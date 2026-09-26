# Acceptance — SPEC-HANDOFF-NEUTRAL-001 (card t1273, M1 run 범위)

> 판정 명령은 전부 환경 스크럽 단일 복합 호출로 실행한다: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && <명령>`.
>
> **Fixture 계약 (1차 감사 결함 정정)**: 저장소는 factory.db(`resume_handoffs`)다 — 테스트 fixture는 **`SavePending` 호출 또는 homestate 직접 세팅**으로 만든다. legacy `pending.json` fixture는 `ReadPending`의 읽기 호환 분기만 검증하는 별도 케이스에만 쓰고 본 판정(AC-HN-001 등)의 소스로 쓰지 않는다 — 파일 fixture만 심으면 DB가 비어 legacy 호환 분기로 공허 초록이 난다.
>
> **Red 셀 계약**: RED는 각 테스트 최초 작성 시점(구현 전) FAIL 출력으로 관측하며, red 파일(`.moai/reports/t1273/red-*.txt`) **첫 줄에 관측 시점 HEAD SHA를 기록**한다 — 관측 없이 red로 기록하지 않고, SHA 없이 red로 인정하지 않는다.

## 등급 구분

- **release-blocking** (R): 이 트리에서 재실행 가능한 go-test/grep 판정. RED(구현 전 FAIL) 관측 후 GREEN 관측으로 채택 완료 (보존 AC는 예외 표기).
- **regression-guard** (G): 외부 상태(Codex 쿼터·격리 환경)에 의존하는 LIVE 관측 — 임의 재실행 불가능하므로 관측 기록 자체가 산출물 (verification-completeness §2.1 undecidable disposition).

## AC 표

| AC | REQ | 등급 | 판정 명령 (단일 호출) | 기대 출력 | red 근거 (관측 예정) |
|---|---|---|---|---|---|
| AC-HN-001 | 001 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_(PendingSource\|ConsumedFallback\|NoHandoffErrors)' -v` | 3 케이스 전부 `--- PASS`. PendingSource=SavePending 세팅 후 Body verbatim 출력, ConsumedFallback=claimed→consumed 전이 후 재출력, NoHandoff=빈 DB에서 exit 1+안내 | 최초 작성 시 `newHandoffShowCmd` 부재 컴파일 실패 또는 FAIL — 구현 전 출력을 그대로 채택 |
| AC-HN-002 | 002 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_DoesNotMutateState' -v` | PASS — show 2회 호출 후 row의 status·consumed_at·Body 불변 (DB 재열람 대조) | 동일 |
| AC-HN-003 | 003 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_JSONOutput' -v` | PASS — JSON에 body·spec·phase·출처(pending/consumed) 포함 | 동일 |
| AC-HN-004 | 004 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_LocaleHeader' -v` | PASS — 저장 언어(ko·en)별 헤더 분기, 본문 미변경 | 동일 |
| AC-HN-004b | 004 | R | `go test ./internal/cli/ -count=1 -run 'TestHandoffShow_LegacyCompatRead' -v` | PASS — legacy pending.json이 있고 DB가 비었을 때 읽기 호환 경로가 그것을 재출력 (호환 분기의 명시적 검증) | 동일 |
| AC-HN-005 | 005 | R | `go test ./internal/cli/ -count=1 -run 'TestNew_SeedsCodexHooksJson' -v` (materializer 소재: `internal/cli/session_worktree.go`, 재심사 실측) | PASS — 새 워크트리에 `.codex/hooks.json` 존재, moai 소유 커맨드 포함 (생성 로직 `internal/codexwiring` wire.go 재사용) | 시딩 코드 부재로 신규 테스트 FAIL |
| AC-HN-006 | 006 | R | `go test ./internal/cli/ -count=1 -run 'TestEnterWorktree_SeedsMissingCodexHooks' -v` | PASS — 부재 트리 진입 시 채워짐, 존재 트리 진입 시 바이트 불변 | 동일 |
| AC-HN-007 | 007 | R | `go test ./internal/cli/ -count=1 -run 'TestNew_SeedFailureFailOpen' -v` | PASS — 시딩 강제 실패(경로 오염)에도 워크트리 생성 성공 + stderr 진단 | 동일 |
| AC-HN-008 | 008 | R (조건부) | `go test ./internal/codexadapter/ -count=1 -run 'TestMapOutput_SessionStartAdditionalContext' -v` | 관문 (b) 통과 시에만 테스트가 작성되고 PASS | 관문 (b) 기각 시 본 AC는 REQ-HN-008과 함께 기각 종결 (기각 사유 기록, 테스트 미작성) |
| AC-HN-009 | 009 | G | (LIVE 절차 — design §C) 관문당 관측 출력을 `.moai/reports/t1273/live-*.txt` | 관문 (a) moai 훅 발화 관측 / (b) 마커 도달 또는 부재 / (c) 길이·강등 여부 — 각각 관측 기록 | 쿼터 회복(9/28 14:37) 전 실행 불가 — 외부 의존이므로 regression-guard |
| AC-HN-010 | 010 | R (보존) | `go test ./internal/cli/ ./internal/hook/ ./internal/codexadapter/ -count=1 -run 'TestHandoffSave_(WritesJSONNotMarkdown\|Schema\|Stdin\|RequiresBody)\|TestHandoffClear\|TestHandoffCmdRegistered\|TestRenderHandoffContext\|TestMapOutput' -v` | 열거한 **실존** 테스트 전부 PASS — M1이 save 표면·렌더 형식·어댑터 매핑을 건드리지 않았음. 주석(재심사 R1): `TestMapOutput` 패턴은 관문 (b) 채택 전까지 이 트리에서 0건 스윕이다 — 기각 분기에서는 이 패턴을 제외한 나머지가 "전부"의 판정 집합이고, 채택 분기에서는 AC-HN-008 테스트가 이를 채운다 | red 없음(보존 AC) — 구현 후에도 변함없이 통과함이 green 판정 |
| AC-HN-011 | 011 | R | `grep -c "moai handoff save" .moai/specs/SPEC-HANDOFF-NEUTRAL-001/design.md` | ≥ 1 — 방향 중립(Codex→Claude save 경로) 문서화가 design §D3에 존재 | red 없음(문서 AC) — 본 문서가 이미 조건을 충족하면 run-phase에서 위반 시에만 red |
| AC-HN-012 | (형식) | R | `go vet ./internal/cli/ ./internal/codexadapter/ ./internal/homestate/ && gofmt -l internal/cli/handoff.go internal/codexadapter/output.go` | vet exit 0, gofmt 빈 목록 | red 없음(형식 게이트) |

## 채택 순서 계약

1. run-phase 첫 액트: AC-HN-001..007(+4b) 테스트 파일을 먼저 작성해 **RED 출력을 관측·기록** (`.moai/reports/t1273/red-*.txt`, 파일 첫 줄 = 관측 시점 HEAD SHA).
2. 구현 최소 커밋 단위: ① homestate 소비이력 조회 + show(+테스트) → ② materializer 시딩+런처 보완(`internal/codexwiring` 재사용,+테스트) → ③ (관문 b 통과 시) 어댑터 매핑(+테스트).
3. AC-HN-009 LIVE는 쿼터 회복 후 격리 실행 — M1 종결 조건이 아니라 판정 기록 의무(P1 채택/기각 확정)로 바인딩.
4. GREEN 일괄 관측 후 AC 표에 각 결과(명령+출력 발췌)를 회기한다.
