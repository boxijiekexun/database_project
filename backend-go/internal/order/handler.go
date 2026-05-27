package order

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"project1/internal/middleware"
	"project1/internal/response"
	"project1/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListCartItems(c *gin.Context) {
	items, err := h.service.ListCart(c.Request.Context(), currentUserID(c))
	if err != nil {
		handleOrderError(c, err)
		return
	}
	populateCartSubtotals(items)
	response.Success(c, items)
}

type cartAddRequest struct {
	DishID   int64 `json:"dish_id" binding:"required"`
	Quantity int   `json:"quantity" binding:"required"`
}

func (h *Handler) AddCartItem(c *gin.Context) {
	var req cartAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.service.AddCartItem(c.Request.Context(), currentUserID(c), CartAddInput{DishID: req.DishID, Quantity: req.Quantity})
	if err != nil {
		handleOrderError(c, err)
		return
	}
	item.Subtotal = item.DishPrice * float64(item.Quantity)
	response.Success(c, item)
}

type cartUpdateRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

func (h *Handler) UpdateCartItem(c *gin.Context) {
	itemID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	var req cartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.service.UpdateCartItem(c.Request.Context(), currentUserID(c), itemID, req.Quantity)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	item.Subtotal = item.DishPrice * float64(item.Quantity)
	response.Success(c, item)
}

func (h *Handler) DeleteCartItem(c *gin.Context) {
	itemID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	if err := h.service.DeleteCartItem(c.Request.Context(), currentUserID(c), itemID); err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ClearCart(c *gin.Context) {
	if err := h.service.ClearCart(c.Request.Context(), currentUserID(c)); err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, gin.H{})
}

type createOrderRequest struct {
	Remark string `json:"remark"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.service.CreateOrder(c.Request.Context(), currentUserID(c), CreateOrderInput{Remark: strings.TrimSpace(req.Remark)})
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) ListUserOrders(c *gin.Context) {
	items, err := h.service.ListUserOrders(c.Request.Context(), currentUserID(c))
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) GetUserOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.GetUserOrder(c.Request.Context(), currentUserID(c), orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) CancelUserOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.CancelOrder(c.Request.Context(), currentUserID(c), orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) PayUserOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.PayOrder(c.Request.Context(), currentUserID(c), orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

type reviewRequest struct {
	Rating  int    `json:"rating" binding:"required"`
	Content string `json:"content"`
}

func (h *Handler) ReviewUserOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	review, err := h.service.ReviewOrder(c.Request.Context(), currentUserID(c), orderID, ReviewInput{Rating: req.Rating, Content: strings.TrimSpace(req.Content)})
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, review)
}

func (h *Handler) ListMerchantOrders(c *gin.Context) {
	filter := OrderQueryFilter{}
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status := OrderStatus(raw)
		if !status.IsValid() {
			response.Error(c, http.StatusBadRequest, "invalid order status")
			return
		}
		filter.Status = &status
	}
	items, err := h.service.ListMerchantOrders(c.Request.Context(), currentUserID(c), filter)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) GetMerchantOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.GetMerchantOrder(c.Request.Context(), currentUserID(c), orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

type orderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

func (h *Handler) UpdateMerchantOrderStatus(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	var req orderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !req.Status.IsValid() {
		response.Error(c, http.StatusBadRequest, "invalid order status")
		return
	}
	item, err := h.service.UpdateMerchantOrderStatus(c.Request.Context(), currentUserID(c), orderID, req.Status)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

func (h *Handler) ListAdminOrders(c *gin.Context) {
	filter := OrderQueryFilter{}
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		status := OrderStatus(raw)
		if !status.IsValid() {
			response.Error(c, http.StatusBadRequest, "invalid order status")
			return
		}
		filter.Status = &status
	}
	if raw := strings.TrimSpace(c.Query("store_id")); raw != "" {
		storeID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid store_id")
			return
		}
		filter.StoreID = &storeID
	}
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		userID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid user_id")
			return
		}
		filter.UserID = &userID
	}
	items, err := h.service.ListAdminOrders(c.Request.Context(), filter)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) GetAdminOrder(c *gin.Context) {
	orderID, ok := parseOrderIDParam(c)
	if !ok {
		return
	}
	item, err := h.service.GetAdminOrder(c.Request.Context(), orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.Success(c, item)
}

func parseOrderIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func currentUserID(c *gin.Context) int64 {
	claims, _ := middleware.ClaimsFromContext(c)
	return claims.UserID
}

func handleOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		response.Error(c, http.StatusNotFound, "store not found")
	case errors.Is(err, ErrCartItemNotFound):
		response.Error(c, http.StatusNotFound, "cart item not found")
	case errors.Is(err, ErrOrderNotFound):
		response.Error(c, http.StatusNotFound, "order not found")
	case errors.Is(err, ErrCartEmpty):
		response.Error(c, http.StatusBadRequest, "cart is empty")
	case errors.Is(err, ErrCartCrossStore):
		response.Error(c, http.StatusBadRequest, "cart items must belong to the same store")
	case errors.Is(err, ErrDishNotOnSale):
		response.Error(c, http.StatusBadRequest, "dish is not on sale")
	case errors.Is(err, ErrInsufficientStock):
		response.Error(c, http.StatusBadRequest, "insufficient stock")
	case errors.Is(err, ErrInvalidQuantity):
		response.Error(c, http.StatusBadRequest, "invalid quantity")
	case errors.Is(err, ErrOrderNotCancelable):
		response.Error(c, http.StatusBadRequest, "only pending payment orders can be cancelled")
	case errors.Is(err, ErrOrderAlreadyReviewed):
		response.Error(c, http.StatusBadRequest, "order already reviewed")
	case errors.Is(err, ErrInvalidRating):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInvalidTransition):
		response.Error(c, http.StatusBadRequest, "invalid order status transition")
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, "conflict")
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func populateCartSubtotals(items []CartItem) {
	for i := range items {
		items[i].Subtotal = items[i].DishPrice * float64(items[i].Quantity)
	}
}
