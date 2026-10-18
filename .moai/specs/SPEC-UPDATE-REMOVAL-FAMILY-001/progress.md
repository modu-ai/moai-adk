# SPEC-UPDATE-REMOVAL-FAMILY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- plan 산출물 작성 완료 (2026-10-17, manager-spec, card t1668): spec.md, plan.md, acceptance.md, decision-index.md, 이 파일. 상태: `draft` (spec.md frontmatter).
- SPEC ID 사전 점검: `SPEC-UPDATE-REMOVAL-FAMILY-001` 정규식 검사 PASS (이번 실행, 원문 출력 인용: `PASS`). 카탈로그 중복 없음 (워크트리 `.moai/specs/` 와 1차 체크아웃 카탈로그에서 매치 없음).
- 질문 Q1-Q4 는 decision-index.md 에 기록됐고 기본값이 적용되었다(2026-10-17T11:54:29Z).
- plan-audit 1차 (`.moai/reports/t1668/plan-audit-iter1.md`, audited_sha `76f530fd2`): 판정 FAIL, 총점 0.875(Tier M 문턱 0.80 초과), MP-7 미통과(마커 4건, D1). 수리 반영(D1-D5, 2026-10-17, manager-spec): plan.md 마커 4건 해결 치환(Q1-Q4 감사 확정 인용), M1 순서 전제 교정(D3, 재배치 지시 + REQ-UPM 잔여 위험 점검 추가), acceptance.md AC-RC30-01 조건자 교정(D2, `-x` 정확 이름 + claude 종료 상태 별도 기록)·AC-RC30-26 구조 근거 각주(D4), spec.md REQ-RC30-003 Items 정정(D5). 델타 재심사 대기(D1-D4 + CN-4).

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-17
tier: M
artifacts: spec.md, plan.md, acceptance.md (3, Tier M) + decision-index.md (decision gate on) + progress.md
plan_audit: PASS-WITH-DEBT — iteration 3/3, overall_score 0.9375 (Tier M threshold 0.80), audited_sha ea47cabd85cacfe1ee478b78efc828fd138ae15a, plan_artifact_hash e6539258997c110d775680e7b5cfbf96a554acfd9099099d05c1724f266271d4, receipts rcpt-e0a2c5415ca6461eab34a774 / rcpt-bdbac887c672d0d9e1b2a6a9 / rcpt-38e8a8e04582dc2ebb4e0ab4, verdict file .moai/reports/t1668/plan-audit-iter3.md

## §E.2 Run-phase Evidence

run 단계 실행 (manager-develop, 2026-10-18, HEAD `afca91df1` 출발, run tn1p4e).

### 마일스톤 커밋

| 마일스톤 | 커밋 | 내용 |
|---|---|---|
| M1 | `479b3c355` | 제거 전 런 백업 + 마이그레이션의 Backup 단계 뒤 재배치 (AC-RC30-03) |
| M2 | `2bab42c4a` | 빈 스킬 디렉터리 정리 (AC-RC30-11) |
| M3 | `6adaca207` | 제거 반쪽 판정 테스트 (AC-RC30-27 제거 반쪽) |
| M4 | `36ef7fc4a` | 제거 게이트 판정 테스트 (AC-RC30-02) |
| 전이 | `3da74a745` | `draft → in-progress` (M1 커밋에서 누락된 소유 전이의 후속 기록; amend 금지 준수) |
| M5/M6 | (커밋 없음 — 검증 전용 / 레인 절차) | 아래 판정 기록 |

### M1 1단계 — 단계 순서 사실 + Gap G-02 측정 (트리 `afca91df1`)

- 단계 순서 사실 (본 트리 실측, plan-audit-iter1 D3와 동일 결론): 마이그레이션 호출은 `update_template_sync.go:1247`(`runTemplateSyncWithProgress` 안)에 있었고, "Backup" 인터셉트(스테이지 표 `:443-452`, 실행 `:705`-)를 소유한 `runTemplateSyncWithReporter`는 그 뒤 `:1256`에서야 호출 — 실행 순서는 마이그레이션이 백업보다 앞이었다.
- **Gap G-02 확정 (미덮음)**: RED 실행(`TestMigrationBackupBeforeRemoval`, `afca91df1`)이 기계적으로 관측 — 제거는 일어났으나 `.moai-backups` 하위에 제거 자산 사본 0건(`found 0`). 기존 런 백업(BackupMoaiConfig = `.moai/config` 한정, 네임스페이스 백업, in-memory-backups 3파일)은 마이그레이션 제거 대상(`.claude/skills/**` 등)을 덮지 않는다. 따라서 제거 지점의 파일별 백업은 필수(잉여 아님).

### 판정 명령 원문 출력 (E1/E8)

**AC-RC30-03 (M1)** — RED (`afca91df1` + 미커밋 신규 테스트, exit 1):

```
=== RUN   TestMigrationBackupBeforeRemoval
--- FAIL: TestMigrationBackupBeforeRemoval (0.80s)
    migrate_removal_backup_test.go:102: expected exactly one run-backup copy of .claude/skills/moai-foundation-core/SKILL.md under .moai-backups, found 0: []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.737s
```

GREEN (`479b3c355`, `go test ./internal/cli/ -run TestMigrationBackupBeforeRemoval -count=1 -v`, exit 0):

```
=== RUN   TestMigrationBackupBeforeRemoval
--- PASS: TestMigrationBackupBeforeRemoval (0.73s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.837s
```

**AC-RC30-11 (M2)** — RED (`479b3c355` 시점, exit 1; 보존 대조 팔은 처음부터 녹색 — 쌍 규칙의 보존 쪽):

```
--- FAIL: TestEmptiedSkillDirRemoved (0.99s)
    --- FAIL: TestEmptiedSkillDirRemoved/emptied_dir_removed_with_its_last_file (0.56s)
    --- PASS: TestEmptiedSkillDirRemoved/dir_with_user_content_stays (0.43s)
    migrate_removal_backup_test.go:154: the emptied skill directory survived its last file's removal (stat err: <nil>)
```

GREEN (`2bab42c4a`, exit 0):

```
--- PASS: TestEmptiedSkillDirRemoved (1.29s)
    --- PASS: TestEmptiedSkillDirRemoved/emptied_dir_removed_with_its_last_file (0.61s)
    --- PASS: TestEmptiedSkillDirRemoved/dir_with_user_content_stays (0.67s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.273s
```

**AC-RC30-27 제거 반쪽 (M3)** — RED (`afca91df1` 시점, M1 구현 전에 저작되어 진짜 red 상태 관측, exit 1):

```
=== RUN   TestMigrationRemovesSharedCopyWithoutProfileReach
--- FAIL: TestMigrationRemovesSharedCopyWithoutProfileReach (0.55s)
    migrate_profile_reach_test.go:132: expected exactly one run-backup copy of .claude/skills/moai-foundation-core/SKILL.md under .moai-backups, found 0: []
```

GREEN (`6adaca207`, `go test ./internal/cli/ -run '^TestMigrationRemovesSharedCopyWithoutProfileReach$' -count=1 -v`, exit 0):

```
=== RUN   TestMigrationRemovesSharedCopyWithoutProfileReach
--- PASS: TestMigrationRemovesSharedCopyWithoutProfileReach (0.58s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.306s
```

**AC-RC30-02 (M4)** — 보존 기준: RED 를 변이 관찰로 목격 (스위치 기본값을 임시 true 로 — MUTANT-PROBE, exit 1):

```
=== RUN   TestMigrationRemovalGate
--- FAIL: TestMigrationRemovalGate (0.39s)
    migrate_removal_switch_test.go:60: default-configuration migration removed [.claude/skills/moai-foundation-core/SKILL.md]; the gate must keep every removal-preconditioned copy
    migrate_removal_switch_test.go:62: .claude/skills/moai-foundation-core/SKILL.md missing: stat ...: no such file or directory
```

변이 반영 복구 후 GREEN (`36ef7fc4a`, exit 0; 기존 무판정 보호 테스트 동반 녹색):

```
=== RUN   TestMigrationDefaultKeepsConfirmedProjectCopy
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (0.35s)
=== RUN   TestMigrationRemovalGate
--- PASS: TestMigrationRemovalGate (0.40s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.633s
```

**AC-RC30-26 (M5)** — 기존 시험 재지정(Q3 확정) 판정 명령 실행 (exit 0, 비빈 스윕: 2 테스트 + 4 서브테스트):

```
=== RUN   TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow
--- PASS: TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow (0.97s)
=== RUN   TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness
--- PASS: TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness (0.00s)
    --- PASS: TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness/no_llm_yaml_defaults_to_claude (0.00s)
    --- PASS: TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness/claude (0.00s)
    --- PASS: TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness/both (0.00s)
    --- PASS: TestIntegrityProbe_ClaudeOnlyMemberFollowsHarness/gpt_codex_only (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.730s
```

프로덕션 변화 없음 (t1578 수선 `ac141b2cb` 선행 착지 — 회귀 가드).

### REQ-UPM-030/031/032 이동 경로 재점검 (M1 4단계 잔여 위험)

옮겨진 호출 경로에서 `removedPaths` → 요약 집계 불변식 재검증 (scoped 회귀 세트, `479b3c355` 시점, exit 0):

```
ok  	github.com/modu-ai/moai-adk/internal/cli	7.778s
```

(세트: TestTallyCountsMigrationRemovals — "migration: removed:" 경로 명명 + "1 removed by the asset migration" 집계 — 및 TestRecordlessBranchTallyCountsReconcileRemovals, TestCancelledUpdateKeepsProjectManagedAssets, 보호 테스트, 리포터 직접 경로 플로우 테스트 15건. 전부 녹색.)

### 번들 G (M6) — 레인 실행 절차 기록 (임시 테스트 파일로 실행 후 삭제; 커밋 없음 — Q2)

원장 매핑 → 마이그레이션 실행 → 보존·표지 관측의 세 줄 (`TestBundleGProfileReachLaneProcedure`, exit 0):

```
BUNDLE-G line 1: ledger mapping /var/folders/.../002 -> work (ResolveLaunchProfileForProject=work)
BUNDLE-G line 2: full update flow ran (runTemplateSyncWithProgress) with the switch on
BUNDLE-G line 3: shared copy present + byte-identical; disk manifest KeptBy=profile-reach after the full flow; no removal copy
```

전체 흐름(`runTemplateSyncWithReporter` 포함)이 끝난 뒤 디스크 매니페스트에서 `KeptByProfileReach` 표지가 살아남음 — plan-audit-iter2 R1이 재현한 유실 갱신 창(KeptBy 소거)이 M1 의 매니페스트 인스턴스 공유로 닫혔음을 keep 쪽에서도 입증.

### Gap 갱신

- **G-01 (AC-RC30-01)**: 미종결 — 운영자 단말 항목 (M7). 레인은 프로필 세션을 가질 수 없어 실행 불가 (SPEC A Q8). 재개 조건: 운영자가 프로필 세션에서 acceptance.md AC-RC30-01 행의 판정 명령(`CLAUDE_CONFIG_DIR="<profile-dir>" claude -p "List every skill name you can invoke, one per line." | grep -cx 'moai-plan'`, claude 종료 상태 별도 기록 의무 포함)을 실행해 이 §E.2 에 기록. 활성화 전까지 제거 스위치는 기본 false 유지 — REQ-RC30-003 의 안전 속성은 AC-RC30-01 실행 상태와 무관하게 성립(Q4 확정).
- **G-02**: 소거됨 (위 M1 1단계 기록 — 미덮음 확정 + 제거 지점 백업 구현으로 폐쇄).
- **G-28**: 재선언하지 않음 (t1642 소유).

### 품질 측정 (M8)

- `go vet ./internal/cli/` — exit 0 (run 단계 전 과정에서 4회).
- `gofmt -l internal/cli/` — 빈 목록 (clean).
- `GOOS=windows GOARCH=amd64 go build ./...` — exit 0 (E2, syscall-free 변경 예상과 일치).
- `golangci-lint` (moai gate 슬롯 보유 실행, `479b3c355` 시점): NEW 0 (run 중 발견된 errcheck 2건 — `verifyMigrationFinalState` 의 fmt.Fprintf 반환값 미점검 — 즉시 수리 반영; 최종 잔여 3건은 `internal/cli/factory_run_adopt.go` 의 기존 baseline, 본 카드 미접촉 파일).
- `go test -cover ./internal/cli/` — **부분 커버리지 기록 + 전체 패키지 수치는 배치 CI 로 이월(리더 판정)**.
  - **부분 커버리지(카드 계열만 실행, 분모는 패키지 전체)** — 85% 패키지 게이트의 판정 근거로 읽지 않는다:
    - 명령: `go test ./internal/cli/ -run 'TestMigration|TestEmptiedSkillDirRemoved|TestIntegrityProbe|TestRunUpdate_CodexOnly|TestTallyCountsMigrationRemovals' -cover -count=1`
    - 원문 출력 (2026-10-18, 이 트리 `6e2c12f5c`, 이번 실행 관측): `ok  	github.com/modu-ai/moai-adk/internal/cli	27.591s	coverage: 9.5% of statements` (코디네이터 전달 레인 관측과 수치 일치: `13.247s coverage: 9.5% of statements`)
  - 경위: 전체 패키지 `-cover` 1차 시도(슬롯 보유, 2026-10-18)는 go 기본 타임아웃 초과로 `panic: test timed out after 10m0s` 중단(수치 미발행); 재시도 슬롯 확보는 lane-17-t1839·lane-23 점유로 exit 3 거부 연속.
  - **리더 판정 (2026-10-18, 상석 규칙 35e9 준수)**: `-timeout` 상향 금지, internal/cli 전체 패키지 실행 금지 — 전체 패키지 커버리지는 **배치 CI 로 이월되는 Gap** 으로 확정. cli 외 패키지는 기본 타임아웃 그대로 `-cover` 가능하나 본 카드는 해당 사항 없음(터치 패키지 = internal/cli 단일).
- 전체 스위트 참고: 1차 커버리지 실행에서 `TestCodexReviewGatePrimaryCheckoutSkip`·`TestCodexReviewGate_TreeScopeRequestShapeUnchanged`·`TestCodexReviewGate_StaleEnvKeepsTreeScope` 3건 FAIL — 기존·환경성(darwin `/private/var` TMPDIR 경로 해소 불일치, t1772 계열; 본 카드 커밋은 해당 파일 미접촉, scoped 재실행으로 동일 실패 재관측 — `FAIL github.com/modu-ai/moai-adk/internal/cli 1.474s`). 본 카드 변경 범위 테스트는 전부 녹색(위 판정·회귀 세트).

## §F Phase 4 Mode Selection

- 입력 매개변수: tier=M, 스코프=~6 파일(internal/cli 소스 3·테스트 파일 3 예상), 도메인 수=1(Go CLI 내부), 언어 혼합=Go+테스트, 동시성 이득=LOW(코딩 중심 작업), agent-team 전제=해당 없음.
- 모드 평가: direct(아님 — 다중 파일·설계 변화) / serial(**선택** — 코딩 중심, 마일스톤 순차 의존 M1<M2<M3) / fanout(아님 — 코딩 중심 병렬화 경고) / sweep(아님 — ≥30파일 기계 변환 아님) / agent-team(명시 요청 없음).
- Decision: serial
- 근거: Anthropic 코딩 과제 병렬성 주의 — 구현 단위가 순차 의존(M1 백업·재배치 → M2 정리 → M3 합성 단언)이라 순차 위임이 안전한 기본. kickf-off 자율 형식 통과(T7 판정 PASS-WITH-DEBT + T8a audit decider, 2026-10-18) 후 첫 run 스폰 전 기록.
- 구현 위임: manager-develop 단일 순차(Tier M → 섹션 A-E 템플릿, 카드 클래스 C → effort high, cycle_type=tdd).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-18
run_commit_sha: 344bb4aee
run_status: complete-with-recorded-gaps
ac_pass_count: 5
ac_fail_count: 0
preserve_list_post_run_count: 2
l44_pre_commit_fetch: not-run (card worktree, no remote ops per dispatch; branch isolation makes origin/main divergence moot for this tree)
l44_post_push_fetch: not-run (no push — lanes never push; leader batch-pushes)
new_warnings_or_lints_introduced: 0 (2 NEW errcheck found during run were fixed before M1 commit; residual 3 pre-existing baseline in factory_run_adopt.go)
cross_platform_build:
  darwin: exit 0
  windows_amd64: exit 0
coverage_cli_pct: 9.5% (partial, card families only — full-package owed to batch CI)
coverage_note: the 9.5% figure is PARTIAL coverage (card-family selectors only; denominator = the whole internal/cli package) and is NOT read against the 85% package gate. Full-package coverage is a Gap carried to batch CI per leader ruling 2026-10-18 (35e9: no -timeout raise, no full internal/cli run; the first slot-held full-package attempt exceeded go's default 10m timeout with no number emitted; outside-cli packages could run -cover at default timeout but none were touched by this card).
total_run_phase_files: 5 (migrate_project_assets.go, update_template_sync.go, migrate_removal_backup_test.go [new], migrate_removal_switch_test.go, migrate_profile_reach_test.go)
m1_to_mN_commit_strategy: per-milestone conventional commits (M1 479b3c355 / M2 2bab42c4a / M3 6adaca207 / M4 36ef7fc4a / transition 3da74a745 / M5+M6 verification-only records + this M8 record commit), no push
ac_matrix:
  AC-RC30-02: PASS (TestMigrationRemovalGate green; RED observed via mutant probe — switch default flipped)
  AC-RC30-03: PASS (TestMigrationBackupBeforeRemoval RED at afca91df1 → GREEN at 479b3c355)
  AC-RC30-11: PASS (TestEmptiedSkillDirRemoved RED at 479b3c355 → GREEN at 2bab42c4a)
  AC-RC30-26: PASS (re-pointed t1578 tests, non-empty sweep 2+4, exit 0)
  AC-RC30-27-removal-half: PASS (TestMigrationRemovesSharedCopyWithoutProfileReach RED at afca91df1 → GREEN at 6adaca207)
  AC-RC30-01: DEFERRED — operator-terminal Gap G-01 (resume condition in §E.2 gap block; machinery gated off so REQ-RC30-003 holds regardless)
gaps:
  - G-01 open (operator terminal, M7)
  - G-02 closed (measured not-covered; closed by M1 per-file backup)
  - full-package coverage owed to batch CI (leader ruling 2026-10-18, 35e9; partial 9.5% card-families figure recorded in §E.2)
```

audit_ready: true

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-18
sync_commit_sha: pending-backfill-sync
sync_status: complete
sync_actor: manager-docs (card t1668)
b12_self_test_a: pass — pre-emission grep `grep -c 'SPEC-UPDATE-REMOVAL-FAMILY-001' CHANGELOG.md` returned 0 (이번 실행 관측, exit 1 = 매치 없음) — 중복 엔트리 없음 확인 후 발행
b12_self_test_b: pass — AC 카운터(acceptance.md, tier M 소스) live=6 declared 기준(AC-RC30-01/-02/-03/-11/-26/-27 제거 반쪽), 모호(identifier) 0. 참고: 기계 카운트는 7 — 일곱째 식별자 AC-RC30-28 은 파일 본문 자체가 t1642 소유 cross-reference 로 기재("G-28 계승 ... 소유 카드의 갭이다")한 것으로, [REF] 표지는 body 수정이라 manager-docs 금지 범위라 미표지. CHANGELOG 엔트리는 파일이 선언한 6 기준을 이름으로 참조(5 PASS + AC-RC30-01 deferred)하며 §E.3 ac_matrix 와 일치.
b12_self_test_c: pass — 엔트리가 명명하는 5 경로 모두 `ls` 실측 존재 확인 (migrate_project_assets.go, update_template_sync.go, migrate_removal_backup_test.go, migrate_removal_switch_test.go, migrate_profile_reach_test.go — 이번 실행)
changelog_entry_position: "[Unreleased] → Added 섹션 선두 bullet (Keep a Changelog 1.1.0 형식)"
frontmatter_status_transitions:
  spec_md: in-progress → implemented → completed (merged close — 이 단일 sync 커밋이 3-phase close; 별도 Mx 커밋 없음)
  plan_md: status 필드 없음 (Artifact Statelessness) — updated: 만 refresh
  acceptance_md: status 필드 없음 (Artifact Statelessness) — updated: 만 refresh
mx_tag_validation: read-only 검증 완료 — 신규 코드는 미Export 헬퍼 단일 호출자(pruneEmptiedSkillDir, backupRemovedAsset, verifyMigrationFinalState, migrationRunContext)로 ANCHOR 팬인 조건 미충족, goroutine/복잡도 ≥15 위험 패턴 없음, 추가·수정할 태그 없음 (소스 파일 수정은 manager-docs 금지 범위)
commit_subject: "chore(SPEC-UPDATE-REMOVAL-FAMILY-001): 3-phase close (card t1668)" — full-ID exact-token 규정 준수
```

audit_ready: true

## §G Override and Refusal Record

- 2026-10-17T16:04:09Z SPEC-UPDATE-REMOVAL-FAMILY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter3.md
- 2026-10-17T16:04:10Z SPEC-UPDATE-REMOVAL-FAMILY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter3.md
- 2026-10-17T16:07:39Z SPEC-UPDATE-REMOVAL-FAMILY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter3.md
- 2026-10-17T16:10:59Z SPEC-UPDATE-REMOVAL-FAMILY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1668/.moai/reports/t1668/plan-audit-iter3.md
