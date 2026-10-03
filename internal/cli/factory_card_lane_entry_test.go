package cli

// factory_card_lane_entry_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M4: the
// factory card verbs refuse a legacy lane label with a message that names the
// lane entry `-l`, never the removed `-f lane` form.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestFactoryCardLegacyLabelErrorsNameLaneEntry(t *testing.T) {
	for _, tc := range []struct{ label, want string }{
		{"worker-2", "lane-2"},
		{"agent-3", "lane-3"},
		{"worker", "lane-<n>"},
		{"agent", "lane-<n>"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			t.Setenv(config.EnvMoaiFactoryWorker, tc.label)
			_, err := factoryLaneLabelFromEnv("stage")
			if err == nil {
				t.Fatalf("legacy label %q was accepted", tc.label)
			}
			msg := err.Error()
			if !strings.Contains(msg, "rejoin with -l") {
				t.Errorf("error does not name the -l rejoin: %s", msg)
			}
			if !strings.Contains(msg, tc.want) {
				t.Errorf("error does not name %q: %s", tc.want, msg)
			}
			if strings.Contains(msg, "-f lane") {
				t.Errorf("error still teaches the removed -f lane form: %s", msg)
			}
		})
	}
}
