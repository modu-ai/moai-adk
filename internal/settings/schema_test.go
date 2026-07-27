package settings

import (
	"os/exec"
	"strings"
	"testing"
)

// expected34FieldNames는 6개 섹션에 걸친 정규 필드명을 렌더 순서대로 나열한다.
// 이 목록이 스키마의 단일 진실 검사 기준이다(AC-WC10-010). model_policy는 콘솔에서
// 제거되어(G3-5) Launch는 이제 3개 필드다.
var expected34FieldNames = []string{
	// Identity (1)
	"user_name",
	// Language (4)
	"conversation_lang", "git_commit_lang", "code_comment_lang", "doc_lang",
	// Launch (3)
	"model", "effort_level", "permission_mode",
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

// TestSchemaFieldNameSet은 기존 6개 섹션의 정규 34개 필드가 스키마에 그대로
// 보존됨을 검증한다(AC-WC10-010 하위호환). SPEC-WEB-CONSOLE-011 M2b가 10섹션
// 확장 필드를 추가했으므로 총계는 하드코딩하지 않고(B11 파생 카운트 원칙)
// "기존 34-필드는 전부 잔존 + 중복 없음"만 고정한다. 확장 필드의 구조 불변식은
// TestSchemaSectionsRegistered(schema_sections_test.go)가 담당한다.
func TestSchemaFieldNameSet(t *testing.T) {
	got := FieldNames()

	if len(got) < len(expected34FieldNames) {
		t.Fatalf("FieldNames() count = %d, want >= %d (legacy 34-field floor)", len(got), len(expected34FieldNames))
	}

	gotSet := make(map[string]bool, len(got))
	for _, n := range got {
		if gotSet[n] {
			t.Errorf("duplicate field name in schema: %q", n)
		}
		gotSet[n] = true
	}

	for _, n := range expected34FieldNames {
		if !gotSet[n] {
			t.Errorf("schema missing legacy field %q", n)
		}
	}
}

// TestSchemaSixSections는 기존 6개 정규 섹션의 필드 구성이 보존됨을 검증한다
// (AC-WC10-003 하위호환). SPEC-WEB-CONSOLE-011 M2b 확장으로 AllSections는 확장
// 섹션을 추가로 포함하며, 확장 섹션 커버리지는 TestSchemaSectionsRegistered가
// 담당한다. 총계 하드코딩은 B11에 따라 두지 않는다.
func TestSchemaSixSections(t *testing.T) {
	wantCounts := map[SectionID]int{
		SectionIdentity:      1,
		SectionLanguage:      4,
		SectionLaunch:        3,
		SectionStatusline:    17,
		SectionQuality:       4,
		SectionGitConvention: 5,
	}

	sections := AllSections()
	seen := map[SectionID]bool{}
	for _, s := range sections {
		seen[s] = true
	}
	for s, want := range wantCounts {
		if !seen[s] {
			t.Errorf("legacy section %q missing from AllSections()", s)
			continue
		}
		if got := len(SectionFields(s)); got != want {
			t.Errorf("legacy section %q field count = %d, want %d", s, got, want)
		}
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
		case PersistProjectConfig, PersistTypedSection:
			if f.Persist.Section == "" || f.Persist.Key == "" {
				t.Errorf("field %q (%s) has empty Persist.Section/Key", f.Name, f.Persist.Kind)
			}
		case PersistSeam:
			if f.Persist.Section == "" || len(f.Persist.Path) == 0 {
				t.Errorf("field %q (seam) has empty Persist.Section/Path", f.Name)
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
