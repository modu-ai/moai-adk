package skillresidue

import "errors"

// RED stub (card t1733, repair round 4). The GREEN commit replaces this file with the
// per-platform implementations and moves the two seams into residue.go. Until then the
// primitive reports "not implemented" and the stage suffix is empty, so the new tests fail
// against the unchanged rename instead of failing to compile.

// errNoReplaceUnsupported is returned when the platform or the filesystem cannot rename
// without replacing.
var errNoReplaceUnsupported = errors.New("no-replace rename is not supported here")

func noReplaceRename(from, to string) error { return errors.New("noReplaceRename: not implemented") }

var (
	renameNoReplace = noReplaceRename
	stageSuffix     = func() string { return "" }
)
