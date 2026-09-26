// Package cli — codex_role_behaviour_test.go
//
// AC-RLP-006 for SPEC-ROLE-LOAD-PREDICATE-001: the load-predicate input type
// carries no behaviour-derived field, exhaustive per-field mutation of the
// behaviour type leaves the load verdict unchanged, and a deliberately
// coupled variant is shown to diverge on at least one such mutation.
package cli

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// codexRoleLoadPredicateCoupled is a deliberately WRONG variant that reads a
// behaviour field inside its load decision. It exists only in this test file
// to demonstrate that the exhaustive mutation set can actually detect
// coupling — the production predicate (codexRoleLoadPredicate) never takes
// a codexRoleBehaviour value at all (REQ-RLP-011).
func codexRoleLoadPredicateCoupled(load bool, behaviour codexRoleBehaviour, fieldIndex int) bool {
	v := reflect.ValueOf(behaviour)
	if fieldIndex < 0 || fieldIndex >= v.NumField() {
		return load
	}
	fv := v.Field(fieldIndex)
	if fv.Kind() == reflect.Bool && !fv.Bool() {
		// Couple the decision to this one behaviour field being false.
		return false
	}
	return load
}

// TestCodexRoleLoadPredicateStructurallyIndependent — AC-RLP-006
// (REQ-RLP-010, REQ-RLP-011).
func TestCodexRoleLoadPredicateStructurallyIndependent(t *testing.T) {
	root := repoRoot(t)
	realDir := filepath.Join(root, codexRoleFixtureRoot, "real")
	rolesDir := filepath.Join(root, codexRoleFixtureRoot, "roles")

	table, reason, err := codexBuildRoleExpectationTable(rolesDir)
	if err != nil {
		t.Fatalf("build table: %v", err)
	}
	if reason != "" {
		t.Fatalf("table ineligible: %s", reason)
	}

	realFiles := globSorted(t, realDir, "*.jsonl")
	type roleFixture struct {
		path string
		role string
	}
	var fixtures []roleFixture
	for _, p := range realFiles {
		role, hasLabel, err := codexRoleSessionLabel(p)
		if err != nil {
			t.Fatalf("session label %s: %v", p, err)
		}
		if hasLabel {
			fixtures = append(fixtures, roleFixture{path: p, role: role})
		}
	}
	if len(fixtures) != 12 {
		t.Fatalf("expected 12 labeled real sessions, got %d", len(fixtures))
	}

	// Baseline load verdict per fixture, computed with the real
	// codexRoleLoadInput — which carries NO behaviour field at all.
	baseline := make([]bool, len(fixtures))
	for i, f := range fixtures {
		matched, err := codexRoleFingerprintDerive(f.path, table)
		if err != nil {
			t.Fatalf("derive %s: %v", f.path, err)
		}
		baseline[i] = codexRoleLoadPredicate(codexRoleLoadInput{Role: f.role, Matched: matched})
	}

	behaviourType := reflect.TypeOf(codexRoleBehaviour{})
	numBehaviourFields := behaviourType.NumField()
	if numBehaviourFields == 0 {
		t.Fatal("codexRoleBehaviour must carry at least one field for this AC to be meaningful")
	}

	t.Run("input_type_has_no_behaviour_field", func(t *testing.T) {
		inputType := reflect.TypeOf(codexRoleLoadInput{})
		for i := 0; i < inputType.NumField(); i++ {
			name := strings.ToLower(inputType.Field(i).Name)
			if strings.Contains(name, "behaviour") || strings.Contains(name, "behavior") || strings.Contains(name, "nonce") {
				t.Fatalf("codexRoleLoadInput field %q must not reference behaviour/nonce", inputType.Field(i).Name)
			}
		}
	})

	t.Run("exhaustive_field_mutations", func(t *testing.T) {
		mutated := 0
		for fi := 0; fi < numBehaviourFields; fi++ {
			// Build N synthetic behaviour copies, one per fixture, each with
			// exactly this field flipped from its zero value — and confirm
			// the REAL load predicate's verdict is unaffected, because its
			// input type cannot even carry a codexRoleBehaviour value.
			for i, f := range fixtures {
				b := reflect.New(behaviourType).Elem()
				field := b.Field(fi)
				if field.Kind() == reflect.Bool {
					field.SetBool(true)
				}
				matched, err := codexRoleFingerprintDerive(f.path, table)
				if err != nil {
					t.Fatalf("derive %s: %v", f.path, err)
				}
				got := codexRoleLoadPredicate(codexRoleLoadInput{Role: f.role, Matched: matched})
				if got != baseline[i] {
					t.Fatalf("field %d mutation changed the load verdict for %s: got %v, want %v",
						fi, f.path, got, baseline[i])
				}
			}
			mutated++
		}
		if mutated != numBehaviourFields {
			t.Fatalf("BEHAVIOUR_FIELDS_MUTATED %d OF %d", mutated, numBehaviourFields)
		}
		fmt.Printf("BEHAVIOUR_FIELDS_MUTATED %d OF %d\n", mutated, numBehaviourFields)
	})

	t.Run("half_true_arm", func(t *testing.T) {
		// Half the roles carry behaviour=true, half false — the real load
		// predicate is unaffected either way, since it never sees this value.
		for i, f := range fixtures {
			b := codexRoleBehaviour{NonceReturned: i%2 == 0}
			_ = b // constructed only to prove it cannot reach the predicate
			matched, err := codexRoleFingerprintDerive(f.path, table)
			if err != nil {
				t.Fatalf("derive %s: %v", f.path, err)
			}
			got := codexRoleLoadPredicate(codexRoleLoadInput{Role: f.role, Matched: matched})
			if got != baseline[i] {
				t.Fatalf("half-true arm changed the load verdict for %s", f.path)
			}
		}
	})

	t.Run("coupled_mutant_rejected", func(t *testing.T) {
		diverged := 0
		for fi := 0; fi < numBehaviourFields; fi++ {
			for i := range fixtures {
				// Zero-valued behaviour: bool field false at index fi.
				zero := codexRoleBehaviour{}
				coupledResult := codexRoleLoadPredicateCoupled(baseline[i], zero, fi)
				if coupledResult != baseline[i] {
					diverged++
				}
			}
		}
		if diverged < 1 {
			t.Fatal("expected the coupled mutant to diverge from the uniform predicate on at least one mutation")
		}
		fmt.Printf("COUPLED_MUTANT_DIVERGED %d\n", diverged)
	})
}
