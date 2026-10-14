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

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, x.BaseURL+"/v3/payment_requests", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-version", x.APIVersion)
	httpReq.SetBasicAuth(x.SecretKey, "")

	resp, err := x.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var apiErr struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("xendit error: status %d: %s: %s", resp.StatusCode, apiErr.ErrorCode, apiErr.Message)
		}
		return nil, fmt.Errorf("xendit error: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result CashinQRResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
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
