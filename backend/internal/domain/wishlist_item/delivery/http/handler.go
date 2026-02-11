package http

import (
	"errors"
	nethttp "net/http"
	"strconv"

	"wish-list/internal/domain/wishlist_item/delivery/http/dto"
	"wish-list/internal/domain/wishlist_item/service"
	"wish-list/internal/pkg/auth"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for wishlist-item relationships
type Handler struct {
	service service.WishlistItemServiceInterface
}

// NewHandler creates a new Handler
func NewHandler(svc service.WishlistItemServiceInterface) *Handler {
	return &Handler{
		service: svc,
	}
}

// GetWishlistItems godoc
//
//	@Summary		Get items in wishlist
//	@Description	Get all gift items in a specific wishlist with pagination. Public wishlists are accessible without auth.
//	@Tags			Wishlists
//	@Produce		json
//	@Param			id		path		string							true	"Wishlist ID"
//	@Param			page	query		int								false	"Page number (default 1)"
//	@Param			limit	query		int								false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	dto.PaginatedItemsResponse		"List of items in wishlist"
//	@Failure		401		{object}	map[string]string				"Not authenticated"
//	@Failure		403		{object}	map[string]string				"Access denied"
//	@Failure		404		{object}	map[string]string				"Wishlist not found"
//	@Failure		500		{object}	map[string]string				"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items [get]
func (h *Handler) GetWishlistItems(c echo.Context) error {
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
		if errors.Is(err, service.ErrWishListNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, service.ErrWishListForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to get wishlist items",
		})
	}

	return c.JSON(nethttp.StatusOK, dto.PaginatedItemsResponseFromService(result))
}

// AttachItemToWishlist godoc
//
//	@Summary		Attach item to wishlist
//	@Description	Attach an existing gift item to a wishlist. Both item and wishlist must be owned by the authenticated user.
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Wishlist ID"
//	@Param			request	body		dto.AttachItemRequest	true	"Item to attach"
//	@Success		204		{object}	nil						"Item attached successfully"
//	@Failure		400		{object}	map[string]string		"Invalid request body"
//	@Failure		401		{object}	map[string]string		"Not authenticated"
//	@Failure		403		{object}	map[string]string		"Access denied"
//	@Failure		404		{object}	map[string]string		"Wishlist or item not found"
//	@Failure		409		{object}	map[string]string		"Item already attached"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items [post]
func (h *Handler) AttachItemToWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")

	// Parse request body
	var req dto.AttachItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(nethttp.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(nethttp.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Attach item
	err = h.service.AttachItem(ctx, wishlistID, req.ItemID, userID)
	if err != nil {
		if errors.Is(err, service.ErrWishListNotFound) || errors.Is(err, service.ErrItemNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "Wishlist or item not found",
			})
		}
		if errors.Is(err, service.ErrWishListForbidden) || errors.Is(err, service.ErrItemForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		if errors.Is(err, service.ErrItemAlreadyAttached) {
			return c.JSON(nethttp.StatusConflict, map[string]string{
				"error": "Item already attached to this wishlist",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to attach item",
		})
	}

	return c.NoContent(nethttp.StatusNoContent)
}

// CreateItemInWishlist godoc
//
//	@Summary		Create item in wishlist
//	@Description	Create a new gift item and immediately attach it to the specified wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Wishlist ID"
//	@Param			item	body		dto.CreateItemRequest	true	"Item data"
//	@Success		201		{object}	dto.ItemResponse	"Item created and attached successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Wishlist not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id}/items/new [post]
func (h *Handler) CreateItemInWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")

	// Parse request body
	var req dto.CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(nethttp.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(nethttp.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Create and attach item
	item, err := h.service.CreateItemInWishlist(ctx, wishlistID, userID, req.ToDomain())
	if err != nil {
		if errors.Is(err, service.ErrWishListNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, service.ErrWishListForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to create item",
		})
	}

	return c.JSON(nethttp.StatusCreated, dto.ItemResponseFromService(item))
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
func (h *Handler) DetachItemFromWishlist(c echo.Context) error {
	// Get authenticated user ID
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil || userID == "" {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	wishlistID := c.Param("id")
	itemID := c.Param("itemId")

	ctx := c.Request().Context()

	// Detach item
	err = h.service.DetachItem(ctx, wishlistID, itemID, userID)
	if err != nil {
		if errors.Is(err, service.ErrWishListNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "Wishlist not found",
			})
		}
		if errors.Is(err, service.ErrWishListForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}
		if errors.Is(err, service.ErrItemNotInWishlist) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "Item not found in this wishlist",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to detach item",
		})
	}

	return c.NoContent(nethttp.StatusNoContent)
}
