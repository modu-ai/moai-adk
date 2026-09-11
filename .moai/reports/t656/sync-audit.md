# Sync 감사 — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (카드 t656)

- 감사자: sync-auditor (독립 판정, 4차원 평가)
- 대상 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t656`, 브랜치 `WT-update-value-merge`, HEAD `0b5a1b3b1`, 작업 트리 깨끗함(감사 시작·종료 시 `git status --short` 빈 출력)
- 카드 기준: `CARD_BASE = git merge-base develop HEAD = 0db675bedcae69c7ade2f02f106649a715fb14f6`, 코드 diff 범위 `41a470641..0b5a1b3b1`
- 평가 프로필: SPEC 에 `evaluator_profile` 없음 → 기본 프로필(Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security)
- 적용 규칙: `.claude/rules/moai/core/verification-claim-integrity.md` §1·§2·§3, `.claude/rules/moai/development/verification-completeness.md` §1.1(빈 선택 집합)·§2(뮤턴트), `.claude/rules/moai/development/spec-frontmatter-schema.md` § SHA placeholder backfill exemption

## 판정

**PASS-WITH-DEBT** — 코드는 통과. 부채는 모두 선택(optional) 항목이며 코드 정확성과 SPEC 요구사항에 영향을 주지 않는다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 93/100 | PASS | merge/backup AC 시험 전부 `--- PASS`(아래 명령 1·2), cli 슬롯 증거에 `--- PASS` 12줄(`slot-1-green.txt`), M-01 재적용 시 AC-001·003·012·013 이 값 단정에서 RED(명령 8) |
| Security (25%) | 95/100 | PASS | 스냅숏 두 경로 모두 `.gitignore:352:.moai/cache/` 로 무시, 추적 파일 대조군은 exit 1, 배포 템플릿 `.gitignore:241` 에도 `.moai/cache/`, 파일 권한 `defs.FilePerm = 0o644`(살아 있는 settings.json 과 동일), 경로는 상수 조합뿐이라 경로 탐색 입력 없음 |
| Craft (20%) | 88/100 | PASS | 커버리지 merge 92.9% / backup 90.3%(M1 기준선 92.1% / 90.1% 이상), `golangci-lint` `0 issues.`, `go vet ./internal/cli/` exit 0, `gofmt -l` 빈 출력 |
| Consistency (15%) | 86/100 | PASS | @MX 태그 프로토콜 준수, 영어 주석, 파일 스타일 일치. CHANGELOG 메커니즘 서술 부정확(F3), §E.3 `run_commit_sha` 미백필(F4) |

- 가중 평균: 0.40×93 + 0.25×95 + 0.20×88 + 0.15×86 = **91.45**
- 조화 평균: 4 / (1/93 + 1/95 + 1/88 + 1/86) = **90.4**
- must-pass 방화벽: Functionality 93, Security 95 — 둘 다 통과

## AC 판정표

| AC | 판정 | 근거 |
|---|---|---|
| AC-USB-001 | 충족 | `TestMergeUserFiles_SnapshotBaseDeliversTemplateValueChange/with_canonical` PASS, 대조 셀 PASS. M-01 재적용 시 `:105 statusLine.command = old, want new` |
| AC-USB-002 | 충족 | `TestMergeUserFiles_SnapshotBaseKeepsUserEdit` PASS. M-02 기록(`mut-M-02.txt`) |
| AC-USB-003 | 충족 | `...BothChangedReportsConflict` PASS. M-01 재적용 시 `:158 conflict line count = 0, want 1` |
| AC-USB-004 | 충족 | `...SnapshotFallbackMatchesDerivedBase` 하위 5개 PASS. M-04 기록 |
| AC-USB-005 | 충족 | 슬롯 GREEN PASS. M-05 는 `:296 (ii) staging ... = ""`, M-06c-w 는 `:293 (i)` 에서 RED |
| AC-USB-006 | 충족 | 두 하위 테스트 PASS. M-06c 기록(`newKey = <nil>, want 1`) |
| AC-USB-007 | 충족(init 판정 위치만 소스 오프셋 대체라 약함) | 하위 6개 PASS. M-07a~e, M-06t-w, M-D5f, M-D5g-wb/wd/s 모두 값 단정 RED. init 판정 위치는 `:483 offset 17355 > 16435` 소스 위치 검사로만 확인 — acceptance 가 허용하고 사유를 기록한 대체다 |
| AC-USB-008 | 충족 | backup 두 셀·merge 셀 PASS, cli 두 셀 슬롯 PASS. M-08(`:511`), M-08L(`:537`), M-08s, M-08p RED |
| AC-USB-009 | 충족 | `TestSectionsSnapshot_UnaffectedBySettingsSubpath` 셀 A·B PASS, 기존 7개 PASS 확인(명령 2 출력에 7개 이름 모두 `--- PASS`) |
| AC-USB-010 | 충족 | 기존 6개 + 신규 1개 PASS(명령 1) |
| AC-USB-011 | 충족 | `...ScopedToSettingsJSON` PASS. M-11 기록 |
| AC-USB-012 | 충족 | 대조 셀 포함 PASS. M-01 재적용 시 `:249 user-deleted statusLine came back` |
| AC-USB-013 | 충족 | PASS. M-01 재적용 시 `:275 conflict line count = 0, want 1` |
| AC-USB-014 | 충족 | 하위 3개 PASS(실제 `template.NewDeployer` 비강제 규칙 사용). cli 대신 backup 에 구현한 사실과 사유가 acceptance.md:211 에 기록됨. M-14 기록 |
| AC-USB-015 | 충족 | 명령 5 표 — 6개 명령 모두 기대와 일치 |
| AC-USB-016 | 충족 | c1~c8 하위 7개 PASS. M-D5a/b/c/d/i 기록, M-D5b 는 예측과 다른 셀에서 죽었고 그 사유가 기록됨 |

## 발견 사항 (구조화 결함 목록)

| id | 심각도 | 차단 | 위치 | 결함 | 수정 지시 | 확신도 |
|---|---|---|---|---|---|---|
| F1 | Low | optional | `internal/cli/update/merge/settings_snapshot_flow.go:17-19` | 주석은 병합 오류가 "파일을 쓰기 전, 매니페스트·임베드 템플릿 로드에서만 난다"고 단언하지만, `MergeUserFilesWithOutcome` 은 루프 안의 `os.WriteFile` 실패로도 반환한다(`merge.go:254,276,297,306`). settings.json 쓰기가 잘린 뒤 실패하면 결과가 기록되지 않아 `Preserved` 가 false → 렌더를 승격하는데, 이때 살아 있는 파일은 렌더가 아니다. 사용자 데이터는 이미 그 쓰기 실패로 손상된 상태라 새 회귀는 아니다 | 주석을 실제 오류 출처에 맞게 고치거나, `err != nil` 이고 settings.json 결과가 비어 있으면 대기본을 버리도록 한 줄 추가 | 높음 |
| F2 | Low | optional | `internal/cli/update/backup/settings_snapshot.go:115` | 대기본 `os.WriteFile` 이 부분 기록 후 실패하면 대기본이 남고, 흐름 끝 settle 이 그것을 승격한다. `LoadSettingsSnapshot` 의 JSON 객체 검증이 잘린 파일을 거부하므로 결과는 유도 base 폴백 — 안전하게 퇴화한다 | 쓰기 실패 경고 뒤 `discardSettingsSnapshot` 호출(위생) | 높음 |
| F3 | Low | optional | `CHANGELOG.md:12` | 이전 메커니즘을 "live file 자체를 base 로 썼다"고 서술하지만, 실제 유도 base 는 새 렌더의 값을 사용자 키로 좁힌 것(`base.go` `pruneToShared` — "Values always come from updated")이다. 결과(건드리지 않은 키의 템플릿 값 변경이 전달되지 않음)는 맞고 원인 서술이 틀렸다. "after the first `moai update`" 도 부정확 — 이 한계는 첫 update 부터 있었다. 또 "closing a hole where an unrelated on-disk edit could masquerade as a template render" 는 출시된 적 없는 plan 단계 대안을 닫은 것처럼 읽힌다 | 메커니즘 문장을 "the derived base took the new template's values for shared keys, so base and updated always agreed and a template value change was invisible" 취지로 교체하고, 미출시 대안 언급 삭제 | 높음 |
| F4 | Low | optional | `progress.md:329` | §E.3 `run_commit_sha: pending-backfill` 이 run 단계 종료 뒤에도 백필되지 않았다. §E.4 의 `sync_commit_sha` placeholder 는 D3 예외상 정당하지만, §E.3 은 run 커밋(`b5b5883e9`) 이후 증거 커밋이 6개나 쌓였는데도 비어 있다. 후보값과 소유자가 §E.4:365 에 기록돼 있어 추적은 가능 | manager-develop 이 후속 커밋에서 `run_commit_sha: b5b5883e9` 로 백필 | 높음 |
| F5 | Info | optional | `settings_snapshot.go` `promoteSettingsSnapshot` | 확정본 자리가 디렉터리로 막히면 한 번의 update 에서 판정 단계와 settle 단계가 각각 경고해 `settings-snapshot-promote-failed:` 줄이 두 번 나올 수 있다(버전 일치 경로는 한 번, AC-008 이 그 경로만 단정). 동작은 비차단이라 무해 | 필요하면 문서에 "경고는 시도당 한 줄"로 명시 | 중간 |
| F6 | Info | optional | `progress.md` §E.4 gaps 마지막 줄 | "sync-auditor was not invoked" 는 이 감사로 낡았다 | 백필 커밋 때 이 감사 경로(`.moai/reports/t656/sync-audit.md`)로 갱신 | 높음 |

## 정확성 위험 점검 결과

- **A1 회귀(사용자 편집 값이 템플릿 값으로 덮이는 경로)**: 확정본은 배포기가 쓴 바이트임이 매니페스트 `template_managed` + `template_hash == HashBytes(파일)` 로 증명될 때만 기록된다(`internal/template/deployer.go:274-278` 에서 해시는 쓴 바이트의 해시). 사용자 편집 키는 current ≠ base 라 사용자 변경으로 읽힌다. init 의 자율성 번들 재기록은 스테이징 뒤라 사용자 변경으로 읽힌다(c6). 병합이 보존 경로를 타면 대기본을 버려 이전 base 를 유지한다(c2). 덮어쓰기 경로를 찾지 못했다.
- **판정 위치 순서**: `runUpdate`(`internal/cli/update.go`, HEAD `0b5a1b3b1`) 에서 `--binary` 반환(:331 부근)과 `--dry-run` 반환(:365) 뒤인 :384 에서 판정하고, `stripRetiredV2DenyEntries`(:404) 보다 앞선다. v2 clean-reinstall 분기와 버전 일치 건너뛰기는 그 뒤라 한 지점이 모든 update 흐름을 덮는다. dry-run 의 clean-reinstall 계획 호출은 `update_clean_install.go:203` 에서 배포 전에 반환한다. init 은 `executor.Execute` 앞.
- **원자성**: 승격은 같은 디렉터리 `os.Rename`(원자적, Windows 에서도 Go 가 기존 파일 교체). 대기본 기록은 비원자적이지만 F2 대로 안전하게 퇴화한다.
- **중단 경로**: Restore 단계 실패나 init 의 `EnsureProjectLayout` 실패로 settle 에 도달하지 않으면 대기본이 남고, 다음 흐름의 판정이 살아 있는 파일과 바이트 비교로 처리한다(c4/c5).
- **Windows**: `filepath.FromSlash`/`ToSlash` 사용, 크로스 빌드 exit 0 은 슬롯 증거(`progress.md:281`)로만 확인 — 이번 감사에서 재실행하지 않음(Gap).
- **실제 홈 접근**: 새 시험은 `homeSeamSpy`·`userHomeDirFn` 시접을 쓰고 `t.Parallel()` 을 쓰지 않는다. 추가된 `Setenv("HOME"` 0건(명령 5).

## 실행한 명령과 원문 결과 (이번 감사, 이 트리)

1. `go test -count=1 -cover ./internal/cli/update/merge/... ./internal/cli/update/backup/...`
   ```
   ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.536s	coverage: 92.9% of statements
   ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.038s	coverage: 90.3% of statements
   ```
2. `go test -count=1 -v -run 'Snapshot|TestMergeAddsTemplateEntryUserNeverHad|...' ./internal/cli/update/merge/` → 최상위 17개 `--- PASS`(기존 6 + 스냅숏 base 8 + 흐름 3), 끝줄 `ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.555s`. 같은 방식의 backup 실행 끝줄 `ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.816s`, `--- FAIL` 0줄, `[no tests to run]` 없음
3. `go vet ./internal/cli/` → exit 0, 출력 없음 / `gofmt -l internal/cli/` → exit 0, 출력 없음
4. `golangci-lint run ./internal/cli/update/...` → `0 issues.`
5. AC-USB-015 (`CARD_BASE=0db675bedcae69c7ade2f02f106649a715fb14f6`, 명령마다 따로 실행)

   | 명령 | 결과 |
   |---|---|
   | `git check-ignore -v .moai/cache/template-snapshot/claude/settings.json` | `.gitignore:352:.moai/cache/	.moai/cache/template-snapshot/claude/settings.json` (exit 0) |
   | `git check-ignore -v .moai/cache/template-snapshot/claude/settings.json.pending` | `.gitignore:352:.moai/cache/	...settings.json.pending` (exit 0) |
   | `git check-ignore -v .claude/settings.json` | exit 1 (대조군) |
   | `git diff --name-only $CARD_BASE..HEAD` | 76줄 (대조군 ≥1) |
   | `git diff --name-only $CARD_BASE..HEAD -- internal/merge/ internal/template/templates/` | 빈 출력 |
   | `git diff $CARD_BASE..HEAD -G 'Setenv\("HOME"' --name-only -- '*_test.go'` | 빈 출력 |
6. `git diff --stat 8a44a68bc..HEAD -- internal/` → 빈 출력 (슬롯 트리 이후 Go 변경 없음 — 슬롯 증거가 현재 코드에 귀속됨)
7. `grep -n "moai/cache" internal/template/templates/.gitignore` → `241:.moai/cache/`
8. 뮤턴트 재적용(M-01, `merge.go:285` 을 `if false && key == settingsJSONPath`) → `go test -count=1 -run 'TestMergeUserFiles_Snapshot' ./internal/cli/update/merge/`:
   ```
   settings_snapshot_base_test.go:105: statusLine.command = old, want new
   settings_snapshot_base_test.go:158: conflict line count = 0, want 1:
   settings_snapshot_base_test.go:249: user-deleted statusLine came back: map[model:sonnet statusLine:map[command:x]]
   settings_snapshot_base_test.go:275: conflict line count = 0, want 1:
   FAIL	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.455s
   ```
   이후 `git restore internal/cli/update/merge/merge.go` → `git status --short` 빈 출력.

## 기록된 증거 판독

- `slot-1-green.txt`: top-level 4개 + 하위 8개 `--- PASS`, 끝줄 `ok  	github.com/modu-ai/moai-adk/internal/cli	1.503s`.
- `slot-5-*.txt` 15개: 모두 `--- FAIL` 과 `update_settings_snapshot_test.go:<줄>` 값 단정 줄을 담는다. 빌드 오류(`undefined`, `build failed`) 0건.
- `slot-2-red-stub.txt`: 스텁 대비 `:296`, `:324`, `:355`, `:376`, `:407`, `:450` 값 단정 RED.
- `slot-4-win-build.txt`: 0바이트 — 빈 출력 자체는 성공/실패를 말하지 않는다. exit 0 은 `progress.md:281` 의 기록에 의존한다.

## Gaps (관측하지 않은 것)

- 루트 `./internal/cli` 패키지 시험·빌드·Windows 크로스 빌드는 레인 규칙상 재실행하지 않았다. cli AC(005·007·008 cli 셀)는 슬롯 증거 판독에 의존한다(명령 6 으로 트리 귀속만 확인).
- `internal/cli` 패키지 커버리지와 `-race` 는 측정되지 않았다.
- `origin/develop` CI 판정은 리드의 일괄 push 뒤에만 존재한다 — 관측 불가.
- run 단계 슬롯 중 관측된 `~/.moai` mtime 변화 두 건의 귀속은 이번 감사에서도 확인하지 않았다. 새 init 시험 셀은 기존 init 시험과 같은 `runInit` 경로를 타며, `homestate.EnsureProjectLayout` 은 임시 디렉터리 안 프로젝트에 대해 홈 레이아웃을 건너뛴다(`internal/homestate/paths.go:166`) — 이 카드가 원인이라는 증거도, 아니라는 증거도 없다.
- 교차 모델 감사(codex/glm)는 수행하지 않았다(Claude 단독 감사).

## 잔여 위험

- F-17(사용자가 지운 템플릿 키는 이후 수정도 받지 않음)은 수용된 한계다. 사용자가 무심코 지운 보안 관련 키가 이후 템플릿 수정을 영영 받지 못할 수 있다.
- 기계를 옮기면 렌더된 `env.PATH` 가 템플릿 변경으로 전달된다(acceptance §F). 사용자가 직접 고친 PATH 는 충돌 경로로 보존되지만, 고치지 않은 PATH 는 이제 update 마다 바뀐다 — 의도된 동작 변화이나 사용자 체감 변화다.
- 동시 update 는 잠금 없이 단일 사용자 모델을 전제한다.
