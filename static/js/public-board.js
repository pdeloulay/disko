// Public Board View functionality
class PublicBoardView {
    constructor() {
        this.boardData = window.boardData || {};
        this.publicLink = this.boardData.publicLink;
        this.board = null;
        this.ideas = [];
        this.init();
    }

    init() {
        this.loadPublicBoard();
        this.bindEvents();
        this.setupWebSocket();
    }

    bindEvents() {
        // Refresh button
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadPublicBoard();
            });
        }
    }

    setupWebSocket() {
        // Initialize WebSocket connection for real-time updates (public view)
        if (this.publicLink && window.WebSocketManager) {
            // Extract board ID from public link or use public link as identifier
            const boardId = this.publicLink;
            this.wsManager = new WebSocketManager(boardId);
            this.wsManager.startKeepAlive();

            // Listen for feedback updates
            document.addEventListener('feedbackUpdated', (event) => {
                this.handleFeedbackUpdate(event.detail);
            });

            // Listen for idea updates
            document.addEventListener('ideaUpdated', (event) => {
                this.handleIdeaUpdate(event.detail);
            });
        }
    }

    handleFeedbackUpdate(detail) {
        // Refresh the specific idea's feedback display
        console.log('Feedback updated for idea:', detail.ideaId);
        this.refreshIdeaFeedback(detail.ideaId);
    }

    handleIdeaUpdate(detail) {
        // Handle real-time idea updates (position, status changes)
        console.log('Idea updated:', detail);
        
        if (detail.type === 'position_update' || detail.type === 'status_update') {
            // Reload the board to reflect changes
            this.loadPublicBoard();
        }
    }

    async refreshIdeaFeedback(ideaId) {
        // Find and refresh the specific idea's feedback display
        const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
        if (ideaCard) {
            try {
                // Re-fetch the idea data and update feedback display
                const response = await fetch(`/api/boards/${this.publicLink}/ideas/public`);
                if (response.ok) {
                    const data = await response.json();
                    const updatedIdea = data.ideas.find(idea => idea.id === ideaId);
                    if (updatedIdea) {
                        this.updateIdeaFeedbackDisplay(ideaCard, updatedIdea);
                    }
                }
            } catch (error) {
                console.error('Error refreshing idea feedback:', error);
            }
        }
    }

    updateIdeaFeedbackDisplay(ideaCard, idea) {
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

    async loadPublicBoard() {
        try {
            // Show loading state
            this.showLoading();

            // Load board data and ideas
            const [boardResponse, ideasResponse] = await Promise.all([
                this.fetchPublicBoard(),
                this.fetchPublicBoardIdeas()
            ]);

            this.board = boardResponse;
            this.ideas = ideasResponse.ideas;

            // Update UI
            this.updateBoardHeader();
            this.renderBoard();

        } catch (error) {
            console.error('Error loading public board:', error);
            this.showError(error.message || 'Failed to load board');
        }
    }

    async fetchPublicBoard() {
        const response = await fetch(`/api/boards/${this.publicLink}/public`);
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error?.message || 'Failed to fetch board');
        }

        return await response.json();
    }

    async fetchPublicBoardIdeas() {
        const response = await fetch(`/api/boards/${this.publicLink}/ideas/public`);
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error?.message || 'Failed to fetch ideas');
        }

        return await response.json();
    }

    updateBoardHeader() {
        const titleElement = document.getElementById('board-title');
        const descriptionElement = document.getElementById('board-description');
        const countElement = document.getElementById('ideas-count');

        if (titleElement) {
            titleElement.textContent = this.board.name;
        }

        if (descriptionElement && this.board.description) {
            descriptionElement.textContent = this.board.description;
            descriptionElement.style.display = 'block';
        }

        if (countElement) {
            countElement.textContent = `${this.ideas.length} ideas`;
        }

        // Update tab counts
        this.updateTabCounts();
    }

    renderBoard() {
        const boardContainer = document.getElementById('drag-drop-board');
        if (!boardContainer) return;

        // Clear existing content
        boardContainer.innerHTML = '';

        // Group ideas by column
        const ideasByColumn = this.groupIdeasByColumn();

        // Create columns based on visible columns
        const columnsContainer = document.createElement('div');
        columnsContainer.className = 'board-columns';

        this.board.visibleColumns.forEach(columnType => {
            const columnElement = this.createColumnElement(columnType, ideasByColumn[columnType] || []);
            columnsContainer.appendChild(columnElement);
        });

        boardContainer.appendChild(columnsContainer);
    }

    groupIdeasByColumn() {
        const grouped = {};
        this.ideas.forEach(idea => {
            if (!grouped[idea.column]) {
                grouped[idea.column] = [];
            }
            grouped[idea.column].push(idea);
        });

        // Sort ideas within each column by position
        Object.keys(grouped).forEach(column => {
            grouped[column].sort((a, b) => a.position - b.position);
        });

        return grouped;
    }

    createColumnElement(columnType, ideas) {
        const column = document.createElement('div');
        column.className = 'board-column';
        column.dataset.column = columnType;

        // Column header
        const header = document.createElement('div');
        header.className = 'column-header';
        header.innerHTML = `
            <h3 class="column-title">${this.getColumnTitle(columnType)}</h3>
            <span class="column-count">${ideas.length}</span>
        `;
        column.appendChild(header);

        // Column content
        const content = document.createElement('div');
        content.className = 'column-content';

        ideas.forEach(idea => {
            const ideaCard = this.createPublicIdeaCard(idea);
            content.appendChild(ideaCard);
        });

        column.appendChild(content);
        return column;
    }

    createPublicIdeaCard(idea) {
        const card = document.createElement('div');
        card.className = `idea-card public-idea-card ${idea.inProgress ? 'in-progress' : ''}`;
        card.dataset.ideaId = idea.id;

        let cardContent = `
            <div class="idea-header">
                <h4 class="idea-title">${this.escapeHtml(idea.oneLiner)}</h4>
                ${idea.inProgress ? '<span class="in-progress-badge">In Progress</span>' : ''}
            </div>
        `;

        // Add description if visible
        if (idea.description && this.board.visibleFields && this.board.visibleFields.includes('description')) {
            cardContent += `
                <div class="idea-description">
                    <p>${this.escapeHtml(idea.description)}</p>
                </div>
            `;
        }

        // Add value statement if visible
        if (idea.valueStatement && this.board.visibleFields && this.board.visibleFields.includes('valueStatement')) {
            cardContent += `
                <div class="idea-value">
                    <p><strong>Value:</strong> ${this.escapeHtml(idea.valueStatement)}</p>
                </div>
            `;
        }

        // Add feedback container
        cardContent += `<div class="idea-feedback-container"></div>`;

        card.innerHTML = cardContent;

        // Create and add feedback widget
        const feedbackWidget = new FeedbackWidget(
            idea.id, 
            idea.thumbsUp, 
            idea.emojiReactions
        );
        
        const feedbackContainer = card.querySelector('.idea-feedback-container');
        feedbackContainer.appendChild(feedbackWidget.getElement());

        // Store widget reference for updates
        card.feedbackWidget = feedbackWidget;

        return card;
    }



    getColumnTitle(columnType) {
        const titles = {
            'parking': 'Parking',
            'now': 'Now',
            'next': 'Next',
            'later': 'Later',
            'release': 'Released',
            'wont-do': "Won't Do"
        };
        return titles[columnType] || columnType;
    }

    showLoading() {
        const boardContainer = document.getElementById('drag-drop-board');
        if (boardContainer) {
            boardContainer.innerHTML = '<div class="board-loading">Loading board...</div>';
        }
    }

    showError(message) {
        const boardContainer = document.getElementById('drag-drop-board');
        if (boardContainer) {
            boardContainer.innerHTML = `
                <div class="board-error">
                    <h3>Error Loading Board</h3>
                    <p>${this.escapeHtml(message)}</p>
                    <button onclick="window.location.reload()" class="btn btn-primary">Try Again</button>
                </div>
            `;
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
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
}

// Initialize public board view if we're in public mode
document.addEventListener('DOMContentLoaded', () => {
    if (window.boardData && window.boardData.isPublic) {
        window.publicBoardView = new PublicBoardView();
    }
});