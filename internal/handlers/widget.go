package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WidgetHandler struct{}

func NewWidgetHandler() *WidgetHandler {
	return &WidgetHandler{}
}

// PublicCar defines the JSON structure for the widget
type PublicCar struct {
	ID         uuid.UUID `json:"id"`
	Make       string    `json:"make"`
	Model      string    `json:"model"`
	Year       int       `json:"year"`
	ImageURL   string    `json:"image_url"`
	PriceCents int       `json:"price_cents"`
}

// GET /api/public/cars?tenant_id=...
func (h *WidgetHandler) GetPublicCars(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Missing tenant_id", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(tenantID); err != nil {
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
		return
	}

	var cars []PublicCar

	// Resolve Schema Name
	var schemaName string
	err := database.DB.QueryRow(r.Context(), "SELECT schema_name FROM public.tenants WHERE id = $1", tenantID).Scan(&schemaName)
	if err != nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	err = database.RunInTenantScope(r.Context(), schemaName, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT id, make, model, year, daily_rate_cents, image_url 
			FROM cars 
			WHERE status = 'available'
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var c PublicCar
			if err := rows.Scan(&c.ID, &c.Make, &c.Model, &c.Year, &c.PriceCents, &c.ImageURL); err != nil {
				return err
			}
			cars = append(cars, c)
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to fetch cars: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   cars,
	})
}

type PublicBookRequest struct {
	CarID         string    `json:"car_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
}

// POST /api/public/book?tenant_id=...
func (h *WidgetHandler) PublicBook(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Missing tenant_id", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(tenantID); err != nil {
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
		return
	}

	var req PublicBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Resolve Schema Name
	var schemaName string
	err := database.DB.QueryRow(r.Context(), "SELECT schema_name FROM public.tenants WHERE id = $1", tenantID).Scan(&schemaName)
	if err != nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	err = database.RunInTenantScope(r.Context(), schemaName, func(tx pgx.Tx) error {
		// 1. Find or Create Customer
		var customerID string
		err := tx.QueryRow(r.Context(), "SELECT id FROM customers WHERE email = $1", req.CustomerEmail).Scan(&customerID)
		if err == pgx.ErrNoRows {
			// Create new customer
			// Split name for simplicity
			firstName := req.CustomerName
			lastName := "" 
			
			err = tx.QueryRow(r.Context(), `
				INSERT INTO customers (email, first_name, last_name) 
				VALUES ($1, $2, $3) 
				RETURNING id
			`, req.CustomerEmail, firstName, lastName).Scan(&customerID)
			if err != nil {
				return fmt.Errorf("failed to create customer: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to lookup customer: %w", err)
		}

		// 2. Pessimistic Locking for Car Availability
		var status models.CarStatus
		err = tx.QueryRow(r.Context(), "SELECT status FROM cars WHERE id = $1 FOR UPDATE", req.CarID).Scan(&status)
		if err != nil {
			return fmt.Errorf("car not found: %w", err)
		}

		if status != models.CarStatusAvailable {
			return fmt.Errorf("car is not available")
		}

		// 3. Create Booking (Pending)
		_, err = tx.Exec(r.Context(), `
			INSERT INTO bookings (car_id, customer_id, start_time, end_time, status)
			VALUES ($1, $2, $3, $4, 'pending')
		`, req.CarID, customerID, req.StartDate, req.EndDate)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Booking request received"})
}
