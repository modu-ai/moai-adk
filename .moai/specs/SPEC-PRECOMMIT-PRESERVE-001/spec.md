---
id: SPEC-PRECOMMIT-PRESERVE-001
title: "Pre-commit hook install must never silently discard a local patch"
version: "0.5.0"
status: in-progress
created: 2026-08-24
updated: 2026-08-25
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
| 2026-08-25 | 0.5.0 | Plan-audit iteration 3 remediation (`.moai/reports/t230/plan-audit-3.md`, PASS-WITH-DEBT 0.8875, flat). **D1 closed**: the release-composition constraint is restored as §A.5 Decision 3 and re-scoped from "the successor card" to **any** hook-body change competing for the classifier's release, naming card t237/#1641 (re-measured this round: issue open, verified patch on `t312-precommit-vet @ b6f478b1a`); §C.2's "moot / any order" is corrected to hold on the merge-conflict axis only; the constraint's three red-moments are named, one of them a DoD item (acceptance.md §D.3) that is not green by construction, and the release-candidate moment is routed outward rather than faked as an internal criterion. **D2 closed**: AC-PCP-005's Decides now runs `-v` with the run's skip status inspected, and its Then clause fails on a skipped sub-case (c) — the treatment AC-PCP-013 already used. **D3 closed**: acceptance.md §D.4 states sub-case (c)'s pin rule and the re-pin obligation on a legitimate body change, separating the REQ-PCP-005 check from the body-stability guard (which stays with plan.md §F's M2 scope gate and AC-PCP-013). **D5 closed** (the edge-case citation now names the creating call). D4 declined — a third notice element weakens a notice whose strength is its size, and REQ-PCP-006's disclosure floor holds without it. D6 needs no action. 12 REQ / 12 AC unchanged. |
| 2026-08-24 | 0.4.0 | Scope reduction after plan-audit iteration 2 (`.moai/reports/t230/plan-audit-2.md`, PASS-WITH-DEBT 0.8875). The `pre-commit.local` extension point (REQ-PCP-011, REQ-PCP-012) and the release-composition binding (REQ-PCP-015) are **removed from this SPEC and moved to a successor card** — see §D Out of Scope → the `pre-commit.local` extension point. Reason: all four of the iteration-2 blocking defects (N1 unenforceable release constraint, N2 three-referent "last released hook body", N3 unresolved t237 collision, N4 unmeasured first-upgrade population) exist only because M3 changes the hook body, which forces this SPEC to reason about release composition. Removing M3 dissolves N1-N3 outright; N4 survives on its own terms and is closed here by AC-PCP-005 sub-case (c). The auditor found M1 and M2 fit to enter run-phase as they stand. Minors N5 (AC-PCP-004 clause (ii) Decides), N6 (§A.4 self-counting numeral) and N7 (non-agent subject) also closed. 15 REQ/AC → 12. **Identifiers are not renumbered**: REQ/AC-PCP-011, -012 and -015 are retired, leaving gaps, because renumbering would invalidate every citation in the two audit reports and in this HISTORY. |

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
`294b4b6ab`; measurements added at v0.3.0 (§A.5 Decision 1's release-magnitude table, REQ-PCP-004,
§C.4) name `7b2f42be0`.
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
gets a handful of hits in this tree — every one of them prose or comments about
`printf … | exec <bin>` command replacement in `handle-stop-goal.sh`, its `.tmpl` twin,
`stop_goal_single_exec_test.go`, and the report/SPEC files that discuss this very sweep. None is a
redirection, and none contradicts the closure below; the looser glob simply matches a different
construct that shares the word `exec`. **No count is recorded for the loose pattern on purpose**:
this paragraph is itself one of its matches, so any numeral written here drifts on the next edit of
this file. The load-bearing figure is the strict pattern's `0`.

The condition does not hold. There is no upstream occurrence to sweep, so this SPEC authors no
requirement against it. Recorded as a measured closure so a later reader does not re-open it.
Per §A.0, this is a claim about **this tree at this SHA** — a different checkout may differ, and the
sweep is cheap to repeat.

### §A.5 Decisions

Three decisions. Each states its rejected alternatives so a reader can disagree with any one of
them independently. Decision 3 — the release composition — was recorded at v0.3.0, retired at v0.4.0
on the reasoning that a SPEC changing no hook bytes has no composition to decide, and **restored at
v0.5.0** on measurement: the hazard is not owned by whoever changes the body, it is owned by the
release that first ships the classifier, and card t237/#1641 is open and carries a verified
body-changing patch (§C.2). What was correctly retired at v0.4.0 was the *mechanical binding*
(REQ-PCP-015/AC-PCP-015), which sat where it could not fail; the decision itself is restated below
together with the moments at which it can.

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
byte-equal to the current `incoming`, so that user hits AC-PCP-005 sub-case (b) — no backup, no
notice, silent and correct — and is stamped for every release thereafter. **This SPEC leaves the
hook body untouched** (§C.1), which is what makes that sub-case reachable, and AC-PCP-005 sub-case
(c) measures it against the actual `v3.1.2` shipped bytes rather than a synthetic fixture. A user on
`v3.0.x` or earlier has a genuinely different body and takes one unnecessary backup plus one notice,
once.

Any change to the hook body inherits this arithmetic and must not be folded into the classifier's
own release: a body change shipping alongside the classifier makes `installed != incoming` true for
**every** user without exception, reproducing the alarm-fatigue failure §A.3 exists to prevent. The
constraint binds the **release**, not a card, so it binds every hook-body change competing for that
release — card t237/#1641 (open, with a verified patch on `t312-precommit-vet @ b6f478b1a`) exactly
as much as the successor card that carries the extension point (§D Out of Scope). Decision 3 below
states it, names the moments at which it can go red, and says what happens if it is overridden.

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

1. **At the release it would first ship in, it is a no-op.** This SPEC lands the classifier with
   the hook body unchanged, so the "previous shipped body" and the incoming body are the *same
   bytes* (measured above: `v3.1.0` through `v3.1.2` and HEAD are byte-identical). A constant equal
   to `sha256(incoming)` decides nothing that comparing against `incoming` does not already decide.
2. **Its only real coverage is version-skippers at some future body-changing release.** It starts
   mattering only when the hook body next changes, and only for a user who skipped the classifier
   release entirely — everyone who took that release already carries a record. That is a strictly
   smaller population than the one the sidecar covers, bought at ongoing cost.
3. **The ongoing cost lands on the exact seam this SPEC already avoided.** The digest must be
   updated in the same commit as any hook-body change, making a third member of the
   constant/template paired edit (§C.1) — the same pair card t237 edits (§C.2), and the same
   maintenance coupling that got the in-body version stamp rejected two rows above. Its failure
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

The notice accompanying a backup names the backup file, which is what makes (a) recoverable: the
user's patch is on disk under a path they were just told about. A *durable* home for those checks —
a `pre-commit.local` extension point the hook consults on every run, so the patch survives the next
update too — was scoped into this SPEC and has been moved to a successor card (§D Out of Scope).
Until it lands, recovery is manual re-application from the backup. That is a real limitation, and it
is stated rather than papered over; it does not weaken (a), because REQ-PCP-006 already guarantees
the user learns of the replacement at the moment it happens.

#### Decision 3 — the classifier's release carries no hook-body change (restored at v0.5.0)

**The decision**: the release that first ships the classifier ships the hook body byte-identical to
`v3.1.2`. Every hook-body change — card t237/#1641, the successor card carrying the extension point,
or any other — ships in a **later** release.

**Why the constraint outlived the requirement that carried it.** Decision 1's magnitude table does
not care which card edits the body; it cares what is in the release. A body change composed into the
classifier's own release makes `installed != incoming` for the entire no-record installed base, so
REQ-PCP-005 classifies 100% of it as user-modified and every user takes one backup and one notice on
first upgrade — the alarm fatigue §A.3 exists to prevent, delivered by the mechanism meant to prevent
it. Landing t237 *before* the classifier's release is the same failure with an extra edge: `v3.1.2`
bytes would then differ from the incoming body, so AC-PCP-005 sub-case (c) would assert the wrong
outcome (acceptance.md §D.4 states what must happen instead).

**Where it can go red, and what makes it red.** Stated explicitly, because the v0.3.0 binding
(REQ-PCP-015/AC-PCP-015) failed exactly here — it was checked at end-of-M2, where it could not fail.

| Moment | Check | Red when | Inside this SPEC's lifecycle |
|---|---|---|---|
| End of M2 — run-phase scope gate (plan.md §F) | `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit \| cmp - internal/template/templates/.git_hooks/pre-commit` → rc 0 | **this SPEC's own diff** touched the body | yes, but green by construction while the branch stays in scope — which is why it is a scope gate and not the composition check |
| Sync-phase, after this SPEC's branch merges into the release line (acceptance.md §D.3) | the same `cmp`, run against the **integration ref** rather than the working tree | a sibling merge — t237 or any other — has already put a body change on the release line | yes, and **not** green by construction: it reads the integration ref, so another card's merge flips it |
| Release-candidate cut | the same `cmp` against the candidate tag | the composed release carries a body change | **no** — this moment is after this SPEC closes, so it is routed outward as a release-checklist item rather than dressed up as an internal criterion |

**If the constraint is overridden.** A maintainer may compose them together anyway; this SPEC has no
authority over a release. That is an accept-the-noise decision, and it is not free: it must be
recorded in the release notes as a one-time backup-and-notice for the entire installed base, and
AC-PCP-005 sub-case (c) must be re-pinned per acceptance.md §D.4 **before** the release is cut,
because an unpinned (c) then asserts the wrong outcome and reads as a false regression.

**Rejected alternative — re-bind this with a requirement inside this SPEC.** Tried at v0.3.0
(REQ-PCP-015) and it failed: the only moments a requirement of this SPEC can be checked lie inside
this SPEC's own diff, where the constraint is green by construction. Inventing a second such check
would reproduce the defect rather than close it. The composition is therefore carried by this
decision plus a named check at each moment above — one of them a DoD item, one of them explicitly
outward — and the requirement count stays at 12.

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
  path and stating that the hook was replaced — on a **warning writer distinct from its progress
  writer**, and both callers shall bind that warning writer to the command's stderr.

  The notice names the backup and nothing else. It **shall not** direct the user to
  `.git/hooks/pre-commit.local`: this SPEC no longer delivers that extension point (§D Out of
  Scope), and an instruction to stage recovered checks in a file the installed hook never reads is
  worse than no instruction — it reads as a supported recovery path and silently is not one. When
  the successor card lands the extension point, extending this notice to name it is a one-line
  change to this requirement, made in the release that makes it true.

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
- **REQ-PCP-010** — When a supporting write fails, the installer shall not fail `moai init` or
  `moai update`; and the two supporting writes have **different precedence**, because they happen on
  different sides of the hook write:
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

12 requirements — within the Tier M ceiling of 16.

**Retired identifiers.** REQ-PCP-011 and REQ-PCP-012 (the `pre-commit.local` extension point) and
REQ-PCP-015 (the release-composition binding) were removed at v0.4.0 and moved to a successor card
(§D Out of Scope). The remaining requirements keep their original numbers, so the sequence reads
001-010, 013, 014 with three gaps. Renumbering was rejected: every citation in
`.moai/reports/t230/plan-audit.md` and `plan-audit-2.md`, and in this document's own HISTORY, is
written against the original numbers, and silently re-pointing them would make the audit trail lie.

## §C Constraints

### §C.1 Paired-edit obligation

`preCommitHookContent` (`internal/cli/hook_install_precommit.go:26 @294b4b6ab`, symbol anchor:
the `preCommitHookContent` const) and `internal/template/templates/.git_hooks/pre-commit` are
enforced byte-identical by `TestPreCommitTemplateMatchesConstant`
(`internal/cli/hook_install_precommit_test.go:38 @294b4b6ab`, verified present).

**This SPEC changes neither.** Every requirement here touches installer logic only and leaves the
hook body untouched — a deliberate consequence of choosing the sidecar over an in-body stamp (§A.5),
and, since v0.4.0, of moving the one body-changing requirement to a successor card (§D Out of
Scope). REQ-PCP-013 therefore has nothing to enforce within this SPEC's diff; it is retained as the
standing invariant that catches a one-sided edit, and AC-PCP-013's rejection of a SKIP result is
what keeps it from passing vacuously. A change to `preCommitHookContent` appearing during run-phase
is a scope violation, not an implementation detail (plan.md §D).

### §C.2 Card t237 — the merge conflict is dissolved, the release order is not

Card t237 (issue #1641) edits `preCommitHookContent` and its template twin on the module-root
`go vet` axis. Measured at v0.5.0: the issue is **open**, and it records a verified patch on branch
`t312-precommit-vet @ b6f478b1a`, so the body change is neither hypothetical nor far off. At v0.3.0
this SPEC collided with it through REQ-PCP-011, and the audit's N3 asked which card yields. The
answer splits by axis, and v0.4.0's "moot / either may land first in any order" was true of only one
of them:

- **Merge-conflict axis — dissolved.** With the extension point moved to a successor card (§D Out of
  Scope), this SPEC touches neither file (§C.1), so the two diffs are disjoint and merge cleanly in
  either order.
- **Release-order axis — open, and bound by §A.5 Decision 3.** Disjoint diffs still compose into one
  release, and a release carrying both makes `installed != incoming` for every user. The classifier
  ships first; t237 ships in a later release. Neither "same release" nor "t237 first" is available.

The paired-edit collision itself travels with the extension point: the successor card inherits it,
and inherits the mitigation with it — land the paired edit as its own commit so a rebase against t237
touches one commit rather than a whole SPEC.

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

### Out of Scope — the `pre-commit.local` extension point (moved to a successor card)

- **What moved**: a `.git/hooks/pre-commit.local` delegation — the installed hook executing that
  script as its final step and exiting with its status when it exists and is executable, and
  behaving exactly as today when it does not. Authored here at v0.2.0-v0.3.0 as REQ-PCP-011 and
  REQ-PCP-012, with AC-PCP-011 and AC-PCP-012; removed at v0.4.0.
- **Where it went**: a successor card, not the bin. It is the facility that makes a replaced patch
  *durable* rather than merely recoverable, and the card that carries it depends on this SPEC
  landing first, so that provenance records already exist when the body change ships.
- **Why it left**: it is the only requirement in this SPEC that changes the distributed hook body,
  and that single fact generated every blocking defect of plan-audit iteration 2 — a release
  constraint (REQ-PCP-015) checked where it cannot fail, a phrase with three referents, an
  unresolved collision with card t237, and an unmeasured first-upgrade population. Splitting it out
  dissolves the first three and leaves the fourth answerable on its own terms (AC-PCP-005 sub-case
  (c)).
- **The trap no body-changing card may walk into**: a hook-body change and a provenance classifier
  in the same release make `installed != incoming` true for **every** installed base without
  exception, so every user takes a backup and a notice on first upgrade — the alarm fatigue §A.3
  exists to prevent, delivered by the mechanism meant to prevent it. This binds the release rather
  than this card, so it binds card t237/#1641 identically (§C.2); the constraint, its red-moments,
  and its override cost are stated once in §A.5 Decision 3. The successor card ships after the
  classifier release, against an installed base that already carries records.
- Also not authored here: the notice's pointer to `pre-commit.local` (REQ-PCP-004 states why naming
  a facility that does not yet exist is worse than naming nothing).

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
| REQ-PCP-013 | AC-PCP-013 |
| REQ-PCP-014 | AC-PCP-014 |

12 requirements, 12 acceptance criteria, 1:1. REQ/AC-PCP-011, -012 and -015 are retired together
(§B, §D Out of Scope) — a requirement and its criterion always leave as a pair, so the parity holds
across the trim. AC-PCP-005 carries three sub-cases rather than a fourth criterion, which is why
closing the audit's N4 changes no count.
