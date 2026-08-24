---
id: SPEC-SEC-SCAN-SURFACE-001
title: Security scan surface — cheap pre-write ast-grep gate + PostToolUse guardian merge
version: 0.1.0
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tags: [security, hook, pretooluse, posttooluse, ast-grep, performance]
tier: M
---

## HISTORY

- 2026-08-24 — v0.1.0 — Initial draft. Evidence base: `.moai/reports/t217/investigation.md`
  (measured on worktree tree `c4e90cd58`). Operator decisions A (keep the pre-write blocking
  capability, make it cheap) and B (merge the PostToolUse guardian into the post-tool handler)
  were taken before authoring and are not re-opened here.

---

## §A Context

This is a **security-surface change**. It touches the only write-path deny in the MoAI hook
system. Cost reduction is the motive; preserving the deny verdict is the constraint. Any change
that makes the gate cheaper by making it deny less is a failure of this SPEC, not a trade-off
within it.

Five findings from the investigation report ground the work. Each is cited, not re-derived.

| # | Finding (investigation.md) | Consequence for this SPEC |
|---|---|---|
| Claim 1 | The pre-write deny reason string has never appeared in a runtime form across 12,265 + 3,373 transcripts on this machine (control: `BRANCH_GUARD_VIOLATION:` appears in 151 files, so deny reasons *are* recorded). | The capability is unexercised **here**; it is not established as unexercised for distributed users. Keep it. |
| Claim 2 | The mechanism reaches an `error`-severity finding under the local ruleset (`sec-hardcoded-credential`). | The deny path is live, not dead code. |
| Claim 3 | `FindRulesConfig` resolves to `""` in a card worktree; `sg` then fails with `No ast-grep project configuration is found`. Every Write in every worktree session pays a temp-file write plus an `sg` spawn and receives zero findings. | Item A2 — the dominant waste, and it removes **no** finding, because there were none to remove. |
| Claim 4 | `IsSupportedExtension` admits 15 languages; the shipped ruleset carries rules for 4 (go / javascript / typescript / python). | Item A3 — 11 languages pay the full cost for a guaranteed-empty result. |
| Claim 5 | PreToolUse (ast-grep, 26 rules, 14 at `error`) and PostToolUse (in-process regex, 10 vulnerability classes) are different engines over different rule sets. **Neither is a superset of the other.** | Constrains item A1: the pre-filter must not be built from the regex table, and constrains item B: the two advisories are not redundant. |
| Claim 6 + side observation | Merging the PostToolUse entry removes one process spawn and **no scan**. The guardian advisory was observed reaching the session live during the investigation. | Item B is a process merge, not a capability removal — and the advisory channel it carries is real. |

Fresh measurements taken while authoring this SPEC (worktree `c4e90cd58`, branch
`security-scan-surface`):

- `ls .moai/config/astgrep-rules` → `No such file or directory` (Claim 3 still holds on this tree).
- `command -v sg` → `/opt/homebrew/bin/sg` (the spawn is real, not a missing-binary no-op).
- The shipped ruleset carries 26 rules across 11 files; `grep -rh "^language:"` over
  `internal/template/templates/.moai/config/astgrep-rules/` yields `20 go, 2 javascript,
  2 python, 2 typescript`.
- `internal/hook/registry.go:180` `mergeHandlerOutput` already accumulates
  `hookSpecificOutput.additionalContext` across handlers with a `"\n"` join, and
  `internal/hook/official_protocol_fix_test.go` already guards that behaviour.

### §A.1 Two premises of the card were found to be inaccurate

Both are recorded here so the plan does not inherit them.

1. **"Both handlers write `hookSpecificOutput` JSON to stdout, so the `additionalContext`
   payloads must be combined."** The collision is not `additionalContext` against
   `additionalContext`. The post-tool handler emits its LSP / MX / ast-grep / navigator advisories
   on the **top-level `systemMessage`** field and leaves `hookSpecificOutput.additionalContext`
   empty; the guardian emits `hookSpecificOutput.additionalContext` and no `systemMessage`. The
   regression risk the card names is real but sits one field over: a naive merge that folds the
   guardian text into `systemMessage`, or that replaces the whole `HookSpecificOutput` struct,
   drops one of the two. The requirement below is therefore stated over **both** carrier fields.
2. **"`moai hook security-scan` may have other callers — verify."** Verified: the only references
   are `internal/cli/hook.go` (the subcommand registration), its own tests, the wrapper script
   `handle-security-scan.sh` (+ `.sh.tmpl`), and the `settings.json` (+ `.json.tmpl`) entry this
   SPEC removes. There is no third-party production caller in-tree. Retaining the subcommand is a
   backward-compatibility choice for user projects whose `settings.json` predates the merge — a
   real reason, but not the "it is already called elsewhere" reason the card assumed.

---

## §B Requirements

Notation: GEARS. `<subject>` names the component, not "the system".

### B.1 The invariant that bounds every other requirement

- **REQ-SSS-001** (Ubiquitous) — The pre-write security gate shall deny every write payload that
  the pre-implementation gate denies. A payload that produced a deny before this change shall
  produce a deny after it, for the same rule set and the same project root.
- **REQ-SSS-002** (Ubiquitous) — The pre-write security gate shall not introduce a new deny. A
  payload that was allowed before shall be allowed after.

REQ-SSS-001 and REQ-SSS-002 together state that the deny verdict is **unchanged as a function**;
only the work done to compute it changes.

### B.2 Item A2 — no rules config, no scan

- **REQ-SSS-003** (Where / When) — **Where** no ast-grep rules configuration resolves for the
  project root, **when** a Write payload reaches the pre-write security gate, the gate shall
  return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-004** (Ubiquitous) — The gate shall resolve the rules configuration exactly once per
  invocation, and shall pass the resolved path to the scanner rather than letting the scanner
  resolve it a second time.

### B.3 Item A3 — no rules for this language, no scan

- **REQ-SSS-005** (Where / When) — **Where** the resolved rules configuration declares no rule
  whose `language` matches the payload file's language, **when** a Write payload reaches the
  gate, the gate shall return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-006** (Ubiquitous) — The covered-language set shall be derived by reading the
  resolved configuration and the rule files it names. The gate shall not carry a second
  hardcoded language list.
- **REQ-SSS-007** (When) — **When** the resolved configuration cannot be read, parsed, or walked,
  the gate shall treat the covered-language set as unknown and shall escalate to `sg` rather than
  skipping. An unreadable config is never evidence that a language is uncovered.

### B.4 Item A1 — literal pre-filter derived from the rules themselves

- **REQ-SSS-008** (Where / When) — **Where** a literal pre-filter has been derived for the
  payload's language, **when** the payload content contains none of the derived literal tokens,
  the gate shall return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-009** (Ubiquitous) — The pre-filter shall be derived only from the `error`-severity
  rules of the resolved configuration that apply to the payload's language. Rules at `warning`,
  `info`, or `hint` severity shall not contribute tokens, because they cannot produce a deny.
- **REQ-SSS-010** (Where) — **Where** any applicable `error`-severity rule yields no token that
  is provably mandatory for a match, the pre-filter shall be marked underivable for that
  language, and the gate shall escalate to `sg` unconditionally for that language.
- **REQ-SSS-011** (Unwanted) — The pre-filter shall not be derived from `GuardianPatterns()` or
  from any other regex table that is not the resolved ast-grep rule set. (Justification: §C.)
- **REQ-SSS-012** (Ubiquitous) — Pre-filter derivation shall be a pure function of the resolved
  rule set, so that it is testable without a filesystem and without `sg`.

### B.5 Item B — merge the PostToolUse guardian into the post-tool handler

- **REQ-SSS-013** (Ubiquitous) — The regex guardian buffer scan shall run within the post-tool
  handler's process for `Write` / `Edit` / `MultiEdit` events.
- **REQ-SSS-014** (When) — **When** both the post-tool handler's own advisory text and the
  guardian's advisory text are non-empty for a single event, the emitted hook output shall carry
  both. Neither carrier field (`systemMessage`, `hookSpecificOutput.additionalContext`) shall be
  overwritten by the other's content, and neither advisory shall be dropped.
- **REQ-SSS-015** (Ubiquitous) — The guardian's advisory shall remain advisory. The merge shall
  introduce no deny, no `decision: block`, and no non-zero exit on the PostToolUse path.
- **REQ-SSS-016** (Ubiquitous) — The `moai hook security-scan` subcommand shall remain
  registered and invocable with its current stdin / stdout contract, so a user project whose
  `settings.json` still names it keeps working.
- **REQ-SSS-017** (Ubiquitous) — The `handle-security-scan.sh` PostToolUse entry shall be absent
  from the `Write|Edit|MultiEdit` matcher in both `.claude/settings.json` and
  `internal/template/templates/.claude/settings.json.tmpl`.

### B.6 Distribution and disclosure

- **REQ-SSS-018** (Ubiquitous) — Every file changed under `.claude/` shall have its mirror under
  `internal/template/templates/.claude/` changed in the same commit, and every touched hook
  wrapper shall have its `.sh` / `.sh.tmpl` pair changed together.
- **REQ-SSS-019** (Ubiquitous) — Template content shall remain neutral: no SPEC ID, no REQ token,
  no date, no commit SHA, no absolute macOS path, no `CLAUDE.local` reference in any file under
  `internal/template/templates/`.
- **REQ-SSS-020** (Ubiquitous) — The pull request for this SPEC shall declare in its **title**
  and in the **first sentence of its body** that this is a change to the security scan surface,
  and shall not present itself as a performance improvement. The cost reduction is the method;
  the security surface is the subject.

---

## §C Design decision — where the pre-filter comes from

The card offered two candidate pre-filters. This section records the choice and its cost, so the
run phase does not re-litigate it.

### C.1 Rejected — reuse `GuardianPatterns()` as the pre-filter

The simplicity ladder argues for reusing the existing regex table rather than building a second
one. It was evaluated and rejected on correctness grounds, not on effort.

Investigation Claim 5 establishes that the two rule sets are not in a superset relation. A
"guardian regex misses ⇒ skip `sg`" pre-filter therefore narrows the deny to the intersection of
the two sets, which violates REQ-SSS-001. Concretely, the following `error`-severity ast-grep
rules in the shipped ruleset have **no** counterpart in the ten guardian classes, and their
constructs would become undeniable:

- `go-error-ignored-blank` — a discarded error return (`$_, $ERR = $FUNC($$$ARGS)`). No guardian
  class covers error handling at all.
- `go-defer-in-loop` — a `defer` inside a loop body. No guardian class covers resource lifetime.
- `sec-hardcoded-jwt-signing-key` — `SignedString([]byte("..."))`. The guardian
  `hardcoded-secret` class matches secret-shaped literals, not this call shape.
- `sec-template-injection-html` — `template.HTML($USER_INPUT)`. The guardian `dom-injection-xss`
  class is browser-DOM shaped and does not cover Go's `html/template` escape hatch.

Only the credential-shaped rules (`sec-hardcoded-api-key`, `sec-hardcoded-credential`) overlap
the guardian's `hardcoded-secret` class, and even there the guardian's regex and the rule's
`regex:` constraint are separately authored and not proven co-extensive.

The reuse is additionally **fragile against drift**: it couples the deny reachability of the
ast-grep rule set to an unrelated table that the guardian SPEC is free to edit. A future guardian
edit could silently narrow the write-path deny with no signal.

### C.2 Chosen — derive the pre-filter from the resolved `error`-severity rules

The pre-filter is a set of **literal substrings** extracted from the rules that can actually
produce a deny, checked against the payload with `strings.Contains`. No second regex table is
introduced; no regex engine is needed at all for the pre-pass, which makes it cheaper than either
candidate the card sketched.

Soundness is by construction plus a fail-open escape hatch:

- A token is admitted only when it is **mandatory** for the rule to match — a literal run in the
  rule's `pattern` outside any metavariable (`$X`, `$$$X`), or a mandatory literal prefix of a
  `regex:` constraint. For a composite rule, `all:` contributes the union of its children's
  mandatory tokens (any one suffices to escalate); `any:` contributes only if **every** branch
  yields a token, since a branch with none can match with none of the others' tokens present.
- Any rule shape the extractor does not confidently understand (`inside:`, `has:`, `follows:`,
  `kind:`-only rules, a `regex:` with no literal anchor, a pattern that is entirely
  metavariables) marks the **whole language** underivable (REQ-SSS-010), and that language then
  escalates unconditionally. Unknown is never treated as absent.

The honest cost of this choice, stated so it is not discovered later:

- `go-error-ignored-blank` contributes tokens as weak as `_` and `=`. Practically every Go source
  file contains them, so **the pre-filter will rarely skip an `sg` spawn for Go** while that rule
  is in the ruleset at `error` severity. The saving for Go is small; the saving for the other
  three covered languages is larger; and the saving for the no-config case (A2) and the eleven
  uncovered languages (A3) is total. This ordering is why A2 and A3 lead the plan and A1 follows.
- The pre-filter never fires when the rule set is user-supplied and unusual, because such a rule
  set is more likely to hit the underivable path. That is the intended failure direction.

---

## §D Exclusions

### Out of Scope — the post-write ast-grep scan

- `postToolHandler.runAstScan` (via `internal/astgrep`) runs a **second** ast-grep pass after the
  write, and carries the same no-config and uncovered-language waste this SPEC removes from the
  pre-write path. It is deliberately untouched here: it is a different package, a different
  analyzer, and an advisory-only surface with no deny to protect. It is a candidate for a
  follow-up card, not a scope extension.

### Out of Scope — changing the rule set

- No ast-grep rule is added, removed, retuned, or re-severitied by this SPEC. The 26 shipped
  rules and their 14 `error` severities are inputs, not subjects. Expanding language coverage
  beyond the current four is separate work.
- No guardian vulnerability class is added or removed.

### Out of Scope — the deny policy itself

- Whether an `error`-severity ast-grep finding *should* block a write at all is not re-opened.
  Operator decision A settled it: the capability is kept.
- The `MOAI_SECURITY_BLOCKING` opt-in on the Stop-layer guardian is untouched.

### Out of Scope — worktree rule-config resolution

- Making `FindRulesConfig` resolve the *primary checkout's* ruleset from inside a card worktree
  would restore rule loading in worktree sessions. That would **increase** scan coverage and is a
  plausible improvement, but it changes which payloads are denied — the opposite of REQ-SSS-001's
  guarantee — and it belongs to its own card with its own risk argument.

### Out of Scope — latency targets

- This SPEC states no latency figure and sets no latency target. The absolute numbers available
  (the audit's ~164 ms and ~260-340 ms) were measured at load 8-10 and the investigation was
  written at load 38; they are usable as ratios only. Cost is expressed here as a **countable**:
  the number of temporary files created and the number of `sg` processes spawned.

### Out of Scope — removing the `moai hook security-scan` subcommand

- The subcommand stays (REQ-SSS-016). Removing it is a separate deprecation with its own
  compatibility window.

---

## §E Traceability

| Requirement | Item | Primary surface |
|---|---|---|
| REQ-SSS-001..002 | invariant | `internal/hook/pre_tool.go`, differential test |
| REQ-SSS-003..004 | A2 | `internal/hook/pre_tool.go`, `internal/hook/security/rules.go` |
| REQ-SSS-005..007 | A3 | `internal/hook/security/rules.go` (config-derived language set) |
| REQ-SSS-008..012 | A1 | new `internal/hook/security/prefilter.go` |
| REQ-SSS-013..017 | B | `internal/hook/post_tool.go`, `internal/cli/deps.go`, `settings.json(.tmpl)` |
| REQ-SSS-018..019 | distribution | `internal/template/templates/**` |
| REQ-SSS-020 | disclosure | pull request title + body |
