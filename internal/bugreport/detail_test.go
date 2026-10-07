package bugreport

import (
	"strings"
	"testing"
)

func TestDetailIsClosedSetMembership(t *testing.T) {
	// Each template/harness token is accepted for the kinds its register row
	// names and rejected for every other kind.
	validPairs := map[TemplateToken][]Kind{
		TokenPathTraversal:     {KindTemplateDeployFailure},
		TokenNotFound:          {KindTemplateDeployFailure},
		TokenPreserveIntegrity: {KindTemplateDeployFailure},
		TokenMissingKey:        {KindHarnessDefect},
		TokenUnexpandedToken:   {KindHarnessDefect},
		TokenInvalidJSON:       {KindHarnessDefect},
	}
	for tok, kinds := range validPairs {
		for _, k := range kinds {
			d, err := ParseDetail(k, tok.Token())
			if err != nil {
				t.Errorf("ParseDetail(%s, %q): %v", k, tok, err)
				continue
			}
			tt, ok := d.(TemplateToken)
			if !ok || tt != tok {
				t.Errorf("ParseDetail(%s, %q) = %#v, want the same token", k, tok, d)
			}
		}
		// ...and rejected for a kind its row does not name.
		for _, k := range AllKinds() {
			if containsKind(kinds, k) {
				continue
			}
			if _, err := ParseDetail(k, tok.Token()); err == nil {
				t.Errorf("token %q accepted for kind %s, which its register row does not name", tok, k)
			}
		}
	}

	// Hook detail accepts only a registered event paired with a registered
	// handler name.
	const testEvent = "TestEventClosedSet"
	const testHandler = "testHandlerClosedSet"
	if err := RegisterEvent(testEvent); err != nil {
		t.Fatalf("RegisterEvent: %v", err)
	}
	if err := RegisterHandlerName(testHandler); err != nil {
		t.Fatalf("RegisterHandlerName: %v", err)
	}
	hd, err := RegisterHookIdentity(testEvent, testHandler)
	if err != nil {
		t.Fatalf("RegisterHookIdentity: %v", err)
	}
	for _, k := range []Kind{KindHookHandlerFailure, KindHookTimeout} {
		d, err := ParseDetail(k, hd.Token())
		if err != nil {
			t.Errorf("ParseDetail(%s, %q): %v", k, hd.Token(), err)
			continue
		}
		got, ok := d.(HookDetail)
		if !ok || got != hd {
			t.Errorf("ParseDetail(%s) = %#v, want %#v", k, d, hd)
		}
	}

	// An unregistered handler name is rejected even in the right shape.
	if _, err := ParseDetail(KindHookHandlerFailure, "event="+testEvent+" handler=neverRegistered"); err == nil {
		t.Error("ParseDetail accepted an unregistered handler name")
	}
	if _, err := ParseDetail(KindHookHandlerFailure, "event=neverRegisteredEvent handler="+testHandler); err == nil {
		t.Error("ParseDetail accepted an unregistered event")
	}
	// A malformed hook token is rejected (shape, not membership).
	for _, bad := range []string{"event-only", "event=" + testEvent, "handler=" + testHandler, "a b c"} {
		if _, err := ParseDetail(KindHookHandlerFailure, bad); err == nil {
			t.Errorf("ParseDetail accepted malformed hook token %q", bad)
		}
	}

	// Kinds whose register rows carry no closed set accept no detail at all.
	for _, k := range []Kind{KindPanic, KindInternalError} {
		for _, s := range []string{"", TokenMissingKey.Token(), "anything"} {
			if _, err := ParseDetail(k, s); err == nil {
				t.Errorf("ParseDetail(%s, %q) accepted a detail for a kind with no closed set", k, s)
			}
		}
	}
}

func containsKind(kinds []Kind, k Kind) bool {
	for _, got := range kinds {
		if got == k {
			return true
		}
	}
	return false
}

func TestDetailRejectsCanaryAndPathShapedStrings(t *testing.T) {
	strs := []string{
		"CANARY-/Users/leak/secret-token",
		"WorktreeCreate//srv/customer/private-project",
		"/etc/passwd",
		"",
	}
	kinds := AllKinds()
	for _, s := range strs {
		for _, k := range kinds {
			if _, err := ParseDetail(k, s); err == nil {
				t.Errorf("ParseDetail(%s, %q) accepted a canary/path/empty string", k, s)
			}
		}
		// Registration validators reject the same strings: no path shape can
		// become a hook identity.
		if err := RegisterEvent(s); err == nil {
			t.Errorf("RegisterEvent(%q) accepted a canary/path/empty string", s)
		}
		if err := RegisterHandlerName(s); err == nil {
			t.Errorf("RegisterHandlerName(%q) accepted a canary/path/empty string", s)
		}
		if _, err := RegisterHookIdentity("testEventCanary", s); err == nil {
			t.Errorf("RegisterHookIdentity with handler %q accepted a canary/path/empty string", s)
		}
	}

	// The registration shape rule keeps "/", ".", and path shapes out even
	// for names that are otherwise ordinary.
	for _, bad := range []string{"has/slash", "has.dot", "has:colon", "has space", "9starts-with-digit"} {
		if err := RegisterHandlerName(bad); err == nil {
			t.Errorf("RegisterHandlerName(%q) accepted a non-conforming name", bad)
		}
	}
}

// TestTemplateTokenSetIsExactlySix pins the closed token set by name: adding
// or renaming a member fails here, making the closure visible.
func TestTemplateTokenSetIsExactlySix(t *testing.T) {
	want := []string{
		"path_traversal",
		"not_found",
		"preserve_integrity",
		"missing_key",
		"unexpanded_token",
		"invalid_json",
	}
	got := AllTemplateTokens()
	if len(got) != len(want) {
		t.Fatalf("AllTemplateTokens = %v, want exactly %d tokens", got, len(want))
	}
	seen := map[string]bool{}
	for i, tok := range got {
		if tok.Token() != want[i] {
			t.Fatalf("token %d = %q, want %q (order is part of the pin)", i, tok, want[i])
		}
		seen[tok.Token()] = true
	}
	if len(seen) != len(want) {
		t.Fatalf("duplicate tokens in AllTemplateTokens: %v", got)
	}
}

// TestDetailTokenRoundTrip pins that the wire form of every valid detail is
// the string ParseDetail accepts back.
func TestDetailTokenRoundTrip(t *testing.T) {
	for _, tok := range AllTemplateTokens() {
		k := KindHarnessDefect
		if tok == TokenPathTraversal || tok == TokenNotFound || tok == TokenPreserveIntegrity {
			k = KindTemplateDeployFailure
		}
		d, err := ParseDetail(k, tok.Token())
		if err != nil {
			t.Fatalf("ParseDetail(%s, %q): %v", k, tok, err)
		}
		if d.Token() != tok.Token() {
			t.Fatalf("token round trip: %q -> %q", tok.Token(), d.Token())
		}
	}
}

// TestNoPathShapeInTokenForms guards the closed forms against path leakage at
// the type level: no token and no well-formed hook identity contains a path
// separator sequence or a home-directory shape.
func TestNoPathShapeInTokenForms(t *testing.T) {
	for _, tok := range AllTemplateTokens() {
		if strings.Contains(tok.Token(), "//") || strings.HasPrefix(tok.Token(), "/") {
			t.Fatalf("token %q carries a path shape", tok)
		}
	}
}
