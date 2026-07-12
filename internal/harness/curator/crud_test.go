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
