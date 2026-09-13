//go:build windows

package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestWindowsStoreLoginRefreshLogout(t *testing.T) {
	s, e := OpenStore(filepath.Join(t.TempDir(), "store"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	ref, e := s.Resolve()
	if e != nil {
		t.Fatal(e)
	}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	if e = ref.Apply(req); e != nil {
		t.Fatal(e)
	}
	if req.Header.Get("Authorization") == "" {
		t.Fatal("missing credential")
	}
	next, e := s.Refresh(context.Background(), gen, brokerFunc(func(ctx context.Context, h string, r bool) error {
		tokenFixture(t, h, "new", time.Now().Add(2*time.Hour))
		return nil
	}), func(context.Context, CredentialRef) error { return nil })
	if e != nil || next != gen+1 {
		t.Fatalf("refresh %d %v", next, e)
	}
	if _, e = s.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Resolve(); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatalf("logout %v", e)
	}
	if e = s.validatePlatformRoot(); e != nil {
		t.Fatal(e)
	}
}

func TestWindowsStoreRejectsRootReplacement(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	// A pinned directory handle must deny rename while the store is live.
	if e = os.Rename(dir, dir+"-moved"); e == nil {
		t.Fatal("live root rename accepted")
	}
	if e = s.validatePlatformRoot(); e != nil {
		t.Fatal(e)
	}
}

func TestWindowsStoreRejectsInheritedBroadBrokerFile(t *testing.T) {
	s, e := OpenStore(filepath.Join(t.TempDir(), "store"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	before, e := s.Status(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Login(context.Background(), brokerFunc(func(ctx context.Context, h string, r bool) error {
		tokenFixture(t, h, "bad", time.Now().Add(time.Hour))
		return setWindowsTestBroadACL(filepath.Join(h, "auth.json"))
	}))
	if !errors.Is(e, ErrAuthState) {
		t.Fatalf("broad auth accepted: %v", e)
	}
	after, e := s.Status(context.Background())
	if e != nil || after.Generation != before.Generation {
		t.Fatalf("prior state changed: %#v %v", after, e)
	}
}

func setWindowsTestBroadACL(path string) error {
	sd, e := windows.SecurityDescriptorFromString("D:P(A;;FA;;;WD)")
	if e != nil {
		return e
	}
	acl, _, e := sd.DACL()
	if e != nil {
		return e
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}

func TestWindowsStoreUncertainCommitBlocksCredential(t *testing.T) {
	s, e := OpenStore(filepath.Join(t.TempDir(), "store"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	ref, e := s.Resolve()
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(s.dir, "state.json")
	info, e := os.Stat(path)
	if e != nil {
		t.Fatal(e)
	}
	// Deterministic post-API readback contradiction. This exercises the actual
	// finalization boundary used after MoveFileEx, not a forced OpenStore flag.
	if e = s.finishWindowsCommit(path, info, []byte("different candidate")); !errors.Is(e, ErrAuthCommitUncertain) {
		t.Fatalf("uncertainty not surfaced: %v", e)
	}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	if e = ref.Apply(req); e == nil || req.Header.Get("Authorization") != "" {
		t.Fatalf("unverified credential escaped: %v", e)
	}
	if _, e = s.Status(context.Background()); e == nil {
		t.Fatal("uncertain Store remained usable")
	}
}

func TestWindowsStoreCrashFixture(t *testing.T) {
	dir := os.Getenv("MOAI_WINDOWS_STORE_CRASH_DIR")
	if dir == "" {
		return
	}
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if os.Getenv("MOAI_WINDOWS_STORE_CRASH_STAGE") == "flushed" {
		if e = writeWindowsPrivateCandidate(filepath.Join(dir, ".state-crash"), []byte("uncommitted")); e != nil {
			t.Fatal(e)
		}
	} else {
		if _, e = s.Logout(context.Background()); e != nil {
			t.Fatal(e)
		}
	}
	fmt.Println("READY")
	time.Sleep(15 * time.Second)
}
func TestWindowsStoreProcessCrashRecovery(t *testing.T) {
	for _, stage := range []string{"flushed", "committed"} {
		t.Run(stage, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "store")
			s, e := OpenStore(dir)
			if e != nil {
				t.Fatal(e)
			}
			gen := loginFixture(t, s)
			if e = s.Close(); e != nil {
				t.Fatal(e)
			}
			exe, e := os.Executable()
			if e != nil {
				t.Fatal(e)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWindowsStoreCrashFixture$")
			cmd.Env = append(os.Environ(), "MOAI_WINDOWS_STORE_CRASH_DIR="+dir, "MOAI_WINDOWS_STORE_CRASH_STAGE="+stage)
			pipe, e := cmd.StdoutPipe()
			if e != nil {
				t.Fatal(e)
			}
			if e = cmd.Start(); e != nil {
				t.Fatal(e)
			}
			waited := false
			t.Cleanup(func() {
				cmd.Process.Kill()
				if !waited {
					cmd.Wait()
				}
			})
			line, e := bufio.NewReader(pipe).ReadString('\n')
			if e != nil || line != "READY\n" {
				t.Fatalf("handshake %q %v", line, e)
			}
			if e = cmd.Process.Kill(); e != nil {
				t.Fatal(e)
			}
			e = cmd.Wait()
			waited = true
			if e == nil {
				t.Fatal("child did not terminate abnormally")
			}
			s, e = OpenStore(dir)
			if e != nil {
				t.Fatal(e)
			}
			defer s.Close()
			status, e := s.Status(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			if stage == "flushed" {
				if status.Generation != gen || !status.LoggedIn {
					t.Fatalf("prior state lost: %#v", status)
				}
			} else {
				if status.Generation != gen+1 || status.LoggedIn {
					t.Fatalf("tombstone lost: %#v", status)
				}
			}
		})
	}
}
func TestWindowsStoreRejectsNullACL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if e = windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, nil, nil); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Status(context.Background()); e == nil {
		t.Fatal("NULL DACL accepted")
	}
}

func setTestFilePermissions(path string, mode os.FileMode) error {
	if mode == 0644 {
		return setWindowsTestBroadACL(path)
	}
	sd, e := windowsPrivateDescriptor(false)
	if e != nil {
		return e
	}
	acl, _, e := sd.DACL()
	if e != nil {
		return e
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
func blockTestCleanup(path string) (func(), error) {
	name, e := windows.UTF16PtrFromString(filepath.Join(path, "file"))
	if e != nil {
		return nil, e
	}
	h, e := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		return nil, e
	}
	return func() { windows.CloseHandle(h) }, nil
}

func TestWindowsStoreReplacementFailurePreservesCanonical(t *testing.T) {
	s, e := OpenStore(filepath.Join(t.TempDir(), "store"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	path := filepath.Join(s.dir, "state.json")
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	ptr, e := windows.UTF16PtrFromString(path)
	if e != nil {
		t.Fatal(e)
	}
	h, e := windows.CreateFile(ptr, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CloseHandle(h)
	// Denying delete-sharing deterministically makes MoveFileEx fail.
	_, e = s.Logout(context.Background())
	if !errors.Is(e, ErrAuthCommitUncertain) {
		t.Fatalf("API failure not fail closed: %v", e)
	}
	after, e := os.ReadFile(path)
	if e != nil || string(after) != string(before) {
		t.Fatal("failed replacement changed canonical")
	}
}

func TestWindowsStorePinsBrokerScratch(t *testing.T) {
	s, e := OpenStore(filepath.Join(t.TempDir(), "store"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	_, e = s.Login(context.Background(), brokerFunc(func(ctx context.Context, h string, r bool) error {
		if err := os.Rename(h, h+"-replaced"); err == nil {
			t.Fatal("broker scratch rename accepted")
		}
		tokenFixture(t, h, "private", time.Now().Add(time.Hour))
		return nil
	}))
	if e != nil {
		t.Fatal(e)
	}
}

func TestWindowsCandidateIdentitySurvivesRename(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	candidate := filepath.Join(dir, "candidate")
	if e = writeWindowsPrivateCandidate(candidate, []byte("verified")); e != nil {
		t.Fatal(e)
	}
	identity, e := privatePathInfo(candidate, false)
	if e != nil {
		t.Fatal(e)
	}
	final := filepath.Join(dir, "final")
	if e = replaceWindowsWriteThrough(candidate, final); e != nil {
		t.Fatal(e)
	}
	finalIdentity, e := privatePathInfo(final, false)
	if e != nil {
		t.Fatal(e)
	}
	if !os.SameFile(identity, finalIdentity) {
		t.Fatal("captured identity depended on vanished candidate path")
	}
	if e = s.validatePlatformRoot(); e != nil {
		t.Fatal(e)
	}
}
