# SPEC Review Report: SPEC-SESSION-DOUBLELOAD-001
Iteration: 1/2 (Tier M 상한 2)
Verdict: FAIL
Overall Score: 0.64 (Tier M 통과 기준 0.80 미달)

- 감사 대상: 워크트리 `.claude/worktrees/t1219`, 커밋 `2fdd1f8c9` (`git rev-parse --short HEAD` → `2fdd1f8c9`, 작업 트리 clean)
- 읽은 아티팩트: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`, 전제 원천 `.moai/reports/t1219/verdict.md` 와 `probe-*.txt`
- 작성자 추론 맥락은 받지 않았다(M1 맥락 격리). 작성자가 표시한 미결 사항은 감사 범위 지정으로만 썼다.
- 교차 모델 감사: `mcp__moai__audit_multi` 호출 결과 codex 는 사용량 한도(`try again at Sep 28th`)로, GLM 은 리뷰 대상 해석 실패로 모두 `inconclusive`. 참여 백엔드는 Claude 하나뿐이다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `spec.md:L46–L66` 에 REQ-SDL-001..015 가 빈칸·중복 없이 3자리로 이어진다. AC 도 `acceptance.md:L15–L39` 에 AC-SDL-001..016 이 연속이다.
- [PASS] MP-2 GEARS 형식 (요구사항 층 기준): 15개 REQ 모두 GEARS 다섯 패턴 중 하나다. 예: `L46` "The measurement milestone shall …"(Ubiquitous), `L49` "When a cap is reached …, the measurement shall …"(Event), `L51` "shall not change …"(Unwanted), `L56` "Where the measurement shows …, the dispatch doctrine shall …"(Where). AC 의 Given-When-Then 은 검증 층이라 여기서 채점하지 않았다.
- [PASS] MP-3 프런트매터: `spec.md:L2–L13` 에 12개 필수 필드가 모두 있고 타입이 맞다(`version: "0.1.0"` 따옴표, `created/updated: 2026-09-26`, `priority: P2`, `lifecycle: spec-anchored`, `tags` 쉼표 문자열). 거부 별칭(`created_at` 등) 없음. `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` → exit 0, "No findings".
- [N/A] MP-4 언어 중립성: 다중 프로그래밍 언어 도구를 다루는 SPEC 이 아니다. 템플릿 본문 중립성은 REQ-SDL-013 / AC-SDL-012 로 따로 다룬다.
- [PASS] MP-5 D7: 본문의 SPEC 참조는 `SPEC-SESSION-WORKTREE-001` 하나이고 `status: completed` 다. retired/superseded/archived 아님 → BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall` 네 파일 모두 0 → 자동 통과.
- [FAIL] MP-7 명확화 게이트: `plan.md:L88–L90` 에 미해결 `[NEEDS CLARIFICATION]` 3건(lane lifecycle, rules-diet ordering, probe tree). `progress.md:L8` 도 "3 markers open" 으로 스스로 적고 있다. 점수와 무관하게 FAIL 이며, 오케스트레이터가 AskUserQuestion 으로 풀어야 한다.

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.60 | 0.50–0.75 사이 | REQ-SDL-008(`L56`) 의 조건과 plan 의 B-대리 허용이 어긋남(D3). REQ-011 과 REQ-015 사이 "exit-first" 해석이 열려 있음(D17). 토큰 수치의 "한 턴" 정의가 경로 C 의 4턴과 맞지 않음(D5) |
| Completeness | 0.75 | 0.75 | 섹션·프런트매터·`### Out of Scope — …` H3(`L79` 등)는 모두 있음. 다만 §D 영향 표면이 같은 절차를 규정하는 문서를 빠뜨림(D2), 서브에이전트 경로 부재(D11) |
| Testability | 0.55 | 0.50 쪽 | AC-001(D7), 002(D4), 006(D6), 007(D8), 010(D15), 013(D14) 이 기계 판정은 되지만 요구와 다른 것을 재거나 오판 가능 |
| Traceability | 0.70 | 0.75 아래 | REQ-SDL-010 을 검사하는 AC 없음(D9). AC-015→REQ-015, AC-016→REQ-012 매핑이 실제 검사 내용과 다름(D16) |

## 전제 검증 (verdict.md 대조)

- `spec.md:L30` "project skills directory 하나" → `probe-skills-dirs.txt` 의 `project=[/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1219/.claude/skills]` 로 뒷받침된다.
- `spec.md:L30` "worktree copy of CLAUDE.local.md" → `probe-memory-files.txt` 는 잘린 훅 systemMessage 한 줄(워크트리 사본이 40,000자 예산을 넘는다는 경고)이다. 워크트리 사본이 실렸다는 증거는 되지만, primary 사본이 실리지 않았다는 증거(양성 대조 포함)는 아니다(D19). M1 이 B 경로를 다시 재므로 선택 사항으로 둔다.
- `spec.md:L31` 경로 C 관측은 작성자 세션 컨텍스트에서 본 것이라 파일 증거가 없다. SPEC 이 이를 M1 재측정 대상으로 삼은 것은 적절하다.
- `spec.md:L34` 의 "Not measured" 목록은 verdict § Gaps 와 일치한다.

## 현재 트리 기준선 측정 (이번 감사에서 실행)

| 검사 | 명령 | 관측 |
|---|---|---|
| T/L 바이트 동일 | `cmp T L` | exit 0 |
| AC-009 음성 대조 | `grep 'wt. names the new card' T \| grep -c 'moai cc -w'` | `0` |
| AC-012 중립성 | `grep -nE '\bt[0-9]{3,4}\b\|SPEC-…\|20..-..-..' T` | 출력 없음, exit 1 |
| AC-012 양성 대조 | 같은 grep 을 `verdict.md` 에 | `6` |
| AC-012 go test 비공허성 | `internal/template/*_test.go` 의 `Neutral\|Leak` 테스트 | `TestLanguageNeutrality`, `TestTemplateNoInternalContentLeak` 등 10개 존재 |
| AC-014 기준 | `grep -c '^\[HARD\]' T` | `32` (전체 `[HARD]` 38) |
| AC-014 앵커 | `WT-<slug>` / `exit any previous one first` / `entered through the launcher` | L93·L181, L179, L161 에 각각 존재 |
| AC-010 기준 | `grep -c '/clear' T`, `grep -ci relaunch T` | `7`, `0` |
| AC-013 기준 | `G:L20` | `moai cc -w <card-id>` 가 이미 있고, 뒤 3줄에 `/clear`·`relaunch` 없음 |

## Defects Found

D1. MP-7 — plan.md:L88–L90 — 미해결 `[NEEDS CLARIFICATION]` 3건. — Severity: critical — Class: blocking — Required fix: 오케스트레이터가 AskUserQuestion 으로 세 항목(레인 수명, t1175 순서, 프로브 트리)을 풀고, 답을 plan.md §I 에 결정 문장으로 바꿔 적는다. 아래 "작성자 미결 사항 평가"의 권고를 선택지로 쓸 수 있다.

D2. 영향 표면 누락 — spec.md:L68–L73 — §D 가 kanban-dispatch.md 두 사본과 lane protocol 만 적는다. 같은 세션 내 진입 절차를 규정하는 문서가 더 있다: `AGENTS.md:L127`("Start a new card in a new worktree. Exit any previous worktree back to the primary checkout first") 과 템플릿 `internal/template/templates/AGENTS.md.tmpl:L136`, `kanban-dispatch-detail.md:L28`(dispatch 예시 `wt: EnterWorktree(t0)`) 과 그 템플릿 사본, `worktree-integration.md:L224`(현재 세션 진입 안내). M3 이후 AGENTS.md 는 Codex 쪽 상시 계약이라 두 문서가 서로 다른 진입 절차를 지시하게 된다. — Severity: major — Class: blocking — Required fix: 이 표면들을 §D 와 REQ-SDL-011/012 범위에 넣거나, 넣지 않을 표면은 §E 에 이유를 붙여 명시적으로 제외한다. AGENTS.md 는 t1175 도 고치고 있으므로 D1 의 순서 결정에 포함한다.

D3. 요구와 계획의 조건 불일치 — spec.md:L56 vs plan.md:L46, acceptance.md:L65 — REQ-SDL-008 은 "경로 A(런처)가 단일 로드를 보일 때"만 런처를 주 진입 형태로 정한다. 그런데 plan 과 edge case 는 A 가 Gap 이면 B 를 대리로 써서 `decision_primary: launcher` 를 허용한다. 이대로면 REQ 조건 없이 교리가 바뀐다. — Severity: major — Class: blocking — Required fix: REQ-008 을 "경로 A, 또는 A 가 Gap 일 때 대리로 선언한 경로 B" 로 고치고, E 에 `path_a_proxy: none|B` 줄을 두어 AC-SDL-008 이 그 값과 결정의 일관성을 확인하게 한다.

D4. 집합 대신 개수 — acceptance.md:L16 — REQ-SDL-002 는 스킬 디렉터리와 지침 파일 경로의 **집합**을 기록하라고 하지만, AC-SDL-002 정규식은 `([0-9]+|gap)` 숫자만 받는다. 경로 C 에서 `1` 이 나와도 그것이 primary 사본인지 워크트리 사본인지 가릴 수 없다. 바로 이 구분이 SPEC 의 핵심 질문이다. — Severity: major — Class: blocking — Required fix: 경로별 `*_paths_<X>: <절대경로 목록>` 줄을 추가하고, AC 가 A/B 목록에 `.claude/worktrees/` 가 있고 primary 경로가 없음을 확인하게 한다.

D5. 토큰 수치 정의 — spec.md:L47, plan.md:L43–L44 — REQ 는 "고정 프로브 한 턴의 프롬프트 토큰"이라 하는데, 경로 C 는 최대 4턴(`--max-turns 4`)이고 plan 은 실행 결과의 `usage` 합산식을 적는다. 결과 JSON 의 `usage` 가 턴 누적인지 이번 감사에서 확인하지 않았다(미검증). 누적이라면 B(1턴)와 C(4턴)는 비교할 수 없는 수치가 된다. — Severity: major — Class: blocking — Required fix: 어느 턴의 수치인지(예: 첫 턴, 또는 이동 뒤 첫 턴) 명시하고, 턴별 usage 를 주는 출력 형식(stream-json 등)에서 뽑도록 plan 을 고친다. 수치를 쓰는 결정이 없으므로, 대안으로 이 필드를 참고값으로 내려도 된다.

D6. 이름으로 프로세스 판정 — acceptance.md:L20 — `pgrep -f 'claude -p'` 는 공유 머신에서 다른 세션의 `claude -p`(다른 레인, 교차 감사 백엔드 등)까지 잡아 거짓 FAIL 을 낸다. 반대로 timeout 으로 종료된 claude 의 자식(MCP 서버 프로세스 등)은 이 패턴에 걸리지 않을 수 있다. REQ-SDL-006 의 "다른 카드 워크트리에 쓰지 않는다"는 부분은 AC 가 아예 없다. — Severity: major — Class: blocking — Required fix: 프로브마다 PID 를 기록하고, 그 PID 와 프로세스 그룹에 대해서만 생존 여부를 확인한다. 다른 워크트리 비쓰기는 M1 전후 `git worktree list` 대상 트리들의 `status --porcelain` 비교 같은 명령으로 확인한다.

D7. 순서 검사 명령 오류 — acceptance.md:L15 — 문장은 "T 를 건드린 첫 커밋"인데 명령은 `git log -1 -- T`(가장 최근 커밋)다. 검사 대상도 T 뿐이라 REQ-SDL-001 이 묶는 L·G 편집 선행은 잡지 못한다. 결정이 `none` 이라 M3 가 없으면 `git log -1 -- T` 가 develop 의 옛 커밋을 돌려주고, `merge-base --is-ancestor <E> <옛 커밋>` 이 실패해 REQ 를 지켰는데도 AC 가 FAIL 한다. — Severity: major — Class: blocking — Required fix: `git log --reverse --format=%H $(git merge-base develop HEAD)..HEAD -- T L G | head -1` 로 이 브랜치에서 처음 교리를 건드린 커밋을 구하고, 비어 있으면 N/A 로 판정한다.

D8. 양성 대조의 모양이 다름 — acceptance.md:L21, plan.md:L45 — 판정하려는 0은 "두 번째 스킬 디렉터리가 없다"이다. 그런데 계획한 양성 대조(이동 없는 primary 시작 프로브가 primary 경로를 보임)는 필드가 파싱된다는 것만 보여 주고, 두 번째 디렉터리가 **있을 때** 추출기가 그것을 보는지는 보여 주지 못한다. 세션 중 이동 뒤 추가되는 스킬이 `Loading skills from` 줄로 로그에 남는지도 확인되지 않았다(미검증). 또 AC 문장은 "프로브 파일을 명시"라고 하지만 정규식 `^positive_control: .*count=[1-9][0-9]*$` 는 그것을 검사하지 않는다. — Severity: major — Class: blocking — Required fix: 두 출처가 있다고 알려진 입력(두 `project=` 경로를 담은 고정 디버그 로그 사본, 또는 중첩 스킬 발견을 유발하는 프로브)에 같은 추출 명령을 돌려 `count=2` 를 보이게 한다. 정규식에 `.moai/reports/t1219/probes/` 경로를 요구한다.

D9. REQ-SDL-010 미추적 — spec.md:L58, acceptance.md:L25, L64 — 전제 반박 시 "해당 전제는 교리를 고치지 않고 E 에 보고한다"를 확인하는 AC 가 없다. edge case L64 의 부류별 결정(지침 파일만 중복, 스킬은 단일)도 기계 판독 필드가 없다. — Severity: major — Class: blocking — Required fix: AC 한도(Tier M 16개)가 이미 찼으므로 AC-SDL-008 에 합친다: `premise_instructions_C: dup|single|gap`, `premise_skills_C: dup|single|gap` 줄을 요구하고, `single` 인 부류는 교리 편집에서 빠졌는지 확인한다.

D10. 프로브의 부수 효과가 덜 적혔다 (미결 3) — plan.md:L21, spec.md:L51 — 이번 감사에서 확인한 것: primary 의 `.claude/settings.local.json` SessionStart 에 LSEL `session_drain.sh` 와 `backlog_check.sh` 가 `$CLAUDE_PROJECT_DIR` 기준으로 걸려 있다(워크트리의 settings.local.json 은 33 B 로 이 훅이 없다). 그래서 primary 에서 시작하는 프로브(경로 C, 양성 대조)는 공유 LSEL 상태와 인박스 오프셋을 건드린다. InstructionsLoaded 훅은 `resolveProjectRoot(input)` 아래 `.moai/logs/rule-load-audit.jsonl` 에 `session_id`·`file_path` 를 남긴다(`internal/hook/instructions_loaded.go:L59–L65, L144`). 이 감사 세션의 환경에는 `MOAI_KANBAN*` 변수가 3개 있어, 레인에서 띄운 `claude -p` 는 칸반 환경을 물려받는다. — Severity: major — Class: blocking — Required fix: (a) 모든 프로브를 한 번의 복합 호출로 실행한다: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 300 claude -p …`. (b) 지침 경로 추출 원천을 `rule-load-audit.jsonl` 에서 프로브 `session_id` 로 거른 행으로 못박는다(다른 세션 행이 섞이는 파일이다). (c) primary 시작 프로브가 LSEL 드레인을 돌린다는 것을 REQ-SDL-006 에 선언된 부수 효과로 적거나, local 설정 원천을 빼고 실행할 수 있는지 먼저 확인한다.

D11. 서브에이전트 경로 부재 (미결 6) — spec.md:L40, L47 — 이번 감사가 직접 관측한 사실: 이 plan-auditor 서브에이전트는 작업 디렉터리가 t1219 워크트리인데, 컨텍스트에 실린 지침 파일 머리말은 `/Users/goos/MoAI/moai-adk-go/CLAUDE.md`, `/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md`, `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/…` 로 모두 primary 경로다. 반면 에이전트 메모리 경로는 워크트리(`…/worktrees/t1219/.claude/agent-memory/plan-auditor/`)로 풀렸다. 즉 primary 에서 시작한 부모가 띄운 서브에이전트는 cwd 와 무관하게 부모의 지침 집합을 싣는 것으로 보인다. 한계: 컨텍스트 머리말이 로드 집합 전체인지는 확인하지 못했다. 레인의 실제 작업은 서브에이전트가 하므로, §B 목표("card sessions carry … the card worktree's")는 진입 형태만으로는 닫히지 않을 수 있다. — Severity: major — Class: blocking — Required fix: REQ-SDL-002 에 경로 D(워크트리에서 시작한 부모가 띄운 서브에이전트)와 D′(primary 에서 시작한 부모가 워크트리 cwd 로 띄운 서브에이전트)를 추가한다. 캡 8회 안에 들어간다(B, C, 대조, A, clear, D, D′ = 7). 추출은 InstructionsLoaded 행이 서브에이전트에도 발화하는지부터 확인하고, 발화하지 않으면 서브에이전트에게 컨텍스트의 지침 머리말 경로를 그대로 출력하게 한다(자기 보고라는 한계를 E 에 적는다).

D12. 대체 경로와 기존 `/clear` 인계 순서의 충돌 (미결 1·4) — spec.md:L57 — 현재 교리는 페이즈가 끝나면 운영자가 레인을 `/clear` 한 **뒤** 리드가 이동(`wt:`)을 담은 dispatch 를 보낸다. `/clear` 가 이동보다 먼저면 이동 뒤 다시 이중 로드가 된다. REQ-009 의 "이동 직후 `/clear`"는 인계 순서를 "이동 → 운영자 `/clear` → 리드가 포인터 재전송"으로 바꿔야 성립하는데, SPEC 은 kanban-dispatch.md § The `/clear` handoff between phases 를 편집 범위(REQ-011)에도 보존 목록(REQ-015)에도 넣지 않았다. — Severity: major — Class: blocking — Required fix: 그 절을 REQ-011 범위에 넣고, 대체 경로가 채택되면 순서를 위와 같이 적도록 요구한다. 이 경우 레인당 추가 비용은 없다(카드마다 하던 `/clear` 한 번의 위치만 옮긴다).

D13. `/clear` 측정 수단 미지정 (미결 4) — plan.md:L47 — "캡 안에서 대화형 세션으로 시도"라고만 한다. 서브에이전트는 대화형 TUI 에 입력할 수 없고, `-p` 모드에서 `/clear` 가 동작하는지는 확인되지 않았다. tmux send-keys 로 구동하거나 운영자 참여가 필요하다. edge case L66 이 gap→relaunch 로 이어지므로 치명적이진 않다. — Severity: minor — Class: optional — Required fix: 수단(tmux 구동 + `timeout`, 또는 운영자 수동 1회)을 plan 에 적거나, 처음부터 Gap 으로 선언해 프로브 한 회를 아낀다. 권고: D12 의 순서 변경을 채택하면 `/clear` 결과가 교리 전체를 좌우하므로, 운영자 수동 1회 측정을 권한다.

D14. AC-SDL-013 이 주 진입 형태를 확인하지 않음 — acceptance.md:L33 — `G:L20` 에는 이미 `moai cc -w <card-id>` 가 있고, `EnterWorktree(<card-id>)` 가 동등한 선택지로 함께 적혀 있다. AC 는 대체 경로 문자열만 확인하므로, 동등 선택지 문장이 그대로 남아도 통과한다. 현재 기준값 0(음성 대조)도 기록되지 않았다. — Severity: minor — Class: blocking — Required fix: 새 카드 진입 문장에서 `EnterWorktree(<card-id>)` 가 동등 선택지로 남지 않았는지 확인하는 조건과 기준값 0 을 AC 에 넣는다.

D15. AC-SDL-010 범위가 파일 전체 — acceptance.md:L30 — T 에는 이미 `/clear` 가 7번 나온다. 파일 어디에 한 번만 더 써도 통과하고, 비교 기준 `develop` 은 움직이는 참조다. — Severity: minor — Class: optional — Required fix: `wt` 줄과 새 카드 절 안으로 범위를 좁히고, 기준을 `git merge-base develop HEAD` 로 고정한다.

D16. 추적표 오매핑 — acceptance.md:L59–L60 — AC-SDL-015 는 §E 의 "Go 코드 없음" 제외를 검사하는데 REQ-SDL-015(HARD 보존)로 매핑했다. REQ-015 의 실제 검사는 AC-014 다. AC-016(lint + build)을 REQ-012 로 매핑한 것도 간접적이다. — Severity: minor — Class: blocking — Required fix: AC-015 는 "§E Out of Scope — Runtime or CLI changes"로, AC-016 은 "품질 게이트"로 표기를 고친다.

D17. REQ-011 과 REQ-015 의 해석 틈 — spec.md:L62, L66 — REQ-011 은 세션 내 `ExitWorktree → EnterWorktree` 절차를 대체하라 하고, REQ-015 는 "exit any previous one first" HARD 절을 약화하지 말라 한다. `T:L179` 본문은 바로 그 세션 내 이동을 규정한다. 런처 방식에서 "exit" 가 무엇을 뜻하는지(이전 카드 세션 종료, 대체 경로에서는 ExitWorktree) SPEC 이 정하지 않았다. — Severity: minor — Class: blocking — Required fix: REQ-015 에 "exit-first 는 주 진입 형태에서 이전 카드 세션을 끝내는 것, 대체 경로에서 `ExitWorktree` 로 읽는다"를 한 문장으로 적는다.

D18. 캡 선언 시점 검증 — acceptance.md:L17 — 파일 안 줄 순서는 캡이 첫 프로브 **전에** 쓰였다는 것을 증명하지 못한다. — Severity: minor — Class: optional — Required fix: 캡 블록을 첫 프로브 전에 별도 커밋하거나, 첫 프로브 출력 파일보다 앞선 타임스탬프를 E 에 남긴다.

D19. 전제 증거의 과대 표현 — spec.md:L30 — 위 "전제 검증" 참조. `probe-memory-files.txt` 는 경로 열거가 아니다. — Severity: minor — Class: optional — Required fix: "worktree copy observed; absence of the primary copy not established" 로 문장을 낮추거나, M1 경로 B 결과로 대체한다.

D20. plan.md 절 번호 — plan.md:L29–L36 — §D 다음이 §F 로 §E 가 없다. — Severity: minor — Class: optional — Required fix: 번호를 고치거나 그대로 둔다.

## 작성자 미결 사항 평가

1. **장수 레인의 카드별 재기동 비용.** 칸반 companion 과 팩토리 레인은 운영자가 터미널마다 손으로 한 번 띄운다. 런처를 주 형태로 삼으면 카드마다 운영자가 레인을 다시 띄워야 하고, `moai cc -w <name>` 으로 다시 띄운 세션이 `lane-k` 같은 역할로 보드에 다시 붙는지는 이번 감사에서 확인하지 않았다(미검증). 반면 대체 경로는 이미 카드마다 하던 `/clear` 의 위치만 옮기면 추가 비용이 없다(D12). 권고: 새로 여는 세션은 런처 시작을 주 형태로, 장수 레인은 "이동 → `/clear` → 포인터 재전송"을 표준으로 두는 이원 구조. 단 `/clear` 가 단일 로드를 복원한다는 측정(REQ-005)이 선행 조건이다. 운영자 결정 사항이므로 D1 로 올린다.
2. **t1175 와의 순서.** `git diff --stat develop...WT-rules-diet` 결과 t1175 는 kanban-dispatch.md 두 사본(각 78줄)과 AGENTS.md(76줄)를 고친다. `wt` 줄과 새 카드 절 자체는 그 diff 에 없고, t1175 판에서도 앵커가 살아 있다(`^\[HARD\]` 32, `exit any previous one first` 1, `entered through the launcher` 1, `wt. names the new card` 1). 권고: M1·M2 는 지금 진행하고, M3·M4 는 t1175 가 develop 에 착지한 뒤 develop 을 흡수하고 나서 한다. D2 대로 AGENTS.md 를 범위에 넣는다면 t1175 뒤 순서는 필수다.
3. **프로브 실행 트리.** 경로 B·A 는 워크트리 안에서 시작해야 하므로 t1219 트리(또는 런처로 만든 별도 프로브 트리)에 `.moai/state`·`.moai/logs` 쓰기가 생기는 것은 피할 수 없다. 둘 다 추적되지 않는 상태라 t1219 자체에서 돌려도 무방하다고 본다. 실제 위험은 primary 쪽이다(D10: LSEL 드레인, 세션 레지스트리 등록, 칸반 환경 상속). 별도 프로브 트리를 만들면 폐기 절차(L1 트리 정리)를 plan 에 적어야 한다.
4. **headless 에서 `/clear` 측정 불가 가능성.** D13 참조. gap→relaunch 로 이어지는 edge case 는 안전하지만, 1번 권고(장수 레인의 대체 경로 표준화)가 이 측정에 기대므로 운영자 수동 1회 측정을 권한다.
6. **서브에이전트가 부모의 primary 지침을 싣는 관측.** 이번 감사 세션이 같은 현상을 보였다(D11). M1 에 경로 D·D′ 를 추가해야 한다. 런처로 레인을 워크트리 안에서 시작하면 서브에이전트도 워크트리 집합을 받을 것이라는 가정이 이 SPEC 결론의 전제인데, 지금은 아무도 재지 않았다.

## Regression Check

해당 없음 (1회차).

## Recommendation

1. 오케스트레이터: D1 의 세 항목을 AskUserQuestion 으로 해결한다(선택지는 위 평가 1·2·3 참조).
2. manager-spec: D2(§D 확대: AGENTS.md·템플릿, kanban-dispatch-detail.md·템플릿, 필요시 worktree-integration.md), D3(REQ-008 조건), D11(경로 D·D′), D12(`/clear` 인계 절을 범위에 포함), D17(exit-first 해석)을 spec.md 에 반영한다.
3. manager-spec: acceptance.md 에서 D4(경로 집합), D6(PID 기반), D7(첫 커밋·N/A), D8(같은 모양의 양성 대조), D9(AC-008 에 부류별 필드), D14, D16 을 고친다. AC 수는 16을 넘기지 않는다(Tier M 한도).
4. manager-spec: plan.md 에서 D5(토큰 턴 정의), D10(환경 정리 복합 호출·추출 원천·부수 효과 선언)을 고친다.
5. 선택 사항 D13, D15, D18, D19, D20 은 오케스트레이터 재량이다.

## Gaps (이번 감사가 관측하지 않은 것)

- `claude -p --output-format json` 의 `usage` 가 턴 누적인지 확인하지 않았다(D5).
- 세션 중 이동 뒤 스킬 추가가 디버그 로그에 어떤 모양으로 남는지 확인하지 않았다(D8).
- `moai cc -w <name>` 재기동 세션이 칸반/팩토리 역할로 다시 붙는지 확인하지 않았다(평가 1).
- `moai spec lint` 는 설치 바이너리 `a8a9b9376`(이 트리 HEAD 의 조상)로 실행했다. `audit_multi` 가 바이너리 지연을 경고했다.
- 서브에이전트 관측(D11)은 이 감사 세션 한 번의 컨텍스트 머리말에 근거하며, 머리말이 로드 집합 전체라는 보장은 없다.
- codex·GLM 교차 감사는 둘 다 inconclusive 여서 독립 2차 의견이 없다.

## Residual-risk

- D2 의 누락 표면 목록은 `EnterWorktree` 문자열 grep 으로 찾았다. 다른 표현으로 같은 절차를 적은 문서가 더 있을 수 있다.
- t1175 가 착지 전에 추가로 바뀌면 평가 2 의 앵커 생존 판단은 다시 재야 한다.
