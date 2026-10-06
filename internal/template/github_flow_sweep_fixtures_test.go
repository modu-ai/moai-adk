package template

import (
	"fmt"
	"strings"
	"testing"
)

// The fixture self-test of the github-flow sweep guard (AC-GFD-021, design D-9/D-27,
// card t1453 M4 step 1). It exercises the same sweepScan the tree run uses, on inputs
// that do not depend on the real tree, so it can land before the cutover.
//
// Rows F-red-1..14, F-green-1..10 and F-empty are the 25 rows of AC-GFD-021 verbatim.
// The rows after them are the carry-over additions (progress.md §J.3/§J.5 D15, D18) and
// the mutant-resistance rows: each pins a decision the 25 rows leave to the class or the
// ratchet. Every row asserts the CLASS and the RULE, not just violation-or-not, because
// since D-27 a weak hit is a violation too — a mutant that downgrades strong to weak is
// invisible to a boolean check.

type sweepWant struct {
	Line  int
	Class sweepClass
	Rule  string
}

type sweepWantExempt struct {
	Line int
	Rule string
}

type sweepFixture struct {
	name    string
	file    string
	content string
	allow   []sweepAllow
	viol    []sweepWant       // exactly the violating lines expected; empty = none
	exempt  []sweepWantExempt // each must be present as an exempt finding with this rule
	none    bool              // green row that must produce no finding at all
}

func sweepFixtures() []sweepFixture {
	const f = "docs/sample.md"
	return []sweepFixture{
		{name: "F-red-1", file: f, content: "카드 브랜치는 develop 에서 만든다\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}},
		{name: "F-red-2", file: f, content: "develop에서 분기한다\n",
			viol: []sweepWant{{1, sweepStrong, "P-D"}}},
		{name: "F-red-3", file: f, content: "```sh\ngit switch -c x origin/develop\n```\n",
			viol: []sweepWant{{2, sweepStrong, "P-A"}}},
		{name: "F-red-4", file: f, content: "then merge into origin/develop when green\n",
			viol: []sweepWant{{1, sweepStrong, "P-A"}}},
		{name: "F-red-5", file: f, content: "push to develop\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}},
		{name: "F-red-6", file: f, content: "Check out develop, pull, delete the local branch\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}},
		{name: "F-red-7", file: f, content: "compare against refs/remotes/origin/develop\n",
			viol: []sweepWant{{1, sweepStrong, "P-A"}}},
		{name: "F-red-8", file: f, content: "develop을 기준으로 만든다\n",
			viol: []sweepWant{{1, sweepStrong, "P-D"}}},
		{name: "F-red-9", file: f, content: "카드 브랜치는 develop 에서 만든다 (legacy)\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}},
		{name: "F-red-10", file: f,
			content: "## RETIRED 2026-10-02 old chain\n\nCards merge into develop here.\n\n## Current rules\n\n카드는 develop 에 병합한다\n",
			viol:    []sweepWant{{7, sweepStrong, "P-B"}},
			exempt:  []sweepWantExempt{{3, "marker-section"}}},
		{name: "F-red-11", file: f, content: "card work flows on develop-based worktrees\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}},
		{name: "F-red-12", file: f, content: "Work starts from develop.\n",
			viol: []sweepWant{{1, sweepWeak, "weak"}}},
		{name: "F-red-13", file: f, content: "Lane PRs go to develop.\n",
			viol: []sweepWant{{1, sweepWeak, "weak"}}},
		{name: "F-red-14", file: f, content: "Compare with the develop tip.\n",
			allow: []sweepAllow{{File: f, Literal: "develop tip", Why: "partial literal"}},
			viol:  []sweepWant{{1, sweepWeak, "weak"}}},

		{name: "F-green-1", file: f, content: "RETIRED 2026-10-02: card branches fork from develop\n",
			exempt: []sweepWantExempt{{1, "marker-line"}}},
		{name: "F-green-2", file: f, content: "the manager-develop agent runs it\n", none: true},
		{name: "F-green-3", file: f, content: "see .claude/worktrees/develop for the tree\n", none: true},
		{name: "F-green-4", file: f, content: "to develop a feature\n",
			exempt: []sweepWantExempt{{1, "verb"}}},
		{name: "F-green-5", file: f, content: "the tip is `develop @ 0cca34439` today\n",
			exempt: []sweepWantExempt{{1, "stamp"}}},
		{name: "F-green-6", file: f, content: "pass --develop-worktree to it\n", none: true},
		{name: "F-green-7", file: f, content: "actions `push-develop` and `push_develop` stay\n", none: true},
		{name: "F-green-8", file: f,
			content: "### RETIRED 2026-10-02 old flow\n\nCards fork from develop.\n\n#### Sub step\n\nMerge into develop first.\n\n### Next live section\n\nplain text\n",
			exempt:  []sweepWantExempt{{3, "marker-section"}, {7, "marker-section"}}},
		{name: "F-green-9", file: f, content: "the develop-worktree helper and develop-notes\n", none: true},
		{name: "F-green-10", file: f, content: "Compare with the develop tip.\n",
			allow:  []sweepAllow{{File: f, Literal: "Compare with the develop tip.", Why: "reviewed: names the tip ref in a reference sentence"}},
			exempt: []sweepWantExempt{{1, "allow"}}},

		// Carry-over additions (progress.md §J.3 / §J.5).
		{name: "F-red-15", file: f, content: "Lanes merge to develop and push.\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}}, // D15: the verb exclusion must not run before the branch-word rule
		{name: "F-red-16", file: f, content: "RETIRED: card branches fork from develop\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}}, // a marker without date or card id does not exempt
		{name: "F-red-17", file: f, content: "Work starts from develop.\n",
			allow: []sweepAllow{{File: f, Literal: "develop", Why: "bare token"}},
			viol:  []sweepWant{{1, sweepWeak, "weak"}}}, // a bare-token entry silences nothing
		{name: "F-red-18", file: f, content: "작업은 develop에서 시작한다\n",
			viol: []sweepWant{{1, sweepStrong, "P-D"}}}, // CJK neighbour with no branch-word token: P-D alone carries it
		{name: "F-red-19", file: f, content: "the develop 'branch' ref and a to develop a thing\n",
			viol: []sweepWant{{1, sweepStrong, "P-B"}}}, // a verb phrase on a line that also names a branch is not exempt
		{name: "F-red-20", file: f, content: "Allow entry for another file does not apply.\nWork starts from develop.\n",
			allow: []sweepAllow{{File: "docs/other.md", Literal: "Work starts from develop.", Why: "wrong file"}},
			viol:  []sweepWant{{2, sweepWeak, "weak"}}},
		{name: "F-red-21", file: f, content: "```text\nWork starts from develop.\n```\n",
			viol: []sweepWant{{2, sweepStrong, "P-A"}}}, // a fenced line is P-A with no other signal on it
		{name: "F-red-22", file: f, content: "Work starts from `develop` today.\n",
			viol: []sweepWant{{1, sweepStrong, "P-A"}}}, // an inline code span is P-A with no other signal on it
		{name: "F-green-11", file: f, content: "`develop @ 0cca34439` and develop @ 0123abc\n",
			exempt: []sweepWantExempt{{1, "stamp"}}},
		{name: "F-green-12", file: f, content: "> RETIRED 2026-10-02 quoted note\n> Cards fork from develop.\n\nplain text\n",
			exempt: []sweepWantExempt{{2, "marker-quote"}}},
	}
}

func TestGitHubFlowSweepFixtures(t *testing.T) {
	for _, fx := range sweepFixtures() {
		t.Run(fx.name, func(t *testing.T) {
			got := sweepScan(fx.file, fx.content, fx.allow)
			var viol []sweepWant
			for _, g := range got {
				if g.violation() {
					viol = append(viol, sweepWant{g.Line, g.Class, g.Rule})
				}
			}
			if fmt.Sprint(viol) != fmt.Sprint(fx.viol) {
				t.Errorf("violations = %v, want %v\nfindings: %v", viol, fx.viol, got)
			}
			for _, we := range fx.exempt {
				found := false
				for _, g := range got {
					if g.Class == sweepExempt && g.Line == we.Line && g.Rule == we.Rule {
						found = true
					}
				}
				if !found {
					t.Errorf("no exempt finding line %d rule %q; findings: %v", we.Line, we.Rule, got)
				}
			}
			if fx.none && len(got) != 0 {
				t.Errorf("expected no finding at all, got %v", got)
			}
		})
	}

	// F-empty: a sweep that visited nothing is a failing guard, never a pass.
	t.Run("F-empty", func(t *testing.T) {
		subtrees := []sweepSubtree{{Name: "a", Floor: 5}, {Name: "b", Floor: 2}}
		probs := sweepVisitProblems(map[string]int{"a": 0, "b": 0}, subtrees, 9)
		if len(probs) == 0 {
			t.Fatal("visited=0 produced no problem — an empty sweep would read as a pass")
		}
		if !strings.Contains(strings.Join(probs, "\n"), "empty sweep") {
			t.Errorf("problems do not name the empty sweep: %v", probs)
		}
	})

	// Visit floors: the total floor and the per-subtree floors move independently.
	t.Run("F-floor-1", func(t *testing.T) {
		subtrees := []sweepSubtree{{Name: "a", Floor: 5}, {Name: "b", Floor: 2}}
		if probs := sweepVisitProblems(map[string]int{"a": 6, "b": 3}, subtrees, 20); len(probs) == 0 {
			t.Error("total below the total floor was accepted")
		}
	})
	t.Run("F-floor-2", func(t *testing.T) {
		subtrees := []sweepSubtree{{Name: "a", Floor: 5}, {Name: "b", Floor: 2}}
		probs := sweepVisitProblems(map[string]int{"a": 100, "b": 0}, subtrees, 20)
		if len(probs) == 0 || !strings.Contains(strings.Join(probs, "\n"), "b") {
			t.Errorf("dropping a whole subtree (total still above its floor) was accepted: %v", probs)
		}
	})
	t.Run("F-floor-3", func(t *testing.T) {
		subtrees := []sweepSubtree{{Name: "a", Floor: 5}, {Name: "b", Floor: 2}}
		if probs := sweepVisitProblems(map[string]int{"a": 6, "b": 3}, subtrees, 9); len(probs) != 0 {
			t.Errorf("a healthy sweep was rejected: %v", probs)
		}
	})

	// P-C: a token-less live sentence is counted for the ratchet, never a violation.
	t.Run("F-pc-1", func(t *testing.T) {
		got := sweepScan("docs/sample.md", "Cards go through the integration worktree.\n일괄 push 는 리더의 일이다\n", nil)
		var pc int
		for _, g := range got {
			if g.violation() {
				t.Errorf("a P-C line was a violation: %v", g)
			}
			if g.Class == sweepRatchet && g.Rule == "P-C" {
				pc++
			}
		}
		if pc != 2 {
			t.Errorf("P-C lines = %d, want 2; findings: %v", pc, got)
		}
	})

	// Marker channel ratchet (progress.md §J.5 D18): a fake marker silences live text, so the
	// number of marker-exempted lines is a ratchet that can fall but not rise.
	t.Run("F-marker-1", func(t *testing.T) {
		c := sweepCountsOf(sweepScan("docs/sample.md", "# Guide (RETIRED 2026-10-02)\n\nCards fork from develop.\n", nil))
		if c.MarkerH1Lines != 1 || c.MarkerLines != 0 {
			t.Fatalf("H1 marker counts = %+v, want MarkerH1Lines=1 MarkerLines=0", c)
		}
		if len(sweepRatchetProblems(c, sweepCeilings{})) == 0 {
			t.Error("a document-wide H1 marker silencing live text was not caught by a zero ceiling")
		}
	})
	t.Run("F-marker-2", func(t *testing.T) {
		c := sweepCountsOf(sweepScan("docs/sample.md", "Cards fork from develop (RETIRED 2026-10-02)\n", nil))
		if c.MarkerLines != 1 {
			t.Fatalf("dated marker on a live line counts = %+v, want MarkerLines=1", c)
		}
		if len(sweepRatchetProblems(c, sweepCeilings{})) == 0 {
			t.Error("a dated marker appended to a live line was not caught by a zero ceiling")
		}
	})
	t.Run("F-marker-3", func(t *testing.T) {
		c := sweepCountsOf(sweepScan("docs/sample.md", "## RETIRED 2026-10-02 fake\n\nCards fork from develop.\nNew cards merge into develop.\n", nil))
		if c.MarkerLines != 2 {
			t.Fatalf("fake section marker counts = %+v, want MarkerLines=2", c)
		}
		if len(sweepRatchetProblems(c, sweepCeilings{MarkerLines: 1})) == 0 {
			t.Error("a fake heading marker that adds exempt lines past the ceiling was not caught")
		}
	})
	t.Run("F-marker-4", func(t *testing.T) {
		c := sweepCountsOf(sweepScan("docs/sample.md", "> RETIRED 2026-10-02 quoted note\n> Cards fork from develop.\n", nil))
		if c.MarkerLines != 1 {
			t.Fatalf("quote marker counts = %+v, want MarkerLines=1", c)
		}
	})
	t.Run("F-marker-5", func(t *testing.T) {
		// Green direction: counts at or below the ceilings are not a problem.
		c := sweepCounts{MarkerLines: 3, MarkerH1Lines: 1, PCLines: 4}
		if probs := sweepRatchetProblems(c, sweepCeilings{MarkerLines: 3, MarkerH1Lines: 1, PCLines: 4}); len(probs) != 0 {
			t.Errorf("counts equal to the ceilings were a problem: %v", probs)
		}
		if probs := sweepRatchetProblems(c, sweepCeilings{MarkerLines: 9, MarkerH1Lines: 9, PCLines: 9}); len(probs) != 0 {
			t.Errorf("counts below the ceilings were a problem: %v", probs)
		}
		if probs := sweepRatchetProblems(sweepCounts{PCLines: 5}, sweepCeilings{PCLines: 4}); len(probs) == 0 {
			t.Error("a P-C count above its ceiling was accepted")
		}
	})

	// Allow-list shape: entries that would hollow the list out are rejected.
	t.Run("F-allow-1", func(t *testing.T) {
		ok := []sweepAllow{{File: "a.md", Literal: "Compare with the develop tip.", Why: "reviewed"}}
		if err := sweepValidateAllow(ok, 40); err != nil {
			t.Errorf("a well-formed entry was rejected: %v", err)
		}
		for name, bad := range map[string]sweepAllow{
			"no file":        {File: "", Literal: "Compare with the develop tip.", Why: "x"},
			"no literal":     {File: "a.md", Literal: "  ", Why: "x"},
			"bare token":     {File: "a.md", Literal: "develop", Why: "x"},
			"no why":         {File: "a.md", Literal: "Compare with the develop tip.", Why: " "},
			"origin token":   {File: "a.md", Literal: "origin/develop", Why: "x"},
			"hyphenated bit": {File: "a.md", Literal: "develop-tip", Why: "x"},
		} {
			if err := sweepValidateAllow([]sweepAllow{bad}, 40); err == nil {
				t.Errorf("entry %q (%+v) was accepted", name, bad)
			}
		}
	})
	t.Run("F-allow-2", func(t *testing.T) {
		var many []sweepAllow
		for i := 0; i < 41; i++ {
			many = append(many, sweepAllow{File: "a.md", Literal: fmt.Sprintf("Line number %d about develop tip.", i), Why: "x"})
		}
		if err := sweepValidateAllow(many, 40); err == nil {
			t.Error("41 entries passed a cap of 40 — the cap is the price of adding an entry")
		}
		if err := sweepValidateAllow(many[:40], 40); err != nil {
			t.Errorf("40 entries (the cap) were rejected: %v", err)
		}
	})
	t.Run("F-allow-3", func(t *testing.T) {
		// An entry is "used" only when it exempted a would-be violation: the indices of the
		// used entries come back on the findings, so a stale entry is detectable.
		allow := []sweepAllow{
			{File: "a.md", Literal: "Work starts from develop.", Why: "reviewed"},
			{File: "a.md", Literal: "This line is not in the file.", Why: "stale"},
		}
		used := map[int]bool{}
		for _, g := range sweepScan("a.md", "Work starts from develop.\n", allow) {
			if g.Class == sweepExempt && g.Rule == "allow" {
				used[g.AllowIdx] = true
			}
		}
		if !used[0] || used[1] {
			t.Errorf("used entries = %v, want exactly {0}", used)
		}
	})
}
