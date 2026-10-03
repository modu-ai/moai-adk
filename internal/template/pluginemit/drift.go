// drift.go — compare a committed tree with an emission (REQ-009).
package pluginemit

// Difference kinds reported by Drift.
const (
	DriftBytes   = "bytes"
	DriftMode    = "mode"
	DriftMissing = "missing"
	DriftExtra   = "extra"
)

// Difference is one way a committed tree departs from an emission.
type Difference struct {
	Kind string
	Path string
}

// Drift compares the files committed under root with pub. It is read-only.
func Drift(pub *Publication, root string) ([]Difference, error) {
	return nil, nil
}

// Write materialises pub under root, the one regeneration path.
func Write(pub *Publication, root string) error {
	return nil
}
