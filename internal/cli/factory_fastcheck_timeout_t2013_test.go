package cli

import "testing"

// card-review P2 #2: the generated `go test` step must carry an explicit
// -timeout of 30m, the repository convention, instead of Go's 10-minute default
// (a plain run of ./internal/cli/... has panicked at 601s with no failing
// assertion).
func TestT2013FastCheckStepsGoTestCarriesThirtyMinuteTimeout(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"internal/pkg/p.go": "package pkg\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"internal/pkg/p.go": "package pkg\n// edit\n"}, nil)
	steps, err := factoryFastCheckSteps(dir, "main")
	if err != nil {
		t.Fatalf("steps: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("want build, lint and test steps, got %q", steps)
	}
	test := steps[2]
	found := false
	for i, a := range test {
		if a == "-timeout" && i+1 < len(test) && test[i+1] == "30m" {
			found = true
		}
	}
	if !found {
		t.Errorf("go test step %q lacks -timeout 30m", test)
	}
}
