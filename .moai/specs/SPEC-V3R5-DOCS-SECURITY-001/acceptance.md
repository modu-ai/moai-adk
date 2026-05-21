# Acceptance Criteria — SPEC-V3R5-DOCS-SECURITY-001

각 AC 는 **binary** 검증 가능 (PASS/FAIL 명확). 모두 grep / find / script exit code 기반.

## AC-DSEC-001 — 4-locale × 4 페이지 = 16 markdown files 작성 완료

**REQ coverage**: REQ-DSEC-001, REQ-DSEC-002, REQ-DSEC-003, REQ-DSEC-007

### Given

- main worktree HEAD (run-phase 진입 시점)
- `docs-site/content/{ko,en,ja,zh}/` 구조 존재

### When

manager-develop 이 M1 + M2 완료 보고

### Then

다음 16 markdown files 가 모두 존재한다:

```bash
# 신규 4 files (security-notes.md 4-locale)
test -f docs-site/content/ko/advanced/security-notes.md
test -f docs-site/content/en/advanced/security-notes.md
test -f docs-site/content/ja/advanced/security-notes.md
test -f docs-site/content/zh/advanced/security-notes.md

# 패치 12 files (3 페이지 × 4 locales)
test -f docs-site/content/ko/advanced/settings-json.md
test -f docs-site/content/en/advanced/settings-json.md
test -f docs-site/content/ja/advanced/settings-json.md
test -f docs-site/content/zh/advanced/settings-json.md
test -f docs-site/content/ko/getting-started/update.md
test -f docs-site/content/en/getting-started/update.md
test -f docs-site/content/ja/getting-started/update.md
test -f docs-site/content/zh/getting-started/update.md
test -f docs-site/content/ko/multi-llm/cg-mode.md
test -f docs-site/content/en/multi-llm/cg-mode.md
test -f docs-site/content/ja/multi-llm/cg-mode.md
test -f docs-site/content/zh/multi-llm/cg-mode.md
```

### Binary Verification

```bash
find docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md 2>/dev/null | wc -l
# Expected: 4

# 패치된 3 페이지가 SECURITY-CRIT-001 cross-reference 포함 검증
for locale in ko en ja zh; do
  grep -q 'security-notes' docs-site/content/${locale}/advanced/settings-json.md && \
  grep -q 'security-notes' docs-site/content/${locale}/getting-started/update.md && \
  grep -q 'security-notes' docs-site/content/${locale}/multi-llm/cg-mode.md || echo "FAIL: ${locale}"
done
# Expected: no FAIL output
```

### PASS Criterion

- `find` 결과 == 4
- 패치 페이지 cross-reference grep 모두 match (12 grep × 1 match)

---

## AC-DSEC-002 — `scripts/docs-i18n-check.sh` PASS

**REQ coverage**: REQ-DSEC-007, REQ-DSEC-008

### Given

- M2 완료 (4-locale 동기 작성됨)
- `scripts/docs-i18n-check.sh` 실행 가능

### When

`scripts/docs-i18n-check.sh` 실행

### Then

Exit code 0

### Binary Verification

```bash
bash scripts/docs-i18n-check.sh
echo "exit=$?"
# Expected: exit=0
```

### PASS Criterion

`exit=0`

---

## AC-DSEC-003 — 금지 URL grep 0 matches

**REQ coverage**: REQ-DSEC-009 (i18n-rules §17.1)

### Given

- M2 완료
- 금지 URL 패턴: `docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` (오타)

### When

본 SPEC 영향 페이지 4 종 × 4 locale = 16 files 대상 grep

### Then

0 matches

### Binary Verification

```bash
grep -rE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr' \
  docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md \
  docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md \
  docs-site/content/{ko,en,ja,zh}/getting-started/update.md \
  docs-site/content/{ko,en,ja,zh}/multi-llm/cg-mode.md 2>/dev/null
echo "exit=$?"
# Expected: exit=1 (grep 0 matches → exit 1)
```

### PASS Criterion

`exit=1` (grep 의 0-match exit code)

---

## AC-DSEC-004 — Mermaid TD-only (LR/BR 0 matches)

**REQ coverage**: REQ-DSEC-009 (i18n-rules §17.2)

### Given

- M2 완료
- 금지 Mermaid 방향: `flowchart LR`, `graph LR`, `flowchart BR`, `graph BR`

### When

본 SPEC 영향 페이지 16 files 대상 grep

### Then

0 matches

### Binary Verification

```bash
grep -rE 'flowchart LR|graph LR|flowchart BR|graph BR' \
  docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md \
  docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md \
  docs-site/content/{ko,en,ja,zh}/getting-started/update.md \
  docs-site/content/{ko,en,ja,zh}/multi-llm/cg-mode.md 2>/dev/null
echo "exit=$?"
# Expected: exit=1
```

### PASS Criterion

`exit=1` (0 matches)

---

## AC-DSEC-005 — CHANGELOG SECURITY 섹션 추가

**REQ coverage**: REQ-DSEC-011

### Given

- M3 완료
- CHANGELOG.md 갱신됨

### When

CHANGELOG.md `[Unreleased]` 섹션 grep

### Then

다음 키워드 중 최소 1 match + CWE 키워드 ≥3 match (CWE-732/214/345):

### Binary Verification

```bash
# SECURITY 헤더 또는 Security 키워드 존재
grep -E '^## \[Unreleased\]|### Security|## Security|SECURITY' CHANGELOG.md | head -3
# Expected: ≥1 match

# CWE 키워드 ≥3 (CWE-732, CWE-214, CWE-345)
grep -cE 'CWE-732|CWE-214|CWE-345' CHANGELOG.md
# Expected: ≥3

# SECURITY-CRIT-001 commit hash 중 최소 1개 포함
grep -E 'b48bd86cb|10776c4b8|ee1335282|b4e7115cb|03a2552a2' CHANGELOG.md | head -1
# Expected: ≥1 match
```

### PASS Criterion

- SECURITY/Security 키워드 ≥1 match
- CWE 키워드 count ≥3
- commit hash ≥1 match

---

## AC-DSEC-006 — 본문 이모지 0 matches

**REQ coverage**: REQ-DSEC-010

### Given

- M2 완료
- Unicode 이모지 범위: `U+1F300` ~ `U+1F9FF` (Misc Symbols And Pictographs + Supplemental Symbols + Symbols And Pictographs Extended-A)

### When

본 SPEC 영향 16 markdown files 대상 grep

### Then

0 matches

### Binary Verification

```bash
grep -P '[\x{1F300}-\x{1F9FF}]' \
  docs-site/content/{ko,en,ja,zh}/advanced/security-notes.md \
  docs-site/content/{ko,en,ja,zh}/advanced/settings-json.md \
  docs-site/content/{ko,en,ja,zh}/getting-started/update.md \
  docs-site/content/{ko,en,ja,zh}/multi-llm/cg-mode.md 2>/dev/null
echo "exit=$?"
# Expected: exit=1 (0 matches)
```

### PASS Criterion

`exit=1` (0 matches)

### Note

`🗿` (U+1FAA8 Moai) 가 commit message trailer 에는 허용되지만, **본문에는 금지**. 본 AC 는 docs-site 본문에만 적용 (commit message 는 별도).

---

## AC-DSEC-007 — `moai spec lint --strict` 본 SPEC 0 NEW 위반

**REQ coverage**: REQ-DSEC-012

### Given

- 본 SPEC 디렉토리 `.moai/specs/SPEC-V3R5-DOCS-SECURITY-001/` 의 3 artifacts (spec.md, plan.md, acceptance.md) 작성 완료
- Frontmatter 12-field canonical schema 준수
- spec.md §5 에 `### N.M Out of Scope` h3 sub-section 존재 (본 SPEC 의 경우 §3 Non-Goals 가 의미적 동등 — spec-lint heading 규약 충족 위해 §5.X 또는 §3.X h3 sub-section 검증)

### When

`go run ./cmd/moai spec lint --strict` 실행 후 본 SPEC 출력 필터

### Then

0 NEW violations (pre-existing baseline 외)

### Binary Verification

```bash
go run ./cmd/moai spec lint --strict 2>&1 | grep -E 'SPEC-V3R5-DOCS-SECURITY-001'
echo "exit_of_grep=$?"
# Expected: exit_of_grep=1 (no violations referenced)

# 또는 명시적:
go run ./cmd/moai spec lint --strict 2>&1 | grep 'SPEC-V3R5-DOCS-SECURITY-001' | grep -E 'ERROR|FAIL' | wc -l
# Expected: 0
```

### PASS Criterion

- grep 결과 본 SPEC 관련 ERROR/FAIL count == 0
- Frontmatter 12 fields 모두 존재 (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + tier 필드 (optional)
- Out of Scope h3 sub-section 존재 (Non-Goals §3 의 명시 또는 §5.X h3)

---

## Summary Matrix

| AC ID | Description | Verification Type | Expected |
|-------|-------------|-------------------|----------|
| AC-DSEC-001 | 4-locale × 4 페이지 = 16 files 작성 + cross-reference | find + grep | 4 + 12 matches |
| AC-DSEC-002 | i18n script PASS | script exit code | 0 |
| AC-DSEC-003 | 금지 URL 0 matches | grep | exit 1 |
| AC-DSEC-004 | Mermaid TD only | grep | exit 1 |
| AC-DSEC-005 | CHANGELOG SECURITY 항목 | grep count | ≥1 + ≥3 CWE + ≥1 hash |
| AC-DSEC-006 | 본문 이모지 0 | grep -P unicode | exit 1 |
| AC-DSEC-007 | spec-lint 0 NEW 위반 | grep + count | 0 ERROR |

**모든 7 AC PASS 시 run-phase 완료**. PR 생성 시 commit body 에 본 matrix 의 실측 결과 포함 의무.

## Definition of Done

- [ ] AC-DSEC-001 ~ AC-DSEC-007 모두 binary PASS
- [ ] 4-locale 페이지 hugo build dry-run 통과 (선택적 검증)
- [ ] cross-reference anchor 작동 확인 (`#cwe-732`, `#cwe-214`, `#cwe-345` slug)
- [ ] SPEC frontmatter version 0.1.0 → 0.2.0 (run-phase 완료 시 manager-develop 이 갱신)
- [ ] SPEC status `draft → implemented` (run-phase 완료 시)
- [ ] Conventional Commits + `🗿 MoAI` trailer
- [ ] Late-Branch 정책: main 직진 commit 후 PR 시점 branch 분리
