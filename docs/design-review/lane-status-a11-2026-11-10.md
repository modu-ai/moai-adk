# A11 레인 상태 표·유휴 원인·이벤트 처리 — 조사 기록 (#2265)

- 작성: agent-4-t63, claim `2265#6763634562`(epoch 2; epoch 1 `2265#6732746936` agent-8-45deae, 인계 6763628473). 2026-11-10.
- 범위: Issue #2265 본문의 `lane-status` op(읽기 전용), 유휴 원인 분류, 이벤트 중복·역순·유실 판정, 배정 제안 초안. 호스트 세션 이벤트 연동(#A12 결론 뒤)과 실호스트·실제 HOME 변경은 범위 밖이다.
- 이전 보유자의 worktree `.claude/worktrees/2265-rc2-a11-leader-lane-status`는 Go `main` 기반(`4fa2aa9db2`)이고, 깨끗하며, A11 작업이 0이었다(`Cargo.toml` 없음, 프로세스 0). 이어 쓸 것이 없으므로 손대지 않고(`.claude/`는 보존 상태), 현재 tip `01b92a2257`에서 `issue/2265-a11`을 새로 시작했다.

## 1. 읽은 자료

| 자료 | 무엇을 확인했나 | 채택 |
|---|---|---|
| `docs/design-review/autonomy-modes-spec.md` §5 | 리더 감독 루프가 읽는 증거: 열린 PR(head·병합 상태·CI·verdict), claim 댓글, 살아 있는 세션. "목록이 비었는데 대기 중인 에이전트가 있으면 사유를 알린다" | 표의 열과 유휴 원인의 출처 |
| `AGENTS.local.md` § Leader supervision | "Stage changes come from the remote (branch, PR, status, merge), never from a lane message or an idle notice alone." 로컬 파일은 잠금이 아니며 소유는 claim 댓글이다 | phase는 원격·로컬 기록에서만 읽는다 |
| `src/claim_verify.rs` `classify_comment`·`evaluate`(A9 #2255) | claim·release·handoff 줄의 형식과 보유자 판정. 만료된 보유자도 보유자로 남는다(REQ-A9-03) | 소유권 열은 `evaluate`를 그대로 쓴다. `leader.rs::author`는 handoff·release를 보지 않으므로 쓰지 않는다 |
| `src/leader.rs`(A3 #2162) | PR 목록·`Refs #n`·`AUDIT … head … verdict:` 줄 읽기. 상태 파일을 쓴다 | verdict 읽기 규칙을 재사용한다. 이 op는 상태 파일을 쓰지 않는다 |
| `src/queue.rs`(A2 #2161) | ready 규칙: claim 줄이 하나라도 있으면 후보가 아님, `Blocked-By:`가 열린 Issue를 가리키면 제외 | ready 판정과 선행 대기 원인에 같은 규칙을 쓴다 |
| `src/forge/report.rs` | 소유 보고 댓글 형식: `<!-- moai:report run=… -->`, `상태: running·blocked·review·done` | 보고 표지가 있는 댓글만 phase 근거로 인정한다. 사용자 결정 대기를 나타내는 구조화된 필드는 없다 |
| `src/worktree.rs` | `git worktree list --porcelain`을 scrub된 git으로 읽는 파서 | 같은 파서를 재사용한다 |
| `docs/design-review/host-session-events-2026-11-09.md`(A12 #2263) §4 | 호스트 이벤트의 중복·역순·유실에 대한 공식 근거와 결론: 멱등 전이로 처리하고 원격 상태를 다시 조회한다 | 이벤트 판정을 매번 원격에서 다시 계산하는 방식의 근거 |
| 지시 원문 `ADK_AUTONOMOUS_LANE_LOOP_20261109_01` §2·§5 | 저장소에 원문이 없다(A12 문서는 §4만 인용) | **not_read**. Issue 본문의 요약만 근거로 삼는다 |

공식 문서(2026-11-10, curl GET, 응답 sha256):

| 출처 | 확인 시각(UTC) | 바이트·sha256 | 확인한 문장 |
|---|---|---|---|
| https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28 | 2026-11-10T09:24:58Z | 905,280 · `4bf8e7544de8e353dd58ce217a746e1ea4ec55ed47e4b82202203807e778ae9a` | "List commits on a pull request — Lists a maximum of 250 commits for a pull request." |
| https://docs.github.com/en/rest/issues/comments?apiVersion=2022-11-28 | 2026-11-10T09:24:59Z | 461,385 · `3c10480696cc0d1d7fe66e13a003d5aa0868464f83f5f8443fe4945b522ce862` | "Issue comments are ordered by ascending ID." |

## 2. 설계 결정

1. **읽기 전용.** GitHub에는 GET만 쓰고, 로컬에는 아무것도 쓰지 않는다(상태 파일 없음). 같은 원격 상태로 다시 부르면 같은 표가 나온다.
2. **행의 키는 (session, issue).** 한 세션이 Issue 두 개를 동시에 가질 수 있다(예: 감사 대기 PR + 새 카드).
3. **소유권은 `claim_verify::evaluate`에서 읽는다.** 보유 claim 댓글 id·epoch·lease·live를 그대로 싣는다. handoff로 닫혔는데 받는 세션의 claim이 아직 없으면 그 세션의 행을 phase `handoff`로 둔다.
4. **레인 이름은 호출자가 준다.** 모든 레인이 GitHub 계정 하나로 게시하므로 댓글 작성자로는 레인을 가릴 수 없다. 세션과 레인 이름의 대응은 요청의 `lanes[]`에서만 가져오고, 문자열 접두사로 추측하지 않는다.
5. **phase(5종)와 근거.** 각 행은 phase를 정한 기록을 `phase_evidence`에 싣는다.
   - `handoff`: handoff 줄(comment id)과, 받는 세션의 claim이 없음.
   - `waiting`: 열린 PR의 현재 head에 `verdict: PASS` 감사 줄(comment id). 통합 대기.
   - `pr_open`: 열린 PR(번호·head)에 현재 head를 가리키는 verdict가 없음.
   - `implementing`: 열린 PR이 없고, Issue 번호의 lane branch(`issue/<n>-…`) worktree가 있거나(경로·HEAD), 현재 head에 `verdict: FAIL`(수리 중)이거나, 보유자 claim 뒤의 소유 보고가 `상태: running`.
   - `claimed`: 살아 있는 claim만 있고 위 기록이 없음.
6. **head 순서는 기록에서만 읽는다.** PR의 commit 목록(최대 250개, 공식 문서) 안의 위치가 순서다. 댓글 id나 시각으로 head 순서를 정하지 않는다. 목록에 없는 40자 SHA(예: base SHA)는 순서를 알 수 없으므로 `ignored_shas`로 센다.
7. **이벤트 판정**(댓글 id 오름차순으로 한 번 훑는다, 같은 입력이면 순서를 섞어도 같은 결과):
   - `duplicate`: 앞선 댓글이 이미 보고한 head를 다시 보고.
   - `out_of_order`: 앞선 댓글이 이미 더 새로운 head를 보고했는데 더 오래된 head를 보고.
   - `unreported_head`: PR의 현재 head를 가리키는 댓글이 하나도 없음. "보고 유실"과 "아직 보고 전"은 구별할 수 없으므로 "보고 없이 head가 바뀜"으로만 적는다.
8. **유휴 원인과 근거.** 근거 기록이 없으면 `unknown`이다. 순서대로 처음 맞는 원인을 쓴다.
   - 호출자가 준 pane 관측(`source`·`observed_at` 필수): `unreachable` → 연결 불가, `permission_prompt` → 권한 프롬프트, `input_pending` → 미제출 입력, `working` → 유휴 아님.
   - 행이 있는 레인: 호출자가 준 결정 대기 기록(`pending_decisions[]`, `source`·`recorded_at` 필수) → `user_decision`; `Blocked-By:`가 열린 Issue를 가리킴 → `waiting_prerequisite`; phase `pr_open` → `waiting_audit`; phase `waiting` → `waiting_integration`.
   - 행이 없는 레인: 검사한 Issue 중 ready(claim 줄 0, 열린 선행 0)가 없음 → `no_ready_card`; 있음 → `ready_card_unassigned`.
   - 그 밖은 `unknown`이고, 근거에 무엇을 확인했는지 적는다.
   - **사용자 결정 대기**를 나타내는 구조화된 신호가 저장소에 없다. 그래서 산문을 매칭하거나 새 표지를 만들지 않고 호출자 기록만 받는다(리더의 질문 장부가 근거). `waiting_integration`·`ready_card_unassigned`는 Issue의 목록에 없는 원인이다. 각각 PASS 뒤 통합 대기와 "ready 카드가 있는데 배정 전"을 `unknown`으로 뭉개지 않기 위해 더했다.
9. **배정 제안 초안.** 행이 없는 레인과 ready Issue를 레인 순서·Issue 번호 순으로 짝짓는다. 게시하지 않는다. 만료된 보유자나 handoff 대기 Issue는 claim 줄이 있으므로 ready가 아니고 제안되지 않는다.
10. **한계값.** 요청의 `issues`는 1–50개, `lanes`는 0–50개. GitHub 목록 읽기는 50페이지까지(`leader-tick`과 같은 값). 모두 사용자 결정이 아닌 보수적 기본값이며 측정해서 정한 값이 아니다.

## 3. 하지 않는 것

호스트 세션 이벤트 연동(#A12 결론 뒤), pane 상태의 직접 관측(호출자 몫), 원격 쓰기·배정 게시(기존 effect 경로), `assignments.json` 갱신, 실제 GitHub 실행(fixture만).
