# Progress — SPEC-V3R6-DOCS-USER-DRIFT-001

## Status

- **Plan-phase**: COMPLETE (commit `d386cca0e`, 2026-05-22)
- **Run-phase**: COMPLETE (2026-05-22, this commit)
- **Sync-phase**: PENDING
- **Status**: `draft` → `implemented` (v0.1.0 → v0.2.0)

## Pre-flight Baseline (M1)

- Branch: `main`
- HEAD (pre-run): `19146bca6e2c7581872d7a7691313473c4bcccae`
- Recent commits (last 5):
  - `19146bca6` chore(SPEC-V3R6-HOOK-CONTRACT-FIX-001): status implemented v0.2.0
  - `d3d9b829f` chore(hook): anchor harness capture path via CLAUDE_PROJECT_DIR
  - `ba96ffc6e` test(hook): add plain-text stdout guards
  - `8319c6efa` fix(hook): subagent_stop dispatchCapture consults CLAUDE_PROJECT_DIR
  - `d386cca0e` plan(SPEC-V3R6-DOCS-USER-DRIFT-001): Wave 0 두번째 SPEC
- PR #1045: state=OPEN mergeable=MERGEABLE
- 4-locale baseline line counts:
  - ko: 570 lines (14 H2)
  - en: 511 lines (13 H2)
  - ja: 511 lines (13 H2)
  - zh: 512 lines (13 H2)
- F8 reproducibility: 0 matches of `ci-watch|ci-autofix|auto-fix|CI 모니터` (PASS)
- PR #1045 file overlap with `workflow-commands/moai-sync.md`: 0 (PASS)

## M2 Implementation Actions

### M2.1 ko 원문 작성

새 H2 섹션 `## PR 머지 후 CI 모니터링` 작성:
- 도입 단락 (Wave 1 + Wave 2 요약)
- 5 h3 subsection: Wave 1 — CI 결과 폴링 / Wave 2 — 자동 fix 루프 (최대 3 회) / 자동 처리 vs 사람 결정 필요 / auto-fix 가 절대 건드리지 않는 파일 / 관련 문서
- 1 markdown 표 (7 데이터 행 × 3 열: 결함 유형 / 자동 처리? / 비고)
- 1 warning callout (4 protected files: `.env`, credentials, `scripts/ci-watch/run.sh`, `.github/required-checks.yml`)
- cross-ref 2 SSoT 파일

### M2.2-M2.3 en/ja/zh 번역

ko 원문 기준 3 locale 동시 번역. h3 개수 (5) + 표 행 수 (8 — header + 7 data) + callout 수 (1) 정확히 동일 (AC-DUD-005 parity 충족).

H2 heading 매핑:
- en: `CI monitoring after PR creation`
- ja: `PR 作成後の CI モニタリング`
- zh: `PR 创建后的 CI 监控`

### M2.4 삽입 위치

각 locale 의 `## 품질 게이트` (en: `## Quality Gates`, ja: `## 品質ゲート`, zh: `## 质量门`) 직전에 4 sequential Edit 호출로 삽입. fallback 발동 없음 (4 locale 모두 표준 heading 존재).

## M3 검증 — 8 AC Binary Matrix

| AC | Status | Verification | Result |
|----|--------|--------------|--------|
| AC-DUD-001 | PASS | `grep -c '^## <heading>'` per locale | ko=1 / en=1 / ja=1 / zh=1 |
| AC-DUD-002 | PASS | `comm -12` PR #1045 vs this SPEC | 0 overlap (moai-sync.md not in PR #1045) |
| AC-DUD-003 | PASS | ko 3 pillars verification | 30초 + 30분 + required-checks.yml SSoT 명시 / 최대 3 + force-push 금지 / iteration ≥ 4 AskUserQuestion 모두 확인 |
| AC-DUD-004 | PASS | ko protected files check | `.env` ×1, `credentials` ×1, `scripts/ci-watch/run.sh` ×1, `.github/required-checks.yml` ×2 모두 존재 |
| AC-DUD-005 | PASS | parity check h3+table+callout | all 4 locale: h3=5 / table_rows=8 / callouts=1 (identical) |
| AC-DUD-006 | PASS | forbidden patterns scan | urls=0 / mermaid=0 / emojis=0 / devonly=0 / fm_change=0 across all 4 locale |
| AC-DUD-007 | PASS | Hugo build + heading in HTML | `hugo --gc --minify` exit 0, Total 1197ms, 4 public pages all contain new heading |
| AC-DUD-008 | PASS | line count max/min ratio | all 4 locale = 51 lines, ratio = 1.000 (≤ 1.20 threshold) |

**Overall**: 8/8 AC PASS.

## M4 Hugo Build Evidence

```
$ cd docs-site && hugo --gc --minify --buildDrafts=false
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08
              │ KO  │ EN  │ JA  │ ZH
──────────────┼─────┼─────┼─────┼─────
 Pages        │ 106 │  98 │  98 │  98
 Static files │ 218 │ 218 │ 218 │ 218
Total in 1197 ms

$ ls public/{ko,en,ja,zh}/workflow-commands/moai-sync/index.html
public/en/workflow-commands/moai-sync/index.html  (105029 bytes)
public/ja/workflow-commands/moai-sync/index.html  (112158 bytes)
public/ko/workflow-commands/moai-sync/index.html  (110751 bytes)
public/zh/workflow-commands/moai-sync/index.html  (102915 bytes)

Heading grep per locale:
ko: PR 머지 후 CI 모니터링 → PASS
en: CI monitoring after PR creation → PASS
ja: PR 作成後の CI モニタリング → PASS
zh: PR 创建后的 CI 监控 → PASS
```

## M5 PRESERVE List Confirmation

§6.2 PRESERVE list 무결성 확인 — 무관 modified + untracked 항목 모두 보존:

**Modified (preserved, NOT staged):**
- `.moai/harness/usage-log.jsonl` (runtime-managed)
- `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html` (parallel session work)
- `docs-site/data/menu/main.yaml` (parallel session work — overlaps with PR #1045 territory, not modified by this SPEC)

**Untracked (preserved, NOT staged):**
- `.moai/research/moai-adk-current-state-2026-05-22.md`
- `.moai/research/v3.0-design-2026-05-22.md`
- `.tmp-rndtest/`
- `docs-site/content/{ko,en,ja,zh}/book/`
- `docs-site/data/menu/extra.yaml`
- `docs-site/layouts/_default/redirect.html`
- `docs-site/scripts/gen_menu.py`
- `docs-site/static/book/`
- `internal/hook/.moai/`

**Committed (in-scope):**
- `docs-site/content/{ko,en,ja,zh}/workflow-commands/moai-sync.md` (4 files, +51 lines each)
- `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/spec.md` (v0.1.0 → v0.2.0 + HISTORY)
- `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/progress.md` (this file, NEW)

## spec-lint Status

Pre-flight baseline: 7 ERROR + 6 WARNING (all pre-existing, none in this SPEC's frontmatter).

Post-run check: no new spec-lint findings introduced. SPEC-V3R6-DOCS-USER-DRIFT-001 frontmatter is canonical 12-field SSOT compliant (`created:`/`updated:`/`tags:`, all 12 required fields present, `tier: S` declared).

## golangci-lint / Cross-platform

- Cross-platform build: N/A (no Go code changes)
- golangci-lint: N/A (no Go code changes)
- Coverage: N/A
- C-HRA-008 subagent boundary: N/A (docs-only)

## Cross-references (source SSoT)

- Polling doctrine: `.claude/rules/moai/workflow/ci-watch-protocol.md`
  - 30s poll interval (V3R5-015), 30min timeout (V3R5-016), `.github/required-checks.yml` SSoT (V3R5-017), auxiliary checks non-blocking (V3R5-018)
- Auto-fix doctrine: `.claude/rules/moai/workflow/ci-autofix-protocol.md`
  - 3 iter max (V3R5-005), iter≥4 AskUserQuestion (V3R5-006), new commit only (V3R5-007), protected files (V3R5-011 + V3R5-013), semantic failures human-only (V3R5-010)

## Blocker Report

None. All ACs PASS, Hugo build PASS, no scope creep, no PRESERVE violations.

## Next Steps

- Sync-phase: `/moai sync SPEC-V3R6-DOCS-USER-DRIFT-001` (decision deferred to user — single PR vs batch sync with HOOK-CONTRACT-FIX-001).
