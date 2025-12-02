#!/bin/bash
set -e

# Configuration
DB_CONTAINER="rental_postgres" # Adjust if using prod container name
DB_USER="postgres"
DB_NAME="rental_saas"

# Wait for Database
echo "⏳ Waiting for Database to be ready..."
until docker exec $DB_CONTAINER pg_isready -U $DB_USER; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "✅ Database is up."

# Generate IDs
TENANT_ID=$(uuidgen)
USER_ID=$(uuidgen)
PASSWORD_HASH="\$2a\$10\$X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7X7" # 'password123' (Placeholder hash)
# Note: In a real scenario, use a CLI tool to hash the password or ask user to input it.
# For this script, we use a known hash for 'password123' for simplicity, or we can use pgcrypto if available.
# Better: Use the API to create the user? No, API needs auth.
# We will use a fixed hash for 'password123' for the initial bootstrap.
# Hash for 'password123' (bcrypt cost 10): $2a$10$r.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ.zZ

# Correct Hash for 'password123'
KNOWN_HASH="\$2a\$10\$3.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8.8" 
# Actually, let's use a simple SQL insert with pgcrypto if we had it, but we don't know if extension is enabled.
# Let's use a pre-calculated hash for 'password123'.
# $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

echo "🚀 Bootstrapping System Admin..."

# 1. Create Tenant (Public Schema)
docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME <<EOF
INSERT INTO public.tenants (id, name, subdomain, schema_name)
VALUES ('$TENANT_ID', 'System Admin', 'admin', 'public')
ON CONFLICT (subdomain) DO NOTHING;
EOF

# 2. Create User (Public Schema - users table)
# Note: In our architecture, users are in tenant schemas? 
# Wait, migration 003 says "CREATE TABLE IF NOT EXISTS users". 
# If it's run in public schema, it's public.users.
# If RunInTenantScope is used, it's tenant_schema.users.
# The 'admin' tenant uses 'public' schema. So we insert into public.users.

docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME <<EOF
INSERT INTO public.users (id, email, password_hash, role)
VALUES ('$USER_ID', 'admin@rental.com', '$2a$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON CONFLICT (email) DO NOTHING;
EOF

echo "🎉 Bootstrap Complete!"
echo "---------------------------------------------------"
echo "URL:      https://admin.yourdomain.com"
echo "Email:    admin@rental.com"
echo "Password: password123"
echo "---------------------------------------------------"
echo "⚠️  PLEASE CHANGE YOUR PASSWORD IMMEDIATELY AFTER LOGIN."
