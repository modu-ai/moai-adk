# t587 재현 기록 (1차 — 재현 방식 결정 대기)

트리: `.claude/worktrees/t587`, 브랜치 `WT-update-repair`. 기반: origin/develop `d060e0d13` + 로컬 develop `93182d137` 흡수(`01059305e`, HEAD^2 = `93182d137`).
바이너리: 이 트리에서 빌드했다(`go build`, Makefile과 같은 ldflags 값 — Version `moai_cp/20260910_130400`, Commit `01059305e`). 출력은 세션 스크래치패드 `t587-bin/moai`, exit 0. 빌드 전 확인: `pgrep -f 'internal/cli|cli\.test'` 출력 없음.

## 1. F2 — 안내하는 복구 플래그가 없다: 재현됨

- 명령: 빈 디렉터리 `/tmp/t587-repro/f2` 에서 `moai update --restore-config /tmp/t587-repro/nonexistent-backup`
- 결과: `repro/f2-restore-config.txt` — `Unknown flag: --restore-config.`, exit 1
- 대조군: 같은 바이너리 `moai update --help` → `--restore  Restore .moai/config from a backup directory ...`, exit 0 (`repro/f2-update-help.txt`)
- 실패한 실행은 디렉터리에 아무것도 만들지 않았다(`repro/f2-created-by-failed-run.txt`, 빈 목록). 앞서 파이프로 센 `3` 은 셸 `ls` alias 가 붙인 `total`/`.`/`..` 줄이었다.
- 카드가 짚은 곳 외의 형제 위치: 안내 문구는 `internal/cli/update_tux.go:187` 한 곳이 아니라 `internal/cli/update/report/outcome.go:44,56,69` 에도 있다(총 4곳). 실제 플래그 정의는 `internal/cli/update.go:86`.

## 2. 실행 재현을 멈춘 이유

init 과 update 는 `/tmp` 프로젝트에서 돌려도 홈 디렉터리에 쓴다.

- `internal/cli/init.go:964`, `internal/cli/update_template_sync.go:567` → `ensureGlobalSettingsEnv()`(`update.go:958`)는 `~/.claude/hooks/moai` 를 지우고(`update.go:967-970`) `~/.claude/settings.json` 의 moai 관리 키를 정리한다.
- `internal/cli/init.go:877` → `homestate.EnsureProjectLayout` 는 프로젝트가 `os.TempDir()` 밖이면 홈 레이아웃을 만든다(`internal/homestate/paths.go:166-170`). macOS 의 `os.TempDir()` 는 `/var/folders/...` 이므로 `/tmp` 는 밖이다.
- `internal/cli/init.go:889-897` → 자율성 번들이 사용자 범위 `~/.claude/settings.json` 에 권한 블록을 이어 붙인다(기본 semi-auto 는 변경 0 이라고 적혀 있으나 실측하지 않았다).

`HOME=/tmp/...` 로 격리한 실행은 세션 가드가 거부했다(「sets HOME, injecting git configuration」). 격리 없이 돌리면 운영자의 실제 전역 설정을 바꾸게 되므로 F3·F5·F6 실행 재현은 하지 않았다.

## 3. 코드 판독(실행 안 함 — 가설 단계)

| 항목 | 현재 트리에서 읽은 것 | 카드 전제와의 차이 |
|---|---|---|
| F3 | 병합 base 는 「사용자 파일이 가진 키로 좁힌 새 템플릿」(`update/merge/base.go:112-131`, 호출 `update/merge/merge.go:216`). 사용자가 건드리지 않은 공유 키의 **값**을 템플릿이 바꾸면 base 와 updated 가 같아져 사용자 쪽 옛값이 남는다 — 파일 주석이 이 한계를 스스로 적고 있다(`base.go:29-34`). 3-way 판정은 `internal/merge/strategies.go:416-433`. 스냅숏 base 선례는 `update/backup/snapshot.go:51`(sections yaml 만). | 「템플릿 변경이 영원히 전달 안 됨」은 **값 변경**에 대해서만 맞다. 새로 생긴 키는 추가로 읽혀 전달되도록 이미 바뀌었다(`base.go:13-27`). |
| F5 | update 의 템플릿 컨텍스트 3곳이 `WithGitMode` 를 부르지 않는다: `update_template_sync.go:271`(검증), `:331`(배포), `update_clean_install.go:445`. 기본값은 `manual`(`internal/template/context.go:95`). 모드에 따라 갈리는 템플릿: `settings.json.tmpl:456,459`, `git-strategy.yaml.tmpl:5`. init 은 `core/project/initializer.go:403,472` 에서 넘긴다. | 위치 수는 카드와 같다. 사용자 파일에 실제로 드러나는지는 이후 병합·복원 단계에 달려 있어 실측이 필요하다. |
| F6 | `.claude/skills/moai*` 삭제(`update/deploy/deploy.go:70-74`, 동기화 단계 `update_template_sync.go:315`)가 `archiveLegacySkills`(`update.go:568`)보다 먼저 돈다. 대상 id 는 `update_archive.go:45-`. | 「조용히 통과」는 낡았을 수 있다 — 누락분을 손실로 알리는 `reportArchiveShortfall` 가 붙어 있다(`update.go:572-575`). 순서 자체는 그대로다. 줄 번호 드리프트: 카드 554 → 현재 568. |
| F7 | `update_wizard.go:318` `_ = yaml.Unmarshal(...)`, `:330` `_ = atomicfile.Write(...)` 가 그대로 있다. | 전제 유지. 대화형 위저드(`update -c`) 안에서만 닿으므로 바이너리 실행 재현보다 함수 단위 테스트가 맞다. |

## 4. 결정이 필요한 것

1. 재현 방식(위 2절).
2. F2 수리 방향: 안내 문구 4곳을 `--restore` 로 고칠지, `--restore-config` 별칭을 추가할지.
3. F6 수리 방향: 순서 교환인지, 스냅숏 기반 보관인지(카드가 둘 다 제시).
4. F3 는 settings.json 에 배포 시점 스냅숏 base 를 도입하는 설계 변경이다. 스냅숏 위치·범위(sections 전용 `SnapshotSubdir` 규칙과의 관계)를 정해야 한다.
