# Progress — SPEC-SEC-HARDEN-004

## F.1 Plan Structure

- **Tier**: S (minimal) — 2 files (update.go, file_changed.go) + tests, additive guards, consistent with SEC-HARDEN-003.
- **cycle_type**: tdd (reproduction-first — 2 mandatory reproduction ACs).
- **Milestones**: M1 (F1 restoreTargetContained parent-chain, 양 walk 동시) → M2 (F2 runMXScan scan-target EvalSymlinks) → M3 (final verify + push).
- **REQ count**: 8 (REQ-SEC4-001..008) + 5 NFR.
- **AC count**: 10 (AC-SEC4-001..010) — SSOT = acceptance.md.
- **Artifacts**: spec.md + plan.md + acceptance.md + progress.md (Tier S 2-file 최소 집합을 초과해 acceptance.md 분리 — reproduction AC + regression AC 명확화 위해).

## E. Audit-Ready Signals

### E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-06-14
- plan_status: audit-ready
- plan_authored_by: manager-spec
- spec_id: SPEC-SEC-HARDEN-004
- tier: S
- era: V3R6
- frontmatter_canonical_12_fields: verified (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags) + optional tier/era
- spec_id_self_check: PASS (decomposition: SPEC ✓ | SEC ✓ | HARDEN ✓ | 004 ✓ → PASS; regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)
- exclusions_section: present (§F + §F.1 Out of Scope h3)
- gears_requirements: REQ-SEC4-001..008 (Event-detected / State-driven / Ubiquitous; no legacy IF/THEN)
- threat_model_closed_classes: F1 (symlinked-intermediate-dir write CWE-22) + F2 (symlink-in-root lexical read CWE-61) declared CLOSED (sync-auditor Recommendation #2 이행)
- reproduction_acs: 2 mandatory (AC-SEC4-001 F1 write-reject, AC-SEC4-004 F2 read-reject)
- regression_acs: AC-SEC4-003 (F1 leaf 가드), AC-SEC4-006 (F2 lexical/re-root 가드)

### E.2 Run-phase Evidence

- run_complete_at: 2026-06-14
- run_authored_by: manager-develop
- run_status: PASS (10/10 AC)
- cycle_type: tdd (reproduction-first)
- baseline_sha (AC-008 diff scope): 9c3a736c4089775af85f8c4fec337781365b279c
- M1 commit: restoreTargetContained parent-chain symlink containment (F1) — status draft → in-progress
- M2 commit: runMXScan symlink-in-root scan-target containment (F2)
- total_run_phase_files: 4 (internal/cli/update.go, internal/cli/update_fileops_test.go, internal/hook/file_changed.go, internal/hook/file_changed_test.go) + spec.md frontmatter (status transition)
- m1_to_mN_commit_strategy: M1 → M2 separate commits, push after M3 final verify (Hybrid Trunk main 직진)

#### AC PASS/FAIL Matrix

| AC | Severity | Status | Verification | Actual |
|----|----------|--------|--------------|--------|
| AC-SEC4-001 | MUST-PASS | PASS | `go test -run TestRestoreMoaiConfig_RejectsSymlinkedParentDir$ ./internal/cli/ -v` | === RUN + --- PASS (modern_walk + legacy_walk), outside/evil.yaml 미생성 |
| AC-SEC4-002 | MUST-PASS | PASS | `grep -nE 'restoreTargetContained\(configDir, targetPath\)' update.go` | 2 매치 (L1993 모던 + L2085 레거시); 공유 헬퍼 1곳 수정으로 양 walk 동시 봉쇄 |
| AC-SEC4-003 | regression | PASS | `go test -run 'TestRestoreMoaiConfig_LegacyBackup\|...3WayMerge\|...SkipsNonYAML' ./internal/cli/` | ok — leaf 가드 + 복원 동작 회귀 없음 |
| AC-SEC4-004 | MUST-PASS | PASS | `go test -run TestRunMXScan_RejectsSymlinkInRootEscapingTarget$ ./internal/hook/ -v` | === RUN + --- PASS, secret MX-tag 사이드카 미기록 |
| AC-SEC4-005 | MUST-PASS | PASS | `go test -run 'TestFileChanged_AsyncReturn_Under100ms\|TestFileChanged_SideEffectsCompleted' ./internal/hook/` | ok — 빈 payload + async 계약 회귀 없음 |
| AC-SEC4-006 | regression | PASS | `go test -run 'TestRunMXScan_RejectsUncontainedFilePath\|...SidecarCWD\|...AllowsInProjectPath' ./internal/hook/` | ok — lexical/re-root 가드 회귀 없음 |
| AC-SEC4-007 | MUST-PASS | PASS | `grep -rnE 'internal/cli/specid' internal/hook/` | 0 매치 (specid import 없음) |
| AC-SEC4-008 | MUST-PASS | PASS | `git diff --name-only 9c3a736c4 -- 'internal/**/*.go' \| grep -vE '_test'` | 정확히 2 (update.go + file_changed.go), 신규 source 0 |
| AC-SEC4-009 | MUST-PASS | PASS | `GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/hook/...` | exit 0 |
| AC-SEC4-010 | regression | PASS | `go test -cover ./internal/cli/ ./internal/hook/` | cli 71.7% (== baseline), hook 81.5% (== baseline) — 회귀 없음 |

- cross_platform_build: go build ./... exit 0 + windows scoped build exit 0
- new_warnings_or_lints_introduced: 0 (golangci-lint full = 0 issues)
- c_hra_008_grep (2-file scoped per plan-audit D2): 0 매치
- full_test_suite: `go test ./...` — 0 FAIL (E6 pre-push 검증; SEC-HARDEN-002 L_push_before_full_test_regression 준수)
- F2 root normalization 정정: EvalSymlinks resolve-recheck 비교 base(root)도 EvalSymlinks 정규화 — macOS /var→/private/var false escape로 AllowsInProjectPath/SideEffectsCompleted 1차 RED 발생 → 정규화 추가로 GREEN (M1 F1 watch-item-2 동일 정규화 요구가 F2에도 적용됨)

### E.3 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-06-14
- sync_authored_by: manager-docs
- sync_status: complete
- status_transition: in-progress → implemented
- ac_count: 10 (acceptance.md SSOT)
- changelog_section: Security
- sync_commit_sha: ddce7fcf41e5a4b2941c4ebad7cc20fcb4b90de5
- files_in_commit: CHANGELOG.md, spec.md (frontmatter), progress.md (this section)

### E.4 Sync-Audit + Post-Audit Fix Record

- sync_auditor_verdict: PASS-WITH-DEBT (조화평균 ≈ 0.78; Func 75 / Sec 70 MUST-PASS 통과 / Craft 78 / Consist 92)
- sync_auditor_central_finding: F2 (CWE-61) genuinely CLOSED; F1 (CWE-22) leaf-form CLOSED이나 deep-nonexistent-under-symlinked-parent same-class escape 잔존 → §B.1 class-level CLOSED 과장 (SHOULD-FIX HIGH confidence, end-to-end 재현)
- user_decision: 즉시 수정 (이 SPEC 내; SEC-HARDEN-001 D1 선례 — SPEC self-claim 위반=즉시수정, L_sync_auditor_scope_finding_vs_false_pass)
- fix_commit: 44f1f83f5 — fix(SPEC-SEC-HARDEN-004): F1 deep-nonexistent-under-symlinked-parent 봉쇄 (deepest-existing-ancestor walk)
- fix_approach: parentChainContained os.IsNotExist 분기 → deepest-existing-ancestor walk (configDir floor 포함). deep-under-symlink reject + legit deep first-restore allow. godoc L2191-2214 정정.
- fix_verification (orchestrator 독립): TestRestoreMoaiConfig_RejectsDeepNonexistentUnderSymlinkedParent (modern+legacy) RED→GREEN + TestRestoreTargetContained_ParentChainBranches 5 branch 전부 PASS + 회귀 0 (leaf/F2/SEC-HARDEN-003 leaf/legacy NewFile floor) + full go test ./... 0 FAIL + cross-platform build exit 0
- closure_status: F1 §B.1 class CLOSED 이제 정확 (leaf + deep 양쪽 봉쇄). 잔여 deferred: TOCTOU(offline 수용) / §F.1 ${IFS} / §F.2 env-trust (별도 SPEC)
- sync_audit_report: .moai/reports/sync-audit/SPEC-SEC-HARDEN-004-2026-06-14.md (gitignored local)

### E.5 Mx-phase Audit-Ready Signal

- mx_complete_at: 2026-06-14
- mx_authored_by: orchestrator-direct
- status_transition: implemented → completed
- era: V3R6 (H-4: §E.3 sync_commit_sha + §E.5 mx_commit_sha 양쪽 present)
- four_phase_close: plan(d954f3fb0) → run(M1 6fd63365b / M2 675ce46e8 / M3 ea9b42085) → sync(ddce7fcf4) → post-audit-fix(44f1f83f5) → Mx(this commit)
- sync_commit_sha: ddce7fcf41e5a4b2941c4ebad7cc20fcb4b90de5
- mx_commit_sha: (to-backfill by orchestrator)
- final_ac: 10/10 + deep-variant reproduction (post-audit) PASS
- sync_auditor: PASS-WITH-DEBT → SHOULD-FIX 해소 후 clean

### E.1.1 Plan Audit Record (Phase 2.3)

- plan_audit_verdict: PASS
- plan_audit_score: 0.91 (Tier S threshold 0.75; iter-1)
- dimensions: Clarity 0.95 / Completeness 0.92 / Testability 0.90 / Traceability 1.00
- must_pass: MP-1 REQ 일관성 PASS / MP-2 GEARS PASS / MP-3 frontmatter PASS / MP-4 N/A auto-pass
- ground_truth_shared_helper_claim: CONFIRMED (restoreTargetContained update.go:2141, 1 def + 2 call sites L1993 modern walk + L2085 legacy walk, identical signature → BLOCKING-risk eliminated)
- open_design_questions: 4/4 채택안 지지 (both-walks shared-helper / EvalSymlinks-parent-chain / F2 resolve-recheck 비대칭 정당 / not-exist fail-closed-except 안전)
- ac_idiom_audit: ALL CLEAN (이전 -008/-009 PASS-WITH-DEBT 원인이던 vacuous test-name infix / 누락 $ anchor / sibling-prefix collision 전부 부재; 모든 grep/test-name AC live-probed)
- defects_orch_resolved: D1 MINOR (AC-SEC4-008 origin/main 이동 baseline → pre-flight C.1 캡처 SHA 고정) + D2 MINOR (E4/B3 C-HRA-008 grep을 변경 2파일로 한정, harness.go/agent_lint.go pre-existing 5 매치 baseline 제외) — orchestrator-direct patch
- run_phase_watch_items: (1) AC-008 baseline SHA 캡처·사용 (2) EvalSymlinks-normalize configDir before filepath.Rel (3) os.IsNotExist 정밀 구분 (coarse any-error→pass 금지)
- skip_eligible_phase_0_5: score 0.91 ≥ 0.90 → run-phase Phase 0.5 재감사 skip-eligible (단 GATE-2는 §19.1 별도 사용자 결정으로 score 무관 강행)

## §E.0 Phase 0.95 Mode Selection (run-phase entry)

- Decision: sub-agent (Mode 5)
- input: tier S / scope 2 files (update.go, file_changed.go) / domain 1 (Go security hardening) / lang 100% Go / concurrency benefit LOW (coding-heavy)
- evaluation: Mode 1 trivial (no — semantic guard logic) / Mode 2 background (no — Write 필요) / Mode 3 agent-team (no — team prereqs 미충족 + <3 domain) / Mode 4 parallel (no — coding-heavy per Anthropic caveat) / Mode 6 workflow (no — <30 files, not mechanical) / Mode 5 sub-agent (SELECTED — coding-heavy sequential default)
- justification: 2-file Go security 봉쇄는 coding-heavy single-domain. Anthropic coding-task parallelism caveat에 따라 sequential sub-agent (manager-develop cycle_type=tdd)가 안전한 기본값. M1→M2→M3 순차 실행.
- gate2_passed: YES (사용자 run-phase 진입 승인, score-independent §19.1)
- phase_0_5_skip: skip-eligible (plan-auditor PASS 0.91 ≥ 0.90, no plan-PR commit since verdict; spec-workflow.md skip policy) — skip 사유를 run delegation Section A에 기록
- baseline_sha (AC-008 D1): 9c3a736c4089775af85f8c4fec337781365b279c

## 다음 단계

- GATE-2: 사용자 run-phase 진입 승인 완료 (plan-auditor PASS 0.91, §19.1 score-independent human gate).
- run-phase: manager-develop (cycle_type=tdd) Mode 5 sequential — M1(F1) → M2(F2) → M3(검증+push).
- plan-auditor: Tier S PASS threshold 0.75 — 충족 (0.91).
