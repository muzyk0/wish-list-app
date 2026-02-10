package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"wish-list/internal/pkg/auth"
	"wish-list/internal/services"

	"github.com/labstack/echo/v4"
)

// WishlistItemHandlerInterface defines the contract for wishlist-item HTTP handlers
type WishlistItemHandlerInterface interface {
	GetWishlistItems(c echo.Context) error
	AttachItemToWishlist(c echo.Context) error
	CreateItemInWishlist(c echo.Context) error
	DetachItemFromWishlist(c echo.Context) error
}

// WishlistItemHandler handles HTTP requests for wishlist-item relationships
type WishlistItemHandler struct {
	service services.WishlistItemServiceInterface
}

// NewWishlistItemHandler creates a new WishlistItemHandler
func NewWishlistItemHandler(service services.WishlistItemServiceInterface) *WishlistItemHandler {
	return &WishlistItemHandler{
		service: service,
	}
}

// Request DTOs

// AttachItemRequest represents the request to attach an existing item to a wishlist
type AttachItemRequest struct {
	ItemID string `json:"itemId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetWishlistItems godoc
//
//	@Summary		Get items in wishlist
//	@Description	Get all gift items in a specific wishlist with pagination. Public wishlists are accessible without auth.
//	@Tags			Wishlists
//	@Produce		json
//	@Param			id		path		string						true	"Wishlist ID"
//	@Param			page	query		int							false	"Page number (default 1)"
//	@Param			limit	query		int							false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	PaginatedItemsResponse		"List of items in wishlist"
//	@Failure		401		{object}	map[string]string			"Not authenticated"
//	@Failure		403		{object}	map[string]string			"Access denied"
//	@Failure		404		{object}	map[string]string			"Wishlist not found"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items [get]
func (h *WishlistItemHandler) GetWishlistItems(c echo.Context) error {
	// Get authenticated user ID (optional for public wishlists)
	userID, _, _, _ := auth.GetUserFromContext(c)

	wishlistID := c.Param("id")

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

	ctx := c.Request().Context()

	// Get items
	result, err := h.service.GetWishlistItems(ctx, wishlistID, userID, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrWishListNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, services.ErrWishListForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get wishlist items",
		})
	}

	return c.JSON(http.StatusOK, result)
}

// AttachItemToWishlist godoc
//
//	@Summary		Attach item to wishlist
//	@Description	Attach an existing gift item to a wishlist. Both item and wishlist must be owned by the authenticated user.
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Wishlist ID"
//	@Param			request	body		AttachItemRequest	true	"Item to attach"
//	@Success		204		{object}	nil					"Item attached successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Wishlist or item not found"
//	@Failure		409		{object}	map[string]string	"Item already attached"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items [post]
func (h *WishlistItemHandler) AttachItemToWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")

	// Parse request body
	var req AttachItemRequest
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

	// Attach item
	err = h.service.AttachItem(ctx, wishlistID, req.ItemID, userID)
	if err != nil {
		if errors.Is(err, services.ErrWishListNotFound) || errors.Is(err, services.ErrItemNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Wishlist or item not found",
			})
		}
		if errors.Is(err, services.ErrWishListForbidden) || errors.Is(err, services.ErrItemForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		if errors.Is(err, services.ErrItemAlreadyAttached) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Item already attached to this wishlist",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to attach item",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateItemInWishlist godoc
//
//	@Summary		Create item in wishlist
//	@Description	Create a new gift item and immediately attach it to the specified wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Wishlist ID"
//	@Param			item	body		CreateItemRequest	true	"Item data"
//	@Success		201		{object}	ItemResponse		"Item created and attached successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Wishlist not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items/new [post]
func (h *WishlistItemHandler) CreateItemInWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")

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

	// Create and attach item
	item, err := h.service.CreateItemInWishlist(ctx, wishlistID, userID, input)
	if err != nil {
		if errors.Is(err, services.ErrWishListNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, services.ErrWishListForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create item",
		})
	}

	return c.JSON(http.StatusCreated, item)
}

// DetachItemFromWishlist godoc
//
//	@Summary		Detach item from wishlist
//	@Description	Remove a gift item from a wishlist. The item itself is not deleted, only the association is removed.
//	@Tags			Wishlists
//	@Produce		json
//	@Param			id		path		string				true	"Wishlist ID"
//	@Param			itemId	path		string				true	"Item ID"
//	@Success		204		{object}	nil					"Item detached successfully"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Wishlist or item not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items/{itemId} [delete]
func (h *WishlistItemHandler) DetachItemFromWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")
	itemID := c.Param("itemId")

	ctx := c.Request().Context()

	// Detach item
	err = h.service.DetachItem(ctx, wishlistID, itemID, userID)
	if err != nil {
		if errors.Is(err, services.ErrWishListNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, services.ErrWishListForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		if errors.Is(err, services.ErrItemNotInWishlist) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Item not found in this wishlist",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to detach item",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
