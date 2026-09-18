// migrations_disabled_binding_test.go — card t795 measurement.
//
// system.yaml's `migrations.disabled` has a SystemConfig field
// (SystemConfig.Migrations) and a real consumer (internal/hook/session_start.go
// runMigration), but systemFileWrapper binds only the `hook` block, so nothing
// carries the YAML value onto the field. This file MEASURES that, across the
// three control cases card t795 requires: disabled: true, disabled: false, and
// the key absent.
//
// The tests log every loaded value so the measurement is readable even when an
// assertion changes, and assert the user-visible contract: a key the user sets
// is honoured. While the binding is missing, the `true` case FAILS — that
// failure IS the reproduction.
//
// Sentinel on failure: SYSTEM_MIGRATIONS_DISABLED_UNBOUND
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeMigrationsFixture builds an isolated project whose
// .moai/config/sections/system.yaml holds body, and returns the path to pass to
// Loader.Load (the `.moai` dir, since Load appends config/sections itself).
func writeMigrationsFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir fixture sections: %v", err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(sections, "system.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("write fixture system.yaml: %v", err)
		}
	}
	return filepath.Join(root, ".moai")
}

// loadMigrationsDisabled loads the fixture and returns the field the consumer reads.
func loadMigrationsDisabled(t *testing.T, body string) bool {
	t.Helper()
	cfg, err := NewLoader().Load(writeMigrationsFixture(t, body))
	if err != nil {
		t.Fatalf("load fixture config: %v", err)
	}
	if cfg == nil {
		t.Fatal("load returned a nil config")
	}
	return cfg.System.Migrations.Disabled
}

const migrationsDisabledTrue = "migrations:\n  disabled: true\n"
const migrationsDisabledFalse = "migrations:\n  disabled: false\n"

// Control 1 — the key the user sets to turn migrations OFF.
func TestSystemMigrations_DisabledTrueIsHonoured(t *testing.T) {
	got := loadMigrationsDisabled(t, migrationsDisabledTrue)
	t.Logf("system.yaml `migrations.disabled: true` → cfg.System.Migrations.Disabled = %v", got)
	if !got {
		t.Errorf("SYSTEM_MIGRATIONS_DISABLED_UNBOUND: the user set migrations.disabled: true, "+
			"but the loaded config reports %v — the edit is ignored with no signal, and "+
			"internal/hook/session_start.go runMigration runs the migrations anyway", got)
	}
}

// Control 2 — explicit false must stay false (migrations run).
func TestSystemMigrations_DisabledFalseKeepsMigrationsOn(t *testing.T) {
	got := loadMigrationsDisabled(t, migrationsDisabledFalse)
	t.Logf("system.yaml `migrations.disabled: false` → cfg.System.Migrations.Disabled = %v", got)
	if got {
		t.Errorf("migrations.disabled: false must leave migrations enabled, got Disabled=%v", got)
	}
}

// Control 3 — the default when the key is absent. This is the shipped shape:
// the template's system.yaml carries moai / github / hook / document_management
// and no `migrations` block at all, so this case is what every distributed
// project actually loads. Migrations ON is the intended default; the assertion
// pins it so a later binding cannot flip the default silently.
func TestSystemMigrations_AbsentKeyDefaultsToEnabled(t *testing.T) {
	got := loadMigrationsDisabled(t, "hook:\n  strict_mode: false\n")
	t.Logf("system.yaml without a `migrations` block → cfg.System.Migrations.Disabled = %v", got)
	if got {
		t.Errorf("an absent migrations block must default to migrations ENABLED (Disabled=false), got %v", got)
	}
}

// Control 3b — no system.yaml at all takes the same default, so "file absent"
// and "key absent" cannot diverge.
func TestSystemMigrations_NoSystemFileDefaultsToEnabled(t *testing.T) {
	got := loadMigrationsDisabled(t, "")
	t.Logf("no system.yaml → cfg.System.Migrations.Disabled = %v", got)
	if got {
		t.Errorf("an absent system.yaml must default to migrations ENABLED (Disabled=false), got %v", got)
	}
}
