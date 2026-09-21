# SPEC-MATRIX-GROUP-COMMENT-001 — Progress

Card: **t1055**. Tree: `.claude/worktrees/t1055`, branch `WT-retention-comment`,
plan-phase HEAD `d8304b49a`.

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **S** (comment text at two sites in one file; one milestone).
- Artifacts emitted: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
  `acceptance.md` is emitted despite Tier S permitting inline ACs, because the
  plan-phase brief named it as a deliverable.
- Requirements: 4 (REQ-MGC-001..004). Acceptance criteria: 6 (AC-MGC-001..006).
- SPEC ID regex check executed as Bash — `PASS`.
- Every coordinate in `spec.md` §A/§B re-measured in this tree at `d8304b49a`;
  the consumer count rests on `grep -rn --include='*.go' 'template\.AgentGroup' .`
  with the `agentfm_ordering_test.go` lookalikes explicitly excluded.
- Status: `draft`. Awaiting Implementation Kickoff Approval before run phase.

_Gaps at plan phase:_ the replacement comment wording in `plan.md` §C is a
proposal, not a measured artifact — the run phase finalises it against
AC-MGC-001. The `:8` profile-axis fix is unresolved pending an operator decision
(`spec.md` §E).

## §E.2 Run-phase Evidence

### Baseline attribution

Every command below was run **in this run, against this tree**, from the worktree
root `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1055`.

| Coordinate | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1055` (`git rev-parse --show-toplevel`) |
| Branch | `WT-retention-comment` (`git branch --show-current`) |
| `<BASE>` (pre-edit tree) | `d8304b49a` (`git rev-parse --short HEAD` before the M1 commit) |
| Working-tree state at measurement | ` M internal/template/profile_matrix.go` + `?? .moai/specs/SPEC-MATRIX-GROUP-COMMENT-001/` |

The lane's own figures were **re-run independently** rather than copied. Every
number below is this session's observation. Disagreements are recorded under
_Findings_; none of the lane's figures disagreed.

### AC-by-AC result matrix

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-MGC-001 | **PASS** | `BLOCK=$(awk '/^\/\/ Lookup is by agent NAME/,/^func ResolveAgentModelEffort/' internal/template/profile_matrix.go)` then `grep -qF` per needle | `internal/cli/model.go -> PRESENT` / `internal/web/agentfm.go -> PRESENT` / `gate -> PRESENT` / `display -> PRESENT` |
| AC-MGC-002 | **PASS** | `grep -c 'only as a display classification' internal/template/profile_matrix.go` | `0` |
| AC-MGC-003 | **PASS** | `grep -n 'display-only' internal/template/profile_matrix.go; echo "exit=$?"` | (no line printed) `exit=1` |
| AC-MGC-004(a) | **PASS** | `git diff -U0 d8304b49a -- internal/template/profile_matrix.go \| grep -E '^[+-]' \| grep -vE '^(\+\+\+\|---)' \| grep -vE '^[+-][[:space:]]*//'` | (no lines) `residual_exit=1`, `residual_count=0` |
| AC-MGC-004(b) | **PASS** | `git diff --name-only d8304b49a` | `internal/template/profile_matrix.go` (exactly one file; the SPEC artifacts were still untracked at measurement time — see Findings F1) |
| AC-MGC-005 | **PASS** | `diff <(git show d8304b49a:…\|sed -n 8p) <(sed -n 8p …)` then `grep -c 'max/medium/low'` | `IDENTICAL` / `1`. Line 8 verbatim: `// profile axis (max/medium/low) consumed via runtime-arg spawn injection rather` |
| AC-MGC-006 | **PASS** | `go build ./internal/template/...` ; `go test ./internal/template/... ./internal/cli/... ./internal/web/...` | `build_exit=0`. Test: see AC-MGC-006 detail below — plain run hit the documented `internal/cli` timeout with **zero** `--- FAIL`; re-run with `-timeout 30m` exits 0. |

### Negative controls (AC-MGC-002, AC-MGC-003)

Both are absence assertions, so each probe was fired against `<BASE>` to prove it
measures something. Both produced **non-empty** pre-edit output:

```
$ git show d8304b49a:internal/template/profile_matrix.go | grep -c 'only as a display classification'
1

$ git show d8304b49a:internal/template/profile_matrix.go | grep -n 'display-only'
261:// key: retained agent NAME (not a group — the group layer is display-only now,
exit=0
```

A third, independent positive control backs AC-MGC-004(a): the same diff carries
**29** changed lines in total (`grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | wc -l`
→ `29`), so the filter returning `0` residual non-comment lines is a real
measurement of a non-empty set, not a vacuous pass over an empty diff.

### AC-MGC-006 detail — the documented `internal/cli` timeout

The plain-form run reproduced the caveat recorded in `acceptance.md` §D verbatim:

```
$ go test ./internal/template/... ./internal/cli/... ./internal/web/...
test_exit=1
--- FAIL count: 0
ok    github.com/modu-ai/moai-adk/internal/template        60.741s
ok    github.com/modu-ai/moai-adk/internal/template/agentemit   0.529s
ok    github.com/modu-ai/moai-adk/internal/template/commandemit 0.332s
FAIL  github.com/modu-ai/moai-adk/internal/cli             601.215s
ok    github.com/modu-ai/moai-adk/internal/web             26.532s
(all 17 internal/cli/* sub-packages: ok)
line 976: panic: test timed out after 10m0s
```

The non-zero exit is the 10-minute package timeout on `internal/cli`, not a
regression: the `--- FAIL` line count is **0**, and the only failure signal is the
`test timed out after 10m0s` panic. Re-run per the caveat:

```
$ go test ./internal/cli/ -count=1 -timeout 30m
cli_30m_exit=0
--- FAIL count: 0
ok    github.com/modu-ai/moai-adk/internal/cli             1007.628s
```

The 1007.628s figure (≈16.8 min) is itself the discriminator: the package genuinely
needs more than the 10-minute default, so the earlier `FAIL` was the budget
expiring on a still-progressing run, not a hang and not a failing assertion.

This change touches no code, so a failure in `internal/cli` would not be
attributable to this card in any case. Scoped package run for the owning package:

```
$ go test ./internal/template/ -count=1
ok    github.com/modu-ai/moai-adk/internal/template        58.795s
```

### Supplementary verifications (beyond the ACs)

| Check | Command | Observed |
|---|---|---|
| Formatting | `gofmt -l internal/template/profile_matrix.go` | (empty) |
| Vet | `go vet ./internal/template/` | `vet_exit=0` |
| Consumer packages build | `go build ./internal/cli/ ./internal/web/ ./internal/template/` | `build_exit=0` |

### Consumer census (re-run independently)

```
$ grep -rn --include='*.go' 'AgentGroup' .
internal/web/agentfm.go:491:      if _, ok := template.AgentGroup(a.Name); !ok {
internal/web/agentfm_ordering_test.go:31:   // TestAgentGroupRankCoversEveryCatalogClass …
internal/web/agentfm_ordering_test.go:35:   func TestAgentGroupRankCoversEveryCatalogClass(t *testing.T) {
internal/web/agentfm_ordering_test.go:63:   // TestAgentGroupLabelMatchesRank …
internal/web/agentfm_ordering_test.go:67:   func TestAgentGroupLabelMatchesRank(t *testing.T) {
internal/template/profile_matrix.go:465,467,486  (self-referential: doc comment + declaration + the new comment)
internal/cli/model.go:101:       if g, ok := template.AgentGroup(agent); ok {

$ grep -rn --include='*.go' 'template\.AgentGroup' .
internal/web/agentfm.go:491
internal/cli/model.go:101
```

Exactly **two** production consumers. The four `agentfm_ordering_test.go` rows
(31, 35, 63, 67) carry no `template.` qualifier — they are name lookalikes about a
local rank/label helper and are **not** counted, per `acceptance.md` §E. Three
self-referential rows in `profile_matrix.go`. This matches the lane's census
exactly.

### Scope provenance — the card and the SPEC differ on purpose

Two comment sites were corrected, and they entered scope by **different routes**.
This distinction is recorded deliberately, not as incidental detail:

| Site | REQ | How it entered scope |
|---|---|---|
| `ResolveAgentModelEffort` doc block (pre-edit `:481-483`) | REQ-MGC-001 | Named directly by card **t1055**. |
| `defaultProfileMatrix` doc block (pre-edit `:261`) | REQ-MGC-002 | **Lane finding** — discovered during the run, carrying the same `display-only` claim. Inclusion was **explicitly approved by the lead**; it is not a unilateral scope expansion. |

The card text names only the first site. The SPEC's scope is wider than the card's
text by that approval, and the two are expected to differ.

### (a) Comments name callers by file + function, never by line number

The corrected comments point at both callers by **file path and function name**:
`internal/cli/model.go` → `resolveModelProfileReport`, and
`internal/web/agentfm.go` → `parseAgentFMForm`. No line number appears in either
comment.

This is the substantive departure from `plan.md` §C, and it is deliberate. A line
coordinate is tree-bound: the card's `481-483` happened to coincide between HEAD
`159dd30df` and `d8304b49a`, and baking a coordinate into a comment would re-plant
the exact defect class this card exists to remove — a comment asserting something
the tree can silently falsify. The lead approved the substitution.

**Both function names were confirmed to exist in THIS tree**, and their semantics
were read at the call site rather than assumed:

```
$ grep -n 'func resolveModelProfileReport' internal/cli/model.go
88:func resolveModelProfileReport(llm config.LLMConfig) modelProfileReport {

$ grep -n 'func parseAgentFMForm' internal/web/agentfm.go
450:func parseAgentFMForm(r *http.Request, agents []agentfm.AgentInfo, llm config.LLMConfig, targetTier string) (…)

$ sed -n '98,105p' internal/cli/model.go
    me, hasGroup := template.ResolveAgentModelEffort(llm, agent)
    group := "-"
    if g, ok := template.AgentGroup(agent); ok {
        group = g
    }                                   ← takes the STRING, defaults to "-", renders a column. Display.

$ sed -n '487,496p' internal/web/agentfm.go
    if _, ok := template.AgentGroup(a.Name); !ok {
        continue
    }                                   ← discards the string, uses only the BOOL to drop a
                                          non-matrix submission. A validation gate, not display.
```

The comment's characterisation of each caller matches the code read in this tree.

### (b) A retracted claim, and the regex trap that produced a wrong count

A first draft of the closing sentence asserted that removing the group-layer gate
would let non-matrix overrides **"start landing"**. That was an **unverified
premise** and was retracted before the comment was finalised. Reading
`internal/config/profile.go` shows `validateAgentOverrides` (line 182) rejects on
`retainedAgentNames` at validation time, so removing the gate changes the **failure
mode**, not the guardedness of the write. The shipped comment says exactly that,
and additionally states the fallback's limit: the two sets are separate literals in
separate packages with nothing keeping them in step.

Recording this retraction is not optional bookkeeping. This SPEC exists because a
comment asserted something the code did not do; the drafting process reproduced the
same failure one layer up, and suppressing that would make the record less honest
than the artifact it describes.

**The regex trap.** The two sets were compared to establish the fallback claim. An
earlier regex-based comparison reported **12 vs 13** — its character class omitted
digits and silently dropped `e2e-tester`. The comparison was redone without a
character class (extract quoted keys by field, sort, `diff`):

```
$ sed -n '/agentGroupMembership = map/,/^}/p' internal/template/profile_matrix.go \
    | grep '"' | cut -d'"' -f2 | sort > membership.txt
$ sed -n '/retainedAgentNames = map/,/^}/p' internal/config/profile.go \
    | grep '"' | cut -d'"' -f2 | sort > retained.txt
membership count: 13
retained count:   13
$ diff membership.txt retained.txt
(no output — SETS IDENTICAL)
$ grep -c 'e2e-tester' membership.txt ; grep -c 'e2e-tester' retained.txt
1
1
```

Both sets carry the same 13 names. The `e2e-tester` check is the targeted control
for the exact token the bad regex dropped — a bare `13 == 13` would not have caught
a compensating pair of errors.

### Findings

- **F1 — AC-MGC-004(b)'s expected output is measurement-time-dependent.**
  `acceptance.md` says the changed-file list should print the one Go file "plus the
  `.moai/specs/` artifacts of this SPEC". At the time of measurement the SPEC
  artifacts were **untracked**, so `git diff --name-only d8304b49a` printed the Go
  file alone. After the M1 commit (`15c70391d`) the same command prints all five
  paths — observed, not predicted:
  ```
  $ git diff --name-only d8304b49a          # post-commit
  .moai/specs/SPEC-MATRIX-GROUP-COMMENT-001/acceptance.md
  .moai/specs/SPEC-MATRIX-GROUP-COMMENT-001/plan.md
  .moai/specs/SPEC-MATRIX-GROUP-COMMENT-001/progress.md
  .moai/specs/SPEC-MATRIX-GROUP-COMMENT-001/spec.md
  internal/template/profile_matrix.go
  ```
  AC-MGC-004(a) was re-run against the committed tree and is unchanged
  (`residual_count=0` against `29` total changed lines). Both readings satisfy the
  criterion as written; the AC does not say which side of the commit it is
  evaluated on. Recorded, not fixed.
- **F2 — SPEC-quality: AC-MGC-002/003 test a string, not the claim.** Both absence
  criteria are literal-string checks (`grep -c 'display-only' == 0`). That form also
  fails on a sentence that **negates** the phrase: the lane's first wording was
  "It is not display-only in general" — a correct statement the AC would have marked
  FAIL. The lane reworded to avoid the token, so the criterion passes, but the
  criterion is measuring vocabulary rather than the assertion. A future correct
  negation would trip it. **Not fixed here** — SPEC body content is outside this
  phase's ownership (`spec-frontmatter-schema.md` § Forbidden ownership crossings).
- **F3 — cosmetic line-wrap artefact, deliberately not repaired.** The
  `defaultProfileMatrix` rewrite left a short line (`// {model, effort}. This` /
  `// is the authoritative fallback …`). `gofmt` does not rewrap comment prose, so
  it is format-clean and every AC passes. Left as-is: the operator scope decision is
  comment-correction only, and the file was declared not-to-be-re-edited in this
  phase.
- **F5 — the status transition reaches one artifact, not four.** The brief asked
  for `draft → in-progress` "across the SPEC artifacts". Measured in this tree,
  only `spec.md` carries YAML frontmatter at all (`grep -c '^status:'` → spec.md
  `1`, plan.md `0`, acceptance.md `0`, progress.md `0`). Per
  `spec-frontmatter-schema.md` § Artifact Statelessness, the sibling artifacts are
  **status-stateless by rule** — they MUST NOT carry a `status:` field, so adding
  one to satisfy the phrasing would introduce a schema violation. The transition
  was therefore applied to `spec.md` alone (`status: in-progress`, `updated:
  2026-09-21`). No body content in any artifact was modified.
  The plan-phase line in §E.1 reading "Status: `draft`" is left unchanged: it
  records what was true at plan phase, and observations are not rewritten to match
  later facts.
- **F4 — no figure disagreed with the lane.** Every lane-reported measurement
  (2 consumers, 3 self-referential rows, 4 lookalikes, residual count 0, `gofmt`
  empty, vet 0, `internal/template` ok ≈60s, consumer builds 0, negative controls
  firing) reproduced identically in this session.

### Gaps (explicitly not observed)

- **Cross-platform build not run.** `GOOS=windows GOARCH=amd64 go build ./...` was
  not executed. No AC requires it, and the change is comment text with a zero
  non-comment diff (AC-MGC-004(a)), so no platform-conditional code path can be
  affected. Unobserved, not inferred-passing.
- **`golangci-lint` not run.** No AC requires it; `gofmt` and `go vet` were run
  instead. Lint status against baseline is therefore unmeasured.
- **Coverage not measured.** No AC requires it; the change adds no executable code,
  so package coverage cannot move. Unmeasured, not claimed.
- **No RED evidence (§E8) exists, by construction.** The operator scope decision
  forbids new test code, so there is no failing-test-first artifact. The ACs are
  grep/diff/build checks and the negative controls above are their equivalent
  falsifiability evidence.
- **`§F Phase 4 Mode Selection` is absent from this file.** That section is the
  orchestrator's to write, not this phase's.
- **Full-suite verdict deferred to CI.** Per repo discipline, `go test ./...` was
  not run locally; only the affected and consumer packages were measured.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-21
run_commit_sha: 15c70391d          # M1 commit; backfilled in the following commit (a commit cannot cite its own hash)
run_status: audit-ready
ac_pass_count: 6
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 1     # internal/template/profile_matrix.go:8 — byte-identical (AC-MGC-005)
new_warnings_or_lints_introduced: 0 # gofmt empty, go vet exit 0; golangci-lint not run (Gap)
cross_platform_build:
  linux_amd64: not-run              # Gap — no AC requires it; zero non-comment diff
  windows_amd64: not-run            # Gap — same
  darwin_native: pass               # go build ./internal/cli/ ./internal/web/ ./internal/template/ → exit 0
total_run_phase_files: 5            # 1 Go file + 4 SPEC artifacts (artifacts ride the M1 commit)
m1_to_mN_commit_strategy: single-commit  # Tier S, one milestone (plan.md §D M1)
status_transition: draft -> in-progress  # spec.md only; siblings are status-stateless by schema
```

_Gaps at run phase:_ the four items listed under §E.2 Gaps. The `:8` profile-axis
observation remains unresolved and is carried forward unchanged for the operator
(`spec.md` §E) — AC-MGC-005 confirms the line was not touched.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
