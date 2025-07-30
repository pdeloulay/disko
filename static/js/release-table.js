// Release Table functionality
class ReleaseTable {
    constructor() {
        this.boardData = null;
        this.boardId = null;
        this.publicLink = null;
        this.isAdmin = false;
        this.isPublic = false;
        this.currentPage = 1;
        this.pageSize = 20;
        this.currentSearch = '';
        this.currentSort = 'created_at:desc';
        this.debounceTimer = null;
        this.initialized = false;
        this.waitForBoardData();
    }

    async waitForBoardData() {
        // Wait for board data to be available
        while (!window.boardData || !window.boardData.boardId) {
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        
        this.boardData = window.boardData;
        this.boardId = this.boardData.boardId;
        this.publicLink = this.boardData.publicLink;
        this.isAdmin = this.boardData.isAdmin || false;
        this.isPublic = this.boardData.isPublic || false;
        
        console.log('[ReleaseTable] Board data loaded:', this.boardData);
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadReleasedIdeas();
    }

    bindEvents() {
        // Search input
        const searchInput = document.getElementById('release-search');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                clearTimeout(this.debounceTimer);
                this.debounceTimer = setTimeout(() => {
                    this.currentSearch = e.target.value.trim();
                    this.currentPage = 1;
                    this.loadReleasedIdeas();
                }, 300);
            });
        }

        // Clear search button
        const clearSearchBtn = document.getElementById('clear-search');
        if (clearSearchBtn) {
            clearSearchBtn.addEventListener('click', () => {
                searchInput.value = '';
                this.currentSearch = '';
                this.currentPage = 1;
                this.loadReleasedIdeas();
            });
        }

        // Sort dropdown
        const sortSelect = document.getElementById('release-sort');
        if (sortSelect) {
            sortSelect.addEventListener('change', (e) => {
                this.currentSort = e.target.value;
                this.currentPage = 1;
                this.loadReleasedIdeas();
            });
        }
    }

    async loadReleasedIdeas() {
        const container = document.getElementById('release-table-container');
        if (!container) return;

        console.log('[ReleaseTable] Loading released ideas...');
        console.log('[ReleaseTable] Board ID:', this.boardId);
        console.log('[ReleaseTable] Current sort:', this.currentSort);
        console.log('[ReleaseTable] Current search:', this.currentSearch);

        // Show loading state
        container.innerHTML = '<div class="release-loading">Loading released ideas...</div>';

        try {
            const [sortBy, sortDir] = this.currentSort.split(':');
            const params = new URLSearchParams({
                page: this.currentPage,
                pageSize: this.pageSize,
                sortBy: sortBy,
                sortDir: sortDir
            });

            if (this.currentSearch) {
                params.append('search', this.currentSearch);
            }

            // Use the API utility for proper authentication
            const endpoint = `/boards/${this.boardId}/release?${params}`;
            console.log('[ReleaseTable] Making API call to:', endpoint);
            
            const response = await window.api.get(endpoint);
            console.log('[ReleaseTable] API response received:', response);

            const data = response.data || response;
            console.log('[ReleaseTable] Processed data:', data);

            this.renderReleasedIdeas(data);
            this.renderPagination(data);

        } catch (error) {
            console.error('[ReleaseTable] Error loading released ideas:', error);
            console.error('[ReleaseTable] Error details:', {
                message: error.message,
                status: error.response?.status,
                data: error.response?.data
            });
            container.innerHTML = `
                <div class="error-message">
                    <p>Failed to load released ideas. Please try again.</p>
                    <button onclick="window.releaseTable.loadReleasedIdeas()" class="btn btn-secondary">Retry</button>
                </div>
            `;
        }
    }

    renderReleasedIdeas(data) {
        const container = document.getElementById('release-table-container');
        if (!container) return;

        console.log('[ReleaseTable] Rendering released ideas with data:', data);
        console.log('[ReleaseTable] Ideas array:', data.ideas);
        console.log('[ReleaseTable] Ideas count:', data.ideas?.length);

        if (!data.ideas || data.ideas.length === 0) {
            console.log('[ReleaseTable] No ideas found, showing empty state');
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">🚀</div>
                    <h3>No Released Ideas</h3>
                    <p>${this.currentSearch ? 'No ideas match your search criteria.' : 'No ideas have been released yet.'}</p>
                </div>
            `;
            return;
        }

        console.log('[ReleaseTable] Rendering table with', data.ideas.length, 'ideas');

        // Create table
        const table = document.createElement('table');
        table.className = 'release-table';

        // Create header
        const thead = document.createElement('thead');
        thead.innerHTML = `
            <tr>
                <th class="col-name">Idea</th>
                <th class="col-description">Description</th>
                <th class="col-value">Value Statement</th>
                ${!this.isPublic ? '<th class="col-rice">RICE Score</th>' : ''}
                <th class="col-feedback">Feedback</th>
                <th class="col-date">Released</th>
            </tr>
        `;
        table.appendChild(thead);

        // Create body
        const tbody = document.createElement('tbody');
        data.ideas.forEach(idea => {
            const row = this.createIdeaRow(idea);
            tbody.appendChild(row);
        });
        table.appendChild(tbody);

        container.innerHTML = '';
        container.appendChild(table);

        // Update release count in tab
        this.updateReleaseCount(data.totalCount);
    }

    createIdeaRow(idea) {
        const row = document.createElement('tr');
        row.className = 'idea-row';

        // Calculate RICE score if admin
        let riceScore = 0;
        if (!this.isPublic && idea.riceScore) {
            const { reach, impact, confidence, effort } = idea.riceScore;
            if (effort > 0) {
                riceScore = ((reach * impact * confidence) / effort).toFixed(1);
            }
        }

        // Format date
        const releaseDate = new Date(idea.createdAt).toLocaleDateString();

        // Format emoji reactions
        let emojiDisplay = '';
        if (idea.emojiReactions && idea.emojiReactions.length > 0) {
            emojiDisplay = idea.emojiReactions
                .map(reaction => `${reaction.emoji} ${reaction.count}`)
                .join(' ');
        }

        row.innerHTML = `
            <td class="col-name">
                <div class="idea-title">${this.escapeHtml(idea.oneLiner)}</div>
            </td>
            <td class="col-description">
                <div class="idea-description">${this.escapeHtml(idea.description)}</div>
            </td>
            <td class="col-value">
                <div class="idea-value">${this.escapeHtml(idea.valueStatement)}</div>
            </td>
            ${!this.isPublic ? `<td class="col-rice"><div class="rice-score">${riceScore}</div></td>` : ''}
            <td class="col-feedback">
                <div class="feedback-display">
                    <span class="thumbs-up">👍 ${idea.thumbsUp}</span>
                    ${emojiDisplay ? `<span class="emoji-reactions">${emojiDisplay}</span>` : ''}
                </div>
            </td>
            <td class="col-date">
                <div class="release-date">${releaseDate}</div>
            </td>
        `;

        return row;
    }

    renderPagination(data) {
        const container = document.getElementById('release-pagination');
        if (!container) return;

        if (data.totalPages <= 1) {
            container.innerHTML = '';
            return;
        }

        const pagination = document.createElement('div');
        pagination.className = 'pagination';

        // Previous button
        if (data.page > 1) {
            const prevBtn = document.createElement('button');
            prevBtn.className = 'pagination-btn';
            prevBtn.textContent = '← Previous';
            prevBtn.addEventListener('click', () => {
                this.currentPage = data.page - 1;
                this.loadReleasedIdeas();
            });
            pagination.appendChild(prevBtn);
        }

        // Page info
        const pageInfo = document.createElement('span');
        pageInfo.className = 'pagination-info';
        pageInfo.textContent = `Page ${data.page} of ${data.totalPages} (${data.totalCount} total)`;
        pagination.appendChild(pageInfo);

        // Next button
        if (data.page < data.totalPages) {
            const nextBtn = document.createElement('button');
            nextBtn.className = 'pagination-btn';
            nextBtn.textContent = 'Next →';
            nextBtn.addEventListener('click', () => {
                this.currentPage = data.page + 1;
                this.loadReleasedIdeas();
            });
            pagination.appendChild(nextBtn);
        }

        container.innerHTML = '';
        container.appendChild(pagination);
    }

    updateReleaseCount(count) {
        const countElement = document.getElementById('release-ideas-count');
        if (countElement) {
            countElement.textContent = count || 0;
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Method to refresh the release table (called from other components)
    refresh() {
        if (this.boardId) {
            this.loadReleasedIdeas();
        } else {
            console.log('[ReleaseTable] Not initialized yet, waiting for board data...');
        }
    }
}

// Initialize release table when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if we're on a board page
    if (document.getElementById('release-view')) {
        window.releaseTable = new ReleaseTable();
    }
});