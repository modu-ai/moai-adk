package cli

// factory_fastcheck_unreadable_t2013_test.go — tracked item 2 (source t2021
// static review): when the card's diff against the release tip cannot be read
// (the tip names a ref that does not resolve), the fast check must widen to the
// whole module. An unreadable diff is not "no change": reading it as no change
// plans `go build ./...` alone, and the check passes without lint or test.

import (
	"reflect"
	"strings"
	"testing"
)

// unresolvableReleaseTip is a non-empty ref that names no object in the fixture.
const unresolvableReleaseTip = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

func TestT2013FastCheckPackagesUnreadableDiffWidensToModule(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"internal/pkg/p.go": "package pkg\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"internal/pkg/p.go": "package pkg\n\nvar X = 1\n"}, nil)

	// Premise: against a resolvable tip the same tree plans exactly the changed package.
	if got, want := factoryFastCheckPackages(dir, "main"), []string{"./internal/pkg"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("premise: packages against a readable tip = %q, want %q", got, want)
	}

	if got, want := factoryFastCheckPackages(dir, unresolvableReleaseTip), []string{"./..."}; !reflect.DeepEqual(got, want) {
		t.Errorf("packages against an unresolvable tip = %q, want %q (an unreadable diff widens to the module)", got, want)
	}
}

// An empty release tip (the caller could not read the release branch) leaves
// the card's change just as unknown as an unresolvable one does.
func TestT2013FastCheckPackagesEmptyTipWidensToModule(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"internal/pkg/p.go": "package pkg\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"internal/pkg/p.go": "package pkg\n\nvar X = 1\n"}, nil)

	if got, want := factoryFastCheckPackages(dir, ""), []string{"./..."}; !reflect.DeepEqual(got, want) {
		t.Errorf("packages against an empty tip = %q, want %q (an unknown change widens to the module)", got, want)
	}
}

func TestT2013FastCheckStepsUnreadableDiffPlansLintAndTest(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"internal/pkg/p.go": "package pkg\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"internal/pkg/p.go": "package pkg\n\nvar X = 1\n"}, nil)

	steps, err := factoryFastCheckSteps(dir, unresolvableReleaseTip)
	if err != nil {
		t.Fatalf("steps: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("want build, lint and test steps over the module, got %q", steps)
	}
	if lint := strings.Join(steps[1], " "); !strings.Contains(lint, "go vet") && !strings.Contains(lint, "golangci-lint") {
		t.Errorf("second step must be a lint step, got %q", steps[1])
	}
	if want := []string{"go", "test", "-timeout", "30m", "./..."}; !reflect.DeepEqual(steps[2], want) {
		t.Errorf("third step = %q, want %q", steps[2], want)
	}
}
