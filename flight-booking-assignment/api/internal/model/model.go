package model

import "time"

type Passenger struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	DateOfBirth    string `json:"dateOfBirth"`
	Nationality    string `json:"nationality"`
	PassportNumber string `json:"passportNumber"`
	PassportExpiry string `json:"passportExpiry"`
}

type BookingRequest struct {
	FlightID     string      `json:"flightId"`
	Cabin        string      `json:"cabin"`
	FareClass    string      `json:"fareClass"`
	Seats        []string    `json:"seats"`
	Passengers   []Passenger `json:"passengers"`
	ContactEmail string      `json:"contactEmail"`
	ContactPhone string      `json:"contactPhone"`
}

type Booking struct {
	BookingID       string      `json:"bookingId"`
	PNR             string      `json:"pnr"`
	Status          string      `json:"status"`
	FlightID        string      `json:"flightId"`
	FlightNumber    string      `json:"flightNumber"`
	Origin          string      `json:"origin"`
	Destination     string      `json:"destination"`
	DepartureAt     time.Time   `json:"departureAt"`
	Cabin           string      `json:"cabin"`
	FareClass       string      `json:"fareClass"`
	Seats           []string    `json:"seats"`
	Passengers      []Passenger `json:"passengers"`
	ContactEmail    string      `json:"contactEmail"`
	ContactPhone    string      `json:"contactPhone"`
	TotalAmount     float64     `json:"totalAmount"`
	Currency        string      `json:"currency"`
	PaymentDeadline time.Time   `json:"paymentDeadline"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type Flight struct {
	FlightID     string
	FlightNumber string
	Origin       string
	Destination  string
	DepartureAt  time.Time
	Cabin        string
	Price        float64
	Currency     string
}
