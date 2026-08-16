# t49 Evidence — ci.yml go_code paths-filter covers non-Go test inputs

Card: ci.yml `go_code` paths-filter가 테스트의 실제 입력 파일을 커버하지 않는다
(#1557 스텁-초록불 근본 원인). Branch `WT-t49`, based on `origin/release/v3.1.1`
(09516ff0c, fast-forward merge — clean).

## (a) Chosen candidates + rationale + rejected alternatives

**Chosen: candidate (1) + candidate (3)** — the card's recommended shape.

- **(1) go_code 필터에 비-Go 테스트 입력 루트 추가.** The filter now contains,
  beyond the original 6 patterns, 7 evidenced non-Go input roots. Every added
  root cites the Go test that reads it from the REAL repo tree (repo-root
  resolver walking up to `go.mod` — not `t.TempDir()` scaffolds, which were
  explicitly separated during the sweep):

  | Added pattern | Reading Go test (fail mode) |
  |---|---|
  | `.moai/**` | `internal/astgrep/gar_rule_test.go:66` — `realRuleFile` os.Stat + ReadFile of `<repo-root>/.moai/astgrep-rules/go/error-handling.yml`, `t.Fatalf` when absent (**the #1557 incident itself: the path moved, 3 tests FAILed**) |
  | `internal/template/templates/**` | `internal/spec/catalog_hash_test.go:113` (fail-loud REQ-CHR-007) + `internal/template/sanitized_pair_parity_test.go:154`, `branch_protection_parity_test.go:45`, and ~20 more `internal/template/*_test.go` parity guards |
  | `internal/template/catalog.yaml` | `internal/spec/catalog_hash_test.go:112` (fail-loud) |
  | `.claude/agents/**` | `internal/template/agent_askuser_audit_test.go:30` — walks ALL of `.claude/agents/**/*.md` (incl. `harness/`); `manager_kanban_depth_test.go:44` (depth-2 Agent-carrier seal), `haiku_effort_guard_test.go:81` |
  | `.claude/rules/moai/**` | `internal/config/token_budget_guard.go:113` — `alwaysLoadedSurface` walks `.claude/rules/moai/**/*.md` (always-loaded budget tripwire, headroom 1.65% per prior measurement) |
  | `.claude/output-styles/moai/**` | `internal/config/token_budget_guard.go:130` (fixed surface slot) |
  | `CLAUDE.md` | `internal/config/token_budget_guard.go:129` (fixed surface slot; `token_budget_guard_test.go:58` gates on its presence) |

- **(3) 스텁 경로 skip-reason 관측성.** The `test-skip-marker` stub step now
  writes the skip reason to `$GITHUB_STEP_SUMMARY` (renders on the run Summary
  page), stating that NO Go test ran and pointing at the filter definition by
  **pointer** (job `detect`, step `f`) — never duplicating the pattern list, so
  the two surfaces cannot drift. Job names were NOT touched (guardrail).

**Rejected alternatives:**

- **Candidate (2) — fixture-path SSOT shared by filter and tests.** Rejected for
  this card: it requires either generating the YAML filter from Go (build step
  in detect job) or generating a Go fixture manifest from YAML consumed by
  tests — cross-language coupling and new moving parts for a Tier S~M card.
  The pointer-style summary from (3) plus per-root evidence comments in the
  filter keep the drift risk visible without the machinery. A future card can
  promote the comment table into a generated SSOT if drift materializes.
- **Blanket `.claude/**` or `**` filter.** `.claude/skills/**`,
  `.claude/commands/**`, `.claude/hooks/**`, `.claude/rules/local/**` have NO
  evidenced Go-test reader (verified by sweep — repo-root `.claude` readers are
  agents/rules-moai/output-styles only; skills/commands readers all operate on
  `t.TempDir()` scaffolds or the templates mirror). Adding unevidenced roots
  would run the Go suite on changes no test observes. (`.claude/agents/**` IS
  taken whole because the askuser audit walks the entire tree, harness/ included.)

**Glob-engine evidence (load-bearing for the pattern syntax):** the pinned
`dorny/paths-filter@ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d (v4.0.3)` bundle
(fetched via `gh api repos/dorny/paths-filter/git/blobs/6ed114c…`, 1,342,139 B,
saved `/tmp/paths-filter-index.js`) contains at its head:

```js
// Minimatch options used in all matchers
const MatchOptions = {
    dot: true
};
```

so `**` traverses dot-directories — `internal/template/templates/**` matches
`templates/.claude/**` and `templates/.moai/**`, and `.moai/**` matches the
dotted root's children. No per-dot-dir pattern duplication needed.

## (b) Diff summary

Single file changed: `.github/workflows/ci.yml` (+ this evidence file).

- `detect` job: 26-line evidence comment above the `dorny/paths-filter` step +
  7 new patterns in the `go_code` filter (6 → 13 patterns).
- `test-skip-marker` job: step renamed `Skip (no Go changes)` →
  `Skip (no Go-code inputs changed)` (STEP name only — the JOB name
  `Test (${{ matrix.os }})` is untouched on both the real and stub jobs),
  `shell: bash` added, and the run block now also writes the skip reason to
  `$GITHUB_STEP_SUMMARY`.

## (c) PR #1557 observed file list + new-filter match (verbatim)

Observed via `gh pr view 1557 --repo modu-ai/moai-adk --json files --jq '.files[].path'`
(2026-08-17) — 17 files, **zero `.go`**:

```
.claude/rules/local/ci-autofix-protocol.md
.claude/rules/local/ci-watch-protocol.md
.claude/rules/local/lifecycle-sync-gate.md
.claude/rules/local/repo-local-pr-policy.md
.claude/skills/hns-workflow-ci-loop/SKILL.md
.moai/astgrep-rules/go/concurrency.yml
.moai/astgrep-rules/go/error-handling.yml
.moai/astgrep-rules/go/hardcoding.yml
.moai/astgrep-rules/go/idioms.yml
.moai/astgrep-rules/go/resource-safety.yml
.moai/astgrep-rules/security/credentials.yml
.moai/astgrep-rules/security/crypto.yml
.moai/astgrep-rules/security/injection.yml
.moai/astgrep-rules/security/secrets.yml
.moai/astgrep-rules/security/web.yml
.moai/astgrep-rules/sgconfig.yml
.moai/config/sections/gate.yaml
```

Replay (matcher implements picomatch semantics for the three pattern shapes
present — literal, `X/**` prefix with dot:true, `**/*.go` suffix — with sanity
asserts against degenerate matching):

- **OLD filter: go_code=false (0/17 matched)** → stub job satisfied required
  check `Test (ubuntu-latest)`. Matches the recorded history (steps=3 stub on
  #1557/#1558/#1559/#1562 while main was broken).
- **NEW filter: go_code=true (12/17 matched)** — all 12 via `.moai/**`:
  the 11 `.moai/astgrep-rules/**` files + `.moai/config/sections/gate.yaml`.
  The real `test` job would have run and surfaced the 3 `internal/astgrep`
  FAILs.
- Matched by neither old nor new filter (correctly, no evidenced reader):
  `.claude/rules/local/*` (4) + `.claude/skills/hns-workflow-ci-loop/SKILL.md`.

The PR did NOT touch Go files — confirmed from the list above, so the honest
statement is: the ONLY input class the old filter missed here was `.moai/**`,
and that alone was sufficient to hide the breakage.

## (d) Verification outputs (verbatim)

1. `actionlint .github/workflows/ci.yml` → exit 0, no output (run twice: after
   first edit and after final edit).
2. Double YAML parse (outer workflow + inner `filters: |` block as jsyaml will
   parse it):

```
outer YAML: OK
inner filters YAML: OK
go_code patterns: 13
job names: real='Test (${{ matrix.os }})' stub='Test (${{ matrix.os }})' (unchanged)
```

3. Filter replay (final filter) — script output:

```
OLD go_code=false (0/17 matched)
NEW go_code=true (12/17 matched)
unmatched by NEW: ['.claude/rules/local/ci-autofix-protocol.md', '.claude/rules/local/ci-watch-protocol.md', '.claude/rules/local/lifecycle-sync-gate.md', '.claude/rules/local/repo-local-pr-policy.md', '.claude/skills/hns-workflow-ci-loop/SKILL.md']
sanity asserts: PASS
```

4. Candidate (3) dry run — the exact `run:` block was extracted from the
   committed YAML (754 bytes → `/tmp/t49-stub-step.sh`) and executed with
   `GITHUB_STEP_SUMMARY=/tmp/t49-summary.md bash /tmp/t49-step.sh`:

```
No Go-code changes per paths-filter — Go test matrix skipped; required check satisfied.
step exit: 0
=== stdout above; summary file below ===
### Test (ubuntu-latest): skipped by paths-filter

The `detect` job classified this change as containing no `go_code`
inputs, so the Go test matrix did not run and this stub job
satisfies the required check.

- Filter definition: `.github/workflows/ci.yml` — job `detect`, step `f` (`go_code`)
- If this change moved or edited a file the Go tests read (see the
  non-Go input roots listed in that filter) and the suite should
  have run, add the path to the filter — a skip here means NO Go
  test ran on this change.
```

No `go test` was run (local full suites are forbidden; CI owns the verdict).
The astgrep path-resolution link is established by reading
`gar_rule_test.go`'s fail-loud `realRuleFile` (the targeted-test option was not
needed: the read path and its `t.Fatalf` are the demonstration).

## (e) Residual risks — what the filter STILL does not cover

1. **`.claude/rules/local/**`, `.claude/skills/**`, `.claude/commands/**`,
   `.claude/hooks/**` (repo root)** — no evidenced Go-test reader today. If a
   future test starts reading one of these (e.g. a hook-wrapper parity test on
   the repo's own `.claude/hooks/moai/*.sh` rather than the templates mirror),
   the filter must grow the root or the same stub-green hazard re-opens there.
2. **docs-site/**, root `*.md` docs, `install.sh`/`install.bat`** — not read by
   the Go suite per the sweep; unverified for FUTURE tests.
3. **Other workflows** (`.github/workflows/*` beyond ci.yml/codeql.yml) — the
   filter gates only what the `test`/`test-race` jobs consume; a workflow file
   that itself shells into Go tooling is not this card's scope.
4. **`e2e/**` and root scripts** — explicitly outside `go_code` by design
   (documented in the stub job's gating comment); unchanged.
5. **Filter ⊇ reader, not ≡**: the mapping is maintained by evidence comments,
   not generated from code (rejected candidate (2)). A new Go test that reads a
   NEW repo path will NOT auto-add that path to the filter — the residual
   process risk is "author forgets to update the filter". Mitigation today is
   the comment table at the filter + the stub summary pointing reviewers at it.
6. **Replay is an approximation** of picomatch for the three pattern shapes
   used (documented in (c)); exotic future patterns (negations, brace
   expansion) would need the real engine to replay.
7. **Cost side-effect**: PRs touching `.moai/**` (incl. docs/reports/specs),
   the template mirror, or `.claude/agents|rules-moai|output-styles` now run
   the ~2-3 min Go suite + race + integration instead of the 3-step stub.
   Accepted by the card (candidate (1) names `.moai/**` whole-root).
