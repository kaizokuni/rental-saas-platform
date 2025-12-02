package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/jackc/pgx/v5"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type WebhookResponse struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	SecretKey string   `json:"secret_key"`
	Active    bool     `json:"active"`
}

func (h *SettingsHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" || len(req.Events) == 0 {
		http.Error(w, "URL and Events are required", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	// Generate Secret Key
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "Failed to generate secret", http.StatusInternalServerError)
		return
	}
	secretKey := hex.EncodeToString(bytes)

	var webhook WebhookResponse
	webhook.SecretKey = secretKey

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `
			INSERT INTO webhooks (url, events, secret_key) 
			VALUES ($1, $2, $3) 
			RETURNING id, url, events, active
		`, req.URL, req.Events, secretKey).Scan(&webhook.ID, &webhook.URL, &webhook.Events, &webhook.Active)
	})

	if err != nil {
		http.Error(w, "Failed to register webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

func (h *SettingsHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	var webhooks []WebhookResponse

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `SELECT id, url, events, secret_key, active FROM webhooks`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var w WebhookResponse
			if err := rows.Scan(&w.ID, &w.URL, &w.Events, &w.SecretKey, &w.Active); err != nil {
				return err
			}
			webhooks = append(webhooks, w)
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to list webhooks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}
