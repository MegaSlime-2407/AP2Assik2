package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type Handler struct {
	uc *usecase.PaymentUseCase
}

func NewHandler(uc *usecase.PaymentUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/payments", h.postPayment)
	mux.HandleFunc("/payments/", h.getPayment)
}

func (h *Handler) postPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request body"})
		return
	}

	if req.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_id is required"})
		return
	}

	payment, err := h.uc.Authorize(r.Context(), req.OrderID, req.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		log.Printf("authorize error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	code := http.StatusOK
	if payment.Status == "Declined" {
		code = http.StatusPaymentRequired
	}

	writeJSON(w, code, map[string]any{
		"id":             payment.ID,
		"order_id":       payment.OrderID,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"status":         payment.Status,
	})
}

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/payments/")
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing order_id"})
		return
	}

	payment, err := h.uc.GetByOrderID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":             payment.ID,
		"order_id":       payment.OrderID,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"status":         payment.Status,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
