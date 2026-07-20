package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// scheduledManifestJSON returns the valid fixture manifest extended with a
// schedule object (interval/mechanism/mode).
func scheduledManifestJSON(interval, mechanism, mode string) string {
	schedule := `,
  "schedule": {
    "interval": "` + interval + `",
    "mechanism": "` + mechanism + `",
    "mode": "` + mode + `"
  }
}`
	return strings.TrimSuffix(validManifestJSON, "\n}") + schedule
}

// newRootedCmd wires child under a parent carrying the persistent
// --project-root flag (the inheritance path resolveProjectRootV4 consults).
func newRootedCmd(child *cobra.Command, projectRoot string) *cobra.Command {
	parent := &cobra.Command{Use: "harness"}
	parent.PersistentFlags().String("project-root", projectRoot, "")
	parent.AddCommand(child)
	return parent
}

// TestListHarnesses_ScheduleSurfaced verifies a schedule-bearing manifest
// populates HarnessEntry.Schedule and the list --json output includes the
// "schedule" key with its three sub-fields.
func TestListHarnesses_ScheduleSurfaced(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", scheduledManifestJSON("nightly", "cron", "discovery-only"))

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 harness, got %d", len(entries))
	}
	s := entries[0].Schedule
	if s == nil {
		t.Fatalf("HarnessEntry.Schedule not populated from manifest")
	}
	if s.Interval != "nightly" || s.Mechanism != "cron" || s.Mode != "discovery-only" {
		t.Fatalf("Schedule fields wrong: %+v", s)
	}

	// --json output includes "schedule".
	listCmd := NewHarnessV4ListCmd()
	var out bytes.Buffer
	listCmd.SetOut(&out)
	parent := newRootedCmd(listCmd, root)
	parent.SetArgs([]string{"list", "--json"})
	if err := parent.Execute(); err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if !strings.Contains(out.String(), `"schedule"`) {
		t.Fatalf("list --json output missing schedule key: %s", out.String())
	}

	// Human-readable output surfaces interval + mechanism.
	listCmd2 := NewHarnessV4ListCmd()
	var out2 bytes.Buffer
	listCmd2.SetOut(&out2)
	parent2 := newRootedCmd(listCmd2, root)
	parent2.SetArgs([]string{"list"})
	if err := parent2.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out2.String(), "schedule: nightly via cron") {
		t.Fatalf("list text output missing schedule surfacing: %s", out2.String())
	}
}

// TestListHarnesses_ScheduleAbsentOmitted verifies the schedule-less baseline
// is preserved byte-identically: no schedule key in --json, no schedule text
// in the human-readable row.
func TestListHarnesses_ScheduleAbsentOmitted(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", validManifestJSON)

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses error: %v", err)
	}
	if entries[0].Schedule != nil {
		t.Fatalf("schedule-less manifest produced a non-nil Schedule: %+v", entries[0].Schedule)
	}

	listCmd := NewHarnessV4ListCmd()
	var out bytes.Buffer
	listCmd.SetOut(&out)
	parent := newRootedCmd(listCmd, root)
	parent.SetArgs([]string{"list", "--json"})
	if err := parent.Execute(); err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if strings.Contains(out.String(), `"schedule"`) {
		t.Fatalf("list --json leaked a schedule key for schedule-less harness: %s", out.String())
	}

	listCmd2 := NewHarnessV4ListCmd()
	var out2 bytes.Buffer
	listCmd2.SetOut(&out2)
	parent2 := newRootedCmd(listCmd2, root)
	parent2.SetArgs([]string{"list"})
	if err := parent2.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	// Baseline expectation: the row for "dev" carries exactly name + domain +
	// entry with no schedule text (byte-identical pre-change row format).
	wantRow := "dev              moai-adk CLI template development        /harness:dev\n"
	if !strings.Contains(out2.String(), wantRow) {
		t.Fatalf("baseline text row changed for schedule-less harness.\nwant substring: %q\ngot: %s", wantRow, out2.String())
	}
	if strings.Contains(out2.String(), "schedule:") {
		t.Fatalf("schedule text leaked into schedule-less list output: %s", out2.String())
	}
}

// TestRemoveHarness_ScheduleUnregisterNotice verifies `moai harness remove`
// prints an unregister notice naming the declared mechanism, computed from
// the manifest BEFORE deletion; cron names CronDelete, loop names loop
// cancellation; a schedule-less harness prints no notice.
func TestRemoveHarness_ScheduleUnregisterNotice(t *testing.T) {
	cases := []struct {
		name       string
		manifest   string
		wantNotice string
		noNotice   bool
	}{
		{"cron_names_crondelete", scheduledManifestJSON("nightly", "cron", "discovery-only"), "CronDelete", false},
		{"loop_names_cancellation", scheduledManifestJSON("30m", "loop", "discovery-only"), "loop", false},
		{"scheduleless_no_notice", validManifestJSON, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			seedHarness(t, root, "dev", tc.manifest)

			removeCmd := NewHarnessV4RemoveCmd()
			var out bytes.Buffer
			removeCmd.SetOut(&out)
			parent := newRootedCmd(removeCmd, root)
			parent.SetArgs([]string{"remove", "dev"})
			if err := parent.Execute(); err != nil {
				t.Fatalf("remove: %v", err)
			}

			// Removal atomicity unchanged: command file gone.
			cmdPath := filepath.Join(root, ".claude", "commands", "harness", "dev.md")
			if fileExists(cmdPath) {
				t.Fatalf("command file not removed: %s", cmdPath)
			}

			if tc.noNotice {
				if strings.Contains(out.String(), "unregister") || strings.Contains(out.String(), "CronDelete") {
					t.Fatalf("schedule-less remove printed a schedule notice: %s", out.String())
				}
				return
			}
			if !strings.Contains(out.String(), "unregister") {
				t.Fatalf("remove output missing unregister notice: %s", out.String())
			}
			if !strings.Contains(out.String(), tc.wantNotice) {
				t.Fatalf("remove notice missing mechanism token %q: %s", tc.wantNotice, out.String())
			}
		})
	}
}

// fileExists reports whether a path exists (test helper).
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestDoctor_ScheduleInvalidModeError verifies a schema-invalid schedule
// (mode: "write") yields an ERROR-severity DoctorFinding through the doctor
// scan path (axis-2 Validate reuse) and a non-zero-exit report state.
func TestDoctor_ScheduleInvalidModeError(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", scheduledManifestJSON("nightly", "cron", "write"))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor error: %v", err)
	}
	if report.ErrorCount < 1 {
		t.Fatalf("expected >= 1 ERROR finding for invalid schedule mode, got report: %+v", report)
	}
	found := false
	for _, f := range report.Findings {
		if f.Severity == SeverityError && strings.Contains(f.Message, "discovery-only") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no ERROR finding naming the discovery-only invariant: %+v", report.Findings)
	}

	// The doctor cobra command exits non-zero (RunE returns error). The doctor
	// command carries its own --project-root flag (existing doctor_test pattern).
	doctorCmd := NewHarnessDoctorCmd()
	var out bytes.Buffer
	doctorCmd.SetOut(&out)
	doctorCmd.SetErr(&out)
	doctorCmd.SetArgs([]string{"--project-root", root})
	if err := doctorCmd.Execute(); err == nil {
		t.Fatalf("doctor did not fail on schema-invalid schedule; output: %s", out.String())
	}
}

// TestDoctor_ScheduleValidNoNewFindings verifies a VALID schedule introduces
// zero new ERROR findings (absent-schedule baseline preserved by AC-HCB-035;
// a valid declared schedule is equally clean).
func TestDoctor_ScheduleValidNoNewFindings(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", scheduledManifestJSON("30m", "loop", "discovery-only"))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor error: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Fatalf("valid schedule produced ERROR findings: %+v", report.Findings)
	}
}
