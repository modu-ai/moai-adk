# Sync-Phase Independent RE-AUDIT (ROUND 3) — SPEC-CLI-WIZARD-RESTRUCTURE-001

- **Auditor**: sync-auditor (independent, adversarial stance)
- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/wizard`
- **Branch / HEAD**: `feat/SPEC-CLI-WIZARD-RESTRUCTURE-001` @ `df31c3392`
- **Prior audits**: `da85ad13e` → FAIL 0.745 · `209c8f9f9` → FAIL 0.834
- **Tier**: L — PASS threshold **0.85** (harmonic mean)
- **Scoring model**: flat weighted 4-dimension (`harness.yaml` sets no `evaluator_mode: hierarchical`)
- **Remediation under review**: `580d49959` (N3) + `a827880e7` (C2/F3/F7) + `df31c3392` (docs sweep + audit record)

All four dimension scores were **re-derived from the current tree**. No round-2 score was carried forward.
Working tree confirmed clean (`git status --short` → empty); this audit wrote only its own report file.

---

## 1. Verdict

| Dimension | Round 2 | Round 3 | Δ | Verdict |
|-----------|--------:|--------:|---:|---------|
| Functionality (40%) | 0.88 | **0.92** | +0.04 | PASS |
| Security (25%) | 0.82 | **0.82** | 0.00 | PASS (no Critical/High — hard threshold not tripped) |
| Craft (20%) | 0.84 | **0.84** | 0.00 | PASS |
| Consistency (15%) | 0.80 | **0.86** | +0.06 | PASS |

**Harmonic mean** = 4 / (1/0.92 + 1/0.82 + 1/0.84 + 1/0.86)
= 4 / (1.086957 + 1.219512 + 1.190476 + 1.162791) = 4 / 4.659736 = **0.858**

## OVERALL VERDICT: **PASS** (0.858 ≥ 0.85)

This is the first passing round, and it passes on genuine repair, not on score drift. **Nine of the nine
round-2 findings routed for fix are closed, and I reproduced every one of them against the built binary
or by direct re-measurement rather than accepting the commit message.** The two behavioural defects that
held Functionality down are now one; the four-way contradiction between validator, template, help text
and docs is gone; and the audit record's evidence path resolves.

**It passes despite a fresh defect, not because I missed one.** The remediation prose introduced
**N9 — a false claim that a correction was made when it was not** — in the exact clause the round-2 audit
flagged (N5). That is the third consecutive round in which this SPEC's own reporting contains a factual
error, and this instance is the worst-shaped of the three: the record asserts a verification it did not
perform. It is a **one-word fix** and it MUST be made before merge (see §5). It does not trip a must-pass
criterion, so it does not force a FAIL, but nothing about this verdict should be read as tolerating it.

**Robustness of the verdict.** The pass is not knife-edge. Holding Security at 0.82, the mean stays above
0.85 for Craft down to 0.82 (0.853) and Consistency down to 0.84 (0.853). It flips to FAIL only if
Functionality is held at its round-2 value of 0.88 (0.849) — i.e. only if closing F3 and N3 is judged to
have bought nothing. Three independent binary reproductions (§3.1) say otherwise.

**Null hypothesis considered** — *"did this round fix anything, or did it just write prose about the
findings?"* Two of the three commits are one-line/one-comment code and template edits whose effects I
reproduced with the real binary; the third is a 12-file docs sweep whose effect I verified by grep across
all four locales. The prose-only portion is the audit record, and that is precisely where the new defect
landed. The pattern across three rounds is consistent and worth naming: **this SPEC's code fixes hold up
under adversarial re-verification; its self-reporting does not.**

---

## 2. Treatment of the deferred debt (S2) — and why Security did not move

The rule I stated in round 2 and am bound to apply consistently:

> An accurately declared deferral converts an *undisclosed defect* into a *known, bounded risk*. It
> substantially reduces the Consistency penalty and partially reduces the Functionality penalty. **It
> reduces the Security penalty by approximately zero.**

**Security scored 0.82 in round 2 and scores 0.82 now, and the reason is mechanical: the shipped security
surface is byte-identical between `209c8f9f9` and `df31c3392`.** `a827880e7` touched the LSP seed (not a
security path), `580d49959` touched a template comment, and `df31c3392` touched only markdown. Nothing
that a crash, an interrupt, or a malicious input can reach changed.

I record explicitly that I considered awarding +0.01 for the improved S2/S3/S1-residual declarations and
**declined**. Nudging a dimension by one point to clear a threshold is the sycophancy gradient the
skeptical stance exists to resist; the declarations improved Consistency and Craft, which is where I
credited them. Had I taken that +0.01, the mean would have been 0.860 instead of 0.858 — the verdict is
unchanged either way, which is how I know the restraint cost nothing and confirms the pass is not an
artifact of score-tuning.

### Is the S2 declaration accurate and complete?

**Accurate on the defect and on the correction — with one arithmetic error, and one omission in its sibling row.**

| Aspect | Assessment |
|--------|-----------|
| "Seven config writers use non-atomic `os.WriteFile`" | **TRUE** — `grep -c 'os.WriteFile' internal/core/project/initializer_expansion.go` → **7**. Unchanged. |
| The round-2 false clause "`atomicfile.Replace` ships unused" | **REMOVED.** `grep -c "ships unused" CHANGELOG.md` → **0**. The single surviving occurrence in `progress.md:604` is a *quotation of the old claim inside the N4 finding row* — legitimate, not a residual falsehood. |
| Replacement claim: "10 production call sites across 7 files" | **Total TRUE, independently re-measured.** I did not trust the number: `grep -rn "atomicfile.Replace" --include="*.go" .` minus `_test.go` → exactly **10**, across exactly **7** files. Note this also corrects *my own* round-2 count of 5 — I had missed `internal/session/store.go` (4) and `internal/session/registry.go` (1). The remediation's number is right and mine was wrong. |
| Per-file breakdown | **WRONG (N10).** Both surfaces write `internal/session/store.go` **(×3)**; the file has **4** (`:129, :415, :452, :508`). The enumerated breakdown sums to **9** against a stated total of **10**. |
| Scope attribution: "it is only the *config writers* in `initializer_expansion.go` that do not route through it" | **TRUE** — the 10 sites are all outside that file; the 7 `os.WriteFile` calls are all inside it. |
| TOCTOU window | Still omitted from the §G row (CHANGELOG (g) also omits it now — round 2 noted the CHANGELOG was the better of the two on this point; that advantage was lost in the rewrite). Minor. |
| Deferral routing | **FIXED (N8).** "deferred and currently unscheduled — no follow-up SPEC exists yet"; `grep -c "deliberately deferred to a follow-up SPEC" CHANGELOG.md` → **0**. |

**How much does S2 still depress Security, concretely?** It is the single largest deduction in the
dimension. Security sits at 0.82 rather than ~0.93 almost entirely because of it: seven non-atomic
read-patch-write sequences against the same ~22 KB deployed config set that AC-WIZ-010a exists to
protect, where a crash between truncate and write leaves a truncated `lsp.yaml`. The repo ships
`atomicfile.Replace` and 10 other sites use it. The declaration is now honest, and honesty is worth
something — but a declared non-atomic write truncates a config file exactly as often as an undeclared
one. **Roughly 0.11 of the 0.18 Security deficit is S2**; the remainder splits between S3, the
S1-residual writer-level gap, and the general "validation lives in callers, not in the writer" posture.

I do not re-litigate the deferral. The user's decision stands and is correctly recorded.

---

## 3. Per-finding re-verification

Legend — **CONFIRMED** = reproduced in this tree, this pass. **PLAUSIBLE** = judgment, not reproduced.

### 3.1 RESOLVED — round-2 findings, each independently re-verified

| ID | Sev | Verification performed this pass | Status |
|----|-----|----------------------------------|--------|
| **N3** | minor | **Template↔validator↔help↔wizard↔docs now agree, and the behaviour is now *correct* rather than a regression.** Template: `project.yaml.tmpl:11` → `# Project mode: personal or team` (the `enterprise` bullet at former :14 is gone; repo-wide `grep enterprise internal/template/templates/` returns only unrelated Claude-Code-settings prose). Validator: `init.go:233` `{"personal", "team"}`. Flag help: `init.go:87` `"Project mode: personal or team (default: personal)"`. Wizard: `questions.go:444-445` select with exactly `personal`/`team`. Docs: 8 rows already `<personal\|team>`. **Binary**: `moai init . --non-interactive --project-mode enterprise` → `Invalid --project-mode value "enterprise": must be one of: personal, team`, **exit 1**, `project.yaml` **not created**; `--project-mode team` → **exit 0**, `project.yaml:14 mode: team`. **Was `enterprise` really accepted before?** `git show da85ad13e:internal/cli/init.go \| grep project-mode` → three hits (registration, `Changed()` gate, `getStringFlag` seed) and **no enum check** — so yes, it was accepted pre-`0e34ac7a3`. The round-2 framing (a user following the deployed template's own comments hits a hard failure on a previously-working value) is now void: the template no longer documents it. | **RESOLVED — CONFIRMED** |
| **F3** | major | **Reproduced on three variants with the binary built from this tree.** (1) `moai init . --non-interactive` (no flags) → `lsp.yaml:45 enabled: true`. (2) `--enable-lsp=false` → `enabled: false` — the explicit false still wins. (3) `--enable-lsp=true` → `enabled: true`. Mechanism: `init.go:391 LSPEnabled: getBoolFlagWithDefault(cmd, "enable-lsp", true)`, and `getBoolFlagWithDefault` (`:171-180`) keys off `Changed()`, so an explicitly-typed `false` is not swallowed by the seed default. **Pinned by a real regression guard**: `init_flag_precedence_test.go:317` `TestSeedMirrorsProductionLSPSeed` reads `init.go` and asserts the literal `getBoolFlagWithDefault(cmd, "enable-lsp", true)` is present — reverting the fix fails the test. | **RESOLVED — CONFIRMED** |
| **C2** | major | **The comment is now true for every field it covers — I checked all five, not just LSP.** `init.go:386` reads *"Page-3 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008)"*. `LSPEnabled` → seed `true` == wizard `lsp_enabled` default `"true"` (`questions.go:423`) ✔. `EnforceQuality` → `true` ✔. `DesignEnabled` → `true` ✔. `ClaudeDesignEnabled` → hardcoded `true`, annotated wizard-only ✔. `ProjectMode` → seed `""`, but `writeProjectModeYAML` (`initializer_expansion.go:58`) coerces empty to `"personal"`, which is the wizard's recommended default — **verified end-to-end**: T1 `project.yaml:14 mode: personal` ✔. `CoverageExemptionsEnabled: false` carries its own "no CLI flag; wizard/default only" annotation ✔. | **RESOLVED — CONFIRMED** |
| **F7** | minor | **All 4 locales × both pages now say `true`.** `cli.md:71` → en `(default: true)` / ko `(기본값: true)` / ja `(デフォルト値: true)` / zh `(默认: true)`; `cli-reference/init.md:50` → same in all 4. `init-wizard.md:123` keeps "enabled (Yes)" — now *agreeing* rather than contradicting. Residual-`false` sweep across the 12 files (`default: false` / `기본값: false` / `默认.false`) → **0 matches**. | **RESOLVED — CONFIRMED** |
| **F2-docs** | major | **The copy-paste CI example no longer teaches an inert flag.** `grep -rn "harness-profile"` across all 12 files: **zero hits in any `init-wizard.md`** (the `--harness-profile default \` line is gone from all 4 locales — `df31c3392` shows `-1` line in each), and exactly **8 flag-table rows** (`cli.md:70` + `cli-reference/init.md:49`, × 4 locales) each now carrying the inertness disclosure in its own language (en "accepted but currently has no persisted effect"; ko "값은 받지만 현재 저장되어 반영되지는 않음"; ja "値は受け付けるが現在は永続化されない"; zh "接受该值，但目前不会持久化生效"). **The CHANGELOG agrees** — same claim, §(4) and §(7). The round-2 "fresh internal contradiction" (CHANGELOG says inert, docs teach it) is gone. **Binary re-confirms inertness**: T1 `harness.yaml:7 default_profile: "default"`; `writeHarnessProfileYAML` still has **zero production callers** (only 2 test call sites). | **RESOLVED — CONFIRMED** |
| **N4** | minor | Corrected total is right and I re-measured it myself rather than trusting it — **10** production `atomicfile.Replace` call sites across **7** files. See §2 for the full table. (Breakdown arithmetic is wrong → N10.) | **RESOLVED — CONFIRMED** |
| **N6** | minor | **Both cited paths resolve to committed files.** `git ls-files .moai/reports/sync-audit/ \| grep WIZARD` → both `…-2026-07-25.md` and `…-2026-07-25-reaudit.md` are **tracked**. `progress.md:557` cites the first, `:573` cites the second — both resolve. The round-2 self-concealing hazard (a merged §G citing an unopenable file) is closed. | **RESOLVED — CONFIRMED** |
| **N7** | minor | **§G is now a materially complete resolution record.** I cross-checked it against the round-1 report's full 14-finding inventory (F1-F7, S1-S3, C1, C2, N1, N2). Round-1 fixed table: F1, S1, F5, F6, F2-CHANGELOG/F4. Round-2 fixed table: F3, C2, F7, F2-docs, N3, N4, N5, N6, N7, N8. Deferred table: S2, S3, S1-residual. **Every actionable finding now appears in exactly one table.** Two info-level items remain outside: C1 (declared as CHANGELOG debt (a) — acceptable, unchanged from round 2) and N2 (round 1 classified it PLAUSIBLE-judgment, not a defect). | **RESOLVED — CONFIRMED** |
| **N8** | minor | **No phantom SPEC remains.** `grep -c "deliberately deferred to a follow-up SPEC" CHANGELOG.md` → **0**. Both surfaces now read "deferred and currently unscheduled — no follow-up SPEC exists yet"; §G closes with "recorded in the CHANGELOG entry as debt (g), (h), (i)", and those three letters are in fact defined there. | **RESOLVED — CONFIRMED** |

### 3.2 STILL-OPEN — declared and unchanged

| ID | Sev | Location | Status this pass |
|----|-----|----------|------------------|
| **S2** | major | `initializer_expansion.go` (7 × `os.WriteFile`) | Unchanged. Declaration accurate (see §2), with the N10 breakdown error. **User-deferred — not re-litigated.** |
| **S3** | minor | `initializer_expansion.go:118,200` | Unchanged; now declared (closes the N7 half). Declaration under-counts its own scope → **N12**. |
| **S1-residual** | minor | `writeProjectModeYAML` / `patchYAMLKey` | Unchanged; now declared as debt (i). Verified still true: the writer interpolates raw, and `--harness-profile` remains **unvalidated** (`init.go:88,390` — registration and `getStringFlag` seed only, no enum check), safe solely because its writer is dead. The declaration states this coupling correctly. |
| **C1** | minor | dead `writeHarnessProfileYAML` + 2 tests | Unchanged; correctly declared as debt (a). Accepted as declared debt. |

### 3.3 NEW — introduced by the `df31c3392` remediation prose

The round-2 report predicted further prose-vs-tree drift in the ~180 lines of §G/CHANGELOG text it did not
line-by-line verify. That prediction held. I read the new prose line by line this pass.

| ID | Sev | Location | Evidence | Status |
|----|-----|----------|----------|--------|
| **N9** | **major (doc)** | `CHANGELOG.md:28` (original bullet) vs `CHANGELOG.md` §(9); `progress.md:606` §G N4/N5 row | **The N5 correction was claimed but never applied — the record asserts a verification it did not perform.** CHANGELOG §(9) states *"the coverage sentence now reads 'unchanged from the package's own baseline'"*, and progress.md §G's N5 row states *"CHANGELOG restated as 'unchanged from'"*. **Neither is true.** At HEAD, `grep -c "above the package's own unchanged baseline" CHANGELOG.md` → **1** — the original clause is verbatim intact. The phrase "unchanged from the package's own baseline" occurs exactly once, **inside the §(9) paragraph describing the correction**, never in the sentence it was supposed to correct. The entry therefore contradicts itself within one bullet. **And the surviving clause is independently false**: I re-measured all four points with `git archive` + `go test -cover ./internal/template/` rather than trusting round 2 — base `b1ea545e2` → **84.9%**, `origin/main` → **86.0%**, HEAD → **86.0%**. Both figures are **exactly equal** to their respective baselines; neither is above. This is the same defect class the round-2 audit raised (N5), now compounded: a live artifact-vs-tree falsehood *plus* a false claim that it was fixed. Under `verification-claim-integrity.md` §1.1 surface 3, an asserted correction that was not made is an unobserved-verification claim. | **CONFIRMED** |
| **N10** | minor | `CHANGELOG.md` §(9); `progress.md:611` §G S2 row | **Per-file breakdown does not sum to its own stated total.** Both surfaces enumerate the `atomicfile.Replace` sites as 7 files with `internal/session/store.go` **(×3)**. Actual: `grep -c "atomicfile.Replace" internal/session/store.go` → **4** (`:129, :415, :452, :508`). The enumeration sums to **9** while the sentence claims **10**. The total is correct; the breakdown is off by one. Same error mirrored in both surfaces — a copy of one wrong count, not two independent slips. | **CONFIRMED** |
| **N11** | minor | `CHANGELOG.md` §(6), §(7) | **Dangling debt-letter references.** §(6) says *"This closes finding (f) above — it is no longer deferred"*, but the `(f)` definition was deleted from the original bullet **in this same commit**. `grep -c "(f)" CHANGELOG.md` → **1**, and that single occurrence *is* the dangling reference — no `(f)` is defined anywhere in the file. A reader is pointed at a label that does not exist. Separately, §(7) writes *"(debt (a) below)"* while `(a)` is defined in the paragraph **above** it — wrong direction. Debt letters actually defined in the entry: (a)(b)(c)(d)(e) in the first paragraph, (g)(h)(i) in the Remaining-debt paragraph; (f) is referenced but undefined. | **CONFIRMED** |
| **N12** | minor | `progress.md:612` §G S3 row; `CHANGELOG.md` debt (h) | **The S3 declaration under-counts its own scope.** Both cite `initializer_expansion.go:118,200` as the unbounded-`os.ReadFile` sites. The file has **four** `os.ReadFile` calls (`:63, :118, :149, :200`), of which **`:63` is the identical unbounded pattern** — `writeProjectModeYAML` reading `projectYAMLPath` with the same `//nolint:govet` annotation and no size limit. `:149` is the guarded `if data, err := …` form. So the declared debt names 2 of the 3 unbounded patch-target reads. Notably `:63` is in `writeProjectModeYAML` — the very function C32 moved onto the reachable CLI path, i.e. the site this SPEC newly exposed. Inherited from my own round-1/round-2 finding text, so this is an accuracy gap in the declaration rather than a fresh remediation error — but it is live in the merged record. | **CONFIRMED** |

### 3.4 Explicitly assessed and judged NOT a defect

**The cobra `DefValue: false` vs effective default `true` asymmetry — ACCEPTABLE, DOCUMENTED, not a new defect.**

The question was put directly, so here is the reasoning rather than a verdict:

1. **No user-visible contradiction in `--help`.** `moai init --help` renders `--enable-lsp   Enable LSP integration (default: true)` with **no cobra-appended default annotation**, because cobra suppresses zero-value defaults. Its siblings render `--enable-design  Enable design workflow (default: true) (true)` — cobra appends `(true)` for them. The asymmetry makes `--enable-lsp`'s help line *cleaner*, not contradictory. There is no surface where a user sees "false".
2. **No production path reads `DefValue`.** Every read is `Changed()`-gated: `init.go:391` (`getBoolFlagWithDefault`) and `init.go:146` (`applyWizardPage3ToOpts`). `grep -rn "DefValue" --include="*.go" internal/ cmd/` returns **5 test sites** (`web_test.go`, `mx_query_test.go`, `update_test.go`) — **none for `enable-lsp`**. The registered default is inert.
3. **The choice is documented at the declaration site**, `init.go:89-92`: *"Registered false but read with a true default (the LSPEnabled seed in runInit), so the effective default matches the wizard's lsp_enabled default; getBoolFlagWithDefault keys off Changed(), so --enable-lsp=false still wins."* That is exactly the "if the asymmetry is intentional, the comment must say so explicitly" remedy the round-1 report demanded.
4. **Why not just register `true`?** Because `Bool("enable-lsp", true)` would make cobra's `NoOptDefVal` and the `Changed()` probe interact differently with the wizard-precedence rule at `:146`, which must distinguish "absent" from "explicitly set to the same value as the default". The current shape is the smaller change and preserves that distinction.

**Residual, stated honestly**: this is the same structural shape as S1-residual — *the invariant lives in the callers, not in the declaration*. A future contributor adding a bare `getBoolFlag(cmd, "enable-lsp")` silently gets `false`. That risk is mitigated but not eliminated by the comment and by `TestSeedMirrorsProductionLSPSeed`, which pins the seed literal. I record it as a low-severity note, not a finding.

### 3.5 Regression sweep

Every item below was run in this tree, this pass.

- **Go suite** — `go test -count=1 ./...` → **exit 0**, **105 ok**, **0 FAIL**. No regression.
- **Lint** — `golangci-lint run --timeout=4m` → **`0 issues.`**, exit 0. No regression.
- **Coverage, all three figures re-measured (not trusted)** — `internal/template` **86.0%**, `internal/cli/wizard` **93.9%**, `internal/core/project` **88.7%**. All three match the CHANGELOG exactly.
- **Era classification** — `moai spec audit --json` (binary built from this tree) reports verbatim: `{"spec_id": "SPEC-CLI-WIZARD-RESTRUCTURE-001", "era": "V3R6", "finding_type": "EraAutoDetected", "severity": "INFO", "details": {"heuristic_matched": "H-4 (§E.2 + §E.4 + sync_commit_sha)"}}`. **Correct SHA**: `progress.md:403 sync_commit_sha: adee2f46b`, bare line-start key, with `sync_status: complete` (:401) and `sync_complete_at: 2026-07-25` (:402). The ~25 new §G lines did **not** pollute the parser — the later token occurrences (:413, :418, :589) are prose and match neither extractor. No regression.
- **Artifact statuses** — all six (`spec/plan/acceptance/design/research/progress`) read `status: completed`. Uniform. No regression.
- **Retirement grep** — `--standard` / `--advanced` across `docs-site/content/`, `internal/cli/`, `internal/core/project/` (non-test) → **0**. Still fully retired.
- **4-locale parity** — heading/table-row counts identical across en/ko/ja/zh for all three page families: `init-wizard.md` 28/10, `cli.md` 44/161, `cli-reference/init.md` 14/37. Line counts identical except `ja/cli.md` at 494 vs 497 — **pre-existing**, verified unchanged at `209c8f9f9` and `da85ad13e` (both 494), so not introduced here. No regression.
- **Hugo** — `hugo --gc --minify` → **exit 0**, **0 WARN/ERROR**, 153/151/151/151 pages. Identical to round 2. No regression.
- **S1 injection, re-reproduced** — `moai init . --non-interactive --project-mode $'personal\ninjected_key: pwned'` → **exit 1**, `project.yaml` **never created**. Still closed.
- **S2 surface, re-counted** — `grep -c 'os.WriteFile' initializer_expansion.go` → **7**. Unchanged, as declared.
- **Commit-message claims** — `df31c3392`'s message asserts "go test ./... exits 0 (105 packages); hugo build warning-free with 4-locale heading/table-row parity preserved" and "No Go source touched". All four reproduce; `git show --stat` confirms zero `.go` files in the commit. **The commit messages are accurate; the defect is in the artifact prose they describe** — the same split as round 2.

---

## 4. Dimension reasoning — what moved and why

### Functionality 0.88 → 0.92 (+0.04)

**Up.** Round 2 held this dimension down for three shipped behavioural defects; two are gone. **F3 was the
larger of them** — `moai init --non-interactive` in a CI pipeline silently produced `lsp.enabled: false`,
contradicting the wizard, its own help text, and `init-wizard.md`. It now produces `true`, reproduced on
three flag variants, and the correct-override behaviour (`--enable-lsp=false` still wins) is reproduced
too, so the fix did not trade one defect for another. **N3** is gone as a defect class entirely: the
template no longer documents a value the validator rejects. 19/19 AC re-confirmed green under my own suite
+ lint run.

**Held back.** One behavioural wart survives in shipped code: **F2** — `--harness-profile` is accepted,
parsed, copied into `opts`, and does nothing. A CI script passing it gets no error and no effect. It is
now honestly documented in code, in 8 doc rows across 4 locales, and in the CHANGELOG, so no user is
*misled* — but disclosure is not repair, and Functionality scores the artifact. **S2**'s truncation path
is also a functional-risk carry. 0.92, not higher, for those two.

### Security 0.82 → 0.82 (0.00)

**Nothing moved, and nothing should have.** The shipped security surface is byte-identical to round 2 — I
verified the three intervening commits touch the LSP seed, a template comment, and markdown only. S1
remains cleanly closed (re-reproduced: exit 1, no file written). S2 remains unmitigated (7 non-atomic
writes, re-counted). S3 and S1-residual are unchanged in the code; they moved from *undeclared* to
*declared*, which I credited to Consistency and Craft, not here — per the rule stated in round 2 §2 and
restated in §2 above. N12 further blunts even the declaration gain, since the S3 row under-counts its own
site list.

No Critical or High finding exists — S1 is closed, S2 is a durability/robustness defect rather than an
exploitable vulnerability, and S3 / S1-residual are defense-in-depth gaps with no live path. **The hard
threshold does not trip**, so Security does not force an overall FAIL.

### Craft 0.84 → 0.84 (0.00)

**Up, on the code.** The C2/F3/F7 fix is the best-shaped work in this SPEC: **one line closes three
findings**, it is pinned by a real regression guard that reads the source and asserts the seed literal, and
the deliberate cobra-`DefValue` asymmetry is explained at the declaration site rather than left for the
next reader to rediscover. The 12-file 4-locale sweep preserved structural parity exactly and builds
warning-free. §G is now a genuinely complete resolution record, cross-checked against the round-1
14-finding inventory.

**Down, on the record — and Craft includes the craft of the audit record.** This repair pass added roughly
four paragraphs of new prose and introduced **four defects in them** (N9, N10, N11, N12) — a *higher*
defect density than the round-2 prose it was repairing. N9 is the serious one: it is not a stale figure or
a typo, it is **a claim that a correction was made when it was not**, sitting in the document whose purpose
is to be the trustworthy record. N10's breakdown does not sum to its own total. N11 points at a label
deleted in the same commit.

The code craft rose and the record craft fell. **Flat is the honest resolution**, not a compromise.

### Consistency 0.80 → 0.86 (+0.06)

**Up — this is where the remediation earned its keep.** Six of the round-2 Consistency defects are closed
and I verified each: the inert flag is gone from the copy-paste CI example in all 4 locales; the 8 flag
rows disclose inertness in 4 languages; the LSP default reads `true` on both pages in all 4 locales,
agreeing with `init-wizard.md`, with the binary, and with the CHANGELOG; the template enum matches the
validator, the help text, the wizard options, and the docs; the evidence path resolves to tracked files;
the phantom follow-up SPEC is gone. Era/H-4/SHA, 4-locale parity, retirement greps, and the warning-free
build are all intact.

**Held back by a self-contradiction inside the primary consistency artifact.** Consistency asks whether the
artifacts describe the tree. The CHANGELOG's surviving *"both are above the package's own unchanged
baseline"* **does not describe the tree** — I measured it: 86.0% vs an 86.0% baseline, and 84.9% vs an 84.9%
baseline. Equal, not above. Worse, the same entry claims that clause was changed. One document, two
mutually exclusive statements (N9), plus a pointer to a label it deleted (N11).

That is a smaller blast radius than round 2's cluster — a coverage percentage in a changelog versus a CI
command users copy and paste — which is why the dimension rose 6 points. It is not zero, which is why it
stopped at 0.86 rather than ~0.90.

---

## 5. Routed defect-list

| ID | Sev | Location | Required fix | Worth making? |
|----|-----|----------|--------------|---------------|
| **N9** | **major (doc)** | `CHANGELOG.md:28` | Replace **`above the package's own unchanged baseline`** with **`unchanged from the package's own baseline`** in the original bullet's coverage parenthetical. **One word.** This simultaneously removes the live falsehood and makes §(9)'s claim true, resolving the self-contradiction. Verify with `grep -c "above the package's own unchanged baseline" CHANGELOG.md` → 0. | **YES — make it before merge.** Lowest-cost, highest-integrity fix available. Merging as-is ships a record that falsely claims a correction, on a SPEC already carrying three rounds of reporting errors. |
| **N10** | minor | `CHANGELOG.md` §(9); `progress.md:611` | Change `internal/session/store.go` **(×3)** → **(×4)** in both surfaces, so the breakdown sums to the stated total of 10. | YES — same edit pass, two characters. |
| **N11** | minor | `CHANGELOG.md` §(6), §(7) | §(6): replace *"This closes finding (f) above"* with a self-contained phrasing (e.g. *"This closes the deferred `--enable-lsp` default finding"*), since `(f)` no longer exists. §(7): change *"(debt (a) below)"* → *"(debt (a) above)"*. | YES — same edit pass. |
| **N12** | minor | `progress.md:612`; `CHANGELOG.md` debt (h) | Change the S3 citation from `118,200` to `63,118,200`, or state "three unbounded reads". Line 63 is in `writeProjectModeYAML` — the function C32 newly exposed — so omitting it understates the SPEC-introduced portion of the debt. | YES — same edit pass. |
| **S2 / S3 / S1-residual** | major / minor / minor | `initializer_expansion.go` | **User-deferred. No action requested.** Recorded here only because they are the arithmetic reason Security sits at 0.82. | Out of scope by user decision. |

### Is the residual gap structural?

**Partly, and the two parts should not be confused.**

- **The Security deficit (0.82 vs 1.00) is structural** under the current decision. It is driven almost
  entirely by the accepted S2 deferral, and it is **not closable without reversing a user decision**. No
  amount of documentation moves it: routing the 7 `os.WriteFile` calls through `atomicfile.Replace` is the
  only thing that would, and that is precisely the cross-cutting refactor the user declined. This is a
  known, bounded, correctly-declared risk — the right shape for deferred debt.
- **The four open findings (N9-N12) are NOT structural.** They are four small text edits in two files,
  touching no code, no tests, and no docs-site page. They cost one commit.

Because the tree already clears the bar at 0.858, the N9-N12 fixes are **not required to reach PASS** — but
N9 in particular is required for the record to be true, which is a different and more important standard
than the score. Making all four would lift Consistency to roughly 0.90 and Craft to roughly 0.87
(mean ≈ 0.875), and would end the three-round pattern of this SPEC's reporting carrying factual errors.

---

## 6. Gaps — what I did NOT verify this pass

Stated explicitly so nothing unobserved passes as a pass.

1. **The 19 ACs were not individually re-derived.** I re-ran the full suite (105 ok / 0 FAIL), lint (0 issues), all three coverage figures, and both baseline coverage points, and I confirmed the only production delta since `209c8f9f9` is the one-line `init.go` LSP seed. The per-AC greps and byte-size / indentation-multiset checks from round 1 are **carried forward with attribution to `da85ad13e`**, not re-observed here.
2. **`GOOS=windows GOARCH=amd64 go build ./...`** — NOT RUN, in any of the three passes.
3. **`go vet ./...`** — not run separately (golangci-lint's govet pass covers part of it, but I did not isolate it).
4. **`-race`** — not run on any package, in any pass.
5. **Interactive wizard through a TTY** — never driven. The claim that the `project_mode` select cannot emit an out-of-enum value rests on source inspection of `questions.go:443-446`, not an interactive run.
6. **Reconfigure path** (`moai update --reconfigure`) — not executed. My producer-set enumeration for `ProjectMode` is a whole-module grep; it would catch a direct assignment but not a reflective or config-driven one.
7. **Non-vacuity of the S1 and LSP tests was established structurally, not by mutation.** I did not remove the validator or revert the seed and observe failures — the tree is read-only for this audit. For the LSP seed the argument is strong (`TestSeedMirrorsProductionLSPSeed` greps for the literal, so reverting it fails deterministically); for S1 it is deductive.
8. **Rendered docs-site HTML** — hugo exits clean, but I inspected markdown source only, not rendered output, links, or shortcode expansion.
9. **Full semantic re-diff of the 3 non-Latin locales.** Parity was measured by heading/table-row/line counts, and the specific `--enable-lsp` / `--harness-profile` lines were read in all 4 locales; I did not semantically re-diff the remaining body text of ko/ja/zh.
10. **Block-scalar edge case** in `patchYAMLPathValue` (declared debt (b)) — no failing fixture constructed, in any pass.
11. **The pre-`0e34ac7a3` `enterprise` acceptance was confirmed at source level, not by running the old binary.** I inspected `git show da85ad13e:internal/cli/init.go` and found no enum check; I did not build and run that revision.
12. **The `base_merge` divergence figures** in §E.4 — not re-run against historical states.
13. **N2 from round 1** (design.md §D2 "real YAML round-trip" characterization) — classified PLAUSIBLE-judgment in round 1 and never routed; I did not re-examine whether the design text was ever re-characterized.
14. **CHANGELOG paragraphs outside this SPEC's bullet** — I read the wizard entry line by line but did not audit the neighbouring `SPEC-SUBAGENT-NESTING-DOCTRINE-001` entry or the rest of the file.

## 7. Residual risk — what could still be wrong despite what I DID observe

- **My coverage baselines came from `git archive` snapshots built in a scratch directory sharing this machine's module cache.** A dependency-resolution difference between those snapshots and a clean CI checkout would change the figures — and the N9 finding rests on the 86.0%-equals-86.0% measurement. I consider this low risk (four measurements, two matched pairs, consistent with round 2's independent run) but it is not a CI-clean measurement.
- **N9's consequence is self-concealing in the same way N6's was.** Once merged, the CHANGELOG carries a false coverage claim next to a paragraph asserting that claim was corrected. Nothing surfaces it until someone compares the two — which is exactly what took three audit rounds to happen here.
- **Three rounds, three sets of reporting errors.** Round 1 found 2 factual errors in the sync report; round 2 found 2 more in the round-1 remediation (N4, N5); round 3 found 4 in the round-2 remediation (N9-N12). The rate is not declining. I read the new prose line by line this pass, but the CHANGELOG entry for this SPEC is now ~5 paragraphs of dense text and I cannot claim exhaustive verification of every embedded claim. **Further unfound instances are plausible.**
- **The F3 fix's correctness is a statement about the current `Changed()`-gated read pattern.** It holds today and silently breaks the moment someone adds a bare `getBoolFlag(cmd, "enable-lsp")`. `TestSeedMirrorsProductionLSPSeed` pins the seed line but does not forbid a second, ungated reader.
- **Two known-flaky CI tests exist in this repo** (preference `-race`, session `TestRegisterSessionConcurrent`). My suite run was green; one green run does not establish stability, and I did not run with `-race`. Note that `internal/session/store.go` — the file whose `atomicfile.Replace` count the record miscounts — is in that neighbourhood.
- **The verdict's sensitivity is concentrated in one dimension.** Holding everything else, the mean falls below 0.85 only if Functionality is scored at its round-2 value. I judge the F3/N3 closures to be worth +0.04 on the strength of three binary reproductions; a reviewer who valued them at +0.01 would reach FAIL 0.849. I record this so the disagreement, if any, is locatable at a single number rather than diffused across the report.
