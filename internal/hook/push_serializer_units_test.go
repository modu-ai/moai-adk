package hook

// push_serializer_units_test.go — table-driven units behind the push
// serializer's AC tests: the matcher, the exit signal, the activation
// document, the fail-open limbs, and the release variants the ACs do not
// name. Coverage discipline: every branch of push_serializer.go is exercised
// from here or from push_serializer_test.go.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsDevelopPush(t *testing.T) {
	cases := []struct {
		name    string
		command string
		want    bool
	}{
		{"bare push", "git push origin develop", true},
		{"no remote", "git push develop", true},
		{"HEAD refspec", "git push origin HEAD:develop", true},
		{"refs/heads dst", "git push origin main:refs/heads/develop", true},
		{"force push", "git push --force-with-lease origin develop", true},
		{"path to git", "/usr/bin/git push origin develop", true},
		{"push option inline", "git push -o=ci skip origin develop", true},
		{"after scrubbing", `git push 'origin' "develop"`, false}, // substituteQuotedArguments collapses the quoted ref to a placeholder — the deliberate under-match design.md §B names (fail-open direction)
		{"push option separate value", "git push -o ci origin develop", true},
		{"second git in line", "echo hi && git push origin develop", true},
		{"push main", "git push origin main", false},
		{"develop as source", "git push origin develop:main", false},
		{"no refspec", "git push", false},
		{"not git", "moai push develop", false},
		{"git without push", "git commit -m develop", false},
		{"fetch develop", "git fetch origin develop", false},
		{"flag value names develop", "git push --repo develop main", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDevelopPush(tc.command); got != tc.want {
				t.Fatalf("isDevelopPush(%q) = %t, want %t", tc.command, got, tc.want)
			}
		})
	}
}

func TestRefspecTargetsDevelop(t *testing.T) {
	cases := []struct {
		arg  string
		want bool
	}{
		{"develop", true},
		{"HEAD:develop", true},
		{"refs/heads/develop", true},
		{"+refs/heads/*:refs/heads/develop", true},
		{"develop:main", false},
		{"main", false},
		{"develop:", false}, // deletion with no explicit dst does not name develop
		{"developments", false},
	}
	for _, tc := range cases {
		if got := refspecTargetsDevelop(tc.arg); got != tc.want {
			t.Fatalf("refspecTargetsDevelop(%q) = %t, want %t", tc.arg, got, tc.want)
		}
	}
}

func TestPushExitNonZero(t *testing.T) {
	cases := []struct {
		name     string
		response json.RawMessage
		want     bool
	}{
		{"empty", nil, false},
		{"exit zero", pushRawJSON(t, map[string]any{"exit": 0}), false},
		{"exit non-zero", pushRawJSON(t, map[string]any{"exit": 1}), true},
		{"exit_code non-zero", pushRawJSON(t, map[string]any{"exit_code": 127}), true},
		{"no exit key", pushRawJSON(t, map[string]any{"stdout": "ok"}), false},
		{"not an object", json.RawMessage(`"done"`), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pushExitNonZero(tc.response); got != tc.want {
				t.Fatalf("pushExitNonZero(%s) = %t, want %t", tc.response, got, tc.want)
			}
		})
	}
}

func TestDecodePushShowJSONAndArmed(t *testing.T) {
	t.Run("undecodable document is an error", func(t *testing.T) {
		if _, err := DecodePushShowJSON([]byte("{not json")); err == nil {
			t.Fatal("DecodePushShowJSON(bad json) = nil error, want error")
		}
	})
	t.Run("armed truth table", func(t *testing.T) {
		cases := []struct {
			name string
			show *PushShowJSON
			want bool
		}{
			{"nil", nil, false},
			{"fully armed", &PushShowJSON{Mode: "contract", Actions: []string{"push-develop"}, PushRequiresLease: true}, true},
			{"guided mode", &PushShowJSON{Mode: "guided", Actions: []string{"push-develop"}, PushRequiresLease: true}, false},
			{"action missing", &PushShowJSON{Mode: "contract", Actions: []string{"commit"}, PushRequiresLease: true}, false},
			{"lease flag false", &PushShowJSON{Mode: "contract", Actions: []string{"push-develop"}, PushRequiresLease: false}, false},
		}
		for _, tc := range cases {
			if got := tc.show.Armed(); got != tc.want {
				t.Fatalf("%s: Armed() = %t, want %t", tc.name, got, tc.want)
			}
		}
	})
}

// TestPushSerializerFailOpenLimbs pins REQ-AP-002's fail-open direction on the
// limbs the AC-AP-004 fixture does not carry: no root, an unresolvable root,
// a session with no id, and a failed record write. Each allows and leaves
// exactly one audit line (where a root exists to carry it).
func TestPushSerializerFailOpenLimbs(t *testing.T) {
	show := pushArmedShow(t)

	t.Run("no project root allows with no audit", func(t *testing.T) {
		var advisory bytes.Buffer
		dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), "", show, pushBound, &advisory)
		if dec != "" {
			t.Fatalf("decision = %q (%s), want allow", dec, reason)
		}
		if !strings.Contains(advisory.String(), "no project root") {
			t.Fatalf("advisory = %q, want the no-root notice", advisory.String())
		}
	})

	t.Run("unresolvable root allows with one audit line", func(t *testing.T) {
		nonRepo := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", nonRepo)
		var advisory bytes.Buffer
		dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), nonRepo, show, pushBound, &advisory)
		if dec != "" {
			t.Fatalf("decision = %q (%s), want allow", dec, reason)
		}
		entries := slotGuardAudit(t, nonRepo)
		if len(entries) != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", len(entries))
		}
		if event, _ := entries[0]["event"].(string); event != "fail-open" {
			t.Fatalf("audit event = %q, want fail-open", event)
		}
	})

	t.Run("missing session id allows with one audit line", func(t *testing.T) {
		root := slotGuardRepo(t)
		var advisory bytes.Buffer
		dec, _ := checkPushSerializer(slotGuardInput(t, "", pushCommand), root, show, pushBound, &advisory)
		if dec != "" {
			t.Fatalf("decision = %q, want allow", dec)
		}
		entries := slotGuardAudit(t, root)
		if len(entries) != 1 {
			t.Fatalf("audit lines = %d, want exactly 1", len(entries))
		}
		if _, err := os.Stat(pushLeasePath(root)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("record written on the fail-open path (%v), want none", err)
		}
	})

	t.Run("unmatched command is silent", func(t *testing.T) {
		root := slotGuardRepo(t)
		var advisory bytes.Buffer
		dec, _ := checkPushSerializer(slotGuardInput(t, pushSessionA, "git push origin main"), root, show, pushBound, &advisory)
		if dec != "" {
			t.Fatalf("decision = %q, want allow", dec)
		}
		if entries := slotGuardAudit(t, root); entries != nil {
			t.Fatalf("audit lines = %d, want none (command not a develop push)", len(entries))
		}
	})

	t.Run("input without a command is silent", func(t *testing.T) {
		root := slotGuardRepo(t)
		raw, err := json.Marshal(map[string]any{"other": "field"})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		dec, _ := checkPushSerializer(&HookInput{SessionID: pushSessionA, ToolName: "Bash", ToolInput: raw}, root, show, pushBound, &bytes.Buffer{})
		if dec != "" {
			t.Fatalf("decision = %q, want allow", dec)
		}
		if entries := slotGuardAudit(t, root); entries != nil {
			t.Fatalf("audit lines = %d, want none", len(entries))
		}
	})

	t.Run("failed record write allows with one audit line", func(t *testing.T) {
		root := slotGuardRepo(t)
		seed := liveForeignLease()
		seed.Resource = "push-develop"
		seed.SessionID = pushSessionA
		seed.PID = os.Getpid()
		seed.AcquiredAt = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
		seed.ExpiresAt = time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
		seedSlotGuardLease(t, root, seed)
		// A zero bound makes the admit path's acquire fail outright.
		var advisory bytes.Buffer
		dec, _ := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, 0, &advisory)
		if dec != "" {
			t.Fatal("decision = deny, want allow on the failed write")
		}
		entries := slotGuardAudit(t, root)
		if len(entries) == 0 || entries[len(entries)-1]["event"] != "fail-open" {
			t.Fatalf("last audit event = %v, want fail-open", entries)
		}
	})
}

// TestPushSerializerAdmitAfterDeniedHolder pins the reacquire limb: the
// calling session re-pushing while already holding restarts its own bound
// without a takeover (the slot lease's existing self-acquire rule).
func TestPushSerializerAdmitAfterDeniedHolder(t *testing.T) {
	root := slotGuardRepo(t)
	show := pushArmedShow(t)

	if dec, _ := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &bytes.Buffer{}); dec != "" {
		t.Fatal("A's first push did not admit")
	}
	// A re-pushes: admitted again, still the holder.
	if dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &bytes.Buffer{}); dec != "" {
		t.Fatalf("A's re-push: decision = %q (%s), want admit", dec, reason)
	}
	if got := pushLeaseHeldBy(t, root); got != pushSessionA {
		t.Fatalf("lease holder = %q, want %q", got, pushSessionA)
	}
}

// TestPushLeaseReleaseVariants covers the release limbs the AC-AP-003 flow
// does not name: a successful exit, an absent exit signal, a foreign holder,
// and a non-push command each keep the record.
func TestPushLeaseReleaseVariants(t *testing.T) {
	seedHolder := func(t *testing.T) string {
		t.Helper()
		root := slotGuardRepo(t)
		show := pushArmedShow(t)
		if dec, _ := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &bytes.Buffer{}); dec != "" {
			t.Fatal("holder setup did not admit")
		}
		return root
	}

	cases := []struct {
		name     string
		response json.RawMessage
		command  string
	}{
		{"successful exit keeps the lease", pushRawJSON(t, map[string]any{"exit": 0}), pushCommand},
		{"absent exit signal keeps the lease", nil, pushCommand},
		{"failed non-push keeps the lease", pushRawJSON(t, map[string]any{"exit": 2}), "git push origin main"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := seedHolder(t)
			release := &HookInput{
				SessionID:     pushSessionA,
				HookEventName: "PostToolUse",
				ToolName:      "Bash",
				ToolInput:     pushRawJSON(t, map[string]any{"command": tc.command}),
				ToolResponse:  tc.response,
			}
			releasePushLeaseOnFailure(release, root, pushArmedShow(t), &bytes.Buffer{})
			if got := pushLeaseHeldBy(t, root); got != pushSessionA {
				t.Fatalf("lease holder = %q, want it kept by %q", got, pushSessionA)
			}
		})
	}

	t.Run("foreign holder is not released", func(t *testing.T) {
		root := seedHolder(t)
		release := &HookInput{
			SessionID:     pushSessionB,
			HookEventName: "PostToolUse",
			ToolName:      "Bash",
			ToolInput:     pushRawJSON(t, map[string]any{"command": pushCommand}),
			ToolResponse:  pushRawJSON(t, map[string]any{"exit": 1}),
		}
		releasePushLeaseOnFailure(release, root, pushArmedShow(t), &bytes.Buffer{})
		if got := pushLeaseHeldBy(t, root); got != pushSessionA {
			t.Fatalf("lease holder = %q, want it kept by the foreign holder %q", got, pushSessionA)
		}
	})

	t.Run("inactive serializer releases nothing", func(t *testing.T) {
		root := seedHolder(t)
		release := &HookInput{
			SessionID:     pushSessionA,
			HookEventName: "PostToolUse",
			ToolName:      "Bash",
			ToolInput:     pushRawJSON(t, map[string]any{"command": pushCommand}),
			ToolResponse:  pushRawJSON(t, map[string]any{"exit": 1}),
		}
		releasePushLeaseOnFailure(release, root, nil, &bytes.Buffer{})
		if got := pushLeaseHeldBy(t, root); got != pushSessionA {
			t.Fatalf("lease holder = %q, want it kept", got)
		}
	})

	t.Run("session without id releases nothing", func(t *testing.T) {
		root := seedHolder(t)
		release := &HookInput{
			HookEventName: "PostToolUse",
			ToolName:      "Bash",
			ToolInput:     pushRawJSON(t, map[string]any{"command": pushCommand}),
			ToolResponse:  pushRawJSON(t, map[string]any{"exit": 1}),
		}
		releasePushLeaseOnFailure(release, root, pushArmedShow(t), &bytes.Buffer{})
		if got := pushLeaseHeldBy(t, root); got != pushSessionA {
			t.Fatalf("lease holder = %q, want it kept", got)
		}
	})

	t.Run("unresolvable root releases nothing and advises", func(t *testing.T) {
		root := seedHolder(t)
		nonRepo := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", nonRepo)
		var advisory bytes.Buffer
		release := &HookInput{
			SessionID:     pushSessionA,
			HookEventName: "PostToolUse",
			ToolName:      "Bash",
			ToolInput:     pushRawJSON(t, map[string]any{"command": pushCommand}),
			ToolResponse:  pushRawJSON(t, map[string]any{"exit": 1}),
		}
		releasePushLeaseOnFailure(release, nonRepo, pushArmedShow(t), &advisory)
		if !strings.Contains(advisory.String(), "shared root") {
			t.Fatalf("advisory = %q, want the root notice", advisory.String())
		}
		if got := pushLeaseHeldBy(t, root); got != pushSessionA {
			t.Fatalf("lease holder = %q, want it kept", got)
		}
	})
}

// TestPushSerializerShowSeam pins the activation seam: unset loader, loader
// error, undecodable document, and a disarmed document each resolve inactive;
// an armed document resolves to the decoded show values.
func TestPushSerializerShowSeam(t *testing.T) {
	orig := pushShowJSONLoader
	t.Cleanup(func() { pushShowJSONLoader = orig })

	reset := func(loader func(string) ([]byte, error)) {
		pushShowJSONLoader = loader
	}

	t.Run("unset loader is inactive", func(t *testing.T) {
		reset(nil)
		if show := pushSerializerShow(t.TempDir()); show != nil {
			t.Fatalf("pushSerializerShow = %+v, want nil", show)
		}
	})
	t.Run("loader error is inactive", func(t *testing.T) {
		reset(func(string) ([]byte, error) { return nil, errors.New("no resolver") })
		if show := pushSerializerShow(t.TempDir()); show != nil {
			t.Fatalf("pushSerializerShow = %+v, want nil", show)
		}
	})
	t.Run("disarmed document is inactive", func(t *testing.T) {
		reset(func(string) ([]byte, error) { return pushShowDoc(t, "guided", []string{"push-develop"}, nil), nil })
		if show := pushSerializerShow(t.TempDir()); show != nil {
			t.Fatalf("pushSerializerShow = %+v, want nil", show)
		}
	})
	t.Run("armed document resolves", func(t *testing.T) {
		reset(func(string) ([]byte, error) { return pushArmedShowBytes(t), nil })
		show := pushSerializerShow(t.TempDir())
		if show == nil || !show.Armed() {
			t.Fatalf("pushSerializerShow = %+v, want armed", show)
		}
	})
}

func pushArmedShowBytes(t *testing.T) []byte {
	t.Helper()
	truev := true
	return pushShowDoc(t, "contract", []string{"push-develop"}, &truev)
}

// TestPushSerializerBoundParsesConfig pins the bound resolution against a
// project config carrying a custom default_max_duration.
func TestPushSerializerBoundParsesConfig(t *testing.T) {
	root := slotGuardRepo(t)
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfg := "workflow:\n  slot_lease:\n    default_max_duration: 45m\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	if got := pushSerializerBound(root); got != 45*time.Minute {
		t.Fatalf("pushSerializerBound = %s, want 45m", got)
	}

	t.Run("unparseable bound resolves zero (fail open)", func(t *testing.T) {
		cfg := "workflow:\n  slot_lease:\n    default_max_duration: soon\n"
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(cfg), 0o644); err != nil {
			t.Fatalf("write workflow.yaml: %v", err)
		}
		if got := pushSerializerBound(root); got != 0 {
			t.Fatalf("pushSerializerBound = %s, want 0", got)
		}
	})
}
