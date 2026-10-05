// memory_budget_test.go — the doctor's index measurement, byte/line budget
// and link-classification report (SPEC-MEMORY-FOLD-BUDGET-001 REQ-MFB-008,
// REQ-MFB-009, REQ-MFB-010, REQ-MFB-011; plan.md M2; acceptance
// AC-MFB-009 … AC-MFB-012).
//
// Every test drives `moai memory doctor` through its cobra command against a
// store under t.TempDir() or a read-only copy of the SPEC fixture — never
// the operator's real store (REQ-MFB-012). Assertions on the report decode
// the command's JSON into a local shape, so a key the doctor does not carry
// yet decodes as zero and fails as an intended assertion: the RED of each
// new behavior is observable on the pre-implementation tree, not hidden in
// a compile error.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
)

// doctorFindingJSON decodes one audit finding of the doctor's --json output.
// AuditFinding marshals without json tags, so the keys are the Go field
// names — pinned by acceptance.md §2 cell E2.
type doctorFindingJSON struct {
	Code   string `json:"Code"`
	Path   string `json:"Path"`
	Detail string `json:"Detail"`
}

// doctorReportJSON decodes one per-store report. The new keys
// (index_bytes … index_link_targets, byte_cap … line_cap) are the M2
// contract: before M2 they decode as zero, which is exactly the failing
// assertion the RED step needs.
type doctorReportJSON struct {
	Store struct {
		Dir    string `json:"dir"`
		Origin string `json:"origin"`
	} `json:"store"`
	Exists           bool                `json:"exists"`
	TopicFiles       int                 `json:"topic_files"`
	Cap              int                 `json:"cap"`
	IndexLines       int                 `json:"index_lines"`
	IndexBytes       int                 `json:"index_bytes"`
	IndexChars       int                 `json:"index_chars"`
	IndexLoadedChars int                 `json:"index_loaded_chars"`
	IndexLinkTargets int                 `json:"index_link_targets"`
	ByteCap          int                 `json:"byte_cap"`
	WarnPercent      int                 `json:"warn_percent"`
	LineCap          int                 `json:"line_cap"`
	Findings         []doctorFindingJSON `json:"findings"`
}

// runMemoryDoctor executes the doctor command with args and returns stdout.
func runMemoryDoctor(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	cmd := newMemoryDoctorCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory doctor %s: %v", strings.Join(args, " "), err)
	}
	return out.String()
}

// decodeMemoryDoctor decodes the --json output into the per-store reports.
func decodeMemoryDoctor(t *testing.T, out string) []doctorReportJSON {
	t.Helper()
	var reports []doctorReportJSON
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &reports); err != nil {
		t.Fatalf("decode doctor JSON: %v\noutput: %s", err, out)
	}
	return reports
}

// singleReport decodes a --dir run's output and returns its one report.
func singleReport(t *testing.T, out string) doctorReportJSON {
	t.Helper()
	reports := decodeMemoryDoctor(t, out)
	if len(reports) != 1 {
		t.Fatalf("doctor resolved %d stores, want 1 (--dir): %+v", len(reports), reports)
	}
	return reports[0]
}

// countFindingCode counts the findings carrying code.
func countFindingCode(findings []doctorFindingJSON, code string) int {
	n := 0
	for _, f := range findings {
		if f.Code == code {
			n++
		}
	}
	return n
}

// requireDoctorFlags is the RED-time existence check for the M2 flags: a
// missing flag fails here under its own name instead of surfacing later as
// a cobra unknown-flag error inside a boundary run.
func requireDoctorFlags(t *testing.T) {
	t.Helper()
	cmd := newMemoryDoctorCmd()
	for _, name := range []string{"byte-cap", "line-cap", "warn-percent"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("doctor flag --%s is not registered — the M2 budget flags do not exist yet", name)
		}
	}
}

const (
	fixtureMemorySHA  = "252722b91fb9b4f151bc30434656d0f59b29458490fb591ea16fa6ee1592c3d9"
	fixtureArchiveSHA = "c353c3c9adc3255c4a15064deb017ff59ad0052edea6df4232ae336418891105"
)

// specFixtureCopy copies the SPEC fixture (acceptance.md §1) read-only into
// a temporary directory, asserting its frozen file count, sizes and index
// hashes first, so a silent fixture edit cannot invalidate the numbers the
// tests assert on.
func specFixtureCopy(t *testing.T) string {
	t.Helper()
	src := filepath.Join(repoRoot(t), ".moai", "specs",
		"SPEC-MEMORY-FOLD-BUDGET-001", "fixtures", "store-A")
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", src, err)
	}
	if len(entries) != 14 {
		t.Fatalf("fixture holds %d files, want the frozen 14 — it changed under the SPEC", len(entries))
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatalf("read fixture file %s: %v", e.Name(), err)
		}
		switch e.Name() {
		case "MEMORY.md":
			if len(data) != 1155 {
				t.Fatalf("MEMORY.md is %d bytes, want the frozen 1155", len(data))
			}
			requireFixtureSHA(t, data, fixtureMemorySHA, "MEMORY.md")
		case "project_card_archive_2026_10.md":
			if len(data) != 373 {
				t.Fatalf("project_card_archive_2026_10.md is %d bytes, want the frozen 373", len(data))
			}
			requireFixtureSHA(t, data, fixtureArchiveSHA, "project_card_archive_2026_10.md")
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			t.Fatalf("copy %s: %v", e.Name(), err)
		}
	}
	return dst
}

// requireFixtureSHA pins one frozen fixture hash (acceptance.md §1).
func requireFixtureSHA(t *testing.T, data []byte, want, name string) {
	t.Helper()
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("%s SHA-256 = %s, want the frozen %s — the fixture changed under the SPEC", name, got, want)
	}
}

// seedIndexStore lays out a store under t.TempDir(): MEMORY.md with the
// given content plus one topic file per entry of topics.
func seedIndexStore(t *testing.T, indexContent string, topics map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	for name, body := range topics {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// storeHashes hashes every file of a store, keyed by name — the byte-equality
// witness for "the doctor rewrites nothing" (AC-MFB-012).
func storeHashes(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read store %s: %v", dir, err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		sum := sha256.Sum256(data)
		out[e.Name()] = hex.EncodeToString(sum[:])
	}
	return out
}

// requireBudgetFinding pins what REQ-MFB-009 demands of every budget
// finding: the axis, the measured value, the cap, the percentage and the
// basis sentence.
func requireBudgetFinding(t *testing.T, findings []doctorFindingJSON, code, axis, value, cap, pct string) {
	t.Helper()
	for _, f := range findings {
		if f.Code != code {
			continue
		}
		for _, part := range []string{axis, value, cap, pct, "raw bytes", "conservative proxy", "unconfirmed"} {
			if !strings.Contains(f.Detail, part) {
				t.Errorf("%s detail %q does not name %q", code, f.Detail, part)
			}
		}
		return
	}
	t.Fatalf("%s finding not found in %+v", code, findings)
}

// TestMemoryDoctor_Measures is AC-MFB-009: the doctor reports bytes,
// characters, loaded-content characters and lines, and the topic-file cap
// machinery is untouched.
func TestMemoryDoctor_Measures(t *testing.T) {
	t.Parallel()

	// (1) The frozen fixture: 1,155 / 1,095 / 1,095 / 17 (acceptance.md §1).
	dir := specFixtureCopy(t)
	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir))
	if !rep.Exists || rep.TopicFiles != 13 {
		t.Fatalf("fixture store: exists=%v topic_files=%d, want true/13", rep.Exists, rep.TopicFiles)
	}
	if rep.IndexBytes != 1155 {
		t.Errorf("index_bytes = %d, want 1155 (the file size)", rep.IndexBytes)
	}
	if rep.IndexChars != 1095 {
		t.Errorf("index_chars = %d, want 1095 (Unicode code points, not bytes)", rep.IndexChars)
	}
	if rep.IndexLoadedChars != 1095 {
		t.Errorf("index_loaded_chars = %d, want 1095 (the fixture index carries no frontmatter or comment)", rep.IndexLoadedChars)
	}
	if rep.IndexLines != 17 {
		t.Errorf("index_lines = %d, want 17 — the line count must not move (REQ-MFB-010)", rep.IndexLines)
	}

	// (2) An index of 3-byte characters: bytes exceed characters.
	hangul := strings.Repeat("한글", 20) + "\n" // 121 bytes, 41 code points
	dir3 := seedIndexStore(t, hangul, map[string]string{"feedback_a.md": "---\n---\n"})
	rep3 := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir3))
	if rep3.IndexBytes != 121 || rep3.IndexChars != 41 {
		t.Errorf("index_bytes=%d index_chars=%d, want 121/41 — bytes must exceed characters on 3-byte runes", rep3.IndexBytes, rep3.IndexChars)
	}

	// (3) The leading frontmatter block and HTML comments are excluded from
	// the loaded-content count, and from nothing else.
	front := "---\nname: x\ndescription: d\ntype: feedback\n---\n"
	body := "body text\n<!-- hidden from the loader -->\n"
	dirF := seedIndexStore(t, front+body, nil)
	repF := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dirF))
	if want := utf8.RuneCountInString("body text\n\n"); repF.IndexLoadedChars != want {
		t.Errorf("index_loaded_chars = %d, want %d (frontmatter and comment removed)", repF.IndexLoadedChars, want)
	}
	if repF.IndexChars == repF.IndexLoadedChars {
		t.Errorf("index_chars=%d equals index_loaded_chars=%d; the exclusions were not removed", repF.IndexChars, repF.IndexLoadedChars)
	}

	// (4) The text render shows the same four figures.
	text := runMemoryDoctor(t, "--dir", dir)
	for _, want := range []string{"1155", "1095", "17"} {
		if !strings.Contains(text, want) {
			t.Errorf("text render does not show %s:\n%s", want, text)
		}
	}
	if strings.Count(text, "1095") < 2 {
		t.Errorf("text render shows 1095 fewer than twice (chars and loaded chars):\n%s", text)
	}
}

// TestMemoryDoctor_TopicCapUnchanged is the regression-guard half of
// AC-MFB-009: the topic-file count, the cap (50 by default, --cap to
// override) and MEMORY_TOPIC_COUNT_OVER_CAP behave exactly as at the pinned
// tree. E2 recorded "cap":50 and "topic_files":13 today; this test keeps
// them there.
func TestMemoryDoctor_TopicCapUnchanged(t *testing.T) {
	t.Parallel()
	names := []string{"feedback_a.md", "feedback_b.md", "feedback_c.md", "feedback_d.md"}
	topics := map[string]string{}
	idx := "# Memory Index\n\n"
	for _, n := range names {
		topics[n] = "---\n---\n"
		idx += "- [T](" + n + ") — hook\n"
	}
	dir := seedIndexStore(t, idx, topics)

	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--cap", "3"))
	if rep.Cap != 3 || rep.TopicFiles != 4 {
		t.Fatalf("cap=%d topic_files=%d, want 3/4", rep.Cap, rep.TopicFiles)
	}
	if n := countFindingCode(rep.Findings, "MEMORY_TOPIC_COUNT_OVER_CAP"); n != 1 {
		t.Errorf("four topic files of a 3-file cap = %d MEMORY_TOPIC_COUNT_OVER_CAP, want 1: %+v", n, rep.Findings)
	}

	repDef := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir))
	if repDef.Cap != config.DefaultMemoryTopicFileCap {
		t.Errorf("flagless cap = %d, want config.DefaultMemoryTopicFileCap (%d)", repDef.Cap, config.DefaultMemoryTopicFileCap)
	}
	if n := countFindingCode(repDef.Findings, "MEMORY_TOPIC_COUNT_OVER_CAP"); n != 0 {
		t.Errorf("four files under the default cap must not warn: %+v", repDef.Findings)
	}
}

// TestMemoryDoctor_BudgetBoundaries is AC-MFB-010: the budget findings fire
// on the configured boundaries — integer comparisons
// value*100 >= warnPercent*cap — and --line-cap governs both the budget
// axis and MEMORY_INDEX_OVERFLOW, so one line cap is in force per
// invocation.
func TestMemoryDoctor_BudgetBoundaries(t *testing.T) {
	t.Parallel()
	requireDoctorFlags(t)

	// The fixture index: 1,155 bytes, 17 lines.
	dir := specFixtureCopy(t)

	// --byte-cap 1444: 79.98 % — no budget finding.
	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--byte-cap", "1444"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN") + countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_AT_CAP"); n != 0 {
		t.Errorf("79.98%% of cap emitted %d budget finding(s): %+v", n, rep.Findings)
	}

	// --byte-cap 1443: 80.04 % — the bytes-axis warning.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--byte-cap", "1443"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 1 {
		t.Fatalf("80.04%% of cap = %d MEMORY_INDEX_BUDGET_WARN, want 1: %+v", n, rep.Findings)
	}
	requireBudgetFinding(t, rep.Findings, "MEMORY_INDEX_BUDGET_WARN", "bytes", "1155", "1443", "80")

	// The text render carries the same finding's detail — axis, value, cap,
	// percentage and the proxy basis are readable without --json
	// (REQ-MFB-009: the doctor reports in text and in --json).
	textWarn := runMemoryDoctor(t, "--dir", dir, "--byte-cap", "1443")
	for _, part := range []string{
		"index bytes: 1155 of a 1443-bytes cap (80%)",
		"raw bytes: conservative proxy; the loader's cut is unconfirmed",
	} {
		if !strings.Contains(textWarn, part) {
			t.Errorf("text render does not carry the budget detail %q:\n%s", part, textWarn)
		}
	}

	// --byte-cap 1155: 100 % — AT_CAP replaces the warning.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--byte-cap", "1155"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_AT_CAP"); n != 1 {
		t.Fatalf("100%% of cap = %d MEMORY_INDEX_BUDGET_AT_CAP, want 1: %+v", n, rep.Findings)
	}
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 0 {
		t.Errorf("AT_CAP must replace the warning, but %d WARN remain: %+v", n, rep.Findings)
	}
	requireBudgetFinding(t, rep.Findings, "MEMORY_INDEX_BUDGET_AT_CAP", "bytes", "1155", "1155", "100")

	// --line-cap 22: 17 lines is 77.3 % — nothing.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--line-cap", "22"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN") + countFindingCode(rep.Findings, "MEMORY_INDEX_OVERFLOW"); n != 0 {
		t.Errorf("17 lines of a 22-line cap (77.3%%) emitted findings: %+v", rep.Findings)
	}

	// --line-cap 21: 81.0 % — a lines-axis budget warning.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--line-cap", "21"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 1 {
		t.Fatalf("17 lines of a 21-line cap (81.0%%) = %d lines-axis WARN, want 1: %+v", n, rep.Findings)
	}
	requireBudgetFinding(t, rep.Findings, "MEMORY_INDEX_BUDGET_WARN", "lines", "17", "21", "80")

	// --line-cap 17: lines equal to the cap, not above it — the lines-axis
	// warning fires and MEMORY_INDEX_OVERFLOW does not.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--line-cap", "17"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 1 {
		t.Fatalf("17 lines of a 17-line cap = %d lines-axis WARN, want 1 (at the cap, not above): %+v", n, rep.Findings)
	}
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_OVERFLOW"); n != 0 {
		t.Errorf("17 lines at a 17-line cap must not overflow: %+v", rep.Findings)
	}

	// --line-cap 16: lines above the cap — MEMORY_INDEX_OVERFLOW and no
	// lines-axis warning. The same flag governs both checks.
	rep = singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--line-cap", "16"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_OVERFLOW"); n != 1 {
		t.Fatalf("17 lines above a 16-line cap = %d MEMORY_INDEX_OVERFLOW, want 1: %+v", n, rep.Findings)
	}
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 0 {
		t.Errorf("lines above the cap belong to MEMORY_INDEX_OVERFLOW, not the budget warning: %+v", rep.Findings)
	}

	// The two line counts agree where it matters: "line\n\n\n" reads as 1
	// line in the display count but 3 in the audit count — the budget axis
	// must judge the audit count, so the overflow finding REPLACES the
	// warning instead of both firing.
	div := seedIndexStore(t, "line\n\n\n", nil)
	repDiv := singleReport(t, runMemoryDoctor(t, "--json", "--dir", div, "--line-cap", "1"))
	if n := countFindingCode(repDiv.Findings, "MEMORY_INDEX_OVERFLOW"); n != 1 {
		t.Errorf("audit-lines 3 above a 1-line cap = %d MEMORY_INDEX_OVERFLOW, want 1: %+v", n, repDiv.Findings)
	}
	if n := countFindingCode(repDiv.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 0 {
		t.Errorf("overflow must replace the lines-axis warning, not accompany it: %+v", repDiv.Findings)
	}

	// Flagless, the constants are followed by reference, never by literal:
	// one byte below the warn point of DefaultMemoryIndexByteCap is silent,
	// exactly at it warns. Each generated index is the x-run plus one
	// trailing newline, so the run length is totalBytes-1.
	warnAt := config.DefaultMemoryIndexByteCap * config.DefaultMemoryIndexWarnPercent / 100
	below := seedIndexStore(t, strings.Repeat("x", warnAt-2)+"\n", map[string]string{"feedback_a.md": "---\n---\n"})
	repBelow := singleReport(t, runMemoryDoctor(t, "--json", "--dir", below))
	if n := countFindingCode(repBelow.Findings, "MEMORY_INDEX_BUDGET_WARN") + countFindingCode(repBelow.Findings, "MEMORY_INDEX_BUDGET_AT_CAP"); n != 0 {
		t.Errorf("one byte below the warn point emitted %d finding(s): %+v", n, repBelow.Findings)
	}

	at := seedIndexStore(t, strings.Repeat("x", warnAt-1)+"\n", map[string]string{"feedback_a.md": "---\n---\n"})
	repAt := singleReport(t, runMemoryDoctor(t, "--json", "--dir", at))
	if n := countFindingCode(repAt.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 1 {
		t.Fatalf("exactly at the warn point = %d MEMORY_INDEX_BUDGET_WARN, want 1: %+v", n, repAt.Findings)
	}
	if repAt.ByteCap != config.DefaultMemoryIndexByteCap ||
		repAt.WarnPercent != config.DefaultMemoryIndexWarnPercent ||
		repAt.LineCap != config.DefaultMemoryIndexLineCap {
		t.Errorf("budget report fields = (byte %d, warn %d, line %d), want the config constants (%d, %d, %d)",
			repAt.ByteCap, repAt.WarnPercent, repAt.LineCap,
			config.DefaultMemoryIndexByteCap, config.DefaultMemoryIndexWarnPercent, config.DefaultMemoryIndexLineCap)
	}
}

// TestMemoryDoctor_BytesProxyWarns is AC-MFB-011: the warning keys on raw
// bytes and says so — a 3-byte-character index warns although its character
// count alone would not.
func TestMemoryDoctor_BytesProxyWarns(t *testing.T) {
	t.Parallel()
	requireDoctorFlags(t)

	// 41 × "한글" plus the trailing newline = 247 bytes but 83 code points:
	// with --byte-cap 300 the bytes sit at 82 % of the cap while the
	// characters sit at 27 % — only the raw-byte axis can warn here
	// (plan.md OD-3).
	content := strings.Repeat("한글", 41) + "\n"
	dir := seedIndexStore(t, content, map[string]string{"feedback_a.md": "---\n---\n"})

	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir, "--byte-cap", "300"))
	if n := countFindingCode(rep.Findings, "MEMORY_INDEX_BUDGET_WARN"); n != 1 {
		t.Fatalf("247 bytes of a 300-byte cap (82%%) = %d MEMORY_INDEX_BUDGET_WARN, want 1 although the characters alone would not warn: %+v",
			n, rep.Findings)
	}
	requireBudgetFinding(t, rep.Findings, "MEMORY_INDEX_BUDGET_WARN", "bytes", "247", "300", "82")

	// The JSON carries both measures, so the divergence is visible.
	if rep.IndexBytes != 247 || rep.IndexChars != 83 {
		t.Errorf("index_bytes=%d index_chars=%d, want 247/83 — both keys must be present", rep.IndexBytes, rep.IndexChars)
	}
}

// TestMemoryDoctor_LinkClasses is AC-MFB-012: the doctor classifies links,
// never collapses distinct targets, names the count by key, and rewrites
// nothing.
func TestMemoryDoctor_LinkClasses(t *testing.T) {
	t.Parallel()
	dir := specFixtureCopy(t)
	before := storeHashes(t, dir)

	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir))

	// Exactly two repo-relative findings, naming the two distinct targets
	// in full — they share the base name verdict.md and must not collapse
	// (the doctor keys by full text, never by base name).
	if n := countFindingCode(rep.Findings, "MEMORY_REPO_RELATIVE_LINK"); n != 2 {
		t.Fatalf("MEMORY_REPO_RELATIVE_LINK findings = %d, want exactly 2: %+v", n, rep.Findings)
	}
	for _, target := range []string{".moai/reports/t9002/verdict.md", ".moai/reports/t9006/verdict.md"} {
		found := false
		for _, f := range rep.Findings {
			if f.Code == "MEMORY_REPO_RELATIVE_LINK" && strings.Contains(f.Detail, target) {
				found = true
			}
		}
		if !found {
			t.Errorf("no MEMORY_REPO_RELATIVE_LINK names %s in full: %+v", target, rep.Findings)
		}
	}

	// No dangling finding names verdict.md …
	for _, f := range rep.Findings {
		if f.Code == "MEMORY_DANGLING_INDEX_LINK" && strings.Contains(f.Detail, "verdict.md") {
			t.Errorf("repo-relative target reported as dangling: %+v", f)
		}
	}
	// … and exactly one dangling finding remains, naming the one truly
	// missing store-local target, with no replacement suggestion
	// (REQ-MFB-012: the doctor never suggests a replacement).
	dangling := 0
	for _, f := range rep.Findings {
		if f.Code != "MEMORY_DANGLING_INDEX_LINK" {
			continue
		}
		dangling++
		if want := "index links feedback_queue_jump_window.md but no such file exists"; f.Detail != want {
			t.Errorf("dangling detail = %q, want exactly %q — no suggestion text", f.Detail, want)
		}
	}
	if dangling != 1 {
		t.Errorf("MEMORY_DANGLING_INDEX_LINK findings = %d, want exactly 1: %+v", dangling, rep.Findings)
	}

	// The distinct MEMORY.md link targets, counted by full text.
	if rep.IndexLinkTargets != 11 {
		t.Errorf("index_link_targets = %d, want 11 (the two verdict.md paths are distinct by full text)", rep.IndexLinkTargets)
	}

	// The text render carries the same figure and each repo-relative link's
	// full path (REQ-MFB-011: report in text and in --json).
	text := runMemoryDoctor(t, "--dir", dir)
	if !strings.Contains(text, "link targets: 11") {
		t.Errorf("text render does not show the 11 link targets:\n%s", text)
	}
	for _, target := range []string{".moai/reports/t9002/verdict.md", ".moai/reports/t9006/verdict.md"} {
		if !strings.Contains(text, "link target "+target+" is repo-relative") {
			t.Errorf("text render does not name %s in full:\n%s", target, text)
		}
	}

	// The doctor rewrites nothing (the doctor half of REQ-MFB-012).
	after := storeHashes(t, dir)
	if len(before) != len(after) {
		t.Fatalf("store file list changed: %d files before, %d after", len(before), len(after))
	}
	for name, sum := range before {
		if after[name] != sum {
			t.Errorf("%s changed across the doctor run", name)
		}
	}
}

// TestMemoryDoctor_IndexSelfLinkNotDangling is the codex-review round-2 P2
// regression (card t1502): a store whose only file is MEMORY.md, indexing
// itself with a `[Index](MEMORY.md)` link — the index IS present, so the
// class-aware dangling audit must not report `index links MEMORY.md but no
// such file exists`.
func TestMemoryDoctor_IndexSelfLinkNotDangling(t *testing.T) {
	dir := seedFoldStore(t, "# Memory Index\n\n- [Index](MEMORY.md) — the index names itself\n", map[string]string{})
	rep := singleReport(t, runMemoryDoctor(t, "--json", "--dir", dir))
	for _, f := range rep.Findings {
		if f.Code == "MEMORY_DANGLING_INDEX_LINK" {
			t.Errorf("MEMORY.md self-link reported as dangling: %+v", f)
		}
	}
}
