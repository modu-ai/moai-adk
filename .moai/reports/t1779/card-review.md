# t1779 card-review (codex_review scope=card)

- round 1: 2026-10-19 00:2xZ · backend codex · base `ec164d45d2` · tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1779` · verdict **fail**(advisory) — P2 2건
  - P2-1 process_probe_unix.go:22 — ps 프로브가 ExitError.Error()("exit status 1")만 반환, "Operation not permitted" 시그니처(stderr) 미반영 → 샌드박스 런치가 등록·기동 진행. 시험이 문자열 오류 주입으로 실경로 미검증. 처방: stderr 캡처 결합 + 가짜 ps(서브셰임)로 실 ExitError 경로 시험.
  - P2-2 codex_lane_precondition.go:85 — TTY 검사가 TERM 의존적: xterm-256color 상속 비TTY(파이프)·stdin 리다이렉트 통과 → Codex 에 같은 FD 전달로 런 참여·레인 점유 선행. 처방: TERM 무관 실제 stdin/stdout 단말 확인(seam).
- round 1 수리: (delegate, 진행 중)
