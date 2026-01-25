# Quickstart Guide: Wish List Application

## Overview
This guide provides instructions for setting up and running the Wish List Application for development and production environments.

## Prerequisites

### Local Development
- Go 1.25 or higher
- Node.js 18+ and npm/yarn
- PostgreSQL 12+
- Docker and Docker Compose
- AWS S3 account (or local S3-compatible service like MinIO)
- Git

### Environment Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/your-org/wish-list.git
   cd wish-list
   ```

2. Install backend dependencies:
   ```bash
   cd backend
   go mod download
   ```

3. Install frontend dependencies:
   ```bash
   cd ../frontend
   npm install
   # or
   yarn install
   ```

4. Install mobile dependencies:
   ```bash
   cd ../mobile
   npm install
   # or
   yarn install
   ```

## Environment Configuration

### Backend (.env)
Create a `.env` file in the `backend` directory:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/wishlist_db
DATABASE_MAX_CONNECTIONS=20

# Server
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_ENV=development

# JWT
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRY_HOURS=24

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET_NAME=your-bucket-name
AWS_S3_REGION=us-east-1

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006

# Email (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SENDER_EMAIL=your-email@gmail.com
```

### Frontend (.env.local)
Create a `.env.local` file in the `frontend` directory:

```env
# API
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws

# Authentication
NEXT_PUBLIC_JWT_SECRET=your-super-secret-jwt-key-here

# AWS S3
NEXT_PUBLIC_S3_BUCKET_URL=https://your-bucket-name.s3.amazonaws.com
```

### Mobile (.env)
Create a `.env` file in the `mobile` directory:

```env
# API
EXPO_PUBLIC_API_URL=http://10.0.2.2:8080/api  # Use 10.0.2.2 for Android emulator
EXPO_PUBLIC_WS_URL=ws://10.0.2.2:8080/ws

# AWS S3
EXPO_PUBLIC_S3_BUCKET_URL=https://your-bucket-name.s3.amazonaws.com
```

## Database Setup

### Using Docker (Recommended for Development)
1. Navigate to the database directory:
   ```bash
   cd database
   ```

2. Start PostgreSQL with Docker Compose:
   ```bash
   docker-compose up -d
   ```

3. Run database migrations:
   ```bash
   cd ../backend
   go run cmd/migrate/main.go
   ```

### Manual Database Setup
1. Create the database:
   ```sql
   CREATE DATABASE wishlist_db;
   ```

2. Run the schema migrations:
   ```bash
   cd backend
   go run cmd/migrate/main.go
   ```

## Running the Application

### Development Mode

#### Backend
1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Run the server:
   ```bash
   go run cmd/server/main.go
   ```
   The API will be available at `http://localhost:8080`

#### Frontend
1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Start the development server:
   ```bash
   npm run dev
   # or
   yarn dev
   ```
   The web app will be available at `http://localhost:3000`

#### Mobile
1. Navigate to the mobile directory:
   ```bash
   cd mobile
   ```

2. Start the Expo development server:
   ```bash
   npx expo start
   ```
   Follow the instructions in the terminal to run on iOS simulator, Android emulator, or physical device.

### Production Mode

#### Backend
1. Build the binary:
   ```bash
   cd backend
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go
   ```

2. Create a Docker image:
   ```bash
   docker build -t wishlist-backend .
   ```

3. Run the container:
   ```bash
   docker run -p 8080:8080 -e DATABASE_URL=... wishlist-backend
   ```

#### Frontend
1. Build the application:
   ```bash
   cd frontend
   npm run build
   ```

2. Serve the application (e.g., with Vercel, Netlify, or nginx):
   ```bash
   # Using serve for local testing
   npx serve -s out
   ```

## API Contracts

The API contracts are defined in the `contracts/` directory:
- `user-api.json` - User authentication and management endpoints
- `wishlist-api.json` - Wish list creation and management endpoints
- `gift-item-api.json` - Gift item operations endpoints
- `reservation-api.json` - Gift reservation endpoints

## Database Schema

The database schema is documented in `docs/database-schema.md` and generated from the SQL migration files in `database/schema.sql`.

## Testing

### Backend Tests
```bash
cd backend
go test ./... -v
```

### Frontend Tests
```bash
cd frontend
npm test
# or
yarn test
```

### Mobile Tests
```bash
cd mobile
npm test
# or
yarn test
```

## Deployment

### Frontend (Vercel)
1. Install Vercel CLI:
   ```bash
   npm install -g vercel
   ```

2. Deploy:
   ```bash
   cd frontend
   vercel --prod
   ```

### Backend (Docker)
1. Build and push Docker image to your registry
2. Deploy to your preferred platform (Fly.io, Railway, Render, etc.)

### Mobile (App Stores)
1. Configure app credentials and certificates
2. Build for each platform:
   ```bash
   cd mobile
   npx expo build:ios
   npx expo build:android
   ```

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Verify PostgreSQL is running
   - Check `DATABASE_URL` in your environment variables
   - Ensure the database exists and user has proper permissions

2. **CORS Errors**
   - Verify `CORS_ALLOWED_ORIGINS` includes your frontend URL
   - Check that the backend is running on the expected port

3. **S3 Upload Issues**
   - Verify AWS credentials are correct
   - Check that the S3 bucket exists and has proper permissions
   - Ensure the region matches your configuration

4. **Authentication Issues**
   - Verify JWT secret is the same across all services
   - Check that the frontend is sending authentication headers correctly

### Development Tips

1. **Hot Reload**
   - Backend: Use `air` for Go hot reloading
   - Frontend: Next.js has built-in hot reload
   - Mobile: Expo has built-in hot reload

2. **API Testing**
   - Use the OpenAPI specs in `contracts/` to generate client code
   - Import the specs into Postman or Insomnia for manual testing

3. **Database Migrations**
   - Always backup your database before running migrations
   - Test migrations on a copy of production data when possible