package http

import (
	nethttp "net/http"

	"wish-list/internal/domain/item/delivery/http/dto"
	"wish-list/internal/domain/item/repository"
	"wish-list/internal/domain/item/service"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for gift items as independent resources
type Handler struct {
	service service.ItemServiceInterface
}

// NewHandler creates a new Handler
func NewHandler(svc service.ItemServiceInterface) *Handler {
	return &Handler{
		service: svc,
	}
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
//	@Param			attached		query		bool						false	"Filter items attached to any wishlist"
//	@Param			include_archived	query		bool						false	"Include archived items (default false)"
//	@Param			search			query		string						false	"Search in title and description"
//	@Success		200				{object}	dto.PaginatedItemsResponse	"List of items retrieved successfully"
//	@Failure		400				{object}	map[string]string			"Invalid query parameters"
//	@Failure		401				{object}	map[string]string			"Not authenticated"
//	@Failure		500				{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/items [get]
func (h *Handler) GetMyItems(c echo.Context) error {
	userID := auth.MustGetUserID(c)
	pagination := helpers.ParsePagination(c)

	// Parse filter parameters
	filters := repository.ItemFilters{
		Sort:            c.QueryParam("sort"),
		Order:           c.QueryParam("order"),
		Unattached:      c.QueryParam("unattached") == "true",
		Attached:        c.QueryParam("attached") == "true",
		IncludeArchived: c.QueryParam("include_archived") == "true",
		Search:          c.QueryParam("search"),
		Page:            pagination.Page,
		Limit:           pagination.Limit,
	}

	ctx := c.Request().Context()

	// Get items from service
	result, err := h.service.GetMyItems(ctx, userID, filters)
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.JSON(nethttp.StatusOK, dto.PaginatedItemsResponseFromService(result))
}

// CreateItem godoc
//
//	@Summary		Create gift item
//	@Description	Create a new gift item without attaching it to a wishlist
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			item	body		dto.CreateItemRequest	true	"Item data"
//	@Success		201		{object}	dto.ItemResponse		"Item created successfully"
//	@Failure		400		{object}	map[string]string		"Invalid request body"
//	@Failure		401		{object}	map[string]string		"Not authenticated"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/items [post]
func (h *Handler) CreateItem(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	var req dto.CreateItemRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Create item via service
	item, err := h.service.CreateItem(ctx, userID, req.ToDomain())
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.JSON(nethttp.StatusCreated, dto.ItemResponseFromService(item))
}

// GetItem godoc
//
//	@Summary		Get gift item
//	@Description	Get a specific gift item by ID
//	@Tags			Items
//	@Produce		json
//	@Param			id	path		string				true	"Item ID"
//	@Success		200	{object}	dto.ItemResponse	"Item retrieved successfully"
//	@Failure		401	{object}	map[string]string	"Not authenticated"
//	@Failure		403	{object}	map[string]string	"Access denied"
//	@Failure		404	{object}	map[string]string	"Item not found"
//	@Security		BearerAuth
//	@Router			/items/{id} [get]
func (h *Handler) GetItem(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	itemID := c.Param("id")
	ctx := c.Request().Context()

	// Get item via service
	item, err := h.service.GetItem(ctx, itemID, userID)
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.JSON(nethttp.StatusOK, dto.ItemResponseFromService(item))
}

// UpdateItem godoc
//
//	@Summary		Update gift item
//	@Description	Update a gift item by ID
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Item ID"
//	@Param			item	body		dto.UpdateItemRequest	true	"Updated item data"
//	@Success		200		{object}	dto.ItemResponse	"Item updated successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Not authenticated"
//	@Failure		403		{object}	map[string]string	"Access denied"
//	@Failure		404		{object}	map[string]string	"Item not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/items/{id} [put]
func (h *Handler) UpdateItem(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	itemID := c.Param("id")

	var req dto.UpdateItemRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Update item via service
	item, err := h.service.UpdateItem(ctx, itemID, userID, req.ToDomain())
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.JSON(nethttp.StatusOK, dto.ItemResponseFromService(item))
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
func (h *Handler) DeleteItem(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	itemID := c.Param("id")
	ctx := c.Request().Context()

	// Soft delete item via service
	err := h.service.SoftDeleteItem(ctx, itemID, userID)
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.NoContent(nethttp.StatusNoContent)
}

// MarkItemAsPurchased godoc
//
//	@Summary		Mark gift item as purchased
//	@Description	Mark a gift item as purchased with the actual purchased price. This is a global status.
//	@Tags			Items
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Item ID"
//	@Param			purchase body	dto.MarkPurchasedRequest	true	"Purchase details"
//	@Success		200		{object}	dto.ItemResponse		"Item marked as purchased"
//	@Failure		400		{object}	map[string]string		"Invalid request body"
//	@Failure		401		{object}	map[string]string		"Not authenticated"
//	@Failure		403		{object}	map[string]string		"Access denied"
//	@Failure		404		{object}	map[string]string		"Item not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/items/{id}/mark-purchased [post]
func (h *Handler) MarkItemAsPurchased(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	itemID := c.Param("id")

	var req dto.MarkPurchasedRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Mark as purchased via service
	item, err := h.service.MarkPurchased(ctx, itemID, userID, req.PurchasedPrice)
	if err != nil {
		return mapItemServiceError(err)
	}

	return c.JSON(nethttp.StatusOK, dto.ItemResponseFromService(item))
}
