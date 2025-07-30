// Board view functionality - Updated to work with drag-drop board
class BoardView {
    constructor() {
        this.boardData = window.boardData || {};
        this.isAdmin = this.boardData.isAdmin || false;
        this.boardId = this.boardData.boardId;
        this.publicLink = this.boardData.publicLink;
        this.searchBar = null;
        this.originalIdeas = null; // Store original ideas for search filtering
        this.init();
    }

    init() {
        this.bindEvents();
        this.setupIdeaManager();
        this.setupWebSocket();
        this.setupSearchBar();
    }

    bindEvents() {
        // Refresh button
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshBoard();
            });
        }

        // Create idea button (admin only)
        const createIdeaBtn = document.getElementById('create-idea-btn');
        if (createIdeaBtn && this.isAdmin) {
            createIdeaBtn.addEventListener('click', () => {
                if (window.ideaManager) {
                    window.ideaManager.setBoardId(this.boardId);
                    window.ideaManager.openCreateModal();
                }
            });
        }

        // Board settings button (admin only)
        const settingsBtn = document.getElementById('board-settings-btn');
        if (settingsBtn && this.isAdmin) {
            settingsBtn.addEventListener('click', () => {
                if (window.boardSettingsManager) {
                    window.boardSettingsManager.setBoardId(this.boardId);
                    window.boardSettingsManager.openSettingsModal();
                }
            });
        }
    }

    setupIdeaManager() {
        // Set board ID for idea manager when it's available
        if (window.ideaManager) {
            window.ideaManager.setBoardId(this.boardId);
        } else {
            // Wait for idea manager to load
            const checkIdeaManager = setInterval(() => {
                if (window.ideaManager) {
                    window.ideaManager.setBoardId(this.boardId);
                    clearInterval(checkIdeaManager);
                }
            }, 100);
        }
    }

    setupWebSocket() {
        // Initialize WebSocket connection for real-time updates (both admin and public)
        if (this.boardId && window.WebSocketManager) {
            this.wsManager = new WebSocketManager(this.boardId);
            // Expose globally for retry functionality
            window.wsManager = this.wsManager;
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
        // Refresh the specific idea or the entire board
        console.log('Feedback updated for idea:', detail.ideaId);
        
        // If drag-drop board is available, refresh it
        if (window.dragDropBoard) {
            window.dragDropBoard.loadBoardData();
        } else {
            this.refreshBoard();
        }
    }

    handleIdeaUpdate(detail) {
        // Handle real-time idea updates
        console.log('Idea updated:', detail);
        
        // Refresh the board to show updates
        if (window.dragDropBoard) {
            window.dragDropBoard.loadBoardData();
        } else {
            this.refreshBoard();
        }
    }

    async refreshBoard() {
        if (window.dragDropBoard) {
            await window.dragDropBoard.loadBoard();
        }
    }

    // Method to refresh ideas (called by idea manager)
    async refreshIdeas() {
        if (window.dragDropBoard) {
            await window.dragDropBoard.refreshBoard();
        }
    }

    setupSearchBar() {
        // Only setup search bar for admin users
        if (!this.isAdmin || !this.boardId) {
            return;
        }

        // Initialize search bar with callback for handling search results
        this.searchBar = new SearchBar(this.boardId, (searchResults, searchInfo) => {
            this.handleSearchResults(searchResults, searchInfo);
        });
    }

    handleSearchResults(searchResults, searchInfo) {
        // Update the drag-drop board with filtered results
        if (window.dragDropBoard) {
            // Store original ideas if not already stored
            if (!this.originalIdeas && !searchInfo.query && !this.hasActiveFilters(searchInfo)) {
                this.originalIdeas = window.dragDropBoard.ideas;
            }

            // Apply search results to the board
            if (searchInfo.query || this.hasActiveFilters(searchInfo)) {
                // Show search results
                window.dragDropBoard.updateIdeasWithSearch(searchResults, searchInfo);
                this.highlightSearchResults(searchInfo.query);
            } else {
                // Restore original ideas when search is cleared
                if (this.originalIdeas) {
                    window.dragDropBoard.updateIdeasWithSearch(this.originalIdeas, null);
                    this.originalIdeas = null;
                }
                this.clearSearchHighlights();
            }
        }
    }

    hasActiveFilters(searchInfo) {
        if (!searchInfo || !searchInfo.filters) return false;
        
        return searchInfo.filters.column || 
               searchInfo.filters.status || 
               searchInfo.filters.inProgress !== null ||
               (searchInfo.sort && searchInfo.sort.by);
    }

    highlightSearchResults(query) {
        if (!query) return;

        // Add highlighting to idea cards
        const ideaCards = document.querySelectorAll('.idea-card');
        ideaCards.forEach(card => {
            const titleElement = card.querySelector('.idea-title');
            const descriptionElement = card.querySelector('.idea-description');
            const valueElement = card.querySelector('.idea-value');

            if (titleElement && this.searchBar) {
                const originalTitle = titleElement.dataset.originalText || titleElement.textContent;
                titleElement.dataset.originalText = originalTitle;
                titleElement.innerHTML = this.searchBar.highlightSearchTerms(originalTitle, query);
            }

            if (descriptionElement && this.searchBar) {
                const originalDesc = descriptionElement.dataset.originalText || descriptionElement.textContent;
                descriptionElement.dataset.originalText = originalDesc;
                descriptionElement.innerHTML = this.searchBar.highlightSearchTerms(originalDesc, query);
            }

            if (valueElement && this.searchBar) {
                const originalValue = valueElement.dataset.originalText || valueElement.textContent;
                valueElement.dataset.originalText = originalValue;
                valueElement.innerHTML = this.searchBar.highlightSearchTerms(originalValue, query);
            }
        });
    }

    clearSearchHighlights() {
        // Remove highlighting from idea cards
        const ideaCards = document.querySelectorAll('.idea-card');
        ideaCards.forEach(card => {
            const titleElement = card.querySelector('.idea-title');
            const descriptionElement = card.querySelector('.idea-description');
            const valueElement = card.querySelector('.idea-value');

            if (titleElement && titleElement.dataset.originalText) {
                titleElement.textContent = titleElement.dataset.originalText;
                delete titleElement.dataset.originalText;
            }

            if (descriptionElement && descriptionElement.dataset.originalText) {
                descriptionElement.textContent = descriptionElement.dataset.originalText;
                delete descriptionElement.dataset.originalText;
            }

            if (valueElement && valueElement.dataset.originalText) {
                valueElement.textContent = valueElement.dataset.originalText;
                delete valueElement.dataset.originalText;
            }
        });
    }
}

// Initialize board view
document.addEventListener('DOMContentLoaded', () => {
    window.boardView = new BoardView();
});