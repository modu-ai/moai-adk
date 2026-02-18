package rank

import (
	"strings"
	"testing"
)

func TestBrowserWindowsURLEscaping(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single ampersand",
			input: "https://example.com/auth?a=1&b=2",
			want:  "https://example.com/auth?a=1^&b=2",
		},
		{
			name:  "multiple ampersands",
			input: "https://example.com/auth?a=1&b=2&c=3",
			want:  "https://example.com/auth?a=1^&b=2^&c=3",
		},
		{
			name:  "no ampersand",
			input: "https://example.com/auth?code=abc",
			want:  "https://example.com/auth?code=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.ReplaceAll(tt.input, "&", "^&")
			if got != tt.want {
				t.Errorf("URL escaping = %q, want %q", got, tt.want)
			}
		})
	}
}
