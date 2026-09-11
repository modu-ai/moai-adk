# Plan — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

카드 **t656** · Tier **M** · 기준 트리 `04a8ab731` (브랜치 `WT-update-value-merge`, `internal/` 은 `81c1d58f9` 와 동일) · v0.4.0 — plan-audit 2회차(FAIL 0.75) 반영, 운영자 승인 1회 연장 최종본

## §A Context

### Tier 판정

Tier **M**. 운영 코드 변경은 한 서브시스템(`internal/cli/update` 와 그 호출 지점)에 머물고, 예상 변경 파일은 테스트를 빼면 8개 안팎이다.

- `update/merge/merge.go`, `update/merge/base.go`
- 대기본·확정본 헬퍼 1~2개(`update/backup`)
- 흐름 시접 1개
- `update.go`(남은 대기본 판정 지점), `update_template_sync.go`, `update_clean_install.go`, init 기록·판정 지점

새 저장물은 기존 캐시 루트 아래 파일 둘이다. plan-auditor 통과 기준은 0.80 이다.

### 이 트리에서 읽은 사실 (실행 없이 판독만)

| 사실 | 위치 |
|---|---|
| `MergeUserFiles(projectRoot, backups, out)` 는 `projectRoot` 를 이미 받는다. 확정본을 찾는 데 시그니처를 바꿀 필요가 없다 | `internal/cli/update/merge/merge.go:173` |
| 병합의 보존 경로는 세 분기다. 배포 파일을 읽지 못하면 사용자 파일을 씀(`:197-204`), base 를 만들 수 없으면 사용자 파일을 씀(`:217-225`), 병합이 실패하면 사용자 파일을 씀(`:229-237`). 사용자 파일이 렌더와 같으면 쓰지 않고 넘어감(`:207`) | `merge.go` |
| settings.json 은 `templateManaged` 판별을 통과한다 | `base.go:53-60`, `base_test.go:143-159` |
| `moai update` 명령의 순서는 다음과 같다. ① `--binary` 조기 반환(`update.go:330`) ② `--dry-run` 조기 반환(`:335-365`) ③ 폐기 v2 deny 규칙 제거(`:384`, 살아 있는 settings.json 을 다시 쓸 수 있는 첫 단계) ④ v2 분기 → clean-reinstall(`:405`, `:420`) ⑤ 템플릿 동기화(`:489`) | `update.go` |
| 템플릿 동기화 안에서 버전이 같으면 건너뛴다(`update_template_sync.go:98-105`, `:641`). 그 뒤 순서는 Backup 단계의 사용자 파일 읽기(`:491`), 정리(`:329-337`), 배포(`:364`), Restore Settings 병합(`:544`) | `update_template_sync.go` |
| clean-reinstall 은 사용자 파일 읽기(`:400`), 배포(`:459`), 병합(`:507`), 폐기 deny 규칙 제거(`:531`) 순이다 | `update_clean_install.go` |
| update 정리 직전 사용자 settings.json 은 실행 백업의 `in-memory-backups/` 에 복사된다 | `update_disk_backup.go:47-53`, `:126-131` |
| `moai update --restore` 는 `RestoreFromBackupDir` 를 거쳐 `RestoreMoaiConfig` 만 두 번 부른다 | `internal/cli/update/backup/restore_entry.go:47-79` |
| init 은 비강제 배포기를 쓴다. 기존 파일은 매니페스트 기록이 없거나 사용자 소유면 건너뛴다. 기록이 없던 파일은 `UserCreated` 로 기록된다 | `init.go:821`, `deployer.go:239-255` |
| init 은 초기화 실행(`init.go:867`) 뒤 `ApplyAutonomyTierBundle`(`:889-900`)이 조건부로 프로젝트 settings.json 을 다시 쓴다 | `init.go`, `autonomy_bundle.go:59-94` |
| sections 스냅숏 읽기는 `sections/` 아래만 본다 | `backup/snapshot.go:114-131`, `backup/base_loader.go:26-67` |
| clean-reinstall 테스트 대역 `stubDeployer.Deploy` 는 파일을 쓰지 않는다 | `update_clean_install_test.go:40-52` |

### 레인 제약 — `internal/cli` 컴파일·테스트 슬롯

[HARD] `internal/cli` 패키지를 컴파일하거나 테스트하려면 **리드의 슬롯 승인**이 필요하다. run 단계는 M4 에 들어가기 전에 이 승인을 요청하며, 승인 없이 `go test ./internal/cli` 나 `go build ./internal/cli/...` 를 돌리지 않는다. 테스트는 `internal/cli/update/merge` 와 `internal/cli/update/backup` 에 두는 것을 우선한다. M1–M3 은 이 두 패키지만으로 끝나도록 짰다.

## §B Known Issues

- **B1 — 같은 흐름 안의 덮어쓰기 (종결: 대기본·확정본 분리).** 뮤턴트 M-06c, M-06t, M-06t-w 로 반증한다.
- **B2 — init 의 후속 재기록.** 대기본은 `ApplyAutonomyTierBundle` 보다 앞서 기록한다(REQ-USB-003).
- **B3 — 승격 조건 (종결: D5 운영자 결정).** 중단 시 살아 있는 파일은 순수 렌더다(감사 F-07). 확정 규칙과 판정 위치는 §E Decision D5.
- **B4 — 형제 SPEC (종결: D6 운영자 결정).**
- **B5 — hook-delivery 특성화 테스트.** `base_test.go:269-332` 는 유도 base 경로의 동작을 고정하며, 이 SPEC 은 그 테스트를 뒤집지 않는다. M3 에서 @MX:NOTE 를 붙인다.
- **B6 — 배포가 파일을 건너뛰는 init (운영자 확정).** REQ-USB-016 과 Decision D7 이 막는다.
- **B7 — 낡은 코드 주석.** `update_clean_install.go:394-397` 을 M4 에서 고친다.
- **B8 — 판정 위치 (종결: N-02 운영자 지시).** 2회차 plan 은 남은 대기본을 "다음 흐름이 새 대기본을 기록하기 전"에 판정한다고 적었다. 그러나 모든 흐름에서 그 지점은 살아 있는 파일이 이미 지워지거나 다시 쓰인 뒤다. 그래서 새 렌더와 비교하게 되고, "중단·복원 없음 → 승격"이 폐기로 바뀐다. 이번에 판정 위치를 흐름의 첫 재기록 앞으로 옮겼다(Decision D5, D4).

## §C Pre-flight (run 단계 진입 점검)

```bash
# 1. 기준 트리 — 흡수한 ref 와의 merge-base 부터 잰다 (gitflow-lane-protocol §8)
git merge-base develop HEAD
git diff --name-only <위 SHA>..HEAD          # 대조군: 1줄 이상

# 2. 인용이 아직 유효한가
grep -n "func MergeUserFiles" internal/cli/update/merge/merge.go
grep -n "stripRetiredV2DenyEntries(cwd\|detectV2Fingerprint(cwd)\|runTemplateSyncWithProgress(cmd)" internal/cli/update.go
grep -n "\"Deploy Templates\"\|updatemerge.MergeUserFiles\|packageVersion == projectVersion" internal/cli/update_template_sync.go
grep -n "deployWithMirrorNotice\|updatemerge.MergeUserFiles\|stripRetiredV2DenyEntries" internal/cli/update_clean_install.go
grep -n "NewDeployerWithRenderer\|executor.Execute\|ApplyAutonomyTierBundle" internal/cli/init.go

# 3. 캐시 루트가 git 무시 대상인가
grep -n "^\.moai/cache/$" .gitignore internal/template/templates/.gitignore

# 4. 기준 스위트와 커버리지 — 슬롯 불필요 패키지만
go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/
```

## §D Constraints (Hard)

1. `internal/merge` 는 수정하지 않는다(REQ-USB-014).
2. sections 스냅숏 코드(`WriteSnapshot`·`HasSnapshot`·`SaveTemplateBase`)의 본문·시그니처와 호출 지점 동작을 바꾸지 않는다(REQ-USB-013).
3. `internal/template/templates/` 는 건드리지 않는다.
4. 대기본 기록, 승격, 남은 대기본 판정, 확정본 읽기 실패는 모두 비차단이다(REQ-USB-009, REQ-USB-010).
5. 테스트는 `t.TempDir()` 과 `userHomeDirFn` 시접만 쓴다. `t.Setenv("HOME", ...)`, 실제 홈 접근, 실제 `moai update`·`moai init` 실행은 금지다.
6. `internal/cli` 컴파일·테스트는 리드 슬롯 승인 뒤에만 한다.
7. 로컬 전체 스위트(`go test ./...`)는 돌리지 않는다.

## §E Decisions (바뀔 가능성이 큰 순서)

### Decision D5 — 대기본 승격 조건과 판정 위치 (운영자 결정 2026-09-11, 판정 위치는 N-02 운영자 지시)

**규칙.** 대기본은 흐름이 끝났을 때 살아 있는 `.claude/settings.json` 이 그 흐름의 렌더를 **반영**하고 있을 때만 확정본으로 승격한다. 반영하고 있지 않으면 이전 확정본을 그대로 둔다. REQ-USB-005 가 이것을 GEARS 문장 셋으로 적는다.

**결정 기록.**

- 1회차 권고안 (a)의 목적은 새 템플릿 키가 사용자의 삭제로 잘못 읽히는 일을 절대 만들지 않는 것이었다. 이 목적은 그대로 지켜진다.
- "배포 뒤 병합 전에 중단되면 이전 스냅숏을 유지한다"는 1회차 설명은 틀린 전제에 기대고 있었다. 그 설명은 이 규칙으로 대체되었다(운영자 결정 2026-09-11).
- **운영자 규칙의 구현 방식, 리드 수용 (2026-09-11).** 판정은 두 시점에서 한다. 정상 종료 흐름은 그 흐름 끝에서, 중단된 흐름은 다음 흐름에서 판정한다. 남은 대기본의 판정 위치는 N-02 배치, 즉 **다음 흐름의 어느 단계도 살아 있는 settings.json 을 지우거나 다시 쓰기 전**이다. 이전 배치("다음 흐름이 새 대기본을 기록하기 전")를 대체한다.

**정상 종료 흐름의 판정 신호 (N-10).** 바이트 비교가 아니라 병합이 **보존 경로**(`merge.go:197-204`, `:217-225`, `:229-237`)를 탔는지를 신호로 쓴다.

- 보존 경로를 탔으면 대기본을 버린다.
- 사용자 파일과 렌더가 같아 건너뛴 경우(`:207`), 병합 대상이 없는 경우(init, 사용자 파일이 원래 없음)는 보존 경로가 아니므로 승격한다.
- 바이트 비교(살아 있는 파일 = 흐름 이전 사용자 파일 → 보존으로 간주)로 구현하면 경우 3 이 폐기되며, 뮤턴트 M-D5i 가 이를 잡는다.

**남은 대기본의 판정 위치 (N-02) — 선택과 근거.** 두 후보 가운데 **"다음 흐름의 어느 단계도 살아 있는 파일을 지우거나 다시 쓰기 전"**을 택한다.

- `moai update` 에서는 `runUpdate` 의 `--binary`·`--dry-run` 조기 반환 뒤, 폐기 v2 deny 규칙 제거(`update.go:384`) 앞의 한 지점이다. dry-run 은 아무것도 바꾸면 안 되므로 그 뒤다.
- `moai init` 에서는 초기화 실행(`init.go:867`) 앞이다.

근거는 이 트리에서 직접 읽은 코드다.

1. `update.go:384` 는 `moai update` 가 살아 있는 settings.json 을 다시 쓸 수 있는 첫 단계다. 그 앞 한 지점이 v2 분기(`:405` → `runCleanReinstall` `:420`, 그 안의 배포 `update_clean_install.go:459`)와 템플릿 동기화(`:489`, 그 안의 정리 `update_template_sync.go:337` 과 배포 `:364`)를 모두 덮는다.
2. 두 버전 일치 건너뛰기(`update_template_sync.go:98-105`, `:641`)도 이 지점보다 뒤다. 따라서 버전이 같아 동기화를 건너뛰는 update 에서도 남은 대기본이 판정된다(감사 N-09 종결).
3. 기각한 후보 "배포 전 백업 바이트와 비교"는 두 가지 이유로 맞지 않는다.
   - 백업 읽기(`update_template_sync.go:491`, `update_clean_install.go:400`)는 `update.go:384` 의 재기록보다 뒤에 있다.
   - `:491` 은 버전 일치 건너뛰기보다도 뒤라서 건너뛰는 update 에서는 실행되지 않는다. init 에는 백업 읽기가 아예 없어 판정 지점을 따로 하나 더 내야 한다.

**경우별 결과**

| # | 경우 | 흐름 끝의 살아 있는 파일 | 결과 | 판정 시점 |
|---|---|---|---|---|
| 1 | 사용자 파일 있음, 병합이 결과를 씀 | 렌더를 반영한 병합 결과 | 승격 | 흐름 끝 |
| 2 | 병합 실패, 사용자 파일을 통째로 보존 | 사용자 파일 | 이전 확정본 유지, 대기본 폐기 | 흐름 끝 |
| 3 | 사용자 파일이 렌더와 바이트 동일(병합 건너뜀) | 렌더 | 승격 | 흐름 끝 |
| 4 | 배포 뒤 병합 전에 중단, 복원 없음 | 순수 렌더 | 승격 | 다음 흐름, 첫 재기록 전 |
| 5 | 4 뒤에 사용자 파일이 되돌려짐 | 사용자 파일 | 이전 확정본 유지, 대기본 폐기 | 다음 흐름, 첫 재기록 전 |
| 6 | init 이 파일을 실제로 씀(이후 자율성 번들이 고칠 수 있음) | 렌더 또는 렌더+번들 | 승격 | 흐름 끝 |
| 7 | init 이 기존 파일을 건너뜀 | 사용자 파일 | 대기본 없음(REQ-USB-016) | — |
| 8 | update, 사용자 settings.json 이 원래 없음 | 렌더 | 승격 | 흐름 끝 |

**중단 뒤 끼어든 쓰기 (N-08, 안전한 쪽으로 벗어남).** 경우 4 에서 다음 흐름이 시작하기 전에 살아 있는 파일에 쓰기가 일어나면(손 편집, Claude Code 의 프로젝트 설정 기록, `moai tool-policy build`, 자율성 번들) 바이트 비교가 실패해 승격이 폐기로 바뀐다.

- 운영자 규칙 문구와 다른 결과이므로 spec.md §E 에 알려진 한계로 적었다.
- 사용자 데이터는 잃지 않는다. 중단된 렌더의 템플릿 변경이 사용자 변경으로 읽혀, 같은 leaf 의 이후 템플릿 변경이 충돌로 남는다.
- 다른 결과를 원하면 운영자에게 올려야 하며, 이 plan 에서 조용히 바꾸지 않는다.

**run 단계 M1 검증 항목 — 복원 동작.** `moai update --restore` 는 `RestoreMoaiConfig` 만 부른다(`restore_entry.go:47-79`, 직접 읽음). 감사 판독에 따르면 그 함수는 `.moai/config` 만 쓴다.

- 그렇다면 경우 5 는 사용자가 손으로 파일을 되돌린 경우에만 생긴다.
- 복원 명령 뒤에는 살아 있는 파일이 렌더로 남아 경우 4 와 같아진다(승격). 규칙은 그대로 옳다.
- M1 에서 `restore.go` 를 읽거나 backup 패키지 테스트로 관측해 `progress.md` §E.2 에 기록한다.

**리드 요구 뮤턴트.** M-D5a("항상 승격")를 경우 2 에 적용하면 보존 흐름의 렌더가 확정본이 된다. 사용자가 JSON 을 고친 뒤 다음 update 에서 그 렌더가 새로 들인 키가 base 에는 있고 사용자 파일에는 없으므로 사용자 삭제로 읽혀 도착하지 않는다. AC-USB-016 c2 가 RED 로 잡는다.

### Decision D6 — 형제 SPEC 과의 관계 (운영자 결정, 2026-09-11)

- **해석.** `SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` REQ-UMC-010 을 그 SPEC 의 REQ-UMC-008/009 개선에 한정한다. 그 SPEC 의 문장·제외 항목·HISTORY 는 이미 개정되었고 이번에 바꾸지 않았다.
- **착지 순서.** 이 SPEC 이 먼저 착지하고, t576 의 M2.1 재개 때 그 SPEC §A.6 전제를 다시 잰다.

### Decision D1 — 하위 경로: 확정본 `claude/settings.json`, 대기본 `claude/settings.json.pending`

같은 디렉터리의 `.pending` 접미사로 두면 승격이 같은 디렉터리 안의 교체가 되고, `*.json` 을 찾는 도구가 대기본을 확정본으로 착각하지 않는다.

### Decision D3 — 같은 흐름의 base 보존: 대기본 → 승격

병합은 늘 확정본 경로를 읽는다. 같은 흐름의 대기본은 병합 단계가 끝나기 전에는 확정본 자리에 오지 않는다. `MergeUserFiles` 의 시그니처와 base 주입 방식은 바뀌지 않는다.

### Decision D7 — "배포가 settings.json 을 실제로 썼는가"의 판정 근거 (REQ-USB-016)

디스크의 현재 파일을 다시 읽는 방식은 금지한다. **근거 2(매니페스트 기록 + 해시)를 우선한다(감사 N-11).**

- **근거 2 (우선).** 배포 뒤 매니페스트 항목이 `TemplateManaged` 이고, 기록된 해시가 배포가 쓴 렌더 바이트의 해시와 같을 때만 "썼다"로 본다. 건너뛴 파일은 기록이 없던 경우 `UserCreated` 로 기록되고, 이미 사용자 소유였던 경우 그 기록이 남는다(`deployer.go:243-253`). 렌더 바이트는 배포기에서 직접 받는다.
- **근거 1 (차선).** `DeployResult.ProtectedSkips` 는 필드 설명이 발행 스킬 경로 건너뜀만 기록한다고 적는다(`skill_mirror.go:69-75`). 실제로는 비강제 모드에서 모든 보호 건너뜀을 기록한다(`deployer.go:245`, `:253`). 이 근거를 쓰면 필드 설명을 함께 고치고, init 에서는 결과를 호출 지점까지 넘기는 경로를 새로 내야 한다(`initializer.go:423-430`).

어느 쪽이든 AC-USB-014 의 M-14 가 RED 여야 한다.

### Decision D2 — base 선택 규칙

- 유효한 확정본의 조건은 존재, 읽힘, JSON 객체 해석 셋이다. 하나라도 어긋나면 유도 base 로 간다.
- 확정본이 유효해도 사용자 파일이나 새 렌더가 JSON 으로 해석되지 않으면 지금의 보존 폴백을 따르며, 이는 D5 의 보존 경로다.

### Decision D4 — 판정·기록·승격 지점

| 흐름 | ① 남은 대기본 판정 (첫 재기록 전) | ② 대기본 기록 (배포 성공 뒤, 첫 재기록 전) | ③ 승격 판정 (흐름 끝) | 금지 위치 |
|---|---|---|---|---|
| `moai update` (template_sync·clean-reinstall 공통) | `runUpdate` 의 조기 반환 뒤, `update.go:384` 앞 한 곳 | — | — | 백업 단계(`update_template_sync.go:491`), 배포 뒤 |
| 일반 update(template_sync) | ①을 공유 | Deploy Templates 배포 성공 뒤(`update_template_sync.go:364-368`) | Restore Settings 의 settings.json 병합 뒤, 백업 경로 유무와 무관 | 기록을 `if configBackupPath != ""` 블록(`:497-532`) 안에 두기 |
| clean-reinstall | ①을 공유 | Step 5 배포 성공 뒤(`update_clean_install.go:459-462`), 병합(`:507`) 앞 | Step 5.5 병합 뒤 | 기록을 `:531` 뒤에 두기 |
| `moai init` | `init.go:867` 앞 | `executor.Execute` 성공 뒤, 오류 분기(`:868-876`) 밖, `ApplyAutonomyTierBundle`(`:889`) 앞 | init 흐름 끝(병합 없음) | 오류 분기 안, `:1030` |

- ②는 세 흐름이 하나의 헬퍼를 부른다(fan_in 3). ①은 두 지점(update, init)이 하나의 판정 헬퍼를 부른다.
- `runUpdateRestore` 는 렌더를 하지 않으므로 ①②③ 어느 지점도 아니다.

### Decision D8 — 흐름 시접과 관측 훅 (감사 F-02, N-05)

"남은 대기본 판정 → (배포) → 대기본 기록 → 병합 → 승격 판정" 순서를 `internal/cli/update/merge` 또는 `internal/cli/update/backup` 의 작은 시접으로 뺀다.

- 시접에는 테스트만 교체하는 **병합 직전 관측 훅**(패키지 수준 함수 변수)을 둔다. AC-USB-005 가 병합 직전의 확정본·대기본 바이트를 이 훅으로 관측한다.
- init 은 `ApplyAutonomyTierBundle` 호출을 테스트가 교체할 수 있는 패키지 수준 함수 변수 뒤에 둔다(N-06). AC-USB-007 `init` 셀이 이것으로 번들의 재기록을 결정적으로 모사한다.
- AC-USB-006·AC-USB-016 은 시접으로 사이클을 돌린다. 세 흐름의 배선과 판정 위치는 AC-USB-007 이 행동으로 확인한다.

## §F Milestones (우선순위 순)

### M1 — RED 기준선과 복원 동작 확인 (Priority High · 슬롯 불필요)

- merge 패키지에 AC-USB-001, 003, 012, 013 테스트를 먼저 쓴다. 오늘의 `MergeUserFiles` 에 확정본을 심어 넘기므로 컴파일되며, 값 단정이나 충돌 줄 단정에서 실패한다.
- **복원 동작 확인(D5 검증 항목).** `internal/cli/update/backup/restore.go` 를 읽거나 backup 패키지 테스트로 `RestoreMoaiConfig` 가 `.claude/settings.json` 을 쓰는지 관측한다.
- 네 요소(명령, 출력 원문, 종료 코드, 트리 SHA)와 걸린 테스트 수를 `progress.md` §E.2 에 기록한다.

### M2 — 대기본·확정본 데이터 모델 (Priority High · 슬롯 불필요)

- backup 패키지에 다음을 추가한다. sections 함수는 건드리지 않는다.
  - 두 경로 상수
  - 대기본 기록
  - 확정본 유효성 판정 읽기
  - 승격(실패 시 `settings-snapshot-promote-failed:` 경고)
  - 남은 대기본 판정·폐기
- AC-USB-004, AC-USB-009, AC-USB-015 를 통과시킨다.

### M3 — 병합 base 선택, 흐름 시접, 승격 규칙 (Priority High · 슬롯 불필요)

- `MergeUserFiles` 가 `.claude/settings.json` 에 한해 유효한 확정본을 base 로 쓰게 하고, 보존 경로 여부를 흐름 시접에 알린다(D2, D5 N-10).
- 흐름 시접, 병합 직전 관측 훅, D5 두 시점 판정을 구현한다. 승격을 빈 동작으로 둔 시접에 대해 AC-USB-006·016 의 RED 를 먼저 관측한다.
- AC-USB-001, 002, 003, 004, 006, 008 `promote_failure`, 010, 011, 012, 013, 016 을 통과시킨다.
- `base.go:29-34` 주석을 갱신하고 B5 의 @MX:NOTE 를 붙인다.

### M4 — 배선과 판정 위치 (Priority Medium · **리드 슬롯 승인 필요**)

- 진입 전에 리드에게 `internal/cli` 슬롯을 요청한다.
- D4 의 ① 판정 지점 두 곳, ② 기록 지점 세 곳, ③ 승격 지점 세 곳을 배선하고 D7 근거 2 를 연결한다.
- **테스트 대역 확장(감사 F-10).** 배포기 대역에 배포 때 지정한 렌더 바이트를 `.claude/settings.json` 에 실제로 쓰고 매니페스트를 `TemplateManaged` 로 기록하는 필드를 더한다. 두 사이클 이상 clean-reinstall 을 돌리는 테스트를 쓴다면 `makeScenarioA` 의 v2 지문이 남아 있어야 하며, 2회차 출력에 `not a v2 project — no-op` 이 없음을 함께 단정한다.
- **관측 훅 선언(감사 N-05, N-06).** D8 의 병합 직전 관측 훅과 init 자율성 번들 교체 변수를 이 마일스톤에서 들인다.
- AC-USB-005, 007, 008, 014 를 통과시킨다.
- `update_clean_install.go:394-397` 의 낡은 주석을 고친다.
- 기존 `TestCleanReinstall_SettingsJSONUserKeysPreserved`, `TestCleanReinstall_MatchesNormalPathProtection`, `TestMergeUserFiles_*`(`update_merge_test.go`), `TestUpdateSubsystem_HomeSeamReach` 가 계속 통과하는지 확인한다.

### M5 — 반증 가능성 확인 (Priority Medium)

- acceptance.md §E 의 뮤턴트 표를 하나씩 적용해 RED 를 관측하고 되돌린다. 전역 `git stash` 는 쓰지 않는다.
- 관측을 네 요소로 `progress.md` §E.2 에 남긴다.

### M6 — 범위 한정 검증 (Priority Low · M4 슬롯 안에서)

```bash
go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/
golangci-lint run ./internal/cli/update/...
GOOS=windows GOARCH=amd64 go build ./internal/cli/...
```

AC-USB-015 의 저장소 위생 명령을 커밋 뒤에 실행한다.

## §G Anti-Patterns

- 병합이 대기본을 읽게 하거나, 병합 단계가 끝나기 전에 대기본을 확정본 자리로 옮기지 않는다.
- 보존 경로를 탄 흐름에서 대기본을 승격하지 않는다(D5 경우 2).
- 정상 종료 판정을 바이트 비교로 구현하지 않는다(M-D5c, M-D5i).
- 중단된 흐름의 승격을 중단 시점에 결정하지 않는다(M-D5e).
- 남은 대기본을 백업 단계, 배포 뒤, 새 대기본 기록 뒤에 판정하지 않는다(M-D5f, M-D5g, M-D5g-w).
- 기존 sections 기록 자리에 settings.json 기록을 덧붙이지 않는다.
- 배포 뒤 디스크 파일을 다시 읽어 대기본을 만들지 않는다(M-14).
- `internal/merge` 에 settings.json 전용 분기를 넣지 않는다.
- 헬퍼만 호출하는 테스트나 끝 상태만 보는 셀로 배선 순서를 증명했다고 보고하지 않는다.
- `t.Setenv("HOME", ...)` 로 홈을 바꾸지 않는다.

## §H @MX 태그 계획

| 대상 | 태그 | 이유 |
|---|---|---|
| 세 지점이 부르는 대기본 기록 헬퍼 | `@MX:ANCHOR` + `@MX:REASON` | fan_in 3. 기록 시점이 계약이다 |
| 남은 대기본 판정 헬퍼와 두 호출 지점 | `@MX:WARN` + `@MX:REASON` | 첫 재기록 뒤로 옮기면 "중단·복원 없음 → 승격"이 폐기로 바뀐다 |
| 승격 판정 | `@MX:WARN` + `@MX:REASON` | 병합 전 승격이나 보존 흐름의 승격은 새 키를 사용자 삭제로 만든다 |
| `MergeUserFiles` 의 확정본 base 선택과 보존 경로 신호 | `@MX:NOTE` | 확정본과 유도 base 폴백의 경계, 보존 경로 세 분기 |
| `base.go` 헤더 한계 주석, `update_clean_install.go:394-397` 주석 | 주석 갱신 | 한계가 확정본이 없는 경로로 좁아짐 |
| `base_test.go` `TestMergeDropsTemplateAdditionInsideCarriedEventKey` | `@MX:NOTE` | 유도 경로 전용 특성화임을 명시 |

태그 설명은 `code_comments: en` 에 따라 영어로 쓴다.

## §I Cross-References

- `spec.md` §A.4(첫 재기록), §B.2(두 상태), §B.4(D5와 판정 위치), §C(요구사항 16개), §E(N-08)
- `acceptance.md` §B 대응표, §E 뮤턴트 표
- `.moai/reports/t656/plan-audit-iter1.md`, `plan-audit-iter2.md`
- `.moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001/spec.md` REQ-UMC-010, `plan.md` M2.1 진입 게이트
- `.claude/rules/local/gitflow-lane-protocol.md` §8 — 범위 판정의 기준 ref
