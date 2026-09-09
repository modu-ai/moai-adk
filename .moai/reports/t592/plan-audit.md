# SPEC Review Report: SPEC-HOME-STATE-ROLLOUT-001

Iteration: 3/3  
Verdict: FAIL  
Overall Score: 0.81  
Tier: L (PASS threshold 0.85)  
Escalation: MAX_ITERATIONS_REACHED

Reasoning context ignored per M1 Context Isolation. 이번 재감사는 v0.3.0의 D1/D2/D5 수정
델타와 D3/D4/D6 회귀만 판정했다.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `spec.md:47-83`의 25개 REQ가 001부터 025까지 연속하며 중복이 없다.
- [PASS] MP-2 GEARS format compliance: 판정은 REQ 층에만 적용했다. `spec.md:47-83`의 REQ는 GEARS 또는 허용 중인 legacy EARS 형식이다. `acceptance.md:15-39`의 Given–When–Then은 검증 층이다.
- [PASS] MP-3 YAML frontmatter validity: canonical 12 fields와 quoted semver `"0.3.0"`이 있다(`spec.md:2-13`). Strict lint는 exit 0, stdout `[]`였다.
- [N/A] MP-4 language neutrality: Go 단일언어 구현 SPEC이다(`spec.md:95`).
- [PASS] MP-5 D7 reconciliation: 참조 선행 SPEC 5개가 모두 존재하며 `completed`이다.
- [PASS] MP-6 D8 discipline: `spec.md`에 `syscall` literal이 없다.
- [PASS] MP-7 clarification gate: `plan.md`와 `research.md`에 `[NEEDS CLARIFICATION]` marker가 없다.
- [N/A] MP-8 RED-now: 25 AC 모두 `regression-guard / future green-path`, release-blocking은 0개다(`acceptance.md:43-47`). 현재 PASS를 주장하지 않고, production 변경 전 named test RED와 같은 exact command의 test-count ≥1 GREEN을 요구한다(`acceptance.md:5-10`). 기존 AC-009 selector는 실제 1건 실행됐다. 미구현 24개 test는 정직하게 표시된 TDD green-path이므로 그 부재 자체는 결함이 아니다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|------:|-------------|----------|
| Clarity | 0.50 | 0.50 | D2의 `claimed`/`recovery-required` CAS 충돌과 D5의 pre/post gate 순서 충돌이 남았다. |
| Completeness | 1.00 | 1.00 | Tier L 5개 artifact, 25 REQ/AC, HISTORY와 구체적 Out of Scope가 있다. |
| Testability | 0.75 | 0.75 | 25개 고유 selector와 binary AC가 있으나 D5 성공 경로가 순환한다. |
| Traceability | 1.00 | 1.00 | 25개 machine-readable 1:1 mapping이 있다(`acceptance.md:117-143`). |

Overall Score = `(0.50 + 1.00 + 0.75 + 1.00) / 4 = 0.8125`, rounded 0.81.
Tier L threshold 미달이며 blocking finding 2건이 남았다.

## Defects Found

### D2 — legacy recovery 상태와 CAS가 실행 가능한 단일 상태기계를 이루지 못함

- Severity: critical
- Artifact: `spec.md:65,83`; `design.md:100-123`; `internal/homestate/factory.go:66-68`
- Confidence: High
- Blocking: Yes
- Class: blocking
- Impact: `design.md:112-120`은 NULL/invalid row를 `recovery-required`로 전이하고 그 status를 CAS 조건으로 삼지만, normative `REQ-HSR-025`(`spec.md:83`)는 current `claimed` status/token을 검증하라고 한다. 같은 migrated row에 두 조건은 동시에 참일 수 없다. 또한 현재 v1 DDL의 CHECK는 `pending,claimed,consumed,failed,expired,cleared`만 허용하지만 설계는 table rebuild 없이 `additive columns`만 명시한다. 따라서 `recovery-required` update는 CHECK를 위반해 v1→v2 승격 자체가 실패할 수 있다.
- Required fix: transaction table rebuild로 v2 CHECK에 `recovery-required`를 추가하고 모든 CAS를 그 상태로 통일하거나, 기존 `claimed`를 유지한 additive recovery-state column으로 모든 artifact를 통일한다. 실제 v1 DDL fixture에서 schema upgrade와 dead-owner requeue/fail, live/indeterminate refusal, old-token finish 거부를 검증한다.

### D5 — verified-live gate 순서가 mutation 이전 거부를 보장하지 못함

- Severity: critical
- Artifact: `acceptance.md:35-36,91-94`; `design.md:203-220`
- Confidence: High
- Blocking: Yes
- Class: blocking
- Impact: AC-HSR-021은 nonce tamper/replay를 target/backup/marker mutation 전에 거부하라고 한다(`acceptance.md:35`). 그러나 설계는 step 4에서 backup을 만든 뒤 step 5에서 nonce CAS를 소비한다(`design.md:209-212`). Replay 판정 전에 backup inventory가 이미 바뀐다. 또한 pre-apply step 2는 ledger named validators를 GREEN으로 요구하지만, AC-HSR-022 validator는 live gate readback을 요구하고(`acceptance.md:91-94`) 그 readback은 apply 종료 뒤에만 생성된다(`design.md:219-220`). AC-022를 pre-apply 집합에서 제외하는 규칙이 없어 성공 gate가 아직 존재하지 않는 사후 evidence를 선행 조건으로 요구한다.
- Required fix: pre-apply authorization validators와 post-apply verdict validator를 명시적으로 분리한다. Nonce authenticity/replay 판정을 backup/marker/target 생성 앞으로 옮기고, stale/tampered/replayed/zero-test/bare apply가 세 inventory를 모두 byte-identical하게 보존하는 named test를 요구한다.

## Regression Check

- D1 — [RESOLVED]: 25 AC 전부 future green-path regression guard이며 25 selector는 고유하다(`acceptance.md:43-75`). AC-009는 test 1건을 재실행했고 AC-022도 exact selector가 있다(`acceptance.md:77-94`).
- D2 — [UNRESOLVED]: recovery 동사는 추가됐지만 status/CHECK 모순이 남았다.
- D3 — [RESOLVED]: `continue:false`, `stopReason`, host prompt/tool 0건이 유지된다(`spec.md:59`, `acceptance.md:24`).
- D4 — [RESOLVED]: exact recover/rollback, owner refusal, restore, marker-last와 no-op이 유지된다(`spec.md:81-82`, `acceptance.md:37-38`).
- D5 — [UNRESOLVED]: exact `--verified-live` entry는 생겼지만 nonce/backup 순서와 AC-022 순환이 남았다.
- D6 — [RESOLVED]: provisional parent-token CAS와 cleaner 경쟁 보호가 유지된다(`spec.md:73`, `design.md:149-157`, `acceptance.md:32`).

| Iteration | Score | Verdict | Blocking findings |
|-----------|------:|---------|-------------------|
| 1 | 0.75 | FAIL | D1-D6 |
| 2 | 0.81 | FAIL | D1, D2, D5 |
| 3 | 0.81 | FAIL | D2, D5 |

3회 상한에 도달했다. 자동 반복을 종료하고 PASS-with-debt, scope reduction, explicit iteration-4 override 중 user gate 선택이 필요하다.

## Recommendation

1. D2의 status 표현과 실제 SQLite v1 CHECK migration 방식을 통일한다.
2. D5의 pre-apply authorization과 post-apply readback 검증을 분리하고 replay 판정을 모든 명시적 mutation 앞으로 이동한다.
3. Blocking 2건이므로 현재 live rollout을 승인하지 않는다.

## Evidence-Bearing Record

### Claim

v0.3.0은 D1/D3/D4/D6을 닫았으나 D2/D5가 남아 FAIL이다.

### Evidence

1. Tree: `git rev-parse HEAD` → `6ea69661c405c1b6a3ef38e598f5653a63cd7af0`.
2. Strict lint: `go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json` → exit `0`, stdout `[]`.
3. Static counts: `REQ=25`, `AC=25`, `RG=25`, `RB=0`, `MAPPINGS=25`, `SELECTORS=25`, `UNIQUE=25`.
4. AC-009 command `go test ./internal/kanban -run '^TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary$' -count=1 -v` → exit `0`; stdout included `=== RUN`, `--- PASS`, `PASS`.
5. D2 static evidence:

   ```text
   design.md:100: 기존 table에 additive columns를 추가한다.
   design.md:112: ... `recovery-required`로 전이
   design.md:120: `status='recovery-required'` CAS
   factory.go:68: CHECK(status IN ('pending','claimed','consumed','failed','expired','cleared'))
   ```

6. D5 static evidence: `design.md:205-212`는 validators→census→backup→nonce CAS 순서다. `acceptance.md:35`는 replay를 backup mutation 전 거부하라고 한다. `acceptance.md:91-94`의 AC-022는 `design.md:219-220`의 post-apply readback을 요구한다.

### Baseline-attribution

모든 측정은 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t592`, HEAD
`6ea69661c405c1b6a3ef38e598f5653a63cd7af0`의 현재 artifact/source를 대상으로 했다.

### Gaps

- Future test 24개와 실제 live apply는 실행하지 않았다. 문서가 이를 현재 PASS로 주장하지 않으므로 test 부재 자체는 결함으로 판정하지 않았다.
- 실제 v1 DB mutation은 수행하지 않았고 문서 계약과 현재 v1 DDL의 정적 양립성을 검사했다.

### Residual-risk

- Strict lint `[]`는 D2/D5의 상태기계와 순서 모순을 검출하지 않는다.
- D2는 upgrade failure/영구 미전달, D5는 replay 전 backup mutation 또는 실행 불가능한 circular gate로 이어질 수 있다.

## Operational Notes (unverified)

- `measured` — strict lint, selector count와 AC-009 실행 결과는 위 Evidence에 기록했다.
- `inferred` — D2/D5 판정은 현재 DDL과 normative artifact의 상태/순서 불변식 대조 결과다.
