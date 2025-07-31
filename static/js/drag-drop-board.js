// Drag and Drop Board Component
// This module provides drag-and-drop functionality for the board view

class DragDropBoard {
    constructor() {
        this.boardData = window.boardData || {};
        this.isAdmin = this.boardData.isAdmin || false;
        this.isPublic = this.boardData.isPublic || false;
        this.boardId = this.boardData.boardId;
        this.publicLink = this.boardData.publicLink;
        
        console.log('[DragDropBoard] Constructor - BoardData:', this.boardData);
        console.log('[DragDropBoard] Constructor - IsAdmin:', this.isAdmin);
        console.log('[DragDropBoard] Constructor - IsPublic:', this.isPublic);
        console.log('[DragDropBoard] Constructor - BoardID:', this.boardId);
        console.log('[DragDropBoard] Constructor - PublicLink:', this.publicLink);
        
        this.ideas = [];
        this.columns = [
            { id: 'parking', title: 'Parking', description: 'Ideas waiting to be prioritized' },
            { id: 'now', title: 'Now', description: 'Currently working on' },
            { id: 'next', title: 'Next', description: 'Up next in the pipeline' },
            { id: 'later', title: 'Later', description: 'Future considerations' },
            { id: 'release', title: 'Release', description: 'Completed and released' },
            { id: 'wont-do', title: "Won't Do", description: 'Decided not to pursue' }
        ];
        this.draggedElement = null;
        this.draggedIdeaId = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupGlobalSort();
        this.loadBoard();
    }

    setupEventListeners() {
        // Global event listeners for drag and drop (only for admin users)
        if (this.isAdmin) {
            document.addEventListener('dragstart', this.handleDragStart.bind(this));
            document.addEventListener('dragend', this.handleDragEnd.bind(this));
            document.addEventListener('dragover', this.handleDragOver.bind(this));
            document.addEventListener('drop', this.handleDrop.bind(this));
            document.addEventListener('dragenter', this.handleDragEnter.bind(this));
            document.addEventListener('dragleave', this.handleDragLeave.bind(this));
        }

        // Click outside to close menus
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.idea-card-menu')) {
                this.closeAllMenus();
            }
        });

        // WebSocket event listeners for real-time updates (only for admin users)
        if (this.isAdmin) {
            document.addEventListener('feedbackUpdated', (event) => {
                this.handleFeedbackUpdate(event.detail);
            });

            document.addEventListener('ideaUpdated', (event) => {
                this.handleIdeaUpdate(event.detail);
            });
        }
    }

    async loadBoard() {
        const boardContainer = document.getElementById('drag-drop-board');
        const boardTitle = document.getElementById('board-title');
        const boardDescription = document.getElementById('board-description');

        try {
            // Load board details - use appropriate endpoint based on public/private
            let endpoint;
            if (this.isPublic) {
                endpoint = `/boards/${this.publicLink}/public`;
            } else {
                endpoint = `/boards/${this.boardId}`;
            }
            
            const response = await window.api.get(endpoint);
            const board = response.data || response;

            // Store board data for column filtering
            this.board = board;
            
            // Update admin status based on board data (only for private boards)
            if (!this.isPublic) {
                this.isAdmin = board.isAdmin || false;
            }
            
            // Update window.boardData for consistency across components
            window.boardData = {
                boardId: this.boardId,
                isAdmin: this.isAdmin,
                isPublic: this.isPublic,
                publicLink: this.publicLink
            };

            if (boardTitle) {
                boardTitle.textContent = board.name || 'Untitled Board';
            }
            
            if (boardDescription) {
                boardDescription.textContent = board.description || '';
                boardDescription.style.display = board.description ? 'block' : 'none';
            }

            // Update page title
            const pageTitle = this.isPublic ? 
                `${board.name || 'Untitled Board'} - Public Board` :
                `${board.name || 'Untitled Board'} - Board ${this.boardId}`;
            document.title = pageTitle;

            // Update board settings manager with current board data (only for admin users)
            if (this.isAdmin && window.boardSettingsManager) {
                window.boardSettingsManager.setBoard(board);
            }

            // Load ideas
            await this.loadIdeas();
            
            // Render the board
            this.renderBoard();

        } catch (error) {
            console.error('[DragDropBoard] Failed to load board:', error);
            console.error('[DragDropBoard] Error details:', {
                message: error.message,
                stack: error.stack,
                boardId: this.boardId,
                isAdmin: this.isAdmin,
                isPublic: this.isPublic
            });
            
            if (boardTitle) {
                boardTitle.textContent = 'Error Loading Board';
            }
            
            boardContainer.innerHTML = `
                <div class="error">
                    <h3>Error</h3>
                    <p>Failed to load board. Please try again.</p>
                    <p>Error: ${error.message}</p>
                    <button class="btn btn-primary" onclick="dragDropBoard.loadBoard()">Try Again</button>
                </div>
            `;
        }
    }

    async loadIdeas() {
        console.log('[DragDropBoard] loadIdeas started - BoardID:', this.boardId, 'IsAdmin:', this.isAdmin, 'IsPublic:', this.isPublic);
        
        try {
            // Use appropriate endpoint based on public/private
            let endpoint;
            if (this.isPublic) {
                endpoint = `/boards/${this.publicLink}/ideas/public`;
            } else {
                endpoint = `/boards/${this.boardId}/ideas`;
            }
            
            const response = await window.api.get(endpoint);
            const data = response.data || response;
            
            this.ideas = data.ideas || [];

            // Update ideas count
            const ideasCount = document.getElementById('ideas-count');
            if (ideasCount) {
                const count = this.ideas.length;
                ideasCount.textContent = `${count} ${count === 1 ? 'idea' : 'ideas'}`;
            }

            // Update tab counts
            this.updateTabCounts();
            
        } catch (error) {
            console.error('[DragDropBoard] Failed to load ideas:', error);
            console.error('[DragDropBoard] Ideas error details:', {
                message: error.message,
                stack: error.stack,
                boardId: this.boardId,
                isAdmin: this.isAdmin,
                isPublic: this.isPublic
            });
            this.ideas = [];
            
            const ideasCount = document.getElementById('ideas-count');
            if (ideasCount) {
                ideasCount.textContent = 'Error loading';
            }
        }
    }

    renderBoard() {
        const boardContainer = document.getElementById('drag-drop-board');
        
        if (!boardContainer) {
            console.error('[DragDropBoard] Board container not found!');
            return;
        }
        
        if (this.ideas.length === 0) {
            boardContainer.innerHTML = this.createEmptyState();
            return;
        }

        // Group ideas by column
        const ideasByColumn = this.groupIdeasByColumn();
        
        // Filter columns based on visibility settings
        const visibleColumns = this.getVisibleColumns();
        
        // Render only visible columns
        const columnsHtml = visibleColumns.map(column => {
            const columnIdeas = ideasByColumn[column.id] || [];
            return this.createColumnView(column, columnIdeas);
        }).join('');
        
        boardContainer.innerHTML = columnsHtml;
        
        // Make idea cards draggable if admin and not public board
        if (this.isAdmin && !this.isPublic) {
            this.makeDraggable();
        }
    }

    getVisibleColumns() {
        // For public boards, show only specific columns: Now, Next, Later, Won't Do
        if (this.isPublic) {
            return this.columns.filter(column => 
                ['now', 'next', 'later', 'wont-do'].includes(column.id)
            );
        }
        
        // For private boards (both admin and non-admin), filter based on board visibility settings
        if (this.board && this.board.visibleColumns && this.board.visibleColumns.length > 0) {
            return this.columns.filter(column => 
                this.board.visibleColumns.includes(column.id)
            );
        }
        
        // Default to all columns if no settings
        return this.columns;
    }

    shouldShowField(fieldName) {
        // oneLiner is always visible
        if (fieldName === 'oneLiner') {
            return true;
        }
        
        // For public boards, show all fields (read-only mode)
        if (this.isPublic) {
            return true;
        }
        
        // For private boards (both admin and non-admin), filter based on board visibility settings
        if (this.board && this.board.visibleFields && this.board.visibleFields.length > 0) {
            return this.board.visibleFields.includes(fieldName);
        }
        
        // Default to showing all fields except RICE score for non-admin users on private boards
        if (!this.isAdmin && fieldName === 'riceScore') {
            return false;
        }
        
        return true;
    }

    groupIdeasByColumn() {
        const grouped = {};
        
        // Initialize all columns
        this.columns.forEach(column => {
            grouped[column.id] = [];
        });
        
        // Group ideas by column and sort by position
        this.ideas.forEach(idea => {
            const column = idea.column || 'parking';
            if (grouped[column]) {
                grouped[column].push(idea);
            }
        });
        
        // Sort each column by position
        Object.keys(grouped).forEach(columnId => {
            grouped[columnId].sort((a, b) => (a.position || 0) - (b.position || 0));
        });
        
        return grouped;
    }

    createColumnView(column, ideas) {
        const isEmpty = ideas.length === 0;
        
        return `
            <div class="board-column" data-column="${column.id}">
                <div class="column-header">
                    <div class="column-title-section">
                        <div class="column-title">${column.title}</div>
                        <div class="column-count">${ideas.length}</div>
                    </div>
                </div>
                <div class="column-ideas ${isEmpty ? 'empty' : ''}" data-column="${column.id}">
                    ${ideas.map(idea => this.createIdeaCard(idea)).join('')}
                </div>
            </div>
        `;
    }

    setupGlobalSort() {
        const globalSortSelect = document.getElementById('global-sort');
        if (globalSortSelect) {
            globalSortSelect.addEventListener('change', (e) => {
                this.sortAllColumns(e.target.value);
            });
        }
    }

    createIdeaCard(idea) {
        const riceScore = this.calculateRICEScore(idea.riceScore);
        const statusClass = idea.inProgress ? 'in-progress' : '';
        
        return `
            <div class="idea-card ${statusClass}" 
                 data-idea-id="${idea.id}" 
                 data-column="${idea.column}"
                 ${this.isAdmin && !this.isPublic ? 'draggable="true"' : ''}>
                <div class="idea-card-header">
                    <h4 class="idea-title idea-oneliner">${this.escapeHtml(idea.oneLiner)}</h4>
                    ${this.isAdmin ? `
                        <div class="idea-card-menu">
                            <button class="btn-menu" onclick="dragDropBoard.toggleIdeaMenu('${idea.id}')">⋮</button>
                            <div class="idea-menu" id="idea-menu-${idea.id}" style="display: none;">
                                <button onclick="dragDropBoard.editIdea('${idea.id}')">✏️ Edit</button>
                                <button onclick="dragDropBoard.toggleInProgress('${idea.id}', ${!idea.inProgress})">
                                    ${idea.inProgress ? '⏸️ Mark as Not In Progress' : '▶️ Mark as In Progress'}
                                </button>
                                ${idea.status !== 'done' ? `
                                    <button onclick="dragDropBoard.updateIdeaStatus('${idea.id}', 'done')">✅ Mark as Done</button>
                                ` : ''}
                                ${idea.status === 'done' ? `
                                    <button onclick="dragDropBoard.updateIdeaStatus('${idea.id}', 'active')">🔄 Reactivate</button>
                                ` : ''}
                                ${idea.status !== 'archived' ? `
                                    <button onclick="dragDropBoard.updateIdeaStatus('${idea.id}', 'archived')">🗃️ Archive</button>
                                ` : ''}
                                <button onclick="dragDropBoard.confirmDeleteIdea('${idea.id}', '${this.escapeHtml(idea.oneLiner)}')">🗑️ Delete</button>
                            </div>
                        </div>
                    ` : ''}
                </div>
                
                <div class="idea-content">
                    ${this.shouldShowField('description') ? `<p class="idea-description">${this.escapeHtml(idea.description)}</p>` : ''}
                    ${this.shouldShowField('valueStatement') ? `<p class="idea-value-statement"><strong>Value:</strong> <span class="idea-value">${this.escapeHtml(idea.valueStatement)}</span></p>` : ''}
                </div>
                
                ${this.shouldShowField('riceScore') && idea.riceScore ? `
                    <div class="idea-rice-score">
                        <span class="rice-label">RICE Score:</span>
                        <span class="rice-value">${riceScore.toFixed(1)}</span>
                        <div class="rice-breakdown">
                            R:${idea.riceScore.reach || 0}% I:${idea.riceScore.impact || 0}% C:${idea.riceScore.confidence || 0} E:${idea.riceScore.effort || 1}%
                        </div>
                    </div>
                ` : ''}
                
                <div class="idea-feedback">
                    <div class="feedback-actions">
                        <button class="feedback-btn thumbs-up" onclick="dragDropBoard.addThumbsUp('${idea.id}')">
                            👍 <span class="count">${idea.thumbsUp || 0}</span>
                        </button>
                        <button class="feedback-btn emoji" onclick="dragDropBoard.showEmojiPicker('${idea.id}')">
                            😊 <span class="count">${this.getEmojiCount(idea.emojiReactions)}</span>
                        </button>
                    </div>
                    ${idea.emojiReactions && idea.emojiReactions.length > 0 ? `
                        <div class="emoji-reactions">
                            ${idea.emojiReactions.map(reaction => 
                                `<span class="emoji-reaction">${reaction.emoji} ${reaction.count}</span>`
                            ).join('')}
                        </div>
                    ` : ''}
                </div>
                <div class="idea-meta">
                    <span class="idea-date">Updated ${this.formatTimeAgo(idea.updatedAt)}</span>
                </div>
            </div>
        `;
    }

    makeDraggable() {
        // Only make cards draggable for admin users on private boards
        if (!this.isAdmin || this.isPublic) {
            return;
        }
        
        const ideaCards = document.querySelectorAll('.idea-card[draggable="true"]');
        
        ideaCards.forEach((card) => {
            card.addEventListener('dragstart', this.handleDragStart.bind(this));
            card.addEventListener('dragend', this.handleDragEnd.bind(this));
        });
    }

    // Drag and Drop Event Handlers
    handleDragStart(e) {
        console.log('[DragDropBoard] Drag start event triggered');
        console.log('[DragDropBoard] IsAdmin:', this.isAdmin, 'IsPublic:', this.isPublic);
        
        if (!this.isAdmin || this.isPublic) {
            console.log('[DragDropBoard] Drag blocked - user not admin or public board');
            return;
        }
        
        const ideaCard = e.target.closest('.idea-card');
        if (!ideaCard) {
            console.log('[DragDropBoard] No idea card found in drag target');
            return;
        }
        
        console.log('[DragDropBoard] Starting drag for idea:', ideaCard.dataset.ideaId);
        
        this.draggedElement = ideaCard;
        this.draggedIdeaId = ideaCard.dataset.ideaId;
        
        ideaCard.classList.add('dragging');
        
        // Set drag data
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('text/html', ideaCard.outerHTML);
        e.dataTransfer.setData('text/plain', this.draggedIdeaId);
        
        // Close any open menus
        this.closeAllMenus();
        
        console.log('[DragDropBoard] Drag started successfully');
    }

    handleDragEnd(e) {
        if (!this.isAdmin || this.isPublic) return;
        
        const ideaCard = e.target.closest('.idea-card');
        if (ideaCard) {
            ideaCard.classList.remove('dragging');
        }
        
        // Remove drag-over styling from all columns
        document.querySelectorAll('.board-column').forEach(column => {
            column.classList.remove('drag-over');
        });
        
        this.draggedElement = null;
        this.draggedIdeaId = null;
    }

    handleDragOver(e) {
        if (!this.isAdmin || this.isPublic || !this.draggedElement) return;
        
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
    }

    handleDragEnter(e) {
        if (!this.isAdmin || this.isPublic || !this.draggedElement) return;
        
        const column = e.target.closest('.board-column');
        if (column) {
            column.classList.add('drag-over');
        }
    }

    handleDragLeave(e) {
        if (!this.isAdmin || this.isPublic || !this.draggedElement) return;
        
        const column = e.target.closest('.board-column');
        if (column && !column.contains(e.relatedTarget)) {
            column.classList.remove('drag-over');
        }
    }

    async handleDrop(e) {
        console.log('[DragDropBoard] Drop event triggered');
        console.log('[DragDropBoard] IsAdmin:', this.isAdmin, 'IsPublic:', this.isPublic, 'DraggedElement:', !!this.draggedElement);
        
        if (!this.isAdmin || this.isPublic || !this.draggedElement) {
            console.log('[DragDropBoard] Drop blocked - not admin, public board, or no dragged element');
            return;
        }
        
        e.preventDefault();
        
        const targetColumn = e.target.closest('.board-column');
        if (!targetColumn) {
            console.log('[DragDropBoard] No target column found');
            return;
        }
        
        const targetColumnId = targetColumn.dataset.column;
        const sourceColumnId = this.draggedElement.dataset.column;
        
        console.log('[DragDropBoard] Drop event:', { targetColumnId, sourceColumnId, draggedIdeaId: this.draggedIdeaId });
        
        // Remove drag-over styling
        targetColumn.classList.remove('drag-over');
        
        // Don't do anything if dropped in the same column
        if (targetColumnId === sourceColumnId) {
            console.log('Dropped in same column, no action needed');
            return;
        }
        
        try {
            // Calculate new position (append to end of target column)
            const targetColumnIdeas = targetColumn.querySelectorAll('.idea-card');
            const newPosition = targetColumnIdeas.length + 1;
            
            console.log('Calculated new position:', newPosition);
            
            // Update idea position via API
            await this.updateIdeaPosition(this.draggedIdeaId, targetColumnId, newPosition);
            
            // Show success message
            this.showSuccessMessage(`Idea moved to ${this.formatColumn(targetColumnId)}!`);
            
            // Reload the board to reflect changes
            await this.loadIdeas();
            this.renderBoard();
            
        } catch (error) {
            console.error('Failed to update idea position:', error);
            this.showErrorMessage('Failed to move idea. Please try again.');
        }
    }

    async updateIdeaPosition(ideaId, column, position) {
        console.log('Updating idea position:', { ideaId, column, position });
        const response = await window.api.put(`/ideas/${ideaId}/position`, {
            column: column,
            position: position
        });
        console.log('Position update response:', response);
        return response;
    }

    // Utility Methods
    calculateRICEScore(riceScore) {
        if (!riceScore || !riceScore.effort || riceScore.effort === 0) return 0;
        // Convert percentages to decimals (0-100 -> 0-1)
        const reach = (riceScore.reach || 0) / 100;
        const impact = (riceScore.impact || 0) / 100;
        const confidence = (riceScore.confidence || 0) / 100;
        return (reach * impact * confidence) / riceScore.effort;
    }

    getEmojiCount(emojiReactions) {
        if (!emojiReactions || !Array.isArray(emojiReactions)) return 0;
        return emojiReactions.reduce((total, reaction) => total + (reaction.count || 0), 0);
    }

    formatStatus(status) {
        const statusMap = {
            'active': 'Active',
            'done': 'Done',
            'draft': 'Draft',
            'archived': 'Archived'
        };
        return statusMap[status] || status;
    }

    formatColumn(column) {
        const columnMap = {
            'parking': 'Parking',
            'now': 'Now',
            'next': 'Next',
            'later': 'Later',
            'release': 'Release',
            'wont-do': "Won't Do"
        };
        return columnMap[column] || column;
    }

    formatDate(dateString) {
        if (!dateString) return '';
        const date = new Date(dateString);
        return date.toLocaleDateString();
    }

    formatTimeAgo(dateString) {
        if (!dateString) return '';
        
        const date = new Date(dateString);
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) {
            return 'just now';
        }
        
        const diffInMinutes = Math.floor(diffInSeconds / 60);
        if (diffInMinutes < 60) {
            return `${diffInMinutes}m ago`;
        }
        
        const diffInHours = Math.floor(diffInMinutes / 60);
        if (diffInHours < 24) {
            return `${diffInHours}h ago`;
        }
        
        const diffInDays = Math.floor(diffInHours / 24);
        if (diffInDays < 7) {
            return `${diffInDays}d ago`;
        }
        
        const diffInWeeks = Math.floor(diffInDays / 7);
        if (diffInWeeks < 4) {
            return `${diffInWeeks}w ago`;
        }
        
        const diffInMonths = Math.floor(diffInDays / 30);
        if (diffInMonths < 12) {
            return `${diffInMonths}mo ago`;
        }
        
        const diffInYears = Math.floor(diffInDays / 365);
        return `${diffInYears}y ago`;
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    createEmptyState() {
        if (this.isAdmin && !this.isPublic) {
            return `
                <div class="ideas-empty-state">
                    <h3>No ideas yet</h3>
                    <p>Create your first idea to get started with your board!</p>
                    <button class="btn btn-primary create-idea-btn" onclick="window.ideaManager.openCreateModal()">
                        Create Your First Idea
                    </button>
                </div>
            `;
        } else {
            return `
                <div class="ideas-empty-state">
                    <h3>No ideas to display</h3>
                    <p>${this.isPublic ? 'This public board has no ideas yet.' : 'The board owner hasn\'t added any ideas yet. Check back later!'}</p>
                </div>
            `;
        }
    }

    // Menu and Action Handlers
    toggleIdeaMenu(ideaId) {
        const menu = document.getElementById(`idea-menu-${ideaId}`);
        const allMenus = document.querySelectorAll('.idea-menu');
        
        // Close all other menus
        allMenus.forEach(m => {
            if (m.id !== `idea-menu-${ideaId}`) {
                m.style.display = 'none';
            }
        });
        
        // Toggle current menu
        if (menu) {
            menu.style.display = menu.style.display === 'none' ? 'block' : 'none';
        }
    }

    closeAllMenus() {
        const allMenus = document.querySelectorAll('.idea-menu');
        allMenus.forEach(menu => {
            menu.style.display = 'none';
        });
    }

    // Delegate to idea manager for these actions
    editIdea(ideaId) {
        console.log('[DragDropBoard] editIdea called with ideaId:', ideaId);
        console.log('[DragDropBoard] ideaManager available:', !!window.ideaManager);
        console.log('[DragDropBoard] ideas available:', !!this.ideas, 'count:', this.ideas?.length);
        
        if (window.ideaManager) {
            window.ideaManager.editIdea(ideaId);
        } else {
            console.error('[DragDropBoard] ideaManager not available');
        }
        this.closeAllMenus();
    }

    async toggleInProgress(ideaId, inProgress) {
        if (window.ideaManager) {
            await window.ideaManager.toggleInProgress(ideaId, inProgress);
            // Reload board after status change
            await this.loadIdeas();
            this.renderBoard();
        }
        this.closeAllMenus();
    }

    async markAsDone(ideaId) {
        if (window.ideaManager) {
            await window.ideaManager.markAsDone(ideaId);
            // Reload board after status change
            await this.loadIdeas();
            this.renderBoard();
        }
        this.closeAllMenus();
    }

    async updateIdeaStatus(ideaId, status, inProgress = null) {
        if (window.ideaManager) {
            await window.ideaManager.updateIdeaStatus(ideaId, status, inProgress);
            // Reload board after status change
            await this.loadIdeas();
            this.renderBoard();
        }
        this.closeAllMenus();
    }

    confirmDeleteIdea(ideaId, ideaTitle) {
        if (window.ideaManager) {
            window.ideaManager.confirmDeleteIdea(ideaId, ideaTitle);
        }
        this.closeAllMenus();
    }

    async addThumbsUp(ideaId) {
        console.log('[DragDropBoard] addThumbsUp called for idea:', ideaId);
        
        try {
            // For public boards, use public API endpoint
            if (this.isPublic) {
                const response = await window.api.post(`/ideas/${ideaId}/thumbsup`);
                console.log('[DragDropBoard] Thumbs up response:', response);
                
                // Update the UI optimistically
                const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
                if (ideaCard) {
                    const countSpan = ideaCard.querySelector('.thumbs-up .count');
                    if (countSpan) {
                        const currentCount = parseInt(countSpan.textContent) || 0;
                        countSpan.textContent = currentCount + 1;
                    }
                }
                
                this.showSuccessMessage('Thumbs up added!');
            } else if (window.ideaManager) {
                // For private boards, use idea manager
                await window.ideaManager.addThumbsUp(ideaId);
                // Update the UI optimistically
                const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
                if (ideaCard) {
                    const countSpan = ideaCard.querySelector('.thumbs-up .count');
                    if (countSpan) {
                        const currentCount = parseInt(countSpan.textContent) || 0;
                        countSpan.textContent = currentCount + 1;
                    }
                }
            }
        } catch (error) {
            console.error('[DragDropBoard] Failed to add thumbs up:', error);
            this.showErrorMessage('Failed to add thumbs up. Please try again.');
        }
    }

    showEmojiPicker(ideaId) {
        console.log('[DragDropBoard] showEmojiPicker called for idea:', ideaId);
        
        // Create emoji picker modal for public boards
        if (this.isPublic) {
            this.createEmojiPickerModal(ideaId);
        } else if (window.ideaManager) {
            // For private boards, use idea manager
            window.ideaManager.showEmojiPicker(ideaId);
        }
    }

    createEmojiPickerModal(ideaId) {
        // Remove existing emoji picker
        const existingPicker = document.getElementById('emoji-picker-modal');
        if (existingPicker) {
            existingPicker.remove();
        }

        const emojis = ['🚀', '💡', '🎯', '🔥', '👍', '❤️', '😊', '🎉', '⭐', '💪'];
        
        const modal = document.createElement('div');
        modal.id = 'emoji-picker-modal';
        modal.className = 'modal show';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Add Emoji Reaction</h3>
                    <button class="modal-close" onclick="dragDropBoard.closeEmojiPicker()">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="emoji-grid">
                        ${emojis.map(emoji => `
                            <button class="emoji-option" onclick="dragDropBoard.addEmojiReaction('${ideaId}', '${emoji}')">
                                ${emoji}
                            </button>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(modal);
    }

    closeEmojiPicker() {
        const modal = document.getElementById('emoji-picker-modal');
        if (modal) {
            modal.remove();
        }
    }

    async addEmojiReaction(ideaId, emoji) {
        console.log('[DragDropBoard] addEmojiReaction called for idea:', ideaId, 'emoji:', emoji);
        
        try {
            // For public boards, use public API endpoint
            if (this.isPublic) {
                const response = await window.api.post(`/ideas/${ideaId}/emoji`, { emoji: emoji });
                console.log('[DragDropBoard] Emoji reaction response:', response);
                
                this.closeEmojiPicker();
                this.showSuccessMessage('Emoji reaction added!');
                
                // Refresh the board to show updated emoji reactions
                await this.loadIdeas();
                this.renderBoard();
            } else if (window.ideaManager) {
                // For private boards, use idea manager
                await window.ideaManager.addEmojiReaction(ideaId, emoji);
            }
        } catch (error) {
            console.error('[DragDropBoard] Failed to add emoji reaction:', error);
            this.showErrorMessage('Failed to add emoji reaction. Please try again.');
        }
    }

    // Message display methods
    showSuccessMessage(message) {
        this.showMessage(message, 'success');
    }

    showErrorMessage(message) {
        this.showMessage(message, 'error');
    }

    showMessage(message, type) {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message-toast');
        existingMessages.forEach(msg => msg.remove());

        // Create new message
        const messageEl = document.createElement('div');
        messageEl.className = `message-toast ${type}`;
        messageEl.textContent = message;
        
        // Add to page
        document.body.appendChild(messageEl);
        
        // Auto remove after 3 seconds
        setTimeout(() => {
            if (messageEl.parentNode) {
                messageEl.remove();
            }
        }, 3000);
        
        // Remove on click
        messageEl.addEventListener('click', () => {
            messageEl.remove();
        });
    }

    // Method to refresh board (called by idea manager)
    async refreshBoard() {
        await this.loadIdeas();
        this.renderBoard();
    }

    // Method to update ideas with search results
    updateIdeasWithSearch(searchResults, searchInfo) {
        // Store the search results as current ideas
        this.ideas = searchResults || [];
        this.searchInfo = searchInfo;
        
        // Re-render the board with search results
        this.renderBoard();
        
        // Update ideas count with search context
        this.updateIdeasCountWithSearch(searchInfo);
        
        // Update tab counts
        this.updateTabCounts();
    }

    updateIdeasCountWithSearch(searchInfo) {
        const ideasCount = document.getElementById('ideas-count');
        if (!ideasCount) return;

        const count = this.ideas.length;
        let countText = `${count} ${count === 1 ? 'idea' : 'ideas'}`;
        
        if (searchInfo && (searchInfo.query || this.hasActiveFilters(searchInfo))) {
            countText += ' (filtered)';
        }
        
        ideasCount.textContent = countText;
    }

    hasActiveFilters(searchInfo) {
        if (!searchInfo || !searchInfo.filters) return false;
        
        return searchInfo.filters.column || 
               searchInfo.filters.status || 
               searchInfo.filters.inProgress !== null ||
               (searchInfo.sort && searchInfo.sort.by);
    }

    // Method to sort ideas within a specific column
    sortColumn(columnId, sortType) {
        if (!sortType) {
            // Reset to default order (by position)
            this.renderBoard();
            return;
        }

        const [sortBy, sortDir] = sortType.split('-');
        
        // Get ideas for this column
        const columnIdeas = this.ideas.filter(idea => idea.column === columnId);
        
        // Sort the ideas
        columnIdeas.sort((a, b) => {
            let comparison = 0;
            
            switch (sortBy) {
                case 'name':
                    comparison = a.oneLiner.localeCompare(b.oneLiner);
                    break;
                case 'rice':
                    const riceA = this.calculateRICEScore(a.riceScore);
                    const riceB = this.calculateRICEScore(b.riceScore);
                    comparison = riceA - riceB;
                    break;
                case 'status':
                    // Sort by in-progress first, then by status
                    if (sortDir === 'progress') {
                        if (a.inProgress && !b.inProgress) return -1;
                        if (!a.inProgress && b.inProgress) return 1;
                        return a.status.localeCompare(b.status);
                    }
                    comparison = a.status.localeCompare(b.status);
                    break;
                case 'created':
                    comparison = new Date(a.createdAt) - new Date(b.createdAt);
                    break;
                default:
                    comparison = a.position - b.position;
            }
            
            return sortDir === 'desc' ? -comparison : comparison;
        });
        
        // Update the column display
        const columnElement = document.querySelector(`[data-column="${columnId}"] .column-ideas`);
        if (columnElement) {
            columnElement.innerHTML = columnIdeas.map(idea => this.createIdeaCard(idea)).join('');
            
            // Re-enable dragging if admin
            if (this.isAdmin) {
                this.makeDraggable();
            }
        }
    }

    // Method to sort all columns with the same sort type
    sortAllColumns(sortType) {
        if (!sortType) {
            // Reset to default order (by position)
            this.renderBoard();
            return;
        }

        const [sortBy, sortDir] = sortType.split('-');
        
        // Group ideas by column
        const groupedIdeas = this.groupIdeasByColumn();
        
        // Sort each column's ideas
        Object.keys(groupedIdeas).forEach(columnId => {
            const columnIdeas = groupedIdeas[columnId];
            
            columnIdeas.sort((a, b) => {
                let comparison = 0;
                
                switch (sortBy) {
                    case 'name':
                        comparison = a.oneLiner.localeCompare(b.oneLiner);
                        break;
                    case 'rice':
                        const riceA = this.calculateRICEScore(a.riceScore);
                        const riceB = this.calculateRICEScore(b.riceScore);
                        comparison = riceA - riceB;
                        break;
                    case 'status':
                        // Sort by in-progress first, then by status
                        if (sortDir === 'progress') {
                            if (a.inProgress && !b.inProgress) return -1;
                            if (!a.inProgress && b.inProgress) return 1;
                            return a.status.localeCompare(b.status);
                        }
                        comparison = a.status.localeCompare(b.status);
                        break;
                    case 'created':
                        comparison = new Date(a.createdAt) - new Date(b.createdAt);
                        break;
                    default:
                        comparison = a.position - b.position;
                }
                
                return sortDir === 'desc' ? -comparison : comparison;
            });
            
            // Update the column display
            const columnElement = document.querySelector(`[data-column="${columnId}"] .column-ideas`);
            if (columnElement) {
                columnElement.innerHTML =
                    columnIdeas.map(idea => this.createIdeaCard(idea)).join('');
                
                // Re-enable dragging if admin
                if (this.isAdmin) {
                    this.makeDraggable();
                }
            }
        });
    }

    updateTabCounts() {
        // Update board tab count (exclude release column ideas)
        const boardIdeas = this.ideas.filter(idea => idea.column !== 'release');
        if (window.boardTabs) {
            window.boardTabs.setBoardCount(boardIdeas.length);
        }

        // Update release tab count
        const releaseIdeas = this.ideas.filter(idea => idea.column === 'release');
        if (window.boardTabs) {
            window.boardTabs.setReleaseCount(releaseIdeas.length);
        }
    }

    // WebSocket event handlers for real-time updates
    handleFeedbackUpdate(detail) {
        console.log('Feedback updated for idea:', detail.ideaId);
        
        // Find the idea card and refresh its feedback display
        const ideaCard = document.querySelector(`[data-idea-id="${detail.ideaId}"]`);
        if (ideaCard) {
            this.refreshIdeaFeedback(detail.ideaId, ideaCard);
        }
    }

    handleIdeaUpdate(detail) {
        console.log('Idea updated:', detail);
        
        // Handle different types of idea updates
        if (detail.type === 'position_update') {
            this.handlePositionUpdate(detail);
        } else if (detail.type === 'status_update') {
            this.handleStatusUpdate(detail);
        }
    }

    async refreshIdeaFeedback(ideaId, ideaCard) {
        try {
            // Re-fetch the board ideas to get updated feedback
            const response = await fetch(`/api/boards/${this.boardId}/ideas`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('clerk-db-jwt')}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                const updatedIdea = data.ideas.find(idea => idea.id === ideaId);
                if (updatedIdea) {
                    this.updateIdeaCardFeedback(ideaCard, updatedIdea);
                }
            }
        } catch (error) {
            console.error('Error refreshing idea feedback:', error);
        }
    }

    updateIdeaCardFeedback(ideaCard, idea) {
        // Update thumbs up count
        const thumbsUpCount = ideaCard.querySelector('.thumbs-up .count');
        if (thumbsUpCount) {
            thumbsUpCount.textContent = idea.thumbsUp || 0;
        }

        // Update emoji button count
        const emojiButtonCount = ideaCard.querySelector('.emoji .count');
        if (emojiButtonCount) {
            const totalEmojiCount = idea.emojiReactions ? 
                idea.emojiReactions.reduce((sum, reaction) => sum + reaction.count, 0) : 0;
            emojiButtonCount.textContent = totalEmojiCount;
        }

        // Update emoji reactions display
        const emojiContainer = ideaCard.querySelector('.emoji-reactions');
        if (emojiContainer) {
            if (idea.emojiReactions && idea.emojiReactions.length > 0) {
                emojiContainer.innerHTML = idea.emojiReactions.map(reaction => 
                    `<span class="emoji-reaction">${reaction.emoji} ${reaction.count}</span>`
                ).join('');
            } else {
                emojiContainer.innerHTML = '';
            }
        }
    }

    handlePositionUpdate(detail) {
        // If this is an update from another client, refresh the board
        // to show the new position
        if (detail.ideaId !== this.draggedIdeaId) {
            this.loadBoard();
        }
    }

    handleStatusUpdate(detail) {
        // Refresh the board to show status changes
        this.loadBoard();
    }
}

// Initialize drag-drop board
document.addEventListener('DOMContentLoaded', () => {
    window.dragDropBoard = new DragDropBoard();
});