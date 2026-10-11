auditor-model: glm-5.3-flash

verdict: FAIL
audited_sha: 17d0f9bac

# SPEC Review Report: SPEC-ALWAYS-LOADED-BUDGET-001
Iteration: 5/3+2 (leader-authorized final re-read)
Verdict: FAIL (final — no further repair round proposed)
Overall Score: 0.82 (exact mean 0.825; iter4 0.81 → iter5 0.825, improving, no STOP signal; Tier L threshold 0.85 not met)
Plan Artifact Hash: sha256 `2beebb7e878947956f92028086175b6a12c24ddbc8f507f5eff6d396a498ac84` (spec·plan·acceptance·design·research·decision-index concatenated in that order)
Auditor Version: plan-auditor (card t1469, iter5)

`audited_sha` = 17d0f9bacd7a64f572a2b2aa11b34a5f7a2a6cdf, measured directly by this session via `git rev-parse` (iter4 could not run git; this session could — no guard refusals on git).

Reasoning context ignored per M1 Context Isolation. Operator/leader decisions (Q1/Q5/Q9/Q10) treated as settled premises per delegation.

## 1. Claim

All seven iter4 defects (N4-1..N4-7) are substantively closed as specified — the repair (commit `17d0f9bac`, the only commit since the pinned tree `2771626b5`, touching only the SPEC directory) did exactly what iter4's §Recommendation asked for each item, with 25/25 REQ/AC numbering unchanged.

However, the N4-2 repair **introduced one new blocking-class defect cluster (N5-1)**: the companion:-location exception rows it created are contradicted by six mechanically-binding statements that still carry the old rationale-only restriction, and the exception has no anchor-source predicate — so the ledger is un-greenable as written for the M3 split scenario it mandates, while honoring the exception naively reopens the exact REQ-ALB-012 lossless bypass that iter1–4 fought. Both required cross-model backends independently found the same cluster (claude P1, codex P2; `disagreement_flag: false`). Combined with the aggregate 0.825 < 0.85 → **FAIL**.

This is the last authorized read: no repair round is proposed. Residual work is one coherent cluster (§ Recommendation).

### Must-Pass (all PASS)
- [PASS] MP-1: REQ-ALB-001..025 all present, no gaps/dups (order anomaly 023–025 between 011/012 remains — known optional D-N8). AC-ALB-001..025 all present.
- [PASS] MP-2 (requirement layer only): 21 GEARS forms + 4 legacy EARS shall-not (REQ-008/013/014/019) within the backward-compat window (through 2026-11-22). ACs are verification-layer Given-When-Then — not graded here.
- [PASS] MP-3: spec.md:L2–L15 carries all 12 canonical fields, correct types, no rejected aliases. Corroborated by `go run ./cmd/moai spec lint SPEC-ALWAYS-LOADED-BUDGET-001 --strict` (built from THIS tree): `0 error(s), 0 warning(s)`, exit 0 (one INFO: OwnershipTransitionUnmeasured — Info never escalates).
- [N/A→PASS] MP-4: not a multi-language-tooling SPEC; REQ-ALB-019 enforces 16-language neutrality of template text.
- [PASS] MP-5: all 5 referenced SPECs `status: completed` (DIET-001, DIET-002, HEADROOM-001, INSTRUCTION-BUDGET-SCOPE-001, INSTRUCTIONS-BUDGET-001 — 5 grep lines observed). No D7 BLOCKING.
- [PASS] MP-6: `grep -c syscall spec.md` → `0` (exit 1).
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → no output, exit 1.
- [PASS] MP-8: all 10 RED-now groups re-executed at the worktree root, verbatim stdout + exit codes matching the table (see §2). Pin is a commit SHA (`2771626b5`); `git diff --name-only 2771626b5..HEAD` shows only the 7 SPEC-dir files — template/hook bytes identical to the pin, so the pins are valid on the current tree.
- [PASS] MP-9: CN-4 verb refused by worktree guard (GAP, see §5) → hand comparison: plan order M0(L55)→M1(L68)→M2(L75)→M3(L85)→M4(L91)→M5(L96), 0 `Exit:` bindings; ordering clauses — AC-002 "M1 RED lands after the anchor" ≅ plan M1 L71; AC-025 "floor > 115,000 → report before any M1 commit" ≅ plan L65; AC-025 "re-pin commit precedes first rule edit" ≅ plan §C.2 step 4 (L50). No conflict.

### Scores
| Dim | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Fixed: stale-term/stale-ref set of iter4 gone (`구속 절` now only spec.md:L26 HISTORY; `AC-ALB-002 RED-now` gone). Deduction: the central location vocabulary (§B L64) contradicts the ledger-scope exception (§B L62) across six normative surfaces (N5-1) — an implementer cannot make the ledger green for the mandated M3 split without violating a stated criterion or silently special-casing. Intent is recoverable (L62's own clarifying clause), so not 0.50. |
| Completeness | 0.80 | — | N4-1/N4-2 scope gaps closed (membership checks, widened ledger scope, M3 split-lossless condition). Deduction: the exception's case-analysis is incomplete everywhere else — vocabulary, REQ-012/015/017, AC-021(1), design §1/§5/§6 — plus the `entry_points`/`delivery` vocabulary gap for companion-origin units and the AC-021(2) new-companion exclusion gap (all N5-1 sub-points). |
| Testability | 0.80 | 0.75+ | All 10 RED groups reproduce verbatim; AC-015 pinned to `/usr/bin/grep` with declared order-insensitive comparison; AC-019 gained the always:-membership mutation fixture; AC-025 gained the anchor-kind git-log check. Deduction: AC-021(1) is precisely binary-testable but unsatisfiable-as-written in the mandated scenario; AC-019 green condition states only 2 of the 3 location after-text lookups (omits `companion:`); no positive/negative fixture pair for the exception class. |
| Traceability | 0.95 | — | Hand-verified bidirectional mapping: all 25 REQs have ≥1 AC (REQ-001→AC-001/002 … REQ-025→AC-019), 0 orphans, all AC→REQ references valid; AC-025 now also maps REQ-ALB-015. COLLECTED: 25 (manual enumeration; verb refused, see Gaps). |

Mean (0.75+0.80+0.80+0.95)/4 = 0.825 < 0.85.

## 2. Evidence

Commands executed by this session at the worktree root (HEAD `17d0f9bac`, branch `WT-always-loaded-budget`, tree clean):

- `git log --oneline 2771626b5..HEAD` → single commit `17d0f9bac docs(SPEC-ALWAYS-LOADED-BUDGET-001): close iter4 defects N4-1..N4-7, re-pin evidence to absorbed tree 2771626b5 (card t1469)`; `git diff --name-only 2771626b5..HEAD` → only the 7 files under `.moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/`.
- RED-now re-execution (verbatim stdout, exit codes):
  - AC-001..007 `grep -rl TestDeployedAlwaysLoadedCharBudget internal/template` → (empty) exit 1 ✓
  - AC-008 `/usr/bin/grep -c 'Intentionally always-loaded' <factory-dispatch.md> <cross-session-messaging.md>` → `factory-dispatch.md:1` / `cross-session-messaging.md:1` exit 0 ✓ (byte-identical to table)
  - AC-009/011/012/013 `grep -l -e role-core -e factory-dispatch internal/hook/session_start_factory.go internal/hook/factory_messages.go internal/hook/subagent_start.go` → (empty) exit 1 ✓
  - AC-014 `grep -rl 'moai:role-core-start' internal/template/templates` → (empty) exit 1 ✓
  - AC-015 `/usr/bin/grep -c moai:role-rules-required <manager-lead.md> <moai-factory-foreman/SKILL.md> <gtd.md>` → `manager-lead.md:0` / `SKILL.md:0` / `gtd.md:0` in argument order, exit 1 ✓ — exactly as the repaired table states (N4-4 verified against live output)
  - AC-019/020/021 `ls internal/template/testdata/binding_ledger.json` → stderr `ls: internal/template/testdata/binding_ledger.json: No such file or directory` exit 1 ✓
  - AC-022 13-path `grep -l '^paths:' …` → (empty) exit 1 ✓
  - AC-023 `grep -rl TestDeployedRuleSectionRefsResolve internal/template` → (empty) exit 1 ✓
  - AC-024 `grep -c -e factory-dispatch -e cross-session-messaging internal/template/rule_template_mirror_test.go` → `1` exit 0 ✓
  - AC-025 `grep -c 'pending run-phase' .moai/specs/SPEC-ALWAYS-LOADED-BUDGET-001/progress.md` → `2` exit 0 ✓
- Greps: `구속 절` → only spec.md:L26 (HISTORY narration — legitimate); `syscall` spec.md → 0; `NEEDS CLARIFICATION` plan/research → none; `AC-ALB-002 RED-now` acceptance.md → none (N4-5 closed).
- Reference SPEC statuses → 5× `status: completed`.
- `factory-dispatch-detail.md:L150` at HEAD: "**Write-capable sub-agents spawned in parallel MUST carry `isolation: "worktree"`**" — confirmed live; the exception rows mandated by N4-2's fix are non-hypothetical.
- Lint: `go run ./cmd/moai spec lint SPEC-ALWAYS-LOADED-BUDGET-001 --strict` → `0 error(s), 0 warning(s)`, exit 0.
- Plan-artifact hash: `cat spec.md plan.md acceptance.md design.md research.md decision-index.md | shasum -a 256` → `2beebb7e878947956f92028086175b6a12c24ddbc8f507f5eff6d396a498ac84`.

audit_multi (project_root=this worktree, target=baseBranch, gates claude+codex required / glm advisory):
- `overall_verdict: fail`, `disagreement_flag: false`, `participant_count: 2`, `fail_open_backends: [glm]`, `audit_receipt: rcpt-ce0997667c1f9cf5f0ef4a8a`.
- `residual_risk_note` (verbatim): "required-backend FAIL: claude, codex".
- claude (P1×1, P2×3, P3×4): P1 = the companion:-exception contradiction cluster (six surfaces enumerated, incl. REQ-012 itself and the smuggling-bypass observation); P2 = entry_points/delivery vocabulary gap; AC-021(2) new-companion exclusion; Q10 resume-plan divergence. P3 = stale refs (spec §H count, acceptance L15, research header, design §6), AC-019 missing companion:-side fixture. **Auditor correction: claude's P3 claim that progress.md:L3 is stale is an overstatement** — L3 is explicitly qualified as the historical plan base ("run M0 re-anchors", §C.2), consistent with acceptance L14.
- codex (P2×2): independently flagged the same exception contradiction (citing the live `factory-dispatch-detail.md` MUST at HEAD) and the same Q10 divergence (decision-index L73 + progress L11 vs plan M3).
- glm: `inconclusive` — "z.ai response carried no text content" (advisory gate; fail-open).

## 3. Baseline-attribution

- Measured in this run, this tree: worktree `agent-a6d780d7ba07bc24c`, branch `WT-always-loaded-budget`, HEAD `17d0f9bac` (full SHA observed directly). All RED-now commands, greps, lint (`go run` — built from this tree, satisfying tool-provenance §2.2), and the hash were executed here.
- Pinned evidence tree `2771626b5` is HEAD's parent; the only intervening commit is docs-only (verified via git diff file list), so template/hook bytes at HEAD equal the pin — RED-now pins carry over.
- audit_multi's build is `0732cc699` (ancestor of HEAD, per its own build_lag) — its verdicts are diff/tree-content based, so binary lag does not affect them; its inability to execute the RED-now commands is covered by this session's direct re-execution.

## 4. Defects-or-Closure

**iter4 defect closure (delta scope)**

| Defect | Status | Evidence |
|---|---|---|
| N4-1 always:/companion: membership unchecked | CLOSED | spec L96 REQ-ALB-015 gains both failure conditions ("`always:` row's path is absent from the deployed-surface member list…", "`companion:` row's path is not a deployed file carrying a top-level `paths:` key"); §B L64 vocabulary carries both; AC-019 L40 mutation fixture = binding row location → role-gated rule full-body path |
| N4-2 edited companion out of ledger scope | CLOSED (primary fix) | §B L62 widened to "배포 규칙 파일 전부 — 상시 표면 파일과 최상위 `paths:` 를 가진 기존 companion 파일을 함께 포함"; companion: exception rows defined; REQ-013 "new relocation only" clarification; plan M3 L87 split-lossless condition (the minimum alternative) also landed. **But introduced N5-1 below** |
| N4-3 AC-019 stale scope term | CLOSED | AC-019 green = "원장 대상 파일 전부의 모든 단위(spec §B 단위 경계)에 정확히 1행"; design L28 "원장 대상 파일 안에 있으므로"; `구속 절` grep → HISTORY only |
| N4-4 AC-015 stdout order | CLOSED | `/usr/bin/grep` pin + argument-order stdout + preamble order-insensitive convention (L14); live re-run byte-matches |
| N4-5 L15 stale example ref (optional) | CLOSED | `AC-ALB-002 RED-now` gone; re-pointed (residual imprecision folded into N5-3) |
| N6 anchor-kind same-row (optional) | CLOSED | AC-025 green carries anchor-kind `git log` check; REQ-ALB-015 added to AC-025 mapping |
| N4-7 fragment rule skills (optional) | CLOSED | AC-021(2) fragment targets now "배포 companion 파일과 배포 skill 파일" |
| REQ/AC 25/25 unchanged | CONFIRMED | Direct read: REQ-ALB-001..025, AC-ALB-001..025, no renumbering |

**New defects (iter5)**

N5-1. COMPANION-EXCEPTION-CONTRADICTION-CLUSTER (NEW, introduced by the N4-2 repair)
- Location: spec.md:L62 (exception) vs L64 (vocabulary), L93 (REQ-ALB-012), L96 (REQ-ALB-015), L98 (REQ-ALB-017), L63 (entry_points mandate); acceptance.md:L42 (AC-021(1), whose carve-out list names only rationale-companion: and role-core: rows); design.md:L17/L19/L80.
- Problem (four sub-points, one cluster):
  (a) L62 mandates exception rows — "원래 companion 에 있던 `binding`·`normative` 단위는 위치 `companion:` 를 허용하는 예외 행으로 원장에 들어간다" — but six mechanically-binding statements still forbid exactly that: L64 "`companion:` 은 종류가 `rationale` 행에만 허용된다"; REQ-015 "when … a `binding` or `normative` row carries a `companion:` location"; REQ-012 "shall … be present in either the deployed always-loaded surface or a role core" (companion-origin units are in neither); REQ-017 "the companion shall receive only rationale…"; AC-021(1) "…행 0"; design L17/L19/L80. A correct ledger for the mandated M3 `factory-dispatch-detail.md` split fails REQ-015/AC-021(1) as written. Non-hypothetical: that file carries live MUST obligations (L150 verified at HEAD).
  (b) The exception has no anchor-source predicate: no field records where a unit lived at the anchor, so an always-surface binding unit relabelled `companion:` is indistinguishable from a legitimate exception row — and AC-021(2) does not check ledger-scope companions. Honoring the exception naively reopens the exact REQ-ALB-012 lossless bypass that iter1–4 FAILed on.
  (c) `entry_points`/`delivery` vocabulary (L63: `always`/`REQ-ALB-007`/`REQ-ALB-024`) has no truthful value for a companion-origin unit delivered by path contact; REQ-025 (L90) read literally would push such units into the always-on stub — inflating the budget this SPEC exists to cut.
  (d) AC-021(2) excludes only lines already in "the same companion" at the anchor; a NEW companion created by the split (design §6, plan M3) is neither ledger-scope nor "the same companion", so legitimate relocation inside the split trips the fragment check.
- Severity: major. Class: **blocking**. Convergence: this audit + claude P1 (conf 0.93) + codex P2 — `disagreement_flag: false`.
- Required fix (for the record, not a proposed round): give the exception a mechanical predicate (anchor source file carries top-level `paths:`, frozen at the anchor like anchor kind) and propagate it through L64, REQ-012/015/017, AC-021(1)/(2), design §1/§5; add delivery value or REQ-025 exemption for companion-origin rows; add both-direction mutation fixtures. Alternative: Q10(a) scope reduction, which removes the cluster entirely.

N5-2. Q10-RECORD-DIVERGENCE (NEW, minor)
- Location: decision-index.md:L68–73 + progress.md:L11 vs plan.md:M3(L87) + HISTORY 0.6.0 row.
- Problem: the recorded resume plan is (a) scope reduction (moves the detail-split out, removing N4-2) + (c) STE-lite; the repair instead kept the split and widened the ledger (iter4's N4-2 primary fix — the path this audit's delegation authorized) and silently dropped STE-lite, recording no Q10 amendment. The delegation's authorization settles the substantive question; what remains is a record-keeping inconsistency between hashed plan artifacts.
- Severity: minor. Class: optional. Required fix: record a Q10 amendment (or Q11) noting the adopted path; or strike the stale resume plan.

N5-3. STALE-REFS-POST-REPAIR (NEW, minor)
- spec.md:L160: "결정 9건(해소 6…)" — decision-index now has Q1..Q10 (10 decisions, 7 resolved; Q10 added after iter4, unreflected).
- acceptance.md:L15: "바로 아래 항목의 `[no tests to run]` 판정" — the example is at L17; L16 (ledger fixture path) sits between. Off by one.
- research.md:L11: header says "§1.1 만" deviates from the `b5815ca80` basis, but §1.2/§4/§5 also carry `2771626b5`/2026-10-04 re-measurements.
- design.md §6 (L87): split-target list ("상시 core 로 남는 11개 규칙과 `CLAUDE.md`") omits the ledger-scope companion splits (factory-dispatch-detail, spec-workflow per anchor re-measurement).
- Severity: minor. Class: optional.

**Regression check (iter4 → iter5)**: N4-1..N4-7 all RESOLVED (table above); no prior-iteration defect remains unresolved. Score 0.81 → 0.825, improving — no LEAN STOP signal. Stagnation: no defect unchanged across iterations; the ledger-scope defect family narrowed again (iter4: scope holes; iter5: the repair's own exception is un-propagated).

## Recommendation (final escalation — leader/operator; no further repair round authorized)

All iter4 defects are closed; what remains is **one coherent cluster (N5-1)** plus two minor record-keeping items (N5-2, N5-3). Dispositions for the leader/operator (consistent with the menu already recorded in decision-index Q10):

1. **Scope reduction (Q10(a))** — move the `factory-dispatch-detail.md` split to a separate card. This eliminates N5-1 entirely (no companion-origin exception rows needed), reverts the ledger scope to iter4's surface, and is the lowest-risk path. N5-2/N5-3 are doc-only touch-ups.
2. **PASS-with-debt** — accept with the documented debt that run-phase M0 will hit the N5-1 contradiction when building the ledger (a loud failure: REQ-015/AC-021 go red on legitimate rows) and will require a blocker-report round-trip to manager-spec before M3 can proceed; risk (b) — the smuggling bypass — is only reachable if the implementer silently special-cases the exception, which the loud failure makes unlikely but not impossible.
3. **Explicit user override** — authorize one more targeted repair round (leader's call; this auditor was authorized for exactly one re-read and issues no further rounds).

## 5. Gaps

- CN-4 and traceability awk verbs refused by the worktree-isolation guard ("runs awk with -f … cannot be shown not to be git") — same refusal shape as iter4. Substitutes: full hand comparison for both (milestone order + ordering clauses enumerated in MP-9; 25/25 REQ→AC mapping table in Scores/Traceability). COLLECTED count is a manual enumeration, not the verb's.
- glm backend inconclusive ("z.ai response carried no text content") — advisory gate, fail-open; convergence rests on claude+codex, both of which returned fail.
- claude backend could not execute the RED-now commands from the diff ("stay inconclusive" per its next_steps) — covered by this session's direct verbatim re-execution (§2).
- `audit_receipt` issued: `rcpt-ce0997667c1f9cf5f0ef4a8a` (cited in the verdict line).
- No other refusals: git, grep, ls, go run all executed normally in this session (unlike iter4).

## 6. Residual-risk

- audit_multi residual_risk_note (verbatim): "required-backend FAIL: claude, codex".
- The ledger-scope defect family has produced a new member every round it was touched (iter1–4: scope holes; iter5: un-propagated exception in the repair). Any disposition that keeps the companion split in scope carries the risk of a further loophole in the exception machinery; Q10(a) scope reduction removes the family's habitat.
- research §2 estimate still suggests M0 likely stops at floor > 115,000 (Q6 path); widening ledger scope (if the exception machinery is repaired instead of scoped down) further increases M0 workload.
- The env `grep` shell-function wrapper remains an order-instability source; acceptance now pins multi-file commands to `/usr/bin/grep` with a declared order-insensitive comparison, and AC-009/AC-022 (plain `grep`, multi-file, empty expected output) are unaffected in practice — cosmetic deviation from the preamble's own convention, noted, no effect observable.
