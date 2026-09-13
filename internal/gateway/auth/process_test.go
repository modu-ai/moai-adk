package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuthStoreProcess(t *testing.T) {
	mode := os.Getenv("MOAI_AUTH_TEST_MODE")
	if mode == "" {
		return
	}
	dir := os.Getenv("MOAI_AUTH_TEST_DIR")
	s, e := OpenStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	switch mode {
	case "refresh", "refresh-late":
		b := brokerFunc(func(ctx context.Context, home string, refresh bool) error {
			f, e := os.OpenFile(filepath.Join(dir, "calls"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if e != nil {
				return e
			}
			if _, err := f.WriteString("call\n"); err != nil {
				_ = f.Close() // error path; the write error is what propagates
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			if mode == "refresh-late" {
				if e := os.WriteFile(filepath.Join(dir, "ready"), []byte("ready"), 0600); e != nil {
					return e
				}
				deadline := time.Now().Add(3 * time.Second)
				for {
					if _, e := os.Stat(filepath.Join(dir, "release")); e == nil {
						break
					}
					if time.Now().After(deadline) {
						return context.DeadlineExceeded
					}
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				time.Sleep(60 * time.Millisecond)
			}
			tokenFixture(t, home, "process", time.Now().Add(2*time.Hour))
			return nil
		})
		generation, e := s.Refresh(context.Background(), 1, b, func(context.Context, CredentialRef) error { return nil })
		if mode == "refresh-late" {
			if e != ErrCredentialChanged {
				t.Fatalf("late refresh %v", e)
			}
			return
		}
		if e != nil || generation != 2 {
			t.Fatalf("refresh %d %v", generation, e)
		}
	case "logout":
		if _, e = s.Logout(context.Background()); e != nil {
			t.Fatal(e)
		}
	case "lock":
		unlock, e := s.lock(context.Background(), "operation.lock")
		if e != nil {
			t.Fatal(e)
		}
		defer unlock()
		fmt.Println("LOCKED")
		for {
			time.Sleep(time.Hour)
		}
	}
}
func authProcess(t *testing.T, dir, mode string) *exec.Cmd {
	t.Helper()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(exe, "-test.run=^TestAuthStoreProcess$")
	cmd.Env = append(os.Environ(), "MOAI_AUTH_TEST_MODE="+mode, "MOAI_AUTH_TEST_DIR="+dir)
	return cmd
}
func TestCrossProcessRefreshSingleExchange(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	loginFixture(t, s)
	one := authProcess(t, s.dir, "refresh")
	two := authProcess(t, s.dir, "refresh")
	type result struct {
		raw []byte
		e   error
	}
	done := make(chan result, 2)
	for _, cmd := range []*exec.Cmd{one, two} {
		t.Cleanup(func() {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		})
		go func() { raw, e := cmd.CombinedOutput(); done <- result{raw, e} }()
	}
	for range 2 {
		r := <-done
		if r.e != nil {
			t.Fatalf("process failed %v %s", r.e, r.raw)
		}
	}
	raw, e := os.ReadFile(filepath.Join(s.dir, "calls"))
	if e != nil || strings.Count(string(raw), "call") != 1 {
		t.Fatalf("exchange count %q %v", raw, e)
	}
}

func TestCrossProcessLockCrashReleasesOwnership(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	cmd := authProcess(t, s.dir, "lock")
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		t.Fatal(e)
	}
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	line, e := bufio.NewReader(stdout).ReadString('\n')
	if e != nil || line != "LOCKED\n" {
		t.Fatalf("lock handshake %q %v", line, e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if unlock, e := s.lock(ctx, "operation.lock"); e == nil {
		unlock()
		t.Fatal("cross-process lock bypass")
	}
	if e = cmd.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	_ = cmd.Wait()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	unlock, e := s.lock(ctx2, "operation.lock")
	if e != nil {
		t.Fatal("crash leaked lock", e)
	}
	unlock()
}
func TestCrossProcessLateRefreshCannotResurrectLogout(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	loginFixture(t, s)
	cmd := authProcess(t, s.dir, "refresh-late")
	done := make(chan error, 1)
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill() // an already-exited child is the expected case
		}
	})
	go func() { done <- cmd.Run() }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, e := os.Stat(filepath.Join(s.dir, "ready")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("refresh readiness missing")
		}
		time.Sleep(10 * time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, e = s.Logout(ctx); e != nil {
		t.Fatal("operation lock blocked logout", e)
	}
	if e = os.WriteFile(filepath.Join(s.dir, "release"), []byte("release"), 0600); e != nil {
		t.Fatal(e)
	}
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if _, e = s.Resolve(); e != ErrCredentialAbsent {
		t.Fatal("late refresh resurrected state", e)
	}
}
