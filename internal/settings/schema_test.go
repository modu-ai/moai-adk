package settings

import (
	"os/exec"
	"strings"
	"testing"
)

// expected34FieldNames는 6개 섹션에 걸친 정규 34개 필드명을 렌더 순서대로 나열한다.
// 이 목록이 스키마의 단일 진실 검사 기준이다(AC-WC10-010).
var expected34FieldNames = []string{
	// Identity (1)
	"user_name",
	// Language (4)
	"conversation_lang", "git_commit_lang", "code_comment_lang", "doc_lang",
	// Launch (4)
	"model", "model_policy", "effort_level", "permission_mode",
	// Statusline (16: theme + 15 segments)
	"statusline_theme",
	"statusline_segment.claude_version",
	"statusline_segment.context",
	"statusline_segment.directory",
	"statusline_segment.effort_thinking",
	"statusline_segment.git_branch",
	"statusline_segment.git_status",
	"statusline_segment.moai_version",
	"statusline_segment.model",
	"statusline_segment.output_style",
	"statusline_segment.pr",
	"statusline_segment.session_time",
	"statusline_segment.task",
	"statusline_segment.usage_5h",
	"statusline_segment.usage_7d",
	"statusline_segment.worktree",
	// Quality (4)
	"development_mode",
	"quality.test_coverage_target",
	"quality.enforce_quality",
	"quality.tdd_settings.min_coverage_per_commit",
	// Git Convention (5)
	"git_convention",
	"git_convention.auto_detection.enabled",
	"git_convention.auto_detection.confidence_threshold",
	"git_convention.auto_detection.sample_size",
	"git_convention.validation.enforce_on_push",
}

// TestSchemaFieldNameSet은 스키마가 정확히 34개 필드를 6개 섹션에 걸쳐 열거하고,
// 필드명 집합이 기대 집합과 정확히 일치함을 검증한다(AC-WC10-010).
func TestSchemaFieldNameSet(t *testing.T) {
	got := FieldNames()

	if len(got) != 34 {
		t.Fatalf("FieldNames() count = %d, want 34", len(got))
	}

	gotSet := make(map[string]bool, len(got))
	for _, n := range got {
		if gotSet[n] {
			t.Errorf("duplicate field name in schema: %q", n)
		}
		gotSet[n] = true
	}

	wantSet := make(map[string]bool, len(expected34FieldNames))
	for _, n := range expected34FieldNames {
		wantSet[n] = true
	}

	for n := range gotSet {
		if !wantSet[n] {
			t.Errorf("schema has unexpected field %q (not in expected 34-field set)", n)
		}
	}
	for n := range wantSet {
		if !gotSet[n] {
			t.Errorf("schema missing expected field %q", n)
		}
	}
}

// TestSchemaSixSections는 6개 정규 섹션이 모두 존재하고, 각 섹션의 필드 수가
// 기대값과 일치함을 검증한다(AC-WC10-003 섹션 구성).
func TestSchemaSixSections(t *testing.T) {
	wantCounts := map[SectionID]int{
		SectionIdentity:      1,
		SectionLanguage:      4,
		SectionLaunch:        4,
		SectionStatusline:    16,
		SectionQuality:       4,
		SectionGitConvention: 5,
	}

	sections := AllSections()
	if len(sections) != 6 {
		t.Fatalf("AllSections() count = %d, want 6", len(sections))
	}

	total := 0
	for _, s := range sections {
		got := len(SectionFields(s))
		want, ok := wantCounts[s]
		if !ok {
			t.Errorf("unexpected section %q", s)
			continue
		}
		if got != want {
			t.Errorf("section %q field count = %d, want %d", s, got, want)
		}
		total += got
	}
	if total != 34 {
		t.Errorf("sum of section field counts = %d, want 34", total)
	}
}

// TestSchemaPerFieldInvariants는 모든 필드가 비어있지 않은 i18n 키와 영속화 대상을
// 가짐을 검증한다(REQ-WC10-002). select/multi-select 필드는 비어있지 않은 옵션
// 목록을, ProfileStore 필드는 비어있지 않은 Field 를, ProjectConfig 필드는
// 비어있지 않은 Section+Key 를 가져야 한다.
func TestSchemaPerFieldInvariants(t *testing.T) {
	for _, f := range AllFields() {
		if f.Name == "" {
			t.Errorf("field with empty Name found")
			continue
		}
		if f.I18nKey == "" {
			t.Errorf("field %q has empty I18nKey", f.Name)
		}
		if f.Section == "" {
			t.Errorf("field %q has empty Section", f.Name)
		}
		switch f.Persist.Kind {
		case PersistProfileStore:
			if f.Persist.Field == "" {
				t.Errorf("field %q (profile-store) has empty Persist.Field", f.Name)
			}
		case PersistProjectConfig:
			if f.Persist.Section == "" || f.Persist.Key == "" {
				t.Errorf("field %q (project-config) has empty Persist.Section/Key", f.Name)
			}
		default:
			t.Errorf("field %q has unknown Persist.Kind %q", f.Name, f.Persist.Kind)
		}
		if f.Type == TypeSelect && len(f.Options) == 0 {
			t.Errorf("select field %q has empty Options", f.Name)
		}
	}
}

// TestSchemaOptionListUniqueness는 각 select 필드의 옵션 값에 중복이 없고
// 모두 비어있지 않음을 검증한다(REQ-WC10-004 단일 원천 무결성).
func TestSchemaOptionListUniqueness(t *testing.T) {
	for _, f := range AllFields() {
		if len(f.Options) == 0 {
			continue
		}
		seen := make(map[string]bool, len(f.Options))
		for _, o := range f.Options {
			if o.Value == "" {
				t.Errorf("field %q has an empty option value (empty handled by EmptyLabel, not Options)", f.Name)
			}
			if seen[o.Value] {
				t.Errorf("field %q has duplicate option value %q", f.Name, o.Value)
			}
			seen[o.Value] = true
			if o.I18nKey == "" {
				t.Errorf("field %q option %q has empty I18nKey", f.Name, o.Value)
			}
		}
	}
}

// TestSchemaValidatePredicates는 검증 술어를 가진 필드가 정규 옵션 값을 통과시키고
// 명백히 잘못된 값을 거부함을 검증한다(검증 술어가 옵션 목록과 정합).
func TestSchemaValidatePredicates(t *testing.T) {
	for _, f := range AllFields() {
		if f.Validate == nil || len(f.Options) == 0 {
			continue
		}
		for _, o := range f.Options {
			if !f.Validate(o.Value) {
				t.Errorf("field %q: Validate rejected canonical option %q", f.Name, o.Value)
			}
		}
		if f.Validate("__definitely_not_a_valid_value__") {
			t.Errorf("field %q: Validate accepted a bogus value", f.Name)
		}
	}
}

// TestSchemaNoPresetField는 retired statusline preset 이 스키마에 존재하지 않음을
// 검증한다(REQ-WC10-010, B3). 어떤 필드명도 "preset" 을 포함하면 안 된다.
func TestSchemaNoPresetField(t *testing.T) {
	for _, f := range AllFields() {
		if strings.Contains(strings.ToLower(f.Name), "preset") {
			t.Errorf("schema contains a preset-related field %q (retired by STATUSLINE-PRESET-RETIRE-001)", f.Name)
		}
	}
}

// TestSchemaImportsNeitherCLINorWeb는 internal/settings 가 internal/cli 또는
// internal/web 를 import 하지 않음을 go list -deps 로 검증한다(AC-WC10-003,
// 역방향 import 금지). 의존성 방향 검사.
func TestSchemaImportsNeitherCLINorWeb(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "github.com/modu-ai/moai-adk/internal/settings/...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}
	deps := string(out)
	for _, forbidden := range []string{
		"github.com/modu-ai/moai-adk/internal/cli",
		"github.com/modu-ai/moai-adk/internal/web",
	} {
		for _, line := range strings.Split(deps, "\n") {
			if strings.TrimSpace(line) == forbidden {
				t.Errorf("internal/settings transitively imports %q (forbidden reverse edge)", forbidden)
			}
		}
	}
}
