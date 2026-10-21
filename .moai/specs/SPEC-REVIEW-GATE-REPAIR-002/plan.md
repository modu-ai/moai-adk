# SPEC-REVIEW-GATE-REPAIR-002 — 구현 계획

## §A 맥락

- **Tier M 판정.** 변경 표면은 두 서브시스템(검토 결과 캐시·표식 가족, 스위치 해석과 셸 래퍼)이다. 인용하는 코드 파일은 10개 안팎이고, 요구사항 10개와 인수 기준 12개다(Tier M 상한은 각각 16). 산출물은 spec.md · plan.md · acceptance.md 셋과 progress.md 다. design.md 와 research.md 는 두지 않는다 — 설계 규칙의 원문은 SPEC-001 design.md §3·§5·§7 이 이미 가지고 있고, 이 SPEC 은 그 규칙을 요구로 옮길 뿐 새 설계를 하지 않는다. 카드가 정한 "정확히 세 파일"에 progress.md 뼈대가 더해진 것은 era 분류와 §E.2 증거 적재를 위한 상태 파일이며 내용 산출물이 아니다.
- **증거 원천 Gap.** 카드가 인용한 `.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md` 와 `plan-audit-split-iter1.md` 는 main 에 추적되지 않고 t1620 워크트리는 존재하지 않는다(이 세션 관측: `git ls-tree main .moai/reports/t1620/` 출력 없음, 디렉터리 부재). 이 계획은 그 두 경로를 읽을 수 있는 근거로 쓰지 않는다. iter10 B1·B2, iter3 부채, R-1·F2·F5, tree_scope-skip 항목의 내용은 카드 본문 요약, SPEC-001 HISTORY v0.2.6–v0.2.11 행, SPEC-001 decision-index.md 에서 왔다. 코드 앵커는 전부 이 세션이 기준 트리 `8249bff8c3` 에서 다시 읽었다(spec.md §A.2).
- **현재 상태.** SPEC-001 은 v0.2.11 개정(in-progress) 상태이며 이 카드는 그 어떤 파일도 건드리지 않는다. 이 SPEC 은 draft 다. plan-audit 통과 전까지 run 단계로 넘어가지 않는다.
- **이 계획이 하지 않는 것.** 구현 코드 변경은 run 단계의 몫이다. 상속 항목 1(F5), 5(F2 잔여), 6(N1·N3·F6)은 구현 범위 밖이다(spec.md §E). multi 게이트와 진입 래퍼는 건드리지 않는다.

## §B 되돌리기 어려운 결정 — 상속 처분 (구현 전 확정)

바뀌기 쉬운 결정부터 적는다. 각 항목은 spec.md 「상속 항목 처분」 표의 같은 항목과 짝을 이룬다.

1. **F5 — SPEC-GATE-BOTTLENECK-001 문언과 코드의 어긋남.** 결정: 기록 처분 지속(Q6 = (b) 승계, 운영자 판정 d-20261010T054047Z-3fc2). 근거: 코드 쪽은 이미 해결되었다 — 진입 훅은 차단하지 않고(SPEC-001 REQ-RG-001, `internal/cli/codex_review_delay.go` 진입 경로 주석 "It never blocks (REQ-RG-001)"), 남는 어긋남은 그 SPEC 본문과 현재 코드 사이의 것이 아니라 그 SPEC 본문과 SPEC-001 사이의 것이다. 대상 SPEC 개정은 그 SPEC 소유자의 개정 흐름 문제이며 이 카드가 대신할 수 없다. 처분 위치: spec.md 「Out of Scope — SPEC-GATE-BOTTLENECK-001 의 편집」. 잔여 위험: 대상 SPEC 이 in-progress 인 채 문언이 뒤집힌 채로 남는다(SPEC-001 F5 가 이미 기록한 위험의 지속).
2. **tree_scope-skip 명명 — 이름 분리 대신 사유 표시.** 결정: (b) 사유 표시. 근거: 열두 상태 어휘는 SPEC-001 design.md §7 이 정하고 v0.2.11 이 동결했다("어휘 불변"). 이름을 쪼개면 서명 변경이 되고 SPEC-001 개정이 필요하다. 운영자 혼란의 원인은 이름이 아니라 메시지가 사유를 담지 않은 것이므로, 상태 이름 `tree_scope 건너뜀` 을 유지한 채 메시지 본문이 명시 설정 제외와 기본 판정 제외를 구별한다(REQ-RGR-010). 스킵 행렬의 재측정은 M1 기준선에서 한다(§F 위험 8).
3. **R-1 — 차단 횟수 경합.** 결정: 수리 계획을 새로 두지 않고 상류 수리를 인정한다. 근거(이 세션 관측): `codexReviewReserveBlock`(`internal/cli/codex_review_gate.go:380`)이 키별 잠금 안에서 검사와 기록을 한 단계로 하고, 경합 가드 `TestReviewGate_ConcurrentBlockCap`(`internal/cli/codex_review_gate_concurrency_test.go:23`)이 있다 — SPEC-001 v0.2.11 Amendments 의 SF-1 수리(커밋 `4bd5d433a` RED, `a13e02705` GREEN)와 같은 것이다. 이행: AC-RGR-011 회귀 가드 재실행(M7). 수리가 이미 있다는 판정을 이 세션이 코드로 확인했다.
4. **래퍼 처리기 실패 경로(83–86행).** 결정: 이 카드가 REQ-RGR-009 로 명명한다. 래퍼 범위 이름 `처리기 실패` (와 바이너리 부재의 `바이너리 없음`)를 002 가 정의하고, SPEC-001 의 열두 처리기 어휘는 그대로 둔다. 처리기가 도달 불가능한 경로만 래퍼가 이름 붙인다는 경계를 REQ 본문에 새겨 어휘가 두 층에서 섞이는 것을 막는다.
5. **F2 잔여 — 게이트 실행 잠금의 링크 추종.** 결정: 부채 행(상한 명시). 근거: 블록 잠금 획득은 이미 no-follow 이고(이 세션 관측, `acquireGateLockNoFollowImpl` 두 플랫폼), 남는 표면은 `moai gate`(SPEC-GATE-THREE-AXES-001)의 잠금 파일이다. 수리는 두 줄이지만 다른 하위시스템의 파일과 테스트를 이 카드의 승인 목록으로 끌어온다 — M5·M6 표면 밖이다. 상한과 승격 조건은 spec.md 「Out of Scope — 게이트 실행 잠금의 no-follow 전환」에 있다.
6. **iter3 부채 N1·N2·N3·F3·F6·F7.** N2(Go 처리기 오류 경로 도달 불가)는 열린 질문 1 로 리더 판정 대상이다. N1(SPEC-001 `spec.md:53` 문장), N3(잔여 행 표기 불일치), F6(decision-index 폐기 식별자 인용)은 기록 처분 — 이 카드는 SPEC-001 파일을 바꾸지 않는다. N1 은 카드 보도이며 이 세션이 검증하지 않았다(Gap). F3(잠금 TTL 초과 경로와 윈도 소유자 파괴 테스트 부재)은 부채로 기록하고, 그 검증 강화는 M7 회귀 정비에서 `TestReviewGate_BlockLockRefusesSymlink` 재실행과 함께 관측만 한다(수리는 범위 밖). F7(SF-5 테스트 이름, SF-8 부재 기본값)은 선택 항목으로 연기한다.
7. **B1 — iter10 필수 항목.** (1) `internal/template/review_gate_registration_test.go` 와 `internal/cli/codex_review_gate_wiring_test.go` 를 §C M6 행과 기대값 변경 목록에 묶었다. (2)의 래퍼 기대값은 AC-RGR-008 이 본다. (3) `internal/cli/review_gate_config_key_test.go` 기대 항목을 D11 커밋값 판독에 맞춰 강화한다(§C M5 행, AC-RGR-005).
8. **B2 — 덮어쓰기 컴파일 전제.** 이 계획의 덮어쓰기 검사(§D 도입부, M5·M6)와 acceptance.md §D·§E 의 같은 검사마다 같은 문장을 둔다: 덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 하고, 컴파일 실패는 `tdd-result-contract.md` 의 TOOL_FAILURE 이며 RED 로 세지 않는다.

## §C 기술 접근 — 파일 지도

현재 역할은 이 세션의 관측(spec.md §A.2)에서 확인한 것이다. 줄 번호는 기준 코드 트리(`8249bff8c3`)의 값이다.

| 파일 | 현재 역할 (근거) | 관련 마일스톤 |
|---|---|---|
| `internal/cli/codex_review_delay.go` | 킥 마커 `codexReviewKickMarkerPath`(:82, 버킷 이름)·`kickInFlight`(:96, 배타적 생성, 내용 시각 한 줄 :112–117). pid 소유·완료 해제·고아 회수가 더해질 곳 | M5 |
| `internal/cli/codex_review_receipt.go` | 배경 검토의 판정 기록 단계(design.md §3 (가) 앵커). REQ-RGR-001 의 해제가 붙는 지점. 행 위치의 재확인은 M1 기준선 측정이 한다 | M5 |
| `internal/cli/codex_review_gate.go` | 캐시 재사용 분기(:227–259, inconclusive 재사용이 REQ-RGR-004 로 바뀐다), 킥 분기(:261–283), 차단 예약 `codexReviewReserveBlock`(:380), `tree_scope` 분기(:199–201), `runCodexReviewGate`(:767, 입력 오류 :770, 처리기 오류 :777–779), 상태 상수(:119–130) | M5·M6 |
| `internal/cli/mcp_codex.go` | `readCodexReviewGateEnabled`(:2602, 작업 트리 판독 — D11 전환 대상) | M5 |
| `internal/cli/codex_review_tree_scope.go` | `treeScopeSkipApplies`(:132 — 사유 판별이 더해질 곳) | M6 |
| `.claude/hooks/moai/handle-codex-review-gate.sh` + `internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh` | 셸 자체 게이트(32–60행)와 조용한 종료 35·36·60·79–80행, 처리기 호출 뒤 무조건 `exit 0`(83–86행). 두 면은 이 세션 `cmp` 로 동일 확인 | M6 |
| `internal/cli/codex_review_cache_test.go` | 기존 테스트. `TestReviewGate_InconclusiveReceiptIsNotPassEvidence`(셀 B-2 t1674 노트의 `:236–270` 표기; 현 트리 행은 M1 이 다시 확인)은 REQ-RGR-004 의 기대값 변경 대상 | M5 |
| `internal/cli/codex_review_delay_test.go` | 기존 킥·마커 테스트. M5 의 기존 기대값 중 마커 관련 항목의 변경 대상 | M5 |
| `internal/cli/codex_review_gate_concurrency_test.go` | `TestReviewGate_ConcurrentBlockCap`(:23) — R-1 회귀 가드(AC-RGR-011). 변경 없음, 재실행 | M5·M7 |
| `internal/cli/codex_review_gate_symlink_test.go` | `TestReviewGate_BlockLockRefusesSymlink`(:21) — 블록 잠금 no-follow 가드. 변경 없음, 재실행 | M7 |
| `internal/cli/review_gate_config_key_test.go` | 작업 트리 중첩 키 경로 고정(:52–78), `tree_scope` 행(:106). D11 강화 대상: 작업 트리만의 변경을 무시하는 경우, HEAD 판독, 판독 실패 시 fail-open + `커밋 값 읽기 실패`(B1 (3)) | M5 |
| `internal/cli/codex_review_inconclusive_test.go` | **신규 RED 테스트 파일**(카드 계획명 유지). AC-RGR-004 | M5 |
| `internal/cli/codex_review_marker_lifetime_test.go` | **신규 RED 테스트 파일**(카드 계획의 마커 가족을 inconclusive 파일에서 갈라 둔 조정 — 카드의 다른 계획 파일 `codex_review_entry_delivery_test.go` 는 SPEC-001 M2 가 이미 만들어 두어 신규 생성 대상이 아니다). AC-RGR-001·002·003 | M5 |
| `internal/cli/codex_review_run_switch_test.go` | **신규 RED 테스트 파일**. 커밋값 판독·재정의 파일·출처 기록. AC-RGR-005·006·007 | M5 |
| `internal/cli/codex_review_gate_test.go` | 기존 게이트 메시지 테스트. `tree_scope 건너뜀` 사유 판별의 신규 테스트 함수는 이 파일에 둔다 | M6 |
| `internal/template/review_gate_registration_test.go` | `TestReviewGateWrappersSelfGateBeforeBinaryResolution`(:214 — 게이트 래퍼 행이 REQ-RGR-008 으로 기대값이 바뀐다; 진입 래퍼 행은 유지), timeout 등록(:24). B1 (1) | M6 |
| `internal/cli/codex_review_gate_wiring_test.go` | `runCodexReviewGate` fail-open 고정(`:73` invalid stdin, `:88` empty stdin, `:111` happy path, `:142` block 전달). 기대값 변경은 **조건부**다 — 열린 질문 1 의 리더 판정이 처리기 오류 경로에 이름을 더하면 그때 함께 바꾼다. B1 (1) | M6 (조건부) |
| `.moai/specs/SPEC-REVIEW-GATE-REPAIR-002/progress.md` (공통 승인 파일) | RED 원문과 기대값 변경 기록, 회귀 비교표가 §E.2 에 쌓인다 | M1·M5·M6·M7 (공통) |

파일 목록은 이 §C 가 정한다. run 단계가 목록 밖 파일을 바꿔야 하면 멈추고 보고한다(마일스톤 종료 시 트리 변화 점검, §E). 바꾸지 않는 파일: `.moai/config/sections/workflow.yaml`, `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/**`, `internal/cli/multi_review_gate.go`, `internal/cli/gate_lock*.go`, `internal/cli/codex_review_entry_delivery_test.go`(이미 SPEC-001 M2 분).

## §D 마일스톤 (우선순위 순)

**RED 선행의 증명 방식.** 각 마일스톤의 RED 선행은 커밋 그래프만으로 검증한다: 마일스톤 커밋 M 의 첫 부모 P 의 트리에 덮어쓰기 단위를 얹어 기록된 명령을 실행하면, 기록된 단언에서 EXPECTED_RED 로 실패해야 한다. 덮어쓰기 단위는 마일스톤 커밋이 추가하거나 바꾼 `_test.go` 파일의 그 커밋 시점 내용이다. **덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 한다. 컴파일 실패는 `tdd-result-contract.md` 의 TOOL_FAILURE 이며 RED 로 세지 않는다.** 이 SPEC 은 승격을 하지 않으므로 순서의 증거는 커밋 그래프와 progress.md §E.2 기록뿐이다. 검사 대상은 RED 단계가 있는 M5·M6 다(M1·M7 은 제외). 마일스톤 순서는 데이터 모델(표식) → 사용자 표면(래퍼) → 기계적 정리다. 각 마일스톤은 한 커밋이고, 래퍼의 두 면은 그 커밋에서 함께 갱신한다(spec.md §D.3).

### M1 — 기준선 측정과 열린 질문 확인 (우선순위 High, 차단 게이트)

- plan-audit 통과와 열린 질문 1(처리기 오류 경로 명명)의 리더 판정을 확인한다. 판정 전에는 M6 로 넘어가지 않는다(spec.md §D.1).
- 기대값 변경 대상 테스트의 현행 기준선을 원문으로 기록한다: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence`(M5), `TestReviewGateWrappersSelfGateBeforeBinaryResolution`(M6). `codex_review_receipt.go` 의 판정 기록 단계와 `codex_review_cache_test.go` 의 현재 행 번호를 확인해 §C 표를 갱신한다.
- `treeScopeSkipApplies` 의 스킵 행렬(설정 유무·세션 클래스·WT- 증거의 조합)을 관측해 REQ-RGR-010 의 사유 판별 기준을 확정한다.
- 완료 조건: 기준선 원문이 progress.md §E.2 에 있고, 열린 질문 1 의 판정이 기록되어 있다.

### M5 — 표식 수명, inconclusive 재검토, 스위치 해석 (우선순위 High, 데이터 모델)

- 요구: REQ-RGR-001–007. 인수 기준: AC-RGR-001–007.
- RED: 판정 기록 뒤 표식이 풀리지 않는다(AC-RGR-001). 죽은 pid 의 표식과 판독 불가 표식이 회수되지 않는다(AC-RGR-002). 표식이 판정 뒤의 재검토를 막는다(AC-RGR-003). fresh inconclusive 턴 종료가 재검토를 시작하지 않는다(AC-RGR-004). 작업 트리만의 스위치 변경이 지켜지고 커밋값 판독 실패가 무메시지다(AC-RGR-005). 재정의 파일의 무시 조건이 없다(AC-RGR-006). 출처 기록이 없다(AC-RGR-007).
- GREEN: `kickInFlight` 표식에 소유자 pid 를 싣고(킥 런처가 자식 pid 를 기록), 판정 기록 단계가 표식을 풀며, 새 킥과 새 판정 앞에서 고아를 회수한다(fail-open). 캐시 적중 inconclusive 분기가 `inconclusive` 메시지와 함께 킥으로 이어지게 한다(표식이 동시 실행만 막는지 확인 — REQ-RGR-003). 커밋값 판독기(D11)와 재정의 파일 해석(design.md §5 판정 순서 (가)–(라)), member lane 조회, 출처 기록을 둔다.
- 기대값 변경: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` — inconclusive 캐시 재사용이 킥 억제에서 허용+재검토로 바뀐다. 변경 기록("이전 기대값 → 새 기대값 → 근거 REQ-RGR-004")은 progress.md §E.2 에 남긴다. `review_gate_config_key_test.go` 는 작업 트리 파일 픽스처에 HEAD 커밋 픽스처를 더해 강화한다(B1 (3)).
- 버킷 이름의 운명: 표식이 pid·완료 해제를 갖추면 버킷 만료는 추가 안전망이다. 이름에 버킷 성분을 남길지는 run 단계가 정하며, REQ-RGR-001–003 의 의미 규약은 어느 쪽이든 같다.

### M6 — 래퍼의 바이너리 호출과 래퍼 범위 명명 (우선순위 High, 사용자 표면)

- 요구: REQ-RGR-008–010. 인수 기준: AC-RGR-008–010.
- RED: 게이트 래퍼가 스위치를 셸에서 해석해 꺼짐일 때 무출력으로 끝난다(AC-RGR-008). 바이너리 부재와 처리기 실패가 상태 이름 없이 사라진다(AC-RGR-009). `tree_scope 건너뜀` 메시지가 사유를 구별하지 않는다(AC-RGR-010).
- GREEN: `handle-codex-review-gate.sh` 의 셸 자체 게이트(32–60행)를 없애고 모든 턴 종료에서 처리기를 부른다. 35·36·60행은 경로가 사라진다. 바이너리 부재(79–80행)와 처리기 실패(83–86행, 종료 상태 확인)에서 래퍼가 systemMessage 를 실어 0 으로 끝난다. `treeScopeSkipApplies` 의 사유 판별을 메시지에 더한다. 래퍼의 두 면을 한 커밋에서 맞춘다.
- 기대값 변경: `TestReviewGateWrappersSelfGateBeforeBinaryResolution` — 게이트 래퍼 행의 전제(자체 게이트 선행)가 "모든 턴 종료에서 처리기 호출"로 바뀐다. 진입 래퍼 행은 유지된다. `codex_review_gate_wiring_test.go` 는 열린 질문 1 판정이 이름을 더하면 그때 함께 바꾼다(조건부).
- 래퍼 검증은 격리 실행으로 한다: 각 경로의 선행 조건만 실패하게 만든 실행에서 종료 코드·stdout JSON·systemMessage 를 progress.md §E.2 에 원문으로 남긴다(SPEC-001 v0.2.10·v0.2.11 이 관측한 방식의 연장).
- 비용 수용의 재확인: 게이트가 꺼진 배포 기본 프로젝트는 턴 종료마다 바이너리 기동 1회를 낸다(v0.2.6/D4 가 수용한 대가, design.md §5 「두 층」).

### M7 — 회귀 정비와 후속 등록 (우선순위 Low, 기계적)

- 회귀 가드 전체를 변경 전후로 다시 실행하고 비교표를 progress.md §E.2 에 남긴다: AC-RGR-011(`TestReviewGate_ConcurrentBlockCap` — R-1 가드), AC-RGR-012(진입 전달·pass 재사용·건너뜀 결정 가드 — SPEC-001 REQ-RG-001·002·005·006 보호), `TestReviewGate_BlockLockRefusesSymlink`, `TestReviewGate_DisabledAllows` 계열.
- 후속 등록: 상속 항목 1(F5 대상 SPEC 개정), 5(F2 잔여 부채), 6(F3·F7)을 후속 항목으로 기록한다. 이 SPEC 의 Follow-up 항목을 그대로 옮긴다.

M2·M3 은 SPEC-001 이 소유하고 M4 는 t1662 의 몫이다. 이 SPEC 의 마일스톤은 M1·M5·M6·M7 이며, 번호 간격은 상속을 보이기 위한 의도다.

## §E 자기 검증 — run 단계 명령 형식

- 기준 코드 확인: `git diff --stat 8249bff8c3 HEAD -- ':!.moai/specs/SPEC-REVIEW-GATE-REPAIR-002'` 의 출력이 비어 있어야 한다. 비어 있지 않으면 멈추고 보고한다. 설치된 빌드 버전 문자열로 기준을 판정하지 않는다.
- 테스트 호출 형식: `go test -count=1 -v -run '^<TestName>$' ./internal/cli/` 처럼 단일 호출로, 출력을 원문으로 남긴다. 종료 코드는 별도 필드로 기록한다.
- 빌드 출처: 측정은 트리의 소스에서 만든 빌드로 한다. 설치된 빌드의 결과를 트리의 증거로 인용하지 않는다.
- 공허 검사 방지: 실행된 테스트 수(`--- PASS` 줄 수)를 함께 기록한다. 0이면 측정 불가로 보고한다.
- 트리 변화 점검: 각 마일스톤 종료 시 `git status --short` 를 본다. 변화가 그 마일스톤의 승인 파일 목록(§C) 밖에 있으면 멈추고 보고한다.
- 변경 세트 점검: M6 커밋마다 `git show --stat <sha>` 로 래퍼의 두 면이 같은 커밋에 있는지 확인한다.
- RED 검사: §D 도입부의 커밋 그래프 검사(덮어쓰기 단위 + 컴파일 전제)를 M5·M6 커밋마다 실행하고 원문을 §E.2 에 남긴다.

## §F 위험

1. **D11 전환은 테스트 픽스처를 바꾼다.** 커밋값 판독은 작업 트리 파일로는 검증할 수 없으므로 `review_gate_config_key_test.go` 에 임시 저장소 커밋 픽스처가 필요하다. 픽스처 비용을 이유로 작업 트리 판독을 유지하는 것이 최악의 귀결이다(D11 위반).
2. **래퍼 기대값 변경의 회귀 오인.** `TestReviewGateWrappersSelfGateBeforeBinaryResolution` 은 현재 PASS 이며 그 PASS 가 제거할 동작을 고정한다. 기대값 변경 근거(REQ-RGR-008, v0.2.6 판정)를 progress.md 에 남긴다.
3. **표식 스키마 과도기.** 이전 형식 표식(시각 한 줄, 빈 파일)이 디스크에 남아 있을 수 있다. REQ-RGR-002 가 이전 형식을 고아로 규정하므로 회수 규칙이 흡수한다. 회수가 fail-open 인지 M5 RED 에서 확인한다.
4. **member lane 판정의 경로 정규화.** 정규화 불가 트리·행은 member 가 아니다(커밋값을 따른다 — 레인 기본값 (나)). 심볼릭 링크 풀기의 실패 경로를 M5 RED 에 픽스처로 둔다.
5. **inconclusive 무간격 재검토의 폭주.** Q4 = (a) 는 상한을 두지 않는다. 표식이 동시 실행만 막는 REQ-RGR-003 을 어기면(표식이 판정 뒤 재검토를 막거나, 반대로 동시 실행을 못 막으면) 같은 inconclusive 키에서 킥이 겹치거나 무한히 반복된다. M5 RED 가 두 방향 모두를 단언한다(AC-RGR-003·004).
6. **템플릿 미러 누락.** 래퍼가 두 면에 있다. 한 면만 바꾸면 배포본과 저장소 본이 갈린다(M6 변경 세트 점검).
7. **wiring 테스트의 조건부 변경.** 열린 질문 1 판정 전에 M6 가 처리기 오류 분기를 바꾸면 SPEC-001 의 예외 기록과 충돌한다. §D 순서 구속 1 이 막는다.
8. **`tree_scope` 스킵 행렬의 미재측정.** 카드 기록과 코드 주석으로만 파악했고 이 세션은 행렬 전체를 재측정하지 않았다. M1 이 확정한다. 확정 전에 REQ-RGR-010 의 사유 판별 기준을 굳히지 않는다.
9. **설치 빌드 오인.** 설치된 빌드의 결과를 트리의 증거로 인용하면 도구 출처 규칙을 어긴다.
10. **N1 은 카드 보도다.** SPEC-001 `spec.md:53` 문장의 오류 내용은 이 세션이 검증하지 않았다. 기록 처분만 하고 본문을 인용하지 않는다.

## §G 안티패턴

- 작업 트리의 미커밋 `workflow.yaml` 값을 스위치 기본값으로 읽는 것(D11 위반).
- 표식을 판정 뒤의 재검토 차단에 쓰는 것(REQ-RGR-003 위반 — 캐시 관할 침범).
- inconclusive 재검토에 간격이나 상한을 몰래 두는 것(Q4 = (a) 위반), 또는 표식 없이 무간격 재시작을 두는 것(동시 킥 폭주).
- 열두 처리기 어휘에 13번째 이름을 더하거나, 래퍼 범위 두 이름을 처리기 어휘와 섞는 것(어휘 경계 위반).
- 처리기 오류 경로를 열린 질문 1 판정 없이 고치는 것(§D 순서 구속 1 위반).
- 래퍼의 두 면을 다른 커밋에 두는 것.
- 결함 고정 테스트를 삭제해서 빨간불을 피하는 것. 기대값 변경만 허용된다.
- 매칭이 0개인 실행을 통과로 세는 것(공허한 통과).

## 열린 질문

1. **[리더 판정 — 필수] Go 처리기 오류 경로의 상태 명명.** `runCodexReviewGate` 의 처리기 오류 분기(`internal/cli/codex_review_gate.go:776–781`; SPEC-001 v0.2.11 표기 773–777행)는 처리기 오류를 표준 오류 한 줄로 삼키고 빈 허용을 낸다. SPEC-001 은 "처분의 확정은 리더 판정 대상"으로 남겼다. 선택지: (a) 처리기 도달 가능 경로이므로 처리기 범위 이름(예: `처리기 오류`)을 더한다 — SPEC-001 어휘 개정이 따르고, `codex_review_gate_wiring_test.go` 기대값이 함께 바뀐다. (b) 문서화된 예외로 남긴다 — wiring 테스트는 불변이고, 이 카드는 AC-RGR-009 로 래퍼 실패만 명명한다. 판정 전 M6 GREEN 은 (b) 대로 둔다.
2. **[기록] N1 — SPEC-001 `spec.md:53` 문장의 오류.** 카드 보도이며 미검증(Gap). SPEC-001 개정은 범위 밖이다.
3. **[선택, 연기] F7** — SF-5 테스트 이름과 SF-8 부재 기본값의 정리. 이 SPEC 은 다루지 않는다.

## §H 참조

- spec.md · acceptance.md · progress.md
- `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` spec.md (HISTORY v0.2.3–v0.2.11, REQ-RG-007 범위 한계 표) · design.md §3·§5·§7 · plan.md §C 패턴 · acceptance.md §C 셀 B-2 · decision-index.md (Q2–Q6, D3·D4·D11, 레인 기본값 (가)/(나))
- `.claude/rules/moai/workflow/tdd-result-contract.md` (RED 기록 형식, TOOL_FAILURE)
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (도구 출처)
- `.claude/rules/moai/development/verification-completeness.md` §2.1 (RED-now 셀의 네 요소)
