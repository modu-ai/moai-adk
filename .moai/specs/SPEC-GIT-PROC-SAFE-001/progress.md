# SPEC-GIT-PROC-SAFE-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-GIT-PROC-SAFE-001
phase: plan
status: draft
tier: M
harness: standard
baseline_tree: "develop c9ceff175"
audit_baseline: "main 2213871af"
findings:
  - {id: AC-01, severity: P1, state: live, scope: "manager-git.md x2 + spec-workflow.md x2"}
  - {id: AC-11, severity: P2, state: remediated, action: verification-only}
  - {id: SX-R04, severity: P1, state: partial, scope: "delivery.md Step 3.4 x2"}
artifacts: [spec.md, plan.md, acceptance.md, research.md, progress.md]
clarifications_pending: 0
plan_complete_at: "2026-09-13T21:02:43Z"
plan_status: audit-ready
```

플랜 아티팩트 세트 완성(2026-09-14, manager-spec). 발신 카드 t782가 설계 방향을 사전 결정 — 미해결 명확화 표식 없음. plan-audit 1차 판정(FAIL 8.1/10) F1-F7 수리 적용 완료(동일 일자). 델타 재감사 2차 판정: **PASS 9.4/10** (plan-auditor, `.moai/reports/t782/plan-audit.md` iteration-2 섹션; `moai spec lint` RED 1 error → GREEN 0 error 직접 관측 포함). run-phase 유의: manager-git.md 는 agents-emit 대상 트리라 run 종료 시 `make agents-emit` 필요성 보고 필수(plan.md §D).

## §F Phase 4 Mode Selection

- 입력: tier M · scope 6 files · domain 3 그룹(agents doctrine / workflow rules / sync skill) · 언어 믹스 markdown 100% · concurrency benefit LOW · agent-team 미요청
- 평가: direct — 기각(다중 파일 교리 재작성, 위임 필요) / **serial — 선택** / fanout — 기각(M1→M2 교차참조 의존, 단일 writer 규율) / sweep — 기각(의미 재작성, 기계 변환 아님)
- Decision: serial — manager-develop 단일 순차 스폰(M1→M2→M3→M4)
- 근거: 문서 절차 재작성은 coding-heavy·의존형 작업으로 Anthropic coding-task 직렬 기본 원칙이 적용되는 영역. 병렬 이점 없음(마일스톤 간 교차참조 정합 의존). Kickoff 은 리드-운영자 채널에서 운영자 완수 지시(2026-09-14)로 갈음됨.

## §E.2 Run-phase Evidence

run-phase 완료(2026-09-14, manager-develop, serial M1→M2→M3→M4, 본 워크트리 WT-gitproc-audit). 문서 전용 RED/GREEN 검증 — 모든 증거는 이번 실행에서 본 트리(HEAD `3686f85fc` 시점)에 대해 직접 실행한 명령의 실측 출력이다.

### E8 RED 선(先)증거 (편집 전 캡처, baseline `e06eae05a` = develop `c9ceff175` + plan 커밋 — 6개 대상 파일 기준으로는 c9ceff175와 동일 트리)

첫 편집 전 Read 도구 직접 판독으로 acceptance.md RED-now 셀과 일치하는 위반 텍스트를 관측했다:

- manager-git.md 로컬 98행 `git checkout main && git pull origin main` / 112행 `git switch -c feat/SPEC-XXX` / 121-124행 Phase D `git checkout main`·`git fetch origin`·`git reset --hard origin/main`·`git pull origin main` / 129행 복구 `git fetch origin && git reset --hard origin/main` / 139행 `main_late_branch` 설명의 `reset --hard origin/main` — 템플릿 사본 96·110·119-122·127·137행 동일 내용 (사본 간 2행 오프셋, 본문 동일 — research.md 관측과 일치).
- spec-workflow.md 양 사본 50행 Route B 사전조건("`git rev-parse --abbrev-ref HEAD == main`... plan-phase commits land directly on `main` and are pushed only after Phase C `git switch -c plan/SPEC-XXX`") + 55-60행 closure 명령 블록(`git checkout main`...`git pull origin main`) + 62행 post-condition — 사본 간 동일 행·동일 내용.
- delivery.md 양 사본 Step 3.4: "Mode conditions (same as `manager-git.md` § PR Auto-Merge)" + 2조건 재인용, "When auto-merge is triggered" 4단계, "Auto-Merge Execution" 5단계(4단계 "Checkout target branch, fetch latest" 포함), "Auto-Merge Failures" 3불릿.

### AC 판정 (E1)

| AC | 판정 | 검증 명령 | 실측 출력(요지 — 전문은 커밋 M1-M4 당시 배치 출력) |
|----|------|-----------|------|
| AC-GP-01a | PASS | `grep -n 'reset --hard\|checkout main\|switch -c'` 양 사본 | 잔존 2적중 전량 비위반: 42/44행은 범위 밖 선존재 롤백 문구(`git reset --hard [checkpoint-tag]`, AC-MIRROR-01 보존), 129/127행은 신규 금지 문장("never a recovery path"). 위반 문맥 0건 |
| AC-GP-01b | PASS | 동일 grep, spec-workflow.md 양 사본 | `grep_exit=1` — 0 적중 |
| AC-GP-01c | PASS | Read 양 사본 옵션 행 | `main_late_branch` 설명이 워크트리-브랜치 모델 서술, `reset --hard`/`git switch -c` 문구 없음 |
| AC-SX-01 | PASS | `grep -n 'Mode conditions (same as\|Auto-Merge Execution\|Auto-Merge Failures\|Checkout target branch'` 양 사본 | `dup_exit=1` — 0 적중. "Only applies when a PR was created in Step 3.2" 프레이밍 + Post-Merge Automatic Cleanup(`workflow.worktree.auto_cleanup`) 절 양 사본 존재 확인 |
| AC-AC11-01 | PASS | `grep -c 'run \`git fetch\` first and wait until it completes, then run \`git rev-list --count --left-right\`'` 양 사본 | 각 `1`; `git diff e06eae05a..HEAD -- (manager-git.md 양 사본) | grep -c 'reads the remote-tracking refs'` = `0` — AC-11 문장 편집 무접촉. 편집 커밋 0건 |
| AC-XREF-01 | PASS | grep `Late-Branch Invocation Pattern` | 양 사본 `### Late-Branch Invocation Pattern` 앵커 존재(88/90행) + 옵션 행 참조. 역참조: spec-workflow 50행 "Late-branch precondition (the Late-Branch closure contract)" + Step 4 "**Late-branch closure (the Late-Branch closure contract):**" — manager-git.md 교차참조 문구와 해석 가능. 끊긴 앵커 0건 |
| AC-MIRROR-01 | PASS | `git diff e06eae05a..HEAD --stat -- (6개 파일)` | manager-git.md 38행, spec-workflow.md 10행, delivery.md 26행 ×양 사본 — 전부 발견 건 범위 행. frontmatter description(무접촉), lint/크로스빌드 블록(무접촉), Route A/B 산문(무접촉), AC-11 행(diff 0) 불변 |
| AC-TN-01 | PASS | `grep -c 'SPEC-GIT-PROC-SAFE-001'` + `grep -c 'rules/local'` 템플릿 사본 3개 | 각 `0` |

### REQ-GP-002 자기규율 grep (§E.1 항목 1)

manager-git.md 양 사본 `grep -n 'reset --hard\|checkout main\|switch -c'`: 42/44행 선존재 롤백 문구(범위 밖) + 129/127행 금지 문장 — primary-checkout 수순 지시 0건. spec-workflow.md: 0건. delivery.md Step 3.4 신규 텍스트: 금지 명령 0건.

### 2차 스윕 — 범위 밖 적중 2건 (blocker 보고, E7 참조)

1. `.claude/skills/moai/workflows/plan/spec-assembly.md:376`(템플릿 사본 동일): Late-branch를 "main commit + late switch... manual `git switch -c feat/SPEC-*` at PR time"으로 서술 — 6개 대상 파일 밖. 운영자 범위 확장 판정 대상(후속 카드 후보).
2. delivery.md 양 사본(로컬 349행/템플릿 324행, Step 3.4 이전 구간): "`git checkout develop && git pull origin develop`... `git checkout main && git pull origin main`" — 6개 파일 안이지만 research.md가 고정한 발견 건 범위(B10이 보존하는 Route A/B 산문 구간) 밖. 통합 워크트리 진입 후 문맥이므로 위반 여부 판정은 범위 확장과 함께 운영자 몫. 무편집 유지.

### E5 SPEC lint (M4 후)

`moai spec lint SPEC-GIT-PROC-SAFE-001` → `0 error(s), 2 warning(s)` — baseline(0 error / 2 warning: OwnershipTransitionUnmeasured INFO + MovingRefUnpinned 2건)과 동일. 신규 오류 0건.

### E6 커밋 상태

| 커밋 | 내용 |
|------|------|
| `75cc17a29` | M1 — manager-git.md 양 사본 워크트리 모델 전환 + spec.md `draft → in-progress` 전이 (Authored-By-Agent: manager-develop) |
| `7f591604f` | M2 — spec-workflow.md 양 사본 SSOT 궤적 수리 |
| `3686f85fc` | M3 — delivery.md 양 사본 auto-merge 중복 SSOT 위임 |
| (M4) | progress.md §E.2/§E.3 기록 + AC-11 검증 전용 폐쇄 |

`git rev-list --count develop..HEAD` 최종치는 §E.3에 기록. push는 레인 규율(로컬 develop 병합은 리드 창, 원격 push는 리드 일괄)에 따라 **의도적으로 수행하지 않음**.

### make agents-emit 필요성 보고 (plan.md §D 5번)

**필요함.** manager-git.md(`.claude/agents/moai/manager-git.md` 양 트리)를 편집했고, emit 대상은 `.claude/agents/moai/*.md` → `.codex/agents/moai/*.toml`이므로 run 종료 후 오케스트레이터가 `make agents-emit`을 실행해야 한다(본 레인은 실행 금지 — 지시 제약).

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec_id: SPEC-GIT-PROC-SAFE-001
run_complete_at: "2026-09-14T00:00:00Z"
run_commit_sha: "3686f85fc"   # M3 — 편집 커밋 최종점; M4 기록 커밋은 본 파일의 커밋 자신이라 자기 SHA 미기입(D3 관례상 자기 참조 불가)
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 6   # 대상 6파일 전량 범위 내 편집; 범위 밖 분기(frontmatter description, lint/크로스빌드, Route A/B 산문, TRACE PROBE) 무손상
l44_pre_commit_fetch: not-performed   # 격리 워크트리 내 전용 브랜치 작업 — pre-spawn sync check 불요(워크트리 격리 상태)
l44_post_push_fetch: not-performed    # push 미수행(레인 규율)
new_warnings_or_lints_introduced: 0
cross_platform_build: not-applicable   # 문서 전용 — 빌드 게이트 없음
total_run_phase_files: 7   # 6 대상 파일 + spec.md(frontmatter 전이) + progress.md
m1_to_mN_commit_strategy: per-milestone commit (M1-M3 + M4 record)
blockers_reported: 2   # 범위 밖 2차 스윕 적중 — §E.2 참조, 운영자 판정 대상
agents_emit_required: true
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec: SPEC-GIT-PROC-SAFE-001
card: t782
sync_complete_at: 2026-09-14
sync_commit_sha: pending-backfill
sync_status: completed
changelog_entry_position: CHANGELOG.md [Unreleased] ### Fixed (first bullet)
frontmatter_status_transitions:
  in-progress: implemented
  implemented: completed
  merged_into_sync_commit: true   # 3-phase close — no separate chore commit
b12_self_test_a_duplicate_grep: 0   # grep -c 'SPEC-GIT-PROC-SAFE-001' CHANGELOG.md before emission
b12_self_test_b_ac_count: 8         # acceptance.md live AC identifiers == CHANGELOG AC references
b12_self_test_c_paths_verified: true # all doctrine paths named in the entry test -f verified
codemaps: not-applicable — docs-only SPEC; no Go code surface changed, codemap extraction scope unchanged
mx_tags: not-applicable — no exported functions or code annotations touched; doctrine prose only
scope: docs-only — 6 doctrine files (local + template copies) + codex agent emit mirror; zero Go/runtime change
```

Sync summary: CHANGELOG entry added under `[Unreleased]` → `### Fixed`; spec.md frontmatter carried `in-progress → implemented → completed` on this single sync commit (`status` + `updated` only, zero body edits in spec.md/plan.md/acceptance.md); the 6 doctrine files from the run phase (commits `75cc17a29`..`1138b8daa`, agents-emit `b6984e8e9`) are frozen and untouched by sync. `sync_commit_sha` is the D3-convention placeholder `pending-backfill` — a commit cannot contain its own SHA — and is backfilled by the orchestrator in a follow-up commit.
