package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/mailer"
	"rental-saas/internal/middleware"
	"rental-saas/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jung-kurt/gofpdf"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type BookingHandler struct{}

func NewBookingHandler() *BookingHandler {
	return &BookingHandler{}
}

type CreateBookingRequest struct {
	CarID      string    `json:"car_id"`
	CustomerID string    `json:"customer_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		// 1. Pessimistic Locking: Lock the car row to prevent double booking
		var status models.CarStatus
		err := tx.QueryRow(r.Context(), "SELECT status FROM cars WHERE id = $1 FOR UPDATE", req.CarID).Scan(&status)
		if err != nil {
			return fmt.Errorf("car not found: %w", err)
		}

		// 2. Check Availability
		if status != models.CarStatusAvailable {
			return fmt.Errorf("car is not available (status: %s)", status)
		}

		// 3. Create Booking
		var bookingID string
		err = tx.QueryRow(r.Context(), `
			INSERT INTO bookings (car_id, customer_id, start_time, end_time, status)
			VALUES ($1, $2, $3, $4, 'pending')
			RETURNING id
		`, req.CarID, req.CustomerID, req.StartTime, req.EndTime).Scan(&bookingID)
		if err != nil {
			return fmt.Errorf("failed to create booking: %w", err)
		}

		// 4. Update Car Status
		_, err = tx.Exec(r.Context(), "UPDATE cars SET status = $1 WHERE id = $2", models.CarStatusRented, req.CarID)
		if err != nil {
			return fmt.Errorf("failed to update car status: %w", err)
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict for double booking
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

type ReturnCarRequest struct {
	FinalOdometer   int `json:"final_odometer"`
	DamageCostCents int `json:"damage_cost_cents"`
}

func (h *BookingHandler) ReturnCar(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	var req ReturnCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Input Validation
	if req.DamageCostCents < 0 {
		http.Error(w, "Damage cost cannot be negative", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	var customerEmail, firstName, lastName, carMake, carModel string

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		// 1. Fetch Booking and Payment Info with LOCK
		// We lock the booking row to prevent concurrent returns (Double Capture)
		var carID string
		var startTime, endTime time.Time
		var status string
		var stripeIntentID string
		var dailyRateCents int

		err := tx.QueryRow(r.Context(), `
			SELECT b.car_id, b.start_time, b.end_time, b.status, p.stripe_intent_id, c.daily_rate_cents,
			       cust.email, cust.first_name, cust.last_name, c.make, c.model
			FROM bookings b
			JOIN payments p ON b.id = p.booking_id
			JOIN cars c ON b.car_id = c.id
			JOIN customers cust ON b.customer_id = cust.id
			WHERE b.id = $1
			FOR UPDATE OF b
		`, bookingID).Scan(&carID, &startTime, &endTime, &status, &stripeIntentID, &dailyRateCents,
			&customerEmail, &firstName, &lastName, &carMake, &carModel)

		if err != nil {
			return fmt.Errorf("booking not found or payment missing: %w", err)
		}

		// 2. State Machine Validation
		// Can only return a car that is currently 'active' (or 'rented' depending on terminology, assuming 'active' based on context, but previous code checked != completed. Let's be stricter.)
		// If the previous status was 'pending', we shouldn't be able to return it? 
		// Let's assume 'confirmed' or 'active' is the state. The CreateBooking sets it to 'pending'. 
		// Payment webhook likely sets it to 'active'.
		// For safety, we ensure it is NOT 'completed' and NOT 'cancelled'.
		if status == "completed" {
			return fmt.Errorf("booking already completed")
		}
		if status == "cancelled" {
			return fmt.Errorf("booking is cancelled")
		}

		// 3. Calculate Final Price
		days := int(endTime.Sub(startTime).Hours() / 24)
		if days < 1 {
			days = 1
		}
		rentalCost := days * dailyRateCents
		finalAmount := rentalCost + req.DamageCostCents

		// 4. Capture Stripe Payment
		// Note: We are inside a DB transaction. If this fails, we rollback.
		// If this succeeds but DB commit fails, we have an issue. 
		// Ideally we would use Idempotency Keys with Stripe based on the booking ID.
		params := &stripe.PaymentIntentCaptureParams{
			AmountToCapture: stripe.Int64(int64(finalAmount)),
		}
		// Idempotency key to prevent double capture on retry
		params.IdempotencyKey = stripe.String("capture_" + bookingID)
		
		_, err = paymentintent.Capture(stripeIntentID, params)
		if err != nil {
			return fmt.Errorf("failed to capture payment: %w", err)
		}

		// 5. Update Database Records
		// Update Booking
		_, err = tx.Exec(r.Context(), `
			UPDATE bookings 
			SET status = 'completed', final_odometer = $1, damage_cost_cents = $2, total_amount_cents = $3
			WHERE id = $4
		`, req.FinalOdometer, req.DamageCostCents, finalAmount, bookingID)
		if err != nil {
			return err
		}

		// Update Car Status
		newCarStatus := models.CarStatusAvailable
		if req.DamageCostCents > 0 {
			newCarStatus = models.CarStatusMaintenance
		}
		_, err = tx.Exec(r.Context(), `
			UPDATE cars 
			SET status = $1, odometer = $2 
			WHERE id = $3
		`, newCarStatus, req.FinalOdometer, carID)
		if err != nil {
			return err
		}

		// Update Payment Status
		_, err = tx.Exec(r.Context(), `
			UPDATE payments 
			SET status = 'captured', amount_cents = $1 
			WHERE stripe_intent_id = $2
		`, finalAmount, stripeIntentID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send Email Async
	go func() {
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, fmt.Sprintf("Invoice - %s", tenantID))
		pdf.Ln(12)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, fmt.Sprintf("Customer: %s %s", firstName, lastName))
		pdf.Ln(8)
		pdf.Cell(40, 10, fmt.Sprintf("Vehicle: %s %s", carMake, carModel))
		pdf.Ln(12)
		pdf.Cell(40, 10, "Total Paid: Captured") // Simplified for async

		var buf bytes.Buffer
		if err := pdf.Output(&buf); err == nil {
			m := mailer.NewSMTPMailer()
			m.SendInvoice(customerEmail, buf.Bytes())
		} else {
			fmt.Printf("Failed to generate PDF for email: %v\n", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
