package http

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"strconv"

	"wish-list/internal/domain/wishlist/delivery/http/dto"
	"wish-list/internal/domain/wishlist/service"
	"wish-list/internal/pkg/auth"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for wishlists
type Handler struct {
	service service.WishListServiceInterface
}

// NewHandler creates a new Handler
func NewHandler(svc service.WishListServiceInterface) *Handler {
	return &Handler{
		service: svc,
	}
}

// CreateWishList godoc
//
//	@Summary		Create a new wish list
//	@Description	Create a new wish list for the authenticated user
//	@Tags			Wish Lists
//	@Accept			json
//	@Produce		json
//	@Param			wish_list	body		dto.CreateWishListRequest	true	"Wish list creation information"
//	@Success		201			{object}	dto.WishListResponse		"Wish list created successfully"
//	@Failure		400			{object}	map[string]string			"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string			"Unauthorized"
//	@Failure		500			{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists [post]
func (h *Handler) CreateWishList(c echo.Context) error {
	var req dto.CreateWishListRequest
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

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	wishList, err := h.service.CreateWishList(ctx, userID, req.ToServiceInput())
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to create wish list: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusCreated, dto.FromWishListOutput(wishList))
}

// GetWishList godoc
//
//	@Summary		Get a wish list by ID
//	@Description	Get a wish list by its ID. If the wish list is private, the user must be the owner.
//	@Tags			Wish Lists
//	@Produce		json
//	@Param			id	path		string					true	"Wish List ID"
//	@Success		200	{object}	dto.WishListResponse	"Wish list retrieved successfully"
//	@Failure		403	{object}	map[string]string		"Access denied"
//	@Failure		404	{object}	map[string]string		"Wish list not found"
//	@Security		BearerAuth
//	@Router			/wishlists/{id} [get]
func (h *Handler) GetWishList(c echo.Context) error {
	wishListID := c.Param("id")

	ctx := c.Request().Context()
	wishList, err := h.service.GetWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(nethttp.StatusNotFound, map[string]string{
			"error": fmt.Errorf("wish list not found: %w", err).Error(),
		})
	}

	// Get user from context to check ownership
	currentUserID, _, _, _ := auth.GetUserFromContext(c)

	// If not the owner and not public, return forbidden
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return c.JSON(nethttp.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromWishListOutput(wishList))
}

// GetWishListsByOwner godoc
//
//	@Summary		Get all wish lists owned by the authenticated user
//	@Description	Get all wish lists owned by the currently authenticated user. Includes item_count for each wishlist.
//	@Tags			Wish Lists
//	@Produce		json
//	@Success		200	{array}		dto.WishListResponse	"List of wish lists retrieved successfully (includes item_count)"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists [get]
func (h *Handler) GetWishListsByOwner(c echo.Context) error {
	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	wishLists, err := h.service.GetWishListsByOwner(ctx, userID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get wish lists: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromWishListOutputs(wishLists))
}

// UpdateWishList godoc
//
//	@Summary		Update a wish list
//	@Description	Update a wish list by its ID. The user must be the owner of the wish list.
//	@Tags			Wish Lists
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string						true	"Wish List ID"
//	@Param			wish_list	body		dto.UpdateWishListRequest	true	"Wish list update information"
//	@Success		200			{object}	dto.WishListResponse		"Wish list updated successfully"
//	@Failure		400			{object}	map[string]string			"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string			"Unauthorized"
//	@Failure		403			{object}	map[string]string			"Forbidden"
//	@Failure		404			{object}	map[string]string			"Wish list not found"
//	@Failure		500			{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id} [put]
func (h *Handler) UpdateWishList(c echo.Context) error {
	wishListID := c.Param("id")

	var req dto.UpdateWishListRequest
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

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	wishList, err := h.service.UpdateWishList(ctx, wishListID, userID, req.ToServiceInput())
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, service.ErrWishListNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "wish list not found",
			})
		}
		// Check if it's a forbidden error
		if errors.Is(err, service.ErrWishListForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to update wish list: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromWishListOutput(wishList))
}

// DeleteWishList godoc
//
//	@Summary		Delete a wish list
//	@Description	Delete a wish list by its ID. The user must be the owner of the wish list.
//	@Tags			Wish Lists
//	@Produce		json
//	@Param			id	path		string				true	"Wish List ID"
//	@Success		204	{object}	nil					"Wish list deleted successfully"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		403	{object}	map[string]string	"Forbidden"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id} [delete]
func (h *Handler) DeleteWishList(c echo.Context) error {
	wishListID := c.Param("id")

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	err = h.service.DeleteWishList(ctx, wishListID, userID)
	if err != nil {
		// Check if it's a forbidden error
		if errors.Is(err, service.ErrWishListForbidden) {
			return c.JSON(nethttp.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to delete wish list: %w", err).Error(),
		})
	}

	return c.NoContent(nethttp.StatusNoContent)
}

// GetWishListByPublicSlug godoc
//
//	@Summary		Get a public wish list by its slug
//	@Description	Get a public wish list by its public slug. The wish list must be marked as public.
//	@Tags			Wish Lists
//	@Produce		json
//	@Param			slug	path		string					true	"Public Slug"
//	@Success		200		{object}	dto.WishListResponse	"Public wish list retrieved successfully"
//	@Failure		404		{object}	map[string]string		"Wish list not found"
//	@Router			/public/wishlists/{slug} [get]
func (h *Handler) GetWishListByPublicSlug(c echo.Context) error {
	publicSlug := c.Param("slug")

	ctx := c.Request().Context()
	wishList, err := h.service.GetWishListByPublicSlug(ctx, publicSlug)
	if err != nil {
		return c.JSON(nethttp.StatusNotFound, map[string]string{
			"error": fmt.Errorf("wish list not found: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromWishListOutput(wishList))
}

// GetGiftItemsByPublicSlug godoc
//
//	@Summary		Get gift items for a public wish list by slug
//	@Description	Get all gift items for a public wish list by its public slug with pagination support.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			slug	path		string						true	"Public Slug"
//	@Param			page	query		int							false	"Page number (default 1)"
//	@Param			limit	query		int							false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	dto.GetGiftItemsResponse	"Gift items retrieved successfully"
//	@Failure		404		{object}	map[string]string			"Wish list not found or not public"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/public/wishlists/{slug}/gift-items [get]
func (h *Handler) GetGiftItemsByPublicSlug(c echo.Context) error {
	publicSlug := c.Param("slug")

	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	ctx := c.Request().Context()

	// Get the wishlist by public slug to verify it's public
	wishList, err := h.service.GetWishListByPublicSlug(ctx, publicSlug)
	if err != nil {
		return c.JSON(nethttp.StatusNotFound, map[string]string{
			"error": "Wish list not found or not public",
		})
	}

	// Get all gift items for this wishlist
	giftItems, err := h.service.GetGiftItemsByWishList(ctx, wishList.ID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get gift items: %w", err).Error(),
		})
	}

	if giftItems == nil {
		giftItems = []*service.GiftItemOutput{}
	}

	// Apply pagination
	total := len(giftItems)
	start := (page - 1) * limit
	end := min(start+limit, total)

	// Handle out of bounds
	if start > total {
		start = total
		end = total
	}

	paginatedItems := giftItems[start:end]

	// Calculate total pages
	pages := (total + limit - 1) / limit

	return c.JSON(nethttp.StatusOK, dto.GetGiftItemsResponse{
		Items: dto.FromGiftItemOutputs(paginatedItems),
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	})
}
