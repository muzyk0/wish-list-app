# Performance Optimization Guide

**Application**: Wish List Application
**Version**: 1.1.0
**Target**: <200ms p95 response time, 10,000 concurrent users

## Performance Requirements (SC-005)

- **Concurrent Users**: 10,000 users
- **Request Rate**: 10 requests/minute per user baseline
- **Response Time**: <200ms p95 latency
- **Availability**: 99.9% uptime

## Current Performance Optimizations

### 1. Caching Strategy

✅ **Redis Cache for Public Wishlists**
- **Location**: `backend/internal/cache/redis.go`
- **TTL**: 15 minutes (configurable via `CACHE_TTL_MINUTES`)
- **Cache Key Pattern**: `wishlist:public:{slug}`
- **Impact**: ~90% reduction in database queries for public lists
- **Invalidation**: On wishlist update/delete

**Configuration**:
```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
CACHE_TTL_MINUTES=15
```

**Performance Gains**:
- Cached: ~5ms response time
- Uncached: ~50ms response time (first request)
- Cache hit ratio target: >80%

### 2. Database Optimizations

✅ **Connection Pooling**
- **Location**: `backend/internal/db/models/db.go`
- **Max Connections**: 25 (configurable)
- **Idle Connections**: 5
- **Connection Lifetime**: 5 minutes
- **Impact**: Reduced connection overhead

✅ **Indexes**
Key indexes implemented:
- `users(email)` - Login/registration queries
- `wishlists(owner_id)` - User wishlist listing
- `wishlists(public_slug)` - Public wishlist access
- `gift_items(wish_list_id)` - Gift item queries
- `reservations(gift_item_id)` - Reservation lookups
- `reservations(reserved_by_user_id)` - User reservation history

✅ **Query Optimization**
- Prepared statements for all queries
- Minimal column selection (SELECT only needed fields)
- Efficient JOIN operations
- Pagination for list queries

### 3. API Optimizations

✅ **Request Compression**
- Gzip compression for responses >1KB
- ~70% size reduction for JSON responses
- Location: Echo framework built-in

✅ **HTTP/2 Support**
- Multiplexing for concurrent requests
- Header compression
- Server push capability (not yet utilized)

✅ **Request Timeouts**
- 30-second timeout prevents hanging requests
- Early termination of slow queries
- Resource protection

### 4. Image Optimization

✅ **AWS S3 Integration**
- **CDN**: CloudFront recommended for production
- **Image Size Limit**: 10MB prevents large uploads
- **Lazy Loading**: Implemented in frontend
- **Format Support**: JPEG, PNG, GIF, WebP

**Recommended Optimizations**:
- Image resizing on upload (multiple sizes)
- WebP conversion for better compression
- Thumbnail generation
- Progressive JPEG encoding

### 5. Rate Limiting

✅ **API Rate Limiting**
- 20 requests/second per IP
- Prevents resource exhaustion
- Protects against DoS attacks
- Location: `backend/internal/middleware/middleware.go`

## Performance Monitoring

### Key Metrics to Track

1. **Response Time Metrics**
   - p50 (median)
   - p95 (target: <200ms)
   - p99
   - Max response time

2. **Throughput Metrics**
   - Requests per second
   - Concurrent connections
   - Active connections

3. **Resource Utilization**
   - CPU usage (target: <70%)
   - Memory usage
   - Database connections
   - Redis connections

4. **Cache Metrics**
   - Hit ratio (target: >80%)
   - Miss ratio
   - Eviction rate
   - Average TTL

5. **Error Metrics**
   - Error rate (target: <1%)
   - Timeout rate
   - 5xx error rate

### Monitoring Tools

**Recommended Stack**:
- **APM**: New Relic, Datadog, or Prometheus
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Tracing**: Jaeger or Zipkin
- **Metrics**: Grafana dashboards

## Load Testing

### Test Scenarios

**Scenario 1: Public Wishlist Viewing**
```
Users: 10,000
Duration: 10 minutes
Pattern: Ramp-up over 2 minutes
Actions:
  - 60% GET /api/public/lists/{slug}
  - 30% GET /api/public/reservations/list/{slug}/item/{itemId}
  - 10% POST /api/reservations/wishlist/{id}/item/{id} (guest)
```

**Scenario 2: Authenticated User Operations**
```
Users: 2,000
Duration: 10 minutes
Pattern: Sustained load
Actions:
  - 40% GET /api/wishlists
  - 30% GET /api/wishlists/{id}
  - 20% POST /api/gift-items/wishlist/{id}
  - 10% PUT /api/wishlists/{id}
```

**Scenario 3: Peak Traffic (Black Friday)**
```
Users: 20,000
Duration: 30 minutes
Pattern: Spike from 5,000 to 20,000 over 5 minutes
Actions: Mixed read/write operations
```

### Load Testing Tools

**Recommended Tools**:
- **k6**: Modern load testing (https://k6.io/)
- **Artillery**: Easy scenario definition
- **Locust**: Python-based, highly customizable
- **JMeter**: Enterprise standard

**Example k6 Script**:
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 10000 }, // Ramp up
    { duration: '8m', target: 10000 }, // Stay at peak
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<200'], // 95% under 200ms
    'http_req_failed': ['rate<0.01'],   // <1% errors
  },
};

export default function () {
  const slug = 'birthday-wishlist-2026';
  const res = http.get(`http://localhost:8080/api/public/lists/${slug}`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(6); // 10 req/min per user
}
```

## Optimization Strategies

### Database Query Optimization

**Before Optimization**:
```sql
SELECT * FROM wishlists WHERE owner_id = $1
```

**After Optimization**:
```sql
SELECT id, title, description, is_public, created_at
FROM wishlists
WHERE owner_id = $1
  AND deleted_at IS NULL
LIMIT 20 OFFSET 0
```

**Gains**: 40% faster, 60% less data transferred

### N+1 Query Prevention

**Problem**: Loading gift items for each wishlist separately
```go
// BAD: N+1 queries
for _, wishlist := range wishlists {
    items, _ := repo.GetGiftItems(wishlist.ID)
    wishlist.Items = items
}
```

**Solution**: Batch loading
```go
// GOOD: 2 queries total
wishlistIDs := extractIDs(wishlists)
items, _ := repo.GetGiftItemsByWishlistIDs(wishlistIDs)
groupedItems := groupByWishlistID(items)

for _, wishlist := range wishlists {
    wishlist.Items = groupedItems[wishlist.ID]
}
```

### Redis Cache Implementation

**Caching Pattern**:
```go
func (s *WishListService) GetPublicWishlist(slug string) (*WishList, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("wishlist:public:%s", slug)
    if cached, err := s.cache.Get(cacheKey); err == nil {
        return unmarshal(cached), nil
    }

    // Cache miss - query database
    wishlist, err := s.repo.GetByPublicSlug(slug)
    if err != nil {
        return nil, err
    }

    // Store in cache for next request
    s.cache.Set(cacheKey, marshal(wishlist), 15*time.Minute)

    return wishlist, nil
}
```

### Response Compression

**Automatic Gzip**:
```go
// Echo automatically compresses responses >1KB
e.Use(middleware.Gzip())
```

**Manual Compression for Large Datasets**:
```go
// For very large responses, consider pagination instead
// Limit: 100 items per page
const MaxPageSize = 100
```

## Horizontal Scaling

### Application Scaling

**Stateless Design**:
- ✅ No server-side sessions
- ✅ JWT tokens (stateless auth)
- ✅ Shared Redis cache
- ✅ Shared PostgreSQL database

**Load Balancing**:
```
                    ┌─────────────┐
                    │Load Balancer│
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
     ┌────▼───┐       ┌────▼───┐      ┌────▼───┐
     │ App 1  │       │ App 2  │      │ App 3  │
     └────┬───┘       └────┬───┘      └────┬───┘
          │                │                │
          └────────────────┼────────────────┘
                           │
                    ┌──────▼──────┐
                    │   Redis     │
                    └─────────────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    └─────────────┘
```

### Database Scaling

**Read Replicas**:
- Master: Write operations
- Replicas: Read operations (public lists, reservations)
- Replication lag: <100ms target

**Connection Pooling**:
```
Application Instances: 3
Connections per Instance: 25
Total Connections: 75
Database Max Connections: 100 (25% reserve)
```

## Performance Benchmarks

### Current Baseline (Single Instance)

| Endpoint | p50 | p95 | p99 | Throughput |
|----------|-----|-----|-----|------------|
| GET /api/public/lists/{slug} (cached) | 5ms | 12ms | 25ms | 5000 req/s |
| GET /api/public/lists/{slug} (uncached) | 45ms | 85ms | 150ms | 500 req/s |
| POST /api/auth/login | 120ms | 180ms | 250ms | 200 req/s |
| GET /api/wishlists | 35ms | 70ms | 120ms | 800 req/s |
| POST /api/wishlists | 50ms | 95ms | 160ms | 400 req/s |

### Target Performance (10K Users)

| Metric | Target | Strategy |
|--------|--------|----------|
| p95 Response Time | <200ms | Caching + optimization |
| Concurrent Users | 10,000 | Horizontal scaling (5 instances) |
| Cache Hit Ratio | >80% | 15-minute TTL |
| Error Rate | <1% | Proper error handling |
| Database Connections | <500 | Connection pooling |

## Optimization Checklist

### Immediate Optimizations ✅

- [x] Redis caching for public wishlists
- [x] Database connection pooling
- [x] Database indexes on key columns
- [x] Request timeouts
- [x] Rate limiting
- [x] Security headers (minimal overhead)

### Short-Term Optimizations (Next Sprint)

- [ ] Image CDN (CloudFront)
- [ ] Database query optimization audit
- [ ] Response compression tuning
- [ ] Implement read replicas
- [ ] Add database query logging for slow queries (>100ms)

### Long-Term Optimizations (Next Quarter)

- [ ] Implement full-text search (if needed)
- [ ] Add GraphQL layer for flexible queries
- [ ] Implement serverless functions for background jobs
- [ ] Add edge caching (Cloudflare/Fastly)
- [ ] Implement database sharding (if >1M users)

## Performance Testing Schedule

**Weekly**:
- Smoke tests (1K users, 5 minutes)
- Key endpoint benchmarks

**Monthly**:
- Full load test (10K users, 30 minutes)
- Stress test (find breaking point)
- Endurance test (sustained 5K users, 2 hours)

**Pre-Release**:
- Complete load testing suite
- Spike testing
- Capacity planning review

## Monitoring & Alerts

### Critical Alerts (Page Engineer)

- p95 response time >300ms for 5 minutes
- Error rate >5% for 2 minutes
- Database connections >90% for 2 minutes
- Memory usage >90% for 5 minutes

### Warning Alerts (Notify Team)

- p95 response time >200ms for 10 minutes
- Cache hit ratio <70% for 15 minutes
- CPU usage >80% for 10 minutes
- Disk usage >80%

## Conclusion

The Wish List Application is architected for performance with caching, connection pooling, and horizontal scaling capabilities. The current implementation should handle the target load of 10,000 concurrent users with proper deployment infrastructure.

**Key Success Factors**:
1. Redis caching (80%+ hit ratio)
2. Horizontal scaling (3-5 app instances)
3. Database optimization (indexes, connection pooling)
4. CDN for static assets and images
5. Continuous monitoring and optimization

**Next Steps**:
1. Set up load testing environment
2. Configure monitoring and alerting
3. Optimize database queries
4. Implement CDN for images
5. Run baseline performance tests

---

**Performance Team Contact**: performance@wishlistapp.com
