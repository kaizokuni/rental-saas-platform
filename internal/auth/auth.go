package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"rental-saas/internal/database"
	"rental-saas/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-change-me")
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login Debug: Handler started\n")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
	if !ok || tenantID == "" {
		log.Printf("Login Debug: Tenant context missing\n")
		http.Error(w, "Tenant context missing", http.StatusInternalServerError)
		return
	}
	log.Printf("Login Debug: Attempting login for %s in tenant %s\n", req.Email, tenantID)

	var user User
	var passwordHash string

	err := database.RunInTenantScope(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), "SELECT id, email, password_hash, role FROM users WHERE email = $1", req.Email).
			Scan(&user.ID, &user.Email, &passwordHash, &user.Role)
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("Login Debug: User not found in DB\n")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
			return
		}
		log.Printf("Login Debug: Database error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error: " + err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		log.Printf("Login Debug: Password mismatch. Hash: %s, Input: %s\n", passwordHash, req.Password)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password mismatch"})
		return
	}
	log.Printf("Login Debug: Login successful for %s\n", user.Email)

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
		User:  user,
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Security Check: Tenant Isolation
		tenantID, ok := r.Context().Value(middleware.TenantKey).(string)
		if !ok {
			http.Error(w, "Tenant context missing", http.StatusInternalServerError)
			return
		}

		tokenTenantID, ok := claims["tenant_id"].(string)
		if !ok || tokenTenantID != tenantID {
			http.Error(w, "Token tenant mismatch (Cross-Tenant Replay Attempt)", http.StatusForbidden)
			return
		}

		// Inject user_id into context
		ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
