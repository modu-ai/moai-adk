# Design — SPEC-V3R4-LINT-SKIP-CLEANUP-001

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.2   | 2026-05-16 | manager-develop (run-phase amend) | run-phase 실측 발견 반영: §9 Non-Goals에 'real status drift 해소'가 본 SPEC scope 외임을 명시 + §10 Future Work에 `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (가설) 긴급 후보 추가. plan-audit D5 §5.4 Mid-Run Crash Recovery 항목 포함 (manager-develop에 의해 본문에 이미 작성). |
| 0.1.1   | 2026-05-16 | plan-audit remediation | plan-audit 0.904 PASS 후 P2 4건 remediation: (1) AC-LSKC-002 placeholder → plan.md §5.2 cross-ref, (2) HISTORY date harmonize 2026-05-16, (3) REQ-005↔AC-002 매핑 rationale 명시, (4) design.md §2.4 redundancy 정리. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 design. 현재 상태(55 SPEC frontmatter)와 목표 상태 sample + bulk edit 알고리즘 pseudocode + 6가지 Edge Case + Rollback plan + Idempotency 검증 절차. Go yaml.v3 Node API 기반 권장. |

---

## 1. Goal

`lint.skip: [StatusGitConsistency]` 회피책을 55개 SPEC frontmatter에서 안전하게 제거하기 위한 구체적인 implementation design.

설계 목표:

1. **Body 완전 보존**: 본문 byte-level 변경 0 (HISTORY 표 1줄 추가는 frontmatter-인접 메타데이터로 간주)
2. **frontmatter key ordering 보존**: 자동 YAML re-serialize로 인한 키 순서 변경 없음
3. **Idempotent**: 2회 실행 시 두 번째는 no-op
4. **Reversible**: PR-level revert로 100% 복원 가능

---

## 2. Current State (cleanup 이전)

### 2.1 frontmatter sample 1 — SPEC-AGENCY-ABSORB-001

```yaml
---
id: SPEC-AGENCY-ABSORB-001
version: 0.3.0
status: completed
created_at: 2026-04-20
updated_at: 2026-04-24
author: GOOS
priority: High
labels: [agency, migration, design, hybrid, absorption]
issue_number: null
merged_pr: 682
merged_commit: 4271fd8a8
deprecation_policy_amended: true
title: "Agency → MoAI-ADK 흡수 및 Claude Design 통합"
created: 2026-04-21
updated: 2026-05-13
phase: "v2.x - Legacy"
module: "agency"
lifecycle: completed
tags: "legacy"
lint:
  skip:
    - StatusGitConsistency
---
```

### 2.2 frontmatter sample 2 — SPEC-AGENT-002

```yaml
---
id: SPEC-AGENT-002
version: 1.0.0
status: completed
created: 2026-04-09
updated: 2026-04-09
author: GOOS
priority: high
issue_number: 0
title: "Agent Definition Optimization - Token Efficiency with Workflow Preservation"
phase: "v2.x - Legacy"
module: "agents"
lifecycle: completed
tags: "legacy"
lint:
  skip:
    - StatusGitConsistency
---
```

### 2.3 frontmatter sample 3 — SPEC-CORE-001 (다른 key ordering)

```yaml
---
id: SPEC-CORE-001
title: Foundation Methodologies
created: 2026-02-03
updated: 2026-05-13
status: completed
priority: Medium
phase: "Phase 5 - Knowledge (Phase 1 과 병렬 구현 가능)"
module: "internal/foundation/"
estimated_loc: "~1000"
dependencies: []
assigned: expert-backend
lifecycle: spec-anchored
tags: "foundation, ears, languages, trust5, domain-patterns"
version: "1.0.0"
author: GOOS
lint:
  skip:
    - StatusGitConsistency
---
```

### 2.4 공통 패턴

- 모든 55개 SPEC frontmatter 가장 아래에 `lint:` 블록이 위치 (`---` closing 직전 3줄)
- `lint:` 다음 줄은 `  skip:` (들여쓰기 2 spaces)
- `  skip:` 다음 줄은 `    - StatusGitConsistency` (들여쓰기 4 spaces)
- 모든 케이스가 단일 엔트리 — 다중 엔트리 케이스 처리 정책은 §6.2 참조

### 2.5 frontmatter key ordering 다양성

55개 SPEC을 sample inspect한 결과:
- `version` 필드 위치: 상단 (id 직후) vs 하단 (tags 직전) 모두 존재
- `updated` 필드 형식: `updated: 2026-04-09` (unquoted ISO date) — 통일
- `updated_at` (구 표준) + `updated` (현 표준) 혼재 SPEC 존재 (SPEC-AGENCY-ABSORB-001 등) — 둘 다 갱신 필요?
- `title` 따옴표 스타일: `"..."` vs unquoted 혼재

---

## 3. Target State (cleanup 이후)

### 3.1 frontmatter sample 1 — SPEC-AGENCY-ABSORB-001 (after)

```yaml
---
id: SPEC-AGENCY-ABSORB-001
version: 0.3.1
status: completed
created_at: 2026-04-20
updated_at: 2026-04-24
author: GOOS
priority: High
labels: [agency, migration, design, hybrid, absorption]
issue_number: null
merged_pr: 682
merged_commit: 4271fd8a8
deprecation_policy_amended: true
title: "Agency → MoAI-ADK 흡수 및 Claude Design 통합"
created: 2026-04-21
updated: 2026-05-15
phase: "v2.x - Legacy"
module: "agency"
lifecycle: completed
tags: "legacy"
---
```

**변경 사항:**
- `lint:` 블록 (3줄) 제거
- `version`: `0.3.0` → `0.3.1` (patch +1)
- `updated`: `2026-05-13` → `2026-05-16`
- `updated_at`: 유지 (별도 필드, 다른 의미) — 본 SPEC scope 외
- HISTORY 표에 row 1줄 추가 (body 영역)

### 3.2 HISTORY 표 row 추가 sample

본문 `## HISTORY` 표 끝에 다음 row 1줄 추가:

```markdown
| 0.3.1   | 2026-05-16 | manager-develop (run-phase) | lint.skip StatusGitConsistency 회피책 제거 — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 walker filter 머지로 불필요해짐. |
```

(version 컬럼은 각 SPEC의 새 버전, 다른 컬럼은 모두 동일)

---

## 4. Bulk Edit Algorithm

### 4.1 옵션 C (Go script) 권장 알고리즘

```
ALGORITHM: cleanup_lint_skip(spec_path)
  INPUT: path to spec.md
  OUTPUT: modified spec.md, idempotent

1. data := os.ReadFile(spec_path)
2. fm_end := find_frontmatter_end(data)
   // first "---\n" at line 0, second "---\n" at fm_end
3. fm_text := data[4 : fm_end]  // skip leading "---\n"
4. body_text := data[fm_end + 4 :]  // skip "---\n"
5. 
6. // YAML parse with node-level fidelity
7. var node yaml.Node
8. yaml.Unmarshal(fm_text, &node)
9. 
10. mapping := node.Content[0]  // top-level mapping node
11. 
12. // Check idempotency: if no `lint` key, skip
13. lint_idx := find_key(mapping, "lint")
14. if lint_idx == -1:
15.   return  // already cleaned
16. 
17. // Remove lint key + value (2 entries in flat list)
18. remove_keys(mapping, "lint")
19. 
20. // Patch bump version
21. version_idx := find_key(mapping, "version")
22. current_version := mapping.Content[version_idx + 1].Value  // string
23. new_version := patch_bump(current_version)  // "0.3.0" -> "0.3.1"
24. mapping.Content[version_idx + 1].Value = new_version
25. 
26. // Update `updated` field (NOT `updated_at`)
27. updated_idx := find_key(mapping, "updated")
28. if updated_idx != -1:
29.   mapping.Content[updated_idx + 1].Value = "2026-05-16"
30. 
31. // Serialize back preserving key order
32. new_fm := yaml.Marshal(&node)
33. 
34. // Add HISTORY row to body
35. new_body := insert_history_row(body_text, new_version, "2026-05-16")
36. 
37. // Write
38. result := "---\n" + new_fm + "---\n" + new_body
39. os.WriteFile(spec_path, result)
```

### 4.2 patch_bump 함수

```go
func patch_bump(version string) string {
    // semver: "X.Y.Z" or "X.Y" (rare)
    // strip quotes if present
    v := strings.Trim(version, `"`)
    parts := strings.Split(v, ".")
    if len(parts) < 3 {
        parts = append(parts, "0")  // "1.0" -> "1.0.0"
    }
    patch, _ := strconv.Atoi(parts[2])
    parts[2] = strconv.Itoa(patch + 1)
    return strings.Join(parts, ".")
}
```

### 4.3 insert_history_row 함수

```go
func insert_history_row(body, new_version, new_date string) string {
    // Find "## HISTORY" heading
    // Find first table row "| X.Y.Z | ... |" in HISTORY section
    // Determine table ordering: top-newest vs bottom-newest
    //   - Sample SPECs vary; check first vs last row date
    // Insert new row at appropriate position (top OR bottom)
    
    new_row := fmt.Sprintf(
        "| %-7s | %-10s | manager-develop (run-phase) | lint.skip StatusGitConsistency 회피책 제거 — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 walker filter 머지로 불필요해짐. |",
        new_version, new_date,
    )
    
    return body_with_row_inserted
}
```

### 4.4 옵션 A polyfill (Edit tool)

manager-develop이 매 SPEC마다 `Read` → 4개 `Edit` op 수행. idempotency는 매 SPEC마다 baseline 비교로 수동 검증.

manager-develop 의사코드:

```
for spec in 55_affected_specs:
    content = Read(spec)
    
    # idempotency check
    if "lint:\n  skip:\n    - StatusGitConsistency" not in content:
        log("already cleaned, skip: " + spec)
        continue
    
    # 1. Remove lint block
    Edit(spec, 
         old_string="lint:\n  skip:\n    - StatusGitConsistency\n",
         new_string="")
    
    # 2. Patch bump version (need to read current version first)
    current_version = extract_version(content)
    new_version = patch_bump(current_version)
    Edit(spec,
         old_string=f"version: {current_version}",
         new_string=f"version: {new_version}")
    
    # 3. Update updated field
    current_updated = extract_updated(content)
    Edit(spec,
         old_string=f"updated: {current_updated}",
         new_string="updated: 2026-05-16")
    
    # 4. Insert HISTORY row (manual placement)
    history_marker = "## HISTORY"
    new_row = f"| {new_version} | 2026-05-16 | manager-develop (run-phase) | ... |"
    # Edit to insert after table header or at appropriate row position
```

옵션 A는 simpler 하지만 sequential Edit op 수가 많고 HISTORY row insertion 위치가 SPEC마다 달라질 수 있음.

---

## 5. Rollback Plan

### 5.1 자동 rollback

PR `feat/SPEC-V3R4-LINT-SKIP-CLEANUP-001` 머지 후 회귀 발견 시:

1. **PR revert**: `gh pr revert <merged-PR-N>` — main에 revert commit 생성
2. **lint --strict 회귀 검증**: revert 후 walker filter는 여전히 적용되어 있으므로 `StatusGitConsistency` WARN 0건 유지될 것 (lint.skip 부재 + walker filter 가 sweep commit skip)
3. **별도 SPEC 발급**: 회귀 원인 분석 후 SPEC-V3R4-LINT-SKIP-CLEANUP-002 로 재시도

### 5.2 수동 rollback (PR 머지 전)

cleanup commit 후 PR 만들기 전 발견 시:
- `git reset --hard origin/main` (단일 commit이므로 안전)

### 5.3 회귀 시나리오 예상

| 시나리오 | 발생 가능성 | rollback 필요? |
|---------|-----------|--------------|
| frontmatter parse 오류 (예: title 따옴표 깨짐) | Low (옵션 C 사용 시 yaml.v3 보장) | YES |
| HISTORY 표 형식 깨짐 (예: 컬럼 수 mismatch) | Medium (옵션 A) | YES |
| 의도하지 않은 SPEC 수정 (55개 외) | Low (M3 verification 통과해야 PR) | YES |
| lint --strict 회귀 (walker filter 동작 안 함) | Very Low (PR #933 머지 확인됨) | NO — predecessor revert 필요 |

### 5.4 Mid-Run Crash Recovery

스크립트가 55개 SPEC 처리 도중 crash(예: OOM, 프로세스 종료) 된 경우:

1. **idempotency 활용**: 스크립트를 재실행하면 이미 처리된 SPEC(lint 블록 부재)은 "already cleaned: skipped" 로그 후 skip됨. 미처리 SPEC은 정상 처리.
2. **검증**: M3 verification은 전체 55 SPEC 기준으로 수행 — partial state에서 재실행 후 전체 AC 통과 확인.
3. **git 상태**: partial commit 없음 (commit은 M4에서 수동 수행). `git stash` 또는 `git diff` 로 처리 진척 파악 가능.

따라서 별도 resume-from-checkpoint 로직 없이 **재실행만으로 recovery 완료**.

---

## 6. Edge Cases

### 6.1 (a) `lint.skip`이 단일 `StatusGitConsistency`만 가진 SPEC (current ALL 55)

**처리**: `lint:` 블록 전체 (3줄) 제거. 빈 `lint:` 블록 또는 `lint.skip: []` 유지 금지.

**근거**: 빈 frontmatter 키는 noise. YAML 표준에서 키 부재가 명시적 빈 값보다 cleaner.

### 6.2 (b) `lint.skip`이 다른 rule + `StatusGitConsistency`를 같이 가진 SPEC (현재 0건, forward-compat)

**처리**: `StatusGitConsistency` 엔트리만 element-level 제거. 나머지 rules + `lint:` 블록 유지.

```yaml
# Before:
lint:
  skip:
    - OrphanBCID
    - StatusGitConsistency
    - MissingExclusions

# After:
lint:
  skip:
    - OrphanBCID
    - MissingExclusions
```

**구현**: 옵션 C 알고리즘 16-19 라인을 `remove_element(skip_array, "StatusGitConsistency")` 로 변경.

### 6.3 (c) 본 SPEC 자체 cleanup (meta self-reference)

**판정**: 본 SPEC (`SPEC-V3R4-LINT-SKIP-CLEANUP-001`)은 cleanup 대상이 아니다 — frontmatter에 `lint:` 블록을 추가하지 않으므로 (REQ-LSKC 가 cleanup의 트리거). M1 baseline scan에서 55 list에 포함되지 않음.

### 6.4 (d) ARCHIVED SPEC (SPEC-I18N-001-ARCHIVED)

**판정**: 정리 대상에 포함. archived 상태와 무관하게 frontmatter consistency 유지.

**근거**: archive 정책상 frontmatter `status: superseded` 또는 `archived`로 표시되지만, `lint.skip` 회피책 metadata는 archive 보존 가치 없음.

### 6.5 (e) `updated_at` vs `updated` 혼재 SPEC

샘플 검증 결과 일부 SPEC은 둘 다 가짐 (SPEC-AGENCY-ABSORB-001):
- `updated_at: 2026-04-24` (구 표준, frontmatter 상단)
- `updated: 2026-05-13` (현 표준)

**판정**: `updated` 필드만 `2026-05-15`로 갱신. `updated_at`은 건드리지 않음 — scope creep 방지.

**근거**: 본 cleanup은 frontmatter 다른 필드 수정 0건 원칙 (REQ-LSKC-007). `updated_at`/`updated` 통합은 별도 SPEC.

### 6.6 (f) 다양한 version 형식

- `version: 0.3.0` (unquoted)
- `version: "1.0.0"` (quoted string)
- `version: 0.3` (rare — patch 부재)

**처리**:
- unquoted/quoted 스타일 보존 (옵션 C yaml.Node)
- patch 부재 시: `0.3` → `0.3.1` (patch 추가)

### 6.7 (g) HISTORY 표 ordering 다양성

샘플 검증:
- top-newest (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001): 새 row 가 위
- bottom-newest (SPEC-AGENT-002): 새 row 가 아래
- single-row (SPEC-CORE-001): 최초 row만 존재

**처리**: M1 baseline에서 SPEC별 ordering 감지 → 적절한 위치 삽입. heuristic:
- 표 헤더 바로 아래 first row 의 version vs last row 의 version 비교
- newer version 이 위면 top-newest → 새 row 를 헤더 바로 아래 삽입
- older version 이 위면 bottom-newest → 새 row 를 표 끝에 append

---

## 7. Idempotency Verification

### 7.1 옵션 C script idempotency

```bash
go run .moai/scripts/lint-skip-cleanup.go
git diff --stat .moai/specs/ > /tmp/run1.txt

go run .moai/scripts/lint-skip-cleanup.go
git diff --stat .moai/specs/ > /tmp/run2.txt

diff /tmp/run1.txt /tmp/run2.txt && echo "PASS: idempotent" || echo "FAIL: not idempotent"
```

**기대 결과**: `/tmp/run1.txt` 와 `/tmp/run2.txt` 동일 (두 번째 실행은 no-op).

### 7.2 옵션 A (Edit tool) idempotency

매 SPEC iteration 시작 시 검사:

```python
content = Read(spec)
if "lint:\n  skip:\n    - StatusGitConsistency" not in content:
    log(f"already cleaned: {spec}")
    continue  # idempotent skip
```

이렇게 하면 manager-develop이 partial run 도중 중단 후 재시작해도 안전.

---

## 8. Architecture Decision Records (ADR)

### ADR-001: Bulk script vs Edit tool

**Decision**: 옵션 C (Go script) 권장. run-phase 최종 결정.

**Rationale**:
- 55 SPEC iteration → Edit tool 옵션 A는 270+ tool calls 발생 → token budget pressure
- 옵션 C는 단일 script 실행으로 처리 + idempotent 자연스러움
- script 재사용성: 향후 유사 metadata cleanup 패턴에 응용 (예: 다른 lint rule deprecation)

**Trade-off**: Go script 작성 LOC 100~150 추가. 단, 검토 가능한 코드이며 PR에 포함하지 않거나 `.moai/scripts/` 보존 결정 (OQ2).

### ADR-002: HISTORY row 추가 위치

**Decision**: SPEC별 ordering 감지 (top-newest vs bottom-newest) 후 맞춤 삽입.

**Rationale**:
- 일관성 vs 보존성 trade-off
- 모든 SPEC을 일괄 top-add 또는 bottom-add 하면 일부 SPEC의 기존 ordering 깨짐
- 기존 ordering 보존이 git history 가독성 우선

### ADR-003: `updated_at` 보존, `updated`만 갱신

**Decision**: `updated_at` 은 건드리지 않음.

**Rationale**:
- `updated_at` 은 구 표준 (deprecation 진행 중), 별도 마이그레이션 SPEC 영역
- 본 SPEC은 lint.skip cleanup만 처리 — scope creep 방지

### ADR-004: SPEC 본문 sha256 비교를 검증 절차에 포함

**Decision**: M3 verification 단계에서 본문 sha256 비교 (HISTORY 표 영역 제외).

**Rationale**:
- "본문 0줄 수정" 약속 (AC-LSKC-003) 의 자동 검증 방법
- manual diff 검토는 55 SPEC × 평균 ~300줄 = 16,500줄 → 무리
- HISTORY 표 영역만 별도 추적 (table row 추가는 frontmatter-인접 메타 변경으로 간주)

---

## 9. Non-Goals

- Lint engine 자체의 동작 변경 (walker filter는 PR #933 완료, 본 SPEC 무관)
- `lint.skip` 메커니즘 자체의 폐기 (다른 lint rule skip 정당한 사례 존재 가능)
- 다른 lint rule 의 lint.skip 엔트리 정리 (예: OrphanBCID, MissingExclusions) — 별도 SPEC
- frontmatter key ordering 표준화 — 별도 SPEC
- `updated_at` ↔ `updated` 필드 통합 — 별도 SPEC
- `version` 필드 quoted vs unquoted 스타일 통일 — 별도 SPEC
- **Real status drift 해소 (54 SPECs)**: run-phase 실측에서 lint.skip suppression 해제 후 노출된 `StatusGitConsistency` WARN — `feat:`/`feat(specs):` commit 기반 status mismatch — 는 본 SPEC scope 외. follow-up SPEC `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` (가설) 에서 해결 권장.

---

## 10. Future Work

- **SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001** (긴급 후보): run-phase 실측 발견 — 본 cleanup 후 64 `StatusGitConsistency` WARN 노출 (cleanup population 55 중 54 + 다른 SPECs 10건 unrelated). 옵션: (a) 영향 SPECs frontmatter `status` 필드를 git-implied status와 동기화, (b) walker filter scope를 `feat(specs):` 등으로 확대, (c) status drift detection rule 자체를 deprecate.
- **SPEC-V3R4-LINT-SKIP-CLEANUP-002** (가설): 다른 lint rule 의 lint.skip 엔트리 정리 (필요 시)
- **SPEC-V3R4-FRONTMATTER-STANDARDIZATION-001** (가설): `updated_at` / `updated` 통합 + key ordering 표준화 + version quoted 통일
- **SPEC-V3R4-LINT-SKIP-DEPRECATION-001** (가설): `lint.skip` 메커니즘 자체를 deprecate (walker filter 등 근본 해결 후 영구 metadata noise 제거 정책)
