package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/flight-booking/api/internal/model"
	"github.com/flight-booking/api/internal/store"
	redislock "github.com/flight-booking/api/internal/redis"
)

const pnrChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type BookingHandler struct {
	store  store.Store
	locker redislock.SeatLocker
}

func NewBookingHandler(s store.Store, l redislock.SeatLocker) *BookingHandler {
	return &BookingHandler{store: s, locker: l}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func generatePNR() (string, error) {
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pnrChars))))
		if err != nil {
			return "", err
		}
		b[i] = pnrChars[n.Int64()]
	}
	return string(b), nil
}

func generateBookingID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("bk_%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req model.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateBookingRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.Seats) != len(req.Passengers) {
		writeError(w, http.StatusUnprocessableEntity, "seats count must equal passengers count")
		return
	}

	flight, err := h.store.GetFlight(req.FlightID)
	if err != nil {
		writeError(w, http.StatusNotFound, "flight not found")
		return
	}

	ctx := context.Background()
	lockedSeats := []string{}
	for _, seat := range req.Seats {
		acquired, err := h.locker.LockSeat(ctx, req.FlightID, seat)
		if err != nil {
			for _, s := range lockedSeats {
				h.locker.UnlockSeat(ctx, req.FlightID, s)
			}
			writeError(w, http.StatusInternalServerError, "seat locking error")
			return
		}
		if !acquired {
			for _, s := range lockedSeats {
				h.locker.UnlockSeat(ctx, req.FlightID, s)
			}
			writeError(w, http.StatusConflict, fmt.Sprintf("seat %s is already locked", seat))
			return
		}
		lockedSeats = append(lockedSeats, seat)
	}

	bookingID, err := generateBookingID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate booking ID")
		return
	}
	pnr, err := generatePNR()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate PNR")
		return
	}

	now := time.Now().UTC()
	booking := &model.Booking{
		BookingID:       bookingID,
		PNR:             pnr,
		Status:          "PENDING",
		FlightID:        flight.FlightID,
		FlightNumber:    flight.FlightNumber,
		Origin:          flight.Origin,
		Destination:     flight.Destination,
		DepartureAt:     flight.DepartureAt,
		Cabin:           req.Cabin,
		FareClass:       req.FareClass,
		Seats:           req.Seats,
		Passengers:      req.Passengers,
		ContactEmail:    req.ContactEmail,
		ContactPhone:    req.ContactPhone,
		TotalAmount:     flight.Price * float64(len(req.Seats)),
		Currency:        flight.Currency,
		PaymentDeadline: now.Add(15 * time.Minute),
		CreatedAt:       now,
	}

	if err := h.store.CreateBooking(booking); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store booking")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": booking})
}

func (h *BookingHandler) GetBookingByID(w http.ResponseWriter, r *http.Request) {
	bookingID := strings.TrimPrefix(r.URL.Path, "/api/v1/bookings/")
	if bookingID == "" {
		writeError(w, http.StatusBadRequest, "missing bookingId")
		return
	}

	booking, err := h.store.GetBookingByID(bookingID)
	if err != nil {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": booking})
}

func (h *BookingHandler) GetBookingByPNR(w http.ResponseWriter, r *http.Request) {
	pnr := strings.TrimPrefix(r.URL.Path, "/api/v1/bookings/pnr/")
	if pnr == "" {
		writeError(w, http.StatusBadRequest, "missing pnr")
		return
	}

	booking, err := h.store.GetBookingByPNR(pnr)
	if err != nil {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": booking})
}

func validateBookingRequest(req *model.BookingRequest) error {
	if req.FlightID == "" {
		return fmt.Errorf("flightId is required")
	}
	if req.Cabin == "" {
		return fmt.Errorf("cabin is required")
	}
	if req.FareClass == "" {
		return fmt.Errorf("fareClass is required")
	}
	if len(req.Seats) == 0 {
		return fmt.Errorf("at least one seat is required")
	}
	if len(req.Passengers) == 0 {
		return fmt.Errorf("at least one passenger is required")
	}
	if req.ContactEmail == "" {
		return fmt.Errorf("contactEmail is required")
	}
	for i, p := range req.Passengers {
		if p.FirstName == "" {
			return fmt.Errorf("passenger[%d].firstName is required", i)
		}
		if p.LastName == "" {
			return fmt.Errorf("passenger[%d].lastName is required", i)
		}
		if p.PassportNumber == "" {
			return fmt.Errorf("passenger[%d].passportNumber is required", i)
		}
	}
	return nil
}
