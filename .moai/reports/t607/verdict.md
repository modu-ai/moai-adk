# t607 판정 기록 — 무거운 테스트 실행 슬롯

## 1. 방향 결정 (plan 진입 전)

- **결정**: B안 — 운영자 결정, 리드가 전달
- **수신**: 2026-09-12T00:10+0900, 리드 세션(`lead`)의 cross-session 메시지. /clear로 유실된 결정을 다시 전달받은 것이다
- **기록한 곳**: lane-10, 워크트리 `.claude/worktrees/t607`, 브랜치 `WT-heavy-test-slot`, 기록 시점 HEAD `85868148c`

### 결정 내용 (리드 메시지 요지)

| 항목 | 결정 |
|---|---|
| 표면 | 범용 자원 임대 명령 — `moai slot acquire\|status\|release --resource <name>` 류 |
| 강제 | 선택형 PreToolUse 가드, 기본값 꺼짐 |
| 배포 범위 | 템플릿과 바이너리로 배포한다(로컬 전용이 아니다). 기본값은 꺼짐, 템플릿 문서는 최소화 |
| 기록 필드 | 자원명 · 보유 세션 id/이름 · pid(스테일 판정용) · 명령 · 시작 시각 · 선언 상한 |
| 대조군 | 필수. 표면 없음 → 두 세션 동시 시작 성공 / 표면 있음 → 한쪽 거절. 테스트로 재현한다 |
| 재사용 경계 | kanban acquire/release/stale 코드 방식만 재사용한다. 통합 창(`moai integration`)과는 분리한다 |
| 문서 편집 | `kanban-dispatch.md`, `gitflow-lane-protocol.md`는 t637(`WT-acquire-branch-record`)이 develop에 병합된 뒤에 고친다 |
| 착수 순서 | 로컬 develop `eb50af5a8` 흡수 → manager-spec으로 SPEC 작성 |

### 기각된 안

- A안(로컬 규칙만): 채택하지 않았다.
- C안(`moai integration` 확장): 채택하지 않았다. 사전 판단 근거는 가드가 `git merge`만 잡는다는 점, 기록 필드가 병합 대상 전용이라는 점, 창이 하나뿐이라 병합 대기와 테스트 대기가 서로를 막는다는 점이다.

### 사전 판단 때 밝힌 공백 (그대로 이어진다)

- lane-10 자기 슬롯에서 본 충돌 대상 `internal/cli/ptycaptest`는 `internal/cli`를 import하지 않는다. 그래서 무거운 패키지끼리 실제로 충돌한 장면은 직접 재현하지 못했다.
- 직접 관측한 증거는 확인 후 시작 사이의 경합뿐이다. `ps`로 비어 있음을 확인한 직후 다른 `go test`가 시작됐다.

## 2. t637 병합 상태 (문서 편집 보류 조건)

측정 명령: `git merge-base --is-ancestor WT-acquire-branch-record develop` / `... origin/develop`
2026-09-12T00:10+0900 기준 두 명령 모두 rc=1이었다. `f680dab46`은 `eb50af5a8`의 조상이 아니므로 문서 편집은 보류한다.

## 3. plan 단계 결과

| 단계 | 커밋 | 결과 |
|---|---|---|
| SPEC 초안 v0.1.0 | `1a178c174` | `SPEC-RESOURCE-SLOT-LEASE-001`, Tier M, REQ 16 / AC 16 |
| plan-audit 1회차 | `e50cfea93` | FAIL 0.78 — 막는 결함 D1–D6 (`plan-audit-iter1.md`) |
| SPEC 수리 v0.2.0 | `e79d6761f` | D1–D12 전부 반영 |
| plan-audit 2회차 | `759009244` | **PASS 0.90**, 막는 결함 없음 (`plan-audit-iter2.md`) |

run 단계로 넘길 선택 사항(2회차 N1–N4): N1은 CLI도 가드와 같은 루트 해석 함수를 쓰게 할 것(D1의 거울상), N2는 AC-RSL-003c 분류, N3은 `<TOOL_TOKENS>`를 파일로 빼 `grep -f`로 돌릴 것, N4는 no-git 행의 감사 로그 경로 명시다.

다음 관문은 Implementation Kickoff Approval이다. 승인 전에는 run에 들어가지 않는다. t637 게이트는 이 기록 시점에도 rc=1이다.
