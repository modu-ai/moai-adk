---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: plan
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Plan — SPEC-DUAL-HARNESS-RECOVERY-001

## §A 맥락

- 카드 t1100, 워크트리 `.claude/worktrees/t1100-recovery`, 브랜치 `WT-dual-harness-recovery`, plan 기준 HEAD `d87e9af2e`.
- 설계 원문은 제안 문서다(`reports/moai-dual-harness-full-design-20260922.md`). 구현 증거로 쓰지 않는다.
- 개발 방식: 새 동작은 TDD(먼저 실패하는 테스트), 기존 동작 보존은 특성 테스트. 변경 패키지 단위로만 로컬 검증하고 전체 판정은 CI 몫이다.

## §B 운영자 결정이 필요한 항목

run 착수 승인 전에 정해야 한다. 각 항목은 기본 제안을 달았다.

1. [NEEDS CLARIFICATION: unwire-trigger] 하네스 프로필이 `claude`로 바뀐 뒤 `moai update`가 Codex 배선을 자동으로 unwire할지, 고아로 보고만 하고 제거는 명시 명령(`moai tool disable codex`, 기존 `moai tool enable codex`의 짝)으로만 할지. 기본 제안: 보고만 하고 명시 명령으로 제거. 사용자 설정 파일을 update가 스스로 고치는 경로를 늘리지 않는다.
2. [NEEDS CLARIFICATION: codex-kanban-roles] `moai codex -k`가 lead와 companion을 모두 지원할지, companion만 지원할지. 기본 제안: 둘 다. `moai codex -f`가 이미 codex lead를 허용한 선례(카드 t865)가 있다. 다만 Codex lead에는 Claude의 `SendMessage` 알림 경로가 없어 알림은 큐와 브로커에만 의존한다.
3. [NEEDS CLARIFICATION: idempotency-scope] factorymsg 멱등 범위를 송신자 세션 UUID에서 송신 lane slot으로 옮길지(REQ-DHR-017). t1074가 착지시킨 스키마를 바꾸고, 초안 단계인 t1082(design.md §8 "idempotency key에 handoff generation 결합")와 문구를 맞춰야 한다. 기본 제안: 옮긴다. 송신자 재시작 뒤 재전송이 중복 적용되는 경로가 코드상 열려 있기 때문이다. 채택하지 않으면 REQ-DHR-017과 AC-DHR-014 (c)를 빼고 AC-MSG-01의 "재시작" 행을 `UNPROVEN`으로 남긴다.
4. [NEEDS CLARIFICATION: live-budget] LIVE 두 항목(AC-DHR-012 역할 로드, AC-DHR-018 4조합 카드 흐름)의 모델 호출 예산 승인. 기본 제안: AC-DHR-018 조합당 모델 호출 8회 이하·조합당 제한 시간 900초, AC-DHR-012 codex 호출 14회 이하(역할 로드 12 + 쓰기 시도 2). 승인 전에는 두 항목을 실행하지 않고 `NOT_RUN`으로 둔다.

## §C 사전 점검 (run 착수 시)

- `git rev-parse --short HEAD`, `git branch --show-current`로 트리를 재확인한다(기대: `WT-dual-harness-recovery`).
- 로컬 develop과의 차이를 재측정하고, t1082 워크트리에서 factorymsg가 바뀌었는지 확인한다(같은 `store.go`를 만진다).
- 기준 테스트: `go test ./internal/codexwiring ./internal/factorymsg ./internal/template/agentemit -count=1` (plan 시점 결과는 `progress.md` §E.1).

## §D 제약

- 새 저장소·데몬·브로커를 만들지 않는다. 소유 기록은 `internal/manifest`, 작업 레코드는 factorymsg SQLite.
- `.codex/agents/moai/*.toml`은 손으로 고치지 않는다. `internal/template/templates/.claude/agents/moai/*.md`나 `agents-codex.yaml`을 고친 뒤 `make agents-emit`.
- 템플릿을 고치면 `make build`. 템플릿 본문에 SPEC ID·카드 ID·날짜를 넣지 않는다(템플릿 중립성).
- 브로커 전달 의미는 at-least-once로 유지한다(REQ-FMH-006).
- 공유 기본 체크아웃과 다른 워크트리(`dual-harness-parity-rebuild`, `t1100`, `t1082`)에 쓰지 않는다.

## §E 자기 검증

run 완료 보고는 `acceptance.md` §C 표를 채운다. 각 행은 명령, 판정식 출력(`true`/`false`), 증거 파일 경로를 가진다. LIVE는 실행했으면 경로, 아니면 `NOT_RUN`. `PARTIAL`은 PASS로 적지 않는다.

## §F 마일스톤

바뀔 가능성이 높은 결정부터 둔다. 데이터 모델과 사용자 동선이 앞, 기계적 테스트 확장이 뒤다.

### M1 — dispatch record와 멱등 범위 (Priority High)

가장 바뀌기 쉬운 결정: 새 테이블 모양, 상태 전이, 멱등 범위 이동(§B-3).

- REQ-DHR-016 ~ 021, AC-DHR-014 ~ 016
- 파일: `internal/factorymsg/store.go`(테이블, 마이그레이션, 적용·재할당 API), `internal/factorymsg/store_test.go` 또는 새 `dispatch_test.go`, `internal/cli/mcp_factory_msg.go`(도구 표면이 필요한 경우만), `internal/mcp/catalog.go`(도구를 추가할 때만)
- 순서: 마이그레이션 특성 테스트(기존 DB 행 보존) → 멱등 범위 이동 → dispatch 테이블 → fenced 적용 → 재할당 → superseded status
- t1082와 같은 파일을 만진다. 병합 순서가 뒤바뀌면 t1082가 이 스키마 위에 올라가도록 조정한다.

### M2 — 배선 소유 기록과 저널 (Priority High)

두 번째로 바뀌기 쉬운 결정: manifest 부분 기록 모양, 저널 형식.

- REQ-DHR-001 ~ 004, AC-DHR-001, 002
- 파일: `internal/manifest/types.go`, `internal/manifest/manifest.go`, `internal/codexwiring/wire.go`, `internal/codexwiring/hooks.go`, `internal/codexwiring/configtoml.go`, 새 저널 파일(패키지 내부), 테스트 파일들
- 기존 동작 보존: `wire_test.go`, `sidecar_test.go`, `hooks_test.go`, `configtoml_test.go`, `statusline_test.go` 전부 통과 유지(REQ-CW-005/006 멱등·보존 계약).

### M3 — unwire 동선과 프로필 전환 (Priority High)

사용자 동선: 명령 이름, 자동/수동 제거(§B-1).

- REQ-DHR-005 ~ 007, AC-DHR-003 ~ 005
- 파일: `internal/cli/tool.go`(disable 하위 명령), `internal/cli/update_codex_wiring.go`, `internal/codexwiring/`(unwire 본체), `internal/cli/update_template_sync.go`, `internal/cli/update_cleanup.go`(분류 재사용), `internal/cli/doctor_codex.go`(고아·복구 보고)

### M4 — Codex worktree anchor와 `-k` (Priority High)

사용자 동선: `-k` 역할 범위(§B-2), 거부 진단 문구.

- REQ-DHR-008 ~ 012, AC-DHR-006 ~ 009
- 파일: `internal/cli/codex_launcher.go`, `internal/cli/codex_direct_posix.go`, `internal/cli/codex_direct_windows.go`, `internal/cli/session_worktree.go`(`moai cc -w` 사전 판정), `internal/cli/kanban.go` / `factory.go`(파서 재사용 지점), `internal/session/anchor_lock.go`(수정 없이 재사용이 목표), `internal/cli/worktree/clean.go`·`done.go`(AC-DHR-008에서 빈틈이 드러날 때만)

### M5 — 역할 권한 계약과 감사 경계 (Priority Medium)

- REQ-DHR-013 ~ 015, AC-DHR-010, 011, 013
- 파일: `internal/template/agentemit/agents-codex.yaml`, `manifest.go`, `writer.go`, 테스트, `internal/cli/worktree/guard.go`(허용 목록), `internal/worktree/state_guard.go`, 감사 호출 단계가 있는 워크플로 본문(`internal/template/templates/.claude/skills/moai/workflows/plan/`, `.../sync/` 하위 중 감사 spawn 지점) → `make agents-emit`, `make build`

### M6 — 결정적 4조합 카드 흐름과 판정식 (Priority Medium)

- REQ-DHR-022, 024, AC-DHR-017, 019
- 파일: `internal/factorymsg/` 새 테스트 파일

### M7 — LIVE 증거 (Priority Low, §B-4 승인 후)

- REQ-DHR-014(LIVE 부분), REQ-DHR-023, AC-DHR-012, 018
- 파일: `internal/cli/factory_live_test.go`(카드 흐름 케이스 추가, 기존 왕복 케이스 유지), 새 `internal/cli/codex_role_live_test.go`
- 실행 조건: 격리된 임시 저장소·`MOAI_HOME`·`CODEX_HOME`, 조합별 예산·시간 제한, 첫 spawn 전 정리 등록. 기존 `requireFactoryLive`처럼 조합마다 따로 실행한다.

## §G 위험

| 위험 | 대응 |
|---|---|
| 멱등 범위 이동이 t1074 기존 DB와 t1082 설계를 깬다 | 마이그레이션 특성 테스트 선행, §B-3 결정 전 착수하지 않음, t1082 담당에 경계 표 전달 |
| 구버전 배선에는 provenance가 없어 unwire가 아무것도 못 지운다 | 의도한 보수 기본값. 한 번 재배선 후 제거 가능하다는 안내를 doctor에 출력 |
| exec 후 lock 해제 주체가 없음 | 죽은 pid lock은 anchor가 아니므로 무해. 다음 launch가 교체 |
| `moai cc -w` 사전 판정이 Claude Code 자체 lock과 겹친다 | 사전 판정은 읽기만 하고 lock을 쓰지 않는다 |
| 감사 경계 호출을 Codex가 보장하지 않음 | 잔여 위험으로 명시. 훅 강제는 M2(범위 밖) |
| LIVE 테스트 비용과 flake | 예산 승인제, 조합별 분리 실행, 실패와 미실행을 분리 집계 |

## §H 참조

- `spec.md` §E(t1082 경계), `design.md`, `research.md`, `acceptance.md`
- SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-006(at-least-once), SPEC-CODEX-WIRING-001 REQ-CW-005/012
