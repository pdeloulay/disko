// Authentication utilities
class Auth {
    constructor() {
        this.clerk = null;
        this.user = null;
        this.isLoaded = false;
        this.authStateListeners = [];
    }

    async init() {
        if (!window.Clerk) {
            console.error('Clerk not loaded');
            return;
        }

        this.clerk = window.Clerk;
        
        // Wait for Clerk to load
        await this.clerk.load();
        this.isLoaded = true;
        
        // Get initial user state
        this.user = this.clerk.user;
        
        // Listen for auth state changes
        this.clerk.addListener('user', (user) => {
            this.user = user;
            this.updateUI();
            this.notifyAuthStateListeners();
        });

        // Initial UI update
        this.updateUI();
        this.notifyAuthStateListeners();
    }

    // Add listener for auth state changes
    addAuthStateListener(callback) {
        this.authStateListeners.push(callback);
    }

    // Remove listener for auth state changes
    removeAuthStateListener(callback) {
        const index = this.authStateListeners.indexOf(callback);
        if (index > -1) {
            this.authStateListeners.splice(index, 1);
        }
    }

    // Notify all auth state listeners
    notifyAuthStateListeners() {
        this.authStateListeners.forEach(callback => {
            try {
                callback(this.user);
            } catch (error) {
                console.error('Error in auth state listener:', error);
            }
        });
    }

    updateUI() {
        const signInBtn = document.getElementById('sign-in-btn');
        const signUpBtn = document.getElementById('sign-up-btn');
        const dashboardBtn = document.getElementById('dashboard-btn');
        const signOutBtn = document.getElementById('sign-out-btn');
        const userName = document.getElementById('user-name');

        if (this.user) {
            // User is signed in
            if (signInBtn) signInBtn.style.display = 'none';
            if (signUpBtn) signUpBtn.style.display = 'none';
            if (dashboardBtn) dashboardBtn.style.display = 'inline-block';
            if (signOutBtn) signOutBtn.style.display = 'inline-block';
            if (userName) userName.textContent = this.user.firstName || this.user.emailAddresses[0].emailAddress;
        } else {
            // User is signed out
            if (signInBtn) signInBtn.style.display = 'inline-block';
            if (signUpBtn) signUpBtn.style.display = 'inline-block';
            if (dashboardBtn) dashboardBtn.style.display = 'none';
            if (signOutBtn) signOutBtn.style.display = 'none';
            if (userName) userName.textContent = '';
        }
    }

    async signIn() {
        if (!this.clerk) return;
        await this.clerk.openSignIn();
    }

    async signUp() {
        if (!this.clerk) return;
        await this.clerk.openSignUp();
    }

    async signOut() {
        if (!this.clerk) return;
        await this.clerk.signOut();
        window.location.href = '/';
    }

    isSignedIn() {
        return !!this.user;
    }

    requireAuth() {
        if (!this.isSignedIn()) {
            this.signIn();
            return false;
        }
        return true;
    }

    redirectToDashboard() {
        window.location.href = '/dashboard';
    }

    // Get the current JWT token
    async getToken() {
        if (!this.clerk || !this.clerk.session) {
            throw new Error('No active session');
        }
        
        try {
            return await this.clerk.session.getToken();
        } catch (error) {
            console.error('Failed to get token:', error);
            throw error;
        }
    }

    // Get user information
    getUserInfo() {
        if (!this.user) {
            return null;
        }

        return {
            id: this.user.id,
            email: this.user.primaryEmailAddress?.emailAddress,
            firstName: this.user.firstName,
            lastName: this.user.lastName,
            fullName: this.user.fullName,
            imageUrl: this.user.imageUrl
        };
    }

    // Check if Clerk is loaded and ready
    isReady() {
        return this.isLoaded && this.clerk;
    }

    // Wait for auth to be ready
    async waitForReady() {
        return new Promise((resolve) => {
            if (this.isReady()) {
                resolve();
                return;
            }

            const checkReady = () => {
                if (this.isReady()) {
                    resolve();
                } else {
                    setTimeout(checkReady, 100);
                }
            };

            checkReady();
        });
    }
}

// Create global auth instance
window.auth = new Auth();

// Initialize auth when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Wait a bit for Clerk to load
    setTimeout(() => {
        window.auth.init();
    }, 100);
});

// Global event listeners for auth buttons
document.addEventListener('click', (e) => {
    if (e.target.id === 'sign-in-btn') {
        window.auth.signIn();
    } else if (e.target.id === 'sign-up-btn') {
        window.auth.signUp();
    } else if (e.target.id === 'sign-out-btn') {
        window.auth.signOut();
    } else if (e.target.id === 'dashboard-btn') {
        window.auth.redirectToDashboard();
    }
});