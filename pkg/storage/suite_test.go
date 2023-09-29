package storage

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/tonkeeper/tonkeeper-twa-api/pkg/storage/migrations"
)

func envOrDefault(name, def string) string {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	return val
}

func createDB(t *testing.T) *pgxpool.Pool {
	connectionStr := envOrDefault("POSTGRES_URI", "postgresql://postgres:postgres@localhost:5432/twatest?sslmode=disable")
	pool, err := pgxpool.Connect(context.TODO(), connectionStr)
	require.Nil(t, err)

	_, err = pool.Exec(context.Background(), "DROP SCHEMA if exists twa CASCADE")
	require.Nil(t, err)
	require.Nil(t, migrations.MigrateDb(connectionStr))
	return pool
}
