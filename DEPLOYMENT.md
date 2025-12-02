# 🚀 Production Flight Manual

This guide covers the deployment of the Car Rental SaaS to a fresh Ubuntu 22.04 LTS server.

## 1. Prerequisites (Run on Server)

Install Docker, Compose, and Make:
```bash
# Update System
sudo apt update && sudo apt upgrade -y

# Install Docker
sudo apt install -y docker.io make git
sudo systemctl enable --now docker
sudo usermod -aG docker $USER

# Install Docker Compose (V2)
mkdir -p ~/.docker/cli-plugins/
curl -SL https://github.com/docker/compose/releases/download/v2.24.6/docker-compose-linux-x86_64 -o ~/.docker/cli-plugins/docker-compose
chmod +x ~/.docker/cli-plugins/docker-compose
```

## 2. Network & Security

### Configure the Firewall (UFW)
```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### DNS Configuration
Point your domain's A Records to this server's IP address:
- `@` (root) -> `123.45.67.89`
- `*` (wildcard) -> `123.45.67.89`

## 3. Deployment

### Clone the Repository
```bash
git clone https://github.com/YOUR_USERNAME/rental-saas-platform.git
cd rental-saas-platform
```

### Configure Environment
Create the production environment file:
```bash
cp .env.example .env
nano .env
```
Update `DOMAIN_NAME`, `STRIPE_SECRET_KEY`, and `JWT_SECRET`.

### Launch
```bash
make deploy
```
*This will build the optimized images and start Traefik.*

## 4. Initialization (First Run Only)
Since the database is empty, create the Admin Tenant:
```bash
./scripts/init_prod.sh
```

## 5. Maintenance
- **View Logs:** `make logs`
- **Update Code:** `git pull && make deploy`
- **Backup DB:** `make backup`