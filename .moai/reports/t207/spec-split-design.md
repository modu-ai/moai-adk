# SPEC-WEB-CONSOLE-015 분할 설계 (t207)

운영자 판정(2차 감사 F4 대응, kickoff 보류)에 따른 분할 설계서. 이 문서는 **분할의 설계**이며,
SPEC 본문 편집은 이 설계가 비준된 뒤에 수행한다. 판정 근거는 `plan-audit-iter2-independent.md`
F1·F4·F5·F7, 1차 감사 `plan-audit.md` N1·N2.

측정 기준 트리: `WT-web-live-todo` @ `52358e72f` (직전 SPEC 편집 커밋 `449736deb` 이후 변경 없음).

---

## 1. 요약 — 무엇이 왜 갈라지는가

세 갈래로 나눈다. 좌표는 두 개다: **예산**(F4 — REQ 25/25, AC 29/25)과 **의존 성분**
(plan.md §E 의 의존 그래프가 이미 `M1→M2→M5 / M3→M5` 와 `M4→M6` 두 개의 연결 성분으로 끊겨 있다).

| SPEC | 범위 | Tier | REQ | AC | 의존 |
|---|---|---|---|---|---|
| **A** `SPEC-SESSION-TELEMETRY-001` (신규) | statusline 세션별 스냅샷 — 경로 분할 + 모델·effort 동승, 리더 1개, `moai tokens` 이관, 독트린 4파일, docs-site 12파일 | L | 9 | 10 | 없음 |
| **B** `SPEC-WEB-TODO-QUEUE-001` (신규) | 큐 루트 해석 이전(순수/입양 분리) + `/todo` 라우트·6번째 내비·배지 | M | 8 | 11 | 없음 |
| **C** `SPEC-WEB-CONSOLE-015` (유지) | `kanban.Record` 레인번호·카드id, 런처 배선, 콘솔 소비(텔레메트리 셀 + 팩토리 레인 섹션) | L | 16 | 17 | **A** |

세 SPEC 모두 각자의 상한 아래에 여유를 남긴다 — F5·F7 이 요구하던 자리가 생기고, 재감사에서
새 REQ/AC 가 필요해져도 상한을 다시 들이받지 않는다. **거버넌스 편차(N1)는 발급하지 않는다.**
분할이 초과 자체를 없애므로 편차를 기록할 대상이 남지 않는다.

병렬 가능: **A 와 B 는 서로 독립**이라 동시에 돌 수 있다. C 는 A 가 착지한 뒤.

---

## 2. F1 이 어디로 가는가 — 그리고 왜 이것이 분할선을 바꾸는가

**F1 은 A 로 간다. C 로 가지 않는다.** 고치는 방식이 M2 를 손보는 것이 아니라 **생산자를 옮기는
것**이기 때문이다.

### 2.1 측정 — 런처는 세션의 모델을 알지 못한다

리드가 재확인한 두 줄(`EffectiveProfile` 이 프로파일 이름을 반환, `ModelEffort` 가 Claude 별칭
어휘)에 더해, 이 트리에서 한 겹 더 측정했다:

```
$ grep -rn '"\-m"\|"--model"\|Model' internal/cli/cc.go
(출력 없음)

$ grep -n "ANTHROPIC_DEFAULT\|--model\|EnvAnthropic" internal/cli/cc.go
36:  -m, --model <model>           Override model selection      ← 도움말 문자열 1건뿐
```

`moai cc` 는 모델을 **파싱하지도 설정하지도 않는다**. `-m` 은 백엔드로 흘려보내는
패스스루이고, 런처 코드가 읽는 지점이 없다. 세션의 실제 모델은 Claude Code 가 settings 와
`/model` 로 정한다 — 런처가 기록할 값을 애초에 손에 쥔 적이 없다.

GLM 쪽은 반대 방향으로 어긋난다. `internal/cli/glm.go:350-353` 이 심는 것은 단일 모델이 아니라
**슬롯 맵**이다(High/Medium/Low/Fable → z.ai 모델명). 세션이 그중 어느 슬롯에서 도는지는
런처가 모른다.

따라서 REQ-WC15-011("런처가 세션의 모델과 effort 를 기록한다")은 **두 백엔드 모두에서 구현
불가**다. F1 은 AC 의 표현 결함이 아니라 요구사항의 생산자 지목 결함이었다.

### 2.2 실제 생산자는 이미 존재하고, 이미 세션별이며, 이미 A 의 소관이다

statusline 은 매 렌더마다 Claude Code 로부터 세션 단위 입력을 받는다:

```
internal/statusline/types.go:69    Effort *EffortInfo `json:"effort"`   // v2.1.139+
internal/statusline/types.go:131   EffortInfo{ Level string }           // low/medium/high/xhigh/max
internal/statusline/types.go:148   ModelInfo{ DisplayName string }      // "Opus" 등
internal/statusline/builder.go:286-287   data.Effort = input.Effort
internal/statusline/metrics.go:51        modelName = resolveGLMModelName(modelName)
```

마지막 줄이 F1 의 GLM 절반을 이미 풀어둔 지점이다 — GLM 백엔드에서 Claude 표시명을 실제 z.ai
모델명으로 바꾸는 해석기가 **이미 있고 이미 호출된다**.

그리고 이 값들이 살아 있는 바로 그 함수가 컨텍스트 스냅샷을 쓴다
(`builder.go:157` `writeContextUsage(...)`). A 가 그 파일을 세션별로 쪼개고 나면, 모델과 effort 는
**같은 레코드에 같은 쓰기로** 실린다. 새 생산자도, 새 파일도, 여덟 개 호출부 배선도 없다.

### 2.3 결과 — 세 축의 값 하나가 축을 옮긴다

| 값 | 종전 계획 | 분할 후 |
|---|---|---|
| model | `kanban.Record` + 런처 배선(M2, 8 호출부) — **구현 불가** | A 의 세션 스냅샷 (백엔드별 해석 포함) |
| effort | 같음 — 구현 불가 | A 의 세션 스냅샷 |
| context % | A 의 세션 스냅샷 | 변동 없음 |
| lane 번호 | `kanban.Record` + 런처 | 그대로 C (런처가 **실제로 아는** 값) |
| card id | `kanban.Record` + 런처(worktree basename) | 그대로 C (같은 이유) |

M1 은 두 필드(lane, card id)로 줄고, M2 는 여덟 호출부를 그 두 값만 들고 지난다. 카드가
지목한 `kanban.Record` 확장은 사라지지 않고 **레인 신원**으로 좁혀진다.

### 2.4 카드 전제를 한 번 더 정정한다는 사실을 명시한다

카드 t207 은 "kanban.Record 확장으로 모델·effort·CW 기록"이라고 적었다. spec.md §A.1 이 이미
1축의 전제(SSE vs 폴링 택일)를 측정으로 정정한 선례를 세웠고, 이것이 **두 번째 정정**이다.
정정되는 것은 기록 **수단**이지 의도가 아니다 — 콘솔이 세션별 모델·effort·CW 를 보여준다는
목표는 그대로다. A 의 spec.md §A 에 이 정정을 §A.1 과 같은 형식(주장 → 증거표)으로 적는다.

### 2.5 리드 비준이 필요한 지점 — D-1

**D-1. 모델·effort 를 컨텍스트 스냅샷 레코드에 동승시킨다** (`schema_version` 상향).
대안은 (b) `kanban.Record` 유지 → §2.1 측정으로 기각, (c) statusline 이 세션당 두 번째 파일을
쓴다 → 같은 쓰기 경로에 파일만 늘어 기각. (a) 를 택하면 레코드의 의미가 "컨텍스트 사용량"에서
"세션 텔레메트리"로 넓어지므로 **파일명/타입명을 그대로 둘지**가 함께 걸린다.

제안: 온디스크 경로는 `.moai/state/context-usage/<session-id>.json` 로 **유지**(독트린·docs-site
12파일이 이미 이 이름을 가리키고, 이름까지 바꾸면 같은 스윕에서 두 가지를 동시에 옮기게 된다),
Go 타입명만 `sessionTelemetryRecord` 로 넓힌다. 비준 대상: 이름 유지 + 타입명 확장 + 스키마
버전 상향.

---

## 3. SPEC A — `SPEC-SESSION-TELEMETRY-001`

세션별 텔레메트리 스냅샷: 경로 분할(하드 컷) + 모델·effort 동승 + 소비자 전량 이관.

**Tier L.** 파일 23개(statusline 4, cli 2, 독트린 4, docs-site 12, `drift_cache.go` 주석 1)로
15 파일 상한을 넘고, 항상 로드되는 독트린의 읽기 절차를 바꾸므로 constitutional 축에서도 L 이다.
LOC 는 작지만 tier 표의 파일 축이 결정한다. 5-아티팩트, 임계 0.85.

### REQ 배정 (9 / 25)

| # | 종전 | 내용 |
|---|---|---|
| A-001 | WC15-020 | 세션별 경로 `.moai/state/context-usage/<session-id>.json` 에 기록하고, 단일 슬롯 경로에는 기록하지 않는다(G-3 하드 컷) |
| A-002 | 신규(F1) | 레코드는 컨텍스트 값과 함께 그 세션의 모델·effort 를 싣는다. 입력에 없으면 키를 비우고 추정하지 않는다 |
| A-003 | 신규(F1) | 기록되는 모델명은 그 세션이 **실제로 도는** 모델이다 — GLM 백엔드에서는 Claude 표시명이 아니라 z.ai 모델명 |
| A-004 | WC15-021(생산자 절반) | `internal/statusline` 은 이 레코드의 리더를 **정확히 하나** 내보낸다 |
| A-005 | WC15-022 | `internal/` 어디에도 이 레코드 스키마의 두 번째 선언이 없다 |
| A-006 | WC15-025 | `moai tokens` 의 중복 리더를 제거하고 A-004 리더로 이관한다. `moai tokens` 출력의 컨텍스트 블록은 계속 나온다 |
| A-007 | AC-052 승격 | 세션 id 가 경로 구성요소가 되므로, 경로를 벗어나는 값은 **거부**하고 리디렉션하지 않으며 렌더 자체는 계속된다 |
| A-008 | WC15-024 | 독트린 두 미러쌍(4파일)을 같은 변경에서 갱신한다 (`Where` → `When`, F13) |
| A-009 | 신규(F5) | docs-site 4로케일 12파일을 같은 변경에서 갱신한다 |

### AC 배정 (10 / 25)

A-010 경로 존재 + 구경로 부재 · A-011 모델·effort 왕복(입력 없으면 키 없음) · A-012 GLM 환경에서
z.ai 모델명 기록(Claude 별칭 아님) · A-013 내보낸 리더 정확히 1개(**baseline 0**) ·
A-014 `"raw_pct"` 전량이 `internal/statusline` 안(**baseline 4 hits / 4 files, 그중 2개가
statusline 밖**) · A-015 tokens 제거 절반 + `moai tokens` 블록 잔존 절반 · A-016 악성 세션 id
4종에서 디렉터리 밖 파일 0 + 렌더 완료 · A-017 독트린 grep 0 + 신경로 4파일 존재 + 미러쌍 diff
공집합(**baseline 4파일**) · A-018 docs-site 구경로 grep 0 + 신경로 12파일 + 4로케일 균등
(**baseline 12파일**) · A-019 구 레코드 읽기 → 없는 필드는 "미기록", 실패 없음

부재로 만족되는 기준은 전부 baseline 을 함께 적는다(교훈: 사전구현 트리 0히트인 grep 만 채택).

### 흡수하는 finding

F1(생산자 재지목·GLM 분기) · F5(docs-site) · F10(AC-025 baseline, `:79`→`:81`) ·
F13 의 `Where`→`When` 과 REQ-025 서술형 전환 · N2/F8(신규 SPEC 이므로 0.1.0 시작).

---

## 4. SPEC B — `SPEC-WEB-TODO-QUEUE-001`

큐 루트 해석 이전 + `/todo` 라우트. plan.md §E 에서 이미 끊겨 있던 `M4 → M6` 성분 그대로다.

**Tier M.** 파일 12-14개(`internal/cli/todo.go`, `internal/kanban` 신규+테스트,
`internal/web` 의 `app.go`/`shell.templ`/`icons.templ`/`screens.go`/신규 templ/viewmodel,
`assets/i18n.js`, 테스트). 5-15 파일 구간, 스키마 변경 없음, 항상 로드 독트린 무관.
3-아티팩트, 임계 0.80.

### REQ 배정 (8 / 16)

| # | 종전 | 내용 |
|---|---|---|
| B-001 | WC15-032 | 콘솔은 백로그 변경 연산을 호출하지 않고 락을 잡지 않는다 |
| B-002 | WC15-030(라우트 절반) | 자체 최상위 라우트 `/todo` + 6번째 내비 항목 + 아이콘 case + `Area` 값 (G-4) |
| B-003 | WC15-030(내용 절반) | 세 상태 전부를 id·본문·상태 배지·SPEC id 와 함께 나열하고 어느 것도 걸러내지 않는다 (G-5) |
| B-004 | WC15-031 | 루트 해석을 두 패키지가 import 하는 자리로 이전하고, **해석과 입양을 분리**한다. 콘솔이 쓰는 진입점은 어떤 분기에서도 파일시스템을 변경하지 않는다 |
| B-005 | 신규(F3) | 비-git 폴백 분기에서 콘솔이 **무엇을 렌더하는지** 정한다 (§4.1) |
| B-006 | WC15-033 | 부재·빈 파일·깨진 JSON → 200 + 빈 상태 |
| B-007 | WC15-034 | 기존 `kanban` 이벤트로 갱신한다. **조건절을 요구사항 본문에 단다**(F6) |
| B-008 | WC15-050(부분) | 새 사용자 노출 문자열 4로케일 |

### AC 배정 (11 / 16)

B-010 세 행 + 배지 · B-011 워크트리에서 primary 의 N 건 · B-012 `todo.go` 위임(**grep 형태로**:
`gitcore.ResolveGitDirs` 0히트 — F11) · B-013 폴백 분기 디스크 무변경(원본 경로·mtime 유지) ·
B-014 **폴백 분기 렌더 결과**(F3) · B-015 `moai todo` 입양 동작 불변 · B-016 락 mtime 불변 +
백로그 바이트 동일 · B-017 부재/깨짐 200 · B-018 `data-live="kanban"` + `watchMap`/`EVENTS`
diff 불변 · B-019 내비 6행 + 아이콘 case + `aria-current` (**baseline 5행, case 없음**) ·
B-020 i18n 거버넌스 테스트 무 allowlist 통과

### 4.1 B-005 가 정해야 하는 것 (F3) — 리드 비준 필요

감사관이 짚은 발산: 순수 리졸버는 `~/.moai/todo/<key>` 를 반환하고 **입양하지 않으므로**,
프로젝트 로컬에 카드가 남아 있는 비-git 실행 맥락에서 콘솔은 **빈 큐**를, `moai todo` 는
**N 건**을 보여준다. 어느 쪽을 먼저 실행했는지에 따라 화면이 갈린다.

**D-2 제안: 읽기 관통(read-through).** 순수 리졸버는 폴백 루트에 큐 파일이 없고 프로젝트 로컬에
있으면 **프로젝트 로컬 경로를 반환**한다 — 여전히 아무것도 쓰지 않는다. 콘솔과 `moai todo` 가
같은 카드를 보고, 입양은 `moai todo` 가 처음 돌 때 종전대로 일어난다. 대안은 발산을 §C.6 식
한계로 기록하는 것인데, 이 SPEC 이 여는 첫 문장이 바로 그 실패("primary 에 30장, 워크트리에서
빈 큐")라 기록으로 닫는 선택은 자기모순에 가깝다.

### 흡수하는 finding

F3(B-005/B-014) · F6(B-007 조건절) · F11 의 AC-031b 절반 · F13 의 REQ-030/031 서술형 전환.

---

## 5. SPEC C — `SPEC-WEB-CONSOLE-015` (id 유지)

레인 신원 기록 + 콘솔 소비(텔레메트리 셀 + 팩토리 레인 섹션). 카드 t207 과 두 차례 감사 이력이
붙어 있는 id 이므로 **유지**한다. 제목은 세 축 → 두 축으로 좁혀 다시 쓴다.

**Tier L.** 3 패키지(`internal/kanban`·`internal/cli`·`internal/web`) + `kanban.Record`
스키마 확장(`@MX:ANCHOR` 가 보이지 않는 리더를 묶는 자리) + 런처 8 호출부.

### REQ 배정 (16 / 25)

프레이밍 2: C-001 전송 계층 무변경 · C-002 콘솔 읽기 전용.
레인 신원 4: C-003 `Lane int`(0=레인 아님, G-6) · C-004 기록 시 레인 번호 보존(`WithRole`
드롭 가드 유지) · C-005 카드 id(worktree basename + env 오버라이드, 둘 다 없으면 공란, G-1) ·
C-006 런처 8 호출부가 두 값을 배선.
콘솔 소비 6: C-007 A 의 리더를 소비하고 자체 리더를 선언하지 않는다 · C-008 해당 세션 레코드가
없으면 "미기록"이며 **다른 세션 값을 대신 쓰지 않는다** · C-009 PID 조인, 새 상태 파일 없음 ·
C-010 **동일 PID 두 레인을 한 세션에 오귀속하지 않는다**(F7) · C-011 레인 행 구성 · C-012 추정
표식.
견고성·횡단 4: C-013 레지스트리 부재/깨짐 → 0 레인, 200 · C-014 4로케일 · C-015 구 레코드 관용 ·
C-016 `screens.templ:192` 의 고정 영문 안내 배너를 갱신한다(F12 — 변경 후 거짓이 되는 문장이고,
i18n 키가 빈 문자열이라 4로케일 규칙에도 걸린다).

### AC 배정 (17 / 25)

REQ 당 1개를 기본으로 하고, C-003(스키마 왕복 2건)·C-008(A/B 두 세션 대조)에만 둘씩. 부재로
만족되는 기준은 baseline 동반.

### 삭제되는 것

**REQ/AC-WC15-012 를 삭제한다**(F2). "미기록" 표식은 `screens.templ:165-175` +
`widgets.templ:122-124` 로 **이미 구현·배선·렌더되고 있고**, `viewmodel_ops.go:253` 이
`Model: ""` 를 하드코딩하므로 AC 의 Given 이 현재 상태 그 자체다. 두 번째 절("빈 `<td>` 없음")은
이 뷰가 `div`/`span` 기반이라 어떤 구현에서도 참이다. 남는 honesty 요구는 C-015(구 레코드 관용)와
C-008(다른 세션 값 금지)이 이미 진다.

### 흡수하는 finding

F2(삭제) · F7(C-010) · F12(C-016) · F9(AC-002 grep 표현) · F11 의 AC-043 절반 ·
F13 의 REQ-040/042 서술형·환경변수 상수화 메모 · **N2/F8**: `version: 0.2.0` + HISTORY 행 1개
("분할: 세션 텔레메트리는 SPEC-SESSION-TELEMETRY-001, todo 큐는 SPEC-WEB-TODO-QUEUE-001 로 이관").

---

## 6. 의존 순서와 착지 계획

```
A ─────────────► C
       B (독립)
```

- **A 먼저.** C-007/C-008 이 A 의 리더를 소비한다. A 가 없으면 C 의 텔레메트리 절반은 검증할
  대상이 없다.
- **B 는 아무 때나.** A·C 와 공유 파일이 없다 — `internal/web` 안에서도 B 는 라우트·내비·신규
  templ 을, C 는 `screens.templ`/`viewmodel_ops.go` 를 만진다. 겹치는 파일은 `i18n.js` 와
  `app.go` 뿐이고 둘 다 추가만 한다. 다만 **같은 워크트리에서 동시에 편집하지 않는다**(카드 하나
  = 워크트리 하나 규율).
- **C 마지막.** M5 가 A 의 산출과 M1/M2 의 산출을 함께 소비하는 유일한 지점이다.

각 SPEC 은 단독 출하 가능하다: A 는 콘솔이 아무것도 읽지 않아도 statusline·`moai tokens`·문서가
정합하고, B 는 `/todo` 하나로 완결되며, C 는 A 착지 후 레코드 확장과 뷰를 함께 닫는다.

## 7. 재감사

수정을 만들지 않은 감사관이 돌린다(2차 감사관 자신의 권고이자 델타 스코핑 교훈). 채택된 규율:

> iteration 2+ 델타 재감사 PASS 는 "델타가 닫혔다"만 증명하며 SPEC 절대 품질을 증명하지 않는다.
> Tier L kickoff 전에는 전면 감사 1회를 별도로 받는다.

세 SPEC 은 각각 전면 감사를 받는다. C 는 iteration 3(최대 3회 중 마지막)이 되므로, C 편집 전에
A·B 의 감사 결과를 먼저 읽고 공통 결함을 C 에 선반영한다.

## 8. 리드 비준이 필요한 항목

| # | 결정 | 제안 |
|---|---|---|
| D-1 | 모델·effort 를 세션 텔레메트리 스냅샷에 동승 (§2.5) | 채택. 온디스크 경로명 유지, Go 타입명만 확장, `schema_version` 상향 |
| D-2 | 비-git 폴백 분기의 콘솔 렌더 (§4.1) | 읽기 관통 — 파일이 없는 폴백 루트 대신 프로젝트 로컬 경로를 반환(무변경) |
| D-3 | 신규 SPEC id 두 개 | `SPEC-SESSION-TELEMETRY-001`, `SPEC-WEB-TODO-QUEUE-001` |
| D-4 | C 의 Tier 유지(L) / A 의 Tier(L) / B 의 Tier(M) | 위 근거대로 |

---

## Evidence

**Claim** — (1) 런처는 두 백엔드 모두에서 세션 모델을 알지 못한다. (2) statusline 입력은 세션별
모델·effort 를 이미 싣는다. (3) 옮길 경로의 소비자는 독트린 4 + docs-site 12 파일이다.
(4) `"raw_pct"` 중복 선언은 4파일 중 2파일이 statusline 밖이다.

**Evidence**

```
$ grep -rn '"\-m"\|"--model"\|Model' internal/cli/cc.go
(no output)

$ grep -n "ANTHROPIC_DEFAULT\|--model\|EnvAnthropic" internal/cli/cc.go
36:  -m, --model <model>           Override model selection

$ sed -n '350,353p' internal/cli/glm.go
	_ = os.Setenv(config.EnvAnthropicDefaultOpusModel, glmConfig.Models.High)
	_ = os.Setenv(config.EnvAnthropicDefaultSonnetModel, glmConfig.Models.Medium)
	_ = os.Setenv(config.EnvAnthropicDefaultHaikuModel, glmConfig.Models.Low)
	_ = os.Setenv(config.EnvAnthropicDefaultFableModel, glmConfig.Models.Fable)

$ grep -n "Effort\|DisplayName" internal/statusline/types.go
69:	Effort         *EffortInfo        `json:"effort"`
131:type EffortInfo struct { Level string `json:"level"` }
148:	DisplayName string `json:"display_name"`

$ sed -n '286,287p' internal/statusline/builder.go
	if input != nil && input.Effort != nil {
		data.Effort = input.Effort

$ sed -n '51p' internal/statusline/metrics.go
	modelName = resolveGLMModelName(modelName)

$ grep -rln "context-usage.json" docs-site/content | wc -l
12

$ grep -rln "state/context-usage.json" .claude internal/template/templates
.claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md

$ grep -rn '"raw_pct"' internal/ | sed 's/:.*//' | sort | uniq -c
   1 internal/cli/tokens.go
   1 internal/cli/tokens_test.go
   1 internal/statusline/context_usage.go
   1 internal/statusline/context_usage_test.go
```

**Baseline-attribution** — 전부 이 실행에서, 이 트리(`WT-web-live-todo` @ `52358e72f`)에 대해
측정했다. SPEC 본문이 인용한 수치를 재사용하지 않았다.

**Gaps** — (1) 코드를 한 줄도 바꾸지 않았다. D-1 의 실제 구현 가능성(레코드에 두 필드를 얹었을 때
`sameSemanticPayload` 쓰기 스로틀이 매 렌더 쓰기로 퇴화하는지)은 **미검증**이다 — 모델·effort 는
세션 내내 대체로 불변이라 스로틀을 깨지 않을 것으로 보이나 측정하지 않았다. A 의 plan 에서
확인할 항목으로 넘긴다. (2) B 의 파일 수 12-14 는 열거 추정이며 실측 diff 가 아니다.
(3) 각 SPEC 의 AC 개수는 배정 설계값이고, 본문 작성 후 재계수해야 확정된다.
(4) Claude Code 가 `effort` 를 실제로 보내는지는 타입 선언으로만 확인했고 런타임 입력을 관측하지
않았다 — 안 보내면 A-002 의 "없으면 비운다" 경로로 떨어지므로 설계는 성립하나, 그 경우 effort 셀은
상시 "미기록"이 된다.

**Residual-risk** — (1) D-1 이 기각되면 §2 전체가 무효가 되고 F1 은 C 안에서 "런처가 알 수 있는
근사값을 기록"하는 형태로만 닫힌다 — 필드의 존재 이유(정직성)를 훼손하므로 권하지 않는다.
(2) 분할이 세 번의 plan-audit 을 만든다. B 는 Tier M(임계 0.80, 3 아티팩트)이라 비용이 낮지만
A 는 Tier L 이라 research/design 을 새로 쓴다. (3) A 와 C 를 다른 라운드에 태우면 콘솔의
텔레메트리 셀은 A 착지 후 C 착지 전까지 계속 "미기록"으로 보인다 — 기능 퇴행은 아니나 관측자에게는
"고쳐지지 않은 것"으로 보인다.

---

## 부록 — effort 런타임 관측 (리드 [HARD] 지시, 본문 작성 전 선행 측정)

**결론: Claude Code 는 `effort` 를 실제로 보낸다.** A-002 의 "없으면 비운다"는 예외 경로로
남고, REQ 문장은 "가용할 때 기록"이 아니라 **"기록한다"**로 쓸 수 있다.

측정 방법: 이 머신의 statusline 래퍼(`.moai/status_line.sh` — gitignore 대상 런타임 생성물)에
stdin 을 스크래치패드로 복사하는 한 줄을 **일시 삽입**하고, 렌더 1회를 받은 뒤 원본으로
복원했다. 워크트리본과 primary 본 두 벌 모두 백업→패치→복원했고, 복원 후 `diff -q` 두 건이
공집합임을 확인했다(`RESTORED both`). 설정 파일은 건드리지 않았다.

관측된 페이로드(2026-08-24 13:02, Claude Code **2.1.241**, 세션 `d281730e…` = `lead`):

```json
{
  "session_id": "d281730e-a47e-4f82-878e-5fd0ddc4dcb9",
  "effort":  { "level": "medium" },
  "model":   { "id": "claude-opus-5[1m]", "display_name": "Opus 5 (1M context)" },
  "version": "2.1.241",
  "context_window": { "used_percentage": 54, "context_window_size": 1000000 }
}
```

세 값 모두 **한 페이로드 안에 세션 단위로** 들어온다 — `session_id`(A-001 의 경로 키),
`effort.level`, `model`. §2 의 생산자 재지목이 런타임 관측으로 확인됐다.

참고로 `~/.moai/cache/statusline_debug.log` 의 2026-05-20 기록(CC 2.1.145)에도
`"effort": {"level": "xhigh"}` 가 있다 — 세 달 전 버전에서도 보내고 있었다는 역사적 데이터
포인트이나, 위 판단의 근거는 이번 실행의 관측이다.

### D-5 — 모델 값으로 무엇을 기록하는가 (신규, 비준 필요)

관측이 새 갈래를 하나 드러냈다: 페이로드는 `id`(`claude-opus-5[1m]`)와
`display_name`(`Opus 5 (1M context)`)을 **둘 다** 보낸다.

**제안: GLM 해석을 거친 `display_name` 한 값만 기록한다.** `resolveGLMModelName`
(`metrics.go:197`)이 이미 `display_name` 을 입력으로 받아 GLM 환경에서 z.ai 모델명으로 바꾸므로,
그 함수의 출력이 곧 "이 세션이 실제로 도는 모델"이다. *기각:* `id` 기록 — Claude 형태의 식별자라
GLM 세션에서는 그 자리에 넣을 값이 애초에 없고(`ANTHROPIC_DEFAULT_*_MODEL` 로 들어온 이름이
실체다), `[1m]` 접미도 따로 벗겨야 한다. *기각:* 두 값 모두 기록 — 콘솔 셀 하나에 값 둘은
소비자가 다시 고르게 만든다.

### 잔여 Gap

GLM 백엔드 세션의 페이로드는 관측하지 못했다(현재 이 머신에 GLM 세션 없음). `effort` 가 GLM
경유에서도 오는지는 **미관측**이며, 안 오면 GLM 세션의 effort 셀만 미기록이 된다 — A-002 의
빈 경로로 정상 처리되므로 설계는 성립한다. A 의 plan 확인 항목에 넣는다.

---

## 부록 2 — 세션 id 세 값의 정체, 그리고 그것이 드러낸 조인 결함 (리드 지시 선행 측정)

**결론 먼저.** 세 값은 **세 개의 id 체계가 아니라 한 체계(Claude Code 의 `session_id`)를 담은 세
개의 슬롯**이다. 그런데 그 확인 과정에서 **A-001 보다 더 아래를 때리는 결함**이 나왔다:
`kanban.Record` 는 그 세션 자신의 id 로 키잉되지 않는다. C 의 조인이 오늘 데이터 위에서 닫히지
않는다.

### 2-1. 세 값이 왜 다른가

| 슬롯 | 값 | 무엇을 담는가 |
|---|---|---|
| statusline 페이로드 `session_id` | `d281730e` (lead) | **렌더한 세션 자신의 id.** 구조상 항상 정확 |
| `moai session current` | `3db058e1` (**t207 = 저**) | `.moai/state/current-session-id.txt` — **프로젝트당 파일 하나**, SessionStart 가 마지막에 쓴 세션이 이긴다 |
| `.moai/state/goal/<id>.json` | `20d94c75` | **goal 을 arm 한 시점의 세션 id.** 그 뒤 `/clear` 로 세션이 바뀌면 파일명은 옛 id 그대로 남는다 |

두 계층 모두 **같은 출처**에서 id 를 받는다 — SessionStart 훅은 `input.SessionID`
(`internal/hook/session_start.go:299-314`), statusline 은 stdin 페이로드의 같은 필드. 새 id 를
발급하는 계층은 없다.

리드가 `moai session current` 로 받은 `3db058e1` 이 **제 세션 id** 인 것이 그 증거다: 사이드카는
12:23 에 제 SessionStart 가 마지막으로 썼다.

```
$ cat .moai/state/current-session-id.txt      → 3db058e1-…   (12:23 작성 = t207 세션)
internal/session/registry.go:52   const CurrentSideChannelFile = ".moai/state/current-session-id.txt"
```

**즉 사이드카는 context-usage.json 과 정확히 같은 결함 계열이다** — 단일 슬롯, 마지막 기록자 승리.
A 가 고치는 그 형태가 세션 id 표면에도 하나 더 있다. **A 의 범위는 아니다**(A 는 페이로드 id 로
키잉하며 사이드카를 읽지 않는다). 별도 카드 후보로 남긴다: 다중 세션에서 `moai session current`
가 남의 id 를 답하면 핸드오프의 `source_session_id` 귀속이 틀어진다.

### 2-2. 그런데 `kanban.Record` 는 그 세션의 id 로 키잉되지 않는다 — **C 블로킹**

```
internal/cli/kanban.go:472-478        recordKanbanSession → resolveLaunchSessionID("")
internal/cli/launcher_blockcap_infinite.go:126-134   resolveLaunchSessionID → resolveCurrentSessionID()
                                                     → 위의 단일 슬롯 사이드카
```

런처(`moai cc` / `moai glm`)는 **새 claude 세션이 생기기 전에** 돈다. 그 시점에 사이드카가 담고
있는 것은 **런처를 띄운 부모 세션의 id**다. 그러므로 레코드는 태어날 세션이 아니라 **띄운
세션의 id 로 파일명이 붙는다.**

실측이 이것을 그대로 보여준다:

```
$ python3 …  .moai/state/active-sessions.json   (살아 있는 세션 3개, CC 실제 id)
2beac221…  pid 15207  cwd …/worktrees/t219
c15d8434…  pid 51045  cwd …/worktrees/t210
3db058e1…  pid 36912  cwd …/worktrees/t207

$ ls .moai/state/kanban/{2beac221…,c15d8434…,3db058e1…}.json
No such file or directory   ×3      ← 살아 있는 세 세션 모두 **자기 레코드가 없다**

$ cat .moai/state/kanban/d281730e-….json      ← 유일하게 맞은 id = 리드
{ "session_id": "d281730e-…", "role": "lane", "backend": "claude",
  "entered_at": "2026-08-23T17:47:22Z" }
```

마지막 줄이 결정적이다. `d281730e` 는 **리드 세션**인데 그 레코드의 `role` 은 **`lane`** 이다.
리드가 레인을 띄울 때, 런처가 레인의 레코드를 **리드의 id 로** 찍은 것이다. 한 번의 실수가 양쪽을
동시에 망친다 — 레인은 자기 레코드가 없고, 리드의 id 에는 레인 레코드가 붙는다.

**따라서 다음이 오늘 데이터 위에서 닫히지 않는다:**

- REQ-WC15-043 의 레인 조인 `workers.json[lane].PID → active-sessions.session_id → kanban.Record`
  — 세션 등록부는 CC 실제 id 를 담고 레코드는 부모 id 로 파일명이 붙으므로, 조회는 **빈손이거나
  남의 레코드**를 집는다. 스펙 §A.5 는 이 체인이 "오늘 데이터로 닫힌다"고 적었는데, 측정하면
  **키가 어긋나 있다**.
- C-007/C-008 의 레코드 ↔ 텔레메트리 스냅샷 조인 — 스냅샷은 CC 실제 id 로 키잉되고 레코드는
  아니므로 같은 이유로 어긋난다.
- 이미 배포된 체인 뷰의 행 귀속도 같은 이유로 어긋나 있을 것이다(콘솔 관측은 하지 않았다).

### 2-3. 제안 — **SPEC D 를 하나 더 세운다** (리드 비준 필요)

**D-6. `SPEC-KANBAN-RECORD-SESSION-KEY-001` (신규, Tier M), C 의 선행 조건.**

고치는 자리는 런처가 아니다 — 런처는 자식의 id 를 **알 수 없다**(자식이 아직 없다). 자식의 id 를
아는 첫 지점은 그 세션의 **SessionStart 훅**이고, 그 훅은 이미 `input.SessionID` 를 받으며 런처가
export 한 `MOAI_KANBAN_*` 환경변수도 그 프로세스 환경에 있다. 즉 **레코드 쓰기를 런처에서
SessionStart 로 옮기면** 키가 자동으로 맞고, C 가 필요로 하던 lane 번호·card id 도 같은 환경에서
같이 실린다.

- 범위: `internal/cli/kanban.go`(+`cc.go`/`glm.go` 호출부), `internal/hook/session_start.go`,
  `internal/kanban/record.go`, 테스트 — 6-8 파일, Tier M.
- 강한 acceptance 가 나온다: "레인 세션의 레코드가 **그 세션 자신의** id 로 존재하고 `role` 이
  그 세션의 역할과 일치한다" — 오늘 트리에서 **측정 가능하게 거짓**이므로 baseline 이 선다.
- C 는 순수 소비자 SPEC 이 된다(M1/M2 의 생산자 몫이 D 로 이동). plan.md §C 이유 1("생산자만
  떼면 acceptance 가 약해진다")은 여기 적용되지 않는다 — D 의 acceptance 는 왕복 테스트가 아니라
  **어긋난 귀속의 교정**이고, 오늘 거짓인 명제를 참으로 만드는 것이라 관측이 성립한다.

의존 그래프가 이렇게 된다:

```
A ─┐
D ─┴──► C        B (독립)
```

A 와 D 와 B 는 서로 독립이라 순서는 자유롭고, C 만 A·D 뒤에 온다.

### 2-4. A-001 은 어떻게 쓰는가 (측정 결과 반영)

- **정본 id = statusline 페이로드의 `session_id`** — CC 가 그 세션에 준 값이며, 구조상 렌더한
  세션 자신의 것이다. A 는 이 값으로 파일을 키잉한다.
- **사이드카는 읽지 않는다.** A 는 `moai session current` 에 의존하지 않는다 — 그 표면의 단일 슬롯
  결함을 A 가 상속하지 않기 위해서다.
- **소비자가 id 를 얻는 경로는 세션 등록부**(`active-sessions.json`, SessionStart 가 같은 CC id 로
  기록)다. `kanban.Record` 경유는 **D 착지 전까지 쓸 수 없다** — 그것이 §2-2 의 내용이다.
- 그러므로 A-001 에는 "세션 자신에게 전달된 식별자로 키잉한다"는 문장이 들어가고, 그 값의 출처가
  런처가 아니라 세션 런타임임을 명시한다.

### 잔여 Gap

콘솔이 실제로 잘못된 행을 그리는지는 **관측하지 않았다**(`moai web` 미기동). §2-2 는 온디스크
데이터로 조인이 어긋난다는 것까지만 증명하며, 렌더 결과는 추론이다. D 의 plan 에서 확인 항목으로
둔다.
