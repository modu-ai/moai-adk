package cli

// The moai update strip step for the retired per-agent model/effort keys
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-014, design §C and D14): its report,
// the retained-key filter, and its three hosts — the template-sync restore
// (a), the version-matched skip (b), and the clean reinstall (c).

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

const stripLLMFixture = "llm:\n  mode: \"\"\n  profile: \"medium\"\n  performance_tier: \"medium\"\n" +
	"  agent_overrides:\n    manager-develop: { model: opus, effort: xhigh }\n  glm:\n    base_url: x\n"

const stripWorkflowFixture = "workflow:\n  agent_model_guard:\n    enabled: true\n" +
	"  workflow_agents:\n    research: { model: opus, effort: high }\n  audit:\n    glm:\n      model: glm-4.6\n"

// stripProject writes a project whose sections carry the given llm.yaml and
// workflow.yaml bodies (skipped when empty) and a system.yaml stamped with
// templateVersion.
func stripProject(t *testing.T, llm, workflow, templateVersion string) string {
	t.Helper()
	withNoShippedRetiredKeys(t)
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"system.yaml": fmt.Sprintf("moai:\n  template_version: %s\n", templateVersion),
	}
	if llm != "" {
		files["llm.yaml"] = llm
	}
	if workflow != "" {
		files["workflow.yaml"] = workflow
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(sections, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// withNoShippedRetiredKeys puts the strip in the state after the template has
// dropped every retired key, the state design §C describes.
func withNoShippedRetiredKeys(t *testing.T) {
	t.Helper()
	prev := shippedRetiredModelKeysFn
	shippedRetiredModelKeysFn = func() (map[string]bool, error) { return map[string]bool{}, nil }
	t.Cleanup(func() { shippedRetiredModelKeysFn = prev })
}

func stripTestCmd(force bool) (*cobra.Command, *bytes.Buffer) {
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "update-strip-test"}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", force, "")
	return cmd, &buf
}

func sectionBytes(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func backupDirs(t *testing.T, root string) []string {
	t.Helper()
	ents, err := os.ReadDir(filepath.Join(root, ".moai-backups"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, filepath.Join(root, ".moai-backups", e.Name()))
		}
	}
	return out
}

func TestWithoutStrippedKeys_DropsStrippedKeysAndTheirChildren(t *testing.T) {
	refs := []backup.RetainedKeyRef{
		{Section: "llm.yaml", Key: "llm.agent_overrides"},
		{Section: "llm.yaml", Key: "llm.agent_overrides.manager-develop"},
		{Section: "llm.yaml", Key: "llm.profiles.high.manager-spec"},
		{Section: "llm.yaml", Key: "llm.profile", KeptOverDefault: true},
		{Section: "llm.yaml", Key: "llm.profile_extra"},
		{Section: "llm.yaml", Key: "llm.glm.base_url"},
		{Section: "workflow.yaml", Key: "workflow.model_routing"},
		{Section: "design.yaml", Key: "llm.profile"},
	}
	removed := []template.RetiredModelKey{
		{Section: "llm.yaml", Key: "llm.agent_overrides"},
		{Section: "llm.yaml", Key: "llm.profiles"},
		{Section: "llm.yaml", Key: "llm.profile"},
		{Section: "workflow.yaml", Key: "workflow.model_routing"},
	}
	got := withoutStrippedKeys(refs, removed)
	want := []backup.RetainedKeyRef{
		{Section: "llm.yaml", Key: "llm.profile_extra"},
		{Section: "llm.yaml", Key: "llm.glm.base_url"},
		{Section: "design.yaml", Key: "llm.profile"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filtered = %+v\nwant       %+v", got, want)
	}
}

func TestStripRetiredModelConfig_ReportsEachKeyOnceAsRemoved(t *testing.T) {
	root := stripProject(t, stripLLMFixture, stripWorkflowFixture, "v0.0.0-other")
	var buf bytes.Buffer

	removed, err := stripRetiredModelConfig(&buf, root, ".moai-backups/20260927-000000")
	if err != nil {
		t.Fatalf("stripRetiredModelConfig: %v", err)
	}
	if len(removed) != 5 {
		t.Fatalf("removed %d keys, want 5: %+v", len(removed), removed)
	}
	out := buf.String()
	for _, k := range []string{"llm.profile", "llm.performance_tier", "llm.agent_overrides", "workflow.agent_model_guard", "workflow.workflow_agents"} {
		if n := strings.Count(out, k+" "); n != 1 {
			t.Errorf("report names %s %d times, want exactly once:\n%s", k, n, out)
		}
	}
	if strings.Contains(out, "retained") || strings.Contains(out, "preserved") {
		t.Errorf("report must name the keys as removed, never retained:\n%s", out)
	}
	if !strings.Contains(out, "user values") || !strings.Contains(out, ".moai-backups/20260927-000000") {
		t.Errorf("the user-edited agent_overrides line must say it carried user values and name the backup:\n%s", out)
	}
	if s := sectionBytes(t, root, "llm.yaml"); strings.Contains(s, "profile") || strings.Contains(s, "agent_overrides") {
		t.Errorf("llm.yaml still carries a retired key:\n%s", s)
	}
}

func TestStripRetiredModelConfigOnVersionMatch_BacksUpThenStrips(t *testing.T) {
	root := stripProject(t, stripLLMFixture, stripWorkflowFixture, version.GetVersion())
	cmd, buf := stripTestCmd(false)

	if err := stripRetiredModelConfigOnVersionMatch(cmd, buf, root); err != nil {
		t.Fatalf("host (b): %v", err)
	}
	if s := sectionBytes(t, root, "workflow.yaml"); strings.Contains(s, "agent_model_guard") || strings.Contains(s, "workflow_agents") {
		t.Errorf("workflow.yaml still carries a retired key:\n%s", s)
	}
	dirs := backupDirs(t, root)
	if len(dirs) != 1 {
		t.Fatalf("backup dirs = %v, want exactly one (host (b) takes its own backup)", dirs)
	}
	saved, err := os.ReadFile(filepath.Join(dirs[0], "sections", "llm.yaml"))
	if err != nil {
		t.Fatalf("backup lacks llm.yaml: %v", err)
	}
	if string(saved) != stripLLMFixture {
		t.Errorf("backup llm.yaml = %q, want the original with every key", saved)
	}
	if !strings.Contains(buf.String(), "llm.agent_overrides") {
		t.Errorf("host (b) printed no report:\n%s", buf.String())
	}
}

func TestStripRetiredModelConfigOnVersionMatch_NoOpWithoutMatchOrKeys(t *testing.T) {
	t.Run("version differs", func(t *testing.T) {
		root := stripProject(t, stripLLMFixture, "", "v0.0.0-other")
		cmd, buf := stripTestCmd(false)
		if err := stripRetiredModelConfigOnVersionMatch(cmd, buf, root); err != nil {
			t.Fatal(err)
		}
		if sectionBytes(t, root, "llm.yaml") != stripLLMFixture || backupDirs(t, root) != nil {
			t.Error("a skip that was not a version match changed the project")
		}
	})
	t.Run("forced", func(t *testing.T) {
		root := stripProject(t, stripLLMFixture, "", version.GetVersion())
		cmd, buf := stripTestCmd(true)
		if err := stripRetiredModelConfigOnVersionMatch(cmd, buf, root); err != nil {
			t.Fatal(err)
		}
		if sectionBytes(t, root, "llm.yaml") != stripLLMFixture || backupDirs(t, root) != nil {
			t.Error("--force is not a version-match skip; the project changed")
		}
	})
	t.Run("no retired keys", func(t *testing.T) {
		clean := "llm:\n  mode: \"\"\n"
		root := stripProject(t, clean, "", version.GetVersion())
		cmd, buf := stripTestCmd(false)
		if err := stripRetiredModelConfigOnVersionMatch(cmd, buf, root); err != nil {
			t.Fatal(err)
		}
		if sectionBytes(t, root, "llm.yaml") != clean || backupDirs(t, root) != nil || buf.Len() != 0 {
			t.Errorf("a clean project got a backup or output: backups=%v out=%q", backupDirs(t, root), buf.String())
		}
	})
}

// A user who cancels the merge prompt gets the same (true, nil) as a version
// match; host (b) tells them apart by re-evaluating the version predicate and
// must leave the files byte-identical (design D14 (b), §C cancel row).
func TestUserCancelledMerge_LeavesConfigByteIdentical(t *testing.T) {
	root := stripProject(t, stripLLMFixture, stripWorkflowFixture, "v0.0.0-other")
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	prev := confirmViaPreviewFn
	confirmViaPreviewFn = func(merge.MergeAnalysis, string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmViaPreviewFn = prev })

	cmd, buf := stripTestCmd(false)
	skipped, err := runTemplateSyncWithProgress(cmd)
	if err != nil || !skipped {
		t.Fatalf("cancelled sync = (%v, %v), want (true, nil)", skipped, err)
	}
	if !strings.Contains(buf.String(), "Merge cancelled by user") {
		t.Fatalf("the cancel seam was not reached:\n%s", buf.String())
	}
	if err := stripRetiredModelConfigOnVersionMatch(cmd, buf, "."); err != nil {
		t.Fatal(err)
	}
	if sectionBytes(t, root, "llm.yaml") != stripLLMFixture || sectionBytes(t, root, "workflow.yaml") != stripWorkflowFixture {
		t.Error("a user-cancelled update changed llm.yaml or workflow.yaml")
	}
	if backupDirs(t, root) != nil {
		t.Error("a user-cancelled update took a config backup")
	}
}

// The three hosts sit where design D14 places them.
func TestRetiredModelKeyStrip_HostPlacement(t *testing.T) {
	read := func(name string) string {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	upd := read("update.go")
	skip := strings.Index(upd, "if syncSkipped {")
	if skip < 0 {
		t.Fatal("syncSkipped block not found in update.go")
	}
	blockEnd := strings.Index(upd[skip:], "\n\t}\n")
	if blockEnd < 0 || !strings.Contains(upd[skip:skip+blockEnd], "stripRetiredModelConfigOnVersionMatch(") {
		t.Error("host (b) is not inside the syncSkipped branch of update.go")
	}

	sync := read("update_template_sync.go")
	restore := strings.Index(sync, "backup.RestoreMoaiConfigRetained(")
	strip := strings.Index(sync, "stripRetiredModelConfig(")
	advisory := strings.Index(sync, "renderRetainedKeyAdvisory(")
	if restore < 0 || strip < 0 || advisory < 0 || restore >= strip || strip >= advisory {
		t.Errorf("host (a) must sit after RestoreMoaiConfigRetained and before renderRetainedKeyAdvisory (restore=%d strip=%d advisory=%d)", restore, strip, advisory)
	}
	if !strings.Contains(sync, "confirmViaPreviewFn(") {
		t.Error("runTemplateSyncWithProgress does not call the confirmViaPreviewFn seam")
	}

	clean := read("update_clean_install.go")
	if strings.Contains(clean, "backup.RestoreMoaiConfig(") {
		t.Error("host (c): the clean reinstall still calls RestoreMoaiConfig, which prints every retained key before any strip")
	}
	cRestore := strings.Index(clean, "backup.RestoreMoaiConfigRetained(")
	cStrip := strings.Index(clean, "stripRetiredModelConfig(")
	if cRestore < 0 || cStrip < 0 || cRestore >= cStrip {
		t.Errorf("host (c) must strip after RestoreMoaiConfigRetained (restore=%d strip=%d)", cRestore, cStrip)
	}
}

// With the real embedded template, a retired key the template still ships stays
// in the user's file (code in this build still reads it); a retired key the
// template does not ship is removed. agent_model_guard has never been shipped
// in YAML, so it is removed now.
func TestStripRetiredModelConfig_LeavesKeysTheTemplateStillShips(t *testing.T) {
	shipped, err := template.ShippedRetiredModelKeys()
	if err != nil {
		t.Fatal(err)
	}
	if shipped["workflow.agent_model_guard"] {
		t.Fatal("premise: the embedded workflow.yaml ships agent_model_guard; this test assumes it does not")
	}
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(stripLLMFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(stripWorkflowFixture), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	removed, err := stripRetiredModelConfig(&buf, root, "backup")
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range removed {
		if shipped[k.Key] {
			t.Errorf("stripped %s, which the embedded template still ships", k.Key)
		}
	}
	if s := sectionBytes(t, root, "workflow.yaml"); strings.Contains(s, "agent_model_guard") {
		t.Errorf("agent_model_guard (never shipped) was not removed:\n%s", s)
	}
	llm := sectionBytes(t, root, "llm.yaml")
	for key := range shipped {
		name := strings.TrimPrefix(key, "llm.")
		if name != key && strings.Contains(stripLLMFixture, "  "+name+":") && !strings.Contains(llm, "  "+name+":") {
			t.Errorf("%s is still shipped but was removed from the user's llm.yaml", key)
		}
	}
}
