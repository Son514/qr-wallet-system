# QR Wallet System

Backend service that creates Xendit QRPh cash-in payment requests and InstaPay cash-out payouts, plus a static frontend for both flows.

## Requirements

- Go 1.27+
- A Xendit account with the **QRPh** channel activated (cash-in) and **Payouts** enabled (cash-out)
- A browser and internet access (frontend loads the QR library from a CDN)

## Setup

1. Clone the repository.
2. `cd xendit`
3. Copy `.env` values into your own `.env` (see below).
4. Install dependencies: `go mod download`

## Environment Variables

| Name | Description | Required |
|------|-------------|----------|
| `XENDIT_SECRET_KEY` | Xendit secret API key (basic auth username) | Yes |
| `XENDIT_API_VERSION` | Payments API version, e.g. `2024-11-11` | Yes |
| `XENDIT_PAYOUT_API_VERSION` | Payouts API version, e.g. `2025-09-01` | Yes |
| `XENDIT_BASE_URL` | Xendit API base URL, e.g. `https://api.xendit.co` | Yes |

## Running

Run the backend and frontend together:

```bash
make dev
```

- Backend: `http://localhost:8080` (Gin default; `PORT` is honored if set).
- Frontend: `http://localhost:5500`.

Or run them separately:

```bash
make backend   # cd xendit && go run .
make web       # cd web && python3 -m http.server 5500
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/cash-in/qr` | Creates a Xendit QRPh payment request and returns the QR string |
| POST | `/v1/cash-in/qr/:id/simulate` | Simulates a payment for a payment request (test mode only) |
| GET | `/v1/cash-in/qr/:id` | Returns the current status of a payment request |
| POST | `/v1/cash-out/payout` | Creates a Xendit payout to a PH bank or e-wallet (InstaPay/PESONet) |
| GET | `/v1/cash-out/payout/:id` | Returns the current status of a payout |

### Cash-in request body:

```json
{
  "reference_id": "test-001",
  "amount": 1500
}
```

Response body:

```json
{
  "payment_request_id": "pr-...",
  "reference_id": "test-001",
  "type": "PAY",
  "country": "PH",
  "currency": "PHP",
  "request_amount": 1500,
  "channel_code": "QRPH",
  "status": "REQUIRES_ACTION",
  "qr_string": "...",
  "actions": [
    { "type": "PRESENT_TO_CUSTOMER", "descriptor": "QR_STRING", "value": "..." }
  ],
  "created": "...",
  "updated": "..."
}
```

`POST /v1/cash-in/qr/:id/simulate` body:

```json
{
  "amount": 1500
}
```

### Cash-out request body:

```json
{
  "reference_id": "cashout-001",
  "amount": 1500,
  "channel": "PH_GCASH",
  "account_holder_name": "Maria Santos",
  "account_number": "09171234567"
}
```

`channel` is one of the supported PH destinations. Banks use `SWIFT` routing with the bank's BIC; e-wallets use `WALLET` routing with the channel code.

| Channel value | Destination | Routing |
|---------------|-------------|---------|
| `PH_BDO` | BDO | SWIFT / `BNORPHMM` |
| `PH_BPI` | BPI | SWIFT / `BOPIPHMM` |
| `PH_MBT` | Metrobank | SWIFT / `MBTCPHMM` |
| `PH_UBP` | UnionBank | SWIFT / `UBPHPHMM` |
| `PH_LBP` | LandBank | SWIFT / `TLBPPHMM` |
| `PH_SEC` | Security Bank | SWIFT / `SETCPHMM` |
| `PH_RCBC` | RCBC | SWIFT / `RCBCPHMM` |
| `PH_PNB` | PNB | SWIFT / `PNBMPHMM` |
| `PH_CBC` | China Bank | SWIFT / `CHBKPHMM` |
| `PH_DBP` | DBP | SWIFT / `DBPHPHMM` |
| `PH_GCASH` | GCash | WALLET / `PH_GCASH` |
| `PH_MAYA` | Maya | WALLET / `PH_PAYMAYA` |
| `PH_GRABPAY` | GrabPay | WALLET / `PH_GRABPAY` |
| `PH_SHOPEE` | ShopeePay | WALLET / `PH_SHOPEE` |
| `PH_COINS` | Coins.ph | WALLET / `PH_COINS` |

Response body:

```json
{
  "payout_id": "po-...",
  "status": "ACCEPTED",
  "reference_id": "cashout-001",
  "type": "B2C",
  "source_currency": "PHP",
  "source_amount": 1500,
  "destination_currency": "PHP",
  "destination_amount": 1500,
  "failure_code": "",
  "estimated_arrival_time": "2026-07-08T16:24:40Z",
  "created": "...",
  "updated": "..."
}
```

`amount` is in PHP (major units). Request and response amounts are both in PHP major units; the value is converted to minor units (`PHP 1500.00` → `150000`) only when calling Xendit. Xendit routes the payout to the destination channel automatically, using InstaPay for eligible amounts (≤ PHP 50,000) and PESONet otherwise. PH payout channels settle in whole pesos, so any centavos may be dropped by Xendit (e.g. `100.50` is sent as `PHP 100.00`); the response reflects the amount actually sent.

## Testing in sandbox

In Xendit test mode the QR string is a placeholder (`some-random-qr-string`), so the
generated QR is **not scannable** and cannot be paid by a real bank or e-wallet app.
To exercise the full payment flow in sandbox, use the frontend's **Simulate payment**
button (or `POST /v1/cash-in/qr/:id/simulate`) and watch the status via
`GET /v1/cash-in/qr/:id`, which polls until it reaches a terminal status
(`SUCCEEDED`, `FAILED`, `EXPIRED`, or `CANCELED`).

Payouts behave differently in test mode: a `POST /v1/cash-out/payout` with any valid
`account_holder_name` and `account_number` returns a payout that transitions to
`SUCCEEDED` without a simulate call. The frontend's cash-out form polls
`GET /v1/cash-out/payout/:id` until the status is terminal. To simulate failures in
test mode, use Xendit's reserved test account numbers, e.g. `98018521` → `TRANSFER_ERROR`,
`123456` → `TEMPORARY_TRANSFER_ERROR`, `999999` → `REJECTED_BY_CHANNEL`,
`121212` → `INVALID_DESTINATION`.

## Project Structure

```
.
├── xendit/          Go backend (Gin + Xendit client)
│   ├── main.go
│   └── models/xendit.go
└── web/             Static frontend (HTML/CSS/JS)
```
