// FeedbackWidget component for public users
class FeedbackWidget {
    constructor(ideaId, initialThumbsUp = 0, initialEmojiReactions = []) {
        this.ideaId = ideaId;
        this.thumbsUp = initialThumbsUp;
        this.emojiReactions = initialEmojiReactions;
        this.isSubmitting = false;
        this.element = null;
        this.init();
    }

    init() {
        this.createElement();
        this.bindEvents();
    }

    createElement() {
        this.element = document.createElement('div');
        this.element.className = 'feedback-widget';
        this.element.innerHTML = this.getHTML();
    }

    getHTML() {
        return `
            <div class="feedback-stats">
                <span class="thumbs-up-count">
                    <span class="thumbs-up-icon">👍</span>
                    <span class="count">${this.thumbsUp}</span>
                </span>
                ${this.renderEmojiReactions()}
            </div>
            <div class="feedback-actions">
                <button class="feedback-btn thumbs-up-btn" ${this.isSubmitting ? 'disabled' : ''}>
                    👍 Like
                </button>
                <button class="feedback-btn emoji-btn" ${this.isSubmitting ? 'disabled' : ''}>
                    😊 React
                </button>
            </div>
            <div class="feedback-message" style="display: none;"></div>
        `;
    }

    renderEmojiReactions() {
        if (!this.emojiReactions || this.emojiReactions.length === 0) {
            return '';
        }

        return this.emojiReactions.map(reaction => 
            `<span class="emoji-reaction">
                <span class="emoji">${reaction.emoji}</span>
                <span class="count">${reaction.count}</span>
            </span>`
        ).join('');
    }

    bindEvents() {
        // Thumbs up button
        const thumbsUpBtn = this.element.querySelector('.thumbs-up-btn');
        if (thumbsUpBtn) {
            thumbsUpBtn.addEventListener('click', () => {
                this.handleThumbsUp();
            });
        }

        // Emoji button
        const emojiBtn = this.element.querySelector('.emoji-btn');
        if (emojiBtn) {
            emojiBtn.addEventListener('click', () => {
                this.showEmojiPicker();
            });
        }
    }

    async handleThumbsUp() {
        if (this.isSubmitting) return;

        this.setSubmitting(true);
        this.showMessage('Adding thumbs up...', 'info');

        try {
            const response = await fetch(`/api/ideas/${this.ideaId}/thumbsup`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.thumbsUp = data.thumbsUp;
                this.updateDisplay();
                this.showMessage('Thanks for your feedback! 👍', 'success');
                
                // Trigger animation
                this.animateThumbsUp();
            } else {
                const errorData = await response.json();
                this.handleError(errorData);
            }
        } catch (error) {
            console.error('Error adding thumbs up:', error);
            this.showMessage('Failed to add thumbs up. Please try again.', 'error');
        } finally {
            this.setSubmitting(false);
        }
    }

    showEmojiPicker() {
        if (this.isSubmitting) return;

        // Create emoji picker modal
        const modal = document.createElement('div');
        modal.className = 'emoji-picker-modal';
        modal.innerHTML = `
            <div class="emoji-picker-backdrop"></div>
            <div class="emoji-picker">
                <div class="emoji-picker-header">
                    <h3>Choose an emoji</h3>
                    <button class="close-btn">&times;</button>
                </div>
                <div class="emoji-grid">
                    ${this.getEmojiGrid()}
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // Bind events
        modal.querySelector('.close-btn').addEventListener('click', () => {
            document.body.removeChild(modal);
        });

        modal.querySelector('.emoji-picker-backdrop').addEventListener('click', () => {
            document.body.removeChild(modal);
        });

        // Bind emoji selection
        modal.querySelectorAll('.emoji-option').forEach(option => {
            option.addEventListener('click', async () => {
                const emoji = option.dataset.emoji;
                document.body.removeChild(modal);
                await this.handleEmojiReaction(emoji);
            });
        });
    }

    getEmojiGrid() {
        const emojis = [
            '😀', '😃', '😄', '😁', '😆', '😅', '😂', '🤣', '😊', '😇',
            '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚',
            '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🤩',
            '🥳', '😏', '😒', '😞', '😔', '😟', '😕', '🙁', '☹️', '😣',
            '😖', '😫', '😩', '🥺', '😢', '😭', '😤', '😠', '😡', '🤬',
            '🤯', '😳', '🥵', '🥶', '😱', '😨', '😰', '😥', '😓', '🤗',
            '👍', '👎', '👌', '✌️', '🤞', '🤟', '🤘', '🤙', '👏', '🙌',
            '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔',
            '⭐', '🌟', '💫', '✨', '🌠', '🔥', '💥', '⚡', '🌈', '🎉',
            '🎊', '🎈', '🎁', '🎀', '🏆', '🥇', '🥈', '🥉', '🏅', '🎖️'
        ];

        return emojis.map(emoji => 
            `<button class="emoji-option" data-emoji="${emoji}">${emoji}</button>`
        ).join('');
    }

    async handleEmojiReaction(emoji) {
        if (this.isSubmitting) return;

        this.setSubmitting(true);
        this.showMessage(`Adding ${emoji} reaction...`, 'info');

        try {
            const response = await fetch(`/api/ideas/${this.ideaId}/emoji`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ emoji })
            });

            if (response.ok) {
                // Refresh the idea to get updated emoji reactions
                await this.refreshEmojiReactions();
                this.showMessage(`Thanks for your ${emoji} reaction!`, 'success');
                
                // Trigger animation
                this.animateEmojiReaction(emoji);
            } else {
                const errorData = await response.json();
                this.handleError(errorData);
            }
        } catch (error) {
            console.error('Error adding emoji reaction:', error);
            this.showMessage('Failed to add emoji reaction. Please try again.', 'error');
        } finally {
            this.setSubmitting(false);
        }
    }

    async refreshEmojiReactions() {
        // This would typically be handled by the parent component
        // For now, we'll trigger a page refresh or emit an event
        if (window.publicBoardView) {
            await window.publicBoardView.loadPublicBoard();
        }
    }

    handleError(errorData) {
        const message = errorData.error?.message || 'An error occurred';
        
        if (errorData.error?.code === 'RATE_LIMITED') {
            this.showMessage('Please wait a moment before giving more feedback.', 'warning');
        } else {
            this.showMessage(message, 'error');
        }
    }

    setSubmitting(isSubmitting) {
        this.isSubmitting = isSubmitting;
        
        const thumbsUpBtn = this.element.querySelector('.thumbs-up-btn');
        const emojiBtn = this.element.querySelector('.emoji-btn');
        
        if (thumbsUpBtn) thumbsUpBtn.disabled = isSubmitting;
        if (emojiBtn) emojiBtn.disabled = isSubmitting;
    }

    showMessage(message, type = 'info') {
        const messageEl = this.element.querySelector('.feedback-message');
        if (messageEl) {
            messageEl.textContent = message;
            messageEl.className = `feedback-message ${type}`;
            messageEl.style.display = 'block';
            
            // Hide message after 3 seconds
            setTimeout(() => {
                messageEl.style.display = 'none';
            }, 3000);
        }
    }

    updateDisplay() {
        const countEl = this.element.querySelector('.thumbs-up-count .count');
        if (countEl) {
            countEl.textContent = this.thumbsUp;
        }

        const statsEl = this.element.querySelector('.feedback-stats');
        if (statsEl) {
            statsEl.innerHTML = `
                <span class="thumbs-up-count">
                    <span class="thumbs-up-icon">👍</span>
                    <span class="count">${this.thumbsUp}</span>
                </span>
                ${this.renderEmojiReactions()}
            `;
        }
    }

    animateThumbsUp() {
        const thumbsUpIcon = this.element.querySelector('.thumbs-up-icon');
        if (thumbsUpIcon) {
            thumbsUpIcon.classList.add('feedback-animation');
            setTimeout(() => {
                thumbsUpIcon.classList.remove('feedback-animation');
            }, 600);
        }
    }

    animateEmojiReaction(emoji) {
        // Create floating emoji animation
        const floatingEmoji = document.createElement('div');
        floatingEmoji.className = 'floating-emoji';
        floatingEmoji.textContent = emoji;
        
        const rect = this.element.getBoundingClientRect();
        floatingEmoji.style.left = rect.left + rect.width / 2 + 'px';
        floatingEmoji.style.top = rect.top + 'px';
        
        document.body.appendChild(floatingEmoji);
        
        // Remove after animation
        setTimeout(() => {
            if (document.body.contains(floatingEmoji)) {
                document.body.removeChild(floatingEmoji);
            }
        }, 1000);
    }

    // Update the widget with new data
    update(thumbsUp, emojiReactions) {
        this.thumbsUp = thumbsUp;
        this.emojiReactions = emojiReactions;
        this.updateDisplay();
    }

    // Get the DOM element
    getElement() {
        return this.element;
    }
}

// Export for use in other modules
window.FeedbackWidget = FeedbackWidget;