package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// fakeProcess is one node of a synthetic process tree.
type fakeProcess struct {
	ppid int
	comm string
}

// withFakeAncestry installs a synthetic process tree plus an explicit live-PID
// set for the duration of the test, restoring both package seams afterwards.
func withFakeAncestry(t *testing.T, tree map[int]fakeProcess, live map[int]bool) {
	t.Helper()
	origInfo, origAlive := procInfo, pidIsAlive
	t.Cleanup(func() { procInfo, pidIsAlive = origInfo, origAlive })

	procInfo = func(pid int) (int, string, bool) {
		p, ok := tree[pid]
		if !ok {
			return 0, "", false
		}
		return p.ppid, p.comm, true
	}
	pidIsAlive = func(pid int) bool { return live[pid] }
}

// TestAncestorSessionPID_SkipsWrapperShells covers the real hook chain:
// the session spawns a wrapper shell, which execs the moai binary. The PID
// worth recording is the session's, never the moai subprocess's own.
func TestAncestorSessionPID_SkipsWrapperShells(t *testing.T) {
	const (
		sessionPID = 4100
		shellPID   = 4200
		hookPID    = 4300
	)
	tree := map[int]fakeProcess{
		hookPID:    {ppid: shellPID, comm: "moai"},
		shellPID:   {ppid: sessionPID, comm: "bash"},
		sessionPID: {ppid: 500, comm: "claude"},
		500:        {ppid: 1, comm: "zsh"},
	}
	withFakeAncestry(t, tree, map[int]bool{sessionPID: true, shellPID: true, hookPID: true, 500: true})

	if got := ancestorSessionPID(hookPID); got != sessionPID {
		t.Errorf("ancestorSessionPID = %d, want the session PID %d", got, sessionPID)
	}
}

// TestAncestorSessionPID_CollapsedChain covers the case where `sh -c` and the
// wrapper's own `exec` collapse the shell away, leaving moai as a direct child
// of the session.
func TestAncestorSessionPID_CollapsedChain(t *testing.T) {
	const (
		sessionPID = 7100
		hookPID    = 7200
	)
	tree := map[int]fakeProcess{
		hookPID:    {ppid: sessionPID, comm: "moai"},
		sessionPID: {ppid: 900, comm: "2.1.235"}, // the binary is version-named
		900:        {ppid: 1, comm: "zsh"},
	}
	withFakeAncestry(t, tree, map[int]bool{sessionPID: true, hookPID: true, 900: true})

	if got := ancestorSessionPID(hookPID); got != sessionPID {
		t.Errorf("ancestorSessionPID = %d, want %d", got, sessionPID)
	}
}

// TestAncestorSessionPID_Unresolvable enumerates every shape that must yield 0
// so the caller falls back rather than recording a wrong PID.
func TestAncestorSessionPID_Unresolvable(t *testing.T) {
	cases := []struct {
		name  string
		tree  map[int]fakeProcess
		live  map[int]bool
		start int
	}{
		{
			name:  "platform reports nothing",
			tree:  map[int]fakeProcess{},
			live:  map[int]bool{},
			start: 10,
		},
		{
			name:  "walk reaches init",
			tree:  map[int]fakeProcess{10: {ppid: 1, comm: "moai"}},
			live:  map[int]bool{10: true},
			start: 10,
		},
		{
			name: "resolved ancestor is dead",
			tree: map[int]fakeProcess{
				10: {ppid: 20, comm: "moai"},
				20: {ppid: 30, comm: "claude"},
			},
			live:  map[int]bool{10: true},
			start: 10,
		},
		{
			name: "shells all the way up past the depth bound",
			tree: func() map[int]fakeProcess {
				tree := map[int]fakeProcess{}
				for i := 0; i < 40; i++ {
					tree[100+i] = fakeProcess{ppid: 101 + i, comm: "sh"}
				}
				return tree
			}(),
			live:  map[int]bool{},
			start: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withFakeAncestry(t, tc.tree, tc.live)
			if got := ancestorSessionPID(tc.start); got != 0 {
				t.Errorf("ancestorSessionPID = %d, want 0 (unresolvable)", got)
			}
		})
	}
}

// TestSessionPIDFromEnv_RejectsUnusableValues asserts the override is honored
// only when it names a live process — a stale PID inherited through the
// environment must not be recorded.
func TestSessionPIDFromEnv_RejectsUnusableValues(t *testing.T) {
	withFakeAncestry(t, nil, map[int]bool{5150: true})

	if pid, ok := sessionPIDFromEnv(" 5150 "); !ok || pid != 5150 {
		t.Errorf("sessionPIDFromEnv(live) = (%d, %v), want (5150, true)", pid, ok)
	}
	for _, raw := range []string{"", "   ", "not-a-number", "0", "-3", "6000" /* dead */} {
		if pid, ok := sessionPIDFromEnv(raw); ok {
			t.Errorf("sessionPIDFromEnv(%q) = (%d, true), want rejected", raw, pid)
		}
	}
}

// TestResolveSessionPID_PrefersEnvOverride pins the resolution order.
//
// The fixture makes the override an ACTUAL ancestor of this process (card
// t958): self -> 9999 (claude) -> 8250. The walk stops at 9999, the first
// non-wrapper ancestor, so the two answers still differ and the assertion
// still measures "an override for THIS session wins over the walk" — which is
// what it was written to measure. Before t958 the override here named a pid
// nowhere in the chain, i.e. the defect scenario itself.
func TestResolveSessionPID_PrefersEnvOverride(t *testing.T) {
	const override = 8250
	withFakeAncestry(t,
		map[int]fakeProcess{
			os.Getpid(): {ppid: 9999, comm: "moai"},
			9999:        {ppid: override, comm: "claude"},
			override:    {ppid: 1, comm: "login"},
		},
		map[int]bool{override: true, 9999: true},
	)
	t.Setenv(config.EnvMoaiSessionPID, "8250")

	if got := resolveSessionPID(); got != override {
		t.Errorf("resolveSessionPID = %d, want the env override %d", got, override)
	}
}

// TestResolveOwnerPID_Precedence pins the exported seam's three outcomes
// (card t298): the env stamp wins, the ancestry walk answers next, and an
// unresolvable owner is reported AS unresolvable rather than papered over with
// os.Getpid(). The third row is the one that matters — a caller whose record
// outlives its own process needs to know it did not find an owner, because
// recording this process's pid there is the integration-lock defect.
func TestResolveOwnerPID_Precedence(t *testing.T) {
	// Card t958 split what was one "env stamp wins" row into two, because the
	// original fixture's override (8250) named a pid that was NOT in this
	// process's ancestry — the defect scenario. Fixing the fixture rather
	// than the assertion keeps the row measuring what it was written to
	// measure; the second row pins the case the first one used to cover by
	// accident, so the precedence table states BOTH outcomes explicitly
	// instead of leaving the new one to a distant test.
	t.Run("env stamp wins when it names an ancestor", func(t *testing.T) {
		const override = 8250
		const near = 9999
		withFakeAncestry(t,
			map[int]fakeProcess{
				os.Getpid(): {ppid: near, comm: "moai"},
				near:        {ppid: override, comm: "claude"},
				override:    {ppid: 1, comm: "login"},
			},
			map[int]bool{override: true, near: true},
		)
		t.Setenv(config.EnvMoaiSessionPID, "8250")

		pid, ok := ResolveOwnerPID()
		if !ok || pid != override {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, override)
		}
		// The walk's own answer differs, so the row is not vacuous.
		if got := ancestorSessionPID(os.Getpid()); got == override {
			t.Fatalf("fixture is vacuous: the walk alone already answers %d", got)
		}
	})

	t.Run("a stamp naming no ancestor is ignored", func(t *testing.T) {
		const foreign = 8250
		const owner = 9999
		withFakeAncestry(t,
			map[int]fakeProcess{
				os.Getpid(): {ppid: owner, comm: "moai"},
				owner:       {ppid: 1, comm: "claude"},
			},
			map[int]bool{foreign: true, owner: true},
		)
		t.Setenv(config.EnvMoaiSessionPID, "8250")

		pid, ok := ResolveOwnerPID()
		if !ok || pid != owner {
			t.Errorf("ResolveOwnerPID = (%d, %v), want the walk's answer (%d, true): a live stamp naming neither this process nor any ancestor is one this process inherited from an unrelated session", pid, ok, owner)
		}
	})

	t.Run("a stamp naming this process itself wins", func(t *testing.T) {
		self := os.Getpid()
		withFakeAncestry(t,
			map[int]fakeProcess{self: {ppid: 9999, comm: "moai"}, 9999: {ppid: 1, comm: "claude"}},
			map[int]bool{self: true, 9999: true},
		)
		t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(self))

		pid, ok := ResolveOwnerPID()
		if !ok || pid != self {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, self)
		}
	})

	t.Run("a stamp is accepted where ancestry is unmeasurable", func(t *testing.T) {
		const foreign = 8250
		withFakeAncestry(t, map[int]fakeProcess{}, map[int]bool{foreign: true})
		t.Setenv(config.EnvMoaiSessionPID, "8250")

		pid, ok := ResolveOwnerPID()
		if !ok || pid != foreign {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (%d, true): where the platform reports no ancestry at all, the stamp is the only signal there is and the fail-open keeps today's behavior", pid, ok, foreign)
		}
	})

	t.Run("ancestry answers when no stamp", func(t *testing.T) {
		const owner = 9999
		withFakeAncestry(t,
			map[int]fakeProcess{os.Getpid(): {ppid: owner, comm: "moai"}, owner: {ppid: 1, comm: "claude"}},
			map[int]bool{owner: true},
		)
		t.Setenv(config.EnvMoaiSessionPID, "")

		pid, ok := ResolveOwnerPID()
		if !ok || pid != owner {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, owner)
		}
	})

	t.Run("unresolvable reports zero, never os.Getpid()", func(t *testing.T) {
		withFakeAncestry(t, map[int]fakeProcess{}, map[int]bool{})
		t.Setenv(config.EnvMoaiSessionPID, "")

		pid, ok := ResolveOwnerPID()
		if ok || pid != 0 {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (0, false)", pid, ok)
		}
		if pid == os.Getpid() {
			t.Error("ResolveOwnerPID fell back to os.Getpid(); that fallback belongs to the registry alone")
		}
	})
}

// TestResolveSessionPID_FallsBackToSelf covers the unsupported-platform path:
// with no override and no readable ancestry, the resolver keeps the
// pre-existing behavior rather than recording nothing.
func TestResolveSessionPID_FallsBackToSelf(t *testing.T) {
	withFakeAncestry(t, map[int]fakeProcess{}, map[int]bool{})
	t.Setenv(config.EnvMoaiSessionPID, "")

	if got := resolveSessionPID(); got != os.Getpid() {
		t.Errorf("resolveSessionPID = %d, want os.Getpid() %d", got, os.Getpid())
	}
}

// TestRegister_RecordsLivePID is the regression assertion the defect report
// asks for: the PID written at registration must answer a liveness probe
// immediately afterwards. Before the fix the registry recorded the hook
// subprocess's PID, which was dead on arrival for every reader.
func TestRegister_RecordsLivePID(t *testing.T) {
	const sessionPID = 3300
	withFakeAncestry(t,
		map[int]fakeProcess{
			os.Getpid(): {ppid: 3400, comm: "moai"},
			3400:        {ppid: sessionPID, comm: "bash"},
			sessionPID:  {ppid: 1, comm: "claude"},
		},
		map[int]bool{os.Getpid(): true, 3400: true, sessionPID: true},
	)

	path := filepath.Join(t.TempDir(), "active-sessions.json")
	reg := NewRegistry(path, nil)
	if err := reg.Register("11111111-2222-3333-4444-555555555555", SpecIDNone, PhaseNone); err != nil {
		t.Fatalf("Register: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entries []Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("unmarshal registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("registry holds %d entries, want 1", len(entries))
	}
	if entries[0].PID != sessionPID {
		t.Errorf("recorded PID = %d, want the session PID %d", entries[0].PID, sessionPID)
	}
	if !pidIsAlive(entries[0].PID) {
		t.Errorf("recorded PID %d is not alive immediately after registration", entries[0].PID)
	}
}

// TestResolveOwnerPID_InheritedStampCollapsesDistinctSessions is the
// reproduction for card t958: two independent sessions that INHERIT the same
// MOAI_SESSION_PID value resolve to the SAME owner pid, because the override
// is honored unconditionally — it is never checked against the resolving
// process's own ancestry.
//
// Attribution: PIDSourceSessionOwner (acquire recording the owning-session
// pid) predates this, from commit 3f3465369 (card t298). Commit ea939d9a7
// (card t951) added releasableBy, making pid a SECOND key on the release
// holder judgment, which WIDENED the exposure surface of this pre-existing
// acquire-era property. It is not a t951 regression.
//
// Each sub-test stands for one session's resolution: the two arms install
// DIFFERENT synthetic ancestries, so any collapse of the two answers is the
// override erasing them.
func TestResolveOwnerPID_InheritedStampCollapsesDistinctSessions(t *testing.T) {
	const (
		ownerA    = 9001
		ownerB    = 9002
		inherited = 7000
	)

	// treeFor builds the ancestry of a session whose own non-wrapper ancestor
	// is owner: this process -> (moai) -> owner (claude).
	treeFor := func(owner int) map[int]fakeProcess {
		return map[int]fakeProcess{
			os.Getpid(): {ppid: owner, comm: "moai"},
			owner:       {ppid: 1, comm: "claude"},
		}
	}

	// Positive control — WITHOUT the inherited stamp the two sessions resolve
	// differently. Without this arm "inheritance is the cause" is not
	// established: the experimental arm alone cannot tell an override-induced
	// collapse from two ancestries that were never distinguishable.
	t.Run("control: no inherited stamp, distinct ancestries stay distinct", func(t *testing.T) {
		if ownerA == ownerB {
			t.Fatal("control is vacuous: the two arms name the same owner pid")
		}
		t.Setenv(config.EnvMoaiSessionPID, "")

		// Merely OMITTING to set the variable does not remove one inherited
		// from the environment. An unasserted control silently becomes a copy
		// of the experimental arm and produces a plausible-looking pass, so
		// assert the override is genuinely not in effect.
		if pid, ok := sessionPIDFromEnv(os.Getenv(config.EnvMoaiSessionPID)); ok {
			t.Fatalf("control arm is contaminated: an override IS in effect (pid %d)", pid)
		}

		var gotA, gotB int
		t.Run("session A", func(t *testing.T) {
			withFakeAncestry(t, treeFor(ownerA), map[int]bool{ownerA: true, inherited: true})
			pid, ok := ResolveOwnerPID()
			if !ok || pid != ownerA {
				t.Fatalf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, ownerA)
			}
			gotA = pid
		})
		t.Run("session B", func(t *testing.T) {
			withFakeAncestry(t, treeFor(ownerB), map[int]bool{ownerB: true, inherited: true})
			pid, ok := ResolveOwnerPID()
			if !ok || pid != ownerB {
				t.Fatalf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, ownerB)
			}
			gotB = pid
		})
		if gotA == gotB {
			t.Fatalf("two distinct sessions both resolved %d; the control cannot separate them", gotA)
		}
	})

	// Experimental — WITH the same inherited stamp the two distinct
	// ancestries are erased and both sessions resolve to the same pid.
	t.Run("inherited stamp erases both ancestries", func(t *testing.T) {
		t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(inherited))

		var gotA, gotB int
		t.Run("session A", func(t *testing.T) {
			withFakeAncestry(t, treeFor(ownerA), map[int]bool{ownerA: true, inherited: true})
			pid, ok := ResolveOwnerPID()
			if !ok {
				t.Fatalf("ResolveOwnerPID = (%d, false), want a resolved owner", pid)
			}
			gotA = pid
			if pid != ownerA {
				t.Errorf("session A resolved %d, want its OWN ancestor %d — an inherited stamp naming neither this process nor any of its ancestors must not win", pid, ownerA)
			}
		})
		t.Run("session B", func(t *testing.T) {
			withFakeAncestry(t, treeFor(ownerB), map[int]bool{ownerB: true, inherited: true})
			pid, ok := ResolveOwnerPID()
			if !ok {
				t.Fatalf("ResolveOwnerPID = (%d, false), want a resolved owner", pid)
			}
			gotB = pid
			if pid != ownerB {
				t.Errorf("session B resolved %d, want its OWN ancestor %d", pid, ownerB)
			}
		})
		if gotA == gotB {
			t.Errorf("both sessions resolved the same owner pid %d: two independent sessions that inherited MOAI_SESSION_PID are indistinguishable to every pid-keyed ownership check", gotA)
		}
	})
}
