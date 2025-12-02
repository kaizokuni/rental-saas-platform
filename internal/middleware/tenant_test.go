package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/jackc/pgx/v5"
)

// MockDB is a placeholder. In a real scenario, we would use a test container or a dedicated test DB.
// For this audit, we will simulate the tenant isolation failure if RunInTenantScope is bypassed.

func TestTenantIsolation(t *testing.T) {
	// 1. Setup: Assume we have two tenants "tenant_a" and "tenant_b" in the DB.
	// This test requires a running DB connection, which we might not have in this environment.
	// So we will write the logic that WOULD fail if isolation is broken.

	if database.DB == nil {
		t.Skip("Skipping test: Database not initialized")
	}

	// Create a car in Tenant A
	var carID string
	err := database.RunInTenantScope(context.Background(), "tenant_a", func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), 
			"INSERT INTO cars (make, model, year, status, daily_rate_cents) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			"Toyota", "Camry", 2023, "available", 5000).Scan(&carID)
	})
	if err != nil {
		t.Fatalf("Failed to insert car in tenant_a: %v", err)
	}

	// Attempt to read that car from Tenant B
	err = database.RunInTenantScope(context.Background(), "tenant_b", func(tx pgx.Tx) error {
		var id string
		err := tx.QueryRow(context.Background(), "SELECT id FROM cars WHERE id = $1", carID).Scan(&id)
		if err == nil {
			return fmt.Errorf("SECURITY ALERT: Tenant B can see Tenant A's car %s", id)
		}
		return nil // Error is expected (sql.ErrNoRows)
	})

	if err != nil {
		t.Fatalf("Tenant Isolation Failed: %v", err)
	}
}
