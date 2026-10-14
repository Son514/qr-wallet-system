# QR Wallet System

Backend service that creates Xendit QRPh cash-in payment requests, plus a static frontend that renders the resulting QR code.

## Requirements

- Go 1.27+
- A Xendit account with the **QRPh** channel activated
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

Request body:

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

## Testing in sandbox

In Xendit test mode the QR string is a placeholder (`some-random-qr-string`), so the
generated QR is **not scannable** and cannot be paid by a real bank or e-wallet app.
To exercise the full payment flow in sandbox, use the frontend's **Simulate payment**
button (or `POST /v1/cash-in/qr/:id/simulate`) and watch the status via
`GET /v1/cash-in/qr/:id`, which polls until it reaches a terminal status
(`SUCCEEDED`, `FAILED`, `EXPIRED`, or `CANCELED`).

## Project Structure

```
.
├── xendit/          Go backend (Gin + Xendit client)
│   ├── main.go
│   └── models/xendit.go
└── web/             Static frontend (HTML/CSS/JS)
```
