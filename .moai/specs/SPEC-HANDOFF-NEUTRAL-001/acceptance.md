# Acceptance — SPEC-HANDOFF-NEUTRAL-001 (card t1273, M1 run 범위)

> 판정 명령은 전부 환경 스크럽 단일 복합 호출로 실행한다: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && <명령>`. RED 관측은 각 테스트의 최초 작성 시점(구현 전) FAIL 출력으로 확보하며, 그 출력이 이 문서의 red 근거 셀을 채운다 — 관측 없이 red로 기록하지 않는다.

## 등급 구분

- **release-blocking** (R): 이 트리에서 재실행 가능한 go-test 판정. RED(구현 전 FAIL)를 관측한 뒤 GREEN을 관측해야 채택 완료.
- **regression-guard** (G): 외부 상태(Codex 쿼터·격리 환경)에 의존하는 LIVE 관측 — 현재 트리에서 임의 재실행 불가능하므로 release-blocking에서 제외하고 관측 기록 자체가 산출물 (verification-completeness §2.1 undecidable disposition).

## AC 표

| AC | REQ | 등급 | 판정 명령 (단일 호출) | 기대 출력 | red 근거 (관측 예정) |
|---|---|---|---|---|---|
| AC-HN-001 | 001 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_(PendingSource\|ConsumedFallback\|NoHandoffErrors)' -v` | 3 케이스 전부 `--- PASS` (NoHandoff는 exit-코드/오류 메시지 판정) | 최초 작성 시 `undefined: newHandoffShowCmd` 컴파일 실패 또는 FAIL — 구현 전 출력을 그대로 채택 |
| AC-HN-002 | 002 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_DoesNotConsume' -v` | PASS — show 2회 호출 후 pending.json 존재·내용 바이트 동일 | 동일 (기능 부재 FAIL) |
| AC-HN-003 | 003 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_JSONOutput' -v` | PASS — JSON에 body·spec·phase·출처 포함 | 동일 |
| AC-HN-004 | 004 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_LocaleHeader' -v` | PASS — ko·en 저장 언어별 헤더 분기 (본문 미변경 포함) | 동일 |
| AC-HN-005 | 005 | R | `go test ./internal/worktree/ -count=1 -run 'TestNew_SeedsCodexHooksJson' -v` (실제 패키지 경로는 구현 시 확정) | PASS — 새 워크트리에 `.codex/hooks.json` 존재, moai 소유 커맨드 포함 | 시딩 코드 부재로 신규 테스트 FAIL |
| AC-HN-006 | 006 | R | `go test ./internal/cli/ -count=1 -run 'TestEnterWorktree_SeedsMissingCodexHooks' -v` (동일·경로 확정) | PASS — 부재 트리 진입 시 채워짐, 존재 트리 진입 시 바이트 불변 | 동일 |
| AC-HN-007 | 007 | R | `go test ./internal/worktree/ -count=1 -run 'TestNew_SeedFailureFailOpen' -v` | PASS — 시딩 강제 실패(경로 오염)에도 워크트리 생성 성공 + stderr 진단 | 동일 |
| AC-HN-008 | 008 | R (조건부) | `go test ./internal/codexadapter/ -count=1 -run 'TestMapOutput_SessionStartAdditionalContext' -v` | 관문 (b) 통과 시에만 테스트가 작성되고 PASS — SessionStart 출력의 systemMessage/additionalContext가 매핑 결과에 존재 | 관문 (b) 기각 시 본 AC는 REQ-HN-008과 함께 기각 종결 (기각 사유 기록, 테스트 미작성) |
| AC-HN-009 | 009 | G | (LIVE 절차 — design §C) 관문당 관측 출력을 `.moai/reports/t1273/` 에 파일로 | 관문 (a) moai 훅 발화 관측 / (b) 마커 도달 또는 부재 / (c) 길이·강등 여부 — 각각 관측 기록 | 쿼터 회복(9/28 14:37) 전 실행 불가 — 외부 의존이므로 regression-guard |
| AC-HN-010 | 010 | R | `go test ./internal/cli/ ./internal/hook/ -count=1 -run 'TestRenderHandoffContext\|TestHandoffSave\|TestNewHandoffSaveCmd' -v` | 기존 테스트 전부 PASS — M1이 렌더 형식·save 표면을 건드리지 않았음 | red 없음(보존 AC) — 대신 이 테스트들이 구현 후에도 변함없이 통과함이 green 판정 |
| AC-HN-011 | (형식) | R | `go vet ./internal/cli/ ./internal/codexadapter/ && gofmt -l internal/cli/handoff.go internal/codexadapter/output.go` | vet exit 0, gofmt 빈 목록 | red 없음(형식 게이트) |

## 채택 순서 계약

1. run-phase 첫 액트: AC-HN-001..007 테스트 파일을 먼저 작성해 **RED 출력을 관측·기록** (`.moai/reports/t1273/red-*.txt`).
2. 구현 최소 커밋 단위: ① show(+테스트) → ② materializer 시딩+런처 보완(+테스트) → ③ (관문 b 통과 시) 어댑터 매핑(+테스트).
3. AC-HN-009 LIVE는 쿼터 회복 후 격리 실행 — M1 종결 조건이 아니라 판정 기록 의무(P1 채택/기각 확정)로 바인딩.
4. GREEN 일괄 관측 후 AC 표에 각 결과(명령+출력 발췌)를 회기한다.
