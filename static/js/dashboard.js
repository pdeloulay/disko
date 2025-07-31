// Dashboard functionality
document.addEventListener('DOMContentLoaded', async () => {
    console.log('[Dashboard] DOM loaded, initializing dashboard...');
    
    try {
        // Load user info (boards will be loaded after Clerk initialization)
        
    } catch (error) {
        console.error('[Dashboard] Error during initialization:', error);
        console.log('[Dashboard] Showing demo mode...');
    }

    // Create board button
    const createBoardBtn = document.getElementById('create-board-btn');
    if (createBoardBtn) {
        createBoardBtn.addEventListener('click', () => {
            console.log('[Dashboard] Create board button clicked');
            openCreateBoardModal();
        });
    }

    // Create board form
    const createBoardForm = document.getElementById('create-board-form');
    if (createBoardForm) {
        createBoardForm.addEventListener('submit', handleCreateBoard);
    }

    // Modal close functionality
    const modalClose = document.querySelector('.modal-close');
    if (modalClose) {
        modalClose.addEventListener('click', closeCreateBoardModal);
    }

    // Close modal when clicking outside
    const modal = document.getElementById('create-board-modal');
    if (modal) {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                closeCreateBoardModal();
            }
        });
    }

    // Delete confirmation modal close functionality
    const deleteModalClose = document.querySelector('#delete-board-modal .modal-close');
    if (deleteModalClose) {
        deleteModalClose.addEventListener('click', closeDeleteBoardModal);
    }

    // Close delete modal when clicking outside
    const deleteModal = document.getElementById('delete-board-modal');
    if (deleteModal) {
        deleteModal.addEventListener('click', (e) => {
            if (e.target === deleteModal) {
                closeDeleteBoardModal();
            }
        });
    }
    
    console.log('[Dashboard] Dashboard initialization complete');
});



// Update dashboard stats
function updateDashboardStats(boardsCount = 0, ideasCount = 0) {
    const boardsCountElement = document.getElementById('boards-count');
    const ideasCountElement = document.getElementById('ideas-count');
    const boardsLabelElement = document.querySelector('.stat-item:first-child .stat-label');
    const ideasLabelElement = document.querySelector('.stat-item:last-child .stat-label');
    
    if (boardsCountElement) {
        boardsCountElement.textContent = boardsCount;
    }
    
    if (ideasCountElement) {
        ideasCountElement.textContent = ideasCount;
    }
    
    // Update labels for singular vs plural
    if (boardsLabelElement) {
        boardsLabelElement.textContent = boardsCount === 1 ? 'Total Board' : 'Total Boards';
    }
    
    if (ideasLabelElement) {
        ideasLabelElement.textContent = ideasCount === 1 ? 'Total Idea' : 'Total Ideas';
    }
}

async function loadBoards() {
    const boardsList = document.getElementById('boards-list');
    
    // Prevent infinite retry loops
    if (window.boardLoadAttempts === undefined) {
        window.boardLoadAttempts = 0;
    }
    
    if (window.boardLoadAttempts > 3) {
        console.error('[Dashboard] Too many board load attempts, stopping retry loop');
        boardsList.innerHTML = `
            <div class="error-state">
                <h3>Failed to load boards</h3>
                <p>There was an error loading your boards. Please try again.</p>
                <button class="btn btn-primary" onclick="location.reload()">Refresh Page</button>
            </div>
        `;
        return;
    }
    
    window.boardLoadAttempts++;
    
    try {
        console.log('[Dashboard] Starting loadBoards function... (attempt', window.boardLoadAttempts, ')');
        boardsList.innerHTML = '<div class="loading">Loading your boards...</div>';
        console.log('[Dashboard] Making API call to /boards...');

        const response = await window.api.get('/boards');
        console.log('[Dashboard] API response received:', response);
        
        // Reset attempts on success
        window.boardLoadAttempts = 0;
        
        const data = response.data || response;
        console.log('[Dashboard] Processed data:', data);
        if (!data.boards || data.boards.length === 0) {
            console.log('[Dashboard] No boards found in response');
            updateDashboardStats(0, 0);
            boardsList.innerHTML = `
                <div class="empty-state">
                    <h3>No boards yet</h3>
                    <p>Create your first board to get started!</p>
                    <button class="btn btn-primary" onclick="openCreateBoardModal()">Create Your First Board</button>
                </div>
            `;
            return;
        }
        console.log('[Dashboard] Found boards:', data.boards.length);
        console.log('[Dashboard] Board details:', data.boards);
        const totalIdeas = data.boards.reduce((total, board) => {
            return total + (board.ideasCount || 0);
        }, 0);
        updateDashboardStats(data.boards.length, totalIdeas);
        boardsList.innerHTML = data.boards.map(board => createBoardCard(board)).join('');
        console.log('[Dashboard] Boards rendered successfully');
    } catch (error) {
        console.error('[Dashboard] Failed to load boards:', error);
        console.error('[Dashboard] Error details:', {
            message: error.message,
            stack: error.stack,
            response: error.response
        });
        
        updateDashboardStats(0, 0);
        boardsList.innerHTML = `
            <div class="error-state">
                <h3>Failed to load boards</h3>
                <p>There was an error loading your boards. Please try again.</p>
                <button class="btn btn-primary" onclick="loadBoards()">Retry</button>
            </div>
        `;
    }
}

function openCreateBoardModal() {
    console.log('[Dashboard] openCreateBoardModal called');
    const modal = document.getElementById('create-board-modal');
    console.log('[Dashboard] Modal element:', modal);
    if (modal) {
        console.log('[Dashboard] Adding show class to modal');
        modal.classList.add('show');
        console.log('[Dashboard] Modal classes:', modal.className);
    } else {
        console.error('[Dashboard] Modal element not found!');
    }
}

function closeCreateBoardModal() {
    const modal = document.getElementById('create-board-modal');
    if (modal) {
        modal.classList.remove('show');
    }
    
    // Reset form
    const form = document.getElementById('create-board-form');
    if (form) {
        form.reset();
    }
}

async function handleCreateBoard(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const boardData = {
        name: formData.get('title'),
        description: formData.get('description')
    };

    // Basic validation
    if (!boardData.name || boardData.name.trim().length === 0) {
        showFormError('board-title', 'Board title is required');
        return;
    }

    if (boardData.name.length > 100) {
        showFormError('board-title', 'Board title must be less than 100 characters');
        return;
    }

    if (boardData.description && boardData.description.length > 500) {
        showFormError('board-description', 'Description must be less than 500 characters');
        return;
    }

    try {
        // Disable form during submission
        const submitBtn = e.target.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating...';

        const response = await window.api.post('/boards', boardData);
        
        // Close modal and reload boards
        closeCreateBoardModal();
        await loadBoards();
        
        // Show success message
        showSuccessMessage('Board created successfully!');
        
    } catch (error) {
        console.error('Failed to create board:', error);
        
        // Show error message
        if (error.response && error.response.data && error.response.data.error) {
            const errorData = error.response.data.error;
            if (errorData.code === 'VALIDATION_ERROR') {
                showFormError('board-title', 'Please check your input and try again');
            } else {
                showFormError('board-title', errorData.message || 'Failed to create board');
            }
        } else {
            showFormError('board-title', 'Failed to create board. Please try again.');
        }
    } finally {
        // Re-enable form
        const submitBtn = e.target.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Create Board';
        }
    }
}

// Board card creation
function createBoardCard(board) {
    console.log('[Dashboard] Creating board card for:', board);
    
    // Check what fields are available in the board object
    console.log('[Dashboard] Board fields:', Object.keys(board));
    console.log('[Dashboard] Board ID field:', board.id);
    console.log('[Dashboard] Board ID type:', typeof board.id);
    
    const createdDate = new Date(board.createdAt).toLocaleDateString();
    const publicUrl = `${window.location.origin}/board/${board.publicLink}`;
    
    // Ensure we have a valid board ID
    const boardId = board.id || board._id || board.boardId;
    if (!boardId) {
        console.error('[Dashboard] No valid board ID found:', board);
        return ''; // Return empty string to skip this board
    }
    
    return `
        <div class="board-card" data-board-id="${boardId}">
            <div class="board-card-header">
                <h3>${escapeHtml(board.name)}</h3>
                <div class="board-card-menu">
                    <button class="btn-menu" onclick="toggleBoardMenu('${boardId}')">⋮</button>
                    <div class="board-menu" id="menu-${boardId}" style="display: none;">
                        <button onclick="editBoard('${boardId}')">Edit</button>
                        <button onclick="copyPublicLink('${publicUrl}')">Copy Public Link</button>
                        <button onclick="confirmDeleteBoard('${boardId}', '${escapeHtml(board.name)}')">Delete</button>
                    </div>
                </div>
            </div>
            ${board.description ? `<p class="board-description">${escapeHtml(board.description)}</p>` : ''}
            <div class="board-meta">
                <span class="board-date">Created ${createdDate}</span>
                <span class="board-stats">${board.ideasCount} ideas • ${board.reactionsCount} reactions</span>
            </div>
            <div class="board-actions">
                <button class="btn btn-primary" onclick="viewBoard('${boardId}')">Open Board</button>
            </div>
        </div>
    `;
}

// Board actions
async function viewBoard(boardId) {
    console.log('[Dashboard] View board clicked:', boardId);
    
    // Validate boardId
    if (!boardId || boardId === 'undefined' || boardId === 'null') {
        console.error('[Dashboard] Invalid boardId:', boardId);
        showErrorMessage('Invalid board ID. Please try again.');
        return;
    }
    
    try {
        // Redirect to the board page
        window.location.href = `/board/${boardId}`;
        
    } catch (error) {
        console.error('[Dashboard] Failed to redirect to board:', error);
        showErrorMessage('Failed to access board. Please try again.');
    }
}

function editBoard(boardId) {
    // This will be implemented in later tasks when board editing is added
    console.log('Editing board:', boardId);
    alert('Board editing will be implemented in later tasks');
}

function toggleBoardMenu(boardId) {
    const menu = document.getElementById(`menu-${boardId}`);
    const allMenus = document.querySelectorAll('.board-menu');
    
    // Close all other menus
    allMenus.forEach(m => {
        if (m.id !== `menu-${boardId}`) {
            m.style.display = 'none';
        }
    });
    
    // Toggle current menu
    menu.style.display = menu.style.display === 'none' ? 'block' : 'none';
    
    // Add click outside handler to close menu
    if (menu.style.display === 'block') {
        setTimeout(() => {
            const handleClickOutside = (event) => {
                if (!menu.contains(event.target) && !event.target.closest('.btn-menu')) {
                    menu.style.display = 'none';
                    document.removeEventListener('click', handleClickOutside);
                }
            };
            document.addEventListener('click', handleClickOutside);
        }, 0);
    }
}

function copyPublicLink(url) {
    navigator.clipboard.writeText(url).then(() => {
        showSuccessMessage('Public link copied to clipboard!');
    }).catch(() => {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = url;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        showSuccessMessage('Public link copied to clipboard!');
    });
}

// Board deletion
let boardToDelete = null;

function confirmDeleteBoard(boardId, boardName) {
    boardToDelete = boardId;
    const modal = document.getElementById('delete-board-modal');
    const boardNameSpan = document.getElementById('delete-board-name');
    
    boardNameSpan.textContent = boardName;
    modal.classList.add('show');
    
    // Close board menu
    const menu = document.getElementById(`menu-${boardId}`);
    if (menu) menu.style.display = 'none';
}

function closeDeleteBoardModal() {
    const modal = document.getElementById('delete-board-modal');
    modal.classList.remove('show');
    boardToDelete = null;
}

async function deleteBoard() {
    if (!boardToDelete) return;
    
    try {
        const deleteBtn = document.getElementById('confirm-delete-btn');
        const originalText = deleteBtn.textContent;
        deleteBtn.disabled = true;
        deleteBtn.textContent = 'Deleting...';
        
        await window.api.delete(`/boards/${boardToDelete}`);
        
        closeDeleteBoardModal();
        await loadBoards();
        showSuccessMessage('Board deleted successfully!');
        
    } catch (error) {
        console.error('Failed to delete board:', error);
        
        let errorMessage = 'Failed to delete board. Please try again.';
        if (error.response && error.response.data && error.response.data.error) {
            errorMessage = error.response.data.error.message || errorMessage;
        }
        
        showErrorMessage(errorMessage);
    } finally {
        const deleteBtn = document.getElementById('confirm-delete-btn');
        if (deleteBtn) {
            deleteBtn.disabled = false;
            deleteBtn.textContent = 'Delete Board';
        }
    }
}

// Utility functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showFormError(fieldId, message) {
    const field = document.getElementById(fieldId);
    if (!field) return;
    
    // Remove existing error
    const existingError = field.parentNode.querySelector('.form-error');
    if (existingError) {
        existingError.remove();
    }
    
    // Add error message
    const errorDiv = document.createElement('div');
    errorDiv.className = 'form-error';
    errorDiv.textContent = message;
    field.parentNode.appendChild(errorDiv);
    
    // Add error styling
    field.classList.add('error');
    
    // Remove error on input
    field.addEventListener('input', function() {
        field.classList.remove('error');
        const error = field.parentNode.querySelector('.form-error');
        if (error) error.remove();
    }, { once: true });
}

function showSuccessMessage(message) {
    showMessage(message, 'success');
}

function showErrorMessage(message) {
    showMessage(message, 'error');
}

function showMessage(message, type) {
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

// Close menus when clicking outside
document.addEventListener('click', (e) => {
    if (!e.target.closest('.board-card-menu')) {
        const allMenus = document.querySelectorAll('.board-menu');
        allMenus.forEach(menu => {
            menu.style.display = 'none';
        });
    }
});