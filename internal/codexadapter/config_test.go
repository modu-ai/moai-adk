package codexadapter

import "testing"

// measuredGoodConfig is the shape observed loading and firing under
// codex-cli 0.147.0 — no top-level version, matcher optional, timeout present.
const measuredGoodConfig = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {"type": "command", "command": "/path/to/hook.sh", "timeout": 10}
        ]
      }
    ]
  }
}`

// TestValidatorRejectsVersionKey — AC-REQ-5a.
//
// Single-variable A/B: the identical file fired without this key and fired
// nothing with it, and Codex reported the parse failure only inside the --json
// stream while still exiting 0.
func TestValidatorRejectsVersionKey(t *testing.T) {
	t.Parallel()

	violations, err := ValidateConfig([]byte(`{"version":1,"hooks":{}}`))
	if err != nil {
		t.Fatalf("ValidateConfig error = %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("violations = %+v, want exactly 1", violations)
	}
	if violations[0].Key != "version" {
		t.Errorf("violation key = %q, want version", violations[0].Key)
	}
	if violations[0].Level != "top-level" {
		t.Errorf("violation level = %q, want top-level", violations[0].Level)
	}
}

// TestValidatorAcceptsMeasuredGoodShape — AC-REQ-5b.
func TestValidatorAcceptsMeasuredGoodShape(t *testing.T) {
	t.Parallel()

	violations, err := ValidateConfig([]byte(measuredGoodConfig))
	if err != nil {
		t.Fatalf("ValidateConfig error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none for the measured-working shape", violations)
	}
}

// TestValidatorAcceptsDescription asserts the second accepted top-level key,
// named by the same Codex error message that named `hooks`.
func TestValidatorAcceptsDescription(t *testing.T) {
	t.Parallel()

	violations, err := ValidateConfig([]byte(`{"description":"probe hooks","hooks":{}}`))
	if err != nil {
		t.Fatalf("ValidateConfig error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %+v, want none", violations)
	}
}

// TestValidatorRejectsPerLevelUnknownKeys — AC-REQ-5b.
func TestValidatorRejectsPerLevelUnknownKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		config    string
		wantLevel string
		wantKey   string
	}{
		{
			name:      "top level",
			config:    `{"hooks":{},"schemaVersion":2}`,
			wantLevel: "top-level",
			wantKey:   "schemaVersion",
		},
		{
			name:      "entry level",
			config:    `{"hooks":{"PreToolUse":[{"matcher":"Bash","when":"always","hooks":[]}]}}`,
			wantLevel: "entry",
			wantKey:   "when",
		},
		{
			name:      "hook level",
			config:    `{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"x","timeoutSec":10}]}]}}`,
			wantLevel: "hook",
			wantKey:   "timeoutSec",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations, err := ValidateConfig([]byte(tc.config))
			if err != nil {
				t.Fatalf("ValidateConfig error = %v", err)
			}
			if len(violations) != 1 {
				t.Fatalf("violations = %+v, want exactly 1", violations)
			}
			if violations[0].Level != tc.wantLevel || violations[0].Key != tc.wantKey {
				t.Errorf("violation = %+v, want level %q key %q", violations[0], tc.wantLevel, tc.wantKey)
			}
		})
	}
}

// TestValidatorNamesKeyAndLevelInMessage keeps the diagnostic actionable.
func TestValidatorNamesKeyAndLevelInMessage(t *testing.T) {
	t.Parallel()

	violations, err := ValidateConfig([]byte(`{"version":1}`))
	if err != nil {
		t.Fatalf("ValidateConfig error = %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("violations = %+v, want 1", violations)
	}
	msg := violations[0].Error()
	if !contains(msg, "version") || !contains(msg, "top-level") {
		t.Fatalf("message %q does not name both the key and its level", msg)
	}
}

// TestValidatorReportsMalformedConfig asserts a parse failure is an error
// rather than a silent pass.
func TestValidatorReportsMalformedConfig(t *testing.T) {
	t.Parallel()

	if _, err := ValidateConfig([]byte(`{not json`)); err == nil {
		t.Fatal("malformed config returned nil error; want a parse failure")
	}
}
