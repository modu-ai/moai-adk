package toolpolicy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestRecordLockPersistentGuardKeepsPrivateIdentity(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	release, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	guard, err := os.Open(path + ".guard")
	if err != nil {
		t.Fatalf("persistent guard missing while acquired: %v", err)
	}
	defer func() {
		if err := guard.Close(); err != nil {
			t.Errorf("close guard observer: %v", err)
		}
	}()
	before, err := guard.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if err := testguard.CheckFileModePolicy(path+".guard", 0o600); err != nil {
		t.Errorf("guard does not carry a real private policy: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	identity := ""
	for _, field := range strings.Fields(string(body)) {
		if strings.HasPrefix(field, "guard=") {
			identity = strings.TrimPrefix(field, "guard=")
		}
	}
	if identity == "" {
		t.Error("record does not bind the persistent guard identity")
	}
	release()
	releaseNext, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseNext()
	after, err := os.Stat(path + ".guard")
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("guard identity changed across release/acquire: %v", err)
	}
	body, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "guard="+identity+" ") {
		t.Errorf("successor changed guard binding: %q", body)
	}
	releaseNext()
	if _, err := os.Stat(path + ".guard"); err != nil {
		t.Errorf("release removed persistent guard: %v", err)
	}
}

func runRecordGuardChildHelper() {
	path := os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD_SETTINGS")
	signal := os.Getenv("MOAI_TOOLPOLICY_LOCK_CHILD_SIGNAL")
	g, err := openRecordGuard(path + ".guard")
	if err != nil {
		os.Exit(10)
	}
	if err := g.lockUntil(time.Now().Add(recordLockRetryBudget)); err != nil {
		os.Exit(11)
	}
	if err := os.WriteFile(signal, []byte("held"), 0o600); err != nil {
		os.Exit(12)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(signal + ".release"); err == nil {
			break
		}
		if time.Now().After(deadline) {
			os.Exit(13)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := unlockRecordGuard(g.file); err != nil {
		os.Exit(14)
	}
	if err := g.file.Close(); err != nil {
		os.Exit(15)
	}
	os.Exit(0)
}

func TestRecordLockKernelGuardExcludesAnotherProcess(t *testing.T) {
	root := t.TempDir()
	settings := filepath.Join(root, ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	signal := filepath.Join(root, "child-held")
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "MOAI_TOOLPOLICY_LOCK_CHILD=guard-hold",
		"MOAI_TOOLPOLICY_LOCK_CHILD_SETTINGS="+path,
		"MOAI_TOOLPOLICY_LOCK_CHILD_SIGNAL="+signal)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	t.Cleanup(func() {
		if !waited {
			if err := cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
				t.Logf("child kill: %v", err)
			}
			if err := cmd.Wait(); err != nil {
				t.Logf("child cleanup wait: %v", err)
			}
		}
	})
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(signal); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child did not hold kernel guard")
		}
		time.Sleep(20 * time.Millisecond)
	}
	release, err := acquireProjectRecordLock(settings)
	if release != nil {
		release()
	}
	if err == nil {
		t.Error("primary writer entered while another process held kernel guard")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("guard contention created primary record: %v", err)
	}
	if err := os.WriteFile(signal+".release", []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("child guard release: %v", err)
	}
	waited = true
	release, err = acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatalf("acquire after child released: %v", err)
	}
	release()
}

func TestRecordLockRefusesUnsafePersistentGuardObjects(t *testing.T) {
	for _, kind := range []string{"directory", "public-file", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
			path := managedAllowLockPath(settings)
			guard := path + ".guard"
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			const sentinel = "unrelated bytes must remain unchanged"
			switch kind {
			case "directory":
				if err := os.Mkdir(guard, 0o700); err != nil {
					t.Fatal(err)
				}
			case "public-file":
				if err := testguard.SeedFileModePolicy(guard, []byte(sentinel), 0o644); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				foreign := filepath.Join(t.TempDir(), "foreign")
				if err := testguard.SeedFileModePolicy(foreign, []byte(sentinel), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(foreign, guard); err != nil {
					t.Fatal(err)
				}
			}
			release, err := acquireProjectRecordLock(settings)
			if release != nil {
				release()
			}
			if err == nil {
				t.Error("unsafe persistent guard accepted")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("unsafe guard refusal created primary record: %v", err)
			}
			if kind != "directory" {
				b, err := os.ReadFile(guard)
				if err != nil || string(b) != sentinel {
					t.Errorf("unsafe guard data changed: %q, %v", b, err)
				}
			}
		})
	}
}
