package config

// llm_save_seam_test.go — REQ-RC30-007 on the ConfigManager.Save path. A dirty llm
// section saved over an existing llm.yaml goes through the yamlpatch seam with no
// typed fallback: a quoted empty value and a removed omitempty key are written as
// patches, so the keys and comments the section does not own survive.

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestSaveLLMDirtyQuotedEmptyIsPatched clears team_mode. The value is written as a
// quoted empty string and keeps its line comment, deployment_mode, an unmodelled
// key, and the header comment survive, and the file keeps its mode.
func TestSaveLLMDirtyQuotedEmptyIsPatched(t *testing.T) {
	t.Parallel()
	root, path := writeLLMSaveFixture(t, "# llm header comment\n"+
		"llm:\n"+
		"  team_mode: glm  # team note\n"+
		"  deployment_mode: plugin\n"+
		"  future_flag: keep-me\n"+
		"  glm:\n"+
		"    models:\n"+
		"      high: glm-5.2\n")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	saveLLMMutation(t, root, func(c *LLMConfig) { c.TeamMode = "" })
	got := readFile(t, path)
	for _, keep := range []string{`team_mode: ""  # team note`, "# llm header comment", "deployment_mode: plugin", "future_flag: keep-me"} {
		if !strings.Contains(got, keep) {
			t.Errorf("a dirty save lost %q:\n%s", keep, got)
		}
	}
	if team := reloadLLMSection(t, root).TeamMode; team != "" {
		t.Errorf("team_mode after the save = %q, want empty", team)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && after.Mode().Perm() != before.Mode().Perm() {
		t.Errorf("a dirty save changed the file mode %v -> %v", before.Mode().Perm(), after.Mode().Perm())
	}
}

// TestSaveLLMDirtyRemovedOmitEmptyFieldIsPatched clears claude_bin, an omitempty
// field, so its key leaves the file. deployment_mode, an unmodelled key, and the
// header comment survive.
func TestSaveLLMDirtyRemovedOmitEmptyFieldIsPatched(t *testing.T) {
	t.Parallel()
	root, path := writeLLMSaveFixture(t, "# llm header comment\n"+
		"llm:\n"+
		"  team_mode: claude\n"+
		"  deployment_mode: plugin\n"+
		"  future_flag: keep-me\n"+
		"  claude_bin: /opt/claude\n"+
		"  glm:\n"+
		"    models:\n"+
		"      high: glm-5.2\n")
	saveLLMMutation(t, root, func(c *LLMConfig) { c.ClaudeBin = "" })
	got := readFile(t, path)
	if strings.Contains(got, "claude_bin") {
		t.Errorf("a cleared omitempty field left its key in the file:\n%s", got)
	}
	for _, keep := range []string{"# llm header comment", "deployment_mode: plugin", "future_flag: keep-me"} {
		if !strings.Contains(got, keep) {
			t.Errorf("a dirty save lost %q:\n%s", keep, got)
		}
	}
	if bin := reloadLLMSection(t, root).ClaudeBin; bin != "" {
		t.Errorf("claude_bin after the save = %q, want empty", bin)
	}
}

// TestSaveLLMDirtyOverBlankFileAppendsSection saves a dirty section over a blank
// llm.yaml. There is no content to keep, so the section is appended and the save
// succeeds rather than failing on the empty document.
func TestSaveLLMDirtyOverBlankFileAppendsSection(t *testing.T) {
	t.Parallel()
	root, _ := writeLLMSaveFixture(t, "")
	saveLLMMutation(t, root, func(c *LLMConfig) { c.TeamMode = "glm" })
	if team := reloadLLMSection(t, root).TeamMode; team != "glm" {
		t.Errorf("team_mode after a save over a blank file = %q, want glm", team)
	}
}
