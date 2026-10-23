package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"barber-one/backend/internal/config"
	"barber-one/backend/internal/controller"
	"barber-one/backend/internal/middleware"
	"barber-one/backend/internal/repository"
	"barber-one/backend/internal/service"
	_ "barber-one/backend/docs"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := config.InitDBPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Koneksi PostgreSQL gagal: %v", err)
	}
	defer db.Close()
	log.Println("Koneksi PostgreSQL berhasil")

	// Initialize repository
	repo := repository.NewRepository(db)

	// Initialize services
	userService := service.NewUserService(repo)
	bookingService := service.NewBookingService(repo)
	serviceService := service.NewServiceService(repo)
	barberService := service.NewBarberService(repo)
	reviewService := service.NewReviewService(repo)
	cashierService := service.NewCashierService(repo)

	// Initialize midtrans service
	midtransService, err := service.NewMidtransService()
	if err != nil {
		log.Printf("Warning: Midtrans not configured: %v", err)
	}

	// Initialize middleware
	sessionStore := middleware.NewSessionStore()

	// Initialize controllers
	authController := controller.NewAuthController(userService, sessionStore)
	bookingController := controller.NewBookingController(bookingService, userService, sessionStore)
	serviceController := controller.NewServiceController(serviceService)
	barberController := controller.NewBarberController(barberService)
	reviewController := controller.NewReviewController(reviewService, sessionStore)
	cashierController := controller.NewCashierController(cashierService, sessionStore)
	paymentController := controller.NewPaymentController(midtransService)
	ticketController := controller.NewTicketController(sessionStore)

	// Initialize workers
	workerService := service.NewWorkerService(db)
	workerService.StartPaymentExpiryWorker()
	workerService.StartBookingAutoStartWorker()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Swagger - just serve JSON and YAML
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger/" || r.URL.Path == "/swagger/index.html" {
			// Simple HTML page that uses CDN for Swagger UI
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Swagger UI</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function() {
window.ui = SwaggerUIBundle({
url: "/swagger/swagger.json",
dom_id: '#swagger-ui',
presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
layout: "StandaloneLayout"
})
}
</script>
</body>
</html>`))
			return
		}
		if r.URL.Path == "/swagger/swagger.json" {
			http.ServeFile(w, r, "docs/swagger.json")
			return
		}
		if r.URL.Path == "/swagger/swagger.yaml" {
			http.ServeFile(w, r, "docs/swagger.yaml")
			return
		}
		http.NotFound(w, r)
	})

	// Status
	mux.HandleFunc("/api/status", paymentController.Status)
	mux.HandleFunc("/", paymentController.Status)

	// Auth
	mux.HandleFunc("/api/register", authController.Register)
	mux.HandleFunc("/api/login", authController.Login)
	mux.HandleFunc("/api/profile", authController.Profile)
	mux.HandleFunc("/api/logout", authController.Logout)

	// Services
	mux.HandleFunc("/api/services", serviceController.ListServices)

	// Barbers
	mux.HandleFunc("/api/barbers", barberController.ListBarbers)

	// Bookings
	mux.HandleFunc("/api/bookings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			bookingController.ListUserBookings(w, r)
		case http.MethodPost:
			bookingController.CreateBooking(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		}
	})
	mux.HandleFunc("/api/bookings/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 3 && parts[0] == "api" && parts[1] == "bookings" && r.Method == http.MethodGet {
			if bookingID, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				bookingController.GetBookingDetail(w, r, bookingID)
				return
			}
		}
		if len(parts) == 5 && parts[0] == "api" && parts[1] == "bookings" && parts[3] == "payment" && r.Method == http.MethodPost {
			if bookingID, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				bookingController.CreatePayment(w, r, bookingID)
				return
			}
		}
		if len(parts) == 4 && parts[0] == "api" && parts[1] == "bookings" {
			if bookingID, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				switch parts[3] {
				case "cancel":
					bookingController.CancelBooking(w, r, bookingID)
				case "reschedule":
					bookingController.RescheduleBooking(w, r, bookingID)
				default:
					http.NotFound(w, r)
				}
				return
			}
		}
		http.NotFound(w, r)
	})

	// Reviews
	mux.HandleFunc("/api/reviews", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			reviewController.ListReviews(w, r)
		case http.MethodPost:
			reviewController.CreateReview(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			middleware.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metode tidak diizinkan."})
		}
	})

	// Availability
	mux.HandleFunc("/api/availability", bookingController.CheckAvailability)

	// My Tickets
	mux.HandleFunc("/api/my-tickets", ticketController.MyTickets)

	// Cashier
	mux.HandleFunc("/api/cashier/queue", cashierController.GetQueue)
	mux.HandleFunc("/api/cashier/summary", cashierController.GetSummary)
	mux.HandleFunc("/api/cashier/walk-in", cashierController.WalkIn)
	mux.HandleFunc("/api/cashier/bookings/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 5 && parts[0] == "api" && parts[1] == "cashier" && parts[2] == "bookings" && r.Method == http.MethodPatch {
			switch parts[4] {
			case "finish":
				middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Service finished"})
			case "cancel":
				middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "Service cancelled"})
			default:
				http.NotFound(w, r)
			}
			return
		}
		http.NotFound(w, r)
	})

	// Midtrans Notification
	mux.HandleFunc("/api/payments/midtrans/notification", paymentController.MidtransNotification)

	// Start server
	fmt.Printf("Backend berjalan di http://localhost:%s\n", cfg.ServerPort)
	handler := middleware.CORSMiddleware(mux, cfg.FrontendOrigin)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatalf("Server gagal: %v", err)
	}
}
