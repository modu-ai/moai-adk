package hook

// protected_zone_shell_repro_test.go — RED reproduction for card t1585
// (follow-up to SPEC-HOOK-ZONE-BACKSLASH-001 / card t1570): three defects in
// the ANSI-C ($'...') half of the protected-zone shell guard, all still
// unreproduced when the card was dispatched — RED comes first, no repair
// without it.
//
// The bash ground truth each row rests on was measured with real bash
// before this file was written (card tree, 2026-10-08):
//
//   $'a\x00b'X          -> 61 58              ("aX": the ANSI-C part
//                                               truncates at the NUL, later
//                                               word parts still append)
//   zone_dir$'\x00/sub' -> 7a..72             ("zone_dir")
//   $'a\0b'             -> 61                 (octal NUL truncates too)
//   $'\xec\xa1\x80'     -> ec a1 80           (\xHH is ONE RAW BYTE, not a
//                                               re-encoded code point)
//   $'⊇'           -> (decoder: e2 8a 87) — the DECODER renders code
//                                               points per Go string(rune);
//                                               this host's bash 3.2.57
//                                               expands NEITHER \u NOR \U
//                                               (both stay literal —
//                                               decisive 2026-10-09, see
//                                               the card's evidence record;
//                                               an earlier "expansion"
//                                               reading was glyph-poisoned:
//                                               the probe had carried the
//                                               glyph, not the escape text)
//
//   ① P1 NUL truncation: zoneUnescapeAnsiC keeps the NUL and everything
//      after it, so the guard checks `zone_dir\x00/sub` against the zone
//      while the shell removes `zone_dir` itself — the protected directory
//      falls to rm (decision allow, measured pre-repair).
//   ② P2 \x byte vs rune: zoneHexEscape renders string(rune(val)), so every
//      \x byte ≥ 0x80 re-encodes as two UTF-8 bytes and a raw-byte spelling
//      of a protected path decodes to text no zone entry matches.
//   ③ P2 no-digit hex: with no hex digit after \x/\u/\U, zoneHexEscape
//      returns v[i:i+2] with i at the LAST index — slice-bounds panic when
//      the escape ends the string — and when a non-digit follows, the return
//      drops the backslash and the loop re-emits that character (doubled).
//      bash renders $\x' (and $\u', $\U') literally.
//
// The rows carry positive controls so a fix cannot pass by blanket-denying
// NUL-bearing or raw-byte commands (verification-completeness §2 mutant
// probe). The demonstration branches exec the real bash — the same authority
// the guard judged — so a red shows the actual landing, not an inference.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"mvdan.cc/sh/v3/syntax"
)

// hzsShellManifest declares the protected zone_dir directory under the probe
// category, the same fixture shape the t1570 repro rows use.
const hzsShellManifest = "  probe_zone:\n    paths: [\"zone_dir/\"]\n"

// hzsHangulDir is the protected directory whose name needs the multi-byte
// UTF-8 spelling the \x raw-byte confusion hides. The hex-spelling row derives
// its $'\xNN' escapes from THESE bytes (measured: this file's spelling of the
// name carries ec a1 b4) so the word the guard decodes and the directory the
// test created are the same bytes by construction — the bypass rows never
// depend on reading a hangul syllable correctly.
const hzsHangulDir = "존"

// hzsRunBashInRoot executes cmd under real bash with cwd=root and reports the
// combined output. The demonstration branches use it to show where the OS
// actually lands the command the guard allowed.
func hzsRunBashInRoot(t *testing.T, root, cmd string) (string, error) {
	t.Helper()
	c := exec.Command("bash", "-c", cmd)
	c.Dir = root
	out, err := c.CombinedOutput()
	return string(out), err
}

// hzsHexWord spells s as one ANSI-C word of \xNN escapes — the raw-byte form
// bash emits byte for byte.
func hzsHexWord(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		fmt.Fprintf(&b, `\x%02x`, s[i])
	}
	return b.String()
}

// TestCheckProtectedZoneShellAnsiCNulTruncationBypass — defect ①: a word
// whose ANSI-C part embeds a NUL must be judged by what the shell receives —
// the part truncated at the NUL, later parts still appended. Pre-repair the
// guard decodes the NUL and the text after it, sees `zone_dir\x00/sub`, and
// allows a command whose real target is `zone_dir` itself. The demonstration
// branch runs the allowed command under real bash and shows the protected
// directory disappearing. The mutant-probe control (the same NUL shape
// truncating OUTSIDE the zone must stay allowed) is its own test below, so
// its pre-repair green is recorded even while this row returns early on red
// (review gate, card t1585).
func TestCheckProtectedZoneShellAnsiCNulTruncationBypass(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the demonstration runs bash ANSI-C quoting")
	}
	swept := 0

	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "zone_dir", "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)

	const nulCmd = `rm -r zone_dir$'\x00/sub'`
	swept++
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": nulCmd})
	if d != DecisionDeny {
		// RED demonstration: the guard let the command through, so the shell
		// it protects runs it — bash truncates the ANSI-C part at the NUL and
		// rm lands on zone_dir itself.
		out, err := hzsRunBashInRoot(t, root, nulCmd)
		if err == nil {
			if _, statErr := os.Stat(filepath.Join(root, "zone_dir")); os.IsNotExist(statErr) {
				t.Errorf("BYPASS — decision=%q reason=%q, want deny; real bash ran the allowed command and the protected directory is GONE (rm output: %q)", d, r, out)
				swept--
				t.Logf("swept=%d", swept)
				return
			}
			t.Errorf("%q: decision=%q reason=%q, want deny; the demonstration run did not remove zone_dir (output %q)", nulCmd, d, r, out)
		} else {
			t.Errorf("%q: decision=%q reason=%q, want deny; and the demonstration run failed: %v (%q)", nulCmd, d, r, err, out)
		}
		swept--
		t.Logf("swept=%d", swept)
		return
	}
	wantZoneDeny(t, "nul truncation", d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("nul truncation: deny observed but the protected marker is gone: %v", err)
	}

	if swept != 1 {
		t.Fatalf("swept %d rows, want exactly 1 (the bypass row; the mutant-probe control lives in its own test)", swept)
	}
	t.Logf("swept=%d", swept)
}

// TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl —
// AC-HZS-002 (mutant probe): the same NUL shape truncating to a path OUTSIDE
// the zone stays allowed, pre- AND post-repair — a fix that blanket-denies
// every NUL-bearing command fails HERE, not on the bypass row. Independent
// of the bypass test so the control's pre-repair green is measured and
// recorded even while the bypass row returns early on red.
func TestCheckProtectedZoneShellAnsiCNulTruncationOutsideZoneControl(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the demonstration runs bash ANSI-C quoting")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": `rm -r docs$'\x00/zone_dir'`})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("positive control outside-zone truncation: decision=%q reason=%q, want allowed — the truncation lands on docs, not the zone", d, r)
	}
	t.Logf("swept=%d", 1)
}

// hzsHangulManifest declares the protected directory hzsHangulDir under the
// probe category — one builder so the manifest and the created directory are
// the same bytes by construction.
func hzsHangulManifest() string {
	return fmt.Sprintf("  probe_zone:\n    paths: [\"%s/\"]\n", hzsHangulDir)
}

// TestCheckProtectedZoneShellHexRawByteBypass — defect ②: \xHH emits one raw
// byte (measured), so a protected path spelled in raw UTF-8 bytes decodes —
// under the current rune rendering — to re-encoded text no zone entry
// matches. The direct-spelling control (the same target written literally
// must stay denied) is its own test below, so its pre-repair green is
// recorded even while this row returns early on red (review gate, card
// t1585).
func TestCheckProtectedZoneShellHexRawByteBypass(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the demonstration runs bash ANSI-C quoting")
	}
	swept := 0

	root := newZoneRoot(t, zoneShippedDoc(hzsHangulManifest()), "")
	if err := os.MkdirAll(filepath.Join(root, hzsHangulDir), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, hzsHangulDir, "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)

	byteCmd := `rm $'` + hzsHexWord(hzsHangulDir) + `/marker.md'`
	swept++
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": byteCmd})
	if d != DecisionDeny {
		// RED demonstration: the guard let it through — real bash emits the
		// raw bytes 존/marker.md and removes the protected file.
		out, err := hzsRunBashInRoot(t, root, byteCmd)
		if err == nil {
			if _, statErr := os.Stat(marker); os.IsNotExist(statErr) {
				t.Errorf("BYPASS — decision=%q reason=%q, want deny; real bash ran the allowed command and the protected marker is GONE (rm output: %q)", d, r, out)
				swept--
				t.Logf("swept=%d", swept)
				return
			}
			t.Errorf("%q: decision=%q reason=%q, want deny; the demonstration run did not remove the marker (output %q)", byteCmd, d, r, out)
		} else {
			t.Errorf("%q: decision=%q reason=%q, want deny; and the demonstration run failed: %v (%q)", byteCmd, d, r, err, out)
		}
		swept--
		t.Logf("swept=%d", swept)
		return
	}
	wantZoneDeny(t, "hex raw byte", d, r, harnessLearnerIdentity, "category", "probe_zone")

	if swept != 1 {
		t.Fatalf("swept %d rows, want exactly 1 (the bypass row; the direct-spelling control lives in its own test)", swept)
	}
	t.Logf("swept=%d", swept)
}

// TestCheckProtectedZoneShellHexDirectSpellingControl — AC-HZS-004
// (control): the same protected target spelled directly is denied on the
// pre-repair tree too — the manifest and the folding are not the defect;
// defect ② is confined to the decoder. Independent of the bypass test so
// the control's pre-repair deny is measured and recorded even while the
// bypass row returns early on red.
func TestCheckProtectedZoneShellHexDirectSpellingControl(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the demonstration runs bash ANSI-C quoting")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsHangulManifest()), "")
	if err := os.MkdirAll(filepath.Join(root, hzsHangulDir), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, hzsHangulDir, "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": `rm ` + hzsHangulDir + `/marker.md`})
	wantZoneDeny(t, "direct spelling control", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestZoneUnescapeAnsiCNoDigitHexStaysLiteral — defect ③ at the decoder
// level: an escape with no hex digit keeps its backslash and its prefix
// letter literally (bash's rendering), the following text survives once. The
// pre-repair decoder panics when the escape ends the string and mangles the
// text after a non-digit. Guard-level consequence (not executed here — a
// panic must not escape into the guard's walk): any command carrying $'\x'
// would crash the analysis instead of being judged.
func TestZoneUnescapeAnsiCNoDigitHexStaysLiteral(t *testing.T) {
	rows := []struct{ in, want string }{
		{`\x`, `\x`},
		{`\u`, `\u`},
		{`\U`, `\U`},
		{`\xZ`, `\xZ`},
		{`a\xZb`, `a\xZb`},
	}
	swept := 0
	for _, row := range rows {
		swept++
		got, panicked := hzsDecodeAnsiC(row.in)
		if panicked {
			t.Errorf("%q: PANICKED — slice bounds out of range on the no-digit escape", row.in)
			continue
		}
		if got != row.want {
			t.Errorf("%q: decoded %q, want the bash literal %q", row.in, got, row.want)
		}
	}
	if swept != len(rows) {
		t.Fatalf("swept %d rows, want exactly %d", swept, len(rows))
	}
	t.Logf("swept=%d", swept)
}

// hzsDecodeAnsiC calls zoneUnescapeAnsiC with the panic contained, so one
// defect row cannot abort the whole test binary.
func hzsDecodeAnsiC(v string) (out string, panicked bool) {
	defer func() {
		if rec := recover(); rec != nil {
			out, panicked = "", true
		}
	}()
	return zoneUnescapeAnsiC(v), false
}

// hzsGuardShell drives the guard's shell half with the panic contained (the
// hzsDecodeAnsiC pattern extended to the guard call): the analysis walk
// carries no recover of its own, so pre-fix a no-digit escape in the command
// panics the walk; post-fix a decision comes back (an empty string is the
// allow decision).
func hzsGuardShell(h *preToolHandler, agent, command string) (decision string, panicked bool) {
	defer func() {
		if rec := recover(); rec != nil {
			decision, panicked = "", true
		}
	}()
	raw, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		return "", false
	}
	return h.checkProtectedZoneShell(agent, raw), false
}

// TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape — AC-HZS-006: a
// real Bash call carrying a no-digit escape must produce a DECISION — the
// analysis walk runs to completion and a decoder panic never escapes into
// the guard. Pre-fix the walk panics (contained here by hzsGuardShell, so
// the row reports the panic instead of crashing the binary); post-fix the
// row observes the decision.
func TestCheckProtectedZoneShellGuardCompletesOnNoDigitEscape(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	const panicCmd = `rm -r zone_dir$'\x'/sub`
	d, panicked := hzsGuardShell(h, harnessLearnerIdentity, panicCmd)
	if panicked {
		t.Errorf("guard walk PANICKED on the no-digit escape in %q — the walk must complete and return a decision", panicCmd)
		return
	}
	t.Logf("decision=%q (no panic — the walk completed)", d)
	t.Logf("swept=%d", 1)
}

// TestZoneUnescapeAnsiCCodePointRenderingPinned — AC-HZS-007: the decoder
// renders code points as UTF-8 (string(rune(val))), pinned POST-repair so a
// mutant of the \\x split cannot quietly re-render \\u as a raw byte. The
// rows feed the ESCAPE TEXTS — byte 0x5C followed by the ASCII characters —
// not the rendered character: a literal glyph would ride the backslash-free
// early path and pin nothing. A maxDigits 4→2 mutant on \\u returns a
// different byte sequence and fails here. In Go source the six-byte escape
// text is written "\\u2287" — the doubled backslash is source syntax for the
// single 0x5C byte at runtime.
func TestZoneUnescapeAnsiCCodePointRenderingPinned(t *testing.T) {
	rows := []struct{ in, name string }{
		{"\\u2287", `escape text 0x5C u2287`},
		{"\\U00002287", `escape text 0x5C U00002287`},
	}
	want := string(rune(0x2287)) // the three UTF-8 bytes e2 8a 87
	swept := 0
	for _, row := range rows {
		swept++
		got, panicked := hzsDecodeAnsiC(row.in)
		if panicked {
			t.Errorf("%s: PANICKED — the code-point pin must decode without a panic", row.name)
			continue
		}
		if got != want {
			t.Errorf("%s: decoded % x, want % x (string(rune(0x2287)))", row.name, got, want)
		}
	}
	if swept != len(rows) {
		t.Fatalf("swept %d rows, want exactly %d", swept, len(rows))
	}
	t.Logf("swept=%d", swept)
}

// TestZoneWordTextAnsiCPartTruncatesAtNul — AC-HZS-009: the part-level NUL
// terminator shape. Bash assembles $'a\x00b'X as "aX" — the ANSI-C part's
// contribution ends at its first NUL byte and the LATER part still appends.
// A word-level truncation reads the same "aX" on the command rows while
// wrongly dropping the X here, which is exactly the mutant this row fails.
// Pre-repair the decoder keeps the NUL and the text after it, so the
// assembled word is "a" NUL "bX".
func TestZoneWordTextAnsiCPartTruncatesAtNul(t *testing.T) {
	file, ok := zoneParse(`rm $'a\x00b'X`)
	if !ok {
		t.Fatal("parse failed")
	}
	call, ok := file.Stmts[0].Cmd.(*syntax.CallExpr)
	if !ok {
		t.Fatalf("stmt[0] is %T, want *syntax.CallExpr", file.Stmts[0].Cmd)
	}
	got, literal := zoneWordText(call.Args[1])
	if !literal {
		t.Fatalf("word $'a\\x00b'X is not fully literal")
	}
	if got != "aX" {
		t.Errorf("zoneWordText($'a\\x00b'X) = %q, want %q — the ANSI-C part ends at its first NUL byte and the later part still appends", got, "aX")
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellOctalNulTruncationDenied — AC-HZS-010: the same
// NUL spelled in OCTAL ($'\0') must reach the same truncation — the octal
// escape renders byte(0) through zoneOctalEscape, so the semantics cannot
// depend on the escape's origin (bash: $'a\0b' -> "a"). Pre-repair the
// decoder keeps the NUL and the trailing text and the guard allows a command
// whose real target is zone_dir itself.
func TestCheckProtectedZoneShellOctalNulTruncationDenied(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "zone_dir", "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": `rm -r zone_dir$'\0/sub'`})
	wantZoneDeny(t, "octal nul truncation", d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("octal nul truncation: deny observed but the protected marker is gone: %v", err)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed — AC-HZS-011
// (control): a non-ASCII-named file OUTSIDE the zone stays ALLOWED in both
// the literal and the raw-byte spelling, pre- and post-repair — a fix (or
// mutant) that blanket-denies non-ASCII or raw-byte-bearing commands fails
// HERE. The raw-byte row derives its escapes from the declared name's bytes
// by construction (hzsHexWord), the same derivation AC-HZS-003 uses, so the
// decoded word and the created file are the same bytes.
func TestCheckProtectedZoneShellNonAsciiOutsideZoneStaysAllowed(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.WriteFile(filepath.Join(root, "개요.md"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	rows := []struct{ name, cmd string }{
		{"literal spelling", `rm 개요.md`},
		{"raw-byte spelling", `rm $'` + hzsHexWord("개요.md") + `'`},
	}
	swept := 0
	for _, row := range rows {
		swept++
		d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": row.cmd})
		if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
			t.Errorf("%s: decision=%q reason=%q, want allowed — the target is outside the zone", row.name, d, r)
		}
	}
	if swept != len(rows) {
		t.Fatalf("swept %d rows, want exactly %d", swept, len(rows))
	}
	t.Logf("swept=%d", swept)
}

// TestCheckProtectedZoneShellCodePointNulDoesNotTruncate — gate round 10 P1
// regression pin (card t1585): the part-level NUL termination is scoped to
// the origins REQ-HZS-001 names — hex 0x5C x 0 0 and the octal escapes, the
// two origins EVERY bash renders as a NUL byte. A code-point escape whose
// value is 0 must NOT terminate the part: its support is version-variant
// (this host's bash 3.2.57 renders the escape text literally — both-literal,
// decisively od-measured on this card), so a pre-4.2 shell acts on the FULL
// literal path while a decoder that truncates at any NUL byte judges a
// SHORTER word and allows. The reviewer's measured shape: a literally-named
// entry "docs" + backslash + u + 0 0 0 0 (text, created below as a
// directory) plus a command of the form rm $'docs<esc0>/../zone_dir/
// marker.md' — bash 3.2 resolves it through the literal-named entry into
// zone_dir and deletes the marker (gate-measured). The judgment keeps the
// pre-fix shape for code-point-origin NUL bytes — the NUL-bearing text
// Clean-collapses into the zone on the lexical arm and the row stays DENY.
// Green pre-M2, RED under the first M2 cut (any-NUL truncation judged
// "docs" and allowed), DENY again under the origin-scoped refinement —
// this row pins that scoping. In the Go source below the escape text is
// written with the doubled backslash (transport-safe source syntax for the
// single 0x5C byte at runtime).
func TestCheckProtectedZoneShellCodePointNulDoesNotTruncate(t *testing.T) {
	if runtime.GOOS == "windows" {
		// gate round 22 P2: the literally-named fixture entry docs+u0000
		// parses as TWO directory levels on windows, so the pre-4.2
		// reading's target becomes docs/zone_dir/marker.md — not the
		// protected file — and wantZoneDeny fails there for the wrong
		// reason. POSIX-specific row.
		t.Skip("POSIX-specific: the literally-named fixture entry carries backslashes, a separator on windows")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "zone_dir", "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The literally-named entry the pre-4.2 shell's argument walks through:
	// d o c s backslash u 0 0 0 0 — TEXT.
	if err := os.MkdirAll(filepath.Join(root, "docs\\u0000"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	const scopingCmd = "rm $'docs\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": scopingCmd})
	wantZoneDeny(t, "code-point nul origin scoping", d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("code-point nul origin scoping: deny observed but the protected marker is gone: %v", err)
	}
	t.Logf("swept=%d", 1)
}

// hzsMixedNulFixture builds the zone root plus the literally-named symlinks
// the pre-4.2 shell's argument walks through: the names are TEXT — backslash
// followed by u (or U) and hex digits — each pointing at the protected
// marker (the reviewer's measured landing: the write goes through the
// literal-named entry into the zone). POSIX-specific: the entry names carry
// backslashes, which are separators on windows — os.Symlink would fail
// there, so the rows skip (gate round 14 P2).
func hzsMixedNulFixture(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the literally-named fixture entries carry backslashes, a separator on windows")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "zone_dir", "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"link\\u0000", "link\\U00000000"} {
		if err := os.Symlink("zone_dir/marker.md", filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// hzsWantMixedDeny drives one mixed-origin NUL row and asserts the zone
// deny with the marker intact — the shared assertion body of the four
// gate-round-13 rows.
func hzsWantMixedDeny(t *testing.T, name, cmd string) {
	t.Helper()
	root := hzsMixedNulFixture(t)
	marker := filepath.Join(root, "zone_dir", "marker.md")
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
	wantZoneDeny(t, name, d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("%s: deny observed but the protected marker is gone: %v", name, err)
	}
	t.Logf("swept=%d", 1)
}

// The four gate-round-13 regression rows (card t1585): MIXED-origin NUL
// words. The part carries an EARLIER code-point-origin NUL (\u0000 or
// \U00000000) and a LATER hex/octal-origin NUL; the origin-scoped
// termination fires at the hex/octal NUL, so the part returns only the
// prefix before it and the judged candidate stops short of the zone —
// while the pre-4.2 shell acts on the FULL literal path through the
// literally-named symlink into the zone (reviewer confidence 1.00, real
// bash). Green at the audit baseline (the whole NUL-bearing text
// Clean-collapsed into the zone), RED under the origin-scoped refinement,
// DENY under the dual-candidate remedy: for a word carrying \u/\U escapes
// the guard judges BOTH the modern-decoded candidate AND the raw
// source-text candidate — the raw text Clean-collapses into the zone on
// the lexical arm and the t1566 raw arm resolves literal-named entries
// through symlinks. Escape texts are written with the doubled backslash
// (transport-safe source syntax for the single 0x5C byte at runtime).

// TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect — the reviewer's
// exact shape: \u0000 then \x00, a write REDIRECT through the literal-named
// symlink.
func TestCheckProtectedZoneShellMixedOriginNulDeniedRedirect(t *testing.T) {
	hzsWantMixedDeny(t, "mixed origin nul redirect",
		"printf changed > $'link\\u0000\\x00/../zone_dir/marker.md'")
}

// TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm — \u0000 then the
// OCTAL NUL, an rm argument candidate.
func TestCheckProtectedZoneShellMixedOriginNulDeniedOctalTerm(t *testing.T) {
	hzsWantMixedDeny(t, "mixed origin nul octal term",
		"rm -r $'link\\u0000\\0/../zone_dir'")
}

// TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex — \U00000000 then
// \x00, a write REDIRECT.
func TestCheckProtectedZoneShellMixedOriginNulDeniedUpperHex(t *testing.T) {
	hzsWantMixedDeny(t, "mixed origin nul upper hex",
		"printf changed > $'link\\U00000000\\x00/../zone_dir/marker.md'")
}

// TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal — \U00000000
// then the OCTAL NUL, an rm argument candidate.
func TestCheckProtectedZoneShellMixedOriginNulDeniedUpperOctal(t *testing.T) {
	hzsWantMixedDeny(t, "mixed origin nul upper octal",
		"rm -r $'link\\U00000000\\0/../zone_dir'")
}

// TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent — gate round
// 14 P1 (card t1585): the old-bash world must DECODE \xHH and octal escapes
// exactly like modern bash (raw bytes, the argument truncated at their NUL)
// and keep ONLY \u/\U as verbatim literal text — bash 3.2 has no \u/\U, so
// an unknown escape keeps its backslash and letter and the digits that
// follow are ordinary characters (measured: the ⊇ escape text passes
// as 5c 75 32 32 38 37). The whole-raw-text candidate judged the
// hex-escaped zone component (0x5C x 7a = "z" spelled \x7a) as an
// undecoded name, matched no entry, and ALLOWED — while the true pre-4.2
// path truncates at the \x00 NUL, `link\u0000`, and the symlink resolves
// INTO the zone (reviewer-measured: base deny, marker overwritten). The
// old-bash candidate lands the deny. Escape texts are written with the
// doubled backslash (transport-safe source syntax for the single 0x5C byte
// at runtime).
func TestCheckProtectedZoneShellMixedOriginNulDeniedHexComponent(t *testing.T) {
	root := hzsMixedNulFixture(t)
	marker := filepath.Join(root, "zone_dir", "marker.md")
	h := zoneTestHandler(t, root)
	const hexCompCmd = "printf changed > $'link\\u0000\\x00/../\\x7aone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": hexCompCmd})
	wantZoneDeny(t, "mixed origin nul hex component", d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("mixed origin nul hex component: deny observed but the protected marker is gone: %v", err)
	}
	t.Logf("swept=%d", 1)
}

// hzsLiteralNameFixture builds the zone root, the protected marker, a
// source.md for the cp shape, and the literally-named docs+backslash+u0000
// DIRECTORY the pre-4.2 shell's argument walks through. POSIX-specific: the
// entry name carries backslashes, a separator on windows (gate round 14
// P2).
func hzsLiteralNameFixture(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the literally-named fixture entry carries backslashes, a separator on windows")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsShellManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "zone_dir", "marker.md")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs\\u0000"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// hzsWantLiteralDeny drives one gate-round-15 row and asserts the zone deny
// with the marker intact.
func hzsWantLiteralDeny(t *testing.T, name, cmd string) {
	t.Helper()
	root := hzsLiteralNameFixture(t)
	marker := filepath.Join(root, "zone_dir", "marker.md")
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": cmd})
	wantZoneDeny(t, name, d, r, harnessLearnerIdentity, "category", "probe_zone")
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("%s: deny observed but the protected marker is gone: %v", name, err)
	}
	t.Logf("swept=%d", 1)
}

// The five gate-round-15 regression rows (card t1585, the reviewer's
// TestReviewDualWorldRegression shapes): every consumption site of word
// text must consume the CANDIDATE SET, not a single world. The dual worlds
// were wired into the path-candidate funnels only; these five shapes each
// reach a consumer still reading one world and ALLOW while the pre-4.2
// shell lands inside the zone through the literally-named docs+u0000
// entry. Green at the M2.3 base for the shapes the base judged whole-text,
// RED under the M2.3 tip, DENY under the consumer-set fix. Escape texts are
// written with the doubled backslash (transport-safe source syntax for the
// single 0x5C byte at runtime).

// TestCheckProtectedZoneShellDualWorldEmptyModernKept — an EMPTY modern
// candidate must not discard the word: the modern reading truncates to ""
// at the leading code-point NUL while the pre-4.2 reading still names the
// ../zone_dir climb.
func TestCheckProtectedZoneShellDualWorldEmptyModernKept(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world empty modern kept",
		"printf changed > $'\\u0000/../zone_dir/marker.md'")
}

// TestCheckProtectedZoneShellDualWorldCdReadings — the cd destination
// tracks BOTH readings: on 3.2 the cd lands in zone_dir through the
// literally-named entry, so the following rm judges from inside the zone.
func TestCheckProtectedZoneShellDualWorldCdReadings(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world cd readings",
		"cd $'docs\\u0000/../zone_dir'; rm marker.md")
}

// TestCheckProtectedZoneShellDualWorldVerbRecognition — the executable
// NAME tests every candidate: the modern reading truncates to "docs" (no
// mutation verb recognized) while the pre-4.2 path resolves to an rm
// executable through the literally-named entry.
func TestCheckProtectedZoneShellDualWorldVerbRecognition(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world verb recognition",
		"$'docs\\u0000/../rm' zone_dir/marker.md")
}

// TestCheckProtectedZoneShellDualWorldLongOptionValue — a long option's
// attached value is extracted from EVERY world: the modern value truncates
// to "docs" while the pre-4.2 value carries the ../zone_dir climb.
func TestCheckProtectedZoneShellDualWorldLongOptionValue(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world long option value",
		"cp source.md $'--target-directory=docs\\u0000/../zone_dir'")
}

// TestCheckProtectedZoneShellDualWorldGitAnchor — the git -C anchor tracks
// BOTH readings: on 3.2 the subcommand's file arguments resolve from
// inside the zone.
func TestCheckProtectedZoneShellDualWorldGitAnchor(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world git anchor",
		"git -C $'docs\\u0000/../zone_dir' rm marker.md")
}

// The five gate-round-17 regression rows (card t1585): each candidate
// funnel's EXISTING semantics — emptiness filters, verb/specialized
// recognition, anchor accumulation, overwrite-wins — must apply PER WORLD,
// not just to worlds[0].

// TestCheckProtectedZoneShellGitFunnelEmptyModernKept — the git file-arg
// funnel's empty-modern filter dropped the word's old-bash candidate (the
// same class zonePathCandidates had): git rm -f must judge the pre-4.2
// reading of the path.
func TestCheckProtectedZoneShellGitFunnelEmptyModernKept(t *testing.T) {
	hzsWantLiteralDeny(t, "git funnel empty modern kept",
		"git rm -f $'\\u0000/../zone_dir/marker.md'")
}

// TestCheckProtectedZoneShellDualWorldGitNameSpecialized — a world whose
// base name is GIT drives the git analysis: the modern reading truncates to
// "docs" (no verb, no git) while the pre-4.2 path executes git through the
// literally-named entry.
func TestCheckProtectedZoneShellDualWorldGitNameSpecialized(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world git name specialized",
		"$'docs\\u0000/../git' rm -f zone_dir/marker.md")
}

// TestCheckProtectedZoneShellDualWorldSedNameSpecialized — the sed variant:
// a world whose base name is SED drives the in-place analysis.
func TestCheckProtectedZoneShellDualWorldSedNameSpecialized(t *testing.T) {
	hzsWantLiteralDeny(t, "dual world sed name specialized",
		"$'docs\\u0000/../sed' -i zone_dir/marker.md")
}

// TestCheckProtectedZoneShellGitAnchorAccumulationBounded — eighteen -C
// options whose values carry two bash readings must accumulate PER WORLD
// (two chains, never crossed) and stay bounded. An ABSOLUTE -C replaces
// the anchor, so the per-world accumulation holds exactly two anchors —
// the modern reading truncates to the project root itself (the deny lands
// through it) while the pre-4.2 reading keeps the escape text as an
// ordinary component — where the cartesian product held 262,144
// candidates for 2 unique paths (reviewer-measured; the pre-fix crawl
// measured 13.53s in-suite). Escape texts are written with the doubled
// backslash (transport-safe source syntax for the single 0x5C byte at
// runtime).
func TestCheckProtectedZoneShellGitAnchorAccumulationBounded(t *testing.T) {
	root := hzsLiteralNameFixture(t)
	h := zoneTestHandler(t, root)
	anchor := "$'" + root + "/\\u0000x'"
	var b strings.Builder
	b.WriteString("git")
	for i := 0; i < 18; i++ {
		b.WriteString(" -C ")
		b.WriteString(anchor)
	}
	b.WriteString(" rm -f zone_dir/marker.md")
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": b.String()})
	wantZoneDeny(t, "git anchor accumulation bounded", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitWorkTreeOverwriteWins — git's LAST
// --work-tree wins: judging the earlier (already-overwritten) anchor is a
// FALSE DENY — an over-block, the inverse direction of the same defect
// class. The row asserts the ALLOW the real option semantics produce.
func TestCheckProtectedZoneShellGitWorkTreeOverwriteWins(t *testing.T) {
	root := hzsLiteralNameFixture(t)
	h := zoneTestHandler(t, root)
	const overBlockCmd = "git --git-dir=docs/.git --work-tree=zone_dir --work-tree=docs rm -f marker.md"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": overBlockCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git work-tree overwrite wins: decision=%q reason=%q, want allowed — the last --work-tree replaces the earlier anchor", overBlockCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// hzsMarkerFileManifest declares the single protected FILE the narrowed
// fixture protects — the cross-generation join lands on exactly this file
// and nothing else (gate round 20 P2: the wholesale zone_dir/ manifest let
// the modern reading trip on zone_dir/x for the wrong reason).
const hzsMarkerFileManifest = "  probe_zone:\n    paths: [\"zone_dir/marker.md\"]\n"

// hzsMarkerFileFixture builds the narrowed-manifest root with the protected
// marker. POSIX-specific skip: the rows sharing it create literally-named
// backslash entries.
func hzsMarkerFileFixture(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: the literally-named fixture entries carry backslashes, a separator on windows")
	}
	root := newZoneRoot(t, zoneShippedDoc(hzsMarkerFileManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "zone_dir", "marker.md"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// The three gate-round-19 regression rows (card t1585): branch dispatch
// must not preempt the dual-world verb classification, git file arguments
// must join their OWN generation's directory, and the candidate cap applies
// AFTER dedup.

// TestCheckProtectedZoneShellDualWorldVerbBeforeDispatch — the cd branch's
// early return skipped the dual-world verb classification: the modern
// reading truncates to "cd" (the walker takes the cd branch and returns)
// while the pre-4.2 path executes rm through the literally-named
// cd+u0000 entry. A word truncating to a DECLARED FUNCTION name is the
// same class — the verb classification completes BEFORE any dispatch.
func TestCheckProtectedZoneShellDualWorldVerbBeforeDispatch(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	if err := os.MkdirAll(filepath.Join(root, "cd\\u0000"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "$'cd\\u0000/../rm' zone_dir/marker.md"})
	wantZoneDeny(t, "dual world verb before dispatch", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitFileArgsOwnGeneration — git file arguments
// join their OWN generation's directory: the cross-generation join
// (modern dir × pre-4.2 file text) is a path NO generation executes, and
// landing it is a FALSE DENY — an over-block. The row asserts the ALLOW:
// generation 0 reads zone_dir/other.txt (not the protected marker — the
// narrowed manifest does not cover it), generation 1 reads
// zone\u0005fdir/marker.md against a literally-named entry that does not
// exist — only the crossed join lands on the protected marker.
func TestCheckProtectedZoneShellGitFileArgsOwnGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const crossCmd = "git -C $'zone\\u005fdir' rm -f $'other.txt\\u0000/../marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": crossCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git file args own generation: decision=%q reason=%q, want allowed — only the cross-generation join lands in the zone, and no generation executes it", crossCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellCandidateCapAfterDedup — dedup BEFORE the
// cap: a unicode-free command contributes the same path from BOTH worlds,
// so 2,050 files pool 4,100 candidates — past the 4,096 cap — while the
// deduped set is 2,050. Capping before dedup false-denies the plain
// command (an over-block). The row asserts the ALLOW.
func TestCheckProtectedZoneShellCandidateCapAfterDedup(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	var b strings.Builder
	b.WriteString("git rm -f")
	for i := 0; i < 2050; i++ {
		fmt.Fprintf(&b, " f%04d.txt", i)
	}
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": b.String()})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("candidate cap after dedup: decision=%q reason=%q, want allowed — the deduped candidate set (2,050) fits the cap", b.String(), r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellDualWorldGitNameBeforeCdDispatch — the
// git/sed SPECIALIZED dispatch had the same early-return hole the mutation
// classification had: the modern reading truncates to "cd" (the walker
// takes the cd branch and returns) while the pre-4.2 path executes git
// through the literally-named entry. ALL name-driven dispatches classify
// across both worlds before any single-world branch runs.
func TestCheckProtectedZoneShellDualWorldGitNameBeforeCdDispatch(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	if err := os.MkdirAll(filepath.Join(root, "cd\\u0000"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": "$'cd\\u0000/../git' rm -f zone_dir/marker.md"})
	wantZoneDeny(t, "dual world git name before cd dispatch", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellDeclaredFunctionShadowsVerb — a declared
// read-only function SHADOWS the external verb: bash runs ONLY the
// function and the protected file survives, so registering the arguments
// as rm targets is a FALSE DENY — an over-block, the inverse direction.
// The row asserts the ALLOW.
func TestCheckProtectedZoneShellDeclaredFunctionShadowsVerb(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const shadowCmd = "rm() { printf 'read-only\\n'; }; rm zone_dir/marker.md"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": shadowCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("declared function shadows verb: decision=%q reason=%q, want allowed — the function shadows the external rm and is read-only", shadowCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellCdReadingGenerationRelation — the cd's
// directory reading and the rm's file-argument reading keep their
// GENERATION RELATION: generation i's cwd joins generation i's file
// arguments only. The reviewer planted a symlink ONLY on the cross path
// (modern cwd zone_dir × pre-4.2 file reading link\u0005fx): the pooled
// join judges it and FALSE-DENIES, while neither true generation touches
// it — generation 0 reads zone_dir/link_x (not the marker, the narrowed
// manifest does not cover it) and generation 1 reads the literally-named
// zone\u0005fdir/link\u0005fx (an entry that does not exist). The row
// asserts the ALLOW. POSIX-specific skip: the planted entry carries
// backslashes, a separator on windows.
func TestCheckProtectedZoneShellCdReadingGenerationRelation(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	if err := os.Symlink(filepath.Join(root, "zone_dir", "marker.md"), filepath.Join(root, "zone_dir", "link\\u005fx")); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	const relationCmd = "cd $'zone\\u005fdir'; rm $'link\\u005fx'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": relationCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("cd reading generation relation: decision=%q reason=%q, want allowed — only the cross-generation cwd×file join reaches the planted symlink, and no generation executes it", relationCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// The four gate-round-23/25 regression rows (card t1585): the two bash
// worlds are FULLY ISOLATED interpretations — each world carries its own
// function registry, its own name→verb binding, and its own argument
// readings; the worlds meet only at the deny union.

// TestCheckProtectedZoneShellCandidateStateSeparated — gate round 23 P1:
// the g candidates' inner declarations must not overwrite each other on
// shared state — each candidate runs from the entry snapshot and the final
// registries MERGE, so the deleting f registration survives and the bare
// `f` call is judged as the mutation it is.
func TestCheckProtectedZoneShellCandidateStateSeparated(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const candCmd = "function g { f(){ :; }; }; g(){ f(){ rm zone_dir/marker.md; }; }; $'g'; f"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": candCmd})
	wantZoneDeny(t, "candidate state separated", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellVerbBindsOwnArgs — gate round 23 P2: the
// verb-name's world binds its own argument readings. Modern world: rm
// "docs" (harmless); pre-4.2 world: printf (harmless). Neither deletes —
// joining the modern name with the pre-4.2 argument is a FALSE DENY. The
// row asserts the ALLOW.
func TestCheckProtectedZoneShellVerbBindsOwnArgs(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const bindCmd = "$'rm\\u0000/../printf' $'docs\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": bindCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("verb binds own args: decision=%q reason=%q, want allowed — modern rm joins only the modern argument, and no world deletes", bindCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellRegistryPerGeneration — gate round 25 P1:
// the function registry is VERSION-SCOPED. The modern world's execution of
// f registers the no-op rm override IN THE MODERN WORLD only; the pre-4.2
// world never ran f (its name was the literal f\u0000/not_f path), so its
// rm stays external and deletes the protected marker — the guard must
// DENY. The row asserts the deny.
func TestCheckProtectedZoneShellRegistryPerGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const regCmd = "f() { rm() { :; }; }; $'f\\u0000/not_f'; rm zone_dir/marker.md"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": regCmd})
	wantZoneDeny(t, "registry per generation", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellVerbBindsOwnArgsNoop — gate round 25 P2, the
// name↔argument binding again: modern rm "docs" (harmless), pre-4.2 noop
// (harmless) — the guard testing the pre-4.2 protected path against the
// modern rm is a FALSE DENY. The row asserts the ALLOW.
func TestCheckProtectedZoneShellVerbBindsOwnArgsNoop(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const noopCmd = "$'rm\\u0000/../noop' $'docs\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": noopCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("verb binds own args noop: decision=%q reason=%q, want allowed — modern rm joins only the modern argument, and no world deletes", noopCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// The three gate-round-27 regression rows (card t1585): the isolation
// principle applied one funnel at a time — the numeric-FD decision, the
// git candidate generation, and the function-body redirections all bind
// to the world that executes them.

// TestCheckProtectedZoneShellNumericFdPerWorld — gate round 27 P1: the
// numeric-FD decision is PER WORLD. The modern reading of the >& word is
// "1" (a descriptor — writes nothing) while the pre-4.2 reading is a real
// path through the literally-named 1u0000 entry that old bash WRITES. The
// row asserts the DENY the pre-4.2 world's file write produces.
func TestCheckProtectedZoneShellNumericFdPerWorld(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const fdCmd = "printf changed >& $'1\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": fdCmd})
	wantZoneDeny(t, "numeric fd per world", d, r, harnessLearnerIdentity, "category", "probe_zone")
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitCandidatesConfinedToCaller — gate round
// 27 P2 (over-block): the git candidates confine to the CALLING
// generation. Modern reads `git rm -f docs` (harmless), pre-4.2 reads
// printf (harmless) — pooling both generations' file readings judged the
// protected path under the modern git → FALSE DENY. The row asserts the
// ALLOW.
func TestCheckProtectedZoneShellGitCandidatesConfinedToCaller(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const gitCmd = "$'git\\u0000/../printf' rm -f $'docs\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": gitCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git candidates confined to caller: decision=%q reason=%q, want allowed — modern git joins only the modern file reading, and no world deletes", gitCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellFunctionRedirectsOwnGeneration — gate round
// 27 P2 (over-block): the function executes only in the modern world
// (writing to docs, harmless) — judging the pre-4.2 redirect reading's
// protected path during that execution is a FALSE DENY. The row asserts
// the ALLOW.
func TestCheckProtectedZoneShellFunctionRedirectsOwnGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const fnCmd = "f(){ printf changed > $'docs\\u0000/../zone_dir/marker.md'; }; $'f\\u0000/../printf'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": fnCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("function redirects own generation: decision=%q reason=%q, want allowed — the function executes only in the modern world, whose redirect lands on docs", fnCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellSedInPlaceOwnGeneration — gate round 28 P2
// (over-block): the in-place option scan binds to the EXECUTING
// generation. The sed NAME is dual such that only the pre-4.2 world
// dispatches sed (the modern reading truncates to a non-sed name), and
// the option word decodes to -i ONLY in the modern world — the pre-4.2
// reading keeps the escape text literal (an invalid option: sed exits,
// touching nothing). The pooled modern scan read "-i" and false-denied
// while NEITHER world's sed runs in-place. The row asserts the ALLOW.
// Escape texts are written with the doubled backslash (transport-safe
// source syntax for the single 0x5C byte at runtime).
func TestCheckProtectedZoneShellSedInPlaceOwnGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	// the name word: the modern reading truncates to "no" (nothing
	// dispatches), the pre-4.2 reading resolves to sed through the
	// literally-named no+u0000p entry.
	const sedCmd = "$'no\\u0000p/../sed' $'\\u002di' 's/a/b/' zone_dir/marker.md"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": sedCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("sed in place own generation: decision=%q reason=%q, want allowed — the executing generation's option reading is the literal escape text, an invalid option: sed touches nothing", sedCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitSubOwnGeneration — gate round 29 P1
// (over-block): the git SUBCOMMAND word binds its own generation. Modern
// reads `git rm -f docs` (docs is outside the zone — harmless); pre-4.2
// reads the subcommand as the literal rm\u0000bogus text (a nonexistent
// subcommand — git exits, harmless). The pooled modern subcommand joined
// with the pre-4.2 file reading false-denied. The row asserts the ALLOW.
// Escape texts are written with the doubled backslash (transport-safe
// source syntax for the single 0x5C byte at runtime).
func TestCheckProtectedZoneShellGitSubOwnGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const subCmd = "git $'rm\\u0000bogus' -f $'docs\\u0000/../zone_dir/marker.md'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": subCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git sub own generation: decision=%q reason=%q, want allowed — modern git rm joins only the modern file reading (docs, outside the zone), and the pre-4.2 subcommand is nonexistent", subCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitNameSubOwnGeneration — gate round 30 P2
// (over-block): the executable word is dual (modern printf, harmless;
// pre-4.2 resolves to git) AND the subcommand word is dual (modern rm,
// mutating; pre-4.2 reads the literal rm\u0000 text — a nonexistent
// subcommand, harmless). The guard joined the pre-4.2 git execution with
// the modern rm subcommand → false deny. The row asserts the ALLOW.
func TestCheckProtectedZoneShellGitNameSubOwnGeneration(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	const nameSubCmd = "$'printf\\u0000/../git' $'rm\\u0000' -f zone_dir/marker.md"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": nameSubCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git name sub own generation: decision=%q reason=%q, want allowed — the pre-4.2 world's git subcommand is the nonexistent literal rm\\u0000 text, and the modern world's name is printf", nameSubCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellSedInPlacePlainFileShape — gate round 29 P2
// confirmation pin: the sed in-place generation binding (zoneSedInPlace
// takes the executing generation) covers the plain-file call shape —
// modern decodes the in-place option and edits a PLAIN file (harmless);
// pre-4.2 exits on the literal option text. The row asserts the ALLOW
// (green-now: the M2.11 binding already covers this shape).
func TestCheckProtectedZoneShellSedInPlacePlainFileShape(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	if err := os.WriteFile(filepath.Join(root, "docs_a"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	const plainCmd = "sed $'\\u002di' 's/a/b/' docs_a"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": plainCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("sed in place plain file shape: decision=%q reason=%q, want allowed — the modern world edits the plain file docs_a, outside the zone", plainCmd, r)
	}
	t.Logf("swept=%d", 1)
}

// TestCheckProtectedZoneShellGitMassFileArgsBounded — gate round 29 P3 /
// gate round 30 P3 / gate round 31 P1 (measurement pin): 80,000 file
// arguments must run BOUNDED — hash-set dedup (linear) and the
// unique-candidate cap firing fail-closed immediately, no per-candidate
// resolution of over-cap sets. The row asserts the deny (the marker file
// is among the arguments) and MEASURES the duration (logged; before/after
// recorded in progress.md §E.2 — not asserted, per the gate's directive).
func TestCheckProtectedZoneShellGitMassFileArgsBounded(t *testing.T) {
	root := hzsMarkerFileFixture(t)
	h := zoneTestHandler(t, root)
	var b strings.Builder
	b.WriteString("git rm -f zone_dir/marker.md")
	for i := 0; i < 80000; i++ {
		b.WriteString(" f")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(".txt")
	}
	start := time.Now()
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": b.String()})
	elapsed := time.Since(start)
	// the marker file is among the arguments: the per-generation join
	// denies; an over-cap set fails closed IMMEDIATELY (loop-unbounded) —
	// both are bounded outcomes; the row MEASURES the duration (logged;
	// before/after in progress.md §E.2 — not asserted, per the gate's
	// directive)
	if d != DecisionDeny || !strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("git mass file args bounded: decision=%q reason=%q, want a bounded deny", b.String(), r)
	}
	t.Logf("elapsed=%v category=%q (bounded-run measurement; before/after in progress.md §E.2)", elapsed, categoryOf(r))
	t.Logf("swept=%d", 1)
}

// categoryOf extracts the category token from a zone deny reason for
// bounded-outcome assertions.
func categoryOf(reason string) string {
	if i := strings.Index(reason, "category="); i >= 0 {
		rest := reason[i+len("category="):]
		if j := strings.Index(rest, " "); j >= 0 {
			return rest[:j]
		}
		return rest
	}
	return ""
}

// TestCheckProtectedZoneShellInvalidManifestRedirectAllow — gate round 31
// P2 (:635, over-block pin): with an INVALID manifest, a never-executed
// generation's redirect reading must not flip the mutating flag — the
// fail-closed invalid-manifest denial (REQ-SIPZ-009) requires a MUTATING
// command, and `f(){ printf read-only >& $'1'; }; $'f'` mutates nothing
// in either world (the fd duplication writes nothing). The row asserts
// the ALLOW the fail-closed gate must not swallow.
func TestCheckProtectedZoneShellInvalidManifestRedirectAllow(t *testing.T) {
	invalidManifest := "version: 1\ncategories:\n  probe_zone:\n    paths: []\n"
	root := newZoneRoot(t, invalidManifest, "")
	h := zoneTestHandler(t, root)
	const invCmd = "f(){ printf read-only >& $'1'; }; $'f'"
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": invCmd})
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Errorf("invalid manifest redirect allow: decision=%q reason=%q, want allowed — the fd-duplicating function mutates nothing in either world, so the fail-closed invalid-manifest denial must not fire", invCmd, r)
	}
	t.Logf("swept=%d", 1)
}
