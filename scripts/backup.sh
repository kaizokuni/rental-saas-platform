#!/bin/bash

# Configuration
BACKUP_DIR="./backups"
CONTAINER_NAME="rental_postgres"
DB_USER="postgres"
DATE=$(date +%F_%H-%M-%S)
FILENAME="$BACKUP_DIR/backup_$DATE.sql.gz"

# Ensure backup directory exists
mkdir -p $BACKUP_DIR

# Perform Backup
echo "Starting backup for $CONTAINER_NAME..."
docker exec -t $CONTAINER_NAME pg_dumpall -c -U $DB_USER | gzip > $FILENAME

# Check if backup was successful
if [ $? -eq 0 ]; then
  echo "Backup successful: $FILENAME"
  
  # Optional: Upload to MinIO/S3
  # aws s3 cp $FILENAME s3://my-backup-bucket/
else
  echo "Backup failed!"
  exit 1
fi

# Cleanup old backups (keep last 7 days)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete
