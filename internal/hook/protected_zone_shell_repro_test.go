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
	"strings"
	"testing"

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
