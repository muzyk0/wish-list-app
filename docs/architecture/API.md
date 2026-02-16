# Wish List Application API Documentation

## Overview

The Wish List Application API provides endpoints for managing wish lists, gift items, and reservations. The API follows REST principles and uses JSON for request and response bodies.

**Base URL**: `http://localhost:8080/api`

**Authentication**: JWT Bearer Token (for protected endpoints)

## API Specifications

The complete OpenAPI 3.0 specifications are available in the `/contracts` directory:

- `user-api.json` - User authentication and profile management
- `wishlist-api.json` - Wish list CRUD operations
- `gift-item-api.json` - Gift item management
- `reservation-api.json` - Gift reservation system

## Authentication

### Register a new user

\`\`\`http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "John Doe"
}
\`\`\`

**Response:**
\`\`\`json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
\`\`\`

### Login

\`\`\`http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
\`\`\`

**Response:**
\`\`\`json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com"
  }
}
\`\`\`

### Using the JWT Token

Include the JWT token in the \`Authorization\` header for protected endpoints:

\`\`\`http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
\`\`\`

## User Profile Management

### Get User Profile (Protected)

\`\`\`http
GET /api/protected/profile
Authorization: Bearer {token}
\`\`\`

### Update User Profile (Protected)

\`\`\`http
PUT /api/protected/profile
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com"
}
\`\`\`

### Delete Account (Protected)

\`\`\`http
DELETE /api/protected/account
Authorization: Bearer {token}
\`\`\`

## Error Responses

All endpoints return standard error responses with appropriate HTTP status codes and error messages.

**Common HTTP Status Codes:**
- \`200 OK\` - Request successful
- \`201 Created\` - Resource created successfully  
- \`204 No Content\` - Request successful, no response body
- \`400 Bad Request\` - Invalid request data
- \`401 Unauthorized\` - Missing or invalid authentication
- \`403 Forbidden\` - Insufficient permissions
- \`404 Not Found\` - Resource not found
- \`409 Conflict\` - Resource conflict
- \`429 Too Many Requests\` - Rate limit exceeded
- \`500 Internal Server Error\` - Server error

## Rate Limiting

API endpoints are rate-limited to prevent abuse:
- **Default**: 100 requests per minute per IP address

## Caching

Public wish list endpoints are cached for improved performance:
- **Cache TTL**: 15 minutes (configurable via \`CACHE_TTL_MINUTES\`)

## Security

- **Authentication**: JWT-based authentication
- **Encryption**: PII data encrypted at rest
- **HTTPS**: All production endpoints use HTTPS
- **Rate Limiting**: Protection against abuse
- **Input Validation**: All inputs validated and sanitized

For detailed API specifications, see the OpenAPI spec files in \`/contracts\`.
