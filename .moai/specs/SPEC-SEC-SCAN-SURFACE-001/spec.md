---
id: SPEC-SEC-SCAN-SURFACE-001
title: Security scan surface — cheap pre-write ast-grep gate + PostToolUse guardian merge
version: 0.2.0
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/hook
lifecycle: spec-anchored
tags: "security, hook, pretooluse, posttooluse, ast-grep, performance"
tier: M
---

## HISTORY

- 2026-08-24 — v0.2.0 — Audit iteration 1 (FAIL 0.65) remediation. `tags` corrected to the
  canonical comma-separated string (D1). §C.2 extraction rules extended to cover the shipped
  ruleset's dominant `kind:` + `regex:` shape and regex top-level alternation, with the
  language-by-language saving claim replaced by a measurement (D2). Requirement count reduced
  20 → 16 by consolidation, with no content dropped (D6). Short-circuit reachability (D7) and
  the empty-derived-set case (D8) adopted as requirement clauses.
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

### §A.1 Measurements taken while authoring

All on worktree `.claude/worktrees/t217`, tree `c4e90cd58`, branch `security-scan-surface`.

| Measurement | Command | Observed |
|---|---|---|
| Worktree has no ruleset | `ls .moai/config/astgrep-rules` | `No such file or directory` — Claim 3 holds on this tree |
| `sg` is installed | `command -v sg` | `/opt/homebrew/bin/sg` — the spawn is real, not a missing-binary no-op |
| Shipped rule languages | `grep -rh "^language:" internal/template/templates/.moai/config/astgrep-rules/` | `20 go, 2 javascript, 2 python, 2 typescript` |
| Error-severity rules per language | per-document YAML split over the shipped ruleset | go 8, javascript 2, python 2, typescript 2 |
| Advisory merge already exists | read of `internal/hook/registry.go:180` `mergeHandlerOutput` | `systemMessage` and `hookSpecificOutput.additionalContext` are each accumulated with a `"\n"` join; guarded by `TestDispatch_MergeAccumulatesAdditionalContext` |
| Pre-write config resolutions today | `grep -rn "FindRulesConfig" internal/` (non-test) | exactly **one** on the pre-write path, at `scanner.go:84` inside `ScanFile` |
| Single-file scan has no injection seam | `grep -n "scanFunc" internal/hook/security/ast_grep.go` | referenced only at `:199-200`, inside `ScanMultiple`; single-file `Scan` execs `sg` directly at `:137` |
| Handler scanner field is concrete | `grep -n "scanner \*security.SecurityScanner" internal/hook/pre_tool.go` | `pre_tool.go:325` — a struct pointer, so no stub can be injected at the handler level without a type change |
| Pre-filter skip rate | `python3 .moai/reports/t217/skiprate.py .` | `go files=2438 wouldSKIP=30 rate=1.2%` · `js files=81 wouldSKIP=78 rate=96.3%` · `py files=14 wouldSKIP=13 rate=92.9%` |

### §A.2 Two premises of the card were found to be inaccurate

Both are recorded here so the plan does not inherit them.

1. **"Both handlers write `hookSpecificOutput` JSON to stdout, so the `additionalContext`
   payloads must be combined."** The collision is not `additionalContext` against
   `additionalContext`. The post-tool handler emits its LSP / MX / ast-grep / navigator advisories
   on the **top-level `systemMessage`** field and leaves `hookSpecificOutput.additionalContext`
   empty; the guardian emits `hookSpecificOutput.additionalContext` and no `systemMessage`. The
   regression risk the card names is real but sits one field over: a naive merge that folds the
   guardian text into `systemMessage`, or that replaces the whole `HookSpecificOutput` struct,
   drops one of the two. REQ-SSS-012 is therefore stated over **both** carrier fields.
2. **"`moai hook security-scan` may have other callers — verify."** Verified: the only references
   are `internal/cli/hook.go` (the subcommand registration), its own tests, the wrapper script
   `handle-security-scan.sh` (+ `.sh.tmpl`), and the `settings.json` (+ `.json.tmpl`) entry this
   SPEC removes. There is no third-party production caller in-tree. Retaining the subcommand is a
   backward-compatibility choice for user projects whose `settings.json` predates the merge — a
   real reason, but not the "it is already called elsewhere" reason the card assumed.

### §A.3 Latent risk accepted with evidence — handler short-circuit

`internal/hook/registry.go:142` short-circuits `Dispatch` when a preceding handler returns a
block decision, and again on `ExitCode == 2`. Today the guardian is immune because Claude Code
invokes it as a separate process. After the merge it becomes order-dependent. The risk is
**latent, not live**: `post_tool.go:144` documents the post-tool handler as observation-only and
it returns `allow` on every path. REQ-SSS-012 nonetheless states the reachability invariant, so a
future handler that does block cannot silently mute the guardian.

---

## §B Requirements

Notation: GEARS. `<subject>` names the component, not "the system".

### B.1 The invariant that bounds every other requirement

- **REQ-SSS-001** (Ubiquitous) — The pre-write security gate shall compute an unchanged deny
  verdict: for a given rule set and project root, it shall deny every payload the
  pre-implementation gate denies, and shall introduce no deny the pre-implementation gate did not
  produce. Only the work done to reach the verdict changes.

### B.2 Item A2 — no rules config, no scan

- **REQ-SSS-002** (Where / When) — **Where** no ast-grep rules configuration resolves for the
  project root, **when** a Write payload reaches the pre-write security gate, the gate shall
  return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-003** (Ubiquitous) — The gate shall resolve the rules configuration in the caller,
  exactly once per invocation, and shall pass the resolved path to the scanner; the scanner shall
  perform no second resolution on the pre-write path.

### B.3 Item A3 — no rules for this language, no scan

- **REQ-SSS-004** (Where / When) — **Where** the resolved rules configuration declares no rule
  whose `language` matches the payload file's language, **when** a Write payload reaches the
  gate, the gate shall return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-005** (Ubiquitous) — The covered-language set shall be derived by reading the
  resolved configuration and the rule files it names. The gate shall not carry a second
  hardcoded language list.
- **REQ-SSS-006** (When) — **When** the resolved configuration cannot be read, cannot be parsed,
  cannot be walked, **or yields an empty covered-language set**, the gate shall treat the
  covered-language set as unknown and shall escalate to `sg` rather than skipping. An unreadable
  or empty result is never evidence that a language is uncovered.

### B.4 Item A1 — literal pre-filter derived from the rules themselves

- **REQ-SSS-007** (Where / When) — **Where** a literal pre-filter has been derived for the
  payload's language, **when** the payload content contains none of the derived literal tokens,
  the gate shall return allow without creating a temporary file and without spawning `sg`.
- **REQ-SSS-008** (Ubiquitous) — The pre-filter shall be a pure function of the resolved rule
  set, derived only from those rules of the resolved configuration that carry `error` severity
  and apply to the payload's language. Rules at `warning`, `info`, or `hint` severity shall not
  contribute tokens, because they cannot produce a deny. The pre-filter shall not be derived from
  `GuardianPatterns()` or from any other table that is not the resolved ast-grep rule set.
- **REQ-SSS-009** (Where) — **Where** any applicable `error`-severity rule yields no token that
  is provably mandatory for a match, the pre-filter shall be marked underivable for that
  language, and the gate shall escalate to `sg` unconditionally for that language.

### B.5 Item B — merge the PostToolUse guardian into the post-tool handler

- **REQ-SSS-010** (Ubiquitous) — The regex guardian buffer scan shall run within the post-tool
  handler's process for `Write` / `Edit` / `MultiEdit` events.
- **REQ-SSS-011** (Ubiquitous) — The guardian's advisory shall remain advisory. The merge shall
  introduce no deny, no `decision: block`, and no non-zero exit on the PostToolUse path.
- **REQ-SSS-012** (When) — **When** both the post-tool handler's own advisory text and the
  guardian's advisory text are non-empty for a single event, the emitted hook output shall carry
  both: neither carrier field (`systemMessage`,
  `hookSpecificOutput.additionalContext`) shall be overwritten by the other's content, and
  neither advisory shall be dropped. No preceding handler's decision shall prevent the guardian
  scan from being reached (§A.3).
- **REQ-SSS-013** (Ubiquitous) — The `moai hook security-scan` subcommand shall remain
  registered and invocable with its current stdin / stdout contract, so a user project whose
  `settings.json` still names it keeps working.
- **REQ-SSS-014** (Ubiquitous) — The `handle-security-scan.sh` PostToolUse entry shall be absent
  from the `Write|Edit|MultiEdit` matcher in both `.claude/settings.json` and
  `internal/template/templates/.claude/settings.json.tmpl`.

### B.6 Distribution and disclosure

- **REQ-SSS-015** (Ubiquitous) — Every file this SPEC changes under `.claude/` shall have its
  mirror under `internal/template/templates/.claude/` changed in the same commit; any hook
  wrapper this SPEC changes shall move as a `.sh` / `.sh.tmpl` pair; and no changed file under
  `internal/template/templates/` shall introduce a SPEC ID, a REQ token, a date, a commit SHA, an
  absolute macOS path, or a `CLAUDE.local` reference.
- **REQ-SSS-016** (Ubiquitous) — The pull request for this SPEC shall declare in its **title**
  and in the **first sentence of its body** that this is a change to the security scan surface,
  and shall not present itself as a performance improvement. The cost reduction is the method;
  the security surface is the subject.

---

## §C Design decision — where the pre-filter comes from

The card offered two candidate pre-filters. This section records the choice, its extraction
rules, and its measured value, so the run phase does not re-litigate it.

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

The reuse is additionally **fragile against drift**: it couples the deny reachability of the
ast-grep rule set to an unrelated table that the guardian SPEC is free to edit. A future guardian
edit could silently narrow the write-path deny with no signal.

### C.2 Chosen — derive the pre-filter from the resolved `error`-severity rules

The pre-filter is a set of **literal substrings** extracted from the rules that can actually
produce a deny, checked against the payload with `strings.Contains`. No second regex table is
introduced; no regex engine is needed for the pre-pass, which makes it cheaper than either
candidate the card sketched.

A token is admitted only when it is **mandatory** for the rule to match. The extraction rules,
each with the reason it is sound:

| Rule shape | Extraction | Why the token stays mandatory |
|---|---|---|
| `pattern:` | Literal runs outside any metavariable (`$X`, `$$$X`) | A pattern match reproduces its literal text |
| `all:` | Union of the children's tokens | Every child must match, so any child's mandatory token is mandatory |
| `any:` | Union — **only if every branch yields ≥ 1 token**; otherwise underivable | A tokenless branch can match with none of the other branches' tokens present |
| `regex:` with a top-level alternation group | Union of each branch's mandatory literal prefix; underivable if any branch has none | Same logic as `any:` — the alternation *is* a disjunction |
| `regex:` without alternation | The mandatory literal prefix | The prefix must appear in any match |
| `kind:` **and** `regex:` together | Tokens from the `regex:` conjunct | `kind` narrows and never widens the match set, so the regex's tokens remain mandatory. This is an `all:`-shaped conjunction |
| `kind:` alone, `inside:`, `has:`, `follows:`, a `regex:` with no literal anchor, a pattern that is entirely metavariables | **Underivable** — marks the whole language underivable | Unknown is never treated as absent |

Underivability is per-language and total: one underivable rule forces unconditional escalation
for its language (REQ-SSS-009).

Worked example, the shipped ruleset's dominant shape. `sec-hardcoded-credential` carries no
`pattern:` — it is `kind:` + `regex:` in all four covered languages:

```yaml
rule:
  kind: interpreted_string_literal          # go; js/ts/python use kind: string
  regex: "^\"(sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})"
```

Under the table above this is derivable: the `kind:` conjunct is ignored, the `regex:` has a
top-level alternation of five branches, and every branch carries a mandatory literal prefix, so
the token set is `{sk-, AKIA, ghp_, xox, AIza}`. Without the alternation row this rule would be
read as `kind:`-only and would mark **all four covered languages** underivable, reducing item A1
to a no-op everywhere. That reading was the defect the plan audit caught, and the table above is
the resolution.

### C.3 Measured value — what the pre-filter actually skips

The v0.1.0 draft asserted that "the saving for the other three covered languages is larger" than
for Go. That was an unverified premise. It is now measured.

Token sets derived by the §C.2 rules from the shipped ruleset's error-severity rules:

- **go** (8 error rules) — `go-error-ignored-blank`'s pattern `$_, $ERR = $FUNC($$$ARGS)`
  contributes `,` and `=`. Adding the other seven rules' tokens can only lower the skip rate, so
  these two bound it.
- **javascript / typescript** (2 error rules) — `child_process.exec`, `cp.exec`, plus the five
  credential prefixes.
- **python** (2 error rules) — `subprocess.call`, `os.system`, plus the five credential prefixes.

Command: `python3 .moai/reports/t217/skiprate.py .` (script committed alongside this SPEC),
worktree `c4e90cd58`. Observed:

```
go files=2438 wouldSKIP=30 rate=1.2%
js files=81  wouldSKIP=78 rate=96.3%
py files=14  wouldSKIP=13 rate=92.9%
```

**Reading of the measurement.** The saving is real for javascript, typescript, and python, and
is close to nil for Go. The asymmetry has a cause worth stating plainly: Go carries 8 of the 12
error rules, and one of them (`go-error-ignored-blank`) matches a construct present in almost
every Go file. So A1 saves most in the language where the gate has least to check, and saves
nothing in the language where it checks most.

**Why A1 is kept rather than cut.** MoAI-ADK ships to sixteen languages on equal terms
(`CLAUDE.local.md` §15). Deciding A1's fate from the 1.2% figure would be deciding it from the
fact that this particular repository is written in Go, and would deny a 93-96% saving to every
JavaScript, TypeScript, and Python project the tool ships to. The mechanism is data-driven, so
the rate self-adjusts as a language's ruleset grows. Item A1 is therefore retained, with its Go
degeneracy stated as a measured fact rather than hidden.

**Gaps in this measurement.** It counts *files present in this repository*, not Writes — the
distribution of what a session actually writes is not observed. The javascript (81) and python
(14) populations are small and skewed toward tooling scripts. A project whose source genuinely
uses `child_process.exec` or `os.system` throughout would skip far less. The figure is a
directional measurement, not a guarantee.

**Residual risk.** The pre-filter's soundness rests on the extractor, which is the most
defect-prone unit in this SPEC and guards a deny. Three things bound it: the underivable ⇒
escalate fallback (REQ-SSS-009), the M0 differential corpus (AC-SSS-001), and the second
assertion of AC-SSS-001 — that for every fixture that denies, the pre-filter would not have
skipped it.

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
  rules and their severities are inputs, not subjects. In particular, whether
  `go-error-ignored-blank` — an error-handling rule, not a security rule — belongs at `error`
  severity in a gate that *denies writes* is a fair question raised by §C.3's measurement, and it
  belongs to the ruleset's owner, not to this card.
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
  the number of temporary files created and the number of scans dispatched.

### Out of Scope — pre-existing wrapper-pair drift

- `.claude/hooks/moai/handle-pre-tool.sh` differs from its `.sh.tmpl` today, in comments only
  (the deployed copy carries a SPEC ID the neutralized template must not, and lacks a
  documentation block the template has). Byte-equality across the `.sh` / `.sh.tmpl` axis is
  therefore **not** a valid repository-wide invariant, and this SPEC does not attempt to restore
  it. This SPEC touches no hook wrapper; its pair check is scoped to the files its own diff
  changes (REQ-SSS-015).

### Out of Scope — removing the `moai hook security-scan` subcommand

- The subcommand stays (REQ-SSS-013). Removing it is a separate deprecation with its own
  compatibility window.

---

## §E Traceability

| Requirement | Item | Primary surface | Closed by |
|---|---|---|---|
| REQ-SSS-001 | invariant | `internal/hook/pre_tool.go` | AC-SSS-001 |
| REQ-SSS-002 | A2 | `internal/hook/pre_tool.go` | AC-SSS-002, AC-SSS-003 |
| REQ-SSS-003 | A2 | `internal/hook/security/scanner.go` | AC-SSS-004 |
| REQ-SSS-004 | A3 | `internal/hook/pre_tool.go` | AC-SSS-005 |
| REQ-SSS-005 | A3 | `internal/hook/security/rules.go` | AC-SSS-006 |
| REQ-SSS-006 | A3 fail-open | `internal/hook/security/rules.go` | AC-SSS-007 |
| REQ-SSS-007 | A1 | `internal/hook/security/prefilter.go` | AC-SSS-011 |
| REQ-SSS-008 | A1 | `internal/hook/security/prefilter.go` | AC-SSS-008 |
| REQ-SSS-009 | A1 fail-open | `internal/hook/security/prefilter.go` | AC-SSS-009, AC-SSS-010 |
| REQ-SSS-010 | B | `internal/hook/post_tool.go`, `internal/cli/deps.go` | AC-SSS-013 |
| REQ-SSS-011 | B | `internal/hook/post_tool.go` | AC-SSS-014 |
| REQ-SSS-012 | B | `internal/hook/registry.go` consumers | AC-SSS-012 |
| REQ-SSS-013 | B | `internal/cli/hook.go` | AC-SSS-015 |
| REQ-SSS-014 | B | `settings.json`, `settings.json.tmpl` | AC-SSS-015 |
| REQ-SSS-015 | distribution | `internal/template/templates/**` | AC-SSS-016 |
| REQ-SSS-016 | disclosure | pull request title + body | AC-SSS-016 |
