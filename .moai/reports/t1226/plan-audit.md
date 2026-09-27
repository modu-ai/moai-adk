auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-ALWAYS-LOADED-HEADROOM-001
Iteration: 1/2 (Tier M — `harness.yaml` `plan_audit_tier_ceilings.M: 2`)
Verdict: FAIL
Overall Score: 0.62 (Tier M PASS 문턱 0.80 — `spec-workflow.md` § SPEC Complexity Tier)

카드 t1226 · 감사 대상 커밋 `10281a857` · 브랜치 `WT-always-loaded-headroom` · 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1226`

Reasoning context ignored per M1 Context Isolation. 오케스트레이터가 넘긴 「실측 사실」 세 가지는 주장으로만 받고 아래에서 다시 쟀다.

## 감사자 재측정 (이 실행, 이 트리)

| 항목 | 명령 | 관측 |
|---|---|---|
| HEAD·브랜치 | `git rev-parse --short HEAD` · `git branch --show-current` | `10281a857` · `WT-always-loaded-headroom`, 작업 트리 깨끗함 |
| 기준 커밋 대비 변경 | `git diff --stat 7fe658815 HEAD` | SPEC 5파일 + `verdict.md` 뼈대, 612줄 추가. 18개 계수 파일 변경 0 |
| `S_live` 합계 | acceptance.md AC-ALH-002 블록 그대로 | `199111 total` — 오케스트레이터 값과 일치 |
| 템플릿 미렌더링 합계 | research.md §1 명령 그대로 | `203611 total` — research.md 와 일치 |
| 템플릿 import | `grep -n '^@' internal/template/templates/CLAUDE.md` | `9:@AGENTS.md` · `107:` · `108:` 두 yaml — research.md 와 일치 |
| develop 이동 여부 | `git rev-parse develop` | `7fe658815eb0…` — 기준 커밋과 같음(현재 시점) |
| `paths:` 없는 룰 | 두 트리의 `.claude/rules/moai/*/*.md` 머리 12줄에서 `^paths:` 검사 | 라이브·템플릿 모두 **13개**. `workflow/skill-routing.md` 는 두 트리 모두 `paths: ".claude/agents/**,.claude/skills/**,.moai/config/sections/delegation.yaml"` 를 가짐, 각 4,966자 |
| sec.py | `ls -la` · `shasum -a 256 /tmp/claude-501/sec.py` | 1,661 bytes, 2026-09-25 13:29, `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547` |
| 원 트리 | `git cat-file -t 172ef22eb` | `commit` |
| 린트 | `go build -o <scratchpad>/moai ./cmd/moai` → `<scratchpad>/moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo exit=$?` | `built` / `0 error(s), 0 warning(s)` / `exit=0`. INFO 1건 `OwnershipTransitionUnmeasured` (커밋에 `Authored-By-Agent` 트레일러 없음) |
| `moai init` 사용법 | `<scratchpad>/moai init --help` | `moai init <project-name>` = 「새 폴더를 만들고 그 안에서 초기화」, `--root` 플래그 존재, 기본은 slim(core-only) 모드, `--all` 별도 |
| `S_init` 절차 실행 가능성 | acceptance.md L91 형태(`unset … && HOME=<scratch>/init-home <scratch>/moai init <scratch>/init-proj --non-interactive`) | **워크트리 격리 가드가 거부**: 「this command sets HOME, injecting git configuration whose effect on where git writes can't be verified. Refusing to run it」 |
| init 의 홈 쓰기 | `internal/cli/init_home_guard_test.go:6-9` | runInit 이 쓰는 홈 경로 넷 — user-scope settings, profile ledger, `MOAI_HOME`, 셸 rc |

교차 모델 감사(`mcp__moai__audit_multi` 등)는 호출하지 않았다. 프로젝트 설정에 `audit_model` 키가 없고(`.moai/config/sections/*.yaml` grep 0건), 백엔드가 읽는 대상은 서버 쪽에서 원격 기본 브랜치로 해석되는 코드 diff 인데 이번 감사 대상은 develop 기반 마크다운 SPEC 이다. Claude 단독 감사다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L64-78` 에 REQ-ALH-001 ~ REQ-ALH-015 가 빈틈·중복 없이 3자리로 이어진다.
- [PASS] MP-2 GEARS 준수(요구사항 층 `spec.md §C` 대상): 15개 REQ 모두 `shall` / `shall not` 을 가지며 Ubiquitous(001·002·003·008·009·010·014·015), Where(004), Unwanted(005·012·013), When(006·007·011) 형태다. REQ-ALH-004 의 `Where` 는 파일 정체라는 정적 속성에 걸려 있어 허용 범위 안이다. AC 는 검증 층(Given-When-Then)이라 여기서 채점하지 않았다. 린트도 0 error.
- [PASS] MP-3 frontmatter: `spec.md:L2-14` — `id`·`title`·`version: "0.1.0"`(따옴표)·`status: draft`·`created`/`updated: 2026-09-27`·`author`·`priority: P1`·`phase`·`module`·`lifecycle: spec-anchored`·`tags`(쉼표 문자열) 12개 모두 있음. 거부 별칭 없음.
- [N/A] MP-4 언어 중립성: 다언어 도구 SPEC 이 아니다(도구명 0건).
- [PASS] MP-5 D7: 본문 참조 SPEC 은 `SPEC-ALWAYS-LOADED-DIET-001` · `SPEC-ALWAYS-LOADED-DIET-002` 둘, 둘 다 `status: completed`(retired/superseded/archived 아님). BLOCKING 0.
- [PASS] MP-6 D8: `grep -c syscall` 5개 파일 모두 0.
- [PASS] MP-7 clarification gate: `[NEEDS CLARIFICATION: <topic>]` 형태의 미해결 마커 0건. 단 `plan.md:L19` 의 설명 문장이 리터럴 `[NEEDS CLARIFICATION]` 을 담아 문자 그대로의 grep 에 걸린다(D20, optional).

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | `U` 의 출처 미정의(acceptance.md:L190 vs L28), F 의 산식이 금지된 외삽을 담는지 불명(spec.md:L70 vs L69), 조건 1 정의에서 선행 SPEC 의 단서 탈락(spec.md:L55), `P절` 선택지 둘로 축소(spec.md:L70) — 합리적인 구현자가 서로 다르게 실행할 수 있다 |
| Completeness | 0.75 | 0.75 | HISTORY·배경·용어·REQ·범위 밖(`### Out of Scope — …` 4개, spec.md:L84-99)·성공 기준·AC 모두 있음. 결손: 런타임 계수기와의 대조 요구, 목적지 열, 상한 열 |
| Testability | 0.50 | 0.50 | AC-ALH-001·002 의 기계 검사가 plan 단계 뼈대만으로 이미 충족(verdict.md:L3-8, L31, L35). AC-ALH-003 은 1행 TSV 로도 `BAD=0`. `UNDETERMINED` 우회 가능. AC-ALH-005 가 살아 있는 선택지를 금지 |
| Traceability | 1.00 | 1.00 | acceptance.md:L224-240 — REQ 15개 모두 AC 에 매핑, AC 8개 모두 존재하는 REQ 를 가리킴. 고아 없음 |

## Defects Found (structured defect-list)

D1. COUNT-SET — spec.md:L30·L43, acceptance.md:L65-84 — 계수 집합 18경로가 런타임이 실제로 세는 집합과 대조되지 않았다. 이 기준 트리에서 `workflow/skill-routing.md` 는 라이브·템플릿 모두 `paths:` frontmatter 를 가진 path-scoped 룰이다(각 4,966자, `paths:` 없는 룰은 두 트리 모두 13개). 선행 SPEC 이 인용한 런타임 경고는 스스로 파일 수와 합계를 찍는다(DIET-002 spec.md:L55 「18 instruction files add up to 249.2k chars」 — `skill-routing` 템플릿에 frontmatter 가 없던 시점). 그런데 이 SPEC 에는 `S_init` 의 `wc -m` 합계를 런타임이 보고하는 파일 수·합계와 맞대는 AC 가 없고, §D.3(acceptance.md:L244)은 「경고가 세는 집합이 다르면」을 가정형으로만 둔다. 이것은 가설이다 — 런타임이 CSV 문자열 `paths:` 를 존중하는지는 이번 감사에서 관측하지 않았다. 그러나 존중한다면 잔여 49,111 중 4,966 이 계수 대상이 아니며, 판정 토큰이 경고의 실체와 어긋날 수 있다. — Severity: major — Class: blocking — Required fix: `S_init` 에서 런타임이 보고하는 파일 수·합계를 관측해 판정서에 적는 REQ/AC 를 추가하고(관측이 불가하면 그 사실을 Gap 으로), 18경로 합계와 다르면 어느 집합으로 `T_min` 을 내는지 규칙으로 고정한다. `skill-routing.md` 를 계속 세려면 그 근거를 증거와 함께 적는다.

D2. S_INIT-EXEC — acceptance.md:L88-95 — 1차 판정 표면 `S_init` 의 산출 절차가 워크트리 레인에서 그대로는 실행되지 않고, 대체 경로가 위험하다. (i) `HOME=<scratch> <binary> init …` 형태는 이번 감사 세션에서 워크트리 격리 가드가 거부했다(위 재측정 표). run 레인도 같은 가드 아래에 있다. (ii) L95 는 HOME 격리를 빠뜨려도 Gaps 에 적기만 하면 된다고 허용하는데, runInit 은 user-scope settings·profile ledger·`MOAI_HOME`·셸 rc 네 곳에 쓴다(`internal/cli/init_home_guard_test.go:6-9`). 비격리 실행은 운영자의 실제 홈을 바꾼다. (iii) `moai init <project-name>` 은 「새 폴더를 만든다」로 문서화돼 있고 절대경로 수용은 미확인이며 `--root` 가 따로 있다. slim 기본값과 `--all` 중 무엇이 사용자 표면인지도 고정되지 않았다. — Severity: major — Class: blocking — Required fix: 비격리 `moai init` 실행을 FAIL 조건으로 명시하고, 가드와 양립하는 격리 경로 하나를 지정한다(예: t1184 처럼 `HOME`·`CLAUDE_CONFIG_DIR` 를 돌려 세운 비워크트리 세션에서 리드가 실행, 또는 기존 `prepareSafeInitHome` 류 격리 헬퍼를 쓰는 커밋하지 않는 일회성 하네스). 플래그 전체(`--non-interactive`, slim/`--all`, `--llm`, `--name`, 경로 인자 또는 `--root`)를 고정한다.

D3. BIN-PROVENANCE — acceptance.md:L102 — 「설치 바이너리로 재면 FAIL」에 기계 증거가 없다. `go build` 는 ldflags 없이 빌드되므로 `moai version` 이 커밋을 보여 주지 않는다(`Commit="none"`), 버전 출력으로 출처를 증명할 수 없다. — Severity: major — Class: blocking — Required fix: 판정서에 `build_head = <git rev-parse HEAD>`, 빌드 명령 원문, `binary_sha256 = <shasum -a 256 결과>`, init 호출 줄의 바이너리 절대경로를 적게 하고 그 줄들을 grep 으로 검사한다.

D4. VACUOUS-HEADERS — acceptance.md:L50-54·L98-99, verdict.md:L3-8·L31·L35 — AC-ALH-001 의 grep 다섯 개와 AC-ALH-002 의 grep 두 개가 plan 단계 뼈대만으로 이미 모두 기대값을 낸다. 측정이 하나도 없어도 기계 판정이 통과하므로 이 두 AC 의 기계 부분은 주장을 확립하지 못한다. 「`total` 관측 줄」은 산문 FAIL 조건에만 있다. — Severity: major — Class: blocking — Required fix: `total_init = N` · `total_live = N` 기계 줄을 요구하고, `total_live` 가 재측정값(기대 199111)과 같은지, `## S_init`·`## S_live` 절 안에 `_미측정` 자리표시가 0건인지를 grep 으로 검사한다.

D5. BACKEND-LITERAL — spec.md:L64, acceptance.md:L50, verdict.md:L3 — REQ-ALH-001 과 AC-ALH-001 이 레인 백엔드 줄의 **값**을 `Claude Opus 5.5 (claude-opus-5-5)` 로 고정했고, plan 단계 뼈대가 run 레인이 존재하기도 전에 그 값을 이미 단언한다. run 레인이 다른 모델로 서빙되면 AC 는 거짓 기재를 하거나 FAIL 하는 두 길만 남긴다 — `verification-claim-integrity.md §1` 의 관측 없는 주장을 강제하는 구조다. — Severity: major — Class: blocking — Required fix: 형식을 `레인 백엔드: <관측값> (출처: <관측 방법>)` 으로 바꾸고 AC 는 줄의 존재와 형식만 검사한다. 뼈대의 값은 자리표시로 되돌린다.

D6. TABLE-COMPLETENESS — spec.md:L66, acceptance.md:L104-126 — REQ-ALH-003 은 「모든 비구속 절과 문단」을 요구하지만 AC-ALH-003 은 행 형식만 본다. 헤더와 1행짜리 TSV 도 `BAD=0` 이다. `bind` 열은 자기 신고라 재계산되지 않는다. `REJECT` 행은 사유 칸이 비어 있지만 않으면 통과하므로, `bind=0`·`c1~c4=Y` 인 행을 임의 사유로 기각해 `A_adm` 을 줄일 수 있다 — `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 를 거짓으로 만들 수 있는 방향이다. — Severity: major — Class: blocking — Required fix: TSV 를 커밋된 스크립트가 생성하도록 요구하고, AC 가 기준 커밋의 각 표면에서 절 분할을 다시 돌려 (file, section, bind, gross 자수) 키 집합을 TSV 와 diff 한다. `REJECT` 는 실패한 열(`bind>0` 또는 `c*=N`)이나 열거된 사유 집합(`governs-scope`·`narrows-scope`·`net-negative`·`rewraps-binding-line`·`dest-over-40k`…) 중 하나를 요구한다.

D7. COND1-PROVISO — spec.md:L55, acceptance.md:L31 — 조건 1 을 「구속 조항 줄 0」으로만 정의해 선행 SPEC 의 단서가 빠졌다. `SPEC-ALWAYS-LOADED-DIET-002/plan.md:L80` 은 구속 조항이 없어도 (a) 구속 조항이 의존하는 절, (b) 구속 조항의 범위를 좁히는 문장은 옮기지 않는다고 적고, `design.md:L554-568` 이 그렇게 기각한 절들(예: `kanban-dispatch.md ## Scope — when this rule is live` 1,309)을 나열한다. 이 단서가 없으면 그 절들이 `ADMIT` 되어 `A_adm` 이 부풀고 거짓 `ACHIEVABLE` 로 갈 수 있다. — Severity: major — Class: blocking — Required fix: 조건 1 정의에 단서 (a)/(b) 를 승계하고 TSV 에 `gov` 열(`N`/`a`/`b` + 근거)을 추가해 `ADMIT` 은 `gov=N` 을 요구한다. `design.md §4.3` 기각 목록을 출발점으로 두 표면에서 재확인한다.

D8. DEST-CAPACITY — spec.md:L72·L78, acceptance.md:L24-37·L124 — 조건 4 가 행 단위다. 같은 companion 으로 가는 `ADMIT` M1 행 여럿이 각자는 40,000 아래여도 합쳐서 넘을 수 있다. 또 REQ-ALH-009 는 `R` 을 목적지별로 묶어 재라고 하는데 TSV 에 목적지 열이 없어 그 묶음이 재현되지 않는다(목적지는 `reason`/`evidence` 산문에만). — Severity: major — Class: blocking — Required fix: `dest` 열을 추가하고, 목적지별로 `현재 크기 + Σ(허용 자수) ≤ 40,000` 을 라이브·템플릿 두 트리에서 각각 검사하는 awk 를 AC-ALH-003 에 넣는다. `R` 도 `dest` 로 묶어 재현한다.

D9. UNDETERMINED-GAMING — spec.md:L69·L73·L74, acceptance.md:L28·L190·L220 — `U` = 「`미측정` 행 자수 합」인데 그 행의 `chars` 칸은 정의상 문자열 `미측정` 이라 `U` 의 출처가 없다. 또한 `미측정` 행 수에 제한이 없고 사유도 요구하지 않는다. 모든 M2 행을 시도하지 않고 두면 `UNDETERMINED` 가 되고, 상신 절차(REQ-ALH-011)는 `INFEASIBLE` 에서만 발동하므로 건너뛰며, §D.1 에 따라 전 AC 가 통과한다. 카드 임무(불가능하면 상신 절차 필수)를 비켜 가는 경로다. — Severity: major — Class: blocking — Required fix: TSV 에 `gross` 열(후보 문단의 전체 자수 — 제거 가능량의 건전한 상한)을 추가하고 `U = Σ gross(미측정 행)` 로 정의한다. `미측정` 행마다 시도 불가 사유를 요구한다. 토큰 규칙을 판정서 줄(`current`·`A_adm`·`R`·`U`·`T_min`)에서 기계적으로 재계산하는 awk 를 AC-ALH-007 에 넣는다. `verdict_init = UNDETERMINED` 일 때도 `미측정` 목록과 `U` 를 상신 자료로 넘기도록 규정한다.

D10. P-RECON-OPTIONS — spec.md:L70, acceptance.md:L153, plan.md:L45 — 재조정 선택지를 (a)/(b) 둘로 잘랐지만 선행 표는 세 행이다(`SPEC-ALWAYS-LOADED-DIET-002/acceptance.md:L110-114` — (a), 공표값, (b); F 도 세 값). 산술로 보면: (a) 의 M1 상한 71,817 에서 `J + A = 92,332 − 71,817 = 20,515`, 공표값 `93,641 − 20,515 = 73,126`. kanban 칸이 1,309 를 선차감했고(`design.md:L282`) `J` 도 그 1,309 를 담는다면(`design.md:L558`), (a) 는 그 절을 두 번 빼고 (b) 는 한 번도 빼지 않으며, 공표값 행이 한 번 빼는 읽기다. 이 해석은 가설이다(재실행하지 않았다). 그러나 AC-ALH-005 의 정규식은 이 세 번째 선택지를 금지한다. 또 `sec.py` 재실행은 절별 원자수만 주고 `J` 의 구성은 알려 주지 않으므로, 재실행만으로는 판별되지 않는다. — Severity: major — Class: blocking — Required fix: 허용 토큰을 `(a)|(b)|(공표값)|(기타)` 로 넓히고, `J` 가 kanban Scope 행을 포함하는지를 `design.md §4.3` 인용과 재실행 출력으로 함께 밝히게 한다.

D11. F-DEFINITION — spec.md:L69-70·L77, acceptance.md:L154 — REQ-ALH-007 이 F 를 단일값으로 요구하지만 F 의 산식(`DIET-002/acceptance.md:L86-88`)은 `M2 기저 × 10.77%` 를 쓴다. 바로 위 REQ-ALH-006 이 금지한 수율 외삽이다. F 는 `172ef22eb` 트리의 양인데 REQ-ALH-014 는 다른 트리 수치를 이 트리 측정값으로 쓰지 말라고 하며, `^F = N$` 줄에는 트리 귀속이 없다. — Severity: major — Class: blocking — Required fix: F 가 `172ef22eb` 트리의 역사적 하한이며 선행 산식(외삽 포함, 그렇게 표시)으로 계산하고 `T_min`·판정 토큰의 입력이 아님을 명시한다. 줄 형식을 `F(172ef22eb) = N` 으로 바꾼다.

D12. TMP-BAN-CONFLICT — acceptance.md:L86·L155, plan.md:L92 — AC-ALH-005 는 판정서 전체에서 `/tmp/` 가 0회이기를 요구한다. 그런데 AC-ALH-002 는 「실제 실행한 명령을 그대로」 옮기라고 하고, 세션 scratchpad 는 `/private/tmp/…` 아래에 있다. plan §E 는 sec.py 가 사라졌을 때 그 사실을 Gaps 에 적으라고 한다. 충실히 기록할수록 AC-ALH-005 가 FAIL 하거나, 반대로 출처를 숨기도록 유도된다. — Severity: major — Class: blocking — Required fix: 금지 대상을 「`/tmp` 경로를 증거로 인용하는 것」으로 좁혀(예: `evidence` 열과 판정서 증거 줄에 대해서만 검사) 기록·출처 서술과 분리하거나, 경로를 `$SCRATCH` 표기로 쓰도록 명시한다.

D13. SEC-PY-PROVENANCE — research.md:L27, plan.md:L44·L92 — `sec.py` 는 지금 `/tmp/claude-501/sec.py` 에 있다(1,661 bytes, sha256 `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547`, 감사자 실측). 그러나 research.md 는 크기와 시각만 적고 해시를 적지 않았고, 복사는 run 단계 M2 로 미뤘다. 그 사이 사라지면 재작성본은 `design.md §4.0` 의 출력을 재현한다는 보장이 없고, 커밋된 사본이 원본과 같다는 것도 증명할 수 없다. — Severity: major — Class: blocking — Required fix: 위 sha256 을 research.md 에 기록하고, Implementation Kickoff 전에 `.moai/reports/t1226/sec.py` 로 복사·커밋한다. AC-ALH-005 에 `shasum -a 256 .moai/reports/t1226/sec.py` 가 그 값과 같다는 검사를 넣고, 재작성한 경우에는 Gap 과 차이 설명을 요구한다.

D14. AC008-ABSORPTION — acceptance.md:L200-207 — AC-ALH-008 은 `7fe658815..HEAD` 를 `.claude/rules/moai/core`·`workflow` 와 템플릿 룰 디렉터리 **전체**에 대고 diff 한다. 레인 규약은 병합 전에 `git merge develop` 흡수를 요구한다(`CLAUDE.local.md §4.1`). 흡수된 develop 이 이 디렉터리의 어느 파일이든 바꾸면(예: 40,000자 초과 companion 을 고치는 t1180 류) 이 카드가 아무것도 편집하지 않았어도 `exit≠0` 이 된다. 반대로 흡수가 18경로를 바꾸면 측정값이 조용히 낡는다. 지금은 `develop == 7fe658815` 라 드러나지 않을 뿐이다. — Severity: major — Class: blocking — Required fix: 카드가 작성한 비병합 커밋만 본다(`git log --no-merges --format=%H 7fe658815..HEAD -- <18경로 + 미러>` 가 빈 출력). pathspec 을 18경로와 그 미러로 좁힌다. 흡수가 18경로를 바꿨다면 재측정하거나 기준 커밋 고정을 판정서에 명시하도록 규정한다.

D15. AGENTS-M1P — spec.md:L67, acceptance.md:L117 — REQ-ALH-004 는 `AGENTS.md` 후보를 「M2 로만」 허용하는데 awk 는 `$4=="M1"` 만 잡는다. `M1p` 로 표기된 `AGENTS.md` 행은 통과한다. — Severity: minor — Class: blocking — Required fix: 조건을 `$1 ~ /AGENTS\.md/ && $4!="M2" && $10=="ADMIT"` 로 바꾼다.

D16. HASH-PER-SURFACE — acceptance.md:L141, plan.md:L27-29 — AC-ALH-004 는 두 TSV 모두에 라이브 해시 `d97b33d9…c6c3` 유지를 요구하지만, plan D2 는 `S_init` 후보를 그 표면 자신의 기준선 해시로 판정한다고 적는다. 두 해시가 다르면(plan 이 그 가능성을 인정한다) `S_init` 의 M2 행은 AC 를 만족할 수 없다. — Severity: minor — Class: blocking — Required fix: 표면별 기대 해시(`hash_init`·`hash_live`)를 판정서 줄로 적게 하고, AC-ALH-004 가 해당 표면의 값을 쓰게 한다.

D17. EVIDENCE-EXISTENCE — acceptance.md:L119·L136 — 증거 파일 존재 검사가 `ADMIT` M2 행에만 있다. M1·M1p·REJECT 행의 `evidence` 는 빈 칸만 아니면 된다. — Severity: minor — Class: optional — Required fix: 존재 검사를 모든 행으로 넓힌다.

D18. MIN-SET-GREEDY — spec.md:L74, plan.md:L68 — 「최소 해제 집합」을 비율 순 정렬로 구한다. 여러 구속 줄이 같은 절을 묶는 경우(겹침)에는 탐욕 정렬이 최소를 보장하지 않는다. — Severity: minor — Class: optional — Required fix: 「탐욕 해제 집합」이라 부르거나, 최소의 정의(줄 수 기준)와 겹침 처리 규칙을 적는다.

D19. SPLIT-COVERAGE — spec.md:L66, `/tmp/claude-501/sec.py` L2-18·L31-33 — 절 분할 스크립트는 16개 마크다운만 다루고 첫 제목 앞의 서문과 yaml 두 개를 세지 않는다. `wc -m` 의 로케일도 고정되지 않았다(C 로케일이면 바이트를 센다). — Severity: minor — Class: optional — Required fix: 서문과 yaml 을 후보 열거에 포함하거나 제외 사유를 적고, 측정 명령에 UTF-8 로케일을 고정한다.

D20. NC-LITERAL — plan.md:L19, progress.md:L15 — 설명 문장이 리터럴 `[NEEDS CLARIFICATION]` 을 담아 MP-7 의 문자 grep 에 걸린다(마커는 아님). — Severity: minor — Class: optional — Required fix: 「clarification 마커」 같은 표현으로 바꾼다.

D21. OWNERSHIP-TRAILER — 커밋 `10281a857` — 린트 INFO `OwnershipTransitionUnmeasured`: `Authored-By-Agent` 트레일러가 없다. — Severity: minor — Class: optional — Required fix: 다음 SPEC 커밋에 트레일러를 단다.

## Regression Check

해당 없음(1회차).

## 확인한 항목 중 결함이 아닌 것

- 시간 예측: 다섯 파일과 뼈대에서 기간 표현 grep 0건. 우선순위 라벨(High/Medium/Low)만 쓴다(plan.md:L35-72).
- `-run` 앵커 규칙: `go test` 를 쓰는 AC 가 없다(grep 결과는 progress.md:L26 의 설명 한 줄뿐). 대상 0.
- 상신 경계: REQ-ALH-011(d)·REQ-ALH-012(spec.md:L74-75)와 plan M6(plan.md:L70)가 결정권자를 운영자로, 상신 경로를 리드의 `AskUserQuestion` 으로 두고, 레인의 결론을 `RECOMMEND:` 로 제한한다. AC-ALH-007 의 결정 동사 음성 grep(acceptance.md:L187)은 약하지만 검토 항목과 함께 경계를 지킨다.
- D1 결정(`S_init` 1차)의 근거: 카드 본문의 「템플릿 rules 풀」, 선행 SPEC 의 격리 `moai init` 경고 관측(DIET-002 spec.md:L53-57), `.tmpl` 3개 때문에 렌더링 없이는 사용자 실물을 잴 수 없다는 점(research.md:L17, 감사자 재측정 203,611 일치) — 근거 자체는 타당하다. 결함은 그 표면을 **재는 방법**(D1·D2·D3)에 있다.
- 18개 계수 파일 무수정: `git diff --stat 7fe658815 HEAD` 에 계수 파일 0건.

## Recommendation

FAIL. 필수 통과 7개는 모두 통과했지만, 판정 토큰의 정확성과 AC 의 반증 가능성을 직접 해치는 blocking 결함이 16건(D1-D16)이다. 점수 0.62 는 Tier M 문턱 0.80 아래다. manager-spec 수정 지시:

1. 계수 집합을 런타임 계수기와 대조하는 요구를 추가하고, `skill-routing.md` 처리 규칙을 정한다(D1).
2. `S_init` 산출을 가드와 양립하는 격리 경로 하나로 고정하고, 비격리 실행을 FAIL 로 둔다. 플래그와 바이너리 출처 증거 줄을 규정한다(D2, D3).
3. AC-ALH-001/002 에 측정 후에만 참이 되는 기계 줄(`total_*`·`build_head`·`binary_sha256`)을 넣고, 백엔드 줄을 관측값 형식으로 바꾼다(D4, D5).
4. TSV 에 `gross`·`dest`·`gov` 열을 추가하고, 스크립트 재생성 diff·REJECT 사유 열거·목적지 누적 수용량·토큰 규칙 재계산을 AC 에 넣는다(D6-D9, D15, D16).
5. `P절` 선택지를 세 행 이상으로 넓히고, F 를 원 트리의 역사적 하한으로 격리한다(D10, D11).
6. `/tmp/` 금지를 증거 인용으로 좁히고, sec.py 해시를 지금 기록한 뒤 kickoff 전에 커밋한다(D12, D13).
7. AC-ALH-008 을 카드 작성 커밋 기준으로 바꾼다(D14).

optional D17-D21 은 오케스트레이터 재량이다.
