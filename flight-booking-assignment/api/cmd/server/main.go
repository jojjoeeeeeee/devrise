package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/flight-booking/api/internal/handler"
	"github.com/flight-booking/api/internal/middleware"
	redislock "github.com/flight-booking/api/internal/redis"
	"github.com/flight-booking/api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	memStore := store.NewMemoryStore()
	locker := redislock.NewRedisLocker(redisAddr)
	bookingHandler := handler.NewBookingHandler(memStore, locker)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/bookings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			bookingHandler.CreateBooking(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// /api/v1/bookings/pnr/:pnr must be matched before /api/v1/bookings/:bookingId
	mux.HandleFunc("/api/v1/bookings/pnr/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			bookingHandler.GetBookingByPNR(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/v1/bookings/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Avoid matching /api/v1/bookings/pnr/ here — the more specific pattern wins
		if strings.HasPrefix(path, "/api/v1/bookings/pnr/") {
			bookingHandler.GetBookingByPNR(w, r)
			return
		}
		if r.Method == http.MethodGet {
			middleware.RequireBearer(http.HandlerFunc(bookingHandler.GetBookingByID)).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})

	handler := middleware.CORS(mux)

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
