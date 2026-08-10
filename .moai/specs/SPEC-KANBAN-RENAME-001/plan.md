---
id: SPEC-KANBAN-RENAME-001
title: "Implementation plan — Factory Mode to Kanban Mode rename"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, refactor, cli, template-mirror, behavior-preserving"
tier: L
---

## §A. Context and measured baseline

Everything below was measured in the `kanban` worktree at HEAD `d39e3cdc6` (clean, at `origin/main`). Figures are attributable to the commands recorded beside them; none is carried over from a prior session.

| Fact | Command | Observed |
|---|---|---|
| Kanban-Mode surface | §D.1 token grep over `internal/ .claude/ .moai/project/` | 28 files (26 + 2 under `.moai/project/`) |
| Surface outside the first scope | same pattern, whole tree | `.moai/project/codemaps/modules.md`, `.moai/project/structure.md` — both name `internal/factory` and `moai cc -f` |
| Bare `-f` residue in contract docs | the bare-short-flag grep of acceptance.md AC-KR-026, run over the six docs on both sides | 8 occurrences (2+1+1 local, mirrored) |
| `-k` collision in the `claude` CLI | the short-flag probe of acceptance.md AC-KR-004 | no match; short flags are `-c -d -h -n -p -r -v -w` |
| Package count (test-run scale) | `go list ./...` piped to `wc -l` | 115 — a `tail -20` hides ~95 lines of a full-suite run |
| docs-site involvement | `grep -rni factory docs-site/content/` | 4 hits, all `ExecutionFactory` (unrelated) |
| CHANGELOG involvement | `grep -ci factory CHANGELOG.md` | 0 |
| `-f` release exposure | `git cat-file -e <tag>:internal/cli/factory.go` | absent in `v3.0.1`; present in `v3.1.0-rc.0`, `-rc.1` |
| Mirror classification | `diff` per pair | 3 byte-identical, 3 sanitized |
| `-k` collision in-repo | `grep -rn '"-k"' --include='*.go' internal/` | none |
| Catalog indexing | `grep -n 'workflows/' internal/template/catalog.yaml` | none — catalog indexes skill **directories** with a content hash |

Two of these corrected the working assumptions this plan started from, and both corrections are load-bearing (§B).

---

## §B. Known issues in the framing (read before executing)

**B-1 — "the mirrors are byte-identical" is false at HEAD.** Three of the six pairs are *sanitized pairs*: the local copy carries `Updated:` dates and SPEC-ID-bearing paragraphs that §25 Template Internal-Content Isolation forbids in template source, so the template copy has them stripped. An implementer who "restores parity" by copying the local file verbatim over the template file will re-introduce the forbidden content and trip the neutrality guard. The invariant to hold is **delta preservation** (REQ-KR-018), not parity. The classification is time-varying and is re-measured in M0, not read from this file.

**B-2 — the completion grep must be token-scoped.** A case-insensitive grep for the bare word `factory` over `internal/` and `.claude/` matches roughly 110 files of unrelated vocabulary: `clientFactory` throughout `internal/lsp/core`, the deliberate "Interface + Factory" anti-pattern example, the Apache-2.0 attribution to `revfactory/harness`, an `@MX:ANCHOR` renderer-factory comment. Enumerating those as an allowlist would be unmaintainable and would make the criterion a judgment call. The §D.1 token pattern was falsified against those trees before adoption and returns zero there.

**B-3 — `catalog.yaml` records a directory hash, not a file list.** Grepping the catalog for `factory` or for `workflows/` returns nothing, which could be misread as "the catalog is unaffected". It is affected: the catalog carries one `hash:` per skill directory, and renaming `factory.md` → `kanban.md` inside `.claude/skills/moai/` changes the `moai` skill's hash. `make build` regenerates it and the result must be committed (REQ-KR-020).

**B-4 — `-k` collision: probed, no collision.** The launcher strips its own switch before passing the remaining argv through to `claude`, so a `claude`-defined `-k` would be silently shadowed. The probe was run at plan-phase: `claude --help` matches no `-k`, and the full short-flag set is `-c -d -h -n -p -r -v -w`. M0 keeps the gate because the `claude` CLI surface drifts between versions and the run-phase tree may sit on a different one — but M0 now **re-confirms a recorded answer** rather than resolving an open question. Recorded limitation: the probe pattern matches `-k ` and `-k,` renderings and would miss a `-k=<value>` form.

**B-5 — two surface files live outside `internal/` and `.claude/`.** `.moai/project/codemaps/modules.md` carries a `### internal/factory` section (heading, a role line documenting `moai cc -f` / `moai glm -f`, an entry-point line naming `internal/cli/factory.go`) and `.moai/project/structure.md` carries a package-count paragraph naming `internal/factory`. The v0.1.0 scope missed both, so the completion grep would have returned 0 while two documents described a deleted package. They are now in scope (REQ-KR-024, AC-KR-028) and the grep scope carries `.moai/project/`. Neither is template-mirrored, so neither joins the six-pair delta comparison.

**B-6 — four command-shape hazards that produce vacuous GREENs.** Recorded here because they are what an implementer copies without noticing:
- `cmd | tail -N; echo "exit=$?"` reports `tail`'s exit code. Redirect to a log, read `$?` first, then grep and tail the log.
- A bare `git diff` compares the working tree to the index, so once M1-M3 are committed it is empty regardless of what happened. Anchor to `d39e3cdc6..HEAD`.
- `\|` and `\b` are GNU-only. They work under this machine's `grep` and silently match nothing under BSD `grep`. Use `-E` with a plain `|`.
- **`go test -run <pattern>` with a pattern matching nothing exits 0 and prints `PASS`.** Measured on `go1.26.4`: `go test ./internal/factory/ -run 'ZZZNoSuchTestName' -v; echo $?` → `testing: warning: no tests to run` / `PASS` / `ok … [no tests to run]` / `0`. Every `-run` here names a **post-rename** test function, so an implementer who renames production code and skips step 6 of M1 selects zero tests and reports green. `AC-KR-001` and `AC-KR-005` therefore pair each run with a name-existence `grep` and an assertion that the literal `[no tests to run]` is absent from the log — that literal, not the `-v`-only warning line, because only it appears in both `-v` and non-`-v` output.

---

## §C. Pre-flight

Run in the worktree, in one batch:

```bash
git rev-parse --show-toplevel          # must print the kanban worktree
git status --porcelain                 # must be empty
go build ./... && go test ./... > /tmp/kr-base.log 2>&1; echo "base-exit=$?"
grep -c '^FAIL' /tmp/kr-base.log
tail -20 /tmp/kr-base.log
```

Require `base-exit=0` AND a `^FAIL` count of `0`, both read from the whole log. The exit status is read before any pipe, and the `FAIL` count is taken over the full file rather than a tail — a piped `tail` reports the exit status of `tail`, and truncates roughly 95 of the 115 package lines, so a red baseline would present as green. Same command shape as AC-KR-011, and required by AP-8 / §B-6.

A red baseline must be resolved or characterized before the rename starts; otherwise a post-rename failure cannot be attributed — which is exactly why this check must not be able to report green on a red tree.

Note the environment constraint: `.git/hooks/pre-commit` runs `moai gate`, which currently exits non-zero on pre-existing ast-grep findings and blocks every commit. Use `SKIP_MOAI_PRECOMMIT=1` for each commit in this SPEC. That defect is out of scope here.

---

## §D. Constraints

1. **Template-First** — template source under `internal/template/templates/` is edited before its local counterpart; `make build` follows; the regenerated `catalog.yaml` is committed.
2. **Delta preservation** — per §B-1, not byte-parity.
3. **Neutrality** — the renamed `templates/.claude/skills/moai/workflows/kanban.md` carries no SPEC ID, REQ/AC token, audit citation, internal date, or commit SHA. The mechanical authority is the shipped guard, not a re-implemented regex.
4. **Full suite** — `go test ./...`, never an affected-packages subset.
5. **Env constants** — `internal/config/envkeys.go` holds the strings; no call site inlines them.
6. **Behavior preservation** — no assertion is added, weakened, or removed.

---

## §E. Self-verification

Before declaring the SPEC done, the implementer produces a 5-section evidence report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) covering: the §D.1 grep at 0, the full-suite result, the six-pair delta comparison, the neutrality guard, and the committed `catalog.yaml`. Each row names the command that decided it and quotes its output.

---

## §F. Milestones

Ordered so the reversible-decision work lands first and the mechanical sweep last. M1, M2, and M3 are separable units, each independently committable and independently verifiable.

### M0 — Baseline measurement and the one open decision

Priority: **High** (gates everything else).

1. Capture the §D.1 grep baseline to a file over `internal/ .claude/ .moai/project/` (`28` files expected) — this is the denominator M4 measures against.
2. Capture `diff` for each of the six mirrored pairs to a file. This is the delta-preservation baseline (§D.2 of spec.md); it must be taken *before* any edit or it cannot be reconstructed. Write each to `/tmp/base-<label>.diff` using the labels AC-KR-017 reads back — `contract` (`skills/moai/workflows/factory.md`, still under its pre-rename name here), `run`, `goal`, `moaidoc`, `modeorch`, `qgates`. The label, not the filename, is the key: the contract document's basename changes at M2, and keying the baseline on the basename would orphan it.
3. Re-confirm the `-k` probe: `claude --help 2>&1 > /tmp/kr-claude-help.txt`, then grep it for `-k`. Record the verbatim output.
   - **No collision** (the plan-phase result — spec.md §A.6) → proceed with `-k` as mapped.
   - **Collision found** → stop and surface it as a blocker report. Do not silently pick a different letter; the mapping was user-confirmed and changing it is a user decision.
4. Capture the bare-`-f` baseline over the six contract documents on both sides (`8` expected) — the denominator for AC-KR-026.
5. Capture the `AC-FM-` count baseline in the test corpus — the denominator for AC-KR-013.

**Exit**: four baseline artifacts recorded (grep, six-pair diff, `-f` count, `AC-FM-` count), `-k` re-confirmed.

### M1 — Go rename

Priority: **High**. Independent of M2.

1. `internal/config/envkeys.go`: `EnvMoaiFactory` → `EnvMoaiKanban` = `"MOAI_KANBAN"`, `EnvMoaiFactorySpec` → `EnvMoaiKanbanSpec` = `"MOAI_KANBAN_SPEC"`; update both doc comments.
2. `git mv internal/factory internal/kanban`; update the package clause in `record.go` and `revision.go`, the package doc comment, and `stateDirSegments` to `{".moai", "state", "kanban"}`.
3. `git mv internal/cli/factory.go internal/cli/kanban.go`; rename the seven identifiers of REQ-KR-005 and the three flag/sentinel constants. Leave `captureEnvState` alone.
4. Update the call sites: `internal/cli/cc.go`, `glm.go`, `cg.go`, `launcher_blockcap_infinite.go` — including the `internal/factory` import path in `cc.go`, `glm.go`, and `cc_test.go`.
5. Update the `@MX:REASON` comment in the renamed file, which names `MOAI_FACTORY` in prose.
6. Rename test function names (REQ-KR-011) and mode prose; leave every `AC-FM-*` identifier and every assertion untouched (REQ-KR-012, REQ-KR-013).

**Exit**: `go build ./... && go test ./...` green; `git diff --stat` shows no assertion-line changes beyond renames.

### M2 — Harness documentation rename

Priority: **High**. Independent of M1; may run in either order.

Template side first (REQ-KR-017), then the local counterpart, holding each pair's measured delta.

1. `git mv` the contract doc to `workflows/kanban.md` on both sides.
2. Apply the rename inside it: title, mode prose, `--kanban` / `-k`, `kanban_chain`, `KANBAN_MODE_UNSUPPORTED_BACKEND`, `.moai/state/kanban/`, `MOAI_KANBAN`.
3. Update the five sibling documents at their **nine edit locations**: `workflows/moai.md` line ~206 (chaining-policy contract bullet, which also carries the `workflows/factory.md` path reference), `workflows/run.md` line ~52 (routing table), `workflows/run/mode-orchestration.md` lines ~82/~84/~108 (Verify Exit Gate heading and body), `workflows/sync/quality-gates-quality.md` lines ~112/~114/~129 (Step 0.55.0 heading and disclosure clause), `rules/moai/workflow/goal-directive.md` line ~15 (block-cap trigger 2).

   > **Nine edit locations ≠ eight grep matches.** The `quality-gates-quality.md` line ~112 heading must be edited but contains no `factory` token, so the AC-KR-015 positive control is **8**, not 9. The two figures count different things and neither is derived from the other; do not reconcile them by adjusting one.

4. Sweep the bare `-f` residue in the same pass (REQ-KR-025): 8 occurrences across `workflows/factory.md` ×2, `workflows/moai.md` ×1, `goal-directive.md` ×1, mirrored on the template side. The token grep cannot see these.
5. Re-run the six-pair `diff` and compare against the M0 baseline under `factory`→`kanban` substitution.

**Exit**: delta comparison clean; no new content added to or removed from any template file beyond the renamed tokens; bare-`-f` count at 0.

### M2b — Project documentation

Priority: **Medium**. Depends on M1 (the package path must already be renamed for the documents to describe the tree accurately).

1. `.moai/project/codemaps/modules.md` lines 157, 158, 161, 246: the `### internal/factory` section heading → `### internal/kanban`, the role line's `moai cc -f` / `moai glm -f` → `-k`, the entry-point line's `internal/cli/factory.go` → `internal/cli/kanban.go`, and the line-246 prose naming `internal/factory`.
2. `.moai/project/structure.md` line 139: the package-count paragraph naming `internal/factory`.

Neither file is template-mirrored, so no delta comparison applies. Both are hand-authored Korean prose — edit the tokens in place; do not regenerate the documents (a regeneration would rewrite the surrounding measurement narrative, which is out of scope).

All four lines above must be edited, not just the two the token grep sees. `$TOK` matches `modules.md` at lines 157 and 246 only: line 158 carries the mode phrase in Korean (`Factory 모드`) and line 161 carries `internal/cli/factory.go`, and neither form is in the pattern (`research.md` §H.4). `AC-KR-028`'s third command — a bare-word grep bounded to these two files, baseline 5 lines — is what catches a partial edit; the token grep alone would read clean with 158 and 161 stale.

**Exit**: the §D.1 grep over `.moai/project/` returns 0 (baseline: 2 files, 3 lines) **and** `grep -niI factory` over the two files returns 0 (baseline: 5 lines).

### M3 — Build and catalog

Priority: **High**. Depends on M2 (and on M1 only insofar as both must be in the tree before the final build).

1. `make build`.
2. `git diff --stat internal/template/catalog.yaml` — the `moai` skill's `hash:` is expected to change (§B-3). Commit it.
3. Confirm no other generated artifact drifted.

**Exit**: `catalog.yaml` staged with a changed `moai` hash; build green.

### M4 — Verification sweep

Priority: **High**. Depends on M1-M3 (and M2b).

Single-turn parallel batch. Every command that decides a criterion writes to a log and reads `$?` before any pipe (§B-6):

```bash
go build ./... > /tmp/kr-build.log 2>&1; echo "build-exit=$?"
go test ./... > /tmp/kr-test.log 2>&1; echo "test-exit=$?"; grep -c '^FAIL' /tmp/kr-test.log
grep -rlniIE "$TOK" internal/ .claude/ .moai/project/ | wc -l                    # expect 0 (baseline 28)
go test ./internal/template/... > /tmp/kr-template.log 2>&1; echo "tmpl-exit=$?"
for f in <six pairs>; do diff .claude/$f internal/template/templates/.claude/$f; done
grep -c 'SPEC-KANBAN-RENAME-001' internal/template/templates/.claude/skills/moai/workflows/kanban.md  # expect 0
git diff --stat d39e3cdc6..HEAD -- docs-site/ .moai/specs/SPEC-FACTORY-MODE-001/  # expect empty
git diff --stat d39e3cdc6..HEAD -- internal/template/catalog.yaml                 # expect non-empty (AC-KR-020)
grep -niI factory .moai/project/codemaps/modules.md .moai/project/structure.md | wc -l  # expect 0 (baseline 5)
```

The `<six pairs>` diff loop and the bare-`-f` sweep of AC-KR-026 run in the same batch. Note the asymmetry in the two `git diff` lines: the exclusion checks and the catalog check are **both** anchored to `d39e3cdc6..HEAD`, because by M4 every milestone is committed and a ref-less diff is empty either way — which reads as PASS for an exclusion check and as FAIL for the catalog check, but is uninformative in both directions. M3 step 2's bare diff is a different moment and stays bare.

**Exit**: the §E 5-section report, with every row attributed.

---

## §G. Anti-patterns for this SPEC

- **AP-1 — parity by verbatim copy.** Copying a local `.claude/` file over its template counterpart to "fix" a diff. Re-introduces §25-forbidden content into template source (§B-1).
- **AP-2 — bare-word grep-zero.** Adopting `grep -ri factory` as the completion criterion and then hand-maintaining a 110-file allowlist (§B-2).
- **AP-3 — narrow test run.** Running `go test ./internal/cli/... ./internal/kanban/...` and reporting it as the suite. The template guards live in `internal/template` and are exactly what a rename touching template source can break.
- **AP-4 — renaming `AC-FM-*`.** They cite a closed SPEC's acceptance criteria (REQ-KR-012).
- **AP-5 — assertion drift.** Adjusting a test expectation "while I was in there". Behavior preservation is the claim; a changed assertion makes it unverifiable.
- **AP-6 — forgetting `catalog.yaml`.** The failure is silent locally and surfaces as a CI parity failure (§B-3).
- **AP-8 — reading `$?` after a pipe.** `go test ./... 2>&1 | tail -20; echo "exit=$?"` reports `tail`'s status. A fully red suite prints `exit=0`, and the "no FAIL lines" fallback does not rescue it because `tail -20` hides ~95 of 115 packages (§B-6).
- **AP-9 — a ref-less `git diff` as an exclusion check.** Empty by construction after the work is committed, so it passes whether or not the excluded path was touched. Anchor to `d39e3cdc6..HEAD` (§B-6).
- **AP-11 — a `go test -run` keyed on a name that does not exist yet.** Exits 0 and prints `PASS` over zero tests, so a criterion built on a post-rename test name is satisfied by not renaming the test (§B-6). Pair every `-run` with a name-existence grep and an absent-`[no tests to run]` assertion.
- **AP-12 — a negative check with no positive control.** "Returns zero" is only evidence when the same command returned non-zero at baseline. `AC-KR-025`'s help-text `factory` grep returned zero **before** the rename too — the help string never named the flag — so it decided nothing and its missing control is what hid that. Every negative criterion here states its measured baseline (acceptance.md §A.1).
- **AP-10 — treating `.moai/project/` like `.moai/specs/`.** The specs directory is preserved verbatim; the project directory is in scope (§B-5). Excluding both would leave two documents describing a package that no longer exists.
- **AP-7 — writing outside the worktree.** This SPEC is authored and implemented in `/Users/goos/.moai/worktrees/kanban`. The primary checkout is diverged and carries unrelated uncommitted work.

---

## §H. Cross-references

- `spec.md` §A.4 (mirror classification), §D.1 (token grep), §C (exclusions).
- `acceptance.md` — the AC matrix.
- `CLAUDE.local.md` §2, §14, §25.
