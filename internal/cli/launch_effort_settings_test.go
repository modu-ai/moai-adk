package cli

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// readSettingsPayload decodes a transient settings file written by one of the
// launcher injection points.
func readSettingsPayload(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read injected settings: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("injected settings is not valid JSON: %v (%s)", err, data)
	}
	return payload
}

// withLaunchEffortPrefs pins the profile preferences the effort resolution
// reads, so no test depends on the host's real profile. Restored on cleanup.
func withLaunchEffortPrefs(t *testing.T, prefs profile.ProfilePreferences, err error) {
	t.Helper()
	orig := launchEffortPrefsFn
	launchEffortPrefsFn = func(string) (profile.ProfilePreferences, error) { return prefs, err }
	t.Cleanup(func() { launchEffortPrefsFn = orig })
}

// withNoLaunchEffort pins an empty profile, so applyLaunchEffort contributes
// nothing. Existing injection tests use it to stay byte-identical to the
// pre-effort payload regardless of what the host's profile happens to say.
func withNoLaunchEffort(t *testing.T) {
	t.Helper()
	withLaunchEffortPrefs(t, profile.ProfilePreferences{}, nil)
}

// TestApplyLaunchEffort covers the three resolution outcomes plus the
// fail-open path. The resolution itself is resolveLaunchEffort's — covered by
// TestResolveLaunchEffort — so what is judged here is which of them reaches the
// payload, and under what key.
func TestApplyLaunchEffort(t *testing.T) {
	t.Run("explicit effort_level lands under effortLevel", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "xhigh", ModelPolicy: "low"}, nil)
		got := applyLaunchEffort(map[string]any{}, "dev")
		if got[effortSettingsKey] != "xhigh" {
			t.Errorf("%s = %v, want xhigh (explicit effort_level wins)", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("model_policy supplies the fallback", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{ModelPolicy: "high"}, nil)
		got := applyLaunchEffort(map[string]any{}, "dev")
		if got[effortSettingsKey] != "high" {
			t.Errorf("%s = %v, want high (model_policy fallback)", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("both empty injects nothing", func(t *testing.T) {
		withNoLaunchEffort(t)
		got := applyLaunchEffort(map[string]any{}, "dev")
		if _, ok := got[effortSettingsKey]; ok {
			t.Errorf("%s present (%v) for an effort-less profile; the launch must stay byte-identical", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("unreadable profile fails open", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, errors.New("read preferences: boom"))
		got := applyLaunchEffort(map[string]any{"crossSessionInbound": "accept"}, "dev")
		if _, ok := got[effortSettingsKey]; ok {
			t.Errorf("%s injected despite a read error; the effort must fail open", effortSettingsKey)
		}
		if got["crossSessionInbound"] != "accept" {
			t.Errorf("existing payload key lost on the fail-open path: %v", got)
		}
	})

	t.Run("nil payload is materialized", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "medium"}, nil)
		got := applyLaunchEffort(nil, "dev")
		if got[effortSettingsKey] != "medium" {
			t.Errorf("%s = %v, want medium from a nil payload", effortSettingsKey, got[effortSettingsKey])
		}
	})
}

// TestClaudeLaunchEnvPreservesInheritedEffort is AC-FM-023c's Claude-path half:
// the launch env the Claude backend builds must carry an inherited
// CLAUDE_CODE_EFFORT_LEVEL through unchanged — neither stripped nor replaced.
//
// It asserts BEHAVIOR, not identity, so it is not a tautology over os.Environ():
// a wrapper that scrubs the variable drops the entry (red), and one that pins a
// profile effort replaces its value (red) — the exact defect card t595 removed.
// The GLM half, where a wrapper still filters the environment, is
// TestACFM023c_KanbanEnvReachesChildEnvironment.
func TestClaudeLaunchEnvPreservesInheritedEffort(t *testing.T) {
	t.Parallel()

	key := config.EnvClaudeCodeEffortLevel
	base := []string{"PATH=/usr/bin", key + "=xhigh", "HOME=/home/user"}

	got := buildEnvForClaudeLaunch(base)

	count, value := 0, ""
	for _, e := range got {
		if strings.HasPrefix(e, key+"=") {
			count++
			value = strings.TrimPrefix(e, key+"=")
		}
	}
	switch {
	case count == 0:
		t.Errorf("%s was stripped from the Claude launch env; an inherited value is the user's own per-session override", key)
	case count > 1:
		t.Errorf("%s appears %d times in the Claude launch env, want exactly 1", key, count)
	case value != "xhigh":
		t.Errorf("%s = %q, want xhigh — the Claude launch env must not pin an effort (that override refuses in-session /effort changes)", key, value)
	}
	if len(got) != len(base) {
		t.Errorf("Claude launch env changed length %d -> %d; it must pass the inherited environment through untouched", len(base), len(got))
	}
}

// TestLaunchEffortReachesGeneralInjection is the load-bearing link for the
// general launch path: a profile effort with an otherwise-neutral
// crosssession.yaml must still produce a --settings file carrying effortLevel.
// Before card t595 the effort travelled in CLAUDE_CODE_EFFORT_LEVEL, which
// Claude Code treats as an override that refuses an in-session /effort or
// /model change; the settings file carries it as a launch DEFAULT instead.
func TestLaunchEffortReachesGeneralInjection(t *testing.T) {
	root := withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "xhigh"}, nil)

	got := appendCrossSessionSettings(root, "dev", []string{"-p", "dev"})
	if len(got) != 4 || got[2] != settingsFlagLong {
		t.Fatalf("args = %v, want [-p dev --settings <path>] (effort-only payload still injects)", got)
	}
	payload := readSettingsPayload(t, got[3])
	if payload[effortSettingsKey] != "xhigh" {
		t.Errorf("%s = %v, want xhigh in the injected settings", effortSettingsKey, payload[effortSettingsKey])
	}
}

// TestLaunchEffortReachesKanbanInjection is the same link for the kanban /
// factory lanes, whose args already carry an injected --settings by the time
// they reach the general funnel — so the lane payload has to carry the effort
// itself or a lane would launch with no profile effort at all.
func TestLaunchEffortReachesKanbanInjection(t *testing.T) {
	withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{ModelPolicy: "high"}, nil)

	flag, cleanup := prepareKanbanSettings("dev", []string{"-p", "dev"})
	t.Cleanup(cleanup)
	if len(flag) != 2 || flag[0] != settingsFlagLong {
		t.Fatalf("flag = %v, want [--settings <path>]", flag)
	}
	payload := readSettingsPayload(t, flag[1])
	if payload[effortSettingsKey] != "high" {
		t.Errorf("%s = %v, want high (model_policy fallback reaches a lane)", effortSettingsKey, payload[effortSettingsKey])
	}
	if payload["crossSessionInbound"] != "accept" {
		t.Errorf("crossSessionInbound = %v, want accept — the dispatch requirement must survive the effort merge", payload["crossSessionInbound"])
	}
}

// TestLaunchEffortNeverInjectedWhenOperatorSuppliedSettings pins the documented
// limitation: an operator who passes --settings themselves owns the settings
// surface, so the profile effort is not merged into their file.
func TestLaunchEffortNeverInjectedWhenOperatorSuppliedSettings(t *testing.T) {
	root := withCrossSessionConfig(t, "crosssession:\n  inbound: accept\n")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)

	args := []string{settingsFlagLong, "/tmp/operator.json"}
	if got := appendCrossSessionSettings(root, "dev", args); len(got) != 2 {
		t.Errorf("args = %v, want unchanged (operator-supplied --settings wins over the effort injection)", got)
	}
}
