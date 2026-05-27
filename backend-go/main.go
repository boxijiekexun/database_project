package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"project1/internal/ai"
	"project1/internal/auth"
	"project1/internal/config"
	"project1/internal/db"
	"project1/internal/dining"
	"project1/internal/middleware"
	"project1/internal/order"
	"project1/internal/response"
	"project1/internal/store"
	"project1/internal/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	var database *db.Connection
	if cfg.MySQLDSN != "" {
		conn, err := db.Open(cfg.MySQLDSN)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		database = conn
		defer database.Close()
	} else {
		log.Println("MYSQL_DSN is empty, starting without database connection")
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		status := gin.H{"status": "ok"}
		if database != nil {
			if err := database.Ping(); err != nil {
				response.Error(c, http.StatusServiceUnavailable, err.Error())
				return
			}
			status["database"] = "up"
		}
		response.Success(c, status)
	})

	if database != nil {
		userRepo := user.NewRepository(database.SQLDB())
		userImportService := user.NewImportService(userRepo)
		userImportHandler := user.NewImportHandler(userImportService)
		userAdminService := user.NewAdminService(userRepo)
		userAdminHandler := user.NewAdminHandler(userAdminService)
		storeRepo := store.NewRepository(database.SQLDB())
		storeService := store.NewService(storeRepo)
		storeHandler := store.NewHandler(storeService)
		diningRepo := dining.NewRepository(database.SQLDB())
		diningService := dining.NewService(diningRepo, storeRepo)
		diningHandler := dining.NewHandler(diningService)
		aiRepo := ai.NewRepository(database.SQLDB())
		aiService := ai.NewService(aiRepo)
		aiHandler := ai.NewHandler(aiService)
		orderRepo := order.NewRepository(database.SQLDB())
		orderService := order.NewService(orderRepo, storeService)
		orderHandler := order.NewHandler(orderService)
		tokenManager := auth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.TokenTTLHours)*time.Hour)
		captchaStore := auth.NewCaptchaStore(10 * time.Minute)
		authHandler := auth.NewHandler(userRepo, tokenManager, captchaStore)

		api := router.Group("/api")
		authGroup := api.Group("/auth")
		authGroup.GET("/captcha", authHandler.Captcha)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)

		browseGroup := api.Group("", middleware.AuthRequired(tokenManager), middleware.RequireAnyRole(user.RoleStudent, user.RoleTeacher, user.RoleAdmin))
		browseGroup.GET("/stores", storeHandler.PublicListStores)
		browseGroup.GET("/stores/:id", storeHandler.PublicGetStore)
		browseGroup.GET("/stores/:id/categories", storeHandler.PublicListCategories)
		browseGroup.GET("/stores/:id/dishes", storeHandler.PublicListDishes)
		browseGroup.GET("/stores/:id/reviews", storeHandler.PublicListReviews)

		loginGroup := api.Group("", middleware.AuthRequired(tokenManager))
		loginGroup.GET("/dining-areas", diningHandler.ListAreas)
		loginGroup.GET("/dining-areas/:id/canteens", diningHandler.ListCanteens)
		loginGroup.GET("/canteens/:id", diningHandler.GetCanteen)
		loginGroup.GET("/canteens/:id/stores", diningHandler.ListStores)
		loginGroup.GET("/canteens/:id/dishes", diningHandler.ListDishes)

		aiGroup := api.Group("/ai", middleware.AuthRequired(tokenManager))
		aiGroup.GET("/recommend/dishes", middleware.RequireAnyRole(user.RoleStudent, user.RoleTeacher, user.RoleAdmin), aiHandler.RecommendDishes)
		aiGroup.POST("/dish-description", middleware.RequireAnyRole(user.RoleMerchant, user.RoleAdmin), aiHandler.DishDescription)
		aiGroup.POST("/dish-image-prompt", middleware.RequireAnyRole(user.RoleMerchant, user.RoleAdmin), aiHandler.DishImagePrompt)

		studentGroup := api.Group("/student", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleStudent))
		teacherGroup := api.Group("/teacher", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleTeacher))
		merchantGroup := api.Group("/merchant", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleMerchant))
		adminGroup := api.Group("/admin", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleAdmin))
		merchantOrderGroup := api.Group("/merchant", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleMerchant))
		adminOrderGroup := api.Group("/admin", middleware.AuthRequired(tokenManager), middleware.RequireRole(user.RoleAdmin))
		cartGroup := api.Group("/cart", middleware.AuthRequired(tokenManager), middleware.RequireAnyRole(user.RoleStudent, user.RoleTeacher))
		ordersGroup := api.Group("/orders", middleware.AuthRequired(tokenManager), middleware.RequireAnyRole(user.RoleStudent, user.RoleTeacher))

		merchantGroup.GET("/store", storeHandler.MerchantGetStore)
		merchantGroup.PUT("/store", storeHandler.MerchantUpdateStore)
		merchantGroup.GET("/categories", storeHandler.MerchantListCategories)
		merchantGroup.POST("/categories", storeHandler.MerchantCreateCategory)
		merchantGroup.PUT("/categories/:id", storeHandler.MerchantUpdateCategory)
		merchantGroup.DELETE("/categories/:id", storeHandler.MerchantDeleteCategory)
		merchantGroup.GET("/dishes", storeHandler.MerchantListDishes)
		merchantGroup.POST("/dishes", storeHandler.MerchantCreateDish)
		merchantGroup.PUT("/dishes/:id", storeHandler.MerchantUpdateDish)
		merchantGroup.PATCH("/dishes/:id/status", storeHandler.MerchantUpdateDishStatus)

		cartGroup.GET("/items", orderHandler.ListCartItems)
		cartGroup.POST("/items", orderHandler.AddCartItem)
		cartGroup.PUT("/items/:id", orderHandler.UpdateCartItem)
		cartGroup.DELETE("/items/:id", orderHandler.DeleteCartItem)
		cartGroup.DELETE("/items", orderHandler.ClearCart)

		ordersGroup.POST("", orderHandler.CreateOrder)
		ordersGroup.GET("", orderHandler.ListUserOrders)
		ordersGroup.GET("/:id", orderHandler.GetUserOrder)
		ordersGroup.POST("/:id/cancel", orderHandler.CancelUserOrder)
		ordersGroup.POST("/:id/pay", orderHandler.PayUserOrder)
		ordersGroup.POST("/:id/review", orderHandler.ReviewUserOrder)

		merchantOrderGroup.GET("/orders", orderHandler.ListMerchantOrders)
		merchantOrderGroup.GET("/orders/:id", orderHandler.GetMerchantOrder)
		merchantOrderGroup.PATCH("/orders/:id/status", orderHandler.UpdateMerchantOrderStatus)

		adminOrderGroup.GET("/orders", orderHandler.ListAdminOrders)
		adminOrderGroup.GET("/orders/:id", orderHandler.GetAdminOrder)
		adminGroup.POST("/users/import", userImportHandler.ImportUsers)
		adminGroup.GET("/users", userAdminHandler.ListUsers)
		adminGroup.GET("/users/:id", userAdminHandler.GetUser)
		adminGroup.PATCH("/users/:id/status", userAdminHandler.UpdateStatus)
		adminGroup.PATCH("/users/:id/role", userAdminHandler.UpdateRole)
		adminGroup.POST("/users/:id/reset-password", userAdminHandler.ResetPassword)

		studentGroup.GET("/dashboard", dashboard("student"))
		teacherGroup.GET("/dashboard", dashboard("teacher"))
		merchantGroup.GET("/dashboard", dashboard("merchant"))
		adminGroup.GET("/dashboard", dashboard("admin"))
	}

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on %s", cfg.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func dashboard(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := middleware.ClaimsFromContext(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		response.Success(c, gin.H{
			"message": fmt.Sprintf("%s dashboard ok", role),
			"user_id": claims.UserID,
			"role":    claims.Role,
		})
	}
}
