package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestGeneratedManagedPartsRoundTrip covers the per-part ownership record a
// generator stores (design §A.1): the new provenance value is valid, parts
// survive a JSON round trip, and an entry without parts serializes exactly
// as before (no "parts" key).
func TestGeneratedManagedPartsRoundTrip(t *testing.T) {
	if !GeneratedManaged.IsValid() {
		t.Fatal("generated_managed must be a valid provenance")
	}
	legacy, err := json.Marshal(FileEntry{Provenance: TemplateManaged, CurrentHash: "h"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacy), "parts") {
		t.Fatalf("an entry without parts must not serialize a parts key: %s", legacy)
	}
	in := FileEntry{Provenance: GeneratedManaged, Parts: []Part{
		{Kind: PartTOMLTable, Key: "mcp_servers.moai", Origin: OriginCreated, Hash: "abc", Region: "\n[mcp_servers.moai]\n"},
		{Kind: PartJSONKey, Key: "description", Origin: OriginUnknown},
	}}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out FileEntry
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Parts) != 2 || out.Parts[0] != in.Parts[0] || out.Parts[1] != in.Parts[1] {
		t.Fatalf("parts round trip: got %+v want %+v", out.Parts, in.Parts)
	}
}
