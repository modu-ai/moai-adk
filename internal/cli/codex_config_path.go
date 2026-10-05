package cli

// codex_config_path.go — the separator seam between a host path and the form
// that goes into ~/.codex/config.toml (SPEC-CODEX-SKILL-PATH-SLASH-001).
//
// Why a seam exists at all, rather than a direct filepath.ToSlash call.
//
// The config's `path` value is published in FORWARD-SLASH form on every host.
// That is not cosmetic: this repository's parser reads the value verbatim and
// decodes no TOML escapes (internal/codexwiring/skills.go), so a path emitted
// with escapes would be an entry Codex reads correctly and every moai reader
// reads as a different, absent path — and the prune verb deletes entries it
// classifies as absolute and cannot stat. Publishing slashes keeps that
// divergence closed by construction instead of by a decoder being right.
//
// filepath.ToSlash and filepath.FromSlash rewrite filepath.Separator, and on
// this card's only executable host that separator is ALREADY '/'. Both are
// therefore the identity function here — measured, not assumed
// (`go run .moai/reports/t540/lab/ts.go` → GOOS=darwin Separator='/',
// changed=false in both directions). Calling them directly would leave every
// backslash-shaped test input untouched on darwin, so the Windows behaviour
// this file exists to fix would have no failing test to drive it and no
// passing test to keep it. Taking the separator as a parameter is what makes
// Windows semantics measurable on a '/' host.
//
// The shape follows the seams this package already uses — a package-level var
// plus explicit parameters (osStatFn, codexUserHomeDir) — rather than
// introducing an abstraction of its own.

import (
	"path/filepath"
	"strings"
)

// configPathSeparator is the injection point. Production call sites pass it;
// a test overrides it with a t.Cleanup restore and stays non-parallel, since
// a package-level var is shared state — the same discipline osStatFn already
// carries.
var configPathSeparator = filepath.Separator

// toConfigPath renders a host path in the config's slash form.
//
// [HARD] The conversion is SEPARATOR-AWARE, never an unconditional
// strings.ReplaceAll(p, "\\", "/"). A unix filename may legally contain a
// backslash, and such a path is refused today by the guard in
// upsertCodexSkillDisable — correctly, because the format cannot carry it
// verbatim. An unconditional replacement would turn that refusal into the
// publication of a wrong path, which is strictly worse than the defect being
// repaired. So when sep is already '/', the input is handed back untouched
// and the guard keeps refusing (REQ-CSPS-010).
func toConfigPath(p string, sep rune) string {
	if sep == '/' {
		return p
	}
	return strings.ReplaceAll(p, string(sep), "/")
}

// fromConfigPath renders a declared config path back in the host's own form,
// for handing to stat. It is the inverse of toConfigPath and is likewise the
// identity where sep is '/'.
func fromConfigPath(p string, sep rune) string {
	if sep == '/' {
		return p
	}
	return strings.ReplaceAll(p, "/", string(sep))
}
