# Progress — SPEC-V3R6-DOCS-CMD-CATALOG-001

## Run-phase Summary

- Start: 2026-05-22 (orchestrator-direct execution per Tier S LEAN optional template)
- End: 2026-05-22
- Baseline HEAD: `8da5367e5` (v3.0.0-rc1 version bump)
- Branch: `main` (Late-Branch policy — sync 단계에서 feat branch + cherry-pick)
- Commits (proposed 3-commit option B per plan.md §11.2):
  - C1: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): R1/R2 false promise 제거 (db/github 4-locale 삭제)`
  - C2: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): G1 harness 페이지 4-locale 신설`
  - C3: `docs(SPEC-V3R6-DOCS-CMD-CATALOG-001): G2 gate 페이지 4-locale 신설 + spec.md status implemented`

## AC Matrix (8/8 PASS)

| AC | Status | Verification | Actual Output | Expected |
|----|--------|--------------|---------------|----------|
| AC-DCC-001 | PASS | `ls .../moai-db.md \| grep -c "No such file"` | `4` | `4` |
| AC-DCC-002 | PASS | `for f in workflow-commands/_meta.yaml; do grep -q "moai-db" ...; done \| grep -c "OK:"` | `4` | `4` |
| AC-DCC-003 | PASS | `grep -c "/workflow-commands/moai-db" main.yaml` | `0` | `0` |
| AC-DCC-004 | PASS | files=4 meta=4 main=0 (R2 통합) | `files=4 meta=4 main=0` | `files=4 meta=4 main=0` |
| AC-DCC-005 | PASS | files=4 weight=4 verbs=4 (G1 harness) | `files=4 weight=4 verbs=4` | `files=4 weight=4 verbs=4` |
| AC-DCC-006 | PASS | meta=4 main=1 | `meta=4 main=1` | `meta=4 main=1` |
| AC-DCC-007 | PASS | files=4 weight=4 content=4 meta=4 main=1 (G2 gate) | `files=4 weight=4 content=4 meta=4 main=1` | `files=4 weight=4 content=4 meta=4 main=1` |
| AC-DCC-008 | PASS | forbidden_url=0 mermaid=0 emoji=0 draft=0 + hugo exit 0 | `forbidden_url=0 mermaid=0 emoji=0 draft=0` + `Total in 1164 ms` | `=0 0 0 0` + `hugo exit 0` |

**Total: 8/8 PASS (binary)**

## Files Affected (29 file operations)

### Deletes (8)
- `docs-site/content/ko/workflow-commands/moai-db.md`
- `docs-site/content/en/workflow-commands/moai-db.md`
- `docs-site/content/ja/workflow-commands/moai-db.md`
- `docs-site/content/zh/workflow-commands/moai-db.md`
- `docs-site/content/ko/utility-commands/moai-github.md`
- `docs-site/content/en/utility-commands/moai-github.md`
- `docs-site/content/ja/utility-commands/moai-github.md`
- `docs-site/content/zh/utility-commands/moai-github.md`

### Creates (8 + 1 progress.md = 9)
- `docs-site/content/ko/workflow-commands/moai-harness.md` (weight: 55, 4 verbs documented)
- `docs-site/content/en/workflow-commands/moai-harness.md`
- `docs-site/content/ja/workflow-commands/moai-harness.md`
- `docs-site/content/zh/workflow-commands/moai-harness.md`
- `docs-site/content/ko/quality-commands/moai-gate.md` (weight: 15, --fix/--staged/--file documented)
- `docs-site/content/en/quality-commands/moai-gate.md`
- `docs-site/content/ja/quality-commands/moai-gate.md`
- `docs-site/content/zh/quality-commands/moai-gate.md`
- `.moai/specs/SPEC-V3R6-DOCS-CMD-CATALOG-001/progress.md` (이 파일)

### Modifies — _meta.yaml (12 files, 16 entry-changes)
- workflow-commands × 4-locale: db remove + harness add (ko가 db 가진 유일 locale, en/ja/zh는 db 부재였음)
  - ko: 2 entry-changes (db remove + harness add)
  - en/ja/zh: 1 entry-change each (harness add only)
- utility-commands × 4-locale: github remove (4 entry-changes)
- quality-commands × 4-locale: gate add (4 entry-changes)
- Total: 2 + 3×1 + 4 + 4 = 13 entry-changes across 12 files (acceptance.md §4.2 명목값 16과 약간 차이 — ko workflow_meta baseline 비대칭으로 실제는 13)

### Modifies — main.yaml (1 file, regenerated via scripts/gen_menu.py)
- 4 block-changes 모두 적용됨:
  - db block removed
  - github block removed
  - harness block added (workflow-commands sub list, line 185-190 around)
  - gate block added (quality-commands sub list, line 247-252 around)
- Generator approach: `scripts/gen_menu.py` reads _meta.yaml + page frontmatter, regenerates main.yaml deterministically

### Modifies — SPEC artifacts (1 file)
- `spec.md` v0.1.1 → v0.2.0, `status: draft → implemented`, `updated: 2026-05-22`, HISTORY entry added

## Hugo Build Result

```
$ cd docs-site && hugo --gc --minify --buildDrafts=false
 images       │     │     │     │
 Aliases      │   6 │   7 │   7 │   7
 Cleaned      │   0 │   0 │   0 │   0

Total in 1164 ms
```

Exit code: 0 (PASS)

## Working Tree PRESERVE Verification

baseline dirty (`/tmp/dirty-before.txt`) → run-phase 종료 (`/tmp/dirty-after.txt`) diff:

**In-scope (28 files, expected)**:
- 12 M docs-site/content/*/_meta.yaml (4-locale × 3 categories)
- 1 M docs-site/data/menu/main.yaml (gen_menu.py regenerated)
- 8 D docs-site/content/*/{workflow-commands/moai-db,utility-commands/moai-github}.md
- 8 ?? docs-site/content/*/workflow-commands/moai-harness.md + quality-commands/moai-gate.md

**Out-of-scope external untracked (parallel session 산물, 본 SPEC 무관)**:
- `.tmp-parsetest/` (Go file, 18:50 timestamp)
- `docs-site/content/{ko,en,ja,zh}/book/` (4 dirs, book landing pages)
- `docs-site/layouts/_default/redirect.html` (Hugo layout)
- `docs-site/static/book/` (book static assets)

**PRESERVE PASS**: spec.md §6.2 enumerated list 모두 보호 ((`.moai/harness/usage-log.jsonl`, `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html`, `.moai/research/*`, `.moai/specs/SPEC-V3R5-*`, `docs-site/scripts/`, `internal/hook/.moai/`) — 모두 변경 0.

## Cross-platform / Subagent Boundary

- **Cross-platform**: n/a (콘텐츠 변경, OS-independent)
- **C-HRA-008 grep**: n/a (no harness/hook Go files touched)
- **spec-lint**: spec.md frontmatter canonical 12-field SSOT 준수 (`title`, `phase`, `module`, `lifecycle`, `created`/`updated`/`tags` snake_case 부재)

## NEW MINOR/SHOULD 흡수 (iter 2 plan-auditor)

- **D-NEW-01 (file count wording 29 vs 31)**: progress.md에 정확한 분해 표 인용 (29 file operations + 1 progress.md = 30 actual touch; spec.md HISTORY v0.2.0에서 29 file operations로 통일).
- **D-NEW-02 (8 create vs 9 incl. progress.md)**: progress.md `### Creates (8 + 1 progress.md = 9)` 명시.
- **D-NEW-03 (provisional SPECs)**: SPEC-V3R6-DOCS-SSOT-CHECK-001 / SPEC-V3R6-DOCS-WEIGHT-NORM-001 가칭 보존 (out of scope per spec.md §3.5).
- **D-NEW-04 (4-locale parity indirect coverage)**: AC-001/004/005/007 각각 `=4` count로 verification 완료.

## Operational Notes

### main.yaml regeneration via scripts/gen_menu.py

Run-phase 중 main.yaml을 처음 manual edit (3 Edits)로 수정했으나, 외부 trigger (parallel session 또는 hugo-related script)에 의해 main.yaml이 부분 재생성되는 현상 발견. 안정성 확보를 위해 `python3 docs-site/scripts/gen_menu.py > docs-site/data/menu/main.yaml`로 재생성 — 결과는 deterministic하며 4 block-changes (db/github remove + harness/gate add) 모두 정확히 반영됨.

### ko workflow _meta.yaml 비대칭 처리

ko/en/ja/zh workflow-commands/_meta.yaml 중 baseline에서 ko만 `moai-db` entry 보유 (en/ja/zh는 이미 부재). 본 SPEC에서 ko의 db entry 제거 + 4-locale 모두 harness entry 추가 처리. acceptance.md §4.2 명목값 (16 entry-changes)는 4-locale 모두 db 보유 가정이었으나, 실제는 13 entry-changes (ko 2 + en/ja/zh 각 1 + utility 4 + quality 4 = 13). AC-DCC-002 (4 OK) + AC-DCC-006 (meta=4) + AC-DCC-007 (meta=4) 모두 PASS.

## Definition of Done

- [x] All 8 ACs PASS (binary)
- [x] Hugo local build exit 0
- [x] Working tree PRESERVE list 보호 검증 PASS
- [x] SPEC frontmatter status: draft → implemented (v0.1.1 → v0.2.0)
- [x] progress.md 생성 (이 파일)
- [x] Self-verification report orchestrator로 반환

## Blocker Report

None. 본 SPEC scope 내 모든 acceptance criteria 충족.

## Next Steps (orchestrator 결정)

- **Sync 전략** (사용자 결정 필요):
  - (a) 단독 PR: 본 SPEC만 별도 PR (Late-Branch Phase C: stash → feat branch from origin/main → cherry-pick 3 commits → push -u → gh pr create)
  - (b) Batch sync: 누적 SPECs (V3R6-AGENT-FOLDER-SPLIT-001 등 진행 중)과 단일 PR
  - (c) 외부 untracked 정리 SPEC 선행: parallel session 산물 (`.tmp-parsetest/`, `book/` dirs, `redirect.html`, `static/book/`) 별도 SPEC으로 정리 후 본 SPEC sync

- **다음 SPEC 후보** (out of scope per spec.md §3.4/§3.5):
  - SPEC-V3R6-DOCS-WEIGHT-NORM-001 (가칭): 기존 페이지 weight 충돌 8건 정렬
  - SPEC-V3R6-DOCS-SSOT-CHECK-001 (가칭): docs ↔ skill drift 자동 detect
  - SPEC-V3R6-DOCS-HIGH-MEDIUM-DRIFT-001 (가칭): baseline 보고서 High/Medium 5건 정정 (cg-mode title alias / statusline preset 별칭 등)
