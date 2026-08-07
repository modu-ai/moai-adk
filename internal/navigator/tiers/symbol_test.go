package tiers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeGoFixture writes a Go source fixture under pkgPath within root.
func writeGoFixture(t *testing.T, root, pkgPath, filename, body string) {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(pkgPath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Initialize a minimal go.mod if absent so go/parser can resolve the file.
	if _, err := os.Stat(filepath.Join(root, "go.mod")); os.IsNotExist(err) {
		if err := os.WriteFile(filepath.Join(root, "go.mod"),
			[]byte("module fixt\n\ngo 1.23\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestSymbol_GoStructure_SignatureDeclRefs exercises AC-NS3-011: given a Go
// fixture package with a function ParseHeader, the deterministic structure
// extraction produces a per-symbol record carrying signature +
// declaration_path + declaration_line + ≥1 reference.
func TestSymbol_GoStructure_SignatureDeclRefs(t *testing.T) {
	root := t.TempDir()
	def := `package header

// ParseHeader extracts the bearer token.
func ParseHeader(h string) string {
	return h
}
`
	call := `package cli

import "fixt/header"

func run() string {
	return header.ParseHeader("Authorization")
}
`
	writeGoFixture(t, root, "internal/header", "header.go", def)
	writeGoFixture(t, root, "internal/cli", "cli.go", call)

	recs, err := extractGoStructures(root, SymbolOptions{})
	if err != nil {
		t.Fatalf("extractGoStructures error: %v", err)
	}
	var parseHeader *SymbolEnrichment
	for i := range recs {
		if recs[i].Identifier == "header.ParseHeader" {
			parseHeader = &recs[i]
			break
		}
	}
	if parseHeader == nil {
		t.Fatalf("ParseHeader record not emitted; got %d records", len(recs))
	}
	if parseHeader.Signature == "" {
		t.Errorf("Signature empty; want non-empty")
	}
	if parseHeader.DeclarationPath == "" {
		t.Errorf("DeclarationPath empty")
	}
	if parseHeader.DeclarationLine == 0 {
		t.Errorf("DeclarationLine=0; want the line of the func decl")
	}
	if len(parseHeader.References) == 0 {
		t.Errorf("References empty; want ≥1 caller (cli.go)")
	}
	// Identifier shape: <pkg-path-last-segment>.<name> (additive, deterministic).
	if parseHeader.Identifier != "header.ParseHeader" {
		t.Errorf("Identifier=%q; want header.ParseHeader", parseHeader.Identifier)
	}
}

// TestSymbol_Dispatch_NonGoLanguage_FailOpen exercises AC-NS3-013: a language
// without a configured indexer (anything but Go at M4) returns
// errIndexerNotConfigured and produces 0 records, exit 0.
func TestSymbol_Dispatch_NonGoLanguage_FailOpen(t *testing.T) {
	root := t.TempDir()
	recs, err := extractStructuresForLanguage(root, "typescript", SymbolOptions{})
	if err != errIndexerNotConfigured {
		t.Errorf("err = %v; want errIndexerNotConfigured", err)
	}
	if len(recs) != 0 {
		t.Errorf("non-Go language emitted %d records; want 0 (SCIP absent degrade)", len(recs))
	}
}

// TestSymbol_Dispatch_GoLanguage_ProducesRecords confirms the dispatch
// delegates to the Go path for "go".
func TestSymbol_Dispatch_GoLanguage_ProducesRecords(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "p", "p.go", "package p\nfunc F() {}\n")
	recs, err := extractStructuresForLanguage(root, "go", SymbolOptions{})
	if err != nil {
		t.Errorf("go dispatch error: %v", err)
	}
	if len(recs) == 0 {
		t.Errorf("go dispatch emitted 0 records; want ≥1")
	}
}

// TestSymbol_ReferencesCap exercises AC-NS3-014: the references list is
// capped at a configurable N to bound output size.
func TestSymbol_ReferencesCap(t *testing.T) {
	root := t.TempDir()
	// One function definition, many call sites.
	def := "package p\nfunc F() {}\n"
	writeGoFixture(t, root, "p", "p.go", def)
	// Generate 50 call sites in 50 files.
	for i := 0; i < 50; i++ {
		body := "package call\nimport \"fixt/p\"\nvar _ = p.F\n"
		writeGoFixture(t, root, "internal/call", "c"+itoa(i)+".go", body)
	}
	recs, err := extractGoStructures(root, SymbolOptions{MaxReferences: 5})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range recs {
		if r.Identifier != "p.F" {
			continue
		}
		if len(r.References) > 5 {
			t.Errorf("References count=%d; want ≤ 5 (cap)", len(r.References))
		}
	}
}

// TestSymbol_Narrative_LastUpdatedCommitGate exercises AC-NS3-012: the
// narrative is re-drafted ONLY when the deterministic record changed since
// last_updated_commit. ShouldRedraft returns true on first run (no metadata)
// and false when the deterministic hash matches the stored one.
func TestSymbol_Narrative_LastUpdatedCommitGate(t *testing.T) {
	root := t.TempDir()
	symDir := filepath.Join(root, ".moai", "project", "navigator", "symbols")
	if err := os.MkdirAll(symDir, 0o755); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(symDir, "p.F.metadata.json")

	// (1) No metadata → redraft.
	if !shouldRedraft(metadataPath, "hash-A") {
		t.Errorf("shouldRedraft=true on missing metadata; got false")
	}
	// (2) Write metadata with matching hash → no redraft.
	if err := writeNarrativeMetadata(metadataPath, "abc1234", "hash-A"); err != nil {
		t.Fatal(err)
	}
	if shouldRedraft(metadataPath, "hash-A") {
		t.Errorf("shouldRedraft=false on matching hash; got true")
	}
	// (3) Hash changed → redraft.
	if !shouldRedraft(metadataPath, "hash-B") {
		t.Errorf("shouldRedraft=true on changed hash; got false")
	}
}

// TestSymbol_TwoTierSeparable_DeterministicWithoutNarrative exercises
// AC-NS3-015: with the narrative path STUBBED/DISABLED, tiers emission
// still produces deterministic structure records with narrative_path empty,
// and exits 0.
func TestSymbol_TwoTierSeparable_DeterministicWithoutNarrative(t *testing.T) {
	root := t.TempDir()
	writeGoFixture(t, root, "p", "p.go", "package p\nfunc G() {}\n")
	recs, _, err := enumerateSymbols(root, SymbolOptions{NarrativeEnabled: false})
	if err != nil {
		t.Fatalf("enumerateSymbols error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatalf("deterministic records empty; want ≥1")
	}
	for _, r := range recs {
		if r.NarrativePath != "" {
			t.Errorf("NarrativePath=%q; want empty with narrative disabled", r.NarrativePath)
		}
		if r.Signature == "" || r.DeclarationPath == "" {
			t.Errorf("deterministic fields empty for %s: %+v", r.Identifier, r)
		}
	}
}

// TestSymbol_NarrativeMetadata_RoundTrip verifies the metadata.json sidecar
// is JSON and round-trips.
func TestSymbol_NarrativeMetadata_RoundTrip(t *testing.T) {
	root := t.TempDir()
	symDir := filepath.Join(root, ".moai", "project", "navigator", "symbols")
	if err := os.MkdirAll(symDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(symDir, "x.metadata.json")
	if err := writeNarrativeMetadata(p, "sha1", "h1"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m narrativeMetadata
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("metadata.json unparseable: %v", err)
	}
	if m.LastUpdatedCommit != "sha1" || m.LastRecordHash != "h1" {
		t.Errorf("round-trip wrong: %+v", m)
	}
}

// itoa is a tiny strconv.Itoa to avoid the import in this test file.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
