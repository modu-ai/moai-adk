# t1488 card-review (advisory)

도구: `mcp__moai__codex_review` scope=uncommitted (커밋 전 실행 — card 범위와 같은 변경 집합), backend codex.

판정: fail (P2 2건, 자문)

1. 테스트 자식의 HOME 미격리 — fixture가 MOAI_HOME을 비워 실제 `~/.moai/run/`에 쓴다. → 수리: 자식 env에 `HOME=<t.TempDir()>`.
2. `#!/bin/sh` 가짜 codex는 Windows에서 실행되지 않음 → 수리: 테스트 파일에 `//go:build !windows`.

수리 후 재실행: `-race -run 'Restamp|RunOwner|PaneDoor|DirectPOSIX|SD_AC003|LaneLoop'` → ok. 재리뷰는 하지 않았다.
