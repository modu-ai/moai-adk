package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

// Permission policy must exist before a caller puts any data in the stage.
func TestPolicyTempHasNativePolicyBeforeFirstWrite(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o644, 0o755} {
		t.Run(mode.String(), func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "target")
			f, err := atomicfile.CreatePolicyTemp(target, ".policy-*.tmp", mode, 0)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = f.Close(); _ = os.Remove(f.Name()) })
			if err := testguard.CheckFileModePolicy(f.Name(), mode); err != nil {
				t.Fatalf("stage policy before data: %v", err)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("staging published a target: %v", err)
			}
		})
	}
}

func TestPolicyTempPreservesObservedOwnerGroupAndCustomDACL(t *testing.T) {
	target := filepath.Join(t.TempDir(), "original")
	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	flags := windows.SECURITY_INFORMATION(windows.OWNER_SECURITY_INFORMATION | windows.GROUP_SECURITY_INFORMATION | windows.DACL_SECURITY_INFORMATION)
	sd, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT, flags)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		t.Fatalf("observe owner: %v", err)
	}
	custom, err := windows.SecurityDescriptorFromString("D:P(A;;0x17019f;;;" + owner.String() + ")(A;;FR;;;SY)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := custom.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(target, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	before, err := windows.GetNamedSecurityInfo(target, windows.SE_FILE_OBJECT, flags)
	if err != nil {
		t.Fatal(err)
	}
	f, err := atomicfile.CreatePolicyTemp(target, ".policy-*.tmp", 0o644, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close(); _ = os.Remove(f.Name()) })
	after, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, flags)
	if err != nil {
		t.Fatal(err)
	}
	if before.String() != after.String() {
		t.Errorf("staged owner/group/DACL changed: before=%q after=%q", before.String(), after.String())
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Errorf("staging changed original bytes: %q, %v", data, err)
	}
}

func TestPolicyTempExplicitModeOverridesExistingGrants(t *testing.T) {
	target := filepath.Join(t.TempDir(), "original")
	if err := testguard.SeedFileModePolicy(target, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := atomicfile.CreatePolicyTemp(target, ".policy-*.tmp", 0o644, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close(); _ = os.Remove(f.Name()) })
	if err := testguard.CheckFileModePolicy(f.Name(), 0o600); err != nil {
		t.Fatalf("explicit private policy was not established before data: %v", err)
	}
	if err := testguard.CheckFileModePolicy(target, 0o644); err != nil {
		t.Fatalf("staging modified the original policy: %v", err)
	}
}

func TestPolicyTempRefusesReadonlyOriginalWithoutChangingIt(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "original")
	if err := testguard.SeedFileModePolicy(target, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(target, 0o400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(target, 0o600) })
	if f, err := atomicfile.CreatePolicyTemp(target, ".policy-*.tmp", 0o644, 0); err == nil {
		_ = f.Close()
		t.Fatal("readonly original accepted")
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Errorf("refusal changed original bytes: %q, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("refusal left a stage: %v, %v", entries, err)
	}
}
