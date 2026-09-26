package config

import "testing"

// SPEC-AUTONOMY-ESCALATION-001 adds exactly one key beside A1's autonomy
// block: workflow.autonomy.escalation.new_api_detector (graph | off). Absent
// or empty takes the default graph; an out-of-set value falls back to graph
// with a warning naming the key, the same shape as every other autonomy key.
func TestResolveAutonomy_NewAPIDetector(t *testing.T) {
	cases := []struct {
		name, body, want string
		warn             bool
	}{
		{"absent", "", "graph", false},
		{"no_key", "workflow:\n  autonomy:\n    mode: contract\n", "graph", false},
		{"graph", "workflow:\n  autonomy:\n    escalation:\n      new_api_detector: graph\n", "graph", false},
		{"off", "workflow:\n  autonomy:\n    escalation:\n      new_api_detector: \"off\"\n", "off", false},
		{"bogus", "workflow:\n  autonomy:\n    escalation:\n      new_api_detector: ast\n", "graph", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := loadAutonomyFixture(t, tc.body)
			if s.NewAPIDetector != tc.want {
				t.Errorf("new_api_detector = %q, want %q", s.NewAPIDetector, tc.want)
			}
			w := warningFor(s, "workflow.autonomy.escalation.new_api_detector")
			if tc.warn && w == "" {
				t.Errorf("no warning names the key; warnings = %v", s.Warnings)
			}
			if !tc.warn && w != "" {
				t.Errorf("unexpected warning %q", w)
			}
		})
	}
}

// The new key is a Config struct addition, so the cache schema must move past
// the version that predates it (see configCacheSchemaVersion).
func TestAutonomy_CacheSchemaBumpedForNewAPIDetector(t *testing.T) {
	if configCacheSchemaVersion < 8 {
		t.Errorf("configCacheSchemaVersion = %d, want >= 8 — escalation gained new_api_detector without a cache schema bump", configCacheSchemaVersion)
	}
}
