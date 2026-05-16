---
id: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002
title: "LSGF-001 side-effect status drift sweep — 17건 per-SPEC remediation"
version: "0.1.0"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: manager-spec
priority: P1
phase: "v3.0.0 R4 — Foundation Cleanup"
module: ".moai/specs/*/spec.md (17 SPEC frontmatter targets — metadata-only)"
lifecycle: spec-anchored
tags: "v3r4, spec-lint, status-drift, status-git-consistency, frontmatter-sync, lsgf-001-aftermath, sdf-001-successor, foundation"
issue_number: null
related_specs:
  - SPEC-V3R4-LINT-SPECID-GREP-FIX-001
  - SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
  - SPEC-V3R4-SPECLINT-DEBT-002
  - SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001
depends_on:
  - SPEC-V3R4-LINT-SPECID-GREP-FIX-001
breaking: false
bc_id: []
related_theme: "Foundation Cleanup — Status drift permanent closure"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — LSGF-001 side-effect status drift sweep (17건 per-SPEC remediation)

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.0   | 2026-05-16 | manager-spec  | 초기 draft. SPEC-V3R4-LINT-SPECID-GREP-FIX-001 (PR #947+#948+#949, main `139c4d9d0`) 머지 후 walker word-boundary 매칭 격상이 substring noise를 제거하면서 그 동안 가려져 있던 17건의 실제 status drift가 노출됨. SDF-001 (Pattern H closeout)의 직계 후속. Category A 5건(forward sync-up) + Category B 12건(per-SPEC analysis: sync-commit / lint.skip / frontmatter downgrade) 두 단계로 처리. BODP 평가: ¬a ¬b ¬c → base `origin/main` fresh branch (plan-in-main, worktree 미사용 per `feedback_worktree_never_use`). 본 SPEC + CI-INFRA-FIX-001은 v3.0.0-rc1 release readiness gate. |

---

## 1. Goal

`SPEC-V3R4-LINT-SPECID-GREP-FIX-001` (PR #947+#948+#949 lifecycle COMPLETE) walker word-boundary precision fix의 의도된 부수효과로 main에서 `moai spec lint --strict` 가 **17 WARNING** 을 보고한다. 모든 17건은 LSGF-001 이전부터 존재하던 진짜 status drift이며 substring collision noise에 의해 가려져 있던 것이다.

본 SPEC은 17건을 두 카테고리로 분류하여 **per-SPEC 분석 + 정밀 remediation** 으로 해소하고 `0 ERROR + 0 WARNING` 상태를 달성한다.

### 1.1 Binary Success Criterion (AC-SDF002-X-001)

본 SPEC 머지 후 main에서 다음 명령이 정확히 `0 warning(s)` 를 출력해야 한다:

```bash
moai spec lint --strict 2>&1 | tail -1
# expected output: "0 error(s), 0 warning(s)"
```

### 1.2 Lineage 위치

본 SPEC은 lint cleanup chain의 5번째 wave다:

```
SPECLINT-DEBT-001 (sweep + skip 51건)
  └─ LINT-STATUS-CHORE-SKIP-001 (engine walker filter)
       └─ LINT-SKIP-CLEANUP-001 (51건 skip 제거, real drift 노출)
            └─ STATUS-DRIFT-FOLLOWUP-001 (64건 → 0; Pattern A~H 처리)
                 └─ LINT-SPECID-GREP-FIX-001 (walker word-boundary; substring noise 영구 차단)
                      └─ STATUS-DRIFT-FOLLOWUP-002 (본 SPEC; substring noise 제거로 노출된 17건 정밀 해소)
```

### 1.3 17 Drift Inventory (verified at main HEAD `139c4d9d0`)

#### Category A: Forward drift (frontmatter < git-implied, 5건)

frontmatter status가 git-implied status보다 뒤쳐진 케이스. 단순 sync-up.

| # | SPEC-ID | frontmatter | git-implied | Recommended fix |
|---|---------|-------------|-------------|-----------------|
| A1 | SPEC-GLM-MCP-001       | in-progress | completed   | frontmatter `status: completed` |
| A2 | SPEC-STATUSLINE-001    | in-progress | implemented | frontmatter `status: implemented` |
| A3 | SPEC-V3R2-WF-002       | in-progress | implemented | frontmatter `status: implemented` |
| A4 | SPEC-V3R4-CATALOG-001  | implemented | completed   | frontmatter `status: completed` |
| A5 | SPEC-WORKTREE-002      | implemented | completed   | frontmatter `status: completed` |

**처리 방향**: Wave 2-A 단일 commit으로 5건 frontmatter `status` + `updated: 2026-05-16` 일괄 갱신.

#### Category B: Suspect downgrade (frontmatter > git-implied, 12건)

frontmatter가 git-implied보다 더 진행됨을 주장. 각 건마다 개별 분석 필수.

| # | SPEC-ID | frontmatter | git-implied | Investigation type |
|---|---------|-------------|-------------|--------------------|
| B1  | SPEC-V3R2-ORC-003               | completed   | implemented | sync commit visibility (#926 bulk closeout) |
| B2  | SPEC-V3R2-RT-001                | implemented | planned     | Big gap — implementation event 미인지 |
| B3  | SPEC-V3R2-RT-007                | completed   | implemented | sync commit visibility (#856 docs(sync)) |
| B4  | SPEC-V3R2-SPC-002               | completed   | planned     | Big gap — feat(mx) T-SPC002-02 만 인식 |
| B5  | SPEC-V3R2-SPC-003               | implemented | planned     | Big gap — feat(spec) Wave 5 + plan만 인식 |
| B6  | SPEC-V3R2-WF-003                | completed   | implemented | sync commit visibility (HRN-003 sync에 묻힘) |
| B7  | SPEC-V3R3-CI-AUTONOMY-001       | completed   | in-progress | sync commit visibility (#927 bulk closeout) |
| B8  | SPEC-V3R3-HARNESS-LEARNING-001  | completed   | in-progress | sync commit visibility (HARNESS-001 sync에 묻힘) |
| B9  | SPEC-V3R3-PROJECT-HARNESS-001   | completed   | implemented | sync commit visibility (HARNESS-001 sync에 묻힘) |
| B10 | SPEC-V3R3-RETIRED-AGENT-001     | completed   | in-progress | sync commit visibility (#856 RT-007 우산) |
| B11 | SPEC-V3R3-RETIRED-DDD-001       | completed   | implemented | sync commit visibility (#858 fix(hook) 우산) |
| B12 | SPEC-V3R4-LINT-SKIP-CLEANUP-001 | completed   | planned     | Big gap — sync commit이 SDF-001 우산 (#939/#940) 아래 |

**처리 방향**: Wave 1에서 per-SPEC git log 분석으로 each fix 결정. Wave 2-B에서 결정된 fix 적용 (mechanism 별 grouped commits: sync(spec) sweep / lint.skip 등록 / frontmatter downgrade).

### 1.4 가설 (run-phase 검증 대상)

plan-phase 임시 가설 (run-phase에서 git log 정밀 분석 + 외부 증거 대조 후 final 결정):

- **B1, B3, B6, B7, B8, B9, B10, B11 (sync visibility 군 — 8건)**: 실제 sync는 발생했으나 commit subject가 다른 SPEC을 지칭하거나 bulk closure 형태였음 (예: `19957efd8 sync(specs): V3 final status closeout (CI-AUTONOMY-001 + HARNESS-001 in-progress → completed)`). 가설 fix:
  - 우선 채택: **개별 `sync(spec): SPEC-X status closeout under FOLLOWUP-002` commit 추가** (walker가 인식하도록 `sync(spec):` prefix, body에 `SPEC-X` 명시)
  - 차선: **`lint.skip: [StatusGitConsistency]` with reason "synced via bulk-closure PR #XXX"**
- **B2, B4, B5 (Big gap 군 — 3건)**: implementation event가 다른 SPEC 우산 아래 진행되었거나 frontmatter가 단순히 잘못됨. 가설 fix:
  - 우선 채택: git log + CHANGELOG + project memory 종합 검증 후 **frontmatter downgrade** (만약 외부 증거가 partial implementation 만 입증)
  - 차선: **lint.skip with 인용 PR #XXX** (만약 외부 증거가 완료를 입증하지만 walker visibility 한계)
- **B12 (SDF-001 chain self-drift)**: LSKC-001 sync commit `b14290946 sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 Wave 7 sync` 이 LSKC-001 closeout을 SDF-001 우산 아래 흡수. SDF-001 sync chain 의 알려진 부작용. 가설 fix:
  - 우선 채택: **`lint.skip: [StatusGitConsistency]` + HISTORY reason "closed under SDF-001 Wave 7 (PR #939/#940)"**
  - 차선: **개별 sync(spec) commit** (그러나 sync(spec): SPEC-V3R4-LINT-SKIP-CLEANUP-001 — already closed by SDF-001 같은 retroactive commit 은 noise 가중)

run-phase는 각 B-N에 대해 (a) git log evidence, (b) 채택된 mechanism, (c) 적용된 변경 을 plan.md "Category B Analysis Table" 갱신 형태로 기록한다.

---

## 2. Scope

### 2.1 In Scope

- `.moai/specs/SPEC-*/spec.md` frontmatter `status` + `updated` 필드 갱신 (최대 17 파일).
- 필요한 경우 `lint.skip: [StatusGitConsistency]` 등록 + 본문 HISTORY entry 에 reason 기록.
- 필요한 경우 `sync(spec): SPEC-X — status closeout under FOLLOWUP-002` 형태의 sync commit (run-phase에서 walker가 인식하도록 `sync(spec):` prefix 사용).
- 본 SPEC의 plan-PR (현 phase) + run-PR + sync-PR 3-PR cycle.
- 각 Category B SPEC에 대한 per-SPEC analysis log (plan.md "Category B Analysis Table" 갱신, run-phase).

### 2.2 Out of Scope

- **walker code 수정 금지**: `internal/spec/drift.go`, `internal/spec/transitions.go` 변경 절대 금지. LSGF-001은 영구 fix이며 추가 변경은 회귀 리스크.
- **lint rule 정의 변경 금지**: `StatusGitConsistency` rule severity, scope, threshold 모두 불변.
- **CI workflow 변경 금지**: 별도 SPEC-V3R4-CI-INFRA-FIX-001 scope.
- **walker behavior 재설계 금지**: 만약 운영상 walker 가 sync visibility 한계로 인해 반복적 false-positive 를 일으킨다면 별도 follow-up SPEC (예: `SPEC-V3R4-WALKER-BULK-SYNC-RECOGNITION-001`).
- **17건 외 SPEC frontmatter touch 금지**: lint clean SPECs 는 건들지 않는다.
- **본문(Body) 콘텐츠 수정 금지**: spec.md / plan.md / acceptance.md 본문 내용 변경 없음. frontmatter + HISTORY entry 만.
- **신규 lint rule 신설 금지**.
- **bulk script (sweep) 신설 금지**: 17건 모두 개별 분석을 거치므로 sweep 스크립트는 부적절. Category A 5건만 단순 sed-style edit 허용.

---

## 3. Requirements (EARS Format)

### REQ-SDF002-001 (Ubiquitous)

The system SHALL maintain `moai spec lint --strict` warning count at exactly **zero** for all `StatusGitConsistency` rules after this SPEC closes.

**Acceptance**: AC-SDF002-X-001

### REQ-SDF002-002 (Event-Driven)

**WHEN** walker substring-noise removal (LSGF-001) exposes previously-hidden drift, **THE followup SPEC SHALL** classify each finding into Category A (forward sync-up) or Category B (suspect downgrade), then apply the appropriate remediation mechanism (frontmatter sync-up / sync-commit / lint.skip / frontmatter downgrade).

**Acceptance**: AC-SDF002-A-1..A-5, AC-SDF002-B-1..B-12

### REQ-SDF002-003 (State-Driven)

**WHILE** the suspect-downgrade Category B is being analyzed, **THE SPEC SHALL** preserve frontmatter values that are validated by external evidence (PRs merged, SPECs documented as complete in CHANGELOG, project memory entries). frontmatter downgrade is permitted **only** when git log + external evidence agree.

**Acceptance**: AC-SDF002-B-1..B-12 (each B-AC enforces evidence-based decision)

### REQ-SDF002-004 (Unwanted)

**IF** a drift cannot be safely resolved by sync-up / sync-commit / lint.skip, **THEN** the SPEC SHALL escalate the case to a separate root-cause SPEC (e.g., `SPEC-V3R4-WALKER-BULK-SYNC-RECOGNITION-001`) rather than force-clearing the warning via blanket frontmatter mutation.

**Acceptance**: AC-SDF002-X-001 (zero-warning state via legitimate mechanisms only)

### REQ-SDF002-005 (Backward-Compatible)

The fix SHALL NOT regress LSGF-001 walker behavior (word-boundary precision) or LSCSK-001 chore-skip semantics. No changes to `internal/spec/drift.go` or `internal/spec/transitions.go`.

**Acceptance**: AC-SDF002-X-002

### REQ-SDF002-006 (Ubiquitous — Scope discipline)

The fix SHALL respect the SPEC discipline — **no other SPEC's content or implementation files are touched** beyond the 17 frontmatter targets. Specifically: no `internal/`, no `pkg/`, no `cmd/`, no `.claude/skills/`, no `.claude/agents/`, no template content modifications.

**Acceptance**: AC-SDF002-X-003

---

## 4. Exclusions (What NOT to Build)

1. **walker 코드 변경 금지** — `internal/spec/drift.go`, `internal/spec/transitions.go` 본문 + 테스트 모두 손대지 않는다. LSGF-001 영구 fix 유지.
2. **bulk sweep script 신설 금지** — Category A 5건만 단순 status field edit. Category B 12건은 각 SPEC별 git log 분석 후 mechanism 결정.
3. **lint.skip 남발 금지** — `lint.skip` 사용 시 반드시 HISTORY entry에 "synced via bulk-closure PR #XXX" 또는 "frontmatter overrides walker due to multi-SPEC sync commit" 같은 명확한 reason 기록.
4. **frontmatter mass-downgrade 금지** — Category B 12건 중 frontmatter downgrade 결정은 git log + 외부 증거 (CHANGELOG, project memory, merged PR) 모두 일치할 때만 허용.
5. **새 lint rule 또는 새 SPEC schema field 신설 금지**.
6. **CI workflow 변경 금지** — 별도 SPEC-V3R4-CI-INFRA-FIX-001 scope.
7. **본문 콘텐츠 수정 금지** — `spec.md` 본문, `plan.md`, `acceptance.md`, `design.md`, `research.md` 내용 변경 없음. frontmatter `status` + `updated` + (선택적) `lint.skip` 만.
8. **HARNESS-001/002/003 spec.md touch 금지** — LSGF-001 fix 후 자동 해소되었고 본 SPEC 17건 inventory에 포함되지 않음.

---

## 5. Constraints

- **Code language**: 본 SPEC은 metadata-only — Go 코드 변경 없음.
- **Comment language**: Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`). frontmatter HISTORY entry는 Korean.
- **Commit message language**: Korean (per `git_commit_messages: ko`).
- **Schema**: 12-field canonical frontmatter (per SPEC-V3R4-SPECLINT-DEBT-002 SSOT `.claude/rules/moai/development/spec-frontmatter-schema.md`).
- **TDD discipline**: 본 SPEC은 metadata-only — TDD cycle 부적용. 대신 **Wave 1 (Analysis) → Wave 2 (Apply) → Wave 3 (Verify)** 3-Wave 적용.
- **Performance budget**: 17건 frontmatter edit + lint run 1회 — 총 작업시간 영향 미미.
- **MX tag**: tagging 대상 신규 코드 없음 — frontmatter 메타데이터 변경만.
- **Worktree 미사용**: per `feedback_worktree_never_use` 정책. plan-in-main + run-in-main + sync-in-main.

---

## 6. Stakeholders

- **Author**: manager-spec (이 SPEC의 plan-phase 작성자, 본 세션 orchestrator 위임)
- **Implementer (run-phase)**: manager-develop (per-SPEC analysis + frontmatter apply)
- **Validator**: plan-auditor (plan-phase audit, target PASS ≥0.85), `moai spec lint --strict` (final binary AC)
- **End user**: v3.0.0-rc1 release gatekeeper — 본 SPEC + CI-INFRA-FIX-001 모두 closure 후 tag 가능.

---

## 7. References

- **Direct trigger (LSGF-001)**: `.moai/specs/SPEC-V3R4-LINT-SPECID-GREP-FIX-001/` — PR #947 plan + #948 run + #949 sync (main `139c4d9d0`).
- **Precedent (SDF-001)**: `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001/` — Pattern A~H 처리 (5 Wave) 모델.
- **SSOT (12-field schema)**: `.claude/rules/moai/development/spec-frontmatter-schema.md` (SPEC-V3R4-SPECLINT-DEBT-002 산출).
- **walker behavior**: `internal/spec/drift.go:96-197` (LSGF-001 post-fix, word-boundary 매칭), `internal/spec/transitions.go:15-117` (ClassifyPRTitle + ExtractSPECIDs).
- **chore-skip filter**: `internal/spec/drift.go::shouldSkipCommitTitle` (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 산출, PR #933).
- **Lint AC evidence**: `moai spec lint --strict` 실행 결과 (main `139c4d9d0`) — 정확히 17 WARNING, 모두 `StatusGitConsistency` rule.
- **BODP**: CLAUDE.local.md §18.12 — orchestrator pre-evaluated `¬a ¬b ¬c` → base `origin/main` fresh branch.
- **Worktree 정책**: `~/.claude/projects/{hash}/memory/feedback_worktree_never_use.md` — plan/run/sync 모두 main checkout 에서 진행.

---

## 8. AC ↔ REQ Mapping (summary)

| AC ID              | REQ ID(s)                                   | Verification |
|--------------------|---------------------------------------------|--------------|
| AC-SDF002-A-1..A-5 | REQ-SDF002-002                              | `moai spec lint --strict 2>&1 \| grep <SPEC-A-N>` 결과 없음 (Category A 5건) |
| AC-SDF002-B-1..B-12| REQ-SDF002-002, REQ-SDF002-003              | `moai spec lint --strict 2>&1 \| grep <SPEC-B-N>` 결과 없음 (Category B 12건 각각) |
| AC-SDF002-X-001    | REQ-SDF002-001, REQ-SDF002-004              | `moai spec lint --strict 2>&1 \| tail -1` = `0 error(s), 0 warning(s)` |
| AC-SDF002-X-002    | REQ-SDF002-005                              | HARNESS-001/002/003 lint clean (LSGF-001 회귀 부재 확인) |
| AC-SDF002-X-003    | REQ-SDF002-006                              | `git diff main --stat` 변경 파일 모두 `.moai/specs/*/spec.md` 만 |

상세 시나리오는 `acceptance.md` 참조.
