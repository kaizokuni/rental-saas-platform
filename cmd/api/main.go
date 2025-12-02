package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"rental-saas/internal/auth"
	"rental-saas/internal/database"
	"rental-saas/internal/handlers"
	"rental-saas/internal/middleware"
	"rental-saas/internal/repository"
	"rental-saas/internal/storage"
)

func main() {
	if err := database.Connect(); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.DB.Close()

	// Initialize MinIO
	minioClient, err := storage.NewMinioClient("localhost:9000", "minioadmin", "minioadmin", "cars")
	if err != nil {
		fmt.Printf("Failed to initialize MinIO: %v\n", err)
		os.Exit(1)
	}

	// Initialize Repositories and Handlers
	carRepo := repository.NewCarRepository(database.DB)
	carHandler := handlers.NewCarHandler(carRepo, minioClient)

	// Widget Handler
	widgetHandler := handlers.NewWidgetHandler()

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// --- Public API Group (Widget) ---
	r.Group(func(r chi.Router) {
		// Open CORS for Widget
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Content-Type"},
		}))
		// Strict Rate Limit (10 req/min)
		r.Use(httprate.LimitByIP(10, 1*time.Minute))

		r.Get("/api/public/cars", widgetHandler.GetPublicCars)
		r.Post("/api/public/book", widgetHandler.PublicBook)
	})

	// --- Tenant / Protected API Group ---
	r.Group(func(r chi.Router) {
		// Restricted CORS
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://admin.localhost:5173", os.Getenv("FRONTEND_URL")},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))

		r.Use(middleware.TenantMiddleware(database.DB))

		// Security Middleware
		r.Use(httprate.LimitByIP(100, 1*time.Minute))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tenant := r.Context().Value(middleware.TenantKey).(string)
			w.Write([]byte(fmt.Sprintf("Car Rental Project Initialized. Tenant: %s", tenant)))
		})

		r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Test OK"))
		})

		// Routes
		r.Route("/api", func(r chi.Router) {
			r.Post("/auth/login", auth.LoginHandler)
			
			// Protected Routes
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				// TenantMiddleware is already applied in the parent group
				
				r.Get("/cars", carHandler.ListCars)
				r.Post("/cars", carHandler.CreateCar)
				r.Put("/cars/{id}", carHandler.UpdateCar)

				r.Post("/payments/intent", handlers.NewPaymentHandler().CreatePaymentIntent)
				r.Post("/bookings", handlers.NewBookingHandler().CreateBooking)
				r.Post("/bookings/{id}/return", handlers.NewBookingHandler().ReturnCar)
				r.Get("/dashboard/stats", handlers.NewDashboardHandler().GetDashboardStats)
				r.Get("/bookings/{id}/invoice", handlers.NewInvoiceHandler().GenerateInvoice)
				r.Post("/settings/webhooks", handlers.NewSettingsHandler().RegisterWebhook)
				r.Get("/settings/webhooks", handlers.NewSettingsHandler().ListWebhooks)
			})
		})

		r.Post("/api/webhooks/stripe", handlers.NewPaymentHandler().HandleStripeWebhook)
	})

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("SERVER STARTED on port %s\n", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %s\n", err)
	}

	fmt.Println("Server exiting")
}
