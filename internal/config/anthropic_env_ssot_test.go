package config

// anthropic_env_ssot_test.go — SPEC-ENVKEY-ANTHROPIC-SSOT-001 (M2)
//
// Repo-root-scoped guard for the ANTHROPIC_* environment-variable name family.
// It asserts that no bare ANTHROPIC_* string literal survives in production Go
// source under internal/, pkg/, and cmd/ — every reference MUST go through the
// EnvAnthropic* constants in envkeys.go (REQ-EAS-004/005/006/007/008).
//
// Two properties distinguish this guard from the pre-existing
// TestNoBareGLMEnvVarLiteralsInCLIProduction in internal/cli:
//
//  1. Walk root is the REPOSITORY ROOT, not the package directory. The older
//     guard walks os.Getwd(), which under `go test` is the package dir, making
//     every file outside internal/cli structurally unreachable.
//  2. The banned set is DERIVED from the envkeys.go constants rather than
//     duplicated as bare literals, so adding a constant without adding it here
//     is a visible omission. Consequently this file contains zero bare
//     ANTHROPIC_* literals by design; set completeness is asserted at runtime
//     by TestAnthropicBannedSetCoversAllNames below.
//
// Exclusions: _test.go files (hardcoding-allowed zone) and the SSOT definition
// file internal/config/envkeys.go itself, which holds the constant definitions.
// The envkeys.go exclusion is an EXACT PATH match, never a package- or
// substring-scoped one: a broader exclusion would silently exempt the whole
// internal/config package.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// bannedAnthropicEnvNames is derived from the envkeys.go constants, so a new
// ANTHROPIC_* constant that is not added here is a visible omission rather than
// a silent gap (REQ-EAS-008).
var bannedAnthropicEnvNames = []string{
	EnvAnthropicPrefix,
	EnvAnthropicAPIKey,
	EnvAnthropicAuthToken,
	EnvAnthropicBaseURL,
	EnvAnthropicDefaultFableModel,
	EnvAnthropicDefaultHaikuModel,
	EnvAnthropicDefaultOpusModel,
	EnvAnthropicDefaultSonnetModel,
	EnvAnthropicReasoningEffort,
}

// anthropicScanRoots are the three ENUMERATED production roots this guard
// walks. They fail independently: a typo'd or omitted entry leaves that root
// unscanned while the others still report clean.
var anthropicScanRoots = []string{"internal", "pkg", "cmd"}

// TestNoBareAnthropicEnvVarLiteralsInProduction walks internal/, pkg/, and cmd/
// from the repository root and fails if any non-test production .go file
// contains a bare ANTHROPIC_* string literal. Offenders are reported one per
// line, each line carrying the offending file's repo-relative path.
func TestNoBareAnthropicEnvVarLiteralsInProduction(t *testing.T) {
	t.Parallel()

	bannedSet := make(map[string]bool, len(bannedAnthropicEnvNames))
	for _, b := range bannedAnthropicEnvNames {
		bannedSet[b] = true
	}

	root := findProjectRoot(t)
	// The single SSOT definition file, excluded by EXACT path.
	ssotFile := filepath.Clean(filepath.Join(root, "internal", "config", "envkeys.go"))

	var offenders []string
	for _, scanRoot := range anthropicScanRoots {
		base := filepath.Join(root, scanRoot)
		if _, err := os.Stat(base); err != nil {
			t.Fatalf("scan root %q not found under %s: %v", scanRoot, root, err)
		}
		_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if filepath.Clean(path) == ssotFile {
				return nil // the SSOT definition site itself
			}
			fset := token.NewFileSet()
			node, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if perr != nil {
				return nil // skip unparseable files; not the concern of this test
			}
			ast.Inspect(node, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				val := strings.Trim(lit.Value, `"`)
				if !bannedSet[val] {
					return true
				}
				pos := fset.Position(lit.Pos())
				rel, rerr := filepath.Rel(root, pos.Filename)
				if rerr != nil {
					rel = pos.Filename
				}
				offenders = append(offenders,
					filepath.ToSlash(rel)+":"+strconv.Itoa(pos.Line)+":"+strconv.Itoa(pos.Column)+": "+val)
				return true
			})
			return nil
		})
	}

	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("bare ANTHROPIC_* env-var string literal(s) found in production source "+
			"(must reference the config.EnvAnthropic* constants instead) — %d offender(s):\n  %s",
			len(offenders), strings.Join(offenders, "\n  "))
	}
}

// TestAnthropicBannedSetCoversAllNames asserts at runtime that the derived
// banned set holds exactly the 9 ANTHROPIC_* names defined in envkeys.go
// (REQ-EAS-008). It is a top-level test function — NOT a sub-test — because the
// acceptance verdict selects it with a bare top-level -run pattern.
//
// This assertion is the only coverage for the three names with zero
// out-of-internal/cli offenders, which no scan-based check can exercise.
func TestAnthropicBannedSetCoversAllNames(t *testing.T) {
	t.Parallel()

	const wantLen = 9
	if len(bannedAnthropicEnvNames) != wantLen {
		t.Fatalf("banned set size = %d, want %d", len(bannedAnthropicEnvNames), wantLen)
	}

	got := make(map[string]bool, len(bannedAnthropicEnvNames))
	for _, b := range bannedAnthropicEnvNames {
		got[b] = true
	}

	for _, tc := range []struct {
		name  string
		value string
	}{
		{"EnvAnthropicPrefix", EnvAnthropicPrefix},
		{"EnvAnthropicAPIKey", EnvAnthropicAPIKey},
		{"EnvAnthropicAuthToken", EnvAnthropicAuthToken},
		{"EnvAnthropicBaseURL", EnvAnthropicBaseURL},
		{"EnvAnthropicDefaultFableModel", EnvAnthropicDefaultFableModel},
		{"EnvAnthropicDefaultHaikuModel", EnvAnthropicDefaultHaikuModel},
		{"EnvAnthropicDefaultOpusModel", EnvAnthropicDefaultOpusModel},
		{"EnvAnthropicDefaultSonnetModel", EnvAnthropicDefaultSonnetModel},
		{"EnvAnthropicReasoningEffort", EnvAnthropicReasoningEffort},
	} {
		if !got[tc.value] {
			t.Errorf("banned set is missing %s (%q)", tc.name, tc.value)
		}
	}
}
