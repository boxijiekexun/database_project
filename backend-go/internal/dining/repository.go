package dining

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

func (r *Repository) ListAreas(ctx context.Context) ([]DiningArea, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM dining_areas
		ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list dining areas: %w", err)
	}
	defer rows.Close()
	items := make([]DiningArea, 0)
	for rows.Next() {
		var item DiningArea
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dining areas: %w", err)
	}
	return items, nil
}

func (r *Repository) GetAreaByID(ctx context.Context, id int64) (DiningArea, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM dining_areas
		WHERE id = ?`, id)
	var item DiningArea
	if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return DiningArea{}, err
	}
	return item, nil
}

func (r *Repository) ListCanteensByAreaID(ctx context.Context, areaID int64) ([]Canteen, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.area_id, a.name, c.name, c.description, c.specialties, c.created_at, c.updated_at
		FROM canteens c
		JOIN dining_areas a ON a.id = c.area_id
		WHERE c.area_id = ?
		ORDER BY c.id ASC`, areaID)
	if err != nil {
		return nil, fmt.Errorf("list canteens by area: %w", err)
	}
	defer rows.Close()
	items := make([]Canteen, 0)
	for rows.Next() {
		var item Canteen
		if err := rows.Scan(&item.ID, &item.AreaID, &item.AreaName, &item.Name, &item.Description, &item.Specialties, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate canteens: %w", err)
	}
	return items, nil
}

func (r *Repository) GetCanteenByID(ctx context.Context, id int64) (Canteen, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.area_id, a.name, c.name, c.description, c.specialties, c.created_at, c.updated_at
		FROM canteens c
		JOIN dining_areas a ON a.id = c.area_id
		WHERE c.id = ?`, id)
	var item Canteen
	if err := row.Scan(&item.ID, &item.AreaID, &item.AreaName, &item.Name, &item.Description, &item.Specialties, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return Canteen{}, err
	}
	return item, nil
}
