// cutover_runbook_test.go: AC-GFD-023 (SPEC-GITHUB-FLOW-DEFAULT-001 M6, design
// D-8 / D-13 / D-25 / D-29) reads the step table of the cutover runbook
// (.moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/cutover-runbook.md) and judges its
// shape: the step order, exactly one executing subject per row, the shared-system
// tag, the operator-held rows, the post-retirement checks, and the precheck's
// position with the card's own exemption.
//
// The runbook is a Korean-body document; the table columns are located by their
// header names, never by position, so a reordering of columns is not a failure
// but a missing or merged column is.
package template_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

type cvoStep struct {
	id        string
	text      string // the whole row
	subject   string
	shared    string
	stepCell  string
	evidence  string
	lineIndex int
}

var cvoStepIDRe = regexp.MustCompile(`^(\d+)([ab]?)$`)

// cvoParseRunbook returns the runbook text and the rows of its step table (the
// table whose header carries both the subject and the shared-system columns).
func cvoParseRunbook(t *testing.T) (string, []cvoStep) {
	t.Helper()
	data, err := os.ReadFile(cvoScript(t, cvoRunbookRel))
	if err != nil {
		t.Fatalf("the cutover runbook is missing: %v", err)
	}
	text := string(data)
	var steps []cvoStep
	subjectCol, sharedCol, stepCol, evidenceCol := -1, -1, -1, -1
	inTable := false
	for i, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "|") {
			inTable = false
			subjectCol, sharedCol = -1, -1
			continue
		}
		cells := strings.Split(strings.Trim(trim, "|"), "|")
		for j := range cells {
			cells[j] = strings.TrimSpace(cells[j])
		}
		if !inTable {
			// A header row: find the columns by name.
			subjectCol, sharedCol, stepCol, evidenceCol = -1, -1, -1, -1
			for j, c := range cells {
				switch {
				case c == "주체":
					subjectCol = j
				case strings.HasPrefix(c, "외부 공유 시스템"):
					sharedCol = j
				case c == "단계":
					stepCol = j
				case strings.HasPrefix(c, "기록할 증거"):
					evidenceCol = j
				}
			}
			inTable = true
			continue
		}
		if subjectCol < 0 || sharedCol < 0 {
			continue // some other table
		}
		if strings.Trim(strings.Join(cells, ""), "-: ") == "" {
			continue // the separator row
		}
		if len(cells) <= subjectCol || len(cells) <= sharedCol || !cvoStepIDRe.MatchString(cells[0]) {
			continue
		}
		s := cvoStep{id: cells[0], text: trim, subject: cells[subjectCol], shared: cells[sharedCol], lineIndex: i}
		if stepCol >= 0 && stepCol < len(cells) {
			s.stepCell = cells[stepCol]
		}
		if evidenceCol >= 0 && evidenceCol < len(cells) {
			s.evidence = cells[evidenceCol]
		}
		steps = append(steps, s)
	}
	return text, steps
}

func cvoFindStep(steps []cvoStep, id string) (cvoStep, bool) {
	for _, s := range steps {
		if s.id == id {
			return s, true
		}
	}
	return cvoStep{}, false
}

// TestCutoverRunbookShape is AC-GFD-023: five sub-tests, one per Then clause.
func TestCutoverRunbookShape(t *testing.T) {
	text, steps := cvoParseRunbook(t)

	t.Run("order", func(t *testing.T) {
		want := []string{"0", "1", "2", "3", "4", "5a", "5b", "6", "7", "8", "9a", "9b"}
		var got []string
		for _, s := range steps {
			got = append(got, s.id)
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("step ids = %v, want %v", got, want)
		}
		// Each step names its subject matter, in the SPEC's order.
		topics := map[string][]string{
			"0":  {"선행 조건", "CI"},
			"1":  {"레인 정지", "/clear", "정리"},
			"2":  {"병합", "push"},
			"3":  {"사전 점검"},
			"4":  {"흡수"},
			"5b": {"병합"},
			"6":  {"트리 항등", "조상"},
			"7":  {"재기동"},
			"8":  {"첫 카드"},
			"9a": {"퇴역"},
			"9b": {"퇴역"},
		}
		for id, words := range topics {
			s, _ := cvoFindStep(steps, id)
			for _, w := range words {
				if !strings.Contains(s.text, w) {
					t.Errorf("step %s does not mention %q: %s", id, w, s.text)
				}
			}
		}
	})

	t.Run("one_subject_and_a_shared_system_tag_per_row", func(t *testing.T) {
		allowed := map[string]bool{"카드": true, "리더": true, "운영자": true}
		if len(steps) == 0 {
			t.Fatal("no step rows were parsed")
		}
		for _, s := range steps {
			if !allowed[s.subject] {
				t.Errorf("step %s: subject %q is not exactly one of 카드/리더/운영자", s.id, s.subject)
			}
			if !strings.HasPrefix(s.shared, "예") && !strings.HasPrefix(s.shared, "아니오") {
				t.Errorf("step %s: shared-system cell %q does not start with 예 or 아니오", s.id, s.shared)
			}
			if strings.TrimSpace(s.evidence) == "" {
				t.Errorf("step %s: no evidence cell (the executor must record what it observed)", s.id)
			}
		}
	})

	t.Run("operator_steps_are_operator_and_shared", func(t *testing.T) {
		for _, id := range []string{"5b", "9a"} {
			s, ok := cvoFindStep(steps, id)
			if !ok {
				t.Fatalf("step %s is missing", id)
			}
			if s.subject != "운영자" {
				t.Errorf("step %s subject = %q, want 운영자", id, s.subject)
			}
			if !strings.HasPrefix(s.shared, "예") {
				t.Errorf("step %s shared-system cell = %q, want 예...", id, s.shared)
			}
		}
		s9a, _ := cvoFindStep(steps, "9a")
		for _, w := range []string{"Release PR Multi-OS Gate", "develop 보호", "삭제"} {
			if !strings.Contains(s9a.text, w) {
				t.Errorf("step 9a does not mention %q: %s", w, s9a.text)
			}
		}
		// Any row whose subject is the card is a no-shared-system row: the card
		// never changes a shared system (REQ-GFD-020).
		for _, s := range steps {
			if s.subject == "카드" && !strings.HasPrefix(s.shared, "아니오") {
				t.Errorf("step %s: the card is the subject of a shared-system step", s.id)
			}
		}
	})

	t.Run("post_retirement_checks", func(t *testing.T) {
		for _, w := range []string{
			"git grep -n -w develop -- .github/workflows",
			"종료 코드 1",
			"gh run list --workflow=CI --branch develop --limit 3",
			"createdAt",
		} {
			if !strings.Contains(text, w) {
				t.Errorf("the runbook lacks the post-retirement phrase %q", w)
			}
		}
		s9b, _ := cvoFindStep(steps, "9b")
		for _, w := range []string{"git grep -n -w develop -- .github/workflows", "createdAt"} {
			if !strings.Contains(s9b.text, w) {
				t.Errorf("step 9b does not carry %q", w)
			}
		}
	})

	t.Run("exemption_and_precheck_position", func(t *testing.T) {
		s2, _ := cvoFindStep(steps, "2")
		s3, ok3 := cvoFindStep(steps, "3")
		s4, _ := cvoFindStep(steps, "4")
		if !ok3 {
			t.Fatal("step 3 is missing")
		}
		if s2.lineIndex >= s3.lineIndex || s3.lineIndex >= s4.lineIndex {
			t.Errorf("the precheck (step 3, line %d) is not between the bundle merge (step 2, line %d) and the develop absorption (step 4, line %d)",
				s3.lineIndex, s2.lineIndex, s4.lineIndex)
		}
		for _, w := range []string{"scripts/cutover-precheck.sh", "--exclude-card", "D-25"} {
			if !strings.Contains(s3.text, w) {
				t.Errorf("step 3 does not carry %q: %s", w, s3.text)
			}
		}
		if !strings.Contains(s3.text, "2단계") || !strings.Contains(s3.text, "4단계") {
			t.Errorf("step 3 does not state that it runs after step 2 and before step 4: %s", s3.text)
		}
	})
}

// TestCutoverRunbookContent pins the content obligations beyond the AC-GFD-023
// shape: the CI precondition is read by tip SHA (D-17, carry-over D16), the
// rollback and fallback paths are recorded (REQ-GFD-018), the lane relaunch
// forms and the once-per-card-transition /clear are stated, the retirement
// stages follow D-13, and nothing the SPEC holds for the operator is a card step.
func TestCutoverRunbookContent(t *testing.T) {
	text, steps := cvoParseRunbook(t)

	t.Run("the_ci_precondition_is_read_by_tip_sha", func(t *testing.T) {
		s0, ok := cvoFindStep(steps, "0")
		if !ok {
			t.Fatal("step 0 is missing")
		}
		for _, w := range []string{"head_sha", "actions/runs", "gh run view", "`git rev-parse origin/develop`"} {
			if !strings.Contains(s0.text, w) {
				t.Errorf("step 0 does not carry %q: %s", w, s0.text)
			}
		}
		for _, banned := range []string{"gh run list --limit", "gh run list --workflow=CI --branch develop --limit 3"} {
			if strings.Contains(s0.text, banned) {
				t.Errorf("step 0 reads CI through an unstable list call %q: %s", banned, s0.text)
			}
		}
	})

	t.Run("rollback_and_fallback_are_recorded", func(t *testing.T) {
		for _, w := range []string{
			"## 되돌리기", "되돌리기 PR", "pre-merge-refs", "force-push",
			"## 폴백", "분할 PR",
		} {
			if !strings.Contains(text, w) {
				t.Errorf("the runbook lacks %q", w)
			}
		}
	})

	t.Run("lane_stop_and_relaunch_follow_the_kanban_rules", func(t *testing.T) {
		s1, _ := cvoFindStep(steps, "1")
		s7, _ := cvoFindStep(steps, "7")
		for _, w := range []string{"/clear", "착지", "폐기"} {
			if !strings.Contains(s1.text, w) {
				t.Errorf("step 1 does not carry %q", w)
			}
		}
		for _, w := range []string{"moai cc -f", "origin/main", "/clear"} {
			if !strings.Contains(s7.text, w) {
				t.Errorf("step 7 does not carry %q", w)
			}
		}
		if !strings.Contains(text, "카드 전이") {
			t.Errorf("the runbook does not state the once-per-card-transition /clear rule")
		}
	})

	t.Run("develop_retirement_stages_follow_d13", func(t *testing.T) {
		for _, w := range []string{"단계 1", "단계 2", "단계 3", "단계 4", "spec-lint.yml", "재대상", "삭제 뒤"} {
			if !strings.Contains(text, w) {
				t.Errorf("the retirement section lacks %q", w)
			}
		}
		i1 := strings.Index(text, "spec-lint.yml")
		i2 := strings.Index(text, "develop 삭제")
		if i1 < 0 || i2 < 0 || i1 > i2 {
			t.Errorf("the spec-lint.yml develop fetch removal must be stated before the develop deletion (indexes %d, %d)", i1, i2)
		}
	})

	t.Run("operator_held_actions_are_never_card_steps", func(t *testing.T) {
		for _, s := range steps {
			if s.subject != "카드" {
				continue
			}
			for _, w := range []string{"constitution amend", "git push", "gh pr merge", "gh api -X", "branch -D"} {
				if strings.Contains(s.text, w) {
					t.Errorf("a card-subject step %s carries the held action %q", s.id, w)
				}
			}
		}
		for _, w := range []string{"constitution amend", "cutover-confirmation.md", "확인 기록"} {
			if !strings.Contains(text, w) {
				t.Errorf("the runbook lacks %q", w)
			}
		}
	})
}
