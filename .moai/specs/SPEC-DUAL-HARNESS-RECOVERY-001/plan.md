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

- 카드 t1100, 워크트리 `.claude/worktrees/t1100-recovery`, 브랜치 `WT-dual-harness-recovery`. plan 초안 기준 HEAD `d87e9af2e`, 초안 SPEC 커밋 `09e24fd08`, iter-2 개정은 plan-audit iter-1(`.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-1.md`, FAIL 0.76) 수리였고(커밋 `2bac34f3e`), 이번 iter-3 개정은 plan-audit iter-2(`.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-2.md`, FAIL 0.81) 결함 ND1~ND14 수리다. 처리표는 `progress.md` "Revision iter-3".
- 설계 원문은 제안 문서다(`reports/moai-dual-harness-full-design-20260922.md`). 구현 증거로 쓰지 않는다.
- 개발 방식: 새 동작은 TDD(먼저 실패하는 테스트), 기존 동작 보존은 특성 테스트. 변경 패키지 단위로만 로컬 검증하고 전체 판정은 CI 몫이다.

## §B 결정 기록

초안의 미결 항목 넷과 plan-audit iter-1의 D6은 모두 결정되었다. 이 SPEC에 남은 미결 표식은 없다.

**출처와 한계.** 아래 다섯 결정은 모두 칸반 리드가 2026-09-23 운영자의 상설 위임 아래에서 Jev 답을 근거로 내린 것이다. 건별로 운영자에게 확인받은 결정이 아니다. 기록: `.moai/reports/t1100/operator-decisions.md`. 이 SPEC 작성 세션은 위임 사실 자체를 직접 확인하지 못했다. 같은 기록에 따르면 `CLAUDE.local.md` §29가 사용자 표면 동작 변경을 Jev에 맡기지 않는 등급으로 두므로, 사용자에게 보이는 동작을 바꾸는 결정 1과 D6에 대해 상설 위임이 §29보다 우선하는지는 리드가 운영자에게 올린 "미해결 긴장"으로 남아 있다. Implementation Kickoff Approval은 여전히 운영자 게이트다.

| # | 항목 | 결정 | SPEC 반영 |
|---|---|---|---|
| 1 | unwire-trigger | `moai update`는 고아 배선을 보고만 한다. 제거는 새 명령 `moai tool disable codex`로만 한다 | REQ-DHR-005, 007, AC-DHR-003, 005 |
| 2 | codex-kanban-roles | `moai codex -k`가 lead와 companion을 모두 지원한다 | REQ-DHR-012, AC-DHR-009 |
| 3 | idempotency-scope | 결정 보류, 실측 먼저. run 첫 마일스톤에서 재현 테스트로 "송신 세션 재시작 뒤 같은 키 재전송이 실제로 두 번 적용되는가"를 잰다. 재현되면 송신 lane slot 기준으로 옮기고, 아니면 현행 `UNIQUE(sender_session, idem_key)`를 유지한다. 옮기게 되면 t1082 영향은 리드가 조정한다 | REQ-DHR-017(조건부), 025, AC-DHR-014(분기), 020 |
| 4 | live-budget | 승인. AC-DHR-018 조합당 모델 호출 8회·900초, AC-DHR-012 codex 호출 14회. 예산을 넘으면 중단하고 `ABORTED`로 보고하며 PASS로 세지 않는다 | REQ-DHR-014, 023, AC-DHR-012, 018, 023 |
| D6 | audit-role-sandbox | (a) Codex에서 `plan-auditor`·`sync-auditor`를 read-only로 두고, 부모 lane 오케스트레이터가 감사자 반환문 그대로 판정 파일을 쓴다. "감사자가 자기 판정 파일을 쓴다"는 계약 변경은 Codex 경로에만 적용하는 예외이며 Claude 경로는 그대로다. 리드가 제시한 Jev 수치(opt_a 0.79, opt_b 0.10, opt_c 0.19, reversible 0.52)는 기록 파일 인용이며 이 세션이 원본을 보지 않았다 | REQ-DHR-015, AC-DHR-013, 023 |

결정 3의 측정 기준(REQ-DHR-025): 바뀌지 않은 트리에는 결과 적용 단계가 없으므로, "두 번째 메시지 행이 생기고 수신자가 서로 다른 메시지 ID 두 개를 claim한다"를 "두 번 적용될 수 있음"의 관측 기준으로 쓴다. 코드 판독(`store.go:282`, `:624`)으로는 재현이 예상되지만 측정 전까지는 가설이다. 한편 REQ-DHR-018의 결과 적용은 attempt 단위라서, 어느 분기든 dispatch 결과는 한 번만 반영된다. 멱등 범위가 바꾸는 것은 수신자가 같은 결과를 두 번 받느냐뿐이다. 리드에게 알릴 사실이다.

D6 결정의 설계 §10 대조: 원문(`reports/moai-dual-harness-full-design-20260922.md:177`)은 "역할별 권한을 호스트가 표현하지 못하면 더 넓은 권한을 조용히 부여하지 않는다. 외부 worker의 sandbox로 강제할 수 있는지 먼저 검증하고 불가능하면 해당 역할을 차단한다"이다. (a)는 쓰기 경계를 sandbox(read-only)로 강제하므로 차단 분기로 가지 않는다. read-only 강제가 실제로 동작하는지는 AC-DHR-012(LIVE)가 잰다. 그 전까지 강제는 가설이며, AC-DHR-012가 `NOT_RUN`이면 AC-AGENT-01은 `PARTIAL`이다.

## §C 사전 점검 (run 착수 시)

- `git rev-parse --short HEAD`, `git branch --show-current`로 트리를 재확인한다(기대: `WT-dual-harness-recovery`).
- 로컬 develop과의 차이를 재측정하고, t1082 워크트리에서 factorymsg가 바뀌었는지 확인한다(같은 `store.go`를 만진다).
- 기준 테스트: `go test ./internal/codexwiring ./internal/factorymsg ./internal/template/agentemit -count=1` (plan 시점 결과는 `progress.md` §E.1).
- M1의 재현 측정(AC-DHR-020)은 factorymsg 스키마를 건드리기 전에 실행한다.

## §D 제약

- 새 저장소·데몬·브로커를 만들지 않는다. 소유 기록은 `internal/manifest`, 작업 레코드는 factorymsg SQLite, 배선 잠금은 `.moai/state/` 아래 파일 하나.
- `.codex/agents/moai/*.toml`은 손으로 고치지 않는다. `agents-codex.yaml`이나 `internal/template/templates/.claude/agents/moai/*.md`를 고친 뒤 `make agents-emit`.
- D6 예외는 Codex 경로에만 둔다. `.claude/agents/moai/plan-auditor.md`, `sync-auditor.md`(C1·C2)와 Claude 감사 워크플로는 바꾸지 않는다.
- 템플릿을 고치면 `make build`. 템플릿 본문에 SPEC ID·카드 ID·날짜를 넣지 않는다(템플릿 중립성).
- 브로커 전달 의미는 at-least-once로 유지한다(REQ-FMH-006).
- 공유 기본 체크아웃과 다른 워크트리(`dual-harness-parity-rebuild`, `t1100`, `t1082`)에 쓰지 않는다.
- 검증 명령에는 실행 중에 계산한 값을 git·go 명령으로 넘기지 않는다(워크트리 세션 가드).

## §E 자기 검증

run 완료 보고는 `acceptance.md` §C 표를 채운다. 각 행은 명령, 판정식 출력(`true`/`false`), 증거 파일 경로를 가진다. LIVE는 실행했으면 경로, 아니면 `NOT_RUN`, 예산으로 멈췄으면 `ABORTED`. `PARTIAL`은 PASS로 적지 않는다. AC-DHR-014는 AC-DHR-020 측정 결과에 맞는 분기 명령만 판정하고 다른 분기는 `N/A (branch)`로 적는다.

## §F 마일스톤

바뀔 가능성이 높은 결정부터 둔다. 측정이 결정을 가르는 항목이 맨 앞, 데이터 모델과 사용자 동선이 그다음, 기계적 테스트 확장이 뒤다.

### M1 — 멱등 범위 재현 측정 (Priority High)

가장 바뀌기 쉬운 결정: 멱등 범위의 분기(§B-3). 측정 결과가 M2의 스키마를 정한다.

- REQ-DHR-025, AC-DHR-020
- 파일: `internal/factorymsg/` 새 테스트 파일 하나. 스키마·프로덕션 코드 변경 없음
- 순서: `git rev-parse --short HEAD`를 `ac020-head.txt`에 기록 → AC-DHR-020 실행 → 결과를 `progress.md` §E.2에 적고 분기 A/B를 확정
- 결과가 `reproduced`이면 리드에게 t1082 조정이 필요함을 알린다(§B-3).

### M2 — dispatch record, 결과 적용 판정 순서, 조건부 범위 이관 (Priority High)

두 번째로 바뀌기 쉬운 결정: 새 테이블 모양, 상태 전이, 판정 순서 표, 같은 attempt 재부여 연산.

- REQ-DHR-016 ~ 021, AC-DHR-014 ~ 016
- 파일: `internal/factorymsg/store.go`(테이블, 적용·재할당·재부여 API, 분기 A일 때만 마이그레이션), 테스트 파일, `internal/cli/mcp_factory_msg.go`(도구 표면이 필요한 경우만), `internal/mcp/catalog.go`(도구를 추가할 때만)
- 순서: dispatch 테이블 → 판정 순서 표(`design.md` §D.3) → 재할당 → superseded status와 재부여 → (분기 A만) 마이그레이션 특성 테스트 → 범위 이관
- t1082와 같은 파일을 만진다. 병합 순서가 뒤바뀌면 t1082가 이 스키마 위에 올라가도록 리드가 조정한다.

### M3 — 배선 소유 기록, 저널, 배선 잠금 (Priority High)

세 번째로 바뀌기 쉬운 결정: manifest 부분 기록 모양, origin 규칙, 저널 형식, 잠금 파일 위치.

- REQ-DHR-001 ~ 004, 006, AC-DHR-001, 002, 021, 022
- 파일: `internal/manifest/types.go`, `internal/manifest/manifest.go`, `internal/codexwiring/wire.go`, `hooks.go`, `configtoml.go`, 새 저널·잠금 파일(패키지 내부), `internal/cli/update_codex_wiring.go`(lock-held 보고), `internal/cli/doctor_codex.go`(읽기 전용 보고), 테스트 파일들
- 기존 동작 보존: `wire_test.go`, `sidecar_test.go`, `hooks_test.go`, `configtoml_test.go`, `statusline_test.go` 전부 통과 유지(REQ-CW-005/006 멱등·보존 계약).

### M4 — `moai tool disable codex`와 프로필 전환 보고 (Priority High)

사용자 동선: 새 명령 이름과 보고 문구(§B-1).

- REQ-DHR-005, 007, AC-DHR-003 ~ 005
- 파일: `internal/cli/tool.go`(disable 하위 명령), `internal/codexwiring/`(unwire 본체), `internal/cli/update_codex_wiring.go`, `internal/cli/update_template_sync.go`(고아 보고)
- 범위 경계: update의 관리 경로 정리 단계(`CleanMoaiManagedPaths`)가 프로필과 무관하게 `.claude/` 관리 뿌리를 지우는 동작은 바꾸지 않는다(`spec.md` §F, `design.md` §A.6). M4의 테스트는 `.codex/` 쪽과 배선 파일만 단언한다.

### M5 — Codex worktree anchor, 폐기 동등성, `-k` (Priority High)

사용자 동선: `-k` 역할(§B-2), 거부 진단 문구.

- REQ-DHR-008 ~ 012, AC-DHR-006 ~ 009
- 파일: `internal/cli/codex_launcher.go`, `codex_direct_posix.go`, `codex_direct_windows.go`, `session_worktree.go`(`moai cc -w` 사전 판정), `kanban.go` / `factory.go`(파서 재사용 지점), `internal/session/anchor_lock.go`(수정 없이 재사용이 목표), `internal/cli/worktree/remove.go`(lock-aware anchor 판정), `clean.go`·`session_worktree_prmerge.go`(AC-DHR-008에서 빈틈이 드러날 때만), `done.go`는 바꾸지 않는다(L1 거부 유지, SPEC-WORKTREE-DONE-TIER-001)

### M6 — 역할 권한 계약과 감사 역할 Codex 예외 (Priority Medium)

- REQ-DHR-013 ~ 015, AC-DHR-010, 011, 013
- 파일: `internal/template/agentemit/agents-codex.yaml`(`sandbox_mode.role_values`에 두 감사 역할 `read-only`, `read-vs-write-distinction` 근거 수정, 축 계약), `manifest.go`, `writer.go`, 테스트, Codex 부모 지시 표면(템플릿 `AGENTS.md.tmpl` capability 표 또는 Codex 전용 발행물) → `make agents-emit`, `make build`

### M7 — 결정적 4조합 카드 흐름과 판정식 (Priority Medium)

- REQ-DHR-022, 024, AC-DHR-017, 019
- 파일: `internal/factorymsg/` 새 테스트 파일

### M8 — LIVE 증거 (Priority Low)

- REQ-DHR-014(LIVE 부분), REQ-DHR-015(LIVE 부분), REQ-DHR-023, AC-DHR-012, 018, 023
- 파일: `internal/cli/factory_live_test.go`(카드 흐름 케이스 추가, 기존 왕복 케이스 유지), 새 `internal/cli/codex_role_live_test.go`
- 실행 조건: 격리된 임시 저장소·`MOAI_HOME`·`CODEX_HOME`, §B-4 예산(AC-DHR-012·023 합쳐 14회, AC-DHR-018 조합당 8회·900초), 예산 초과 시(예산 다음 호출이 필요해지거나 900초를 넘으면) 그 호출 전에 중단·`ABORTED`(예산과 정확히 같은 호출 수로 끝난 실행은 `ABORTED`가 아님), 첫 spawn 전 정리 등록, 조합마다 따로 실행.

## §G 위험

| 위험 | 대응 |
|---|---|
| 범위 이관(분기 A)이 t1074 기존 DB와 t1082 문구를 깬다 | M1 측정 후에만 착수, 마이그레이션 특성 테스트 선행, 리드가 t1082와 조정 |
| 구버전 설치의 `[mcp_servers.moai]`·`status_line`·`description`은 `unknown`이 되어 unwire가 지우지 못한다. 다시 배선해도 바뀌지 않는다 | 의도한 보수 기본값. doctor가 해당 부분과 손 제거 방법을 안내. handler는 명령 식별자로 분류되므로 다음 배선 한 번이면 제거 가능 |
| manifest 손상·초기화 뒤 이미 `created`였던 config 부분이 `unknown`이 된다 | 제거 가능성만 잃고 사용자 부분 오분류는 없음. 보고로 드러냄 |
| exec 후 lock 해제 주체가 없음 | 죽은 pid lock은 anchor가 아니므로 무해. 다음 launch가 가드 아래서 교체 |
| Windows에서 부모 moai만 강제 종료되면 codex 자식이 살아 있어도 lock이 죽은 것으로 보인다 | 잔여 위험으로 명시. Windows 실측은 범위 밖 |
| `remove`는 지금 레지스트리만 본다. `done`은 L1 트리를 늘 거부한다 | M5에서 `remove`에 lock-aware 판정 추가. `done`은 바꾸지 않음. AC-DHR-008이 측정 |
| gpt 프로필 update가 `.claude/` 관리 뿌리를 지운 뒤 다시 배포하지 않을 수 있다(코드 판독, 미측정) | 이 카드 범위 밖. `spec.md` §F에 후속 카드 후보로 기록 |
| `moai cc -w` 사전 판정이 Claude Code 자체 lock과 겹친다 | 사전 판정은 읽기만 하고 lock을 쓰지 않는다 |
| 부모 오케스트레이터가 감사 반환문을 바꿔 적을 수 있다 | AC-DHR-023(LIVE)이 세션 기록과 판정 파일을 해시로 대조. 세션 기록에 반환문이 없으면 `NOT_RUN` |
| 결정 출처가 건별 운영자 확인이 아니다(§B) | 기록에 한계를 명시. Kickoff 게이트는 운영자 |
| LIVE 테스트 비용과 flake | 예산 승인제, 예산 도달 시 `ABORTED`, 조합별 분리 실행, 실패·미실행·중단을 분리 집계 |

## §H 참조

- `spec.md` §E(t1082 경계), `design.md`, `research.md`, `acceptance.md`
- `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-1.md`, `-review-2.md`, `.moai/reports/t1100/operator-decisions.md`
- SPEC-WORKTREE-DONE-TIER-001(`moai worktree done`의 L1 거부, completed)
- SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-006(at-least-once), SPEC-CODEX-WIRING-001 REQ-CW-005/012
