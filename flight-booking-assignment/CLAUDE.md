# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Flight Booking Service** — a full-stack assignment covering a Go REST API, TypeScript/Bun frontend or CLI, shell scripting, and Linux deployment.

## Planned Structure

```
flight-booking-assignment/
├── api/               # Go REST API
├── client/            # TypeScript/Bun frontend or CLI
├── scripts/           # Shell scripts (bookings.psv processor)
├── systemd/           # .service config files for deployment
└── assignment.md      # Full project spec
```

## Go API

### Commands
```bash
cd api
go run ./cmd/server          # Start dev server
go build -o bin/server ./cmd/server  # Build binary
go test ./...                # Run all tests
go test ./... -run TestName  # Run a single test
go vet ./...                 # Lint
```

### Endpoints
- `POST   /api/v1/bookings`            — Create pending booking (201), locks seat in Redis 15 min
- `GET    /api/v1/bookings/:bookingId` — Fetch by internal UUID (requires Bearer JWT)
- `GET    /api/v1/bookings/pnr/:pnr`  — Fetch by 6-char PNR (public)

### Error contracts
| Code | Meaning |
|------|---------|
| 400  | Validation error / missing required fields |
| 404  | flightId or bookingId not found |
| 409  | Seat already locked by another session |
| 422  | Seats count ≠ passengers count |

### Key design rules
- Seat locking is Redis-based with a 15-minute TTL (set on `POST`, cleared on confirm/cancel).
- `bookingId` uses the `bk_` prefix; `pnr` is a 6-char uppercase alphanumeric.
- CORS must allow only the deployed frontend origin — no wildcard in production.
- Internal `GET /:bookingId` requires `Authorization: Bearer <jwt>` (used by Payment Service).

## TypeScript / Bun Client

```bash
cd client
bun install
bun run dev          # Dev mode
bun build index.ts --outdir dist   # Bundle for browser (Option A: Web)
bun run book.ts -- --flight fl_xxx --seat 14A  # CLI usage (Option B: CLI)
bun test             # Run tests
```

- No `any` types — all request/response shapes must be fully typed.
- Web option: static files served by nginx on the VM.
- CLI option: compilable to a standalone binary via `bun build --compile`.

## Shell Script

```bash
cd scripts
bash process_bookings.sh    # Reads bookings.psv, writes report.txt
```

- Input: `bookings.psv` — pipe-separated, header row `bookingId|pnr|status|totalAmount|currency|paymentDeadline`
- Output: `report.txt` — status counts, total confirmed revenue, list of overdue PENDING bookings
- Uses `awk -F'|'` for parsing and `date -u +%s` for deadline comparison; no external tools.

## Testing Requirements

Three required cases for the Go API:
1. Happy path — booking created (201)
2. Seat conflict — same seat booked twice (409)
3. Missing required fields (400)

Choose unit tests (mocked deps), integration tests (real DB + Redis), or both.

## Deployment

- Go API and frontend each run as **systemd services** (auto-start, auto-restart).
- **nginx** acts as reverse proxy in front of the API; serves static frontend files.
- **HTTPS** via Let's Encrypt (`certbot`) — use `sslip.io` if no custom domain.
- Systemd `.service` files must be committed to `systemd/`.

## Dependencies (expected)

| Layer | Tech |
|-------|------|
| API   | Go, Redis (seat locking), PostgreSQL or SQLite (bookings store) |
| Client | Bun, TypeScript |
| Proxy | nginx + certbot |
| Infra | systemd on Debian/Linux VM |
