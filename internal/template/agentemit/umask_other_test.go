//go:build !unix

package agentemit_test

import "testing"

// restrictUmask reports false where the platform has no process umask.
func restrictUmask(t *testing.T) bool {
	t.Helper()
	return false
}
