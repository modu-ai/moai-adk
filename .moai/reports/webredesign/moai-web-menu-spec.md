# MoAI Web Console — 메뉴·기능 정의서

> `moai web`을 **설정 편집기**에서 **moai-adk 운영 콘솔**로 확장하기 위한 기능 정의서입니다.
> 짝 문서: `moai-web-redesign-brief.md` (시각 재설계 지침). 이 문서는 *무엇을 담을지*, 짝 문서는 *어떻게 보일지*를 다룹니다.

작성 기준: 2026-08-14 / 브랜치 `main` / 조사 대상 리비전에서 코드를 직접 읽어 작성

---

## 0. MoAI-Kanban — 이 콘솔이 비춰야 할 대상

`moai cc -k` / `moai glm -k`로 진입하는 **MoAI-Kanban 모드**는 하나의 작업을 5개 터미널 세션에 나눠 태우는 개발 방법론입니다. 세션 하나가 전 과정을 끌고 가는 대신, 역할마다 세션을 따로 열고 세션끼리 메시지로 넘깁니다.

```
  0. Lead ─┬─▶ 1. Plan ──▶ 2. Run ──▶ 3. Review ──▶ 4. Sync
           │      계획         구현        코드리뷰        문서동기화+커밋
           └────────────── 완료 증거를 읽고 다음 단계로 지시 ──────────────┘
```

### 왜 이렇게 나누는가 — 토큰 경제성

역할마다 요구되는 추론 깊이가 다릅니다. 그래서 **역할별로 모델과 백엔드를 다르게 태웁니다.**

| # | 역할 | 모델 | effort | 과금 |
|---|---|---|---|---|
| 0 | Lead | Opus 5 (1M) | medium | Claude Max (정액) |
| 1 | Plan | Opus 5 (1M) | high | Claude Max (정액) |
| 2 | Run | GLM-5.2 (1M) | xhigh (ultracode) | API 종량 |
| 3 | Review | Opus 5 (1M) | medium | Claude Max (정액) |
| 4 | Sync | GLM-5.2 (1M) | high | API 종량 |

설계·판단이 걸린 자리(Plan·Review)는 Opus에 두고, 분량이 많고 기계적인 자리(Run·Sync)는 GLM으로 내립니다. 그리고 세션을 나누는 것 자체가 절약입니다 — 한 세션이 계획부터 커밋까지 다 하면 컨텍스트가 계속 불어나 매 턴 전체를 다시 읽지만, 역할을 나누면 각 세션이 **자기 단계에 필요한 만큼만** 들고 갑니다.

### 콘솔이 답해야 할 질문

이 방법론이 돌아가는 동안 운영자가 알고 싶은 것은 설정값이 아니라 **체인의 현재 상태**입니다.

1. 지금 카드가 어느 역할에 있는가? 누가 일하고 누가 노는가?
2. 각 세션이 어떤 모델·백엔드로 돌고 있고, 컨텍스트를 얼마나 썼는가?
3. 어디서 막혔는가 — 다음 역할 세션이 안 떠 있진 않은가?
4. 이번 체인이 종량 과금(GLM)을 얼마나 태웠는가?

**따라서 칸반 화면의 1순위는 SPEC 파이프라인이 아니라 "5개 세션의 살아있는 상태"입니다.** §4.2가 이 방향으로 정의됩니다.

> 방법론의 규율(완료는 증거로 판정, 단계마다 `/clear` 요청, 세션은 다른 세션을 띄울 수 없음)은
> `.claude/rules/moai/workflow/kanban-dispatch.md`가 정본입니다. 콘솔은 그것을 **비추기만** 하고 지시하지 않습니다.

---

## 1. 확정된 설계 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 칸반 카드는 **SPEC에서 유도**한다 (전용 보드 저장소를 만들지 않음) | 상태의 단일 진실을 SPEC frontmatter 한 곳에 유지. 보드가 별도 상태를 들면 frontmatter와 불일치가 생기고, 그 불일치를 조정하는 코드가 이 도구의 가치보다 커진다 |
| D2 | 실시간 갱신은 **SSE(Server-Sent Events)** | 모니터링은 전부 서버→브라우저 단방향. Go 표준 라이브러리만으로 되고 외부 의존성이 0이며, 현재 오프라인 불변식(네트워크 요청 0)을 깨지 않는다 |
| D3 | 이번 산출물은 **메뉴·기능 정의서까지** | 메뉴 정의가 흔들리면 SPEC과 화면 시안이 함께 흔들린다. 정의 확정 후 SPEC → 구현 순서 |

---

## 2. 데이터 층 원칙 — "MCP 연동"의 실체

### 2.1 웹과 MCP는 형제다

`moai mcp-server`가 노출하는 17개 도구는 전부 **내부 Go 패키지의 얇은 래퍼**입니다. `internal/cli/mcp_server.go`의 도구 설명이 그대로 말해줍니다.

| MCP 도구 | 감싸는 내부 함수 |
|---|---|
| `session_list` | `session.QueryActiveWork` |
| `goal_status` / `goal_arm` | `goal.LoadGoal` / `goal.NewGoal`+`SaveGoal` |
| `spec_progress` | `spec.ListDocs` |
| `verify_snapshot` / `verify_trend` | `verify.Load` (+ `verify.RecordCheck`) |
| `spec_audit` / `spec_drift` | `spec.Audit` |

`moai web`은 같은 바이너리 안의 Go 코드이므로, **MCP 프로토콜을 거칠 이유가 없습니다.** 이미 SPEC 보드(`internal/web/board.go:89`)가 `spec.ListDocs` + `spec.Audit`을 직접 호출하고 있고, 이것이 정답 패턴입니다.

```
                    ┌─────────────────┐
   Claude Code ───▶ │  moai mcp-server │ ─┐
                    └─────────────────┘  │   ┌──────────────────┐
                                          ├──▶│ internal/spec    │
                    ┌─────────────────┐  │   │ internal/goal    │  ← 단일 진실
   브라우저    ───▶ │  moai web        │ ─┘   │ internal/session │
                    └─────────────────┘      │ internal/verify  │
                                              └──────────────────┘
```

**따라서 "moai-mcp와 실시간 동기화"의 구현은 이렇게 나뉩니다.**

- **동기화**: 별도 작업이 없습니다. 같은 패키지·같은 파일을 읽으면 이미 동기화된 상태입니다.
- **실시간**: 파일이 바뀐 것을 **감지해서 브라우저에 알리는 채널**이 필요합니다. 이것만이 신규 작업입니다.

### 2.2 읽기 전용 원칙

콘솔의 설정 탭은 쓰기(저장)를 하지만, **모니터링 영역(SPEC·칸반·모니터)은 전부 읽기 전용**으로 둡니다.

이유: SPEC 상태 전이는 소유권 매트릭스가 정해져 있습니다(`internal/spec/lint_ownership.go:64`). `draft→in-progress`는 manager-develop, `in-progress→implemented`는 manager-docs, `implemented→completed`는 오케스트레이터의 Mx 커밋이 소유합니다. 웹에서 상태를 손으로 바꾸면 이 소유권 규율이 무너지고, 감사 로그에 소유자 없는 전이가 남습니다.

기존 SPEC 보드도 이미 GET 외 메서드를 405로 거부합니다(`app.go:154` 주석, REQ-WC11-044/045/046).

---

## 3. 메뉴 구조

```
┌─ 개요        Overview     지금 무슨 일이 벌어지는지 한 화면        [신규]
├─ 칸반        Kanban       SPEC 유도 실행 보드                     [신규]
├─ SPEC        Specs        목록 · 상태 · 드리프트 · 상세            [기존 확장]
├─ 모니터      Monitor      세션 · 목표 · 검증 · 에픽                [신규]
└─ 설정        Settings     현행 10개 탭                            [기존]
```

설계 의도: **왼쪽으로 갈수록 "보는 것", 오른쪽으로 갈수록 "바꾸는 것"** 입니다. 세션을 열면 대개 개요에서 시작해 필요한 곳으로 들어가고, 설정은 가끔 들릅니다.

---

## 4. 화면별 정의

각 화면마다 다음 5가지를 정의합니다: **목적 / 데이터 원천 / 표시 항목 / 상호작용 / 실시간**

---

### 4.1 개요 (Overview) — `GET /`

**목적**: 세션을 시작할 때 "지금 이 프로젝트가 어떤 상태인가"를 한 화면에서 파악한다.

**데이터 원천**

| 블록 | 원천 | 비용 |
|---|---|---|
| SPEC 상태 분포 | `spec.ListDocs` | 파일 스캔 |
| 드리프트 요약 | `spec.Audit` | 파일 스캔 |
| 활성 세션 | `session.QueryActiveWork` (`.moai/state/active-sessions.json`) | 작음 |
| 무장된 목표 | `.moai/state/goal/<session>.json` | 작음 |
| 검증 최근 결과 | `verify.Load` (`.moai/state/verify/`) | 작음 |
| 컨텍스트 사용률 | `.moai/state/context-usage.json` | 작음 |

**표시 항목**

- 상단 지표 4개: 활성 SPEC 수 / 드리프트 건수 / 활성 세션 수 / 마지막 검증 결과
- 진행 중인 SPEC 카드 (in-progress 상태) — ID, 제목, Tier, 마지막 갱신
- 활성 세션 목록 — 세션 ID 앞 8자, 작업 중인 SPEC, 백엔드(claude/glm), 진입 시각
- 주의가 필요한 것 — MUST-FIX 드리프트, 실패한 검증, 정체된 목표
- **칸반 체인 요약** — 진행 중인 체인이 있으면 5개 역할의 기동 여부를 한 줄로 압축해 보여주고, 미기동 역할이 있으면 경고합니다 (상세는 칸반 화면). ※ §4.6 선행 필요

**상호작용**: 클릭하면 해당 상세 화면으로. 이 화면 자체는 쓰기 없음.

**실시간**: 예 — 전 블록.

**신규 구현**: 뷰모델 + 템플릿. 데이터 원천은 전부 기존.

---

### 4.2 칸반 (Kanban) — `GET /kanban`

**목적**: 진행 중인 칸반 체인이 지금 어느 역할에 있고, 각 세션이 어떤 조건으로 돌고 있는지 본다.

화면을 **두 개의 뷰**로 나눕니다. 위가 주(主), 아래가 보조입니다.

---

#### 뷰 A — 체인 세션 보드 (주)

MoAI-Kanban의 5개 역할을 **가로 5열**로 놓고, 각 열에 그 역할 세션의 살아있는 상태를 얹습니다. 화면이 §0의 다이어그램과 같은 모양이 되도록 합니다.

```
┌─ Lead ───┬─ Plan ───┬─ Run ────┬─ Review ─┬─ Sync ───┐
│ ● 활성   │ ● 활성   │ ● 활성   │ ○ 미기동 │ ● 활성   │
│ Opus 5   │ Opus 5   │ GLM-5.2  │   —      │ GLM-5.2  │
│ medium   │ high     │ xhigh    │   —      │ high     │
│ CW 26%   │ CW 14%   │ CW 41%   │   —      │ CW 8%    │
│          │ ✅ 완료  │ ▶ 작업중 │ ⏸ 대기   │ ⬜ 대기  │
└──────────┴──────────┴──────────┴──────────┴──────────┘
              현재 카드: SPEC-XXX-001
```

열마다 표시할 것:

| 항목 | 의미 |
|---|---|
| 기동 여부 | 세션이 떠 있는가 (미기동이면 체인이 그 자리에서 멈춥니다) |
| 세션 라벨 | `plan-tjq3pi` 형태 |
| 모델·effort | `Opus 5 / medium`, `GLM-5.2 / xhigh` |
| 백엔드 | claude(정액) / glm(종량) — 배지로 구분 |
| 컨텍스트 사용률 | `/clear` 시점 판단용 |
| 단계 상태 | 완료 / 작업중 / 대기 / 막힘 |
| 마지막 활동 | 하트비트 경과 시간 |

**미기동 역할을 눈에 띄게 표시하는 것이 이 뷰의 핵심 가치입니다.** `kanban-dispatch.md`가 못 박은 것처럼, 카드가 갈 곳의 세션이 없으면 그건 대기가 아니라 **결함**이고, 조용히 멈춰 있는 것이 가장 진단하기 비싼 실패 모양입니다.

**⚠️ 이 뷰에 필요한 데이터 중 상당수가 지금 기록되지 않습니다.** §4.6을 반드시 함께 읽으세요.

---

#### 뷰 B — SPEC 파이프라인 (보조)

체인이 아니라 **카탈로그 전체**를 봅니다. SPEC들이 생애주기 어디에 몰려 있는지, 병목이 어느 단계인지.

**데이터 원천**: `spec.ListDocs`의 frontmatter `status` 하나로 컬럼을 계산합니다. 보조로 `spec.Audit`의 드리프트를 카드 배지로 얹습니다.

#### ⚠️ 컬럼 매핑 — 6컬럼 중 4개만 자동으로 채워집니다

SPEC 생애주기는 `(none) → draft → in-progress → implemented → completed`이고, 종단 상태로 `superseded / archived / rejected`가 있습니다 (`internal/spec/lint_ownership.go:64` 전이 매트릭스).

| 칸반 컬럼 | SPEC status | 자동 |
|---|---|---|
| backlog | — | ❌ 대응 status 없음 |
| plan | `draft` (별칭 `planned`) | ✅ |
| run | `in-progress` | ✅ |
| review | — | ❌ 대응 status 없음 |
| sync | `implemented` | ✅ |
| done | `completed` | ✅ |
| (보드 밖) | `superseded` / `archived` / `rejected` | ✅ 필터로만 |

**빈 두 컬럼을 어떻게 할지가 이 화면의 유일한 미결 사항입니다.** 세 가지 길이 있습니다.

- **(a) 4컬럼으로 축소** — `plan · run · sync · done`. 가장 정직하고 구현이 0에 가깝지만, 칸반 모드의 6컬럼 어휘(`kanban-dispatch.md`)와 이름이 어긋납니다.
- **(b) 6컬럼 유지 + 두 칸을 비워둠** — 어휘는 맞지만 늘 비어 있는 칸이 두 개 보입니다.
- **(c) 6컬럼 + 보조 신호로 채움** — `backlog`는 `/moai todo` 큐에서, `review`는 열린 PR에서. 다만 todo 큐는 Go 구현이 없고(스킬이 파일을 직접 다룸), PR 조회는 `gh` 호출이 필요해 SPEC 보드의 "git 비의존" 원칙을 깹니다.

권장은 **(a) 4컬럼**입니다. 뷰 A가 이미 `plan → run → review → sync` 역할 체인을 그대로 보여주므로, 뷰 B까지 6컬럼을 흉내 낼 이유가 없습니다. 뷰 B는 SPEC status가 정직하게 말해주는 4단계만 그리고, 역할 흐름은 뷰 A에 맡깁니다.

**카드에 담을 것**: SPEC ID, 제목, Tier(S/M/L), 마지막 갱신일, 드리프트 배지, 담당 세션(있으면).

**상호작용**: 카드 클릭 → SPEC 상세. **드래그 이동은 넣지 않습니다** (§2.2 읽기 전용 원칙).

**실시간**: 예 — 뷰 A는 세션 상태가 바뀔 때마다, 뷰 B는 SPEC 파일이 바뀔 때마다.

**신규 구현**
- 뷰 B: 컬럼 계산 로직 + 뷰모델 + 템플릿. 저장소 없음. **지금 데이터로 바로 됩니다.**
- 뷰 A: 뷰모델 + 템플릿에 더해 **생산자 쪽 작업이 선행**되어야 합니다 (§4.6).

---

### 4.3 SPEC — `GET /specs`, `GET /specs/{id}`

**목적**: SPEC 카탈로그를 훑고, 한 건을 깊이 본다.

**데이터 원천**: `spec.ListDocs` + `spec.Audit`. 기존 `board.go`가 이미 하는 일입니다.

**목록 화면** (기존 개선)
- 상태·Tier·시대(era)로 거르기, ID/제목 검색
- 정렬: 갱신순 / ID순 / 상태순
- 드리프트 심각도 표시 (MUST-FIX 우선)

**상세 화면** (신규)
- frontmatter 전체 (status, tier, era, 관련 SHA)
- 문서 4종 링크 — `spec.md` / `plan.md` / `acceptance.md` / `progress.md`
- 해당 SPEC의 드리프트 findings
- 해당 SPEC의 검증 스냅샷 이력 (`verify.Load`)

**상호작용**: 읽기 전용. 파일 열기 링크는 경로 표시까지만 (편집기 실행은 범위 밖).

**실시간**: 예.

**신규 구현**: 상세 라우트 + 필터/검색. 목록은 기존 확장.

---

### 4.4 모니터 (Monitor) — `GET /monitor`

**목적**: 실행 중인 것들의 상태를 본다. 개요보다 깊고, SPEC보다 넓다.

네 개 패널로 나눕니다.

| 패널 | 데이터 원천 | 표시 |
|---|---|---|
| 세션 | `session.QueryActiveWork` | 세션 ID, cwd, SPEC, 백엔드, 진입시각, 하트비트 |
| 목표 | `.moai/state/goal/*.json` | 조건, 진행 턴 수, 정체 여부, 판정문(verdict) |
| 검증 | `verify.Load` / `verify_trend` | 키별 최근 결과와 추이 |
| 에픽 | `moai epic status` (`internal/cli/epic.go:55`) | 접두사별 마일스톤 진척 |

**⚠️ 세션 레지스트리 주의**: 레지스트리에는 이미 죽은 PID의 항목이 남을 수 있습니다(알려진 성질). 화면에서는 "활성"이라고 단정하지 말고, 하트비트 시각을 함께 보여주거나 PID 생존을 확인한 뒤 표시하세요. 확인 없이 "N개 세션 활성"이라고 쓰면 거짓 정보가 됩니다.

**칸반 세션 기록**: `.moai/state/kanban/<session>.json`은 `{session_id, spec_id, backend, entered_at, verify_rung, verify_reentries}` 형태입니다(`internal/kanban/record.go:55`). **보드 상태가 아니라 세션 진입 기록**이므로, 칸반 화면이 아니라 이 세션 패널에 붙는 것이 맞습니다.

**상호작용**: 읽기 전용.

**실시간**: 예 — 이 화면이 실시간의 주 수요처입니다.

**신규 구현**: 4개 패널 전부 신규. 다만 데이터 원천은 모두 기존 패키지.

---

### 4.5 설정 (Settings) — `GET /settings`

**목적**: 현행 그대로 — 프로필 환경설정 + 프로젝트 설정 편집.

**변경점**: 라우트만 `/`에서 `/settings`로 옮기고(개요가 `/`를 차지), **탭 구성·필드·폼 계약은 손대지 않습니다.**

기존 10개 탭 유지: `Identity · Language · LLM · 3rd Party LLM · Workflow · Git & Worktree · Audit · Agents · Report · MCP`

**실시간**: 아니오 — 편집 중 화면이 밑에서 바뀌면 안 됩니다. 다만 외부에서 설정 파일이 바뀌면 **배너로 알리기만** 하고 자동 갱신하지 않습니다.

**신규 구현**: 라우트 이동 + 폼 계약 보존(`name` 속성·action·POST 경로 그대로).

---

### 4.6 ⚠️ 데이터 공백 — 지금은 기록되지 않는 것들

§4.2 뷰 A(체인 세션 보드)를 만들려면 **먼저 생산자 쪽을 고쳐야 합니다.** 앞선 §2에서 "데이터 원천은 전부 기존"이라고 적었지만, 칸반 체인 모니터에 한해서는 그 말이 성립하지 않습니다. 파일을 직접 열어 확인한 결과는 다음과 같습니다.

#### (1) 역할이 어디에도 저장되지 않습니다

역할 라벨(`plan-tjq3pi`)은 실행 시점에 파싱되지만(`internal/cli/cc.go:115`, `glm.go:184`), 어떤 상태 파일에도 남지 않습니다.

- `internal/kanban/record.go:55`의 `Record` 필드는 `{session_id, spec_id, backend, entered_at, deepscan_dir, verify_rung, verify_reentries}` — **role 없음**
- `.moai/state/active-sessions.json` 항목은 `{session_id, spec_id, phase, started_at, last_heartbeat, pid, host, cwd}` — **role 없음**

→ 어떤 세션이 Plan이고 어떤 세션이 Run인지 콘솔이 알 방법이 없습니다.

#### (2) 컨텍스트 사용률은 세션당이 아니라 **단일 슬롯**입니다

`.moai/state/context-usage.json`은 프로젝트당 **파일 하나**이고, 상태줄을 그리는 세션이 매번 통째로 덮어씁니다. 실제로 이 문서를 쓰는 동안 눈앞에서 관측했습니다.

```
읽은 시점 A:  session_id 368a2bd9…   tokens_used 260000   raw_pct 26
읽은 시점 B:  session_id e463a3c9…   tokens_used      0   raw_pct  0   ← 다른 세션이 덮어씀
```

파일 자체가 `writer_pid`와 `session_id`를 담고 있고 마지막 기록자가 이깁니다. **5개 세션의 컨텍스트 사용률을 여기서 읽는 것은 불가능합니다** — 가장 최근에 상태줄을 그린 세션 하나만 보입니다.

#### (3) 모델·effort는 아무 데도 저장되지 않습니다

상태줄이 보여주는 `Opus 5 / medium`, `GLM-5.2 / xhigh`는 Claude Code 런타임에서 실시간으로 읽는 값입니다. 디스크에 남지 않으므로 콘솔이 사후에 조회할 수 없습니다. `Record.Backend`(claude/glm)만 유일하게 저장됩니다.

#### 필요한 생산자 작업

| 항목 | 해야 할 일 | 크기 |
|---|---|---|
| 역할 | `kanban.Record`에 `role` 필드 추가 + 진입 시 기록 | 작음 |
| 모델·effort | 같은 레코드에 진입 시점 값 스냅샷 | 작음 |
| 컨텍스트 사용률 | **세션당 파일로 분리** — `.moai/state/context-usage/<session-id>.json` | 중간 (기존 소비자 호환 필요) |
| 단계 상태 | 체인 진행(완료/작업중/대기)을 어디에 남길지 결정 | 중간 — 미결 |

**(3)의 "단계 상태"가 가장 까다롭습니다.** 현재 체인 진행은 리드 세션의 판단으로만 존재하고 디스크에 없습니다. 세 가지 길이 있습니다.

- **(가) SPEC status에서 유추** — `in-progress`면 Run이 일하는 중으로 본다. 추가 작업 0이지만 Review 단계를 구분할 수 없고, SPEC 없는 체인 초기에는 아무것도 못 봅니다.
- **(나) 리드가 전이를 기록** — 카드가 컬럼을 옮길 때 리드 세션이 상태 파일에 한 줄 남긴다. 정확하지만 리드의 규율에 의존합니다(기록을 빠뜨리면 화면이 조용히 낡습니다).
- **(다) 하트비트로 추정** — 최근 활동이 있는 세션을 "작업중"으로 본다. 기록 부담이 0이지만 추정이므로 단정해서 표시하면 안 됩니다.

권장은 **(다)로 시작해 (나)로 올리는 것**입니다. (다)는 지금 있는 `last_heartbeat`만으로 되고, 화면에는 "추정"임이 드러나게 표기합니다. 정확한 전이 기록이 필요해지면 그때 (나)를 얹습니다.

#### 단계 나누기 권고

이 공백 때문에 **칸반 화면은 한 번에 완성되지 않습니다.** 이렇게 나누기를 권합니다.

| 단계 | 내용 | 선행 조건 |
|---|---|---|
| 1단계 | 뷰 B(SPEC 파이프라인) + 개요 + SPEC + 설정 | 없음 — 지금 데이터로 가능 |
| 2단계 | 뷰 A 기본형 (기동 여부 · 역할 · 백엔드 · 하트비트) | `Record.role` 추가 |
| 3단계 | 뷰 A 완성형 (모델·effort·컨텍스트 사용률·비용) | 컨텍스트 파일 세션 분리 + 모델 스냅샷 |

1단계만으로도 콘솔은 충분히 쓸모가 있고, 2·3단계는 생산자 SPEC이 끝난 뒤에 붙습니다.

---

## 5. 실시간 채널 설계

### 5.1 전송 방식

```
GET /events            Content-Type: text/event-stream
```

Go 표준 라이브러리 `http.Flusher`만 사용합니다. 외부 의존성 0.

### 5.2 이벤트 종류

| 이벤트 | 발생 조건 | 소비 화면 |
|---|---|---|
| `spec` | `.moai/specs/**/spec.md` 변경 | 개요 · 칸반 · SPEC |
| `session` | `.moai/state/active-sessions.json` 변경 | 개요 · 모니터 |
| `goal` | `.moai/state/goal/*` 변경 | 개요 · 모니터 |
| `verify` | `.moai/state/verify/**` 변경 | 개요 · 모니터 |
| `kanban` | `.moai/state/kanban/*` 변경 | 개요 · 칸반(뷰 A) |
| `config` | `.moai/config/sections/*` 변경 | 설정 (배너만) |

이벤트 본문에는 데이터를 싣지 않고 **"이 영역이 바뀌었다"는 신호만** 보냅니다. 브라우저는 신호를 받으면 해당 조각을 htmx로 다시 가져옵니다. 이렇게 하면 렌더링 진실이 서버 한 곳에만 남습니다.

### 5.3 변경 감지

`fsnotify`로 위 경로들을 감시합니다. 현재 `go.mod`에 **간접 의존으로 이미 들어와 있어**(`go.mod:64`) 직접 의존으로 승격만 하면 됩니다.

주의 두 가지:
- **디바운스 필수** — 저장 한 번에 파일 이벤트가 여러 개 옵니다. 200~300ms 정도로 묶으세요.
- **감시 대상 수 제한** — `.moai/state/verify/`는 이미 105개 키가 있습니다. 개별 파일이 아니라 디렉터리 단위로 감시하세요.

### 5.4 ⚠️ htmx SSE 확장이 없습니다

번들된 `htmx.min.js`는 **2.0.4 코어이고 SSE 확장이 포함돼 있지 않습니다** (`EventSource` 문자열 0건, 직접 확인). CDN 금지 제약 때문에 확장을 받아올 수도 없습니다.

따라서 `hx-ext="sse"` / `sse-swap` 문법은 쓸 수 없고, **`app.js`에 직접 배선**해야 합니다:

```js
const es = new EventSource("/events");
es.addEventListener("spec", () => {
  document.querySelectorAll('[data-live="spec"]').forEach(el => htmx.trigger(el, "refresh"));
});
```

각 조각은 `hx-trigger="refresh"` + `hx-get="/fragment/..."`로 자신을 다시 가져옵니다. 확장 없이 코어 htmx만으로 성립합니다.

### 5.5 폴백

`EventSource` 연결이 끊기면 브라우저가 자동 재연결합니다. 재연결이 반복 실패하면 30초 폴링으로 내려가고, 화면에 "실시간 끊김" 표시를 남깁니다. 조용히 멈춰서 낡은 화면을 보여주는 것이 가장 나쁜 실패 모드입니다.

---

## 6. 신규 구현 범위

| 구분 | 파일 | 작업 |
|---|---|---|
| 라우트 | `internal/web/app.go` | `/`(개요) `/kanban` `/monitor` `/settings` `/specs/{id}` `/events` 추가, 기존 `/`→`/settings` 이동 |
| 실시간 | `internal/web/events.go` (신규) | SSE 핸들러 + fsnotify 감시 + 디바운스 |
| 뷰모델 | `internal/web/overview.go` `kanban.go` `monitor.go` (신규) | 각 화면 데이터 조립 |
| 템플릿 | `internal/web/overview.templ` `kanban.templ` `monitor.templ` (신규) | 화면 |
| 템플릿 | `internal/web/board.templ` | SPEC 목록 필터·검색 확장 + 상세 |
| 스크립트 | `internal/web/assets/app.js` | EventSource 배선 |
| 사전 | `internal/web/assets/i18n.js` | 신규 문구를 en/ko/ja/zh **네 곳 모두**에 추가 |
| 의존성 | `go.mod` | `fsnotify` 간접 → 직접 승격 |

**건드리지 않는 것**: `internal/spec` · `internal/goal` · `internal/verify` · `internal/settings` · `internal/cli/mcp_server.go`. 전부 읽기만 합니다.

### 6.1 생산자 쪽 선행 작업 (칸반 뷰 A 전용 — 별도 SPEC 권장)

아래는 **웹 UI 작업이 아니라 상태 기록 쪽 작업**이며, §4.6의 공백을 메웁니다. 웹 SPEC과 분리해서 진행하기를 권합니다 — 소비자(웹)와 생산자(상태 층)가 한 SPEC에 섞이면 어느 쪽이 막혔는지 구분이 어려워집니다.

| 파일 | 작업 | 단계 |
|---|---|---|
| `internal/kanban/record.go` | `Role` 필드 추가 (`plan`/`run`/`review`/`sync`/`lead`) | 2단계 |
| `internal/cli/cc.go` · `glm.go` | 진입 시 파싱한 역할을 레코드에 기록 | 2단계 |
| `internal/kanban/record.go` | 모델·effort 진입 시점 스냅샷 필드 | 3단계 |
| `internal/statusline/` | `context-usage.json` → `context-usage/<session-id>.json` 분리 | 3단계 |
| `.claude/rules/.../context-window-management.md` | 위 경로 변경에 맞춰 소비자 규칙 갱신 | 3단계 |

**⚠️ 컨텍스트 파일 분리는 기존 소비자가 있습니다.** `context-window-management.md`가 이 파일을 "권위 있는 스냅샷"으로 규정하고 읽는 절차(세션 ID 일치 검사, `writer_pid` 판별, 신선도 검사)까지 정의해 두었습니다. 경로를 바꾸면 그 규칙과 읽기 코드를 함께 옮겨야 하며, 단일 슬롯을 전제로 쓰인 검증 로직 상당수가 불필요해집니다.

---

## 7. 이번 범위에서 하지 않는 것

명시적으로 제외합니다. 나중에 필요하면 별도로 다룹니다.

- 웹에서 SPEC 상태 변경 / 카드 드래그 이동 (§2.2 소유권 규율)
- 웹에서 명령 실행 (`/moai run` 등 구동)
- 다중 프로젝트 전환 (현재 콘솔은 단일 checkout 전제)
- 인증·다중 사용자 (루프백 단일 사용자 전제 유지)
- 로그 실시간 tail
- 6컬럼 완성을 위한 backlog 저장소 / PR 조회 (§4.2 (c)안)

---

## 8. 남은 결정 사항

1. **뷰 B 컬럼 수** — §4.2의 (a) 4컬럼 / (b) 6컬럼 빈칸 / (c) 보조신호 중 택일. 권장 (a).
2. **체인 단계 상태를 어떻게 알아낼지** — §4.6의 (가) SPEC 유추 / (나) 리드가 기록 / (다) 하트비트 추정. 권장 (다)→(나).
3. **생산자 작업을 별도 SPEC으로 뺄지** — §6.1. 권장 분리.
4. **비용을 표시할지** — GLM 종량 과금 사용량을 콘솔에 띄우려면 z.ai 쪽 사용량 조회가 필요합니다. 현재 그 데이터를 읽는 코드는 없습니다(미확인 영역). 1단계에서는 **백엔드 배지(정액/종량)만** 보여주고 금액은 넣지 않기를 권합니다.
5. **개요를 `/`로 둘지** — 지금 `/`는 설정입니다. 개요를 `/`로 올리면 설정에 즐겨찾기를 걸어둔 흐름이 바뀝니다.
6. **SPEC 상세를 페이지로 할지 패널로 할지** — 별도 라우트(공유 가능) vs 목록 옆 슬라이드 패널(맥락 유지).

---

## 9. 짝 문서

- `moai-web-redesign-brief.md` — 시각 재설계 지침 (토큰·레이아웃·컴포넌트·제약·검증)
- `current-01~03.png` — 현행 화면 스크린샷
