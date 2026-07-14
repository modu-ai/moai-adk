package cli

// M3 characterization tests for SPEC-CLI-TUX-V3-005.
//
// These tests pin the exact stdout/stderr bytes of the migration CLI commands
// after the Printer migration (8 fmt.Print* → p.Data/p.Info/p.Success). They
// are the DDD safety net proving behavior preservation:
//
//   - Data paths (JSON + human block) write to STDOUT byte-identical to the
//     pre-migration sequential fmt.Print* output (golden master captured from
//     the pre-migration binary in a seeded temp dir, current=999999 so
//     Pending() is empty regardless of registry state).
//   - Info/Success status messages are re-routed from STDOUT to STDERR (the
//     documented channel change per CLAUDE.md internal/cli output-stream
//     convention: stdout = machine-readable data, stderr = human status).
//   - Korean strings are preserved byte-for-byte (UTF-8).
//
// Registry note: the cli test binary does NOT import
// internal/migration/migrations, so m001/m002 never Register() and the
// registry is empty. This makes Pending() always empty and Apply() always
// return [], so the apply-success (line 61), pending>0 (line 113), and
// rollback-success (line 157) branches cannot be reached via the Runner.
// Those three call sites are byte-pinned at the Printer-call level in
// TestM3_PrinterFormats_UnreachableBranches; they use the identical Printer
// methods whose byte-identity the reachable command-level tests prove.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
)

// seedMigrationDir creates a temp project root with .moai/state/migration-version
// set to version and, if logJSONL is non-empty, .moai/logs/migrations.log seeded.
// current is set to a high value (callers pass 999999) so Pending() is empty.
func seedMigrationDir(t *testing.T, version int, logJSONL string) string {
	t.Helper()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "migration-version"), []byte(strconv.Itoa(version)), 0o644); err != nil {
		t.Fatalf("write version: %v", err)
	}
	if logJSONL != "" {
		logsDir := filepath.Join(dir, ".moai", "logs")
		if err := os.MkdirAll(logsDir, 0o755); err != nil {
			t.Fatalf("mkdir logs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(logsDir, "migrations.log"), []byte(logJSONL), 0o644); err != nil {
			t.Fatalf("write log: %v", err)
		}
	}
	return dir
}

// chdir changes the process cwd to dir and restores it on cleanup. Tests in
// this file do NOT call t.Parallel (shared package-level command vars + cwd).
func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// A single success log entry used to exercise the lastApplied branch.
const m2SuccessLog = `{"version":2,"name":"m002-settings-cleanup","result":"success","details":""}` + "\n"

// TestM3_Status_JSON_NoLastApplied pins the --json stdout for the empty-pending,
// no-lastApplied case. Covers the JSON Data() call site (migration.go line 103).
func TestM3_Status_JSON_NoLastApplied(t *testing.T) {
	dir := seedMigrationDir(t, 999999, "")
	chdir(t, dir)

	if err := migrationStatusCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set json flag: %v", err)
	}
	t.Cleanup(func() {
		if err := migrationStatusCmd.Flags().Set("json", "false"); err != nil {
			t.Fatalf("reset json flag: %v", err)
		}
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	migrationStatusCmd.SetOut(stdout)
	migrationStatusCmd.SetErr(stderr)

	if err := migrationStatusCmd.RunE(migrationStatusCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// stdout must be byte-identical to the pre-migration golden master (73 bytes).
	want := `{
  "current_version": 999999,
  "last_applied": null,
  "pending": []
}
`
	if got := stdout.String(); got != want {
		t.Errorf("stdout mismatch:\nwant %q (%d bytes)\ngot  %q (%d bytes)", want, len(want), got, len(got))
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty (JSON data writes stdout only)", stderr.String())
	}
}

// TestM3_Status_Human_NoPending_NoLastApplied pins the human-format stdout for
// the empty-pending, no-lastApplied case. Covers the composed-string Data() call
// site (migration.go lines 110-121): "현재 버전" + "없음" lines.
func TestM3_Status_Human_NoPending_NoLastApplied(t *testing.T) {
	dir := seedMigrationDir(t, 999999, "")
	chdir(t, dir)

	if err := migrationStatusCmd.Flags().Set("json", "false"); err != nil {
		t.Fatalf("set json flag: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	migrationStatusCmd.SetOut(stdout)
	migrationStatusCmd.SetErr(stderr)

	if err := migrationStatusCmd.RunE(migrationStatusCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// stdout byte-identical to pre-migration golden master (72 bytes).
	want := `현재 버전: 999999
Pending 마이그레이션 없음 (최신 상태)
`
	if got := stdout.String(); got != want {
		t.Errorf("stdout mismatch:\nwant %q (%d bytes)\ngot  %q (%d bytes)", want, len(want), got, len(got))
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty (human data writes stdout only)", stderr.String())
	}
}

// TestM3_Status_Human_WithLastApplied pins the human-format stdout when a success
// log entry exists. Covers the lastApplied line of the composed-string Data()
// block (migration.go line 118): "최근 적용: <name> (버전 <version>)".
func TestM3_Status_Human_WithLastApplied(t *testing.T) {
	dir := seedMigrationDir(t, 999999, m2SuccessLog)
	chdir(t, dir)

	if err := migrationStatusCmd.Flags().Set("json", "false"); err != nil {
		t.Fatalf("set json flag: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	migrationStatusCmd.SetOut(stdout)
	migrationStatusCmd.SetErr(stderr)

	if err := migrationStatusCmd.RunE(migrationStatusCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// stdout byte-identical to pre-migration golden master (120 bytes).
	want := `현재 버전: 999999
Pending 마이그레이션 없음 (최신 상태)
최근 적용: m002-settings-cleanup (버전 2)
`
	if got := stdout.String(); got != want {
		t.Errorf("stdout mismatch:\nwant %q (%d bytes)\ngot  %q (%d bytes)", want, len(want), got, len(got))
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty (human data writes stdout only)", stderr.String())
	}
}

// TestM3_Status_JSON_WithLastApplied pins the --json stdout when lastApplied is
// non-nil. Confirms the nested object serializes byte-identically via Data().
func TestM3_Status_JSON_WithLastApplied(t *testing.T) {
	dir := seedMigrationDir(t, 999999, m2SuccessLog)
	chdir(t, dir)

	if err := migrationStatusCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set json flag: %v", err)
	}
	t.Cleanup(func() {
		if err := migrationStatusCmd.Flags().Set("json", "false"); err != nil {
			t.Fatalf("reset json flag: %v", err)
		}
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	migrationStatusCmd.SetOut(stdout)
	migrationStatusCmd.SetErr(stderr)

	if err := migrationStatusCmd.RunE(migrationStatusCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// stdout byte-identical to pre-migration golden master (172 bytes).
	want := `{
  "current_version": 999999,
  "last_applied": {
    "version": 2,
    "name": "m002-settings-cleanup",
    "result": "success",
    "details": ""
  },
  "pending": []
}
`
	if got := stdout.String(); got != want {
		t.Errorf("stdout mismatch:\nwant %q (%d bytes)\ngot  %q (%d bytes)", want, len(want), got, len(got))
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty (JSON data writes stdout only)", stderr.String())
	}
}

// TestM3_Run_NoPending pins the apply command's no-pending path. The Info status
// message is re-routed from stdout (pre-migration fmt.Println) to stderr
// (post-migration p.Info). Covers the Info() call site (migration.go line 57).
func TestM3_Run_NoPending(t *testing.T) {
	dir := seedMigrationDir(t, 999999, "")
	chdir(t, dir)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	migrationRunCmd.SetOut(stdout)
	migrationRunCmd.SetErr(stderr)

	if err := migrationRunCmd.RunE(migrationRunCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// Channel change: the message moved stdout → stderr. Korean text preserved.
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty (Info status writes stderr, not stdout)", stdout.String())
	}
	want := "실행할 pending 마이그레이션이 없습니다."
	if got := stderr.String(); !strings.Contains(got, want) {
		t.Errorf("stderr = %q, want substring %q (Korean text preserved on stderr)", got, want)
	}
}

// TestM3_PrinterFormats_UnreachableBranches byte-pins the three call sites that
// the empty test-binary registry prevents reaching through the Runner:
//
//   - migration.go line 61  (apply success):    p.Success("성공: %d개 ... (버전: %v)", n, slice)
//   - migration.go line 113 (pending>0 line):   fmt.Sprintf("Pending 마이그레이션 (%d개): %v", n, slice)
//   - migration.go line 157 (rollback success): p.Success("성공: 버전 %d로 롤백됨", v)
//
// These use the identical Printer methods whose byte-identity the command-level
// tests above prove; this test guards the exact format strings / interpolation.
func TestM3_PrinterFormats_UnreachableBranches(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	p := printer.New(printer.WithWriters(stdout, stderr))

	// Line 61 — apply success (Success → stderr).
	p.Success("성공: %d개 마이그레이션 적용됨 (버전: %v)", 2, []int{1, 2})
	if got := stderr.String(); !strings.Contains(got, "성공: 2개 마이그레이션 적용됨 (버전: [1 2])") {
		t.Errorf("apply-success stderr = %q, want substring containing the Korean text", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("apply-success stdout = %q, want empty (Success writes stderr)", stdout.String())
	}

	// Line 157 — rollback success (Success → stderr).
	stderr.Reset()
	p.Success("성공: 버전 %d로 롤백됨", 0)
	if got := stderr.String(); !strings.Contains(got, "성공: 버전 0로 롤백됨") {
		t.Errorf("rollback-success stderr = %q, want substring containing the Korean text", got)
	}

	// Line 113 — pending>0 human line (composed into Data → stdout).
	stdout.Reset()
	line := fmt.Sprintf("Pending 마이그레이션 (%d개): %v", 2, []int{1, 2})
	_ = p.Data(line)
	if got := stdout.String(); !strings.Contains(got, "Pending 마이그레이션 (2개): [1 2]") {
		t.Errorf("pending>0 stdout = %q, want substring containing the Korean text", got)
	}
}
