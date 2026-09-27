# SPEC Review Report: SPEC-SESSION-MIDMOVE-001
Iteration: 2/3 (Tier L ceiling)
Verdict: FAIL
Overall Score: 0.74

- 감사 대상: `.claude/worktrees/t1279` HEAD `75f761f43` (브랜치 `WT-session-double-load`). `git merge-base develop HEAD` = `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`(= 현재 로컬 develop tip).
- 입력(Tier L 5종 전부 읽음): spec.md, plan.md, acceptance.md, design.md, research.md, progress.md. 전제 대조: `.moai/reports/t1279/verdict.md` §①–§④. 1회차: `.moai/reports/t1279/plan-audit.md`.
- 작성자의 수정 주장("D1–D22 해결", "AC는 가드를 피해 작성됨")은 M1 Context Isolation에 따라 입력으로 쓰지 않았고, 모두 직접 재측정했다.
- 교차 모델: `mcp__moai__audit_multi`(project_root = 이 워크트리). codex는 도구 판정 필드가 `inconclusive`지만 요약은 FAIL이며 결함 9건을 냈다. 이 감사의 N1–N6과 독립적으로 겹치고, 새로 N3의 후행 명령 뮤턴트, N2의 `listing_count=0` 뮤턴트, AC-015의 짝 비교 결함을 더했다. 이 감사에서 모두 재현하거나 코드로 확인했다. GLM은 리뷰할 diff가 없어 `inconclusive`였다. 판정은 이 감사자가 내린다.
- 도구 경고: 실행 중인 moai MCP 바이너리는 `a8a9b9376`(HEAD의 조상)에서 빌드된 것이다. spec lint는 `go run`으로 현재 트리에서 따로 실행했다.

## 1. D1–D22 해결 여부 (재측정 근거)

| 결함 | 판정 | 근거 |
|---|---|---|
| D1 BASE 리터럴 | 해결 | `CARD_BASE=b59a5d69c…; git rev-parse --verify "$CARD_BASE^{commit}"` → SHA 출력, rc 0. `git diff --name-only "$CARD_BASE"..HEAD \| wc -l` → `9`. 음성 대조 `git rev-parse --verify "BASE^{commit}"` → `fatal: Needed a single revision`, rc=128. AC-SMM-021에 「이 항목이 FAIL이면 범위 기반 항목도 FAIL」 규칙이 있다(acceptance.md:L41) |
| D2 흡수 후 범위 | 해결 | 읽는 시점의 merge-base(acceptance.md:L22, plan.md:L21), M3 G1에서 §3 동일성 재측정과 §D 재측정(plan.md:L73-78) |
| D3 예산 여유 | 해결 | 재측정: `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)`, `--- PASS:` 2건. 상수: `grep -c '^const AlwaysLoadedTokenBudget = ' internal/config/token_budget_guard.go` → `1`, `git diff "$CARD_BASE"..HEAD -- internal/config/token_budget_guard.go \| wc -l` → `0`. REQ-SMM-015의 순증 0 규칙(spec.md:L115-118) |
| D4 명령 형식 | **부분 해결 → N3** | 파일 분리(REQ-SMM-007), 전사 `cwd` 검사(AC-SMM-003)는 들어갔다. 그러나 프로브 줄 형식으로는 시작 cwd를 적을 수 없다 |
| D5 대조가 추출기를 거치지 않음 | 해결 | JSONL 모양의 양성·음성 대조를 같은 추출기로 돌린다(AC-SMM-005). 실제 전사 줄 404의 `names` 필드에 접두 이름이 42개 있음을 확인했다(`names` 60개 중 `^\.claude/worktrees/[^:]+:` 42). 따라서 names만 남긴 대조도 성립한다. 발동 조건이 REQ-004의 OR와 일치한다 |
| D6 tree 정의 | **부분 해결 → N2** | 네임스페이스 제외 규칙은 들어갔다(spec.md:L67). 대신 이름 기반 귀속이 새 결함을 만들었다 |
| D7 측정 스키마 | 해결(잔여 N2) | 합성 `$E`로 AC-002 세 명령을 실제 실행해 `12`, `12`, `3`을 얻었다. 바이트·세션 id·정규식이 모두 있다 |
| D8 비용 요구 | 해결 | REQ-SMM-008, AC-SMM-009. 수치는 §3에서 재현했다 |
| D9 REQ015↔plan | 해결 | REQ-009/018이 CLAUDE.local.md를 포함하고, AC-010이 `$CL`을 검사한다(기준 0). LP의 공허 검사는 AC-012의 「또는 EnterWorktree」 1→0으로 바꿨다 |
| D10 [HARD] 보존 | 해결(잔여 N5) | AC-018 명령을 실행했다. 추출 3줄, 누락 `0`, `asks→tells` 변이 사본에서 누락 `1`. 작성자가 적은 관측값(3/0/1)과 일치한다 |
| D11 DP-1 비대칭 | 해결 | 표의 두 열이 면제 판별과 모드 게이트를 똑같이 요구한다(spec.md:L196-197). 출력 채널 인용을 확인했다: `internal/hook/types.go:333` `AdditionalContext`, `:366` `SystemMessage … // Warning message shown to user`, `post_tool_worktree.go:48`은 `SystemMessage`만 반환 |
| D12 Kanban 구조 | **부분 해결 → N1** | 상설 세션을 재기동하지 않는 흐름을 정했다(REQ-010, design.md §2). 그러나 「/clear 한 번」 주장이 REQ-017의 원문 동결과 충돌한다 |
| D13 시간대 | 해결 | `%ct` epoch 비교와 ISO→epoch 교차 검사(AC-SMM-001) |
| D14 자리표시자 | 해결(잔여 N8) | 1회차에 지적한 6곳이 모두 명령으로 바뀌었다 |
| D15 `-d` 누락 | 해결 | 양성 대조 실행 → `2` |
| D16 대조 한 줄 | 해결 | 한 줄에 대안 하나씩 넣은 대조 실행 → `4` |
| D17 바이트 기준 | 해결 | `content` UTF-8 길이로 정의했다. 43,566 B 재현(§3) |
| D18 템플릿 주장 무조건 | 해결 | REQ-012가 P2 두 트리를 조건으로 건다 |
| D19 문구 | 해결 | REQ-005 「shall either … or」, REQ-019 「the hook layer shall not change」 |
| D20 Tier | 해결 | `tier: L`, REQ 19 / AC 21, design.md·research.md 추가 |
| D21 상한 계산 | 해결(잔여 N9) | 세션 수 = 고유 전사 id, 벽시계는 전사 첫 행부터 마지막 행까지 |
| D22 DP-2 비용 | 해결 | spec.md:L206 |

해결 19, 부분 해결 3(D4, D6, D12). Retry Loop Contract에 따라 전 회차 결함이 완전히 해소되지 않으면 이번 회차는 FAIL이다.

## 2. Must-Pass Results
- [PASS] MP-1 REQ 번호: `REQ-SMM-001`…`019`가 공백·중복 없이 이어진다. AC도 `AC-SMM-001`…`021` 전부 1회씩 있다(grep 추출).
- [PASS] MP-2 GEARS(요구 계층에서 판정): 19개 REQ 모두 패턴 라벨이 있고 shall 구조다. 예: REQ-010 While(spec.md:L96), REQ-013 Where(L110), REQ-019 Where와 docs-only 문장 「the hook layer shall not change」(L133). AC의 Given-When-Then은 이 기준으로 평가하지 않았다.
- [PASS] MP-3 frontmatter: spec.md:L2-L13에 12개 정규 필드가 있다(`version: "0.2.0"`, `status: draft`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 문자열). `tier: L`은 L14. lint 경고는 0건이다.
- [N/A] MP-4 언어 중립성: 다중 프로그래밍 언어 도구를 다루지 않는다.
- [PASS] MP-5 D7: 참조 SPEC은 `SPEC-SESSION-DOUBLELOAD-001` 하나이고 `status: draft`다.
- [PASS] MP-6 D8: 5개 산출물의 `syscall` 출현 수는 모두 0이다.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → rc=1(일치 없음).

**spec lint**: `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` → **rc=0**, `0 error(s), 0 warning(s)`. INFO 1건: `OwnershipTransitionUnmeasured`(커밋 `a14fcf851`에 `Authored-By-Agent` 트레일러가 없음).

## 3. 비용 주장 재측정 (전사 bb145fe7)

scratchpad의 `cost2.py`로 `…/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t1279/bb145fe7-….jsonl`을 읽었다.

- 줄 404: `ts=2026-09-27T03:57:49.175Z skillCount=60 content_bytes=43566 scoped=42 before=278878 after=299707 delta=20829`. spec.md:L34-35의 네 수치가 **정확히 재현**된다.
- 사이 행 바이트: 개행 포함 **20,547**, 사이 21행의 개행을 빼면 20,526. SPEC의 20,526은 개행 제외 기준이다. REQ-SMM-008은 이 바이트 기준을 정의하지 않는다(N7).
- 줄 1126: 55 skills, 45,703 B, 접두 이름 44개(t1279 43개, t1219 1개). spec.md:L36과 research.md 표가 맞다.
- **「upper bound」 전제**: delta는 두 턴 사이 프롬프트 증가분이므로, 그 사이에 프롬프트 내용이 제거되지 않을 때만 상한이다. 줄 392–415 사이에 compact 행은 없었으므로 줄 404에 대해서는 성립한다. 그러나 REQ-008은 모든 행에 이 라벨을 강제하고(AC-009 정규식 `bound=upper$`), 전제를 적지 않았다. 또 줄 1000은 사이 행 150 KB에 delta가 22.8k 토큰이다. 전사 행 바이트는 프롬프트 토큰의 척도가 아니다(queue-operation, custom-title 같은 비프롬프트 행이 섞인다). 20,829를 상한으로 보고한 것은 줄 404에 대해서는 옳다. 다만 일반 규칙으로는 근거가 부족하다(N7).

## 4. 가드 아래 실행 가능성

| 형식 | 결과 |
|---|---|
| `CARD_BASE=<sha>; git rev-parse --verify "$CARD_BASE^{commit}"` | 실행됨 |
| `CARD_BASE=<sha> && git rev-parse … ; git diff … \| wc -l; BAD=…` (복합) | **거부됨** |
| `git show "$CARD_BASE:$KT" \| awk '…' > file` (AC-018) | 실행됨 |
| `git diff -U0 … \| grep … \| grep -cE …` (AC-017) → `0` | 실행됨 |
| `git log --no-merges … -- "$KT" \| wc -l` (AC-015 뒷부분) → `0` | 실행됨 |
| `comm -23 <(git log …) <(git log …) \| wc -l` (**AC-015 앞부분**) | **거부됨**: 「names git in a form too complex to verify」 |
| `diff <(sed …) <(sed …)` (AC-015 §3) → rc 0, 36/36줄 | 실행됨 |
| AC-012 여러 줄 `python3 -c` → `5` | 실행됨(작성자 기준값 5와 일치) |
| `$(grep … .claude/rules/local/gitflow-lane-protocol.md)` | **거부됨**: 경로 문자열 `gitflow` 안의 `git`에 반응한다(참고 관측, AC에는 없는 형식) |

「AC는 가드를 피해 작성됐다」는 주장은 AC-SMM-015 한 곳에서 거짓이다(N4).

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | 0.75 | REQ-010과 REQ-017이 /clear 시점을 두고 충돌한다(spec.md:L96-102 vs L120-121과 kanban-dispatch.md:153). 「tree」와 「listing」이 섞였다(L67 vs L105-107). 「turn」이 정의되지 않았다(L58) |
| Completeness | 0.85 | 0.75–1.0 | 절·frontmatter·Out of Scope(H3 6개, spec.md:L161-185)와 Tier L 산출물이 모두 있다. fixture의 이름 분리 조건과 프로브 cwd 설정 방법이 빠졌다 |
| Testability | 0.60 | 0.50–0.75 | AC-015는 가드 아래 실행되지 않는다. 뮤턴트가 쓰이는 AC가 AC-002, 003, 007, 012, 018, 009 여섯 곳이다. AC-006에 산문 계산이 남아 있다 |
| Traceability | 0.90 | 0.75–1.0 | 19 REQ ↔ 21 AC 표가 완전하다(acceptance.md:L103-123). LP의 진입 형식 검사가 AC-010에서 빠졌다(N10) |

조화평균 = 4 / (1/0.70 + 1/0.85 + 1/0.60 + 1/0.90) = **0.74**. Tier L 기준 0.85 미만이다. 1회차 0.55보다 높으므로 STOP 신호는 없다.

## Defects Found (structured defect-list)

N1. CLEAR-ORDER-VS-VERBATIM — spec.md:L96-L102, L120-L121; design.md:L28; `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`:L153 — REQ-010은 상설 세션의 카드 전환 /clear를 「이동 뒤 단 한 번이며 이전 clear→이동 순서를 대체한다」고 정한다. 그런데 REQ-017은 「The `/clear` handoff between phases」 절의 모든 비지 않은 줄을 원문 그대로 두라고 요구한다. 그 절 첫 줄은 [HARD]이며 「When a phase completes and the lead has read its evidence, the lead asks the operator to `/clear` that session … which session is instructed once the clear is done」라고 적혀 있다. 즉 phase 완료 직후 /clear, 그다음 지시라는 순서다. Kanban companion에게 phase 완료는 곧 카드 사이다. 따라서 두 문장이 공존하면 카드 전환마다 /clear가 두 번이거나 순서가 상충한다. design.md §2의 「기존 /clear를 이동 뒤로 옮긴다」는 원문 동결 아래에서는 불가능하다. 1회차 D12 후반부(두 순서가 합쳐지면 어떻게 되는가)가 해소되지 않았다. codex도 독립적으로 같은 결함을 냈다. — Severity: major — Class: blocking — Required fix: REQ-017에 이 [HARD] 첫 줄만을 위한 좁은 개정 예외와 정확한 최종 문구를 정의한다(예: 「When a phase completes … the lead asks the operator to `/clear` that session — for a session that next moves into another card's worktree, after that move」). AC-018의 원문 대조에서 그 한 줄만 새 문구와 비교하도록 바꾼다. 또는 반대로 REQ-010에서 「대체한다」를 빼고 /clear 횟수를 명시한다. 어느 쪽이든 두 REQ가 같은 파일에서 동시에 참일 수 있음을 한 문장으로 보인다.

N2. TREE-ATTRIBUTION — spec.md:L61, L67, L105-L107; plan.md:L34-L36, L60-L62; design.md:L36-L38; acceptance.md:L46, L90 — 결함은 두 겹이다. ① 이름 기반 귀속은 물리 트리를 구분하지 못한다. `git worktree add`로 만든 w1은 fixture 커밋의 `.claude/skills/fx-primary-*`를 그대로 체크아웃한다(codex가 실제 fixture로 확인). w1에서 시작한 P1은 목록이 하나여도 `primary,wt`로 기록되고, P3도 마찬가지다. 그러면 REQ-012 판정이 체계적으로 `no`가 된다. SPEC과 plan은 w1이 `fx-primary-*`를 갖지 않도록 만드는 조건을 적지 않았다. ② 판정 키 세 개가 경로별 **합집합** 트리 집합에서 도출된다. 「이동이 두 번째 목록을 더했다」와 「한 목록이 두 이름군을 담았다」가 구분되지 않는다. P-LAUNCH 문구 「gives one skill listing」은 목록 수에 관한 주장인데, 게이트는 트리 수로 건다. 여기에 `listing_count_P*: 0`인데 트리 값이 긍정인 뮤턴트가 AC-002의 `12/12/3`을 그대로 통과한다(정규식이 `[0-9]+`를 허용). §D.1의 「목록 없으면 gap」 규칙이 AC로 강제되지 않는다. — Severity: major — Class: blocking — Required fix: fixture 구성에 「w1 브랜치에서 `fx-primary-*`를 삭제한다(또는 primary 스킬은 w1 분기 뒤에 primary에만 추가한다)」를 요구로 적고, fixture.txt 검사(예: `test ! -e "$FX/.claude/worktrees/w1/.claude/skills/fx-primary-a"`)를 AC로 둔다. 판정 키는 목록별 트리 집합과 `listing_count`로 도출한다. 예: `move_adds_listing=yes` ⇔ `listing_count_P2 ≥ 2` 이고 이동 뒤 목록이 wt를 담음. `launcher_single_listing=yes` ⇔ `listing_count_P1 = 1`. AC-002에 「non-gap이면 listing_count ≥ 1, listing_bytes 항목 수 = listing_count」 검사를 넣는다.

N3. PROBE-CWD-AND-SUFFIX — plan.md:L39-L40; acceptance.md:L47, L51 — ① AC-007은 주석 아닌 모든 프로브 줄이 `unset MOAI_KANBAN … && timeout -k 10 300 claude `로 **시작**하도록 요구한다. 이 형식에는 P1(w1에서 시작)과 P2(fixture 루트에서 시작)의 시작 디렉터리를 적을 자리가 없다. Bash 도구 cwd는 매 호출 워크트리 루트로 돌아가므로, 기록된 줄을 그대로 실행하면 프로브가 **저장소 워크트리에서** 뜬다. 이것은 REQ-006이 금지하고 AC-003이 잡는 바로 그 상황이다. 결국 「기록한 명령 = 실행한 명령」(REQ-003)과 AC-007을 동시에 지킬 수 없다. ② 접두만 보므로 `… claude -p ok; rm -rf /outside` 같은 후행 명령이 통과한다(실측: `grep -vc` → `0`). ③ AC-003의 `c.startswith(r)`는 구분자를 보지 않는다. `/tmp/scr/fx-evil/x`가 `fixture_root: /tmp/scr/fx`를 통과했다(합성 `$E` 실측 → `0`). 또 P1의 첫 cwd가 정확히 w1인지, P2의 첫 cwd가 정확히 루트인지는 검사하지 않는다. — Severity: major — Class: blocking — Required fix: 줄 형식을 `unset … && cd -- <fixture 하위 경로> && timeout -k 10 300 claude …`로 정하고, `cd` 대상이 `fixture_root` 아래임을 검사한다. `claude` 인자 뒤의 `;`, `&&`, `||`, `$(`, 리다이렉션은 거부한다. cwd 포함 검사는 `os.path.commonpath([r, c]) == r`처럼 경로 성분 단위로 하고, `probe_cwd_P1`의 첫 값 = `$fixture_root/.claude/worktrees/w1`, `probe_cwd_P2`의 첫 값 = `$fixture_root`를 따로 검사한다.

N4. AC015-GUARD — acceptance.md:L75; plan.md:L22-L27; research.md:L50 — AC-SMM-015의 `comm -23 <(git log …) <(git log …) | wc -l`은 이 워크트리 세션 가드가 거부한다(재현: 「names git in a form too complex to verify」). 작성자는 가드를 피해 AC를 썼다고 주장하지만 이 항목에서는 사실이 아니다. 비교도 세 로컬 경로의 합집합과 세 템플릿 경로의 합집합을 대조하므로, 로컬 A와 템플릿 B를 같은 커밋에서 고치는 잘못된 짝도 통과한다(codex 지적, 논리로 확인). — Severity: major — Class: blocking — Required fix: 쌍 `(KT,KL)`, `(DT,DL)`, `(WT,WL)`마다 `git log --no-merges --format=%H "$CARD_BASE"..HEAD -- <path> > <scratch file>`를 따로 실행하고, 각 git의 rc를 확인한 뒤 파일끼리 `comm`한다. process substitution 안에 git을 두지 않는다. `AL`↔`AT` 쌍도 같은 방식으로 넣는다.

N5. HARD-OFFSET-WEAKENS — plan.md:L82-L85; spec.md:L120-L124; acceptance.md:L78; kanban-dispatch.md:L179 — plan M4의 상쇄 후보는 new-card [HARD] 문단에서 「세 문장을 **제외한** `EnterWorktree(<card-id>)` cannot run … 설명」을 지우는 것이다. 그런데 그 설명 문장이 곧 「a lane still anchored in the previous card's tree **MUST** `ExitWorktree` … or it does the new card's work on the old card's branch」라는 안전 조항이다. REQ-017의 일반 조항(「shall not remove or weaken an existing [HARD] clause」)과 계획이 충돌한다. AC-018은 세 문장과 [HARD] 개수만 보므로 이 삭제를 잡지 못한다. [HARD] 개수는 무관한 [HARD]를 하나 더해 상쇄할 수 있고, 줄 보존은 파일 **어디든** 있으면 통과하므로 절 밖으로 옮겨도 통과한다(codex 지적). REQ-010 흐름의 「이동」도 이 exit-first 규칙에 기대고 있다. — Severity: major — Class: blocking — Required fix: MUST-ExitWorktree 문장(또는 동일 의미의 개정문)을 REQ-017 보존 목록에 넣고 AC-018에서 대조한다. 상쇄 후보에서 이 문장을 뺀다. 줄 보존은 파일 전체가 아니라 해당 절 범위에서 대조하고, [HARD] 개수 대신 기준 시점의 [HARD] 줄 각각이 남았는지 대조한다.

N6. AC012-LEXICAL-EXCLUSION — acceptance.md:L62-L69; spec.md:L103 — 「이동 뒤 /clear 없는 제시」 계수기는 단위 안에 부분 문자열 `/clear`가 있으면 무조건 제외한다. 실측 결과, 가장 중요한 이동 처방인 new-card [HARD] 문단(`EnterWorktree(<card-id>)` 포함)이 이미 **기준 시점부터 EXCL**이다. 무관한 구절 「reuse without a `/clear` in between」 때문이다(단위별 출력으로 확인). REQ-011은 /clear **와** 포인터 재전송을 요구하는데, 계수기는 /clear만 본다. 「moves with `EnterWorktree(<path>)`; do not `/clear`」 같은 뮤턴트가 통과한다. — Severity: major — Class: blocking — Required fix: 제외 조건을 정식 문구 `move → \`/clear\` → re-send`(P-FLOW) 또는 P-FLOW에 대한 포인터가 있을 때로 바꾼다. 기준 시점 단위별 COUNT/EXCL 목록을 research에 기록하고, new-card 문단이 기준에서 COUNT로 잡히는지 확인한다.

N7. COST-BOUND-PREMISE — spec.md:L35, L86-L91; acceptance.md:L56; research.md:L23 — 「upper bound」는 두 턴 사이에 프롬프트 내용이 제거되지 않는다는 전제 위에서만 성립한다. 이 전제가 적혀 있지 않은데도 모든 행에 `bound=upper`가 강제된다. 「other rows」의 바이트 기준도 정의되지 않았다(20,526 개행 제외 / 20,547 개행 포함). AC-009는 추출기를 다시 돌려 파일과 비교하지 않으므로, 형식만 맞는 가짜 행이 통과한다. — Severity: minor — Class: blocking — Required fix: REQ-008에 「두 턴 사이에 compact/boundary 행이 있으면 `bound=confounded`」 규칙과 행 바이트 기준(개행 제외 등)을 적는다. AC-009에는 전사가 읽히는 동안 `diff <(python3 … --cost <transcript> | grep '^row:') <(grep '^row:' m1-cost.md)`에 해당하는 재실행 대조를 넣는다(git이 없으므로 가드 아래에서 실행된다).

N8. RESIDUAL-PROSE — acceptance.md:L45(`<first_probe_ts value from $E>`), L50(`wall_clock_min` 계산, `probes_run` 범위, operator-run id 대조가 산문), L70(AC-013 합산이 산문), L78(WT- 줄 전체를 CARD_BASE에서 뽑는 명령 없음, `grep -cxF`에는 전체 줄이 필요) — 「Every criterion is … binary」(L3) 선언과 맞지 않는 곳이 남았다. — Severity: minor — Class: blocking — Required fix: 각 자리를 `python3 -c` 한 줄 명령과 기대 stdout으로 바꾼다.

N9. TURN-CAP-UNVERIFIED — spec.md:L58; acceptance.md:L51 — 「at most 4 turns per probe」가 무엇을 한 턴으로 세는지 정의되지 않았다(에이전트 턴인지 `--resume` 호출인지). 어떤 AC도 실제 턴 수나 `--max-turns 4`를 검사하지 않는다. — Severity: minor — Class: optional — Required fix: 턴 = 전사의 assistant 행 수 등으로 정의하고, `$E`의 `turn_tokens_P*` 항목 수 ≤ 4를 검사한다.

N10. LP-ENTRY-UNGUARDED — acceptance.md:L60, L122; spec.md:L95, L125 — REQ-009/018은 lane protocol에도 런처 진입 문구를 요구한다. AC-010은 `$LP`를 뺐다(기준 시점에 이미 1이라서다). 이후 그 문구가 삭제되어도 매핑된 AC가 모두 통과한다. — Severity: minor — Class: optional — Required fix: `$LP`에 대한 회귀 가드(`grep -c 'moai cc -w <card-id>'` ≥ 1, RED-now 아님을 명시)를 AC-010에 추가한다.

N11. AC020-ENV — acceptance.md:L85-L86 — warn의 「without `MOAI_KANBAN`」 경우가 같은 호출 안에서 변수를 해제하지 않는다. Kanban 레인에서 돌리면 주변 환경에 이미 변수가 있다. block의 「flag enabled」에는 설정 명령이 없다. — Severity: minor — Class: optional — Required fix: `unset MOAI_KANBAN … && go run …` 형식으로 쓰고, 플래그를 설정하는 방법을 명시한다.

N12. AC018-RESIDUE — acceptance.md:L78 — `> clear-base.txt`가 워크트리 루트에 미추적 파일을 남긴다. — Severity: minor — Class: optional — Required fix: 출력 경로를 scratch 디렉터리나 `$R/`로 지정한다.

N13. SELF-REPORTED-BASELINES — acceptance.md:L51, L76 — `repo_head_before/after`와 `always_loaded_before`는 `$E`에 적힌 자기 보고값이다. 격리 검사는 HEAD와 브랜치만 보고, 작업 트리 상태는 보지 않는다. 상위 세션에서 상속된 `CLAUDE_PROJECT_DIR`를 사용자 수준 훅이 쓰면 저장소 쪽에 파일이 생길 수 있는데, 이것도 보지 않는다. — Severity: minor — Class: optional — Required fix: 프로브 전후 `git status --porcelain` 동일성을 증거에 넣고, 프로브 줄의 `unset`에 `CLAUDE_PROJECT_DIR`를 더할지 판단해 적는다.

N14. CLAUDE-LOCAL-GROWTH — design.md:L15; spec.md:L118 — CLAUDE.local.md는 예산 표면 밖이다(`token_budget_guard.go:65` 주석). 그런데 verdict §① Claim 2·4에 따르면 모든 L1 워크트리 세션에서 두 번 로드된다. 이 SPEC이 거기에 문장을 더하는 비용은 어떤 요구로도 제한되지 않는다. — Severity: minor — Class: optional — Required fix: CLAUDE.local.md 추가분에도 순증 0 규칙을 적용할지 한 줄로 결정해 적는다.

N15. PHRASE-GATING-LEXICAL — acceptance.md:L70-L71 — AC-013은 정식 문구 고정 문자열만 본다. 바꿔 쓴 미측정 주장(예: 「the move doubles the listing」)이 통과하고, `$CL`, `$LP`는 검사 대상이 아니다. AC-014는 「정확히 두 이동」을 검사하지 않는다. — Severity: minor — Class: optional — Required fix: 검사 대상에 `$CL`, `$LP`를 더한다. 금지어 목록(`listing` 근처의 `adds|doubles|single|one`)으로 보조 검사를 둔다.

## Regression Check (Iteration 2)
Defects from previous iteration:
- D1, D2, D3, D5, D7, D8, D9, D10, D11, D13–D22: RESOLVED — 근거는 §1 표.
- D4: UNRESOLVED(부분) — 파일 분리와 전사 cwd 검사는 들어갔으나 프로브 cwd를 적을 수 없다(N3).
- D6: UNRESOLVED(부분) — 네임스페이스 제외는 해결, 이름 기반 귀속이 새 결함(N2).
- D12: UNRESOLVED(부분) — 상설 세션 흐름은 정했으나 /clear 병합 규칙이 REQ-017과 충돌(N1).
- 세 회차 연속 무변 결함 없음(정체 신호 없음). 점수 0.55 → 0.74 상승.

## Recommendation

Verdict FAIL. must-pass 7개(N/A 1)는 모두 통과했다. 그러나 blocking 결함 8건(major 6: N1–N6, minor 2: N7–N8)이 있고, 집계 0.74가 Tier L 기준 0.85에 못 미친다. 이번이 3회 상한 중 2회차다. manager-spec 수정 순서는 다음과 같다.

1. N1: REQ-010과 REQ-017 중 어느 쪽을 개정할지 정하고, kanban-dispatch.md:153의 [HARD] 첫 줄에 대한 정확한 최종 문구와 대조 명령을 적는다.
2. N2: fixture에서 w1이 `fx-primary-*`를 갖지 않게 하는 요구와 검사를 추가한다. 판정 키를 목록별 트리 집합과 `listing_count`로 다시 정의한다.
3. N3: 프로브 줄 형식을 `unset … && cd -- <fixture 경로> && timeout … claude …`로 바꾸고, 후행 연결자를 거부하며, 경로 성분 단위로 포함 검사를 한다.
4. N4: AC-015를 쌍별 파일 비교로 다시 쓰고 process substitution 안의 git을 없앤다.
5. N5, N6: MUST-ExitWorktree 문장을 보존 목록에 넣고 상쇄 후보에서 뺀다. 절 범위 원문 대조로 바꾸고, AC-012의 제외 조건을 P-FLOW 문구 기준으로 바꾼다.
6. N7, N8: 상한 전제와 바이트 기준을 적고, AC-009에 재실행 대조를 넣는다. 산문 계산을 명령으로 바꾼다.
7. optional N9–N15는 오케스트레이터 재량이다. 목록이 길다는 이유로 FAIL을 만들지 않았다. 판정은 N1–N8과 점수에 근거한다.

3회차에서 FAIL이면 상한 규정에 따라 PASS-with-debt, 범위 축소, 명시적 연장 중 하나를 사용자에게 올려야 한다.

---
증거(비커밋, 세션 scratchpad `…/bb145fe7-…/scratchpad/`): `lint2.txt`(lint rc 0), `clear-base.txt`, `kt-mut.md`(AC-018 대조), `E.md`(합성 증거 — AC-002/003/006 실행과 경로 접두 뮤턴트), `cost2.py`(§3 재측정). 이 보고서는 커밋하지 않았다.
