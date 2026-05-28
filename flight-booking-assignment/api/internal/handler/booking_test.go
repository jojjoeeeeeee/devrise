package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flight-booking/api/internal/handler"
	"github.com/flight-booking/api/internal/model"
)

// --- Mock Store ---

type mockStore struct {
	flights  map[string]*model.Flight
	bookings map[string]*model.Booking
	pnrIndex map[string]*model.Booking
}

func newMockStore() *mockStore {
	s := &mockStore{
		flights:  make(map[string]*model.Flight),
		bookings: make(map[string]*model.Booking),
		pnrIndex: make(map[string]*model.Booking),
	}
	dep, _ := time.Parse(time.RFC3339, "2026-06-01T08:30:00+07:00")
	s.flights["fl_8a2b3c4d-5e6f-7a8b-9c0d"] = &model.Flight{
		FlightID:     "fl_8a2b3c4d-5e6f-7a8b-9c0d",
		FlightNumber: "QL101",
		Origin:       "BKK",
		Destination:  "NRT",
		DepartureAt:  dep,
		Cabin:        "economy",
		Price:        12500.00,
		Currency:     "THB",
	}
	return s
}

func (m *mockStore) GetFlight(flightID string) (*model.Flight, error) {
	f, ok := m.flights[flightID]
	if !ok {
		return nil, fmt.Errorf("flight not found")
	}
	return f, nil
}

func (m *mockStore) CreateBooking(b *model.Booking) error {
	m.bookings[b.BookingID] = b
	m.pnrIndex[b.PNR] = b
	return nil
}

func (m *mockStore) GetBookingByID(id string) (*model.Booking, error) {
	b, ok := m.bookings[id]
	if !ok {
		return nil, fmt.Errorf("booking not found")
	}
	return b, nil
}

func (m *mockStore) GetBookingByPNR(pnr string) (*model.Booking, error) {
	b, ok := m.pnrIndex[pnr]
	if !ok {
		return nil, fmt.Errorf("booking not found")
	}
	return b, nil
}

// --- Mock SeatLocker ---

type mockLocker struct {
	locked map[string]bool
}

func newMockLocker() *mockLocker {
	return &mockLocker{locked: make(map[string]bool)}
}

func (m *mockLocker) LockSeat(ctx context.Context, flightID, seat string) (bool, error) {
	key := fmt.Sprintf("seat:%s:%s", flightID, seat)
	if m.locked[key] {
		return false, nil
	}
	m.locked[key] = true
	return true, nil
}

func (m *mockLocker) UnlockSeat(ctx context.Context, flightID, seat string) error {
	key := fmt.Sprintf("seat:%s:%s", flightID, seat)
	delete(m.locked, key)
	return nil
}

// --- Helpers ---

func validBookingBody() map[string]any {
	return map[string]any{
		"flightId":  "fl_8a2b3c4d-5e6f-7a8b-9c0d",
		"cabin":     "economy",
		"fareClass": "M",
		"seats":     []string{"14A"},
		"passengers": []map[string]any{
			{
				"firstName":      "Somchai",
				"lastName":       "Jaidee",
				"dateOfBirth":    "1990-05-15",
				"nationality":    "THA",
				"passportNumber": "AA1234567",
				"passportExpiry": "2030-12-31",
			},
		},
		"contactEmail": "somchai@example.com",
		"contactPhone": "+66812345678",
	}
}

func makeRequest(t *testing.T, h *handler.BookingHandler, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	return rr
}

// --- Tests ---

func TestCreateBooking_HappyPath(t *testing.T) {
	s := newMockStore()
	l := newMockLocker()
	h := handler.NewBookingHandler(s, l)

	rr := makeRequest(t, h, validBookingBody())

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'data' field")
	}
	if data["status"] != "PENDING" {
		t.Errorf("expected status PENDING, got %v", data["status"])
	}
	bookingID, _ := data["bookingId"].(string)
	if len(bookingID) < 4 || bookingID[:3] != "bk_" {
		t.Errorf("bookingId should start with bk_, got %q", bookingID)
	}
	pnr, _ := data["pnr"].(string)
	if len(pnr) != 6 {
		t.Errorf("pnr should be 6 chars, got %q", pnr)
	}
}

func TestCreateBooking_SeatConflict(t *testing.T) {
	s := newMockStore()
	l := newMockLocker()
	h := handler.NewBookingHandler(s, l)

	// First booking acquires the seat lock
	rr1 := makeRequest(t, h, validBookingBody())
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first booking expected 201, got %d", rr1.Code)
	}

	// Second booking for the same seat should get 409
	rr2 := makeRequest(t, h, validBookingBody())
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d — body: %s", rr2.Code, rr2.Body.String())
	}
}

func TestCreateBooking_MissingContactEmail(t *testing.T) {
	s := newMockStore()
	l := newMockLocker()
	h := handler.NewBookingHandler(s, l)

	body := validBookingBody()
	delete(body, "contactEmail")

	rr := makeRequest(t, h, body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
	}
}
