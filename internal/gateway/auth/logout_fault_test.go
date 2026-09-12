package auth

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Preserve the independent auditor's actual permission-failure probe as a
// regression. Only test-owned scratch is chmodded, and cleanup restores access.
func TestLogoutScratchCleanupFailurePreservesCanonical(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permission fault; Windows execution is a separate gate")
	}
	for _, newLogin := range []bool{false, true} {
		name := "tombstone"
		if newLogin {
			name = "newer_login"
		}
		t.Run(name, func(t *testing.T) {
			s, err := OpenStore(privateStorePath(t))
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			second, err := OpenStore(s.dir)
			if err != nil {
				t.Fatal(err)
			}
			defer second.Close()
			generation := loginFixture(t, s)
			var scratch string
			var before []byte
			out, err := s.LogoutWithBroker(context.Background(), logoutFunc(func(ctx context.Context, home string) error {
				scratch = home
				t.Cleanup(func() {
					if err := os.Chmod(home, 0700); err != nil && !errors.Is(err, os.ErrNotExist) {
						t.Errorf("restore owned scratch permissions: %v", err)
					}
					if err := os.RemoveAll(home); err != nil {
						t.Errorf("remove owned scratch: %v", err)
					}
				})
				status, err := second.Status(ctx)
				if err != nil || status.LoggedIn || status.Generation != generation+1 {
					t.Fatal("local revoke not published before broker", status, err)
				}
				if newLogin {
					_, err = second.Login(ctx, brokerFunc(func(_ context.Context, fresh string, _ bool) error {
						if fresh == home {
							t.Fatal("new login reused logout scratch")
						}
						tokenFixture(t, fresh, "newer", time.Now().Add(time.Hour))
						return nil
					}))
					if err != nil {
						t.Fatal(err)
					}
				}
				before, err = os.ReadFile(filepath.Join(s.dir, "state.json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(filepath.Join(home, "retained"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(home, 0000); err != nil {
					t.Fatal(err)
				}
				return nil
			}))
			if err == nil {
				t.Skip("platform permitted removal despite mode 000; failure branch not observed")
			}
			if !errors.Is(err, ErrAuthState) || !out.LocalCompleted || out.Generation != generation+1 || out.BrokerOutcome != "cleanup_failed" || out.RemoteOutcome != "unknown" {
				t.Fatal(out, err)
			}
			if _, err := os.Lstat(scratch); err != nil {
				t.Fatal("failed cleanup path unexpectedly absent", err)
			}
			after, err := os.ReadFile(filepath.Join(s.dir, "state.json"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("cleanup failure changed canonical bytes", err)
			}
			status, err := second.Status(context.Background())
			expected := generation + 1
			if newLogin {
				expected++
			}
			if err != nil || status.Generation != expected || status.LoggedIn != newLogin {
				t.Fatal("wrong final canonical state", status, err)
			}
			t.Log("actual scratch cleanup failure returned local completed, cleanup_failed, remote unknown; canonical bytes preserved")
		})
	}
}
