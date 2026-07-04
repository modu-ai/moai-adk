package yamlpatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// copyFixture는 testdata fixture를 t.TempDir 아래로 복사하고 사본 경로와 원본
// 내용을 반환한다. 실제 프로젝트 파일은 절대 건드리지 않는다.
func copyFixture(t *testing.T, name string) (string, string) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	dst := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(dst, src, 0o644); err != nil {
		t.Fatalf("write fixture copy: %v", err)
	}
	return dst, string(src)
}

// writeTempYAML은 인라인 yaml 내용을 임시 파일로 기록한다.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "section.yaml")
	if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return dst
}

// changedLines는 before/after를 라인 단위로 비교해 (변경된 라인 쌍, before 전용
// 라인, after 전용 라인)을 반환한다. 라인 수가 같으면 인덱스 비교, 다르면 선두
// 공통 프리픽스 이후를 각각 전용 라인으로 취급한다.
func changedLines(before, after string) (changed [][2]string, beforeOnly, afterOnly []string) {
	b := splitLines(before)
	a := splitLines(after)
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

// splitLines는 라인 분할 후 trailing newline이 만드는 마지막 빈 요소를 제거한다.
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// commentLines는 '#'로 시작하는(trim 후) 라인 전체를 순서대로 반환한다.
func commentLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

// TestYAMLPatchScalarReplace_WorkflowFixture는 실제 workflow.yaml fixture에서
// role_profiles.implementer.model 스칼라 하나만 교체될 때 (i) diff가 편집 라인
// 1개에 국한되고 (ii) 주석 전량과 (iii) 미모델링 키(team.patterns)·키 순서가
// 보존됨을 검증한다 (AC-WC11-003, GWT-2).
func TestYAMLPatchScalarReplace_WorkflowFixture(t *testing.T) {
	t.Parallel()
	path, before := copyFixture(t, "workflow.yaml")

	err := PatchFile(path, []KeyEdit{
		{Path: []string{"workflow", "team", "role_profiles", "implementer", "model"}, Value: "opus"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	after := string(raw)

	changed, beforeOnly, afterOnly := changedLines(before, after)
	if len(beforeOnly) > 0 || len(afterOnly) > 0 {
		t.Fatalf("line count drifted (whitespace normalization?):\nbefore-only: %q\nafter-only: %q", beforeOnly, afterOnly)
	}
	if len(changed) != 1 {
		t.Fatalf("want exactly 1 changed line, got %d: %v", len(changed), changed)
	}
	if !strings.Contains(changed[0][1], "model: opus") {
		t.Errorf("changed line = %q, want it to carry %q", changed[0][1], "model: opus")
	}

	// 주석 전량 보존.
	if got, want := commentLines(after), commentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("comments not preserved:\nbefore: %v\nafter:  %v", want, got)
	}

	// 미모델링 키 team.patterns 잔존 (EXCL-WSE-004 보존).
	for _, key := range []string{"patterns:", "design_implementation:", "full_stack:", "investigation:"} {
		if !strings.Contains(after, key) {
			t.Errorf("unmodeled key %q lost after patch", key)
		}
	}
	// role-profile effort 키(Go-invisible, REQ-WEM-006)도 잔존.
	if got, want := strings.Count(after, "effort:"), strings.Count(before, "effort:"); got != want {
		t.Errorf("effort key count drifted: before %d, after %d", want, got)
	}
}

// TestYAMLPatchUpsert_MissingPathCreated는 부재 경로(workflow_agents 블록 최초
// 기록, REQ-WC11-073 대비)의 upsert가 기존 내용을 전부 보존한 채 신규 블록만
// 추가함을 검증한다 (AC-WC11-003 upsert 케이스, GWT-7).
func TestYAMLPatchUpsert_MissingPathCreated(t *testing.T) {
	t.Parallel()
	path, before := copyFixture(t, "workflow.yaml")

	if strings.Contains(before, "workflow_agents") {
		t.Fatal("fixture precondition violated: workflow_agents already present")
	}

	err := PatchFile(path, []KeyEdit{
		{Path: []string{"workflow", "workflow_agents", "implementer", "model"}, Value: "haiku"},
		{Path: []string{"workflow", "workflow_agents", "implementer", "effort"}, Value: "low"},
	})
	if err != nil {
		t.Fatalf("PatchFile upsert: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	after := string(raw)

	changed, beforeOnly, afterOnly := changedLines(before, after)
	if len(changed) != 0 || len(beforeOnly) != 0 {
		t.Fatalf("upsert must be additive-only: changed=%v before-only=%q", changed, beforeOnly)
	}
	joined := strings.Join(afterOnly, "\n")
	for _, want := range []string{"workflow_agents:", "implementer:", "model: haiku", "effort: low"} {
		if !strings.Contains(joined, want) {
			t.Errorf("upsert block missing %q in added lines:\n%s", want, joined)
		}
	}

	// 주석 전량 보존.
	if got, want := commentLines(after), commentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("comments not preserved across upsert:\nbefore: %v\nafter:  %v", want, got)
	}
}

// TestYAMLPatchPreservesQuotedStyle은 double-quoted 빈 문자열 값(db.yaml `orm: ""`
// 형상)을 교체해도 인용 스타일이 유지됨을 검증한다 (EC-4 유사 — 정규화 금지).
func TestYAMLPatchPreservesQuotedStyle(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "db:\n  orm: \"\"\n  multi_tenant: \"none\"\n")

	if err := PatchFile(path, []KeyEdit{{Path: []string{"db", "orm"}, Value: "prisma"}}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, _ := os.ReadFile(path)
	after := string(raw)
	if !strings.Contains(after, "orm: \"prisma\"") {
		t.Errorf("quoted style not preserved, got:\n%s", after)
	}
	if !strings.Contains(after, "multi_tenant: \"none\"") {
		t.Errorf("sibling quoted scalar drifted, got:\n%s", after)
	}
}

// TestYAMLPatchPreservesTypedScalars는 plain int/bool 스칼라 교체가 따옴표 없는
// 원래 타입 형상으로 기록됨을 검증한다.
func TestYAMLPatchPreservesTypedScalars(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "ralph:\n    enabled: true\n    loop:\n        max_iterations: 10\n")

	err := PatchFile(path, []KeyEdit{
		{Path: []string{"ralph", "loop", "max_iterations"}, Value: "20"},
		{Path: []string{"ralph", "enabled"}, Value: "false"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, _ := os.ReadFile(path)
	after := string(raw)
	if !strings.Contains(after, "max_iterations: 20") {
		t.Errorf("int scalar drifted, got:\n%s", after)
	}
	if strings.Contains(after, "\"20\"") || strings.Contains(after, "'20'") {
		t.Errorf("int scalar was quoted, got:\n%s", after)
	}
	if !strings.Contains(after, "enabled: false") {
		t.Errorf("bool scalar drifted, got:\n%s", after)
	}
}

// TestYAMLPatchEmptyEditsIsNoop은 빈 edits가 파일을 재기록하지 않음을 검증한다.
func TestYAMLPatchEmptyEditsIsNoop(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "feedback:\n    repository: modu-ai/moai-adk\n")
	before, _ := os.ReadFile(path)

	if err := PatchFile(path, nil); err != nil {
		t.Fatalf("PatchFile(nil): %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Error("empty edits mutated the file")
	}
}

// TestYAMLPatchErrors는 오류 경로를 검증한다: 파일 부재, 파싱 불가, 빈 경로,
// 스칼라가 아닌 대상(매핑) 교체 시도.
func TestYAMLPatchErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		err := PatchFile(filepath.Join(t.TempDir(), "nope.yaml"), []KeyEdit{{Path: []string{"a"}, Value: "1"}})
		if err == nil {
			t.Fatal("want error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "a: [unclosed\n")
		err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: "1"}})
		if err == nil {
			t.Fatal("want error for invalid yaml")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "a: 1\n")
		err := PatchFile(path, []KeyEdit{{Path: nil, Value: "1"}})
		if err == nil {
			t.Fatal("want error for empty edit path")
		}
	})

	t.Run("non-scalar target", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "a:\n    b: 1\n")
		before, _ := os.ReadFile(path)
		err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: "x"}})
		if err == nil {
			t.Fatal("want error when target is a mapping, not a scalar")
		}
		after, _ := os.ReadFile(path)
		if string(before) != string(after) {
			t.Error("failed patch must not mutate the file")
		}
	})

	t.Run("empty document", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "")
		err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: "1"}})
		if err == nil {
			t.Fatal("want error for empty yaml document")
		}
	})

	t.Run("non-mapping root", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "- item1\n- item2\n")
		err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: "1"}})
		if err == nil {
			t.Fatal("want error for sequence root document")
		}
	})

	t.Run("scalar intermediate segment", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "a: 1\n")
		before, _ := os.ReadFile(path)
		err := PatchFile(path, []KeyEdit{{Path: []string{"a", "b", "c"}, Value: "x"}})
		if err == nil {
			t.Fatal("want error when an intermediate segment is a scalar, not a mapping")
		}
		after, _ := os.ReadFile(path)
		if string(before) != string(after) {
			t.Error("failed patch must not mutate the file")
		}
	})

	t.Run("sequence target", func(t *testing.T) {
		t.Parallel()
		path := writeTempYAML(t, "a:\n    - x\n    - y\n")
		err := PatchFile(path, []KeyEdit{{Path: []string{"a"}, Value: "flat"}})
		if err == nil {
			t.Fatal("want error when target is a sequence, not a scalar")
		}
	})
}

// TestYAMLPatchDefaultIndentFallback은 들여쓰기 검출이 불가능한 flat 문서의
// 중첩 upsert가 섹션 관례 기본값(4-space)으로 기록됨을 검증한다.
func TestYAMLPatchDefaultIndentFallback(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "a: 1\n")

	if err := PatchFile(path, []KeyEdit{{Path: []string{"b", "c"}, Value: "v"}}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, _ := os.ReadFile(path)
	after := string(raw)
	if !strings.Contains(after, "b:\n    c: v") {
		t.Errorf("default 4-space indent not applied on upsert, got:\n%s", after)
	}
	if !strings.Contains(after, "a: 1") {
		t.Errorf("existing key lost, got:\n%s", after)
	}
}

// TestYAMLPatchMultiEditSingleWrite는 복수 edit이 한 번의 원자적 기록으로
// 모두 반영됨을 검증한다 (M2b 폼 저장 시 다중 필드 제출 대비).
func TestYAMLPatchMultiEditSingleWrite(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "s:\n    a: 1\n    b: two\n    c: true\n")

	err := PatchFile(path, []KeyEdit{
		{Path: []string{"s", "a"}, Value: "9"},
		{Path: []string{"s", "c"}, Value: "false"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, _ := os.ReadFile(path)
	after := string(raw)
	for _, want := range []string{"a: 9", "b: two", "c: false"} {
		if !strings.Contains(after, want) {
			t.Errorf("multi-edit missing %q, got:\n%s", want, after)
		}
	}
}

// TestYAMLPatchIndentDetection은 2-space 파일(db.yaml 형상)의 upsert가 원본
// 들여쓰기 폭을 따름을 검증한다.
func TestYAMLPatchIndentDetection(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "db:\n  enabled: false\n  orm: \"\"\n")

	if err := PatchFile(path, []KeyEdit{{Path: []string{"db", "extra", "key"}, Value: "v"}}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, _ := os.ReadFile(path)
	after := string(raw)
	if !strings.Contains(after, "  extra:\n    key: v") {
		t.Errorf("2-space indent not preserved on upsert, got:\n%s", after)
	}
}

// TestYAMLPatchAtomicWriteErrors는 atomicWrite의 오류 분기를 직접 검증한다:
// 대상 파일 부재(stat 실패) + 쓰기 불가 디렉터리(temp 생성 실패).
func TestYAMLPatchAtomicWriteErrors(t *testing.T) {
	t.Parallel()

	t.Run("stat missing target", func(t *testing.T) {
		t.Parallel()
		err := atomicWrite(filepath.Join(t.TempDir(), "gone.yaml"), []byte("a: 1\n"))
		if err == nil {
			t.Fatal("want error when target file does not exist")
		}
	})

	t.Run("read-only directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "section.yaml")
		if err := os.WriteFile(path, []byte("a: 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Skipf("cannot chmod dir read-only on this platform: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		if err := atomicWrite(path, []byte("a: 2\n")); err == nil {
			t.Fatal("want error when temp file cannot be created in a read-only dir")
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil || string(raw) != "a: 1\n" {
			t.Errorf("failed atomic write mutated the target: %q (err %v)", raw, readErr)
		}
	})
}

// TestYAMLPatchEncodeError는 인코딩 불가 노드(unknown kind)가 오류로 전파됨을
// 직접 검증한다 (encode 오류 분기).
func TestYAMLPatchEncodeError(t *testing.T) {
	t.Parallel()
	bad := &yaml.Node{
		Kind: yaml.DocumentNode,
		// Kind 0 + 비어있지 않은 노드 — yaml.v3가 "unknown kind"로 거부한다
		// (Kind 0의 zero 노드는 null로 인코딩되므로 Value를 채워야 오류 분기 진입).
		Content: []*yaml.Node{{Value: "x"}},
	}
	if _, err := encode(bad, 4); err == nil {
		t.Fatal("want error when encoding a node with unknown kind")
	}
}
