package atomicfile_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

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

// Characterize only the native creation round trip, before any staged data.
// Keep the complete-policy assertion strict even when the kernel changes it.
func TestPolicyTempNativeSecurityDescriptorRoundTrip(t *testing.T) {
	for _, custom := range []bool{false, true} {
		name := "inherited-control"
		if custom {
			name = "custom-protected"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			original := filepath.Join(dir, "original")
			// Empty, owned fixture: no application payload or user file.
			if err := os.WriteFile(original, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			originalHandle, err := windows.CreateFile(policyRoundTripPath(t, original), windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
				windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
				windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
			if err != nil {
				t.Fatalf("open original metadata handle: %v", err)
			}
			t.Cleanup(func() {
				if err := windows.CloseHandle(originalHandle); err != nil {
					t.Errorf("close original metadata handle: %v", err)
				}
			})
			readPolicy := func(h windows.Handle) *windows.SECURITY_DESCRIPTOR {
				t.Helper()
				sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT,
					windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
				if err != nil || sd == nil || !sd.IsValid() {
					t.Fatalf("read held owner/group/DACL: %v", err)
				}
				return sd
			}
			if custom {
				initial := readPolicy(originalHandle)
				owner, _, err := initial.Owner()
				if err != nil || owner == nil || !owner.IsValid() {
					t.Fatalf("read original owner: %v", err)
				}
				fixture, err := windows.SecurityDescriptorFromString("D:P(A;;0x17019f;;;" + owner.String() + ")(A;;FR;;;SY)")
				if err != nil {
					t.Fatal(err)
				}
				acl, _, err := fixture.DACL()
				if err != nil || acl == nil {
					t.Fatalf("read custom fixture DACL: %v", err)
				}
				// Exactly the existing custom test's one fixture-only DACL setup.
				if err := windows.SetNamedSecurityInfo(original, windows.SE_FILE_OBJECT,
					windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
					t.Fatalf("seed custom fixture DACL: %v", err)
				}
				t.Logf("fixture_requested_dacl=%q", fixture.String())
			}
			before := readPolicy(originalHandle)
			beforeText := before.String()
			policyRoundTripSnapshot(t, "before", originalHandle, before)
			wanted := before
			wantedText := beforeText
			policyRoundTripSnapshot(t, "wanted", originalHandle, wanted)
			sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: wanted}
			// Access, sharing, creation and attributes match CreatePolicyTemp.
			t.Logf("create desired_access=%#x share=%#x disposition=%#x attributes=%#x security_attributes_length=%d",
				windows.GENERIC_WRITE|windows.READ_CONTROL|windows.WRITE_DAC,
				windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.CREATE_NEW,
				windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, sa.Length)
			stageHandle, err := windows.CreateFile(policyRoundTripPath(t, filepath.Join(dir, "empty-stage")),
				windows.GENERIC_WRITE|windows.READ_CONTROL|windows.WRITE_DAC,
				windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, &sa,
				windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
			if err != nil {
				t.Fatalf("CREATE_NEW with original security descriptor: %v", err)
			}
			t.Cleanup(func() {
				if err := windows.CloseHandle(stageHandle); err != nil {
					t.Errorf("close empty stage handle: %v", err)
				}
			})
			actual := readPolicy(stageHandle)
			policyRoundTripSnapshot(t, "actual", stageHandle, actual)
			if wantedText == "" || actual.String() != wantedText {
				t.Errorf("native creation changed complete owner/group/DACL: before=%q wanted=%q actual=%q", beforeText, wantedText, actual.String())
			}
		})
	}
}

func policyRoundTripPath(t *testing.T, path string) *uint16 {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(abs, `\\?\`) {
		if strings.HasPrefix(abs, `\\`) {
			abs = `\\?\UNC\` + strings.TrimPrefix(abs, `\\`)
		} else {
			abs = `\\?\` + abs
		}
	}
	name, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func policyRoundTripSnapshot(t *testing.T, phase string, h windows.Handle, sd *windows.SECURITY_DESCRIPTOR) {
	t.Helper()
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		t.Fatalf("%s owner: %v", phase, err)
	}
	group, _, err := sd.Group()
	if err != nil || group == nil || !group.IsValid() {
		t.Fatalf("%s group: %v", phase, err)
	}
	control, revision, err := sd.Control()
	if err != nil {
		t.Fatalf("%s descriptor control: %v", phase, err)
	}
	if control&windows.SE_SELF_RELATIVE == 0 {
		t.Fatalf("%s GetSecurityInfo returned a non-self-relative descriptor", phase)
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil {
		t.Fatalf("%s DACL: %v", phase, err)
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		t.Fatalf("%s held metadata: %v", phase, err)
	}
	if info.FileSizeHigh != 0 || info.FileSizeLow != 0 {
		t.Errorf("%s diagnostic fixture unexpectedly contains data", phase)
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_READONLY) != 0 || info.NumberOfLinks != 1 {
		t.Errorf("%s diagnostic object is not a writable regular single-link file", phase)
	}
	sdSize := uintptr(sd.Length())
	t.Logf("phase=%s owner=%s group=%s complete_sddl=%q control=%#x revision=%d P=%t AI=%t AR=%t attributes=%#x readonly=%t ace_count=%d descriptor_bytes=%d file_id=%x:%x:%x",
		phase, owner, group, sd.String(), control, revision,
		control&windows.SE_DACL_PROTECTED != 0, control&windows.SE_DACL_AUTO_INHERITED != 0,
		control&windows.SE_DACL_AUTO_INHERIT_REQ != 0, info.FileAttributes,
		info.FileAttributes&windows.FILE_ATTRIBUTE_READONLY != 0, acl.AceCount, sdSize,
		info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			t.Fatalf("%s ACE%d: %v", phase, i, err)
		}
		if ace == nil || ace.Header.AceSize < uint16(unsafe.Offsetof(ace.SidStart)) {
			t.Fatalf("%s ACE%d is truncated", phase, i)
		}
		offset := uintptr(unsafe.Pointer(ace)) - uintptr(unsafe.Pointer(sd))
		size := uintptr(ace.Header.AceSize)
		if offset > sdSize || size > sdSize-offset {
			t.Fatalf("%s ACE%d lies outside its held security descriptor", phase, i)
		}
		data := unsafe.Slice((*byte)(unsafe.Pointer(ace)), int(ace.Header.AceSize))
		t.Logf("phase=%s ace=%d type=%#x flags=%#x mask=%#x bytes=%x", phase, i, ace.Header.AceType, ace.Header.AceFlags, ace.Mask, data)
	}
	runtime.KeepAlive(sd)
}
