# Plan — SPEC-MAIN-COMMIT-BAN-001

> Tier M · cycle_type: tdd (Scope 2 carries the RED→GREEN core; Scopes 1/3/4 are doc+config
> surface verified mechanically). Worktree: card worktree on `WT-main-commit-ban`
> (base 145c3d98c = develop tip). Commit nothing in plan phase.

## §A Design Decisions (settled)

### D1 — Guard layer: extend the PreToolUse BranchGuard, NOT a pre-commit hook, NOT a new guard family

- **Chosen**: a second, branch-conditional deny class inside the same `checkBranchState` seam of
  `internal/hook/branch_guard.go`, invoked from the same gated call site in `pre_tool.go`
  (after the existing branch-state check, before the allow fall-through). New function pair:
  `matchProtectedCommitCommand(command) bool` + `checkProtectedCommit(input, projectDir)` —
  sharing Seams A (`resolveProjectRootFromInputOrEnv` → `gitcore.IsPrimaryCheckout`),
  `isExemptAgent`, the fail-open advisory path, and the command normalization pipeline
  (quoted-argument / heredoc / comment / PowerShell substitution via `matchBranchStateCommand`'s
  normalize step).
- **Rejected — git pre-commit hook**: cannot see Claude session context; would also bind the
  operator's own terminal, where the residue procedure (REQ-4) must stay executable; needs
  per-clone installation; cannot ride the existing `Workflow.BranchGuard.Enabled` opt-in
  semantics.
- **Rejected — new guard family (`commit_guard` sibling)**: duplicates the primary-checkout
  discriminant, exemption axes, fail-open advisory, and audit-log machinery `checkBranchState`
  already owns; one more config key + one more PreToolUse scan pass for zero added hazard
  coverage. The commit deny is the same hazard domain (branch state of the shared primary
  checkout).

### D2 — Deny scope: commit-CREATING commands, primary checkout, protected branch only

- **Chosen matcher**: `\bgit\s+(commit|revert|cherry-pick)\b` (case-insensitive, same
  compilation convention) evaluated ONLY for the protected-branch deny. `git commit` covers
  `--amend` and `-a` by the word-anchor. Detached HEAD → allow (no named branch to protect —
  deliberate, stated in code comment).
- **Rejected — deny ALL commits in the primary checkout**: breaks the doctrine's legitimate
  "commits to the already-checked-out branch are permitted" for develop (the standing §2 clause
  survives for non-protected branches).
- **Rejected — broader verb set (`pull`/`push`/`tag`)**: pull/fetch of main is legitimate
  synchronization (lane protocol §1: main is a sync-only reference); scope discipline — the
  operator directive names commits. Recorded as accepted residual E-1/§D Non-Goals.
- Note: `merge`/`rebase`/branch-mutations on main need no new work — the existing pattern set
  already denies them in the primary checkout on ANY branch.

### D3 — Branch awareness: HEAD resolution at the command's actual cwd, bounded subprocess cost

- `internal/core/git` gains `ResolveHeadBranch(cwd string) (string, error)` (hosted beside
  `ResolveGitDirs`), implementing `git branch --show-current` (git 2.22+; empty output on
  detached HEAD → `("" , nil)` → allow) at the Seam-A-resolved cwd.
- **Evaluation order** (bounds cost to the positive path): command normalization →
  protected-commit pattern match → exemption check → primary-checkout discriminant → THEN (and
  only then) HEAD resolution → list membership → deny. With the list empty, the whole
  commit-family check short-circuits BEFORE the pattern scan (REQ-2.5) — unconfigured users pay
  one string-list `len()` check.

### D4 — Config: `deny_commits_on` list, empty default; NO new opt-in flag

- `BranchGuardConfig` gains `DenyCommitsOn []string \`yaml:"deny_commits_on"\``;
  `defaults.go` → `[]` (template-neutral, sibling of `Enabled: false`); the TEMPLATE
  `internal/template/templates/.moai/config/sections/workflow.yaml` (plain `.yaml`, no `.tmpl`
  suffix — measured) has NO `branch_guard` block today (grep empty, measured), so M1 CREATES
  the block there: `branch_guard:\n        enabled: false\n        deny_commits_on: []` plus a
  neutral comment sentence. This is NEW template-config surface — stated, not an edit of an
  existing block. LOCAL `.moai/config/sections/workflow.yaml` (whose existing commented
  `branch_guard:` block is measured at lines 158-165) sets `deny_commits_on: [main]`.
- **Rejected — hardcoding `"main"`** in the Go code: violates the hardcoding ban
  (AGENTS.local.md §14, the develop-tracked local doctrine — names/thresholds belong in config; a list also generalizes to
  release/* style protections).
- **Rejected — separate `deny_commits: true` boolean**: a boolean still needs a branch name, so
  the list subsumes it; a second flag doubles the config surface for one behavior.
- Flag/default question settled: rides the existing `Workflow.BranchGuard.Enabled` gate —
  dogfood already `enabled: true`; template default stays `false`. Double-inert default for
  distributed users (flag off AND list empty).

### D5 — Doctrine edits: parallel-but-divergent local/template pair; incident lives local-only

Every lane-editable target below is measured to exist in this worktree. The designated probe
sentinel for the local doctrine set is the literal English token **`commit-dead`** (embedded in
each file's clause — Korean prose embeds English tokens per house style), so AC-16 probes ONE
exact phrase instead of a weak generic verb.

| File | Mirror class | Edit |
|---|---|---|
| `AGENTS.md` (local, tracked) | local-only | §2 permitted clause reworded: pushing the checked-out branch + inspection stay permitted; commits permitted EXCEPT on `main` — commit-dead, pointer to `.moai/docs/gitflow-integration-chain.md`. Sentinel: `commit-dead` |
| `internal/template/templates/AGENTS.md.tmpl` | TEMPLATE (neutral) | Same clause position, GENERIC form: "except on a branch the workflow declares commit-protected (the git-strategy mode decides which)". Zero card ids / dates / provenance narrative. Enforced classes if violated: C1 SPEC-ID, C3 audit-citation, C4a date, C4b short-sha, C5 memory-path (measured catalogue, `internal_content_leak_test.go`); card-id/CLAUDE.local-ref prohibitions are prose-level, guarded de facto by sanitized-pair parity |
| `.claude/rules/moai/workflow/main-checkout-branch-guard.md` + template mirror | sanitized-pair | "Mechanical Enforcement" bullet list gains ONE bullet: protected-branch commit deny (config key, sentinel, same fail-open norm). Template side stays class-neutral |
| `AGENTS.local.md` §4.1 (develop-TRACKED — measured `git ls-files`; the develop-side home of the gitflow chain doctrine) | local-only | [HARD] bullet: local main commit-dead + procedure pointer + `lead_push_threshold` key citation. Sentinel: `commit-dead` |
| `.claude/rules/local/gitflow-lane-protocol.md` §1 (measured in worktree) | local-only | strengthens the existing "main은 동기화만 하는 참조점" clause: main은 commit-dead — 어느 세션도 그 안에서 커밋하지 않는다. Sentinel: `commit-dead` |
| `.moai/docs/gitflow-integration-chain.md` (measured in worktree) | local-only | hosts the residue procedure (Scope 3) + threshold trigger semantics (Scope 4) + incident reference |
| `.moai/docs/git-workflow-doctrine.md` (~line 64 block, measured) | local-only | one clause added to the 2026-08-27 git-flow revision note: main is commit-dead, not merely release-PR-updated |
| `CLAUDE.local.md` §4.1 (primary-ONLY: untracked, gitignored `.gitignore:276`, absent from worktrees — measured) | OPERATOR-side | NOT a lane edit. The same clause lands in the primary's working copy at disposition time (REQ-4.6); the LEAD verifies it at the primary (AC-16 primary probe set). Editing it from the lane would violate worktree discipline and the one-writer rule |

- **Execution surface rule (audit F1)**: the lane writes only inside its worktree; the
  primary-only file is handled exactly like Scope 3's real-primary steps — operator-side,
  lead-verified.
- **Parity gates to keep green**: `TestSanitizedPairParity` (lists
  `main-checkout-branch-guard.md` + `-detail.md` as sanitized pairs),
  `TestRuleTemplateMirrorDrift` + `TestDeclaredRuleMirrorForks` (byte-parity allowlist — the
  guard rule is NOT byte-parity; sanitized parity governs),
  `TestTemplateNeutralityAudit` + `TestTemplateNoInternalContentLeak` (C1/C3/C4/C5 classes) on
  touched template files.
- **Rejected — editing only the local AGENTS.md**: the template mirror's §2 carries the same
  permitted-commit sentence; leaving it would ship the contradiction to every user project whose
  git-strategy mode declares a protected mainline, and the two copies would teach opposite
  contracts.
- **Rejected — putting the incident (`c8f245c2c`, card ids, dates) in template copies**:
  leak classes C3/C4a/C4b enforce audit-citations, dates, and short-SHAs; the card-provenance
  prose prohibition plus sanitized-pair parity close the rest.

### D6 — Residue procedure: operator-terminal execution, origin/main target, fixed ordering

- Home: `.moai/docs/gitflow-integration-chain.md` (local-only; already the operational-procedure
  owner). Structure per REQ-4.1-4.5. The three coordinator-fixed specifics are encoded:
  ordering as precondition; preservation copy cited reference-only (NO new artifact — develop
  already treats `CLAUDE.local.md` as untracked+gitignored, which IS the normal post-switch
  state); incident `c8f245c2c` recorded as motivation.
- **Reset target — chosen `origin/main`, rejected `develop`**: main's only legitimate role is
  the synchronized mirror of the release-PR landing surface; `branch -f main develop` would
  manufacture a second develop head and a false main-ahead-of-origin reading (statusline ↓N
  honesty + the exact c8f245c2c defect shape). `origin/main` makes the ↓N signal read true.
- **Execution channel — operator's own terminal, stated as design not workaround**: the branch
  switch and `branch -f` are precisely what the guards deny to agent sessions in the primary
  checkout; the procedure says so explicitly so a session-side deny reads as correct behavior.
  The lane NEVER executes steps 2-3 on the real primary (REQ-4.4); run phase produces a scratch
  clone rehearsal + the observed-refusal log only.
- Switch mechanism — REWRITTEN from the measured tree (audit F5): `main` does NOT track
  `CLAUDE.local.md` (deleted from tracking by `c8f245c2c`; `git cat-file -e main:CLAUDE.local.md`
  → absent) — the file is untracked+gitignored on BOTH sides of the switch and survives it
  silently; the earlier "tracked-file denial" story was false. The boundary the switch actually
  crosses is the primary's develop-state working copy: uncommitted modifications to files still
  tracked on main (modified-tracked-set refusal) plus untracked working files that develop
  tracks (untracked-overwrite refusal). The procedure therefore opens with a measured inventory
  (`git status --porcelain` in the primary, operator-side) and a content-safety precondition —
  `git diff develop --stat` EMPTY (the working tree already equals develop's tree, so the
  switch is content-preserving by construction); non-empty → STOP, operator inventories each
  entry. Post-procedure note: `main = origin/main` re-tracks `CLAUDE.local.md`, so a future
  primary switch to main refuses "untracked would be overwritten" — a documented re-armed
  boundary (REQ-4.6), not a defect.
- Rehearsal step is mandatory and must reproduce the primary's REAL state class: a fresh clone
  at origin/main (where the file is tracked+clean) would observe a SILENT REMOVAL on switch,
  not a refusal — so the rehearsal recipe is: clone → `git switch c8f245c2c` → materialize
  develop-state content as unstaged modifications over main's tracked set (`git restore
  --source=develop -- <develop-tracked paths>`) → run the procedure steps → record the actual
  refusal set. Whatever git does is what the procedure documents (AGENTS.md §1: no unobserved
  claims).

### D7 — Threshold: config-key SSOT, doctrine cites the key; value initial=20, operator-given

- `git_strategy.manual.lead_push_threshold: 20` in LOCAL `.moai/config/sections/git-strategy.yaml`;
  template mirror carries the key with `0` (= disabled — manual-mode default has
  `push_to_remote: false`, so a push threshold is meaningless; neutrality preserved).
  Go: git-strategy manual struct gains `LeadPushThreshold int`, default `0`,
  symmetry test updated (`CONFIG_STRUCT_YAML_MISMATCH` enforces this — the YAML key cannot
  exist without the field).
- **Rejected — doctrine-only (doc states 20, no key)**: not machine-readable; the AC requires a
  parse assertion; also `moai update` re-application of git-strategy.yaml (AGENTS.local.md §2.3
  reapplication ritual)
  would not carry a prose value.
- **Rejected — new section file**: 5-step new-section procedure (loader, struct, defaults,
  loader file, wiring) for one integer; the existing `git_strategy` section is the natural home.
- Semantics (REQ-5): the lead reads `git rev-list --count origin/develop..develop`; count ≥
  threshold → close the batch (collect lane merge SHAs → push once → verify landing). The
  threshold never interrupts an open integration window and never authorizes a lane push —
  develop push remains the lead's sole, window-outside, batch act. Doctrine names the key and
  counter; config is authoritative for the current value; prose records "20" as the initial
  operator-given value, not a derivation. The key citation lands in `AGENTS.local.md` §4.1
  (develop-tracked, lane-editable — audit F1 re-aim); the primary-only `CLAUDE.local.md` copy
  receives it via the REQ-4.6 operator-side sync.

### D8 — Template-key instrument: a dedicated presence test, NOT a deeper symmetry harness (audit F4)

- **Chosen**: a NEW test `internal/config/template_main_commit_ban_keys_test.go`
  (`TestTemplateConfigCarriesMainCommitBanKeys`) that reads BOTH template yamls
  (`internal/template/templates/.moai/config/sections/workflow.yaml` →
  `workflow.branch_guard.deny_commits_on` present; `.../git-strategy.yaml.tmpl` →
  `git_strategy.manual.lead_push_threshold` present) AND both LOCAL tracked yamls (asserting
  `deny_commits_on` contains `main` and `lead_push_threshold == 20`). One test closes the whole
  F4 gap: template-key presence (a `moai update` re-application dropping the template key would
  go red) and the operator-given local values.
- **Why not extend `checkSymmetry` deeper**: the instrument is one level deep BY DESIGN
  (measured: `checkSymmetry` walks top-level keys only; `symmetryCases` covers Constitution /
  Context / Interview / Design / Statusline / GitConvention / Gate — workflow and git-strategy
  are absent). Deepening it would sweep every section's nested keys into enforcement at once —
  a blast radius far beyond this SPEC, likely red on pre-existing sections. The targeted
  presence test enforces exactly the two new keys, and the existing
  `^TestStructYAMLSymmetry_` family stays untouched (it would be green whether or not the keys
  land — a vacuous criterion for them, which is precisely why it is not the AC's instrument).

## §B Milestones

### M1 — Config surface (code, no behavior change alone)

Files:
- `internal/config/types.go` — `BranchGuardConfig.DenyCommitsOn`; git-strategy manual struct is
  `ModeProfile` (MEASURED, `types.go:180` `Manual ModeProfile \`yaml:"manual"\``) — gains
  `LeadPushThreshold int`.
- `internal/config/defaults.go` — both defaults (`[]` / `0`) with neutrality comments.
- NEW `internal/config/template_main_commit_ban_keys_test.go` — the D8 presence test (the
  instrument for both new keys; the existing symmetry harness cannot see them — its
  `symmetryCases` covers 7 sections, workflow/git-strategy absent, and `checkSymmetry` walks
  one level deep, both measured). No `symmetryCases` edit — the new keys are two levels deep
  and out of that harness's scope by design.
- `internal/template/templates/.moai/config/sections/workflow.yaml` (plain `.yaml`) — CREATES
  the `branch_guard:` block (`enabled: false` + `deny_commits_on: []` + neutral comment); no
  such block exists there today (measured).
- `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` —
  `lead_push_threshold: 0` under `manual:` (line 17 block, measured).
- `.moai/config/sections/workflow.yaml` (LOCAL) — `deny_commits_on: [main]` under the existing
  commented `branch_guard:` block (lines 158-165, measured).
- `.moai/config/sections/git-strategy.yaml` (LOCAL) — `lead_push_threshold: 20` under
  `manual:`.

Verify: `go test ./internal/config/... -count=1` green; the new presence test selects via
`go test ./internal/config/ -run '^TestTemplateConfigCarriesMainCommitBanKeys$' -count=1`.

### M2 — Protected-commit deny (RED→GREEN core)

Files:
- `internal/hook/branch_guard.go` — `protectedCommitPattern` (compiled once),
  `matchProtectedCommitCommand`, `checkProtectedCommit`; call site in `pre_tool.go` after the
  existing `checkBranchState` block, same `Workflow.BranchGuard.Enabled` gate;
  `internal/core/git` `ResolveHeadBranch` + its unit test.
- NEW `internal/hook/branch_guard_protected_commit_test.go` — table-driven per house idiom
  (sibling: `branch_guard_flagclass_test.go`); package-var swap idiom for the HEAD resolver
  (counting stub, per the M6 deny-origin precedent that swaps `branchStatePatterns`).

RED cell (why red): `checkBranchState` has no commit-family matcher today — `git commit` on a
protected branch in the primary returns allow (`("","")` fall-through); the new tests fail on
the pre-implementation tree for that stated reason. RED baseline SHA: `145c3d98c` (the iter1
audit tree, where the deny provably does not exist); at run-phase RED execution the four
elements (command, verbatim output, exit code, actual tree SHA) are captured and re-pinned in
the evidence path per verification-completeness §2.1. GREEN: all cases pass
post-implementation.

Test matrix (audit F3 corrected): commit/revert/cherry-pick on main@primary+configured → deny
(prefix asserted); commit on develop@primary+configured → allow; commit on
main@worktree cwd → allow; list empty → allow with ZERO `ResolveHeadBranch` invocations (stub
count == 0); HEAD query error → allow + audit-log line; detached HEAD (empty resolver output)
→ allow; `MOAI_BRANCH_GUARD_EXEMPT=1` → allow; `AgentType == "manager-git"` → allow;
`git commit -m "git switch main"` @main+primary+configured → **DENY** — the quoted text is
data, but the command verb IS a commit on the protected branch (the earlier "allow" cell was a
matrix contradiction, audit F3); normalization's allow direction is covered by a NON-commit
carrier: `moai todo add "git commit -m x"` → allow (no commit verb at command position —
quoted git prose in a foreign command stays inert); compound `git status && git commit -m x`
@main+primary → deny; PowerShell payload form → deny. Known under-match documented in the test
notes (E-1): `git -C <primary-path> commit` from a worktree cwd — the cwd-based discriminant
classifies it as worktree, accepted fail-open residual.

Verify: `go test ./internal/hook/ ./internal/core/git/... -count=1` green with selector
`-run '^TestProtectedCommit'` for the new family; `go vet` those packages;
`golangci-lint run` (CI version, v2.1.6) on touched packages.

### M3 — Doctrine edits (Scopes 1+3+4 prose)

Files: the §D5 table — 7 lane-editable files + 1 operator-side file (`CLAUDE.local.md` §4.1,
landed at disposition time per REQ-4.6, lead-verified) + `gitflow-integration-chain.md`
procedure/trigger sections (Scopes 3+4 semantics, including the modified-tracked-set boundary
and the post-procedure re-armed switch-to-main note).

Verify: `commit-dead` sentinel grep over the lane-editable set (AC-16 worktree probe set);
template-side guards with exact anchors:
`go test ./internal/template/ -run '^(TestTemplateNeutralityAudit|TestSanitizedPairParity|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestTemplateNoInternalContentLeak)$' -count=1`.

### M4 — Rehearsal + verification batch

- Scratch-clone rehearsal of the Scope 3 procedure per D6's recipe (clone → `git switch
  c8f245c2c` → develop-state content as unstaged modifications → procedure steps), recording
  the ACTUAL refusal set to `.moai/reports/t1337/` evidence. The real primary checkout is NOT
  touched; its state is verified read-only from the worktree via the shared object store
  (`git rev-parse`/`rev-list` refs) — never `git -C <primary>`.
- Lane-local verification batch (single turn, parallel Bash): scoped `go test` (hook + core/git +
  config + template targets above, all `-count=1`), `go vet` on touched packages,
  golangci-lint (CI version), `commit-dead` sentinel batch, and the D8 presence test.
  Primary-run probes (CLAUDE.local.md §4.1 clause; primary HEAD state) are LEAD-run from the
  primary side — the lane reports them as a pending lead checklist, never executes them.

## §C Per-AC → Milestone map

AC-1..AC-3 → M3 · AC-4..AC-11 → M1+M2 · AC-12..AC-13 → M3+M4 · AC-14..AC-15 → M1+M3 ·
AC-16 → M4. Full matrix: `acceptance.md`.
