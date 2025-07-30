// Landing page functionality
document.addEventListener('DOMContentLoaded', async () => {
    console.log('[Landing] DOM loaded, initializing landing page...');
    
    // Initialize auth buttons immediately
    initializeAuthButtons();
    
    // Wait for auth to be ready
    await window.auth.waitForReady();
    console.log('[Landing] Auth ready, user signed in:', window.auth.isSignedIn());
    
    // Stats are server-side templated and always visible
    // No authentication required to view stats
    
    const getStartedBtn = document.getElementById('get-started-btn');
    const ctaSignupBtn = document.getElementById('cta-signup-btn');
    const ctaDemoBtn = document.getElementById('cta-demo-btn');
    const demoBtn = document.getElementById('demo-btn');
    
    if (getStartedBtn) {
        getStartedBtn.addEventListener('click', () => {
            console.log('[Landing] Get started button clicked');
            if (window.auth.isSignedIn()) {
                console.log('[Landing] User signed in, redirecting to dashboard');
                window.location.href = '/dashboard';
            } else {
                console.log('[Landing] User not signed in, opening sign up');
                window.auth.signUp();
            }
        });
    }
    
    if (ctaSignupBtn) {
        ctaSignupBtn.addEventListener('click', () => {
            console.log('[Landing] CTA signup button clicked');
            if (window.auth.isSignedIn()) {
                window.location.href = '/dashboard';
            } else {
                window.auth.signUp();
            }
        });
    }
    
    if (ctaDemoBtn) {
        ctaDemoBtn.addEventListener('click', () => {
            console.log('[Landing] CTA demo button clicked');
            // TODO: Implement demo functionality
            alert('Demo coming soon!');
        });
    }
    
    if (demoBtn) {
        demoBtn.addEventListener('click', () => {
            console.log('[Landing] Demo button clicked');
            // TODO: Implement demo functionality
            alert('Demo coming soon!');
        });
    }

    // Test API connection
    try {
        const response = await window.api.ping();
        console.log('[Landing] API connection successful:', response);
    } catch (error) {
        console.error('[Landing] API connection failed:', error);
    }

    // Listen for auth state changes to handle redirects
    window.auth.addAuthStateListener((user) => {
        console.log('[Landing] Auth state changed:', user ? 'signed in' : 'signed out');
        if (user) {
            // User signed in, redirect to dashboard
            console.log('[Landing] User signed in on landing page, redirecting to dashboard');
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1000); // Small delay to show success state
        } else {
            // User signed out, but stats remain visible (server-side templated)
            console.log('[Landing] User signed out, stats remain visible');
        }
    });
    
    console.log('[Landing] Landing page initialization complete');
});

// Initialize auth buttons visibility
function initializeAuthButtons() {
    console.log('[Landing] Initializing auth buttons...');
    
    const signInBtn = document.getElementById('sign-in-btn');
    const signUpBtn = document.getElementById('sign-up-btn');
    const dashboardBtn = document.getElementById('dashboard-btn');
    const signOutBtn = document.getElementById('sign-out-btn');
    
    // Show sign in/up buttons by default (they'll be hidden if user is signed in)
    if (signInBtn) {
        signInBtn.style.display = 'inline-block';
        console.log('[Landing] Sign in button shown');
    } else {
        console.error('[Landing] Sign in button not found!');
    }
    if (signUpBtn) {
        signUpBtn.style.display = 'inline-block';
        console.log('[Landing] Sign up button shown');
    } else {
        console.error('[Landing] Sign up button not found!');
    }
    
    // Hide dashboard/sign out buttons by default
    if (dashboardBtn) {
        dashboardBtn.style.display = 'none';
    }
    if (signOutBtn) {
        signOutBtn.style.display = 'none';
    }
    
    // Add click handlers
    if (signInBtn) {
        signInBtn.addEventListener('click', () => {
            console.log('[Landing] Sign in button clicked');
            window.auth.signIn();
        });
    }
    
    if (signUpBtn) {
        signUpBtn.addEventListener('click', () => {
            console.log('[Landing] Sign up button clicked');
            window.auth.signUp();
        });
    }
    
    if (dashboardBtn) {
        dashboardBtn.addEventListener('click', () => {
            console.log('[Landing] Dashboard button clicked');
            window.auth.redirectToDashboard();
        });
    }
    
    if (signOutBtn) {
        signOutBtn.addEventListener('click', () => {
            console.log('[Landing] Sign out button clicked');
            window.auth.signOut();
        });
    }
}

// Stats are server-side templated and always visible
// No authentication required to view stats