package sync

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteGraph serializes g to JSON and writes it atomically to outPath
// (REQ-NS-010). The atomic-rename pattern (write to `outPath.tmp`, then
// `os.Rename`) guarantees no reader observes a partial file. The
// NAVIGATOR_PRE_RENAME_BARRIER environment variable is honored as a
// synchronized test hook: when set, WriteGraph writes "ready" to the barrier
// path after creating the .tmp file and blocks until the barrier is removed
// before the rename lands. Mirrors 003's `navigator_enrich.go:128 atomicWrite`.
func WriteGraph(outPath string, g Graph) error {
	data, err := marshalGraph(g)
	if err != nil {
		return err
	}
	return atomicWrite(outPath, data)
}

// marshalGraph serializes g to deterministic JSON (sorted nodes and edges).
func marshalGraph(g Graph) ([]byte, error) {
	return json.MarshalIndent(struct {
		Provenance Provenance `json:"provenance"`
		Nodes      []Node     `json:"nodes"`
		Edges      []Edge     `json:"edges"`
	}(g), "", "  ")
}

// atomicWrite writes data to <path>.tmp then renames it into place. It honors
// the NAVIGATOR_PRE_RENAME_BARRIER test hook (REQ-NS-010).
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	if barrier := os.Getenv("NAVIGATOR_PRE_RENAME_BARRIER"); barrier != "" {
		_ = os.Unsetenv("NAVIGATOR_PRE_RENAME_BARRIER")
		_ = os.WriteFile(barrier, []byte("ready"), 0o644)
		for {
			if _, err := os.Stat(barrier); err != nil {
				break
			}
		}
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s.tmp: %w", path, err)
	}
	return nil
}
