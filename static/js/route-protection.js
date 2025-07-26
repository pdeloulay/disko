// Route protection utilities
class RouteProtection {
    constructor() {
        this.protectedRoutes = ['/dashboard', '/board'];
        this.publicRoutes = ['/'];
        this.authRequiredMessage = 'Please sign in to access this page.';
    }

    // Check if current route requires authentication
    isProtectedRoute(path = window.location.pathname) {
        return this.protectedRoutes.some(route => path.startsWith(route));
    }

    // Check if current route is public
    isPublicRoute(path = window.location.pathname) {
        return this.publicRoutes.some(route => path === route);
    }

    // Protect the current page
    async protectPage() {
        await window.auth.waitForReady();

        const currentPath = window.location.pathname;
        
        // If it's a protected route and user is not signed in
        if (this.isProtectedRoute(currentPath) && !window.auth.isSignedIn()) {
            this.redirectToSignIn();
            return false;
        }

        // If it's the landing page and user is signed in, redirect to dashboard
        if (currentPath === '/' && window.auth.isSignedIn()) {
            this.redirectToDashboard();
            return false;
        }

        return true;
    }

    // Redirect to sign in
    redirectToSignIn() {
        // Store the intended destination
        sessionStorage.setItem('redirectAfterAuth', window.location.pathname);
        
        // Show sign in modal or redirect to landing page
        if (window.auth.isReady()) {
            window.auth.signIn();
        } else {
            window.location.href = '/';
        }
    }

    // Redirect to dashboard
    redirectToDashboard() {
        window.location.href = '/dashboard';
    }

    // Handle post-authentication redirect
    handlePostAuthRedirect() {
        const redirectPath = sessionStorage.getItem('redirectAfterAuth');
        if (redirectPath && redirectPath !== '/') {
            sessionStorage.removeItem('redirectAfterAuth');
            window.location.href = redirectPath;
        } else {
            this.redirectToDashboard();
        }
    }

    // Initialize route protection for the current page
    async init() {
        // Wait for auth to be ready
        await window.auth.waitForReady();

        // Listen for auth state changes
        window.auth.addAuthStateListener((user) => {
            this.handleAuthStateChange(user);
        });

        // Protect the current page
        return await this.protectPage();
    }

    // Handle auth state changes
    handleAuthStateChange(user) {
        const currentPath = window.location.pathname;

        if (user) {
            // User signed in
            if (currentPath === '/') {
                // If on landing page, redirect to dashboard
                this.handlePostAuthRedirect();
            }
        } else {
            // User signed out
            if (this.isProtectedRoute(currentPath)) {
                // If on protected route, redirect to landing page
                window.location.href = '/';
            }
        }
    }
}

// Create global route protection instance
window.routeProtection = new RouteProtection();

// Auto-initialize route protection when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Wait a bit for auth to initialize
    setTimeout(async () => {
        try {
            await window.routeProtection.init();
        } catch (error) {
            console.error('Route protection initialization failed:', error);
        }
    }, 200);
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RouteProtection;
}