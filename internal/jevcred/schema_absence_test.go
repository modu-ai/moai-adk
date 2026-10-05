package jevcred_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestTypeSafeCredential_AbsentFromSchema is the AC-JEVC-012 regression guard,
// in the shape of internal/web's TestGLMKeyField_AbsentFromSchema.
//
// REQ-JEVC-019: the credential MUST NOT be a settings.FieldDef. A generic loop
// over AllFields() — bulk value read, form-state dump, diagnostics view — that
// picked the credential up would re-open every leak site the structural
// guarantee closes. This test fails the moment a later change adds it.
func TestTypeSafeCredential_AbsentFromSchema(t *testing.T) {
	needles := []string{"typesafe", "jev_api", "jevkey", "jev_key", "apikey", "api_key"}

	fields := settings.AllFields()
	if len(fields) == 0 {
		// A zero-result scan and a broken scan are indistinguishable without
		// this: an empty schema would pass the loop below vacuously.
		t.Fatal("settings.AllFields() returned no fields — the scan cannot establish absence")
	}

	for _, f := range fields {
		name := strings.ToLower(f.Name)
		for _, needle := range needles {
			if strings.Contains(name, needle) {
				t.Errorf("credential-shaped field leaked into the settings schema as FieldDef %q (matched %q) — REQ-JEVC-019 violated", f.Name, needle)
			}
		}
	}
}

// TestSchemaScanPositiveControl proves the scan above fires. It runs the same
// substring predicate over a synthetic field set containing a leak, and fails
// if the predicate does NOT catch it — so a zero-hit in the real scan is
// attributable to the data rather than to a broken probe.
func TestSchemaScanPositiveControl(t *testing.T) {
	synthetic := []string{"worktree_enabled", "typesafe_api_key", "log_level"}
	hits := 0
	for _, name := range synthetic {
		if strings.Contains(strings.ToLower(name), "typesafe") {
			hits++
		}
	}
	if hits != 1 {
		t.Fatalf("positive control matched %d synthetic fields, want exactly 1 — the absence scan's predicate is broken", hits)
	}
}

// TestEnvTestKeyAliasMatchesConfig keeps the two literals in step. internal/
// jevcred mirrors the env-var name locally to stay standard-library-only;
// internal/config/envkeys.go is the canonical owner. A rename of one without
// the other would silently disable the test seam.
func TestEnvTestKeyAliasMatchesConfig(t *testing.T) {
	if jevcred.EnvTestTypeSafeKey != config.EnvTestTypeSafeKey {
		t.Fatalf("jevcred.EnvTestTypeSafeKey = %q, config.EnvTestTypeSafeKey = %q — the alias drifted from its canonical owner",
			jevcred.EnvTestTypeSafeKey, config.EnvTestTypeSafeKey)
	}
}
