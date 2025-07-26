// Dashboard functionality
document.addEventListener('DOMContentLoaded', async () => {
    // Wait for auth to be ready
    await window.auth.waitForReady();
    
    // Route protection will handle authentication check
    // If we reach here, user is authenticated
    
    // Load user info and boards
    await loadUserInfo();
    await loadBoards();

    // Create board button
    const createBoardBtn = document.getElementById('create-board-btn');
    if (createBoardBtn) {
        createBoardBtn.addEventListener('click', openCreateBoardModal);
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
});

async function loadUserInfo() {
    try {
        // Use user context for consistent user information
        window.userContext.addListener((user) => {
            const userName = document.getElementById('user-name');
            if (userName) {
                userName.textContent = window.userContext.getDisplayName() || 'User';
            }
        });

        // Test the protected API endpoint
        const response = await window.api.get('/user');
        console.log('User API response:', response);
        
    } catch (error) {
        console.error('Failed to load user info:', error);
    }
}

async function loadBoards() {
    const boardsList = document.getElementById('boards-list');
    
    try {
        boardsList.innerHTML = '<div class="loading">Loading your boards...</div>';
        
        const response = await window.api.get('/boards');
        const data = response.data || response;
        
        if (!data.boards || data.boards.length === 0) {
            boardsList.innerHTML = `
                <div class="empty-state">
                    <h3>No boards yet</h3>
                    <p>Create your first board to get started!</p>
                    <button class="btn btn-primary" onclick="openCreateBoardModal()">Create Your First Board</button>
                </div>
            `;
            return;
        }

        // Render board cards
        boardsList.innerHTML = data.boards.map(board => createBoardCard(board)).join('');
        
    } catch (error) {
        console.error('Failed to load boards:', error);
        boardsList.innerHTML = `
            <div class="error">
                <h3>Failed to load boards</h3>
                <p>Please try again or contact support if the problem persists.</p>
                <button class="btn btn-primary" onclick="loadBoards()">Retry</button>
            </div>
        `;
    }
}

function openCreateBoardModal() {
    const modal = document.getElementById('create-board-modal');
    if (modal) {
        modal.style.display = 'flex';
    }
}

function closeCreateBoardModal() {
    const modal = document.getElementById('create-board-modal');
    if (modal) {
        modal.style.display = 'none';
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
    const createdDate = new Date(board.createdAt).toLocaleDateString();
    const publicUrl = `${window.location.origin}/board/${board.publicLink}`;
    
    return `
        <div class="board-card" data-board-id="${board.id}">
            <div class="board-card-header">
                <h3>${escapeHtml(board.name)}</h3>
                <div class="board-card-menu">
                    <button class="btn-menu" onclick="toggleBoardMenu('${board.id}')">⋮</button>
                    <div class="board-menu" id="menu-${board.id}" style="display: none;">
                        <button onclick="editBoard('${board.id}')">Edit</button>
                        <button onclick="copyPublicLink('${publicUrl}')">Copy Public Link</button>
                        <button onclick="confirmDeleteBoard('${board.id}', '${escapeHtml(board.name)}')">Delete</button>
                    </div>
                </div>
            </div>
            ${board.description ? `<p class="board-description">${escapeHtml(board.description)}</p>` : ''}
            <div class="board-meta">
                <span class="board-date">Created ${createdDate}</span>
                <span class="board-columns">${board.visibleColumns.length} columns</span>
            </div>
            <div class="board-actions">
                <button class="btn btn-primary" onclick="viewBoard('${board.id}')">Open Board</button>
                <button class="btn btn-secondary" onclick="copyPublicLink('${publicUrl}')">Share</button>
            </div>
        </div>
    `;
}

// Board actions
function viewBoard(boardId) {
    // Navigate to board view - this will be implemented in later tasks
    window.location.href = `/board/${boardId}`;
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
    modal.style.display = 'flex';
    
    // Close board menu
    const menu = document.getElementById(`menu-${boardId}`);
    if (menu) menu.style.display = 'none';
}

function closeDeleteBoardModal() {
    const modal = document.getElementById('delete-board-modal');
    modal.style.display = 'none';
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
// Ut
ility functions
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