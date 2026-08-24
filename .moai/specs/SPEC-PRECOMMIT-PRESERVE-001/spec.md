---
id: SPEC-PRECOMMIT-PRESERVE-001
title: "Pre-commit hook install must never silently discard a local patch"
version: "0.1.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec (card t230)
priority: P2
tier: M
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hooks, cli, pre-commit, data-loss, provenance, install-safety"
---

# SPEC-PRECOMMIT-PRESERVE-001 — Pre-commit hook install must never silently discard a local patch

> GEARS notation (canonical since v3.0.0). Tier M. Card: t230.

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-24 | 0.1.0 | Initial plan-phase authoring from card t230. Requirement C closed as non-applicable by measurement (§A.4). |

## §A Overview

### §A.1 Problem

`moai update` and `moai init` overwrite an installed pre-commit hook whenever the hook carries the
MoAI marker, without ever asking whether the user has since modified it — and they report the
overwrite as a plain success.

`InstallPreCommitHook` (`internal/cli/hook_install_precommit.go:126-156`) does exactly two things
with an existing file: it checks the first three lines for the marker
`# MoAI-ADK pre-commit hook` (`:139-147`), then calls `os.WriteFile` (`:151`). There is no byte
comparison, no backup, and no diagnostic. The question it asks is "did MoAI write this?"; the
question it never asks is "has the user changed it since?".

The disguise is what makes this expensive. `installPreCommitHookOptional` (`:166-180`) prints the
same line — `  Pre-commit hook installed (.git/hooks/pre-commit)` — whether it created a hook from
nothing or destroyed a locally-patched one. The user reads a success message at the moment work is
lost.

Downstream evidence (the `mo.ai.kr` checkout, reported on card t230): an installed hook of 17,901 B
carrying three separate local patches was replaced by the distributed content in a single
`moai update`, with no output indicating anything had been removed.

### §A.2 Measured baseline (this tree, `294b4b6ab`)

| Measurement | Command | Value today |
|---|---|---|
| Backup logic present | `grep -c 'pre-commit\.bak' internal/cli/hook_install_precommit.go` | `0` |
| Provenance digest present | `grep -c 'sha256' internal/cli/hook_install_precommit.go` | `0` |
| Local extension point present | `grep -c 'pre-commit\.local' internal/template/templates/.git_hooks/pre-commit` | `0` |
| Distributed hook size | `wc -c < internal/template/templates/.git_hooks/pre-commit` | `3245` |
| Installer size | `wc -l < internal/cli/hook_install_precommit.go` | `179` |
| Call sites | `grep -n 'installPreCommitHookOptional' internal/cli/update_template_sync.go internal/cli/init.go` | `update_template_sync.go:575`, `init.go:898` |

The card cited call sites `:557` / `:773` and a distributed size of 6,454 B. The functions named are
correct; the line numbers have drifted, and the 6,454 B figure belongs to the downstream checkout's
copy, not to this tree's template (3,245 B). Neither discrepancy changes the defect.

### §A.3 The design crux — byte-difference alone cannot attribute a change

Any design that treats "installed bytes ≠ bytes about to be written" as "the user patched it" is
wrong, because a routine upstream version bump produces the same signal. Such a design would emit a
backup file and a warning on every release that touches the hook, for every user — noise that
trains people to ignore the one notice that matters.

The tool therefore needs to distinguish two questions:

- *Is this file different from what I am about to write?* — a byte comparison answers this, and it
  conflates the two causes.
- *Is this file different from what I last wrote?* — this is the attribution question, and it needs
  a record of what was last written.

§A.5 records the chosen answer.

### §A.4 Requirement C — closed by measurement, not implemented

Card t230's requirement C (narrowed by the 2026-08-24 sweep to the `exec` redirection axis) was
conditional: the downstream repair applied to an `exec` redirection swallowing stderr, and the
requirement asked whether the distributed template carried the same pattern.

Measured in this tree at `294b4b6ab`:

```
grep -rEn 'exec [0-9][0-9]*[<>]' --include='*.sh' --include='*.tmpl' --include='pre-commit' \
  --include='pre-push' --include='*.go' --include='*.md' . | grep -v '^\./\.git/'
→ 0 hits
```

The condition does not hold. There is no upstream occurrence to sweep, so this SPEC authors no
requirement against it. Recorded as a measured closure so a later reader does not re-open it.

### §A.5 Decision — what happens when the installed hook differs

Two decisions were required. Both are stated with their rejected alternatives so a reader can
disagree with either independently.

**Decision 1 — attribution mechanism: a provenance sidecar, not an in-body stamp.**

The installer records `sha256(preCommitHookContent)` into `.git/hooks/.moai-pre-commit.sha256`
immediately after every successful write. On the next install it re-hashes the installed file and
compares against that record:

| Sidecar state | Digest comparison | Meaning | Action |
|---|---|---|---|
| present | matches | untouched since MoAI wrote it, whatever version | overwrite silently |
| present | differs | user modified it after MoAI wrote it | back up, notify, overwrite |
| absent (legacy install) | — | undecidable | fall back to comparing against `preCommitHookContent`; identical → silent, differing → treat as modified |

A routine version bump lands in row 1: the installed file still matches the *previous* constant,
which is what the sidecar records, so it is silently replaced. This is the property the design
exists for.

*Rejected — an in-body version stamp* (a `# moai-hook-version:` line inside the hook). It works, but
it forces every attribution change through the `preCommitHookContent` / template pair (§C.1),
colliding with card t237, and it is defeated by a user who edits the hook and regenerates the stamp.
The sidecar lives outside the file being edited, which is precisely why editing the file is
detectable.

*Known limitation of the sidecar:* it lives in `.git/`, so it is per-clone and not distributed. A
fresh clone of a repository whose hook was patched by hand has no sidecar and lands in the legacy
row — correctly, since a fresh clone has no `.git/hooks/pre-commit` either.

**Decision 2 — on a detected difference: back up and overwrite, never preserve.**

| Option | Consequence |
|---|---|
| (a) back up, then overwrite — **chosen** | the patch is recoverable from `pre-commit.bak.<ts>`; the project converges on the distributed hook; the user is told, once, per divergence |
| (b) back up, then refuse to overwrite | the project silently freezes on an old hook and stops receiving upstream fixes — the same silent-loss failure, pointing the other way, and harder to notice because nothing is ever printed again |
| (c) prompt interactively | `moai update` runs non-interactively in CI and in scripted flows; a prompt there is a hang, not a safeguard |

Option (b) is the tempting one and is rejected deliberately: "never overwrite" is not the card's
goal. The card's goal is that overwriting is never silent. (a) delivers that while keeping the tool
converging.

The notice that accompanies a backup points the user at `.git/hooks/pre-commit.local` (REQ-PCP-011)
as the durable home for the checks they had patched in — that extension point is what makes (a)
survivable rather than merely recoverable.

## §B Requirements (GEARS)

### Attribution and provenance

- **REQ-PCP-001** — The pre-commit installer shall record the SHA-256 digest of the content it
  wrote into `.git/hooks/.moai-pre-commit.sha256` after every successful hook write.
- **REQ-PCP-002** — When the installed hook carries the MoAI marker and its digest matches the
  recorded provenance digest, the installer shall overwrite it and shall produce no output beyond
  the existing installation line (behaviour-preserving).
- **REQ-PCP-005** — Where no provenance record exists (a hook installed by a version predating this
  SPEC), the installer shall compare the installed bytes against `preCommitHookContent` and shall
  treat any difference as user-modified.

### Backup and disclosure

- **REQ-PCP-003** — When the installed hook is determined to be user-modified, the installer shall
  copy it to `.git/hooks/pre-commit.bak.<UTC-timestamp>` before writing the new content.
- **REQ-PCP-004** — When a backup is written, the installer shall emit a notice on stderr naming the
  backup path, stating that the hook was replaced, and naming `.git/hooks/pre-commit.local` as the
  place to re-apply local checks.
- **REQ-PCP-006** — The installer shall not overwrite a hook it has determined to be user-modified
  without producing both a backup file and a notice.
- **REQ-PCP-007** — The installer shall not report the unqualified success line as its only output
  on a run in which a backup was taken.
- **REQ-PCP-009** — The installer shall not overwrite an existing backup file; when the chosen
  backup path is occupied, it shall select a distinct path.

### Preserved behaviour

- **REQ-PCP-008** — When the installed hook does not carry the MoAI marker, the installer shall
  return `ErrUserHookExists` without writing a backup and without modifying the hook, exactly as
  today.
- **REQ-PCP-010** — Where a backup or provenance write fails, the installer shall report a warning
  and shall not fail `moai init` or `moai update`.
- **REQ-PCP-013** — The installer's `preCommitHookContent` constant and
  `internal/template/templates/.git_hooks/pre-commit` shall remain byte-identical.

### Local extension point

- **REQ-PCP-011** — Where `.git/hooks/pre-commit.local` exists and is executable, the pre-commit
  hook shall execute it as its final step and shall exit with that script's status.
- **REQ-PCP-012** — Where `.git/hooks/pre-commit.local` is absent or not executable, the pre-commit
  hook shall behave as it does today.

## §C Constraints

### §C.1 Paired-edit obligation

`preCommitHookContent` (`internal/cli/hook_install_precommit.go:26`) and
`internal/template/templates/.git_hooks/pre-commit` are enforced byte-identical by
`TestPreCommitTemplateMatchesConstant` (`internal/cli/hook_install_precommit_test.go:38`, verified
present at `294b4b6ab`). REQ-PCP-011 changes the hook body and therefore requires both edits in the
same commit, followed by `make build`.

REQ-PCP-001 through REQ-PCP-010 touch installer logic only and leave the hook body untouched — a
deliberate consequence of choosing the sidecar over an in-body stamp (§A.5).

### §C.2 Collision with card t237

Card t237 (issue #1641) edits the same constant/template pair on the module-root `go vet` axis.
Only REQ-PCP-011 collides. Run-phase should land REQ-PCP-011 as its own commit, last, so a rebase
against t237 touches one commit rather than the whole SPEC.

### §C.3 Non-interactive

`moai update` and `moai init` run unattended. No requirement here may introduce a prompt.

## §D Out of Scope

### Out of Scope — the `exec` redirection sweep (card requirement C)

- Not authored: the repository contains zero `exec <n><redirect>` occurrences (§A.4). There is no
  pattern here to fix.

### Out of Scope — `--no-hooks` over-blocking

- Confirmed but not fixed here: the single `--no-hooks` flag gates both
  `installPrePushHookOptional` and `installPreCommitHookOptional`
  (`update_template_sync.go:572`/`:575`, `init.go:895`/`:898`), so a user cannot skip one tier
  without skipping the other. Separate concern from silent loss; belongs to its own card.

### Out of Scope — the pre-push installer

- `InstallPrePushHook` shares the marker-based design and presumably the same defect. Fixing it is a
  sibling change, deliberately excluded to keep this SPEC's diff attributable.

### Out of Scope — restoring the three downstream patches

- The go.mod-detection axis belongs to card t237/#1641 and the gate-serialization axis to card
  t235/#1639. This SPEC makes such losses visible and recoverable; it does not re-implement them.

### Out of Scope — implementation shape

- Function decomposition, helper naming, and test file organization are run-phase decisions.

## §E Traceability

| REQ | AC |
|---|---|
| REQ-PCP-001 | AC-PCP-001 |
| REQ-PCP-002 | AC-PCP-002 |
| REQ-PCP-003 | AC-PCP-003 |
| REQ-PCP-004 | AC-PCP-004 |
| REQ-PCP-005 | AC-PCP-005 |
| REQ-PCP-006 | AC-PCP-006 |
| REQ-PCP-007 | AC-PCP-007 |
| REQ-PCP-008 | AC-PCP-008 |
| REQ-PCP-009 | AC-PCP-009 |
| REQ-PCP-010 | AC-PCP-010 |
| REQ-PCP-011 | AC-PCP-011 |
| REQ-PCP-012 | AC-PCP-012 |
| REQ-PCP-013 | AC-PCP-013 |
