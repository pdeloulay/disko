// Board view functionality - Updated to work with drag-drop board
class BoardView {
    constructor() {
        this.boardData = window.boardData || {};
        this.isAdmin = this.boardData.isAdmin || false;
        this.boardId = this.boardData.boardId;
        this.publicLink = this.boardData.publicLink;
        this.searchBar = null;
        this.originalIdeas = null; // Store original ideas for search filtering
        
        console.log('[BoardView] Constructor called - BoardID:', this.boardId, 'IsAdmin:', this.isAdmin, 'PublicLink:', this.publicLink);
        console.log('[BoardView] Board data:', this.boardData);
        
        this.init();
    }

    init() {
        console.log('[BoardView] Initializing board view...');
        this.bindEvents();
        this.setupIdeaManager();
        this.setupWebSocket();
        this.setupSearchBar();
        console.log('[BoardView] Board view initialization complete');
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
        console.log('[BoardView] Setting up idea manager...');
        // Set board ID for idea manager when it's available
        if (window.ideaManager) {
            console.log('[BoardView] Idea manager available, setting board ID');
            window.ideaManager.setBoardId(this.boardId);
        } else {
            console.log('[BoardView] Idea manager not available, waiting...');
            // Wait for idea manager to load
            const checkIdeaManager = setInterval(() => {
                if (window.ideaManager) {
                    console.log('[BoardView] Idea manager now available, setting board ID');
                    window.ideaManager.setBoardId(this.boardId);
                    clearInterval(checkIdeaManager);
                }
            }, 100);
        }
    }

    setupWebSocket() {
        console.log('[BoardView] Setting up WebSocket...');
        // Initialize WebSocket connection for real-time updates (both admin and public)
        if (this.boardId && window.WebSocketManager) {
            console.log('[BoardView] WebSocket manager available, initializing connection');
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
            console.log('[BoardView] WebSocket setup complete');
        } else {
            console.log('[BoardView] WebSocket setup skipped - BoardID:', this.boardId, 'WebSocketManager available:', !!window.WebSocketManager);
        }
    }

    handleFeedbackUpdate(detail) {
        // Refresh the specific idea or the entire board
        console.log('[BoardView] Feedback updated for idea:', detail.ideaId);
        
        // If drag-drop board is available, refresh it
        if (window.dragDropBoard) {
            console.log('[BoardView] Refreshing drag-drop board for feedback update');
            window.dragDropBoard.loadBoardData();
        } else {
            console.log('[BoardView] Refreshing board for feedback update');
            this.refreshBoard();
        }
    }

    handleIdeaUpdate(detail) {
        // Handle real-time idea updates
        console.log('[BoardView] Idea updated:', detail);
        
        // Refresh the board to show updates
        if (window.dragDropBoard) {
            console.log('[BoardView] Refreshing drag-drop board for idea update');
            window.dragDropBoard.loadBoardData();
        } else {
            console.log('[BoardView] Refreshing board for idea update');
            this.refreshBoard();
        }
    }

    async refreshBoard() {
        console.log('[BoardView] Refreshing board...');
        if (window.dragDropBoard) {
            console.log('[BoardView] Using drag-drop board to refresh');
            await window.dragDropBoard.loadBoard();
        } else {
            console.log('[BoardView] No drag-drop board available for refresh');
        }
        console.log('[BoardView] Board refresh complete');
    }

    // Method to check if drag-drop board is available
    checkDragDropBoard() {
        if (window.dragDropBoard) {
            console.log('[BoardView] Drag-drop board is available');
            return true;
        } else {
            console.log('[BoardView] Drag-drop board is not available');
            return false;
        }
    }

    // Method to refresh ideas (called by idea manager)
    async refreshIdeas() {
        console.log('[BoardView] Refreshing ideas...');
        if (window.dragDropBoard) {
            console.log('[BoardView] Using drag-drop board to refresh ideas');
            await window.dragDropBoard.refreshBoard();
        } else {
            console.log('[BoardView] No drag-drop board available for idea refresh');
        }
        console.log('[BoardView] Ideas refresh complete');
    }

    setupSearchBar() {
        console.log('[BoardView] Setting up search bar...');
        // Only setup search bar for admin users
        if (!this.isAdmin || !this.boardId) {
            console.log('[BoardView] Skipping search bar setup - IsAdmin:', this.isAdmin, 'BoardID:', this.boardId);
            return;
        }

        console.log('[BoardView] Initializing search bar for admin user');
        // Initialize search bar with callback for handling search results
        this.searchBar = new SearchBar(this.boardId, (searchResults, searchInfo) => {
            this.handleSearchResults(searchResults, searchInfo);
        });
        console.log('[BoardView] Search bar setup complete');
    }

    handleSearchResults(searchResults, searchInfo) {
        console.log('[BoardView] Handling search results:', searchResults?.length || 0, 'ideas, SearchInfo:', searchInfo);
        
        // Update the drag-drop board with filtered results
        if (window.dragDropBoard) {
            // Store original ideas if not already stored
            if (!this.originalIdeas && !searchInfo.query && !this.hasActiveFilters(searchInfo)) {
                console.log('[BoardView] Storing original ideas for search restoration');
                this.originalIdeas = window.dragDropBoard.ideas;
            }

            // Apply search results to the board
            if (searchInfo.query || this.hasActiveFilters(searchInfo)) {
                console.log('[BoardView] Applying search results to board');
                // Show search results
                window.dragDropBoard.updateIdeasWithSearch(searchResults, searchInfo);
                this.highlightSearchResults(searchInfo.query);
            } else {
                console.log('[BoardView] Restoring original ideas from search');
                // Restore original ideas when search is cleared
                if (this.originalIdeas) {
                    window.dragDropBoard.updateIdeasWithSearch(this.originalIdeas, null);
                    this.originalIdeas = null;
                }
                this.clearSearchHighlights();
            }
        } else {
            console.log('[BoardView] No drag-drop board available for search results');
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
    console.log('[BoardView] DOM loaded, initializing board view...');
    window.boardView = new BoardView();
    console.log('[BoardView] Board view initialized and available as window.boardView');
});