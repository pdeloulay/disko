# Authentication System Documentation

## Overview

The Disko application uses Clerk for authentication, providing secure JWT-based authentication for admin users. The system includes both frontend and backend components for managing user authentication and authorization.

## Architecture

### Frontend Components

1. **Auth Class** (`static/js/auth.js`)
   - Manages Clerk integration
   - Handles sign-in/sign-up flows
   - Provides authentication state management
   - Offers token retrieval for API calls

2. **User Context** (`static/js/user-context.js`)
   - Centralized user state management
   - Reactive user information updates
   - Consistent user data across components

3. **Route Protection** (`static/js/route-protection.js`)
   - Automatic route-based authentication checks
   - Redirects unauthenticated users
   - Handles post-authentication routing

### Backend Components

1. **Authentication Middleware** (`middleware/auth.go`)
   - JWT token validation using Clerk
   - User context injection
   - Protected route enforcement

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
```

#### Route Protection
Routes are automatically protected based on configuration in `RouteProtection` class:
- Protected routes: `/dashboard`, `/board/*`
- Public routes: `/`

### Backend Authentication

#### Protected Endpoints
```go
// Apply authentication middleware
protected := api.Group("/")
protected.Use(middleware.AuthMiddleware())
{
    protected.GET("/user", func(c *gin.Context) {
        userID, err := middleware.GetUserID(c)
        if err != nil {
            // Handle error
            return
        }
        // Use userID for business logic
    })
}
```

#### Optional Authentication
```go
// Apply optional authentication middleware
optional := api.Group("/")
optional.Use(middleware.OptionalAuthMiddleware())
{
    optional.GET("/public-with-context", func(c *gin.Context) {
        if middleware.RequireAuth(c) {
            // User is authenticated
            userID, _ := middleware.GetUserID(c)
            // Provide authenticated experience
        } else {
            // User is not authenticated
            // Provide public experience
        }
    })
}
```

## API Endpoints

### Authentication Endpoints

#### GET /api/user
**Protected**: Yes  
**Description**: Get current user information  
**Response**:
```json
{
    "userID": "user_123",
    "sessionID": "sess_456"
}
```

#### GET /api/protected
**Protected**: Yes  
**Description**: Test protected endpoint  
**Response**:
```json
{
    "message": "This is a protected endpoint",
    "userID": "user_123"
}
```

## Environment Variables

Required environment variables for authentication:

```env
# Clerk Configuration
CLERK_SECRET_KEY=your_clerk_secret_key_here
CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key_here
```

## Error Handling

### Frontend Errors
- Network errors: Automatic retry with exponential backoff
- Authentication errors: Redirect to sign-in
- Token expiration: Automatic token refresh

### Backend Errors
- Missing authorization header: `401 UNAUTHORIZED`
- Invalid token format: `401 UNAUTHORIZED`
- Expired/invalid token: `401 UNAUTHORIZED`

Example error response:
```json
{
    "error": {
        "code": "UNAUTHORIZED",
        "message": "Authorization header is required"
    }
}
```

## Security Features

1. **JWT Validation**: All tokens are validated against Clerk's public keys
2. **Secure Headers**: Authorization headers use Bearer token format
3. **Context Isolation**: User context is properly isolated per request
4. **Route Protection**: Automatic protection of sensitive routes
5. **Token Refresh**: Automatic token refresh handling

## Testing

### Running Authentication Tests
```bash
# Run middleware tests
go test ./middleware -v

# Run all tests
go test ./... -v
```

### Test Coverage
- Authentication middleware validation
- Token parsing and validation
- User context extraction
- Route protection logic
- Error handling scenarios

## Integration with Clerk

### Setup Requirements
1. Create a Clerk application
2. Configure allowed origins and redirect URLs
3. Set up environment variables
4. Initialize Clerk in both frontend and backend

### Clerk Configuration
- **Frontend**: Clerk JavaScript SDK loaded via CDN
- **Backend**: Clerk Go SDK for JWT validation
- **Token Format**: Standard JWT with Clerk-specific claims

## Future Enhancements

1. **Role-based Access Control**: Extend middleware for role-based permissions
2. **Session Management**: Enhanced session handling and cleanup
3. **Multi-factor Authentication**: Integration with Clerk's MFA features
4. **Audit Logging**: Track authentication events and user actions