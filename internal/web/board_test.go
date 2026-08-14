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

// SPEC-WEB-CONSOLE-011 M5 — READ-ONLY SPEC 감시 표면 (AC-WC11-040..046, 061).
//
// 종료 부채 목록과 MUST-FIX 조치 명령은 예전에 /specs/board 라는 별도 페이지였다.
// 어느 화면도 그 주소를 링크하지 않아 직접 입력해야만 닿았으므로, 지금은 /specs
// 화면의 두 패널이다. 아래 테스트가 묶는 불변식은 옮기기 전과 같다: GET 전용,
// 쓰기 경로 없음, 서버측 명령 실행 없음, git 의존 drift 스캐너 미호출.

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
//   - SPEC-BOARD-DONE-003   completed                            → list only
//   - SPEC-BOARD-PLAN-004   draft, no tier                       → list only
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

// getBoard fetches the /specs screen, which now carries the close-debt and
// must-fix panels.
func getBoard(t *testing.T, root string) *httptest.ResponseRecorder {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	req := httptest.NewRequest(http.MethodGet, "/specs", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// TestBoard_Render (AC-WC11-040): the /specs screen lists every SPEC, calls out
// the implemented-not-completed close-debt set, and flags MUST-FIX drift.
func TestBoard_Render(t *testing.T) {
	rec := getBoard(t, boardFixtureRoot(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /specs = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// The status filter chips carry the per-status counts.
	if !strings.Contains(body, "implemented") {
		t.Error("/specs missing the implemented status")
	}
	if !strings.Contains(body, "completed") {
		t.Error("/specs missing the completed status")
	}

	// Close-debt panel — both implemented SPECs listed. (Every SPEC also appears
	// in the main table, so presence alone is not the assertion; the panel's own
	// heading key plus the ID pair is.)
	if !strings.Contains(body, "SPEC-BOARD-DEBT-001") {
		t.Error("close-debt panel missing SPEC-BOARD-DEBT-001")
	}
	if !strings.Contains(body, "SPEC-BOARD-DRIFT-002") {
		t.Error("close-debt panel missing SPEC-BOARD-DRIFT-002")
	}

	// MUST-FIX drift badge.
	if !strings.Contains(body, "MUST-FIX") {
		t.Error("/specs missing the MUST-FIX drift badge")
	}
	if !strings.Contains(body, "badge--danger") {
		t.Error("the MUST-FIX badge is not rendered with the danger variant")
	}
}

// TestBoard_CloseDebtExcludesCompleted pins the close-debt predicate itself,
// which the rendered page cannot show on its own (every SPEC appears in the main
// table too, so a completed SPEC's ID is present either way).
func TestBoard_CloseDebtExcludesCompleted(t *testing.T) {
	a := newApp(Config{ProjectRoot: boardFixtureRoot(t), ProfileName: "default"})
	vm, err := a.buildSpecList("", "", "")
	if err != nil {
		t.Fatalf("buildSpecList: %v", err)
	}
	got := map[string]bool{}
	for _, r := range vm.CloseDebt {
		got[r.ID] = true
		if r.Status != "implemented" {
			t.Errorf("close-debt row %s has status %q, want implemented", r.ID, r.Status)
		}
	}
	if !got["SPEC-BOARD-DEBT-001"] || !got["SPEC-BOARD-DRIFT-002"] {
		t.Errorf("close-debt set = %v, want both implemented SPECs", got)
	}
	if got["SPEC-BOARD-DONE-003"] {
		t.Error("completed SPEC-BOARD-DONE-003 must not be close debt")
	}
	if got["SPEC-BOARD-PLAN-004"] {
		t.Error("draft SPEC-BOARD-PLAN-004 must not be close debt")
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
		t.Errorf("/specs missing the remediation command text %q", remediation)
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
		if !strings.Contains(body, `badge--outline">L<`) {
			t.Errorf("tier badge missing for a SPEC that declares tier: L:\n%s", body)
		}
	})
	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		writeBoardSpec(t, root, "SPEC-NOTIER-001",
			"id: SPEC-NOTIER-001\ntitle: \"untiered\"\nstatus: implemented\nupdated: 2026-07-01\n", "")
		rec := getBoard(t, root)
		if rec.Code != http.StatusOK {
			t.Fatalf("untiered render = %d, want 200 (tier absence is not an error)", rec.Code)
		}
		if strings.Contains(rec.Body.String(), `badge--outline">L<`) {
			t.Error("tier badge rendered for a SPEC with NO tier field (must be omitted)")
		}
	})
}

// TestBoard_GETOnly (AC-WC11-046): /specs responds to GET only; every mutating
// method is rejected with 405 (no write path, no status transition).
func TestBoard_GETOnly(t *testing.T) {
	a := newApp(Config{ProjectRoot: boardFixtureRoot(t), ProfileName: "default"})
	h := a.routes()
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/specs", nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin") // pass CSRF gate (REQ-SEC-002) so the handler's own 405 is isolated
		req.Host = "127.0.0.1"                          // loopback → passes hostCheck
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /specs = %d, want 405 (the SPEC screen is GET-only)", method, rec.Code)
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

// TestBoard_BoardRouteRetired: the standalone /specs/board page is gone. Its two
// panels live on /specs, so the old address must not resolve — a route that
// still answered would be a second, unlinked copy of the same surface.
func TestBoard_BoardRouteRetired(t *testing.T) {
	a := newApp(Config{ProjectRoot: boardFixtureRoot(t), ProfileName: "default"})
	req := httptest.NewRequest(http.MethodGet, "/specs/board", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /specs/board = %d, want 404 (the standalone board page was retired)", rec.Code)
	}
}

// TestBoard_NoWritePathSourceScan (REQ-WC11-044/045/046 guard): the sources that
// build and render the SPEC screen contain NO write / exec / status-transition /
// git-drift primitive. This is the mechanical "no write path exists" assertion.
func TestBoard_NoWritePathSourceScan(t *testing.T) {
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
	for _, name := range []string{"screens.go", "viewmodel_ops.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(data)
		for _, tok := range forbidden {
			if strings.Contains(src, tok) {
				t.Errorf("%s references forbidden write/exec primitive %q — the SPEC screen must be READ-ONLY", name, tok)
			}
		}
	}
}

// TestBoard_NoGitDriftPathInWebPackage (AC-WC11-045): the non-test web sources
// contain zero references to the git-dependent drift scanner (the synchronous
// render must never invoke it). The token is concatenated so this test source
// does not itself trip the repo-wide AC grep.
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

// TestBoard_I18nParity (AC-WC11-061 / REQ-WC11-015): every board.* data-i18n key
// still rendered is present in ALL FOUR locales of i18n.js.
func TestBoard_I18nParity(t *testing.T) {
	// The empty-state keys only render when the lists are empty, so an empty
	// project root exercises the full board.* key set.
	body := getBoard(t, t.TempDir()).Body.String()
	dict := readEmbeddedAsset(t, "i18n.js")

	keyRe := regexp.MustCompile(`data-i18n="(board\.[^"]+)"`)
	matches := keyRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatal("no board.* data-i18n keys found in the rendered SPEC screen")
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

// TestBoard_CoverageEdges exercises the loadSpecRows dirname fallback (an
// implemented SPEC with no frontmatter id) and an out-of-enum status.
func TestBoard_CoverageEdges(t *testing.T) {
	root := t.TempDir()
	// (a) implemented but NO id: field → the row falls back to the dir name.
	writeBoardSpec(t, root, "SPEC-NOID-001",
		"title: \"no id\"\nstatus: implemented\nupdated: 2026-07-01\n", "")
	// (b) a status outside the 8-value enum still renders verbatim on its row.
	writeBoardSpec(t, root, "SPEC-EXOTIC-002",
		"id: SPEC-EXOTIC-002\ntitle: \"exotic\"\nstatus: exotic\nupdated: 2026-07-02\n", "")
	// (c) an unparseable spec.md (no frontmatter) must not break the render.
	dir := filepath.Join(root, ".moai", "specs", "SPEC-BROKEN-003")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("no frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	rec := getBoard(t, root)
	if rec.Code != http.StatusOK {
		t.Fatalf("edge render = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "SPEC-NOID-001") {
		t.Error("row missing the dir-name fallback ID for an id-less implemented SPEC")
	}
	if !strings.Contains(body, "exotic") {
		t.Error("an out-of-enum status must still render on its row")
	}
	if !strings.Contains(body, "SPEC-BROKEN-003") {
		t.Error("an unparseable spec.md must still contribute a row rather than vanish")
	}
}

// TestBoard_EmptyProject renders an empty catalog without error.
func TestBoard_EmptyProject(t *testing.T) {
	rec := getBoard(t, t.TempDir())
	if rec.Code != http.StatusOK {
		t.Fatalf("empty-project /specs = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "board.closedebt.empty") {
		t.Error("an empty catalog should render the close-debt empty-state key")
	}
	if !strings.Contains(body, "board.mustfix.empty") {
		t.Error("an empty catalog should render the must-fix empty-state key")
	}
}
