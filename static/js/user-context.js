// User context management
class UserContext {
    constructor() {
        this.user = null;
        this.isLoading = true;
        this.listeners = [];
    }

    // Initialize user context
    async init() {
        await window.auth.waitForReady();
        
        // Get initial user state
        this.user = window.auth.getUserInfo();
        this.isLoading = false;
        
        // Listen for auth state changes
        window.auth.addAuthStateListener((clerkUser) => {
            this.user = window.auth.getUserInfo();
            this.notifyListeners();
        });

        // Initial notification
        this.notifyListeners();
    }

    // Add listener for user context changes
    addListener(callback) {
        this.listeners.push(callback);
        
        // Immediately call with current state if not loading
        if (!this.isLoading) {
            callback(this.user);
        }
    }

    // Remove listener
    removeListener(callback) {
        const index = this.listeners.indexOf(callback);
        if (index > -1) {
            this.listeners.splice(index, 1);
        }
    }

    // Notify all listeners
    notifyListeners() {
        this.listeners.forEach(callback => {
            try {
                callback(this.user);
            } catch (error) {
                console.error('Error in user context listener:', error);
            }
        });
    }

    // Get current user
    getUser() {
        return this.user;
    }

    // Check if user is authenticated
    isAuthenticated() {
        return !!this.user;
    }

    // Check if context is loading
    isContextLoading() {
        return this.isLoading;
    }

    // Get user display name
    getDisplayName() {
        if (!this.user) return null;
        
        return this.user.fullName || 
               `${this.user.firstName || ''} ${this.user.lastName || ''}`.trim() ||
               this.user.email ||
               'User';
    }

    // Get user avatar URL
    getAvatarUrl() {
        return this.user?.imageUrl || null;
    }

    // Get user email
    getEmail() {
        return this.user?.email || null;
    }

    // Get user ID
    getUserId() {
        return this.user?.id || null;
    }
}

// Create global user context instance
window.userContext = new UserContext();

// Auto-initialize user context when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Wait a bit for auth to initialize first
    setTimeout(async () => {
        try {
            await window.userContext.init();
        } catch (error) {
            console.error('User context initialization failed:', error);
        }
    }, 300);
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UserContext;
}