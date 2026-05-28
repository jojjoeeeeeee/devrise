import type { BookingRequest, BookingResponse, ErrorResponse } from "./types";

function parseArgs(argv: string[]): Map<string, string> {
  const args = new Map<string, string>();
  for (let i = 0; i < argv.length; i++) {
    const current = argv[i];
    if (current !== undefined && current.startsWith("--")) {
      const key = current.slice(2);
      const next = argv[i + 1];
      const value = next !== undefined && !next.startsWith("--") ? next : "";
      args.set(key, value);
      if (value) i++;
    }
  }
  return args;
}

function usage(): never {
  console.error(
    "Usage: bun run book.ts --flight <flightId> --seat <seat> --passenger-first <firstName> --passenger-last <lastName> [--cabin economy] [--fare-class M] [--dob 1990-01-01] [--nationality THA] [--passport AA1234567] [--passport-expiry 2030-12-31] [--email <email>] [--phone <phone>]"
  );
  process.exit(1);
}

async function main(): Promise<void> {
  const args = parseArgs(process.argv.slice(2));

  const flightId = args.get("flight");
  const seat = args.get("seat");
  const firstName = args.get("passenger-first");
  const lastName = args.get("passenger-last");

  if (!flightId || !seat || !firstName || !lastName) {
    console.error("Error: --flight, --seat, --passenger-first, and --passenger-last are required.");
    usage();
  }

  const cabin = args.get("cabin") ?? "economy";
  const fareClass = args.get("fare-class") ?? "M";
  const dateOfBirth = args.get("dob") ?? "1990-01-01";
  const nationality = args.get("nationality") ?? "THA";
  const passportNumber = args.get("passport") ?? "AA1234567";
  const passportExpiry = args.get("passport-expiry") ?? "2030-12-31";
  const contactEmail = args.get("email") ?? "traveler@example.com";
  const contactPhone = args.get("phone") ?? "+66800000000";

  const apiBaseUrl = process.env["API_URL"] ?? "http://localhost:8080";

  const payload: BookingRequest = {
    flightId,
    cabin,
    fareClass,
    seats: [seat],
    passengers: [
      {
        firstName,
        lastName,
        dateOfBirth,
        nationality,
        passportNumber,
        passportExpiry,
      },
    ],
    contactEmail,
    contactPhone,
  };

  let response: Response;
  try {
    response = await fetch(`${apiBaseUrl}/api/v1/bookings`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
  } catch (err) {
    console.error("Request failed:", err instanceof Error ? err.message : String(err));
    process.exit(1);
  }

  const body = await response.text();

  if (response.ok) {
    const result = JSON.parse(body) as BookingResponse;
    console.log(JSON.stringify(result, null, 2));
  } else {
    const errResult = JSON.parse(body) as ErrorResponse;
    console.error(`Error ${response.status}:`, errResult.error);
    process.exit(1);
  }
}

main();
