package curator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixture writes a file under a t.TempDir()-isolated directory and
// returns its absolute path. Every test in this package uses t.TempDir()
// isolation per REQ-HEV2-034 / CLAUDE.local.md §6.
func writeFixture(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// mustRead reads a file, failing the test on error.
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// --- AC-HEV2-001: WriteManagedBlock signature exists and is callable ---

func TestWriteManagedBlock_Signature(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "k1", Text: "use context-first discovery before non-trivial work"},
		},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock returned error: %v", err)
	}
	data := mustRead(t, path)
	if !strings.Contains(string(data), "## MOAI:LEARNED-WORKFLOW") {
		t.Error("heading not written")
	}
	if !strings.Contains(string(data), "<!-- moai:learned-start -->") {
		t.Error("start marker not written")
	}
	if !strings.Contains(string(data), "<!-- moai:learned-end -->") {
		t.Error("end marker not written")
	}
}

// --- AC-HEV2-002: atomic replace — existing block content is replaced, leaving one block ---

func TestWriteManagedBlock_AtomicReplace(t *testing.T) {
	initial := strings.Join([]string{
		"# Project",
		"",
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"- old bullet <!-- key: old -->",
		"<!-- moai:learned-end -->",
		"",
	}, "\n")
	path := writeFixture(t, initial)

	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "new", Text: "new distilled workflow knowledge"},
		},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock: %v", err)
	}
	data := string(mustRead(t, path))

	if c := strings.Count(data, "## MOAI:LEARNED-WORKFLOW"); c != 1 {
		t.Errorf("heading count = %d, want 1 (atomic replace must not duplicate)", c)
	}
	if c := strings.Count(data, "<!-- moai:learned-start"); c != 1 {
		t.Errorf("start marker count = %d, want 1", c)
	}
	if c := strings.Count(data, "<!-- moai:learned-end -->"); c != 1 {
		t.Errorf("end marker count = %d, want 1", c)
	}
	if strings.Contains(data, "old bullet") {
		t.Error("old bullet still present after replace")
	}
	if !strings.Contains(data, "new distilled workflow knowledge") {
		t.Error("new bullet missing after replace")
	}
}

// --- AC-HEV2-003: BlockType enum includes HarnessGenerated; marker registry
// matches the legacy InjectMarker format byte-for-byte (D1 load-bearing) ---

func TestBlockTypeEnum_IncludesHarnessGenerated(t *testing.T) {
	// The two enum values must be distinct and usable.
	if BlockTypeLearnedWorkflow == BlockTypeHarnessGenerated {
		t.Fatal("BlockTypeLearnedWorkflow and BlockTypeHarnessGenerated must be distinct")
	}

	// The HarnessGenerated marker registry entry MUST produce a block
	// byte-identical to the legacy buildMarkerBlock output. This is the D1
	// load-bearing constraint: install.go:85 calls InjectMarker, which
	// delegates to WriteManagedBlock; the production output must not change.
	path := writeFixture(t, "# Project\n")

	body := "**Domain**: ios-mobile\n**Harness level**: standard\n**Updated**: 2026-07-12\n\nSee @.moai/harness/main.md\n"
	startAttrs := ` id="SPEC-PROJ-INIT-001" generated="2026-07-12T00:00:00Z"`

	err := WriteManagedBlock(path, BlockTypeHarnessGenerated, BlockContent{
		RawBody:   body,
		StartAttrs: startAttrs,
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock HarnessGenerated: %v", err)
	}

	got := string(mustRead(t, path))
	want := "# Project\n" +
		"\n" +
		"## Project-Specific Configuration (Harness-Generated)\n" +
		`<!-- moai:harness-start id="SPEC-PROJ-INIT-001" generated="2026-07-12T00:00:00Z" -->` + "\n" +
		"**Domain**: ios-mobile\n" +
		"**Harness level**: standard\n" +
		"**Updated**: 2026-07-12\n" +
		"\n" +
		"See @.moai/harness/main.md\n" +
		"<!-- moai:harness-end -->\n"

	if got != want {
		t.Errorf("HarnessGenerated block not byte-identical to legacy format:\n--- got (%d bytes) ---\n%s\n--- want (%d bytes) ---\n%s", len(got), got, len(want), want)
	}
}

// --- AC-HEV2-004: idempotent re-application produces zero byte-diff ---

func TestWriteManagedBlock_Idempotent_ZeroByteDiff(t *testing.T) {
	path := writeFixture(t, "# Project\npre-existing content\n")

	content := BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "k1", Text: "first distilled rule"},
			{LedgerKey: "k2", Text: "second distilled rule"},
		},
	}

	if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, content); err != nil {
		t.Fatalf("first write: %v", err)
	}
	afterFirst := mustRead(t, path)

	// Re-apply the SAME content — the file must not change.
	if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, content); err != nil {
		t.Fatalf("second write: %v", err)
	}
	afterSecond := mustRead(t, path)

	if string(afterFirst) != string(afterSecond) {
		t.Errorf("idempotency violated: re-applying identical content changed the file\n--- first ---\n%s\n--- second ---\n%s", afterFirst, afterSecond)
	}
}

// --- AC-HEV2-005: byte preservation — pre-block and post-block bytes unchanged ---

func TestWriteManagedBlock_PreBlockPostBlockPreserved(t *testing.T) {
	// The fixture carries an EXISTING block so WriteManagedBlock exercises
	// the atomic-replace path (not append). Bytes before and after the
	// marker block must survive the replace verbatim.
	preBlock := "# Title\n\nSome introductory prose that must survive verbatim.\n\n"
	existingBlock := "## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n- old bullet <!-- key: old -->\n<!-- moai:learned-end -->\n"
	postBlock := "\n## Another Section\n\nMore content after the block.\n"
	path := writeFixture(t, preBlock+existingBlock+postBlock)

	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{{LedgerKey: "k1", Text: "a rule"}},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock: %v", err)
	}
	data := string(mustRead(t, path))

	if !strings.HasPrefix(data, preBlock) {
		t.Errorf("pre-block bytes not preserved:\n--- want prefix ---\n%s\n--- got start ---\n%s", preBlock, data[:min(len(preBlock), len(data))])
	}
	if !strings.HasSuffix(data, postBlock) {
		t.Errorf("post-block bytes not preserved:\n--- want suffix ---\n%s\n--- got end ---\n%s", postBlock, data[max(0, len(data)-len(postBlock)):])
	}
}

// --- AC-HEV2-006: append-mode respects newline convention ---

func TestWriteManagedBlock_AppendMode_NewlineConvention(t *testing.T) {
	t.Run("file_ending_with_newline", func(t *testing.T) {
		path := writeFixture(t, "# Project\n")
		err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
			Bullets: []Bullet{{LedgerKey: "k1", Text: "rule"}},
		})
		if err != nil {
			t.Fatalf("WriteManagedBlock: %v", err)
		}
		data := string(mustRead(t, path))
		// A single blank line must separate the existing content from the
		// new block (not a double-blank-line, not a missing separator).
		if strings.Contains(data, "# Project\n\n\n## MOAI:LEARNED-WORKFLOW") {
			t.Error("double blank line inserted before block")
		}
		if !strings.Contains(data, "# Project\n\n## MOAI:LEARNED-WORKFLOW") {
			t.Errorf("missing blank-line separator before block:\n%s", data)
		}
	})

	t.Run("file_without_trailing_newline", func(t *testing.T) {
		path := writeFixture(t, "# Project (no trailing newline)")
		err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
			Bullets: []Bullet{{LedgerKey: "k1", Text: "rule"}},
		})
		if err != nil {
			t.Fatalf("WriteManagedBlock: %v", err)
		}
		data := string(mustRead(t, path))
		// The writer must insert a separating newline so the last existing
		// line is not concatenated with the block heading.
		if strings.Contains(data, "no trailing newline)## MOAI:LEARNED-WORKFLOW") {
			t.Error("existing line concatenated with block heading (missing newline)")
		}
	})
}

// --- Edge case: empty block (zero bullets) ---

func TestWriteManagedBlock_EmptyBlock(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{})
	if err != nil {
		t.Fatalf("WriteManagedBlock empty: %v", err)
	}
	data := string(mustRead(t, path))
	// An empty block carries heading + markers with no bullets between them.
	if !strings.Contains(data, "## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->") {
		t.Errorf("empty block structure incorrect:\n%s", data)
	}
}

// --- Edge case: provisional bullet (empty LedgerKey) omits the key comment ---

func TestWriteManagedBlock_ProvisionalBullet(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
		Bullets: []Bullet{
			{LedgerKey: "", Text: "early observation", Provisional: true},
		},
	})
	if err != nil {
		t.Fatalf("WriteManagedBlock: %v", err)
	}
	data := string(mustRead(t, path))
	if !strings.Contains(string(data), "- early observation") {
		t.Error("provisional bullet text missing")
	}
	if strings.Contains(string(data), "<!-- key:") {
		t.Error("provisional bullet should not carry a key comment")
	}
}

// --- WriteManagedBlock error paths ---

func TestWriteManagedBlock_EmptyPath(t *testing.T) {
	if err := WriteManagedBlock("", BlockTypeLearnedWorkflow, BlockContent{}); err == nil {
		t.Fatal("expected error on empty path")
	}
}

func TestWriteManagedBlock_FileNotFound(t *testing.T) {
	if err := WriteManagedBlock(filepath.Join(t.TempDir(), "nope.md"), BlockTypeLearnedWorkflow, BlockContent{}); err == nil {
		t.Fatal("expected error on missing file")
	}
}
