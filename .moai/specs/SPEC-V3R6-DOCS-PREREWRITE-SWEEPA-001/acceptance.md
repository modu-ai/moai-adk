# SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001 — Acceptance Criteria

## A. Conventions

### A.1 Pin, patterns, and pathspecs

Document-level pin: tree commit `2bcce5d9dc4c5f57bb98bcfa553e0061fa373681`. It binds every RED-now cell, control, and evidence entry in §C that carries no pin of its own. The commands in §B judge the sweep tip, so they read `HEAD` of the card branch, not the pin.

The patterns and pathspecs below are defined once and used by name. `E01` is the umbrella's family F1 pattern (the card's contract). `W01` is `E01` with `[[:space:]]+` in place of the single space; it is the mutant guard, because a stale citation respelled with two spaces passes `E01` and still tells the reader to run an absent verb.

```
E01 = (^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)
W01 = (^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)
P-BAS = 'docs-site/content/*/advanced/bas-navigator.md'
P-FAC = 'docs-site/content/*/advanced/factory-mode.md'
P-PRO = 'docs-site/content/*/advanced/profile-matrix.md'
```

### A.2 Command form

- Every command runs from the card worktree root. `git` stands at the head of a plain pipeline or of a `&&` / `||` chain, never inside `$(...)` or `<(...)`, because the worktree-session guard refuses those forms.
- A verdict word (`MATCH` / `NOMATCH`, `UNCHANGED` / `CHANGED`) comes from `&& echo ... || echo ...`, because the Bash tool does not display the exit status of a command that prints nothing. A count or a path list is read directly.
- Commands live in fenced blocks, never in table cells, so an alternation `|` is a regex alternation.
- The left end `main` in `main...HEAD` is deliberately not pinned: the claim is about this card's own contribution, so the left end is the merge-base with the absorbed local `main`, recomputed at read time. It cannot be used after the card has merged into `main`.

### A.3 Classification

AC-PRC-001-001 and AC-PRC-001-002 are release-blocking and flip from red to green. AC-PRC-001-003 to AC-PRC-001-005 are preservation criteria: green at the pin by construction, so each carries a control, a command run on a known historical input that turns the same form red.

## B. Acceptance criteria

| AC | Requirement | Depends on | Class | RED-now or control | Green path |
|----|-------------|------------|-------|--------------------|------------|
| AC-PRC-001-001 | REQ-PRC-001-001 | Q13, Q17 | release-blocking | E-SWA-2, E-SWA-3 (4 matches, RED) | M2 |
| AC-PRC-001-002 | REQ-PRC-001-002 | Q-SWA-1, Q13, Q17 | release-blocking | E-SWA-2, E-SWA-3 (8 matches, RED under branch a); E-SWA-11 (control, branch b) | M3 |
| AC-PRC-001-003 | REQ-PRC-001-003 | Q17 | preservation | E-SWA-12 (control) | M2, M3 |
| AC-PRC-001-004 | REQ-PRC-001-004 | none | preservation | E-SWA-7 (control) | M5 |
| AC-PRC-001-005 | REQ-PRC-001-004 | none | preservation | E-SWA-13 (control) | M4, M5 |

### AC-PRC-001-001 — the Class C lines are gone

AC-PRC-001-001 maps REQ-PRC-001-001.

- **Given** the sweep tip is `HEAD` of the card branch and Q13 carries a verdict,
- **When** these three commands run,
- **Then** the first two print `NOMATCH` and the third prints exactly four paths.

```
git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/bas-navigator.md' && echo MATCH || echo NOMATCH
git grep -q -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/bas-navigator.md' && echo MATCH || echo NOMATCH
git ls-files -- 'docs-site/content/*/advanced/bas-navigator.md'
```

The third command is the swept-set control: a pass over an empty set asserts nothing, and deleting the files would otherwise satisfy the first two. RED-now at the pin: E-SWA-2 prints four `bas-navigator.md` lines ending `:1`, and E-SWA-4 prints `MATCH`.

### AC-PRC-001-002 — the Class N lines match the Q-SWA-1 verdict

AC-PRC-001-002 maps REQ-PRC-001-002.

**Branch (a): Q-SWA-1 treats the eight lines as stale citations.**

- **Given** the verdict is (a),
- **When** these three commands run,
- **Then** the first two print `NOMATCH` and the third prints exactly eight paths.

```
git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md' && echo MATCH || echo NOMATCH
git grep -q -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md' && echo MATCH || echo NOMATCH
git ls-files -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'
```

RED-now at the pin: E-SWA-2 prints eight lines for these two pages, each ending `:1`.

**Branch (b): Q-SWA-1 treats the eight lines as retained retirement notices.**

- **Given** the verdict is (b),
- **When** these two commands run,
- **Then** the first prints exactly eight lines, each ending `:1`, and the second prints `UNCHANGED`.

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'
git diff --quiet main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md' && echo UNCHANGED || echo CHANGED
```

Branch (b) is true at the pin. E-SWA-11 is its control: the second command prints `CHANGED` on a commit that does change these pages.

### AC-PRC-001-003 — the three pages change in all four locales or in none

AC-PRC-001-003 maps REQ-PRC-001-003.

- **Given** Q17 carries a verdict that the edits land in all four locales in one change,
- **When** the three commands run, one per page,
- **Then** each prints either nothing or exactly four paths (one per locale).

```
git diff --name-only main...HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'
git diff --name-only main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md'
git diff --name-only main...HEAD -- 'docs-site/content/*/advanced/profile-matrix.md'
```

Control: the same form on commit 609ac419eb, which changed only the en copy of `factory-mode.md`, prints one path, which this criterion rejects (E-SWA-12). The repository's own parity gate is a confirmatory check, not the criterion: `bash scripts/docs-i18n-check.sh | rg -n "^(Errors|Warnings): +0$"` prints two lines. It prints those two lines at the pin (E-SWA-9), so on its own it cannot go red for this change.

### AC-PRC-001-004 — nothing outside the twelve files changes

AC-PRC-001-004 maps REQ-PRC-001-004.

- **Given** the sweep tip,
- **When** this command runs,
- **Then** it prints nothing.

```
git diff --name-only main...HEAD -- . ':!docs-site/content/*/advanced/bas-navigator.md' ':!docs-site/content/*/advanced/factory-mode.md' ':!docs-site/content/*/advanced/profile-matrix.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001' ':!.moai/reports/t1721'
```

Control (E-SWA-7): the same exclusion form over the range of commit 84f9387535 prints nine paths, so the form does print when something outside the list changes. At plan time the command prints nothing, which is the expected state before any edit.

### AC-PRC-001-005 — no line owned by another card is touched

AC-PRC-001-005 maps REQ-PRC-001-004.

- **Given** the sweep tip,
- **When** these four commands run,
- **Then** the first prints nothing; the second prints nothing under Q-SWA-1 (b) or `4` under (a); the third prints nothing under (b) or four numstat lines whose second column is `1` under (a); the fourth prints nothing under (b) or `4` under (a).

```
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^[-+].*(--factory-run|--clear-policy)"
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^@@"
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md'
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^-[^-]"
```

The first command keeps the sweep off the `--factory-run` and `--clear-policy` lines that card t1726 owns. The other three confine the sweep's edit to the one "Refused forms" row per locale, which is what keeps it clear of the lines card t1863 adds next to it. Controls (E-SWA-13): over the range of commit 84f9387535 the first command prints `8`; over the range of commit 449b1c9931 the hunk count prints `8`, the numstat prints `2 2` per file, and the removed-line count prints `8`.

## C. Evidence ledger

Each entry gives the command, the observed output (verbatim unless marked), and the tree it read. "No failure shown" means the Bash tool displayed no error status, which on a command that prints nothing does not distinguish exit 0 from exit 1; where that matters, the entry uses a verdict-word wrapper.

### E-SWA-0 — pin and tool provenance

```
git rev-parse HEAD                    -> 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681
git rev-parse main                    -> 2bcce5d9dc4c5f57bb98bcfa553e0061fa373681
moai version (version line)           -> moai_cp/20261007_195927   moai_cp/20261007_195927-1888-g7705b365e2   built 2026-10-18T10:45:08Z
git merge-base --is-ancestor 7705b365e2 2bcce5d9dc && echo ANCESTOR || echo NOT-ANCESTOR   -> ANCESTOR
git rev-list --count 2bcce5d9dc..7705b365e2   -> 0
git rev-list --count 7705b365e2..2bcce5d9dc   -> 48
```

The installed build is a strict ancestor of the tree HEAD, 48 commits behind. No Go build was made; CLI claims in this SPEC are attributed to "installed build 7705b365e2 against tree 2bcce5d9dc" (plan.md §A.4).

### E-SWA-1 — the three pathspecs expand to twelve paths

Command: `git ls-files -- 'docs-site/content/*/advanced/bas-navigator.md' 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'` at the pin. Output, 12 paths: `docs-site/content/{en,ja,ko,zh}/advanced/{bas-navigator,factory-mode,profile-matrix}.md`, listed per page and locale in the order en, ja, ko, zh.

### E-SWA-2 — RED-now, pattern E01 over the twelve files

Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc -- 'docs-site/content/*/advanced/bas-navigator.md' 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'`. Exit 0 (output present, no failure shown). Output:

```
2bcce5d9dc:docs-site/content/en/advanced/bas-navigator.md:1
2bcce5d9dc:docs-site/content/en/advanced/factory-mode.md:1
2bcce5d9dc:docs-site/content/en/advanced/profile-matrix.md:1
2bcce5d9dc:docs-site/content/ja/advanced/bas-navigator.md:1
2bcce5d9dc:docs-site/content/ja/advanced/factory-mode.md:1
2bcce5d9dc:docs-site/content/ja/advanced/profile-matrix.md:1
2bcce5d9dc:docs-site/content/ko/advanced/bas-navigator.md:1
2bcce5d9dc:docs-site/content/ko/advanced/factory-mode.md:1
2bcce5d9dc:docs-site/content/ko/advanced/profile-matrix.md:1
2bcce5d9dc:docs-site/content/zh/advanced/bas-navigator.md:1
2bcce5d9dc:docs-site/content/zh/advanced/factory-mode.md:1
2bcce5d9dc:docs-site/content/zh/advanced/profile-matrix.md:1
```

Why red: each of the twelve files carries one line that matches `E01`. Four of them are Class C (`bas-navigator.md`) and eight Class N (`factory-mode.md`, `profile-matrix.md`); the per-line inventory is plan.md §A.2. A green path exists because the sweep edits exactly those lines.

### E-SWA-3 — RED-now, pattern W01 over the twelve files

Command: the same as E-SWA-2 with `W01` in place of `E01`. Output: the same twelve lines, line for line. The mutant guard therefore adds no match today and closes only the respelling route.

### E-SWA-4 — RED-now verdict word for AC-PRC-001-001

Command: `git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc -- 'docs-site/content/*/advanced/bas-navigator.md' && echo MATCH || echo NOMATCH`. Output: `MATCH`.

### E-SWA-5 — the other umbrella families are silent on the twelve files

Command: `git grep -q -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)|(^|[^/])moai (run|sync) SPEC-ID|/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 2bcce5d9dc -- 'docs-site/content/*/advanced/bas-navigator.md' 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md' && echo MATCH || echo NOMATCH`. Output: `NOMATCH`. Control, the same pattern over `docs-site/content/en/guides/mcp-server.md` at the same tree: `MATCH`, so the combined pattern fires where the families do occur.

### E-SWA-6 — CLI surface and source corroboration

Installed build 7705b365e2 against tree 2bcce5d9dc. Each probe's stderr is the whole output.

```
moai codemaps --help  -> exit 1   Unknown command "codemaps" for "moai".
moai gpt --help       -> exit 1   Unknown command "gpt" for "moai"   (suggestions: glm, gate, gtd, pr)
moai model --help     -> exit 1   Unknown command "model" for "moai"   (suggestion: codex)
moai audit --help     -> exit 1   Unknown command "audit" for "moai".
moai cg --help        -> exit 1   moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration
moai migrate --help   -> lists    cg [--flags]   Preview migration of legacy CG teammate roles
moai navigator-sync --help -> help text printed ("Writes a single nav-graph.json atomically."); the verb is not in the `moai --help` listing
moai --help | rg -c "^ +(cg|gpt|audit|codemaps|model)( |$)"      -> no output
moai --help | rg -c "^ +(cc|glm|codex|migrate|spec)( |$)"        -> 5   (control: the listing pattern fires on listed verbs)
```

Source at the pin: `git grep -n -E "Use:[[:space:]]*[\"\`](cg|gpt|audit|codemaps|model)([[:space:]\"\`]|$)" 2bcce5d9dc -- internal cmd` prints four lines: `internal/cli/cg.go:13` (`Use: "cg"`, `Hidden: true`, "Retired CG launcher"), `internal/cli/codex_audit_launch.go:880` (`Use: "audit <role>"`), `internal/cli/migrate_cg.go:22` (`Use: "cg"`, "Preview migration of legacy CG teammate roles"), and `internal/cli/spec_audit.go:64` (`Use: "audit"`). Neither `audit` line is a top-level verb: they are sub-verbs of `codex` and `spec`. `internal/cli/navigator_sync.go` declares `Use: "navigator-sync"` with `Hidden: true`. `git diff --name-only 7705b365e2 2bcce5d9dc -- internal/cli cmd` lists 15 files, none of them `cg.go`, `migrate_cg.go`, or `navigator_sync.go`.

Codemaps equivalence: `git ls-files -- .claude/skills/moai/workflows/codemaps.md .claude/commands/moai/codemaps.md` prints both paths. `git grep -c "nav-graph" 2bcce5d9dc -- <those two paths>` prints nothing, and `git grep -c -i "codemaps" 2bcce5d9dc -- .claude/skills/moai/workflows/codemaps.md` prints `22`, so the zero is a real zero on a file that is read. The command's `description:` line reads, in Korean, "scan the codebase and generate architecture docs in codemaps/".

### E-SWA-7 — control for AC-PRC-001-004

Command: `git diff --name-only 84f9387535~1...84f9387535 -- . ':!docs-site/content/*/advanced/bas-navigator.md' ':!docs-site/content/*/advanced/factory-mode.md' ':!docs-site/content/*/advanced/profile-matrix.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPA-001' ':!.moai/reports/t1721'`. Output, nine paths:

```
.claude/skills/moai/workflows/factory.md
.moai/reports/t1863/progress.md
README.ja.md
README.ko.md
README.md
README.zh.md
internal/cli/factory_leader_orphan_test.go
internal/template/catalog.yaml
internal/template/templates/.claude/skills/moai/workflows/factory.md
```

The same command with `main...HEAD` at plan time prints nothing (local `main` equals HEAD, E-SWA-0).

### E-SWA-8 — sequencing facts

```
git merge-base --is-ancestor 84f9387535 main && echo ANCESTOR-OF-MAIN || echo NOT-ANCESTOR-OF-MAIN   -> NOT-ANCESTOR-OF-MAIN
git merge-base --is-ancestor 84f9387535 HEAD && echo ANCESTOR-OF-HEAD || echo NOT-ANCESTOR-OF-HEAD   -> NOT-ANCESTOR-OF-HEAD
git diff -U0 84f9387535~1 84f9387535 -- 'docs-site/content/en/advanced/factory-mode.md'  (hunk header)  -> @@ -57,0 +58,6 @@
```

The en hunk adds the six-line section "Leader start over an orphaned run" after the blank line 57 that follows the "Refused forms" table. Churn: `git log -n 25 --name-only` over the three pages, at the pin, lists 18 commits touching `factory-mode.md`, 5 touching `profile-matrix.md`, none touching `bas-navigator.md`, and 3 merges listing no file.

### E-SWA-9 — parity baseline

Command: `bash scripts/docs-i18n-check.sh` at the pin; exit 0. Tail of the output:

```
=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

### E-SWA-10 — decision-board check

Commands: `moai decision read --scope card:t1655 --project-root /Users/goos/MoAI/moai-adk-go` (134 records) and `moai decision read --scope card:t1721 --project-root /Users/goos/MoAI/moai-adk-go` (122 records), each output saved to a file. `rg -c "Q13|Q17"` on each file printed nothing. Controls: `rg -c "id=d-20261014T124444Z-1aba"` on the t1655 file matched, and `rg -c "id=d-20261018T170912Z-14f9"` on the t1721 file printed `1`.

### E-SWA-11 — control for AC-PRC-001-002 branch (b)

Command: `git diff --quiet 84f9387535~1 84f9387535 -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md' && echo UNCHANGED || echo CHANGED`. Output: `CHANGED`. The same form over `main...HEAD` at plan time prints `UNCHANGED`.

### E-SWA-12 — control for AC-PRC-001-003

Command: `git diff --name-only 609ac419eb~1 609ac419eb -- 'docs-site/content/*/advanced/factory-mode.md'`. Output: `docs-site/content/en/advanced/factory-mode.md` (one path, neither none nor four).

### E-SWA-13 — controls for AC-PRC-001-005

```
git diff -U0 84f9387535~1 84f9387535 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^[-+].*(--factory-run|--clear-policy)"   -> 8
git diff -U0 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^@@"                                  -> 8
git diff --numstat 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md'                                         -> 2 2 <path>   (one line per locale)
git diff -U0 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^-[^-]"                               -> 8
```

At plan time the same four commands over `main...HEAD` print nothing, the expected state before any edit.

## D. Gaps and residual risk

**Gaps (not observed in this pass):**

- **G1: the card's named sources are absent.** The card cites `split-proposal.md` and `child-cards.md` of the t1655 tree. The t1655 worktree is gone: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1655` and `.moai/reports/t1655` do not exist, and a search of `.moai/**/reports` found `split-proposal.md` only under the t1631 and t1714/t1715 report directories, none of them t1655's, and no `child-cards.md` at all. Nothing in this SPEC is taken from those two files; the file list, the E-01 pattern, and Q13 and Q17 come from the umbrella's committed spec, acceptance, and decision-index (rows at commit bdb7b764 and revision 0.3.1).
- **G2: no build of the tree.** The CLI probes ran on an installed build 48 commits behind the tree (E-SWA-0). The source read at the pin corroborates the registrations but is not a run of the tree's binary.
- **G3: forms outside the four families.** The three pages cite other `moai` forms (for example `moai factory runs`, `moai factory decide`, `moai profile setup`, `moai todo next`, `moai chain`, `moai paths`). Only the top-level listing of `moai --help` was compared with them; their sub-verbs were not probed, and no claim is made about them.
- **G4: no trial merge with card t1863.** E-SWA-8 shows the hunks are disjoint and 84f9387535 is not on `main`; no merge was attempted.
- **G5: GATE card state.** The lane's brief says card t1717 is on plan-audit hold. That state was not re-measured, and nothing in this SPEC depends on it.
- **G6: the i18n rules document.** The umbrella cites `.moai/docs/docs-site-i18n-rules.md` lines 61 and 65 for the four-locale rule. This pass did not re-read that document.

**Residual risk:** a green E01 and W01 over the twelve files proves the matched spellings are gone, not that the page still reads coherently (plan.md R2) or that a respelling outside both patterns, such as a differently quoted verb name, is absent. The reviewer reads the edited lines in all four locales.

## E. Definition of Done

- AC-PRC-001-001 and AC-PRC-001-002 print their expected output at the sweep tip, and their results are carried into `.moai/reports/t1721/verdict.md` with the command and the deciding lines.
- AC-PRC-001-003 to AC-PRC-001-005 print their expected output at the sweep tip.
- Q17 and Q-SWA-1 carry an operator verdict, and Q13 carries its verdict, before the first file edit.
- The umbrella's closing measurement is not part of this Definition of Done; it is the umbrella's M6.
