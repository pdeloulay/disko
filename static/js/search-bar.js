// Search Bar functionality with debouncing and sorting
class SearchBar {
    constructor(boardId, onSearchResults) {
        this.boardId = boardId;
        this.onSearchResults = onSearchResults;
        this.searchTimeout = null;
        this.currentQuery = '';
        this.currentFilters = {
            column: '',
            status: '',
            inProgress: null
        };
        this.currentSort = {
            by: '',
            direction: 'asc'
        };
        this.init();
    }

    init() {
        this.createSearchUI();
        this.bindEvents();
    }

    createSearchUI() {
        // Create search container if it doesn't exist
        let searchContainer = document.getElementById('search-container');
        if (!searchContainer) {
            searchContainer = document.createElement('div');
            searchContainer.id = 'search-container';
            searchContainer.className = 'search-container';
            
            // Insert after the board header
            const boardHeader = document.querySelector('.board-header');
            if (boardHeader) {
                boardHeader.insertAdjacentElement('afterend', searchContainer);
            } else {
                // Fallback: insert at the beginning of main content
                const mainContent = document.querySelector('.main-content') || document.body;
                mainContent.insertBefore(searchContainer, mainContent.firstChild);
            }
        }

        searchContainer.innerHTML = `
            <div class="search-bar-wrapper">
                <div class="search-input-group">
                    <input type="text" 
                           id="search-input" 
                           class="search-input" 
                           placeholder="Search ideas..." 
                           value="${this.currentQuery}">
                    <button id="clear-search" class="clear-search-btn" title="Clear search">
                        <span class="clear-icon">×</span>
                    </button>
                </div>
                
                <div class="search-filters">
                    <select id="column-filter" class="filter-select">
                        <option value="">All Columns</option>
                        <option value="parking">Parking</option>
                        <option value="now">Now</option>
                        <option value="next">Next</option>
                        <option value="later">Later</option>
                        <option value="release">Release</option>
                        <option value="wont-do">Won't Do</option>
                    </select>
                    
                    <select id="status-filter" class="filter-select">
                        <option value="">All Status</option>
                        <option value="active">Active</option>
                        <option value="done">Done</option>
                        <option value="archived">Archived</option>
                    </select>
                    
                    <select id="progress-filter" class="filter-select">
                        <option value="">All Progress</option>
                        <option value="true">In Progress</option>
                        <option value="false">Not In Progress</option>
                    </select>
                </div>
                
                <div class="search-sorting">
                    <select id="sort-by" class="sort-select">
                        <option value="">Default Order</option>
                        <option value="name">Name</option>
                        <option value="rice">RICE Score</option>
                        <option value="status">Status</option>
                        <option value="created">Created Date</option>
                    </select>
                    
                    <button id="sort-direction" class="sort-direction-btn" title="Toggle sort direction">
                        <span class="sort-icon">↑</span>
                    </button>
                </div>
                
                <div class="search-results-info">
                    <span id="search-results-count"></span>
                </div>
            </div>
        `;
    }

    bindEvents() {
        const searchInput = document.getElementById('search-input');
        const clearSearchBtn = document.getElementById('clear-search');
        const columnFilter = document.getElementById('column-filter');
        const statusFilter = document.getElementById('status-filter');
        const progressFilter = document.getElementById('progress-filter');
        const sortBy = document.getElementById('sort-by');
        const sortDirection = document.getElementById('sort-direction');

        // Search input with debouncing
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.currentQuery = e.target.value;
                this.debouncedSearch();
            });

            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.performSearch();
                }
            });
        }

        // Clear search button
        if (clearSearchBtn) {
            clearSearchBtn.addEventListener('click', () => {
                this.clearSearch();
            });
        }

        // Filter changes
        if (columnFilter) {
            columnFilter.addEventListener('change', (e) => {
                this.currentFilters.column = e.target.value;
                this.performSearch();
            });
        }

        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.currentFilters.status = e.target.value;
                this.performSearch();
            });
        }

        if (progressFilter) {
            progressFilter.addEventListener('change', (e) => {
                const value = e.target.value;
                this.currentFilters.inProgress = value === '' ? null : value === 'true';
                this.performSearch();
            });
        }

        // Sort changes
        if (sortBy) {
            sortBy.addEventListener('change', (e) => {
                this.currentSort.by = e.target.value;
                this.performSearch();
            });
        }

        if (sortDirection) {
            sortDirection.addEventListener('click', () => {
                this.currentSort.direction = this.currentSort.direction === 'asc' ? 'desc' : 'asc';
                this.updateSortDirectionUI();
                this.performSearch();
            });
        }
    }

    debouncedSearch() {
        // Clear existing timeout
        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }

        // Set new timeout for debounced search
        this.searchTimeout = setTimeout(() => {
            this.performSearch();
        }, 300); // 300ms delay
    }

    async performSearch() {
        try {
            // Build query parameters
            const params = new URLSearchParams();
            
            if (this.currentQuery.trim()) {
                params.append('q', this.currentQuery.trim());
            }
            
            if (this.currentFilters.column) {
                params.append('column', this.currentFilters.column);
            }
            
            if (this.currentFilters.status) {
                params.append('status', this.currentFilters.status);
            }
            
            if (this.currentFilters.inProgress !== null) {
                params.append('inProgress', this.currentFilters.inProgress.toString());
            }
            
            if (this.currentSort.by) {
                params.append('sortBy', this.currentSort.by);
                params.append('sortDir', this.currentSort.direction);
            }

            // Make API request
            const response = await fetch(`/api/boards/${this.boardId}/search?${params.toString()}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${window.userContext?.sessionToken}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Search failed: ${response.status}`);
            }

            const data = await response.json();
            
            // Update results count
            this.updateResultsCount(data.count, data.query);
            
            // Call the callback with search results
            if (this.onSearchResults) {
                this.onSearchResults(data.ideas, {
                    query: data.query,
                    filters: data.filters,
                    sort: data.sort,
                    count: data.count
                });
            }

        } catch (error) {
            console.error('Search error:', error);
            this.updateResultsCount(0, this.currentQuery, error.message);
        }
    }

    updateResultsCount(count, query, error = null) {
        const resultsInfo = document.getElementById('search-results-count');
        if (!resultsInfo) return;

        if (error) {
            resultsInfo.textContent = `Search error: ${error}`;
            resultsInfo.className = 'search-error';
        } else if (query && query.trim()) {
            resultsInfo.textContent = `Found ${count} result${count !== 1 ? 's' : ''} for "${query}"`;
            resultsInfo.className = 'search-results';
        } else if (this.hasActiveFilters()) {
            resultsInfo.textContent = `${count} idea${count !== 1 ? 's' : ''} match current filters`;
            resultsInfo.className = 'search-results';
        } else {
            resultsInfo.textContent = '';
            resultsInfo.className = '';
        }
    }

    hasActiveFilters() {
        return this.currentFilters.column || 
               this.currentFilters.status || 
               this.currentFilters.inProgress !== null ||
               this.currentSort.by;
    }

    updateSortDirectionUI() {
        const sortDirectionBtn = document.getElementById('sort-direction');
        const sortIcon = sortDirectionBtn?.querySelector('.sort-icon');
        
        if (sortIcon) {
            sortIcon.textContent = this.currentSort.direction === 'asc' ? '↑' : '↓';
            sortDirectionBtn.title = `Sort ${this.currentSort.direction === 'asc' ? 'ascending' : 'descending'}`;
        }
    }

    clearSearch() {
        // Reset all search parameters
        this.currentQuery = '';
        this.currentFilters = {
            column: '',
            status: '',
            inProgress: null
        };
        this.currentSort = {
            by: '',
            direction: 'asc'
        };

        // Reset UI elements
        const searchInput = document.getElementById('search-input');
        const columnFilter = document.getElementById('column-filter');
        const statusFilter = document.getElementById('status-filter');
        const progressFilter = document.getElementById('progress-filter');
        const sortBy = document.getElementById('sort-by');

        if (searchInput) searchInput.value = '';
        if (columnFilter) columnFilter.value = '';
        if (statusFilter) statusFilter.value = '';
        if (progressFilter) progressFilter.value = '';
        if (sortBy) sortBy.value = '';

        this.updateSortDirectionUI();
        this.updateResultsCount(0, '');

        // Perform search to reset results
        this.performSearch();
    }

    // Method to highlight search terms in text
    highlightSearchTerms(text, query) {
        if (!query || !text) return text;
        
        const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
        return text.replace(regex, '<mark class="search-highlight">$1</mark>');
    }

    // Method to set board ID (useful when switching boards)
    setBoardId(boardId) {
        this.boardId = boardId;
    }

    // Method to get current search state
    getCurrentSearchState() {
        return {
            query: this.currentQuery,
            filters: { ...this.currentFilters },
            sort: { ...this.currentSort }
        };
    }

    // Method to restore search state
    restoreSearchState(state) {
        if (!state) return;

        this.currentQuery = state.query || '';
        this.currentFilters = { ...this.currentFilters, ...state.filters };
        this.currentSort = { ...this.currentSort, ...state.sort };

        // Update UI elements
        const searchInput = document.getElementById('search-input');
        const columnFilter = document.getElementById('column-filter');
        const statusFilter = document.getElementById('status-filter');
        const progressFilter = document.getElementById('progress-filter');
        const sortBy = document.getElementById('sort-by');

        if (searchInput) searchInput.value = this.currentQuery;
        if (columnFilter) columnFilter.value = this.currentFilters.column || '';
        if (statusFilter) statusFilter.value = this.currentFilters.status || '';
        if (progressFilter) {
            const progressValue = this.currentFilters.inProgress === null ? '' : 
                                 this.currentFilters.inProgress.toString();
            progressFilter.value = progressValue;
        }
        if (sortBy) sortBy.value = this.currentSort.by || '';

        this.updateSortDirectionUI();
        this.performSearch();
    }
}

// Export for use in other modules
window.SearchBar = SearchBar;