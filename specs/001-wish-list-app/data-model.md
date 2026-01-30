# Data Model: Wish List Application

## Overview
This document defines the data model for the wish list application, including entities, relationships, and validation rules derived from the feature specification.

## Core Entities

### User
Represents an account holder who can create and manage wish lists.

**Fields:**
- `id` (UUID, primary key): Unique identifier for the user
- `email` (VARCHAR(255), unique, not null): User's email address for authentication
- `password_hash` (VARCHAR(255), nullable): Hashed password (nullable for magic-link-only users)
- `first_name` (VARCHAR(100), nullable): User's first name
- `last_name` (VARCHAR(100), nullable): User's last name
- `avatar_url` (TEXT, nullable): URL to user's avatar image
- `is_verified` (BOOLEAN, default: false): Email verification status
- `created_at` (TIMESTAMPTZ, not null): Account creation timestamp
- `updated_at` (TIMESTAMPTZ, not null): Last update timestamp
- `last_login_at` (TIMESTAMPTZ, nullable): Last login timestamp
- `deactivated_at` (TIMESTAMPTZ, nullable): Account deactivation timestamp

**Validation Rules:**
- Email must be valid format
- Password must meet complexity requirements if present
- First and last name must be 1-100 characters if provided

### WishList
A collection of gift items for a specific occasion, owned by a user.

**Fields:**
- `id` (UUID, primary key): Unique identifier for the wish list
- `owner_id` (UUID, foreign key to users.id, not null): Reference to the owner user
- `title` (VARCHAR(200), not null): Title of the wish list
- `description` (TEXT, nullable): Detailed description of the wish list
- `occasion` (VARCHAR(100), nullable): Occasion type (e.g., "Birthday", "Wedding", "Holiday")
- `occasion_date` (DATE, nullable): Date of the occasion
- `template_id` (VARCHAR(50), default: "default"): Template identifier for presentation
- `is_public` (BOOLEAN, default: false): Visibility status for public access
- `public_slug` (VARCHAR(255), unique, nullable): URL-friendly identifier for public access
- `view_count` (INTEGER, default: 0): Number of times the list has been viewed publicly
- `created_at` (TIMESTAMPTZ, not null): Creation timestamp
- `updated_at` (TIMESTAMPTZ, not null): Last update timestamp

**Validation Rules:**
- Title must be 1-200 characters
- Public slug must be unique when list is public
- Occasion date must be in the future if provided

### GiftItem
An item on a wish list with details and status information.

**Fields:**
- `id` (UUID, primary key): Unique identifier for the gift item
- `wishlist_id` (UUID, foreign key to wishlists.id, not null): Reference to the parent wish list
- `name` (VARCHAR(255), not null): Name/title of the gift item
- `description` (TEXT, nullable): Detailed description of the gift
- `link` (TEXT, nullable): URL to the gift item (e.g., product page)
- `image_url` (TEXT, nullable): URL to the gift image (stored in S3)
- `price` (DECIMAL(10,2), nullable): Price of the gift item
- `priority` (INTEGER, default: 0): Priority level (0-10 scale)
- `reserved_by_user_id` (UUID, foreign key to users.id, nullable): Reference to user who reserved the item
- `reserved_at` (TIMESTAMPTZ, nullable): Timestamp when item was reserved
- `purchased_by_user_id` (UUID, foreign key to users.id, nullable): Reference to user who purchased the item
- `purchased_at` (TIMESTAMPTZ, nullable): Timestamp when item was marked as purchased
- `purchased_price` (DECIMAL(10,2), nullable): Actual purchase price
- `notes` (TEXT, nullable): Private notes from the list owner
- `position` (INTEGER, default: 0): Display position in the list
- `created_at` (TIMESTAMPTZ, not null): Creation timestamp
- `updated_at` (TIMESTAMPTZ, not null): Last update timestamp

**Validation Rules:**
- Name must be 1-255 characters
- Link must be a valid URL if provided
- Price must be positive if provided
- Priority must be between 0-10
- Reserved and purchased statuses are mutually exclusive
- Position must be non-negative

### Reservation
A record of a guest reserving a gift item.

**Fields:**
- `id` (UUID, primary key): Unique identifier for the reservation
- `gift_item_id` (UUID, foreign key to gift_items.id, not null): Reference to the reserved gift item
- `reserved_by_user_id` (UUID, foreign key to users.id, nullable): Reference to user who made the reservation (null for anonymous)
- `guest_name` (VARCHAR(200), nullable): Name of the guest if not authenticated
- `guest_email` (VARCHAR(255), nullable): Email of the guest if not authenticated
- `reservation_token` (UUID, unique, not null): Token for anonymous reservation management
- `status` (ENUM: active, cancelled, fulfilled, expired): Current status of the reservation
- `reserved_at` (TIMESTAMPTZ, not null): Timestamp when item was reserved
- `expires_at` (TIMESTAMPTZ, nullable): Expiration timestamp for anonymous reservations
- `canceled_at` (TIMESTAMPTZ, nullable): Timestamp when reservation was cancelled
- `canceled_reason` (TEXT, nullable): Reason for cancellation
- `notification_sent` (BOOLEAN, default: false): Whether reservation notification was sent

**Validation Rules:**
- Either reserved_by_user_id or (guest_name and guest_email) must be provided
- Status must be one of the defined enum values
- Expires_at must be in the future if provided

### Template
Presentation style options for wish lists.

**Fields:**
- `id` (VARCHAR(50), primary key): Unique identifier for the template
- `name` (VARCHAR(100), not null): Display name of the template
- `description` (TEXT, nullable): Description of the template
- `preview_image_url` (TEXT, nullable): URL to preview image
- `config` (JSONB, not null): Configuration options for the template
- `is_default` (BOOLEAN, default: false): Whether this is the default template
- `created_at` (TIMESTAMPTZ, not null): Creation timestamp
- `updated_at` (TIMESTAMPTZ, not null): Last update timestamp

**Validation Rules:**
- ID must be alphanumeric with hyphens/underscores
- Name must be 1-100 characters

## Relationships

### User → WishList (One-to-Many)
- A user can own multiple wish lists
- Foreign key: `wishlists.owner_id` references `users.id`
- Cascade delete: When a user is deleted, their wish lists are also deleted

### WishList → GiftItem (One-to-Many)
- A wish list contains multiple gift items
- Foreign key: `gift_items.wishlist_id` references `wishlists.id`
- Cascade delete: When a wish list is deleted, its gift items are also deleted

### User → GiftItem (Reserved) (Many-to-Many via GiftItem)
- A user can reserve multiple gift items
- A gift item can be reserved by one user at a time
- Foreign key: `gift_items.reserved_by_user_id` references `users.id`

### User → GiftItem (Purchased) (Many-to-Many via GiftItem)
- A user can purchase multiple gift items
- A gift item can be purchased by one user
- Foreign key: `gift_items.purchased_by_user_id` references `users.id`

### GiftItem → Reservation (One-to-Many)
- A gift item can have multiple reservation records (historical)
- Foreign key: `reservations.gift_item_id` references `gift_items.id`

## State Transitions

### Gift Item Status Transitions
```
Available → Reserved → Purchased
    ↓           ↓         ↓
Purchased ← Reserved ← Available
```

- An item starts as "Available"
- When reserved, `reserved_by_user_id` and `reserved_at` are set
- When purchased, `purchased_by_user_id` and `purchased_at` are set, and reservation is fulfilled
- When unreserved, `reserved_by_user_id` and `reserved_at` are cleared
- When unpurchased, `purchased_by_user_id` and `purchased_at` are cleared, returning to available

### Reservation Status Transitions
```
Active → Cancelled | Fulfilled | Expired
```

- Starts as "Active" when created
- Becomes "Cancelled" when user removes reservation
- Becomes "Fulfilled" when gift is marked as purchased
- Becomes "Expired" for anonymous reservations after timeout period

## Indexes

### Essential Indexes
- `users.email` (unique): For authentication lookups
- `wishlists.public_slug` (unique): For public URL lookups
- `wishlists.owner_id`: For user's wish list queries
- `gift_items.wishlist_id`: For wish list item queries
- `gift_items.reserved_by_user_id`: For user's reserved items
- `reservations.gift_item_id`: For checking item reservations
- `reservations.reservation_token` (unique): For anonymous reservation lookups

### Performance Indexes
- `gift_items.position`: For ordered display
- `wishlists.created_at`: For chronological listings
- `gift_items.priority`: For priority-based sorting
- Composite: `(wishlists.owner_id, wishlists.is_public)` for dashboard queries

## Constraints

### Referential Integrity
- All foreign key relationships enforce referential integrity
- Cascade delete behaviors are explicitly defined

### Business Logic Constraints
- A gift item cannot be both reserved and purchased simultaneously
- A gift item can only be reserved by one user at a time
- Public wish lists must have a unique slug
- Reservation expiration applies only to anonymous reservations
