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

// sectionCommentLines는 파일의 주석 라인 시퀀스를 반환한다 (주석 보존 검증용).
func sectionCommentLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

// TestWriteSectionViaSeamRejectsM3ReclassifiedSections는 SPEC-WEBCONF-SIMPLIFY-001
// M3가 RouteExcluded로 재분류한 8개 전 seam 섹션에 대한 WriteSectionViaSeam 호출이
// 전부 오류로 거부되고, 디스크의 섹션 파일이 바이트 단위로 무변경임을 검증한다
// (REQ-WC-003 — config keys persist, web write path removed). M3 이전에는 이
// 섹션들이 seam-writable이었고 golden round-trip이 성공했다; M3 이후 웹 쓰기
// 경로가 제거되어 전원 거부된다.
func TestWriteSectionViaSeamRejectsM3ReclassifiedSections(t *testing.T) {
	t.Parallel()

	cases := []struct {
		section string
		edit    yamlpatch.KeyEdit
	}{
		{"workflow", yamlpatch.KeyEdit{Path: []string{"workflow", "execution_mode"}, Value: "auto"}},
		{"harness", yamlpatch.KeyEdit{Path: []string{"harness", "default_profile"}, Value: "strict"}},
		{"ralph", yamlpatch.KeyEdit{Path: []string{"ralph", "loop", "max_iterations"}, Value: "20"}},
		{"feedback", yamlpatch.KeyEdit{Path: []string{"feedback", "repository"}, Value: "example-org/fork"}},
		{"observability", yamlpatch.KeyEdit{Path: []string{"observability", "retention_days"}, Value: "60"}},
		{"security", yamlpatch.KeyEdit{Path: []string{"security", "permission", "strict_mode"}, Value: "true"}},
		{"handoff", yamlpatch.KeyEdit{Path: []string{"handoff", "mode"}, Value: "auto"}},
		{"cache", yamlpatch.KeyEdit{Path: []string{"cacheStrategy", "enabled"}, Value: "true"}},
	}

	for _, tc := range cases {
		t.Run(tc.section, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			before := seedSectionFixture(t, root, tc.section)

			err := WriteSectionViaSeam(root, tc.section, []yamlpatch.KeyEdit{tc.edit})
			if err == nil {
				t.Fatalf("WriteSectionViaSeam(%s): want rejection (M3 RouteExcluded), got nil", tc.section)
			}
			// 거부 경로는 파일을 건드리지 않는다 — config keys persist 무변경.
			if got := readSection(t, root, tc.section); got != before {
				t.Errorf("%s: rejected write mutated the section file (must be byte-unchanged — REQ-WC-003)", tc.section)
			}
		})
	}
}

// TestWriteSectionViaSeamRejectsHarnessLearningRoot는 harness.yaml의 두 번째
// 최상위 키(learning)에 대한 seam 쓰기도 M3 재분류로 거부됨을 검증한다.
func TestWriteSectionViaSeamRejectsHarnessLearningRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "harness")

	err := WriteSectionViaSeam(root, "harness", []yamlpatch.KeyEdit{
		{Path: []string{"learning", "auto_apply"}, Value: "true"},
	})
	if err == nil {
		t.Fatal("WriteSectionViaSeam(harness/learning): want rejection (M3 RouteExcluded), got nil")
	}
	if got := readSection(t, root, "harness"); got != before {
		t.Error("rejected harness write mutated harness.yaml (must be byte-unchanged)")
	}
}

// TestWriteSectionViaSeamRejectsCachePreservesFile은 cache.yaml에 대한 seam 쓰기가
// M3 재분류로 거부되고, 미노출 키(spec_ttl, min_cacheable_tokens)와 주석이
// 무변경임을 검증한다 (SPEC-WEB-CONSOLE-013 REQ-WC13-006 보존 불변식은 M3 이후에도
// 유효 — 거부 경로는 파일을 건드리지 않는다).
func TestWriteSectionViaSeamRejectsCachePreservesFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "cache")

	err := WriteSectionViaSeam(root, "cache", []yamlpatch.KeyEdit{
		{Path: []string{"cacheStrategy", "enabled"}, Value: "true"},
	})
	if err == nil {
		t.Fatal("WriteSectionViaSeam(cache): want rejection (M3 RouteExcluded), got nil")
	}
	after := readSection(t, root, "cache")
	// 미노출 키 + 주석 전량 무변경 (거부 경로는 파일을 건드리지 않는다).
	for _, keep := range []string{`spec_ttl: "5m"`, "min_cacheable_tokens: 2048", `session_ttl: "1h"`} {
		if !strings.Contains(after, keep) {
			t.Errorf("unexposed key line %q lost after rejected write:\n%s", keep, after)
		}
	}
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("cache.yaml comments not preserved verbatim after rejected write")
	}
}

// TestWriteSectionViaSeamRejectsNonSeamSections는 seam 대상이 아닌 섹션 —
// typed 경로(llm, git-strategy, quality), statusline, 제외군(state, tool-policy),
// 미지명 — 및 SPEC-WEBCONF-SIMPLIFY-001 M3로 재분류된 8개 전 seam 섹션이 전부
// 오류로 거부되고 어떤 파일도 생성/변경되지 않음을 검증한다.
func TestWriteSectionViaSeamRejectsNonSeamSections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	for _, section := range []string{
		// typed 경로.
		"llm", "git-strategy", "quality", "user", "language", "git-convention",
		// 전용 경로.
		"statusline",
		// 기존 제외군.
		"state", "system", "project", "sunset",
		"tool-policy", "lsp", "mx",
		"constitution", "context", "design", "interview",
		"research", "db",
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections reclassified to
		// RouteExcluded (tabs removed, web write path gone).
		"workflow", "harness", "ralph", "feedback", "observability", "security", "handoff", "cache",
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

// TestWriteSectionViaSeamRejectsResearchPreservesFile은 폐선된 research 섹션에
// 대한 seam 쓰기가 not-seam-writable 오류로 거부되고, 디스크의 research.yaml이
// 바이트 단위로 무변경임을 검증한다 (SPEC-WEB-CONSOLE-012 REQ-WC12-012).
func TestWriteSectionViaSeamRejectsResearchPreservesFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "research")

	err := WriteSectionViaSeam(root, "research", []yamlpatch.KeyEdit{
		{Path: []string{"research", "enabled"}, Value: "false"},
	})
	if err == nil {
		t.Fatal("want not-seam-writable rejection for research, got nil")
	}
	if got := readSection(t, root, "research"); got != before {
		t.Error("rejected research write mutated research.yaml (must be byte-unchanged)")
	}
}

// TestWriteSectionViaSeamRejectsForeignRootKey는 섹션 파일 밖의 최상위 키 주입
// (upsert 오남용)을 차단함을 검증한다. M3 이후 workflow는 RouteExcluded이므로
// 라우트 게이트에서 거부된다 (외래 키 검사 도달 전).
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
