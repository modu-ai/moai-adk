// Package runner는 네트워크 egress 검증 기능을 테스트합니다.
// Package runner tests network egress validation functionality.
package runner

import (
	"context"
	"errors"
	"testing"
)

func TestEgressValidator_ValidateGitHubAPI(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() EgressValidator
		wantErr bool
	}{
		{
			name: "valid github api connection",
			setup: func() EgressValidator {
				return &mockEgressValidator{
					validateAPIFunc: func(ctx context.Context) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "github api connection failed",
			setup: func() EgressValidator {
				return &mockEgressValidator{
					validateAPIFunc: func(ctx context.Context) error {
						return errors.New("lookup failed")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := tt.setup()
			err := validator.ValidateGitHubAPI(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitHubAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEgressValidator_ValidateRunnerDownload(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() EgressValidator
		wantErr bool
	}{
		{
			name: "valid runner download connection",
			setup: func() EgressValidator {
				return &mockEgressValidator{
					validateDownloadFunc: func(ctx context.Context) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "runner download connection failed",
			setup: func() EgressValidator {
				return &mockEgressValidator{
					validateDownloadFunc: func(ctx context.Context) error {
						return errors.New("connection refused")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := tt.setup()
			err := validator.ValidateRunnerDownload(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRunnerDownload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockEgressValidator struct {
	validateAPIFunc      func(ctx context.Context) error
	validateDownloadFunc func(ctx context.Context) error
}

func (m *mockEgressValidator) ValidateGitHubAPI(ctx context.Context) error {
	if m.validateAPIFunc != nil {
		return m.validateAPIFunc(ctx)
	}
	return nil
}

func (m *mockEgressValidator) ValidateRunnerDownload(ctx context.Context) error {
	if m.validateDownloadFunc != nil {
		return m.validateDownloadFunc(ctx)
	}
	return nil
}

func TestAuditLogger_LogInstall(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "log install success",
			version: "v2.317.0",
			wantErr: false,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mockAuditLogger{
				logInstallFunc: func(ctx context.Context, version string) error {
					if version == "" {
						return &InvalidVersionError{}
					}
					return nil
				},
			}

			err := logger.LogInstall(context.Background(), tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogInstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuditLogger_LogRegister(t *testing.T) {
	tests := []struct {
		name     string
		runnerID string
		wantErr  bool
	}{
		{
			name:     "log register success",
			runnerID: "runner-123",
			wantErr:  false,
		},
		{
			name:     "empty runner id",
			runnerID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &mockAuditLogger{
				logRegisterFunc: func(ctx context.Context, runnerID string) error {
					if runnerID == "" {
						return &InvalidRunnerIDError{}
					}
					return nil
				},
			}

			err := logger.LogRegister(context.Background(), tt.runnerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LogRegister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockAuditLogger struct {
	logInstallFunc  func(ctx context.Context, version string) error
	logRegisterFunc func(ctx context.Context, runnerID string) error
}

func (m *mockAuditLogger) LogInstall(ctx context.Context, version string) error {
	if m.logInstallFunc != nil {
		return m.logInstallFunc(ctx, version)
	}
	return nil
}

func (m *mockAuditLogger) LogRegister(ctx context.Context, runnerID string) error {
	if m.logRegisterFunc != nil {
		return m.logRegisterFunc(ctx, runnerID)
	}
	return nil
}

type InvalidVersionError struct{}

func (e *InvalidVersionError) Error() string {
	return "invalid version"
}

type InvalidRunnerIDError struct{}

func (e *InvalidRunnerIDError) Error() string {
	return "invalid runner ID"
}
