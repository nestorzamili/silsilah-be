//go:build integration
// +build integration

package integration_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

const (
	defaultDBURL = "postgres://user:password@localhost:5432/silsilah_db?sslmode=disable"
)

type TestEnv struct {
	DB *sql.DB
}

func SetupTestEnv(t *testing.T) *TestEnv {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultDBURL
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	// Wait for DB to be ready
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err, "Database not ready")

	// Cleanup data (optional, be careful in production)
	_, err = db.Exec("TRUNCATE TABLE users, persons, relationships, audit_logs CASCADE")
	require.NoError(t, err)

	return &TestEnv{
		DB: db,
	}
}

func (e *TestEnv) Teardown() {
	if e.DB != nil {
		e.DB.Close()
	}
}
