package handlers

import (
	"fmt"
	"net/http"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jung-kurt/gofpdf"
)

type InvoiceHandler struct{}

func NewInvoiceHandler() *InvoiceHandler {
	return &InvoiceHandler{}
}

func (h *InvoiceHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok {
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}

	var firstName, lastName, carMake, carModel string
	var startTime, endTime time.Time
	var totalAmountCents int

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `
			SELECT c.first_name, c.last_name, car.make, car.model, b.start_time, b.end_time, b.total_amount_cents
			FROM bookings b
			JOIN customers c ON b.customer_id = c.id
			JOIN cars car ON b.car_id = car.id
			WHERE b.id = $1
		`, bookingID).Scan(&firstName, &lastName, &carMake, &carModel, &startTime, &endTime, &totalAmountCents)
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Booking not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice - %s", tenantID))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Customer: %s %s", firstName, lastName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Vehicle: %s %s", carMake, carModel))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Rental Period: %s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Total: $%.2f", float64(totalAmountCents)/100.0))

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", bookingID))
	
	if err := pdf.Output(w); err != nil {
		// Log error, but header already sent
		fmt.Printf("Error generating PDF: %v\n", err)
	}
}
