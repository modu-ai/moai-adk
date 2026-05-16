# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — Task Breakdown

> **Methodology**: TDD/DDD cycle 부적용 (metadata-only). 대신 **3-Wave evidence-driven workflow** 적용.
> **Sequential**: Wave 1 → Wave 2 → Wave 3. 각 Wave 내부도 sequential.

## Wave 1 — Analysis (Category B per-SPEC investigation)

목표: Category B 12건 각각에 대해 evidence-driven mechanism 결정.

### T1-1: Verify lint baseline (sanity check)

- **Command**: `moai spec lint --strict 2>&1 | tail -1`
- **Expected**: `0 error(s), 17 warning(s)` (또는 동등한 17 WARNING)
- **Output**: progress.md 의 Wave 1 섹션에 baseline 기록
- **Note**: 만약 17 이 아닌 다른 수 면 main 이 변경된 것 — plan.md §3 inventory 재확인 필요

### T1-2: Category B git log evidence capture (B1..B12)

각 Category B SPEC 에 대해:

- **Command**: `git log main --oneline --no-merges --grep="\bSPEC-X\b" -10`
- **Output**: plan.md §3 Category B Analysis Table 의 해당 row "git log evidence" column 갱신
- **Iterations**: 12 SPECs sequential (B1 SPEC-V3R2-ORC-003 부터 B12 SPEC-V3R4-LINT-SKIP-CLEANUP-001 까지)

### T1-3: Bulk closure commit body inspection

OQ-2 resolution. 다음 bulk closure commits 의 body 점검:

- **Command**: `git show --format=full <sha>` (no diff)
- **Targets**:
  - `2c81e17b6` (PR #926, V3 bulk status drift 일괄 해소)
  - `19957efd8` (PR #927, V3 final status closeout)
  - `325a6492a` (PR #856, RT-007 closure + status drift 4건)
- **Output**: 각 commit body 가 명시한 SPEC-ID 목록 → plan.md §3 의 B-N rows mechanism column 결정 보조

### T1-4: Walker prefix recognition verification

OQ-1 resolution. 다음 prefixes 가 transitions.go 에 인식되는지 확인:

- **Command**: `Read /Users/goos/MoAI/moai-adk-go/internal/spec/transitions.go` (검증 only, 미수정)
- **Targets**: `docs(sync):`, `fix(hook):`, `feat(mx):`, `chore(post-X):`
- **Output**: 각 prefix 의 walker 인식 여부 → B3 (docs(sync)), B11 (fix(hook)), B4 (feat(mx) + T-SPC002-*) mechanism 결정 보조

### T1-5: External evidence cross-check (CHANGELOG + project memory)

각 Category B SPEC 에 대해 외부 완료 증거 확인:

- **Sources**:
  - `CHANGELOG.md` 또는 `internal/template/templates/CHANGELOG.md`
  - project memory (`~/.claude/projects/{hash}/memory/project_*<spec>*.md`)
- **Command**: `grep -l "SPEC-X" <sources>`
- **Output**: 완료 증거 유무 → B2/B4/B5 (Big gap 군) 의 frontmatter downgrade vs sync-commit 결정

### T1-6: Per-SPEC mechanism decision

각 Category B SPEC 에 대해 final mechanism 채택:

- **Decision matrix** (design.md §3 + §7 기반):
  - Walker visibility 한계 명확 + 외부 완료 증거 있음 → `lint.skip`
  - Walker visibility 한계 + 외부 완료 증거 부재 → `frontmatter downgrade`
  - Bulk closure commit body 에 SPEC-ID 명시 부재 → `sync(spec) commit` 추가
  - SDF-001 chain self-drift → `lint.skip with chain-rolled-up reason`
- **Output**: plan.md §3 Category B Analysis Table 의 "Mechanism (plan 가설 → run 결정)" column 갱신

---

## Wave 2 — Apply (frontmatter + sync-commits)

### T2-A-1: Category A bulk sync-up (Wave 2-A)

5 SPECs frontmatter 갱신:

- **Files**:
  - `.moai/specs/SPEC-GLM-MCP-001/spec.md`: `status: in-progress → completed`
  - `.moai/specs/SPEC-STATUSLINE-001/spec.md`: `status: in-progress → implemented`
  - `.moai/specs/SPEC-V3R2-WF-002/spec.md`: `status: in-progress → implemented`
  - `.moai/specs/SPEC-V3R4-CATALOG-001/spec.md`: `status: implemented → completed`
  - `.moai/specs/SPEC-WORKTREE-002/spec.md`: `status: implemented → completed`
- **Additionally**:
  - 각 파일 `updated: 2026-05-16` 갱신
  - 각 파일 HISTORY 신규 row 추가 (1 line): `0.X.Y | 2026-05-16 | manager-develop (FOLLOWUP-002 Wave 2-A) | <SPEC-A-N> sync-up: status <old> → <new>, walker LSGF-001 post-fix evidence aligned. |`
- **Tool**: Edit (sed 금지)
- **Commit**: `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-A — Category A 5건 frontmatter sync-up`

### T2-B-sync: Category B sync-commit mechanism apply (Wave 2-B-sync)

Wave 1 에서 sync-commit mechanism 채택된 SPECs 일괄 처리:

- **Mechanism**: empty body commit (또는 minimal body) 로 `sync(spec): SPEC-X — status closeout under FOLLOWUP-002` form 추가. frontmatter 미수정.
- **Predicted targets**: B6, B9, B10, B11 (run-phase Wave 1 결정에 따라 변경 가능)
- **Commit pattern** (multiple sync-commits one per SPEC OR single grouped commit — run-phase 결정):
  - Option A (preferred): 각 SPEC 별 단일 commit. 4 commits 예상.
  - Option B: 단일 commit 에 multi-SPEC body. 1 commit 예상.
- **Plan-auditor 선호**: Option A (개별 commit 으로 walker 가 명확히 인식). 그러나 PR squash 후 단일 commit 으로 압축됨.

### T2-B-skip: Category B lint.skip mechanism apply (Wave 2-B-skip)

Wave 1 에서 lint.skip 채택된 SPECs frontmatter 수정:

- **Mechanism**: 각 SPEC frontmatter 에 `lint.skip: [StatusGitConsistency]` 추가 + HISTORY entry 에 reason 기록
- **Predicted targets**: B1, B4, B7, B8, B12 (run-phase 결정 변경 가능)
- **Files**: 각 SPEC `.moai/specs/SPEC-X/spec.md` (frontmatter 만)
- **HISTORY entry pattern**:
  ```
  | 0.X.Y | 2026-05-16 | manager-develop (FOLLOWUP-002 Wave 2-B-skip) | walker visibility 한계 (synced via bulk-closure PR #XXX / prefix not recognized / chain rolled up) — lint.skip [StatusGitConsistency] 등록. |
  ```
- **Commit**: `docs(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-B (skip) — Category B lint.skip 등록 + reason 기록`

### T2-B-downgrade: Category B frontmatter downgrade (Wave 2-B-downgrade)

Wave 1 에서 frontmatter downgrade 채택된 SPECs 수정:

- **Mechanism**: 각 SPEC frontmatter `status` 하향 + `updated: 2026-05-16` 갱신 + HISTORY entry reason 기록
- **Predicted targets**: B2 (RT-001), B5 (SPC-003) — Big gap 군 중 외부 완료 증거 부재 시
- **Note**: run-phase 시점 외부 완료 증거 검증 후 채택. 만약 모든 Big gap 군이 외부 증거 있다면 sync-commit 또는 lint.skip 으로 처리되어 downgrade 0건 가능.
- **HISTORY entry pattern**:
  ```
  | 0.X.Y | 2026-05-16 | manager-develop (FOLLOWUP-002 Wave 2-B-downgrade) | git log + 외부 증거 검토 결과 implementation 미완료 — frontmatter status `implemented → planned` 정정. |
  ```
- **Commit**: `fix(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 Wave 2-B (downgrade) — Category B frontmatter 정정`

---

## Wave 3 — Verification

### T3-1: Lint clean confirmation

- **Command**: `moai spec lint --strict 2>&1 | tail -1`
- **Expected**: `0 error(s), 0 warning(s)`
- **Failure path**: 만약 still > 0 warning 시 → 해당 SPEC mechanism 재검토 → Wave 2 추가 commit
- **AC verification**: AC-SDF002-X-001

### T3-2: LSGF-001 회귀 부재 확인

- **Command**: `moai spec lint --strict 2>&1 | grep -E "SPEC-V3R4-HARNESS-00[123]\b" | wc -l`
- **Expected**: `0`
- **AC verification**: AC-SDF002-X-002

### T3-3: Scope discipline 확인

- **Command**: `git diff main...HEAD --name-only | grep -vE "^\.moai/specs/.+/(spec|plan|acceptance|design|tasks|progress)\.md$" | grep -v "^CHANGELOG\.md$" | wc -l`
- **Expected**: `0`
- **AC verification**: AC-SDF002-X-003

### T3-4: Per-SPEC AC verification (17건 individual check)

각 SPEC 별:

- **Command**: `moai spec lint --strict 2>&1 | grep "<SPEC-ID>"`
- **Expected**: empty output
- **AC verification**: AC-SDF002-A-1..A-5, AC-SDF002-B-1..B-12

### T3-5: progress.md 작성

`.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/progress.md` 생성. plan.md §7.3 형식 따라 Wave 1/2/3 결과 inline 기록.

---

## Sync-phase (post-run-PR merge)

### T-Sync-1: Open run-PR with auto-merge

- **Branch**: `feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002` (Wave 2 commits 모두 push)
- **Base**: `main`
- **Title**: `feat(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 run-phase — 17건 status drift binary 0 달성`
- **Body**: Wave 1 analysis summary + Wave 2 mechanism breakdown + Wave 3 AC verification logs
- **Auto-merge**: `gh pr merge <PR> --auto --squash`

### T-Sync-2: spec.md frontmatter sync (post-run-merge)

run-PR 머지 후:

- **Files**: `.moai/specs/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002/spec.md`
- **Changes**:
  - `status: draft → completed`
  - `updated: 2026-05-16`
  - `version: "0.1.0" → "0.2.0"`
  - HISTORY 0.2.0 entry 추가 (sync-phase 완료 + AC binary 0 달성 + lessons reference)
- **Commit**: `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 lifecycle COMPLETE — 17건 status drift binary 0 달성`

### T-Sync-3: CHANGELOG entry

`CHANGELOG.md` Unreleased 섹션에 entry 추가 (ko + en):

```markdown
### Fixed
- **Lint**: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — LSGF-001 walker word-boundary fix 후 노출된 17건 status drift 해소. Category A 5건 forward sync-up + Category B 12건 per-SPEC mechanism (sync-commit / lint.skip / downgrade). `moai spec lint --strict` 0 ERROR / 0 WARNING.

### Fixed (English)
- **Lint**: SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 — Resolved 17 status drift warnings exposed by LSGF-001 walker word-boundary fix. Category A 5 forward sync-ups + Category B 12 per-SPEC mechanisms (sync-commit / lint.skip / downgrade). `moai spec lint --strict` reports 0 error(s), 0 warning(s).
```

### T-Sync-4: Open sync-PR with auto-merge

- **Branch**: `sync/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002`
- **Base**: `main`
- **Title**: `sync(spec): SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 lifecycle COMPLETE — 17건 status drift binary 0`
- **Auto-merge**: `gh pr merge <PR> --auto --squash`

---

## Task Dependency Graph

```
T1-1 (baseline) → T1-2 (B git log) → T1-3 (bulk closure body) → T1-4 (prefix verify)
                                                                      ↓
                                                                T1-5 (ext evidence)
                                                                      ↓
                                                                T1-6 (mechanism decision)
                                                                      ↓
T2-A-1 ────────────────────┐
T2-B-sync ─────────────────┼─→ T3-1 (lint clean) → T3-2 (HARNESS regression) → T3-3 (scope)
T2-B-skip ─────────────────┤             ↓                                              ↓
T2-B-downgrade ────────────┘    T3-4 (per-SPEC AC)                              T3-5 (progress.md)
                                                                                      ↓
                                                                          T-Sync-1 (run-PR open)
                                                                                      ↓
                                                                          T-Sync-2 (spec sync)
                                                                                      ↓
                                                                          T-Sync-3 (CHANGELOG)
                                                                                      ↓
                                                                          T-Sync-4 (sync-PR open)
```

- T2-A-1, T2-B-sync, T2-B-skip, T2-B-downgrade 는 mutually independent (서로 다른 SPEC 들 수정) → parallel 가능. 그러나 main checkout 에서 sequential commit 권장 (commit history 명확성).
- T3-1, T3-2, T3-3 는 mutually independent → parallel 가능.

---

## Out of scope (Wave 4+ 미실시)

- 새 lint rule 신설
- walker code 변경
- `internal/spec/transitions.go` prefix expansion (docs(sync), fix(hook))
- 17건 외 SPEC frontmatter 일괄 audit
- v3.0.0-rc1 tagging (별도 release SPEC 또는 chore)
- CHANGELOG 4-locale 동기화 (CLAUDE.local.md §17 정책 — 별도 SPEC)
