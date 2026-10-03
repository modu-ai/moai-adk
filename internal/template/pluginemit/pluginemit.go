// Package pluginemit generates the moai marketplace and the manifests of the
// derived moai core plugin (SPEC-PLUGIN-MARKETPLACE-001).
//
// RED stub: the types and entry points compile and return empty output, so
// the M1 tests fail at their assertions. The implementation lands in the
// GREEN commit.
package pluginemit

import "io/fs"

// EnvUpdate is the environment switch that flips the golden tests into
// regeneration mode (the maintainer path); unset or empty, they compare.
const EnvUpdate = "PLUGIN_EMIT_UPDATE"

// Options selects the inputs of one emission.
type Options struct {
	// Version is the version SSOT value, with or without a leading "v".
	Version string
	// MCPSource is the fs-relative path of the template .mcp.json.
	MCPSource string
}

// DefaultOptions returns the options for the real template tree.
func DefaultOptions() Options {
	return Options{}
}

// Publication is the deterministic output of one emission.
type Publication struct {
	// Files maps each repository-root-relative path (forward slashes) to its
	// bytes.
	Files map[string][]byte
}

// Emit produces the publication.
func Emit(fsys fs.FS, opts Options) (*Publication, error) {
	return &Publication{Files: map[string][]byte{}}, nil
}
