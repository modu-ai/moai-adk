auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-CONTAINMENT-MAPPING-DOC-001
Iteration: 3/3
Verdict: FAIL
Overall Score: 0.90
Plan Artifact Hash: 20386d925add70aa38f9274b8054856f45085feedbad6471248ab5e7eff21697
Auditor Version: plan-auditor/v1 (Tier M threshold 0.80)

verdict: FAIL
audited_sha: 04890fc1336a8d4b4843be7208afbdbdb3ee8150
overall_score: 0.90
must_pass_failed: 0
blocking_count: 2
plan_artifact_hash: 20386d925add70aa38f9274b8054856f45085feedbad6471248ab5e7eff21697 (auditor-computed `cat .moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/*.md | shasum -a 256`; not the runtime ComputeHash, see Gaps G5)
scope: delta
fix_scope: spec.md#REQ-CMM-001 acceptance.md#E10a acceptance.md#E10b acceptance.md#AC-CMM-010 acceptance.md#Definition-of-Done acceptance.md#AC-CMM-008 acceptance.md#E08
defect_class: N3=ac-wording N4=ac-wording
reread_hunks: spec.md#REQ-CMM-001 acceptance.md#E10a acceptance.md#E10b acceptance.md#AC-CMM-008
convergence_overall: fail
required_backend: codex fail

Reasoning context ignored per M1 Context Isolation (the card prompt's description of fix round 2 was used only to know which commands to re-run; every finding below rests on the artifacts and on commands run in this audit).

The aggregate (0.90) is above the Tier M threshold (0.80) and higher than round 2 (0.875), so no STOP signal applies. The verdict is FAIL because the required codex backend returned `fail` with two P2 findings, and my independent reproduction confirmed both (D-N3, D-N4). A required-gate failure is not absorbed into the aggregate.

## Round-2 defect status

| Id | Status | Basis |
|---|---|---|
| N1 (E10a could never pass: tracked `.moai/reports/t1760/*`) | RESOLVED | E10a exit 0 and E10b empty at HEAD 04890fc13; the fix-round-1 form still exits 1 (control) |
| N2 (E08 missed `october 16`, `OCTOBER 16`, `16 October`) | RESOLVED for the shapes named | 3 of 3 counted; 45 of 46 hit-fixture lines counted, the 1 miss is a listed Known-limit shape (`on 16th of Oct`); 0 of 51 look-alike lines counted |
| D1 (round 1, E08 `grep -w` vs Hangul particle) | no regression | `t1760은`, `2026년`, `develop를`, `카드t1760`, `한글2026년` all counted |
| D2 (round 1, E07 refused / counted occurrences / suffix) | no regression | one-path mutant `TOO-FEW-DISTINCT n=1` exit 1; `.md.NONEXISTENT` `BAD-SUFFIX` exit 1; missing path `MISSING` exit 1; passing fixture exit 0 |
| N3 (new) | OPEN, blocking | `.moai/reports` exclusion hides other cards' evidence |
| N4 (new) | OPEN, blocking | AC-CMM-008 text promises "any letter case" for months; `mAy 4` passes unseen |

## Must-Pass Results

- [PASS] MP-1 REQ numbers: REQ-CMM-001 to REQ-CMM-011, sequential, no gap or duplicate (spec.md:L65-L75). Traceability verb `COLLECTED: 11 REQ definitions (acceptance input: read)`.
- [PASS] MP-2 GEARS/EARS compliance, judged on the requirement layer (REQ entries in spec.md); ACs not graded here. Ubiquitous 001-003, 006, 007, 009, 011; Event-driven 004 ("When the author assigns..."), 008 ("When no MoAI defense..."); State-driven 005 ("While..."); REQ-CMM-010 is the legacy-EARS negative form ("shall not contain"), accepted as a legacy equivalent within the window.
- [PASS] MP-3 frontmatter, spec.md:L2-L14: id, title, version "0.1.0" (quoted), status draft, created/updated ISO dates, author, priority P2, phase "v3.2.0" (a release label), module, lifecycle spec-anchored, tags (string); optional `tier: M`. No rejected alias.
- [N/A] MP-4 language neutrality: SPEC covers one reference document, not multi-language tooling.
- [PASS] MP-5 D7: the only SPEC id in spec.md is its own (`grep -Eo` printed `SPEC-CONTAINMENT-MAPPING-DOC-001`), status draft; no retired/superseded reference, no REVIEW line possible.
- [PASS] MP-6 D8: `grep -c syscall spec.md` printed `0`; D8-4 auto-pass.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` printed only the missing-research.md warning, no match in plan.md (exit 1 = no match); research.md absent at Tier M.
- [PASS] MP-8 RED-now re-execution (see Evidence E1): E01 to E09 are release-blocking; each re-executed on the current tree with the document absent and each RED reproduces (E01 exit 1; E02/E04a/E04b/E06/E08/E09 `ugrep: warning ... No such file`, exit 2; E03/E05/E07 `awk: can't open file`, exit 2). Pins `cf5a52f55`, `1b2319571`, `629bb51cd` exist and the document is absent at each (`git cat-file -e`). E10a/E10b are declared regression-guards (green by construction), correctly outside MP-8.
- [PASS] MP-9: CN-4 verb `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates`, no `CONFLICT:` line. The 8 candidates were read: acceptance.md:L7, L24, L136, L192, L211, L219, L234 are pin/RED-now/ledger prose; L247 ("AC-CMM-010 ... evaluated at the run-phase tip ... before the sync phase begins") agrees with plan M4 `Exit: AC-CMM-008, 009, 010` and sync is not a plan milestone. No cross-artifact ordering conflict.

Required backend (audit plan, not an MP): codex `required` returned `fail` (receipt rcpt-304f9db25d40cef69764ae74). Reviewed defect, not an unmet gate: `moai verify audit-plan --result-file` printed `convergence_check: {ok: true, unmet: []}`.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | REQ-CMM-001 (spec.md:L65) allows all of `.moai/reports/` where only this card's directory is meant; AC-CMM-008 (acceptance.md:L44) promises "any letter case" for months |
| Completeness | 1.00 | 1.0 | HISTORY L19, Why L27, What L31, Requirements L61, AC (acceptance.md), four `### Out of Scope — <topic>` H3s with `-` bullets L88-L105, 12-field frontmatter |
| Testability | 0.75 | 0.75 | Every AC binary, but E08 over-promises (N4) and E10 cannot see deletion of other cards' evidence (N3); reviewer-checked ACs 011-014 are declared as such |
| Traceability | 1.00 | 1.0 | Verb printed no `UNCOVERED:` and no `ORPHAN:`; 11 REQs, 14 ACs (Tier M ceiling 16/16) |

## Defects Found

D1. N3 — spec.md:L65 and acceptance.md:L181-L195, L52-L53, L250 — REQ-CMM-001 and the pathspec `':!.moai/reports'` allow the whole `.moai/reports/` tree (1789 tracked files under it, measured by `git ls-files .moai/reports | wc -l`). Reproduced in a throwaway repo: deleting another card's tracked evidence leaves E10a exit 0 and E10b empty. The narrow form `':!.moai/reports/t1760'` still passes on this tree (E10a exit 0, E10b empty) and still flags the same deletion (exit 1). The acceptance text states the cost ("neither command sees a change under `.moai/reports/`"), but the requirement itself is what over-allows. — Severity: major — Class: blocking — Required fix: narrow REQ-CMM-001 to "the card-evidence directory `.moai/reports/t1760/`" and the pathspec in E10a and E10b to `':!.moai/reports/t1760'`; update AC-CMM-010 prose, the E10a/E10b ledger entries (re-observe green on the tree, exit 0 / stdout empty, plus one probe that a deletion under another card's directory now reads red), and the Definition of Done last bullet in the same edit (cross-layer sweep, verification-completeness.md §3).

D2. N4 — acceptance.md:L44 (AC-CMM-008), L164 (E08) — AC text says month names match "in any letter case except that lowercase `may` is not a month", but the command spells `May|MAY` only. Reproduced: `mAy 4`, `maY 4`, `4 MAy` each count 0 (exit 1) while `May 4` and `MAY 4` count 1 each. — Severity: minor — Class: blocking (the criterion's own stated scope is false) — Required fix: either reword AC-CMM-008 and the Known-limit paragraph to "`May` and `MAY` count; every other letter case of `may` does not", or extend the command; add `mAy 4` / `4 MAy` to the Known-limit uncounted list or to a must-hit fixture, whichever is chosen.

D3. O5 — acceptance.md:L52, L247, L250 vs spec.md:L75 — Optional. REQ-CMM-011 (no Template-First mirror) has no phase bound, but AC-CMM-010 is evaluated only at the run-phase tip; the Definition of Done says "in no phase does a file under `internal/template/templates/` change" with no command for the sync phase. A one-line `git diff --quiet <base> -- internal/template/templates` re-run at sync would close it. — Severity: minor — Class: optional.

D4. O6 — acceptance.md:L113-L127 (E08, date branch `(0[1-9]|1[0-2])[-/](0[1-9]|[12][0-9]|3[01])`) — Optional, status inferred from reading the regex, not run: a Korean-or-English section range such as `10-12` would match as a month-day pair. The document is unlikely to carry one; the cost is a false red a reviewer can resolve. — Severity: minor — Class: optional.

## Regression Check (Iteration 2+)

- N1: RESOLVED (E10a exit 0, E10b empty, both at HEAD 04890fc13; control without `':!.moai/reports'` exits 1; older base `46960ac7d` exits 1).
- N2: RESOLVED for `october 16`, `OCTOBER 16`, `16 October`, plus 45 of 46 shapes in a wider fixture.
- O1, O2, O3 (round 2 optionals): O3 confirmed fixed (CN-4 reads 4 milestones); O1 and O2 text present (acceptance.md:L7, L242).
- Round 1 D1, D2: no regression (Evidence E4, E5).
- No score regression: 0.90 vs 0.875.

## Recommendation

1. Apply D1 (N3) and D2 (N4) as a delta; both are confined to `fix_scope`. The two fixes touch REQ-CMM-001 body text, which `manager-spec` owns (re-delegation), and the acceptance anchors listed.
2. Keep E10a/E10b green on the tree after narrowing and record one red probe for the narrowed form.
3. D3 and D4 are optional; fold them into the same edit only if free.
4. Iteration 3/3: the remaining scope is two anchors, so it fits the ceiling policy's delta-round rule (diff inside `fix_scope`, REQ/AC id sets unchanged). The re-audit must still re-run the CN-4 ordering verb in full and the required codex leg.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim

C1. N1 fixed. C2. N2 fixed for the named shapes. C3. D1/D2 of round 1 did not regress. C4. Two new blocking defects (N3, N4). C5. All nine release-blocking RED-now cells reproduce RED. C6. MP-1 to MP-9 pass; the required codex leg fails.

### Evidence

E1. RED-now re-execution (document absent), cited commands run verbatim, observed:

```text
test -f .moai/docs/multi-agent-containment-mapping.md                      -> exit=1
grep -c -i -F 'reference-only' <DOC>                                       -> ugrep: warning: <DOC>: No such file or directory ; exit=2
awk -F'|' -v rowre=... (E03) <DOC>                                         -> awk: can't open file <DOC> ; exit=2
grep -c -F 'https://' <DOC>  (E04a) / grep -c -E 'https://.*\| ...' (E04b) -> ugrep: warning ...No such file ; exit=2 / exit=2
awk -F'|' -v rows=... (E05) <DOC>                                          -> awk: can't open file ; exit=2
grep -c -E '\| gap \|$' <DOC> (E06)                                        -> ugrep: warning ... ; exit=2
awk -v re=... (E07) <DOC>                                                  -> awk: can't open file ; exit=2
grep -c -E '<E08 pattern>' <DOC> (E08, full ledger pattern)                -> ugrep: warning ... ; exit=2
grep -c -F -e '[HARD]' -e '[ZONE' <DOC> (E09)                              -> ugrep: warning ... ; exit=2
```

E2. N1:

```text
git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'   -> E10a exit=0
git status --porcelain --untracked-files=all -- . (same three pathspecs)                                                                   -> stdout empty, E10b exit=0
same E10a without ':!.moai/reports'                                                                                                          -> exit=1  (control)
same E10a, base 46960ac7d8043f1bee3f352688cceba8fb405738                                                                                     -> exit=1
git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 HEAD -> .moai/reports/t1760/plan-audit.md, .moai/reports/t1760/progress.md, and the four SPEC files only
```

E3. N2 and look-alikes, E08 full ledger pattern run verbatim against fixtures in session scratch (not committed):

```text
e08_n2.txt   (october 16 / OCTOBER 16 / 16 October)         lines=3   -> 3
e08_hit.txt  (46 shapes incl. Oct.16, (October 16), 16th Oct) lines=46 -> 45 ; grep -n -v shows the one miss: "46:on 16th of Oct" (listed Known limit)
e08_zero.txt (51 look-alikes: developer, 30,000건, Opus 5.5, Mark 5, Marketing 5, Mayor 5, Octal 7, "Fit is aligned", a register row...) -> 0, exit=1
e08_may.txt  (mAy 4, maY 4, 4 MAy, may 4, 4 may, May 4, MAY 4, maY) lines=8 -> 2   (only May 4 and MAY 4 counted; mAy 4 / maY 4 / 4 MAy are NOT)
```

E4. D1 (Hangul boundaries) is inside E3 (`t1760은`, `2026년`, `develop를`, `카드t1760`, `한글2026년` counted; `developer를` and `v3.2.0은` not).

E5. D2/E03/E05/E07 regression: passing Korean fixture (4 register rows, 7 mapping rows incl. one gap, 9+ distinct existing paths, Hangul particles after paths, `developer를`) gave E03 exit 0 / E05 exit 0 / E07 exit 0, and E04a=4, E04b=4, E06=1, E02=1, E09=0. Mutants:

```text
E03: only BP-1 rows  -> COUNT BP-2=0 COUNT BP-5=0 COUNT BP-6=0 ; BP-1 twice -> COUNT BP-1=2 COUNT BP-2=0 ; short fragment -> BAD-FRAGMENT ; flag MAYBE -> BAD-FLAG ; exit 1
E05: no section marker -> NO-PATH-SECTION ; swapped header -> HEADER-COUNT=0 ; Fit maybe -> BAD-FIT x3 ; theme dropped -> THEME-MISSING: merge-queue-alternative ; section marker only in claim cell -> NO-PATH-SECTION
E07: one path cited everywhere -> TOO-FEW-DISTINCT n=1 exit 1 ; .md.NONEXISTENT -> BAD-SUFFIX exit 1 ; nonexistent .md -> MISSING exit 1
```

E6. N3 reproduction (throwaway repo in session scratch, base commit with `.moai/reports/t9999/verdict.md` and `.moai/reports/t1760/plan-audit.md`; `rm .moai/reports/t9999/verdict.md`):

```text
E10a (cited form, ':!.moai/reports')      -> exit=0
E10b (cited form)                          -> stdout empty, exit=0
narrow form ':!.moai/reports/t1760' E10a  -> exit=1
on the real tree, narrow form: E10a exit=0, E10b stdout empty exit=0
git ls-files .moai/reports | wc -l        -> 1789
```

E7. Verbs: CN-4 `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 12 exit bindings, 8 ordering candidates` (no CONFLICT). Traceability `COLLECTED: 11 REQ definitions (acceptance input: read)` (no UNCOVERED, no ORPHAN). `mcp__moai__spec_audit` (project_root = this worktree, filter this SPEC): `modern_era_clean: 1`, one INFO `EraAutoDetected` (H-5), `total_specs: 1`.

E8. Cross-model: `moai verify audit-plan --project-root <toplevel>` printed `model: codex`, `cross_model_active: true`, `enforced_required: ["codex"]`. `mcp__moai__audit_multi` (project_root = toplevel, target baseBranch, no `gates` argument): `overall_verdict: fail`, `per_backend_verdicts: [{codex, required, fail}]`, `participant_count: 1`, `audit_receipt: rcpt-304f9db25d40cef69764ae74`, `plan_source: config`; two P2 findings (reports exclusion; mixed-case `may`). Result digest written to `.moai/state/audit-plan-result.json`, then `moai verify audit-plan --result-file ...` printed `convergence_check: {"ok": true, "unmet": []}` (the required leg ran and failed on merit).

### Baseline-attribution

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1760`, HEAD `04890fc1336a8d4b4843be7208afbdbdb3ee8150`, branch WT-10-15-release-3, working tree clean at start and at the end (`git status --short` empty). Commands are those printed above, run in this audit. Judging build for the MCP and `moai verify` tooling: `moai-adk v3.2.0-rc.36-t1822c`, server commit `ee981a932`; `git merge-base --is-ancestor ee981a932 HEAD` printed exit 1, so that commit is not a lagging ancestor of this tree. Comparison base for E10: `629bb51cda73fc70422769988b8c28182c9bd2d5` (pinned, not a moving ref).

### Gaps

G1. The deliverable document does not exist, so E01 to E09 were observed RED on the absent document and exercised on constructed fixtures only; their green on the real document is a run-phase observation.
G2. AC-CMM-011 to AC-CMM-014 are reviewer-checked by design and were not executed; the quoted source fragments were not re-fetched from the network in this audit (plan-phase claim of `grep -c -F` counts was not re-observed).
G3. Codex's two findings are reproduced here, but the codex review itself ran as a black box (no transcript); I did not independently read its prompt.
G4. No tool refusal was encountered: every cited command, the auditor-side awk programs (`awk -f`), and the MCP calls ran without a guard refusal. The `tool_failure:*` rows of `.moai/lessons-inbox.jsonl` were not inspected, so absence of refusal rows is not confirmed there.
G5. `plan_artifact_hash` is an auditor-computed sha256 of the four SPEC markdown files (`cat ... | shasum -a 256`), not the runtime `ComputeHash` over the tier's artifact set; admission may compute a different value.
G6. HTML-entity spellings and the other date shapes listed in acceptance.md Known limit were not exercised beyond `on 16th of Oct`.

### Residual-risk

R1. After N3/N4 are fixed, E08 still counts lines and enumerates date shapes; shapes outside the list pass unseen (the document is Korean prose with few dates, so the risk is low but not zero).
R2. A change under `.moai/reports/t1760/` stays invisible to E10 after narrowing (intended; the evidence reader owns it).
R3. If the run phase merges `main` into the card branch, E10a (fixed base `629bb51cd`) would see other cards' files and read red at the run-phase tip; inferred from the fixed-base design, not tested here.
R4. Paraphrase drift in defense cells remains reviewer-only (AC-CMM-014).

## Operational Notes (unverified)

- [measured] Narrow the scope guard and re-observe: `git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports/t1760'` printed exit 0 on this tree (this audit).
- [inferred] Rule applied: `verification-completeness.md` §2 (mutant probe) and §3 (cross-layer revision sweep): a deletion mutant satisfies the criterion, and the requirement and criterion must be rescoped together.
- [assumption] Measure whether the landing sequence merges `main` into the card branch before the run-phase tip check (R3) — command: `git log --merges --oneline 629bb51cd..HEAD` on the run-phase tip.

AUDIT-VERDICT: FAIL spec=SPEC-CONTAINMENT-MAPPING-DOC-001 receipts=rcpt-304f9db25d40cef69764ae74
