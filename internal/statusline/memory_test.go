package statusline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectMemory(t *testing.T) {
	// Disable auto-compact scaling for existing tests
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "100")

	tests := []struct {
		name       string
		input      *StdinData
		wantUsed   int
		wantBudget int
		wantAvail  bool
	}{
		{
			name: "valid context window data (legacy fields)",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
			},
			wantUsed:   50000,
			wantBudget: 200000,
			wantAvail:  true,
		},
		{
			name:       "nil input",
			input:      nil,
			wantUsed:   0,
			wantBudget: 0,
			wantAvail:  false,
		},
		{
			name:       "nil context window",
			input:      &StdinData{Model: &ModelInfo{Name: "claude-sonnet-4"}},
			wantUsed:   0,
			wantBudget: 0,
			wantAvail:  false,
		},
		{
			name: "zero values - session start state",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{Used: 0, Total: 0},
			},
			wantUsed:   0,
			wantBudget: 200000, // Default context window size
			wantAvail:  true,
		},
		{
			name: "full context window (legacy fields)",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{Used: 200000, Total: 200000},
			},
			wantUsed:   200000,
			wantBudget: 200000,
			wantAvail:  true,
		},
		{
			name: "used_percentage takes priority",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(25.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:   50000, // 25% of 200000
			wantBudget: 200000,
			wantAvail:  true,
		},
		{
			name: "current_usage calculation",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					ContextWindowSize: 200000,
					CurrentUsage: &CurrentUsageInfo{
						InputTokens:         30000,
						CacheCreationTokens: 10000,
						CacheReadTokens:     10000,
					},
				},
			},
			wantUsed:   50000, // 30000 + 10000 + 10000
			wantBudget: 200000,
			wantAvail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectMemory(tt.input)

			if got.TokensUsed != tt.wantUsed {
				t.Errorf("TokensUsed = %d, want %d", got.TokensUsed, tt.wantUsed)
			}
			if got.TokenBudget != tt.wantBudget {
				t.Errorf("TokenBudget = %d, want %d", got.TokenBudget, tt.wantBudget)
			}
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
		})
	}
}

// TestCollectMemory_AutoCompactScaling verifies that TokenBudget is scaled
// to the auto-compact threshold so the CW bar shows 100% at compact point.
func TestCollectMemory_AutoCompactScaling(t *testing.T) {
	tests := []struct {
		name          string
		threshold     string // CLAUDE_AUTOCOMPACT_PCT_OVERRIDE value
		input         *StdinData
		wantUsed      int
		wantBudget    int
		wantPctApprox int // expected usagePercent (approximate)
	}{
		{
			name:      "default threshold 85%: 83% used → ~97% display",
			threshold: "85",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(83.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:      166000, // 83% of 200K
			wantBudget:    170000, // 85% of 200K
			wantPctApprox: 97,     // 166000/170000 = 97.6%
		},
		{
			name:      "default threshold 85%: 85% used → 100% display",
			threshold: "85",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(85.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:      170000, // 85% of 200K
			wantBudget:    170000, // 85% of 200K
			wantPctApprox: 100,    // exact 100%
		},
		{
			name:      "threshold 90%: 83% used → 92% display",
			threshold: "90",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(83.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:      166000,
			wantBudget:    180000, // 90% of 200K
			wantPctApprox: 92,     // 166000/180000 = 92.2%
		},
		{
			name:      "threshold 100% (no scaling)",
			threshold: "100",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(83.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:      166000,
			wantBudget:    200000,
			wantPctApprox: 83,
		},
		{
			name:      "exceeded threshold capped at 100%",
			threshold: "80",
			input: &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    new(85.0),
					ContextWindowSize: 200000,
				},
			},
			wantUsed:      170000,
			wantBudget:    160000, // 80% of 200K
			wantPctApprox: 100,    // capped at 100%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", tt.threshold)

			got := CollectMemory(tt.input)

			if got.TokensUsed != tt.wantUsed {
				t.Errorf("TokensUsed = %d, want %d", got.TokensUsed, tt.wantUsed)
			}
			if got.TokenBudget != tt.wantBudget {
				t.Errorf("TokenBudget = %d, want %d", got.TokenBudget, tt.wantBudget)
			}

			pct := usagePercent(got.TokensUsed, got.TokenBudget)
			if pct != tt.wantPctApprox {
				t.Errorf("usagePercent = %d%%, want ~%d%%", pct, tt.wantPctApprox)
			}
		})
	}
}

func TestContextUsageLevel(t *testing.T) {
	tests := []struct {
		name  string
		used  int
		total int
		want  contextLevel
	}{
		{"green - low usage 25%", 50000, 200000, levelOk},
		{"green - zero usage", 0, 200000, levelOk},
		{"green - 49%", 98000, 200000, levelOk},
		{"yellow - exactly 50%", 100000, 200000, levelWarn},
		{"yellow - 65%", 130000, 200000, levelWarn},
		{"yellow - 79%", 158000, 200000, levelWarn},
		{"red - exactly 80%", 160000, 200000, levelError},
		{"red - 90%", 180000, 200000, levelError},
		{"red - 100%", 200000, 200000, levelError},
		{"green - zero total", 0, 0, levelOk},
		{"green - negative total", 100, -1, levelOk},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contextUsageLevel(tt.used, tt.total)
			if got != tt.want {
				t.Errorf("contextUsageLevel(%d, %d) = %d, want %d",
					tt.used, tt.total, got, tt.want)
			}
		})
	}
}

func TestUsagePercent(t *testing.T) {
	tests := []struct {
		name  string
		used  int
		total int
		want  int
	}{
		{"25%", 50000, 200000, 25},
		{"50%", 100000, 200000, 50},
		{"100%", 200000, 200000, 100},
		{"0%", 0, 200000, 0},
		{"zero total", 100, 0, 0},
		{"negative total", 100, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usagePercent(tt.used, tt.total)
			if got != tt.want {
				t.Errorf("usagePercent(%d, %d) = %d, want %d",
					tt.used, tt.total, got, tt.want)
			}
		})
	}
}

// TestCollectMemory_GLMContextOverride verifies that when a GLM model is
// active (via ANTHROPIC_DEFAULT_*_MODEL env), the context window size is
// overridden from the Claude slot's reported value (e.g., 1M) to the actual
// GLM model's limit. Issue #653.
func TestCollectMemory_GLMContextOverride(t *testing.T) {
	pct := 23.0 // near the reported ~23% context-full point on 1M gauge

	cases := []struct {
		name         string
		envSlot      string
		modelName    string
		reportedSize int
		wantBudget   int // expected TokenBudget (contextSize * 85 / 100)
	}{
		{
			name:         "glm-5.1 opus slot overrides 1M to 200K",
			envSlot:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			modelName:    "glm-5.1",
			reportedSize: 1_000_000,
			wantBudget:   200_000 * 85 / 100,
		},
		{
			name:         "glm-4.5-air haiku slot",
			envSlot:      "ANTHROPIC_DEFAULT_HAIKU_MODEL",
			modelName:    "glm-4.5-air",
			reportedSize: 200_000,
			wantBudget:   128_000 * 85 / 100,
		},
		{
			name:         "claude model preserves reported size",
			envSlot:      "ANTHROPIC_DEFAULT_OPUS_MODEL",
			modelName:    "claude-opus-4-6",
			reportedSize: 1_000_000,
			wantBudget:   1_000_000 * 85 / 100,
		},
		{
			name:         "unknown non-claude model falls back to reported",
			envSlot:      "ANTHROPIC_DEFAULT_SONNET_MODEL",
			modelName:    "custom-unknown-llm",
			reportedSize: 200_000,
			wantBudget:   200_000 * 85 / 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear all model slots so only tc.envSlot drives detection.
			t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
			t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
			t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")
			t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
			t.Setenv(tc.envSlot, tc.modelName)

			input := &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    &pct,
					ContextWindowSize: tc.reportedSize,
				},
			}
			got := CollectMemory(input)
			if got == nil || !got.Available {
				t.Fatalf("CollectMemory returned unavailable")
			}
			if got.TokenBudget != tc.wantBudget {
				t.Errorf("TokenBudget = %d, want %d", got.TokenBudget, tc.wantBudget)
			}
		})
	}
}

// TestCollectMemory_ExplicitOverride verifies that MOAI_STATUSLINE_CONTEXT_SIZE
// wins over both Claude reported size and GLM auto-detection (issue #653).
func TestCollectMemory_ExplicitOverride(t *testing.T) {
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.1")
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "131072")

	pct := 50.0
	input := &StdinData{
		ContextWindow: &ContextWindowInfo{
			UsedPercentage:    &pct,
			ContextWindowSize: 1_000_000,
		},
	}
	got := CollectMemory(input)
	if got == nil || !got.Available {
		t.Fatalf("CollectMemory returned unavailable")
	}
	wantBudget := 131072 * 85 / 100
	if got.TokenBudget != wantBudget {
		t.Errorf("TokenBudget = %d, want %d (explicit override wins)", got.TokenBudget, wantBudget)
	}
}

// TestCollectMemory_LLMYAMLOverride verifies that .moai/config/sections/llm.yaml
// `glm.context_windows` entries take precedence over the built-in
// glmContextWindows table when MOAI_STATUSLINE_CONTEXT_SIZE is not set.
// Priority: env > llm.yaml > built-in table (issue #653).
func TestCollectMemory_LLMYAMLOverride(t *testing.T) {
	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yamlBody := `llm:
  glm:
    context_windows:
      glm-5.1: 230000        # user-configured override (larger than built-in 200K)
      custom-model: 96000    # previously unknown model
`
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	// resolveContextWindowOverride walks from getwd() upward. t.Chdir restores
	// the original wd at test teardown.
	t.Chdir(tempDir)
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	cases := []struct {
		name       string
		glmModel   string
		wantBudget int
	}{
		{
			name:       "llm.yaml overrides built-in for glm-5.1 (200K→230K)",
			glmModel:   "glm-5.1",
			wantBudget: 230_000 * 85 / 100,
		},
		{
			name:       "llm.yaml adds previously unknown custom-model",
			glmModel:   "custom-model",
			wantBudget: 96_000 * 85 / 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", tc.glmModel)

			pct := 50.0
			input := &StdinData{
				ContextWindow: &ContextWindowInfo{
					UsedPercentage:    &pct,
					ContextWindowSize: 1_000_000,
				},
			}
			got := CollectMemory(input)
			if got == nil || !got.Available {
				t.Fatalf("CollectMemory returned unavailable")
			}
			if got.TokenBudget != tc.wantBudget {
				t.Errorf("TokenBudget = %d, want %d", got.TokenBudget, tc.wantBudget)
			}
		})
	}
}

// TestCollectMemory_EnvOverridesLLMYAML verifies env var wins over llm.yaml.
func TestCollectMemory_EnvOverridesLLMYAML(t *testing.T) {
	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	_ = os.MkdirAll(sectionsDir, 0o755)
	yamlBody := `llm:
  glm:
    context_windows:
      glm-5.1: 230000
`
	_ = os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(yamlBody), 0o644)

	t.Chdir(tempDir)
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.1")
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "65536")

	pct := 50.0
	input := &StdinData{
		ContextWindow: &ContextWindowInfo{
			UsedPercentage:    &pct,
			ContextWindowSize: 1_000_000,
		},
	}
	got := CollectMemory(input)
	if got == nil || !got.Available {
		t.Fatalf("CollectMemory returned unavailable")
	}
	wantBudget := 65536 * 85 / 100
	if got.TokenBudget != wantBudget {
		t.Errorf("TokenBudget = %d, want %d (env wins over llm.yaml)", got.TokenBudget, wantBudget)
	}
}
