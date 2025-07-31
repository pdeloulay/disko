// API utility functions
class API {
    constructor() {
        this.baseURL = '/api';
    }

    async request(endpoint, options = {}) {
        // Check if this is a public endpoint (doesn't require authentication)
        const isPublicEndpoint = endpoint.includes('/public') || 
                               endpoint.includes('/health') || 
                               endpoint.includes('/thumbsup') || 
                               endpoint.includes('/emoji');
        
        // Only wait for Clerk if this is not a public endpoint
        if (!isPublicEndpoint && !window.Clerk) {
            console.log('[API] Clerk not available, waiting...');
            await new Promise(resolve => {
                const checkClerk = setInterval(() => {
                    if (window.Clerk) {
                        clearInterval(checkClerk);
                        resolve();
                    }
                }, 100);
            });
        }
        
        const url = `${this.baseURL}${endpoint}`;
        
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        // Add auth token if user is signed in (skip for public endpoints)
        let token = null;
        
        if (!isPublicEndpoint) {
            // First try to get token from localStorage (for board navigation)
            const storedToken = localStorage.getItem('clerk-db-jwt');
            if (storedToken && storedToken.length > 100) {
                // Check if token is expired by trying to decode it
                try {
                    const payload = JSON.parse(atob(storedToken.split('.')[1]));
                    const currentTime = Math.floor(Date.now() / 1000);
                    
                    if (payload.exp && payload.exp < currentTime) {
                        console.warn('[API] Stored token is expired, clearing');
                        localStorage.removeItem('clerk-db-jwt');
                    } else {
                        token = storedToken;
                    }
                } catch (error) {
                    console.warn('[API] Failed to decode stored token, clearing');
                    localStorage.removeItem('clerk-db-jwt');
                }
            } else if (storedToken) {
                console.warn('[API] Stored token is invalid (too short), clearing');
                localStorage.removeItem('clerk-db-jwt');
            }
            
            // If no stored token, try to get from Clerk session
            if (!token && window.Clerk) {
                try {
                    // Wait for Clerk to be fully loaded
                    if (!window.Clerk.session) {
                        await new Promise(resolve => {
                            const checkSession = setInterval(() => {
                                if (window.Clerk.session) {
                                    clearInterval(checkSession);
                                    resolve();
                                }
                            }, 100);
                        });
                    }
                    
                    token = await window.Clerk.session.getToken();
                    
                    // Validate token format
                    if (!token || token.length < 100) {
                        console.error('[API] Invalid token from Clerk session');
                        token = null;
                    }
                } catch (error) {
                    console.error('[API] Failed to get auth token from Clerk:', error);
                }
            }
        }
        
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('[API] Request failed:', response.status, errorText);
                
                // If it's an authentication error, clear the stored token
                if (response.status === 401) {
                    console.warn('[API] Authentication failed, clearing stored token');
                    localStorage.removeItem('clerk-db-jwt');
                }
                
                throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
            }
            
            const contentType = response.headers.get('content-type');
            
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return await response.text();
        } catch (error) {
            console.error('[API] Request failed:', error);
            throw error;
        }
    }

    // GET request
    async get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    }

    // POST request
    async post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    // PUT request
    async put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    // DELETE request
    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // Health check
    async healthCheck() {
        return this.get('/health');
    }

    // Ping endpoint
    async ping() {
        return this.get('/ping');
    }
}

// Create global API instance
window.api = new API();