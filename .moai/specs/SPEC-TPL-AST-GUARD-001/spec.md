---
id: SPEC-TPL-AST-GUARD-001
title: "AST-based workflow.worktree.* key-honesty guard — type-resolved reader map, superseding the t682 text scan in role"
version: "0.1.3"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/template"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "template, guard, ast, go-types, config-honesty, worktree, workflow-yaml, anti-rot, tdd"
related_specs: [SPEC-CONFIG-KEY-HONESTY-001, SPEC-WORKTREE-KEY-WIRING-001]
---

# SPEC-TPL-AST-GUARD-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-03 | Initial draft. Card t1377 (operator-directed, v3.2.0 mission): adopt an AST-based guard mechanism for the `workflow.worktree.*` template key-honesty contract, on the evidence basis of PR #1707 (CLOSED, unmerged) and its review thread. All file:line evidence verified in this session against worktree baseline `7c7c84b5c` (branch `WT-ast-template-guard`, = local develop tip). Reader map measured by grep + direct file reads; scanner-precedent runtime measured 24.0s (`go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -count=1` → `ok ... 23.997s`). Contract decision (D1) and scope decision (D2) presented with recommendations in plan.md §F and research.md §3-§4. |
| 0.1.1 | 2026-10-03 | Plan-audit iteration-1 revision (verdict FAIL 0.87 vs Tier M 0.80 threshold; four blocking defects + optional fold-ins; verdict file `.moai/reports/t1377/plan-audit.md`). Audit D1: both `-run` selectors anchored; AC-001 gains an executed-test-count assertion (a zero-match selector is a vacuous pass). Audit D2: §A.2/§A.3 narrative corrected — PR #1707's **final** diff was already AST-based (`ast.Inspect` scan, `TestFieldReadersIn_FollowsCopiesAndIgnoresOtherStructs`; the alias-blindness Major was fixed inside the PR at d102e1e); the PR closed UNMERGED, so the retained alias-blind exposure on develop is the t682 text guard, not the PR. Audit D3: REQ-004 is now an independent table-content assertion (set equality alone is defeated by a same-commit read+entry removal — the plan's own maintenance path), with a third AC-004 mutation arm. Audit D4: REQ-002's reader definition restricted to value-consuming uses; pure writes excluded, with a write-only fixture negative (AC-005b). Optional D5-D9 folded: template↔table residual named (§A.3 + Exclusions), build-context scope stated (C-7), regression-guard classification line (acceptance.md), AC-010/011 mapped, `Selections[sel].Obj()` made the primary resolution route (plan M1). |
| 0.1.2 | 2026-10-03 | Plan-audit iteration-2 revision (verdict FAIL 0.85; N1-N4 blocking + optional N5-N10 folded; verdict file `.moai/reports/t1377/plan-audit-iter2.md`). N1: REQ-002's read definition rewritten as **exclusion** — every type-resolved occurrence is a read except the direct left-hand operand of a plain `=` assignment; `op=`/`++`/`--`/address-of are reads; the definition is verified against all four real production readers (three are if-condition operands); the fixture gains a compound-assign positive (`compoundassign.go` — three fixture files, four new source files total). N2: file count unified on four source files across plan §D/M4, AC-011, and a path-filtered DoD item 2. N3: PR #1707's final state corrected to **syntax-only** (`go/parser` + syntactic `worktreeAliases` tracking, zero `go/types`; already write-excluding) — this SPEC's true delta is `go/types` resolution + the REQ-004 independent assertion + the C-7 scope statement. N4: the fixture gets its own test `TestWorkflowWorktreeKeyHonestyAliasFixture`, named in the GUARD anchored alternation; AC-001 requires both tests to execute. N6: the false "Uses comes back empty" claim removed — an executed probe at HEAD `84873b3c2` found `Uses == Selections.Obj()` at all four reader sites; route choice free, `Selections` primary. Optional N5 (arm counts), N7 (`types.go` citations 673-690 / 691-697), N8 (grep-hypothesis epistemics + unified probe regex), N9 (AC-012 negative-control classification + AC-008 cause-named), N10 (§F.0 refs, Exclusions renamed §D, title wording, REQ-004 exception in plan §C, doubled Whens collapsed) all applied. |
| 0.1.3 | 2026-10-03 | Kickoff gate-debt micro-fix (leader decision record `.moai/reports/t1377/kickoff-decision.md`: Kickoff APPROVED, PASS-WITH-DEBT 0.94 accepted as entry verdict; run-phase Phase 1 enters BYPASSED per that decision). F1: AC-011's gofmt command now lists all four new source files. F2: §D.12's Tested line classifies AC-012 as a negative control (captured PASS, not RED). F3 (M1-gate ordering notation) and F4 (REQ-005 reserved-key source notation) ruled notation-only by the leader — recorded as gate debt in progress.md §E.1, carried per leader decision 2026-10-03. |

## §A Problem / Motivation

### §A.1 The honesty claim being guarded

The shipped template `internal/template/templates/.moai/config/sections/workflow.yaml` (the file users receive) documents each `workflow.worktree.*` key with a comment clause claiming which production code reads the matching `config.WorkflowWorktreeConfig` field — or that none does. The struct (doc comment `internal/config/types.go:673-690`, struct body `:691-697`; baseline `7c7c84b5c`) declares exactly five fields:

| template key | struct field | production readers (verified this session, non-test `.go`) |
|---|---|---|
| `auto_create` | `AutoCreate` | `internal/cli/worktree_advisory.go` (`readWorktreeAutoCreate`, :66) |
| `auto_merge` | `AutoMerge` | `internal/cli/session_worktree_automerge.go` (`sessionExitAutoMerge`, :162) |
| `auto_cleanup` | `AutoCleanup` | `internal/cli/session_worktree.go` (`cleanupSessionWorktree`, :656) **and** `internal/cli/session_worktree_prmerge.go` (`prMergeCleanup`, :150) |
| `tmux_preferred` | `TmuxPreferred` | **none** (only `internal/config` test files reference it) |
| `session_name_pattern` | `SessionNamePattern` | **none** (declared-but-not-read; the template marks it reserved) |

The landed template clauses are prose-style and do not name `.go` files (except `auto_create`, which names `internal/cli/worktree_advisory.go`). The struct's own doc comment (`internal/config/types.go:673-690`, from SPEC-CONFIG-KEY-HONESTY-001 M5 + SPEC-WORKTREE-KEY-WIRING-001 REQ-WKW-012) carries the same reader status **with explicit file names** — it is the in-repo documentation SSOT for this map.

This map is a grep-derived hypothesis: the guard's first M1 AST run is the authoritative recomputation, and a reader the scan surfaces that the table lacks becomes either a table entry or a justified exclusion (the shipped-key precedent's handling).

### §A.2 The two prior guards, and why neither suffices

1. **The text-based guard (GH #1705 / card t682)** — `internal/config/workflow_key_honesty_test.go` (171 lines, verified this session). It scans production sources with `strings.Contains(src, ".Workflow.Worktree.AutoCleanup")` and fails when a template line calls a reader-bearing key "reserved". Two structural limits:
   - **Alias-blind** (the defect class PR #1707's review flagged MAJOR): production code that copies the struct first — `w := cfg.Workflow.Worktree; w.AutoCleanup` — contains neither the full accessor string nor any `.Worktree.`-suffixed read on one line, so the text scan misses the read and can produce a false "no reader" verdict. Text/line matching cannot see through local aliases or multi-line expressions.
   - **One-directional**: it only polices template *phrasing* against reader *existence* (`readers["auto_cleanup"] >= 1`). It does not require every reader file to be accounted for, does not validate a per-key file set, and historically accepted a single `auto_cleanup` reader site (the PR review's both-sites requirement, fixed at be8b2dc in the PR thread).

2. **PR #1707's bidirectional contract, closed unmerged** — the PR proposed `internal/template/workflow_worktree_key_honesty_test.go`: for each documented key, every file reading the field must be accounted for by the key's claim, and every file the claim accounts for must actually read the field. The review's alias-blindness Major was raised against the PR's **initial** text scan and was fixed inside the PR (d102e1e): the PR's final diff is AST-based but **syntax-only** — `go/parser` per-file parsing with a syntactic `worktreeAliases` alias tracker (alias-aware and write-excluding, but zero `go/types`/`TypesInfo`), pinned by `TestFieldReadersIn_FollowsCopiesAndIgnoresOtherStructs` — auditor-verified from `gh pr diff 1707`. The PR then CLOSED UNMERGED: its test file is absent on develop (verified: `git cat-file -e develop:internal/template/workflow_worktree_key_honesty_test.go` → absent) and absent on this worktree's HEAD — so neither the bidirectional contract nor the AST mechanism was ever adopted on develop. The alias-blind text scan that remains in-tree is the t682 guard of item 1.

### §A.3 What this SPEC builds

The card's deliverable: adopt that mechanism — re-implement PR #1707's final AST-based bidirectional contract on this branch, where it is absent — as a Go test in `internal/template` that computes each field's production reader set by resolving selector expressions through `go/types` (which sees through local alias copies, method-internal reads, and multi-line expressions by construction) and asserts it **exactly equals** a per-key expected reader table. The SPEC's delta over the PR's final state is threefold: (1) true type resolution — `go/types` selector resolution replaces the PR's syntax-only alias tracking, making the reader map alias-proof by construction rather than by tracked alias sets (REQ-001/002); (2) the independent REQ-004 table-content assertion; (3) the explicit build-context scope statement (C-7). The template stays untouched (contract decision D1, recommendation A — plan.md §F.0). The existing text guard in `internal/config` is retained, unchanged: its template-phrase check is complementary (the AST guard under contract A does not parse template comments at all). Known residual under contract A: a template claim diverging from the expectation **table** (rather than from the code) is caught by neither guard — the t682 check compares template prose against code, the new guard compares table against code — recorded as a follow-up candidate (research.md §3; revisit contract D1 if it bites).

## §B Requirements (GEARS)

- **REQ-001** The AST-based worktree-key honesty guard (a Go test in `internal/template`) shall compute the production reader set of every `WorkflowWorktreeConfig` field (`AutoCreate`, `AutoMerge`, `AutoCleanup`, `TmuxPreferred`, `SessionNamePattern`) by resolving Go selector expressions through type information (`go/types`), never by matching source text, line content, or accessor substrings.
- **REQ-002** When a production (non-test) Go file accesses a `WorkflowWorktreeConfig` field through any type-resolved expression — a direct selector (`cfg.Workflow.Worktree.AutoCleanup`), a local alias copy (`w := cfg.Workflow.Worktree; w.AutoCleanup`), an if-condition operand, or any other shape — the guard shall attribute that access to the file containing it, with **access defined by exclusion**: every type-resolved occurrence of the field is a read EXCEPT the field selector appearing as the direct left-hand operand of a plain `=` assignment (`w.AutoCleanup = false` is not a read). The other assignment forms are reads and are not excepted: compound assignments (`w.SessionNamePattern += s` consumes the old value), `++`/`--` operands, and address-taking (`&w.AutoCleanup`). The positional examples in this entry are illustrative; the exclusion rule is the decision procedure, and it classifies all four real production readers (`worktree_advisory.go:66` return operand; `session_worktree.go:656`, `session_worktree_prmerge.go:150`, `session_worktree_automerge.go:162` if-condition operands) as reads.
- **REQ-003** When the computed reader set of a key differs from that key's expected reader set — an expected file no longer reads the field, or a file absent from the expectation reads it — the guard shall fail with findings naming the key and each offending file.
- **REQ-004** When the `auto_cleanup` expected reader set does not contain both `internal/cli/session_worktree.go` and `internal/cli/session_worktree_prmerge.go`, the guard shall fail — enforced as an independent table-content assertion that holds regardless of the computed reader set (the PR review's both-sites requirement; a same-commit removal of a read and its table entry therefore still fails).
- **REQ-005** When a key's expected reader set is empty (unread/reserved — `TmuxPreferred` and `SessionNamePattern` on the current tree), the guard shall fail if any production reader of that field exists, and shall reject the expectation table itself as invalid when it names any file under that key.
- **REQ-006** When `WorkflowWorktreeConfig` declares a field with no entry in the expectation table, the guard shall fail — the table is checked for completeness against the live struct, so the table cannot silently go stale when a field is added.
- **REQ-007** When any package scanned by the reader index carries type errors, the guard shall fail rather than pass on a possibly-incomplete reader set.
- **REQ-008** While computing the production reader set, the guard shall exclude `*_test.go` files from the reader set and from the unnamed-reader comparison.
- **REQ-009** Where the alias-detection characterization runs, the guard shall scan a fixture package under `internal/template/testdata/` — which imports `internal/config` and reads a `WorkflowWorktreeConfig` field through a local alias copy — using the same scan path as the production run, and shall detect that fixture file as a reader of the aliased field.

## §C Constraints

- **C-1 Zero new dependencies.** `golang.org/x/tools v0.49.0` is already a direct dependency (`go.mod:32`, verified). The guard uses `golang.org/x/tools/go/packages` — the in-repo precedent (`internal/config/shipped_key_reader_test.go:395-470`) already loads the whole module this way. The stdlib-only alternative (`go/parser` + `go/importer` source importer) was weighed and rejected; tradeoff recorded in research.md §5.
- **C-2 No template content changes** (contract A). The shipped workflow.yaml comment clauses stay exactly as landed on develop. Run phase therefore needs **no `make build` regeneration**.
- **C-3 No production config code changes.** No new readers are wired for `TmuxPreferred` / `SessionNamePattern`; the guard enforces honesty of the reader map, not reader existence.
- **C-4 Existing guards retained.** `internal/config/workflow_key_honesty_test.go` (t682) and `internal/config/shipped_key_reader_test.go` are not modified or retired by this SPEC.
- **C-5 Methodology: TDD.** The deliverable IS a test; its RED state is demonstrated by mutation (run the mutation, observe the failure, revert), never by assertion alone.
- **C-6 Runtime envelope.** The whole-module `packages.Load("./...")` scan measured 24.0s in this session on the in-repo precedent; the new guard is expected to stay in that order of magnitude (guidance, not an acceptance criterion — machine-load-dependent).
- **C-7 Baseline and build-context scope.** All §A file:line evidence verified against worktree baseline `7c7c84b5c` (branch `WT-ast-template-guard`); run-phase pre-flight re-verifies the reader map before implementing (plan.md §C). The production reader set is defined for the build context the guard runs in (`packages.Load` resolves the current GOOS/GOARCH): a future platform-constrained reader file (e.g. `_windows.go`) is invisible to a guard run on darwin — cross-GOOS readers are out of scope until a union scan is justified (no build-tagged reader files exist on the current tree).

## §D Exclusions

### Out of Scope — template wording and shipped keys

- No edits to `internal/template/templates/.moai/config/sections/workflow.yaml` — the comment clauses stay prose-style as landed (contract A; option B is a scope alternative documented in plan.md §F.0, not planned).
- The commented-out `sparse_paths` example in the template is untouched: `WorkflowWorktreeConfig` declares no such field, and wiring one is a separate config-surface concern.
- No new production readers are wired for `TmuxPreferred` or `SessionNamePattern` (the shipped active `tmux_preferred: true` value staying unread is exactly what the guard formalizes).

### Out of Scope — existing guards and house-wide migration

- `internal/config/workflow_key_honesty_test.go` (the t682 text guard) is retained unchanged — its template-phrase direction is complementary under contract A; its retirement or consolidation into the new guard is a follow-up candidate only.
- No template-prose ↔ expectation-table consistency check (the contract A residual named in §A.3): the shipped workflow.yaml claims are guarded against code drift only by the retained t682 phrase check; a table↔prose divergence follow-up is a separate candidate.
- `internal/config/shipped_key_reader_test.go` (key-liveness classification, SPEC-CONFIG-KEY-HONESTY-001 REQ-CKH-008) is untouched — different contract (is the key read at all), not reader-set honesty.
- No migration of other `internal/template/*_audit_test.go` text-based guards to AST scanning (card scope option (c)) — explicitly a follow-up candidate only.
- No shared/reusable scanner-helper extraction beyond what the single test file needs (card scope option (b)) — no second consumer exists today; extraction is mechanical later if one appears.
