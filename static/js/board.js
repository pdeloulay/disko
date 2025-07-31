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
        // loadBoardData() will be called after Clerk is initialized
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
        // Remove existing event listeners to prevent duplicates
        this.removeEventListeners();
        
        // Refresh button
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshBoard();
            });
        }

        // Create idea button (admin only)
        const createIdeaBtn = document.getElementById('create-idea-btn');
        console.log('[BoardView] Create idea button found:', !!createIdeaBtn, 'IsAdmin:', this.isAdmin);
        if (createIdeaBtn && this.isAdmin) {
            console.log('[BoardView] Adding create idea button event listener');
            createIdeaBtn.addEventListener('click', () => {
                console.log('[BoardView] Create idea button clicked');
                console.log('[BoardView] Idea manager available:', !!window.ideaManager);
                if (window.ideaManager) {
                    window.ideaManager.setBoardId(this.boardId);
                    console.log('[BoardView] About to open create idea modal...');
                    window.ideaManager.openCreateModal();
                    console.log('[BoardView] Create idea modal opened');
                } else {
                    console.error('[BoardView] Idea manager not available');
                }
            });
        } else {
            console.log('[BoardView] Create idea button not found or user not admin - Button exists:', !!createIdeaBtn, 'IsAdmin:', this.isAdmin);
        }

        // Board settings button (admin only)
        const settingsBtn = document.getElementById('board-settings-btn');
        if (settingsBtn && this.isAdmin) {
            console.log('[BoardView] Adding settings button event listener - IsAdmin:', this.isAdmin);
            settingsBtn.addEventListener('click', async () => {
                console.log('[BoardView] Settings button clicked');
                console.log('[BoardView] Board settings manager available:', !!window.boardSettingsManager);
                console.log('[BoardView] Current board ID:', this.boardId);
                
                if (window.boardSettingsManager) {
                    window.boardSettingsManager.setBoardId(this.boardId);
                    console.log('[BoardView] About to open settings modal...');
                    await window.boardSettingsManager.openSettingsModal();
                    console.log('[BoardView] Settings modal opened');
                } else {
                    console.error('[BoardView] Board settings manager not available');
                }
            });
        } else {
            console.log('[BoardView] Settings button not found or user not admin - Button exists:', !!settingsBtn, 'IsAdmin:', this.isAdmin);
        }

        // Publish button (admin only)
        const publishBtn = document.getElementById('publish-btn');
        if (publishBtn && this.isAdmin) {
            console.log('[BoardView] Adding publish button event listener - IsAdmin:', this.isAdmin);
            publishBtn.addEventListener('click', async () => {
                console.log('[BoardView] Publish button clicked');
                await this.publishBoard();
            });
        } else {
            console.log('[BoardView] Publish button not found or user not admin - Button exists:', !!publishBtn, 'IsAdmin:', this.isAdmin);
        }

        // Invite button (admin only, enabled only when board is published)
        const inviteBtn = document.getElementById('invite-btn');
        if (inviteBtn && this.isAdmin) {
            console.log('[BoardView] Adding invite button event listener - IsAdmin:', this.isAdmin);
            inviteBtn.addEventListener('click', () => {
                console.log('[BoardView] Invite button clicked');
                this.openInviteModal();
            });
        } else {
            console.log('[BoardView] Invite button not found or user not admin - Button exists:', !!inviteBtn, 'IsAdmin:', this.isAdmin);
        }

        // Invite form submission
        const inviteForm = document.getElementById('invite-form');
        if (inviteForm) {
            inviteForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                await this.sendInvite();
            });
        }
    }

    removeEventListeners() {
        // Remove existing event listeners to prevent duplicates
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.replaceWith(refreshBtn.cloneNode(true));
        }

        const createIdeaBtn = document.getElementById('create-idea-btn');
        if (createIdeaBtn) {
            createIdeaBtn.replaceWith(createIdeaBtn.cloneNode(true));
        }

        const settingsBtn = document.getElementById('board-settings-btn');
        if (settingsBtn) {
            settingsBtn.replaceWith(settingsBtn.cloneNode(true));
        }

        const publishBtn = document.getElementById('publish-btn');
        if (publishBtn) {
            publishBtn.replaceWith(publishBtn.cloneNode(true));
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
            window.dragDropBoard.loadBoard();
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
            window.dragDropBoard.loadBoard();
        } else {
            console.log('[BoardView] Refreshing board for idea update');
            this.refreshBoard();
        }
    }

    async loadBoardData() {
        console.log('[BoardView] Loading board data...');
        
        // Wait for Clerk to be available
        if (!window.Clerk) {
            console.log('[BoardView] Clerk not available, waiting...');
            await new Promise(resolve => {
                const checkClerk = setInterval(() => {
                    if (window.Clerk) {
                        clearInterval(checkClerk);
                        resolve();
                    }
                }, 100);
            });
        }
        
        try {
            const authToken = await this.getAuthToken();
            if (!authToken) {
                console.error('[BoardView] No auth token available');
                return;
            }
            
            const response = await fetch(`/api/boards/${this.boardId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${authToken}`
                }
            });

            if (response.ok) {
                const boardData = await response.json();
                console.log('[BoardView] Board data loaded:', boardData);
                
                // Update board data
                this.boardData = boardData;
                this.isAdmin = boardData.isAdmin || false;
                this.publicLink = boardData.publicLink;
                
                // Update window.boardData for other components
                window.boardData = this.boardData;
                
                console.log('[BoardView] Updated board data - IsAdmin:', this.isAdmin, 'PublicLink:', this.publicLink);
                
                // Update invite button state based on board publication status
                if (this.isAdmin) {
                    const isPublished = boardData.isPublic || boardData.publicLink;
                    this.updateInviteButtonState(isPublished);
                    this.updatePublishButtonState(isPublished);
                }
                
                // Re-bind events with updated admin status
                this.bindEvents();
            } else {
                console.error('[BoardView] Failed to load board data:', response.status);
            }
        } catch (error) {
            console.error('[BoardView] Error loading board data:', error);
        }
    }

    async getAuthToken() {
        if (!window.Clerk) {
            console.log('[BoardView] Clerk not available for auth token');
            return null;
        }
        
        try {
            // Wait for Clerk to be fully loaded
            if (!window.Clerk.session) {
                console.log('[BoardView] Waiting for Clerk session...');
                await new Promise(resolve => {
                    const checkSession = setInterval(() => {
                        if (window.Clerk.session) {
                            clearInterval(checkSession);
                            resolve();
                        }
                    }, 100);
                });
            }
            
            const token = await window.Clerk.session.getToken();
            console.log('[BoardView] Auth token obtained successfully');
            return token;
        } catch (error) {
            console.error('[BoardView] Error getting auth token:', error);
            return null;
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

    async publishBoard() {
        console.log('[BoardView] Toggling board publish state...');
        
        const publishBtn = document.getElementById('publish-btn');
        if (!publishBtn) {
            console.error('[BoardView] Publish button not found');
            return;
        }

        // Get current publish state
        const isCurrentlyPublished = publishBtn.getAttribute('data-published') === 'true';
        const newPublishState = !isCurrentlyPublished;
        
        console.log('[BoardView] Current publish state:', isCurrentlyPublished, 'New state:', newPublishState);

        // Store original button state
        const originalText = publishBtn.textContent;
        const originalDisabled = publishBtn.disabled;

        try {
            // Update button state
            publishBtn.disabled = true;
            publishBtn.textContent = newPublishState ? '🔄 Publishing...' : '🔄 Unpublishing...';

            console.log('[BoardView] Making API call to toggle board public state');
            const response = await window.api.put(`/boards/${this.boardId}`, {
                isPublic: newPublishState
            });
            
            console.log('[BoardView] Toggle publish API response:', response);
            
            if (response) {
                // Update the board data
                if (window.boardData) {
                    window.boardData.publicLink = response.publicLink;
                    window.boardData.isPublic = newPublishState;
                }
                
                // Update button state and text
                publishBtn.setAttribute('data-published', newPublishState.toString());
                publishBtn.textContent = newPublishState ? '🔒 Unpublish' : '🌐 Publish';
                
                // Update button styling
                if (newPublishState) {
                    publishBtn.classList.remove('btn-primary');
                    publishBtn.classList.add('btn-warning');
                } else {
                    publishBtn.classList.remove('btn-warning');
                    publishBtn.classList.add('btn-primary');
                }
                
                // Update invite button state
                this.updateInviteButtonState(newPublishState);
                
                // Show success message
                if (newPublishState) {
                    this.showSuccessMessage(`Board published successfully! New public link: ${response.publicLink}`, response.publicLink);
                    console.log('[BoardView] Board published successfully with new public link:', response.publicLink);
                } else {
                    this.showSuccessMessage('Board unpublished successfully!');
                    console.log('[BoardView] Board unpublished successfully');
                }
            } else {
                throw new Error('No response received from API');
            }

        } catch (error) {
            console.error('[BoardView] Failed to toggle board publish state:', error);
            this.showErrorMessage(`Failed to ${newPublishState ? 'publish' : 'unpublish'} board. Please try again.`);
        } finally {
            // Restore button state
            publishBtn.disabled = originalDisabled;
            publishBtn.textContent = originalText;
        }
    }

    updateInviteButtonState(enabled) {
        const inviteBtn = document.getElementById('invite-btn');
        if (inviteBtn) {
            inviteBtn.disabled = !enabled;
            if (enabled) {
                inviteBtn.classList.remove('btn-secondary');
                inviteBtn.classList.add('btn-primary');
            } else {
                inviteBtn.classList.remove('btn-primary');
                inviteBtn.classList.add('btn-secondary');
            }
        }
    }

    updatePublishButtonState(isPublished) {
        const publishBtn = document.getElementById('publish-btn');
        if (publishBtn) {
            // Update data attribute
            publishBtn.setAttribute('data-published', isPublished.toString());
            
            // Update button text
            publishBtn.textContent = isPublished ? '🔒 Unpublish' : '🌐 Publish';
            
            // Update button styling
            if (isPublished) {
                publishBtn.classList.remove('btn-primary');
                publishBtn.classList.add('btn-warning');
            } else {
                publishBtn.classList.remove('btn-warning');
                publishBtn.classList.add('btn-primary');
            }
        }
    }

    openInviteModal() {
        const modal = document.getElementById('invite-modal');
        if (modal) {
            // Set default subject
            const subjectInput = document.getElementById('invite-subject');
            if (subjectInput) {
                subjectInput.value = `[Disko] 🚀 You're Invited to the ${this.boardData.name || 'board!'}`;
            }
            
            modal.classList.add('show');
        }
    }

    closeInviteModal() {
        const modal = document.getElementById('invite-modal');
        if (modal) {
            modal.classList.remove('show');
            // Reset form
            const form = document.getElementById('invite-form');
            if (form) {
                form.reset();
            }
        }
    }

    async sendInvite() {
        const emailInput = document.getElementById('invite-email');
        const subjectInput = document.getElementById('invite-subject');
        
        if (!emailInput || !subjectInput) {
            this.showErrorMessage('Form elements not found');
            return;
        }

        const email = emailInput.value.trim();
        const subject = subjectInput.value.trim();

        // Validation
        if (!email) {
            this.showErrorMessage('Email address is required');
            return;
        }

        if (!subject) {
            this.showErrorMessage('Subject is required');
            return;
        }

        // Basic email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            this.showErrorMessage('Please enter a valid email address');
            return;
        }

        try {
            console.log('[BoardView] Sending invite email to:', email);
            
            const response = await window.api.post(`/boards/${this.boardId}/invite`, {
                email: email,
                subject: subject
            });
            
            console.log('[BoardView] Invite API response:', response);
            
            if (response && response.success) {
                this.showSuccessMessage('Invitation email sent successfully!');
                this.closeInviteModal();
            } else {
                throw new Error('Failed to send invitation');
            }

        } catch (error) {
            console.error('[BoardView] Failed to send invite:', error);
            this.showErrorMessage('Failed to send invitation email. Please try again.');
        }
    }

    showSuccessMessage(message, publicLink = null) {
        // Create a temporary success message
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message-toast success show';
        
        let messageContent = message;
        
        // Add view link if public link is provided
        if (publicLink) {
            const viewUrl = `${window.location.origin}/public/${publicLink}`;
            messageContent += `<br><a href="${viewUrl}" target="_blank" class="toast-link">🔗 View Public Board</a>`;
        }
        
        messageDiv.innerHTML = messageContent;
        document.body.appendChild(messageDiv);

        // Remove after 6 seconds (longer duration)
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, 6000);
    }

    showErrorMessage(message) {
        // Create a temporary error message
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message-toast error show';
        messageDiv.textContent = message;
        document.body.appendChild(messageDiv);

        // Remove after 3 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, 3000);
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