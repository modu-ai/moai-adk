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
- A quiet check (`git grep -q`, `git diff --quiet`) is followed by `; echo "exit=$?"`, because the Bash tool does not display the exit status of a command that prints nothing. The three outcomes stay distinct: `exit=0` means a match (or a difference), `exit=1` means none, and `exit=128` or any other value means git failed. A failure is never read as "no match". A count or a path list is read directly.
- Commands live in fenced blocks, never in table cells, so an alternation `|` is a regex alternation.
- The left end `main` in `main...HEAD` is deliberately not pinned: the claim is about this card's own contribution, so the left end is the merge-base with the absorbed local `main`, recomputed at read time. It cannot be used after the card has merged into `main`.

### A.3 Classification

AC-PRC-001-001 is release-blocking and flips from red to green. AC-PRC-001-002 is release-blocking under branch (a), where it flips from red to green, and a regression-guard under branch (b), where it is true at the pin. AC-PRC-001-003 and AC-PRC-001-004 are preservation criteria: green at the pin by construction, so each carries a control, a command run on a known historical input that turns the same form red. AC-PRC-001-005 mixes both. Its command 2 (the bas-navigator numstat bound, which requires a non-empty edit within the edit unit) is the release-blocking part: it is red at the pin and flips to green at M2, and E-SWA-15 is its RED-now cell. Commands 1 and 6 are green at the pin and are preservation checks (command 6, the dangling-step check, prints `moai codemaps` at the pin). Commands 3 to 5 (the Class N bounds) depend on the Q-SWA-1 branch: under (b) they expect nothing, so they are green at the pin and stay green (preservation); under (a) they expect four lines, four lines and the count 4, print nothing at the pin, and are therefore red and flip to green at M3. Their controls are E-SWA-13 and E-SWA-14. The flips are command 2 at M2 and, under (a) only, commands 3 to 5 at M3.

## B. Acceptance criteria

| AC | Requirement | Depends on | Class | RED-now or control | Green path |
|----|-------------|------------|-------|--------------------|------------|
| AC-PRC-001-001 | REQ-PRC-001-001 | Q13, Q17 | release-blocking | E-SWA-2, E-SWA-3 (4 matches, RED) | M2 |
| AC-PRC-001-002 | REQ-PRC-001-002 | Q-SWA-1, Q17 | release-blocking under (a); regression-guard under (b) | E-SWA-2, E-SWA-3 (8 matches, RED under branch a); E-SWA-11 (control, branch b) | M3 |
| AC-PRC-001-003 | REQ-PRC-001-003 | Q17 | preservation | E-SWA-12 (control) | M2, M3 |
| AC-PRC-001-004 | REQ-PRC-001-004 | none | preservation | E-SWA-7 (control) | M5 |
| AC-PRC-001-005 | REQ-PRC-001-004 | Q13, Q17 (command 2); Q-SWA-1 branch (a) (commands 3 to 5) | release-blocking (command 2; commands 3 to 5 under (a)); preservation (commands 1 and 6, and commands 3 to 5 under (b)) | E-SWA-15 (RED-now, commands 2 to 5); E-SWA-13, E-SWA-14 (controls) | M2, M3, M4, M5 |

### AC-PRC-001-001 — the Class C lines are gone

AC-PRC-001-001 maps REQ-PRC-001-001.

- **Given** the sweep tip is `HEAD` of the card branch and Q13 carries an operator verdict,
- **When** these three commands run,
- **Then** the first two print `exit=1` and the third prints exactly four paths.

```
git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'; echo "exit=$?"
git grep -q -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'; echo "exit=$?"
git ls-files -- 'docs-site/content/*/advanced/bas-navigator.md'
```

The third command is the swept-set control: a pass over an empty set asserts nothing. It does not stop a page from being emptied; the per-file bounds of AC-PRC-001-005 do. RED-now at the pin: E-SWA-2 prints four `bas-navigator.md` lines ending `:1`, and E-SWA-4 prints `exit=0`.

### AC-PRC-001-002 — the Class N lines match the Q-SWA-1 verdict

AC-PRC-001-002 maps REQ-PRC-001-002.

**Branch (a): Q-SWA-1 treats the eight lines as stale citations.**

- **Given** the verdict is (a) (a1 or a2),
- **When** these three commands run,
- **Then** the first two print `exit=1` and the third prints exactly eight paths.

```
git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"
git grep -q -E "(^|[^/])moai[[:space:]]+(cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"
git ls-files -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'
```

RED-now at the pin: E-SWA-2 prints eight lines for these two pages, each ending `:1`.

**Branch (b): Q-SWA-1 treats the eight lines as retained retirement notices.**

- **Given** the verdict is (b),
- **When** these two commands run,
- **Then** the first prints exactly eight lines, each ending `:1`, and the second prints `exit=0` (no difference).

```
git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'
git diff --quiet main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"
```

Branch (b) is true at the pin. E-SWA-11 is its control: the second command prints `exit=1` on a commit that does change these pages. Under (a), AC-PRC-001-005 bounds how much of each file may change.

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

### AC-PRC-001-005 — every change stays inside its edit unit

AC-PRC-001-005 maps REQ-PRC-001-004. The edit units and their bounds are the table in plan.md §A.2; this criterion measures them for all twelve files, so a page cannot be emptied, and deleting only the target rows cannot pass.

- **Given** the sweep tip and the Q-SWA-1 verdict,
- **When** these commands run (the last one once per locale: ko, en, ja, zh),
- **Then** command 1 prints nothing; command 2 prints four numstat lines, one per locale, each with removed in 2 to 4 and added in 0 to 3 (empty output is red: the Class C pair must be gone); command 3 prints nothing under Q-SWA-1 (b), or under (a) four lines each with removed `1` and added `0` or `1`; command 4 prints nothing under (b), or under (a) four lines each with removed in 1 to 4 and added in 0 to 4; command 5 prints nothing under (b) or `4` under (a); command 6 prints a line that does not begin with `# 3)`.

```
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^[-+].*(--factory-run|--clear-policy)"
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md'
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/profile-matrix.md'
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^@@"
git show HEAD:docs-site/content/<locale>/advanced/bas-navigator.md | awk '/^```bash/ && !b {b=1; next} b && /^```/ {print last; exit} b && NF {last=$0}'
```

Command 1 keeps the sweep off the `--factory-run` and `--clear-policy` lines that card t1726 owns. Commands 3 and 5 confine the factory-mode edit to the one "Refused forms" row per locale, which keeps it clear of the lines card t1863 adds next to it. Commands 2 to 4 are the per-file bounds. Command 6 prints the last non-blank line of the page's first bash block; a step-3 comment left without its command prints `# 3) ...` and fails, while a deleted or replaced pair does not. Mutants this closes: blanking any of the twelve pages exceeds the removed-line bound; deleting only the target lines in `bas-navigator.md` removes the comment or the command alone, which leaves `# 3)` last in the block; leaving a page untouched fails command 2 (bas-navigator) or the AC-PRC-001-002 absence check. Reading the line-65 sentence stays a reviewer read.

Controls (E-SWA-13, E-SWA-14): over the range of commit 84f9387535 command 1 prints `8`; over the range of commit 449b1c9931 command 5 prints `8` and command 3 prints `2 2` per file (outside the bound); over the range of commit 948a565e8f command 2 prints `1 1` per file (below the lower bound, so empty and one-line edits are red); over the range of commit 5f4f199efd command 4 prints `113 131`, `38 111`, `79 101`, `38 111` (above the upper bound); fed a block whose last line is a bare step-3 comment, command 6 prints that comment.

## C. Evidence ledger

Each entry gives the command, the observed output (verbatim unless marked), and the tree it read. "No failure shown" means the Bash tool displayed no error status, which on a command that prints nothing does not distinguish exit 0 from exit 1; where that matters, the entry uses the explicit `; echo "exit=$?"` form of §A.2 (E-SWA-0 and E-SWA-8 were run earlier with an `&& echo ... || echo ...` wrapper and are unchanged).

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

Command: `git grep -q -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" 2bcce5d9dc -- 'docs-site/content/*/advanced/bas-navigator.md'; echo "exit=$?"`. Output: `exit=0` (a match).

### E-SWA-5 — the other umbrella families are silent on the twelve files

Command: `git grep -q -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)|(^|[^/])moai (run|sync) SPEC-ID|/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" 2bcce5d9dc -- 'docs-site/content/*/advanced/bas-navigator.md' 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"`. Output: `exit=1` (no match). Control, the E-02 part of the same pattern over `docs-site/content/en/guides/mcp-server.md` at the same tree: `exit=0`, so the combined pattern fires where the families do occur.

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

Command: `git diff --quiet 84f9387535~1 84f9387535 -- 'docs-site/content/*/advanced/factory-mode.md' 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"`. Output: `exit=1` (the pages differ). The same form over `main...HEAD` at plan time prints `exit=0`.

### E-SWA-12 — control for AC-PRC-001-003

Command: `git diff --name-only 609ac419eb~1 609ac419eb -- 'docs-site/content/*/advanced/factory-mode.md'`. Output: `docs-site/content/en/advanced/factory-mode.md` (one path, neither none nor four).

### E-SWA-13 — controls for AC-PRC-001-005

```
git diff -U0 84f9387535~1 84f9387535 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^[-+].*(--factory-run|--clear-policy)"   -> 8
git diff -U0 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^@@"                                  -> 8
git diff --numstat 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md'                                         -> 2 2 <path>   (one line per locale)
git diff -U0 449b1c9931~1 449b1c9931 -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^-[^-]"                               -> 8
```

At plan time under Q-SWA-1 branch (b) the same four commands over `main...HEAD` print nothing, the expected state before any edit; under branch (a) the status follows the branch-by-branch statement in A.3, the AC-PRC-001-005 row, and E-SWA-15, where commands 3 to 5 are red at plan time.

### E-SWA-14 — controls for the per-file bounds and the dangling-step check

```
git diff --numstat 948a565e8f~1 948a565e8f -- 'docs-site/content/*/advanced/bas-navigator.md'   -> 1 1 <path>   (one line per locale; below the 2-to-4 removed bound)
git diff --numstat 5f4f199efd~1 5f4f199efd -- 'docs-site/content/*/advanced/profile-matrix.md'   -> 113 131 en, 38 111 ja, 79 101 ko, 38 111 zh   (above the 1-to-4 bound)
git show HEAD:docs-site/content/en/advanced/bas-navigator.md | awk '<command 6 of AC-PRC-001-005>'   -> moai codemaps   (at the plan-time tree, last line of the first bash block)
printf '```bash\n# 2) x\ncmd\n\n# 3) Rebuild the graph\n```\n' | awk '<command 6>'   -> # 3) Rebuild the graph   (a dangling step comment is printed)
git grep -n -E "^# 3\)" HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'   -> 12 lines: line 74, 97, 135 in each locale (`# 3)` also opens later blocks, so the check is scoped to the first block)
```

At plan time commands 2 to 4 of AC-PRC-001-005 over `main...HEAD` print nothing. That is red for command 2 under every branch, and red for commands 3 and 4 under Q-SWA-1 (a) but the expected state under (b); see E-SWA-15.

### E-SWA-15 — RED-now cell for AC-PRC-001-005 command 2

Tree: HEAD `640d2471cd78ff43e5b69e355f423620ff9913c9` (the docs-site tree equals the pin 2bcce5d9dc, because the plan commits touch only the SPEC directory; local `main` is the pin). Commands, each followed by `; echo "exit=$?"`, run from the card worktree root:

```
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/bas-navigator.md'; echo "exit=$?"   -> (no stdout)   exit=0
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md'; echo "exit=$?"    -> (no stdout)   exit=0
git diff --numstat main...HEAD -- 'docs-site/content/*/advanced/profile-matrix.md'; echo "exit=$?"  -> (no stdout)   exit=0
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^[-+].*(--factory-run|--clear-policy)"; echo "exit=$?"   -> (no stdout)   exit=1 (rg, no match)
git diff -U0 main...HEAD -- 'docs-site/content/*/advanced/factory-mode.md' | rg -c "^@@"; echo "exit=$?"   -> (no stdout)   exit=1 (rg, no match)
for each locale of bas-navigator.md: git show HEAD:docs-site/content/<locale>/advanced/bas-navigator.md | awk '<command 6>'   -> moai codemaps (ko, en, ja, zh)
```

Reading: command 2 is expected to print four numstat lines after M2 and prints none now, so it is red at the pin; it is red under every Q-SWA-1 branch. Commands 3, 4, and 5 also print nothing at the pin: under (b) that is the expected state (green, preservation); under (a) the Then-clause requires four lines, four lines, and `4`, so they are red at the pin and flip at M3. Command 1 prints nothing, the expected state before any edit; command 6 prints `moai codemaps`, which does not begin with `# 3)`, so it is green at the pin and its red form is the `printf` fixture of E-SWA-14.

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
- AC-PRC-001-003 to AC-PRC-001-005 print their expected output at the sweep tip, including the per-file bounds and the dangling-step check of AC-PRC-001-005.
- Q13, Q17, and Q-SWA-1 each carry an operator verdict before the first file edit (the gate set is stated in decision-index.md, Gate).
- The umbrella's closing measurement is not part of this Definition of Done; it is the umbrella's M6.
