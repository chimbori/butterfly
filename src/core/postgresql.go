package core

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Migrations are run using the stdlib [database/sql] driver using the pgx compatibility wrapper,
// not [pgx] directly because Goose does not support [pgx.Conn] or [pgxpool.Pool] natively.
func RunMigrations(dbUrl string, migrationsDir embed.FS) error {
	slog.Info("Running migrations...", "database", dbUrl)
	goose.SetBaseFS(migrationsDir)

	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("Setting goose dialect=pgx failed: %w", err)
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return fmt.Errorf("sql.Open failed: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Error running migrations: %w", err)
	}

	return nil
}
