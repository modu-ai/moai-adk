---
description: >
  Multi-perspective code review with security, performance, quality, and UX analysis.
  Supports staged changes, branch comparison, and security-focused review.
  Use when performing code review, security audit, or quality assessment.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "review, code-review, security, performance, quality, ux, audit"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["review", "code review", "security audit", "quality check", "code analysis"]
  agents: ["sync-auditor"]
  phases: ["review"]
---

# Workflow: Review - Code Review

Purpose: Multi-perspective code review analyzing security, performance, quality, and UX dimensions. Produces a consolidated, prioritized report of findings.

Flow: Identify Changes -> Analyze Perspectives -> Consolidate -> Report

## Relationship to /moai loop (read-only lens, layered under loop)

`/moai review` is a **read-only, report-only lens**: it produces findings and modifies nothing. Its behavior is unchanged by the loop-sweep redefinition. The relationship to `/moai loop` is layered, not competing — `/moai review` is **layered under loop** as a queue supplier: `/moai loop`'s scan stage may INVOKE the review lenses (the security lens and the `@MX` lens) so their findings are **consumed by** the loop as fixable queue items. Standalone `/moai review` still only REPORTS those findings; the loop is what enqueues and fixes them.

Non-overlap: run a `/moai review` to SEE findings without changing anything; run a `/moai loop` to FIX the finite set of issues the scan (including review lenses) found. The loop-side view of this layering is documented in `.claude/skills/moai/workflows/loop.md` (§ Scan Stage / § Relationship to /moai review and /moai fix).

## Supported Flags

- --staged: Review only staged (git add) changes
- --branch BRANCH: Compare current branch against BRANCH (default: main)
- --security: Focus primarily on security review (OWASP, injection, auth)
- --file PATH: Review specific file(s) only
- --design: Extract design patterns from UI code and create/update `.moai/design/system.md`
- --critique: Post-build craft review focusing on subtle layering, surface elevation, token architecture, and typography hierarchy
- --lean: Over-engineering-ONLY lean audit mode. Short-circuits the comprehensive 4-perspective analysis (Security / Performance / Quality / UX) and runs ONLY the over-engineering scan with the 5-tag finding format + net-reduction summary. Read-only and advisory: applies no fixes, modifies no files, renders no PASS/FAIL verdict. See the "--lean Mode" section below.
- --repo: Repo-wide scope. With --lean, sweeps the WHOLE tree instead of the diff-scope default. With --deep, scopes the deep scan to the whole tree. Ignored without --lean or --deep.
- --deep: On-demand multi-agent DEEP vulnerability scan mode — a six-phase, adversarially-verified pipeline (architecture map -> threat model -> hunt -> adversarial verify -> report -> patch). A bare `--deep` is treated as a security-focused deep scan and composes with `--security`. Additive and non-breaking: it never modifies the existing single-pass `--security` lens. See the "--deep Mode" section below.
- --patch: (--deep only) Opt-in patch drafting. OFF by default: while `--patch` is absent, the deep scan stops after the report phase and drafts no patch. Present only when the user explicitly passes it.
- --commit SHA: (--deep diff scope) Scope the deep scan to a single commit `<SHA>` (the single-commit diff scope, alongside `--staged` and `--branch`).

## Phase 1: Identify Changes

Determine the scope of code to review:

If --staged: Use `git diff --staged` to get staged changes.
If --branch: Use `git diff {BRANCH}...HEAD` to get branch changes.
If --file: Read the specified file(s) directly.
If no flag: Use `git diff HEAD~1` for the most recent commit changes.

Collect:
- List of modified files with change types (added, modified, deleted)
- Diff summary with line counts
- Affected modules and their responsibilities

## Phase 2: Multi-Perspective Analysis

If --lean flag: SHORT-CIRCUIT this phase entirely. Skip the comprehensive 4-perspective analysis (Perspectives 1-4 below) and jump directly to the "--lean Mode — Over-Engineering-Only Lean Audit" section. The narrowness IS the feature: correctness, security, and performance findings stay in the default (non-`--lean`) comprehensive review.

[HARD] The 4 perspectives execute as a Mode-4 parallel read-only fan-out: up to 4 concurrent read-only judges — one per perspective (Security / Performance / Quality / UX) — spawned in a single turn, within the 3-5 concurrent `Agent()` ceiling (`orchestration-mode-selection.md` §C.2). The sync-auditor subagent remains the binding synthesis and verdict owner (independent skeptical quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2): the parallel judges' findings feed the sync-auditor synthesis, which renders the consolidated assessment. The fan-out changes execution shape only — never verdict ownership. (The `--team` parallel-review mode remains retired with the Agent Teams static layer; this is Mode 4 subagent fan-out, not a team.)

Per-perspective skill injection (skill-routing.md §1): each read-only judge is a `Agent(general-purpose)` spawned with the perspective's skill injected — Security → `At start, invoke Skill("moai-ref-owasp-checklist")`; Quality → `Skill("moai-foundation-quality")`; UX → `Skill("moai-ref-react-patterns")` (per `.moai/config/sections/delegation.yaml`).

At the finding stage, report every issue you find, including ones you are uncertain about or consider low-severity, each with a confidence level and an estimated severity. Do not filter for importance or confidence while finding — the verdict stage (must-pass thresholds + harmonic scoring) does the filtering downstream. The goal at this stage is coverage: surfacing a finding that later gets filtered out is preferable to silently dropping a real bug.

### Perspective 1: Security Review

- OWASP Top 10 compliance check
- Input validation and sanitization
- Authentication and authorization logic
- Secrets exposure (API keys, passwords, tokens)
- Injection risks (SQL, command, XSS, CSRF)

#### Dependency Vulnerability Scan

Enumerate project manifest files and run a vulnerability scan for each detected file:
`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.

Auto-detect language from project markers; run the dependency vulnerability scan via a per-spawn `Agent(general-purpose)` security reviewer (security whitelist + OWASP instructions per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 9) with the detected manifest.
Full OWASP checklist: load the retained `moai-ref-owasp-checklist` skill (OWASP Top 10 + dependency-scan + secrets patterns), which supplements the inline dependency and secrets scans above.

#### Secrets Scan (Incremental with Checkpoint)

Scan git history for credential leaks incrementally. A last-scanned-SHA checkpoint is recorded under `.moai/state/` (`.moai/state/secrets-scan-checkpoint.txt` — the HEAD SHA of the last completed scan).

Where a checkpoint SHA exists, scan only the new commit range plus the working tree, then update the checkpoint to the current HEAD:

```bash
git log -p <last-sha>..HEAD -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'
```

Where no checkpoint exists (first run) OR an explicit full-scan flag is passed, run the full-history scan covering all commits reachable via `--all`, then record the checkpoint:

```bash
git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'
```

Cross-reference findings against `.gitignore` to distinguish historical leaks from working-tree exposure.
This scan is separate from working-tree-only scanners. The incremental range plus checkpoint update produces the same coverage over time as the former every-review full scan — no finding class is dropped; only redundant re-scanning of already-covered history is removed.

#### Data Isolation Check

Verify the following boundaries are intact:
- **Multi-tenant**: No cross-tenant data flow; tenant ID is enforced at every query boundary.
- **PII separation**: PII is never written to logs, metrics, or telemetry endpoints.
- **Shared-state leakage**: No mutable globals that carry request-scoped data across concurrent requests.

For all three subsections above, the canonical security procedure is this workflow's `--security` phase (Perspective 1: Security Review) plus the sync-phase dependency-manifest quality gate (`workflows/sync/quality-gates-quality.md`), with the OWASP baseline supplied by `Skill("moai-ref-owasp-checklist")`.

If --security flag: This perspective receives primary focus with deeper analysis.

### Perspective 2: Performance Review

- Algorithmic complexity analysis (O(n) considerations)
- Database query efficiency (N+1 queries, missing indexes)
- Memory usage patterns (leaks, excessive allocation)
- Caching opportunities
- Bundle size impact (frontend changes)
- Concurrency safety (race conditions, deadlocks)

### Perspective 3: Quality Review

- TRUST 5 compliance (Tested, Readable, Unified, Secured, Trackable)
- Naming conventions and code readability
- Error handling completeness
- Test coverage for changed code
- Documentation for public APIs
- Consistency with project patterns and conventions

### Perspective 4: UX Review

- User flow integrity (do changes break existing flows?)
- Error states and edge cases from user perspective
- Accessibility compliance (WCAG, ARIA)
- Loading states and feedback mechanisms
- Breaking changes in public interfaces

### Native /code-review compose (Axis A)

Where native `/code-review` is auto-invocable, the orchestrator MAY invoke it via `Skill("code-review")` as one Phase 2 finding source, covering the correctness-bug + reuse/simplification/efficiency portion; its findings feed the sync-auditor synthesis. The Security review (Perspective 1), `@MX` tag-compliance (Phase 3), UX review (Perspective 4), and design review (Phase 5) composition is preserved — native `/code-review` augments, never replaces, the MoAI-specific perspectives.

Conditional-PROGRAMMATIC caveat: before relying on `Skill("code-review")`, verify auto-invocability at runtime — a bundled skill with `disable-model-invocation: true`, a session with `disableBundledSkills`, or a denied `Skill` tool all remove auto-invocability.

Compose fallback: where native `/code-review` is not auto-invocable, Phase 2 runs entirely via the sync-auditor as today. See `native-invocation-model.md` Axis A.

## Phase 3: MX Tag Compliance Check

After perspective analysis, check @MX tag compliance for changed files:

- New exported functions: Should have @MX:NOTE or @MX:ANCHOR
- High fan_in functions (>=3 callers): Must have @MX:ANCHOR
- Dangerous patterns: Should have @MX:WARN
- Untested public functions: Should have @MX:TODO

Report missing or outdated @MX tags as findings.

## Phase 4: Report Consolidation

Produce a consolidated review report organized by severity:

### Report Structure

```markdown
## Code Review Report - {target}

### Critical Issues (must fix)
- [SECURITY] file:line: Description
- [PERFORMANCE] file:line: Description

### Warnings (should fix)
- [QUALITY] file:line: Description
- [UX] file:line: Description

### Suggestions (nice to have)
- [QUALITY] file:line: Description

### MX Tag Compliance
- Missing tags: N
- Outdated tags: N
- Compliant files: N/M

### Overall Assessment
- Security: PASS/FAIL
- Performance: PASS/WARN
- Quality: PASS/WARN
- UX: PASS/WARN
- TRUST 5 Score: N/5
```

## Phase 6: Next Steps

Present options via AskUserQuestion:

- Auto-fix issues (Recommended): Run /moai fix to automatically resolve Level 1-2 issues found in the review. Critical and complex issues will require manual attention.
- Create fix tasks: Create TaskList items for each finding so they can be addressed individually. Useful for team coordination.
- Export report: Save the review report to .moai/reports/ for future reference and tracking.
- Dismiss: Acknowledge the review without taking immediate action.

## --lean Mode — Over-Engineering-Only Lean Audit

When the --lean flag is present, the review runs ONLY this mode and nothing else. It is a narrow, single-purpose over-engineering audit — a focused "what can be cut" lens. The hard scope boundary is the mechanism's whole value: by EXCLUDING correctness bugs, security findings, and performance findings (those stay in the default comprehensive 4-perspective review), the lean audit gives a high-signal leanness scan that the broad review dilutes across four perspectives. The narrowness IS the feature.

The mode is inspired by the "lazy senior dev" minimalist-coding review pattern: a read-only, one-shot audit that applies NO fixes, hunts ONLY over-engineering, emits findings under 5 fixed tags, and closes with a net-reduction summary.

### Scope (two variants — review vs audit split)

The --lean mode supports two scopes, mirroring the diff-vs-repo split of the minimalist-coding audit pattern:

- Diff-scope (default): audit ONLY the changed code, reusing the existing Phase 1 scope selection (`--staged` / `--branch BRANCH` / `--file PATH`, or the most recent commit when no scope flag is given). This is the "review the changes" variant.
- Repo-scope (with --repo): sweep the WHOLE tree. This is the "sweep everything" variant. The `--repo` flag is honored only in --lean mode.

Both variants run the identical 5-tag scan and net-reduction summary below; only the set of files scanned differs.

### The 5 finding tags

The mode emits findings under exactly these 5 fixed tags, and no others:

| Tag | What it flags |
|-----|---------------|
| `delete:` | Unused or speculative code — dead branches, never-called helpers, write-only config |
| `stdlib:` | Reimplemented standard library — hand-rolled logic that the language's standard library already provides |
| `native:` | A dependency or code duplicating a platform-native feature the platform already provides |
| `yagni:` | Premature generality — single-implementation abstraction, single-caller indirection layer, dead config knob |
| `shrink:` | Logic reducible to fewer lines without loss of clarity |

Language neutrality: `stdlib:` names "the language's standard library" and `native:` names "a platform-native feature" generically — across all supported languages. Do NOT reference any single language's standard library module, package-manager name, or platform feature by name; the same 5 tags apply whatever language the audited code is written in.

### Output format

Emit one finding per line in this format:

```
L<line>: <tag> <what to cut>. <replacement>. [path]
```

Report every over-engineering finding you observe; the mode is finding-only and does not filter for importance. Do NOT report correctness, security, or performance issues here — those belong to the default comprehensive review (Perspectives 1, 2 above), not the lean audit.

### `@MX:DEBT` one-directional cross-link

A `yagni:` finding (single-implementation abstraction, single-caller layer, dead config) is exactly the deliberate-simplification case that an `@MX:DEBT` marker records. Before reporting a `yagni:` finding as a fresh discovery, consult the existing deferred-debt harvest:

- Read the `@MX:DEBT` harvest via `moai mx query --kind DEBT` (read-only — never write or modify markers).
- When a `yagni:` finding lands on a site that ALREADY carries an `@MX:DEBT` marker, annotate it as already-tracked deferred debt rather than re-flagging it as new:

  ```
  L<line>: yagni: <site> [already tracked @MX:DEBT — deferred]. [path]
  ```

  This avoids re-surfacing a simplification the author already recorded and deliberately deferred — the noise the `@MX:DEBT` mechanism exists to suppress.

The link is one-directional: the lean audit READS the `moai mx query --kind DEBT` harvest but NEVER creates, modifies, or removes an `@MX:DEBT` marker. Authoring `@MX:DEBT` markers remains the run-phase author's responsibility; an advisory read-only lens must not mutate source.

### Closing summary

Close the audit with exactly one of these forms:

- When at least one removal is warranted (diff-scope): `net: -<N> lines possible`
- When at least one removal is warranted (repo-scope, dependencies also removable): `net: -<N> lines, -<M> deps possible`
- When nothing warrants removal: the literal line `Lean already. Ship.`

### Read-only, advisory, no verdict

The --lean mode is read-only and advisory. It applies NO fixes, modifies NO files, and renders NO PASS/FAIL verdict — it is distinct from `/moai clean` (which removes code) and from the sync-auditor gate (which scores a verdict). It produces only a "what can be cut" list plus the net-reduction estimate. Remediation routes through the existing Phase 6 Next Steps (run `/moai fix`, create fix tasks, export the report, or dismiss) — the same as the rest of this skill.

### Doctrine cross-references (reuse, do not duplicate)

- The 5 lean tags are the OPERATIONAL scan surface for the over-engineering anti-patterns already catalogued in `.claude/skills/moai/references/anti-patterns.md` (the Premature Abstraction and Over-Engineering categories, mapped to Agent Core Behavior #4 Enforce Simplicity). Consult that catalogue for the wrong/right examples; this section does not restate it.
- The lean audit is the post-hoc DETECTION counterpart to the pre-code PREVENTION ladder in `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (the simplicity decision ladder). The ladder prevents over-engineering before code is written; the lean audit detects what slipped through. Consult that ladder for the ordered prevention steps; this section does not restate it.

## --deep Mode — On-Demand Multi-Agent Deep Vulnerability Scan

When the `--deep` flag is present, `/moai review` runs an on-demand, adversarially-verified deep vulnerability scan. `--deep` is a *depth* modifier layered on the existing *breadth* review: it reuses the same scope-selection machinery (`--repo` / `--staged` / `--branch B` / `--commit SHA`) and the Security perspective, then adds independent adversarial verification, a machine-readable findings artifact, and (opt-in) reviewer-vouched patch drafts. A bare `--deep` is treated as a security-focused deep scan and composes with `--security`; `--security --deep` is the explicit form. `--deep` is additive and non-breaking — the existing single-pass `--security` lens is unchanged.

This deep scan is the heavy, explicitly-invoked layer. It composes with the lighter always-on review lens rather than replacing it. It lives entirely under `/moai review --deep`; there is no separate top-level security subcommand (the former `/moai security` entry was retired and is not revived — the deep scan is a mode of `/moai review`, and no `security` / `audit` / `sec` alias is added).

### Job menu — mapped onto review scope flags

`/moai review --deep` offers the following jobs, each mapped onto the existing scope-selection mechanism (Phase 1). The orchestrator collects the job + scope + `--patch` opt-in via `AskUserQuestion` BEFORE the pipeline launches (see the User-Interaction Boundary below).

| Job | Scope flag | What it scans |
|-----|-----------|---------------|
| Scan the whole repository | `--repo` | The entire working tree |
| Scan staged changes | `--staged` | The staged diff |
| Scan a branch diff | `--branch B` | `current...B` |
| Scan a single commit | `--commit SHA` | Exactly the one commit `<SHA>` |
| Draft patches for confirmed findings | `--patch` (opt-in) | Enables the patch phase (see below); OFF by default |

`--patch` is an independent opt-in and is OFF by default: while `--patch` is absent, the deep scan stops after the report phase and drafts, writes, or applies no patch. Patch drafting happens only when `--patch` is explicitly present.

### Prerequisite & graceful degradation ladder

The deep scan's PRIMARY execution path requires Dynamic Workflows (Claude Code v2.1.154+). Availability is checked BEFORE launch (pipeline agents cannot prompt mid-run, so the degradation choice is made up front). Observable signals that mean Dynamic Workflows are unavailable: the runtime version is below v2.1.154, the `CLAUDE_CODE_DISABLE_WORKFLOWS` environment variable is set, or workflows are disabled in settings. When Dynamic Workflows are unavailable the scan degrades gracefully rather than failing:

| Rung | Condition | Path | Rigor |
|------|-----------|------|-------|
| PRIMARY | Dynamic Workflows available (v2.1.154+, not disabled) | Runtime-constructed `Workflow()` — full fan-out across phases | Full |
| FALLBACK | Workflows unavailable, but bounded parallel is viable | Mode-4 bounded parallel fan-out (3-5 concurrent read-only agents), findings batched | **2-of-3 quorum preserved** — degradation drops concurrency/scale, NOT verification rigor |
| DEGRADED | Neither viable | Single-pass `/moai review --security` plus the native `/security-review` | Reduced (single-pass, no per-finding 3-voter panel) — the report MUST label this rung as rigor-reduced so a reduced result is never mistaken for a full adversarially-verified scan |

The FALLBACK rung is a genuine degradation of scale only: it preserves the 3-voter panel and the 2-of-3 quorum. Only the DEGRADED last-resort rung reduces rigor, and it self-labels as such in the report.

The shipped artifact of this mode is THIS playbook (template-safe markdown), not a static workflow script. The orchestrator constructs the `Workflow()` at runtime from the phase descriptions below. A user MAY save the generated script into their own `.claude/workflows/` for reuse, but no static script is shipped.

### The six-phase pipeline

The deep scan executes six phases in order. Phases 1-4 are strictly READ-ONLY against the working tree; phase 5 writes only under the results directory; phase 6 writes only inside an isolated scratch clone.

1. **Architecture map** — a single read-only recon agent (or `Explore`) maps modules, entry points, and trust boundaries. READ-ONLY (Read / Grep / Glob only).
2. **Threat model** — a single agent consumes the phase-1 map and enumerates the attack surface. READ-ONLY. Loads `Skill("moai-ref-owasp-checklist")` for the baseline vocabulary.
3. **Vulnerability hunt** — a parallel fan-out of hunt agents (per area / per manifest) surfaces candidate findings. READ-ONLY. Each hunt agent loads the relevant security reference skill(s) on demand via `Skill()` injection (NOT via static frontmatter preload): `Skill("moai-ref-owasp-checklist")` for web-app classes, `Skill("moai-ref-llm-security")` for LLM/agentic classes, `Skill("moai-ref-secops")` for CI/CD, container, and API-operational classes, and `Skill("moai-ref-supply-chain")` for dependency and provenance classes.
4. **Adversarial verification** — each candidate finding is cross-examined by an independent 3-voter panel (detail below). READ-ONLY.
5. **Report** — the orchestrator (or a synthesizer agent) writes the results directory (schema below). Writes ONLY under the results directory.
6. **Patch** — gated by `--patch`; drafts one reviewer-vouched patch per confirmed finding in an isolated scratch clone (detail below). Writes ONLY inside the scratch clone; never the live tree.

The scan reasons across all supported languages equally — it names vulnerability classes and platform-native features generically, not by any single language's toolchain, standard library, or package manager.

### Phase 4 — adversarial verification panel

Every candidate finding from the hunt phase passes a 3-voter panel before it may enter the report. The three voters are perspective-diverse and REFUTE-skewed:

- **REACHABILITY** — Is the vulnerable code path actually reachable by attacker-controlled input? (affirm | refute)
- **IMPACT** — If reached, what is the concrete blast radius? Is the stated impact real, not theoretical? (affirm | refute)
- **DEFENSES** — Do existing defenses (validation, authorization, framework guards) already neutralize this? (affirm undefended | refute defended)

Quorum admission rule:

- A finding is ADMITTED to the report only when at least **2 of the 3** voters affirm it (the 2-of-3 quorum).
- When all 3 voters affirm (unanimous), the finding's confidence MAY be stated as "high".
- A non-unanimous panel (2-of-3, not 3-of-3) caps the finding's stated confidence at medium and MUST NOT state "high" — the cap binds confidence only, not severity (a high-severity finding may keep high severity while its confidence is capped at medium).
- A candidate that fails the 2-of-3 quorum is REJECTED: it is excluded from the confirmed-findings body and recorded only under a trailing "Unconfirmed candidates" appendix, never mixed with confirmed findings.

Voter independence: each voter is a separate agent given ONLY the finding claim plus the surrounding code context — never the hunt agent's own reasoning chain. Each voter's default posture is to disprove the finding (REFUTE-skewed), reusing the skeptical-evaluation stance. No voter reuses the hunt agent's reasoning as its sole basis.

### Phase 6 — patch drafting & reviewer vouch (`--patch` only)

When `--patch` is present and at least one confirmed finding exists, each patch is drafted in isolation and independently reviewed:

- **Drafter** — an `Agent(isolation: "worktree")` operates in an isolated scratch clone of the repository, drafts the minimal fix for exactly one finding, and emits a unified diff. The drafter never touches the user's live working tree.
- **Independent reviewer** — a SEPARATE agent spawn (distinct from the drafter) reads the drafted diff plus the finding and surrounding code (read-only) and vouches, all-or-nothing, for THREE claims: (a) it addresses only the one finding, (b) it introduces no new vulnerability, (c) it leaves behavior otherwise unchanged.
- When all three claims are vouched, the diff is written into the results directory as a patch artifact. When the reviewer cannot vouch for all three, the pipeline emits a short explanatory note for that finding instead of a patch; other findings' patches are unaffected.

Patches are NEVER auto-applied: the pipeline never runs `git apply`, and never stages, commits, or pushes any drafted patch to the user's repository. Each confirmed-and-vouched finding yields exactly one patch artifact the user applies manually — one finding = one patch = one PR. The drafter and reviewer are always different spawns (independence is a hard requirement).

### Results directory — a report, not a SPEC

The deep scan writes its outputs to a timestamped results directory classified as a REPORT (an analysis of existing code), under `.moai/reports/` — never under `.moai/specs/` (a scan is a report, not a specification). The timestamp is injected as an input argument or stamped after the run returns; it is never generated inside the workflow script body (so resume-caching stays valid).

```
.moai/reports/security-deepscan-<timestamp>/
  .gitignore        # a single line: *  (so a stray `git add` can never sweep the results into a commit)
  report.md         # the human report
  findings.jsonl    # machine-readable, one finding per line
  revision.json     # the revision stamp
  patches/          # only when --patch: one F<i>.patch or F<i>.note.md per confirmed finding
```

Every results directory ships its own `.gitignore` (ignoring its entire contents) so a stray `git add` never sweeps a scan into a commit.

`report.md` — each confirmed finding carries a stable finding ID (`F1`, `F2`, ...) plus five fields:

```
### F<i> — <title>
- Severity:        <critical | high | medium | low>
- Confidence:      <high | medium>        (medium max when the panel was non-unanimous)
- Impact:          <concrete blast radius>
- Exploit scenario: <step-by-step reachability narrative>
- Recommendation:  <fix direction>
- Panel:           REACHABILITY=<affirm|refute> IMPACT=<...> DEFENSES=<...> (<N>/3)
```

Rejected candidates appear ONLY under a trailing `## Unconfirmed candidates (did not reach 2-of-3)` appendix.

`findings.jsonl` — machine-readable, exactly one finding per line (JSONL); each line carries the finding ID plus its structured fields:

```json
{"id":"F1","severity":"high","confidence":"medium","title":"...","impact":"...","exploit":"...","recommendation":"...","panel":{"reachability":true,"impact":true,"defenses":false},"location":{"path":"...","hint":"..."}}
```

`revision.json` — the revision stamp records which commit was scanned, the effort tier used, and whether the working tree was included in the scan:

```json
{"scanned_commit":"<sha>","effort_tier":"<low|medium|high|xhigh|max>","working_tree_included":true,"scope":"repo|branch|commit|staged|file","generated_at":"<injected-timestamp>"}
```

### User-interaction boundary

All user decisions for the deep scan — scope selection, the `--patch` opt-in, and the degradation-path choice — are collected by the orchestrator via `AskUserQuestion` BEFORE the pipeline launches. The pipeline agents (recon, threat-modeler, hunters, voters, drafter, reviewer) never prompt the user: an agent that lacks a required input returns a structured blocker report to the orchestrator, which resolves it and re-delegates. No pipeline or workflow agent asks the user anything directly.

### Edge cases

- **Zero confirmed findings** — the scan completes cleanly, writes a results directory stating "0 confirmed findings", and drafts no patch. No error, no empty patch.
- **Confirmed finding but `--patch` absent** — the report is written; the patch phase does not run.
- **Reviewer cannot vouch** — a note (not a patch) is emitted for that finding; other findings' patches are unaffected.
- **Non-unanimous panel on a high-severity finding** — severity may remain high, but the stated CONFIDENCE is capped at medium (severity and confidence are independent axes).
- **Dynamic Workflows disabled** — the availability signal is checked BEFORE launch; the degradation rung is chosen up front, not mid-run.
- **Large repository, hundreds of candidates** — the PRIMARY `Workflow()` path handles scale; the FALLBACK Mode-4 path batches findings within the 3-5 concurrent ceiling. Neither path drops the quorum.

## Task Tracking

[HARD] Task management tools mandatory:
- Each critical finding tracked as a pending task via TaskCreate
- Warnings grouped by file as aggregate tasks
- Suggestions listed in report but not tracked as tasks

## Phase 5: Design Review (Conditional)

When to run: --design or --critique flag is present, OR changed files include UI components (tsx, jsx, vue, svelte, css, scss)

### --design: Extract Design Patterns

Agent: per-spawn `Agent(general-purpose)` frontend specialist (frontend whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 8) — inject `At start, invoke Skill("moai-ref-react-patterns")` and `Skill("moai-domain-frontend")` (per `.moai/config/sections/delegation.yaml` domain_skills.frontend; skill-routing.md §1)

Tasks:
1. Scan UI files for repeated patterns: spacing values, radius values, color tokens, button/card patterns, depth strategy (borders vs shadows)
2. Identify existing design conventions and inconsistencies
3. If `.moai/design/system.md` exists: Compare extracted patterns against system.md, report deviations
4. If `.moai/design/system.md` does not exist: Create system.md from extracted patterns
5. Present extraction summary with option to update system.md

Output: Design pattern report with deviation list (file:line references)

### --critique: Post-Build Craft Review

Agent: per-spawn `Agent(general-purpose)` frontend specialist (frontend whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 8) — inject `At start, invoke Skill("moai-ref-react-patterns")` and `Skill("moai-domain-frontend")` (per `.moai/config/sections/delegation.yaml` domain_skills.frontend; skill-routing.md §1)

Tasks:
1. Read `.moai/design/system.md` for design direction context
2. Review built UI against craft principles:
   - **Composition**: Layout rhythm, proportions, focal point clarity
   - **Craft**: Spacing grid adherence, typography hierarchy, surface elevation consistency
   - **Content**: String coherence, data truthfulness
   - **Structure**: CSS quality (no negative margin hacks, no absolute positioning escapes)
3. Run quality checks: swap test, squint test, signature test, token test
4. Identify specific locations where defaults won over intentional design decisions
5. Provide actionable rebuild recommendations with file:line references

Output: Craft critique report with severity-ranked findings and rebuild suggestions

## Agent Chain Summary

- Phase 1: MoAI orchestrator (change identification via git)
- Phase 2-3: sync-auditor subagent (multi-perspective analysis; the Security perspective receives deeper focus when --security is set) — the dependency vulnerability sub-scan additionally delegates to a per-spawn `Agent(general-purpose)` security reviewer
- Phase 4-5: MoAI orchestrator (consolidation and user interaction)
- Phase 5 (conditional): per-spawn `Agent(general-purpose)` frontend specialist (if --design or --critique)

## Execution Summary

1. Parse arguments (extract flags: --staged, --branch, --security, --file, --design, --critique)
2. (The `--team` review fan-out is retired — always delegate to the sync-auditor subagent)
3. Identify code changes (git diff based on flags)
4. Delegate multi-perspective review to the sync-auditor subagent
5. Check @MX tag compliance for changed files
6. If --design or --critique: Run design review phase 4.5 (per-spawn `Agent(general-purpose)` frontend specialist per the frontend whitelist)
7. Consolidate findings by severity
8. Present report with next step options

---

Version: 1.0.0
