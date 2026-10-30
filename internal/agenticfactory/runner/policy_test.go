package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// stubFacts answers per Issue number; unlisted Issues fail the lookup.
type stubFacts map[int]IssueFacts

func (s stubFacts) IssueFacts(_ context.Context, i core.IssueRef) (IssueFacts, error) {
	f, ok := s[i.Number]
	if !ok {
		return IssueFacts{}, errors.New("host lookup failed")
	}
	return f, nil
}

func foreignFacts() IssueFacts {
	return IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets", UserID: 7, AuthorID: 99, Owner: core.OwnerNone}
}

func collabOn() core.Collaboration {
	return core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{"acme/widgets"}}
}

func policyRunner(t *testing.T, facts Facts, c core.Collaboration) (*Runner, core.IssueRef) {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 10)
	if err != nil {
		t.Fatal(err)
	}
	r := New(st)
	r.Facts, r.Collab = facts, c
	ref, _ := core.NewIssueRef(1948)
	return r, ref
}

func noRecord(t *testing.T, r *Runner, ref core.IssueRef) {
	t.Helper()
	if _, err := r.Store.Read(ref); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("a refused claim left a record: %v", err)
	}
}

// Without a host lookup nobody's identity is established, so nothing is claimed.
func TestClaimWithoutHostFactsIsRefused(t *testing.T) {
	r, ref := policyRunner(t, nil, core.Collaboration{})
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); !errors.Is(err, core.ErrIdentityUnknown) {
		t.Fatalf("no lookup: %v", err)
	}
	noRecord(t, r, ref)
	// A lookup that fails is the same as an unknown identity.
	r, ref = policyRunner(t, stubFacts{}, core.Collaboration{})
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); !errors.Is(err, core.ErrIdentityUnknown) {
		t.Errorf("failing lookup: %v", err)
	}
	noRecord(t, r, ref)
}

func TestOwnIssueIsClaimedWithoutAnyOption(t *testing.T) {
	own := foreignFacts()
	own.AuthorID = own.UserID
	r, ref := policyRunner(t, stubFacts{1948: own}, core.Collaboration{})
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); err != nil {
		t.Fatalf("own issue: %v", err)
	}
}

func TestAnotherAuthorsIssueIsNotClaimedByDefault(t *testing.T) {
	r, ref := policyRunner(t, stubFacts{1948: foreignFacts()}, core.Collaboration{})
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Fatalf("foreign issue: %v", err)
	}
	noRecord(t, r, ref)
}

// A claim on another author's Issue needs the option, a listed repository, a
// confirmed identity and a confirmed owner state, all together.
func TestForeignClaimNeedsEveryCondition(t *testing.T) {
	cases := []struct {
		name   string
		facts  func(*IssueFacts)
		collab func(*core.Collaboration)
		want   error
	}{
		{"all conditions", nil, nil, nil},
		{"owner handed over", func(f *IssueFacts) { f.Owner = core.OwnerHandedOver }, nil, nil},
		{"option off", nil, func(c *core.Collaboration) { c.Enabled = false }, core.ErrNotIssueAuthor},
		{"repository not listed", nil, func(c *core.Collaboration) { c.AllowedRepositories = []string{"acme/other"} }, core.ErrRepoNotAllowed},
		{"renamed repository", func(f *IssueFacts) { f.ActualRepo = "acme/renamed" }, nil, core.ErrRepoUnverified},
		{"redirected lookup", func(f *IssueFacts) { f.Redirected = true }, nil, core.ErrRepoUnverified},
		{"owner active", func(f *IssueFacts) { f.Owner = core.OwnerActive }, nil, core.ErrOwnerUnconfirmed},
		{"owner unknown", func(f *IssueFacts) { f.Owner = core.OwnerUnknown }, nil, core.ErrOwnerUnconfirmed},
		{"user unknown", func(f *IssueFacts) { f.UserID = 0 }, nil, core.ErrIdentityUnknown},
		{"author unknown", func(f *IssueFacts) { f.AuthorID = 0 }, nil, core.ErrIdentityUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, col := foreignFacts(), collabOn()
			if c.facts != nil {
				c.facts(&f)
			}
			if c.collab != nil {
				c.collab(&col)
			}
			r, ref := policyRunner(t, stubFacts{1948: f}, col)
			_, err := r.Claim(ref, "sess-a", "d1", "run-1")
			if c.want == nil {
				if err != nil {
					t.Fatalf("want a claim, got %v", err)
				}
				return
			}
			if !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
			noRecord(t, r, ref)
		})
	}
}

// Collaboration never lifts the rule against a second active claim.
func TestADuplicateClaimIsStillRefusedUnderCollaboration(t *testing.T) {
	r, ref := policyRunner(t, stubFacts{1948: foreignFacts()}, collabOn())
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Claim(ref, "sess-b", "d2", "run-2"); !errors.Is(err, core.ErrHeld) {
		t.Errorf("second active claim: %v", err)
	}
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("same dispatch again: %v", err)
	}
}

func TestAgentClaimsAndRepairClaimsAreGuardedToo(t *testing.T) {
	r, ref := policyRunner(t, stubFacts{1948: foreignFacts()}, core.Collaboration{})
	a, _ := r.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-000000000001")
	if _, err := r.ClaimAs(ref, a, "d1", "run-1"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("ClaimAs: %v", err)
	}
	if _, err := r.ClaimRepair(ref, "sess-a", "d1", "run-1", "WT-x"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("ClaimRepair: %v", err)
	}
	if _, err := r.ClaimRepairAs(ref, a, "d1", "run-1", "WT-x"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("ClaimRepairAs: %v", err)
	}
	noRecord(t, r, ref)
}

// A caller can record the refusal under the right reason.
func TestRefusalsMapToBlockReasons(t *testing.T) {
	for err, want := range map[error]core.BlockReason{
		core.ErrNotIssueAuthor:       core.ReasonNotIssueAuthor,
		core.ErrIdentityUnknown:      core.ReasonIdentityUnknown,
		core.ErrRepoNotAllowed:       core.ReasonRepoNotAllowed,
		core.ErrRepoUnverified:       core.ReasonRepoUnverified,
		core.ErrPermissionNotGranted: core.ReasonPermissionNotGranted,
		core.ErrOwnerUnconfirmed:     core.ReasonOwnerUnconfirmed,
		core.ErrForeignClosing:       core.ReasonForeignClosing,
	} {
		if got, ok := core.ReasonFor(err); !ok || got != want {
			t.Errorf("%v -> %q %v, want %q", err, got, ok, want)
		}
	}
}
