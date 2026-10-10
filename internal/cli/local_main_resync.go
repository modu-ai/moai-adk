// local_main_resync.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1: the
// fast-forward re-sync of the primary checkout's local main (REQ-LMF-014,
// plan §B3a).
//
// COMPILE-ONLY STUBS. This file exists so that local_main_resync_test.go
// compiles. Every function here returns "not implemented" and carries no
// behavior. No cobra command is registered: the verb name is
// `moai integration resync`, and its registration belongs to the GREEN stage.
package cli

import "errors"

// errLocalMainResyncNotImplemented is the stub's only result.
var errLocalMainResyncNotImplemented = errors.New("local-main resync: not implemented")

// localMainResyncFetch is the fetch seam of the re-sync (plan §B3a step 4). Its
// production default runs `git fetch origin main`. The stub returns the
// not-implemented error, and the tests replace it with a no-op.
var localMainResyncFetch = func(repoRoot string) error { return errLocalMainResyncNotImplemented }

// runLocalMainResync fast-forwards local main to origin/main inside the
// integration window and returns the report line. The stub reports nothing.
func runLocalMainResync(repoRoot string) (string, error) {
	return "", errLocalMainResyncNotImplemented
}
