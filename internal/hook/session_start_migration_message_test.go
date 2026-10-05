package hook_test

// The SystemMessage half of REQ-V3R2-RT-007-021 (card t796).
//
// REQ-021 carries three obligations for a failed migration: do not advance the
// version file, write the failure to .moai/logs/migrations.log, and emit
// HookResponse.SystemMessage naming the migration and its error. The first two
// landed with the runner; this file covers the third, which recorded the
// failure only in HookOutput.Data (json:"-", never serialized) and slog (whose
// stderr the hook wrappers discard) — so the failure was recorded and invisible.
//
// Boundary that shapes these assertions: runMigration holds only the error, and
// the two failure shapes differ in what that error can name.
//
//	pre-flight — readVersion fails ("version 읽기 실패: …"). No migration ran,
//	             so there is no number to name. This is the shape the fixture
//	             below triggers, because the test binary does not link
//	             internal/migration/migrations and the registry is empty.
//	apply      — a registered migration fails ("마이그레이션 %d 적용 실패: …",
//	             runner.go:109). The number is already inside the error text.
//
// The message therefore does not carry its own %d slot: it wraps the error,
// which names the migration where a migration exists. Asserting a literal
// number here would pin a value the pre-flight shape cannot produce.
//
// The Korean error text is deliberately passed through rather than translated:
// rewriting the runner's error surface is a separate change (card filed by the
// lead), and processing a root-cause error to look tidy is worse than a mixed
// sentence.
//
// Sentinel on failure: MIGRATION_FAILURE_NOT_SURFACED
//
// @MX:ANCHOR: [AUTO] the only test proving a migration failure reaches the user
// @MX:REASON: HookOutput.Data is json:"-" and hook stderr is discarded, so every
// other record of the failure is invisible to the session; if this assertion
// goes, the failure goes silent again with nothing turning red.

import (
	"strings"
	"testing"
)

// A failed migration MUST reach the user through SystemMessage.
func TestSessionStart_MigrationFailure_SurfacesViaSystemMessage(t *testing.T) {
	root, _ := newMigrationFixture(t, malformedVersion)
	out, data := runSessionStart(t, nil, root)

	// Premise: the runner actually failed. Without this, an empty SystemMessage
	// would read as "no message needed" rather than "message missing".
	failure, _ := data["migration_error"].(string)
	if failure == "" {
		t.Fatalf("premise: no migration failure occurred, so this test measures nothing (Data=%v)", data)
	}

	if out.SystemMessage == "" {
		t.Fatalf("MIGRATION_FAILURE_NOT_SURFACED: the migration failed (%q) but SystemMessage is empty — "+
			"the failure lives only in Data (json:\"-\") and slog, so the user never sees it", failure)
	}

	// The error text itself must ride along: a bare "a migration failed" tells
	// the user nothing actionable, and for an apply failure this text is what
	// carries the migration number.
	if !strings.Contains(out.SystemMessage, failure) {
		t.Errorf("SystemMessage does not carry the error text.\nwant substring: %q\ngot: %q", failure, out.SystemMessage)
	}

	// The three pointers the lead's wording fixes: what did NOT happen, where
	// the detail is, and which command inspects it.
	for _, want := range []string{
		"migration failed",
		"version file was not advanced",
		".moai/logs/migrations.log",
		"moai doctor --check Migration",
	} {
		if !strings.Contains(out.SystemMessage, want) {
			t.Errorf("SystemMessage is missing %q.\ngot: %q", want, out.SystemMessage)
		}
	}

	// The doctor check name is matched exactly and case-sensitively
	// (internal/cli/doctor.go: `c.name != filterCheck`), and the registered
	// name is "Migration" (doctor.go:218). The lowercase spelling the SPEC's
	// AC-06 suggested would run no check at all, so naming it would be worse
	// than naming nothing.
	if strings.Contains(out.SystemMessage, "--check migration") {
		t.Errorf("SystemMessage names the lowercase check, which matches no check and runs nothing.\ngot: %q", out.SystemMessage)
	}
}

// The control that keeps the assertions above honest: a session whose
// migrations did NOT fail must carry no migration SystemMessage. Without it, a
// handler that unconditionally emitted the text would satisfy every check.
func TestSessionStart_MigrationSuccess_EmitsNoMigrationMessage(t *testing.T) {
	root, _ := newMigrationFixture(t, "")
	out, data := runSessionStart(t, nil, root)

	if _, failed := data["migration_error"]; failed {
		t.Fatalf("premise: this fixture must not fail the runner (Data=%v)", data)
	}
	if strings.Contains(out.SystemMessage, "migration failed") {
		t.Errorf("no migration failed, yet SystemMessage announces one.\ngot: %q", out.SystemMessage)
	}
}
