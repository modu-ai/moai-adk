# SPEC Review Report: SPEC-SESSION-DOUBLELOAD-001
Iteration: 2/2 (Tier M 상한 2 — 마지막 회차, 사용자 에스컬레이션 대상)
Verdict: FAIL
Overall Score: 0.68 (Tier M 통과 기준 0.80 미달, 1회차 0.64 대비 상승 — STOP 신호 없음)

- 감사 대상: 워크트리 `.claude/worktrees/t1219`, 커밋 `a103df537`(`git rev-parse --short HEAD` → `a103df537`, 작업 트리 clean), 브랜치 `WT-local-doc-dedup`
- 기준: `git merge-base develop HEAD` → `e464fd5d0`. 로컬 develop tip 은 `18573765b` 로 이미 앞서 있다.
- 읽은 아티팩트: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`, 1회차 보고서 `plan-audit.md`, 전제 원천 `verdict.md`·`probe-*.txt`
- 작성자 추론 맥락은 받지 않았다(M1 맥락 격리). 오케스트레이터가 지목한 점검 축은 감사 범위 지정으로만 썼다.
- 교차 모델: `mcp__moai__audit_multi`(project_root = 이 워크트리). codex 는 본문을 돌려줬지만 판정 필드가 `inconclusive` 였고, GLM 은 리뷰 대상 해석 실패로 `inconclusive`. codex 지적은 이 감사가 직접 재확인한 것만 채택했다(아래 각 결함에 표시).

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L60–L100` 에 REQ-SDL-001..016 이 빈칸·중복 없이 이어진다(`grep -nE '^- \*\*REQ-SDL-[0-9]+'` 16행). AC 는 `acceptance.md:L30–L153` 에 AC-SDL-001..016.
- [PASS] MP-2 GEARS 형식 (요구사항 층 기준): 16개 REQ 모두 다섯 패턴 중 하나다. 예: `L60` "The measurement milestone shall complete …"(Ubiquitous), `L73` "When a cap is reached …, the measurement shall stop"(Event), `L75` "The measurement probes shall not:"(Unwanted), `L84` "Where the measurement shows …, the dispatch doctrine shall name …"(Where). REQ-011(`L90–L95`)은 Ubiquitous 본문에 Where 절을 하나 덧붙인 복합형으로 허용 범위다. AC 의 Given-When-Then 은 검증 층이라 여기서 채점하지 않았다.
- [PASS] MP-3 프런트매터: `spec.md:L2–L14` 에 12개 필수 필드(`version: "0.2.0"` 따옴표, `created: 2026-09-26`, `updated: 2026-09-27`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열)와 `tier: M`. 거부 별칭 없음. `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` → `✓ No findings`, **rc=0**(설치 바이너리 `a8a9b9376`). 트리 소스로도 `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` → rc=0.
- [N/A] MP-4 언어 중립성: 다중 프로그래밍 언어 도구 SPEC 이 아니다. 템플릿 본문 중립성은 REQ-SDL-013/AC-SDL-012 가 다룬다.
- [PASS] MP-5 D7: 본문 SPEC 참조는 `SPEC-SESSION-WORKTREE-001` 하나, `status: completed`. BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall` 네 파일 모두 0 → 자동 통과.
- [PASS] MP-7 명확화 게이트: `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/` → 출력 없음, rc=1. 세 결정은 `spec.md:L144–L148` §F 와 `plan.md:L148–L152` 에 결정 문장으로 옮겨졌다.

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.70 | 0.50–0.75 | REQ 문장 자체는 명료. 그러나 AC-007 의 발동 조건 "zero duplicate count" 에 대응하는 E 필드가 없음(N6), 경로 A·#7 의 실행 형태 미정(N5), M2 결정표에 "워크트리 시작이 primary 파일도 싣는다" 행이 없음(N2) |
| Completeness | 0.75 | 0.75 | 섹션·프런트매터·`### Out of Scope — …` H3(`spec.md:L119` 등) 모두 있음. 결정 분기 누락(N2·N7), 품질 게이트가 M3 편집이 실제로 걸리는 가드를 빠뜨림(N9) |
| Testability | 0.50 | 0.50 | 거짓 FAIL 이 확정인 검사(AC-006, N4), 거짓 PASS 가 가능한 검사(AC-001 t1175 조건 N3, AC-003 중복 줄 N12, AC-008 `none` 우회 N7), 이름표만 있고 경로가 없는 셸 토큰(N13) |
| Traceability | 0.90 | 0.75–1.0 | 16 REQ 모두 AC 있음, 추적표(`acceptance.md:L162–L179`) 정합. REQ-012 의 "템플릿 먼저 편집" 순서만 검사가 없다(선택 사항) |

조화평균 ≈ 0.68.

## D1–D20 해소 여부 (커밋본 대조)

| 결함 | 판정 | 근거 |
|---|---|---|
| D1 미결 마커 3건 | 해소 | 마커 grep rc=1. `spec.md:L144–L148`, `plan.md:L148–L152` |
| D2 영향 표면 누락 | 해소 | §D 에 DT/DL/AT/A 추가(`spec.md:L104–L113`). worktree-integration.md 는 사유와 함께 제외(`L127–L129`). 기준 트리에서 `cmp T L`=0, `cmp DT DL`=0, AGENTS 단락 diff=0 확인 |
| D3 REQ-008 조건 | 해소 | `spec.md:L84` 가 B 대리를 명시, `acceptance.md:L84` `path_a_proxy` |
| D4 집합 대신 개수 | 부분 해소 | 경로 목록 필드(`spec.md:L68–L71`, `acceptance.md:L47–L50`)와 primary 경로 부재 검사(`L92`)는 들어왔다. 그러나 SRC 목록에 `.claude/worktrees/` 경로가 **있다**는 양성 존재 검사가 없어, 추출이 사용자 층 경로만 돌려줘도 (a) 가 통과한다 |
| D5 토큰 턴 정의 | 해소 | `spec.md:L71`, `plan.md:L57`, `acceptance.md:L51` |
| D6 이름으로 프로세스 판정 | 형태는 해소, 대체 검사 결함 | pgid 기반(`acceptance.md:L66`)·세션 id 귀속(`L67–L69`)으로 바뀌었다. 그러나 "다른 워크트리" 집합이 primary 를 포함해 거짓 FAIL 이 확정이다(N4) |
| D7 순서 검사 명령 | 해소 | `acceptance.md:L32–L35` 가 `$BASE..HEAD` 첫 커밋과 N/A 를 쓴다. 단 같은 AC 에 새로 붙은 t1175 조건이 이미 참이다(N3) |
| D8 양성 대조 모양 | 해소 | REQ-007(`spec.md:L80`), `acceptance.md:L74–L77` 가 `count=2` 와 probes 경로를 요구. 발동 조건 정의는 N6 |
| D9 REQ-010 미추적 | 해소 | `acceptance.md:L85–L86, L94–L95` |
| D10 프로브 부수 효과 | **미해소(회귀)** | (a) 복합 호출 `plan.md:L42`, (b) session_id 필터 `L52–L55` 는 해소. (c) 를 `--setting-sources user,project` 로 풀었는데, 이 플래그가 CLAUDE.local.md 로드를 막는다는 것을 이번 감사가 실측했다(N1). 측정 대상 파일을 측정에서 빼는 결과가 된다 |
| D11 서브에이전트 경로 | 해소 | `spec.md:L66–L67`, `plan.md:L69–L70`. 보조 사항은 N16 |
| D12 `/clear` 인계 순서 | 해소 | `spec.md:L90–L95`, AC-010 |
| D13 `/clear` 측정 수단 | 해소(수단 명시) | `plan.md:L72, L152`. 실행 형태가 §D·AC-006 과 맞지 않는 문제는 N5 |
| D14 AC-013 주 진입 형태 | 해소 | `acceptance.md:L131`. 기준 트리에서 `grep 'moai cc -w <card-id>' G \| grep -c 'EnterWorktree(<card-id>)'` → `1`, `grep -A3 … \| grep -cE '/clear\|relaunch'` → `0` 을 이번에 재확인 |
| D15 AC-010 범위 | 해소 | 절 단위 `sed` 범위(`acceptance.md:L108`), 기준 `BASE`. 기준 트리에서 절 9줄, `EnterWorktree` 0, `relaunch` 0, `re-sends the full pointer` 1 |
| D16 추적표 오매핑 | 해소 | `acceptance.md:L178–L179` |
| D17 exit-first 해석 | 해소 | `spec.md:L99` 마지막 문장 |
| D18 캡 선언 시점 | 해소(선택 사항 권고안 채택) | REQ-003(`spec.md:L72`), AC-002(`acceptance.md:L40–L42`). 잔여: 커밋 조상 관계는 실행 순서를 증명하지 않는다(N11) |
| D19 전제 과대 표현 | 문장은 해소, 전제는 반증됨 | `spec.md:L31` 이 한계를 적었다. 그런데 같은 트리의 기록이 primary 사본 로드를 보여 준다(N2) |
| D20 plan 절 번호 | 해소 | plan §A–§H 연속 |

요약: 20건 중 해소 16, 부분 해소 2(D4, D6), 미해소 1(D10 — 수정이 새 결함을 만듦), 문장만 해소되고 전제가 반증된 것 1(D19).

## 이번 감사에서 실행한 측정

| 검사 | 명령 | 관측 |
|---|---|---|
| spec lint | `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` | `✓ No findings`, rc=0 |
| 전체 lint(AC-016 둘째 항목) | `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` | `0 error(s), 4845 warning(s)`, `baseline: OK`, rc=0 |
| setting-sources 영향 | 스크래치 디렉터리에 `CLAUDE.md`(MANGO)·`CLAUDE.local.md`(PINEAPPLE), `unset MOAI_KANBAN… && timeout 180 claude -p … --max-turns 1 --model haiku [--setting-sources user,project] --output-format json` | 세션 `2d5df932`(플래그 있음) 전사: `Contents of …/ssprobe/CLAUDE.md` 1건뿐. 세션 `bbef73ec`(플래그 없음) 전사: `CLAUDE.md` 와 `CLAUDE.local.md` 둘 다. 모델 응답도 각각 MANGO / MANGO+PINEAPPLE |
| 경로 B 전제 | verdict 프로브 세션 `79ed0553` 전사(`…/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t1219/79ed0553-….jsonl`, `cwd` 29회 모두 t1219, `entrypoint: sdk-cli`, `2.1.283`) | 22행에 `Contents of /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md (user's private project instructions…)` 와 워크트리 사본이 **둘 다** 있다. t1219 의 `rule-load-audit.jsonl` 에도 같은 세션의 `/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md`, `load_reason: session_start` 행이 있다 |
| t1175 게이트 술어 | `git log --format='%h %s' --grep='t1175' develop` | `8d88ee013 … (card t1243)` 등 2건 — t1243 커밋 본문이 t1175 를 언급 |
| 그 커밋의 조상 여부 | `git merge-base --is-ancestor 8d88ee013 HEAD` | exit 0 |
| t1175 실제 착지 여부 | `git merge-base --is-ancestor WT-rules-diet develop` | exit 1 (미착지) |
| t1175 판 앵커 생존 | `git show WT-rules-diet:<T>` 에 AC-009/010/014 앵커 grep | 전부 1 이상, `^\[HARD\]` 32 |
| 증거 경로 ignore | `git check-ignore -v .moai/reports/t1219/{m1-measure.md,m1-caps.md,probes/*}` | 전부 `.gitignore:235:.moai/reports/*` 에 걸림. 기존 `verdict.md`·`plan-audit.md` 는 `git ls-files` 에 있음(강제 추가로 추적) |
| 워크트리 목록 | `git worktree list --porcelain \| grep '^worktree '` | 190개, 첫 줄이 primary `/Users/goos/MoAI/moai-adk-go` |
| InstructionsLoaded 기록 필드 | `internal/hook/instructions_loaded.go:L122–L127` | `timestamp, session_id, file_path, load_reason, globs, trigger_file_path` — 에이전트 식별 필드 없음 |
| 프로젝트 설정 권한 | `jq '.permissions' .claude/settings.json` | allow 에 `Agent`, `EnterWorktree`, `ExitWorktree` 포함 → `-p` 프로브의 도구 권한은 local 층 없이도 성립 |
| AGENTS 크기 가드 | `internal/config/token_budget_guard_test.go:L84 TestCodexContractByteCeiling`, `L54 TestAlwaysLoadedTokenBudget` | 존재. 루트 `AGENTS.md` 16,441 B, `AGENTS.md.tmpl` 19,177 B |

## Defects Found

N1. setting-sources 가 측정 대상을 지운다 — plan.md:L42, L46–L50 — `--setting-sources user,project` 는 local 층과 함께 **CLAUDE.local.md 로드 자체를 끈다**(위 측정표: 플래그 있음 → `CLAUDE.md` 만, 없음 → 둘 다; 전사의 `Contents of` 머리말로 확인, Claude Code 2.1.283). 이 SPEC 이 재려는 중복의 핵심이 바로 두 CLAUDE.local.md 사본이므로, 모든 프로브가 이 파일을 "0" 으로 보고하고 `premise_instructions_C` 가 거짓 `single` 로 기울 수 있다. `plan.md:L50` 의 "memory 로드는 바뀌지 않는다고 가정" 은 반증됐다. — Severity: critical — Class: blocking — Required fix: `--setting-sources` 를 프로브 명령에서 뺀다. LSEL 드레인은 부수 효과로 선언하고 `<primary>/.moai/state/lsel/` 목록과 인박스 오프셋을 전후 기록한다(또는 운영자 결정으로 다른 격리 수단을 고른다). 어떤 플래그 조합을 쓰든, 사전 점검(§C)에 이번 감사와 같은 "두 표식 파일 스크래치 프로브" 를 넣어 CLAUDE.local.md 가 실린다는 것을 전사로 확인한 뒤에 본 프로브를 돌린다.

N2. 경로 B 전제가 트리 안의 기록으로 반증되고, 결정표에 그 경우가 없다 — spec.md:L29–L31, L48–L52, L84; plan.md:L96–L101; verdict.md:L16 — 워크트리 cwd 에서 띄운 headless 세션 `79ed0553`(verdict Evidence 1 의 바로 그 프로브로 보인다: 시각 11:29Z, probe 파일 20:30 KST)이 primary `CLAUDE.local.md` 를 지침으로 실었다(전사 22행 `Contents of /Users/goos/MoAI/moai-adk-go/CLAUDE.local.md`, InstructionsLoaded `session_start` 행). L1 워크트리가 primary 체크아웃 **안**(`.claude/worktrees/`)에 있어서 상위 디렉터리 탐색으로 primary 사본이 붙는 것으로 보인다(원인은 추정, 로드 사실은 관측). 따라서 (1) verdict Evidence 1 의 "워크트리 사본 하나" 는 틀렸고, (2) §B 목표 "one instruction set" 은 진입 형태만으로는 L1 트리에서 닫히지 않을 가능성이 크며, (3) M2 결정표(`plan.md:L96–L101`)와 AC-008 에는 "A/B 가 상위 탐색으로 primary 지침 파일을 싣는다" 는 결과에 대한 행이 없다. AC-008(a) 의 정규식은 이 경우 런처 결정을 막아 주지만, 그다음 어떤 결정을 내려야 하는지는 정해져 있지 않다. — Severity: critical — Class: blocking — Required fix: verdict.md Evidence 1 을 정정한다. REQ-008/결정표에 이 행을 추가한다(예: `decision_primary: launcher-skills-only` 처럼 부류별로 나누거나, 지침 부류는 "진입 형태로 해결 불가" 로 E 에 기록하고 별도 카드로 넘김). 측정에 경로 E(`~/.moai/worktrees/` 아래 L2 트리처럼 primary 밖에 있는 트리에서 시작)를 넣을지 운영자에게 묻는다 — 캡은 예비 1회를 쓰거나, 캡 커밋 전에 9로 올린다.

N3. t1175 게이트 술어가 이미 참이다 — plan.md:L108–L109; acceptance.md:L36 — `git log --format=%H -1 --grep='t1175' develop` 는 지금 `8d88ee013`(t1243 커밋, 본문에 "t1175 merge window" 언급)을 돌려주고, 그 커밋은 HEAD 의 조상이다(exit 0). 반면 t1175 브랜치 `WT-rules-diet` 는 develop 에 없다(exit 1). 즉 Gate G1 두 조건과 AC-001 의 t1175 조건이 t1175 착지 전에 이미 통과한다. REQ-SDL-016(운영자 결정)을 기계적으로 지키지 못한다. — Severity: major — Class: blocking — Required fix: 커밋 메시지 grep 대신 착지 자체를 묻는다. 예: `git merge-base --is-ancestor <t1175 착지 커밋 SHA 또는 WT-rules-diet tip> develop` 과 같은 SHA 가 HEAD 의 조상인지. 브랜치 이름은 병합 뒤 지워질 수 있으므로 G1 통과 시점에 SHA 를 E 에 `t1175_landed: <sha>` 로 기록하고 AC-001 은 그 줄을 읽는다. 양성 대조: 지금 이 판정식이 exit 1 을 내는 것을 기록한다.

N4. AC-006 "다른 워크트리 비쓰기" 가 primary 를 포함해 거짓 FAIL 이 확정이다 — acceptance.md:L67–L69 — `git worktree list --porcelain` 의 첫 항목이 primary 다. primary 시작 프로브는 설계상 primary `rule-load-audit.jsonl` 에 자기 session_id 를 남기므로(`plan.md:L59`), "t1219 이외의 모든 목록 트리" 에서 session_id 가 나오지 않아야 한다는 조건은 반드시 깨진다. 같은 AC 의 양성 대조(`L69`)는 primary 적중을 오히려 기대하고 있어 AC 가 자기모순이다. — Severity: major — Class: blocking — Required fix: 검사 대상에서 t1219 와 primary 를 모두 뺀다(`git worktree list --porcelain | awk '/^worktree /{print $2}' | grep -vxF -e <primary> -e <t1219>`). 190개 트리를 도는 만큼 스윕 개수(검사한 트리 수)를 E 에 기록한다.

N5. 경로 A(#6)와 `/clear`(#7)의 실행 형태가 없고, §D·AC-006 과 양립하지 않는다 — plan.md:L39–L45, L71–L72; acceptance.md:L65–L66 — §D 는 모든 프로브를 `unset … && timeout 300 claude -p …` 한 형태로 적고 `commands.txt` 에 전부 남기라 하며, AC-006 은 그 파일의 **모든** 줄이 `… && timeout 300 claude ` 로 시작하길 요구한다. 그런데 #6 은 `moai cc -w t1219`, #7 은 대화형 세션이다. 기록하면 AC 가 깨지고, 빼면 "모든 프로브를 기록" 이 깨진다(codex 지적 #2, 이번 감사가 본문으로 재확인). 또 대화형 세션에서 스킬 추출 원천인 `--debug-file` 을 어떻게 켜는지(`moai cc` 가 인자를 넘기는지 미확인), pgid 를 어떻게 기록하는지, `/clear` 뒤 session_id 가 바뀌면 session_id 필터가 뒤쪽 로드 행을 놓치는지(미확인)가 정해져 있지 않다. #7 이 운영자 결정(레인 표준)의 유일한 근거이므로, 수단이 모호하면 결과가 조용히 `gap → relaunch` 로 떨어진다. — Severity: major — Class: blocking — Required fix: #6·#7 의 정확한 명령 줄을 plan 에 적는다(예: #7 은 `unset … && timeout -k 10 900 claude --debug-file P/clear.debug` 로 대화형 시작 → 운영자가 `EnterWorktree` → `/clear` → 한 턴). `/clear` 전후 session_id 를 둘 다 E 에 기록하게 한다. AC-006 의 환경 정리 검사는 "`unset MOAI_KANBAN… && timeout ` 으로 시작" 까지만 요구하도록 고치고, pgid 는 프로브별 캡처 방법을 적는다.

N6. AC-007 의 발동 조건이 E 필드와 연결돼 있지 않다 — acceptance.md:L72–L77; spec.md:L80 — "E records any zero duplicate count" 라고 하지만 E 에는 중복 개수 필드가 없다(경로 목록과 `dup|single|gap` 만 있다). 런처 결정의 근거인 "SRC 에 primary 경로 0" 이라는 0 도 대조를 요구받지 않는다. D4 에서 남은 양성 존재 검사도 여기 속한다. — Severity: major — Class: blocking — Required fix: 발동 조건을 "`premise_*_C: single` 이 하나라도 있거나 `decision_primary: launcher` 일 때" 로 기계적으로 적는다. AC-008(a) 에 `grep -E "^(skills|instruction)_paths_${SRC}: " E | grep -c '/\.claude/worktrees/'` 가 `2` 임을 추가한다.

N7. AC-008 에 역방향 조건이 없어 교리 AC 전체를 `none` 으로 비껴갈 수 있다 — acceptance.md:L87–L95 — (c) 는 "둘 다 single 이면 none" 만 말하고, "SRC 가 워크트리 경로만이고 전제가 dup 이면 launcher 여야 한다" 는 반대 방향이 없다. 그래서 `decision_primary: none` 을 적으면 AC-001 이 N/A, AC-009..014 가 N/A 가 되어 M3/M4 없이 DoD 를 통과한다. (b) 도 한 방향뿐이라 `clear_restores_single_load: yes` + `relaunch` 조합이 통과한다(codex #4, 이번 감사가 조건식을 읽어 재확인). — Severity: major — Class: blocking — Required fix: 결정을 증거의 함수로 완결한다 — `decision_primary` 는 (a) 조건이 성립하면 `launcher`, (c) 조건이면 `none`, N2 의 새 행이면 그 값, 그 밖이면 FAIL. `clear_restores_single_load: yes` ⇔ `move-clear-resend`(단 decision_primary 가 none 이 아닐 때).

N8. 증거 경로가 gitignore 대상이고, 원시 로그를 커밋·공개하게 된다 — plan.md:L60, L90; acceptance.md:L35, L39 — `.gitignore:235 .moai/reports/*` 가 E·CAPS·probes 를 모두 무시한다. plan 은 "commit E and probes/" 라고만 해서 `git add -f` 가 없으면 AC-001/002 의 `git log --diff-filter=A` 가 비어 실패한다. 반대로 강제 추가하면 `--debug-file` 과 stream-json 원본이 리드 일괄 push 로 공개 저장소 `origin/develop` 에 실린다. 그 안에 무엇이 들어가는지(계정 식별자, 훅 출력, 개인 메모리 본문 등)는 이번에 확인하지 않았다(미검증 위험). — Severity: minor — Class: blocking — Required fix: E, CAPS, `commands.txt`, 추출된 줄만 담은 발췌 파일을 `git add -f` 로 커밋한다고 명시한다. 원시 debug/stream-json 은 커밋하지 않거나, 커밋 전 검사(이메일·토큰 패턴 grep)와 크기 상한을 둔다. AC-002 의 `F` 는 커밋되는 파일(예: `commands.txt`)로 계산한다.

N9. 품질 게이트가 M3 편집이 실제로 걸리는 가드를 돌리지 않는다 — acceptance.md:L127, L153–L158, L192 — M3 은 상시 로드 파일(`kanban-dispatch.md`, `paths:` 없음)과 `AGENTS.md`/`AGENTS.md.tmpl` 에 문장을 더한다. 이 둘을 묶는 가드는 `./internal/config` 의 `TestCodexContractByteCeiling`(`token_budget_guard_test.go:L84`, AGENTS 두 사본 바이트 상한)과 `TestAlwaysLoadedTokenBudget`(`L54`)이다. SPEC 은 `./internal/template` 의 두 테스트만 돌린다. t1175 가 바로 상시 로드 다이어트라 흡수 뒤 여유가 줄어 있을 수 있다. 상시 로드 예산 테스트가 kanban-dispatch.md 를 실제로 세는지는 이번에 열거하지 않았다(가드 존재만 확인). — Severity: major — Class: blocking — Required fix: AC-016 에 `go test ./internal/config/ -run '^(TestCodexContractByteCeiling|TestAlwaysLoadedTokenBudget)$' -count=1 -v` 를 추가하고, 출력의 `--- PASS:` 두 줄로 스윕이 비지 않았음을 확인한다.

N10. SessionStart 훅의 primary 쪽 쓰기가 REQ-006 허용 범위를 넘는데 선언되지 않았다 — spec.md:L77; plan.md:L59 — `--setting-sources` 와 무관하게 추적되는 프로젝트 `settings.json` 의 SessionStart 는 돈다. 그 처리기는 `.claude/settings.local.json` 쓰기 체인, 세션 레지스트리, `.claude/skills/` 진화 스킬 심볼릭 링크, 마이그레이션 파일을 건드린다(`internal/hook/session_start.go:L224–L279`; codex #1, 이번 감사가 코드를 읽어 재확인). 워크트리의 `settings.local.json` 이 `{"teammateMode":"in-process"}` 33 B 인 것도 이 훅이 쓴 흔적과 맞는다. REQ-006 은 primary 의 `.moai/` 런타임 상태만 허용한다. 멱등이라 실제 바이트 변화가 없을 수는 있으나 확인되지 않았다. — Severity: minor — Class: blocking — Required fix: primary 시작 프로브 전후로 `<primary>/.claude/settings.local.json` 의 sha256 과 `<primary>/.claude/skills/` 목록을 E 에 남기고, 달라지면 선언된 부수 효과로 기록한다. REQ-006 허용 목록을 그에 맞게 고친다.

N11. AC-001·002 의 순서 검사가 빈 뼈대와 같은 커밋을 통과시킨다 — acceptance.md:L35, L42 — E 가 처음 **추가된** 커밋만 보므로 빈 E 를 먼저 커밋하고 교리를 고친 뒤 측정을 채워도 통과한다. E 와 교리가 한 커밋이면 `merge-base --is-ancestor X X` 도 exit 0 이다(codex #5, 재확인). 커밋 조상 관계는 캡이 프로브 실행보다 먼저였다는 것도 증명하지 않는다. — Severity: minor — Class: blocking — Required fix: `git show "$FIRST^:.moai/reports/t1219/m1-measure.md" | grep -c '^decision_primary: '` 가 `1` 이고 E 커밋 ≠ `FIRST` 임을 요구한다. 캡은 `git log -1 --format=%ct $C` 가 첫 debug 로그 줄의 시각보다 이르다는 것을 E 에 기록한다.

N12. 개수 검사가 서로 다른 키를 세지 않는다 — acceptance.md:L47–L49, L66, L70 — AC-003 은 줄 수 15 를 세므로 한 필드를 15번 반복해도 통과한다. AC-006 의 브랜치 검사는 `before` 한 줄만 있어도 고유값 1 이 된다. pgid 중복도 걸러지지 않는다(codex #4·#6, 재확인). — Severity: minor — Class: optional — Required fix: `grep -oE '^(skills_paths|instruction_paths|prompt_tokens)_(A|B|C|D|Dp):' E | sort -u | wc -l` → 15, `primary_branch_before`·`after` 각 1줄, pgid `sort -u` 개수 = `probes_run`.

N13. 수용 기준의 셸 토큰이 경로로 풀리지 않는다 — acceptance.md:L11–L20 — 이름표는 `T`/`L`/`DT`/`DL` 를 "template / local `kanban-dispatch.md`" 로만 적어, `cmp T L` 이나 `git log … -- T L …` 를 그대로 치면 `No such file` 또는 빈 결과가 나온다(codex #3, 재확인). 빈 `FIRST` 는 N/A 로 읽혀 공허 통과로 이어질 수 있다. — Severity: minor — Class: optional — Required fix: 표 아래에 `T=internal/template/…/kanban-dispatch.md` 식 대입 블록을 붙이거나 표에 전체 경로를 적는다.

N14. 프로브가 이 카드의 유일본 트리에 들어간다 — plan.md:L68, L70, L72 — #3·#5·#7 은 primary 에서 시작한 세션이 t1219 로 `EnterWorktree` 한다. 이 트리는 미푸시 브랜치와 아직 커밋되지 않은 프로브 출력을 담는다. headless 세션이 끝날 때 들어간 워크트리를 어떻게 처리하는지는 확인하지 않았다(미검증). plan 에 보호 장치가 없다. — Severity: minor — Class: blocking — Required fix: primary 시작 프로브마다 직전에 커밋해 두고, 직후 `git worktree list` 에 t1219 가 남아 있는지와 HEAD 불변을 E 에 기록한다.

N15. 서브에이전트 귀속이 설계상 자기 보고로 떨어진다 — plan.md:L52–L56 — `rule-load-audit.jsonl` 행에는 에이전트 식별 필드가 없다(`instructions_loaded.go:L122–L127`). 서브에이전트는 부모와 같은 session_id 를 쓸 가능성이 높아 D/D′ 의 지침 행이 부모 행과 섞인다. 스킬 쪽은 디버그 로그의 `Loading skills from` 줄이 에이전트를 구분하지 않아 자기 보고 대체 수단조차 정해져 있지 않다. — Severity: minor — Class: optional — Required fix: D/D′ 는 처음부터 `source: self-report` 로 계획하고, `skills_paths_D`/`Dp` 는 귀속 불가하면 `gap` 으로 둔다고 명시한다.

N16. `timeout` 에 강제 종료가 없다 — plan.md:L42 — GNU `timeout` 은 기본으로 TERM 만 보낸다. TERM 을 무시하는 자식이 남으면 AC-006 의 pgid 검사가 뒤늦게 붙잡는다. — Severity: minor — Class: optional — Required fix: `timeout -k 10 300`.

## 점검 축별 소견 (오케스트레이터 지정)

- **AC 기계 판정성:** 대부분 명령으로 적혔으나, 이름표-경로 미해소(N13), 거짓 FAIL 확정(N4), 거짓 PASS 가능(N3, N7, N12)이 남았다.
- **0 결과의 양성 대조:** AC-009·010·011·013 은 BASE 기준값을 적었고 이번에 모두 재현됐다(`0`/`0`/`1`/`1`·`0`). AC-012 는 verdict.md 대조, AC-015 는 SPEC 디렉터리 대조가 있다. 빈 곳은 AC-007 발동 조건(N6)과, 무엇보다 N1 — 대조 자체가 플래그 때문에 0을 낼 수 있다는 점이다.
- **캡 선커밋:** REQ-003·AC-002 로 들어왔다. 커밋 순서 ≠ 실행 순서(N11).
- **운영자 결정:** 두 단계 진입은 REQ-008/009 로, 프로브 트리는 §F·§D 로 반영됐다. t1175 뒤 M3/M4 는 REQ-016 으로 적혔으나 술어가 이미 참이라 지켜지지 않는다(N3). 레인 표준의 근거인 #7 은 실행 형태가 없다(N5). 두 단계 구조의 전제 자체가 N2 로 흔들린다 — 운영자에게 다시 올릴 사안이다.
- **템플릿 중립성:** REQ-013·AC-012 가 추가 줄의 카드 id·SPEC id·날짜·해시를 막고, 양성 대조와 스윕 비공백 검사가 있다. M3 문안 계획(`plan.md:L120`)도 일반 자리표시자만 쓴다. 결함 없음.
- **REQ↔AC 추적:** 16:16 정합. 선택 사항으로 REQ-012 "템플릿 먼저" 순서 검사가 없다.
- **`--setting-sources user,project` 와 메모리 로드:** 작성자 가정은 **틀렸다**. 실측으로 CLAUDE.local.md 가 빠진다(N1).

## Regression Check (1회차 결함)

- D1, D2, D3, D5, D7, D8, D9, D11, D12, D13, D14, D15, D16, D17, D18, D20 — RESOLVED (근거는 위 표).
- D4 — PARTIALLY RESOLVED: 양성 존재 검사 누락(N6 로 이관).
- D6 — PARTIALLY RESOLVED: 대체 검사가 거짓 FAIL(N4).
- D10 — UNRESOLVED: (c) 의 해법이 측정 대상을 지운다(N1).
- D19 — 문장은 RESOLVED, 전제는 반증(N2).
- 정체(3회 연속 무변) 판정 대상 없음 — 이번이 2회차이자 Tier M 마지막 회차다.

## Recommendation

판정 FAIL. Tier M 상한(2회)에 도달했으므로 오케스트레이터는 추가 반복 대신 사용자에게 세 갈래를 제시해야 한다: 범위 축소 / 부채를 기록한 PASS / 명시적 연장.

우선순위 순 수정 지시(연장 또는 부채 수용 시):

1. **N2 를 먼저 운영자에게 올린다.** 경로 B(워크트리 안에서 시작)도 primary CLAUDE.local.md 를 싣는다는 기록이 이미 트리에 있다. 두 단계 구조의 전제가 흔들리므로, 측정 설계(경로 E 추가 여부)와 결정표 새 행을 운영자 결정으로 확정한다. verdict.md Evidence 1 을 정정한다.
2. **N1:** `--setting-sources` 를 프로브에서 빼고, LSEL 부수 효과는 선언·전후 기록으로 다룬다. 사전 점검에 두 표식 파일 스크래치 프로브를 넣는다.
3. **N3:** Gate G1 과 AC-001 을 t1175 착지 커밋 SHA 기반 조상 판정으로 바꾸고, 지금 exit 1 을 대조로 기록한다.
4. **N4·N5·N6·N7:** AC-006 대상에서 primary 제외, #6·#7 명령 줄 확정과 AC-006 환경 정리 검사 완화, AC-007 발동 조건 기계화와 양성 존재 검사, AC-008 결정의 역방향 조건.
5. **N9:** AC-016 에 `./internal/config` 의 두 예산 가드를 추가한다.
6. **N8·N10·N11·N14:** `git add -f` 와 원시 로그 비커밋, SessionStart 부수 효과 전후 해시, 순서 검사 강화, 트리 보존 확인.
7. 선택 사항 N12·N13·N15·N16 은 오케스트레이터 재량.

## Gaps (이번 감사가 관측하지 않은 것)

- `79ed0553` 이 verdict Evidence 1 의 프로브라는 것은 시각·cwd·entrypoint 로 추정했다. 세션 id 가 verdict 에 적혀 있지 않다.
- primary CLAUDE.local.md 가 상위 디렉터리 탐색 때문에 붙었다는 원인은 추정이다. 로드 사실만 관측했다. 왜 primary `CLAUDE.md` 는 같은 세션에 실리지 않았는지도 확인하지 않았다.
- `/clear` 후 session_id 변경, `moai cc -w` 의 인자 전달, headless 세션 종료 시 워크트리 처리 — 확인하지 않았다(N5, N14).
- 상시 로드 예산 테스트가 kanban-dispatch.md 를 세는지 열거하지 않았다(N9).
- 원시 debug/stream-json 에 민감 정보가 들어가는지 확인하지 않았다(N8).
- 설치 바이너리는 `a8a9b9376`(HEAD 의 조상)이다. spec lint 는 트리 소스(`go run`)로도 재실행해 같은 결과를 얻었다.
- codex 지적 중 CLAUDE.local.md 의 테스트 경로 일반화(#9)는 이 SPEC 범위 밖(§E 첫 항목)이라 채택하지 않았다.

## Residual-risk

- 이번 감사가 돌린 스크래치 프로브 두 번(`2d5df932`, `bbef73ec`)은 haiku, 1턴, 비워크트리 디렉터리였다. 다른 모델이나 워크트리 안에서 `--setting-sources` 동작이 다를 가능성은 낮지만 배제하지 않았다.
- N2 가 사실이면 이 SPEC 의 교리 변경은 스킬 목록 중복만 줄이고 지침 파일 중복은 그대로 둘 수 있다. 토큰 비용이 어느 쪽에 더 큰지는 여전히 재지 않았다.
- 190개 워크트리가 살아 있는 환경이라, 다른 레인이 같은 시간대에 primary 의 공유 로그와 레지스트리에 쓴다. session_id 귀속은 그 섞임을 걸러 내지만 session_id 를 담지 않는 쓰기는 보지 못한다.
