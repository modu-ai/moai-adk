# SPEC Review Report: SPEC-SPECLINT-GITBLIND-001

Iteration: 4/4 (operator-authorized single override — Tier M ceiling is 2 and was exceeded at iter-3; this is the final authorized retry)
Version audited: **0.4.0**
Verdict: **PASS**
Overall Score: **0.87** (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. The resume record
(`.moai/reports/t371/verdict.md` § 재개 기록 — 2026-09-02) and progress notes were read as
artifacts under audit; every load-bearing claim in them was re-verified with commands run in
this session against the audit tree.

## Audit tree (pinned)

```
$ git rev-parse HEAD && git branch --show-current
06d9084553b86f682fabe8a3a252131ab695ec7d
WT-lint-shallow-clone

$ git merge-base --is-ancestor ad272be20 HEAD && echo ABSORBED
ABSORBED   (origin/develop ad272be20 fully absorbed via merge 4db7ad66e)

$ git merge-base --is-ancestor d31bf744d HEAD && echo ANCESTOR
ANCESTOR   (2-cell ledger measurement tree d31bf744d is an ancestor of HEAD)

$ git diff --stat d31bf744d..HEAD -- internal/spec .github/workflows/spec-lint.yml
(empty — neither the measured code nor the workflow changed since the ledger's tree)
```

Fix commits audited: `322dffdf2` (coordinate unification + D1 premise + 2-cell ledger) and
`06d908455` (label re-stamp). Both touch SPEC markdown only — no code, no REQ/AC semantics.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 9 unique ids `REQ-SLGB-001 … REQ-SLGB-009`,
  sequential, no gaps, no duplicates, uniform padding. `REQ-SLGB-005` and `REQ-SLGB-009`
  each appear twice in `grep -o` output — both second occurrences are back-references
  (spec.md §4 residual-risk mention; REQ→AC coverage table), not duplicate definitions.
- **[PASS] MP-2 GEARS format compliance** — judgment made against the requirement layer
  (`spec.md` §2) only. All nine match a GEARS pattern: Ubiquitous (001, 005, 006, 008, 009),
  Event-driven `When …` (002), State-driven `While …` (003, 007), Unwanted `내지 않는다` =
  shall not (004). The Given/When/Then entries in `acceptance.md` are verification-layer
  AC-XXX criteria in their correct format (M3 § Scope) and are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct
  types (spec.md L2-L15): `version: "0.4.0"` quoted semver, ISO dates, `priority: P1`,
  `phase: "v3.1.4 target"` (a release label — not a prohibited lifecycle-stage value),
  `lifecycle: spec-anchored`, `tags` comma-separated string. No snake_case alias. `tier: M`
  and `era: V3R6` are additive optional fields.
- **[N/A] MP-4 Section 22 language neutrality** — SPEC scoped to this Go repository's own
  `internal/spec` lint and one GitHub workflow; single-language by construction. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — five external SPEC ids referenced across the
  three artifacts; statuses read from disk this session:
  `SPEC-KANBAN-TODO-CLI-001 → in-progress`, `SPEC-LSPMCP-001 → superseded`,
  `SPEC-UPDATE-DOC-DRIFT-001 → draft`, `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 → implemented`,
  `SPEC-V3R6-SKILL-GEARS-ALIGN-001 → implemented`. The one terminal-status reference
  (SPEC-LSPMCP-001) is reconciled at its reference site — spec.md §1.3 cites it *because of*
  its terminal status, and the code premise holds on this tree (I verified
  `applyEraDemotion` at `lint.go:344-361` carries the `SeverityWarning → Advisory=true`
  branch and `terminalStatusEnum` includes `superseded` at `lint.go:1340`). No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall' .moai/specs/SPEC-SPECLINT-GITBLIND-001/`
  exits 1. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' <SPEC dir>` exits 1.
  Tier M artifact set (spec/plan/acceptance + progress.md); no `research.md` exists — N/A for
  that file, and `plan.md` is clean.

**No must-pass failure.**

---

## Delta dispositions (iter-3 blocking defects)

### D1 — vacuous RED baseline in AC-SLGB-001/004 — **RESOLVED**

- acceptance.md L41-45 (AC-SLGB-001) and L77-80 (AC-SLGB-004) now carry the [HARD] fixture
  premise: the `fixtureSpecMD` (`drift_characterization_test.go:109`) precedent is followed
  for **frontmatter only**, the body-copy is forbidden, and the fixture body MUST carry
  `## 4. Scope` + `### 4.1 Out of Scope — <suffix>` + one item. AC-004 incorporates AC-001's
  premise wholesale by reference.
- The verified form exists and I read it: `.moai/reports/t371/repro/withscope/spec.md` —
  12-field schema-valid frontmatter, `status: draft` (non-terminal), `## 4. Scope` (L30),
  `### 4.1 Out of Scope — unrelated surfaces` (L32), one item (L34).
- **I re-ran the RED-now command myself on the audit tree** (not consumed from the resume
  record):

  ```
  $ go run ./cmd/moai spec lint .moai/reports/t371/repro/withscope/spec.md
  ✓ No findings — all SPEC documents are valid
  rc=0
  ```

  The RED baseline is now **reachable and measured**: pre-M1 the `✓ No findings` line
  actually prints, so the ACs' "line absent" assertions are live rather than vacuous — the
  exact defect iter-3 measured is closed. The premise is normative prose in the AC itself
  (the repro file is evidence, not the carrier), so it survives worktree disposal.

### D2 — two-cell adoption for all 11 ACs — **RESOLVED**

acceptance.md L181-201 carries the "2-Cell 채택 대장" with one row per AC (RED-now + why-red +
green-path + classification), document-level tree pin `d31bf744d` (valid: ancestor of HEAD;
measured surfaces unchanged since — empty diff). Judgment against
`verification-completeness.md` §1.1/§1.2/§2/§2.1:

- **실측 2 — correctly adopted.** I re-ran both cells verbatim on the audit tree:
  - AC-SLGB-009: `grep -n 'fetch-depth' .github/workflows/spec-lint.yml` → no stdout,
    **exit 1** (reproduced). All four elements present: single-invocation read-only command;
    empty stdout recorded as absent (the rule's §2.1 itself calls an empty stdout with
    non-zero exit "a complete and common observation"); exit code as its own field; tree pin
    (document-level, permitted); why-red names the missing `with:` block and fetch step — I
    confirmed `spec-lint.yml:31` is the bare `actions/checkout@v7` with no `with:`.
  - AC-SLGB-010: `grep -n -A6 'paths:' …` → **exit 0** (reproduced); both trigger lists carry
    only `.moai/specs/**` (`:5-6` pull_request, `:14-15` push) — the (b)/(c) coverage gaps the
    cell names are real. One form gap: the stdout is recorded as a characterization
    ("두 목록이 각각 `.moai/specs/**` 단독") rather than raw bytes — minor, on a
    non-release-blocking criterion (see D2 of this iteration's defects).
- **seam 8 — sanctioned disposition, not evasion.** The deferral rests on three verified
  legs: (i) each row states a genuine construction blocker — the target finding does not
  exist pre-implementation (AC-002/003/006), `cachedMainBranch` is unexported and callable
  only in-package (AC-007/008, airtight), and shallow fixtures are cwd-dependent while the
  worktree isolation guard rejects cwd changes (AC-004a/b — the blocker is real: the same
  guard rejected this session's compound git loop during the D7 check); (ii) the [HARD] lead
  ruling (verdict.md § 순환 해소) correctly re-frames the main-absent observation as M2's
  **output** — the seam M2 builds is the instrument that makes the observation possible, and
  pulling it before M1 means fighting the isolation guard rather than building the SPEC's own
  deliverable; (iii) the deferred observation has contractual carriers — plan.md §E's
  three-layer vacuous-green guard, acceptance.md DoD, and the manager-develop §E.8
  verbatim-RED obligation. Decisively: wherever a plan-time RED *was* executable (the D1
  baseline, AC-009/010), it was executed and recorded — the deferral is applied only where
  construction requires implementation artifacts, and every deferred row still states why it
  is red today. AC-SLGB-005 explicitly names its own vacuous green and substitutes the §2
  mutant-probe clause — the rule's own prescription for the vacuous direction.
- **CI 1 — correct §2.1 application.** AC-SLGB-011 applies the undecidable disposition
  verbatim: regression-guard, release-blocking=false, closes via CI job log post-sync; its
  body already mandated the CI-only judgment place and plan.md §G forbids reading CI green
  as evidence. This is the disposition, not one option among several.

### D3 — stale forward coordinates — **RESOLVED** (one minor residual, see D1 below)

Every dispatch spot-check verified on the audit tree, plus a full sweep beyond them:

| Citation | Claim | Verified |
|---|---|---|
| plan.md:101 + plan.md:170 | both cite err skip `:1373-1376` | **agree**; `lint.go:1373-1376` is exactly the `if err != nil { return nil }` block inside `Check` (func at `:1355`) — the cross-check invariant holds |
| plan.md §H parent | `lint.go:1355-1391` Check | exact; terminal `:1367-1369`, emission `:1379-1388` both inside; `Advisory: true` at `:1384` (spec.md §1.3) exact |
| plan.md:172 | `gitquery_cache.go:108-136` cachedMainBranch | exact; cache early return `:115-119` exact (also acceptance.md:126); `.Dir`-absent sites `:121`/`:132` exact |
| plan.md §H | `drift.go:300-367` getGitImpliedStatus | exact; three error shapes at `:312` (`git log failed`), `:316` (`no git history found`), `:366` (`no classifiable commit within window`) — content-verified, all inside; `drift.go:68` wiring exact |
| plan.md §H / spec.md §1.2 | `lint_ownership.go:380-400` Unreachable | contained: the construct spans ~387-399 (Code literal `:395`); range opens inside the preceding Skipped block — the same harmless range-start imprecision iter-3 explicitly examined and recorded as a non-defect; only a line-count-neutral rename changed in this file since iter-3 |
| plan.md §H / spec.md §4 | `spec_lint.go:113` printTable, short-circuit `:115-118` (msg `:116`), summary `:136-142` | exact for the first three; summary range covers the error/warning counting arms (`:139-142`) with the print at `:145` 2-3 lines beyond the range end — benign, inside printTable |
| spec.md §4 | `lint.go:33-45` Finding, `applyEraDemotion :344-361` | exact; signature is `cause demotionCause` as the resume record describes |
| spec.md §5 / plan.md §H | test precedents `:53 :55 :98 :103 :109 :280`; ci.yml six `fetch-depth` lines | all exact (`:129 :264 :382 :431 :486 :543`) |

- **Containment**: every backtick sub-coordinate in plan.md §H falls inside its entry's
  declared parent range — the failure shape iter-3 caught by eye is gone.
- **plan.md:16 label** re-stamped to `ad272be20` with the "11 measured positions unchanged
  between absorptions" claim — verified consistent: the second absorption touched neither
  `internal/spec` nor the workflow (empty diff `d31bf744d..HEAD`), and every measured
  coordinate above holds on HEAD `06d908455`.
- **Residual (this iteration's D1 defect)**: plan.md:29 still cites the cache globals as
  `gitquery_cache.go:21-23`. The measured value — recorded in the resume record's own table
  (verdict.md:293, "전역 뮤텍스 `:23-25`") and carried correctly at acceptance.md:10 — never
  reached plan.md §B. Actual fields sit at 23-24. One write-through miss from the
  `322dffdf2` unification; lands inside the same var block, cannot mislead an implementer,
  and no completeness claim rides on it (unlike iter-3's D3). Minor/optional.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | between 0.75 and 1.0 | Unchanged from iter-3 (both named blemishes — REQ-SLGB-009 rationale folded into normative text; §2.2's cache-inactive exception unreconciled in-artifact with REQ-SLGB-001's absolute — persist; neither was in delta scope). The delta artifacts add precision: the fixture premise is now unambiguous (acceptance.md:41-45) and the adoption classification is explicit per AC. |
| Completeness | 0.90 | between 0.75 and 1.0 | Both iter-3 deductions closed: the two-cell structure now exists for all 11 ACs (acceptance.md:181-201); plan.md:16's baseline label is re-stamped from a fresh absorption-verified read. All sections/frontmatter present; three Out of Scope H3 sub-headings with bullets. Remaining deductions: plan.md:29's stale coordinate (D1 below); AC-010's stdout recorded as characterization (D2 below). |
| Testability | 0.83 | between 0.75 and 1.0 | The two iter-3 drivers are closed with re-runnable evidence: the flagship RED baseline is measured and I reproduced it (`✓ No findings`, rc=0); every AC carries the two-cell pair; both measured cells reproduce exactly on the audit tree. Remaining deductions: 8 of 11 RED-now cells are sanctioned deferrals rather than plan-time observations (carriers contracted to run-phase §E.2/§E.8); AC-010's stdout form gap; iter-3's D6 shallow-mutant on AC-002 remains unadopted (optional). |
| Traceability | 0.90 | between 0.75 and 1.0 | D3 closed coordinate-by-coordinate on the audit tree — the §H surface an implementer navigates by is now accurate, with the err-skip cross-check invariant (`:1373-1376` at both plan.md:101 and :170) holding. REQ→AC coverage table intact; all nine REQs covered; one declared orphan (AC-SLGB-011, deliberate). Deduction: the plan.md:29 residual. |

Aggregate (harmonic mean): `4 / (1/0.85 + 1/0.90 + 1/0.83 + 1/0.90) = 4 / 4.6035 = 0.869` → **0.87**.

**Sensitivity (disclosed):** the verdict is not thin. Holding the other three dimensions,
Testability would have to fall below **0.62** — fourteen points under iter-3's 0.68, whose
drivers (0/11 adoption, broken flagship RED) are both now measurably closed — for the
aggregate to miss 0.80.

---

## Defects Found

**D1.** `plan.md`:L29 — stale forward coordinate: cache globals cited as
`internal/spec/gitquery_cache.go:21-23`, but the fields sit at 23-24 and the unification
pass's own measured value (`:23-25`, verdict.md:293; carried at acceptance.md:10) was never
written into plan.md §B. The two artifacts now disagree about the same construct.
— Severity: **minor** — Class: **optional** — Required fix: change `:21-23` to `:23-25`
(one-line edit; can ride any later plan.md touch).

**D2.** `acceptance.md`:L200 (2-cell ledger, AC-SLGB-010 row) — the RED-now stdout is
recorded as a characterization ("두 목록이 각각 `.moai/specs/**` 단독") rather than the raw
grep bytes. The discriminating content is present and I reproduced the command (exit 0, both
lists `.moai/specs/**`-only), and §2.1's verbatim-stdout element binds release-blocking
criteria (this one is not classified as such), so this is a form gap, not an adoption gap.
— Severity: **minor** — Class: **optional** — Required fix: quote the 12-line grep output
verbatim in the cell or in a cited evidence-ledger entry.

**Carried from iter-3 (optional there, unchanged, not regressed — orchestrator discretion):**
iter-3 D4 (lint.skip ordering vs the once-per-run suppression flag), D5 (t382 disproof stated
one notch wide), D6 (AC-SLGB-002 message mutant: assert `origin/main`/`origin/master`), D7
(AC-SLGB-011 REQ-less, declared in the coverage table).

---

## Regression Check (iter-1 closed defects + iter-3 must-pass)

| Prior defect | Status | Evidence (this tree) |
|---|---|---|
| iter-1 D1 — error shapes are three | CLOSED, holds | spec.md §2.1 table; `drift.go:312/:316/:366` literals content-verified |
| iter-1 D2 — output-volume decision | CLOSED, holds | spec.md §2.2 + AC-SLGB-003 bound of exactly 1 |
| iter-1 D4 — REQ-SLGB-002 untraced | CLOSED, holds | AC-SLGB-002 present (acceptance.md:51-58) |
| iter-1 D7 — terminal-status fixture trap | CLOSED, holds | AC-001 Given names the non-terminal constraint; AC-005 [HARD] cites `lint.go:1367-1369` — verified exact |
| iter-1 D8 class — no observed red on record | CLOSED by this delta | 2-cell ledger (11/11) + D1 measured baseline; pre-implementation RED baseline re-confirmed: `grep -rn "StatusGitUnreachable" internal .github \| wc -l` → **0** |
| iter-3 must-pass 7/7 | PASS, holds | re-verified independently above (MP-1 … MP-7) |
| New defects from fix commits `322dffdf2`/`06d908455` | none of blocking shape | both diffs touch SPEC markdown only; no REQ/AC semantic changes; the single new-shape item is D1 above (a missed write-through) |

Tier M budget: REQ 9 ≤ 16; AC 11 headings (12 with the 004a/b split) ≤ 16. Within budget.

---

## Gaps (explicitly not observed)

- The full-corpus lint (`0 error / 1096 warning`) was not re-run; no AC depends on it (all
  fixture ACs run against `t.TempDir()` fixtures). Consumed from
  `.moai/reports/t371/lint-local-35bc0715f.txt` as in iter-3.
- A genuinely shallow fixture (`--is-shallow-repository == true`) was not constructed — same
  gap as iter-3; it is now explicitly classified as the M1 seam in the ledger with its
  observation scheduled for run-phase §E.2.
- The 18-SPEC ID *set* behind the StatusGitConsistency count was not re-verified (count
  only, per iter-3).
- No cross-model backends invoked: no `audit_model: multi` is configured in this project
  (grep of `.moai/config/` finds no `audit_model` key), consistent with the Claude-only
  iter-1/iter-3 audits of this SPEC.

## Residual risk

- The coordinate-drift class is structural: any card touching `internal/spec` re-shifts
  these citations at integration. plan.md:29 is a live instance. A coordinate pass will be
  needed again after merge regardless of this verdict.
- The seam-classified RED observations (8 ACs) land in run-phase §E.2 — the adoption is
  contingent on those observations actually being recorded there. The kickoff delegation
  prompt should carry that obligation explicitly (plan.md §E and acceptance.md DoD already
  mandate it; the delegation should point at them).

## Recommendation

**PASS.** All three iter-3 blocking defects are resolved with independently re-measured
evidence (D1's RED baseline re-run by this auditor, rc=0; both 실측 cells reproduced verbatim;
D3's coordinates verified line-by-line on the pinned audit tree), no must-pass failure
exists, no closed defect regressed, and the aggregate clears the Tier M threshold with
margin robust to any defensible Testability reading. The two new findings are minor/optional
and do not gate run-phase entry; the orchestrator may fold D1's one-line coordinate fix into
any later plan.md touch. Implementation Kickoff Approval remains the next mandatory human
gate and is not bypassed by this verdict.

Per the retry-loop contract: this was the single operator-authorized iteration-4 retry. No
further iteration is proposed or performed.
