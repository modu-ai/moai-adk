# t1779 card-review (codex_review scope=card)

- round 1: 2026-10-19 00:2xZ · backend codex · base `ec164d45d2` · tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1779` · verdict **fail**(advisory) — P2 2건
  - P2-1 process_probe_unix.go:22 — ps 프로브가 ExitError.Error()("exit status 1")만 반환, "Operation not permitted" 시그니처(stderr) 미반영 → 샌드박스 런치가 등록·기동 진행. 시험이 문자열 오류 주입으로 실경로 미검증. 처방: stderr 캡처 결합 + 가짜 ps(서브셰임)로 실 ExitError 경로 시험.
  - P2-2 codex_lane_precondition.go:85 — TTY 검사가 TERM 의존적: xterm-256color 상속 비TTY(파이프)·stdin 리다이렉트 통과 → Codex 에 같은 FD 전달로 런 참여·레인 점유 선행. 처방: TERM 무관 실제 stdin/stdout 단말 확인(seam).
- round 1 수리: (delegate, 진행 중)
- round 1 수리: 커밋 `31775ba19b` — P2-1(ps 프로브 stderr 결합, 가짜 ps PATH 주입으로 실 exec 경로 시험)·P2-2(stdio seam 무조건 검사, TERM 무관) RED→GREEN.
- round 2: FAIL — P2 1건(신규 ps 시험 2종 Unix 전용 무태그 → Windows CI 차단) → 수리 `df6aeb24f3`(파일 수준 //go:build !windows 태그 + _unix_test.go 접미사, fake-ps 시험 분리, GOOS=windows vet 직접 확인).
- round 3: FAIL — P2 1건(TTY 단위 시험이 실제 샌드박스 프로브 실행 — 환경 의존 시험) → 수리 `da4b49ee74`(프로브 핀 + t.Cleanup 복원, +3/−2).
- round 4: **pass**(결함 0건, "확실히 지적할 결함은 발견하지 못했어") — 4라운드 수렴 완료.
- 레인 독립 스팟검증: DumbTerm ok 3.660s · Recursive ok 3.077s · SandboxFakePS ok 1.780s · NonTerminalStdio ok 1.502s · gofmt 클린(각 라운드 시점).
