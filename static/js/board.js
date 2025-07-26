// Board view functionality - Updated to work with drag-drop board
class BoardView {
    constructor() {
        this.boardData = window.boardData || {};
        this.isAdmin = this.boardData.isAdmin || false;
        this.boardId = this.boardData.boardId;
        this.publicLink = this.boardData.publicLink;
        this.init();
    }

    init() {
        this.bindEvents();
        this.setupIdeaManager();
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
}

// Initialize board view
document.addEventListener('DOMContentLoaded', () => {
    window.boardView = new BoardView();
});