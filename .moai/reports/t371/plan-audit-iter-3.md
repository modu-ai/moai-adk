# SPEC Review Report: SPEC-SPECLINT-GITBLIND-001

Iteration: 3/3 (Tier M ceiling is 2 — see § Procedural notes)
Version audited: **0.4.0**
Verdict: **FAIL**
Overall Score: **0.79** (Tier M PASS threshold 0.80 — below threshold; sensitivity disclosed in § Scoring sensitivity)

Reasoning context ignored per M1 Context Isolation. The SPEC author's account in `progress.md`
§E.1 was read as an artifact under audit, never as a verdict.

## Audit tree (pinned)

Re-read at audit start, not carried from the dispatch:

```
$ git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t371
bff12c54d
WT-lint-shallow-clone
```

```
$ git log --oneline -3
bff12c54d docs(SPEC-SPECLINT-GITBLIND-001): refresh lint.go citations after develop absorption, record iter-3 baseline (t371)
35bc0715f Merge remote-tracking branch 'origin/develop' into WT-lint-shallow-clone
3b637a290 docs(SPEC-SPECLINT-GITBLIND-001): plan-phase artifacts and audit evidence (t371)
```

**Tool provenance (§2.2).** All lint measurements below were produced by `./bin/moai`, built in
this tree at `35bc0715f` (mtime 19:15). No code-bearing commit exists between that build and HEAD:

```
$ git log --oneline 35bc0715f..HEAD -- internal/ cmd/ pkg/
(empty)
```

So the judging build and the judged tree are the same code. The PATH binary was not used.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-SLGB-001 … REQ-SLGB-009`, nine unique ids, no gaps,
  no duplicates, uniform zero-padding.
  ```
  $ grep -o 'REQ-SLGB-[0-9]*' .moai/specs/SPEC-SPECLINT-GITBLIND-001/spec.md | sort -u
  REQ-SLGB-001 … REQ-SLGB-009   (9 lines)
  ```
- **[PASS] MP-2 GEARS format compliance** — judgment made against the **requirement layer**
  (`REQ-XXX` in `spec.md` §2), not the verification layer. All nine match a GEARS pattern:
  Ubiquitous (001, 005, 006, 008, 009), Event-driven `When …` (002), State-driven `While …`
  (003, 007), Unwanted `… 내지 않는다` = shall not (004). The `Given/When/Then` entries in
  `acceptance.md` are `AC-XXX` verification-layer criteria and are correctly formatted as such
  (M3 § Scope); they are graded under Group 4 and are **not** penalized here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md` L2-L14): `id`, `title`, `version: "0.4.0"` (quoted semver), `status: draft`,
  `created`/`updated` ISO dates, `author`, `priority: P1`, `phase`, `module`, `lifecycle:
  spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias
  (`created_at`/`updated_at`/`labels`/`spec_id`) present. Extra `tier: M` and `era: V3R6` are
  additive, not substitutions.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC is scoped to this Go repository's own
  `internal/spec` lint and its GitHub workflow. It is single-language by construction and names no
  multi-language tooling. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — six external SPEC ids referenced. Statuses read
  from disk:
  ```
  SPEC-KANBAN-TODO-CLI-001            status: in-progress
  SPEC-LSPMCP-001                     status: superseded
  SPEC-INTEGRATION-LOCK-ATOMIC-001    status: implemented
  SPEC-UPDATE-DOC-DRIFT-001           status: draft
  SPEC-V3R6-SKILL-GEARS-ALIGN-001     status: implemented
  SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001  status: implemented
  ```
  All six exist. One is `superseded` (`SPEC-LSPMCP-001`), and it is reconciled at the reference
  site: `spec.md` §1.3 references it *because of* its terminal status — "SPEC-LSPMCP-001 은
  terminal status 라 `applyEraDemotion` 이 advisory 로 낮춘다". I verified that reasoning holds on
  this tree (`lint.go:239` computes `demote` from `terminalStatusEnum[…Status]`; `applyEraDemotion`
  `:307-308` sets `Advisory` on the Warning branch). No BLOCKING D7 finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall' .moai/specs/SPEC-SPECLINT-GITBLIND-001/`
  exits 1 (no match). D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-SPECLINT-GITBLIND-001/`
  exits 1 (no match) across `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.

**No must-pass failure.** The FAIL verdict below is driven by the rubric scores and three blocking
defects, not by the M5 firewall.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | between 0.75 and 1.0 | Requirements carry single interpretations and measurable content. Two named blemishes: `spec.md` REQ-SLGB-009 folds rationale into normative text ("현재 필터로는 … 닫히지 않는다"); `spec.md` §2.2's cache-inactive exception is never reconciled in the artifact with REQ-SLGB-001's absolute "실행당 최대 1건". I resolved the second myself (`lint.go:186` `startGitQueryCache()` is unconditional at `Lint()` entry, so the exception cannot reach the `moai spec lint` path) — but a reader must go to source for that. |
| Completeness | 0.80 | between 0.75 and 1.0 | All sections present: HISTORY (L18-27), WHY/§1, WHAT/§2, §3, §4, and three `### Out of Scope — <topic>` H3 sub-headings each carrying `-` bullets. Frontmatter complete. Deductions: `acceptance.md` carries none of the §2 two-cell adoption structure (D2); `plan.md:16` still states the pre-absorption baseline `develop 1e5199b88` after an iteration whose stated purpose was re-measurement. |
| Testability | 0.68 | below the 0.75 band | AC content is binary and weasel-free (no "적절히"/"잘"/"reasonable" anywhere), and the two inherited-green ACs each carry an explicit mutation clause — genuinely good work. Against that: **zero of 11 ACs carry a RED-now cell** (D2), and the flagship RED baseline of AC-SLGB-001/004 is **measurably broken as written** (D1). Both defects sit exactly on the instrument's ability to deliver a verdict. |
| Traceability | 0.90 | between 0.75 and 1.0 | Explicit REQ→AC coverage table; all nine REQs covered, verified id-by-id. One declared orphan (AC-SLGB-011 maps to no REQ — stated deliberately in the table). Deduction: `plan.md` §H's artifact→code coordinates are stale and self-contradictory (D3), which is the traceability surface a run-phase implementer actually navigates by. |

Aggregate (harmonic mean, per `agent-common-protocol.md` § Skeptical Evaluation Stance):
`4 / (1/0.85 + 1/0.80 + 1/0.68 + 1/0.90) = 0.7987` → **0.79**.

### Scoring sensitivity (disclosed, not hidden)

The aggregate is close to the threshold and moves with the Testability band alone. At Testability
0.70 the aggregate is 0.805; at 0.68 it is 0.799; at 0.65 it is 0.788. I chose 0.68 and I state the
reason so it can be disputed: 11 of 11 ACs lack the [HARD] adoption structure, and I *measured* one
of the two flagship RED baselines failing. A margin this thin is not evidence of a pass — it is
evidence that the verdict rests on the defect list, which is where I have put it.

---

## Defects Found

**D1.** `acceptance.md`:L36-42 (AC-SLGB-001 [HARD] 픽스처 전제) and L74-77 (AC-SLGB-004 [HARD] 픽스처
전제) — **The named fixture precedent measurably breaks the RED baseline these ACs depend on.**
Both ACs assert the *absence* of the line `✓ No findings — all SPEC documents are valid`, and both
make that assertion meaningful by requiring a fixture that produces zero findings before M1 — "
`fixtureSpecMD` 선례를 따른다". I built exactly that fixture and ran the tree binary:

```
$ cat > $SC/.moai/specs/SPEC-FIXTURE-001/spec.md   # verbatim fixtureSpecMD output
  (frontmatter per drift_characterization_test.go:109-130, body "# SPEC-FIXTURE-001")
$ cd $SC && git init -q -b feature . && git add -A && git commit -qm "chore: fixture"
$ git rev-parse --verify main
fatal: Needed a single revision            (rc=128 — the AC-001 Given holds: no main/master)
$ bin/moai spec lint
SEVERITY  CODE               FILE          LINE  MESSAGE
--------  ----               ----          ----  -------
WARNING   MissingExclusions  …/spec.md     1     'Out of Scope' section missing — minimum one item in Out of Scope subsection required [grandfathered era — downgraded to warning]

0 error(s), 1 warning(s)
rc=0
```

The `✓ No findings` line **does not print**, because `fixtureSpecMD` has no `Out of Scope` section
and `MissingExclusions` fires. `printTable`'s short-circuit (`internal/cli/spec_lint.go:115-118`)
therefore never runs, and AC-SLGB-001 Then#2 / AC-SLGB-004a / AC-SLGB-004b become vacuous — the
exact outcome the AC's own closing sentence warns about ("다른 finding 이 섞여 있으면 short-circuit
이 애초에 발화하지 않아 2번 단언이 공허해진다"). The premise is **reachable**, not impossible — the
same fixture with an Out-of-Scope section prints it:

```
$ bin/moai spec lint          # fixture + "### Out of Scope — nothing" + one bullet
✓ No findings — all SPEC documents are valid
rc=0
```

(That second run also independently corroborates the SPEC's whole thesis: in a repo with no `main`
ref, `StatusGitConsistency` emitted nothing and the tool declared every document valid.)
— Severity: **major** — Class: **blocking** — Required fix: in AC-SLGB-001 and AC-SLGB-004, replace
the bare "`fixtureSpecMD` 선례를 따른다" with the actual zero-finding requirement — the fixture SPEC
must additionally carry an `Out of Scope` section with at least one item — and record the verified
zero-finding output as the RED cell (see D2).

**D2.** `acceptance.md`: whole file — **No acceptance criterion is adopted under the two-cell
discipline that `.claude/rules/moai/development/verification-completeness.md` §2 makes [HARD] for
`**/.moai/specs/**`.** That rule requires each AC to be authored as a pair: a **RED-now cell** (the
criterion observed red on the pre-implementation tree, pinned to the tree it was measured on, with
the reason it is red stated) and a **green path cell** (which milestone flips it and what the
passing output becomes). `acceptance.md` carries neither, for any of the 11 criteria, and carries
no tree pin at all:

```
$ grep -nE '[0-9a-f]{7,40}' .moai/specs/SPEC-SPECLINT-GITBLIND-001/acceptance.md
(no output, exit 1)
```

What the file does carry is prose *instructions to observe red later* (AC-001/004: "구현 전 그 줄이
찍히는 것을 관측한 뒤 구현한다") and two mutation clauses (AC-005/008). Those are the right
substance and they are why this is major rather than critical — but a future obligation is not a
recorded starting observation, and §2 is explicit that "a green path alone is a promise with no
starting observation". D1 is the direct cost of the omission: had AC-001's RED cell carried a
command and its verbatim stdout, the `MissingExclusions` line would have been on the page at
authoring time. I do not assess §2.1's four-element form, because no criterion in this SPEC is
classified release-blocking and §2.1 is conditional on that classification.
— Severity: **major** — Class: **blocking** — Required fix: give each of AC-SLGB-001…010 a RED-now
cell (command, its output, the tree SHA `bff12c54d`, and why it is red) and a green-path cell
naming the closing milestone. AC-SLGB-005 and AC-SLGB-008 are green today and must say so in their
RED cell, citing the mutation as the substitute red they already define. AC-SLGB-011 is CI-only and
records that as its reason.

**D3.** `plan.md`:L143-144 (§H 상호참조) — **The iter-3 "전량 재측정" completeness claim is false, and
the surviving stale coordinates are self-contradictory on their face.** `spec.md` HISTORY v0.4.0
asserts "인용 좌표를 전량 재측정" and "그 밖의 파일 인용은 불변으로 확인"; `progress.md` §E.1 iter-3
repeats it. `plan.md` §H refreshed only the outer range and left all three parentheticals at their
pre-absorption values:

```
$ sed -n '143,144p' .moai/specs/SPEC-SPECLINT-GITBLIND-001/plan.md
- `internal/spec/lint.go:1306-1342` — `StatusGitConsistencyRule.Check`
  (terminal 조기반환 `:1299-1301`, err skip `:1305-1308`, emission `:1310-1319`)
```

Two of the three sub-ranges fall **outside** the parent range stated on the same line (1299 < 1306)
— visible without opening the tree. On this tree they resolve to unrelated constructs:

```
$ awk 'NR>=1297 && NR<=1312' internal/spec/lint.go
1299-1301: // Severity: advisory warning — the git-implied status is a heuristic … (doc comment)
1305:      (blank)
1306:      func (r *StatusGitConsistencyRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
1307-1308: fm := doc.Frontmatter / var findings []Finding
$ awk 'NR>=1310 && NR<=1320' internal/spec/lint.go
1310-1313: if fm.ID == "" || fm.Status == "" { … return nil }
1318-1320: if terminalStatusEnum[fm.Status] { return nil }
```

Correct values, measured: terminal early return `:1318-1320`, err-skip `:1324-1327`, emission block
`:1331-1340`. `plan.md` §F M1.4 already uses the correct `:1324-1327`, so the file contradicts
itself. An implementer following §H's "err skip `:1305-1308`" lands on the function signature. A
second staleness rides along: `plan.md`:L16 still declares the baseline as `develop 1e5199b88`,
which the absorption of `9328a5242` superseded. — Severity: **major** — Class: **blocking** —
Required fix: correct the three `plan.md` §H sub-coordinates to `:1318-1320` / `:1324-1327` /
`:1331-1340`, update `plan.md`:L16 to the current tree, and amend the v0.4.0 HISTORY line so the
completeness claim matches what was actually re-measured.

**D4.** `spec.md`:§2.2, `plan.md`:§F M1.3 — **The once-per-run suppression flag can silently lose
the only Info of a run.** The design attaches `StatusGitUnreachable` to a particular SPEC document
and suppresses further emission via a per-run flag. In `Lint()`, per-document findings pass through
`applylintSkip(ruleFindings, doc.LintSkip)` at `lint.go:236` *after* the rule returns. If the
document that happens to trigger first carries `lint.skip: [StatusGitUnreachable]`, its finding is
dropped while the run-level flag is already set — no later document re-emits, and the run reverts
to exactly the silent state REQ-SLGB-001 exists to end. Neither §2.2 nor §F M1.3 addresses the
ordering. iter-1's D2 required-fix option (b) — the directory-level `dirFindings` channel, on this
tree assigned at `lint.go:204` — would not have this hazard. No SPEC carries such a skip entry
today, which is why this is minor rather than major. — Severity: **minor** — Class: **optional** —
Required fix: state in §2.2 that the suppression flag is set only after the finding survives
`applylintSkip`, or route the finding through `dirFindings`.

**D5.** `progress.md`:§E.1 iter-3 (t382 disproof) — **The disproof is sound for the code it cites
but is stated one notch wider than it was verified.** I checked every step and each holds on this
tree: `StatusGitConsistency` is emitted with `Advisory: true` unconditionally at `lint.go:1335`;
`applyEraDemotion`'s Warning branch (`:307-308`) only re-sets an already-true flag and removes no
finding; `eraDemotableCodes` (`:272-275`) contains only `MissingExclusions` and `FrontmatterInvalid`;
the switch has no Info case, so a future `StatusGitUnreachable` passes through; and `--strict`
escalation is gated on `!f.Advisory` at `lint.go:61`. The claim as written — "t382(era.go H-3)가
어느 방향으로 착지하든" — is a statement about a change whose content is not yet fixed. It holds for
any t382 confined to the era classifier or to `eraDemotableCodes`, and would not survive a t382
that restructured `applyEraDemotion` itself. — Severity: **minor** — Class: **optional** —
Required fix: bound the claim to "any t382 change confined to `era.go`'s classifier and
`eraDemotableCodes`".

**D6.** `acceptance.md`:L52-53 (AC-SLGB-002) — **A shallow mutant satisfies it.** AC-SLGB-002
requires only that the message contain the strings `main` and `master`. An implementation that
hard-codes a candidate-list literal into the message without ever attempting resolution passes it
while violating REQ-SLGB-002's payload ("시도한 후보 ref 이름"). The gap is narrow — the chain *is*
a fixed ordered list, so the observable difference is thin — and AC-SLGB-007 constrains the
resolution behaviour separately. Recording it because the mutant probe is writable.
— Severity: **minor** — Class: **optional** — Required fix: additionally assert that the message
names `origin/main` and `origin/master`, so the message is tied to the four-candidate chain
AC-SLGB-007 verifies rather than to two bare words.

**D7.** `acceptance.md`:L189-190 (커버리지 표) — AC-SLGB-011 traces to no REQ ("— (반-공허 관측)").
It is declared, not hidden, and every REQ is covered in the other direction. Recording it because
Group 4 AC-4 counts orphaned ACs. — Severity: **minor** — Class: **optional** — Required fix:
either map it to REQ-SLGB-008 (it observes that the checkout fix took effect) or state in the table
why it is deliberately REQ-less.

---

## iter-1 blocking defects — closure verdict, one by one

| iter-1 defect | Verdict | Evidence |
|---|---|---|
| **D1** — error shapes are three, not two | **CLOSED** | `spec.md` §2.1 carries the three-shape table with per-shape emission decisions. I verified all three return points and their literal strings on this tree: `drift.go:312` `git log failed: %s`, `:316` `no git history found for %s`, `:366` `no classifiable commit within window of %d for %s` with `gitLogWindowSize = 50` at `drift.go:284`. The table's "window of 50" matches. |
| **D2** — output-volume amplification unexamined | **CLOSED** | `spec.md` §2.2 makes the decision explicit (repository-level cause → one finding per `Lint()` run) and AC-SLGB-003 bounds the count at exactly 1 against a 10-SPEC fixture. iter-1 demanded "make the decision explicit … add an AC that bounds the count either way"; both are done. Residual: D4 above. |
| **D4** — REQ-SLGB-002 untraced | **CLOSED** | AC-SLGB-002 exists (`acceptance.md`:L46-53) and asserts the message content directly. The implementation can no longer be omitted with all ACs still passing. Residual: D6 above (the assertion is shallower than the requirement). |
| **D7** — terminal-status fixture trap | **CLOSED** | AC-SLGB-001's Given now names the constraint ("`terminalStatusEnum` … 에 속하지 **않는**"), and AC-SLGB-005 carries it as a [HARD] constraint with the mechanism spelled out. I verified the trap is real and the cited coordinate is correct: `lint.go:1318-1320` returns `nil` for terminal statuses before `getGitImpliedStatus` is reached. |
| **D8** — two ACs cannot show RED | **PARTIALLY CLOSED** | The two *specific instances* are closed: the over-emission guard (now AC-SLGB-005) and the cache guard (now AC-SLGB-008) each carry a [HARD] mutation clause naming the mutant and the expected failure, and `plan.md` §E.3 names both as green-today. I confirmed both are green today — `grep -rn 'StatusGitUnreachable' internal .github` exits 1, and `gitquery_cache.go:96-100` already memoizes via `mainBranchSet`. **The defect class is not closed**: the acceptance artifact still records no observed red for any criterion (D2), and I measured a *new* instance of D8's exact failure class in AC-SLGB-001/004 (D1). |

---

## Answers to the dispatch's specific challenges

1. **Refreshed citations.** The five `spec.md` refreshes are all correct on this tree
   (`:1306`, `:1318-1320`, `:1324-1327`, `:1335`, `:296-312`), as are every citation into
   `drift.go` (`:68`, `:300`, `:312`, `:316`, `:366`), `gitquery_cache.go` (`:96-100`, `:102`,
   `:113`), `cli/spec_lint.go` (`:113`, `:115-118`, `:116`), `lint_ownership.go` (`:429` — the
   emission literal carries no `Advisory` field, as claimed), `drift_characterization_test.go`
   (`:53`, `:55`, `:98`, `:103`, `:280`), and `ci.yml`'s six `fetch-depth: 0` lines
   (`:129 :264 :382 :431 :486 :543` — exact). **Three coordinates were missed**: `plan.md` §H (D3).
   Two range starts are off by one in a harmless direction (`gitquery_cache.go:88-118` begins on the
   last doc-comment line, `lint_ownership.go:380-400` opens inside the preceding block); recorded
   here, not as defects.
2. **The disproved t382 coupling.** Verified sound; see D5 for the one over-statement.
3. **iter-1's five blocking defects.** Four closed, one partially — table above.
4. **AC RED-now adoption and the mutant probe.** No AC carries a RED-now cell (D2). On the mutant
   probe, the AC set is stronger than it looks in one important direction and weaker in another:
   an unconditional emitter (fires exactly one Info per run regardless of git state) satisfies
   AC-001/002/003/004a/004b/006 — and is killed by **AC-SLGB-005**, which demands zero findings in
   a non-shallow main-resolved repo. That guard is load-bearing and correctly placed. The mutant I
   *could* write is the message hard-coder (D6).
5. **The `✓ No findings` assertion.** The RED baseline is **reachable but not as the AC specifies
   it** — measured both ways in D1. The corpus's 1096 warnings are irrelevant to it, since the ACs
   run against a `t.TempDir()` fixture, not this repository.
6. **Tier M budget.** REQ = 9 (≤16). AC = 11 headings, 12 counting the 004a/004b split (≤16). Within
   budget on both axes.
7. **AC-SLGB-011.** Acceptable as a declared plan-phase gap, not a defect. It names its judgment
   place, forbids marking it PASS in run-phase, and `plan.md` §G forbids reading CI green as its
   evidence. Its only blemish is traceability (D7).

---

## Independent re-measurement of the consumed figures

I re-derived the dispatch's key numbers rather than consuming them:

```
$ tail -3 .moai/reports/t371/lint-local-35bc0715f.txt
0 error(s), 1096 warning(s)
$ grep -c 'StatusGitConsistency' .moai/reports/t371/lint-local-35bc0715f.txt
18
$ wc -l .moai/reports/t371/lint-local-1e5199b88.txt
1098
$ tail -2 .moai/reports/t371/lint-local-1e5199b88.txt
2 error(s), 1091 warning(s)
```

The unit-error correction is confirmed independently: `1098` is the pre-absorption report's line
count, and that tree's own declared total is `2 error / 1091 warning`. `1091 + 5 = 1096` checks out.
StatusGitConsistency = 18, unchanged. I did **not** re-run the full corpus lint myself, so the
`SyncSHASlotFormat 0 → 5` per-rule delta attribution remains attributed to
`.moai/reports/t371/remeasure-35bc0715f.md`, not to me (see Gaps).

---

## Gaps (explicitly not observed)

- I did not re-run `moai spec lint` over the full repository corpus. The `0 error / 1096 warning`
  and `StatusGitConsistency = 18` figures were verified against the saved report file, not
  regenerated. The per-rule delta table is consumed from `remeasure-35bc0715f.md`.
- I did not verify that the 18 SPEC ID **set** is unchanged; I verified only the count.
- I did not construct a genuinely shallow fixture (`--is-shallow-repository == true`) to test
  AC-SLGB-004a/b's premise. Their fixture-construction cost is unassessed.
- I did not check the `--json` / `--sarif` output paths. The SPEC forbids any AC from resting on
  them, so no AC depends on the gap.
- I did not audit `classification-18.md`'s per-SPEC classification of the 18 findings.
- `mcp__moai__audit_multi` / cross-model backends were not invoked; this is a Claude-only audit.

## Residual risk

- Three cards touch `internal/spec/lint.go`. Whichever lands first re-shifts the others'
  coordinates — D3 is a live instance of that class arriving *before* any of the three has landed.
  A coordinate-refresh pass will be needed again after integration regardless of this verdict.
- D1's fix changes only the fixture prose, but the fixture must be re-measured after the change to
  confirm the zero-finding state — a new rule arriving from `develop` (as `SyncSHASlotFormat` just
  did) can break it again silently.

## Procedural notes

- **Iteration ceiling.** `harness.yaml` `plan_audit_tier_ceilings` sets Tier M = 2. This is
  iteration 3. It is within the absolute 3-iteration cap but **exceeds the Tier-resolved ceiling**;
  the orchestrator should treat any further iteration as requiring the explicit user-override
  branch of the retry-loop contract.
- **No score regression.** iter-1 scored 0.75 against a 0.75 threshold; iter-3 scores 0.79 against
  0.80. The score improved; the threshold moved further. No STOP escalation signal.

## Recommendation

FAIL, with three blocking defects that are all cheap to close and none of which require rethinking
the SPEC's design. In priority order:

1. **D1** — `acceptance.md` AC-SLGB-001 and AC-SLGB-004: replace the `fixtureSpecMD` precedent
   reference with the explicit zero-finding requirement (fixture SPEC must carry an `Out of Scope`
   section with ≥1 item). Both measurements needed to write the corrected clause are in D1 above.
2. **D2** — `acceptance.md`: add a RED-now cell and a green-path cell to AC-SLGB-001…010, pinned to
   `bff12c54d`. AC-SLGB-005 and AC-SLGB-008 state green-today plus their mutation as substitute red;
   AC-SLGB-011 states CI-only as its reason.
3. **D3** — `plan.md` §H: correct the three sub-coordinates to `:1318-1320` / `:1324-1327` /
   `:1331-1340`; update `plan.md`:L16's baseline; amend the v0.4.0 HISTORY completeness claim.

D4-D7 are optional and left to the orchestrator's discretion. None of them changes the design, and
routing all of them into a revision would add speculative surface the SPEC has not claimed.

A confirming re-audit should be scoped to this enumerated defect delta plus a regression check over
iter-1's D1/D2/D4/D7/D8, not repeated from scratch.
