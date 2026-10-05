package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

var errCGRetired = errors.New("moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration")

// cgCmd is an unregistered diagnostic boundary for legacy internal callers.
// It is never an alias for another launcher.
var cgCmd = &cobra.Command{Use: "cg", Short: "Retired CG launcher", Hidden: true, DisableFlagParsing: true, RunE: runCG}

func runCG(_ *cobra.Command, _ []string) error { return errCGRetired }
