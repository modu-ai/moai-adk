package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFactoryAgeAlarmReviewWarningKeepsStatusGrammar(t *testing.T) {
	for _, tc := range []struct{ name, body, key string }{
		{"duplicate-enabled", "workflow:\n  card_age_alarm:\n    enabled: true\n    enabled: false\n", "workflow.card_age_alarm"},
		{"null-threshold", "workflow:\n  card_age_alarm:\n    enabled: true\n    gate_warn_after: null\n", "workflow.card_age_alarm.gate_warn_after"},
		{"merged-malformed-threshold", "base: &base\n  gate_warn_after: [6h]\nworkflow:\n  card_age_alarm:\n    <<: *base\n    enabled: true\n", "workflow.card_age_alarm.gate_warn_after"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := fcFixture(t)
			caaSeedAlarmCards(t, root)
			baseline, _, err := runFactory(t, "status")
			if err != nil {
				t.Fatal(err)
			}
			caaEnableAlarm(t, root, tc.body)
			out, _, err := runFactory(t, "status")
			if err != nil {
				t.Fatalf("invalid alarm refused the status command: %v", err)
			}
			var withoutWarning []string
			warnings := 0
			for _, line := range strings.Split(out, "\n") {
				if strings.HasPrefix(line, "age-alarm warning:") {
					warnings++
					if !strings.Contains(line, tc.key) {
						t.Errorf("warning did not name %s: %s", tc.key, line)
					}
				} else {
					withoutWarning = append(withoutWarning, line)
				}
			}
			if warnings != 1 || strings.Join(withoutWarning, "\n") != baseline {
				t.Errorf("expected exactly one warning line and byte-identical remaining status; warnings=%d\nbaseline:\n%s\nactual:\n%s", warnings, baseline, out)
			}
			outJSON, _, err := runFactory(t, "status", "--json")
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal([]byte(outJSON), &got); err != nil {
				t.Fatal(err)
			}
			warning, _ := got["age_alarm_warning"].(string)
			if !strings.Contains(warning, tc.key) {
				t.Errorf("JSON warning did not name %s: %s", tc.key, outJSON)
			}
			if _, present := got["age_alarms"]; present {
				t.Errorf("refused alarm still rendered rows: %s", outJSON)
			}
		})
	}
}
