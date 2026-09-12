package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCredentialReferencesNeverLeakAcrossEndpointsOrGenerations(t *testing.T) {
	a := Absent{ProviderID: ProviderOpenAI}
	if a.Provider() != ProviderOpenAI || a.Redacted() != "openai:absent" {
		t.Fatal("absent provider")
	}
	if _, e := a.Generation(); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
	if e := a.Apply(nil); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	ref, e := s.Resolve()
	if e != nil {
		t.Fatal(e)
	}
	if ref.Provider() != ProviderOpenAI {
		t.Fatal("wrong provider")
	}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	req.Header = nil
	if e = ref.Apply(req); e != nil || req.Header.Get("ChatGPT-Account-Id") != "account-1" {
		t.Fatal(e)
	}
	for _, endpoint := range []string{APIEndpoint, SubscriptionEndpoint + "?x=1", SubscriptionEndpoint + "#fragment", "http://chatgpt.com/backend-api/codex/responses", "https://evil.example/responses"} {
		r, _ := http.NewRequest("POST", endpoint, nil)
		if e = ref.Apply(r); !errors.Is(e, ErrWrongProvider) || r.Header.Get("Authorization") != "" {
			t.Fatalf("endpoint %s %v", endpoint, e)
		}
	}
	_, e = s.Refresh(context.Background(), gen, brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "newer", time.Now().Add(2*time.Hour))
		return nil
	}), func(ctx context.Context, r CredentialRef) error {
		if r.Provider() != ProviderOpenAI || r.Redacted() != "openai:subscription" {
			t.Fatal("candidate metadata")
		}
		if g, e := r.Generation(); e != nil || g != gen+1 {
			t.Fatal("candidate generation")
		}
		request, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
		return r.Apply(request)
	})
	if e != nil {
		t.Fatal(e)
	}
	stale, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	if e = ref.Apply(stale); !errors.Is(e, ErrCredentialChanged) {
		t.Fatal(e)
	}
	fresh, _ := s.Resolve()
	if _, e = s.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e = fresh.Apply(stale); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
}
func TestStoreRejectsCorruptionPermissionsAndCanceledMutations(t *testing.T) {
	if s, e := OpenStore("relative"); e == nil {
		s.Close()
		t.Fatal("relative store")
	}
	for _, raw := range []string{`{`, `{"generation":0}`, `{"generation":1,"tombstone":true,"auth":{}}`, `{"generation":1,"auth":{}}`} {
		t.Run(raw, func(t *testing.T) {
			s, e := OpenStore(privateStorePath(t))
			if e != nil {
				t.Fatal(e)
			}
			defer s.Close()
			if e = os.WriteFile(filepath.Join(s.dir, "state.json"), []byte(raw), 0600); e != nil {
				t.Fatal(e)
			}
			if _, e = s.Resolve(); !errors.Is(e, ErrAuthState) {
				t.Fatal(e)
			}
			if _, e = s.Logout(context.Background()); !errors.Is(e, ErrAuthState) {
				t.Fatal(e)
			}
		})
	}
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = s.Status(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e = s.Login(ctx, brokerFunc(func(context.Context, string, bool) error { return nil })); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e = s.Logout(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e = s.Login(context.Background(), nil); !errors.Is(e, ErrBroker) {
		t.Fatal(e)
	}
	if _, e = s.Refresh(context.Background(), gen, nil, nil); !errors.Is(e, ErrRefresh) {
		t.Fatal(e)
	}
	if e = setTestFilePermissions(filepath.Join(s.dir, "state.json"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Status(context.Background()); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
	if e = setTestFilePermissions(filepath.Join(s.dir, "state.json"), 0600); e != nil {
		t.Fatal(e)
	}
	if e = setTestFilePermissions(filepath.Join(s.dir, "state.lock"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Status(context.Background()); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
}
func TestBrokerTokenValidationRejectsInvalidExpiryAndHeaderInjection(t *testing.T) {
	for _, token := range []string{"plain", "a.?.c", "a.e30.c", "a.ImJhZCI.c", "a.eyJleHAiOi0xfQ.c", "a.eyJleHAiOjF9.c\nforged"} {
		raw, _ := json.Marshal(authFile{Mode: "chatgpt", Tokens: tokenData{Access: token, Refresh: "refresh", ID: "id", Account: "acct"}})
		if _, _, e := parseAuth(raw); e == nil {
			t.Fatalf("token accepted %q", token)
		}
	}
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if _, e = s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "expired", time.Now().Add(-time.Hour))
		return nil
	})); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	if e = applySubscription(req, authFile{}, time.Now().Add(-time.Hour)); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
	if _, e = s.Refresh(context.Background(), 1, brokerFunc(func(context.Context, string, bool) error { return nil }), func(context.Context, CredentialRef) error { return nil }); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
}
func TestSendRejectsInvalidOptionsBodiesProxyAndGeneration(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	options := SendOptions{WriteTimeout: time.Second, PollInterval: time.Millisecond, MaxBodyBytes: 10}
	request, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader("{}"))
	transport := &http.Transport{}
	if _, e = s.SendAuthorized(context.Background(), gen, request, transport, SendOptions{}); !errors.Is(e, ErrSend) {
		t.Fatal(e)
	}
	if _, e = s.SendAuthorized(context.Background(), gen+1, request, transport, options); !errors.Is(e, ErrCredentialChanged) {
		t.Fatal(e)
	}
	proxy, _ := url.Parse("http://localhost:1234")
	transport.Proxy = http.ProxyURL(proxy)
	if _, e = s.SendAuthorized(context.Background(), gen, request, transport, options); !errors.Is(e, ErrSend) {
		t.Fatal(e)
	}
	transport.Proxy = nil
	request.GetBody = nil
	if _, e = s.SendAuthorized(context.Background(), gen, request, transport, options); !errors.Is(e, ErrSend) {
		t.Fatal(e)
	}
	request, _ = http.NewRequest("POST", APIEndpoint, strings.NewReader("{}"))
	if _, e = s.SendAuthorized(context.Background(), gen, request, transport, options); !errors.Is(e, ErrWrongProvider) {
		t.Fatal(e)
	}
}

func TestBrokerCannotReplaceScratchHomeWithSymlink(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	outside, e := privateBrokerTestHome(t)
	if e != nil {
		t.Fatal(e)
	}
	tokenFixture(t, outside, "outside", time.Now().Add(time.Hour))
	_, e = s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		if e := os.Rename(h, h+"-moved"); e != nil {
			return e
		}
		return os.Symlink(outside, h)
	}))
	if !errors.Is(e, ErrAuthState) {
		t.Fatalf("replaced scratch accepted: %v", e)
	}
	if _, e = s.Resolve(); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal("outside credential copied")
	}
}

func TestBrokerCannotSeedAPIKeyFallback(t *testing.T) {
	home := t.TempDir()
	tokenFixture(t, home, "subscription", time.Now().Add(time.Hour))
	raw, e := os.ReadFile(filepath.Join(home, "auth.json"))
	if e != nil {
		t.Fatal(e)
	}
	var fields map[string]any
	if e = json.Unmarshal(raw, &fields); e != nil {
		t.Fatal(e)
	}
	fields["OPENAI_API_KEY"] = "paid-fallback-key"
	raw, _ = json.Marshal(fields)
	if _, _, e = parseAuth(raw); !errors.Is(e, ErrAuthState) {
		t.Fatalf("API fallback seed accepted: %v", e)
	}
}

func TestSendRejectsBodyFailuresAndDialFailure(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	options := SendOptions{WriteTimeout: time.Second, PollInterval: time.Millisecond, MaxBodyBytes: 100}
	tr := &http.Transport{DialTLSContext: func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("secret-dial-error") }}
	for _, kind := range []string{"get-body", "body-length", "dial"} {
		req, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader("{}"))
		switch kind {
		case "get-body":
			req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("secret-body-error") }
		case "body-length":
			req.ContentLength = 3
		}
		if _, e = s.SendAuthorized(context.Background(), gen, req, tr, options); e == nil || strings.Contains(e.Error(), "secret-") {
			t.Fatalf("unsafe body/dial failure %v", e)
		}
	}
}
func TestBrokerRejectsInvalidLaunchConfiguration(t *testing.T) {
	home, e := privateBrokerTestHome(t)
	if e != nil {
		t.Fatal(e)
	}
	os.Chmod(home, 0700)
	for _, b := range []CodexBroker{{}, {Executable: "relative", Timeout: time.Second}, {Executable: "/missing-executable", Timeout: time.Second}, {Executable: "/missing-executable", Timeout: time.Second, OnLogin: func(LoginPrompt) error { return nil }}} {
		if e = b.Run(context.Background(), home, false); !errors.Is(e, ErrBroker) {
			t.Fatal(e)
		}
	}
}
func TestStoreOverflowAndCanceledBrokerCannotPublish(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	loginFixture(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	_, e = s.Login(ctx, brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "cancel", time.Now().Add(time.Hour))
		cancel()
		return nil
	}))
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(state{Generation: ^uint64(0), Tombstone: true})
	if e = os.WriteFile(filepath.Join(s.dir, "state.json"), raw, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Logout(context.Background()); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
	if _, e = s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "overflow", time.Now().Add(time.Hour))
		return nil
	})); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
	s.Close()
	if _, e = s.Status(context.Background()); !errors.Is(e, ErrAuthState) {
		t.Fatal(e)
	}
}

func TestScratchCleanupFailureDoesNotPublish(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	var locked string
	var releaseCleanup func()
	_, e = s.Login(context.Background(), brokerFunc(func(c context.Context, h string, r bool) error {
		tokenFixture(t, h, "cleanup", time.Now().Add(time.Hour))
		locked = filepath.Join(h, "locked")
		if e := os.Mkdir(locked, 0700); e != nil {
			return e
		}
		if e := os.WriteFile(filepath.Join(locked, "file"), []byte("private"), 0600); e != nil {
			return e
		}
		releaseCleanup, e = blockTestCleanup(locked)
		return e
	}))
	if releaseCleanup != nil {
		releaseCleanup()
	}
	if e == nil {
		t.Fatal("cleanup failure reported login success")
	}
	status, e := s.Status(context.Background())
	if e != nil || status.Generation != gen {
		t.Fatalf("cleanup failure published generation %d %v", status.Generation, e)
	}
}
