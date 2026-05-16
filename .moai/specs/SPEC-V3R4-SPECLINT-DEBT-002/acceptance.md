# SPEC-V3R4-SPECLINT-DEBT-002 — Acceptance Criteria

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.0   | 2026-05-16 | manager-spec | Iteration 2 revision (plan-auditor review-1 0.62 FAIL → revise). D2: AC-004를 Case C (team/plan.md에 신설)로 재구성 — Case A/B 양분 제거 (plan-phase에서 file 존재 + checklist 부재 실측 완료). D5: AC-002 finding count "3개 누락" → "정확히 3개" 강화 (≥3 모호성 제거). D6: AC-003 verification command #3을 binary `wc -l = 0` 형식으로 명료화 (1-month deferred 제거). D7: line citation 정확도 (lint.go:518-533). EC-006 신설: dual-field cleanup scope creep mitigation. |
| 0.1.0   | 2026-05-16 | manager-spec | 초기 draft. 5개 EARS requirements에 매핑된 5개 acceptance scenarios (Given-When-Then 형식). Edge cases + Definition of Done + Quality Gates 포함. |

---

## 1. Overview

본 acceptance.md는 SPEC-V3R4-SPECLINT-DEBT-002의 5개 EARS requirements가 implementation 완료 시점에 검증 가능함을 정의한다. 각 AC는 Given-When-Then 시나리오 + 구체적 verification command를 포함한다.

### 1.1 AC ↔ REQ Mapping

| AC ID | REQ ID | Wave | Verification |
|-------|--------|------|--------------|
| AC-SDBT-002-001 | REQ-SDBT-002-001 | W1 | `moai spec lint --strict` exit 0 |
| AC-SDBT-002-002 | REQ-SDBT-002-002 | W1 | grep + lint dry-run |
| AC-SDBT-002-003 | REQ-SDBT-002-003 | W1 + W3 | Hotfix scenario reproduction test |
| AC-SDBT-002-004 | REQ-SDBT-002-004 | W3 | File existence + sync diff |
| AC-SDBT-002-005 | REQ-SDBT-002-005 | W2 | Cross-reference grep (3+ locations) |

---

## 2. Acceptance Scenarios

### AC-SDBT-002-001 (manager-spec workflow 일관성)

**maps REQ-SDBT-002-001**

**Given**:
- `.claude/skills/moai/workflows/plan.md` § "Pre-Write Frontmatter Checklist"가 12-field canonical로 갱신되었다.
- 갱신된 plan workflow를 가이드로 manager-spec이 새 SPEC 작성을 시작한다.

**When**:
- manager-spec이 plan.md 12-field 목록을 따라 frontmatter를 작성한다.
- 작성된 spec.md를 `moai spec lint --strict`로 검증한다.

**Then**:
- 검증 결과 FrontmatterInvalid finding이 **0건**이어야 한다.
- exit code 0 이어야 한다.
- 누락 field 경고 ("Frontmatter required field missing: ...") 0건이어야 한다.

**Verification Command**:
```bash
# 본 SPEC spec.md 자체를 dry-run 검증 (이미 12-field로 작성됨)
moai spec lint --strict .moai/specs/SPEC-V3R4-SPECLINT-DEBT-002/spec.md

# Expected output:
# (no findings)
# exit code: 0
```

---

### AC-SDBT-002-002 (snake_case rejection)

**maps REQ-SDBT-002-002**

**Given**:
- plan.md "Rejected legacy aliases" 섹션이 반전되어 `created_at:`, `updated_at:`, `labels:`, `spec_id:`를 reject로 명시한다.
- Wave 3 fixture `internal/spec/testdata/frontmatter-schema/invalid-snake-case-only/spec.md`가 snake_case alias만 보유 (`created_at:`, `updated_at:`, `labels:`)하고 canonical (`created:`, `updated:`, `tags:`)을 누락. 다른 9 required fields는 모두 정상값.

**When**:
- fixture를 lint.go FrontmatterSchemaRule.Check()로 검증한다.

**Then**:
- lint.go는 **정확히 3개 FrontmatterInvalid findings**를 보고해야 한다 — 각각 "Frontmatter required field missing: created", "...: updated", "...: tags".
- 정확히 3개 (≥3 아님) — unknown YAML keys (`created_at`/`updated_at`/`labels`)는 struct decoder가 silently drop하므로 findings에 추가 기여 없음 (lint.go 동작 근거: `internal/spec/lint.go:534-545`).

**Verification Command**:
```bash
# Wave 3 fixture로 검증
go test ./internal/spec/ -run TestFrontmatterSchemaRule_SnakeCaseRejected -v

# Expected:
# === RUN   TestFrontmatterSchemaRule_SnakeCaseRejected
# --- PASS: TestFrontmatterSchemaRule_SnakeCaseRejected (0.00s)
#     ✓ exactly 3 findings: {created, updated, tags} missing
# PASS

# Manual cross-verification
moai spec lint --strict internal/spec/testdata/frontmatter-schema/invalid-snake-case-only/spec.md 2>&1 | \
  grep -c "Frontmatter required field missing"
# Expected: 3
```

---

### AC-SDBT-002-003 (regression evidence — SDF-001 hotfix reproduction)

**maps REQ-SDBT-002-003**

**Given**:
- SDF-001 (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001) run-phase에서 hotfix commit `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강`이 필요했다.
- 본 SPEC-V3R4-SPECLINT-DEBT-002 implementation이 완료되어 plan workflow가 12-field canonical로 정렬되었다.

**When**:
- SDF-001과 동일한 시나리오를 재현한다: manager-spec이 plan workflow를 충실히 따라 frontmatter를 작성한다 (당시 9-field에서 7-field 부족).
- 작성된 spec.md를 `moai spec lint --strict`로 검증한다.

**Then**:
- hotfix commit이 **더 이상 필요하지 않다** — manager-spec이 plan workflow 가이드만으로 12-field frontmatter를 작성하므로 lint pass.
- sync-phase 시점 (본 SPEC 머지 직전) baseline: 본 SPEC 머지 이후 추가 frontmatter hotfix commit 0건 검증.

**Verification Command**:
```bash
# 1. plan workflow body 직접 따라하기 dry-run (12-field 명시 확인)
cat .claude/skills/moai/workflows/plan.md | grep -E '^- \[ \] ' | grep -E '(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):' | wc -l
# Expected: 12

# 2. 본 SPEC spec.md 자체가 plan workflow 가이드만으로 lint pass
moai spec lint --strict .moai/specs/SPEC-V3R4-SPECLINT-DEBT-002/spec.md
# Expected: 0 findings, exit 0

# 3. Binary regression check (sync-phase에서 실행 — 본 SPEC 머지 이후 추가 hotfix 0건)
git log --since=2026-05-16 --grep "fix(spec).*frontmatter canonical" --oneline | wc -l
# Expected: 0 (sync-phase baseline — 본 SPEC 자체 commit은 제외, fix(spec) 패턴만 카운트)
```

---

### AC-SDBT-002-004 (team/plan.md Pre-Write Checklist 신설 — Case C)

**maps REQ-SDBT-002-004**

**Given**:
- Plan-phase 실측 baseline: `.claude/skills/moai/team/plan.md` 존재 (8065 bytes, 2026-05-14) + Pre-Write Frontmatter Checklist 섹션 부재 (`grep -i "frontmatter\|12 required\|9 required" team/plan.md` → 0 matches).
- team 모드에서도 SPEC 작성은 manager-spec subagent에 위임 (team/plan.md:157 "Delegate SPEC creation to manager-spec subagent").

**When (Case C — 신설)**:
- Wave 3 T-Wave3-001에서 team/plan.md의 manager-spec delegation step 직후 (line 155 다음, "## Phase 4: User Approval" 헤더 직전)에 solo plan.md (lines 446-471)의 Pre-Write Frontmatter Checklist 섹션을 verbatim mirror하여 신설. 12 required fields + optional fields + rejected aliases + pre-write gate behavior 모두 포함. team-mode 적응 (예: "manager-spec coordinator (team-mode)" 명시)만 차이.

**Then (Case C)**:
- team/plan.md에 "Pre-Write Frontmatter Checklist" 섹션이 존재.
- 12 required fields가 solo plan.md와 동일 (whitespace/team-mode adaptation 제외 의미적 동등).
- Rejected aliases (`created_at:`, `updated_at:`, `labels:`) 명시.
- ~75 LOC 삽입.

**Verification Command**:
```bash
# 1. 섹션 존재
grep -A 5 'Pre-Write Frontmatter Checklist' .claude/skills/moai/team/plan.md | head -10
# Expected: header + 본문 시작 출력

# 2. 12 fields 모두 명시
grep -c -E '^- \[ \] `(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):' .claude/skills/moai/team/plan.md
# Expected: 12

# 3. Rejected aliases 명시
grep -c -E '`(created_at|updated_at|labels):`' .claude/skills/moai/team/plan.md
# Expected: ≥3 (in rejected section context)

# 4. Solo vs team 의미적 동등성 (Required fields 블록)
diff <(grep -A 12 'Required 12 fields' .claude/skills/moai/workflows/plan.md) \
     <(grep -A 12 'Required 12 fields' .claude/skills/moai/team/plan.md)
# Expected: 차이 없음 또는 team-mode 단어 (manager-spec coordinator 등)만 다름
```

---

### AC-SDBT-002-005 (SSOT cross-reference 3-layer)

**maps REQ-SDBT-002-005**

**Given**:
- Wave 2에서 `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT 문서가 신설되었다.
- 3개 enforcement layer (plan workflow body, lint.go FrontmatterSchemaRule, plan-auditor)가 모두 SSOT 문서를 cross-reference 한다.

**When**:
- 코드베이스 grep으로 SSOT 파일명 발견 위치를 카운트한다.

**Then**:
- `grep -rln "spec-frontmatter-schema.md"` 결과가 **최소 3개 파일**에서 매칭되어야 한다:
  1. `.claude/skills/moai/workflows/plan.md` (Pre-Write Checklist 상단 cross-reference)
  2. `internal/spec/lint.go` (FrontmatterSchemaRule godoc 주석)
  3. `.claude/rules/moai/development/coding-standards.md` 또는 `.claude/rules/moai/workflow/spec-workflow.md` (hub link)
- 추가로 `.claude/rules/moai/development/spec-frontmatter-schema.md` 자체가 존재.

**Verification Command**:
```bash
# 1. SSOT 파일 존재 확인
ls .claude/rules/moai/development/spec-frontmatter-schema.md
# Expected: file exists

# 2. Cross-reference count
grep -rln "spec-frontmatter-schema.md" .claude/ internal/ | grep -v "spec-frontmatter-schema.md:" | sort -u | wc -l
# Expected: ≥ 3

# 3. Verify specific layers
grep -l "spec-frontmatter-schema.md" .claude/skills/moai/workflows/plan.md
grep -l "spec-frontmatter-schema.md" internal/spec/lint.go
# Expected: both match
```

---

## 3. Edge Cases

### EC-001: lint.go에 13th field 추가가 SPEC 작업 중 발생

**Scenario**: 본 SPEC implementation 중 다른 SPEC이 lint.go에 새 field 추가 → 본 SPEC plan.md 갱신 시점에 13-field가 진실.

**Mitigation**:
- run-phase Wave 1 시작 시 `git log -p internal/spec/lint.go` 검토 + FrontmatterSchemaRule.Check() `required[]` slice (`internal/spec/lint.go:518-533`) 실측.
- field 수가 12와 다를 경우 plan.md 갱신을 실측에 맞춤. spec.md REQ-SDBT-002-001 본문은 "12 fields" → "12 fields (실측 기준 N fields)" 추가 메모.

### EC-002: 197 SPECs corpus baseline dual-field 패턴 (plan-phase 실측 baseline)

**Scenario**: Plan-phase 실측 결과 — 197 SPECs 중 51건이 `created_at:` snake_case + canonical `created:` dual-field, 53건이 `labels:` + canonical `tags:` dual-field 보유 (lint regression 0건). Run-phase Wave 1 시작 시 동일 baseline 재확인.

**Mitigation**:
- Wave 1 시작 시 `grep -l "^created_at:" .moai/specs/*/spec.md | wc -l` 재측정 (expected: 51 ±5, 다른 SPEC 머지 영향 흡수).
- 신규 SPEC이 추가되었을 가능성 — count 변동은 정상, lint regression 0건 유지가 핵심.
- 본 SPEC scope에서 dual-field cleanup 제외 (F-OUT-001 + EC-006). 별도 SPEC 후보로 분리.

### EC-003: plan-auditor가 plan.md schema 본문을 검증하지 않음

**Scenario**: REQ-SDBT-002-005 "plan-auditor가 SSOT 참조"가 vacuous — plan-auditor가 본문 schema 검증을 수행하지 않을 수 있음.

**Mitigation**:
- Wave 3 T-Wave3-004에서 plan-auditor 코드 분석. 검증 미수행 시 REQ-005를 "plan-auditor 문서/주석에 SSOT 참조 추가"로 약화 (plan-auditor README 또는 godoc).

### EC-004: SSOT 문서 신설 후 다른 SPEC이 기존 frontmatter 관련 문서를 신설 (중복)

**Scenario**: SSOT 신설 시점에 다른 SPEC이 별도 frontmatter 가이드를 작성하여 중복 발생.

**Mitigation**:
- Wave 2 시작 전 `find .claude/rules/moai/development/ -name '*frontmatter*' -o -name '*schema*'` 사전 검토.
- 중복 발견 시 PR description에 명시 + reviewer가 통합 결정.

### EC-005: lint.go 주석 추가가 Go test 실패 트리거

**Scenario**: T-Wave2-003에서 lint.go 주석 추가가 godoc linter 규칙 위반 (예: comment too long, link format 잘못).

**Mitigation**:
- Wave 2 후 `go test ./internal/spec/... && go vet ./internal/spec/...` 실행.
- 실패 시 주석 형식 조정 (한 줄로 줄이거나 // (See ...) 형식 사용).

### EC-006: dual-field SPECs cleanup scope creep 압박

**Scenario**: 본 SPEC implementation 중 51건 `created_at:` + 53건 `labels:` dual-field SPECs cleanup이 "당장 처리" 요구로 압박. plan workflow는 정렬되지만 redundant 메타데이터는 그대로 남음.

**Mitigation**:
- spec.md §1.2 #1, §5 (Exclusions) #1, F-OUT-001 모두에서 "opportunistic cleanup, 본 SPEC scope 외" 명시.
- 별도 follow-up SPEC 후보 (예: SPEC-V3R4-SPECLINT-DEBT-003 dual-field cleanup) 제안.
- 본 SPEC 머지 이후 plan workflow 정렬로 신규 SPEC에서는 redundant dual-field 생성 차단됨 — 기존 dual-field는 시간 경과 + 별도 SPEC으로 자연 감소.

---

## 4. Quality Gates

### 4.1 Wave-level Gates

| Gate | Verification | Block on fail |
|------|--------------|---------------|
| Wave 1 verification | T-Wave1-005 grep 3건 모두 expected output | Yes |
| Wave 2 verification | SSOT 파일 존재 + cross-ref 3+ 매칭 | Yes |
| Wave 3 verification | regression test 2건 PASS | Yes |
| `moai spec lint --strict` (본 SPEC 디렉토리) | 0 ERROR / 0 WARNING | Yes |
| `go test ./internal/spec/...` | All PASS | Yes |
| spec-lint CI job (sync-phase) | GREEN | Yes |

### 4.2 TRUST 5 Compliance

- **T (Tested)**: Wave 3 regression test 2건 (TestFrontmatterSchemaRule_Valid12Field, TestFrontmatterSchemaRule_SnakeCaseRejected). lint.go FrontmatterSchemaRule 자체는 lint_test.go 기존 coverage 활용.
- **R (Readable)**: plan.md 갱신 + SSOT 문서 신설이 12-field schema를 명시적으로 표현. naming 변경 없음 (기존 lint.go field name 보존).
- **U (Unified)**: SSOT 문서 신설로 plan.md / lint.go / plan-auditor가 단일 source 참조 → schema 변경 시 한 곳만 수정.
- **S (Secured)**: schema 검증 강화로 frontmatter 위반이 CI 단계에서 detect (PR merge 차단 가능).
- **T (Trackable)**: HISTORY entry 추가 + CHANGELOG `[Unreleased]` 갱신 + commit message에 SPEC ID 명시.

---

## 5. Definition of Done

### 5.1 spec.md / plan.md / acceptance.md (planning artifacts)

- [x] spec.md 12-field frontmatter (lint.go canonical 준수, 본 SPEC 자체로 dogfooding)
- [x] plan.md Wave 분해 + risks + OOS 명시
- [x] acceptance.md 5 AC 모두 Given-When-Then + verification command

### 5.2 Implementation (run-phase)

- [ ] Wave 1: plan.md Pre-Write Checklist 12-field 갱신
- [ ] Wave 2: SSOT 문서 신설 + 3+ cross-reference
- [ ] Wave 3: team/plan.md에 Pre-Write Checklist 섹션 신설 (Case C, ~75 LOC) + regression fixture 2건 + test 2건

### 5.3 Verification (run-phase 완료 시)

- [ ] AC-SDBT-002-001: `moai spec lint --strict` 본 SPEC 디렉토리 exit 0
- [ ] AC-SDBT-002-002: `go test ./internal/spec/ -run TestFrontmatterSchemaRule_SnakeCaseRejected` PASS — 정확히 3 findings
- [ ] AC-SDBT-002-003: plan workflow body grep 12 fields + 본 SPEC spec.md lint pass + `git log --since=2026-05-16 --grep "fix(spec).*frontmatter canonical" --oneline | wc -l` = 0
- [ ] AC-SDBT-002-004: team/plan.md에 Pre-Write Checklist 섹션 신설 (Case C) — 12 fields + rejected aliases + gate behavior 모두 명시
- [ ] AC-SDBT-002-005: `grep -rln "spec-frontmatter-schema.md"` ≥ 3 파일

### 5.4 Sync (sync-phase)

- [ ] CHANGELOG `[Unreleased]` 항목 추가
- [ ] frontmatter `status: draft → implemented → completed` transition
- [ ] spec-lint CI job GREEN
- [ ] PR merge (plan / run / sync 모두 squash)
- [ ] sync HISTORY entry 추가 (divergence 분석 포함)

---

## 6. Out of Scope (Acceptance-level)

다음 항목은 본 SPEC acceptance에서 **검증하지 않는다**:

- 197 기존 SPECs frontmatter 양방향 검증 (모두 lint pass; dual-field cleanup은 별도 SPEC)
- title/phase/module/lifecycle 값 의미 검증 (별도 SPEC 후보)
- issue_number tracking 정책 검증
- plan-auditor 전체 동작 검증 (본 SPEC은 SSOT 참조 추가만)
- spec-lint CI workflow YAML 변경 (별도 SPEC)
- IDE/editor plugin 의 frontmatter 자동완성 (out of moai scope)
