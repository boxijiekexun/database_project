package order

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"project1/internal/store"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrCartEmpty            = errors.New("cart is empty")
	ErrCartCrossStore       = errors.New("cart items must belong to the same store")
	ErrCartItemNotFound     = errors.New("cart item not found")
	ErrDishNotOnSale        = errors.New("dish is not on sale")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrInvalidQuantity      = errors.New("invalid quantity")
	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderNotCancelable   = errors.New("order cannot be cancelled")
	ErrOrderAlreadyReviewed = errors.New("order already reviewed")
	ErrInvalidRating        = errors.New("rating must be between 1 and 5")
	ErrInvalidTransition    = errors.New("invalid order status transition")
	ErrConflict             = errors.New("conflict")
)

type Service struct {
	repo     *Repository
	storeSvc *store.Service
}

func NewService(repo *Repository, storeSvc *store.Service) *Service {
	return &Service{repo: repo, storeSvc: storeSvc}
}

func (s *Service) ListCart(ctx context.Context, userID int64) ([]CartItem, error) {
	return s.repo.listCartItems(ctx, s.repo.db, userID, false)
}

func (s *Service) AddCartItem(ctx context.Context, userID int64, input CartAddInput) (CartItem, error) {
	if input.Quantity <= 0 {
		return CartItem{}, ErrInvalidQuantity
	}
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return CartItem{}, err
	}
	defer func() { _ = tx.Rollback() }()

	dish, err := s.repo.getDishByIDForUpdate(ctx, tx, input.DishID)
	if err != nil {
		return CartItem{}, normalizeCartLookupError(err)
	}
	if dish.Status != store.DishStatusOnSale {
		return CartItem{}, ErrDishNotOnSale
	}
	if dish.Stock < input.Quantity {
		return CartItem{}, ErrInsufficientStock
	}
	if storeID, err := s.repo.getCartStoreIDByUserID(ctx, tx, userID); err == nil {
		if storeID != dish.StoreID {
			return CartItem{}, ErrCartCrossStore
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return CartItem{}, err
	}

	existing, err := s.repo.getCartItemByDishID(ctx, tx, userID, input.DishID)
	if err == nil {
		if existing.Quantity+input.Quantity > dish.Stock {
			return CartItem{}, ErrInsufficientStock
		}
		if err := s.repo.updateCartItemQuantity(ctx, tx, userID, existing.ID, existing.Quantity+input.Quantity); err != nil {
			return CartItem{}, normalizeDBError(err)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		if err := s.repo.createCartItem(ctx, tx, userID, dish.StoreID, dish.ID, input.Quantity); err != nil {
			return CartItem{}, normalizeDBError(err)
		}
	} else {
		return CartItem{}, err
	}

	updated, err := s.repo.getCartItemByDishID(ctx, tx, userID, input.DishID)
	if err != nil {
		return CartItem{}, normalizeCartLookupError(err)
	}
	if err := tx.Commit(); err != nil {
		return CartItem{}, err
	}
	return updated, nil
}

func (s *Service) UpdateCartItem(ctx context.Context, userID, itemID int64, quantity int) (CartItem, error) {
	if quantity <= 0 {
		return CartItem{}, ErrInvalidQuantity
	}
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return CartItem{}, err
	}
	defer func() { _ = tx.Rollback() }()

	item, err := s.repo.getCartItemByID(ctx, tx, userID, itemID)
	if err != nil {
		return CartItem{}, normalizeCartLookupError(err)
	}
	dish, err := s.repo.getDishByIDForUpdate(ctx, tx, item.DishID)
	if err != nil {
		return CartItem{}, normalizeCartLookupError(err)
	}
	if dish.Status != store.DishStatusOnSale {
		return CartItem{}, ErrDishNotOnSale
	}
	if quantity > dish.Stock {
		return CartItem{}, ErrInsufficientStock
	}
	if err := s.repo.updateCartItemQuantity(ctx, tx, userID, itemID, quantity); err != nil {
		return CartItem{}, normalizeDBError(err)
	}
	updated, err := s.repo.getCartItemByID(ctx, tx, userID, itemID)
	if err != nil {
		return CartItem{}, normalizeCartLookupError(err)
	}
	if err := tx.Commit(); err != nil {
		return CartItem{}, err
	}
	return updated, nil
}

func (s *Service) DeleteCartItem(ctx context.Context, userID, itemID int64) error {
	if err := s.repo.deleteCartItem(ctx, s.repo.db, userID, itemID); err != nil {
		return normalizeCartLookupError(err)
	}
	return nil
}

func (s *Service) ClearCart(ctx context.Context, userID int64) error {
	return s.repo.clearCart(ctx, s.repo.db, userID)
}

func (s *Service) CreateOrder(ctx context.Context, userID int64, input CreateOrderInput) (Order, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	cartItems, err := s.repo.lockCartItemsForCheckout(ctx, tx, userID)
	if err != nil {
		return Order{}, err
	}
	if len(cartItems) == 0 {
		return Order{}, ErrCartEmpty
	}

	storeID := cartItems[0].StoreID
	for _, item := range cartItems {
		if item.StoreID != storeID {
			return Order{}, ErrCartCrossStore
		}
	}

	dishIDs := make([]int64, 0, len(cartItems))
	for _, item := range cartItems {
		dishIDs = append(dishIDs, item.DishID)
	}

	lockedDishes, err := s.repo.lockDishesByIDs(ctx, tx, dishIDs)
	if err != nil {
		return Order{}, err
	}
	dishMap := make(map[int64]store.Dish, len(lockedDishes))
	for _, dish := range lockedDishes {
		dishMap[dish.ID] = dish
	}

	total := 0.0
	orderItems := make([]orderItemRecord, 0, len(cartItems))
	for _, item := range cartItems {
		dish, ok := dishMap[item.DishID]
		if !ok {
			return Order{}, ErrOrderNotFound
		}
		if dish.Status != store.DishStatusOnSale {
			return Order{}, ErrDishNotOnSale
		}
		if dish.Stock < item.Quantity {
			return Order{}, ErrInsufficientStock
		}
		subtotal := dish.Price * float64(item.Quantity)
		total += subtotal
		orderItems = append(orderItems, orderItemRecord{
			DishID:    dish.ID,
			DishName:  dish.Name,
			DishPrice: dish.Price,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
		})
	}

	orderNo := generateOrderNo(userID)
	orderID, err := s.repo.createOrder(ctx, tx, orderCreateRecord{
		OrderNo:     orderNo,
		UserID:      userID,
		StoreID:     storeID,
		TotalAmount: total,
		Status:      OrderStatusPendingPayment,
		Remark:      input.Remark,
	})
	if err != nil {
		return Order{}, normalizeDBError(err)
	}

	for i := range orderItems {
		orderItems[i].OrderID = orderID
		if err := s.repo.decrementDishStock(ctx, tx, orderItems[i].DishID, orderItems[i].Quantity); err != nil {
			return Order{}, err
		}
	}
	if err := s.repo.createOrderItems(ctx, tx, orderItems); err != nil {
		return Order{}, normalizeDBError(err)
	}
	if err := s.repo.clearCart(ctx, tx, userID); err != nil {
		return Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return Order{}, err
	}

	created, err := s.repo.getOrderByID(ctx, s.repo.db, `o.id = ? AND o.user_id = ?`, orderID, userID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	createdItems, err := s.repo.listOrderItems(ctx, s.repo.db, orderID)
	if err != nil {
		return Order{}, err
	}
	created.Items = createdItems
	return created, nil
}

func (s *Service) ListUserOrders(ctx context.Context, userID int64) ([]Order, error) {
	return s.repo.listOrders(ctx, s.repo.db, `o.user_id = ?`, userID)
}

func (s *Service) GetUserOrder(ctx context.Context, userID, orderID int64) (Order, error) {
	item, err := s.repo.getOrderByID(ctx, s.repo.db, `o.id = ? AND o.user_id = ?`, orderID, userID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	items, err := s.repo.listOrderItems(ctx, s.repo.db, orderID)
	if err != nil {
		return Order{}, err
	}
	item.Items = items
	return item, nil
}

func (s *Service) CancelOrder(ctx context.Context, userID, orderID int64) (Order, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	item, err := s.repo.getOrderByID(ctx, tx, `o.id = ? AND o.user_id = ? FOR UPDATE`, orderID, userID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	if item.Status != OrderStatusPendingPayment {
		return Order{}, ErrOrderNotCancelable
	}
	items, err := s.repo.listOrderItems(ctx, tx, orderID)
	if err != nil {
		return Order{}, err
	}
	if err := s.repo.updateOrderStatus(ctx, tx, orderID, OrderStatusCancelled); err != nil {
		return Order{}, err
	}
	for _, orderItem := range items {
		if err := s.repo.restoreDishStock(ctx, tx, orderItem.DishID, orderItem.Quantity); err != nil {
			return Order{}, err
		}
	}
	item.Status = OrderStatusCancelled
	item.Items = items
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return item, nil
}

func (s *Service) PayOrder(ctx context.Context, userID, orderID int64) (Order, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	item, err := s.repo.getOrderByID(ctx, tx, `o.id = ? AND o.user_id = ? FOR UPDATE`, orderID, userID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	if item.Status != OrderStatusPendingPayment {
		return Order{}, ErrInvalidTransition
	}
	if err := s.repo.updateOrderStatus(ctx, tx, orderID, OrderStatusPaid); err != nil {
		return Order{}, err
	}
	item.Status = OrderStatusPaid
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return item, nil
}

func (s *Service) ReviewOrder(ctx context.Context, userID, orderID int64, input ReviewInput) (Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return Review{}, ErrInvalidRating
	}
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Review{}, err
	}
	defer func() { _ = tx.Rollback() }()

	item, err := s.repo.getOrderByID(ctx, tx, `o.id = ? AND o.user_id = ? FOR UPDATE`, orderID, userID)
	if err != nil {
		return Review{}, normalizeLookupError(err)
	}
	if item.Status != OrderStatusCompleted {
		return Review{}, ErrInvalidTransition
	}
	if _, err := s.repo.getReviewByOrderID(ctx, tx, orderID); err == nil {
		return Review{}, ErrOrderAlreadyReviewed
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Review{}, err
	}
	review, err := s.repo.createReview(ctx, tx, Review{
		OrderID: orderID,
		UserID:  userID,
		StoreID: item.StoreID,
		Rating:  input.Rating,
		Content: strings.TrimSpace(input.Content),
	})
	if err != nil {
		return Review{}, normalizeDBError(err)
	}
	if err := tx.Commit(); err != nil {
		return Review{}, err
	}
	return review, nil
}

func (s *Service) ListMerchantOrders(ctx context.Context, merchantUserID int64, filter OrderQueryFilter) ([]Order, error) {
	storeItem, err := s.storeSvc.GetMerchantStore(ctx, merchantUserID)
	if err != nil {
		return nil, err
	}
	where := "o.store_id = ?"
	args := []any{storeItem.ID}
	if filter.Status != nil {
		where += " AND o.status = ?"
		args = append(args, *filter.Status)
	}
	return s.repo.listOrders(ctx, s.repo.db, where, args...)
}

func (s *Service) GetMerchantOrder(ctx context.Context, merchantUserID, orderID int64) (Order, error) {
	storeItem, err := s.storeSvc.GetMerchantStore(ctx, merchantUserID)
	if err != nil {
		return Order{}, err
	}
	item, err := s.repo.getOrderByID(ctx, s.repo.db, `o.id = ? AND o.store_id = ?`, orderID, storeItem.ID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	items, err := s.repo.listOrderItems(ctx, s.repo.db, orderID)
	if err != nil {
		return Order{}, err
	}
	item.Items = items
	return item, nil
}

func (s *Service) UpdateMerchantOrderStatus(ctx context.Context, merchantUserID, orderID int64, next OrderStatus) (Order, error) {
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	storeItem, err := s.storeSvc.GetMerchantStore(ctx, merchantUserID)
	if err != nil {
		return Order{}, err
	}
	item, err := s.repo.getOrderByID(ctx, tx, `o.id = ? AND o.store_id = ? FOR UPDATE`, orderID, storeItem.ID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	if !item.Status.CanMerchantTransitionTo(next) {
		return Order{}, ErrInvalidTransition
	}
	if err := s.repo.updateOrderStatus(ctx, tx, orderID, next); err != nil {
		return Order{}, err
	}
	item.Status = next
	items, err := s.repo.listOrderItems(ctx, tx, orderID)
	if err == nil {
		item.Items = items
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return item, nil
}

func (s *Service) ListAdminOrders(ctx context.Context, filter OrderQueryFilter) ([]Order, error) {
	whereParts := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if filter.Status != nil {
		whereParts = append(whereParts, "o.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.StoreID != nil {
		whereParts = append(whereParts, "o.store_id = ?")
		args = append(args, *filter.StoreID)
	}
	if filter.UserID != nil {
		whereParts = append(whereParts, "o.user_id = ?")
		args = append(args, *filter.UserID)
	}
	return s.repo.listOrders(ctx, s.repo.db, strings.Join(whereParts, " AND "), args...)
}

func (s *Service) GetAdminOrder(ctx context.Context, orderID int64) (Order, error) {
	item, err := s.repo.getOrderByID(ctx, s.repo.db, `o.id = ?`, orderID)
	if err != nil {
		return Order{}, normalizeLookupError(err)
	}
	items, err := s.repo.listOrderItems(ctx, s.repo.db, orderID)
	if err != nil {
		return Order{}, err
	}
	item.Items = items
	return item, nil
}

func generateOrderNo(userID int64) string {
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprintf("ORD%d%d", time.Now().UnixNano(), userID)
	}
	return fmt.Sprintf("ORD%d%d%s", time.Now().UnixNano(), userID, hex.EncodeToString(raw))
}

func normalizeLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrOrderNotFound
	}
	return err
}

func normalizeCartLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCartItemNotFound
	}
	return err
}

func normalizeDBError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrConflict
		case 1452:
			return ErrOrderNotFound
		}
	}
	return err
}
