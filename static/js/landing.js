// Landing page functionality
document.addEventListener('DOMContentLoaded', async () => {
    // Wait for auth to be ready
    await window.auth.waitForReady();
    
    const getStartedBtn = document.getElementById('get-started-btn');
    
    if (getStartedBtn) {
        getStartedBtn.addEventListener('click', () => {
            if (window.auth.isSignedIn()) {
                window.location.href = '/dashboard';
            } else {
                window.auth.signUp();
            }
        });
    }

    // Test API connection
    try {
        const response = await window.api.ping();
        console.log('API connection successful:', response);
    } catch (error) {
        console.error('API connection failed:', error);
    }

    // Listen for auth state changes to handle redirects
    window.auth.addAuthStateListener((user) => {
        if (user && window.location.pathname === '/') {
            // User signed in on landing page, redirect to dashboard
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1000); // Small delay to show success state
        }
    });
});