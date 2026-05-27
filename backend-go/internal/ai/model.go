package ai

type RecommendedDish struct {
	DishID        int64   `json:"dish_id"`
	DishName      string  `json:"dish_name"`
	StoreID       int64   `json:"store_id"`
	StoreName     string  `json:"store_name"`
	Price         float64 `json:"price"`
	SalesCount    int64   `json:"sales_count"`
	OrderCount    int64   `json:"order_count"`
	TotalQuantity int64   `json:"total_quantity"`
	TotalAmount   float64 `json:"total_amount"`
	Score         float64 `json:"score"`
	Reason        string  `json:"reason"`
}

type DishDescriptionRequest struct {
	DishName    string `json:"dish_name"`
	Ingredients string `json:"ingredients"`
	Style       string `json:"style"`
}

type DishDescriptionResponse struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	RecommendReason string `json:"recommend_reason"`
	Message         string `json:"message,omitempty"`
}

type DishImagePromptRequest struct {
	DishName    string `json:"dish_name"`
	Description string `json:"description"`
	Style       string `json:"style"`
}

type DishImagePromptResponse struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	Note           string `json:"note"`
	Message        string `json:"message,omitempty"`
}

type recommendationStat struct {
	DishID        int64
	DishName      string
	StoreID       int64
	StoreName     string
	Price         float64
	SalesCount    int64
	OrderCount    int64
	TotalQuantity int64
	TotalAmount   float64
	StoreRating   float64
}
