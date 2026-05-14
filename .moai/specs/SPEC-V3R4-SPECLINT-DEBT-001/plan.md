# Plan — SPEC-V3R4-SPECLINT-DEBT-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.1   | 2026-05-15 | manager-develop run | OQ4 lock-in (user AskUserQuestion 1 응답): 5 categorical commits 채택 (Option B). plan-auditor D1 (tasks.md L28-30 time predictions) follow-up: §5.1 Option B 의 effort 분류 시간 추정 제거는 본 sync 단계에서 처리. ParseFailure 4건 + ID format 2건 등 expanded scope 결정 정리. |
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. 3 Wave / 11 task 분할. Wave 1 (P0 ERROR 해소, 6 task) → Wave 2 (P1 WARNING 해소, 2 task) → Wave 3 (검증 + 머지, 3 task). 각 task는 T-SLD-NNN 패턴으로 식별. @MX:NOTE 태그를 SPEC 분석 pivot 지점에 적용. |

---

## 1. Overview

@MX:NOTE — 본 SPEC 은 메타데이터 정정 전용 SPEC 이다. 일반적인 feature SPEC 과 달리 새 코드 산출물이 거의 없고, 기존 SPEC 188개 의 frontmatter / depends_on / AC reference 만 수정한다. 이로 인해 일반적인 milestone (M1/M2/M3) 분할 대신 lint 카테고리별 Wave 분할이 더 적합하다.

본 plan 은 3개 Wave 로 구성된다:

- **Wave 1**: P0 ERROR 6개 카테고리 해소 (FrontmatterInvalid → MissingDependency → DependencyCycle → ModalityMalformed → MissingExclusions → CoverageIncomplete 순). 복잡도 오름차순.
- **Wave 2**: P1 WARNING 2개 카테고리 해소 (StatusGitConsistency 자동화 + OrphanBCID 단일 케이스).
- **Wave 3**: 통합 검증 + PR 생성 + main 머지.

총 11 task (Wave 1: 6 ERROR 해소 + Wave 2: 2 WARNING 해소 + Wave 3: 3 verification/PR). priority: 모두 P0 (CI-blocking 해소가 본 SPEC 목적).

---

## 2. Wave 1 — P0 ERROR 해소

### T-SLD-001 — FrontmatterInvalid 7+4 필드 보충

**Target**: SPEC-V3R2-RT-001 (7 필드 누락), SPEC-V3R4-HARNESS-002 (4 필드 누락).

**Action**:
1. SPEC-V3R2-RT-001/spec.md 의 frontmatter 를 읽고 누락 필드 식별.
2. 인접 V3R2-RT-* SPEC (예: RT-002, RT-003) 의 frontmatter 컨벤션 참조.
3. git log `--follow` 로 SPEC-V3R2-RT-001 의 최초 commit 시각을 `created` 값으로 사용.
4. 최근 modify commit 시각을 `updated` 값으로 사용.
5. `phase` / `module` / `lifecycle` / `tags` 는 SPEC 본문 분석 + 인접 SPEC 비교로 추론.
6. SPEC-V3R4-HARNESS-002/spec.md 에 대해서도 동일 절차 (title, created, updated, tags 4건).

**Output**: 2개 SPEC frontmatter 완전화.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "FrontmatterInvalid"` → 0.

**Dependencies**: 없음.

**Tools**: Read, Edit, Bash (git log).

**Reference**: REQ-SLD-001, AC-SLD-001.

### T-SLD-002 — MissingDependency 2건 해소

**Target**: SPEC-V3R2-RT-005 (SCH-001 참조), SPEC-V3R3-COV-001 (ARCH-003 참조).

**Action**:
1. 두 SPEC 의 spec.md 본문을 읽고 SCH-001 / ARCH-003 가 sentinel 인지 별칭인지 확인.
2. spec.md §1.2 §1 결정대로 depends_on 목록에서 해당 항목 제거.
3. 만약 본문에서 명확히 "SCH-001 의 X 기능에 의존" 같은 문구가 발견되면, run-phase 결정을 (a) 제거 에서 (b) repoint 로 변경하고 사용자 확인.

**Output**: 2개 SPEC frontmatter depends_on 정정.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "MissingDependency"` → 0.

**Dependencies**: 없음.

**Tools**: Read, Edit, Grep.

**Reference**: REQ-SLD-003, AC-SLD-003.

### T-SLD-003 — DependencyCycle RT-004 ↔ RT-005 해소

**Target**: SPEC-V3R2-RT-004 의 depends_on.

**Action**:
1. SPEC-V3R2-RT-004 와 SPEC-V3R2-RT-005 의 spec.md 본문을 모두 읽고 의존성 방향 의도 확인.
2. RT cluster 의 sequential numbering convention 상 RT-005 → RT-004 (후행 → 선행) 방향이 자연스러우므로, RT-004 → RT-005 백엣지를 제거.
3. 만약 본문 분석 결과 반대 방향이 자연스럽다면 RT-005 → RT-004 를 제거하고 사용자 확인.

**Output**: SPEC-V3R2-RT-004 frontmatter depends_on 에서 RT-005 제거 (또는 그 반대).

**Verification**: `moai spec lint --strict 2>&1 | grep -c "DependencyCycle"` → 0.

**Dependencies**: T-SLD-002 무관 (별개 SPEC).

**Tools**: Read, Edit.

**Reference**: REQ-SLD-004, AC-SLD-004.

### T-SLD-004 — ModalityMalformed REQ-SPC-003-041 해소

**Target**: SPEC-V3R2-SPC-003 line 95 REQ-SPC-003-041.

**Action**:
1. SPEC-V3R2-SPC-003/spec.md line 95 를 Read 로 확인.
2. 원본 텍스트: "WHERE `moai spec lint --format table` is specified, the default human-readable output is explicitly selected (redundant with default, useful for script clarity)."
3. SHALL 키워드 삽입: "WHERE `moai spec lint --format table` is specified, the system **SHALL** explicitly select the default human-readable output (redundant with default, useful for script clarity)."
4. Edit tool 로 단일 라인 교체.

**Output**: SPEC-V3R2-SPC-003 line 95 EARS modality 정합.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "ModalityMalformed"` → 0.

**Dependencies**: 없음.

**Tools**: Read, Edit.

**Reference**: REQ-SLD-005, AC-SLD-005.

### T-SLD-005 — MissingExclusions HARNESS-002 해소

**Target**: SPEC-V3R4-HARNESS-002/spec.md `## Out of Scope` 섹션.

**Action**:
1. SPEC-V3R4-HARNESS-002/spec.md 본문 전체 Read.
2. SPEC 의 in-scope 와 대조하여 deferred / out-of-scope 항목 최소 1개 식별.
3. `## Out of Scope` 섹션에 명시적 bullet 추가.
4. (이미 §1.3 Non-Goals 가 존재할 수 있으므로 형식 확인 후 적절히 배치.)

**Output**: SPEC-V3R4-HARNESS-002 `## Out of Scope` 섹션에 최소 1 항목 포함.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "MissingExclusions"` → 0.

**Dependencies**: T-SLD-001 (HARNESS-002 frontmatter 정정 후 본문 검토가 수월).

**Tools**: Read, Edit.

**Reference**: REQ-SLD-002, AC-SLD-002.

### T-SLD-006 — CoverageIncomplete 일괄 분석 및 해소

**Target**: SPC-001, SPC-002, SPC-003, SPC-004 등 다수 SPEC.

**Action**:
1. `moai spec lint --strict` 를 다시 실행하여 CoverageIncomplete 전체 케이스 enumeration.
2. 각 미참조 REQ-XXX-NNN 마다:
   - spec.md 의 해당 REQ 본문 Read.
   - acceptance.md 의 모든 AC Read 후 REQ-XXX-NNN 가 어떤 AC 에 매칭되는지 분석.
   - **결정 (a)**: REQ 가 명확한 의도를 가지면, acceptance.md 에 매칭 AC 추가 (테스트 가능한 Given-When-Then).
   - **결정 (b)**: REQ 가 orphan placeholder 로 판단되면, spec.md 에서 REQ 삭제.
   - **결정 (c)**: REQ 가 기존 AC 에서 implicit 으로 검증되나 reference 가 누락된 경우, 기존 AC 에 `Reference: REQ-XXX-NNN` 라인 추가.
3. 각 SPEC 별로 batch 처리.

**Output**: 25+ 미참조 REQ 모두 해소.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "CoverageIncomplete"` → 0.

**Dependencies**: T-SLD-001 ~ T-SLD-005 무관 (별개 카테고리).

**Tools**: Read, Edit, Grep, Bash.

**Risk**: @MX:WARN — 실제 케이스 수가 raw output 25건이 아니라 50건+ 일 수 있음. Wave 1 초입에 정확한 수치 측정 필수. 만약 60건 이상이면 T-SLD-006 을 sub-task 로 분할 (T-SLD-006a / 006b / 006c per SPEC).

**Reason** (@MX:REASON): CoverageIncomplete 가 가장 case-specific 한 분석을 요구하므로 단일 task 로 추정한 effort 가 가장 부정확함.

**Reference**: REQ-SLD-006, AC-SLD-006.

---

## 3. Wave 2 — P1 WARNING 해소

### T-SLD-007 — StatusGitConsistency 자동화 + 잔존 수동 정리

**Target**: ~140건 SPEC 의 frontmatter `status` 필드.

**Action**:
1. **자동화 스크립트 작성** (`scripts/spec-status-sync.go`):
   - 모든 SPEC 디렉토리 순회.
   - 각 SPEC 의 git log 에서 PR 머지 commit 식별 (`gh pr list --search "head:<branch> is:merged"`).
   - 매 SPEC 마다 lifecycle 추론:
     - PR 0개 머지 → `draft`
     - Plan PR 머지만 → `planned` 또는 `in_review`
     - Run PR 머지 → `implemented`
     - Run + Sync PR 모두 머지 → `completed`
   - frontmatter `status` 와 추론값 비교, 불일치 시 정정.
2. **스크립트 실행 + 결과 검토**: `go run scripts/spec-status-sync.go --dry-run` 로 변경 사항 검토.
3. **수동 검토**: 자동화 결과 중 false-positive 후보 (예: PR 없이 commit, 다중 PR 경유, supersede 관계) 를 manual 검토. False-positive 임계치: 자동화 추론과 수동 결정이 다를 때 `--review-only` 출력으로 분리하고, 30% 이상 false-positive 발견 시 batch apply 중단 후 사용자에게 재설계 요청 (T-SLD-007.5 escalation).
4. **batch apply**: `--apply` 플래그로 frontmatter 일괄 정정. **카테고리별 commit 분리** (예: `status: draft → in_review` 1 commit, `status: implemented → completed` 1 commit) 로 rollback granularity 확보.
5. **Rollback 절차** (false-positive 발견 시):
   - 옵션 A: `git revert <commit>` — 단일 카테고리 commit revert (가장 안전, 권장).
   - 옵션 B: `go run scripts/spec-status-sync.go --rollback --from-snapshot <snapshot.json>` — 스크립트가 dry-run 시 생성한 snapshot 으로 복원 (구현 가능 시).
   - 옵션 C: 잘못된 frontmatter 만 manual Edit 으로 복원 (단건).
   - 권장 순서: A → C → B.

**Output**: SPEC corpus 의 status frontmatter 일관화. 잔존 WARNING ≤ 5건.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "StatusGitConsistency"` → ≤ 5.

**Dependencies**: Wave 1 완료 (ERROR 가 잔존하면 자동화 스크립트가 동작하지 않을 수 있음).

**Tools**: Bash (go run, gh pr), Read, Edit, Write (스크립트 작성).

**Risk**: @MX:WARN — 자동화 false-positive 가 예상보다 많을 가능성. dry-run 결과 정확도가 낮으면 수동 검토 부담 증가.

**Reference**: REQ-SLD-007, AC-SLD-007.

### T-SLD-008 — OrphanBCID SPEC-V3R3-ARCH-007 해소

**Target**: SPEC-V3R3-ARCH-007/spec.md frontmatter.

**Action**:
1. SPEC-V3R3-ARCH-007/spec.md frontmatter 의 `breaking` 값 확인 (false).
2. `bc_id` 필드 값 확인.
3. **결정**: `bc_id` 필드 제거 (또는 빈 리스트 `[]` 명시).

**Output**: SPEC-V3R3-ARCH-007 frontmatter 의 `bc_id` 가 OrphanBCID 룰을 satisfy.

**Verification**: `moai spec lint --strict 2>&1 | grep -c "OrphanBCID"` → 0.

**Dependencies**: 없음.

**Tools**: Read, Edit.

**Reference**: REQ-SLD-008, AC-SLD-008.

---

## 4. Wave 3 — 검증 + PR 생성 + 머지

### T-SLD-009 — 통합 lint 재실행 및 카운트 검증

**Action**:
1. `moai spec lint --strict` 실행 후 stdout 전체 캡쳐.
2. ERROR 카테고리별 카운트 측정 (FrontmatterInvalid / MissingExclusions / MissingDependency / DependencyCycle / ModalityMalformed / CoverageIncomplete).
3. WARNING 카테고리별 카운트 측정 (StatusGitConsistency / OrphanBCID).
4. 목표 vs 실제 비교: ERROR 0 / WARNING ≤ 5 인지 확인.
5. 실패 시 해당 카테고리 task 로 복귀.

**Output**: lint 통과 보고서 (`.moai/specs/SPEC-V3R4-SPECLINT-DEBT-001/lint-final.md`).

**Verification**: `moai spec lint --strict; echo $?` → 0.

**Dependencies**: T-SLD-001 ~ T-SLD-008 모두 완료.

**Tools**: Bash, Write.

**Reference**: Gate G1, G2, G3.

### T-SLD-010 — plan-auditor self-review

**Action**:
1. orchestrator 가 plan-auditor 를 호출하여 본 SPEC 의 plan.md / spec.md / acceptance.md 를 audit.
2. plan-auditor 가 lint debt 신규 생성 여부 확인.
3. 발견된 lint debt 가 있으면 본 SPEC 의 plan 자체에 반영.

**Output**: plan-auditor PASS 확인.

**Verification**: plan-auditor 보고서가 `PASS` 또는 점수 ≥ 0.85.

**Dependencies**: T-SLD-009 (전체 lint 통과 후 audit).

**Tools**: Agent (plan-auditor).

**Reference**: Gate G6.

### T-SLD-011 — Run PR 생성 및 CI GREEN 확인

**Action**:
1. 모든 변경사항을 worktree `feat/SPEC-V3R4-SPECLINT-DEBT-001` 브랜치에 commit.
2. `gh pr create --base main --title "feat(specs): SPEC-V3R4-SPECLINT-DEBT-001 spec-lint debt 일괄 해소" --body "..."` 로 PR 생성.
3. GitHub Actions `spec-lint` workflow 실행 대기.
4. `gh pr checks <PR-number>` 로 GREEN 확인.
5. 머지 (admin squash, plan-in-main + run-in-worktree doctrine).

**Output**: Run PR 머지 완료. main HEAD 가 spec-lint --strict exit 0 상태.

**Verification**: `gh pr checks <PR-number>` 의 `spec-lint` job 이 success.

**Dependencies**: T-SLD-010 (audit 통과 후 PR).

**Tools**: Bash (gh, git), manager-git delegation.

**Reference**: REQ-SLD-009, AC-SLD-009, Gate G4.

---

## 5. Technical Approach

### 5.1 SPEC frontmatter 컨벤션 (T-SLD-001 reference)

V3R2 / V3R3 / V3R4 series 의 frontmatter schema 는 다음 7개 mandatory 필드를 공유한다:

```yaml
---
id: SPEC-<DOMAIN>-<NUM>
version: "<MAJOR>.<MINOR>.<PATCH>"  # quoted string
status: <draft|in_review|in_progress|implemented|completed|superseded>
created: <YYYY-MM-DD>
updated: <YYYY-MM-DD>
author: <agent name | role>
priority: <P0|P1|P2|P3|High|Medium|Low|Critical>
tags: "<comma-separated string>"  # OR YAML array (legacy)
title: <Human readable title>
phase: "<phase identifier>"
module: "<affected directories or skills>"
lifecycle: <spec-first|spec-anchored|spec-as-source>
---
```

추가 optional 필드: `dependencies`, `depends_on`, `supersedes`, `related_specs`, `breaking`, `bc_id`, `issue_number`, `target_release`, `related_theme`.

### 5.2 CoverageIncomplete 분석 휴리스틱 (T-SLD-006 reference)

REQ-XXX-NNN 가 어떤 AC 에서도 참조되지 않을 때의 결정 트리:

```
1. spec.md 의 REQ-XXX-NNN 본문이 명확한 EARS 패턴인가?
   - YES → 결정 (a) 또는 (c). 본문이 testable 한 의도를 담고 있음.
   - NO  → 결정 (b). orphan placeholder.

2. acceptance.md 에 implicit 으로 검증하는 AC 가 있는가?
   - YES → 결정 (c). 기존 AC 에 `Reference: REQ-XXX-NNN` 추가.
   - NO  → 결정 (a). 신규 AC 작성.

3. REQ 본문이 다른 REQ 와 거의 동일한 의도인가?
   - YES → 결정 (b). 중복 REQ 제거 + 다른 REQ 의 본문에 의도 통합.
   - NO  → 결정 (a) 또는 (c).
```

### 5.3 Status 추론 로직 (T-SLD-007 reference)

```go
func InferStatus(specID string) string {
    plans := gh.PullRequests("head:plan/" + specID, "is:merged")
    runs  := gh.PullRequests("head:feat/" + specID, "is:merged")
    syncs := gh.PullRequests("head:sync/" + specID, "is:merged")

    if len(syncs) > 0 && len(runs) > 0 {
        return "completed"
    }
    if len(runs) > 0 {
        return "implemented"
    }
    if len(plans) > 0 {
        return "planned"
    }
    return "draft"
}
```

False-positive 케이스 (수동 검토 필요):
- SPEC 가 다른 SPEC 의 PR 에 포함되어 머지된 경우 (예: bundled SPEC).
- supersede 관계로 후속 SPEC 에 의해 status 가 superseded 로 변경된 경우.
- Hotfix 또는 cherry-pick 으로 별도 PR 없이 머지된 경우.

---

## 6. Risks

(spec.md §6 참조. plan.md 에서는 task 수준의 추가 risk 만 기재.)

- **R6 (T-SLD-006 specific)**: CoverageIncomplete 실제 케이스가 60건+ 이면 단일 task 로 처리 시 PR 크기 과대. Wave 1 초입에 sub-task 분할 결정.
- **R7 (T-SLD-007 specific)**: Go 자동화 스크립트의 false-positive rate 가 30% 이상이면 수동 검토 부담이 자동화 이득을 상회. 이 경우 수동 batch 처리로 전환.
- **R8 (T-SLD-011 specific)**: PR 크기가 188 SPEC 전반에 걸쳐 수천 라인 diff 가 될 수 있음. reviewer 부담 완화를 위해 commit 을 카테고리별로 분할 (frontmatter / depends_on / coverage / status / orphan-bcid 5 commit).

---

## 7. Open Questions

@MX:NOTE — 다음은 plan-phase 에서 답을 미루고 run-phase 에서 결정할 항목:

- **OQ1**: SCH-001, ARCH-003 가 실제로 다른 SPEC 의 별칭/리네임인지 (spec.md §6 R2). run-phase T-SLD-002 시작 전에 git log + grep 으로 확인.
- **OQ2**: RT-004 ↔ RT-005 의 의존성 방향 의도 (spec.md §6 R3). run-phase T-SLD-003 시작 전에 두 SPEC 본문 검토.
- **OQ3**: StatusGitConsistency 자동화 스크립트의 false-positive 임계값. Wave 2 T-SLD-007 dry-run 결과를 보고 결정.
- **OQ4**: PR commit 분할 전략 (단일 commit vs 카테고리별 5 commit). T-SLD-011 직전에 reviewer 와 협의.

---

## 8. References

- spec.md §3 REQ-SLD-001 ~ REQ-SLD-010
- acceptance.md AC-SLD-001 ~ AC-SLD-010 + Gate G1 ~ G6
- `.claude/rules/moai/workflow/spec-workflow.md` § Plan Phase / Run Phase
- `.claude/rules/moai/workflow/mx-tag-protocol.md` § MX Tag Types
- CLAUDE.local.md §18.3 (Merge Strategy), §18.12 (BODP)
