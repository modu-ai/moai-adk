auditor-model: claude-haiku-5-5

# Plan audit iteration 2: SPEC-CAPACITY-RECLAIM-001 (factory card t1631)

This is the LAST audit spawn at Tier M (tier ceiling: two spawns).

## Raw audit_multi return

> {"per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"fail","summary":"Verdict: fail\n\n- [P1] D14 미해결: 출시 차단 기준의 RED 증거가 MP-8을 충족하지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:15\n  RB 셀의 종료 코드를 run-phase로 미뤘다. f59ae36be에서 AC-004의 명령은 PASS했고, AC-011·013의 테스트 명령은 `[no tests to run]`으로 종료했다. 확장 fixture의 실패는 관측되지 않았다.\n  신뢰도: 높음. 권고: 감사 전에 네 요소가 갖춰진 RED 셀을 기록하고 재실행해. Tier M 임계값은 0.80이며, 점수와 무관하게 MP-8 실패는 통과를 차단한다.\n\n- [P1] 거절 판정도 Q1·Q4 변경의 실행 조건을 만족한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11\n  M1은 Q4에 “operator verdict”가 있으면 적용하며, Q1도 같은 조건을 사용한다(:21). 거절도 판정이므로, 승인된 경우만 변경하도록 한 REQ-CAP-004·006 및 결정별 완료 분기와 충돌한다.\n  신뢰도: 높음. 권고: 두 조건 모두 명시적인 승인 판정을 요구하고, 거절·미결정에서는 기존 동작을 유지하도록 맞춰.\n\n- [P2] AC-016의 관측 시점과 채택 규칙이 완료를 막는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:99\n  AC-016은 미래 B0·B1 커밋이 필요하지만, :15는 pre-flight에서 RED를 관측하지 못한 기준의 완료를 차단한다. 계획대로 커밋을 만들면 ancestry 검사는 처음부터 GREEN이며, 존재하지 않는 SHA의 오류는 순서 위반의 RED가 아니다.\n  신뢰도: 높음. 권고: 잘못된 커밋 순서를 가진 fixture로 실패를 관측하고, 실제 B0·B1의 순서 검증은 두 커밋 생성 후 수행하도록 채택 조건을 정리해.\n\n- [P2] D7 부분 미해결: 잘못된 브랜치 연결과 용량도 list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:171\n  경로와 브랜치를 각각 집합으로 비교하고 용량은 숫자인지만 검사한다. 현재 등록된 270개 레코드의 브랜치를 순환 교환하고 모든 `size_kib`를 -1로 만든 mutant가 명시된 JSON 검사 네 가지를 모두 만족했다.\n  신뢰도: 높음. 권고: 경로별 브랜치를 비교하고, 알려진 용량의 fixture에서 JSON·표 양쪽의 실제 크기를 검증해.","findings":[{"severity":"P1","title":"D14 미해결: 출시 차단 기준의 RED 증거가 MP-8을 충족하지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:15","body":"D14 미해결: 출시 차단 기준의 RED 증거가 MP-8을 충족하지 않는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:15\nRB 셀의 종료 코드를 run-phase로 미뤘다. f59ae36be에서 AC-004의 명령은 PASS했고, AC-011·013의 테스트 명령은 `[no tests to run]`으로 종료했다. 확장 fixture의 실패는 관측되지 않았다.\n신뢰도: 높음. 권고: 감사 전에 네 요소가 갖춰진 RED 셀을 기록하고 재실행해. Tier M 임계값은 0.80이며, 점수와 무관하게 MP-8 실패는 통과를 차단한다.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":15,"confidence":0,"recommendation":""},{"severity":"P1","title":"거절 판정도 Q1·Q4 변경의 실행 조건을 만족한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11","body":"거절 판정도 Q1·Q4 변경의 실행 조건을 만족한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:11\nM1은 Q4에 “operator verdict”가 있으면 적용하며, Q1도 같은 조건을 사용한다(:21). 거절도 판정이므로, 승인된 경우만 변경하도록 한 REQ-CAP-004·006 및 결정별 완료 분기와 충돌한다.\n신뢰도: 높음. 권고: 두 조건 모두 명시적인 승인 판정을 요구하고, 거절·미결정에서는 기존 동작을 유지하도록 맞춰.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md","line":11,"confidence":0,"recommendation":""},{"severity":"P2","title":"AC-016의 관측 시점과 채택 규칙이 완료를 막는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:99","body":"AC-016의 관측 시점과 채택 규칙이 완료를 막는다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:99\nAC-016은 미래 B0·B1 커밋이 필요하지만, :15는 pre-flight에서 RED를 관측하지 못한 기준의 완료를 차단한다. 계획대로 커밋을 만들면 ancestry 검사는 처음부터 GREEN이며, 존재하지 않는 SHA의 오류는 순서 위반의 RED가 아니다.\n신뢰도: 높음. 권고: 잘못된 커밋 순서를 가진 fixture로 실패를 관측하고, 실제 B0·B1의 순서 검증은 두 커밋 생성 후 수행하도록 채택 조건을 정리해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":99,"confidence":0,"recommendation":""},{"severity":"P2","title":"D7 부분 미해결: 잘못된 브랜치 연결과 용량도 list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:171","body":"D7 부분 미해결: 잘못된 브랜치 연결과 용량도 list 인수조건을 통과한다 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:171\n경로와 브랜치를 각각 집합으로 비교하고 용량은 숫자인지만 검사한다. 현재 등록된 270개 레코드의 브랜치를 순환 교환하고 모든 `size_kib`를 -1로 만든 mutant가 명시된 JSON 검사 네 가지를 모두 만족했다.\n신뢰도: 높음. 권고: 경로별 브랜치를 비교하고, 알려진 용량의 fixture에서 JSON·표 양쪽의 실제 크기를 검증해.","file":".moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md","line":171,"confidence":0,"recommendation":""}],"next_steps":[]}],"overall_verdict":"fail","disagreement_flag":null,"participant_count":1,"residual_risk_note":"required-backend FAIL: codex","fail_open_backends":[],"audit_receipt":"rcpt-0a7cef923221b27a1ac96ed8","plan_source":"config","build_commit":"4f8aba061","tree_root":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631"}

## Verdict

verdict: FAIL
audited_sha: f59ae36bed13f23961d96e7df640af3bf4c93d9c
overall_score: 0.69
must_pass_failed: 2
blocking_count: 8
plan_artifact_hash: none (mcp__moai__audit_cache is not callable in this session; the hash was not replayed; see Gaps G1)
scope: full
fix_scope: spec.md#REQ-CAP-004; spec.md#REQ-CAP-006; spec.md#REQ-CAP-008; spec.md#REQ-CAP-009; spec.md#REQ-CAP-011; spec.md#§B.2; spec.md#A11; spec.md#A12; spec.md#A14; plan.md#M1; plan.md#M5; acceptance.md#§0; acceptance.md#AC-CAP-004; acceptance.md#AC-CAP-005; acceptance.md#AC-CAP-011; acceptance.md#AC-CAP-012; acceptance.md#AC-CAP-013; acceptance.md#AC-CAP-014; acceptance.md#AC-CAP-016; decision-index.md#Q1; decision-index.md#Q2; decision-index.md#Q4
defect_class: B1=ac-wording; B2=design; B3=ac-wording; B4=ac-wording; B5=traceability; B6=ac-wording; B7=design; B8=design
convergence_overall: fail
required_backend: codex fail
required_backend_fail: codex

Convergence is fail (codex, required, receipt rcpt-0a7cef923221b27a1ac96ed8). The verdict is FAIL on three independent grounds: MP-8 fails (must-pass), MP-9 is UNVERIFIED in this pass (counted as a must-pass failure under M4), and the score 0.69 is below the Tier M threshold 0.80.

## Evidence-bearing report

### Claim

The revision resolves most iteration-1 defects at the plan level (D1, D2, D3, D6, D8, D9, D10, D11, D13, O1, O2). It does not pass: release-blocking RED cells are not four-element (MP-8), and the Q1/Q4 trigger wording admits a rejected verdict as an approval (new blocking). Several criteria still pass on the wrong thing (AC-CAP-012, AC-CAP-014, AC-CAP-016). Two product-level behaviour changes are adopted without an operator decision record.

### Evidence

- E1 `date -u` → `Sat Oct 10 07:54:33 UTC 2026` (start); `Sat Oct 10 08:03:32 UTC 2026` (last read).
- E2 `git rev-parse HEAD` → `f59ae36bed13f23961d96e7df640af3bf4c93d9c`; short `f59ae36be`.
- E3 `moai verify audit-plan --project-root …/.moai/worktrees/t1631` → `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`; backends: claude `off`, codex `required`, glm `off`.
- E4 `mcp__moai__audit_multi` (project_root = worktree, target = SPEC directory, no `gates`) → `overall_verdict: fail`; codex `verdict: fail`; `audit_receipt: rcpt-0a7cef923221b27a1ac96ed8`; `build_commit: 4f8aba061`.
- E5 `mcp__moai__audit_cache` (op `compute_hash`) → `No such tool available: mcp__moai__audit_cache`.
- E6 RB re-execution on the current tree, each single-invocation:
  - AC-CAP-005 `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` → five paths (the RED reproduces; no exit code recorded in the cell).
  - AC-CAP-011 `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` → `243:  if profileBytes <= config.DefaultProfileMaxBytes {` (reproduces; no exit code in the cell).
  - AC-CAP-012 and AC-CAP-013 `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` → no output (reproduces; exit code not recorded).
  - AC-CAP-014 `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` → not run as written (zsh glob error, `no matches found: --include=*.go`); the unquoted form was refused by the worktree guard. The earlier shape returned no output. Not re-executed with an exit code (Gap G3).
  - AC-CAP-007 `grep -n -E 'usage-log' internal/hygiene/sinks.go` → line 91 only (cell matches).
  - AC-CAP-009 `grep -n -E 'mode: (report|apply)' …/workflow.yaml` → `318:        mode: report`.
  - AC-CAP-004 premise: `grep -n -E 'check-ignore|no-index|TempDir' internal/template/gitignore_local_artifacts_test.go` → `repo := t.TempDir()`, `check-ignore -q --no-index`. The isolated probe exists, but no extended case for `usage-log`, `lessons-inbox`, or `corrupt` is in the test (`grep` of the template file found no state-file rule). The cited RED therefore does not reproduce on the current tree; codex also measured the command as PASS at f59ae36be.
- E7 Template gitignore: `grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore` → no output (confirms AC-CAP-004 premise).
- E8 Repository gitignore: `grep -n -E 'usage-log|lessons-inbox' .gitignore` → lines 311, 352, 353, 354. Issue 1 (verdict path ignored) is true for `.moai/reports/*` by convention.
- E9 Flags: `grep -n -E '"force"|"home"|"apply"|"audit-logs"' internal/cli/clean.go` → `--force` 106, `--home` 107, `--audit-logs` 111, `--apply` 113. A18 (`--force` for the home cleanup) is correct.
- E10 `grep -n -E 'usage-log' internal/hygiene/sinks.go` → `91: "usage-log.jsonl": "internal/harness/retention.go (30-day prune to monthly gz)"`. The owner exception is as stated.
- E11 `grep -rn -i owner internal/worktree internal/cli/worktree internal/core/git --include=*.go` → refused by the worktree guard (unquoted glob argument; "This session is isolated in the worktree … Refusing to run it"). A14's claim that no owner field exists is NOT re-observed (Gap G2).
- E12 GiB search → `spec.md:97` (REQ-CAP-014, `10 GiB (10,485,760 KiB)`), `spec.md:142` (A5), `spec.md:148` (A11, states that the 5 GiB cap is lost for users without the key), `spec.md:149` (A12: "The operator confirms it when deciding Q2"), `acceptance.md:24, :50, :55, :146`. `decision-index.md` contains no GiB or 10,485,760 line. The Q2 row does not ask the operator to confirm the threshold, so A12's confirmation has no surface.
- E13 `NEEDS CLARIFICATION` grep over the SPEC directory → no output (MP-7 holds; research.md absent in Tier M).
- E14 `git merge-base --is-ancestor 4f8aba061 f59ae36be` → exit 1 (the build commit is NOT an ancestor of HEAD).
- E15 `git merge-base --is-ancestor 2aab5f797 4f8aba061` → exit 0, no output (the tree's last code base commit IS an ancestor of the build commit).
- E16 Revised texts read in full: spec.md v0.1.2 (REQ-CAP-001 to -014 at lines 75–97; §B.2 at 60–67; A10–A18 at 147–156), plan.md (M0 at 27, M1 at 9 with "Applies only when decision Q4 carries an operator verdict" at :11 and "Applies only when decision Q1 carries an operator verdict" at :21), acceptance.md (§0 at 7–17, §D.0 at 19–38, AC-CAP-001 to -016, §D.6 at 182–193), decision-index.md (Q1 to Q4; Q2 and Q3 now carry Default and Alternate).
- E17 Clock at the end: `Sat Oct 10 08:03:32 UTC 2026`; lease deadline 2026-10-10T08:08:58Z.

### Baseline-attribution

- Tree measured: HEAD `f59ae36bed13f23961d96e7df640af3bf4c93d9c` (short `f59ae36be`), branch `WT-capacity-reclaim`, clean at the start of this audit.
- Code observations (E6, E7, E9, E10) were taken in this run on this worktree.
- Judging build: `moai version` → `v3.2.0-rc.29`, `moai_cp/20261007_195927-158-g4f8aba061`, built `2026-10-09T16:02:30Z`. Build commit `4f8aba061`. Status per verification-claim-integrity §2.2: NOT built from this tree. The build is not a strict ancestor of HEAD (E14). Its ancestor `2aab5f797` is the tree's last code commit before the SPEC-only commits (E15), so the build descends from the code base the SPEC measures and carries later commits. The codex result (E4) and the audit-plan config read (E3) are attributed to build `4f8aba061` beside tree HEAD `f59ae36be`. The file-based findings rest on reads in this worktree, not on the build's output, so the build lag does not change them.

### Gaps

- G1 `plan_artifact_hash` NOT obtained. `mcp__moai__audit_cache` (op `compute_hash`) is not available in this session (E5). The hash was not replayed by hand, per the instruction. The machine line carries `none`, so the admission rule refuses this verdict on the hash field until a hash is recorded from the tool.
- G2 A14's claim (no owner field in `internal/worktree`, `internal/cli/worktree`, `internal/core/git`) was NOT re-observed: the grep was refused by the worktree guard (E11). Recorded as a refused-tool degradation under verification-claim-integrity §3.1. D13 is marked resolved only at the plan-text level.
- G3 AC-CAP-014 RED was not re-executed with a fixed quoting form; the zsh glob refused the unquoted `--include=*.go` (E6). It is recorded as not re-executed.
- G4 The awk-based verification verbs (Group A traceability, Group B CN-4 ordering, Group D D8) were NOT run in this pass. Refused or omitted under the worktree guard's stated rule against awk programs. The CN-4 ordering judgment for MP-9 was therefore NOT made in this iteration. MP-9 is counted as UNVERIFIED (M4), not as PASS.
- G5 The AC-CAP-002 pipeline (`git worktree list … | xargs -0 du -skc`) was read but not executed (it is a 10-minute-class traversal and outside the deadline). D9 is marked resolved at the plan-text level only.
- G6 Codex is the only backend (`claude` and `glm` gates are `off`). No second opinion on the rubric scores.
- G7 The `codex` auxiliary report, if any, was not read in full.

### Residual-risk

- The verdict would not change on MP-9 if the ordering judgment were made: the codex audit did not raise an ordering conflict, but the mechanical check was not run here (G4).
- The overall score is rubric-anchored by hand (no second backend). A different reader could move a band between 0.75 and 0.50 on Testability or Clarity; the verdict still falls below 0.80 unless Testability and Clarity both rise.
- The Q1 to Q4 decisions remain OPEN by design. Their presence is not a plan defect (the codex audit agrees). The trigger wording (B2) is a defect.

## Must-pass results

| Criterion | Result | Evidence and layer |
|---|---|---|
| MP-1 REQ number consistency | PASS | REQ-CAP-001 to REQ-CAP-014, sequential, spec.md:75–97. Requirement layer. |
| MP-2 GEARS format compliance | PASS (with notes) | REQ-CAP-008 and -009 now begin "When the operator runs …" (spec.md:88–89); REQ-CAP-010 to -012 begin "When" or "Where"; REQ-CAP-005 Ubiquitous. Residual notes (optional O-new-2): REQ-CAP-004 carries a second, non-"shall" sentence; REQ-CAP-009 carries "No cap value shall be compiled" (legacy negative). Requirement layer only. |
| MP-3 YAML frontmatter validity | PASS | spec.md:2–15 has all 12 canonical fields; `version: "0.1.2"` quoted; `created`/`updated` as dates; `tags` a string; no snake_case alias. |
| MP-4 Language neutrality | PASS (N/A-equivalent) | Go-scoped SPEC; no language-specific tool enumeration. |
| MP-5 D7 cross-SPEC reconciliation | PASS | References are to completed SPECs (SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001, SPEC-WORKTREE-REAPER-001) per iteration-1 E13; no BLOCKING line in this pass. D7 verb not re-run (G4). |
| MP-6 D8 cross-platform discipline | PASS | `syscall` absent in SPEC files (iteration-1 E8; not re-run, G4). |
| MP-7 clarification gate | PASS | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CAPACITY-RECLAIM-001` → no output (E13). |
| MP-8 RED-now cell re-execution | FAIL | RB criteria (AC-CAP-004, -005, -011, -012, -013, -014) lack the four-element cell: no recorded exit code and paraphrased, not verbatim, stdout (acceptance.md §0 states "no RB criterion is adopted at this revision"). AC-CAP-004's RED does not reproduce: the extended cases are absent from the test (E6), and codex measured its command as PASS at f59ae36be. Per verification-completeness §2.1, an RB cell with three elements or fewer is unadopted, and an unreproducible RED loses RB eligibility. |
| MP-9 cross-artifact ordering | UNVERIFIED | The awk CN-4 verb was not run (G4). No mechanical or hand ordering scan was completed in this pass. Counted as a must-pass failure under M4. Iteration-1 judged PASS by hand at 40cc1031c; the plan changed (M0 added, ordering clauses changed), so that judgment does not carry over. |

## Category scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity in one or two requirements | "carries an operator verdict" (plan.md:11, :21) admits a rejection (B2); the AC-CAP-014 branch comparison has no normalization rule (`refs/heads/` vs short name). |
| Completeness | 0.75 | all sections present; frontmatter complete; Out of Scope has H3 headings with bullets (spec.md:101–134); decision outcome branches now in §D.6 | Q2 row does not carry the 10 GiB threshold that A12 says the operator confirms (E12); two product behaviour changes (REQ-CAP-011 narrowing, REQ-CAP-009 cap removal) have no decision row. |
| Testability | 0.50 | several criteria pass on the wrong thing | AC-CAP-012 checks only that keys exist (acceptance.md:156), so placeholder values pass; AC-CAP-014 passes a mutant that swaps branches and sets every `size_kib` to -1 (codex, acceptance.md:171); AC-CAP-004 RED premise not observed (E6); AC-CAP-016 ancestry check is GREEN from creation (codex, acceptance.md:99). |
| Traceability | 0.75 | mapping complete (14 of 14 REQs have an AC; 16 of 16 ACs map to a REQ); depth gaps closed for archives (AC-CAP-006), `--yes` (AC-CAP-013), fields (AC-CAP-014); the 10 GiB source is not in the decision that confirms it (E12) | 14 of 14 REQs covered (E16). |

Overall score = mean of the four dimensions = (0.75 + 0.75 + 0.50 + 0.75) / 4 = 0.6875, reported as 0.69. Threshold for Tier M: 0.80. Not met.

## Iteration-1 defect status (D1 to D14, O1 to O5)

| ID | Status at f59ae36be | Basis |
|---|---|---|
| D1 usage-log rotation mechanism | RESOLVED at plan level | M1 declares `.moai/harness` as the sink directory, keeps the pass lock under `LogDir`, serializes with the prune-state lock, and does not retire the prune (plan.md:13–18). Code paths cited at rotate.go:26–27, 102 (iteration-1 measurements; not re-read in this pass). |
| D2 `--yes` disposal guard | RESOLVED | AC-CAP-013 runs report and `--yes` over U, W, L, C fixtures (acceptance.md:159–165). |
| D3 AC-CAP-007 passes without rotation | RESOLVED | AC-CAP-007 (a) to (e) require displacement with hash and size, empty primary, no `.2`, audit row, and lock contention (acceptance.md:105–112). |
| D4 DoD vs decision outcomes | PARTIAL | §D.6 maps outcome branches (acceptance.md:187–191). The Q1/Q4 trigger still admits a rejection (new B2). |
| D5 keep-3 rule, cap key, commands | RESOLVED | REQ-CAP-008 matches the code (current plus three non-current, A10); the cap key is named (REQ-CAP-009, A11); the apply command is `moai clean --home --force` (E9); AC-CAP-010 and -011 name `go test` targets. |
| D6 ignore criterion on comments | RESOLVED | AC-CAP-004 uses the isolated `git check-ignore -q --no-index` in `t.TempDir()` (E6, confirmed in `gitignore_local_artifacts_test.go:46, :85`). The RED premise is still unobserved (B1). |
| D7 AC-CAP-014 count-only | PARTIAL | Path and branch sets and numeric size are now checked (acceptance.md:171), but codex's branch-swap and `-1` size mutant passes (B4). |
| D8 archives not verified | RESOLVED | AC-CAP-006 hashes the three `.corrupt` files and checks presence (acceptance.md:86–92). |
| D9 AC-CAP-002 masks errors and truncates paths | RESOLVED at plan level | The command uses `cut -c10-`, `tr '\n' '\0'`, `xargs -0 du -skc`, with exit 123 on `du` failure (acceptance.md:54–56). Not executed (G5). |
| D10 size target no REQ states | RESOLVED with a residual | REQ-CAP-014 states the 10 GiB target (spec.md:97). Its confirmation surface is missing from Q2 (B5). |
| D11 MP-2 "On apply," | RESOLVED | REQ-CAP-008 and -009 begin with "When the operator runs `moai clean --home --force`". |
| D12 baseline-first ordering | PARTIAL | B0 precedes B1 as a separate commit (plan.md:27–32); AC-CAP-016 checks it. Its RED is not observable before B0 exists (B6). |
| D13 owner / confirmed-owner / live-session predicates | RESOLVED at plan level | §B.2 defines the four predicates with code references (spec.md:60–67). The A14 grep was refused (G2). |
| D14 release-blocking classification vs DoD | UNRESOLVED | Classes RB, RG, MS are now explicit (acceptance.md §0), but the RB RED cells are not four-element (B1). |
| O1 Unwanted form | RESOLVED | REQ-CAP-005 and -011 are now Ubiquitous and Event-driven. |
| O2 Q2 and Q3 Default/Alternate | RESOLVED | decision-index.md Q2 and Q3 carry Default and Alternate lines. |
| O3 status duplicated outside spec.md | OPEN (by design) | progress.md:3 keeps `status: draft` (plan.md §D.6 rationale: sync closer rewrite). Optional. |
| O4 Q3 measurement | RESOLVED | Q3 row gives the per-clone recipe. |
| O5 pin wording | RESOLVED (wording) | acceptance.md:5 states the plan commit changes only SPEC artifacts. |

## New issues (seven, as listed in the revision)

| # | Issue | Judgment | Class |
|---|---|---|---|
| 1 | `.gitignore` ignores `.moai/reports/*` (line 235) and usage-log (line 352) | Not a plan defect. The verdict path is local by convention, and B0 avoids `.moai/reports` for that reason (plan.md:30). Line 352 verified (E8). | not a defect |
| 2 | AC-CAP-006 relaxed from "decrease is a failure" | Mostly sound. Residual wording conflict: "truncation to zero is the failure" (acceptance.md:91) while the 30-day prune can legitimately rewrite to zero. | optional O-new-1 |
| 3 | Q4 alternate discards the `.1` chunk | Disclosed in the Q4 row, A17, and §F. Decision-gated, not a plan defect. The consequence is operator-visible. | optional (residual risk) |
| 4 | Home cleanup switch is `--force` | Verified at clean.go:106 (E9). A18 is correct. | not a defect |
| 5 | Profile cap has no shipped default; loader precedent falls back to a compiled default | The REQ forbids a compiled cap. The removal of the 5 GiB default for every user without the key is an unrecorded product decision (A11, §F). | BLOCKING B8 |
| 6 | `--yes` disposes unowned landed trees; the owner predicate narrows it | REQ-CAP-011 changes existing `sweep --yes` behaviour for every user with no decision row. Q2 gates only the bulk run, not this narrowing. | BLOCKING B7 |
| 7 | Size proxy omits `.codex/worktrees/` | §B labels the two-root sum as a proxy, and A16 defines the criterion set. Not a defect. | not a defect |

Q1 to Q4 are open by design. Their presence is not a plan defect. The defects are the trigger wording (B2), the missing threshold surface (B5), and the missing outcome record for B7 and B8.

## Blocking defects (blocking_count = 8)

- B1 (MP-8, must-pass): RB RED cells in acceptance.md §0 and §D.0 (AC-CAP-004, -005, -011, -012, -013, -014) lack a recorded exit code and verbatim stdout; AC-CAP-004's RED does not reproduce on the current tree (E6, codex). Fix: record four-element cells (command, verbatim stdout, exit code, tree SHA) and re-execute at the pinned tree, or demote the criterion to RG per verification-completeness §2.1.
- B2 (plan.md:11, :21; spec.md REQ-CAP-004, -006): "Applies only when decision Q4 / Q1 carries an operator verdict" also triggers on a rejection. Fix: require an explicit approved verdict and keep the default branch for rejection or no verdict.
- B3 (AC-CAP-012, acceptance.md:156): checks key presence only; placeholder values pass. Fix: compare `owner`, `owner_state`, `unpushed` with a fixture of known values (a done card, a run card, an unpushed branch).
- B4 (AC-CAP-014, acceptance.md:171): the branch-swap and `size_kib = -1` mutant passes (codex). Fix: compare branch per path and check size values against `du -sk` on a fixture.
- B5 (Q2 row, decision-index.md:15–22; A12, spec.md:149): the 10 GiB threshold has no confirmation surface. Fix: add the threshold and its source to the Q2 row.
- B6 (AC-CAP-016, acceptance.md:94–99): the ancestry check is GREEN from creation and its RED is unobservable (codex P2). Fix: observe a RED on a wrong-order fixture before adoption, or state the adoption rule so it is not a close blocker.
- B7 (REQ-CAP-011, spec.md:94; A14): the owner predicate narrows the existing `sweep --yes` for every user (manual trees become non-disposable through sweep). This is an unrecorded product decision. Fix: add a decision row, or scope the narrowing explicitly.
- B8 (REQ-CAP-009, spec.md:89; A11; plan.md M5): removal of the compiled profile cap leaves most users with no profile cap until they set the key. This is an unrecorded product decision. Fix: add a decision row for the shipped default, or record the removal as an operator-approved change.

## Optional findings (not counted in blocking_count)

- O-new-1 AC-CAP-006 "truncation to zero is the failure" vs the prune's legitimate rewrite to zero (acceptance.md:91). Tighten the wording.
- O-new-2 REQ-CAP-004 second sentence and REQ-CAP-009 "No cap value shall be compiled" are non-canonical GEARS forms (legacy negative). Consider splitting.
- O3 (carried) progress.md:3 `status: draft` duplicates lifecycle state outside spec.md.

## Operational Notes (unverified)

- Re-run the AC-CAP-002 pipeline on a path containing a space. Label: inferred (shell field handling); not executed here (G5).
- Record the plan_artifact_hash from the moai audit tool once it is reachable. Label: measured (tool absent, E5); the hash must come from the tool, not a replay.

## Commands run (with observed output)

- `date -u` → `Sat Oct 10 07:54:33 UTC 2026`
- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631`
- `git rev-parse HEAD` → `f59ae36bed13f23961d96e7df640af3bf4c93d9c`
- `moai version` → `v3.2.0-rc.29   moai_cp/20261007_195927-158-g4f8aba061   built 2026-10-09T16:02:30Z`
- `moai verify audit-plan --project-root …/.moai/worktrees/t1631` → `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`
- `git merge-base --is-ancestor 4f8aba061 f59ae36be` → exit 1
- `git merge-base --is-ancestor 2aab5f797 4f8aba061` → exit 0 (no output)
- `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` → five paths
- `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` → `243: if profileBytes <= config.DefaultProfileMaxBytes {`
- `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` → no output
- `grep -rn 'Use:.*"list"' internal/cli/worktree --include=*.go` → refused by zsh glob (not re-run, G3)
- `grep -n -E 'usage-log' internal/hygiene/sinks.go` → `91: "usage-log.jsonl": "internal/harness/retention.go (30-day prune to monthly gz)",`
- `grep -n -E 'mode: (report|apply)' …/workflow.yaml` → `318:        mode: report`
- `grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore` → no output
- `grep -n -E 'check-ignore|no-index|TempDir' internal/template/gitignore_local_artifacts_test.go` → `46: repo := t.TempDir()`, `85: check-ignore -q --no-index`, `95: t.Fatalf`
- `grep -n -E 'usage-log|lessons-inbox' .gitignore` → `311`, `352`, `353`, `354`
- `grep -n -E '"force"|"home"|"apply"|"audit-logs"' internal/cli/clean.go` → `106 --force`, `107 --home`, `111 --audit-logs`, `113 --apply`
- `grep -rn -i owner internal/worktree internal/cli/worktree internal/core/git --include=*.go` → refused by the worktree guard (G2)
- `grep -n -E 'GiB|10,485,760' …/spec.md …/plan.md …/acceptance.md …/decision-index.md` → spec.md:39, :97, :142, :148, :149; acceptance.md:24, :50, :55, :146
- `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CAPACITY-RECLAIM-001` → no output
- `mcp__moai__audit_cache` (op `compute_hash`) → `No such tool available`
- `date -u` (last) → `Sat Oct 10 08:03:32 UTC 2026`

## Recommendation

FAIL. Revise the SPEC (manager-spec, via the orchestrator) to address B1 to B8 in the fix_scope above. Then re-audit: this is the last Tier M audit spawn, so the next step is the tier ceiling policy (harness.plan_audit_ceiling_policy), not a further spawn in this card. The operator must record verdicts for Q1 to Q4 (with approve and reject distinguished) and confirm the 10 GiB threshold (B5), the shipped profile-cap default (B8), and the narrowing of `sweep --yes` (B7).

## Machine-line placement note

The machine lines appear once each, in the Verdict section above. The `plan_artifact_hash` line is `none` because the audit tool was not callable (G1); the verdict is not admissible on that field until the tool hash is recorded.
