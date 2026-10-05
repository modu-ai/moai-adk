---
id: SPEC-CODEX-EVENT-COVERAGE-001
title: "plan — Codex hook 이벤트 커버리지 (Interrupt 행 + 발화 실측 캠페인)"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
---

# plan — SPEC-CODEX-EVENT-COVERAGE-001

## A. 컨텍스트

- 카드 t496 (Class C) · 트리 `.claude/worktrees/t496` · development_mode: tdd (`quality.yaml:2`)
- 대상: `internal/codexadapter/events.go` (테이블+Resolve), 테스트 파일 3종(codexadapter 2개 갱신 + codexwiring/hooks_test.go 신설 테스트 — M1 파일 목록 참조), 그리고 M2 캠페인(코드 무변경, 런 페이즈 실행)
- 본 트리 재측정 완료(2026-09-07): `EventInterrupt` grep 0히트(대조군 EventStop 검출), EventTable 6+5=11행, `internal/cli/hook.go`에 interrupt 서브커맨드 0히트, `codexwiring/hooks.go:99-100`의 `!row.Adapted` skip 확인, `codex --version` → `codex-cli 0.153.4`

## B. 알려진 이슈

- `TestDispatcherArgsExist` (`dispatcher_registration_test.go:20-45`)는 **Adapted=false 행 포함 전 행**이 `internal/cli/hook.go`에 등록된 디스패처 인자를 가져야 한다고 단언한다. 빈 `DispatcherArg` 행을 추가하면 이 테스트가 깨진다 — M1이 반드시 함께 고치는 계약된 갱신점이다(§D1).
- `events.go:86`의 ErrUnadapted 메시지는 `(dispatcher arg %q exists; ...)`라고 단언한다 — 빈 인자에 대해 거짓이 된다(§D1).
- `events_test.go:15` `TestEventTableRowCount`는 `wantRows = 11`로 고정 — 12로 갱신 필요.
- `internal/hook/types.go:249`의 `IsInterrupt bool`은 페이로드 구조체 필드로, EventType 상수와 무관(혼동 금지).

## C. 사전 점검 (Pre-flight)

- [x] SPEC ID 정규식 Bash 검사: `SPEC-CODEX-EVENT-COVERAGE-001` → PASS (2026-09-07 실행)
- [x] ID 중복: `.moai/specs/`에 `SPEC-CODEX-EVENT-COVERAGE*` 선행 존재 없음 (grep, 출력 0행)
- [x] frontmatter 12 필드 스키마 적합 (schema SSOT § Canonical 12 대조)
- [x] GEARS 표기 / Out of Scope H3+불릿 규약

## D. 설계 결정 (Class C — 명시적 결정 기록)

### D1. EventTable 행 형태 — 빈 DispatcherArg + ErrUnadapted 의도적 재사용 (채택)

**채택**: `{CodexEventInterrupt, "", false}` 행을 추가한다. `Resolve`는 ErrUnadapted를 그대로 감싸되, 메시지의 dispatcher-arg 존재 단언을 빈 인자에서 생략한다. `TestDispatcherArgsExist`는 빈 인자 행을 skip하도록 갱신한다(주석에 이유 기재 — "Interrupt has no MoAI dispatcher counterpart; an empty arg is the marker").

**기각 1 — `moai hook interrupt` 서브커맨드 등록**: Claude 측에 Interrupt에 대응하는 훅이 없어 등록할 핸들러 동작이 없다. 테스트(`TestDispatcherArgsExist`)를 통과시키려고 죽은 경로를 만드는 것은 방향이 틀렸고, "테이블의 모든 행이 등록 인자를 갖는다"는 기존 단언의 취지(오탐 방지)도 훼손한다.

**기각 2 — 3번째 센티널 `ErrNoCounterpart`**: "known but no MoAI counterpart"와 "recognized but not adapted"를 구별하는 소비자가 현재 0명이다(§E 전수 조사 — `IsUnadapted`/`IsUnknownEvent`의 production 호출부 중 refusal 클래스를 분기하는 곳은 0이고, 유일한 외부 호출부는 테스트 1건이다; harness 리퓨절 경로는 `%w` 래핑만 하고 분기하지 않는다). 3번째 sentinel+헬퍼는 소비자 없는 API 표면이다. `IsUnadapted(err)`가 true인 상태에서 메시지가 뉘앙스를 운반하는 것으로 충분하며, 소비자가 구별이 필요해지는 시점이 업그레이드 트리거다.

### D2. Interrupt 이름 상수의 위치 — internal/codexadapter (채택)

불변 원문 (`events.go:8-9`): *"this package sits in FRONT of the dispatcher rather than inside it, and nothing under internal/hook is modified (SPEC-CODEX-HOOK-ADAPTER-001 REQ-7)"*.

- **문자 판정**: internal/hook에 additive 상수 하나라도 REQ-7 문자 위반이다 — "nothing ... is modified"에 추가는 포함된다.
- **의도 판정**: internal/hook은 Claude 측 디스패처 이벤트 어휘이다. Interrupt는 Claude 측 대응 훅이 없는 Codex 전용 이벤트라, 그 어휘에 들어가면 어댑터 앞단 배치의 의미(두 하네스스 간 차이를 어댑터가 흡수)가 무너진다.
- **채택**: `const CodexEventInterrupt hook.EventType = "Interrupt"` 를 `events.go`에 둔다. `hook.EventType`이 string 기반이므로 internal/hook 무변경으로 타입 일관성이 유지된다. M1 이후 internal/ 전역 `EventInterrupt` grep은 (테스트 문자열 제외) 0행을 유지하는 것이 회귀 판별식이다.

### D3. EventTable 소비자 전수와 축별 영향 (본 트리에서 검증한 목록)

| 소비자 | 위치 | M1(Axis A) 영향 | M2(Axis B) 영향 |
|---|---|---|---|
| `RenderHooks` (설치기) | `internal/codexwiring/hooks.go:97` | **없음** — `!row.Adapted { continue }` (:99-100)이라 빈 인자 행이 설치 표면에 도달하지 않음. `.codex/hooks.json` 바이트 불변이 회귀 판별식 | 없음 (코드 무변경) |
| 런타임 Resolve | `internal/cli/hook_harness_codex.go:54,58` | 사용자가 손으로 `.codex/hooks.json`에 Interrupt를 넣으면 기존 5종과 같은 ErrUnadapted 리퓨절 클래스로 거부 — 동작 변화 없음, 메시지만 D1에 따라 수정 | 없음 |
| `ValidateConfig` | `internal/cli/doctor_codex.go:132`, `internal/cli/codex_readiness.go:116`, `internal/codexwiring/wire.go:86` | 없음 — `config.go:53-82` 직독상 EventTable을 조회하지 않고 구조만 검증 | 없음 |
| `TestDispatcherArgsExist` | `internal/codexadapter/dispatcher_registration_test.go:31-45` | **갱신 필요** — 빈 인자 행 skip (§D1) | 없음 |
| `TestEventTableRowCount`/`TestEventTableMapping`/Resolve 테스트 | `internal/codexadapter/events_test.go` | **갱신 필요** — wantRows 12, Interrupt 엔트리, Resolve 케이스 | 없음 |
| 디스패처 등록 교차검증 주석 | `internal/cli/hook.go:290` | 없음 (읽기 전용 대조 주석) | 없음 |

### D4. 주석 갱신 (code_comments: en)

- `events.go:40-44` "All eleven Codex events have a counterpart: ..." → 12종 + "Interrupt has no MoAI dispatcher counterpart (empty DispatcherArg)" 로 수정.
- `events.go:46-50` "Six rows are adapted... Four are held back..." → 행 수 상태를 M1/M2 후 참으로. M2가 발화 재판정을 마치면 SubagentStop 문장(측정 "NOT to fire", 0.147.0 basis)을 0.153.4 재측정 결과로 갱신.
- 패키지 doc `events.go:11-12` "Measurement basis: codex-cli 0.147.0" → M2 종료 시 `codex-cli 0.153.4 (campaign record: .moai/reports/t496/codex-event-campaign.md)` 로 재스탬프.

## E. EventTable/Resolve 소비자 검증 명령 (plan 페이즈 실행분)

```bash
grep -rn 'codexadapter\.' internal/ cmd/ pkg/ | grep -v 'internal/codexadapter/' | grep -v '_test.go'
```
→ `cli/codex_readiness.go:116`, `cli/hook.go:290(주석)`, `cli/doctor_codex.go:132`, `cli/hook_harness_codex.go:44,54,58,81,90`, `codexwiring/hooks.go:43,97`, `codexwiring/wire.go:86`. 비테스트 소비자는 위 6파일뿐(테스트 소비자 별도 4파일: hook_harness_codex_test.go, codex_readiness_test.go, codexwiring/hooks_test.go, wire_test.go). 이 목록이 D3 표의 근거다.

## F. 마일스톤

### M1 — Axis A: Interrupt 12번째 행 (High)

결정 역순 위 원칙에 따라, 바뀔 가능성이 가장 높은 결정(D1/D2 — 표 형태·상수 위치)을 앞세우고 기계적 갱신을 뒤로:

1. **표 형태+상수** (§D1/§D2): `events.go`에 `CodexEventInterrupt` 상수 + `{CodexEventInterrupt, "", false}` 행. `Resolve` 메시지 조건 처리.
2. **테스트 RED→GREEN**: `TestEventTableRowCount` 11→12, `TestEventTableMapping` 엔트리, Resolve Interrupt 케이스(ErrUnadapted + 메시지에 "dispatcher arg" 존재 단언 부재), `TestDispatcherArgsExist` 빈 인자 skip.
3. **doc comment 갱신** (§D4 — SubagentStop 문장은 M2 결과 대기 중 M1 시점엔 "measured NOT to fire on 0.147.0; re-verification pending (M2)"로 중간 상태 기재).
4. **회귀 판별식**: `grep -rn 'EventInterrupt' internal/` → 테스트 외 0행. `go test ./internal/codexadapter/`·`go test ./internal/codexwiring/ ./internal/cli/ -run 'Codex|Hooks'` (영향 패키지 한정).

파일: `internal/codexadapter/events.go`, `internal/codexadapter/events_test.go`, `internal/codexadapter/dispatcher_registration_test.go`, `internal/codexwiring/hooks_test.go` (AC-CEV-005의 신설 "Interrupt 미설치" 테스트 소재 — `TestRenderHooks_*` 패밀리가 이 파일에 있다). **프로덕션(비테스트) 파일은 위 events.go 1개뿐이며 그 외 프로덕션 무변경** — 테스트 파일 3개는 신설/갱신된다.

### M2 — Axis B: 발화 실측 캠페인 (High) — 런 페이즈 실행

**P0 전제 (캠페인 최우선 단계)**: `CODEX_HOME` 지원 검증 — 임시 디렉터리를 `CODEX_HOME`으로 두고 hooks.json에 SessionStart 로거를 심은 뒤 비대화형 `codex exec` 1회 실행, 로그 착지를 임시 홈에서 확인. 지원 실패 시 폴백: scratch repo(임시 디렉터리의 git repo)에 project-scope `.codex/hooks.json` 배치 + trust 절차 문서화. 어느 경로든 연산자 실제 `~/.codex` 무변경 (REQ-CEV-008).

**측정 대상 6종과 트리거 방법 (0.153.4 기준, 런 페이즈에서 실행·확인)**:

| 이벤트 | 트리거 방법 (안) |
|---|---|
| PreCompact / PostCompact | 컴팩션이 발화하도록 충분히 긴 컨텍스트로 `codex exec` 실행 — **[미검증]** 이 트리거의 비대화형 실현 가능성은 plan 페이즈에서 입증되지 않았다. 런 페이즈에서 확립하거나, 확립 실패 시 해당 이벤트는 재분류 대상이다 (아래 not-fired 판정 규칙 참조) |
| PermissionRequest | approval-mode 설정 상태에서 승인 요구 명령 실행 |
| SubagentStart / SubagentStop | 위임(collaboration) 수행 실행 — 종전 "SubagentStop 불발화(0.147.0)" 관측 재판정 (REQ-CEV-011) |
| Interrupt | 실행 중 런 인터럽트 (SIGINT 등) |

캡처: 훅 커맨드가 이벤트별 로그 파일에 JSON 페이로드 append. 판정: fired/not-fired + 페이로드 키 셋, 각각 실행 명령+관측 출력 귀속.

**not-fired 판정의 공허 방지 규칙**: not-fired 판정은 트리거 전제가 실제로 도달했음을 보이는 독립 증거(예: codex 자체의 compaction 알림이 `--json` 이벤트 스트림/세션 로그에 나타남)를 함께 운반해야 한다. 전제 미도달이 확인되거나 입증할 수 없으면 "does not fire"로 기재하지 않고 **"trigger-not-achieved"** 로 구별 기재한다 — PreCompact/PostCompact가 이 규칙의 직접 대상이다(위 트리거의 실현 가능성 미검증). "명령을 실행했으나 전제에 도달한 적 없는 런의 무발화"를 not-fired로 기록하는 것은 금지된다.

**산출물**: `.moai/reports/t496/codex-event-campaign.md` — 이벤트별 판정 테이블 + 페이로드 형상 + 처분(adapt-now / 후속 카드 / not-fire 문서화) + P0 경로 기록.

### M3 — 조건부 적응 (Medium)

M2 처분이 adapt-now로 표시한 이벤트만: 테이블 행 `Adapted` true 전환 + 디스패처 인자 확인 + 필요 시 golden/output 매핑. adapt-now가 0이면 M3는 no-op이고 M2 기록이 종결 산출물이다. 서브에이전트가 이 SPEC의 구조적 특성(불확정 측정 결과)상 M2 결과를 본 뒤에야 크기를 알 수 있어, 초과 시 레인이 분할을 리드에 보고한다 (REQ-CEV-012).

## G. 안티 패턴 (금지)

- `ErrUnadapted` 메시지 포맷을 빈 인자에서 그대로 재사용해 "dispatcher arg \"\" exists"라는 거짓을 새기는 것
- 테스트 통과 목적의 `moai hook interrupt` 유령 서브커맨드 등록
- 문서 부재를 not-fired 판정의 근거로 사용 (REQ-CEV-009)
- 캠페인의 연산자 `~/.codex` 직접 변형
- `internal/hook`으로의 어떤 변경 (REQ-CEV-002/005)
- 전체 스위트 로컬 실행 (영향 패키지 한정 — CLAUDE.local.md §4)

## H. 상호 참조

- spec.md §C (본 트리 측정 근거 M1-M5) · acceptance.md (AC 전문) · `.moai/reports/t494/codex-doc-survey.md` §3 (12종 열거의 근거인 외부 문서 주장 — 본 트리 재측정 불가, §C M6 참조; M2 캠페인이 설계된 후속 검증) · `.moai/reports/t83/precondition-measurement.md` (0.147.0 캠페인 선례)
