package cli

import (
	"testing"
)

// TestSearchCmd_HasSubcommand은 search 커맨드가 루트에 등록되어 있는지 확인한다.
func TestSearchCmd_HasSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'search' 서브커맨드가 rootCmd에 등록되지 않았다")
	}
}

// TestSearchCmd_FlagParsing은 search 커맨드에 모든 플래그가 정의되어 있는지 확인한다.
func TestSearchCmd_FlagParsing(t *testing.T) {
	// search 커맨드 찾기
	var cmd = searchCmd

	tests := []struct {
		flagName string
	}{
		{"branch"},
		{"since"},
		{"until"},
		{"role"},
		{"limit"},
		{"index-session"},
		{"project-path"},
		{"git-branch"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			f := cmd.Flags().Lookup(tt.flagName)
			if f == nil {
				t.Errorf("플래그 '--%s'가 search 커맨드에 없다", tt.flagName)
			}
		})
	}
}
