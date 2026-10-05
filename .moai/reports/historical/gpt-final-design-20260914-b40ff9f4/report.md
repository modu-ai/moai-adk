---
title: "최종 검토안 v3 — moai gpt Gateway와 App Server 재설계"
date: "2026-09-14"
lang: ko
---

## 0. 지정 문서 재조사 — 최종 검토안 v3

**권고: 외부 명령은 `moai gpt`로 통일하고, 기존 gateway·codexbridge를 재사용하여 일반 Anthropic Messages 입구와 Codex App Server 실행부를 연결한다.** 새 프록시 제품을 추가하거나 Claude Apps Gateway를 GPT 공급자로 개조하는 안이 아니다. 문서상 가능한 통합 표면과 현재 제품의 동작 증거를 분리했다.

### 읽은 범위와 남은 범위

사용자가 지정한 아래 **5개 페이지의 본문 및 페이지 내 모든 하위 목차를 끝까지 읽었다.** 한국어 HTML 페이지를 가져오지 못한 경우 같은 공식 경로의 `.md` 원문을 사용했다. gateway의 연결·배포·프로토콜 3개 하위 페이지도 포함한다.

| 지정 문서 | 읽기 상태 | 설계에 반영한 주제 |
|---|---|---|
| [Codex App Server](https://learn.chatgpt.com/ko-KR/docs/app-server) | 전체 | 인증·초기화·모델·thread/turn·동적 도구·취소·복구·실험 API |
| [LLM Gateway](https://code.claude.com/docs/ko/llm-gateway) | 전체 | 일반 게이트웨이의 범위, 구독과 인증, 비-Claude 비지원 |
| [연결](https://code.claude.com/docs/ko/llm-gateway-connect) | 전체 | 인증 변수·설정 우선순위·실행 표면별 차이·오류 점검 |
| [배포](https://code.claude.com/docs/ko/llm-gateway-rollout) | 전체 | 사용자별 인증·설정 배포·모델 제한·네트워크·점진 검증 |
| [프로토콜](https://code.claude.com/docs/ko/llm-gateway-protocol) | 전체 | Messages·SSE·기능 변환·오류 복구·모델 검색 |

추가로 SDK 본문 전체, 모델 설정의 gateway 관련 절, agent-view의 설정·모델·gateway·백그라운드 전환 절을 읽었다. prompt-caching은 이력·모델·도구 변경 관련 절을 읽었다. **전체 사이트의 모든 교차 링크를 재귀적으로 읽었다는 뜻은 아니다.** 전체 하위 목차와 직접 링크 목록 및 미독 범위는 [문서 조사 장부](documentation-review.md)에 있다. 한국어 settings-reference는 curl 두 번과 웹 열기에 실패했지만 공식 영어 원문의 modelPicker 절을 끝까지 읽어 보완했다. 오류 토큰 상세 앵커는 내려받은 대상 원문에 없어 해당 목록이 미확정이다.

### 지원·약관 판단

**“약관 문제가 전혀 없고 양사가 지원하는 Claude Code 안의 GPT”라고 보증할 수 없다.** 공식 일반 게이트웨이 문서는 비-Claude 모델 라우팅을 지원하지 않는다고 명시한다. 비지원과 약관 위반 확정은 다른 판단이며, 이번 조사로 위반 또는 무위험 어느 쪽도 확정하지 않는다. [Anthropic 지원 범위](https://code.claude.com/docs/ko/llm-gateway)

OpenAI 쪽은 문서화된 App Server의 관리형 로그인을 사용하고, 기존 토큰을 추출해 별도 비공개 endpoint로 보내거나 다른 클라이언트를 사칭하지 않는 통합을 제안한다. 계정 공유·한도 우회·보호 조치 우회는 배제한다. 이용 계정의 계약과 조직 정책 확인은 별도 승인 조건이다. [OpenAI 이용약관](https://openai.com/policies/terms-of-use/), [Anthropic 상업용 약관](https://www.anthropic.com/legal/commercial-terms)

| 대안 | 이번 판단 |
|---|---|
| Claude Code + MoAI Messages + App Server | 요청한 UX에 맞는 **조건부 구현 후보**. 양사 공식 지원을 모두 받는 경로는 아님 |
| Codex 자체 CLI/App Server/SDK로 Factory 실행 | Claude Code 내부라는 조건을 포기할 때 지원 경계가 더 명확한 대안. 자동으로 이 방식으로 바꾸지 않음 |
| Claude Apps Gateway | Claude 공급자·SSO용 제품. GPT bridge 해결책으로 채택하지 않음 |
| LiteLLM 등 추가 API 프록시 | 이 프로젝트의 App Server 세션·동적 도구 소유권 문제를 대신 해결한다는 증거가 없어 새 의존성으로 추가하지 않음 |
| 구독 토큰 추출·비공개 HTTP 재사용 | 새 제품 경로의 기준으로 채택하지 않음. 공식 인증 수명주기를 실행부에 맡김 |

[Codex SDK 문서](https://learn.chatgpt.com/codex/codex-sdk)는 자동화 작업과 풍부한 클라이언트 통합을 구분한다. TypeScript SDK와 안정 릴리스 Python SDK도 후보지만, 이미 Go `internal/codexbridge`가 있는 프로젝트에 Node/Python 실행층을 추가할 이유는 아직 없다. 인증·도구 요청 응답·이벤트를 직접 연결해야 하는 이번 설계에는 기존 Go bridge + App Server를 우선한다. SDK의 안정 릴리스가 동적 도구 계약 전체의 안정성을 보증하지는 않는다.

## 1. 결정 요약 — 구현 승인 전 최종안

**사용자가 확정한 모델 매핑은 Fable→Astra, Opus→Sol, Sonnet→Terra, Haiku→Luna다. 이전 보고서의 예시 매핑을 폐기한다.**

**칸반은 별도 실행 모드와 현행 제품 용어에서 제거한다.** 이를 다른 이름의 별도 모드로 되살리지 않는다. 자동화 실행은 Factory, 작업 목록은 Todo, 작업 배정은 Dispatch, 단계는 Plan→Run→Sync로 구분한다.

이 문서는 구현 전 설계 보고다. 제품 코드·저장된 작업 상태·기존 세션은 변경하지 않는다. 새 명칭, 기존 입력의 폐기 방식, 이력 보존 정책은 아래 권고안을 승인받은 뒤 적용한다.

| 구분 | 최종안 |
|---|---|
| 사용자 실행 | `moai gpt`, `moai gpt -f`, `moai gpt -f N`, `moai gpt -f lane-N` |
| 별도 모드 제거 | `-k/--kanban` 실행 분기 폐기. 다른 이름의 대체 모드 신설 없음 |
| 실행 제어 패키지 | `internal/kanban` → `internal/orchestration` |
| GPT 식별자 | `kanban.BackendGPT` → `orchestration.BackendGPT`, 저장 값 `"gpt"` 유지 |
| 모델 역할 | Fable=Astra / Opus=Sol / Sonnet=Terra / Haiku=Luna |
| GPT 실행 | Messages gateway + 공식 Codex App Server |
| 실제 도구 실행 | Claude Code의 승인·hooks·도구 실행 경로 유지 |
| 보존 | Todo 작업·UUID·소유권·DB·완료 근거·과거 커밋과 보고서 |

`orchestration`은 실행을 조정하는 공통 계층이라는 뜻이다. 코드가 실제로 담고 있는 세션·배정·상태·잠금·작업 큐 지원을 포괄한다. 이미 존재하는 `internal/workflow`에 억지로 합치지 않는다. 첫 변경에서 새 패키지를 여러 개로 분할하는 별도 리팩터링도 하지 않는다.

## 2. Baseline-attribution — 이전 보고서와 달라진 기준

| 항목 | 이번 확인 |
|---|---|
| 작업 트리 | 기존 `.claude/worktrees/develop` |
| 브랜치·HEAD | `develop` · `15f3eacd7` |
| 최신 병합 | `15f3eacd7 Merge branch 'WT-gateway-launchers' into develop (t654)` |
| 이전 보고서 | `c9ceff175` 기준이므로 t654 미병합 진술은 현재 기준으로 대체 |
| 이번 세션 | source session `01a09c3b-3734-7c30-b65d-650e3c63d0aa`; 디렉터리의 기존 조사 식별자는 유지 |
| tracked 파일 | 18,834개 |
| 텍스트 검사 | 18,565개, 바이너리 269개 제외, 누락·symlink 0개 |
| 검색식 | 대소문자 무시 `kanban|칸반`, 내용과 경로를 별도 검사 |
| 내용 적중 | 1,728개 파일 / 13,901개 적중 줄 |
| 경로 적중 | 262개 경로 |
| 내용 또는 경로 적중 | 1,733개 파일, 중복 제거 |
| 스캔 범위 | 해당 develop의 git tracked working-tree 파일 전체. ignore 규칙에 가려진 tracked 파일도 포함 |
| 제외 범위 | 다른 worktree, git 역사 전체 blob, untracked/ignored runtime 데이터, 외부 홈 디렉터리, 바이너리 내용 |

최초 `rg --json` 출력은 대형 보고서 때문에 잘렸다. 그 수치를 전수 결과로 사용하지 않았다. 이어 `git ls-files -z` 목록 전체를 Node로 검사하고 적중 파일·줄 번호만 추출해 완전한 JSON을 판독했다. [전수 목록 JSON](inventory.json), [목록 MD](inventory.md), [재현 스캐너](scan-names.js).

분류표는 **내용 또는 경로 적중 1,733개** 기준이다. 적중 수는 변경할 파일 수와 같지 않다. 특히 역사 보고서·SPEC 식별자는 보존 또는 검토 대상으로 분류한다.

| 분류 | 파일 수 | 내용 적중 줄 |
|---|---:|---:|
| 기타 | 21 | 187 |
| 에이전트·스킬·규칙 | 17 | 90 |
| 기존 보고서 | 937 | 6952 |
| SPEC | 334 | 3384 |
| 문서 사이트 | 93 | 435 |
| 구현·테스트·템플릿 | 331 | 2853 |

## 3. 명칭 사전 — 역할에 따라 하나의 이름만 사용

| 현재 명칭·표면 | 제안 명칭 | 의미와 처리 |
|---|---|---|
| Kanban Mode / 칸반 모드 | 제거 | 자동화 모드는 Factory만 유지 |
| `-k`, `--kanban` | 폐기 입력 | 명확한 종료 오류와 `-f` 안내, 실행하지 않음 |
| `internal/kanban`, `package kanban` | `internal/orchestration`, `package orchestration` | 첫 단계는 패키지 이동과 import 변경, 동작 보존 |
| `kanban.BackendGPT/Claude/GLM` | `orchestration.BackendGPT/Claude/GLM` | Backend는 런처의 초기 백엔드. 요청별 model과 별개 |
| `kanbanEntryParse` | `launcherEntry` | `Mode: interactive\|factory`, SPEC·lane·worker 수를 분리 |
| `KanbanEnabled`, `kanbanBranch*` | 제거 | Factory 여부를 거짓 Kanban 플래그로 표현하지 않음 |
| `prepareKanbanSettings` | `prepareDispatchSettings` | 세션 간 작업 전달 설정이라는 실제 역할 |
| `moai-kanban` 임시 파일 접두어 | `moai-dispatch` | 생성 파일 이름 |
| `kanban-dispatch.md`, `-detail.md` | `factory-dispatch.md`, `-detail.md` | Factory 배정 규칙. 단일 세션 milestone 지침은 orchestration 용어 |
| `moai-kanban-foreman` 참조·경로 | `moai-factory-foreman` | 실제 존재 표면을 목록으로 확정 후 스킬 metadata·호출자 동시 갱신 |
| Kanban 화면·`KanbanVM`·`buildKanban` | 작업 현황 / `TasksVM` / `buildTasks` | Todo와 Factory의 조회 화면 |
| `/kanban` | `/tasks` | 새 canonical route, 구 GET route는 명시적 redirect |
| `nav.kanban` 등 i18n key | `nav.tasks` 및 `tasks.*` | 한국어·영어·일본어·중국어 등 실제 locale 모두 동기화 |
| SSE `kanban` / `data-live="kanban"` | `tasks` | 서버·브라우저·watcher·테스트 동시 갱신 |
| `kanban-board` 저장 경로 | `task-board` | 사용 중인 board consumer가 필요로 하는 범위만 명시적 migration. 단순 rename 아님 |
| 기존 queue의 legacy `kanban` 경로 | migration reader의 legacy 상수만 보존 | 현재 canonical Todo DB 경로를 다시 옮기지 않음 |
| 새 milestone 보고 경로 `.moai/reports/kanban` | `.moai/reports/orchestration` | 새 출력만 변경. 기존 근거 링크는 보존 |
| `SPEC-KANBAN-*`, 과거 commit·verdict | 역사 식별자 유지 | 새 문서에 대응표 추가. 기존 증거를 새 이름으로 위장하지 않음 |

`board`라는 일반 자료구조 이름까지 무조건 `factory`로 바꾸지 않는다. 작업 큐는 Todo, 실행은 Factory, 읽기 화면은 Tasks, 내부 실행 조정은 Orchestration이라는 사전을 적용한다. 저장소 전체의 `kanban` 문자열 0건 대신 **현행 계약 0건 + 승인된 역사·호환 예외 목록**을 완료 기준으로 삼는다.

### 환경변수 계약

| 현재 | 새 canonical 변수 | 원칙 |
|---|---|---|
| `MOAI_KANBAN` | `MOAI_RUN_MODE=factory` | old-mode enabled 비트 제거, interactive는 명시 모델 또는 미설정 |
| `MOAI_KANBAN_ID` | `MOAI_RUN_ID` | Factory·세션 연결의 동일 run ID |
| `MOAI_KANBAN_SPEC` | `MOAI_SPEC_ID` | 실행 모드와 독립인 SPEC |
| `MOAI_KANBAN_BACKEND` | `MOAI_BACKEND` | 값 `gpt/claude/glm` 유지 |
| `MOAI_KANBAN_CARD` | `MOAI_CARD_ID` | 기존 카드 식별값 보존 |
| `MOAI_KANBAN_LABEL` | `MOAI_SESSION_ROLE` + `MOAI_SESSION_LABEL` | 역할과 표시 이름을 분리; 문자열 단순 복사 금지 |
| `MOAI_KANBAN_LEAD_ADDR` | `MOAI_LEAD_ADDR` | 동일 leader 연결 계약 |
| `MOAI_KANBAN_LEAD_NAME` | `MOAI_LEAD_NAME` | resolved 이름 유지 |
| `MOAI_KANBAN_SETTINGS_INJECTED` | `MOAI_DISPATCH_SETTINGS_INJECTED` | 설정 주입 확인 |
| `MOAI_FACTORY_WORKERS/WORKER` | 이름 유지 | 기존 Factory 전용 변수 재사용 |

새 생산자는 새 변수만 쓴다. 과도기 입력 해석기는 구·신 값이 모두 있고 다르면 오류로 멈춘다. 호환성이 확인된 run ID·backend·card metadata만 정규화하며, 구 mode 비트나 plan/run/sync companion을 Factory lane으로 자동 해석하지 않는다. 이미 실행 중인 세션은 강제로 환경을 바꾸거나 종료하지 않는다. 혼합 버전 join은 프로토콜 버전 검사로 거절한다.

## 4. 모델 매핑 — 사용자 지정 계약

**확정: 별도 MoAI 모델 티어를 만들지 않는다. 기존 네 모델 별칭을 실제 GPT 모델에 직접 1:1 매핑한다.**

| Claude Code 모델 별칭 | 실제 GPT ID | child 환경 |
|---|---|---|
| `fable` | `gpt-6-astra` | `ANTHROPIC_DEFAULT_FABLE_MODEL` |
| `opus` | `gpt-5.6-sol` | `ANTHROPIC_DEFAULT_OPUS_MODEL` |
| `sonnet` | `gpt-5.6-terra` | `ANTHROPIC_DEFAULT_SONNET_MODEL` |
| `haiku` | `gpt-5.6-luna` | `ANTHROPIC_DEFAULT_HAIKU_MODEL` |

이는 사용자 지정 역할 매핑이다. 모델 간 품질·가격을 이번에 벤치마크한 결과가 아니다.

**모델 매핑은 승인됐으며 아래 설정 문법은 구현 전 제안이다.** GPT 모델 설정의 키는 `fable / opus / sonnet / haiku`만 사용한다. 모델 선택을 위한 중간 티어, 별도의 등급명, 티어 변환 테이블은 추가하지 않는다.

### 모델 선택과 추론 강도 분리

- 모델 별칭은 사용할 모델을 선택한다. `--model opus`와 Agent의 `model: opus`는 GPT 모드에서 Sol을 선택한다.
- `effort`는 선택한 모델의 추론 강도다. 모델 별칭에서 값을 추정하거나 자동으로 강제하지 않는다. 예를 들어 `opus`를 선택했다는 이유로 effort를 `high`로 설정하지 않는다.
- effort를 지정하면 해당 모델의 지원 범위로 검증한다. 미지정 시에는 실행부의 기본 동작을 사용하며 MoAI가 임의 기본값을 새로 만들지 않는다.
- 모델 변경 후 기존 effort가 지원되지 않으면 명시 오류 또는 사용자 선택이 필요하다. 다른 값으로 조용히 바꾸지 않는다.
- `inherit`는 추가 모델 별칭이나 등급이 아니라 부모의 모델 선택을 상속하는 방식이다. 부모 effort의 상속 정책까지 함께 확정됐다고 해석하지 않는다.

### GPT 모델 설정

```yaml
llm:
  gpt:
    transport: codex-app-server
    models:
      fable: gpt-6-astra
      opus: gpt-5.6-sol
      sonnet: gpt-5.6-terra
      haiku: gpt-5.6-luna
```

`gateway_prepare.go:104–116`의 현재 non-GLM 공통 모델 반복을 GPT 전용 매핑 분기로 교체한다. GLM과 Claude의 기존 설정 동작은 유지한다. child 환경은 다음 네 값으로 조립한다.

```text
ANTHROPIC_DEFAULT_FABLE_MODEL=gpt-6-astra
ANTHROPIC_DEFAULT_OPUS_MODEL=gpt-5.6-sol
ANTHROPIC_DEFAULT_SONNET_MODEL=gpt-5.6-terra
ANTHROPIC_DEFAULT_HAIKU_MODEL=gpt-5.6-luna
```

### 모델 선택 로직

1. 시작 시 계정에 허용된 model catalog와 네 ID를 검증한다. 모델이 없으면 명확히 실패하고 다른 모델·계정·API 과금으로 자동 대체하지 않는다.
2. 네 슬롯의 map은 실행 시작 때 고정하고 모든 lane과 child에게 전달한다. 각 lane이 다른 계정의 map을 덮어쓸 수 없다.
3. main 모델은 명시 `--model` → 새 대화가 아닌 resume의 저장 선택 → 기존 사용자 설정·picker 순으로 보존한다. 모두 없을 때만 **권고 기본값 Opus→Sol**을 사용한다. 이 기본값은 이번 승인 대상이며 사용자 지정 네 매핑과는 별개다.
4. `--model fable`은 main만 Astra로 선택한다. 네 슬롯 전체를 Astra로 덮어쓰지 않는다.
5. Agent의 `model: opus/sonnet/haiku/fable`은 해당 슬롯으로, `inherit`는 부모의 **그 시점 실제 선택 모델**로 처리한다.
6. 직접 입력한 허용 GPT ID는 그 ID를 쓴다. 의미가 다른 Claude 전체 버전 ID를 문자열 추측으로 GPT에 매핑하지 않는다.
7. `opusplan`, `best`, `default`, `[1m]` 등 복합 alias·접미사는 별도 계약 테스트를 만든다. 검증되지 않은 조합을 임의 해석하거나 지원한다고 광고하지 않는다. `opusplan`의 목표는 plan=Sol / 그 외=Terra다.
8. 모델 변경은 thread idle에서 허용하고 pending tool 왕복 중에는 적용하지 않는다. 요청·선택·실제 turn 모델을 구분해서 기록한다.
9. reasoning effort는 모델 이름에서 추정하지 않는다. 모델별 허용 목록으로 검증하고, 알 수 없는 effort는 명시 오류로 처리한다.
10. startup env뿐 아니라 picker allowlist·custom model 표시·agent frontmatter·summary 경로까지 같은 resolver 계약을 확인한다.

네 alias 환경변수는 Claude Code 공식 설정 표면이다. 다만 그 표면이 존재한다는 사실과 비-Claude 모델 조합을 Anthropic이 지원한다는 것은 다르다. [모델 설정 문서](https://code.claude.com/docs/en/model-config)

## 5. Factory 실행 계약 — 별도 칸반 체인 제거

| 입력 | 목표 동작 |
|---|---|
| `moai gpt` | 대화형 GPT 세션, Factory 자동 시작 없음 |
| `moai gpt -f` | 기존 의미대로 1 lane 구성의 lead 진입 |
| `moai gpt -f 4` | 4 lane 구성의 lead 진입 |
| `moai gpt -f lane-2` | lane-2에 해당하는 worker 세션 진입 |
| `moai gpt -f 4 --model fable` | lead main=Astra, 네 역할 슬롯은 지정 매핑 유지 |
| `moai gpt -k ...` | 부작용 없이 폐기 안내 오류. `-f` 실행으로 몰래 전환하지 않음 |

이 표는 **런처 진입 의미**다. `-f N`만으로 N개의 창이 모두 즉시 열리거나 작업이 자동 승인된다고 확대하지 않는다. 기존 spawn·소유권·작업 배정 규칙을 유지한다.

이전 `-k SPEC-ID`의 대상 지정은 **제안 `-f --spec SPEC-ID`**로 흡수한다. 새 flag라서 기존 parser, pass-through `--`, 오류 문구, 테스트가 필요하다. bare `-k` 및 `-k --name plan/run/sync`는 자동변환할 동등 경로가 없으므로 종료 후 명시 재시작을 안내한다. 이미 진행 중인 구 체인의 상태를 새 Factory 상태로 추측해 재개하지 않는다.

Plan→Run→Sync는 작업의 단계로 남는다. 세 개의 별도 companion 세션을 의미하는 구 topology는 신규 실행에서 폐기한다. Factory lead가 Todo에서 허용된 작업을 lane에 배정하고, lane이 기존 단계별 전문 역할과 승인 경계를 지킨다. 완료는 메시지의 주장만으로 처리하지 않고 소유 token·generation·근거·저장 결과로 확인한다.

`launcherEntry`는 `Mode`, `SpecID`, `Workers`, `WorkerNumber`, `Rest`를 담는 중립 구조로 바꾼다. role/label은 별도 resolver에서 검증한다. 단순 package rename과 모드 폐기를 하나의 무차별 문자열 변경으로 섞지 않는다.

## 6. GPT 실행부 — gateway와 App Server를 결합

목표 경로는 다음과 같다.

1. Factory 런처가 run·lane·session·agent identity와 네 슬롯 map을 만든다.
2. Messages gateway는 인증·요청 한도·모델 권한·소유권을 검사한다.
3. 공식 Codex App Server의 thread/turn으로 GPT 실행을 이어 간다.
4. dynamic tool call은 Claude의 `tool_use`로 돌려준다.
5. Claude가 승인·hooks·실제 도구를 실행한다.
6. 다음 Messages의 `tool_result`를 저장된 pending RPC에 정확히 한 번 연결한다.
7. 텍스트 delta·heartbeat를 전달하고 durable completion 후 성공 terminal event를 보낸다.

thread binding에는 계정 scope·프로젝트·run·lane·Claude session·agent를 포함한다. backend는 `gpt`, transport는 `codex-app-server`, model은 네 ID 중 실제 선택값이다. 세 개념을 같은 필드로 합치지 않는다. thread ID와 tool call ID는 재개할 때도 유지한다.

Codex native shell·patch·MCP를 초기 bridge에서 비활성화해 도구 실행 주체가 둘이 되지 않게 한다. `dynamicTools`와 `item/tool/call`은 experimental이므로 버전·스키마를 고정하고 capability 확인 후 사용한다. [공식 OpenAI App Server 문서](https://developers.openai.com/codex/app-server/)

프로세스는 계정 profile lock 소유자를 하나로 명확히 하고 lane별 thread를 격리한다. 프로세스 하나에서 여러 lane을 multiplex할지, 독립 profile로 분리할지는 lock·재시작 실측으로 결정한다. **한 profile을 lane 수만큼 동시에 독점 lock할 수 있다고 가정하지 않는다.** 이 항목은 구현 전 spike 통과 조건이다.

### 현재 t654 병합을 반영한 판단

t654는 현재 develop에 병합됐다. 그러나 `productionGatewayHandlerFactory`는 현재도 `installedGPTBrokerForGateway()` → `auth.CodexBroker`와 `translate.PolicyGPTNative` → HTTP handler 조립 경로를 사용한다. `internal/cli` non-test source 검색에서 `NewAppServerAdapter` 호출은 찾지 못했다.

따라서 “t654 미병합”은 철회하지만, “실제 App Server transport 연결과 라이브 인수가 끝났다”는 판단으로 바꾸지는 않는다. `as5-verdict.md`도 gate-open, PTY, 실계정 이중 모드, Windows CI, 배포, appliedEpoch 생산 호출자 연결을 6개 Gap으로 기록한다. `as5-launcher-integration.md`의 `authMethod/outputPolicy` overlay는 표시 계약이며 실제 transport 연결의 대체 증거가 아니다.

### 일반 게이트웨이 설정 계약 — Apps Gateway 로그인과 분리

아래는 **MoAI의 제안 정책**이다. 공식 연결·배포 문서를 적용하되 조직 관리 정책을 덮어쓰는 방식은 쓰지 않는다. [연결 문서](https://code.claude.com/docs/ko/llm-gateway-connect), [배포 문서](https://code.claude.com/docs/ko/llm-gateway-rollout)

1. loopback 주소와 실행별 로컬 gateway token을 함께 전달한다. `ANTHROPIC_BASE_URL`만 바꾸지 않는다. Claude의 구독 토큰은 MoAI 실행부로 전달하지 않는다.
2. 일반 gateway에 `forceLoginMethod: "gateway"`를 주입하지 않는다. 저장된 Claude 로그인도 지우지 않는다. 관리 로그인 강제 정책과 충돌하면 이유를 표시하고 중단한다.
3. 최종 설정의 base URL·인증 종류·네 모델·허용 목록을 실행 전에 검사한다. 설정 파일의 `env`가 셸보다 우선하므로 “export했으니 적용됐다”는 판정을 하지 않는다. 비밀값은 로그에 남기지 않는다.
4. Factory lane마다 소유권이 확인된 scoped 설정을 전달한다. background supervisor가 셸 env를 항상 상속한다고 가정하지 않는다. `--settings` 전달 및 supervisor 재개 시 실제 endpoint readback을 필수 검사한다. [백그라운드 gateway](https://code.claude.com/docs/ko/agent-view#llm-gateway)
5. 현재 `gateway_session.go`의 `modelPicker`·`availableModels` 조립을 재사용해 네 항목의 역할 표시를 추가한다. 순수 GPT ID는 gateway 자동 검색의 필터를 통과하지 않으므로 `/v1/models` 검색에만 기대지 않는다. `claude-`를 붙여 GPT를 Claude 모델처럼 속이지 않는다. **modelPicker는 project/local 설정에서는 무시되므로 소유된 `--settings` overlay를 쓴다.** 관리 설정이 우선하며 모든 custom row가 탈락하면 기본 목록이 다시 보일 수 있어 allowlist·실제 picker readback을 함께 검사한다. [공식 modelPicker 계약](https://code.claude.com/docs/en/settings-reference#modelpicker)
6. `*_SUPPORTED_CAPABILITIES`는 일반 BASE_URL gateway에서 효력이 없으므로 이를 기능 협상 수단으로 쓰지 않는다. 표시명 변수와 기능 변수는 다르다. 모델별 실제 지원·변환 가능 여부는 MoAI resolver가 판단한다. [프로토콜](https://code.claude.com/docs/ko/llm-gateway-protocol), [모델 설정](https://code.claude.com/docs/ko/model-config)
7. 시스템 블록 재구성이 필요하면 문서화된 `CLAUDE_CODE_ATTRIBUTION_HEADER=0` 설정을 scoped profile에 적용한다. system-reminder나 시스템 지침을 정규식으로 지우는 수리는 하지 않는다.
8. 베타 헤더만 지우지 않는다. `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`도 adaptive thinking과 모든 도구 검색을 꺼 주는 만능 설정이 아니다. 각 필드를 변환·로컬 처리·명시 거부 중 하나로 분류한다.
9. Remote Control·voice·hosted web/Slack·Desktop까지 지원한다고 확대하지 않는다. 초기 인수 표면은 CLI의 대화형 세션·Agent·Factory lane이다. /fast와 자동 모델 fallback은 GPT 정책으로 검증되기 전 활성화하지 않는다.
10. `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`만으로 외부 트래픽 0을 주장하지 않는다. WebFetch 안전 검사 등의 별도 통신과 인증·업데이트 범위를 점검하며 안전 검사를 편의상 끄지 않는다.

### 요청 변환 계약 — 필드별로 의미를 보존

아래 표는 문서 복사가 아니라 **MoAI 구현 요구사항**이다. Anthropic 형식 그대로 전달하는 proxy와 다른 스키마로 바꾸는 adapter를 구분한다. [프로토콜 기준](https://code.claude.com/docs/ko/llm-gateway-protocol)

| 입력·상황 | 새 bridge의 처리 | 필수 음성 검사 |
|---|---|---|
| `/v1/messages?beta=true` | query를 허용하고 body를 독립 검증 | query 유무로 경로가 달라지지 않음 |
| 최초 system 지침 | base/developer instructions로 승인된 의미 변환 | MoAI가 프로젝트 지침을 중복 import하지 않음 |
| 대화 중 `role: system` | 해당 위치·권한의 context event로 처리 | 임의 user 입력을 system 권한으로 승격하지 않음 |
| `thinking`, `output_config.effort` | 모델이 허용한 effort·summary에만 명시 대응 | Anthropic signature를 Codex reasoning처럼 재사용하지 않음 |
| `cache_control` | 지원 차이를 명시; 캐시 성공·절감량을 만들지 않음 | 마커 제거를 실제 provider cache 성공으로 표시하지 않음 |
| `context_management` | client context epoch 변경을 감지·검증 | 삭제된 이미지나 이전 도구 결과를 몰래 다시 넣지 않음 |
| `tools`, `strict`, `defer_loading`, `tool_reference` | 인증된 현재 도구 snapshot에만 연결 | 존재하지 않는 참조·schema 의미 손실은 명시 거부 |
| `tool_use` / `tool_result` | RPC ID·call ID·owner·turn·digest ledger | 다른 lane 결과, 변형된 재전달, 중복 실행 차단 |
| `max_tokens`, stop 조건 | App Server와 동등한 계약이 아닌 부분을 표시 | byte limit을 token cap으로 광고하지 않음 |
| 이미지·구조화 출력 | 지원되는 입력과 outputSchema의 교집합만 허용 | 불지원 콘텐츠를 조용히 삭제하지 않음 |
| SSE 응답 | delta·ping을 흘리고 terminal만 durable 경계 뒤에 전송 | 전체 segment 완료 전 무응답·가짜 성공 금지 |
| `/count_tokens` | 정확한 계산 가능 여부를 표시 | 문자 추정을 실제 provider token 계측으로 표시하지 않음 |

**현재 어댑터 관찰:** `appserver.go`는 `Engine.Step` 이후 segment를 받아 메모리 buffer로 SSE를 만든다. 이는 소스에서 명시한 구간 buffering 정책이다. 라이브 지연을 측정한 것은 아니다. 새 설계에서는 결과 완료 판정의 안전성은 유지하면서 heartbeat/delta 전달 경로를 추가해야 한다. 현재 usage 0도 실제 무료 사용이나 실제 토큰 사용량 0의 근거가 아니다.

### App Server 실행 상태와 안전 경계

`New → Initialized → ThreadBound → Running → ToolPending → Running → Completed`를 기본 상태로 두고, 어느 단계에서든 `Cancelling / RecoveryRequired / Failed`로 갈 수 있게 한다. HTTP 응답 완료와 Codex turn 완료를 같은 상태로 취급하지 않는다.

- `initialize`와 `initialized`는 연결당 한 번이다. 로컬 stdio가 기본이며, 외부 websocket listener를 열지 않는다. 클라이언트 식별자는 MoAI 자신의 이름을 쓴다.
- 인증은 Codex 관리형 account 수명주기에 맡긴다. login 요청 발행과 login 완료를 구분하고, 세션 종료 때 공유 계정을 logout하지 않는다. API key 모드로 바꾸거나 크레딧을 소비하는 동작은 별도 승인이다.
- `model/list`의 계정별 결과와 effort·입력 종류를 확인한다. `modelProvider/capabilities/read`는 설치 schema에서 imageGeneration/namespaceTools/webSearch를 반환하며, 모든 기능을 망라하는 manifest로 과장하지 않는다.
- `dynamicTools`로 등록된 Claude 도구만 요청한다. Codex native shell·patch·MCP·자체 Agent는 초기 범위에서 차단하고, 예상하지 않은 native 승인 요청은 거절한다. `process/*`, `thread/shellCommand`, 임의 파일 API를 gateway 공개 API로 노출하지 않는다.
- `item/tool/call`에 대한 pending 응답은 같은 JSON-RPC request ID로 반환한다. 별도 `turn/start.toolOutput`을 그 응답의 대체물로 사용하지 않는다. 성공·실패 모두 durable ledger를 확인하고 반환한다.
- HTTP 연결이 tool_use에서 끝나도 Codex turn은 도구 응답을 기다릴 수 있다. 따라서 프로세스·RPC 연결·pending state는 해당 HTTP request context보다 오래 살아야 한다. 사용자 취소만 명시적 turn interrupt로 전달하고, 단순 HTTP 종료가 모든 lane을 종료하지 않게 한다.
- child와 summary는 identity·요청 목적이 검증된 독립 thread로 관리한다. 헤더의 agent ID 자체는 인증 증거가 아니다. 부모가 실행 중인 native fork를 기본 경로로 삼지 않는다.
- `model/rerouted`를 감지해 요청 모델과 실제 모델을 별도로 표시한다. 시작 시 `allowProviderModelFallback=false`만으로 모든 서비스 측 재라우팅을 막는다고 주장하지 않는다. 매핑 보장 모드에서는 예상과 다른 모델이 확인되면 성공으로 조용히 처리하지 않고 명시 정책에 따라 중단·보고한다. 안전 판정을 다른 모델로 우회하지 않는다.
- compact 요청의 즉시 응답은 완료가 아니다. contextCompaction 이벤트와 저장 상태를 확인하고, Claude와 Codex의 압축을 중복 수행하지 않도록 epoch을 분리한다.
- 목록 조회 시 appServer source 포함 여부를 확인한다. 조회·종료 편의를 위해 thread archive/delete나 externalAgentConfig/import를 자동 실행하지 않는다.
- 계정 한도와 lane별 대기열로 backpressure를 건다. `rateLimitsByLimitId`의 알 수 없는 값은 0으로 추정하지 않고, 한도 재설정 크레딧 소비·계정 교체를 자동 복구 수단으로 쓰지 않는다.

**버전 차이 확인:** 설치 버전은 Codex 0.154.0 / Claude Code 2.1.270이다. 이전 단계에서 생성한 0.154.0 schema를 이번에 다시 검사했을 때 `ThreadStartParams`에는 `instructionSources`가 없고 `ThreadResumeParams`에는 `dynamicTools`가 없었다. 또한 `multiAgentMode`는 deprecated/ignored로 설명된다. 최신 문서의 예제를 그대로 전송하거나 이 무시되는 플래그로 native Agent를 껐다고 판단하면 안 된다. 실행 바이너리에서 schema를 새로 생성하고 필요한 필드·차단 설정을 검증하는 것을 구현 진입 조건으로 둔다. 이 schema 관찰은 실제 서버가 임의 필드를 어떻게 처리하는지에 대한 런타임 검증이 아니다.

### 이력 조정의 결정표

| 비교 결과 | 제안 동작 |
|---|---|
| 같은 owner·epoch, 기존 부분 불변 + 새 user 입력 | 기존 thread에 새 입력 한 번 추가 |
| pending tool 결과 | 원 RPC에 결과 연결; 완료된 결과의 중복은 저장 응답 재사용 |
| 정상 system/context 추가 | 위치와 변경 근거를 검증한 context delta |
| 요약·rewind·이미지 제거·도구 registry 변경 | 검증된 전환 종류를 판별하여 새 epoch/thread 또는 지원되는 native 절차 사용 |
| 기존 tool 결과 본문 변경 | 같은 결과로 묵인하지 않음; 변경 원인과 증거를 표시하고 복구 필요 상태 |
| account·family·권한 불일치 | 실패 폐쇄. 같은 요청을 재시도하거나 다른 계정으로 보내지 않음 |
| 변경 이유를 식별할 수 없음 | 원문 보호·bounded 진단·사용자 복구 안내. 무조건 strip/replay 금지 |

Claude의 요청 이력은 항상 append-only인 API 계약이 아니다. 압축·이미지 제거·도구 정의 변경은 따로 고려해야 한다. 다만 일반 파일 편집은 과거 Read 결과를 소급 변경한다는 근거가 아니며, 문서는 알림을 추가한다고 설명한다. 이 구분을 t851의 성공/실패 wire에 대조해야 한다. [이력·캐시 동작](https://code.claude.com/docs/ko/prompt-caching)

오류는 `origin = client | gateway | codex-runtime | provider`, 내부 code, upstream HTTP status를 별도 기록한다. 실제 context 초과만 문서에 나온 `capability_rejected: prompt_too_long` 복구 표식 후보로 사용하고, receipt 불일치를 prompt-too-long으로 위장하지 않는다. 다른 token 이름은 참조 앵커 누락 때문에 아직 승인하지 않는다. SSE 시작 후에는 HTTP status를 뒤늦게 바꾸는 대신 규격에 맞는 error event와 중단을 사용하며, 부분 완료 뒤에는 자동 replay하지 않는다.

## 7. 400·이력·재개 로직

명칭 정리는 400 오류를 고치지 않는다. 네 모델 map만 분리해도 전체 transcript의 불변 prefix 가정이 남으면 오류가 재발할 수 있다. 이 작업은 이름 변경과 실행 경로·이력 수리를 각각 검증해야 한다.

- **새 입력:** 사용자 메시지·새 tool 결과·검증된 context 변경을 한 번만 반영한다.
- **기존 tool event:** owner·call ID·digest를 검증한다. 재전달이 곧 재실행을 뜻하지 않는다.
- **과거 context 변경:** 무조건 strip하지 않는다. 전달 가능한 새 context event로 정규화하거나 명시 새 thread로 분기한다.
- **자식 Agent:** 독립 child thread, 부모는 결과를 기다린다. 실행 중인 부모 turn의 native fork를 필수 전제로 삼지 않는다.
- **compact/resume:** 실제 Claude 동작과 epoch·binding 복구를 연결한다. 성공 메시지만으로 회상 성공을 판정하지 않는다.
- **400:** malformed request, receipt chain, reasoning, scope 오류를 구분한다. 같은 잘못된 이력을 무한 재시도하지 않는다.
- **429/5xx:** 제한된 backoff. 스트리밍 시작 후나 도구 실행 여부가 불명확하면 자동 replay하지 않는다.
- **crash:** pending RPC의 완료가 불명확하면 복구 상태를 노출한다. 파일을 바꾸는 tool을 추측해서 다시 실행하지 않는다.

이전 보고서의 synthetic 이력 실험은 이전 HEAD에서 수행한 증거로만 인용한다. 이번 새 HEAD에서 현장 wire 원인을 추가 확정했다고 주장하지 않는다. 실제 오류의 최초 변형 위치·값은 성공/실패 wire의 구조·digest 계측으로 확인해야 한다.

## 8. 데이터·역사·호환성 처리

1. 현재 Todo DB는 `~/.moai/db/<project-key>/todo/backlog.db` 경로를 이미 사용한다. 패키지 이름 변경 때문에 DB를 새로 만들거나 기존 큐를 초기화하지 않는다.
2. 세션 JSON의 `backend: "gpt"`, UUID, run/card ID, generation, claim token, 완료 근거는 보존한다. `BackendGPT`의 import 경로를 바꾼다고 저장값을 바꾸지 않는다.
3. `kanban-board`나 legacy queue 경로는 사용 여부와 실제 writer를 확인한 후 migration 목록으로 분리한다. read endpoint와 `LoadPure`는 migration을 실행하지 않는다.
4. migration은 dry-run 목록 → 세션 drain·writer admission → backup/manifest → 트랜잭션 또는 원자 이동 → readback → 새 canonical 확정 순서다. 충돌하는 두 원본은 합치거나 덮어쓰지 않는다.
5. 새 web route는 `/tasks`. 구 `/kanban` GET은 새 route로 redirect하고, 상태 변경 메서드를 새로 허용하지 않는다.
6. SSE 새 서버는 `tasks`를 한 번 emit한다. 전환기 새 client가 구 `kanban` event도 수용할 수 있으나, 같은 이벤트를 두 이름으로 중복 발행하지 않는다. endpoint/version 교체 범위를 테스트한다.
7. `.templ` 원본을 고친 뒤 생성 Go를 재생성한다. 규칙·스킬·agent와 `internal/template` mirror, catalog hash·문서 nav를 함께 갱신한다.
8. 과거 SPEC ID·보고·로그·commit history는 보존한다. 활동 중인 SPEC의 현행 설명과 참조는 소유권을 확인해 변경하되 원래 acceptance 의미를 지우지 않는다.
9. 예외 목록은 정확한 path·literal·유형·이유·제거 조건을 갖는다. 단순히 `.moai/**` 전체를 제외하는 방식은 금지한다.
10. 롤백은 새 작업이 쓰인 DB를 과거 backup으로 덮는 행위가 아니다. 쓰기 정지·새 변경 식별·호환 reader 또는 별도 복구 절차를 통해 처리한다.

## 9. 구현 순서와 인수 조건

| 단계 | 변경 | 완료 증거 |
|---|---|---|
| A — 명칭·mapping 계약 승인 | 이 문서의 명칭 사전, -k 폐기, default main, 역사·호환 정책 확정 | 운영자 승인 |
| A-1 — 핵심 연결 사전 검증 | 지원 경계 수용, 실행 버전 schema, native 도구 차단, pending RPC, profile lock | 파일 쓰기 없는 1개 dynamic tool 왕복, 독립 child, 모델 4슬롯 실제 관찰; 실패하면 대규모 rename 전에 재설계 |
| B — 동작 보존 rename | 패키지/import/중립 타입·출력 명칭, rule/template 동기화 | 관련 package tests·build·생성물 일치, 저장 JSON 불변 |
| C — Factory 단일 실행 모드 | 구 분기 제거, 새 env·launcherEntry, SPEC 지정 | -f positive, -k no-side-effect negative, pass-through·mixed-version tests |
| D — GPT 네 슬롯 | config·env·picker·Agent·summary resolver | 아래 네 모델 matrix, main override가 map을 바꾸지 않음 |
| E — App Server 수직 연결 | runtime·identity·pending tools·stream·reconciliation | 1 lane + child + Read/Edit/후속 turn 양성·음성 쌍대 |
| F — 상태·다중 lane 인수 | migration·resume·compact·cancel·429·crash·N lane | 실계정 증거와 상태 readback, 계정별 한도 준수 |
| G — 배포 판정 | 바이너리 provenance·Windows CI·사용자 UI | 실제 실행 경로·모델·완료 결과가 일치 |

### 필수 모델 인수 matrix

| 요청 | 예상 실제 모델 | 검사 위치 |
|---|---|---|
| main `--model fable` | `gpt-6-astra` | launch env + gateway request + Codex turn |
| Agent `model: opus` | `gpt-5.6-sol` | child request + child thread |
| Agent `model: sonnet` | `gpt-5.6-terra` | child request + child thread |
| Agent `model: haiku` | `gpt-5.6-luna` | child request + child thread |
| Agent `model: fable` | `gpt-6-astra` | child request + child thread |
| main Sol 상태에서 `inherit` | `gpt-5.6-sol` | 부모 선택 snapshot + child |
| main Astra로 변경 후 `inherit` | `gpt-6-astra` | 변경 완료 경계 + child |
| main override Astra 뒤 다른 Agent의 `sonnet` | `gpt-5.6-terra` | 슬롯 map이 오염되지 않음 |
| lane-1 Sol / lane-2 Terra | 각각 지정 모델 | 독립 binding·tool result 오배달 0 |
| 없는 model / 잘못된 effort | 명시 실패 | silent fallback·billing 전환 0 |

이름 변경 완료는 활성 CLI/help/런타임 식별자/화면/새 규칙·템플릿에서 구 명칭이 0건이고, 모든 나머지 적중이 승인된 역사·migration 예외로 분류됐을 때다. 제품 완료는 실제 Factory 작업이 끝나고 저장 상태와 결과를 확인한 뒤 별도로 판정한다.

## 10. Evidence — 이번 검사

명령과 결과의 전체 범위는 [evidence.md](evidence.md)에 기록한다.

```text
git branch --show-current
develop

git rev-parse --short HEAD
15f3eacd7

선택 테스트 package 결과:
ok  github.com/modu-ai/moai-adk/internal/cli 1.000s
```

선택자는 `TestGPTLaunchPreservesCommonEntry`, `TestGatewayProviderContractPickerAndSlots`, `TestGatewayLaunchAssemblyCarriesAuthDisplay`, `TestGatewayLaunchTransportGateControl`, `TestParseFactoryFlag`다.

이 결과는 **현재 old mapping과 gate 대조군 계약**의 통과다. 새 이름·새 네 슬롯 map·실제 App Server 팩토리는 아직 구현되지 않았으므로 새 설계 PASS라고 표현하지 않는다.

## 11. Gaps·Residual-risk·승인 항목

- 이름 적중 전체 목록은 확보했지만 1,733개 파일의 모든 의미·writer·외부 소비자를 전부 추적한 것은 아니다. migration·삭제 전에 각 소비자 graph를 검증한다.
- 바이너리 이미지 안의 글자, git history, 다른 worktree, 설치된 사용자 홈의 배포 복사본은 이번 텍스트 전수 범위 밖이다. 배포 검증 때 해당 표면을 별도로 검사한다.
- 새 HEAD에서 실제 400 실패 wire, 유료 모델 호출, 다중 lane, Windows CI, DB migration을 실행하지 않았다.
- App Server dynamic tool API의 실험 상태·profile lock 전략·기존 session resume는 운영 위험으로 남는다.
- 비-Claude gateway 라우팅에 대한 Anthropic 비지원 경계는 명칭이나 모델 map을 바꿔도 사라지지 않는다. 공식 OpenAI 통합 경로를 쓰는 것과 양사 지원·약관 무위험 보증은 별개다. [Anthropic 지원 범위](https://code.claude.com/docs/en/llm-gateway)
- 이번 설계는 계정 공유·제한 우회·토큰 전달·API 자동 과금 전환을 허용하지 않는다.

**확정된 모델 계약:** 별도 티어 없이 `fable→Astra`, `opus→Sol`, `sonnet→Terra`, `haiku→Luna`를 직접 매핑한다. effort는 독립 설정이다. 이번 승인을 제품 전체의 인수 완료로 표시하지 않는다.

**나머지 설계 승인 범위:** Factory 단일 모드, `orchestration` 공통 패키지, Todo/Tasks/Dispatch 역할 사전, 구 `-k`의 자동변환 없는 폐기, 역사·데이터 보존이다. 기존 main 설정이 없을 때 Opus→Sol을 기본값으로 제안한다. 모델 매핑 확인만으로 나머지 선택과 실계정 실행까지 승인된 것으로 확대하지 않는다.

**이번 문서 수정 기준:** `develop` / `4056f69e1`. 앞선 코드 조사·테스트·전수 목록의 기준은 `15f3eacd7`로 보존하며, 새 HEAD에서 재측정한 결과로 취급하지 않는다.

## 12. 주요 근거

- [현재 GPT entry](../../internal/cli/gpt.go), [모델 env 조립](../../internal/cli/gateway_prepare.go)
- [현재 GPT handler](../../internal/cli/gateway_product_binding.go): 198–257행
- [현재 mode parser](../../internal/cli/factory.go), [구 mode parser](../../internal/cli/kanban.go)
- [backend 저장 계약](../../internal/kanban/record.go), [현재 Todo 경로](../../internal/kanban/state_dir.go)
- [환경변수](../../internal/config/envkeys.go), [웹 route](../../internal/web/app.go), [SSE 계약](../../internal/web/events.go)
- [t654 App Server 최종 판정](../../.moai/reports/t654/as5-verdict.md), [조립 증거](../../.moai/reports/t654/as5-launcher-integration.md)
- [이전 연구 보고 — 해당 baseline의 오류 실험과 라이브러리 비교](../gpt-factory-20260914-01a09c3b/report.md)
