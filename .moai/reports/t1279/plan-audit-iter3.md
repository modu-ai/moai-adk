auditor-model: claude-opus-5-5[1m]

# SPEC Review Report: SPEC-SESSION-MIDMOVE-001
Iteration: 3/3 (Tier L 상한, 마지막 회차)
Verdict: FAIL
Overall Score: 0.83

- 감사 대상: `.claude/worktrees/t1279` HEAD `8122d749d`(브랜치 `WT-session-double-load`). `git merge-base develop HEAD` = `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`(= 로컬 develop tip).
- 입력: Tier L 산출물 전부(spec.md, plan.md, acceptance.md, design.md, research.md, progress.md). 전제·결정은 `.moai/reports/t1279/verdict.md` §①–§⑤, 이전 회차는 `plan-audit.md`, `plan-audit-iter2.md`에서 읽었다.
- 작성자의 해결 주장과 research.md §R11 원장은 M1 Context Isolation에 따라 입력으로 쓰지 않았다. 대조용으로만 쓰고, 모든 AC는 직접 다시 실행했다. heredoc 본문은 acceptance.md에서 기계적으로 추출해 그대로 실행했다(scratchpad `i3/acx.py`).
- 교차 모델: `mcp__moai__audit_multi`(project_root = 이 워크트리). codex는 도구 판정 필드가 `inconclusive`지만 요약은 FAIL이고 결함 8건을 냈다. 그중 5건(AC-013 인덱스, AC-022 출처, P3 연결, `$E` 교차 검증, 프로브 플래그)은 이 감사에서 재현하거나 코드로 확인해 아래 결함에 넣었다. GLM은 응답 본문이 없어 `inconclusive`였다. 판정은 이 감사자가 내린다.
- 도구 경고: 실행 중인 moai MCP 바이너리는 `a8a9b9376`(HEAD의 조상)에서 빌드됐다. spec lint는 `go run`으로 현재 트리에서 따로 돌렸다.

## 1. N1–N15 해결 여부

| 결함 | 판정 | 근거(이번 회차 직접 실행) |
|---|---|---|
| N1 /clear 순서 ↔ 원문 동결 | **해결** | design.md §0의 개정문은 옛 줄의 문장을 모두 그대로 담고 한 문장만 끼워 넣는다. 끼운 문장은 「On a card change … after the move … a phase end that is not a card change is cleared as before」이다. REQ-010(spec.md:L118)이 바꿀 줄을 지목하고, REQ-017(L138)이 그 한 줄만 예외로 둔다. AC-018 통제군: 치환본은 보존 `0`·절 안 새 줄 `1`·옛 줄 `0`, MUST 줄 이동본은 보존 `1`. 순서 게이트는 REQ-020/AC-022로 들어갔다. 잔여: ND2(게이트 입력의 출처), ND9(문구 명확성) |
| N2 트리 귀속 | **해결** | fixture가 구성상 서로소다(w1을 스킬 생성 전에 만들고 스킬은 미추적으로 둔다, spec.md:L62-66). 판정 키는 목록 수와 목록별 집합에서 도출한다(L81-86). 합성 `$E`: 정상 `0`, count 0 뮤턴트 `2`, 키 반전 `1`. 새 결함 ND3(귀속 규칙의 전제) |
| N3 프로브 cwd·후행 명령 | **해결** | `cd --` 형식, 금지 문자, `commonpath` 포함 검사, 첫 cwd 검사가 들어갔다. AC-007: 정상 `2 0`, `; rm -rf x` `2 1`, `fx-evil` `2 1`, `-d` `2 1`, 상태 digest 불일치 `2 1`. AC-003: `fx-evil` `2`, P2 첫 cwd 오류 `1`. 잔여: ND8 |
| N4 AC-015 가드 | **해결**(잔여 ND6) | 쌍별 `git log … > file` → `comm -23`. 리터럴 scratch 경로로 한 번에 실행되고 `0 0 0 0`, `$KT` 커밋 수 `0`, 외래 SHA 통제군 `1`이 나왔다. 다만 acceptance.md:L21의 `S=$(mktemp -d)`를 git과 같은 호출에 두면 가드가 거부한다(ND6) |
| N5 [HARD] 상쇄 약화 | **해결** | MUST-ExitWorktree 줄이 절 단위 보존 대상이 됐다(spec.md:L139). 상쇄 후보는 비-[HARD] 줄로 한정됐다(design.md:L41). 이동 통제군 `1` |
| N6 이동 계수기 | **해결** | 제외 조건이 P-FLOW 문구로 바뀌었다. 단위별 분류 결과, 기준 시점 new-card [HARD] 문단이 COUNT로 잡혔다. 합계 `6`(R7과 일치) |
| N7 상한 전제·바이트 기준 | **해결**(잔여 ND7) | 전제, `bound=confounded` 규칙, 개행 제외 기준(spec.md:L105-107), 재실행 `cmp`가 들어갔다. 전사 판독: compact 행 `0`, `local_command` 행은 11행 하나이고 `/clear`를 담지 않는다. R10과 일치 |
| N8 잔여 산문 | **해결** | iter-2가 지목한 네 곳(첫 프로브 시각, 벽시계·probes_run·operator id, AC-013 합산, WT- 줄 추출)이 모두 heredoc이 됐다. AC-006 통제군 `1`/`1`/`1`. 새로 드러난 산문은 ND10에 따로 적었다 |
| N9 턴 상한 | **해결** | 턴을 정의했고(spec.md:L59), `turn_tokens` 5개 뮤턴트에서 AC-002가 `1`을 냈다. 잔여: `--max-turns 40` 부분 문자열(ND8) |
| N10 LP 진입 회귀 | **해결** | AC-010에 `$LP`가 들어갔다(현재 `1`) |
| N11 AC-020 환경 | **해결** | warn의 unset이 같은 호출에 있고, block의 설정 키 기록 위치를 적었다. 잔여: 네 번째 명령 부재(ND10) |
| N12 산출 잔여물 | **해결** | 모든 출력이 `$S`로 간다(ND6과 별개) |
| N13 자기 보고 기준선 | **해결** | `repo_status_sha` 전후 비교와 `CLAUDE_PROJECT_DIR` unset이 들어갔다. 상태 불일치 통제군 `2 1` |
| N14 CLAUDE.local.md 증가 | **해결** | 600 B 상한(spec.md:L136). AC-016 통제군(before 63700) `1` |
| N15 문구 게이트 | **해결** | 추가된 줄 가운데 `listing`을 담은 모든 줄이 정식 문구를 써야 하고, `$LP`·`$CL`도 대상이다. 바꿔 쓴 문구 뮤턴트 `+the move doubles the skill listing` → `2`(기준 `1`) |

해결 15/15. 이전 회차 결함 가운데 남은 것은 없다. 대신 이번 수정에서 새 결함이 드러났다(§4). 세 회차 연속으로 그대로 남은 결함은 없다(정체 신호 없음).

## 2. N1 집중 검토

- **의무를 옮기되 약화하지 않는가**: 옛 줄의 네 요소가 개정문에 그대로 있다. `[HARD]` 표지, 「one card's context」 규칙, 「user-typed command」 규칙, 세 부분 메시지 구조다. 옛 줄의 모든 문장이 부분 문자열로 남는다(design.md:L18 vs L24 대조). 끼운 문장은 카드 전환의 `/clear` 시점만 이동 뒤로 옮기고, 재전송 주체를 리드로 묶는다. 횟수는 「exactly once」로 유지된다. **약화 없음으로 판정.** 단 두 가지 해석 여지가 남는다(ND9). 첫째, phase가 끝났는데 다음 카드가 아직 없을 때 어느 규칙이 적용되는지 정하지 않았다. 둘째, 이동과 `/clear` 사이에 카드 작업을 하지 않는다는 말이 개정문에 없다(design.md §2의 「work starts」는 [HARD] 줄 밖에 있다).
- **카드 전환에만 닿는가**: 「On a card change …; a phase end that is not a card change is cleared as before」. 비-[HARD] 문장 「Where the next phase reuses a just-cleared session …」은 상쇄 후보에서 명시적으로 빠졌다(design.md:L41, plan.md:L81). 다만 이 문장을 지키는 AC는 없다. 이 문장을 지운 뮤턴트가 AC-018을 통과했다(ND11).
- **REQ-010 ↔ REQ-017**: REQ-010이 줄머리 문자열로 한 줄을 지목하고, REQ-017이 그 한 줄만 예외로 둔다. 두 REQ는 같은 파일에서 동시에 참일 수 있다. AC-018 통제군이 이를 보인다(치환본: 보존 `0`, 절 안 새 줄 `1`, 파일 전체 새 줄 `1`, 옛 줄 `0`). 새 줄을 다른 절로 옮기면 절 안 `0`, 옛 줄을 함께 남기면 옛 줄 `1`, 새 줄을 두 번 넣으면 `2`가 되어 모두 걸린다.
- **실행 게이트(REQ-020/AC-022)**: 구조상 순서는 맞다. `FIRST_KT`는 범위 안에서 `$KT`를 건드린 가장 오래된 커밋이다. 흡수 전에 편집한 커밋은 흡수 뒤에도 범위에 남으므로 `is-ancestor "$T1175" "$FIRST_KT"`가 exit 1이 된다. 현재 RED-now는 `git merge-base --is-ancestor WT-rules-diet HEAD` → rc `1`로 재현했다. **그러나 `T1175`는 `$E`에 적힌 자기 보고값이다. `CARD_BASE` 같은 임의의 조상을 적으면 두 검사가 모두 공허하게 통과한다.** t1257의 absorbed/not-landed 분기 명령은 코드 블록 밖 산문에만 있다(ND2).
- **t1175 tip 이동**: tip이 `4989ea6b0`에서 `deab44714`로 다시 움직였다. `deab44714`에서도 옛 줄은 `1`/`1`, `[HARD]` `38`, `MUST \`ExitWorktree\`` `1`, `EnterWorktree(<card-id>)` `2`, 두 사본 `cmp` rc `0`이다. 새로 생긴 `kanban-dispatch-mechanics.md`에는 `EnterWorktree`와 `/clear`가 0건이다. design.md §0 초안은 지금도 그대로 적용된다. M3 step 4가 흡수 시점에 다시 확인한다.

## 3. §R11 대조 (가드 아래 재실행)

| AC | 이번 회차 출력 | R11 | 일치 |
|---|---|---|---|
| 021 | SHA rc `0`; `10` | 동일 | 예 |
| 022 | `is-ancestor WT-rules-diet HEAD` rc `1` | 동일 | 예 |
| 010 | KT 0, DT 0, WT 0, AT 0, AL 0, CL 0, LP 1 | 동일 | 예 |
| 011 | `0 0 0 0`; `0`; `0`; `0`; `0` | 동일 | 예 |
| 012 | `2`,`2`; `1`; `6` | 동일 | 예 |
| 013 | 합성 `$E`(clear gap) `1`; 바꿔 쓴 문구 `2` | `1`; P-CLEAR 통제군 `2` | 예, 단 **ND1**(argv 인덱스)은 R11에 없음 |
| 014 | `0`,`0`,`0` | 동일 | 예 |
| 015 | `cmp` ×3 무출력; 36/36; `diff` rc 0; `0 0 0 0`; `0`; 통제군 `1` | 동일 | 예. 단 리터럴 `S`일 때만 실행된다(ND6) |
| 016 | rc `0`, `77530 … headroom 70`, PASS 두 줄; 합성 `0`; 통제군 `1`/`1`; 상수 diff `0`; grep `1` | 동일 | 예 |
| 017 | `0`; `4`; `0` | 동일 | 예 |
| 018 | old/new 각 `1`줄; HEAD 보존 `0`·옛 줄 `1`/`1`·절 새 줄 `0`·파일 새 줄 `0`; 비-kanban `0`; 이동 `1`; 치환 `0`/`1`/`0` | 동일 | 예 |
| 019 | rc `1`; `0`; `1` | 동일 | 예 |
| 020 | `0` | 동일 | 예 |
| 001 | `0`; CAPS_CT 초과 `1` | 동일 | 예 |
| 002 | `0`; count 0 `2`; 키 반전 `1` | 동일 | 예 |
| 003 | `3`; `0`; `fx-evil` `2`; 첫 cwd `1` | 동일 | 예 |
| 006 | `0`; gap 사유 제거 `1`; 벽시계 `1` | 동일 | 예 |
| 007 | `2 0`; 후행 `2 1`; `fx-evil` `2 1`; `-d` `2 1`; 상태 `2 1` | 동일 | 예 |

R11의 수치는 모두 재현됐다. R11이 놓친 것은 이번 회차에 추가한 뮤턴트에서 드러났다. `--max-turns 40`과 `--dangerously-skip-permissions --add-dir /Users/goos`가 AC-007을 `2 0`으로 통과했다. P3 count가 gap인데 `turn_tokens_P3`가 숫자인 경우는 AC-002 `0`이었다. 비-[HARD] 재전송 문장을 삭제한 경우는 AC-018에서 보존 `0`이었다. 가드 거부 관측: `S=$(mktemp -d)`와 git을 한 호출에 두면 거부된다(「names git in a form too complex to verify」, 두 번 재현). 여러 줄의 평문 git 호출은 한 호출 안에서도 실행됐다.

## 4. Must-Pass Results
- [PASS] MP-1 REQ 번호: `REQ-SMM-001`…`020`이 각각 1회씩 있고 빈 번호가 없다(`uniq -c` → 20개 모두 1). 문서 순서는 `…018, 020, 019`이다(spec.md:L142-150). 공백도 중복도 아니므로 PASS지만 선택 결함으로 적는다(ND11). AC는 `AC-SMM-001`…`022`가 22개 고유값이고 중복 0이다.
- [PASS] MP-2 GEARS(요구 계층에서 판정): 20개 REQ 모두 패턴 라벨과 shall 구조가 있다. 예: REQ-010 While(spec.md:L112), REQ-013 Where(L126), REQ-020 Unwanted 「shall not begin before」(L143), REQ-019 Where와 「the hook layer shall not change」(L150-154). AC의 Given-When-Then은 이 기준으로 평가하지 않았다.
- [PASS] MP-3 frontmatter: spec.md:L2-13에 12개 정규 필드가 있다(`version: "0.3.0"`, `status: draft`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 문자열). `tier: L`은 L14에 있다. lint 경고는 0건이다.
- [N/A] MP-4 언어 중립성: 다중 프로그래밍 언어 도구를 다루지 않는다.
- [PASS] MP-5 D7: 참조 SPEC은 `SPEC-SESSION-DOUBLELOAD-001` 하나이고 `status: draft`다. BLOCKING 없음.
- [PASS] MP-6 D8: 다섯 산출물의 `syscall` 수는 모두 `0`이다.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → rc `1`(일치 없음).

**spec lint**: `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` → **rc=0**, `0 error(s), 0 warning(s)`. INFO 1건: `OwnershipTransitionUnmeasured`(커밋 `a14fcf851`에 `Authored-By-Agent` 트레일러가 없음).

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75–1.0 | 요구 문장은 정밀하다(spec.md:L57-L154). 다만 REQ-002의 「unprefixed name comes from the session's start directory」(L68)는 측정되지 않은 런타임 전제를 규칙으로 적었고(ND3), 개정문은 다음 카드가 없는 phase 종료를 정하지 않았다(ND9) |
| Completeness | 0.90 | 0.75–1.0 | 절, frontmatter, Out of Scope H3 6개(spec.md:L182-206), Tier L 산출물 6종이 모두 있다. AC-020의 warn/block 분기와 AC-022의 t1257 분기는 명령이 불완전하다(ND10, ND2) |
| Testability | 0.72 | 0.50–0.75 | 대부분의 AC가 가드 아래 실행되고 통제군이 발화한다(§3). 그러나 AC-013은 잘못된 파일을 읽는다(ND1). AC-022는 공허하게 통과할 수 있다(ND2). P3와 `/clear`의 연결(ND4), `$E`와 `extract.txt`의 대조(ND5)는 검사되지 않는다. `S=$(mktemp -d)` 전제부는 거부된다(ND6) |
| Traceability | 0.95 | 1.0 | 20 REQ ↔ 22 AC 표가 완전하다(acceptance.md:L359-380). REQ-013의 covered 분기는 `$KT`만 검사한다(ND10) |

조화평균 = 4 / (1/0.80 + 1/0.90 + 1/0.72 + 1/0.95) = **0.83**. Tier L 기준 0.85 미만이다. 0.55 → 0.74 → 0.83으로 올랐으므로 STOP(회귀) 신호는 없다.

## Defects Found (structured defect-list)

ND1. AC013-ARGV-INDEX — acceptance.md:L208, L215 — heredoc 인자 순서는 `"$E" "$S/added.txt" "$KT" "$DT" "$WT" "$AT" "$LP" "$CL"`이다. 따라서 `sys.argv[4]`=`$DT`, `sys.argv[6]`=`$AT`인데, P-RELAUNCH 검사가 `open(sys.argv[6])`, 즉 **AGENTS.md.tmpl**을 읽는다. AC 설명(L205 「not yet in `$DT`」)과 plan M4 step 2는 이 문구를 detail companion에 두라고 한다. 실측: P-RELAUNCH를 DT에만 넣은 사본 → `1`(FAIL), AT에만 넣은 사본 → `0`(PASS). 계획대로 구현하면 AC-013이 거짓 FAIL을 낸다. 그렇다고 통과시키려고 문구를 always-loaded 계약 파일에 넣으면 예산 규칙(REQ-015)과 부딪친다. codex도 독립적으로 같은 결함을 냈다. — Severity: major — Class: blocking — Required fix: `sys.argv[6]` → `sys.argv[4]`. 통제군 두 개를 추가한다(DT에만 있으면 `0`, AT에만 있으면 `1`).

ND2. T1175-PROVENANCE — spec.md:L143-146; acceptance.md:L24, L49-53 — `T1175`는 `$E`의 `t1175_absorbed:` 자기 보고값이다. 그 값이 실제 t1175 착지 커밋인지 확인하는 명령이 없다. 임의의 조상(예: 현재 `CARD_BASE`)을 적으면 두 `is-ancestor`가 모두 exit 0이 된다. 리드 결정 §⑤가 요구한 「t1175 흡수 뒤 치환」이 이 입력 하나에 달려 있다. 또 t1257의 `absorbed` 분기(`is-ancestor <t1257 sha> "$FIRST_KT"`)와 `not-landed` 분기(`grep -c '^t1257_notified: yes$'`)는 코드 블록 밖 산문에만 있다. — Severity: major — Class: blocking — Required fix: 코드 블록에 출처 검사를 넣는다. 예: `git log -1 --format=%B "$T1175" | grep -c 't1175'` ≥ 1(카드 id는 커밋 메시지의 필수 운반자다), `git merge-base --is-ancestor "$T1175" develop` exit 0. t1257 두 분기도 명령과 기대 출력으로 옮긴다.

ND3. ATTRIBUTION-PREMISE — spec.md:L68; design.md:L70 — REQ-002는 「접두 없는 이름은 세션 시작 디렉터리의 `.claude/skills`에서 온다」를 규칙으로 정한다. 이것은 Claude Code 동작에 대한 측정되지 않은 전제다. 이 감사 세션(cwd = t1279 워크트리)의 스킬 목록에는 접두 없는 `hns-lsel-applier`와 `.claude/worktrees/t1279:hns-lsel-applier`가 함께 있다. 전사 bb145fe7 404행의 접두 `.claude/worktrees/t1219:`도 primary 기준 상대 경로다. 즉 관측된 접두는 저장소 주 루트를 기준으로 한다. 이 관측은 가설의 근거일 뿐 P1 동작을 확정하지 않는다. P1(w1에서 시작)이 fixture 루트를 프로젝트 루트로 삼는다면, 접두 없는 `fx-primary-*`가 w1로 귀속되어 `attribution_conflict`가 나고 증거 전체가 무효가 된다. 이 경우를 gap으로 돌리는 경로가 없다. REQ-005는 이동·`/clear` 불가와 상한 도달만 다룬다. 따라서 AC-002를 통과할 수 없게 되어 M1이 멈춘다. 잘못된 교리 문구로 이어지지는 않는다(fail-closed). — Severity: major — Class: blocking — Required fix: REQ-002/REQ-005에 규칙 하나를 더한다. 「경로에 `attribution_conflict`가 나오면 그 경로의 모든 값을 `gap`(사유 `attribution_conflict`)으로 기록하고, AC-002의 conflict 계수는 non-gap 경로에만 적용한다」. 교리는 REQ-012의 no/gap 문구로 돌아간다. 가능하면 P1 결과로 접두 기준을 기록한다.

ND4. P3-CLEAR-LINK — spec.md:L71; acceptance.md:L92-105, L346-347 — P3가 정말 P2 세션의 `/clear` **뒤** 구간인지 검사하지 않는다. AC-003의 P3 항목에는 첫 cwd 검사가 없다. `session_id_P3`와 P2의 관계(같은 id이면 `/clear` 행 뒤만 센다, 새 id이면 `/clear`로 생긴 것이다)도 §D.1 산문에만 있다. 별개의 새 세션을 w1에서 띄워 P3로 적으면 `clear_restores_single_listing: yes`가 나온다. 그러면 P-CLEAR(「`/clear` leaves only the new tree's listing」)가 관측 없이 교리에 들어간다. REQ-012가 막으려던 바로 그 형태의 미관측 주장이다(codex 독립 지적). — Severity: major — Class: blocking — Required fix: `extract.txt`에 `/clear` 경계 행(시각·행 번호)을 출력하게 한다. AC-003(또는 AC-006)에 P3의 id 관계와 경계 뒤 목록만 셌음을 검사하는 heredoc을 넣는다. 이 검사가 들어오기 전까지 `clear_restores_single_listing`은 `gap`으로 취급한다.

ND5. E-NOT-CROSSCHECKED — acceptance.md:L68-90, L107-112 — AC-002는 `$E`의 내부 일관성만 본다. `$E`의 `listing_count/trees/bytes/turn_tokens`가 `extract.txt`의 해당 세션 `listing`·`turn` 줄과 같은지는 비교하지 않는다. 또 count가 `gap`인 경로의 `turn_tokens`가 숫자여도 통과한다(뮤턴트 `turn_tokens_P3: 1,2,3,4,5,6` → `0`). REQ-002 「A path with listing count 0 … yields gap for every value」와 어긋난다. — Severity: minor — Class: blocking — Required fix: `extract.txt`를 파싱해 세션 id별 값을 `$E`와 대조하는 heredoc을 추가하고, gap 경로의 `turn_tokens`도 `gap`이어야 함을 검사한다.

ND6. SCRATCH-VAR-GUARD — acceptance.md:L5, L21, L24 — 변수 블록은 「set in the same shell as each check」와 `S=$(mktemp -d)`를 요구한다. 그런데 이 전제부를 git 명령과 한 호출에 두면 가드가 거부한다(두 번 재현). 별도 호출로 두면 매번 새 셸이라 `$S`가 전달되지 않는다. 이 전제부를 쓰는 AC는 AC-013/015/016/018이다. `$S`에 파일을 쓰고 다음 명령이 읽는 구조라 문서대로는 실행되지 않는다. 문서 스스로 read-time 값에 쓰는 관례(한 번 출력하고 리터럴로 대입)를 `S`에는 적용하지 않았다. — Severity: minor — Class: blocking — Required fix: `S`를 read-time 값 목록으로 옮긴다(`mktemp -d`를 단독 실행 → `S=<path>` 리터럴 대입).

ND7. COST-LIVE-TRANSCRIPT — spec.md:L102-107; acceptance.md:L163-171 — `bb145fe7`은 이 레인 세션 자신의 전사이고 지금도 자란다(19:07 수정, 1,397행. R2 시점의 마지막 목록은 1,126행). REQ-008은 「각 non-initial scoped listing」을 시점 제한 없이 요구한다. M2 이후 이 세션이 한 번이라도 워크트리를 옮기면(예: 통합 창의 develop 워크트리 진입) M6의 재실행이 행을 더 내고 `cmp`가 실패한다. — Severity: minor — Class: blocking — Required fix: `m1-cost.md`에 `cutoff_line: <N>`(또는 cutoff ts)을 기록하고, 추출기 `--cost`와 AC-009 재실행이 같은 cutoff를 쓰게 한다.

ND8. PROBE-FLAG-ALLOWLIST — acceptance.md:L141-147; spec.md:L93 — 프로브 인자 검사는 금지 목록이고 `--max-turns 4`를 부분 문자열로 확인한다. 실측으로 `--max-turns 40`, `--max-turns 4 --max-turns 99`, `--settings /etc/evil.json`, `--allowedTools Bash --add-dir /Users/goos --dangerously-skip-permissions`가 모두 `2 0`으로 통과했다. `repo_status_sha`는 이 저장소만 보므로 fixture 밖 다른 경로의 변경은 잡지 못한다. — Severity: minor — Class: blocking — Required fix: `shlex.split`으로 허용 목록 검사를 한다(`-p <text>`, `--model haiku`, 정확히 한 번의 `--max-turns 4`, `--output-format json`, 선택적 `--allowedTools EnterWorktree`, UUID 형식 `--resume`).

ND9. AMENDMENT-EDGE — design.md:L24; spec.md:L112-118 — 개정문은 「On a card change」를 정의하지 않는다. phase가 끝났는데 다음 카드가 아직 없는 상설 세션은 이전 카드 컨텍스트를 무기한 들고 기다려야 하는가, 아니면 기존처럼 바로 clear하는가. 또 이동과 `/clear` 사이에 카드 작업(예: WT- 이름 변경)을 하지 않는다는 말이 [HARD] 줄 안에 없다. 핵심 의무는 약화되지 않았으므로 설명 보강 문제다. — Severity: minor — Class: optional — Required fix: [HARD] 줄은 그대로 두고, detail companion `### Card change in a standing session`에 두 가지를 적는다. 다음 카드가 없으면 기존 흐름대로 clear한다. 이동과 clear 사이에는 이동 외의 작업을 하지 않는다.

ND10. RESIDUAL-PROSE-2 — acceptance.md:L258(AC-016의 PASS 두 줄 검사가 산문), L336-337(AC-020 warn의 네 번째 명령과 block의 모든 명령 부재), L49(AC-022 t1257 분기, ND2와 겹침), L221-226(AC-014 covered 분기가 `$KT`만 검사하고 REQ-013이 요구하는 `$LP`·`$CL`은 보지 않음) — 「each is decided by the fenced commands」(L3)와 어긋난다. AC-020과 AC-014의 해당 분기는 DP-1 ≠ docs-only, DP-2 = covered일 때만 쓰인다. — Severity: minor — Class: blocking(선택된 분기에 한함) — Required fix: 각 분기를 명령과 기대 출력으로 바꾼다. AC-016에 `grep -c '^--- PASS: TestAlwaysLoadedTokenBudget ('` → `1`을 더한다.

ND11. MINOR-HYGIENE — spec.md:L142-150(REQ-020이 REQ-019 앞에 있음); acceptance.md:L194(AC-012 제외어 `develop`이 어휘 기준이라 develop을 언급한 카드 진입 단위도 면제됨); acceptance.md:L279-324(비-[HARD] 재전송 문장 「Where the next phase reuses …」을 지켜 주는 AC가 없어 삭제 뮤턴트가 통과함); AC-013의 `listing` 검사가 한국어 표면(`$CL`, `$LP`)의 「목록」에는 닿지 않음 — Severity: minor — Class: optional — Required fix: REQ 순서를 정렬한다. 재전송 문장에 `grep -cxF` 보존 검사를 둔다. 제외어는 통합 절 범위로 좁힌다.

## Regression Check (Iteration 3)
Defects from previous iteration:
- N1–N15: RESOLVED(근거는 §1 표). 잔여 관찰은 새 번호 ND1–ND11로 따로 적었다.
- 1회차 D1–D22: iter-2에서 해소 확인. 이번 회차 재측정(AC-021 범위, AC-016 예산, AC-018 보존)에서도 회귀 없음.
- 정체: 세 회차 연속 무변 결함 없음. 점수 0.55 → 0.74 → 0.83.

## Recommendation

**Verdict FAIL (0.83 < 0.85).** must-pass 7개(N/A 1)는 모두 통과했고 N1–N15는 모두 해결됐다. 그러나 이번 수정에서 blocking major 4건(ND1–ND4)이 새로 드러났고, 집계 점수가 Tier L 기준에 못 미친다. 이번이 3회 상한의 마지막 회차이므로 수정·재감사는 더 하지 않는다. 오케스트레이터는 상한 규정에 따라 PASS-with-debt, 범위 축소, 명시적 연장 중 하나를 사용자에게 올려야 한다.

**부채 이월 가능성**: 남은 결함 11건은 모두 설계 결정(N1의 (a), DP-1/DP-2 권고, Gate G1)을 다시 열지 않는다. 모두 검사 명령이나 기록 규칙을 국소적으로 고치는 것이다. 따라서 PASS-with-debt를 택하더라도 아래 조건을 run 착수 계약에 넣으면 안전하게 이월할 수 있다.

1. **M1 착수 전 적용**: ND3(conflict → gap 경로), ND4(P3 경계 검사. 그 전까지 `clear_restores_single_listing`은 gap), ND5, ND6, ND8. 모두 M1 산출물의 판정 도구이므로 측정 전에 고쳐야 한다. 측정 뒤에 고치면 이미 기록된 증거를 다시 판정해야 한다.
2. **M2 착수 전 적용**: ND7(cutoff 기록).
3. **M3(Gate G1) 전 적용**: ND2(T1175 출처 검사와 t1257 분기 명령). 리드 결정 §⑤의 순서 조건이 이 검사에 달려 있으므로 가장 우선한다.
4. **M6 전 적용**: ND1(인덱스 한 글자), ND10(선택된 분기에 한함).
5. **선택**: ND9, ND11. 오케스트레이터 재량이며, 목록이 길다는 이유로 FAIL을 만들지 않았다.

이월하지 말아야 할 결함은 없다. 다만 ND2와 ND4는 고치지 않으면 각각 「t1175 흡수 전 치환」과 「관측 없는 P-CLEAR 문구」를 막지 못한다. 그래서 이 두 건은 「나중에」가 아니라 해당 마일스톤 착수 조건으로 걸어야 한다.

---
증거(비커밋, 세션 scratchpad `…/bb145fe7-…/scratchpad/i3/`): `lint3.txt`(lint rc 0), `acx.py`(acceptance.md heredoc 추출·실행기), `mk.py`/`mk2.py`/`mk18.py`(합성 `$E`·프로브·kanban 뮤턴트), `run13.py`(ND1 재현), `run18.py`(AC-018 통제군), `run7b.py`(ND8), `listings.py`/`bound.py`(전사 목록·경계 판독), `t1175-kt.md`/`t1175-kl.md`/`t1175-mech.md`(`deab44714` 사본), `budget.txt`. 이 보고서는 커밋하지 않았다.
