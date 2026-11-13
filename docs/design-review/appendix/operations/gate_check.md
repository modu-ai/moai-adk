# `gate-check`

operation `gate.check` · 코드 `src/ops/gate_check.rs`, 규칙 `src/gate.rs` · Issue #2674 S1

`gate-check` 는 표준 입력의 PR 스냅숏(댓글, check run, PR·tip 변경 파일, merge-tree 결과, Self-Audit 검사 결과, 판정 base 가 아는 커밋인지, 사용자가 확인한 REVIEW.md PR 번호)으로 리더 병합 게이트 규칙 G1~G7 을 차례로 본다. 모두 맞으면 `pass: true` 와 판정 base·허용한 겹침 파일을, 아니면 `pass: false` 와 멈춘 규칙(`rule`)과 이유를 낸다. 네트워크·git 을 읽지 않고, 상태 게시·병합을 하지 않는다. 규칙 표는 Issue #2674 댓글 6849084101. 판정은 첫 줄이 `AUDIT #` 로 시작하는 댓글뿐이다(본문 중간에 인용한 줄은 판정이 아니다). 예약 철회는 첫 줄 `moai-audit-claim 철회: <레인> (<예약 id>)` 형식만 읽는다.
