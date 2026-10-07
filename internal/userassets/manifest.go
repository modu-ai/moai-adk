// manifest.go — the per-user manifest at ~/.moai/user-assets.json
// (SPEC-USER-ASSET-INSTALL-001 REQ-006/REQ-021; design §2.2).
//
// Schema (design §2.2): schema_version, bundles (the recorded opt-in
// selection), files{path → {sha256, bundle, installed_at, moai_version}},
// collisions[{path, first_seen_at}]. The installing moai version is PER FILE
// (REQ-006) — there is deliberately no top-level moai_version, because
// REQ-013's partial-failure continuation makes mixed-version states real.
//
// REQ-021 (iter4 D27): EVERY write — under a known or an unknown schema
// version — carries through fields the writing binary does not understand
// (top-level and per-file). The decode captures unknown fields as raw JSON
// and the encode re-emits them, so nothing is silently shrunk.
package userassets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SchemaVersion is the schema this binary reads and writes.
const SchemaVersion = 1

// Sentinel errors — callers branch on these (errors.Is).
var (
	// ErrPathInvalid rejects a root-relative path that is absolute, escapes
	// its root, or is empty.
	ErrPathInvalid = errors.New("userassets: path invalid")
	// ErrSchemaUnknown is the REQ-021 removal refusal against a foreign
	// schema version.
	ErrSchemaUnknown = errors.New("userassets: manifest schema_version unknown — manifest-driven removal refused")
	// ErrLocked is the user-level lock contention signal.
	ErrLocked = errors.New("userassets: manifest lock held by another run")
)

// CorruptError reports a manifest that exists but does not parse (AC-021
// second clause): removal refuses, the corruption is reported, a
// rebuild-from-scan is offered — never an auto-delete.
type CorruptError struct {
	Path string
	Err  error
}

func (e *CorruptError) Error() string {
	return fmt.Sprintf("userassets: manifest corrupt at %s: %v (removal refused; rebuild-from-scan offered)", e.Path, e.Err)
}

func (e *CorruptError) Unwrap() error { return e.Err }

// AsCorrupt reports whether err is a *CorruptError.
func AsCorrupt(err error, ce **CorruptError) bool {
	return errors.As(err, ce)
}

// FileEntry is one manifest-tracked file's record.
type FileEntry struct {
	SHA256      string `json:"sha256"`
	Bundle      string `json:"bundle"`
	InstalledAt string `json:"installed_at"`
	MoaiVersion string `json:"moai_version"`

	// installedByJournal marks a record completed by the journal
	// reconciliation within the current run (the recovery lattice's case 2 —
	// the retry claimed its own interrupted install). It is run-scoped
	// coordination state, never serialized.
	installedByJournal bool `json:"-"`

	// unknown carries fields this binary does not model (REQ-021); captured
	// on decode, re-emitted on encode.
	unknown map[string]json.RawMessage
}

// Collision records a user-created file found at an install target
// (REQ-010) — never overwritten, reported.
type Collision struct {
	Path        string `json:"path"`
	FirstSeenAt string `json:"first_seen_at"`
	unknown     map[string]json.RawMessage
}

// Manifest is the per-user install record.
type Manifest struct {
	SchemaVersion int                  `json:"schema_version"`
	Bundles       []string             `json:"bundles"`
	Files         map[string]FileEntry `json:"files"`
	Collisions    []Collision          `json:"collisions"`

	unknownTop map[string]json.RawMessage
}

// SchemaKnown reports whether the loaded schema_version is one this binary
// fully understands. Unknown schemas permit append-only install/refresh
// writes but refuse manifest-driven removal (REQ-021).
func (m *Manifest) SchemaKnown() bool { return m.SchemaVersion == SchemaVersion }

// CanRemove gates manifest-driven removal (REQ-009/REQ-021).
func (m *Manifest) CanRemove() error {
	if !m.SchemaKnown() {
		return fmt.Errorf("%w (read %d, know %d)", ErrSchemaUnknown, m.SchemaVersion, SchemaVersion)
	}
	return nil
}

// Load reads the manifest at path. Absent → a fresh, writable manifest and a
// nil error (the recovery lattice's "absent → install" case). Corrupt → a
// *CorruptError; the file is never deleted.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return freshManifest(), nil
		}
		return nil, fmt.Errorf("userassets: read manifest: %w", err)
	}
	m := freshManifest()
	if err := json.Unmarshal(data, m); err != nil {
		return nil, &CorruptError{Path: path, Err: err}
	}
	if m.Files == nil {
		m.Files = map[string]FileEntry{}
	}
	return m, nil
}

func freshManifest() *Manifest {
	return &Manifest{
		SchemaVersion: SchemaVersion,
		Files:         map[string]FileEntry{},
	}
}

// Save writes the manifest atomically, preserving unknown fields (REQ-021).
func (m *Manifest) Save(path string) error {
	if m.Files == nil {
		m.Files = map[string]FileEntry{}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("userassets: encode manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("userassets: mkdir manifest home: %w", err)
	}
	return atomicWrite(path, data, 0o644)
}

// UnmarshalJSON decodes the manifest, capturing unknown top-level fields.
func (m *Manifest) UnmarshalJSON(data []byte) error {
	type alias Manifest
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = Manifest(a)
	m.unknownTop = map[string]json.RawMessage{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch k {
		case "schema_version", "bundles", "files", "collisions":
		default:
			m.unknownTop[k] = v
		}
	}
	return nil
}

// MarshalJSON encodes the manifest, re-emitting unknown top-level fields.
func (m *Manifest) MarshalJSON() ([]byte, error) {
	type alias Manifest
	a := alias(*m)
	base, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return mergeUnknownFields(base, m.unknownTop)
}

// UnmarshalJSON decodes a file entry, capturing unknown per-file fields.
func (e *FileEntry) UnmarshalJSON(data []byte) error {
	type alias FileEntry
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*e = FileEntry(a)
	return captureUnknownFileFields(data, e)
}

func captureUnknownFileFields(data []byte, e *FileEntry) error {
	e.unknown = map[string]json.RawMessage{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch k {
		case "sha256", "bundle", "installed_at", "moai_version":
		default:
			e.unknown[k] = v
		}
	}
	return nil
}

// MarshalJSON encodes a file entry, re-emitting unknown fields. The value
// receiver is load-bearing: FileEntry rides a map, whose values are not
// addressable, so a pointer receiver would never fire during marshal and
// the unknown fields would drop (REQ-021).
func (e FileEntry) MarshalJSON() ([]byte, error) {
	type alias FileEntry
	a := alias(e)
	base, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return mergeUnknownFields(base, e.unknown)
}

// UnmarshalJSON decodes a collision record, capturing unknown fields.
func (c *Collision) UnmarshalJSON(data []byte) error {
	type alias Collision
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = Collision(a)
	c.unknown = map[string]json.RawMessage{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for k, v := range raw {
		switch k {
		case "path", "first_seen_at":
		default:
			c.unknown[k] = v
		}
	}
	return nil
}

// MarshalJSON encodes a collision record, re-emitting unknown fields.
func (c *Collision) MarshalJSON() ([]byte, error) {
	type alias Collision
	a := alias(*c)
	base, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return mergeUnknownFields(base, c.unknown)
}

// mergeUnknownFields overlays the captured unknown fields onto a base
// encoding. If a captured key collides with a modeled key the unknown value
// loses — the modeled field is this binary's truth.
func mergeUnknownFields(base []byte, unknown map[string]json.RawMessage) ([]byte, error) {
	if len(unknown) == 0 {
		return base, nil
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal(base, &merged); err != nil {
		return nil, err
	}
	for k, v := range unknown {
		if _, exists := merged[k]; !exists {
			merged[k] = v
		}
	}
	return json.Marshal(merged)
}

// atomicWrite writes data to path via a temp file in the destination
// directory plus a rename (the deployer's atomicWriteFile pattern).
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".userassets-*")
	if err != nil {
		return fmt.Errorf("userassets: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("userassets: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("userassets: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("userassets: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("userassets: rename into place: %w", err)
	}
	return nil
}
