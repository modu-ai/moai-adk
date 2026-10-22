//go:build windows

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"golang.org/x/sys/windows"
)

// Deny just FILE_EXECUTE to the current SID on a private fixture. Reading,
// WRITE_DAC and DELETE remain allowed; a Unix chmod fixture cannot express
// this native Windows boundary.
func nonExecutableClaudePinFixture(t *testing.T) (string, string) {
	t.Helper()
	pin := writeNativeToolFixture(t, "claude-denied", "package main\nfunc main() {}\n")
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sid := user.User.Sid.String()
	setDACL := func(sddl string) error {
		sd, err := windows.SecurityDescriptorFromString(sddl)
		if err != nil {
			return err
		}
		acl, _, err := sd.DACL()
		if err != nil {
			return err
		}
		return windows.SetNamedSecurityInfo(pin, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
	}
	t.Cleanup(func() {
		if err := setDACL("D:P(A;;FA;;;" + sid + ")"); err != nil {
			t.Errorf("restore private pin cleanup access: %v", err)
		}
	})
	if err := setDACL("D:P(D;;0x20;;;" + sid + ")(A;;FA;;;" + sid + ")"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(pin); err != nil {
		t.Fatalf("execute-only DENY must retain read access: %v", err)
	}
	path, err := windows.UTF16PtrFromString(pin)
	if err != nil {
		t.Fatal(err)
	}
	const share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE
	handle, err := windows.CreateFile(path, windows.FILE_EXECUTE, share, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err == nil {
		_ = windows.CloseHandle(handle)
		t.Fatal("execute DENY positive control unexpectedly opened FILE_EXECUTE")
	}
	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("native execute DENY=%v, want ACCESS_DENIED", err)
	}
	// Cleanup permission remains real, rather than depending on an elevated
	// runner or a cleanup exemption.
	handle, err = windows.CreateFile(path, windows.DELETE|windows.WRITE_DAC, share, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatalf("execute DENY must retain cleanup access: %v", err)
	}
	if err := windows.CloseHandle(handle); err != nil {
		t.Fatal(err)
	}
	return pin, "execute permission"
}

func TestValidateClaudeBinaryPinWindowsExecuteDenyFailsWithoutPATHFallback(t *testing.T) {
	dir, _ := fakePATHDir(t)
	pin, _ := nonExecutableClaudePinFixture(t)
	t.Setenv("PATH", dir)
	t.Setenv(config.EnvClaudeBin, pin)
	got, err := resolveClaudeBinaryAt("")
	if got != "" || !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("denied pin resolved=%q err=%v, want wrapped ACCESS_DENIED without fallback", got, err)
	}
	for _, want := range []string{pin, config.EnvClaudeBin, "not executable", "execute permission"} {
		if !strings.Contains(fmt.Sprint(err), want) {
			t.Errorf("denied pin diagnostic=%v, missing %q", err, want)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "claude.exe")); err != nil {
		t.Fatalf("PATH fallback positive control missing: %v", err)
	}
}
