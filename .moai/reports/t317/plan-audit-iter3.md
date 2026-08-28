# SPEC 감사 보고서 (iter-3, 최종): SPEC-AGENT-EMIT-LINEAGE-001

- Iteration: **3/3 — 상한 도달.** 이 판정으로 카드의 plan-phase 감사는 종료된다
- **Verdict: PASS**
- **Overall Score: 0.90** (조화평균) — Tier M PASS 임계 **0.80** 초과
- **점수 이동: 0.74 → 0.82 → 0.90, 단조 상승.** 회귀 없음 → STOP 에스컬레이션 조건 미충족
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317` @ `48eb945df`, branch `WT-agent-emit-lineage`, spec `version: 0.4.0`, `tier: M`
- 저작자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. `measurement.md` 와 저작자의 반론은 **증거로 읽되 권위로 받지 않았고**, 하중이 걸리는 인용은 전부 이 트리 이 실행에서 다시 쟀다
- 범위: iter-2 의 **D7·D8 델타 + Tier 재판정 + 회귀**. iter-1·iter-2 에서 닫은 D1-D6 은 깨지지 않았는지만 확인했다
- 감사 중 트리 변경: **없음.** 판정에 뮤테이션이 필요하지 않아 읽기 전용으로만 쟀다. 종료 시 `git status --short` 재확인 결과는 말미에 있다

---

## Must-Pass 결과

| 기준 | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | `grep -n '^\*\*REQ-AEL-' spec.md` → `:68,:70,:74,:76,:88,:90,:92` 정의 **7건**(001-007). 결번·중복 0, 3자리 padding 일관. 008 은 정의 0건이고 폐기 서술로만 3회 등장 |
| MP-2 GEARS 형식 준수 | **PASS — 요구 층(`REQ-XXX`)에 대해 판정** | 001 Event-driven(`When … the build shall`), 002 Unwanted(`shall not write`), 003 Ubiquitous+Unwanted(`shall occur only through` / `The build path shall never regenerate`), 004 Ubiquitous + Where(`Where the judgment point is not applicable, it shall report ok`) + While(`While … applicable, … shall exit failure`), 005 Event-driven, 006·007 Ubiquitous. IF/THEN 0건. `acceptance.md` 의 Given-When-Then 은 **검증 층**이므로 이 기준으로 감점하지 않았다(Group 4 소관) |
| MP-3 YAML frontmatter 유효성 | **PASS** | 정본 12필드 전부 존재(`spec.md:2-13`) + 선택 `tier: M`/`era: V3R6`. 거부 별칭 0건. `version: "0.4.0"` 인용 semver, `status: draft` 유효 enum, `phase: "v3.1.4 target"` 은 릴리스 타깃이라 금지된 lifecycle 토큰이 아니다 |
| MP-4 §22 언어 중립성 | **N/A (auto-pass)** | 이 저장소의 Go/Makefile 배선 한정 단일 언어 SPEC. 16개 프로그래밍 언어 도구 배선을 건드리지 않는다. 문서 게재를 `CLAUDE.local.md` 로 못박아 배포 템플릿 중립성도 침범하지 않는다 |
| MP-5 D7 교차 SPEC 정합 | **PASS** | `grep -o 'SPEC-…-[0-9][0-9][0-9]' spec.md | sort -u` → 자기 자신 1건뿐. retired/superseded 참조 0. BLOCKING 0 |
| MP-6 D8 크로스 플랫폼 규율 | **PASS (auto)** | `grep -rn 'syscall' <spec dir>` → 무출력(0건) |
| MP-7 clarification gate | **PASS** | `grep -rn 'NEEDS CLARIFICATION' <spec dir>` → 무출력. Tier M 이라 `research.md` 는 없고 `plan.md` 는 있으며 둘 다 훑었다. `progress.md §E.1` 도 "미해결 결정: 없음" 으로 일치 |

---

## 차원별 점수

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.85 | 0.75 | D7 수리로 iter-2 의 감점 사유가 사라졌다 — 적용가능성 술어가 **glob·기준점·문턱·금지**를 모두 이름으로 못박는다(`spec.md:78`). 남은 모호성 하나: 술어의 기준점이 "project root under check" 인데 doctor 배선이 실제로 넘기는 값은 `os.Getwd()` 원값이다(`internal/cli/doctor.go:180`). 요구 문면은 옳고 구현이 그 해석을 스스로 해야 하는 형태라 감점은 소폭에 그친다 → D9 |
| Completeness | 0.95 | 1.0 | 필수 절 전부, frontmatter 12/12, `### Out of Scope —` 5건 각각 구체 불릿 보유, **Tier M 산출물 3종 실재**(spec.md + plan.md + acceptance.md, + progress.md). `module:` 에 `internal/cli` 반영됨(`spec.md:11`). 감점은 영향 파일 열거가 스스로 "전수"라 적으면서 골든 고정본 3개를 빠뜨린 것 → D10 |
| Testability | 0.82 | 0.75 | AC 7건 전부 판정 명령·종료 코드·기수·원복을 동반한다. weasel word 0건. 4건이 뮤테이션 확립이고, CI 금지 게이트는 "도착 시점부터 초록"임을 **스스로 선언한 뒤** 반증용 뮤턴트를 제시해 공허 판정을 피했다. 감점 둘: 판정 명령에 run-phase 미확정 토큰(`embed-check`, `<항목명>`)이 남아 문자 그대로는 실행 불가하다는 점(SPEC 이 명시), 그리고 하위 디렉터리에서 실행하는 입력 상태를 어떤 AC 도 밟지 않는다는 점(D9) |
| Traceability | 1.00 | 1.0 | REQ 7 / AC 7. `spec.md:102-110` 사상표와 `acceptance.md` 헤딩 병기(`## AC-AEL-00N *(REQ-AEL-…)*`)를 **양방향 전수 대조**했다: 001→{AC1,AC4}, 002→{AC2,AC4}, 003→{AC2,AC5}, 004→{AC3}, 005→{AC3}, 006→{AC7}, 007→{AC6}. 역방향도 완전 일치. 미피복 REQ 0, 고아 AC 0. 분리 이후에도 깨지지 않았다 |

조화평균 = 4 / (1/0.85 + 1/0.95 + 1/0.82 + 1/1.00) = 4 / 4.4486 = **0.899 → 0.90**.

---

## D7 델타 판정 — 적용가능성 술어

### (a) 술어가 run-phase 가 바꿔치기할 수 없을 만큼 구체적인가 → **YES**

`spec.md:78` 은 네 가지를 각각 이름으로 적는다: glob(`internal/template/templates/.codex/agents/moai/*.toml`), 기준점(project root under check), 문턱(1건 이상 ↔ 0건), 그리고 **금지되는 대체 대상**(배포 프로젝트 루트의 `.codex/agents/moai/*.toml`). 술어를 "커밋 경로 단독에 건다"고 못박아 두었으므로, 다른 술어를 고르는 것은 해석이 아니라 문면 위반이 된다.

금지 조항이 세 표면에 모두 있음을 확인했다 — `spec.md:78`("shall not be substituted for the committed emission set"), `acceptance.md` AC-AEL-003 적용 불가 게이트("**함정** — … 대조 대상으로 삼지 않는다"), `plan.md` M1("**금지**다 — 배포 산출물이지 커밋 산출물이 아니어서"). run-phase 가 어느 문서를 먼저 읽어도 같은 금지에 닿는다. **airtight 로 판정한다.**

이 저장소에서 술어가 참임을 실측했다:

```console
$ ls internal/template/templates/.codex/agents/moai/*.toml | wc -l
      11
```

### (b) D3 의 공허성 봉투가 다시 열렸는가 → **NO**

두 부재가 서로 다른 축에 걸려 있다. 적용가능성은 **커밋 산출물 존재**에 걸리고(`spec.md:78`), "0 comparisons → pass 금지"는 `While the judgment point is applicable` 로 **범위가 좁혀진 채** 그대로 남아 있다(`:80`). 비교가 의미를 갖는 트리에서는 봉투가 온전하고, 비교 대상이 애초에 없는 트리에서만 `ok` 로 흐른다. 기수 조항(`shall exit non-zero when that count is lower than the committed artifact count`)도 같은 `While` 아래에 있다.

우회 경로 하나를 따로 확인했다 — 커밋 산출물을 지워 "적용 불가"로 위장하는 상태는 기존 단언이 막는다:

```console
$ grep -n "count" internal/template/agentemit/golden_test.go
284:	if count != 11 {
285:		t.Errorf("committed .codex/agents/moai carries %d TOMLs, want 11", count)
```

`ok` 로 흘리는 형태가 이 코드베이스의 선례와 같다는 SPEC 의 주장도 **문자 그대로 참**이다:

```console
$ grep -n "" internal/cli/doctor_mcp_version.go | sed -n '54,58p'
54:	if len(live) == 0 {
55:		check.Status = uikit.CheckOK
56:		check.Message = "no running moai MCP server recorded"
57:		if verbose {
58:			check.Detail = "the server stamps .moai/state/mcp-server/<pid>.json while it runs"
```

인용 `doctor_mcp_version.go:54-58` 은 오차 없이 정확하다.

### (c) AC-AEL-003 의 4개 단계가 v0.3.0 세 조항 + 적용 불가를 판정하는가 → **YES, 4/4**

| 조항 | 판정 단계 | 판정 가능성 |
|---|---|---|
| "reachable … as an explicit maintainer verb" | verb 도달 게이트(`grep -n '^embed-check:' Makefile` + `make embed-check` exit=0) | 이분 판정 가능 |
| "reachable … as a `moai doctor` check item" | doctor 도달 게이트 **[뮤테이션 확립]** — 뮤턴트 상태 exit≠0 / 원복 exit=0 + 필터가 한 항목만 남겼음을 카운터 합 1 로 확인 | 이분 판정 가능. 선례 형태를 실측으로 인용해 둠 |
| "shall not be attached to a CI build job" | CI 미부착 금지 게이트(`grep -rn` rc=1, baseline 측정 + 반증 뮤턴트 제시) | 이분 판정 가능 |
| 적용가능성 술어 + not-applicable 거동 | 적용 불가 게이트(스크래치 배포 → 항목 `ok`, `doctor_exit=0`) + 함정 명시 | 이분 판정 가능 |

CI 금지 게이트의 baseline 을 이 트리에서 재현했다:

```console
$ grep -rn "embed-check" .github/workflows/
(무출력 — 0 히트)
```

---

## D8 델타 판정 — 열거의 정직성과 Tier 재판정

### 열거 5건의 검증

| # | 파일 | 주장 | 판정 |
|---|---|---|---|
| 1 | `Makefile` | `.PHONY:16` 에 `agents-emit` 부재, `build:23` 선행은 `templ-generate` 뿐 | **확인.** `sed -n '14,30p' Makefile` → `.PHONY` 행에 `agents-emit` 없음, `build: templ-generate`, `agents-emit:` 타깃 별도 존재 |
| 2 | `internal/cli/doctor_<name>.go` (신규) | "doctor 항목은 파일 1개에 사는 것이 규약" | **부분 반증 — D11.** MoAI-ADK 그룹 13개 항목 중 자기 파일을 가진 것은 2건뿐(`doctor_mcp_version.go:39`, `doctor_disk.go:66`)이고 6건은 `doctor.go` 안에 인라인이다(`:417,:478,:495,:641,:941,:987`). "규약"은 과잉 주장이다 |
| 3 | `internal/cli/doctor_<name>_test.go` (신규) | 같은 규약 | 위와 동일. 기존 인라인 항목들은 `doctor_test.go`/`doctor_new_test.go` 를 공유한다 |
| 4 | `internal/cli/doctor.go` 등록 1행 | `moaiChecks` 슬라이스, 선례 `:201` | **확인.** `:201` = `{mcpServerVersionCheckName, func(v bool) DiagnosticCheck { return checkMCPServerVersion(cwd, v) }}` — 인용 정확 |
| 5 | `CLAUDE.local.md` | REQ-AEL-006 게재처 | 문서 결정 사항, 반증 대상 아님 |

### `doctor_golden_test.go` 는 정말 6번째가 아닌가 → **테스트 소스는 맞고, 그 고정본은 틀리다**

SPEC 의 근거("항목 이름 목록을 고정하지 않으므로")는 **테스트 소스에 한해 참**이다. `TestDoctor_CheckCount` 는 `len(checks) < 19` 만 단언하므로 항목이 하나 늘어도 깨지지 않는다(`doctor_golden_test.go:181-185`).

그러나 같은 파일이 doctor **출력 전체**를 골든으로 고정하고(`checkDoctorGolden(t, "doctor-light"/"doctor-dark"/"doctor-nocolor", got)`), 그 고정본은 항목 행을 한 줄씩 + 그룹별 카운터를 담는다:

```console
$ grep -n "MoAI-ADK" -A 16 internal/cli/testdata/doctor-nocolor.golden
13:│  MoAI-ADK ---
14-│    STATUS  CHECK                  MESSAGE
…
20-│    ok      MCP Server Version     no running moai MCP server recorded
…
26-│    8 ok, 3 warn, 0 fail
```

doctor 항목을 하나 더하면 이 세 고정본이 **반드시** 바뀐다(각 7,526 B, 바이트 동일 3본). 즉 편집 파일은 5건이 아니라 **8건**이다. → D10

### Tier 재판정 — **M 유지, 그리고 두 방향 모두에서 견고하다**

SSOT 밴드는 S `< 5 files`, M `5 - 15`(`spec-workflow.md § SPEC Complexity Tier`).

- 열거대로 5 → M
- 골든 3본을 더하면 8 → M
- D11 을 반영해 항목을 `doctor.go` 인라인으로 구현(파일 2·3 소멸)해도 3 + 골든 3 = 6 → **M**

세 시나리오 전부 밴드 안이므로 **Tier M 판정은 열거의 두 흠결에 흔들리지 않는다.** 승격 방향 자체도 이 리포가 금지한 안티패턴("오버헤드를 피하려 큰 SPEC 을 S 로")의 반대 방향이다.

### Tier M 산출물 의무 — 선언이 아니라 이행됐는가 → **이행됨**

- `acceptance.md` 실재(12,089 B), AC 헤딩 7건(`## AC-AEL-001` … `007`) 실측
- `spec.md §3` 은 포인터 + REQ↔AC 사상표만 남았고 AC 본문 0건(`:98-112` 전량 확인)
- `module:` 에 `internal/cli` 추가됨(`:11`)
- 임계 0.80 을 `plan.md §B.3` 과 `progress.md §E.1` 이 모두 명시 — 이 감사는 그 기준으로 판정했다

---

## 저작자 반론 3건에 대한 판정

### 반론 1 — "행을 내지 않는다"는 처방은 구현 불가하다 → **수용(정정 포함)**

근거 두 축이 측정으로 확인된다. 상태 열거는 셋뿐이고 "건너뜀"이 없다:

```console
$ grep -n "" internal/cli/uikit/types.go | sed -n '12,17p'
12:	// CheckOK indicates the check passed.
13:	CheckOK CheckStatus = "ok"
14:	// CheckWarn indicates a non-fatal issue.
15:	CheckWarn CheckStatus = "warn"
16:	// CheckFail indicates a critical failure.
17:	CheckFail CheckStatus = "fail"
```

등록 슬라이스에도 조건부 생략 경로가 없다 — `run` 은 이름 필터에 걸리지 않은 항목의 결과를 **무조건** append 한다(`doctor.go:229-243`).

정정을 하나 덧붙인다. Go 문법상 `moaiChecks` 를 조건부로 조립하는 것 자체는 **불가능하지 않다**. 그러므로 "구현 불가"보다 정확한 진술은 — **등록 계약을 바꾸지 않고는 닿을 수 없고, 바꾸면 골든 고정본이 환경 의존이 되어 깨진다**(위 D10 에서 잰 사실이 저작자가 든 근거보다 오히려 강하다). 어느 쪽이든 iter-2 의 그 처방은 채택 불가였다. **iter-2 처방을 철회한다.**

### 반론 2 — 처방 (나)는 거짓 딜레마였다 → **수용**

공유 로직 조항이 요구하는 것은 "두 표면이 같은 판정을 공유한다"이고, v0.4.0 이 실제로 쓴 것은 **표면 조건부**가 아니라 **입력 상태 조건부**다. `spec.md:78` 의 주어는 일관되게 "the judgment point"(두 표면 모두)이고, `moai doctor` 종료 코드 문장은 그 결과 서술이지 표면 한정이 아니다. 하나의 판정 함수가 적용가능성으로 분기하면 두 표면이 같은 입력에 같은 답을 낸다. **iter-2 의 (나) 논증을 철회한다.**

### 반론 3 — 인용이 구현이 아니라 문서를 가리켰다 → **수용, 양쪽 모두 확인**

```console
$ grep -n "" internal/cli/doctor.go | sed -n '47,48p;121p;140,146p'
47:	Long: "Run comprehensive system health checks including …
48:		"Exit codes: 0=no failing checks (warnings are advisory and do not fail the run), 1=one or more checks reported Fail.",
121:	return doctorExitStatus(failCount)
140:func doctorExitStatus(failCount int) error {
141:	if failCount == 0 {
142:		return nil
143:	}
144:	return &exitCodeError{
145:		code: 1,
146:		msg:  fmt.Sprintf("doctor: %d check(s) failed", failCount),
```

`:48` 은 `Long` 도움말 문자열 — **저작자 지적이 옳다.** 실제 승격은 `:121` + `:140-146` 이며 — **결론도 옳다**(Fail 1건 → exit 1). 두 반쪽 모두 확인했고, v0.4.0 본문(`spec.md:84`)은 이미 올바른 인용을 담고 있다.

---

## 저작자가 스스로 신고한 갭 2건 — 차단인가

### 갭 1 — `make build` 미실행 → **차단 아님. 오히려 옳은 절제다**

v0.4.0 이 새로 세운 주장 중 `make build` 실행을 요구하는 것은 **하나도 없다.** AC-AEL-001/003/004 의 `make build` 는 전부 run-phase 전제이지 plan-phase baseline 이 아니다. 반대로 실행했다면 `*_templ.go` 와 `catalog.yaml` 을 in-place 로 써서(실측 5) 감사 트리를 오염시켰을 것이다 — 이 SPEC 이 REQ-AEL-002 로 겨눈 바로 그 성질이다. 잔여 위험으로만 남긴다.

### 갭 2 — 파일 #2·#3 이 규약 추론 → **차단 아님. 다만 규약 주장은 과잉(D11)**

위에서 잰 대로 추론의 전제가 부분적으로 틀렸다. 다만 Tier 판정이 세 시나리오 전부에서 M 이므로 결과에는 하중이 없다. 열거 문장의 정정으로 닫힌다.

---

## 회귀 검사 (D1-D6, iter-1/iter-2 에서 닫힌 항목)

| 항목 | 상태 | 근거 |
|---|---|---|
| D1 결정 기록(doctor 편입) | **RESOLVED 유지** | HISTORY 0.3.0 행 + `spec.md:82` 호출 지점 주석 + `plan.md` M1 「자동 호출 지점 — 결정됨」. 근거 2 의 인용문을 원문 대조했다 — `doctor_mcp_version.go:6-8` 에 `"…leaves the host talking to the previous build. Nothing surfaces that…"` 축자 일치 |
| D2 REQ/AC-008 폐기 | **RESOLVED 유지** | 정의 0건. 008 은 `spec.md` 3회 / `acceptance.md` 1회 모두 폐기 서술 |
| D3 공허성 봉투 | **RESOLVED 유지, 범위만 좁혀짐** | `While … applicable` 아래로 이동. 위 D7 (b) 참조 |
| D4 REQ-AEL-002 주어 확대 | **RESOLVED 유지** | `spec.md:70` — "Every check this SPEC introduces — … of M2 and … of M1 alike" |
| D5 간접 사상 | **RESOLVED 유지** | 분리된 `acceptance.md` 에서도 7/7 헤딩에 REQ id 병기 |
| D6 AC-AEL-006 대조군 | **RESOLVED 유지** | Control(RED) 블록 + 판정을 레시피 실행 여부로 이동, make 3.81 비대칭 인용 주석 유지 |
| 번호·피복 | **무결** | REQ 001-007 / AC 001-007, 008 부재, 미피복 0 / 고아 0 |
| 마커 | **무결** | `NEEDS CLARIFICATION` 0건 |
| 범위 밖 침범 | **무결** | 5개 Out of Scope 영역 어디로도 REQ·AC 가 흘러가지 않는다. AC-AEL-003 의 CI grep 은 **부재 단언**이지 CI 신설이 아니고, 「공허성 대조」는 기록용이며 `agentemit` 밖을 건드리지 않는다 |
| `progress.md §E.1` 정합 | **무결** | Tier M, REQ 7 / AC 7, 임계 0.80, 산출물 3종, "미해결 결정: 없음" — 본문과 전량 일치 |

**정체(stagnation) 없음.** 3회 모두 서로 다른 결함 집합이 나왔고, 이전 회차 결함이 미해결로 남은 것은 0건이다.

---

## 발견된 결함 (구조화 목록)

D9. anchor-ambiguity — `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/spec.md:78` — 적용가능성 술어의 기준점이 "project root under check" 인데, doctor 배선은 모든 항목에 `os.Getwd()` 원값을 넘긴다(`internal/cli/doctor.go:180`). 이 저장소의 하위 디렉터리(예: `internal/cli/`)에서 `moai doctor` 를 돌리면 glob 이 0건을 잡아 **적용 가능한 트리가 적용 불가로 판정된다** — D3 이 닫은 공허성의 좁은 재발 형태다. 요구 문면 자체는 옳고(project root 를 해석하라고 이미 적혀 있다), 빠진 것은 그 문면을 확인하는 AC 다 — 어떤 AC 도 이 입력 상태를 밟지 않는다. — Severity: major — Class: optional — Required fix: AC-AEL-003 에 한 줄 게이트 추가 — `( cd internal/cli && "$REPO/bin/moai" doctor --check "<항목명>" )` 가 저장소 루트에서와 **같은 판정**을 내는지. 구현은 프로젝트 루트를 marker 로 거슬러 올라가 해석한다

D10. enumeration-omits-golden — `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md §B.1` — "전수 열거 5건" 이 `internal/cli/testdata/doctor-{light,dark,nocolor}.golden` 3본을 빠뜨린다. 세 고정본은 doctor 항목 행 전체와 그룹별 카운터(`8 ok, 3 warn, 0 fail`)를 담으므로 항목 추가 시 **반드시** 재생성된다. "`doctor_golden_test.go` 는 여섯 번째 편집 파일이 되지 않는다"는 서술은 테스트 **소스**에 한해 참이고 그 **고정본**에는 거짓이라, run-phase 가 골든 3본 실패에 놀랄 여지를 남긴다. 실제 편집 파일은 8건이며 Tier 는 M 그대로다 — Severity: minor — Class: optional — Required fix: §B.1 표에 골든 3본을 행으로 추가하고 총계를 8 로 정정, 해당 문장을 "테스트 소스는 항목 이름을 고정하지 않지만 골든 고정본은 재생성이 필요하다"로 교체

D11. convention-overclaim — `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md §B.1` — "이 리포의 doctor 항목은 파일 1개에 사는 것이 규약이다"는 측정과 어긋난다. MoAI-ADK 그룹 13개 중 자기 파일 보유는 2건(`doctor_mcp_version.go:39`, `doctor_disk.go:66`)이고 6건은 `doctor.go` 인라인이다(`:417,:478,:495,:641,:941,:987`). 파일 #2·#3 은 "규약"이 아니라 **선택 가능한 형태 중 하나**다. Tier 판정에는 하중이 없다(어느 형태든 M) — Severity: minor — Class: optional — Required fix: "규약이다" → "이 리포에 두 선례가 있는 형태이며 `doctor.go` 인라인 선례도 6건 있다. 어느 쪽을 고르든 Tier 는 M" 으로 정정

D12. citation-span — `.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/spec.md:86` — `golden_test.go:285 의 if count != 11` 인용에서 `if` 는 `:284`, 인용된 `want 11` 문자열은 `:285` 다. 조항 전체를 메시지 행 번호로 가리켰다 — Severity: minor — Class: optional — Required fix: `:284-285` 로 표기

**must-pass 실패 0건. blocking 등급 결함 0건.** D9-D12 는 전부 optional 이며, 어느 것도 구현을 막지 않는다.

---

## 판정 근거 요약 — 이 SPEC 은 run-phase 에 들어가도 되는가

**들어가도 된다.** 남은 결함 4건은 모두 (i) 요구·수락의 정합성을 깨지 않고, (ii) 구현 순서를 바꾸지 않으며, (iii) 착지 조건을 흐리지 않는다. D9 만이 실질을 가진 항목인데, 그마저도 **요구 문면은 이미 옳고** 빠진 것은 그 문면을 확인하는 AC 한 줄이다 — run-phase 가 M1 에서 자연히 밟는 자리이므로 부채로 지고 갈 수 있다. D10-D12 는 `plan.md`·`spec.md` 서술의 정확도 문제이고 Tier 판정에는 하중이 없다.

M6 규율에 따라 덧붙인다: 위 4건 중 어느 것도 FAIL 을 만들기 위해 승격하지 않았다. 이 SPEC 의 착지 조건 — 뮤테이션으로 죽는 가드 4건, 기수 게이트, 부재 게이트, 4개 도달/금지 게이트 — 는 이 카드가 겨눈 공허성을 실제로 봉쇄한다.

---

## Gaps — 관측하지 않은 것

- `make build` / `make agents-emit` / `make embed-check` 를 **실행하지 않았다.** 트리를 변형하는 명령이고 판정에 필요하지 않았다. 따라서 AC-AEL-001/003/004/005 의 GREEN 경로는 이 감사에서 **실증되지 않았다** — run-phase 의 몫이다
- `moai doctor` 를 스크래치 배포 프로젝트에서 다시 돌리지 않았다. iter-2 가 잰 값(`doctor_exit=0`)을 재측정 없이 판정 근거로 쓰지 않고 **참고로만** 두었다. 이 회차의 판정은 코드 대조(`doctor.go:121`·`:140-146`·`:180`, `uikit/types.go:12-17`, `doctor_mcp_version.go:54-58`)로 성립한다
- D9 의 하위 디렉터리 시나리오는 **코드 독해로 도출**했고 실제 `moai doctor` 실행으로 재현하지 않았다(검사 자체가 아직 없어 재현 대상이 존재하지 않는다). 근거는 `doctor.go:180` 이 `os.Getwd()` 원값을 넘긴다는 관측 사실이다
- `internal/cli` 패키지 테스트를 돌리지 않았다 — 이 SPEC 은 아직 코드를 바꾸지 않는다

## Residual risk

- 골든 3본 재생성이 run-phase 에서 누락되면 `internal/cli` 테스트가 붉어진다. D10 이 이를 미리 알리지 못하는 상태였다
- 판정 명령의 미확정 토큰(`embed-check`, `<항목명>`)이 run-phase 에서 다른 이름으로 확정되면 AC 문면과 갈린다. SPEC 이 그 사실을 명시하므로 함정은 아니지만, 확정 후 AC 문면 갱신이 필요하다
- 추출 경로(`moai init <scratch>`)는 무거운 쓰기 동작이다. REQ-AEL-002 가 이를 직접 구속하지만 실제 안전성은 구현이 `$TMPDIR` 밖으로 새지 않는지에 달려 있다 — run-phase 의 §E 증거로 확인되어야 한다

---

## 감사 중 트리 변경

```console
$ git status --short
?? .moai/reports/t317/
?? .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/
```

감사 시작 시점과 동일하다(추적 파일 수정 0, 삭제 0). 이 회차는 뮤턴트를 심지 않았으므로 원복할 것도 없다.

---

## Recommendation

**PASS (0.90 ≥ Tier M 임계 0.80).** run-phase 진입을 권고한다.

1. **D9 를 M1 착수 시 첫 항목으로 흡수한다.** 적용가능성 술어의 기준점을 프로젝트 루트 marker 해석으로 구현하고, AC-AEL-003 에 하위 디렉터리 게이트 한 줄을 더한다. 이 카드가 근절하겠다고 선언한 공허성의 마지막 좁은 통로다
2. **D10·D11·D12 는 `plan.md`·`spec.md` 서술 정정으로 닫는다.** 판정·순서·착지 조건에 영향이 없으므로 착수를 막지 않는다. M1 커밋에 함께 실으면 충분하다
3. **골든 3본 재생성을 M1 종료 조건에 넣는다** — `UPDATE_GOLDEN=1` 계열로 갱신하고, 갱신 diff 가 새 항목 행 + 그룹 카운터 증가에 한정되는지 확인한다
4. 위임은 **Tier M 이므로 Section A-E 5절 템플릿 필수**다(`manager-develop-prompt-template.md` § Applicability). §E 에 AC-AEL-003 의 6게이트 판정 결과를 개별 행으로 요구한다
5. Implementation Kickoff Approval 은 이 PASS 로 대체되지 않는다 — plan→run HUMAN GATE 는 그대로 남는다

**iteration 상한 도달.** 이후 SPEC 이 다시 바뀌면 그것은 재감사가 아니라 새 감사 사이클이며, 위 4건이 그대로 남아 있어도 blocking 이 아니므로 재감사 없이 run-phase 가 진행될 수 있다.
