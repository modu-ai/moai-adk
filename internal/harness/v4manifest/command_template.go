// Package v4manifest — thin-wrapper command generator (design §B.2).
//
// CommandTemplate is the markdown thin-wrapper the Builder GENERATE phase stamps
// per-harness as .claude/commands/harness/<name>.md. The wrapper dispatches
// /harness:<name> invocations to that harness's Runner Workflow
// (harness-<name>-run.js) via the dynamic-workflow primitive. Claude Code
// subdirectory-command resolution maps .claude/commands/harness/<name>.md to
// /harness:<name>.
//
// Storage decision (design §B.2 + E8): the template lives as a Go-embedded
// string in this package (option a — same as the Runner template), NOT as a
// file under internal/template/templates/.claude/commands/harness/. Rationale:
// the emitted per-harness <name>.md is user-owned per M1 §24.4 (the user owns
// .claude/commands/harness/); the TEMPLATE the Builder stamps from is
// moai-distributable Go source. Keeping it out of internal/template/templates/
// avoids leaking a harness-* command into the distributable template surface
// (C-HV4-005 / §24 namespace policy).
//
// The template is generic and carries NO internal-state markers (C-HV4-005
// template neutrality): no SPEC IDs, REQ tokens, AC tokens, or commit SHAs. The
// two per-harness substitutions are the harness <name> and the Runner Workflow
// filename, both derived from the validated manifest.
package v4manifest

import (
	"fmt"
	"strings"
)

// commandNamePlaceholder and commandRunnerPlaceholder are the two substitution
// tokens the Builder GENERATE phase replaces when stamping a per-harness
// command file. They are intentionally distinctive so a naive string replace
// cannot accidentally clobber unrelated content.
const (
	commandNamePlaceholder   = "__HARNESS_NAME__"
	commandRunnerPlaceholder = "__HARNESS_RUNNER_WORKFLOW__"
)

// CommandTemplate is the thin-wrapper command markdown. The Builder GENERATE
// phase writes this string (with the harness name and Runner filename
// substituted) to .claude/commands/harness/<name>.md.
//
// The wrapper is a thin router: it tells Claude to run the harness's Runner
// Workflow, which reads manifest.json and dispatches specialists per primitive
// (design §F). The wrapper itself carries NO dispatch logic — that lives in the
// Runner (the single source of truth per AC-HV4-006b).
//
// @MX:ANCHOR: [AUTO] CommandTemplate is the single source of the thin-wrapper command body.
// @MX:REASON: [AUTO] fan_in >= 3 candidate: GENERATE phase stamp, lifecycle list scan, Claude Code command resolution
const CommandTemplate = `---
description: Run the __HARNESS_NAME__ harness (dispatches its Runner Workflow which reads manifest.json)
argument-hint: "[optional task description]"
allowed-tools: Skill
---

Run the harness "__HARNESS_NAME__" by invoking its Runner Workflow __HARNESS_RUNNER_WORKFLOW__.

The Runner reads .claude/commands/harness/__HARNESS_NAME__/manifest.json and dispatches each specialist per its declared primitive. Do not re-derive specialist assignments — the manifest is the single source of truth.

Use Skill("moai") with arguments: harness-run __HARNESS_NAME__ $ARGUMENTS
`

// GenerateCommand produces the per-harness thin-wrapper command markdown for
// the given manifest. It validates the manifest first (a malformed manifest
// must not yield a wrapper that dispatches to a non-existent Runner), then
// substitutes the harness name and Runner filename into CommandTemplate.
//
// The returned string is the full file content to write at
// .claude/commands/harness/<name>.md.
func GenerateCommand(m Manifest) (string, error) {
	if err := Validate(m); err != nil {
		return "", fmt.Errorf("v4manifest: GenerateCommand: manifest invalid: %w", err)
	}
	out := CommandTemplate
	out = strings.ReplaceAll(out, commandNamePlaceholder, m.Name)
	out = strings.ReplaceAll(out, commandRunnerPlaceholder, m.RunnerWorkflow)
	return out, nil
}
