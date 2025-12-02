package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/handlers"
	"rental-saas/internal/middleware"

	"github.com/jackc/pgx/v5"
)

func TestBookingRaceCondition(t *testing.T) {
	// Setup
	if err := database.Connect(); err != nil {
		t.Skip("Skipping test: Database not available")
	}
	// defer database.DB.Close() // Keep open for test duration

	tenantID := "test_race_tenant"
	
	// Create Test Data
	var carID, customerID string
	err := database.RunInTenantScope(context.Background(), tenantID, func(tx pgx.Tx) error {
		// Create Car
		err := tx.QueryRow(context.Background(), 
			"INSERT INTO cars (make, model, year, status, daily_rate_cents) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			"Race", "Car", 2024, "available", 10000).Scan(&carID)
		if err != nil {
			return err
		}
		// Create Customer
		err = tx.QueryRow(context.Background(),
			"INSERT INTO customers (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING id",
			fmt.Sprintf("racer-%d@example.com", time.Now().UnixNano()), "Speed", "Racer").Scan(&customerID)
		return err
	})
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	// Execute Race
	workers := 10
	var wg sync.WaitGroup
	results := make(chan int, workers)

	handler := handlers.NewBookingHandler()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			reqBody := handlers.CreateBookingRequest{
				CarID:      carID,
				CustomerID: customerID,
				StartTime:  time.Now(),
				EndTime:    time.Now().Add(24 * time.Hour),
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/bookings", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), middleware.TenantKey, tenantID))
			w := httptest.NewRecorder()

			handler.CreateBooking(w, req)
			results <- w.Code
		}()
	}

	wg.Wait()
	close(results)

	// Verify Results
	successCount := 0
	conflictCount := 0
	otherCount := 0

	for code := range results {
		if code == http.StatusCreated {
			successCount++
		} else if code == http.StatusConflict {
			conflictCount++
		} else {
			otherCount++
		}
	}

	t.Logf("Race Results: Success=%d, Conflict=%d, Other=%d", successCount, conflictCount, otherCount)

	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	if conflictCount != workers-1 {
		t.Errorf("Expected %d conflicts, got %d", workers-1, conflictCount)
	}
}
