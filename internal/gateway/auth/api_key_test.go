package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestExplicitAPIKeyIsSeparateFromSubscription(t *testing.T) {
	for _, key := range []string{"", "key\nforged"} {
		if _, e := NewAPIKey(key); e == nil {
			t.Fatal("invalid key accepted")
		}
	}
	ref, e := NewAPIKey("secret-api-marker")
	if e != nil {
		t.Fatal(e)
	}
	if ref.Provider() != ProviderOpenAI || strings.Contains(ref.Redacted(), "secret") {
		t.Fatal("unsafe API metadata")
	}
	if gen, e := ref.Generation(); e != nil || gen != 1 {
		t.Fatal("API generation")
	}
	req, _ := http.NewRequest("POST", APIEndpoint, nil)
	if e = ref.Apply(req); e != nil || req.Header.Get("Authorization") != "Bearer secret-api-marker" {
		t.Fatal(e)
	}
	foreign, _ := http.NewRequest("POST", SubscriptionEndpoint, nil)
	if e = ref.Apply(foreign); !errors.Is(e, ErrWrongProvider) || foreign.Header.Get("Authorization") != "" {
		t.Fatal("API key sent to subscription")
	}
}
