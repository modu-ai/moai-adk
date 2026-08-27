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

```
$ git diff --stat 294b4b6ab -- internal/hook/ ':(exclude)internal/hook/astgrep_corpus_pin_test.go'
(no output)                     # every pre-existing hook file byte-unchanged from pinned base
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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: "(this commit)"   # placeholder backfill per D3 exemption; M-final backfills
run_status: M1 COMPLETE (M2/M3 pending)
milestones_done: [M1]
m1_stop_condition: harness green (21/21) from repo-side root; both mutants observed failing
  ([Missing] rc=4, [Noisy] rc=4); id-keying answer recorded (design.md §3.3); corpus pinning
  test present and green incl. RED probes on all three instruments
new_warnings_or_lints_introduced: 0
cross_platform_build.windows_vet: PASS
cross_platform_build.local_build: PASS
total_run_phase_files_m1: 48 (c1) + 1 (pinning test)
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
