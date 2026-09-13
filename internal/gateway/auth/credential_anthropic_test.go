package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAnthropicExplicitAPIKeyBoundary(t *testing.T) {
	for _, key := range []string{"", "x\r\ny", "x\x00y"} {
		if _, e := NewAnthropicAPIKey(key); e == nil {
			t.Fatal("bad key")
		}
	}
	r, e := NewAnthropicAPIKey("synthetic-anthropic")
	if e != nil {
		t.Fatal(e)
	}
	if r.Provider() != ProviderAnthropic || strings.Contains(r.Redacted(), "synthetic") {
		t.Fatal("identity")
	}
	if g, e := r.Generation(); e != nil || g != 1 {
		t.Fatal(g, e)
	}
	req, _ := http.NewRequest("POST", AnthropicEndpoint, nil)
	req.Header.Set("Authorization", "inherited")
	if e = r.Apply(req); e != nil {
		t.Fatal(e)
	}
	if req.Header.Get("X-Api-Key") != "synthetic-anthropic" || req.Header.Get("Authorization") != "" {
		t.Fatal("key header")
	}
	for _, endpoint := range []string{APIEndpoint, SubscriptionEndpoint, "https://api.anthropic.com/v1/messages?x=1", "https://evil.invalid/v1/messages"} {
		req, _ := http.NewRequest("POST", endpoint, nil)
		if e = r.Apply(req); !errors.Is(e, ErrWrongProvider) || req.Header.Get("X-Api-Key") != "" {
			t.Fatal(endpoint, e)
		}
	}
	if e = r.Apply(nil); !errors.Is(e, ErrWrongProvider) {
		t.Fatal(e)
	}
}
