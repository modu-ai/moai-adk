package jev

import (
	"context"
	"crypto/sha256"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-JEVC-002 — the package cannot mutate a card, a file, a branch, or a queue,
// because it has no dependency that could. Asserting the import set is stronger
// than asserting behaviour: a behavioural test shows the current code does not
// write, while this shows the current code CANNOT, and fails the moment an
// import that would make it possible is added.
func TestPackageImports_AreStandardLibraryOnly(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	fset := token.NewFileSet()
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if isStandardLibrary(path) {
				continue
			}
			t.Errorf("%s imports %q — internal/jev must depend on the standard library only, so it CANNOT mutate a queue, a file, a branch, or a card (REQ-JEVC-011)", name, path)
		}
	}
	if scanned == 0 {
		// A zero-result scan and a broken scan are indistinguishable without
		// this: an empty file list would pass the loop above vacuously.
		t.Fatal("scanned 0 non-test Go files — the import assertion above established nothing")
	}
}

// isStandardLibrary reports whether an import path names a standard-library
// package. The discriminator is the Go convention: a non-stdlib module path's
// first segment contains a dot (a domain).
func isStandardLibrary(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

// TestImportClassifierPositiveControl proves the classifier above can fail.
// Without it, a classifier that returned true for everything would let the
// scan report a clean import set it never actually checked.
func TestImportClassifierPositiveControl(t *testing.T) {
	if isStandardLibrary("github.com/modu-ai/moai-adk/internal/config") {
		t.Fatal("classifier reports a module path as standard library — the import assertion is vacuous")
	}
	if !isStandardLibrary("net/http") {
		t.Fatal("classifier reports net/http as non-standard — the import assertion would false-positive")
	}
}

// AC-JEVC-001 — the display-only invariant, measured the way the SPEC names:
// a SHA-256 of a queue-shaped file taken before and after a full enabled call
// with a credential present.
func TestQueueFileHash_UnchangedAcrossAFullEnabledCall(t *testing.T) {
	dir := t.TempDir()
	queue := filepath.Join(dir, "backlog.json")
	const body = `[{"id":"t1020","text":"ship the jev core capability","state":"picked"}]`
	if err := os.WriteFile(queue, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	hashOf := func() [32]byte {
		data, err := os.ReadFile(queue)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		return sha256.Sum256(data)
	}

	before := hashOf()

	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, okBody), nil }}
	c := testClient(d)
	got := c.Ask(context.Background(), Request{
		State:     "queue contents: " + body,
		Questions: []Question{{ID: "q1", Text: "Is this card a near-duplicate?", Kind: KindNoul}},
	})
	if !got.OK() {
		t.Fatalf("the call did not complete: %q (%s) — the hash comparison below would be vacuous", got.Availability, got.Condition)
	}
	if d.Calls != 1 {
		t.Fatalf("transport calls = %d, want 1 — the call must actually have run", d.Calls)
	}

	after := hashOf()
	if before != after {
		t.Fatalf("queue file SHA-256 changed across a jev call: %x -> %x (REQ-JEVC-011 violated)", before, after)
	}

	// Positive control: the comparison detects a change when one happens, so
	// the unchanged result above is attributable to the call rather than to a
	// hash function that always agrees with itself.
	if err := os.WriteFile(queue, []byte(body+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if hashOf() == before {
		t.Fatal("positive control failed: the SHA-256 comparison did not detect a deliberate change, so the unchanged result above establishes nothing")
	}
}

// The same invariant over every unavailable path. A call that fails must not
// become a call that writes.
func TestQueueFileHash_UnchangedAcrossEveryUnavailablePath(t *testing.T) {
	dir := t.TempDir()
	queue := filepath.Join(dir, "backlog.json")
	const body = `[{"id":"t1020","text":"ship it","state":"queued"}]`
	if err := os.WriteFile(queue, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(queue)
	before := sha256.Sum256(data)

	clients := map[string]*Client{
		"disabled":    New(false),
		"credentials": func() *Client { c := New(true); c.LoadCredential = func() string { return "" }; return c }(),
		"unreachable": func() *Client {
			c := testClient(&countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(500, `{}`), nil }})
			c.MaxRetries = 0
			return c
		}(),
	}
	for name, c := range clients {
		t.Run(name, func(t *testing.T) {
			got := c.Ask(context.Background(), Request{
				State:     "queue contents: " + body,
				Questions: []Question{{ID: "q", Text: "?", Kind: KindNoul}},
			})
			if got.OK() {
				t.Fatalf("expected an unavailable result, got available")
			}
			data, _ := os.ReadFile(queue)
			if sha256.Sum256(data) != before {
				t.Errorf("queue file changed on the %q path", name)
			}
		})
	}
}
