package ai

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListRecommendationStats(ctx context.Context) ([]recommendationStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.name, d.store_id, s.name, d.price, d.sales_count,
		       COALESCE(stats.order_count, 0), COALESCE(stats.total_quantity, 0), COALESCE(stats.total_amount, 0),
		       COALESCE(s.rating, 0)
		FROM dishes d
		JOIN stores s ON s.id = d.store_id
		LEFT JOIN (
			SELECT oi.dish_id,
			       COUNT(1) AS order_count,
			       SUM(oi.quantity) AS total_quantity,
			       SUM(oi.subtotal) AS total_amount
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.status <> 'CANCELLED'
			GROUP BY oi.dish_id
		) stats ON stats.dish_id = d.id
		WHERE d.status = 'ON_SALE'
		ORDER BY d.sales_count DESC, d.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list recommendation stats: %w", err)
	}
	defer rows.Close()
	items := make([]recommendationStat, 0)
	for rows.Next() {
		var item recommendationStat
		if err := rows.Scan(&item.DishID, &item.DishName, &item.StoreID, &item.StoreName, &item.Price, &item.SalesCount, &item.OrderCount, &item.TotalQuantity, &item.TotalAmount, &item.StoreRating); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recommendation stats: %w", err)
	}
	return items, nil
}
