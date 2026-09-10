package hook_test

// Session-start migration contracts (SPEC-V3R2-RT-007).
//
// Observation seam: runMigration builds the real migration.NewRunner and there is
// no injection point. This test binary does not link internal/migration/migrations,
// so the registry is empty and Apply's only observable effect is reading
// .moai/state/migration-version. A malformed version file makes Apply return an
// error, which runMigration records as Data["migration_error"] — that key has no
// other writer, so its presence proves the runner was invoked.
//
// Known gaps (reported, not asserted — asserting them would pin the defect):
//   - REQ-V3R2-RT-007-021 requires the failure to surface via SystemMessage. The
//     implementation records it only in HookOutput.Data (json:"-") and slog, so
//     the user never sees it. No test here asserts on SystemMessage.
//   - REQ-V3R2-RT-007-032 names system.yaml as the source of migrations.disabled,
//     but the config loader binds only the hook block of system.yaml, so the key
//     never reaches cfg.System.Migrations. The disabled/enabled tests below drive
//     the flag through ConfigProvider, which is the path the handler reads.
// Evidence: .moai/reports/t617/verdict.md

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// migrationCfgProvider is a fixed ConfigProvider for the migration tests.
type migrationCfgProvider struct{ cfg *config.Config }

func (p migrationCfgProvider) Get() *config.Config { return p.cfg }

// newMigrationFixture returns a project root whose migration-version file holds
// the given content; an empty content leaves the file absent.
func newMigrationFixture(t *testing.T, content string) (root, versionFile string) {
	t.Helper()
	root = t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	versionFile = filepath.Join(stateDir, "migration-version")
	if content != "" {
		if err := os.WriteFile(versionFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, versionFile
}

// runSessionStart invokes the session-start handler and decodes its internal Data map.
func runSessionStart(t *testing.T, provider hook.ConfigProvider, root string) (*hook.HookOutput, map[string]any) {
	t.Helper()
	h := hook.NewSessionStartHandler(provider)
	out, err := h.Handle(context.Background(), &hook.HookInput{
		SessionID:  "t617-migration",
		ProjectDir: root,
	})
	if err != nil {
		t.Fatalf("Handle returned error, session-start must be non-blocking: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}
	data := map[string]any{}
	if err := json.Unmarshal(out.Data, &data); err != nil {
		t.Fatalf("decode HookOutput.Data: %v (raw=%s)", err, out.Data)
	}
	return out, data
}

const malformedVersion = "not-a-number"

// TestSessionStart_InvokesMigrationRunner verifies that the session-start hook calls the runner.
// REQ-V3R2-RT-007-020: the session-start hook calls MigrationRunner.Apply.
func TestSessionStart_InvokesMigrationRunner(t *testing.T) {
	t.Run("malformed version file is read by the runner", func(t *testing.T) {
		root, _ := newMigrationFixture(t, malformedVersion)
		_, data := runSessionStart(t, nil, root)
		if msg, _ := data["migration_error"].(string); msg == "" {
			t.Fatalf("Data[migration_error] absent: runner.Apply was not invoked (Data=%v)", data)
		}
	})
	t.Run("control: absent version file yields no migration_error", func(t *testing.T) {
		root, _ := newMigrationFixture(t, "")
		_, data := runSessionStart(t, nil, root)
		if v, ok := data["migration_error"]; ok {
			t.Fatalf("Data[migration_error]=%v on a clean project; the key must not be unconditional", v)
		}
	})
}

// TestSessionStart_MigrationFailure_DoesNotBlockSession verifies the fail-open half of the contract.
// REQ-V3R2-RT-007-021: a migration failure does not block the session and does not advance
// the version file. The SystemMessage half is unimplemented — see the file header.
func TestSessionStart_MigrationFailure_DoesNotBlockSession(t *testing.T) {
	root, versionFile := newMigrationFixture(t, malformedVersion)
	out, data := runSessionStart(t, nil, root)

	if out.Continue != nil && !*out.Continue {
		t.Fatalf("Continue=false on migration failure; the session must not be blocked (StopReason=%q)", out.StopReason)
	}
	if out.StopReason != "" {
		t.Fatalf("StopReason=%q on migration failure; the session must not be blocked", out.StopReason)
	}
	if msg, _ := data["migration_error"].(string); msg == "" {
		t.Fatalf("Data[migration_error] absent: the failure was swallowed without a record (Data=%v)", data)
	}
	if _, ok := data["migrations_applied"]; ok {
		t.Fatalf("Data[migrations_applied] present on a failed apply (Data=%v)", data)
	}
	got, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("read version file: %v", err)
	}
	if string(got) != malformedVersion {
		t.Fatalf("version file changed on failure: got %q, want %q", got, malformedVersion)
	}
}

// TestSessionStart_MigrationsDisabled_SkipsRunner verifies that migrations.disabled: true skips
// the runner, measured against the enabled input on the identical fixture.
// REQ-V3R2-RT-007-032. The system.yaml loading path is unwired — see the file header.
func TestSessionStart_MigrationsDisabled_SkipsRunner(t *testing.T) {
	disabled := config.NewDefaultConfig()
	disabled.System.Migrations.Disabled = true
	enabled := config.NewDefaultConfig()
	enabled.System.Migrations.Disabled = false

	t.Run("disabled: runner skipped", func(t *testing.T) {
		root, _ := newMigrationFixture(t, malformedVersion)
		_, data := runSessionStart(t, migrationCfgProvider{cfg: disabled}, root)
		if v, ok := data["migration_error"]; ok {
			t.Fatalf("Data[migration_error]=%v with migrations.disabled=true; the runner must be skipped", v)
		}
	})
	t.Run("control: enabled on the same fixture runs the runner", func(t *testing.T) {
		root, _ := newMigrationFixture(t, malformedVersion)
		_, data := runSessionStart(t, migrationCfgProvider{cfg: enabled}, root)
		if msg, _ := data["migration_error"].(string); msg == "" {
			t.Fatalf("Data[migration_error] absent with migrations.disabled=false (Data=%v)", data)
		}
	})
}

// TestSessionStart_EnabledByDefault verifies that migration is enabled by default.
// REQ-V3R2-RT-007-032: when migrations.disabled is missing or false, the runner is called.
func TestSessionStart_EnabledByDefault(t *testing.T) {
	cases := []struct {
		name     string
		provider hook.ConfigProvider
	}{
		{"nil ConfigProvider", nil},
		{"provider returning nil config", migrationCfgProvider{cfg: nil}},
		{"default config (key missing)", migrationCfgProvider{cfg: config.NewDefaultConfig()}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := newMigrationFixture(t, malformedVersion)
			_, data := runSessionStart(t, tc.provider, root)
			if msg, _ := data["migration_error"].(string); msg == "" {
				t.Fatalf("Data[migration_error] absent: runner not invoked by default (Data=%v)", data)
			}
		})
	}
}
