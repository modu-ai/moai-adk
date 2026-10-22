package update

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

func createRestoreTemp(path string, mode os.FileMode) (*os.File, error) {
	return atomicfile.CreatePolicyTemp(path, ".moai-merge-*", mode, 0)
}

// Native permissions were established before the stage was exposed to data.
func chmodRestoreTemp(_ string, _ os.FileMode) error { return nil }
