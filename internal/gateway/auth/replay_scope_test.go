package auth

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReplayScopeSurvivesRefreshButRejectsReplacement(t *testing.T) {
	s, err := OpenStore(privateStorePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	gen := loginFixture(t, s)
	ref, _ := s.Resolve()
	first, err := s.ReplayScope(context.Background(), ref)
	if err != nil || first == "" {
		t.Fatal(err)
	}
	_, err = s.Refresh(context.Background(), gen, brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "new", time.Now().Add(2*time.Hour))
		return nil
	}), func(context.Context, CredentialRef) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReplayScope(context.Background(), ref); err == nil {
		t.Fatal("stale ref accepted")
	}
	ref, _ = s.Resolve()
	second, err := s.ReplayScope(context.Background(), ref)
	if err != nil || first != second {
		t.Fatal("refresh changed owner scope", err)
	}
	_, err = s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "other", time.Now().Add(3*time.Hour))
		p := filepath.Join(h, "auth.json")
		raw, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		return os.WriteFile(p, bytes.ReplaceAll(raw, []byte("account-1"), []byte("account-2")), 0600)
	}))
	if err != nil {
		t.Fatal(err)
	}
	ref, _ = s.Resolve()
	third, err := s.ReplayScope(context.Background(), ref)
	if err != nil || first == third {
		t.Fatal("account replacement shared scope", err)
	}
}

func TestReplayScopeRejectsAbsentForeignAndCancelledReferences(t *testing.T) {
	s, err := OpenStore(privateStorePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	loginFixture(t, s)
	ref, _ := s.Resolve()
	for _, bad := range []CredentialRef{nil, Absent{ProviderID: ProviderOpenAI}, (*storeRef)(nil)} {
		if _, err = s.ReplayScope(context.Background(), bad); err == nil {
			t.Fatal("foreign ref accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.ReplayScope(ctx, ref); err == nil {
		t.Fatal("cancelled scope accepted")
	}
	if _, err := s.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReplayScope(context.Background(), ref); err == nil {
		t.Fatal("logged out scope accepted")
	}
}
