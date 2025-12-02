package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantContextKey string

const TenantKey TenantContextKey = "tenant"

func TenantMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			log.Printf("TenantMiddleware: Processing request for host: %s\n", host)
			// Extract subdomain (e.g., tenant.example.com -> tenant)
			// This is a simplified extraction for demonstration
			parts := strings.Split(host, ".")
			var subdomain string
			if len(parts) > 2 {
				subdomain = parts[0]
			} else {
				// For localhost or direct IP, default to 'public' schema directly
				// OR we could look for a 'default' tenant.
				// For now, let's assume 'public' schema for root domain if not found.
				subdomain = "public" 
			}
			
			log.Printf("TenantMiddleware: Extracted subdomain: %s\n", subdomain)

			var tenantSchema string
			// Special case for "public" subdomain or dev environment fallback
			if subdomain == "public" || host == "localhost:8080" || strings.HasPrefix(host, "localhost") {
				tenantSchema = "public"
			} else {
				// Query the database to get the schema_name for this subdomain
				// We use the public schema to query the tenants table
				err := db.QueryRow(r.Context(), "SELECT schema_name FROM public.tenants WHERE subdomain = $1", subdomain).Scan(&tenantSchema)
				if err != nil {
					log.Printf("TenantMiddleware: Tenant not found for subdomain %s: %v\n", subdomain, err)
					http.Error(w, "Tenant not found", http.StatusNotFound)
					return
				}
			}

			log.Printf("TenantMiddleware: Resolved tenant schema: %s\n", tenantSchema)
			w.Header().Set("X-Tenant-Schema", tenantSchema)
			ctx := context.WithValue(r.Context(), TenantKey, tenantSchema)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
