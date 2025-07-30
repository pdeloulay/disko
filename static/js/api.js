// API utility functions
class API {
    constructor() {
        this.baseURL = '/api';
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        
        console.log('[API] Making request to:', url);
        console.log('[API] Request method:', options.method || 'GET');
        console.log('[API] Request headers:', options.headers);
        
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        // Add auth token if user is signed in
        let token = null;
        
        // First try to get token from localStorage (for board navigation)
        const storedToken = localStorage.getItem('clerk-db-jwt');
        if (storedToken && storedToken.length > 100) {
            // Check if token is expired by trying to decode it
            try {
                const payload = JSON.parse(atob(storedToken.split('.')[1]));
                const currentTime = Math.floor(Date.now() / 1000);
                
                if (payload.exp && payload.exp < currentTime) {
                    console.warn('[API] Stored token is expired, clearing:', storedToken);
                    localStorage.removeItem('clerk-db-jwt');
                } else {
                    token = storedToken;
                    console.log('[API] Using stored auth token from localStorage, length:', token.length);
                }
            } catch (error) {
                console.warn('[API] Failed to decode stored token, clearing:', error);
                localStorage.removeItem('clerk-db-jwt');
            }
        } else if (storedToken) {
            console.warn('[API] Stored token is invalid (too short), clearing:', storedToken);
            localStorage.removeItem('clerk-db-jwt');
        }
        
        // If no stored token, try to get from Clerk session
        if (!token && window.Clerk && window.Clerk.user) {
            try {
                token = await window.Clerk.session.getToken();
                console.log('[API] Got auth token from Clerk session, length:', token ? token.length : 0);
                
                // Validate token format
                if (!token || token.length < 100) {
                    console.error('[API] Invalid token from Clerk session:', token);
                    token = null;
                }
            } catch (error) {
                console.error('[API] Failed to get auth token from Clerk:', error);
            }
        }
        
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
            console.log('[API] Added auth token to request');
        } else {
            console.log('[API] No auth token available, proceeding without auth');
        }

        try {
            console.log('[API] Sending fetch request to:', url);
            const response = await fetch(url, config);
            
            console.log('[API] Response status:', response.status);
            console.log('[API] Response headers:', Object.fromEntries(response.headers.entries()));
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('[API] Response not ok:', response.status, errorText);
                
                // If it's an authentication error, clear the stored token
                if (response.status === 401) {
                    console.warn('[API] Authentication failed, clearing stored token');
                    localStorage.removeItem('clerk-db-jwt');
                }
                
                throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
            }
            
            const contentType = response.headers.get('content-type');
            console.log('[API] Response content-type:', contentType);
            
            if (contentType && contentType.includes('application/json')) {
                const jsonResponse = await response.json();
                console.log('[API] JSON response received:', jsonResponse);
                return jsonResponse;
            }
            
            const textResponse = await response.text();
            console.log('[API] Text response received:', textResponse);
            return textResponse;
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