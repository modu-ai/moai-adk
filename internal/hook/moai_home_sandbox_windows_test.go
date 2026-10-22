package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/sqliteguard"
	"golang.org/x/sys/windows"
)

func TestFactoryHookHomeFixtureNativePrivacy(t *testing.T) {
	rebindEnv(t, "native-home-fixture", "lane-1", "codex")
	home := os.Getenv(config.EnvHome)
	if err := privatefile.CheckDirectory(home); err != nil {
		t.Fatalf("factory fixture home native privacy: %v", err)
	}
	root := rebindRoot(t, "native-home-fixture=active")
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "native fixture factory", db)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := privatefile.Check(db.Path + suffix); err != nil {
			t.Errorf("factory object suffix=%q native privacy: %v", suffix, err)
		}
	}
}

func TestFactoryHookRefusesBroadExistingHomeWithoutMutation(t *testing.T) {
	home := privateHookHome(t)
	marker := filepath.Join(home, "existing-marker")
	const payload = "existing namespace payload"
	if err := os.WriteFile(marker, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")(A;OICI;FR;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(home, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	policy := func() string {
		t.Helper()
		sd, err := windows.GetNamedSecurityInfo(home, windows.SE_FILE_OBJECT,
			windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		return sd.String()
	}
	before := policy()
	if err := privatefile.CheckDirectory(home); err == nil {
		t.Fatal("broad existing fixture unexpectedly private")
	}
	if err := sqliteguard.EnsureDirectory(home); err == nil {
		t.Fatal("broad existing namespace accepted")
	}
	t.Setenv(config.EnvHome, home)
	root := t.TempDir()
	if db, err := homestate.OpenFactory(root); err == nil {
		_ = db.Close()
		t.Fatal("factory accepted broad existing home")
	}
	if after := policy(); after != before {
		t.Fatalf("existing home policy changed: before=%s after=%s", before, after)
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != payload {
		t.Fatalf("existing payload changed: %q, %v", got, err)
	}
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("factory database created before home refusal: %v", err)
	}
}
