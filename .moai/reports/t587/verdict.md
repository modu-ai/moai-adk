# t587 verdict — moai update 수리 (F2·F5·F6·F7)

트리: `.claude/worktrees/t587`, 브랜치 `WT-update-repair`. 기반은 origin/develop `d060e0d13` 에 로컬 develop `93182d137` 을 흡수한 `01059305e`, 재현 증거 커밋은 `eabad2c59`.
범위: 리드 결정에 따라 F3 는 새 카드 t656 으로 분리됐다. 이 카드는 F2·F5·F6·F7 만 다룬다(Class B).
재현 방식: 리드 결정 A — internal/cli Go 테스트, 홈은 `userHomeDirFn` 시접과 `t.TempDir()` 로만 바꾸고 `t.Setenv("HOME", …)` 는 쓰지 않았다.

## 주장

1. F2 — 업데이트 결과 화면이 안내하는 복구 명령이 존재하지 않는 플래그를 가리켰다. 이제 안내 네 곳이 모두 실제 플래그 `--restore` 를 가리킨다.
2. F5 — update 가 템플릿을 렌더링할 때 프로젝트의 git 모드를 넘기지 않아 team·personal 프로젝트가 manual 로 렌더링됐다. 이제 동기화 경로와 v2→v3 재설치 경로 모두 git-strategy.yaml 의 모드로 렌더링한다.
3. F6 — 동기화의 정리 단계가 `.claude/skills/moai*` 를 지운 뒤에야 레거시 스킬 보관이 돌아 보관할 원본이 없었다. 이제 보관이 정리 단계 바로 앞에서 돈다.
4. F7 — `update -c` 의 모델 정책 저장이 system.yaml 읽기·파싱·쓰기 오류를 버렸고, 파싱 실패 시 파일을 덮어써 원래 내용을 잃었다. 이제 오류를 반환하고 파일을 건드리지 않는다.
5. 새 테스트 네 개는 수리 전 트리에서 모두 맞는 이유로 실패했고, 수리 뒤 이웃 테스트 64개와 함께 통과했다. 수리마다 넣은 뮤턴트가 해당 테스트를 빨갛게 만들었다(한 곳은 첫 위치에서 살아남아 위치를 바꿨다).
6. 실행 내내 운영자의 실제 홈 설정은 바뀌지 않았다.

## 증거

- 주장 1:
  - 재현: `repro/f2-restore-config.txt` — 트리 빌드 바이너리로 `moai update --restore-config <dir>` 실행 → `Unknown flag: --restore-config.`, exit 1. 대조 `repro/f2-update-help.txt` 에 `--restore` 존재, exit 0.
  - 테스트: `slot/red.txt` 에서 렌더 4곳과 소스 리터럴 4곳이 FAIL. `slot/green.txt` 에서 PASS.
  - 뮤턴트: `slot/m1.txt` FAIL.
- 주장 2:
  - RED: `slot/red.txt` — 동기화는 `git add` 줄은 있고 `git commit`·`git push` 줄이 없음, 재설치는 `GitMode = "manual"`.
  - GREEN: `slot/green.txt` PASS.
  - 뮤턴트: `slot/m2.txt` 생존, `slot/m2b.txt` FAIL. 이유는 `slot/summary.md` 뮤턴트 절.
- 주장 3:
  - RED: `slot/red.txt` — 보관 경로에 파일 없음.
  - GREEN: `slot/green.txt` PASS.
  - 뮤턴트: `slot/m3.txt` FAIL.
  - 건너뛴 동기화는 여전히 보관하지 않음(REQ-UAC-004): `slot/skip-sync.txt` PASS.
- 주장 4:
  - RED: `slot/red.txt` — 두 서브테스트 모두 nil, 깨진 파일이 `moai:\n    model_policy: high\n` 로 덮어써짐.
  - GREEN: `slot/green.txt` PASS.
  - 뮤턴트: `slot/m4.txt` 파싱 서브테스트 FAIL.
- 주장 5: `slot/summary.md` 실행 결과 표 — GREEN·최종 GREEN 모두 RUN 68 / PASS 68 / FAIL 0 / SKIP 0, 뮤턴트 복원 `cmp` exit 전부 0, `shasum -c` OK.
- 주장 6: `slot/summary.md` 홈 누출 안전장치 절 — `~/.claude/settings.json` sha256 `86e2d9b6…` 가 실행 전·RED 뒤·GREEN 뒤·최종 GREEN 뒤 네 번 모두 같고, `~/.claude/hooks/moai/` 는 계속 없음.

### 바뀐 파일

- `internal/cli/update/report/outcome.go` — 안내 3곳 `--restore`.
- `internal/cli/update_tux.go` — 안내 1곳 `--restore`.
- `internal/cli/update_tux_test.go` — 옛 문구 `--restore-config` 를 고정하던 기대값을 `--restore` 로.
- `internal/cli/update/backup/restore.go` — 없는 플래그를 적은 주석 1줄.
- `internal/config/loader_git_mode.go` (새 파일) — `LoadGitMode`. `LoadWorktreeBaseBranch` 와 같은 모양의 단일 키 읽기.
- `internal/cli/update_template_sync.go` — 정리 단계 전에 git 모드를 읽어 Validate·Deploy 컨텍스트에 넘김. 레거시 스킬 보관과 누락 보고를 정리 단계 앞으로 옮김.
- `internal/cli/update_clean_install.go` — 재설치 배포 컨텍스트에 git 모드 전달.
- `internal/cli/update.go` — 동기화 뒤에 돌던 보관 블록과 사전 스냅숏 제거.
- `internal/cli/update_wizard.go` — system.yaml 읽기·파싱·마샬·쓰기 오류 반환.
- `internal/cli/update_repair_test.go` (새 파일) — 테스트 네 개.

`internal/cli/testdata/tuxiu/**/update.*.stdout.golden` 6개에는 옛 문구가 그대로 남아 있다. 이 파일들은 과거 캡처끼리만 비교하는 고정 기록이라(`tuxiu_characterization_test.go:11-33`) 라이브 출력과 비교하지 않으므로 손대지 않았다. 최종 GREEN 에서 해당 테스트 네 개가 통과했다.

## 기준선 귀속

- 재현(F2): 트리 `01059305e` 에서 빌드한 바이너리, 빈 `/tmp/t587-repro/f2`.
- RED: `eabad2c59` + 미커밋 테스트 파일, 수리 전.
- GREEN·뮤턴트·최종 GREEN: `eabad2c59` + 미커밋 수리와 테스트. 이 기록을 담는 커밋과 코드·테스트 내용이 같다.
- 슬롯의 모든 실행 직전에 사전 확인을 했다. 다른 internal/cli 컴파일 0, 대조 zsh 36~38.
- grep 은 모두 `/usr/bin/grep` 이다.

## 미검증

- 이웃 테스트 셋은 돌리지 않았다: `TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently`, `TestRunUpdate_ThreeRunIdempotency_V3Project`, `TestReproduction_NonProjectDirectoryPollution_Issue1086`. 모두 `updateCmd.RunE` 로 update 전체를 돌리는데 홈 시접을 바꾸지 않아, 실제 `~/.claude/settings.json` 에 닿을 수 있다. F6 의 순서 변경을 update 전체 흐름에서 확인하는 테스트는 이 셋뿐이다.
- 같은 이유로 `TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive` 도 돌리지 않았다.
- `config.LoadGitMode` 에는 internal/config 패키지 단위 테스트가 없다. 동작은 internal/cli 테스트의 두 경로로만 확인했다.
- F7 쓰기 오류 분기는 테스트가 닿지 않는다. 경로를 디렉터리로 만든 픽스처는 읽기 단계에서 먼저 실패하므로 `atomicfile.Write` 오류 반환을 되돌리는 뮤턴트는 살아남을 것이다. 돌려 보지는 않았다.
- F5 는 설정 파일을 렌더링하는 순간만 확인했다. `git-strategy.yaml` 의 모드가 설정 복원(3-way 병합) 뒤에 그대로 남는지는 재지 않았다.
- CI(darwin·windows 매트릭스)는 보지 않았다.

## 잔여 위험

- **누락 손실 보고는 여전히 의미가 있다.** 판단 근거는 `slot/m3.txt` 32행이다. 보관이 돌지 않았을 때 `archived 0 of 1 skills present before sync` 경고가 찍혔다. 순서를 바꾼 뒤에도 보관에 실패한 스킬은 바로 뒤 정리 단계에서 지워지므로, 이 경고는 실제 손실을 알린다.
- **그러나 경고의 원인 설명은 이제 틀렸다.** `update_archive.go:424` 는 "정리 단계가 보관보다 먼저 원본을 지웠다"고 말하지만, 이제 그 순서로는 일어나지 않는다. 남은 원인은 보관 자체의 실패다. 문구 수정 여부는 리드 판단으로 남겼다. `TestReportArchiveShortfall` 은 "0 of 3" 과 `!` 표시만 고정하므로 문구를 바꿔도 깨지지 않는다.
- 렌더링을 실제로 결정하는 곳은 Validate 컨텍스트다(배포기의 렌더 캐시). 누군가 Validate 단계를 없애거나 배포기를 새로 만들면 Deploy 쪽 줄이 결정하게 된다. 두 곳 모두에 넘기도록 해 두었다.
- **코드만 읽어서 세운 가설:** 동기화의 Deploy 컨텍스트가 넘기는 `readHookOptInEnabled(projectRoot)` 는 정리 단계가 `.moai/config` 를 지운 뒤에 system.yaml 을 읽는다. 다만 위 캐시 때문에 실제 렌더링은 Validate 쪽 값(정리 전에 읽음)을 쓸 가능성이 높다. 이 카드 범위 밖이다.

## t656 입력 — F3 코드 판독 (실행하지 않음)

- 병합 base 는 「사용자 파일이 가진 키로 좁힌 새 템플릿」이다: `internal/cli/update/merge/base.go:112-131`, 호출은 `internal/cli/update/merge/merge.go:216`.
- 사용자가 건드리지 않은 공유 키의 **값**을 템플릿이 바꾸면 base 와 updated 가 같아져 사용자 쪽 옛값이 남는다. 파일 주석이 이 한계를 스스로 적고 있다(`base.go:29-34`). 3-way 판정은 `internal/merge/strategies.go:416-433`.
- 새로 생긴 키는 추가로 읽혀 전달되도록 이미 바뀌었다(`base.go:13-27`). 카드의 「영원히 전달 안 됨」은 값 변경에 대해서만 맞다.
- 스냅숏 base 선례: `internal/cli/update/backup/snapshot.go:22,51` — `.moai/cache/template-snapshot/sections/` 에 sections yaml 만 둔다(REQ-TBS-015 가 범위를 sections 로 제한). settings.json 에 적용하려면 이 범위 규칙과의 관계를 정해야 한다.
- 재설치 경로의 주석은 settings.json base 가 임베드 FS 에 없어(`.tmpl` 로 배포) 사용자 파일을 통째로 보존한다고 적는다(`internal/cli/update_clean_install.go:393-396`).
