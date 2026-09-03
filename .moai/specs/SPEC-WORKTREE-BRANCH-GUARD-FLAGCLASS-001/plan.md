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
- Doctrine disposition (audit iteration 1, D1): the doctrine update is IN
  scope as a bounded M3 edit (REQ-WBG-F-009) — one bullet + one table-row
  enumeration, Template-First with sanitized-pair parity. Leaving v1.3.2
  asserting "`-f`/`-u` … pass" post-fix would recreate the doc-code
  shared-discrimination break the dispatch mandate named.
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
  this SPEC directory, and the bounded doctrine edit's two copies (template
  `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`
  + local mirror, same commit). No CLAUDE.local.md, no other doctrine or
  detail-companion edits.
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

- Expected observation: every pre-fix allow row in acceptance.md §D.1
  FAILs (RED — the defect). As of v0.6.0 that is 28 rows / 29 command
  variants (§D.0's 10 guard-level measurements + audit-measured git-form
  liveness M-20/M-22/M-23/M-24/M-25..M-29/M-31/M-33 + inferred-pending-M1
  M-07/M-14/M-17/M-18/M-21/M-30/M-32). All query cells (Q-01..Q-17) and
  existing deny cells PASS (green-now). §D.1 is the normative expectation
  authority — the M1 run's row set is read against §D.1, never against
  this count.
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
   changes. Per D7 (audit iteration 1): also add the small negative-path
   condition test — subagent-shaped `HookInput` (no `AgentType`, env
   unset) + mutating `git branch -f` → deny stands; labeled as pinning
   guard LOGIC only, not the payload shape. Per D12 (audit iteration 2):
   the `AgentType` axis is contested by `hooks-system.md:114` ("All hook
   events include `agent_id` and `agent_type` fields when triggered from a
   subagent context (v2.1.69+)") vs the guard's reading
   (`branch_guard.go:30-33`, t43 runtime observations) — M3 captures one
   real tool-spawned PreToolUse payload and records which holds; a capture
   contradicting the guard's reading becomes a doc-reconciliation blocker
   report, never a silent re-classification.
4. Bounded doctrine alignment (REQ-WBG-F-009 / AC-009): update the v1.3.2
   Query-vs-mutate bullet (`main-checkout-branch-guard.md:90-95`) and
   forbidden-table row 2's flag enumeration to name the extended mutation
   class (`-f`/`--force`, `-u`/`--set-upstream-to`/`--unset-upstream`,
   `--delete`/`--move`/`--copy`, `-t`/`--track`/`--no-track`,
   `--edit-description`, combined clusters) — Template-First (template
   copy first, local mirror in the same commit, sanitized-pair parity).
   Bounded to that bullet + row; nothing else in the rule changes.
5. Verify: `go test ./internal/hook/ -count=1` + `golangci-lint run`
   (affected scope only); report per the 5-section evidence format.

Milestones M1→M2→M3 are strictly sequential (M2 needs M1's matrix; M3 needs
M2's fix). No further milestones — Tier-S scope discipline.

## §G Technical approach (fix shape — corrected at audit iteration 1, D3)

Both original sketch shapes were defective as written: (1) the sketched
regex's trailing `[^\s-]` after the optional flag group cannot match any
long-mutation-flag-with-operand form (independent trace: `--force topic`,
`-f topic`, `--unset-upstream topic`, `--track topic …`, `-vD old` all
fail to match); (2) the token sketch's unqualified "bare non-flag token →
create/deny" would deny pinned query cells whose query flags take operands
(`--contains HEAD`, `--merged main`, `--list develop -v`, `--format
%(refname)`). The corrected discrimination rule the M2 fix implements:

**Deny iff either holds for the token stream after `git branch`** (after
quoted-span collapse, `(?i)`):

1. **Positional creation (extended at audit iteration 2, D9)**: a
   non-flag operand is a creation when NO list action is selected and the
   operand is not consumed as the value of a preceding value-taking flag.
   This covers first-token creation (`git branch <name> [start-point]`)
   AND option-prefixed creation (`git branch -- <name>`,
   `git branch -v <name>`, `git branch -q <name>`,
   `git branch --no-force <name>` — all auditor-measured creating branches
   on git 2.50.1) and creation modifiers (`--recurse-submodules`,
   `--create-reflog`) alongside a creation operand — including
   `git branch --create-reflog <name>` (M-30). Value-consumption arity
   rule (run-gate G1, refined at gate #2 by H1/H4, corrected at gate #3
   by L1/L3): a long flag with a REQUIRED space-separated value consumes
   the following token — pinned space-consuming set: `--contains`,
   `--no-contains`, `--merged`, `--no-merged`, `--points-at`, `--format`,
   `--sort` (`git branch --sort committerdate`,
   `git branch --no-contains HEAD`). ATTACHED-ONLY optional-value flags —
   `--color`, `--abbrev`, `--column` — consume a space-separated token
   NEVER: an attached `=<value>` is consumed, but a following
   space-separated token is a positional/creation operand (measured:
   `git branch --color colprobe` created `colprobe`, rc 0 — M-31/M-32);
   `git branch --color=always` (attached, no positional) is a query
   (Q-17). `--list` and `--show-current` are list-action selectors taking
   NO value. The short flag `-l` is the SHORT-FORM LIST SELECTOR
   (`git branch -l lpattern` — measured rc 0, no branch created: a live
   filter, unlike `-v <name>` = create), and list/FILTER mode is VARIADIC:
   selected by `--list`, `-l`, OR any filter selector (`--contains`,
   `--no-contains`, `--merged`, `--no-merged`, `--points-at`) — ALL
   remaining positionals after the selector's consumed value are filter
   patterns (`git branch -l foo bar`, `git branch --list foo bar`,
   `git branch --contains HEAD main` — measured rc 0, output `* main`,
   Q-16); Q-15 pins `-l`.
   Sibling resolved by measurement: `git branch -a <name>` is
   git-rejected (rc 128) → fail-closed-safe, not an over-match concern.
   Note: `git branch -v vbranch`
   therefore denies — matching git's measured semantics (flag + name =
   create); the doctrine Permitted list's `-v`/`-vv` entries name
   operand-free inquiry forms, noted for the REQ-WBG-F-009 bullet wording.
2. **Mutating flag anywhere**: any token is a mutating flag —
   - a short-flag cluster containing any of `d D m M c C f t u` — leading
     OR mid-cluster `u` (`git branch -umain` parses; `git branch -vu
     main` parsed and `git branch -vux` completed a mutation, measured at
     gate #2 — and no query short flag contains `u`, per the measured
     usage table; e.g. `-d`, `-df`, `-vD`, `-u`, `-f`, `-t`, `-vt` —
     `git branch -vt vtbranch main`, auditor-measured setting tracking,
     M-29, exercises the cluster-`t` scan);
   - a long flag whose name — taken as the full token, split at `=` first
     when an attached value is present (`--set-upstream-to=origin/main` →
     name `--set-upstream-to`) — exactly matches a member of the mutation
     set: `--force`, `--delete`, `--move`, `--copy`, `--set-upstream`,
     `--set-upstream-to`, `--unset-upstream`, `--track`, `--no-track`,
     `--edit-description`.

**Everything else allows**: query long flags (`--list`, `--show-current`,
`--contains`, `--merged`, `--no-merged`, `--points-at`, `--format`,
`--sort`, …) and their value operands; query short clusters (`-l`, `-v`,
`-vv`, `-a`, `-r`, `-q`, `-i`); unknown flags (fail-open, E-5 — including
git prefix-abbreviation long flags and the `git -C <path> branch …`
wrapper shape, both named residuals).

Per-flag operand-consumption semantics matter in both directions: a query
flag's operand (`--contains HEAD`, `--format %(refname)`, `--list
develop`) must not be read as a creation operand (rule 1 is first-token
only), and a mutation flag's operand (`--delete old`,
`--set-upstream-to origin/main topic`) must not make the token fail to
match (rule 2 is token-membership, not a trailing anchor). Long-flag
classification splits on `=` BEFORE name matching, so `--format` never
prefixes into `--force`, and the embedded letters in
`--contains`/`--merged`/`--no-merged` are never scanned.

Implementation may be a single regex or a small token classifier slotting
into the `branchStatePatterns` matching surface — M1's matrix decides
which; either must preserve quoted-span collapse (see
`branch_guard_quoted_test.go`) and `(?i)` for the `git branch` entry, and
must not change the other pattern entries.

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

- spec.md §C REQ-WBG-F-001..009; acceptance.md §D.1 matrix (incl. audit
  iteration-1 cells M-20..M-23) + §D.0 RED-now baseline.
- `verification-completeness.md` §2 two-cell adoption discipline (RED-now
  cell = acceptance.md §D.0 probe output on `d592b0551`; green path = M2).
- gitflow lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) —
  local develop merge via integration window; no direct push.
