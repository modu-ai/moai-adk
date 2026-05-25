---
id: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001
title: "Acceptance — Template Internal-Content Isolation"
version: "0.1.1"
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

### AC-TII-001 — 35 leak files cleanup (5-class pattern)

**Command** (5 leak classes: 4-class prose + commit sha):
```bash
( grep -rln 'SPEC-V3R6-\|REQ-ATR-\|Audit 3\|Finding A[1-6]\|archive-2026-05-25' internal/template/templates/ ; \
  grep -rlnE '\b[0-9a-f]{40}\b|\b[0-9a-f]{7,8}\b ' internal/template/templates/ ) \
  | sort -u | wc -l
```
**Pattern decomposition** (matches REQ-TII-013 5-class enumeration verbatim):
1. SPEC ID literal: `SPEC-V3R6-...`
2. REQ token literal: `REQ-ATR-...`
3. Audit citation: `Audit 3` or `Finding A[1-6]`
4. Archive date: `archive-2026-05-25`
5. Commit sha: 40-char full sha (`\b[0-9a-f]{40}\b`) OR 7-8 char short sha followed by a space (`\b[0-9a-f]{7,8}\b `) — trailing space disambiguates short sha from random hex (e.g., color codes, hash truncations)

**Expected**: `0` (정확히 0). Allowlist 항목 (예: NOTICE.md Apache 2.0 attribution 날짜는 commit sha 형식이 아님 — `2026-04-26` 같은 ISO date는 hex 패턴 불일치이므로 false-positive 없음; 만약 NOTICE.md가 future에 commit sha를 attribution용으로 인용해야 할 경우 REQ-TII-007 allowlist 메커니즘 적용)는 grep 결과에서 제외되어야 함 — allowlist 메커니즘은 lint test 내부에서만 적용; raw grep은 0이어야 함.

**False-positive risk note** (D-001 amendment rationale): 7-8 char short sha 패턴에 trailing space를 강제함으로써 random hex string (예: CSS color `#abcdef`, base64 fragments, type hashes)과의 충돌을 차단. 40-char full sha는 길이 자체로 false-positive 위험 낮음.

**Verification owner**: manager-develop (M4 종료 후)

### AC-TII-002 — Substitution pattern 일관성

**Command** (SPEC-scoped attribution range — D-003 amendment):
```bash
# Predecessor cleanup commits에서 사용된 substitution 패턴 표본
git show 20a66df85 -- '*.md' | grep -E '^\+' | grep -iE 'predecessor|선행|earlier|prior' | head -5

# 본 SPEC의 cleanup commits를 SPEC ID grep으로 선별 (race-absorbed parallel commits 자동 제외):
#   - attribution base: plan-phase commit b7d1528c8 (canonical anchor, spec.md §A.3 row 추가됨)
#   - 본 SPEC ID를 commit subject에 포함하는 commit만 평가 대상
SPEC_COMMITS=$(git log --grep='SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001' \
  --format=%H b7d1528c8..HEAD)
for c in $SPEC_COMMITS; do
  git show "$c" -- 'internal/template/templates/**/*.md' \
    | grep -E '^\+' \
    | grep -iE 'predecessor|선행|earlier|prior|generic'
done | wc -l
```
**Expected**: 본 SPEC cleanup commits에서 유사 substitution vocabulary 사용 count ≥ 10 (predecessor와 일관)

**Range rationale** (D-003 amendment): `HEAD~N..HEAD` 하드코딩은 race-absorbed parallel session commits에 fragile. SPEC-scoped `git log --grep` 선별은 (a) plan-phase anchor `b7d1528c8` 이후 commit만 검토하고, (b) commit subject에 본 SPEC ID를 포함하는 commit만 평가하여 (c) parallel session의 무관 commit (예: TEST-REFACTOR-001 M3/M4/M5)을 자동 제외함.

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

**Command** (SPEC-scoped attribution range — D-003 amendment):
```bash
find .github/workflows/ -name '*.yml' -o -name '*.yaml' | xargs grep -l 'go test' | head -3

# 본 SPEC cleanup commits 범위에서 workflow 변경 inspection (race-absorbed commits 제외):
#   - attribution base: plan-phase commit b7d1528c8
SPEC_COMMITS=$(git log --grep='SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001' \
  --format=%H b7d1528c8..HEAD)
for c in $SPEC_COMMITS; do
  git show "$c" -- '.github/workflows/'
done | head -50

# Docstring cross-ref 존재 확인
grep -c 'feedback_template_internal_content_isolation' .github/workflows/*.yml 2>/dev/null || echo "0"
```
**Expected**: 영향 받은 workflow 파일 식별 + `feedback_template_internal_content_isolation` cross-ref count ≥ 1

**Range rationale**: D-003 amendment — `HEAD~6..HEAD` 하드코딩 대신 SPEC-scoped `git log --grep` 사용. 본 SPEC commit 이외 parallel session commit (e.g., TEST-REFACTOR-001)은 평가 대상에서 자동 제외.

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

**Command** (SPEC-scoped attribution range — D-003 amendment):
```bash
# M6 audit에서 발견된 경우만 적용; 발견 없으면 trivially PASS
git log --oneline -10 --all -- '.gitignore' | head -5

# 본 SPEC commit 범위에서 .gitignore 변경 확인 (race-absorbed 제외):
SPEC_COMMITS=$(git log --grep='SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001' \
  --format=%H b7d1528c8..HEAD)
CHANGE_FOUND=0
for c in $SPEC_COMMITS; do
  if git show --stat "$c" -- '.gitignore' | grep -q '.gitignore'; then
    CHANGE_FOUND=1
    git show "$c" -- '.gitignore'
  fi
done
[ "$CHANGE_FOUND" -eq 0 ] && echo "no gitignore changes (no maintainer-only files found in templates)"
```
**Expected**: 발견 0건 시 .gitignore 변경 없음 (PASS); 발견 N건 시 본 SPEC commit log에 `git rm` 흔적 + .gitignore에 가드 entry 추가

**Range rationale**: D-003 amendment — `HEAD~6..HEAD` 하드코딩 대신 SPEC-scoped `git log --grep`. parallel session의 무관 .gitignore 변경 (만약 존재) 자동 제외.

**Verification owner**: manager-develop (M6 종료, 조건부)

### AC-TII-011 — Anti-pattern enforcement (§25-scoped self-check)

**Command** (§25 section-scoped via awk range — D-002 amendment):
```bash
# 본 SPEC 산출물 자체 (3 plan-phase + 추후 run/sync) 가 SPEC ID 또는 REQ token을 인용할 때는 본인의 ID/REQ만 허용
# CLAUDE.local.md §25 본문에서 다른 SPEC ID 인용 금지 (design.md §D.D4 doctrine)
# awk 범위: §25 헤더부터 §26 헤더 또는 `---` Status footer 또는 EOF 까지
awk '/^## 25\. Template Internal-Content Isolation/,/^## 26\.|^---$/' CLAUDE.local.md \
  | grep -oE 'SPEC-V3R6-[A-Z0-9-]+-[0-9]+' \
  | sort -u \
  | grep -v 'TEMPLATE-INTERNAL-ISOLATION-001'
```
**Expected**: empty stdout (§25 본문에 본 SPEC ID `TEMPLATE-INTERNAL-ISOLATION-001` 외 다른 SPEC ID 리터럴 부재)

**Awk range 의미론** (D-002 amendment — 3 edge cases):

| 케이스 | §25 존재 | §26 존재 | 동작 | 결과 |
|--------|---------|---------|------|------|
| 1 (post-run normal) | yes | yes (or following section) | awk이 §25 본문을 §26 헤더까지 capture | §25 본문만 grep |
| 2 (file ends after §25) | yes | no | awk이 EOF 또는 `^---$` Status footer 까지 capture | §25 본문만 grep |
| 3 (pre-run state, §25 미작성) | no | (n/a) | awk이 빈 range 출력 | grep input empty → stdout empty → AC-TII-011 vacuously PASSES |

**Scope rationale** (D-002 amendment): CLAUDE.local.md §17/§18/§21/§23/§24는 현재 5개의 다른 SPEC-V3R6 리터럴을 정당하게 인용함 (AGENT-TEAM-REBUILD-001, HARNESS-RENAME-001, LEGACY-CLEANUP-001, SEQ-THINKING-RETIRE-001, UPDATE-NAMESPACE-PROTECT-001). 전체 파일 grep은 deterministic FAIL이므로 §25 section-scoped로 제한. doctrine 적용 범위는 본 SPEC이 신규 작성하는 §25 본문에 한정 (REQ-TII-013 → design.md §D.D4 → AC-TII-011).

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
| v0.1.1 | 2026-05-25 | manager-spec | iter-1 amendment — 4 SHOULD-FIX 해소: (D-001) AC-TII-001 grep 5-class (4-class prose + commit sha 40-char/7-8-char-space) + false-positive risk note; (D-002) AC-TII-011 §25-scoped awk range + 3 edge cases + scope rationale (other §17/18/21/23/24 SPEC ID 인용은 정당); (D-003) AC-TII-002/008/010 `HEAD~N..HEAD` 하드코딩 제거 → SPEC-scoped `git log --grep` attribution range (plan-phase anchor `b7d1528c8`); (D-005) terminology canonicalization 적용 범위 확인 (acceptance.md 본문 영향 없음 — spec.md/design.md/research.md에서 처리). REQ/AC count 불변 (13 REQ / 12 AC). Verification command 논리적 정확성 회복. |
