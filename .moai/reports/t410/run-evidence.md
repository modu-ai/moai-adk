# SPEC-DRIFT-CLOSE-BODY-001 — run-phase 원장 (카드 t410)

각 마일스톤 경계에서 `verification-claim-integrity.md` §3의 5절(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)로 기록한다. Evidence에는 **명령과 축자 출력과 exit code**를 적는다 — 요약은 증거가 아니다.

## 좌표 (run-phase 착수 시점)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410
$ git branch --show-current
WT-drift-false-positive
$ git rev-parse --short HEAD
c323bb491
$ git status --short
?? .moai/reports/t410/drift-before.txt
```

착수 HEAD `c323bb491` · 브랜치 `WT-drift-false-positive` · 트리 `.claude/worktrees/t410`.

## §C 사전 점검 — 오케스트레이터 실행분 인용 (재실행하지 않음)

| 명령 | 결과 |
|---|---|
| `go build -o /tmp/moai-t410 ./cmd/moai` | rc 0 |
| `/tmp/moai-t410 spec drift --no-cache > .moai/reports/t410/drift-before.txt` | rc 0, 637 SPEC 행 |
| `grep SPEC-V3R6-SESSION-HANDOFF-AUTO-001 .moai/reports/t410/drift-before.txt` | `completed / in-progress / DRIFT` |

plan.md §C의 정지 조건(확정 대상 행이 이미 DRIFT가 아니면 blocker 반환)은 **불성립** — 결함이 이 트리에 재현되므로 착수한다.

**귀속 주의** — 위 세 줄은 오케스트레이터가 이 트리에서 실행한 측정을 **인용**한 것이지 이 에이전트가 재측정한 것이 아니다. AC-DCB-004/005의 "수리 전" 표는 M3에서 이 트리 좌표와 함께 다시 확인한다.
