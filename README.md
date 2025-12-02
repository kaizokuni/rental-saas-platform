# Multi-Tenant Car Rental SaaS Platform

![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)
![Vue Version](https://img.shields.io/badge/Vue.js-3.0-4FC08D?logo=vue.js&logoColor=white)
![Postgres](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Production-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

> **A high-performance, event-driven SaaS platform engineered for scale.**
> Built with strict tenant isolation, financial-grade integer math, and pessimistic locking for concurrency safety.

---

## 🏗️ Architecture

This project implements a **Schema-per-Tenant** architecture to ensure strict data isolation while sharing a single database instance for resource efficiency.

### Key Technical Features

*   **Strict Multi-Tenancy:**
    *   **Isolation:** Middleware intercepts the subdomain (e.g., `tenant.app.com`) and injects `SET LOCAL search_path` for every transaction.
    *   **Security:** Impossible for Tenant A to query Tenant B's data due to Postgres schema boundaries.

*   **Concurrency Safety (The "Double-Booking" Problem):**
    *   **Solution:** Implements **Pessimistic Locking** (`SELECT ... FOR UPDATE`) during the booking flow.
    *   **Result:** Prevents race conditions even under high load, ensuring a car is never booked twice for the same dates.

*   **Financial Integrity:**
    *   **Zero-Float Math:** All monetary values are stored as `INTEGER` (cents) to prevent floating-point rounding errors.
    *   **Two-Step Payments:** Integrates Stripe for `Authorization` (Deposit Hold) and `Capture` (Final Charge) workflows.

*   **Production Operations:**
    *   **Docker:** Multi-stage builds result in a `<20MB` production image (Distroless).
    *   **CI/CD:** GitHub Actions pipeline for automated testing and build verification.
    *   **Observability:** SQL Debug Tracing enabled for audit logs.

---

## 🛠️ Technology Stack

### Backend (Go 1.22)
*   **Router:** `go-chi` (Lightweight, idiomatic)
*   **Database:** `pgx/v5` (High-performance driver with connection pooling)
*   **Auth:** JWT (Stateless) + Bcrypt
*   **Testing:** Native `testing` package + Race Detector

### Frontend (Vue 3 + Vite)
*   **State:** Pinia
*   **Data Fetching:** TanStack Query (Auto-caching & re-fetching)
*   **UI:** Tailwind CSS + Headless UI
*   **Widget:** Vanilla JS embeddable widget for third-party integration.

### Infrastructure
*   **Database:** PostgreSQL 16
*   **Proxy:** Traefik v3 (Auto-SSL via Let's Encrypt)
*   **Containerization:** Docker Compose

---

## 🚀 Quick Start

### Prerequisites
*   Docker & Docker Compose
*   Go 1.22+ (Optional, for local dev)

### 1. Clone & Launch
```bash
git clone https://github.com/kaizokuni/rental-saas-platform.git
cd rental-saas-platform/rental-saas

# Start the stack (Postgres + API + Frontend)
make dev
```

### 2. Bootstrap the Admin
Since the database is empty, create the first tenant:
```bash
make db-shell
```
*Paste the SQL from `seed_user.sql` to create the `admin` tenant.*

### 3. Access the Dashboard
*   **URL:** `http://admin.localhost:5173`
*   **Email:** `admin@rental.com`
*   **Password:** `password123`

---

## 🧪 Testing

Run the full suite, including concurrency race tests:
```bash
go test -v -race ./...
```

## 📜 License
MIT
