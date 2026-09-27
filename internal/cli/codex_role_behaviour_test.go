// Package cli — codex_role_behaviour_test.go
//
// AC-RLP-006 for SPEC-ROLE-LOAD-PREDICATE-001: the load-predicate input type
// carries no behaviour-derived field, exhaustive per-field mutation of the
// behaviour type leaves the load verdict unchanged, and a deliberately
// coupled variant is shown to diverge on at least one such mutation.
//
// Every arm below runs against the PRODUCTION predicate and input type.
// Behaviour values are planted into the input by reflection at every site
// whose type can carry them, so a later change that adds such a site — and
// reads it — is caught here rather than passing unnoticed (card t1260: the
// earlier form built behaviour values it never handed to anything, and a
// mutant that added one and read it passed every test).
package cli

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// codexRoleLoadCase is one predicate input: a role and the matched set
// derived for it.
type codexRoleLoadCase struct {
	role    string
	matched map[string]bool
}

// codexRoleLoadInputCoupled is a deliberately WRONG input shape that can
// carry a behaviour value. It exists only in this test file, as the positive
// control proving the divergence harness detects a coupled predicate.
type codexRoleLoadInputCoupled struct {
	Role    string
	Matched map[string]bool
	Extra   codexRoleBehaviour
}

// codexRoleBehaviourMutations returns, by reflection over codexRoleBehaviour,
// one value per field with exactly that field set to a non-zero value. A
// field kind it cannot flip fails the test, so the returned count always
// equals the field count.
func codexRoleBehaviourMutations(t *testing.T) []codexRoleBehaviour {
	t.Helper()
	bt := reflect.TypeOf(codexRoleBehaviour{})
	out := make([]codexRoleBehaviour, 0, bt.NumField())
	for fi := 0; fi < bt.NumField(); fi++ {
		b := reflect.New(bt).Elem()
		f := b.Field(fi)
		switch f.Kind() {
		case reflect.Bool:
			f.SetBool(true)
		default:
			t.Fatalf("codexRoleBehaviour field %q has kind %s; extend codexRoleBehaviourMutations so the mutation stays exhaustive",
				bt.Field(fi).Name, f.Kind())
		}
		out = append(out, b.Interface().(codexRoleBehaviour))
	}
	return out
}

// codexRoleInjectBehaviour writes b into every field of the struct value v —
// recursively through nested structs and non-nil struct pointers — whose type
// is codexRoleBehaviour or *codexRoleBehaviour, and returns how many such
// sites it found.
func codexRoleInjectBehaviour(v reflect.Value, b codexRoleBehaviour) int {
	bt := reflect.TypeOf(codexRoleBehaviour{})
	sites := 0
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch {
		case f.Type() == bt:
			f.Set(reflect.ValueOf(b))
			sites++
		case f.Type() == reflect.PointerTo(bt):
			nb := b
			f.Set(reflect.ValueOf(&nb))
			sites++
		case f.Kind() == reflect.Struct:
			sites += codexRoleInjectBehaviour(f, b)
		}
	}
	return sites
}

// codexRoleBehaviourDivergence evaluates predicate on every case twice — once
// with no behaviour value present, once with behaviourFor(i) planted at every
// site the input type offers — and counts the cases whose verdict changed.
// inputType must carry Role (string) and Matched (map[string]bool) fields.
func codexRoleBehaviourDivergence(t *testing.T, inputType reflect.Type, predicate func(reflect.Value) bool,
	cases []codexRoleLoadCase, behaviourFor func(i int) codexRoleBehaviour) (sites, diverged int) {
	t.Helper()
	build := func(c codexRoleLoadCase) reflect.Value {
		v := reflect.New(inputType).Elem()
		role, matched := v.FieldByName("Role"), v.FieldByName("Matched")
		if !role.IsValid() || !matched.IsValid() {
			t.Fatalf("%s must carry Role and Matched fields", inputType)
		}
		role.SetString(c.role)
		matched.Set(reflect.ValueOf(c.matched))
		return v
	}
	for i, c := range cases {
		base := predicate(build(c))
		v := build(c)
		sites = codexRoleInjectBehaviour(v, behaviourFor(i))
		if predicate(v) != base {
			diverged++
		}
	}
	return sites, diverged
}

// TestCodexRoleLoadPredicateStructurallyIndependent — AC-RLP-006
// (REQ-RLP-010, REQ-RLP-011).
func TestCodexRoleLoadPredicateStructurallyIndependent(t *testing.T) {
	root := repoRoot(t)
	realDir := filepath.Join(root, codexRoleFixtureRoot, "real")
	rolesDir := filepath.Join(root, codexRoleFixtureRoot, "roles")
	otherVersionDir := filepath.Join(root, codexRoleFixtureRoot, "roles-other-version")

	table, reason, err := codexBuildRoleExpectationTable(rolesDir)
	if err != nil {
		t.Fatalf("build table: %v", err)
	}
	if reason != "" {
		t.Fatalf("table ineligible: %s", reason)
	}
	otherTable, reason, err := codexBuildRoleExpectationTable(otherVersionDir)
	if err != nil {
		t.Fatalf("build other-version table: %v", err)
	}
	if reason != "" {
		t.Fatalf("other-version table ineligible: %s", reason)
	}

	// Two verdict populations over the same 12 labeled sessions: the real
	// table (load true) and the other-version table (load false). Both are
	// required — a coupling that can only turn false into true (OR) is
	// invisible on an all-true population, and one that can only turn true
	// into false (AND) is invisible on an all-false one.
	var cases []codexRoleLoadCase
	labeled := 0
	for _, p := range globSorted(t, realDir, "*.jsonl") {
		role, hasLabel, err := codexRoleSessionLabel(p)
		if err != nil {
			t.Fatalf("session label %s: %v", p, err)
		}
		if !hasLabel {
			continue
		}
		labeled++
		for _, tb := range []codexRoleExpectationTable{table, otherTable} {
			matched, err := codexRoleFingerprintDerive(p, tb)
			if err != nil {
				t.Fatalf("derive %s: %v", p, err)
			}
			cases = append(cases, codexRoleLoadCase{role: role, matched: matched})
		}
	}
	if labeled != 12 {
		t.Fatalf("expected 12 labeled real sessions, got %d", labeled)
	}
	trueCount := 0
	for _, c := range cases {
		if codexRoleLoadPredicate(codexRoleLoadInput{Role: c.role, Matched: c.matched}) {
			trueCount++
		}
	}
	if trueCount != 12 || len(cases)-trueCount != 12 {
		t.Fatalf("expected 12 load-true and 12 load-false cases, got %d true of %d", trueCount, len(cases))
	}

	realPredicate := func(v reflect.Value) bool {
		return codexRoleLoadPredicate(v.Interface().(codexRoleLoadInput))
	}
	inputType := reflect.TypeOf(codexRoleLoadInput{})
	mutations := codexRoleBehaviourMutations(t)
	numBehaviourFields := reflect.TypeOf(codexRoleBehaviour{}).NumField()
	if numBehaviourFields == 0 {
		t.Fatal("codexRoleBehaviour must carry at least one field for this AC to be meaningful")
	}

	t.Run("input_type_has_no_behaviour_field", func(t *testing.T) {
		// Exact allowlist: the input type is these two fields and nothing
		// else. A field under any name is rejected, not only one whose name
		// mentions behaviour or nonce.
		want := []struct {
			name string
			typ  reflect.Type
		}{
			{"Role", reflect.TypeOf("")},
			{"Matched", reflect.TypeOf(map[string]bool{})},
		}
		if inputType.NumField() != len(want) {
			names := make([]string, 0, inputType.NumField())
			for i := 0; i < inputType.NumField(); i++ {
				names = append(names, inputType.Field(i).Name+" "+inputType.Field(i).Type.String())
			}
			t.Fatalf("codexRoleLoadInput must carry exactly {Role string, Matched map[string]bool}; got %v", names)
		}
		for i, w := range want {
			f := inputType.Field(i)
			if f.Name != w.name || f.Type != w.typ {
				t.Fatalf("codexRoleLoadInput field %d = %s %s, want %s %s", i, f.Name, f.Type, w.name, w.typ)
			}
			name := strings.ToLower(f.Name)
			if strings.Contains(name, "behaviour") || strings.Contains(name, "behavior") || strings.Contains(name, "nonce") {
				t.Fatalf("codexRoleLoadInput field %q must not reference behaviour/nonce", f.Name)
			}
		}
		if sites := codexRoleInjectBehaviour(reflect.New(inputType).Elem(), codexRoleBehaviour{}); sites != 0 {
			t.Fatalf("codexRoleLoadInput offers %d site(s) that can carry a codexRoleBehaviour value", sites)
		}
	})

	t.Run("exhaustive_field_mutations", func(t *testing.T) {
		mutated := 0
		for fi, b := range mutations {
			sites, diverged := codexRoleBehaviourDivergence(t, inputType, realPredicate, cases,
				func(int) codexRoleBehaviour { return b })
			if sites != 0 {
				t.Fatalf("field %d mutation found %d behaviour site(s) in codexRoleLoadInput", fi, sites)
			}
			if diverged != 0 {
				t.Fatalf("field %d mutation changed the load verdict on %d of %d cases", fi, diverged, len(cases))
			}
			mutated++
		}
		if mutated != numBehaviourFields {
			t.Fatalf("BEHAVIOUR_FIELDS_MUTATED %d OF %d", mutated, numBehaviourFields)
		}
		fmt.Printf("BEHAVIOUR_FIELDS_MUTATED %d OF %d\n", mutated, numBehaviourFields)
	})

	t.Run("half_true_arm", func(t *testing.T) {
		// Half the sessions carry every behaviour field true, half carry all
		// false, planted through the same injection the mutations use. cases
		// holds each session twice (load-true, then load-false), so the split
		// is by session (i/2), giving both verdict populations a true half.
		allTrue := codexRoleBehaviour{}
		for _, b := range mutations {
			bv, av := reflect.ValueOf(b), reflect.ValueOf(&allTrue).Elem()
			for i := 0; i < bv.NumField(); i++ {
				if bv.Field(i).Bool() {
					av.Field(i).SetBool(true)
				}
			}
		}
		_, diverged := codexRoleBehaviourDivergence(t, inputType, realPredicate, cases,
			func(i int) codexRoleBehaviour {
				if (i/2)%2 == 0 {
					return allTrue
				}
				return codexRoleBehaviour{}
			})
		if diverged != 0 {
			t.Fatalf("half-true arm changed the load verdict on %d of %d cases", diverged, len(cases))
		}
	})

	t.Run("coupled_mutant_rejected", func(t *testing.T) {
		// Positive control: the SAME harness, pointed at an input type that
		// can carry a behaviour value and a predicate reading field fi of it,
		// must diverge — for every behaviour field, not merely one.
		coupledType := reflect.TypeOf(codexRoleLoadInputCoupled{})
		total := 0
		for fi, b := range mutations {
			coupled := func(v reflect.Value) bool {
				in := v.Interface().(codexRoleLoadInputCoupled)
				return in.Matched[in.Role] || reflect.ValueOf(in.Extra).Field(fi).Bool()
			}
			sites, diverged := codexRoleBehaviourDivergence(t, coupledType, coupled, cases,
				func(int) codexRoleBehaviour { return b })
			if sites != 1 {
				t.Fatalf("coupled input type: expected 1 behaviour site, found %d", sites)
			}
			if diverged < 1 {
				t.Fatalf("coupled mutant reading behaviour field %d did not diverge on any case", fi)
			}
			total += diverged
		}
		fmt.Printf("COUPLED_MUTANT_DIVERGED %d\n", total)
	})
}
