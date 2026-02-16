package http

import (
	nethttp "net/http"

	"wish-list/internal/domain/wishlist_item/delivery/http/dto"
	"wish-list/internal/domain/wishlist_item/service"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

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
	pagination := helpers.ParsePagination(c)

	ctx := c.Request().Context()

	// Get items
	result, err := h.service.GetWishlistItems(ctx, wishlistID, userID, pagination.Page, pagination.Limit)
	if err != nil {
		return mapWishlistItemServiceError(err)
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
	userID := auth.MustGetUserID(c)

	wishlistID := c.Param("id")

	var req dto.AttachItemRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Attach item
	err := h.service.AttachItem(ctx, wishlistID, req.ItemID, userID)
	if err != nil {
		return mapWishlistItemServiceError(err)
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
	userID := auth.MustGetUserID(c)

	wishlistID := c.Param("id")

	var req dto.CreateItemRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Create and attach item
	item, err := h.service.CreateItemInWishlist(ctx, wishlistID, userID, req.ToDomain())
	if err != nil {
		return mapWishlistItemServiceError(err)
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
	userID := auth.MustGetUserID(c)

	wishlistID := c.Param("id")
	itemID := c.Param("itemId")

	ctx := c.Request().Context()

	// Detach item
	err := h.service.DetachItem(ctx, wishlistID, itemID, userID)
	if err != nil {
		return mapWishlistItemServiceError(err)
	}

	return c.NoContent(nethttp.StatusNoContent)
}
