package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// seedSectionFixture는 testdata/sections/<name>.yaml fixture를 임시 프로젝트
// 루트의 .moai/config/sections/ 아래로 복사하고 (루트 경로, 원본 내용)을
// 반환한다. 실제 프로젝트 파일은 절대 건드리지 않는다.
func seedSectionFixture(t *testing.T, root, name string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "sections", name+".yaml"))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), src, 0o644); err != nil {
		t.Fatalf("seed fixture copy: %v", err)
	}
	return string(src)
}

// readSection은 임시 루트의 섹션 파일 내용을 읽는다.
func readSection(t *testing.T, root, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", name+".yaml"))
	if err != nil {
		t.Fatalf("read back %s: %v", name, err)
	}
	return string(raw)
}

// sectionSplitLines / sectionChangedLines / sectionCommentLines는 golden diff
// 검증 헬퍼다 (yamlpatch 패키지의 테스트 헬퍼와 동일 로직 — 패키지 경계상 복제).
func sectionSplitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func sectionChangedLines(before, after string) (changed [][2]string, beforeOnly, afterOnly []string) {
	b := sectionSplitLines(before)
	a := sectionSplitLines(after)
	n := len(b)
	if len(a) < n {
		n = len(a)
	}
	prefix := 0
	for prefix < n && b[prefix] == a[prefix] {
		prefix++
	}
	if len(b) == len(a) {
		for i := prefix; i < n; i++ {
			if b[i] != a[i] {
				changed = append(changed, [2]string{b[i], a[i]})
			}
		}
		return changed, nil, nil
	}
	return nil, b[prefix:], a[prefix:]
}

func sectionCommentLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

// keySequence는 매핑 키 라인의 키 이름 시퀀스를 순서대로 추출한다 (키 순서 보존
// 검증용). 시퀀스 항목(-)과 주석/빈 줄은 제외한다.
func keySequence(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		if i := strings.Index(trimmed, ":"); i > 0 {
			out = append(out, trimmed[:i])
		}
	}
	return out
}

// nonBlankLines는 빈 줄(공백만 있는 줄 포함)을 제거한 라인 시퀀스를 반환한다.
func nonBlankLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

// TestYAMLPatchGoldenSections는 8개 seam 섹션 각각의 실제 파일 사본(golden
// fixture)에 스칼라 1개를 seam으로 기록하고 (i) diff가 편집 라인 1개에
// 국한되며 (ii) 주석 전량 (iii) 키 순서가 보존됨을 검증한다 (AC-WC11-017,
// REQ-WC11-017 — design.md §A.4 golden round-trip).
//
// 공백-only 정규화 고정 (design.md §A.4 사전 승인 경로, run-phase 실측):
// yaml.v3 재인코딩은 매핑 항목 사이의 빈 줄을 제거한다. 8개 섹션 중 항목-사이
// 빈 줄을 가진 security.yaml / db.yaml에서만 관측됐다 (blankNormalized=true로
// 범위 고정 — 빈 줄 외 모든 라인은 편집 라인 1개를 제외하고 byte-identical,
// 라인 신규 추가 0). 주석/키 소실이 아니므로 blocker가 아닌 허용 범위이며,
// 나머지 6개 섹션은 strict 1-line diff를 유지한다.
func TestYAMLPatchGoldenSections(t *testing.T) {
	t.Parallel()

	cases := []struct {
		section string
		edit    yamlpatch.KeyEdit
		want    string // 편집 라인이 담아야 할 부분 문자열
		// blankNormalized는 yaml.v3 빈 줄 제거 정규화가 관측된 섹션 표시다
		// (실측: security, db). true면 빈 줄 필터 후 1-line diff를 검증한다.
		blankNormalized bool
	}{
		{"workflow", yamlpatch.KeyEdit{Path: []string{"workflow", "team", "default_model"}, Value: "sonnet"}, "default_model: sonnet", false},
		{"harness", yamlpatch.KeyEdit{Path: []string{"harness", "default_profile"}, Value: "strict"}, "default_profile: strict", false},
		{"ralph", yamlpatch.KeyEdit{Path: []string{"ralph", "loop", "max_iterations"}, Value: "20"}, "max_iterations: 20", false},
		{"research", yamlpatch.KeyEdit{Path: []string{"research", "enabled"}, Value: "true"}, "enabled: true", false},
		{"feedback", yamlpatch.KeyEdit{Path: []string{"feedback", "repository"}, Value: "example-org/fork"}, "repository: example-org/fork", false},
		{"observability", yamlpatch.KeyEdit{Path: []string{"observability", "retention_days"}, Value: "60"}, "retention_days: 60", false},
		{"security", yamlpatch.KeyEdit{Path: []string{"security", "permission", "strict_mode"}, Value: "true"}, "strict_mode: true", true},
		{"db", yamlpatch.KeyEdit{Path: []string{"db", "orm"}, Value: "prisma"}, "orm: \"prisma\"", true},
	}

	for _, tc := range cases {
		t.Run(tc.section, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			before := seedSectionFixture(t, root, tc.section)

			if err := WriteSectionViaSeam(root, tc.section, []yamlpatch.KeyEdit{tc.edit}); err != nil {
				t.Fatalf("WriteSectionViaSeam(%s): %v", tc.section, err)
			}
			after := readSection(t, root, tc.section)

			cmpBefore, cmpAfter := before, after
			if tc.blankNormalized {
				// 정규화 범위 고정: 허용되는 유일한 라인 변동은 빈 줄 제거 —
				// 라인 신규 추가 0 + 비-빈 줄 수 동일.
				beforeLines := sectionSplitLines(before)
				afterLines := sectionSplitLines(after)
				if len(afterLines) > len(beforeLines) {
					t.Fatalf("%s: lines were ADDED (%d → %d) — beyond pinned blank-line normalization",
						tc.section, len(beforeLines), len(afterLines))
				}
				nbBefore := nonBlankLines(beforeLines)
				nbAfter := nonBlankLines(afterLines)
				if len(nbBefore) != len(nbAfter) {
					t.Fatalf("%s: non-blank line count drifted (%d → %d) — beyond pinned normalization",
						tc.section, len(nbBefore), len(nbAfter))
				}
				cmpBefore = strings.Join(nbBefore, "\n")
				cmpAfter = strings.Join(nbAfter, "\n")
			}

			changed, beforeOnly, afterOnly := sectionChangedLines(cmpBefore, cmpAfter)
			if len(beforeOnly) > 0 || len(afterOnly) > 0 {
				t.Fatalf("%s: line count drifted (normalization noise):\nbefore-only: %q\nafter-only: %q",
					tc.section, beforeOnly, afterOnly)
			}
			if len(changed) != 1 {
				t.Fatalf("%s: want exactly 1 changed line, got %d: %v", tc.section, len(changed), changed)
			}
			if !strings.Contains(changed[0][1], tc.want) {
				t.Errorf("%s: changed line = %q, want it to carry %q", tc.section, changed[0][1], tc.want)
			}

			// 주석 전량 보존 (AC-WC11-017).
			if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("%s: comments not preserved", tc.section)
			}
			// 키 순서 보존 (unknown key 포함 — 1-line diff가 이미 함의하나 명시 고정).
			if got, want := keySequence(after), keySequence(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("%s: key order drifted", tc.section)
			}
		})
	}
}

// TestYAMLPatchGoldenHarnessLearningRoot는 harness.yaml의 두 번째 최상위 키
// (learning)도 seam으로 기록 가능함을 검증한다 (sectionRootKeys 실측 반영).
func TestYAMLPatchGoldenHarnessLearningRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "harness")

	err := WriteSectionViaSeam(root, "harness", []yamlpatch.KeyEdit{
		{Path: []string{"learning", "auto_apply"}, Value: "true"},
	})
	if err != nil {
		t.Fatalf("WriteSectionViaSeam(harness/learning): %v", err)
	}
	after := readSection(t, root, "harness")
	changed, beforeOnly, afterOnly := sectionChangedLines(before, after)
	if len(beforeOnly) > 0 || len(afterOnly) > 0 || len(changed) != 1 {
		t.Fatalf("learning edit not line-scoped: changed=%v beforeOnly=%q afterOnly=%q", changed, beforeOnly, afterOnly)
	}
	if !strings.Contains(changed[0][1], "auto_apply: true") {
		t.Errorf("changed line = %q, want auto_apply: true", changed[0][1])
	}
}

// TestWriteSectionViaSeamRejectsNonSeamSections는 seam 대상이 아닌 섹션 —
// typed 경로(llm, git-strategy, quality), statusline, 제외군(state, tool-policy),
// 미지명 — 이 전부 오류로 거부되고 어떤 파일도 생성/변경되지 않음을 검증한다
// (AC-WC11-002 거부 케이스의 행동 계층, REQ-WC11-018).
func TestWriteSectionViaSeamRejectsNonSeamSections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	for _, section := range []string{
		"llm", "git-strategy", "quality", "user", "language", "git-convention",
		"statusline",
		"state", "system", "project", "cache", "sunset",
		"tool-policy", "lsp", "mx",
		"constitution", "context", "design", "interview",
		"nonexistent",
	} {
		err := WriteSectionViaSeam(root, section, []yamlpatch.KeyEdit{
			{Path: []string{section, "anything"}, Value: "x"},
		})
		if err == nil {
			t.Errorf("section %q: want rejection error, got nil", section)
		}
	}

	// 거부 경로는 파일을 일절 만들지 않는다.
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Errorf("rejected writes must not touch the filesystem: entries=%v err=%v", entries, err)
	}
}

// TestWriteSectionViaSeamRejectsForeignRootKey는 섹션 파일 밖의 최상위 키 주입
// (upsert 오남용)을 차단함을 검증한다.
func TestWriteSectionViaSeamRejectsForeignRootKey(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "workflow")

	err := WriteSectionViaSeam(root, "workflow", []yamlpatch.KeyEdit{
		{Path: []string{"llm", "mode"}, Value: "glm"},
	})
	if err == nil {
		t.Fatal("want rejection for foreign top-level key, got nil")
	}
	if got := readSection(t, root, "workflow"); got != before {
		t.Error("rejected edit mutated the section file")
	}
}

// TestWriteSectionViaSeamDBKeySplit는 REQ-WC11-019의 db 3/5 분리를 검증한다:
// 인터뷰 입력 3키는 기록되고, system 5키 및 미지명 키 기록 시도는 거부되며
// 파일 값이 불변이다 (AC-WC11-019의 데이터 계층 준비).
func TestWriteSectionViaSeamDBKeySplit(t *testing.T) {
	t.Parallel()

	t.Run("editable interview keys", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		seedSectionFixture(t, root, "db")

		err := WriteSectionViaSeam(root, "db", []yamlpatch.KeyEdit{
			{Path: []string{"db", "orm"}, Value: "prisma"},
			{Path: []string{"db", "multi_tenant"}, Value: "schema"},
			{Path: []string{"db", "migration_tool"}, Value: "alembic"},
		})
		if err != nil {
			t.Fatalf("editable keys rejected: %v", err)
		}
		after := readSection(t, root, "db")
		for _, want := range []string{"orm: \"prisma\"", "multi_tenant: \"schema\"", "migration_tool: \"alembic\""} {
			if !strings.Contains(after, want) {
				t.Errorf("editable key write missing %q", want)
			}
		}
	})

	t.Run("system keys read-only", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		before := seedSectionFixture(t, root, "db")

		for _, key := range DBSystemKeys() {
			err := WriteSectionViaSeam(root, "db", []yamlpatch.KeyEdit{
				{Path: []string{"db", key}, Value: "tampered"},
			})
			if err == nil {
				t.Errorf("system key %q: want rejection, got nil", key)
			}
		}
		if got := readSection(t, root, "db"); got != before {
			t.Error("system-key write attempts mutated db.yaml")
		}
	})

	t.Run("unknown or nested db paths rejected", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		before := seedSectionFixture(t, root, "db")

		for _, edit := range []yamlpatch.KeyEdit{
			{Path: []string{"db", "unknown_key"}, Value: "x"},
			{Path: []string{"db", "auto_sync", "enabled"}, Value: "false"},
			{Path: []string{"db"}, Value: "x"},
		} {
			if err := WriteSectionViaSeam(root, "db", []yamlpatch.KeyEdit{edit}); err == nil {
				t.Errorf("edit %v: want rejection, got nil", edit.Path)
			}
		}
		if got := readSection(t, root, "db"); got != before {
			t.Error("rejected db edits mutated db.yaml")
		}
	})
}

// TestWriteSectionViaSeamKeySplitConstantsMatchFixture는 3/5 키 상수가 실측
// db.yaml fixture의 키 집합과 정합함을 검증한다 (드리프트 가드).
func TestWriteSectionViaSeamKeySplitConstantsMatchFixture(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile(filepath.Join("testdata", "sections", "db.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(src)
	for _, key := range append(DBEditableKeys(), DBSystemKeys()...) {
		if !strings.Contains(content, "\n  "+key+":") {
			t.Errorf("db.yaml fixture missing top-level key %q", key)
		}
	}
	if got := len(DBEditableKeys()) + len(DBSystemKeys()); got != 8 {
		t.Errorf("db key split total = %d, want 8 (3 interview + 5 system)", got)
	}
}
