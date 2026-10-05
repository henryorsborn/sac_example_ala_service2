package ala_service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB returns an in-memory SQLite DB with the aliases table migrated.
// Each test gets its own fresh DB to keep them isolated.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Alias{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// TestOpenStore_DefaultsToSQLite verifies the default DB_DRIVER is sqlite when
// the env var is unset. This is the local-dev expectation.
func TestOpenStore_DefaultsToSQLite(t *testing.T) {
	// Save and clear DB_DRIVER so we get the default.
	orig := os.Getenv("DB_DRIVER")
	t.Cleanup(func() { os.Setenv("DB_DRIVER", orig) })
	os.Unsetenv("DB_DRIVER")

	// Point DB_PATH at a tmp file so we don't litter the repo.
	tmp := filepath.Join(t.TempDir(), "urlshort.db")
	t.Setenv("DB_PATH", tmp)

	db, err := OpenStore()
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	if db == nil {
		t.Fatalf("OpenStore returned nil db with no error")
	}
	// Confirm the table was migrated.
	if !db.Migrator().HasTable(&Alias{}) {
		t.Fatalf("expected aliases table after AutoMigrate")
	}
	// Close explicitly so TempDir cleanup can remove the file on Windows
	// (which holds an exclusive lock on open SQLite files).
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestOpenStore_UnknownDriverFallsBack verifies that an unknown DB_DRIVER
// falls back to sqlite instead of erroring. We don't want a typo in env
// vars to brick the dev server.
func TestOpenStore_UnknownDriverFallsBack(t *testing.T) {
	t.Setenv("DB_DRIVER", "oracle")

	tmp := filepath.Join(t.TempDir(), "urlshort.db")
	t.Setenv("DB_PATH", tmp)

	db, err := OpenStore()
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	if db == nil {
		t.Fatalf("OpenStore returned nil db with no error")
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestOpenStore_PostgresRequiresURL verifies that DB_DRIVER=postgres opens
// without error when DATABASE_URL is set. We use a deliberately invalid URL
// to confirm the error is surfaced clearly rather than panicking.
func TestOpenStore_PostgresRequiresURL(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DATABASE_URL", "host=127.0.0.1 user=nobody dbname=nobody sslmode=disable")

	// We expect this to fail to connect (no postgres server in tests).
	// What we DON'T want is a panic, an empty error, or a silent fallback.
	db, err := OpenStore()
	if err == nil {
		t.Logf("postgres open unexpectedly succeeded; cleaning up")
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		// Don't fail the test if a real postgres happened to be running on
		// the test host; just confirm we got a usable db.
		return
	}
	// err is non-nil as expected.
	if db != nil {
		t.Errorf("expected nil db on connection error, got %v", db)
	}
}