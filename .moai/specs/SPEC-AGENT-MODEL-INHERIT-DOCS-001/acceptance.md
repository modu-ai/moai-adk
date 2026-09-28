# acceptance.md — SPEC-AGENT-MODEL-INHERIT-DOCS-001

Binary-testable criteria. Each AC carries a RED-now observation (measured on the baseline tree)
and the green path (which milestone flips it). All greps run from the repo root; doc paths are
`docs-site/content/<locale>/…` with `<locale>` ∈ {en, ko, ja, zh}.

## D. AC Matrix

### Baseline RED-now (tree `8a969dfc0`, branch `WT-model-docs-sweep`, measured 2026-09-29)

| # | Command (single invocation) | Verbatim output | Exit | Notes |
|---|---|---|---|---|
| R1 | `grep -c 'model_policy' internal/cli/wizard/translations.go` | `5` | 0 | H24 residue |
| R2 | `grep -c 'The three profiles\|Per-agent assignment table\|moai model profile' docs-site/content/en/multi-llm/model-policy.md` | `7` | 0 | A-cluster residue |
| R3 | `grep -c 'The three profiles\|Per-agent assignment table\|moai model profile' docs-site/content/en/advanced/profile-matrix.md` | `3` | 0 | A-cluster residue |
| R4 | `grep -cin 'per-spawn model injection' docs-site/content/en/cost-optimization/prompt-caching.md` | `1` | 0 | line 219 |
| R5 | `ls internal/cli/model.go` | `No such file or directory` | ≠0 | premise: accessor deleted |
| R6 | `grep -c 'Session model policy' internal/cli/profile_setup_translations.go` | `1` (line 165) | 0 | already clean — regression guard |

### AC-AMD-001 — A-cluster rewritten to inheritance narrative, 4 locales (REQ-AMD-001, 002)

Given the five A-cluster pages exist in all four locales with 4-locale existence symmetry,
When the M1 rewrite lands,
Then for each of the 20 files: (a) `grep -c 'The three profiles\|Per-agent assignment table\|moai model profile'` returns `0`; (b) `grep -c 'inherit'`-equivalent per-locale narrative markers exist (each locale's native wording for "subagents inherit the main session's model and effort" — per-locale marker list fixed in M1 evidence before editing); (c) the page's heading count parity holds across locales (same `##`/`###` count per page, per the hns-oss-docs-verify section-parity check).
RED-now: R2/R3 (en only; ko/ja/zh equivalents measured at M1 start).

### AC-AMD-002 — B-cluster deprecated-stub docs (REQ-AMD-003)

Given `getting-started/cli.md` documents `--model-policy`/`--profile` as live flags at en 72/116/452,
When the M2 rewrite lands,
Then each of the 5 getting-started/core-concepts pages × 4 locales describes the flags as deprecated stubs emitting a warning that names `moai profile setup` (per-locale presence of the `moai profile setup` string in the flag context), and no page presents them as live configuration.
RED-now: `grep -n 'model-policy' docs-site/content/en/getting-started/cli.md` hits lines 72/116 (measured via survey; re-measured at M2 start).

### AC-AMD-003 — cli-reference/profile.md disposition executed (REQ-AMD-004)

Given the page documents the surviving `moai profile` command family,
When M2 completes,
Then the page exists in all 4 locales (no removal), contains zero model-policy-matrix residue (`grep -c 'model profile\|profile matrix'` context-checked at edit time), and `docs-site/vercel.json` contains no new redirect for `cli-reference/profile`.
GREEN evidence: `git diff --stat` shows the 4 in-place edits; vercel.json diff shows no cli-reference/profile entry.

### AC-AMD-004 — redirect conditional (REQ-AMD-005)

Given the default disposition is zero page removals,
When the run phase completes,
Then either (a) `git diff docs-site/vercel.json` is empty AND `find docs-site/content -name '*.md' | wc -l` equals its M-start count per locale (no removals), or (b) every removed path has a vercel.json redirect entry in the same diff, recorded as an explicit deviation in the milestone evidence.

### AC-AMD-005 — prompt-caching bullets cleaned (REQ-AMD-006)

Given the per-spawn-injection bullet is measured at en cost-optimization line 219 plus a link description at 308,
When M3 completes,
Then `grep -cin 'per-spawn model injection' docs-site/content/<locale>/cost-optimization/prompt-caching.md` returns `0` for all four locales, the "per-agent model injection" link description is reworded, and the context-memory sibling's all-locale verification result is recorded (baseline en: 0 hits).
RED-now: R4.

### AC-AMD-006 — H24 wizard strings rewritten (REQ-AMD-007)

Given the 5 `model_policy` occurrences measured in `internal/cli/wizard/translations.go`,
When M4 completes,
Then `grep -c 'model_policy' internal/cli/wizard/translations.go` returns `0` (or only the retained map-key identifiers, with every user-visible title/description string carrying main-session wording — the run records which form holds), each locale's string is native wording naming the main-session effort fallback, `go build ./...` exits 0, and the affected `go test ./internal/cli/...` packages pass with observed output.
RED-now: R1.

### AC-AMD-007 — profile_setup regression guard (REQ-AMD-007)

Given `internal/cli/profile_setup_translations.go` carries main-session wording at baseline,
When M4 completes,
Then `grep -c 'Session model policy' internal/cli/profile_setup_translations.go` still returns ≥1 and no per-agent-assignment wording is introduced (`grep -cin 'assigning optimal models to each agent'` returns 0).
RED-now: R6.

### AC-AMD-008 — i18n HARD rules hold (REQ-AMD-008, 009)

Given the edits touch 44 docs-site files across 4 locales,
When the `hns-oss-docs-verify` recipe runs,
Then it passes: warning-free hugo build, sitemap existence, URL-blacklist grep clean, Mermaid direction grep clean (TD-only), 4-locale file-existence and section parity, body-emoji scan clean — each check's exit code observed, not assumed.

### AC-AMD-009 — Korean register (REQ-AMD-009)

Given Korean is the canonical locale authored first,
When M1-M3 land,
Then the ko copy reads as clean native written register (문어): no translationese (no English-syntax carry-over, no calqued figurative nouns), verified by a native-register read-through recorded in M1-M3 evidence; the check is judgment-based and its reviewer record is the evidence.

## D.1 Severity

AC-AMD-001, 002, 005, 006, 008: must-pass (release-blocking). AC-AMD-003, 004, 007: must-pass
(scoping/regression guards). AC-AMD-009: must-pass (quality judgment with recorded evidence).

## D.2 Traceability

REQ-AMD-001→AC-AMD-001; 002→AC-AMD-001; 003→AC-AMD-002; 004→AC-AMD-003; 005→AC-AMD-004;
006→AC-AMD-005; 007→AC-AMD-006+AC-AMD-007; 008→AC-AMD-008; 009→AC-AMD-009; 010→AC-AMD-008
(docs half) + AC-AMD-006 (Go half).

## D.3 Definition of Done

- All 10 REQs trace to a PASS AC with observed command output.
- M1-M5 milestone evidence files exist under the card's evidence path.
- Five-section evidence-bearing completion report delivered (VCI §3).
- Worktree disposes only after remote integration (card discipline).
