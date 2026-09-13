package gateway

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/auth"
)

func TestCatalogExactRoutesAndImmutableSnapshots(t *testing.T) {
	cat, err := NewSessionCatalog([]string{"glm-opus", "glm-sonnet", "glm-sonnet", "glm-haiku"})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "claude-opus-5", "claude-sonnet-5", "glm-opus", "glm-sonnet", "glm-haiku"} {
		entry, err := cat.Resolve(id)
		if err != nil || entry.UpstreamID != id {
			t.Fatalf("%s: %+v %v", id, entry, err)
		}
	}
	for _, id := range []string{"gpt-5.3", "gpt-6-astro", "gpt-6", "opus", "GPT-6-ASTRA", "gpt-6-astra "} {
		if _, err := cat.Resolve(id); !errors.Is(err, ErrUnknownModel) {
			t.Fatalf("unexpected route %q: %v", id, err)
		}
	}
	entries := cat.Entries()
	if len(entries) != 11 {
		t.Fatalf("duplicate GLM IDs not deduplicated: %d", len(entries))
	}
	entries[0].RouteID = "mutated"
	if cat.Entries()[0].RouteID == "mutated" {
		t.Fatal("snapshot escaped")
	}
	first, _ := cat.Resolve("gpt-6-astra")
	second, _ := cat.Resolve("glm-sonnet")
	if first.Provider != ProviderOpenAI || second.Provider != ProviderZAI {
		t.Fatal("request routes interfere")
	}
	if _, err := NewSessionCatalog([]string{"gpt-6-astra"}); !errors.Is(err, ErrConflictingModel) {
		t.Fatalf("cross-provider collision: %v", err)
	}
	if _, err := NewSessionCatalog([]string{""}); err == nil {
		t.Fatal("empty configured GLM ID accepted")
	}
}

func TestCatalogCopiesInputAndRejectsConflicts(t *testing.T) {
	input := []ModelEntry{{RouteID: "one", Provider: ProviderZAI, UpstreamID: "up", AuthMethod: AuthExistingGLM}}
	c, err := NewCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0].UpstreamID = "changed"
	e, _ := c.Resolve("one")
	if e.UpstreamID != "up" {
		t.Fatal("input aliases catalog")
	}
	for _, entries := range [][]ModelEntry{{{}}, {{RouteID: "x", Provider: "bogus", UpstreamID: "x"}}, {{RouteID: "x", Provider: ProviderZAI, UpstreamID: "x"}, {RouteID: "x", Provider: ProviderZAI, UpstreamID: "y"}}} {
		if _, err := NewCatalog(entries); err == nil {
			t.Fatalf("invalid entries accepted: %+v", entries)
		}
	}
	var empty CatalogSnapshot
	if _, err := empty.Resolve("anything"); !errors.Is(err, ErrUnknownModel) {
		t.Fatal(err)
	}
}

func TestAbsentCredentialContract(t *testing.T) {
	var ref auth.CredentialRef = auth.Absent{ProviderID: auth.ProviderOpenAI}
	if ref.Provider() != ProviderOpenAI {
		t.Fatal("wrong provider")
	}
	if _, err := ref.Generation(); !errors.Is(err, auth.ErrCredentialAbsent) {
		t.Fatal(err)
	}
	if err := ref.Apply(nil); !errors.Is(err, auth.ErrCredentialAbsent) {
		t.Fatal(err)
	}
	if ref.Redacted() != "openai:absent" {
		t.Fatal("unsafe description")
	}
}

func TestConcurrentRequestRouting(t *testing.T) {
	c, err := NewSessionCatalog([]string{"glm-test"})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"gpt-6-astra", "claude-opus-5", "glm-test"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			for range 100 {
				entry, err := c.Resolve(id)
				if err != nil || entry.RouteID != id {
					t.Fatalf("route %s: %+v %v", id, entry, err)
				}
				copy := c.Entries()
				copy[0].RouteID = id
			}
		})
	}
}
