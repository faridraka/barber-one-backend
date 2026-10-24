package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// MidtransService handles Midtrans payment gateway integration
type MidtransService struct {
	serverKey string
	baseURL   string
	client    *http.Client
}

func NewMidtransService() (*MidtransService, error) {
	// Try to load from config first, then fallback to file
	key := getEnvValue("MIDTRANS_SERVER_KEY", "")
	if key == "" {
		return nil, fmt.Errorf("MIDTRANS_SERVER_KEY belum dikonfigurasi")
	}
	
	baseURL := "https://app.sandbox.midtrans.com/snap/v1/transactions"
	if strings.EqualFold(strings.TrimSpace(getEnvValue("MIDTRANS_ENVIRONMENT", "sandbox")), "production") {
		baseURL = "https://app.midtrans.com/snap/v1/transactions"
	}
	return &MidtransService{serverKey: key, baseURL: baseURL, client: &http.Client{Timeout: 15 * time.Second}}, nil
}

func getEnvValue(key, fallback string) string {
	// First try from loaded env values
	if value, exists := envValues[key]; exists && value != "" {
		return strings.TrimSpace(value)
	}
	// Then try from os.Getenv
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (m *MidtransService) CreateSnap(ctx context.Context, orderID, name, phone string, amount float64, service, finishURL string, paymentStartedAt time.Time) (map[string]string, error) {
	payload := map[string]any{
		"transaction_details": map[string]any{"order_id": orderID, "gross_amount": int64(amount)},
		"customer_details":    map[string]string{"first_name": name, "phone": phone},
		"item_details":        []map[string]any{{"id": "booking", "price": int64(amount), "quantity": 1, "name": service}},
		"callbacks":           map[string]string{"finish": finishURL},
		"expiry": map[string]any{
			"start_time": paymentStartedAt.In(time.FixedZone("WIB", 7*60*60)).Format("2006-01-02 15:04:05 -0700"),
			"duration":   int(5),
			"unit":       "minutes",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(m.serverKey, "")
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Midtrans returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result["token"] == "" || result["redirect_url"] == "" {
		return nil, fmt.Errorf("Midtrans response missing token or redirect URL")
	}
	return result, nil
}

func (m *MidtransService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + m.serverKey))
	return strings.EqualFold(signatureKey, hex.EncodeToString(sum[:]))
}

var envValues map[string]string
