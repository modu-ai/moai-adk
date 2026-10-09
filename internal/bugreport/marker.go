package bugreport

import (
	"errors"
	"fmt"
)

// internalMarker is the attribution marker (design.md section 3, row M2): an
// emit site that knows the failure is a moai internal defect wraps its error
// with MarkInternal, and the capture-time attribution finds the marker with
// errors.As. The marker is defined here so that every package that can name it
// can also import this package (bugreport depends on nothing but the standard
// library and internal/config).
type internalMarker struct{ cause error }

func (m *internalMarker) Error() string {
	return fmt.Sprintf("bugreport: internal defect: %v", m.cause)
}

func (m *internalMarker) Unwrap() error { return m.cause }

// MarkInternal wraps cause with the internal marker. Call sites use it where
// the tree's own invariant is violated (for example the preference store's
// "store is not *fileStore" branch) so attribution has an explicit,
// non-textual signal — error text is never read anywhere in this package.
func MarkInternal(cause error) error {
	if cause == nil {
		return nil
	}
	return &internalMarker{cause: cause}
}

// IsInternalMarker reports whether the error chain carries the internal
// marker, via errors.As — type inspection only, never text matching.
func IsInternalMarker(err error) bool {
	var m *internalMarker
	return errors.As(err, &m)
}
