package statusline

import "testing"

func TestCollectMemory(t *testing.T) {
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
