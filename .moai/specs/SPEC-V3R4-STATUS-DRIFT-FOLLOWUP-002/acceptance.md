# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — Acceptance Criteria

> **Binary AC discipline**: 모든 AC는 PASS/FAIL 이분 결정. 부분 충족, partial credit, soft threshold 없음.
> **Verification environment**: main checkout, post-merge `.moai/specs/*/spec.md` 상태 + `moai spec lint --strict` (cwd = repo root).

## 1. Per-SPEC Acceptance Criteria

각 SPEC frontmatter 수정 후, 그 SPEC만 grep으로 격리하여 lint 결과 부재를 확인.

### 1.1 Category A (Forward sync-up, 5건)

#### AC-SDF002-A-1: SPEC-GLM-MCP-001 status sync-up

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-GLM-MCP-001"`
**Then**: 출력이 비어 있어야 한다 (정확히 0 라인)
**Mechanism applied**: frontmatter `status: in-progress → completed`, `updated: 2026-05-16`

#### AC-SDF002-A-2: SPEC-STATUSLINE-001 status sync-up

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-STATUSLINE-001"`
**Then**: 출력이 비어 있어야 한다
**Mechanism applied**: frontmatter `status: in-progress → implemented`, `updated: 2026-05-16`

#### AC-SDF002-A-3: SPEC-V3R2-WF-002 status sync-up

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-WF-002"`
**Then**: 출력이 비어 있어야 한다
**Mechanism applied**: frontmatter `status: in-progress → implemented`, `updated: 2026-05-16`

#### AC-SDF002-A-4: SPEC-V3R4-CATALOG-001 status sync-up

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R4-CATALOG-001"`
**Then**: 출력이 비어 있어야 한다
**Mechanism applied**: frontmatter `status: implemented → completed`, `updated: 2026-05-16`

#### AC-SDF002-A-5: SPEC-WORKTREE-002 status sync-up

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-WORKTREE-002"`
**Then**: 출력이 비어 있어야 한다
**Mechanism applied**: frontmatter `status: implemented → completed`, `updated: 2026-05-16`

### 1.2 Category B (Per-SPEC mechanism, 12건)

각 B-AC는 Wave 1 결정 후 채택된 mechanism (sync-commit / lint.skip / frontmatter downgrade) 으로 해소.

#### AC-SDF002-B-1: SPEC-V3R2-ORC-003 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-ORC-003"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `lint.skip + reason "synced via bulk-closure PR #926"` (run-phase 검증 후 확정)

#### AC-SDF002-B-2: SPEC-V3R2-RT-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-RT-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: frontmatter downgrade `implemented → planned` (외부 증거 부재 시) OR sync-commit (증거 있을 시) — run-phase 확정

#### AC-SDF002-B-3: SPEC-V3R2-RT-007 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-RT-007"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R2-RT-007 — status closeout under FOLLOWUP-002` commit 추가 OR `lint.skip + reason "synced via docs(sync) PR #856"`

#### AC-SDF002-B-4: SPEC-V3R2-SPC-002 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-SPC-002"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `lint.skip + reason "implementation tracked via T-SPC002-* subtask commits (walker word-boundary miss)"`

#### AC-SDF002-B-5: SPEC-V3R2-SPC-003 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-SPC-003"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R2-SPC-003 — implementation closeout` commit 추가 OR frontmatter downgrade `implemented → planned`

#### AC-SDF002-B-6: SPEC-V3R2-WF-003 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R2-WF-003"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R2-WF-003 — status closeout under FOLLOWUP-002` commit 추가

#### AC-SDF002-B-7: SPEC-V3R3-CI-AUTONOMY-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R3-CI-AUTONOMY-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `lint.skip + reason "synced via #927 bulk-closure; walker reads later fix(bodp) commit"`

#### AC-SDF002-B-8: SPEC-V3R3-HARNESS-LEARNING-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R3-HARNESS-LEARNING-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `lint.skip + reason "sync rolled up under HARNESS-001 closeout; PR #812 docs(spec) prefix not recognized by walker"`

#### AC-SDF002-B-9: SPEC-V3R3-PROJECT-HARNESS-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R3-PROJECT-HARNESS-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R3-PROJECT-HARNESS-001 — status closeout` commit 추가 OR `lint.skip + reason`

#### AC-SDF002-B-10: SPEC-V3R3-RETIRED-AGENT-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R3-RETIRED-AGENT-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R3-RETIRED-AGENT-001 — status closeout` commit 추가

#### AC-SDF002-B-11: SPEC-V3R3-RETIRED-DDD-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R3-RETIRED-DDD-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `sync(spec): SPEC-V3R3-RETIRED-DDD-001 — status closeout` commit 추가

#### AC-SDF002-B-12: SPEC-V3R4-LINT-SKIP-CLEANUP-001 remediation

**Given**: main HEAD에 본 SPEC run/sync PR 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep "SPEC-V3R4-LINT-SKIP-CLEANUP-001"`
**Then**: 출력이 비어 있어야 한다
**Plan-phase 가설 mechanism**: `lint.skip + reason "implementation tracked via chore(spec) PR #937 (walker chore-skip); sync rolled up under SDF-001 PR #939/#940"`

---

## 2. Cross-cutting Acceptance Criteria

### AC-SDF002-X-001: Binary lint clean (전체 목표)

**Given**: main HEAD에 본 SPEC run + sync PR 모두 머지 완료
**When**: `moai spec lint --strict 2>&1 | tail -1`
**Then**: 출력이 정확히 다음 문자열이어야 한다
```
0 error(s), 0 warning(s)
```
**Verification**: 단일 lint 실행 결과의 마지막 줄 string 비교. PASS/FAIL 이분.

### AC-SDF002-X-002: HARNESS-001/002/003 LSGF-001 회귀 부재

**Given**: main HEAD에 본 SPEC run + sync PR 모두 머지 완료
**When**: `moai spec lint --strict 2>&1 | grep -E "SPEC-V3R4-HARNESS-00[123]\b" | wc -l`
**Then**: 출력이 정확히 `0` 이어야 한다
**Rationale**: REQ-SDF002-005 — 본 SPEC이 LSGF-001 walker word-boundary 정밀도를 회귀시키지 않았음을 확인.

### AC-SDF002-X-003: Scope discipline (Go 코드 무수정)

**Given**: 본 SPEC run-PR 의 file diff
**When**: `git diff main...HEAD --name-only | grep -vE "^\.moai/specs/.+/(spec|plan|acceptance|design|tasks|progress)\.md$" | grep -v "^CHANGELOG\.md$" | wc -l`
**Then**: 출력이 정확히 `0` 이어야 한다 (모든 변경 파일이 `.moai/specs/*/[spec|plan|acceptance|design|tasks|progress].md` 또는 `CHANGELOG.md` 패턴)
**Rationale**: REQ-SDF002-006 — Go source, template, agent/skill, CI workflow, internal/* 모두 무수정 확인.

---

## 3. Definition of Done

본 SPEC은 다음 모든 조건이 만족될 때 closure 가능:

- [ ] AC-SDF002-A-1..A-5 (5건) 모두 PASS
- [ ] AC-SDF002-B-1..B-12 (12건) 모두 PASS
- [ ] AC-SDF002-X-001 (lint 0 warning) PASS
- [ ] AC-SDF002-X-002 (LSGF-001 회귀 부재) PASS
- [ ] AC-SDF002-X-003 (scope discipline) PASS
- [ ] plan-PR (현 phase) `--squash` auto-merge 등록 + plan-auditor PASS ≥ 0.85
- [ ] run-PR `--squash` auto-merge 등록 + CI all-green
- [ ] sync-PR `--squash` auto-merge 등록 + CI all-green
- [ ] spec.md frontmatter `status: draft → completed`, `version: "0.1.0" → "0.2.0"`
- [ ] CHANGELOG.md Unreleased 섹션에 FOLLOWUP-002 entry 추가 (ko + en)

---

## 4. Quality Gate Criteria

본 SPEC은 metadata-only 이므로 표준 TRUST-5 quality gate 중 다음 항목만 적용:

| TRUST 5 Pillar | 적용 여부 | Verification |
|----------------|-----------|--------------|
| Tested         | N/A       | 코드 변경 없음 |
| Readable       | 적용       | frontmatter YAML 구문 valid + HISTORY entry 가독성 |
| Unified        | 적용       | 12-field canonical schema 준수 |
| Secured        | N/A       | 민감 정보 없음 |
| Trackable      | 적용       | HISTORY entry + commit message + PR description 모두 SPEC-ID 명시 |

추가 gate:
- **plan-auditor**: target PASS ≥ 0.85 (plan-PR open 후 자동 실행)
- **moai spec lint --strict**: 0 ERROR / 0 WARNING (AC-SDF002-X-001 동치)

---

## 5. Edge Cases

### Edge-1: Wave 1 에서 가설 검증 실패

**Scenario**: plan-phase Category B 가설 mechanism 이 run-phase 실측에서 부적절 판명.

**Expected behavior**:
- plan.md §3 Category B Analysis Table 의 해당 row inline 갱신
- mechanism 변경 + 사유 기록
- AC-SDF002-B-N 충족 여부 final mechanism 기준으로 평가

**Out of scope**: plan-phase 가설 100% 정확도 — plan-auditor는 evidence-driven process 를 평가하므로 가설 정확도 자체는 비평가.

### Edge-2: 새 WARNING 노출

**Scenario**: 17건 처리 후 lint 가 18번째 새 WARNING 노출.

**Expected behavior**:
- 본 SPEC scope 확장 금지 (per REQ-SDF002-006)
- 별도 follow-up SPEC `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-003` 발급
- 본 SPEC AC-SDF002-X-001은 17건 한정으로 평가 (18번째는 follow-up scope)

**Mitigation**: PR description에 lint 전체 결과 첨부하여 18번째 노출 시 즉시 발견.

### Edge-3: sync-PR 자체가 또 다른 status drift 유발

**Scenario**: 본 SPEC sync-PR commit `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002` 머지 후 본 SPEC 자체에 대해 walker 가 inconsistency 보고.

**Expected behavior**:
- sync commit subject 가 자기 SPEC ID 만 명시하면 walker는 `("sync-merge", "completed")` 인식
- 본 SPEC frontmatter `status: completed` 와 일치 → 회귀 없음

**Mitigation**: sync commit subject 에 다른 SPEC ID 본문 미포함 (Risk-5 mitigation 동일).

### Edge-4: HARNESS-001/002/003 가 LSGF-001 회귀로 다시 WARNING 발생

**Scenario**: 본 SPEC Wave 2 적용이 walker behavior 변경 없음에도 불구하고 HARNESS-001 lint 결과가 달라짐.

**Expected behavior**:
- 본 SPEC Wave 2 는 frontmatter + lint.skip 만 변경 — walker code 변경 0
- HARNESS-001/002/003 frontmatter 본 SPEC 미수정 (17건 inventory 미포함)
- AC-SDF002-X-002 회귀 부재 확인 통과 보장

**Mitigation**: Wave 3 verification 에서 explicit grep 으로 회귀 부재 확인.

---

## 6. AC ↔ REQ Mapping (verification matrix)

| AC ID                | REQ ID(s)                              | Test Command (verification)                                                                       |
|----------------------|----------------------------------------|---------------------------------------------------------------------------------------------------|
| AC-SDF002-A-1        | REQ-SDF002-002                         | `moai spec lint --strict 2>&1 \| grep "SPEC-GLM-MCP-001"` → empty                                  |
| AC-SDF002-A-2        | REQ-SDF002-002                         | `moai spec lint --strict 2>&1 \| grep "SPEC-STATUSLINE-001"` → empty                              |
| AC-SDF002-A-3        | REQ-SDF002-002                         | `moai spec lint --strict 2>&1 \| grep "SPEC-V3R2-WF-002"` → empty                                  |
| AC-SDF002-A-4        | REQ-SDF002-002                         | `moai spec lint --strict 2>&1 \| grep "SPEC-V3R4-CATALOG-001"` → empty                            |
| AC-SDF002-A-5        | REQ-SDF002-002                         | `moai spec lint --strict 2>&1 \| grep "SPEC-WORKTREE-002"` → empty                                |
| AC-SDF002-B-1..B-12  | REQ-SDF002-002, REQ-SDF002-003         | `moai spec lint --strict 2>&1 \| grep "<Category-B-SPEC-ID>"` → empty (per AC)                    |
| AC-SDF002-X-001      | REQ-SDF002-001, REQ-SDF002-004         | `moai spec lint --strict 2>&1 \| tail -1` = "0 error(s), 0 warning(s)"                            |
| AC-SDF002-X-002      | REQ-SDF002-005                         | `moai spec lint --strict 2>&1 \| grep -cE "SPEC-V3R4-HARNESS-00[123]\\b"` = 0                      |
| AC-SDF002-X-003      | REQ-SDF002-006                         | `git diff main...HEAD --name-only \| grep -vE "<allowed pattern>" \| wc -l` = 0                   |
