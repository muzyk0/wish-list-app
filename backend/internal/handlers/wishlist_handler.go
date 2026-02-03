package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"wish-list/internal/auth"
	"wish-list/internal/services"

	"github.com/labstack/echo/v4"
)

type WishListHandler struct {
	service services.WishListServiceInterface
}

func NewWishListHandler(service services.WishListServiceInterface) *WishListHandler {
	return &WishListHandler{
		service: service,
	}
}

type CreateWishListRequest struct {
	Title        string `json:"title" validate:"required,max=200"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	TemplateID   string `json:"template_id" default:"default"`
	IsPublic     bool   `json:"is_public"`
}

type UpdateWishListRequest struct {
	Title        *string `json:"title" validate:"omitempty,max=200"`
	Description  *string `json:"description"`
	Occasion     *string `json:"occasion"`
	OccasionDate *string `json:"occasion_date"`
	TemplateID   *string `json:"template_id"`
	IsPublic     *bool   `json:"is_public"`
}

type CreateGiftItemRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Link        string  `json:"link" validate:"omitempty,url"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
	Price       float64 `json:"price" validate:"omitempty,min=0"`
	Priority    int     `json:"priority" validate:"omitempty,min=0,max=10"`
	Notes       string  `json:"notes"`
	Position    int     `json:"position" validate:"omitempty,min=0"`
}

type UpdateGiftItemRequest struct {
	Name        *string  `json:"name" validate:"omitempty,max=255"`
	Description *string  `json:"description"`
	Link        *string  `json:"link" validate:"omitempty,url"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url"`
	Price       *float64 `json:"price" validate:"omitempty,min=0"`
	Priority    *int     `json:"priority" validate:"omitempty,min=0,max=10"`
	Notes       *string  `json:"notes"`
	Position    *int     `json:"position" validate:"omitempty,min=0"`
}

// WishListResponse is the handler-level DTO for wishlist data
type WishListResponse struct {
	ID           string `json:"id" validate:"required"`
	OwnerID      string `json:"owner_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	TemplateID   string `json:"template_id"`
	IsPublic     bool   `json:"is_public"`
	PublicSlug   string `json:"public_slug"`
	ViewCount    string `json:"view_count" validate:"required"`
	CreatedAt    string `json:"created_at" validate:"required"`
	UpdatedAt    string `json:"updated_at" validate:"required"`
}

// GiftItemResponse is the handler-level DTO for gift item data
type GiftItemResponse struct {
	ID                string  `json:"id" validate:"required"`
	WishlistID        string  `json:"wishlist_id" validate:"required"`
	Name              string  `json:"name" validate:"required"`
	Description       string  `json:"description"`
	Link              string  `json:"link"`
	ImageURL          string  `json:"image_url"`
	Price             float64 `json:"price"`
	Priority          int     `json:"priority"`
	ReservedByUserID  string  `json:"reserved_by_user_id"`
	ReservedAt        string  `json:"reserved_at"`
	PurchasedByUserID string  `json:"purchased_by_user_id"`
	PurchasedAt       string  `json:"purchased_at"`
	PurchasedPrice    float64 `json:"purchased_price"`
	Notes             string  `json:"notes"`
	Position          int     `json:"position"`
	CreatedAt         string  `json:"created_at" validate:"required"`
	UpdatedAt         string  `json:"updated_at" validate:"required"`
}

type GetGiftItemsResponse struct {
	Items []*GiftItemResponse `json:"items" validate:"required"`
	Total int                 `json:"total" validate:"required"`
	Page  int                 `json:"page" validate:"required"`
	Limit int                 `json:"limit" validate:"required"`
	Pages int                 `json:"pages" validate:"required"`
}

type PurchaseRequest struct {
	PurchasedPrice float64 `json:"purchased_price"`
}

// toWishListResponse maps service layer WishListOutput to handler layer WishListResponse
func (h *WishListHandler) toWishListResponse(wl *services.WishListOutput) *WishListResponse {
	if wl == nil {
		return nil
	}
	return &WishListResponse{
		ID:           wl.ID,
		OwnerID:      wl.OwnerID,
		Title:        wl.Title,
		Description:  wl.Description,
		Occasion:     wl.Occasion,
		OccasionDate: wl.OccasionDate,
		TemplateID:   wl.TemplateID,
		IsPublic:     wl.IsPublic,
		PublicSlug:   wl.PublicSlug,
		ViewCount:    fmt.Sprintf("%d", wl.ViewCount),
		CreatedAt:    wl.CreatedAt,
		UpdatedAt:    wl.UpdatedAt,
	}
}

// toGiftItemResponse maps service layer GiftItemOutput to handler layer GiftItemResponse
func (h *WishListHandler) toGiftItemResponse(item *services.GiftItemOutput) *GiftItemResponse {
	if item == nil {
		return nil
	}
	return &GiftItemResponse{
		ID:                item.ID,
		WishlistID:        item.WishlistID,
		Name:              item.Name,
		Description:       item.Description,
		Link:              item.Link,
		ImageURL:          item.ImageURL,
		Price:             item.Price,
		Priority:          item.Priority,
		ReservedByUserID:  item.ReservedByUserID,
		ReservedAt:        item.ReservedAt,
		PurchasedByUserID: item.PurchasedByUserID,
		PurchasedAt:       item.PurchasedAt,
		PurchasedPrice:    item.PurchasedPrice,
		Notes:             item.Notes,
		Position:          item.Position,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

// toWishListResponses maps array of service layer WishListOutput to handler layer WishListResponse
func (h *WishListHandler) toWishListResponses(wishlists []*services.WishListOutput) []*WishListResponse {
	if wishlists == nil {
		return nil
	}
	responses := make([]*WishListResponse, len(wishlists))
	for i, wl := range wishlists {
		responses[i] = h.toWishListResponse(wl)
	}
	return responses
}

// toGiftItemResponses maps array of service layer GiftItemOutput to handler layer GiftItemResponse
func (h *WishListHandler) toGiftItemResponses(items []*services.GiftItemOutput) []*GiftItemResponse {
	if items == nil {
		return nil
	}
	responses := make([]*GiftItemResponse, len(items))
	for i, item := range items {
		responses[i] = h.toGiftItemResponse(item)
	}
	return responses
}

// CreateWishList godoc
//
//	@Summary		Create a new wish list
//	@Description	Create a new wish list for the authenticated user
//	@Tags			Wish Lists
//	@Accept			json
//	@Produce		json
//	@Param			wish_list	body		CreateWishListRequest	true	"Wish list creation information"
//	@Success		201			{object}	WishListResponse	"Wish list created successfully"
//	@Failure		400			{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists [post]
func (h *WishListHandler) CreateWishList(c echo.Context) error {
	var req CreateWishListRequest
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

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	wishList, err := h.service.CreateWishList(ctx, userID, services.CreateWishListInput{
		Title:        req.Title,
		Description:  req.Description,
		Occasion:     req.Occasion,
		OccasionDate: req.OccasionDate,
		TemplateID:   req.TemplateID,
		IsPublic:     req.IsPublic,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to create wish list: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusCreated, h.toWishListResponse(wishList))
}

// GetWishList godoc
//
//	@Summary		Get a wish list by ID
//	@Description	Get a wish list by its ID. If the wish list is private, the user must be the owner.
//	@Tags			Wish Lists
//	@Produce		json
//	@Param			id	path		string				true	"Wish List ID"
//	@Success		200	{object}	WishListResponse	"Wish list retrieved successfully"
//	@Failure		403	{object}	map[string]string	"Access denied"
//	@Failure		404	{object}	map[string]string	"Wish list not found"
//	@Security		BearerAuth
//	@Router			/wishlists/{id} [get]
func (h *WishListHandler) GetWishList(c echo.Context) error {
	wishListID := c.Param("id")

	ctx := c.Request().Context()
	wishList, err := h.service.GetWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Errorf("wish list not found: %w", err).Error(),
		})
	}

	// Get user from context to check ownership
	currentUserID, _, _, _ := auth.GetUserFromContext(c)

	// If not the owner and not public, return forbidden
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(http.StatusOK, h.toWishListResponse(wishList))
}

// GetWishListsByOwner godoc
//
//	@Summary		Get all wish lists owned by the authenticated user
//	@Description	Get all wish lists owned by the currently authenticated user
//	@Tags			Wish Lists
//	@Produce		json
//	@Success		200	{array}		WishListResponse	"List of wish lists retrieved successfully"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists [get]
func (h *WishListHandler) GetWishListsByOwner(c echo.Context) error {
	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	wishLists, err := h.service.GetWishListsByOwner(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get wish lists: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, h.toWishListResponses(wishLists))
}

// UpdateWishList godoc
//
//	@Summary		Update a wish list
//	@Description	Update a wish list by its ID. The user must be the owner of the wish list.
//	@Tags			Wish Lists
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Wish List ID"
//	@Param			wish_list	body		UpdateWishListRequest	true	"Wish list update information"
//	@Success		200			{object}	WishListResponse		"Wish list updated successfully"
//	@Failure		400			{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		403			{object}	map[string]string		"Forbidden"
//	@Failure		404			{object}	map[string]string		"Wish list not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{id} [put]
func (h *WishListHandler) UpdateWishList(c echo.Context) error {
	wishListID := c.Param("id")

	var req UpdateWishListRequest
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

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	wishList, err := h.service.UpdateWishList(ctx, wishListID, userID, services.UpdateWishListInput{
		Title:        req.Title,
		Description:  req.Description,
		Occasion:     req.Occasion,
		OccasionDate: req.OccasionDate,
		TemplateID:   req.TemplateID,
		IsPublic:     req.IsPublic,
	})

	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, services.ErrWishListNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "wish list not found",
			})
		}
		// Check if it's a forbidden error
		if errors.Is(err, services.ErrWishListForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to update wish list: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, h.toWishListResponse(wishList))
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
func (h *WishListHandler) DeleteWishList(c echo.Context) error {
	wishListID := c.Param("id")

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	err = h.service.DeleteWishList(ctx, wishListID, userID)
	if err != nil {
		// Check if it's a forbidden error
		if errors.Is(err, services.ErrWishListForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to delete wish list: %w", err).Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// CreateGiftItem godoc
//
//	@Summary		Create a new gift item
//	@Description	Create a new gift item in a wish list. The user must be the owner of the wish list.
//	@Tags			Gift Items
//	@Accept			json
//	@Produce		json
//	@Param			wishlistId	path		string					true	"Wish List ID"
//	@Param			gift_item	body		CreateGiftItemRequest	true	"Gift item creation information"
//	@Success		201			{object}	GiftItemResponse		"Gift item created successfully"
//	@Failure		400			{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		403			{object}	map[string]string		"Forbidden"
//	@Failure		404			{object}	map[string]string		"Wishlist not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{wishlistId}/gift-items [post]
func (h *WishListHandler) CreateGiftItem(c echo.Context) error {
	wishListID := c.Param("wishlistId")

	// Get user from context to verify ownership
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req CreateGiftItemRequest
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

	// Verify user owns the wishlist before creating gift item
	wishlist, err := h.service.GetWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}
	if wishlist.OwnerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to modify this wishlist",
		})
	}

	giftItem, err := h.service.CreateGiftItem(ctx, wishListID, services.CreateGiftItemInput{
		Name:        req.Name,
		Description: req.Description,
		Link:        req.Link,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Priority:    req.Priority,
		Notes:       req.Notes,
		Position:    req.Position,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to create gift item: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusCreated, h.toGiftItemResponse(giftItem))
}

// GetGiftItem godoc
//
//	@Summary		Get a gift item by ID
//	@Description	Get a gift item by its ID. If the parent wish list is private, the user must be the owner.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			id	path		string				true	"Gift Item ID"
//	@Success		200	{object}	GiftItemResponse	"Gift item retrieved successfully"
//	@Failure		403	{object}	map[string]string	"Access denied"
//	@Failure		404	{object}	map[string]string	"Gift item not found"
//	@Security		BearerAuth
//	@Router			/gift-items/{id} [get]
func (h *WishListHandler) GetGiftItem(c echo.Context) error {
	giftItemID := c.Param("id")

	ctx := c.Request().Context()
	giftItem, err := h.service.GetGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Errorf("gift item not found: %w", err).Error(),
		})
	}

	// Fetch parent wishlist to check access
	wishList, err := h.service.GetWishList(ctx, giftItem.WishlistID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}

	// Get user from context to check ownership
	currentUserID, _, _, _ := auth.GetUserFromContext(c)

	// Check if user has access to the wishlist
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(http.StatusOK, h.toGiftItemResponse(giftItem))
}

// GetGiftItemsByWishList godoc
//
//	@Summary		Get all gift items in a wish list
//	@Description	Get all gift items in a wish list. If the wish list is private, the user must be the owner.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			wishlistId	path		string					true	"Wish List ID"
//	@Param			page		query		int						false	"Page number (default 1)"
//	@Param			limit		query		int						false	"Items per page (default 10, max 100)"
//	@Success		200			{object}	GetGiftItemsResponse	"List of gift items retrieved successfully"
//	@Failure		403			{object}	map[string]string		"Access denied"
//	@Failure		404			{object}	map[string]string		"Wishlist not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/wishlists/{wishlistId}/gift-items [get]
func (h *WishListHandler) GetGiftItemsByWishList(c echo.Context) error {
	wishListID := c.Param("wishlistId")

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

	// Check access to the wishlist before fetching items
	wishList, err := h.service.GetWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}

	// Get user from context to check ownership
	currentUserID, _, _, _ := auth.GetUserFromContext(c)

	// Check if user has access to the wishlist
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	giftItems, err := h.service.GetGiftItemsByWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get gift items: %w", err).Error(),
		})
	}

	if giftItems == nil {
		giftItems = []*services.GiftItemOutput{}
	}

	// Apply pagination
	total := len(giftItems)
	start := (page - 1) * limit
	end := min(start+limit, total)

	// Handle out of bounds
	if start > total {
		start = total
	}

	pagedItems := giftItems[start:end]

	response := GetGiftItemsResponse{
		Items: h.toGiftItemResponses(pagedItems),
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: (total + limit - 1) / limit,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateGiftItem godoc
//
//	@Summary		Update a gift item
//	@Description	Update a gift item by its ID. The user must be the owner of the parent wish list.
//	@Tags			Gift Items
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Gift Item ID"
//	@Param			gift_item	body		UpdateGiftItemRequest	true	"Gift item update information"
//	@Success		200			{object}	GiftItemResponse		"Gift item updated successfully"
//	@Failure		400			{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string		"Unauthorized"
//	@Failure		403			{object}	map[string]string		"Forbidden"
//	@Failure		404			{object}	map[string]string		"Gift item or wishlist not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Security		BearerAuth
//	@Router			/gift-items/{id} [put]
func (h *WishListHandler) UpdateGiftItem(c echo.Context) error {
	giftItemID := c.Param("id")

	// Get user from context to verify ownership
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req UpdateGiftItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Get the gift item to find which wishlist it belongs to
	existingGiftItem, err := h.service.GetGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Gift item not found",
		})
	}

	// Verify user owns the wishlist before updating gift item
	wishlist, err := h.service.GetWishList(ctx, existingGiftItem.WishlistID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}
	if wishlist.OwnerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to modify this wishlist",
		})
	}

	giftItem, err := h.service.UpdateGiftItem(ctx, giftItemID, services.UpdateGiftItemInput{
		Name:        req.Name,
		Description: req.Description,
		Link:        req.Link,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Priority:    req.Priority,
		Notes:       req.Notes,
		Position:    req.Position,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to update gift item: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, h.toGiftItemResponse(giftItem))
}

// DeleteGiftItem godoc
//
//	@Summary		Delete a gift item
//	@Description	Delete a gift item by its ID. The user must be the owner of the parent wish list.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			id	path		string				true	"Gift Item ID"
//	@Success		204	{object}	nil					"Gift item deleted successfully"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		403	{object}	map[string]string	"Forbidden"
//	@Failure		404	{object}	map[string]string	"Gift item or wishlist not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/gift-items/{id} [delete]
func (h *WishListHandler) DeleteGiftItem(c echo.Context) error {
	giftItemID := c.Param("id")

	// Get user from context to verify ownership
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	// Get the gift item to find which wishlist it belongs to
	existingGiftItem, err := h.service.GetGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Gift item not found",
		})
	}

	// Verify user owns the wishlist before deleting gift item
	wishlist, err := h.service.GetWishList(ctx, existingGiftItem.WishlistID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}
	if wishlist.OwnerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to modify this wishlist",
		})
	}

	err = h.service.DeleteGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to delete gift item: %w", err).Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// MarkGiftItemAsPurchased godoc
//
//	@Summary		Mark a gift item as purchased
//	@Description	Mark a gift item as purchased with an optional purchased price. The user must be the owner of the parent wish list.
//	@Tags			Gift Items
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string				true	"Gift Item ID"
//	@Param			purchase_request	body		PurchaseRequest		true	"Purchase information"
//	@Success		200					{object}	GiftItemResponse	"Gift item marked as purchased successfully"
//	@Failure		400					{object}	map[string]string	"Invalid request body or validation error"
//	@Failure		401					{object}	map[string]string	"Unauthorized"
//	@Failure		403					{object}	map[string]string	"Forbidden"
//	@Failure		404					{object}	map[string]string	"Gift item or wishlist not found"
//	@Failure		500					{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/gift-items/{id}/purchase [post]
func (h *WishListHandler) MarkGiftItemAsPurchased(c echo.Context) error {
	giftItemID := c.Param("id")

	var req PurchaseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	if req.PurchasedPrice < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Purchased price must be >= 0",
		})
	}

	// Get user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()

	// Get the gift item to find which wishlist it belongs to
	existingGiftItem, err := h.service.GetGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Gift item not found",
		})
	}

	// Verify user owns the wishlist before marking as purchased
	wishlist, err := h.service.GetWishList(ctx, existingGiftItem.WishlistID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wishlist not found",
		})
	}
	if wishlist.OwnerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "You do not have permission to modify this wishlist",
		})
	}

	giftItem, err := h.service.MarkGiftItemAsPurchased(ctx, giftItemID, userID, req.PurchasedPrice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to mark gift item as purchased: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, h.toGiftItemResponse(giftItem))
}

// GetWishListByPublicSlug godoc
//
//	@Summary		Get a public wish list by its slug
//	@Description	Get a public wish list by its public slug. The wish list must be marked as public.
//	@Tags			Wish Lists
//	@Produce		json
//	@Param			slug	path		string				true	"Public Slug"
//	@Success		200		{object}	WishListResponse	"Public wish list retrieved successfully"
//	@Failure		404		{object}	map[string]string	"Wish list not found"
//	@Router			/public/wishlists/{slug} [get]
func (h *WishListHandler) GetWishListByPublicSlug(c echo.Context) error {
	publicSlug := c.Param("slug")

	ctx := c.Request().Context()
	wishList, err := h.service.GetWishListByPublicSlug(ctx, publicSlug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Errorf("wish list not found: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, h.toWishListResponse(wishList))
}

// GetGiftItemsByPublicSlug godoc
//
//	@Summary		Get gift items for a public wish list by slug
//	@Description	Get all gift items for a public wish list by its public slug with pagination support.
//	@Tags			Gift Items
//	@Produce		json
//	@Param			slug	path		string					true	"Public Slug"
//	@Param			page	query		int						false	"Page number (default 1)"
//	@Param			limit	query		int						false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	GetGiftItemsResponse	"Gift items retrieved successfully"
//	@Failure		404		{object}	map[string]string		"Wish list not found or not public"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/public/wishlists/{slug}/gift-items [get]
func (h *WishListHandler) GetGiftItemsByPublicSlug(c echo.Context) error {
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
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Wish list not found or not public",
		})
	}

	// Get all gift items for this wishlist
	giftItems, err := h.service.GetGiftItemsByWishList(ctx, wishList.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get gift items: %w", err).Error(),
		})
	}

	if giftItems == nil {
		giftItems = []*services.GiftItemOutput{}
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

	return c.JSON(http.StatusOK, GetGiftItemsResponse{
		Items: h.toGiftItemResponses(paginatedItems),
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	})
}
