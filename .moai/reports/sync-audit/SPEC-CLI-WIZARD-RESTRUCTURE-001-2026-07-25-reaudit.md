# Sync-Phase Independent RE-AUDIT — SPEC-CLI-WIZARD-RESTRUCTURE-001

- **Auditor**: sync-auditor (independent, adversarial stance)
- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/wizard`
- **Branch / HEAD**: `feat/SPEC-CLI-WIZARD-RESTRUCTURE-001` @ `209c8f9f9`
- **Prior audit**: `da85ad13e` → FAIL 0.745
- **Tier**: L — PASS threshold **0.85** (harmonic mean)
- **Scoring model**: flat weighted 4-dimension (no `evaluator_mode: hierarchical`)
- **Remediation under review**: `0e34ac7a3` (S1 code fix) + `209c8f9f9` (docs/metadata)

All scores below were **re-derived from the current tree**, not carried forward.

---

## 1. Verdict

| Dimension | Prior | Now | Δ | Verdict |
|-----------|------:|----:|---:|---------|
| Functionality | 0.80 | **0.88** | +0.08 | PASS |
| Security | 0.70 | **0.82** | +0.12 | PASS (no Critical/High — hard threshold not tripped) |
| Craft | 0.82 | **0.84** | +0.02 | PASS |
| Consistency | 0.68 | **0.80** | +0.12 | PASS (weakest dimension) |

**Harmonic mean** = 4 / (1/0.88 + 1/0.82 + 1/0.84 + 1/0.80)
= 4 / (1.136364 + 1.219512 + 1.190476 + 1.250000) = 4 / 4.796352 = **0.834**

## OVERALL VERDICT: **FAIL** (0.834 < 0.85) — narrowly, and for different reasons than before

The remediation is real and substantial. **Five of seven findings are genuinely fixed**, and I
reproduced each fix rather than accepting the commit message. The critical F1 defect — which
would have exempted this SPEC from lifecycle drift detection permanently — is confirmed gone via
the domain's own tool. S1 is confirmed closed at the boundary and reproduced.

The residual FAIL splits into two distinct causes, and the distinction matters for what to do next:

1. **The two user-deferred findings (F3, S2)** — these hold Security and Functionality below the
   bar. Declaring them does not repair them (see §2 for my treatment).
2. **A cluster of docs-site / declaration inconsistencies that were NOT part of the user's
   deferral decision** — the `--harness-profile` documentation half of F2, F7, C2, and six new
   findings introduced by the remediation itself. This second cluster is fixable inside this PR
   and is the difference between 0.834 and a pass.

**Robustness check on the verdict**: the FAIL is not an artifact of score-tuning. Even granting
Functionality 0.90 and Consistency 0.82, the mean is 0.844 — still below 0.85. To pass, Security
would have to rise, and nothing about S2 changed.

**Null hypothesis considered**: *"did the remediation actually fix anything, or did it just write
prose about the findings?"* — It genuinely fixed things. F1 is tool-confirmed, S1 is
binary-reproduced, F5/F6 are measured. This is not paper-over work. But the remediation is
**documentation-weighted**: 4 of 6 changed files are markdown, and the one code fix is 12 lines.

---

## 2. Treatment of declared-and-deferred debt (F3, S2) — stated explicitly

This is the crux of the re-score, so I state the rule I applied before applying it.

**An accurately declared deferral converts an *undisclosed defect* into a *known, bounded risk*.
It substantially reduces the Consistency penalty and partially reduces the Functionality penalty.
It reduces the Security penalty by approximately zero.**

Rationale: Consistency measures whether the artifacts describe the tree. Once the CHANGELOG and
`progress.md` §G correctly state that `--enable-lsp` diverges and that the writes are non-atomic,
the artifact-vs-tree contradiction is *gone* — the documents are now true. That penalty is
legitimately discharged by declaration.

Security and Functionality measure the shipped artifact, not the honesty of its description. A
declared non-atomic write truncates a config file exactly as often as an undeclared one. Writing
"this may corrupt your config" in a CHANGELOG is disclosure, not mitigation.

Applied:

| Finding | Consistency penalty | Functionality / Security penalty |
|---------|--------------------|----------------------------------|
| F3 (declared, accurate) | discharged in full | retained, reduced ~40% (Functionality) |
| S2 (declared, one false clause) | mostly discharged | **retained in full** (Security) |

This is why Security only reaches 0.82 despite S1 being cleanly closed: S2 is a major durability
finding on which literally nothing changed, and its declaration buys no durability.

**Are the declarations accurate and complete?**

- **F3 — ACCURATE and COMPLETE.** It states the actual defect, not a softened one: it names the
  cobra default (`false`), the wizard default (`true`), the user-visible consequence
  (`moai init --non-interactive` yields `lsp.enabled: false`), and the sibling-flag asymmetry. I
  reproduced the consequence with the real binary. It does not hedge.
- **S2 — ACCURATE on the defect, but carries one FALSE supporting claim.** "Seven config writers
  use non-atomic `os.WriteFile`" is correct (verified: 7). The clause **"while
  `internal/atomicfile.Replace` ships unused" is false** — `Replace` has 5 production call sites
  (`preference/filestore.go:495`, `harness/tier/tier.go:296`, `migration/version.go:74`,
  `hook/handoff/persist.go:312`, `goal/state.go:92`). My own prior finding said "unused *here*";
  the qualifier was dropped in transcription. This does not *soften* the defect — if anything it
  overstates the neglect — but it is an unverified claim in an audit record, which is exactly the
  class of error the record exists to prevent. The §G row also omits the TOCTOU window; the
  CHANGELOG (g) is the better of the two because it does state the AC-WIZ-010a consequence.

**The declaration set is INCOMPLETE.** §G presents "Findings fixed" + "Findings deferred" as a
complete resolution record of a 14-finding audit. Three findings appear in neither table and are
still live in the tree: **C2 (major)**, **F7 (minor)**, **S3 (minor)**. C1 is acceptable — it is
declared as CHANGELOG debt (a). A resolution record that silently drops a major finding is worse
than one that lists it as "won't fix".

---

## 3. Per-finding re-verification

Legend — **CONFIRMED** = reproduced in this tree this pass. **PLAUSIBLE** = judgment, not reproduced.

### 3.1 RESOLVED

| ID | Prior sev | Verification performed this pass | Status |
|----|-----------|----------------------------------|--------|
| **F1** | critical | Built `moai` from `209c8f9f9`; ran `moai spec audit --json` myself. Verbatim: `{"spec_id": "SPEC-CLI-WIZARD-RESTRUCTURE-001", "era": "V3R6", "finding_type": "EraAutoDetected", "severity": "INFO", "details": {"heuristic_matched": "H-4 (§E.2 + §E.4 + sync_commit_sha)"}}`. **Value correctness**: replicated both `era.go` regexes against `progress.md` — `sync_commit_sha` resolves via the YAML pattern at **line 403** to exactly **`adee2f46b`**, with no trailing prose (`sync_status`→`complete` L401, `sync_complete_at`→`2026-07-25` L402). **Not polluted**: the three later occurrences of the token (L413 prose, L418 prose, L568 table row) match neither pattern — verified by running the regexes, not by eye. **SHA is the right one**: `git log -S'status: completed' -- spec.md` → `adee2f46b`, the sync commit. **No `§E.*` heading broken**: all of §E.1/§E.2/§E.2.1-2.6/§E.3/§E.3.1/§E.4 present as headings; `§E.5` occurs **0** times in the file, so the legacy H-4 branch cannot be spuriously triggered by the new prose. | **RESOLVED — CONFIRMED** |
| **S1** | major | **Wiring**: `PreRunE: validateInitFlags` at `internal/cli/init.go:62` — the validator is reached, not merely defined. **Enum agreement**: validator `{personal, team}` (init.go:233) == flag help `"Project mode: personal or team (default: personal)"` (init.go:87) == wizard `project_mode` options `personal`/`team` (questions.go:444-445). **Reproduced against the real binary**: `moai init … --project-mode $'personal\ninjected_key: pwned'` → `Invalid --project-mode value …: must be one of: personal, team`, **exit 1**, and `project.yaml` is **never created** (init aborts pre-write). `--project-mode bogus` → exit 1. `--project-mode team` → exit 0, `project.yaml:15 mode: team` — no false positive. **Tests non-vacuous**: `TestValidateInitFlags_InvalidProjectMode` sets the flag via `initCmd.Flags().Set` and asserts `err != nil` on 4 out-of-set rows (2 newline-bearing). If the validator were removed, `validateInitFlags` returns nil → all 4 rows fail. Non-vacuity is structural, not assumed. The `resetInitFlagsForProjectMode` helper prevents cross-test flag bleed on the shared global `initCmd`. | **RESOLVED — CONFIRMED** |
| **F4** (+ CHANGELOG half of F2) | major | The corrected sentence — *"the pre-existing `--harness-profile` CLI flag is still accepted and parsed, but it currently has **no persisted effect**"* — is **TRUE against the tree, not merely softer**. Verified two ways: (1) `writeHarnessProfileYAML` has **zero** production callers (only `initializer_expansion_test.go:42,66`); (2) real binary — `moai init … --harness-profile strict` → `harness.yaml:7 default_profile: "default"`, flag ignored. | **RESOLVED — CONFIRMED** |
| **F5** | minor | All six artifacts re-read: `spec/plan/acceptance/design/research/progress` = `status: completed`. Uniform. | **RESOLVED — CONFIRMED** |
| **F6** | minor | **Re-measured myself, did not trust the CHANGELOG**: `go test -count=1 -cover ./internal/template/...` → `coverage: 86.0% of statements`. Matches the corrected figure exactly. (See N5 for the baseline clause attached to it.) | **RESOLVED — CONFIRMED** |

### 3.2 STILL-OPEN — declared

| ID | Sev | Location | Status this pass |
|----|-----|----------|------------------|
| **F3** | major | `internal/cli/init.go:87-91, 387` | **Unchanged, accurately declared.** Reproduced: `moai init $T --non-interactive` (no flags) → `lsp.yaml:45 enabled: false`. Flag registration confirms the asymmetry: `Bool("enable-lsp", false)` vs `Bool("enforce-quality", true)` / `Bool("enable-design", true)`. Declaration is complete and does not soften. |
| **S2** | major | `internal/core/project/initializer_expansion.go` (7 × `os.WriteFile`) | **Unchanged; declaration accurate on the defect, false on one supporting clause** (see §2). `grep -c 'os.WriteFile'` → **7**. `atomicfile.Replace` → **5 production call sites**, so "ships unused" is false. |
| **C1** | minor | `initializer_expansion.go:91` + 2 tests | Unchanged; correctly declared as CHANGELOG debt (a). Accepted **as declared debt**. |

### 3.3 STILL-OPEN — undeclared

| ID | Sev | Location | Evidence | Status |
|----|-----|----------|----------|--------|
| **F2-docs** | major | `docs-site/content/{en,ko,ja,zh}/getting-started/init-wizard.md:165`; `.../getting-started/cli.md:70`; `.../cli-reference/init.md:49` | **Only the CHANGELOG half of F2 was fixed.** The 4-locale docs — rewritten *by this SPEC's own sync phase* — still ship `--harness-profile default \` inside the recommended copy-paste non-interactive CI example, and 8 flag-table rows still present the flag with its four accepted values and no inert marker. The CHANGELOG now says the flag persists nothing; the docs still teach users to pass it. **The internal contradiction is fresh**: before the fix, CHANGELOG and docs agreed (both wrong); now they disagree. §G marks F2 "fixed" without noting the docs half survives. | **CONFIRMED** |
| **C2** | major | `internal/cli/init.go:382` | Comment *"Page-3 non-interactive overrides — defaults match wizard defaults (REQ-IWE-008)"* sits directly above `LSPEnabled: getBoolFlag(cmd, "enable-lsp")` (cobra default `false`) while the wizard default is `true`. Still factually false. **Absent from both §G tables.** | **CONFIRMED** |
| **F7** | minor | `docs-site/content/*/getting-started/init-wizard.md:123` vs `.../cli.md:71` and `.../cli-reference/init.md:50` | init-wizard.md: *"The default is **enabled (Yes)**"*; cli.md / cli-reference: *"(default: false)"*. Both present in all 4 locales. Each is true in its own context; a reader cannot reconcile them, and neither page cross-references the other. Partially mitigated by the debt (f) declaration, which a CHANGELOG reader (not a docs reader) would see. **Absent from §G.** | **CONFIRMED** |
| **S3** | minor | `initializer_expansion.go:118,200` | Unbounded `os.ReadFile` of the patch target. Unchanged. **Absent from §G.** | **CONFIRMED** |
| **S1-residual** | minor | `initializer_expansion.go:56` `writeProjectModeYAML` / `patchYAMLKey` | **Asked explicitly: is flag-boundary validation sufficient?** **For the current producer set, yes.** I enumerated every production producer of `opts.ProjectMode`: (1) `init.go:385` `getStringFlag(cmd,"project-mode")` — now validated at `PreRunE`; (2) `init.go:139-140` from `result.ProjectMode`, whose only writer is `wizard.go:403` fed by a huh **select** constrained to the two options. No third producer exists in-module. So there is **no live injection path**, and I do not score this as an open vulnerability. It remains a **defense-in-depth gap**: the writer still interpolates raw (`"  " + key + ": " + newValue`), so validation is a property of the *callers*, not of the *writer*. A future third producer silently re-opens S1. Note the latent coupling with debt (a): `--harness-profile` is likewise unvalidated and is safe **only because its writer is dead** — restoring `writeHarnessProfileYAML` without adding a validator re-introduces the identical injection. | **CONFIRMED (residual, not live)** |

### 3.4 NEW — introduced or exposed by the remediation itself

| ID | Sev | Location | Evidence | Status |
|----|-----|----------|----------|--------|
| **N3** | minor | `internal/template/templates/.moai/config/sections/project.yaml.tmpl:11,14` vs `internal/cli/init.go:233` | **The S1 fix contradicts the config template this tool itself deploys.** The template comments read `# Project mode: personal, team, or enterprise` and `# - enterprise: Full governance and compliance features`. The new enum rejects it: `moai init --non-interactive --project-mode enterprise` → **exit 1**. Before `0e34ac7a3` this was accepted and written verbatim. Functional impact ≈ 0 (nothing in the codebase consumes `enterprise`; the 12 docs-site flag rows already said `<personal\|team>`), but a user following the comments in their own deployed `project.yaml` now hits a hard failure on a previously-working value. Fix: drop `enterprise` from the template comments, or widen the enum. | **CONFIRMED** |
| **N4** | minor | `progress.md` §G S2 row; `CHANGELOG.md` debt (g) | "`internal/atomicfile.Replace` ships unused" — false repo-wide (5 production call sites). Unverified claim inside the audit record. | **CONFIRMED** |
| **N5** | minor | `CHANGELOG.md` coverage clause | *"both are above the package's own unchanged baseline"* — **both are exactly equal to it, not above it.** Measured by `git archive` + `go test -cover` at three points: pre-merge base `b1ea545e2` → **84.9%**; run-phase tip `216dcbb2e` → **84.9%**; `origin/main` → **86.0%**; HEAD → **86.0%**. The underlying no-regression claim is TRUE; the word "above" is not. A stale figure was replaced by a smaller overstatement. | **CONFIRMED** |
| **N6** | minor | `progress.md` §G `sync_audit_report:` | §G cites `.moai/reports/sync-audit/SPEC-CLI-WIZARD-RESTRUCTURE-001-2026-07-25.md` as its evidence path. That file is **untracked** and appears in **no commit** (`git log --all -- <path>` → empty). Three sibling reports in the same directory **are** tracked, so the repo convention is to commit them, and the directory is not gitignored. The audit record therefore cites evidence that will not exist for any other reader — an evidence-persistence failure. | **CONFIRMED** |
| **N7** | minor | `progress.md` §G | The "Findings fixed / Findings deferred" tables omit **C2 (major)**, **F7**, **S3** while presenting themselves as the complete resolution record of the audit. | **CONFIRMED** |
| **N8** | minor | `CHANGELOG.md` debt (f)/(g); §G "Findings deferred (user-scoped, follow-up SPEC)" | Both surfaces route the deferred debt to "a follow-up SPEC". **No such SPEC exists** — no `.moai/specs/` entry references `enable-lsp` or `atomicfile` other than this SPEC and the unrelated `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`. Debt routed to a non-existent destination is operationally indistinguishable from dropped debt. Fix: create the SPEC, or reword to "deferred, not yet scheduled". | **CONFIRMED** |

### 3.5 Regression sweep on the remediation commits

Checked specifically, since the remediation is now in scope:

- **§E.4 restructure broke no other heading** — all `§E.*` headings intact; `§E.5` count 0 (no spurious legacy-heuristic trigger). **No regression.**
- **§G addition did not pollute the parser** — replicated regexes confirm `sync_commit_sha` still resolves to `adee2f46b` from L403; §G's own `sync_audit_*` keys collide with nothing `era.go` reads. **No regression.**
- **4-locale parity intact** — remediation touched **0** docs-site files (`git diff --name-only da85ad13e..HEAD -- docs-site/` → 0). Re-measured anyway: init-wizard 28/10, cli.md 44/161, cli-reference/init.md 14/37 — identical across en/ko/ja/zh. **No regression.**
- **`hugo --gc --minify`** re-run: **exit 0, 0 WARN/ERROR**, 153/151/151/151 pages. **No regression.**
- **Full suite** `go test -count=1 ./...`: **exit 0, 105 ok, 0 FAIL**. **`golangci-lint run --timeout=3m`: exit 0, "0 issues."** **No regression.**
- **Other coverage figures re-measured**: `internal/cli/wizard` **93.9%**, `internal/core/project` **88.7%** — both match the CHANGELOG. (`internal/cli` 75.2%, pre-existing, unclaimed.)
- **New claims in the remediation commit messages** — `0e34ac7a3` and `209c8f9f9` each assert verification results; every asserted result I checked (era V3R6, 86.0%, exit-0 suite, hugo clean) reproduced. The two *inaccuracies* found (N4, N5) are in the artifact prose, not in the commit-message verification claims.

---

## 4. Dimension reasoning — what moved and why

### Functionality 0.80 → 0.88 (+0.08)

**Up**: F1 was the single largest functional deduction — the sync deliverable's own close marker
was unreadable by the tool that consumes it, and the SPEC was misclassified into a
grandfather-protected era. That is now tool-confirmed fixed, with the extracted value verified
correct rather than merely non-empty. The 19 ACs remain green under my own suite + lint run.

**Held back**: two major behavioural gaps survive in shipped code — `--harness-profile` accepts
input and does nothing (F2), `moai init --non-interactive` yields an LSP default that contradicts
both the wizard default and its sibling flags (F3). Declaring them does not change what the binary
does. N3 subtracts a hair: a template-documented value is now rejected.

### Security 0.70 → 0.82 (+0.12)

**Up**: S1 was the only *live* finding and it is cleanly closed — validator reached via `PreRunE`,
enum agreeing across all three surfaces (validator / flag help / wizard options), injection
reproduced as rejected with exit 1 and no file written, and a non-vacuous reproduction-first test.
I additionally closed the producer-set question the fix leaves open, and found the current producer
set genuinely closed.

**Held back**: S2 is unchanged and unmitigated — 7 non-atomic writes against the same ~22 KB config
set that AC-WIZ-010a exists to protect. Its declaration buys zero durability, and the declaration
itself carries a false supporting clause (N4). S3 unchanged and now undeclared. The residual
writer-level exposure keeps validation a caller property rather than a writer invariant.

No Critical/High finding → the hard threshold does not trip and Security does not force an
overall FAIL on its own.

### Craft 0.82 → 0.84 (+0.02)

**Up**: the S1 fix is above-average work. Reproduction-first with four discriminating rows (two
newline-bearing), a flag-reset helper that defends against cross-test bleed on the shared global
`initCmd`, assertions on the error message naming the closed set, empty-value semantics matched to
every sibling validator, and an in-code comment explaining *why* this SPEC introduced the exposure.
12 lines of production code for the whole fix — proportionate, no over-engineering.

**Held back, nearly cancelling the gain**: C2 (major) is still live and now also undeclared; the
§G resolution record is incomplete (N7) and carries two inaccurate claims (N4, N5); the cited
evidence path does not exist in the repo (N6). Craft includes the craft of the audit record, and
that record is the weakest artifact in the remediation.

### Consistency 0.68 → 0.80 (+0.12)

**Up**: this is where the remediation did its best work. F1's marker now matches every sibling
SPEC's form; all six artifacts are uniformly `completed`; the coverage figure matches a fresh
measurement; and the `--harness-profile` CHANGELOG sentence is now true rather than merely hedged.
The docs-site 4-locale parity and warning-free build are intact and re-verified.

**Held back — and this is the binding cluster**: the inconsistencies now concentrate in the SPEC's
own sync deliverable. The 12 rewritten docs-site files still teach an inert flag in a copy-paste CI
example (F2-docs) and still contradict each other on the LSP default (F7), while the CHANGELOG has
moved on. The shipped config template contradicts the new validator (N3). The audit record cites a
file that does not exist (N6), omits a major finding (N7), and points the deferred debt at a SPEC
that has not been written (N8). None of these were part of the user's deferral decision.

---

## 5. Routed defect-list (ordered — the first three are what stand between 0.834 and a pass)

| ID | Required fix |
|----|--------------|
| **F2-docs** | Either (a) drop `--harness-profile default \` from the non-interactive example in all 4 `init-wizard.md` files and mark the 8 flag-table rows as "accepted but currently inert (see debt (a))", or (b) restore the writer via `patchYAMLPathValue(content, "harness.default_profile", …)`. Whichever is chosen, code + docs + CHANGELOG must state the same thing. |
| **C2** | `internal/cli/init.go:382` — either change `LSPEnabled` to `getBoolFlagWithDefault(cmd, "enable-lsp", true)` (which also closes F3 and F7), or rewrite the comment to say the LSP default deliberately diverges and why. The comment as written is false. |
| **N7 / N4 / N5 / N6 / N8** | Repair the audit record: add C2, F7, S3 to §G (fixed or won't-fix); strike "ships unused" (5 production call sites exist) or restore the "here" qualifier; change "above the … baseline" to "unchanged from"; commit the cited report file (siblings are tracked) or cite a path that will exist; name the follow-up SPEC or say it is unscheduled. |
| **F7** | Resolved automatically by the C2/F3 fix; otherwise add a cross-reference so init-wizard.md and cli.md stop contradicting each other. |
| **N3** | Remove `enterprise` from `project.yaml.tmpl:11,14`, or widen the validator enum. |
| **S1-residual** | Defense-in-depth: reject `\n` / `\r` / `:` inside `patchYAMLKey` / `patchYAMLPathValue`, or quote the emitted scalar — so the invariant lives at the writer, not in each caller. Add `--harness-profile` to `validateInitFlags` before debt (a) is ever resolved. |
| **F3, S2** | User-deferred. No action requested; they remain the arithmetic reason Security and Functionality sit where they do. |

---

## 6. Gaps — what I did NOT verify this pass

Stated explicitly so nothing unobserved passes as a pass.

1. **The 19 ACs were not individually re-derived this pass.** I re-ran the full suite (105 ok / 0 FAIL), lint (0 issues), and the three coverage figures, and I verified the only production delta since the audited state is the +12-line `init.go` change. The per-AC greps and the byte-size / indentation-multiset checks from the prior pass are **carried forward with attribution to `da85ad13e`**, not re-observed at `209c8f9f9`.
2. **Windows cross-platform build** (`GOOS=windows GOARCH=amd64 go build ./...`) — NOT RUN, in either pass.
3. **`go vet ./...`** — not run separately.
4. **`-race`** — not run on any package.
5. **Interactive wizard rendering** — never driven through a TTY. The claim that the wizard `project_mode` select cannot emit an out-of-enum value rests on source inspection of `questions.go:443-446` + `wizard.go:403`, not on an interactive run.
6. **Reconfigure path** (`moai update --reconfigure`) — not executed; I did not verify that path cannot construct `InitOptions.ProjectMode` outside the two producers I enumerated. My producer-set enumeration is a whole-module grep, which would catch a direct assignment but not a reflective or config-driven one.
7. **Non-vacuity of the S1 tests was established structurally, not by mutation.** I did not remove the validator and observe the tests fail — the tree is read-only for this audit. The argument is deductive: the negative test asserts `err != nil` on values the validator is the only thing rejecting.
8. **Rendered docs-site HTML** — hugo exits clean, but I inspected markdown source only, not rendered output, links, or shortcode expansion.
9. **The 3 non-Latin locales' full documents** — parity re-measured by heading/row counts and the specific `--harness-profile` / `--enable-lsp` lines were read in all 4 locales; I did not semantically re-diff the remaining body text.
10. **Block-scalar edge case** in `patchYAMLPathValue` (declared debt (b)) — no failing fixture constructed, in either pass.
11. **Whether `enterprise` was ever functional** — I confirmed nothing consumes it today; I did not trace its history to determine whether it once did.
12. **The `base_merge` divergence figures** in §E.4 — not re-run against historical states.

## 7. Residual risk — what could still be wrong despite what I DID observe

- My coverage baselines (84.9% / 86.0%) come from `git archive` snapshots built in a scratch
  directory sharing this machine's module cache. A dependency-resolution difference between that
  snapshot and a clean CI checkout would change the figures.
- N6's consequence is self-concealing: once this PR merges, `progress.md` §G will cite an audit
  report that no reader can open, and nothing will surface that until someone follows the link.
- The S1 fix's sufficiency is a **statement about today's producer set**. It is correct now and
  silently becomes incorrect the moment a third producer is added. Nothing in the tree enforces it.
- Two known-flaky CI tests exist in this repo (preference `-race`, session
  `TestRegisterSessionConcurrent`). My suite run was green; a single green run does not establish
  stability, and I did not run with `-race`.
- The prior pass caught two factual errors in this SPEC's own sync reporting; this pass caught two
  more (N4, N5) in the *remediation* prose. The rate of prose-vs-tree drift in this SPEC's
  artifacts is non-trivial, so further unfound instances in the ~180 lines of §G / CHANGELOG text
  I did not line-by-line verify are plausible.
