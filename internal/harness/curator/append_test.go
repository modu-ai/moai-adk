package curator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureLocalWithEntries creates a CLAUDE.local.md with a populated
// MOAI:LEARNED-WORKFLOW-LOCAL block containing the given entries (key → text).
func fixtureLocalWithEntries(t *testing.T, entries ...[2]string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.local.md")
	var b strings.Builder
	b.WriteString("# Local\n\n")
	b.WriteString("## MOAI:LEARNED-WORKFLOW-LOCAL\n")
	b.WriteString("<!-- moai:learned-local-start -->\n")
	for _, e := range entries {
		b.WriteString("- ")
		b.WriteString(e[1])
		b.WriteString(" <!-- key: ")
		b.WriteString(e[0])
		b.WriteString(" -->\n")
	}
	b.WriteString("<!-- moai:learned-local-end -->\n")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// fixtureLocalEmpty creates a CLAUDE.local.md with an EMPTY LOCAL block (start
// and end markers present, zero entries between them).
func fixtureLocalEmpty(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.local.md")
	body := strings.Join([]string{
		"# Local",
		"",
		"## MOAI:LEARNED-WORKFLOW-LOCAL",
		"<!-- moai:learned-local-start -->",
		"<!-- moai:learned-local-end -->",
		"",
	}, "\n")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// --- AC-HEV2-018: append-only — existing entry bytes are unchanged on append
// (REQ-HEV2-012). The Tier-3 surface is the permanent-record layer: append
// MUST NOT modify or delete any existing entry's bytes. ---

func TestAppendOnly_ExistingBytesUnchanged(t *testing.T) {
	path := fixtureLocalWithEntries(t,
		[2]string{"alpha", "observed pattern alpha"},
		[2]string{"beta", "observed pattern beta"},
		[2]string{"gamma", "observed pattern gamma"},
	)
	before := mustRead(t, path)

	err := AppendLearnedLocal(path, Bullet{
		LedgerKey: "delta",
		Text:      "observed pattern delta",
	})
	if err != nil {
		t.Fatalf("AppendLearnedLocal: %v", err)
	}
	after := mustRead(t, path)

	// Every existing entry's bytes MUST be present unchanged in the result.
	for _, frag := range []string{
		"- observed pattern alpha <!-- key: alpha -->",
		"- observed pattern beta <!-- key: beta -->",
		"- observed pattern gamma <!-- key: gamma -->",
	} {
		if !strings.Contains(string(after), frag) {
			t.Errorf("existing entry bytes modified by append; missing %q in:\n%s", frag, after)
		}
	}
	// The new entry MUST be appended.
	if !strings.Contains(string(after), "- observed pattern delta <!-- key: delta -->") {
		t.Error("new entry delta not appended")
	}
	// Exactly one new line was inserted (single-line append).
	beforeLines := strings.Count(string(before), "\n")
	afterLines := strings.Count(string(after), "\n")
	if afterLines-beforeLines != 1 {
		t.Errorf("line delta = %d, want 1 (append-only single-line insert)", afterLines-beforeLines)
	}
}

// --- AC-HEV2-020: dedup — appending an entry whose ledger_key already exists
// returns ErrDuplicateAppend without writing (REQ-HEV2-014). ---

func TestAppendOnly_DedupSameLedgerKey_ErrDuplicateAppend(t *testing.T) {
	path := fixtureLocalWithEntries(t,
		[2]string{"alpha", "observed pattern alpha"},
		[2]string{"beta", "observed pattern beta"},
	)
	before := mustRead(t, path)

	err := AppendLearnedLocal(path, Bullet{
		LedgerKey: "beta", // duplicate key
		Text:      "different text for the same key",
	})
	if !errors.Is(err, ErrDuplicateAppend) {
		t.Fatalf("expected ErrDuplicateAppend, got %v", err)
	}
	after := mustRead(t, path)

	// NO bytes may be written on a dedup no-op (REQ-HEV2-014: "no bytes are written").
	if string(before) != string(after) {
		t.Errorf("dedup append modified the file:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	// The rejected entry's text MUST NOT appear.
	if strings.Contains(string(after), "different text for the same key") {
		t.Error("rejected duplicate text leaked into the file")
	}
}

func TestAppendOnly_DedupDoesNotMatchOtherSection(t *testing.T) {
	// The dedup guard is scoped to the LOCAL section: a ledger_key that appears
	// ONLY in the digest (LEARNED-WORKFLOW) block MUST NOT trigger dedup on a
	// LOCAL append. The two surfaces are distinct layers.
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.local.md")
	body := strings.Join([]string{
		"# Local",
		"",
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"- distilled rule <!-- key: shared-key -->",
		"<!-- moai:learned-end -->",
		"",
		"## MOAI:LEARNED-WORKFLOW-LOCAL",
		"<!-- moai:learned-local-start -->",
		"<!-- moai:learned-local-end -->",
		"",
	}, "\n")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	err := AppendLearnedLocal(p, Bullet{
		LedgerKey: "shared-key", // exists in digest block, NOT in LOCAL block
		Text:      "local observation sharing a key with a digest rule",
	})
	if err != nil {
		t.Fatalf("append should succeed (key not in LOCAL section): %v", err)
	}
	data := string(mustRead(t, p))
	if !strings.Contains(data, "local observation sharing a key") {
		t.Error("entry not appended despite key being unique to the LOCAL section")
	}
}

// --- Coverage: fresh-block append, provisional entry, error paths ---

func TestAppendLearnedLocal_FreshBlockFirstEntry(t *testing.T) {
	path := fixtureLocalEmpty(t)
	err := AppendLearnedLocal(path, Bullet{
		LedgerKey: "first",
		Text:      "the very first local observation",
	})
	if err != nil {
		t.Fatalf("AppendLearnedLocal on empty block: %v", err)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(data, "- the very first local observation <!-- key: first -->") {
		t.Errorf("first entry not appended to empty block:\n%s", data)
	}
	// Exactly one bullet line now exists between the markers.
	localRe := compiledPatterns[BlockTypeLearnedLocal]
	match := localRe.FindString(data)
	if c := strings.Count(match, "\n- "); c != 1 {
		t.Errorf("expected exactly 1 bullet in LOCAL block, got %d", c)
	}
}

func TestAppendLearnedLocal_ProvisionalNoDedup(t *testing.T) {
	// A provisional entry (empty LedgerKey) bypasses the dedup guard — there
	// is no key to dedup against (evidence-or-null, REQ-HEV2-024).
	path := fixtureLocalEmpty(t)

	first := AppendLearnedLocal(path, Bullet{
		LedgerKey: "",
		Text:      "provisional observation one",
	})
	if first != nil {
		t.Fatalf("first provisional append: %v", first)
	}
	// A second provisional entry with different text MUST also append (no key,
	// no dedup collision).
	second := AppendLearnedLocal(path, Bullet{
		LedgerKey: "",
		Text:      "provisional observation two",
	})
	if second != nil {
		t.Fatalf("second provisional append: %v", second)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(data, "provisional observation one") {
		t.Error("first provisional entry missing")
	}
	if !strings.Contains(data, "provisional observation two") {
		t.Error("second provisional entry missing")
	}
}

func TestAppendLearnedLocal_EmptyPath(t *testing.T) {
	if err := AppendLearnedLocal("", Bullet{LedgerKey: "k", Text: "t"}); err == nil {
		t.Fatal("expected error on empty path")
	}
}

func TestAppendLearnedLocal_BlockNotFound(t *testing.T) {
	// A file with no LOCAL marker block MUST return ErrBlockNotFound.
	path := writeFixture(t, "# Project with no LOCAL block\n")
	err := AppendLearnedLocal(path, Bullet{LedgerKey: "k", Text: "t"})
	if !errors.Is(err, ErrBlockNotFound) {
		t.Fatalf("expected ErrBlockNotFound, got %v", err)
	}
}

func TestAppendLearnedLocal_FileNotFound(t *testing.T) {
	err := AppendLearnedLocal(filepath.Join(t.TempDir(), "nope.md"), Bullet{
		LedgerKey: "k",
		Text:      "t",
	})
	if err == nil {
		t.Fatal("expected error on missing file")
	}
}
