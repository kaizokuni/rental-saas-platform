package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &PaymentHandler{}
}

type CreateIntentRequest struct {
	BookingID string `json:"booking_id"`
	Amount    int64  `json:"amount"` // Amount in cents
}

func (h *PaymentHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	var req CreateIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	// Check if payment intent already exists for this booking
	var existingIntentID string
	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), "SELECT stripe_intent_id FROM payments WHERE booking_id = $1", req.BookingID).Scan(&existingIntentID)
	})
	if err == nil && existingIntentID != "" {
		// Return existing intent
		// We need to fetch the client secret from Stripe if we don't store it (we don't seem to store it in DB based on previous code, only intent ID)
		// Or we can retrieve the intent from Stripe.
		pi, err := paymentintent.Get(existingIntentID, nil)
		if err != nil {
			http.Error(w, "Failed to retrieve existing payment intent: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"client_secret": pi.ClientSecret,
		})
		return
	}

	// Create Stripe PaymentIntent (Manual Capture)
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(req.Amount),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		CaptureMethod: stripe.String("manual"), // Two-step payment
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"tenant_id":  tenantID,
			"booking_id": req.BookingID,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		http.Error(w, "Failed to create payment intent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store intent in database
	err = database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(),
			"INSERT INTO payments (booking_id, stripe_intent_id, amount_cents, status) VALUES ($1, $2, $3, $4)",
			req.BookingID, pi.ID, req.Amount, "pending_auth")
		return err
	})

	if err != nil {
		// Note: In a real app, we might want to cancel the intent here if DB save fails
		http.Error(w, "Failed to save payment record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"client_secret": pi.ClientSecret,
	})
}

func (h *PaymentHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Type == "payment_intent.amount_capturable_updated" {
		var pi stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tenantID := pi.Metadata["tenant_id"]
		if tenantID == "" {
			fmt.Println("Missing tenant_id in metadata")
			w.WriteHeader(http.StatusOK) // Acknowledge receipt even if we can't process
			return
		}

		// Update status in DB
		err = database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(r.Context(),
				"UPDATE payments SET status = 'authorized' WHERE stripe_intent_id = $1",
				pi.ID)
			return err
		})

		if err != nil {
			fmt.Printf("Failed to update payment status: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("Payment authorized for intent %s (Tenant: %s)\n", pi.ID, tenantID)
	}

	w.WriteHeader(http.StatusOK)
}
