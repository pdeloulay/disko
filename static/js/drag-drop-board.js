// Drag and Drop Board Component
// This module provides drag-and-drop functionality for the board view

class DragDropBoard {
    constructor() {
        this.boardData = window.boardData || {};
        this.isAdmin = this.boardData.isAdmin || false;
        this.boardId = this.boardData.boardId;
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
        this.loadBoard();
    }

    setupEventListeners() {
        // Global event listeners for drag and drop
        document.addEventListener('dragstart', this.handleDragStart.bind(this));
        document.addEventListener('dragend', this.handleDragEnd.bind(this));
        document.addEventListener('dragover', this.handleDragOver.bind(this));
        document.addEventListener('drop', this.handleDrop.bind(this));
        document.addEventListener('dragenter', this.handleDragEnter.bind(this));
        document.addEventListener('dragleave', this.handleDragLeave.bind(this));

        // Click outside to close menus
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.idea-card-menu')) {
                this.closeAllMenus();
            }
        });

        // WebSocket event listeners for real-time updates
        document.addEventListener('feedbackUpdated', (event) => {
            this.handleFeedbackUpdate(event.detail);
        });

        document.addEventListener('ideaUpdated', (event) => {
            this.handleIdeaUpdate(event.detail);
        });
    }

    async loadBoard() {
        const boardContainer = document.getElementById('drag-drop-board');
        const boardTitle = document.getElementById('board-title');
        const boardDescription = document.getElementById('board-description');

        console.log('[DragDropBoard] loadBoard started - BoardID:', this.boardId, 'IsAdmin:', this.isAdmin);

        try {
            // Load board details
            const endpoint = `/boards/${this.boardId}`;
            console.log('[DragDropBoard] Making API call to:', endpoint);
            console.log('[DragDropBoard] API base URL:', window.api.baseURL);
            console.log('[DragDropBoard] Full URL will be:', window.api.baseURL + endpoint);
            
            const response = await window.api.get(endpoint);
            console.log('[DragDropBoard] API response received:', response);
            const board = response.data || response;

            // Store board data for column filtering
            this.board = board;

            if (boardTitle) {
                boardTitle.textContent = board.name || 'Untitled Board';
            }
            
            if (boardDescription) {
                boardDescription.textContent = board.description || '';
                boardDescription.style.display = board.description ? 'block' : 'none';
            }

            // Update board settings manager with current board data
            if (window.boardSettingsManager) {
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
                isAdmin: this.isAdmin
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
        console.log('[DragDropBoard] loadIdeas started - BoardID:', this.boardId, 'IsAdmin:', this.isAdmin);
        
        try {
            const endpoint = `/boards/${this.boardId}/ideas`;
            console.log('[DragDropBoard] Making ideas API call to:', endpoint);
            
            const response = await window.api.get(endpoint);
            console.log('[DragDropBoard] Ideas API response received:', response);
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
                isAdmin: this.isAdmin
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
        
        if (this.ideas.length === 0) {
            boardContainer.innerHTML = this.createEmptyState();
            return;
        }

        // Group ideas by column
        const ideasByColumn = this.groupIdeasByColumn();
        
        // Filter columns based on visibility settings
        const visibleColumns = this.getVisibleColumns();
        
        // Render only visible columns
        const columnsHtml = visibleColumns.map(column => 
            this.createColumnView(column, ideasByColumn[column.id] || [])
        ).join('');
        
        boardContainer.innerHTML = columnsHtml;
        
        // Make idea cards draggable if admin
        if (this.isAdmin) {
            this.makeDraggable();
        }
    }

    getVisibleColumns() {
        // For admin users, show all columns
        if (this.isAdmin) {
            return this.columns;
        }
        
        // For public users, filter based on board visibility settings
        if (this.board && this.board.visibleColumns) {
            return this.columns.filter(column => 
                this.board.visibleColumns.includes(column.id)
            );
        }
        
        // Default to all columns if no settings
        return this.columns;
    }

    shouldShowField(fieldName) {
        // For admin users, show all fields except RICE score is admin-only
        if (this.isAdmin) {
            return true;
        }
        
        // oneLiner is always visible
        if (fieldName === 'oneLiner') {
            return true;
        }
        
        // riceScore is admin-only
        if (fieldName === 'riceScore') {
            return false;
        }
        
        // For public users, filter based on board visibility settings
        if (this.board && this.board.visibleFields) {
            return this.board.visibleFields.includes(fieldName);
        }
        
        // Default to showing all fields except RICE score
        return fieldName !== 'riceScore';
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
                    ${this.isAdmin ? this.createColumnSortControls(column.id) : ''}
                </div>
                <div class="column-ideas ${isEmpty ? 'empty' : ''}" data-column="${column.id}">
                    ${ideas.map(idea => this.createIdeaCard(idea)).join('')}
                </div>
            </div>
        `;
    }

    createColumnSortControls(columnId) {
        return `
            <div class="column-sort-controls">
                <select class="column-sort-select" data-column="${columnId}" onchange="dragDropBoard.sortColumn('${columnId}', this.value)">
                    <option value="">Sort by...</option>
                    <option value="name-asc">Name A-Z</option>
                    <option value="name-desc">Name Z-A</option>
                    <option value="rice-desc">RICE Score ↓</option>
                    <option value="rice-asc">RICE Score ↑</option>
                    <option value="status-progress">In Progress First</option>
                    <option value="created-desc">Newest First</option>
                    <option value="created-asc">Oldest First</option>
                </select>
            </div>
        `;
    }

    createIdeaCard(idea) {
        const riceScore = this.calculateRICEScore(idea.riceScore);
        const statusClass = idea.inProgress ? 'in-progress' : '';
        
        return `
            <div class="idea-card ${statusClass}" 
                 data-idea-id="${idea.id}" 
                 data-column="${idea.column}"
                 ${this.isAdmin ? 'draggable="true"' : ''}>
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
                
                ${this.shouldShowField('riceScore') ? `
                    <div class="idea-rice-score">
                        <span class="rice-label">RICE Score:</span>
                        <span class="rice-value">${riceScore.toFixed(1)}</span>
                        <div class="rice-breakdown">
                            R:${idea.riceScore.reach}% I:${idea.riceScore.impact}% C:${idea.riceScore.confidence} E:${idea.riceScore.effort}%
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
                    <span class="idea-status">${this.formatStatus(idea.status)}</span>
                    <span class="idea-column">${this.formatColumn(idea.column)}</span>
                    <span class="idea-date">${this.formatDate(idea.createdAt)}</span>
                </div>
            </div>
        `;
    }

    makeDraggable() {
        const ideaCards = document.querySelectorAll('.idea-card[draggable="true"]');
        ideaCards.forEach(card => {
            card.addEventListener('dragstart', this.handleDragStart.bind(this));
            card.addEventListener('dragend', this.handleDragEnd.bind(this));
        });
    }

    // Drag and Drop Event Handlers
    handleDragStart(e) {
        if (!this.isAdmin) return;
        
        const ideaCard = e.target.closest('.idea-card');
        if (!ideaCard) return;
        
        this.draggedElement = ideaCard;
        this.draggedIdeaId = ideaCard.dataset.ideaId;
        
        ideaCard.classList.add('dragging');
        
        // Set drag data
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('text/html', ideaCard.outerHTML);
        e.dataTransfer.setData('text/plain', this.draggedIdeaId);
        
        // Close any open menus
        this.closeAllMenus();
    }

    handleDragEnd(e) {
        if (!this.isAdmin) return;
        
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
        if (!this.isAdmin || !this.draggedElement) return;
        
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
    }

    handleDragEnter(e) {
        if (!this.isAdmin || !this.draggedElement) return;
        
        const column = e.target.closest('.board-column');
        if (column) {
            column.classList.add('drag-over');
        }
    }

    handleDragLeave(e) {
        if (!this.isAdmin || !this.draggedElement) return;
        
        const column = e.target.closest('.board-column');
        if (column && !column.contains(e.relatedTarget)) {
            column.classList.remove('drag-over');
        }
    }

    async handleDrop(e) {
        if (!this.isAdmin || !this.draggedElement) return;
        
        e.preventDefault();
        
        const targetColumn = e.target.closest('.board-column');
        if (!targetColumn) return;
        
        const targetColumnId = targetColumn.dataset.column;
        const sourceColumnId = this.draggedElement.dataset.column;
        
        console.log('Drop event:', { targetColumnId, sourceColumnId, draggedIdeaId: this.draggedIdeaId });
        
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
        if (!riceScore || riceScore.effort === 0) return 0;
        return (riceScore.reach * riceScore.impact * riceScore.confidence) / riceScore.effort;
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

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    createEmptyState() {
        if (this.isAdmin) {
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
                    <p>The board owner hasn't added any ideas yet. Check back later!</p>
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
        if (window.ideaManager) {
            window.ideaManager.editIdea(ideaId);
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
        if (window.ideaManager) {
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
    }

    showEmojiPicker(ideaId) {
        if (window.ideaManager) {
            window.ideaManager.showEmojiPicker(ideaId);
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
        const thumbsUpCount = ideaCard.querySelector('.thumbs-up-count');
        if (thumbsUpCount) {
            thumbsUpCount.textContent = idea.thumbsUp || 0;
        }

        // Update emoji reactions
        const emojiContainer = ideaCard.querySelector('.emoji-reactions');
        if (emojiContainer && idea.emojiReactions) {
            emojiContainer.innerHTML = '';
            idea.emojiReactions.forEach(reaction => {
                const emojiSpan = document.createElement('span');
                emojiSpan.className = 'emoji-reaction';
                emojiSpan.innerHTML = `${reaction.emoji} ${reaction.count}`;
                emojiContainer.appendChild(emojiSpan);
            });
        }
    }

    handlePositionUpdate(detail) {
        // If this is an update from another client, refresh the board
        // to show the new position
        if (detail.ideaId !== this.draggedIdeaId) {
            this.loadBoardData();
        }
    }

    handleStatusUpdate(detail) {
        // Refresh the board to show status changes
        this.loadBoardData();
    }
}

// Initialize drag-drop board
document.addEventListener('DOMContentLoaded', () => {
    window.dragDropBoard = new DragDropBoard();
});