package cli

// AC-ITI-014 (REQ-ITI-013) — profileSetupText half: every field of the struct
// is referenced at least once in non-test code. Fields the absorbed wizard
// stopped reading are deleted, not carried.

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// keySweepCorpus concatenates the cli package's non-test .go files, EXCLUDING
// profile_setup_translations.go — the seed assignments would otherwise count
// as their own references (a seed is a definition, not a read).
func keySweepCorpus(t *testing.T) string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the cli package dir: %v", err)
	}
	var b strings.Builder
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || name == "profile_setup_translations.go" || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		b.Write(raw)
		b.WriteString("\n")
	}
	return b.String()
}

// TestProfileTextFields_Referenced is AC-ITI-014's profileSetupText half:
// every field of the struct is read (selector form) in non-test code at least
// once. Fields the absorbed wizard stopped reading are deleted, not carried.
func TestProfileTextFields_Referenced(t *testing.T) {
	corpus := keySweepCorpus(t)
	typ := reflect.TypeOf(profileSetupText{})
	if typ.NumField() == 0 {
		t.Fatal("profileSetupText has no fields; the sweep went stale")
	}
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if !strings.Contains(corpus, "."+name) {
			t.Errorf("profileSetupText.%s is never read in non-test code — delete the key and its seeds (REQ-ITI-013)", name)
		}
	}
}
