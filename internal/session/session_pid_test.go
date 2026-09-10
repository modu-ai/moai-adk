package session

import (
	"encoding/json"
	"os"
	"path/filepath"
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
func TestResolveSessionPID_PrefersEnvOverride(t *testing.T) {
	const override = 8250
	withFakeAncestry(t,
		map[int]fakeProcess{os.Getpid(): {ppid: 9999, comm: "moai"}, 9999: {ppid: 1, comm: "claude"}},
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
	t.Run("env stamp wins", func(t *testing.T) {
		const override = 8250
		withFakeAncestry(t,
			map[int]fakeProcess{os.Getpid(): {ppid: 9999, comm: "moai"}, 9999: {ppid: 1, comm: "claude"}},
			map[int]bool{override: true, 9999: true},
		)
		t.Setenv(config.EnvMoaiSessionPID, "8250")

		pid, ok := ResolveOwnerPID()
		if !ok || pid != override {
			t.Errorf("ResolveOwnerPID = (%d, %v), want (%d, true)", pid, ok, override)
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
