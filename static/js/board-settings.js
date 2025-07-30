// Board Settings Modal Component
// This module provides UI components for managing board settings including column visibility

class BoardSettingsManager {
    constructor() {
        this.currentBoardId = null;
        this.currentBoard = null;
        console.log('[BoardSettings] Constructor called');
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // Settings form submission
        document.addEventListener('submit', (e) => {
            if (e.target.id === 'board-settings-form') {
                e.preventDefault();
                this.handleUpdateSettings(e);
            }
        });

        // Modal close events
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-close') || 
                (e.target.classList.contains('modal') && e.target.id === 'board-settings-modal')) {
                this.closeModal();
            }
        });

        // Escape key to close modal
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModal();
            }
        });
    }

    setBoardId(boardId) {
        this.currentBoardId = boardId;
        console.log('[BoardSettings] Board ID set to:', boardId);
    }

    setBoard(board) {
        this.currentBoard = board;
        console.log('[BoardSettings] Board data set:', board);
    }

    // Board Settings Modal Component
    createBoardSettingsModal(board) {
        const allColumns = [
            { id: 'parking', title: 'Parking', description: 'Ideas waiting to be prioritized' },
            { id: 'now', title: 'Now', description: 'Currently working on' },
            { id: 'next', title: 'Next', description: 'Up next in the pipeline' },
            { id: 'later', title: 'Later', description: 'Future considerations' },
            { id: 'release', title: 'Release', description: 'Completed and released' },
            { id: 'wont-do', title: "Won't Do", description: 'Decided not to pursue' }
        ];

        const allFields = [
            { id: 'oneLiner', title: 'One-liner', description: 'Brief description (always visible)' },
            { id: 'description', title: 'Description', description: 'Detailed description' },
            { id: 'valueStatement', title: 'Value Statement', description: 'Value proposition' },
            { id: 'riceScore', title: 'RICE Score', description: 'Priority scoring (admin only)' }
        ];

        return `
            <div id="board-settings-modal" class="modal show">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Board Settings</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="board-settings-form">
                            <div class="settings-section">
                                <h4>Column Visibility</h4>
                                <p class="section-description">Choose which columns are visible to public users on your board.</p>
                                <div class="column-visibility-grid">
                                    ${allColumns.map(column => `
                                        <div class="column-visibility-item">
                                            <label class="checkbox-label">
                                                <input type="checkbox" 
                                                       name="visibleColumns" 
                                                       value="${column.id}"
                                                       ${board.visibleColumns.includes(column.id) ? 'checked' : ''}>
                                                <span class="checkbox-custom"></span>
                                                <div class="column-info">
                                                    <strong>${column.title}</strong>
                                                    <small>${column.description}</small>
                                                </div>
                                            </label>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>

                            <div class="settings-section">
                                <h4>Field Visibility</h4>
                                <p class="section-description">Choose which idea fields are visible to public users.</p>
                                <div class="field-visibility-grid">
                                    ${allFields.map(field => `
                                        <div class="field-visibility-item">
                                            <label class="checkbox-label ${field.id === 'oneLiner' ? 'disabled' : ''}">
                                                <input type="checkbox" 
                                                       name="visibleFields" 
                                                       value="${field.id}"
                                                       ${board.visibleFields.includes(field.id) ? 'checked' : ''}
                                                       ${field.id === 'oneLiner' ? 'checked disabled' : ''}>
                                                <span class="checkbox-custom"></span>
                                                <div class="field-info">
                                                    <strong>${field.title}</strong>
                                                    <small>${field.description}</small>
                                                </div>
                                            </label>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                            
                            <div class="form-actions">
                                <button type="button" class="btn btn-secondary" onclick="boardSettingsManager.closeModal()">Cancel</button>
                                <button type="submit" class="btn btn-primary">Save Settings</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
    }

    // Open settings modal
    async openSettingsModal() {
        console.log('[BoardSettings] Opening settings modal...');
        console.log('[BoardSettings] Current board ID:', this.currentBoardId);
        console.log('[BoardSettings] Current board data:', this.currentBoard);
        
        try {
            // Fetch current board settings if not available
            if (!this.currentBoard) {
                console.log('[BoardSettings] Fetching board data...');
                const response = await window.api.get(`/boards/${this.currentBoardId}`);
                this.currentBoard = response.data || response;
                console.log('[BoardSettings] Board data fetched:', this.currentBoard);
            }

            // Remove existing modal if any
            const existingModal = document.getElementById('board-settings-modal');
            if (existingModal) {
                console.log('[BoardSettings] Removing existing modal');
                existingModal.remove();
            }

            // Add modal to page
            console.log('[BoardSettings] Creating modal HTML...');
            const modalHtml = this.createBoardSettingsModal(this.currentBoard);
            console.log('[BoardSettings] Modal HTML created, length:', modalHtml.length);
            console.log('[BoardSettings] Modal HTML preview:', modalHtml.substring(0, 200) + '...');
            document.body.insertAdjacentHTML('beforeend', modalHtml);
            console.log('[BoardSettings] Modal added to page');
            
            // Verify modal was added
            const modal = document.getElementById('board-settings-modal');
            console.log('[BoardSettings] Modal element found:', !!modal);
            if (modal) {
                console.log('[BoardSettings] Modal classes:', modal.className);
                console.log('[BoardSettings] Modal computed style:', window.getComputedStyle(modal).display);
                console.log('[BoardSettings] Modal z-index:', window.getComputedStyle(modal).zIndex);
            }
            
            // Setup event listeners for form updates
            this.setupFormUpdates();
            console.log('[BoardSettings] Form updates setup complete');
            
        } catch (error) {
            console.error('[BoardSettings] Failed to load board settings:', error);
            this.showErrorMessage('Failed to load board settings. Please try again.');
        }
    }

    closeModal() {
        const modal = document.getElementById('board-settings-modal');
        if (modal) {
            modal.remove();
        }
    }

    setupFormUpdates() {
        const modal = document.getElementById('board-settings-modal');
        if (!modal) return;

        // Listen for changes to checkboxes
        const checkboxes = modal.querySelectorAll('input[type="checkbox"]');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                console.log('[BoardSettings] Form field updated:', checkbox.name, checkbox.value, checkbox.checked);
            });
        });
    }

    // Handle settings form submission
    async handleUpdateSettings(e) {
        console.log('[BoardSettings] Form submission started');
        
        const formData = new FormData(e.target);
        
        // Get selected columns and fields
        const visibleColumns = formData.getAll('visibleColumns');
        const visibleFields = formData.getAll('visibleFields');

        console.log('[BoardSettings] Form data - VisibleColumns:', visibleColumns);
        console.log('[BoardSettings] Form data - VisibleFields:', visibleFields);

        // Ensure oneLiner is always included
        if (!visibleFields.includes('oneLiner')) {
            visibleFields.push('oneLiner');
            console.log('[BoardSettings] Added oneLiner to visible fields');
        }

        const settingsData = {
            visibleColumns,
            visibleFields
        };

        console.log('[BoardSettings] Settings data to send:', settingsData);
        console.log('[BoardSettings] Current board ID:', this.currentBoardId);

        try {
            const submitBtn = e.target.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Saving...';

            console.log('[BoardSettings] Making API call to PUT /boards/' + this.currentBoardId);
            const response = await window.api.put(`/boards/${this.currentBoardId}`, settingsData);
            
            console.log('[BoardSettings] API response received:', response);
            
            // Update current board data
            this.currentBoard = response.data || response;
            
            console.log('[BoardSettings] Updated current board data:', this.currentBoard);
            
            this.closeModal();
            this.showSuccessMessage('Board settings updated successfully!');
            
            console.log('[BoardSettings] Triggering board refresh...');
            
            // Trigger refresh of board view
            if (window.boardView && window.boardView.refreshIdeas) {
                console.log('[BoardSettings] Refreshing board view...');
                await window.boardView.refreshIdeas();
            }
            
            // Trigger refresh of drag-drop board
            if (window.dragDropBoard && window.dragDropBoard.loadBoard) {
                console.log('[BoardSettings] Refreshing drag-drop board...');
                await window.dragDropBoard.loadBoard();
            }
            
            console.log('[BoardSettings] Board settings update completed successfully');
            
        } catch (error) {
            console.error('[BoardSettings] Failed to update board settings:', error);
            console.error('[BoardSettings] Error details:', {
                message: error.message,
                status: error.response?.status,
                data: error.response?.data
            });
            this.handleFormError(error);
        } finally {
            const submitBtn = e.target.querySelector('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Save Settings';
            }
        }
    }

    handleFormError(error) {
        let errorMessage = 'Failed to update settings. Please try again.';
        
        if (error.response && error.response.data && error.response.data.error) {
            errorMessage = error.response.data.error.message || errorMessage;
        }
        
        this.showErrorMessage(errorMessage);
    }

    // Utility methods
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

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize board settings manager
document.addEventListener('DOMContentLoaded', () => {
    window.boardSettingsManager = new BoardSettingsManager();
});