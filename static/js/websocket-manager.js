// WebSocket Manager for real-time updates
class WebSocketManager {
    constructor(boardId) {
        this.boardId = boardId;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.isConnected = false;
        this.messageHandlers = new Map();
        this.init();
    }

    init() {
        this.connect();
        this.setupMessageHandlers();
    }

    connect() {
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/ws/boards/${this.boardId}`;
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.onConnectionStatusChange(true);
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.isConnected = false;
                this.onConnectionStatusChange(false);
                this.handleReconnect();
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.handleReconnect();
        }
    }

    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 30000); // Max 30 seconds
            
            console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`);
            
            this.reconnectTimeout = setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
            this.onConnectionStatusChange(false, true);
        }
    }

    setupMessageHandlers() {
        // Handle feedback animations
        this.onMessage('feedback_animation', (data) => {
            this.handleFeedbackAnimation(data);
        });

        // Handle idea updates
        this.onMessage('idea_update', (data) => {
            this.handleIdeaUpdate(data);
        });

        // Handle pong responses
        this.onMessage('pong', () => {
            // Keep-alive response
        });
    }

    handleMessage(message) {
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
            handler(message.data, message);
        } else {
            console.log('Unhandled WebSocket message type:', message.type);
        }
    }

    handleFeedbackAnimation(data) {
        console.log('Feedback animation received:', data);
        
        // Find the idea card
        const ideaCard = document.querySelector(`[data-idea-id="${data.ideaId}"]`);
        if (!ideaCard) {
            console.warn('Idea card not found for animation:', data.ideaId);
            return;
        }

        // Create feedback animation
        this.createFeedbackAnimation(ideaCard, data);
        
        // Update feedback counts (refresh the idea data)
        this.refreshIdeaFeedback(data.ideaId);
    }

    createFeedbackAnimation(ideaCard, data) {
        // Create animation element
        const animation = document.createElement('div');
        animation.className = 'feedback-animation-popup';
        
        let content = '';
        if (data.feedbackType === 'thumbsup') {
            content = '👍 +1';
            animation.classList.add('thumbsup-animation');
        } else if (data.feedbackType === 'emoji') {
            content = `${data.emoji} +1`;
            animation.classList.add('emoji-animation');
        }
        
        animation.textContent = content;
        
        // Position the animation
        const rect = ideaCard.getBoundingClientRect();
        animation.style.position = 'fixed';
        animation.style.left = (rect.left + rect.width / 2) + 'px';
        animation.style.top = (rect.top + rect.height / 2) + 'px';
        animation.style.zIndex = '1000';
        animation.style.pointerEvents = 'none';
        
        document.body.appendChild(animation);
        
        // Add glow effect to idea card
        ideaCard.classList.add('feedback-glow');
        
        // Remove animation and glow after completion
        setTimeout(() => {
            if (document.body.contains(animation)) {
                document.body.removeChild(animation);
            }
            ideaCard.classList.remove('feedback-glow');
        }, 2000);
    }

    async refreshIdeaFeedback(ideaId) {
        // This would typically trigger a refresh of the specific idea's feedback data
        // For now, we'll emit a custom event that the board manager can listen to
        const event = new CustomEvent('feedbackUpdated', {
            detail: { ideaId }
        });
        document.dispatchEvent(event);
    }

    handleIdeaUpdate(data) {
        console.log('Idea update received:', data);
        
        // Emit custom event for idea updates
        const event = new CustomEvent('ideaUpdated', {
            detail: data
        });
        document.dispatchEvent(event);
    }

    onMessage(type, handler) {
        this.messageHandlers.set(type, handler);
    }

    send(message) {
        if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.warn('WebSocket not connected, message not sent:', message);
        }
    }

    sendPing() {
        this.send({ type: 'ping' });
    }

    onConnectionStatusChange(connected, maxAttemptsReached = false) {
        // Update UI to show connection status
        const statusIndicator = document.getElementById('ws-status');
        if (statusIndicator) {
            // Clear any existing content
            statusIndicator.innerHTML = '';
            
            if (connected) {
                statusIndicator.className = 'ws-status connected';
                statusIndicator.textContent = 'Live';
                statusIndicator.title = 'Real-time updates active';
            } else if (maxAttemptsReached) {
                statusIndicator.className = 'ws-status error';
                statusIndicator.innerHTML = 'Offline <button class="retry-btn" onclick="window.wsManager?.retryConnection()">Retry</button>';
                statusIndicator.title = 'Real-time updates unavailable - Click retry to reconnect';
            } else {
                statusIndicator.className = 'ws-status reconnecting';
                statusIndicator.textContent = 'Connecting...';
                statusIndicator.title = 'Reconnecting to real-time updates';
            }
        }

        // Emit custom event
        const event = new CustomEvent('websocketStatusChanged', {
            detail: { connected, maxAttemptsReached }
        });
        document.dispatchEvent(event);
    }

    disconnect() {
        // Clear any pending reconnection attempts
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
        this.stopKeepAlive();
    }

    // Manual retry connection
    retryConnection() {
        this.reconnectAttempts = 0;
        this.connect();
    }

    // Start periodic ping to keep connection alive
    startKeepAlive() {
        this.keepAliveInterval = setInterval(() => {
            if (this.isConnected) {
                this.sendPing();
            }
        }, 30000); // Ping every 30 seconds
    }

    stopKeepAlive() {
        if (this.keepAliveInterval) {
            clearInterval(this.keepAliveInterval);
            this.keepAliveInterval = null;
        }
    }
}

// Export for use in other modules
window.WebSocketManager = WebSocketManager;