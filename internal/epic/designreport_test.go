package epic

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseDesignReport_BASSliceTable verifies the parser extracts the M0..M5
// canonical list + labels from the BAS design report slice table. The fixture
// mirrors the live `.moai/reports/navigator-redesign-bas-20260805.html` §7
// slice-table structure.
func TestParseDesignReport_BASSliceTable(t *testing.T) {
	html := `<!DOCTYPE html>
<html><body>
<h2>6. 구현 로드맵 — 마일스톤</h2>
<p>lead text</p>
<h2>7. 슬라이스 — 범위·산출물·의존성</h2>
<table>
<thead><tr><th>슬라이스</th><th>핵심 산출물</th><th>의존</th><th>Tier 분류</th></tr></thead>
<tbody>
<tr><td>M0 그래프 결합층</td><td>SSOT 토큰 명세, 결합 그래프 스키마</td><td>—</td><td>L</td></tr>
<tr><td>M1 Detect 훅</td><td>PostToolUse hook, path→rows 매핑</td><td>M0</td><td>M</td></tr>
<tr><td>M2 Route 승격</td><td>audit → work item 변환, owner=path 결합</td><td>M0 (M1과 병렬)</td><td>M</td></tr>
<tr><td>M3 Fix 증분 갱신</td><td>AI-drafted rewrite, --compare-to 패턴</td><td>M1+M2</td><td>L</td></tr>
<tr><td>M4 4-tier 지도</td><td>blueprint-first 모드, astx 확장</td><td>M0</td><td>L</td></tr>
<tr><td>M5 Brownfield 역추출</td><td>document --code 스타일 추출기</td><td>M4</td><td>M</td></tr>
</tbody>
</table>
<h2>8. 리스크 그리드</h2>
</body></html>`

	path := filepath.Join(t.TempDir(), "report.html")
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := ParseDesignReport(path)
	if err != nil {
		t.Fatalf("ParseDesignReport: %v", err)
	}
	if len(got.Milestones) != 6 {
		t.Fatalf("expected 6 milestones, got %d (%+v)", len(got.Milestones), got.Milestones)
	}
	wantIDs := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	for i, want := range wantIDs {
		if got.Milestones[i].ID != want {
			t.Errorf("milestone[%d].ID = %q, want %q", i, got.Milestones[i].ID, want)
		}
	}
	// label is the cell text after "M0 " through the first < or end.
	if got.Milestones[0].Label != "그래프 결합층" {
		t.Errorf("M0 label = %q, want 그래프 결합층", got.Milestones[0].Label)
	}
}

// TestParseDesignReport_FailOpenOnMalformed verifies KI-1 / edge case E3: a
// design report without the slice table returns no milestones, NOT an error.
func TestParseDesignReport_FailOpenOnMalformed(t *testing.T) {
	html := `<html><body><h2>1. 왜 다시 설계하는가</h2><p>no slice table here</p></body></html>`
	path := filepath.Join(t.TempDir(), "report.html")
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	got, err := ParseDesignReport(path)
	if err != nil {
		t.Fatalf("malformed report should fail-open, got: %v", err)
	}
	if len(got.Milestones) != 0 {
		t.Errorf("expected 0 milestones on malformed report, got %d", len(got.Milestones))
	}
}

// TestDiscoverDesignReport_NamingRule verifies design.md §4: the discovery
// rule matches `<basename>-<token-lowercased>-<stamp>.html` and returns the
// lexicographically-first match.
func TestDiscoverDesignReport_NamingRule(t *testing.T) {
	dir := t.TempDir()
	// Three files; only the two with `-bas-` should match.
	mustWrite(t, filepath.Join(dir, "navigator-redesign-bas-20260805.html"), "<html/>")
	mustWrite(t, filepath.Join(dir, "zzz-extra-bas-20260901.html"), "<html/>")
	mustWrite(t, filepath.Join(dir, "unrelated-report.html"), "<html/>")

	got, err := DiscoverDesignReport("BAS", dir)
	if err != nil {
		t.Fatalf("DiscoverDesignReport: %v", err)
	}
	// lexicographically-first match: "navigator-redesign-bas-..." < "zzz-extra-bas-..."
	want := filepath.Join(dir, "navigator-redesign-bas-20260805.html")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestDiscoverDesignReport_NoMatch verifies the fail-open path.
func TestDiscoverDesignReport_NoMatch(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "unrelated.html"), "<html/>")

	got, err := DiscoverDesignReport("BAS", dir)
	if err != nil {
		t.Fatalf("expected fail-open empty path, got error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty path on no match, got %q", got)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
