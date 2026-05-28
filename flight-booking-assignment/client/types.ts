export interface Passenger {
  firstName: string;
  lastName: string;
  dateOfBirth: string;
  nationality: string;
  passportNumber: string;
  passportExpiry: string;
}

export interface BookingRequest {
  flightId: string;
  cabin: string;
  fareClass: string;
  seats: string[];
  passengers: Passenger[];
  contactEmail: string;
  contactPhone: string;
}

export interface BookingData {
  bookingId: string;
  pnr: string;
  status: string;
  flightId: string;
  flightNumber: string;
  origin: string;
  destination: string;
  departureAt: string;
  cabin: string;
  fareClass: string;
  seats: string[];
  passengers: Passenger[];
  contactEmail: string;
  contactPhone: string;
  totalAmount: number;
  currency: string;
  paymentDeadline: string;
  createdAt: string;
}

export interface BookingResponse {
  data: BookingData;
}

export interface ErrorResponse {
  error: string;
}
