package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestAnthropicOAuthRequestBoundary(t *testing.T) {
	for _, h := range []http.Header{nil, {"Authorization": {"Bearer "}}, {"Authorization": {"Basic abc"}}, {"Authorization": {"Bearer a b"}}, {"Authorization": {"Bearer a\n"}}, {"Authorization": {"Bearer abc", "Bearer def"}}, {"Authorization": {"Bearer abc"}, "authorization": {"Bearer def"}}, {"Authorization": {"Bearer abc"}, "X-Api-Key": {"conflict"}}} {
		if _, err := NewAnthropicOAuth(context.Background(), h); !errors.Is(err, ErrCredentialAbsent) {
			t.Fatalf("accepted invalid credential: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := http.Header{"Authorization": {"Bearer synthetic-oauth"}}
	ref, err := NewAnthropicOAuth(ctx, h)
	if err != nil {
		t.Fatal(err)
	}
	h.Set("Authorization", "Bearer changed")
	req, _ := http.NewRequest("POST", AnthropicEndpoint, nil)
	req.Header.Set("X-Api-Key", "old")
	if err := ref.Apply(req); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "Bearer synthetic-oauth" || req.Header.Get("X-Api-Key") != "" {
		t.Fatal("credential isolation")
	}
	if strings.Contains(fmt.Sprintf("%v %#v %s", ref, ref, ref.Redacted()), "synthetic") {
		t.Fatal("secret in diagnostic")
	}
	for _, target := range []string{SubscriptionEndpoint, "https://api.anthropic.com/v1/messages?x=1", "http://api.anthropic.com/v1/messages"} {
		req, _ := http.NewRequest("POST", target, nil)
		if !errors.Is(ref.Apply(req), ErrWrongProvider) {
			t.Fatal("destination accepted")
		}
	}
	req, _ = http.NewRequest("POST", AnthropicEndpoint, nil)
	req.Host = "evil.invalid"
	if !errors.Is(ref.Apply(req), ErrWrongProvider) {
		t.Fatal("host override accepted")
	}
	cancel()
	req, _ = http.NewRequest("POST", AnthropicEndpoint, nil)
	if !errors.Is(ref.Apply(req), context.Canceled) {
		t.Fatal("cancel lost")
	}
	if _, err := ref.Generation(); !errors.Is(err, context.Canceled) {
		t.Fatal("generation after cancel")
	}
}

func TestAnthropicOAuthMalformedPaddingAndContext(t *testing.T) {
	for _, value := range []string{"Bearer =", "Bearer a=b", "Bearer café", "Bearer " + strings.Repeat("x", 16385)} {
		if _, err := NewAnthropicOAuth(context.Background(), http.Header{"Authorization": {value}}); !errors.Is(err, ErrCredentialAbsent) {
			t.Fatal("invalid token accepted")
		}
	}
	if _, err := NewAnthropicOAuth(nil, nil); !errors.Is(err, ErrCredentialAbsent) {
		t.Fatal("nil context")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewAnthropicOAuth(ctx, http.Header{"Authorization": {"Bearer valid"}}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled resolution")
	}
	ref, err := NewAnthropicOAuth(context.Background(), http.Header{"authorization": {"bearer valid=="}})
	if err != nil {
		t.Fatal(err)
	}
	if ref.Provider() != ProviderAnthropic {
		t.Fatal("provider")
	}
	if g, err := ref.Generation(); g != 1 || err != nil {
		t.Fatal("generation")
	}
	req, _ := http.NewRequest("POST", AnthropicEndpoint, nil)
	req.Header = nil
	if err := ref.Apply(req); err != nil || req.Header.Get("Authorization") != "Bearer valid==" {
		t.Fatal("normalized immutable reference")
	}
	cancelled := req.WithContext(ctx)
	if err := ref.Apply(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatal("outbound cancellation lost")
	}
}
