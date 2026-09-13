package receipt

import (
	"errors"
	"testing"
)

// chainedManifest publishes parent -> c1 -> c2 -> c3 in order and returns the
// manifest plus the chain.
func chainedManifest(t *testing.T) (*Manifest, []Candidate) {
	t.Helper()
	m, err := New(testUUID)
	if err != nil {
		t.Fatal(err)
	}
	previous := Hash([]byte("root"))
	chain := make([]Candidate, 0, 3)
	for _, label := range []string{"c1", "c2", "c3"} {
		c := candidate(label, "ignored", "o")
		c.Previous = previous
		if err = m.Publish(c); err != nil {
			t.Fatal(err)
		}
		chain = append(chain, c)
		previous = c.Prefix
	}
	return m, chain
}

func TestForkAtCopiesOnlyTheBoundaryChain(t *testing.T) {
	m, chain := chainedManifest(t)
	const child = "00000000-0000-4000-8000-000000000001"
	fork, err := m.ForkAt(child, chain[1].Prefix)
	if err != nil {
		t.Fatal(err)
	}
	got := fork.Candidates()
	if len(got) != 2 {
		t.Fatal("boundary fork candidate count", len(got))
	}
	if got[0] != chain[0] || got[1] != chain[1] {
		t.Fatal("boundary fork carried the wrong chain", got)
	}
	// The parent may keep completing turns; none of that may reach the child.
	post := Observation{Prefix: chain[2].Prefix, Previous: chain[1].Prefix, Provider: chain[2].Provider, Opaque: chain[2].Opaque, Items: chain[2].Items}
	if err = fork.Check(child, []Observation{post}); !errors.Is(err, ErrInvalid) {
		t.Fatal("post-boundary fact validated in the child", err)
	}
	pre := Observation{Prefix: chain[1].Prefix, Previous: chain[0].Prefix, Provider: chain[1].Provider, Opaque: chain[1].Opaque, Items: chain[1].Items}
	if err = fork.Check(child, []Observation{pre}); err != nil {
		t.Fatal("pre-boundary fact lost", err)
	}
}

func TestForkAtRejectionGroups(t *testing.T) {
	m, chain := chainedManifest(t)
	const child = "00000000-0000-4000-8000-000000000002"
	if _, err := m.ForkAt(child, Hash([]byte("future-turn"))); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown boundary accepted", err)
	}
	if _, err := m.ForkAt(child, Digest{}); !errors.Is(err, ErrInvalid) {
		t.Fatal("zero boundary accepted", err)
	}
	broken, _ := New(testUUID)
	brokenC := candidate("x", "ignored", "o")
	brokenC.Previous = Hash([]byte("nowhere"))
	_ = broken.Publish(brokenC)
	// A dangling Previous link is a chain root, not a tamper signal: the fork
	// keeps exactly the boundary candidate itself.
	dangling, err := broken.ForkAt(child, brokenC.Prefix)
	if err != nil || len(dangling.Candidates()) != 1 {
		t.Fatal("dangling chain root rejected", err, dangling)
	}
	looped, _ := New(testUUID)
	x := candidate("x", "ignored", "o")
	x.Previous = Hash([]byte("y"))
	_ = looped.Publish(x)
	y := candidate("y", "ignored", "o")
	y.Previous = x.Prefix
	_ = looped.Publish(y)
	if _, err := looped.ForkAt(child, x.Prefix); !errors.Is(err, ErrInvalid) {
		t.Fatal("cyclic chain accepted", err)
	}
	if _, err := m.ForkAt("not-a-uuid", chain[0].Prefix); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid child uuid accepted", err)
	}
}

func TestForkAtTipMatchesFork(t *testing.T) {
	m, chain := chainedManifest(t)
	tipFork, err := m.Fork("00000000-0000-4000-8000-000000000003")
	if err != nil {
		t.Fatal(err)
	}
	atFork, err := m.ForkAt("00000000-0000-4000-8000-000000000004", chain[2].Prefix)
	if err != nil {
		t.Fatal(err)
	}
	a, b := tipFork.Candidates(), atFork.Candidates()
	if len(a) != len(b) {
		t.Fatal(len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("tip fork diverged from the boundary chain", a[i], b[i])
		}
	}
}

func TestChainToKeepsAlternativeEvidenceAtTheBoundary(t *testing.T) {
	m, chain := chainedManifest(t)
	alt := chain[1]
	alt.Opaque = Hash([]byte("alternative"))
	if err := m.Publish(alt); err != nil {
		t.Fatal(err)
	}
	fork, err := m.ForkAt("00000000-0000-4000-8000-000000000005", chain[1].Prefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(fork.Candidates()) != 3 {
		t.Fatal("boundary evidence alternative dropped", fork.Candidates())
	}
}
