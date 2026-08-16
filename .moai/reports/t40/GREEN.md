# t40 GREEN — 관측성 결함 3건, 수정 후 통과 증거

## Claim (주장)

1. dry-run 아카이브 예고가 실제 동작과 일치 ("will NOT be archived" + "total: N legacy skills present, 0 will be archived")하며, 실제 실행은 sync 전 스냅샷과 대비해 손실(0 of N)을 경고로 보고.
2. 결과 요약이 관리 경로 재배포를 총계에 포함(예: 32+143=175)하고 삭제 파일 수 + 미복구(로컬 전용) 수를 내역으로 표기.
3. --dry-run이 관리 경로 삭제 목록을 미리보기 — 템플릿 복구 파일은 "re-deployed"로 집계, 로컬 전용 파일은 "not restored"로 개별 명시. 읽기 전용(파일 무변경).

## Evidence (증거)

명령 (워크트리 t40, 구현 완료 후):

```
go test ./internal/cli/ -run 'TestDryRunArchive|TestPresentLegacySkillIDs|TestReportArchiveShortfall|TestRenderUpdateOutcome|TestPreviewManagedCleanup' -count=1 -v
```

결과 (전문: `green-run.log`):

```
--- PASS: TestRenderUpdateOutcome (기존)
--- PASS: TestRenderUpdateOutcome_SingularFile (기존)
--- PASS: TestRenderUpdateOutcome_NoColor (기존)
--- PASS: TestRenderUpdateOutcome_ZeroDetailUnchanged (신규)
--- PASS: TestRenderUpdateOutcome_ManagedBreakdown (신규)
--- PASS: TestDryRunArchive_NoSkills_HonestZero (신규)
--- PASS: TestReportArchiveShortfall (신규)
--- PASS: TestPreviewManagedCleanup_EmptyProject (신규)
--- PASS: TestDryRunArchive_NoSkills (기존, 문구 갱신)
--- PASS: TestPreviewManagedCleanup (신규)
--- PASS: TestDryRunArchive (기존)
--- PASS: TestDryRunArchive_PredictsRemovalNotArchive (신규)
--- PASS: TestPresentLegacySkillIDs (신규)
ok  github.com/modu-ai/moai-adk/internal/cli
```

하위 패키지:

```
go test ./internal/cli/update/... -count=1
ok  github.com/modu-ai/moai-adk/internal/cli/update
ok  github.com/modu-ai/moai-adk/internal/cli/update/backup
ok  github.com/modu-ai/moai-adk/internal/cli/update/deploy
ok  github.com/modu-ai/moai-adk/internal/cli/update/merge
ok  github.com/modu-ai/moai-adk/internal/cli/update/plan
ok  github.com/modu-ai/moai-adk/internal/cli/update/report
```

정적 게이트:

```
go vet ./internal/cli/...          → exit 0 (출력 없음)
golangci-lint run ./internal/cli/... → 0 issues. (exit 0)
go build ./internal/...            → exit 0 (출력 없음)
```

## Baseline-attribution (baseline 귀속)

- 위 전부 동일 워크트리(구현 커밋 후, 커밋 전 작업 트리)에서 이번 실행으로 관측.

## Gaps (미검증)

- `go test ./internal/cli/ -count=1 -timeout 540s` 전체 수트: `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` 1건 실패(184s) — **사전 존재 결함**: 변경 없는 원본 main 체크아웃(primary)에서 동일 명령으로 동일 실패 확인됨 (`expected BLOCK ... got decision="" err=<nil>`, 라이브 codex 판정 의존). t40 변경과 무관.
- 로컬 전체 저장소 수트(`go test ./...`)는 규율(CLAUDE.local.md §4)상 미실행 — 전체 판정은 CI.

## Residual-risk (잔여 위험)

- 관리 재배포 카운트는 `deployer.ListTemplates()`의 IsMoaiManaged 템플릿 수 기준(배포가 실제로 쓰는 파일 집합과 동일 소스에서 도출). 배포 도중 렌더 실패 등으로 일부가 안 써졌다면 총계가 실제 쓰기보다 클 수 있음 — 다만 그 경우 배포 스텝 자체가 실패 경로로 감.
- 아카이브 순서 결함(원본이 wipe로 소실) 자체는 t38 카드가 소관 — 본 카드는 보고 정직화만.
