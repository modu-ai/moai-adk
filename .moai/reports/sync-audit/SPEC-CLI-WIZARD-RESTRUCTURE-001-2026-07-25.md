# Sync-Phase Independent Audit — SPEC-CLI-WIZARD-RESTRUCTURE-001

- **Auditor**: sync-auditor (independent, adversarial stance)
- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/wizard`
- **Branch / HEAD**: `feat/SPEC-CLI-WIZARD-RESTRUCTURE-001` @ `da85ad13e`
- **Tier**: L — PASS threshold **0.85** (harmonic mean)
- **Scoring model**: flat weighted 4-dimension (no `evaluator_mode: hierarchical` in `harness.yaml`)

---

## 1. Verdict

| Dimension | Score | Verdict |
|-----------|-------|---------|
| Functionality | **0.80** | PASS (below Tier-L bar) |
| Security | **0.70** | PASS (no Critical/High — no hard-threshold trip) |
| Craft | **0.82** | PASS |
| Consistency | **0.68** | FAIL |

**Harmonic mean = 4 / (1/0.80 + 1/0.70 + 1/0.82 + 1/0.68) = 0.745**

## OVERALL VERDICT: **FAIL** (0.745 < 0.85)

The FAIL is **score-driven, not must-pass-driven**. The Security dimension records
no Critical/High OWASP finding, so the hard threshold does not trip. All 19 run-phase
acceptance criteria genuinely pass and were independently reproduced. The failure is
concentrated in the **sync-phase deliverable**, where one tool-confirmed critical defect
(the 3-phase close marker is not machine-readable) and several artifact-vs-tree
contradictions pull Consistency to 0.68 and drag the harmonic mean below the bar.

**Null hypothesis considered**: *"did the sync phase actually change anything that
matters, or did it merely restate the run phase?"* — Answer: it did real work. The
12-file 4-locale docs-site rewrite is substantive (44 stale flag references → 0,
semantic parity verified across ko/ja/zh, hugo build warning-free). It is not a
restatement. But it also shipped a broken close marker and a CHANGELOG claim the
tree contradicts.

---

## 2. Dimension detail

### 2.1 Functionality — 0.80

**All 19 ACs re-derived from the CURRENT tree.** I did not trust progress.md §E.2.2.
Where an AC states a grep with an expected count, I ran the grep. Where an AC binds a
test, the test ran inside my own `go test -count=1 ./...`. For AC-WIZ-010 / 010a / 017 I
built the binary from this tree and ran a real deployer-path `moai init`.

| AC | Independent verification performed | Result |
|----|-----------------------------------|--------|
| AC-WIZ-001 | `questions.go` inspected: no `advanced_bridge`; 3 Group labels | PASS |
| AC-WIZ-002 | Page membership read from source: Basic=[conversation_language, user_name, project_name], Model & Report=[model_policy, report_format], Quality & Workflow=[lsp_enabled, enforce_quality, project_mode, design_enabled, claude_design_enabled] | PASS |
| AC-WIZ-003 | `Condition: func(r *WizardResult) bool { return r.DesignEnabled }` — no `StandardMode` | PASS |
| AC-WIZ-004 | `saveAnswer` carries `case "conversation_language"`; test-bound, suite green | PASS |
| AC-WIZ-005 | `Default: "medium"`; exactly one `(Recommended)` label, on value `"medium"`; `Value: "high"` still present | PASS |
| AC-WIZ-006 | `DefaultModelPolicy = ModelPolicyMedium` (model_policy.go:26); case-insensitive grep on context.go+profile.go → **exit 1 / 0 matches**; `ModelPolicyHigh` const retained | PASS |
| AC-WIZ-007 | `Default: "true"`; CJK-colon-class grep on translations.go → **0**; `default: No` in questions.go → **0** | PASS |
| AC-WIZ-008 | `harness_profile` absent from question set; **real init**: `harness.yaml` → `default_profile: "default"` | PASS |
| AC-WIZ-009 | `coverage_exemptions_enabled` absent; template default `enabled: false` intact | PASS |
| AC-WIZ-010 | **Real `moai init --non-interactive --project-mode team --enable-lsp --enforce-quality=false --enable-design=false`**: `project.yaml → mode: team`; `lsp.yaml → enabled: true`; `quality.yaml → enforce_quality: false`; `design.yaml → enabled: false`. Structural: `if !opts.StandardMode` → 0; `WritePhase1Configs` called at `initializer.go:213`, outside `generateConfigsFallback` | PASS |
| AC-WIZ-010a | Post-init sentinels all present (`delegate_to_astgrep`, `circuit_breaker`, `figma`, `brand_context`, `default_profile`); sizes **lsp 11305 B / design 2868 B / harness 8165 B** — matching progress.md §E.2.4 exactly, and all above the stated floors | PASS |
| AC-WIZ-011 | `advanced_bridge|harness_profile|coverage_exemptions_enabled` in translations.go → **0** | PASS |
| AC-WIZ-012 | `ReconfigureQuestions` splices Git block immediately after `report_format` | PASS |
| AC-WIZ-012a | `ReconfigureQuestions = DefaultQuestions + Git` — `Page3Questions` never appended | PASS |
| AC-WIZ-013 | `saveAnswer`/`saveBoolAnswer` case list contains none of the 3 removed IDs | PASS |
| AC-WIZ-014 | `go test -count=1 ./...` **exit 0, 105 ok, 0 FAIL**; wizard **93.9%**, core/project **88.7%**; `golangci-lint run --timeout=3m` **exit 0, "0 issues."** | PASS |
| AC-WIZ-015 | `advanced_gate.go` GONE; all six `internal/`-scoped greps → **0**; `.github/` → **0** | PASS |
| AC-WIZ-016 | `applyWizardPage3ToOpts` uses `cmd.Flags().Changed(name)` for all four settings (not value comparison); test-bound, suite green | PASS |
| AC-WIZ-017 | **Real init** indentation multisets: design.yaml `{2sp:1, 4sp:3, 6sp:1}`, lsp.yaml `{2sp:1, 4sp:1}` — exactly the depth-aware expectation, refuting the naive `patchYAMLKey` shape `{2sp:5}` / `{2sp:2}` | PASS |

**19/19 AC PASS is CONFIRMED, not merely accepted.** This is the strongest part of the work.

Deductions against 1.00:

- **F1 (critical)** — the sync-phase 3-phase close is not machine-readable (§3 F1). The
  domain's dedicated tool (`moai spec audit`) classifies this SPEC **V3R5**, not V3R6.
  This is a functional failure of the sync deliverable itself, not covered by any AC
  (the ACs are run-phase-scoped).
- **F2 (major)** — `--harness-profile` is a documented, accepted, and re-documented
  CLI flag that writes nothing (§3 F2).
- **F3 (major)** — non-interactive `moai init` yields `lsp.enabled: false` while the
  interactive wizard default is now `true` (§3 F3).

### 2.2 Security — 0.70

I read the implementation rather than the SPEC prose, then probed the boundary with the
real binary.

**What holds up:**

- `patchYAMLPathValue` is sound as a primitive. Path segments are compile-time constants
  (`"lsp.enabled"`, `"design.enabled"`, `"design.claude_design.enabled"`) — no
  user-controlled key path, so no path-injection surface into the key matcher. Only the
  **first** match is rewritten. When the path is absent it returns `ok=false` and the
  caller leaves the document **byte-identical** rather than appending a duplicate
  top-level mapping — a deliberate and correct anti-corruption choice.
- **No path traversal.** `sectionsDir = filepath.Clean(filepath.Join(opts.ProjectRoot, defs.MoAIDir, defs.SectionsSubdir))`;
  every filename is a `defs` constant. No user string reaches a path component.
- **The "non-destructive patch" claim is TRUE**, verified empirically rather than from
  prose: sentinel keys survive, byte sizes land at 11305/2868/8165 (total 22 338 B — the
  exact figure the SPEC says was at risk), and `project.yaml` diffs against a control run
  by exactly the two expected lines. `harness.yaml` is byte-untouched.
- **No secrets** in the 12 sync-edited docs-site files (`ghp_|github_pat_|glpat-|sk-ant|AKIA|BEGIN … PRIVATE KEY` → 0 matches).
- File mode `0o644` on non-secret config files is appropriate.

**What does not:**

- **S1 (major, CONFIRMED, reproduced)** — YAML injection via unvalidated `--project-mode`
  (§3 S1). `validateInitFlags` validates `--mode`, `--git-mode`, `--git-provider`,
  `--model-policy`, `--profile`, `--github-username`, `--gitlab-instance-url` — but
  **not** `--project-mode` or `--harness-profile`. The value is interpolated raw into
  `"  " + key + ": " + newValue`. This SPEC (C32) is what made that write path reachable
  from the CLI at all, so the gap is newly exposed by this change. Exploitability is low
  (local, self-inflicted, no privilege boundary crossed), which is why it is not scored
  as Critical/High and does not trip the hard threshold — but it is a TRUST-5 "Secured"
  input-validation failure at a write boundary.
- **S2 (major)** — the patch is written with non-atomic `os.WriteFile` (7 sites in
  `initializer_expansion.go`) even though this repo ships `internal/atomicfile.Replace`
  precisely for write-temp-then-rename durability. An interrupt or crash mid-write
  truncates the 11 KB `lsp.yaml` — destroying the very file AC-WIZ-010a exists to
  protect. The AC guards against *logical* clobber but not against *durability* clobber.
- **S3 (minor)** — unbounded `os.ReadFile` of the target config before patching. Bounded
  in practice by template-deployed sizes; no limit enforced.

### 2.3 Craft — 0.82

**`patchYAMLPathValue` is a defensible design, and design.md §D2 option C is not
obviously the better future fix.** A real YAML round-trip (`gopkg.in/yaml.v3`) would
destroy the comments and key ordering of an 11 KB hand-annotated deployed config — the
exact property the SPEC is trying to preserve. The line-walker is the *right* shape for
"patch one value, preserve every other byte". It is well-documented (the godoc explains
why `patchYAMLKey` was not reused), carries a clear `(string, bool)` contract, and its
scope-stack unwinding (`for len(stack) > 0 && stack[len(stack)-1].indent >= indent`) is
correct. I would call the block-scalar limitation genuinely minor and correctly
disclosed. **Recommend downgrading design.md §D2 option C from "the documented future
fix" to "an option with a comment-loss cost" — as written it understates the tradeoff.**

**The tests are meaningful, not vacuous.** I checked for the failure mode where tests
assert tautologies:

- `restructure_test.go` `expectedGroupCount` **independently restates the
  `buildFormGroups` partition contract** instead of hardcoding a number, so it keeps
  verifying that merging actually happens as questions change. This is above-average
  test design.
- `question_removal_test.go` is refutation-style: it *calls* `saveAnswer("harness_profile", …)`
  and asserts the result field stays zero — it would fail if the capture branch returned.
- `TestWritePhase1Configs_PatchesExistingFiles` had its non-vacuity proven by temporarily
  forcing the wholesale-write branch (recorded in §E.2.4, and the mechanism is visible in
  the test body).
- AC-WIZ-017's assertion distinguishes `{2sp:1, 4sp:3, 6sp:1}` from the naive
  `{2sp:5}` — it genuinely discriminates the depth-aware implementation from the
  depth-blind one. I reproduced the passing side with the real binary.

**Craft deductions:**

- **C1 (minor)** — `writeHarnessProfileYAML` retained as dead code *with two live tests*.
  Declared debt (a), and the §G-amendment rationale for not deleting is legitimate. But
  the two tests exercise a function with no production caller, contributing false
  coverage confidence to the 88.7% figure. Acceptable **as declared debt**, not as
  finished work.
- **C2 (major)** — stale-and-now-false code comment at `internal/cli/init.go:368`:
  *"Page-3 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008)"*.
  After REQ-WIZ-010 flipped the wizard `lsp_enabled` default to `true`, the sibling line
  `LSPEnabled: getBoolFlag(cmd, "enable-lsp")` (cobra default `false`) makes that comment
  factually wrong. Its two neighbours correctly use `getBoolFlagWithDefault(…, true)`.
- **C3 (minor)** — non-atomic writes despite an in-repo atomic primitive (see S2).
- **C4 (informational)** — the 4 removal-tombstone comments are fine as declared debt
  (e); they carry no AC-WIZ-015 token and cannot defeat that grep. I verified this.

### 2.4 Consistency — 0.68

**What is genuinely consistent (and I checked semantically, not by line count):**

- The rewritten docs-site content **matches the actual question set in
  `internal/cli/wizard/questions.go` item by item**: 3 pages, 10 questions, correct
  per-page membership and order, `claude_design_enabled` correctly described as nested on
  `design_enabled`, and the Git-questions-are-not-asked-at-init callout is accurate
  (`GitQuestions()` is reachable only from `ReconfigureQuestions`, confirmed in source).
- **4-locale parity is real, not superficial.** I read the rewritten region in ko / ja /
  zh, not just the counts: each carries the same "fixed 3-page flow, no mode flag" thesis,
  the same 3-row page table with the same per-page question lists, and the same Git
  auto-detection callout. Counts also match exactly (init-wizard 28 headings / 10 rows;
  cli.md 44 / 161; cli-reference/init.md 14 / 37 — identical across all four locales).
- `hugo --gc --minify` **independently re-run: exit 0, 0 WARN/ERROR lines**, 153/151/151/151 pages.
- `grep -rnE -- '--standard|--advanced' docs-site/content/` → **0**, against an
  independently re-measured baseline of **44 matches across 12 files** at `216dcbb2e`.
- The CHANGELOG's merge explanation is accurate — I enumerated the 9 commits merged by
  `96d35723c` and confirmed both `SPEC-DB-RETIRE-001` (#1155) and the prompt-cache
  retirement (#1148) are among them, which is exactly the stated cause of 107 → 105 packages.
- Debt items (a)–(e) in the CHANGELOG are all present in the tree and accurately described.

**Consistency defects — this is where the score is lost:**

- **F1 (critical)** — `sync_commit_sha` format deviates from every sibling SPEC and
  breaks the era parser (§3 F1).
- **F4 (major)** — CHANGELOG claim contradicted by the tree (§3 F4).
- **F5 (major)** — `design.md` and `research.md` still carry `status: in-progress`
  while the other four artifacts are `completed` (§3 F5).
- **F6 (minor)** — CHANGELOG cites `internal/template` coverage **84.9%**; measured
  **86.0%** at HEAD (§3 F6).
- **F7 (minor)** — in-repo contradiction between sync-edited docs-site files about the
  LSP default (§3 F7).

---

## 3. Findings

Legend — **CONFIRMED** = I reproduced it in this tree. **PLAUSIBLE** = suspected, not reproduced.

| # | Sev | Location | Finding | Evidence | Status |
|---|-----|----------|---------|----------|--------|
| **F1** | **critical** | `.moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/progress.md:402` | **The 3-phase close marker is not machine-readable; the SPEC is misclassified as grandfather era and is now permanently exempt from lifecycle drift detection.** The line is written as `- **`+"`"+`sync_commit_sha: adee2f46b`+"`"+`**` — bold-wrapped. Neither era.go extractor matches it: `progressFieldYAMLPattern` = `` `(?m)^\s*([^:\s]+)\s*:\s*(.+?)\s*$` `` requires the key at line start; `progressFieldListPattern` = `` "(?m)^\\s*[-*]\\s*`?([^:\\s`]+)`?\\s*:\\s*(.+?)\\s*$" `` allows at most **one optional backtick** after the bullet, not `**`+"`"+`. So `extractProgressField(…, "sync_commit_sha")` returns `""`, H-3 fires instead of H-4, and `era_final: true` / `IsModern() == false`. | `moai spec audit --json` (built from this tree) reports verbatim: `{'spec_id': 'SPEC-CLI-WIZARD-RESTRUCTURE-001', 'era': 'V3R5', 'finding_type': 'EraAutoDetected', 'severity': 'INFO', 'details': {'heuristic_matched': 'H-3 (§E.2 present, sync_commit_sha missing)'}}`. Every comparably-recent closed SPEC writes a bare line-start key and classifies V3R6: `SPEC-DB-RETIRE-001` `sync_commit_sha: 1c46c206e` → `H-4 (§E.2 + §E.4 + sync_commit_sha)`; likewise `SPEC-SUBAGENT-NESTING-DOCTRINE-001`, `SPEC-CLI-TUX-INIT-UPDATE-001`, `SPEC-CONFIG-AUDIT-REPAIR-001`. No `era:` frontmatter override is present (`grep -c '^era:' spec.md` → 0) to rescue the classification. | **CONFIRMED** |
| **F2** | major | `internal/cli/init.go:88,374`; `internal/core/project/initializer_expansion.go:91`; `docs-site/content/{en,ko,ja,zh}/getting-started/init-wizard.md:165`, `.../getting-started/cli.md:70`, `.../cli-reference/init.md:49` | **`--harness-profile` is an accepted, documented, silently inert flag.** It is registered, validated nowhere, and copied into `opts.HarnessProfile` — but its only consumer `writeHarnessProfileYAML` was removed from `WritePhase1Configs` by C36 and has **no production caller**. The sync phase then re-published `--harness-profile default` in the non-interactive CI example across all 4 locales, teaching users to pass a flag that does nothing. spec.md §C declares the *write path* out of scope, which justifies the code decision — it does not justify continuing to document the flag as functional. | Real binary: `moai init … --harness-profile strict` → `harness.yaml` still contains `default_profile: "default"`. `grep -rn 'HarnessProfile' --include='*.go' .` (non-test) shows the only reader is the dead `writeHarnessProfileYAML`. | **CONFIRMED** |
| **F3** | major | `internal/cli/init.go:89,371` | **Non-interactive and interactive `moai init` now disagree on the LSP default.** REQ-WIZ-010 flipped the wizard default to `true`, but the flag path still uses bare `getBoolFlag(cmd, "enable-lsp")` → cobra default `false`, while its two siblings use `getBoolFlagWithDefault(…, true)`. A CI/CD `moai init --non-interactive` silently gets LSP **disabled**, contradicting `init-wizard.md:123` ("The default is **enabled (Yes)**"). The immediately-preceding code comment asserting the defaults match is now false (see C2). | Real binary: `moai init $TMP --non-interactive` (no flags) → `lsp.yaml` → `enabled: false`. | **CONFIRMED** |
| **F4** | major | `CHANGELOG.md:28` | **CHANGELOG asserts a capability the tree does not support**: *"`harness_profile` remains selectable only as the pre-existing `--harness-profile` CLI flag, never as a wizard question."* "Selectable" implies the flag takes effect. It does not (F2). | Same evidence as F2. | **CONFIRMED** |
| **F5** | major | `.moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/{design,research}.md` frontmatter | **2 of the 6 Tier-L artifacts were left at `status: in-progress` on a `completed` SPEC.** progress.md §E.4 `frontmatter_status_transitions` explicitly scopes the transition to 4 artifacts, so this was deliberate — but it deviates from the observed repo convention: of the completed SPECs that ship a `design.md`, `SPEC-AGENT-ARCH-V2-001` and `SPEC-GLM-EFFORT-TUNE-001` both carry `design=completed`. This SPEC is the only one in the catalog showing `spec=completed design=in-progress`. | `for d in .moai/specs/SPEC-*/; do …` survey output: `SPEC-CLI-WIZARD-RESTRUCTURE-001  spec=completed design=in-progress` — unique in the catalog among design.md-bearing completed SPECs. | **CONFIRMED** |
| **S1** | major | `internal/cli/init.go` `validateInitFlags`; `internal/core/project/initializer_expansion.go:66` `patchYAMLKey` | **YAML injection through unvalidated `--project-mode`.** The flag value is interpolated raw into `"  " + key + ": " + newValue` with no enum validation, unlike the 7 sibling flags that are validated. A newline in the value escapes the `project:` mapping and writes an arbitrary top-level key. This SPEC's C32 is what moved this writer onto the reachable CLI path (it previously lived inside the structurally-unreachable `generateConfigsFallback`), so the SPEC newly exposes the gap. `--harness-profile` has the same missing validation (currently masked by F2's inertness). | Reproduced: `moai init $TMP --non-interactive --project-mode $'personal\ninjected_key: pwned'` → `project.yaml:16` contains `injected_key: pwned` at column 0, outside the `project:` mapping. | **CONFIRMED** |
| **S2** | major | `internal/core/project/initializer_expansion.go` (7 × `os.WriteFile`) | **Non-atomic write of the very files AC-WIZ-010a protects.** The read-patch-write sequence ends in `os.WriteFile`, which truncates before writing. An interrupt mid-write leaves a truncated 11 KB `lsp.yaml` — the same data-loss outcome the wholesale-writer hazard was removed to prevent, arrived at by a different route. The repo already ships `internal/atomicfile.Replace` (`replace.go:20`) for exactly this. Also a TOCTOU window between `os.ReadFile` and `os.WriteFile`. | `grep -c 'os.WriteFile' internal/core/project/initializer_expansion.go` → 7; `ls internal/atomicfile/` + `grep -n 'func Replace' internal/atomicfile/replace.go` → present and unused here. | **CONFIRMED** |
| **F6** | minor | `CHANGELOG.md:28` | **Stale coverage figure.** CHANGELOG states `internal/template` **84.9%**; measured **86.0%** at HEAD. The parenthetical "(all ≥ 85% floor or unchanged baseline)" also glosses that 84.9% is *below* the stated 85% floor. progress.md §E.4 Gaps honestly discloses these were not re-measured in sync; the CHANGELOG carries the number without that provenance qualifier. | `go test -count=1 -cover ./internal/template/...` → `coverage: 86.0% of statements`. | **CONFIRMED** |
| **F7** | minor | `docs-site/content/*/getting-started/init-wizard.md:123` vs `.../getting-started/cli.md:71` and `.../cli-reference/init.md:50` (all 4 locales) | **Two sync-edited docs pages contradict each other on the LSP default** — one says "enabled (Yes)", the others say "(default: false)". Each statement is individually true (wizard default vs flag default), but a reader cannot reconcile them. Downstream symptom of F3. | Both strings present in all 4 locales; see grep output in §4. | **CONFIRMED** |
| **S3** | minor | `initializer_expansion.go:118,200` | Unbounded `os.ReadFile` of the patch target. Bounded in practice by deployed template size; no explicit limit. | Code inspection. | **CONFIRMED** |
| **C1** | minor | `internal/core/project/initializer_expansion.go:91` + `initializer_expansion_test.go:33,57` | Dead `writeHarnessProfileYAML` retained with two live tests, contributing coverage for an unreachable function. Correctly declared as debt (a) with a legitimate §G-amendment rationale — flagged as incomplete-work signal, not as an undeclared defect. | `grep -rn 'TestWriteHarnessProfileYAML'` → 2 tests; no production caller. | **CONFIRMED** |
| **C2** | major | `internal/cli/init.go:368-369` | Code comment "Page-3 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008)" is now factually false for `LSPEnabled` (see F3). | Source read + F3 reproduction. | **CONFIRMED** |
| **N1** | info | `.moai/reports/sync-audit/` | `.gitignore` covers `.moai/reports/*.md` (line 275) and `.moai/reports/plan-audit/*.md` (line 212) but **not** `.moai/reports/sync-audit/`. This report will surface as an untracked file in `git status`. | `git check-ignore -v <this file>` → exit 1 (not ignored). | **CONFIRMED** |
| **N2** | info | `design.md` §D2 | The "real YAML round-trip" recorded as the future fix would destroy comments and key ordering in the 11 KB annotated deployed configs — the exact property this SPEC preserves. Recommend re-characterising it as a tradeoff, not an upgrade. | Design reasoning + observed comment density in deployed `lsp.yaml` / `design.yaml`. | PLAUSIBLE (judgment, not a defect) |

---

## 4. Evidence appendix (commands run in this tree, verbatim results)

```
$ git -C <worktree> rev-parse HEAD
da85ad13e18afc53504db414f1f88cc9151826ba

$ go test -count=1 ./...
exit=0 · 105 ok · 0 FAIL · 3 "no test files"

$ golangci-lint run --timeout=3m
exit=0 · "0 issues."

$ go test -count=1 -cover ./internal/cli/wizard/... ./internal/core/project/... ./internal/template/...
internal/cli/wizard      coverage: 93.9% of statements
internal/core/project    coverage: 88.7% of statements
internal/template        coverage: 86.0% of statements   <-- CHANGELOG says 84.9%

$ test -f internal/cli/wizard/advanced_gate.go            -> GONE
$ grep -rn 'IsAdvancedWizardReady\|AdvancedGate'  internal/  -> 0
$ grep -rn 'advancedMode\|standardMode'           internal/  -> 0
$ grep -rn 'StandardMode\|AdvancedMode'           internal/  -> 0
$ grep -rn 'RunWithDefaultsModes\|Phase2Questions' internal/ -> 0
$ grep -rn '\.Bool("standard"\|\.Bool("advanced"' internal/  -> 0
$ grep -rn -- '--standard|--advanced' .github/               -> 0
$ grep -rnE -- '--standard|--advanced' docs-site/content/    -> 0
$ git grep -nE -- '--standard|--advanced' 216dcbb2e -- docs-site/content/ | wc -l
44                                    (baseline confirmed: 12 files, 44 matches)

$ grep -rin 'DefaultModelPolicy *= *"high"\|default.*"high"' internal/template/context.go internal/config/profile.go
exit=1  (0 matches)
$ grep -cE '기본값[：:] *아니오|デフォルト[：:] *いいえ|默认[：:] *否' internal/cli/wizard/translations.go   -> 0
$ grep -c 'default: No' internal/cli/wizard/questions.go                                                -> 0
$ grep -c 'advanced_bridge\|harness_profile\|coverage_exemptions_enabled' internal/cli/wizard/translations.go -> 0

# Real deployer-path init (binary built from this tree)
$ moai init $T/proj1 --non-interactive --project-mode team --enable-lsp \
      --enforce-quality=false --enable-design=false --harness-profile strict
exit=0
  project.yaml:15   mode: team                      (AC-WIZ-010 row 1)
  lsp.yaml:45       enabled: true                   (AC-WIZ-010 row 2)
  quality.yaml:13   enforce_quality: false          (AC-WIZ-010 row 3)
  design.yaml:8     enabled: false                  (AC-WIZ-010 Scenario B row 6)
  harness.yaml:7    default_profile: "default"      <-- --harness-profile strict IGNORED (F2)
  sentinels: delegate_to_astgrep=1 circuit_breaker=1 figma=1 brand_context=1
  sizes:     lsp 11305 B · design 2868 B · harness 8165 B  (total 22 338 B)
  indents:   design {2sp:1, 4sp:3, 6sp:1} · lsp {2sp:1, 4sp:1}   (AC-WIZ-017)
  project.yaml diff vs control run: exactly 2 lines (name, mode) — no clobber

$ moai init $T/proj3 --non-interactive          # no flags at all
  lsp.yaml   enabled: false                     <-- wizard default is true (F3)

$ moai init $T/proj2 --non-interactive --project-mode $'personal\ninjected_key: pwned'
  project.yaml:16   injected_key: pwned          <-- YAML injection (S1)

$ moai spec audit --json
  total 526 · grandfathered 277 · modern_era_clean 249 · severity {INFO: 381}
  SPEC-CLI-WIZARD-RESTRUCTURE-001 -> V3R5 | H-3 (§E.2 present, sync_commit_sha missing)   <-- F1
  SPEC-DB-RETIRE-001              -> V3R6 | H-4 (§E.2 + §E.4 + sync_commit_sha)
  SPEC-SUBAGENT-NESTING-DOCTRINE-001 -> V3R6 | H-4 (§E.2 + §E.4 + sync_commit_sha)
  SPEC-CLI-TUX-INIT-UPDATE-001    -> V3R6 | H-4 (§E.2 + §E.4 + sync_commit_sha)
  SPEC-CONFIG-AUDIT-REPAIR-001    -> V3R6 | H-4 (§E.2 + §E.4 + sync_commit_sha)

$ moai spec lint --strict | grep -i 'WIZARD-RESTRUCTURE'
(no output — clean for this SPEC)

$ hugo --gc --minify        (from docs-site/)
exit=0 · 0 WARN/ERROR lines · KO 153 / EN 151 / JA 151 / ZH 151 pages

$ per-locale parity (grep -cE '^#{1,3} ' / '^\|')
init-wizard.md        en/ko/ja/zh = 28 headings / 10 rows  (identical)
cli.md                en/ko/ja/zh = 44 headings / 161 rows (identical)
cli-reference/init.md en/ko/ja/zh = 14 headings / 37 rows  (identical)

$ git show --stat adee2f46b   -> 17 files, +286/-369, carries all 4 frontmatter transitions
$ git log --oneline -S'status: completed' -- .../spec.md -> adee2f46b   (sync_commit_sha correct)
$ git log --oneline 216dcbb2e..96d35723c | wc -l -> 9   (merge count claim confirmed)
```

---

## 5. Required fixes (routed defect-list)

Ordered by severity. F1 is the only one that must block.

| ID | Required fix |
|----|--------------|
| **F1** | Rewrite `progress.md` §E.4 `sync_commit_sha` as a **bare line-start YAML key**, matching every sibling SPEC: `sync_commit_sha: adee2f46b` (drop the `- **`+"`"+`…`+"`"+`**` wrapper; move the backfill prose to a following line). Then re-run `moai spec audit --json` and confirm the SPEC reports `era: V3R6` with `H-4 (§E.2 + §E.4 + sync_commit_sha)`. Do **not** paper over it with an `era: V3R6` frontmatter override — that hides the format bug rather than fixing it. Consider applying the same bare-key form to `sync_status` / `sync_complete_at` for consistency. |
| **F2 / F4** | Choose one and make code, docs, and CHANGELOG agree: **(a)** restore a `harness.default_profile` write path (a depth-aware `patchYAMLPathValue(content, "harness.default_profile", …)` call — the primitive already exists and would be non-destructive, unlike the removed wholesale writer); or **(b)** deprecate/remove `--harness-profile`, drop it from the 4-locale non-interactive example + the two flag tables (12 files), and correct the CHANGELOG sentence to state the flag is inert. |
| **F3 / C2** | Change `internal/cli/init.go:371` to `LSPEnabled: getBoolFlagWithDefault(cmd, "enable-lsp", true)` so the non-interactive default matches the wizard default, update the cobra help text at line 89 to `(default: true)`, and update the "(default: false)" rows in `cli.md` / `cli-reference/init.md` across all 4 locales. Then the comment at :368 becomes true again. If the asymmetry is intentional, the comment must say so explicitly and the docs must state both defaults. |
| **S1** | Add `--project-mode` (and `--harness-profile`, if retained) to the `validateInitFlags` enum checks alongside `--mode` / `--git-mode` / `--git-provider`. Belt-and-braces: reject values containing `\n` / `\r` / `:` in `patchYAMLKey` and `patchYAMLPathValue`, or quote the emitted scalar. |
| **S2** | Route the 7 `os.WriteFile` calls in `initializer_expansion.go` through `internal/atomicfile.Replace` (write temp + replace), so an interrupted patch cannot truncate a deployed config. |
| **F5** | Transition `design.md` and `research.md` to `status: completed` to match the other four artifacts and the convention set by `SPEC-AGENT-ARCH-V2-001` / `SPEC-GLM-EFFORT-TUNE-001`; or record in progress.md §E.4 why the Tier-L 5-artifact set closes at 4. |
| **F6** | Correct the CHANGELOG `internal/template` figure to the re-measured **86.0%**, or attach the run-phase provenance qualifier the §E.4 Gaps section already carries. |
| **F7** | Resolved by the F3 fix. |
| **N2** | Re-characterise design.md §D2 option C as a tradeoff (comment/ordering loss) rather than an unqualified future fix. |

---

## 6. Gaps — what I did NOT verify

Stated explicitly so nothing unobserved passes as a pass.

1. **Windows cross-platform build** (`GOOS=windows GOARCH=amd64 go build ./...`) — **NOT RUN**. AC-WIZ-014's cross-platform half is inherited from progress.md, not re-observed by me. My AC-WIZ-014 PASS covers the test suite, coverage, and lint legs only.
2. **`go vet ./...`** — NOT run separately (golangci-lint covers a superset in practice, but I did not observe `go vet` output).
3. **Interactive wizard rendering** — never exercised. Every persistence assertion went through `--non-interactive`. The huh form composition, the Page-1 live locale re-render (AC-WIZ-004), and the `claude_design_enabled` nested hide behaviour (AC-WIZ-003) are verified only by source inspection plus their bound tests running green inside my suite run — I did not drive a TTY.
4. **AC-WIZ-016 bespoke procedure** — I did not run `init_flag_precedence_test.go` in isolation or hand-execute its 4 × 2 × 2 matrix. I verified the implementation uses `cmd.Flags().Changed()` (the property the AC says a value-only implementation would fail) and that the test file exists and passes inside the full-suite run.
5. **`-race`** — not run on any package.
6. **Reconfigure path** (`moai update --reconfigure`) — not executed. AC-WIZ-012 / 012a verified by reading `ReconfigureQuestions` and its splice logic, not by running the command.
7. **Rendered docs-site output** — I verified the hugo build is warning-free and the markdown source is correct, but did not inspect rendered HTML, links, or shortcode expansion.
8. **The 3 non-Latin locales' full documents** — I read the rewritten §"Wizard structure" region and the non-interactive block in ko/ja/zh and confirmed semantic equivalence there. I did **not** semantically diff the remaining ~180 lines of each locale, nor `cli.md` (161 table rows × 4) or `cli-reference/init.md` beyond the flag rows.
9. **The `base_merge` divergence figures** (`0 21` / `9 20`) — the `9` is corroborated by my own count of 9 merged commits, and the left-right reading is now internally consistent; I did not re-run `git rev-list --count --left-right` against the historical states.
10. **Block-scalar edge case** in `patchYAMLPathValue` (declared debt (b)) — I confirmed by inspection that a `|`/`>` block whose body resembles `key:` would be mis-walked, and that no shipped config currently triggers it. I did not construct a failing fixture.
11. **S1 exploitability beyond the local CLI** — I reproduced the injection; I did not analyse whether any CI/automation surface feeds untrusted input into `--project-mode`.
12. **Historical CHANGELOG entry** at line 38 mentions `--standard`/`--advanced` — I judged this correct (it describes the v3.0.1 release as it shipped) and did not treat it as a stale reference.

## 7. Residual risk (what could still be wrong despite what I DID observe)

- My persistence verification used `--non-interactive`. If `applyWizardPage3ToOpts` mis-maps a wizard answer that the flag path happens to set correctly, my run would not catch it. The bound tests cover this, but I did not independently drive the wizard.
- Coverage figures are single-run; `internal/cli` at 75.2% is well below the 90% "critical package" target in CLAUDE.local.md §6, though that is pre-existing and outside this SPEC's scope.
- The repo carries two known-flaky CI tests (preference `-race`, session `TestRegisterSessionConcurrent`); my suite run was green, but a green single run does not establish stability.
- F1's downstream effect (permanent drift-detection exemption) is silent by construction — nothing will surface it again until someone re-runs `moai spec audit` and reads the era column.
