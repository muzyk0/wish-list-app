package http

import (
	nethttp "net/http"

	"wish-list/internal/domain/wishlist/delivery/http/dto"
	"wish-list/internal/domain/wishlist/service"
	"wish-list/internal/pkg/apperrors"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

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
	userID := auth.MustGetUserID(c)

	var req dto.CreateWishListRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	wishList, err := h.service.CreateWishList(ctx, userID, req.ToServiceInput())
	if err != nil {
		return mapWishlistServiceError(err)
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
		return mapWishlistServiceError(err)
	}

	// Get user from context to check ownership (optional for public wishlists)
	currentUserID, _, _, _ := auth.GetUserFromContext(c)

	// If not the owner and not public, return forbidden
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return apperrors.Forbidden("Access denied")
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
	userID := auth.MustGetUserID(c)

	ctx := c.Request().Context()
	wishLists, err := h.service.GetWishListsByOwner(ctx, userID)
	if err != nil {
		return mapWishlistServiceError(err)
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
	userID := auth.MustGetUserID(c)

	wishListID := c.Param("id")

	var req dto.UpdateWishListRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	wishList, err := h.service.UpdateWishList(ctx, wishListID, userID, req.ToServiceInput())
	if err != nil {
		return mapWishlistServiceError(err)
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
	userID := auth.MustGetUserID(c)

	wishListID := c.Param("id")

	ctx := c.Request().Context()
	err := h.service.DeleteWishList(ctx, wishListID, userID)
	if err != nil {
		return mapWishlistServiceError(err)
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
		return mapWishlistServiceError(err)
	}

	return c.JSON(nethttp.StatusOK, dto.FromWishListOutput(wishList))
}

// GetGiftItemsByPublicSlug godoc
//
//	@Summary		Get gift items for a public wish list by slug
//	@Description	Get gift items for a public wish list by its public slug with pagination, search, status filter, and sort support.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			slug	path		string						true	"Public Slug"
//	@Param			page	query		int							false	"Page number (default 1)"
//	@Param			limit	query		int							false	"Items per page (default 12, max 100)"
//	@Param			search	query		string						false	"Case-insensitive search on name and description"
//	@Param			status	query		string						false	"Filter by status"					Enums(available, reserved, purchased)
//	@Param			sort_by	query		string						false	"Sort order"						Enums(position, name_asc, name_desc, price_asc, price_desc, priority_desc)
//	@Success		200		{object}	dto.GetGiftItemsResponse	"Gift items retrieved successfully"
//	@Failure		404		{object}	map[string]string			"Wish list not found or not public"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/public/wishlists/{slug}/gift-items [get]
func (h *Handler) GetGiftItemsByPublicSlug(c echo.Context) error {
	publicSlug := c.Param("slug")
	pagination := helpers.ParsePagination(c)

	search := c.QueryParam("search")
	status := c.QueryParam("status")
	sortBy := c.QueryParam("sort_by")

	// Silently reset invalid enum values to empty string (no error to caller)
	validStatuses := map[string]bool{"available": true, "reserved": true, "purchased": true}
	if !validStatuses[status] {
		status = ""
	}
	validSortBys := map[string]bool{"position": true, "name_asc": true, "name_desc": true, "price_asc": true, "price_desc": true, "priority_desc": true}
	if !validSortBys[sortBy] {
		sortBy = ""
	}

	ctx := c.Request().Context()

	// Verify the wishlist exists and is public
	_, err := h.service.GetWishListByPublicSlug(ctx, publicSlug)
	if err != nil {
		return apperrors.NotFound("Wish list not found or not public")
	}

	offset := (pagination.Page - 1) * pagination.Limit
	giftItems, totalCount, err := h.service.GetGiftItemsByPublicSlugFiltered(ctx, publicSlug, service.PublicItemFiltersInput{
		Limit:  pagination.Limit,
		Offset: offset,
		Search: search,
		Status: status,
		SortBy: sortBy,
	})
	if err != nil {
		return apperrors.Internal("Failed to get gift items").Wrap(err)
	}

	if giftItems == nil {
		giftItems = []*service.GiftItemOutput{}
	}

	// Calculate total pages
	pages := (totalCount + pagination.Limit - 1) / pagination.Limit

	return c.JSON(nethttp.StatusOK, dto.GetGiftItemsResponse{
		Items: dto.FromGiftItemOutputs(giftItems),
		Total: totalCount,
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Pages: pages,
	})
}
