//go:build !unix

package cli

import "testing"

// auditHygieneUnixCells is empty off Unix: Windows has no FIFOs.
func auditHygieneUnixCells(_ *testing.T, _ *leaderAnchorVerdict) {}
