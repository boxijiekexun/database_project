package order

import "time"

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusAccepted       OrderStatus = "ACCEPTED"
	OrderStatusPreparing      OrderStatus = "PREPARING"
	OrderStatusDelivering     OrderStatus = "DELIVERING"
	OrderStatusCompleted      OrderStatus = "COMPLETED"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPendingPayment, OrderStatusPaid, OrderStatusAccepted, OrderStatusPreparing, OrderStatusDelivering, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (s OrderStatus) CanMerchantTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderStatusPaid:
		return next == OrderStatusAccepted
	case OrderStatusAccepted:
		return next == OrderStatusPreparing
	case OrderStatusPreparing:
		return next == OrderStatusDelivering
	case OrderStatusDelivering:
		return next == OrderStatusCompleted
	default:
		return false
	}
}

type CartItem struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	StoreID    int64     `json:"store_id"`
	DishID     int64     `json:"dish_id"`
	Quantity   int       `json:"quantity"`
	StoreName  string    `json:"store_name,omitempty"`
	DishName   string    `json:"dish_name,omitempty"`
	DishPrice  float64   `json:"dish_price,omitempty"`
	Subtotal   float64   `json:"subtotal,omitempty"`
	DishStatus string    `json:"dish_status,omitempty"`
	DishStock  int       `json:"dish_stock,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Order struct {
	ID          int64       `json:"id"`
	OrderNo     string      `json:"order_no"`
	UserID      int64       `json:"user_id"`
	StoreID     int64       `json:"store_id"`
	StoreName   string      `json:"store_name,omitempty"`
	UserName    string      `json:"user_name,omitempty"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	Remark      string      `json:"remark,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Items       []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	DishID    int64     `json:"dish_id"`
	DishName  string    `json:"dish_name"`
	DishPrice float64   `json:"dish_price"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
}

type Review struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	UserID    int64     `json:"user_id"`
	StoreID   int64     `json:"store_id"`
	Rating    int       `json:"rating"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CartAddInput struct {
	DishID   int64
	Quantity int
}

type CartUpdateInput struct {
	Quantity int
}

type CreateOrderInput struct {
	Remark string
}

type ReviewInput struct {
	Rating  int
	Content string
}

type OrderStatusUpdateInput struct {
	Status OrderStatus
}

type OrderQueryFilter struct {
	Status  *OrderStatus
	StoreID *int64
	UserID  *int64
}
