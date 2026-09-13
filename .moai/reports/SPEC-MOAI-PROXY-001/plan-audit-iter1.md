# SPEC Review Report: SPEC-MOAI-PROXY-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.70** (Tier L PASS 문턱 0.85)

감사 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`,
브랜치 `WT-moai-proxy-unified`, HEAD `d060e0d13`. 감사 시점 2026-09-10.
Reasoning context ignored per M1 Context Isolation — 저자의 추론 맥락은 받지 않았고,
판정은 아래 5종 아티팩트와 이 트리에 대한 직접 재측정만으로 내렸다.

부재 판정에는 전부 `/usr/bin/grep`을 썼다(이 셸의 `grep`은 ugrep 래퍼로 gitignore·바이너리
필터를 조용히 적용한다). 모든 부재 검사에 동일 범위 양성 대조군을 붙였다.

---

## 총평 — 먼저 말해 둘 것

이 SPEC은 증거 규율이 이례적으로 좋다. 각 AC에 **검증 가능성** 줄을 두어 PASS와 Gap을
분리했고, 부재 주장에 양성 대조군을 붙였으며, 미해결 교차 렌즈 모순(C1~C4)을 봉합하지
않고 그대로 기록했다. 인용한 `file:line`을 12곳 열어 봤고 **전부 성립했다**. 이 점은
감점 사유가 아니라 이 문서의 강점이므로 먼저 기록한다.

점수를 끌어내린 것은 단 하나의 축에 몰려 있다 — **추적성**이다. 25개 요구사항 중 7개가
어떤 AC에도 닿지 않는데, `spec.md` §F는 25개 전부가 대응한다고 스스로 단언한다. 그리고
수용 기준 예산에는 **여유 5칸이 남아 있다**. 즉 이 결함은 범위 압박의 결과가 아니라
누락이며, 대부분 현재 예산 안에서 고칠 수 있다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `/usr/bin/grep -c '^\*\*REQ-MP-' spec.md` → `25`.
  `REQ-MP-001`~`REQ-MP-025` 연속, 결번 0, 중복 0(`sort | uniq -d` 빈 출력), 3자리 zero-pad 일관.
- **[PASS] MP-2 GEARS 형식 준수** — 요구사항 층(`spec.md`의 `REQ-XXX`)에 대해서만 판정했다.
  25개 전부가 패턴 라벨을 달고 있다: Ubiquitous 16, Event-driven 5, Unwanted 3,
  Where 2(`REQ-MP-009` Where Windows, `REQ-MP-016` Where 측정 통과 — 둘 다 capability gate로
  적법). `acceptance.md`의 Given/When/Then은 검증 층의 올바른 형식이므로 여기서 감점하지
  않았다(Group 4에서 별도 채점).
- **[PASS] MP-3 YAML frontmatter 유효성** — `sed -n '2,14p' spec.md | cut -d: -f1` →
  `id title version status created updated author priority phase module lifecycle tags tier`.
  canonical 12필드 정순 + 선택 필드 `tier: L`. 거부 alias(`created_at`/`updated_at`/`labels`/
  `spec_id`) 검색 exit=1(0건). `phase: "v3.3.0 target"`은 릴리스 타깃이며 금지된 lifecycle
  토큰(`plan`/`run`/`sync`/`mx`)이 아니다.
  ID 정규식: `internal/spec/lint.go:1131` `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`에 대해 매치 True.
  (스키마 문서의 표 행은 단일 세그먼트 형태 `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`로 적혀 있어
  구현과 어긋나지만, 그 문서가 스스로 지목한 집행 지점은 `lint.go`이므로 구현이 정본이다.
  이 드리프트는 이 SPEC의 결함이 아니라 규칙 트리의 결함이다.)
- **[N/A] MP-4 언어 중립성** — 이 SPEC은 moai-adk-go 자체의 Go 내부 구현을 다루는
  단일 언어 범위이고 배포 template 표면을 규정하지 않는다. 자동 통과.
- **[PASS] MP-5 D7 교차 SPEC 정합** — 참조된 SPEC ID 10개의 status를 전수 판독했다.
  `superseded` 1건(`SPEC-MODEL-ROUTING-WIRE-001`)이 있으나 `research.md` §9라는 **명시적으로
  형제 SPEC에 이관된 근거 절** 안에서만 인용되고, 그 carve-out은 `spec.md` §G에 선언되어
  있다. BLOCKING 아님(관련 advisory A6 참조).
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `syscall`이 spec.md 4회 등장한다. 리터럴
  `//go:build`는 없으나 그보다 강한 것이 있다: `REQ-MP-009`(Where Windows)가 별도 계약을
  세우고, `design.md` §3.5가 `launch_exec_windows.go:31-60`의 spawn-and-wait 경로를 명시하며,
  `AC-MP-006`이 Windows 절반을 Gap으로 표시하고, `plan.md` §H에 검증 환경 미확보가 열린
  항목으로 남아 있다. 이 가드가 막으려는 실패(Windows 경로를 POSIX 계약으로 뭉개기)는
  명시적으로 차단되어 있다.
- **[FAIL] MP-7 clarification gate** — `plan.md` §H에 `[NEEDS CLARIFICATION]` **6건 미해소**.
  `/usr/bin/grep -c 'NEEDS CLARIFICATION' plan.md` → `7`(헤더 문장 1 + 마커 6).
  이것은 **점수와 무관한 게이트**이며 단독으로 FAIL을 강제한다. 다만 성격을 정확히
  적어 둔다 — 이는 저자의 결함이 아니라 **설계된 인계**다. 6건 모두 실재하는 차단 결정을
  가리키고 있고(패딩 아님), 해소 주체는 SPEC 저자가 아니라 Implementation Kickoff Approval
  게이트에서 `AskUserQuestion`을 도는 오케스트레이터다. 저자에게 개정을 돌려보낼 사유가
  **아니다**.
  배치 자체는 정확하다: 마커는 `plan.md`에만 있고 `spec.md`(0) / `acceptance.md`(0) /
  `research.md`(0) / `design.md`(0)는 깨끗하다.

---

## Category Scores (rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분 단일 해석. 감점 3: `spec.md:173-175`(미측정 `s` 의미론을 "사실"로 진술), `spec.md:181-182`(세 지점 중 하나의 판정 기구를 잘못 특징화), `spec.md:141-143`(REQ-MP-014가 세 의무를 한 요구사항에 압축) |
| Completeness | 0.95 | 1.0 근접 | HISTORY / Context / 범위 / 용어 / 요구사항 / 제약 / 성공 기준 / Out of Scope / 교차참조 전부 존재. Out of Scope는 `### Out of Scope — <주제>` H3 4개 각각에 구체 bullet 보유. Tier L 5종 아티팩트 전부 존재. §F가 거짓 자기주장을 담은 것만 소폭 감점 |
| Testability | 0.75 | 0.75 | AC 20개 전부 Given/When/Then + 검증 가능성 줄. 감점 2: `acceptance.md:101-104`(AC-MP-015가 이미 거짓), `acceptance.md:46`(AC-MP-005를 "검증 가능"으로 단정하나 전제인 `s` 경로가 미측정) |
| Traceability | 0.50 | 0.50 | "Multiple REQs lack ACs" 밴드에 정확히 해당. 25개 중 7개 무연결(커버리지 72%). 반대 방향(AC → 존재하지 않는 REQ)은 0건 |

**집계(조화평균, skeptical-stance 규정)**: 4 / (1/0.75 + 1/0.95 + 1/0.75 + 1/0.50) = **0.6994 ≈ 0.70**.
(산술평균으로 계산해도 0.7375로 문턱 0.85 미달이다. 결합 방식이 판정을 바꾸지 않는다.)

---

## 예산 대조 — 독립 재유도

| 축 | 이번 감사 실측 | Tier L 상한 | 이전 계수와 일치? |
|---|---|---|---|
| 요구사항 | **25** | 25 | 예 |
| 수용 기준 | **20** | 25 | 예 |

상한 출처: `.claude/rules/moai/workflow/spec-workflow.md:146-150` — 두 축에 **독립** 적용,
합산 아님. 이전 계수(REQ 25/25, AC 20/25)에 **동의한다**.

요구사항이 상한에 정확히 닿은 것은 `plan.md` §I가 스스로 적었듯 범위 압박 신호다.
다만 이번 감사가 덧붙이는 관측이 하나 있다 — **압박의 실제 비용은 상한 도달이 아니라
무연결 7건으로 나타났다.** AC 예산에는 5칸이 남아 있으므로, 이 결함은 예산 완화 없이
대부분 해소 가능하다. 상한을 이유로 무연결을 정당화할 수 없다.

---

## 배정 시험 충실도 — T01~T04, T11~T19

13개 전부가 `acceptance.md` §A에 AC-MP-001~013으로 **1:1 생존했다**. 축약·삭제·산문화
없음. 미배정 T05~T10 · T20은 §D 표에 이관처와 함께 가시적으로 남아 있고, T09는 이관이
아니라 `REQ-MP-016` 게이트 측정으로 유지된다고 명시된다. `handoff` §6의
"핵심 항목을 축약하거나 삭제하지 않는다" 요구를 충족한다.

다만 이 축에서 하나가 갈라져 나온다. `handoff` §6과 설계 보고서 §11의 **T08**은
"API fallback 0건, 선택한 endpoint·credential 종류 일치"인데, 운영자 결정으로 T08이
형제 SPEC으로 나갔다. 그런데 그 실질을 담은 `REQ-MP-022`(무단 fallback 금지)는 **이 SPEC에
남았다.** 결과적으로 의무는 남고 그것을 검증하던 시험만 나갔으며, 대체 AC가 작성되지
않았다. 이것은 "조용한 약화"라고 부를 정도의 은폐는 아니지만(라벨은 설계 보고서와
충실하다), 안전 요구사항 하나가 이 SPEC 안에서 검증 불가 상태로 남는 실질적 구멍이다.
아래 B1에 포함한다.

---

## Defects Found

### Blocking

**B1 — 요구사항 7개가 어떤 AC에도 닿지 않고, SPEC은 반대를 단언한다.**
`spec.md:216` — Severity: critical — Class: blocking

`spec.md` §F 첫 줄: "§D의 25개 요구사항 각각이 `acceptance.md`의 하나 이상 AC에 대응한다."
이 문장은 거짓이다. AC 헤더의 괄호 추적을 전수 파싱한 결과 커버 18개, **무연결 7개**:

| 무연결 REQ | 내용 | 왜 문제인가 |
|---|---|---|
| `REQ-MP-003` | `moai gpt` login/logout 닫힌 동사 집합, 집합 밖 첫 인수는 명시 오류 | 사용자에게 보이는 CLI 계약. AC-MP-014는 `--help`만 보고 동사 라우팅을 보지 않는다 |
| `REQ-MP-014` | `count_tokens` 정확도 표기 / `/v1/models` 범위 / 미등록 path 정책 | 세 의무 전부 무검증. AC-MP-013은 "등록되지 않은 path"만 스치고 앞의 둘을 다루지 않는다 |
| `REQ-MP-017` | GPT 정확한 4개 ID만 전달, 미지원 ID 대체 금지 | **금지 사항 직결 요구사항**(`handoff` §8). 검증 표면 0 |
| `REQ-MP-018` | GLM adapter 기존 UX 보존, beta 헤더 제거 이관 | 기존 사용자 회귀 축. 검증 표면 0 |
| `REQ-MP-022` | provider 간 무단 fallback 금지, 미선택 유료 경로 송신 금지 | **안전 요구사항**. 원래 검증하던 T08이 형제 SPEC으로 나갔고 대체가 없다 |
| `REQ-MP-009` | Windows supervisor 계약 | AC-MP-006의 *검증 가능성 줄* 괄호에만 등장. 헤더 추적 아님 |
| `REQ-MP-016` | Anthropic passthrough 측정 게이트 | §D 이관 표 각주에만 등장. 헤더 추적 아님 |

Required fix: AC 예산 여유 5칸에 `REQ-MP-017`·`REQ-MP-022`·`REQ-MP-003`·`REQ-MP-014`·
`REQ-MP-018`을 겨냥한 AC 5개를 신설한다(`REQ-MP-017`/`REQ-MP-022`는 registry 등록 목록
단언과 fallback 부재 가드로 각각 결정적 단위 시험이 가능하다 — `plan.md` M2/M8이 이미
그 작업을 계획하고 있으므로 AC만 없는 상태다). `REQ-MP-009`·`REQ-MP-016`은 각각
AC-MP-006 헤더와 §D 표에서 **헤더 괄호로 승격**한다(Gap 판정 자체는 그대로 유지). 그리고
`spec.md:216`을 실제 커버리지와 일치시키거나, 위 조치 후 참으로 만든다.

---

**B2 — AC-MP-015는 작성 시점에 이미 거짓이며 영구히 실패한다.**
`acceptance.md:101-104` (그리고 `spec.md:39-41`의 같은 측정) — Severity: major — Class: blocking

AC-MP-015는 "`/usr/bin/grep -rn "moai gg" .`를 실행하면 결과가 `0`행"을 PASS 조건으로
삼는다. 이 트리에서 지금 재측정한 결과:

```
$ /usr/bin/grep -rn "moai gg" . | wc -l
       9
$ /usr/bin/grep -rln "moai gg" .
./.moai/specs/SPEC-MOAI-PROXY-001/research.md
./.moai/specs/SPEC-MOAI-PROXY-001/spec.md
./.moai/specs/SPEC-MOAI-PROXY-001/acceptance.md
$ /usr/bin/grep -rn "moai cc" . | wc -l     # AC가 지정한 양성 대조군
    1584
```

9건 전부가 **이 SPEC 자신의 산출물**이다. 부재를 주장하는 문서가 주장 대상 토큰을 담아
자기 판정을 뒤집었다. 대조군 규율은 옳게 지켰지만 **탐색 범위에서 자기 자신을 빼지
않았다**. `spec.md` §HISTORY D3의 `→ 0`도 그 이름 붙인 트리에서 지금 재현되지 않는다.

Required fix: AC-MP-015를 요구사항이 실제로 말하는 것 — `REQ-MP-004`의 "help 출력과
command tree 어디에도 `gg`가 나타나지 않는다" — 로 다시 쓴다. 전 트리 문자열 grep은
애초에 그 요구사항의 계기가 아니었다. 구체 형태: (a) `moai --help` 출력에 `gg` 부재 +
동일 출력에 `cc` 존재(대조군), (b) cobra command tree 열거에 `gg` 부재 +
`helpGroupFrequency["launch"]` 항목 존재(대조군). 문자열 스윕을 굳이 남긴다면 범위에서
`.moai/specs/SPEC-MOAI-PROXY-001/`을 제외하고 그 제외를 AC 본문에 적는다.

---

**B3 — `REQ-MP-021`이 세 지점 중 하나의 판정 기구를 잘못 특징화했고, 그 하나가 쓰기
부작용을 가진 지점이다.**
`spec.md:177-182`, `design.md:151-157` — Severity: critical — Class: blocking

요구사항은 "현재 **부분 문자열 판정**을 쓰는 세 지점"으로 세 곳을 지목한다. 실측:

```
$ /usr/bin/grep -rn 'Contains(.*EnvAnthropicBaseURL' internal pkg cmd --include='*.go' \
    | /usr/bin/grep -v '_test.go'
internal/tmux/cg_detect.go:188:  if strings.Contains(env[config.EnvAnthropicBaseURL], "z.ai") {
internal/tmux/cg_detect.go:207:  if strings.Contains(os.Getenv(config.EnvAnthropicBaseURL), "z.ai") {
internal/hook/session_start_glm_guardrail.go:40:  return strings.Contains(os.Getenv(config.EnvAnthropicBaseURL), glmProcessEnvSubstring)

$ /usr/bin/grep -n 'z\.ai\|Contains' internal/hook/session_end.go
608:        if strings.Contains(line, "(attached)") {     # 무관 (tmux 출력 파싱)
```

두 가지가 어긋난다.

1. **`internal/hook/session_end.go:672-677`은 부분 문자열 판정이 아니다.** 같은 파일
   `:678` 주석이 기구를 명시한다 — "ANTHROPIC_BASE_URL is used as the GLM-active indicator:
   Claude Code's OAuth flow never sets this variable, so its **presence** reliably signals
   GLM mode." **존재 여부** 판정이다.
2. **부분 문자열 술어는 세 파일이 아니라 두 파일에 셋이다** — `cg_detect.go` 안에 서로
   다른 두 술어(`:188` settings-env, `:207` `hasGLMEnv` process-env)가 있다. 즉 이관 대상은
   부분 문자열 술어 3 + 존재 술어 1 = **4개 지점**이다.

이 구분이 사소하지 않은 이유: proxy 아래에서 두 기구는 **반대 방향으로** 깨진다.
부분 문자열 술어는 `127.0.0.1`에 `z.ai`가 없으므로 일관되게 `false`를 낸다(GLM 세션을
놓친다). 존재 술어는 세 launcher 전부가 `ANTHROPIC_BASE_URL`을 설정하므로 일관되게
`true`를 낸다 — **`moai cc`와 `moai gpt` 세션까지 GLM으로 오판하고**, 그 판정이 부르는
`cleanupGLMSettingsLocal`이 `settings.local.json`을 **쓴다**. `design.md:157`의
"이 판정은 전부 같은 답을 낸다"는 문자 그대로는 참이지만, 쓰기 부작용을 가진 오판을
읽기 오판과 같은 줄에 묶어 그 위험을 지운다.

Required fix: `REQ-MP-021`의 지점 목록을 기구별로 분리해 다시 쓴다 — 부분 문자열 술어
3개(`cg_detect.go`의 두 술어를 각각 세고, `session_start_glm_guardrail.go:40`)와 존재 술어
1개(`session_end.go`의 `cleanupGLMSettingsLocal`). 존재 술어에 대해서는 이관 요구에 더해
**proxy 아래에서 비-GLM 세션에 GLM teardown이 실행되지 않아야 한다**는 불변식을 명시하고,
`design.md` §5에 그 쓰기 부작용을 적는다. AC-MP-018에 "비-GLM proxy 세션 종료 시
`settings.local.json`이 변경되지 않는다"를 추가한다(해시 비교로 기계 판정 가능).

---

**B4 — 미측정 능력을 사실로 진술하고, 그 위에 선 AC를 "검증 가능"으로 단정한다.**
`spec.md:173-175`, `acceptance.md:44-46` — Severity: major — Class: blocking

`REQ-MP-020` 본문: "세션에만 적용하는 선택은 picker의 `s` 경로이고, Enter와 `/model <name>`은
사용자 기본값을 저장한다는 **사실**을 문서와 시험에 반영해야 한다."

같은 SPEC의 `plan.md:181-184`는 정반대로 적는다: "`s` 경로가 대상 Claude Code 버전에서
실제로 존재하는지 **확인되지 않으면** REQ-MP-020의 '전역 settings 불변' 계약을 사용자
안내만으로 지킬 수 없다." 설계 보고서 §3도 이 동작을 공식 문서 인용으로 서술할 뿐
측정하지 않았고, 같은 절에서 "동일 대화 `/model` 연속 전환 — 격리 TTY가 workspace trust
화면에서 멈추어 inference를 만들지 못함 / **아직 PASS 아님**"이라고 적는다.

그 위에서 `AC-MP-005`는 Given에 "`/model`을 `s` 경로로 전환한 뒤"를 넣고 검증 가능성을
"**검증 가능** — 해시 비교와 부모 env 비교 모두 기계적이다"로 단정한다. 해시 비교가
기계적인 것은 맞지만, 그 전제인 `s` 경로 자체가 미측정이므로 AC 전체는 검증 가능하지
않다. `AC-MP-003`도 같은 전제 위에 서 있으나 그쪽은 Gap을 병기하고 있어 정도가 덜하다.

Required fix: `REQ-MP-020`에서 "사실" 진술을 걷어내고 `REQ-MP-016`과 같은 게이트 형태로
바꾼다 — "**Where** picker의 `s` 세션 전용 선택이 대상 Claude Code 버전에 존재함이 실측된
경우에만 …" — 그리고 미측정 상태에서 전역 settings 불변을 무엇으로 보장할지(예: overlay
주입만으로 `~/.claude/settings.json`을 애초에 쓰기 대상에서 배제)를 명시한다.
`AC-MP-005`의 검증 가능성 줄을 "부분 — 해시·env 비교는 기계적이나 `s` 경로 존재는
미측정 전제이므로 그 부분은 **Gap**"으로 정정한다.

---

**B5 — `[NEEDS CLARIFICATION]` 6건 미해소 (clarification gate).**
`plan.md:153-184` — Severity: critical — Class: blocking (게이트)

MP-7 위반. 다만 성격은 위 네 건과 다르다. 6건 전부가 실재하는 차단 결정을 가리키고
있고(패딩 0건), 해소 주체는 저자가 아니라 kickoff 게이트의 오케스트레이터다.
검토한 6건: `moai gpt`의 kanban/factory 진입 형태, provider-activity signal의 구체 키·
경로, T09 측정을 M6 안에서 할지 별도 spike로 뺄지, `CredentialRef` 경계를 어느 SPEC이
먼저 고정할지, Windows 검증 환경 확보 여부, picker `s` 경로 존재 확인.

Required fix: SPEC 개정이 아니라 **오케스트레이터가 Implementation Kickoff Approval 이전에
`AskUserQuestion`으로 6건을 해소**한다. 그중 4번째(`CredentialRef`)와 6번째(`s` 경로)는
각각 아래 seam 항목 및 B4와 같은 결정을 가리키므로 함께 처리할 수 있다.

### Advisory

**A1** — `spec.md:195-196` / `plan.md:39-40`이 `ANTHROPIC_*` 가드를 "**빌드 실패**로 만드는
AST 가드" / "AST 가드가 **빌드를 깬다**"고 적는다. 실제로는 AST 기반이 맞지만(`go/ast`,
`go/parser` import) **테스트**다 — `internal/config/anthropic_env_ssot_test.go:83`
`TestNoBareAnthropicEnvVarLiteralsInProduction`. 실패하는 테스트는 빌드를 깨지 않는다.
`research.md:107-110`은 이를 정확히 테스트로 적고 있으므로 spec/plan 쪽이 드리프트했다.
구현자가 `go build`가 잡아 줄 것으로 믿고 테스트를 건너뛸 수 있다.
Required fix: 두 곳을 "변경 패키지 테스트에서 실패하는 AST 기반 가드 시험"으로 정정.

**A2** — `research.md:294` "`.moai/specs` 아래 `SPEC-*PROXY*` / `SPEC-*GPT*` — 없다.
착수된 작업이 없다." 재측정: `ls .moai/specs/ | /usr/bin/grep -E 'PROXY|GPT'` →
`SPEC-MOAI-PROXY-001`(양성 대조군: 전체 SPEC 디렉터리 832개). B2와 같은 자기 무효화
계열이다. 어떤 AC도 여기 걸려 있지 않아 advisory로 둔다.
Required fix: 탐색 범위에서 자기 디렉터리를 제외하고 그 제외를 주장 본문에 적는다.

**A3** — `spec.md:51-53`이 "`ANTHROPIC_BASE_URL`이 `http://127.0.0.1:<port>`를 가리키면,
세션 안에서 `/model`로 고른 모델 ID가 실제 요청의 `model` 필드에 실려 오고"를 직설법으로
서술한다. `plan.md:55-58`은 같은 명제를 "이 SPEC 전체가 서 있는 **가정**이며 연속 전환은
아직 관측되지 않았다"고 적는다. §A만 읽는 독자는 확립된 사실로 읽는다.
Required fix: §A에 "설계 보고서 §3이 단발 요청 2건까지 관측했고 연속 전환은 미관측"
한 줄을 붙인다. (M1이 이 가정을 첫 마일스톤으로 세운 판단 자체는 올바르다.)

**A4** — `REQ-MP-001`이 하드코딩 launcher 이름 목록을 "**두 곳**"이라 한다. 실측 세 곳이다:
`help_order.go:31`, `help.go:44-49`, 그리고 호출 지점 리터럴
(`cc.go:140` / `cg.go:93` / `glm.go:205`의 `spawnLaunch(..., "cc"|"cg"|"glm", ...)`).
`plan.md:88`은 세 번째를 올바로 포함하므로 계획이 요구사항을 앞선다.
Required fix: `REQ-MP-001`을 "최소 세 곳"으로 정정하고 spawn 경로를 명시.

**A5** — `research.md:80-81`이 launcher 리터럴을 `spawn.go:152-192 spawnLaunch(out,
"cc"|"cg"|"glm", args)`로 국소화한다. 실제 리터럴은 `spawn.go` 안이 아니라 **호출 지점**
`cc.go:140` / `cg.go:93` / `glm.go:205`에 있다(`spawn.go:152`는 `subcommand string` 파라미터를
받는 함수 정의일 뿐이다). 구현자를 잘못된 파일로 보낸다.
Required fix: 세 호출 지점을 file:line으로 직접 적는다.

**A6** — `research.md:239`가 "끊기는 교차 SPEC 계약 7건"에 `SPEC-MODEL-ROUTING-WIRE-001`을
포함하는데, 그 SPEC은 이미 `status: superseded` / `superseded_by: SPEC-AGENT-ARCH-V2-001`
이다. 이미 대체된 계약을 세면 형제 SPEC의 비용 추정이 부풀려진다. (덧붙여 그 SPEC은
agent 모델 라우팅이지 provider 요청 라우팅이 아니므로 이 SPEC과 도메인이 겹치지 않는다 —
이 점은 `REQ-MP-011`/`REQ-MP-019`와의 충돌 우려를 해소한다.)
Required fix: 7건 목록에 각 status를 병기하고 이미 대체된 것을 표시.

**A7** — `REQ-MP-001`과 `REQ-MP-021`이 **규범 요구사항 본문에 행 범위 인용**을 담는다
(`help_order.go`의 map 키, `cg_detect.go:185-207` 등). 요구사항은 SPEC 수명 내내 규범으로
남는데 행 번호는 파일이 바뀌는 즉시 낡는다. §HISTORY의 HEAD 고정이 완화하지만 없애지는
못한다.
Required fix: 요구사항 본문은 심볼 이름으로 지목하고(예: `hasGLMEnv`,
`hookProcessEnvHasGLM`, `cleanupGLMSettingsLocal`), 행 범위는 날짜와 HEAD가 붙어 있는
`research.md`에만 둔다.

**A8** — `REQ-MP-014`가 세 개의 독립 의무(count_tokens 정확도 표기 / `/v1/models` 범위 /
미등록 path 정책)를 하나로 압축했고 AC가 없다(B1에 포함). 요구사항 상한 25 도달이 이
압축의 원인일 가능성이 높다.
Required fix: B1의 AC 신설 시 세 의무를 각각 판정 가능한 문장으로 분해해 AC 본문에 담는다
(요구사항 분할은 상한 때문에 불가하므로 AC 쪽에서 흡수).

**A9** — `plan.md:181-182`가 설계 보고서의 "확인되지 않았다"를 `s` 경로 존재에 귀속시키는데,
보고서 §3의 그 문장은 **native picker의 disabled 행 지원**에 대한 것이다. 결론(=`s`는
미측정)은 옳지만 귀속이 어긋났다.
Required fix: 귀속을 정정하고, `s` 미측정은 보고서가 그 축을 아예 재지 않았다는 사실로
근거를 바꾼다.

### 검사했고 결함이 아닌 것 (기록)

- **R3 `moving-ref-ok` 예외 — 적법하다.** `internal/spec/lint_movingref.go:123-127`이
  `<!-- moving-ref-ok: <reason> -->`를 1급 예외 기구(REQ-MRG-002/003)로 정의하고,
  이유 없는 마커는 거부한다(`:276-278`). 저자의 이유는 비어 있지 않고, 고정 anchor
  (`d060e0d13`)가 같은 줄 또는 바로 앞줄에 있다. 결정적으로 같은 파일 `:74-77`이
  적는다 — "the detector reads SHAPE, never SUBJECT. It cannot apply the
  anchor-or-subject predicate, so every finding is a question put to a human, never a
  verdict." 저자는 그 질문에 정확히 답했다. 덧붙여 이 rule의 finding은 전부
  `Advisory: true`라 게이트를 열지 않으므로, 마커가 무언가를 통과시킨 것도 아니다.
  경고가 가리키던 실질(`0 0` 읽기의 시간 부패)은 `plan.md:32-33`이 착수 시점 재측정으로
  이미 처리한다.
- `moai spec lint SPEC-MOAI-PROXY-001` → `✓ No findings`, exit 0. (재실행함)
- 금지 사항(`handoff` §8) 준수: `gpt-5.3` 2회 등장은 **부재 가드 문맥에서만**
  (`spec.md:159` "후보에 넣지 않는다", `plan.md:74` "부재 가드"). `gpt-6-astra`는 4개 ID 중
  하나로 등록되나 기본값은 `gpt-5.6-sol`(`REQ-MP-019`)이므로 기본값 금지 준수.
  `--provider mixed` 0건. `claude_glm` 2건은 전부 기존 코드 인용이며 새 API 부활 아님.
  Codex credential은 `REQ-MP-025`/`AC-MP-020`/`design.md` §8에서 **읽기 전용 보존**으로
  다뤄지고 logout 구현은 형제 SPEC으로 이관. 과거 기록 일괄 치환은 §G에서 명시 제외.
- 인용 12곳 열람 결과 전부 성립: `launch_exec_posix.go:24-27`, `glm.go:363-392`,
  `help_order.go:31`, `cg_detect.go:185-207`, `session_start_glm_guardrail.go:37-40`,
  `session_end.go:672-679`, `launcher.go:57,130,158-171`, `kanban/record.go:23-24`,
  `codex_launcher.go:359`, `web/server.go:44,47,56`, `launch_exec_windows.go:31-40`,
  `envkeys.go` 상수 85개.
- 부재 주장 재검증(전부 양성 대조군 동반): PKCE/`code_verifier` 0건(대조군: oauth 포함
  파일 15개), `httputil.ReverseProxy` 비-테스트 0건, `login`/`logout` cobra 0건,
  `find . -type d -name "*proxy*"` 빈 결과.

---

## Recommendation

**FAIL — 저자에게 반려하되 개정 범위는 좁다.** 이 SPEC은 구조·형식·증거 규율에서
문턱을 넘어섰고, 실패는 추적성 한 축과 세 개의 국소 정정에 몰려 있다.

우선순위 순으로 고칠 것:

1. **B1** — AC 5개 신설(`REQ-MP-017`, `REQ-MP-022`, `REQ-MP-003`, `REQ-MP-014`, `REQ-MP-018`)
   + `REQ-MP-009`·`REQ-MP-016` 헤더 승격 + `spec.md:216` 정정. AC 예산 20 → 25로 상한에
   정확히 닿는다. 이것만으로 Traceability 0.50 → 1.0, 집계 약 0.70 → 0.86.
2. **B3** — `REQ-MP-021`의 지점 목록을 기구별로 재작성하고, 존재 술어의 쓰기 부작용
   불변식을 `AC-MP-018`에 추가.
3. **B4** — `REQ-MP-020`을 `REQ-MP-016` 형태의 측정 게이트로 전환, `AC-MP-005` 검증
   가능성 줄을 "부분/Gap"으로 정정.
4. **B2** — `AC-MP-015`를 help 출력 + command tree 판정으로 재작성.
5. **A1·A3·A4·A5** — 네 줄짜리 정정. 함께 처리.

**B5(clarification gate)는 저자 작업이 아니다.** 위 1~5가 반영되어 iter2가 PASS를 받더라도
6건이 열려 있는 한 Implementation Kickoff Approval로 진입할 수 없다. 오케스트레이터가
`AskUserQuestion` 라운드로 해소해야 하며, B4와 §H 6번 항목은 같은 결정이므로 묶어서
물을 수 있다.

**iter2 재감사 범위**: 위 열거된 결함 delta로 한정한다(전면 재감사 아님). 회귀 검사는
B1~B5 각각에 대해 수행한다.

---

## Claim

이 SPEC(`SPEC-MOAI-PROXY-001`)의 plan 단계 5종 아티팩트는 Tier L PASS 문턱 0.85에
미달하며 집계 0.70, 판정 FAIL이다. 근거는 (1) 25개 요구사항 중 7개가 어떤 수용 기준에도
닿지 않으면서 SPEC이 반대를 단언하고, (2) `AC-MP-015`가 작성 시점에 이미 거짓이며,
(3) `REQ-MP-021`이 이관 대상 세 지점 중 하나의 판정 기구를 잘못 특징화했고 그 지점이
쓰기 부작용을 가지며, (4) 미측정 picker 능력이 사실로 진술된 위에 "검증 가능" AC가
서 있고, (5) `[NEEDS CLARIFICATION]` 6건이 미해소인 것이다. 배정 시험 13개는 전부 충실히
생존했고, 열람한 인용 12곳은 전부 성립했다.

## Evidence

명령과 그 출력은 본문 각 항목에 인라인으로 배치했다. 판정을 지탱하는 핵심 6건:

```
$ git rev-parse --show-toplevel && git branch --show-current && git rev-parse --short HEAD
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
WT-moai-proxy-unified
d060e0d13

$ /usr/bin/grep -c '^\*\*REQ-MP-' spec.md ; /usr/bin/grep -c '^\*\*AC-MP-' acceptance.md
25
20

$ python3 <AC 헤더 괄호 전수 파싱>
REQ total 25 covered 18
ORPHAN REQs (no AC header trace): ['REQ-MP-003','REQ-MP-009','REQ-MP-014',
                                   'REQ-MP-016','REQ-MP-017','REQ-MP-018','REQ-MP-022']
AC refs to nonexistent REQ: []

$ /usr/bin/grep -rn "moai gg" . | wc -l          # AC-MP-015의 PASS 조건은 0
       9
$ /usr/bin/grep -rn "moai cc" . | wc -l          # AC가 지정한 양성 대조군
    1584

$ /usr/bin/grep -n 'z\.ai\|Contains' internal/hook/session_end.go
608:            if strings.Contains(line, "(attached)") {
  # 대조군: 같은 grep이 cg_detect.go와 session_start_glm_guardrail.go에서는 z.ai를 찾는다

$ moai spec lint SPEC-MOAI-PROXY-001
✓ No findings — all SPEC documents are valid   (exit 0)
```

## Baseline-attribution

모든 측정은 이번 실행에서 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`),
이 HEAD(`d060e0d13`), 2026-09-10에 수행했다. 다른 트리·다른 시점의 수치를 이번 실측으로
제시한 항목은 없다. 예산 상한 25/25는 `.claude/rules/moai/workflow/spec-workflow.md:146-150`
에서 직접 읽었다. frontmatter 스키마는 `.claude/rules/moai/development/spec-frontmatter-schema.md`
에서, 실제 집행 정규식은 `internal/spec/lint.go:1131`에서 읽었다. 설계 보고서 §3/§11과
핸드오프 §6/§8은 primary checkout의 `reports/`에서 읽기 전용으로 판독했다.
이전 감사의 계수(REQ 25/AC 20)는 인용하지 않고 독립 재유도한 뒤 일치를 확인했다.

## Gaps — 이번 감사가 관측하지 **않은** 것

- `design.md`의 Go 타입 스케치(`LaunchPlan`/`ModelEntry`/`RequestContext`)가 컴파일 가능한지,
  기존 `internal/cli` 타입과 이름 충돌하는지 — 정적으로도 확인하지 않았다.
- 설계 보고서 §1·§2·§4~§10 — §3과 §11만 읽었다. 나머지 절에 이 SPEC이 위배하는 결정이
  있는지 모른다.
- `internal/cli/launcher.go` 전량(55KB) — `research.md` §12가 스스로 gap으로 남긴 항목을
  감사도 메우지 않았다. proxy 기동에 관련된 추가 env/exec seam이 더 있을 수 있다.
- `moai gg`가 primary checkout이나 미병합 브랜치에 있는지 — 이 워크트리만 쟀다.
- 배정 13개 시험 각각이 설계 보고서 §11의 **합격 판정 열**까지 완전히 담았는지 —
  ID 생존과 제목 대응만 대조했고 판정 기준 문구의 의미 손실은 T08을 제외하고 개별
  대조하지 않았다.
- 형제 SPEC 두 개의 ID 예약 절차(`SPEC-MOAI-GPT-AUTH-001`, `SPEC-MOAI-CG-RETIRE-001`) —
  `.moai/specs` 부재만 확인했고 예약 레지스트리가 따로 있는지는 확인하지 않았다.
- `research.md` §9의 cg 철거 규모 수치(문서 303건, template 73건, 테스트 18파일 등) —
  형제 SPEC 소관이라 재측정하지 않았다. A6에서 지적한 "7건"만 검증했다.
- 런타임 검증 0건. 이 감사의 모든 항목은 정적 판독이다.

## Residual-risk

- **`s` 경로가 존재하지 않는 것으로 판명되면 이 SPEC의 사용자 가치 명제가 흔들린다.**
  B4를 게이트로 바꾸면 문서는 정직해지지만, `REQ-MP-020`(전역 settings 불변)과
  `REQ-MP-019`(세 launcher 공통 catalog)가 동시에 성립하는 경로가 남는지는 M1이 답해야
  한다. 이 위험은 SPEC 개정으로 제거되지 않는다.
- **B3의 존재 술어 오판은 `settings.local.json`을 쓴다.** 이관을 놓치면 `moai cc` 세션
  종료마다 GLM teardown이 돌아 사용자 설정을 조용히 변형한다. 회귀가 조용하므로 AC 없이
  구현 단계에 들어가면 발견 시점이 매우 늦다.
- **무연결 7건을 AC로 덮어도 검증 가능성이 곧 실측은 아니다.** 특히 `REQ-MP-016`
  (Anthropic passthrough)와 `REQ-MP-017`(GPT ID 접근성)은 계정 권한이 필요해 run 단계에서도
  Gap으로 남을 공산이 크다. AC를 신설하면서 그 Gap을 미리 선언해 두지 않으면, run 단계에서
  "AC가 있으니 PASS해야 한다"는 압력이 근거 없는 PASS를 만든다.
- **요구사항이 상한 25에 닿아 있어 위 정정 중 요구사항 분할이 필요한 것(A8)을 수용할
  여지가 없다.** iter2에서 새 요구사항이 필요해지면 상한 완화가 아니라 형제 SPEC으로의
  추가 이관을 먼저 검토해야 한다.
- **이 감사 자신이 자기 무효화 계열 결함을 재생산한다.** 이 보고서는 `moai gg` 문자열을
  담고 있으므로, 앞으로 이 트리에서 그 토큰을 세는 모든 스윕의 분모를 키운다. B2의
  수정안(범위 제외 명시)은 이 보고서에도 적용되어야 한다.
