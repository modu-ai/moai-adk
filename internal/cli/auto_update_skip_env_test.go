package cli

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// TestBuildAutoUpdateFunc_HonorsSkipBinaryUpdateEnv pins GitHub #1714: the
// SessionStart auto-update must honor MOAI_SKIP_BINARY_UPDATE=1 the same way
// `moai update` does. The unset arm is the control — it proves the fixture
// reaches the orchestrator, so the skip arm's silence is not vacuous.
func TestBuildAutoUpdateFunc_HonorsSkipBinaryUpdateEnv(t *testing.T) {
	cases := []struct {
		name        string
		skip        string
		wantUpdated bool
		wantChecked bool
	}{
		{name: "unset updates", skip: "", wantUpdated: true, wantChecked: true},
		{name: "set skips", skip: "1", wantUpdated: false, wantChecked: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origDeps := deps
			defer func() { deps = origDeps }()
			origVersion := version.Version
			defer func() { version.Version = origVersion }()

			// Unique release-shaped version and an empty HOME keep the
			// update cache from short-circuiting either arm.
			version.Version = fmt.Sprintf("v97.97.%d", time.Now().UnixNano()%10000)
			t.Setenv("HOME", t.TempDir())
			t.Setenv(config.EnvSkipBinaryUpdate, tc.skip)

			checked, replaced := false, false
			deps = &Dependencies{
				UpdateChecker: &mockUpdateChecker{
					isUpdateAvailFunc: func(string) (bool, *update.VersionInfo, error) {
						checked = true
						return true, &update.VersionInfo{Version: "v100.0.0"}, nil
					},
				},
				UpdateOrch: &mockUpdateOrchestrator{
					updateFunc: func(context.Context) (*update.UpdateResult, error) {
						replaced = true
						return &update.UpdateResult{PreviousVersion: "v97.97.0", NewVersion: "v100.0.0"}, nil
					},
				},
			}

			result, err := buildAutoUpdateFunc()(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != tc.wantUpdated || replaced != tc.wantUpdated {
				t.Errorf("Updated=%v replaced=%v, want %v", result.Updated, replaced, tc.wantUpdated)
			}
			if checked != tc.wantChecked {
				t.Errorf("update check ran=%v, want %v", checked, tc.wantChecked)
			}
		})
	}
}
