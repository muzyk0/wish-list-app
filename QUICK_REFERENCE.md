# Quick Reference Card

## ⚡ Most Common Commands

### Daily Development

```bash
# Start your day
make db-up                                    # Start database
make backend                                  # Start backend API

# Or use local database override
DATABASE_URL="postgres://user:password@localhost:5432/wishlist_db?sslmode=disable" make backend

# Frontend & Mobile (in separate terminals)
make frontend                                 # Start Next.js web app
make mobile                                   # Start Expo mobile app
```

### Database Operations

```bash
make migrate-up                               # Apply migrations
make migrate-down                             # Rollback migrations
make migrate-create                           # Create new migration

# Check database
docker exec -it wishlist_postgres psql -U user -d wishlist_db
```

### Docker Services

```bash
make docker-up                                # Start all services
make docker-down                              # Stop all services
make docker-logs                              # View all logs
make docker-logs-backend                      # Backend logs only
make docker-ps                                # Show running containers
```

### Testing & Quality

```bash
make test                                     # Run all tests
make test-backend                             # Backend tests only
make lint                                     # Run linters
make format                                   # Format code
```

### Documentation

```bash
make docs                                     # Generate API docs
# Then visit: http://localhost:8080/swagger/index.html
```

---

## 🔗 Important URLs

| Service | URL |
|---------|-----|
| Backend API | http://localhost:8080 |
| Swagger Docs | http://localhost:8080/swagger/index.html |
| Frontend | http://localhost:3000 |
| Mobile | http://localhost:8081 |

---

## 🗄️ Database Credentials

### Local Docker Database
```
Host: localhost
Port: 5432
User: user
Password: password
Database: wishlist_db
```

**Connection String:**
```
postgres://user:password@localhost:5432/wishlist_db?sslmode=disable
```

---

## 🐛 Quick Fixes

### Problem: Docker not running
```bash
# Open Docker Desktop app and wait for it to start
```

### Problem: Port already in use
```bash
make docker-down                              # Stop all services
docker ps                                     # Check running containers
```

### Problem: Database connection failed
```bash
make db-down                                  # Stop database
make db-up                                    # Start database
sleep 5                                       # Wait for startup
make migrate-up                               # Retry migration
```

### Problem: Tables not found
```bash
make migrate-up                               # Run migrations
docker exec wishlist_postgres psql -U user -d wishlist_db -c "\dt"
```

### Problem: Fresh start needed
```bash
make docker-down                              # Stop everything
docker volume rm database_postgres_data       # Clear data
make db-up                                    # Start fresh
make migrate-up                               # Create tables
```

---

## 📝 Useful PostgreSQL Commands

```sql
\dt                                           -- List tables
\d table_name                                 -- Describe table
\du                                           -- List users
\l                                            -- List databases
\q                                            -- Quit

-- Quick queries
SELECT * FROM users LIMIT 10;
SELECT * FROM wishlists WHERE owner_id = 'uuid';
SELECT COUNT(*) FROM gift_items;
```

---

## 🎯 Environment Variables

```bash
# Use local database
export DATABASE_URL="postgres://user:password@localhost:5432/wishlist_db?sslmode=disable"

# Or prefix commands
DATABASE_URL="..." make backend
DATABASE_URL="..." make migrate-up
```

---

## 📦 Migration Files

Location: `backend/internal/app/database/migrations/`

Current migrations:
- `000001_init_schema.up.sql` - Creates all tables
- `000001_init_schema.down.sql` - Drops all tables

---

## ✅ Health Checks

```bash
# Is database running?
docker ps | grep postgres

# Is backend running?
curl http://localhost:8080/health

# Check migration version
cd backend && go run cmd/migrate/main.go -action version

# View database tables
docker exec wishlist_postgres psql -U user -d wishlist_db -c "\dt"
```

---

**💡 Tip**: Keep this file open in a terminal for quick copy-paste!

For detailed instructions, see [GETTING_STARTED.md](./GETTING_STARTED.md)
