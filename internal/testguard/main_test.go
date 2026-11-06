package testguard

import (
	"fmt"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain scrubs repository-scoping git variables from the process and pins
// GIT_CONFIG_GLOBAL to a temp file (card t1801). Tests here run
// `git config --global` under RedirectHome; the scrub keeps their git children
// out of the caller's repository, and the pin keeps a test that forgets
// RedirectHome off the operator's real global config. RedirectHome uses
// t.Setenv, so it restores this value afterwards.
//
// Unlike internal/cli, this binary never re-executes itself, so the pin needs
// no inherited marker or reuse path.
func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: %v\n", err)
		os.Exit(1)
	}
	restore, err := pinGitConfigGlobal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	restore()
	os.Exit(code)
}

// pinGitConfigGlobal points GIT_CONFIG_GLOBAL at a temp file carrying a minimal
// [user] identity and returns the cleanup that removes the file.
func pinGitConfigGlobal() (func(), error) {
	f, err := os.CreateTemp("", "moai-testguard-gitconfig-")
	if err != nil {
		return nil, fmt.Errorf("create GIT_CONFIG_GLOBAL pin: %w", err)
	}
	path := f.Name()
	remove := func() { _ = os.Remove(path) }
	const body = "[user]\n\tname = moai-testguard-test\n\temail = moai-testguard-test@example.invalid\n"
	_, werr := f.WriteString(body)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Setenv(gitGlobalConfigEnv, path)
	}
	if werr != nil {
		remove()
		return nil, fmt.Errorf("set up GIT_CONFIG_GLOBAL pin: %w", werr)
	}
	return remove, nil
}
