// Idea Management Components
// This module provides UI components for creating, editing, and managing ideas

class IdeaManager {
    constructor() {
        this.currentBoardId = null;
        this.editingIdeaId = null;
        console.log('[IdeaManager] Constructor called');
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // Create idea form submission
        document.addEventListener('submit', (e) => {
            console.log('[IdeaManager] Form submit event triggered:', e.target.id);
            if (e.target.id === 'create-idea-form') {
                console.log('[IdeaManager] Create idea form submitted');
                e.preventDefault();
                this.handleCreateIdea(e);
            }
            if (e.target.id === 'edit-idea-form') {
                console.log('[IdeaManager] Edit idea form submitted');
                e.preventDefault();
                this.handleEditIdea(e);
            }
        });

        // Modal close events
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-close') || e.target.classList.contains('modal')) {
                this.closeModals();
            }
        });

        // Keyboard events
        document.addEventListener('keydown', (e) => {
            // Escape key to close modals
            if (e.key === 'Escape') {
                this.closeModals();
            }
            
            // Enter key to submit form
            if (e.key === 'Enter' && !e.shiftKey) {
                const activeElement = document.activeElement;
                if (activeElement && activeElement.closest('.modal')) {
                    // Check if we're in a textarea (allow Shift+Enter for new lines)
                    if (activeElement.tagName === 'TEXTAREA') {
                        return; // Allow normal textarea behavior
                    }
                    
                    // Validate One-liner field before submitting
                    const modal = activeElement.closest('.modal');
                    const oneLinerField = modal.querySelector('input[name="oneLiner"]');
                    if (oneLinerField && oneLinerField.value.trim() === '') {
                        // One-liner is empty, don't submit
                        oneLinerField.focus();
                        return;
                    }
                    
                    // Find the submit button and trigger it
                    const submitBtn = modal.querySelector('button[type="submit"]');
                    if (submitBtn) {
                        console.log('[IdeaManager] Enter key - found submit button, clicking it');
                        e.preventDefault();
                        submitBtn.click();
                    } else {
                        console.log('[IdeaManager] Enter key - submit button not found');
                    }
                }
            }
        });
    }

    setBoardId(boardId) {
        this.currentBoardId = boardId;
        console.log('[IdeaManager] Board ID set to:', boardId);
    }

    // Create Idea Form Component
    createIdeaCreationForm() {
        return `
            <div id="create-idea-modal" class="modal show">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Create New Idea</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="create-idea-form">
                                                    <div class="form-group">
                            <label for="idea-oneliner">One-liner *</label>
                            <input type="text" id="idea-oneliner" name="oneLiner" required maxlength="200" 
                                   placeholder="Brief description of your idea">
                            <small class="form-help">Maximum 200 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="idea-description">Description</label>
                            <textarea id="idea-description" name="description" maxlength="1000" rows="4"
                                      placeholder="Detailed description of your idea (optional)"></textarea>
                            <small class="form-help">Maximum 1000 characters (optional)</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="idea-value-statement">Value Statement</label>
                            <textarea id="idea-value-statement" name="valueStatement" maxlength="500" rows="3"
                                      placeholder="What value does this idea provide? (optional)"></textarea>
                            <small class="form-help">Maximum 500 characters (optional)</small>
                        </div>
                            
                            <div class="rice-score-section">
                                <h4>RICE Score</h4>
                                <div class="rice-grid">
                                                                    <div class="form-group">
                                    <label for="rice-reach">Reach (%)</label>
                                    <input type="number" id="rice-reach" name="reach" min="0" max="100" 
                                           value="100" placeholder="0-100">
                                    <small class="form-help">Percentage of users affected</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-impact">Impact (%)</label>
                                    <input type="number" id="rice-impact" name="impact" min="0" max="100" 
                                           value="50" placeholder="0-100">
                                    <small class="form-help">Impact per user</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-confidence">Confidence (%)</label>
                                    <input type="number" id="rice-confidence" name="confidence" min="0" max="100" 
                                           value="50" placeholder="0-100">
                                    <small class="form-help">Confidence in the estimate</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-effort">Effort</label>
                                    <select id="rice-effort" name="effort">
                                        <option value="1" selected>1 - Low (hours)</option>
                                        <option value="3">3 - Medium (days)</option>
                                        <option value="8">8 - High (weeks)</option>
                                        <option value="21">21 - Very High (months)</option>
                                    </select>
                                    <small class="form-help">Development effort required</small>
                                </div>
                                </div>
                                <div class="rice-score-display">
                                    <span>RICE Score: <strong id="rice-score-value">0</strong></span>
                                </div>
                            </div>
                        </form>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="ideaManager.closeCreateModal()">Cancel</button>
                        <button type="submit" form="create-idea-form" class="btn btn-primary" onclick="console.log('[IdeaManager] Create button clicked')">Create Idea</button>
                    </div>
                </div>
            </div>
        `;
    }

    // Idea Card Component
    createIdeaCard(idea, isAdmin = false) {
        const riceScore = this.calculateRICEScore(idea.riceScore);
        const statusClass = idea.inProgress ? 'in-progress' : '';
        const animationClass = idea.inProgress ? 'pulse-animation' : '';
        
        return `
            <div class="idea-card ${statusClass} ${animationClass}" data-idea-id="${idea.id}">
                <div class="idea-card-header">
                    <h4 class="idea-oneliner">${this.escapeHtml(idea.oneLiner)}</h4>
                    ${isAdmin ? `
                        <div class="idea-card-menu">
                            <button class="btn-menu" onclick="ideaManager.toggleIdeaMenu('${idea.id}')">⋮</button>
                            <div class="idea-menu" id="idea-menu-${idea.id}" style="display: none;">
                                <button onclick="ideaManager.editIdea('${idea.id}')">✏️ Edit</button>
                                <button onclick="ideaManager.toggleInProgress('${idea.id}', ${!idea.inProgress})">
                                    ${idea.inProgress ? '⏸️ Mark as Not In Progress' : '▶️ Mark as In Progress'}
                                </button>
                                ${idea.status !== 'done' ? `
                                    <button onclick="ideaManager.updateIdeaStatus('${idea.id}', 'done')">✅ Mark as Done</button>
                                ` : ''}
                                ${idea.status === 'done' ? `
                                    <button onclick="ideaManager.updateIdeaStatus('${idea.id}', 'active')">🔄 Reactivate</button>
                                ` : ''}
                                ${idea.status !== 'archived' ? `
                                    <button onclick="ideaManager.updateIdeaStatus('${idea.id}', 'archived')">🗃️ Archive</button>
                                ` : ''}
                                <button onclick="ideaManager.confirmDeleteIdea('${idea.id}', '${this.escapeHtml(idea.oneLiner)}')">🗑️ Delete</button>
                            </div>
                        </div>
                    ` : ''}
                </div>
                
                <div class="idea-content">
                    <p class="idea-description">${this.escapeHtml(idea.description)}</p>
                    <p class="idea-value-statement"><strong>Value:</strong> ${this.escapeHtml(idea.valueStatement)}</p>
                </div>
                
                ${isAdmin ? `
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
                        <button class="feedback-btn thumbs-up" onclick="ideaManager.addThumbsUp('${idea.id}')">
                            👍 <span class="count">${idea.thumbsUp || 0}</span>
                        </button>
                        <button class="feedback-btn emoji" onclick="ideaManager.showEmojiPicker('${idea.id}')">
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
                    <span class="idea-date">Updated ${this.formatTimeAgo(idea.updatedAt)}</span>
                </div>
            </div>
        `;
    }

    // Idea Edit Modal Component
    createIdeaEditModal(idea) {
        return `
            <div id="edit-idea-modal" class="modal show">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Edit Idea</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="edit-idea-form">
                        <input type="hidden" name="ideaId" value="${idea.id}">
                        
                        <div class="form-group">
                            <label for="edit-idea-oneliner">One-liner *</label>
                            <input type="text" id="edit-idea-oneliner" name="oneLiner" required maxlength="200" 
                                   value="${this.escapeHtml(idea.oneLiner)}" placeholder="Brief description of your idea">
                            <small class="form-help">Maximum 200 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-idea-description">Description</label>
                            <textarea id="edit-idea-description" name="description" maxlength="1000" rows="4"
                                      placeholder="Detailed description of your idea (optional)">${this.escapeHtml(idea.description)}</textarea>
                            <small class="form-help">Maximum 1000 characters (optional)</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-idea-value-statement">Value Statement</label>
                            <textarea id="edit-idea-value-statement" name="valueStatement" maxlength="500" rows="3"
                                      placeholder="What value does this idea provide? (optional)">${this.escapeHtml(idea.valueStatement)}</textarea>
                            <small class="form-help">Maximum 500 characters (optional)</small>
                        </div>
                        
                        <div class="rice-score-section">
                            <h4>RICE Score</h4>
                            <div class="rice-grid">
                                <div class="form-group">
                                    <label for="edit-rice-reach">Reach (%)</label>
                                    <input type="number" id="edit-rice-reach" name="reach" min="0" max="100" 
                                           value="${idea.riceScore.reach}" placeholder="0-100">
                                    <small class="form-help">Percentage of users affected</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-impact">Impact (%)</label>
                                    <input type="number" id="edit-rice-impact" name="impact" min="0" max="100" 
                                           value="${idea.riceScore.impact}" placeholder="0-100">
                                    <small class="form-help">Impact per user</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-confidence">Confidence (%)</label>
                                    <input type="number" id="edit-rice-confidence" name="confidence" min="0" max="100" 
                                           value="${idea.riceScore.confidence}" placeholder="0-100">
                                    <small class="form-help">Confidence in the estimate</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-effort">Effort</label>
                                    <select id="edit-rice-effort" name="effort">
                                        <option value="">Select effort level</option>
                                        <option value="1" ${idea.riceScore.effort === 1 ? 'selected' : ''}>1 - Low (hours)</option>
                                        <option value="3" ${idea.riceScore.effort === 3 ? 'selected' : ''}>3 - Medium (days)</option>
                                        <option value="8" ${idea.riceScore.effort === 8 ? 'selected' : ''}>8 - High (weeks)</option>
                                        <option value="21" ${idea.riceScore.effort === 21 ? 'selected' : ''}>21 - Very High (months)</option>
                                    </select>
                                    <small class="form-help">Development effort required</small>
                                </div>
                            </div>
                            <div class="rice-score-display">
                                <span>RICE Score: <strong id="edit-rice-score-value">${this.calculateRICEScore(idea.riceScore).toFixed(1)}</strong></span>
                            </div>
                        </div>
                        </form>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="ideaManager.closeEditModal()">Cancel</button>
                        <button type="submit" form="edit-idea-form" class="btn btn-primary">Update Idea</button>
                    </div>
                </div>
            </div>
        `;
    }

    // Delete Confirmation Modal
    createDeleteConfirmationModal(ideaId, ideaTitle) {
        return `
            <div id="delete-idea-modal" class="modal" style="display: flex;">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Delete Idea</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <div class="modal-body">
                        <p>Are you sure you want to delete the idea "<strong>${this.escapeHtml(ideaTitle)}</strong>"?</p>
                        <p class="warning-text">This action cannot be undone.</p>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="ideaManager.closeDeleteModal()">Cancel</button>
                        <button type="button" class="btn btn-danger" onclick="ideaManager.deleteIdea('${ideaId}')">Delete Idea</button>
                    </div>
                </div>
            </div>
        `;
    }

    // Event Handlers
    openCreateModal() {
        console.log('[IdeaManager] Opening create idea modal...');
        
        // Remove existing modal if any
        const existingModal = document.getElementById('create-idea-modal');
        if (existingModal) {
            existingModal.remove();
        }

        // Add modal to page
        document.body.insertAdjacentHTML('beforeend', this.createIdeaCreationForm());
        
        // Setup RICE score calculation
        this.setupRICECalculation('create');
        
        // Focus first input
        const modal = document.getElementById('create-idea-modal');
        const firstInput = modal.querySelector('input[type="text"]');
        if (firstInput) {
            firstInput.focus();
        }
        
        console.log('[IdeaManager] Create idea modal opened');
    }

    closeCreateModal() {
        const modal = document.getElementById('create-idea-modal');
        if (modal) {
            modal.remove();
        }
    }

    async editIdea(ideaId) {
        try {
            console.log('[IdeaManager] editIdea called with ideaId:', ideaId);
            
            // Find the idea in the current board data
            let idea = null;
            
            // Try to get idea from drag-drop board if available
            if (window.dragDropBoard && window.dragDropBoard.ideas) {
                idea = window.dragDropBoard.ideas.find(i => i.id === ideaId);
            }
            
            // If not found in drag-drop board, try to get from board view
            if (!idea && window.boardView && window.boardView.ideas) {
                idea = window.boardView.ideas.find(i => i.id === ideaId);
            }
            
            if (!idea) {
                console.error('[IdeaManager] Idea not found:', ideaId);
                this.showErrorMessage('Idea not found. Please refresh the page and try again.');
                return;
            }
            
            console.log('[IdeaManager] Found idea for editing:', idea);
            this.editingIdeaId = ideaId;
            
            // Remove existing modal if any
            const existingModal = document.getElementById('edit-idea-modal');
            if (existingModal) {
                existingModal.remove();
            }

            // Add modal to page
            document.body.insertAdjacentHTML('beforeend', this.createIdeaEditModal(idea));
            
            // Setup RICE score calculation
            this.setupRICECalculation('edit');
            
            // Focus first input
            const firstInput = document.querySelector('#edit-idea-modal input[type="text"]');
            if (firstInput) {
                firstInput.focus();
            }
            
            // Close idea menu
            this.closeIdeaMenus();
            
        } catch (error) {
            console.error('[IdeaManager] Failed to load idea for editing:', error);
            this.showErrorMessage('Failed to load idea details. Please try again.');
        }
    }

    closeEditModal() {
        const modal = document.getElementById('edit-idea-modal');
        if (modal) {
            modal.remove();
        }
        this.editingIdeaId = null;
    }

    confirmDeleteIdea(ideaId, ideaTitle) {
        // Remove existing modal if any
        const existingModal = document.getElementById('delete-idea-modal');
        if (existingModal) {
            existingModal.remove();
        }

        // Add modal to page
        document.body.insertAdjacentHTML('beforeend', this.createDeleteConfirmationModal(ideaId, ideaTitle));
        
        // Close idea menu
        this.closeIdeaMenus();
    }

    closeDeleteModal() {
        const modal = document.getElementById('delete-idea-modal');
        if (modal) {
            modal.remove();
        }
    }

    closeModals() {
        this.closeCreateModal();
        this.closeEditModal();
        this.closeDeleteModal();
    }

    // API Handlers
    async handleCreateIdea(e) {
        console.log('[IdeaManager] handleCreateIdea called');
        const formData = new FormData(e.target);
        
        // Ensure RICE score always has default values
        const riceScore = {
            reach: parseInt(formData.get('reach')) || 100,
            impact: parseInt(formData.get('impact')) || 50,
            confidence: parseInt(formData.get('confidence')) || 50,
            effort: parseInt(formData.get('effort')) || 1
        };
        
        const ideaData = {
            oneLiner: formData.get('oneLiner'),
            description: formData.get('description') || '',
            valueStatement: formData.get('valueStatement') || '',
            riceScore: riceScore,
            column: 'parking', // New ideas start in parking
            status: 'active'
        };
        console.log('[IdeaManager] Form data collected:', ideaData);
        console.log('[IdeaManager] RICE score data:', ideaData.riceScore);

        // Validate form
        console.log('[IdeaManager] Validating form data...');
        if (!this.validateIdeaForm(ideaData)) {
            console.log('[IdeaManager] Form validation failed');
            return;
        }
        console.log('[IdeaManager] Form validation passed');

        try {
            const modal = e.target.closest('.modal');
            const submitBtn = modal.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Creating...';

            console.log('[IdeaManager] Making API call to create idea...');
            const response = await window.api.post(`/boards/${this.currentBoardId}/ideas`, ideaData);
            console.log('[IdeaManager] API response received:', response);
            
            this.closeCreateModal();
            this.showSuccessMessage('Idea created successfully!');
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to create idea:', error);
            this.handleFormError(error, 'create');
        } finally {
            const modal = e.target.closest('.modal');
            const submitBtn = modal.querySelector('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Create Idea';
            }
        }
    }

    async handleEditIdea(e) {
        const formData = new FormData(e.target);
        
        // Ensure RICE score always has default values
        const riceScore = {
            reach: parseInt(formData.get('reach')) || 100,
            impact: parseInt(formData.get('impact')) || 50,
            confidence: parseInt(formData.get('confidence')) || 50,
            effort: parseInt(formData.get('effort')) || 1
        };
        
        const ideaData = {
            oneLiner: formData.get('oneLiner'),
            description: formData.get('description') || '',
            valueStatement: formData.get('valueStatement') || '',
            riceScore: riceScore
        };

        // Validate form
        if (!this.validateIdeaForm(ideaData)) {
            return;
        }

        try {
            const modal = e.target.closest('.modal');
            const submitBtn = modal.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Updating...';

            const response = await window.api.put(`/ideas/${this.editingIdeaId}`, ideaData);
            
            this.closeEditModal();
            this.showSuccessMessage('Idea updated successfully!');
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to update idea:', error);
            this.handleFormError(error, 'edit');
        } finally {
            const modal = e.target.closest('.modal');
            const submitBtn = modal.querySelector('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Update Idea';
            }
        }
    }

    async deleteIdea(ideaId) {
        try {
            const deleteBtn = document.querySelector('#delete-idea-modal .btn-danger');
            const originalText = deleteBtn.textContent;
            deleteBtn.disabled = true;
            deleteBtn.textContent = 'Deleting...';

            await window.api.delete(`/ideas/${ideaId}`);
            
            this.closeDeleteModal();
            this.showSuccessMessage('Idea deleted successfully!');
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to delete idea:', error);
            this.showErrorMessage('Failed to delete idea. Please try again.');
        } finally {
            const deleteBtn = document.querySelector('#delete-idea-modal .btn-danger');
            if (deleteBtn) {
                deleteBtn.disabled = false;
                deleteBtn.textContent = 'Delete Idea';
            }
        }
    }

    async toggleInProgress(ideaId, inProgress) {
        try {
            // Add visual feedback during status change
            const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
            if (ideaCard) {
                ideaCard.classList.add('status-changing');
            }

            await window.api.put(`/ideas/${ideaId}/status`, { inProgress });
            
            this.showSuccessMessage(`Idea marked as ${inProgress ? 'in progress' : 'not in progress'}!`);
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to update idea status:', error);
            this.showErrorMessage('Failed to update idea status. Please try again.');
        } finally {
            // Remove status changing animation
            const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
            if (ideaCard) {
                setTimeout(() => {
                    ideaCard.classList.remove('status-changing');
                }, 500);
            }
        }
        
        this.closeIdeaMenus();
    }

    async markAsDone(ideaId) {
        // Use the updateIdeaStatus method for consistency
        await this.updateIdeaStatus(ideaId, 'done');
    }

    // New method to handle status changes with automatic column transitions
    async updateIdeaStatus(ideaId, status, inProgress = null) {
        try {
            // Add visual feedback during status change
            const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
            if (ideaCard) {
                // Add specific animation class based on status
                switch (status) {
                    case 'done':
                        ideaCard.classList.add('status-done');
                        break;
                    case 'archived':
                        ideaCard.classList.add('status-archived');
                        break;
                    case 'active':
                        ideaCard.classList.add('status-reactivated');
                        break;
                    default:
                        ideaCard.classList.add('status-changing');
                }
            }

            const updateData = { status };
            if (inProgress !== null) {
                updateData.inProgress = inProgress;
            }

            await window.api.put(`/ideas/${ideaId}/status`, updateData);
            
            // Show appropriate success message based on status
            let message = 'Idea status updated!';
            let icon = '✅';
            switch (status) {
                case 'done':
                    message = 'Idea marked as done and moved to Release!';
                    icon = '🎉';
                    break;
                case 'archived':
                    message = 'Idea archived and moved to Won\'t Do!';
                    icon = '🗃️';
                    break;
                case 'active':
                    message = 'Idea reactivated and moved to Parking!';
                    icon = '🔄';
                    break;
            }
            
            this.showSuccessMessage(`${icon} ${message}`);
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to update idea status:', error);
            this.showErrorMessage('Failed to update idea status. Please try again.');
        } finally {
            // Remove all status animation classes
            const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
            if (ideaCard) {
                setTimeout(() => {
                    ideaCard.classList.remove('status-changing', 'status-done', 'status-archived', 'status-reactivated');
                }, 800);
            }
        }
        
        this.closeIdeaMenus();
    }

    // Menu management utilities
    toggleIdeaMenu(ideaId) {
        // Close all other menus first
        this.closeIdeaMenus();
        
        // Toggle the specific menu
        const menu = document.getElementById(`idea-menu-${ideaId}`);
        if (menu) {
            const isVisible = menu.style.display !== 'none';
            menu.style.display = isVisible ? 'none' : 'block';
        }
    }

    closeIdeaMenus() {
        const menus = document.querySelectorAll('.idea-menu');
        menus.forEach(menu => {
            menu.style.display = 'none';
        });
    }

    // Feedback handlers (placeholder for future implementation)
    async addThumbsUp(ideaId) {
        try {
            await window.api.post(`/ideas/${ideaId}/thumbsup`);
            
            // Update the thumbs up count in the UI
            const ideaCard = document.querySelector(`[data-idea-id="${ideaId}"]`);
            if (ideaCard) {
                const countSpan = ideaCard.querySelector('.thumbs-up .count');
                if (countSpan) {
                    const currentCount = parseInt(countSpan.textContent) || 0;
                    countSpan.textContent = currentCount + 1;
                }
            }
            
        } catch (error) {
            console.error('Failed to add thumbs up:', error);
            this.showErrorMessage('Failed to add thumbs up. Please try again.');
        }
    }

    showEmojiPicker(ideaId) {
        // Placeholder for emoji picker implementation
        console.log('Emoji picker for idea:', ideaId);
        this.showInfoMessage('Emoji reactions will be implemented in a future task.');
    }

    // Utility Methods
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

    closeIdeaMenus() {
        const allMenus = document.querySelectorAll('.idea-menu');
        allMenus.forEach(menu => {
            menu.style.display = 'none';
        });
    }



    validateIdeaForm(ideaData) {
        console.log('[IdeaManager] validateIdeaForm called with:', ideaData);
        console.log('[IdeaManager] validateIdeaForm - oneLiner:', ideaData.oneLiner);
        console.log('[IdeaManager] validateIdeaForm - riceScore:', ideaData.riceScore);
        let isValid = true;

        // Clear previous errors
        document.querySelectorAll('.form-error').forEach(error => error.remove());
        document.querySelectorAll('.error').forEach(field => field.classList.remove('error'));

        // Validate one-liner
        console.log('[IdeaManager] Validating one-liner:', ideaData.oneLiner);
        if (!ideaData.oneLiner || ideaData.oneLiner.trim().length === 0) {
            console.log('[IdeaManager] One-liner validation failed - empty');
            this.showFieldError('oneLiner', 'One-liner is required');
            isValid = false;
        } else if (ideaData.oneLiner.length > 200) {
            console.log('[IdeaManager] One-liner validation failed - too long');
            this.showFieldError('oneLiner', 'One-liner must be less than 200 characters');
            isValid = false;
        } else {
            console.log('[IdeaManager] One-liner validation passed');
        }

        // Validate description (optional)
        if (ideaData.description && ideaData.description.length > 1000) {
            this.showFieldError('description', 'Description must be less than 1000 characters');
            isValid = false;
        }

        // Validate value statement (optional)
        if (ideaData.valueStatement && ideaData.valueStatement.length > 500) {
            this.showFieldError('valueStatement', 'Value statement must be less than 500 characters');
            isValid = false;
        }

        // Validate RICE score
        const rice = ideaData.riceScore;
        console.log('[IdeaManager] Validating RICE score:', rice);
        
        if (isNaN(rice.reach) || rice.reach < 0 || rice.reach > 100) {
            console.log('[IdeaManager] Reach validation failed:', rice.reach);
            this.showFieldError('reach', 'Reach must be between 0 and 100');
            isValid = false;
        } else {
            console.log('[IdeaManager] Reach validation passed:', rice.reach);
        }
        
        if (isNaN(rice.impact) || rice.impact < 0 || rice.impact > 100) {
            console.log('[IdeaManager] Impact validation failed:', rice.impact);
            this.showFieldError('impact', 'Impact must be between 0 and 100');
            isValid = false;
        } else {
            console.log('[IdeaManager] Impact validation passed:', rice.impact);
        }
        
        if (isNaN(rice.confidence) || rice.confidence < 0 || rice.confidence > 100) {
            console.log('[IdeaManager] Confidence validation failed:', rice.confidence);
            this.showFieldError('confidence', 'Confidence must be between 0 and 100');
            isValid = false;
        } else {
            console.log('[IdeaManager] Confidence validation passed:', rice.confidence);
        }
        
        if (isNaN(rice.effort) || ![1, 3, 8, 21].includes(rice.effort)) {
            console.log('[IdeaManager] Effort validation failed:', rice.effort);
            this.showFieldError('effort', 'Please select a valid effort level');
            isValid = false;
        } else {
            console.log('[IdeaManager] Effort validation passed:', rice.effort);
        }

        console.log('[IdeaManager] Final validation result:', isValid);
        return isValid;
    }

    showFieldError(fieldName, message) {
        const fieldId = this.editingIdeaId ? `edit-idea-${fieldName}` : `idea-${fieldName}`;
        let field = document.getElementById(fieldId);
        
        // Handle RICE fields
        if (!field) {
            const riceFieldId = this.editingIdeaId ? `edit-rice-${fieldName}` : `rice-${fieldName}`;
            field = document.getElementById(riceFieldId);
        }
        
        if (!field) return;
        
        // Add error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'form-error';
        errorDiv.textContent = message;
        field.parentNode.appendChild(errorDiv);
        
        // Add error styling
        field.classList.add('error');
    }

    handleFormError(error, formType) {
        if (error.response && error.response.data && error.response.data.error) {
            const errorData = error.response.data.error;
            if (errorData.code === 'VALIDATION_ERROR' && errorData.details) {
                // Handle field-specific validation errors
                Object.keys(errorData.details).forEach(field => {
                    this.showFieldError(field, errorData.details[field]);
                });
            } else {
                this.showErrorMessage(errorData.message || `Failed to ${formType} idea`);
            }
        } else {
            this.showErrorMessage(`Failed to ${formType} idea. Please try again.`);
        }
    }

    calculateRICEScore(riceScore) {
        if (!riceScore || riceScore.effort === 0) return 0;
        // Convert percentages to decimals (0-100 -> 0-1)
        const reach = riceScore.reach / 100;
        const impact = riceScore.impact / 100;
        const confidence = riceScore.confidence / 100;
        return (reach * impact * confidence) / riceScore.effort;
    }

    getEmojiCount(emojiReactions) {
        if (!emojiReactions || !Array.isArray(emojiReactions)) return 0;
        return emojiReactions.reduce((total, reaction) => total + reaction.count, 0);
    }

    formatStatus(status) {
        const statusMap = {
            'draft': 'Draft',
            'active': 'Active',
            'done': 'Done',
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
        const date = new Date(dateString);
        const now = new Date();
        const diffTime = Math.abs(now - date);
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) return `${diffDays} days ago`;
        if (diffDays < 30) return `${Math.ceil(diffDays / 7)} weeks ago`;
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

    // Menu management utilities
    toggleIdeaMenu(ideaId) {
        // Close all other menus first
        this.closeIdeaMenus();
        
        // Toggle the specific menu
        const menu = document.getElementById(`idea-menu-${ideaId}`);
        if (menu) {
            const isVisible = menu.style.display !== 'none';
            menu.style.display = isVisible ? 'none' : 'block';
        }
    }

    closeIdeaMenus() {
        const menus = document.querySelectorAll('.idea-menu');
        menus.forEach(menu => {
            menu.style.display = 'none';
        });
    }

    // Message utilities
    showSuccessMessage(message) {
        this.showMessage(message, 'success');
    }

    showErrorMessage(message) {
        this.showMessage(message, 'error');
    }

    showInfoMessage(message) {
        this.showMessage(message, 'info');
    }

    showMessage(message, type) {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message-toast');
        existingMessages.forEach(msg => msg.remove());
        
        // Create message element
        const messageDiv = document.createElement('div');
        messageDiv.className = `message-toast ${type}`;
        messageDiv.textContent = message;
        
        // Add to page
        document.body.appendChild(messageDiv);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.remove();
            }
        }, 5000);
        
        // Allow manual close
        messageDiv.addEventListener('click', () => {
            messageDiv.remove();
        });
    }

    // RICE score calculation setup
    setupRICECalculation(formType) {
        const prefix = formType === 'edit' ? 'edit-' : '';
        const reachInput = document.getElementById(`${prefix}rice-reach`);
        const impactInput = document.getElementById(`${prefix}rice-impact`);
        const confidenceInput = document.getElementById(`${prefix}rice-confidence`);
        const effortInput = document.getElementById(`${prefix}rice-effort`);
        const scoreDisplay = document.getElementById(`${prefix}rice-score-value`);

        const calculateScore = () => {
            const reach = parseInt(reachInput.value) || 0;
            const impact = parseInt(impactInput.value) || 0;
            const confidence = parseInt(confidenceInput.value) || 0;
            const effort = parseInt(effortInput.value) || 1;

            // RICE Score formula: (Reach × Impact × Confidence) ÷ Effort
            // All values are now properly scaled:
            // - Reach: 0-100 (percentage)
            // - Impact: 0-100 (percentage) 
            // - Confidence: 0-100 (percentage)
            // - Effort: 1-21 (scale: 1=Low, 3=Medium, 8=High, 21=Very High)
            const score = effort > 0 ? (reach * impact * confidence) / effort : 0;
            
            if (scoreDisplay) {
                scoreDisplay.textContent = score.toFixed(1);
            }
        };

        // Add event listeners
        [reachInput, impactInput, confidenceInput, effortInput].forEach(input => {
            if (input) {
                input.addEventListener('input', calculateScore);
                input.addEventListener('change', calculateScore);
            }
        });

        // Calculate initial score
        calculateScore();
    }


}

// Initialize idea manager
window.ideaManager = new IdeaManager();

// Close menus when clicking outside
document.addEventListener('click', (e) => {
    if (!e.target.closest('.idea-card-menu')) {
        window.ideaManager.closeIdeaMenus();
    }
});