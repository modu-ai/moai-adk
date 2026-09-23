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
		got, _ := applyLaunchEffort(map[string]any{}, "dev")
		if got[effortSettingsKey] != "xhigh" {
			t.Errorf("%s = %v, want xhigh (explicit effort_level wins)", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("model_policy supplies the fallback", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{ModelPolicy: "high"}, nil)
		got, _ := applyLaunchEffort(map[string]any{}, "dev")
		if got[effortSettingsKey] != "high" {
			t.Errorf("%s = %v, want high (model_policy fallback)", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("both empty injects nothing", func(t *testing.T) {
		withNoLaunchEffort(t)
		got, _ := applyLaunchEffort(map[string]any{}, "dev")
		if _, ok := got[effortSettingsKey]; ok {
			t.Errorf("%s present (%v) for an effort-less profile; the launch must stay byte-identical", effortSettingsKey, got[effortSettingsKey])
		}
	})

	t.Run("unreadable profile fails open", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, errors.New("read preferences: boom"))
		got, _ := applyLaunchEffort(map[string]any{"crossSessionInbound": "accept"}, "dev")
		if _, ok := got[effortSettingsKey]; ok {
			t.Errorf("%s injected despite a read error; the effort must fail open", effortSettingsKey)
		}
		if got["crossSessionInbound"] != "accept" {
			t.Errorf("existing payload key lost on the fail-open path: %v", got)
		}
	})

	t.Run("max returns launch argv and stays out of the payload", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)
		got, args := applyLaunchEffort(map[string]any{}, "dev")
		if _, ok := got[effortSettingsKey]; ok {
			t.Errorf("%s = %v; max is not an accepted settings level", effortSettingsKey, got[effortSettingsKey])
		}
		if len(args) != 2 || args[0] != launchEffortFlag || args[1] != "max" {
			t.Errorf("launch args = %v, want [%s max]", args, launchEffortFlag)
		}
	})

	t.Run("nil payload is materialized", func(t *testing.T) {
		withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "medium"}, nil)
		got, _ := applyLaunchEffort(nil, "dev")
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

// launchEffortFlag is the Claude Code launch flag that sets the effort for one
// session. Spelled literally here so the tests do not borrow the production
// constant they are checking.
const launchEffortFlag = "--effort"

// injectedSettingsEffort returns the effortLevel carried by the transient
// settings file an argv points at via --settings, and whether a --settings
// flag was present at all.
func injectedSettingsEffort(t *testing.T, args []string) (effort any, hasSettings bool) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == settingsFlagLong {
			payload := readSettingsPayload(t, args[i+1])
			return payload[effortSettingsKey], true
		}
	}
	return nil, false
}

// countEffortFlags counts `--effort` occurrences in an argv and returns the
// value following the last one.
func countEffortFlags(args []string) (count int, value string) {
	for i := 0; i < len(args); i++ {
		if args[i] == launchEffortFlag && i+1 < len(args) {
			count++
			value = args[i+1]
		}
	}
	return count, value
}

// TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection pins the K4 fix on the
// general launch path: Claude Code does not accept `max` as a settings
// effortLevel, so a resolved max must never be written there and must reach
// the launch argv as `--effort max` (session-scoped) instead.
func TestLaunchEffortMaxTravelsAsArgvOnGeneralInjection(t *testing.T) {
	root := withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)

	got := appendCrossSessionSettings(root, "dev", []string{"-p", "dev"})
	if effort, ok := injectedSettingsEffort(t, got); ok && effort != nil {
		t.Errorf("settings %s = %v; max must never be written to the settings payload", effortSettingsKey, effort)
	}
	if n, v := countEffortFlags(got); n != 1 || v != "max" {
		t.Errorf("args = %v, want exactly one %s max", got, launchEffortFlag)
	}
}

// TestLaunchEffortMaxTravelsAsArgvOnKanbanInjection is the same pin for the
// kanban / factory lane injection.
func TestLaunchEffortMaxTravelsAsArgvOnKanbanInjection(t *testing.T) {
	withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)

	flag, cleanup := prepareKanbanSettings("dev", []string{"-p", "dev"})
	t.Cleanup(cleanup)
	effort, ok := injectedSettingsEffort(t, flag)
	if !ok {
		t.Fatalf("flag = %v, want a --settings pair (crossSessionInbound: accept is still required)", flag)
	}
	if effort != nil {
		t.Errorf("settings %s = %v; max must never be written to the settings payload", effortSettingsKey, effort)
	}
	if n, v := countEffortFlags(flag); n != 1 || v != "max" {
		t.Errorf("flag = %v, want exactly one %s max", flag, launchEffortFlag)
	}
}

// TestLaunchEffortXHighStaysOnSettingsPath guards the other direction: every
// level the settings key accepts keeps travelling as a launch DEFAULT, so an
// in-session /effort change can still replace it — no --effort flag appears.
func TestLaunchEffortXHighStaysOnSettingsPath(t *testing.T) {
	root := withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "xhigh"}, nil)

	got := appendCrossSessionSettings(root, "dev", []string{"-p", "dev"})
	if effort, _ := injectedSettingsEffort(t, got); effort != "xhigh" {
		t.Errorf("settings %s = %v, want xhigh", effortSettingsKey, effort)
	}
	if n, _ := countEffortFlags(got); n != 0 {
		t.Errorf("args = %v carry %s; only max leaves the settings path", got, launchEffortFlag)
	}

	flag, cleanup := prepareKanbanSettings("dev", []string{"-p", "dev"})
	t.Cleanup(cleanup)
	if effort, _ := injectedSettingsEffort(t, flag); effort != "xhigh" {
		t.Errorf("kanban settings %s = %v, want xhigh", effortSettingsKey, effort)
	}
	if n, _ := countEffortFlags(flag); n != 0 {
		t.Errorf("kanban flag = %v carries %s; only max leaves the settings path", flag, launchEffortFlag)
	}
}

// countEffortTokens counts every --effort token in an argv, in both the
// `--effort X` and `--effort=X` spellings.
func countEffortTokens(args []string) int {
	n := 0
	for _, a := range args {
		if a == launchEffortFlag || strings.HasPrefix(a, launchEffortFlag+"=") {
			n++
		}
	}
	return n
}

// TestLaunchEffortOperatorEffortAnywhereSuppressesInjection: the launcher
// forwards everything after `--` to Claude Code and appends its injected flags
// after it, so an operator --effort on either side of `--`, in either
// spelling, must suppress the profile's `--effort max` on both injection
// paths — otherwise Claude Code receives two --effort flags.
func TestLaunchEffortOperatorEffortAnywhereSuppressesInjection(t *testing.T) {
	shapes := [][]string{
		{launchEffortFlag, "low"},
		{launchEffortFlag + "=low"},
		{"--", launchEffortFlag, "low"},
		{"--", launchEffortFlag + "=low"},
	}
	for _, op := range shapes {
		t.Run("general "+strings.Join(op, " "), func(t *testing.T) {
			root := withCrossSessionConfig(t, "")
			withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)
			got := appendCrossSessionSettings(root, "dev", append([]string(nil), op...))
			if n := countEffortTokens(got); n != 1 {
				t.Errorf("op=%v argv=%v effortFlags=%d want 1 (the operator's)", op, got, n)
			}
		})
		t.Run("kanban "+strings.Join(op, " "), func(t *testing.T) {
			withCrossSessionConfig(t, "")
			withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)
			flag, cleanup := prepareKanbanSettings("dev", append([]string(nil), op...))
			t.Cleanup(cleanup)
			if n := countEffortTokens(flag); n != 0 {
				t.Errorf("op=%v injected=%v effortFlags=%d want 0", op, flag, n)
			}
		})
	}
}

// TestLaunchEffortMaxDefersToOperatorEffortFlag: an operator who passes
// --effort themselves owns the session effort; the profile max adds no
// second flag.
func TestLaunchEffortMaxDefersToOperatorEffortFlag(t *testing.T) {
	root := withCrossSessionConfig(t, "")
	withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)

	got := appendCrossSessionSettings(root, "dev", []string{launchEffortFlag, "low"})
	if n, v := countEffortFlags(got); n != 1 || v != "low" {
		t.Errorf("args = %v, want only the operator's %s low", got, launchEffortFlag)
	}
}

// TestLaunchEffortValuePositionEffortStillInjectsMax: an --effort token that is
// the VALUE of a free-text option (the operator's prompt text happens to read
// `--effort...`) is not an operator --effort flag, so the profile's resolved
// max must still be injected as `--effort max` on both injection paths.
func TestLaunchEffortValuePositionEffortStillInjectsMax(t *testing.T) {
	shapes := [][]string{
		{"--append-system-prompt", launchEffortFlag + "=low"},
		{"--append-system-prompt", launchEffortFlag},
		{"--system-prompt", launchEffortFlag},
		{"--", "--append-system-prompt", launchEffortFlag + "=low"},
	}
	hasMaxPair := func(args []string) bool {
		for i := 0; i+1 < len(args); i++ {
			if args[i] == launchEffortFlag && args[i+1] == "max" {
				return true
			}
		}
		return false
	}
	for _, op := range shapes {
		t.Run("general "+strings.Join(op, " "), func(t *testing.T) {
			root := withCrossSessionConfig(t, "")
			withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)
			got := appendCrossSessionSettings(root, "dev", append([]string(nil), op...))
			if !hasMaxPair(got[len(op):]) {
				t.Errorf("op=%v argv=%v: want an injected %s max after the operator args", op, got, launchEffortFlag)
			}
		})
		t.Run("kanban "+strings.Join(op, " "), func(t *testing.T) {
			withCrossSessionConfig(t, "")
			withLaunchEffortPrefs(t, profile.ProfilePreferences{EffortLevel: "max"}, nil)
			flag, cleanup := prepareKanbanSettings("dev", append([]string(nil), op...))
			t.Cleanup(cleanup)
			if !hasMaxPair(flag) {
				t.Errorf("op=%v injected=%v: want %s max", op, flag, launchEffortFlag)
			}
		})
	}
}

// TestLaunchEffortOperatorEffortAfterPromptValueStillSuppresses guards the
// value-position skip from over-reaching: only the one token after a prompt
// option is skipped, so a real operator --effort that follows the prompt text
// still suppresses the injection.
func TestLaunchEffortOperatorEffortAfterPromptValueStillSuppresses(t *testing.T) {
	for _, op := range [][]string{
		{"--append-system-prompt", "be terse", launchEffortFlag, "low"},
		{"--append-system-prompt=be terse", launchEffortFlag + "=low"},
		{"--system-prompt", "x", "--", launchEffortFlag, "low"},
	} {
		if !operatorSuppliedEffort(op) {
			t.Errorf("operatorSuppliedEffort(%v) = false, want true (the operator's --effort follows the prompt value)", op)
		}
	}
}
