# t63 GREEN — retained-key advisory TUI 단일 스트림 합치기, 수정 후 통과 증거

## Claim (주장)

1. update 템플릿 동기화의 Restore Settings 스텝은 `RestoreMoaiConfigRetained`로 병합을 수행하며, 이 경로에서 **retained-key advisory 텍스트가 stderr sink로 한 줄도 나가지 않는다** (단위 실측: sink 버퍼 0바이트). advisory는 데이터(키 목록)로 수집돼 렌더 계층으로 반환된다.
2. 렌더는 진행 줄과 **동일한 stdout 채널**(`out`, tui.ProgressLine이 쓰는 writer)로: 기본 **요약 1줄**(`✓ N user settings key(s) preserved (run with --verbose to list)`), `--verbose`에서만 키 목록 펼침(dim 스타일 `section: key` 줄 per 키). verbose 게이트는 `updateVerboseMode` — `recordMergeFallback`(update_noise.go)이 쓰는 동일 ledger.
3. 키 0개 → 출력 없음(클린 런 무소음). NO_COLOR → SGR 0개(기존 렌더 헬퍼와 동일 열화).
4. 레거시 진입점(`MergeYAML3Way` / `RestoreMoaiConfig`)은 REQ-UYP-007 advisory-텍스트-on-sink 계약을 바이트 동등하게 유지 — 기존 계약 테스트 3종(uyp_ac_test, merge_useradd_test, config/template_removed_key_test AC-CKH-014) 전부 무수정 통과. clean-reinstall·--restore-config 경로(아직 TUI 렌더로 마이그레이션 안 됨)의 stderr 보고는 legacy 래퍼의 재출력으로 보존.
5. 병합/백업 의미론 불변 — diff는 출력 라우팅과 verbose 게이팅만 (3-way 판정 로직 0행 변경, 단 `mergeErr != nil`시 수집된 키 폐기 추가: 2-way 폴백에서는 advisory 의미론이 적용되지 않으므로).

## Evidence (증거)

명령 (워크트리 t63, 구현 완료 후 — 전문 로그: `green-run-*.log`, `green-run-static.log`):

신규 테스트 + 기존 restore/merge 계약 (42 PASS):

```
go test ./internal/cli/ -run 'TestRenderRetainedKeyAdvisory|TestRestoreMoaiConfig|TestMergeYAML3Way|TestDeepMerge3Way' -count=1 -v
→ --- PASS: TestRenderRetainedKeyAdvisory_DefaultSingleSummaryLine (신규)
→ --- PASS: TestRenderRetainedKeyAdvisory_VerboseExpandsKeyList (신규)
→ --- PASS: TestRenderRetainedKeyAdvisory_EmptySilent (신규)
→ --- PASS: TestRenderRetainedKeyAdvisory_NoColorZeroSGR (신규)
→ (기존 TestRestoreMoaiConfig* 30종 + TestMergeYAML3Way* / TestDeepMerge3Way / TestRestoreMoaiConfigLegacy* 전부 PASS)
PASS  ok  github.com/modu-ai/moai-adk/internal/cli
```

백업 패키지 신규 테스트 (전문: `green-run-backup-new.log`):

```
go test ./internal/cli/update/backup/ -run 'TestMergeYAML3WayRetained|TestMergeYAML3Way_LegacySinkTextPreserved|TestRestoreMoaiConfigRetained|TestRestoreMoaiConfig_LegacyWrapperReemits' -count=1 -v
--- PASS: TestMergeYAML3WayRetained_CollectsWithoutSinkText (0.00s)
--- PASS: TestMergeYAML3Way_LegacySinkTextPreserved (0.00s)
--- PASS: TestRestoreMoaiConfigRetained_CollectsRefsAndStaysSilentOnSink (0.01s)
--- PASS: TestRestoreMoaiConfig_LegacyWrapperReemitsAdvisoryText (0.01s)
PASS  ok  github.com/modu-ai/moai-adk/internal/cli/update/backup
```

하위 패키지 전체 (내부 시그니처 변경 `deepMerge3WayTo`/`mergeMappingNode3Way`의 14개 테스트 호출부 기계 갱신 포함):

```
go test ./internal/cli/update/... -count=1
ok  github.com/modu-ai/moai-adk/internal/cli/update
ok  github.com/modu-ai/moai-adk/internal/cli/update/backup
ok  github.com/modu-ai/moai-adk/internal/cli/update/deploy
ok  github.com/modu-ai/moai-adk/internal/cli/update/merge
ok  github.com/modu-ai/moai-adk/internal/cli/update/plan
ok  github.com/modu-ai/moai-adk/internal/cli/update/report
```

크로스 패키지 계약 (AC-CKH-014, 무수정):

```
go test ./internal/config/ -run 'TestTemplateRemovedKeySurvivesUserConfig' -count=1 -v
--- PASS: TestTemplateRemovedKeySurvivesUserConfig (0.00s)
PASS  ok  github.com/modu-ai/moai-adk/internal/config
```

tuxiu 골든 (갱신 불필요 — 아래 참고):

```
go test ./internal/cli/ -run 'TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXDataValuesPreserved|TestTUXPostM4CarriesNewPresentation' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli
```

정적 게이트 (전문: `green-run-static.log`):

```
go vet ./internal/cli/...            → exit 0 (출력 없음)
go build ./internal/cli/...          → exit 0 (출력 없음)
golangci-lint run ./internal/cli/... → 0 issues. (exit 0)
gofmt -l <내 파일 7종>              → 빈 목록 (clean)
```

### 골든 파일 갱신 불필요 — 근거

카드는 "수정 시 tuxiu 골든(tty/notty/nocolor) 갱신이 동반된다"로 예상했으나, 실측 결과 **갱신 대상이 없다**:

1. `grep -rl 'advisory|retained' internal/cli/testdata/tuxiu/` → 매치 0건 (exit 1). 골든 캡처 시나리오(일회용 신규 프로젝트)에는 retained key가 없어 advisory 줄이 골든에 존재하지 않았다.
2. 본 수정은 "retained key가 있는 출력"에만 변화를 만든다 → 골든 캡처 시나리오의 출력은 불변.
3. 골든 재캡처 하네스(스크래치패드 capture.sh + pty_capture.py)는 저장소에 더 존재하지 않음(find 0건) — 골든 비교는 커밋된 2개 fixture 세트(tuxiu_characterization_test.go) 간 대비로만 수행.
4. 따라서 검증 = 특성화 수트 4종 전부 통과(위) + `git status`에서 골든 0수정. DATA 불변(AC-TUXIU-016)은 기계적으로 확인됨.

## Baseline-attribution (baseline 귀속)

- 위 전부 동일 워크트리(구현 완료 후, 커밋 전 작업 트리, HEAD `f4ad34f04` 기반)에서 이번 실행으로 관측.

## Gaps (미검증)

- `internal/cli` 패키지 전체 수트(`go test ./internal/cli/ -count=1`) 미실행 — t40 GREEN.md에 기록된 사전 존재 결함(`TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey`, 라이브 codex 의존)이 섞여 있고, 카드 규율상 타깃 테스트만 로컬 실행(전체 판정은 CI).
- `runUpdate` 전체 경로 E2E(요약 줄의 화면 배치 실측) 미실행 — 간접 검증(헬퍼 단위 + 배선 리뷰)으로 대체, t40과 동일 방식.

## Residual-risk (잔여 위험)

- **clean-reinstall(`update_clean_install.go`)과 `--restore-config`(`update_restore.go` → `RestoreFromBackupDir`) 경로는 여전히 legacy `RestoreMoaiConfig`를 통해 stderr에 advisory를 뿌린다**(래퍼 재출력으로 바이트 동등 보존). 이 경로들은 `[clean-reinstall]` 평문 줄 관행을 써서 TUI 렌더 도입 시 새 이질 서식이 되므로 본 카드(카드가 지목한 update 진행 화면)에서는 의도적으로 미마이그레이션 — 후속 카드 후보.
- `RestoreFromBackupDir`의 fixed-point 2패스는 legacy 래퍼를 2회 호출하므로 advisory가 2회 출력된다 — 수정 전과 동일한 거동(회귀 아님).
- 최상단 slog WARN logfmt 이질성은 t62 소관.
