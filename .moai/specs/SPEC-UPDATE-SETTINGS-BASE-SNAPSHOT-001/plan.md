# Plan — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

카드 **t656** · Tier **M** · 기준 트리 `81c1d58f9` (브랜치 `WT-update-value-merge`) · v0.3.0 — plan-audit 1회차(FAIL 0.73) 반영과 운영자 답변(D5 · F-05 · F-17, 2026-09-11) 반영본

## §A Context

### Tier 판정

Tier **M**. 운영 코드 변경은 한 서브시스템(`internal/cli/update` 와 그 호출 지점)에 머물고, 예상 변경 파일은 테스트를 빼면 7개 안팎이다.

- `update/merge/merge.go`, `update/merge/base.go`
- 대기본·확정본 헬퍼 1~2개(`update/backup`)
- 흐름 시접 1개
- `update_template_sync.go`, `update_clean_install.go`, init 기록 지점

새 저장물은 기존 캐시 루트 아래 파일 둘이다. plan-auditor 통과 기준은 0.80 이다.

### 이 트리에서 읽은 사실 (실행 없이 판독만)

| 사실 | 위치 |
|---|---|
| `MergeUserFiles(projectRoot, backups, out)` 는 `projectRoot` 를 이미 받는다. 확정본을 찾는 데 시그니처를 바꿀 필요가 없다 | `internal/cli/update/merge/merge.go:173` |
| settings.json 은 `templateManaged` 판별을 통과한다(`.tmpl` 형태 인정) | `internal/cli/update/merge/base.go:53-60`, `base_test.go:143-159` |
| 일반 update 에서는 Backup 단계가 사용자 파일을 메모리에 뜨고, 정리 단계가 파일을 지우며, Deploy Templates 가 렌더를 쓰고, Restore Settings 가 병합한다 | `update_template_sync.go:487-494`, `:299-340`, `:341-371`, `:543-547` |
| update 정리 직전 사용자 settings.json 은 실행 백업의 `in-memory-backups/` 에 복사된다 | `update_disk_backup.go:47-53`, `:126-131` |
| clean-reinstall 은 Step 4.5 백업, Step 5 배포, Step 5.5 병합, 그 뒤 폐기 deny 규칙 제거 순이다 | `update_clean_install.go:398-403`, `:459`, `:507`, `:531` |
| v3 일반 경로의 폐기 deny 규칙 제거는 백업 이전에 사용자 파일(current)을 고친다 | `update.go:379-387` |
| init 은 비강제 배포기를 쓴다. 이미 있는 파일은 매니페스트 기록이 없거나 사용자 소유면 건너뛰고, 건너뛴 경로는 `DeployResult.ProtectedSkips` 에 남는다 | `init.go:821`, `deployer.go:239-255`, `skill_mirror.go:66-77` |
| init 은 배포 뒤 `ApplyAutonomyTierBundle` 이 `tool-policy.yaml` 이 있으면 프로젝트 settings.json 을 다시 쓴다 | `init.go:867`, `:889-900`, `autonomy_bundle.go:89-93` |
| init 의 기존 sections 스냅숏 기록은 자율성 번들 **뒤**(`:1030`)다 | `init.go:1030` |
| sections 스냅숏 읽기는 `sections/` 아래만 본다 | `backup/snapshot.go:114-131`, `backup/base_loader.go:26-67` |
| 병합이 실패하면 사용자 파일을 그대로 쓴다 | `merge.go:229-236` |
| clean-reinstall 테스트 대역 `stubDeployer.Deploy` 는 파일을 쓰지 않고 `ResultDeployer` 도 구현하지 않는다 | `update_clean_install_test.go:40-52` |
| clean-reinstall 은 v2 지문이 없으면 `[clean-reinstall] not a v2 project — no-op` 을 찍고 끝난다 | `update_clean_install.go:172-178` |

### 레인 제약 — `internal/cli` 컴파일·테스트 슬롯

[HARD] `internal/cli` 패키지를 컴파일하거나 테스트하려면 **리드의 슬롯 승인**이 필요하다. run 단계는 M4 에 들어가기 전에 이 승인을 요청하며, 승인 없이 `go test ./internal/cli` 나 `go build ./internal/cli/...` 를 돌리지 않는다. 테스트는 `internal/cli/update/merge` 와 `internal/cli/update/backup` 에 두는 것을 우선한다. M1–M3 은 이 두 패키지만으로 끝나도록 짰다.

## §B Known Issues

- **B1 — 같은 흐름 안의 덮어쓰기 (종결: 대기본·확정본 분리).** 배포 직후 기록한 스냅숏이 그 흐름의 병합 base 를 대체하면, base 와 updated 가 같아져 새 키 전달이 깨진다. spec.md §B.2 가 기록물을 대기본과 확정본으로 나눠 구조적으로 막았다. 뮤턴트 M-06c(clean-reinstall 순서)와 M-06t(template_sync 순서)로 반증한다.
- **B2 — init 의 후속 재기록.** `ApplyAutonomyTierBundle` 은 배포 뒤 프로젝트 settings.json 의 deny/ask 를 다시 만들 수 있다. 대기본은 그보다 앞서 기록한다(REQ-USB-003).
- **B3 — 승격 조건 (종결: D5 운영자 결정).** 1회차 plan 은 "병합 실패 보존"과 "중단"을 같은 모양으로 적었으나 틀렸다(감사 F-07). 배포 뒤 병합 전에 update 가 중단되면 살아 있는 파일은 **순수 렌더**다. 정리 단계가 파일을 지우고 배포가 렌더를 썼으며, 사용자 사본은 디스크 백업에만 남기 때문이다. 확정 규칙은 §E Decision D5 에 있다.
- **B4 — 형제 SPEC (종결: D6 운영자 결정).** REQ-UMC-010 을 그 SPEC 의 REQ-UMC-008/009 개선에 한정하도록 개정했다. 이 SPEC 이 먼저 착지하며, t576 의 M2.1 재개 때 §A.6 전제를 다시 잰다(§E Decision D6).
- **B5 — hook-delivery 특성화 테스트.** `base_test.go:269-332` `TestMergeDropsTemplateAdditionInsideCarriedEventKey` 는 유도 base 경로의 동작을 고정한다. 이 SPEC 은 그 테스트를 뒤집지 않는다. 확정본 경로에서는 사용자 배열이 이전 렌더와 같을 때 같은 추가가 **도착한다**. M3 에서 해당 테스트 옆에 @MX:NOTE 를 붙인다.
- **B6 — 배포가 파일을 건너뛰는 init (감사 F-05, 운영자 확정).** 이미 있는 settings.json 이 사용자 파일로 남은 채 init 이 끝날 수 있다. 그 파일을 읽어 기록하면 사용자 편집이 다음 update 에서 덮인다. REQ-USB-016 과 Decision D7 이 막는다.
- **B7 — 낡은 코드 주석 (감사 F-16).** `update_clean_install.go:394-397` 의 "settings.json base is unavailable in the embedded FS … preserves the user's file wholesale" 는 이미 `base.go:53-60` 과 어긋나며, 이 SPEC 뒤에는 더 어긋난다. M4 에서 `base.go:29-34` 주석과 함께 고친다.

## §C Pre-flight (run 단계 진입 점검)

```bash
# 1. 기준 트리 — 흡수한 ref 와의 merge-base 부터 잰다 (gitflow-lane-protocol §8)
git merge-base develop HEAD
git diff --name-only <위 SHA>..HEAD          # 대조군: plan 커밋 뒤라면 1줄 이상

# 2. 인용이 아직 유효한가
grep -n "func MergeUserFiles" internal/cli/update/merge/merge.go
grep -n "func templateManaged\|func pruneToShared" internal/cli/update/merge/base.go
grep -n "recordProtectedSkip" internal/template/deployer.go
grep -n "deployWithMirrorNotice\|updatemerge.MergeUserFiles\|stripRetiredV2DenyEntries" internal/cli/update_clean_install.go
grep -n "\"Deploy Templates\"\|updatemerge.MergeUserFiles" internal/cli/update_template_sync.go
grep -n "NewDeployerWithRenderer\|executor.Execute\|ApplyAutonomyTierBundle" internal/cli/init.go

# 3. 캐시 루트가 git 무시 대상인가
grep -n "^\.moai/cache/$" .gitignore internal/template/templates/.gitignore

# 4. 기준 스위트와 커버리지 — 슬롯 불필요 패키지만
go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/
```

## §D Constraints (Hard)

1. `internal/merge` 는 수정하지 않는다(REQ-USB-014).
2. sections 스냅숏 코드(`WriteSnapshot`·`HasSnapshot`·`SaveTemplateBase`)의 본문·시그니처와 호출 지점 동작을 바꾸지 않는다. 같은 패키지에 새 함수를 추가하는 것은 허용한다(REQ-USB-013).
3. `internal/template/templates/` 는 건드리지 않는다.
4. 대기본 기록, 승격, 확정본 읽기 실패는 모두 비차단이다(REQ-USB-009, REQ-USB-010).
5. 테스트는 `t.TempDir()` 과 `userHomeDirFn` 시접만 쓴다. `t.Setenv("HOME", ...)`, 실제 홈 접근, 실제 `moai update`·`moai init` 실행은 금지다.
6. `internal/cli` 컴파일·테스트는 리드 슬롯 승인 뒤에만 한다.
7. 로컬 전체 스위트(`go test ./...`)는 돌리지 않는다.

## §E Decisions (바뀔 가능성이 큰 순서)

### Decision D5 — 대기본 승격 조건 (운영자 결정 2026-09-11: 정교화 규칙)

**규칙.** 대기본은 흐름이 끝났을 때 살아 있는 `.claude/settings.json` 이 그 흐름의 렌더를 **반영**하고 있을 때만 확정본으로 승격한다. 반영하고 있지 않으면 이전 확정본을 그대로 둔다. 규칙 문장은 REQ-USB-005 에 덧붙였다.

**결정 기록.**

- 운영자가 처음 고른 1회차 권고안 (a)의 목적은 새 템플릿 키가 사용자의 삭제로 잘못 읽히는 일을 절대 만들지 않는 것이었다. 이 목적은 그대로 지켜진다. 병합이 사용자 파일을 통째로 보존한 흐름에서는 렌더를 승격하지 않으므로, 그 렌더가 새로 들인 키가 다음 update 에서 사용자 삭제로 읽히지 않는다.
- "배포 뒤 병합 전에 중단되면 이전 스냅숏을 유지한다"는 1회차 설명은 틀린 전제(중단 시 살아 있는 파일이 사용자 파일이라는 가정)에 기대고 있었다. 실제로는 순수 렌더다(감사 F-07). 그 설명은 이 규칙으로 대체되었다(운영자 결정 2026-09-11).

**경우별 결과**

| # | 경우 | 흐름 끝의 살아 있는 파일 | 결과 | 판정 시점 |
|---|---|---|---|---|
| 1 | 사용자 파일 있음, 병합이 결과를 씀 | 렌더를 반영한 병합 결과 | 승격 | 흐름 끝 |
| 2 | 병합 실패, 사용자 파일을 통째로 보존(사용자 JSON 손상, `merge.go:229-236`) | 사용자 파일 | 이전 확정본 유지 | 흐름 끝 |
| 3 | 사용자 파일이 렌더와 바이트 동일(`merge.go:207`) | 렌더 | 승격 | 흐름 끝 |
| 4 | 배포 뒤 병합 전에 중단, 복원 없음 | 순수 렌더 | 승격 | 다음 흐름 시작 |
| 5 | 4 뒤에 복원이 사용자 파일을 되돌림 | 사용자 파일 | 이전 확정본 유지 | 다음 흐름 시작 |
| 6 | init 이 파일을 실제로 씀(이후 자율성 번들이 고칠 수 있음) | 렌더 또는 렌더+번들 | 승격 | 흐름 끝 |
| 7 | init 이 기존 파일을 건너뜀 | 사용자 파일 | 대기본 없음(REQ-USB-016) | — |
| 8 | update, 사용자 settings.json 이 원래 없음(백업 항목 없음, `update_template_sync.go:489-493`) | 렌더 | 승격 | 흐름 끝 |

**"반영"의 판정 — 두 시점.** 운영자 결정의 경우 4·5 를 함께 성립시키려면 판정 시점이 둘이어야 한다. 이 절은 운영자 결정 문장에서 도출한 적용 방법이다.

- **정상적으로 끝난 흐름.** 바이트 동일로 판정하지 않는다. 그 흐름의 배포가 렌더를 쓴 뒤 흐름 이전의 사용자 파일이 통째로 되돌려 쓰이지 않았으면(병합 보존 폴백을 타지 않았으면) 반영으로 본다. 경우 1·3·6·8 은 승격하고, 경우 2 는 유지한다. 바이트 동일로 판정하면 병합 결과(경우 1)와 자율성 번들이 고친 파일(경우 6)이 승격되지 못한다.
- **승격 판정 전에 중단된 흐름.** 중단 시점에 결정하지 않고 대기본을 남긴다. 다음 흐름이 시작해 남은 대기본을 발견하면, 그 시점의 살아 있는 파일이 대기본과 바이트 동일한지로 판정한다. 같으면 승격하고(경우 4), 다르면 대기본을 버리고 확정본을 유지한다(경우 5). 이 판정은 다음 흐름이 자기 대기본을 기록하기 **전에** 끝낸다. 중단 시점에 바로 결정하면, 중단 직후에는 살아 있는 파일이 렌더이므로 승격해 버리고, 그 뒤 복원이 사용자 파일을 되돌린 경우 5 를 볼 수 없다.

**run 단계 M1 검증 항목 — 복원 동작.** 경우 5 는 `moai update --restore <dir>`(`runUpdateRestore` → `RestoreFromBackupDir`)가 `.claude/settings.json` 을 실제로 되돌린다는 전제에 선다. 이 트리에서는 확인하지 않았다(`restore_entry.go` 토큰 grep 0건까지만 봄). M1 에서 `internal/cli/update/backup/restore_entry.go` 와 백업 디렉터리의 `in-memory-backups/` 처리를 읽거나 backup 패키지 테스트로 관측해 결과를 `progress.md` §E.2 에 기록한다.

- 되돌리지 않는다면 복원 뒤 살아 있는 파일은 렌더로 남아 경우 4 와 같아진다. 규칙은 그대로 옳다.
- 이 경우 AC-USB-016 c5 는 "사용자가 손으로 파일을 되돌린 경우"로 이름만 바꾼다.

**리드 요구 뮤턴트.** M-D5a("항상 승격")를 경우 2 에 적용하면 병합이 보존한 흐름의 렌더가 확정본이 된다. 그러면 사용자가 JSON 을 고친 뒤 다음 update 에서, 렌더가 새로 들인 키가 base 에는 있고 사용자 파일에는 없으므로 사용자 삭제로 읽혀 영영 도착하지 않는다. AC-USB-016 c2 가 이것을 RED 로 잡는다.

### Decision D6 — 형제 SPEC 과의 관계 (운영자 결정, 2026-09-11)

- **해석.** `SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` REQ-UMC-010 을 그 SPEC 의 REQ-UMC-008/009 개선에 한정한다. 그 SPEC 의 REQ-UMC-010 문장과 `§D` 제외 항목을 같은 방향으로 고쳤고, HISTORY 에 기록했다.
- **착지 순서.** 이 SPEC 이 먼저 착지한다. t576 의 M2.1 이 재개될 때 그 SPEC §A.6 전제(공유 leaf 에서 "템플릿만 바꿈" 갈래에 도달할 수 없음)를 다시 잰다. 그 SPEC 의 plan.md M2.1 진입 게이트와 spec.md §A.6 에 이 사실을 적었다.
- 그 SPEC 의 status 는 in-progress 로 유지했다.

### Decision D1 — 하위 경로: 확정본 `claude/settings.json`, 대기본 `claude/settings.json.pending`

확정본 경로의 근거는 spec.md §B.1 에 있다(대칭, 점 없는 디렉터리, 정리 반경 밖과 git 무시).

대기본을 같은 디렉터리의 `.pending` 접미사로 두는 이유는 세 가지다.

- 승격이 같은 디렉터리 안의 교체가 되어 파일 시스템 경계를 넘지 않는다.
- `*.json` 을 찾는 도구가 대기본을 확정본으로 착각하지 않는다.
- sections 스냅숏 읽기는 `sections/` 만 보므로 영향이 없다.

### Decision D3 — 같은 흐름의 base 보존: 대기본 → 승격

- 병합은 늘 확정본 경로를 읽는다.
- 같은 흐름의 대기본은 병합 단계가 끝나기 전에는 확정본 자리에 오지 않는다(REQ-USB-005).
- `MergeUserFiles` 의 시그니처와 base 주입 방식은 바뀌지 않으므로 별도의 base 주입 시접이 필요 없다(감사 F-06).

### Decision D7 — "배포가 settings.json 을 실제로 썼는가"의 판정 근거 (REQ-USB-016, 운영자 확정)

디스크의 현재 파일을 다시 읽는 방식은 금지한다. 비강제 init 배포기가 건너뛴 사용자 파일을 렌더로 착각하기 때문이다. 쓸 수 있는 근거는 둘이며, run 단계가 고른다.

1. **배포 결과의 건너뜀 기록.** `DeployResult.ProtectedSkips` 에 `.claude/settings.json` 이 있으면 쓰지 않은 것이다. init 은 `initializer.go:423-430` 이 결과를 호출 지점까지 넘기지 않으므로 전달 경로를 새로 내야 한다. 테스트 대역은 `ResultDeployer` 를 구현해야 한다.
2. **렌더 바이트 자체.** 배포가 쓴 바이트를 배포기에서 직접 받는다. 또는 배포 뒤 매니페스트 항목이 `TemplateManaged` 이고 기록된 해시가 디스크 파일 해시와 같을 때만 쓴 것으로 본다. 건너뛴 파일은 `UserCreated` 로 기록되거나 기존 사용자 소유 기록이 남는다(`deployer.go:243-253`).

어느 쪽이든 AC-USB-014 의 뮤턴트 M-14("건너뜀과 무관하게 디스크에서 기록")가 RED 여야 한다.

### Decision D2 — base 선택 규칙

- 유효한 확정본의 조건은 세 가지다. 존재하고, 읽히며, JSON 객체로 해석되어야 한다. 하나라도 어긋나면 유도 base 로 간다.
- 확정본이 유효해도 사용자 파일이나 새 렌더가 JSON 으로 해석되지 않으면 지금의 보존 폴백을 따른다(`merge.go:217-225`, `:229-236`).
- `deriveTemplateBase` 시그니처는 바꾸지 않는다.

### Decision D4 — 기록·승격 지점

| 흐름 | 대기본 기록 위치 (배포 성공 뒤, 첫 재기록 전) | 승격 판정 위치 | 금지 위치 |
|---|---|---|---|
| `moai init` | `executor.Execute` 성공 뒤(`init.go:867` 이후, 오류 분기 `:868-876` 밖), `ApplyAutonomyTierBundle`(`:889`) 앞 | init 흐름 끝(병합 없음) | 오류 분기 안, 기존 sections 기록 자리 `:1030` |
| 일반 update | Deploy Templates 단계의 배포 성공 뒤(`update_template_sync.go:364-368`) | Restore Settings 의 settings.json 병합 뒤(백업 경로 유무와 무관) | Restore Settings 의 `if configBackupPath != ""` 블록 안(`:497-532`, 백업이 없으면 실행되지 않음) |
| clean-reinstall | Step 5 배포 성공 뒤(`update_clean_install.go:459-462`), Step 5.5 병합(`:507`) 앞 | Step 5.5 병합 뒤 | `:531` 폐기 deny 규칙 제거 뒤에 대기본 기록 |

- 세 흐름 모두 대기본을 기록하기 **전에**, 이전 흐름이 남긴 대기본이 있으면 D5 의 "중단된 흐름" 판정을 먼저 끝낸다.
- 세 지점은 하나의 대기본 기록 헬퍼를 부른다(fan_in 3).
- `runUpdateRestore` 는 렌더를 하지 않으므로 기록 지점이 아니다.

### Decision D8 — 흐름 시접 (감사 F-02 대응)

일반 update 의 순서(배포 단계와 Restore Settings 단계가 분리됨)를 슬롯 없이 반증하려고, "남은 대기본 판정 → 대기본 기록 → 병합 → 승격 판정" 순서를 `internal/cli/update/merge` 또는 `internal/cli/update/backup` 의 작은 시접으로 뺀다.

- 세 흐름은 이 시접이나 그 조각을 부른다.
- AC-USB-006 과 AC-USB-016 은 이 시접으로 사이클을 돌린다.
- 세 흐름이 실제로 시접을 부르는지는 AC-USB-007 이 행동으로 확인한다.

## §F Milestones (우선순위 순)

### M1 — RED 기준선과 복원 동작 확인 (Priority High · 슬롯 불필요)

- merge 패키지에 AC-USB-001, AC-USB-003, AC-USB-012, AC-USB-013 테스트를 먼저 쓴다. 오늘 존재하는 `MergeUserFiles` 에 확정본을 `claude/settings.json` 에 심어 넘기므로 컴파일되며, 값 단정이나 충돌 줄 단정에서 실패한다.
- **복원 동작 확인(D5 검증 항목).** `moai update --restore` 경로가 `.claude/settings.json` 을 되돌리는지 읽거나 backup 패키지 테스트로 관측한다. 결과에 따라 AC-USB-016 c5 의 이름을 조정한다.
- 네 요소(명령, 출력 원문, 종료 코드, 트리 SHA)와 걸린 테스트 수를 `progress.md` §E.2 에 기록한다.

### M2 — 대기본·확정본 데이터 모델 (Priority High · 슬롯 불필요)

- backup 패키지에 다음을 추가한다. sections 함수는 건드리지 않는다.
  - 두 경로 상수
  - 대기본 기록
  - 확정본 유효성 판정 읽기
  - 승격 원시 동작
  - 남은 대기본 발견·폐기 원시 동작
- AC-USB-004, AC-USB-009(sections 불변), AC-USB-015(git 무시)를 통과시킨다.

### M3 — 병합 base 선택, 흐름 시접, 승격 규칙 (Priority High · 슬롯 불필요)

- `MergeUserFiles` 가 `.claude/settings.json` 에 한해 유효한 확정본을 base 로 쓰게 한다(D2, REQ-USB-015).
- 흐름 시접(D8)과 D5 승격 규칙(두 판정 시점)을 구현한다. 승격을 빈 동작으로 둔 시접에 대해 AC-USB-006 과 AC-USB-016 의 RED 를 먼저 관측한다.
- AC-USB-001, 002, 003, 004, 006, 010, 011, 012, 013, 016 을 통과시킨다.
- `base.go:29-34` 주석을 갱신하고 B5 의 @MX:NOTE 를 붙인다.

### M4 — 기록 지점 배선 (Priority Medium · **리드 슬롯 승인 필요**)

- 진입 전에 리드에게 `internal/cli` 슬롯을 요청한다.
- D4 의 세 지점에 남은 대기본 판정, 대기본 기록, 승격 판정 호출을 배선하고, D7 의 판정 근거를 연결한다.
- **테스트 대역 확장(감사 F-10).** clean-reinstall 테스트용 배포기 대역에 두 가지를 더한다. 하나는 배포 때 지정한 렌더 바이트를 `.claude/settings.json` 에 실제로 쓰는 필드, 다른 하나는 D7 의 근거 1을 쓸 경우 `ResultDeployer` 구현이다. AC-005 와 AC-008 은 한 사이클만 돌리므로 v2 지문 유지 조건이 필요 없다. 두 사이클 이상 clean-reinstall 을 돌리는 테스트를 쓴다면 `makeScenarioA` 의 `moai.version: v2.16.1` 과 `.agency/` 가 남아 있어야 한다. 이때 2회차 출력에 `not a v2 project — no-op` 이 없음을 함께 단정한다.
- AC-USB-005, AC-USB-007, AC-USB-008, AC-USB-014 를 통과시킨다.
- **주석 갱신(감사 F-16).** `update_clean_install.go:394-397` 의 낡은 주석을 고친다.
- 기존 `TestCleanReinstall_SettingsJSONUserKeysPreserved`, `TestCleanReinstall_MatchesNormalPathProtection`, `TestMergeUserFiles_*`(`internal/cli/update_merge_test.go`), `TestUpdateSubsystem_HomeSeamReach` 가 계속 통과하는지 확인한다.

### M5 — 반증 가능성 확인 (Priority Medium)

- acceptance.md §E 의 뮤턴트 표를 하나씩 적용해 RED 를 관측하고 되돌린다. 적용과 되돌림은 이 워크트리 안의 임시 커밋이나 되돌림 커밋으로 하며, 전역 `git stash` 는 쓰지 않는다.
- 관측은 네 요소로 `progress.md` §E.2 에 남긴다.

### M6 — 범위 한정 검증 (Priority Low · M4 슬롯 안에서)

```bash
go test ./internal/cli/update/merge/... ./internal/cli/update/backup/... -count=1
go test -cover ./internal/cli/update/merge/ ./internal/cli/update/backup/
golangci-lint run ./internal/cli/update/...
GOOS=windows GOARCH=amd64 go build ./internal/cli/...
```

AC-USB-015 의 저장소 위생 명령을 커밋 뒤에 실행한다. 슬롯 안에서도 `internal/cli` 는 영향받는 테스트만 선택자로 돌린다.

## §G Anti-Patterns

- 병합이 대기본을 읽게 하거나, 병합 단계가 끝나기 전에 대기본을 확정본 자리로 옮기지 않는다(B1).
- 병합이 사용자 파일을 통째로 보존한 흐름에서 대기본을 승격하지 않는다(D5 경우 2).
- 중단된 흐름의 승격을 중단 시점에 결정하지 않는다. 다음 흐름 시작에서 판정한다(D5 경우 5).
- 정상 종료 흐름의 반영 여부를 바이트 동일로 판정하지 않는다(D5 경우 1·6).
- 기존 sections 기록 자리에 settings.json 기록을 덧붙이지 않는다.
- 배포 뒤 디스크 파일을 다시 읽어 대기본을 만들지 않는다(B6).
- `internal/merge` 에 settings.json 전용 분기를 넣지 않는다.
- 헬퍼만 호출하는 테스트나, 조건문 안 배치를 통과시키는 소스 위치 검사로 배선을 증명했다고 보고하지 않는다.
- `t.Setenv("HOME", ...)` 로 홈을 바꾸지 않는다.

## §H @MX 태그 계획

| 대상 | 태그 | 이유 |
|---|---|---|
| 세 지점이 부르는 대기본 기록 헬퍼 | `@MX:ANCHOR` + `@MX:REASON` | fan_in 3. 기록 시점(배포가 실제로 쓴 뒤, 첫 재기록 전)이 계약이다 |
| 승격 판정 | `@MX:WARN` + `@MX:REASON` | 병합 전 승격이나 보존 흐름의 승격은 새 키를 사용자 삭제로 만든다 |
| 남은 대기본 판정 | `@MX:NOTE` | 중단된 흐름을 다음 흐름 시작에서 판정하는 이유(복원 뒤 상태를 봐야 함) |
| `MergeUserFiles` 의 확정본 base 선택 | `@MX:NOTE` | 확정본과 유도 base 폴백의 경계 |
| `base.go` 헤더 한계 주석, `update_clean_install.go:394-397` 주석 | 주석 갱신 | 한계가 확정본이 없는 경로로 좁아짐 |
| `base_test.go` `TestMergeDropsTemplateAdditionInsideCarriedEventKey` | `@MX:NOTE` | 유도 경로 전용 특성화임을 명시 |

태그 설명은 `code_comments: en` 에 따라 영어로 쓴다.

## §I Cross-References

- `spec.md` §B.1(A1 보완), §B.2(두 상태), §B.3(D6), §B.4(D5), §B.5, §C(요구사항 16개)
- `acceptance.md` §B 대응표, §E 뮤턴트 표
- `.moai/reports/t656/plan-audit-iter1.md` — 1회차 감사 판정
- `.moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001/spec.md` REQ-UMC-010, `plan.md` M2.1 진입 게이트
- `.claude/rules/local/gitflow-lane-protocol.md` §8 — 범위 판정의 기준 ref
