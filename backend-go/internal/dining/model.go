package dining

import "time"

type DiningArea struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Canteen struct {
	ID          int64     `json:"id"`
	AreaID      int64     `json:"area_id"`
	AreaName    string    `json:"area_name,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Specialties string    `json:"specialties,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
