package order

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"project1/internal/store"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repository) getDishByID(ctx context.Context, q dbExecutor, dishID int64) (store.Dish, error) {
	return r.getDishByIDWithLock(ctx, q, dishID, false)
}

func (r *Repository) getDishByIDForUpdate(ctx context.Context, q dbExecutor, dishID int64) (store.Dish, error) {
	return r.getDishByIDWithLock(ctx, q, dishID, true)
}

func (r *Repository) getDishByIDWithLock(ctx context.Context, q dbExecutor, dishID int64, forUpdate bool) (store.Dish, error) {
	query := `
		SELECT id, store_id, category_id, name, description, price, stock, status, image_url, sales_count, created_at, updated_at
		FROM dishes
		WHERE id = ?`
	if forUpdate {
		query += " FOR UPDATE"
	}
	row := q.QueryRowContext(ctx, query, dishID)
	return scanDish(row)
}

func (r *Repository) getCartStoreIDByUserID(ctx context.Context, q dbExecutor, userID int64) (int64, error) {
	var storeID int64
	if err := q.QueryRowContext(ctx, `SELECT store_id FROM cart_items WHERE user_id = ? LIMIT 1`, userID).Scan(&storeID); err != nil {
		return 0, err
	}
	return storeID, nil
}

func (r *Repository) listCartItems(ctx context.Context, q dbExecutor, userID int64, forUpdate bool) ([]CartItem, error) {
	query := `
		SELECT ci.id, ci.user_id, ci.store_id, ci.dish_id, ci.quantity,
		       s.name, d.name, d.price, d.status, d.stock, ci.created_at, ci.updated_at
		FROM cart_items ci
		JOIN dishes d ON d.id = ci.dish_id
		JOIN stores s ON s.id = ci.store_id
		WHERE ci.user_id = ?
		ORDER BY ci.id ASC`
	if forUpdate {
		query += " FOR UPDATE"
	}
	rows, err := q.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list cart items: %w", err)
	}
	defer rows.Close()
	items := make([]CartItem, 0)
	for rows.Next() {
		item, err := scanCartItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart items: %w", err)
	}
	return items, nil
}

func (r *Repository) getCartItemByID(ctx context.Context, q dbExecutor, userID, itemID int64) (CartItem, error) {
	row := q.QueryRowContext(ctx, `
		SELECT ci.id, ci.user_id, ci.store_id, ci.dish_id, ci.quantity,
		       s.name, d.name, d.price, d.status, d.stock, ci.created_at, ci.updated_at
		FROM cart_items ci
		JOIN dishes d ON d.id = ci.dish_id
		JOIN stores s ON s.id = ci.store_id
		WHERE ci.id = ? AND ci.user_id = ?`, itemID, userID)
	return scanCartItem(row)
}

func (r *Repository) getCartItemByDishID(ctx context.Context, q dbExecutor, userID, dishID int64) (CartItem, error) {
	row := q.QueryRowContext(ctx, `
		SELECT ci.id, ci.user_id, ci.store_id, ci.dish_id, ci.quantity,
		       s.name, d.name, d.price, d.status, d.stock, ci.created_at, ci.updated_at
		FROM cart_items ci
		JOIN dishes d ON d.id = ci.dish_id
		JOIN stores s ON s.id = ci.store_id
		WHERE ci.user_id = ? AND ci.dish_id = ?`, userID, dishID)
	return scanCartItem(row)
}

func (r *Repository) createCartItem(ctx context.Context, q dbExecutor, userID, storeID, dishID int64, quantity int) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, store_id, dish_id, quantity)
		VALUES (?, ?, ?, ?)`, userID, storeID, dishID, quantity)
	return err
}

func (r *Repository) updateCartItemQuantity(ctx context.Context, q dbExecutor, userID, itemID int64, quantity int) error {
	_, err := q.ExecContext(ctx, `UPDATE cart_items SET quantity = ? WHERE id = ? AND user_id = ?`, quantity, itemID, userID)
	return err
}

func (r *Repository) deleteCartItem(ctx context.Context, q dbExecutor, userID, itemID int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM cart_items WHERE id = ? AND user_id = ?`, itemID, userID)
	return err
}

func (r *Repository) clearCart(ctx context.Context, q dbExecutor, userID int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = ?`, userID)
	return err
}

func (r *Repository) lockCartItemsForCheckout(ctx context.Context, q dbExecutor, userID int64) ([]CartItem, error) {
	return r.listCartItems(ctx, q, userID, true)
}

func (r *Repository) lockDishesByIDs(ctx context.Context, q dbExecutor, dishIDs []int64) ([]store.Dish, error) {
	if len(dishIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(dishIDs))
	args := make([]any, 0, len(dishIDs))
	for _, id := range dishIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	query := fmt.Sprintf(`
		SELECT id, store_id, category_id, name, description, price, stock, status, image_url, sales_count, created_at, updated_at
		FROM dishes
		WHERE id IN (%s)
		FOR UPDATE`, strings.Join(placeholders, ","))
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lock dishes: %w", err)
	}
	defer rows.Close()
	items := make([]store.Dish, 0, len(dishIDs))
	for rows.Next() {
		item, err := scanStoreDish(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locked dishes: %w", err)
	}
	return items, nil
}

func (r *Repository) createOrder(ctx context.Context, q dbExecutor, input orderCreateRecord) (int64, error) {
	result, err := q.ExecContext(ctx, `
		INSERT INTO orders (order_no, user_id, store_id, total_amount, status, remark)
		VALUES (?, ?, ?, ?, ?, ?)`, input.OrderNo, input.UserID, input.StoreID, input.TotalAmount, input.Status, nullString(input.Remark))
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	return result.LastInsertId()
}

func (r *Repository) createOrderItems(ctx context.Context, q dbExecutor, items []orderItemRecord) error {
	if len(items) == 0 {
		return nil
	}
	valueParts := make([]string, 0, len(items))
	args := make([]any, 0, len(items)*6)
	for _, item := range items {
		valueParts = append(valueParts, "(?, ?, ?, ?, ?, ?)")
		args = append(args, item.OrderID, item.DishID, item.DishName, item.DishPrice, item.Quantity, item.Subtotal)
	}
	query := fmt.Sprintf(`INSERT INTO order_items (order_id, dish_id, dish_name, dish_price, quantity, subtotal) VALUES %s`, strings.Join(valueParts, ","))
	_, err := q.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) decrementDishStock(ctx context.Context, q dbExecutor, dishID int64, quantity int) error {
	_, err := q.ExecContext(ctx, `UPDATE dishes SET stock = stock - ? WHERE id = ?`, quantity, dishID)
	return err
}

func (r *Repository) restoreDishStock(ctx context.Context, q dbExecutor, dishID int64, quantity int) error {
	_, err := q.ExecContext(ctx, `UPDATE dishes SET stock = stock + ? WHERE id = ?`, quantity, dishID)
	return err
}

func (r *Repository) listOrders(ctx context.Context, q dbExecutor, where string, args ...any) ([]Order, error) {
	query := `
		SELECT o.id, o.order_no, o.user_id, o.store_id, s.name, u.username, o.total_amount, o.status, o.remark, o.created_at, o.updated_at
		FROM orders o
		JOIN stores s ON s.id = o.store_id
		JOIN users u ON u.id = o.user_id`
	if strings.TrimSpace(where) != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY o.id DESC"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()
	items := make([]Order, 0)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return items, nil
}

func (r *Repository) getOrderByID(ctx context.Context, q dbExecutor, where string, args ...any) (Order, error) {
	query := `
		SELECT o.id, o.order_no, o.user_id, o.store_id, s.name, u.username, o.total_amount, o.status, o.remark, o.created_at, o.updated_at
		FROM orders o
		JOIN stores s ON s.id = o.store_id
		JOIN users u ON u.id = o.user_id`
	if strings.TrimSpace(where) != "" {
		query += " WHERE " + where
	}
	row := q.QueryRowContext(ctx, query, args...)
	return scanOrder(row)
}

func (r *Repository) listOrderItems(ctx context.Context, q dbExecutor, orderID int64) ([]OrderItem, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, order_id, dish_id, dish_name, dish_price, quantity, subtotal, created_at
		FROM order_items
		WHERE order_id = ?
		ORDER BY id ASC`, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()
	items := make([]OrderItem, 0)
	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}
	return items, nil
}

func (r *Repository) updateOrderStatus(ctx context.Context, q dbExecutor, orderID int64, status OrderStatus) error {
	_, err := q.ExecContext(ctx, `UPDATE orders SET status = ? WHERE id = ?`, status, orderID)
	return err
}

func (r *Repository) getReviewByOrderID(ctx context.Context, q dbExecutor, orderID int64) (Review, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, order_id, user_id, store_id, rating, content, created_at
		FROM reviews
		WHERE order_id = ?`, orderID)
	return scanReview(row)
}

func (r *Repository) createReview(ctx context.Context, q dbExecutor, input Review) (Review, error) {
	result, err := q.ExecContext(ctx, `
		INSERT INTO reviews (order_id, user_id, store_id, rating, content)
		VALUES (?, ?, ?, ?, ?)`, input.OrderID, input.UserID, input.StoreID, input.Rating, nullString(input.Content))
	if err != nil {
		return Review{}, fmt.Errorf("create review: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Review{}, fmt.Errorf("review id: %w", err)
	}
	item, err := r.getReviewByOrderID(ctx, q, input.OrderID)
	if err == nil {
		item.ID = id
		return item, nil
	}
	return Review{ID: id, OrderID: input.OrderID, UserID: input.UserID, StoreID: input.StoreID, Rating: input.Rating, Content: input.Content}, nil
}

func scanCartItem(row rowScanner) (CartItem, error) {
	var item CartItem
	var dishStatus sql.NullString
	if err := row.Scan(&item.ID, &item.UserID, &item.StoreID, &item.DishID, &item.Quantity, &item.StoreName, &item.DishName, &item.DishPrice, &dishStatus, &item.DishStock, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return CartItem{}, err
	}
	if dishStatus.Valid {
		item.DishStatus = dishStatus.String
	}
	return item, nil
}

func scanDish(row rowScanner) (store.Dish, error) {
	var item store.Dish
	var description, imageURL sql.NullString
	if err := row.Scan(&item.ID, &item.StoreID, &item.CategoryID, &item.Name, &description, &item.Price, &item.Stock, &item.Status, &imageURL, &item.SalesCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return store.Dish{}, err
	}
	if description.Valid {
		item.Description = description.String
	}
	if imageURL.Valid {
		item.ImageURL = imageURL.String
	}
	return item, nil
}

func scanStoreDish(row rowScanner) (store.Dish, error) {
	return scanDish(row)
}

func scanOrder(row rowScanner) (Order, error) {
	var item Order
	var remark sql.NullString
	if err := row.Scan(&item.ID, &item.OrderNo, &item.UserID, &item.StoreID, &item.StoreName, &item.UserName, &item.TotalAmount, &item.Status, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Order{}, err
	}
	if remark.Valid {
		item.Remark = remark.String
	}
	return item, nil
}

func scanOrderItem(row rowScanner) (OrderItem, error) {
	var item OrderItem
	if err := row.Scan(&item.ID, &item.OrderID, &item.DishID, &item.DishName, &item.DishPrice, &item.Quantity, &item.Subtotal, &item.CreatedAt); err != nil {
		return OrderItem{}, err
	}
	return item, nil
}

func scanReview(row rowScanner) (Review, error) {
	var item Review
	var content sql.NullString
	if err := row.Scan(&item.ID, &item.OrderID, &item.UserID, &item.StoreID, &item.Rating, &content, &item.CreatedAt); err != nil {
		return Review{}, err
	}
	if content.Valid {
		item.Content = content.String
	}
	return item, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

type orderCreateRecord struct {
	OrderNo     string
	UserID      int64
	StoreID     int64
	TotalAmount float64
	Status      OrderStatus
	Remark      string
}

type orderItemRecord struct {
	OrderID   int64
	DishID    int64
	DishName  string
	DishPrice float64
	Quantity  int
	Subtotal  float64
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
