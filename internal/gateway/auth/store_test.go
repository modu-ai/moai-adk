package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type brokerFunc func(context.Context, string, bool) error

func (f brokerFunc) Run(c context.Context, h string, r bool) error { return f(c, h, r) }
func tokenFixture(t *testing.T, home, marker string, expiry time.Time) {
	t.Helper()
	claims, _ := json.Marshal(map[string]any{"exp": expiry.Unix()})
	token := "e30." + base64.RawURLEncoding.EncodeToString(claims) + "." + marker
	raw, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": token, "refresh_token": "refresh-" + marker, "id_token": "identity", "account_id": "account-1"}})
	if err := os.WriteFile(filepath.Join(home, "auth.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func loginFixture(t *testing.T, s *Store) uint64 {
	t.Helper()
	gen, err := s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "old", time.Now().Add(time.Hour))
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	return gen
}
func TestStoreLoginAtomicPrivateAndLogoutGeneration(t *testing.T) {
	s, err := OpenStore(privateStorePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.Resolve(); !errors.Is(err, ErrCredentialAbsent) {
		t.Fatal(err)
	}
	gen := loginFixture(t, s)
	ref, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if got, e := ref.Generation(); e != nil || got != gen {
		t.Fatalf("generation %d %v", got, e)
	}
	if strings.Contains(ref.Redacted(), "refresh") {
		t.Fatal("secret exposed")
	}
	info, err := os.Stat(filepath.Join(s.dir, "state.json"))
	if err != nil || !privateRegular(filepath.Join(s.dir, "state.json"), info) {
		t.Fatalf("private state %v %v", info, err)
	}
	next, err := s.Logout(context.Background())
	if err != nil || next <= gen {
		t.Fatalf("logout %d %v", next, err)
	}
	if _, err = ref.Generation(); !errors.Is(err, ErrCredentialAbsent) {
		t.Fatal(err)
	}
}
func TestStoreRejectsPartialFailureAndLateLoginAfterLogout(t *testing.T) {
	s, err := OpenStore(privateStorePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	for _, b := range []brokerFunc{
		func(c context.Context, h string, r bool) error {
			return os.WriteFile(filepath.Join(h, "auth.json"), []byte(`{"tokens":`), 0600)
		},
		func(c context.Context, h string, r bool) error {
			tokenFixture(t, h, "bad", time.Now().Add(time.Hour))
			return errors.New("secret-marker")
		},
	} {
		if _, err = s.Login(context.Background(), b); err == nil || strings.Contains(err.Error(), "secret-marker") {
			t.Fatalf("unsafe failure %v", err)
		}
	}
	ref, _ := s.Resolve()
	if got, _ := ref.Generation(); got != gen {
		t.Fatal("failure changed generation")
	}
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, e := s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
			close(started)
			<-release
			tokenFixture(t, h, "late", time.Now().Add(time.Hour))
			return nil
		}))
		done <- e
	}()
	<-started
	if _, err = s.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err = <-done; !errors.Is(err, ErrCredentialChanged) {
		t.Fatal(err)
	}
	if _, err = s.Resolve(); !errors.Is(err, ErrCredentialAbsent) {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(s.dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "broker-") {
			t.Fatal("scratch retained")
		}
	}
}
func TestRefreshRequiresChangedTokenAndProviderAcceptance(t *testing.T) {
	s, err := OpenStore(privateStorePath(t))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	unchanged := brokerFunc(func(context.Context, string, bool) error { return nil })
	accept := func(context.Context, CredentialRef) error { return nil }
	if _, err = s.Refresh(context.Background(), gen, unchanged, accept); err == nil {
		t.Fatal("unchanged refresh accepted")
	}
	changed := brokerFunc(func(c context.Context, h string, r bool) error {
		if !r {
			t.Fatal("not refresh")
		}
		tokenFixture(t, h, "new", time.Now().Add(2*time.Hour))
		return nil
	})
	if _, err = s.Refresh(context.Background(), gen, changed, func(context.Context, CredentialRef) error { return errors.New("401") }); err == nil {
		t.Fatal("rejected token published")
	}
	next, err := s.Refresh(context.Background(), gen, changed, accept)
	if err != nil || next != gen+1 {
		t.Fatalf("refresh %d %v", next, err)
	}
	called := false
	got, err := s.Refresh(context.Background(), gen, brokerFunc(func(context.Context, string, bool) error { called = true; return nil }), accept)
	if err != nil || got != next || called {
		t.Fatalf("duplicate refresh %d %v %v", got, err, called)
	}
}
func TestStoreRejectsSymlinkAndPublicPermissions(t *testing.T) {
	dir := t.TempDir()
	public := filepath.Join(dir, "public")
	if err := os.Mkdir(public, 0755); err != nil {
		t.Fatal(err)
	}
	if s, e := OpenStore(public); e == nil {
		s.Close()
		t.Fatal("public store accepted")
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	if s, e := OpenStore(link); e == nil {
		s.Close()
		t.Fatal("symlink store accepted")
	}
}

func privateStorePath(t *testing.T) string {
	t.Helper()
	dir, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return filepath.Join(dir, "auth")
}
