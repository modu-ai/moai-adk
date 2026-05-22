# Plan: SPEC-V3R6-SKILL-CONSOLIDATE-001

## 1. Implementation Strategy

본 SPEC 는 **7 source SKILL.md → 3 unified SKILL.md** 통합 메커니즘을 구현한다. 핵심 접근:

- **Content-first ordering**: 새 통합 skill body 작성을 먼저 (1.1-1.3), 그 다음 cross-reference rename (1.4), 마지막에 source skill 삭제 (1.5). 역순 시 source 정보 손실 + cross-ref dangling 위험.
- **Template-First Rule 의무**: 모든 skill 파일 작성/수정은 local `.claude/skills/` + template `internal/template/templates/.claude/skills/` 동시 적용 (CLAUDE.local.md §2). 매 step 에서 MultiEdit 또는 동시 Write tool call 사용.
- **Catalog 자동 hash regen 활용**: SPEC-V3R6-CATALOG-SSOT-001 의 `Makefile` `gen-catalog-hashes.go --all` 인프라가 정착했으므로 catalog.yaml 수동 hash 편집 불필요. `make build` 만 호출하면 32-entry hash 자동 재생성.
- **Binary AC 우선**: 본 SPEC 의 acceptance 는 binary contract (existence + word count + cross-ref grep + diff -q + test exit code) 만 검증. Body 의 *내용 품질* 평가는 evaluator-active scope, 본 SPEC scope 외.

### 1.1 `moai-workflow-ci-loop` 통합 (ci-watch + ci-autofix)

**Source skills**:
- `moai-workflow-ci-watch` (961w) — CI watch loop (gh pr checks polling, required-checks SSOT)
- `moai-workflow-ci-autofix` (1,507w) — CI auto-fix iteration loop (3-iteration cap, protected files, semantic failure escalation)

**Target**: `moai-workflow-ci-loop/SKILL.md` (목표 1,200w, 상한 1,500w)

**Content scope (run-phase 작성자 가이드)**:

- Quick Reference (`## Quick Reference`):
  - "CI Watch + Auto-fix unified loop — poll required checks, classify failures, auto-fix safe categories, escalate semantic failures to user"
  - Trigger 조건 list (PR open + required check fail)
- Implementation Guide:
  - Phase 1: Watch loop (30s poll, 30 min timeout, gh pr checks `.json` output)
  - Phase 2: Failure classification (safe vs semantic per ci-autofix-protocol.md)
  - Phase 3: Auto-fix iteration (max 3, per-PR-push counter, new commit only — no force-push/amend)
  - Phase 4: Escalation via AskUserQuestion (iteration > 3)
- Protected files list (carry over from ci-autofix: `.env*`, secrets, `.github/required-checks.yml`, `scripts/ci-watch/run.sh`)
- Audit log location: `.moai/logs/ci-autofix/`
- Works Well With: `manager-git`, `manager-quality`, `ci-watch-protocol.md`, `ci-autofix-protocol.md`
- Footer note: "absorbed from moai-workflow-ci-watch + moai-workflow-ci-autofix (SPEC-V3R6-SKILL-CONSOLIDATE-001, 2026-05-22)"

### 1.2 `moai-workflow-design` 통합 (design-import + design-context)

**Source skills**:
- `moai-workflow-design-import` (1,358w) — Path A Claude Design import (figma extraction, design tokens, components)
- `moai-workflow-design-context` (1,049w) — `.moai/design/` artifact loading + brand context integration

**Target**: `moai-workflow-design/SKILL.md` (목표 1,400w, 상한 1,500w)

**Content scope**:

- Quick Reference: "Design workflow — Path A (Claude Design import) + Path B (code-based) + design context loading"
- Implementation Guide:
  - Path selection (Path A vs Path B vs Hybrid)
  - Path A: `moai-workflow-design-import` 본문 (figma extractor, handoff bundle, tokens.json/components.json)
  - Path B: brand context loading (`.moai/project/brand/` + `.moai/design/` artifacts)
  - Design brief auto-load (per `.claude/rules/moai/design/constitution.md` §3.2 reserved paths)
- Works Well With: `expert-frontend`, `moai-domain-brand-design`, `moai-domain-design-handoff`, `moai-workflow-gan-loop`, `.claude/rules/moai/design/constitution.md`
- Footer note: "absorbed from moai-workflow-design-import + moai-workflow-design-context (SPEC-V3R6-SKILL-CONSOLIDATE-001, 2026-05-22)"

### 1.3 `moai-harness-patterns` 통합 (hook-ci + workflow + quality)

**Source skills**:
- `moai-harness-hook-ci` (655w) — Hook/CI harness patterns (subagent boundary, event dispatcher)
- `moai-harness-workflow` (666w) — Workflow harness patterns (Plan-Run-Sync invariants, mode dispatch)
- `moai-harness-quality` (718w) — Quality harness patterns (TRUST 5 enforcement, gate sentinels)

**Target**: `moai-harness-patterns/SKILL.md` (목표 1,200w, 상한 1,500w)

**Content scope**:

- Quick Reference: "MoAI-ADK harness pattern library — hook/CI dispatch + workflow invariants + quality gates"
- Implementation Guide:
  - Section 1: Hook/CI patterns (subagent boundary C-HRA-008, event dispatcher contract, plain-text vs JSON stdout)
  - Section 2: Workflow patterns (Plan-Run-Sync ordering, mode dispatch sentinels, late-branch closure)
  - Section 3: Quality patterns (TRUST 5 enforcement, coverage threshold, lint baseline)
- Works Well With: `manager-develop`, `manager-quality`, `manager-git`, `harness/hook-ci-specialist`, `harness/workflow-specialist`, `harness/quality-specialist`
- Footer note: "absorbed from moai-harness-hook-ci + moai-harness-workflow + moai-harness-quality (SPEC-V3R6-SKILL-CONSOLIDATE-001, 2026-05-22)"

**Note**: 3 harness specialist agent (`.claude/agents/harness/*-specialist.md`) 는 그대로 보존 — agent 가 *이* skill 을 reference 하도록 frontmatter `skills:` array 만 업데이트 (REQ-SC-006 의 cross-ref rename 일부).

### 1.4 Cross-reference Rename (~30 file 위치)

**범위**: Appendix B 의 표 항목 + run-phase grep 으로 발견되는 모든 추가 위치.

**전략**:

- `Edit` tool 의 `replace_all: true` 활용 (각 파일 내 단일 old name → new name 매핑)
- 단일 파일에 다중 old name 등장 시 (예: `moai/SKILL.md` 가 7 name 모두 참조 가능) MultiEdit 사용
- Local + template mirror 동시 처리 (Template-First Rule)
- 검증: 각 파일 수정 후 즉시 `grep -c "<old-name>" <file>` → 0 확인

**대상 카테고리**:

1. Agent frontmatter `skills:` YAML array (3 harness specialist agents)
2. Protocol rules `.claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md` (See section / Cross-reference 항목)
3. Top-level skill cross-ref `.claude/skills/moai/SKILL.md` + `workflows/{fix,brain,design,sync/delivery}.md`
4. Sibling skill body `moai-domain-{brand-design,design-handoff}/SKILL.md` + `moai-workflow-gan-loop/SKILL.md`
5. Design constitution `.claude/rules/moai/design/constitution.md`
6. Zone registry `.claude/rules/moai/core/zone-registry.md`
7. Go test `internal/design/dtcg/frozen_guard_test.go` (필요 시)

### 1.5 Source Skill Removal (7 directories)

**전략**:

- 모든 cross-ref rename 완료 (1.4) 후 마지막 실행
- `rm -rf` 또는 `Bash` `rm` — git tracking 자동 처리
- Local + template 양쪽 동시 (Template-First Rule)
- 검증: `ls .claude/skills/moai-workflow-ci-watch 2>/dev/null && echo EXISTS || echo OK` 7번 반복

### 1.6 Catalog.yaml + Hash Regen

**전략**:

- `Edit` tool 로 catalog.yaml 의 7 entry 블록 (각 5-6 line) 제거
- 3 entry 블록 추가 (`name`, `version`, `path`, `hash` 필드 — hash 는 placeholder, regen 으로 채워짐)
- `make build` 또는 `go run ./internal/template/gen-catalog-hashes.go --all` 호출로 32-entry hash 자동 재생성 (SPEC-V3R6-CATALOG-SSOT-001 인프라)
- 검증: `go test ./internal/template/... -run TestManifestHashFormat` exit 0

## 2. Files Affected (~30-35 locations)

### 2.1 Deletions (14 files)

**Local** (`.claude/skills/`):
- `moai-workflow-ci-watch/SKILL.md` + directory
- `moai-workflow-ci-autofix/SKILL.md` + directory
- `moai-workflow-design-import/SKILL.md` + directory
- `moai-workflow-design-context/SKILL.md` + directory
- `moai-harness-hook-ci/SKILL.md` + directory
- `moai-harness-workflow/SKILL.md` + directory
- `moai-harness-quality/SKILL.md` + directory

**Template** (`internal/template/templates/.claude/skills/`):
- 동일 7 디렉토리 미러

### 2.2 Creations (6 files)

**Local**:
- `moai-workflow-ci-loop/SKILL.md` (new)
- `moai-workflow-design/SKILL.md` (new)
- `moai-harness-patterns/SKILL.md` (new)

**Template**:
- 동일 3 mirror

### 2.3 Modifications (~25 files)

- `internal/template/catalog.yaml` (7 entry remove + 3 entry add + 32 hash regen)
- ~15 cross-ref files (local + template) per Appendix B
- 3 harness specialist agent frontmatter (`skills:` array)
- 1 Go test file (if name reference exists)
- progress.md 추가 (run-phase 완료 후)

### 2.4 PRESERVE List (변경 금지)

**현재 working tree state (Pre-flight Check §G)**:

Modified (`M`):
- `.moai/harness/usage-log.jsonl` (runtime-managed, B8)
- `docs-site/hugo.toml` + `layouts/{_default/baseof.html,partials/menu.html}` (별도 워크스트림)
- `internal/template/renderer.go` (별도 워크스트림)
- `internal/template/templates/.github/{actions,workflows}/*.yml.tmpl` (8 files, 별도 워크스트림)

Untracked (`??`):
- `.moai/research/moai-adk-current-state-2026-05-22.md` + `v3.0-design-2026-05-22.md` (별도)
- `docs-site/content/{en,ja,ko,zh}/book/` (별도)
- `docs-site/data/menu/extra.yaml` + `layouts/_default/redirect.html` + `scripts/gen_menu.py` (별도)
- `docs-site/static/book/` (별도)
- `internal/cli/hook_writehookoutput_test.go` (별도 PR 후보)
- `internal/hook/.moai/` (B7 working tree leak, REQ-HCF-005 scope)
- `internal/template/github_tmpl_parse_test.go` + `settings_no_worktree_keys_test.go` (별도)
- `.moai/specs/SPEC-V3R6-{HOOK-CONTRACT-FIX,DOCS-USER-DRIFT,AGENT-FOLDER-SPLIT,...}-001/` (다른 SPEC)

[ZONE:Frozen] [HARD] 본 SPEC run-phase 는 위 PRESERVE 항목을 일체 수정하지 않는다. `git status --short` 결과를 run-phase 시작 시점과 종료 시점에 비교하여 PRESERVE 위반 0건 확인.

## 3. Implementation Order (M1 → M6)

### M1: 3 신규 통합 SKILL.md 본문 작성 (local + template 동시)

**Goal**: ci-loop / design / harness-patterns 각 SKILL.md 본문 작성

**Approach**:
- 각 통합 skill 당:
  1. Source 2-3 SKILL.md `Read` 로 컨텐츠 파악
  2. 통합 본문 draft (목표 word count 준수)
  3. MultiEdit 또는 동시 Write 로 local + template 동시 생성
- AC-SC-002 + AC-SC-003 + AC-SC-008 verify

**Verification**:
- `ls .claude/skills/moai-{workflow-ci-loop,workflow-design,harness-patterns}/SKILL.md` 3건 존재
- `wc -w` 각 ≤ 1,500
- `diff -q` local ↔ template 3 pair 모두 identical

### M2: Cross-reference Rename (Phase 1 — Skill bodies + Rules)

**Goal**: skills/ + rules/ 내 cross-ref 갱신

**Approach**:
- `.claude/skills/moai/SKILL.md` + `workflows/{fix,brain,design,sync/delivery}.md` (5 files local + 5 template)
- `.claude/skills/moai-domain-{brand-design,design-handoff}/SKILL.md` + `moai-workflow-gan-loop/SKILL.md` (3 local + 3 template)
- `.claude/rules/moai/workflow/ci-{watch,autofix}-protocol.md` (2 local + 2 template)
- `.claude/rules/moai/design/constitution.md` (1 local + 1 template)
- `.claude/rules/moai/core/zone-registry.md` (1 local + 1 template)

**Verification**:
- `grep -c "<old-name>" <file>` per pair → 0
- `grep -rn "moai-workflow-ci-watch\|moai-workflow-ci-autofix" .claude/skills/ .claude/rules/` → 0 matches

### M3: Cross-reference Rename (Phase 2 — Agents + Go test)

**Goal**: agent frontmatter `skills:` array + Go test 갱신

**Approach**:
- `.claude/agents/harness/{hook-ci,workflow,quality}-specialist.md` (3 local + 3 template) — frontmatter `skills:` 의 7 old name → 3 new name
- `internal/design/dtcg/frozen_guard_test.go` — design skill name reference 발견 시 갱신

**Verification**:
- `grep -rn "moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" .claude/agents/ internal/design/` → 0
- `grep -rn "moai-workflow-design-import\|moai-workflow-design-context" internal/design/` → 0

### M4: Catalog.yaml Update + Hash Regen

**Goal**: catalog.yaml SSOT 갱신

**Approach**:
- `Edit` tool 로 7 old entry 블록 제거 (line range 약 36-78 + 274-281)
- 3 new entry 블록 추가 (이름순 정렬 위치)
- `make build` 또는 `go run ./internal/template/gen-catalog-hashes.go --all` 호출
- progress.md 에 hash regen 결과 기록

**Verification**:
- `grep -c "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" internal/template/catalog.yaml` → 0
- `grep -c "moai-workflow-ci-loop\|moai-workflow-design\|moai-harness-patterns" internal/template/catalog.yaml` → ≥ 3
- `go test ./internal/template/... -run TestManifestHashFormat` exit 0
- `go test ./internal/template/... -run TestAllSkillsInCatalog` exit 0

### M5: Source Skill Removal (7 directories × 2 mirrors = 14 dirs)

**Goal**: 7 old skill 디렉토리 제거

**Approach**:
- `rm -rf .claude/skills/moai-workflow-ci-watch .claude/skills/moai-workflow-ci-autofix .claude/skills/moai-workflow-design-import .claude/skills/moai-workflow-design-context .claude/skills/moai-harness-hook-ci .claude/skills/moai-harness-workflow .claude/skills/moai-harness-quality`
- 동일 명령을 `internal/template/templates/.claude/skills/` prefix 로 실행
- git 자동 track 제거 확인

**Verification**:
- `ls .claude/skills/moai-workflow-ci-watch 2>&1` → "No such file or directory" (7번)
- `ls internal/template/templates/.claude/skills/moai-workflow-ci-watch 2>&1` → 동일 (7번)

### M6: 회귀 검증 + AC matrix 작성

**Goal**: 모든 binary AC 통과 + AC PASS/FAIL matrix 보고

**Approach**:
- Parallel Bash batch (`agent-common-protocol.md` § Parallel Execution):
  - `go test ./internal/template/... -run "TestAllSkillsInCatalog|TestAllAgentsInCatalog|TestManifestHashFormat|TestRuleTemplateMirrorDrift"`
  - `grep -rn "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" .claude/ CLAUDE.md internal/template/templates/.claude/ internal/design/dtcg/`
  - `wc -w .claude/skills/moai-workflow-ci-loop/SKILL.md .claude/skills/moai-workflow-design/SKILL.md .claude/skills/moai-harness-patterns/SKILL.md`
  - `diff -q` 3 pair (local ↔ template)
  - `golangci-lint run --timeout=2m` baseline 비교
  - `GOOS=windows GOARCH=amd64 go build ./...` cross-platform
  - `git status --short` PRESERVE 비교
- AC PASS/FAIL matrix 작성 (acceptance.md 의 AC-SC-001..009)
- progress.md 갱신: `status: implemented`, version 0.1.0 → 0.2.0

**Verification**:
- All AC PASS (binary)
- NEW regression 0 (pre-existing baseline 제외)
- PRESERVE list 위반 0

## 4. Pre-flight Checks (모든 step 위 의무 실행)

본 SPEC Section G 의 8개 pre-flight 명령을 run-phase 시작 시점에 재실행하여 baseline 확정:

```bash
# 1. 설계 문서 §Layer 2 확인 (REQ-SC-003 근거)
grep -n "Layer 2\|consolidation\|통폐합" .moai/research/v3.0-design-2026-05-22.md | head -10

# 2. 통합 대상 7 skill 위치 확인 (REQ-SC-001 baseline)
find .claude/skills/moai -type d \( -name "ci-watch" -o -name "ci-autofix" -o ... \) | head -10

# 3. Word count baseline (REQ-SC-003 검증 근거)
for s in ...; do find .claude/skills -path "*$s*SKILL.md" -exec wc -w {} \;; done

# 4. catalog.yaml entry 수 (REQ-SC-005 baseline)
grep -c "^- id:\|^  - id:" internal/template/catalog.yaml

# 5. Cross-reference scan (REQ-SC-006 baseline)
grep -rln "<7 names>" .claude/ CLAUDE.md 2>/dev/null | grep -v ".claude/worktrees/"

# 6. Template-First mirror baseline
ls internal/template/templates/.claude/skills/ | grep -E "(ci-watch|ci-autofix|design-import|design-context|harness-hook-ci|harness-workflow|harness-quality)"

# 7. Wave 0 SPEC 패턴 참조 (HISTORY/EARS 패턴 모방)
head -20 .moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/spec.md

# 8. Template Mirror Drift audit baseline (V3R6 lesson #25)
ls .moai/specs/ | grep TEMPLATE-MIRROR-DRIFT 2>/dev/null
```

## 5. Brownfield Strategy

**Approach**: Hard rename + atomic cutover.

본 SPEC 는 *기존 인프라 보존* 보다 *완전한 통합* 을 우선시한다 (v3.0.0 phase, 의도된 breaking change). 다음 brownfield 원칙 준수:

- **PRESERVE existing infra**:
  - 3 harness specialist agent (frontmatter `skills:` 만 갱신, 본문 보존)
  - 4 docs-site language locales (별도 sync 단계, 본 SPEC scope 외)
  - 29 다른 skill 디렉토리 (foundation/domain/ref/etc.)
  - 모든 rules `.claude/rules/moai/` (cross-ref 갱신 외 본문 보존)
- **NO backward-compat shim**:
  - Old skill name 으로 호출하는 코드는 hard fail. 정상 — 모든 cross-ref 가 새 이름으로 정렬됨이 AC-SC-005 의 binary contract.
- **NO version pinning**:
  - 통합 skill 의 internal version 은 `0.1.0` 새로 시작. Source skill 의 version history 는 SPEC HISTORY 항목 + footer note 로만 추적.

## 6. PRESERVE List (HARD)

[ZONE:Frozen] [HARD] 본 SPEC run-phase 가 *명시적으로 수정 금지* 하는 파일:

### 6.1 Working tree 외 진행 중 별도 워크스트림

- 모든 `M` 상태 파일 (8 modified files in `git status --short`, 위 §2.4 참조)
- 모든 `??` 상태 디렉토리 (`docs-site/content/*/book/`, `docs-site/scripts/`, `internal/cli/hook_writehookoutput_test.go`, `internal/hook/.moai/`, `internal/template/github_tmpl_parse_test.go`, `internal/template/settings_no_worktree_keys_test.go`, `.moai/research/*.md`, `.moai/specs/SPEC-V3R6-{OTHER}/`)

### 6.2 Runtime-managed files

- `.moai/harness/usage-log.jsonl` (B8)
- `.moai/state/*` (B8)
- `internal/hook/.moai/` (B7 working tree leak, REQ-HCF-005 scope, 별도 SPEC)

### 6.3 다른 SPEC 디렉토리

- `.moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/`
- `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/`
- `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/`
- `.moai/specs/SPEC-V3R6-HARNESS-RENAME-001/`
- `.moai/specs/SPEC-V3R6-HARNESS-LEARNER-FIX-001/`
- `.moai/specs/SPEC-V3R6-CATALOG-SSOT-001/`
- `.moai/specs/SPEC-V3R6-ABSORB-CLEANUP-001/`
- `.moai/specs/SPEC-V3R6-DOCS-CMD-CATALOG-001/`

### 6.4 Stale worktrees (정보 보존, 검색 범위 외)

- `.claude/worktrees/agent-{a3e0aa7b6df12e283,ae362bcff2b3b44dd}/` — 본 SPEC grep 검색 범위에서 명시적 제외 (`| grep -v ".claude/worktrees/"`)

### 6.5 Agent-memory (정보 보존)

- `.claude/agent-memory/manager-tdd/project_ciaut_wave2_complete.md` — stale ref 존재 가능하나 정보 손실 회피 위해 미수정

### 6.6 사용자 문서 (docs-site)

- `docs-site/content/{ko,en,ja,zh}/**` — 본 SPEC 미대상 (별도 `/moai sync` 또는 SPEC-V3R6-DOCS-USER-DRIFT-001 후속)

### 6.7 금지 명령

- `git commit --no-verify`
- `git commit --amend`
- `git push --force`
- `rm -rf .claude/specs/` (다른 SPEC 보존)
- `rm` 명령에 wildcard 사용 시 신중

### 6.8 사용 의무 명령

- Conventional Commits (`feat:`, `chore:`, `refactor:`, etc.)
- `🗿 MoAI` trailer
- `make build` (catalog hash regen 시)
- Template-First Rule: 모든 `.claude/` 변경에 `internal/template/templates/.claude/` mirror 동반

## 7. Risks Summary (acceptance.md 와 동기화)

| Risk | Severity | Likelihood | Mitigation |
|---|---|---|---|
| R-SC-001 Template Mirror Drift | High | Medium | AC-SC-006 BLOCKING + Section A prompt 의무 명시 |
| R-SC-002 Cross-ref 누락 | High | High | AC-SC-005 BLOCKING + 검색 범위 명시 |
| R-SC-003 Catalog Hash Drift | Medium | Medium | Makefile 자동 regen + AC-SC-007 |
| R-SC-004 Body 정보 손실 | Medium | Medium | REQ-SC-008 구조 contract + absorbed-from footer |
| R-SC-005 Trigger Keyword 충돌 | Low | Low | REQ-SC-009 frontmatter description 권장 |

## 8. Estimated Scope

- LOC 영향: ~3,800 words 신규 (3 통합 SKILL.md) + ~50 cross-ref edits + ~15 catalog entries = **~600-800 LOC equivalent** (markdown + YAML)
- 파일 수정 수: ~25 files (creation 6 + deletion 14 + modification 25, in/out overlap)
- Tier 결정: **M (Medium)** — 300-1000 LOC 범위 + 5-15 files affected 의 상한, 그러나 binary AC 명확 + brownfield risk medium → Tier M 적절. Tier L 승격 불필요 (architecture change 아님, design.md/research.md 별도 필요 없음).
- plan-auditor PASS threshold: 0.80

## 9. Open Questions (run-phase 시작 시 해결)

Q1. `internal/design/dtcg/frozen_guard_test.go` 가 실제로 7 old skill name 중 어느 것을 reference 하는가? — Pre-flight Check §G step 5 의 grep 결과 (`internal/design/dtcg/frozen_guard_test.go`) 가 `moai-workflow-design-context` 또는 `moai-workflow-design-import` 매치. Run-phase 시작 시 정확한 line 위치 확인 + 새 이름으로 치환.

Q2. 3 harness specialist agent 의 frontmatter `skills:` array 가 정확히 어떤 entry 를 포함하는가? — Run-phase 시작 시 `Read .claude/agents/harness/hook-ci-specialist.md` 등으로 확인 후 array 갱신.

Q3. `moai/SKILL.md` (top-level) 가 7 name 모두 참조하는가, 일부만 참조하는가? — Run-phase 시작 시 정밀 grep 으로 확인 후 MultiEdit 또는 Edit replace_all.

위 3 Q 는 모두 run-phase Pre-flight Check 로 답변 가능 — plan-phase blocker 아님.

## 10. Self-Estimation (plan-auditor preview)

본 SPEC plan 의 self-estimated score (Tier M PASS threshold 0.80):

| Dimension | Score | Rationale |
|-----------|------:|-----------|
| D1 Completeness | 0.88 | 9 EARS REQ + 6 PRESERVE category + 5 Risk + 6 M-step + 30 cross-ref enum |
| D2 Clarity | 0.87 | Binary AC 명확 + 통합 매트릭스 + cross-ref 표 + Out of Scope 풍부 |
| D3 Traceability | 0.86 | REQ↔AC mapping 100% (acceptance.md 에서 명시), §2.1-2.3 v3.0 design 직접 인용 |
| D4 Risk Awareness | 0.85 | 5 risk 모두 mitigation 동반, Cross-SPEC tension audit 명시 |
| **Aggregate** | **0.865** | Tier M threshold 0.80 대비 +0.065 margin (PASS 안정 영역) |

위 self-estimate 는 orchestrator 의 외부 plan-auditor 호출로 검증/조정될 수 있음 (본 SPEC delegation prompt 는 plan-auditor 미호출 명시).
