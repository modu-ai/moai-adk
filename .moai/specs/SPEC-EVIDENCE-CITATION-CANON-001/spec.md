---
id: SPEC-EVIDENCE-CITATION-CANON-001
title: "증거 인용 경로 정본화 — gitignore된 경로를 감사 시점 근거로 쓰지 않는다"
version: "0.3.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules, .claude/agents, .claude/output-styles, .gitignore"
lifecycle: spec-anchored
tags: "evidence, citation, gitignore, verification-claim-integrity, doctrine"
tier: M
---

# SPEC-EVIDENCE-CITATION-CANON-001 — 증거 인용 경로 정본화

## HISTORY

### 2026-08-31 — 최초 작성 (카드 t375)

카드 t375는 "`.gitignore`가 `.moai/state/`를 무시하는데 프로토콜은 거기에 증거를 남기라고 한다"는 모순으로 열렸다. 리드 판정으로 프레이밍이 한 번 바뀌었다: **두 규칙의 충돌이 아니라 규칙 한쪽이 거짓 전제를 깔고 있는 것**이다. 조사 중 이 판정을 뒷받침하는 문장이 규칙 본문에서 그대로 발견됐다(§1.2). 또한 카드가 지목하지 않은 파일 `manager-lead.md`에 같은 전제가 **가장 강한 형태로** 박혀 있는 것을 확인했다.

동시에 운영자 판정으로 카드 t373이 결정을 미룬 ignore 정책 2건(`.moai/observability/`, `.moai/project/navigator/`)이 이 SPEC에 합류했다. 셋 다 "생성된 산출물을 추적할 것인가, 인용 전에 반출할 것인가"라는 같은 질문이고, 규칙이 갈라지면 안 된다는 것이 합류 근거다.

### 2026-08-31 — iter1 감사 FAIL(0.70) 수리

독립 감사가 blocking 11건으로 FAIL을 냈고(`.moai/reports/t375/plan-audit-iter1.md`), 그 전부를 수리했다. 가장 무거운 것은 두 부류였다.

**하나 — 이 문서 자신의 수치가 미귀속이었다.** §1.1 표가 적은 명령이 §1.1 표의 값을 만들지 못했다(D3·D4). 귀속 무결성을 입법하는 문서 안에서 벌어진 일이라 그냥 고치고 넘기지 않고 §1.1.1에 경위를 남긴다.

**둘 — 가드의 반(反)공허 장치가 그 자체로 공허했다**(D1·D2). 하한 7이 실제 모집단(루트 363 + 미러 338 = 701)보다 두 자릿수 작아, 범위가 무너진 가드가 통과할 수 있었다.

그 밖에 D5(carve-out 누락)·D6(판정력 없는 AC)·D7(허용목록 단위)·D8·D9(ignore 판정의 미명시 귀결)·D10(요구 층 공백)·D11(인용 넓이 상한)·D12·D13·D15를 수리했다. **D9는 감사가 제안한 방향으로 가지 않았다** — 이유는 §4.3에 측정과 함께 적는다.

### 2026-08-31 — iter2 감사 PASS-WITH-DEBT(0.85) 부채 종결

Tier M 임계 0.80을 상회하고 재감사 상한에 도달해 종결 판정을 받았다(`.moai/reports/t375/plan-audit-iter2.md`). 남은 부채 4건을 이 판에서 닫았다: **N1** 증거 기반이 이 SPEC 자신의 plan-close를 견디도록 모집단 배제를 도입하고 그 근거를 §1.1.2에 적었다. **D1 잔여** 하한이 원리상 못 하는 하위트리 소실 검출을 집합 상등 단언으로 옮겼다. **D7 잔여** 허용목록 단위 단언에 판정 명령을 붙였다. **N2·N3·N4·N5** 단위·순서 미끄러짐 넷을 정리했다 — 셋이 이 SPEC이 입법하는 바로 그 형태였다.

감사는 iter1의 D9를 **철회했다** — `fix-drafts/`에 관한 §4.3의 반박이 옳다고 판정했다.

**측정 기준**: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t375`, 브랜치 `WT-state-evidence-canon`, HEAD `b64043481`. primary 체크아웃(`/Users/goos/MoAI/moai-adk-go`)에서 잰 값은 그렇게 표시한다.

---

## 1. 배경 — 무엇이 사실인가

### 1.1 측정된 규모

아래 수치는 전부 스크립트 2개가 만든다. 스크립트가 곧 출처이고, 재측정은 그것을 실행하는 것이다. 두 파일은 이 SPEC과 **같은 plan-close 커밋으로 추적 집합에 들어간다** — 그 커밋 이전에는 미추적이므로, "커밋돼 있다"가 아니라 "이 카드가 커밋한다"가 참인 서술이다.

- `.moai/reports/t375/measure_citations.py`
- `.moai/reports/t375/measure_resolution.py`

| 측정 항목 | 값 | 출처 |
|---|---|---|
| 이 문자열을 담은 **추적** `*.md` 파일 | **184** | `git grep -l '\.moai/state/verify' -- '*.md' \| wc -l` |
| 경로 출현 (뒤에 `/`가 오는 형태) | **515** | `measure_citations.py` |
| 그중 **구체** 출현 | **346** | `measure_citations.py` |
| 구체 인용을 1개 이상 담은 파일 | **124** | `measure_citations.py` |
| 중복 제거한 구체 인용 경로 | **231** | `measure_citations.py` |
| primary 체크아웃에서 해석되는 것 | **42 / 231 (18.2%)** | `measure_resolution.py` |
| 갓 만든 워크트리에서 해석되는 것 | **0 / 231 (0%)** | `measure_resolution.py` |

스크립트 출력 원문은 `.moai/reports/t375/citation-figures.txt`. 파생 표는 `cited-path-resolution.txt`(232행 = 헤더 1 + 231)와 `citing-files.txt`(184행).

**판별식**: `.moai/state/verify/` 뒤 꼬리가 ASCII 영숫자로 시작하면 그 인용은 **구체**다. 플레이스홀더(`<session>`, `$MOAI_SESSION_ID`, `...`)는 이 조건에서 자동으로 떨어진다. 두 스크립트가 이 판별식을 **각자 리터럴로** 정의한다 — 한쪽이 다른 쪽에서 import 하지 않으므로 오늘 동일할 뿐 갈라지지 않는다는 보장은 없다. 판별식을 고칠 때는 두 파일을 함께 고치고 둘 다 다시 돌린다.

**두 스크립트 모두 추적 집합만 읽는다** (`git grep -l`, `grep -r`이 아니다). `grep -r`은 미추적 파일까지 훑으므로 카드가 진행될수록 값이 움직인다.

### 1.1.2 [HARD] 모집단은 이 SPEC 자신을 배제한다 — 그래야 추적 집합 논거가 끝까지 성립한다

추적 집합을 고른 논거에는 유효 기간이 있다. 이 SPEC의 3개 산출물은 전부 `.moai/state/verify` 문자열을 담으므로, **plan 단계가 닫혀 SPEC이 추적 집합에 들어가는 순간 `git grep`이 그것을 세기 시작한다.** 그러면 위 표는 이 카드 자신의 정상적인 다음 동작 하나로 스테일해진다 — §1.1.1이 고백한 결함과 같은 계열이고, 기제만 명령 오기가 아니라 **시간**이다.

그래서 두 스크립트는 모집단에서 이 SPEC의 디렉터리를 배제한다(`SPEC_OWN_DIR`, 각 스크립트 32행·필터 1행). 배제가 실제로 값을 고정하는지 확인했다(이 트리, HEAD `b64043481`):

| 입력 모집단 | 입력 파일 수 | 스캔 | 출현 | 구체 | 파일 | 경로 |
|---|---|---|---|---|---|---|
| 추적 집합만 | 184 | 184 | 515 | 346 | 124 | 231 |
| 추적 집합 + 이 SPEC 산출물 | 188 | **184** | **515** | **346** | **124** | **231** |

값이 하나도 움직이지 않는다. 배제가 없으면 같은 비교에서 184→187 / 515→539 / 346→351 / 124→127 / 231→234로 전부 이동한다(iter2 감사 N1 실측).

**이 배제는 측정을 좁히는 것이 아니라 측정 대상을 정의하는 것이다.** 부채는 이 SPEC이 오기 **전부터** 있던 문서들이고, 그것을 세는 수가 이 SPEC의 진행 상태에 따라 흔들리면 애초에 부채의 크기를 말할 수 없다.

primary 체크아웃 `.moai/state/verify/` 실체: 최상위 디렉터리 **124개** — 그중 1개는 `snapshots`(§1.4의 기계 저장소)이므로 **세션 디렉터리는 123개**다. 파일 **905개**, **13 MB**(`find … -maxdepth 1 -mindepth 1 -type d | wc -l`, `find … -type f | wc -l`, `du -sh`). 대조군으로 이미 추적 중인 `.moai/reports/`: 추적 파일 **974개**(`git ls-files .moai/reports | wc -l`, t375 트리), primary 디스크 전체 **2,421 파일 / 29 MB**(미추적 포함).

### 1.1.1 이 문서 자신이 저지른 미귀속 — 경위를 남긴다

초판 §1.1의 표는 값 3개(184 / 532 / 131)를 적으면서 **그 값을 만들지 못하는 명령**을 출처로 달았고, 바로 아래에 "위 표의 명령이 이 문서가 쓰는 값의 유일한 출처다"라고 단언했다. 그 단언이 거짓이었다.

- 184·532는 **추적 집합**(`git grep`)에 대해 옳고, 적혀 있던 **작업 트리 명령**(`grep -r`)에 대해 틀렸다. 같은 트리에서 후자는 187·562를 낸다 — 차이는 이 SPEC 자신의 산출물 3개다.
- 131은 어느 명령으로도 재현되지 않았다. 지목된 증거표에서 재도출하면 97이 나왔다.

**532와 §1.1 표의 515는 모순이 아니라 단위가 다르다.** 초판이 센 것은 맨 문자열 `.moai/state/verify`이고, 지금 표가 세는 것은 **뒤에 `/`가 오는** 출현이다. 차이 17은 `/`가 뒤따르지 않는 출현 — 디렉터리 자체를 가리키는 산문 언급이라 인용 경로로 셀 대상이 아니다.

```
git grep -o '\.moai/state/verify'  -- '*.md' | wc -l   # 532
git grep -o '\.moai/state/verify/' -- '*.md' | wc -l   # 515
```

단위가 좁아진 것이고, 그 사실을 적지 않으면 두 절이 서로를 반박하는 것처럼 읽힌다 — 이 SPEC이 입법하는 바로 그 종류의 미끄러짐이라 여기 적는다.

**왜 이렇게 됐나**: "구체적 인용"이 무엇인지 **글로 정해진 적이 없었고**, 손으로 돌린 변형 셋이 129 / 131 / 132로 갈렸다. 갈린 지점은 어느 플레이스홀더 표기를 제외 목록에 기억해 넣었느냐 하나뿐이었다. 그래서 지금의 판별식은 **제외 목록이 아니라 양성 문자 클래스**로 쓰여 있다 — 무엇을 빼야 하는지 기억할 필요가 없는 형태다.

이것을 지우지 않고 적는 이유: **이 SPEC이 금지하려는 바로 그 결함을 이 SPEC의 증거 안에서 저질렀다.** 수치만 조용히 갈아끼우면 다음 사람은 같은 자리에서 같은 방식으로 미끄러진다.

### 1.2 거짓 전제 — 규칙 한 문장 안에 모순이 들어 있다

`.claude/rules/moai/core/agent-common-protocol-reference.md:60`:

> To satisfy this reachability obligation, evidence SHALL be persisted under `.moai/state/verify/<session>/` (**gitignored runtime state**, …)

같은 문단이 요구하는 것은 "감사 시점에 해석되는 경로"이고, 같은 문장이 괄호 안에서 그 경로가 gitignore 대상임을 **스스로 밝힌다**. gitignore된 경로는 clone·CI·다른 머신 어디에도 따라가지 않으므로 감사 시점 도달성을 만족시킬 수 없다. `0 / 231`이 그 증명이다.

`.claude/rules/moai/core/agent-common-protocol.md:268`이 같은 내용을 요약해 다시 말한다.

즉 **두 규칙이 다투는 것이 아니라, 한쪽이 성립하지 않는 전제 위에 서 있다.** 수리는 두 문장을 화해시키는 것이 아니라 거짓 전제를 걷어내는 것이다.

### 1.3 카드가 지목하지 않은 더 강한 사례

`.claude/agents/moai/manager-lead.md:150` (같은 문장이 템플릿 미러와 방출된 codex `.toml`에도 있다):

> The path MUST resolve at audit time … `.moai/state/verify/` is the **canonical persistence location** (NOT `/tmp`).

여기서는 gitignore 사실조차 적혀 있지 않아 읽는 쪽이 모순을 알아챌 단서가 없다. 같은 파일이 증거 표 열과 `mkdir -p` 레시피까지 제공하므로, 이 파일은 거짓 전제를 **지시로 바꾸는** 지점이다. REQ-ECC-002의 주어가 "규칙 문서"가 아니라 **doctrine 표면 문서**인 것은 이 사실 때문이다.

### 1.4 옳은 `.moai/state/verify/` 사용처는 둘이다

**둘 다 기계 소비자이고, 둘 다 문서가 사람에게 인용하지 않는다.**

1. **`internal/verify/store.go:15`** — `SnapshotDir = ".moai/state/verify/snapshots"`. `moai verify record` / `moai verify check --key-current`가 읽고 쓰는 저장소로, HEAD SHA로 키가 걸려 같은 트리·같은 머신 안에서만 유효하도록 설계돼 있다.
2. **`internal/web/events.go:29`** — `"verify": {".moai/state/verify"}`. fsnotify 감시 맵이 `snapshots`가 아니라 **디렉터리 전체**를 SSE 이벤트 소스로 본다.

**따라서 이 SPEC은 `.moai/state/verify/`를 없애지 않는다.** 없애는 것은 그 경로의 *인용 대상* 역할 하나뿐이다. 스크래치·스냅샷·감시 소스 역할은 그대로 둔다. REQ-ECC-006의 carve-out은 이 둘을 모두 이름 붙여야 한다 — 초판은 첫째만 적고 "하나뿐"이라고 단정했다.

### 1.5 이 저장소는 "생성물"을 일괄 배제하지 않는다

`.moai/project/codemaps/`에는 추적 `.md`가 **6개** 있고(`git ls-files '.moai/project/codemaps/*.md' | wc -l` → 6), `provenance.json`을 더하면 디렉터리 총 7개다. 생성물이면서 **의도적으로 추적**된다. 그러므로 "생성됐다"는 사실만으로는 이 저장소에서 ignore 근거가 되지 않는다. §4의 두 판정은 이 전제 위에서 내려진다.

---

## 2. 요구사항 (GEARS)

### 2.1 인용 경로 정본

- **REQ-ECC-001** — 문서가 검증·판정의 근거로 인용하는 증거 경로는 버전 관리로 추적되는 경로여야 한다(`shall`). 이 저장소의 정본 위치는 `.moai/reports/<card-id>/`다.
- **REQ-ECC-002** — **doctrine 표면 문서**(`.claude/rules/`, `.claude/agents/`, `.claude/output-styles/`, `.claude/skills/` 하위)는 `.moai/state/verify/<session>/`를 `/tmp` 소거를 견디는 **머신-로컬 스크래치**로 명명해야 하며, 감사 시점 인용 대상으로 명명해서는 안 된다(`shall not`).
- **REQ-ECC-003** — When 어떤 산출물이 판정 근거로 인용될 때, 행위자는 인용 **이전에** 그것을 추적 경로로 반출해야 한다(`shall`). 반출되지 않은 경로를 인용하는 것은 미귀속 주장이다.
- **REQ-ECC-004** — 인용은 **파일 하나를 이름 붙여야 하며**(`shall`), 디렉터리를 이름 붙여서는 안 된다(`shall not`). 반출 대상은 그렇게 이름 붙여진 파일로 한정되고, 스크래치 디렉터리 통째 반출은 금지된다.
- **REQ-ECC-005** — 규칙 본문은 REQ-ECC-004의 선택 기준을 서술해야 한다(`shall`): 판정을 결정한 명령과 그 명령의 판정 결정선(exit 코드, 실패 요약, 인용된 수치)만 반출하고, 그 외 원문은 스크래치에 남긴 채 손실 위험을 Residual-risk에 적는다.
- **REQ-ECC-006** — Where 소비자가 기계인 경우, 규칙은 그 사용처를 REQ-ECC-002의 명시적 예외로 서술해야 한다(`shall`). 열거 대상은 §1.4의 **둘 다**이다: 스냅샷 저장소(`internal/verify/store.go`, `moai verify record|check`)와 SSE 감시 소스(`internal/web/events.go`).

> **REQ-ECC-004와 005가 하나의 상한을 이룬다.** 초판은 둘이 서로 다른 상한을 걸어, 인용문을 넓게 쓰면(`evidence: .moai/state/verify/<session>/`) 양쪽을 동시에 만족하면서 전부 반출할 수 있었다. 인용의 **넓이 자체**를 004가 구속하고, 그 안에서 무엇을 남길지를 005가 정한다.

### 2.2 기계적 가드

- **REQ-ECC-007** — 저장소는 doctrine 표면 하위 `.md`에서 `.moai/state/verify` 인용을 검출하는 가드를 가져야 한다(`shall`).
- **REQ-ECC-008** — 가드는 저장소 루트 사본과 `internal/template/templates/` 미러 **양쪽**을 대상으로 하고, **두 트리 각각**에 대해 방문 사실과 스캔 하한을 단언해야 한다(`shall`). 한쪽만 보는 가드는 다른 쪽에서 규칙이 삭제돼도 초록으로 남는다.
- **REQ-ECC-009** — Where 가드가 예외 허용목록을 가지는 경우, 항목은 **파일 + 정확한 리터럴** 단위여야 하며(`shall`), 항목 하나가 파일 전체를 면제해서는 안 된다(`shall not`). 가드는 목록의 항목 수를 단언해야 한다.
- **REQ-ECC-010** — 가드는 위반 문서를 실제로 잡아내는 것이 증명돼야 한다(`shall`) — 통과 방향만 보이는 시연은 규칙이 꺼진 상태와 구별되지 않는다.
- **REQ-ECC-011** — §C.3의 경계 사례 3건은 파일별로 판별식 적용 결과와 그 이유가 `progress.md`에 기록돼야 한다(`shall`).

### 2.3 ignore 정책 2건

- **REQ-ECC-012** — 저장소 루트 `.gitignore`와 템플릿 미러는 `.moai/observability/*.jsonl`을 무시해야 한다(`shall`). `.gitkeep` 예외 줄을 두어서는 안 되고, `.moai/observability/` 디렉터리 스캐폴드를 저장소나 템플릿에 만들어서는 안 된다(`shall not`) — 근거는 §4.2.
- **REQ-ECC-013** — 저장소는 `.moai/project/navigator/` 아래 어떤 경로도 무시해서는 안 된다(`shall not`) — 디렉터리 전체도, `fix-drafts/`도, `capability-map.md`도. 근거는 §4.3.

---

## 3. 인용 정본 판정 — 왜 Option A이고 Option B가 아닌가

**채택(Option A)**: 프로토콜을 고쳐 `.moai/state/verify/`를 스크래치로 명명하고, 인용은 추적 경로의 **파일**을 이름 붙이며, **인용 전 반출**을 의무로 세운다.

**기각(Option B — `.gitignore`에 `.moai/state/verify/` 예외 추가)**. 두 근거를 남긴다.

1. **저장소 비대.** primary 한 대에서만 905 파일 / 13 MB이고, 세션마다 자란다. 상한이 없다.
2. **이력에 무엇이 들어가야 하는가의 차이.** `.moai/reports/`가 담는 것은 사람이 추린 판정 산문이고, `.moai/state/verify/`가 담는 것은 원시 로그다. 크기 차이가 아니라 **종류의 차이**다. 로그 원문을 이력에 넣는 것은 판정을 더 검증 가능하게 만들지 않는다 — 판정을 결정한 몇 줄이 그렇게 만든다.

### 3.1 [HARD] 범위 경계 — 반출은 인용이 이름 붙인 파일만

이 경계가 없으면 Option A는 Option B의 비대를 목적지만 바꿔 재현한다: 같은 13 MB가 `.moai/reports/`로 옮겨 갈 뿐이다. 그래서 규칙 본문은 **인용 넓이 상한**(REQ-ECC-004)과 **선택 기준**(REQ-ECC-005)을 **둘 다** 실어야 하고, 둘이 같은 상한을 가리켜야 한다.

**작동 사례** (출처 등급을 함께 적는다):

- **전문(傳聞) — 리드가 전한 것이고 이 카드가 관측하지 않았다.** 카드 t341의 lane-5는 338 KB / 1,095행짜리 축어 전사를 이력에 넣지 않기로 했고, 판정문이 실제로 인용한 요약 줄만 커밋했으며, 전문(全文)의 손실 위험을 판정문 Residual-risk에 적었다.
- **직접 측정 — 이 카드가 primary에서 쟀다.** `.moai/reports/t341/`은 파일 13개 / 52 KB이고 하위에 `verify/` 하나를 가진다. 이 측정은 338 KB가 반출되지 않았다는 서술과 **모순되지 않는다**. 확증은 아니다 — 52 KB라는 사실은 338 KB 이야기가 참일 때도, 애초에 그런 전사가 없었을 때도 똑같이 나온다.
- **이 사례 자체가 결함의 실례다.** t341의 판정문은 이 SPEC 초판을 쓰는 시점에 읽을 수 없었다(아직 반출되지 않았다).

---

## 4. ignore 정책 2건 — 판정과 근거

### 4.1 t373 선례에서 뽑은 판별식

t373은 `.moai/chain/`과 `.moai/.migrate-tx-*.json`을 **일부러 무시하지 않았다** (`/Users/goos/MoAI/moai-adk-go/.moai/reports/t373/verdict.md`, 131-147행). `.gitignore` 항목은 "이 파일은 무시해도 된다"는 선언인데, 저 두 경로의 잔여물은 정확히 그 반대 — **처분이나 트랜잭션이 끝나지 않았다**는 뜻이고, `git status`에 뜨는 것이 그 사실의 유일한 신호이기 때문이다.

**판별식**: 그 자리에 파일이 남아 있다는 사실이 *무언가 잘못됐다는 신호*인가? 그렇다면 무시하지 않는다.

**판별식의 숨은 전제**: `chain/`에서 이것이 성립하는 이유는 **코드가 성공 경로에서 그것을 처분하기 때문**이다(`disposeLegacyChainDir`). 처분이 있어야 잔존이 실패를 뜻한다. 처분하는 코드가 없으면 잔존은 성공한 실행에서도 똑같이 남으므로 아무 신호도 아니다. 아래 두 판정은 이 전제를 각각 확인한다.

### 4.2 `.moai/observability/` — `*.jsonl`만 무시하고, `.gitkeep`은 만들지 않는다

**측정된 사실** (t375 트리 + primary):

- `.gitignore`에 없다. 디렉터리는 t375에도 primary에도 **존재하지 않는다**(`ls -d` 둘 다 실패).
- 생산자: `internal/hook/post_tool_duration.go`, `hookMetricsRelPath = ".moai/observability/hook-metrics.jsonl"` (19행).
- 쓰기 앞 관문 둘, 소스에서 확인(100-125행): `durationMs <= threshold`면 조기 반환(`slow_hook_threshold_ms`, 로컬 5000), 그리고 `.moai/observability/`가 없으면 — 자체 주석이 REQ-CC2122-HOOK-001-003을 인용하며 — 조용히 조기 반환. **디렉터리를 만드는 행위 자체가 opt-in이다.**
- 형제 격 `.moai/evolution/telemetry/*.jsonl`은 이미 무시되고(`.gitignore:229-230`) `.gitkeep` 예외를 단다. 그 `.gitkeep`은 **추적되고**(`git ls-files .moai/evolution/telemetry` → `.gitkeep`), **템플릿에도 있다**(`internal/template/templates/.moai/evolution/telemetry/.gitkeep`).

**판정**: `.moai/observability/*.jsonl`을 무시한다. **`!.moai/observability/.gitkeep` 예외 줄은 두지 않고, `.gitkeep` 파일도 만들지 않는다.**

**층위 분리 논증** (초판에서 유지). t373이 갈리지 않는다고 본 두 읽기(일관성 → ignore / opt-in은 공유 의도 → track)는 **같은 층에 있지 않다.** 디렉터리 생성이 여는 opt-in은 *훅이 기록하도록 하는 것*이지 *이력에 남기는 것*이 아니고, 코드는 이 둘을 묶지 않는다.

**초판이 적지 않았던 귀결, 그리고 그것이 판정을 바꾼 지점.** 초판은 "텔레메트리 형제와 같은 두 줄 형태"를 택했다. 그 형태를 끝까지 따르면 `.gitkeep`이 추적되고 템플릿에 실린다. 그러면:

- 이 저장소의 모든 clone에서 `.moai/observability/`가 존재하게 되어 `os.Stat` 관문이 항상 통과한다.
- 더 무겁게 — 템플릿에 실린 `.gitkeep`은 `moai init`이 **모든 사용자 프로젝트에 배포한다.** 배포판 사용자 전원에게 훅 기록이 기본 활성화된다.

코드가 일부러 세운 opt-in(REQ-CC2122-HOOK-001-003)이 opt-out으로 뒤집히고, 그것도 `.gitignore` 카드의 부수 효과로 뒤집힌다. **이 SPEC은 그 판정을 내릴 자리가 아니다.** 관측 기록을 기본 활성화할 것인가는 별도의 의도적 결정이고, 그렇게 하려는 사람은 이 SPEC이 아니라 그 결정을 다루는 카드에서 해야 한다.

**반대 뿔에 대한 답.** `.gitkeep`을 만들지 않으면 남는 것은 "존재한 적 없는 디렉터리에 대한 `*.jsonl` 무시 규칙"이고, 이는 §4.3이 반대하는 "예측으로 무시하기"처럼 보인다. 두 수는 같은 종류가 아니다.

- `*.jsonl` 패턴은 **불활성**이다. 누군가 로컬에서 디렉터리를 만들어 opt-in 하는 순간 효력이 생기고, 그전까지 아무것도 하지 않는다. 비용 0.
- `.gitkeep` 커밋은 **불활성이 아니다.** 배포되는 순간 모든 사용자의 런타임 동작을 바꾼다.

§4.1의 판별식도 같은 답을 준다: `hook-metrics.jsonl`에 줄이 있다는 것은 무언가 실패했다는 뜻이 아니라 훅이 느렸다는 뜻이다. 보이는 것 자체가 신호를 담지 않는다. `!` 예외 줄은 이 판정에서 할 일이 없으므로 뺀다.

### 4.3 `.moai/project/navigator/` — 아무것도 무시하지 않는다

**측정된 사실** (t375 트리):

- `.gitignore`에 없다. 오늘 파일이 정확히 **1개**이고 추적된다: `symbols/narrative.template.md` — 템플릿, 즉 **입력**이지 생성물이 아니다.
- 생산자들이 요구 시 쓰는 산출물: `work-items.{md,json}`(`navigator_route.go`), `fix-drafts/<draft-id>/request.json`(`navigator_fix.go:20,60`), `capability-map.md`(`navigator_enrich.go`), `tiers.json`(`navigator_tiers.go`), `nav-graph.json` · `audit-report.json`(`navigator_sync.go`). **현재 디스크에 하나도 없다.**
- `capability-map.md`는 **입력으로 읽힌다**: `navigator_enrich.go:75`가 그 경로를 기본값으로 열고 출력은 `.moai/project/codemaps`로 나간다. 디렉터리 통째 ignore는 위험하다.

**판정**: 아무 항목도 추가하지 않는다. 초판은 `fix-drafts/`만 무시하려 했고, 그 판정을 **철회한다**.

**§4.1 판별식을 `fix-drafts/`에 실제로 적용한 결과.** 감사(D9)는 남아 있는 `request.json`이 "요청됐으나 완료되지 않은 위임"을 뜻하므로 `chain/`과 같은 모양이고, 따라서 무시하면 안 된다고 제안했다. 그 가설을 §4.1의 숨은 전제로 검사했다 — **처분하는 코드가 있는가?**

```
grep -rn 'RemoveAll' internal/navigator/fix/ internal/cli/navigator_fix.go   (테스트 제외) → 0행
```

**없다.** `internal/navigator/fix/types.go:77`은 승인 후 `fix-drafts/<draft-id>/applied.json`을 쓴다고 적지만, 성공 경로에서 디렉터리를 치우는 코드는 어디에도 없다. 그러므로 `fix-drafts/<id>/`의 **존재 자체**는 성공한 실행에서도 똑같이 남고, 아무 신호도 아니다 — `chain/`과 모양이 다르다. 감사의 반대 독해는 여기서 성립하지 않는다.

다만 그것이 무시해도 된다는 결론으로 이어지지는 않는다. 두 가지가 남는다.

1. **더 가는 신호는 실재한다.** 완료 여부를 가르는 것은 디렉터리의 존재가 아니라 그 안의 `applied.json` 유무다. `.gitignore`는 경로에 걸리지 내용에 걸리지 않으므로, `fix-drafts/`를 무시하면 그 구분도 함께 사라진다.
2. **존재-부재 논거의 비대칭이 해소되지 않는다.** 감사가 지적한 대로, 나머지 navigator 산출물을 무시하지 않는 근거로 "아직 만들어진 적 없는 것을 예측으로 무시하지 않는다"를 들면서 `fix-drafts/`만 예외로 두는 이유가 없다. **오늘 그것도 존재하지 않는다.**

두 이유가 같은 방향을 가리키므로 판정을 통일한다: navigator 아래는 아무것도 무시하지 않는다. codemaps가 일부러 추적되는 것(§1.5)과도 일관된다.

**이 판정이 t373의 미결을 닫는 방식**: t373은 "관측으로 갈리지 않는다"며 넘겼다. 여기서 갈린 것은 새 관측(처분 코드 부재)이 아니라 그 관측이 **감사의 반대 독해를 반증했다**는 사실이고, 반증 후에 남은 두 근거가 모두 같은 답을 준다는 것이다.

---

## 5. 잔여 부채 — 소급 정정은 범위 밖

이 규칙은 **앞으로만** 구속한다. 기존 문서를 고치는 것은 이 SPEC의 범위가 아니다.

**남는 부채(측정치, §1.1 출처)**: 구체 인용을 담은 문서 **124개**, 구체 인용 경로 **231건**. 그중 **189건(231 − 42)이 이미 primary에서조차 해석되지 않고**, **231건 전부가 clone에서 해석되지 않는다.** 부채는 이 SPEC이 만든 것이 아니라 이 SPEC이 측정해 드러낸 것이다.

**부채를 닫으려면**: 124개 문서를 훑어 각 인용마다 (a) 원본이 아직 `.moai/state/verify/`에 남아 있으면 추려서 `.moai/reports/<card-id>/`로 반출하고 인용을 갈아끼우거나, (b) 남아 있지 않으면 그 인용을 미귀속으로 표시한다 — 후속 배치의 카드 후보.

**별건 후속 후보 하나 더**: `internal/cli/mcp_glm.go:110`의 **코드 주석**이 측정 근거로 `.moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md`를 인용한다. 이 SPEC이 없애려는 결함의 살아 있는 실례인데, 가드 범위가 `.md` 한정이라 잡히지 않는다. `.go` 주석까지 스캔 범위를 넓히는 것은 별개 판단이므로 카드로 넘긴다.

---

## 6. 범위에서 제외

### Out of Scope — 소급 정정

- 기존 124개 인용 문서의 경로 교체·반출.
- `.moai/state/verify/` 아래 현존 905개 파일의 이관·삭제·정리.
- `internal/cli/mcp_glm.go:110`의 코드 주석 인용, 그리고 가드 범위를 `.go`로 넓히는 것.

### Out of Scope — 메커니즘 변경

- `.moai/state/verify/snapshots/` 스냅샷 저장소와 `internal/web/events.go`의 SSE 감시 대상 변경(§1.4 carve-out).
- `moai verify record` / `moai verify check` CLI 동작 변경.
- 관측 기록(`.moai/observability/`)을 기본 활성화할 것인지의 판정 — §4.2가 명시적으로 이 SPEC 밖으로 뺀다.
- 반출을 자동화하는 새 CLI 동사나 훅.

### Out of Scope — 다른 ignore 경로

- `.moai/chain/`, `.moai/.migrate-tx-*.json` — t373이 이미 판정했고 이 SPEC은 그 판정을 유지한다.
- `.gitignore`의 그 밖의 항목 전반에 대한 재검토.
