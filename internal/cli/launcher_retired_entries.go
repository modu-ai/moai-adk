package cli

// launcher_retired_entries.go holds the retired `-k` / `--kanban` entry
// (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-010). Kanban Mode was removed; the
// spelling stays recognized only so every launcher refuses it with one line
// that names the entries to use, before any branch resolves, with nothing
// launched and nothing written.
//
// The two token constants and the refusal carry names without the retired
// word, so this file is the one place in the launcher surface that spells it
// (REQ-017).

import (
	"errors"
	"strings"
)

// The retired entry tokens. `-k` is unbound on cc / glm / codex; the commands
// that do bind it (`doctor config dump`, `state`) are distinct and unaffected.
const (
	retiredLongFlag  = "--kanban"
	retiredShortFlag = "-k"
)

// retiredEntryRefusal is the one line every launcher prints for a retired
// entry: it states the mode is retired and names the replacement entries —
// `-f` starts a factory leader on cc and glm, `-l` joins as a lane on all three
// launchers (codex has no leader entry, so its lane entry is named in full).
const retiredEntryRefusal = "-k/--kanban is retired: Kanban Mode was removed; " +
	"start a factory leader with 'moai cc -f' or 'moai glm -f', and join as a lane with -l ('moai codex -l' on Codex)"

// refuseRetiredEntry returns the refusal when args carry a retired entry
// spelling (bare or `=`-joined) before the pass-through marker, and nil
// otherwise. The `--` discipline matches the other launcher parsers: nothing
// from the marker on is read, and the marker plus everything after it reaches
// the backend untouched.
func refuseRetiredEntry(args []string) error {
	for _, arg := range args {
		if arg == "--" {
			return nil
		}
		if arg == retiredShortFlag || arg == retiredLongFlag ||
			strings.HasPrefix(arg, retiredShortFlag+"=") || strings.HasPrefix(arg, retiredLongFlag+"=") {
			return errors.New(retiredEntryRefusal)
		}
	}
	return nil
}
