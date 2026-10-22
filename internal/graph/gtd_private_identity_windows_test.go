package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"golang.org/x/sys/windows"
)

func TestPrivateGTDWindowsRealSIDAndPolicyDenial(t *testing.T) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sid := user.User.Sid.String()
	if sid == "" {
		t.Fatal("current token SID is unavailable")
	}
	checker := SystemPrivateGTDPermissionChecker()
	current, err := checker.CurrentUID()
	if err != nil || current != sid {
		t.Fatalf("current identity %q differs from token SID: %v", current, err)
	}
	for _, tc := range []struct {
		name, policy string
		directory    bool
		protected    bool
	}{
		{"file-read-only", "D:P(A;;FR;;;" + sid + ")", false, true},
		{"file-inheritance", "D:(A;;FA;;;" + sid + ")", false, false},
		{"directory-missing-child-policy", "D:P(A;;FA;;;" + sid + ")", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "todo")
			projection, err := BuildPrivateGTDProjection(context.Background(), filepath.Join(dir, "backlog.db"), 1, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := AuthorizePrivateGTDProjection(projection, current, checker); err != nil {
				t.Fatalf("native private control denied: %v", err)
			}
			for _, path := range []string{dir, projection.DataPath, projection.MetaPath} {
				owner, err := checker.OwnerUID(path)
				if err != nil || owner != sid {
					t.Fatalf("owner %q differs from token SID: %v", owner, err)
				}
			}
			path := projection.MetaPath
			if tc.directory {
				path = dir
			}
			setPrivateGTDTestDACL(t, path, tc.policy, tc.protected)
			owner, err := checker.OwnerUID(path)
			if err != nil || owner != current {
				t.Fatalf("DACL-only mutation changed identity: owner=%q error=%v", owner, err)
			}
			if err := AuthorizePrivateGTDProjection(projection, current, checker); err == nil {
				t.Fatal("current owner bypassed invalid native private policy")
			}
		})
	}
	if _, err := checker.OwnerUID(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing owner accepted")
	}
	if _, err := privateGTDSIDIdentity(nil); err == nil {
		t.Fatal("unknown SID accepted")
	}
}

func TestPrivateGTDWindowsReplaceRetainsPrivateDACLAndIdentity(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "todo")
	publisher := osPrivateGTDPublisher{}
	if err := publisher.PrepareDir(dir); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, privateGTDDataName)
	if err := privatefile.WriteFile(target, []byte("old nonsecret fixture")); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	setPrivateGTDTestDACL(t, target, "D:(A;;FA;;;"+user.User.Sid.String()+")", false)
	if err := privatefile.Check(target); err == nil {
		t.Fatal("unprotected destination control unexpectedly private")
	}
	staging, err := publisher.WriteTemp(dir, ".gtd-native-", []byte("new nonsecret fixture"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := staging.Close(); err != nil {
			t.Error(err)
		}
	})
	before, err := privateGTDTestFileInfo(staging.Name())
	if err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(staging.Name()); err != nil {
		t.Fatal(err)
	}
	if err := publisher.Install(staging, target); err != nil {
		t.Fatal(err)
	}
	if err := staging.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Lstat(target)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("MoveFileEx did not retain private source object: %v", err)
	}
	if err := privatefile.Check(target); err != nil {
		t.Fatalf("destination retained unsafe prior policy: %v", err)
	}
}

func setPrivateGTDTestDACL(t *testing.T, path, policy string, protected bool) {
	t.Helper()
	sd, err := windows.SecurityDescriptorFromString(policy)
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION)
	if !protected {
		flags = windows.DACL_SECURITY_INFORMATION | windows.UNPROTECTED_DACL_SECURITY_INFORMATION
	}
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, flags, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
}
