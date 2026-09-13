package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func freshFixture(t *testing.T, expired bool) *Store {
	t.Helper()
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	loginFixture(t, s) // Establish the explicit-login store and its sidecar locks first.
	expiry := time.Now().Add(time.Hour)
	if expired {
		expiry = time.Now().Add(-time.Hour)
	}
	home := t.TempDir()
	tokenFixture(t, home, "existing", expiry)
	raw, e := os.ReadFile(filepath.Join(home, "auth.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = s.write(state{Generation: 1, Auth: raw}); e != nil {
		t.Fatal(e)
	}
	return s
}
func TestResolveFreshDoesNotAutoLoginOrRefreshValid(t *testing.T) {
	s := freshFixture(t, false)
	var calls atomic.Int32
	b := brokerFunc(func(context.Context, string, bool) error { calls.Add(1); return errors.New("unexpected") })
	v := func(context.Context, CredentialRef) error { calls.Add(1); return nil }
	ref, e := s.ResolveFresh(context.Background(), b, v)
	if e != nil || !s.Owns(ref) || calls.Load() != 0 {
		t.Fatalf("valid ref %v calls %d", e, calls.Load())
	}
	if _, e = s.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	ref, e = s.ResolveFresh(context.Background(), b, v)
	if !errors.Is(e, ErrCredentialAbsent) || ref != nil || calls.Load() != 0 {
		t.Fatalf("logout %v calls %d", e, calls.Load())
	}
	empty, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer empty.Close()
	if ref, e = empty.ResolveFresh(context.Background(), b, v); !errors.Is(e, ErrCredentialAbsent) || ref != nil || calls.Load() != 0 {
		t.Fatal("missing store auto-login")
	}
}
func TestResolveFreshConcurrentSingleRefresh(t *testing.T) {
	s := freshFixture(t, true)

	var brokers, verifiers atomic.Int32
	b := brokerFunc(func(c context.Context, h string, refresh bool) error {
		if !refresh {
			return errors.New("login forbidden")
		}
		brokers.Add(1)
		tokenFixture(t, h, "fresh", time.Now().Add(time.Hour))
		return nil
	})
	v := func(context.Context, CredentialRef) error { verifiers.Add(1); return nil }
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ref, e := s.ResolveFresh(context.Background(), b, v)
			if e != nil || !s.Owns(ref) {
				t.Errorf("resolve %v", e)
			}
		}()
	}
	wg.Wait()
	if brokers.Load() != 1 || verifiers.Load() != 1 {
		t.Fatalf("broker=%d verifier=%d", brokers.Load(), verifiers.Load())
	}
}
func TestResolveFreshRejectsFailureAndLateLogout(t *testing.T) {
	for _, mode := range []string{"broker", "verify", "expired", "logout"} {
		t.Run(mode, func(t *testing.T) {
			s := freshFixture(t, true)
			b := brokerFunc(func(c context.Context, h string, refresh bool) error {
				if mode == "broker" {
					return errors.New("failure")
				}
				expiry := time.Now().Add(time.Hour)
				if mode == "expired" {
					expiry = time.Now().Add(-time.Minute)
				}
				tokenFixture(t, h, "candidate", expiry)
				if mode == "logout" {
					_, e := s.Logout(c)
					return e
				}
				return nil
			})
			v := func(context.Context, CredentialRef) error {
				if mode == "verify" {
					return errors.New("401")
				}
				return nil
			}
			ref, e := s.ResolveFresh(context.Background(), b, v)
			if e == nil || ref != nil {
				t.Fatalf("failed candidate returned %v %v", ref, e)
			}
			status, e := s.Status(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			if mode == "logout" {
				if status.LoggedIn {
					t.Fatal("resurrected")
				}
			} else if status.Generation != 1 {
				t.Fatal("failed refresh changed generation")
			}
		})
	}
}
func TestResolveFreshContextBoundsStateLock(t *testing.T) {
	s := freshFixture(t, false)
	unlock, e := s.lock(context.Background(), "state.lock")
	if e != nil {
		t.Fatal(e)
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	ref, e := s.ResolveFresh(ctx, nil, nil)
	if !errors.Is(e, context.DeadlineExceeded) || ref != nil {
		t.Fatalf("lock context %v", e)
	}
}

func TestResolveFreshReusesConcurrentValidGenerationAfterFailedAttempt(t *testing.T) {
	s := freshFixture(t, true)
	b := brokerFunc(func(c context.Context, h string, refresh bool) error {
		tokenFixture(t, h, "concurrent-login", time.Now().Add(time.Hour))
		raw, e := os.ReadFile(filepath.Join(h, "auth.json"))
		if e != nil {
			return e
		}
		unlock, e := s.lock(c, "state.lock")
		if e != nil {
			return e
		}
		defer unlock()
		if e = s.write(state{Generation: 2, Auth: raw}); e != nil {
			return e
		}
		return errors.New("stale broker failed")
	})
	ref, e := s.ResolveFresh(context.Background(), b, func(context.Context, CredentialRef) error { return nil })
	if e != nil || !s.Owns(ref) {
		t.Fatalf("newer ref %v", e)
	}
	if generation, e := ref.Generation(); e != nil || generation != 2 {
		t.Fatalf("generation %d %v", generation, e)
	}
}

func TestResolveFreshMissingRefreshDependenciesReturnsNoExpiredRef(t *testing.T) {
	s := freshFixture(t, true)
	for _, verify := range []func(context.Context, CredentialRef) error{nil, func(context.Context, CredentialRef) error { return nil }} {
		if ref, e := s.ResolveFresh(context.Background(), nil, verify); e == nil || ref != nil {
			t.Fatalf("expired ref %v %v", ref, e)
		}
	}
}
