package contract

import (
	"slices"
	"testing"
)

// TestAC_CONTRACT_008 — escalation completeness (REQ-CONTRACT-007).
func TestAC_CONTRACT_008(t *testing.T) {
	cases := map[string][]string{
		"five of the six triggers": slices.Clone(sixTriggers[:5]),
		"all six plus boredom":     append(slices.Clone(sixTriggers), "boredom"),
	}
	for i, trigger := range sixTriggers {
		without := slices.Delete(slices.Clone(sixTriggers), i, i+1)
		cases["missing "+trigger] = without
	}
	for name, set := range cases {
		t.Run(name, func(t *testing.T) {
			r := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{escalateOn: set}))))
			if r.Valid || !slices.Contains(r.Reasons, ReasonEscalateOnIncomplete) {
				t.Errorf("escalate_on %v: valid=%v reasons=%v, want %s", set, r.Valid, r.Reasons, ReasonEscalateOnIncomplete)
			}
		})
	}

	t.Run("positive control: exactly the six, reordered", func(t *testing.T) {
		reordered := slices.Clone(sixTriggers)
		slices.Reverse(reordered)
		r := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{escalateOn: reordered}))))
		if !r.Valid || slices.Contains(r.Reasons, ReasonEscalateOnIncomplete) {
			t.Errorf("complete set: valid=%v reasons=%v, want valid", r.Valid, r.Reasons)
		}
	})
}

// TestAC_CONTRACT_021 — second-review and push-develop coupling
// (REQ-CONTRACT-017, REQ-CONTRACT-018).
func TestAC_CONTRACT_021(t *testing.T) {
	verify := func(o fixtureOpts, edit func(*Policy)) Report {
		in := fixtureInputs(signFixture(renderFixture(o)))
		if edit != nil {
			edit(&in.Policy)
		}
		return Verify(in)
	}

	t.Run("second_review required, second_model none", func(t *testing.T) {
		r := verify(fixtureOpts{secondModel: strPtr("none")}, nil)
		if r.Valid || !slices.Contains(r.Reasons, ReasonSecondReviewMissing) {
			t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSecondReviewMissing)
		}
	})

	t.Run("second_review required, second_model empty", func(t *testing.T) {
		r := verify(fixtureOpts{secondModel: strPtr("")}, nil)
		if r.Valid || !slices.Contains(r.Reasons, ReasonSecondReviewMissing) {
			t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSecondReviewMissing)
		}
	})

	for _, mode := range []string{"advisory", "off"} {
		t.Run("second_review "+mode+" accepts none", func(t *testing.T) {
			r := verify(fixtureOpts{secondModel: strPtr("none")}, func(p *Policy) { p.SecondReview = mode })
			if slices.Contains(r.Reasons, ReasonSecondReviewMissing) || !r.Valid {
				t.Errorf("valid=%v reasons=%v, want valid without %s", r.Valid, r.Reasons, ReasonSecondReviewMissing)
			}
			if r.SecondReview != mode {
				t.Errorf("Report.SecondReview = %q, want %q", r.SecondReview, mode)
			}
		})
	}

	t.Run("push_develop false with push-develop action", func(t *testing.T) {
		r := verify(fixtureOpts{}, func(p *Policy) { p.PushDevelop = false })
		if r.Valid || !slices.Contains(r.Reasons, ReasonPushDevelopDisabled) {
			t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonPushDevelopDisabled)
		}
	})

	t.Run("push_develop true: push_requires_lease", func(t *testing.T) {
		r := verify(fixtureOpts{}, func(p *Policy) { p.PushDevelop = true })
		if slices.Contains(r.Reasons, ReasonPushDevelopDisabled) || !r.Valid {
			t.Errorf("valid=%v reasons=%v, want valid", r.Valid, r.Reasons)
		}
		if !r.PushRequiresLease {
			t.Errorf("push_requires_lease = false, want true")
		}
	})

	t.Run("no push-develop action: no lease required", func(t *testing.T) {
		r := verify(fixtureOpts{actions: []string{"commit", "worktree"}}, nil)
		if r.PushRequiresLease || !r.Valid {
			t.Errorf("push_requires_lease=%v valid=%v reasons=%v, want false/valid", r.PushRequiresLease, r.Valid, r.Reasons)
		}
	})
}

// TestAC_CONTRACT_023 — ownership and invariant well-formedness (REQ-CONTRACT-020).
func TestAC_CONTRACT_023(t *testing.T) {
	registry := []string{"CONST-V3R2-001"}
	specGlob := ".moai/specs/" + fixtureSpecID + "/**"
	verify := func(o fixtureOpts) Report {
		if o.invariants == nil {
			o.invariants = []string{"constitution:CONST-V3R2-*", "frozen-files"}
		}
		in := fixtureInputs(signFixture(renderFixture(o)))
		in.RegistryRuleIDs = registry
		return Verify(in)
	}

	t.Run("baseline is valid", func(t *testing.T) {
		if r := verify(fixtureOpts{}); !r.Valid {
			t.Fatalf("baseline: reasons=%v", r.Reasons)
		}
	})

	ownershipInvalid := []struct {
		name string
		o    fixtureOpts
	}{
		{"ownership.write: []", fixtureOpts{write: []string{}}},
		{"write glob /etc/**", fixtureOpts{write: []string{"/etc/**", specGlob}}},
		{"write glob ../x/**", fixtureOpts{write: []string{"../x/**", specGlob}}},
		{"internal/a/** in write and never", fixtureOpts{write: []string{"internal/a/**", specGlob}, never: []string{"internal/a/**"}}},
		{"write [internal/**, .moai/specs/*.md] misses contract.yaml", fixtureOpts{write: []string{"internal/**", ".moai/specs/*.md"}}},
		{"scratch glob ../tmp/**", fixtureOpts{scratch: []string{"../tmp/**"}}},
		{"tmp/** in scratch and never", fixtureOpts{scratch: []string{"tmp/**"}, never: []string{"tmp/**"}}},
	}
	for _, tc := range ownershipInvalid {
		t.Run("ownership_invalid: "+tc.name, func(t *testing.T) {
			r := verify(tc.o)
			if r.Valid || !slices.Contains(r.Reasons, ReasonOwnershipInvalid) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonOwnershipInvalid)
			}
		})
	}

	t.Run("invariant_unresolved: constitution:CONST-NOPE-*", func(t *testing.T) {
		r := verify(fixtureOpts{invariants: []string{"constitution:CONST-NOPE-*"}})
		if r.Valid || !slices.Contains(r.Reasons, ReasonInvariantUnresolved) {
			t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonInvariantUnresolved)
		}
	})

	for _, glob := range []string{".moai/**", ".moai/specs/**", specGlob} {
		t.Run("write "+glob+" covers contract.yaml", func(t *testing.T) {
			r := verify(fixtureOpts{write: []string{"internal/fixture/**", glob}})
			if slices.Contains(r.Reasons, ReasonOwnershipInvalid) || !r.Valid {
				t.Errorf("valid=%v reasons=%v, want valid without %s", r.Valid, r.Reasons, ReasonOwnershipInvalid)
			}
		})
	}

	t.Run("constitution:CONST-V3R2-* resolves", func(t *testing.T) {
		r := verify(fixtureOpts{invariants: []string{"constitution:CONST-V3R2-*"}})
		if slices.Contains(r.Reasons, ReasonInvariantUnresolved) || !r.Valid {
			t.Errorf("valid=%v reasons=%v, want valid without %s", r.Valid, r.Reasons, ReasonInvariantUnresolved)
		}
	})
}
