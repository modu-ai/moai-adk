# SPEC Review Report: SPEC-SESSION-MIDMOVE-001
Iteration: 1/2 (Tier M ceiling)
Verdict: FAIL
Overall Score: 0.55

- 감사 대상: `.claude/worktrees/t1279` HEAD `a14fcf851` (브랜치 `WT-session-double-load`). BASE(`2370c5b31`)와 HEAD의 차이는 SPEC 산출물 4개뿐이다(`git diff --name-only 2370c5b31 HEAD`).
- 입력: spec.md, plan.md, acceptance.md, progress.md(Tier M), 전제 대조용 `.moai/reports/t1279/verdict.md` §①–§④.
- 작성자의 추론 맥락은 M1 Context Isolation에 따라 배제했다.
- 교차 모델: `mcp__moai__audit_multi`(project_root = 이 워크트리). codex는 도구 판정 필드가 `inconclusive`였지만 요약문은 FAIL이었고, 결함 7건을 냈다. 그중 D1(BASE 리터럴)은 이 감사가 놓친 것을 codex가 찾았고, 여기서 재현했다. GLM은 리뷰할 diff가 없어 `inconclusive`였다. 최종 판정은 이 감사자가 내린다.

## 전제 대조 (verdict.md §①–§④)

| 전제 | SPEC 근거 | 판정 |
|---|---|---|
| 로드 판정은 전사로만, 디버그 로그 불가 | spec.md:L58 REQ-SMM-003, L36(디버그 로그로만 본 항목을 명시적으로 불채택) | 일치 |
| A 보류 | spec.md:L120-123 (차단 3곳 인용 + 부활 조건 2개) | 일치 |
| B 범위 밖 | spec.md:L125-127 | 일치 |
| 범위는 D만 | spec.md:L29, L42, §C | 일치 — 단 §④ D 범위의 「스킬 목록 중복의 토큰 비용 측정」은 요구로 떨어지지 않았다(D8) |

spec.md:L31의 전사 관측도 재측정했다. `bb145fe7` 전사의 non-initial `skill_listing`(`2026-09-27T03:57:49.175Z`, skillCount 60)에서 `worktrees/t1219:`가 **84회** 나온다(python3 추출, `.../scratchpad/sl.py`). 인용된 바이트 수 46,541 B는 재현되지 않았다(D17).

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: REQ-SMM-001…016이 spec.md:L48–L88에 공백·중복 없이 3자리로 이어진다.
- [PASS] MP-2 EARS/GEARS 준수(요구 계층에서 판정): 16개 REQ 전부 패턴 라벨을 달았고 `shall` 구조다. 사소한 예외 두 건(REQ-005 L60의 규범문 `may`, REQ-016 L92의 docs-only 문장이 shall 형식이 아님)은 D19로 분리했다. AC 계층의 Given-When-Then은 이 기준에서 평가하지 않았다.
- [PASS] MP-3 YAML frontmatter validity: spec.md:L2–L13에 12개 필드가 정규 이름으로 있다(`status: draft`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 문자열, 날짜 ISO). `tier: M`은 L14. `moai spec lint` 경고는 0건이다.
- [N/A] MP-4 언어 중립성: 다중 프로그래밍 언어 도구를 다루는 SPEC이 아니다.
- [PASS] MP-5 D7: 참조하는 SPEC은 `SPEC-SESSION-DOUBLELOAD-001` 하나(`status: draft`)이고, retired·superseded·archived인 참조는 없다.
- [PASS] MP-6 D8: spec/plan/acceptance의 `syscall` 출현 수는 0, 0, 0이다.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION'` rc=1(일치 없음). research.md는 Tier M이라 없다.

**spec lint rc**: `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` → **rc=0**, `0 error(s), 0 warning(s)`. INFO 1건: `OwnershipTransitionUnmeasured`(커밋 `a14fcf851`에 `Authored-By-Agent` 트레일러가 없음).

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.55 | 0.50 | 「tree」 정의(L55), 바이트 기준(L56), warn의 면제 판별(L88), 「model-visible」 필드, 「near」·「their paragraph」(AC-011) 모두 해석이 필요하다 |
| Completeness | 0.70 | 0.75 | 절 구성은 모두 있다(HISTORY L20, §A–§F, `### Out of Scope —` L120 외). 그러나 §B의 「비용 측정」이 요구가 되지 않았고(D8), Kanban 역할 세션 구조가 빠졌다(D12) |
| Testability | 0.40 | 0.50 | `BASE` 리터럴로 AC 4건이 공허 통과 또는 영구 적색(D1). 자리표시자가 여러 곳이고(D14), 양성 대조가 추출기를 거치지 않는다(D5) |
| Traceability | 0.65 | 0.50–0.75 | 표는 16↔16 대응(acceptance.md:L73–L90)이지만 REQ-002 바이트, REQ-004 OR 분기, REQ-015 CLAUDE.local.md 절반이 AC로 덮이지 않는다 |

집계는 조화평균 0.55다. Tier M PASS 기준 0.80(spec-workflow.md § SPEC Complexity Tier)에 못 미친다.

## Defects Found (structured defect-list)

D1. AC-BASE-LITERAL — acceptance.md:L34, L39, L46, L47, L56 — 변수 블록(L20)은 `BASE=2370c5b31`로 정의하는데, 명령은 `$BASE`가 아니라 맨 `BASE..HEAD` / `git show BASE:"$KT"`를 쓴다. `BASE`라는 ref는 존재하지 않는다. 재현: `git diff --name-only BASE..HEAD -- internal/hook/` → `fatal: bad revision 'BASE..HEAD'`, rc=128, stdout 비어 있음. 결과는 다음과 같다. ① AC-016 docs-only(「prints nothing」)는 무엇을 바꿔도 **공허 통과**한다. ② AC-006의 `git log … | grep -cE '\.jsonl$|debug'`는 `0`을 찍어 **공허 통과**한다. ③ AC-013은 자기 대조(추가 줄 수 ≥ 1)가 `0`이 되어 **영구 적색**이다. ④ AC-008의 「BASE 시점 문장 보존」은 기준 파일을 읽지 못한다. 최초 발견은 codex이고, 여기서 재현했다. — Severity: critical — Class: blocking — Required fix: 모든 범위를 `"$BASE"`(D2의 `$RUN_BASE`)로 쓴다. 파이프라인 앞단 git의 exit code를 따로 확인하는 형태(예: 범위를 먼저 `git rev-parse --verify "$RUN_BASE^{commit}"`로 검증)를 AC에 넣는다.

D2. AC-RANGE-VS-DP3 — acceptance.md:L20, L34, L46, L47, L56; spec.md:L172 — `$BASE`로 고쳐도 문제가 남는다. 권장 DP-3(wait)은 develop과 t1175를 이 브랜치에 흡수하므로, 리터럴 plan 시점 SHA에서 시작하는 범위에는 다른 카드의 커밋이 섞인다. 실측으로 확인했다. t1175 tip `3a48485af` 기준으로 `AGENTS.md` §3은 바뀌지만(`git grep`: WT-rules-diet:AGENTS.md:112, :128) `AGENTS.md.tmpl` §3은 옛 문구 그대로다(:126, :142). merge-base `a0b78213d` 기준 t1175는 tmpl을 건드리지 않는다. 그러면 AC-012의 §3 동일성 diff는 **이 카드와 무관한 이유로 적색**이 된다. AC-016 docs-only는 develop이 싣고 오는 `internal/hook/` 변경을 이 카드의 것으로 세고, AC-013은 t1175가 추가한 줄을 스캔한다. 레인 프로토콜 §8 [HARD](「흡수한 ref와의 merge-base부터 잰다 — 리터럴 base SHA로 재지 않는다」) 위반이다. — Severity: major — Class: blocking — Required fix: 흡수 시점에 `RUN_BASE=$(git merge-base develop HEAD)`를 증거 파일에 기록하고, 범위 판정은 `"$RUN_BASE"..HEAD`(커밋 대응은 `--no-merges`)로 한다. 대조군 `git diff --name-only "$RUN_BASE"..HEAD | wc -l` ≥ 1을 둔다. M2 G1에 「흡수 후 `AGENTS.md` §3 ↔ `AGENTS.md.tmpl` §3 동일성을 다시 재고, 어긋나면 이 카드 범위 밖의 선행 결함으로 리드에게 보고한다」를 추가한다. 흡수 후 §D 줄 번호와 [HARD] 기준값도 다시 측정한다(t1175는 `kanban-dispatch-mechanics.md`를 새로 만든다).

D3. BUDGET-HEADROOM — spec.md:L81; acceptance.md:L46; plan.md:L120 — `go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v` → `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)`, `--- PASS:` 2건. 여유는 **70 토큰**이다. 그런데 REQ-007/008/010/011은 항상 로드되는 `kanban-dispatch.md`(파일 머리말이 「Intentionally always-loaded」라고 밝힌다)와 CLAUDE.md가 import하는 `AGENTS.md`에 문장을 더하라고 요구한다. plan §F는 「a few sentences」라고만 하고 실측 여유도 상쇄 규칙도 없다. 게다가 AC-012는 상수 `AlwaysLoadedTokenBudget`를 올려도 통과한다(codex #3). — Severity: major — Class: blocking — Required fix: 실측 여유(흡수 후 재측정값)를 SPEC에 기록한다. 추가하는 always-loaded 문장은 같은 파일 안에서 순증 0으로 상쇄하거나, 경로 한정 detail companion에 두고 포인터만 남기도록 요구한다. AC-012에 「`AlwaysLoadedTokenBudget` 상수 불변」(`git diff "$RUN_BASE"..HEAD -- internal/config/` 비어 있음 + 대조군)을 추가한다.

D4. CMDS-FORMAT — acceptance.md:L32, L33, L34; spec.md:L58, L61–L66; plan.md:L34–L38 — 모순이 세 겹이다. ① AC-004는 「`$CMDS`에 기록된 추출 명령」을 다시 돌리는데, AC-006은 `$CMDS`의 주석 아닌 모든 줄이 `unset MOAI_KANBAN … && timeout -k 10 300 claude `로 시작해야 한다고 요구한다. python3 추출 줄(과 fixture 구성 줄)은 둘 중 하나를 어긴다. ② P1과 P2는 시작 cwd만 다른데, 그 접두 형식으로는 cwd를 적을 수 없다. 그래서 「repository root path를 포함하지 않음」 검사는 저장소 루트에서 실행해도 통과한다(뮤턴트가 성립한다). REQ-006 첫 항목(저장소에서 시작하지 않음)을 검증할 수 없다. ③ operator-run은 `# operator-run:` 주석이라 AC-006 검사 대상에서 빠지고, env 정리·timeout·cwd 제약 없이 돌아도 통과한다(codex #6). — Severity: major — Class: blocking — Required fix: 파일을 나눈다(`probes.txt`: 프로브 줄만, `extract.txt`: 추출 명령, `fixture.txt`: 구성). 각 프로브 세션에 대해 **전사의 `cwd` 필드**를 추출해 `probe_cwd_<P>:`로 기록하고, 그 값이 scratch fixture 아래라는 것을 AC로 검사한다. 양성 대조로 저장소 경로를 넣은 가짜 행이 적발되는지도 본다. operator-run도 같은 전사 `cwd` 검사와 세션 id 기록을 거친다.

D5. CONTROL-NOT-EXERCISING-EXTRACTOR — plan.md:L27, L44; acceptance.md:L32; spec.md:L59 — plan은 `m1-control.txt`에 이미 추출한 필드(타임스탬프, isInitial, skillCount, 바이트, 접두 집합)만 담는다. 이 요약본에 전사용 추출기를 돌리면 ⓐ 형식이 달라 파싱하지 못하거나 ⓑ 이미 추출된 접두만 grep한다. 어느 쪽이든 「실제 `skill_listing` 모양을 못 읽는 눈먼 추출기」를 적발하지 못한다. 또 AC-004의 발동 조건은 「tree 하나」뿐이다. REQ-004의 OR 분기(「worktree 범위 목록 없음」)가 빠졌다(codex #4). — Severity: major — Class: blocking — Required fix: 대조 파일을 실제 전사와 같은 JSONL 모양으로 만들되, 필드를 `{type, isInitial, skillCount, names}`로 줄인다(스킬 식별자일 뿐 대화 내용이 아니다). 같은 추출기를 그대로 돌린다. 음성 대조(worktree 범위 이름이 없는 JSONL → 0)를 함께 둔다. AC-004의 발동 조건을 REQ-004의 OR 조건과 같게 맞춘다.

D6. TREE-DEFINITION — spec.md:L55; plan.md:L41 — 「a tree is named by its scope prefix, or by the absence of one for the start tree」라는 정의는 실제 목록에서 모호하다. 이 세션의 목록만 봐도 `vercel:`, `anthropic-skills:`, `typesafe:`(플러그인 네임스페이스), `moai:`, `harness:`(명령 네임스페이스) 같은 트리가 아닌 접두가 있다. 프로브는 사용자 `CLAUDE_CONFIG_DIR`를 그대로 쓰므로 사용자·플러그인 스킬도 목록에 들어온다. 순진한 접두 집합은 P1에서 여러 「tree」를 세게 되고, `launcher_single_listing`(REQ-010의 문구 조건)이 거짓 `no`가 된다. — Severity: major — Class: blocking — Required fix: fixture가 이미 준 판별자를 쓴다. 트리 귀속은 스킬 이름(`fx-primary-*` / `fx-wt-*`)과 `.claude/worktrees/<x>:` 형태의 접두로 정하고, 그 밖의 네임스페이스는 제외 집합으로 명시한다. `launcher_single_listing`·`clear_restores_single_listing`의 「one tree」를 「fixture 트리 중 하나만」으로 정의한다.

D7. MEASURE-SCHEMA — spec.md:L56; plan.md:L45–L56; acceptance.md:L30, L61 — REQ-002는 「각 `skill_listing`의 바이트 크기」를 요구하지만 plan의 `$E` 키 목록에는 바이트 키가 없고, AC-002도 검사하지 않는다. AC-002는 키 **이름**의 개수만 센다. 값이 `garbage`거나 비어 있어도 9와 2를 찍는다(codex가 뮤턴트로 재현). acceptance §D.1이 요구하는 P2/P3 세션 id 기록도 AC가 없다. `clear_restores_single_listing: yes` 같은 값은 REQ-010을 거쳐 템플릿 문구를 결정하므로, 검증되지 않은 값이 독트린으로 흘러간다. — Severity: major — Class: blocking — Required fix: `listing_bytes_P1..P3`, `session_id_P2`, `session_id_P3`(서로 달라야 함) 키를 추가한다. 값 형식을 정규식으로 검사한다(예: `^turn_tokens_P[123]: ([0-9]+(,[0-9]+)*|gap)$`). `launcher_single_listing`이 `listing_trees_P1`에서 기계적으로 도출되는지 대조하는 명령을 AC에 넣는다.

D8. COST-NOT-REQUIRED — spec.md:L42, L145; plan.md:L51; verdict.md §④ — 리드의 D 범위는 「스킬 목록 중복의 토큰 비용 측정을 포함한다」이고, §B도 「cost … is measured, not assumed」라고 약속한다. 그런데 비용 수치를 내는 REQ나 AC가 없다. `skill_dup_tokens_P2`는 plan에만 있고 「reference only」다. §E L145는 「no requirement … depends on them」이라고 말한다. fixture는 3+3 스킬인데, 실측 목록은 60 스킬·46,307 B(첨부 JSON)이므로 fixture의 차이값은 그 비용이 아니다. — Severity: major — Class: blocking — Required fix: 비용 키 하나를 REQ와 AC로 올린다. 방법은 ⓐ 대표 규모 fixture(실제와 같은 스킬 수)의 이동 전후 입력 합계 차이, 또는 ⓑ 이미 있는 실제 전사의 `usage` 필드로 03:57:49Z 첨부 전후 차이 중 하나다(전사만 쓰므로 REQ-003과 일관된다). 캐시 효과로 해석이 흐려지면 `gap` 사유를 기록하게 한다.

D9. REQ015-VS-PLAN — spec.md:L84; plan.md:L80; acceptance.md:L38, L49 — REQ-015는 lane protocol과 **CLAUDE.local.md §4.1**이 「같은 진입 형식과 순서」를 말하라고 요구한다. 그런데 plan M3.6은 CLAUDE.local.md를 「line 340 (DP-2 only)」로 한정해 진입 형식을 넣지 않는다. 요구와 계획이 모순이다. AC-015는 CLAUDE.local.md를 전혀 검사하지 않는다. 게다가 `moai cc -w <card-id>`는 BASE 시점에 이미 lane protocol에서 1회 나온다(실측). 그래서 AC-007의 LP 부분과 AC-015 앞 절반은 작업 전부터 녹색(공허)이다. — Severity: major — Class: blocking — Required fix: REQ-015를 plan에 맞게 줄이거나 plan에 CLAUDE.local.md §4.1 진입 문장을 추가한다. 그 문장의 고정 문자열 검사를 AC-015에 넣는다(BASE 시점 0 → 이후 ≥ 1). LP 검사는 BASE에서 적색인 것(예: `/clear` 0 → ≥ 1, 이미 있음)만 남긴다.

D10. HARD-PRESERVATION — acceptance.md:L39, L48; spec.md:L83 — 「`/clear` handoff between phases」의 [HARD] 보존은 AC-008의 「BASE 시점의 모든 문장이 그대로 남는다」에 기대고 있다. 이 절에는 명령이 없다(게다가 D1 때문에 BASE를 읽지도 못한다). AC-014는 제목 문자열과 [HARD] 개수의 하한만 본다. 제목과 개수를 유지하면서 L153의 [HARD] 문단을 약화시키는 뮤턴트가 통과한다. — Severity: major — Class: blocking — Required fix: 명령으로 명시한다. 예: `git show "$BASE":"$KT"`에서 해당 절의 비지 않은 줄을 뽑아 현재 파일에 `grep -qF --`로 하나씩 대조하고 누락 0을 기대한다. 양성 대조로 한 줄을 바꾼 사본에서 누락 1이 나오는지 본다. 같은 대조를 L179(새 카드)와 L181(WT- 개명)의 [HARD] 문단에도 건다. L181은 「wording only if needed」(spec.md:L104)이므로, 허용되는 변경 범위를 문장 단위로 지정한다.

D11. DP1-ASYMMETRY — spec.md:L153–L159, L88–L90; plan.md:L93; `internal/hook/post_tool_worktree.go`:L47–L54 — Block을 반대하는 근거(「면제를 경로 목록에 기대면 정확성이 명명 규칙을 따라가야 한다」, 「non-card sessions unless each is exempted」)는 Warn에도 똑같이 적용된다. REQ-016 warn은 「카드 워크트리이고 면제 통합 워크트리가 아닌」 이동에만 발화해야 하고, AC-016 warn에는 면제 음성 사례가 있다. 그런데 plan M4a는 「no config flag needed」라면서 면제를 어떻게 판별하는지 말하지 않는다. Warn 행의 비용 칸도 「already emits a message」라고 하지만, 기존 핸들러가 내는 것은 `SystemMessage`다. REQ-016이 요구하는 「세션 모델에게 보이는」 통지가 그 필드로 충족되는지는 SPEC 안에서 확인된 적이 없다. 배포 템플릿에서는 Kanban 모드가 아닌 사용자의 `.claude/worktrees/*` 이동(`claude -w`, isolation 서브에이전트)에도 「리드가 포인터를 다시 보낸다」는 통지가 뜬다. — Severity: major — Class: blocking — Required fix: DP-1 Warn 행에 면제 판별 비용·위험과 Kanban 모드 한정 여부를 추가한다. 통지를 실을 출력 필드를 이름으로 명시하고, 그 필드가 모델에게 보이는지를 어떻게 측정할지 적는다. 이렇게 두 행의 비교 기준을 맞춘 뒤 권고를 다시 쓴다.

D12. KANBAN-TOPOLOGY — spec.md:L70, L73; kanban-dispatch.md:L13(역할별 companion은 「launched by hand, one per terminal」), L266(Factory 레인 `worker-N`이 카드를 차례로 운반) — Kanban companion(plan/run/sync) 세션은 여러 카드를 서로 다른 워크트리에서 처리하는 상설 역할 세션이다. Factory 레인도 이름이 고정된 장수 세션이다. REQ-007은 카드 작업의 진입 형식을 런처 시작으로 정하고, REQ-009는 세션 내 `EnterWorktree`를 동등한 대안으로 제시하는 것을 금지한다. 그러나 카드마다 역할·레인 세션을 누가 어떻게 다시 띄우는지, 다시 띄운 세션이 디스패치 주소(`worker-N`, 역할 이름)를 유지하는지, 운영자 부담이 얼마인지 어디에도 없다. 결과적으로 이 구조에서는 「불가피한 이동」 분기가 정상 경로가 된다. 기존 「/clear handoff」(/clear → 재전송)와 새 순서(이동 → /clear → 재전송)가 이어지면 카드당 /clear가 두 번 필요한지도 정리되지 않았다. — Severity: major — Class: blocking — Required fix: 역할·레인 세션의 카드별 재기동 절차를 요구로 적거나(주소 보존 방법 포함), 이 판단을 DP-4로 올려 선택지와 비용을 제시한다. 기존 절과 새 절의 /clear 순서가 합쳐지면 어떻게 되는지 한 문장으로 확정한다.

D13. AC001-TIMEZONE — acceptance.md:L29 — `git log -1 --format=%cI`는 로컬 오프셋(+09:00)을 붙이고, 전사 타임스탬프는 `Z`(UTC)다(실측 전사 줄 모두 `Z`). 문자열로 비교하면 오판한다. — Severity: minor — Class: blocking — Required fix: 두 값을 epoch 초로 바꿔 비교한다(`git log -1 --format=%ct`와 전사 시각의 UTC epoch).

D14. PLACEHOLDERS — acceptance.md:L29(`<that commit>`), L33(「every key whose value is gap is listed」에 명령 없음), L41(`<section from AC-SMM-008>`), L42(「their paragraph」, 「near」), L46(커밋 대응 루프에 명령 없음, §3 추출이 양쪽 다 비면 diff가 0으로 통과 — codex #5), L54(페이로드 자리표시자, 「model-visible context field」 이름 없음) — 「Every criterion is … binary」(L3)라는 선언과 맞지 않는다. — Severity: minor — Class: blocking — Required fix: 각 자리를 완전한 명령과 기대 stdout·exit code로 바꾼다. §3 비교 앞에는 추출 줄 수 > 0 가드를 둔다(BASE 시점 36줄 실측).

D15. AC003-LEXICAL — acceptance.md:L31 — `grep -ciE 'debug-file|--debug|/debug/'`는 `claude -d`를 놓친다. `claude --help`의 78행이 `-d, --debug [filter]`다. — Severity: minor — Class: optional — Required fix: `(^| )-d( |$)`를 패턴에 추가하거나, 금지 플래그를 명령 파싱으로 검사한다.

D16. AC013-CONTROL — acceptance.md:L47 — 양성 대조가 네 대안을 모두 담은 한 줄이다. 대안 셋이 깨져도 `1`을 찍는다. 이 환경의 `grep`은 ugrep 7.8.4로 해석되며, 여기서는 대안 넷이 각각 적중했다(개별 실측). 다른 grep 구현에서 `\b` 의미가 달라져도 대조가 잡지 못한다. — Severity: minor — Class: optional — Required fix: 대안마다 한 줄씩, 네 줄짜리 대조로 `4`를 기대한다.

D17. BYTE-BASIS — spec.md:L31, L56 — 「46,541 B」는 재현되지 않는다. 같은 첨부가 첨부 JSON으로는 46,307 B, JSONL 줄 전체로는 90,584 B다. REQ-002의 「byte size」는 기준이 정의되어 있지 않다. — Severity: minor — Class: blocking — Required fix: 바이트 기준(예: `content` 필드의 UTF-8 길이)을 한 가지로 정의하고, §A 수치를 그 기준으로 다시 잰다.

D18. TEMPLATE-CLAIM-UNGATED — spec.md:L109; plan.md:L77 — worktree-integration 템플릿에 들어갈 「mid-session move carries the start tree's skill listing alongside」는 Claude Code 런타임 동작에 대한 주장이다. REQ-010은 런처와 /clear에 관한 문구만 측정에 묶고, 이 문구는 P2 결과에 묶지 않는다(이 저장소 전사 증거는 있다). — Severity: minor — Class: optional — Required fix: REQ-010에 「P2가 두 트리를 기록했을 때만」이라는 조건을 추가한다.

D19. REQ-WORDING — spec.md:L60, L92 — REQ-005의 「may substitute」는 규범문 안의 `may`이고, REQ-016의 docs-only 문장은 shall 형식이 아니다. — Severity: minor — Class: optional — Required fix: 「the measurement shall either … or …」로 쓰고, docs-only 문장은 「the hook layer shall not change」로 쓴다.

D20. TIER-CEILING — spec.md:L14; plan.md:L8 — REQ 16개, AC 16개로 Tier M 상한(각 16)에 정확히 닿아 있다. D8·D12를 고치면 REQ가 늘어 Tier L이 된다. 영향 파일도 문서 10 + Go 2 + 증거·진행 4로 경계(15)에 걸린다. 수정 범위인 `AGENTS.md`는 모든 하네스의 상설 계약이다. block 선택 시 Tier L로 재분류하면 PASS 기준이 0.85가 되는데, 재감사가 필요하다는 말이 없다. — Severity: minor — Class: optional — Required fix: 수정 후 REQ/AC 수로 Tier를 다시 판정하고, 재분류 시 plan-audit을 다시 받는다고 명시한다.

D21. CAPS-ACCOUNTING — spec.md:L48; acceptance.md:L61 — 「6 probe sessions」에서 `--resume` 호출과 `/clear` 뒤의 새 세션 id가 각각 세션으로 세어지는지 정의가 없다. `wall_clock_min`은 출처 없는 자기 보고다. — Severity: minor — Class: optional — Required fix: 세션 수는 「고유 전사 session id 수」로, 벽시계는 「첫 프로브 전사 첫 줄 ~ 마지막 프로브 전사 마지막 줄」로 정의한다.

D22. DP2-COST-OMITTED — spec.md:L165 — 「Exempt (recommended)」 행은 이점만 적는다. 통합 워크트리로 들어갔다 돌아올 때마다 그 트리 범위의 스킬 목록이 추가되고 세션이 끝날 때까지 남는다는 비용(D 범위의 바로 그 문제)이 빠졌다. — Severity: minor — Class: optional — Required fix: Exempt 행에 이 비용을 한 줄로 적는다.

## Regression Check (Iteration 2+ only)
해당 없음(1회차).

## 결정 지점 평가 (DP-1 / DP-2 / DP-3)

- **DP-1**: 선택지 셋과 비용·위험 표는 갖췄다. 그러나 D11의 비대칭(면제 판별 부담을 Block에만 귀속시킴, 모델 가시성 미확인) 때문에 **중립적 제시로 인정하지 않는다.** Tier 판단(block → L)은 plan.md:L8·L16과 AC-016 block 분기(L55)가 서로 맞물려 있어 형식은 온전하다.
- **DP-2**: 두 분기가 REQ-011과 AC-011에 모두 대응되어 있고 형식은 온전하다. Exempt 행이 비용을 빠뜨렸다(D22, 선택 사항).
- **DP-3**: 사실관계는 실측과 일치한다. `git merge-base --is-ancestor WT-rules-diet HEAD` rc=1, tip `3a48485af`, merge-base `a0b78213d` 기준으로 kanban-dispatch 78줄, AGENTS.md 76줄이 바뀐다. 그러나 권장안(wait)이 AC 범위(D1·D2)와 §3 동일성(D2)에 주는 영향이 SPEC에 반영되지 않았다.
- 세 DP 모두 `[NEEDS CLARIFICATION]` 표식이 아니라 오케스트레이터에 넘긴 결정으로 제시되어 있어, MP-7에는 해당하지 않는다(spec.md:L149).

## 통과 확인 항목 (근거)

- 새 린트 게이트: `-run` 패턴은 두 곳(acceptance.md:L46, plan.md:L99)뿐이고, 둘 다 대안마다 `^…$`로 고정되어 있다(`^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$`). 두 테스트는 `internal/config/token_budget_guard_test.go:54, :84`에 있다. `$`로 고정한 덕분에 `TestAlwaysLoadedTokenBudget_OverBudgetFails`(:124)는 빠진다. `--- PASS:` 주장은 AC-012 한 곳이고 `--- PASS: TestX (` 형태로 Go 구분자를 갖췄다. 실행 결과 두 줄 모두 관측했다.
- 측정 상한 선커밋: REQ-001(L48)과 AC-001(L29)이 「프로브 출력이 없는 단독 커밋」과 「첫 프로브보다 이른 커밋 시각」을 요구한다. 시각 비교의 오류만 D13에 있다.
- 프로브 격리: scratch fixture 한정(plan.md:L23, L106), `--setting-sources` 금지(spec.md:L64, AC-006), `MOAI_KANBAN*` 해제(L65). scratchpad의 `git -C <fixture> init`은 이 워크트리 세션 가드를 통과한다(rc=0 실측). 따라서 fixture 구성 자체는 실행 가능하다.
- Template-First 결합: 세 쌍 모두 `cmp` rc=0이다. AGENTS.md §3 ↔ tmpl §3 `diff` rc=0(36줄). 템플릿 미러가 없어야 할 로컬 전용 파일은 `test -e internal/template/templates/.claude/rules/local` rc=1, 템플릿 내 `lane-protocol` 언급 0건이다(대조: 로컬 디렉터리 1건).
- 기준값: §D.L114의 수치는 모두 실측과 일치한다(`EnterWorktree(<card-id>)` 2/2, `wt: EnterWorktree(t0)` 1, [HARD] 38/8/18/21/48/0). AC-014의 다섯 문자열도 현재 모두 1회씩 있다.
- 템플릿 중립성: REQ-013·AC-013의 금지 토큰(카드 id, SPEC id, 날짜, 해시)이 올바르게 설정되어 있다. 대안 넷 모두 개별 적중했고, `t0`(한 자리)는 적중하지 않았다.

## Recommendation

Verdict FAIL. must-pass는 모두 통과했지만, blocking 결함 15건(critical 1, major 11, minor 3)이 있고 집계 0.55가 Tier M 기준 0.80에 못 미친다. manager-spec 수정 순서는 다음과 같다.

1. D1 → D2: 모든 git 범위를 `$RUN_BASE`(흡수 시점 merge-base)로 바꾸고, git exit code 확인과 대조군을 AC에 넣는다. M2 G1에 AGENTS.md §3 ↔ tmpl §3 재측정을 추가한다.
2. D3: always-loaded 여유 70 토큰을 SPEC에 기록하고, 순증 0 또는 detail companion 배치를 요구한다. 예산 상수 불변도 AC에 넣는다.
3. D4–D7: 측정 산출물을 설계를 다시 한다 — 명령 파일 분리, 전사 `cwd`로 격리 증명, JSONL 모양의 양성·음성 대조, fixture 스킬 이름 기반 트리 정의, 바이트·세션 id·값 형식 키.
4. D8: 리드 §④의 비용 측정을 REQ/AC로 올린다(대표 규모 또는 실제 전사의 usage).
5. D9, D10: REQ-015와 plan M3.6의 모순을 해소하고, [HARD] 문단 보존을 줄 단위 대조 명령과 양성 대조로 기계화한다.
6. D11, D12: DP-1 표를 대칭으로 다시 쓰고(면제 판별, Kanban 한정, 출력 필드), Kanban·Factory 역할 세션의 카드별 재기동을 요구나 DP-4로 다룬다.
7. D13, D14, D17: 시간대, 자리표시자, 바이트 기준을 정리한다.
8. 수정 뒤 REQ/AC 수로 Tier를 다시 판정한다(D20). optional 항목(D15, D16, D18, D19, D21, D22)은 오케스트레이터 재량이다.

---
증거 경로(비커밋, 세션 scratchpad): `.../scratchpad/lint.txt`, `.../scratchpad/budget.txt`, `.../scratchpad/sl.py`. 이 보고서는 커밋하지 않았다.
