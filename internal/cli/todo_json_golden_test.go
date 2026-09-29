// todo_json_golden_test.go — SPEC-TODO-TRANSITION-STAMPS-001 AC-TST-012:
// the existing `todo list --json` disclosure is byte-identical across the
// transition-stamp change.
//
// The mechanism is the AC's own, and its ordering is load-bearing: the
// golden (testdata/ac_tst_012_golden.json) was captured by rendering the
// checked-in PRE-change seed fixture (testdata/ac_tst_012_seed.sql) through
// `todo list --json` on the pre-change tree and committed as a testdata file
// BEFORE the implementation landed — the ordering rule of
// verification-claim-integrity §2.3. The post-change binary renders the same
// fixture and must reproduce every pre-existing key and value byte for byte;
// the only permitted difference is the new `omitempty` stamp keys, and on
// THIS fixture there are none — the seed predates the columns, so every
// stamp reads NULL and the whole output is byte-identical.
//
// The fixture is materialized with the raw sqlite driver, deliberately NOT
// through the store engine: openBacklogEngine would run the additive
// migration and upgrade the fixture before the render, which would weaken
// the check into exercising nothing but the fresh path. Rendering a genuine
// pre-change database through the pure reader (LoadPure, no DDL) is what
// makes this a real upgrade-path regression test.
package cli

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// updateGolden regenerates the golden file from the current binary's render.
// It exists for the one legitimate use — capturing the golden on the
// pre-change tree before the implementation commit — and for a future
// schema-audited regeneration; a golden produced by a changed binary against
// itself proves nothing and must never be refreshed casually.
var updateGolden = flag.Bool("update-ac-tst-012-golden", false, "rewrite the AC-TST-012 golden from the current render")

// acTST012Fixture materializes the checked-in pre-change seed into the
// fixture queue's engine artifact, exactly where the store would create it.
func acTST012Fixture(t *testing.T, enginePath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(enginePath), 0o700); err != nil {
		t.Fatalf("create queue dir: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join("testdata", "ac_tst_012_seed.sql"))
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	db, err := sql.Open("sqlite", enginePath)
	if err != nil {
		t.Fatalf("open fixture db: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(), string(raw)); err != nil {
		t.Fatalf("materialize seed: %v", err)
	}
}

// TestTodoListJSON_GoldenByteIdentity — AC-TST-012. The post-change render
// of the pre-change fixture is byte-identical to the committed golden.
func TestTodoListJSON_GoldenByteIdentity(t *testing.T) {
	_, store := todoFixture(t)
	enginePath := store.EnginePath()
	acTST012Fixture(t, enginePath)

	stdout, _, err := runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	// stderr is not asserted here: the temp-root guard legitimately speaks on
	// stderr inside t.TempDir fixtures, and stdout is the AC's subject.

	goldenPath := filepath.Join("testdata", "ac_tst_012_golden.json")
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(stdout), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("list --json diverged from the pre-change golden\n--- golden ---\n%s\n--- actual ---\n%s", want, stdout)
	}
}
