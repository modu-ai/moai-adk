package cli

import (
	"errors"
	"reflect"
	"testing"
)

// saveActiveClaudeOptionModel captures the active option model and restores
// it at cleanup, so a test that injects a model cannot leak it into a later
// test in the package run.
func saveActiveClaudeOptionModel(t *testing.T) {
	t.Helper()
	prev := activeClaudeOptionModel
	t.Cleanup(func() { activeClaudeOptionModel = prev })
}

// fixtureClaudeHelp is a commander-shaped `claude --help` excerpt carrying
// one option per measured class, a two-alias line, and a synopsis so long
// that commander wraps its description onto the next line — the real 2.1.289
// layout this parser must read.
const fixtureClaudeHelp = `Claude Code

  Usage: claude [options] [prompt]

  Options:
    -v, --version                        display version number
    -d, --debug [filters]                enable debug mode with optional filters
    -p, --print                          print response and exit (non-interactive)
    -w, --worktree [name]                run in a worktree
        --settings <file...>             load settings from files
        --append-system-prompt <prompt>  append a system prompt to the default
        --teleport [target]              teleport to a session
        --remote-control-session-name-prefix <prefix>
                                         prefix for remote control session names
    -r, --resume [sessionId]             resume a conversation

  Commands:
    auth    manage authentication
`

// TestClaudeOptionModelDerivation (AC-SCV-012, REQ-SCV-012) — the option
// model is derived from fixture `--help` text injected through the seam, and
// every derivation failure degrades to the compile-time snapshot silently:
// no error escapes, and no test here spawns a process (the seam carries it).
func TestClaudeOptionModelDerivation(t *testing.T) {
	t.Run("fixture help classifies through the parser", func(t *testing.T) {
		m := parseClaudeOptionModel(fixtureClaudeHelp)
		want := claudeOptionModel{
			// required <value> synopses
			"--settings":                           claudeOptionRequiredValue,
			"--append-system-prompt":               claudeOptionRequiredValue,
			"--remote-control-session-name-prefix": claudeOptionRequiredValue,
			// optional [value] synopses
			"-d":         claudeOptionOptionalValue,
			"--debug":    claudeOptionOptionalValue,
			"-w":         claudeOptionOptionalValue,
			"--worktree": claudeOptionOptionalValue,
			"--teleport": claudeOptionOptionalValue,
			// measured booleans
			"-v":        claudeOptionBoolean,
			"--version": claudeOptionBoolean,
			"-p":        claudeOptionBoolean,
			"--print":   claudeOptionBoolean,
		}
		if !reflect.DeepEqual(m, want) {
			t.Fatalf("parseClaudeOptionModel(fixture) = %v, want %v", m, want)
		}
		// The resume tokens are special-cased by the walk before the model is
		// consulted (REQ-SCV-009), so the parser never carries them.
		if _, ok := m[resumeFlag]; ok {
			t.Errorf("--resume carried in the model; the walk special-cases it before the model")
		}
		if _, ok := m[resumeFlagShort]; ok {
			t.Errorf("-r carried in the model; the walk special-cases it before the model")
		}
	})

	t.Run("derivation installs the parsed model", func(t *testing.T) {
		saveActiveClaudeOptionModel(t)
		prevSynopsis := claudeHelpSynopsis
		claudeHelpSynopsis = func(string) (string, error) { return fixtureClaudeHelp, nil }
		t.Cleanup(func() { claudeHelpSynopsis = prevSynopsis })

		applyClaudeOptionModel("/fixture/claude")
		want := parseClaudeOptionModel(fixtureClaudeHelp)
		if !reflect.DeepEqual(activeClaudeOptionModel, want) {
			t.Fatalf("active model after derivation = %v, want the parsed fixture model", activeClaudeOptionModel)
		}
	})

	t.Run("failing derivation degrades to the snapshot silently", func(t *testing.T) {
		saveActiveClaudeOptionModel(t)
		prevSynopsis := claudeHelpSynopsis
		claudeHelpSynopsis = func(string) (string, error) { return "", errors.New("exec failed") }
		t.Cleanup(func() { claudeHelpSynopsis = prevSynopsis })

		applyClaudeOptionModel("/fixture/claude") // must not panic, must not swap
		if !reflect.DeepEqual(activeClaudeOptionModel, claudeOptionModelSnapshot) {
			t.Fatalf("active model changed on a failing derivation: %v", activeClaudeOptionModel)
		}
	})

	t.Run("help with no option lines degrades to the snapshot", func(t *testing.T) {
		saveActiveClaudeOptionModel(t)
		prevSynopsis := claudeHelpSynopsis
		claudeHelpSynopsis = func(string) (string, error) { return "Claude Code\n\n  auth    manage authentication\n", nil }
		t.Cleanup(func() { claudeHelpSynopsis = prevSynopsis })

		applyClaudeOptionModel("/fixture/claude")
		if !reflect.DeepEqual(activeClaudeOptionModel, claudeOptionModelSnapshot) {
			t.Fatalf("active model swapped to an empty parse: %v", activeClaudeOptionModel)
		}
	})
}
