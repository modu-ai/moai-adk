package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SPEC-WEB-CONSOLE-011 M5 — READ-ONLY SPEC board (AC-WC11-040..046, 061).
//
// The board is READ-ONLY: GET-only route, no write path, no server-side command
// execution, no status transition. These tests bind those invariants.

// writeBoardSpec writes root/.moai/specs/<id>/{spec.md,progress.md} for a board
// fixture. progressMD "" skips the progress.md write.
func writeBoardSpec(t *testing.T, root, id, frontmatter, progressMD string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	content := "---\n" + frontmatter + "---\n\n# " + id + "\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
	if progressMD != "" {
		if err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte(progressMD), 0o644); err != nil {
			t.Fatalf("write progress.md: %v", err)
		}
	}
}

// boardFixtureRoot builds a project root with a representative SPEC catalog:
//   - SPEC-BOARD-DEBT-001   implemented, tier L, no sync markers → close-debt only
//   - SPEC-BOARD-DRIFT-002  implemented, era V3R6, sync markers  → close-debt + MUST-FIX drift
//   - SPEC-BOARD-DONE-003   completed                            → distribution only
//   - SPEC-BOARD-PLAN-004   draft, no tier                       → distribution only
func boardFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeBoardSpec(t, root, "SPEC-BOARD-DEBT-001",
		"id: SPEC-BOARD-DEBT-001\ntitle: \"Debt one\"\nstatus: implemented\nupdated: 2026-07-01\ntier: L\nera: V3R6\n",
		"## §E.2 Run-phase Evidence\nrun evidence here\n")
	// A V3R6 SPEC with the full sync markers but status != completed → SyncStatusDrift.
	writeBoardSpec(t, root, "SPEC-BOARD-DRIFT-002",
		"id: SPEC-BOARD-DRIFT-002\ntitle: \"Drift two\"\nstatus: implemented\nupdated: 2026-07-02\nera: V3R6\n",
		"## §E.2 Run-phase Evidence\nrun\n\n## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: \"abc123def456\"\n")
	writeBoardSpec(t, root, "SPEC-BOARD-DONE-003",
		"id: SPEC-BOARD-DONE-003\ntitle: \"Done three\"\nstatus: completed\nupdated: 2026-07-02\nera: V3R6\n", "")
	writeBoardSpec(t, root, "SPEC-BOARD-PLAN-004",
		"id: SPEC-BOARD-PLAN-004\ntitle: \"Plan four\"\nstatus: draft\nupdated: 2026-07-03\n", "")
	return root
}

func getBoard(t *testing.T, root string) *httptest.ResponseRecorder {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	req := httptest.NewRequest(http.MethodGet, "/specs", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// TestBoard_Render (AC-WC11-040): the board renders the status distribution, the
// implemented-not-completed close-debt column, and MUST-FIX drift badges.
func TestBoard_Render(t *testing.T) {
	rec := getBoard(t, boardFixtureRoot(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /specs = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// Status distribution (the 8-value enum) — implemented count present.
	if !strings.Contains(body, "board-summary") {
		t.Error("board missing the status distribution summary section")
	}
	if !strings.Contains(body, "implemented") {
		t.Error("status distribution missing the implemented status")
	}
	if !strings.Contains(body, "completed") {
		t.Error("status distribution missing the completed status")
	}

	// Close-debt column — both implemented SPECs listed.
	if !strings.Contains(body, "SPEC-BOARD-DEBT-001") {
		t.Error("close-debt column missing SPEC-BOARD-DEBT-001")
	}
	if !strings.Contains(body, "SPEC-BOARD-DRIFT-002") {
		t.Error("close-debt column missing SPEC-BOARD-DRIFT-002")
	}
	// A completed SPEC must NOT be in close-debt.
	if strings.Contains(body, "SPEC-BOARD-DONE-003") {
		t.Error("completed SPEC-BOARD-DONE-003 must not appear in close-debt")
	}

	// MUST-FIX drift badge.
	if !strings.Contains(body, "MUST-FIX") {
		t.Error("board missing the MUST-FIX drift badge")
	}
	if !strings.Contains(body, "board-badge--mustfix") {
		t.Error("board missing the must-fix badge class")
	}
}

// TestBoard_RemediationCopyable (AC-WC11-043 / REQ-WC11-044): the remediation
// command renders as COPYABLE TEXT (a <code> block plus a data-copy button) and
// is NEVER executed server-side.
func TestBoard_RemediationCopyable(t *testing.T) {
	rec := getBoard(t, boardFixtureRoot(t))
	body := rec.Body.String()

	remediation := "moai spec close SPEC-BOARD-DRIFT-002 --backfill-only"
	if !strings.Contains(body, remediation) {
		t.Errorf("board missing the remediation command text %q", remediation)
	}
	// Rendered as copyable: a <code class="board-remedy"> plus a data-copy button.
	if !strings.Contains(body, "board-remedy") {
		t.Error("remediation not rendered inside a copyable <code class=\"board-remedy\">")
	}
	if !strings.Contains(body, `data-copy="`+remediation+`"`) {
		t.Errorf("remediation copy button missing data-copy=%q attribute", remediation)
	}
}

// TestBoard_TierBadge (AC-WC11-042): tier renders as an OPTIONAL badge — present
// when the frontmatter carries it, omitted (no error) when absent.
func TestBoard_TierBadge(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		root := t.TempDir()
		writeBoardSpec(t, root, "SPEC-TIER-001",
			"id: SPEC-TIER-001\ntitle: \"tiered\"\nstatus: implemented\nupdated: 2026-07-01\ntier: L\n", "")
		body := getBoard(t, root).Body.String()
		if !strings.Contains(body, "board-tier") {
			t.Error("tier badge (board-tier) missing for a SPEC that declares tier: L")
		}
	})
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		writeBoardSpec(t, root, "SPEC-NOTIER-001",
			"id: SPEC-NOTIER-001\ntitle: \"untiered\"\nstatus: implemented\nupdated: 2026-07-01\n", "")
		rec := getBoard(t, root)
		if rec.Code != http.StatusOK {
			t.Fatalf("untiered board render = %d, want 200 (tier absence is not an error)", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "board-tier") {
			t.Error("tier badge rendered for a SPEC with NO tier field (must be omitted)")
		}
	})
}

// TestBoard_GETOnly (AC-WC11-046): the board route responds to GET only; every
// mutating method is rejected with 405 (no write path, no status transition).
func TestBoard_GETOnly(t *testing.T) {
	a := newApp(Config{ProjectRoot: boardFixtureRoot(t), ProfileName: "default"})
	h := a.routes()
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/specs", nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin") // pass CSRF gate (REQ-SEC-002) so handleBoard's 405 is isolated
		req.Host = "127.0.0.1" // loopback → passes hostCheck; isolates handleBoard's own 405
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /specs = %d, want 405 (board is GET-only)", method, rec.Code)
		}
	}
	// GET is allowed.
	req := httptest.NewRequest(http.MethodGet, "/specs", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /specs = %d, want 200", rec.Code)
	}
}

// TestBoard_NoWritePathSourceScan (REQ-WC11-044/045/046 guard): the board source
// contains NO write / exec / status-transition / git-drift primitive. This is the
// mechanical "no write path exists" assertion.
func TestBoard_NoWritePathSourceScan(t *testing.T) {
	data, err := os.ReadFile("board.go")
	if err != nil {
		t.Fatalf("read board.go: %v", err)
	}
	src := string(data)
	// Tokens built via concatenation where naming the literal here would itself
	// trip the repo-wide "0 matches in internal/web" AC greps.
	forbidden := []string{
		"exec.Command", "os/exec", // no server-side command execution (REQ-WC11-044)
		"Detect" + "Drift",                                   // no git-dependent drift path (REQ-WC11-045)
		"Update" + "Status",                                  // no status transition (REQ-WC11-046)
		"WriteSectionViaSeam",                                // no config write seam
		"WritePreferences",                                   // no profile write
		"os.WriteFile", "os.Create", "os.Remove", "os.Mkdir", // no filesystem mutation
		"PatchFile", // no yaml patch write
	}
	for _, tok := range forbidden {
		if strings.Contains(src, tok) {
			t.Errorf("board.go references forbidden write/exec primitive %q — the board must be READ-ONLY", tok)
		}
	}
}

// TestBoard_NoGitDriftPathInWebPackage (AC-WC11-045): the non-test web sources
// contain zero references to the git-dependent drift scanner (the synchronous
// board render must never invoke it). The token is concatenated so this test
// source does not itself trip the repo-wide AC grep.
func TestBoard_NoGitDriftPathInWebPackage(t *testing.T) {
	gitDriftTok := "Detect" + "Drift"
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(data), gitDriftTok) {
			t.Errorf("%s references the git drift scanner — the web package must not invoke it (AC-WC11-045)", name)
		}
	}
}

// TestBoard_I18nParity (AC-WC11-061 / REQ-WC11-015): every board data-i18n key in
// the rendered board is present in ALL FOUR locales of i18n.js.
func TestBoard_I18nParity(t *testing.T) {
	body := getBoard(t, boardFixtureRoot(t)).Body.String()
	dict := readEmbeddedAsset(t, "i18n.js")

	keyRe := regexp.MustCompile(`data-i18n="(board\.[^"]+)"`)
	matches := keyRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatal("no board.* data-i18n keys found in the rendered board")
	}
	seen := map[string]bool{}
	for _, m := range matches {
		key := m[1]
		if seen[key] {
			continue
		}
		seen[key] = true
		// Each key must appear exactly once per locale (en/ko/ja/zh) → 4 total.
		if got := strings.Count(dict, `"`+key+`":`); got != 4 {
			t.Errorf("board i18n key %q appears %d times in i18n.js, want 4 (one per locale)", key, got)
		}
	}
}

// TestBoard_CoverageEdges exercises the boardSpecID dirname fallback (implemented
// SPEC with no frontmatter id) and orderedStatusCounts out-of-enum + unknown
// buckets (a non-enum status and an unparseable spec.md).
func TestBoard_CoverageEdges(t *testing.T) {
	root := t.TempDir()
	// (a) implemented but NO id: field → close-debt row falls back to the dir name.
	writeBoardSpec(t, root, "SPEC-NOID-001",
		"title: \"no id\"\nstatus: implemented\nupdated: 2026-07-01\n", "")
	// (b) a non-enum status → orderedStatusCounts "extras" (non-empty) branch.
	writeBoardSpec(t, root, "SPEC-EXOTIC-002",
		"id: SPEC-EXOTIC-002\ntitle: \"exotic\"\nstatus: exotic\nupdated: 2026-07-02\n", "")
	// (c) an unparseable spec.md (no frontmatter) → status "" → "(unknown)" bucket.
	dir := filepath.Join(root, ".moai", "specs", "SPEC-BROKEN-003")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("no frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	rec := getBoard(t, root)
	if rec.Code != http.StatusOK {
		t.Fatalf("edge board = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "SPEC-NOID-001") {
		t.Error("close-debt row missing the dir-name fallback ID for an id-less implemented SPEC")
	}
	if !strings.Contains(body, "exotic") {
		t.Error("status distribution missing the out-of-enum status bucket")
	}
	if !strings.Contains(body, "(unknown)") {
		t.Error("status distribution missing the (unknown) bucket for an unparseable spec.md")
	}
}

// TestBoard_EmptyProject renders an empty board (no .moai/specs) without error.
func TestBoard_EmptyProject(t *testing.T) {
	rec := getBoard(t, t.TempDir())
	if rec.Code != http.StatusOK {
		t.Fatalf("empty-project board = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "board.closedebt.empty") {
		t.Error("empty board should render the close-debt empty-state key")
	}
}
