// Authentication utilities
class Auth {
    constructor() {
        this.clerk = null;
        this.user = null;
        this.isLoaded = false;
        this.authStateListeners = [];
    }

    async init() {
        console.log('[Auth] Initializing auth...');
        
        // Get Clerk configuration from global variables set in base.html
        const clerkPublishableKey = window.clerkPublishableKey;
        const clerkFrontendApiUrl = window.clerkFrontendApiUrl;
        
        console.log('[Auth] Clerk publishable key:', clerkPublishableKey ? 'Set' : 'Not set');
        console.log('[Auth] Clerk frontend API URL:', clerkFrontendApiUrl ? 'Set' : 'Not set');
        
        // Wait for Clerk to be available
        let attempts = 0;
        while (!window.Clerk && attempts < 50) {
            console.log('[Auth] Waiting for Clerk to load... attempt', attempts + 1);
            await new Promise(resolve => setTimeout(resolve, 100));
            attempts++;
        }
        
        if (!window.Clerk) {
            console.error('[Auth] Clerk failed to load after 50 attempts');
            return;
        }
        
        this.clerk = window.Clerk;
        console.log('[Auth] Clerk loaded successfully');
        console.log('[Auth] Clerk object type:', typeof this.clerk);
        //console.log('[Auth] Clerk methods:', Object.getOwnPropertyNames(this.clerk));
        //console.log('[Auth] Clerk prototype methods:', Object.getOwnPropertyNames(Object.getPrototypeOf(this.clerk)));
        
        // Initialize Clerk with configuration
        if (clerkPublishableKey) {
            const clerkConfig = {
                publishableKey: clerkPublishableKey
            };
            
            // Add frontend API URL if provided
            if (clerkFrontendApiUrl) {
                clerkConfig.frontendApi = clerkFrontendApiUrl;
                console.log('[Auth] Clerk frontend API URL set:', clerkFrontendApiUrl);
            }
            
            try {
                await this.clerk.load(clerkConfig);
                console.log('[Auth] Clerk initialized with configuration');
            } catch (error) {
                console.error('[Auth] Failed to initialize Clerk:', error);
                return;
            }
        }
        
        // Get the current user
        this.user = this.clerk.user;
        console.log('[Auth] Current user:', this.user ? this.user.id : 'None');
        
        // Set up authentication state listener
        const mainAuthListener = ({ user }) => {
            console.log('[Auth] Auth state changed, user:', user ? user.id : 'None');
            if (user) {
                console.log('[Auth] User authenticated:', user.id);
                console.log('[Auth] User email:', user.emailAddresses[0]?.emailAddress);
            } else {
                console.log('[Auth] User signed out');
            }
            this.user = user;
            this.notifyAuthStateListeners();
            this.updateUI();
        };
        
        // Add main auth listener using the correct method
        if (this.clerk.addListener) {
            this.clerk.addListener(mainAuthListener);
        } else if (this.clerk.on) {
            this.clerk.on('userChanged', mainAuthListener);
        }
        
        // Update UI initially
        this.updateUI();
        this.isLoaded = true;
        console.log('[Auth] Auth initialization complete');
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
                console.error('[Auth] Error in auth state listener:', error);
            }
        });
    }

    updateUI() {
        console.log('[Auth] Updating UI...');
        
        const signInBtn = document.getElementById('sign-in-btn');
        const signUpBtn = document.getElementById('sign-up-btn');
        const dashboardBtn = document.getElementById('dashboard-btn');
        const signOutBtn = document.getElementById('sign-out-btn');
        const userName = document.getElementById('user-name');

        if (this.user) {
            // User is signed in
            console.log('[Auth] User signed in, hiding sign in/up buttons');
            if (signInBtn) signInBtn.style.display = 'none';
            if (signUpBtn) signUpBtn.style.display = 'none';
            if (dashboardBtn) dashboardBtn.style.display = 'inline-block';
            if (signOutBtn) signOutBtn.style.display = 'inline-block';
            if (userName) userName.textContent = this.user.firstName || this.user.emailAddresses[0].emailAddress;
        } else {
            // User is signed out
            console.log('[Auth] User signed out, showing sign in/up buttons');
            if (signInBtn) signInBtn.style.display = 'inline-block';
            if (signUpBtn) signUpBtn.style.display = 'inline-block';
            if (dashboardBtn) dashboardBtn.style.display = 'none';
            if (signOutBtn) signOutBtn.style.display = 'none';
            if (userName) userName.textContent = '';
        }
    }

    async signIn() {
        console.log('[Auth] Opening sign in popup...');
        if (!this.clerk) {
            console.error('[Auth] Clerk not initialized');
            alert('Authentication not ready. Please refresh the page and try again.');
            return;
        }
        
        const currentPath = window.location.pathname;
        console.log('[Auth] Current path:', currentPath);
        console.log('[Auth] Clerk object:', this.clerk);
        //console.log('[Auth] Clerk methods:', Object.getOwnPropertyNames(this.clerk));
        
        try {
            // Check if openSignIn method exists
            if (typeof this.clerk.openSignIn !== 'function') {
                console.error('[Auth] openSignIn method not available');
                console.log('[Auth] Available methods:', Object.getOwnPropertyNames(this.clerk));
                alert('Sign-in method not available. Please refresh the page.');
                return;
            }
            
            // Use Clerk's built-in sign-in method
            await this.clerk.openSignIn();
            
            // Set up a one-time listener for this sign-in attempt
            const authListener = ({ user }) => {
                if (user) {
                    console.log('[Auth] Sign in successful, redirecting to dashboard...');
                    // Remove listener using the correct method
                    if (this.clerk.removeListener) {
                        this.clerk.removeListener(authListener);
                    } else if (this.clerk.off) {
                        this.clerk.off('userChanged', authListener);
                    }
                    window.location.href = '/dashboard';
                } else {
                    console.log('[Auth] Sign in failed or cancelled');
                    // Remove listener using the correct method
                    if (this.clerk.removeListener) {
                        this.clerk.removeListener(authListener);
                    } else if (this.clerk.off) {
                        this.clerk.off('userChanged', authListener);
                    }
                    // If we're not already on the landing page, go back to "/"
                    if (currentPath !== '/') {
                        console.log('[Auth] Redirecting back to landing page...');
                        window.location.href = '/';
                    }
                }
            };
            
            // Add listener using the correct method
            if (this.clerk.addListener) {
                this.clerk.addListener(authListener);
            } else if (this.clerk.on) {
                this.clerk.on('userChanged', authListener);
            }
        } catch (error) {
            console.error('[Auth] Failed to open sign in popup:', error);
            // Show error to user
            alert('Failed to open sign-in. Please try again.');
        }
    }

    async signUp() {
        console.log('[Auth] Opening sign up popup...');
        if (!this.clerk) {
            console.error('[Auth] Clerk not initialized');
            alert('Authentication not ready. Please refresh the page and try again.');
            return;
        }
        
        const currentPath = window.location.pathname;
        console.log('[Auth] Current path for sign up:', currentPath);
        console.log('[Auth] Clerk object:', this.clerk);
        //console.log('[Auth] Clerk methods:', Object.getOwnPropertyNames(this.clerk));
        
        try {
            // Check if openSignUp method exists
            if (typeof this.clerk.openSignUp !== 'function') {
                console.error('[Auth] openSignUp method not available');
                console.log('[Auth] Available methods:', Object.getOwnPropertyNames(this.clerk));
                alert('Sign-up method not available. Please refresh the page.');
                return;
            }
            
            // Use Clerk's built-in sign-up method
            await this.clerk.openSignUp();
            
            // Set up a one-time listener for this sign-up attempt
            const authListener = ({ user }) => {
                if (user) {
                    console.log('[Auth] Sign up successful, redirecting to dashboard...');
                    // Remove listener using the correct method
                    if (this.clerk.removeListener) {
                        this.clerk.removeListener(authListener);
                    } else if (this.clerk.off) {
                        this.clerk.off('userChanged', authListener);
                    }
                    window.location.href = '/dashboard';
                } else {
                    console.log('[Auth] Sign up failed or cancelled');
                    // Remove listener using the correct method
                    if (this.clerk.removeListener) {
                        this.clerk.removeListener(authListener);
                    } else if (this.clerk.off) {
                        this.clerk.off('userChanged', authListener);
                    }
                    // If we're not already on the landing page, go back to "/"
                    if (currentPath !== '/') {
                        console.log('[Auth] Redirecting back to landing page...');
                        window.location.href = '/';
                    }
                }
            };
            
            // Add listener using the correct method
            if (this.clerk.addListener) {
                this.clerk.addListener(authListener);
            } else if (this.clerk.on) {
                this.clerk.on('userChanged', authListener);
            }
        } catch (error) {
            console.error('[Auth] Failed to open sign up popup:', error);
            // Show error to user
            alert('Failed to open sign-up. Please try again.');
        }
    }

    async signOut() {
        console.log('[Auth] Signing out...');
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
            console.error('[Auth] Failed to get token:', error);
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

    // Check if Clerk components are ready
    areComponentsReady() {
        return this.clerk && this.clerk.isReady && this.clerk.isReady();
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

    // Wait for Clerk components to be ready
    async waitForComponentsReady() {
        return new Promise((resolve) => {
            if (this.areComponentsReady()) {
                resolve();
                return;
            }

            const checkComponentsReady = () => {
                if (this.areComponentsReady()) {
                    resolve();
                } else {
                    setTimeout(checkComponentsReady, 200);
                }
            };

            checkComponentsReady();
        });
    }
}

// Create global auth instance
window.auth = new Auth();

// Initialize auth when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('[Auth] DOM loaded, initializing auth...');
    
    // Show auth buttons immediately (they'll be updated when auth loads)
    const signInBtn = document.getElementById('sign-in-btn');
    const signUpBtn = document.getElementById('sign-up-btn');
    const dashboardBtn = document.getElementById('dashboard-btn');
    const signOutBtn = document.getElementById('sign-out-btn');
    
    if (signInBtn) {
        signInBtn.style.display = 'inline-block';
        console.log('[Auth] Sign in button shown immediately');
    }
    if (signUpBtn) {
        signUpBtn.style.display = 'inline-block';
        console.log('[Auth] Sign up button shown immediately');
    }
    if (dashboardBtn) {
        dashboardBtn.style.display = 'none';
    }
    if (signOutBtn) {
        signOutBtn.style.display = 'none';
    }
    
    // Wait for Clerk to be available and then initialize
    const initAuth = async () => {
        if (window.Clerk) {
            console.log('[Auth] Clerk available, initializing...');
            await window.auth.init();
        } else {
            console.log('[Auth] Clerk not yet available, retrying in 100ms...');
            setTimeout(initAuth, 100);
        }
    };
    
    // Start the initialization process
    initAuth();
});

// Global event listeners for auth buttons
document.addEventListener('click', (e) => {
    if (e.target.id === 'sign-in-btn') {
        console.log('[Auth] Sign in button clicked via global listener');
        window.auth.signIn();
    } else if (e.target.id === 'sign-up-btn') {
        console.log('[Auth] Sign up button clicked via global listener');
        window.auth.signUp();
    } else if (e.target.id === 'sign-out-btn') {
        console.log('[Auth] Sign out button clicked via global listener');
        window.auth.signOut();
    } else if (e.target.id === 'dashboard-btn') {
        console.log('[Auth] Dashboard button clicked via global listener');
        window.auth.redirectToDashboard();
    }
});