package curator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureWithBullets creates a CLAUDE.md with a populated LEARNED-WORKFLOW
// block containing the given bullets (key → text).
func fixtureWithBullets(t *testing.T, bullets ...[2]string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.md")
	var b strings.Builder
	b.WriteString("# Project\n\n")
	b.WriteString("## MOAI:LEARNED-WORKFLOW\n")
	b.WriteString("<!-- moai:learned-start -->\n")
	for _, bl := range bullets {
		b.WriteString("- ")
		b.WriteString(bl[1])
		b.WriteString(" <!-- key: ")
		b.WriteString(bl[0])
		b.WriteString(" -->\n")
	}
	b.WriteString("<!-- moai:learned-end -->\n")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// --- AC-HEV2-009: AddBullet inserts a single new line, preserves all existing bullets ---

func TestAddBullet_SingleLine_RewritesOnlyTarget(t *testing.T) {
	path := fixtureWithBullets(t,
		[2]string{"a", "rule A"},
		[2]string{"b", "rule B"},
		[2]string{"c", "rule C"},
		[2]string{"d", "rule D"},
	)
	before := mustRead(t, path)

	err := AddBullet(path, BlockTypeLearnedWorkflow, Bullet{
		LedgerKey: "e",
		Text:      "rule E",
	})
	if err != nil {
		t.Fatalf("AddBullet: %v", err)
	}
	after := mustRead(t, path)

	// Every existing bullet's bytes must be unchanged.
	for _, frag := range []string{"rule A", "rule B", "rule C", "rule D"} {
		if !strings.Contains(string(after), frag) {
			t.Errorf("existing bullet %q clobbered by AddBullet", frag)
		}
	}
	// The new bullet must be present.
	if !strings.Contains(string(after), "rule E") {
		t.Error("new bullet E missing after AddBullet")
	}
	if !strings.Contains(string(after), "<!-- key: e -->") {
		t.Error("new bullet key marker missing")
	}
	// Exactly one new line was inserted (after has exactly one more line than before).
	beforeLines := strings.Count(string(before), "\n")
	afterLines := strings.Count(string(after), "\n")
	if afterLines-beforeLines != 1 {
		t.Errorf("line delta = %d, want 1 (single-line insert)", afterLines-beforeLines)
	}
}

// --- AC-HEV2-010: UpdateBullet rewrites the text of a bullet matched by ledger_key ---

func TestUpdateBullet_ByLedgerKey(t *testing.T) {
	path := fixtureWithBullets(t,
		[2]string{"a", "original A"},
		[2]string{"b", "original B"},
		[2]string{"c", "original C"},
	)

	err := UpdateBullet(path, BlockTypeLearnedWorkflow, "b", "updated B text")
	if err != nil {
		t.Fatalf("UpdateBullet: %v", err)
	}
	data := string(mustRead(t, path))

	if !strings.Contains(data, "- updated B text <!-- key: b -->") {
		t.Error("bullet B text not updated")
	}
	if strings.Contains(data, "original B") {
		t.Error("old text 'original B' still present after update")
	}
	// Other bullets untouched.
	if !strings.Contains(data, "original A") || !strings.Contains(data, "original C") {
		t.Error("non-target bullets modified by UpdateBullet")
	}
}

func TestUpdateBullet_NotFound(t *testing.T) {
	path := fixtureWithBullets(t, [2]string{"a", "rule A"})
	err := UpdateBullet(path, BlockTypeLearnedWorkflow, "nonexistent", "text")
	if err == nil {
		t.Fatal("expected error for nonexistent ledger_key")
	}
}

// --- AC-HEV2-011: DeleteBullet removes the matched bullet, preserves all others ---

func TestDeleteBullet_ByLedgerKey_PreservesOthers(t *testing.T) {
	path := fixtureWithBullets(t,
		[2]string{"a", "rule A"},
		[2]string{"b", "rule B"},
		[2]string{"c", "rule C"},
	)

	err := DeleteBullet(path, BlockTypeLearnedWorkflow, "b")
	if err != nil {
		t.Fatalf("DeleteBullet: %v", err)
	}
	data := string(mustRead(t, path))

	if strings.Contains(data, "rule B") {
		t.Error("deleted bullet B still present")
	}
	if strings.Contains(data, "<!-- key: b -->") {
		t.Error("deleted bullet B key marker still present")
	}
	// Remaining bullets preserved.
	if !strings.Contains(data, "rule A") || !strings.Contains(data, "rule C") {
		t.Error("non-target bullets modified by DeleteBullet")
	}
}

func TestDeleteBullet_NotFound(t *testing.T) {
	path := fixtureWithBullets(t, [2]string{"a", "rule A"})
	err := DeleteBullet(path, BlockTypeLearnedWorkflow, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent ledger_key")
	}
}

// --- CRUD error paths (coverage + robustness) ---

func TestAddBullet_EmptyPath(t *testing.T) {
	if err := AddBullet("", BlockTypeLearnedWorkflow, Bullet{LedgerKey: "k", Text: "t"}); err == nil {
		t.Fatal("expected error on empty path")
	}
}

func TestAddBullet_BlockNotFound(t *testing.T) {
	path := writeFixture(t, "# Project with no block\n")
	err := AddBullet(path, BlockTypeLearnedWorkflow, Bullet{LedgerKey: "k", Text: "t"})
	if err == nil {
		t.Fatal("expected ErrBlockNotFound when file has no marker block")
	}
}

func TestAddBullet_FileNotFound(t *testing.T) {
	err := AddBullet(filepath.Join(t.TempDir(), "nope.md"), BlockTypeLearnedWorkflow,
		Bullet{LedgerKey: "k", Text: "t"})
	if err == nil {
		t.Fatal("expected error on missing file")
	}
}

func TestUpdateBullet_EmptyPath(t *testing.T) {
	if err := UpdateBullet("", BlockTypeLearnedWorkflow, "k", "t"); err == nil {
		t.Fatal("expected error on empty path")
	}
}

func TestUpdateBullet_EmptyLedgerKey(t *testing.T) {
	path := fixtureWithBullets(t, [2]string{"a", "rule A"})
	if err := UpdateBullet(path, BlockTypeLearnedWorkflow, "", "t"); err == nil {
		t.Fatal("expected error on empty ledgerKey")
	}
}

func TestDeleteBullet_EmptyPath(t *testing.T) {
	if err := DeleteBullet("", BlockTypeLearnedWorkflow, "k"); err == nil {
		t.Fatal("expected error on empty path")
	}
}

func TestDeleteBullet_EmptyLedgerKey(t *testing.T) {
	path := fixtureWithBullets(t, [2]string{"a", "rule A"})
	if err := DeleteBullet(path, BlockTypeLearnedWorkflow, ""); err == nil {
		t.Fatal("expected error on empty ledgerKey")
	}
}

// --- AC-HEV2-015: ledger_key cross-layer linkage — each digest bullet carries
// a `<!-- key: <ledger_key> -->` trailing HTML comment linking the digest
// bullet to a ledger-layer entry (REQ-HEV2-010). ---

func TestBullet_LedgerKeyTrailingHTMLComment(t *testing.T) {
	path := writeFixture(t, "# Project\n")

	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "req-validation-fanout", Text: "validate cross-file reachability before declaring PASS"},
			{LedgerKey: "context-first-discovery", Text: "run socratic interview when intent is ambiguous"},
		},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock: %v", err)
	}
	data := string(mustRead(t, path))

	// Each bullet MUST carry its ledger_key in a trailing HTML comment of the
	// exact form `<!-- key: <ledger_key> -->`.
	for _, key := range []string{"req-validation-fanout", "context-first-discovery"} {
		want := "<!-- key: " + key + " -->"
		if !strings.Contains(data, want) {
			t.Errorf("ledger_key linkage marker missing for %q; want %q in:\n%s", key, want, data)
		}
	}

	// The marker MUST be on the SAME line as the bullet text (trailing, not on
	// a separate line), so the ledger_key binds to its bullet.
	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, "<!-- key:") {
			if !strings.HasPrefix(strings.TrimSpace(line), "- ") {
				t.Errorf("ledger_key marker not on a bullet line: %q", line)
			}
		}
	}
}

// --- AC-HEV2-015 (provisional): a provisional bullet (empty LedgerKey) omits
// the key marker (evidence-or-null, REQ-HEV2-010 + REQ-HEV2-024). ---

func TestBullet_ProvisionalOmitsKeyMarker(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "", Text: "early single observation", Provisional: true},
		},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock: %v", err)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(data, "- early single observation") {
		t.Error("provisional bullet text missing")
	}
	if strings.Contains(data, "<!-- key:") {
		t.Error("provisional bullet should not carry a key marker")
	}
}

// TestBullet_ProvisionalNullLedgerKey (AC-HEV2-023, Scenario 9): a Tier-1
// provisional observation carries a null (empty) ledger_key — the bullet renders
// with NO key marker and the recall-layer DigestEntry.Provisional() reports true
// (evidence-or-null, REQ-HEV2-017/024). A later Tier-3 promotion replaces the
// provisional marker with a real ledger_key.
func TestBullet_ProvisionalNullLedgerKey(t *testing.T) {
	provisional := Bullet{LedgerKey: "", Text: "single occurrence, no aggregated evidence", Provisional: true}
	if provisional.LedgerKey != "" {
		t.Fatalf("provisional bullet must carry a null (empty) ledger_key")
	}
	// The recall-layer DigestEntry mirror reports the provisional state.
	if de := (DigestEntry{Summary: provisional.Text, LedgerKey: provisional.LedgerKey}); !de.Provisional() {
		t.Errorf("DigestEntry.Provisional() should report true for a null ledger_key")
	}

	// Write the provisional bullet → renders with NO key marker.
	path := writeFixture(t, "# Project\n")
	if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{provisional},
	}); err != nil {
		t.Fatalf("WriteManagedBlock (provisional): %v", err)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(data, "- single occurrence, no aggregated evidence") {
		t.Fatal("provisional bullet text missing")
	}
	if strings.Contains(data, "<!-- key:") {
		t.Errorf("provisional bullet must not carry a key marker (null ledger_key)")
	}

	// Tier-3 promotion: the same pattern gains a real ledger_key. Re-writing the
	// block with the promoted bullet replaces the provisional marker with the key.
	promoted := Bullet{LedgerKey: "lw-promoted-001", Text: provisional.Text}
	if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{promoted},
	}); err != nil {
		t.Fatalf("WriteManagedBlock (promoted): %v", err)
	}
	promotedData := string(mustRead(t, path))
	if !strings.Contains(promotedData, "<!-- key: lw-promoted-001 -->") {
		t.Errorf("promoted bullet must carry the real ledger_key marker")
	}
	if de := (DigestEntry{Summary: promoted.Text, LedgerKey: promoted.LedgerKey}); de.Provisional() {
		t.Errorf("promoted DigestEntry.Provisional() should report false")
	}
}
