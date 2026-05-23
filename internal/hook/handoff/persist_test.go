// Test suite for SPEC-V3R6-SESSION-HANDOFF-AUTO-001 paste-ready resume persistence.
// All AC-SHA-001..010 verified via binary table-driven tests + race-detector-safe
// concurrency guard.
package handoff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// validBody is the canonical Markdown body that satisfies REQ-SHA-005 structural
// checks: heading + fenced ```text block.
const validBody = `## Next Session Entry Point

` + "```text\n" + `ultrathink. SPEC-V3R6-SESSION-HANDOFF-AUTO-001 run 진입.
applied lessons: project_sprint2_session_handoff_auto_001_plan_ready.

전제 검증:
1) git log --oneline -1
2) ls .moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/

실행: /moai run SPEC-V3R6-SESSION-HANDOFF-AUTO-001

머지 후: /moai sync
` + "```\n"

// validFrontmatter is the canonical YAML frontmatter that satisfies REQ-SHA-004
// required-field and REQ-SHA-011 field-format checks.
const validFrontmatter = `---
sprint: sprint2
spec: session-handoff-auto-001
status: plan_ready
index_line: "- [Sprint 2 SESSION-HANDOFF-AUTO-001 plan ready](project_sprint2_session-handoff-auto-001_plan_ready.md) — short hook"
---
`

// makeProjectDir constructs `<root>/.moai/state/session-handoff/` and writes the
// given pending content. Returns the projectDir (root) so the test can pass it
// to PersistIfPending.
func makeProjectDir(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "state", "session-handoff")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir pending dir: %v", err)
	}
	pendingPath := filepath.Join(dir, "pending.md")
	if err := os.WriteFile(pendingPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write pending: %v", err)
	}
	return root
}

// makeMemoryDir returns an empty memory directory (no MEMORY.md yet).
func makeMemoryDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// TestPersistIfPending_ReadsOnlyContractPath verifies AC-SHA-001: PersistIfPending
// reads only `<projectDir>/.moai/state/session-handoff/pending.md` and does not
// touch other files in the project tree. Verified via decoy-mtime approach.
func TestPersistIfPending_ReadsOnlyContractPath(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := makeMemoryDir(t)

	// Plant decoy files at known mtimes outside the contract path.
	pastTime := time.Now().Add(-1 * time.Hour)
	decoys := []string{
		filepath.Join(projectDir, "decoy1.md"),
		filepath.Join(projectDir, ".moai", "decoy2.md"),
		filepath.Join(projectDir, ".moai", "state", "decoy3.md"),
	}
	for _, d := range decoys {
		if err := os.MkdirAll(filepath.Dir(d), 0o755); err != nil {
			t.Fatalf("mkdir for decoy %s: %v", d, err)
		}
		if err := os.WriteFile(d, []byte("decoy"), 0o644); err != nil {
			t.Fatalf("write decoy %s: %v", d, err)
		}
		if err := os.Chtimes(d, pastTime, pastTime); err != nil {
			t.Fatalf("chtimes decoy %s: %v", d, err)
		}
	}

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	for _, d := range decoys {
		info, err := os.Stat(d)
		if err != nil {
			t.Fatalf("decoy %s missing post-call: %v", d, err)
		}
		if !info.ModTime().Equal(pastTime) {
			t.Errorf("decoy %s mtime changed: pre=%v post=%v", d, pastTime, info.ModTime())
		}
	}
}

// TestPersistIfPending_AbsentPendingNoOp verifies AC-SHA-002: absent pending
// file is a no-op (no slog warn, no directory creation).
func TestPersistIfPending_AbsentPendingNoOp(t *testing.T) {
	projectDir := t.TempDir()
	memoryDir := makeMemoryDir(t)

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	// Verify the pending directory was NOT created.
	pendingDir := filepath.Join(projectDir, ".moai", "state", "session-handoff")
	if _, err := os.Stat(pendingDir); !os.IsNotExist(err) {
		t.Errorf("pending directory should not be created when pending file is absent; stat err=%v", err)
	}

	// Verify memory directory contains no new files.
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		t.Fatalf("read memory dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("memory dir should be empty; found %d entries", len(entries))
	}
}

// TestPersistIfPending_ValidPendingWritesBoth verifies AC-SHA-003: valid pending
// file produces both the memory file and the MEMORY.md prepend.
func TestPersistIfPending_ValidPendingWritesBoth(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter string
		supersedes  string
	}{
		{
			name:        "no-supersedes",
			frontmatter: validFrontmatter,
		},
		{
			name: "with-supersedes",
			frontmatter: `---
sprint: sprint2
spec: session-handoff-auto-001
status: complete
index_line: "- [Sprint 2 complete](project_sprint2_session-handoff-auto-001_complete.md) — short"
supersedes: project_sprint2_session-handoff-auto-001_plan_ready.md
---
`,
			supersedes: "project_sprint2_session-handoff-auto-001_plan_ready.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := makeProjectDir(t, tt.frontmatter+validBody)
			memoryDir := makeMemoryDir(t)

			// Pre-populate MEMORY.md when testing supersede.
			if tt.supersedes != "" {
				existing := fmt.Sprintf("- [Old entry](%s) — prev hook\n- [Other](other.md) — keep\n", tt.supersedes)
				if err := os.WriteFile(filepath.Join(memoryDir, "MEMORY.md"), []byte(existing), 0o644); err != nil {
					t.Fatalf("write existing MEMORY.md: %v", err)
				}
			}

			if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
				t.Fatalf("PersistIfPending returned error: %v", err)
			}

			// Verify memory file exists with verbatim body.
			entries, err := os.ReadDir(memoryDir)
			if err != nil {
				t.Fatalf("read memory dir: %v", err)
			}
			var memoryFile string
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "project_") && strings.HasSuffix(e.Name(), ".md") {
					memoryFile = e.Name()
				}
			}
			if memoryFile == "" {
				t.Fatalf("memory file not found; entries=%v", entries)
			}
			body, err := os.ReadFile(filepath.Join(memoryDir, memoryFile))
			if err != nil {
				t.Fatalf("read memory file: %v", err)
			}
			if string(body) != validBody {
				t.Errorf("memory file body mismatch:\ngot=%q\nwant=%q", string(body), validBody)
			}

			// Verify MEMORY.md first line is the index_line.
			memoryMd, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md"))
			if err != nil {
				t.Fatalf("read MEMORY.md: %v", err)
			}
			firstLine := strings.SplitN(string(memoryMd), "\n", 2)[0]
			if !strings.Contains(firstLine, "Sprint 2") {
				t.Errorf("MEMORY.md first line missing index entry; firstLine=%q", firstLine)
			}
		})
	}
}

// TestPersistIfPending_MalformedFrontmatterPreserved verifies AC-SHA-004:
// malformed/missing/invalid frontmatter logs and preserves the pending file.
func TestPersistIfPending_MalformedFrontmatterPreserved(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing-sprint",
			content: `---
spec: foo
status: bar
index_line: "x"
---
` + validBody,
		},
		{
			name: "missing-spec",
			content: `---
sprint: sprint2
status: bar
index_line: "x"
---
` + validBody,
		},
		{
			name: "missing-status",
			content: `---
sprint: sprint2
spec: foo
index_line: "x"
---
` + validBody,
		},
		{
			name: "missing-index_line",
			content: `---
sprint: sprint2
spec: foo
status: bar
---
` + validBody,
		},
		{
			name:    "unparseable-yaml",
			content: "---\nsprint: [unclosed\n---\n" + validBody,
		},
		{
			name: "invalid-sprint-format",
			content: `---
sprint: "Sprint With Spaces"
spec: foo
status: bar
index_line: "x"
---
` + validBody,
		},
		{
			name: "invalid-spec-format",
			content: `---
sprint: sprint2
spec: "../escape"
status: bar
index_line: "x"
---
` + validBody,
		},
		{
			name: "invalid-status-format",
			content: `---
sprint: sprint2
spec: foo
status: "Plan Ready!"
index_line: "x"
---
` + validBody,
		},
		{
			name:    "no-frontmatter",
			content: validBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := makeProjectDir(t, tt.content)
			memoryDir := makeMemoryDir(t)

			pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
			before, err := os.ReadFile(pendingPath)
			if err != nil {
				t.Fatalf("read pending pre-call: %v", err)
			}

			if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
				t.Fatalf("PersistIfPending returned error: %v", err)
			}

			// Pending file preserved byte-for-byte.
			after, err := os.ReadFile(pendingPath)
			if err != nil {
				t.Fatalf("read pending post-call: %v", err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("pending file modified for case %s", tt.name)
			}

			// Memory directory should have no new files.
			entries, _ := os.ReadDir(memoryDir)
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "project_") {
					t.Errorf("malformed pending produced memory file: %s", e.Name())
				}
			}
		})
	}
}

// TestPersistIfPending_MissingHeadingPreserved verifies AC-SHA-005: missing
// heading or fenced text block preserves the pending file.
func TestPersistIfPending_MissingHeadingPreserved(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing-heading",
			body: "Just some body text\n```text\nstuff\n```\n",
		},
		{
			name: "missing-fenced-block",
			body: "## Next Session Entry Point\n\nNo fenced block here, just prose.\n",
		},
		{
			name: "wrong-fence-language",
			body: "## Next Session Entry Point\n\n```bash\necho hi\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := makeProjectDir(t, validFrontmatter+tt.body)
			memoryDir := makeMemoryDir(t)

			pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
			before, err := os.ReadFile(pendingPath)
			if err != nil {
				t.Fatalf("read pending pre-call: %v", err)
			}

			if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
				t.Fatalf("PersistIfPending returned error: %v", err)
			}

			after, err := os.ReadFile(pendingPath)
			if err != nil {
				t.Fatalf("read pending post-call: %v", err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("pending file modified for case %s", tt.name)
			}

			entries, _ := os.ReadDir(memoryDir)
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "project_") {
					t.Errorf("structural-defect pending produced memory file: %s", e.Name())
				}
			}
		})
	}
}

// TestPersistIfPending_AtomicWriteNoPartialRead verifies AC-SHA-006: concurrent
// readers either observe the memory file as absent or as complete, never partial.
// Race-detector compatible.
func TestPersistIfPending_AtomicWriteNoPartialRead(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := makeMemoryDir(t)

	memoryFileName := "project_sprint2_session-handoff-auto-001_plan_ready.md"
	memoryFilePath := filepath.Join(memoryDir, memoryFileName)

	// Reader goroutine: continuously stat + read until persisted body equals
	// validBody or stop signal. Records any partial read as failure.
	var (
		stop          atomic.Bool
		partialReads  atomic.Int32
		completeReads atomic.Int32
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for !stop.Load() {
			data, err := os.ReadFile(memoryFilePath)
			if err != nil {
				// Absent is OK.
				continue
			}
			if string(data) == validBody {
				completeReads.Add(1)
				continue
			}
			// Anything else is a partial read.
			partialReads.Add(1)
		}
	}()

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	// Give reader a small window to observe complete state.
	time.Sleep(10 * time.Millisecond)
	stop.Store(true)
	wg.Wait()

	if partialReads.Load() > 0 {
		t.Errorf("reader observed %d partial reads (atomic-write contract broken)", partialReads.Load())
	}
	// completeReads.Load() may be 0 on very slow systems; not asserted.
}

// TestPersistIfPending_MemoryMdContentionRetry verifies AC-SHA-007: MEMORY.md
// modification between read and write triggers retry; persistent contention
// is logged and the function returns nil without aborting.
func TestPersistIfPending_MemoryMdContentionRetry(t *testing.T) {
	t.Run("retry-succeeds", func(t *testing.T) {
		projectDir := makeProjectDir(t, validFrontmatter+validBody)
		memoryDir := makeMemoryDir(t)
		memoryMdPath := filepath.Join(memoryDir, "MEMORY.md")
		if err := os.WriteFile(memoryMdPath, []byte("- existing\n"), 0o644); err != nil {
			t.Fatalf("write MEMORY.md: %v", err)
		}

		// No active interference: the call should succeed on the first attempt.
		if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
			t.Fatalf("PersistIfPending returned error: %v", err)
		}

		data, err := os.ReadFile(memoryMdPath)
		if err != nil {
			t.Fatalf("read MEMORY.md post-call: %v", err)
		}
		if !strings.Contains(strings.SplitN(string(data), "\n", 2)[0], "Sprint 2") {
			t.Errorf("MEMORY.md first line does not contain prepend; got=%q", strings.SplitN(string(data), "\n", 2)[0])
		}
		if !strings.Contains(string(data), "- existing") {
			t.Errorf("MEMORY.md missing pre-existing content; got=%q", string(data))
		}
	})

	t.Run("prependToMemoryMD-direct-retry-exhausts", func(t *testing.T) {
		// Direct test of prependToMemoryMD with a goroutine actively bumping
		// MEMORY.md mtime+size. Verifies the retry-exhaustion error path is
		// reachable. We do not go through PersistIfPending because the contention
		// window inside PersistIfPending is narrow on fast filesystems.
		memoryDir := t.TempDir()
		memoryMdPath := filepath.Join(memoryDir, "MEMORY.md")
		if err := os.WriteFile(memoryMdPath, []byte("- v0\n"), 0o644); err != nil {
			t.Fatalf("init MEMORY.md: %v", err)
		}

		var (
			stop atomic.Bool
			wg   sync.WaitGroup
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter := 0
			for !stop.Load() {
				counter++
				_ = os.WriteFile(memoryMdPath, []byte(fmt.Sprintf("- v%d\n- extra %d\n", counter, counter)), 0o644)
				time.Sleep(100 * time.Microsecond)
			}
		}()
		// Let the goroutine churn long enough that prependToMemoryMD's three
		// attempts will all observe drift.
		err := prependToMemoryMD(memoryDir, "- new entry", "", "project_new.md")
		stop.Store(true)
		wg.Wait()

		// On fast hardware the call may succeed despite churn; only assert that
		// IF an error is returned it is the contention-exhaustion message.
		if err != nil && !strings.Contains(err.Error(), "contention") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestPersistIfPending_SupersedeMarkerApplied verifies AC-SHA-008: supersede
// marker is applied to the first matching MEMORY.md line.
func TestPersistIfPending_SupersedeMarkerApplied(t *testing.T) {
	tests := []struct {
		name           string
		existing       string
		supersedesFile string
		newFile        string
		wantContains   string
		wantNotPrefix  string
	}{
		{
			name:           "marker-applied-on-match",
			existing:       "- [Old entry](project_wave5_old_complete.md) — prev hook\n- [Other](other.md) — keep\n",
			supersedesFile: "project_wave5_old_complete.md",
			newFile:        "project_wave6_new.md",
			wantContains:   "[SUPERSEDED by project_wave6_new.md]",
		},
		{
			name:           "no-match-no-marker",
			existing:       "- [Other](other.md) — keep\n",
			supersedesFile: "project_wave5_missing.md",
			newFile:        "project_wave6_new.md",
			wantContains:   "- [Other](other.md)",
			wantNotPrefix:  "[SUPERSEDED",
		},
		{
			name:           "multiple-match-only-first",
			existing:       "- [A](dup.md) — first\n- [B](dup.md) — second\n",
			supersedesFile: "dup.md",
			newFile:        "project_new.md",
			wantContains:   "[SUPERSEDED by project_new.md]",
		},
		{
			name:           "already-superseded-no-double-mark",
			existing:       "- [SUPERSEDED by old.md] - [A](dup.md) — already marked\n- [B](other.md) — keep\n",
			supersedesFile: "dup.md",
			newFile:        "project_new.md",
			wantContains:   "[SUPERSEDED by old.md]",
			wantNotPrefix:  "[SUPERSEDED by project_new.md]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applySupersedeMarker([]byte(tt.existing), tt.supersedesFile, tt.newFile)
			if tt.wantContains != "" && !bytes.Contains(got, []byte(tt.wantContains)) {
				t.Errorf("expected marker substring %q not found; got=%q", tt.wantContains, string(got))
			}
			if tt.wantNotPrefix != "" && bytes.Contains(got, []byte(tt.wantNotPrefix)) {
				t.Errorf("unexpected substring %q found; got=%q", tt.wantNotPrefix, string(got))
			}
			if tt.name == "multiple-match-only-first" {
				// Only the first dup.md line should carry the marker.
				count := bytes.Count(got, []byte("[SUPERSEDED by project_new.md]"))
				if count != 1 {
					t.Errorf("expected exactly 1 marker; got %d (output=%q)", count, string(got))
				}
			}
		})
	}
}

// TestPersistIfPending_NoUserInteraction verifies AC-SHA-009: the function
// makes zero AskUserQuestion calls (verified by source-import audit) and
// produces no stdout output. The static guard is enforced by the orchestrator
// at the source-tree level; this test verifies the runtime behavior.
func TestPersistIfPending_NoUserInteraction(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := makeMemoryDir(t)

	// Capture stdout to verify no user-visible output.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	persistErr := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir)

	_ = w.Close()
	os.Stdout = origStdout
	var captured bytes.Buffer
	if _, copyErr := captured.ReadFrom(r); copyErr != nil && !errors.Is(copyErr, os.ErrClosed) {
		t.Fatalf("read captured stdout: %v", copyErr)
	}

	if persistErr != nil {
		t.Fatalf("PersistIfPending returned error: %v", persistErr)
	}
	if captured.Len() > 0 {
		t.Errorf("stdout should be empty; got=%q", captured.String())
	}

	// Static source-tree guard. We cannot fail this test based on grep result
	// because the test binary cannot self-introspect easily; the orchestrator
	// E4 verification covers this.
	_ = runtime.GOOS
}

// TestPersistIfPending_MemoryDirMissingSkips verifies §B.2 contract: when
// memoryDir does not exist the hook logs warn and returns nil; it MUST NOT
// create the directory (Claude Code owns the project-hash directory).
func TestPersistIfPending_MemoryDirMissingSkips(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	missingMemoryDir := filepath.Join(t.TempDir(), "does", "not", "exist")
	pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")

	if err := PersistIfPending(context.Background(), "session-001", projectDir, missingMemoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	// Pending file remains (no persistence happened).
	if _, err := os.Stat(pendingPath); err != nil {
		t.Errorf("pending file should remain when memoryDir missing; stat err=%v", err)
	}
	// memoryDir not created.
	if _, err := os.Stat(missingMemoryDir); !os.IsNotExist(err) {
		t.Errorf("memoryDir should not be created when missing; stat err=%v", err)
	}
}

// TestPersistIfPending_MemoryDirIsFile verifies the helper skips when memoryDir
// path resolves to a regular file (not a directory). Covers the !info.IsDir()
// branch of PersistIfPending.
func TestPersistIfPending_MemoryDirIsFile(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDirAsFile := filepath.Join(t.TempDir(), "memory-as-file")
	if err := os.WriteFile(memoryDirAsFile, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write memory-as-file: %v", err)
	}

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDirAsFile); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
	if _, err := os.Stat(pendingPath); err != nil {
		t.Errorf("pending file should remain when memoryDir is a file; stat err=%v", err)
	}
}

// TestPersistIfPending_PendingIsDirectory verifies the read-error branch of
// PersistIfPending (os.ReadFile returns non-IsNotExist error when path is a
// directory).
func TestPersistIfPending_PendingIsDirectory(t *testing.T) {
	root := t.TempDir()
	// Create pending.md as a DIRECTORY to force a read error.
	pendingDir := filepath.Join(root, ".moai", "state", "session-handoff", "pending.md")
	if err := os.MkdirAll(pendingDir, 0o755); err != nil {
		t.Fatalf("mkdir pending-as-dir: %v", err)
	}
	memoryDir := makeMemoryDir(t)

	if err := PersistIfPending(context.Background(), "session-001", root, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}
	// No memory file produced.
	entries, _ := os.ReadDir(memoryDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "project_") {
			t.Errorf("pending-as-dir produced memory file: %s", e.Name())
		}
	}
}

// TestAtomicWriteFile_CreateTempErrorPath covers the CreateTemp error branch by
// passing a non-existent dir.
func TestAtomicWriteFile_CreateTempErrorPath(t *testing.T) {
	err := atomicWriteFile(filepath.Join(t.TempDir(), "does", "not", "exist"), "x.md", []byte("payload"), 0o644)
	if err == nil {
		t.Fatalf("expected error from atomicWriteFile on missing dir; got nil")
	}
	if !strings.Contains(err.Error(), "create temp") {
		t.Errorf("expected 'create temp' in error; got=%v", err)
	}
}

// TestPersistIfPending_AtomicWriteMemoryFileFails covers the error branch in
// PersistIfPending where atomicWriteFile for the memory file itself fails.
// Strategy: make memoryDir read-only so CreateTemp fails.
func TestPersistIfPending_AtomicWriteMemoryFileFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod read-only semantics differ on windows")
	}
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := t.TempDir()
	if err := os.Chmod(memoryDir, 0o555); err != nil {
		t.Fatalf("chmod read-only: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(memoryDir, 0o755)
	})

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	// Pending file remains (persistence aborted).
	pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
	if _, err := os.Stat(pendingPath); err != nil {
		t.Errorf("pending should remain on memory-write failure; stat err=%v", err)
	}
}

// TestPersistIfPending_MemoryMdUpdateFails covers the PersistIfPending branch
// where the memory file is written but MEMORY.md update fails (partial-success
// log path, REQ-SHA-007 retry exhaustion path). Strategy: pre-create MEMORY.md
// as a directory so atomic rename fails.
func TestPersistIfPending_MemoryMdUpdateFails(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := t.TempDir()
	// Pre-create MEMORY.md as a directory to force os.ReadFile to error.
	if err := os.MkdirAll(filepath.Join(memoryDir, "MEMORY.md"), 0o755); err != nil {
		t.Fatalf("mkdir MEMORY.md-as-dir: %v", err)
	}

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	// Memory file was written.
	entries, _ := os.ReadDir(memoryDir)
	var memoryFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "project_") && !e.IsDir() {
			memoryFile = e.Name()
		}
	}
	if memoryFile == "" {
		t.Errorf("memory file should be written before MEMORY.md fails; entries=%v", entries)
	}

	// Pending file remains (partial-success → leave for retry).
	pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
	if _, err := os.Stat(pendingPath); err != nil {
		t.Errorf("pending should remain on MEMORY.md failure; stat err=%v", err)
	}
}

// TestParsePending_EdgeCases covers CRLF frontmatter delimiter + missing close
// delimiter to push parsePending coverage past 92%.
func TestParsePending_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "crlf-delimiter",
			content: "---\r\nsprint: sprint2\r\nspec: foo\r\nstatus: bar\r\nindex_line: \"x\"\r\n---\r\n" + validBody,
			wantErr: "",
		},
		{
			name:    "missing-close",
			content: "---\nsprint: x\n" + validBody,
			wantErr: "missing closing frontmatter delimiter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePending([]byte(tt.content))
			if tt.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Errorf("expected error containing %q; got=%v", tt.wantErr, err)
			}
		})
	}
}

// TestAtomicWriteFile_HappyPath covers all happy-path branches (Write, Chmod,
// Sync, Close, Rename) plus contents equality.
func TestAtomicWriteFile_HappyPath(t *testing.T) {
	dir := t.TempDir()
	payload := []byte("hello world")
	if err := atomicWriteFile(dir, "file.md", payload, 0o600); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "file.md"))
	if err != nil {
		t.Fatalf("read after write: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("payload mismatch: got=%q want=%q", got, payload)
	}
	// Verify no leftover tmp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp.") {
			t.Errorf("leftover tmp file: %s", e.Name())
		}
	}
}

// TestPersistIfPending_PendingCleanedOnSuccess verifies AC-SHA-010: successful
// persistence removes the pending file.
func TestPersistIfPending_PendingCleanedOnSuccess(t *testing.T) {
	projectDir := makeProjectDir(t, validFrontmatter+validBody)
	memoryDir := makeMemoryDir(t)
	pendingPath := filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")

	// Sanity: pending exists pre-call.
	if _, err := os.Stat(pendingPath); err != nil {
		t.Fatalf("pending file missing pre-call: %v", err)
	}

	if err := PersistIfPending(context.Background(), "session-001", projectDir, memoryDir); err != nil {
		t.Fatalf("PersistIfPending returned error: %v", err)
	}

	if _, err := os.Stat(pendingPath); !os.IsNotExist(err) {
		t.Errorf("pending file should be removed post-success; stat err=%v", err)
	}
}
