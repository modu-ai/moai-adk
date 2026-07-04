package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
// (v) read-only 키(llm.mode/team_mode, db system 5키)는 편집 필드에 없음.
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

	// read-only 키는 편집 필드로 존재하지 않는다 (REQ-WC11-013/019).
	for _, name := range []string{
		"llm.mode", "llm.team_mode",
		"db.enabled", "db.dir", "db.engine", "db.auto_sync", "db.migration_patterns",
	} {
		if seen[name] {
			t.Errorf("read-only key %q must not be an editable field", name)
		}
	}

	// db 편집 필드는 정확히 인터뷰 3키다 (REQ-WC11-019).
	dbFields := SectionFields(SectionDB)
	if len(dbFields) != 3 {
		t.Errorf("db editable fields = %d, want 3", len(dbFields))
	}
}

// TestApplySchemaEditsSeamRoundTrip은 seam 필드 편집이 WriteSectionViaSeam 경유로
// 파일에 반영되고 주석이 보존됨을 검증한다 (REQ-WC11-017; AC-WC11-004 인프라 계층).
func TestApplySchemaEditsSeamRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "workflow")

	err := ApplySchemaEdits(root, map[string]string{
		"workflow.team.max_teammates": "7",
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	after := readSection(t, root, "workflow")
	if !strings.Contains(after, "max_teammates: 7") {
		t.Errorf("seam edit not persisted:\n%s", after)
	}
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("comments not preserved by seam routing")
	}
}

// TestApplySchemaEditsGitStrategyTyped는 git-strategy 편집이 typed dirty-flag
// Save 경로로 반영됨을 검증한다 (REQ-WC11-010, AC-WC11-010).
func TestApplySchemaEditsGitStrategyTyped(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")

	err := ApplySchemaEdits(root, map[string]string{
		"git_strategy.mode":                          "personal",
		"git_strategy.personal.automation.auto_push": "false",
		"git_strategy.team.required_reviews":         "2",
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
			Mode     string `yaml:"mode"`
			Personal struct {
				Automation struct {
					AutoPush bool `yaml:"auto_push"`
				} `yaml:"automation"`
			} `yaml:"personal"`
			Team struct {
				RequiredReviews int  `yaml:"required_reviews"`
				DraftPR         bool `yaml:"draft_pr"`
			} `yaml:"team"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.GS.Mode != "personal" {
		t.Errorf("mode = %q, want personal", doc.GS.Mode)
	}
	if doc.GS.Personal.Automation.AutoPush {
		t.Error("personal.automation.auto_push still true")
	}
	if doc.GS.Team.RequiredReviews != 2 {
		t.Errorf("team.required_reviews = %d, want 2", doc.GS.Team.RequiredReviews)
	}
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

	if err := ApplySchemaEdits(root, map[string]string{"llm.performance_tier": "high"}); err != nil {
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
	if !strings.Contains(string(llmAfter), "performance_tier: high") {
		t.Errorf("llm.performance_tier not persisted:\n%s", llmAfter)
	}
}

// TestApplySchemaEditsRejectsUnknownAndReadOnly는 미지명 필드와 read-only 키가
// 거부됨을 검증한다 (REQ-WC11-013/018).
func TestApplySchemaEditsRejectsUnknownAndReadOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")

	for _, name := range []string{"llm.mode", "llm.team_mode", "db.enabled", "state.anything", "nope"} {
		if err := ApplySchemaEdits(root, map[string]string{name: "x"}); err == nil {
			t.Errorf("edit %q: want rejection, got nil", name)
		}
	}
}

// TestSchemaCurrentValuesReadsAllSections는 제네릭 읽기 seam이 (i) 확장 필드 전부
// (ii) read-only 표시 키의 현재 값을 읽음을 검증한다.
func TestSchemaCurrentValuesReadsAllSections(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality",
		"workflow", "harness", "ralph", "research", "feedback", "observability", "security", "db")

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues: %v", err)
	}

	cases := map[string]string{
		"workflow.team.max_teammates":                  "10",
		"harness.default_profile":                      "default",
		"learning.enabled":                             "true",
		"ralph.loop.max_iterations":                    "10",
		"research.enabled":                             "false",
		"feedback.repository":                          "modu-ai/moai-adk",
		"observability.retention_days":                 "30",
		"security.permission.strict_mode":              "false",
		"db.multi_tenant":                              "none",
		"git_strategy.mode":                            "team",
		"git_strategy.team.required_reviews":           "0",
		"llm.performance_tier":                         "",
		"quality.coverage_threshold":                   "0",
		"quality.ddd_settings.max_transformation_size": "small",
		// read-only 표시 키.
		"llm.mode":      "",
		"llm.team_mode": "",
		"db.enabled":    "false",
		"db.dir":        ".moai/project/db",
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
	if got := values["workflow.team.max_teammates"]; got != "" {
		t.Errorf("missing file should read empty, got %q", got)
	}
}

// TestRawBlockValues는 REQ-WC11-062 raw view 블록이 표시용 텍스트로 추출됨을
// 검증한다.
func TestRawBlockValues(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "workflow", "harness", "security", "db")

	blocks, err := RawBlockValues(root)
	if err != nil {
		t.Fatalf("RawBlockValues: %v", err)
	}
	if !strings.Contains(blocks["workflow.team.patterns"], "design_implementation:") {
		t.Errorf("workflow.team.patterns raw block missing content: %q", blocks["workflow.team.patterns"])
	}
	if !strings.Contains(blocks["harness.levels"], "thorough:") {
		t.Errorf("harness.levels raw block missing content: %q", blocks["harness.levels"])
	}
	if !strings.Contains(blocks["db.migration_patterns"], "prisma/schema.prisma") {
		t.Errorf("db.migration_patterns raw block missing content: %q", blocks["db.migration_patterns"])
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

// TestApplySchemaEditsAllFieldsRoundTrip은 M2b 확장 편집 필드 전수(163개)에
// 유효 값을 단일 ApplySchemaEdits 호출로 기록하고 제네릭 리더로 전량 재확인한다
// — typed applier(git_strategy/llm/quality) 전 분기 + seam 라우팅 + 읽기 seam의
// 전 필드 커버리지 가드다 (AC-WC11-010 "git-strategy 전체 필드" 포함).
func TestApplySchemaEditsAllFieldsRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality",
		"workflow", "harness", "ralph", "research", "feedback", "observability", "security", "db")

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
