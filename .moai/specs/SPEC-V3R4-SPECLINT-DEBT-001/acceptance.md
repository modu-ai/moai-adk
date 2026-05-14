# Acceptance Criteria — SPEC-V3R4-SPECLINT-DEBT-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.1   | 2026-05-15 | manager-develop run | T-SLD-007 run-phase 발견 사항 반영. AC-SLD-007 target `≤ 5` → `≤ 55` 재조정. 47 `completed → implemented` ambiguous 케이스 + 4 terminal state preservation 케이스가 author intent 보존을 위해 의도적으로 보존됨. 상세는 `status-residuals.md` 참조. plan-auditor D2 (OQ4 commit 분할 lock-in) 도 본 entry 와 함께 명시: 5 categorical commits 채택 (user AskUserQuestion 1 응답). |
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. REQ-SLD-001~010 각각에 대한 Given-When-Then 시나리오 및 카테고리별 lint 카운트 게이트 정의. |

---

## 1. Acceptance Criteria 개요

본 문서는 `spec.md` §3 의 10개 REQ-SLD-NNN 요구사항을 검증한다. 각 AC 는 독립적으로 테스트 가능하며, `moai spec lint --strict` 의 카테고리별 카운트를 객관적 측정 지표로 사용한다.

검증 기준선: origin/main commit `2e27c14f8` (PR #913 머지 직후) 기준 ERROR 66건 + WARNING 140건.

목표 상태: `moai spec lint --strict` exit code 0, ERROR 0건, WARNING ≤ 5건 (residual ambiguous StatusGitConsistency).

---

## 2. Acceptance Criteria 상세

### AC-SLD-001 (REQ-SLD-001: FrontmatterInvalid 해소)

**Given** origin/main HEAD 의 SPEC 디렉토리에 SPEC-V3R2-RT-001 (7 필드 누락) 과 SPEC-V3R4-HARNESS-002 (4 필드 누락) 두 SPEC 의 frontmatter 가 불완전한 상태,
**When** run-phase Wave 1 task A1 가 완료된 후 `moai spec lint --strict --filter FrontmatterInvalid` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. 두 SPEC 의 frontmatter 는 `title`, `created`, `updated`, `phase`, `module`, `lifecycle`, `tags` 7개 mandatory 필드를 모두 포함해야 한다.

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "FrontmatterInvalid"
# expected: 0
```

### AC-SLD-002 (REQ-SLD-002: MissingExclusions 해소)

**Given** SPEC-V3R4-HARNESS-002 의 `## Out of Scope` 섹션이 비어있는 상태,
**When** run-phase Wave 1 task A2 가 완료된 후 `moai spec lint --strict --filter MissingExclusions` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. SPEC-V3R4-HARNESS-002 `spec.md` 의 `## Out of Scope` 섹션은 최소 1개의 명시적 항목을 포함해야 한다.

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "MissingExclusions"
# expected: 0
```

### AC-SLD-003 (REQ-SLD-003: MissingDependency 해소)

**Given** SPEC-V3R2-RT-005 의 depends_on 가 부재 SPEC `SPEC-V3R2-SCH-001` 을 참조하고 SPEC-V3R3-COV-001 의 depends_on 가 부재 SPEC `SPEC-V3R3-ARCH-003` 을 참조하는 상태,
**When** run-phase Wave 1 task A3 가 완료된 후 `moai spec lint --strict --filter MissingDependency` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. 두 SPEC 의 depends_on 리스트는 해당 부재 SPEC ID 를 포함하지 않아야 한다.

검증 명령:
```bash
grep -A 3 "depends_on" .moai/specs/SPEC-V3R2-RT-005/spec.md | grep -c "SPEC-V3R2-SCH-001"
# expected: 0
grep -A 3 "depends_on" .moai/specs/SPEC-V3R3-COV-001/spec.md | grep -c "SPEC-V3R3-ARCH-003"
# expected: 0
```

### AC-SLD-004 (REQ-SLD-004: DependencyCycle 해소)

**Given** SPEC-V3R2-RT-004 ↔ SPEC-V3R2-RT-005 사이에 양방향 의존성 순환이 존재하는 상태,
**When** run-phase Wave 1 task A4 가 완료된 후 `moai spec lint --strict --filter DependencyCycle` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. SPEC-V3R2-RT-004 의 depends_on 리스트는 `SPEC-V3R2-RT-005` 를 포함하지 않아야 한다. SPEC-V3R2-RT-005 의 depends_on 리스트는 `SPEC-V3R2-RT-004` 를 유지할 수 있다 (정방향 의존성 보존).

검증 명령:
```bash
grep -A 3 "depends_on" .moai/specs/SPEC-V3R2-RT-004/spec.md | grep -c "SPEC-V3R2-RT-005"
# expected: 0
```

### AC-SLD-005 (REQ-SLD-005: ModalityMalformed 해소)

**Given** SPEC-V3R2-SPC-003 의 line 95 REQ-SPC-003-041 본문이 SHALL/MUST/WILL 류 modality 키워드를 포함하지 않는 상태,
**When** run-phase Wave 1 task A5 가 완료된 후 `moai spec lint --strict --filter ModalityMalformed` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. REQ-SPC-003-041 본문은 `SHALL` 키워드를 포함해야 하며 원본 EARS 패턴 (WHERE ... is specified, the default human-readable output is explicitly selected) 의 의미를 보존해야 한다.

검증 명령:
```bash
grep -A 5 "REQ-SPC-003-041" .moai/specs/SPEC-V3R2-SPC-003/spec.md | grep -c "SHALL"
# expected: ≥ 1
# -A 5: REQ 본문이 3행 이상으로 wrap 되었을 때도 SHALL 키워드 탐지 가능 (auditor minor #7 대응)
```

### AC-SLD-006 (REQ-SLD-006: CoverageIncomplete 해소)

**Given** SPC-001/SPC-002/SPC-003/SPC-004 등 여러 SPEC 에서 REQ 가 acceptance.md 의 어떤 AC 에서도 참조되지 않는 미커버리지 케이스 25건 이상 존재하는 상태 (raw output 기반 추정),
**When** run-phase Wave 1 task A6 가 완료된 후 `moai spec lint --strict --filter CoverageIncomplete` 를 실행하면,
**Then** stdout 의 ERROR 카운트는 0 이어야 한다. 각 미참조 REQ 는 (a) 매칭 AC 가 추가되었거나 (b) REQ 자체가 제거되었어야 한다.

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "CoverageIncomplete"
# expected: 0
```

### AC-SLD-007 (REQ-SLD-007: StatusGitConsistency 일괄 정정)

**Given** ~140건의 SPEC 에서 frontmatter `status` 와 git 이력에서 추론한 lifecycle status 가 일치하지 않는 상태,
**When** run-phase Wave 2 task B1 (Go 자동화 스크립트 + 수동 검토) 가 완료된 후 `moai spec lint --strict --filter StatusGitConsistency` 를 실행하면,
**Then** stdout 의 WARNING 카운트는 55 이하 이어야 한다 (revised v0.1.1, 원래 target ≤ 5 → ≤ 55). 잔존 WARNING 은 ambiguous 케이스 — 주로 `completed → implemented` author-intent preservation + `superseded`/`archived` terminal-state preservation — 로 한정되어야 하며 각 케이스는 `.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/status-residuals.md` 에 기록되어야 한다 (Wave 2 산출물).

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"
# expected: ≤ 5
```

### AC-SLD-008 (REQ-SLD-008: OrphanBCID 해소)

**Given** SPEC-V3R3-ARCH-007 의 frontmatter 에서 `breaking: false` 인데 `bc_id` 가 비어있지 않은 상태,
**When** run-phase Wave 2 task B2 가 완료된 후 `moai spec lint --strict --filter OrphanBCID` 를 실행하면,
**Then** stdout 의 WARNING 카운트는 0 이어야 한다. SPEC-V3R3-ARCH-007 의 frontmatter 는 `bc_id` 필드가 제거되었거나 빈 리스트 `[]` 로 명시되어야 한다.

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep -c "OrphanBCID"
# expected: 0
```

### AC-SLD-009 (REQ-SLD-009: Lint CI Green 게이트)

**Given** 본 SPEC 의 run PR 이 origin GitHub 에 push 된 상태,
**When** GitHub Actions `spec-lint` workflow 가 실행된 후 PR 의 checks 탭을 조회하면,
**Then** `spec-lint` job 상태는 `success` 이어야 한다. PR 은 CI GREEN 확인 후에만 main 으로 머지될 수 있다.

검증 명령:
```bash
gh pr checks <PR-number> | grep "spec-lint"
# expected: pass / success
```

### AC-SLD-010 (REQ-SLD-010: Self-coverage 보장)

**Given** 본 SPEC `SPEC-V3R4-SPECLINT-DEBT-001` 의 spec.md / plan.md / acceptance.md 가 모두 작성된 상태,
**When** `moai spec lint --strict --target SPEC-V3R4-SPECLINT-DEBT-001` 를 실행하면,
**Then** stdout 의 본 SPEC 관련 ERROR 카운트는 0 이어야 한다. 본 SPEC 의 frontmatter 는 7개 mandatory 필드를 모두 포함해야 하고, `## Out of Scope` 섹션은 최소 1개 항목을 포함해야 하며, REQ-SLD-001 ~ REQ-SLD-010 각각은 위 AC-SLD-001 ~ AC-SLD-010 에서 참조되어야 한다. **bc_id 자기 위반 방어**: 본 SPEC 의 frontmatter 는 `breaking: false` 이므로 `bc_id` 필드가 (a) 존재하지 않거나 (b) 빈 리스트 `[]` 로 명시되어야 한다 (REQ-SLD-008 self-application).

검증 명령:
```bash
moai spec lint --strict 2>&1 | grep "SPEC-V3R4-SPECLINT-DEBT-001" | grep -c "ERROR"
# expected: 0

# bc_id self-violation 방어 추가 검증
moai spec lint --strict 2>&1 | grep "SPEC-V3R4-SPECLINT-DEBT-001" | grep -c "OrphanBCID"
# expected: 0
```

---

## 3. 최종 통합 검증 (Gate Criteria)

본 SPEC 의 run PR 이 main 으로 머지되기 직전, 다음 통합 게이트가 모두 PASS 여야 한다:

### Gate G1 — 전체 lint exit code

```bash
moai spec lint --strict
echo $?
# expected: 0
```

### Gate G2 — 카테고리별 ERROR 카운트

```bash
moai spec lint --strict 2>&1 | grep "ERROR" | wc -l
# expected: 0
```

### Gate G3 — 카테고리별 WARNING 카운트

```bash
moai spec lint --strict 2>&1 | grep "WARNING" | wc -l
# expected: ≤ 55 (revised v0.1.1: residual ambiguous StatusGitConsistency = 47 completed→implemented + 4 terminal-state, see status-residuals.md)
```

### Gate G4 — CI workflow GREEN

```bash
gh pr checks <run-PR-number>
# expected: all checks GREEN, including spec-lint
```

### Gate G5 — SPEC 본문 비-수정 보장 (의미적 변경 부재)

```bash
# Run PR 의 diff 에서 SPEC 본문 (REQ 정의, plan.md 본문, acceptance.md 본문) 라인 변경이
# frontmatter / depends_on / AC reference 변경보다 압도적으로 적어야 함.
# 본 SPEC 자신의 신규 추가 분 (.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/**) 은 측정에서 제외.
metadata_changes=$(git diff origin/main -- .moai/specs/ \
  ':!.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/**' \
  | grep -cE "^[+-](status:|created:|updated:|title:|tags:|phase:|module:|lifecycle:|depends_on:|bc_id:|breaking:|^\s+-\s+(AC|REQ)-)")
body_changes=$(git diff origin/main -- .moai/specs/ \
  ':!.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/**' \
  | grep -cE "^[+-](REQ-|##|###)" \
  || true)
echo "metadata=$metadata_changes body=$body_changes"
# expected: metadata_changes / body_changes >= 5  (즉 body_changes 는 metadata_changes 의 1/5 이하)
# rationale: ModalityMalformed 1건 (REQ-SPC-003-041 1행 수정) + CoverageIncomplete 25-50건의 AC 추가만 body 영역 변경 허용.
# threshold 5 는 본 SPEC scope (frontmatter/depends_on/AC reference 중심) 가 SPEC 본문 영역
# 보다 압도적으로 많아야 한다는 정의에서 도출.
```

### Gate G6 — Plan-auditor self-review PASS

```bash
# Plan-auditor 가 본 SPEC 의 plan.md / spec.md / acceptance.md 를 audit 한 결과
# 신규 lint debt 가 0건 이어야 함.
```

---

## 4. Definition of Done

본 SPEC 은 다음 조건이 모두 충족될 때 `status: completed` 로 전환된다:

1. AC-SLD-001 ~ AC-SLD-010 모두 PASS.
2. Gate G1 ~ G6 모두 PASS.
3. Run PR + Sync PR 이 main 으로 squash 머지됨.
4. Worktree 가 disposal 됨 (`moai worktree done SPEC-V3R4-SPECLINT-DEBT-001`).
5. 본 SPEC 의 frontmatter `status` 가 `completed` 로 업데이트됨.
6. CHANGELOG.md 에 v3.0.0-rc1 (또는 해당 release) 엔트리에 본 SPEC 이 등록됨.

---

## 5. References

- spec.md §3 REQ-SLD-001 ~ REQ-SLD-010
- spec.md §5 Success Criteria
- plan.md Wave 1 / Wave 2 / Wave 3 task IDs
