package cli

// codex_config_path_test.go — SPEC-CODEX-SKILL-PATH-SLASH-001 (t540) M1.
//
// These tests measure Windows separator semantics on a host whose separator
// is '/'. That is possible only because the conversion takes the separator as
// a parameter rather than reading filepath.Separator directly: on this host
// filepath.ToSlash and filepath.FromSlash are both the IDENTITY function —
// measured, not assumed (spec.md M6, `go run .moai/reports/t540/lab/ts.go`
// prints `changed=false` for both). A test written against the bare stdlib
// call would therefore assert nothing here and would still read green.
//
// The pure functions take their separator as an argument, so these tests do
// NOT touch the package-level configPathSeparator var and are safe to keep
// free of the non-parallel discipline that seam-overriding tests need.

import (
	"path/filepath"
	"testing"
)

// TestConfigPathToConfigPathConvertsHostSeparator is the publisher half of
// the seam: a Windows-shaped path reaches the config in slash form, so no
// TOML escape is ever emitted and the verbatim-reading parser and Codex
// cannot disagree about which file the entry names.
func TestConfigPathToConfigPathConvertsHostSeparator(t *testing.T) {
	got := toConfigPath(`C:\Users\u\.codex\skills\probe\SKILL.md`, '\\')
	want := "C:/Users/u/.codex/skills/probe/SKILL.md"
	if got != want {
		t.Fatalf("toConfigPath(windows form, '\\\\') = %q, want %q", got, want)
	}
}

// TestConfigPathFromConfigPathRestoresHostSeparator is AC-CSPS-004 arm B: the
// reader half. A declared slash path is converted back to the host's own form
// before it is handed to stat.
func TestConfigPathFromConfigPathRestoresHostSeparator(t *testing.T) {
	got := fromConfigPath("C:/Users/u/SKILL.md", '\\')
	want := `C:\Users\u\SKILL.md`
	if got != want {
		t.Fatalf("fromConfigPath(slash form, '\\\\') = %q, want %q", got, want)
	}
}

// TestConfigPathIsIdentityOnSlashSeparator pins REQ-CSPS-006 and REQ-CSPS-010
// together, and it is the assertion that forbids an unconditional
// strings.ReplaceAll(p, "\\", "/").
//
// A unix filename may legally contain a backslash. Today such a path is
// REFUSED by the guard in upsertCodexSkillDisable, which is the correct
// outcome: the config format cannot carry it verbatim. An unconditional
// replacement would convert that refusal into the publication of a DIFFERENT,
// wrong path — strictly worse than the defect this card repairs. So when the
// separator is '/', both directions must hand back their input untouched,
// backslashes included.
func TestConfigPathIsIdentityOnSlashSeparator(t *testing.T) {
	inputs := []string{
		"/tmp/x/SKILL.md",
		"relative/path/SKILL.md",
		`/tmp/weird\name/SKILL.md`, // a legal unix filename carrying a backslash
		`C:\Users\u\SKILL.md`,      // a windows literal seen on a '/' host
		"",
	}
	for _, in := range inputs {
		if got := toConfigPath(in, '/'); got != in {
			t.Errorf("toConfigPath(%q, '/') = %q, want the input unchanged", in, got)
		}
		if got := fromConfigPath(in, '/'); got != in {
			t.Errorf("fromConfigPath(%q, '/') = %q, want the input unchanged", in, got)
		}
	}
}

// TestConfigPathRoundTripsUnderWindowsSeparator states the property the two
// functions exist to hold: what the publisher writes, the reader resolves back
// to the file the publisher was naming.
func TestConfigPathRoundTripsUnderWindowsSeparator(t *testing.T) {
	native := `C:\Users\u\.codex\skills\probe\SKILL.md`
	if got := fromConfigPath(toConfigPath(native, '\\'), '\\'); got != native {
		t.Fatalf("round trip = %q, want %q", got, native)
	}
}

// TestConfigPathSeparatorSeamTracksTheHost pins the injection point itself.
// Production call sites pass configPathSeparator; if that var were ever wired
// to a literal instead of the host separator, every production conversion
// would silently be measuring the wrong platform.
func TestConfigPathSeparatorSeamTracksTheHost(t *testing.T) {
	if configPathSeparator != filepath.Separator {
		t.Fatalf("configPathSeparator = %q, want filepath.Separator %q", configPathSeparator, filepath.Separator)
	}
}
