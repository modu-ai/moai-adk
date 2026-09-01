package guardliveness

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// countingProducer counts arrivals at the query layer — the point at which the
// refresh asks the state model for subject verdicts. AC-GDL-003 measures here
// and not at the call site: a filter moved one frame inward still satisfies a
// call-site reading, which is the mutant that survived three earlier audits.
type countingProducer struct {
	mu    sync.Mutex
	calls int
	roots []string
}

func (p *countingProducer) Produce(_ context.Context, act Activation) (Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.roots = append(p.roots, act.Root)
	return resultA(), nil
}

func (p *countingProducer) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// gitRepoWithChange builds a repository whose working tree carries an
// uncommitted modification to relPath, so a mutant reading `git diff` sees real
// content to filter on. Without a genuine diff the two fixtures of clause (c)
// would differ only in file presence, and a diff-reading mutant would look
// identical under both.
func gitRepoWithChange(t *testing.T, relPath string) string {
	t.Helper()
	root := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git %v unavailable in this environment: %v (%s)", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "fixture@example.invalid")
	run("config", "user.name", "fixture")

	path := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(path, []byte("committed\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "fixture")
	if err := os.WriteFile(path, []byte("committed\nmodified\n"), 0o644); err != nil {
		t.Fatalf("modify fixture: %v", err)
	}
	return root
}

// AC-GDL-003 — the refresh is initiated on every host activation, reaches the
// query layer on every activation, and the count is independent of what the
// activation's diff touches.
func TestRefreshReachesTheQueryLayerOnEveryActivation(t *testing.T) {
	const n = 5

	counts := map[string]int{}
	for _, tc := range []struct {
		name string
		root string
	}{
		{"diff touches no workflow file", gitRepoWithChange(t, "docs/notes.md")},
		{"diff touches a workflow file", gitRepoWithChange(t, ".github/workflows/probe.yml")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &countingProducer{}
			ev := New(p)

			initiations := 0
			for i := 0; i < n; i++ {
				refresh := ev.OnActivation(context.Background(), Activation{Root: tc.root})
				if refresh == nil {
					t.Fatalf("activation %d yielded no refresh", i+1)
				}
				initiations++
				if _, err := refresh.Wait(); err != nil {
					t.Fatalf("activation %d refresh: %v", i+1, err)
				}
			}

			// (a) every activation initiated a refresh.
			if initiations != n {
				t.Fatalf("refresh initiations = %d, want %d", initiations, n)
			}
			// (b) every refresh reached the query layer.
			if got := p.count(); got != n {
				t.Fatalf("query-layer arrivals = %d, want %d — a filter sits between the activation and the subject queries", got, n)
			}
			counts[tc.name] = p.count()
		})
	}

	// (c) the count is unchanged by the fixture's diff content.
	if len(counts) == 2 {
		var seen []int
		for _, c := range counts {
			seen = append(seen, c)
		}
		if seen[0] != seen[1] {
			t.Fatalf("query-layer arrivals differ by diff content: %v — the refresh is subject-matter conditional", counts)
		}
	}
}

// blockingProducer holds the query layer open until released, so a caller that
// awaits the refresh is observable as a caller that has not returned.
type blockingProducer struct{ release chan struct{} }

func (p *blockingProducer) Produce(_ context.Context, _ Activation) (Result, error) {
	<-p.release
	return resultA(), nil
}

// REQ-GDL-012's full criterion (AC-GDL-012) is M2's; what M1's own wiring must
// not do is hold session start open while the subject queries run. Asserted as
// a return that happens while the query layer is still blocked.
func TestOnActivationReturnsWithoutAwaitingTheRefresh(t *testing.T) {
	p := &blockingProducer{release: make(chan struct{})}
	ev := New(p)

	returned := make(chan *Refresh, 1)
	go func() { returned <- ev.OnActivation(context.Background(), Activation{Root: t.TempDir()}) }()

	var refresh *Refresh
	select {
	case refresh = <-returned:
	case <-time.After(2 * time.Second):
		close(p.release)
		t.Fatal("OnActivation did not return while the query layer was still blocked — the refresh is awaited")
	}

	close(p.release)
	if _, err := refresh.Wait(); err != nil {
		t.Fatalf("refresh: %v", err)
	}
}

// The producer is the seam to SPEC-GUARD-STATE-MODEL-001, which has not landed.
// Until it does, the wired producer must still be REACHED and must report that
// it is unwired — an early return short of the query layer would make the
// production path conditional on the seam's existence, and would report nothing
// while nothing was ever asked.
func TestUnwiredProducerIsStillReached(t *testing.T) {
	ev := New(Unwired())
	refresh := ev.OnActivation(context.Background(), Activation{Root: t.TempDir()})
	if refresh == nil {
		t.Fatal("activation yielded no refresh against the unwired producer")
	}
	if _, err := refresh.Wait(); err == nil {
		t.Fatal("unwired producer returned no error — an absent seam read as a successful evaluation")
	}
}
