---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Research — verified ground truth for the wizard restructure + persistence chain"
version: "0.2.1"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: L
---

# Research — SPEC-CLI-WIZARD-RESTRUCTURE-001

Added at v0.2.0 with the Tier M → L re-tier. Every measurement below was
executed against the dedicated worktree `moai-adk-go-wt-wizard` on 2026-07-25
during the plan-audit review-2 fold, and is recorded with its command so a
later phase can re-derive rather than re-litigate it
(`verification-claim-integrity.md` §2 baseline attribution).

Two review-2 auditor claims were **falsified** on re-verification; both are
recorded in §R4 and §R5 as corrections rather than carried forward.

## §R1 — The Page-3 persistence chain (three gates)

**Question:** is removing `init.go`'s `if result.StandardMode` gate sufficient
for a Page-3 answer to reach disk?

**Answer: no.** Three gates, verified independently:

| Gate | Location | Evidence |
|---|---|---|
| 1 | `internal/cli/init.go:464-477` | `// Apply Phase 1 wizard results (only when StandardMode was active)` at `:464`; `if result.StandardMode {` at `:465`; guards `ProjectMode`, `HarnessProfile`, `LSPEnabled`, `EnforceQuality`, `CoverageExemptionsEnabled`, `DesignEnabled`, `ClaudeDesignEnabled`; closes at `:477`. |
| 2 | `internal/core/project/initializer_expansion.go:25-28` | `func WritePhase1Configs(opts InitOptions, result *InitResult) error {` then `if !opts.StandardMode { return nil }`. `opts.StandardMode` seeded at `init.go:336` from the retired flags. |
| 3 | `internal/core/project/initializer.go:438` | `grep -rn 'WritePhase1Configs' --include='*.go' .` → the ONLY production reference besides the definition. It sits inside `generateConfigsFallback` (def `:334`), reached from `:175` only in the `else` branch of `if i.deployer != nil` (`:167-179`). |

**Gate 3 is stronger than "usually skipped".** `moai init` builds the
initializer at `internal/cli/init.go:529`:

```go
initializer := project.NewInitializer(deployer, mgr, nil)
```

`deployer` is assigned in **both** branches of `if shouldDistributeAll(cmd)`
(`:517-527`) — `template.NewDeployerWithRenderer(...)` or
`template.NewSlimDeployerWithRenderer(...)` — and every error path returns
early. There is no code path through `runInit` that produces a nil deployer.
Therefore `generateConfigsFallback`, and with it every Page-3 write, is
**structurally unreachable from `moai init`**. `WritePhase1Configs` today runs
only from its own unit tests.

**The both-paths precedent already exists in the same file.**
`writeReportConfig` is invoked at `initializer.go:195` as Step 3c, positioned
after the deployer if/else block, with the doc comment: *"This runs on BOTH the
deployer path (overriding the template default html+md with the wizard/flag-
selected value) and the fallback path (which does not emit report.yaml
otherwise)."* This is the shape the Page-3 writes need (plan.md C32).

## §R2 — The five Page-3 writers and their write semantics

`grep -n` over `internal/core/project/initializer_expansion.go`:

| Function | Line | Semantics | Target (deployed size) | Verdict |
|---|---|---|---|---|
| `writeProjectModeYAML` | 52 | read → `patchYAMLKey` → write | `project.yaml` (799 B, `.tmpl`) | safe today |
| `writeHarnessProfileYAML` | 82 | **wholesale** `os.WriteFile`, 2 lines | `harness.yaml` (8,165 B, static) | destructive |
| `writeLSPYAML` | 98 | **wholesale** `os.WriteFile`, 2 lines | `lsp.yaml` (11,306 B, `.tmpl`) | destructive |
| `writeQualityExpansionYAML` | 111 | read → `patchYAMLKey` + append-if-absent | `quality.yaml` (6,536 B, `.tmpl`) | safe today |
| `writeDesignYAML` | 157 | **wholesale** `os.WriteFile`, 4 lines | `design.yaml` (2,867 B, static) | destructive |

Verbatim from `writeLSPYAML`:

```go
content := fmt.Sprintf("lsp:\n  enabled: %t\n", opts.LSPEnabled)
```

Total exposure if the three wholesale writers become reachable without
conversion: **22,338 bytes** of deployed configuration destroyed per
`moai init`, including the 16-language `lsp.yaml` that the template
language-neutrality policy (CLAUDE.local.md §15) exists to protect.

Command: `ls -la internal/template/templates/.moai/config/sections/`

Note on terminology: `harness.yaml` and `design.yaml` carry **no `.tmpl`
suffix** — they are static template assets deployed verbatim, not Go-template
rendered. Only `lsp.yaml.tmpl` and `quality.yaml.tmpl` are rendered. The
clobber hazard is identical either way; the accurate term is
"template-**deployed**".

## §R3 — Shipped template defaults for the removed/derived settings

`grep -n` over `internal/template/templates/.moai/config/sections/`:

| Setting | File | Line | Shipped value | Consequence |
|---|---|---|---|---|
| `harness.default_profile` | `harness.yaml` | 7 | `"default"` | REQ-WIZ-012 satisfied by the template alone; no write path needed (plan.md C36) |
| `constitution.coverage_exemptions.enabled` | `quality.yaml.tmpl` | 54 | `false` | REQ-WIZ-013 satisfied by the template alone; no write path needed |
| `lsp.enabled` | `lsp.yaml.tmpl` | 45 | `false` | REQ-WIZ-010's flip to enabled-by-default is **inert for persisted state** unless the write path works — a concrete consequence of the R1 chain being broken |
| `design.enabled` | `design.yaml` | 8 | `true` | user-answerable; needs a working write path |
| `design.claude_design.enabled` | `design.yaml` | 44 | `true` | user-answerable; needs a working write path |
| `project.mode` | `project.yaml.tmpl` | 15 | `personal` | user-answerable; existing read-patch already correct |

The `quality.yaml.tmpl` case has a second-order finding: because the deployed
file **already contains** `coverage_exemptions:` (L52),
`writeQualityExpansionYAML`'s `yamlContains(content, "coverage_exemptions:")`
guard returns true on the deployer path and the append branch never fires. The
function would patch `constitution.enforce_quality` (a live Page-3 answer) but
would never write `coverage_exemptions.enabled`. Since REQ-WIZ-013 fixes that
setting at `false` and the template already ships `false`, this is benign for
this SPEC — but it is benign by coincidence, so the SPEC states it explicitly
rather than leaving it implicit (spec.md §C).

## §R4 — CORRECTION 1: the read-patch/wholesale split was mis-stated by the audit

**Audit claim (review-2 N1):** *"Only `writeProjectModeYAML` read-patches
(`patchYAMLKey`, `:64`)."*

**Verified: FALSE — undercount.** `writeQualityExpansionYAML` (`:111-152`) is
also a read-patch. Verbatim structure:

```go
if data, err := os.ReadFile(qualityPath); err == nil { existing = string(data) }
...
content = patchYAMLKey(existing, "constitution", "enforce_quality", fmt.Sprintf("%t", opts.EnforceQuality))
if !yamlContains(content, "coverage_exemptions:") { content += exemptBlock }
```

The split is **2 read-patch + 3 wholesale**, not 1 + 4. The audit's *repair
target count* (three functions) happened to be right; its characterization of
the remainder was not. Consequence for the plan: `enforce_quality` needs no
writer conversion, so C35 covers only `lsp` and `design`, and C36 drops
`harness` entirely rather than converting it.

**Why the distinction matters beyond bookkeeping:** had the plan carried the
audit's framing forward, a run-phase agent would likely have "converted"
`writeQualityExpansionYAML` too — churning a function that is already correct
and risking a regression in the one Page-3 answer (`enforce_quality`) whose
write path works today.

## §R5 — CORRECTION 2: `patchYAMLKey` is depth-blind; the audit's prescribed repair is unsafe

**Audit claim (review-2 N1 required-fix 2):** *"Convert `writeLSPYAML` /
`writeHarnessProfileYAML` / `writeDesignYAML` to read-patch (`patchYAMLKey`)."*

**Verified: the target helper cannot safely do this.** `patchYAMLKey`'s
matching and rewriting logic:

```go
stripped := trimLeadingSpaces(line)
...
if len(stripped) > len(key)+2 && stripped[:len(key)+2] == key+": " {
    lines[i] = "  " + key + ": " + newValue
}
```

It strips leading whitespace before matching, so it matches a key of that name
at **any** nesting depth inside the section; and it rewrites with a
**hardcoded 2-space indent**, flattening whatever it matches to depth 2. It
also does not `break`, so it rewrites every match.

Collision inventory against the real deployed files
(`grep -n '^ *enabled:' <file>`):

| File | `enabled:` occurrences | Target | Collateral damage under `patchYAMLKey` |
|---|---|---|---|
| `lsp.yaml.tmpl` | L45 (2sp), **L323 (4sp)** | L45 | L323 hoisted to depth 2 → `delegate_to_astgrep:` (L322) orphaned |
| `design.yaml` | L8 (2sp), **L25 (6sp), L44 (4sp), L55 (4sp), L76 (4sp)** | L8 + L44 | all five flattened to depth 2 → `gan_loop`, `claude_design`, `figma`, `adaptation` all orphaned |

The two existing callers are safe only by **accident of key uniqueness**:

```
grep -n '^ *mode:' project.yaml.tmpl          → 1 hit  (L15)
grep -n 'enforce_quality' quality.yaml.tmpl   → 1 hit  (L13)
```

Following the audit's prescription literally would trade a **visible clobber**
(a 2-line file where an 11 KB file was) for a **silent structural corruption**
(a plausible-looking YAML file with keys at the wrong depth) — strictly worse,
because the clobber is caught by a size assertion and the corruption is not.

Hence plan.md C34 introduces an additive depth-aware helper, spec.md adds
REQ-WIZ-021/022, and acceptance.md adds AC-WIZ-017 with an explicit
non-vacuity proof that distinguishes a correct implementation from the naive
one (`grep -c` indentation multiset: correct → `1/3/1` for design.yaml, naive
→ `5/0/0`).

## §R6 — Retirement residue inventory (N2/N5/N6 baselines)

| Measurement | Value | Command |
|---|---|---|
| `--standard` / `--advanced` under `.github/` | **0** (exit 1) | `grep -rn '\-\-standard\|\-\-advanced' .github/` |
| `StandardMode`/`AdvancedMode` repo-wide (Go) | 83 lines | `grep -rn 'StandardMode\|AdvancedMode' --include='*.go' .` |
| …outside `internal/cli/` | **15 lines** | `… internal/ \| grep -v '^internal/cli/'` |
| Breakdown of the 15 | `initializer.go` 2 (`:56`, `:437`) · `initializer_expansion.go` 4 (`:5,23,24,26`) · `initializer_expansion_test.go` 9 (`:34` false, `:60,85,116,161,184,219,260,344` true) | as above |
| Test files touching retired symbols | **7** | `grep -rln 'StandardMode\|AdvancedMode\|advanced_bridge\|harness_profile\|coverage_exemptions_enabled\|IsAdvancedWizardReady\|Phase2Questions\|RunWithDefaultsModes' --include='*_test.go' .` |
| The 7 files | `internal/core/project/initializer_expansion_test.go`, `internal/cli/init_test.go`, `internal/cli/wizard/{translations_completeness,questions,unified_form,wizard,expansion}_test.go` | as above |
| cobra flag registration idiom | `.Bool("standard"…)` / `.Bool("advanced"…)` — bare name, no `--`, not `BoolVar` | `grep -rn '\.Bool("standard"\|\.Bool("advanced"' internal/` → 4 hits (`init.go:84,85`, `init_update_notice_test.go:49,50`) |

`.github/` at 0 confirms review-2 N5: the plan's prior claim that CI scripts
"break" was a phantom. It is retained only as an expected-0 non-regression
guard in AC-WIZ-015.

## §R7 — The N4 grep-vacuity measurement

AC-WIZ-006's second grep, as written at v0.1.2:

```
grep -rn 'default: "high"|default is "high"|default.*"high"' internal/template/context.go internal/config/profile.go
```

**Executed: 1 match** — `internal/template/context.go:64`
(`ModelPolicy string // "high", "medium", "low" (default: "high")`), and
**zero** for `internal/config/profile.go`.

Cause: the pattern is case-sensitive; `profile.go:59`'s real text is

```
// The divergent legacy init-selection constant DefaultModelPolicy = "high"
```

— capital `D`, no lowercase `default`. A case-insensitive probe
(`grep -in 'default.*"high"'`) hits it. So the grep read 0 for `profile.go`
both before AND after the C11 edit, leaving C11 (a comment-only change)
unverified by any AC. Fixed in acceptance.md AC-WIZ-006 by anchoring on the
real token and adding `-i`; the corrected grep returns **2** against the
current tree, satisfying the standing non-vacuity requirement (§D.3).

## §R8 — File-count inventory driving the Tier M → L decision

Production (12):

```
internal/cli/wizard/questions.go          internal/cli/init.go
internal/cli/wizard/wizard.go             internal/cli/init_update_notice.go
internal/cli/wizard/translations.go       internal/template/model_policy.go
internal/cli/wizard/types.go              internal/template/context.go
internal/cli/wizard/advanced_gate.go(del) internal/config/profile.go
internal/core/project/initializer.go      internal/core/project/initializer_expansion.go
```

Test (7, per §R6), plus **3 NEW test files** made explicit at v0.2.1 (review-3
D2) — `internal/core/project/initializer_persist_test.go` (C39),
`internal/core/project/yaml_patch_test.go` (C40) and
`internal/cli/init_flag_precedence_test.go` (C41). Each is a NEW file rather
than an edit to an existing sibling so that M4/M5 deliverables do not collide
with M7's C37/C38 edits to `initializer_expansion_test.go` / `init_test.go`.
**Total 22.**

Tier M's envelope is 5-15 files. 22 exceeds it, and the excess is not
mechanical: it is an architectural relocation of a config write onto a
different execution path, guarded by a data-loss hazard and requiring a new
YAML-patch primitive. See design.md §D4 for the tier rationale.

## §R9 — SELF-CAUGHT: the v0.2.0 AC-WIZ-007 locale grep was itself vacuous for zh

Recorded because it is the review-2 N4 failure mode recurring **inside the fix
for N4**, caught by applying the standing non-vacuity rule (acceptance.md §D.3)
to a grep this fold had just authored.

The first draft of AC-WIZ-007's locale grep was:

```
grep -c '기본값: 아니오\|デフォルト: いいえ\|默认: 否' internal/cli/wizard/translations.go
```

**Executed: 6 matches** — not the 9 the file actually contains. The three zh
entries were missed. Cause: **zh writes a FULLWIDTH colon**.

| Locale | Actual title text | Parens | Colon |
|---|---|---|---|
| ko | `"LSP 통합을 활성화할까요? (기본값: 아니오)"` | halfwidth `()` | halfwidth `:` |
| ja | `"LSP 統合を有効にしますか？（デフォルト: いいえ）"` | **fullwidth `（）`** | halfwidth `:` |
| zh | `"启用 LSP 集成？（默认：否）"` | **fullwidth `（）`** | **fullwidth `：`** |

The corrected colon-agnostic form returns the full 9:

```
grep -cE '기본값[：:] *아니오|デフォルト[：:] *いいえ|默认[：:] *否' internal/cli/wizard/translations.go   → 9
```

Distribution: ko L43/125/133, ja L155/237/245, zh L267/349/357 — exactly 3 per
locale, and all 9 belong to the three questions this SPEC removes or flips
(`advanced_bridge`, `lsp_enabled`, `coverage_exemptions_enabled`), so the
post-change expectation is a clean **0**.

Had the halfwidth-only pattern shipped, AC-WIZ-007 would have reported the
zh locale clean while `启用 LSP 集成？（默认：否）` still told Chinese users the
default was off — a 4-locale parity break invisible to its own acceptance
check. Fixed in acceptance.md AC-WIZ-007 (character class `[：:]` + a measured
9/3 non-vacuity proof), in plan.md C13 (exact per-locale punctuation recorded
so run-phase does not guess), and in plan.md §G (new anti-pattern:
halfwidth-punctuation greps over CJK locales).
