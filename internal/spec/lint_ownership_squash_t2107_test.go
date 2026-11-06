package spec

import (
	"strings"
	"testing"
)

// lint_ownership_squash_t2107_test.go — card t2107: a squash-merge commit body
// concatenates the Authored-By-Agent trailers of every squashed sub-commit.
// The rule used to read only the FIRST trailer, so the verdict depended on
// sub-commit order rather than on who changed the status. A body naming more
// than one distinct actor cannot be attributed: it is Unmeasured (Info), never
// Invalid. Tests use os.Chdir and must not call t.Parallel().

// squashFixture builds a repo whose spec moves draft -> in-progress in one
// commit carrying the given body, and returns that repo and the spec path.
func squashFixture(t *testing.T, body string) (*gbRepo, string) {
	t.Helper()
	r := newGBRepo(t)
	const id = "SPEC-SQ-001"
	r.spec(id, "draft", "")
	r.commit("feat(SPEC-SQ-001): plan-phase artifacts", "Authored-By-Agent: manager-spec")
	r.spec(id, "in-progress", "")
	r.commit("[t1] feat: squash merge (#1)", body)
	return r, r.specPath(id)
}

func ownershipFindings(t *testing.T, r *gbRepo, specPath string) []Finding {
	t.Helper()
	chdirForTest(t, r.dir)
	doc := &SPECDoc{Path: specPath, Frontmatter: SPECFrontmatter{ID: "SPEC-SQ-001", Status: "in-progress"}}
	return (&OwnershipTransitionRule{}).Check(doc, nil)
}

func TestOwnershipSquashBody_MultiActorIsUnmeasuredNotInvalid(t *testing.T) {
	// First trailer is NOT the owner (manager-spec), the owner's trailer comes later:
	// exactly the shape of squash commits 55b20438fd (#1930) and 5ef49c4625 (#1901).
	body := "squashed\n\nAuthored-By-Agent: manager-spec\n\nAuthored-By-Agent: manager-develop\n\nAuthored-By-Agent: manager-spec"
	r, p := squashFixture(t, body)
	got := ownershipFindings(t, r, p)
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding, got %d: %+v", len(got), got)
	}
	if got[0].Code != "OwnershipTransitionUnmeasured" || got[0].Severity != SeverityInfo {
		t.Fatalf("want Info OwnershipTransitionUnmeasured, got %s/%v: %s", got[0].Code, got[0].Severity, got[0].Message)
	}
	if !strings.Contains(got[0].Message, "manager-develop") || !strings.Contains(got[0].Message, "manager-spec") {
		t.Errorf("message should name the distinct actors: %s", got[0].Message)
	}
}

func TestOwnershipSquashBody_SameActorRepeatedStillJudged(t *testing.T) {
	body := "squashed\n\nAuthored-By-Agent: manager-develop\n\nAuthored-By-Agent: manager-develop"
	r, p := squashFixture(t, body)
	if got := ownershipFindings(t, r, p); len(got) != 0 {
		t.Fatalf("one distinct owner actor must pass silently, got %+v", got)
	}
	// wrong single actor repeated: still Invalid
	body = "squashed\n\nAuthored-By-Agent: manager-docs\n\nAuthored-By-Agent: manager-docs"
	r, p = squashFixture(t, body)
	got := ownershipFindings(t, r, p)
	if len(got) != 1 || got[0].Code != "OwnershipTransitionInvalid" {
		t.Fatalf("one distinct wrong actor must stay Invalid, got %+v", got)
	}
}

func TestOwnershipSquashBody_SingleTrailerUnchanged(t *testing.T) {
	r, p := squashFixture(t, "Authored-By-Agent: manager-docs")
	got := ownershipFindings(t, r, p)
	if len(got) != 1 || got[0].Code != "OwnershipTransitionInvalid" {
		t.Fatalf("single wrong trailer must stay Invalid, got %+v", got)
	}
	r, p = squashFixture(t, "no trailer here")
	got = ownershipFindings(t, r, p)
	if len(got) != 1 || got[0].Code != "OwnershipTransitionUnmeasured" {
		t.Fatalf("no trailer must stay Unmeasured, got %+v", got)
	}
}

// The batched lookup (inside Lint) must agree with the per-SPEC oracle.
func TestOwnershipSquashBody_BatchEqualsOracle(t *testing.T) {
	body := "squashed\n\nAuthored-By-Agent: manager-spec\n\nAuthored-By-Agent: manager-develop"
	r, p := squashFixture(t, body)
	chdirForTest(t, r.dir)
	want, werr := lookupOwnershipTransitionPerSpec(p, "SPEC-SQ-001")
	startGitQueryCache()
	defer stopGitQueryCache()
	got, gerr := lookupOwnershipTransitionFromGit(p, "SPEC-SQ-001")
	if a, b := gbOwnershipString(got, gerr), gbOwnershipString(want, werr); a != b {
		t.Fatalf("batched %q != oracle %q", a, b)
	}
}

func TestDistinctAuthoredByAgents(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"none", "plain", nil},
		{"one", "Authored-By-Agent: Manager-Develop", []string{"manager-develop"}},
		{"dedup_order", "Authored-By-Agent: b\nAuthored-By-Agent: a\nAuthored-By-Agent: B", []string{"b", "a"}},
	}
	for _, c := range cases {
		got := distinctAuthoredByAgents(c.body)
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
