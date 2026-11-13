# `gate-check`

operation `gate.check` · 코드 `src/ops/gate_check.rs`, 규칙 `src/gate.rs`, 수집 `src/gate_collect.rs` · Issue #2674 S1·S1b, #2777(G8)

`gate-check` 는 표준 입력의 PR 스냅숏(댓글, check run, PR·tip 변경 파일, merge-tree 결과, Self-Audit 검사 결과, 판정 base 가 아는 커밋인지, 사용자가 확인한 REVIEW.md PR 번호, tip 의 push CI 판정 `tip_ci`, 본문의 수리 표지 `tip_repair`)으로 리더 병합 게이트 규칙 G8·G1~G7 을 차례로 본다. 모두 맞으면 `pass: true` 와 판정 base·허용한 겹침 파일을, 아니면 `pass: false` 와 멈춘 규칙(`rule`)과 이유를 낸다. 네트워크·git 을 읽지 않고, 상태 게시·병합을 하지 않는다. 규칙 표는 Issue #2674 댓글 6849084101. 판정은 첫 줄이 `AUDIT #` 로 시작하는 댓글뿐이다(본문 중간에 인용한 줄은 판정이 아니다). 예약 철회는 첫 줄 `moai-audit-claim 철회: <레인> (<예약 id>)` 형식만 읽는다.

`gate-check --pr N [--repo OWNER/REPO] [--dir .] [--tip refs/remotes/origin/v2026.11.0] [--review-confirmed 1,2]` 는 스냅숏을 직접 모은다(S1b): PR head·본문·댓글(쪽 넘김)·head 의 check run 을 GitHub 에서 GET 으로 읽고, 로컬 저장소에서 `merge-tree`·변경 파일·판정 base 존재·겹친 `docs/` 의 양쪽 추가 줄 보존·병합 tree 의 manifest sha256 을 잰다. PR 본문은 저장소의 `tools/check_pr_self_audit.py` 로 검사하고, 그 도구가 없으면 G4 에서 멈춘다. 가져오기(fetch)는 하지 않는다: head 가 로컬에 없으면 `refs/pull/N/head` 를 먼저 가져오라는 오류를 낸다.

G8(tip 빨강 멈춤, #2665·#2777)은 맨 앞에서 본다. 수집은 tip 커밋의 check run 을 모든 시도(`filter=all`)까지 읽고, `cancelled` 와 아직 도는 시도를 뺀 뒤 `flow::tip_ci_judgment`(필수 `ci`)로 판정한다. tip 의 CI 가 아직 돌거나 취소만 있으면 첫 부모를 따라 이전 tip 커밋을 최대 10개(`TIP_WALK`) 읽고, 처음 완료된 CI 가 판정한다(`decided_by`). 읽기 실패나 10개 안에 완료된 CI 가 없으면 `unknown`. `red` 면 본문에 `moai-tip-repair:` 줄이 있는 PR 만 통과하고, `unknown`(또는 `tip_ci` 없음)이면 수리 PR 도 멈춘다. `flaky`(같은 커밋에서 실패 뒤 재실행 성공)와 `green` 은 통과한다. `flow-gate-verdict` 와 달리 취소는 실패로 보지 않는다: 새 push 가 앞 실행을 취소하므로 취소는 코드에 대해 말하지 않는다.
