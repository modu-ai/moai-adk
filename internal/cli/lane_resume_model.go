// lane_resume_model.go — the claude option model behind the resume
// interpretation (SPEC-SESSION-CC-VERSION-002 REQ-SCV-012/013): one measured
// knowledge surface — option token → value class — derived at launcher entry
// from the resolved claude binary's own `--help` synopsis, with the
// compile-time snapshot as the silent fallback. The two resume interpreters
// read the active model as pure package state (C.2 of the SPEC): the
// derivation happens behind the seam at the launcher-entry call sites, never
// inside the walk, and no test of the derivation spawns a process.

package cli

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// claudeOptionClass is the three measured value classes of claude's option
// surface, plus the zero value for a token the model cannot classify — the
// polarity-safe state both modes resolve without consumption (REQ-SCV-013).
type claudeOptionClass int

const (
	// claudeOptionUnknown classifies nothing: the token consumes nothing in
	// either mode, so the guard judges the token after it (fail-closed) and
	// the validator judges it too.
	claudeOptionUnknown claudeOptionClass = iota
	// claudeOptionRequiredValue marks a `<value>` synopsis: the next token is
	// the option's value in both modes — even when flag-shaped (r2) — unless
	// that token is `--`.
	claudeOptionRequiredValue
	// claudeOptionOptionalValue marks a `[value]` synopsis: the value-ness of
	// the next token depends on its shape. The validator consumes it silently
	// (ambiguity never refuses); the guard judges it (fail-closed).
	claudeOptionOptionalValue
	// claudeOptionBoolean marks a valueless synopsis: nothing is consumed in
	// either mode.
	claudeOptionBoolean
)

// claudeOptionModel maps an option token (long or short, as the help synopsis
// spells it) to its measured class. Missing keys read as claudeOptionUnknown.
type claudeOptionModel map[string]claudeOptionClass

// claudeHelpTimeout bounds the derivation probe: a claude that cannot answer
// `--help` inside it leaves the snapshot in place — no derivation failure
// changes any exit status (REQ-SCV-012).
const claudeHelpTimeout = 3 * time.Second

// claudeHelpWaitDelay bounds the post-Cancel pipe wait: exec's default of
// zero makes Output() wait for stdout EOF even after the context's Cancel
// kills the direct child, so a PATH wrapper that exits while a backgrounded
// child of its own still holds the pipe would block launcher entry until
// that grandchild exits (card-review r1, P2①). After the grace the wait is
// abandoned and the probe degrades like any other derivation failure.
const claudeHelpWaitDelay = time.Second

// claudeHelpSynopsis is the derivation probe, in the procInfoFunc seam
// pattern (session_pid.go): a package var the launcher entries call through
// and tests substitute. The default runs the resolved binary's `--help`,
// bounded in time.
var claudeHelpSynopsis = func(binaryPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), claudeHelpTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binaryPath, "--help")
	cmd.WaitDelay = claudeHelpWaitDelay
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// activeClaudeOptionModel is the model the two resume interpreters read as
// pure package state. It starts at the compile-time snapshot and is swapped
// once per process when the derivation succeeds.
var activeClaudeOptionModel = claudeOptionModelSnapshot

// claudeOptionModelSnapshot is the compile-time fallback the interpretation
// degrades to when the derivation fails or the binary is absent. Its safety
// does not depend on completeness: a snapshot gap manifests only as one of
// the three named outcomes of REQ-SCV-014 — a false guard fire (a restatable
// launch), a validator pass (claude's own error surfaces at launch), or the
// validator false-refusal residual named under plan.md §A.4 — never a
// silently leaked resume under the relaunch policy.
//
// The snapshot's source-version provenance: every entry below was measured
// from `claude --help` (Claude Code 2.1.289, 2026-10-04; the same
// measurement the predecessor's tables carried) and re-measured live in the
// run phase against the same installed version (2.1.289, 2026-10-05 —
// /Users/<u>/.local/share/claude/versions/2.1.289, `claude --help` exit 0 in
// 0.098s): all 43 entries reproduce from the re-measure, and
// --remote-control-session-name-prefix joins the required-value class from
// it (card-review round 5, P2① — REQ-SCV-015). REQ-SCV-014 requires the
// snapshot to keep carrying its source version in these comments.
//
// @MX:DEBT: the snapshot fallback may lag claude's option surface — under the compound condition (derivation down + option snapshot-absent + resume-shaped value → validator false refusal; never a silently leaked resume under relaunch)
// @MX:CEILING: correct while the installed claude's option surface stays within the classes this snapshot measured (2.1.289, re-measured live 2026-10-05); the guard side is unaffected (unknown tokens are judged — fail-closed), and a snapshot gap manifests only as a false guard fire (a restatable launch), a validator pass (claude's own error surfaces at launch), or this compound-condition false refusal
// @MX:UPGRADE: another r-class round finds a snapshot gap, or claude ships a machine-readable option surface the derivation can consume without parsing help prose
var claudeOptionModelSnapshot = claudeOptionModel{
	// Required <value> synopses (2.1.289, measured):
	"--add-dir": claudeOptionRequiredValue, "--agent": claudeOptionRequiredValue,
	"--agents": claudeOptionRequiredValue, "--allowedTools": claudeOptionRequiredValue,
	"--allowed-tools": claudeOptionRequiredValue, "--append-system-prompt": claudeOptionRequiredValue,
	"--autocompact": claudeOptionRequiredValue, "--betas": claudeOptionRequiredValue,
	"--debug-file": claudeOptionRequiredValue, "--disallowedTools": claudeOptionRequiredValue,
	"--disallowed-tools": claudeOptionRequiredValue, "--effort": claudeOptionRequiredValue,
	"--environment": claudeOptionRequiredValue, "--fallback-model": claudeOptionRequiredValue,
	"--file": claudeOptionRequiredValue, "--input-format": claudeOptionRequiredValue,
	"--json-schema": claudeOptionRequiredValue, "--max-budget-usd": claudeOptionRequiredValue,
	"--mcp-config": claudeOptionRequiredValue, "--model": claudeOptionRequiredValue,
	"--name": claudeOptionRequiredValue, "-n": claudeOptionRequiredValue,
	"--output-format": claudeOptionRequiredValue, "--permission-mode": claudeOptionRequiredValue,
	"--permission-prompts": claudeOptionRequiredValue, "--plugin-dir": claudeOptionRequiredValue,
	"--plugin-url": claudeOptionRequiredValue,
	// r5 P2①: the measured synopsis is `--remote-control-session-name-prefix
	// <prefix>` — required-value. Its absence was the r5 false refusal: a
	// legit call whose prefix value literally reads `--resume` was refused.
	"--remote-control-session-name-prefix": claudeOptionRequiredValue,
	"--session-id":                         claudeOptionRequiredValue,
	"--setting-sources":                    claudeOptionRequiredValue, "--settings": claudeOptionRequiredValue,
	"--system-prompt": claudeOptionRequiredValue, "--system-prompt-snapshot": claudeOptionRequiredValue,
	"--tools": claudeOptionRequiredValue,

	// Optional [value] synopses (2.1.289, measured): the parser refuses to
	// consume a flag-shaped token as the value, so whether the next token is
	// the value depends on its shape — the two modes resolve the ambiguity in
	// opposite directions (r3 leader ruling).
	"--cloud": claudeOptionOptionalValue, "-d": claudeOptionOptionalValue,
	"--debug": claudeOptionOptionalValue, "--from-pr": claudeOptionOptionalValue,
	"--prompt-suggestions": claudeOptionOptionalValue, "--remote-control": claudeOptionOptionalValue,
	"--teleport": claudeOptionOptionalValue, "-w": claudeOptionOptionalValue,
	"--worktree": claudeOptionOptionalValue,

	// Measured booleans: `-p` is claude's `--print` — the launcher's own `-p`
	// (profile) precedes it in the launcher segment only (r4), so the claude
	// class is recorded here to keep the two surfaces distinct.
	"-p": claudeOptionBoolean,
}

// claudeModelDeriveOnce bounds the derivation to one attempt per process:
// the probe is launch-path latency once, never per launch.
var claudeModelDeriveOnce sync.Once

// refreshActiveClaudeOptionModel derives the option model once per process
// from the claude binary the launcher resolves on PATH. Every failure — the
// binary absent, the probe erroring or timing out, a help text with no
// option lines — leaves the active model untouched, silently.
func refreshActiveClaudeOptionModel() {
	claudeModelDeriveOnce.Do(func() {
		path, err := claudeLookPath(claudeBinaryName)
		if err != nil {
			return
		}
		applyClaudeOptionModel(path)
	})
}

// applyClaudeOptionModel derives the model from the binary at path and
// installs it when the derivation succeeds. The derivation itself never
// fails outward: a failure or an unusable parse leaves the active model —
// the snapshot, unless a prior derivation swapped it — in place (REQ-SCV-012).
func applyClaudeOptionModel(binaryPath string) {
	text, err := claudeHelpSynopsis(binaryPath)
	if err != nil {
		return
	}
	m := parseClaudeOptionModel(text)
	if len(m) == 0 {
		return
	}
	activeClaudeOptionModel = m
}

// claudeHelpGapRe matches the commander layout's gap between an option
// synopsis and its description (two or more spaces). The synopsis itself may
// carry single spaces (the alias, the value placeholder).
var claudeHelpGapRe = regexp.MustCompile(`\s{2,}`)

// claudeFlagTokenRe matches one flag spelling inside a synopsis: `-d`,
// `--debug`, `--append-system-prompt`.
var claudeFlagTokenRe = regexp.MustCompile(`-{1,2}[A-Za-z][\w-]*`)

// parseClaudeOptionModel reads a `--help` text into the option model. Each
// option line's synopsis (the flag spellings plus any `<value>`/`[value]`
// placeholder) sits before the commander gap; the class is the placeholder
// it carries, applied to every spelling on the line. Classification is
// FIRST-WINS per flag: an option's synopsis precedes any prose that
// mentions it, so once a flag carries a class, later flag-prefixed lines
// naming it — description references, help cross-references, even ones the
// synopsis-indent gate admits — are prose and must not reclassify it
// (card-review r1, P1). Lines that are not option synopses — descriptions,
// wrapped description continuations, command names — start with something
// other than a flag and are skipped. The resume tokens are never carried:
// the walk special-cases them before the model is consulted (REQ-SCV-009).
func parseClaudeOptionModel(help string) claudeOptionModel {
	m := claudeOptionModel{}
	for _, line := range strings.Split(help, "\n") {
		synopsis, ok := claudeHelpSynopsisSegment(line)
		if !ok {
			continue
		}
		class := claudeOptionBoolean
		switch {
		case strings.Contains(synopsis, "<"):
			class = claudeOptionRequiredValue
		case strings.Contains(synopsis, "["):
			class = claudeOptionOptionalValue
		}
		for _, flag := range claudeFlagTokenRe.FindAllString(synopsis, -1) {
			if flag == resumeFlag || flag == resumeFlagShort {
				continue
			}
			if _, classified := m[flag]; classified {
				continue // first-classification-wins: this line is a reference, not the synopsis
			}
			m[flag] = class
		}
	}
	return m
}

// claudeHelpSynopsisMaxIndent is the widest indent a line may carry and
// still be an option synopsis. Commander aligns wrapped description
// continuations under the description column — dozens of spaces, and the
// measured help's continuations can BEGIN with a flag mention
// ("…--append-system-prompt included — sent,") that would otherwise parse
// as a valueless synopsis and silently reclassify the option — while the
// option synopses sit at the option block's shallow indent (two spaces in
// the measured 2.1.289 help).
const claudeHelpSynopsisMaxIndent = 6

// claudeHelpSynopsisSegment returns the option synopsis a help line carries:
// the flag-prefixed head of the line, before the commander gap. A synopsis
// too long for its description wraps — commander puts the description on the
// next line and leaves the synopsis alone on its own — so a flag-prefixed
// line at the option indent with no gap is a synopsis too.
func claudeHelpSynopsisSegment(line string) (string, bool) {
	indent := 0
	for indent < len(line) && (line[indent] == ' ' || line[indent] == '\t') {
		indent++
	}
	if indent > claudeHelpSynopsisMaxIndent {
		return "", false // a wrapped description continuation, not an option line
	}
	trimmed := line[indent:]
	if !strings.HasPrefix(trimmed, "-") {
		return "", false
	}
	if loc := claudeHelpGapRe.FindStringIndex(trimmed); loc != nil {
		return trimmed[:loc[0]], true
	}
	return trimmed, true
}
