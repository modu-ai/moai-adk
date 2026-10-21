auditor-model: GPT-6 (Codex)
Verdict: fail

검토 SHA: `61b5af4b14505926b3435bc9daf6d3e38eaafecc`  
부모 SHA: `8249bff8c3933d647b4f99e59048595f9808e83a`  
보고서 반환 대상: `.moai/reports/t1916/plan-audit.md`  
이번 검토는 읽기 전용이며, 아래 구현 근거는 해당 SHA의 정적 증거다.

## Defects Found

- [P1] deadline 만료를 host load의 증거로 취급하면 원인 오분류가 그대로 남는다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/spec.md:94
  - **D1 · 확신 높음 · major · blocking · defect_class: design**
  - 근거: AC-FLB-001(c)는 **write lock을 유지하면서 deadline을 단축**하고, AC-FLB-002는 그 결과에 host-load 설명을 요구한다. 실제 `retryFactoryBusy`는 잠금 경합 중 context가 만료되면 `errors.Join(ctx.Err(), err)`를 반환한다(`internal/homestate/factory.go:642–655`). 따라서 DB 잠금 때문에 만료된 요청도 host load로 설명하면서 AC를 통과한다. 현재 구현의 주석은 부하 측정 증거가 아니다.
  - 권고: 관측한 deadline 만료와 추정 원인을 구분하고, busy와 deadline이 함께 있는 경우의 우선순위를 명시하라. (b)의 `flHoldRecord`도 deadline까지 유지될 수 있으므로, (b)/(c)를 구별하는 입력·context 상태를 각각 정의하라.

- [P1] serial-slot 점유와 SQLite busy 오류를 혼동해 M3의 구현 대상이 성립하지 않는다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/plan.md:103
  - **D2 · 확신 높음 · major · blocking · defect_class: design**
  - 근거: 인용한 `factory_card.go:1857–1861`은 슬롯 점유에 관한 주석이다. 실제 `factorySerialInFlightExcluding`은 이미 읽은 카드 목록에서 **bool**을 계산한다(`factory_card.go:330`). nominated 경로는 이 값이 참이면 정상적으로 `serial-slot`을 반환한다(`factory_card.go:1367`). DB 읽기 오류는 별도 `ListCards` 경로다(`factory_card.go:1402`).
  - 권고: REQ-FLB-009와 AC-FLB-007의 `busy`를 슬롯 점유 또는 실제 DB 읽기 실패 중 하나로 확정하라. 슬롯 점유라면 REQ-FLB-010의 보존 의무와 충돌하고, DB 실패라면 실제 오류 발생 지점과 검증 seam을 지정해야 한다.

- [P1] compensation AC가 기존 동작 보존 대신 상태 변경을 요구한다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/acceptance.md:37
  - **D3 · 확신 높음 · major · blocking · defect_class: ac-wording**
  - 근거: AC-FLB-006은 promotion 뒤 세 claim 실패 모두에서 `queued` 복구를 요구한다. 그러나 직접 인용한 기존 테스트 `factory_lease_bound_test.go:302–307`은 중간 실패 뒤 **record=`assigned`, queue=`picked` 유지**를 검증한다. 구현도 이 lane의 assigned row를 보존한다(`factory_card.go:1658–1662`). 이를 `queued`로 바꾸면 REQ-FLB-008의 “same compensation … no store state”를 위반한다.
  - 권고: 실패 시점을 분리해 record 없음/unowned picked, 자기 lane assigned, 다른 holder 소유 각각의 기대 상태를 명시하라. 기존 중간 실패 복구·재시도 동작을 보존하는 AC로 고쳐야 한다.

- [P1] `raced` 독점 요구와 queue-lock 예외를 동시에 만족할 수 없다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/spec.md:98
  - **D4 · 확신 높음 · major · blocking · defect_class: design**
  - 근거: REQ-FLB-005는 다른 claimant에게 진 실패가 아니면 `raced`를 금지한다. 반면 같은 문서 134–137행은 queue-lock timeout의 `raced`를 보존하고, AC-FLB-010도 해당 테스트를 변경 없이 통과시키도록 요구한다. 실제 queue-lock 거절은 승자를 확인하지 않고 `raced`를 반환한다(`factory_card.go:1525–1528`).
  - 권고: queue-lock 예외를 REQ-FLB-005의 적용 범위에 직접 반영하라. AC-FLB-002도 deadline 사례 하나가 아닌, 합의된 독점 범위 전체를 검증하도록 맞춰라.

- [P1] status 측정 절차에 실행 대상과 격리가 없어 공유 DB를 변경할 수 있다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/plan.md:106
  - **D5 · 확신 높음 · major · blocking · defect_class: design**
  - 근거: M4와 AC-FLB-008/009에는 표본 수·load·percentile만 있고 실행 바이너리, fixture, root 설정, 측정 명령이 없다. worktree의 기본 `factoryCardRoot()`는 primary checkout을 가리킨다(`factory_card.go:90–98`). status는 `OpenFactory`를 호출하고, 그 경로는 디렉터리 생성·chmod·DDL·마이그레이션을 수행한다(`internal/homestate/factory.go:171–185,292–350`). “read-only”라는 설명만으로 수동 반복 실행의 무변경성이 보장되지 않는다.
  - 권고: 별도 임시 프로젝트와 격리된 homestate, 고정 데이터, revision별 바이너리를 사용하는 실행 가능한 recipe를 계획에 넣어라. 출력·종료 코드도 검증하고, 같은 입력으로 전후를 측정하도록 고정하라. 기본 live 프로젝트에서의 반복 측정을 승인된 검증으로 취급하면 안 된다.

- [P2] 지연 AC의 배타 조건이 정상 개선을 거절하고 큰 회귀를 허용한다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/acceptance.md:50
  - **D6 · 확신 높음 · major · blocking · defect_class: ac-wording**
  - 근거: 이번 실행에서 문서의 조건식을 계산했다. 기준선 **100ms → 40ms**는 REQ-FLB-012의 OR 조건을 만족하지만 두 AC 조건이 모두 참이라 `exactly one`에는 실패한다. **100ms → 1,000ms**는 두 번째 조건만 참이라 AC를 통과한다.
  - 권고: 기준선에 따라 먼저 경로를 선택하라. `<200ms`이면 변경 없이 재현 실패로 종료하거나, 변경한다면 허용 회귀 조건을 별도로 정의해야 한다. 나머지는 50% 개선을 요구하도록 REQ와 AC를 일치시켜라.

- [P2] REQ–AC 연결표가 실제 검증 누락을 가린다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/acceptance.md:28
  - **D7 · 확신 높음 · major · blocking · defect_class: traceability**
  - 근거: REQ-FLB-007은 `backend-skip`까지 포함해 기존 pre-claim 거절 **문구와 exit code**를 보존한다. AC-FLB-004에는 `backend-skip`이 없고 나머지도 exit만 검사한다. 따라서 quota-hold/foreign-worktree 문구를 바꿔도 해당 AC는 통과할 수 있다. REQ-FLB-003의 record-lock detail 역시 AC-FLB-001의 token 검사로는 보장되지 않는다.
  - 권고: 해당 REQ가 요구하는 원인별 detail과 기존 pre-claim 문구·exit를 직접 검증하라. 연결표의 ID 존재와 실제 assertion coverage를 구분해야 한다.

- [P2] 두 REQ가 지정된 GEARS 패턴을 벗어난다 — .moai/specs/SPEC-FACTORY-LEASE-BUSY-001/spec.md:100
  - **D8 · 확신 높음 · minor · blocking · defect_class: format**
  - 근거: REQ-FLB-006/007은 `Every lease-path refusal … shall`로 시작한다. 적용된 MP-2 rubric의 ubiquitous 형식은 `The <subject> shall …`이다. 전체 12개 requirement entry를 검사해 이 두 항목을 확인했다. Given–When–Then인 **AC 계층에는 이 판단을 적용하지 않았다**.
  - 권고: 두 항목을 `The lease path shall …` 형태로 정규화하라. MP-2는 모든 REQ의 준수를 요구한다.

## Must-Pass Results

- **MP-1 PASS:** REQ-FLB-001–012, 연속·중복 없음(`spec.md:83–128`).
- **MP-2 FAIL:** requirement 계층의 REQ-FLB-006/007 형식 위반. D8 참조.
- **MP-3 PASS:** canonical 12개 필드와 값 형식 확인(`spec.md:2–13`).
- **MP-4 N/A:** 이 저장소의 Go CLI 구현 범위이며 universal template 변경이 아니다(`spec.md:11`).
- **MP-5 PASS:** 참조 SPEC의 실제 상태는 `in-progress`, `completed`, `completed`; retired/superseded/archived 대상 없음.
- **MP-6 PASS:** 전체 spec 본문에서 `syscall` 언급 0개.
- **MP-7 PASS:** `plan.md` 직접 확인 결과 clarification marker 0개. `research.md`는 없으며 Tier M의 필수 입력이 아니다.
- **MP-8 N/A:** acceptance에 release-blocking 분류가 없다. 17–19행의 RED 설명은 실행 명령·stdout·exit code를 갖춘 재현 기록으로 인정하지 않았다.
- **MP-9 PASS, 수동 확인:** 검증 verb는 목록형 milestone을 수집하지 못해 `GAP`을 출력했다. 문서를 직접 읽어 `M0 → M1 → … → M6`을 확인했다. M0의 PR landing·local main 포함·leader resume 조건(`plan.md:90–92`), M4 baseline commit → M5 변경 순서(`plan.md:106–113`)는 acceptance의 선행 조건과 양립한다. **보류가 실제 해제됐다는 판단은 아니다.**

평가 점수: **0.6875 / Tier M 기준 0.80**. Clarity 0.50(D1–D4), Completeness 1.00(필수 artifact·요구·계획·검증·제외 범위 존재), Testability 0.50(D5–D6), Traceability 0.75(D7). MP-2 실패는 점수와 독립적으로 FAIL이다.

## Claim / Evidence

이번 실행에서 확인한 주요 명령과 출력:

```text
git log -1 --format='%H %P' 61b5af4b
61b5af4b14505926b3435bc9daf6d3e38eaafecc 8249bff8c3933d647b4f99e59048595f9808e83a

traceability verification
COLLECTED: 12 REQ definitions (acceptance input: read)
[UNCOVERED/ORPHAN 출력 없음]

ordering verification
COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 5 ordering candidates
GAP: 0 milestone headings collected from the plan — milestone order not observed

추가 구조 검사
Manual milestone order: M0 -> M1 -> M2 -> M3 -> M4 -> M5 -> M6
AC count: 10
CLARIFICATION plan.md markers=0
D8 syscall mentions: 0

조건식 반례 계산
baseline_ms=100 post_ms=40 REQ012_OR=True AC009_exactly_one=False
baseline_ms=100 post_ms=1000 REQ012_OR=True AC009_exactly_one=True

작업 트리 확인
reviewed_paths_diff_exit=0
git status --porcelain=v1
[출력 없음]
```

`moai spec lint SPEC-FACTORY-LEASE-BUSY-001 --json`은 `OwnershipTransitionUnmeasured` info와 `VacuousTestAssertion` warning을 출력했다. 후자는 문서가 의도와 count 검증을 명시하므로 별도 결함으로 올리지 않았다. linter의 침묵을 의미상 traceability나 GEARS 준수의 증거로 사용하지 않았다.

## Baseline-attribution

검토 범위 6개 파일과 관련 구현을 지정 SHA에서 확인했다. 부모 대비 변경은 문서 6개·522행 추가이며 CLI/homestate 구현 변경은 없다. plan hash는 현재 구현의 파일명·길이·원본 바이트 구분 규칙으로 계산했다.

## Gaps

- Go 테스트, 빌드, 성능 측정, live factory 명령은 실행하지 않았다. 미래 RED/GREEN 결과를 이번 증거로 간주하지 않았다.
- PR #1859의 현재 원격 상태와 leader resume은 확인하지 않았다. survey의 과거 기록과 현재 landing 상태는 구분했다.
- Q1의 operator verdict는 비어 있고, Q2의 측정도 아직 없다(`decision-index.md:23,33`). kickoff 전에 필요한 처분과 M4에서 얻을 증거를 구분해 기록해야 한다.
- 재귀 감사나 다른 auditor 호출은 하지 않았다. receipt는 발급받지 않았다.
- 파일은 작성하지 않았다. 이 응답이 launcher의 보고서 저장용 반환 본문이다.

## Residual-risk

명목상 연결은 12개 REQ 전체에 있지만, 원인 분류·보상 상태·지연 검증의 의미상 결함 때문에 현재 계획으로 구현을 시작하면 AC 통과와 요구 충족이 어긋날 수 있다. D1–D8을 수정한 뒤 해당 변경과 전체 ordering 검사를 재검토해야 한다.

```text
verdict: FAIL
audited_sha: 61b5af4b14505926b3435bc9daf6d3e38eaafecc
overall_score: 0.6875
must_pass_failed: 1
blocking_count: 8
plan_artifact_hash: ad465517a5d2a605ba974bdbd8d6411de0a44a33d0f2c876ba6c6afb99a89c97
scope: full
fix_scope: spec.md#REQ-FLB-003, spec.md#REQ-FLB-004, spec.md#REQ-FLB-005, spec.md#REQ-FLB-006, spec.md#REQ-FLB-007, spec.md#REQ-FLB-008, spec.md#REQ-FLB-009, spec.md#REQ-FLB-011, spec.md#REQ-FLB-012, acceptance.md#AC-FLB-001, acceptance.md#AC-FLB-002, acceptance.md#AC-FLB-004, acceptance.md#AC-FLB-006, acceptance.md#AC-FLB-007, acceptance.md#AC-FLB-008, acceptance.md#AC-FLB-009, plan.md#§F
reread_hunks: []
convergence_overall: fail
required_backend: codex fail
required_backend_fail: codex
```

AUDIT-VERDICT: FAIL spec=SPEC-FACTORY-LEASE-BUSY-001 receipts=none