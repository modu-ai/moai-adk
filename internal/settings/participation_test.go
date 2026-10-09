package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// participationFixture builds a project root whose feedback section file
// exists, points MOAI_HOME at a temporary directory, and returns both roots.
func participationFixture(t *testing.T) (projectRoot, home string) {
	t.Helper()
	projectRoot = t.TempDir()
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "feedback:\n  repository: modu-ai/moai-adk\n  auto_submit: false\n"
	if err := os.WriteFile(filepath.Join(dir, "feedback.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write feedback.yaml: %v", err)
	}
	home = t.TempDir()
	t.Setenv("MOAI_HOME", home)
	return projectRoot, home
}

func userParticipationPath(home string) string {
	return filepath.Join(home, "config", "participation.yaml")
}

func readHomeParticipation(t *testing.T, home string) config.UserParticipation {
	t.Helper()
	return config.ReadUserParticipation()
}

func TestFeedbackParticipationFieldIsBool(t *testing.T) {
	f, ok := Field(ParticipationField)
	if !ok {
		t.Fatalf("schema has no %q field", ParticipationField)
	}
	if f.Type != TypeBool {
		t.Fatalf("field type = %q, want bool (a checkbox widget would fail TestBoolFieldsRenderAsRadio)", f.Type)
	}
	if f.Persist.Kind != PersistUserScoped {
		t.Fatalf("persist kind = %q, want %q", f.Persist.Kind, PersistUserScoped)
	}
	// The absent-key polarity is declared: an absent key reads false, so an
	// unchanged submission that submits false is a no-op.
	if f.AbsentDefault != "false" {
		t.Fatalf("AbsentDefault = %q, want \"false\"", f.AbsentDefault)
	}
	if f.Section != SectionFeedback {
		t.Fatalf("section = %q, want feedback", f.Section)
	}
	// participation_asked has no FieldDef: it must not be rendered anywhere.
	for _, all := range AllFields() {
		if strings.HasSuffix(all.Name, "participation_asked") {
			t.Fatalf("asked marker key %q is a schema field; it must never render", all.Name)
		}
	}
}

func TestUserScopedEditWritesHomeFileOnly(t *testing.T) {
	projectRoot, home := participationFixture(t)
	projFile := filepath.Join(projectRoot, ".moai", "config", "sections", "feedback.yaml")
	before, err := os.ReadFile(projFile)
	if err != nil {
		t.Fatalf("read project feedback.yaml: %v", err)
	}

	if err := ApplySchemaEdits(projectRoot, map[string]string{
		ParticipationField: "true",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	up := readHomeParticipation(t, home)
	if !up.Enabled || !up.Asked {
		t.Fatalf("user-scoped value after save = %+v, want enabled+asked", up)
	}

	after, err := os.ReadFile(projFile)
	if err != nil {
		t.Fatalf("read project feedback.yaml: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("project feedback.yaml changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestUserScopedValueInvariantTouchesNothing(t *testing.T) {
	projectRoot, home := participationFixture(t)
	consent := userParticipationPath(home)

	// (a) absent key + a false submission is a no-op: absence already IS the
	// unset state.
	if err := ApplySchemaEdits(projectRoot, map[string]string{
		ParticipationField: "false",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits (absent, false): %v", err)
	}
	if _, err := os.Stat(consent); !os.IsNotExist(err) {
		t.Fatalf("consent file created by a false submission over an absent key: %v", err)
	}

	// (b) an unchanged submission over a persisted value writes nothing.
	if err := ApplySchemaEdits(projectRoot, map[string]string{
		ParticipationField: "true",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits (enable): %v", err)
	}
	info, err := os.Stat(consent)
	if err != nil {
		t.Fatalf("stat consent after enable: %v", err)
	}
	beforeBody, err := os.ReadFile(consent)
	if err != nil {
		t.Fatalf("read consent: %v", err)
	}
	if err := ApplySchemaEdits(projectRoot, map[string]string{
		ParticipationField: "true",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits (unchanged): %v", err)
	}
	info2, err := os.Stat(consent)
	if err != nil {
		t.Fatalf("stat consent after unchanged save: %v", err)
	}
	if !info.ModTime().Equal(info2.ModTime()) {
		t.Fatal("unchanged submission rewrote the consent file (mtime moved)")
	}
	afterBody, _ := os.ReadFile(consent)
	if string(beforeBody) != string(afterBody) {
		t.Fatalf("unchanged submission changed the consent file:\n%s\n%s", beforeBody, afterBody)
	}
}

func TestUserScopedWriteSetsAsked(t *testing.T) {
	projectRoot, home := participationFixture(t)

	// A change of enabled also sets asked true — a user who toggled in the
	// console before any prompt is never asked again (the console description
	// carries the full disclosure, so the consent is informed).
	if err := ApplySchemaEdits(projectRoot, map[string]string{
		ParticipationField: "true",
	}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	up := readHomeParticipation(t, home)
	if !up.Asked {
		t.Fatalf("console save left asked false: %+v", up)
	}

	// The writer preserves an existing repository key (unknown-key survival).
	if err := WriteUserParticipation(config.UserParticipation{
		Enabled:    true,
		Asked:      true,
		Repository: "someone/mirror",
	}); err != nil {
		t.Fatalf("WriteUserParticipation: %v", err)
	}
	up = readHomeParticipation(t, home)
	if up.Repository != "someone/mirror" {
		t.Fatalf("repository key lost: %+v", up)
	}

	// The write is direct to the user file, not through the project seam:
	// the project feedback.yaml stays untouched.
	projFile := filepath.Join(projectRoot, ".moai", "config", "sections", "feedback.yaml")
	body, err := os.ReadFile(projFile)
	if err != nil {
		t.Fatalf("read project feedback.yaml: %v", err)
	}
	if strings.Contains(string(body), "participation") {
		t.Fatalf("project feedback.yaml gained a participation key:\n%s", body)
	}
}
