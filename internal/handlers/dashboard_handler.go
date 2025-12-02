package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/jackc/pgx/v5"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

type DashboardStats struct {
	RevenueCents   int64            `json:"revenue_cents"`
	UtilizationPct float64          `json:"utilization_pct"`
	ActiveRentals  int              `json:"active_rentals"`
	RecentBookings []RecentBooking  `json:"recent_bookings"`
}

type RecentBooking struct {
	ID        string    `json:"id"`
	CarMake   string    `json:"car_make"`
	CarModel  string    `json:"car_model"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *DashboardHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	stats := DashboardStats{
		RecentBookings: []RecentBooking{},
	}

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		// 1. Revenue (This Month)
		err := tx.QueryRow(r.Context(), `
			SELECT COALESCE(SUM(amount_cents), 0) 
			FROM payments 
			WHERE status = 'captured' 
			AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
		`).Scan(&stats.RevenueCents)
		if err != nil {
			return err
		}

		// 2. Utilization Rate
		var totalCars, rentedCars int
		err = tx.QueryRow(r.Context(), `SELECT COUNT(*) FROM cars`).Scan(&totalCars)
		if err != nil {
			return err
		}
		err = tx.QueryRow(r.Context(), `SELECT COUNT(*) FROM cars WHERE status = 'rented'`).Scan(&rentedCars)
		if err != nil {
			return err
		}

		if totalCars > 0 {
			stats.UtilizationPct = (float64(rentedCars) / float64(totalCars)) * 100
		}
		stats.ActiveRentals = rentedCars

		// 3. Recent Activity
		rows, err := tx.Query(r.Context(), `
			SELECT b.id, c.make, c.model, b.status, b.created_at 
			FROM bookings b
			JOIN cars c ON b.car_id = c.id
			ORDER BY b.created_at DESC 
			LIMIT 5
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rb RecentBooking
			if err := rows.Scan(&rb.ID, &rb.CarMake, &rb.CarModel, &rb.Status, &rb.CreatedAt); err != nil {
				return err
			}
			stats.RecentBookings = append(stats.RecentBookings, rb)
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to fetch dashboard stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
