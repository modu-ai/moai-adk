# Acceptance Criteria — SPEC-V3R6-DOCS-CMD-CATALOG-001

## 1. Binary Acceptance Criteria (8건)

각 AC는 단일 verification command + expected output 으로 PASS/FAIL 결정 가능.

### AC-DCC-001 (R1 페이지 삭제 검증)

**Given**: SPEC-V3R6-DOCS-CMD-CATALOG-001 run-phase 완료 직후
**When**: `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md` 4 파일 존재 여부 확인
**Then**: 4 파일 모두 부재 (4 ENOENT)

**Verification**:
```bash
ls docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md 2>&1 | grep -c "No such file"
```
**Expected output**: `4` (정확히 4개 locale에서 부재)

**REQ↔AC mapping**: REQ-DCC-001 (delete 4 files)

---

### AC-DCC-002 (R1 _meta.yaml entry 제거 검증)

**Given**: AC-DCC-001 PASS
**When**: `workflow-commands/_meta.yaml` 4 파일에서 `moai-db` 키 검색
**Then**: 4 파일 모두 `moai-db` 키 부재

**Verification**:
```bash
grep -l "moai-db" docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml 2>&1
```
**Expected output**: 빈 출력 (0 매치, exit 1) — 즉 `grep -L "moai-db" ...` 가 4 파일 모두 listing

**Alternate verification** (positive form):
```bash
for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml; do grep -q "moai-db" "$f" && echo "FAIL: $f" || echo "OK: $f"; done | grep -c "OK:"
```
**Expected**: `4`

**REQ↔AC mapping**: REQ-DCC-002 (_meta.yaml entry 제거)

---

### AC-DCC-003 (R1 main.yaml reference 제거 검증)

**Given**: AC-DCC-002 PASS
**When**: `docs-site/data/menu/main.yaml` 에서 `/workflow-commands/moai-db` reference 검색
**Then**: 0 매치

**Verification**:
```bash
grep -c "/workflow-commands/moai-db" docs-site/data/menu/main.yaml
```
**Expected output**: `0`

**REQ↔AC mapping**: REQ-DCC-003 (main.yaml reference 제거)

---

### AC-DCC-004 (R2 페이지 + meta + main.yaml 제거 검증, 통합)

**Given**: SPEC run-phase 완료
**When**: `moai-github.md` 4 파일 + `_meta.yaml` 4 파일 + `main.yaml` 단일 파일에서 github 흔적 검색
**Then**: 모두 부재

**Verification**:
```bash
# (a) 4 files 부재
files_absent=$(ls docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md 2>&1 | grep -c "No such file")
# (b) _meta.yaml 부재
meta_absent=$(for f in docs-site/content/{ko,en,ja,zh}/utility-commands/_meta.yaml; do grep -q "moai-github" "$f" || echo "ok"; done | wc -l | tr -d ' ')
# (c) main.yaml 부재
main_absent=$(grep -c "/utility-commands/moai-github" docs-site/data/menu/main.yaml)
echo "files_absent=$files_absent meta_absent=$meta_absent main_absent=$main_absent"
```
**Expected output**: `files_absent=4 meta_absent=4 main_absent=0`

**REQ↔AC mapping**: REQ-DCC-004 (delete 4 files), REQ-DCC-005 (_meta.yaml), REQ-DCC-006 (main.yaml)

---

### AC-DCC-005 (G1 harness 페이지 4-locale 신설 검증)

**Given**: SPEC run-phase 완료
**When**: `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md` 4 파일 존재 + frontmatter 검증 + 본문 핵심 키워드 검증
**Then**: 4 파일 모두 존재, `weight: 55`, `draft: false`, `title` 존재, 본문에 `status`/`apply`/`rollback`/`disable` 4 verbs 모두 등장

**Verification**:
```bash
# (a) 4 files 존재
files_exist=$(ls docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md 2>&1 | grep -v "No such file" | wc -l | tr -d ' ')
# (b) weight: 55 4-locale 모두
weight_ok=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md; do grep -q "^weight: 55$" "$f" && echo "ok"; done | wc -l | tr -d ' ')
# (c) draft: false 또는 미명시 (Hugo default)
draft_ok=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md; do grep -q "^draft: true" "$f" && echo "fail" || echo "ok"; done | grep -c "ok")
# (d) 4 verbs 본문 존재
verbs_ok=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md; do grep -q "status" "$f" && grep -q "apply" "$f" && grep -q "rollback" "$f" && grep -q "disable" "$f" && echo "ok"; done | wc -l | tr -d ' ')
echo "files_exist=$files_exist weight_ok=$weight_ok draft_ok=$draft_ok verbs_ok=$verbs_ok"
```
**Expected output**: `files_exist=4 weight_ok=4 draft_ok=4 verbs_ok=4`

**REQ↔AC mapping**: REQ-DCC-007 (harness 4 files 신설 + 본문 contract)

---

### AC-DCC-006 (G1 _meta.yaml + main.yaml 추가 검증)

**Given**: AC-DCC-005 PASS
**When**: `workflow-commands/_meta.yaml` 4 파일에서 `moai-harness` 키 + `main.yaml` 에서 `/workflow-commands/moai-harness` reference 검색
**Then**: 4 _meta + 1 main 추가됨

**Verification**:
```bash
meta_added=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml; do grep -q "moai-harness" "$f" && echo "ok"; done | wc -l | tr -d ' ')
main_added=$(grep -c "/workflow-commands/moai-harness" docs-site/data/menu/main.yaml)
echo "meta_added=$meta_added main_added=$main_added"
```
**Expected output**: `meta_added=4 main_added=1` (main.yaml은 단일 entry로 4-locale 라벨 포함이므로 `ref:` 라인 1개)

**REQ↔AC mapping**: REQ-DCC-008 (_meta + main.yaml 추가)

---

### AC-DCC-007 (G2 gate 페이지 + meta + main.yaml 통합 검증 + i18n parity)

**Given**: SPEC run-phase 완료
**When**: gate 4 페이지 신설 + meta/main 추가 + 4-locale parity 검증
**Then**: 모두 통과

**Verification**:
```bash
# (a) 4 files 존재
files_exist=$(ls docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>&1 | grep -v "No such file" | wc -l | tr -d ' ')
# (b) weight: 15 4-locale
weight_ok=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md; do grep -q "^weight: 15$" "$f" && echo "ok"; done | wc -l | tr -d ' ')
# (c) 핵심 키워드 (--fix, --staged, lint/format/type-check/test)
content_ok=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md; do grep -q -- "--fix" "$f" && grep -q -- "--staged" "$f" && grep -q "lint" "$f" && grep -q "test" "$f" && echo "ok"; done | wc -l | tr -d ' ')
# (d) _meta.yaml 4 + main.yaml 1
meta_added=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/_meta.yaml; do grep -q "moai-gate" "$f" && echo "ok"; done | wc -l | tr -d ' ')
main_added=$(grep -c "/quality-commands/moai-gate" docs-site/data/menu/main.yaml)
echo "files_exist=$files_exist weight_ok=$weight_ok content_ok=$content_ok meta_added=$meta_added main_added=$main_added"
```
**Expected output**: `files_exist=4 weight_ok=4 content_ok=4 meta_added=4 main_added=1`

**REQ↔AC mapping**: REQ-DCC-009 (gate 4 files 신설 + 본문 contract), REQ-DCC-010 (_meta + main.yaml 추가)

---

### AC-DCC-008 (Cross-cutting: forbidden URL/Mermaid LR/emoji/draft 위반 0 + 4-locale parity + Hugo build)

**Given**: 모든 ACs (001-007) PASS
**When**: 신규/수정 파일에서 정책 위반 검증 + Hugo 로컬 빌드 실행
**Then**: 위반 0건 + `hugo --gc --minify --buildDrafts=false` exit 0

**Verification**:
```bash
# (a) Forbidden URL 0 (canonical은 adk.mo.ai.kr만 허용)
forbidden_url=$(grep -rEn "https?://[a-z0-9.-]+" docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>/dev/null | grep -v "adk\.mo\.ai\.kr" | grep -v "github\.com/modu-ai/moai-adk" | wc -l | tr -d ' ')
# (b) Mermaid LR/BT/RL 0
mermaid_violation=$(grep -rEn "^(graph|flowchart) (LR|BT|RL)" docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>/dev/null | wc -l | tr -d ' ')
# (c) Emoji 0 (대표 패턴 🤖, 🎯, 🚀, 📋 등 검사)
emoji_violation=$(grep -rEn "[🤖🎯🚀📋✅❌⚡🔧🛠️📝]" docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>/dev/null | wc -l | tr -d ' ')
# (d) draft: true 0 (신규 페이지)
draft_violation=$(grep -l "^draft: true" docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>/dev/null | wc -l | tr -d ' ')
# (e) 4-locale parity: 신규 페이지 8 + 삭제 페이지 8 = 정확히 8 file create + 8 file delete = clean state
# (f) Hugo 빌드
cd docs-site && hugo --gc --minify --buildDrafts=false 2>&1 | tail -5
hugo_exit=$?
cd ..
echo "forbidden_url=$forbidden_url mermaid_violation=$mermaid_violation emoji_violation=$emoji_violation draft_violation=$draft_violation hugo_exit=$hugo_exit"
```
**Expected output**: `forbidden_url=0 mermaid_violation=0 emoji_violation=0 draft_violation=0 hugo_exit=0`

**REQ↔AC mapping**: REQ-DCC-011 (4-locale parity), REQ-DCC-012 (forbidden URL/Mermaid/emoji/draft 금지) + Hugo build 사전 검증 (EC-DCC-006)

---

## 2. REQ ↔ AC Traceability Matrix

| Requirement | AC | Coverage |
|-------------|-----|----------|
| REQ-DCC-001 (R1 4 files delete) | AC-DCC-001 | Full |
| REQ-DCC-002 (R1 _meta.yaml) | AC-DCC-002 | Full |
| REQ-DCC-003 (R1 main.yaml) | AC-DCC-003 | Full |
| REQ-DCC-004 (R2 4 files delete) | AC-DCC-004(a) | Full |
| REQ-DCC-005 (R2 _meta.yaml) | AC-DCC-004(b) | Full |
| REQ-DCC-006 (R2 main.yaml) | AC-DCC-004(c) | Full |
| REQ-DCC-007 (G1 harness 4 files 신설 + 본문) | AC-DCC-005 | Full |
| REQ-DCC-008 (G1 _meta + main 추가) | AC-DCC-006 | Full |
| REQ-DCC-009 (G2 gate 4 files 신설 + 본문) | AC-DCC-007(a-c) | Full |
| REQ-DCC-010 (G2 _meta + main 추가) | AC-DCC-007(d-e) | Full |
| REQ-DCC-011 (4-locale parity) | AC-DCC-008(e) + AC-DCC-001/004/005/007 (4-count) | Full |
| REQ-DCC-012 (forbidden URL/Mermaid/emoji/draft) | AC-DCC-008(a-d) | Full |

**Coverage**: 12 REQs × 8 ACs, 100% (모든 REQ가 최소 1개 AC에 매핑됨).

---

## 3. Edge Case Coverage

| Edge Case | Covered By | Notes |
|-----------|------------|-------|
| EC-DCC-001 (_meta.yaml 4-locale 동일 키) | AC-DCC-002, AC-DCC-004(b) | 4 file × 1 entry × 4 locale = 4 deletion 카운트로 검증 |
| EC-DCC-002 (main.yaml 단일 entry 4-locale 라벨) | AC-DCC-003, AC-DCC-004(c), AC-DCC-006, AC-DCC-007(e) | `grep -c "/category/page" main.yaml` 결과 0 (삭제) 또는 1 (추가) 로 검증 |
| EC-DCC-003 (db 삭제 후 sync=50 잔존) | implicit (weight 검증 X — 본 SPEC scope 외 per §3.5) | 기존 weight 충돌 해소는 본 SPEC 부수효과로만 발생 |
| EC-DCC-004 (github 삭제 후 utility weight 정렬) | implicit | 동일 |
| EC-DCC-005 (gate weight: 15 review 앞) | AC-DCC-007(b) (weight: 15 검증) | review와 mental model 부합은 본문 §Difference from Other Workflows 비교 표로 보강 |
| EC-DCC-006 (Vercel 빌드 실패 위험) | AC-DCC-008(f) (Hugo 로컬 빌드) | 로컬 `hugo --gc --minify` exit 0 가 Vercel 빌드 PASS의 강한 proxy |

---

## 4. Quality Gate Criteria

본 SPEC run-phase 완료 시 다음 quality gate 모두 통과해야 함:

### 4.1 Mandatory Quality Gates

- [ ] All 8 ACs PASS (binary)
- [ ] `hugo --gc --minify --buildDrafts=false` exit 0 (AC-DCC-008(f))
- [ ] `git status` 깨끗 (Section 1.4 PRESERVE list 보호 검증 — 본 SPEC scope 외 파일 변경 없음)
- [ ] Working tree dirty list 변동 0 (spec.md §6.2 PRESERVE 정확히 유지)

### 4.2 Self-Verification Deliverables (manager-develop 보고 의무)

run-phase 완료 시 manager-develop 또는 orchestrator가 보고:

- [ ] AC matrix 8/8 PASS (위 verification commands 실행 결과 + 예상 output 비교)
- [ ] Files affected: 정확히 **8 delete** (db × 4-locale + github × 4-locale) + 정확히 **8 create** (harness × 4-locale + gate × 4-locale) + **12 _meta.yaml file 모두 수정** (workflow × 4-locale 각 2 entry-change [db remove + harness add], utility × 4-locale 각 1 entry-change [github remove], quality × 4-locale 각 1 entry-change [gate add] = 합계 16 entry-changes across 12 _meta files) + **1 main.yaml 수정** (4 block-change — db remove / github remove / harness add / gate add)
- [ ] 신규 페이지 4-locale parity: ko/en/ja/zh 본문 구조 동일 (heading 수, 섹션 순서)
- [ ] Cross-platform 무관 (콘텐츠 변경만)
- [ ] C-HRA-008 n/a (no harness/hook Go files touched)
- [ ] Lint NEW=0 (md 파일 lint 미적용 — Hugo 빌드 PASS로 대체)

### 4.3 Definition of Done

본 SPEC 은 다음 모두 만족 시 DONE:

1. 8/8 ACs PASS
2. Hugo 로컬 빌드 exit 0
3. Working tree PRESERVE list 보호 검증 PASS
4. SPEC frontmatter `status: draft` → `status: implemented` 갱신 (v0.1.0 → v0.2.0)
5. progress.md 생성 (run-phase 산출물)
6. Self-verification report orchestrator로 반환

---

## 5. Test Execution Plan

run-phase 완료 직후 다음 명령으로 8 ACs 일괄 검증 (orchestrator 또는 manager-develop self-verification):

```bash
#!/usr/bin/env bash
set -e
echo "=== AC-DCC-001 (R1 4 files delete) ==="
ls docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md 2>&1 | grep -c "No such file"  # expect 4

echo "=== AC-DCC-002 (R1 _meta.yaml) ==="
for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml; do grep -q "moai-db" "$f" && echo "FAIL: $f" || echo "OK: $f"; done | grep -c "OK:"  # expect 4

echo "=== AC-DCC-003 (R1 main.yaml) ==="
grep -c "/workflow-commands/moai-db" docs-site/data/menu/main.yaml  # expect 0

echo "=== AC-DCC-004 (R2 통합) ==="
files=$(ls docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md 2>&1 | grep -c "No such file")
meta=$(for f in docs-site/content/{ko,en,ja,zh}/utility-commands/_meta.yaml; do grep -q "moai-github" "$f" || echo "ok"; done | wc -l | tr -d ' ')
main=$(grep -c "/utility-commands/moai-github" docs-site/data/menu/main.yaml)
echo "files=$files meta=$meta main=$main"  # expect files=4 meta=4 main=0

echo "=== AC-DCC-005 (G1 harness) ==="
files=$(ls docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md 2>&1 | grep -v "No such file" | wc -l | tr -d ' ')
weight=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md; do grep -q "^weight: 55$" "$f" && echo "ok"; done | wc -l | tr -d ' ')
verbs=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md; do grep -q "status" "$f" && grep -q "apply" "$f" && grep -q "rollback" "$f" && grep -q "disable" "$f" && echo "ok"; done | wc -l | tr -d ' ')
echo "files=$files weight=$weight verbs=$verbs"  # expect files=4 weight=4 verbs=4

echo "=== AC-DCC-006 (G1 meta + main) ==="
meta=$(for f in docs-site/content/{ko,en,ja,zh}/workflow-commands/_meta.yaml; do grep -q "moai-harness" "$f" && echo "ok"; done | wc -l | tr -d ' ')
main=$(grep -c "/workflow-commands/moai-harness" docs-site/data/menu/main.yaml)
echo "meta=$meta main=$main"  # expect meta=4 main=1

echo "=== AC-DCC-007 (G2 gate 통합) ==="
files=$(ls docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md 2>&1 | grep -v "No such file" | wc -l | tr -d ' ')
weight=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md; do grep -q "^weight: 15$" "$f" && echo "ok"; done | wc -l | tr -d ' ')
content=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md; do grep -q -- "--fix" "$f" && grep -q -- "--staged" "$f" && grep -q "lint" "$f" && grep -q "test" "$f" && echo "ok"; done | wc -l | tr -d ' ')
meta=$(for f in docs-site/content/{ko,en,ja,zh}/quality-commands/_meta.yaml; do grep -q "moai-gate" "$f" && echo "ok"; done | wc -l | tr -d ' ')
main=$(grep -c "/quality-commands/moai-gate" docs-site/data/menu/main.yaml)
echo "files=$files weight=$weight content=$content meta=$meta main=$main"  # expect files=4 weight=4 content=4 meta=4 main=1

echo "=== AC-DCC-008 (cross-cutting + Hugo build) ==="
# Bash brace expansion으로 8 path를 array로 수집 (eval 없이)
target_files=(
  docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md
  docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md
)
forbidden_url=$(grep -rEn "https?://" "${target_files[@]}" 2>/dev/null | grep -v "adk\.mo\.ai\.kr" | grep -v "github\.com/modu-ai/moai-adk" | wc -l | tr -d ' ')
mermaid=$(grep -rEn "^(graph|flowchart) (LR|BT|RL)" "${target_files[@]}" 2>/dev/null | wc -l | tr -d ' ')
draft=$(grep -l "^draft: true" "${target_files[@]}" 2>/dev/null | wc -l | tr -d ' ')
echo "forbidden_url=$forbidden_url mermaid=$mermaid draft=$draft"
echo "--- Hugo build ---"
cd docs-site && hugo --gc --minify --buildDrafts=false 2>&1 | tail -3
echo "hugo_exit=$?"
```

**Expected aggregate output**: 모든 카운터가 spec 의 expected output과 일치 + hugo_exit=0
