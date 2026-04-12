package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"order-service/internal/usecase"
)

type PaymentHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentHTTPClient(baseURL string, client *http.Client) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		baseURL: baseURL,
		client:  client,
	}
}

func (c *PaymentHTTPClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResult, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"order_id": orderID,
		"amount":   amount,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		TransactionID string `json:"transaction_id"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &usecase.PaymentResult{
		TransactionID: body.TransactionID,
		Status:        body.Status,
	}, nil
}
