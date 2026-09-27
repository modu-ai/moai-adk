package session

import (
	"strconv"
	"testing"
)

// syntheticView builds a ProcessView over a synthetic process tree in which
// every listed process is alive.
func syntheticView(self int, tree map[int]fakeProcess) ProcessView {
	return ProcessView{
		Self: self,
		Info: func(pid int) (int, string, bool) {
			p, ok := tree[pid]
			if !ok {
				return 0, "", false
			}
			return p.ppid, p.comm, true
		},
		Alive: func(pid int) bool {
			_, ok := tree[pid]
			return ok
		},
	}
}

// TestProcessView_WalksTheInjectedTable pins the seam SPEC-HOOK-STOP-PARSE-CAP-001
// plan.md §C 6 asks for: the real ancestry walk runs over an injected table,
// so a caller outside this package can test owner resolution against a
// synthetic tree without replacing the resolver itself.
func TestProcessView_WalksTheInjectedTable(t *testing.T) {
	t.Parallel()

	tree1 := map[int]fakeProcess{
		400: {ppid: 300, comm: "moai"},
		300: {ppid: 200, comm: "bash"},
		200: {ppid: 100, comm: "bash"},
		100: {ppid: 50, comm: "claude"},
		50:  {ppid: 1, comm: "zsh"},
	}
	tree2 := map[int]fakeProcess{
		401: {ppid: 201, comm: "moai"},
		201: {ppid: 100, comm: "bash"},
		100: {ppid: 50, comm: "claude"},
		50:  {ppid: 1, comm: "zsh"},
	}

	for name, v := range map[string]ProcessView{
		"two shells": syntheticView(400, tree1),
		"one shell":  syntheticView(401, tree2),
	} {
		pid, ok := v.ResolveOwnerPID("")
		if !ok || pid != 100 {
			t.Errorf("%s: ResolveOwnerPID = (%d, %v), want (100, true) — the walk must step through every wrapper shell", name, pid, ok)
		}
	}
}

// TestProcessView_StampRules keeps the stamp semantics of ResolveOwnerPID on
// the injected view: an ancestor stamp wins, a foreign stamp is ignored, and
// an unresolvable tree yields (0, false).
func TestProcessView_StampRules(t *testing.T) {
	t.Parallel()

	tree := map[int]fakeProcess{
		400: {ppid: 300, comm: "moai"},
		300: {ppid: 100, comm: "bash"},
		100: {ppid: 50, comm: "claude"},
		50:  {ppid: 1, comm: "zsh"},
		777: {ppid: 1, comm: "claude"},
	}
	v := syntheticView(400, tree)

	if pid, ok := v.ResolveOwnerPID(strconv.Itoa(50)); !ok || pid != 50 {
		t.Errorf("ancestor stamp: got (%d, %v), want (50, true)", pid, ok)
	}
	if pid, ok := v.ResolveOwnerPID(strconv.Itoa(777)); !ok || pid != 100 {
		t.Errorf("foreign stamp: got (%d, %v), want the walk's answer (100, true)", pid, ok)
	}

	empty := ProcessView{
		Self:  400,
		Info:  func(int) (int, string, bool) { return 0, "", false },
		Alive: func(int) bool { return false },
	}
	if pid, ok := empty.ResolveOwnerPID(""); ok || pid != 0 {
		t.Errorf("unresolvable tree: got (%d, %v), want (0, false)", pid, ok)
	}
}

// TestLiveProcessView_ReadsThePackageSeams keeps ResolveOwnerPID and the live
// view on one code path: the live view must read the same procInfo/pidIsAlive
// seams the existing resolver tests swap.
func TestLiveProcessView_ReadsThePackageSeams(t *testing.T) {
	self := LiveProcessView().Self
	withFakeAncestry(t, map[int]fakeProcess{
		self: {ppid: 900, comm: "moai"},
		900:  {ppid: 800, comm: "sh"},
		800:  {ppid: 1, comm: "claude"},
	}, map[int]bool{800: true, 900: true})

	if pid, ok := LiveProcessView().ResolveOwnerPID(""); !ok || pid != 800 {
		t.Errorf("LiveProcessView().ResolveOwnerPID = (%d, %v), want (800, true)", pid, ok)
	}
}
