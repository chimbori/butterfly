package db

import (
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

var Pool *pgxpool.Pool
