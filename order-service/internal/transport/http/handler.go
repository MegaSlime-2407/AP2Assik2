package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/orders", h.handleOrders)
	mux.HandleFunc("/orders/", h.handleOrderByID)
}

func (h *OrderHandler) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getAllOrders(w, r)
		return
	}
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		CustomerID string `json:"customer_id"`
		ItemName   string `json:"item_name"`
		Amount     int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.CustomerID == "" || req.ItemName == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "customer_id and item_name are required"})
		return
	}

	order, err := h.uc.CreateOrder(r.Context(), usecase.CreateOrderInput{
		CustomerID: req.CustomerID,
		ItemName:   req.ItemName,
		Amount:     req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrPaymentFailed) {
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "payment service unavailable"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "something went wrong"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"id":     order.ID,
		"status": order.Status,
	})
}

func (h *OrderHandler) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	
	if strings.HasSuffix(path, "/task") {
		if r.Method != http.MethodGet {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		customerID := strings.TrimSuffix(path, "/task")
		h.task(w, r, customerID)
		return
	}

	if strings.HasSuffix(path, "/cancel") {
		if r.Method != http.MethodPatch {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id := strings.TrimSuffix(path, "/cancel")
		h.cancelOrder(w, r, id)
		return
	}

	if r.Method == http.MethodDelete {
		h.deleteOrder(w, r, path)
		return
	}

	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	h.getOrder(w, r, path)
}

type orderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func (h *OrderHandler) getAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.uc.GetAllOrders(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	var result []orderResponse
	for _, o := range orders {
		result = append(result, orderResponse{
			ID:         o.ID,
			CustomerID: o.CustomerID,
			ItemName:   o.ItemName,
			Amount:     o.Amount,
			Status:     o.Status,
			CreatedAt:  o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	if result == nil {
		result = []orderResponse{}
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request, id string) {
	order, err := h.uc.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	respondJSON(w, http.StatusOK, orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request, id string) {
	order, err := h.uc.CancelOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCancelNotAllowed) {
			respondJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	respondJSON(w, http.StatusOK, orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *OrderHandler) deleteOrder(w http.ResponseWriter, r *http.Request, id string) {
	err := h.uc.DeleteOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "order deleted"})
}

func respondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *OrderHandler) task(w http.ResponseWriter, r *http.Request, customerID string) {
	result, err := h.uc.Task(r.Context(), customerID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"customer_id":  customerID,
		"total_amount": result.TotalAmount,
		"total_orders": result.TotalOrders,
	})
}