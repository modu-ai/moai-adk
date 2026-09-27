# SPEC-LOCAL-INSTR-RECEPTION-001 — design

> Authored by card t1290 (plan phase). Carries this SPEC's own decision — the worktree
> reception mechanism — with the measurement each choice rests on. Every prior measurement
> cited here is attributed in research.md §A; this file argues from them.

## §A The decision: force-track `AGENTS.local.md` (git-tracked delivery)

**The migrated file is committed to git (force-added past `.gitignore:275`), so every worktree
receives it through git's own checkout, and the tracked `CLAUDE.md`'s `@AGENTS.local.md`
import resolves inside the project.**

Why this is the recommended option, in the order the reasons weigh:

1. **Coverage is git's, not a tool's.** A tracked file is present at the root of every
   worktree, from every entry path — `moai cc -w`, `claude --worktree`, `EnterWorktree`
   re-entry into existing trees, `moai codex -w` children, even raw `git worktree add`. No
   copy step, no creation-time window, no list of launchers to keep in sync. Every competing
   mechanism covers a subset of these; git covers all of them by semantics.
2. **Content flows, not snapshots.** Tracked content updates through the normal merge/absorb
   path (this repository's develop integration). A copied gitignored file is frozen at
   creation — a long-lived tree silently diverges from the canonical copy. With tracking, the
   file every tree holds IS the committed copy, which is exactly what `CLAUDE.local.md` §0.1
   already declares canonical. The delivery mechanism and the canonical-copy doctrine become
   the same statement.
3. **It is this repository's existing posture.** `CLAUDE.local.md` is tracked today despite
   its ignore rule (`.gitignore:276`; `git ls-files` confirms) — the maintainer-owned
   local-instruction file is already a force-tracked exception here. The migrated file
   inherits a posture, it does not invent one.
4. **It is the iter4 D1 fix anyway.** The audit's D1 required force-adding so the content
   criteria measure a committed file. Choosing tracked delivery makes that fix load-bearing
   instead of incidental: the same commit that satisfies the criteria is the one that delivers
   reception.
5. **Zero code.** No launcher change, no materializer change, no hook. The one unmeasured link
   — whether the `@AGENTS.local.md` import fires in a worktree when the file exists there —
   is exactly what M1's probe measures before anything depends on it.

## §B The unmeasured premise M1 settles

`AGENTS.md:262` states that in a linked worktree the import "points outside the project and is
skipped silently." That statement was written against (and §M1a measured) the **untracked**
geometry: the file exists only at the primary root, so the worktree's import target is absent.
The tracked geometry is different — `CLAUDE.md` is tracked and present at the worktree root,
its `@AGENTS.local.md` import is project-relative, and the target exists in-project. Whether
Claude Code resolves that import in a linked worktree is **the** open measurement question; no
prior measurement covers it (§M1a's fixture was untracked). M1's probe answers it before the
migration depends on the answer — the positive control (the tracked contract, which reaches the
session through the tracked `CLAUDE.md`'s `@AGENTS.md` import on the same mechanism) is what
makes the answer discriminating: control present + token absent means imports fire and the file
is what failed; both absent means the leg measured nothing.

## §C The fallback ladder (applied at M1 step 5, only if a Claude-side leg is red)

1. **`.worktreeinclude` entry.** Add `/AGENTS.local.md` to `.worktreeinclude` (present in this
   tree and template-mirrored at `internal/template/templates/.worktreeinclude`; documented
   coverage: `claude --worktree`, `moai cc|glm|cg -w`, `isolation: worktree` — creation-time).
   Template-First: both mirrors, neutral wording. Cost: covers creation-time only — re-entry
   into pre-existing trees and raw-created trees stay uncovered, and content becomes a
   creation-time snapshot. Accepted as fallback, not as primary, for exactly those two gaps
   (AC-LIR-003 exists because of the first).
   - Composition note: `.worktreeinclude` copies only files that are both listed and
     gitignored ("tracked files are never duplicated"). If the fallback fires, the ladder
     implies untracking the fixture — the two mechanisms are alternatives, not a stack.
2. **Re-probe every leg** under the fallback geometry; the same ACs judge it.
3. **If the ladder cannot close the gate: blocker report** with the measured matrix. The
   migration does not proceed on a partially-proven channel — REQ-LIR-004 is a stop condition,
   not a preference.

Codex-side coverage does not enter the ladder: the launcher path (design §D row 3) is
independent of file tracking, because it reads the original project root directly.

## §D Rejected mechanisms, with the measurement that rejected each

| Mechanism | Verdict | Grounds |
|---|---|---|
| Directory-ancestor walk (today's `CLAUDE.local.md` path) | Rejected | §M1a measured `LOCAL_AGENTS_TOKEN = NOT_PRESENT` for `AGENTS.local.md` in a real nested worktree while the walk found `CLAUDE.local.md` — the discovery set is per-filename; `AGENTS.local.md` is not in it. Not an assumption; a measurement. |
| Native `.worktreeinclude` as primary | Rejected as primary, kept as fallback | Creation-time only (its own header); re-entry and pre-existing trees uncovered; snapshot staleness. §C. |
| Codex launcher `developer_instructions` as the sole channel | Rejected for Claude legs | It covers `moai codex -w` children only (§8; §M1b showed bare Codex discovers zero local files). Claude Code sessions never see it. Correct role: AC-LIR-004's mechanism, verified not assumed. |
| `moai worktree new` materializer copy (MoAI-side) | Rejected | Covers only materializer-created trees; new Go code; dominated by git's own coverage (§A.1) at strictly greater cost and strictly less reach. |

## §E Probe design

Two tokens, one probe, three signals (exit code, control, local token):

- **Control token** — the `AGENTS.md` version line ("Version: …"). It is tracked content that
  reaches every worktree session through `CLAUDE.md`'s `@AGENTS.md` import — the same import
  mechanism the local file must ride — and it is stable enough for a probe while varying
  harmlessly (any non-`ABSENT` line passes).
- **Local token** — `LOCAL_AGENTS_TOKEN = <value>`, a line inside the fixture file. Never
  derived from file presence (`ls`/`test -f`): the probe asks the session what it *loaded*,
  because the unit under test is the load path, not the filesystem.
- **Negative control** — the same probe on a geometry without the file (AC-LIR-005), proving
  the probe's red exists and is observable.

The pattern is §M1a's (prompt-only, no tools, haiku model, `timeout 180`, env-scrubbed single
invocation) — reused because it is the measured template this repository already trusts, not
reinvented.

## §F Cross-references

- research.md §A — every measurement cited above, with command, output, and tree SHA.
- plan.md §D M1 — the milestone that executes §B's measurement and §C's ladder.
- `SPEC-INSTRUCTION-FILES-UNIFY-001/design.md` — the parent-of-parent's read-order analysis;
  not duplicated here.
