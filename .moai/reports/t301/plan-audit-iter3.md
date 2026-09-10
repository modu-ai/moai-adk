# SPEC 감사 보고서 (iteration 3 · 최종): SPEC-FULL-SUITE-DOCTRINE-001 v0.3.0

Iteration: **3/3** (운영자가 Tier M 상한 2를 넘겨 승인한 마지막 회차)
Verdict: **FAIL**
Overall Score: **0.81** — iter-1 `0.57` 대비 **+0.24**, iter-2 `0.75` 대비 **+0.06** (단조 개선)
Tier M PASS 기준선 **0.80** (`spec-workflow.md:141` 실측) — **점수만 놓고 보면 처음으로 기준선을 넘었다.**
Run-phase 진입: **아직 아니다** (아래 § 진입 권고 — 좁은 관문 1개)

측정 트리: `.claude/worktrees/t301`, HEAD `d29b8942e`. iter-1·iter-2와 동일 트리.
저자 추론 맥락은 M1 Context Isolation에 따라 무시했다 (Reasoning context ignored per M1 Context Isolation). 저자가 진술한 조치·근거는 전부 **주장으로만** 취급했고, 아래 모든 판정은 이 세션에서 직접 실행한 명령 출력에 귀속한다.

---

## 판정 요약 — 점수가 아니라 결함 하나가 verdict를 가른다

iter-2가 남긴 차단 결함 4건(N1 · D3' · D9' · N2)은 **전부 닫혔다.** 검증은 주장 대조가 아니라 명령 재실행으로 했고, 기재된 baseline **13개 전부가 이 트리에서 그대로 재현된다.** 새로 도입된 세 개의 구간 길이 단언(STEP 4 = 14, STEP 5 = 10, 배치 1번 = 4)도 세 사본에서 정확히 재현된다.

그럼에도 FAIL이다. 이유는 하나다 — **C3(`.codex/.../manager-develop.toml`)는 손으로 고치는 미러가 아니라 C2에서 기계적으로 생성되는 산출물인데, SPEC이 그 사실을 모른 채 "손으로 고치라"고 지시한다.** SPEC 전체에서 생성기(`internal/template/agentemit`)와 그 재생성 타깃(`make agents-emit`), 그리고 그 드리프트를 지키는 골든 테스트가 **한 번도 언급되지 않는다**(실측 0회). 지시대로 하면 커밋된 골든 가드가 깨지거나, 다음 재생성이 수리를 조용히 되돌린다.

이것은 이 SPEC이 스스로 이름 붙인 위험 그대로다. iter-2에서 C3를 범위에 넣은 것은 옳았는데, **C3에 잘못된 유지보수 모델을 가져다 붙였다.**

---

## Must-Pass 결과 (7/7 통과)

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o '^- \*\*REQ-FSD-[0-9]*' spec.md | sort | uniq -c` → REQ-FSD-001~011 각 1회, 결번·중복 없음, 3자리 패딩 일관.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 층)** — 11개 전부 GEARS. 001/004/005 Ubiquitous · 002/003/007/008/010 Unwanted · 006/011 Event-driven · 009 Capability gate(`Where`). REQ-FSD-006은 `and … in the same sentence …; carrying only one of the two is not compliance` 로 확장됐으나 `When …, the … shall …` 골격이 유지된다.
- **[PASS] MP-3 YAML 프론트매터 유효성** — 실측 필드 13개: `id title version status created updated author priority phase module lifecycle tier tags`. 정본 12 + 선택 `tier: M`(enum 유효). snake_case 별칭 0건. `version: "0.3.0"` 인용부호 있는 semver 문자열.
- **[N/A] MP-4 §22 언어 중립성** — 다중 언어 툴링 서술이 아니다. `spec.md §D` 가 지시문 층(중립)과 예시 층(기존 Go)을 명시 분리한다.
- **[PASS] MP-5 D7 교차 SPEC 정합** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 자기 참조 1건. retired/superseded/archived 0건.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall'` → 산출물 4종 모두 `0`. D8-4 자동 통과.
- **[PASS] MP-7 clarification 게이트** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/` → 매치 없음(`rc=1`, 파이프 없이 읽음).

부수 확인 — **Tier M 예산**: REQ 11 / AC 15, 상한 각 16(`spec-workflow.md:148-150`). AC 헤딩 수 `15`, 고유 ID `AC-FSD-001`~`015` 실측. 저자가 AC-FSD-002를 **재배정**해 총수를 늘리지 않은 결과이며, 상한까지 여전히 1칸 남았다.

**FAIL은 M5 방화벽이 아니다** — 일곱 항목 모두 통과했다. 아래 N3(critical, blocking)이 verdict를 가른다.

---

## 1. 전 AC 명령 재실행 — 기재 baseline 13/13 재현

`acceptance.md` 의 **모든** AC 명령을 이 트리에서 직접 실행했다. 파이프를 통해 `rc` 를 읽지 않았다.

### 에이전트 정의 축 (사본 3벌)

| AC | 명령 요지 | 기재 baseline | 실측 C1 / C2 / C3 | 판정 |
|---|---|---|---|---|
| AC-FSD-001 | 리터럴 4패턴 | 3 / 3 / 3 | `3` / `3` / `3` | ✓ |
| AC-FSD-002 | 어휘족 `(full\|complete\|entire\|whole)[ -](test )?suite`, `-i` | 4 / 4 / 4 | `4` / `4` / `4` | ✓ |
| AC-FSD-003 | `full suite, coverage` | 1 / 1 / 1 | `1` / `1` / `1` | ✓ |
| AC-FSD-014 | `LARGE_SCALE` | 3 / 3 / 3 | `3` / `3` / `3` | ✓ |
| AC-FSD-012 | STEP 4 구간 길이 | 14 / 14 / 14 | `14` / `14` / `14` | ✓ |
| AC-FSD-012 | 구간 내 정본 문구 | 0 / 0 / 0 | `0` / `0` / `0` | ✓ RED-now |
| AC-FSD-007 | STEP 5 구간 길이 | 10 / 10 / 10 | `10` / `10` / `10` | ✓ |
| AC-FSD-007 | `integration branch` (구간 내) | 0 / 0 / 0 | `0` / `0` / `0` | ✓ RED-now |
| AC-FSD-007 | `PENDING at report time` (구간 내) | 0 / 0 / 0 | `0` / `0` / `0` | ✓ RED-now |

**C3 열이 `.md` 사본과 같은 값이라는 사실을 저자가 가정하지 않고 따로 쟀다는 주장은 참이다** — 나도 따로 쟀고 값이 일치한다. 다만 그 일치에는 저자가 모르는 이유가 있다(§ N3).

`AC-FSD-002` 의 `4` 가 `AC-FSD-001` 의 `3` 보다 1 큰 까닭도 직접 확인했다. C1에서 어휘족이 잡는 줄은 `92`·`126`·`132`·`135` 이고, 리터럴 집합이 놓치는 한 줄이 `:135`(S4의 `full suite, coverage`)다. `-i` 가 필요한 이유도 `:132` 가 `COMPLETE` 대문자라는 데 있다. **어휘족 패턴은 네 지점 전부를 덮는 상위집합이다.**

### 배치 정의 축 (codex 미러 없음)

| AC | 기재 baseline | 실측 | 판정 |
|---|---|---|---|
| AC-FSD-004 `Full test suite` | 각 1 | 로컬 `1` / 템플릿 `1` | ✓ |
| AC-FSD-005 Group A 행 내 전량 호출 | 각 1 | `1` / `1` | ✓ |
| AC-FSD-006 구간 길이 | 4 | `4` | ✓ |
| AC-FSD-006 구간 내 전량 호출 | 각 1 | `1` / `1` | ✓ |
| AC-FSD-013 구간 내 `internal/<pkg>` | 각 0 | `0` / `0` | ✓ RED-now |
| AC-FSD-015 Group A 행 내 `[0-9]+-[0-9]+ s` | 각 1 | `1` / `1` | ✓ |

### 동등형 · 중립성

| AC | 기재 baseline | 실측 | 판정 |
|---|---|---|---|
| AC-FSD-008 세 쌍 diff | `1a2 > isolation: worktree`, 무출력, 무출력 | 동일 (`rc1=1`, `rc2=0`, `rc3=0`) | ✓ |
| AC-FSD-009 diff 훑은 줄 | 0 (사전 미성립이 의도) | `exit=0`, `grep -c '^+'` → `0` | ✓ |
| — `origin/develop` in templates | rc=1 | 파이프 없이 `rc=1` | ✓ |
| — C3 로컬 쌍 부재 | 없음 | `No such file or directory` | ✓ |

**형식 불량 패턴 0건. 사후 기대치에서 출발해 아무것도 증명하지 못하는 항목 0건.** AC-FSD-008·010은 사전=사후지만 불변 보존형이라 유형상 정상이고, AC-FSD-009는 사전에 의도적으로 미성립한다(훑은 줄 하한이 공허 스윕을 막는 장치).

---

## 2. mutant 재구성 (질문 2)

문자열 프로브로 각 패턴의 실제 동작을 확인한 뒤 판정했다.

| mutant | 내용 | 결과 | 잡는 AC |
|---|---|---|---|
| **M-1 (C3 표적)** | C1·C2만 완전 수리, C3 무손 | **잡힘** | AC-001(C3 `3`) · AC-002(`4`) · AC-003(`1`) · AC-014(`3`) · AC-012(`0`) · AC-007(`0`,`0`) — **독립 6중** |
| **M-2 (C3 부분)** | C3의 S1·S2·S3 리터럴만 고치고 S4 방치 | **잡힘** | AC-002(프로브 실측 `1`) · AC-003(`1`) — 2중 |
| **M-3 (iter-2 생존자 재실행)** | S1 → `Step 5 runs the complete suite regardless of project size.`, 정본 문구는 S2에만 | **잡힘** | AC-002 — 프로브 실측 `1`. **iter-2에서 전 항목을 통과했던 mutant가 이제 RED다** |
| **M-4 (신규 · 우회형)** | S1 → `Step 5 runs every package's tests regardless of project size.` | **생존** | 프로브 실측: 리터럴 `0`, 어휘족 `0`. 전 MUST 초록 |
| **M-5 (토큰 분리)** | 두 정본 토큰을 STEP 5의 서로 다른 두 불릿에 배치 | **생존(경미)** | AC-007은 각 `1`로 통과. 같은 문장 여부는 눈 확인 |
| **M-6 (구간 폭주)** | 편집 중 `### STEP 5` / `### Checkpoint` 헤딩을 깨뜨림 | **잡힘** | AC-012 길이 ≠ 14, AC-007 길이 ≠ 10 — Then 안에 들어간 길이 단언이 정확히 이것을 막는다 |

**M-4 판정 — 차단 결함으로 올리지 않는다.** 이 mutant는 전량 어휘족 전체를 버리고 부자연스러운 우회 표현(`every package's tests`)을 일부러 골라야 성립한다. 리터럴 4종 · 어휘족 4종 · 판별자 · 구간 고정 정본 문구를 동시에 피하면서 여전히 전량을 지시해야 한다. grep 기반 AC 집합으로 임의의 의역까지 닫는 것은 원리적으로 불가능하고, **닫으라고 요구하면 도착 시점에도 완료 시점에도 RED인 항목을 만드는 것**이다. 남은 공간은 "그럴듯한 구현 실패" 로 도달되지 않는다 — 적대적 구성으로만 도달된다. iter-1의 M-C, iter-2의 M-A가 **평범한 편집으로도 나올 수 있었던** 것과는 성격이 다르다.

---

## 3. 범위 규약이 실제로 모든 요구사항을 묶는가 (질문 3)

`spec.md:100` 의 규약은 **`the manager-develop agent definition`** 이라는 구(句)에 C1·C2·C3를 결속한다. 요구사항별로 대조했다:

| REQ | 주어 표기 | 규약 결속 | 실질 결속 |
|---|---|---|---|
| 001 | `The manager-develop agent definition` | ✓ 직접 | AC-012가 C1·C2·C3 3회 측정 |
| 002 | 같음 + `in any of the three copies` | ✓ 이중 | AC-001·002·003이 각 3회 |
| 003 | `The repaired manager-develop agent definition` | ✓ 직접 | AC-014가 3회 |
| 004·005 | 배치 정의 대상 — 에이전트 정의 아님 | 해당 없음 | codex 미러 없음(실측) |
| **006** | **`the agent definition`** (`manager-develop` 생략) | **문자적으로는 규약 밖** | AC-007이 C1·C2·C3에 걸쳐 6회 측정 |
| **007** | **`The agent definition`** (동일) | **문자적으로는 규약 밖** | AC-009(템플릿 전체) + AC-008(C1↔C2 델타)로 전이 결속 |
| 008 | `The completion-report instruction` | 규약 밖 | AC-007이 담당 |
| 009 | 파일 경로를 직접 명명(`.codex/` 포함) | 규약 불필요 | — |
| 010 | `distributed files` | 규약 불필요 | — |

**판정: 실질적으로는 새는 요구사항이 없다.** REQ-006·007·008이 축약형 주어를 쓰지만, 이들을 재는 AC가 세 사본 전부를 직접 측정하므로 한 사본만 고쳐서는 어느 것도 충족되지 않는다. REQ-007의 C1 축은 AC-009가 직접 훑지 않지만 AC-008이 C1↔C2 동등성을 강제하므로 C1에 브랜치명이 들어가면 델타가 깨져 RED가 된다.

다만 규약이 "이 절의 모든 요구사항에서 에이전트 정의를 가리키는 표현" 이라고 넓게 쓰였다면 이 대조 자체가 필요 없었다. **minor / optional** 로 기록한다.

---

## 4. AC-FSD-008이 못 덮는 C3를 직접 측정이 메우는가 (질문 4)

**메운다 — 저자가 든 다섯 개보다 실제로 하나 더 많다.**

C3에는 로컬 쌍이 없으므로(실측 확인) 쌍 동등성 판정 자체가 성립하지 않는다. 그 자리를 메우는 직접 측정은:

| AC | C3에 대해 재는 것 | 실측 baseline |
|---|---|---|
| AC-FSD-001 | 리터럴 4패턴 부재 | `3` |
| AC-FSD-002 | 전량 어휘족 부재 | `4` |
| AC-FSD-003 | S4 열거 부재 | `1` |
| AC-FSD-012 | STEP 4 구간 내 정본 문구 존재 + 구간 길이 | `0`, `14` |
| AC-FSD-014 | 판별자 부재 | `3` |
| **AC-FSD-007** | **STEP 5 구간 내 두 토큰 각각 존재 + 구간 길이** | **`0`, `0`, `10`** |
| AC-FSD-009 | 중립성 — 경로가 `internal/template/templates/` 전체라 `.codex/` 자동 포함 | diff 훑기 |
| AC-FSD-010 | `//go:embed all:templates` 범위에 포함 | — |

쌍 동등성이 제공했을 유일한 추가 가치는 "패턴에 걸리지 않는 무관한 내용 드리프트 탐지" 인데, C3는 애초에 **포맷이 다른 파일**(md vs toml)이라 쌍 동등성이 의미를 갖지 않는다. 저자의 판단은 옳다.

**단, `§D.2` 의 REQ-FSD-009 행이 직접 측정 AC로 `001·002·003·012·014` 다섯만 열거한다.** AC-FSD-007도 C3를 직접 재므로 실제 커버리지는 여섯이다. 커버리지 공백이 아니라 문서 기술의 과소 표기다 — **minor / optional**.

---

## 5. AC-FSD-007의 "같은 문장" 잔여 갭 (질문 5)

**판정: run-phase 진입을 막는 사유가 아니다. 수용 가능하다.**

근거 넷:

1. **기계 층이 이미 하중을 지고 있다.** 보증되는 것은 "두 토큰이 **각각**, **STEP 5 블록(10줄) 안에**, **세 사본 모두**" 다 — 실측 6개 측정점. iter-2의 구멍(`-e A -e B` 가 한쪽만으로 통과)은 프로브로 확인한 대로 닫혔다.
2. **위반 시 실제 손해가 작다.** 두 토큰이 같은 10줄 완료 보고 지시 블록 안에 있는 한, 그 블록에서 생성되는 보고는 판정 주체와 미결 상태를 **둘 다** 말한다. "같은 문장" 은 문체 강화이지 하중 부재가 아니다. VCI §1.1 surface 2가 막으려는 것 — 미결을 말하지 않는 위임 — 은 `PENDING at report time` 토큰의 단독 판정이 이미 막는다.
3. **갭이 은폐되지 않고 세 곳에 기록됐다** — `acceptance.md:207`(AC 본문), `:312`(§D.3 완료 정의), `progress.md:112`(잔여 미결). 관측되지 않은 것을 관측된 것처럼 적지 않았다.
4. **더 강한 판정을 원하면 run-phase에서 AC 수정 없이 얻을 수 있다.** 구현자가 두 토큰을 한 문장에 쓰고(REQ-FSD-006이 이미 요구) 그 줄을 `progress.md` 에 verbatim으로 남기면, 눈 확인이 인용 가능한 증거가 된다. 이는 AC 축소가 아니라 증거 추가이므로 소유권 교차에 해당하지 않는다.

세 회차를 거친 SPEC에서 이 항목을 차단으로 올리는 것은 M6이 경계하는 과잉 교정이다. **optional 로 분류한다.**

---

## 6. 회귀 점검 (질문 6) — 없음

iter-1에서 닫힌 7건(D1·D2·D4·D5·D6·D7·D8)과 iter-2에서 처리된 optional 2건(D11·D12)이 v0.3.0에서 그대로 유지된다:

- **D1** — `origin/develop` 이 배포 템플릿에 0회(파이프 없이 `rc=1` 재확인). REQ-FSD-007 존치, AC-FSD-009 패턴에 브랜치명 2종 포함.
- **D2** — AC-FSD-006이 구간 한정 확정 판정. 구간 길이 단언까지 추가돼 오히려 강화됐다.
- **D4** — 판별자 제거 확정, AC-FSD-014가 세 사본으로 확장. 토큰 의존성은 iter-2에서 리포 전역으로 확인했고 v0.3.0 `§A.2:73` 이 이를 명시한다.
- **D5** — AC-FSD-015 존치, baseline 각 `1` 재현.
- **D6** — 기준 SHA 고정 + 훑은 줄 하한 존치.
- **D7** — `tier: M` 존치, plan.md:13 · progress.md:5와 일치.
- **D8** — AC-FSD-005 자기완결 판정 존치.
- **D11 · D12** — `spec.md §D` 의 언어 중립성 분리와 `progress.md` 의 린트 파서 기록 모두 존치.

**AC-FSD-011은 손대지 않았다** — v0.2.0 본문과 문자 그대로 동일하다(귀속 명시 · 갭 선언 · SHOULD(지연) · §D.3의 "미결을 통과로 적지 않는다" 전부 유지). 물려받은 49분 40초를 이 SPEC이 측정한 값처럼 쓰지 않는 규율이 세 회차 내내 지켜졌다.

**AC-FSD-002 재배정 검증 (질문 2 부속)** — v0.2.0의 AC-FSD-002는 **템플릿 사본(C2)에 대한 리터럴 4패턴 검사**였다. v0.3.0에서 그 역할은 AC-FSD-001의 인자 목록으로 흡수됐고, 실측 출력에 C2 행이 `:3` 으로 존재함을 확인했다. **구 AC-FSD-002가 잡던 것 중 지금 놓치는 것은 없다.** 오히려 어휘족 판정이 새로 들어와 M-3을 잡는다. 재배정은 정당하고 AC 총수는 15로 유지된다.

---

## 신규 차단 결함

**N3. C3는 손으로 고치는 미러가 아니라 C2에서 생성되는 산출물인데, SPEC이 손으로 고치라고 지시한다 — `plan.md:90` · `plan.md:51` · `spec.md:134`(C1) · `spec.md:140`(C7) · `spec.md:61` — Severity: critical — Class: blocking**

### 측정된 사실

`internal/template/agentemit` 패키지가 `.codex/agents/moai/*.toml` 을 **`.md` 에서 결정론적으로 생성**한다. 패키지 문서(`agentemit.go:1-15`)의 표현 그대로:

> "the .md files ARE the neutral core... The Codex publication is a deterministic transform of (.md × manifest) emitting one TOML per agent under .codex/agents/."

그리고 골든 테스트(`golden_test.go:1-12`)가 그 드리프트를 지킨다:

> "These tests run the emitter against the REAL 11 template .md sources and pin the committed artifacts under templates/.codex/agents/moai/. They are the drift guard: **a hand-edited .toml** or a behavior change in the emitter or manifest **fails here** until regenerated via: `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/...` (the `make agents-emit` target wraps this)."

추가 실측:

- `Makefile:27-28` — `agents-emit: ## Regenerate the .codex/agents/moai TOMLs from the neutral .md layer` → `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission`
- `Makefile:23-25` — **`make build` 은 `agents-emit` 을 호출하지 않는다** (`templ-generate` → catalog hashes → `go build`). 따라서 AC-FSD-010의 `make build` 는 C3 재생성도, 스테일 탐지도 하지 않는다.
- 골든 테스트 현재 상태 — 타깃 실행(전량 아님, C4 위반 아님):
  ```
  $ go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1
  ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.416s
  ```
  즉 **지금 커밋된 C3는 C2로부터 생성한 결과와 정확히 일치한다.** C3의 STEP 4/STEP 5 구간 길이가 C1·C2와 `14`·`10` 으로 똑같은 것도 우연이 아니라 `TestEmitAllRoundTripsBodyVerbatim` 이 보장하는 본문 verbatim 통과의 귀결이다.
- SPEC의 인지 여부 — `grep -rn -i -e 'agentemit' -e 'agents-emit' -e 'emit' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/` 의 유일한 히트는 `AC-FSD-010 — 임베드 재생성` 과 `plan.md:107` 이며, **둘 다 `//go:embed` 이야기이지 생성기 이야기가 아니다.** 생성기·`make agents-emit`·골든 가드는 **0회** 등장한다.

### SPEC이 지시하는 것

- `plan.md:90` (M2) — "순서: C2(템플릿 md) → **C3(템플릿 toml)** → `make build` → C1" — C3를 **직접 편집 대상**으로 세운다.
- `plan.md:51` · `spec.md:140` (C7) — "C3는 TOML이다 … 문면을 옮길 때 문자열 리터럴 경계를 깨지 않는다" — 손으로 TOML을 쓰는 것을 전제한 주의사항이다. 생성기를 쓰면 이 걱정 자체가 존재하지 않는다.
- `spec.md:134` (C1) — "산출물은 **markdown · toml 문면 편집**뿐" — toml을 편집 산출물로 규정한다.
- `spec.md:61` — "형제 10개와 함께 있는 **정식 미러 트리**" — 미러라는 말은 맞지만 **유지 방식이 생성이라는 사실**을 말하지 않는다.

### 왜 차단인가 — 세 경로

| 경로 | 무슨 일이 일어나는가 |
|---|---|
| **A (올바름)** | C2 편집 → `make agents-emit` → C3 재생성 → 15개 AC 초록, 골든 초록. 정상 |
| **B (SPEC 지시대로)** | C2·C3 둘 다 손 편집. 손 편집이 생성기 출력과 바이트 단위로 일치하지 않으면 **`TestGoldenCommittedArtifactsMatchEmission` 이 깨진다.** 일치하면 우연히 통과 — 어느 쪽인지가 운에 달렸다 |
| **C (최악)** | C3만 손 편집하거나, C2만 고치고 재생성을 잊음 → **다음 `make agents-emit` 이 수리를 되돌린다.** 커밋 시점에는 AC가 전부 초록이므로 아무도 모른다 |

경로 C는 이 SPEC이 `plan.md §B7`("로컬 편집 후 템플릿 미러를 잊기 — 다음 `moai update` 가 되돌린다")로 **스스로 이름 붙인 위험**이다. iter-2에서 C3를 범위에 넣은 판단은 옳았는데, C3에 `.md` 쌍의 유지보수 모델을 그대로 가져다 붙이면서 이 위험을 새 축에서 재생산했다.

**15개 AC 중 어느 것도 A와 B/C를 구별하지 못한다.** AC들은 C3의 *내용*을 재는데, 손 편집된 C3는 재생성이 되돌리기 전까지 올바른 내용을 갖기 때문이다. 구별할 수 있는 유일한 기계 층은 골든 테스트인데 SPEC이 그것을 부르지 않고, `make build` 도 부르지 않는다.

### 필요한 수정 (전부 plan-phase 소관)

1. **`spec.md §A.1`** — C3 행에 "**C2로부터 `internal/template/agentemit` 이 생성**하며 원천은 C2" 를 명시한다. `.codex/` 트리에 `rules` 가 없다는 이미 적힌 사실 옆에 나란히 둘 자리다.
2. **`spec.md §C` C7 교체** — TOML 이스케이프 주의(무의미해짐) 대신 "**C3를 손으로 편집하지 않는다. C2 편집 후 `make agents-emit` 으로 재생성한다**" 를 [HARD]로 둔다. C1도 "markdown 편집 + 생성 실행" 으로 정정.
3. **`plan.md` M2 순서 교정** — `C2 편집 → make agents-emit → make build → C1 반영 확인`. C3는 편집 대상 목록에서 뺀다. §G 안티패턴에 "C3를 손으로 고치기" 를 추가한다.
4. **골든 가드를 증거 집합에 넣는다** — AC-FSD-010의 Then에 `go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` 종료 코드 0을 더한다(AC 총수 변화 없음). 또는 새 AC로 분리해도 16/16으로 예산 안이다. **어느 쪽이든 `acceptance.md` 편집이므로 plan-phase에서 해야 한다** — run-phase 행위자는 `acceptance.md` 본문을 고칠 수 없다(iter-1 D2에서 확립된 소유권 규칙).

덧붙일 만한 정합성 하나: `go test ./internal/template/agentemit/...` 은 **변경 범위로 좁힌 타깃 실행**이다. 이 SPEC이 복원하려는 바로 그 독트린을 따르는 검증이므로 C4와 충돌하지 않는다.

---

## 남은 결함 목록

| ID | 요지 | Severity | Class |
|---|---|---|---|
| **N3** | C3가 생성 산출물인데 SPEC이 손 편집을 지시. 생성기·`make agents-emit`·골든 가드 언급 0회. `make build` 는 재생성하지 않음. 15개 AC가 올바른 경로와 되돌려질 경로를 구별 못 함 | **critical** | blocking |
| N4 | 범위 규약이 `the manager-develop agent definition` 구에 결속되는데 REQ-006·007·008은 축약형 주어를 씀. 실질 결속은 AC가 담당하므로 새는 요구사항 없음 | minor | optional |
| N5 | `§D.2` REQ-FSD-009 행이 C3 직접 측정 AC를 5개로 적었으나 실제는 6개(AC-FSD-007 누락). 커버리지 공백 아님, 기술 과소 표기 | minor | optional |
| N6 | AC-FSD-007의 "같은 문장" 은 눈 확인 잔존. 세 곳에 기록됐고 실질 위험 작음 — §5 판정 | minor | optional |
| N7 | mutant M-4(전량 어휘족을 버린 우회 표현)가 생존. grep 기반으로 닫는 것이 원리적으로 불가하고, 요구하면 영구 RED 항목이 됨 | minor | optional |

**차단 결함은 N3 하나뿐이다.**

---

## 항목별 점수 (루브릭 앵커)

| 차원 | iter-1 | iter-2 | iter-3 | 밴드 | 근거 |
|---|---|---|---|---|---|
| Clarity | 0.50 | 0.75 | **0.80** | 0.75↔1.0 | 범위 규약 신설, 사본 축 재구조화, 계측점 4종 문자 고정(C6), AC별 근거 기술. 남은 모호성 1건: REQ-006·007·008의 축약형 주어(N4) |
| Completeness | 0.75 | 0.75 | **0.75** | 0.75 | 구조·프론트매터·범위 제외(H3 7개) 완비, C3 편입으로 iter-2 누락 해소. 그러나 C3의 **유지보수 기제**가 통째로 빠졌고 C7이 적용되지 않는 걱정을 문서화한다(N3) |
| Testability | 0.50 | 0.75 | **0.85** | 0.75↔1.0 | 구간 길이 3종을 Then 안으로, 토큰별 분리 판정, 어휘족 확장, C3 6중 직접 측정, 존재형 전부 RED-now 실측. 남은 비이진 2건: 같은 문장 눈 확인(N6), 재생성 상태를 재는 AC 부재(N3) |
| Traceability | 0.60 | 0.75 | **0.85** | 0.75↔1.0 | REQ 절 단위 매핑 완비, REQ-006·008이 AC-007의 서로 다른 토큰 판정으로 실질 분리, 구 번호 잔재 0건. 남은 공백: REQ-FSD-009의 "the repair shall land in that distributed copy" 절이 **착지 기제**를 재는 AC 없음(N3), §D.2 과소 표기(N5) |

조화평균: `4 / (1/0.80 + 1/0.75 + 1/0.85 + 1/0.85)` = `4 / 4.9363` = **0.8103 → 0.81**

**Δ(iter-1) = +0.24 · Δ(iter-2) = +0.06.** 세 회차 단조 상승, 회귀 없음.

**점수는 Tier M 기준선 0.80을 넘었다.** 그럼에도 FAIL인 것은 N3가 must-pass-equivalent가 아님에도 차단이기 때문이다: 지시대로 수행하면 커밋된 가드를 깨거나 수리가 되돌려지는데, 그것을 SPEC의 어떤 AC도 잡지 못한다. 집계 점수가 이 결함을 흡수하도록 두면 "AC 전부 초록 + 커밋 완료" 상태에서 drift가 살아 있는 완료 보고가 나온다 — VCI §1.1 surface 2의 관측되지 않은 완료 주장이 설계 단계에서 예약되는 형태다.

---

## Run-phase 진입 권고

**지금 상태로는 진입 불가. 다만 관문은 좁고 하나뿐이다.**

권고 경로 — **범위 한정 plan-phase 교정 후 델타 확인만으로 진입**:

1. manager-spec이 N3 수정 4항(위 § N3)만 적용한다. `spec.md §A.1`·§C, `plan.md` M2·§G, `acceptance.md` AC-FSD-010 — **다른 어떤 것도 건드리지 않는다.** 나머지 SPEC은 이 감사에서 0.81로 건전함이 확인됐고, 광범위 재작업은 검증된 부분을 흔들 뿐이다.
2. 확인은 **from-scratch 재감사가 아니라 열거된 결함 델타 검사**로 족하다: (a) `spec.md`·`plan.md` 에서 C3가 생성 산출물로 서술되는가, (b) `make agents-emit` 이 M2 순서에 들어갔는가, (c) 골든 테스트가 증거 집합에 들어갔는가, (d) AC 총수가 15 또는 16인가. 네 항목 전부 grep 한 줄로 판정된다.
3. N4·N5·N6·N7은 **run-phase로 넘겨도 무방하다.** N5는 문서 표기 정정 1줄이라 위 교정에 얹으면 되고, N6은 §5에서 판정한 대로 구현자가 한 문장에 쓰고 그 줄을 `progress.md` 에 verbatim으로 남기면 증거가 된다(AC 수정 아님, 소유권 교차 아님).

**PASS-with-debt를 권하지 않는다.** N3는 "덜 다듬어진 부분" 이 아니라 **잘못된 실행 지시**다. 부채로 넘기면 그 부채를 갚을 시점이 run-phase인데, run-phase 행위자는 `acceptance.md` 를 고칠 수 없고 M2 순서는 plan 소관이라 — iter-1의 D2와 정확히 같은 구조의 막다른 길이 된다.

**scope-reduction도 권하지 않는다.** C3를 다시 범위 밖으로 빼는 것은 iter-2의 N1으로 되돌아가는 것이고, 이번엔 "배포되는 drift를 알면서 남긴다" 가 문서에 남는다.

---

## Gaps (관측하지 않은 것)

- 대체 문면은 여전히 존재하지 않는다(run-phase 산출). §2의 mutant 판정은 **AC 집합의 판별력**에 대한 것이지 아직 쓰이지 않은 텍스트에 대한 것이 아니다.
- 전체 테스트 스위트를 실행하지 않았다. 실행한 유일한 Go 테스트는 `./internal/template/agentemit/...` 의 `-run TestGoldenCommittedArtifactsMatchEmission` 단일 타깃(0.416s)이며, `AGENTEMIT_UPDATE` 는 설정하지 않았다 — 읽기 전용 확인이고 어떤 파일도 재생성하지 않았다.
- `.codex/` 의 나머지 10개 `.toml` 이 각자 어떤 상태인지는 판정하지 않았다(`spec.md §D` 가 범위 밖으로 선언). `plan-auditor.toml` 이 전량 호출을 담고 있다는 사실만 iter-2에서 관측했다.
- `agentemit` 매니페스트(`agents-codex.yaml`)가 본문 변경을 어떻게 다루는지 세부는 읽지 않았다 — N3 판정에는 패키지 문서·골든 테스트 문서·Makefile 타깃·실행 결과로 충분했다.
- `.claude/skills/` 층의 같은 형태 지시는 세지 않았다(범위 밖 선언).
- 어떤 SPEC 산출물·에이전트 정의·규칙·템플릿도 수정하지 않았다.

**감사 산출물**: `/Users/goos/MoAI/moai-adk-go/.moai/reports/t301/plan-audit-iter3.md`
(iter-1 `plan-audit.md`, iter-2 `plan-audit-iter2.md` 는 그대로 둠)
**측정 트리**: `.claude/worktrees/t301` @ `d29b8942e`
