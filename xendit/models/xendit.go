package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Xendit struct {
	SecretKey  string
	BaseURL    string
	APIVersion string
	HTTPClient *http.Client
}

type CashinQRRequest struct {
	ReferenceID string  `json:"reference_id"`
	Amount      float64 `json:"amount"`
}

type CashinQRResponse struct {
	ID          string         `json:"payment_request_id"`
	ReferenceID string         `json:"reference_id"`
	Type        string         `json:"type"`
	Country     string         `json:"country"`
	Currency    string         `json:"currency"`
	Amount      float64        `json:"request_amount"`
	ChannelCode string         `json:"channel_code"`
	Status      string         `json:"status"`
	QRString    string         `json:"qr_string"`
	Actions     []XenditAction `json:"actions"`
	Created     string         `json:"created"`
	Updated     string         `json:"updated"`
}

type SimulateCashinQRRequest struct {
	Amount float64 `json:"amount"`
}

type SimulateCashinQRResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type XenditAction struct {
	Type       string `json:"type"`
	Descriptor string `json:"descriptor"`
	Value      string `json:"value"`
}

type cashinQRPayload struct {
	ReferenceID   string  `json:"reference_id"`
	Type          string  `json:"type"`
	Country       string  `json:"country"`
	Currency      string  `json:"currency"`
	RequestAmount float64 `json:"request_amount"`
	CaptureMethod string  `json:"capture_method"`
	ChannelCode   string  `json:"channel_code"`
}

func NewXendit() *Xendit {
	return &Xendit{
		SecretKey:  os.Getenv("XENDIT_SECRET_KEY"),
		BaseURL:    os.Getenv("XENDIT_BASE_URL"),
		APIVersion: os.Getenv("XENDIT_API_VERSION"),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (x *Xendit) doJSON(method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequest(method, x.BaseURL+path, body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("api-version", x.APIVersion)
	if payload != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.SetBasicAuth(x.SecretKey, "")

	resp, err := x.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var apiErr struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("xendit error: status %d: %s: %s", resp.StatusCode, apiErr.ErrorCode, apiErr.Message)
		}
		return fmt.Errorf("xendit error: status %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return err
		}
	}

	return nil
}

func (x *Xendit) CashinQR(req CashinQRRequest) (*CashinQRResponse, error) {
	payload := cashinQRPayload{
		ReferenceID:   req.ReferenceID,
		Type:          "PAY",
		Country:       "PH",
		Currency:      "PHP",
		RequestAmount: req.Amount,
		CaptureMethod: "AUTOMATIC",
		ChannelCode:   "QRPH",
	}

	var result CashinQRResponse
	if err := x.doJSON(http.MethodPost, "/v3/payment_requests", payload, &result); err != nil {
		return nil, err
	}

	for _, action := range result.Actions {
		if action.Type == "PRESENT_TO_CUSTOMER" && action.Descriptor == "QR_STRING" {
			result.QRString = action.Value
			break
		}
	}

	return &result, nil
}

func (x *Xendit) SimulateCashinQR(paymentRequestID string, amount float64) (*SimulateCashinQRResponse, error) {
	payload := SimulateCashinQRRequest{Amount: amount}

	var result SimulateCashinQRResponse
	if err := x.doJSON(http.MethodPost, "/v3/payment_requests/"+paymentRequestID+"/simulate", payload, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (x *Xendit) GetCashinQR(paymentRequestID string) (*CashinQRResponse, error) {
	var result CashinQRResponse
	if err := x.doJSON(http.MethodGet, "/v3/payment_requests/"+paymentRequestID, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
