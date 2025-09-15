package router

import (
	"net/http"

	_ "github.com/Temisaputra/warOnk/docs" // wajib untuk register doc
	"github.com/Temisaputra/warOnk/internal/delivery/handler"
	"github.com/Temisaputra/warOnk/internal/delivery/middleware"
	"github.com/Temisaputra/warOnk/internal/infrastructure/config"
	"github.com/Temisaputra/warOnk/pkg/auth"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type Handlers struct {
	ProductHandler *handler.ProductHandler
	UserHandler    *handler.UserHandler
	AuthHandler    *handler.AuthHandler
	Logger         *zap.Logger
	JwtService     auth.JwtService
}

// NewRouter bikin router dan register semua endpoint
func NewRouter(handlers *Handlers) http.Handler {
	router := mux.NewRouter()
	config := config.Get()
	// contoh custom metric
	opsProcessed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "my_app_processed_ops_total",
		Help: "Total number of processed events",
	})

	prometheus.MustRegister(opsProcessed)

	authMW := middleware.NewAuthMiddleware(handlers.JwtService, opsProcessed)

	router.Handle("/metrics", promhttp.Handler())
	router.Use(middleware.LoggingMiddleware(handlers.Logger, config)) // <- inject logger

	// Swagger endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	api := router.PathPrefix("/api/war-onk").Subrouter()

	// ---------------- Public ----------------
	// Auth endpoints
	api.HandleFunc("/register", handlers.AuthHandler.Register).Methods("POST")
	api.HandleFunc("/login", handlers.AuthHandler.Login).Methods("POST")

	// ---------------- Protected ----------------
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMW.Authorization)
	// Product endpoints
	protected.HandleFunc("/products", handlers.ProductHandler.GetAllProduct).Methods("GET")
	protected.HandleFunc("/product/{id}", handlers.ProductHandler.GetProductByID).Methods("GET")
	protected.HandleFunc("/product-create", handlers.ProductHandler.CreateProduct).Methods("POST")
	protected.HandleFunc("/product-update/{id}", handlers.ProductHandler.UpdateProduct).Methods("PUT")
	protected.HandleFunc("/product-delete/{id}", handlers.ProductHandler.DeleteProduct).Methods("DELETE")

	// User endpoints
	protected.HandleFunc("/users", handlers.UserHandler.GetAllUser).Methods("GET")
	protected.HandleFunc("/user/{id}", handlers.UserHandler.GetUserByID).Methods("GET")
	protected.HandleFunc("/user-create", handlers.UserHandler.CreateUser).Methods("POST")
	protected.HandleFunc("/user-update/{id}", handlers.UserHandler.UpdateUser).Methods("PUT")
	protected.HandleFunc("/user-delete/{id}", handlers.UserHandler.DeleteUser).Methods("DELETE")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"POST", "GET", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Mode"},
		MaxAge:             60,
		AllowCredentials:   true,
		OptionsPassthrough: false,
	})

	return c.Handler(router)
}
