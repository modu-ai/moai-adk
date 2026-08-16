# t40 RED — 관측성 결함 3건, 수정 전 실패 증거

카드: `moai todo list` t40 — moai update 관측성 결함 3건 (조용한 실패). Tier S.

## Claim (주장)

수정 전에는 아래 3결함을 잡는 테스트가 존재하지 않고, 신규 테스트를 추가하면 컴파일조차 실패한다 (대상 API 미존재).

1. dry-run의 아카이브 예고가 실제 결과와 상시 불일치 ("N skills archived" 예고 ↔ 실제 항상 0건 — wipe가 먼저 원본 삭제).
2. "Updated N files"가 IsMoaiManaged 경로를 제외해 실제 쓰기 파일 수의 일부만 보고 (실측 32 vs 실제 175), 삭제는 요약에 미표시.
3. --dry-run이 CleanMoaiManagedPaths 삭제 예정 목록을 미리보기하지 않음 (로컬 전용 파일 사전 인지 불가).

## Evidence (증거)

명령 (워크트리 t40, RED 시점 — 구현 커밋 전):

```
go test ./internal/cli/ -run 'TestDryRunArchive_PredictsRemovalNotArchive|TestDryRunArchive_NoSkills_HonestZero|TestPresentLegacySkillIDs|TestReportArchiveShortfall|TestRenderUpdateOutcome_ManagedBreakdown|TestRenderUpdateOutcome_ZeroDetailUnchanged|TestPreviewManagedCleanup' -count=1
```

출력 (전문: `red-compile.log`):

```
internal/cli/update_observability_test.go:99:2: undefined: reportArchiveShortfall
internal/cli/update_observability_test.go:129:12: undefined: updateOutcomeDetail
internal/cli/update_observability_test.go:134:59: too many arguments in call to renderUpdateOutcome
internal/cli/update_observability_test.go:197:12: undefined: previewManagedCleanup
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

또한 기존 `TestDryRunArchive_NoSkills`(update_dry_run_test.go)는 수정 전 문구 "0 skills archived" — 정직하지 않은 예고를 그대로 단언하던 상태.

## Baseline-attribution (baseline 귀속)

- 워크트리: `.claude/worktrees/t40` (branch `worktree-t40`, base = release/v3.1.1 `36a12cf82` 병합 완료 시점, 구현 0커밋 상태).
- 위 출력은 이 트리에서 이번 실행으로 관측된 것.

## Gaps (미검증)

- 실 runUpdate 전체 경로(통합)에서의 손실 경고 출력은 간접 검증(헬퍼 단위 테스트 + 배선 코드 리뷰)이며, runUpdate를 그대로 실행하는 E2E는 로컬에서 돌리지 않음(로컬 전체 수트 금지 규율 §4).

## Residual-risk (잔여 위험)

- 삭제 로직 자체(t38 카드 소관)는 이번에 고치지 않음 — 아카이브 순서 결함은 존재하며, 본 카드는 그 손실이 "조용히" 발생하지 않도록 보고만 정직화.
