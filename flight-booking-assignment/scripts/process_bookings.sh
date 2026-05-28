#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INPUT="$SCRIPT_DIR/bookings.psv"
OUTPUT="$SCRIPT_DIR/report.txt"

if [[ ! -f "$INPUT" ]]; then
  echo "Error: $INPUT not found" >&2
  exit 1
fi

NOW_EPOCH=$(date -u +%s)
GENERATED=$(date -u +"%Y-%m-%d %H:%M:%S")

# Count bookings by status
count_confirmed=$(awk -F'|' 'NR>1 && $3=="CONFIRMED"{c++} END{print c+0}' "$INPUT")
count_pending=$(awk -F'|' 'NR>1 && $3=="PENDING"{c++} END{print c+0}' "$INPUT")
count_cancelled=$(awk -F'|' 'NR>1 && $3=="CANCELLED"{c++} END{print c+0}' "$INPUT")

# Sum confirmed revenue
confirmed_revenue=$(awk -F'|' 'NR>1 && $3=="CONFIRMED"{sum+=$4} END{printf "%.2f", sum+0}' "$INPUT")

# Get currency from first data row
currency=$(awk -F'|' 'NR==2{gsub(/[\r\n]/, "", $5); print $5}' "$INPUT")
currency="${currency:-THB}"

# Build expired PENDING list
expired_lines=""
while IFS='|' read -r booking_id pnr status amount curr deadline; do
  # Skip header
  [[ "$booking_id" == "bookingId" ]] && continue

  # Only process PENDING bookings
  [[ "$status" != "PENDING" ]] && continue

  # Strip trailing whitespace/newline from deadline
  deadline="${deadline%$'\r'}"
  deadline="${deadline%$'\n'}"
  deadline="${deadline%%[[:space:]]*}"

  # Convert deadline to epoch — support both GNU date (Linux) and BSD date (macOS)
  deadline_epoch=0
  if date --version >/dev/null 2>&1; then
    # GNU date
    deadline_epoch=$(date -u -d "$deadline" +%s 2>/dev/null || echo 0)
  else
    # BSD date (macOS)
    deadline_epoch=$(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$deadline" +%s 2>/dev/null || echo 0)
  fi

  if [[ "$deadline_epoch" -lt "$NOW_EPOCH" ]]; then
    expired_lines="${expired_lines}  ${booking_id}  ${pnr}  deadline: ${deadline}"$'\n'
  fi
done < "$INPUT"

# Write the report
{
  echo "=== Booking Report (generated: $GENERATED) ==="
  echo ""
  echo "Status Summary:"
  printf "  %-10s: %s\n" "CONFIRMED"  "$count_confirmed"
  printf "  %-10s: %s\n" "PENDING"    "$count_pending"
  printf "  %-10s: %s\n" "CANCELLED"  "$count_cancelled"
  echo ""
  echo "Total Revenue (CONFIRMED): $confirmed_revenue $currency"
  echo ""
  echo "Expired PENDING Bookings (payment overdue):"
  if [[ -z "$expired_lines" ]]; then
    echo "  (none)"
  else
    printf "%s" "$expired_lines"
  fi
} > "$OUTPUT"

echo "Report written to $OUTPUT"
