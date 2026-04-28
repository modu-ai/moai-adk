// Package cli는 GitHub status 명령에 대한 테스트를 제공합니다.
// Package cli provides tests for GitHub status command.
package cli

import (
	"bytes"
	"testing"
)

func TestNewStatusCmd(t *testing.T) {
	cmd := newStatusCmd()

	if cmd == nil {
		t.Fatal("newStatusCmd returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("expected Use 'status', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRunStatus(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "default status check",
			args:    []string{},
			wantErr: false, // GREEN: 구현 완료
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newStatusCmd()
			cmd.SetArgs(tt.args)

			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(out)

			err := cmd.RunE(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
