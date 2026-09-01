# SPEC Review Report: SPEC-VERSION-STAMP-PREDICATE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.79** (Tier M PASS threshold 0.80 — `spec-workflow.md` § SPEC Complexity Tier)

Audit tree: worktree `.claude/worktrees/t392`, HEAD `9a3e2dabe`, branch `WT-version-stamp-predicate`
(`git rev-parse --show-toplevel` / `--short HEAD` / `git branch --show-current`, run in this session).

Reasoning context ignored per M1 Context Isolation. The orchestrator's framing was read only to
locate artifacts; every judgment below rests on the artifact files and on commands re-run here.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n '^\*\*REQ-'` over `spec.md` returns exactly
  REQ-VSP-001…011 at L180, 184, 188, 191, 195, 200, 204, 209, 214, 218, 221. Sequential, no gap,
  no duplicate, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md` §4), not the verification layer. All 11 match a GEARS pattern: ubiquitous
  (001/002/005/006/007/008/010/011), event-driven (003 "When the sweep finds…", 004 "When a
  registry entry…", 009 "When the two sets differ…"), with canonical negative forms present
  (002 "shall contain no glob…", 004 "shall not be judged", 006 "shall not admit"). The
  `Given/When/Then` entries in `acceptance.md` are AC-layer and are graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md` L2-14) plus optional `tier: M`. `version: "0.1.0"` quoted semver; `status: draft` in
  the 8-value enum; `created`/`updated` ISO; `priority: Medium` in the allowed enum;
  `lifecycle: spec-anchored`; `tags` comma-separated string; `phase: "v3.1.5 target"` is a release
  label, not a lifecycle stage. No rejected snake_case alias. Independently confirmed:
  `moai spec lint` → `0 error(s), 1096 warning(s)`, rc=0, and
  `grep -c 'VERSION-STAMP-PREDICATE' <lint output>` → **0** (no finding of any severity against
  this SPEC).
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own Go code, its
  docs-site and its release stamps. It makes no multi-language tooling claim, so the 16-language
  enumeration criterion does not bind. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md`
  yields exactly two IDs: itself and `SPEC-VERSION-STAMP-GUARD-001`.
  `grep -n '^status:' .moai/specs/SPEC-VERSION-STAMP-GUARD-001/spec.md` → `5:status: completed`.
  `completed` ∉ {retired, superseded, archived} ⇒ no reconciliation obligation, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/`
  → rc=1, no match. `research.md` is absent, which is correct for Tier M (3-artifact set).

The must-pass firewall does **not** force this verdict. The FAIL comes from the rubric and from
three blocking correctness defects below.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 band — minor ambiguity a reasonable engineer would resolve, but two decisions are stated in mutually incompatible terms | The sweep mechanism is `filepath.WalkDir` (`plan.md:86`) while every number the SPEC stands on is a `git grep` measurement (`spec.md:90,94,117`; `plan.md:37`) — D1. §3's 2×2 attributes the whole registered/stale cell to REQ-VSP-004, which `spec.md:191-194` explicitly exempts `prose` from — D3. |
| Completeness | 0.70 | between the 0.50 and 0.75 bands — all sections present and frontmatter complete, but the registry's maintenance contract is absent, not merely unmeasured | All required sections present: HISTORY (L18), §1 WHY, §2/§3 WHAT, §4 REQUIREMENTS, §4.1 + `acceptance.md` §D ACCEPTANCE, §7 with five `### Out of Scope — <topic>` H3 sub-headings each carrying `-` bullets. Missing: who edits the registry, when, and what the check does between a bump and that edit — D2. |
| Testability | 0.70 | between the 0.50 and 0.75 bands — every AC is binary and carries a pre-pinned failure string, but the central constant is not attributable to the population the check will actually read | `acceptance.md` splits "측정됨(plan)" from "RED 대기(run)" and `plan.md` §D.2 fixes the expected failure strings before measurement — genuinely strong. But AC-VSP-002/003/005 all judge against `28`, and 28 was derived from the tracked-file population, not the walked one — D1. |
| Traceability | 1.00 | 1.0 band | `spec.md` §4.1 maps REQ-VSP-001…011 to AC-VSP-001…011 one-to-one; `grep -n '^### AC-' acceptance.md` returns exactly 11 headings in the same order; `moai spec lint` emits **no** `CoverageIncomplete` finding for this SPEC ID. No orphan AC, no uncovered REQ. |

Arithmetic mean = (0.75 + 0.70 + 0.70 + 1.00) / 4 = **0.7875 → 0.79**, below the Tier M
threshold of 0.80.

---

## What I re-measured and found sound

Recording these explicitly, because a FAIL verdict that does not distinguish the verified parts
from the defective ones is not usable as a fix route.

**The exclusion attribution (§2.3) is exact and the partition is disjoint.** Re-run here at
`9a3e2dabe`:

```
git grep -lF v3.1.3 -- . | wc -l                                   → 155
git grep -lF v3.1.3 -- . ':!.moai/reports' ':!.moai/specs' \
  ':!.moai/release-notes' ':!CHANGELOG.md' ':!*_test.go' \
  ':!docs-site/content/*/changelog*' | wc -l                        → 34
.moai/reports 57 · .moai/specs 58 · .moai/release-notes 1 · CHANGELOG.md 1 · *_test.go 4 · changelog-pages 0
```

Sum 121; `121 + 34 = 155` ✓. Disjointness verified separately rather than inferred from the
identity — the union of the six groups measured as a single pathspec is also **121**
(`git grep -lF v3.1.3 -- <all six> | sort -u | wc -l`), so no file is double-counted and the
identity is not an overlap cancelling a gap. Each excluded group carries its own distinct reason
in the §2.3 table; none of the six is a blanket clause wearing several labels. The t388 first
anti-pattern is not repeated here.

**The 34-file classification is exact.** Listing the 34 confirms 7 stamps (README×4 +
`.moai/config/sections/system.yaml` + `docs-site/hugo.toml` + `pkg/version/version.go`), 6 goldens,
20 docs-site prose pages (5 documents × 4 locales), and `.moai/docs/version-management.md`.
7 + 6 + 20 + 1 = 34.

**The `-h` trap measurement (§2.4) holds.** Eight name-token files inside the deny-list scope, six
of them under `.moai/release/` — one character from the excluded `.moai/release-notes/` — and
`git grep -lF v3.1.3 -- .moai/release/` → 0. Verified by listing all eight.

**§1.1 is consistent with the corrected baseline, not the original premise.** `baseline.md` §B.3's
`[정정 2026-09-01]` block retracts the "reachable RED candidate" reading of
`origin/release/v3.1.4`. `spec.md` §1.1, `plan.md` §D, and `progress.md` §E.1 all carry the
corrected account — the authoritative token there is `v3.1.4`, the stale goldens carry `v3.1.3`, so
the current-token sweep is silent on that tree. The SPEC does not re-import the wrong premise
anywhere. This is the single strongest thing in the document.

**The ordering constraint is real and names its tree.** `plan.md` §D pins the RED tree by SHA
(`9a3e2dabe` and its M1-M3 successors, pre-M4), forbids borrowing (AP-4, AP-6), and separates
AC-VSP-005's synthetic empty-sweep RED from M3's count-mismatch red. `plan.md` §D.1 also refuses to
make CI the RED observer. This satisfies t388 §5.1.

**Tier M derivation is sound and is not reverse-engineered.** §5 explicitly invalidates the card's
595-file basis and re-derives from 10 touched files (1 new check + 2 pin-edited tests + 6
regenerated goldens + 1 doc section). The SSOT column is "Files affected", which plainly counts
files in the diff, so counting the goldens is correct and the SPEC's rejection of the
data-not-files reading is argued rather than assumed. Note also that the disputed reading would
land Tier S, whose PASS threshold is *lower* (0.75) and whose artifact set is *smaller* — so the
author chose the stricter classification, which is the safe direction even if one disagrees.

**Every source citation I spot-checked resolves.** `pkg/version/version.go:8` = `Version = "v3.1.3"`;
`version_sync_list_test.go:59` = `expectedVersionStampEntries = 7`, `:106` `parseVersionStampEntries`,
`:142` `isBoldLabel`; `hook_flush_test.go:22` `repoRootFromCLITest`; `version_test.go:180` the pin
precedent with `defer` restore; `docs-site/hugo.toml:55-56` the version/date pair; `ci.yml` ~208
`go test … ./...`. Each golden carries exactly one `v3.1.3` occurrence, as `plan.md` M4 claims.

---

## Defects Found

### D1 — the sweep reads a population the constants were never measured against
**Artifact:** `plan.md:86` (mechanism) vs `spec.md:90,94,117` and `plan.md:37` (measurements);
`acceptance.md:33-52` (the `28` constant) — **Severity: critical — Class: blocking**

`plan.md:86` specifies the driver as `filepath.WalkDir` excluding `.git/` plus the six literal
groups. Every number the SPEC holds — 34, 28, 121, 155, the per-group counts, the `-h` trap's 8 —
was produced by `git grep`, which sees only **tracked** files. `filepath.WalkDir` sees the
filesystem, including everything `.gitignore` hides. The two populations are not the same set, and
REQ-VSP-010 forbids the check from asking git which files are tracked, so the check cannot close
the gap at runtime.

What I ran:

```
grep -rlF --exclude-dir=.git v3.1.3 . | wc -l          → 161      (git grep: 155)
```

Six token-carrying files exist on disk in this worktree that `git grep` does not see. Today they
all happen to fall inside excluded groups, so the post-deny-list count coincides at 34 — I verified
that by listing the filesystem survivors, and they are the same 34 paths. That coincidence is luck,
not a property the SPEC establishes, and it does not survive leaving this worktree:

```
ls -1d /Users/goos/MoAI/moai-adk-go/.claude/worktrees/*/ | wc -l   → 181
grep -n 'worktrees' .gitignore                                     → 200:.claude/worktrees/
grep -c 'v3.1.3' .../.claude/worktrees/t392/pkg/version/version.go → 1
```

In the primary checkout the walk descends into 181 gitignored worktree copies, each a full tree
carrying the authoritative token. REQ-VSP-003 would name hundreds of unregistered paths and the
check would be unrunnable locally while staying green in a clean CI checkout — a
measurement-tree ≠ judgment-tree split of exactly the kind this SPEC's own §8 R-7 warns about,
introduced by the SPEC rather than inherited. `bin/`, `dist/`, `/docs-site/public/`,
`node_modules/`, `.moai/cache|logs|state/` are gitignored on the same terms (`.gitignore` lines
12, 27, 295, 274, 306-308) and would join the sweep after any local build.

The author was aware the mechanism is a walk — `spec.md:151` argues REQ-VSP-006 exists because
"`filepath.WalkDir` 에서 경로가 손 닿는 곳에 있어" — yet never reconciled the walked population
with the measured one.

**Required fix:** state the sweep's file population as a requirement, and make the constants
attributable to it. Either (a) enumerate the walk-pruning set to cover the gitignored roots
(`.claude/worktrees/`, `.moai/worktrees/`, `bin/`, `dist/`, `docs-site/public/`, `node_modules/`,
`.moai/{cache,logs,state}/`) as literal groups in §2.3 with their own measured counts, and re-derive
34/28/121 with `grep -rl` rather than `git grep`; or (b) narrow the walk to an explicit root
allow-list. Then re-run §2.3's attribution against the chosen population and record the command
that produced each number.

### D2 — a version bump breaks the check wholesale, and the SPEC names no maintenance contract
**Artifact:** `spec.md` §8 R-4 (L…, "R-4 — 등록부 유지비를…"); `spec.md` §6 reason 1;
`plan.md` §D.3 — **Severity: critical — Class: blocking**

21 of the 28 registry entries are `prose`: 20 docs-site narrative pages plus
`.moai/docs/version-management.md`. Whether a bump updates them is decidable, and I decided it:

```
git show --numstat --format='%h %s' 61921f1ba
  → .moai/config/sections/system.yaml, README.{md,ko,ja,zh}, docs-site/hugo.toml, pkg/version/version.go
git show --numstat --format='%h %s' eba919e44
  → the same set minus hugo.toml (that is the miss t388 documented)
```

A bump touches **only the 7 stamps**. The 20 docs-site prose pages and the doc keep the old token.
So the instant `pkg/version/version.go` moves to `v3.1.5`, the authoritative-token sweep matches
**7** files, not 28, and REQ-VSP-005 fails with `sweep matched=7 expected=28` — a bare count with
no path named, on the single most routine event in this check's life. Every release then requires a
human to rewrite a 28-entry Go literal and both constants, and nothing tells them to:
`spec.md` §6 reason 1 deliberately keeps the 21 prose paths **out** of
`.moai/docs/version-management.md`'s "Files Requiring Version Sync", which is the only list a
person performing a bump reads.

§8 R-4 records the opposite, milder risk — "how many prose paths get refreshed to the new token,
creating new *unregistered* items" — and marks it unmeasured. The measurement was one `git show
--numstat` away, and §2.3 already cites that very numstat for CHANGELOG.md. The answer is *none of
them get refreshed*, which produces collapse rather than churn.

Note this is not the *unattributable* kind of gap: the SPEC does not say "we did not measure the
maintenance cost", it says something about the maintenance cost that the measurement contradicts.

**Required fix:** add a requirement and an AC for the registry's maintenance contract — who edits
it, at what point in the bump sequence, and what the check reports during the window. Add the
registry file to the bump instruction list in `version-management.md` (naming the file, not the
21 paths, so §6 reason 1 survives). Rewrite §8 R-4 to record the measured numstat result and the
sweep-collapse consequence. Consider judging prose entries by "carries *some* `vX.Y.Z` token"
rather than the authoritative one, so the count is stable across a bump — but that is a design
choice for the author, not a fix I am prescribing.

### D3 — the 2×2 claims a cell that REQ-VSP-004 covers for only 7 of 28 entries, and the
classification predicate is undefined
**Artifact:** `spec.md` §3 table (registered × stale cell) vs `spec.md:191-194` (REQ-VSP-004);
`spec.md:184-187` (REQ-VSP-002) — **Severity: critical — Class: blocking**

§3's grid says the registered/stale cell is caught by REQ-VSP-004. REQ-VSP-004's own second
sentence reads: "Entries classified `prose` shall not be judged for freshness." 21 of 28 entries
are prose. So the cell is closed for 25% of the registry and open for the rest, and the grid — the
document's central coverage claim — says otherwise without qualification.

The grid's two axes are also the wrong two. Coverage is decided by (registered, **classified
stamp**, carries token), a three-way split; collapsing classification into registration is what
produces the overstatement. This leaves a fifth case entirely outside the drawn grid and outside
the named fourth quadrant:

> A file that is genuinely a stamp site, entered in the registry as `prose`, and absent from
> `version-management.md`'s Version Stamps list.

REQ-VSP-003 passes (it is registered). REQ-VSP-004 skips it (classified prose). REQ-VSP-009 passes
(it compares the stamp set to the doc list, and it is in neither). It is invisible to all three
assertions, and the next bump breaks it — **which is precisely the `hugo.toml` shape t388 existed
to close**, reintroduced through the classification axis. REQ-VSP-002 requires each entry to be
classified but defines no predicate for choosing, so the classification is unconstrained judgment
on which the whole freshness assertion's reach depends.

The grid additionally does not state its population. Files inside the six excluded groups are in
neither row (see also §8 R-2's `_test.go`-inline hole, which the SPEC itself names).

**Required fix:** redraw §3 over three axes (registered / classified / carries-token), or state the
grid's population and split the registered row by classification. Add a definition of `stamp` vs
`prose` to REQ-VSP-002 that a reader can apply to a new file — the operative property is
"a bump must rewrite this file or a test breaks", which is exactly what
`red-observation.md` §R.3 demonstrated for the goldens. Name the misclassification case in §7 or
§8 alongside the fourth quadrant.

### D4 — the replacement wording's closed "셋" enumeration understates, repeating the exact t388
failure mode
**Artifact:** `plan.md` §E replacement text (the "여전히 잡지 못하는 것이 셋이다" paragraph);
`acceptance.md` AC-VSP-011(b) — **Severity: major — Class: blocking**

The narrowing itself is defensible and I grade it as earned: the replacement correctly writes
"등록된 **스탬프**가 권위 토큰을 담지 않은 것" rather than the looser "등록된 것", and it retains
the "목록이 더는 썩지 않는다는 뜻이 아니다" sentence (verified: `grep -n '썩'` finds it at
`plan.md:192` and no "no longer rots" claim anywhere in `spec.md`).

The defect is the **closed count**. "셋이다" asserts the uncovered set has exactly three members.
It omits at least two the SPEC has already identified or that follow from its own design:

- the `*_test.go`-inline stamp site — `spec.md` §8 R-2 names it explicitly and marks it
  unclosed, and the doc's enumeration drops it;
- the misclassified-as-prose stamp site (D3).

A document that repairs an overclaim by publishing a smaller, differently-shaped overclaim has not
repaired it. This is the failure the whole card exists to correct, so it blocks.

**Required fix:** replace "셋이다" with an open form ("적어도 다음이 남는다" / "among the cases
still not caught are"), and add the `_test.go`-inline site and the misclassification site to the
list. Extend AC-VSP-011(b) with grep assertions for the added items so the enumeration is checked,
not just written.

### D5 — M5 deterministically breaks the GREEN that M4 establishes
**Artifact:** `plan.md` §C M5 + §E; `.moai/docs/version-management.md:90` —
**Severity: major — Class: blocking**

`grep -n 'v3\.1\.3' .moai/docs/version-management.md` returns **exactly one line**: L90 — and L90
is the very sentence `plan.md` §E replaces. The pinned replacement text contains no version token
(I read it; it names `hugo.toml`, `releaseDate`, `system.yaml.tmpl`, and no `vX.Y.Z`). After M5 the
doc drops out of the sweep: sweep 27, registry 28, expected 28 → REQ-VSP-005 fails at the end of
the SPEC's own last milestone. The plan's ordering section (§D) reasons carefully about the M3→M4
boundary and does not consider M4→M5 at all. §8 R-5 describes this failure shape but frames it as
a future hazard, not as something the plan schedules.

**Required fix:** decide the doc's registry status before M5 — either require the replacement text
to carry the authoritative token, or drop `.moai/docs/version-management.md` from the registry and
set both constants to 27, or exclude the doc from the sweep as a seventh literal group with its own
reason. Whichever, state it in §D as an ordering constraint and adjust §D.3's constant table.

### D6 — six volatile counts are written into a document with no staleness contract and no check
**Artifact:** `spec.md` §2.3 table; REQ-VSP-007 (`spec.md:204-208`); AC-VSP-007
(`acceptance.md:128-143`) — **Severity: major — Class: optional**

REQ-VSP-007 obliges the doc to carry the measured file count per excluded group. Two of the six
counts move constantly: `.moai/reports` measured 57 here, and it includes files written by this
session — `baseline.md` and `red-observation.md` — and **this audit report adds another**, so 57 is
stale before run-phase starts. AC-VSP-007's judgment command checks only that the table's row count
equals the group count; nothing checks the numbers. A card whose subject is documentation rot is
therefore specifying six new unchecked, rotting numbers.

The `changelog-pages` clause is handled better than the task's worry suggests — §2.3 does record it
as "0을 감추는 조항" rather than leaving it unmeasured — but it stops there. Nothing says what
happens when that clause starts hiding something: no re-measure trigger, no owner, and AC-VSP-007
would not notice, because the row would still be present.

**Required fix:** pin each count to the measuring tree SHA and label them point-in-time
(`9a3e2dabe: 57`), or record only the stable groups' counts and state a range for the volatile
ones. Add one sentence to §2.3 saying that a nonzero measurement on `changelog-pages` (or any group
whose reason no longer matches its count) obliges a re-derivation of the exclusion table.

### D7 — the 28-path registry has no ghost-entry guard; t388's 7-entry list does
**Artifact:** `spec.md` §6, §7 "Out of Scope — t388이 세운 검사"; REQ-VSP-002 —
**Severity: minor — Class: optional**

t388's landed check exists precisely because a list can name a path that no longer exists
(`version_sync_list_test.go:82` compares parsed entries against `expectedVersionStampEntries`, and
the surrounding test asserts each named path is in the tree). The new 28-path registry inherits no
such assertion: REQ-VSP-002 constrains the entries' *shape* (exact path, no glob) but never
requires them to resolve. A deleted or mistyped `prose` path is not silent — the sweep drops to 27
and the count assertion fires — but it fires with a bare count and no path, and the cheapest human
response is to decrement the constant to 27 and leave the ghost in place. That is a degraded
diagnosis of the exact defect class the predecessor card landed a check for.

**Required fix:** add to REQ-VSP-002 (or a new AC on REQ-VSP-002) that every registry entry must
resolve to a file in the working tree, failing with the offending path named. One `os.Stat` per
entry; no new dependency.

### D8 — a requirement names an implementation identifier
**Artifact:** `spec.md:209-213` (REQ-VSP-008) — **Severity: minor — Class: optional**

REQ-VSP-008 prescribes pinning `version.Version`, a package variable name, inside a requirement
statement. Group 3 RQ-4 asks requirements to express WHAT, not HOW. The behavioural form —
"the golden-fixture tests shall render a fixed version token independent of the build-time value"
— says the same thing without binding the requirement layer to an identifier. Low impact; the
mechanism is genuinely settled by an in-repo precedent (`version_test.go:180`), and `plan.md` M4 is
the right home for the identifier.

**Required fix:** optional. Restate REQ-VSP-008 behaviourally and leave `version.Version` in
`plan.md` M4 and AC-VSP-008.

---

## Regression Check

Iteration 1 — not applicable.

---

## Recommendation

FAIL. Five blocking defects (D1-D5), three of them critical. Fix route, in the order that minimises
rework — D3 first, because D4's wording depends on its outcome and D2's contract depends on the
classification predicate:

1. **D3** — redraw §3 over (registered / classified / carries-token), or split the registered row
   by classification and state the grid's population. Define the `stamp` vs `prose` predicate in
   REQ-VSP-002 ("a bump must rewrite this file or a test breaks"). Name the
   misclassified-as-prose case as a second uncovered case in §7.
2. **D1** — choose the sweep's file population explicitly and make it a requirement. Re-derive
   34 / 28 / 121 against that population with the matching command (`grep -rl` if the walk stays
   filesystem-wide), and add the gitignored roots — starting with `.claude/worktrees/` — to the
   §2.3 enumeration with their own reasons and counts.
3. **D2** — add a maintenance-contract requirement + AC (owner, timing relative to the bump,
   behaviour during the window). Correct §8 R-4 to the measured numstat result and its
   sweep-collapse consequence. Name the registry file in `version-management.md`'s bump list.
4. **D5** — decide `.moai/docs/version-management.md`'s registry status before M5 and record it as
   an ordering constraint in `plan.md` §D; adjust the §D.3 constant table accordingly.
5. **D4** — open the "셋" enumeration in `plan.md` §E and add the `_test.go`-inline site and the
   misclassification site; extend AC-VSP-011(b) so the additions are grep-checked.

D6-D8 are surfaced for the orchestrator's discretion; none of them needs to gate the re-audit,
though D6's fix is one line and D7's is one `os.Stat`.

Retry ceiling: Tier M → 2 iterations (`harness.plan_audit_tier_ceilings`). The confirming re-audit
is scoped to D1-D5 plus a regression check over them, not a from-scratch re-read.

**Note for the next iteration's measurements:** this report contains the literal `v3.1.3`, so it
joins `.moai/reports/` and the group count moves 57 → 58. That is D6 demonstrating itself, and the
re-audit should expect the drift rather than read it as a defect in the SPEC's original arithmetic.
