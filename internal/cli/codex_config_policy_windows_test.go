package cli

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Read the real ACL independently of synthetic Stat mode bits. Compare every
// ACE byte and the inheritance-protection setting, including broad/deny rules.
func codexConfigPolicyFingerprint(t *testing.T, path string) string {
	t.Helper()
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
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
	prefix := fmt.Sprintf("protected=%t;readonly=%t", control&windows.SE_DACL_PROTECTED != 0, attrs&windows.FILE_ATTRIBUTE_READONLY != 0)
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
