package models

import (
	"time"

	"github.com/google/uuid"
)

type CarStatus string

const (
	CarStatusAvailable   CarStatus = "available"
	CarStatusRented      CarStatus = "rented"
	CarStatusInspecting  CarStatus = "inspecting"
	CarStatusMaintenance CarStatus = "maintenance"
)

type Car struct {
	ID             uuid.UUID `json:"id"`
	Make           string    `json:"make"`
	Model          string    `json:"model"`
	LicensePlate   string    `json:"license_plate"`
	Status         CarStatus `json:"status"`
	DailyRateCents int       `json:"daily_rate_cents"`
	ImageURL       string    `json:"image_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
