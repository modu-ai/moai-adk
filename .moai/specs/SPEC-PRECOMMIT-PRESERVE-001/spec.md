---
id: SPEC-PRECOMMIT-PRESERVE-001
title: "Pre-commit hook install must never silently discard a local patch"
version: "0.3.0"
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
| 2026-08-24 | 0.3.0 | Plan-audit iteration 1 remediation (`.moai/reports/t230/plan-audit.md`, PASS-WITH-DEBT 0.875). D1 closed: REQ-PCP-010 now states the backup-failure precedence explicitly and separates it from the provenance-failure case. D2 closed: REQ-PCP-004's "on stderr" reconciled with the installer's single-writer signature by adding a warning writer and permitting the two-line caller change (plan.md §A). D3 closed: the one-entry previous-digest corpus is evaluated and rejected on its own terms (§A.5), the release composition is decided (§A.5 Decision 3 — the hook-body change ships after the classifier), and REQ-PCP-015/AC-PCP-015 bind that composition mechanically. D4 (sweep-pattern semantics, §A.4), D5 (backup-succeeded/write-failed edge row), D6 (AC-PCP-002 falsification label), D7 (`Where` → `When` on four requirements) also closed. 14 REQ/AC → 15. |

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

**Two SHAs appear in this document, and both are current.** The table above was measured at
`294b4b6ab`; measurements added at v0.3.0 (§A.5 Decision 3, REQ-PCP-004, §C.4) name `7b2f42be0`.
The difference between those two commits is the four SPEC files in this directory and nothing else,
so every source-file value reproduces identically at either — the anchor is the symbol, per §A.0.

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

**What the pattern means.** `exec [0-9][0-9]*[<>]` matches a **file-descriptor redirection** —
`exec` followed by *one or more* digits naming a descriptor, then `<` or `>`. The digits are not
optional. A reader repeating this sweep with the looser `exec [0-9]*[<>]` (zero or more digits)
gets 6 hits in this tree, all of them prose or comments about `printf … | exec <bin>` command
replacement in `handle-stop-goal.sh`, its `.tmpl` twin, `stop_goal_single_exec_test.go`, and two
report/spec files. None is a redirection, and none contradicts the closure below; the looser glob
simply matches a different construct that shares the word `exec`.

The condition does not hold. There is no upstream occurrence to sweep, so this SPEC authors no
requirement against it. Recorded as a measured closure so a later reader does not re-open it.
Per §A.0, this is a claim about **this tree at this SHA** — a different checkout may differ, and the
sweep is cheap to repeat.

### §A.5 Decisions

Three decisions. Each states its rejected alternatives so a reader can disagree with any one of
them independently.

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
noisy direction. The alternative — treating "unknown" as "unmodified" — is silent data loss for
exactly the population the card is about, since a hand-patched legacy hook is the most likely thing
to be found without a record.

**How loud is that first upgrade, actually?** This was previously written as if it were incidental
("a hook that *happens* to differ"). It is not incidental — it is decided entirely by whether the
release that introduces the classifier also changes the hook body, and the magnitude is measurable.
Measured in this worktree at `7b2f42be0`:

| Measurement | Command | Result |
|---|---|---|
| Last change to the distributed hook body | `git log -1 --format='%h %ad' --date=short -- internal/template/templates/.git_hooks/pre-commit` | `883d53852` 2026-07-28 |
| Releases carrying that body | `git tag --contains 883d53852` | `v3.1.0`, `v3.1.0-rc.0/1/2`, `v3.1.1`, `v3.1.2` |
| `v3.1.2` body vs HEAD body | `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit \| cmp - internal/template/templates/.git_hooks/pre-commit` | identical (rc 0) |
| `v3.0.1` installer constant vs HEAD constant | `sed -n '/preCommitHookContent = /,/^\`/p'` on each, piped to `shasum -a 256` | `f79adf7f…` vs `3442efc9…` — **differ** |

So the population splits cleanly. A user whose hook came from `v3.1.0`-`v3.1.2` has `installed`
byte-equal to the current `incoming`; if the classifier ships in a release that does **not** touch
the hook body, that user hits AC-PCP-005 sub-case (b) — no backup, no notice, silent and correct —
and is stamped for every release thereafter. A user on `v3.0.x` or earlier has a genuinely different
body and takes one unnecessary backup plus one notice, once. If instead REQ-PCP-011 (which changes
the body *by construction*) ships in the same release, `installed != incoming` holds for **every**
user without exception, and the design's own alarm-fatigue failure (§A.3) is reproduced in full by
the release that introduces the design. That is not a residual risk to note — it is a release-
composition decision, taken below as Decision 3.

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
| Compare against the **immediately-previous** shipped body only — a one-entry corpus (a single 64-char digest constant) | Rejected on its own terms, not as the expensive form above — see below. |

On the **one-entry corpus** (the cheap form of "compare against history", evaluated separately
because rejecting the ever-growing form is not a rejection of this one): store
`sha256(previous shipped preCommitHookContent)` as a second constant, and treat an installed hook
matching it as unmodified even with no record. It is genuinely cheap — one 64-character string, no
per-clone storage, no growth. It is rejected here for three measured reasons, and it stays available
because adopting it later re-opens no decision (it is simply an additional source of a `recorded`
value; the §A.3 classifier contract is unchanged).

1. **At the release it would first ship in, it is a no-op.** Decision 3 lands the classifier in a
   release whose hook body is unchanged, so the "previous shipped body" and the incoming body are
   the *same bytes* (measured above: `v3.1.0` through `v3.1.2` and HEAD are byte-identical). A
   constant equal to `sha256(incoming)` decides nothing that comparing against `incoming` does not
   already decide.
2. **Its only real coverage is version-skippers at the later release.** It starts mattering at the
   REQ-PCP-011 release, and only for a user who skipped the classifier release entirely — everyone
   who took that release already carries a record. That is a strictly smaller population than the
   one Decision 3 covers, bought at ongoing cost.
3. **The ongoing cost lands on the exact seam this SPEC already avoided.** The digest must be
   updated in the same commit as any hook-body change, making a third member of the
   constant/template paired edit (§C.1) — the same pair that collides with card t237 (§C.2), and the
   same maintenance coupling that got the in-body version stamp rejected two rows above. Its failure
   mode is benign (a stale digest degrades to today's noise, never to a missed detection), but it is
   silent, and it recurs on every future body change rather than once.

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

#### Decision 3 — release composition: the classifier ships before the hook-body change

The magnitude measurement above forces a choice `plan.md §F` previously left open (it stated two
resolutions and picked neither). The choice is taken here.

| Option | First-upgrade false alarms | Consequence |
|---|---|---|
| (a) M1+M2 ship first, body **unchanged**; M3 follows in a later release — **chosen** | only `v3.0.x`-and-earlier installs | `v3.1.0`-`v3.1.2` users are silent and correct, and are stamped; the M3 body change then lands against an installed base that has records, so it too is silent |
| (b) M1+M2+M3 ship together | **every** user, without exception | the alarm-fatigue outcome §A.3 exists to prevent, delivered by the release that introduces the prevention |
| (c) M3 ships first, M1+M2 later | none at the M1+M2 release | rejected: it ships a hook-body change through the *unprotected* installer, so the one cohort with patched hooks loses them silently in that release — paying with the card's own failure mode to buy quieter output later |

Concretely: REQ-PCP-001 through REQ-PCP-010 and REQ-PCP-014 target `v3.1.4` and leave
`preCommitHookContent` byte-identical to `v3.1.2`; REQ-PCP-011 and REQ-PCP-012 target the following
release. REQ-PCP-013 (constant/template parity) is a standing invariant and binds in **both**
releases — it is not deferred; it simply has nothing to enforce until the body changes. REQ-PCP-015
binds the composition mechanically so it is checked rather than remembered, and AC-PCP-015 is the
check. The SPEC's `phase:` frontmatter names `v3.1.4` because that is where its core lands; the M3
deferral is recorded here rather than by splitting the SPEC.

**The accepted wart.** In the `v3.1.4` release the notice names `.git/hooks/pre-commit.local`
(REQ-PCP-004) while the installed hook does not yet consult it — the facility arrives one release
later. This is accepted rather than papered over: the notice's reader is someone whose patch was
just replaced, whose immediate action is to recover it from the backup (which exists now) and stage
it in `pre-commit.local` (which begins running at the next update). A file created early is inert,
not wrong. The alternative — a two-phase notice text whose content depends on which release it ships
in — makes REQ-PCP-004 release-dependent, and buys one release of precision with a permanently more
complicated requirement.

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
- **REQ-PCP-005** — When no provenance record exists, or the record is unreadable, the installer
  shall compare the installed bytes against the incoming content and shall treat any difference as
  user-modified.

### Disclosure

- **REQ-PCP-003** — When the installed hook is classified user-modified, the installer shall copy it
  to `.git/hooks/pre-commit.bak.<UTC-timestamp>` before writing the new content.
- **REQ-PCP-004** — When a backup is written, the installer shall emit a notice — naming the backup
  path, stating that the hook was replaced, and naming `.git/hooks/pre-commit.local` as the place to
  re-apply local checks — on a **warning writer distinct from its progress writer**, and both
  callers shall bind that warning writer to the command's stderr.

  The two-writer split is not decoration. `installPreCommitHookOptional(projectRoot string, skip
  bool, out io.Writer)` (`internal/cli/hook_install_precommit.go:166 @7b2f42be0`) has exactly one
  writer today, and the `moai update` caller binds it to **stdout** (`out := cmd.OutOrStdout()` at
  `internal/cli/update_template_sync.go:69 @7b2f42be0`, inside `runTemplateSyncWithReporter` —
  the durable anchor, since a second unrelated `out := cmd.OutOrStdout()` exists at `:604` in a
  later function — passed at `:575`) while `moai init` binds it
  to stderr (`cmd.ErrOrStderr()` at `internal/cli/init.go:898 @7b2f42be0`). A data-loss notice on
  stdout under `moai update` is swallowed whole by a scripted `moai update >/dev/null` — the card's
  own silent-loss failure, reintroduced through the stream choice — and it contradicts the module
  convention that stderr carries warnings and stdout carries machine-readable output
  (`internal/cli/CLAUDE.md` § Conventions → Output streams, `@7b2f42be0`). The `moai update` caller
  already holds `errOut := cmd.ErrOrStderr()` at `update_template_sync.go:72 @7b2f42be0`, so the
  caller change is two lines; `plan.md §A` permits exactly that change and no more.
- **REQ-PCP-007** — The installer shall not report the unqualified success line as its only output
  on a run in which a backup was taken.
- **REQ-PCP-009** — The installer shall not overwrite an existing backup file; when the chosen
  backup path is occupied it shall select a distinct path and shall still produce a backup.

### Preserved behaviour

- **REQ-PCP-008** — When the installed hook does not carry the MoAI marker, the installer shall
  return `ErrUserHookExists` without writing a backup and without modifying the hook, exactly as
  today.
- **REQ-PCP-010** — Failure of a supporting write shall never fail `moai init` or `moai update`, and
  the two supporting writes have **different precedence** because they happen on different sides of
  the hook write:
  - **When the backup write fails** — which is *before* the hook write — the installer shall report
    a warning, shall **not overwrite the hook**, and shall not fail the caller. The user's bytes
    stay on disk; nothing is lost.
  - **When the provenance write fails** — which is *after* a hook write that already succeeded —
    the installer shall report a warning and shall not fail the caller. The hook stays replaced;
    the missing record is self-healing, since the next run finds no record, compares installed
    against incoming, finds them equal, and re-stamps (REQ-PCP-005 → AC-PCP-005 sub-case (b)).

  The first clause is the one that must be written down. Without it, "warn and do not fail the
  caller" admits an implementation that warns and overwrites anyway — destroying the patch with no
  recoverable backup, which is precisely the outcome REQ-PCP-006 forbids and this SPEC exists to
  prevent. Realistic triggers: a read-only `.git/hooks/`, quota exhaustion, or a link/rename backup
  strategy failing across a filesystem boundary while the 3,245-byte overwrite still succeeds. This
  precedence is deliberate and is **not** a carve-out from REQ-PCP-006: no backup means no
  replacement, so the invariant "never replace a user-modified hook silently" holds by never
  replacing it at all.
- **REQ-PCP-013** — `preCommitHookContent` and
  `internal/template/templates/.git_hooks/pre-commit` shall remain byte-identical.

### Local extension point

- **REQ-PCP-011** — When the pre-commit hook runs and `.git/hooks/pre-commit.local` exists and is
  executable, the hook shall execute it as its final step and shall exit with that script's status.
- **REQ-PCP-012** — When the pre-commit hook runs and `.git/hooks/pre-commit.local` is absent or not
  executable, the hook shall behave as it does today.

### Release composition

- **REQ-PCP-015** — The release that first ships REQ-PCP-001 through REQ-PCP-010 and REQ-PCP-014
  shall leave `preCommitHookContent` byte-identical to the most recently released hook body, and
  REQ-PCP-011's hook-body change shall ship in a later release (§A.5 Decision 3).

15 requirements — within the Tier M ceiling of 16.

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
t237 touches one commit rather than the whole SPEC. §A.5 Decision 3 makes this stronger than a
sequencing preference: REQ-PCP-011 ships in a *later release* than the rest, so the collision window
is a separate release cycle rather than a separate commit within one.

### §C.3 Non-interactive

`moai update` and `moai init` run unattended. No requirement here may introduce a prompt.

### §C.4 Permitted caller change (the one exception)

REQ-PCP-004 requires a warning writer distinct from the progress writer, which the current
single-writer signature cannot express. This SPEC therefore permits **exactly one** change at each
call site: passing the already-available stderr writer as the new parameter —
`errOut` at `internal/cli/update_template_sync.go:72 @7b2f42be0`, and `cmd.ErrOrStderr()` at
`internal/cli/init.go:898 @7b2f42be0`. Nothing else in either caller changes: not the `--no-hooks`
gate, not the ordering relative to `installPrePushHookOptional`, not the existing progress writer.
`plan.md §A` carries the same permission in the same words; a wider caller change is out of scope
and belongs to its own card.

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
| REQ-PCP-015 | AC-PCP-015 |
