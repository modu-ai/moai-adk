# t587 통합 창 재측정 기록

창 보유: lane-8(`moai integration acquire --name lane-8`). 리드 지명 후 acquire.
워크트리 `.claude/worktrees/t587`, 브랜치 `WT-update-repair`.
승인 범위: 창 안 internal/cli 4회 — 흡수 트리 GREEN, m5, m6, 최종 GREEN.

## 흡수

- 흡수 전: 리드 지시로 창 전 편집을 `6cb2b8e34` 로 커밋. 메시지에 "창 안 GREEN 대기" 명시.
- 흡수 대상: 로컬 `develop` `1bea05daafcb4d52d58ef57443da02dd564b1ea1`(리드가 알린 tip 과 `git rev-parse develop` 일치).
- 명령: `git merge --no-ff develop` → `../window-merge.log`, exit 0. 로그에 `CONFLICT`·`Automatic merge failed`·`Unable to write index`·`index.lock` 없음.
- 흡수 병합 커밋: `f7c21fe365725cb04616d14719fc170b5614651b`, 부모 `6cb2b8e34` · `1bea05daa`, 트리 `9050c7e231987f9da2889d5596e2e8cda05fa3e0`.
- 병합 직후 `git rev-parse -q --verify MERGE_HEAD` → 출력 없음, exit 1.
- develop 이 들여온 파일 157개(`git diff --name-only HEAD^1 HEAD | wc -l`). 그중 `internal/cli/update*`·`internal/config/`·`internal/template/deployer*`·`internal/cli/wizard*` 는 0건(grep exit 1).

## 매 실행 전 사전 확인

네 실행 모두 직전에 따로 실행했다.

- `pgrep -lx cli.test` → 출력 없음, exit 1
- `pgrep -lf 'go (test|build|vet).*internal/cli'` → 출력 없음, exit 1
- 대조 `pgrep -x zsh | wc -l` → 41·43·43·44

## 홈 누출 안전장치

창 전(스크래치 `window-pre.txt`), 흡수 트리 GREEN 뒤, m5·m6 뒤, 최종 GREEN 뒤에 같은 값이었다.

- `~/.claude/settings.json` sha256 `86e2d9b63abd4d027b4f65bcdc3b41e127b55c9d65c5c8061d17f36b247cc5eb`
- mtime: `~/.zshrc` 1786867378, `~/.zprofile` 1788158817, `~/.zshenv` 1786867378, `~/.bashrc` 1779596844, `~/.bash_profile` 1779596844, `~/.profile` 1779596844
- `~/.claude/hooks/moai/` 는 계속 없음(`ls` exit 1)

## 실행 결과

| 실행 | 파일 | exit | 판독 |
|---|---|---|---|
| 흡수 트리 GREEN(슬롯과 같은 정확 이름 패턴) | `green.txt` / `green.exit` | 0 | RUN 68 / PASS 68 / FAIL 0 / SKIP 0. `unwritable_directory_is_reported` PASS(skip 아님), `TestReportArchiveShortfall` PASS |
| m5 — `update_archive.go` 원인 문구를 옛 문장으로 되돌림 | `m5.txt` / `m5.exit` | 1 | `TestReportArchiveShortfall` FAIL(`update_observability_test.go:110` 새 단언), `TestUpdateRepair_*` 전부 PASS |
| m6 — `update_wizard.go` 쓰기 오류 반환을 `_ = atomicfile.Write(...)` 로 되돌림 | `m6.txt` / `m6.exit` | 1 | `unwritable_directory_is_reported` 만 FAIL(`update_repair_test.go:316`), 나머지 PASS |
| 최종 GREEN(같은 패턴) | `final-green.txt` / `final-green.exit` | 0 | RUN 68 / PASS 68 / FAIL 0 / SKIP 0, 두 새 단언 PASS |

m5·m6 선택자는 `^(TestReportArchiveShortfall|TestUpdateRepair_.*)$`.

m6 에서 기존 픽스처 `unwritable_path_is_reported`(경로를 디렉터리로 만든 것)는 PASS 였다. 그 픽스처는 읽기 단계에서 먼저 실패해 쓰기 분기에 닿지 않는다는 슬롯 판독이 실행으로 확인됐다.

각 뮤턴트는 적용 직후 `/usr/bin/grep -c` 로 바뀐 줄을 세고 gofmt 가 통과한 뒤 실행했다.

## 복원

- 백업: 세션 스크래치패드의 `cp -p` 사본, 해시 `pre-mutant-sha.txt`.
- `cmp` exit: m5 0, m6 0.
- `shasum -a 256 -c pre-mutant-sha.txt` → 두 파일 OK, exit 0.
- 뮤턴트는 커밋하지 않았다.

## 재측정 트리

- 네 실행 모두 흡수 병합 커밋 `f7c21fe36` 의 트리 `9050c7e2…` 에서 했다(최종 GREEN 뒤 `git status --short` 는 이 증거 파일만 미추적).
- 이 기록을 담는 증거 커밋은 `.moai/reports/t587/` 아래만 더한다.

## 미실행

- `TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently`, `TestRunUpdate_ThreeRunIdempotency_V3Project`, `TestReproduction_NonProjectDirectoryPollution_Issue1086`, `TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive` — 홈 시접 없음, 리드 결정으로 미실행, t661 소관.
