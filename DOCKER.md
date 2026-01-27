# Docker Setup Guide

This guide explains how to run the Wish List application using Docker.

## Prerequisites

- Docker (version 20.10+)
- Docker Compose (version 1.29+)

## Quick Start

### 1. Start All Services

```bash
make docker-up
```

This will start:
- PostgreSQL database (port 5432)
- Redis cache (port 6379)
- Backend API (port 8080)

### 2. View Logs

```bash
# All services
make docker-logs

# Backend only
make docker-logs-backend
```

### 3. Stop Services

```bash
make docker-down
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make docker-up` | Start all services (database + backend) |
| `make docker-down` | Stop all services |
| `make docker-build` | Build the backend Docker image |
| `make docker-logs` | Show logs from all services |
| `make docker-logs-backend` | Show logs from backend only |
| `make docker-restart` | Restart all services |
| `make docker-restart-backend` | Restart backend only |
| `make docker-ps` | Show running containers |
| `make docker-clean` | Remove all containers, volumes, and images |
| `make db-up` | Start only database services (Postgres + Redis) |
| `make db-down` | Stop database services |

## Environment Variables

The backend service uses environment variables for configuration. You can:

1. **Use the default values** in docker-compose.yml (suitable for development)

2. **Set environment variables** before running docker-compose:
   ```bash
   export JWT_SECRET=your-secret-key
   export AWS_ACCESS_KEY_ID=your-access-key
   make docker-up
   ```

3. **Use a .env file** in the database directory:
   ```bash
   cd database
   cp ../.env.example .env
   # Edit .env with your values
   cd ..
   make docker-up
   ```

### Required Environment Variables for Production

- `JWT_SECRET` - Secret key for JWT token signing
- `SERVER_ENV` - Set to "production" for production environment
- `AWS_ACCESS_KEY_ID` - AWS access key (if using S3)
- `AWS_SECRET_ACCESS_KEY` - AWS secret key (if using S3)
- `AWS_S3_BUCKET_NAME` - S3 bucket name (if using S3)
- `ENCRYPTION_DATA_KEY` - Base64-encoded 32-byte key for PII encryption

## Service Details

### Backend Service

- **Port**: 8080
- **Health Check**: `http://localhost:8080/health`
- **API Documentation**: Available at the `/docs` endpoint

### PostgreSQL Database

- **Port**: 5432
- **Database**: wishlist_db
- **User**: user
- **Password**: password (change in production!)

### Redis Cache

- **Port**: 6379
- **Persistence**: AOF (Append Only File) enabled

## Running Migrations

Migrations are automatically included in the Docker image. To run them manually:

```bash
# Connect to the backend container
docker exec -it wishlist_backend sh

# Run migrations
./migrate -path internal/db/migrations -database "${DATABASE_URL}" up
```

Or use the Makefile command (requires Go installed locally):

```bash
make migrate-up
```

## Development Workflow

### 1. Database Only (for local development)

Start only the database services while running the backend locally:

```bash
# Start database
make db-up

# Run backend locally
make backend
```

### 2. Full Docker Stack

Run everything in Docker:

```bash
# Build and start all services
make docker-build
make docker-up

# View logs
make docker-logs
```

### 3. Rebuild After Code Changes

```bash
# Rebuild and restart backend
make docker-build
make docker-restart-backend
```

## Networking

All services are connected via the `wishlist_network` bridge network. Services can communicate using their service names:

- Backend → Postgres: `postgres:5432`
- Backend → Redis: `redis:6379`

## Troubleshooting

### Backend Container Won't Start

Check logs:
```bash
make docker-logs-backend
```

Common issues:
1. Database not ready: Wait for health check to pass
2. Missing environment variables: Check docker-compose.yml
3. Port already in use: Stop conflicting services

### Database Connection Failed

Verify database is running:
```bash
docker exec wishlist_postgres pg_isready -U user -d wishlist_db
```

### Clear All Data and Start Fresh

```bash
make docker-clean
make docker-up
```

## Production Deployment

For production deployment:

1. **Set environment variables** properly (never use defaults)
2. **Use secrets management** (e.g., Docker secrets, AWS Secrets Manager)
3. **Enable SSL/TLS** for database connections
4. **Use a reverse proxy** (e.g., nginx) in front of the backend
5. **Configure proper logging** and monitoring
6. **Set up backups** for PostgreSQL and Redis

Example production docker-compose override:

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  backend:
    environment:
      SERVER_ENV: production
      DATABASE_URL: ${DATABASE_URL}  # From secrets
      JWT_SECRET: ${JWT_SECRET}      # From secrets
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

Run with:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Health Checks

All services have health checks configured:

- **Backend**: HTTP check on `/health` endpoint
- **PostgreSQL**: `pg_isready` command
- **Redis**: `redis-cli ping` command

View health status:
```bash
make docker-ps
```

## Accessing Services

- **Backend API**: http://localhost:8080
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## FAQ - Frequently Asked Questions

### How do I restart Docker containers?

**Restart all services:**
```bash
make docker-restart
```

**Restart only backend:**
```bash
make docker-restart-backend
```

**Restart specific service:**
```bash
cd database && docker-compose restart postgres
cd database && docker-compose restart redis
```

### How do I apply code changes?

After making code changes to the backend:

```bash
# 1. Rebuild the Docker image
make docker-build

# 2. Restart the backend service
make docker-restart-backend

# 3. Check logs to verify
make docker-logs-backend
```

### How do I completely reset everything?

To start with a clean slate (removes all data):

```bash
make docker-clean  # Removes containers, volumes, and images
make docker-up     # Start fresh
```

### How do I apply environment variable changes?

Environment variables are set when containers start, so you need to restart:

```bash
# Stop all services
make docker-down

# Start with new environment variables
make docker-up
```

### How do I check if containers are running?

```bash
make docker-ps
```

Or use docker-compose directly:
```bash
cd database && docker-compose ps
```

### How do I view logs?

**All services:**
```bash
make docker-logs
```

**Backend only:**
```bash
make docker-logs-backend
```

**Follow logs in real-time:**
```bash
cd database && docker-compose logs -f backend
```

**Last 100 lines:**
```bash
cd database && docker-compose logs --tail=100 backend
```

### Backend won't start - what do I check?

1. **Check logs:**
   ```bash
   make docker-logs-backend
   ```

2. **Verify database is ready:**
   ```bash
   docker exec wishlist_postgres pg_isready -U user -d wishlist_db
   ```

3. **Check if port is in use:**
   ```bash
   lsof -i :8080
   ```

4. **Restart with clean slate:**
   ```bash
   make docker-down
   make docker-up
   ```

### How do I run migrations?

**Option 1: Using local Go (recommended):**
```bash
make migrate-up
```

**Option 2: Inside Docker container:**
```bash
docker exec -it wishlist_backend sh
./migrate -path internal/db/migrations -database "${DATABASE_URL}" up
```

### How do I connect to the database?

**Using psql:**
```bash
docker exec -it wishlist_postgres psql -U user -d wishlist_db
```

**Using a GUI tool (e.g., TablePlus, DBeaver):**
- Host: localhost
- Port: 5432
- Database: wishlist_db
- User: user
- Password: password

### How do I backup the database?

```bash
docker exec wishlist_postgres pg_dump -U user wishlist_db > backup.sql
```

**Restore from backup:**
```bash
docker exec -i wishlist_postgres psql -U user -d wishlist_db < backup.sql
```

### How do I access Redis CLI?

```bash
docker exec -it wishlist_redis redis-cli
```

**Test Redis connection:**
```bash
docker exec -it wishlist_redis redis-cli ping
# Should return: PONG
```

### Container keeps restarting - what's wrong?

1. **Check logs for errors:**
   ```bash
   make docker-logs-backend
   ```

2. **Check health status:**
   ```bash
   docker inspect --format='{{.State.Health.Status}}' wishlist_backend
   ```

3. **Common issues:**
   - Database not ready → Wait for health check to pass
   - Missing environment variables → Check docker-compose.yml
   - Port conflict → Change port in docker-compose.yml
   - Code compilation error → Check build logs

### How do I rebuild after changing dependencies?

```bash
# Backend (Go modules changed)
make docker-build
make docker-restart-backend

# If you need to rebuild everything from scratch
make docker-clean
make docker-up
```

### How do I run only the database for local development?

```bash
# Start only database services
make db-up

# Run backend locally
cd backend && go run cmd/server/main.go
```

### How do I check backend health?

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "checks": {
    "database": "ok"
  }
}
```

### How do I update to the latest Docker images?

```bash
cd database && docker-compose pull
make docker-up
```

### How do I see resource usage?

```bash
docker stats wishlist_backend wishlist_postgres wishlist_redis
```

### How do I stop containers without removing them?

```bash
cd database && docker-compose stop
```

**Resume later:**
```bash
cd database && docker-compose start
```

### What's the difference between `stop`, `down`, and `clean`?

- **`docker-compose stop`**: Stops containers (data preserved, can restart quickly)
- **`make docker-down`**: Stops and removes containers (volumes preserved)
- **`make docker-clean`**: Removes containers, volumes, AND images (complete reset)

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [Redis Docker Image](https://hub.docker.com/_/redis)
