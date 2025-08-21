package main

import (
	"fmt"
	"gin-freemarket/controllers"
	"gin-freemarket/infra"
	"net/http"
	"os"
	"runtime"

	//	"gin-freemarket/models"
	"gin-freemarket/middlewares"
	"gin-freemarket/repositories"
	"gin-freemarket/services"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

var (
	// HTTP request counter
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request duration histogram
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func initMetrics() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "go_goroutines",
			Help: "Number of goroutines",
		},
		func() float64 { return float64(runtime.NumGoroutine()) },
	))
}

// Structure for setting up dependencies
type Dependencies struct {
	IItemController     controllers.IItemController
	IAuthController     controllers.IAuthController
	IPurchaseController controllers.IPurchaseController
	AuthMiddleware      gin.HandlerFunc
	SessionMiddleware   gin.HandlerFunc
	WebMonitoring       middlewares.WebMonitoring
}

// Function to initialize dependencies
func setupDependencies(db *gorm.DB) *Dependencies {
	// Item
	itemRepository := repositories.NewItemRepository(db)
	itemService := services.NewItemService(itemRepository, db)
	itemController := controllers.NewItemController(itemService)

	// Auth
	authRepository := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepository, db)
	authController := controllers.NewAuthController(authService)

	// auth middlware
	authMiddleware := middlewares.AuthMiddleware(authService)
	//session middleware
	sessionMiddleware := middlewares.SessionMiddleware(authService)

	// Purchase
	purchaseRepository := repositories.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepository, itemRepository, db)
	purchaseController := controllers.NewPurchaseController(purchaseService)

	// monitoring
	webMonitoring := middlewares.NewPrometheusMonitorWebRequest()

	return &Dependencies{
		IItemController:     itemController,
		IAuthController:     authController,
		IPurchaseController: purchaseController,
		AuthMiddleware:      authMiddleware,
		SessionMiddleware:   sessionMiddleware,
		WebMonitoring:       webMonitoring,
	}
}

func main() {
	fmt.Println("main started")
	infra.Initialize()
	db := infra.SetupDB()
	//initMetrics()

	deps := setupDependencies(db)

	router := gin.Default()
	router.Use(gin.Logger())

	// monitoring
	router.Use(deps.WebMonitoring.MonitorWebRequest())
	router.GET("/metrics", deps.WebMonitoring.Metrics())

	// health check
	router.GET("/alive", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "this is alive",
		})
	})

	// item controllers
	itemRouter := router.Group("/items")
	{
		itemRouter.GET("", deps.IItemController.FindAll)
		itemRouter.GET("/:id", deps.IItemController.FindById)

		itemRouter.Use(deps.AuthMiddleware)
		itemRouter.POST("", deps.IItemController.Create)
		itemRouter.PUT("/:id", deps.IItemController.Update)
		itemRouter.DELETE("/:id", deps.IItemController.Delete)
	}

	// auth controllers
	authRouter := router.Group("/auth")
	{
		authRouter.POST("/register", deps.IAuthController.Register)
		authRouter.POST("/login", deps.IAuthController.Login)
	}

	// purchase controllers
	purchaseRouter := router.Group("/purchases")
	{
		purchaseRouter.Use(deps.AuthMiddleware, deps.SessionMiddleware)
		purchaseRouter.POST("", deps.IPurchaseController.Create)
		purchaseRouter.GET("", deps.IPurchaseController.FindAll)
		purchaseRouter.GET("/:id", deps.IPurchaseController.FindById)
	}

	// Setup metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	router.Run(":" + os.Getenv("APP_PORT"))
}
