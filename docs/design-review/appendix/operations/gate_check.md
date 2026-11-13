# `gate-check`

operation `gate.check` · 코드 `src/ops/gate_check.rs`, 규칙 `src/gate.rs`, 수집 `src/gate_collect.rs` · Issue #2674 S1·S1b

`gate-check` 는 표준 입력의 PR 스냅숏(댓글, check run, PR·tip 변경 파일, merge-tree 결과, Self-Audit 검사 결과, 판정 base 가 아는 커밋인지, 사용자가 확인한 REVIEW.md PR 번호)으로 리더 병합 게이트 규칙 G1~G7 과 G9 를 차례로 본다. G9(#2674 R1)는 마지막 판정의 감사자가 이 PR 커밋의 `Moai-Claim` 트레일러가 가리키는 claim 의 세션(저자)이 아닌지 본다. 스냅숏의 `pr_authors` 가 없으면 모른다는 뜻이라 멈춘다. 모두 맞으면 `pass: true` 와 판정 base·허용한 겹침 파일을, 아니면 `pass: false` 와 멈춘 규칙(`rule`)과 이유를 낸다. 네트워크·git 을 읽지 않고, 상태 게시·병합을 하지 않는다. 규칙 표는 Issue #2674 댓글 6849084101. 판정은 첫 줄이 `AUDIT #` 로 시작하는 댓글뿐이다(본문 중간에 인용한 줄은 판정이 아니다). 예약 철회는 첫 줄 `moai-audit-claim 철회: <레인> (<예약 id>)` 형식만 읽는다.

`gate-check --pr N [--repo OWNER/REPO] [--dir .] [--tip refs/remotes/origin/v2026.11.0] [--review-confirmed 1,2]` 는 스냅숏을 직접 모은다(S1b): PR head·본문·댓글(쪽 넘김)·head 의 check run 을 GitHub 에서 GET 으로 읽고, 로컬 저장소에서 `merge-tree`·변경 파일·판정 base 존재·겹친 `docs/` 의 양쪽 추가 줄 보존·병합 tree 의 manifest sha256 을 잰다. PR 본문은 저장소의 `tools/check_pr_self_audit.py` 로 검사하고, 그 도구가 없으면 G4 에서 멈춘다. 가져오기(fetch)는 하지 않는다: head 가 로컬에 없으면 `refs/pull/N/head` 를 먼저 가져오라는 오류를 낸다.
