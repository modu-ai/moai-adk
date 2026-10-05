// Package jevcred owns the single implementation of the TypeSafe API
// credential writer and reader for ~/.moai/.env.typesafe (REQ-JEVC-018).
//
// @MX:ANCHOR: [AUTO] TypeSafe credential SSOT — exactly one writer implementation
// @MX:REASON: two writers would mean two file-mode policies and two escaping rules; every consumer (internal/jev, the doctor check, and the later opt-in surfaces) MUST delegate here
//
// The package is modelled on internal/glmcred, deliberately and down to the
// two details that are easy to lose:
//
//   - Mode tightening on write. os.WriteFile's perm argument applies only at
//     file creation, so a pre-existing 0644 credential file stays 0644 without
//     an explicit Chmod. The explicit Chmod in Save closes that latent defect.
//   - The four-character disclosure floor. View returns Configured: true with
//     an EMPTY hint for a credential of four characters or fewer; a naive
//     "last four or the whole key" fallback would disclose a short credential
//     entirely, the exact inverse of REQ-JEVC-020.
//
// It depends only on the standard library and on stdlib-only internal leaf
// packages (internal/paths, internal/defs) so it cannot participate in an
// import cycle, and so both internal/cli and internal/web can import it while
// the one-way internal/cli → internal/web dependency stays acyclic.
//
// The credential is deliberately OUT of settings.AllFields(): no generic
// schema-walking loop (bulk value read, form-state dump, diagnostics view) can
// read, render, or write it (REQ-JEVC-019). schema_absence_test.go asserts it.
package jevcred

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// EnvTestTypeSafeKey is the env-var test seam honoured by Load. A non-empty
// value short-circuits the on-disk read and is returned verbatim, so tests can
// simulate a stored credential without touching ~/.moai/.env.typesafe. The
// constant string itself is owned by internal/config/envkeys.go
// (config.EnvTestTypeSafeKey) — this alias keeps the package stdlib-only,
// exactly as glmcred.EnvTestGLMKey does, without introducing a second literal.
const EnvTestTypeSafeKey = "MOAI_TEST_TYPESAFE_KEY"

// dotenvKey is the assignment name inside the credential file.
const dotenvKey = "TYPESAFE_API_KEY"

// hintLength is the number of trailing characters View may disclose, and the
// floor below which it discloses nothing at all (REQ-JEVC-020).
const hintLength = 4

// HomeDirFn resolves the operator's home directory. It is a function variable
// (mirroring glmcred.HomeDirFn) so tests can redirect it at a t.TempDir() and
// never touch the developer's real home directory — and so no test needs
// t.Setenv("HOME", ...), which pollutes parallel tests.
var HomeDirFn = paths.Home

// Path returns the absolute path to the TypeSafe credential file
// (~/.moai/.env.typesafe). A non-empty absolute MOAI_HOME redirects the file
// under the overridden root; otherwise the HomeDirFn seam resolves the home and
// the join consumes defs segments. It returns the empty string when resolution
// fails — never a relative fallback.
func Path() string {
	if v := os.Getenv(paths.EnvHome); v != "" && filepath.IsAbs(v) {
		root, err := paths.MoaiHome()
		if err != nil {
			return ""
		}
		return filepath.Join(root, defs.TypeSafeEnvFileName)
	}
	home, err := HomeDirFn()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, defs.MoAIDir, defs.TypeSafeEnvFileName)
}

// Save writes the TypeSafe API credential to ~/.moai/.env.typesafe at file mode
// 0600, in the quoted dotenv form Load reads back.
//
// On a pre-existing file whose mode is wider than 0600, the mode is tightened to
// 0600 as part of the write (REQ-JEVC-018 / AC-JEVC-011).
func Save(key string) error {
	envPath := Path()
	if envPath == "" {
		return fmt.Errorf("cannot determine home directory")
	}

	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	content := fmt.Sprintf("# TypeSafe API credential for MoAI-ADK\n# Written by internal/jevcred\n%s=\"%s\"\n",
		dotenvKey, EscapeValue(key))
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Explicit mode assertion: os.WriteFile's perm is honoured only at file
	// creation. An existing credential file created at a wider mode by an
	// older writer is narrowed here so the credential is never world-readable
	// after a save.
	if err := os.Chmod(envPath, 0o600); err != nil {
		return fmt.Errorf("chmod credential file: %w", err)
	}
	return nil
}

// Load reads the TypeSafe API credential from ~/.moai/.env.typesafe.
//
// EnvTestTypeSafeKey is honoured first (test seam) and short-circuits the
// on-disk read. When the credential file is absent, unreadable, or carries no
// entry, Load returns the empty string — an absence, never an error. That is
// the fail-open contract's first link (REQ-JEVC-005): the caller cannot turn
// this into a non-zero exit because there is nothing to propagate.
func Load() string {
	if testKey := os.Getenv(EnvTestTypeSafeKey); testKey != "" {
		return testKey
	}

	envPath := Path()
	if envPath == "" {
		return ""
	}

	file, err := os.Open(envPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		after, ok := strings.CutPrefix(line, dotenvKey+"=")
		if !ok {
			continue
		}
		value := after
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = UnescapeValue(value[1 : len(value)-1])
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
		return value
	}
	return ""
}

// Configured reports whether a credential is stored. It is the presence
// predicate the doctor check reads; it never discloses the value.
func Configured() bool { return Load() != "" }

// ViewHint carries the bounded disclosure a surface may show about a stored
// credential. The credential itself NEVER crosses into this value — only a
// Configured boolean and, for a credential longer than four characters, its
// final four characters (REQ-JEVC-020).
type ViewHint struct {
	Configured bool
	Hint       string
}

// View returns the bounded disclosure.
//
// For a credential longer than four characters, Hint is exactly the final four
// characters and nothing else. For a credential of four characters or fewer,
// Hint is EMPTY and only Configured is true — a naive "last four or the whole
// key" fallback would disclose a short credential entirely, which is the exact
// inverse of the requirement (AC-JEVC-014).
//
// There is deliberately no reveal route. The GLM precedent has one
// (glmKeyRevealPath); REQ-JEVC-020 is satisfied without it, so this package
// ships no path by which the stored credential leaves the process.
func View() ViewHint {
	key := Load()
	if key == "" {
		return ViewHint{Configured: false}
	}
	if len(key) <= hintLength {
		return ViewHint{Configured: true}
	}
	return ViewHint{Configured: true, Hint: key[len(key)-hintLength:]}
}

// EscapeValue escapes the characters that are special inside a dotenv
// double-quoted value: backslash, double-quote, and dollar. It deliberately
// does NOT escape newlines — a newline in a credential is malformed input a
// boundary validator rejects, not a value to round-trip.
func EscapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "$", "\\$")
	return value
}

// UnescapeValue reverses EscapeValue.
func UnescapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\$", "$")
	value = strings.ReplaceAll(value, "\\\"", "\"")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return value
}
