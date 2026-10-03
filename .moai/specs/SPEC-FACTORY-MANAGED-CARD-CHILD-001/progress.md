# progress.md — SPEC-FACTORY-MANAGED-CARD-CHILD-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-plan-audit-iter1 (0.2.0, 재감사 대기)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 14) · plan.md · acceptance.md (AC 13) · design.md (D-1..D-6) · decision-index.md · progress.md
tier: M
measured_tree: 2b9e4a4d0
open_clarifications: 0 (Q1-Q3 은 리더가 기본값으로 결정 — decision-index.md; mission contract 11c79e1a)

### 감사 이력과 처분

| 회차 | 판정 | 점수 | audited_sha | 처분 |
|---|---|---|---|---|
| 1 | FAIL | 0.81 | `d26215f25` | MP-7(D1)·MP-6(D2) 차단, 주요 D3–D6, 경미 D7–D12. 개정 0.2.0: D1 결정 기록, D2 `cross-platform exemption` 문장, D3 AC-CC-013 정확히 88, D4 AC-CC-002 `*exec.Cmd` 전체·stdio·펌프 미시작, D5 어댑터 재설계(줄 단위·종료 토큰 인식)와 REQ-CC-010 재서술, D6 어댑터를 레인 전용 이음새로 한정, D7 R-1 관측 RED(정리 접두 한 호출), D8 store.go 인용 정정, D9 REQ-CC-007 분리(007/013/014), D10 REQ-CC-004/005 구현 용어 축소, D11 §F 문구, D12 운영 모드 명시 |

## Residual-risk (plan 단계)

- **무인 레인은 첫 카드 이후로 진행하지 않는다(Q1).** 관리 카드 자식은 우선 턴 뒤 유휴이고 카드 시작 프롬프트를 주입하지 않으므로 운영자 입력이나 브로커 메시지 없이는 카드 작업을 시작하지 않는다. 이 SPEC 의 이득은 운영자가 붙은 관리 레인에서만 닿는다. 문서 known-limits 불릿(AC-CC-012)에 같은 문장을 적는다.
- **연속 시작 실패 상한 없음(Q3).** 관리 시작이 계속 실패하면(예: `app-server` 하위 명령 부재, 핸드셰이크 시간 초과) 루프가 큐의 카드를 연속으로 lease 하며 소진할 수 있다. 카드는 lease 만료까지 묶인다. 현행 직접 문 루프도 상한이 없다.
- **stdin EOF 레인(Q2)**: 세션이 `/exit`·`/quit`·치명 오류 없이는 끝나지 않는다.
- **입력 보장 범위**: 종료 토큰 `/exit`·`/quit` 이후 입력만 보장한다. 종료 토큰 없이 끝난 세션이 이미 받은 줄은 되찾지 않는다(design.md D-6 잔여). 어댑터는 드라이버의 토큰 두 값을 복제한다.
- **미관측 전제**: P-1(POSIX 직접 문의 `syscall.Exec` 교체)·P-2(임대 루트와 브로커 루트의 동일성)·P-4(개발자 지침 쌍의 스레드 적용)는 읽기만 했거나 관측 불가다. M0 에서 측정한다.
- **도구 출처**: plan-audit 보고서가 설치된 `moai` 빌드(`0732cc699`)의 지연을 기록했다. 이 plan 의 `moai spec lint` 결과도 같은 빌드에서 얻었다.

## 후속 카드 문안 (리더가 큐에 올릴 준비된 한 단락, 한국어)

Codex 관리 카드 자식에 카드 시작 프롬프트 주입: t1440 으로 `moai codex -f lane` 의 카드별 자식이 옵트인(`MOAI_FACTORY_MANAGED`) 아래 관리 Codex 소유자로 뜨게 됐지만, 관리 소유자의 드라이버는 우선 턴("준비 완료라고 한 줄로 답해, 아직 작업은 시작하지 마") 뒤 유휴로 들어가 운영자 입력이나 브로커 메시지가 올 때까지 카드 작업을 시작하지 않는다. 그래서 운영자가 붙지 않은 무인 레인은 첫 카드 이후로 진행하지 않는다. 레인 루프가 우선 턴 직후 리스한 카드의 id와 시작 지시를 첫 작업 턴으로 주입하는 방법(주입 위치, 부모 REQ-MS-014의 "카드 처분 판정 금지" 경계, 재시작·재배달 시 중복 주입 방지, 헤드리스라 운영자가 보지 못하는 점)을 SPEC으로 정하고 구현하는 카드다. 기본 꺼짐(옵트인)은 유지하고, t1408(TUI 부착)·t1459(시그널·Start/Close) 범위는 건드리지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
