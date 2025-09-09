package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBPool(databaseUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	schemaPath := filepath.Join("db", "schema.sql")
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("could not read schema file at %s: %w", schemaPath, err)
	}

	if _, err := pool.Exec(context.Background(), string(schema)); err != nil {
		pool.Close()
		return nil, fmt.Errorf("could not apply schema: %w", err)
	}

	fmt.Println("Database connection successful and schema applied.")
	return pool, nil
}
