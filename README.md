# Disko

A web-based project management tool that allows solopreneurs to share their work progress with customers through public boards.

## Project Structure

```
disko/
├── backend/           # Go backend with Gin framework and HTML templates
│   ├── templates/     # HTML templates
│   ├── static/        # CSS, JavaScript, and static assets
│   ├── models/        # Data models
│   ├── handlers/      # API handlers
│   ├── middleware/    # Custom middleware
│   └── utils/         # Utility functions
└── .kiro/specs/       # Feature specifications and design documents
```

## Tech Stack

### Frontend
- Single Page Application (SPA) using Go HTML templates
- Vanilla JavaScript for interactivity
- Clerk JavaScript SDK for authentication
- CSS for styling

### Backend
- Go with Gin framework
- MongoDB with Go driver 2.0
- HTML template rendering
- Static file serving
- JWT authentication via Clerk
- RESTful API design

## Getting Started

### Prerequisites
- Go (v1.21 or higher)
- MongoDB Atlas account or local MongoDB instance

### Setup
```bash
cd backend
go mod download
cp .env.example .env
# Edit .env with your MongoDB URI and Clerk keys
go run main.go
```

Visit http://localhost:8080 to access the application.

## Environment Variables

### Backend (.env)
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name (default: disko_board)
- `PORT`: Server port (default: 8080)
- `CLERK_SECRET_KEY`: Your Clerk secret key
- `CLERK_PUBLISHABLE_KEY`: Your Clerk publishable key
- `ENV`: Environment (development/production)

## Development

The project is set up with:
- HTML template rendering with Go
- Static file serving for CSS/JS
- Clerk authentication integration
- MongoDB connection utilities
- Basic project structure for scalable development

## Application Routes

- `/` - Landing page
- `/dashboard` - Admin dashboard (protected)
- `/board/:publicLink` - Public board view
- `/api/*` - API endpoints
- `/health` - Health check endpoint

## Changelog

### [v0.2.0] - Simplified Authentication System
- **🔐 Direct Clerk Integration**: Removed complex auth.js file and implemented direct Clerk integration like numi project
- **🎯 Simplified Flow**: Clean authentication flow with direct Clerk initialization in HTML templates
- **🛡️ Server-side Protection**: Authentication handled primarily through server-side middleware
- **📝 Cleaner Code**: Removed unnecessary complexity and auth-related JavaScript files
- **🔧 Better Performance**: Faster page loads with less JavaScript overhead
- **⚡ Direct Integration**: Clerk initialized directly in templates following numi pattern
- **🎨 Consistent UI**: Unified authentication experience across all pages
- **📱 Mobile Friendly**: Simplified auth flow works better on mobile devices

### [v0.1.5] - Board ID Template Debugging
- **🔍 Template Debugging**: Added console logging to debug board ID template variable rendering
- **🐛 Issue Investigation**: Investigating why board ID is showing as "undefined" in API calls
- **📝 Debug Logging**: Added template variable debugging to identify template rendering issues
- **🔧 Variable Tracking**: Tracking board ID, public link, and ownership flags from server-side template

### [v0.1.4] - Board Page Header & Layout Consistency
- **🎨 Header Consistency**: Updated board page header to match dashboard header structure
- **📱 Layout Improvements**: Moved board info and actions to main content area for better organization
- **🎯 Visual Hierarchy**: Cleaner header with logo and user menu, board details in content section
- **📝 Responsive Design**: Added responsive styling for board info section on mobile devices
- **🔧 WebSocket Status**: Enhanced WebSocket status indicator styling with proper color coding
- **⚡ Performance**: Simplified header structure for faster rendering

### [v0.1.3] - Frontend Authentication & Modal Styling
- **🔓 Route Access**: Removed AuthMiddleware from `/board/:id` route to allow frontend authentication handling
- **🎨 Modal Styling**: Enhanced Create New Board modal with professional styling and proper form layout
- **📱 Responsive Design**: Improved modal responsiveness and form element spacing
- **🎯 Frontend Auth**: Board pages now handle authentication through JavaScript instead of route-level protection
- **🔧 Form Actions**: Better styling for modal action buttons with proper background and spacing
- **📝 CSS Variables**: Added missing success colors and improved design system consistency

### [v0.1.2] - Authentication Loop Fixes & Token Management
- **🔐 Token Validation**: Enhanced JWT token validation with length and expiration checks
- **🔄 Retry Logic**: Implemented retry limits to prevent infinite authentication loops
- **🧹 Token Cleanup**: Automatic clearing of invalid/expired tokens from localStorage and sessionStorage
- **🎯 Error Handling**: Better error messages and user feedback for authentication issues
- **🔧 Debug Logging**: Comprehensive logging for token validation and authentication flow
- **⚡ Performance**: Improved authentication performance with proper token caching

### [v0.1.1] - Code Organization & Handler Cleanup
- **🧹 Code Cleanup**: Removed all inline handlers from main.go and moved them to proper handler files
- **📁 Better Organization**: Created dedicated handler files for different concerns (user.go, health.go)
- **🔧 Maintainability**: Improved code structure with proper separation of concerns
- **📝 Documentation**: Added comprehensive handler documentation and logging
- **🎯 Consistency**: Standardized handler patterns across all endpoints
- **⚡ Performance**: Reduced main.go complexity and improved readability
- **🔧 API Endpoint Fix**: Fixed JavaScript to use correct private endpoints for board data loading

### [v0.1.0] - Authentication & UI Improvements
- **🔐 Authentication System**: Complete Clerk integration with JWT token management
- **🛡️ Route Protection**: Separate private (`/board/:id`) and public (`/public/:publicLink`) board routes
- **🔑 Bearer Token Auth**: Authenticated API requests with proper Authorization headers
- **📊 Public Stats**: Server-side templated statistics always visible on landing page
- **🎨 UI Consistency**: Unified auth setup across dashboard and board pages
- **🔄 Event Listeners**: Fixed Clerk event listener compatibility for different API versions
- **📱 User Menu**: Added user menu with sign-out functionality to all pages
- **🏷️ Version Display**: App version (v0.1.0) shown in footer and headers
- **🎯 Error Handling**: Improved error messages and user feedback
- **🔧 Debug Logging**: Enhanced logging for authentication and API calls
- **🎨 Landing Page Redesign**: Complete redesign focused on solopreneurs with minimal, pixel-perfect design and "less is more" approach

### [v0.0.9] - Board Creation Enhancements & Modal Redesign
- **Enhanced**: Boards are now private by default (`isPublic: false`) for better security
- **Added**: Default welcome idea automatically created with each new board
- **Improved**: Modern modal design with gradient headers, smooth animations, and enhanced UX
- **Updated**: Form styling with better typography, spacing, and visual feedback
- **Enhanced**: Button design with hover effects and improved accessibility
- **Security**: Public board endpoints now require `isPublic: true` - boards must be explicitly made public
- **Fixed**: Board card hamburger menu now displays as proper dropdown overlay instead of inline buttons
- **Fixed**: Board owners can now access their own boards via direct URL without requiring public access
- **Fixed**: Board access workflow now properly handles non-existent or private boards with clear error pages
- **Fixed**: Board route now uses OptionalAuthMiddleware to properly handle authenticated user access
- **Enhanced**: Restructured board routes for better security and clarity:
  - `GET /board/:id` - Private route with JWT enforcement (for board owners)
  - `GET /board/public/:publicLink` - Public route with rate limiting (for public access)
- **Security**: Enhanced public link generation to use full UUID (37 characters) for maximum security
- **API**: Added missing GET `/boards/:id` endpoint for authenticated board access
- **Dashboard**: Enhanced `/api/boards` endpoint to include ideas count for each board
- **Navigation**: Fixed board view loading by using proper page navigation instead of HTML replacement
- **Authentication**: Updated board route to use AuthMiddleware for authenticated users only
- **Routes**: Separated private board access (`/board/:id`) from public board access (`/public/:publicLink`)

### [Previous] - Short UUID Implementation
- **Enhanced**: Public board links now use short Google UUIDs (12 characters) with "p" prefix
- **Improved**: Better user experience with shorter, more manageable public links
- **Added**: Utility functions in `utils/uuid.go` for consistent UUID generation
- **Updated**: Board IDs use "b" prefix, Idea IDs use "i" prefix for easy identification
- **Standardized**: All UUIDs now have consistent prefixes for better organization

### Previous Updates
- **Added**: Clerk authentication integration
- **Added**: MongoDB database setup and models
- **Added**: Board and idea management APIs
- **Added**: Public board sharing functionality
- **Added**: Real-time feedback system with thumbs up and emoji reactions

## Next Steps

This is the foundation setup. The next tasks will implement:
1. Enhanced UI components and interactions
2. Advanced board customization features
3. Analytics and reporting capabilities
4. Team collaboration features
5. Real-time features and more...