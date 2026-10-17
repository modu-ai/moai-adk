//go:build windows

package cli

import "os/exec"

// bisectRunPrepare is a no-op on Windows: a cancelled context ends only the
// direct child, so a judged command's own children can outlive the search.
func bisectRunPrepare(_ *exec.Cmd) {}
