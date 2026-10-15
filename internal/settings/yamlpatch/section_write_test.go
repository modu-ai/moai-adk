package yamlpatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// readTestYAMLValue는 path의 YAML을 파싱해 llm.team_mode 값을 돌려준다. 값이 null이면
// nil이므로, 빈 문자열("")과 null을 구별하는 데 쓴다.
func readTestYAMLValue(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v\n%s", path, err, raw)
	}
	return doc["llm"]["team_mode"]
}

// TestPatchFileQuotedValueReadsBackAsEmptyString는 KeyEdit.Quoted 가 "" 를 큰따옴표
// 스칼라로 쓰므로, 기존 plain 노드와 새로 만든 키 모두 null이 아닌 "" 로 읽힘을 검증한다.
func TestPatchFileQuotedValueReadsBackAsEmptyString(t *testing.T) {
	t.Parallel()
	existing := writeTempYAML(t, "llm:\n  team_mode: glm # note\n")
	if err := PatchFile(existing, []KeyEdit{{Path: []string{"llm", "team_mode"}, Quoted: true}}); err != nil {
		t.Fatalf("PatchFile existing plain node: %v", err)
	}
	raw, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `team_mode: "" # note`) {
		t.Errorf("existing plain node was not rewritten double-quoted with its comment kept:\n%s", raw)
	}
	if got := readTestYAMLValue(t, existing); got != "" {
		t.Errorf("existing node reads back as %#v, want the empty string", got)
	}

	created := writeTempYAML(t, "llm:\n  harness: claude\n")
	if err := PatchFile(created, []KeyEdit{{Path: []string{"llm", "team_mode"}, Quoted: true}}); err != nil {
		t.Fatalf("PatchFile upsert: %v", err)
	}
	if got := readTestYAMLValue(t, created); got != "" {
		t.Errorf("upserted quoted key reads back as %#v, want the empty string", got)
	}
}

// TestPatchFileWithModeSeedsBlankAndKeepsCommentOnlyFiles는 PatchFileWithMode 가
// 빈 파일을 거부하지 않고 섹션으로 채우며, 주석만 있는 파일의 주석을 지키고 섹션을
// 그 뒤에 덧붙임을 검증한다. PatchFile의 빈 문서 거부는 TestYAMLPatchErrors가 고정한다.
func TestPatchFileWithModeSeedsBlankAndKeepsCommentOnlyFiles(t *testing.T) {
	t.Parallel()
	edits := []KeyEdit{{Path: []string{"llm", "team_mode"}, Value: "glm"}}

	blank := writeTempYAML(t, "")
	if err := PatchFileWithMode(blank, edits, 0o600); err != nil {
		t.Fatalf("PatchFileWithMode on a blank file: %v", err)
	}
	if got := readTestYAMLValue(t, blank); got != "glm" {
		t.Errorf("blank file: team_mode = %#v, want glm", got)
	}

	commentOnly := writeTempYAML(t, "# only a comment\n")
	if err := PatchFileWithMode(commentOnly, edits, 0o600); err != nil {
		t.Fatalf("PatchFileWithMode on a comment-only file: %v", err)
	}
	raw, err := os.ReadFile(commentOnly)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "# only a comment\n") {
		t.Errorf("comment-only file lost its comment:\n%s", raw)
	}
	if got := readTestYAMLValue(t, commentOnly); got != "glm" {
		t.Errorf("comment-only file: team_mode = %#v, want glm", got)
	}
}

// TestPatchFileWithModeLeavesEmptyEditsAlone는 편집이 없으면 PatchFileWithMode 도
// 파일을 건드리지 않음을 검증한다 (PatchFile 과 같은 계약).
func TestPatchFileWithModeLeavesEmptyEditsAlone(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "absent.yaml")
	if err := PatchFileWithMode(path, nil, 0o600); err != nil {
		t.Fatalf("PatchFileWithMode with no edits: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("empty edits created a file: stat err = %v", err)
	}
}
