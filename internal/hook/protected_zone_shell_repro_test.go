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
//   $'⊇'           -> e2 8a 87           (\u renders the code point as
//                                               UTF-8 — the current decoder
//                                               is right for \u/\U only)
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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
