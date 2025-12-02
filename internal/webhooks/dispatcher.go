package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rental-saas/internal/database"

	"github.com/jackc/pgx/v5"
)

type Dispatcher struct{}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Dispatch(ctx context.Context, tenantID string, eventType string, payload interface{}) {
	// Run in a goroutine to not block the main request
	go func() {
		// Create a new context for the background task since the request context might be cancelled
		bgCtx := context.Background()

		err := database.RunInTenantScope(bgCtx, tenantID, func(tx pgx.Tx) error {
			rows, err := tx.Query(bgCtx, `
				SELECT url, secret_key 
				FROM webhooks 
				WHERE active = true AND $1 = ANY(events)
			`, eventType)
			if err != nil {
				return err
			}
			defer rows.Close()

			type WebhookTarget struct {
				URL       string
				SecretKey string
			}
			var targets []WebhookTarget

			for rows.Next() {
				var t WebhookTarget
				if err := rows.Scan(&t.URL, &t.SecretKey); err != nil {
					continue
				}
				targets = append(targets, t)
			}

			// Send to all targets
			for _, target := range targets {
				d.send(target.URL, target.SecretKey, eventType, payload)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error dispatching webhooks for tenant %s: %v\n", tenantID, err)
		}
	}()
}

func (d *Dispatcher) send(url, secret, eventType string, payload interface{}) {
	body, err := json.Marshal(map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	})
	if err != nil {
		fmt.Printf("Error marshaling webhook payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating webhook request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RentalSaaS-Webhook/1.0")

	// HMAC Signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Rental-Signature", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending webhook to %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		fmt.Printf("Webhook to %s failed with status: %d\n", url, resp.StatusCode)
	}
}
