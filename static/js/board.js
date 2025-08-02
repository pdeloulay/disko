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
        
        // Set initial status indicator based on available data
        if (this.boardData) {
            const isPublished = this.boardData.isPublic || this.boardData.publicLink;
            this.updateStatusIndicator(isPublished);
        }
        
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

        // Publish toggle (admin only) - will be handled in updateStatusIndicator
        console.log('[BoardView] Publish toggle will be handled in updateStatusIndicator - IsAdmin:', this.isAdmin);

        // Delete board button (admin only)
        const deleteBoardBtn = document.getElementById('delete-board-btn');
        if (deleteBoardBtn && this.isAdmin) {
            console.log('[BoardView] Adding delete board button event listener - IsAdmin:', this.isAdmin);
            deleteBoardBtn.addEventListener('click', () => {
                console.log('[BoardView] Delete board button clicked');
                this.openDeleteModal();
            });
        } else {
            console.log('[BoardView] Delete board button not found or user not admin - Button exists:', !!deleteBoardBtn, 'IsAdmin:', this.isAdmin);
        }

        // Invite form submission
        const inviteForm = document.getElementById('invite-form');
        if (inviteForm) {
            inviteForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                console.log('[BoardView] Invite form submitted, preventing duplicate submission');
                
                // Disable submit button to prevent multiple submissions
                const submitBtn = inviteForm.querySelector('button[type="submit"]');
                if (submitBtn) {
                    submitBtn.disabled = true;
                    submitBtn.textContent = 'Sending...';
                }
                
                try {
                    await this.sendInvite();
                } finally {
                    // Re-enable submit button
                    if (submitBtn) {
                        submitBtn.disabled = false;
                        submitBtn.textContent = '📧 Send Email';
                    }
                }
            });
        }
    }

    removeEventListeners() {
        // Remove refresh button listener
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.replaceWith(refreshBtn.cloneNode(true));
        }

        // Remove create idea button listener
        const createIdeaBtn = document.getElementById('create-idea-btn');
        if (createIdeaBtn) {
            createIdeaBtn.replaceWith(createIdeaBtn.cloneNode(true));
        }

        // Remove settings button listener
        const settingsBtn = document.getElementById('board-settings-btn');
        if (settingsBtn) {
            settingsBtn.replaceWith(settingsBtn.cloneNode(true));
        }

        // Remove delete board button listener
        const deleteBoardBtn = document.getElementById('delete-board-btn');
        if (deleteBoardBtn) {
            deleteBoardBtn.replaceWith(deleteBoardBtn.cloneNode(true));
        }

        // Remove invite form listener
        const inviteForm = document.getElementById('invite-form');
        if (inviteForm) {
            inviteForm.replaceWith(inviteForm.cloneNode(true));
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
                const isPublished = boardData.isPublic || boardData.publicLink;
                
                if (this.isAdmin) {
                    this.updatePublishButtonState(isPublished);
                } else {
                    // For non-admin users, still show the status indicator
                    this.updateStatusIndicator(isPublished);
                }
                
                // Always update status indicator for all users
                this.updateStatusIndicator(isPublished);
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

    async publishBoard(isPublished = null) {
        console.log('[BoardView] Toggling board publish state...');
        
        const publishToggle = document.getElementById('publish-toggle');
        if (!publishToggle) {
            console.error('[BoardView] Publish toggle not found');
            return;
        }

        // Get current publish state
        const isCurrentlyPublished = publishToggle.checked;
        const newPublishState = isPublished !== null ? isPublished : !isCurrentlyPublished;
        
        console.log('[BoardView] Current publish state:', isCurrentlyPublished, 'New state:', newPublishState);

        // Store original toggle state
        const originalChecked = publishToggle.checked;

        try {
            // Add loading state
            publishToggle.disabled = true;
            publishToggle.parentElement.classList.add('loading');

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
                
                // Update toggle state
                publishToggle.checked = newPublishState;
                console.log('[BoardView] Updated publish toggle state to:', publishToggle.checked);
                
                // Update invite button state
                this.updatePublishButtonState(newPublishState);
                
                // Update status indicator with the new public link
                this.updateStatusIndicator(newPublishState, response.publicLink);
                
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
            
            // Restore toggle state only on error
            publishToggle.checked = originalChecked;
        } finally {
            // Remove loading state
            publishToggle.disabled = false;
            publishToggle.parentElement.classList.remove('loading');
        }
    }

    updatePublishButtonState(isPublished) {
        const publishToggle = document.getElementById('publish-toggle');
        if (publishToggle) {
            console.log('[BoardView] updatePublishButtonState called - IsPublished:', isPublished);
            
            // Update toggle state
            publishToggle.checked = isPublished;
            
            // Update data attribute
            publishToggle.setAttribute('data-published', isPublished.toString());
            
            console.log('[BoardView] Updated publish toggle state to:', publishToggle.checked);
        } else {
            console.error('[BoardView] Publish toggle not found in updatePublishButtonState');
        }

        // Update status indicator
        this.updateStatusIndicator(isPublished);
    }

    updateStatusIndicator(isPublished, publicLink = null) {
        const statusContainer = document.getElementById('board-status');
        if (!statusContainer) return;

        // Use the provided publicLink or fall back to boardData
        const linkToUse = publicLink || (this.boardData ? this.boardData.publicLink : null);
        
        console.log('[BoardView] Updating status indicator - IsPublished:', isPublished, 'PublicLink:', linkToUse, 'BoardData:', this.boardData);

        if (isPublished && linkToUse) {
            // Create public status indicator with toggle and view icon
            statusContainer.innerHTML = `
                <div class="status-indicator public">
                    <span class="toggle-label">Private</span>
                    <label class="toggle-switch">
                        <input type="checkbox" id="publish-toggle" data-published="true" checked>
                        <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">Public</span>
                    <a href="/public/${linkToUse}" target="_blank" class="view-link-btn" title="View public board">View</a>
                    <button class="share-link-btn" onclick="boardView.openInviteModal()" title="Share board via email">Share</button>
                </div>
            `;
        } else if (isPublished) {
            // Create public status indicator with toggle (fallback)
            statusContainer.innerHTML = `
                <div class="status-indicator public">
                    <span class="toggle-label">Private</span>
                    <label class="toggle-switch">
                        <input type="checkbox" id="publish-toggle" data-published="true" checked>
                        <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">Public</span>
                    <button class="share-link-btn" onclick="boardView.openInviteModal()" title="Share board via email">Share</button>
                </div>
            `;
        } else {
            // Create private status indicator with toggle
            statusContainer.innerHTML = `
                <div class="status-indicator private">
                    <span class="toggle-label">Private</span>
                    <label class="toggle-switch">
                        <input type="checkbox" id="publish-toggle" data-published="false">
                        <span class="toggle-slider"></span>
                    </label>
                    <span class="toggle-label">Public</span>
                </div>
            `;
        }

        // Re-attach event listener to the new toggle
        const publishToggle = document.getElementById('publish-toggle');
        if (publishToggle && this.isAdmin) {
            publishToggle.addEventListener('change', async (e) => {
                console.log('[BoardView] Publish toggle changed:', e.target.checked);
                await this.publishBoard(e.target.checked);
            });
        }
    }

    openInviteModal() {
        const modal = document.getElementById('invite-modal');
        if (modal) {
            // Set default subject with board name and public link
            const subjectInput = document.getElementById('invite-subject');
            const boardName = this.boardData?.name || 'my board';
            const publicLink = this.boardData?.publicLink;
            
            if (subjectInput) {
                if (publicLink) {
                    subjectInput.value = `[Disko] 🚀 You're invited to view ${boardName}`;
                } else {
                    subjectInput.value = `[Disko] 🚀 You're invited to ${boardName}`;
                }
            }
            
            // Pre-fill email body with public link if available
            const emailInput = document.getElementById('invite-email');
            if (emailInput) {
                emailInput.value = '';
                emailInput.placeholder = 'Enter recipient email address';
            }
            
            // Pre-fill message with public link
            const messageInput = document.getElementById('invite-message');
            if (messageInput && publicLink) {
                const appUrl = window.location.origin;
                const publicUrl = `${appUrl}/public/${publicLink}`;
                messageInput.value = `Hi! I'd like to share my board "${boardName}" with you. Let me know what you think!`;
            } else if (messageInput) {
                messageInput.value = '';
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
        console.log('[BoardView] sendInvite method called');
        
        const emailInput = document.getElementById('invite-email');
        const subjectInput = document.getElementById('invite-subject');
        const messageInput = document.getElementById('invite-message');
        
        if (!emailInput || !subjectInput) {
            console.error('[BoardView] Invite form inputs not found');
            return;
        }
        
        const emailTo = emailInput.value.trim();
        const subject = subjectInput.value.trim();
        const message = messageInput ? messageInput.value.trim() : '';
        
        if (!emailTo || !subject) {
            this.showErrorMessage('Please fill in all required fields.');
            return;
        }
        
        try {
            console.log('[BoardView] Sending invite for board:', this.boardId, 'Email:', emailTo, 'Subject:', subject);
            
            // Include public link in the request if available
            const inviteData = {    
                emailTo: emailTo,
                subject: subject,
                message: message,
                publicLink: this.boardData?.publicLink || null
            };
            
            console.log('[BoardView] Making API call to send invite');
            const response = await window.api.post(`/boards/${this.boardId}/invite`, inviteData);
            console.log('[BoardView] API call completed');
            
            if (response) {
                this.showSuccessMessage('Invitation sent successfully!');
                this.closeInviteModal();
                
                // Clear form
                emailInput.value = '';
                subjectInput.value = '';
                if (messageInput) {
                    messageInput.value = '';
                }
            } else {
                this.showErrorMessage('Failed to send invitation. Please try again.');
            }
        } catch (error) {
            console.error('[BoardView] Failed to send invite:', error);
            this.showErrorMessage('Failed to send invitation. Please try again.');
        }
    }

    openDeleteModal() {
        console.log('[BoardView] Opening delete modal for board:', this.boardId);
        
        // Get the board name to display in the confirmation
        const boardTitle = document.getElementById('board-title');
        const boardName = boardTitle ? boardTitle.textContent : 'this board';
        
        // Set the board name in the confirmation field
        const boardNameConfirm = document.getElementById('board-name-to-confirm');
        if (boardNameConfirm) {
            boardNameConfirm.textContent = boardName;
        }
        
        // Clear the confirmation input
        const confirmInput = document.getElementById('confirm-board-name');
        if (confirmInput) {
            confirmInput.value = '';
        }
        
        // Show the modal
        const modal = document.getElementById('delete-board-modal');
        if (modal) {
            modal.classList.add('show');
        }
        
        // Focus on the confirmation input
        if (confirmInput) {
            setTimeout(() => confirmInput.focus(), 100);
        }
        
        // Add form submission handler
        const form = document.getElementById('delete-board-form');
        if (form) {
            form.onsubmit = (e) => {
                e.preventDefault();
                this.confirmDeleteBoard();
            };
        }
    }

    closeDeleteModal() {
        console.log('[BoardView] Closing delete modal');
        
        const modal = document.getElementById('delete-board-modal');
        if (modal) {
            modal.classList.remove('show');
        }
        
        // Remove form submission handler
        const form = document.getElementById('delete-board-form');
        if (form) {
            form.onsubmit = null;
        }
    }

    async confirmDeleteBoard() {
        const confirmInput = document.getElementById('confirm-board-name');
        const boardNameConfirm = document.getElementById('board-name-to-confirm');
        
        if (!confirmInput || !boardNameConfirm) {
            console.error('[BoardView] Delete confirmation elements not found');
            return;
        }
        
        const enteredName = confirmInput.value.trim();
        const expectedName = boardNameConfirm.textContent.trim();
        
        if (enteredName !== expectedName) {
            this.showErrorMessage('Board name does not match. Please enter the exact board name to confirm deletion.');
            return;
        }
        
        try {
            console.log('[BoardView] Confirming board deletion for board:', this.boardId);
            
            const response = await window.api.delete(`/boards/${this.boardId}`);
            
            if (response) {
                this.showSuccessMessage('Board deleted successfully!');
                this.closeDeleteModal();
                
                // Redirect to dashboard after a short delay
                setTimeout(() => {
                    window.location.href = '/dashboard';
                }, 1500);
            } else {
                this.showErrorMessage('Failed to delete board. Please try again.');
            }
        } catch (error) {
            console.error('[BoardView] Failed to delete board:', error);
            this.showErrorMessage('Failed to delete board. Please try again.');
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