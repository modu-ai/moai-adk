# t503 run-phase evidence — SPEC-CODEX-COMMAND-SKILLS-001

Collected in worktree `.claude/worktrees/t503`, branch `WT-codex-command-skills`, 2026-09-07.
File set: this index + verbatim command outputs in the sibling files listed per row.

| # | Claim | Evidence file (verbatim output) | Baseline-attribution |
|---|-------|--------------------------------|----------------------|
| 1 | AC-005 RED observed before the collision guard landed | `t503-red-ac005.txt` (rc=1; `Emit succeeded over a colliding name; want refusal diagnostic`) | guard call removed from `emit.go`, this tree, pre-e7d2a1658 |
| 2 | AC-010 RED observed on a hand-mutated committed artifact | `t503-red-ac010.txt` (`make commands-emit-check` → recipe exit 1, `command-skill drift: … run make commands-emit`) | mutation of `moai-run/SKILL.md`, this tree, post-6ae337e60 wiring |
| 3 | AC-010 GREEN after regeneration + check-only run writes nothing | inline: `make commands-emit` ok; re-check exit 0; `git status --porcelain templates/.agents/` → empty | this tree, post-regeneration |
| 4 | Affected packages green (internal/template tree) | `pkg-tests.txt` (3× `ok`, rc=0) | HEAD at export time, this run |
| 5 | Emitter package coverage 90.0% (≥85% gate) | `coverage.txt` (`coverage: 90.0% of statements`) | this run |
| 6 | `make build` (both emit-checks + templ-generate + catalog hashes + go build) | `make-build.txt` (rc=0) | this run |
| 7 | Cross-platform build | `windows-build.txt` (GOOS=windows exit 0; host exit 0) | this run |
| 8 | Lint: 0 issues | `lint.txt` (`0 issues.`) | this run |
| 9 | AC-001/003/004/007 count greps | inline in report: 16 / 16 / 0 / 16 | this run |
| 10 | AC-006 body byte-equality (16 pairs in test + todo spot diff rc=0) | `TestBodiesByteEqual` in pkg-tests.txt run; spot diff exit 0 | this run |
| 11 | AC-011 idempotent regeneration (`diff -r` snapshot form) | `/tmp` snapshot vs tree → empty (exit 0), twice executed | this run |
| 12 | AC-002 command sources untouched after emission runs | `git status --porcelain …/commands/moai/` → empty (0 lines) | this run |
| 13 | gofmt clean, go vet clean | gofmt -l → empty; vet exit 0 | this run |

Measured nuances (recorded, not smoothed over):
- `make commands-emit-check` failure: the recipe exits 1 with the drift diagnostic; `make` itself surfaces exit code 2 on a failed target. Identical behavior to the shipped `agents-emit-check`; the AC-010 "exits 1" holds at the recipe level.
- `TestSkillMirror_SlimSetEqualsCanonicalAndIsSmaller` was re-expressed (not deleted): `.agents/skills` now holds mirror entries PLUS the 16 published real directories, so the invariant is asserted on the mirror partition with zero-dangling-links preserved.

PRESERVE verification: `git status --porcelain` empty on `templates/.claude/commands/moai/` after every emission run; `internal/codexwiring/`, `internal/template/agentemit/` untouched (`git log` shows no commits touching them on this branch).
