package store

import "time"

type StoreStatus string

const (
	StoreStatusOpen     StoreStatus = "OPEN"
	StoreStatusClosed   StoreStatus = "CLOSED"
	StoreStatusDisabled StoreStatus = "DISABLED"
)

func (s StoreStatus) IsValid() bool {
	switch s {
	case StoreStatusOpen, StoreStatusClosed, StoreStatusDisabled:
		return true
	default:
		return false
	}
}

type DishStatus string

const (
	DishStatusOnSale  DishStatus = "ON_SALE"
	DishStatusOffSale DishStatus = "OFF_SALE"
)

func (s DishStatus) IsValid() bool {
	switch s {
	case DishStatusOnSale, DishStatusOffSale:
		return true
	default:
		return false
	}
}

type Store struct {
	ID          int64       `json:"id"`
	MerchantID  int64       `json:"merchant_id"`
	CanteenID   int64       `json:"canteen_id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	Address     string      `json:"address,omitempty"`
	Status      StoreStatus `json:"status"`
	Rating      float64     `json:"rating"`
	SalesCount  int64       `json:"sales_count"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type DishCategory struct {
	ID        int64     `json:"id"`
	StoreID   int64     `json:"store_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Dish struct {
	ID          int64      `json:"id"`
	StoreID     int64      `json:"store_id"`
	CategoryID  int64      `json:"category_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	Status      DishStatus `json:"status"`
	ImageURL    string     `json:"image_url,omitempty"`
	SalesCount  int64      `json:"sales_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Review struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	StoreID   int64     `json:"store_id"`
	Rating    int       `json:"rating"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreUpdateInput struct {
	Name        *string
	Description *string
	Phone       *string
	Address     *string
	Status      *StoreStatus
}

type CategoryCreateInput struct {
	Name      string
	SortOrder int
}

type CategoryUpdateInput struct {
	Name      *string
	SortOrder *int
}

type DishCreateInput struct {
	CategoryID  int64
	Name        string
	Description string
	Price       float64
	Stock       int
	Status      DishStatus
	ImageURL    string
}

type DishUpdateInput struct {
	CategoryID  *int64
	Name        *string
	Description *string
	Price       *float64
	Stock       *int
	ImageURL    *string
}

type DishStatusUpdateInput struct {
	Status DishStatus
}

type DishQueryFilter struct {
	CategoryID *int64
	Status     *DishStatus
	Keyword    string
}
