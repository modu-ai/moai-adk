# AC-GDP-030 outcome — case (A): no published-skill change

Judged after the command-source commit `2f4dfd803b1b7dace279fea7583dad04e6637a23` (X2), on branch
`WT-git-procedure-fixes`, CARD_BASE `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` (`card-base.txt`).

## Positive control (c), pre-flight (before any edit)

| Step | File | Result |
|---|---|---|
| drift planted, `make commands-emit-check` | `ac030-red.txt`, `ac030-red.exit` | make exit 2 (recipe `Error 1`), `--- FAIL: TestGoldenCommittedArtifactsMatchEmission … sha256 mismatch` — drift detected |
| restore + `cmp` | `ac030-restore-cmp.exit` | exit 0 |
| `make commands-emit-check` again | `ac030-control-green.txt`, `.exit` | exit 0 |
| `git status --porcelain -- <T>/.agents/skills/ .agents/skills/` | `ac030-control-clean.txt` | `test -e` 0, `test -s` 1 |
| tracked published pair | `ac030-tracked.txt`, `ac030-tracked-count.txt` | 2 |

## Right after the source edit (before commit)

| Command | File | Result |
|---|---|---|
| `make commands-emit` | `ac030-emit.txt`, `ac030-emit.exit` | exit 0 |
| `make commands-emit-check` | `ac030-check.txt`, `ac030-check.exit` | exit 0 |
| `git status --porcelain -- internal/template/templates/.agents/skills/ .agents/skills/` | `ac030-emit-status.txt` | `test -e` 0, `test -s` 1 (0 bytes) — case (A) candidate |

## After the source commit

| Command | File | Result |
|---|---|---|
| `git diff --name-only develop...HEAD -- internal/template/templates/.agents/skills/ .agents/skills/` | `ac030-changed.txt` | exit 0, `test -e` 0, `test -s` 1 — **empty** |
| `git log --format=%H HEAD --not develop -- internal/template/templates/.claude/commands/moai/sync.md.tmpl` | `ac030-src-commits.txt` | 1 line: `2f4dfd803b1b7dace279fea7583dad04e6637a23` (`test -s` 0) |
| `git log --format=%H HEAD --not develop -- <T>/.agents/skills/moai-sync/SKILL.md .agents/skills/moai-sync/SKILL.md` | `ac030-artifact-commits.txt` | exit 0, `test -e` 0, `test -s` 1 — **empty** |
| `git diff --name-only develop...HEAD -- <T>/.claude/commands/moai/sync.md.tmpl .claude/commands/moai/sync.md` (path-form control) | `ac030-src-changed.txt`, `ac030-src-changed-count.txt` | 2 — the same diff form does see changed files, so the empty `ac030-changed.txt` is not a path error |
| `grep -v -x -F -f ac030-src-commits.txt ac030-artifact-commits.txt` | `ac030-orphan-artifact-commits.txt` | exit 1, `test -e` 0, `test -s` 1 — no orphan published commit |
| `test -s <T>/.agents/skills/moai-sync/SKILL.md` | — | exit 0 |
| `diff .agents/skills/moai-sync/SKILL.md <T>/.agents/skills/moai-sync/SKILL.md` | `ac030-published-lt.diff` | exit 0 |
| AC-026 (ii) flag detector on the template published skill | `ac030-published-flags.txt` | `test -e` 0, `test -s` 1 |

## Contents of the three case-deciding files

- `ac030-changed.txt`: empty (0 bytes)
- `ac030-src-commits.txt`: `2f4dfd803b1b7dace279fea7583dad04e6637a23`
- `ac030-artifact-commits.txt`: empty (0 bytes)

Case: **(A)** — the source edit changed only `argument-hint`, which the emitter does not carry into
the published skill (consistent with `loader.go:4-6` as read at plan time); the published pair is
unchanged and still byte-identical, and no published commit exists outside the source commit.
