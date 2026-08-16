# t63 RED — retained-key advisory 스트림 분리 결함, 수정 전 실패 증거

카드: `moai todo` t63 — moai update 진행 출력 깨짐 (stdout/stderr 스트림 분리가 원인). Tier S.

## Claim (주장)

수정 전에는 아래 계약을 잡는 테스트가 존재하지 않고, 신규 테스트를 추가하면 컴파일조차 실패한다 (대상 API 미존재).

1. 진행 표시(stdout, tui.ProgressLine 커서 제어 `[2K]` 재그리기 루프)와 advisory(stderr, `node_merge.go` retainedKeySink = os.Stderr 무조건 append)가 스트림 분리되어 '○ Restoring user settings...advisory: retained key ...' 형태로 출력이 깨진다.
2. advisory에 verbose 게이트가 없다 — 바로 옆 `recordMergeFallback`(update_noise.go:93)은 verbose bool을 받아 억제하는데 advisory만 그 ledger를 우회해 49줄을 무조건 뿌린다.
3. 49줄이 전달하는 실질 정보는 "사용자 설정 키 49개 보존" 한 줄이다.

## Evidence (증거)

명령 (워크트리 t63, RED 시점 — 구현 커밋 전):

```
go test ./internal/cli/ -run 'TestRenderRetainedKeyAdvisory' -count=1
go test ./internal/cli/update/backup/ -run 'TestMergeYAML3WayRetained|TestMergeYAML3Way_LegacySinkTextPreserved|TestRestoreMoaiConfigRetained|TestRestoreMoaiConfig_LegacyWrapperReemits' -count=1
```

출력 (전문: `t63-red-cli.log`, `t63-red-backup.log`):

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_retained_advisory_test.go:30:36: undefined: backup.RetainedKeyRef
internal/cli/update_retained_advisory_test.go:31:18: undefined: backup.RetainedKeyRef
internal/cli/update_retained_advisory_test.go:44:2: undefined: renderRetainedKeyAdvisory
internal/cli/update_retained_advisory_test.go:68:2: undefined: renderRetainedKeyAdvisory
internal/cli/update_retained_advisory_test.go:89:2: undefined: renderRetainedKeyAdvisory
internal/cli/update_retained_advisory_test.go:94:2: undefined: renderRetainedKeyAdvisory
internal/cli/update_retained_advisory_test.go:105:2: undefined: renderRetainedKeyAdvisory
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
(exit 1)

# github.com/modu-ai/moai-adk/internal/cli/update/backup [github.com/modu-ai/moai-adk/internal/cli/update/backup.test]
internal/cli/update/backup/retained_collect_test.go:73:23: undefined: MergeYAML3WayRetained
internal/cli/update/backup/retained_collect_test.go:115:15: undefined: RestoreMoaiConfigRetained
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup [build failed]
FAIL
(exit 1)
```

또한 수정 전 코드 상태 (배선 리뷰):
- `internal/cli/update_template_sync.go:450` — Restore Settings 스텝이 `backup.RestoreMoaiConfig`(legacy)를 호출 → 내부 `MergeYAML3Way`가 키마다 `advisory:` 줄을 `retainedKeySink`(=os.Stderr)에 직접 append. verbose 게이트 없음(grep `updateVerboseMode|Verbose` node_merge.go 0건 — 카드 실측과 동일).
- `internal/tui/progress_line.go` — 진행 줄은 stdout `out`에 `\r\x1b[2K` 커서 제어로 재그리기. stderr 끼어들면 깨짐 (카드 실측: mink-code에서 49줄).

## Baseline-attribution (baseline 귀속)

- 워크트리: `.claude/worktrees/agent-af71696d26ebce0bc` (branch `WT-t63`, base = release/v3.1.1 병합 완료 시점 `f4ad34f04`, 구현 0커밋 상태).
- 위 출력은 이 트리에서 이번 실행으로 관측된 것.

## Gaps (미검증)

- 실제 PTY 터미널에서의 커서 깨짐 재현은 로컬에서 수행하지 않음 — 결함 실측(2026-08-16, mink-code)은 카드에 기록된 관측을 근거로 사용. 본 카드의 검증은 단위 수준(스트림 미혼입 + 렌더 계약)으로 수행.
- `runTemplateSyncWithReporter` 전체 경로(통합)에서의 요약 줄 출력은 간접 검증(헬퍼 단위 테스트 + 배선 코드 리뷰)이며, runUpdate를 그대로 실행하는 E2E는 로컬에서 돌리지 않음(로컬 전체 수트 금지 규율 CLAUDE.local.md §4).

## Residual-risk (잔여 위험)

- 최상단 slog WARN의 logfmt 서식 이질성은 t62 카드 소관 — 본 카드는 건드리지 않음(새 이질 서식 도입 없음).
