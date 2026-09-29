// todo_add_dash_test.go — SPEC-TODO-SURFACE-POLISH-001 M1 (card t1349): the
// `todo add` argument surface.
//
// AC-TSP-010 pins the measured defect: a card body starting with `-f` was
// consumed by pflag's interspersed parsing (no `--force` shorthand is
// registered) and the add failed with "unknown shorthand flag". The body is
// TEXT and must reach the store verbatim.
//
// The remaining tests are the AC-TSP-011 regression contract: the known
// flags (--pick, --force, --classification-file), the `--` separator, and
// the parent fallthrough (t69) / mistyped-verb guard (t203) behave exactly
// as before the add path took over its own argument scanning.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// AC-TSP-010 — `moai todo add "-f hello world"` creates the card with the
// verbatim body (RED on the pre-fix tree: pflag rejects the body as an
// unknown shorthand flag).
func TestTodoAddLeadingDashText(t *testing.T) {
	_, store := todoFixture(t)
	const text = "-f hello world"

	out, _, err := runTodo(t, "add", text)
	if err != nil {
		t.Fatalf("add %q: %v", text, err)
	}
	if got := strings.TrimSpace(out); got != "t1 1" {
		t.Errorf("add output = %q, want %q", got, "t1 1")
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("record holds %d items, want 1", len(rec.Items))
	}
	if got := rec.Items[0].Text; got != text {
		t.Errorf("stored text = %q, want the verbatim body %q", got, text)
	}
}

// AC-TSP-011 — `add --pick <text>` still appends AND picks in one locked
// write, printing the picked confirmation.
func TestTodoAddPick(t *testing.T) {
	_, store := todoFixture(t)

	out, _, err := runTodo(t, "add", "--pick", "pick me")
	if err != nil {
		t.Fatalf("add --pick: %v", err)
	}
	if !strings.HasPrefix(out, "picked t1 ") {
		t.Errorf("add --pick output = %q, want the picked confirmation prefix", out)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 || rec.Items[0].State != kanban.BacklogStatePicked {
		t.Errorf("record = %+v, want one picked item", rec.Items)
	}
}

// AC-TSP-011 — `add --force <dup>` still admits an analyser-exact duplicate
// and records the forced admission.
func TestTodoAddForce(t *testing.T) {
	_, store := todoFixture(t)
	const text = "duplicate candidate card"

	if _, _, err := runTodo(t, "add", text); err != nil {
		t.Fatalf("add first: %v", err)
	}
	out, _, err := runTodo(t, "add", "--force", text)
	if err != nil {
		t.Fatalf("add --force duplicate: %v", err)
	}
	if !strings.HasPrefix(out, "t2 ") {
		t.Errorf("add --force output = %q, want the issued id line for t2", out)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("record holds %d items, want 2 (the forced duplicate admitted)", len(rec.Items))
	}
}

// AC-TSP-011 — `add --classification-file <path> <text>` still validates the
// classification BEFORE the locked write and records it on the card.
func TestTodoAddClassificationFile(t *testing.T) {
	_, store := todoFixture(t)
	path := filepath.Join(t.TempDir(), "class.json")
	// The closed value sets the file validates against — an out-of-set value
	// is a usage refusal with nothing written, so a valid judgement keeps
	// the add succeeding exactly as before.
	body := `{"priority":"low","blocked":false,"mode":"serial","decider":"human","reason":"t1349 fixture"}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write classification file: %v", err)
	}

	out, _, err := runTodo(t, "add", "--classification-file", path, "classified card")
	if err != nil {
		t.Fatalf("add --classification-file: %v", err)
	}
	if !strings.HasPrefix(out, "t1 ") {
		t.Errorf("add --classification-file output = %q, want the issued id line", out)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("record holds %d items, want 1", len(rec.Items))
	}
}

// AC-TSP-011 — the `--` separator still ends flag scanning, and the parent
// fallthrough (t69) plus the mistyped-verb guard (t203) are untouched.
func TestTodoBareFallthrough(t *testing.T) {
	_, store := todoFixture(t)

	// t69: multi-word natural language falls through to add.
	out, _, err := runTodo(t, "fix", "the flaky queue tests")
	if err != nil {
		t.Fatalf("bare fallthrough: %v", err)
	}
	if !strings.HasPrefix(out, "t1 ") {
		t.Errorf("bare fallthrough output = %q, want the issued id line", out)
	}

	// `--` still escapes a dash-leading body through the explicit add verb.
	out, _, err = runTodo(t, "add", "--", "-not-a-flag body")
	if err != nil {
		t.Fatalf("add -- dash body: %v", err)
	}
	if !strings.HasPrefix(out, "t2 ") {
		t.Errorf("add -- output = %q, want the issued id line", out)
	}

	// t203: verb-shaped first token addressing a card id stays refused.
	_, errOut, err := runTodo(t, "pick", "t1")
	if err == nil {
		t.Fatalf("mistyped verb pick t1: want the guard refusal, got nil (stdout context)")
	}
	if !strings.Contains(errOut, "is not a todo verb") {
		t.Errorf("mistyped verb stderr = %q, want the guard refusal naming the verb list", errOut)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("record holds %d items, want 2", len(rec.Items))
	}
}
