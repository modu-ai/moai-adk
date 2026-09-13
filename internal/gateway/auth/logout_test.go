package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type logoutFunc func(context.Context, string) error

func (f logoutFunc) Logout(c context.Context, h string) error { return f(c, h) }

func TestLogoutOwnedSnapshotAfterLocalCommitWithoutOperationLock(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	g := loginFixture(t, s)
	operation, e := s.lock(context.Background(), "operation.lock")
	if e != nil {
		t.Fatal(e)
	}
	defer operation()
	var scratch string
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, e := s.LogoutWithBroker(ctx, logoutFunc(func(c context.Context, h string) error {
		scratch = h
		status, e := s.Status(c)
		if e != nil || status.LoggedIn || status.Generation != g+1 {
			t.Fatal("broker preceded tombstone", status, e)
		}
		data, e := os.ReadFile(filepath.Join(h, "auth.json"))
		if e != nil || !strings.Contains(string(data), "refresh-old") {
			t.Fatal("owned snapshot missing", e)
		}
		for _, p := range []string{h, filepath.Join(h, "auth.json")} {
			i, e := os.Lstat(p)
			if e != nil || !testPrivatePath(p, i) {
				t.Fatal("non-private scratch")
			}
		}
		return errors.New("private broker error")
	}))
	if e != nil || !out.LocalCompleted || out.Generation != g+1 || out.BrokerOutcome != "failed" || out.RemoteOutcome != "unknown" {
		t.Fatal(out, e)
	}
	if _, e := os.Stat(scratch); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("scratch not cleaned")
	}
	status, e := s.Status(ctx)
	if e != nil || status.LoggedIn {
		t.Fatal("remote failure restored auth")
	}
}

func TestLogoutOldSnapshotCannotOverwriteNewLogin(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	g := loginFixture(t, s)
	out, e := s.LogoutWithBroker(context.Background(), logoutFunc(func(c context.Context, h string) error {
		newGen, e := s.Login(c, brokerFunc(func(c context.Context, newHome string, _ bool) error {
			if newHome == h {
				t.Fatal("scratch reused")
			}
			tokenFixture(t, newHome, "new", time.Now().Add(time.Hour))
			return nil
		}))
		if e != nil || newGen != g+2 {
			t.Fatal(newGen, e)
		}
		return os.Remove(filepath.Join(h, "auth.json"))
	}))
	if e != nil || out.BrokerOutcome != "completed" || out.RemoteOutcome != "unknown" {
		t.Fatal(out, e)
	}
	status, e := s.Status(context.Background())
	if e != nil || !status.LoggedIn || status.Generation != g+2 {
		t.Fatal("old logout changed new canonical", status, e)
	}
}

func TestLogoutLocalFailureAndAbsentBroker(t *testing.T) {
	for _, kind := range []string{"local failure", "no credential", "no broker"} {
		t.Run(kind, func(t *testing.T) {
			s, e := OpenStore(privateStorePath(t))
			if e != nil {
				t.Fatal(e)
			}
			defer s.Close()
			if kind != "no credential" {
				loginFixture(t, s)
			}
			calls := 0
			var broker LogoutBroker = logoutFunc(func(context.Context, string) error { calls++; return nil })
			if kind == "no broker" {
				broker = nil
			}
			if kind == "local failure" {
				if e := os.WriteFile(filepath.Join(s.dir, "state.json"), []byte("invalid"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			out, e := s.LogoutWithBroker(context.Background(), broker)
			if calls != 0 {
				t.Fatal("unexpected broker")
			}
			if kind == "local failure" {
				if e == nil || out.LocalCompleted {
					t.Fatal("local failure called success")
				}
				return
			}
			if e != nil || !out.LocalCompleted || out.RemoteOutcome != "not_attempted" {
				t.Fatal(out, e)
			}
		})
	}
}

func TestLogoutCanceledBrokerKeepsTombstoneAndCleansScratch(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var scratch string
	out, e := s.LogoutWithBroker(ctx, logoutFunc(func(c context.Context, home string) error {
		scratch = home
		deadline, ok := c.Deadline()
		if !ok || time.Until(deadline) > 30*time.Second {
			t.Fatal("broker deadline absent")
		}
		cancel()
		<-c.Done()
		// Even a misreported nil result cannot override the canceled context.
		return nil
	}))
	if e != nil || !out.LocalCompleted || out.BrokerOutcome != "failed" || out.RemoteOutcome != "unknown" {
		t.Fatal(out, e)
	}
	if _, e := os.Stat(scratch); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("canceled scratch remains")
	}
	status, e := s.Status(context.Background())
	if e != nil || status.LoggedIn {
		t.Fatal("cancellation restored credential")
	}
}
