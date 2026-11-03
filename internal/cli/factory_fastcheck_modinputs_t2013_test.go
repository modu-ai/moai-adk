package cli

import (
	"reflect"
	"strings"
	"testing"
)

// A card whose change is only a module input (go.mod or go.sum) owns no Go
// package, yet it can break any package's build or tests. The fast required
// check must still plan lint and test, over the whole module, instead of
// stopping at `go build ./...`.
func TestT2013FastCheckStepsModuleInputOnlyChangePlansLintAndTest(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"go.sum":            "a v1 h1:x\n",
		"internal/pkg/p.go": "package pkg\n",
		"docs/guide.md":     "g\n",
	}
	cases := []struct {
		name   string
		writes map[string]string
	}{
		{"go.mod only", map[string]string{"go.mod": "module example.com/m\n\ngo 1.23\n"}},
		{"go.sum only", map[string]string{"go.sum": "a v1 h1:x\nb v1 h1:y\n"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := fastCheckRepo(t, base, tc.writes, nil)
			steps, err := factoryFastCheckSteps(dir, "main")
			if err != nil {
				t.Fatalf("steps: %v", err)
			}
			if len(steps) != 3 {
				t.Fatalf("want build, lint and test steps, got %q", steps)
			}
			lint := strings.Join(steps[1], " ")
			if !strings.Contains(lint, "go vet") && !strings.Contains(lint, "golangci-lint") {
				t.Errorf("second step must be a lint step, got %q", steps[1])
			}
			if want := []string{"go", "test", "./..."}; !reflect.DeepEqual(steps[2], want) {
				t.Errorf("third step = %q, want %q", steps[2], want)
			}
		})
	}
}

// A change that touches neither a Go file nor a module input still plans only
// the build, as before.
func TestT2013FastCheckStepsDocOnlyChangeKeepsBuildOnly(t *testing.T) {
	base := map[string]string{
		"go.mod":            "module example.com/m\n",
		"internal/pkg/p.go": "package pkg\n",
		"docs/guide.md":     "g\n",
	}
	dir := fastCheckRepo(t, base, map[string]string{"docs/guide.md": "e\n"}, nil)
	steps, err := factoryFastCheckSteps(dir, "main")
	if err != nil {
		t.Fatalf("steps: %v", err)
	}
	if want := [][]string{{"go", "build", "./..."}}; !reflect.DeepEqual(steps, want) {
		t.Errorf("steps = %q, want %q", steps, want)
	}
}
