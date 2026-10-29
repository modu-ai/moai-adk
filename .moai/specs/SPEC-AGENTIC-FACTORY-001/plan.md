# SPEC-AGENTIC-FACTORY-001 — 구현 계획

순서는 바뀔 가능성이 큰 결정부터 둔다(데이터 모델, 역할·명령 경계, 정책). 기계적인 단계는 뒤에 둔다. 시간 예측은 쓰지 않는다.

## §A 맥락

- 규범 원천: Issue #1948(유일). 계약 범위는 계약 + 정책 + 설정 초안이며, 이 lane이 유일한 writer다.
- 측정 트리: 0.1.0은 `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4`, 0.2.0의 추가 측정(활성 판정, 보호 설정, 소스 문자열)은 `e14b5bbac2eb2499cbaa3ae9434595001c91cb39`(코드는 같음).
- 산출물: Tier L 다섯 개(spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md.

## §B 알려진 문제

- 세션 구조·명칭·명령은 상위 설계 권고를 PROPOSAL로 반영했지만 리더·사용자 확인 전까지 미확정이다(design.md §13). 이를 근거로 코드를 쓰지 않는다. 제안 명령 `moai factory status`는 기존 명령과 이름이 겹친다(OQ-12).
- `run`의 범위 정의가 두 갈래다(design.md §7.3).
- 설정·Codex 파일에는 쓰기 배정이 없다(t1984, lane-14/16 ACK 대기).
- 충돌 PR: #1937·#1944(설정), #1943·#1916·#1921(factory, launcher). 카드 t1965는 읽기 전용 설계 입력.

## §C 사전 점검 (run 진입 전)

1. 기존 리더·writer의 안전 중단 ACK와 단일 owner·writer 확정(OQ-8).
2. 시나리오용 실제 Issue 선정과 운영자 승인(OQ-1).
3. 상위 설계 권고(PROPOSAL)의 확정 또는 미확정 상태로 진행할지에 대한 결정(OQ-2).
4. `go test ./internal/agenticfactory/... -count=1`로 대상 부재 참고 관측 재측정(acceptance.md EV-REF-0). 이것은 RED-now가 아니다. 각 AC의 RED-now는 해당 시험을 먼저 쓰고 `EXPECTED_RED`로 기록한다.
5. AC-AF-019 기준값 재측정(acceptance.md §B.2 명령).

## §D 제약

- 강제 종료·터미널 종료·작업 트리 삭제·미커밋 변경 폐기 금지.
- 기존 PR 닫기·브랜치 덮어쓰기·카드 일괄 done 금지.
- GLM managed 사용 금지, Codex는 명시 설정 시에만.
- 검사 삭제·보호 약화·관리자 우회 금지.

## §E 자기 검증

- 산출물마다 `moai spec lint SPEC-AGENTIC-FACTORY-001` 결과를 progress.md §E.1에 남긴다.

## §F 마일스톤 (우선순위 순)

### M1 — 데이터 모델과 claim 결정 [우선순위 High, 바뀔 가능성 가장 큼]

- `state` / `claim` 범위의 상태 열거, 전이표, claim 기록 모양(Issue당 하나, attempt 순번)을 코드 타입으로 굳힌다.
- 이 단계에서 확정된 명칭과 `run` 범위 정의(OQ-2), 배정 epoch(design.md §14)를 반영한다. 확정이 없으면 미확정 표시를 유지한 채 진행할지 운영자가 정한다.
- 대상 AC: AC-AF-008, AC-AF-021, AC-AF-025(#1926·#1924 시험).

### M2 — 소유권·복구 [High]

- 원자적 claim, 중복 dispatch 거절, 소유권 확인 실패 시 쓰기 거절, 실제 실행 우선 조회, 살아 있는 소유자 인계 금지.
- 활성 판정은 `homestate.ClassifyOwnerWith`를 정본으로 쓴다(spec.md §F.2). claim 때 pid와 프로세스 시작 시각을 함께 기록한다. `FactoryProcessAlive`는 claim 판정에 쓰지 않는다.
- 잠금 기반은 `internal/factory`의 상태 잠금 기반을 재사용 후보로 검토하되, `slot_lease.go` 파일 자체는 수정하지 않는다.
- 대상 AC: AC-AF-009, 010, 025(#1938·#1847·#1813 시험). 시나리오 AC-AF-001·002·004의 부품은 여기서 만들지만, 시나리오 판정은 묶음이 끝까지 도는 M4 뒤다.

### M3 — 전달 상태와 원장 [High]

- 네 단계 전달 상태, 원장 필드 권한, 비밀 가림, 저장소-거울 조정.
- Issue 쓰기는 기존 `internal/github` 도우미를 넓히지 않고 `ledger` 범위에서 감싼다.
- 대상 AC: AC-AF-011, 012. 시나리오 AC-AF-003의 부품은 여기서 만든다.

### M4 — 병합 게이트와 증거 분리 [High]

- 병합 수용 조건, 증거 저장소 분리, 병합 결과 미상 처리, `done` 수용 조건(원격 반영 + 로컬 동기화 + 배포 또는 N/A), 종료·다음 dispatch.
- 증거 생산자 결속(design §11)과 **공급자 중립 리뷰 경로**(design §11.1: 리뷰 세션 또는 운영자 기록 리뷰)를 여기서 넣는다. 이 경로가 있어야 Codex 미설정 상태에서 병합 게이트가 통과할 수 있고, 시나리오 06·07이 이 마일스톤 뒤 초록이 될 수 있다(spec.md §F.7 D2 개정). M4까지 `internal/agenticfactory`에는 Codex 호출 지점이 없으며, 묶음은 Codex 호출 계수기 0을 함께 단언한다. Codex 명시 설정 관문과 그 계수기 시험(AC-AF-013)은 M5다.
- 시나리오 하네스 `TestScenarioSuite`(acceptance.md §A.1: 05→06 증거 생산 단계, 분기 사례 격리, 재실행 규칙, 두 번째 Issue 사후 상태 포함)를 여기서 완성한다. 실제 Issue 쓰기와 보호된 `main` 병합은 운영자 승인 뒤에만 실행한다.
- 대상 AC: AC-AF-018, 시나리오 AC-AF-001~007(이 마일스톤 뒤 초록), AC-AF-019(regression-guard, 시나리오 묶음 전후 측정).

### M5 — 정책과 설정 초안 [Medium]

- GLM 제외, Codex 명시 설정 관문, 인증·429 처리, 환경 매트릭스, 브라우저 어댑터, 설정 검증.
- 실제 설정 파일 삽입은 이 마일스톤에 포함하지 않는다(§F.3 2단계 이후, 설정 담당 레인).
- 실행한 Codex 전환 단계는 progress.md `## §J Codex Migration Stage Record`에 `| CX-<n> |` 행으로 적는다(design.md §9.7). 세 단계는 각각 별도 커밋이며, 순서는 커밋 그래프로 증명한다(AC-AF-020).
- 대상 AC: AC-AF-013, 014, 015, 016, 017, 020.

### M6 — 명령 계층과 호환 [Medium, 기계적]

- **첫 커밋은 호환 기준 golden 파일만** 담는다(`internal/cli/testdata/agentic_factory_compat/`, design.md §9.6): 도움말·거절 문구·종료 코드·lane 레이블/세션 ID/owner/handle 스냅숏을 변경 전 빌드로 기록한다. S1 명령 등록은 그 다음 커밋부터다(VCI §2.3).
- §8 명령의 S1 도입(기존 명령 불변), 관찰 모니터 표시 항목(§8.1)과 무변경, 재연결·선택·재개 계약.
- 호환 시험은 launcher 담당 레인이 `internal/cli`에 쓴다.
- 대상 AC: AC-AF-022, 023, 024.

### F.3 기존 Codex 기본값의 전환 순서 (지금은 끄지 않음)

1. **관문 도입**: AgenticFactory 경로에 `workflow.agentic_factory.codex.enabled` 관문을 넣는다. 기존 `workflow.codex.*`, `audit.model: codex`, `audit.gates.codex: required`는 그대로다.
2. **미설정 0건 검증**: AC-AF-013 PASS를 기록한다.
3. **기본값 변경**: 그 뒤에만 기존 Codex 기본값(이 저장소의 로컬 설정과 PR #1937이 다루는 출시 기본값 포함)의 변경을 별도 작업으로 제안한다. 기존 기능은 대체 동작이 검증된 뒤에만 제거한다.

모든 단계에서 점검, 독립 감사, 인계 의무는 그대로 유지한다. PR #1937과의 관계는 OQ-9다.

### F.4 호환·단계 폐기 (미확정)

design.md §9.4의 S1 → S2 → S3 → S4(lane-22 대응표와 같은 네 단계)를 따른다. 각 단계는 진입 조건과 증거가 있어야 하고, 이관 행 C1~C11과 PR 착지 순서 후보는 design.md §9.3·§9.5에 있다. 명칭은 상위 설계 권고를 PROPOSAL로 반영했으며 확정 전이다(OQ-2).

### F.5 Phase 2 경계

남은 카드의 Issue 이관은 이 SPEC의 마일스톤이 아니다. AgenticFactory 수용 기준 충족과 재개 승인 뒤의 하위 단계이며, 이관 자체는 구현 완료로 세지 않는다.

## §G 위험

| 위험 | 대응 |
|---|---|
| GitHub API 실패·429로 원장 거울이 늦어짐 | 단일 규칙(REQ-AF-014, design.md §2.1a): 실패한 동작에 차단을 기록하고 재시도 고리는 없다. 거울 쓰기만 실패하면 `mirror_state: blocked(<사유>)`로 거울에 기록하고 Issue `state`는 바꾸지 않는다. 실행은 저장소 기준으로 계속한다 |
| 새 claim과 기존 카드 임대가 같은 작업을 이중으로 소유 | Phase 2 전까지 Issue 경로와 카드 경로가 같은 작업을 동시에 맡지 않도록 claim 시 기존 카드 임대 존재를 읽기 전용으로 확인 |
| 시나리오 재실행마다 실제 Issue 한 쌍과 실제 병합이 새로 든다 | 시나리오 묶음은 단위 AC가 모두 초록이 된 뒤 한 번 돌린다. 재실행 규칙은 acceptance.md §A.1 |
| 증거 생산자를 증거 파일이 스스로 주장 | 생산자 결속은 ControlService가 등록 때 관측한 값만 쓴다(design §11) |
| `slot` 명칭 충돌 | 확정 전까지 새 실행 자리 이름을 코드 식별자로 쓰지 않음(OQ-2) |
| 시나리오 시험이 실제 Issue에 씀 | 운영자 승인 뒤에만 실행, 시험용 Issue를 하나로 고정 |

## §H 참조

- `spec.md`, `design.md`, `acceptance.md`, `research.md`
- `.claude/rules/moai/workflow/factory-dispatch*.md`, `.claude/rules/local/repo-local-pr-policy.md`
