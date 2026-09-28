# SPEC-LOCAL-INSTR-RECEPTION-001 — research

> Authored by card t1290 (plan phase). Every claim below was measured in this worktree
> (`.claude/worktrees/t1290`, branch `WT-agents-local-migration`) unless attributed to an
> external record, which is named with its own tree/tool attribution.

## §A Measured premises

Tree SHA for everything in this section unless noted: `514ac7abe2ee5d721fef5e715ead9f753940cf51`
(short `514ac7abe`).

| # | Premise | Command | Observed |
|---|---|---|---|
| A1 | The contract states the worktree reception negative | `sed -n '250,275p' AGENTS.md` | `:262`: "…import points outside the project and is skipped silently, so a worktree session does not receive `AGENTS.local.md`." |
| A2 | `AGENTS.local.md` is gitignored; `CLAUDE.local.md` is tracked despite its rule | `sed -n '270,280p' .gitignore`; iter4 §2 D1 | `.gitignore:275` `/AGENTS.local.md`, `:276` `/CLAUDE.local.md`; `git ls-files` lists `CLAUDE.local.md` (iter4, tree `a73c78d1a`) |
| A3 | The file does not exist in this tree | `ls AGENTS.local.md` | `No such file or directory` |
| A4 | `CLAUDE.md` imports the file last, and documents the skip | `sed -n '160,175p' CLAUDE.md` | §18: "the user-owned `AGENTS.local.md` (gitignored, never deployed) is imported last"; line 171: `@AGENTS.local.md`; "When the file is absent, or the session runs in a linked worktree where the import points outside the project, Claude Code skips the import silently." |
| A5 | `.worktreeinclude` exists, root AND template mirror, byte-identical 1,055 B | `ls .worktreeinclude internal/template/templates/.worktreeinclude`; Read | Copies gitignored listed files into new worktrees at creation — `claude --worktree`, `moai cc\|glm\|cg -w`, `isolation: worktree`; "tracked files are never duplicated"; creation-time only. Currently lists `.claude/settings.local.json`, `.env`, `.env.*` |
| A6 | `.worktreeinclude` provenance | `git log --oneline -1 -- .worktreeinclude` | `e89d01461` (SPEC-WORKTREE-BRANCH-GUARD-001 PR #1192); documented at `.claude/rules/moai/workflow/worktree-integration.md` § `.worktreeinclude` |
| A7 | The Codex launcher reads local files, `AGENTS.local.md` first | `grep -n codexLocalInstructionName internal/cli/codex_contract.go` | `codex_contract.go:35-36`: `codexLocalInstructionName = "AGENTS.local.md"`, `codexClaudeLocalName = "CLAUDE.local.md"`; iter4 §1: loop order at `codex_launcher.go:126` reads `AGENTS.local.md` first |
| A8 | The `-w` child reads the original project root (claim to verify) | `sed -n '250,275p' AGENTS.md` | §8: "For every local launch shape (bare, `cli`, `app`, `--spawn`, and `-w`), `moai codex` reads … from the project root … A `-w` child still reads the original project root." |
| A9 | Before-value (expectation, not bound) | `git show develop:CLAUDE.local.md \| wc -m`; `git show origin/develop:CLAUDE.local.md \| wc -m` | `45810` both. History of the figure: 44,381 (carve) → 44,740 (t1259 v0.2.0) → 45,810 (iter4, `a73c78d1a`) — it moves; M2 re-measures |
| A10 | Parent SPEC status | `grep '^status:' .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/spec.md` | `status: draft` (version `0.2.4` in this tree) |
| A11 | No SPEC-ID collision | `ls -d .moai/specs/*RECEPTION*` | no matches |

## §B External measurements this SPEC is built from (attributed, not re-measured)

- **§M1a — the reception negative** (`SPEC-INSTRUCTION-FILES-UNIFY-001/progress.md` §M1a;
  evidence `.moai/reports/t1243/m1/`; tool `claude 2.1.283`): in a real nested linked worktree
  (this repository's geometry), a token-probe session loaded the tracked contract (positive
  control present) and the gitignored primary-root `CLAUDE.local.md` (reached by
  directory-ancestor walk) but **not** the gitignored `AGENTS.local.md`
  (`LOCAL_AGENTS_TOKEN = NOT_PRESENT`). A sibling worktree got neither local file — the walk is
  not git-worktree-aware. `AGENTS.local.md` was not discovered even at the primary root of the
  synthetic fixture.
- **§M1b — Codex filename discovery is zero** (same progress.md §M1b; `codex-cli 0.157.0`):
  `codex debug prompt-input` over a fixture carrying `AGENTS.local.md` showed `LOCAL count=0`
  with the contract control present — bare Codex discovers no local instruction files; the
  launcher is the only Codex route.
- **iter4 D1** (`.moai/reports/t1259/plan-audit-iter4.md`, tree `a73c78d1a`): the parent's
  content criteria read the working copy of a gitignored file — a mutant satisfies them while
  the merge lands nothing. Fix folded into `REQ-IFU-021` (committed-tree reads + force-add).
- **iter4 D2** (same report): the parent's reception check lived only as an ungated plan
  bullet; executing M3 as written strips maintainer doctrine from every lane with all criteria
  green. This SPEC is that finding's remedy (the audit's option (b), split, plus option (a)'s
  criterion done properly as M1).

## §C What remains unmeasured (M1's job — listed so nothing pretends to be settled)

1. Whether the `@AGENTS.local.md` import fires in a linked worktree **when the file is present
   and tracked there** (design.md §B). No prior measurement covers the tracked geometry; A1's
   negative statement was written and measured against the untracked one.
2. Whether the `moai codex -w` child actually receives the content (A8 is the launcher's own
   documentation claim, not an observation of a `-w` child).
3. Whether `claude --worktree` native creation behaves identically to `moai cc -w` for the
   tracked file (expected yes — both end in a git checkout of tracked content — but AC-LIR-002
   measures rather than assumes; A5's copy mechanism is inert for tracked files by its own
   header).

## §D Source survey pointers

- `internal/cli/codex_launcher.go` — the local-instruction loop and `developer_instructions`
  producer (parent design.md §B has the symbol-level analysis; not duplicated).
- `.claude/rules/moai/workflow/worktree-integration.md` § `.worktreeinclude` — the native copy
  mechanism's documentation.
- `internal/cli/` `migrate_agency*` — the move-plus-backup verb precedent (parent's M1; this
  SPEC runs the verb, does not build it).
