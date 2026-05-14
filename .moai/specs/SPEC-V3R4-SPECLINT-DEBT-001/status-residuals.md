# Status Residuals — SPEC-V3R4-SPECLINT-DEBT-001 Wave 2 T-SLD-007 산출물

## HISTORY

| Version | Date       | Author              | Description |
|---------|------------|---------------------|-------------|
| 0.1.0   | 2026-05-15 | manager-develop run | T-SLD-007 batch status sync 완료 후 잔존 51 StatusGitConsistency WARNING 의 ambiguous 케이스 enumeration. 모두 author intent 보존을 위해 의도적으로 보존됨. AC-SLD-007 target 재조정 결정 (사용자 AskUserQuestion 1 응답 → Option A) 의 근거 문서. |

---

## 1. 요약

`moai spec lint --strict` 실행 결과 잔존 StatusGitConsistency WARNING **51건**. 분포:

| Pair (frontmatter → git-implied) | Count | 보존 이유                                                                                                |
|----------------------------------|-------|---------------------------------------------------------------------------------------------------------|
| `completed` → `implemented`      | 47    | SPEC author 가 sync 단계까지 완료했지만 lint git 추론이 latest commit 의 `feat:` prefix 만 인식. author intent 보존. |
| `superseded` → `completed`       | 2     | terminal state. author 가 명시적으로 supersede 선언 (V3R3 → V3R4 마이그레이션 등). 보존.                |
| `archived` → `implemented`       | 1     | terminal state. archived 마커는 변경 금지.                                                              |
| `archived` → `in-progress`       | 1     | terminal state. 동일.                                                                                   |

---

## 2. 보존 사유 상세

### 2.1 `completed` → `implemented` (47 cases)

이 47개 SPEC 들의 frontmatter `status: completed` 는 다음 lifecycle 을 명시한다:
- Run PR 머지 완료 → 코드가 main 에 존재
- Sync PR 머지 완료 → 문서/codemap/MX tags 동기화 완료
- 전체 lifecycle 종료 → author 가 SPEC 닫음을 선언

반면 lint 의 `getGitImpliedStatus` 는 `git log main --grep=<SPEC-ID> -1` 으로 latest commit 1건만 보고 분류한다. 47 SPEC 들의 latest commit 은 다음 시나리오 중 하나로 `feat:` prefix 를 가진다:

- (a) Follow-up bugfix 가 SPEC 을 mention (예: `fix: ... related to SPEC-XXX`)
- (b) Batch SPEC update commit 이 mention (예: `feat(specs): batch frontmatter sync` mentioning 50+ SPECs)
- (c) Cross-SPEC refactor mention

따라서 lint 가 `implemented` 로 추론해도 SPEC 의 실제 lifecycle 은 `completed` 이다. 47 cases 모두 author intent (`completed`) 가 정답이며, git 추론이 latest-1 휴리스틱 한계로 over-conservative.

#### 47 SPECs (sorted)

```
SPEC-AGENCY-ABSORB-001, SPEC-AGENT-002, SPEC-CC2122-HOOK-001, SPEC-CC2122-STATUSLINE-001,
SPEC-CC297-001, SPEC-CICD-001, SPEC-DB-SYNC-HARDEN-001, SPEC-DB-SYNC-RELOC-001, SPEC-DESIGN-001,
SPEC-DOCS-SB-REMOVE-001, SPEC-GLM-001, SPEC-HOOK-008, SPEC-HOOK-009, SPEC-KARPATHY-001,
SPEC-PSR-001, SPEC-QUALITY-001, SPEC-REFACTOR-001, SPEC-SKILL-002, SPEC-SKILL-GATE-001,
SPEC-SLE-001, SPEC-SLV3-001, SPEC-SRS-001, SPEC-SRS-002, SPEC-SRS-003, SPEC-STATUS-AUTO-001,
SPEC-STATUSLINE-002, SPEC-TEAM-001, SPEC-UI-003, SPEC-UPDATE-002, SPEC-UTIL-003,
SPEC-V3R2-ORC-001, SPEC-V3R2-ORC-005, SPEC-V3R2-RT-004, SPEC-V3R2-RT-005, SPEC-V3R2-SPC-004,
SPEC-V3R2-WF-005, SPEC-V3R2-WF-006, SPEC-V3R3-ARCH-007, SPEC-V3R3-BRAIN-001,
SPEC-V3R3-CMD-CLEANUP-001, SPEC-V3R3-COV-001, SPEC-V3R3-DEF-001, SPEC-V3R3-DEF-007,
SPEC-V3R3-DESIGN-PIPELINE-001, SPEC-V3R4-STATUS-LIFECYCLE-001, SPEC-WF-AUDIT-GATE-001,
SPEC-WORKTREE-002
```

### 2.2 `superseded` → `completed` (2 cases)

Terminal state. SPEC author 가 명시적으로 supersede 결정. `completed` 로 다운그레이드 시 후속 SPEC supersede 관계 정보 손실. 보존.

- **SPEC-LSP-001**: superseded by 후속 LSP SPEC.
- **SPEC-V3R3-HARNESS-001**: superseded by SPEC-V3R4-HARNESS-001 (BC-V3R4-HARNESS-001-CLI-RETIREMENT).

### 2.3 `archived` (2 cases)

Terminal state. 외부 archive (moai-docs 등) 또는 deprecated SPEC. 변경 불가.

- **SPEC-I18N-001-ARCHIVED** (id field 는 본 SPEC 에서 `SPEC-I18N-001` 로 정규화됨): moai-docs 에서 이관된 archive.
- **SPEC-V3R3-WEB-001**: deprecated web SPEC.

---

## 3. False-positive 분류

본 51 cases 는 모두 lint rule 자체의 한계 (latest-1 git commit 휴리스틱) 에서 비롯된 false-positive 이다. 다음 방법으로 lint side 에서 향후 개선 가능:

- (a) **Lint 개선**: `getGitImpliedStatus` 에서 `status(XXX)` 또는 `sync` prefix 우선 탐색 후 fallback. 그러나 본 SPEC 범위 밖 (lint rule 수정 = §1.3 Non-Goals).
- (b) **`status(completed)` batch commit**: PR squash merge 특성상 개별 commit prefix 가 squash title 로 collapse 되므로 효과 없음. Merge commit 전략 (§18.3 위반) 만 효과 있음. 본 SPEC 에서는 시도 안 함.
- (c) **Lint exemption mechanism**: `frontmatter.lint.skip: [StatusGitConsistency]` 추가 (REQ-SPC-003-040 에 정의됨). 그러나 47 SPEC 의 frontmatter 일괄 수정은 음의 ROI.

본 SPEC 에서는 **(d) AC-SLD-007 target 재조정 + 문서화** 채택. 향후 follow-up SPEC 에서 (a) lint rule 개선 가능.

---

## 4. AC-SLD-007 Target Revision

원래 `acceptance.md` AC-SLD-007:

> Then stdout 의 WARNING 카운트는 **5 이하** 이어야 한다.

본 SPEC run-phase 결정 (사용자 AskUserQuestion 1 응답) 으로 다음과 같이 재조정:

> Then stdout 의 WARNING 카운트는 **55 이하** 이어야 한다 (잔존 51 ambiguous 보존 + 임의의 사소한 신규 drift 4 여유분).

---

## 5. Follow-up Recommendations

본 SPEC 완료 후 다음 follow-up 가능 (별도 SPEC 으로 관리):

- **F1**: `internal/spec/lint.go::getGitImpliedStatus` 개선 — `sync(` / `docs(sync)` / `status(completed)` prefix 우선 탐색 후 fallback.
- **F2**: 47 cases 중 일부 SPEC 의 supersede 관계를 frontmatter 에 추가 (`supersedes` / `superseded_by` 필드 활용) 하여 자동 archive 식별 가능 환경 구축.
- **F3**: Status lifecycle 의 `completed` 단계 정의를 lint rule 에서도 인식하도록 `git log --grep=sync(` AND `git log --grep=feat(` 동시 조회 후 둘 다 머지된 케이스에만 `completed` 추론.

위 follow-up 은 본 SPEC 범위 밖이며, V3R5 또는 V4 timeline 의 lint enhancement SPEC 에서 처리.

---

## 6. References

- `acceptance.md` AC-SLD-007 (revised target ≤ 55)
- `spec.md` §5 Success Criteria (revised WARNING budget)
- `plan.md` §1.2 §4 StatusGitConsistency 해소 전략
- `internal/spec/lint.go:879 StatusGitConsistencyRule`
- `internal/spec/drift.go:99 getGitImpliedStatus`
- `internal/spec/transitions.go:70 ClassifyPRTitle`
