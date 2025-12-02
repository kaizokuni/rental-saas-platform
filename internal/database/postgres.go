package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var DB *pgxpool.Pool

type Logger struct{}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	if msg == "Query" {
		if sql, ok := data["sql"]; ok {
			log.Printf("[SQL] %s args=%v", sql, data["args"])
		}
	} else {
		log.Printf("[%s] %s %v", level, msg, data)
	}
}

func Connect() error {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://postgres:password@localhost:5432/rental_saas"
	}

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	// Enable SQL Tracing if DB_DEBUG is set
	if os.Getenv("DB_DEBUG") == "true" {
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   &Logger{},
			LogLevel: tracelog.LogLevelDebug,
		}
		fmt.Println("🔧 SQL Debug Tracing Enabled")
	}

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %v", err)
	}

	fmt.Println("Connected to PostgreSQL")
	return nil
}

func RunInTenantScope(ctx context.Context, tenantID string, fn func(tx pgx.Tx) error) error {
	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Sanitize tenantID to prevent SQL injection
	ident := pgx.Identifier{tenantID}
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL search_path TO %s", ident.Sanitize()))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
