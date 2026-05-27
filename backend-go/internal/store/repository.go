package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetMerchantStoreByUserID(ctx context.Context, userID int64) (Store, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.merchant_id, s.canteen_id, s.name, s.description, s.phone, s.address, s.status,
		       s.rating, s.sales_count, s.created_at, s.updated_at
		FROM stores s
		JOIN merchants m ON m.id = s.merchant_id
		WHERE m.user_id = ?
		LIMIT 1`, userID)
	return scanStore(row)
}

func (r *Repository) GetStoreByID(ctx context.Context, storeID int64) (Store, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, merchant_id, canteen_id, name, description, phone, address, status,
		       rating, sales_count, created_at, updated_at
		FROM stores
		WHERE id = ?`, storeID)
	return scanStore(row)
}

func (r *Repository) ListPublicStores(ctx context.Context) ([]Store, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, merchant_id, canteen_id, name, description, phone, address, status,
		       rating, sales_count, created_at, updated_at
		FROM stores
		WHERE status = 'OPEN'
		ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	defer rows.Close()
	return scanStores(rows)
}

func (r *Repository) ListStoresByCanteenID(ctx context.Context, canteenID int64) ([]Store, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, merchant_id, canteen_id, name, description, phone, address, status,
		       rating, sales_count, created_at, updated_at
		FROM stores
		WHERE canteen_id = ? AND status = 'OPEN'
		ORDER BY id DESC`, canteenID)
	if err != nil {
		return nil, fmt.Errorf("list stores by canteen: %w", err)
	}
	defer rows.Close()
	return scanStores(rows)
}

func (r *Repository) UpdateStoreByID(ctx context.Context, storeID int64, input StoreUpdateInput) (Store, error) {
	setParts := make([]string, 0, 5)
	args := make([]any, 0, 5)

	if input.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *input.Name)
	}
	if input.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *input.Description)
	}
	if input.Phone != nil {
		setParts = append(setParts, "phone = ?")
		args = append(args, *input.Phone)
	}
	if input.Address != nil {
		setParts = append(setParts, "address = ?")
		args = append(args, *input.Address)
	}
	if input.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *input.Status)
	}

	if len(setParts) == 0 {
		return r.GetStoreByID(ctx, storeID)
	}

	args = append(args, storeID)
	query := fmt.Sprintf(`UPDATE stores SET %s WHERE id = ?`, strings.Join(setParts, ", "))
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return Store{}, fmt.Errorf("update store: %w", err)
	}
	return r.GetStoreByID(ctx, storeID)
}

func (r *Repository) ListCategoriesByStoreID(ctx context.Context, storeID int64) ([]DishCategory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, store_id, name, sort_order, created_at, updated_at
		FROM dish_categories
		WHERE store_id = ?
		ORDER BY sort_order ASC, id ASC`, storeID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	return scanCategories(rows)
}

func (r *Repository) GetCategoryByID(ctx context.Context, storeID, categoryID int64) (DishCategory, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, store_id, name, sort_order, created_at, updated_at
		FROM dish_categories
		WHERE id = ? AND store_id = ?`, categoryID, storeID)
	return scanCategory(row)
}

func (r *Repository) CreateCategory(ctx context.Context, storeID int64, input CategoryCreateInput) (DishCategory, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO dish_categories (store_id, name, sort_order)
		VALUES (?, ?, ?)`, storeID, input.Name, input.SortOrder)
	if err != nil {
		return DishCategory{}, fmt.Errorf("create category: %w", err)
	}
	categoryID, err := result.LastInsertId()
	if err != nil {
		return DishCategory{}, fmt.Errorf("category id: %w", err)
	}
	return r.GetCategoryByID(ctx, storeID, categoryID)
}

func (r *Repository) UpdateCategory(ctx context.Context, storeID, categoryID int64, input CategoryUpdateInput) (DishCategory, error) {
	setParts := make([]string, 0, 2)
	args := make([]any, 0, 3)

	if input.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *input.Name)
	}
	if input.SortOrder != nil {
		setParts = append(setParts, "sort_order = ?")
		args = append(args, *input.SortOrder)
	}

	if len(setParts) == 0 {
		return r.GetCategoryByID(ctx, storeID, categoryID)
	}

	args = append(args, categoryID, storeID)
	query := fmt.Sprintf(`UPDATE dish_categories SET %s WHERE id = ? AND store_id = ?`, strings.Join(setParts, ", "))
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return DishCategory{}, fmt.Errorf("update category: %w", err)
	}
	return r.GetCategoryByID(ctx, storeID, categoryID)
}

func (r *Repository) DeleteCategory(ctx context.Context, storeID, categoryID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM dish_categories WHERE id = ? AND store_id = ?`, categoryID, storeID)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("category rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) CountDishesByCategory(ctx context.Context, storeID, categoryID int64) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM dishes
		WHERE store_id = ? AND category_id = ?`, storeID, categoryID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count dishes by category: %w", err)
	}
	return count, nil
}

func (r *Repository) ListDishes(ctx context.Context, storeID int64, filter DishQueryFilter, onlyOnSale bool) ([]Dish, error) {
	query := `
		SELECT id, store_id, category_id, name, description, price, stock, status, image_url, sales_count, created_at, updated_at
		FROM dishes
		WHERE store_id = ?`
	args := []any{storeID}

	if filter.CategoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *filter.CategoryID)
	}
	if onlyOnSale {
		query += " AND status = 'ON_SALE'"
	} else if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, *filter.Status)
	}
	if strings.TrimSpace(filter.Keyword) != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(filter.Keyword)+"%")
	}
	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dishes: %w", err)
	}
	defer rows.Close()
	return scanDishes(rows)
}

func (r *Repository) ListDishesByCanteenID(ctx context.Context, canteenID int64, filter DishQueryFilter, onlyOnSale bool) ([]Dish, error) {
	query := `
		SELECT d.id, d.store_id, d.category_id, d.name, d.description, d.price, d.stock, d.status, d.image_url, d.sales_count, d.created_at, d.updated_at
		FROM dishes d
		JOIN stores s ON s.id = d.store_id
		WHERE s.canteen_id = ?`
	args := []any{canteenID}

	if filter.CategoryID != nil {
		query += " AND d.category_id = ?"
		args = append(args, *filter.CategoryID)
	}
	if onlyOnSale {
		query += " AND d.status = 'ON_SALE'"
	} else if filter.Status != nil {
		query += " AND d.status = ?"
		args = append(args, *filter.Status)
	}
	if strings.TrimSpace(filter.Keyword) != "" {
		query += " AND d.name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(filter.Keyword)+"%")
	}
	query += " ORDER BY d.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dishes by canteen: %w", err)
	}
	defer rows.Close()
	return scanDishes(rows)
}

func (r *Repository) ListReviewsByStoreID(ctx context.Context, storeID int64) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.order_id, r.user_id, u.username, r.store_id, r.rating, r.content, r.created_at
		FROM reviews r
		JOIN users u ON u.id = r.user_id
		WHERE r.store_id = ?
		ORDER BY r.id DESC`, storeID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	items := make([]Review, 0)
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}
	return items, nil
}

func (r *Repository) GetDishByID(ctx context.Context, storeID, dishID int64) (Dish, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, store_id, category_id, name, description, price, stock, status, image_url, sales_count, created_at, updated_at
		FROM dishes
		WHERE id = ? AND store_id = ?`, dishID, storeID)
	return scanDish(row)
}

func (r *Repository) CreateDish(ctx context.Context, storeID int64, input DishCreateInput) (Dish, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO dishes (store_id, category_id, name, description, price, stock, status, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, storeID, input.CategoryID, input.Name, input.Description, input.Price, input.Stock, input.Status, nullIfEmpty(input.ImageURL))
	if err != nil {
		return Dish{}, fmt.Errorf("create dish: %w", err)
	}
	dishID, err := result.LastInsertId()
	if err != nil {
		return Dish{}, fmt.Errorf("dish id: %w", err)
	}
	return r.GetDishByID(ctx, storeID, dishID)
}

func (r *Repository) UpdateDish(ctx context.Context, storeID, dishID int64, input DishUpdateInput) (Dish, error) {
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 8)

	if input.CategoryID != nil {
		setParts = append(setParts, "category_id = ?")
		args = append(args, *input.CategoryID)
	}
	if input.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *input.Name)
	}
	if input.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *input.Description)
	}
	if input.Price != nil {
		setParts = append(setParts, "price = ?")
		args = append(args, *input.Price)
	}
	if input.Stock != nil {
		setParts = append(setParts, "stock = ?")
		args = append(args, *input.Stock)
	}
	if input.ImageURL != nil {
		setParts = append(setParts, "image_url = ?")
		args = append(args, nullIfEmpty(*input.ImageURL))
	}

	if len(setParts) == 0 {
		return r.GetDishByID(ctx, storeID, dishID)
	}

	args = append(args, dishID, storeID)
	query := fmt.Sprintf(`UPDATE dishes SET %s WHERE id = ? AND store_id = ?`, strings.Join(setParts, ", "))
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return Dish{}, fmt.Errorf("update dish: %w", err)
	}
	return r.GetDishByID(ctx, storeID, dishID)
}

func (r *Repository) UpdateDishStatus(ctx context.Context, storeID, dishID int64, status DishStatus) (Dish, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE dishes SET status = ? WHERE id = ? AND store_id = ?`, status, dishID, storeID); err != nil {
		return Dish{}, fmt.Errorf("update dish status: %w", err)
	}
	return r.GetDishByID(ctx, storeID, dishID)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStore(row rowScanner) (Store, error) {
	var item Store
	var description, phone, address sql.NullString
	if err := row.Scan(&item.ID, &item.MerchantID, &item.CanteenID, &item.Name, &description, &phone, &address, &item.Status, &item.Rating, &item.SalesCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Store{}, err
	}
	if description.Valid {
		item.Description = description.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if address.Valid {
		item.Address = address.String
	}
	return item, nil
}

func scanCategory(row rowScanner) (DishCategory, error) {
	var item DishCategory
	if err := row.Scan(&item.ID, &item.StoreID, &item.Name, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return DishCategory{}, err
	}
	return item, nil
}

func scanDish(row rowScanner) (Dish, error) {
	var item Dish
	var description, imageURL sql.NullString
	if err := row.Scan(&item.ID, &item.StoreID, &item.CategoryID, &item.Name, &description, &item.Price, &item.Stock, &item.Status, &imageURL, &item.SalesCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Dish{}, err
	}
	if description.Valid {
		item.Description = description.String
	}
	if imageURL.Valid {
		item.ImageURL = imageURL.String
	}
	return item, nil
}

func scanReview(row rowScanner) (Review, error) {
	var item Review
	var content sql.NullString
	if err := row.Scan(&item.ID, &item.OrderID, &item.UserID, &item.Username, &item.StoreID, &item.Rating, &content, &item.CreatedAt); err != nil {
		return Review{}, err
	}
	if content.Valid {
		item.Content = content.String
	}
	return item, nil
}

func scanStores(rows *sql.Rows) ([]Store, error) {
	items := make([]Store, 0)
	for rows.Next() {
		item, err := scanStore(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stores: %w", err)
	}
	return items, nil
}

func scanCategories(rows *sql.Rows) ([]DishCategory, error) {
	items := make([]DishCategory, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return items, nil
}

func scanDishes(rows *sql.Rows) ([]Dish, error) {
	items := make([]Dish, 0)
	for rows.Next() {
		item, err := scanDish(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dishes: %w", err)
	}
	return items, nil
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
