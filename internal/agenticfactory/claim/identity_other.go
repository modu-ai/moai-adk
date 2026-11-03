//go:build !unix

package claim

import "os"

// fileIdentity is not available here; the comparisons fall back to os.SameFile
// between two paths and to the canonical text for the reservation key.
func fileIdentity(os.FileInfo) (string, bool) { return "", false }
