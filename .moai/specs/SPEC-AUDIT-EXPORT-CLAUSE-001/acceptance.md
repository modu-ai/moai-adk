# SPEC-AUDIT-EXPORT-CLAUSE-001 — acceptance criteria

> Every AC names a command and the output that closes it. Commands run from the
> repository root of the tree under test, as **one compound invocation each**,
> with no command substitution and no writes to the tree outside `/tmp`
> (plan §E). A zero-count result is admissible only when the AC's paired positive
> control returned non-zero **in the same run** (§D.3).
>
> **[HARD] Named deviation — this is looser than `verification-completeness.md`
> §2.1's single-invocation form, and the difference is stated rather than
> glossed.** §2.1 places *"pipes, redirection, `&&`, `;` chaining, and
> subshells"* outside that form. Measured over the fifteen criteria below, **ten
> fall outside it**: pipes in AC-AEC-003, -004, -007, -008, -013, -014;
> `/tmp` redirection in AC-AEC-003 and -008; `&&` / `||` chaining in AC-AEC-002
> and -011; and `; echo "exit=$?"` in AC-AEC-001, -002, -003 and -012. AC-AEC-015
> additionally reads `$?` in a following invocation. Five criteria — AC-AEC-005,
> -006, -009, -010, -015 — are single invocations throughout.
>
> **The shape is forced, not chosen.** The conforming form would put the
> substitution inside the command (`$(git merge-base …)`, `<(git ls-tree …)`),
> and the worktree-session guard refuses both: it cannot statically verify a git
> command inside a substitution, so every such form is denied in this tree. The
> deviation is documented at acceptance §AC-AEC-003's closing note and
> plan §E.
>
> **These criteria keep release-blocking status, and the reason is §2.1's own
> trigger.** §2.1's consequence for a non-conforming citation is the
> *undecidable disposition*, whose stated trigger is a cited RED that **cannot be
> re-executed on the current tree** — a historical event, an already-merged
> state, or an externally observed CI result. That trigger is not met here: every
> RED cell below was re-executed verbatim on this tree and produced a decidable
> result with a firing control. A criterion that is decidable, reproducible, and
> paired is not what the undecidable disposition was written to demote. Where a
> future tree makes any RED below unreproducible, that criterion takes the
> disposition then — regression-guard, not a pass.

---

## §A Given-When-Then scenarios

### AC-AEC-001 — the negation and its explanatory block are gone

**Given** the `.gitignore` withdrawal specified in plan §A.2,
**When** the file is searched for any negation re-including a **verdict**
artifact under a card report directory, and for the comment block that
introduced it,
**Then** neither is found.

```bash
grep -nE '^!\.moai/reports/[^/]*/.*verdict' .gitignore ; echo "negation_exit=$?"
grep -nc 'Narrow exception (card t1039)' .gitignore ; echo "block_exit=$?"
```

Expected, **stated per command** because the two behave differently:

| Command | Expected stdout | Expected exit |
|---|---|---|
| the `-nE` negation probe | no output | `negation_exit=1` |
| the `-nc` block probe | the single line `0` | `block_exit=1` |

`grep -c` **always** prints a count — it prints `0` and exits 1 on no match — so
"no output from either" could never be met by the command as written. A reader
closing this AC literally against the former wording had to mark it FAIL; the
table above is what the commands actually do.

Paired control — the directive block and its blanket, which this card does **not**
touch, MUST still match, proving the greps reach the file:

```bash
grep -c 'evidence for lead verdicts stays on disk' .gitignore
grep -cE '^\.moai/reports/\*$' .gitignore
```

Expected: both `1`.

> **The probe is anchored on the verdict name, not on the glob form, and that is
> the repair.** The former probe `^!\.moai/reports/\*/` could only match a
> literal `*` path component, so it was structurally incapable of seeing a
> negation written with a literal card id — it returned `negation_exit=1`
> whatever the file contained below that line. Four such negations do survive at
> HEAD by design (REQ-AEC-001 carve-out); the requirement was therefore narrowed
> to *verdict artifact* and the probe widened to `[^/]*` in the same pass, so
> criterion and requirement now have the same reach. Measured: the widened probe
> matches the withdrawn line `!.moai/reports/*/verdict.md` at `269fb89c1` and
> matches none of the four surviving fixture negations at `113e487c2`.
>
> **RED-now / green path**, both single invocations:
>
> ```
> $ git show 269fb89c1:.gitignore > /tmp/pre.txt
> $ grep -nE '^!\.moai/reports/[^/]*/.*verdict' /tmp/pre.txt
> 259:!.moai/reports/*/verdict.md
> exit 0
> $ grep -nE '^!\.moai/reports/[^/]*/.*verdict' .gitignore
> exit 1
> ```
>
> `269fb89c1` is the pinned pre-withdrawal commit (the merge base of this branch
> with `develop`), not a branch name — the ref is the address at which the red
> was measured, so it takes the ANCHOR remedy.
>
> **Marker-string collision — a trap for the next editor.** The withdrawn block's
> marker was `Narrow exception (card t1039)`, singular. The surviving fixture
> block at `.gitignore:255` is headed `# Narrow exceptions: test guard fixtures…`
> — near-identical prose, plural, different rule. The second probe distinguishes
> them correctly today because it matches the singular form with its
> parenthetical, but an editor who relaxes it to `Narrow exception` will silently
> start matching the block this card preserves.

Maps REQ-AEC-001

### AC-AEC-002 — the blanket rule is in effect again for a new path

**Given** that the withdrawal restores `.moai/reports/*` to full effect,
**When** two names under a card report directory are tested — the one the
negation used to re-include and one it never did —
**Then** both report as ignored, and the plan-audit carve-out still behaves as
before.

```bash
git check-ignore --no-index .moai/reports/ZZAC002/verdict.md && echo IGNORED || echo NOT-IGNORED
git check-ignore --no-index .moai/reports/ZZAC002/plan-audit.md && echo IGNORED || echo NOT-IGNORED
git check-ignore --no-index .moai/reports/plan-audit/.gitkeep && echo IGNORED || echo NOT-IGNORED
```

Expected, in order: `IGNORED`, `IGNORED`, `NOT-IGNORED`.

Paired control — a path outside the reports tree MUST report `NOT-IGNORED`,
proving the probe can produce that answer at all:

```bash
git check-ignore --no-index README.md && echo IGNORED || echo NOT-IGNORED
```

Expected: `NOT-IGNORED`.

**Instrument self-check — the two forms must be shown to disagree somewhere in
this tree.**

```bash
git check-ignore -v --no-index .moai/reports/t338/ac-count-baseline.txt ; echo "verbose_exit=$?"
git check-ignore    --no-index .moai/reports/t338/ac-count-baseline.txt ; echo "plain_exit=$?"
```

Expected: the verbose form prints the matched rule line and `verbose_exit=0`; the
plain form prints nothing and `plain_exit=1`.

> **The plain form is load-bearing; `-v` answers a different question.** The
> verbose form reports that *a pattern matched* and exits 0 even when the
> matching pattern is a **negation** — i.e. when the path is **not** ignored. A
> criterion keyed on its exit code reads "matched" as "ignored" and passes on a
> tree where a negation survived.
>
> **Scoping "here" — the divergence is measured on a surviving fixture, not on
> `verdict.md`.** The original statement of this note attributed the divergence
> to `verdict.md` **measured before the withdrawal, with the negation live**, and
> that attribution was correct for the pre-withdrawal tree it named. It is no
> longer true of the current one: `113e487c2` withdrew the only negation covering
> `verdict.md`, so the two forms now **agree** on that path and the note would
> read as a claim about this tree that this tree refutes. The probe above uses
> `.moai/reports/t338/ac-count-baseline.txt` instead — one of the four
> CI-fixture negations REQ-AEC-001 deliberately preserves — where the two forms
> still disagree at HEAD `113e487c2` (measured: verbose `0`, plain `1`). The
> criterion therefore carries a name on which the wrong instrument is still
> observably wrong, rather than resting on a historical attribution.
>
> **AC-AEC-002's own target is unchanged by this.** Its job is detecting an
> *incomplete* withdrawal: had the negation survived, `verdict.md` would report
> `NOT-IGNORED` under the plain form and the first probe would fail, while a
> `-v`-instrumented version would exit 0 and pass falsely. No probe file is
> created — the `--no-index` form answers for a path that does not exist, so this
> AC writes nothing and has no cleanup step.

Maps REQ-AEC-001, REQ-AEC-002

### AC-AEC-003 — the tracked population differs from the remote by exactly the two ruled removals

**Given** the withdrawal (which untracks nothing, SPEC §A.3) and the ruled index
act (which removed exactly two tracked verdict files, SPEC §A.3a),
**When** the tracked set and the remote-present set are each enumerated with the
anchored criterion,
**Then** every tracked verdict file is on the remote (`comm -23` **empty**), the
remote-only remainder is **exactly** the two ruled names —
`.moai/reports/t1039/verdict.md` and `.moai/reports/t1048/verdict.md` — and
nothing under `.moai/reports/` remains as an uncommitted index change.

```bash
git ls-files '.moai/reports/*/verdict.md' | sort > /tmp/aec003-tracked.txt
git ls-tree -r --name-only origin/develop -- .moai/reports | grep -E '^\.moai/reports/[^/]+/verdict\.md$' | sort > /tmp/aec003-remote.txt
comm -23 /tmp/aec003-tracked.txt /tmp/aec003-remote.txt
comm -13 /tmp/aec003-tracked.txt /tmp/aec003-remote.txt
git status --porcelain -- .moai/reports ; echo "status_exit=$?"
```

Expected: the `comm -23` output **empty**; the `comm -13` output **exactly the
two lines** `.moai/reports/t1039/verdict.md` and `.moai/reports/t1048/verdict.md`
and no others; the `git status` output empty with `status_exit=0`.

> **The Then-clause was re-baselined by the operator's disposition of
> 2026-09-22, and the RED that forced it was observed on this tree.** The
> pre-amendment Then-clause read *"neither difference direction is non-empty"*.
> Measured at HEAD `ce6de9407` (2026-09-22, this run): `comm -23` empty,
> `comm -13` returning exactly `.moai/reports/t1039/verdict.md` and
> `.moai/reports/t1048/verdict.md`, `status_exit=0` — a **FAIL** against the old
> text and the **PASS** state against the ruled one, because the disposition
> ruled that state final (SPEC §A.3a): the two files stay on `origin/develop` and
> stay intentionally absent from this branch's index. The command is unchanged;
> only the expected state moved, and any third name in either direction still
> fails.
>
> **A count pair cannot distinguish the outcome from its failure modes.** At HEAD
> `ce6de9407` the sides measure `10` and `12`, and a `10` / `12` pair would also
> read the same if two *different* files had left the index while two other files
> arrived on the remote. The two `comm` directions separate those cases; the
> counts do not, which is why the closing evidence is the exact two-way
> difference and the counts are not the criterion at all.
>
> **Directional reading.** A non-empty `comm -23` means a tracked verdict file
> has not reached the remote — under §A.3a's predicate that is the state whose
> removal is authorized, so it is a finding to route to the operator rather than
> to act on silently. A `comm -13` whose set is exactly the two ruled names is
> the operator-ruled intentional state, recorded as such at the 2026-09-22
> disposition (SPEC §A.3a); a `comm -13` carrying **any other name** is the state
> REQ-AEC-003 forbids creating, and it fails this criterion outright.
>
> **A substring criterion is a FAIL, not a near-pass.** `grep 'verdict.md'` on
> the remote side also matches `.moai/reports/t965/plan-audit-verdict.md` and
> returns **11** rather than 10 (measured at `113e487c2`). The anchored form
> `^\.moai/reports/[^/]+/verdict\.md$` is used on **both** sides; a criterion
> anchored on one side only would compare a 10-element set against an
> 11-element one and read the surplus as a real difference.
>
> **`origin/develop` is left unpinned deliberately** — the four tests of
> `verification-claim-integrity-detail.md` § Moving-ref predicate were applied
> (`verification-claim-integrity.md` §2.1) and land **SUBJECT / S2 → remedy R4**:
> the claim is about what the remote **currently** carries, so pinning the ref to
> a SHA would narrow it into a statement about one past commit and let the
> criterion keep passing after new verdict files reach the remote. R4's form is
> satisfied here — the measuring command leads, the criterion carries no count at
> all, and the 10/10 figures appear only as a dated reference pinned to
> `113e487c2`. A non-empty `comm -13` is therefore a true signal about the
> subject, not upstream drift.
>
> Both `comm` invocations read files written by the two preceding commands rather
> than by process substitution: the worktree guard refuses a git command it
> cannot statically verify, and `<(git …)` is such a form. The two temp files are
> the substitution-free equivalent and write nothing into the tree.

Maps REQ-AEC-002, REQ-AEC-003

### AC-AEC-004 — no surface designates a card-report artifact as tracked

**Given** the twelve wording surfaces enumerated in plan §C (scope rows #2–#13),
**When** they are swept with the markup-insensitive pattern that located them,
**Then** no match remains.

```bash
grep -rnE 'tracked\*{0,2} (citation target|verdict file|path)' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the replacement phrasing MUST match in the same trees, proving
the sweep reaches the repaired files rather than missing them:

```bash
grep -rlE 'local\*{0,2} record' \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md | wc -l
```

Expected: `6`.

> **Why an enumerated file list rather than tree roots.** REQ-AEC-004's subject is
> the plan §C surface set, so the sweep's reach must be exactly that set — and an
> enumerated list is the form a reader can compare against the table row by row.
> The former six-tree-root form swept whole trees and reached files outside the
> table: measured at HEAD `ce6de9407`, it returned one match,
> `internal/template/templates/.moai/README.md:77` (*"regenerable artifacts out
> of tracked paths"*), a sentence pre-existing since `d97980654` (2026-07-07) and
> not in the scope table. All 8 in-scope lines were repaired; the criterion's
> trees were wider than its requirement's scope. Narrowed by the operator
> disposition of 2026-09-22.
>
> **Why `.gitignore` (scope row #1) is not in the list either.** Its act in this
> card is the withdrawal (REQ-AEC-001), measured by AC-AEC-001 and AC-AEC-002 —
> not the wording repair. Sweeping it with this pattern would re-catch a
> pre-existing sentence this card did not author (the `t196` block added by
> `d791fd29d`, *"The tracked citation target is the verdict file…"*), the same
> out-of-scope shape the narrowing removes, one file over. Its added lines are
> swept where they belong, in AC-AEC-014.
>
> **RED-now cell — re-derived against the pinned pre-repair tree, not carried
> forward.** The eight pre-repair lines live in six files, each read from its
> blob at `269fb89c1` (the merge base of this branch with `origin/develop`;
> measured 2026-09-22):
>
> ```
> $ git show 269fb89c1:.claude/rules/moai/core/agent-common-protocol-reference.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 1                                  # line 62
> $ git show 269fb89c1:.claude/rules/moai/core/agent-common-protocol.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 1                                  # line 274
> $ git show 269fb89c1:.claude/agents/moai/manager-lead.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 2                                  # lines 63, 152
> $ git show 269fb89c1:internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 1                                  # line 62
> $ git show 269fb89c1:internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 1                                  # line 274
> $ git show 269fb89c1:internal/template/templates/.claude/agents/moai/manager-lead.md \
>     | grep -cE 'tracked\*{0,2} (citation target|verdict file|path)'
> 2                                  # lines 65, 154
> ```
>
> Total: **8** matching lines, all exit 0 — the red the amended criterion starts
> from. Its firing positive control on the same tree:
> `git show 269fb89c1:.claude/rules/moai/core/agent-common-protocol.md | grep -c 'citation target'`
> → `1`, exit 0, proving the instrument reads the pinned blobs. **Green path:**
> M4's rewrite (SPEC §C.5); the same sweep at HEAD `ce6de9407` returns no output,
> `exit=1`.
>
> **The pattern carries `path` as a third alternative deliberately.** The v0.2.0
> draft measured these surfaces as `tracked** path` and excluded them; the
> `develop` absorb rewrote that phrase out of existence, so a sweep for it now
> returns a clean zero on unrepaired files (SPEC §A.6). Enumerating the retired
> phrasing alongside the current one costs nothing and stops the next absorb from
> producing the same false zero in reverse.

Maps REQ-AEC-004

### AC-AEC-005 — the withdrawn sentence is gone from every auditor copy

**Given** the four agent definitions enumerated in SPEC §A.4,
**When** they are searched for the removed sentence,
**Then** every file reports zero.

```bash
grep -rc 'Never write the verdict to a gitignored location' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:0`.

Paired control — a string known to remain present MUST report non-zero in every
file, proving the instrument can find text in them at all:

```bash
grep -rc 'Export mandate' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher. `Export mandate` is chosen
because the pattern crosses no emphasis marker: the clause carries `**bold**`
runs, and a pattern spanning a `**` delimiter matches nothing regardless of the
file's content.

Maps REQ-AEC-004, REQ-AEC-008

### AC-AEC-006 — the local-by-design statement is present in every auditor copy

**Given** the SPEC §C.1 replacement tail,
**When** the four files are searched for the statement the repair adds,
**Then** every file reports at least one match.

```bash
grep -rc 'local by design' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher.

Maps REQ-AEC-005

### AC-AEC-007 — no repaired wording names a remote-placing act

**Given** REQ-AEC-006,
**When** every file this card changes is searched for a forced stage or any other
verb that would place a card-report artifact on the remote,
**Then** none is found.

```bash
grep -rnE 'add -f|--force.*reports|push.*verdict' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the prohibition the replacement states instead MUST match,
proving the sweep reaches the repaired files:

```bash
grep -rlc 'do not force it into the tree' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md | wc -l
```

Expected: `4`.

> **This criterion exists because the defect it guards dissolved rather than being
> fixed.** The iteration-2 audit found that the permission clause (*"which this
> mandate permits for exactly this destination"*) had no criterion, so a wording
> omitting it would pass every criterion. Under the operator's ruling that clause
> is withdrawn, so there is nothing left to test for presence — and the residual
> risk inverts: the hazard is now that the verb **survives** a partial edit. The
> criterion is therefore a prohibition test over every changed file, not a
> presence test over four.

Maps REQ-AEC-006

### AC-AEC-008 — the FORBIDDEN prohibition survives and rests on a property that still holds

**Given** that each auditor copy carries its own FORBIDDEN wording, and that the
convention's § Where bullet justified the directory by the ignore rules,
**When** each surface is inspected,
**Then** the per-copy prohibition is intact and the convention's justification no
longer rests on being gitignored, while retaining its `.gitignore` pointer.

```bash
grep -c 'FORBIDDEN' .claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md
grep -c 'plan-audit/` is FORBIDDEN' .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md | tr '\n' ' ' | grep -cE 'deliberately[^.]{0,30}gitignored'
awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md | grep -cE '`\.gitignore` comment|never read as a card'
```

Expected: the two `FORBIDDEN` counts non-zero; the third command `0` (the
withdrawn justification); the fourth command non-zero (the retained `.gitignore`
pointer and the read-based distinguishing property).

Paired control for the **negative** branch — the same normalised pipeline, run
over the pre-repair blob, MUST return non-zero. Without it a `0` on the third
command is indistinguishable from a pattern that cannot fire:

```bash
git show 269fb89c1:.moai/docs/audit-artifact-convention.md > /tmp/aec008-pre.md
awk '/^## Where/,/^## When/' /tmp/aec008-pre.md | tr '\n' ' ' | grep -cE 'deliberately[^.]{0,30}gitignored'
```

Expected: `1`.

The control reads a **pinned blob** (`269fb89c1`), so it returns `1` whatever the
repair does — which is what makes it a firing control for the negative branch
rather than a second copy of the probe. Measured at HEAD `113e487c2`, before M3
runs, the probe **also** returns `1`: that is this criterion's RED-now cell, and
its green path is the same probe returning `0` once §C.3 is applied to the
convention document.

> **The `\n?` is removed, and that removal is the repair.** The former pattern
> was `deliberately.{0,3}\n?.{0,20}gitignored`, and grep is line-oriented: no
> POSIX grep can match across a line break, so the alternative containing `\n?`
> could only fire on a grep that spans lines. In the convention document
> `deliberately` ends one line and `gitignored` begins the next — exactly the
> layout the branch exists to catch. Measured on the **unmodified** document at
> HEAD `113e487c2`, the two greps on this machine disagree:
>
> ```
> $ awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md \
>     | grep -cE 'deliberately.{0,3}\n?.{0,20}gitignored'
> 1                              # PATH grep == ugrep 7.8.4, which spans lines
> $ awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md \
>     | /usr/bin/grep -cE 'deliberately.{0,3}\n?.{0,20}gitignored'
> 0                              # BSD grep — cannot span lines
> ```
>
> Opposite verdicts on one unmodified file, and CI runs the flavour that returns
> `0` — so the negative assertion passed **vacuously against the unrepaired
> document**. The normalised form above (`tr '\n' ' '` before the match) has one
> reading on every flavour; measured on the same unmodified document it returns
> `1` under **both** greps, which is the RED this criterion needs.
>
> **The two branches are split into separate commands deliberately.** The former
> single alternation mixed a presence assertion and an absence assertion into one
> count, so a non-zero result could not say which branch produced it. Separated,
> each has its own expected value and the absence branch has its own firing
> control.

Maps REQ-AEC-007

### AC-AEC-009 — the export obligation is not weakened

**Given** REQ-AEC-008,
**When** the four auditor copies are searched,
**Then** the incomplete-audit declaration still stands in each.

```bash
grep -rc 'incomplete audit' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four non-zero lines.

Maps REQ-AEC-008

### AC-AEC-010 — the convention document's two false claims are withdrawn and its check still holds

**Given** § Committing and § What makes the convention stick,
**When** both sections are read,
**Then** neither claims the artifact is tracked or reaches the integration branch,
the disposal hazard survives with its mitigation, and the mechanical check names
presence on disk rather than branch reachability.

```bash
awk '/^## Committing/,/^## What makes/' .moai/docs/audit-artifact-convention.md
awk '/^## What makes/,/^## Cross-references/' .moai/docs/audit-artifact-convention.md
```

Expected, read together: the first output contains `is a local file` and
`disposal hazard` and does **not** contain `Audit artifacts are tracked files` or
`reach the integration branch`; the second contains `ls .moai/reports/<card-id>/`
and does **not** name `git ls-files --error-unmatch` or any other
branch-reachability check.

```bash
grep -c 'Audit artifacts are tracked files' .moai/docs/audit-artifact-convention.md
grep -c 'ls-files --error-unmatch' .moai/docs/audit-artifact-convention.md
grep -c 'disposal hazard' .moai/docs/audit-artifact-convention.md
```

Expected: `0`, `0`, non-zero. The third is the paired control: it is a sentence
this card preserves, so a zero there means the greps did not reach the file and
the two zeros above are void.

Maps REQ-AEC-004, REQ-AEC-013

### AC-AEC-011 — the two convention copies are byte-identical

**Given** the local convention doc and its template mirror,
**When** they are compared,
**Then** they are byte-identical.

```bash
cmp internal/template/templates/.moai/docs/audit-artifact-convention.md \
    .moai/docs/audit-artifact-convention.md && echo BYTE-IDENTICAL
```

Expected: `BYTE-IDENTICAL`, exit 0.

Maps REQ-AEC-009

### AC-AEC-012 — the emitted codex layer matches its regenerated source

**Given** that three C2 agent copies were modified,
**When** the emission drift check runs,
**Then** it exits 0 without regenerating anything, and the emitted layer carries
neither withdrawn text.

```bash
make agents-emit-check ; echo "exit=$?"
grep -rc 'Never write the verdict to a gitignored location' \
  internal/template/templates/.codex/agents/moai/plan-auditor.toml \
  internal/template/templates/.codex/agents/moai/sync-auditor.toml
grep -rcE 'tracked\*{0,2} verdict file' \
  internal/template/templates/.codex/agents/moai/manager-lead.toml
```

Expected: `exit=0`; the second command's two lines each ending `:0`; the third
ending `:0`.

Paired control — the three TOMLs MUST still carry text the emission preserves,
proving the greps reached them:

```bash
grep -rc 'Export mandate' \
  internal/template/templates/.codex/agents/moai/plan-auditor.toml \
  internal/template/templates/.codex/agents/moai/sync-auditor.toml
grep -c 'Context-Folding' internal/template/templates/.codex/agents/moai/manager-lead.toml
```

Expected: all non-zero. Measured before implementation, the three probes return
`:1`, `:1`, `:2` — so each has an observed red and a zero here is a real change.

Maps REQ-AEC-011

### AC-AEC-013 — template neutrality on every template-tree edit

**Given** the six files changed under `internal/template/templates/`,
**When** **the lines this card adds** are swept for the forbidden content classes,
**Then** no match is found in any class.

```bash
git diff -U0 develop...HEAD -- internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -E 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|REQ-[A-Z]|CLAUDE\.local|/Users/|\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b|\b[0-9a-f]{9,40}\b'
echo "exit=$?"
```

Expected: no output, `exit=1`. **This is the sole closing command for this AC.**

Paired control — the same added-line extraction, filtered for a token the repair
introduces, MUST match; otherwise the empty result above means the diff pipeline
reached nothing rather than that the classes are absent:

```bash
git diff -U0 develop...HEAD -- internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' | grep -c 'local by design'
```

Expected: non-zero.

> **Revision-pinned, and three-dot deliberately.** A working-tree `git diff` with
> no revision compares against the index, so both the probe and its control return
> empty the moment the change is staged — and under §D.3 the criterion then cannot
> be closed at all. The three-dot form re-derives the merge base at read time, so
> it survives staging, commit, and a later absorb of `develop`.
>
> **[HARD] It is empty before the change is committed, so this criterion closes
> only after the milestone's commit lands.** `develop...HEAD` compares the merge
> base to the **commit** `HEAD`; uncommitted working-tree edits are invisible to
> it whether staged or not. During M2-M4, before anything is committed, both the
> probe and its control return empty and §D.3's zero-result rule correctly refuses
> the close — the risk is a blocked close, not a false pass, which is why the
> ordering is stated rather than mechanised. Measured at HEAD `113e487c2` with
> only M1 committed: the AC-AEC-014 pathspec yields `0` added lines, while the
> same pathspec with `.gitignore` included yields `39`. The three-dot form is what
> makes the criterion *runnable* once committed; it does not make it runnable
> *before*. See plan §E and M6's exit condition.
>
> The moving ref is
> kept rather than pinned because the claim is *what this card added relative to
> its branch point*; a literal SHA would falsify it at the first absorb
> (`verification-claim-integrity.md` §2.1, SUBJECT class). `$(git merge-base …)` is
> not used: the worktree guard refuses a git command inside a command
> substitution, which is what made the v0.2.0 form unrunnable where cards are
> worked.
>
> **A whole-file form is not an alternative.** Measured against the
> pre-implementation tree, it returns 19 matching lines in
> `internal/template/templates/.claude/agents/moai/plan-auditor.md` alone —
> rubric placeholders and documentation examples the regex over-matches — so it
> reads as a FAIL that is false by construction.

Maps REQ-AEC-010

### AC-AEC-014 — no introduced sentence asserts this repository's tree state

**Given** REQ-AEC-012,
**When** the lines this card adds across **all thirteen** scope files are swept
for declarative tree-state forms, in every verb form that carries the claim and
across an interposed adverbial,
**Then** the match set and the declared exception set below are **equal** —
every match is a member of the table, and every entry of the table is matched;
neither difference direction is non-empty.

> **Equality, not subset — and the empty-match case is why.** A subset test is
> satisfied vacuously by an empty match set, which is not hypothetical here: it
> is exactly the state of the declared pre-commit window (§D.1 item 6), and it
> recurs once this card merges and `develop...HEAD` empties. Equality refuses
> that close instead of passing it, and matches what the Expected line and
> §D.1 item 7 already required — the three now state one strength rather than
> two.

```bash
git diff -U0 develop...HEAD -- \
  .gitignore \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -nEi '\b(is|are|was|were|remains?|stays?|stayed|becomes?|became|has been|have been)( not)?( [a-z]+ly| already| still| now)? (ignore-matched|gitignored|tracked|untracked|on the remote)\b'
echo "exit=$?"
```

Expected: exactly the nine lines of the declared exception set below, and no
others.

**Declared exception set — 9 entries, every one in a class REQ-AEC-012 permits.**
The regex is a screen over surface forms; REQ-AEC-012's permission is about tense
and subject, which no such screen can express. Rather than narrow the regex until
it stops matching permitted lines — the move that produced this criterion's
previous two defects — the permitted matches are enumerated here and the
criterion closes on the **set difference** being empty:

| # | File | Matched line (verbatim) | Why permitted |
|---|---|---|---|
| E1 | `.gitignore` | `# untrack a file that is already tracked, so the entries already in the index` | General git behaviour. The subject is *a file*, indefinite — a statement of how git behaves, not a claim about this tree. REQ-AEC-012 permits it explicitly. |
| E2 | `.gitignore` | `# its own leaves whatever is already tracked exactly where it was.` | General git behaviour. Subject *whatever is already tracked*, a generic quantifier; the sentence is the same mechanism statement continued. |
| E3 | `.claude/rules/moai/core/agent-common-protocol.md` | `it is gitignored, so it reaches no clone, no CI runner` | Pre-existing prose this card did not author: the line is re-added whole only because a line-granularity diff re-adds a line whose other half changed (the § Evidence export bullet, M4). The subject is the machine-local scratch directory `.moai/state/verify/<session>/`, not a card-report artifact, and the sentence states that directory's design intent — outside REQ-AEC-012's bind, which covers sentences *this SPEC introduces*. |
| E4 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | `it is gitignored, so it reaches no clone, no CI runner` | As E3 — C2 mirror of the same pre-existing sentence. |
| E5 | `.claude/agents/moai/manager-lead.md` | `That directory is scratch and nothing more: it is gitignored, so it reaches no clone, no CI runner` | As E3 — pre-existing prose, subject the machine-local scratch directory, design intent; surfaced only by the line-granularity re-add. |
| E6 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | `That directory is scratch and nothing more: it is gitignored, so it reaches no clone, no CI runner` | As E5 — C2 mirror. |
| E7 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | `That directory is scratch and nothing more: it is gitignored, so it reaches no clone, no CI runner` | As E5 — the machine-emitted copy of E6's source (REQ-AEC-011). It appears in this table only as a recorded matched line and is never edited. |
| E8 | `.moai/docs/audit-artifact-convention.md` | `Ignore-matching does not untrack a file that is already tracked, so a project` | SPEC §C.2's own specified closing paragraph: general git behaviour plus a conditional about *a* project (indefinite), not an assertion about this one — identical in kind to E1/E2 and the class REQ-AEC-012 explicitly permits. |
| E9 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | `Ignore-matching does not untrack a file that is already tracked, so a project` | As E8 — byte-identical mirror (REQ-AEC-009). |

A match outside this table fails the criterion. Adding an entry to the table is a
SPEC edit with its own justification, not a close-time judgement call. The
extension from 2 to 9 entries is such an edit: the seven additions were
re-measured at HEAD `ce6de9407` on 2026-09-22 — the aggregate sweep above
returned exactly these nine lines (`exit=0`), all in REQ-AEC-012-permitted
classes, and the extension was ruled by the operator disposition of the same date
rather than decided at close time.

> **[HARD] The declared false-negative class, stated rather than hidden.** One
> added line in `.gitignore` carries a past-tense record of this repository's
> index — *"the entries already in the index stayed there and had to be removed
> deliberately"* — and **no regex above matches it**, because `stayed there`
> takes no participle from the object list. That line is permitted by
> REQ-AEC-012's past-tense clause, so the miss costs nothing here; what it
> establishes is that this regex is a screen and not a decision procedure. A
> successor card that **narrows REQ-AEC-012's past-tense permission** — i.e.
> widens the prohibition to reach past-tense narration, which v0.3.2 currently
> permits (SPEC §A.3b) — must replace the
> instrument, not extend the alternation — the alternation has now been extended
> twice (verb set, then adverbial) and been defeated a third time.

Paired control — a hedged or obligation form the repair introduces MUST match,
proving the diff pipeline and the regex both reach the added lines:

```bash
git diff -U0 develop...HEAD -- \
  .gitignore \
  .claude/agents/moai/plan-auditor.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' | grep -cEi 'local by design|is a local file|local\*{0,2} record|exported so a later reader'
```

Expected: non-zero.

> **Scope arithmetic — the sweep now covers thirteen files, and the thirteenth is
> where the interesting lines are.** The prose claimed *"all thirteen"* while the
> pathspec listed six explicit paths plus `internal/template/templates` (covering
> plan §C rows #4, #5, #7, #10, #11, #13) — **twelve**. The omitted file was row
> #1, `.gitignore`, which is the one file in the scope table that had already been
> edited: `113e487c2` added 11 comment lines to it, three of which the regex
> alternation touches. The criterion written to enforce REQ-AEC-012 was not
> looking at the only file where REQ-AEC-012-adjacent sentences had actually
> landed. `.gitignore` is included rather than exempted: exempting it would have
> closed the arithmetic gap while leaving the requirement unenforced exactly
> where it was being exercised.
>
> **The adverbial group is the second repair, and it has an observed red.** The
> former pattern required the verb and the participle to be **adjacent**, so one
> intervening adverb defeated it. Three mutants that violate REQ-AEC-012, with a
> control that does not (measured on ugrep 7.8.4 and `/usr/bin/grep` alike):
>
> ```
> $ printf '+The destination is already gitignored in this repository.\n+The verdict file remains currently untracked here.\n+The two entries stayed tracked until removed.\n' > /tmp/aec014-mutants.txt
> $ grep -cEi '<former pattern>' /tmp/aec014-mutants.txt
> 0                              # all three mutants pass the old criterion
> $ grep -cEi '<pattern above>'  /tmp/aec014-mutants.txt
> 3                              # all three now fire
> $ printf '+The destination is gitignored.\n' > /tmp/aec014-control.txt
> $ grep -cEi '<pattern above>'  /tmp/aec014-control.txt
> 1                              # the plain form still fires
> ```
>
> The mutant file above is this criterion's own mutant probe under
> `verification-completeness.md` §2 and is re-run at close time. Note also that
> `stayed` was added to the verb list: the former set carried `stays?` but not
> the past participle.
>
> **The first repair — the verb set — retains its earlier red.** The v0.2.0 form
> enumerated only the copula, and the assertion actually present in that draft's
> own specified wording was `remain tracked`, so the criterion written to catch
> the defect returned 0 against the defect while its control returned 1. The
> pattern above returns `1` against that v0.2.0 string (*"artifacts exported
> before this was written remain tracked with no exception recorded for them"*)
> and `0` against the §C.2 wording specified here.
>
> **`[a-z]+ly` rather than `\w+ly`.** `\w` is a GNU extension; the bracket form
> reads identically on BSD grep, which is what CI runs. Measured: the pattern
> returns `3` on the mutant file and `1` on the control under both flavours.
>
> **Scope is the added lines, not the files.** Every one of the thirteen files
> legitimately contains declarative tree-state sentences this card did not write;
> REQ-AEC-012 binds sentences *this SPEC introduces*, so the diff is the correct
> subject and a whole-file form would be false by construction.
>
> **This criterion is empty before the milestone's commit lands**, for the reason
> AC-AEC-013 states at length: `develop...HEAD` reads commits, not the working
> tree. It closes after the commit, never during the edit.

Maps REQ-AEC-012

### AC-AEC-015 — no repaired document tells its reader to decide ignore status with `-v`

**Given** REQ-AEC-014 and the discriminant in SPEC §A.2a,
**When** the twelve repaired target documents are swept for the verbose
`check-ignore` form,
**Then** none is found.

```bash
grep -rnE 'check-ignore[^`]*-v' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control (reachability) — the specified check the repair preserves MUST
match in the two convention copies, proving the sweep reaches the repaired files:

```bash
grep -c 'ls .moai/reports/' \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
```

Expected: two lines, each ending `:1` or higher.

> **[HARD] Classification: regression-guard, not release-blocking — and it is
> vacuous against the current tree by construction.** Measured before
> implementation, the probe already returns zero: `check-ignore` appears **0
> times** across all twelve target documents, so a green here proves nothing
> about the repair. Read as a pass it would be an empty-swept-set claim
> (`verification-completeness.md` §1.1). What it actually guards is the repair
> **introducing** the form — and that hazard is real rather than theoretical,
> because the predecessor draft specified exactly this form for four of these
> twelve files (SPEC §A.2a).
>
> **The instrument has an observed red, taken against that draft at a pinned
> commit.** The same pattern run over the v0.2.0 specified wording fires, and over
> the wording specified here it does not:
>
> ```
> $ git show 622e25d22:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md > /tmp/aec015-red.md
> exit 0
> $ grep -cE 'check-ignore[^`]*-v' /tmp/aec015-red.md
> 3
> exit 0
> $ sed -n '/^## §C Specified wording/,/^## §D Exclusions/p' \
>     .moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/spec.md \
>     | grep -cE 'check-ignore[^`]*-v'
> 0
> exit 1
> ```
>
> That pair is this criterion's RED-now cell and its green path: `3` against text
> carrying the defect, `0` against the text specified here. Without it the
> criterion would be indistinguishable from a pattern that matches nothing.
> Re-measured at HEAD `113e487c2`.
>
> **[HARD] The RED cell is pinned to `622e25d22`, and the former `HEAD:` form was
> the defect.** `622e25d22` is the commit carrying the v0.2.0 draft (`version:
> "0.2.0"`, verified in that blob). The cell previously cited
> `git show HEAD:…` and recorded `3`; re-executed verbatim on this tree the same
> command returns **`1`**, because `HEAD` was `cf45febae`'s parent when the cell
> was authored and is `113e487c2` now. The surviving single match is AC-AEC-002's
> own explanatory note — so the "red" had degraded into the criterion quoting
> itself. This is the ANCHOR case of `verification-claim-integrity.md` §2.1: the
> ref is the address at which a measurement was taken, not the subject of the
> claim, so R1 (pin the literal SHA) applies. `verification-completeness.md` §2.1
> independently requires the fourth element of a RED cell to be a commit SHA and
> never a branch name. The same pin is applied to spec.md §A.2a, which cited
> `git show HEAD:` for the same purpose.
>
> **Scope is the target documents, never this SPEC's own artifacts.** `spec.md`
> §A.2a and `plan.md` §E both quote the verbose form deliberately — that is the
> discriminant being *stated*, not a reader being *instructed*. A sweep widened
> to the SPEC directory would fail on the section that exists to prevent the
> defect.

Maps REQ-AEC-014

---

## §B Edge cases

| Case | Expected handling |
|---|---|
| A copy's FORBIDDEN wording differs from both variants in SPEC §C.1 | Blocker report, not an improvised third variant — the divergence is a finding about the copy |
| `cmp` reports the convention copies already divergent before the edit | Blocker report; the mirror is then edited section-by-section and the pre-existing divergence recorded, never flattened by copying |
| `make agents-emit-check` fails after `make agents-emit` | Emission is non-deterministic or the source layer is malformed; blocker, never resolved by editing a `.toml` |
| Withdrawing the negation appears to change `git status` under `.moai/reports/` | Stop. The withdrawal untracks nothing (SPEC §A.3), so a status change from it alone is a misreading of the tool or a second act that was not the withdrawal; re-read the output and report before proceeding (REQ-AEC-003) |
| AC-AEC-003's `comm -23` is non-empty — a tracked verdict file is not on the remote | Not a failure of this criterion's requirement, but a finding: §A.3a's predicate would authorize its removal and the predicate is the operator's to apply, not the implementer's. Report it; do not remove it |
| AC-AEC-003's `comm -13` contains exactly the two ruled names (`t1039/verdict.md`, `t1048/verdict.md`) | **PASS.** This is the operator-ruled intentional state (2026-09-22 disposition, SPEC §A.3a) — the two files stay on `origin/develop` and intentionally absent from this branch's index |
| AC-AEC-003's `comm -13` contains any name other than the two ruled names | FAIL. That is the state REQ-AEC-003 forbids creating — an index removal cannot undo a publication. Report as a blocker |
| AC-AEC-014 matches a line that is not in its declared exception set | FAIL, whatever the line's apparent justification. Adding an entry is a SPEC edit with its own reasoning, never a close-time judgement |
| The plan-audit carve-out's behaviour changes after the withdrawal | Blocker. The `.gitignore` notes the ordering was load-bearing; AC-AEC-002's third probe is what detects it |
| A grep positive control returns zero | Every absence claim in the same run is void. Re-derive the instrument; do not report the zeros |
| A phrase this card searches for has been rewritten upstream | The zero is not absence. SPEC §A.6 is the standing instance; re-locate the claim by meaning and re-measure before reporting |

---

## §C Quality gates

- `internal/template/templates/.gitignore` does not appear in the change set — it
  carries no verdict negation and needs none.
- No file under `internal/template/templates/.codex/` appears in the change set as
  a hand edit; its presence is acceptable only as `make agents-emit` output.
- No index change on any path under `.moai/reports/` that has already reached
  `origin/develop`, and no history rewrite on any such path in any case. The two
  index removals of SPEC §A.3a were performed on already-published paths (their
  removal premise was measured against a stale remote-tracking ref, §A.3a), and
  the operator's 2026-09-22 disposition rules the resulting state **final**:
  `.moai/reports/t1039/verdict.md` and `.moai/reports/t1048/verdict.md` stay on
  `origin/develop` and stay intentionally absent from this branch's index.
  Restoration and re-staging of either file is forbidden — REQ-AEC-003's first
  clause (no index alteration on an already-published path) prohibits it in both
  directions — and no further index change under `.moai/reports/` is in scope.
- The card's verdict is written to `.moai/reports/t1059/verdict.md` and **left
  local**. Force-staging it would be the card refuting its own direction; the lead
  reads it on disk, and the worktree is not disposed of until that read has
  happened.

---

## §D Definition of Done

### §D.1 Mandatory

1. AC-AEC-001 through AC-AEC-015 all pass, with each absence claim preceded in the
   same run by its positive control. AC-AEC-015 is a regression-guard and is not
   release-blocking: it is vacuous against the current tree by construction, and
   its verdict is read as "the repair introduced nothing", never as evidence the
   repair worked. The ten criteria that fall outside
   `verification-completeness.md` §2.1's single-invocation form **retain**
   release-blocking status and are recorded as passes: §2.1's undecidable
   disposition triggers on an unreproducible RED, and every RED here
   re-executes on this tree with a firing control. The deviation, its forced
   cause, and the criteria it covers are named in the preamble above rather than
   left to a reader who knows §2.1 to discover the mismatch.
2. `make agents-emit-check` exits 0.
3. The two convention copies are byte-identical.
4. The tracked `.moai/reports/` verdict set is a subset of the `origin/develop`
   verdict set (`comm -23` empty) and the remote-only remainder is **exactly**
   the two ruled removals, `.moai/reports/t1039/verdict.md` and
   `.moai/reports/t1048/verdict.md` — reported as the two set differences rather
   than as counts (AC-AEC-003). Any third name in either direction is a finding,
   and the two directions mean different things — see AC-AEC-003's directional
   reading.
5. The card's verdict artifact exists on disk at `.moai/reports/t1059/verdict.md`
   and has been read by the lead before the worktree is disposed of.
6. AC-AEC-013 and AC-AEC-014 are closed **after** their milestone's commit lands,
   not during the edit: their three-dot diff reads commits and is empty on an
   uncommitted working tree, so a close attempted earlier is refused by §D.3
   rather than passed.
7. AC-AEC-014's match set equals its declared exception set exactly. A match
   outside that table fails the criterion; the table is extended only by a SPEC
   edit carrying its own justification, never at close time.

### §D.2 Evidence format

The card's verdict is written in the five-section evidence-bearing format
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). Every AC above
appears with its command and its verbatim observed output — a summary line is not
evidence. Because the verdict stays local, the verdict file carries the deciding
lines themselves rather than pointing at scratch that resolves nowhere else.

### §D.3 The zero-result rule

[HARD] No AC above may be closed on a zero-count result unless its paired positive
control returned non-zero **in the same run**. Two hazards make this non-optional
in this card specifically: the clause under repair carries `**bold**` runs, so a
literal grep whose pattern crosses a `**` delimiter matches nothing regardless of
the file's content; and a phrase this card searches for has already been rewritten
once by an upstream absorb, producing a clean zero on unrepaired files (SPEC §A.6).
In both cases a clean zero means either the target is absent or the pattern never
reached the file, and only the control separates the two readings.
