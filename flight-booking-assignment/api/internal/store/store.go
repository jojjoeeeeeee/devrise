package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/flight-booking/api/internal/model"
)

type Store interface {
	GetFlight(flightID string) (*model.Flight, error)
	CreateBooking(booking *model.Booking) error
	GetBookingByID(bookingID string) (*model.Booking, error)
	GetBookingByPNR(pnr string) (*model.Booking, error)
}

type MemoryStore struct {
	mu       sync.RWMutex
	flights  map[string]*model.Flight
	bookings map[string]*model.Booking
	pnrIndex map[string]*model.Booking
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		flights:  make(map[string]*model.Flight),
		bookings: make(map[string]*model.Booking),
		pnrIndex: make(map[string]*model.Booking),
	}
	s.seedFlights()
	return s
}

func (s *MemoryStore) seedFlights() {
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

	dep2, _ := time.Parse(time.RFC3339, "2026-06-15T14:00:00+07:00")
	s.flights["fl_1a2b3c4d-5e6f-7a8b-9c0e"] = &model.Flight{
		FlightID:     "fl_1a2b3c4d-5e6f-7a8b-9c0e",
		FlightNumber: "QL202",
		Origin:       "NRT",
		Destination:  "BKK",
		DepartureAt:  dep2,
		Cabin:        "economy",
		Price:        11800.00,
		Currency:     "THB",
	}
}

func (s *MemoryStore) GetFlight(flightID string) (*model.Flight, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flights[flightID]
	if !ok {
		return nil, fmt.Errorf("flight not found")
	}
	return f, nil
}

func (s *MemoryStore) CreateBooking(booking *model.Booking) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookings[booking.BookingID] = booking
	s.pnrIndex[booking.PNR] = booking
	return nil
}

func (s *MemoryStore) GetBookingByID(bookingID string) (*model.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bookings[bookingID]
	if !ok {
		return nil, fmt.Errorf("booking not found")
	}
	return b, nil
}

func (s *MemoryStore) GetBookingByPNR(pnr string) (*model.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.pnrIndex[pnr]
	if !ok {
		return nil, fmt.Errorf("booking not found")
	}
	return b, nil
}
