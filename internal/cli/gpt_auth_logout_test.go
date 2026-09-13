package cli

import (
	"bytes"
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

type cliLogoutBroker struct {
	cliGPTBrokerFunc
	logout func(context.Context, string) error
}

func (b cliLogoutBroker) Logout(c context.Context, h string) error { return b.logout(c, h) }

func TestGPTLogoutSeparatesLocalBrokerAndUnknownRemote(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "completed", true: "failed"}[failed], func(t *testing.T) {
			home := t.TempDir()
			home, _ = filepath.EvalSymlinks(home)
			t.Setenv("MOAI_HOME", home)
			legacy := t.TempDir()
			t.Setenv("CODEX_HOME", legacy)
			legacyFile := filepath.Join(legacy, "auth.json")
			if e := os.WriteFile(legacyFile, []byte("preserve"), 0600); e != nil {
				t.Fatal(e)
			}
			var out bytes.Buffer
			calls := 0
			b := cliLogoutBroker{cliGPTBrokerFunc: func(c context.Context, h string, r bool) error {
				claims, _ := json.Marshal(map[string]int64{"exp": time.Now().Add(time.Hour).Unix()})
				token := "e30." + base64.RawURLEncoding.EncodeToString(claims) + ".fixture"
				raw, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": token, "refresh_token": "private-refresh-fixture", "id_token": "identity", "account_id": "account"}})
				return os.WriteFile(filepath.Join(h, "auth.json"), raw, 0600)
			}, logout: func(c context.Context, h string) error {
				calls++
				if h == legacy || !strings.HasPrefix(h, filepath.Join(home, "gateway-auth", "logout-")) {
					t.Fatal("unowned scratch")
				}
				if failed {
					return errors.New("private failure")
				}
				return os.Remove(filepath.Join(h, "auth.json"))
			}}
			s := newGPTAuthServices(&out, b)
			if e := s.Login(context.Background()); e != nil {
				t.Fatal(e)
			}
			out.Reset()
			if e := s.Logout(context.Background()); e != nil {
				t.Fatal(e)
			}
			want := "broker logout completed"
			if failed {
				want = "broker logout failed"
			}
			if calls != 1 || !strings.Contains(out.String(), "local logout completed") || !strings.Contains(out.String(), want) || !strings.Contains(out.String(), "remote revocation unknown") || strings.Contains(out.String(), "private") {
				t.Fatal(calls, out.String())
			}
			out.Reset()
			if e := s.Status(context.Background()); e != nil || !strings.Contains(out.String(), "not logged in") {
				t.Fatal(e, out.String())
			}
			raw, e := os.ReadFile(legacyFile)
			if e != nil || string(raw) != "preserve" {
				t.Fatal("legacy changed")
			}
		})
	}
}

func TestGPTInstalledLogoutMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if e := (installedGPTBroker{Out: &bytes.Buffer{}}).Logout(context.Background(), t.TempDir()); e == nil {
		t.Fatal("missing logout broker accepted")
	}
}
