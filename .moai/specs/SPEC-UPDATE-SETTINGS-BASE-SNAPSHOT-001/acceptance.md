# Acceptance — SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001

## §A 검증 원칙

- 모든 수용 기준은 통과와 실패가 둘로만 갈린다. 판정 신호는 테스트 종료 코드, 테스트가 단정하는 값, 읽기 전용 명령의 출력이다.
- 테스트는 `t.TempDir()` 안에서만 쓰며, 홈이 필요하면 `userHomeDirFn` 시접으로 주입한다. `t.Setenv("HOME", ...)`, 실제 홈 접근, 실제 `moai update`·`moai init` 실행은 모두 금지다.
- 선택자로 테스트를 돌릴 때는 걸린 테스트 수를 함께 센다. 출력에 `[no tests to run]` 이 있으면 통과가 아니라 **측정 불가**다.
- 패키지 약칭: **merge** = `internal/cli/update/merge`, **backup** = `internal/cli/update/backup`, **cli** = `internal/cli`(리드 슬롯 승인 뒤에만 실행).
- 경로 약칭: **확정본** = `.moai/cache/template-snapshot/claude/settings.json`, **대기본** = `.moai/cache/template-snapshot/claude/settings.json.pending`.
- **관측 시점.** 흐름 안의 중간 상태는 plan.md Decision D8 의 병합 직전 관측 훅으로만 단정한다. 그 밖의 단정은 흐름과 흐름 **사이**(테스트가 흐름을 따로 호출하므로 관측 가능) 또는 흐름이 끝난 뒤의 상태만 본다.

## §B 요구사항 ↔ 수용 기준

| REQ | AC |
|---|---|
| REQ-USB-001 (배포가 쓰면 대기본 기록) | AC-USB-005, AC-USB-007, AC-USB-014 |
| REQ-USB-002 (경로, git 무시) | AC-USB-005, AC-USB-009, AC-USB-015 |
| REQ-USB-003 (렌더 뒤 수정 배제) | AC-USB-005, AC-USB-007 |
| REQ-USB-004 (유효 확정본을 base 로) | AC-USB-001 |
| REQ-USB-005 (승격 시점·조건·남은 대기본 판정 위치) | AC-USB-005, AC-USB-006, AC-USB-007, AC-USB-016 |
| REQ-USB-006 (안 바꾼 leaf 에 템플릿 값) | AC-USB-001, AC-USB-006, AC-USB-007 |
| REQ-USB-007 (사용자 편집 유지) | AC-USB-002 |
| REQ-USB-008 (둘 다 바꿈 → 사용자 값 + 충돌) | AC-USB-003, AC-USB-013 |
| REQ-USB-009 (부재·손상 폴백, 비차단) | AC-USB-004 |
| REQ-USB-010 (기록·승격 실패 비차단, 구별되는 경고) | AC-USB-008 |
| REQ-USB-011 (새 키 전달 유지) | AC-USB-006, AC-USB-007, AC-USB-010, AC-USB-016 |
| REQ-USB-012 (사용자 키 삭제 존중) | AC-USB-012, AC-USB-016 |
| REQ-USB-013 (sections 스냅숏 불변) | AC-USB-009 |
| REQ-USB-014 (공용 엔진 불변) | AC-USB-015 |
| REQ-USB-015 (settings.json 에만 적용) | AC-USB-011 |
| REQ-USB-016 (건너뛴 배포는 기록 금지) | AC-USB-014 |

## §C RED 기준선 표기

"RED-now" 는 이 SPEC 을 구현하기 전 트리에서 해당 테스트가 **값 단정에서** 실패해야 한다는 뜻이다. 컴파일 오류는 RED 로 치지 않는다. 아직 없는 시접이 필요한 기준은 **RED-stub** 으로 표기한다. 시접의 틀만 만들고 동작을 비워 둔 상태에서 값 단정으로 실패하는 것을 뜻한다.

[HARD] plan 단계에서는 어떤 RED 도 관측하지 않았다(이번 위임은 `go test` 실행 금지). 아래 RED 칸은 모두 **미관측 예측**이다. run 단계가 명령, 출력 원문, 종료 코드, 트리 SHA 네 요소를 `progress.md` §E.2 에 기록하기 전까지는 어떤 기준도 릴리스 차단 자격을 갖지 않는다(`.claude/rules/moai/development/verification-completeness.md` §2.1).

| 표기 | 뜻 |
|---|---|
| **RED-now** | 구현 전 트리에서 실패해야 한다. 이유를 칸에 적는다 |
| **RED-stub** | 시접의 빈 구현에 대해 실패해야 한다 |
| **guard** | 구현 전에도 통과한다. 깨뜨리면 안 되는 동작을 지키며, §E 뮤턴트로 반증 가능성을 확보한다 |

## §D 수용 기준 (Given-When-Then)

### AC-USB-001 — 사용자가 안 바꾼 공유 leaf 의 템플릿 값 변경이 전달된다 · **RED-now**

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseDeliversTemplateValueChange`
- **Given** `t.TempDir()` 프로젝트에 `.moai/manifest.json` 이 있다. 확정본에 이전 렌더 `{"statusLine":{"command":"old"},"env":{"PATH":"/old"},"permissions":{"deny":["A"]}}` 가 있고, 사용자 백업은 그 렌더와 같다. 새 렌더 `{"statusLine":{"command":"new"},"env":{"PATH":"/new"},"permissions":{"deny":["A","B"]}}` 가 `.claude/settings.json` 에 배포돼 있다.
- **When** `MergeUserFiles(root, []FileBackup{{Path: ".claude/settings.json", Data: user}}, &out)` 를 부른다.
- **Then** 결과 파일에서 `statusLine.command == "new"`, `env.PATH == "/new"`, `permissions.deny == ["A","B"]` 이다.
- **대조 셀** — 확정본이 없는 같은 입력에서는 세 값이 `old`, `/old`, `["A"]` 로 남는다(오늘도 통과).
- **RED 이유** — 지금은 확정본을 읽지 않고 base 를 유도하므로 세 leaf 가 모두 "사용자만 바꿈"으로 읽힌다.
- **green 경로** — M3.

### AC-USB-002 — 사용자가 편집한 키는 유지된다 · guard

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseKeepsUserEdit`
- **Given** 확정본 `{"model":"sonnet"}`, 사용자 백업 `{"model":"opus"}`, 새 렌더 `{"model":"sonnet"}`.
- **Then** 결과는 `model == "opus"` 이고, 출력에 `merged with conflicts` 가 없다.

### AC-USB-003 — 둘 다 바꾸면 사용자 값을 유지하고 충돌을 보고한다 · **RED-now**

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseBothChangedReportsConflict`
- **Given** 확정본 `{"model":"sonnet"}`, 사용자 백업 `{"model":"opus"}`, 새 렌더 `{"model":"haiku"}`.
- **Then** 결과는 `model == "opus"` 이고, `out` 에 `.claude/settings.json merged with conflicts (user version preferred)` 가 정확히 한 번 나온다.
- **RED 이유** — 유도 base 에서는 "사용자만 바꿈"으로 끝나 충돌 줄이 없다.
- **green 경로** — M3.

### AC-USB-004 — 확정본이 없거나 손상되면 이 SPEC 이전과 같은 결과로 폴백하고 막히지 않는다 · guard

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotFallbackMatchesDerivedBase` (하위 테스트 `absent`, `unreadable_dir`, `invalid_json`, `json_array`, `json_null`)
- **Given** AC-USB-001 의 사용자 백업과 새 렌더가 있다. 확정본 자리는 하위 테스트마다 다르다. 파일이 없거나, 같은 이름의 디렉터리이거나, `{"a":` 이거나, `[1]` 이거나, `null` 이다.
- **Then** 반환값은 `nil` 이고, 결과 파일 바이트는 확정본이 전혀 없는 프로젝트의 결과와 `bytes.Equal` 이다.
- **반증** — M-04.

### AC-USB-005 — 대기본은 병합 전 렌더 바이트이고, 병합이 끝나기 전에는 확정본을 바꾸지 않는다 · **RED-now**

- 패키지·테스트: cli(슬롯) · `TestCleanReinstall_SettingsSnapshotStagedBeforeMerge`
- **Given** `makeScenarioA` 기반 프로젝트가 있다. 사용자 `.claude/settings.json` 에는 사용자만 가진 키 `"userOnly": true` 와 폐기 v2 deny 항목 1개가 있다. 확정본 자리에는 표식 바이트 `{"marker":"prior"}` 를 미리 심는다. 테스트 배포기 대역(plan.md M4 확장)은 배포 때 렌더 `R`(`userOnly` 없음, 폐기 deny 없음)을 `.claude/settings.json` 에 쓴다. **관측은 plan.md Decision D8 의 병합 직전 관측 훅(M4 에서 도입)으로 한다.** 홈은 `homeSeamSpy` 로 주입한다.
- **When** `runCleanReinstall` 을 실행한다.
- **Then** 네 가지를 단정한다.
  - (i) 관측 훅에서 본 확정본 바이트가 `{"marker":"prior"}` 다.
  - (ii) 관측 훅에서 본 대기본 바이트가 `R` 이다.
  - (iii) 흐름이 끝난 뒤 확정본 바이트가 `R` 이다(D5 경우 1).
  - (iv) 결과 `.claude/settings.json` 은 `userOnly` 를 가지므로 `R` 과 다르다.
- **RED 이유** — 지금은 대기본도 확정본 갱신도 없다.
- **반증** — M-05, M-06c.
- **green 경로** — M4.

### AC-USB-006 — 두 사이클에서 한 흐름의 base 는 이전 흐름의 렌더다 (두 흐름 순서 모두) · **RED-stub**

- 패키지·테스트: merge 또는 backup(시접이 사는 곳) · `TestSettingsSnapshotFlow_TwoCycles_BaseIsPreviousRender` (하위 테스트 `single_call_order`, `split_deploy_then_restore_order`)
- **Given** 1회차 렌더 `R1 = {"model":"sonnet","userOnly":false}` 와 사용자 파일 `{"model":"sonnet","userOnly":true}` 로 1회차가 병합 결과를 쓴다(D5 경우 1). 2회차 렌더는 `R2 = {"model":"haiku","userOnly":false,"newKey":1}` 이고, 사용자는 1회차 결과 파일을 편집하지 않는다.
  - `single_call_order` 는 clean-reinstall 처럼 배포·병합·승격을 한 번에 호출한다.
  - `split_deploy_then_restore_order` 는 일반 update 처럼 배포 단계 호출과 병합·승격 단계 호출을 따로 한다.
- **When** 흐름 시접으로 두 사이클을 돌린다.
- **Then** 2회차 뒤 `.claude/settings.json` 은 `model == "haiku"`, `newKey == 1`, `userOnly == true` 이고, 2회차 뒤 확정본은 `R2` 다.
- **RED 이유** — 승격이 빈 구현이면 2회차가 유도 base 로 병합해 `model` 이 `sonnet` 으로 남는다.
- **반증** — 각 하위 테스트에서 M-06c 와 M-06t 는 `newKey` 단정이 RED 여야 한다.
- **green 경로** — M3.

### AC-USB-007 — 배선: 기록·승격·남은 대기본 판정이 실제 흐름의 올바른 위치에서 실행된다 · **RED-now**

- 패키지·테스트: cli(슬롯) · `TestSettingsSnapshot_WriteSites`

**공통 렌더.** `R1 = {"a":1}`, `R2 = {"a":2,"K":1}`, `R3 = {"a":3,"K":1,"L":1}`.

| 하위 테스트 | Given | When | Then |
|---|---|---|---|
| `clean_reinstall` | AC-USB-005 하네스, 확정본 `R1`, 사용자 파일 `R1`, 배포 렌더 `R2` | `runCleanReinstall` | 결과 `a == 2`, `K == 1`, 흐름 뒤 확정본 `R2` |
| `template_sync_backup_empty` | 백업 경로가 비는 조건(`configBackupPath == ""`), 확정본 `R1`, 사용자 파일 `R1`, 배포 렌더 `R2` | 템플릿 동기화 흐름 | 결과 `a == 2`, `K == 1`, 흐름 뒤 확정본 `R2` |
| `template_sync_backup_filled` | 위와 같되 백업 경로가 차는 조건 | 템플릿 동기화 흐름 | 위와 같다 |
| `update_leftover_abort` | 확정본 `R1`, 이전 흐름이 남긴 대기본 `R2`, 살아 있는 파일 `R2`(중단 모사), 배포 렌더 `R3` | `runUpdate` 경로(dry-run 아님, 버전 다름) | 결과 `a == 3`, `K == 1`, `L == 1`, 흐름 뒤 확정본 `R3` |
| `update_leftover_version_skip` | 확정본 `R1`, 남은 대기본 `R2`, 살아 있는 파일 `R2`, 프로젝트 버전 = 패키지 버전 | `runUpdate` 경로(버전 일치로 동기화 건너뜀) | 반환 뒤 확정본 `R2`, 대기본 없음 |
| `init` | 새 디렉터리(settings.json 없음), 배포 렌더 `R2`. 자율성 번들 호출을 plan.md D8 의 교체 변수로 바꿔 살아 있는 파일의 `a` 를 `9` 로 고치는 함수로 둔다(티어·게이트 환경과 무관) | init 흐름 | 흐름 뒤 확정본 바이트가 `R2` 이고, 살아 있는 파일은 `a == 9` 라 `R2` 와 다르다 |

- **행동 검사를 세울 수 없는 흐름의 대체 기준.** 소스 위치 검사로 대신하되 조건을 조인다.
  - 기록 호출은 배포 호출의 성공 경로와 같은 블록 안의 무조건 문장이어야 한다.
  - 남은 대기본 판정 호출은 `runUpdate` 의 조기 반환 뒤, `stripRetiredV2DenyEntries(` 앞의 무조건 문장이어야 하고, `runInit` 에서는 `executor.Execute(` 앞이어야 한다.
  - 대체를 택한 사실과 이유를 `progress.md` §E.2 에 적는다. `update_leftover_*` 셀은 대체를 허용하지 않는다(판정 위치가 핵심이라서다).
- **RED 이유** — 기록·승격·판정이 어느 흐름에도 없다.
- **반증** — M-07a~e, M-06t-w(`template_sync_*` 두 셀), M-D5g-w(`update_leftover_*` 두 셀).
- **green 경로** — M4.

### AC-USB-008 — 대기본 기록·승격 실패는 흐름을 막지 않고 구별되는 경고를 남긴다 · **RED-now / RED-stub**

- 패키지·테스트: cli(슬롯) · `TestSettingsSnapshot_WriteFailureDoesNotBlock` (하위 테스트 `clean_reinstall`, `helper`). 흐름 시접 패키지 · `TestSettingsSnapshotFlow_PromoteFailureDoesNotBlock`.

| 셀 | Given | When | Then | 표기 |
|---|---|---|---|---|
| `clean_reinstall` | AC-USB-005 하네스, `.moai/cache/template-snapshot/claude` 자리에 일반 파일(대기본 기록 실패, sections 는 영향 없음) | `runCleanReinstall` | 반환 `nil`. `settings-snapshot-write-failed:` 로 시작하는 줄 정확히 1개. `Warning: template snapshot write failed:` 줄 0개. settings.json 은 폴백 병합 결과 | RED-now |
| `helper` | 위와 같은 조건 | 세 흐름이 부르는 기록 헬퍼 직접 호출 | 위와 같은 문구 단정 | RED-now |
| `promote_failure` | 흐름이 대기본 `R` 을 기록했고, 확정본 자리 `claude/settings.json` 이 비어 있지 않은 디렉터리여서 승격(교체)이 실패 | 흐름 시접으로 흐름을 끝냄(경우 1) | 반환 `nil`. `settings-snapshot-promote-failed:` 로 시작하는 줄 정확히 1개. `settings-snapshot-write-failed:` 줄 0개. 확정본 자리 디렉터리는 그대로 남음 | RED-stub |

- **REQ-USB-010 의 init·template_sync 적용 범위.** 세 흐름은 같은 기록 헬퍼와 같은 승격 원시 동작을 부르며, 그 호출 사실은 AC-USB-007 이 행동으로 확인한다. 이 둘의 조합을 init·template_sync 의 비차단 근거로 삼는다.
- **반증** — M-08, M-08s, M-08p.
- **green 경로** — M3(`promote_failure`), M4(나머지).

### AC-USB-009 — sections 스냅숏의 파일과 동작이 바뀌지 않는다 · guard

- 패키지·테스트: backup · `TestSectionsSnapshot_UnaffectedBySettingsSubpath`, 그리고 기존 테스트 7개
  - `TestSnapshot_SurvivesCleanStep`, `TestWriteSnapshot_FailureDoesNotBlock`, `TestSnapshot_ScopeLimitedToSections`
  - `TestMerge_BaseFromSnapshot_NotFromEmbedded`, `TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal`, `TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal`, `TestMerge_QualityYaml_Real3Way_WithSnapshot`
- **Given**
  - 셀 A: 같은 sections 입력을 가진 두 프로젝트가 있고, 하나에만 확정본과 대기본이 있다.
  - 셀 B: sections 스냅숏은 **없고** 확정본만 있는 프로젝트다.
- **When**
  - 셀 A: 두 프로젝트에서 `WriteSnapshot`, `HasSnapshot`, `SaveTemplateBase(destDir, root)` 를 실행한다.
  - 셀 B: `WriteSnapshot` 을 부르지 않고 `HasSnapshot` 과 `SaveTemplateBase(destDir, root)` 를 실행한다.
- **Then**
  - 셀 A: `sections/` 트리의 파일 목록과 바이트, `HasSnapshot` 반환값이 두 프로젝트에서 같다.
  - 셀 B: `HasSnapshot == false` 이고, `destDir/sections/` 의 바이트가 `SaveTemplateDefaults(otherDir)` 결과와 같다.
  - 기존 테스트 7개가 모두 `--- PASS` 다(7개 모두 걸렸는지 센다).
- **반증** — M-09 는 셀 B 에서 RED 여야 한다.

### AC-USB-010 — 새 키 전달이 확정본 유무와 관계없이 유지된다 · guard

- 패키지·테스트: merge · 기존 테스트 6개와 신규 `TestMergeUserFiles_SnapshotBaseAddsNewTemplateKey`
  - 기존: `TestMergeAddsTemplateEntryUserNeverHad`, `TestMergeAddsNestedTemplateKey`, `TestMergeUserFilesAddsTemplateEntry`, `TestMergeKeepsUserEditToSharedKey`, `TestMergeKeepsUserDeletionInCarriedEventKey`, `TestMergeDropsTemplateAdditionInsideCarriedEventKey`
- **Given** 신규 테스트 입력은 확정본 `{"a":1}`, 사용자 `{"a":1}`, 새 렌더 `{"a":1,"hooks":{"New":[1]}}` 다.
- **Then** 신규 테스트 결과에 `hooks.New` 가 있고, 기존 테스트 6개가 모두 `--- PASS` 다.

### AC-USB-011 — 확정본 base 는 `.claude/settings.json` 에만 적용된다 · guard

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseScopedToSettingsJSON`
- **Given** 확정본이 있다. 같은 캐시 루트 아래에 가짜 스냅숏 `claude/mcp.json` 과 `.mcp.json` 이 있으며, 두 파일 모두 AC-USB-001 의 **이전 렌더** 바이트를 담는다. `.mcp.json` 사용자 백업은 이전 렌더와 같고, 배포된 `.mcp.json` 은 AC-USB-001 의 새 렌더다.
- **When** `.mcp.json` 만 넘겨 `MergeUserFiles` 를 부른다.
- **Then** `.mcp.json` 결과 바이트는 스냅숏이 전혀 없을 때의 결과와 같다(옛 값이 남는다).
- **반증** — M-11.

### AC-USB-012 — 확정본이 있으면 사용자가 지운 템플릿 키는 되살아나지 않는다 · **RED-now**

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseHonorsUserKeyDeletion`
- **Given** 확정본 `{"statusLine":{"command":"x"},"model":"sonnet"}`, 사용자 백업 `{"model":"sonnet"}`, 새 렌더 `{"statusLine":{"command":"x"},"model":"sonnet"}`.
- **Then** 결과에 `statusLine` 이 없다.
- **대조 셀** — 확정본이 없으면 결과에 `statusLine` 이 있다(오늘도 통과).
- **RED 이유** — 유도 base 에서는 키가 되살아난다.
- **green 경로** — M3.

### AC-USB-013 — 배열 통째 비교 한계를 특성화한다 · **RED-now**

- 패키지·테스트: merge · `TestMergeUserFiles_SnapshotBaseArrayIsWholeLeaf`
- **Given** 확정본 `{"permissions":{"allow":["Read"]}}`, 사용자 `{"permissions":{"allow":["Read","Bash(ls)"]}}`, 새 렌더 `{"permissions":{"allow":["Read","Grep"]}}`.
- **Then** 결과 `permissions.allow == ["Read","Bash(ls)"]` 이고 `Grep` 은 없다. 출력에 `merged with conflicts` 가 한 번 나온다.
- **RED 이유** — 충돌 줄 단정이 유도 base 에서 실패한다.
- **green 경로** — M3.

### AC-USB-014 — init 배포가 기존 settings.json 을 건너뛰면 대기본을 기록하지 않는다 · **RED-now**

- 패키지·테스트: cli(슬롯) · `TestInitSettingsSnapshot_SkippedDeployRecordsNothing` (하위 테스트 `fresh_dir`, `untracked_existing`, `user_modified_existing`)
- **Given** init 배포 대역은 실제 비강제 배포기 규칙(`internal/template/deployer.go:239-255`)을 따르며 렌더 `R` 을 쓰려 한다.
  - `fresh_dir`: settings.json 이 없다.
  - `untracked_existing`: 매니페스트 기록이 없는 사용자 파일 `U` 가 있다.
  - `user_modified_existing`: 매니페스트에 `user_modified` 로 기록된 사용자 파일 `U` 가 있다.
- **When** init 흐름을 실행한다.
- **Then**
  - `fresh_dir`: 흐름이 끝난 뒤 확정본 바이트가 `R` 이다(D5 경우 6).
  - `untracked_existing`, `user_modified_existing`: 대기본과 확정본이 둘 다 **없고**, `.claude/settings.json` 은 `U` 그대로다(D5 경우 7).
- **RED 이유** — 오늘은 기록이 없어 `fresh_dir` 셀이 실패한다. 나머지 두 셀은 오늘도 통과하며, 뮤턴트로 반증한다.
- **반증** — M-14.
- **green 경로** — M4.

### AC-USB-015 — 저장소 위생: git 무시, 공용 엔진·템플릿 원본 무수정, 실제 홈 무접촉 · guard

아래 명령은 **커밋 뒤** 카드 워크트리에서 하나씩 실행한다. `<CARD_BASE>` 는 `git merge-base develop HEAD` 의 출력 SHA 를 그대로 붙인 값이다.

| 명령 | 기대 |
|---|---|
| `git check-ignore -v .moai/cache/template-snapshot/claude/settings.json` | 종료 코드 0, 규칙 `.moai/cache/` |
| `git check-ignore -v .moai/cache/template-snapshot/claude/settings.json.pending` | 종료 코드 0 |
| `git check-ignore -v .claude/settings.json` | 종료 코드 1 (대조군 — 추적 파일) |
| `git diff --name-only <CARD_BASE>..HEAD` | 1줄 이상 (대조군 — 0이면 측정 불가) |
| `git diff --name-only <CARD_BASE>..HEAD -- internal/merge/ internal/template/templates/` | 빈 출력 |
| `git diff <CARD_BASE>..HEAD -G 'Setenv\("HOME"' --name-only -- '*_test.go'` | 빈 출력 |

`-G` 는 커밋 범위에서 해당 정규식이 추가되거나 삭제된 줄이 있는 파일만 보고한다. 작업 트리의 미커밋 변경은 보지 않으므로 커밋 뒤에 실행한다.

### AC-USB-016 — 승격 규칙: 운영자 D5 의 다섯 결과와 판정 위치

- 패키지·테스트: 흐름 시접이 사는 곳(merge 또는 backup) · `TestSettingsSnapshotFlow_PromotionRule` (하위 테스트 `c1` … `c8`, 번호는 plan.md Decision D5 경우 번호)
- **공통 Given.**
  - `R1 = {"a":1}` 이 확정본으로 있다.
  - 이번 흐름의 렌더는 `R2 = {"a":2,"K":1}` 이다.
  - 다음 흐름의 렌더는 `R3 = {"a":3,"K":1,"L":1}` 이며, `R2` 와 다르다.
- **관측 시점.** "이번 흐름 뒤"와 "다음 흐름 전"은 흐름 사이의 상태다. "다음 흐름 뒤"는 다음 흐름이 끝난 상태다. 흐름 안의 중간 상태는 단정하지 않는다.

| 셀 | 운영자 결과 | 이번 흐름 조건 | Then — 이번 흐름 뒤 | Then — 다음 흐름 뒤 | 표기 |
|---|---|---|---|---|---|
| c1 | 병합이 결과를 씀 → 승격 | 사용자 `{"a":1,"u":1}`, 병합이 결과를 씀(살아 있는 파일 ≠ `R2` 바이트) | 확정본 `R2`, 대기본 없음 | — | RED-stub |
| c2 | 병합 실패·통째 보존 → 이전 확정본 유지 | 사용자 파일 JSON 손상 `{"a":`, 병합이 보존 경로를 탐. 다음 흐름 전 사용자가 파일을 `{"a":1}` 로 고침 | 확정본 `R1`, 대기본 없음 | `a == 3`, `K == 1`, `L == 1` | 이번 흐름 뒤: RED-stub / 다음 흐름 뒤: guard |
| c3 | (병합 건너뜀) → 승격 | 사용자 파일이 `R2` 와 바이트 동일 | 확정본 `R2` | — | RED-stub |
| c4 | 배포 뒤 병합 전 중단, 복원 없음 → 승격 | 배포 뒤 병합 전에 중단(살아 있는 파일 = `R2`, 대기본 `R2` 남음) | 확정본 `R1`, 대기본 `R2` | `a == 3`, `K == 1`, `L == 1`, 확정본 `R3` | RED-stub |
| c5 | 중단 뒤 사용자 파일 복귀 → 이전 확정본 유지 | c4 처럼 중단한 뒤, 다음 흐름 전에 사용자 파일 `{"a":1}` 을 되돌려 씀(손 복원. 복원 명령은 settings.json 을 되돌리지 않는다는 판독이며 M1 이 관측) | 확정본 `R1`, 대기본 `R2` | `a == 3`, `K == 1`, `L == 1`, 확정본 `R3` | 다음 흐름 뒤: guard, 뮤턴트로 반증 |
| c6 | init 이 파일을 씀 → 승격 | init 이 `R2` 를 씀, 이후 자율성 번들 모사가 `a` 를 `9` 로 고침 | 확정본 `R2` | — | RED-stub |
| c8 | (사용자 파일 없음) → 승격 | update, 사용자 settings.json 이 원래 없음 | 확정본 `R2` | — | RED-stub |

- D5 경우 7(init 이 기존 파일을 건너뜀)은 AC-USB-014 가 담당한다. 판정 위치의 실제 흐름 배선은 AC-USB-007 `update_leftover_*` 가 담당한다.
- **다음 흐름 뒤 값의 계산.**
  - c4: 올바른 규칙은 base `R2`, current `R2`, updated `R3` 이다. `a` 는 템플릿만 바꿈이라 3, `L` 은 추가된다. base 가 `R1` 로 남는 뮤턴트에서는 `a` 가 1→2→3 으로 양쪽이 바꾼 셈이 되어 충돌이 나고 2 로 남는다.
  - c2·c5: 올바른 규칙은 base `R1`, current `{"a":1}`, updated `R3` 이다. `a` 는 3, `K`·`L` 은 추가된다. base 가 `R2` 가 되는 뮤턴트에서는 `a` 가 충돌로 1 에 남고, `K` 는 base 에 있고 current 에 없어 사용자 삭제로 빠진다.
- **반증**

| 뮤턴트 | RED 가 나야 하는 셀 | 까닭 |
|---|---|---|
| M-D5a — 항상 승격 (리드 요구) | c2 다음 흐름, c5 다음 흐름 | base `R2` 가 되어 새 키 `K` 가 사용자 삭제로 읽혀 빠지고, `a` 가 1 로 남는다 |
| M-D5b — 병합이 결과를 쓸 때만 승격 | c4 다음 흐름, c6, c8 | 병합이 없거나 중단된 흐름에서 승격하지 못해 base 가 `R1` 로 남는다(c4 `a == 2`) |
| M-D5c — 정상 종료를 살아 있는 파일 = 렌더 바이트로 판정 | c1, c6 | 병합 결과와 번들이 고친 파일은 렌더와 바이트가 달라 승격되지 않는다 |
| M-D5d — 남은 대기본을 판정 없이 버림 | c4 다음 흐름 | base 가 `R1` 로 남아 `a == 2` |
| M-D5e — 중단 흐름을 중단 시점에 판정 | c5 다음 흐름 | 중단 시점의 살아 있는 파일은 `R2` 라 승격되고, 뒤이은 복귀를 보지 못한다 |
| M-D5f — 새 대기본을 기록한 뒤 남은 대기본 판정 | c4 다음 흐름 | 새 대기본 `R3` 가 병합 전에 확정본이 되어 base = updated. `a` 가 2 로 남고 `L` 이 사용자 삭제로 빠진다 |
| M-D5g — 남은 대기본을 흐름의 첫 재기록(배포) 뒤에 판정 | c4 다음 흐름 | 살아 있는 파일이 이미 `R3` 라 남은 `R2` 와 달라 버려지고, base 가 `R1` 로 남아 `a == 2` |
| M-D5i — 정상 종료 보존 판정을 "살아 있는 파일 = 흐름 이전 사용자 파일" 바이트 비교로 구현 | c3 | 사용자 파일과 렌더가 같아 보존으로 오판하고 승격하지 않는다 |

## §E 뮤턴트 표 (M5 에서 적용 → RED 관측 → 되돌림)

| ID | 뮤턴트 | RED 가 나야 하는 기준 |
|---|---|---|
| M-01 | 확정본 분기를 지워 늘 유도 base 사용 | AC-001, AC-003, AC-012, AC-013 |
| M-02 | 확정본이 있으면 새 렌더를 통째로 씀 | AC-002, AC-003 |
| M-04 | 손상 확정본이면 병합이 오류 반환 | AC-004 |
| M-05 | 대기본을 병합 뒤 파일 복사로 만듦 | AC-005 (ii) |
| M-06c | clean-reinstall 형 순서에서 병합 전에 대기본을 확정본으로 승격 | AC-005 (i), AC-006 `single_call_order` |
| M-06t | template_sync 형 순서(시접의 배포 단계 호출 안)에서 확정본을 새 렌더로 덮음 | AC-006 `split_deploy_then_restore_order` |
| M-06t-w | 실제 Deploy Templates 단계에서 확정본을 직접 씀(대기본·시접 우회) | AC-007 `template_sync_backup_empty`, `template_sync_backup_filled` (`K` 가 사용자 삭제로 빠지고 `a` 가 1 로 남음) |
| M-07a | init 의 기록을 `ApplyAutonomyTierBundle` 뒤로 옮김 | AC-007 `init` |
| M-07b | template_sync 의 기록 삭제 | AC-007 `template_sync_*` |
| M-07c | clean_reinstall 의 기록을 `stripRetiredV2DenyEntries` 뒤로 옮김 | AC-007 `clean_reinstall` |
| M-07d | template_sync 의 기록을 `if configBackupPath != ""` 블록 안으로 옮김 | AC-007 `template_sync_backup_empty` |
| M-07e | init 의 기록을 `executor.Execute` 오류 분기 안으로 옮김 | AC-007 `init` |
| M-08 | 기록 실패를 흐름 오류로 반환 | AC-008 `clean_reinstall`, `helper` |
| M-08s | settings 기록 실패 경고를 sections 경고 문구로 출력 | AC-008 문구 단정 |
| M-08p | 승격 실패를 흐름 오류로 반환하거나 기록 실패 문구로 출력 | AC-008 `promote_failure` |
| M-09 | `HasSnapshot` 이 스냅숏 루트 전체의 비어 있음 여부를 봄 | AC-009 셀 B |
| M-11 | 확정본 base 를 모든 JSON 병합 대상에 적용 | AC-011 |
| M-14 | 건너뜀과 무관하게 배포 뒤 디스크 파일을 읽어 대기본 기록 | AC-014 기존 파일 두 셀 |
| M-D5a | 항상 승격 | AC-016 c2·c5 다음 흐름 |
| M-D5b | 병합이 결과를 쓸 때만 승격 | AC-016 c4 다음 흐름·c6·c8 |
| M-D5c | 정상 종료를 바이트 동일로 판정 | AC-016 c1·c6 |
| M-D5d | 남은 대기본을 판정 없이 버림 | AC-016 c4 다음 흐름 |
| M-D5e | 중단 흐름을 중단 시점에 판정 | AC-016 c5 다음 흐름 |
| M-D5f | 새 대기본 기록 뒤 남은 대기본 판정 | AC-016 c4 다음 흐름 |
| M-D5g | 남은 대기본을 첫 재기록(배포) 뒤에 판정 | AC-016 c4 다음 흐름 |
| M-D5g-w | 실제 흐름에서 남은 대기본 판정을 백업 단계나 배포 뒤로 옮김 | AC-007 `update_leftover_abort`(`a == 2`), `update_leftover_version_skip`(판정이 실행되지 않아 확정본 `R1`) |
| M-D5i | 정상 종료 보존 판정을 바이트 비교로 구현 | AC-016 c3 |

각 관측은 명령, 출력 원문, 종료 코드, 트리 SHA(뮤턴트 적용 트리 포함)를 `progress.md` §E.2 에 남긴다.

## §F Edge Cases

- **버전 일치로 update 를 건너뜀** — 배포가 없어 새 대기본은 기록되지 않는다. 남은 대기본 판정은 건너뛰기보다 앞에서 실행된다(AC-USB-007 `update_leftover_version_skip`).
- **`--dry-run`** — 남은 대기본 판정을 포함해 아무것도 바꾸지 않는다. 판정 위치가 dry-run 조기 반환 뒤다.
- **사용자가 `.moai/cache/` 를 지움** — 다음 update 는 폴백으로 병합하고, 그 흐름의 대기본이 승격된 뒤부터 base 가 다시 생긴다.
- **여러 버전을 건너뛴 update** — 확정본은 마지막으로 반영된 렌더다. 그 뒤로 몇 버전이 바뀌었든 올바른 base 다.
- **기계를 바꾼 뒤 update** — 렌더된 `env.PATH` 가 달라져 템플릿 변경으로 전달된다. 사용자가 직접 고친 값이면 AC-USB-003 과 같은 충돌 경로를 탄다.
- **중단 뒤 끼어든 쓰기** — 남은 대기본이 버려진다(spec.md §E, N-08). 안전한 쪽으로 벗어난 결과이며, 이를 고정하는 AC 는 두지 않는다.
- **동시 update** — 단일 사용자 오프라인 모델이며 잠금을 두지 않는다.

## §G Definition of Done

- AC-USB-001~016 이 모두 통과한다.
- RED-now·RED-stub 셀(001, 003, 005, 006, 007, 008, 012, 013, 014, 016 의 표기 셀)의 구현 전 RED 관측이 네 요소로 `progress.md` §E.2 에 있다.
- §E 뮤턴트 27개의 RED 관측이 `progress.md` §E.2 에 있다.
- M1 의 복원 동작 확인 결과가 `progress.md` §E.2 에 있다.
- `go test ./internal/cli/update/merge/... ./internal/cli/update/backup/...` 가 통과하고, `internal/cli` 영향 테스트가 슬롯 안에서 통과한다.
- `golangci-lint run ./internal/cli/update/...` 가 깨끗하고, `GOOS=windows GOARCH=amd64 go build ./internal/cli/...` 가 성공한다.
- merge·backup 패키지 커버리지가 M1 기준선 이상이다.
- 전 패키지 판정은 push 뒤 `origin/develop` CI 결과로 확인한다.
