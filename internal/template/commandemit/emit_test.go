// emit_test.go — fixture-side emitter tests (AC-005 collision refusal, the
// fail-closed parse contract, and the quoting edge cases). The real-tree
// golden tests live in golden_test.go.
package commandemit_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template/commandemit"
)

// fixtureOptions scopes the fixtures to small synthetic roots.
func fixtureOptions() commandemit.Options {
	return commandemit.Options{
		CommandsRoot: "cmd",
		SkillsRoot:   "skills",
		EmittedRoot:  "out",
	}
}

// fixtureCommand builds one minimal command source with the given
// description raw value and body.
func fixtureCommand(desc, body string) string {
	return "---\n" +
		"description: " + desc + "\n" +
		"argument-hint: \"x\"\n" +
		"allowed-tools: Skill\n" +
		"---\n" +
		"\n" + body
}

func fixtureSkill(name string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte("---\nname: " + name + "\n---\nbody\n")}
}

// fixtureFS builds a synthetic tree: two commands plus a canonical skills
// root. The skills entries simulate the .claude/skills directory names.
func fixtureFS(t *testing.T, skills map[string]*fstest.MapFile) fstest.MapFS {
	t.Helper()
	fsys := fstest.MapFS{
		"cmd/plan.md.tmpl": &fstest.MapFile{Data: []byte(fixtureCommand(
			`{{if eq .ConversationLanguage "ko"}}한국어 설명{{else if eq .ConversationLanguage "ja"}}日本語{{else if eq .ConversationLanguage "zh"}}中文{{else}}Create SPEC documents{{end}}`,
			"Use Skill(\"moai\") with arguments: plan $ARGUMENTS\n"))},
		"cmd/todo.md": &fstest.MapFile{Data: []byte(fixtureCommand(
			"Backlog queue for cards",
			"Use Skill(\"moai\") with arguments: todo $ARGUMENTS\n"))},
	}
	for name, f := range skills {
		fsys["skills/"+name+"/SKILL.md"] = f
	}
	return fsys
}

// TestCollisionRefused is AC-005 (R-004): a derived skill name matching an
// existing canonical skill directory refuses emission with a diagnostic
// naming the derived name and the colliding skill, and no partial artifact
// set is written. Never suffixes or overwrites: the failure carries no
// moai-plan-2-style fallback.
func TestCollisionRefused(t *testing.T) {
	fsys := fixtureFS(t, map[string]*fstest.MapFile{
		"moai-plan":  fixtureSkill("moai-plan"), // collides with cmd/plan
		"moai-other": fixtureSkill("moai-other"),
	})

	pub, err := commandemit.Emit(fsys, fixtureOptions())
	if err == nil {
		t.Fatalf("Emit succeeded over a colliding name; want refusal diagnostic")
	}
	if pub != nil {
		t.Fatalf("collision refusal must produce no artifact set, got %d artifacts", len(pub.Skills))
	}
	for _, want := range []string{"moai-plan", "collides", "canonical skill"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal diagnostic missing %q:\n%s", want, err.Error())
		}
	}
	if strings.Contains(err.Error(), "moai-plan-2") {
		t.Errorf("refusal diagnostic suggests a suffixed fallback:\n%s", err.Error())
	}
}

// TestCollisionRefusedReportsAllCollisions pins that one run names every
// offending pair, not just the first.
func TestCollisionRefusedReportsAllCollisions(t *testing.T) {
	fsys := fixtureFS(t, map[string]*fstest.MapFile{
		"moai-plan": fixtureSkill("moai-plan"),
		"moai-todo": fixtureSkill("moai-todo"),
	})

	_, err := commandemit.Emit(fsys, fixtureOptions())
	if err == nil {
		t.Fatalf("Emit succeeded over two colliding names; want refusal")
	}
	if !strings.Contains(err.Error(), "plan.md.tmpl") || !strings.Contains(err.Error(), "todo.md") {
		t.Errorf("refusal diagnostic must name every colliding source:\n%s", err.Error())
	}
}

// TestNonCollidingEmissionSucceeds keeps the guard honest: the same fixture
// without the colliding skill emits both commands.
func TestNonCollidingEmissionSucceeds(t *testing.T) {
	fsys := fixtureFS(t, map[string]*fstest.MapFile{
		"moai-other": fixtureSkill("moai-other"),
	})

	pub, err := commandemit.Emit(fsys, fixtureOptions())
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if len(pub.Skills) != 2 {
		t.Fatalf("emitted %d artifacts, want 2", len(pub.Skills))
	}
}

// TestParseFailClosed covers the fail-closed parse contract: each bad input
// aborts with a diagnostic naming the file.
func TestParseFailClosed(t *testing.T) {
	cases := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name:    "missing_opening_delimiter",
			source:  "description: x\n---\nbody",
			wantErr: "missing opening",
		},
		{
			name:    "missing_closing_delimiter",
			source:  "---\ndescription: x\nbody without close",
			wantErr: "missing closing",
		},
		{
			name:    "no_description",
			source:  "---\nargument-hint: \"x\"\n---\nbody\n",
			wantErr: "no description",
		},
		{
			name:    "empty_description",
			source:  "---\ndescription:\n---\nbody\n",
			wantErr: "description is empty",
		},
		{
			name:    "conditional_without_else",
			source:  "---\ndescription: {{if eq .ConversationLanguage \"ko\"}}한글{{end}}\n---\nbody\n",
			wantErr: "no {{else}} branch",
		},
		{
			name:    "unbalanced_conditional",
			source:  "---\ndescription: {{if eq .ConversationLanguage \"ko\"}}한글{{else}}x\n---\nbody\n",
			wantErr: "unbalanced template actions",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{"cmd/broken.md.tmpl": &fstest.MapFile{Data: []byte(tc.source)}}
			_, err := commandemit.Emit(fsys, fixtureOptions())
			if err == nil {
				t.Fatalf("Emit succeeded over a broken source; want fail-closed diagnostic")
			}
			if !strings.Contains(err.Error(), "broken.md.tmpl") {
				t.Errorf("diagnostic must name the offending file:\n%s", err.Error())
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("diagnostic missing %q:\n%s", tc.wantErr, err.Error())
			}
		})
	}
}

// TestDescriptionQuoting pins the YAML round trip: a description with a
// colon, a quote, and non-ASCII is emitted quoted and decodes back equal.
func TestDescriptionQuoting(t *testing.T) {
	fsys := fixtureFS(t, map[string]*fstest.MapFile{
		"moai-other": fixtureSkill("moai-other"),
	})
	fsys["cmd/plan.md.tmpl"] = &fstest.MapFile{Data: []byte(fixtureCommand(
		`"Uses: colons, \"quotes\" — and dashes"`,
		"body\n"))}

	pub, err := commandemit.Emit(fsys, fixtureOptions())
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	data := pub.Skills["out/moai-plan/SKILL.md"]
	if data == nil {
		t.Fatalf("moai-plan/SKILL.md not emitted")
	}
	rendered := string(data)
	if !strings.Contains(rendered, `description: "Uses: colons, \"quotes\" — and dashes"`) {
		t.Errorf("description not emitted as escaped double-quoted scalar:\n%s", rendered)
	}
}
