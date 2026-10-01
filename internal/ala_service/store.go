package ala_service

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenStore connects to the configured database and runs AutoMigrate.
// Driver is selected via DB_DRIVER env var:
//   - "sqlite" (default) — uses DB_PATH (default "urlshort.db") for a file-based SQLite DB
//   - "postgres"         — uses DATABASE_URL for a Postgres connection
//
// SQLite is the default for local dev because it has zero external dependencies.
// Postgres is the production target (auto-migrate is dev-only; real deployments
// should use a real migration tool like golang-migrate or goose).
func OpenStore() (*gorm.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	var (
		db  *gorm.DB
		err error
	)

	switch driver {
	case "postgres":
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "host=localhost user=postgres password=postgres dbname=ala_service port=5432 sslmode=disable"
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		path := os.Getenv("DB_PATH")
		if path == "" {
			path = "urlshort.db"
		}
		db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	default:
		log.Printf("unknown DB_DRIVER %q, falling back to sqlite", driver)
		db, err = gorm.Open(sqlite.Open("urlshort.db"), &gorm.Config{})
	}
	if err != nil {
		return nil, err
	}

	// AutoMigrate creates the table if it doesn't exist, and adds missing columns.
	// It does NOT delete columns or change types — those require manual migrations.
	if err := db.AutoMigrate(&Alias{}); err != nil {
		return nil, err
	}

	log.Printf("connected to %s, aliases table ready", driver)
	return db, nil
}
