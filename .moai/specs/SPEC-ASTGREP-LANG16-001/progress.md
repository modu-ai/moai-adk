# SPEC-ASTGREP-LANG16-001 — Progress

Card t228, Class C, Tier L. Worktree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`,
HEAD `294b4b6ab`.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-25
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md]
requirements: 23
acceptance_criteria: 19
milestones: 3
scope: contract (M1-M3)
successor: SPEC-ASTGREP-BREADTH-001
audit_iteration: 5 audits (0.68 -> 0.69 -> 0.71 -> 0.84 -> PASS 0.88)
spec_version: 0.7.0
audit_verdict: PASS 0.88 (iter5); open minor findings: N5 §A.6 half (note-half closed), N6
gate_recheck: ".moai/reports/plan-audit/SPEC-ASTGREP-LANG16-001-2026-08-27.md (narrow-delta; F1+F2-note fixed here)"
```

Plan-phase basis: `.moai/reports/t228/plan-measurements.md` (M1-M4; M5 superseded — the
differential corpus is committed on `origin/main` via `a9eb896ce` / PR #1637 / card t227), plus
`.moai/reports/t228/plan-audit-iter1.md`.

Revision history: lead amendments A1-A3 at v0.2.0; corrections C1-C7 at v0.3.0; plan-audit
iteration 1 revision and the D8 split at v0.4.0; iteration 2's two-column criteria rewrite at
v0.5.0; iteration 3's bounded propagation-only pass at v0.6.0.

**Split (D8, binding).** This SPEC now carries the contract half only — M1-M3. The breadth half
(M4-M10) moved to `SPEC-ASTGREP-BREADTH-001`. Total card scope is unchanged: requirements (2) and
(4) are discharged by the successor, in full, with no external gate between the two SPECs. The
split was forced by the budget — REQ and AC were saturated at 25/25, so the audit's own fixes
could not be added without breaching the Tier L ceiling.

Budget after revision: 23 requirements, 19 acceptance criteria against a Tier L ceiling of 25 —
headroom of 2 and 6 respectively. The propagation pass at v0.6.0 added **no** requirement and
**no** criterion; D1, D3, and D2 are amendments to existing entries.

Iteration-2 findings addressed (v0.5.0, `acceptance.md`): E1 the central criterion passed
vacuously — now asserts count-equals-rule-count, which also discharges REQ-A16-010; E5 REQ-A16-010
had no criterion; E8 the corpus-baseline criterion gained a requirement via renumbering. E2, E3,
E4, E6, E7, and E9 were applied to `acceptance.md` only and their other halves were left
unpropagated — that gap is what iteration 3 found, and what v0.6.0 closes.

Iteration-3 findings addressed (v0.6.0, propagation-only across the five artifacts iteration 2 did
not open): D1 requirement/criterion contradiction on `catalog.yaml` — REQ-A16-018 inverted to the
measured truth, `plan.md` §A.3 and §D corrected; D2 AC-A16-016's citation-or-probe half given a
requirement (REQ-A16-011 third clause) and a work item (`plan.md` M3 item 4); D3 REQ-A16-021 and
REQ-A16-001 re-scoped off `internal/template/templates/**`, with §0's discipline extended to the
requirement layer; D4 the pinning test given a path, an M1 item, an explicit PRESERVE carve-out,
and a restated §E diff command; D5 and D11 `design.md` §3.1/§4 D13 placement stated one way plus
the four-class count and three stale citations; D6 nine renumbering sites; D7 four §I rows
re-derived; D8 this signal block; D9 the two missing HISTORY rows; D10 AC-A16-001's open count
clause bound to a derivable number.

Iteration-1 findings addressed: D1 (corpus gate skips, not fails — four passages corrected, no-skip
obligation added), D2 (R2 mitigation withdrawn, `metadata.cwe` anchor added), D3 (archived
`SPEC-ASTG-UPGRADE-001` reconciled, R6 owned here), D4 (contiguous renumbering), D5 (REQ-A16-012
in `Where … shall` form), D6 (corpus baseline given a criterion), D7 (checker class 1 is a set
comparison, class 4 added), D8 (split), D9 (AC-A16-020 names its sanctioned exemption — that criterion was folded into AC-A16-019 at
v0.5.0; the id is retained here as the audit trail of what iteration 1 asked for), D10
(dangling reference removed), D11 (tautological criterion retired). Optional findings applied:
D12 (`catalog.yaml` named), D13 (rule-tests outside the distributed tree), D14 (neutrality widened
to any human-language text), D15 (unobservable "shall read" moved to `plan.md`), D16 (stray
directory in the pre-flight).

## §E.2 Run-phase Evidence

M1 (card t228), tree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`.
Commits: `20657baa3 -> 5c02f1546` (harness/assets/design-R1/frontmatter transition)
and the pinning-test commit this block rides. Nothing pushed (git-flow override B9;
integration lands into develop by the lead).

### Pre-flight (all verbatim, before first edit)

| check | command | result |
|---|---|---|
| sg version | `sg --version` | `ast-grep 0.40.5` rc=0 (matches SPEC pin A2) |
| branch/HEAD | `git rev-parse --short HEAD && git branch --show-current` | `20657baa3 / WT-astgrep-16-langs` |
| cleanliness | `git status --porcelain \| wc -l` | `0` |
| module build | `go build ./...` | exit 0 |
| B2 conflict scan | `grep -rn 'Retired\|superseded\|TestHarnessRetirement' internal/astgrep/ internal/cli/` | no ast-grep-related conflicts (hits only unrelated cli subsystems) |

§A.5 pre-flight residue removal: stray empty dir
`internal/template/templates/.moai/config/astgrep-rules/security/.claude/agent-memory/manager-spec/`
verified file-less (`find ... -type f | wc -l` → `0`) then removed; untracked so
`git status` stayed `0`.

### Protected-before baseline — template ruleset SHA256s at HEAD 20657baa3 (pre-edit)

```
934e9eed15b7551a9233840868511b8872741ce71e18fe052eacaefe1ffd80da  go/concurrency.yml
935741b6b2db21938ededa54d4f49e6c5b0016a1562371cb757da4c1c7e66afa  go/error-handling.yml
6586624ceb3759e4712536448d4b1ade0fd6d43ac2e3e1ba3a55858e85d41c07  go/hardcoding.yml
5e5f19ce089a4b09a5c3e5a06e03f65bb03940f6008c8e22b3a1a48a40b52cff  go/idioms.yml
d648495afdb8c62938b50fade2d079c059990f9b0e9f5ad7981e7ac354069161  go/resource-safety.yml
e227766eb1e9654f8e4b811091eb1a48011299854894a8c8cd66368eba845168  security/credentials.yml
9fa00d04c869a93787ec9b3b2d480dbf9bc174efeae512299bbe908c011b378d  security/crypto.yml
1ab931e02a362682b455b6adfa9f6216dd38a5fc990f17b4feb9446e17e5f690  security/injection.yml
15446281266aecca6b60afdba5ae46df9c0db9010a808147701220c887e770db  security/secrets.yml
3da57a2cadd7cee8bc785ba97463d4eea0bbc270efcbd92a15752b08c42d46b1  security/web.yml
54ee886e21d1c6a4986d1304aba5fd18f0808727040c78f291a39da0982cf40e  sgconfig.yml
```

### R1 id-keying — measured, not assumed

Full transcript recorded in `design.md` §3.3 (settled): ast-grep keys cases by id ALONE with one
global snapshot name; silent drop ("Configuration not found!") for unmatched ids; inline snippets
are single-language-routed. Decision: convention unchanged (rename branch blocked by PRESERVE
`internal/hook/security/prefilter_test.go` pinning `sec-hardcoded-credential` across four
languages). Reported count therefore 21 (= distinct rule ids over 26 rules) — recorded as the
AC-A16-001 deviation below rather than declared.

### Latent-dead patterns surfaced (each fixed structurally, metadata untouched)

Four shipped patterns matched nothing under ast-grep 0.40.5 before M1 gave them cases:
`sec-template-injection-html`, `go-defer-in-loop`, `sec-hardcoded-api-key`,
`sec-hardcoded-jwt-signing-key`. Each now uses the vetted structural form
(`kind:`+anchored-regex or all/inside composition); real-file scan probes fired and the
negative controls stayed quiet before the change was applied.

### Harness green + rule mutants (plan M1 item 5)

```
$ sg test --config internal/template/templates/.moai/config/astgrep-rules/sgconfig.yml
test result: ok. 21 passed; 0 failed;
rc=0                                       (post-revert rerun; also after make build)

# Mutant A: sec-weak-hash-md5 pattern -> zzz("sh")
[Missing] Expect rule sec-weak-hash-md5 to report issues, but none found in:
FAIL sec-weak-hash-md5  .M
Error: test failed. 20 passed; 1 failed;     mutA-rc=4   (reverted)

# Mutant B: go-interface-empty-not-any pattern -> var cache map[string]$T
[Noisy] Expect go-interface-empty-not-any to report no issue, but some issues found in:
[Wrong] go-interface-empty-not-any snapshot is different from baseline.
Error: test failed. 20 passed; 1 failed;     mutB-rc=4   (reverted; post-revert sg test rc=0)
```

### Corpus pinning test — RED probes on known failing inputs (instruments proven live)

Probe A — appended one comment byte to `py_deny_os_system.py`:
```
--- FAIL: TestAstgrepCorpusFixturesPinned
    py_deny_os_system.py changed since 294b4b6ab: sha256 = da3814..., want 3d2b44...
FAIL github.com/modu-ai/moai-adk/internal/hook
```

Probe B — flipped one wantDeny value:
```
--- FAIL: TestAstgrepCorpusTablePinned
    py_clean.py: wantDeny flipped from false to true since 294b4b6ab — a flip is how
    a red differential row turns green without any rule change; ...
```

Probe C — added "rust" to coveredCorpusLanguages (green-by-skip condition):
```
pre_tool_scan_differential_test.go:242: corpus rejected: no denying fixture for covered language(s) rust
--- FAIL: TestAstgrepCorpusRunDoesNotSkip
    differential re-run SKIPPED; green-by-skip proves nothing.
    differential re-run reported a rejected corpus: ...
```

All three restored (`git checkout -- <path>`; restored digest re-printed matching base blob);
final pin suite `PASS ok internal/hook` again.

### AC-A16-006 negative validation (deployed-shaped tree without repo-side root)

```
$ cp -R $T/* tmp/proj/.moai/config/astgrep-rules/ ; sample.go absent-root scan:
$ sg scan --config .moai/config/astgrep-rules/sgconfig.yml sample.go
error[sec-weak-hash-md5]: MD5 is unsuitable for security use. ...
scan completed; rc=0 path-wise (finding expected; "1 error(s) found" notice only)
```
No configuration error; `testConfigs` present; `ruleDirs` unchanged ([go, security]).

### AC-A16-008 standing diff

REBASE RE-PIN (the §E baseline rule applied): measured literally against plan base `294b4b6ab`
the standing diff shows 40 files / +6776 — all of it UPSTREAM work that entered through the
2026-08-27 rebase onto origin/develop (t227/t217 lineage), none of it ours. Re-pinned to the true
fork point:

```
$ git merge-base HEAD origin/develop        -> d29b8942e
$ git diff --stat d29b8942e -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
(no output)                     # vs fork point: only the carve-out file exists; every
                                # pre-existing hook file byte-unchanged by THIS SPEC
```

Corpus fixture digests and the extracted wantDeny/covered-language pins live as constants inside
`internal/hook/astgrep_corpus_pin_test.go` (blob-level equality against `294b4b6ab` cross-checked
at authoring time; the source-file prose delta from t217's upstream assertion-(ii) is deliberately
NOT whole-file-pinned — only the table values are, per the criterion's wording).

### Builds / vet / lint / boundary

```
$ make build            → success; catalog.yaml diff EMPTY (plan §D requirement)
$ go build ./...        → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/template/... ./internal/astgrep/...
                        → exit 0
$ go test -count=1 ./internal/template/...   → ok (after make build: stale-embed trap excluded)
$ go test -count=1 -run 'TestAstgrepCorpus|TestScanWriteContentDifferential' ./internal/hook/
                        → ok (rerun after make build)
$ golangci-lint run --timeout=2m ./internal/hook/... ./internal/astgrep/... ./internal/template/...
                        → 0 issues (two NEW findings introduced mid-work — errcheck fh.Close,
                          staticcheck De Morgan — fixed in-place before completion)
$ grep -rn 'AskUserQuestion\|mcp__askuser' <created files> → zero hits (rc=1)

---

## §E.2b Run-phase Evidence — M2 (coverage matrix, evidence rule, four-class checker)

M2 (card t228), tree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`, base HEAD
`075e79344`. Everything below was re-measured in this run; nothing carried forward.

### Item 1 — parser probe re-run (NOT copied forward)

```
$ sg --version
ast-grep 0.40.5                                   rc=0   (matches SPEC pin A2)

$ printf 'x <- 1\n' | sg run -l r --stdin
error: invalid value 'r' for '--lang <LANG>': r is not supported!
                                                  rc=2

$ printf 'void main() {}\n' | sg run -l flutter --stdin
error: invalid value 'flutter' for '--lang <LANG>': flutter is not supported!
                                                  rc=2
```

Full 16-language axis re-derivation (`sg run -p 'zzz' -l <L> --stdin`, one invocation per id):
14 ACCEPTED (go, javascript, python, typescript, rust, java, kotlin, csharp, ruby, php, elixir,
cpp, scala, swift) / 2 REJECTED (r rc=2, flutter rc=2). The 14/2 split is therefore measured
here, not inherited. Transcript: `.moai/state/verify/t228-m2m3/lang-axis-probe.txt`.

### Items 2-3 — the document

`coverage-matrix.md`: 112 cells (8 families × 14 parseable languages), 14 IMPLEMENTED seeded from
the measured inventory, 98 PENDING for `SPEC-ASTGREP-BREADTH-001`. Excluded-languages record
carries both languages, the verbatim rc=2 refusals, the version, and the equal-opportunity idiom.

### Items 4-6 — the checker, and the RED that changed it

RED (authored before the fix, verbatim):

```
--- FAIL: TestCheckerClassesFireThroughDocumentPath/(c)_bare-assertion_exemption
    defect (c) bare-assertion exemption misclassified as dangling rule id:
    [{Class:dangling rule id Key:F3/ruby Detail:names rule id "ruby has no weak-hash sink
     worth flagging" which is absent from the shipped ruleset}]
--- FAIL: TestParseCoverageMatrixRoutesSharedColumnByState
    EXEMPT cell = {... State:EXEMPT RuleID:no csrf token surface in this ecosystem Rationale: ...};
    want Rationale set, RuleID empty
```

Cause: column 4 is shared (`rule id / rationale`) and the parser read it as a rule id
unconditionally. The four classes were proven against hand-built structs but the **wired gate**
reads markdown first, so a document-level defect (c) landed in class 4 instead of class 3 — the
right count of failures under the wrong name, which is the exact confusion the four-class split
exists to prevent. Fixed by routing column 4 on `State` (2 hunks, `ParseCoverageMatrix`).

Two adjacent defects fixed in the same pass, both inside the files this milestone certifies: a
dead test helper (`matrixLanguageAt`, declared and never called — a new `unused` finding waiting
to happen) and an unguarded `strings.Fields(...)[0]` index that panics on an empty leading field.

GREEN — all four classes fire through the document path, each named distinctly:

```
--- PASS: TestCheckerClassesFireThroughDocumentPath
    --- PASS: /(a)_deleted_cell                        (key-set mismatch, names F3/ruby)
    --- PASS: /(b)_substitution_at_constant_count      (key-set mismatch at count 112)
    --- PASS: /(c)_bare-assertion_exemption            (unevidenced exemption; class 4 forbidden)
    --- PASS: /(d)_dangling_rule_id                    (dangling rule id, names F3/go)
--- PASS: TestCoverageMatrixDocumentMatchesShippedRuleset
--- PASS: TestExcludedLanguagesRecordedWithVersionAndProbeOutput
```

Case (c) asserts the *absence* of the classes it must not emit, so a future re-read of column 4
cannot make it green again by landing in the wrong bucket.

### Mutant guard — the wired gate is live on the REAL document

```
# Mutant C: F5/scala row substituted for a duplicate F5/php row (count stays 112)
--- FAIL: TestCoverageMatrixDocumentMatchesShippedRuleset
    key-set mismatch findings on the real matrix: [{Class:key-set mismatch Key:matrix
    Detail:cell key set differs from the 112-key Cartesian product: missing=[F5/scala];
    duplicated/substituted=[F5/php x2]}]
```

Reverted; `cmp` reports the restored document byte-identical; post-revert re-run `ok`.

### M2 follow-up fix — the class-3 positive control was vacuous

Found in lead review of the M2 commit (`49375fd77`), reproduced here before fixing rather than
taken on report. `TestCheckerEvidencedExemptionsPass` was a **vacuous positive control**:

```go
citeCell := base[0]                 // MatrixCell is a VALUE type -- this copies
citeCell.State = StateExempt        // ... so the write never reaches base
```

`fullMatrix()` returns `[]MatrixCell`, so the two writes landed on copies. The matrix handed to
`CheckCoverageMatrix` stayed all-IMPLEMENTED, **no EXEMPT cell reached class 3 at all**, and the
assertion `len(unevidenced) == 0` held regardless of what the acceptance branch did.

Reproduced independently — mutant G disables `cite:`/`probe:` acceptance entirely, so every EXEMPT
cell must be flagged:

```
# coverage_matrix.go:162
-  if !strings.Contains(lower, "cite:") && !strings.Contains(lower, "probe:") {
+  if !strings.Contains(lower, "ZZZNEVERMATCH") {

$ go test -count=1 -run 'TestCheckerEvidencedExemptionsPass|TestCheckerUnevidencedExemptionDetected' -v
--- PASS: TestCheckerUnevidencedExemptionDetected (0.00s)
--- PASS: TestCheckerEvidencedExemptionsPass (0.00s)      <- passes with acceptance DISABLED
ok      0.155s
```

The detection direction of class 3 was guarded; the **acceptance** direction was not. The line-162
logic is correct as written, but nothing would have caught its regression — so plan §B M2 item 4
and the M2 stop condition were not in fact proven by the committed suite.

RED — the corrected control (pointer writes) under the same mutant G:

```
--- FAIL: TestCheckerEvidencedExemptionsPass
    coverage_matrix_test.go:203: cite:/probe:-evidenced exemptions reported unevidenced (2):
    [{Class:unevidenced exemption Key:F1/go Detail:EXEMPT evidence carries neither cite: nor
      probe: ("cite: cpp stdlib reference, process-spawn section")}
     {Class:unevidenced exemption Key:F1/javascript Detail:EXEMPT evidence carries neither
      cite: nor probe: ("probe: sg run -p '...' -l elixir --stdin -> no match")}]
```

Mutant reverted (`cmp` byte-identical; `git diff --stat` on the file empty; zero `ZZZNEVERMATCH`
occurrences remain), then GREEN:

```
--- PASS: TestCheckerUnevidencedExemptionDetected (0.00s)
--- PASS: TestCheckerEvidencedExemptionsPass (0.00s)
ok      github.com/modu-ai/moai-adk/internal/astgrep    0.348s
```

Two guards added beyond the pointer fix, so the control cannot silently go vacuous again: it counts
the EXEMPT cells actually reaching the checker and fails if the writes stop landing, and it asserts
the two cells trip neither class 2 nor class 4 — a wrong-bucket pass is as bad as a vacuous one.

Same-defect sweep across both test files (the fix is worth nothing if a sibling carries it):

| pattern searched | result |
|---|---|
| element copied to a local, then field-written | 3 hits, none defective — one slice reslice (`cells[:0]`, filtering) and two map-value copies read only in assertions |
| range loop-variable copy, then field-written | none |
| surviving mutations | all write through `cells[i]` or `&base[i]` |

`go vet` carries no `-unusedwrite` flag (the analyzer is not in the default vet set), so the sweep
above is the grep, not a claim that a tool cleared it. Gates after the fix: `go test -count=1
./internal/astgrep/...` rc=0; `go vet` rc=0; `gofmt -l` empty; `golangci-lint` 0 issues.

### M2 AC matrix

| AC | claim | evidence | status |
|---|---|---|---|
| AC-A16-009 | key set == 112-key Cartesian product, set comparison not count | `TestCoverageMatrixDocumentMatchesShippedRuleset` PASS; mutant C names both sides of a constant-count substitution | PASS |
| AC-A16-010 | every cell resolves to exactly one state | class-2 branch + document gate PASS; 14/98 accounting asserted | PASS |
| AC-A16-011 | every exemption carries cite:/probe: | `TestCheckerUnevidencedExemptionDetected`, `TestCheckerEvidencedExemptionsPass`, and document-path case (c) PASS | PASS |
| AC-A16-012 | no cell names a rule absent from the ruleset | `TestCheckerDanglingRuleIDDetected` + document-path case (d), resolved against the real shipped ruleset | PASS |
| AC-A16-013 | four synthetic defects, four distinct class names | `TestCheckerClassesFireThroughDocumentPath` 4/4 subtests PASS, (b) load-bearing, (c) forbids misclassification | PASS |
| AC-A16-014 | excluded record: both languages, verbatim refusal, version, equal-opportunity idiom | `TestExcludedLanguagesRecordedWithVersionAndProbeOutput` PASS; probes re-run this session | PASS |

### M2 gates

```
$ go test -count=1 ./internal/astgrep/...                       rc=0   ok
$ go vet ./internal/astgrep/...                                 rc=0
$ GOOS=windows GOARCH=amd64 go vet ./internal/astgrep/...       rc=0
$ gofmt -l internal/astgrep/                                    (empty)
$ golangci-lint run --timeout=3m ./internal/astgrep/...         rc=0   0 issues.
```

new_warnings_or_lints_introduced: 0. No file under `internal/template/templates/**` touched by
M2, so no `make build` and no neutrality surface. Nothing pushed.

Gaps (M2): the full repository suite was NOT run — verification is scoped to the affected package
per the standing local-load prohibition; CI owns the cross-package verdict. `sg test` was not
re-run at M2 (M1 owns it; M2 touches no rule YAML). Residual risk: the 98 PENDING cells assert
nothing, by design — the contract they carry is enforced only when the successor fills them, and a
PENDING cell that is never filled fails no test here.

---

## §E.2c Run-phase Evidence — M3 (severity reclassification and the CWE anchor)

M3 (card t228), tree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`, base HEAD
`49375fd77` (the M2 commit). Re-measured in this run.

### Item 1 — the two-clause predicate applied to all 26

The predicate was applied rule by rule, to each rule's OWN cases, never to family membership.
Measured outcome — **12 `error` / 14 `warning`**; the count did not move, and that is a result, not
an assumption: every promotion was re-derived rather than inherited.

| rule (security family) | clause 1 | clause 2 — benign same-shape `valid` | severity |
|---|---|---|---|
| `sec-hardcoded-credential` (go/python/javascript/typescript) | yes | **strengthened this milestone** — the sole `valid` case was `token := os.Getenv(...)`, a *different* shape (call, not literal). Added `bearer := "not-a-secret-placeholder"`: an assignment whose RHS IS a string literal, benign, zero findings | `error` |
| `sec-weak-hash-md5` | yes | `sha256.New()` vs `md5.New()` — same shape (hash constructor call) | `error` |
| `sec-command-injection-shell` (go) | yes | `exec.Command("git", "status", ...)` vs `exec.Command("sh","-c",...)` — same call, benign argument list | `error` |
| `sec-command-injection-shell` (python) | yes | shares the id's case document; benign argument-list form | `error` |
| `sec-command-injection-exec` (javascript/typescript) | yes | `child_process.execFile(...)` vs `child_process.exec(...)` — same module, non-shell sibling | `error` |
| `sec-hardcoded-api-key` | yes | `const usageNote = "..."` vs `const openAiKey = "sk-live-alpha"` — same `const_spec` shape, benign literal | `error` |
| `sec-hardcoded-jwt-signing-key` | yes | `claims.SignedString(signingKey)` vs `claims.SignedString([]byte("..."))` — same call, non-literal key | `error` |
| `sec-template-injection-html` | yes | `template.HTMLEscapeString(...)` vs `template.HTML(...)` — adjacent symbol, and the anchored regex `^template\.HTML\(` requires the literal `(`, so the escaper does not match | `error` |
| `sec-csrf-no-token-check` | yes | **FAILS** — the pattern is the shape of every Go HTTP handler; a CSRF-protected handler is the benign same-shape construct and still matches. No satisfiable negative exists | `warning` |
| `sec-log-injection-unsanitized` | yes | **FAILS** — evaluated against the predicate rather than assumed. The pattern is `log.Printf($FORMAT, $$$ARGS)`, which matches every `log.Printf` carrying arguments; a sanitized call is the benign same-shape construct and still matches. The shipped `valid` case (`slog.Info(...)`) is a *different* function, so it discharges nothing | `warning` |

The twelve `go-*` rules fail clause 1 (not a security family) and all sit at `warning` — asserted
by `TestErrorSeverityFollowsTheTwoClausePredicate`, which fails any `error` outside a security
family.

### Item 2 — the two known cases

`sec-csrf-no-token-check` stays `warning` with its precision limitation recorded as the F6/go
matrix annotation. `sec-log-injection-unsanitized` was evaluated, not assumed: it fails clause 2
for the same structural reason, so it also stays `warning` and gains an F7/go annotation. Both are
now mechanically pinned by `TestShapeMatchersStayWarningWithRecordedLimitation`, which asserts BOTH
the severity and the presence of the matrix annotation.

### Item 3 — `metadata.cwe`

Measured: **14 of 14 security rules already carried `cwe`; 0 lacking.** Each `invalid` case was
read against its declared weakness class and instantiates it in idiomatic code: CWE-798 by a
literal credential/API-key assignment, CWE-327 by the MD5 constructor, CWE-78 by a shell-routed
execution call, CWE-321 by a literal signing key, CWE-79 by the unescaped-HTML conversion, CWE-117
by a raw external value in a log format, CWE-352 by a state-changing handler with no token check.

### Item 4 — the citation-or-probe anchor (RED at 0 of 14)

RED, verbatim (14 identical failures, one shown):

```
--- FAIL: TestSecurityRulesCarryCweAndAnchor
    sec-hardcoded-credential (go): metadata.anchor is ""; want a [cite: probe:] prefixed
    entry naming the reference documenting its matched head symbol, or a recorded probe
    ... (14 rules, one line each — the plan's "measured today at 0 of 14", re-measured here)
```

Every anchor was **measured before it was written**. Head-symbol probes actually run (rc=0 each):

```
$ printf 'h := md5.New()\n'                   | sg run -p 'md5.New()' -l go --stdin
STDIN:1:h := md5.New()
$ printf 'exec.Command("sh", "-c", userCmd)\n' | sg run -p 'exec.Command("sh", "-c", $CMD)' -l go --stdin
STDIN:1:exec.Command("sh", "-c", userCmd)
$ printf 'os.system(cmd)\n'                    | sg run -p 'os.system($CMD)' -l python --stdin
STDIN:1:os.system(cmd)
$ printf 'subprocess.run(cmd, shell=True)\n'   | sg run -p 'subprocess.run($CMD, shell=True)' -l python --stdin
STDIN:1:subprocess.run(cmd, shell=True)
$ printf 'child_process.exec(script);\n'       | sg run -p 'child_process.exec($CMD)' -l javascript --stdin
STDIN:1:child_process.exec(script);
   (typescript: identical output)
$ printf 'log.Printf("user %s", name)\n'      | sg run -p 'log.Printf($FORMAT, $$$ARGS)' -l go --stdin
STDIN:1:log.Printf("user %s", name)
```

The five `kind:`+regex rules cannot be probed through `sg run -p`, so they were probed through
their own rule documents against real-shaped files:

```
$ sg scan --config sgconfig.yml cred.go
error[sec-hardcoded-credential]: A credential appears to be hardcoded in source. ...   (x2)
error[sec-hardcoded-api-key]:   API key hardcoded in source. ...
$ sg scan --config sgconfig.yml jwt.go
error[sec-hardcoded-jwt-signing-key]: JWT signing key hardcoded as a literal byte slice. ...
$ sg scan --config sgconfig.yml tpl.go
error[sec-template-injection-html]: Marking user input as trusted HTML via html/template.HTML() enables XSS.
$ sg scan --config sgconfig.yml csrf.go
warning[sec-csrf-no-token-check]: POST handler without CSRF token verification. ...
$ sg scan ... cred.py / cred.js / cred.ts
error[sec-hardcoded-credential]: ...   (each language variant fires independently)
```

Anchors then written into `metadata.anchor`, alongside `metadata.cwe` as the plan requires:
`probe:` where the matched head symbol is a **vendor token prefix** rather than a documented API
symbol (the four credential variants and the API-key rule — five probes, each carrying its
invocation and observed output); `cite:` for the remaining nine, whose head symbols are documented
standard-library or framework calls (`crypto/md5.New`, `os/exec.Command`, `os.system` /
`subprocess shell=True`, `child_process.exec` ×2, `golang-jwt Token.SignedString`,
`html/template.HTML`, `log.Printf`, `net/http.HandlerFunc`).

### Item 5 — the handed record corrected

`sec-template-injection-html` **measures `error`** in the shipped tree; the handed measurement
record listed it as `warning`. Pinned by `TestTemplateInjectionRuleMeasuresError` so the stale
value cannot be re-derived from memory.

### Mutant guards — all three new assertions proven live

```
# Mutant D: metadata.anchor stripped from sec-csrf-no-token-check
--- FAIL: TestSecurityRulesCarryCweAndAnchor
    sec-csrf-no-token-check (go): metadata.anchor is ""; want a [cite: probe:] prefixed entry ...

# Mutant E: sec-csrf-no-token-check promoted to error
--- FAIL: TestShapeMatchersStayWarningWithRecordedLimitation
    sec-csrf-no-token-check: severity "error"; a shape matcher must stay warning (REQ-A16-014)
--- FAIL: TestSeveritySplitAcrossAllTwentySix
    severity split = 13 error / 13 warning, want 12 / 14
    1 security rules sit at warning, want 2 (both shape matchers)

# Mutant F: F6/go precision-limitation annotation stripped from the matrix
--- FAIL: TestShapeMatchersStayWarningWithRecordedLimitation
    sec-csrf-no-token-check cell (F6/go) evidence = "rule-tests case pair"; REQ-A16-015
    requires a recorded precision-limitation annotation ...
```

All reverted; `cmp` reports both mutated files byte-identical to their pre-mutant copies;
post-revert package run `ok`.

### M3 AC matrix

| AC | claim | evidence | status |
|---|---|---|---|
| AC-A16-015 | every `error` is backed by a benign-shape negative case | `TestErrorSeverityFollowsTheTwoClausePredicate` PASS over all 26; `sg test` 21/21 proves the zero-findings half; credential family strengthened to a genuine same-shape negative | PASS |
| AC-A16-016 | every security rule carries `metadata.cwe` AND a cite/probe anchor for its matched head symbol | `TestSecurityRulesCarryCweAndAnchor` PASS 14/14, RED→GREEN from 0/14; every anchor measured before written; mutant D live | PASS |
| AC-A16-017 | shape matcher stays `warning`, demotion visible in the matrix | `TestShapeMatchersStayWarningWithRecordedLimitation` PASS; mutants E and F live | PASS |
| AC-A16-018 | severity across all 26 follows from evidence, not assertion | per-rule predicate table above + `TestSeveritySplitAcrossAllTwentySix` PASS (12/14, 2 security at warning) | PASS |
| AC-A16-019 | neutrality (partial — close-phase criterion) | ruleset + rule-test trees scanned for SPEC-ID / date / SHA / macOS path / CLAUDE.local: the only hit is a pre-existing redacted Slack-token placeholder (`xoxb-000000000000-...`) already present at HEAD, matched by the hex-shaped SHA pattern, not a commit SHA. `make build` leaves `internal/template/catalog.yaml` **unchanged** | PARTIAL — close-phase owns the remaining clauses |

### M3 gates

```
$ sg test --config .../sgconfig.yml                             rc=0   21 passed; 0 failed
$ go test -count=1 ./internal/astgrep/...                       rc=0   ok
$ go test -count=1 ./internal/template/...                      rc=0   ok
$ go vet ./internal/astgrep/...                                 rc=0
$ GOOS=windows GOARCH=amd64 go vet ./internal/astgrep/...       rc=0
$ gofmt -l internal/astgrep/                                    (empty)
$ golangci-lint run --timeout=3m ./internal/astgrep/...         rc=0   0 issues.
$ make build                                                    rc=0; catalog.yaml diff EMPTY
$ grep -rn 'AskUserQuestion|mcp__askuser' <the four created files>   rc=1 (no hits)
```

new_warnings_or_lints_introduced: 0.

### Pre-existing failure surfaced, NOT introduced here — for the lead

`go test ./internal/hook/...` fails on three subtests:

```
--- FAIL: TestScanWriteContentNoConfigNoTempFile/control:_resolvable_config_creates_exactly_one_temp_file
    pre_tool_scan_config_test.go:127: expected 1 ScanFile call for the control, got 0
--- FAIL: TestScanWriteContentUncoveredLanguage/control_sample.go
    pre_tool_scan_config_test.go:239: expected 1 ScanFile call for the .go control, got 0
--- FAIL: TestScanWriteContentCoveredLanguageFollowsConfig/shipped_ruleset_scans_go
    pre_tool_scan_config_test.go:253: expected 1 ScanFile call, got 0
```

Attributed by bisection rather than assumed: the five edited ruleset files were restored to their
HEAD (`49375fd77`) blobs, `make build` re-run, and the same three subtests re-run — **they fail
identically with M3's template edits absent**. The edits were then restored and `make build` re-run
(catalog diff empty, 14 anchors present, `sg test` 21/21). The failure is therefore pre-existing on
this branch and outside M3's scope; M3 changes only `metadata` fields, which do not affect matching.
Not repaired here — repairing it would be scope expansion into a subsystem this SPEC declares
PRESERVE. The corpus pinning suite (`internal/hook/security`, `TestAstgrepCorpus*`) is green.

Gaps (M3): the full repository suite was NOT run (standing local-load prohibition; CI owns the
cross-package verdict). Clause-2 "same shape" is a structural judgement recorded per rule in the
table above, not a mechanical measurement — the mechanical half is the zero-findings assertion
`sg test` already makes. The nine `cite:` anchors name references rather than executing them, which
is exactly the evidence class the requirement permits; the five `probe:` anchors carry executed
output. Residual risk: an anchor can go stale if an upstream library renames the symbol it cites,
and nothing re-checks a citation the way a probe re-checks itself.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: "(this commit)"   # placeholder backfill per D3 exemption; M-final backfills
run_status: M1-M3 COMPLETE (contract half closed; close-phase pending)
milestones_done: [M1, M2, M3]
m1_stop_condition: harness green (21/21) from repo-side root; both mutants observed failing
  ([Missing] rc=4, [Noisy] rc=4); id-keying answer recorded (design.md §3.3); corpus pinning
  test present and green incl. RED probes on all three instruments
new_warnings_or_lints_introduced: 0
cross_platform_build.windows_vet: PASS
cross_platform_build.local_build: PASS
total_run_phase_files_m1: 48 (c1) + 1 (pinning test)
total_run_phase_files_m2: 4 (coverage-matrix.md, coverage_matrix.go, coverage_matrix_test.go,
  progress.md)
total_run_phase_files_m3: 9 (5 security rule YAML, 1 rule-test case doc, rule_severity.go,
  rule_severity_test.go, coverage-matrix.md; progress.md)
m2_stop_condition: checker fails on all four synthetic defects with four distinct class
  names -- proven through the DOCUMENT path, not only against hand-built structs -- and
  passes on the seeded 14; mutant on the real matrix observed failing and reverted
m3_stop_condition: all 26 severities re-derived from each rule's own cases (12 error /
  14 warning); 14/14 security rules carry metadata.cwe AND a cite:/probe: anchor measured
  before it was written (RED at 0/14); both warning-classified security rules carry a
  recorded precision-limitation annotation; three mutants observed failing and reverted
preexisting_failure_not_introduced: internal/hook pre_tool_scan_config_test.go 3 subtests
  ("expected 1 ScanFile call, got 0") -- attributed by bisection against HEAD blobs, fails
  identically with M3 template edits absent; outside scope, not repaired
m1_to_mN_commit_strategy: separate conventional commits per milestone; catalog diff asserted
  empty on every template-touching commit (AC-A16-019 posture preserved pre-close)
l44_pre_commit_fetch: n/a (worktree-only branch; origin copy intentionally stale per B9)
l44_post_push_fetch: n/a (nothing pushed)

```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

```yaml
tier: L
scope_files: ~8 (template rules YAML, sgconfig.yml, rule-tests tree, corpus pin test, coverage-matrix.md, checker Go test)
domains: 3 (ast-grep YAML configs, Go test code, markdown contract docs)
file_language_mix: yaml/go/markdown
concurrency_benefit: LOW (M1 harness -> M2 matrix -> M3 severity are strictly sequential)
agent_team_prereqs: not requested
```

| Mode    | Selected      | Rationale                                                              |
|---------|---------------|------------------------------------------------------------------------|
| direct  | not selected  | Multi-milestone Tier L, no trivial path                                 |
| serial  | **selected**  | Coding-heavy, sequential dependency chain, one milestone per spawn      |
| fanout  | not selected  | No independent research fan-out; coding caveat binds                    |
| sweep   | not selected  | Not a uniform mechanical transform; worktree-editing                    |

Decision: serial

Justification: Anthropic's coding-task parallelism caveat applies — the three milestones form a
strict dependency chain (harness before matrix before reclassification), so concurrency buys only
cache pressure and reconciliation cost. Each milestone delegates once to manager-develop with the
Tier L Section A-E template. Boundary case: none — all thresholds resolve unambiguously toward
serial.

Kickoff approval: operator approved "승인 — M1~M3 연속 실행" this session, AFTER the lead dispatch
and prior to any run-phase implementation spawn.
