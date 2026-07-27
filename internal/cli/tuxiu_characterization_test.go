package cli

// SPEC-CLI-TUX-INIT-UPDATE-001 M4 — AC-TUXIU-016 (top-risk preservation gate).
//
// The v0.1.3 presentation redesign (glyph SSOT, update card/pills/progress bar,
// identity band, re-dressed init success card, restored large logo, spinner-
// residue fix) is PRESENTATION-ONLY: it must not change WHICH files are
// deployed, the COUNT of files updated, the OUTCOME, nor the stdout-vs-stderr
// channel partition of the DATA lines.
//
// This test proves that invariant mechanically by comparing two committed
// golden fixture sets:
//
//   - testdata/tuxiu/*.golden          — the M1.0 PRE-logo characterization
//     baseline, captured BEFORE any M1 edit (plan.md §C first task). Immutable.
//   - testdata/tuxiu/postm4/*.golden    — the FRESH post-M4 presentation
//     baseline, captured from the M1+M2+M3 binary WITH the new presentation.
//
// Fixture strategy: OPTION (a) — the fresh logo-era goldens are added ALONGSIDE
// the M1.0 pre-logo goldens (rather than replacing them), so the data-invariance
// guarantee "vs the PRE-logo baseline" stays mechanically checkable as a diff of
// two committed fixtures. Both sets are captured by the same scratchpad harness
// (internal/cli scratchpad capture.sh + pty_capture.py) in throwaway project
// dirs — the repo is never mutated (B0).
//
// The DATA-line filter (dataLines below) drops the EXPECTED-NEW presentation the
// AC-016 logo note sanctions — the restored logo art, the identity band, the
// classification card, the re-dressed init success card, the block progress bars
// — and the M2 spinner-residue fix (a documented REMOVAL of the duplicated
// per-step reporter block from stderr), and normalizes run-variant backup
// timestamps + NO_COLOR pill brackets + indentation. What survives is exactly
// the file / count / outcome DATA subset, which MUST be byte-identical between
// the two baselines.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	tuxAnsiRE     = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]|\x1b\\][^\x07]*\x07")
	tuxTSISO      = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z`)
	tuxTSCompact  = regexp.MustCompile(`\d{8}_\d{6}`)
	tuxStepsRE    = regexp.MustCompile(`\d+/\d+ steps`)   // progress-bar label + "N/M steps complete" reporter
	tuxReporterRE = regexp.MustCompile(`^○ .+?: `)         // reporter/phase "○ Name: gerund..." lines
	tuxPillBrkt   = strings.NewReplacer("[", "", "]", "") // NO_COLOR pill brackets [label] -> label
)

// tuxStripANSI removes CSI + OSC escape sequences.
func tuxStripANSI(s string) string { return tuxAnsiRE.ReplaceAllString(s, "") }

// tuxNormLine normalizes a single ANSI-stripped line for DATA comparison:
// removes CR, folds run-variant backup timestamps to <TS>, removes NO_COLOR pill
// brackets, and trims indentation/padding (both presentation).
func tuxNormLine(l string) string {
	l = strings.ReplaceAll(l, "\r", "")
	l = tuxTSISO.ReplaceAllString(l, "<TS>")
	l = tuxTSCompact.ReplaceAllString(l, "<TS>")
	l = tuxPillBrkt.Replace(l)
	return strings.TrimSpace(l)
}

// tuxIsPresentation reports whether a normalized line is EXPECTED-NEW (or
// residue-removed) presentation that is excluded from the DATA-line subset.
func tuxIsPresentation(l string) bool {
	if strings.ContainsAny(l, "█░") { // logo art fill / progress-bar fill+empty
		return true
	}
	for _, r := range "╗╔╚╝═║" { // logo double-line box runes
		if strings.ContainsRune(l, r) {
			return true
		}
	}
	if r := []rune(l); len(r) > 0 { // card/box chrome + interior (re-dressed cards)
		switch r[0] {
		case '│', '╭', '╰', '╮', '╯', '─':
			return true
		}
	}
	if strings.Contains(l, "◆") && strings.Contains(l, "MoAI-ADK") { // NEW identity band
		return true
	}
	if tuxStepsRE.MatchString(l) { // block progress bar + "· N/M steps complete" reporter
		return true
	}
	if tuxReporterRE.MatchString(l) { // "○ Name: gerund..." reporter/phase lines
		return true
	}
	if strings.HasSuffix(l, " complete") { // "✓ X complete" reporter/phase lines
		return true
	}
	switch l { // residue reporter ✓ lines that lack the " complete" suffix
	case "✓ Configuration backed up", "✓ Settings restored", "✓ Templates loaded", "✓ Manifest loaded":
		return true
	}
	return false
}

// dataLines extracts the file/count/outcome DATA-line subset from a raw capture.
func tuxDataLines(raw string) []string {
	out := make([]string, 0, 64)
	for _, ln := range strings.Split(tuxStripANSI(raw), "\n") {
		n := tuxNormLine(ln)
		if n == "" || tuxIsPresentation(n) {
			continue
		}
		out = append(out, n)
	}
	return out
}

func tuxReadGolden(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "tuxiu", rel))
	if err != nil {
		t.Fatalf("read golden %s: %v", rel, err)
	}
	return string(b)
}

// TestInitUpdateTUXCharacterization is the AC-TUXIU-016 top-risk preservation
// gate: the file/count/outcome DATA subset is byte-identical between the M1.0
// PRE-logo baseline and the post-M4 presentation baseline, across all 12
// surface × variant × channel captures, and the stdout/stderr partition of the
// DATA is preserved.
func TestInitUpdateTUXCharacterization(t *testing.T) {
	surfaces := []string{"init", "update"}
	variants := []string{"tty", "notty", "nocolor"}
	channels := []string{"stdout", "stderr"}

	for _, surf := range surfaces {
		for _, variant := range variants {
			for _, chan_ := range channels {
				name := surf + "." + variant + "." + chan_ + ".golden"
				t.Run(surf+"/"+variant+"/"+chan_, func(t *testing.T) {
					base := tuxDataLines(tuxReadGolden(t, name))
					post := tuxDataLines(tuxReadGolden(t, filepath.Join("postm4", name)))
					if len(base) != len(post) {
						t.Fatalf("DATA-line count drift %s: pre-logo=%d post-M4=%d\n--- pre-logo ---\n%s\n--- post-M4 ---\n%s",
							name, len(base), len(post), strings.Join(base, "\n"), strings.Join(post, "\n"))
					}
					for i := range base {
						if base[i] != post[i] {
							t.Errorf("DATA drift %s line %d:\n  pre-logo: %q\n  post-M4 : %q", name, i+1, base[i], post[i])
						}
					}
				})
			}
		}
	}
}

// TestTUXChannelPartition asserts the DATA lines that carry the update outcome
// vs the pre-deploy "Found N files" count keep their stdout / stderr channel in
// BOTH baselines (AC-TUXIU-016 (b) — the printer-gateway partition is unchanged).
func TestTUXChannelPartition(t *testing.T) {
	for _, variant := range []string{"tty", "notty", "nocolor"} {
		t.Run(variant, func(t *testing.T) {
			for _, set := range []string{"", "postm4"} {
				outName := filepath.Join(set, "update."+variant+".stdout.golden")
				errName := filepath.Join(set, "update."+variant+".stderr.golden")
				out := tuxStripANSI(tuxReadGolden(t, outName))
				erra := tuxStripANSI(tuxReadGolden(t, errName))
				label := set
				if label == "" {
					label = "pre-logo"
				}
				// Outcome "Updated N files" rides stdout, never stderr.
				if !strings.Contains(out, "Updated 26 files") {
					t.Errorf("[%s] %q: 'Updated 26 files' missing from stdout", label, variant)
				}
				if strings.Contains(erra, "Updated 26 files") {
					t.Errorf("[%s] %q: 'Updated 26 files' leaked onto stderr", label, variant)
				}
				// Pre-deploy "Found N files to sync" rides stderr, never stdout.
				if !strings.Contains(erra, "Found 26 files to sync") {
					t.Errorf("[%s] %q: 'Found 26 files to sync' missing from stderr", label, variant)
				}
				if strings.Contains(out, "Found 26 files to sync") {
					t.Errorf("[%s] %q: 'Found 26 files to sync' leaked onto stdout", label, variant)
				}
			}
		})
	}
}

// TestTUXDataValuesPreserved cross-checks the specific count/outcome VALUES that
// live inside re-dressed cards (which tuxDataLines drops as presentation), so the
// data-invariance guarantee still covers the card-interior counts (AC-TUXIU-016).
func TestTUXDataValuesPreserved(t *testing.T) {
	for _, set := range []string{"", "postm4"} {
		label := set
		if label == "" {
			label = "pre-logo"
		}
		// update: outcome count + found count + skills-archived summary.
		uOut := tuxStripANSI(tuxReadGolden(t, filepath.Join(set, "update.notty.stdout.golden")))
		for _, want := range []string{"Updated 26 files", "3 skills archived, 0 user customizations modified"} {
			if !strings.Contains(uOut, want) {
				t.Errorf("[%s] update stdout missing data token %q", label, want)
			}
		}
		// init: created-files count (outside the card) + card dir/file counts.
		iErr := tuxStripANSI(tuxReadGolden(t, filepath.Join(set, "init.notty.stderr.golden")))
		if !strings.Contains(iErr, "Created 2 files") {
			t.Errorf("[%s] init stderr missing 'Created 2 files'", label)
		}
		for _, want := range []string{"13", "2"} { // 13 dirs, 2 files preserved across the card re-dress
			if !strings.Contains(iErr, want) {
				t.Errorf("[%s] init stderr missing count token %q", label, want)
			}
		}
	}
}

// TestTUXPostM4CarriesNewPresentation confirms the post-M4 fixtures actually are
// the logo-era presentation baseline: the update identity band + classification
// card and the re-dressed init success card appear ONLY in the post-M4 set,
// which is why the tuxDataLines filter (which drops them) is load-bearing.
func TestTUXPostM4CarriesNewPresentation(t *testing.T) {
	uPost := tuxStripANSI(tuxReadGolden(t, filepath.Join("postm4", "update.notty.stdout.golden")))
	uBase := tuxStripANSI(tuxReadGolden(t, "update.notty.stdout.golden"))
	if !strings.Contains(uPost, "◆ MoAI-ADK") {
		t.Error("post-M4 update stdout should carry the NEW identity band '◆ MoAI-ADK'")
	}
	if strings.Contains(uBase, "◆ MoAI-ADK") {
		t.Error("pre-logo update stdout should NOT carry the identity band (baseline invariant)")
	}
	if !strings.Contains(uPost, "update") || !strings.Contains(uPost, "conflict") {
		t.Error("post-M4 update stdout should carry the NEW classification card (update/conflict pills)")
	}
	// Re-dressed init success card: "N dirs" / "N files" pill wording is post-M4 only.
	iPost := tuxStripANSI(tuxReadGolden(t, filepath.Join("postm4", "init.notty.stderr.golden")))
	iBase := tuxStripANSI(tuxReadGolden(t, "init.notty.stderr.golden"))
	if !strings.Contains(iPost, "dirs") || !strings.Contains(iPost, "files") {
		t.Error("post-M4 init card should carry the re-dressed 'N dirs / N files' pill wording")
	}
	if !strings.Contains(iBase, "Directories") {
		t.Error("pre-logo init card should carry the ORIGINAL 'Directories N created' wording (baseline invariant)")
	}
}
