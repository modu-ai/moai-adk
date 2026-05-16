# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — Implementation Plan

> **Plan owner**: manager-spec (plan-phase) → manager-develop (run-phase) → manager-docs (sync-phase)
> **Strategy**: plan-in-main + run-in-main + sync-in-main (NO worktree per `feedback_worktree_never_use`)
> **PR lifecycle**: 3 PRs (plan / run / sync) — 모두 `--squash` auto-merge

## 1. Implementation Strategy

본 SPEC은 **metadata-only**다. 17 SPEC frontmatter의 `status` + `updated` 필드 + (선택적) `lint.skip` 필드만 갱신한다. TDD/DDD cycle은 부적용하며, 대신 **3-Wave evidence-driven workflow** 를 적용한다.

```
Wave 1 (Analysis)         Wave 2 (Apply)            Wave 3 (Verify)
─────────────────         ──────────────            ───────────────
B-12 SPECs git log    →   Category A bulk apply  →  moai spec lint --strict
+ external evidence  →    Category B mech apply  →  17 → 0 warning 확인
+ decision recorded  →    grouped by mechanism  →   AC-SDF002-X-001 충족
```

### 1.1 Wave 1 — Analysis (per-SPEC Category B)

목적: Category B 12건 각각에 대해 (a) git log evidence, (b) 외부 증거 (CHANGELOG / project memory / merged PR), (c) 채택된 mechanism (sync-commit / lint.skip / frontmatter downgrade) 을 결정.

수행자: manager-develop (run-phase 시작 직후)

산출물: 본 plan.md §3 "Category B Analysis Table" 갱신 (decision + evidence inline)

### 1.2 Wave 2 — Apply

목적: 결정된 mechanism 을 17 SPECs 에 적용.

수행자: manager-develop (Wave 1 직후)

산출물:
- Wave 2-A commit (1개): Category A 5건 frontmatter sync-up (`status` + `updated` 갱신)
- Wave 2-B-sync commit (1개, 필요 시): Category B 중 sync-commit mechanism 채택 SPECs 일괄 `sync(spec): SPEC-X — status closeout under FOLLOWUP-002` form
- Wave 2-B-skip commit (1개, 필요 시): Category B 중 lint.skip mechanism 채택 SPECs frontmatter `lint.skip: [StatusGitConsistency]` 추가 + HISTORY entry reason 기록
- Wave 2-B-downgrade commit (1개, 필요 시): Category B 중 frontmatter downgrade mechanism 채택 SPECs frontmatter `status` 하향 + HISTORY entry 사유 기록

총 commit 개수 예상: 2-4개 (Category A 1개 + Category B mechanism 별 1-3개).

### 1.3 Wave 3 — Verify

목적: `moai spec lint --strict` 가 `0 error(s), 0 warning(s)` 를 출력하는지 확인 + HARNESS-001/002/003 LSGF-001 회귀 부재 확인.

수행자: manager-develop (Wave 2 직후) → 자동 PR CI

산출물: lint 결과 log를 progress.md 또는 PR description 에 첨부.

---

## 2. Milestones (Priority-Ordered, No Time Estimates)

| Priority | Milestone                                  | Output | Phase |
|----------|--------------------------------------------|--------|-------|
| P1-1     | Wave 1 Category B per-SPEC analysis        | plan.md §3 갱신 (12-row decision table) | run-phase |
| P1-2     | Wave 2-A Category A bulk sync-up apply     | 1 commit, 5 SPEC files | run-phase |
| P1-3     | Wave 2-B Category B mechanism apply        | 1-3 commits (mechanism 별 grouped), 최대 12 SPEC files | run-phase |
| P1-4     | Wave 3 verification                        | lint clean log + AC-SDF002-X-001 충족 | run-phase |
| P2-1     | run-PR open with `--squash` auto-merge     | PR #N+1 (target main) | run-phase |
| P2-2     | sync-phase commit (status: completed)      | 1 commit, frontmatter status update + HISTORY 0.2.0 | sync-phase |
| P2-3     | sync-PR open with `--squash` auto-merge    | PR #N+2 (target main) | sync-phase |

**Sequential execution**: P1-1 → P1-2 → P1-3 → P1-4 → P2-1. P2-1 머지 후 P2-2 → P2-3.

---

## 3. Category B Analysis Table (run-phase 갱신 대상)

각 B-N 행은 plan-phase 시점 가설 + 증거 placeholder. run-phase에서 evidence를 채우고 mechanism을 확정한다.

| ID  | SPEC-ID                          | front → git    | git log evidence (top 2-3 commits)                                                                                                                                            | Walker miss 원인 가설                            | Mechanism (plan 가설 → run 결정) |
|-----|----------------------------------|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|-----------------------------------|
| B1  | SPEC-V3R2-ORC-003                | completed → implemented | `2c81e17b6 sync(specs): V3 bulk status drift 일괄 해소 (implemented → completed, 6 SPECs) (#926)` + `5a529892a feat(agents): SPEC-V3R2-ORC-003 M1 — Effort-Level Calibration Matrix test scaffolding (#897)` + `94e26cb6c plan(spec): SPEC-V3R2-ORC-003 — Effort-Level Calibration Matrix for 17 agents (#835)` | bulk sync commit `#926` 가 SPEC-V3R2-ORC-003 를 body에 명시했는지 확인 필요. 명시 안 했으면 walker는 feat M1 commit만 보고 `implemented` 추론. | 가설: `lint.skip + reason "synced via bulk-closure PR #926"`. 검증: run-phase에서 `git show --stat 2c81e17b6` body 확인. |
| B2  | SPEC-V3R2-RT-001                 | implemented → planned | `bdcb57f8d chore(spec): status drift 11건 sweep (#930)` + `3b6420c27 plan(spec): SPEC-V3R2-RT-001 — Hook JSON-OR-ExitCode Dual Protocol (#820)` | chore commit walker SKIP → plan commit만 hit → `planned`. RT-001 implementation event가 다른 SPEC 우산 아래 진행되었거나 frontmatter 가 잘못됨. | 가설: 외부 증거(CHANGELOG, project memory) 검증 후 (a) sync-commit OR (b) frontmatter downgrade `implemented → planned` 결정. 검증: run-phase에서 RT-001 implementation PR 존재 여부 확인. |
| B3  | SPEC-V3R2-RT-007                 | completed → implemented | `325a6492a docs(sync): RT-007 closure + status drift 일괄 정정 (4건) (#856)` + `aa10ce386 feat(migration): SPEC-V3R2-RT-007 — Hardcoded Path Fix + Versioned Migration (#846)` | docs(sync) 는 transitions.go에서 미인식 (prefix 가 `sync(` 아님 → walker는 `feat(migration)` 만 hit → `implemented`. | 가설: `sync(spec): SPEC-V3R2-RT-007 — status closeout under FOLLOWUP-002` commit 추가. 또는 `lint.skip + reason "synced via docs(sync) PR #856 — walker prefix limit"`. 검증: walker가 `docs(sync):` 도 인식하는지 transitions.go 확인. |
| B4  | SPEC-V3R2-SPC-002                | completed → planned | `2c81e17b6 sync(specs): V3 bulk status drift 일괄 해소 (implemented → completed, 6 SPECs) (#926)` + `4b810faa4 feat(mx): T-SPC002-02 Add HookSpecificOutput.MxTags field (#896)` + `73742e3ee plan(spec): SPEC-V3R2-SPC-002 — @MX TAG v2 hook JSON integration + sidecar index (#836)` | feat(mx) commit이 `T-SPC002-02` 형태(SPEC-V3R2-SPC-002 substring 매칭 안 됨)로 walker miss. plan commit만 hit → `planned`. | 가설: `lint.skip + reason "implementation tracked via T-SPC002-* subtask commits; synced via bulk-closure PR #926"`. 검증: T-SPC002 prefix 사용 다른 SPEC-V3R2-SPC-002 commit 존재 여부. |
| B5  | SPEC-V3R2-SPC-003                | implemented → planned | `bdcb57f8d chore(spec): status drift 11건 sweep (#930)` + `c959115e5 plan(spec): SPEC-V3R2-SPC-003 — backfill plan-phase artifacts (post-Wave 5) (#816)` + `03146d1ae feat(spec): SPEC-V3R2-SPC-003 — moai spec lint CLI (Wave 5) (#745)` | 가장 newest non-skip commit `c959115e5 plan(spec): backfill plan-phase artifacts` → `planned`. 실제 implementation은 `03146d1ae feat(spec): Wave 5` 이지만 plan backfill commit이 더 최신이라 walker는 plan을 채택. | 가설: `sync(spec): SPEC-V3R2-SPC-003 — implementation closeout` commit 추가. 또는 frontmatter `implemented` 유지 + lint.skip with reason "plan backfill commit overrode implementation; lint walker newest-first bias". 검증: 03146d1ae 가 진짜 implementation 완료 commit 인지 PR #745 확인. |
| B6  | SPEC-V3R2-WF-003                 | completed → implemented | `ecc40c4d9 sync(spec): SPEC-V3R2-HRN-003 — sync 단계 마무리 (#890)` + `2d153c471 feat(workflow): SPEC-V3R2-WF-003 Multi-Mode Router (#805)` | newest sync commit이 다른 SPEC (HRN-003) 을 지칭 → walker는 SPEC-V3R2-WF-003 명시 commit 인 `feat(workflow)` 만 hit → `implemented`. | 가설: `sync(spec): SPEC-V3R2-WF-003 — status closeout under FOLLOWUP-002` commit 추가. 검증: HRN-003 PR #890 body가 WF-003 closeout 도 포함했는지. |
| B7  | SPEC-V3R3-CI-AUTONOMY-001        | completed → in-progress | `19957efd8 sync(specs): V3 final status closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed) (#927)` + `6f9e035e2 docs(spec): SPEC-V3R3-CI-AUTONOMY-001 closure — 8-Wave + follow-up #795 (#796)` + `9ecd8c765 fix(bodp): anchor audit trail to git repo root, not os.Getwd() (#795)` | `sync(specs):` 는 transitions.go `strings.HasPrefix("sync", ...)` 로 인식 → `completed`. 그러나 SPEC-ID matching이 `SPEC-V3R3-CI-AUTONOMY-001` 정확 매칭일 때 #927 body가 이를 포함하는지 (LSGF-001 word-boundary 이후) 확인 필요. word-boundary 통과 시에도 다음 commit인 `fix(bodp)` → `in-progress` 가 newer 라면 walker는 fix 를 채택. | 가설: `lint.skip + reason "synced via #927 bulk-closure; walker reads later fix(bodp) commit"`. 검증: #927 body의 SPEC-ID 명시 + LSGF-001 word-boundary 통과 확인. |
| B8  | SPEC-V3R3-HARNESS-LEARNING-001   | completed → in-progress | `e8e38b17b sync(SPEC-V3R4-HARNESS-001): status transition draft → implemented (#911)` + `5c04b87a1 docs(spec): SPEC-V3R3-HARNESS-LEARNING-001 sync cleanup — implemented status + Wave B.3-5 done (#812)` + `68f023289 feat(harness): SPEC-V3R3-HARNESS-LEARNING-001 — Self-Learning Dynamic Harness (5-wave) (#728)` | `sync(SPEC-V3R4-HARNESS-001):` 는 LEARNING-001 명시 아님 (word-boundary post-LSGF-001 통과 못함). `docs(spec):` prefix 는 walker 미인식 → `feat(harness)` 만 hit → `implemented` 아닌 `implemented` (frontmatter `completed`). 음, 이는 git-implied `implemented` 인데 lint 보고는 `in-progress`. Walker가 더 newer non-target commit (e.g. `fix(hook)` `chore(...)` 등) 을 hit하는 것으로 보임. | 가설: `lint.skip + reason "sync rolled up under HARNESS-001 closeout; PR #812 docs(spec) prefix not recognized by walker"`. 검증: walker가 hit한 정확한 commit identify. |
| B9  | SPEC-V3R3-PROJECT-HARNESS-001    | completed → implemented | `e8e38b17b sync(SPEC-V3R4-HARNESS-001): ... (#911)` + `74169ec86 feat(harness): SPEC-V3R3-PROJECT-HARNESS-001 — 16Q 인터뷰 + 5-Layer 통합 (#727)` | sync commit이 HARNESS-001 명시이지 PROJECT-HARNESS-001 아님 (word-boundary 통과 못함). walker는 `feat(harness)` 만 hit → `implemented`. | 가설: `sync(spec): SPEC-V3R3-PROJECT-HARNESS-001 — status closeout under FOLLOWUP-002` commit 추가. 또는 lint.skip with reason. 검증: HARNESS-001 sync commit (#911) body가 PROJECT-HARNESS-001 closeout도 포함했는지. |
| B10 | SPEC-V3R3-RETIRED-AGENT-001      | completed → in-progress | `325a6492a docs(sync): RT-007 closure + status drift 일괄 정정 (4건) (#856)` + `90b849669 chore(changelog): SPEC-V3R3-RETIRED-AGENT-001 [Unreleased] entry backfill (#777)` + `20d77d931 feat(agent-runtime): SPEC-V3R3-RETIRED-AGENT-001 — retired stub 호환성 + manager-cycle 템플릿 정합화 (#776)` | `docs(sync):` prefix 미인식 + `chore(changelog):` 은 chore-skip filter SKIP → `feat(agent-runtime)` 만 hit → `implemented` 아닌 `in-progress` (이상함). 가장 가능성 높은 시나리오: walker가 hit한 newest non-skip commit이 다른 SPEC을 지칭하면서 RETIRED-AGENT-001을 body에 mention → in-progress classification. | 가설: `sync(spec): SPEC-V3R3-RETIRED-AGENT-001 — status closeout` commit 추가. 또는 lint.skip with reason. 검증: 실제 walker hit commit 식별. |
| B11 | SPEC-V3R3-RETIRED-DDD-001        | completed → implemented | `e8e38b17b sync(SPEC-V3R4-HARNESS-001): ... (#911)` + `15345b7a7 fix(hook): hooks 패키지 전면 개선 P0/P1/P2 — ddd_handler 삭제 + retired stub 정합화 (#858)` + `ea8fb3e6b feat(agent-runtime): SPEC-V3R3-RETIRED-DDD-001 — manager-ddd retired stub 표준화 (#781)` | `fix(hook)` prefix 미인식 (transitions.go에 없음) → newer non-target commits만 → `feat(agent-runtime)` → `implemented`. | 가설: `sync(spec): SPEC-V3R3-RETIRED-DDD-001 — status closeout` commit 추가. 검증: PR #781 sync 단계 존재 여부 + project memory. |
| B12 | SPEC-V3R4-LINT-SKIP-CLEANUP-001  | completed → planned | `b14290946 sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 7 sync — Pattern H 잔여 closeout + CHANGELOG (#940)` + `ce779f9ee feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 77건 status drift 일괄 해소 + terminal-state exemption (#939)` + `602a07c84 plan(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 plan 초기 작성 (#938)` + `758341089 chore(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — 55 SPECs lint.skip 일괄 제거 (#937)` | `758341089 chore(spec):` 가 LSKC-001 implementation 인데 chore-skip filter에 의해 SKIP. SDF-001 commits 는 LSKC-001 body 미포함 → walker는 LSKC-001 commit 중 chore-skip 후 next hit가 plan commit이 아닌 다른 SPEC 우산. 결과적으로 walker hit는 `plan(spec): LSKC-001` commit 인 듯. (#935 plan PR) → `planned`. | 가설: `lint.skip + reason "implementation tracked via chore(spec) PR #937 (walker chore-skip); sync rolled up under SDF-001 PR #939/#940"`. 이는 SDF-001 chain self-drift 패턴의 직계 후속. 검증: 실제 walker hit commit. |

> **Note**: Wave 1 run-phase 시점에 위 표의 "git log evidence" 와 "Mechanism (plan 가설 → run 결정)" 두 열을 update. 만약 plan 가설이 검증 실패 시 mechanism 변경 + 사유 inline 기록.

---

## 4. Technical Approach

### 4.1 Wave 2-A — Category A bulk sync-up

수행 절차 (run-phase, manager-develop):

```bash
# Wave 2-A: Category A 5건 frontmatter sync-up
for SPEC_DIR in SPEC-GLM-MCP-001 SPEC-STATUSLINE-001 SPEC-V3R2-WF-002 SPEC-V3R4-CATALOG-001 SPEC-WORKTREE-002; do
  # Edit tool 사용 (sed 금지, 검증 가능한 Edit operation)
  # status: in-progress|implemented  → status: <git-implied target>
  # updated: <old date>              → updated: 2026-05-16
  # + HISTORY 신규 row 추가: "0.X.Y | 2026-05-16 | manager-develop (FOLLOWUP-002 Wave 2-A) | <SPEC-A-N> sync-up: status <old> → <new>, walker LSGF-001 post-fix evidence aligned."
done
git add -A
git commit -m "$(cat <<'EOF'
sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-A — Category A 5건 frontmatter sync-up

LSGF-001 walker word-boundary fix 후 노출된 forward drift 5건:
- SPEC-GLM-MCP-001       (in-progress → completed)
- SPEC-STATUSLINE-001    (in-progress → implemented)
- SPEC-V3R2-WF-002       (in-progress → implemented)
- SPEC-V3R4-CATALOG-001  (implemented → completed)
- SPEC-WORKTREE-002      (implemented → completed)

각 SPEC frontmatter `status` + `updated` 갱신 + HISTORY entry 추가.
AC-SDF002-A-1..A-5 충족 예정.

🗿 MoAI <email@mo.ai.kr>
EOF
)"
```

### 4.2 Wave 2-B — Category B per-mechanism apply

mechanism 별로 grouped commit. Wave 1 결정 기반:

```bash
# Mechanism 1: sync-commit (Category B 중 sync-commit 채택 SPECs)
# 예시 — Wave 1 결정 결과 X건 채택 시:
for SPEC in <determined-sync-commit-targets>; do
  : # frontmatter는 손대지 않고 commit message만 walker에 visible 하게
done
git commit -m "$(cat <<'EOF'
sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-B (sync) — Category B sync visibility closeout

walker visibility 한계로 sync event 미인식되었던 SPECs:
- <list-from-wave-1-decisions>

Each SPEC status는 git-implied < frontmatter 상태이지만 실제 sync는 bulk closure
PR 우산 아래 완료됨. 본 commit이 `sync(spec):` prefix + SPEC-X 명시로
walker가 인식할 수 있는 sync visibility 생성.

🗿 MoAI <email@mo.ai.kr>
EOF
)"

# Mechanism 2: lint.skip (Category B 중 lint.skip 채택 SPECs)
for SPEC in <determined-lint-skip-targets>; do
  : # spec.md frontmatter에 lint.skip: [StatusGitConsistency] 추가
  : # HISTORY entry에 reason 기록
done
git commit -m "$(cat <<'EOF'
docs(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-B (skip) — Category B lint.skip 등록 + reason 기록

walker가 인식 불가한 mechanism(docs(sync):, fix(hook):, T-SPC* subtask 등)으로
implementation/sync 완료된 SPECs. 각 SPEC frontmatter lint.skip:
[StatusGitConsistency] 추가 + HISTORY entry에 root cause + 인용 PR 기록.

대상:
- <list-from-wave-1-decisions>

🗿 MoAI <email@mo.ai.kr>
EOF
)"

# Mechanism 3: frontmatter downgrade (Category B 중 downgrade 채택 SPECs)
for SPEC in <determined-downgrade-targets>; do
  : # spec.md frontmatter status 하향 + HISTORY entry 사유 기록
done
git commit -m "$(cat <<'EOF'
fix(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-B (downgrade) — Category B frontmatter 정정

git log + 외부 증거(CHANGELOG, project memory) 모두 implementation 미완료 또는
partial 만 입증되는 SPECs frontmatter status 하향:
- <list-from-wave-1-decisions>

🗿 MoAI <email@mo.ai.kr>
EOF
)"
```

### 4.3 Wave 3 — Verification

```bash
# 1) Lint clean 확인 (AC-SDF002-X-001)
moai spec lint --strict 2>&1 | tail -1
# expected: "0 error(s), 0 warning(s)"

# 2) HARNESS-001/002/003 회귀 부재 확인 (AC-SDF002-X-002)
moai spec lint --strict 2>&1 | grep "HARNESS-00[123]" | wc -l
# expected: 0

# 3) Scope discipline 확인 (AC-SDF002-X-003)
git diff main --stat | grep -v "\.moai/specs/.*spec\.md" | wc -l
# expected: 0 (모든 변경 파일이 .moai/specs/*/spec.md 패턴)
```

### 4.4 Sync-phase commit

```bash
# spec.md frontmatter:
#   status: draft → completed
#   updated: 2026-05-16
#   version: "0.1.0" → "0.2.0"
# + HISTORY 0.2.0 entry 추가 (sync-phase 완료 + AC binary 0 달성 기록)
# + CHANGELOG.md Unreleased 섹션에 FOLLOWUP-002 entry 추가 (ko + en)

git commit -m "$(cat <<'EOF'
sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 lifecycle COMPLETE — 17건 status drift binary 0 달성

LSGF-001 walker word-boundary fix 후 노출된 17건 status drift 모두 해소.
- Category A 5건: frontmatter forward sync-up
- Category B 12건: per-SPEC mechanism (sync-commit / lint.skip / downgrade)

AC-SDF002-X-001..X-003 모두 PASS.
`moai spec lint --strict` → 0 error(s), 0 warning(s).
v3.0.0-rc1 release gate (FOLLOWUP-002 + CI-INFRA-FIX-001) 중 본 SPEC closure 완료.

🗿 MoAI <email@mo.ai.kr>
EOF
)"
```

---

## 5. Risks and Mitigation

### Risk-1: Walker가 lint.skip 추가로 처리 못한 hidden drift 노출

**상황**: 17건 처리 후 lint 가 다른 SPEC에서 새로운 WARNING 을 노출할 수 있음 (예: lint.skip 추가가 walker filter pattern 변경에 인접).

**Mitigation**: Wave 3 verification 에서 `moai spec lint --strict` 전체 결과를 PR description 에 첨부. 만약 새 WARNING 노출 시 SPEC 범위 확장 (17 → N) 하지 않고 별도 follow-up SPEC 발급.

### Risk-2: Wave 1 가설이 실측과 어긋남 (특히 B-군 walker miss 원인)

**상황**: plan-phase 가설은 `git log --grep` 의 commit listing 만으로 추론. 실제 walker가 hit 한 commit이 다를 수 있음.

**Mitigation**: Wave 1 에서 `moai spec lint --strict --verbose` 또는 walker debug log(존재 시) 활용. 만약 가설 검증 실패 시 mechanism 변경 + plan.md §3 evidence column inline 갱신. plan-auditor 가 본 plan 의 가설 voracity 보다는 process 의 evidence-driven 성을 평가하므로 가설 정확도 자체는 critical 아님.

### Risk-3: frontmatter downgrade 결정이 잘못된 SPEC closure 신호

**상황**: Category B 중 frontmatter downgrade 결정 시, 만약 실제로는 완료된 SPEC을 partial-implementation으로 잘못 marking하면 프로젝트 추적성 손상.

**Mitigation**: REQ-SDF002-003 에 따라 downgrade는 (a) git log evidence 부재 + (b) 외부 증거 (CHANGELOG, project memory, merged PR) 부재 두 조건 모두 충족 시에만 적용. 단 하나라도 완료 증거 있으면 sync-commit 또는 lint.skip 선택.

### Risk-4: lint.skip 남발로 walker 의 실제 가치 훼손

**상황**: 17건 모두 lint.skip 으로 처리 시 walker의 status drift detection 기능 자체가 무력화.

**Mitigation**: lint.skip 는 **walker visibility 한계가 명백한 경우에만** 사용. 가능한 한 sync-commit mechanism 우선. REQ-SDF002-004 escalation 조건 명시 (sync-commit / lint.skip / downgrade 모두 부적절 시 별도 root-cause SPEC).

### Risk-5: 본 SPEC 자체의 sync commit이 또 다른 status drift 유발

**상황**: 본 SPEC sync-phase commit `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002` 이 다른 SPEC의 git-implied status를 또 우회하는 경우 (SDF-001 self-drift 패턴 재현).

**Mitigation**: 본 SPEC sync commit은 `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 lifecycle COMPLETE` 형태로 자기 SPEC ID만 명시 + 다른 SPEC ID 본문 미포함. walker 가 본 commit 을 다른 SPEC drift signal 로 오인식 못하도록.

---

## 6. Rollback Plan

### 6.1 Wave 2 rollback (run-phase 중)

만약 Wave 2 적용 commit이 lint 결과를 악화시키거나 회귀 유발 시:

```bash
git reset --hard HEAD~<N>  # Wave 2 commits 수만큼 revert
moai spec lint --strict 2>&1 | tail -1  # 원래 17 warning 으로 복귀 확인
```

run-PR 미머지 상태이므로 main 영향 없음.

### 6.2 sync-phase rollback (sync-PR open 후 lint regression 노출 시)

```bash
gh pr close <sync-PR-number>
gh pr close <run-PR-number>
# run-phase commit 도 main 미머지 상태 → revert 불필요
```

### 6.3 Post-merge rollback (만약 main 머지 후 회귀 발견)

```bash
# 새 fix-up PR 발급 (revert PR)
git revert <merge-commit-sha> --no-commit
# + 추가 fix 적용 후 새 PR
```

run/sync PR 머지 commits은 모두 `--squash` 이므로 단일 commit revert 로 깔끔히 되돌릴 수 있음.

---

## 7. Telemetry & Observability

### 7.1 Pre-fix baseline (main `139c4d9d0`)

```
moai spec lint --strict 2>&1 | tail -1
# 0 error(s), 17 warning(s)
```

### 7.2 Post-fix target

```
moai spec lint --strict 2>&1 | tail -1
# 0 error(s), 0 warning(s)
```

### 7.3 Run-phase progress.md 형식

run-phase에서 `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/progress.md` 생성:

```markdown
# Progress — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002

## Wave 1 (Analysis) — <timestamp>

- B-12 each git log evidence captured
- mechanism 결정: <sync-commit count> + <lint.skip count> + <downgrade count> = 12
- 가설 정확도: <X/12 verified, Y/12 revised>

## Wave 2-A (Category A) — <timestamp>

- 5 files modified: SPEC-GLM-MCP-001, ..., SPEC-WORKTREE-002
- 1 commit: <sha>

## Wave 2-B (Category B) — <timestamp>

- N commits applied (mechanism: <sync/skip/downgrade counts>)
- Files modified: <list>

## Wave 3 (Verification) — <timestamp>

- `moai spec lint --strict 2>&1 | tail -1`: "0 error(s), 0 warning(s)"
- HARNESS-001/002/003 회귀 부재 확인
- Scope discipline: `git diff main --stat` all files in `.moai/specs/*/spec.md`

## PR Status

- run-PR #<N+1>: <status>
- sync-PR #<N+2>: <status>
```

---

## 8. Open Questions (OQ block)

### OQ-1: docs(sync): prefix walker recognition 검토

**Question**: `transitions.go` 가 `docs(sync):` prefix 를 `sync-merge` 로 인식하는가?

**Status**: 미해결. plan-phase 미검증.

**Resolution Path**: run-phase Wave 1 에서 `internal/spec/transitions.go:transitions` map 확인. 만약 미인식 시:
- Option (a): 본 SPEC scope 내 transitions.go 추가 (REQ-SDF002-006 위반 — out of scope)
- Option (b): B3 (SPEC-V3R2-RT-007) 에 sync(spec) commit 추가
- Option (c): B3 에 lint.skip with reason "docs(sync) prefix per design not recognized"

**Plan-phase 권장**: Option (b) 또는 (c). 별도 walker 확장은 새 SPEC.

### OQ-2: Walker bulk-sync-commit recognition 한계

**Question**: walker가 `sync(specs): V3 bulk closeout (SPEC-X + SPEC-Y → completed) (#NNN)` 같은 multi-SPEC sync commit 의 body 를 어떻게 처리하는가? word-boundary 통과 후 commit body 도 grep 대상인가?

**Status**: 미해결. run-phase 검증.

**Resolution Path**: `git log --grep` 기본 동작은 commit message (subject + body) 매칭. word-boundary 격상 후에도 body 매칭은 유지. 따라서 만약 bulk sync commit body 에 SPEC-V3R2-ORC-003 명시되어 있으면 walker가 인식해야 함. 만약 명시 안 되어 있으면 walker miss → lint.skip 필요.

**Plan-phase 결정**: Wave 1 에서 각 bulk sync commit (`#926`, `#927`, `#856`) body 를 `git show --format=full <sha>` 로 점검 + mechanism 결정.

### OQ-3: feat(mx): T-SPC002-02 같은 subtask prefix 인식

**Question**: walker가 `T-SPC002-02` 같은 subtask 식별자 (SPEC-V3R2-SPC-002 substring 의 변형) 를 SPEC-V3R2-SPC-002 로 인식하는가?

**Status**: 미해결.

**Resolution Path**: post-LSGF-001 word-boundary 매칭은 `\bSPEC-V3R2-SPC-002\b` (정확 매칭) 이므로 `T-SPC002-02` 는 매칭 실패. walker miss 확정. → B4 lint.skip 채택 권장.

### OQ-4: SDF-001 chain self-drift 처리 정책

**Question**: SDF-001 chain (LSCSK-001, LSKC-001, SDF-001, FOLLOWUP-002) 의 closeout commit 이 항상 next-chain SPEC 우산 아래 발생하는 self-drift 패턴은 walker 한계인가 정책 한계인가?

**Status**: 미해결.

**Resolution Path**: 본 SPEC scope 내에서는 lint.skip with reason 채택 (B12 case). 영구 해소는 별도 SPEC `SPEC-V3R4-WALKER-BULK-SYNC-RECOGNITION-001` 후보 (out of scope).

---

## 9. References (cross-file)

- `spec.md` — REQ ↔ AC mapping
- `acceptance.md` — per-SPEC AC 시나리오 17 + cross-cutting AC 3
- `design.md` — design rationale (Category A vs B 구분 + mechanism 선택 기준)
- `tasks.md` — Wave 1/2/3 task breakdown
- `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/plan.md` — walker word-boundary fix 디자인 (직계 trigger)
- `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/plan.md` — Pattern A~H 처리 precedent (5-Wave 모델)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema (SPECLINT-DEBT-002 SSOT)
