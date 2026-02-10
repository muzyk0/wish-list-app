package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"wish-list/internal/pkg/auth"
	"wish-list/internal/repositories"
	"wish-list/internal/services"

	"github.com/labstack/echo/v4"
)

// ItemHandler handles HTTP requests for gift items as independent resources
type ItemHandler struct {
	service services.ItemServiceInterface
}

// NewItemHandler creates a new ItemHandler
func NewItemHandler(service services.ItemServiceInterface) *ItemHandler {
	return &ItemHandler{
		service: service,
	}
}

// Request/Response DTOs

// CreateItemRequest represents the request to create a gift item
type CreateItemRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255" example:"iPhone 15 Pro"`
	Description string  `json:"description" validate:"max=2000" example:"256GB, Blue Titanium"`
	Link        string  `json:"link" validate:"omitempty,url" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string  `json:"imageUrl" validate:"omitempty,url" example:"https://example.com/image.jpg"`
	Price       float64 `json:"price" validate:"omitempty,gte=0" example:"999.99"`
	Priority    int     `json:"priority" validate:"omitempty,gte=0,lte=5" example:"3"`
	Notes       string  `json:"notes" validate:"max=1000" example:"Preferred color: Blue"`
}

// UpdateItemRequest represents the request to update a gift item
type UpdateItemRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=2000"`
	Link        *string  `json:"link" validate:"omitempty,url"`
	ImageURL    *string  `json:"imageUrl" validate:"omitempty,url"`
	Price       *float64 `json:"price" validate:"omitempty,gte=0"`
	Priority    *int     `json:"priority" validate:"omitempty,gte=0,lte=5"`
	Notes       *string  `json:"notes" validate:"omitempty,max=1000"`
}

// ItemResponse represents a gift item in API responses
type ItemResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     string  `json:"ownerId" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title       string  `json:"title" example:"iPhone 15 Pro"`
	Description string  `json:"description" example:"256GB, Blue Titanium"`
	Link        string  `json:"link" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string  `json:"imageUrl" example:"https://example.com/image.jpg"`
	Price       float64 `json:"price" example:"999.99"`
	Priority    int     `json:"priority" example:"3"`
	Notes       string  `json:"notes" example:"Preferred color: Blue"`
	IsPurchased bool    `json:"isPurchased" example:"false"`
	IsArchived  bool    `json:"isArchived" example:"false"`
	CreatedAt   string  `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   string  `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
}

// PaginatedItemsResponse represents paginated list of items
type PaginatedItemsResponse struct {
	Items      []ItemResponse `json:"items"`
	TotalCount int64          `json:"totalCount" example:"42"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"10"`
	TotalPages int            `json:"totalPages" example:"5"`
}

// MarkPurchasedRequest represents the request to mark item as purchased
type MarkPurchasedRequest struct {
	PurchasedPrice float64 `json:"purchasedPrice" validate:"required,gte=0" example:"899.99"`
}

// GetMyItems godoc
//
//	@Summary		Get my gift items
//	@Description	Get all gift items owned by the authenticated user with pagination and filters
//	@Tags			Items
//	@Produce		json
//	@Param			page			query		int							false	"Page number (default 1)"
//	@Param			limit			query		int							false	"Items per page (default 10, max 100)"
//	@Param			sort			query		string						false	"Sort field (created_at, updated_at, title, price)"
//	@Param			order			query		string						false	"Sort order (asc, desc)"
//	@Param			unattached		query		bool						false	"Filter items not attached to any wishlist"
//	@Param			include_archived	query		bool						false	"Include archived items (default false)"
//	@Param			search			query		string						false	"Search in title and description"
//	@Success		200				{object}	PaginatedItemsResponse		"List of items retrieved successfully"
//	@Failure		400				{object}	map[string]string			"Invalid query parameters"
//	@Failure		401				{object}	map[string]string			"Not authenticated"
//	@Failure		500				{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/items [get]
func (h *ItemHandler) GetMyItems(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Parse filter parameters
	filters := repositories.ItemFilters{
		Sort:            c.QueryParam("sort"),
		Order:           c.QueryParam("order"),
		Unattached:      c.QueryParam("unattached") == "true",
		IncludeArchived: c.QueryParam("include_archived") == "true",
		Search:          c.QueryParam("search"),
		Page:            page,
		Limit:           limit,
	}

	ctx := c.Request().Context()

	// Get items from service
	result, err := h.service.GetMyItems(ctx, userID, filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get items",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// CreateItem godoc
//
//	@Summary		Create gift item
//	@Description	Create a new gift item without attaching it to a wishlist
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			item	body		CreateItemRequest	true	"Item data"
//	@Success		201		{object}	ItemResponse		"Item created successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/items [post]
func (h *ItemHandler) CreateItem(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	// Parse request body
	var req CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert to service input
	input := services.CreateItemInput{
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Priority:    req.Priority,
		Notes:       req.Notes,
	}

	ctx := c.Request().Context()

	// Create item via service
	item, err := h.service.CreateItem(ctx, userID, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create item",
		})
	}

	return c.JSON(http.StatusCreated, item)
}

// GetItem godoc
//
//	@Summary		Get gift item
//	@Description	Get a specific gift item by ID
//	@Tags			Items
//	@Produce		json
//	@Param			id	path		string				true	"Item ID"
//	@Success		200	{object}	ItemResponse		"Item retrieved successfully"
//	@Failure		401	{object}	map[string]string	"Not authenticated"
//	@Failure		403	{object}	map[string]string	"Access denied"
//	@Failure		404	{object}	map[string]string	"Item not found"
//	@Security		BearerAuth
//	@Router			/items/{id} [get]
func (h *ItemHandler) GetItem(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	itemID := c.Param("id")
	ctx := c.Request().Context()

	// Get item via service
	item, err := h.service.GetItem(ctx, itemID, userID)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Item not found",
			})
		}
		if errors.Is(err, services.ErrItemForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get item",
		})
	}

	return c.JSON(http.StatusOK, item)
}

// UpdateItem godoc
//
//	@Summary		Update gift item
//	@Description	Update a gift item by ID
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Item ID"
//	@Param			item	body		UpdateItemRequest	true	"Updated item data"
//	@Success		200		{object}	ItemResponse		"Item updated successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Item not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/items/{id} [put]
func (h *ItemHandler) UpdateItem(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	itemID := c.Param("id")

	// Parse request body
	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert to service input
	input := services.UpdateItemInput{
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Priority:    req.Priority,
		Notes:       req.Notes,
	}

	ctx := c.Request().Context()

	// Update item via service
	item, err := h.service.UpdateItem(ctx, itemID, userID, input)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Item not found",
			})
		}
		if errors.Is(err, services.ErrItemForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update item",
		})
	}

	return c.JSON(http.StatusOK, item)
}

// DeleteItem godoc
//
//	@Summary		Delete gift item (soft delete)
//	@Description	Archive a gift item by setting archived_at timestamp. Item is removed from all queries but data is preserved.
//	@Tags			Items
//	@Produce		json
//	@Param			id	path		string				true	"Item ID"
//	@Success		204	{object}	nil					"Item archived successfully"
//	@Failure		401	{object}	map[string]string	"Not authenticated"
//	@Failure		403	{object}	map[string]string	"Access denied"
//	@Failure		404	{object}	map[string]string	"Item not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/items/{id} [delete]
func (h *ItemHandler) DeleteItem(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	itemID := c.Param("id")
	ctx := c.Request().Context()

	// Soft delete item via service
	err = h.service.SoftDeleteItem(ctx, itemID, userID)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Item not found",
			})
		}
		if errors.Is(err, services.ErrItemForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete item",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// MarkItemAsPurchased godoc
//
//	@Summary		Mark gift item as purchased
//	@Description	Mark a gift item as purchased with the actual purchased price. This is a global status.
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Item ID"
//	@Param			purchase body	MarkPurchasedRequest	true	"Purchase details"
//	@Success		200		{object}	ItemResponse			"Item marked as purchased"
//	@Failure		400		{object}	map[string]string		"Invalid request body"
//	@Failure		401		{object}	map[string]string		"Not authenticated"
//	@Failure		403		{object}	map[string]string		"Access denied"
//	@Failure		404		{object}	map[string]string		"Item not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/items/{id}/mark-purchased [post]
func (h *ItemHandler) MarkItemAsPurchased(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	itemID := c.Param("id")

	// Parse request body
	var req MarkPurchasedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Mark as purchased via service
	item, err := h.service.MarkPurchased(ctx, itemID, userID, req.PurchasedPrice)
	if err != nil {
		if errors.Is(err, services.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Item not found",
			})
		}
		if errors.Is(err, services.ErrItemForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to mark item as purchased",
		})
	}

	return c.JSON(http.StatusOK, item)
}
