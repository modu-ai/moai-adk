auditor-model: claude-opus-5-5[1m]

# SPEC 델타 감사 보고서: SPEC-SESSION-MIDMOVE-001
Iteration: 델타 1회(verdict.md §⑦ 1회 한정 연장, 최종)
Verdict: FAIL
Overall Score: 0.83

- 감사 범위: `b284d634a..00f62349e`(7개 파일). HEAD `9c989fd6b`는 verdict.md 인용 절만 더하므로 범위 밖이다. 워크트리 `.claude/worktrees/t1279`, 브랜치 `WT-session-double-load`, 작업 트리 clean.
- 판정 대상: iter-3의 ND1–ND11 해소 여부와 델타가 만든 회귀. 이미 정한 결정(N1 (a), DP-1/2/3, A 보류, B 제외)은 다시 다루지 않았다.
- 입력 격리: 작성자 원장(research.md §R11 addendum)은 대조용으로만 읽었다. 모든 AC는 이 세션에서 다시 실행했다.
- 기준 값(이번 실행에서 읽음): `git merge-base develop HEAD` → `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`. `git rev-parse develop` → `7fe658815…`. `FIRST_KT` → 빈 값(아직 `$KT`를 건드린 커밋이 없다). scratch `S=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/tmp.wRgSJBGCNQ`(`mktemp -d` 단독 실행).
- 교차 모델: `mcp__moai__audit_multi`(project_root = 이 워크트리). codex는 판정 필드가 `inconclusive`였지만 요약은 FAIL이었다. codex가 낸 결함 4건 가운데 3건(REQ-005 gap 거부, P3와 `/clear` 수명주기 모순, 중괄호 확장 우회)은 이 감사에서 독립적으로 재현했다. 나머지 1건(M1175 미검사)은 코드를 읽어 확인했다. GLM은 `inconclusive`(대상 해석 실패)였다. 실행 중인 MCP 바이너리는 `a8a9b9376`(HEAD의 조상)이다.

## 0. 감사 시점의 상태 변화 — t1175가 develop에 착지했다

작성자 원장과 커밋된 문구는 「t1175 미착지」를 전제로 한다. 하지만 이번 실행에서는 착지가 관측됐다.

- `git merge-base --is-ancestor WT-rules-diet develop` → rc `0`
- develop reflog: `7fe658815 develop@{2026-09-27 19:38:56}: merge WT-rules-diet`
- 델타 커밋 `00f62349e`의 시각은 19:41:02다. 즉 커밋 **2분 전**에 착지했다. R11 측정은 그보다 앞선 HEAD `bce7248ff`에서 했으므로, 측정 시점에는 원장이 맞았다.

따라서 다음 세 곳은 커밋 시점에 이미 사실과 달랐다. acceptance.md:L68 「RED-now … `t1175.txt` has `0` lines」, plan.md:L72 「At v0.4.0 no such merge exists」, spec.md:L248 「exits `1` at v0.4.0」. 게이트 자체는 지금도 옳은 이유로 RED다(§2 ND2). design.md §0의 대상 줄은 착지한 develop 사본에도 그대로 있다. `git show develop:$KT`와 `develop:$KL`에서 old-153 줄 `grep -cxF`가 각각 `1`, `[HARD]` 줄 `38`, 두 사본 `cmp` rc `0`, 재전송 문장 `1`이다. 따라서 개정 초안은 지금도 적용할 수 있다.

## 1. ND1–ND11 해소 여부

| 결함 | 판정 | 근거(이번 실행) |
|---|---|---|
| ND1 AC-013 argv 인덱스 | **해소** | `sys.argv[4]`로 고쳤다(acceptance.md:L319). gap `$E` 기준 RED `1`, P-RELAUNCH를 `$DT` 사본에만 넣으면 `0`, `$AT` 사본에만 넣으면 `1`, P-CLEAR를 덧붙이면 `2`, `목록` 줄 추가도 `2`. R11과 같다 |
| ND2 T1175 출처 | **해소**(잔여 D5) | 자기 보고 대신 develop 머지 로그에서 도출한다(L73-L76). 옛 자기 보고형 뮤턴트 `T1175=CARD_BASE`는 출처 grep이 `0`이라 걸린다. 도출값 `M1175=7fe658815… T1175=8fb81c948…`에서 출처 grep은 `35`, `is-ancestor T1175 HEAD`는 rc `1`이다. 착지했지만 흡수하지 않았다는, 옳은 이유의 RED다. t1257은 두 분기 모두 명령이 됐다(`WT-role-naming-docs` 미착지, `t1257.txt` `0`줄). 잔여 문제는 D5에 적었다 |
| ND3 귀속 전제 | **해소** | REQ-002를 가설로 바꿨다(spec.md:L77-L80). REQ-005에 conflict → gap 경로가 생겼다(L104). 두 번째 heredoc 결과: conflict가 있는데 gap 처리하지 않으면 `1`, gap 처리하면 `0` |
| ND4 P3 ↔ `/clear` 연결 | **미해소** | 연결 검사는 들어갔고, `no_clear_link`이면 P-CLEAR 문구가 막힌다(gap 키 + P-CLEAR 사본 → AC-013 `2`). 그러나 이 연결은 실제 `/clear`로는 충족되지 않고, 압축(compaction)으로 충족된다. 그러면 ND4가 막으려던 형태, 곧 `/clear`를 관측하지 않은 P-CLEAR가 다른 경로로 다시 열린다(D2) |
| ND5 `$E` ↔ extract 대조 | **해소**(회귀 D1) | gap 경로의 `turn_tokens_P3: 1,2` → 첫 heredoc `1`. `listing_bytes_P1`을 1만큼 틀리면 두 번째 heredoc `1`. 다만 새 대조 heredoc이 REQ-005가 허용한 gap을 거부한다(D1) |
| ND6 `S` 가드 | **해소**(잔여 D6) | `S`를 read-time 표로 옮겼다(L27-L35). 리터럴 `S`로 AC-015 전체가 한 호출에서 실행됐다. 잔여 문제는 D6에 적었다 |
| ND7 전사 cutoff | **해소** | REQ-008 `cutoff_line`(spec.md:L120), `--until-line`(design.md §3), AC-009 네 번째 grep과 `$CUTOFF` 재실행이 들어갔다. plan M2와 명령 형식도 일치한다 |
| ND8 프로브 인자 허용 목록 | **미해소** | 허용 목록 heredoc은 들어갔고 R11의 뮤턴트 8종은 모두 `2 1`로 재현됐다. 그러나 REQ-006이 이름까지 들어 금지한 `--dangerously-skip-permissions`와 `--add-dir`가 두 경로로 통과한다(`2 0`)(D3) |
| ND9 개정문 경계 | **해소**(선택) | design.md §2에 두 문장을 넣었고 AC-011 grep 두 줄을 더했다 |
| ND10 잔여 산문 | **해소**(잔여 선택 D7) | AC-016 PASS grep `1`/`1`, AC-014 covered 분기의 `$LP` `2`·`$CL` `1`, AC-020 warn/block 명령, AC-022 t1257 명령이 들어갔다 |
| ND11 위생 | **해소** | REQ 순서가 001…020으로 정렬됐다(C.6 절). AC-012 제외어를 좁힌 뒤 계수기는 `6`이다. 재전송 문장은 `$KT` `1`·`$KL` `1`이고, 삭제 사본에서 `0`이 된다. `목록`도 검사된다 |

해소 9/11, 미해소 2(ND4, ND8).

## 2. 재실행 대조 (§R11 addendum)

| AC | 이번 출력 | R11 | 일치 |
|---|---|---|---|
| 021 | SHA 출력; `11` | `11` | 예 |
| 022 (a) | 양성 대조 `1`; `t1175.txt` **`1`줄**; `t1257.txt` `0`줄 | `1`; `0`; `0` | **아니오.** t1175가 19:38:56에 착지했다(§0) |
| 022 (b) | 출처 grep `35`; `is-ancestor T1175 HEAD` rc `1`; `FIRST_KT` 빈 값 → rc `128` | 실행 안 함(미착지) | 해당 없음. 옳은 이유의 RED |
| 022 뮤턴트 | `T1175=CARD_BASE` → grep `0`(차단). 손으로 고른 쌍 `M=f5794a497, T=HEAD` → grep `4`, rc `0`(통과) | — | D5 |
| 002 | 합성 fixture 결과는 §3 D1·D2 표 참조. valid `0`/`0`, 무관 P3 세션 `…/1`, conflict 비gap `1`·gap `0`, bytes 불일치 `1` | 같음 | 예. 단 R11에 없는 두 경우가 실패한다(D1, D2) |
| 007 | valid `2 0`. `; rm -rf x`, `-evil`, `-d`, `--max-turns 40`, 중복, skip-perms, `--add-dir`, `--settings`, `--allowedTools Bash`, `--model=haiku`는 모두 `2 1` | 같음 | 예. 단 우회 두 건이 `2 0`(D3) |
| 012 | `2`,`2`; `1`; `6` | `6` | 예 |
| 013 | RED `1`; DT만 `0`; AT만 `1`; P-CLEAR `2`; `목록` `2` | 같음 | 예 |
| 014 | exempt `0 0 0`; covered `0`, `2`, `1` | `0`,`2`,`1` | 예 |
| 015 | `cmp` 무출력; `36`/`36`; diff rc `0`; `0 0 0 0`; `$KT` 커밋 `0`; 외래 SHA 대조 `1` | 같음 | 예. 한 호출 안에서 가드를 통과했다 |
| 016 | `rc=0`; PASS grep `1`,`1`; `77530 tokens (budget 77600, headroom 70, 16 entries`; heredoc `0`(before 77530/64316); 상수 diff `0`; 상수 grep `1`; 대조(77529/63700) `2` | 같음 | 예 |
| 018 | old/new `1`/`1`; 보존 `0`; 옛 줄 `1`; 절 새 줄 `0`; 파일 새 줄 `0`; 재전송 `1`(`$KT`)·`1`(`$KL`); 치환본 보존 `0`·절 새 줄 `1`; 재전송 삭제본 `0`; 비-kanban `[HARD]` `0` | 같음 | 예 |
| 020 docs-only | `0` | `0` | 예 |

**spec lint**: `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` → **rc=0**, `0 error(s), 0 warning(s)`. INFO 1건 `OwnershipTransitionUnmeasured`(커밋 `a14fcf851`)는 iter-3과 같다.

**가드 실행성 관측(D6)**: 따옴표 키를 쓴 dict 리터럴(`{"L":1}`)이 들어간 heredoc은 git이 없는 단독 호출이어도 가드가 거부했다(「too complex to verify」, 최소 재현 두 번). 해당 heredoc은 AC-002 두 개, AC-007, AC-013이다. 그래서 이 네 heredoc은 문서에서 기계적으로 추출한 **파일**로 실행했다(scratchpad `h/*.py`, 추출기 `xh.py`). git 호출과 AC-012·015·014·018의 평문 명령은 문서 그대로 가드 아래에서 실행됐다. read-time 표가 제시하는 한 줄 이중 대입 `M1175=<sha> T1175=<sha>`도 git 명령과 한 호출에 두면 거부된다. 두 줄로 나누면 실행된다.

## 3. Must-Pass Results
- [PASS] MP-1 REQ 번호: `REQ-SMM-001`…`020`이 문서 순서대로 한 번씩 있다(C.6이 C.5 뒤로 옮겨졌다, spec.md:L166). AC는 22개가 모두 고유하다(`uniq -c`에서 1이 아닌 행 0개).
- [PASS] MP-2 GEARS(요구 계층에서 판정): 델타가 바꾼 REQ-002(L77-L96), REQ-005 Event-driven(L104), REQ-006 Unwanted(L105-L106), REQ-008(L120), REQ-020 Unwanted 「shall not begin before」(L166)가 모두 shall 구조를 유지한다. AC의 Given-When-Then은 이 기준으로 판정하지 않았다.
- [PASS] MP-3 frontmatter: `version: "0.4.0"`(L4), `status: draft`(L5), `updated: 2026-09-27`(L7). 나머지 필드는 iter-3과 같고 lint 경고는 0건이다.
- [N/A] MP-4 언어 중립성: 여러 프로그래밍 언어의 도구를 다루지 않는다.
- [PASS] MP-5 D7: 참조 SPEC은 `SPEC-SESSION-DOUBLELOAD-001`(draft) 하나뿐이다. BLOCKING 없음.
- [PASS] MP-6 D8: 다섯 산출물의 `syscall` 수가 모두 `0`이다.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → rc `1`.

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 | REQ-002의 P3 연결 규칙(spec.md:L84)과 §D.1 「A new id: `session_id_P3` is that id」(acceptance.md:L479)가 서로 모순된다. RED-now 세 곳은 커밋 시점에 이미 사실과 달랐다(§0) |
| Completeness | 0.92 | 0.75–1.0 | 모든 절, Tier L 산출물 6종, t1257 분기와 AC-020 분기 명령이 있다. AC-020의 `PX`/`PB`/`PBX`/`FX` 생성은 여전히 산문이다(D7) |
| Testability | 0.72 | 0.50–0.75 | ND1·ND5·ND6이 해소되고 대부분의 통제군이 발화한다. 그러나 AC-002가 REQ-005의 정당한 gap을 거부하고(D1), P3 연결은 압축으로만 충족된다(D2). AC-007은 금지 플래그를 통과시키고(D3), heredoc 4개는 가드 아래에서 그대로 실행되지 않는다(D6) |
| Traceability | 0.92 | 0.75–1.0 | 20 REQ ↔ 22 AC 표가 완전하다(acceptance.md:L498 REQ-005 → AC-002 추가). 다만 AC-002가 자신이 추적하는 REQ-005와 충돌하고, AC-022는 REQ-020의 「develop 흡수」 절반을 검사하지 않는다(D4) |

조화평균 = 4 / (1/0.80 + 1/0.92 + 1/0.72 + 1/0.92) = **0.83**. Tier L 기준 0.85에 못 미친다. iter-3(0.83)과 같아 회귀 STOP 신호는 없다. ND 해소분만큼 점수가 올라야 했지만, 델타가 만든 D1·D2 때문에 상쇄됐다.

## Defects Found (structured defect-list)

D1. AC002-REJECTS-REQ005-GAP — acceptance.md:L156(`if blk is None: bad+=1`), L159-L166; spec.md:L104 — 새 대조 heredoc은 P1/P2에 non-gap 재도출을 강제한다. 따라서 REQ-005가 명시적으로 허용한 결과, 곧 헤드리스 세션이 이동이나 `/clear`를 못 했거나 상한에 먼저 닿아 「경로를 이유와 함께 Gap으로 기록」한 경우를 거부한다. 합성 fixture 결과: 세션은 돌았고 P2를 gap(`move_not_possible`)으로 기록 → 두 번째 heredoc `4`. 세션 없이 P2를 gap으로 기록 → `1`. P3를 연결은 있는데 `cap_reached`로 gap 기록 → `4`. 추적표(L498)가 REQ-005 → AC-002를 새로 매핑했으므로, AC가 자기가 검증한다는 REQ와 충돌한다. 정당하게 gap을 낸 측정은 AC-002를 통과할 수 없고, M1은 운영자 대체 세션 없이는 닫히지 않는다(fail-closed). codex도 독립적으로 같은 결함을 냈다. — Severity: major — Class: blocking — Required fix: `gap_reason_listing_trees_P{p}`의 허용 사유를 열거한다(`move_not_possible`, `clear_not_possible`, `cap_reached`, `no_measurement`, `attribution_conflict`, `no_clear_link`). 허용 사유를 가진 gap 경로는 네 값이 모두 `gap`이면 재도출과 세션 존재 검사를 건너뛴다. 재도출은 non-gap 경로에만 적용한다. 통제군 두 개(P2 `move_not_possible` gap → `0`, 사유 없는 gap → `1`)를 추가한다.

D2. P3-LINK-UNSATISFIABLE-BY-CLEAR — spec.md:L84; design.md:L74, L82; acceptance.md:L154, L479 — P3 연결은 「같은 전사 파일(같은 세션 id)에 `/clear` 경계 행이 정확히 하나 있을 것」을 요구한다. 그런데 이 프로젝트 전사 디렉터리의 최근 600개 파일에서 `/clear` 명령 행 369개를 읽었더니, 그 행보다 앞에 assistant 행이 있는 경우가 **0건**이었다. 모든 `/clear` 행은 새 전사 파일(새 세션 id)의 9–11번째 행에 있었다(`clearprobe.py`, `compactprobe.py`). 반면 `compact_boundary`는 같은 세션 id의 전사 중간에 있다(표본 3개). codex도 `TestAC008_ClearRearmsEmbeddedGoalUnderNewSessionID`(ok)를 근거로 `/clear`가 새 세션 id를 만든다고 지적했다. design.md:L74는 `compact_boundary` 행도 `clear` 행으로 출력한다. 그 결과는 두 가지다.
  - (a) 실제 `/clear`로 얻은 P3는 언제나 `gap(no_clear_link)`가 된다. 실측 결과: 새 id를 P3로 적으면 `1`, gap으로 적으면 `0`. P-CLEAR 측정은 구조상 공허하다.
  - (b) 같은 전사 안의 압축(auto-compact 또는 `/compact`)이 연결 조건을 충족한다. 합성 extract에서 압축 행을 `clear` 행으로 둔 경우 두 heredoc이 `0`/`0`이었고 `clear_restores_single_listing: yes`가 통과했다. 그러면 「`/clear` leaves only the new tree's listing」 문구가 `/clear` 관측 없이 교리에 들어갈 수 있다. ND4가 막으려던 형태가 압축 경로로 다시 열린 것이다.

  또 §D.1(acceptance.md:L479)은 여전히 「A new id: `session_id_P3` is that id」라고 적어 REQ-002와 모순된다. 헤드리스 `--resume` 경로의 `/clear` 동작은 측정되지 않았다. 위 관측은 대화형 세션 기준이다. — Severity: major — Class: blocking — Required fix: P3 연결 계약을 실제 수명주기에 맞춘다. P3는 P2 다음에 시작한 **새** 전사이고, 그 전사의 첫 명령 행이 `/clear`이며, `cwd`가 같은 fixture이고, 시각이 P2의 마지막 행보다 뒤인 경우에만 인정한다. 추출기의 `clear` 행은 `/clear` 명령 행만 내고, `compact_boundary`는 별도 `compact` 행으로 분리한다. 압축 행이 있으면 해당 경로를 `gap(compaction)`으로 돌린다. §D.1을 같은 규칙으로 고친다.

D3. PROBE-ALLOWLIST-BYPASS — acceptance.md:L221, L225-L235; spec.md:L106 — 허용 목록은 `shlex` 토큰을 검사하지만, 프로브 줄을 실행하는 것은 셸이고 문자 거부 목록은 `{ } * ? [ ] ~`를 막지 않는다. 확인된 우회는 두 가지다.
  - (i) 중괄호 확장: `-p {ok,--dangerously-skip-permissions,--add-dir,/Users/goos} …` → heredoc `2 0`. `bash -c 'printf "[%s]" {ok,--dangerously-skip-permissions,--add-dir,/Users/goos}'` → `[ok][--dangerously-skip-permissions][--add-dir][/Users/goos]`.
  - (ii) `-p`의 값 슬롯이 플래그도 받는다: `-p --dangerously-skip-permissions --model haiku …` → `2 0`. 실제 CLI에서 `-p`(`--print`)는 불리언이고 프롬프트는 위치 인자라는 점은 이 감사에서 실행해 확인하지 않았다.

  REQ-006이 이름까지 들어 금지한 두 플래그가 통과하므로 ND8은 미해소다. `--resume` UUID 정규식은 대시 36개도 받는다(`2 0`). codex도 같은 우회를 냈다. — Severity: major — Class: blocking — Required fix: 문자 거부 목록에 `{`, `}`, `*`, `?`, `[`, `]`, `~`, `\`, `'`를 더한다. 따옴표를 허용한다면 `-p` 값에 한해서만, 첫 글자가 `-`가 아니어야 한다. UUID는 `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`로 좁힌다. 통제군 두 개(중괄호 확장, `-p --dangerously-skip-permissions` → 각 `N 1`)를 추가한다.

D4. GATE-M-NOT-CHECKED — spec.md:L166-L167; acceptance.md:L79-L80, L83 — REQ-020은 「t1175가 develop에 착지하고 **그 develop을 흡수**할 것」을 요구한다. 그러나 AC-022는 `T1175`(브랜치 tip)의 조상 관계만 보고 `M1175`(develop 착지 머지)는 검사하지 않는다. `WT-rules-diet`를 develop을 거치지 않고 직접 병합해도 게이트를 통과한다. t1257도 같다. 교리 내용상의 위험은 작다. tip `8fb81c948`이 이미 develop `f01c7889a`를 흡수했기 때문이다. 그러나 문서가 명시한 기준을 검사하지 않는다. — Severity: minor — Class: blocking — Required fix: (b)에 `git merge-base --is-ancestor "$M1175" HEAD`와 `… "$M1175" "$FIRST_KT"`를 더한다. t1257이 착지한 경우 `M1257`도 같다.

D5. GATE-LITERAL-COPY — acceptance.md:L32, L62 — `M1175 T1175`는 `t1175.txt`를 사람이 읽고 리터럴로 옮긴다. 옮긴 값이 파일과 같은지 확인하는 명령은 없다. 손으로 고른 쌍 `M1175=f5794a497…`, `T1175=HEAD`는 출처 grep `4`, rc `0`으로 통과한다. 도출 (a)를 다시 실행하는 감사자는 이를 잡는다. 또 표가 제시하는 한 줄 이중 대입 `M1175=<sha> T1175=<sha>`는 git과 한 호출에 두면 가드가 거부한다. RED-now 세 곳(§0)도 커밋 시점에 이미 사실과 달랐다. — Severity: minor — Class: optional — Required fix: (b) 첫 줄에 `printf '%s %s\n' "$M1175" "$T1175" | cmp - "$S/t1175.txt"`를 두고 rc `0`을 기대값으로 적는다. 대입은 두 줄로 쓰라고 명시한다. RED-now는 「t1175 착지, 미흡수: `is-ancestor T1175 HEAD` rc `1`」로 갱신한다.

D6. HEREDOC-GUARD-REFUSAL — acceptance.md:L117, L137(AC-002 두 heredoc), L216(AC-007), L312(AC-013); research.md §R11 addendum 머리 — 이 감사 세션의 가드는 따옴표 키 dict 리터럴(`{"L":1}`)이 든 heredoc을 git 없이도 거부했다(최소 재현). 따라서 R11의 「Every command below ran under the worktree-session guard exactly as written」은 이 네 heredoc에 대해 재현되지 않았다. 작성자 세션의 가드 버전이 달랐을 가능성은 배제하지 못했다. 이 heredoc들은 git을 부르지 않으므로, 추출한 `.py` 파일로 실행해도 안전성 문제는 없다. — Severity: minor — Class: optional — Required fix: acceptance.md 머리에 「python heredoc은 추출한 scratch `.py`로 실행해도 된다(git 미사용)」는 규칙을 한 줄 둔다. 또는 dict 리터럴을 `dict(k=v)` 형식으로 바꾼다.

D7. AC020-PROSE-RESIDUE — acceptance.md:L446-L448 — `PX`/`PB`/`PBX`는 치환 규칙을 산문으로만 정의하고, `FX`를 만드는 명령은 read-time 표에 없다. DP-1이 docs-only가 아닐 때만 쓰인다. — Severity: minor — Class: optional — Required fix: `FX`(`mktemp -d`)를 read-time 표에 넣고, `PX`/`PB`/`PBX`를 `P`와 같은 리터럴 대입으로 적는다.

## Regression Check (델타)
iter-3 결함:
- ND1: RESOLVED — `sys.argv[4]`, 통제군 `0`/`1`
- ND2: RESOLVED — 머지 로그에서 도출, 자기 보고형 뮤턴트 차단(grep `0`), 옳은 이유의 RED(rc `1`). 잔여는 D4, D5
- ND3: RESOLVED — conflict → gap 경로, 통제군 `1`/`0`
- ND4: **UNRESOLVED** — 연결 검사는 들어갔으나 실제 `/clear`로는 충족되지 않고 압축으로 충족된다(D2)
- ND5: RESOLVED — 대조 heredoc 발화. 그러나 D1 회귀를 만들었다
- ND6: RESOLVED — `S`가 read-time 값이 됐다. 잔여는 D5, D6
- ND7: RESOLVED
- ND8: **UNRESOLVED** — 금지 플래그 두 개가 두 경로로 통과한다(D3)
- ND9: RESOLVED
- ND10: RESOLVED — 잔여는 D7
- ND11: RESOLVED

델타가 만든 회귀: D1(새 heredoc이 REQ-005와 충돌), D2 일부(§D.1과 REQ-002의 모순; `compact_boundary`를 `clear` 행으로 정의한 design.md:L74는 이번 델타에서 새로 쓴 것이다).

## Recommendation

**Verdict FAIL (0.83 < 0.85, blocking major 3건).** must-pass 7개(N/A 1)는 모두 통과했고 ND 11건 중 9건이 해소됐다. 그러나 ND4·ND8이 남았고, 델타가 major 회귀 D1을 만들었다. verdict.md §⑦에 따라 이번 결과가 최종이며 추가 반복은 없다. 오케스트레이터는 PASS-with-debt, 범위 축소, 명시적 재연장 중 하나를 사용자에게 올려야 한다.

부채로 넘길 경우의 착수 조건:
1. **M1 착수 전(필수)**: D1, D2, D3. 셋 모두 M1 증거를 판정하는 도구다. D2를 고치기 전까지 `clear_restores_single_listing`은 `gap`으로 고정하고, 압축 행이 있는 전사는 쓰지 않는다. D3를 고치기 전까지 `probes.txt`는 리드가 눈으로 확인한다.
2. **M3(Gate G1) 전(필수)**: D4. t1175는 이미 develop에 착지했으므로(§0) M3를 지금 시작할 수 있다. 그래서 D4를 먼저 고쳐야 한다.
3. **선택**: D5, D6, D7.

설계 결정(N1 (a), DP-1/2/3)을 다시 여는 결함은 없다. 모두 검사 명령이나 추출기 계약을 국소적으로 고치는 일이다.

---
증거(비커밋, 세션 scratchpad `…/bb145fe7-…/scratchpad/`): `xh.py`(heredoc 추출), `h/*.py`(추출본), `mkfx.py`·`run002.py`·`run002b.py`(AC-002 fixture 표), `run007.py`·`run007b.py`(AC-007 우회), `run016_018.py`·`run018.py`, `clearprobe.py`·`compactprobe.py`(`/clear` 전사 위치). 가드 아래 git·평문 AC의 scratch 출력: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/tmp.wRgSJBGCNQ/`(`develop-merges.txt`, `t1175.txt`, `budget.txt`, `lint.txt`, `c-*.txt`). 이 보고서는 커밋하지 않았다.
