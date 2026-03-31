package statusline

import "testing"

// clearGLMEnv clears GLM environment variables so tests run in isolation
// regardless of whether GLM mode is active in the developer's shell.
func clearGLMEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")
}

func TestCollectMetrics(t *testing.T) {
	clearGLMEnv(t)

	tests := []struct {
		name      string
		input     *StdinData
		wantModel string
		wantCost  float64
		wantAvail bool
	}{
		{
			name: "valid cost and model data",
			input: &StdinData{
				Model: &ModelInfo{Name: "claude-sonnet-4-20250514"},
				Cost:  &CostData{TotalUSD: 0.05, InputTokens: 1000, OutputTokens: 500},
			},
			wantModel: "Sonnet 4",
			wantCost:  0.05,
			wantAvail: true,
		},
		{
			name:      "nil input",
			input:     nil,
			wantModel: "",
			wantCost:  0,
			wantAvail: false,
		},
		{
			name: "nil cost",
			input: &StdinData{
				Model: &ModelInfo{Name: "claude-opus-4-5-20251101"},
			},
			wantModel: "Opus 4.5",
			wantCost:  0,
			wantAvail: true,
		},
		{
			name: "empty model",
			input: &StdinData{
				Cost: &CostData{TotalUSD: 0.15},
			},
			wantModel: "",
			wantCost:  0.15,
			wantAvail: false,
		},
		{
			name: "zero cost",
			input: &StdinData{
				Model: &ModelInfo{Name: "claude-haiku-3-5-20241022"},
				Cost:  &CostData{TotalUSD: 0},
			},
			wantModel: "Haiku 3.5",
			wantCost:  0,
			wantAvail: true,
		},
		{
			name: "display name takes priority over name",
			input: &StdinData{
				Model: &ModelInfo{DisplayName: "Opus 4.5", Name: "claude-opus-4-5-20251101"},
			},
			wantModel: "Opus 4.5",
			wantCost:  0,
			wantAvail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectMetrics(tt.input)

			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
			if got.CostUSD != tt.wantCost {
				t.Errorf("CostUSD = %f, want %f", got.CostUSD, tt.wantCost)
			}
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
		})
	}
}

func TestShortenModelName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"claude-sonnet-4-20250514", "Sonnet 4"},
		{"claude-opus-4-5-20251101", "Opus 4.5"},
		{"claude-haiku-3-5-20241022", "Haiku 3.5"},
		{"claude-sonnet-4", "Sonnet 4"},
		{"claude-opus-4-5", "Opus 4.5"},
		{"gpt-4", "gpt-4"},
		{"", ""},
		{"claude-", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ShortenModelName(tt.input)
			if got != tt.want {
				t.Errorf("ShortenModelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCollectMetrics_SessionDuration(t *testing.T) {
	tests := []struct {
		name   string
		input  *StdinData
		wantMS int
	}{
		{
			name:   "extracts duration",
			input:  &StdinData{Cost: &CostData{TotalDurationMS: 4980000}, Model: &ModelInfo{DisplayName: "Opus"}},
			wantMS: 4980000,
		},
		{
			name:   "zero duration",
			input:  &StdinData{Cost: &CostData{TotalDurationMS: 0}, Model: &ModelInfo{DisplayName: "Opus"}},
			wantMS: 0,
		},
		{
			name:   "nil cost",
			input:  &StdinData{Model: &ModelInfo{DisplayName: "Opus"}},
			wantMS: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := CollectMetrics(tt.input)
			if m.SessionDurationMS != tt.wantMS {
				t.Errorf("SessionDurationMS = %d, want %d", m.SessionDurationMS, tt.wantMS)
			}
		})
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.05, "$0.05"},
		{0.15, "$0.15"},
		{1.234, "$1.23"},
		{0, "$0.00"},
		{10.5, "$10.50"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatCost(tt.input)
			if got != tt.want {
				t.Errorf("formatCost(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestResolveGLMModelName verifies GLM model name detection from env vars.
func TestResolveGLMModelName(t *testing.T) {
	clearGLMEnv(t)

	tests := []struct {
		name        string
		displayName string
		envKey      string
		envValue    string
		want        string
	}{
		{
			name:        "Opus display name with GLM env",
			displayName: "Opus",
			envKey:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			envValue:    "glm-5.1",
			want:        "glm-5.1",
		},
		{
			name:        "Opus 4.6 display name with GLM env",
			displayName: "Opus 4.6",
			envKey:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			envValue:    "glm-5.1",
			want:        "glm-5.1",
		},
		{
			name:        "Sonnet display name with GLM env",
			displayName: "Sonnet",
			envKey:      "ANTHROPIC_DEFAULT_SONNET_MODEL",
			envValue:    "glm-4.7",
			want:        "glm-4.7",
		},
		{
			name:        "Haiku display name with GLM env",
			displayName: "Haiku",
			envKey:      "ANTHROPIC_DEFAULT_HAIKU_MODEL",
			envValue:    "glm-4.7-air",
			want:        "glm-4.7-air",
		},
		{
			name:        "Opus without GLM env (Claude mode)",
			displayName: "Opus",
			envKey:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			envValue:    "",
			want:        "Opus",
		},
		{
			name:        "Opus with Claude model env (not GLM)",
			displayName: "Opus",
			envKey:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			envValue:    "claude-opus-4-6-20250514",
			want:        "Opus",
		},
		{
			name:        "Non-Claude display name unchanged",
			displayName: "GPT-4o",
			envKey:      "",
			envValue:    "",
			want:        "GPT-4o",
		},
		{
			name:        "Empty display name",
			displayName: "",
			envKey:      "",
			envValue:    "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}
			got := resolveGLMModelName(tt.displayName)
			if got != tt.want {
				t.Errorf("resolveGLMModelName(%q) = %q, want %q", tt.displayName, got, tt.want)
			}
		})
	}
}

// TestCollectMetrics_GLMMode verifies CollectMetrics shows GLM model name in GLM mode.
func TestCollectMetrics_GLMMode(t *testing.T) {
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.1")

	input := &StdinData{
		Model: &ModelInfo{DisplayName: "Opus"},
	}
	got := CollectMetrics(input)
	if got.Model != "glm-5.1" {
		t.Errorf("Model = %q in GLM mode, want %q", got.Model, "glm-5.1")
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{50000, "50K"},
		{200000, "200K"},
		{1000, "1K"},
		{500, "500"},
		{0, "0"},
		{999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatTokens(tt.input)
			if got != tt.want {
				t.Errorf("formatTokens(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
