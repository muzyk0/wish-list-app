# Wish List Application

A full-stack wish list application allowing users to create and share gift lists with friends and family. The system includes public holiday pages showing gift lists with reservation functionality to avoid duplicates, along with personal accounts for managing wish lists.

## Architecture

The Wish List application is a full-stack application consisting of three main components:

- **Backend**: Go-based REST API using Echo framework with PostgreSQL database, utilizing sqlx for database operations (migrated from sqlc)
- **Frontend**: Next.js 16 application with React 19 and TypeScript
- **Mobile**: Expo/React Native application for iOS and Android

### Database Layer

The backend has migrated from sqlc (SQL code generation) to sqlx (direct database operations) for the following benefits:
- Simplified build process without code generation steps
- Increased flexibility in query handling
- Better control over database operations
- Reduced dependency complexity

## Project Structure

```
wish-list/
├── backend/          # Go backend with Echo framework
├── frontend/         # Next.js 16 frontend application
├── mobile/           # Expo/React Native mobile application
├── database/         # Database schema and migrations
├── api/              # OpenAPI specifications
├── specs/            # Feature specifications
├── contracts/        # API contracts
└── docs/             # Documentation files
```

## Setup

### Prerequisites
- Go 1.25+
- Node.js 18+
- Docker (for database)
- pnpm

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   cd wish-list
   pnpm install
   ```
3. Set up environment variables (copy from .env.example)
4. Run database migrations:
   ```bash
   make db-up
   make migrate-up
   ```
5. Start the applications:
   ```bash
   make backend   # Start backend server
   make frontend  # Start frontend server
   make mobile    # Start mobile development server
   ```

## Development

### Available Commands

- `make setup` - Set up the development environment
- `make backend` - Start the backend server
- `make frontend` - Start the frontend server
- `make mobile` - Start the mobile development server
- `make db-up` - Start the database with Docker
- `make migrate-up` - Run database migrations
- `make test` - Run tests for all components
- `make lint` - Run lint for all components
- `make format` - Format all components with Biome

### Mobile-First Architecture

The application follows a mobile-first approach where:
- User account creation and management happens in the mobile app
- Public wish list viewing and gift reservation happen in the frontend
- Account access redirects from web to mobile app at lk.domain.com

## Features

- User authentication and account management via mobile app
- Creation and sharing of public wish lists
- Gift reservation system to avoid duplicates
- Image upload support for gift items
- Multiple template options for wish list presentation
- Responsive design for both web and mobile viewing

## Technologies Used

### Backend
- Go 1.25 with Echo framework
- PostgreSQL database with sqlx for database operations
- JWT authentication system
- AWS S3 for image uploads
- Database migrations with golang-migrate

### Frontend
- Next.js 16 with App Router
- TypeScript with strict typing
- Tailwind CSS for styling
- TanStack Query for data fetching
- Shadcn/ui components

### Mobile
- Expo 54 with Expo Router
- React Native with TypeScript
- React Navigation for routing
- TanStack Query for data fetching
