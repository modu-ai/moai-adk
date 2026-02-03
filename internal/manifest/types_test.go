package manifest

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestProvenanceStringValues(t *testing.T) {
	tests := []struct {
		name     string
		prov     Provenance
		expected string
	}{
		{"TemplateManaged", TemplateManaged, "template_managed"},
		{"UserModified", UserModified, "user_modified"},
		{"UserCreated", UserCreated, "user_created"},
		{"Deprecated", Deprecated, "deprecated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.prov) != tt.expected {
				t.Errorf("Provenance %s = %q, want %q", tt.name, string(tt.prov), tt.expected)
			}
		})
	}
}

func TestProvenanceIsValid(t *testing.T) {
	tests := []struct {
		name  string
		prov  Provenance
		valid bool
	}{
		{"TemplateManaged", TemplateManaged, true},
		{"UserModified", UserModified, true},
		{"UserCreated", UserCreated, true},
		{"Deprecated", Deprecated, true},
		{"EmptyString", Provenance(""), false},
		{"InvalidValue", Provenance("invalid"), false},
		{"CaseSensitive", Provenance("Template_Managed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prov.IsValid(); got != tt.valid {
				t.Errorf("Provenance(%q).IsValid() = %v, want %v", tt.prov, got, tt.valid)
			}
		})
	}
}

func TestProvenanceJSONRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		prov Provenance
		json string
	}{
		{"TemplateManaged", TemplateManaged, `"template_managed"`},
		{"UserModified", UserModified, `"user_modified"`},
		{"UserCreated", UserCreated, `"user_created"`},
		{"Deprecated", Deprecated, `"deprecated"`},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_marshal", func(t *testing.T) {
			data, err := json.Marshal(tt.prov)
			if err != nil {
				t.Fatalf("json.Marshal(%v) error: %v", tt.prov, err)
			}
			if string(data) != tt.json {
				t.Errorf("json.Marshal(%v) = %s, want %s", tt.prov, data, tt.json)
			}
		})

		t.Run(tt.name+"_unmarshal", func(t *testing.T) {
			var got Provenance
			if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) error: %v", tt.json, err)
			}
			if got != tt.prov {
				t.Errorf("json.Unmarshal(%s) = %v, want %v", tt.json, got, tt.prov)
			}
		})
	}
}

func TestManifestJSONRoundtrip(t *testing.T) {
	original := &Manifest{
		Version:    "1.14.0",
		DeployedAt: "2026-02-03T10:30:00Z",
		Files: map[string]FileEntry{
			".claude/agents/moai/expert-backend.md": {
				Provenance:   TemplateManaged,
				TemplateHash: "sha256:a1b2c3d4",
				DeployedHash: "sha256:a1b2c3d4",
				CurrentHash:  "sha256:a1b2c3d4",
			},
			"CLAUDE.md": {
				Provenance:   UserModified,
				TemplateHash: "sha256:e5f6g7h8",
				DeployedHash: "sha256:e5f6g7h8",
				CurrentHash:  "sha256:i9j0k1l2",
			},
		},
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent error: %v", err)
	}

	if !json.Valid(data) {
		t.Fatal("serialized manifest is not valid JSON")
	}

	var restored Manifest
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if restored.Version != original.Version {
		t.Errorf("Version = %q, want %q", restored.Version, original.Version)
	}
	if restored.DeployedAt != original.DeployedAt {
		t.Errorf("DeployedAt = %q, want %q", restored.DeployedAt, original.DeployedAt)
	}
	if len(restored.Files) != len(original.Files) {
		t.Fatalf("Files count = %d, want %d", len(restored.Files), len(original.Files))
	}

	for path, origEntry := range original.Files {
		resEntry, ok := restored.Files[path]
		if !ok {
			t.Errorf("missing file entry for %q", path)
			continue
		}
		if resEntry.Provenance != origEntry.Provenance {
			t.Errorf("Files[%q].Provenance = %v, want %v", path, resEntry.Provenance, origEntry.Provenance)
		}
		if resEntry.TemplateHash != origEntry.TemplateHash {
			t.Errorf("Files[%q].TemplateHash = %q, want %q", path, resEntry.TemplateHash, origEntry.TemplateHash)
		}
		if resEntry.DeployedHash != origEntry.DeployedHash {
			t.Errorf("Files[%q].DeployedHash = %q, want %q", path, resEntry.DeployedHash, origEntry.DeployedHash)
		}
		if resEntry.CurrentHash != origEntry.CurrentHash {
			t.Errorf("Files[%q].CurrentHash = %q, want %q", path, resEntry.CurrentHash, origEntry.CurrentHash)
		}
	}
}

func TestFileEntryFieldCompleteness(t *testing.T) {
	entry := FileEntry{
		Provenance:   TemplateManaged,
		TemplateHash: "sha256:abc123",
		DeployedHash: "sha256:def456",
		CurrentHash:  "sha256:ghi789",
	}

	if entry.Provenance == "" {
		t.Error("Provenance should not be empty")
	}
	if entry.TemplateHash == "" {
		t.Error("TemplateHash should not be empty")
	}
	if entry.DeployedHash == "" {
		t.Error("DeployedHash should not be empty")
	}
	if entry.CurrentHash == "" {
		t.Error("CurrentHash should not be empty")
	}
}

func TestChangedFileStruct(t *testing.T) {
	cf := ChangedFile{
		Path:       ".claude/settings.json",
		OldHash:    "sha256:old",
		NewHash:    "sha256:new",
		Provenance: UserModified,
	}

	data, err := json.Marshal(cf)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var restored ChangedFile
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if restored.Path != cf.Path {
		t.Errorf("Path = %q, want %q", restored.Path, cf.Path)
	}
	if restored.OldHash != cf.OldHash {
		t.Errorf("OldHash = %q, want %q", restored.OldHash, cf.OldHash)
	}
	if restored.NewHash != cf.NewHash {
		t.Errorf("NewHash = %q, want %q", restored.NewHash, cf.NewHash)
	}
	if restored.Provenance != cf.Provenance {
		t.Errorf("Provenance = %v, want %v", restored.Provenance, cf.Provenance)
	}
}

func TestNewManifest(t *testing.T) {
	m := NewManifest()
	if m == nil {
		t.Fatal("NewManifest() returned nil")
	}
	if m.Files == nil {
		t.Fatal("NewManifest().Files is nil, want initialized map")
	}
	if len(m.Files) != 0 {
		t.Errorf("NewManifest().Files length = %d, want 0", len(m.Files))
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrManifestNotFound", ErrManifestNotFound, "manifest: file not found"},
		{"ErrManifestCorrupt", ErrManifestCorrupt, "manifest: JSON parse error"},
		{"ErrEntryNotFound", ErrEntryNotFound, "manifest: entry not found"},
		{"ErrHashMismatch", ErrHashMismatch, "manifest: hash verification failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.msg)
			}
		})
	}

	// Verify errors.Is works with wrapped errors
	t.Run("ErrorsIs_wrapping", func(t *testing.T) {
		wrapped := errors.New("outer: " + ErrManifestCorrupt.Error())
		// Direct sentinel comparison
		if !errors.Is(ErrManifestCorrupt, ErrManifestCorrupt) {
			t.Error("errors.Is(ErrManifestCorrupt, ErrManifestCorrupt) should be true")
		}
		// Different errors should not match
		if errors.Is(wrapped, ErrManifestCorrupt) {
			t.Error("newly created error should not match sentinel via errors.Is")
		}
	})
}

func TestManifestEmptyFilesJSON(t *testing.T) {
	m := NewManifest()
	m.Version = "1.0.0"
	m.DeployedAt = "2026-02-03T00:00:00Z"

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent error: %v", err)
	}

	if !json.Valid(data) {
		t.Fatal("empty manifest is not valid JSON")
	}

	var restored Manifest
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if len(restored.Files) != 0 {
		t.Errorf("restored Files length = %d, want 0", len(restored.Files))
	}
}
