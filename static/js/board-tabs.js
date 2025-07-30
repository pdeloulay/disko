// Board Tab Navigation functionality
class BoardTabs {
    constructor() {
        this.activeTab = 'board';
        this.init();
    }

    init() {
        this.bindEvents();
        this.updateTabCounts();
    }

    bindEvents() {
        // Tab buttons
        const tabButtons = document.querySelectorAll('.tab-button');
        tabButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const tabName = e.currentTarget.dataset.tab;
                this.switchTab(tabName);
            });
        });
    }

    switchTab(tabName) {
        if (this.activeTab === tabName) return;

        // Update active tab
        this.activeTab = tabName;

        // Update tab buttons
        const tabButtons = document.querySelectorAll('.tab-button');
        tabButtons.forEach(button => {
            if (button.dataset.tab === tabName) {
                button.classList.add('active');
            } else {
                button.classList.remove('active');
            }
        });

        // Update tab content
        const tabContents = document.querySelectorAll('.tab-content');
        tabContents.forEach(content => {
            if (content.id === `${tabName}-view`) {
                content.classList.add('active');
            } else {
                content.classList.remove('active');
            }
        });

        // Load content for the active tab
        if (tabName === 'release') {
            console.log('[BoardTabs] Switching to release tab');
            console.log('[BoardTabs] Release table available:', !!window.releaseTable);
            // Load release table if not already loaded
            if (window.releaseTable) {
                console.log('[BoardTabs] Calling release table refresh');
                window.releaseTable.refresh();
            } else {
                console.log('[BoardTabs] Release table not available, creating new instance');
                window.releaseTable = new ReleaseTable();
            }
        } else if (tabName === 'board') {
            console.log('[BoardTabs] Switching to board tab');
            // Refresh board if needed
            if (window.dragDropBoard) {
                window.dragDropBoard.refreshBoard();
            }
        }
    }

    updateTabCounts() {
        // Update board count
        this.updateBoardCount();
        
        // Update release count (will be updated by release table)
        this.updateReleaseCount();
    }

    updateBoardCount() {
        // This will be called by the drag-drop board when it loads
        if (window.dragDropBoard && window.dragDropBoard.ideas) {
            const boardIdeas = window.dragDropBoard.ideas.filter(idea => 
                idea.column !== 'release'
            );
            const countElement = document.getElementById('board-ideas-count');
            if (countElement) {
                countElement.textContent = boardIdeas.length;
            }
        }
    }

    updateReleaseCount() {
        // This will be called by the release table when it loads
        // The release table handles updating its own count
    }

    // Method to be called by other components to update counts
    setBoardCount(count) {
        const countElement = document.getElementById('board-ideas-count');
        if (countElement) {
            countElement.textContent = count || 0;
        }
    }

    setReleaseCount(count) {
        const countElement = document.getElementById('release-ideas-count');
        if (countElement) {
            countElement.textContent = count || 0;
        }
    }

    // Get current active tab
    getActiveTab() {
        return this.activeTab;
    }
}

// Initialize board tabs when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on a board page with tabs
    if (document.querySelector('.board-tabs')) {
        window.boardTabs = new BoardTabs();
    }
});