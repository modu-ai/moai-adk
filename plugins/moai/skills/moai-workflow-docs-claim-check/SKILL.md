---
name: moai-workflow-docs-claim-check
description: >
  Read-only check of whether the claims a public-facing document makes
  (README, release notes, install and usage guides) are supported by
  user-supplied evidence. Runs Preflight, Claim Triage, and Validation, splits
  composite claims into atomic ones, and labels each. Runs no commands and
  writes no fixes.

when_to_use: >
  Use when asked whether documentation claims are backed by evidence, to audit
  a README or release note before publishing, or to find unsupported or
  outdated statements in user-facing docs.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-07-24"
  modularized: "true"
  tags: "documentation, claim-check, evidence, verification, readme, release-notes, read-only"
  related-skills: "moai-workflow-project, moai-foundation-quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# Documentation Claim Check

Assess whether the statements a public-facing document makes are actually
supported by the evidence supplied alongside it. The assessment reads, judges,
and reports. It changes nothing and runs nothing.

Typical subjects: a README, a release note, an install guide, a quickstart, a
migration note, a feature or compatibility table.

## Hard boundaries

Three boundaries are absolute. They hold even when the request asks for more;
in that case perform the assessment and decline the rest in Boundary Notes.

1. **No command execution.** Do not run builds, tests, package managers,
   linters, network requests, or any shell command as part of the assessment.
   When a claim can only be settled by running something, **name** the exact
   command and the file it should run against and label the claim
   `needs-human`. Naming the command is the deliverable; running it is not.
2. **No fixes.** Do not produce patches, diffs, rewritten passages, or file
   edits. Describe what a maintainer would change and where, then stop.
3. **No code review and no security review.** Do not assess code quality,
   architecture, performance, or vulnerabilities. If asked, state the boundary
   in Boundary Notes and continue with the claim assessment only.

Opening the files the user pointed at, to locate the evidence they supplied, is
in scope — that is reading, not executing.

## Phase 1 — Preflight

Complete all three steps before triaging a single claim. If a step cannot be
completed, report that and stop rather than guessing.

1. **Confirm the document is public-facing.** This skill judges documents
   written for users of the software: README, release notes, install and usage
   guides, published site pages. Internal design notes, task trackers, and
   private runbooks are out of scope — say so and stop.
2. **Inventory the supplied evidence.** For every item record what it is, where
   it came from, its **version** identifier, and its **timestamp**. An item with
   neither is still usable, but record it as undated: it cannot later support a
   freshness judgment.
3. **Flag secrets for redaction before proceeding.** Scan the supplied evidence
   for credentials, tokens, private keys, connection strings, and personal
   data. On a hit, flag the location for redaction, never reproduce the secret
   in any output, and continue only once a redacted copy is available.

## Phase 2 — Claim Triage

Turn the document into an inventory of atomic claims.

**Extract.** Walk the document and pull out every statement that asserts
something checkable about the software: supported platforms and versions,
install and usage steps, defaults, limits, guarantees, availability, counts.

**Decompose composite claims.** A composite claim bundles several
independently-checkable assertions into one sentence. Split it so that **each
atomic claim carries exactly one assertion and therefore receives exactly one
label**. A sentence that would otherwise need two labels is not yet atomic.

> "Installs with a single command on macOS and Linux" splits into three atomic
> claims: single-command install, macOS support, Linux support. Each is
> evidenced — and can fail — independently.

**Set aside the subjective.** Statements of taste or ambition ("fast",
"developer-friendly", "production-grade") cannot be checked against evidence.
Exclude them from labeling and list them under Input Scope Reviewed with a
one-line reason. When a subjective adjective wraps a checkable core, split it:
label the core, exclude the adjective.

**Bind evidence.** For each atomic claim, note which supplied evidence items
bear on it — or record that none does.

## Phase 3 — Validation

Assign **exactly one** label to every atomic claim by walking the ordered
decision tree. Stop at the first gate that fires; do not re-open an earlier gate.

```
needs-human  ->  stale-suspected  ->  verified  ->  unsupported
```

| Label | Gate condition |
|-------|----------------|
| `needs-human` | Settling the claim requires something outside this skill: running a command, reaching a private system, exercising a UI, or a call only a maintainer can make. |
| `stale-suspected` | Evidence indicates the claim was true earlier, but current evidence disagrees on a version, date, count, or name. A temporal mismatch, not a contradiction of substance. |
| `verified` | Supplied evidence directly supports the claim and the supporting item can be named. |
| `unsupported` | None of the above fired: the evidence does not carry the claim. |

`unsupported` always carries **exactly one** reason:

| Reason | Meaning |
|--------|---------|
| `missing-evidence` | No supplied evidence speaks to the claim at all. |
| `contradicted` | Supplied evidence asserts the opposite. |
| `insufficient-coverage` | Evidence is on-topic but narrower than the claim — one platform of three, one version of a declared range, one path of several. |

Anchoring rules: every `verified` names its evidence anchor; every
`stale-suspected` names the mismatched field and both values; every
`needs-human` names the command or file that would settle it; every
`unsupported` carries its reason.

Gate criteria in full, tie-breaks between adjacent gates, and a claim-type
table live in `references/label-decision-tree.md`. Worked end-to-end
assessments live in `references/worked-examples.md`.

## Output contract

Emit exactly these three sections, in this order, every time — including when
the claim inventory is empty.

### 1. Input Scope Reviewed

- Documents read, with version or date when known.
- Evidence items inventoried, each with version and timestamp (or `undated`).
- Statements excluded as subjective, each with a one-line reason.
- Evidence that was needed but **not** supplied, named specifically.

### 2. Claim Assessments

One row per atomic claim:

| # | Atomic claim | Label | Reason | Evidence anchor or what is missing |
|---|--------------|-------|--------|------------------------------------|

`Reason` is filled only for `unsupported` rows.

### 3. Boundary Notes

- The certification, stated literally: **no commands executed** during this
  assessment.
- Anything declined, with the boundary that forbade it (fix requests, code or
  security review requests).
- Residual risk: what this assessment still does not prove even where every
  row is `verified`.
- For each `needs-human` row, the specific command or file a maintainer needs.

## Relationship to the claim-integrity policy

This skill is the document-facing procedure for the project rule at
`.claude/rules/moai/core/verification-claim-integrity.md`. That rule owns the
norm and its reporting format; this skill only applies the norm to a published
document. Read the rule directly when the norm itself is in question — it is
not restated here.

## Bundled references

| File | Contents |
|------|----------|
| `references/label-decision-tree.md` | Full gate criteria, adjacent-gate tie-breaks, claim-type table, composite-splitting guidance |
| `references/worked-examples.md` | End-to-end assessments across several language ecosystems |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "One quick command settles this claim, so I will just run it" | The no-execution boundary has no size exemption. Name the command and label the claim `needs-human`. |
| "The claim is obviously true, I know this ecosystem" | Background knowledge is not supplied evidence. Absent an evidence anchor the label is `unsupported`, reason `missing-evidence`. |
| "No evidence contradicts it, so it passes" | Nothing arriving to contradict a claim is not support for it. Silence maps to `missing-evidence`, never to `verified`. |
| "The doc is wrong, I will just fix the sentence" | Findings only. Describe the change and where it belongs; the edit is a maintainer action. |
| "This sentence has two parts but one obvious verdict" | Two assertions need two labels. Split it before labeling. |
| "The Linux log is missing but the macOS log is close enough" | Narrower evidence than the claim is `insufficient-coverage`, not `verified`. |
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- A claim row carries two labels, or a label plus a hedge.
- A `verified` row names no evidence anchor.
- An `unsupported` row carries no reason, or more than one.
- A subjective adjective was labeled instead of excluded.
- Boundary Notes are missing, or omit the literal no-commands-executed line.
- A command was run "just to confirm" during the assessment.
- Zero evidence was supplied and some claim still came back `verified`.
- Evidence was inventoried without a version or timestamp and then used to
  judge freshness.
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] All three Preflight steps completed and recorded before any labeling.
- [ ] Every composite claim split until each atomic claim holds one assertion.
- [ ] Every atomic claim carries exactly one label from the four-value set.
- [ ] The decision tree was walked in order for each claim.
- [ ] Every `unsupported` row carries exactly one reason.
- [ ] All three output sections are present, in order.
- [ ] Boundary Notes state the literal no-commands-executed certification.
- [ ] No patch, diff, or edited passage appears anywhere in the output.
<!-- moai:evolvable-end -->
