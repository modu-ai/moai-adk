# Implementation Plan — SPEC-V3R6-DOCS-CMD-CATALOG-001

## 1. Tier Classification

**Tier: Tier S (LEAN minimal form)**

근거 (LEAN workflow rule + spec-workflow.md § SPEC Complexity Tier):
- 영향 파일 = 8 delete + 8 create + 12 _meta.yaml modify + 1 main.yaml modify = **29 files affected** (Tier S 임계 <5 초과)
- 그러나 **변경 성격이 일률적이고 mechanical** (delete은 단순 rm, create는 template 복제 + locale 번역, _meta.yaml/main.yaml은 entry 추가/삭제 단순 패턴)
- **LOC**: 신규 페이지 4 locale × 약 200 lines = ~800 LOC, 본문 작성 부담은 있으나 복잡한 로직 없음
- **Tier S로 분류** 가능 근거:
  - 콘텐츠 변경만 (코드 0)
  - 테스트 신규 작성 0 (Hugo 빌드 + grep verification으로 충분)
  - Cross-platform 무관
  - C-HRA-008 / subagent boundary 무관
- 그러나 4-locale × 4 files = 16 file 영향 + plan-auditor 호출 권장 (이미 spec-workflow.md § LEAN Tier S 임계 근사) → **Tier S 분류 채택하되 plan-auditor 호출 by orchestrator 권장**

**적용 형식**: minimal form (Section A-E 일부 약식) + plan-auditor 호출 옵션 (orchestrator 결정)

---

## 2. Pre-flight 결과 요약 (spec.md §1.4 인용)

(이미 spec.md §1.4 에 detail 명시됨. plan은 핵심 4개 사실만 재인용)

1. **R1 db**: 4 files × 3,300 bytes 동일 + cmd/moai/`grep` 0 매치 → 완전 false promise 확정
2. **R2 github**: 4 files locale별 본문 번역 + `.claude/commands/moai/github.md` 부재 + dev-only 97/98 prefix도 `.claude/commands/`에 부재 → 완전 false promise 확정
3. **G1 harness**: skill v2.0.0 (4 verbs / Tier-4 / 5-layer safety) + command 263 bytes 존재 + docs 페이지 4-locale 부재 → 신설 필요
4. **G2 gate**: skill v1.0.0 (lint+format+type-check+test 병렬 <30s) + command 188 bytes 존재 + docs 페이지 4-locale 부재 → 신설 필요

**weight 자동 결정 (OQ4 default)**:
- G1 weight: 55 (workflow-commands에서 sync=50 다음 자연 정렬)
- G2 weight: 15 (quality-commands에서 review=20 앞, pre-commit gate mental model)

---

## 3. Section A — Context (위치 + 분기 + 산출물 경로)

- Working directory: `/Users/goos/MoAI/moai-adk-go` (main checkout, no worktree per user policy 2026-05-17 default)
- Current branch: `main`
- HEAD SHA: `a809e0b98` (docs(CLAUDE.local): §23 Local Git Workflows + Hook Setup)
- SPEC artifacts: `.moai/specs/SPEC-V3R6-DOCS-CMD-CATALOG-001/{spec.md, plan.md, acceptance.md}`
- Run-phase 산출물 예정: `progress.md` (run-phase 종료 시 생성)
- 기존 인프라 (PRESERVE 대상): docs-site Hugo 빌드 시스템 (hugo.toml, layouts/, partials/) — 수정 금지
- 변경 대상 (EXTEND 대상): docs-site/content/ko/en/ja/zh/{workflow,quality,utility}-commands/ + docs-site/data/menu/main.yaml

---

## 4. Section B — Known Issues (자동 주입)

본 SPEC 도메인 (docs-site)에 적용 가능한 known issues:

| 카테고리 | 적용 여부 | 적용 시 가이드 |
|----------|-----------|----------------|
| B1 Cross-platform build tags | n/a | Go 코드 0 변경 |
| B2 Cross-SPEC 정책 충돌 | **확인 필요** | 진행 중 SPEC (CATALOG-SSOT-001 / ABSORB-CLEANUP-001 / HARNESS-RENAME-001 / DOCS-SECURITY-001 등) 과 paths 충돌 점검: docs-site/ vs .claude/agents/ + .moai/config/ + internal/ — **충돌 없음 확인** |
| B3 C-HRA-008 / Subagent boundary | n/a | hook/harness Go 파일 0 변경 |
| B4 Frontmatter canonical schema | **적용** | SPEC frontmatter는 canonical 12-field schema 의무 (`created:` / `updated:` / `tags:` / `title:` / `phase:` / `module:` / `lifecycle:`) — `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT. spec.md v0.1.0 작성 시 snake_case alias (`created_at:` / `updated_at:` / `labels:`) 사용 + 4 required field 누락 발견 → plan-auditor iter 1 verdict REVISE → spec.md v0.1.1 iter 1 정정 완료 (canonical 12-field + `tier: S` 명시) |
| B5 CI 3-tier 인지 | **부분 적용** | spec-lint + golangci-lint + Test (per OS) — 본 SPEC은 docs-only이므로 docs-site Vercel 빌드만 영향 + spec-lint 영향 (spec.md/plan.md/acceptance.md heading 규약) |
| B6 spec-lint heading 규약 | **적용** | spec.md `## 3. Out of Scope` 가 아니라 `### 3.X Out of Scope` 패턴 사용 — 본 SPEC spec.md §3에서 `### 3.1` ~ `### 3.8` 8개 sub-section으로 구성 (W3 lessons #19 적용) |
| B7 observer.go capture path | n/a | 비코드 변경 |
| B8 Working tree hygiene | **적용** | spec.md §6.2 PRESERVE list 정확히 enumerate — runtime-managed files (`.moai/harness/usage-log.jsonl`) + 무관 untracked SPEC 디렉토리 (4개 V3R5-001) + docs-site dirty (hugo.toml / layouts/) + .moai/research/ 3건 + internal/hook/.moai/ 절대 변경 금지 |

**docs-site 도메인 특화 known issues**:
- **4-locale 동기화 의무**: ko 원문 → en/ja/zh 번역 시 본문 구조(heading/section count) 동일 유지 (REQ-LCK-1, `.moai/docs/docs-site-i18n-rules.md`)
- **Mermaid TD-only**: harness/gate 본문에 diagram 추가 시 `graph TD` 또는 `flowchart TD` 만 사용 (`LR`/`BT`/`RL` 금지)
- **canonical URL**: 본문 내 cross-link은 `adk.mo.ai.kr` 또는 GitHub repo `github.com/modu-ai/moai-adk` 만 허용
- **Emoji ban**: 신규 페이지 `🤖`, `🎯`, `🚀` 등 emoji 절대 사용 금지 (CWE 보안 정책)
- **Hugo weight 충돌**: spec.md §1.4 weight 분석 결과 신규 weight (55, 15) 는 충돌 없음 확인

---

## 5. Section C — Pre-flight Check List (run-phase 착수 전 의무 검증)

run-phase 위임 받은 manager-develop 또는 orchestrator가 콘텐츠 변경 전 실행:

```bash
# 1. 현재 main HEAD + branch 확인
git log --oneline -1  # expect: a809e0b98 또는 그 이후 (forward only)
git branch --show-current  # expect: main

# 2. 4-locale 디렉토리 존재 확인
ls docs-site/content/{ko,en,ja,zh}/workflow-commands/ | head -5
ls docs-site/content/{ko,en,ja,zh}/quality-commands/ | head -5
ls docs-site/content/{ko,en,ja,zh}/utility-commands/ | head -5

# 3. Hugo 빌드 baseline (변경 전 PASS 확인)
cd docs-site && hugo --gc --minify --buildDrafts=false 2>&1 | tail -3
# expect: "Total in XX ms" exit 0
cd ..

# 4. PRESERVE 대상 파일 정확한 list 출력
git status --short | head -30  # SPEC 시작 시 dirty list snapshot

# 5. 신규 페이지 source 본문 재확인
cat .claude/commands/moai/harness.md  # 263 bytes
cat .claude/commands/moai/gate.md  # 188 bytes
head -100 .claude/skills/moai/workflows/harness.md
head -50 .claude/skills/moai/workflows/gate.md

# 6. Cross-SPEC 충돌 재확인 (진행 중 SPECs paths)
grep -l "docs-site/content" .moai/specs/SPEC-V3R6-*/spec.md 2>/dev/null  # expect: 본 SPEC만 (다른 SPEC 0)
```

---

## 6. Section D — Constraints (DO NOT VIOLATE)

- **PRESERVE list 절대 수정 금지** (spec.md §6.2 enumerate 참조)
- **4-locale 동기화 필수**: ko 작성 후 en/ja/zh 3-locale 동시 적용 (partial 금지 — AC-DCC-008(e))
- **spec-lint NEW 위반 0**: spec.md `### X.Y Out of Scope` h3 sub-section 패턴 준수 (이미 §3에 8개 sub-section 적용)
- **weight 충돌 0 (신규 페이지 한정)**: G1=55, G2=15 — 다른 페이지와 미충돌 검증 완료 (spec.md §1.4)
- **Forbidden URL 0**: canonical `adk.mo.ai.kr` + `github.com/modu-ai/moai-adk` 외 외부 URL 금지
- **Mermaid LR/BT/RL 0 위반**: TD/topdown만 허용
- **Emoji `🤖` `🎯` 등 0**: 신규 페이지 body 검사 (AC-DCC-008(c))
- **삭제 시 메뉴 데이터 파일도 동시 정리**: db/github 삭제 시 _meta.yaml entry + main.yaml block 동시 제거 (orphan link 방지)
- **신규 페이지에 stub disclaimer 금지**: `draft: false` (또는 frontmatter 미명시 — Hugo default false), 본문 완성 의무 (AC-DCC-008(d))
- **사용 의무**: Conventional Commits + `🗿 MoAI` trailer (git commit 시) + `--no-verify` 금지 + `--amend` 금지 + force-push to main 금지

---

## 7. Task Breakdown

### T-DCC-001: R1 db false promise 제거 (REQ-DCC-001/002/003)

**Inputs**: spec.md §1.4 fact-check, 4 db 페이지 list
**Output**: 4 files deleted + 4 _meta.yaml entry 제거 + 1 main.yaml block 제거
**Steps**:
1. `rm docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md` (4 deletions)
2. 4-locale `_meta.yaml` 에서 `"moai-db":` block + `title: "/moai db"` (locale-translated) entry 제거 (Edit tool)
3. `docs-site/data/menu/main.yaml` 에서 `- name:` ~ `ref: /workflow-commands/moai-db` 1 block 제거 (Edit tool, 약 10 lines)

**Verification**: AC-DCC-001, AC-DCC-002, AC-DCC-003

### T-DCC-002: R2 github false promise 제거 (REQ-DCC-004/005/006)

**Inputs**: spec.md §1.4 fact-check, 4 github 페이지 list
**Output**: 4 files deleted + 4 _meta.yaml entry 제거 + 1 main.yaml block 제거
**Steps**:
1. `rm docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md` (4 deletions)
2. 4-locale `_meta.yaml` 에서 `"moai-github":` block 제거
3. `main.yaml` 에서 `- name:` ~ `ref: /utility-commands/moai-github` 1 block 제거

**Verification**: AC-DCC-004

### T-DCC-003: G1 harness 페이지 신설 (REQ-DCC-007/008)

**Inputs**: `.claude/commands/moai/harness.md` (263B) + `.claude/skills/moai/workflows/harness.md` (V3R4 본문) + SPEC-V3R4-HARNESS-001
**Output**: 4 new files + 4 _meta.yaml entry 추가 + 1 main.yaml block 추가
**Steps**:
1. **ko 원문 작성** (`docs-site/content/ko/workflow-commands/moai-harness.md`):
   - Frontmatter: `title: "/moai harness"`, `linkTitle: "/moai harness"`, `description: "Self-Evolving Harness 4-tier 진화 + 5-layer 안전 (status/apply/rollback/disable)"`, `weight: 55`, `geekdocCollapseSection: true`
   - Sections (ko 원문 구조, 약 200 lines):
     - `## 개요` — V3R4 Self-Evolving Harness 개념 (observer + 4-tier ladder + 5-layer safety) 요약
     - `## 명령어 형식` — `/moai harness {status|apply|rollback <YYYY-MM-DD>|disable}`
     - `## 4 verbs 상세`
       - `### status` — 현재 harness 상태 + pending proposals + 7-day rate-limit 윈도우 표시
       - `### apply` — Tier-4 proposal 승인 후 적용 (orchestrator AskUserQuestion gate 필수)
       - `### rollback <YYYY-MM-DD>` — 특정 날짜 snapshot으로 복원
       - `### disable` — harness 학습 일시 중단 (`learning.enabled: false`)
     - `## 4-tier evolution ladder`
       - Tier-1 observation (passive), Tier-2 heuristic (suggest), Tier-3 rule (auto-apply non-frozen), Tier-4 frozen-zone (user approval)
     - `## 5-layer safety pipeline`
       - L1 frozen-guard, L2 canary, L3 contradiction, L4 rate-limit, L5 human oversight
     - `## 사용 예시` — 3-4 코드 블록 예시
     - `## 관련 자료` — cross-ref to `/moai plan`, `/moai run`, GitHub repo, SPEC-V3R4-HARNESS-001
   - 출처: skill 본문 v2.0.0 (4 verbs section / Tier-4 gate section / Authoritative Sources section)
2. **en 번역** (`docs-site/content/en/workflow-commands/moai-harness.md`): ko 구조 동일, 영어 본문
3. **ja 번역** (`docs-site/content/ja/workflow-commands/moai-harness.md`): ko 구조 동일, 일본어 본문
4. **zh 번역** (`docs-site/content/zh/workflow-commands/moai-harness.md`): ko 구조 동일, 중국어(간체) 본문
5. **_meta.yaml 4-locale 추가**: `"moai-harness": title: "/moai harness"` (각 locale 적절히 번역 — ko/en은 `/moai harness`, ja/zh는 동일)
6. **main.yaml 추가** (workflow-commands 카테고리, sync 항목 다음):
   ```yaml
         - name:
             ko: /moai harness
             en: /moai harness
             ja: /moai harness
             zh: /moai harness
           ref: /workflow-commands/moai-harness
   ```

**Verification**: AC-DCC-005, AC-DCC-006

### T-DCC-004: G2 gate 페이지 신설 (REQ-DCC-009/010)

**Inputs**: `.claude/commands/moai/gate.md` (188B) + `.claude/skills/moai/workflows/gate.md` (v1.0.0)
**Output**: 4 new files + 4 _meta.yaml entry 추가 + 1 main.yaml block 추가
**Steps**:
1. **ko 원문 작성** (`docs-site/content/ko/quality-commands/moai-gate.md`):
   - Frontmatter: `title: "/moai gate"`, `linkTitle: "/moai gate"`, `description: "lint + format + type-check + test 병렬 실행 (pre-commit 품질 게이트, <30s)"`, `weight: 15`, `geekdocCollapseSection: true`
   - Sections (ko 원문 구조, 약 200 lines):
     - `## 개요` — Lightweight pre-commit quality gate 개념 + <30s 목표
     - `## 명령어 형식` — `/moai gate [--fix] [--staged] [--file PATH]`
     - `## 옵션`
       - `### --fix` — auto-fix lint/format issues
       - `### --staged` — staged files만 검증 (`git diff --staged`)
       - `### --file PATH` — 특정 파일만 검증
     - `## 병렬 실행되는 4 단계`
       - lint (golangci-lint / ruff / eslint / clippy ...), format (gofmt / black / prettier ...), type-check (mypy / tsc ...), test (go test / pytest / jest ...)
     - `## 16-language 자동 감지` — project_markers 기반 (`go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml` 등)
     - `## /moai gate vs /moai review vs sync Phase 0.5` — skill 본문 §Difference from Other Workflows 비교 표 (3-row table)
     - `## 사용 예시` — 3-4 코드 블록 예시
     - `## 관련 자료` — cross-ref to `/moai review`, `/moai sync`, GitHub repo
   - 출처: skill 본문 v1.0.0 (Purpose / Input / Phase 1 sections)
2. **en/ja/zh 번역**: ko 구조 동일
3. **_meta.yaml 4-locale 추가**: `"moai-gate": title: "/moai gate"`
4. **main.yaml 추가** (quality-commands 카테고리, review 항목 앞 — weight: 15 < review weight: 20):
   ```yaml
         - name:
             ko: /moai gate
             en: /moai gate
             ja: /moai gate
             zh: /moai gate
           ref: /quality-commands/moai-gate
   ```

**Verification**: AC-DCC-007

### T-DCC-005: Cross-cutting + Hugo build 검증 (REQ-DCC-011/012)

**Inputs**: T-DCC-001/002/003/004 완료
**Output**: 8 ACs PASS + Hugo 빌드 exit 0 + working tree PRESERVE 확인
**Steps**:
1. acceptance.md §5 Test Execution Plan 일괄 실행 (8 ACs verification)
2. `cd docs-site && hugo --gc --minify --buildDrafts=false` exit 0 확인
3. `git status --short` 출력 비교 (PRESERVE list 정확히 유지 + 본 SPEC 변경만 추가)
4. SPEC frontmatter 갱신: `version: "0.1.0"` → `"0.2.0"`, `status: draft` → `implemented`, `updated_at: 2026-05-22`
5. `progress.md` 생성 (Tier S minimal — run-phase 산출물)

**Verification**: AC-DCC-008 + Definition of Done (acceptance.md §4.3)

---

## 8. REQ ↔ AC ↔ Task Matrix

| Task | REQs Covered | ACs Covered | Files Affected |
|------|--------------|-------------|----------------|
| T-DCC-001 | REQ-DCC-001/002/003 | AC-DCC-001/002/003 | 4 delete + 4 _meta + 1 main block remove |
| T-DCC-002 | REQ-DCC-004/005/006 | AC-DCC-004 | 4 delete + 4 _meta + 1 main block remove |
| T-DCC-003 | REQ-DCC-007/008 | AC-DCC-005/006 | 4 create + 4 _meta add + 1 main block add |
| T-DCC-004 | REQ-DCC-009/010 | AC-DCC-007 | 4 create + 4 _meta add + 1 main block add |
| T-DCC-005 | REQ-DCC-011/012 | AC-DCC-008 | Hugo build + SPEC frontmatter + progress.md |

**Coverage**: 12 REQs → 5 Tasks → 8 ACs, 100%.

**Total files affected**: 8 delete + 8 create + 12 _meta.yaml modify + 1 main.yaml modify + SPEC frontmatter (1) + progress.md (1 new) = **31 file operations**.

---

## 9. Risks (Section E)

### R-DCC-001 (Medium): main.yaml YAML syntax error

**Symptoms**: Vercel/Hugo 빌드 실패, `yaml: line N: did not find expected ...` 에러
**Mitigation**:
- Edit tool 사용 시 정확한 indentation (Hugo Geekdoc 패턴: 4-space) 유지
- T-DCC-005 step 2에서 `hugo --gc --minify --buildDrafts=false` 로 사전 검증
- Block 삭제 시 인접 entry indentation 영향 없는지 확인 (특히 list item `-` 기호)

### R-DCC-002 (Low): harness/gate 본문이 skill source와 drift

**Symptoms**: skill v2.0.0 → v2.1.0 업데이트 시 docs와 어긋남
**Mitigation**:
- "관련 자료" section에 skill source 경로 명시 (`.claude/skills/moai/workflows/harness.md`) — SSoT 참조 의무
- 본문은 핵심 개념만 (verb signature / 4-tier / 5-layer)에 집중, detail은 skill 본문 link로 위임
- 향후 별도 SPEC `SPEC-V3R6-DOCS-SSOT-CHECK-001` (provisional) 로 docs ↔ skill drift 자동 detect 가능

### R-DCC-003 (Low): 4-locale 번역 품질 부족 (ko 원문 → en/ja/zh 직역)

**Symptoms**: en/ja/zh 본문이 ko 직역체로 어색, 또는 기술 용어 부정확
**Mitigation**:
- ko를 원문으로 하되 en은 ko-en 대응 쌍 (예: "프록시" → "proxy") 표준화
- 기존 페이지 (sync.md, plan.md 등) 의 4-locale 본문을 참조하여 톤 일치
- ja/zh 본문은 ko 구조 동일 + 기술 용어 (status, apply, rollback, disable, lint, format)는 그대로 보존
- 번역 품질은 본 SPEC scope 외 — 추후 native speaker 검토 가능

### R-DCC-004 (Low): _meta.yaml 4-locale entry 누락 가능성

**Symptoms**: ko에는 추가했으나 en/ja/zh 누락 → sidebar 메뉴 비대칭
**Mitigation**:
- T-DCC-003/004에서 4-locale을 batch로 처리 (ko → en → ja → zh 순차)
- AC-DCC-006/007의 `meta_added=4` 카운터로 자동 검증

---

## 10. Section E — Self-Verification Deliverables (run-phase 보고 의무)

run-phase 완료 시 manager-develop 또는 orchestrator가 다음 보고:

### E.1 AC Binary PASS/FAIL Matrix

acceptance.md §1 의 8 ACs 모두에 대해 실제 verification command 실행 결과 표 + 예상 output 일치 여부:

| AC | Status | Verification | Actual |
|----|--------|--------------|--------|
| AC-DCC-001 | PASS/FAIL | `ls ... \| grep -c "No such file"` | (실제 출력) |
| ... | ... | ... | ... |

### E.2 Hugo Build 결과

```
$ cd docs-site && hugo --gc --minify --buildDrafts=false
(stdout/stderr)
exit code: (0 or non-zero)
```

### E.3 Working Tree PRESERVE 검증

`git status --short` 출력에서 spec.md §6.2 PRESERVE list (`.moai/harness/usage-log.jsonl`, `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html`, `.moai/research/`, `.moai/specs/SPEC-V3R5-*`, `docs-site/scripts/`, `internal/hook/.moai/`) 정확히 유지 + 본 SPEC 변경 사항만 추가됨 확인.

### E.4 Files Affected Count (정확한 분해)

**Deletes (8 files)**:
- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-db.md` (4 files)
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-github.md` (4 files)

**Creates (9 files)**:
- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-harness.md` (4 files)
- `docs-site/content/{ko,en,ja,zh}/quality-commands/moai-gate.md` (4 files)
- `.moai/specs/SPEC-V3R6-DOCS-CMD-CATALOG-001/progress.md` (1 file, run-phase 종료 산출물)

**Modifies — `_meta.yaml` (12 files total, 16 entry-changes)**:
- workflow-commands/_meta.yaml × 4-locale: 각 파일당 `moai-db` entry 제거 + `moai-harness` entry 추가 = **2 entry-changes/file × 4 = 8**
- utility-commands/_meta.yaml × 4-locale: 각 파일당 `moai-github` entry 제거 = **1 entry-change/file × 4 = 4**
- quality-commands/_meta.yaml × 4-locale: 각 파일당 `moai-gate` entry 추가 = **1 entry-change/file × 4 = 4**

**Modifies — `main.yaml` (1 file, 4 block-changes)**:
- `/workflow-commands/moai-db` block 제거 (-1 block)
- `/utility-commands/moai-github` block 제거 (-1 block)
- `/workflow-commands/moai-harness` block 추가 (+1 block)
- `/quality-commands/moai-gate` block 추가 (+1 block)

**Modifies — SPEC artifacts (3 files, run-phase 종료 직전)**:
- spec.md frontmatter: `status: draft` → `status: implemented`, `version: 0.1.1` → `0.2.0`, `updated: 2026-05-22`
- plan.md: progress 링크 추가 (선택)
- acceptance.md: (변경 없음)

**Grand total**: 8 delete + 9 create + 12 _meta modify + 1 main modify + 1 spec.md frontmatter rewrite = **31 file operations**.

### E.5 Cross-Platform / Subagent Boundary

- Cross-platform: n/a (콘텐츠 변경, OS-independent)
- C-HRA-008: n/a (no harness/hook Go files)
- Lint NEW=0: Hugo build PASS로 대체

### E.6 Branch HEAD + Commit 상태

- 새 commits SHA 리스트 (run-phase에서 1개 또는 2개 commits 예상: `fix(docs-site): R1/R2 false promise removal` + `feat(docs-site): G1/G2 harness/gate 4-locale pages`, 또는 단일 squash commit `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): Critical 4건 정정`)
- main 직진 (Late-Branch policy, sync 단계에서 별도 branch 생성 + cherry-pick)

### E.7 Blocker Report (있을 시)

- 본 SPEC scope 외 사용자 결정 필요 항목 발견 시 structured 보고 (AskUserQuestion 절대 호출 금지)
- 예: 번역 품질 검토 필요, 또는 main.yaml syntax 충돌

---

## 11. Operational Notes

### 11.1 Late-Branch policy (sync 단계)

본 SPEC의 run-phase는 main 직진 commits 생성. sync 단계에서:
1. 누적 dirty (본 SPEC 변경 + PRESERVE list 모두) 를 stash
2. `git switch -c feat/SPEC-V3R6-DOCS-CMD-CATALOG-001 origin/main`
3. 본 SPEC commits cherry-pick
4. `git push -u origin feat/SPEC-V3R6-DOCS-CMD-CATALOG-001`
5. `gh pr create` (4-locale parity + Hugo build PASS 명시)

### 11.2 Commit 전략 옵션

**옵션 A (단일 squash)**: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): 4-locale command catalog drift 정정 (R1/R2 삭제 + G1/G2 신설)` — 단일 atomic commit (29 files)

**옵션 B (3 commits)**:
- C1: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): R1/R2 false promise 제거 (db/github 4-locale 삭제)`
- C2: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): G1 harness 페이지 4-locale 신설`
- C3: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): G2 gate 페이지 4-locale 신설`

**권장**: 옵션 B (작업 단위별 commits → diff 검토 용이) — run-phase 위임 시 manager-develop 또는 orchestrator 결정.

### 11.3 Plan-auditor 호출 옵션

Tier S optional이지만 4-locale × 4 files = 16 file 영향이므로 orchestrator가 plan-auditor 호출 결정 가능. plan-auditor 호출 시:
- Tier S threshold: 0.75
- 1-pass PASS 목표
- 호출 형식: orchestrator가 별도 `Agent(plan-auditor)` 위임 (manager-spec은 호출하지 않음 — plan-phase 산출물 작성만)

### 11.4 Open Questions resolution

spec.md §7 OQ1-OQ4 모두 default 진행으로 가정. orchestrator가 spec.md 검토 후 AskUserQuestion 통해 최종 결정 가능:
- default 채택: 본 SPEC 그대로 run-phase 진입
- default 거부: spec.md 수정 + plan.md/acceptance.md 영향 반영 후 재진행

---

## 12. Task Priority (시간 추정 금지 — Priority 라벨 사용)

`.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule 준수: "Never use time predictions in plans or reports". 시간 추정 대신 **Priority 라벨** 사용.

| Task | Priority | 작업 성격 | 의존성 |
|------|----------|---------|--------|
| T-DCC-001 (R1 제거) | **High** | Mechanical delete + Edit, 부담 적음 | 없음 (병렬 가능) |
| T-DCC-002 (R2 제거) | **High** | Mechanical delete + Edit, 부담 적음 | 없음 (병렬 가능) |
| T-DCC-003 (G1 신설 4-locale) | **High** | ko 원문 + en/ja/zh 번역 + meta/main, 작업 부담 大 (4-locale 번역) | skill source 본문 변경 없음 가정 |
| T-DCC-004 (G2 신설 4-locale) | **High** | ko 원문 + en/ja/zh 번역 + meta/main, 작업 부담 中 (skill 본문이 더 짧음) | skill source 본문 변경 없음 가정 |
| T-DCC-005 (검증 + Hugo build + frontmatter) | **High** | AC 8개 일괄 verification + Hugo 빌드 + SPEC frontmatter 갱신 | T-DCC-001/002/003/004 완료 후 |

**Execution order**: T-DCC-001/002 (병렬) → T-DCC-003/004 (병렬) → T-DCC-005 (sequential, 검증). 5개 Task 모두 Priority High — 본 SPEC scope 자체가 Critical drift cleanup이기 때문.

---

## References

- spec.md (본 SPEC 의 SSoT)
- acceptance.md (verification 명령 + REQ↔AC matrix)
- baseline: `.moai/research/moai-adk-current-state-2026-05-22.md` §2 Critical Drift Inventory
- spec-workflow.md § SPEC Complexity Tier (Tier S 분류 근거)
- manager-develop-prompt-template.md § Section A-E 5-section (본 plan의 구조)
- 선례 SPEC: SPEC-V3R5-DOCS-SECURITY-001 (commit `c94d8b203`, 4-locale 신설 패턴)
- skill sources: `.claude/skills/moai/workflows/{harness,gate}.md`
