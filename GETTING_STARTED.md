# Getting Started with Wish List Application

Welcome! This guide will help you set up and run the Wish List application on your local machine.

## 📋 Prerequisites

Before you begin, make sure you have:

- **Docker Desktop** installed and running ([Download here](https://www.docker.com/products/docker-desktop))
- **Go 1.25.5+** installed ([Download here](https://go.dev/dl/))
- **Node.js 18+** and **pnpm** installed (for frontend/mobile)
- A terminal/command line application

## 🚀 Quick Start (5 Minutes)

Get the application running in 3 simple steps:

### 1. Start the Database

```bash
make db-up
```

This starts PostgreSQL and Redis in Docker containers. You'll see:
```
✓ Container wishlist_postgres Started
✓ Container wishlist_redis Started
```

### 2. Run Database Migrations

```bash
make migrate-up
```

This creates all the tables your application needs. You'll see:
```
Running database migrations...
Migration completed successfully
```

### 3. Start the Backend Server

**Option A: Run Locally (Recommended for Development)**
```bash
make backend
```

**Option B: Run Everything in Docker**
```bash
make docker-up
```

### 4. Open in Your Browser

The backend API is now running at:
- **API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html

---

## 🔧 Detailed Setup

### First Time Setup

If this is your first time running the project:

```bash
# Install all dependencies
make setup

# Start database
make db-up

# Run migrations
make migrate-up

# Start backend
make backend
```

### Environment Configuration

The application uses different databases for different environments:

#### For Local Development (Docker Database)

Create a file `backend/.env.local`:

```env
DATABASE_URL=postgres://user:password@localhost:5432/wishlist_db?sslmode=disable
```

Run commands with the local database:
```bash
DATABASE_URL="postgres://user:password@localhost:5432/wishlist_db?sslmode=disable" make backend
```

#### For Production (Remote Database)

The default `backend/.env` file contains your production database URL.

**⚠️ Important**: Never run destructive commands (like `migrate-down`) with the production database URL!

---

## 📚 Common Tasks

### Starting the Application

```bash
# Start database only
make db-up

# Start backend (with local database)
DATABASE_URL="postgres://user:password@localhost:5432/wishlist_db?sslmode=disable" make backend

# Start frontend (in a new terminal)
make frontend

# Start mobile (in a new terminal)
make mobile
```

### Stopping Services

```bash
# Stop database containers
make db-down

# Stop all Docker services
make docker-down

# Stop backend - Press Ctrl+C in the terminal where it's running
```

### Database Migrations

```bash
# Check current migration version
cd backend && go run cmd/migrate/main.go -action version

# Apply migrations (move forward)
make migrate-up

# Rollback migrations (move backward)
make migrate-down

# Create a new migration
make migrate-create
# (Enter migration name when prompted)
```

### Viewing Logs

```bash
# View all Docker logs
make docker-logs

# View backend logs only
make docker-logs-backend

# View database logs
docker logs wishlist_postgres
```

### Running Tests

```bash
# Run all tests
make test

# Run backend tests only
make test-backend

# Run backend tests with coverage
make test-backend-advanced
```

---

## 🗄️ Database Management

### Checking Database Contents

Connect to your local database:

```bash
docker exec -it wishlist_postgres psql -U user -d wishlist_db
```

Useful PostgreSQL commands:
```sql
\dt              -- List all tables
\d users         -- Describe users table
\q               -- Quit psql

SELECT * FROM users;          -- View all users
SELECT * FROM wishlists;      -- View all wishlists
```

### Resetting the Database

If you need a fresh start:

```bash
# Stop database
make db-down

# Remove all data
docker volume rm database_postgres_data database_redis_data

# Start fresh
make db-up
make migrate-up
```

---

## 🐛 Troubleshooting

### "Docker daemon not running"

**Problem**: Error connecting to Docker

**Solution**:
1. Open Docker Desktop application
2. Wait for it to fully start
3. Try your command again

### "Migration failed" or "Database connection error"

**Problem**: Can't connect to database

**Solution**:
```bash
# Check if database is running
make docker-ps

# If not running, start it
make db-up

# Wait 5 seconds for database to fully start, then retry
make migrate-up
```

### "Port 5432 already in use"

**Problem**: Another PostgreSQL instance is using the port

**Solution**:
```bash
# Option 1: Stop other PostgreSQL
# (On macOS with Homebrew)
brew services stop postgresql

# Option 2: Change port in database/docker-compose.yml
# Change "5432:5432" to "5433:5432"
```

### "Connection refused" when running backend

**Problem**: Database not accessible

**Solution**:
```bash
# Check database is running
docker ps | grep postgres

# Restart database if needed
make db-down
make db-up

# Wait 10 seconds, then try backend again
```

### Tables not found

**Problem**: Application can't find database tables

**Solution**:
```bash
# Run migrations
make migrate-up

# Verify tables exist
docker exec wishlist_postgres psql -U user -d wishlist_db -c "\dt"

# Should show: users, wishlists, gift_items, wishlist_items, reservations
```

---

## 🎯 Development Workflows

### Starting Your Day

```bash
# 1. Start database
make db-up

# 2. Start backend (in one terminal)
DATABASE_URL="postgres://user:password@localhost:5432/wishlist_db?sslmode=disable" make backend

# 3. Start frontend (in another terminal)
make frontend
```

### After Pulling New Code

```bash
# Update dependencies
make setup

# Run new migrations
make migrate-up

# Restart backend
make backend
```

### Before Committing Code

```bash
# Format code
make format

# Run linter
make lint

# Run tests
make test
```

---

## 📖 Additional Resources

### Useful Commands Reference

```bash
make help              # Show all available commands
make build             # Build all components
make clean             # Clean build artifacts
make docs              # Generate API documentation
make format            # Format all code
make lint              # Run linters
```

### Project Structure

```
wish-list-app/
├── backend/           # Go API server
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   └── .env          # Environment variables
├── frontend/          # Next.js web application
├── mobile/            # Expo mobile application
├── database/          # Docker compose for databases
└── Makefile          # Development commands
```

### Getting Help

- **API Documentation**: http://localhost:8080/swagger/index.html (when backend is running)
- **Project Documentation**: See `/docs` folder
- **Migration Specs**: See `/specs/004-db-init-migration`

---

## ✅ Checklist for New Developers

- [ ] Docker Desktop installed and running
- [ ] Go 1.25.5+ installed
- [ ] Node.js and pnpm installed
- [ ] Repository cloned
- [ ] `make setup` completed successfully
- [ ] `make db-up` starts database
- [ ] `make migrate-up` creates tables
- [ ] `make backend` starts server
- [ ] Can access http://localhost:8080/swagger/index.html
- [ ] `make test` passes

**Welcome to the team! 🎉**

If you encounter any issues not covered here, please ask for help!
