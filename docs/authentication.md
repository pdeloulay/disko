# Authentication System Documentation

## Overview

The Disko application uses a comprehensive authentication system built with Clerk, providing secure JWT-based authentication for admin users. The system includes both frontend and backend components for managing user authentication and authorization.

## Architecture

### Frontend Components

1. **Auth Class** (`static/js/auth.js`)
   - Manages Clerk integration with improved error handling
   - Handles sign-in/sign-up flows with better user experience
   - Provides authentication state management with reactive updates
   - Offers token retrieval for API calls with validation
   - Includes comprehensive logging and debugging capabilities

2. **User Context** (`static/js/user-context.js`)
   - Centralized user state management with reactive updates
   - Consistent user data across components
   - Enhanced error handling and validation
   - Debug tools for troubleshooting

3. **Route Protection** (`static/js/route-protection.js`)
   - Automatic route-based authentication checks
   - Redirects unauthenticated users with stored destination
   - Handles post-authentication routing
   - Configurable protected and public routes

### Backend Components

1. **Authentication Middleware** (`middleware/auth.go`)
   - JWT token validation using Clerk
   - User context injection
   - Protected route enforcement
   - Comprehensive logging

## Features

### Enhanced Authentication Flow
- **Improved Error Handling**: Better error messages and user feedback
- **Configurable Timeouts**: Adjustable retry logic and timeout settings
- **Comprehensive Logging**: Detailed logging for debugging and monitoring
- **Reactive Updates**: Real-time UI updates based on authentication state
- **Token Validation**: JWT token validation with expiration checking

### User Context Management
- **Centralized State**: Single source of truth for user information
- **Reactive Listeners**: Automatic updates when user state changes
- **Safe Property Access**: Null-safe property access methods
- **Debug Tools**: Built-in debugging and status methods

### Route Protection
- **Flexible Configuration**: Easy to add/remove protected routes
- **Stored Redirects**: Remembers intended destination after authentication
- **Status Monitoring**: Real-time route protection status
- **Debug Information**: Comprehensive debugging tools

## Usage

### Frontend Authentication

#### Basic Authentication Check
```javascript
// Check if user is signed in
if (window.auth.isSignedIn()) {
    // User is authenticated
    const userInfo = window.auth.getUserInfo();
    console.log('User:', userInfo);
}
```

#### Using User Context
```javascript
// Listen for user state changes
window.userContext.addListener((user) => {
    if (user) {
        console.log('User signed in:', user);
    } else {
        console.log('User signed out');
    }
});

// Get current user
const currentUser = window.userContext.getUser();
const displayName = window.userContext.getDisplayName();
const email = window.userContext.getEmail();
```

#### Route Protection
Routes are automatically protected based on configuration in `RouteProtection` class:
- Protected routes: `/dashboard`, `/board/*`
- Public routes: `/`

```javascript
// Check if current route is protected
const isProtected = window.routeProtection.isProtectedRoute();

// Check if user can access current route
const canAccess = window.routeProtection.canAccessRoute();

// Get route protection status
const status = window.routeProtection.getRouteStatus();
```

#### Debugging Authentication
```javascript
// Debug authentication system
window.auth.debug();

// Debug user context
window.userContext.debug();

// Debug route protection
window.routeProtection.debug();
```

### Backend Authentication

#### Protected Endpoints
```go
// Apply authentication middleware
protected := api.Group("/")
protected.Use(middleware.AuthMiddleware())

// Get user ID from context
userID, err := middleware.GetUserID(c)
if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
    return
}
```

#### Token Validation
The middleware automatically validates JWT tokens and extracts user information:
- Validates token format and expiration
- Extracts user ID and session ID
- Stores claims in request context
- Provides helper functions for accessing user data

## Configuration

### Frontend Configuration
The authentication system uses global variables set in `base.html`:
```html
<script>
    window.clerkPublishableKey = '{{ .ClerkPublishableKey }}';
    window.clerkFrontendApiUrl = '{{ .ClerkFrontendApiUrl }}';
</script>
```

### Backend Configuration
Environment variables required:
```bash
CLERK_SECRET_KEY=your_clerk_secret_key
CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
```

## Error Handling

### Frontend Errors
- **Authentication Not Ready**: Handled with user-friendly messages
- **Token Validation**: Automatic token validation with expiration checking
- **Network Errors**: Graceful handling of network issues
- **User Feedback**: Clear error messages for users

### Backend Errors
- **Invalid Tokens**: Proper error responses for invalid JWT tokens
- **Missing Headers**: Clear error messages for missing authorization headers
- **Expired Tokens**: Automatic detection and rejection of expired tokens
- **Logging**: Comprehensive logging for debugging and monitoring

## Security Features

### JWT Token Security
- **Automatic Validation**: All tokens are validated on the backend
- **Expiration Checking**: Tokens are checked for expiration
- **Format Validation**: Token format is validated before processing
- **Secure Storage**: Tokens are handled securely in memory

### Route Protection
- **Automatic Redirects**: Unauthenticated users are redirected to sign-in
- **Stored Destinations**: Intended destinations are remembered after authentication
- **Session Management**: Proper session handling and cleanup
- **Access Control**: Fine-grained control over route access

## Performance Optimizations

### Frontend Optimizations
- **Configurable Timeouts**: Adjustable retry logic for better performance
- **Efficient Listeners**: Optimized event listener management
- **Lazy Loading**: Authentication components load only when needed
- **Memory Management**: Proper cleanup of listeners and resources

### Backend Optimizations
- **Efficient Token Validation**: Fast JWT token validation
- **Context Caching**: User context is cached for performance
- **Minimal Database Calls**: Authentication doesn't require database queries
- **Async Processing**: Non-blocking authentication checks

## Troubleshooting

### Common Issues

#### Authentication Not Working
1. Check browser console for error messages
2. Verify Clerk configuration in `base.html`
3. Ensure environment variables are set correctly
4. Check network connectivity to Clerk services

#### Route Protection Issues
1. Verify route configuration in `RouteProtection` class
2. Check authentication state with `window.auth.isSignedIn()`
3. Review browser console for error messages
4. Use debug methods to inspect current state

#### Token Issues
1. Check token format and expiration
2. Verify Clerk secret key configuration
3. Review backend logs for token validation errors
4. Ensure proper authorization headers are sent

### Debug Tools
```javascript
// Debug authentication system
window.auth.debug();

// Debug user context
window.userContext.debug();

// Debug route protection
window.routeProtection.debug();

// Check authentication status
console.log('Auth ready:', window.auth.isReady());
console.log('User signed in:', window.auth.isSignedIn());
console.log('User context loading:', window.userContext.isContextLoading());
```

## Migration Guide

### From Previous Version
The new authentication system is backward compatible but includes several improvements:

1. **Enhanced Error Handling**: Better error messages and user feedback
2. **Improved Logging**: More detailed logging for debugging
3. **Better Performance**: Optimized authentication flow
4. **Debug Tools**: Built-in debugging capabilities
5. **Cleaner Code**: More maintainable and readable code structure

### Breaking Changes
- None - the new system maintains the same public API
- All existing code should continue to work without changes
- New features are additive and don't break existing functionality

## Future Enhancements

### Planned Features
- **Multi-factor Authentication**: Support for MFA
- **Role-based Access Control**: Fine-grained permissions
- **Session Management**: Advanced session handling
- **Audit Logging**: Comprehensive audit trail
- **API Rate Limiting**: Protection against abuse

### Performance Improvements
- **Caching**: Token and user data caching
- **Lazy Loading**: On-demand authentication components
- **Optimized Validation**: Faster token validation
- **Reduced Network Calls**: Minimized API requests