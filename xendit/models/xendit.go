package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type Xendit struct {
	SecretKey        string
	BaseURL          string
	APIVersion       string
	PayoutAPIVersion string
	HTTPClient       *http.Client
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

type CashoutRequest struct {
	ReferenceID       string  `json:"reference_id"`
	Amount            float64 `json:"amount"`
	Channel           string  `json:"channel"`
	AccountHolderName string  `json:"account_holder_name"`
	AccountNumber     string  `json:"account_number"`
}

type PayoutResponse struct {
	PayoutID             string  `json:"payout_id"`
	Status               string  `json:"status"`
	ReferenceID          string  `json:"reference_id"`
	Type                 string  `json:"type"`
	SourceCurrency       string  `json:"source_currency"`
	SourceAmount         float64 `json:"source_amount"`
	DestinationCurrency  string  `json:"destination_currency"`
	DestinationAmount    float64 `json:"destination_amount"`
	FailureCode          string  `json:"failure_code"`
	EstimatedArrivalTime string  `json:"estimated_arrival_time"`
	Created              string  `json:"created"`
	Updated              string  `json:"updated"`
}

type payoutChannel struct {
	RoutingType  string
	RoutingValue string
}

var payoutChannels = map[string]payoutChannel{
	"PH_BDO":     {"SWIFT", "BNORPHMM"},
	"PH_BPI":     {"SWIFT", "BOPIPHMM"},
	"PH_MBT":     {"SWIFT", "MBTCPHMM"},
	"PH_UBP":     {"SWIFT", "UBPHPHMM"},
	"PH_LBP":     {"SWIFT", "TLBPPHMM"},
	"PH_SEC":     {"SWIFT", "SETCPHMM"},
	"PH_RCBC":    {"SWIFT", "RCBCPHMM"},
	"PH_PNB":     {"SWIFT", "PNBMPHMM"},
	"PH_CBC":     {"SWIFT", "CHBKPHMM"},
	"PH_DBP":     {"SWIFT", "DBPHPHMM"},
	"PH_GCASH":   {"WALLET", "PH_GCASH"},
	"PH_MAYA":    {"WALLET", "PH_PAYMAYA"},
	"PH_GRABPAY": {"WALLET", "PH_GRABPAY"},
	"PH_SHOPEE":  {"WALLET", "PH_SHOPEE"},
	"PH_COINS":   {"WALLET", "PH_COINS"},
}

type payoutPayload struct {
	ReferenceID   string          `json:"reference_id"`
	Recipient     payoutRecipient `json:"recipient"`
	PayoutDetails payoutDetails   `json:"payout_details"`
	SourceOfFund  string          `json:"source_of_fund"`
	PurposeCode   string          `json:"purpose_code"`
}

type payoutRecipient struct {
	Type           string                  `json:"type"`
	GivenName      string                  `json:"given_name"`
	Surname        string                  `json:"surname"`
	Relationship   string                  `json:"relationship"`
	Details        *payoutRecipientDetails `json:"details,omitempty"`
	AccountDetails payoutAccountDetails    `json:"account_details"`
	Address        payoutAddress           `json:"address"`
}

type payoutRecipientDetails struct {
	PersonalMobileNumber string `json:"personal_mobile_number"`
}

type payoutAccountDetails struct {
	Currency          string `json:"currency"`
	AccountCountry    string `json:"account_country"`
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	RoutingType1      string `json:"routing_type_1"`
	RoutingValue1     string `json:"routing_value_1"`
}

type payoutAddress struct {
	Country       string `json:"country"`
	ProvinceState string `json:"province_state"`
	City          string `json:"city"`
	StreetLine1   string `json:"street_line_1"`
	PostalCode    string `json:"postal_code"`
}

type payoutDetails struct {
	SourceCurrency      string `json:"source_currency"`
	SourceAmount        int64  `json:"source_amount"`
	DestinationCurrency string `json:"destination_currency"`
}

func NewXendit() *Xendit {
	return &Xendit{
		SecretKey:        os.Getenv("XENDIT_SECRET_KEY"),
		BaseURL:          os.Getenv("XENDIT_BASE_URL"),
		APIVersion:       os.Getenv("XENDIT_API_VERSION"),
		PayoutAPIVersion: os.Getenv("XENDIT_PAYOUT_API_VERSION"),
		HTTPClient:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (x *Xendit) doJSON(method, path, apiVersion string, headers map[string]string, payload any, out any) error {
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
	httpReq.Header.Set("api-version", apiVersion)
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
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
			ErrorCode string          `json:"error_code"`
			Message   string          `json:"message"`
			Errors    json.RawMessage `json:"errors"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			if len(apiErr.Errors) > 0 {
				return fmt.Errorf("xendit error: status %d: %s: %s: %s", resp.StatusCode, apiErr.ErrorCode, apiErr.Message, string(apiErr.Errors))
			}
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
	if err := x.doJSON(http.MethodPost, "/v3/payment_requests", x.APIVersion, nil, payload, &result); err != nil {
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
	if err := x.doJSON(http.MethodPost, "/v3/payment_requests/"+paymentRequestID+"/simulate", x.APIVersion, nil, payload, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (x *Xendit) GetCashinQR(paymentRequestID string) (*CashinQRResponse, error) {
	var result CashinQRResponse
	if err := x.doJSON(http.MethodGet, "/v3/payment_requests/"+paymentRequestID, x.APIVersion, nil, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (x *Xendit) CreateCashout(req CashoutRequest) (*PayoutResponse, error) {
	channel, ok := payoutChannels[req.Channel]
	if !ok {
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}

	nameParts := strings.Fields(req.AccountHolderName)
	givenName := ""
	surname := ""
	if len(nameParts) > 0 {
		givenName = nameParts[0]
		surname = strings.Join(nameParts[1:], " ")
		if surname == "" {
			surname = givenName
		}
	}

	recipient := payoutRecipient{
		Type:         "INDIVIDUAL",
		GivenName:    givenName,
		Surname:      surname,
		Relationship: "CUSTOMER",
		AccountDetails: payoutAccountDetails{
			Currency:          "PHP",
			AccountCountry:    "PH",
			AccountHolderName: req.AccountHolderName,
			AccountNumber:     req.AccountNumber,
			RoutingType1:      channel.RoutingType,
			RoutingValue1:     channel.RoutingValue,
		},
		Address: payoutAddress{
			Country:       "PH",
			ProvinceState: "Metro Manila",
			City:          "Manila",
			StreetLine1:   "123 Rizal Avenue",
			PostalCode:    "1000",
		},
	}
	if channel.RoutingType == "WALLET" {
		recipient.Details = &payoutRecipientDetails{PersonalMobileNumber: req.AccountNumber}
	}

	payload := payoutPayload{
		ReferenceID: req.ReferenceID,
		Recipient:   recipient,
		PayoutDetails: payoutDetails{
			SourceCurrency:      "PHP",
			SourceAmount:        int64(math.Round(req.Amount * 100)),
			DestinationCurrency: "PHP",
		},
		SourceOfFund: "BUSINESS_REVENUE",
		PurposeCode:  "SELF",
	}

	idempotencyKey := req.ReferenceID
	if len(idempotencyKey) < 10 {
		idempotencyKey = "xendit-payout-" + idempotencyKey
	}

	headers := map[string]string{"Idempotency-key": idempotencyKey}

	var result PayoutResponse
	if err := x.doJSON(http.MethodPost, "/v3/payouts", x.PayoutAPIVersion, headers, payload, &result); err != nil {
		return nil, err
	}
	result.SourceAmount /= 100
	result.DestinationAmount /= 100
	if result.SourceAmount == 0 && result.SourceCurrency == result.DestinationCurrency {
		result.SourceAmount = result.DestinationAmount
	}

	return &result, nil
}

func (x *Xendit) GetPayout(payoutID string) (*PayoutResponse, error) {
	var result PayoutResponse
	if err := x.doJSON(http.MethodGet, "/v3/payouts/"+payoutID, x.PayoutAPIVersion, nil, nil, &result); err != nil {
		return nil, err
	}
	result.SourceAmount /= 100
	result.DestinationAmount /= 100
	if result.SourceAmount == 0 && result.SourceCurrency == result.DestinationCurrency {
		result.SourceAmount = result.DestinationAmount
	}

	return &result, nil
}
