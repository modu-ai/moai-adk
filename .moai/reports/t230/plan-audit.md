# SPEC Review Report: SPEC-PRECOMMIT-PRESERVE-001

Card: t230 · Tier M · Iteration: 1/2 (Tier M ceiling)
Audited tree: worktree `.claude/worktrees/t230`, branch `WT-precommit-preserve`, HEAD `7b2f42be0`
(diff vs `294b4b6ab` = the four SPEC files only; every source-file citation in the SPEC is therefore
measurable at HEAD unchanged).

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.875** (Tier M PASS threshold `0.80`, per `spec-workflow.md` § SPEC Complexity Tier)

Reasoning context ignored per M1 Context Isolation. Every figure below was measured in this tree in
this run; nothing was inherited from the dispatch or from the SPEC's own claims.

Two BLOCKING defects (D1, D2) sit on the SPEC's core disclosure surface. Neither is a design failure
— both are unresolved wording, each closable in a paragraph — so the firewall does not force FAIL and
the score clears the tier threshold. They are gated: **D1 and D2 must be closed before M2 lands.**
M1 (the classifier) touches neither and may proceed. A stricter reading — FAIL until amended — is
defensible on D1's severity; the delta is a gate placement, not a disagreement about the finding.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE '^- \*\*REQ-PCP-[0-9]{3}\*\*' spec.md | sort` →
  exactly `001` through `014`, 14 entries, no gap, no duplicate, consistent 3-digit zero-padding.
  The §B presentation order is thematic (006, 014, 001, 002, 005, …), which is a grouping choice, not
  a numbering defect.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-PCP-*` in
  `spec.md §B`); AC entries were graded under Group 4, not here. All 14 match a GEARS pattern:
  Ubiquitous (001, 013, 014), Event-driven (002, 003, 004, 008, 009), Where (005, 010, 011, 012),
  Unwanted / `shall not` (006, 007). See D7 for a modality-precision MINOR on the `Where` uses.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`id`, `title`, `version: "0.2.0"`, `status: draft`, `created`/`updated` ISO dates, `author`,
  `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-separated string),
  plus `tier: M`. No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`).
  Corroborated: `~/go/bin/moai spec lint .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md` →
  `✓ No findings — all SPEC documents are valid`.
- **[PASS] MP-4 language neutrality** — the SPEC reaches template-distributed content
  (`internal/template/templates/.git_hooks/pre-commit`, REQ-PCP-011), so this is not N/A. It passes:
  the `pre-commit.local` delegation is shell-level and language-agnostic, and no requirement names a
  language-specific tool as primary. The existing Go-specific steps in the hook body are already
  guarded by `command -v gofmt` / `command -v go` and are untouched by this SPEC.
- **[N/A] MP-5 D7 cross-SPEC reconciliation** — `grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` across all
  four artifacts returns exactly one id, the SPEC's own. There is no external SPEC reference to
  reconcile, so the D7 verb has no input. N/A per the MP-4 precedent.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. D8 auto-PASS
  per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/`
  → no match (rc=1) across `plan.md`, `research.md` (absent, correct for Tier M), and the rest.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 band | Decisions state rejected alternatives (`spec.md §A.5`); citation convention is explicit (`§A.0`). Two ambiguities, both on core requirements: REQ-PCP-004's "on stderr" (D2) and REQ-PCP-010's silence on the hook's fate (D1). |
| Completeness | 0.85 | 0.75-1.0 boundary | All required sections present; six `### Out of Scope — <topic>` H3s each with specific bullets (`spec.md:280,285,292,297,302,307`); edge-case table `§D.1`; DoD `§D.3`. Deducted for the unanalyzed first-upgrade noise magnitude (D3) and the unspecified backup-succeeded/write-failed state (D5). |
| Testability | 0.85 | 0.75-1.0 boundary | Strongest dimension. Every AC carries Decides / Baseline / Mutant / Failing-input. The headline mutant is constructed once and jointly defeated by AC-PCP-004/006/007. AC-PCP-006 correctly demands both artifacts *within one case*; AC-PCP-013 correctly rejects SKIP; AC-PCP-014 correctly mandates both cases. Deducted for AC-PCP-010's missing post-state assertion (the hole D1 rides through) and AC-PCP-002's imprecise falsifiability label (D6). |
| Traceability | 1.00 | 1.0 | `grep -cE '^### AC-PCP-' acceptance.md` → `14`; REQ definitions → `14`; `spec.md §E` matrix is 1:1 across `001`-`014`. No orphaned AC, no uncovered REQ. |

Aggregate = (0.80 + 0.85 + 0.85 + 1.00) / 4 = **0.875**.

---

## Verification of the dispatch's eight attack points

Recorded because several were checks of the SPEC's own claims, and a claim verified is evidence
while a claim assumed is not.

1. **Core mutant** — CONSTRUCTED AND DEFEATED. An implementation that overwrites correctly, backs up
   correctly and prints nothing fails AC-PCP-004 (output must contain all three of: backup path,
   replacement statement, `pre-commit.local`), AC-PCP-006 (both artifacts in one run) and AC-PCP-007
   clause (ii). No AC in the set asserts post-state file contents *only*. The SPEC's claim here holds.
2. **Falsifiability** — 11 criteria fail against the untouched tree; the 3 labelled
   behaviour-preserving (008, 012, 013) are honestly labelled — each restates behaviour I measured as
   already true at HEAD. One label imprecision found (D6), not a vacuous criterion.
3. **Three-way attribution / migration** — the design is sound; its *magnitude* is not analyzed (D3).
   The record is written on both call paths (both callers funnel through
   `installPreCommitHookOptional` → `InstallPreCommitHook`, verified at
   `update_template_sync.go:575` and `init.go:898`), and once-per-clone is genuinely once.
4. **Decision 2 / unrecoverable-loss path** — a real hole found (D1).
5. **Paired-edit constraint** — VERIFIED. `TestPreCommitTemplateMatchesConstant` exists at
   `internal/cli/hook_install_precommit_test.go:38`, with the `t.Skipf` escape at `:45` exactly as
   the SPEC describes. The paired edit is bound by a requirement (REQ-PCP-013) and an AC
   (AC-PCP-013), not by prose alone; AC-PCP-013 correctly requires PASS and rejects SKIP.
6. **t237 / #1641 collision** — ACTIONABLE, not decorative. `git log origin/main --oneline -200 |
   grep -iE '1641|t237'` → no match: t237 has **not** landed, so the collision is live. `plan.md §C`
   step 4 ("check whether card t237 has landed; if it has, rebase before starting M3") is a real
   pre-flight, and `§C.2`'s "land REQ-PCP-011 as its own commit, last" is carried into the DoD.
7. **Scope / counts + requirement C** — counts verified (14/14, above). The `exec` sweep was **re-run
   in this tree** with the SPEC's recorded command:
   `grep -rEn 'exec [0-9][0-9]*[<>]' --include='*.sh' --include='*.tmpl' --include='pre-commit'
   --include='pre-push' --include='*.go' --include='*.md' . | grep -v '^\./\.git/'` → **0 hits**
   (rc=1). The closure is correct and is recorded as a measured closure in `§A.4`, not omitted.
   Caveat raised as D4.
8. **Citation integrity** — the v0.2.0 retraction is **accurate**, verified against the object
   database without touching the primary checkout: `git show a1b1ca696:internal/cli/init.go |
   sed -n '773p'` → `installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"),
   cmd.ErrOrStderr())`, and `:557` of `update_template_sync.go @a1b1ca696` → the sibling call. Both
   card citations were correct for their tree; nothing had drifted; the retraction says so. Every
   remaining `file:line` in the SPEC carries a `@<sha>` anchor, and I re-measured the whole `§A.2`
   baseline table at HEAD: backup `0`, sha256 `0`, `pre-commit.local` `0`, template `3245` B,
   installer `179` lines, call sites `575` / `898` — **all seven values reproduce exactly**.

---

## Defects Found

**D1 — `.moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md` REQ-PCP-010 ⊥ REQ-PCP-006 — a backup-write failure may destroy the user's patch unrecoverably, and no requirement or criterion forbids it — Severity: critical — Class: blocking**

REQ-PCP-006 forbids replacing a user-modified hook "without producing **both** a backup file and a
notice naming it". REQ-PCP-010 says that where the backup write fails, the installer "shall report a
warning and shall not fail `moai init` or `moai update`" — and says nothing about whether the
overwrite still proceeds. AC-PCP-010's Then clause is "returns normally, emits a warning naming the
failure, and does not panic or abort the caller" — no assertion on the hook's post-state.

Two implementations conform to REQ-PCP-010 and AC-PCP-010:

- backup fails → warn → **skip** the overwrite → the patch survives;
- backup fails → warn → **overwrite anyway** → the patch is destroyed with no recoverable backup.

The second violates REQ-PCP-006, and nothing catches it: AC-PCP-006's Given is "the AC-PCP-003
scenario", i.e. a *successful*-backup run, so the failure path is outside every criterion that
asserts the invariant. AC-PCP-010 names two mutants ("never attempts a backup", "swallows the failure
with no warning"); the destroying mutant is not among them.

This is the card's own failure mode — an unrecoverable silent-ish loss — surviving inside the
requirement written to prevent it. Realistic triggers: read-only `.git/hooks/`, quota exhaustion, an
`os.Link`/`os.Rename` backup strategy failing across a filesystem boundary while the 3,245-byte
overwrite succeeds.

**Required fix**: make REQ-PCP-010 state the precedence explicitly — e.g. "…shall report a warning,
shall **not overwrite the hook**, and shall not fail `moai init` or `moai update`" — and add the
corresponding clause to AC-PCP-010's Then ("…and the hook's bytes are unchanged"), with a matching
mutant ("overwrites anyway after a failed backup"). If the opposite precedence is genuinely intended,
say so in REQ-PCP-006 as a named carve-out with its rationale; silence is what makes this blocking.

**D2 — `spec.md` REQ-PCP-004 ("on stderr") ⊥ `plan.md §A` ("Neither caller changes in this SPEC") ⊥ `acceptance.md` AC-PCP-004 / AC-PCP-007 (single captured writer) — the disclosure requirement is unsatisfiable as written — Severity: critical — Class: blocking**

Measured: `installPreCommitHookOptional(projectRoot string, skip bool, out io.Writer)` has exactly
one writer. At the `moai update` call site that writer is **stdout** —
`internal/cli/update_template_sync.go:69` is `out := cmd.OutOrStdout()`, passed through to `:575`. At
the `moai init` call site it is stderr (`init.go:898`, `cmd.ErrOrStderr()`). So under `moai update`
the function cannot emit "on stderr" without either a caller change — which `plan.md §A` forbids —
or writing to `os.Stderr` directly, which breaks AC-PCP-004 ("runs against a captured writer") and
AC-PCP-007 clause (ii) ("the captured output … contains the backup path"), both of which assert
against `out`.

The likely run-phase cascade is the concerning part: an implementer follows REQ-PCP-004 literally,
AC-PCP-004 then fails, and the cheapest repair is to weaken the criterion to capture `os.Stderr` —
which also detaches AC-PCP-007 from its "differs from the bare success line on the same writer"
comparison. That is criterion erosion driven by an unresolved contradiction, in exactly the SPEC
whose discipline is built to prevent it. (`errOut := cmd.ErrOrStderr()` does already exist at
`update_template_sync.go:72`, so the caller-change route is cheap — it is simply forbidden by the
plan as written.)

**Required fix**: pick one and make all three surfaces agree. Either (i) drop "on stderr" from
REQ-PCP-004 and require the notice on the installer's existing writer — the minimal edit, consistent
with the AC set as written; or (ii) keep stderr, and amend `plan.md §A` to permit the two-line caller
change plus add a second writer parameter, updating AC-PCP-004/007 to name which writer each clause
inspects.

**D3 — `spec.md §A.5` (Decision 1, migration case) + `plan.md §F` (M3 ordering) — the first upgrade produces a false alarm for ~100% of the installed base, and the rejection of the cheap mitigation is argued against a strawman — Severity: major — Class: blocking**

REQ-PCP-005 treats "no provenance record" as user-modified whenever the installed bytes differ from
the incoming content. Every existing installation reaches the new code with no record. If M1+M2+M3
ship in one release, M3 (REQ-PCP-011) *changes the hook body by construction*, so `installed !=
incoming` holds for **every** user — including every user who never touched the hook. Each therefore
receives a backup file plus a notice that, per REQ-PCP-004, tells them to re-apply local checks in
`pre-commit.local` — checks they never had.

`§A.5` acknowledges "one backup and one notice it did not strictly need, once", but frames it as
incidental ("an unmodified legacy hook that *happens* to differ"). It does not: with M3 in the same
release it is certain and universal. That is the alarm-fatigue outcome `§A.3` names as the reason
the three-way design exists — reproduced in full, once, by the release that introduces the design.

The rejection reasoning is also incomplete. `§A.5` rejects "compare against every historically-shipped
hook version" for requiring "an ever-growing corpus". It never evaluates the cheap variant: a
**one-entry** corpus — the digest of the immediately-previous shipped `preCommitHookContent`, a
single 64-character constant — which collapses the first-upgrade false-positive rate from ~100% to
~0% at no ongoing cost. Rejecting the expensive form of an alternative is not a rejection of its
cheap form.

`plan.md §F` compounds this: it says M3 must either land before the SPEC closes, or the notice names
a facility users cannot yet use. Landing M3 in-release is what creates the 100% noise; deferring it
is what makes the notice misleading. The two stated resolutions conflict and neither is chosen.

**Required fix**: evaluate the one-entry-corpus variant explicitly in `§A.5` — take it or reject it
on its own terms — or split the release so M1/M2 ship with the hook body unchanged (making
`installed == incoming` for unmodified users, silent, correct) and M3 follows once records exist.
Add a criterion or a DoD line that measures the first-upgrade path with an unmodified legacy hook
that equals the *previous* shipped content; today no criterion covers it (AC-PCP-005 sub-case (b)
requires the legacy hook to equal the *incoming* content, which is the case M3 makes impossible).

**D4 — `spec.md §A.4` — the recorded sweep command is correct but a reader repeating it under a looser reading gets a non-empty result — Severity: minor — Class: optional**

The SPEC records `exec [0-9][0-9]*[<>]` (one or more digits) → 0 hits, which I reproduced exactly.
The card's axis is elsewhere written `exec [0-9]*[<>]` (zero or more), which in this tree returns 6
hits — all prose or comments about `printf … | exec <bin>` (command replacement) in
`handle-stop-goal.sh`, its `.tmpl` twin, `stop_goal_single_exec_test.go`, and two report/spec files.
None is a file-descriptor redirection. The SPEC's pattern is the semantically correct one; the risk
is that a later reader repeating the "same" sweep with the looser glob concludes the closure was
wrong.

**Required fix**: one clause in `§A.4` naming what the pattern means — file-descriptor redirection,
one or more digits — and noting that the zero-digit form matches unrelated `exec <command>` prose.

**D5 — `acceptance.md §D.1` (edge cases) — the backup-succeeded/hook-write-failed state is unspecified — Severity: minor — Class: optional**

The edge table covers a missing hooks directory, `skip=true`, a malformed record, an unreadable hook,
same-second backup collisions, and a hand-restored hook. It does not cover: the backup is written,
then the hook write fails. The result is an orphan `pre-commit.bak.<ts>` beside an unchanged hook,
while REQ-PCP-004's notice states the hook "was replaced" — which would be false. Harmless to data,
but it produces a misleading notice on the SPEC's own disclosure surface.

**Required fix**: one row in `§D.1` stating the expected outcome (most plausibly: no replacement
notice when no replacement occurred; the warning path of REQ-PCP-010 covers the failure).

**D6 — `acceptance.md` AC-PCP-002 + `progress.md` §E.1 — the "eleven fail against the untouched tree" claim is slightly overstated — Severity: minor — Class: optional**

AC-PCP-002's Given is "an installed hook whose bytes match the recorded digest". No sidecar mechanism
exists at `294b4b6ab`, so the precondition is unconstructible: the criterion goes red at setup, not
by observing today's behaviour. Its own Baseline field concedes as much ("Today's installer reaches
this outcome by accident … not by decision"), and its stated failing input is correctly an
*implementation* mutant — the same class as the three criteria labelled behaviour-preserving. The
labelling, not the criterion, is imprecise; `progress.md`'s 11/3 split reads as stronger evidence
than it is.

**Required fix**: either label AC-PCP-002's falsification target as implementation-mutant-only (as
the behaviour-preserving three are), or restate `progress.md`'s count as "eleven are falsifiable —
ten against the untouched tree, one against a named implementation mutant".

**D7 — `spec.md §B` (REQ-PCP-005, -010, -011, -012) — `Where` used for a runtime state rather than a capability gate — Severity: minor — Class: optional**

GEARS reframes `Where` as a capability gate / feature flag / static configuration. These four use it
for runtime state ("Where no provenance record exists", "Where a backup … write fails", "Where
`.git/hooks/pre-commit.local` exists and is executable"), which reads more naturally as `While` or
`When`. Each REQ still matches a GEARS pattern syntactically and none is ambiguous, so MP-2 passes;
this is precision, not compliance.

**Required fix**: optional — swap `Where` for `When` / `While` on those four.

---

## Recommendation

The design is sound and the criteria are, on the whole, unusually well constructed: the mutant
discipline is real rather than ceremonial, AC-PCP-006 and AC-PCP-014 are each built to defeat a
specific plausible-but-wrong implementation, and AC-PCP-013's PASS-not-SKIP requirement closes a live
hazard I confirmed exists at `hook_install_precommit_test.go:45`. The v0.2.0 retraction is honest and
I verified it against both trees. Traceability is complete.

Three things must change before implementation reaches the disclosure surface:

1. **Close D1** — state in REQ-PCP-010 whether the overwrite proceeds when the backup write fails,
   and assert the hook's post-state in AC-PCP-010. Until this is written down, a conforming
   implementation can destroy the exact data this SPEC exists to protect. Gate: before M2.
2. **Close D2** — reconcile REQ-PCP-004's "on stderr" with `plan.md §A`'s no-caller-change constraint
   and with AC-PCP-004/007's single captured writer. Measured: `moai update` passes
   `cmd.OutOrStdout()`. Gate: before M2.
3. **Close D3** — evaluate the one-entry previous-digest corpus in `§A.5` on its own terms, or split
   M3 into a later release, and add a criterion covering an unmodified legacy hook equal to the
   *previous* shipped content. Gate: before M3, or before the release plan is fixed — whichever comes
   first.

D4-D7 are optional and can be folded into the next edit of these artifacts or left; none affects
correctness, and routing all four into a revision would buy documentation polish at the cost of
another audit cycle.

M1 (the three-way classifier) is unaffected by all three blocking defects and may proceed on the
current artifacts.
