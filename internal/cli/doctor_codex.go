package cli

// doctor_codex.go — SPEC-CODEX-WIRING-001 M4, the `moai doctor` "Codex
// Wiring" diagnostic (REQ-CW-010 / AC-CW-012).
//
// Advisory and fail-open (checkBinaryFreshness t184 precedent): the check
// reports, it never gates — an inactive project on a codex-less machine is an
// informational skip, and every finding is an action directive rather than a
// claim about Codex's undocumented internal trust store (spec §H: the
// sidecar divergence is an INDIRECT signal; the advice says what to DO).
//
// The check READS only. It never creates, repairs, or removes a .codex/ file,
// and never touches the user-layer ~/.codex/config.toml: wiring creation is
// the explicit `moai init --agent codex` opt-in (REQ-CW-009) and a user-owned
// table is the doctor's to report, never the writer's to repair (REQ-CW-005).

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// codexWiringLookPath is the PATH-resolution seam (stubbed in tests so the
// check's verdict never depends on the machine running the test suite).
var codexWiringLookPath = exec.LookPath

// The user-layer config is located through resolveCodexHomeDir (mcp_codex.go)
// rather than a home seam of this file's own: CODEX_HOME is a first-class
// convention here (REQ-CL-005), and a second resolution that ignored it read
// a file Codex does not, while leaving the file Codex DOES read unchecked.
// That resolver already carries the codexUserHomeDir seam tests pin.

// reTrustAdvice is the divergence action directive (AC-CW-012 token).
const reTrustAdvice = "run codex /hooks to re-trust the changed hooks"

// initCodexAdvice is the unwired-project action directive. `moai init` is the
// ONLY path that creates wiring — RefreshWiring is an existence gate that
// creates nothing (REQ-CW-009), so an unwired project stays unwired until the
// user opts in explicitly.
const initCodexAdvice = "run moai init --agent codex"

// codexHomeConfigDisplay / codexHomeConfigEnvDisplay are how the user-layer
// config is NAMED in a finding summary. The symbolic form keeps the summary
// inside the width band; the exact resolved path rides in Detail, where the
// evidence belongs and the width is the reader's own opt-in.
const (
	codexHomeConfigDisplay    = "~/.codex/config.toml"
	codexHomeConfigEnvDisplay = "$CODEX_HOME/config.toml"
)

// codexMessageWidthCeiling bounds a Message in RUNES. The doctor panel sizes
// itself to its widest row, so one long message widens the whole box past the
// terminal and breaks every other row's alignment. 113 is the longest message
// the committed golden fixtures already carry (internal/cli/testdata/
// doctor-nocolor.golden, the Constitution Registry row) — this check stays
// inside the band the existing rows occupy rather than defining a new one.
const codexMessageWidthCeiling = 113

// codexFinding is one problem in two registers: a SHORT summary for Message,
// which a plain `moai doctor` renders, and the full text for Detail, which
// renders only under --verbose. Splitting them is what lets an action
// directive stay visible without the enumeration behind it blowing the panel.
type codexFinding struct {
	summary string
	detail  string
}

// checkCodexWiring verifies the Codex wiring of the project at root:
// hooks.json presence + whitelist validity, sidecar-hash divergence, the
// moai binary's PATH resolution, and the config.toml [mcp_servers.moai]
// table's presence/canonical shape — plus, when Codex is in play at all, the
// user-layer skill registrations. A project without wiring files on a machine
// without codex is an informational skip (the opt-in marker, REQ-CW-009).
func checkCodexWiring(root string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Codex Wiring"}

	hooksRaw, hooksErr := os.ReadFile(filepath.Join(root, codexwiring.HooksRelPath))
	cfgRaw, cfgErr := os.ReadFile(filepath.Join(root, codexwiring.ConfigRelPath))

	wired := hooksErr == nil || cfgErr == nil
	_, codexPathErr := codexWiringLookPath("codex")
	codexInstalled := codexPathErr == nil

	if !wired && !codexInstalled {
		// Claude-only project on a claude-only machine: nothing about Codex
		// is in play, so the check stays silent. This is the un-nagging
		// invariant — it must survive every addition below.
		check.Status = uikit.CheckOK
		check.Message = "not wired (claude-only project) — skipped"
		return check
	}

	var problems []codexFinding
	// extraDetail carries observations that are not findings in their own
	// right (an unreadable sidecar, a verbose-only note).
	var extraDetail []string

	if !wired {
		// Codex IS installed but this project was never wired. Silence here
		// costs the user the whole integration with no signal: no
		// project-layer MCP registration and every generated hook dead. The
		// directive rides IN the summary because Detail renders only under
		// --verbose.
		//
		// The claim is scoped to what was READ. Only the two PROJECT files
		// were inspected, so this says nothing about whether the user-layer
		// config registers the MoAI server — asserting a machine-wide "not
		// registered" from a project-local absence is an unobserved premise.
		problems = append(problems, codexFinding{
			summary: "codex installed, project not wired — " + initCodexAdvice,
			detail: fmt.Sprintf(
				"codex resolves on PATH but this project declares no Codex wiring (%s and %s are both absent), so this project registers no MoAI MCP server and the generated hooks cannot fire here; %s",
				codexwiring.HooksRelPath, codexwiring.ConfigRelPath, initCodexAdvice),
		})
	} else {
		// hooks.json: presence, whitelist validity, sidecar divergence.
		if hooksErr != nil {
			problems = append(problems, plainCodexFinding(fmt.Sprintf("%s missing (wiring active but the hook layer is gone)", codexwiring.HooksRelPath)))
		} else {
			if violations, verr := codexadapter.ValidateConfig(hooksRaw); verr != nil {
				problems = append(problems, plainCodexFinding(fmt.Sprintf("hooks.json unparseable: %v", verr)))
			} else if len(violations) > 0 {
				// Codex silently disables the whole file on one stray key (t83
				// Finding D) — this is the observability backstop.
				names := make([]string, len(violations))
				for i, v := range violations {
					names[i] = v.Error()
				}
				problems = append(problems, codexFinding{
					summary: "hooks.json fails the key whitelist — Codex would silently ignore the file",
					detail:  "hooks.json fails the key whitelist (" + strings.Join(names, "; ") + ") — Codex would silently ignore the file",
				})
			}

			if sidecar, present, serr := codexwiring.LoadSidecar(root); serr != nil {
				extraDetail = append(extraDetail, fmt.Sprintf("sidecar unreadable: %v", serr))
			} else if present {
				sum := sha256.Sum256(hooksRaw)
				if sidecar.HooksSHA256 != hex.EncodeToString(sum[:]) {
					// The directive rides IN the message (not Detail) so a plain
					// `moai doctor` surfaces it — Detail only renders under
					// --verbose (AC-CW-012 clause 2's plain-doctor reading).
					problems = append(problems, codexFinding{
						summary: "hooks.json changed since generation — " + reTrustAdvice,
						detail:  "hooks.json differs from the last generated content (sidecar hash mismatch) — " + reTrustAdvice,
					})
				}
			} else if verbose {
				extraDetail = append(extraDetail, "no trust sidecar recorded (never wired by this generator)")
			}
		}

		// moai binary PATH resolution: the generated hook commands are
		// `moai hook ...` — an unresolvable binary means none of them can fire.
		if _, lerr := codexWiringLookPath("moai"); lerr != nil {
			problems = append(problems, plainCodexFinding("moai binary not found on PATH — the generated hook commands cannot fire"))
		}

		// config.toml: table presence + canonical shape (drift is REPORTED; the
		// writer never repairs a user-owned table, REQ-CW-005).
		if cfgErr != nil {
			problems = append(problems, plainCodexFinding(fmt.Sprintf("%s missing (wiring active but the MCP registration is gone)", codexwiring.ConfigRelPath)))
		} else {
			status := codexwiring.InspectMCPTable(cfgRaw)
			switch {
			case !status.Present:
				problems = append(problems, plainCodexFinding("[mcp_servers.moai] table missing from config.toml"))
			case !status.Canonical:
				problems = append(problems, plainCodexFinding("[mcp_servers.moai] table differs from the canonical registration (user-owned; left untouched)"))
			}
		}
	}

	// User-layer skill registrations. Reached only when Codex is in play at
	// all (wired project, or codex on PATH) — the guard above already
	// returned for the claude-only case.
	if finding, ok := codexStaleSkillFinding(); ok {
		problems = append(problems, finding)
	}

	if len(problems) == 0 {
		check.Status = uikit.CheckOK
		check.Message = "wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical)"
		check.Detail = joinCodexDetails(nil, extraDetail)
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = joinCodexSummaries(problems)
	check.Detail = joinCodexDetails(problems, extraDetail)
	if check.Detail == "" && verbose {
		check.Detail = "advisory check — rerun `moai init --agent codex` to refresh the wiring"
	}
	return check
}

// plainCodexFinding is a finding whose full text is already short enough to
// ride in Message unchanged.
func plainCodexFinding(text string) codexFinding {
	return codexFinding{summary: text, detail: text}
}

// joinCodexSummaries renders the Message, bounded by codexMessageWidthCeiling.
//
// Summaries are kept in order and dropped from the TAIL when they do not fit,
// so a directive-bearing finding — which is always raised before the
// user-layer sweep — survives truncation and stays visible on a plain
// `moai doctor`. The dropped count is reported rather than hidden, and the
// full text of every finding remains in Detail.
//
// One exception is deliberate: a single leading summary that alone exceeds
// the ceiling is emitted whole. Cutting it mid-sentence would truncate its
// own directive, which is the outcome the bound exists to prevent.
func joinCodexSummaries(findings []codexFinding) string {
	if len(findings) == 0 {
		return ""
	}
	summaries := make([]string, len(findings))
	for i, f := range findings {
		summaries[i] = f.summary
	}
	if utf8.RuneCountInString(strings.Join(summaries, "; ")) <= codexMessageWidthCeiling {
		return strings.Join(summaries, "; ")
	}
	// Upper bound on the marker, computed from the largest count it can name.
	markerBudget := utf8.RuneCountInString(codexOverflowMarker(len(summaries)))
	kept := 1
	for i := 2; i <= len(summaries); i++ {
		if utf8.RuneCountInString(strings.Join(summaries[:i], "; "))+markerBudget > codexMessageWidthCeiling {
			break
		}
		kept = i
	}
	return strings.Join(summaries[:kept], "; ") + codexOverflowMarker(len(summaries)-kept)
}

// codexOverflowMarker names how many findings Message dropped for width.
func codexOverflowMarker(dropped int) string {
	return fmt.Sprintf(" (+%d more, see --verbose)", dropped)
}

// joinCodexDetails renders Detail as one line per finding. Detail renders
// only under --verbose, and the panel sizes itself to its widest row, so the
// per-finding text is kept on separate lines rather than concatenated into
// one row that would widen the box all over again.
func joinCodexDetails(findings []codexFinding, extra []string) string {
	var texts []string
	for _, f := range findings {
		if f.detail != "" {
			texts = append(texts, f.detail)
		}
	}
	texts = append(texts, extra...)
	if len(texts) == 0 {
		return ""
	}
	var lines []string
	for _, t := range texts {
		lines = append(lines, wrapCodexDetail(t, codexMessageWidthCeiling)...)
	}
	return strings.Join(lines, "\n      ")
}

// wrapCodexDetail folds one detail text onto lines of at most width runes.
// Detail becomes a row of its own under --verbose, so an unwrapped detail
// widens the panel exactly as an unwrapped Message does — the width bound has
// to hold on both surfaces or it holds on neither.
//
// A single token longer than width is emitted whole rather than split: the
// tokens that get that long here are filesystem paths, and a path broken
// across lines stops being copy-pasteable, which is the whole point of citing
// it.
func wrapCodexDetail(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) > width {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur += " " + w
	}
	return append(lines, cur)
}

// codexSkillPathShape classifies a declared [[skills.config]] path by shape,
// BEFORE any filesystem check (SPEC-CODEX-SKILL-PATH-001). Only shapes that
// resolve deterministically in this environment (absolute, and home-relative
// after expansion) can feed the missing verdict; every other shape is a
// declared classification that is reported and never acted on.
type codexSkillPathShape int

const (
	// codexPathAbsolute is filepath.IsAbs. IsAbs is evaluated FIRST, so a
	// Windows host classifies its own native C:\... paths here — never as
	// oddly-formed — while a backslash-bearing fragment on any host does
	// not.
	codexPathAbsolute codexSkillPathShape = iota
	// codexPathHomeRelative is exactly "~" or a "~/"-prefixed path — the
	// user's own home, expandable here.
	codexPathHomeRelative
	// codexPathRelative declares a path with no observed resolution base in
	// this repository. Guessing one would manufacture a verdict, so the
	// entry is reported as relative and never stat'ed.
	codexPathRelative
	// codexPathOddlyFormed is a backslash-bearing non-absolute fragment (a
	// Windows-shaped separator in a non-absolute position, or a residual
	// un-decoded TOML escape) or another user's "~user" home form.
	codexPathOddlyFormed
)

// classifyCodexSkillPath maps a declared path onto its shape. Ordering is
// load-bearing: IsAbs runs before the backslash check (a native Windows
// absolute stays absolute), and the home-relative forms are peeled off
// before the residual "~"-prefixed shapes fall to oddly-formed.
func classifyCodexSkillPath(p string) codexSkillPathShape {
	if filepath.IsAbs(p) {
		return codexPathAbsolute
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		return codexPathHomeRelative
	}
	if strings.HasPrefix(p, "~") || strings.ContainsRune(p, '\\') {
		return codexPathOddlyFormed
	}
	return codexPathRelative
}

// expandCodexHomeRelativePath expands a "~" or "~/"-prefixed declaration
// against the USER home resolved through the existing codexUserHomeDir seam
// (os.UserHomeDir) — the same seam resolveCodexHomeDir uses, never a second
// resolver, and never CODEX_HOME: the env var locates the config, the user
// home expands "~". ok=false means the home itself is unresolvable, which
// the caller must treat as indeterminate, never absent.
func expandCodexHomeRelativePath(p string) (expanded string, ok bool) {
	home, err := codexUserHomeDir()
	if err != nil || home == "" {
		return "", false
	}
	if p == "~" {
		return home, true
	}
	return filepath.Join(home, p[2:]), true
}

// codexStaleSkillFinding reports the user-layer [[skills.config]] entries
// whose declared path no longer exists — registrations Codex neither prunes
// nor complains about, so nothing else surfaces them.
//
// The config is located through resolveCodexHomeDir, so CODEX_HOME decides
// which file is read, exactly as it does everywhere else in this binary. A
// resolution that ignored it would warn about a file Codex does not read
// while leaving the file Codex DOES read unexamined — wrong in both
// directions at once.
//
// Fail-open in every direction: an unresolvable home, an absent or unreadable
// config, or a config declaring no entries all yield ok=false (a silent
// skip), so a missing input never becomes a finding. The function READS only.
func codexStaleSkillFinding() (codexFinding, bool) {
	codexHome, source := resolveCodexHomeDir()
	if codexHome == "" {
		return codexFinding{}, false
	}
	cfgPath := filepath.Join(codexHome, path.Base(codexwiring.ConfigRelPath))
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return codexFinding{}, false
	}
	entries := codexwiring.ParseSkillEntries(raw)
	if len(entries) == 0 {
		return codexFinding{}, false
	}

	var missingEnabled, missingDisabled, missingUnspecified, indeterminate int
	var relativeCount, oddlyFormed int
	for _, e := range entries {
		if e.Path == "" {
			// An entry declaring no path says nothing about a file's
			// existence — counted in the total, never as a missing path.
			continue
		}
		var statPath string
		switch classifyCodexSkillPath(e.Path) {
		case codexPathAbsolute:
			statPath = e.Path
		case codexPathHomeRelative:
			expanded, ok := expandCodexHomeRelativePath(e.Path)
			if !ok {
				// The user home is unresolvable, so the path's existence is
				// indeterminate — never absent (fail-open, not a guess).
				indeterminate++
				continue
			}
			statPath = expanded
		case codexPathRelative:
			// The base a relative path resolves against is not observed in
			// this repository; stat'ing it against process cwd would decide
			// the verdict on a base nobody chose. Reported as its own
			// classification, never counted missing.
			relativeCount++
			continue
		default:
			// A backslash-bearing non-absolute fragment or a "~user" form:
			// a declared shape this check cannot resolve here. Reported as
			// its own classification, never counted missing.
			oddlyFormed++
			continue
		}
		_, serr := os.Stat(statPath)
		switch {
		case serr == nil:
			// The path resolves. A DIRECTORY resolves too, and is likewise
			// not a missing path: how Codex treats a directory here is not
			// observed, and reporting one as missing would advise deleting a
			// registration on a guess.
		case errors.Is(serr, fs.ErrNotExist):
			switch e.Enabled {
			case codexwiring.SkillEnabledTrue:
				missingEnabled++
			case codexwiring.SkillEnabledFalse:
				missingDisabled++
			default:
				missingUnspecified++
			}
		default:
			// Permission denied, a symlink loop, an I/O error: the path's
			// existence is INDETERMINATE, not absent. The finding tells the
			// user to REMOVE the entry, so an unobserved absence must never
			// reach the missing count — that would advise deleting a healthy
			// registration on the strength of a stat this process was not
			// allowed to complete. Surfaced in Detail, never acted on.
			indeterminate++
		}
	}
	missing := missingEnabled + missingDisabled + missingUnspecified
	unresolvedShape := relativeCount + oddlyFormed
	if missing == 0 && unresolvedShape == 0 {
		// No missing path and no unresolvable shape: nothing to say. An
		// indeterminate-only config stays silent too — the t451 posture,
		// preserved verbatim.
		return codexFinding{}, false
	}

	display := codexHomeConfigDisplay
	if source == codexHomeSourceEnv {
		display = codexHomeConfigEnvDisplay
	}

	// The summary carries the count and the file; the denominator, the split
	// and the directive ride in Detail. The split is the severity axis, so it
	// is quantified rather than summed away — but it is reported as DECLARED:
	// an entry with no `enabled` key is counted as unspecified rather than
	// folded into either side, because Codex's default for an absent key is
	// not observed anywhere in this repository. The remove directive is
	// bound to the missing segment alone: relative and oddly-formed entries
	// are reported per classification and never advised away.
	detail := fmt.Sprintf(
		"%s declares %d [[skills.config]] %s",
		cfgPath, len(entries), pluralCodexEntries(len(entries)))
	if missing > 0 {
		detail += fmt.Sprintf(
			"; %d with a path that no longer exists (%d enabled, %d disabled, %d unspecified) — remove the stale entries or restore the skill files",
			missing, missingEnabled, missingDisabled, missingUnspecified)
	}
	if relativeCount > 0 {
		detail += fmt.Sprintf(
			"; %d relative %s (not checked: the resolution base is not observed)",
			relativeCount, pluralCodexEntries(relativeCount))
	}
	if oddlyFormed > 0 {
		detail += fmt.Sprintf(
			"; %d oddly-formed %s (not checked: backslash or ~other-user shape)",
			oddlyFormed, pluralCodexEntries(oddlyFormed))
	}
	if indeterminate > 0 {
		detail += fmt.Sprintf("; a further %d could not be checked and are NOT counted as missing", indeterminate)
	}

	summary := fmt.Sprintf("%s: %d stale skill %s", display, missing, pluralCodexEntries(missing))
	if missing == 0 {
		// Non-destructive summary for a config whose only finding is an
		// unresolvable path SHAPE — no "stale", nothing to remove.
		summary = fmt.Sprintf("%s: %d skill %s with a path shape this check cannot resolve",
			display, unresolvedShape, pluralCodexEntries(unresolvedShape))
	}

	return codexFinding{
		summary: summary,
		detail:  detail,
	}, true
}

// pluralCodexEntries picks the noun for an entry count.
func pluralCodexEntries(n int) string {
	if n == 1 {
		return "entry"
	}
	return "entries"
}
