package verify

import (
	"strings"
	"testing"
	"time"
)

// runTestState builds the receipt state that matches runTestSnapshot's entry.
func runTestState(cmd string) ReceiptState {
	return ReceiptState{
		Head:         "abc123",
		TreeDigest:   "0123456789abcdef",
		ConfigDigest: "cfg-1",
		Command:      cmd,
		ToolVersion:  "go1.26",
	}
}

// runTestSnapshot builds a snapshot holding one verify-run style entry
// recorded at at.
func runTestSnapshot(cmd string, exit int, verdict string, at time.Time) *Snapshot {
	return &Snapshot{
		Key:        "abc123:0123456789abcdef",
		RecordedAt: at,
		Checks: []CheckEntry{{
			CheckID:      "run",
			Command:      cmd,
			ExitCode:     exit,
			RecordedAt:   at,
			DurationMS:   42,
			ConfigDigest: "cfg-1",
			ToolVersion:  "go1.26",
			Verdict:      verdict,
		}},
	}
}

func TestDecideReuseHit(t *testing.T) {
	now := time.Now()
	snap := runTestSnapshot("go test ./...", 0, "pass", now.Add(-time.Minute))
	hit, entry, reason := DecideReuse(snap, runTestState("go test ./..."), now, 0)
	if !hit || entry == nil {
		t.Fatalf("want hit, got hit=%v entry=%v reason=%q", hit, entry, reason)
	}
	if entry.DurationMS != 42 {
		t.Errorf("entry.DurationMS = %d, want 42", entry.DurationMS)
	}
}

func TestCanonicalCommand(t *testing.T) {
	cases := []struct {
		name string
		argv []string
		want string
	}{
		{"plain", []string{"go", "test", "./..."}, "go test ./..."},
		{"element with space is quoted", []string{"a b"}, `"a b"`},
		{"two elements differ from one spaced element", []string{"a", "b"}, "a b"},
		{"empty element is quoted", []string{"x", ""}, `x ""`},
		{"quote char is quoted", []string{"say", `"hi"`}, `say "\"hi\""`},
		{"tab is quoted", []string{"a\tb"}, `"a\tb"`},
	}
	for _, tc := range cases {
		if got := CanonicalCommand(tc.argv); got != tc.want {
			t.Errorf("%s: CanonicalCommand(%q) = %q, want %q", tc.name, tc.argv, got, tc.want)
		}
	}
	if CanonicalCommand([]string{"a b"}) == CanonicalCommand([]string{"a", "b"}) {
		t.Error("one element \"a b\" and two elements a,b must canonicalize differently")
	}
}

func TestDecideReuseKeyMismatch(t *testing.T) {
	now := time.Now()
	snap := runTestSnapshot("go test", 0, "pass", now)
	state := runTestState("go test")
	state.TreeDigest = "fedcba9876543210"
	if hit, _, reason := DecideReuse(snap, state, now, 0); hit || reason == "" {
		t.Errorf("tree digest differs: hit=%v reason=%q, want miss with a reason", hit, reason)
	}
	state = runTestState("go test")
	state.Head = "def456"
	if hit, _, _ := DecideReuse(snap, state, now, 0); hit {
		t.Error("head differs: want miss")
	}
	if hit, _, reason := DecideReuse(nil, runTestState("go test"), now, 0); hit || reason == "" {
		t.Errorf("nil snapshot: hit=%v reason=%q, want miss with a reason", hit, reason)
	}
}

func TestDecideReuseCommandBytes(t *testing.T) {
	now := time.Now()
	snap := runTestSnapshot("go test ./a", 0, "pass", now)
	for _, cmd := range []string{"go test ./a ", "go  test ./a", "go test  ./a", "go test ./b", "go test"} {
		if hit, _, _ := DecideReuse(snap, runTestState(cmd), now, 0); hit {
			t.Errorf("command %q differs by bytes: want miss", cmd)
		}
	}
}

func TestDecideReuseNonzeroExit(t *testing.T) {
	now := time.Now()
	snap := runTestSnapshot("go test", 1, "fail", now)
	if hit, _, reason := DecideReuse(snap, runTestState("go test"), now, 0); hit {
		t.Errorf("failed entry reused (reason %q)", reason)
	}
	// A nonzero exit code is never reused even when the verdict word says pass.
	snap = runTestSnapshot("go test", 2, "pass", now)
	if hit, _, _ := DecideReuse(snap, runTestState("go test"), now, 0); hit {
		t.Error("nonzero exit code with verdict pass was reused")
	}
	// A verdict other than pass is not reused even with exit 0.
	snap = runTestSnapshot("go test", 0, "fail", now)
	if hit, _, _ := DecideReuse(snap, runTestState("go test"), now, 0); hit {
		t.Error("verdict fail with exit 0 was reused")
	}
	// An entry written by `moai verify record` has no verdict, digest or tool.
	snap = &Snapshot{Key: "abc123:0123456789abcdef", Checks: []CheckEntry{{Command: "go test", RecordedAt: now}}}
	if hit, _, _ := DecideReuse(snap, runTestState("go test"), now, 0); hit {
		t.Error("hand-recorded entry (no config_digest/tool_version) was reused")
	}
}

func TestDecideReuseTTL(t *testing.T) {
	recorded := time.Now()
	snap := runTestSnapshot("go test", 0, "pass", recorded)
	state := runTestState("go test")
	if hit, _, _ := DecideReuse(snap, state, recorded.Add(9*time.Minute), 0); !hit {
		t.Error("inside the default TTL: want hit")
	}
	if hit, _, reason := DecideReuse(snap, state, recorded.Add(11*time.Minute), 0); hit || !strings.Contains(reason, "older") {
		t.Errorf("past the default TTL: hit=%v reason=%q, want miss naming age", hit, reason)
	}
	if hit, _, _ := DecideReuse(snap, state, recorded.Add(2*time.Second), time.Second); hit {
		t.Error("past an explicit 1s TTL: want miss")
	}
}

func TestEnvDigest(t *testing.T) {
	env := map[string]string{"GOFLAGS": "-a"}
	lookup := func(name string) (string, bool) { v, ok := env[name]; return v, ok }

	base := EnvDigest([]string{"GOFLAGS"}, lookup)
	if base == "" {
		t.Fatal("digest must be non-empty")
	}
	if again := EnvDigest([]string{"GOFLAGS"}, lookup); again != base {
		t.Errorf("same input, different digest: %q vs %q", base, again)
	}

	env["GOFLAGS"] = "-b"
	changed := EnvDigest([]string{"GOFLAGS"}, lookup)
	delete(env, "GOFLAGS")
	unset := EnvDigest([]string{"GOFLAGS"}, lookup)
	env["GOFLAGS"] = ""
	empty := EnvDigest([]string{"GOFLAGS"}, lookup)

	seen := map[string]string{"-a": base, "-b": changed, "unset": unset, "empty": empty}
	for a, da := range seen {
		for b, db := range seen {
			if a < b && da == db {
				t.Errorf("digests for %q and %q collide: %q", a, b, da)
			}
		}
	}

	// A variable set to the literal marker text must not read as unset.
	env["GOFLAGS"] = "\x00unset"
	if EnvDigest([]string{"GOFLAGS"}, lookup) == unset {
		t.Error("a value equal to the unset marker collides with unset")
	}

	// No names yields the digest of the empty input set, equal to ConfigDigest.
	if got, want := EnvDigest(nil, lookup), ConfigDigest(map[string]string{}); got != want || got == "" {
		t.Errorf("EnvDigest(nil) = %q, want ConfigDigest(empty) = %q", got, want)
	}
	// Names are order independent.
	env["A"], env["B"] = "1", "2"
	if EnvDigest([]string{"A", "B"}, lookup) != EnvDigest([]string{"B", "A"}, lookup) {
		t.Error("digest depends on name order")
	}
}
