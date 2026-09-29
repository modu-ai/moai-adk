# SPEC-WEB-SAVE-LOSSLESS-001 Progress

> 카드 t1314 · GitHub issue #1731 · branch `WT-web-save-lossless`

## Phase Log

- 2026-09-29 plan-phase: manager-spec이 4종 산출(spec/plan/acceptance/progress) 초안 작성. 코드 근거 12곳 워크트리 직접 확인 (develop `2b1233b13`). 설계 결정 A+B+백스톱 확정 (plan.md §A). 다음: plan-audit.
- 2026-09-29 plan-audit iter-1: **FAIL 0.85** (차단 3건) → v0.1.1 정정 (커밋 `8feb4bcdc` 대상). F1 — D2 수리 공허성(UserConfig `name` 전용, `saveSection` 무병합 재마샬) → user.yaml `name:` 행 seam 스플라이스로 재설계. F2 — quality_extras 강제 폐기 결정 확정(§A.4 Q1). F3 — AC-WSL-002 술어 변이 분리. F4 — M1 seam 허용 확장(`sectionwrite.go:56-64`) 명시. Q2/Q3 판정 승인 기록. 정정 근거 신규 실측: `manager.go:462-470`, `schema_sections_test.go:285-290`, `sectionwrite.go:56-64`, `config.go:32-37`. 다음: delta re-audit (iter-2/2).
- 2026-09-29 plan-audit iter-2: **CONCERNS 0.90 (PASS-with-debt)** — F1-F4 해소 확인, 처분 요구 F5 + 선택 F6 → v0.1.2 반영. F5 — 잔여 재마샬 3경로를 M1 공동 범위로 편입(REQ-WSL-002/003 명시적 포괄 + AC-WSL-009 신설). **인용 경로 정정**: 감사자 인용 `internal/cli/projectconfig.go`·`internal/cli/nested.go`는 본 트리에 부재(ls 실측) — 행 번호가 일치하는 실제 경로 `internal/web/projectconfig.go:231-238/240-247`, `internal/settings/nested.go:112-156`(웹/TUI 공유 seam, TUI 호출점 `internal/cli/profile_setup.go:226`)로 검증 후 반영. F6 — AC-WSL-005 name-부재 변이에 C3 허용 1행. 다음: run 진행 (Implementation Kickoff Approval 경유).
- 2026-09-29 run-phase: manager-develop (cycle_type=tdd). RED-first 신규 무손실 테스트 4패키지 18건 작성, 전건 RED 확인(E8 — `.moai/state/verify/t1314/` 원문 보존) 후 GREEN. 주요 실측: (1) F5의 "TUI 호출점 `profile_setup.go:226`"은 실제로는 web `writeProjectConfig`의 TUI 쌍둥이 `persistProjectConfig`로, nested 공유 seam 호출점이 아님 — 본래 residual 재마샬 3경로에 **네 번째 동류 경로**로 확인되어 동일 공유 원시(`settings.WriteProjectScalars`)로 함께 전환됨(범위 판정: plan §F M1의 "devMode/convention 스칼라 쓰기 seam" 잔여 경로 부류에 귀속, C4 공유 seam 수혜). (2) 사용자 이름 스플라이스는 M3의 user 더티 게이트가 없으면 잔여 `Save()`가 재마샬로 덮어쓰므로(splice 뒤 덮어쓰기 실측), M3 백스톱을 M1 GREEN의 필수 전제로 조기 구현. (3) seam이 sections 디렉터리 부재 프로젝트(EC-5)에서 실패 — `WriteSectionViaSeam`에 MkdirAll 보강. 이슈 재현 시나리오(#1731)는 web `TestHandleSave*` 전체 POST /save 경로 테스트로 기계 고정(수동 브라우저 재현 대체). 다음: M5 마감.
- 2026-09-29 sync-audit 수리 (manager-develop 재위임): sync-audit FAIL 차단 F-1 — `profile/sync.go`의 language 쓰기가 SetSection("language")+Save() 전체-재마샬로, 콘솔 언어 셀렉트 4종이 도달하는 **다섯 번째 잔여 재마샬 경로**였음(감사자 프로브 재현 → `TestSyncToProjectConfig_LanguageEditSplicesRows` RED: 14행→9행, 주석 5건+미모델링 키 소실+미존재 `error_messages` 기본값 주입). user.yaml name-스플라이스 선례와 동일한 yamlpatch 행-스플라이스로 전환(GREEN). **계약 정정 실측**: `git_commit_messages`/`code_comments` pref가 config 기본값(`en`, defaults.go:14-15)과 같고 디스크에 행이 없으면 기준값이 이미 일치하므로 행을 생성하지 않는다 — 구 재마샬의 기본값 물질화는 부작용이었고 기존 `TestSyncToProjectConfig_Languages`를 데이터-레벨 계약(resolved 값)으로 갱신. 수반: F-2 `projectscalars.go` @MX:WARN+@MX:REASON 신설(§E.4 mx_tag_validation 주장이 이제 참이 됨 — §E.4 본문 불요 변경), F-3 `projectconfig.go` 스테일 @MX:REASON을 스플라이스 경로 서술로 개정, F-5-note `write_safety_test.go` git-strategy 픽스처에 주석+미모델링 키 추가, `sectionwrite.go` typedSeamFiles 언어 배제 주석의 거짓 전제("콘솔 폼이 언어 파일을 만지지 않는다") 정정.

## §E.1 Plan-phase Audit-Ready Signal

- artifacts: spec.md 0.1.0 / plan.md / acceptance.md / progress.md (본 파일)
- SPEC ID 사전-검증: `SPEC-WEB-SAVE-LOSSLESS-001` regex PASS (Bash 실행, verbatim 출력 인용 완료)
- ID 유일성: `.moai/specs/` 983개 중 SPEC-WEB-SAVE-LOSSLESS-001 부재 확인 (ls 실측)
- 관련 SPEC 존재 확인: SPEC-WEB-WRITE-SAFETY-001, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-SEAM-GREENFIELD-001/002, SPEC-WEB-CONSOLE-011 (ls 실측)
- Tier: M — 영향 파일 추정 8-12 (settings 4 + config 1 + profile 1 + web 2 + 테스트) — >15 파일 아님
- 향후 진입: plan-audit → Implementation Kickoff Approval → `/moai run SPEC-WEB-SAVE-LOSSLESS-001`

## §E.2 Run-phase Evidence

Run-phase: manager-develop (cycle_type=tdd), worktree `WT-web-save-lossless`, 행위 커밋 `9be71a4f1`.

### 마일스톤

- **M1 typed→seam 라우팅**: `applyTypedEdits`를 yamlpatch 라인-스플라이스 경로로 재작성(기존 스칼라=1행 재작성, 부재 키만 upsert 폴백 — C3). typed applier는 검증 SSOT로 유지(merge_method enum/bool). `sectionwrite.go`에 `typedSeamFiles`로 user/quality/git-strategy/git-convention/llm 5파일 개방(루트 키 가드 유지 — quality 루트 키는 `constitution`; sections 디렉터리 부재 EC-5 허용 MkdirAll 보강). `nested.go` 중첩 쓰기 seam을 스플라이스로 전환(F5 제3경로). 신규 `projectscalars.go` `WriteProjectScalars` — devMode/convention 편집의 공유 seam으로 web `writeProjectConfig`와 TUI `persistProjectConfig`(F5 3경로 외 실측으로 발견한 네 번째 동류 호출점, 동일 공유 원시로 자연 커버)가 모두 위임. `profile/sync.go` 사용자 이름 편집을 `name:` 행 스플라이스로 전환(D2/F1 재설계, AC-WSL-005).
- **M2 차이-게이트**: seam no-op 게이트에 부재 키+`""` 제출 skip 분기(REQ-WSL-004). 존재 키 `""`는 삭제 시맨틱 유지(`TestCrossSessionEmptySubmitsRoundTrip` 회귀 통과).
- **M3 Save 백스톱 게이트**: `ConfigManager.Save`의 user/language/quality/llm을 dirty-or-absent 게이트로(git-strategy/git-convention 선례 확장), `resetSectionDirtyLocked`으로 EC-3 리셋 계약 6섹션 통일. greenfield(absent) 분기 유지(`TestSaveGreenfieldSectionsStillCreated`).
- **M4 테스트 전환**: `TestApplySchemaEditsForcesQualityExtrasTrue` → `TestApplySchemaEditsQualityExtrasForceRetired`(§A.4 Q1 폐기 착지). `TestWriteSectionViaSeamRejectsNonSeamSections` 거부 목록을 language로 수축(typed 5파일은 신규 `TestWriteSectionViaSeamAcceptsTypedFiles`가 수락 담당). 파손 파일 pre-flight 가드 신설(AC-WSL-006a).
- **M5 재검증**: 아래 증거 블록.

### RED-first 증거 (E8 — GREEN 이전, tree `c80d062bc`)

RED 원문 출력 보관: `.moai/state/verify/t1314/t1314_red_{settings,profile,web,config}.txt`(본 run 실측). 요지:

- `TestApplySchemaEditsOneFieldEditSplicesOneLine`: line count changed 22→23 (typed Save 전체 재마샬)
- `TestApplySchemaEditsGitStrategyEditPreservesUnmodeledKeysAndComments`: 85→103행
- `TestApplySchemaEditsQualityEditPreservesIssueNamedKey`: 82→72행 (미모델링 키 소실 + quality_extras 주입)
- `TestNestedSeamEditPreservesIssueNamedKey`: 82→72행
- `TestApplySchemaEditsAbsentKeyEmptySubmissionIsNoOp`: 빈 제출이 workflow.yaml 재기록
- `TestSyncToProjectConfig_NameEditSplicesOneRow` / `_NameAbsentUpsert`: FAIL
- `TestHandleSave{DevModeEdit,ConventionEdit,UserNameEdit,LLMEdit}` 4건: FAIL
- `TestSaveRewritesOnlyDirtySections` / `_WithoutDirtySectionsTouchesNothing`: FAIL

### 검증 (this run, this tree, HEAD `9be71a4f1`)

| 검사 | 명령 | 결과 |
|---|---|---|
| 빌드 | `go build ./...` | exit 0 |
| 크로스 빌드 | `GOOS=windows go build ./internal/{settings,config,profile,web,cli}/` | exit 0 |
| settings | `go test ./internal/settings/...` | ok 0.5s, coverage 87.6% |
| config | `go test ./internal/config/...` | ok 3.3s, coverage 82.5% |
| profile | `go test ./internal/profile/...` | ok 0.5s, coverage 85.0% |
| web | `go test ./internal/web/...` | ok 29.5s, coverage 74.5% |
| cli | `go test -timeout 30m ./internal/cli/` (slot lease go-test-heavy) | ok 1463.0s(24.4분 — 고부하 머신, 기존 실측 t1288과 동일 조건; 기본 10m 타임아웃 1회 도달 후 30m 재측정) |
| lint | `golangci-lint run` (v2.1.6 — CI 판 동일 버전) 4패키지 | 0 issues |
| spec lint | `moai spec lint SPEC-WEB-SAVE-LOSSLESS-001` | ✓ No findings |
| C1 grep | web 패키지 프로덕션 `yaml.Marshal`/`os.WriteFile` | 신규 0건(기존 주석만) |

### AC 매트릭스 (AC-WSL-001..009)

| AC | 판정 | 근거(본 run 실측) |
|---|---|---|
| AC-WSL-001 | PASS | `TestSaveRewritesOnlyDirtySections`+`TestSaveWithoutDirtySectionsTouchesNothing`(무더티 이중 Save 읽기전용 디렉터리에서도 성공=무기록)+`TestHandleSaveLLMEditIsolatesOtherSections`(미편집 3파일 byte-identical) |
| AC-WSL-002 | PASS | `TestApplySchemaEditsOneFieldEditSplicesOneLine`(llm 정확 1행), `TestApplySchemaEditsGitStrategyEditPreservesUnmodeledKeysAndComments`(1행+주석+미모델링 키), web llm 편집 1행 단언 |
| AC-WSL-003 | PASS | 상동 git-strategy 케이스 + `TestLLMTypedSavePreservesLegacyGhostKeys`(기존) 회귀 |
| AC-WSL-004 | PASS | `TestApplySchemaEditsAbsentKeyEmptySubmissionIsNoOp`(부재+"" 무기록) + `TestApplySchemaEditsExistingKeyEmptySubmissionStillWrites`(존재 키 삭제 시맨틱) + `_AbsentKeyRealValueStillUpserts`(변이 b) |
| AC-WSL-005 | PASS | `TestSyncToProjectConfig_NameEditSplicesOneRow`(name 1행만) + `_NameUnchangedNoWrite`(mtime 불변) + `_NameAbsentUpsert`(F6 upsert, 데이터 수준) + web `TestHandleSaveUserNameEditSplicesOneRow` |
| AC-WSL-006 | PASS | `TestPatchFileBrokenFilePreflightFailsNoWrite`(파손 파일 pre-flight 0쓰기); per-file 원자 쓰기는 `atomicWrite` temp+rename 기존 계약; EC-2 검증-선결·logSaveFailure는 기존 web 테스트 회귀로 유지 |
| AC-WSL-007 | PASS | M4 전환 목록(아래) — 손실 행위 긍정 단언 0건 |
| AC-WSL-008 | PASS | settings/web write_safety 전체 + `TestCrossSessionEmptySubmitsRoundTrip` + `TestPatchFileValueInvariantPreservesBytes` + git-strategy Save 게이트 테스트 전부 통과(4패키지 스위트 green) |
| AC-WSL-009 | PASS | settings `TestApplySchemaEditsQualityEditPreservesIssueNamedKey`+`TestNestedSeamEditPreservesIssueNamedKey`; web `TestHandleSaveDevModeEditPreservesIssueNamedKey`+`TestHandleSaveConventionEditLeavesQualityByteIdentical` — `session_effort_default` 키+주석 행 원문 생존, 정확 1행 변경 |

### M4 전환 목록 (REQ-WSL-010 / AC-WSL-007)

| 테스트 | 기존 단언 | 처분 |
|---|---|---|
| `TestApplySchemaEditsForcesQualityExtrasTrue` (settings) | quality 편집 시 quality_extras_enabled 강제 true 기록 | **전환** → `TestApplySchemaEditsQualityExtrasForceRetired`(강제 분기 부재 — 미제출 키 미기록) |
| `TestWriteSectionViaSeamRejectsNonSeamSections` (settings) | typed 5파일의 seam 쓰기 거부 | **수축** → 거부 목록에서 5파일 제거(language만 잔류), `TestWriteSectionViaSeamAcceptsTypedFiles` 수락 케이스 신설 |
| `TestGitStrategyWorktreeBaseBranchRoundTripPreservesManualKeys` (settings) | typed 경로의 미모델링 키 보존 | **유지** — 동일 내부 함수명의 seam 재구현에서 단언이 그대로 참(보존이 더 강해짐) |
| 그 외 write_safety(settings 12건 / web 7건) | 값-불변 무기록 | **유지** — 신규 계약과 동방향, 전건 회귀 통과 |

### 크로스 플랫폼 / lint

- `GOOS=windows` 빌드 통과(4패키지+cli).
- golangci-lint v2.1.6(CI 판) 4패키지 0 issues; `gofmt` 0 차이; `go vet` 0 출력.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: 9be71a4f1
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "n/a (카드 워크트리 — 레인 push 없음, 리더 일괄)"
l44_post_push_fetch: "n/a (push는 리더 일괄)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows: pass
total_run_phase_files: 15
m1_to_mn_commit_strategy: "행위 단일 통합 커밋 9be71a4f1 (M1+M2+M3+M4 전환) + 문서 커밋"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: "6470b61df"
sync_status: complete
b12_self_test_a: "PASS — grep -c 'SPEC-WEB-SAVE-LOSSLESS-001' CHANGELOG.md = 1 (기존 run 문서 커밋 687bac776에 이미 착지, 중복 emission 없음)"
b12_self_test_b: "PASS — acceptance.md 고유 AC 수 9건 (AC-WSL-001..009, grep -oE | sort -u | wc -l 실측) = run §E.2 AC 매트릭스 9/9와 일치"
b12_self_test_c: "PASS — CHANGELOG 항목이 인용하는 8개 구현 경로 전부 ls 실측 존재 (internal/settings/{sectionapply,sectionwrite,nested,projectscalars}.go, internal/config/manager.go, internal/profile/sync.go, internal/web/projectconfig.go, internal/cli/profile_setup.go)"
mx_tag_validation: "PASS — 본 SPEC이 건드린 8개 Go 파일 전수 스캔(grep -rn '@MX:') 실측: 신규 exported 함수 3건(WriteProjectScalars, WriteSectionViaSeam, WriteProjectNestedConfig) 모두 @MX:WARN+@MX:REASON 보유(seam 재마샬 금지 계약 경고), ConfigManager.Save @MX:ANCHOR+@MX:DEBT/CEILING/UPGRADE 기존 유지 — 추가 태그 불요"
canary_compliance_check: "n/a — 본 SPEC은 장래 정책 선언 SPEC 아님(무손실 Save 동작 구현)"
spec_status_transition: "in-progress -> completed (본 sync 커밋에 편승; body 무변경, frontmatter status만)"
docs_rotation_check: "PASS — .moai/project/{tech,structure}.md의 settings/web/config/profile 언급 구문 확인: ADR-011 'Config YAML generated via yaml.Marshal'은 init 생성 맥락(생성은 여전히 typed 직렬화)으로 여전히 참; structure.md manager.go 로드 경로 서술 무변동 — project docs 수정 불요"
changelog_entry_position: "기존 항목 CHANGELOG.md [Unreleased] (커밋 687bac776, 문서 커밋) — sync 신규 emission 없음"
```
