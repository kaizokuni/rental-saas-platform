# Car Rental SaaS

A production-grade, multi-tenant car rental platform built with Go and Vue.js.

## Go Live (Production Deployment)

### Prerequisites
- Docker & Docker Compose installed.
- A domain name pointing to your server.
- Port 80 and 443 open.
- **Local Development**: Update your `/etc/hosts` or `C:\Windows\System32\drivers\etc\hosts` file:
  ```text
  127.0.0.1 admin.localhost hertz.localhost
  ```

### Deployment Steps

1.  **Configure Environment**:
    Ensure `.env` contains production values (STRIPE keys, JWT secret).

2.  **Build Production Image**:
    ```bash
    docker compose -f docker-compose.prod.yml build
    ```

3.  **Start the Stack**:
    ```bash
    docker compose -f docker-compose.prod.yml up -d
    ```

4.  **Verify Deployment**:
    - Check logs: `docker compose -f docker-compose.prod.yml logs -f`
    - Visit `https://app.yourdomain.com` (configured in docker-compose.prod.yml).

### Maintenance

- **Backup Database**:
    ```bash
    ./scripts/backup.sh
    ```
