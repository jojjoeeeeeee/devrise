import type { BookingRequest, BookingResponse, ErrorResponse } from "./types";

declare const API_URL: string | undefined;
const apiBaseUrl = typeof API_URL !== "undefined" ? API_URL : "http://localhost:8080";

const form = document.getElementById("booking-form") as HTMLFormElement;
const submitBtn = document.getElementById("submit-btn") as HTMLButtonElement;
const resultBox = document.getElementById("result") as HTMLDivElement;

function showSuccess(booking: BookingResponse["data"]): void {
  resultBox.className = "success";
  resultBox.innerHTML = `
    <div class="result-label">Booking Confirmed</div>
    <div class="booking-summary">
      <span><strong>Booking ID:</strong> ${booking.bookingId}</span>
      <span><strong>PNR:</strong> ${booking.pnr}</span>
      <span><strong>Status:</strong> ${booking.status}</span>
      <span><strong>Flight:</strong> ${booking.flightNumber} · ${booking.origin} → ${booking.destination}</span>
      <span><strong>Departure:</strong> ${new Date(booking.departureAt).toLocaleString()}</span>
      <span><strong>Seat(s):</strong> ${booking.seats.join(", ")}</span>
      <span><strong>Cabin:</strong> ${booking.cabin} (${booking.fareClass})</span>
      <span><strong>Total:</strong> ${booking.totalAmount.toLocaleString()} ${booking.currency}</span>
      <span><strong>Pay by:</strong> ${new Date(booking.paymentDeadline).toLocaleString()}</span>
    </div>
  `;
}

function showError(message: string): void {
  resultBox.className = "error";
  resultBox.innerHTML = `<div class="result-label">Error</div><div>${message}</div>`;
}

function getFormData(f: HTMLFormElement): BookingRequest {
  const data = new FormData(f);
  const seats = (data.get("seats") as string)
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

  return {
    flightId: data.get("flightId") as string,
    cabin: data.get("cabin") as string,
    fareClass: (data.get("fareClass") as string).toUpperCase(),
    seats,
    passengers: [
      {
        firstName: data.get("firstName") as string,
        lastName: data.get("lastName") as string,
        dateOfBirth: data.get("dateOfBirth") as string,
        nationality: (data.get("nationality") as string).toUpperCase(),
        passportNumber: data.get("passportNumber") as string,
        passportExpiry: data.get("passportExpiry") as string,
      },
    ],
    contactEmail: data.get("contactEmail") as string,
    contactPhone: data.get("contactPhone") as string,
  };
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  resultBox.className = "";
  resultBox.style.display = "none";
  submitBtn.disabled = true;
  submitBtn.textContent = "Booking…";

  const payload = getFormData(form);

  try {
    const response = await fetch(`${apiBaseUrl}/api/v1/bookings`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const text = await response.text();

    if (response.ok) {
      const result = JSON.parse(text) as BookingResponse;
      showSuccess(result.data);
    } else {
      const err = JSON.parse(text) as ErrorResponse;
      showError(`${response.status}: ${err.error}`);
    }
  } catch (err) {
    showError(err instanceof Error ? err.message : "Network error — is the API running?");
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = "Book Flight";
  }
});
