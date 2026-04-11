package provider

import (
	"context"
	"database/sql"
	"os"

	"github.com/DATA-DOG/go-txdb"
	"github.com/go-errors/errors"
	_ "github.com/lib/pq"
)

func NewDBProvider(env *EnvProvider) (*sql.DB, error) {
	db, err := sql.Open("postgres", env.databaseURL)
	if err != nil {
		return nil, errors.Errorf("unable to connect to database: %w", err)
	}

	if pingErr := db.PingContext(context.Background()); pingErr != nil {
		return nil, errors.Errorf("unable to ping database: %w", pingErr)
	}

	db.SetMaxOpenConns(env.databaseMaxConns)

	return db, nil
}

func RegisterTestTxDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	txdb.Register("txdb", "postgres", databaseURL)
}

func NewTestDBProvider(_ *EnvProvider) (*sql.DB, error) {
	db, err := sql.Open("txdb", "TestTransactionDB")
	if err != nil {
		return nil, errors.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}
