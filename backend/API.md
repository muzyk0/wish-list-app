# API Documentation: Wish List Application

## Overview
The Wish List application provides a REST API for managing wish lists and gift items. The API is built with Go and Echo framework, using PostgreSQL as the database with sqlx for database operations.

## Authentication
The API uses JWT-based authentication for protected endpoints. Public endpoints do not require authentication.

### Headers
- For authenticated requests: `Authorization: Bearer {jwt-token}`
- Content-Type: `application/json`

## Endpoints

### Authentication

#### POST /api/auth/register
Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "secure_password",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_verified": false,
    "created_at": "2026-01-12T10:00:00Z"
  },
  "token": "jwt-token-string"
}
```

#### POST /api/auth/login
Log in an existing user.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_verified": false,
    "created_at": "2026-01-12T10:00:00Z"
  },
  "token": "jwt-token-string"
}
```

### Wish Lists

#### POST /api/wishlists (Requires Auth)
Create a new wish list.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Request Body:**
```json
{
  "title": "Birthday Gifts",
  "description": "Gifts I'd like for my birthday",
  "occasion": "Birthday",
  "occasion_date": "2026-05-15",
  "template_id": "default",
  "is_public": true
}
```

**Response:**
```json
{
  "id": "uuid-string",
  "owner_id": "uuid-string",
  "title": "Birthday Gifts",
  "description": "Gifts I'd like for my birthday",
  "occasion": "Birthday",
  "occasion_date": "2026-05-15",
  "template_id": "default",
  "is_public": true,
  "public_slug": "birthday-gifts-1234",
  "view_count": 0,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z"
}
```

#### GET /api/wishlists/:id (Requires Auth)
Get a specific wish list by ID.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
```json
{
  "id": "uuid-string",
  "owner_id": "uuid-string",
  "title": "Birthday Gifts",
  "description": "Gifts I'd like for my birthday",
  "occasion": "Birthday",
  "occasion_date": "2026-05-15",
  "template_id": "default",
  "is_public": true,
  "public_slug": "birthday-gifts-1234",
  "view_count": 5,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z"
}
```

#### GET /api/wishlists (Requires Auth)
Get all wish lists for the authenticated user.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
```json
[
  {
    "id": "uuid-string",
    "owner_id": "uuid-string",
    "title": "Birthday Gifts",
    "description": "Gifts I'd like for my birthday",
    "occasion": "Birthday",
    "occasion_date": "2026-05-15",
    "template_id": "default",
    "is_public": true,
    "public_slug": "birthday-gifts-1234",
    "view_count": 5,
    "created_at": "2026-01-12T10:00:00Z",
    "updated_at": "2026-01-12T10:00:00Z"
  }
]
```

#### PUT /api/wishlists/:id (Requires Auth)
Update a specific wish list.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Request Body:**
```json
{
  "title": "Updated Birthday Gifts",
  "description": "Updated description",
  "is_public": false
}
```

**Response:**
```json
{
  "id": "uuid-string",
  "owner_id": "uuid-string",
  "title": "Updated Birthday Gifts",
  "description": "Updated description",
  "occasion": "Birthday",
  "occasion_date": "2026-05-15",
  "template_id": "default",
  "is_public": false,
  "public_slug": null,
  "view_count": 5,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T11:00:00Z"
}
```

#### DELETE /api/wishlists/:id (Requires Auth)
Delete a specific wish list.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
- Status: 204 No Content

### Gift Items

#### POST /api/gift-items/wishlist/:wishlistId (Requires Auth)
Create a new gift item in a wish list.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Request Body:**
```json
{
  "name": "Wireless Headphones",
  "description": "High-quality noise cancelling headphones",
  "link": "https://example.com/headphones",
  "image_url": "https://example.com/image.jpg",
  "price": 199.99,
  "priority": 8,
  "notes": "Black color preferred",
  "position": 1
}
```

**Response:**
```json
{
  "id": "uuid-string",
  "wishlist_id": "uuid-string",
  "name": "Wireless Headphones",
  "description": "High-quality noise cancelling headphones",
  "link": "https://example.com/headphones",
  "image_url": "https://example.com/image.jpg",
  "price": 199.99,
  "priority": 8,
  "reserved_by_user_id": null,
  "reserved_at": null,
  "purchased_by_user_id": null,
  "purchased_at": null,
  "purchased_price": null,
  "notes": "Black color preferred",
  "position": 1,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z"
}
```

#### GET /api/gift-items/:id (Requires Auth)
Get a specific gift item by ID.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
```json
{
  "id": "uuid-string",
  "wishlist_id": "uuid-string",
  "name": "Wireless Headphones",
  "description": "High-quality noise cancelling headphones",
  "link": "https://example.com/headphones",
  "image_url": "https://example.com/image.jpg",
  "price": 199.99,
  "priority": 8,
  "reserved_by_user_id": null,
  "reserved_at": null,
  "purchased_by_user_id": null,
  "purchased_at": null,
  "purchased_price": null,
  "notes": "Black color preferred",
  "position": 1,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z"
}
```

#### GET /api/gift-items/wishlist/:wishlistId (Requires Auth)
Get all gift items in a specific wish list.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
```json
{
  "items": [
    {
      "id": "uuid-string",
      "wishlist_id": "uuid-string",
      "name": "Wireless Headphones",
      "description": "High-quality noise cancelling headphones",
      "link": "https://example.com/headphones",
      "image_url": "https://example.com/image.jpg",
      "price": 199.99,
      "priority": 8,
      "reserved_by_user_id": null,
      "reserved_at": null,
      "purchased_by_user_id": null,
      "purchased_at": null,
      "purchased_price": null,
      "notes": "Black color preferred",
      "position": 1,
      "created_at": "2026-01-12T10:00:00Z",
      "updated_at": "2026-01-12T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "pages": 1
}
```

#### PUT /api/gift-items/:id (Requires Auth)
Update a specific gift item.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Request Body:**
```json
{
  "name": "Updated Wireless Headphones",
  "description": "Updated description",
  "priority": 9
}
```

**Response:**
```json
{
  "id": "uuid-string",
  "wishlist_id": "uuid-string",
  "name": "Updated Wireless Headphones",
  "description": "Updated description",
  "link": "https://example.com/headphones",
  "image_url": "https://example.com/image.jpg",
  "price": 199.99,
  "priority": 9,
  "reserved_by_user_id": null,
  "reserved_at": null,
  "purchased_by_user_id": null,
  "purchased_at": null,
  "purchased_price": null,
  "notes": "Black color preferred",
  "position": 1,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T11:00:00Z"
}
```

#### DELETE /api/gift-items/:id (Requires Auth)
Delete a specific gift item.

**Headers:**
- `Authorization: Bearer {jwt-token}`

**Response:**
- Status: 204 No Content

### Public Endpoints

#### GET /api/public/lists/:slug
Get a public wish list by its slug.

**Response:**
```json
{
  "id": "uuid-string",
  "owner_id": "uuid-string",
  "title": "Birthday Gifts",
  "description": "Gifts I'd like for my birthday",
  "occasion": "Birthday",
  "occasion_date": "2026-05-15",
  "template_id": "default",
  "is_public": true,
  "public_slug": "birthday-gifts-1234",
  "view_count": 5,
  "created_at": "2026-01-12T10:00:00Z",
  "updated_at": "2026-01-12T10:00:00Z"
}
```

## Error Responses
All error responses follow this format:
```json
{
  "error": "Descriptive error message"
}
```

## Status Codes
- 200: Success
- 201: Created
- 204: No Content
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 409: Conflict
- 500: Internal Server Error
