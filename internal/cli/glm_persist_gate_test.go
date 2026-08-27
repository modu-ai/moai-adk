package cli

// glm_persist_gate_test.go — RC1 regression guards for the launcher-side
// llm.yaml change-detection gate (glm-settings-persist).
//
// persistTeamMode used to rewrite .moai/config/sections/llm.yaml on EVERY
// `moai glm` / `moai cg` launch: a zero-seeded reload + whole-file typed
// re-marshal destroyed hand-written comments, flipped the file mode 0644→0600
// (writeFileAtomic's perm), and touched mtime on every launch — and a
// semantically-identical rewrite reopened a lost-update window against
// concurrent writers. The gate must write ONLY when the desired section state
// actually differs from the persisted one.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// writeHandEditedLLMYAML writes a hand-maintained llm.yaml whose semantic
// content equals the compiled defaults plus the given team_mode — the state a
// second launch would want to persist. Comments and 0644 mode are deliberately
// hand-file properties a typed re-marshal destroys.
func writeHandEditedLLMYAML(t *testing.T, root, teamMode string) string {
	t.Helper()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	var b strings.Builder
	b.WriteString("# llm.yaml — maintained by hand; comments must survive launches\n")
	b.WriteString("llm:\n")
	if teamMode != "" {
		b.WriteString("  team_mode: " + teamMode + " # written by the launcher\n")
	}
	b.WriteString("  glm_env_var: " + config.DefaultGLMEnvVar + "\n")
	b.WriteString("  glm:\n")
	b.WriteString("    base_url: " + config.DefaultGLMBaseURL + "\n")
	b.WriteString("    models:\n")
	b.WriteString("      high: " + config.DefaultGLMHigh + "\n")
	b.WriteString("      medium: " + config.DefaultGLMMedium + "\n")
	b.WriteString("      low: " + config.DefaultGLMLow + "\n")
	b.WriteString("      fable: " + config.DefaultGLMFable + "\n")
	b.WriteString("      # legacy aliases kept in sync by hand\n")
	b.WriteString("      opus: " + config.DefaultGLMOpus + "\n")
	b.WriteString("      sonnet: " + config.DefaultGLMSonnet + "\n")
	b.WriteString("      haiku: " + config.DefaultGLMHaiku + "\n")

	path := filepath.Join(sectionsDir, "llm.yaml")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	return path
}

// assertFileUnchanged fails when the file's bytes or mode differ from the
// snapshot taken before the call under test.
func assertFileUnchanged(t *testing.T, path string, before []byte, modeBefore os.FileMode) {
	t.Helper()
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read llm.yaml after persist: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("llm.yaml was rewritten although the desired state was already persisted:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat llm.yaml: %v", err)
	}
	if info.Mode().Perm() != modeBefore.Perm() {
		t.Errorf("llm.yaml mode flipped %v → %v on a no-op persist", modeBefore.Perm(), info.Mode().Perm())
	}
}

func snapshotLLMYAML(t *testing.T, path string) ([]byte, os.FileMode) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat llm.yaml: %v", err)
	}
	return data, info.Mode()
}

// TestPersistTeamMode_NoRewriteWhenUnchanged pins the RC1 gate: when
// team_mode already equals the target and no default fill is needed, the
// launcher must not rewrite llm.yaml — comments survive, the hand-set 0644
// mode survives, and no lost-update window is reopened.
func TestPersistTeamMode_NoRewriteWhenUnchanged(t *testing.T) {
	root := t.TempDir()
	path := writeHandEditedLLMYAML(t, root, "glm")

	before, modeBefore := snapshotLLMYAML(t, path)
	if err := persistTeamMode(root, "glm"); err != nil {
		t.Fatalf("persistTeamMode: %v", err)
	}
	assertFileUnchanged(t, path, before, modeBefore)
}

// TestPersistTeamMode_WritesWhenTeamModeDiffers keeps the real write path
// alive: a differing team_mode must still be persisted.
func TestPersistTeamMode_WritesWhenTeamModeDiffers(t *testing.T) {
	root := t.TempDir()
	writeHandEditedLLMYAML(t, root, "cg")

	if err := persistTeamMode(root, "glm"); err != nil {
		t.Fatalf("persistTeamMode: %v", err)
	}
	got, err := loadLLMSectionOnly(filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir))
	if err != nil {
		t.Fatalf("loadLLMSectionOnly: %v", err)
	}
	if got.TeamMode != "glm" {
		t.Errorf("team_mode = %q, want %q", got.TeamMode, "glm")
	}
}

// TestPersistTeamMode_FillsEmptySlotsOnFirstWrite keeps the first-launch fill
// behavior: empty GLM slots are still populated with compiled defaults.
func TestPersistTeamMode_FillsEmptySlotsOnFirstWrite(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	minimal := "llm: {}\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(minimal), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	if err := persistTeamMode(root, "glm"); err != nil {
		t.Fatalf("persistTeamMode: %v", err)
	}
	got, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		t.Fatalf("loadLLMSectionOnly: %v", err)
	}
	want := config.NewDefaultLLMConfig()
	if got.GLM.BaseURL != want.GLM.BaseURL || got.GLM.Models.High != want.GLM.Models.High ||
		got.GLM.Models.Medium != want.GLM.Models.Medium || got.GLM.Models.Low != want.GLM.Models.Low ||
		got.GLM.Models.Fable != want.GLM.Models.Fable || got.GLMEnvVar != want.GLMEnvVar {
		t.Errorf("first-launch fill did not populate defaults: got %+v", got)
	}
}

// TestDisableTeamMode_NoRewriteWhenAlreadyNeutral pins the gate on the
// disableTeamMode path: resetting an already-absent team_mode to "" is a no-op
// and must not rewrite the file.
func TestDisableTeamMode_NoRewriteWhenAlreadyNeutral(t *testing.T) {
	root := t.TempDir()
	path := writeHandEditedLLMYAML(t, root, "")

	before, modeBefore := snapshotLLMYAML(t, path)
	if err := disableTeamMode(root); err != nil {
		t.Fatalf("disableTeamMode: %v", err)
	}
	assertFileUnchanged(t, path, before, modeBefore)
}

// TestPersistTeamMode_PreservesUnrelatedKeys pins that a REAL write (team_mode
// transition) still round-trips semantically-unrelated persisted state —
// per-tier GLM effort and context_windows survive the typed re-marshal.
func TestPersistTeamMode_PreservesUnrelatedKeys(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	src := "llm:\n" +
		"  team_mode: cg\n" +
		"  glm:\n" +
		"    base_url: " + config.DefaultGLMBaseURL + "\n" +
		"    effort:\n" +
		"      high: low\n" +
		"      fable: max\n" +
		"    context_windows:\n" +
		"      " + config.DefaultGLM52 + ": 1000000\n" +
		"    models:\n" +
		"      high: " + config.DefaultGLMHigh + "\n" +
		"      medium: " + config.DefaultGLMMedium + "\n" +
		"      low: " + config.DefaultGLMLow + "\n" +
		"      fable: " + config.DefaultGLMFable + "\n" +
		"      opus: " + config.DefaultGLMOpus + "\n" +
		"      sonnet: " + config.DefaultGLMSonnet + "\n" +
		"      haiku: " + config.DefaultGLMHaiku + "\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(src), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	if err := persistTeamMode(root, "glm"); err != nil {
		t.Fatalf("persistTeamMode: %v", err)
	}
	got, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		t.Fatalf("loadLLMSectionOnly: %v", err)
	}
	if got.GLM.Effort.High != "low" || got.GLM.Effort.Fable != "max" {
		t.Errorf("per-tier effort lost across real write: %+v", got.GLM.Effort)
	}
	if got.GLM.ContextWindows[config.DefaultGLM52] != 1000000 {
		t.Errorf("context_windows lost across real write: %+v", got.GLM.ContextWindows)
	}
}
