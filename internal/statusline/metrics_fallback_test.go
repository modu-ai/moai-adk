package statusline

import (
	"testing"
)

// TestCollectMetrics_ModelFallbackChain tests AC-SF-003: Model Name Env Fallback.
// When stdin JSON has no model field and MOAI_LAST_MODEL env is set,
// CollectMetrics shall return the model from env.
//
// RED Phase: This test should FAIL because CollectMetrics doesn't check env yet.
func TestCollectMetrics_ModelFallbackChain(t *testing.T) {
	tests := []struct {
		name          string
		input         *StdinData
		envModel      string // MOAI_LAST_MODEL value
		cacheModel    string // cached model (for homeDir setup)
		wantAvailable bool
		wantModel     string
	}{
		{
			name: "AC-SF-003: stdin has model → use stdin model",
			input: &StdinData{
				Model: &ModelInfo{DisplayName: "Opus"},
			},
			envModel:      "Sonnet 4",
			wantAvailable: true,
			wantModel:     "Opus", // stdin priority
		},
		{
			name:          "AC-SF-003: no stdin model + MOAI_LAST_MODEL env → use env",
			input:         &StdinData{},
			envModel:      "claude-opus-4-7-20250514",
			wantAvailable: true,
			wantModel:     "Opus 4.7", // ShortenModelName applied
		},
		{
			name:          "no stdin model, empty env → unavailable",
			input:         &StdinData{},
			envModel:      "", // empty env string
			wantAvailable: false,
			wantModel:     "",
		},
		{
			name:       "AC-SF-004: no stdin model, no env, cache file → use cache",
			input:      &StdinData{},
			envModel:   "", // no env
			cacheModel: "Sonnet 4",
			// cacheModel will be written to tempDir/.moai/state/last-model.txt
			wantAvailable: true,
			wantModel:     "Sonnet 4",
		},
		{
			name:          "all fallbacks exhausted → unavailable",
			input:         &StdinData{},
			envModel:      "",
			cacheModel:    "", // no cache
			wantAvailable: false,
			wantModel:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearGLMEnv(t)

			// Set up environment
			if tt.envModel != "" {
				t.Setenv("MOAI_LAST_MODEL", tt.envModel)
			} else {
				t.Setenv("MOAI_LAST_MODEL", "")
			}

			// Set up cache file if needed
			var homeDir string
			if tt.cacheModel != "" {
				tempDir := t.TempDir()
				homeDir = tempDir
				if err := WriteModelCache(homeDir, tt.cacheModel); err != nil {
					t.Fatalf("failed to write cache: %v", err)
				}
			}

			// Call CollectMetrics with homeDir parameter
			met := CollectMetrics(tt.input, homeDir)

			if met.Available != tt.wantAvailable {
				t.Errorf("Available = %v, want %v", met.Available, tt.wantAvailable)
			}

			if met.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", met.Model, tt.wantModel)
			}
		})
	}
}
