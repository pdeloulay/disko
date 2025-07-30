// Board Settings Modal Component
// This module provides UI components for managing board settings including column visibility

class BoardSettingsManager {
    constructor() {
        this.currentBoardId = null;
        this.currentBoard = null;
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
    }

    setBoard(board) {
        this.currentBoard = board;
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
            <div id="board-settings-modal" class="modal" style="display: flex;">
                <div class="modal-content board-settings-modal">
                    <div class="modal-header">
                        <h3>Board Settings</h3>
                        <button class="modal-close">&times;</button>
                    </div>
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

                        <div class="settings-preview">
                            <h4>Preview</h4>
                            <p class="section-description">This is how your board will appear to public users:</p>
                            <div class="preview-container">
                                <div class="preview-columns" id="settings-preview">
                                    <!-- Preview will be generated here -->
                                </div>
                            </div>
                        </div>
                        
                        <div class="form-actions">
                            <button type="button" class="btn btn-secondary" onclick="boardSettingsManager.closeModal()">Cancel</button>
                            <button type="submit" class="btn btn-primary">Save Settings</button>
                        </div>
                    </form>
                </div>
            </div>
        `;
    }

    // Open settings modal
    async openSettingsModal() {
        try {
            // Fetch current board settings if not available
            if (!this.currentBoard) {
                const response = await window.api.get(`/boards/${this.currentBoardId}`);
                this.currentBoard = response.data || response;
            }

            // Remove existing modal if any
            const existingModal = document.getElementById('board-settings-modal');
            if (existingModal) {
                existingModal.remove();
            }

            // Add modal to page
            document.body.insertAdjacentHTML('beforeend', this.createBoardSettingsModal(this.currentBoard));
            
            // Setup event listeners for preview updates
            this.setupPreviewUpdates();
            
            // Generate initial preview
            this.updatePreview();
            
        } catch (error) {
            console.error('Failed to load board settings:', error);
            this.showErrorMessage('Failed to load board settings. Please try again.');
        }
    }

    closeModal() {
        const modal = document.getElementById('board-settings-modal');
        if (modal) {
            modal.remove();
        }
    }

    setupPreviewUpdates() {
        const modal = document.getElementById('board-settings-modal');
        if (!modal) return;

        // Listen for changes to checkboxes
        const checkboxes = modal.querySelectorAll('input[type="checkbox"]');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                this.updatePreview();
            });
        });
    }

    updatePreview() {
        const modal = document.getElementById('board-settings-modal');
        const previewContainer = document.getElementById('settings-preview');
        if (!modal || !previewContainer) return;

        // Get selected columns and fields
        const selectedColumns = Array.from(modal.querySelectorAll('input[name="visibleColumns"]:checked'))
            .map(cb => cb.value);
        const selectedFields = Array.from(modal.querySelectorAll('input[name="visibleFields"]:checked'))
            .map(cb => cb.value);

        // Generate preview
        const columnTitles = {
            'parking': 'Parking',
            'now': 'Now',
            'next': 'Next',
            'later': 'Later',
            'release': 'Release',
            'wont-do': "Won't Do"
        };

        const previewHtml = selectedColumns.length > 0 ? `
            <div class="preview-board">
                ${selectedColumns.map(columnId => `
                    <div class="preview-column">
                        <div class="preview-column-header">
                            <h5>${columnTitles[columnId]}</h5>
                            <span class="preview-count">0</span>
                        </div>
                        <div class="preview-idea-card">
                            <div class="preview-idea-header">
                                <h6>Sample Idea</h6>
                            </div>
                            <div class="preview-idea-content">
                                ${selectedFields.includes('description') ? '<p class="preview-description">This is a sample description...</p>' : ''}
                                ${selectedFields.includes('valueStatement') ? '<p class="preview-value"><strong>Value:</strong> Sample value statement...</p>' : ''}
                            </div>
                            <div class="preview-feedback">
                                <span class="preview-thumbs">👍 0</span>
                                <span class="preview-emoji">😊 0</span>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        ` : '<p class="preview-empty">No columns selected. Public users will not see any content.</p>';

        previewContainer.innerHTML = previewHtml;
    }

    // Handle settings form submission
    async handleUpdateSettings(e) {
        const formData = new FormData(e.target);
        
        // Get selected columns and fields
        const visibleColumns = formData.getAll('visibleColumns');
        const visibleFields = formData.getAll('visibleFields');

        // Ensure oneLiner is always included
        if (!visibleFields.includes('oneLiner')) {
            visibleFields.push('oneLiner');
        }

        const settingsData = {
            visibleColumns,
            visibleFields
        };

        try {
            const submitBtn = e.target.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Saving...';

            const response = await window.api.put(`/boards/${this.currentBoardId}`, settingsData);
            
            // Update current board data
            this.currentBoard = response.data || response;
            
            this.closeModal();
            this.showSuccessMessage('Board settings updated successfully!');
            
            // Trigger refresh of board view
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
            // Trigger refresh of drag-drop board
            if (window.dragDropBoard && window.dragDropBoard.loadBoard) {
                await window.dragDropBoard.loadBoard();
            }
            
        } catch (error) {
            console.error('Failed to update board settings:', error);
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