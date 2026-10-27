package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const localInstructionsVerb = "moai migrate local-instructions"

// runDoctorInProject runs a full `moai doctor` with root as the working
// directory and returns its stdout.
func runDoctorInProject(t *testing.T, root string) (string, error) {
	t.Helper()
	t.Chdir(root)
	var out, errB bytes.Buffer
	cmd := &cobra.Command{Use: "doctor-local-instructions-test"}
	cmd.SetOut(&out)
	cmd.SetErr(&errB)
	cmd.SetContext(context.Background())
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().Bool("fix", false, "")
	cmd.Flags().String("export", "", "")
	cmd.Flags().String("check", "", "")
	err := runDoctor(cmd, nil)
	return out.String(), err
}

// updateThenDoctor runs a real, non-dry-run `moai update` (its template sync
// deploying the embedded tree) followed by `moai doctor`, both in root.
func updateThenDoctor(t *testing.T, root string) (updateOut, doctorOut string) {
	t.Helper()
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)
	out, errOut, err := snapRunUpdate(t, root)
	if err != nil {
		t.Fatalf("moai update: %v\nout:\n%s\nerr:\n%s", err, out, errOut)
	}
	docOut, err := runDoctorInProject(t, root)
	if err != nil {
		t.Fatalf("moai doctor: %v\nout:\n%s", err, docOut)
	}
	return out, docOut
}

// AC-IFU-015 (REQ-IFU-011, direction-reversed by SPEC-CLAUDE-INSTRUCTION-
// FILES-001): `moai update` then `moai doctor` both exit 0 and leave the
// legacy AGENTS.local.md byte-identical across the whole sequence — the
// migration is never implicit.
func TestLocalInstructions_UpdateDoctorPreserveFile(t *testing.T) {
	root := snapNewUpdateProject(t, "0.0.0") // older template_version: the sync runs
	body := "# 사용자 지침\nruntime-written: /abs/path\n"
	writeLocalFile(t, root, legacyLocalInstructionName, body)
	path := filepath.Join(root, legacyLocalInstructionName)
	before := localFileDigest(t, path)

	updateThenDoctor(t, root)

	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("AGENTS.local.md gone after update+doctor: %v", err)
	}
	if got := localFileDigest(t, path); got != before {
		t.Error("AGENTS.local.md changed across update+doctor")
	}
	if _, err := os.Lstat(filepath.Join(root, canonicalLocalInstructionName)); !os.IsNotExist(err) {
		t.Errorf("update+doctor created CLAUDE.local.md (err=%v) — migration must be explicit", err)
	}
	// Guard against a vacuous pass: the sync really deployed into this tree.
	if _, err := os.Lstat(filepath.Join(root, ".claude")); err != nil {
		t.Fatalf("template sync did not deploy (.claude absent): %v", err)
	}
}

// AC-IFU-030 (REQ-IFU-012, direction-reversed): each of the two commands'
// stdout carries the migration advisory exactly when a legacy local file is
// present — parity across both is the assertion.
func TestLocalInstructions_UpdateDoctorAdvisoryParity(t *testing.T) {
	cases := []struct {
		name   string
		files  []string
		advise bool
	}{
		{"legacy-only", []string{legacyLocalInstructionName}, true},
		{"both", []string{legacyLocalInstructionName, canonicalLocalInstructionName}, true},
		{"canonical-only", []string{canonicalLocalInstructionName}, false},
		{"neither", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := snapNewUpdateProject(t, "dev") // version match: fast path
			for _, f := range tc.files {
				writeLocalFile(t, root, f, "body of "+f+"\n")
			}
			digests := map[string][32]byte{}
			for _, f := range tc.files {
				digests[f] = sha256.Sum256([]byte("body of " + f + "\n"))
			}

			updateOut, doctorOut := updateThenDoctor(t, root)

			for label, out := range map[string]string{"update": updateOut, "doctor": doctorOut} {
				if got := strings.Contains(out, localInstructionsVerb); got != tc.advise {
					t.Errorf("%s stdout names %q = %v, want %v\n%s", label, localInstructionsVerb, got, tc.advise, out)
				}
			}
			for f, want := range digests {
				if got := localFileDigest(t, filepath.Join(root, f)); got != want {
					t.Errorf("%s modified by update+doctor", f)
				}
			}
		})
	}
}

// The advisory text is one function for all surfaces; both of its advising
// branches name the verb and the legacy file it concerns.
func TestLocalInstructionsAdvisory_Text(t *testing.T) {
	if got := localInstructionsAdvisory(false, false); got != "" {
		t.Errorf("neither: advisory = %q, want empty", got)
	}
	if got := localInstructionsAdvisory(true, false); got != "" {
		t.Errorf("canonical-only: advisory = %q, want empty", got)
	}
	for _, canonical := range []bool{false, true} {
		got := localInstructionsAdvisory(canonical, true)
		if !strings.Contains(got, localInstructionsVerb) || !strings.Contains(got, legacyLocalInstructionName) {
			t.Errorf("canonical=%v legacy=true: advisory = %q, want it to name %q and %q", canonical, got, localInstructionsVerb, legacyLocalInstructionName)
		}
	}
}
