package handlers

import (
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
	Name        string  `json:"name" validate:"max=255"`
	Description string  `json:"description"`
	Link        string  `json:"link" validate:"omitempty,url"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
	Price       float64 `json:"price" validate:"omitempty,min=0"`
	Priority    int     `json:"priority" validate:"omitempty,min=0,max=10"`
	Notes       string  `json:"notes"`
	Position    int     `json:"position" validate:"omitempty,min=0"`
}

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

	return c.JSON(http.StatusCreated, wishList)
}

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

	// If not the owner and not public, return unauthorized
	isOwner := currentUserID == wishList.OwnerID
	if !isOwner && !wishList.IsPublic {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(http.StatusOK, wishList)
}

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

	return c.JSON(http.StatusOK, wishLists)
}

func (h *WishListHandler) UpdateWishList(c echo.Context) error {
	wishListID := c.Param("id")

	var req UpdateWishListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to update wish list: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, wishList)
}

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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to delete wish list: %w", err).Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

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

	return c.JSON(http.StatusCreated, giftItem)
}

func (h *WishListHandler) GetGiftItem(c echo.Context) error {
	giftItemID := c.Param("id")

	ctx := c.Request().Context()
	giftItem, err := h.service.GetGiftItem(ctx, giftItemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Errorf("gift item not found: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, giftItem)
}

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
	giftItems, err := h.service.GetGiftItemsByWishList(ctx, wishListID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get gift items: %w", err).Error(),
		})
	}

	if giftItems == nil {
		giftItems = []*services.GiftItemOutput{}
	}

	type GetGiftItemsResponse struct {
		Items []*services.GiftItemOutput `json:"items"`
		Total int                        `json:"total"`
		Page  int                        `json:"page"`
		Limit int                        `json:"limit"`
		Pages int                        `json:"pages"`
	}

	response := GetGiftItemsResponse{
		Items: giftItems,
		Total: len(giftItems),
		Page:  page,
		Limit: limit,
		Pages: (len(giftItems) + limit - 1) / limit,
	}

	return c.JSON(http.StatusOK, response)
}

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

	return c.JSON(http.StatusOK, giftItem)
}

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

// MarkGiftItemAsPurchased marks a gift item as purchased
func (h *WishListHandler) MarkGiftItemAsPurchased(c echo.Context) error {
	giftItemID := c.Param("id")

	var req struct {
		PurchasedPrice float64 `json:"purchased_price"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
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
	giftItem, err := h.service.MarkGiftItemAsPurchased(ctx, giftItemID, userID, req.PurchasedPrice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to mark gift item as purchased: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, giftItem)
}

func (h *WishListHandler) GetWishListByPublicSlug(c echo.Context) error {
	publicSlug := c.Param("slug")

	ctx := c.Request().Context()
	wishList, err := h.service.GetWishListByPublicSlug(ctx, publicSlug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": fmt.Errorf("wish list not found: %w", err).Error(),
		})
	}

	return c.JSON(http.StatusOK, wishList)
}
