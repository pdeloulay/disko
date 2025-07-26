// Idea Management Components
// This module provides UI components for creating, editing, and managing ideas

class IdeaManager {
    constructor() {
        this.currentBoardId = null;
        this.editingIdeaId = null;
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // Create idea form submission
        document.addEventListener('submit', (e) => {
            if (e.target.id === 'create-idea-form') {
                e.preventDefault();
                this.handleCreateIdea(e);
            }
            if (e.target.id === 'edit-idea-form') {
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

        // Escape key to close modals
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModals();
            }
        });
    }

    setBoardId(boardId) {
        this.currentBoardId = boardId;
    }

    // Create Idea Form Component
    createIdeaCreationForm() {
        return `
            <div id="create-idea-modal" class="modal" style="display: none;">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Create New Idea</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <form id="create-idea-form">
                        <div class="form-group">
                            <label for="idea-oneliner">One-liner *</label>
                            <input type="text" id="idea-oneliner" name="oneLiner" required maxlength="200" 
                                   placeholder="Brief description of your idea">
                            <small class="form-help">Maximum 200 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="idea-description">Description *</label>
                            <textarea id="idea-description" name="description" required maxlength="1000" rows="4"
                                      placeholder="Detailed description of your idea"></textarea>
                            <small class="form-help">Maximum 1000 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="idea-value-statement">Value Statement *</label>
                            <textarea id="idea-value-statement" name="valueStatement" required maxlength="500" rows="3"
                                      placeholder="What value does this idea provide?"></textarea>
                            <small class="form-help">Maximum 500 characters</small>
                        </div>
                        
                        <div class="rice-score-section">
                            <h4>RICE Score</h4>
                            <div class="rice-grid">
                                <div class="form-group">
                                    <label for="rice-reach">Reach (%) *</label>
                                    <input type="number" id="rice-reach" name="reach" required min="0" max="100" 
                                           placeholder="0-100">
                                    <small class="form-help">Percentage of users affected</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-impact">Impact (%) *</label>
                                    <input type="number" id="rice-impact" name="impact" required min="0" max="100" 
                                           placeholder="0-100">
                                    <small class="form-help">Impact per user</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-confidence">Confidence *</label>
                                    <select id="rice-confidence" name="confidence" required>
                                        <option value="">Select confidence level</option>
                                        <option value="1">1 - Low (hours)</option>
                                        <option value="2">2 - Medium (days)</option>
                                        <option value="4">4 - High (weeks)</option>
                                        <option value="8">8 - Very High (months)</option>
                                    </select>
                                    <small class="form-help">Time investment confidence</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="rice-effort">Effort (%) *</label>
                                    <input type="number" id="rice-effort" name="effort" required min="0" max="100" 
                                           placeholder="0-100">
                                    <small class="form-help">Development effort required</small>
                                </div>
                            </div>
                            <div class="rice-score-display">
                                <span>RICE Score: <strong id="rice-score-value">0</strong></span>
                            </div>
                        </div>
                        
                        <div class="form-actions">
                            <button type="button" class="btn btn-secondary" onclick="ideaManager.closeCreateModal()">Cancel</button>
                            <button type="submit" class="btn btn-primary">Create Idea</button>
                        </div>
                    </form>
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
                                <button onclick="ideaManager.editIdea('${idea.id}')">Edit</button>
                                <button onclick="ideaManager.toggleInProgress('${idea.id}', ${!idea.inProgress})">
                                    ${idea.inProgress ? 'Mark as Not In Progress' : 'Mark as In Progress'}
                                </button>
                                <button onclick="ideaManager.markAsDone('${idea.id}')">Mark as Done</button>
                                <button onclick="ideaManager.confirmDeleteIdea('${idea.id}', '${this.escapeHtml(idea.oneLiner)}')">Delete</button>
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
                </div>
                
                <div class="idea-meta">
                    <span class="idea-status">${this.formatStatus(idea.status)}</span>
                    <span class="idea-column">${this.formatColumn(idea.column)}</span>
                    <span class="idea-date">${this.formatDate(idea.createdAt)}</span>
                </div>
            </div>
        `;
    }

    // Idea Edit Modal Component
    createIdeaEditModal(idea) {
        return `
            <div id="edit-idea-modal" class="modal" style="display: flex;">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3>Edit Idea</h3>
                        <button class="modal-close">&times;</button>
                    </div>
                    <form id="edit-idea-form">
                        <input type="hidden" name="ideaId" value="${idea.id}">
                        
                        <div class="form-group">
                            <label for="edit-idea-oneliner">One-liner *</label>
                            <input type="text" id="edit-idea-oneliner" name="oneLiner" required maxlength="200" 
                                   value="${this.escapeHtml(idea.oneLiner)}" placeholder="Brief description of your idea">
                            <small class="form-help">Maximum 200 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-idea-description">Description *</label>
                            <textarea id="edit-idea-description" name="description" required maxlength="1000" rows="4"
                                      placeholder="Detailed description of your idea">${this.escapeHtml(idea.description)}</textarea>
                            <small class="form-help">Maximum 1000 characters</small>
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-idea-value-statement">Value Statement *</label>
                            <textarea id="edit-idea-value-statement" name="valueStatement" required maxlength="500" rows="3"
                                      placeholder="What value does this idea provide?">${this.escapeHtml(idea.valueStatement)}</textarea>
                            <small class="form-help">Maximum 500 characters</small>
                        </div>
                        
                        <div class="rice-score-section">
                            <h4>RICE Score</h4>
                            <div class="rice-grid">
                                <div class="form-group">
                                    <label for="edit-rice-reach">Reach (%) *</label>
                                    <input type="number" id="edit-rice-reach" name="reach" required min="0" max="100" 
                                           value="${idea.riceScore.reach}" placeholder="0-100">
                                    <small class="form-help">Percentage of users affected</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-impact">Impact (%) *</label>
                                    <input type="number" id="edit-rice-impact" name="impact" required min="0" max="100" 
                                           value="${idea.riceScore.impact}" placeholder="0-100">
                                    <small class="form-help">Impact per user</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-confidence">Confidence *</label>
                                    <select id="edit-rice-confidence" name="confidence" required>
                                        <option value="">Select confidence level</option>
                                        <option value="1" ${idea.riceScore.confidence === 1 ? 'selected' : ''}>1 - Low (hours)</option>
                                        <option value="2" ${idea.riceScore.confidence === 2 ? 'selected' : ''}>2 - Medium (days)</option>
                                        <option value="4" ${idea.riceScore.confidence === 4 ? 'selected' : ''}>4 - High (weeks)</option>
                                        <option value="8" ${idea.riceScore.confidence === 8 ? 'selected' : ''}>8 - Very High (months)</option>
                                    </select>
                                    <small class="form-help">Time investment confidence</small>
                                </div>
                                
                                <div class="form-group">
                                    <label for="edit-rice-effort">Effort (%) *</label>
                                    <input type="number" id="edit-rice-effort" name="effort" required min="0" max="100" 
                                           value="${idea.riceScore.effort}" placeholder="0-100">
                                    <small class="form-help">Development effort required</small>
                                </div>
                            </div>
                            <div class="rice-score-display">
                                <span>RICE Score: <strong id="edit-rice-score-value">${this.calculateRICEScore(idea.riceScore).toFixed(1)}</strong></span>
                            </div>
                        </div>
                        
                        <div class="form-actions">
                            <button type="button" class="btn btn-secondary" onclick="ideaManager.closeEditModal()">Cancel</button>
                            <button type="submit" class="btn btn-primary">Update Idea</button>
                        </div>
                    </form>
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
        // Remove existing modal if any
        const existingModal = document.getElementById('create-idea-modal');
        if (existingModal) {
            existingModal.remove();
        }

        // Add modal to page
        document.body.insertAdjacentHTML('beforeend', this.createIdeaCreationForm());
        
        // Show modal
        const modal = document.getElementById('create-idea-modal');
        modal.style.display = 'flex';
        
        // Setup RICE score calculation
        this.setupRICECalculation('create');
        
        // Focus first input
        const firstInput = modal.querySelector('input[type="text"]');
        if (firstInput) {
            firstInput.focus();
        }
    }

    closeCreateModal() {
        const modal = document.getElementById('create-idea-modal');
        if (modal) {
            modal.remove();
        }
    }

    async editIdea(ideaId) {
        try {
            // Fetch idea details
            const response = await window.api.get(`/ideas/${ideaId}`);
            const idea = response.data || response;
            
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
            console.error('Failed to load idea for editing:', error);
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
        const formData = new FormData(e.target);
        const ideaData = {
            oneLiner: formData.get('oneLiner'),
            description: formData.get('description'),
            valueStatement: formData.get('valueStatement'),
            riceScore: {
                reach: parseInt(formData.get('reach')),
                impact: parseInt(formData.get('impact')),
                confidence: parseInt(formData.get('confidence')),
                effort: parseInt(formData.get('effort'))
            },
            column: 'parking', // New ideas start in parking
            status: 'active'
        };

        // Validate form
        if (!this.validateIdeaForm(ideaData)) {
            return;
        }

        try {
            const submitBtn = e.target.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Creating...';

            const response = await window.api.post(`/boards/${this.currentBoardId}/ideas`, ideaData);
            
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
            const submitBtn = e.target.querySelector('button[type="submit"]');
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Create Idea';
            }
        }
    }

    async handleEditIdea(e) {
        const formData = new FormData(e.target);
        const ideaData = {
            oneLiner: formData.get('oneLiner'),
            description: formData.get('description'),
            valueStatement: formData.get('valueStatement'),
            riceScore: {
                reach: parseInt(formData.get('reach')),
                impact: parseInt(formData.get('impact')),
                confidence: parseInt(formData.get('confidence')),
                effort: parseInt(formData.get('effort'))
            }
        };

        // Validate form
        if (!this.validateIdeaForm(ideaData)) {
            return;
        }

        try {
            const submitBtn = e.target.querySelector('button[type="submit"]');
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
            const submitBtn = e.target.querySelector('button[type="submit"]');
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
            await window.api.put(`/ideas/${ideaId}/status`, { inProgress });
            
            this.showSuccessMessage(`Idea marked as ${inProgress ? 'in progress' : 'not in progress'}!`);
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to update idea status:', error);
            this.showErrorMessage('Failed to update idea status. Please try again.');
        }
        
        this.closeIdeaMenus();
    }

    async markAsDone(ideaId) {
        try {
            await window.api.put(`/ideas/${ideaId}/status`, { status: 'done' });
            
            this.showSuccessMessage('Idea marked as done and moved to Release!');
            
            // Trigger refresh of ideas list
            if (window.boardView && window.boardView.refreshIdeas) {
                await window.boardView.refreshIdeas();
            }
            
        } catch (error) {
            console.error('Failed to mark idea as done:', error);
            this.showErrorMessage('Failed to mark idea as done. Please try again.');
        }
        
        this.closeIdeaMenus();
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

    setupRICECalculation(prefix) {
        const reachInput = document.getElementById(`${prefix === 'create' ? '' : 'edit-'}rice-reach`);
        const impactInput = document.getElementById(`${prefix === 'create' ? '' : 'edit-'}rice-impact`);
        const confidenceInput = document.getElementById(`${prefix === 'create' ? '' : 'edit-'}rice-confidence`);
        const effortInput = document.getElementById(`${prefix === 'create' ? '' : 'edit-'}rice-effort`);
        const scoreDisplay = document.getElementById(`${prefix === 'create' ? '' : 'edit-'}rice-score-value`);

        const calculateScore = () => {
            const reach = parseInt(reachInput.value) || 0;
            const impact = parseInt(impactInput.value) || 0;
            const confidence = parseInt(confidenceInput.value) || 0;
            const effort = parseInt(effortInput.value) || 1;

            const score = effort > 0 ? (reach * impact * confidence) / effort : 0;
            scoreDisplay.textContent = score.toFixed(1);
        };

        [reachInput, impactInput, confidenceInput, effortInput].forEach(input => {
            if (input) {
                input.addEventListener('input', calculateScore);
                input.addEventListener('change', calculateScore);
            }
        });

        // Initial calculation
        calculateScore();
    }

    validateIdeaForm(ideaData) {
        let isValid = true;

        // Clear previous errors
        document.querySelectorAll('.form-error').forEach(error => error.remove());
        document.querySelectorAll('.error').forEach(field => field.classList.remove('error'));

        // Validate one-liner
        if (!ideaData.oneLiner || ideaData.oneLiner.trim().length === 0) {
            this.showFieldError('oneLiner', 'One-liner is required');
            isValid = false;
        } else if (ideaData.oneLiner.length > 200) {
            this.showFieldError('oneLiner', 'One-liner must be less than 200 characters');
            isValid = false;
        }

        // Validate description
        if (!ideaData.description || ideaData.description.trim().length === 0) {
            this.showFieldError('description', 'Description is required');
            isValid = false;
        } else if (ideaData.description.length > 1000) {
            this.showFieldError('description', 'Description must be less than 1000 characters');
            isValid = false;
        }

        // Validate value statement
        if (!ideaData.valueStatement || ideaData.valueStatement.trim().length === 0) {
            this.showFieldError('valueStatement', 'Value statement is required');
            isValid = false;
        } else if (ideaData.valueStatement.length > 500) {
            this.showFieldError('valueStatement', 'Value statement must be less than 500 characters');
            isValid = false;
        }

        // Validate RICE score
        const rice = ideaData.riceScore;
        if (isNaN(rice.reach) || rice.reach < 0 || rice.reach > 100) {
            this.showFieldError('reach', 'Reach must be between 0 and 100');
            isValid = false;
        }
        if (isNaN(rice.impact) || rice.impact < 0 || rice.impact > 100) {
            this.showFieldError('impact', 'Impact must be between 0 and 100');
            isValid = false;
        }
        if (isNaN(rice.confidence) || ![1, 2, 4, 8].includes(rice.confidence)) {
            this.showFieldError('confidence', 'Please select a valid confidence level');
            isValid = false;
        }
        if (isNaN(rice.effort) || rice.effort < 0 || rice.effort > 100) {
            this.showFieldError('effort', 'Effort must be between 0 and 100');
            isValid = false;
        }

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
        return (riceScore.reach * riceScore.impact * riceScore.confidence) / riceScore.effort;
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

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
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
}

// Initialize idea manager
window.ideaManager = new IdeaManager();

// Close menus when clicking outside
document.addEventListener('click', (e) => {
    if (!e.target.closest('.idea-card-menu')) {
        window.ideaManager.closeIdeaMenus();
    }
});