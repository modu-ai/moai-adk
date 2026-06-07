package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestFieldsetStatuslineConditionalSegments verifies REQ-WC8-010 (SLR-3 UX): the
// 15 segment checkboxes are editable only when the preset is "custom"; for a
// non-custom preset (full/compact/minimal) they render display-only (disabled),
// because the server expands the preset into the segments map at save time.
func TestFieldsetStatuslineConditionalSegments(t *testing.T) {
	tests := []struct {
		name         string
		preset       string
		wantDisabled bool
	}{
		{name: "custom preset: segments editable", preset: "custom", wantDisabled: false},
		{name: "full preset: segments display-only", preset: "full", wantDisabled: true},
		{name: "compact preset: segments display-only", preset: "compact", wantDisabled: true},
		{name: "minimal preset: segments display-only", preset: "minimal", wantDisabled: true},
		{name: "empty preset: segments display-only", preset: "", wantDisabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := renderIndexBody(t, profile.ProfilePreferences{StatuslinePreset: tt.preset})

			// Count segment_<key> checkbox inputs and how many carry `disabled`.
			segInputs := strings.Count(body, `name="segment_`)
			if segInputs != 15 {
				t.Fatalf("expected 15 segment_<key> inputs, got %d", segInputs)
			}

			// A disabled segment checkbox renders `disabled` on the input. Count
			// disabled occurrences within the segments markup.
			disabledSegInputs := strings.Count(body, ` value="1" disabled`) +
				strings.Count(body, ` checked disabled`)

			if tt.wantDisabled {
				if disabledSegInputs != 15 {
					t.Errorf("preset %q: expected all 15 segment checkboxes disabled, got %d", tt.preset, disabledSegInputs)
				}
				if !strings.Contains(body, "segments--disabled") {
					t.Errorf("preset %q: expected #custom-segments to carry segments--disabled class", tt.preset)
				}
			} else {
				if disabledSegInputs != 0 {
					t.Errorf("preset=custom: expected 0 disabled segment checkboxes, got %d (segments must be editable)", disabledSegInputs)
				}
				if strings.Contains(body, "segments--disabled") {
					t.Errorf("preset=custom: #custom-segments must NOT carry segments--disabled class")
				}
			}
		})
	}
}
