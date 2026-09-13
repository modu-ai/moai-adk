# t587 internal/cli 슬롯 기록

트리: 워크트리 `.claude/worktrees/t587`, 브랜치 `WT-update-repair`, HEAD `eabad2c59` + 미커밋 수리·테스트.
승인: 리드 1회 슬롯, 한 세션 안 직렬 묶음(RED → 수리 → GREEN+이웃 → 뮤턴트 → 최종 GREEN). 전체 스위트는 돌리지 않았다.

## 매 실행 전 사전 확인

모든 `go test` 직전에 같은 세 줄을 따로 실행했다.

- `pgrep -lx cli.test` → 출력 없음, exit 1
- `pgrep -lf 'go (test|build|vet).*internal/cli'` → 출력 없음, exit 1
- 대조 `pgrep -x zsh | wc -l` → 36~38 (pgrep 이 살아 있는 프로세스를 본다는 증거)

9번 실행 모두 다른 레인의 internal/cli 컴파일은 0이었다. 대기한 적은 없다.

## 홈 누출 안전장치

- 실행 전: `~/.claude/settings.json` 을 세션 스크래치패드로 복사하고 sha256 `86e2d9b63abd4d027b4f65bcdc3b41e127b55c9d65c5c8061d17f36b247cc5eb` 를 기록했다. `~/.claude/hooks/moai/` 는 실행 전부터 없었다(`ls` exit 1).
- 테스트 쪽: 새 테스트 네 개가 모두 `isolateHome(t)` 로 시작한다. 시접 홈이 센티널과 다르거나 `os.UserHomeDir()` 와 같으면 즉시 `t.Fatal` 한다.
- RED 뒤, GREEN 뒤, 최종 GREEN 뒤: sha256 이 세 번 모두 같았고 `hooks/moai` 는 계속 없었다. 누출 흔적은 없다.

## 실행 결과

| 단계 | 파일 | exit | 판독 |
|---|---|---|---|
| RED (`-run 'TestUpdateRepair_'`) | `red.txt` / `red.exit` | 1 | 4개 모두 FAIL, 이유는 아래 절 |
| GREEN (정확한 이름 42개 패턴, 최상위 68개) | `green.txt` / `green.exit` | 0 | RUN 68 / PASS 68 / FAIL 0 / SKIP 0 |
| 건너뛴 동기화 서브테스트만 | `skip-sync.txt` / `skip-sync.exit` | 0 | `TestSkipSyncNoArchive/version_match_returns_skipped_true` PASS |
| m1 | `m1.txt` / `m1.exit` | 1 | 아래 표 |
| m2 | `m2.txt` / `m2.exit` | 0 | **생존** — 아래 표 |
| m2b | `m2b.txt` / `m2b.exit` | 1 | 아래 표 |
| m4 | `m4.txt` / `m4.exit` | 1 | 아래 표 |
| m3 | `m3.txt` / `m3.exit` | 1 | 아래 표 |
| 최종 GREEN (GREEN 과 같은 패턴) | `final-green.txt` / `final-green.exit` | 0 | RUN 68 / PASS 68 / FAIL 0 / SKIP 0 |

### RED 가 실패한 이유 (`red.txt`)

- F2 `RecoverHintFlagExists`: 렌더 4곳과 소스 리터럴 4곳이 모두 `--restore-config` 를 가리켰다. 안내 문구 렌더(도달성)는 통과했다.
- F5 `RenderUsesProjectGitMode/template_sync`: `"Bash(git add:*)"` 는 있었고(도달성 통과) `git commit`·`git push` 줄이 없었다. `/clean_reinstall`: `TemplateContext.GitMode = "manual"`.
- F6 `SyncArchivesLegacySkillsBeforeCleanup`: 보관 경로 `.moai/archive/skills/v2.16/moai-domain-mobile/SKILL.md` 가 없었다.
- F7 `WizardModelPolicySurfacesSystemYAMLErrors`: 두 서브테스트 모두 nil 이 반환됐다. 깨진 `system.yaml` 은 `moai:\n    model_policy: high\n` 로 덮어써져 원래 내용이 사라졌다.

### 뮤턴트

| 뮤턴트 | 바꾼 것 | 기대 | 결과 |
|---|---|---|---|
| m1 | `update/report/outcome.go` AlreadyUpToDate 안내를 `--restore-config` 로 되돌림 | F2 FAIL | `RecoverHintFlagExists` FAIL(`rendered_report_outcomes`, `source_literals`), 나머지 PASS |
| m2 | `update_template_sync.go` Deploy 컨텍스트의 `WithGitMode(gitMode)` 제거 | F5 FAIL | **생존** — 4개 전부 PASS |
| m2b | 같은 파일 Validate 컨텍스트의 `WithGitMode(gitMode)` 제거 | F5 FAIL | `RenderUsesProjectGitMode/template_sync` FAIL, `clean_reinstall` 포함 나머지 PASS |
| m4 | `update_wizard.go` 파싱 오류를 `_ = yaml.Unmarshal(...)` 로 되돌림 | F7 FAIL | `unparseable_file_is_reported_and_kept` FAIL, `unwritable_path_is_reported` PASS |
| m3 | `update_template_sync.go` 정리 단계 앞 `archiveLegacySkills(...)` 를 `0, error(nil)` 로 대체 | F6 FAIL | `SyncArchivesLegacySkillsBeforeCleanup` FAIL, 나머지 PASS |

m2 가 살아남은 이유: 템플릿 배포기가 `ValidateAll` 의 렌더 결과를 캐시하고, `Deploy` 는 캐시가 비었을 때만 다시 렌더링한다(`internal/template/deployer.go:81-88`, `:193-199`, `:365-374`). 동기화는 같은 배포기 인스턴스로 Validate 뒤에 Deploy 를 돌리므로, 렌더링을 결정하는 컨텍스트는 Validate 쪽이다. m2b 가 그 자리를 잘라 FAIL 을 냈다. Deploy 컨텍스트의 줄은 캐시가 비었을 때를 위해 남겨 두었다.

각 뮤턴트는 적용 직후 `/usr/bin/grep` 으로 바뀐 줄을 확인하고 gofmt 가 통과한 뒤에 실행했다.

## 복원

- 백업은 세션 스크래치패드의 `cp -p` 사본(`outcome.go`, `update_template_sync.go`, `update_wizard.go`), 해시는 `pre-mutant-sha.txt`.
- 뮤턴트마다 복원 후 `cmp` exit: m1 0, m2 0, m2b 0, m4 0, m3 0.
- 최종 `shasum -a 256 -c pre-mutant-sha.txt` → 세 파일 OK, exit 0.
- 뮤턴트는 커밋하지 않았다.

## 선택자에서 뺀 이웃 테스트

- `TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently`, `TestRunUpdate_ThreeRunIdempotency_V3Project`, `TestReproduction_NonProjectDirectoryPollution_Issue1086`: `updateCmd.RunE` 로 update 전체를 돌리는데 파일 안에 홈 시접 교체가 없다(`userHomeDirFn`·`homeSeamSpy` grep 0건). 실제 `~/.claude/settings.json` 에 닿을 수 있어 돌리지 않았다.
- `TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive`: `--force` 로 동기화 전체를 돌 수 있는데 시접이 없다. 같은 이유로 돌리지 않았고, 버전 일치 서브테스트만 따로 돌렸다.
- 처음 요청문에 적은 선택자(`TestUpdateSkipSync`, `TestCleanReinstall` 등)는 실제 함수 이름과 달라 일부 테스트를 뽑지 못했을 것이다. 실행 전에 함수 이름을 grep 해 정확한 이름 목록으로 바꿨다.
