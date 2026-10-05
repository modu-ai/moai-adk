package hook

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// countingGuardProducer counts arrivals at the guard-liveness query layer, as
// reached through the real session-start handler rather than through the
// evaluator directly. This is where AC-GDL-003's mutants would actually live:
// a filter added to the handler, or to the refresh it calls.
type countingGuardProducer struct {
	mu    sync.Mutex
	calls int
}

func (p *countingGuardProducer) Produce(_ context.Context, _ guardliveness.Activation) (guardliveness.Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return guardliveness.Result{Clean: &guardliveness.Designation{Values: []string{"alpha"}}}, nil
}

func (p *countingGuardProducer) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// stubGuardLivenessProducer installs a counting producer in the ONE seam the
// session-start invocation reads, and restores the real one afterwards.
func stubGuardLivenessProducer(t *testing.T) *countingGuardProducer {
	t.Helper()
	orig := guardLivenessProducer
	t.Cleanup(func() { guardLivenessProducer = orig })
	p := &countingGuardProducer{}
	guardLivenessProducer = p
	return p
}

// projectRootTouchingWorkflows returns a throwaway project directory that does
// (or does not) carry workflow files, so the two fixtures of AC-GDL-003 clause
// (c) differ in exactly the subject matter a filter would test for.
func projectRootTouchingWorkflows(t *testing.T, withWorkflows bool) string {
	t.Helper()
	root := t.TempDir()
	rel := filepath.Join("docs", "notes.md")
	if withWorkflows {
		rel = filepath.Join(".github", "workflows", "probe.yml")
	}
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(path, []byte("fixture\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return root
}

// AC-GDL-003 — N host activations produce N refreshes that reach the query
// layer, and the count does not move with the activation's subject matter.
//
// TestMain sets deferredScansAsync=false for this binary, so the invocation
// runs inline here and the count is read after the handler returns rather than
// raced against a goroutine.
func TestSessionStart_GuardLivenessRefreshOnEveryActivation(t *testing.T) {
	const n = 5

	counts := map[string]int{}
	for _, tc := range []struct {
		name          string
		withWorkflows bool
	}{
		{"activation touches no workflow file", false},
		{"activation touches workflow files", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := stubGuardLivenessProducer(t)
			root := projectRootTouchingWorkflows(t, tc.withWorkflows)
			h := NewSessionStartHandler(nil)

			for i := 0; i < n; i++ {
				if _, err := h.Handle(context.Background(), &HookInput{
					SessionID:     "sess-guard-liveness",
					CWD:           root,
					ProjectDir:    root,
					HookEventName: "SessionStart",
				}); err != nil {
					t.Fatalf("Handle %d: %v", i+1, err)
				}
			}

			if got := p.count(); got != n {
				t.Fatalf("query-layer arrivals = %d, want %d — the session-start invocation is conditional", got, n)
			}
			counts[tc.name] = p.count()
		})
	}

	if len(counts) == 2 {
		var seen []int
		for _, c := range counts {
			seen = append(seen, c)
		}
		if seen[0] != seen[1] {
			t.Fatalf("query-layer arrivals differ by subject matter: %v", counts)
		}
	}
}
