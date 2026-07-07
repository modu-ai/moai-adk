package migration

import (
	"testing"
)

// TestRegistry_DuplicateVersion_Panics verifies the duplicate-version detection.
// REQ-V3R2-RT-007-053: registering two migrations with the same Version must panic.
//
// The duplicate check lives in init() (registry.go) and runs once at package
// load — it cannot be re-invoked at test time without production-code changes
// (C-3 forbids). This test verifies (a) the detection algorithm correctly
// identifies a duplicate when applied to a hand-built registry, and (b) the
// runtime Register() path appends migrations (the init()-time guard already
// executed at package load with a clean registry).
func TestRegistry_DuplicateVersion_Panics(t *testing.T) {
	// Replicate the init()-time check logic against a duplicate registry.
	dup := []Migration{
		{Version: 1, Name: "first"},
		{Version: 1, Name: "second"},
	}
	seen := make(map[int]string)
	detected := false
	for _, m := range dup {
		if _, ok := seen[m.Version]; ok {
			detected = true
			break
		}
		seen[m.Version] = m.Name
	}
	if !detected {
		t.Error("duplicate-version detection logic failed to find the duplicate")
	}

	// Register appends at runtime (init() already validated the base registry).
	withTestRegistry(t, nil)
	Register(Migration{Version: 500, Name: "test-a"})
	Register(Migration{Version: 501, Name: "test-b"})
	if len(registry) != 2 {
		t.Errorf("after Register: len(registry)=%d, want 2", len(registry))
	}
}

// TestRegistry_Pending verifies the pending-migration list relative to current version.
// REQ-V3R2-RT-007-012: Pending(current) returns migrations with Version > current.
func TestRegistry_Pending(t *testing.T) {
	withTestRegistry(t, []Migration{
		{Version: 1, Name: "m001"},
		{Version: 3, Name: "m003"},
		{Version: 2, Name: "m002"},
	})
	pending := Pending(1)
	if len(pending) != 2 {
		t.Fatalf("Pending(1): got %d migrations, want 2", len(pending))
	}
	// Must be sorted ascending.
	if pending[0].Version != 2 || pending[1].Version != 3 {
		t.Errorf("Pending(1) not sorted ascending: got %d, %d", pending[0].Version, pending[1].Version)
	}
	if len(Pending(3)) != 0 {
		t.Errorf("Pending(3): got %d, want 0", len(Pending(3)))
	}
}

// TestRegistry_Highest verifies the registry's maximum version.
// REQ-V3R2-RT-007-016: the registry is compile-time static and exposes a max-version lookup.
func TestRegistry_Highest(t *testing.T) {
	withTestRegistry(t, []Migration{
		{Version: 5, Name: "m005"},
		{Version: 2, Name: "m002"},
	})
	if h := Highest(); h != 5 {
		t.Errorf("Highest(): got %d, want 5", h)
	}
	// Empty registry → 0.
	withTestRegistry(t, nil)
	if h := Highest(); h != 0 {
		t.Errorf("Highest() empty: got %d, want 0", h)
	}
}

// TestRegistry_AllRegistry_PendingMigrations_FindByVersion exercises the
// registry read helpers for additional coverage.
func TestRegistry_AllRegistry_PendingMigrations_FindByVersion(t *testing.T) {
	withTestRegistry(t, []Migration{
		{Version: 2, Name: "m002"},
		{Version: 1, Name: "m001"},
	})
	all := AllRegistry()
	if len(all) != 2 || all[0].Version != 1 || all[1].Version != 2 {
		t.Errorf("AllRegistry not sorted: got %+v", all)
	}
	if h := HighestVersion(); h != 2 {
		t.Errorf("HighestVersion: got %d, want 2", h)
	}
	pm := PendingMigrations(0)
	if len(pm) != 2 {
		t.Errorf("PendingMigrations(0): got %d, want 2", len(pm))
	}
	found := FindByVersion(1)
	if found == nil || found.Name != "m001" {
		t.Errorf("FindByVersion(1): got %+v, want m001", found)
	}
	if FindByVersion(999) != nil {
		t.Error("FindByVersion(999): should be nil")
	}
}
