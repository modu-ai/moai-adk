# SPEC Review Report: SPEC-MOAI-GATEWAY-001

Iteration: 2/3
Verdict: **FAIL**
Overall Score: **0.80** (Tier L PASS 문턱 0.85, iter1 0.70 대비 **+0.10**)
STOP 신호: **발령하지 않음** — 점수가 iter1보다 올랐다(retry-loop 계약의 하락 조건 불충족).

감사 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`,
브랜치 `WT-unified-gateway`, HEAD `d060e0d13`. 감사 시점 2026-09-10.
Reasoning context ignored per M1 Context Isolation — 오케스트레이터가 전달한 처분 대응표와
해석(예: "B3 저자가 옳다", "research.md:353은 오탐")은 결론으로 채택하지 않고 전부 이 트리에서
다시 쟀다.

부재 판정은 전부 `/usr/bin/grep`으로 했고 동일 범위 양성 대조군을 붙였다. 0.1.0 산출물 원문은
트리에서 사라졌으므로, iter1 감사 에이전트가 당시 `Read`로 받은 **도구 결과**(transcript의
tool_result 레코드)에서만 복원했다. 저자(`spec-proxy-core`)의 transcript는 읽지 않았다.

---

## 총평

iter1의 핵심 결함이던 추적성은 실제로 고쳐졌다(25/25 REQ 커버, 역참조 0). B2·B4·B5와 A1~A9도
인용 지점을 열어 확인했고 해소되었다. 점수 상승(+0.10)은 대부분 이 축에서 나왔다.

그런데 B3을 고치는 과정에서 **새 결함 셋이 같은 영역에 생겼다.** 공통 원인은 하나다 —
"gateway launch가 `settings.local.json`에 무엇을 쓰는가, 무엇을 지우는가"를 SPEC이 정하지 않은
채, 그 파일에 대한 훅의 동작만 launch 시점 signal로 막았다는 것이다. 여기에 REQ 번호 결번
(must-pass)과 Windows 종료 코드 의무의 조용한 축소가 더해져 FAIL이다.

비차단 결함 목록은 길지만(M6), 판정을 FAIL로 만든 것은 아래 **차단 5건**뿐이다.

---

## Must-Pass Results

- **[FAIL] MP-1 REQ 번호 일관성** — 25개 정의, 중복 0, zero-pad 일관. 그러나 **`REQ-MG-007`
  결번**이다. 독립 파싱 결과 `nums: [1..6, 8..26]`, `missing in 1..26: [7]`. 폐기와 재사용 금지는
  `spec.md:63-70` HISTORY에 선언되어 있지만, 결번 자리인 `spec.md` §D.2(`:175` REQ-MG-006 다음
  `:180` REQ-MG-008) 본문에는 아무 표시가 없다. MP-1의 문구는 "Even one gap = FAIL"이며, 판정
  계층은 요구사항 층 §D다. **이 항목만 보면 한 줄로 고칠 수 있다**(G2-MP1). 판정을 FAIL로
  만든 이유가 이것 하나는 아니다.
- **[PASS] MP-2 GEARS 형식** — 요구사항 층(`spec.md` §D의 `REQ-MG-*`)에 대해서만 판정했다.
  25개 전부 패턴 라벨 보유: Ubiquitous 12, Event-driven 6, Unwanted 3, Where 2(`REQ-MG-009`,
  `REQ-MG-016`), 복합 2(`REQ-MG-020` Ubiquitous+Where, `REQ-MG-021` Ubiquitous+While — GEARS
  복합절로 적법). `acceptance.md`의 Given/When/Then은 검증 층의 올바른 형식이므로 여기서 보지
  않았다.
- **[PASS] MP-3 YAML frontmatter** — `sed -n '1,15p' spec.md | cut -d: -f1` → `id title version
  status created updated author priority phase module lifecycle tags tier` (12필드 정순 + `tier`).
  거부 alias 검색 exit=1. 형제 산출물 `status:` 검색 exit=1(무상태 규칙 준수).
- **[N/A] MP-4 언어 중립성** — moai-adk-go 자체 Go 내부 구현을 다루는 단일 언어 SPEC. 배포
  template 표면을 규정하지 않는다.
- **[PASS] MP-5 D7 교차 SPEC 정합** — 5종 산출물에서 추출한 SPEC ID 13개의 status를 전수 판독.
  `superseded` 1건(`SPEC-MODEL-ROUTING-WIRE-001`, `superseded_by: SPEC-AGENT-ARCH-V2-001`)은
  `research.md:414`에 status와 "이미 대체" 문구로 명시 조정되어 있다 → BLOCKING 아님. SHOULD 3건:
  `SPEC-MOAI-GPT-AUTH-001`·`SPEC-MOAI-CG-RETIRE-001`(둘 다 "제안"으로 표기), `SPEC-MOAI-PROXY-001`
  (iter1 보고서 경로에만 등장하는 옛 ID). 단, **이 SPEC이 참조하지 않는** 두 SPEC과의 계약 충돌을
  별도로 발견했다(G2-B3). D7 동사는 미참조 SPEC을 보지 못하므로 MP-5가 아니라 차단 결함으로 올린다.
- **[PASS, 판단] MP-6 D8 크로스플랫폼** — 동사를 문자 그대로 돌리면 `syscall` 6회, `//go:build`·
  EXCL 0회로 BLOCKING이 찍힌다. iter1과 같은 판단으로 PASS를 유지한다. `spec.md`의 `syscall`
  언급은 전부 **기존** `internal/cli/launch_exec_posix.go`를 보존한다는 내용이고, 그 파일은 이미
  `:1 //go:build !windows`(짝 `launch_exec_windows.go:1 //go:build windows`)를 갖는다. lessons #21이
  막으려는 "새 syscall 도입 + 빌드 태그 누락"은 `spec.md` 본문에 없다. 새 POSIX 전용 호출은
  `design.md:151`(`kill(pid, 0)`)에만 있으며, advisory G2-A7로 따로 적었다.
- **[PASS] MP-7 clarification gate** — `/usr/bin/grep -c 'NEEDS CLARIFICATION'` → `plan.md:0`,
  `research.md:0`(나머지 4개 파일도 0). 대조군: 같은 파일에서 `REQ-MG` 토큰 `plan.md:13`,
  `research.md:2`.

---

## Category Scores (rubric-anchored)

| 차원 | iter1 | iter2 | 증감 | 밴드 | 근거 |
|---|---|---|---|---|---|
| Clarity | 0.75 | **0.75** | 0 | 0.75 | B4·A3 해소로 올라갈 몫이, 새로 생긴 해석 분기 두 곳에 상쇄되었다 — `spec.md:115-119`의 "env 주입 단계만 교체"가 `removeGLMEnv` 존속 여부를 정하지 않음(G2-B1), `design.md:234-235`의 "항상 거짓"(G2-B3) |
| Completeness | 0.95 | **0.90** | −0.05 | 1.0 근접 | 필수 절·Out of Scope H3 5개(`spec.md:321,332,342,348,353`) 모두 존재. 감점: stale GLM 키 정리 책임이 어디에도 배정되지 않음(G2-B1), HISTORY 대응표 누락(G2-A3) |
| Testability | 0.75 | **0.70** | −0.05 | 0.75와 0.50 사이 | `AC-MG-018`이 올바른 구현에서도 적색(G2-B2, 개정으로 새로 생김). 판정 입력이 정의되지 않은 AC 2건(`AC-MG-021` (c), `AC-MG-014` (d)) |
| Traceability | 0.50 | **0.90** | +0.40 | 1.0 근접 | 헤더 괄호 기준 25/25 커버, 미정의 REQ 참조 0. 감점: `AC-MG-014`가 어떤 REQ에도 근거가 없는 배지 단언을 운반(G2-A1) |

**집계(조화평균)**: 4 / (1/0.75 + 1/0.90 + 1/0.70 + 1/0.90) = **0.8026 ≈ 0.80**.
산술평균 0.8125. 어느 쪽이든 문턱 0.85 미달이고, MP-1 FAIL만으로도 판정은 FAIL이다.

**Windows 항목의 점수 처리(울타리).** `REQ-MG-009`, `AC-MG-006`의 Windows 절반, `plan.md` M4는 운영자가
다시 결정하는 중이므로 위 점수에 넣지 않았다. 현재 문구 그대로 채점하면: `REQ-MG-009`는 GEARS
Where로 적법하고, `AC-MG-006` Windows 절반은 **공허하지 않다**(판정이 없으면 Gap으로 남는다고
스스로 적는다). 다만 CI 판정이 없다는 결론은 읽지 않은 워크플로를 빠뜨렸다(G2-A4). 반영하더라도
Clarity 밴드는 바뀌지 않고, Testability에도 영향이 없다(이미 Gap으로 선언되어 있다).

---

## iter1 결함 처분

| iter1 | 처분 | 근거 |
|---|---|---|
| **B1** 무연결 REQ 7건 | **RESOLVED** | 독립 파싱: `covered by header parens: 25`, `orphan REQs: []`, `AC refs to undefined REQ: []`. `REQ-MG-009` → `acceptance.md:54` 헤더, `REQ-MG-016` → `acceptance.md:175` 헤더. `spec.md:311` §F 문장이 이제 참이다 |
| **B2** AC-015 자기 무효화 | **RESOLVED** | `acceptance.md:123-133` — help 출력 + cobra 트리 판정, 양쪽 모두 `cc` 대조군, 문자열 스윕은 두 경로 제외를 명시. `spec.md:40-46` D3는 작성 전 측정으로 한정 |
| **B3** 판정 기구 오특징화 | **PARTIALLY RESOLVED** | 존재 술어는 정확히 고쳤다(아래 편차 1 판정). 남은 것: `cg_detect.go`의 두 술어도 토큰 존재 disjunct를 가지는데 여전히 "부분 문자열 술어"로 분류되어 있다(G2-B3). 수정이 새 결함 G2-B1·G2-B2를 낳았다 |
| **B4** 미측정 능력의 사실 진술 | **RESOLVED** | `spec.md:241-248` Where 게이트화, 전역 settings 불변은 쓰기 배제로 보장. `acceptance.md:51-52` Gap 병기, `:35-37` AC-MG-003에도 게이트 Gap 추가 |
| **B5** clarification 6건 | **RESOLVED** | 마커 0(MP-7). `plan.md:204-244` §H 결정 6건. kickoff 게이트는 생략되지 않는다고 명시(`plan.md:36-37`, `:207`) |
| **A1** 가드=빌드 | **RESOLVED** | `spec.md:278-280`, `:301-302`, `plan.md:45-47`, `acceptance.md:160-163` |
| **A2** 자기 무효화 부재 주장 | **RESOLVED** | `research.md:475-477`에 자기 제외 명시. 재측정 `ls .moai/specs \| /usr/bin/grep -cE 'PROXY\|GPT\|GATEWAY'` → `1`(GATEWAY 자신). 대조군 라벨 소오류는 G2-A10 |
| **A3** §A 가정 표시 | **RESOLVED** | `spec.md:121-123` |
| **A4** launcher 이름 세 곳 | **RESOLVED** | `spec.md:150-153`. 인용 재확인: `help_order.go:31`, `help.go:44-49`, `cc.go:140`, `cg.go:93`, `glm.go:205` 전부 성립 |
| **A5** spawn 리터럴 위치 | **RESOLVED** | `research.md:128-133` 호출 지점 직접 인용 |
| **A6** superseded 계약 | **RESOLVED** | `research.md:403-417` status 열. 7건 status를 이번에 다시 읽었고 전부 일치하며, `superseded_by`도 확인 |
| **A7** REQ 본문 행 범위 | **RESOLVED** | `sed -n '145,296p' spec.md \| /usr/bin/grep -nE '\.go:[0-9]\|:[0-9]+-[0-9]+'` exit=1 |
| **A8** REQ-014 의무 압축 | **RESOLVED** | `acceptance.md:202-213`이 세 의무를 독립 판정으로 분해(REQ 분할은 예산상 불가 — 수용) |
| **A9** `s` 미측정 근거 귀속 | **RESOLVED** | 설계 보고서 `moai-proxy-three-provider-redesign-20260910.md:60`의 "확인되지 않았다"는 native picker disabled 행에 대한 문장이고, `:73`은 `Enter`/`s`를 공식 문서 인용으로만 서술한다. `plan.md:242-244`의 정정과 일치 |

---

## 저자가 스스로 밝힌 편차 8건 — 판정

1. **B3을 조건부로 서술한 것 — 수용(저자가 옳다).** 오케스트레이터 판독과 무관하게 직접 확인했다.
   `internal/hook/session_end.go:95-104`는 projectDir만 있으면 **조건 없이**
   `cleanupGLMSettingsLocal`을 부르고, `:682` 함수는 `.claude/settings.local.json` **파일**을 읽으며,
   `:726-731`에서 파일 `env` 블록에 키가 **있는지**만 본다. iter1의 "세 launcher가 모두 설정하므로
   일관되게 참"은 프로세스 env와 파일 env를 섞은 서술이었다. 조건부 문구는 쓰기 부작용
   (`design.md:239-242`)과 비-GLM teardown 금지 불변식(`spec.md:260-263`)을 온전히 운반한다.
   `internal/hook/session_start.go:809`의 `ensureGLMCredentials`가 `:884`에서
   `ANTHROPIC_BASE_URL = DefaultGLMBaseURL`을 다시 써넣는 것도 확인했다. **다만 이 수정이 G2-B1과
   G2-B2를 새로 만들었다.**
2. **REQ-MG-016을 AC-MG-021 헤더에 넣은 것 — 수용.** (c)가 게이트가 닫힌 쪽 동작("측정 전 활성화
   금지")을 실제로 판정한다. 무관한 두 REQ를 한 AC에 묶은 응집 문제와 게이트 표현이 정의되지 않은
   문제는 G2-A5로 따로 적었다.
3. **두 가지 범위 추가 — 둘 다 수용하되, 하나는 REQ에 닻을 내려야 한다.** `internal/web/widgets.templ:69-82`는
   `glm`이 아닌 모든 값을 `flat rate`로 그린다(직접 확인). kanban `Backend` 값을 렌더하는 정당한
   소비자이지, 흡수된 범위 밖 작업이 아니다. 그러나 `spec.md`에는 근거 문장이 없다
   (`/usr/bin/grep -n '배지\|badge\|flat\|정액' spec.md` exit=1). AC만 단언하는 상태다(G2-A1).
   `internal/cli/model.go:88-96`의 `rpt.Backend`는 LLM 설정에서 유도한 별개 개념으로, 건드리지
   않는다고 한 판단이 옳다.
4. **Windows** — 울타리 절 참조(G2-A4).
5. **M0 T09 spike 선행 — 정당하고 기록되었다.** 근거가 `plan.md:29-30`(출시 여부를 가르는 측정),
   `:73-74`(핸드오프 §7과의 차이를 명시), `:224-227`에 남아 있다. 핸드오프 §7 1번(T03 mock)은 첫
   **구축** 마일스톤으로 그대로 남는다. 실행 가능성 보완은 G2-A6.
6. **HISTORY 0.1.0 옛 이름 유지 — 수용.** 한 가지 불일치가 있다. 저자는 역사 보존을 이유로 이름을
   바꾸지 않았지만, 같은 0.1.0 항목의 D3 본문(`spec.md:40-46`)에는 0.2.0 내용("이 SPEC 문서와 iter1
   감사 보고서가 그 토큰을 담게 된 뒤로는 … 0.2.0의 재측정은 …")을 제자리에 덧붙였다. 사실을
   왜곡하지는 않았으므로 선택 항목으로만 남긴다(G2-A13).
7. **C1 인계 — 명시가 부족하다.** `research.md:431-435`는 "착수 전에 조율한다"고만 적는다. 누가
   언제 하는지가 없고, `plan.md` §C Pre-flight(`:36-40`)에도 항목이 없다. 게다가 이번 감사에서
   이관 대상 두 술어가 **프로덕션에서 도달 불가**하고 다른 두 SPEC이 고정한 동작을 담고 있음을
   확인했다. 이 편차는 G2-B3의 일부로 올린다.
8. **`.moai/reports/SPEC-MOAI-PROXY-001/` 미추적 — 결함 아님.** `git status --short` →
   `?? .moai/reports/SPEC-MOAI-PROXY-001/`, `git check-ignore -q` exit=1(무시 대상 아님). 들어 있는
   파일은 `plan-audit-iter1.md` 하나(mtime 15:02)로, iter1 감사가 만든 증거다.

## 새 내용 감사

### (a) `BackendGPT` / `REQ-MG-026` / `REQ-MP-007` 통합

- **`REQ-MP-007` 의무 보존 — REQ 문구에서는 손실 없음.** 복원한 0.1.0 원문: "supervisor는
  `MOAI_SESSION_PID` 각인, 종료 코드 전파, signal 전달(Ctrl-C·Ctrl-Z·fg), PTY 동작, job control을
  현재 launcher와 동일하게 보존해야 한다." `REQ-MG-005`(`spec.md:169-173`)에 다섯 항목이 모두 있다.
- **그러나 적용 범위가 조용히 좁아졌다(G2-B4).** 옛 문장은 플랫폼 무관이었다. 새 문장은 그 보장을
  "POSIX 경로의 `syscall.Exec` 호출과 그 호출에 올라타 있는 보장"으로 묶는다. Windows는 오늘
  `launch_exec_windows.go:57-60`에서 종료 코드를 전파하지만, 이제 그 의무를 요구하는 REQ가 없다
  (`REQ-MG-009`는 수명 결속만 말한다). `AC-MG-006`은 여전히 Windows 종료 코드 일치를 단언한다.
  `MOAI_SESSION_PID`는 오늘도 POSIX에서만 각인되므로(`withSessionPID` 비테스트 호출자는
  `launch_exec_posix.go:26` 하나) 그 항목의 손실은 아니다.
- **번호 폐기·재사용 금지 — 지켜졌다.** `REQ-MG-007` 토큰은 spec 어디에도 없다(`REQ tokens anywhere
  in spec`에 부재). 결번 자체는 MP-1 참조.
- **HISTORY 대응표 정확성 — 부분 부정확(G2-A3).** 0.1.0 대비 본문 diff(`proxy→gateway`,
  `MP→MG` 정규화 후) 결과는 REQ 변경 9건, AC 변경 10건이다.
  - 표에 없는 **실질** 변경: `REQ-MG-005`(REQ-MP-007 흡수), `AC-MG-003`(`s` 게이트 Gap 추가),
    `AC-MG-019`(`MOAI_LAUNCH_PROVIDER`·"`go build`로는 잡히지 않는다" 추가).
  - `REQ-MG-003`은 500자 출력 안에서 차이가 보이지 않았다(공백 차이로 추정, 미확인).
  - `AC-MG-002`·`AC-MG-010`·`AC-MG-013`은 헤더 표기 변경만 있다(`019/020` → `019, 020`).
  - `spec.md:92-95` 산문이 `REQ-MG-005` 흡수를 설명하므로 은폐는 아니지만, 표의 `001~006 | 유지 (001
    본문 정정)` 행은 틀렸다.
- **`AC-MG-014` — 응집이 무너지기 시작했다(G2-A2).** 한 AC가 (a)~(d) 네 판정에 더해 세 단언
  (kanban 상수 집합, 감사 상수 불변, 배지)을 운반하고, REQ 셋과 REQ 없는 행동 하나에 걸쳐 있다.
  iter1 A8과 같은 압축 냄새이며, 예산 25/25 때문에 생긴 것으로 보인다. (d)의 "`-k`와 `-f`"가
  `cc.go:163-230`의 네 분기(factory lead / factory worker / kanban lead / kanban companion) 중 어느
  것을 덮는지도 정해지지 않았다.
- **두 함정 — 둘 다 이름이 붙어 있다.** `internal/kanban/record.go:20-22`의 "no third value
  exists" 주석을 직접 확인했다. `design.md:289-303`은 `Backend`를 "launcher의 초기 provider"로
  재정의하고 그 주석을 고쳐 쓰라고 명시한다. `internal/cli/mcp_convergence.go:57-64`의 감사 backend
  상수도 직접 확인했으며, `spec.md:294-295`, `design.md:307-311`, `acceptance.md:116-117`이 `gpt`를
  감사 집합에 넣지 말라고 명시한다.

### (b) clarification 6건 결정

1. BackendGPT — 위 (a).
2. **`MOAI_LAUNCH_PROVIDER`** — 정직한 경계가 명시되어 있다. env는 exec에서 고정되고, 요청별
   provider는 gateway 내부에서만 알 수 있다(`spec.md:262-263`, `design.md:216-220`, `plan.md:219-220`).
   teardown 판정도 launch 시점 값을 쓴다(`design.md:257-267`). `envkeys.go` 등록 의무도 있다
   (`spec.md:251-252`, `plan.md:158`). 새 이름 미사용 측정은 `research.md:92-97`에 대조군 21과 함께
   있다. **그러나 판정 기준의 전제**("launch 때 GLM credential이 `settings.local.json`에
   주입되었는가", `design.md:222-223`)가 SPEC의 나머지와 맞지 않는다 → G2-B1.
3. T09 spike — 편차 5.
4. Windows — 울타리.
5. **`CredentialRef`** — `design.md:65-100`에 provider 중립 seam 계약(`Provider`/`Generation`/
   `Apply`/`Redacted`, `ErrCredentialAbsent`)과 구체 타입 소관 표가 있다. 확인.
6. picker `s` — B4 처분.

---

## 회귀 탐색

- **능력 주장.**
  - `s` 경로: 게이트됨(`spec.md:244-248`).
  - GPT 모델 접근성: 선언된 Gap(`acceptance.md:183-185`).
  - Claude 구독 passthrough: M0 게이트(`spec.md:220-224`, `plan.md:68-70`).
  - T03 연속 전환: 전제로 표시(`spec.md:121-123`).
  - 네 항목 모두 회귀 없음. **새로 생긴 미측정 주장** 하나는 `design.md:234-235`의 "항상 거짓"이다
    (G2-B3).
- **예산·추적성(독립 방법).** 정의는 줄 머리 굵은 `**REQ-MG-nnn**` / `**AC-MG-nnn**`로 셌고,
  보조로 파일 전체 토큰 집합도 셌다.
  - REQ 25 ≤ 25, AC 25 ≤ 25, 고아 REQ 0, 미정의 참조 0.
  - AC 본문에만 등장하고 헤더에는 없는 REQ 0.
  - 오케스트레이터 계수와 일치한다.
- **자기 무효화.**
  - `research.md:44-60`(`moai gg`)과 `acceptance.md:129-132`는 두 경로를 모두 제외한다.
  - **`research.md:92-97`(`MOAI_LAUNCH_PROVIDER` 미사용)은 SPEC 디렉터리만 빼고 `.moai/reports/`는
    빼지 않는다.** `plan.md:56-57` 자기 규칙 위반이다. 이 iter2 보고서가 그 토큰을 담으므로 다음
    재실행은 `0`을 재현하지 못한다 → G2-A9.
- **명칭 위생.**
  - `REQ-MP`/`AC-MP`는 대응표, 명칭 대응 절, 폐기 설명, 0.1.0 항목에만 등장한다(`spec.md:22,64-73,92,101`,
    `plan.md:12,211,250`, `research.md:23`, `acceptance.md:12`).
  - `PROXY-001`은 iter1 보고서 경로와 `research.md:58` grep 출력에만 등장한다.
  - 남은 "proxy"는 워크트리 경로, 0.1.0 브랜치명, 보고서 파일명, `ccmproxy`, `httputil.ReverseProxy`,
    `ConfigProxy`, `HTTP_PROXY`, `proxy.golang.org`, "DIRECTORY PROXY", 렌즈 이름과 인용뿐이다.
  - 전부 의도된 맥락이다. **PASS.**
- **핸드오프 §8 금지 사항** — 직접 판독했다.
  - `gpt-5.3`: 부재 가드 맥락뿐(`spec.md:228`, `plan.md:100`, `acceptance.md:177`은 거절 대상 요청).
  - `gpt-6-astra`: 기본값 아님(`spec.md:239` 기본은 `gpt-5.6-sol`).
  - `--provider mixed`: 0건.
  - `claude_glm`: 기존 코드 인용 2건(`research.md:121,379`).
  - Codex credential 쓰기 경로: `mcp_codex.go`에서 `WriteFile|os.Create|OpenFile` 0건.
  - 과거 기록 일괄 치환: `spec.md:348-351`에서 제외.
  - **PASS.**
- **Lint** — `moai spec lint SPEC-MOAI-GATEWAY-001` → `✓ No findings`, exit 0. 참고로 lint에는 REQ
  번호 결번 규칙이 없다(`internal/spec/lint*.go`에서 gap/sequential 관련 규칙 미발견). 그래서 MP-1
  결번은 lint 통과와 무관하다.
- **`research.md:353` 비밀값 탐지 — 오탐 확인.** 원문은 "dotenv `GLM_API_KEY="…"`"로,
  `internal/glmcred` 파일 **형식**을 설명하는 말줄임 자리표시자다. 실제 키 문자열이 아니다.

---

## Windows 검증 — 울타리 안 판단(점수 미반영)

현재 문구가 **정직하고 공허하지 않은가**만 판정했다.

- **공허하지 않다.** `acceptance.md:58-62`, `plan.md:131-134`, `design.md:167-169`는 모두 "판정이
  실리지 않으면 Gap"이라고 적는다. 존재하지 않는 CI 판정을 PASS로 흘리는 문장은 없다.
- **범위 표기는 정직하다.** 세 곳 모두 "`ci.yml` 안에서"로 한정한다. 한정한 범위 안의 사실은
  맞다 — `ci.yml:123` 주 test 매트릭스는 `[ubuntu-latest]`, `:376` `test-integration`만 3-OS,
  `:400`은 `./test/integration/harness/...`만 실행한다.
- **결론은 근거보다 강하다(G2-A4).** `.github/workflows/release-pr-multi-os.yml`은 `release/*` →
  `main` PR에서 `:91` `[ubuntu-latest, macos-latest, windows-latest]`로 `:203`
  `go test -json -race -timeout 25m ./...`를 **전 패키지 실행**한다. `ci.yml:108-110` 주석도
  "Cross-platform runtime coverage (macOS + Windows race …) moved to release time via
  release-pr-multi-os.yml"이라고 적는다. 즉 supervisor 단위 시험을 어느 패키지에 두든 **릴리스 PR
  시점에는** Windows 실행 판정이 생긴다. 이 워크플로는 `research.md:370-371`, `:489`에 "읽지 않았다"
  Gap으로 남아 있고, 그런데도 AC·plan·design은 "그 경로에 두지 않으면 판정 자체가 생기지 않는다"고
  단정한다.
- **운영자 재결정에 대한 함의.** 오케스트레이터가 운영자에게 제시한 실측 목록(`:123`, `:112-113`,
  `:376`, `:400`, `:471`)에도 이 릴리스 워크플로가 빠져 있다. 판정이 카드 병합 시점이 아니라 **릴리스
  시점**에 온다는 차이가 결정의 핵심 변수이므로, 결정 전에 전제로 보완해야 한다. 릴리스 시점
  판정이면 develop 병합 후에 발견되므로 비용 구조가 다르다.

---

## Defects Found

### Blocking

**G2-MP1** — `spec.md:175-180` (§D.2, `REQ-MG-006` 다음 `REQ-MG-008`) — `REQ-MG-007` 결번. 폐기
사실은 HISTORY(`:63-70`)에만 있고 결번 자리에는 없다. — Severity: minor — Class: blocking(must-pass)
— Required fix: 결번 자리에 묘비 한 줄을 넣는다. 예: `**REQ-MG-007** — [RETIRED] 0.2.0에서
REQ-MG-005에 통합. 번호 재사용 금지.` 묘비에는 GEARS 라벨을 달지 않고, 예산 계수에서 제외된다고
같은 줄에 적는다. 번호를 당겨 재부여하는 방식은 iter1↔iter2 대응을 깨므로 권하지 않는다.

**G2-B1** — `spec.md:115-119`, `spec.md:260-263`, `design.md:222-223`, `design.md:257-267`,
`design.md:277-279` — **gateway launch가 `settings.local.json`의 `ANTHROPIC_*` 키를 쓰는지·지우는지가
정해지지 않았고, 그 공백 위에 "비-GLM이면 훅 정리 금지"를 올렸다.** 네 사실이 서로 맞지 않는다.
- (i) 판정 기준의 전제는 "launch 때 GLM credential이 `settings.local.json`에 주입되었는가"다
  (`design.md:222-223`). 그런데 `REQ-MG-006`은 base URL을 자식 env에만 싣게 하고, `REQ-MG-023`은
  credential을 gateway 내부 `CredentialRef`로 해석하게 하며, `design.md:281-283`은 overlay가
  `settings.local.json`을 쓰지 않는다고 한다. gateway `moai glm`이 오늘의 `injectGLMEnv`
  (`internal/cli/glm.go:1007-1008`, base URL과 token을 파일에 기록)를 계속 부르는지는 어디에도
  없다.
- (ii) `spec.md:115-119`는 "mode switch의 env 주입 단계만 교체"라고 한다. `moai cc`의 `applyCCMode`가
  launch 시 `removeGLMEnv`(`internal/cli/launcher.go:221-245`)로 stale GLM 키를 지우는 동작이 그
  "주입 단계"에 포함되는지, 즉 남는지가 불명확하다. `moai gpt`에는 대응 경로 자체가 없다.
- (iii) `design.md:277-279`는 **settings 값이 gateway base URL을 되돌려 놓는다**는 전제를 스스로
  세운다. 그렇다면 비정상 종료한 이전 GLM 세션이 남긴 `ANTHROPIC_BASE_URL`/`ANTHROPIC_AUTH_TOKEN`은
  `moai cc`/`moai gpt` gateway 세션을 우회시켜 요청을 Z.AI로 곧장 보낸다. 이는 `REQ-MG-022`가 금지한
  "사용자가 선택하지 않은 유료 경로"다.
- (iv) `REQ-MG-021`의 While 절은 바로 그런 세션에서 SessionEnd 정리와 SessionStart 재주입을
  금지한다. 오늘 이 stale 키를 치우는 두 경로 중 훅 쪽이 막히고, launch 쪽은 존속이 불명확하다.
  `session_end.go:95-98` 주석이 밝히는 원래 목적("user ran 'moai glm' but ended the session without
  running 'moai cc'")이 조용히 사라질 수 있다.
- **성격**: 코드 판독과 SPEC 문장에 근거한 추론이다. settings `env`가 자식 프로세스 env보다 우선한다는
  사실은 이번에 측정하지 않았고, `design.md:277-279`의 전제를 인용했다.

Severity: critical — Class: blocking. Required fix(REQ·AC 추가 없이):
1. `REQ-MG-006` 또는 `REQ-MG-021` 본문에 한 문장을 넣는다 — "세 gateway launcher는 exec 전에
   `settings.local.json`의 `env`에서 자식 env를 덮어쓸 `ANTHROPIC_BASE_URL`·`ANTHROPIC_AUTH_TOKEN`·
   `ANTHROPIC_DEFAULT_*_MODEL`을 제거하거나, 제거할 수 없으면 명시 오류로 launch를 거부해야 하며,
   gateway launch는 이 키들을 그 파일에 기록해서는 안 된다." 제거인지 거부인지는 결정 사항이므로
   필요하면 오케스트레이터가 운영자에게 묻는다.
2. `design.md` §5.2의 판정 전제를 새 사실에 맞게 고친다. gateway 아래에서 훅의 GLM 정리는
   "gateway 이전 방식이 남긴 stale 상태 치유" 전용이 되며, 그 치유는 위 1의 launch 단계가 맡는다.
3. `AC-MG-018`에 판정 하나를 흡수한다 — "stale GLM 키가 든 `settings.local.json`으로 `moai cc`/
   `moai gpt`를 launch하면 exec 직전 파일에 세 키가 없거나 launch가 거부된다." 새 AC는 만들지 않는다.

**G2-B2** — `acceptance.md:151-155` — **`AC-MG-018`의 파일 전체 해시 동일 단언은 올바른 구현에서도
적색이다.** SessionStart 체인의 `ensureTeammateMode`(`internal/hook/session_start.go:995`)는
`teammateMode`가 원하는 값과 다르면(`:1026`) 파일을 다시 쓴다(`:1060`). `moai cc` launch의
`removeGLMEnv`는 비어 있지 않은 파일에서 매번 `delete(m, "teammateMode")`를 한다(`launcher.go:402`).
따라서 "SessionStart 직전 → SessionEnd 직후" 구간에는 GLM과 무관한 쓰기가 사실상 항상 들어간다.
파일이 없어도 `ensureTeammateMode`가 새로 만든다. 즉 비-GLM 불변식을 완벽히 지켜도 이 AC는
적색이다(잘못된 이유의 적색). iter1이 제안한 형태는 "SessionEnd 시점 비교"였고, 구간을
SessionStart까지 넓힌 개정이 이 결함을 만들었다. 성격: 코드 판독 추론이며 런타임에서 관측한 적색은
아니다. — Severity: major — Class: blocking — Required fix: 해시 대상을 **두 함수가 만지는 키 집합**
(`ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_*_MODEL`, `MOAI_BACKUP_AUTH_TOKEN`,
`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, context-window 키)의 정규화 투영으로 좁히거나, 판정을
`cleanupGLMSettingsLocal` 호출 전후와 `ensureGLMCredentials` 호출 전후로 각각 괄호 친다. 대조군
(signal 없는 세션은 정리 유지)은 그대로 둔다.

**G2-B3** — `spec.md:255-256`, `design.md:234-235`, `acceptance.md:146-149` — **`sessionEnvHasGLM`과
`hasGLMEnv`는 부분 문자열 술어가 아니라 이중 disjunct이고, 도달 불가하며, 다른 두 SPEC이 고정한
동작을 담는다.**
- (i) 두 함수 모두 `ANTHROPIC_AUTH_TOKEN`이 비어 있지 않으면 `true`를 먼저 반환한다
  (`internal/tmux/cg_detect.go:185`, `:204`). z.ai 부분 문자열 검사는 두 번째 disjunct다
  (`:188`, `:207`). `design.md:234-235`의 "항상 거짓 — GLM 세션을 놓친다"는 gateway 세션 토큰을
  어느 env로 싣느냐에 따라 뒤집힌다. 그런데 토큰 운반 키는 `design.md:129`에 "(+ 세션 토큰)"으로만
  적혀 있고 정해지지 않았다. `ANTHROPIC_AUTH_TOKEN`이면 두 술어는 모든 gateway 세션에서 **항상 참**이
  된다.
- (ii) 두 함수의 비테스트 호출자는 `IsCGMode`(`cg_detect.go:99`, `:125`) 하나다. `IsCGMode(`의
  비테스트 호출은 정의(`:91`) 말고 0건이다(`internal/template/glm_effort_overlay.go:307`은 주석).
  이관해도 프로덕션 동작은 바뀌지 않는다.
- (iii) `hasGLMEnv`의 토큰 disjunct는 `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001`(status: implemented)의
  `IsCGMode` 시험을 살리려고 **의도적으로 보존**된 것이다(`cg_detect.go:195-202` 주석,
  `SPEC-V3R6-CG-MODE-HARDENING-001`(status: completed) `spec.md:106` C-7). `internal/tmux/cg_detect_test.go`와
  `cg_detect_ssot_test.go`가 토큰을 설정한다.
- 이 SPEC은 두 SPEC을 참조하지 않으므로 D7 동사가 잡지 못한다. 편차 7(C1 인계)과 같은 뿌리다.

Severity: major — Class: blocking. Required fix(둘 중 하나):
- (권장) 두 `cg_detect.go` 술어를 `REQ-MG-021`의 이관 목록과 `AC-MG-018` 판정 지점에서 **빼고**,
  `SPEC-MOAI-CG-RETIRE-001`(제안) 소관으로 이름을 붙여 넘긴다. 이관 대상은 `hookProcessEnvHasGLM`과
  `cleanupGLMSettingsLocal` 두 지점으로 줄어든다. C1 조율 책임자를 `plan.md` §C Pre-flight에 한 줄로
  적는다.
- (대안) 남긴다면 두 disjunct를 모두 적고, gateway 세션 토큰의 운반 env 키를 `design.md` §3.2에서
  정하며, 위 두 SPEC과 `REQ-CGH-006`/C-7을 교차 참조로 명시하고 그 시험을 어떻게 유지할지 적는다.

**G2-B4** — `spec.md:169-173`, `spec.md:92-95` — **`REQ-MP-007` 통합이 종료 코드 전파 의무를 POSIX로
좁혔다.** 옛 `REQ-MP-007`은 플랫폼 무관이었다. 새 `REQ-MG-005`는 보장을 "`syscall.Exec` 호출에
올라타 있는 보장"으로 한정하고, `spec.md:94-95`는 "모두 `syscall.Exec` 보존에서 파생된다"고 적는다.
Windows 종료 코드 전파는 `child.Wait()`(`launch_exec_windows.go:57-60`)에서 나오므로 이 설명은
Windows에 대해 거짓이다. `REQ-MG-009`에는 종료 코드 문장이 없는데, `AC-MG-006`은 두 플랫폼 모두에서
종료 코드 `N` 일치를 단언한다. — Severity: minor — Class: blocking — Required fix: `REQ-MG-005`에
"종료 코드 전파는 플랫폼과 무관하게 보존한다(Windows는 `REQ-MG-009`의 spawn-and-wait 경로에서)"
한 구절을 넣고, `spec.md:94-95`의 파생 설명을 POSIX 한정으로 고친다. `REQ-MG-009` 자체는 울타리
대상이므로 건드리지 않아도 된다.

### Advisory

**G2-A1** — `acceptance.md:116-118` — `AC-MG-014`의 배지 단언("`gpt`를 근거 없이 flat rate로 표시하지
않는다")에 REQ 근거가 없다. — Severity: minor — Class: optional에 가까운 blocking 후보(추적성 역방향)
— Fix: `REQ-MG-026`에 "웹 콘솔은 `gpt` 기록의 과금 방식을 단정 표시해서는 안 된다" 한 구절을 넣거나,
단언을 `design.md` §7.3으로만 둔다.

**G2-A2** — `acceptance.md:105-121` — `AC-MG-014`가 판정 7개(REQ 3개 + REQ 없는 행동 1개)를 운반한다.
(d)가 네 진입 분기 중 무엇을 덮는지도 불명확하다. — Severity: minor — Class: optional — Fix: (d)에
"factory lead·worker, kanban lead·companion 네 분기 각각"을 명시한다.

**G2-A3** — `spec.md:67-74` — HISTORY 대응표가 실제 diff와 다르다: `REQ-MG-005`, `AC-MG-003`,
`AC-MG-019`의 실질 변경이 표에 없다. — Severity: minor — Class: optional — Fix: 표의 "본문 정정"
목록을 위 세 건으로 보완하고, 헤더 표기만 바뀐 `AC-MG-002/010/013`은 "표기만"으로 적는다.

**G2-A4** — `acceptance.md:58-62`, `plan.md:131-134`, `design.md:167-169` — (울타리, 점수 미반영)
`release-pr-multi-os.yml:91,203`의 release-time Windows 전 패키지 실행을 빠뜨린 채 "판정이 존재하지
않는다"고 단정한다. — Severity: major(결정 입력으로서) — Class: optional(울타리) — Fix: 세 곳과
`research.md` §10에 해당 워크플로를 판독해 넣고, "카드 병합 시점에는 판정 없음 / 릴리스 PR 시점에는
전 패키지 Windows 실행"으로 두 시점을 구분해 적는다. 운영자 재결정 전에 전제로 전달한다.

**G2-A5** — `acceptance.md:175-185` — `AC-MG-021`이 무관한 두 REQ(GPT ID 정확성, passthrough 게이트)를
묶는다. (c)의 "T09 spike 결과가 기록되지 않았거나 음성인 상태"가 코드에서 어떻게 표현되는지(설정 키?
빌드 플래그? 코드 부재?)는 `design.md`에 정의되지 않았다. — Severity: minor — Class: optional — Fix:
`design.md` §1 또는 §2.1에 게이트 값의 표현을 한 줄로 정한다.

**G2-A6** — `plan.md:61-74` — M0는 로컬 gateway 인증과 OAuth 전달의 공존을 재지만, 그걸 잴 loopback
forwarder는 M1/M2 이후에야 생긴다. "서로의 결과에 의존하지 않는다"(`:73-74`)는 맞지만 필요한
하네스가 적혀 있지 않다. — Class: optional — Fix: "M0는 버리는 spike용 forwarder로 측정하며 M-계열
코드를 전제하지 않는다"를 적는다.

**G2-A7** — `design.md:151-157` — 새 POSIX 전용 감시(`kill(pid, 0)`)에 `//go:build` 파일 분할 언급이
없다(CI `build` 잡의 windows 교차 컴파일이 `ci.yml:485`에서 잡기는 한다). "세션 토큰을 함께 본다"로
PID 재사용을 막는 기구도 정해지지 않았다(`homestate.ProbeProcessIdentity` 같은 기존 선례가 있다). —
Class: optional.

**G2-A8** — `acceptance.md:54-57`, `:135-138` — `REQ-MG-005`의 `MOAI_SESSION_PID` 각인 값과 PTY 동작을
판정하는 절이 `AC-MG-006`/`AC-MG-016`에 없다. 0.1.0 `AC-MP-006`/`AC-MP-016`과 같은 본문이므로 회귀는
아니다. — Class: optional — Fix: `AC-MG-016`의 소스 가드에 "`withSessionPID(env, os.Getpid())` 인수
형태" 단언을 추가한다.

**G2-A9** — `research.md:92-97` — `MOAI_LAUNCH_PROVIDER` 부재 측정이 `.moai/reports/`를 제외하지
않는다(`plan.md:56-57` 자기 규칙 위반). 이 보고서가 그 토큰을 담으므로 다음 재실행에서 `0`이
재현되지 않는다. — Class: optional(B2와 같은 계열) — Fix: `/usr/bin/grep -v '^\./\.moai/reports/'`를
추가하고 제외를 본문에 적는다.

**G2-A10** — `research.md:477` — "대조군: 전체 SPEC 디렉터리 832개". 재측정 결과
`ls .moai/specs | wc -l` → `832`(항목 수), `ls -d .moai/specs/SPEC-*/ | wc -l` → `828`(SPEC 디렉터리).
라벨이 측정 대상과 다르다. — Class: optional.

**G2-A11** — `design.md:97-100`, `acceptance.md:196-198` — 이 SPEC만 출시되면 `moai gpt`는 초기
모델 요청이 전부 거절되는 launcher가 되는데, kanban/factory 완전 지원(`REQ-MG-026`)까지 붙어 있고
출시·도움말 노출 게이팅 문장이 없다. `login`/`logout` 오류가 내부 SPEC ID를 사용자에게 보인다. —
Class: optional.

**G2-A12** — `spec.md:222-223` — `REQ-MG-016`이 "run 단계의 첫 항목인 선행 spike로 수행한다"는 절차
(HOW)를 담는다. — Class: optional.

**G2-A13** — `spec.md:40-46` — 0.1.0 항목 D3에 0.2.0 내용을 제자리에서 덧붙였다. 편차 6의 역사 보존
원칙과 결이 다르다. — Class: optional — Fix: 그 두 문장을 0.2.0 항목으로 옮긴다.

---

## Regression Check (iter1 → iter2)

- B1: RESOLVED — `orphan REQs: []`, `covered: 25`.
- B2: RESOLVED — `acceptance.md:123-133`.
- B3: **UNRESOLVED(부분)** — 존재 술어는 해소, `cg_detect.go` 두 술어 오특징화 잔존(G2-B3). 수정이
  G2-B1·G2-B2를 유발.
- B4: RESOLVED — `spec.md:241-248`, `acceptance.md:51-52`.
- B5: RESOLVED — MP-7 PASS.
- A1~A9: 전부 RESOLVED(위 처분 표).
- 정체(stagnation) 판정: B3 계열이 두 반복에 걸쳐 남았지만 **같은 문장이 그대로인 것은 아니다**
  (기구 분류가 진전됐다). "진전 없음" 표시 대상 아님.

## Recommendation

**FAIL — 저자에게 반려. 개정 범위는 여전히 좁고, REQ/AC 예산을 늘리지 않고 고칠 수 있다.**

우선순위 순:

1. **G2-B1** — gateway launch의 `settings.local.json` 계약(제거/거부, 기록 금지)을 `REQ-MG-006` 또는
   `REQ-MG-021` 본문에 한 문장으로 넣고, `design.md` §5.2 전제를 고치고, `AC-MG-018`에 stale 키 판정을
   흡수한다. 제거와 거부 중 무엇을 택할지는 운영자 결정일 수 있다.
2. **G2-B3** — 두 `cg_detect.go` 술어를 이관 목록에서 빼 형제 SPEC으로 넘긴다(권장). 이러면 C1 인계와
   `AC-MG-018` 범위가 같이 정리된다.
3. **G2-B2** — `AC-MG-018` 해시를 GLM 키 집합 투영 또는 함수별 괄호로 좁힌다. 1·2와 같은 AC이므로
   한 번에 고친다.
4. **G2-B4** — `REQ-MG-005`에 종료 코드 플랫폼 무관 구절을 넣는다.
5. **G2-MP1** — `REQ-MG-007` 묘비 한 줄.
6. 선택: G2-A1·A3·A9를 함께 처리하면 비용이 거의 없다.

**오케스트레이터 몫(저자 작업 아님)**: G2-A4 — 운영자에게 준 Windows 전제에
`release-pr-multi-os.yml`(릴리스 PR 시점 3-OS 전 패키지 실행)을 보완한 뒤 재결정을 받는다.

**iter3 범위**: 위 차단 5건 delta와 G2-A4 반영 여부만 본다(전면 재감사 아님). Tier L 상한 3회 중
마지막이다.

---

## Claim

`SPEC-MOAI-GATEWAY-001` 0.2.0의 plan 단계 5종 산출물은 집계 0.80(문턱 0.85)으로 FAIL이다. 근거는
다섯 가지다.

1. REQ 번호 결번이 must-pass MP-1을 위반한다.
2. gateway launch의 `settings.local.json` 쓰기·정리 계약이 정해지지 않은 채 훅 정리가 금지되어,
   stale GLM 키가 gateway를 우회할 여지가 있다(코드 판독 추론).
3. `AC-MG-018`의 전체 파일 해시 단언은 GLM과 무관한 `teammateMode` 쓰기 때문에 올바른 구현에서도
   적색이다(코드 판독 추론).
4. `cg_detect.go`의 두 술어가 오특징화되었고, 도달 불가하며, 두 SPEC이 고정한 동작과 충돌한다.
5. `REQ-MP-007` 통합이 Windows 종료 코드 전파 의무를 조용히 뺐다.

iter1의 B1·B2·B4·B5와 A1~A9는 인용 지점을 열어 해소를 확인했다.

## Evidence

```
$ git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
d060e0d13
WT-unified-gateway

$ python3 <spec.md 줄머리 **REQ-MG-nnn** / acceptance.md 줄머리 **AC-MG-nnn**(...) 파싱>
REQ defs (bold-start): 25 dups: []
missing in 1..26: [7]
AC defs: 25 dups: [] True
covered by header parens: 25
orphan REQs: []
AC refs to undefined REQ: []
REQ tokens anywhere in acceptance: 25 not-in-header: []

$ /usr/bin/grep -c 'NEEDS CLARIFICATION' plan.md research.md
plan.md:0
research.md:0
$ /usr/bin/grep -c 'REQ-MG' plan.md research.md          # 대조군
plan.md:13
research.md:2

$ moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid    (lint exit=0)

$ sed -n '726,731p' internal/hook/session_end.go    (발췌)
	if _, glmActive := env[config.EnvAnthropicBaseURL]; !glmActive {
		// Not in GLM mode — nothing to clean.
		return
	}

$ /usr/bin/grep -n 'EnvAnthropicAuthToken' internal/tmux/cg_detect.go
185:	if env[config.EnvAnthropicAuthToken] != "" {
204:	if os.Getenv(config.EnvAnthropicAuthToken) != "" {

$ /usr/bin/grep -rn 'IsCGMode(' --include='*.go' . | /usr/bin/grep -v _test.go
./internal/tmux/cg_detect.go:91:func IsCGMode(settingsPath string, stderrSink io.Writer) (bool, error) {

$ /usr/bin/grep -n '^status:' .moai/specs/SPEC-V3R6-CG-MODE-HARDENING-001/spec.md .moai/specs/SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/spec.md
.../SPEC-V3R6-CG-MODE-HARDENING-001/spec.md:5:status: completed
.../SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001/spec.md:5:status: implemented

$ /usr/bin/grep -n 'delete(m, "teammateMode")' internal/cli/launcher.go
402:		delete(m, "teammateMode")
$ /usr/bin/grep -n 'func ensureTeammateMode\|if current == desired\|writeSettingsSecure(settingsPath' internal/hook/session_start.go
995:func ensureTeammateMode(projectDir string) string {
1026:	if current == desired {
1060:	if err := writeSettingsSecure(settingsPath, newData); err != nil {

$ /usr/bin/grep -rn 'withSessionPID(' internal/cli --include='*.go' | /usr/bin/grep -v _test.go
internal/cli/launch_session_pid.go:32:func withSessionPID(env []string, pid int) []string {
internal/cli/launch_exec_posix.go:26:	return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))

$ /usr/bin/grep -n 'go:build' internal/cli/launch_exec_posix.go internal/cli/launch_exec_windows.go
internal/cli/launch_exec_posix.go:1://go:build !windows
internal/cli/launch_exec_windows.go:1://go:build windows

$ /usr/bin/grep -n 'go test' .github/workflows/release-pr-multi-os.yml
203:          go test -json -race -timeout 25m ./... > test-stream.json || rc=$?
$ sed -n '91p' .github/workflows/release-pr-multi-os.yml
        os: [ubuntu-latest, macos-latest, windows-latest]

$ python3 -c "<iter1 감사 transcript의 tool_result에서 0.1.0 spec/acceptance 복원 후 정규화 diff>"
old spec lines 260 old acc lines 153 REQ old/new 25 25 AC old/new 20 25
REQ changed: ['REQ-MG-001', 'REQ-MG-003', 'REQ-MG-005', 'REQ-MG-009', 'REQ-MG-016', 'REQ-MG-020', 'REQ-MG-021', 'REQ-MG-023', 'REQ-MG-024']
AC changed: ['AC-MG-002', 'AC-MG-003', 'AC-MG-005', 'AC-MG-006', 'AC-MG-010', 'AC-MG-013', 'AC-MG-014', 'AC-MG-015', 'AC-MG-018', 'AC-MG-019']

$ /usr/bin/grep -n '배지\|badge\|flat\|정액' .moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
(출력 없음, exit=1)    # 대조군: 같은 파일에서 'REQ-MG-026' 2회 이상 적중(파싱 출력 참조)

$ git status --short
?? .moai/reports/SPEC-MOAI-PROXY-001/
?? .moai/specs/SPEC-MOAI-GATEWAY-001/
$ git check-ignore -q .moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md; echo $?
1
```

## Baseline-attribution

- 모든 측정은 이번 실행에서 이 트리, HEAD `d060e0d13`, 2026-09-10에 수행했다. iter1과 오케스트레이터가
  제시한 계수(25/25, 고아 0)는 인용하지 않고 독립 파싱으로 다시 유도한 뒤 일치를 확인했다.
- 0.1.0 산출물 원문은 트리에 없다. 그래서 iter1 감사 에이전트 transcript
  (`…/subagents/agent-aplan-audit-proxy-1d168c7042898a6e.jsonl`)의 `type: user` 레코드, 즉 당시
  `Read` 도구 결과만을 원천으로 복원했다(spec 260줄, acceptance 153줄). 복원 텍스트는 iter1 보고서의
  인용(`AC-MP-006` Windows Gap 문구 등)과 대조해 일치했다.
- 설계 보고서와 핸드오프는 primary checkout `/Users/goos/MoAI/moai-adk-go/reports/`에서 읽기 전용으로
  판독했다.
- 예산 상한 25/25의 출처는 `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier다.

## Gaps — 이번 감사가 관측하지 **않은** 것

- **런타임 검증 0건.** G2-B1(settings `env`의 자식 env 우선 여부)과 G2-B2(`ensureTeammateMode` 쓰기로
  인한 해시 변화)는 코드 판독 추론이며 실제 세션에서 재현하지 않았다. G2-B1의 우선순위 전제는
  `design.md:277-279`의 문장을 인용한 것이다.
- **교차 모델 감사 미실행.** `audit_model` 키가 `.moai/config/sections/`에 없고(grep 0), SPEC
  디렉터리가 미추적이라 diff 기반 backend가 볼 입력이 없다고 판단해 `mcp__moai__audit_multi`를
  호출하지 않았다. codex/GLM 2차 의견은 없다.
- `REQ-MG-003`의 0.1.0 대비 차이가 무엇인지 — 500자 출력 안에서 보이지 않았고 문자 단위 diff를
  출력하지 않았다.
- `release-pr-multi-os.yml`의 `detect-release` 조건(`release/*` 브랜치 판정)과 `test-install.yml`
  windows 레그의 실행 내용 — 본문 발췌만 읽었다.
- `SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001` 본문 — status만 읽었다. 토큰 disjunct 보존 요구는
  `cg_detect.go` 주석과 `SPEC-V3R6-CG-MODE-HARDENING-001:106` C-7에서 가져왔다.
- `internal/hook/pre_tool.go`, `glm_tmux.go`, `settings_io.go`가 `settings.local.json`을 쓰는지 — 파일
  목록만 확인했다. `AC-MG-018` 구간에 쓰기 주체가 더 있을 수 있다.
- 설계 보고서의 §3·§11 외 절, `launcher.go` 전량, `design.md` Go 타입 스케치 컴파일 가능성 —
  iter1과 마찬가지로 보지 않았다.
- CodeRabbit·CI 결과 — 커밋이 없어 대상이 아니다.

## Residual-risk

- **G2-B1이 실제로 성립하면 가장 비싼 실패는 조용하다.** `moai gpt` 세션이 요금 방식이 다른 Z.AI로
  흘러가도, gateway 로그에는 요청이 한 건도 남지 않는다(요청이 gateway를 거치지 않으므로).
  `REQ-MG-022`의 mock 계수 AC(`AC-MG-022`)는 gateway를 경유한 요청만 세므로 이 경로를 원리상 보지
  못한다.
- **G2-B3을 권장안으로 고치면 `cg_detect.go` 쪽 위험은 형제 SPEC으로 옮겨갈 뿐 사라지지 않는다.**
  그 SPEC이 파일을 지우기 전에 이 SPEC이 먼저 출시되면, 두 술어는 gateway 아래에서 계속 틀린 답을
  낸다. 다만 도달 불가 코드이므로 사용자 영향은 없다.
- **Windows 판정이 릴리스 시점에만 생긴다는 사실이 운영자 결정에 반영되지 않으면**, supervisor의
  Windows 결함이 develop에 병합된 뒤에야 드러난다(G2-A4).
- **예산이 25/25로 꽉 차 있어**, 위 수정이 흡수 방식(기존 REQ/AC 본문 확장)에 기대야 한다. 흡수가
  반복되면 `AC-MG-014`·`AC-MG-018`처럼 한 AC가 여러 판정을 운반하는 압축이 심해진다. iter3에서 또
  흡수가 필요하면 확장보다 형제 SPEC으로 밀어내는 쪽을 먼저 검토해야 한다.
- **이 보고서도 자기 무효화 계열 토큰**(`moai gg`, `MOAI_LAUNCH_PROVIDER`)을 담는다. 이후 스윕은
  `.moai/reports/`를 제외해야 한다.
