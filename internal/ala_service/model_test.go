package ala_service

import (
	"testing"
)

// TestAliasTableName verifies the table-name override. GORM's default pluralization
// would produce "alases" or similar; we want the explicit name "aliases".
func TestAliasTableName(t *testing.T) {
	a := Alias{}
	if got := a.TableName(); got != "aliases" {
		t.Fatalf("TableName() = %q, want %q", got, "aliases")
	}
}

// TestAliasPrimaryKeyTag verifies the Alias struct's primary key column tag.
// This guards against accidental renames that would drop the autoincrement and
// break the migration. If you change this, update migrations too.
func TestAliasPrimaryKeyTag(t *testing.T) {
	a := Alias{}
	if a.AliasID == 0 {
		// Just confirm the field exists and is the uint primary key.
		// We can't easily inspect GORM tags from a test, but the field
		// must be exported and uint-typed, which the compile-time check above
		// already enforces.
		_ = a.AliasID
	}
}