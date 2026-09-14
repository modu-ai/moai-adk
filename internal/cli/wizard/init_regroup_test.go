package wizard

// AC-ITI-018/021/022 (REQ-ITI-017, lead ruling Q5): the init wizard's
// agent_wiring and autonomy_tier questions regroup into one "Agents &
// Autonomy" group — two pages total, stepper denominator 4, and the group
// label itself never renders (no translation keys for it).
//
// The three properties are judged INDEPENDENTLY: AC-ITI-018 counts groups
// (never the denominator), AC-ITI-021 judges the stepper denominator (never
// the group count), and AC-ITI-022 pins the label non-rendering. The two
// mutants named by the AC — splitting the group apart, and adding one more
// question — must flip exactly one of 018/021 each.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// groupReadRe matches a code line that touches a .Group member (the same
// shape AC-ITI-022's render-path clause scans for).
var groupReadRe = regexp.MustCompile(`\.Group([^A-Za-z0-9_.]|$)`)

// TestInitRegroup_TwoPages is AC-ITI-018: buildFormGroups over InitQuestions
// yields exactly 2 groups; conversation_language + user_name open the first,
// agent_wiring + autonomy_tier close the last (in order), both regrouped
// questions carry Group "Agents & Autonomy", and no init question keeps the
// retired "Quality & Workflow" label. The stepper denominator is deliberately
// NOT asserted here (AC-ITI-021's property).
func TestInitRegroup_TwoPages(t *testing.T) {
	questions := InitQuestions("/tmp/init-regroup")
	groups := buildFormGroups(questions, &WizardResult{}, new(string))
	if len(groups) != 2 {
		t.Fatalf("init groups = %d, want exactly 2", len(groups))
	}

	// Membership + relative order (index-independent: the AC-ITI-021 denominator
	// mutant inserts an extra question right after user_name, and this AC must
	// still PASS under it). conversation_language + user_name open the set,
	// agent_wiring + autonomy_tier close it, all four in this relative order.
	position := map[string]int{}
	for i := range questions {
		position[questions[i].ID] = i
	}
	for _, id := range []string{"conversation_language", "user_name", "agent_wiring", "autonomy_tier"} {
		if _, ok := position[id]; !ok {
			t.Fatalf("init set lacks %q", id)
		}
	}
	if position["conversation_language"] >= position["user_name"] ||
		position["user_name"] >= position["agent_wiring"] ||
		position["agent_wiring"] >= position["autonomy_tier"] {
		t.Errorf("init order broken: %v", position)
	}
	for _, tc := range []struct{ id, want string }{
		{"conversation_language", "Basic"},
		{"user_name", "Basic"},
		{"agent_wiring", "Agents & Autonomy"},
		{"autonomy_tier", "Agents & Autonomy"},
	} {
		if got := questions[position[tc.id]].Group; got != tc.want {
			t.Errorf("%s Group = %q, want %q", tc.id, got, tc.want)
		}
	}
	for i := range questions {
		if questions[i].Group == "Quality & Workflow" {
			t.Errorf("%s still carries the retired Group %q", questions[i].ID, "Quality & Workflow")
		}
	}
}

// TestInitStepper_Denominator4 is AC-ITI-021: walking the init form's groups,
// each group's first line carries 4 ●/○ marks and ends "<k> / 4" with k the
// group's first question's visible position. The group count is deliberately
// NOT asserted here (AC-ITI-018's property).
func TestInitStepper_Denominator4(t *testing.T) {
	result := &WizardResult{}
	form := buildUnifiedForm(InitQuestions("/tmp/init-denominator"), result, "")
	d := ptycaptest.NewFormDriver(t, form)
	if got := d.View(); strings.TrimSpace(got) == "" {
		t.Fatal("no group rendered")
	}

	pages := []struct {
		questions int
		first     int
	}{{questions: 2, first: 1}, {questions: 2, first: 3}}
	for i, p := range pages {
		frame := ptycaptest.StripANSI(d.View())
		var firstLine string
		for _, line := range strings.Split(frame, "\n") {
			if strings.TrimSpace(line) != "" {
				firstLine = strings.TrimRight(line, " ")
				break
			}
		}
		if got := strings.Count(firstLine, "●") + strings.Count(firstLine, "○"); got != 4 {
			t.Errorf("page %d first line carries %d ●/○ marks, want 4; line: %q", i+1, got, firstLine)
		}
		if want := strconv.Itoa(p.first) + " / 4"; !strings.HasSuffix(firstLine, want) {
			t.Errorf("page %d first line %q must end with %q", i+1, firstLine, want)
		}
		for range p.questions {
			d.Enter()
		}
	}
	if form.State != huh.StateCompleted {
		t.Fatalf("init form must complete after the last page, state=%v", form.State)
	}
}

// TestGroupLabel_NotRendered is AC-ITI-022: group labels are partition keys
// only — never rendered, and carrying no translation keys.
func TestGroupLabel_NotRendered(t *testing.T) {
	// Fixture half: two unconditional fixture questions with distinct titles
	// whose Group carries the sentinel label; the single page renders both
	// titles and never the sentinel.
	sentinel := "ZZ-GROUP-LABEL-SENTINEL"
	fixture := []Question{
		{ID: "fixture_a", Group: sentinel, Type: QuestionTypeSelect, Title: "Fixture A title",
			Options: []Option{{Label: "fa", Value: "fa"}}, Default: "fa"},
		{ID: "fixture_b", Group: sentinel, Type: QuestionTypeSelect, Title: "Fixture B title",
			Options: []Option{{Label: "fb", Value: "fb"}}, Default: "fb"},
	}
	fixtureForm := buildUnifiedForm(fixture, &WizardResult{}, "")
	fd := ptycaptest.NewFormDriver(t, fixtureForm)
	fixtureFrame := ptycaptest.StripANSI(fd.View())
	for _, title := range []string{"Fixture A title", "Fixture B title"} {
		if !strings.Contains(fixtureFrame, title) {
			t.Errorf("fixture frame lacks %q — the walk did not cover the group", title)
		}
	}
	if strings.Contains(fixtureFrame, sentinel) {
		t.Errorf("fixture frame renders the group label sentinel %q", sentinel)
	}
	fd.Enter()
	fd.Enter()
	if fixtureForm.State != huh.StateCompleted {
		t.Fatalf("fixture form must complete, state=%v", fixtureForm.State)
	}

	// Real init half: page 2 (after answering the Basic page) carries the
	// agent_wiring question title and never the group label. The title string
	// tracks the REQ-IH-013 (SPEC-INIT-HARNESS-001) deployment-consequence
	// wording — this is the render pin, updated with the wording change.
	result := &WizardResult{}
	initForm := buildUnifiedForm(InitQuestions("/tmp/init-label"), result, "")
	id := ptycaptest.NewFormDriver(t, initForm)
	id.Enter() // conversation_language
	id.Enter() // user_name -> regrouped page
	initFrame := ptycaptest.StripANSI(id.View())
	if !strings.Contains(initFrame, "Select the agent harness to deploy and wire") {
		t.Error("init frames lack the agent_wiring title — the walk did not cover the regrouped group")
	}
	if strings.Contains(initFrame, "Agents & Autonomy") {
		t.Error("init frames render the group label Agents & Autonomy")
	}
	for range 2 {
		id.Enter()
	}
	if initForm.State != huh.StateCompleted {
		t.Fatalf("init form must complete, state=%v", initForm.State)
	}

	// Translation-key clause: the label has no key in the translations table.
	data, err := os.ReadFile(filepath.Join("translations.go"))
	if err == nil {
		if strings.Contains(string(data), "Agents & Autonomy") {
			t.Error("translations.go carries a key for the group label Agents & Autonomy")
		}
		if !strings.Contains(string(data), "ConfirmYes") {
			t.Error("control failed: translations.go lacks ConfirmYes (the scan went stale)")
		}
	}

	// Render-path clause: .Group reads in non-test wizard code may only be the
	// huh.Group TYPE NAME or the partition binding-compare/assign inside
	// buildFormGroups and buildProfileForm. Any other read — a render path in
	// any file, new or old — fails this guard.
	allowedPrefix := func(file, line string) bool {
		switch {
		case strings.Contains(line, "huh.Group"):
			return true
		case file == "wizard.go" && (strings.Contains(line, "q.Group != pendingLabel") || strings.Contains(line, "pendingLabel = q.Group")):
			return true
		case file == "profile_wizard.go" && (strings.Contains(line, "q.Group != pendingLabel") || strings.Contains(line, "pendingLabel = q.Group")):
			return true
		}
		return false
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read wizard dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for i, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if groupReadRe.MatchString(trimmed) && !allowedPrefix(name, line) {
				t.Errorf("%s:%d renders or reads .Group outside the partition sites: %q", name, i+1, trimmed)
			}
		}
	}
}

// TestInitRegroup_SecondGroupGolden is the regression guard AC-ITI-021 asks
// for: the regrouped second page (two question titles, stepper ending
// "3 / 4"). It changes under BOTH mutants (group split, extra question), so
// it is a regression guard only — never counted as evidence for either
// property.
func TestInitRegroup_SecondGroupGolden(t *testing.T) {
	result := &WizardResult{}
	form := buildUnifiedForm(InitQuestions("/tmp/init-regroup-golden"), result, "")
	d := ptycaptest.NewFormDriver(t, form)
	d.Enter() // conversation_language
	d.Enter() // user_name -> Agents & Autonomy page
	frame := ptycaptest.StripANSI(d.View())
	if err := ptycaptest.CompareGolden("testdata/axis", "init-regroup-second-group", frame, *updateAxisGolden); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		d.Enter()
	}
	if form.State != huh.StateCompleted {
		t.Fatalf("init form must complete, state=%v", form.State)
	}
}
