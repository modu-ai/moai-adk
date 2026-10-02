package hook

// session_start_no_factory_notice_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M5b
// (card t1399), AC-014 / REQ-013: SessionStart emits no kanban notice, whatever
// retired marker a surviving session still carries, and the factory notices are
// unchanged. The marker names are string literals on purpose: M5b deletes the
// three constants, and a session that outlived the upgrade carries the names
// whether or not any Go constant does.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestSessionStartEmitsNoKanbanNotice runs the real SessionStart handler for
// sessions that carry each shape of the retired kanban markers and asserts that
// neither channel (the agent context, the operator system message) names the
// retired mode and that the handler stays non-blocking (netSessionStart fails
// the test on an error or a nil output). A factory leader and a factory lane
// still receive their own notices.
func TestSessionStartEmitsNoKanbanNotice(t *testing.T) {
	retired := []struct {
		name string
		env  map[string]string
	}{
		{"no marker", nil},
		{"leader marker alone", map[string]string{"MOAI_KANBAN": "1"}},
		{"companion label alone", map[string]string{"MOAI_KANBAN_LABEL": "plan"}},
		{"leader marker with run id, socket, and SPEC", map[string]string{
			"MOAI_KANBAN":             "1",
			"MOAI_KANBAN_SPEC":        "SPEC-RETIRED-001",
			config.EnvFactoryRunID:    "oldrun01",
			config.EnvFactoryLeadAddr: "/tmp/moai-socket-kanban/oldrun01",
		}},
		{"both markers", map[string]string{"MOAI_KANBAN": "1", "MOAI_KANBAN_LABEL": "run-2"}},
	}
	for _, tc := range retired {
		t.Run(tc.name, func(t *testing.T) {
			root := newMoaiProjectRoot(t)
			netScrubFactoryEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			ctx, sys := netSessionStart(t, root, "retired-marker-"+strings.ReplaceAll(tc.name, " ", "-"))
			for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
				if strings.Contains(strings.ToLower(text), "kanban") {
					t.Errorf("a session with %v received a kanban notice on %s:\n%s", tc.env, channel, text)
				}
			}
		})
	}

	t.Run("factory leader still receives its notice", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv("MOAI_KANBAN", "1")
		t.Setenv(config.EnvMoaiFactoryWorkers, "1")
		t.Setenv(config.EnvFactoryRunID, "netrun02")
		t.Setenv(config.EnvFactoryLeadAddr, "/tmp/moai-socket-factory/netrun02")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)

		ctx, sys := netSessionStart(t, root, "retired-marker-factory-leader")
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			if !strings.Contains(text, "netrun02") {
				t.Errorf("the factory leader notice on %s lacks the run id:\n%s", channel, text)
			}
			if strings.Contains(text, "Kanban Mode") {
				t.Errorf("the factory leader received a kanban notice on %s:\n%s", channel, text)
			}
		}
	})

	t.Run("factory lane still receives its notice", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv("MOAI_KANBAN_LABEL", "plan")
		t.Setenv(config.EnvMoaiFactoryWorker, factory.FactoryLaneLabel(2))
		t.Setenv(config.EnvMoaiFactoryWorkers, "0")
		t.Setenv(config.EnvFactoryRunID, "netrun02")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)

		ctx, sys := netSessionStart(t, root, "retired-marker-factory-lane")
		for channel, text := range map[string]string{"additionalContext": ctx, "systemMessage": sys} {
			if !strings.Contains(text, factory.FactoryLaneLabel(2)) {
				t.Errorf("the factory lane notice on %s lacks the lane label:\n%s", channel, text)
			}
			if strings.Contains(text, "Kanban Mode") {
				t.Errorf("the factory lane received a kanban notice on %s:\n%s", channel, text)
			}
		}
	})
}

// TestSessionRecordIgnoresRetiredFactoryMarkers pins the session-record role
// reader after M5b: the leader and companion markers of the retired mode map to
// no role, so a session carrying only them writes no record, and the SPEC
// marker never reaches a factory record (its SPEC field is the empty string).
func TestSessionRecordIgnoresRetiredFactoryMarkers(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"leader marker":    {"MOAI_KANBAN": "1"},
		"companion label":  {"MOAI_KANBAN_LABEL": "plan"},
		"companion bumped": {"MOAI_KANBAN_LABEL": "sync-2"},
	} {
		t.Run(name, func(t *testing.T) {
			root := newMoaiProjectRoot(t)
			netScrubFactoryEnv(t)
			for k, v := range env {
				t.Setenv(k, v)
			}
			if role, lane, ok := factoryRoleFromEnv(); ok {
				t.Errorf("factoryRoleFromEnv with %v = (%q, %d, true), want ok=false", env, role, lane)
			}
			writeFactorySessionRecord(&HookInput{SessionID: "retired-marker-sess", ProjectDir: root, CWD: root})
			if _, err := factory.Read(root, "retired-marker-sess"); err == nil {
				t.Errorf("a session with %v wrote a session record", env)
			}
		})
	}

	t.Run("factory leader record carries no SPEC", func(t *testing.T) {
		root := newMoaiProjectRoot(t)
		netScrubFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorkers, "2")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
		t.Setenv("MOAI_KANBAN_SPEC", "SPEC-RETIRED-001")

		writeFactorySessionRecord(&HookInput{SessionID: "spec-marker-sess", ProjectDir: root, CWD: root})
		rec, err := factory.Read(root, "spec-marker-sess")
		if err != nil {
			t.Fatalf("the factory leader wrote no session record: %v", err)
		}
		if rec.SpecID != "" {
			t.Errorf("the record SPEC field = %q, want the empty string", rec.SpecID)
		}
	})
}
