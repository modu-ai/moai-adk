package statusline

import "testing"

func TestCollectMetrics(t *testing.T) {
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
				Model: "claude-sonnet-4",
				Cost:  &CostData{TotalUSD: 0.05, InputTokens: 1000, OutputTokens: 500},
			},
			wantModel: "sonnet-4",
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
				Model: "claude-opus-4-5-20251101",
			},
			wantModel: "opus-4-5",
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
			wantAvail: true,
		},
		{
			name: "zero cost",
			input: &StdinData{
				Model: "claude-haiku-3-5",
				Cost:  &CostData{TotalUSD: 0},
			},
			wantModel: "haiku-3-5",
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
		{"claude-sonnet-4", "sonnet-4"},
		{"claude-opus-4-5-20251101", "opus-4-5"},
		{"claude-haiku-3-5", "haiku-3-5"},
		{"claude-sonnet-4-20250514", "sonnet-4"},
		{"gpt-4", "gpt-4"},
		{"", ""},
		{"claude-", ""},
		{"claude-opus-4-5", "opus-4-5"},
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
