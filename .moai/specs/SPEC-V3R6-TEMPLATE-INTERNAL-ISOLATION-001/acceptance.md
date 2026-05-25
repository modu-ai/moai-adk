---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Acceptance — Template Internal-Content Isolation"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "template, isolation, internal-content, acceptance"
tier: M
---

# Acceptance Criteria — Template Internal-Content Isolation

## §A. AC ↔ REQ Bidirectional Traceability

| AC ID | Bound REQ | Title | Severity |
|-------|-----------|-------|----------|
| AC-TII-001 | REQ-TII-001 | 35 잔여 leak files 모두 cleanup 완료 (grep count → 0) | MUST-PASS |
| AC-TII-002 | REQ-TII-002 | Substitution pattern 일관성 (predecessor pass 1/2와 호환) | MUST-PASS |
| AC-TII-003 | REQ-TII-003 | Mirror parity (make build clean + embedded.go regeneration) | MUST-PASS |
| AC-TII-004 | REQ-TII-004 | CLAUDE.local.md §25 신규 작성 (4 subsection ≥80 lines) | MUST-PASS |
| AC-TII-005 | REQ-TII-005 | §25 pre-commit self-check checklist 5-item 존재 | MUST-PASS |
| AC-TII-006 | REQ-TII-006 | Go lint test `TestTemplateNoInternalContentLeak` 신규 작성 + 위치 정확 | MUST-PASS |
| AC-TII-007 | REQ-TII-007, REQ-TII-008 | Lint test red-green proof (cleaned HEAD PASS + synthetic leak FAIL) | MUST-PASS |
| AC-TII-008 | REQ-TII-009, REQ-TII-010 | CI workflow extension + docstring cross-reference | MUST-PASS |
| AC-TII-009 | REQ-TII-011 | Maintainer-only file audit 통과 (templates 내 §21 dev-only 부재) | MUST-PASS |
| AC-TII-010 | REQ-TII-012 | Audit 결과에 따른 git rm + .gitignore 갱신 (조건부) | MUST-PASS |
| AC-TII-011 | REQ-TII-013 | Anti-pattern enforcement: 본 SPEC 자체가 안티패턴 도입하지 않음 | MUST-PASS |
| AC-TII-012 | REQ-TII-001~013 (cross-cutting) | TRUST 5 회귀 없음 + 기존 Go test 모두 PASS | MUST-PASS |

REQ↔AC 양방향 traceability: 13 REQ → 12 AC, 0 orphan REQ, 0 orphan AC (REQ-TII-001~013가 AC-TII-001~012에 각각 1:1 + cross-cutting 매핑됨). 100% coverage.

---

## §B. Per-AC Verifiable Commands

### AC-TII-001 — 35 leak files cleanup

**Command**:
```bash
grep -rln 'SPEC-V3R6-\|REQ-ATR-\|Audit 3\|Finding A[1-6]\|archive-2026-05-25' internal/template/templates/ | wc -l
```
**Expected**: `0` (정확히 0, allowlist 항목이 있더라도 grep 결과에서 제외되어야 함 — allowlist 메커니즘은 lint test 내부에서만 적용)
**Verification owner**: manager-develop (M4 종료 후)

### AC-TII-002 — Substitution pattern 일관성

**Command**:
```bash
# Predecessor cleanup commits에서 사용된 substitution 패턴 표본
git show 20a66df85 -- '*.md' | grep -E '^\+' | grep -iE 'predecessor|선행|earlier|prior' | head -5
# 본 SPEC의 cleanup이 같은 substitution vocabulary를 사용했는지 확인
git diff HEAD~4..HEAD -- 'internal/template/templates/**/*.md' | grep -E '^\+' | grep -iE 'predecessor|선행|earlier|prior|generic' | wc -l
```
**Expected**: 본 SPEC cleanup commits에서 유사 substitution vocabulary 사용 count ≥ 10 (predecessor와 일관)
**Verification owner**: manager-develop + orchestrator independent review

### AC-TII-003 — Mirror parity (make build clean)

**Command**:
```bash
# Baseline (M1 pre-flight) shasum 기록 후 비교
shasum internal/template/embedded.go > /tmp/embedded.before
make build
shasum internal/template/embedded.go > /tmp/embedded.after
# Expectation: embedded.go DOES change (templates were edited), but make build returns 0 with no errors
diff /tmp/embedded.before /tmp/embedded.after
echo "make build exit: $?"
```
**Expected**: `make build` exit code 0, no errors, embedded.go updated cleanly
**Verification owner**: manager-develop (M4 sub-batch 마지막 commit 직후)

### AC-TII-004 — CLAUDE.local.md §25 신규 작성

**Command**:
```bash
grep -c '## 25. Template Internal-Content Isolation' CLAUDE.local.md
awk '/^## 25\. Template Internal-Content Isolation/,/^## 26\.|^---$/{c++} END{print c}' CLAUDE.local.md
```
**Expected**: 첫 번째 grep `1`, 두 번째 awk count ≥ 80 (line count, ≥80 lines body)
**Verification owner**: orchestrator-direct (M2 종료)

### AC-TII-005 — §25 self-check checklist 존재

**Command**:
```bash
awk '/^## 25\. Template Internal-Content Isolation/,/^## 26\.|^---$/{print}' CLAUDE.local.md | grep -cE '^\s*-\s+\[\s*\]' 
```
**Expected**: ≥5 (markdown checklist item `- [ ]` 형식 ≥5개)
**Verification owner**: orchestrator-direct (M2)

### AC-TII-006 — Go lint test 신규 작성 + 위치 정확

**Command**:
```bash
ls -la internal/template/internal_content_leak_test.go
grep -c '^func TestTemplateNoInternalContentLeak' internal/template/internal_content_leak_test.go
```
**Expected**: 파일 존재 + 함수 정의 1개
**Verification owner**: manager-develop (M3 종료)

### AC-TII-007 — Lint test red-green proof

**Command**:
```bash
# Green proof: cleaned HEAD에서 PASS
go test -run TestTemplateNoInternalContentLeak ./internal/template/... -v
echo "green exit: $?"

# Red proof: synthetic leak 일시 주입 → FAIL → 복원
TARGET="internal/template/templates/.claude/agents/core/manager-spec.md"
cp "$TARGET" "$TARGET.bak"
echo "<!-- synthetic-leak-for-proof: SPEC-V3R6-PROOF-TEST-001 -->" >> "$TARGET"
go test -run TestTemplateNoInternalContentLeak ./internal/template/... 2>&1 | grep -i 'fail\|synthetic-leak-for-proof'
RED_EXIT=$?
mv "$TARGET.bak" "$TARGET"

# Verify restoration
go test -run TestTemplateNoInternalContentLeak ./internal/template/... -v
echo "post-restoration exit: $?"
```
**Expected**: green exit 0 + red proof matches FAIL output containing target file path + post-restoration exit 0
**Verification owner**: manager-develop (M3 종료 + acceptance.md AC-TII-007 reproducible)

### AC-TII-008 — CI workflow extension

**Command**:
```bash
find .github/workflows/ -name '*.yml' -o -name '*.yaml' | xargs grep -l 'go test' | head -3
# Inspection: 어떤 workflow file에 변경이 적용됐는지
git diff HEAD~6..HEAD -- '.github/workflows/' | head -50
# Docstring cross-ref 존재 확인
grep -c 'feedback_template_internal_content_isolation' .github/workflows/*.yml 2>/dev/null || echo "0"
```
**Expected**: 영향 받은 workflow 파일 식별 + `feedback_template_internal_content_isolation` cross-ref count ≥ 1
**Verification owner**: manager-develop (M5 종료)

### AC-TII-009 — Maintainer-only file audit 통과

**Command**:
```bash
# §21 forbidden classes in templates
find internal/template/templates -type f \( -name '97-*.md' -o -name '98-*.md' -o -name '99-*.md' -o -name 'settings.local.json' -o -name 'last-cc-version.json' \) | head -10
find internal/template/templates -path '*moai/state*' -o -path '*moai/research*' -o -path '*moai/cache*' -o -path '*moai/logs*' | head -10
```
**Expected**: 첫 번째 find empty, 두 번째 find empty (templates에 §21 dev-only file 부재)
**Verification owner**: manager-develop (M6 종료)

### AC-TII-010 — Audit 결과 git rm + .gitignore (조건부)

**Command**:
```bash
# M6 audit에서 발견된 경우만 적용; 발견 없으면 trivially PASS
git log --oneline -10 --all -- '.gitignore' | head -5
git diff HEAD~6..HEAD -- '.gitignore' || echo "no gitignore changes (no maintainer-only files found in templates)"
```
**Expected**: 발견 0건 시 .gitignore 변경 없음 (PASS); 발견 N건 시 commit log에 `git rm` 흔적 + .gitignore에 가드 entry 추가
**Verification owner**: manager-develop (M6 종료, 조건부)

### AC-TII-011 — Anti-pattern enforcement (self-check)

**Command**:
```bash
# 본 SPEC 산출물 자체 (3 plan-phase + 추후 run/sync) 가 SPEC ID 또는 REQ token을 인용할 때는 본인의 ID/REQ만 허용
# CLAUDE.local.md §25 작성 시 다른 SPEC ID 인용 금지
grep -oE 'SPEC-V3R6-[A-Z0-9-]+' CLAUDE.local.md | sort -u | grep -v 'TEMPLATE-INTERNAL-ISOLATION-001'
# Expected: empty (다른 SPEC ID 인용 없음 — §25는 generic doctrine만 명시)
```
**Expected**: empty result (§25 본문에서 본 SPEC ID 외 다른 SPEC ID 인용 없음)
**Verification owner**: orchestrator independent verification

### AC-TII-012 — TRUST 5 회귀 없음

**Command**:
```bash
# 기존 Go test 전체 PASS (8 PROCEED-WITH-DEBT 테스트는 별개 SPEC scope이므로 회귀 baseline 기준으로 평가)
go test ./... 2>&1 | tail -20
# Lint baseline 회귀 없음
golangci-lint run --timeout=2m 2>&1 | tail -10
```
**Expected**: 새로운 test failure 0건 (predecessor SPEC-V3R6-TEST-REFACTOR-001 scope의 8 known failures는 baseline으로 인정; 본 SPEC이 추가로 도입한 failure 없음); golangci-lint 회귀 없음
**Verification owner**: manager-develop (M4/M5 종료 후 verification batch)

---

## §C. Definition of Done

본 SPEC은 다음을 모두 만족할 때 DONE:

- [ ] **AC-TII-001 ~ AC-TII-012 모두 PASS** (12개 MUST-PASS, soft-deferred 항목 없음)
- [ ] 4 plan-phase artifacts (spec.md, plan.md, acceptance.md, design.md, research.md) 모두 frontmatter `status: completed` 또는 sync-phase `status: implemented` 로 전환
- [ ] CLAUDE.local.md §25 영구 존재 + cross-reference 통합 (§15, §21, §24)
- [ ] `internal/template/internal_content_leak_test.go` CI에서 PASS
- [ ] M1~M6 6-milestone 모두 closed
- [ ] sync-phase: CHANGELOG.md entry (Korean per `git_commit_messages: ko`) + frontmatter status transition
- [ ] Mx-phase: EVALUATE-SKIP 예상 (markdown/Go test 중심, 추가 mx code 없음)

## §D. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| v0.1.0 | 2026-05-25 | manager-spec | Initial 12 AC-TII bound to 13 REQ-TII; 100% bidirectional traceability; verifiable commands attached |
