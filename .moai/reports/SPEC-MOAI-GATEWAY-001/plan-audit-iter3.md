# SPEC Review Report: SPEC-MOAI-GATEWAY-001

Iteration: 3/3 (Tier L 상한의 마지막 회차)
Verdict: **FAIL**
Overall Score: **0.84** (조화평균, Tier L PASS 문턱 0.85, iter2 0.80 대비 **+0.04**)
STOP 신호: **발령하지 않음** — 점수가 iter2보다 올랐다(하락 조건 불충족). 다만 3회 상한에 닿았으므로
오케스트레이터는 운영자에게 PASS-with-debt / 범위 축소 / 명시 연장 세 선택지를 제시해야 한다.

감사 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, 브랜치
`WT-unified-gateway`, HEAD `d060e0d13`, SPEC 버전 `0.3.2`. 감사 시점 2026-09-10.

Reasoning context ignored per M1 Context Isolation — 요청문 §1(개정 경위)과 §3(오케스트레이터 측정값)은
결론으로 채택하지 않았다. 판정에 쓴 사실은 전부 이 트리에서 다시 쟀다. 부재 판정은 `/usr/bin/grep`으로
했고 가능한 곳에는 동일 범위 양성 대조군을 붙였다. 교차 모델 감사는 `audit_model` 미설정이라 호출하지 않았다.

---

## 총평

iter2의 차단 5건과 권고 13건은 **인용 지점을 열어 보니 모두 해소되었다.** 0.3.1·0.3.2에서 새로 들어온
내용 — 14키 정리 집합, 두 분기 억제, 결정적 픽스처(T1·T2), release PR 게이트 기반 Windows 판정 — 도
코드와 대조해 정확했다. 특히 `AC-MG-018` (c)의 파서 형식 설명은 `loadGLMKeyFromEnvFile`의 줄 단위
동작과 행 번호까지 일치한다.

그런데 FAIL이다. 이유는 둘이다.

1. **tmux 세션 env라는 두 번째 표면에 계약이 없다(G3-B1, 신규).** iter2 G2-B1이 요구한 것은
   "gateway launch가 무엇을 쓰고 지우는가"의 계약이었고, 저자는 그것을 `settings.local.json`에 대해서만
   정했다. 그런데 오늘의 `moai glm`은 tmux 안에서 **GLM token과 Z.AI base URL을 tmux 세션 env에 쓴다**
   (현재 바이너리의 살아 있는 쓰기 주체다). `moai cc`는 그것을 지운다. SessionStart 훅
   `ensureTmuxGLMEnv`도 그 표면에 쓴다. SPEC은 §A에서 이 표면을 스스로 언급하면서도(`spec.md:188-189`)
   REQ-MG-021·022의 계약과 억제 목록에서 뺐다. 구현자가 이 경로를 유지하느냐 버리느냐에 따라
   "GLM credential은 gateway child에만 전달"(`spec.md:366`)과 "기존 `moai glm` 경험 보존"(`REQ-MG-018`) 중
   하나가 깨진다.
2. **`AC-MG-018` (c)의 `cleanupGLMSettingsLocal` 억제 판정이 자연스러운 읽기에서는 공허하다(G3-B2).**
   저자가 `ensureGLMCredentials`에 대해 정확히 짚어 고친 함정 — 앞 단계 정리가 입구 조건을 지워 억제 누락을
   볼 수 없다 — 이 같은 절의 옆 함수에는 그대로 남아 있다.

두 차단 결함 모두 REQ·AC 예산을 늘리지 않고 문장 몇 개로 고칠 수 있다. G3-B1은 무엇으로 고칠지가 제품
결정이다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — 독립 파싱: 줄머리 `**REQ-MG-nnn**` 26줄, 중복 0, `missing 1..max []`,
  `retired [7]`, `live 25`. 결번 자리 묘비 `spec.md:269-271` ("[RETIRED] … 요구사항 예산 계수에서 제외한다").
  `moai spec lint` exit 0.
- **[PASS] MP-2 GEARS 형식** — 요구사항 층(`spec.md` §D의 `REQ-MG-*`)에 대해서만 판정했다. live 25개 전부
  괄호 라벨 보유: Ubiquitous 12, Event-driven 6, Unwanted 3, Where 2(`REQ-MG-009` L277, `REQ-MG-016` L314),
  복합 2(`REQ-MG-020` L335 Ubiquitous+Where, `REQ-MG-021` L344 Ubiquitous+While). 묘비 `REQ-MG-007`은 라벨이
  없고 요구사항이 아니라고 스스로 적는다(L270-271). `acceptance.md`의 Given/When/Then은 검증 층이므로 여기서
  보지 않았다.
- **[PASS] MP-3 YAML frontmatter** — `spec.md:2-14`: `id`, `title`, `version: "0.3.2"`(따옴표 semver),
  `status: draft`, `created`/`updated: 2026-09-10`, `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags`(쉼표 문자열), 추가 `tier: L`. 거부 alias 없음.
- **[N/A] MP-4 언어 중립성** — moai-adk-go 자신의 Go 내부 구현을 다루는 단일 언어 SPEC이며 배포 template
  표면을 규정하지 않는다.
- **[PASS] MP-5 D7 교차 SPEC 정합** — 6개 산출물 전체에서 SPEC ID 15개를 추출해 status를 읽었다.
  `superseded` 1건 `SPEC-MODEL-ROUTING-WIRE-001`(`superseded_by: SPEC-AGENT-ARCH-V2-001`)은
  `research.md:692`에 status와 "이미 대체되었으므로 … 빼야 한다"로 명시 조정되어 있다 → BLOCKING 아님.
  SHOULD 3건: `SPEC-MOAI-GPT-AUTH-001`·`SPEC-MOAI-CG-RETIRE-001`(둘 다 "제안"으로 표기),
  `SPEC-MOAI-PROXY-001`(iter1 보고서 경로의 옛 ID).
- **[PASS, 판단] MP-6 D8 크로스플랫폼** — 동사를 `spec.md`에 문자 그대로 돌리면 `syscall` 6회, `//go:build`·
  EXCL 0회로 BLOCKING이 찍힌다. iter1·iter2와 같은 판단으로 PASS를 유지한다. `spec.md`의 `syscall`은 전부
  **기존** `launch_exec_posix.go`(`//go:build !windows`) 보존에 관한 문장이다. 새 POSIX 전용 감시 코드에 대해서는
  이번 판에서 `design.md:184-186`이 플랫폼별 파일 분할과 빌드 태그를 명시했다(iter2 G2-A7 해소). lessons #21이
  막으려는 "새 syscall 도입 + 빌드 태그 누락"은 없다.
- **[PASS] MP-7 clarification gate** — `/usr/bin/grep -c 'NEEDS CLARIFICATION'` → 6개 파일 모두 `0`
  (`plan.md:0`, `research.md:0` 포함). 대조군: 같은 파일에서 `GLM 정리 키 집합` `plan.md:3`, `research.md:2`.

---

## Category Scores (rubric-anchored)

| 차원 | iter2 | iter3 | 증감 | 밴드 | 근거 |
|---|---|---|---|---|---|
| Clarity | 0.75 | **0.75** | 0 | 0.75 | G2-B1·B3의 해석 분기는 사라졌다(`spec.md:348-388`, `design.md` §5.3·§6). 대신 tmux 세션 env 표면에서 합리적 구현자가 서로 다른 세 방식으로 갈릴 수 있다(G3-B1). `AC-MG-018` (c)의 "같은 세션에서"와 "활성 분기가 실행되면"이 서로 다른 픽스처를 가리킨다(G3-B2) |
| Completeness | 0.90 | **0.85** | −0.05 | 0.75와 1.0 사이 | 필수 절과 Out of Scope H3 5개(`spec.md:457,468,481,487,492`) 모두 존재. 감점: §A가 두 표면을 명명하는데(`:188-189`) 계약은 한 표면만 덮는다(G3-B1). `progress.md`가 0.3.2를 반영하지 않았다(G3-A3) |
| Testability | 0.70 | **0.85** | +0.15 | 0.75와 1.0 사이 | G2-B2의 잘못된 이유의 적색이 사라졌고(`acceptance.md:153-167`), iter2가 지적한 판정 입력 미정의 2건이 정해졌다(`acceptance.md:118-121`, `:249-252`). 픽스처가 결정적이다(`:169-216`). 감점: `cleanupGLMSettingsLocal` 억제 판정의 공허 가능성(G3-B2), Windows 판정의 실행 주체·기록 위치 부재와 아티팩트 7일 보존(G3-A2) |
| Traceability | 0.90 | **0.95** | +0.05 | 1.0 근접 | 독립 파싱: live REQ 25 전부 AC 헤더 괄호에 등장, 고아 0, 폐기·미정의 참조 0. 배지 단언이 `REQ-MG-026`(`spec.md:423-426`)에 닻을 내렸다. 감점: `AC-MG-014`·`018`·`021`이 여러 REQ를 운반하는 압축(예산 25/25의 결과) |

**집계(조화평균)**: 4 / (1/0.75 + 1/0.85 + 1/0.85 + 1/0.95) = 4 / 4.7389 = **0.8441 ≈ 0.84**.
산술평균 0.850. 조화평균 기준 문턱 미달이고, 산술평균으로 보더라도 미해결 차단 결함 2건이 있어 PASS를 줄 수 없다.

---

## iter2 결함 처분

| iter2 | 처분 | 근거(열어 본 지점) |
|---|---|---|
| **G2-MP1** `REQ-MG-007` 결번 | **RESOLVED** | `spec.md:269-271` 묘비. 파싱 `missing 1..max []`, `retired [7]`. lint exit 0 |
| **G2-B1** `settings.local.json` 계약 부재 | **RESOLVED (그 표면에 대해)** | 정리·기록 금지·우선순위 미측정: `spec.md:348-369`. 판정 전제 교체: `design.md:270-285`. stale 키 판정: `acceptance.md:169-178`. **인접 표면(tmux 세션 env)은 덮지 않았다 → G3-B1(신규).** 수정 문장 자체의 회귀는 없다 |
| **G2-B2** 파일 전체 해시 적색 | **RESOLVED** | 14키 투영 + `teammateMode` 명시 제외: `acceptance.md:153-167`. 전제 문장 한 곳이 부정확하다(G3-A1) — 올바른 구현을 적색으로 만들지는 않는다 |
| **G2-B3** `cg_detect.go` 술어 | **RESOLVED** | 이관 목록에서 제외, 형제 SPEC 인계: `spec.md:386-388`, `:471-473`, `plan.md:41-44`, `design.md:308-317`. 운반 키 결정 인계 입력: `design.md:150-155` |
| **G2-B4** 종료 코드 POSIX 축소 | **RESOLVED** | `spec.md:259-262` "종료 코드 전파는 플랫폼과 무관하게 보존해야 한다". 0.2.0 파생 설명 정정 `spec.md:96-99` |
| **G2-A1** 배지 REQ 근거 | **RESOLVED** | `spec.md:423-426`, `acceptance.md:123-125` |
| **G2-A2** `AC-MG-014` (d) 분기 | **RESOLVED** | `acceptance.md:118-121` 네 분기 명시, `plan.md:132-134` |
| **G2-A3** HISTORY 대응표 | **RESOLVED** | `spec.md:70-74`. iter2 diff 결과(REQ 001·003·005·009·016·020·021·023·024 / AC 002·003·005·006·010·013·014·015·018·019)와 표가 일치 |
| **G2-A4** Windows | **RESOLVED (울타리 해제, 본문 판정)** | 아래 "C7 — Windows" 절. 릴리스 PR 워크플로를 읽어 넣었고, 카드 병합 시점과 릴리스 PR 시점을 구분했다. 인용 행 전부 재확인. 남은 권고는 G3-A2 |
| **G2-A5** passthrough 게이트 표현 | **RESOLVED** | `design.md:102-113`, `acceptance.md:249-252`. HISTORY의 절 번호 표기만 틀렸다(G3-A5) |
| **G2-A6** M0 하네스 | **RESOLVED** | `plan.md:71-73` |
| **G2-A7** 빌드 태그·PID 재사용 | **RESOLVED** | `design.md:177-186`. `CurrentProcessFingerprint` `profile_lease.go:195`, `ProbeProcessIdentity` `:202`, 파일 1행 `package homestate`(태그 없음) 확인 |
| **G2-A8** `withSessionPID` 형태 | **RESOLVED** | `acceptance.md:142-146` |
| **G2-A9** `.moai/reports/` 제외 | **RESOLVED** | `research.md:92-104`. 재측정 `0`, 대조군 `20` |
| **G2-A10** 대조군 라벨 | **RESOLVED** | `research.md:757-768`. 재측정 `828`(SPEC 디렉터리), `832`(항목) |
| **G2-A11** 출시 판단·내부 ID 노출 | **RESOLVED** | `spec.md:245-246`, `design.md:520-531`, `acceptance.md:275`, `plan.md:138-140` |
| **G2-A12** REQ-016 절차 문구 | **RESOLVED** | `spec.md:314-318`에 절차 문구 없음. 같은 종류의 냄새가 `REQ-MG-009`에 새로 생겼다(G3-A6) |
| **G2-A13** 0.1.0 항목의 0.2.0 내용 | **RESOLVED** | `spec.md:40-43`(작성 전 측정만), `:52-54`(0.2.0으로 이전) |

정체(stagnation) 판정: 세 회차 내내 같은 문장으로 남은 결함은 없다. `settings.local.json`·훅 계열은 매 회차
진전했다. G3-B1은 같은 **주제**의 다른 표면이며, 수정 실패가 아니라 수정이 닿지 않은 곳이다.

---

## 0.3.1·0.3.2 신규 내용 감사

### C1 — `injectGLMEnv`를 살아 있는 경로로 단언하는 문장: 없음 (PASS)

6개 파일의 `injectGLMEnv` 언급 23곳을 전부 읽었다. 모두 "프로덕션 호출자가 없다", "옛 동작의 기록", "되살리지
않는다"는 맥락이다(`design.md:276-277,367-369,396-397`, `plan.md:188-190`, `research.md:110-150`,
`spec.md:142-146`의 0.3.1 HISTORY). 코드 재측정: 비테스트 `-w` 검색은 `glm.go:981`(주석)과 `:987`(정의) 2줄.
대조군 `removeGLMEnv`는 8줄이 잡힌다고 `research.md:116-125`가 적은 대로다. `applyGLMMode`의 의도적 생략
주석(`launcher.go:267-272`)도 확인했다.

### C3 — 14키 정리 집합의 일관성 (PASS)

- **구성원 대조.** `removeGLMEnv`(`launcher.go:402-431`)의 `env` 키는 `MOAI_BACKUP_AUTH_TOKEN`,
  `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, 네 모델 슬롯, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`,
  `API_TIMEOUT_MS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
  `MOAI_STATUSLINE_CONTEXT_SIZE`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`로 13개다. 여기에
  `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 더하면 14개다.
- **층 (b) 유도 검증.** `injectGLMEnv`(`glm.go:1003-1029`)와 `ensureGLMCredentials`(`session_start.go:882-898`)가
  쓰는 키의 합집합에서 층 (a)를 빼면 `BETAS`, `API_TIMEOUT_MS`, `AUTO_COMPACT_WINDOW`,
  `MAX_CONTEXT_TOKENS` 4개다. `design.md:356-361` 표의 "쓰는 주체" 열까지 일치한다.
- **네 곳의 서술.** `spec.md:225`(용어), `spec.md:350-360`(REQ-MG-021), `design.md:344-384`(§6.1),
  `acceptance.md:153-161`(AC-MG-018), `plan.md:184-187`·`:294-296`이 같은 14개와 같은 층 구분을 쓴다.
- **옛 서술 잔존 스윕.** `/usr/bin/grep -n '라우팅 키 집합\|13키\|열세 키\|removeGLMEnv와 같은\|같은 키 집합'`
  결과, 옛 서술은 `spec.md:110`·`:114`(0.3.0 HISTORY)뿐이다. 나머지 적중은 정확한 사실 문장이다:
  `acceptance.md:173`(14−1=13), `design.md:341`·`research.md:442`·`spec.md:154`("같은 키 집합이 아니다").
  대조군 `GLM 정리 키 집합` 적중: spec 4, design 5, plan 3, acceptance 2, research 2.
- **네 삭제 목록 표**(`research.md:444-453`)를 코드로 대조했다. `removeGLMEnv`는 위와 같다.
  `stripGLMCredsAndSetTeammateMode`(`settings.go:179-207`)에는 FABLE·`AUTO_COMPACT`·`MAX_CONTEXT`가 없다.
  `buildTmuxClearVars`(`glm.go:595-622`)에는 `AUTH_TOKEN`·`TEAMMATE_DISPLAY`가 없고, `MAX_CONTEXT`·
  `REASONING_EFFORT`·`CLAUDE_CONFIG_DIR`·`DISABLE_PROMPT_CACHING`이 있다. `cleanupGLMSettingsLocal`
  (`session_end.go:733-744`)은 FABLE이 없고, BACKUP은 조건 없이 지운다. 표의 모든 칸이 코드와 일치한다.

### C5 — 두 분기 억제 (PASS)

`spec.md:382-384`는 억제가 "그 함수가 파일에 하는 **모든** 쓰기"를 덮어야 한다고 적는다. `design.md:418-422`는
판정을 함수 입구에 둔다. 코드상 토큰 있음 분기(`session_start.go:851-869`)도 `persistSettingsEnv`로 파일을
쓴다(`:864`). 두 보조 함수의 호출은 `:860-861`, `:894`, `:898` 네 곳이며 모두 이 함수 안이다.
`AC-MG-018` (c)의 토큰 있음 픽스처와 무신호 대조군(`acceptance.md:191-192`, `:211-214`)이 이 분기를 실제로
밟는다.

### C6·T1·T2 — 픽스처 사슬 (PASS)

- **T1이 실제 홈으로의 조용한 폴백을 막는가 — 막는다.** `paths.MoaiHome()`은
  `os.Getenv(EnvHome) != "" && filepath.IsAbs(v)`일 때만 `MOAI_HOME`을 따른다(`paths.go:69`,
  `EnvHome = "MOAI_HOME"` `:30`). `GlmEnvFile()`(`:95`)은 `joinUnderMoaiHome`(`:108-114`)을 거쳐 같은 root를
  쓴다. 훅 `loadGLMKeyFromEnvFile`은 `paths.GlmEnvFile()`을 직접 연다(`session_start.go:1345`). 따라서 훅
  실행 전에 `GlmEnvFile()`이 격리 디렉터리 아래로 풀리는지 단언하면, 상대 경로 폴백은 그 자리에서 적색이 된다.
- **T2 줄 형식이 파서와 맞는가 — 맞는다.** `:1357` `TrimSpace`, `:1358` 빈 줄·`#` 건너뜀, `:1361` 첫 `=` 분할,
  `:1365` 키 `TrimSpace`, `:1366-1367` 값 `TrimSpace` 후 `strings.Trim(val, "\"'")`, `:1369-1370` 정확 비교
  후 비어 있지 않은 값 반환. `acceptance.md:200-206`의 서술과 행 번호가 모두 일치한다. `export` 접두 줄은
  키가 `export GLM_API_KEY`가 되어 버려진다는 설명도 옳다.
- **무신호 대조군의 노출력이 T1을 조건으로 한다는 서술 — 옳다.** `acceptance.md:206-209`. 형식이 틀린 줄이면
  `:872` `apiKey == ""`로 쓰기 없이 반환한다(`:873-878`). 그러면 대조군의 `ANTHROPIC_AUTH_TOKEN`이 생기지 않아
  적색이 된다. 이 노출은 실제 홈 `.env.glm`으로 샐 수 없을 때만 성립하고, 그것을 T1이 보장한다.
- **C6 두 변형 — `removeGLMEnv` 복원 의미와 맞다.** `launcher.go:406-411`: 백업이 비어 있지 않은 문자열이면
  `AUTH_TOKEN = backup` 후 백업 삭제, 아니면 `AUTH_TOKEN` 삭제. 따라서 백업 있음 변형은 "AUTH_TOKEN = 표지
  값, 나머지 13키 부재"(`acceptance.md:172-174`), 백업 없음 변형은 "14키 전부 부재"(`:174-175`)가 정답이다.
  경계 사례 하나는 계약에 없다: 빈 문자열 백업 키는 `removeGLMEnv`가 지우지 않는다(G3-A8).
- **CG 판정 거짓 격리 디렉터리** — 훅 로컬 `isCGMode`는 `projectDir/.moai/config/sections/llm.yaml`의
  `team_mode: cg` 부분 문자열을 본다(`session_start.go:976-985`). 격리 디렉터리라면 파일이 없어 거짓이다.
  서술이 정확하다.

### C7 — Windows, 본문 판정 (PASS, 권고 1건)

인용 행 재확인(`release-pr-multi-os.yml`): 트리거 `:14` `branches: [main]`, `:35` `startsWith(github.head_ref,
'release/')`, 매트릭스 `:91`, "Every OS leg blocks the release gate" `:93`, 실행 `:203`(`-tags=integration`
없음), 업로드 단계 `:213`, 아티팩트 이름 `:217`, `retention-days: 7` `:219`, `if-no-files-found: warn` `:220`,
상단 주석 "no artifact upload" `:23`. `ci.yml`: 정책 주석 `:108`, 주 매트릭스 `:123` `[ubuntu-latest]`, 3-OS
`:376`, harness 전용 실행 `:400`.

| 점검 항목 | 결과 | 위치 |
|---|---|---|
| supervisor 시험이 `integration` 태그 없는 평범한 시험 | 성립 | `acceptance.md:59-60`, `plan.md:149-152`, `design.md:198-202` |
| 판정이 이름을 정한 시험의 `"Action":"pass"` | 성립 | `acceptance.md:61-63`, `design.md:211-213`, `plan.md:270-271` |
| 아티팩트·시험 부재가 PASS가 아님(warn 명시) | 성립 | `acceptance.md:63-65`, `design.md:214-215` |
| `skip`도 PASS가 아님 | 성립 | `acceptance.md:64` "(`skip` 포함)" |
| 카드·develop CI 미실행을 잔여 위험으로 기록 | 성립 | `acceptance.md:67-68`, `spec.md:441-442`, `plan.md:153-154`, `design.md:207-210` |
| TTY·job control을 Gap으로 | 성립 | `acceptance.md:66-67`, `design.md:216-217`, `research.md:791-792` |

남은 공백은 G3-A2로 적었다. 이 판정을 **누가, 언제, 어디에 기록하는지**가 없다. 그리고 아티팩트는 7일 뒤
사라진다(`:219`).

---

## 저자가 밝힌 편차 8건 — 판정

1. **C6 단언을 두 변형으로 나눈 것 — 수용.** `launcher.go:406-411`의 복원 의미에서 백업이 있으면
   `ANTHROPIC_AUTH_TOKEN`이 남는 것이 정답이다. 단일 단언("14키 전부 부재")은 올바른 구현에서 적색이었을 것이다.
2. **REQ-MG-022 문구 확장 — 수용(선택 권고 1).** "이전 방식의 세션이나 옛 바이너리가"는 C1·C2와 맞다. 현재
   바이너리의 `ensureGLMCredentials`도 모델 슬롯이 남은 파일에 base URL을 다시 써넣으므로(`session_start.go:884`)
   "이전 방식의 세션"이라는 표현은 필요하다. 다만 `REQ-MG-021`(`spec.md:362`)은 출처로 "사람이 고쳐"를 함께 드는데
   `REQ-MG-022`에는 없다. 의미를 바꾸지는 않는 비대칭이다.
3. **용어 "GLM 라우팅 키 집합" → "GLM 정리 키 집합" — 수용.** 층 (b)·비회귀 키는 라우팅 키가 아니므로 새 이름이
   정확하다. 스윕상 옛 이름은 0.3.0 HISTORY(`spec.md:110,114`)에만 남는다.
4. **`research.md:611`의 IsAbs 조건 누락 — 선택 권고(G3-A4).** 이 줄은 §9 `glmcred` 요약이다. 조건 전체는
   `research.md:430-431`(§6.5)에 있고, AC가 인용하는 곳도 §6.5다. §9만 읽은 사람은 상대 경로
   `MOAI_HOME`도 통한다고 오독할 수 있다.
5. **`research.md:12`의 개정 목록에서 0.3.2 누락 — 선택 권고(G3-A3).** `research.md`의 mtime은 0.3.2 산출물과
   같은 18:20이다. IsAbs 판독은 0.3.2 내용인데 그 개정이 귀속 목록에 없다. **밝히지 않은 같은 누락이
   `progress.md:6-8`에도 있다.**
6. **0.3.0 HISTORY의 옛 용어 보존 — 수용.** 0.1.0 항목의 선례와 같고, 0.3.1 C3(`spec.md:150-153`)이 대체
   사실을 기록한다.
7. **`plan.md:265`의 옛 Windows 문구 — 수용.** 문장 자체가 "이 결정으로 대체되었다"는 결정 기록이다.
8. **`AC-MG-018`이 두 REQ에 걸쳐 (a)(b)(c)를 운반 — 수용하되 결함 하나가 그 안에 숨었다.** 세 판정은 각각
   이진 판정이 가능하다. 그러나 `AC-MG-021`(`acceptance.md:254-255`)과 달리 하위 판정이 서로 독립적으로
   PASS/Gap을 받는다는 문장이 없다. 그리고 70줄로 압축된 이 절 안에서 G3-B2가 보이지 않게 되었다.
   iter2 G2-A2와 같은 냄새다(선택 권고 G3-A10).

---

## 회귀 탐색

- **능력 주장.**
  - picker `s`는 게이트되어 있다(`spec.md:338-342`, `acceptance.md:35-37`).
  - GPT 모델 접근성은 선언된 Gap이다(`acceptance.md:257-259`).
  - Claude 구독 passthrough는 코드 부재로 게이트되어 있다(`design.md:102-113`).
  - T03은 전제로 표시되어 있다(`spec.md:201-203`).
  - 우선순위는 여섯 곳에서 일관되게 미측정으로 적었다(`spec.md:368-369`, `:439-440`, `design.md` §6.4,
    `plan.md:52-53`, `acceptance.md:223-224`, `research.md:778-779`). 회귀는 없다.
  - **새 미측정 주장**: `spec.md:367` "Claude child의 base URL은 언제나 loopback gateway다"는 tmux pane
    teammate에 대해서는 보장되지 않는다(G3-B1). `design.md:449-452`의 "gateway가 맞지 않는 토큰을 명시 오류로
    거절한다"는 운반 키가 `ANTHROPIC_AUTH_TOKEN`일 때만 성립한다(G3-A9).
- **자기 무효화 부재 주장.**
  - `research.md` §1.4는 SPEC 디렉터리와 `.moai/reports/`를 둘 다 제외한다. 재측정 `0` / 대조군 `20`.
  - §1.1 `moai gg`는 같은 제외로 `0`을 재측정했다(이번에 cc 대조군은 다시 돌리지 않았다 — Gaps).
  - §13 SPEC 디렉터리 스윕은 `828`/`832`, 적중은 자기 자신 1건.
  - **부정확한 부재 주장 1건**: `design.md:429-431`, `research.md:397-398`은 `ensureTeammateMode`가
    `teammateMode`만 쓴다고 적는다. 실제로는 `CLAUDE_CODE_TEAMMATE_DISPLAY`(14키 구성원)도 지운다
    (`session_start.go:1033-1047`) → G3-A1.
- **명칭 위생 — PASS.**
  - "proxy" 적중은 이런 맥락뿐이다: 워크트리 경로·0.1.0 브랜치명, 보고서 파일명 `moai-proxy-*-20260910.md`,
    렌즈 이름 `http-proxy-precedent`, `ccmproxy`, `httputil.ReverseProxy`/`NewSingleHostReverseProxy`,
    `ConfigProxy`, `proxy.golang.org`, "reverse-proxy"(부재 서술), 명칭 대응·명칭 변경 절, "in-binary proxy"
    인용, `find -name "*proxy*"` 측정 인용.
  - `REQ-MP`/`AC-MP`는 대응표·명칭 대응·0.1.0 항목·결정 1에만 있다
    (`spec.md:22,65-74,94,177`, `plan.md:12,246,307`, `research.md:23`, `acceptance.md:12`).
- **핸드오프 §8 금지 사항 — PASS.**
  - `gpt-5.3`은 거절·부재 가드 맥락뿐이다(`spec.md:322`, `plan.md:112`, `acceptance.md:243`).
  - `gpt-6-astra`는 기본값이 아니다(`spec.md:333` 기본은 `gpt-5.6-sol`).
  - `provider mixed`는 0건이다.
  - `claude_glm`은 기존 코드 인용뿐이다(`research.md:171,655`).
  - `mcp_codex.go`의 `WriteFile|os.Create|OpenFile|os.Remove`는 0건(exit 1)이고, 대조군 `auth.json` 참조는 5건이다.
  - 과거 기록 일괄 치환은 `spec.md:487-490`에서 제외했다.
- **예산 — PASS.** REQ 25 ≤ 25, AC 25 ≤ 25(독립 적용). 묘비는 계수에서 제외된다고 명시했다.
- **내부 모순.** 치명적 모순은 없다. 사소한 불일치: `progress.md` 0.3.2 미반영(G3-A3), HISTORY A5 절 번호(G3-A5).

---

## Defects Found

### Blocking

**G3-B1 — tmux 세션 env 표면에 launch·훅 계약이 없다 (신규)**
위치: `spec.md:188-189`, `spec.md:196`, `spec.md:365-367`, `spec.md:379-385`, `spec.md:392-396`, `design.md:321-325`

같은 트리의 코드는 이렇다.

- **`moai glm`이 tmux 세션 env에 쓴다(살아 있는 경로).** `applyGLMMode`는 tmux 안이면 `injectTmuxSessionEnv`를
  부른다(`launcher.go:278-284`). 그 변수 집합에는 `ANTHROPIC_AUTH_TOKEN: apiKey`와
  `ANTHROPIC_BASE_URL: glmConfig.BaseURL`이 있다(`glm.go:507-518`). 실패 시 경고문은 "Teammates spawned in new
  tmux panes may not have GLM credentials"다. 즉 코드 작성자는 tmux pane teammate가 launcher 프로세스 env가
  아니라 tmux 세션 env에서 GLM 라우팅을 받는다고 본다.
- **`moai cc`가 그것을 지운다.** `applyCCMode` → `clearTmuxSessionEnv`(`launcher.go:223-225`).
- **SessionStart 훅도 쓴다.** `ensureTmuxGLMEnv`(`glm_tmux.go:78-140`, 호출 `session_start.go:638`)는 TMUX가 있고,
  `teammateMode == "tmux"`이고, 파일 env에 token이 있으면 `ANTHROPIC_BASE_URL`을 포함한 `glmTmuxKeys`를 tmux
  세션 env에 주입한다.
- **SessionEnd도 조건 없이 tmux 세션 env를 지운다.** `session_end.go:93`, 목록 `:634-640`.

SPEC 쪽 상황은 이렇다.

- §A는 두 표면을 모두 명명한다(`:188-189` "`settings.local.json`과 tmux 세션 env에 남은 GLM 키"). 그러나
  `REQ-MG-021`의 정리·기록 금지·억제 계약은 `settings.local.json`만 덮는다. `REQ-MG-022`의 우회 경로 조항도
  `settings.local.json`만 든다(`:394`).
- "교체하는 것은 mode switch의 env 주입 단계"(`:196`)라는 문장은 `applyGLMMode`의 tmux 주입과
  `applyCCMode`의 tmux 정리가 남는지 사라지는지를 정하지 않는다.
- While 절의 억제 목록(`:379-385`)과 `design.md` §5.4 표에는 `ensureTmuxGLMEnv`와 SessionEnd의
  `clearTmuxSessionEnv`가 없다.

구현자는 세 방식으로 갈릴 수 있다. 어느 쪽이든 SPEC의 다른 문장과 부딪힌다.

- **(i) gateway `moai glm`이 tmux 주입을 유지한다.** GLM token과 Z.AI base URL이 tmux 세션 env로 간다.
  `spec.md:366` "GLM credential은 gateway child에만 전달"과 모순된다. teammate pane은 gateway를 우회한다.
- **(ii) tmux 주입을 버린다.** tmux pane teammate는 GLM 라우팅을 잃는다(코드 경고문의 상황). `REQ-MG-018`의
  "기존 `moai glm` 사용자 경험 보존"과 충돌한다. teammate에는 `MOAI_LAUNCH_PROVIDER`도 없으므로 teammate의
  훅은 gateway 이전 판정을 돈다.
- **(iii) 정리를 넣지 않는다.** 같은 tmux 세션에서 이전에 돈 현재 바이너리의 `moai glm`(또는 형제 SPEC이
  철거하기 전의 cg 경로)이 남긴 tmux 세션 env 때문에, gateway `moai gpt` 세션의 teammate pane이 Z.AI로 곧장
  간다. G2-B1이 `settings.local.json`에서 막은 우회 계열이 두 번째 표면에서 그대로 열린다. 게다가 이 경우의
  출처는 옛 바이너리가 아니라 **현재 바이너리**다.

성격: 코드 판독 추론이다. tmux pane teammate가 어떤 env를 받는지는 Claude Code·tmux 동작이며 이번에 측정하지
않았다. 근거는 코드 주석(`launcher.go:280-282`, `glm_tmux.go:15-16`, `settings.go:201-203`)이다.

Severity: major — Class: blocking — **Tag: (O) 운영자 결정** (대안 경로 R).
왜 O인가: 세 방식 각각이 `REQ-MG-018`·`021`·`022` 중 무엇을 양보할지를 정한다. 기술적으로 옳은 답이 하나로
정해지지 않는 제품 결정이다.

Required fix (REQ·AC 추가 없이):
1. 운영자가 tmux pane teammate의 gateway 아래 동작을 정한다. 예: "gateway launch는 tmux 세션 env에 GLM
   credential·Z.AI base URL을 쓰지 않고, loopback base URL·`MOAI_LAUNCH_PROVIDER`·세션 접근 토큰만 싣는다" 또는
   "tmux pane teammate는 이 SPEC의 지원 범위 밖".
2. `REQ-MG-021`의 계약 목록에 tmux 세션 env 한 항목을 넣는다(launch 단계 정리 대상, 기록 금지 대상).
   While 절 억제 목록에 `ensureTmuxGLMEnv`를 넣고, SessionEnd `clearTmuxSessionEnv`의 처리를 정한다.
   `REQ-MG-022` 우회 조항에 tmux 세션 env를 더한다.
3. `AC-MG-018` (a)에 "stale GLM 키가 든 tmux 세션 env 픽스처(가짜 `SessionManager`)로 비-GLM gateway launch
   → exec 직전 tmux 세션 env에 `ANTHROPIC_BASE_URL`·`ANTHROPIC_AUTH_TOKEN` 없음" 한 줄을 흡수한다.
   재사용할 seam: `injectTmuxSessionEnvVia`(`glm.go:546`)의 recording-fake 선례.
4. **(R) 대안**: 1에서 "지원 범위 밖"을 택하면 §G에 `### Out of Scope — tmux pane teammate 라우팅` H3와
   후속 카드 이름을 두고, `design.md` §6.5에 잔여 위험(현재 바이너리가 남긴 tmux env로 teammate가 우회)을
   기록한다. 이 경우에도 3의 정리 판정은 남기는 것을 권한다.

**G3-B2 — `AC-MG-018` (c)의 `cleanupGLMSettingsLocal` 억제 판정은 자연스러운 읽기에서 공허하다**
위치: `acceptance.md:184-186`, `acceptance.md:215-216`

(c)는 "Given 같은 세션에서"로 시작한다. 그 세션은 (b)의 세션, 즉 launch 정리가 `ANTHROPIC_BASE_URL`을 이미
지운 파일이다. `cleanupGLMSettingsLocal`은 그 키가 없으면 억제 여부와 무관하게 `session_end.go:728-731`에서
돌아간다. 따라서 억제를 빠뜨린 구현도 이 판정은 녹색이다. 저자는 같은 절에서 `ensureGLMCredentials`에 대해
정확히 이 함정을 짚었다("(b)만으로는 억제 누락을 볼 수 없기 때문이다", `acceptance.md:187-188`). 그리고
모델 슬롯을 남긴 픽스처와 무신호 대조군으로 고쳤다. 옆 함수에는 그 처방이 없다. "활성 분기가 실행되면"을 키가
있는 픽스처로 읽는 사람도 있겠지만, 두 읽기의 시험 효력이 다르다.

이 억제는 형식적인 것이 아니다. `settings.local.json`은 같은 프로젝트의 동시 세션이 공유한다. gateway 세션
도중 다른 세션이 키를 다시 쓰면, 억제되지 않은 SessionEnd가 남의 키를 지운다. `design.md:305-306`이 "두 장치가
겹쳐 있어 어느 한쪽이 빠져도"라고 기대는 바로 그 장치다.

Severity: minor — Class: blocking(명시된 REQ-MG-021 While 절의 검증이 공허함) — **Tag: (D) 부채 적격**.
왜 D인가: 결정 사항이 없고, `plan.md:180`이 M7에서 `AC-MG-018` 시험을 구현보다 먼저 세우도록 정해 두었다.
부채 메모를 시험 작성자에게 묶으면 run 단계 첫 작업에서 닫힌다.

Required fix: (c)에 `ANTHROPIC_BASE_URL` 키(와 모델 슬롯·백업)가 있는 픽스처로 `cleanupGLMSettingsLocal`을
**직접** 호출한다고 적는다. 대조군을 한 줄 더한다: `MOAI_LAUNCH_PROVIDER`가 없으면 같은 픽스처에서 키가 지워진다.

### Advisory (M6: optional — 판정을 단독으로 FAIL로 만들지 않음)

- **G3-A1** — `design.md:429-431`, `research.md:397-398`, `acceptance.md:219-222`
  - 문제: `ensureTeammateMode`가 `teammateMode`만 쓴다는 서술은 틀렸다. `session_start.go:1033-1047`은
    `teammateMode`를 다시 쓸 때 env의 `CLAUDE_CODE_TEAMMATE_DISPLAY`(14키 구성원)도 지운다.
  - 영향: 그 키는 launch 정리가 이미 지웠으므로 (b)가 올바른 구현에서 적색이 되지는 않는다. 결론은 살아남고
    전제 문장만 틀렸다.
  - research §6.3 검색이 대입(`= `) 패턴만 봐서 삭제를 놓쳤다.
  - 수정: 세 문장에 "legacy `CLAUDE_CODE_TEAMMATE_DISPLAY` 삭제"를 더한다.
- **G3-A2** — `acceptance.md:58-68`, `design.md:196-217`, `plan.md:264-276`
  - Windows 판정의 **실행 주체·시점·기록 위치**가 없다. 이 판정은 카드 run/sync가 닫히고 develop에 병합된
    **뒤** release PR에서만 생긴다.
  - 그러나 카드 DoD(`acceptance.md:304-305`)는 AC-MG-006 Windows 절반을 카드 종료 시점에 무엇으로 표시하는지
    정하지 않는다. 미리 선언한 Gap 목록에는 TTY·job control만 있다.
  - 아티팩트는 `retention-days: 7`(`release-pr-multi-os.yml:219`)이다. 7일 안에 아무도 읽지 않으면 "부재는 PASS
    아님"에 따라 판정이 영구 Gap이 된다. 재실행은 `workflow_dispatch`로 가능하다.
  - 수정: "카드 종료 시 Windows 절반은 Gap(판정 대기)으로 기록하고, release PR에서 리드가 아티팩트를 읽어
    `<경로>`에 기록한다. 보존 7일" 한 문단을 둔다.
- **G3-A3** — `progress.md:6-8`, `research.md:12`
  - 0.3.2 개정이 기록되지 않았다. `research.md` §6.5의 IsAbs 판독이 귀속 목록에 없는 개정에서 왔다.
- **G3-A4** — `research.md:611`
  - `MOAI_HOME` 우회에 IsAbs 조건이 없다(편차 4).
- **G3-A5** — `spec.md:123`
  - 0.3.0 HISTORY가 A5 반영 위치를 `design.md` §2.1로 적는다. 실제는 §2.2다(`design.md:102`,
    `acceptance.md:249`).
- **G3-A6** — `spec.md:280-281`
  - `REQ-MG-009`가 검증 장소("release PR 게이트의 Windows 레그에서 판정")를 요구사항 층에 담는다.
    G2-A12에서 `REQ-MG-016`의 절차 문구를 뺀 것과 결이 다르다.
  - 수정: `plan.md`·`acceptance.md`로 옮기거나, 운영자 결정의 제약으로 §E에만 둔다.
- **G3-A7** — `design.md:412`, `research.md:421-423`
  - "두 값이 바뀐 경우에만 다시 쓴다"는 둘 다 바뀐 경우로 읽힌다. 코드는 연결 문자열 비교
    `after != before`(`session_start.go:858-866`)로, 하나만 바뀌어도 쓴다. "어느 하나라도"로 고친다.
- **G3-A8** — `spec.md:353-354`, `design.md:348`
  - 빈 문자열 `MOAI_BACKUP_AUTH_TOKEN`의 처리가 계약에 없다. `removeGLMEnv`는 비어 있지 않을 때만 백업 키를
    지운다(`launcher.go:406-411`). `removeGLMEnv` + `MAX_CONTEXT` 삭제로 구현하면 빈 백업 키가 남아 "14키
    부재" 취지와 어긋날 수 있다. `AC-MG-018` (a) 두 변형 모두 이 경우를 덮지 않는다.
- **G3-A9** — `design.md:448-452`
  - 복원된 백업 토큰이 "드러나는 실패"가 된다는 잔여 위험 서술은 운반 키가 `ANTHROPIC_AUTH_TOKEN`일 때만
    성립한다. 운반 키가 별도 헤더면 복원된 사용자 토큰이 gateway로 조용히 전송되고 무시된다. 조건을 명시한다.
- **G3-A10** — `acceptance.md:153-224`
  - `AC-MG-018` 하위 판정의 독립성 문장이 없다(편차 8). `AC-MG-021:254-255`와 같은 한 줄을 둔다.

---

## Regression Check (iter2 → iter3)

- G2-MP1: RESOLVED — 파싱 `missing 1..max []`, `retired [7]`.
- G2-B1: RESOLVED(`settings.local.json` 표면) — `spec.md:348-369`. 인접 표면은 G3-B1(신규).
- G2-B2: RESOLVED — `acceptance.md:153-167`.
- G2-B3: RESOLVED — `spec.md:386-388`, `plan.md:41-44`.
- G2-B4: RESOLVED — `spec.md:259-262`.
- G2-A1~A13: 전부 RESOLVED(처분 표). G2-A4는 울타리를 걷고 본문으로 판정했다.

## Recommendation

**FAIL — 3회 상한 도달. 오케스트레이터는 운영자에게 에스컬레이션한다.**

차단 결함은 2건뿐이고 둘 다 좁다. 운영자 선택지별 권고:

1. **PASS-with-debt** — G3-B2(D)만 부채로 남긴다면 성립한다. **G3-B1(O)은 부채로 넘기면 안 된다.** 결정 없이
   run 단계에 들어가면 M7 구현자가 세 방식 중 하나를 조용히 고르게 된다. 이것이 Core Behavior #1이 금지하는
   침묵 가정이다. 운영자가 G3-B1의 방향(유지 / 폐기 / loopback만 주입 / 범위 밖)을 한 줄로 정하고, 그 결정을
   Implementation Kickoff 전제로 기록한다면 PASS-with-debt가 가능하다.
2. **범위 축소(R)** — G3-B1에서 "tmux pane teammate는 범위 밖"을 택하고 Out of Scope H3와 잔여 위험을 둔다.
   예산을 쓰지 않는다.
3. **명시 연장(iter4)** — 수정 범위는 `spec.md` REQ-MG-021·022 문장 몇 개, `design.md` §5.4·§6, `acceptance.md`
   AC-MG-018 (a)(c) 두 줄이다. iter4는 이 delta만 보면 된다.

선택 권고 중 G3-A1·A2·A7은 비용이 거의 없으니 함께 고치기를 권한다.

---

## Claim

`SPEC-MOAI-GATEWAY-001` 0.3.2의 plan 단계 산출물은 must-pass 7개를 모두 통과하고 iter2 결함 18건을 전부
해소했다. 하지만 조화평균 0.84로 Tier L 문턱 0.85에 못 미치고, 미해결 차단 결함 2건이 있어 FAIL이다.

- **G3-B1** — `moai glm`이 오늘 tmux 세션 env에 GLM token과 Z.AI base URL을 쓰는데, gateway 아래에서 그
  표면을 어떻게 다룰지 SPEC이 정하지 않았다(코드 판독 추론).
- **G3-B2** — `AC-MG-018` (c)의 `cleanupGLMSettingsLocal` 억제 판정은 앞 단계 정리 때문에 공허해질 수 있다.

## Evidence

```
$ git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
d060e0d13
WT-unified-gateway

$ python3 <spec.md 줄머리 **REQ-MG-nnn** / acceptance.md 줄머리 **AC-MG-nnn** (...) 독립 파싱>
REQ header lines: 26 distinct 26
dups []
missing 1..max []
retired [7] live 25
AC defs 25 dups [] missing []
orphan live REQs []
AC refs to undefined/retired []

$ moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
lint_exit=0

$ /usr/bin/grep -c 'NEEDS CLARIFICATION' *.md        # SPEC 디렉터리
acceptance.md:0  design.md:0  plan.md:0  progress.md:0  research.md:0  spec.md:0
$ /usr/bin/grep -c 'GLM 정리 키 집합' *.md            # 대조군
acceptance.md:2  design.md:5  plan.md:3  progress.md:0  research.md:2  spec.md:4

$ /usr/bin/grep -n '라우팅 키 집합\|13키\|13개 키\|열세 키\|removeGLMEnv와 같은\|같은 키 집합\|removeGLMEnv`와 같은' *.md
acceptance.md:173:  `env`에는 `ANTHROPIC_AUTH_TOKEN`을 뺀 열세 키가 하나도 없고, ...
design.md:341:개는 서로 같은 키 집합이 아니며(`research.md` §6.5), ...
research.md:442:**네 GLM 삭제 목록 — 같은 키 집합이 아니다.** ✓ = 지운다.
spec.md:110:    exec 전에 GLM 라우팅 키 집합을 `removeGLMEnv`와 같은 의미로 정리하고, ...
spec.md:114:  - G2-B2 — `AC-MG-018`의 파일 전체 해시 단언을 GLM 라우팅 키 집합 투영 비교로 바꾸고,
spec.md:154:  - C4 — 저장소의 네 GLM 삭제 목록은 같은 키 집합이 아니다. ...

$ sed -n '402,431p' internal/cli/launcher.go   (발췌)
		delete(m, "teammateMode")
			if backup, bok := env["MOAI_BACKUP_AUTH_TOKEN"].(string); bok && backup != "" {
				env[config.EnvAnthropicAuthToken] = backup
				delete(env, "MOAI_BACKUP_AUTH_TOKEN")
			} else {
				delete(env, config.EnvAnthropicAuthToken)
			}
			delete(env, config.EnvAnthropicBaseURL)
			... HaikuModel / SonnetModel / OpusModel / FableModel
			delete(env, config.EnvClaudeCodeDisableExperimentalBetas)
			delete(env, "API_TIMEOUT_MS")
			delete(env, config.EnvClaudeCodeDisableNonessentialTraffic)
			delete(env, config.EnvClaudeCodeTeammateDisplay)
			delete(env, "MOAI_STATUSLINE_CONTEXT_SIZE")
			delete(env, config.EnvClaudeCodeAutoCompactWindow)

$ sed -n '278,284p' internal/cli/launcher.go
	if tmux.NewDetector().InTmuxSession() {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject GLM env into tmux session: %v\n"+
				"  Teammates spawned in new tmux panes may not have GLM credentials.\n"+

$ sed -n '507,511p' internal/cli/glm.go
func buildTmuxInjectVars(glmConfig *GLMConfigFromYAML, apiKey string) map[string]string {
	vars := map[string]string{
		config.EnvAnthropicAuthToken:          apiKey,
		config.EnvAnthropicBaseURL:            glmConfig.BaseURL,

$ /usr/bin/grep -rn 'func injectTmuxSessionEnv\|func clearTmuxSessionEnv\|func ensureTmuxGLMEnv\|injectTmuxSessionEnv(\|ensureTmuxGLMEnv(' internal --include='*.go' | /usr/bin/grep -v '_test.go'
internal/cli/glm.go:490:func injectTmuxSessionEnv(glmConfig *GLMConfigFromYAML, apiKey string) error {
internal/cli/glm.go:574:func clearTmuxSessionEnv() error {
internal/cli/launcher.go:279:		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
internal/hook/session_start.go:638:	if msg := ensureTmuxGLMEnv(input.ProjectDir); msg != "" {
internal/hook/session_end.go:646:func clearTmuxSessionEnv(ctx context.Context) {
internal/hook/glm_tmux.go:78:func ensureTmuxGLMEnv(projectDir string) string {

$ for f in spec plan acceptance design; do /usr/bin/grep -n 'tmux\|teammate' $f.md; done   (요지)
spec.md:189 (§A 서술), :115-116·:386·:471 (teammateMode 제외, internal/tmux 술어 인계) — tmux 세션 env 계약 문장 없음
design.md:152·312-313·577 (internal/tmux 술어), :386-387·:430 (teammateMode) — 동일
plan.md:21·41·196 (internal/tmux 술어 인계), :229 (teammateMode) — 동일
acceptance.md:163-164 (teammateMode 제외) — 동일
대조군: /usr/bin/grep -c 'settings.local' → spec.md:13, design.md:9

$ sed -n '726,731p' internal/hook/session_end.go
	// ANTHROPIC_BASE_URL is the GLM-active indicator.
	// Claude Code's own OAuth flow never sets this variable.
	if _, glmActive := env[config.EnvAnthropicBaseURL]; !glmActive {
		// Not in GLM mode — nothing to clean.
		return
	}

$ cat -n internal/hook/session_start.go | sed -n '1033,1038p'
  1033		// Clean up legacy env var if present.
  1034		if envRaw, ok := raw["env"]; ok {
  1035			var env map[string]string
  1036			if err := json.Unmarshal(envRaw, &env); err == nil {
  1037				if _, legacy := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]; legacy {
  1038					delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")

$ /usr/bin/grep -n 'if isCGMode(projectDir)\|if token := settings.Env\|maybeSet1MAutoCompactWindow(settings.Env)\|maybeDeclareGLMContextWindow(settings.Env)\|persistSettingsEnv(settings.Env\|apiKey := loadGLMKeyFromEnvFile\|= apiKey$\|func loadGLMKeyFromEnvFile' internal/hook/session_start.go
829:	if isCGMode(projectDir) {
851:	if token := settings.Env[config.EnvAnthropicAuthToken]; token != "" {
860:		maybeSet1MAutoCompactWindow(settings.Env)
861:		maybeDeclareGLMContextWindow(settings.Env)
864:			if err := persistSettingsEnv(settings.Env, data, settingsPath); err != nil {
872:	apiKey := loadGLMKeyFromEnvFile()
882:	settings.Env[config.EnvAnthropicAuthToken] = apiKey
894:	maybeSet1MAutoCompactWindow(settings.Env)
898:	maybeDeclareGLMContextWindow(settings.Env)
900:	if err := persistSettingsEnv(settings.Env, data, settingsPath); err != nil {
1344:func loadGLMKeyFromEnvFile() string {

$ /usr/bin/grep -n 'EnvHome =\|IsAbs' internal/paths/paths.go ; /usr/bin/grep -n -A2 'func joinUnderMoaiHome' internal/paths/paths.go
30:const EnvHome = "MOAI_HOME"
69:	if v := os.Getenv(EnvHome); v != "" && filepath.IsAbs(v) {
108:func joinUnderMoaiHome(segment string) (string, error) {
109-	root, err := MoaiHome()

$ /usr/bin/grep -n 'branches: \[main\]\|startsWith(github.head_ref\|os: \[ubuntu-latest, macos-latest, windows-latest\]\|Every OS leg\|go test -json -race\|name: Upload test event stream\|test-stream-release-verify-\|retention-days\|if-no-files-found\|no artifact upload' .github/workflows/release-pr-multi-os.yml
14:    branches: [main]
23:# any GITHUB_TOKEN write surface (no artifact upload, no PR mutation).
35:    if: startsWith(github.head_ref, 'release/') || github.event_name == 'workflow_dispatch'
91:        os: [ubuntu-latest, macos-latest, windows-latest]
93:    # Every OS leg blocks the release gate. The windows leg was previously
203:          go test -json -race -timeout 25m ./... > test-stream.json || rc=$?
213:      - name: Upload test event stream
217:          name: test-stream-release-verify-${{ matrix.os }}
219:          retention-days: 7
220:          if-no-files-found: warn
$ /usr/bin/grep -n 'Cross-platform runtime coverage\|os: \[ubuntu-latest, macos-latest, windows-latest\]\|test/integration/harness/\.\.\.' .github/workflows/ci.yml
108:  # Cross-platform runtime coverage (macOS + Windows race + MX validator) moved
376:        os: [ubuntu-latest, macos-latest, windows-latest]
400:          go test -json -tags=integration -race -timeout 180s ./test/integration/harness/... > test-stream.json || rc=$?
(ci.yml:123 `os: [ubuntu-latest]`는 sed -n '104,126p'로 확인)

$ <D7: SPEC ID 추출 → 각 spec.md status>
SPEC-AGENT-ARCH-V2-001 status=completed
SPEC-FACTORY-MODE-001 status=completed
SPEC-FACTORY-WORKER-FANOUT-001 status=implemented
SPEC-INFINITE-GOAL-001 status=completed
SPEC-KANBAN-BOOTSTRAP-001 status=draft
SPEC-KANBAN-RECORD-SESSION-KEY-001 status=completed
SPEC-MOAI-CG-RETIRE-001 NOT-FOUND
SPEC-MOAI-GATEWAY-001 status=draft
SPEC-MOAI-GPT-AUTH-001 NOT-FOUND
SPEC-MOAI-PROXY-001 NOT-FOUND
SPEC-MODEL-ROUTING-WIRE-001 status=superseded superseded_by: SPEC-AGENT-ARCH-V2-001
SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001 status=completed
SPEC-V3R6-CG-MODE-HARDENING-001 status=completed
SPEC-V3R6-TOOL-POLICY-SSOT-001 status=completed
SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 status=implemented

$ /usr/bin/grep -c 'syscall' spec.md ; /usr/bin/grep -c '//go:build\|EXCL.*syscall' spec.md
6
0

$ <research §1.4 재측정: SPEC 디렉터리·.moai/reports/ 제외>
MOAI_LAUNCH_PROVIDER|EnvMoaiLaunchProvider → 0
MOAI_KANBAN_BACKEND|EnvMoaiKanbanBackend  → 20   (대조군)
"moai gg" (같은 제외)                      → 0
$ ls -d .moai/specs/SPEC-*/ | wc -l ; ls .moai/specs | wc -l
828
832

$ /usr/bin/grep -n 'WriteFile\|os.Create\|OpenFile\|os.Remove' internal/cli/mcp_codex.go ; echo "codex_exit=$?"
codex_exit=1
$ /usr/bin/grep -c 'auth.json\|codexAuthFileName' internal/cli/mcp_codex.go     # 대조군
5
```

## Baseline-attribution

- 모든 측정은 이번 실행, 이 트리, HEAD `d060e0d13`, 2026-09-10에 수행했다. 요청문 §3이 제시한 계수(26줄/25 live,
  고아 0)와 행 번호는 인용하지 않고 독립 파싱과 `grep -n`으로 다시 유도해 일치를 확인했다.
- 요청문 §3이 가리킨 행 중 `release-pr-multi-os.yml:214-220`은 이번 측정에서 업로드 단계 시작이 `:213`이었다.
  `design.md:212`의 `:213-220`이 옳다(acceptance.md는 행 번호를 인용하지 않는다).
- 점수의 기준선은 iter2 보고서(`.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md`)의 차원별 점수이며,
  그 보고서를 전량 읽었다. 문턱 0.85의 출처는 `spec-workflow.md` § SPEC Complexity Tier(Tier L)다.
- 이번 판정에 쓴 SPEC 산출물은 6종 전량을 읽었다(spec 505줄, plan 324줄, acceptance 320줄, design 579줄,
  research 800줄, progress 28줄).

## Gaps — 이번 감사가 관측하지 **않은** 것

- **런타임 검증 0건.** G3-B1(tmux pane teammate가 tmux 세션 env에서 라우팅을 받는지)은 코드 주석과 호출 구조에
  근거한 추론이다. Claude Code가 teammate pane을 띄울 때 넘기는 env는 측정하지 않았다.
- **교차 모델 감사 미실행.** `audit_model` 미설정(요청문), 단일 backend 감사다.
- `launch_exec_windows.go:57,60,64`(REQ-MG-005의 Windows 근거)를 이번에 다시 열지 않았다. iter2가 `:57-60`을
  확인한 기록에 기대었다.
- `moai gg` 재측정에서 `moai cc` 양성 대조군은 다시 돌리지 않았다. 같은 제외 필터의 대조군으로
  `MOAI_KANBAN_BACKEND` `20`만 이번에 쟀다.
- `applyCGMode`(`launcher.go:290~`)가 tmux 세션 env에 무엇을 쓰는지는 끝까지 읽지 않았다.
  `injectTmuxSessionEnv(`의 비테스트 호출이 `launcher.go:279` 하나임만 확인했다.
- `cc.go`·`glm.go`의 kanban 네 분기 행 번호(`research.md:512-522`)는 이번에 재확인하지 않았다(iter2 확인분).
- `statusline.ResolveGLMContextWindow`가 어떤 GLM 모델 ID를 아는지 — 토큰 있음 픽스처가 요구하는 "해석기가
  아는 ID"의 구체값은 확인하지 않았다.
- `design.md` Go 타입 스케치의 컴파일 가능성, 설계 원문·핸드오프 보고서 본문 — 이번 회차에서 다시 읽지 않았다.
- release PR 워크플로가 실제로 windows-latest에서 초록으로 끝난 실행 기록 — 조회하지 않았다.

## Residual-risk

- **G3-B1이 실제로 성립하면 가장 비싼 실패는 이번에도 조용하다.** tmux teammate 요청이 gateway를 거치지
  않으면 gateway 로그에 흔적이 없고, `AC-MG-022`의 mock 계수는 원리상 그 경로를 보지 못한다. 출처가 옛
  바이너리가 아니라 현재 바이너리의 `moai glm`·cg 경로이므로, 업그레이드 사용자만이 아니라 같은 tmux 세션에서
  두 명령을 번갈아 쓰는 모든 사용자가 대상이다.
- **G3-B2를 부채로 넘기면**, M7 시험 작성자가 부채 메모를 읽지 않는 순간 억제 누락이 초록으로 통과한다.
  동시 세션이 같은 `settings.local.json`을 쓸 때만 드러나므로 재현이 드물다.
- **Windows 판정은 구조적으로 카드 종료 뒤에 온다.** 판정 주체와 기록 위치가 정해지지 않으면(G3-A2), 판정이
  PASS로 오표기되지는 않지만 아무도 읽지 않은 채 Gap으로 사라질 수 있다. 아티팩트 보존은 7일이다.
- **예산이 25/25로 꽉 차 있다.** G3-B1 수정도 흡수 방식(REQ-MG-021 확장, AC-MG-018 한 줄)에 기댈 수밖에 없다.
  `AC-MG-018`은 이미 70줄이다. 이번 결함 하나(G3-B2)가 그 압축 속에 숨었다는 사실이, 다음 흡수 전에 형제 SPEC으로
  밀어낼 항목을 먼저 정해야 한다는 신호다.
- **이 보고서도 자기 무효화 계열 토큰**(`moai gg`, `MOAI_LAUNCH_PROVIDER`, `injectGLMEnv`)을 담는다. 이후 부재
  스윕은 `.moai/reports/`를 제외해야 한다.
