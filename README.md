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

### [Latest] - Board Creation Enhancements & Modal Redesign
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
- **Authentication**: Fixed board navigation to use authenticated fetch requests with bearer token
- **Auth Consistency**: Updated board.html to have same auth setup as dashboard.html
- **Public Stats**: Removed authentication requirement for landing page stats - now server-side templated and always visible
- **Auth Fix**: Fixed Clerk authentication methods to use proper API calls instead of incorrect function calls
- **Listener Fix**: Fixed Clerk event listener methods to handle both `addListener/removeListener` and `on/off` patterns
- **Version Display**: Added app version display to footer (index) and headers (dashboard/board) from static/.version file

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