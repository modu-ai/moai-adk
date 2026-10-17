//go:build !linux && !darwin && !windows

package skillresidue

// noReplaceRename reports that this platform has no rename that refuses to replace, so the
// held-back rename uses the link-based path.
func noReplaceRename(from, to string) error { return errNoReplaceUnsupported }
