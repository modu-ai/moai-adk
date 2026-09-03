# plan.md — SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001

## §A Context

- Worktree: `.claude/worktrees/t467`, branch `WT-branchguard-flagclass`,
  plan-phase base HEAD `d592b0551`.
- Defect locus: `internal/hook/branch_guard.go:120` —
  `\bgit\s+branch\s+(-[dDmMcC]\s+)?[^\s-]`. Single-char mutation class +
  "non-flag token" rule; `f`, `u`, long config-mutation flags, and combined
  short-flag clusters all fall outside it.
- Doctrine target: `.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  v1.3.2 § Rules (mutation table row 2: `git branch <name>` /
  `-d|-D|-m|-M|-c|-C` forbidden) + § Permitted (query list) + the
  Query-vs-mutate discrimination bullet. The v1.3.2 table does not yet name
  `-f`/`--force`/`-u` explicitly; the SPEC's position is that they are
  mutation forms under the table's "leaves a branch other sessions did not
  expect" rationale (a `-f` pointer rewrite is worse than a `-d` delete),
  and the M1 measurement — not the table's literal text — is the
  classification authority.
- Plan-phase ground truth (already measured on `d592b0551`; verbatim output
  preserved in acceptance.md §D.0): 10 mutation forms pass the matcher
  today; all probed query forms already pass (allowed).

### §A.5 PRESERVE list (byte-identical behavior)

- `branchStatePatterns` entries other than the `git branch` entry
  (`branch_guard.go:118-119`, `:121-138`): switch, checkout, reset --hard,
  stash, rebase, merge.
- `branchGuardViolationPrefix` sentinel, exemption logic
  (`MOAI_BRANCH_GUARD_EXEMPT` + `manager-git` identity), opt-in gate
  (`Workflow.BranchGuard.Enabled` call-site gating), primary-vs-worktree
  discriminant, audit-log path, fail-open contract.
- Existing tests: `branch_guard_test.go`, `branch_guard_quoted_test.go`,
  `branch_guard_worktree_test.go`, `branch_guard_pr1338_test.go`,
  `pre_tool_branch_guard_optin_test.go`,
  `pre_tool_branch_guard_integration_test.go` — all stay green unmodified
  (except where M3 extends them with new cases).

## §B Known Issues (filtered, Tier-relevant)

- **B4 Frontmatter schema** — spec.md uses canonical 12 fields; verified
  pre-write (ID regex PASS).
- **B5 CI 3-tier** — go vet / golangci-lint / go test run as separate CI
  lanes; measure lint baseline before editing.
- **B8 Working-tree hygiene** — this card's worktree also holds the SPEC
  artifacts; stage by explicit pathspec, never `git add -A`.
- **B10 Scope discipline** — touch only `branch_guard.go`, its test files,
  and this SPEC directory. No doctrine files, no CLAUDE.local.md.
- Regex hazards specific to this fix (see §E technical approach):
  `--force` vs `--format` prefix collision; `--contains`/`--merged`/
  `--no-merged` embedded-letter collision with any `[dDmMcCf]`-style class;
  `(?i)` case-insensitive compilation already applies.

## §C Pre-flight (run-phase, before M1)

```bash
git branch --show-current && git rev-parse --short HEAD   # expect WT-branchguard-flagclass
go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5             # lint baseline
go test ./internal/hook/ -run 'TestBranchGuard|TestPreTool_BranchGuard' -count=1 2>&1 | tail -3
```

## §D Constraints

- No commits from the plan phase (orchestrator commits after reading
  artifacts). Run-phase commits follow Conventional Commits with card id
  `t467` in the message per the gitflow lane protocol.
- No local full suite; affected packages only
  (`go test ./internal/hook/...`), then the integration-window CI verdict.
- Do NOT touch non-`branch` patterns even if a nearby refactor looks
  tempting (scope discipline).
- Escalation path: if M1's measured matrix contradicts the SPEC's expected
  matrix (§D.1 of acceptance.md) on any cell, return a blocker report — do
  not silently re-classify.

## §E Self-Verification (run-phase §E skeleton pointer)

Run-phase evidence lands in `progress.md` §E.2 per the attribution triple
(command + verbatim output + HEAD SHA). M1's RED output is the E8-class
evidence (TDD RED before GREEN).

## §F Milestones

### M1 — Synthetic measurement matrix (Priority: High; decides the fix class)

Create a permanent table-driven Go test (e.g.
`internal/hook/branch_guard_flagclass_test.go`) driving
`matchBranchStateCommand` with the full flag-form matrix of acceptance.md
§D.1. Expected values are the doctrine-based TARGET matrix (mutation →
deny/matched, query → allow/unmatched). Run it on the pre-fix tree:

- Expected observation: the 10 under-match cells FAIL (RED — the defect),
  all query cells and existing deny cells PASS (green-now).
- Record verbatim RED output + tree SHA in progress.md §E.2; this is the
  RED-now baseline that pins the fix class.
- Decide the fix class from the matrix (see §G): if any probed form behaves
  contrary to §D.1 expectations, stop and blocker-report.

### M2 — Matcher fix (Priority: High)

Widen/distinguish the `git branch` pattern (or replace that single pattern
entry with token-level discrimination) so the M1 matrix goes fully green:
mutation flags (incl. `-f`, `--force`, `-u` family, `-t` family,
`--edit-description`, combined clusters containing d/D/m/M/c/C/f) deny;
query forms stay allowed; whole-token long-flag classification (REQ-WBG-F-004).
Preserve fail-open, sentinel, and all §A.5 surfaces. RED→GREEN: M1 test
flips to green in this milestone.

### M3 — Test completion (Priority: High)

1. Denial cases end-to-end through `checkBranchState`/handler (not just the
   matcher): `-f`, `--force`, `-df`, `-fm`, `-vD`, `-u`, `--set-upstream-to`,
   `--unset-upstream`, `-t`, `--edit-description` → deny with
   `BRANCH_GUARD_VIOLATION:` prefix (extends the shape of
   `branch_guard_test.go:610-660`).
2. Query-allowlist regression cases: `--merged`, `--no-merged`,
   `--points-at`, `--format`, `--sort`, `--list develop -v` → allow, added
   alongside the existing pins at `branch_guard_test.go:308-320`.
3. Opt-in / exemption condition tests: keep
   `pre_tool_branch_guard_optin_test.go` green; add documented condition
   comments (or table cases) encoding REQ-WBG-F-008 — tests enable the flag
   explicitly; both exemption axes fire only for main-thread
   `claude --agent` launches / hook-process env, unreachable from
   tool-spawned subagents. These are documented conditions, NOT code
   changes.
4. Verify: `go test ./internal/hook/ -count=1` + `golangci-lint run`
   (affected scope only); report per the 5-section evidence format.

Milestones M1→M2→M3 are strictly sequential (M2 needs M1's matrix; M3 needs
M2's fix). No further milestones — Tier-S scope discipline.

## §G Technical approach (fix-class options, decided by M1)

Two candidate fix shapes; M1's matrix decides:

1. **Regex widening (minimal).** Keep a single regex entry; extend the flag
   alternation to cover long mutation flags and multi-char short clusters,
   e.g. shape: `\bgit\s+branch\s+(--force\b|--set-upstream(-to)?(=\S*)?\b|--unset-upstream\b|--edit-description\b|--track\b|--no-track\b|-[dDmMcCf]\S*|-u\S*\s+\S+|-[dDmMcC]\s+)?[^\s-]`
   — with the whole-token rule that any `--long` flag NOT in the mutation
   set is a query flag and must not match. Hazards: the `--format`/`--force`
   prefix collision requires full-token anchors (`\b` or `[\s=]`); embedded
   letters in `--contains`/`--merged`/`--no-merged` forbid any bare
   character-class scan of long flags.
2. **Token-level discriminator (clearer).** Replace the single `git branch`
   regex entry with a small matcher function for the `git branch` prefix:
   tokenize the command tail, classify each token (`--long` exact-match
   against the mutation set vs the query set; `-xyz` cluster → mutate iff it
   contains any of `dDmMcCf` or is `-u`; bare non-flag token → create/deny;
   unknown flag → treat as query, fail-open). This expresses REQ-WBG-F-004
   directly and is the recommended shape if the regex grows beyond
   readable form; it must slot into `branchStatePatterns`' matching surface
   without changing the other entries.

Either shape must keep the quoted-span collapse (see
`branch_guard_quoted_test.go`) and `(?i)` behavior intact for the `git
branch` entry.

## §H Anti-patterns

- Estimating a flag's classification from git docs without a matrix cell —
  forbidden (REQ-WBG-F-007).
- Substring long-flag matching (`--for` prefix classes) — would deny
  `--format` queries (REQ-WBG-F-004).
- Character-class scan over long-flag names — would deny `--contains`,
  `--merged`, `--no-merged` via embedded `c`/`m`.
- Tightening fail-open to catch obfuscated forms (`bash -c` wrappers) —
  direction is documented under-match, out of scope.
- Touching sibling patterns "while in the file" — §A.5 forbids.

## §I Cross-references

- spec.md §C REQ-WBG-F-001..008; acceptance.md §D.1 matrix + §D.0 RED-now
  baseline.
- `verification-completeness.md` §2 two-cell adoption discipline (RED-now
  cell = acceptance.md §D.0 probe output on `d592b0551`; green path = M2).
- gitflow lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) —
  local develop merge via integration window; no direct push.
