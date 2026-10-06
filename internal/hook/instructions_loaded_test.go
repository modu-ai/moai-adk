package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestInstructionsLoadedHandler_EventType(t *testing.T) {
	h := NewInstructionsLoadedHandler()
	if h.EventType() != EventInstructionsLoaded {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventInstructionsLoaded)
	}
}

func TestInstructionsLoadedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *HookInput
		createFile    bool
		fileContent   string
		expectMessage bool
	}{
		{
			name: "no instruction file path",
			input: &HookInput{
				SessionID:     "test-session",
				HookEventName: "InstructionsLoaded",
			},
			createFile:    false,
			expectMessage: false,
		},
		{
			name: "small file within budget",
			input: &HookInput{
				SessionID:           "test-session",
				InstructionFilePath: "CLAUDE.md",
				CWD:                 "",
				HookEventName:       "InstructionsLoaded",
			},
			createFile:    true,
			fileContent:   "# Small file\n\nThis is well within budget.",
			expectMessage: false,
		},
		{
			name: "file exceeding budget",
			input: &HookInput{
				SessionID:           "test-session",
				InstructionFilePath: "CLAUDE.md",
				CWD:                 "",
				HookEventName:       "InstructionsLoaded",
			},
			createFile:    true,
			fileContent:   string(make([]byte, 45000)), // 45KB
			expectMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewInstructionsLoadedHandler()

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.InstructionFilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filePath)
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.InstructionFilePath = filePath
				tt.input.CWD = tempDir
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			if tt.expectMessage && out.SystemMessage == "" {
				t.Error("expected SystemMessage for budget violation")
			}
			if !tt.expectMessage && out.SystemMessage != "" {
				t.Errorf("unexpected SystemMessage: %v", out.SystemMessage)
			}
		})
	}
}

func TestInstructionsLoadedHandler_CheckCharacterBudget(t *testing.T) {
	t.Parallel()

	h := &instructionsLoadedHandler{}

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "small file",
			content:     "# Hello\n\nWorld",
			expectError: false,
		},
		{
			name:        "exactly at limit",
			content:     string(make([]byte, 40000)),
			expectError: false,
		},
		{
			name:        "exceeds limit by one",
			content:     string(make([]byte, 40001)),
			expectError: true,
		},
		{
			name:        "far exceeds limit",
			content:     string(make([]byte, 50000)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			tempFile, err := os.CreateTemp("", "budget-test-*.md")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tempFile.Name()) }()

			// Write content
			if _, err := tempFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			_ = tempFile.Close()

			// Check budget
			err = h.checkCharacterBudget(tempFile.Name())
			if tt.expectError && err == nil {
				t.Error("expected error for budget violation")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// writeBudgetFixture lays out the shared fixture project used by the
// SPEC-INSTRUCTIONS-BUDGET-001 regression trio:
//
//	root/
//	  .moai/                                  (dir — resolveProjectRoot requires it)
//	  CLAUDE.md                               @-imports AGENTS.md (twice, for dedup)
//	                                          and missing.md (dangling — excluded)
//	  AGENTS.md                               plain import-closure member
//	  .claude/rules/moai/always-core.md       no frontmatter  → always-loaded
//	  .claude/rules/moai/scoped.md            top-level paths: → excluded
//	  .claude/rules/moai/notes.txt            non-md          → excluded
func writeBudgetFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude", "rules", "moai"), 0o750); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"CLAUDE.md":                         "# Project\n\n@AGENTS.md\n@AGENTS.md\n@missing.md\n",
		"AGENTS.md":                         "# Agents\n\nThe standing contract body.\n",
		".claude/rules/moai/always-core.md": "# Always-loaded core rule\n\nNo frontmatter block.\n",
		".claude/rules/moai/scoped.md":      "---\npaths:\n  - \"**/scoped/**\"\n---\n\n# Scoped rule\n",
		".claude/rules/moai/notes.txt":      "not markdown\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestInstructionFileSetDerivation covers the mechanically-derived set
// semantics (SPEC-INSTRUCTIONS-BUDGET-001 REQ-INSTRBUDGET-004): anchor +
// transitive @-import closure with dedup and dangling-member skipping, plus
// the always-loaded rule selector. A hardcoded-list mutant fails the growth
// assertions.
func TestInstructionFileSetDerivation(t *testing.T) {
	t.Parallel()

	root := writeBudgetFixture(t)
	want := []string{
		filepath.Join(root, ".claude", "rules", "moai", "always-core.md"),
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "CLAUDE.md"),
	}
	if got := instructionFileSet(root); !reflect.DeepEqual(got, want) {
		t.Errorf("instructionFileSet() = %v, want %v", got, want)
	}

	// A new always-loaded rule grows the set by exactly one.
	added := filepath.Join(root, ".claude", "rules", "moai", "added-always.md")
	if err := os.WriteFile(added, []byte("# new always-loaded rule\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	grown := instructionFileSet(root)
	if len(grown) != len(want)+1 {
		t.Errorf("always-loaded growth: set has %d members, want %d", len(grown), len(want)+1)
	}

	// A paths:-scoped rule changes nothing (the hardcoded-list mutant fails here).
	scoped := filepath.Join(root, ".claude", "rules", "moai", "added-scoped.md")
	if err := os.WriteFile(scoped, []byte("---\npaths:\n  - \"x\"\n---\nscoped body\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if res := instructionFileSet(root); len(res) != len(grown) {
		t.Errorf("paths:-scoped rule changed the set: %d members, want %d", len(res), len(grown))
	}

	// Malformed frontmatter (opening --- with no closer) is always-loaded.
	malformed := filepath.Join(root, ".claude", "rules", "moai", "malformed.md")
	if err := os.WriteFile(malformed, []byte("---\ntitle: never closed\nbody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if res := instructionFileSet(root); len(res) != len(grown)+1 {
		t.Errorf("malformed-frontmatter rule missing from set: %d members, want %d", len(res), len(grown)+1)
	}

	// An indented paths: is not top-level — the rule stays always-loaded.
	indented := filepath.Join(root, ".claude", "rules", "moai", "indented.md")
	if err := os.WriteFile(indented, []byte("---\nmeta:\n  paths:\n    - \"y\"\n---\nbody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if res := instructionFileSet(root); len(res) != len(grown)+2 {
		t.Errorf("indented paths: wrongly scoped the rule: %d members, want %d", len(res), len(grown)+2)
	}
}

// TestInstructionFileSetCycleSafe pins the visited-set behavior: a mutual
// @-import between CLAUDE.md and AGENTS.md must terminate and yield both
// members exactly once.
func TestInstructionFileSetCycleSafe(t *testing.T) {
	t.Parallel()

	root := writeBudgetFixture(t)
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# Agents\n@CLAUDE.md\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := instructionFileSet(root)
	want := []string{
		filepath.Join(root, ".claude", "rules", "moai", "always-core.md"),
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "CLAUDE.md"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cycle-safe set = %v, want %v", got, want)
	}
}

// TestAggregateInstructionCharsMovement pins the metric unit and freshness
// (REQ-INSTRBUDGET-001): appending exactly N runes to one member moves the
// aggregate by exactly N. Wrong-unit (byte) and stale-count mutants fail.
func TestAggregateInstructionCharsMovement(t *testing.T) {
	t.Parallel()

	root := writeBudgetFixture(t)
	before := aggregateInstructionChars(root)
	if before == 0 {
		t.Fatal("fixture aggregate is zero; fixture is broken")
	}

	const growBy = 137
	path := filepath.Join(root, "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	padded := append(data, []byte(strings.Repeat("x", growBy))...)
	if err := os.WriteFile(path, padded, 0o600); err != nil {
		t.Fatal(err)
	}

	after := aggregateInstructionChars(root)
	if delta := after - before; delta != growBy {
		t.Errorf("aggregate moved %d for a %d-rune append; want exact movement", delta, growBy)
	}
}

// TestInstructionsLoadedAggregateBudget is the Handle-integration boundary
// trio (REQ-INSTRBUDGET-002): grown-past fixture → advisory SystemMessage
// naming aggregate, budget, and file count with NO block decision;
// under-budget fixture → empty HookOutput; exactly-at-budget fixture → no
// trip (strictly greater-than). Per-file 40,000 behavior is unchanged.
//
// No t.Parallel(): the subtests pin the project root via t.Setenv.
func TestInstructionsLoadedAggregateBudget(t *testing.T) {
	newInput := func(root string) *HookInput {
		return &HookInput{SessionID: "t1318", HookEventName: "InstructionsLoaded", CWD: root}
	}

	t.Run("over budget emits advisory SystemMessage with no block decision", func(t *testing.T) {
		root := writeBudgetFixture(t)
		rulesDir := filepath.Join(root, ".claude", "rules", "moai")
		for i := 0; i < 12; i++ {
			bulk := filepath.Join(rulesDir, fmt.Sprintf("bulk-%02d.md", i))
			if err := os.WriteFile(bulk, []byte(strings.Repeat("a", 21000)), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv(config.EnvClaudeProjectDir, root)

		out, err := NewInstructionsLoadedHandler().Handle(context.Background(), newInput(root))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out == nil {
			t.Fatal("expected non-nil output")
		}
		if out.SystemMessage == "" {
			t.Fatal("expected advisory SystemMessage past sessionCharBudget")
		}
		if agg := aggregateInstructionChars(root); !strings.Contains(out.SystemMessage, strconv.Itoa(agg)) {
			t.Errorf("SystemMessage %q missing aggregate figure %d", out.SystemMessage, agg)
		}
		if !strings.Contains(out.SystemMessage, strconv.Itoa(sessionCharBudget)) {
			t.Errorf("SystemMessage %q missing budget %d", out.SystemMessage, sessionCharBudget)
		}
		if n := len(instructionFileSet(root)); !strings.Contains(out.SystemMessage, strconv.Itoa(n)) {
			t.Errorf("SystemMessage %q missing derived file count %d", out.SystemMessage, n)
		}
		// Advisory-only (D1): no block decision, no continue-halt, no reason,
		// no suppressed output — the event itself stays unblocked.
		if out.Decision != "" || out.Continue != nil || out.Reason != "" || out.SuppressOutput {
			t.Errorf("advisory breach mutated control fields: %+v", out)
		}
	})

	t.Run("under budget returns empty output", func(t *testing.T) {
		root := writeBudgetFixture(t)
		t.Setenv(config.EnvClaudeProjectDir, root)

		out, err := NewInstructionsLoadedHandler().Handle(context.Background(), newInput(root))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out == nil || out.SystemMessage != "" {
			t.Errorf("under-budget fixture must return an empty HookOutput, got %+v", out)
		}
	})

	t.Run("exactly at sessionCharBudget does not trip", func(t *testing.T) {
		root := writeBudgetFixture(t)
		rulesDir := filepath.Join(root, ".claude", "rules", "moai")
		for i := 0; i < 9; i++ {
			edge := filepath.Join(rulesDir, fmt.Sprintf("edge-%02d.md", i))
			if err := os.WriteFile(edge, []byte(strings.Repeat("a", 20000)), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		// Pad one member so the aggregate lands on exactly 210000.
		pad := sessionCharBudget - aggregateInstructionChars(root)
		if pad <= 0 || pad >= 40000 {
			t.Fatalf("boundary pad %d outside the per-file-safe band; fixture design broken", pad)
		}
		padded := filepath.Join(rulesDir, "edge-pad.md")
		if err := os.WriteFile(padded, []byte(strings.Repeat("b", pad)), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := aggregateInstructionChars(root); got != sessionCharBudget {
			t.Fatalf("boundary fixture aggregates %d, want exactly %d", got, sessionCharBudget)
		}
		t.Setenv(config.EnvClaudeProjectDir, root)

		out, err := NewInstructionsLoadedHandler().Handle(context.Background(), newInput(root))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out == nil || out.SystemMessage != "" {
			t.Errorf("exactly-at-budget fixture must not trip (strictly greater-than), got %+v", out)
		}
	})

	t.Run("per-file 40000 check unchanged on same fixtures", func(t *testing.T) {
		root := writeBudgetFixture(t)
		if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(strings.Repeat("c", 45000)), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv(config.EnvClaudeProjectDir, root)

		// The per-file branch requires an instruction file path on the input;
		// the CWD fallback only logs, so point the input at the oversized file.
		in := newInput(root)
		in.InstructionFilePath = filepath.Join(root, "CLAUDE.md")

		out, err := NewInstructionsLoadedHandler().Handle(context.Background(), in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out == nil {
			t.Fatal("expected non-nil output")
		}
		if !strings.Contains(out.SystemMessage, "40,000 char budget") {
			t.Errorf("per-file violation must keep its original message, got %q", out.SystemMessage)
		}
		if strings.Contains(out.SystemMessage, "session budget") {
			t.Errorf("aggregate advisory leaked into the per-file message: %q", out.SystemMessage)
		}
	})
}
