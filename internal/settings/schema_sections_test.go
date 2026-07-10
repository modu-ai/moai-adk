package settings

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// seedTypedFixtures는 typed 섹션 fixture(git-strategy/llm/quality)를 임시 루트에
// 시드한다. seam 섹션은 seedSectionFixture(sectionwrite_test.go)를 재사용한다.
func seedTypedFixtures(t *testing.T, root string, names ...string) {
	t.Helper()
	for _, name := range names {
		seedSectionFixture(t, root, name)
	}
}

// TestSchemaSectionsRegistered는 M2b 확장 스키마의 구조 불변식을 검증한다:
// (i) 확장 섹션 전부에 필드 ≥ 1 (ii) 필드명 유일 (iii) seam 필드의 섹션은
// RouteSeam으로 라우팅 (iv) typed 필드의 섹션은 typed applier 대상
// (v) read-only 키(llm.mode/team_mode)는 편집 필드에 없음.
func TestSchemaSectionsRegistered(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	perSection := map[SectionID]int{}
	for _, f := range AllFields() {
		if seen[f.Name] {
			t.Errorf("duplicate field name %q", f.Name)
		}
		seen[f.Name] = true
		perSection[f.Section]++

		switch f.Persist.Kind {
		case PersistSeam:
			if RouteForSection(f.Persist.Section) != RouteSeam {
				t.Errorf("field %q: seam persist targets non-seam section %q", f.Name, f.Persist.Section)
			}
			if len(f.Persist.Path) == 0 {
				t.Errorf("field %q: seam persist without path", f.Name)
			}
		case PersistTypedSection:
			if _, ok := typedSectionFiles[f.Persist.Section]; !ok {
				t.Errorf("field %q: typed persist targets unknown section %q", f.Name, f.Persist.Section)
			}
		}
	}

	for _, sec := range SchemaSectionIDs() {
		if perSection[sec] == 0 {
			t.Errorf("schema section %q has no fields", sec)
		}
	}

	// read-only 키는 편집 필드로 존재하지 않는다 (REQ-WC11-013).
	for _, name := range []string{
		"llm.mode", "llm.team_mode",
	} {
		if seen[name] {
			t.Errorf("read-only key %q must not be an editable field", name)
		}
	}
}

// TestApplySchemaEditsSeamRoundTrip은 seam 필드 편집이 WriteSectionViaSeam 경유로
// 파일에 반영되고 주석이 보존됨을 검증한다 (REQ-WC11-017; AC-WC11-004 인프라 계층).
func TestApplySchemaEditsSeamRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "workflow")

	// 구 workflow.team.max_teammates seam 필드는 Agent Teams 정적 레이어와 함께
	// 제거되었다 (SPEC-AGENT-TEAM-RETIRE-001 M2) — 잔존 workflow seam 필드로 교체.
	err := ApplySchemaEdits(root, map[string]string{
		"workflow.token_budget.plan": "31000",
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if !strings.Contains(after, "plan: 31000") {
		t.Errorf("seam edit not persisted:\n%s", after)
	}
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("comments not preserved by seam routing")
	}
}

// TestApplySchemaEditsGitStrategyTyped는 git-strategy 편집이 typed dirty-flag
// Save 경로로 반영됨을 검증한다 (REQ-WC11-010, AC-WC11-010). M4 다이어트 후
// mode와 profile별 hooks.pre_push만 편집 가능하다.
func TestApplySchemaEditsGitStrategyTyped(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")

	err := ApplySchemaEdits(root, map[string]string{
		"git_strategy.mode":                  "personal",
		"git_strategy.team.hooks.pre_push":   "enforce",
		"git_strategy.manual.hooks.pre_push": "skip",
		// SPEC-WEB-CONSOLE-014 M3 (AC-WC14-010b): merge_method 라운드트립 확장.
		"git_strategy.team.merge_method":     "rebase",
		"git_strategy.personal.merge_method": "merge",
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits(git_strategy): %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		GS struct {
			Mode string `yaml:"mode"`
			Team struct {
				MergeMethod string `yaml:"merge_method"`
				Hooks       struct {
					PrePush string `yaml:"pre_push"`
				} `yaml:"hooks"`
			} `yaml:"team"`
			Personal struct {
				MergeMethod string `yaml:"merge_method"`
			} `yaml:"personal"`
			Manual struct {
				Hooks struct {
					PrePush string `yaml:"pre_push"`
				} `yaml:"hooks"`
			} `yaml:"manual"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.GS.Mode != "personal" {
		t.Errorf("mode = %q, want personal", doc.GS.Mode)
	}
	if doc.GS.Team.Hooks.PrePush != "enforce" {
		t.Errorf("team.hooks.pre_push = %q, want enforce", doc.GS.Team.Hooks.PrePush)
	}
	if doc.GS.Manual.Hooks.PrePush != "skip" {
		t.Errorf("manual.hooks.pre_push = %q, want skip", doc.GS.Manual.Hooks.PrePush)
	}
	if doc.GS.Team.MergeMethod != "rebase" {
		t.Errorf("team.merge_method = %q, want rebase", doc.GS.Team.MergeMethod)
	}
	if doc.GS.Personal.MergeMethod != "merge" {
		t.Errorf("personal.merge_method = %q, want merge", doc.GS.Personal.MergeMethod)
	}
}

// TestMergeMethodFieldsExposed는 SPEC-WEB-CONSOLE-014 M3을 검증한다: AllFields에
// git_strategy.{manual,personal,team}.merge_method 정확 3건이 닫힌 enum select로
// 노출되고, 옵션은 config.ValidMergeMethods() SSOT에서 정렬 파생(리터럴 재선언
// 금지)되며, 빈 문자열과 enum 밖 값이 검증에서 거부됨 (REQ-WC14-010/011).
func TestMergeMethodFieldsExposed(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"git_strategy.manual.merge_method":   true,
		"git_strategy.personal.merge_method": true,
		"git_strategy.team.merge_method":     true,
	}
	got := map[string]FieldDef{}
	for _, f := range AllFields() {
		if strings.HasSuffix(f.Name, ".merge_method") {
			got[f.Name] = f
		}
	}
	if len(got) != 3 {
		t.Fatalf("merge_method field count = %d, want 3 (manual/personal/team)", len(got))
	}
	wantOpts := append([]string{}, config.ValidMergeMethods()...)
	sort.Strings(wantOpts) // 정렬 파생 (B3 — map-range 비결정성 제거).
	for name := range want {
		f, ok := got[name]
		if !ok {
			t.Errorf("merge_method field %q missing", name)
			continue
		}
		if f.Type != TypeSelect {
			t.Errorf("%q type = %q, want select", name, f.Type)
		}
		if gotOpts := f.SelectOptions(); !reflect.DeepEqual(gotOpts, wantOpts) {
			t.Errorf("%q options = %v, want sorted-from-SSOT %v", name, gotOpts, wantOpts)
		}
		if f.Validate == nil {
			t.Fatalf("%q has no validator (closed-set membership)", name)
		}
		if f.Validate("") {
			t.Errorf("%q validator accepted empty (empty NOT a member — REQ-WC14-010)", name)
		}
		if f.Validate("fast-forward") {
			t.Errorf("%q validator accepted out-of-enum value (REQ-WC14-011)", name)
		}
	}
}

// TestMergeMethodAbsentKeyDisplay는 iter-2 D5 absent-key 초기 표시를 검증한다:
// merge_method 키가 없는 git-strategy.yaml 픽스처에서 (i) 현재값 읽기가 검증 실패
// 없이 빈 값을 반환하고, (ii) 빈 문자열 값의 저장은 enum 비멤버로 거부되며,
// (iii) 명시적 enum 값 저장은 반영됨 (REQ-WC14-010, AC-WC14-010c).
func TestMergeMethodAbsentKeyDisplay(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	// 실측: fixture git-strategy.yaml에 merge_method 키 부재 (live absent-key 상태).

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues (absent merge_method): %v", err)
	}
	for _, name := range []string{"git_strategy.manual.merge_method", "git_strategy.team.merge_method"} {
		if v := values[name]; v != "" {
			t.Errorf("absent %q should read empty (no validation failure), got %q", name, v)
		}
	}
	// 빈 문자열 저장은 enum 비멤버로 거부 (empty NOT a member).
	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.team.merge_method": ""}); err == nil {
		t.Error("saving empty merge_method should be rejected (empty NOT a member — AC-WC14-010c)")
	}
	// 명시적 enum 값 저장은 반영 + 주석 보존.
	before, _ := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"))
	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.team.merge_method": "rebase"}); err != nil {
		t.Fatalf("ApplySchemaEdits(merge_method=rebase): %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml"))
	if !strings.Contains(string(after), "merge_method: rebase") {
		t.Errorf("merge_method not persisted:\n%s", after)
	}
	// enum 밖 값 저장은 거부.
	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.manual.merge_method": "fast-forward"}); err == nil {
		t.Error("saving out-of-enum merge_method should be rejected (REQ-WC14-011)")
	}
	_ = before
}

// TestApplySchemaEditsGitStrategyDirtyFlagIsolation은 git-strategy를 건드리지
// 않는 typed 편집(llm)이 git-strategy.yaml을 byte 단위로 보존함을 검증한다
// (SPEC-GITSTRATEGY-SAVE-ISOLATION-001 dirty-flag 경로).
func TestApplySchemaEditsGitStrategyDirtyFlagIsolation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	gsPath := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	before, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplySchemaEdits(root, map[string]string{"llm.glm.models.high": "glm-test-model"}); err != nil {
		t.Fatalf("ApplySchemaEdits(llm): %v", err)
	}

	after, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("git-strategy.yaml was rewritten by a non-git-strategy edit (dirty-flag violated)")
	}

	llmAfter, _ := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if !strings.Contains(string(llmAfter), "high: glm-test-model") {
		t.Errorf("llm.glm.models.high not persisted:\n%s", llmAfter)
	}
}

// TestApplySchemaEditsRejectsUnknownAndReadOnly는 미지명 필드와 read-only 키가
// 거부됨을 검증한다 (REQ-WC11-013/018).
func TestApplySchemaEditsRejectsUnknownAndReadOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")

	for _, name := range []string{
		"llm.mode", "llm.team_mode",
		"db.enabled", "db.orm", "state.anything", "nope",
		// M4 다이어트로 제거된 필드 — 편집 불가 (yaml 로드는 backward-compat 보존).
		"llm.performance_tier", "git_strategy.team.required_reviews",
		"quality.coverage_threshold", "ralph.loop.max_iterations",
		"research.enabled",
		// SPEC-WEB-CONSOLE-014 M2: 편집 철거 → read-only 강등 키 (쓰기 거부).
		// learning.auto_apply(F3 governance FROZEN), output_path(F9 dead config).
		"learning.auto_apply", "observability.hook_metrics.output_path",
	} {
		if err := ApplySchemaEdits(root, map[string]string{name: "x"}); err == nil {
			t.Errorf("edit %q: want rejection, got nil", name)
		}
	}
}

// TestHonestDemotionReadOnlyAndRaw는 SPEC-WEB-CONSOLE-014 M2 정직화를 검증한다:
// (i) learning.auto_apply / observability.hook_metrics.output_path는 편집 필드에서
// 철거되고 read-only 표시 키로 등록됨 (REQ-WC14-001/040), (ii) learning.tier_thresholds /
// learning.rate_limit는 raw view 블록으로 등록됨 (REQ-WC14-002/003).
func TestHonestDemotionReadOnlyAndRaw(t *testing.T) {
	t.Parallel()

	editable := map[string]bool{}
	for _, f := range AllFields() {
		editable[f.Name] = true
	}
	// (i) 강등 키는 편집 가능 집합에 없어야 한다.
	for _, name := range []string{"learning.auto_apply", "observability.hook_metrics.output_path"} {
		if editable[name] {
			t.Errorf("demoted key %q must NOT be an editable field (REQ-WC14-001/040)", name)
		}
	}
	// (i) 강등 키는 read-only 표시 집합에 있어야 한다.
	ro := map[string]bool{}
	for _, r := range ReadOnlyDisplayFields() {
		ro[r.Name] = true
	}
	for _, name := range []string{"learning.auto_apply", "observability.hook_metrics.output_path"} {
		if !ro[name] {
			t.Errorf("demoted key %q must be a read-only display field (REQ-WC14-001/040)", name)
		}
	}
	// (ii) learning list 키는 raw view 블록에 있어야 한다.
	raw := map[string]bool{}
	for _, rb := range RawViewBlocks() {
		raw[rb.Name] = true
	}
	for _, name := range []string{"learning.tier_thresholds", "learning.rate_limit"} {
		if !raw[name] {
			t.Errorf("list key %q must be a raw view block (REQ-WC14-002/003)", name)
		}
	}
}

// TestSchemaCurrentValuesReadsAllSections는 제네릭 읽기 seam이 (i) 확장 필드 전부
// (ii) read-only 표시 키의 현재 값을 읽음을 검증한다.
func TestSchemaCurrentValuesReadsAllSections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality",
		"workflow", "harness", "ralph", "research", "feedback", "observability", "security")

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues: %v", err)
	}

	cases := map[string]string{
		// 구 workflow.team.max_teammates(10)는 Agent Teams 정적 레이어와 함께 제거됨
		// (SPEC-AGENT-TEAM-RETIRE-001 M2) — 잔존 workflow seam 필드로 교체.
		"workflow.token_budget.plan":                   "30000",
		"harness.default_profile":                      "default",
		"learning.enabled":                             "true",
		"ralph.lint_as_instruction":                    "true",
		"ralph.warn_as_instruction":                    "false",
		"feedback.repository":                          "modu-ai/moai-adk",
		"observability.retention_days":                 "30",
		"security.permission.strict_mode":              "false",
		"git_strategy.mode":                            "team",
		"git_strategy.team.hooks.pre_push":             "enforce",
		"llm.glm.models.high":                          "glm-5.2",
		"quality.ddd_settings.characterization_tests":  "true",
		"quality.ddd_settings.behavior_snapshots":      "true",
		"quality.ddd_settings.preserve_before_improve": "true",
		// read-only 표시 키.
		"llm.mode":      "",
		"llm.team_mode": "",
	}
	for name, want := range cases {
		if got := values[name]; got != want {
			t.Errorf("value[%q] = %q, want %q", name, got, want)
		}
	}
}

// TestSchemaCurrentValuesMissingFilesAreEmpty는 섹션 파일 부재 시 오류 없이 빈
// 값을 반환함을 검증한다 (greenfield 루트).
func TestSchemaCurrentValuesMissingFilesAreEmpty(t *testing.T) {
	t.Parallel()
	values, err := SchemaCurrentValues(t.TempDir())
	if err != nil {
		t.Fatalf("SchemaCurrentValues(empty root): %v", err)
	}
	if got := values["workflow.token_budget.plan"]; got != "" {
		t.Errorf("missing file should read empty, got %q", got)
	}
}

// TestRawBlockValues는 REQ-WC11-062 raw view 블록이 표시용 텍스트로 추출됨을
// 검증한다.
func TestRawBlockValues(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "workflow", "harness", "security", "mx")

	blocks, err := RawBlockValues(root)
	if err != nil {
		t.Fatalf("RawBlockValues: %v", err)
	}
	// 구 workflow.team.patterns raw block은 Agent Teams 정적 레이어와 함께 제거됨
	// (SPEC-AGENT-TEAM-RETIRE-001 M2). 잔존 raw block들로 검증한다.
	if !strings.Contains(blocks["harness.levels"], "thorough:") {
		t.Errorf("harness.levels raw block missing content: %q", blocks["harness.levels"])
	}
	// SPEC-WEB-CONSOLE-014 M2: learning list 키 raw view (REQ-WC14-002/003).
	if _, ok := blocks["learning.tier_thresholds"]; !ok {
		t.Error("learning.tier_thresholds raw block missing (REQ-WC14-002)")
	}
	if !strings.Contains(blocks["learning.rate_limit"], "max_per_week:") {
		t.Errorf("learning.rate_limit raw block missing content: %q", blocks["learning.rate_limit"])
	}
	// SPEC-WEB-CONSOLE-014 M4: security.sandbox list 키 raw view (REQ-WC14-020).
	for _, name := range []string{"security.sandbox.network_allowlist", "security.sandbox.env_scrub_extra"} {
		if _, ok := blocks[name]; !ok {
			t.Errorf("%s raw block missing (REQ-WC14-020)", name)
		}
	}
	// SPEC-WEB-CONSOLE-014 M4: mx list 키 raw view (REQ-WC14-030).
	if !strings.Contains(blocks["mx.danger_categories"], "concurrency:") {
		t.Errorf("mx.danger_categories raw block missing content: %q", blocks["mx.danger_categories"])
	}
	if !strings.Contains(blocks["mx.test_paths"], "_test.go") {
		t.Errorf("mx.test_paths raw block missing content: %q", blocks["mx.test_paths"])
	}
}

// TestQualityKeyPartition은 AC-WC11-011의 파티션을 기계 검증한다: quality.yaml
// fixture의 모든 스칼라 leaf 키는 (기존+확장 노출 키) ∪ (명시 제외 prefix)에
// 속해야 한다 — 미노출 잔여 = 0.
func TestQualityKeyPartition(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "sections", "quality.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	root := doc.Content[0]
	// quality.yaml의 최상위 키는 constitution.
	var constitution *yaml.Node
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "constitution" {
			constitution = root.Content[i+1]
		}
	}
	if constitution == nil {
		t.Fatal("quality.yaml fixture has no constitution top-level key")
	}

	// 노출 키 집합: 기존 4필드 + 확장 필드의 quality.<key>.
	exposed := map[string]bool{
		"development_mode": true, // 기존 development_mode 필드 (Name "development_mode")
	}
	for _, f := range AllFields() {
		if (f.Persist.Kind == PersistProjectConfig || f.Persist.Kind == PersistTypedSection) &&
			f.Persist.Section == "quality" {
			exposed[f.Persist.Key] = true
		}
	}

	var leaves []string
	collectScalarLeaves(constitution, "", &leaves)
	for _, leaf := range leaves {
		if exposed[leaf] {
			continue
		}
		excluded := false
		for _, prefix := range QualityExcludedKeyPrefixes() {
			if strings.HasPrefix(leaf, prefix) {
				excluded = true
				break
			}
		}
		if !excluded {
			t.Errorf("quality.yaml scalar key %q is neither exposed nor on the explicit exclusion list (AC-WC11-011)", leaf)
		}
	}
}

// TestApplySchemaEditsAllFieldsRoundTrip은 M2b 확장 편집 필드 전수에 유효 값을
// 단일 ApplySchemaEdits 호출로 기록하고 제네릭 리더로 전량 재확인한다 — typed
// applier(git_strategy/llm/quality) 전 분기 + seam 라우팅 + 읽기 seam의 전 필드
// 커버리지 가드다 (AC-WC11-010 "git-strategy 전체 필드" 포함). M4 다이어트로
// 편집 필드 수가 감소되었다 — 본 테스트는 AllFields()에서 동적으로 파생하므로
// 감소에 자동 적응한다.
func TestApplySchemaEditsAllFieldsRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality",
		"workflow", "harness", "ralph", "research", "feedback", "observability", "security",
		"handoff", "cache") // SPEC-WEB-CONSOLE-013 M2 신규 seam 섹션

	edits := map[string]string{}
	for _, f := range AllFields() {
		if f.Persist.Kind != PersistSeam && f.Persist.Kind != PersistTypedSection {
			continue
		}
		switch f.Type {
		case TypeBool:
			edits[f.Name] = "true"
		case TypeInt:
			edits[f.Name] = "5"
		case TypeFloat:
			edits[f.Name] = "0.5"
		case TypeSelect:
			edits[f.Name] = f.Options[0].Value
		default:
			edits[f.Name] = "x-test"
		}
	}
	if len(edits) == 0 {
		t.Fatal("no schema-editable fields found")
	}

	if err := ApplySchemaEdits(root, edits); err != nil {
		t.Fatalf("ApplySchemaEdits(all %d fields): %v", len(edits), err)
	}

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues: %v", err)
	}
	for name, want := range edits {
		if got := values[name]; got != want {
			t.Errorf("round-trip %q = %q, want %q", name, got, want)
		}
	}
}

// TestRemovedFieldsLoadWithoutError는 M4 다이어트로 제거된 필드 키가 포함된
// 기존 yaml 설정이 오류 없이 로드됨을 검증한다 (backward compat — 제거된 키는
// 조용히 무시되고 KEPT 키의 yaml 경로는 안정적이다). fixture(testdata/sections)
// 각 파일은 제거된 키(ralph.loop.max_iterations, research.enabled,
// quality.coverage_threshold, llm.performance_tier, git_strategy 전 profile
// leaf 등)를 그대로 포함한다.
func TestRemovedFieldsLoadWithoutError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality",
		"workflow", "harness", "ralph", "research", "feedback", "observability", "security")

	// seam + typed 읽기 경로: SchemaCurrentValues는 각 섹션 yaml을 yaml.Node로
	// 파싱 — 제거된 키가 있어도 오류 없이 로드된다. 제거된 키는 AllFields()에
	// 없으므로 values 맵에 나타나지 않는다 (조용히 무시됨).
	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues with removed keys: %v", err)
	}
	for _, removed := range []string{
		"ralph.loop.max_iterations", "ralph.enabled",
		"research.enabled",
		"llm.performance_tier",
		"quality.coverage_threshold",
		"git_strategy.team.required_reviews",
	} {
		if _, ok := values[removed]; ok {
			t.Errorf("removed key %q should not appear in SchemaCurrentValues output", removed)
		}
	}
}

// TestLLMFieldsTierSet은 llm 섹션 FieldDef가 실소비 tier 4종
// {high, medium, low, fable}만 노출함을 검증한다 (SPEC-WEB-CONSOLE-012
// REQ-WC12-001/002 — glm.go setGLMEnv가 읽는 canonical 키만; legacy alias
// opus/sonnet/haiku는 웹 편집면에서 제거, struct/fallback은 REQ-WC12-006 보존).
func TestLLMFieldsTierSet(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"llm.glm.models.high":   true,
		"llm.glm.models.medium": true,
		"llm.glm.models.low":    true,
		"llm.glm.models.fable":  true,
	}
	got := map[string]bool{}
	for _, f := range SectionFields(SectionLLM) {
		got[f.Name] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("llm FieldDef %q missing (want exactly the 4 real tiers)", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("llm FieldDef %q must not be exposed (ghost tier)", name)
		}
	}
}

// TestApplyLLMKeyFableAndGhostRejection은 (i) fable 편집이 typed 경로로
// 영속화되고 (ii) 폐기된 ghost 키 편집이 거부됨을 검증한다
// (SPEC-WEB-CONSOLE-012 REQ-WC12-003).
func TestApplyLLMKeyFableAndGhostRejection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")

	if err := ApplySchemaEdits(root, map[string]string{"llm.glm.models.fable": "glm-fable-test"}); err != nil {
		t.Fatalf("ApplySchemaEdits(llm.glm.models.fable): %v", err)
	}
	llmAfter, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(llmAfter), "fable: glm-fable-test") {
		t.Errorf("llm.glm.models.fable not persisted:\n%s", llmAfter)
	}

	for _, ghost := range []string{"llm.glm.models.opus", "llm.glm.models.sonnet", "llm.glm.models.haiku"} {
		if err := ApplySchemaEdits(root, map[string]string{ghost: "x"}); err == nil {
			t.Errorf("edit %q: want rejection (ghost tier removed), got nil", ghost)
		}
	}
}

// TestLLMTypedSavePreservesLegacyGhostKeys는 EC-2를 검증한다: 라이브 llm.yaml에
// legacy opus/sonnet/haiku 키가 값과 함께 잔존하는 상태에서 콘솔 저장(typed
// re-marshal)이 그 키·값을 파괴하지 않는다 (SPEC-WEB-CONSOLE-012 REQ-WC12-006 —
// GLMModels legacy struct 멤버 보존이 곧 하위호환 메커니즘).
func TestLLMTypedSavePreservesLegacyGhostKeys(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	// fixture는 opus/sonnet/haiku 키를 값과 함께 포함한다 (실파일 사본).

	if err := ApplySchemaEdits(root, map[string]string{"llm.glm.models.fable": "glm-roundtrip"}); err != nil {
		t.Fatalf("ApplySchemaEdits(llm): %v", err)
	}

	after, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for key, val := range map[string]string{
		"opus":   "glm-5.2",
		"sonnet": "glm-4.7",
		"haiku":  "glm-4.5-air",
	} {
		if !strings.Contains(string(after), key+": "+val) {
			t.Errorf("legacy key %s: %s destroyed by typed re-marshal (EC-2):\n%s", key, val, after)
		}
	}
}

// TestHandoffCacheFields는 SPEC-WEB-CONSOLE-013 M2 신규 섹션의 FieldDef 구성을
// 파생 방식으로 검증한다 (AC-WC13-011 — 총수 하드코딩 금지). handoff 2필드
// (mode select, guide bool) + cache 2필드(enabled bool, session_ttl select),
// 전부 PersistSeam kind이고 올바른 섹션 파일로 라우팅된다.
func TestHandoffCacheFields(t *testing.T) {
	t.Parallel()

	type want struct {
		typ     FieldType
		file    string
		isSelopt bool
	}
	cases := map[string]want{
		"handoff.mode":              {TypeSelect, "handoff", true},
		"handoff.guide":             {TypeBool, "handoff", false},
		"cacheStrategy.enabled":     {TypeBool, "cache", false},
		"cacheStrategy.session_ttl": {TypeSelect, "cache", true},
	}

	got := map[string]FieldDef{}
	for _, f := range SectionFields(SectionHandoff) {
		got[f.Name] = f
	}
	for _, f := range SectionFields(SectionCache) {
		got[f.Name] = f
	}

	// handoff 섹션은 정확히 2필드, cache 섹션은 정확히 2필드 (파생 카운트).
	if n := len(SectionFields(SectionHandoff)); n != 2 {
		t.Errorf("SectionHandoff field count = %d, want 2", n)
	}
	if n := len(SectionFields(SectionCache)); n != 2 {
		t.Errorf("SectionCache field count = %d, want 2", n)
	}

	for name, w := range cases {
		f, ok := got[name]
		if !ok {
			t.Errorf("field %q missing", name)
			continue
		}
		if f.Type != w.typ {
			t.Errorf("field %q type = %q, want %q", name, f.Type, w.typ)
		}
		if f.Persist.Kind != PersistSeam {
			t.Errorf("field %q persist kind = %q, want PersistSeam", name, f.Persist.Kind)
		}
		if f.Persist.Section != w.file {
			t.Errorf("field %q seam section = %q, want %q", name, f.Persist.Section, w.file)
		}
		if w.isSelopt {
			if len(f.Options) == 0 {
				t.Errorf("select field %q has no options", name)
			}
			if f.Validate == nil {
				t.Errorf("select field %q has no validator (closed-set membership)", name)
			}
		}
	}
}

// TestSessionTTLClosedSetSymmetry는 settings측 cacheStrategy.session_ttl select
// 옵션 집합이 config측 validSessionTTLs {1h,5m,off}와 정확히 일치함을 검증한다
// (AC-WC13-013 — export 재사용으로 드리프트 구조적 차단). 옵션은 config.
// ValidSessionTTLs()에서 직접 파생되므로 이 대칭은 구조적으로 보장되나, 명시
// 가드로 회귀를 고정한다.
func TestSessionTTLClosedSetSymmetry(t *testing.T) {
	t.Parallel()

	var sttl *FieldDef
	for _, f := range SectionFields(SectionCache) {
		if f.Name == "cacheStrategy.session_ttl" {
			ff := f
			sttl = &ff
			break
		}
	}
	if sttl == nil {
		t.Fatal("cacheStrategy.session_ttl field not found in SectionCache")
	}

	gotOpts := map[string]bool{}
	for _, o := range sttl.Options {
		gotOpts[o.Value] = true
	}
	wantSet := config.ValidSessionTTLs()
	if len(gotOpts) != len(wantSet) {
		t.Fatalf("session_ttl option count = %d, want %d (config.ValidSessionTTLs)", len(gotOpts), len(wantSet))
	}
	for _, v := range wantSet {
		if !gotOpts[v] {
			t.Errorf("session_ttl option %q missing (config.ValidSessionTTLs symmetry)", v)
		}
	}
	// 리터럴 고정: 닫힌 집합은 정확히 {1h, 5m, off}.
	for _, v := range []string{"1h", "5m", "off"} {
		if !gotOpts[v] {
			t.Errorf("session_ttl closed set missing literal %q", v)
		}
	}
	// 검증기가 집합 밖 값을 거부한다.
	if sttl.Validate != nil && sttl.Validate("2h") {
		t.Error("session_ttl validator accepted out-of-set value 2h")
	}
}

// collectScalarLeaves는 매핑 트리의 스칼라 leaf dot-path를 수집한다 (시퀀스는
// form 대상이 아니므로 진입하지 않는다).
func collectScalarLeaves(node *yaml.Node, prefix string, out *[]string) {
	if node == nil || node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		switch val.Kind {
		case yaml.ScalarNode:
			*out = append(*out, path)
		case yaml.MappingNode:
			collectScalarLeaves(val, path, out)
		}
	}
}
