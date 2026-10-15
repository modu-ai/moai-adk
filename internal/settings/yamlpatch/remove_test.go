package yamlpatch

import (
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestPatchFileRemoveKeyDropsKeyAndItsComments는 KeyEdit.Remove 가 키-값 쌍과 그 키에
// 붙은 head·line 주석을 함께 떼고, 나머지 키와 주석은 그대로 남김을 검증한다.
func TestPatchFileRemoveKeyDropsKeyAndItsComments(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "# header\n"+
		"llm:\n"+
		"  team_mode: cg # kept line comment\n"+
		"  glm:\n"+
		"    models:\n"+
		"      high: glm-5.3\n"+
		"      # legacy alias note\n"+
		"      opus: glm-5.2 # alias line comment\n"+
		"      sonnet: glm-5.2\n")
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "glm", "models", "opus"}, Remove: true},
		{Path: []string{"llm", "glm", "models", "sonnet"}, Remove: true},
	}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, gone := range []string{"opus", "sonnet", "legacy alias note", "alias line comment"} {
		if strings.Contains(got, gone) {
			t.Errorf("removed key or its comment %q survived:\n%s", gone, got)
		}
	}
	for _, keep := range []string{"# header", "team_mode: cg # kept line comment", "high: glm-5.3"} {
		if !strings.Contains(got, keep) {
			t.Errorf("removal lost %q:\n%s", keep, got)
		}
	}
}

// TestPatchFileRemoveAbsentKeyLeavesBytesAlone는 없는 키의 Remove 가 아무것도 바꾸지
// 않음을 검증한다. 빈 편집 목록과 달리 파일은 읽히지만 바이트는 그대로여야 한다.
func TestPatchFileRemoveAbsentKeyLeavesBytesAlone(t *testing.T) {
	t.Parallel()
	const src = "llm:\n  team_mode: cg\n"
	path := writeTempYAML(t, src)
	if err := PatchFile(path, []KeyEdit{{Path: []string{"llm", "glm", "models", "opus"}, Remove: true}}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != src {
		t.Errorf("removing an absent key changed the file:\n%s", raw)
	}
}

// TestPatchFileRemoveDeletesExactlyTheGivenPaths pins the scope of KeyEdit.Remove.
// It deletes the keys at exactly the paths it is given, together with the comments
// attached to them. Sibling keys, unknown keys, a same-named key under another
// parent, a same-named key at the document root, and comments not attached to a
// removed key all survive.
func TestPatchFileRemoveDeletesExactlyTheGivenPaths(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "# top comment\n"+
		"opus: root-keep\n"+
		"llm:\n"+
		"  team_mode: cg # trailing on a kept key\n"+
		"  deployment_mode: plugin\n"+
		"  future_flag: keep-me\n"+
		"  custom:\n"+
		"    opus: nested-keep\n"+
		"  glm:\n"+
		"    models:\n"+
		"      high: glm-5.3\n"+
		"      # about medium\n"+
		"      medium: glm-5.3\n"+
		"      # head of a removed key\n"+
		"      opus: glm-5.2 # line of a removed key\n"+
		"      sonnet: glm-5.2\n"+
		"      # about low\n"+
		"      low: glm-5.3\n"+
		"      haiku: glm-5.2\n")
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "glm", "models", "opus"}, Remove: true},
		{Path: []string{"llm", "glm", "models", "sonnet"}, Remove: true},
		{Path: []string{"llm", "glm", "models", "haiku"}, Remove: true},
	}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, gone := range []string{"opus: glm-5.2", "sonnet: glm-5.2", "haiku: glm-5.2", "# head of a removed key", "# line of a removed key"} {
		if strings.Contains(got, gone) {
			t.Errorf("removed key or its comment %q survived:\n%s", gone, got)
		}
	}
	for _, keep := range []string{"# top comment", "opus: root-keep", "opus: nested-keep", "team_mode: cg # trailing on a kept key", "deployment_mode: plugin", "future_flag: keep-me", "# about medium", "high: glm-5.3", "medium: glm-5.3", "# about low", "low: glm-5.3"} {
		if !strings.Contains(got, keep) {
			t.Errorf("removal lost %q:\n%s", keep, got)
		}
	}
	tree := parseTestYAMLTree(t, raw)
	llm, _ := tree["llm"].(map[string]any)
	custom, _ := llm["custom"].(map[string]any)
	glm, _ := llm["glm"].(map[string]any)
	models, _ := glm["models"].(map[string]any)
	if keys := slices.Sorted(maps.Keys(models)); !slices.Equal(keys, []string{"high", "low", "medium"}) {
		t.Errorf("models keys after removal = %v, want exactly [high low medium]", keys)
	}
	if custom["opus"] != "nested-keep" || tree["opus"] != "root-keep" {
		t.Errorf("a same-named key outside the removal changed: root=%v nested=%v", tree["opus"], custom["opus"])
	}
	if llm["deployment_mode"] != "plugin" || llm["future_flag"] != "keep-me" {
		t.Errorf("unknown keys changed: deployment_mode=%v future_flag=%v", llm["deployment_mode"], llm["future_flag"])
	}
}

// TestPatchFileRemoveAbsentPathIsNoOp states what a Remove does when its path does
// not resolve. A missing key, a missing parent, or a scalar where a mapping is
// expected is a no-op: it returns no error and leaves the file's bytes unchanged.
func TestPatchFileRemoveAbsentPathIsNoOp(t *testing.T) {
	t.Parallel()
	const src = "llm:\n  team_mode: cg\n  glm:\n    models:\n      high: glm-5.3\n"
	for _, c := range []struct {
		name string
		path []string
	}{
		{name: "missing key under a present mapping", path: []string{"llm", "glm", "models", "opus"}},
		{name: "missing parent", path: []string{"llm", "nothere", "deep", "opus"}},
		{name: "scalar where a mapping is expected", path: []string{"llm", "team_mode", "child"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := writeTempYAML(t, src)
			if err := PatchFile(path, []KeyEdit{{Path: c.path, Remove: true}}); err != nil {
				t.Fatalf("PatchFile: %v", err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(raw) != src {
				t.Errorf("an absent path changed the file:\n%s", raw)
			}
		})
	}
}

// TestPatchFileRemoveAbsentPathAmongPresentOnes pins that an absent path in a call
// does not stop the present removals in the same call. The present key goes, the
// absent path is ignored, and no error is returned.
func TestPatchFileRemoveAbsentPathAmongPresentOnes(t *testing.T) {
	t.Parallel()
	path := writeTempYAML(t, "llm:\n  glm:\n    models:\n      high: glm-5.3\n      opus: glm-5.2\n      sonnet: glm-5.2\n")
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"llm", "nothere", "deep", "opus"}, Remove: true},
		{Path: []string{"llm", "glm", "models", "opus"}, Remove: true},
	}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	llm, _ := parseTestYAMLTree(t, raw)["llm"].(map[string]any)
	glm, _ := llm["glm"].(map[string]any)
	models, _ := glm["models"].(map[string]any)
	if keys := slices.Sorted(maps.Keys(models)); !slices.Equal(keys, []string{"high", "sonnet"}) {
		t.Errorf("models keys = %v, want [high sonnet]", keys)
	}
}

// parseTestYAMLTree parses raw YAML into a generic tree for structural assertions.
func parseTestYAMLTree(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var tree map[string]any
	if err := yaml.Unmarshal(raw, &tree); err != nil {
		t.Fatalf("parse: %v\n%s", err, raw)
	}
	return tree
}
