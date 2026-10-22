package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

// Read the real ACL independently of synthetic Stat mode bits. Compare every
// ACE byte and the inheritance-protection setting, including broad/deny rules.
func codexConfigPolicyFingerprint(t *testing.T, path string) string {
	t.Helper()
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		t.Fatalf("read original owner: %v", err)
	}
	control, _, err := sd.Control()
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	attrs, err := windows.GetFileAttributes(p)
	if err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("owner=%s;protected=%t;readonly=%t", owner.String(), control&windows.SE_DACL_PROTECTED != 0, attrs&windows.FILE_ATTRIBUTE_READONLY != 0)
	if acl == nil {
		return prefix + ";null-dacl"
	}
	var rows []string
	for i := range uint32(acl.AceCount) {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, hex.EncodeToString(unsafe.Slice((*byte)(unsafe.Pointer(ace)), int(ace.Header.AceSize))))
	}
	return prefix + ";" + strings.Join(rows, "|")
}

func TestRestoreCodexConfigPolicyPreservesOwnerAndNativeRules(t *testing.T) {
	dir := t.TempDir()
	original, staged := filepath.Join(dir, "config.toml"), filepath.Join(dir, "staged")
	for _, path := range []string{original, staged} {
		if err := privatefile.WriteFile(path, []byte("synthetic config")); err != nil {
			t.Fatal(err)
		}
	}
	testguard.MakeFilePublic(t, original)
	before := codexConfigPolicyFingerprint(t, original)
	if err := restoreCodexConfigPolicy(staged, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := codexConfigPolicyFingerprint(t, staged); got != before {
		t.Fatalf("restored native policy=%s, want %s", got, before)
	}
}

func TestRestoreCodexConfigPolicyReportsOwnerRestorationFailure(t *testing.T) {
	dir := t.TempDir()
	original, staged := filepath.Join(dir, "config.toml"), filepath.Join(dir, "staged")
	for _, path := range []string{original, staged} {
		if err := privatefile.WriteFile(path, []byte("synthetic config")); err != nil {
			t.Fatal(err)
		}
	}
	sd, err := windows.GetNamedSecurityInfo(original, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	wantOwner, _, err := sd.Owner()
	if err != nil {
		t.Fatal(err)
	}
	previous := setCodexConfigSecurityInfo
	t.Cleanup(func() { setCodexConfigSecurityInfo = previous })
	called := false
	setCodexConfigSecurityInfo = func(handle windows.Handle, object windows.SE_OBJECT_TYPE, flags windows.SECURITY_INFORMATION, owner, group *windows.SID, dacl, sacl *windows.ACL) error {
		called = true
		if flags&windows.OWNER_SECURITY_INFORMATION == 0 || owner == nil || !owner.Equals(wantOwner) {
			t.Error("original owner was not included in the restoration")
		}
		if flags&windows.DACL_SECURITY_INFORMATION == 0 || dacl == nil {
			t.Error("original DACL was not included in the restoration")
		}
		return windows.ERROR_ACCESS_DENIED
	}
	if err := restoreCodexConfigPolicy(staged, original, 0o600); !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("owner restoration failure=%v, want ACCESS_DENIED", err)
	}
	if !called {
		t.Fatal("owner restoration was not attempted")
	}
	if data, err := os.ReadFile(original); err != nil || string(data) != "synthetic config" {
		t.Fatalf("original config changed: %q %v", data, err)
	}
}
