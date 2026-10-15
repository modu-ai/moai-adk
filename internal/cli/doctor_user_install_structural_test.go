package cli

// doctor_user_install_structural_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// finding A: the doctor's project-vs-lock check must judge the project manifest
// read-only, including a manifest that is valid JSON with the wrong shape.
//
// checkProjectVsLock guarded only JSON validity before it called manifest.Load.
// Load moves a manifest it cannot decode to .corrupt and overwrites any prior
// recovery copy, so a structurally invalid manifest such as {"files":123} was
// renamed by a read-only diagnostic command.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestCheckProjectVsLockKeepsStructurallyInvalidManifestInPlace reproduces finding A.
// The manifest is valid JSON whose "files" value is a number, and a prior .corrupt
// recovery copy already exists. After the doctor check the manifest must still sit
// at its path with its original bytes, the recovery copy must be unchanged, and the
// check must report at the failure grade.
func TestCheckProjectVsLockKeepsStructurallyInvalidManifestInPlace(t *testing.T) {
	root := t.TempDir()
	moaiDir := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const structurallyInvalid = `{"files":123}`
	const priorRecoveryCopy = "PRIOR-RECOVERY-COPY"
	manifestPath := filepath.Join(moaiDir, "manifest.json")
	corruptPath := manifestPath + ".corrupt"
	if err := os.WriteFile(manifestPath, []byte(structurallyInvalid), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corruptPath, []byte(priorRecoveryCopy), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkProjectVsLock(root, true)

	if check.Status != uikit.CheckFail {
		t.Errorf("a structurally invalid manifest was reported at grade %v, want the failure grade; message: %q", check.Status, check.Message)
	}
	got, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Errorf("the structurally invalid manifest left its path during a read-only doctor check: %v", err)
	} else if string(got) != structurallyInvalid {
		t.Errorf("the manifest bytes at %s changed during a read-only doctor check: got %q, want %q", manifestPath, got, structurallyInvalid)
	}
	recovery, err := os.ReadFile(corruptPath)
	if err != nil {
		t.Fatalf("the prior .corrupt recovery copy disappeared during a read-only doctor check: %v", err)
	}
	if string(recovery) != priorRecoveryCopy {
		t.Errorf("the prior .corrupt recovery copy was overwritten during a read-only doctor check: now holds %q", recovery)
	}
}
