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

## Next Steps

This is the foundation setup. The next tasks will implement:
1. Database models and MongoDB setup
2. Authentication system with Clerk
3. Board and idea management APIs
4. Enhanced UI components and interactions
5. Real-time features and more...