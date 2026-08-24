---
id: SPEC-PRECOMMIT-PRESERVE-001
title: "Pre-commit hook install must never silently discard a local patch"
version: "0.2.0"
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
| 2026-08-24 | 0.2.0 | Retracted a false "stale line numbers" claim (§A.2); all citations now SHA-anchored. Three-way comparison raised to REQ-PCP-014; silence prohibition restated as a policy-independent invariant (REQ-PCP-006). |

## §A.0 Citation convention

[HARD] Every `file:line` reference in this SPEC names the tree it was measured against, as
`path:line @<sha>`. **The durable anchor is the symbol name; the line number is a
measured-at-a-SHA convenience and is expected to differ between trees.**

This convention exists because it was already violated once. Version 0.1.0 of this document asserted
that card t230's citation `init.go:773` was "stale" and had "drifted". It had not: `:773` is correct
on the primary checkout at `a1b1ca696`, and `:898` is correct in the card worktree at `294b4b6ab`.
Two correct measurements of two different trees were mistaken for one measurement gone stale. A
reader who finds a line number here that does not resolve should re-measure before concluding
anything drifted.

## §A Overview

### §A.1 Problem

`moai update` and `moai init` overwrite an installed pre-commit hook whenever the hook carries the
MoAI marker, without ever asking whether the user has since modified it — and they report the
overwrite as a plain success.

`InstallPreCommitHook` (`internal/cli/hook_install_precommit.go:126-156 @294b4b6ab`) does exactly
two things with an existing file: it checks the first three lines for the marker
`# MoAI-ADK pre-commit hook` (`:139-147 @294b4b6ab`), then calls `os.WriteFile`
(`:151 @294b4b6ab`). There is no byte comparison, no backup, and no diagnostic. The question it asks
is "did MoAI write this?"; the question it never asks is "has the user changed it since?".

The disguise is what makes this expensive. `installPreCommitHookOptional`
(`:166-180 @294b4b6ab`) prints the same line —
`  Pre-commit hook installed (.git/hooks/pre-commit)` — whether it created a hook from nothing or
destroyed a locally-patched one. The user reads a success message at the moment work is lost.

Downstream evidence (the `mo.ai.kr` checkout, reported on card t230): an installed hook of 17,901 B
carrying three separate local patches was replaced by the distributed content in a single
`moai update`, with no output indicating anything had been removed.

### §A.2 Measured baseline (card worktree, `294b4b6ab`)

| Measurement | Command | Value |
|---|---|---|
| Backup logic present | `grep -c 'pre-commit\.bak' internal/cli/hook_install_precommit.go` | `0` |
| Provenance record present | `grep -c 'sha256' internal/cli/hook_install_precommit.go` | `0` |
| Local extension point present | `grep -c 'pre-commit\.local' internal/template/templates/.git_hooks/pre-commit` | `0` |
| Distributed hook size | `wc -c < internal/template/templates/.git_hooks/pre-commit` | `3245` |
| Installer size | `wc -l < internal/cli/hook_install_precommit.go` | `179` |
| Call site (update) | `grep -n 'installPreCommitHookOptional' internal/cli/update_template_sync.go` | `575` |
| Call site (init) | `grep -n 'installPreCommitHookOptional' internal/cli/init.go` | `898` |

**Retraction.** Version 0.1.0 recorded that the card's call-site citations `:557` / `:773` were
stale. That claim is withdrawn: it compared a measurement of this worktree against a measurement of
the primary checkout and attributed the difference to time rather than to place. The card's numbers
are correct for the primary checkout at `a1b1ca696`; the numbers in the table above are correct for
this worktree at `294b4b6ab`. Nothing drifted. The symbol `installPreCommitHookOptional` resolves in
both.

The card's "distributed original 6,454 B" is likewise a figure from the downstream `mo.ai.kr`
checkout, not from this tree (3,245 B). Not a contradiction — a different tree. Neither figure is
load-bearing for the fix.

### §A.3 The design crux — two-way comparison cannot attribute a change

Any design that treats "installed bytes ≠ bytes about to be written" as "the user patched it" is
wrong, because a routine upstream version bump produces exactly that signal. Such a design emits a
backup file and a warning on every release that touches the hook, for every user — noise that trains
people to ignore the one notice that matters.

The comparison is two-way and has only two operands:

- **installed** — what is on disk now
- **incoming** — what is about to be written

Those two cannot separate "the user changed it" from "we changed it". A third operand is required:

- **recorded** — what this tool last wrote to that path

The attribution then follows from `installed` vs `recorded`, and the version question follows
separately from `recorded` vs `incoming`:

| `installed` vs `recorded` | Meaning | Consequence |
|---|---|---|
| equal | untouched since MoAI wrote it | any difference against `incoming` is a version bump — overwrite quietly |
| differs | the user edited it after MoAI wrote it | the loss-bearing case this card is about |

This is a requirement, not an implementation note — see REQ-PCP-014. The overwrite policy (§A.5
Decision 2) sits *on top of* this distinction, not instead of it.

### §A.4 Requirement C — closed by measurement, not implemented

Card t230's requirement C (narrowed by the 2026-08-24 sweep to the `exec` redirection axis) was
conditional: the downstream repair applied to an `exec` redirection swallowing stderr, and the
requirement asked whether the distributed template carried the same pattern.

Measured in this worktree at `294b4b6ab`:

```
grep -rEn 'exec [0-9][0-9]*[<>]' --include='*.sh' --include='*.tmpl' --include='pre-commit' \
  --include='pre-push' --include='*.go' --include='*.md' . | grep -v '^\./\.git/'
→ 0 hits
```

The condition does not hold. There is no upstream occurrence to sweep, so this SPEC authors no
requirement against it. Recorded as a measured closure so a later reader does not re-open it.
Per §A.0, this is a claim about **this tree at this SHA** — a different checkout may differ, and the
sweep is cheap to repeat.

### §A.5 Decisions

Two decisions. Both state their rejected alternatives so a reader can disagree with either
independently.

#### Decision 1 — how `recorded` is stored: a digest sidecar

The installer writes `sha256(preCommitHookContent)` to `.git/hooks/.moai-pre-commit.sha256`
immediately after every successful hook write. On the next install it hashes the installed file and
compares.

Four questions were put to this mechanism.

**Where does it live?** `.git/hooks/`, beside the hook it describes. `.git/` is per-clone, never
committed, and already the hook's own home, so the record shares the hook's lifetime exactly: wipe
`.git/hooks/`, and both the hook and its provenance go together, which is the correct coupling.

**What happens on first upgrade from a version that never wrote one?** This is the migration case,
and it is universal: *every* existing installation reaches the new code with no record. It is
handled by REQ-PCP-005 — with no record, the tool cannot attribute, so it falls back to comparing
`installed` against `incoming` and treats any difference as user-modified. That is deliberately the
noisy direction: an unmodified legacy hook that happens to differ from the new content gets one
backup and one notice it did not strictly need, once, and is stamped thereafter. The alternative —
treating "unknown" as "unmodified" — is silent data loss for exactly the population the card is
about, since a hand-patched legacy hook is the most likely thing to be found without a record.

**Can it drift or be deleted?** Yes: a user can delete it, and a hook restored from a backup by hand
will not match it. Both land in the same place — a mismatch, or an absence — and both are treated as
`user-modified`, which is the safe reading. The record can only cause a *missed* detection if it is
forged to match a hand-edited hook, which requires deliberately hashing one's own edit into it.

**Is there a simpler mechanism?** Three were considered.

| Alternative | Verdict |
|---|---|
| In-body version stamp (`# moai-hook-version:` inside the hook) | Rejected. Forces every attribution change through the constant/template pair (§C.1), colliding with card t237, and is defeated by a user who edits the hook and regenerates the stamp. The sidecar's whole virtue is living *outside* the file being edited. |
| Full-content snapshot of what was written, rather than a digest | Rejected for now, on cost rather than principle — see below. |
| Compare against every historically-shipped hook version | Rejected. Requires shipping an ever-growing corpus of past hook bodies, and still cannot recognise a version older than the corpus. |

On the **full-content snapshot**: it satisfies the three-way structure identically — equality against
a stored copy is the same test — and it buys one real thing a digest cannot, namely the ability to
show the user *what they changed*, by diffing the snapshot against the backup. That is genuinely
valuable for a card about "what did I just lose". It is deferred rather than dismissed, for two
reasons: it stores ~3 KB of duplicated hook body per clone, and a stale-but-present snapshot is
mistakable for a hook by tooling that globs `.git/hooks/*`. Decisively, the *decision logic is
identical either way* — the classifier's contract in §A.3 does not change — so swapping the digest
for a full snapshot later is a local change that re-opens no decision here. If run-phase finds the
diff worth having, take it; this SPEC does not require it.

#### Decision 2 — on a detected user modification: back up and overwrite

| Option | Consequence |
|---|---|
| (a) back up, then overwrite — **chosen** | the patch is recoverable from `pre-commit.bak.<ts>`; the project converges on the distributed hook; the user is told, once, per divergence |
| (b) back up, then refuse to overwrite | the project silently freezes on an old hook and stops receiving upstream fixes — the same silent-loss failure pointing the other way, and harder to notice because nothing is ever printed again |
| (c) prompt interactively | `moai update` runs unattended, in CI and in scripted flows; a prompt there is a hang, not a safeguard |

Option (b) is the tempting one and is rejected deliberately: "never overwrite" is not the card's
goal. The card's goal is that overwriting is never silent — and that goal is REQ-PCP-006, which
binds regardless of which option is in force, so a later reader who prefers (b) may take it without
weakening anything.

The notice accompanying a backup points at `.git/hooks/pre-commit.local` (REQ-PCP-011) as the
durable home for the checks the user had patched in — that extension point is what makes (a)
survivable rather than merely recoverable.

## §B Requirements (GEARS)

### The invariant

- **REQ-PCP-006** — The installer shall not replace a hook it has classified as user-modified
  without producing **both** a backup file and a notice naming it. This binds regardless of which
  overwrite policy (§A.5 Decision 2) is in force: a future change from overwrite to preserve, or to
  any conditional variant, does not license silence. No policy choice may trade this away.

### Attribution

- **REQ-PCP-014** — The installer shall classify an existing marker-bearing hook by comparing three
  operands — the installed file, the content this tool last recorded writing, and the content about
  to be written — and shall not attribute a difference to the user on the basis of a two-way
  comparison between the installed file and the incoming content alone.
- **REQ-PCP-001** — The installer shall record the SHA-256 digest of the content it wrote into
  `.git/hooks/.moai-pre-commit.sha256` after every successful hook write.
- **REQ-PCP-002** — When the installed hook carries the MoAI marker and its digest matches the
  recorded digest, the installer shall overwrite it and shall produce no output beyond the existing
  installation line, even when the incoming content differs (behaviour-preserving for version
  bumps).
- **REQ-PCP-005** — Where no provenance record exists, or the record is unreadable, the installer
  shall compare the installed bytes against the incoming content and shall treat any difference as
  user-modified.

### Disclosure

- **REQ-PCP-003** — When the installed hook is classified user-modified, the installer shall copy it
  to `.git/hooks/pre-commit.bak.<UTC-timestamp>` before writing the new content.
- **REQ-PCP-004** — When a backup is written, the installer shall emit a notice on stderr naming the
  backup path, stating that the hook was replaced, and naming `.git/hooks/pre-commit.local` as the
  place to re-apply local checks.
- **REQ-PCP-007** — The installer shall not report the unqualified success line as its only output
  on a run in which a backup was taken.
- **REQ-PCP-009** — The installer shall not overwrite an existing backup file; when the chosen
  backup path is occupied it shall select a distinct path and shall still produce a backup.

### Preserved behaviour

- **REQ-PCP-008** — When the installed hook does not carry the MoAI marker, the installer shall
  return `ErrUserHookExists` without writing a backup and without modifying the hook, exactly as
  today.
- **REQ-PCP-010** — Where a backup or provenance write fails, the installer shall report a warning
  and shall not fail `moai init` or `moai update`.
- **REQ-PCP-013** — `preCommitHookContent` and
  `internal/template/templates/.git_hooks/pre-commit` shall remain byte-identical.

### Local extension point

- **REQ-PCP-011** — Where `.git/hooks/pre-commit.local` exists and is executable, the pre-commit
  hook shall execute it as its final step and shall exit with that script's status.
- **REQ-PCP-012** — Where `.git/hooks/pre-commit.local` is absent or not executable, the pre-commit
  hook shall behave as it does today.

14 requirements — within the Tier M ceiling of 16.

## §C Constraints

### §C.1 Paired-edit obligation

`preCommitHookContent` (`internal/cli/hook_install_precommit.go:26 @294b4b6ab`, symbol anchor:
the `preCommitHookContent` const) and `internal/template/templates/.git_hooks/pre-commit` are
enforced byte-identical by `TestPreCommitTemplateMatchesConstant`
(`internal/cli/hook_install_precommit_test.go:38 @294b4b6ab`, verified present). REQ-PCP-011 changes
the hook body and therefore requires both edits in the same commit, followed by `make build`.

REQ-PCP-001 through REQ-PCP-010 and REQ-PCP-014 touch installer logic only and leave the hook body
untouched — a deliberate consequence of choosing the sidecar over an in-body stamp (§A.5).

### §C.2 Collision with card t237

Card t237 (issue #1641) edits the same constant/template pair on the module-root `go vet` axis. Only
REQ-PCP-011 collides. Run-phase should land REQ-PCP-011 as its own commit, last, so a rebase against
t237 touches one commit rather than the whole SPEC.

### §C.3 Non-interactive

`moai update` and `moai init` run unattended. No requirement here may introduce a prompt.

## §D Out of Scope

### Out of Scope — the `exec` redirection sweep (card requirement C)

- Not authored: this tree contains zero `exec <n><redirect>` occurrences (§A.4). There is no pattern
  here to fix.

### Out of Scope — `--no-hooks` over-blocking

- Confirmed but not fixed here: one `--no-hooks` flag gates both `installPrePushHookOptional` and
  `installPreCommitHookOptional` (`update_template_sync.go:572`/`:575` and `init.go:895`/`:898`,
  both `@294b4b6ab`), so a user cannot skip one tier without skipping the other. A separate concern
  from silent loss; belongs to its own card.

### Out of Scope — the pre-push installer

- `InstallPrePushHook` shares the marker-based design and is likely to share this defect. Not
  verified here, and deliberately excluded to keep this SPEC's diff attributable.

### Out of Scope — rendering the user's diff

- Showing *what* the user changed requires the full-content snapshot evaluated in §A.5 Decision 1.
  Deferred; the classifier contract is unchanged either way.

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
| REQ-PCP-014 | AC-PCP-014 |
