# DevRise

Self-learning repository covering Go, Linux, Web Security, TypeScript, and Testing.

## Flight Booking Assignment

A full-stack flight booking service built as part of the DevRise program.

**Live URL:** https://104-154-23-198.sslip.io

### What's inside

| Layer | Tech |
|-------|------|
| API | Go — 3 REST endpoints, Redis seat locking, JWT Bearer auth |
| Frontend | TypeScript + Bun — web booking form, fully typed (no `any`) |
| Script | Bash — processes `bookings.psv`, writes `report.txt` with overdue detection |
| Infra | GCP e2-micro VM, nginx reverse proxy, systemd services, HTTPS via Let's Encrypt |

### Run locally

**API**
```bash
cd flight-booking-assignment/api
go run ./cmd/server
```

**Frontend**
```bash
cd flight-booking-assignment/client
bun install
bun run dev
```

**Shell script**
```bash
cd flight-booking-assignment/scripts
bash process_bookings.sh
```

**Tests**
```bash
cd flight-booking-assignment/api
go test ./...
```

### Submit criteria

| # | What | Status |
|---|------|--------|
| 1 | 3 booking endpoints running | ✅ |
| 2 | CORS demo — blocked vs allowed | ✅ |
| 3 | TypeScript client, no `any` | ✅ |
| 4 | Tests green — 201, 409, 400 | ✅ |
| 5 | systemd on HTTPS VM | ✅ |
| 6 | Shell script with overdue detection | ✅ |
