# AC-GDP-001 reading record (M3)

Read by: manager-develop (run-phase part 1, card t622). Sources: `ac001-local-paras.md` and
`ac001-template-paras.md` — every paragraph of the `## Synchronization` section that contains both
`fetch` and `rev-list` (1 paragraph per copy, `ac001-{local,template}-paras.txt`). The local and
template sections are byte-identical (`diff` exit 0), so the single paragraph below is read once for
both copies. The generated `.toml` copy of this section is regenerated in M4 and is read there; it
is NOT covered by this record.

Paragraph (both copies):

> Pre-flight status reads are read-only, but `git rev-list --count --left-right` reads the
> remote-tracking refs that `git fetch` updates, so the two are not independent: run `git fetch`
> first and wait until it completes, then run `git rev-list --count --left-right` — never in the
> same batch as the fetch. Reads that do not consume the fetch result (`git status`,
> `gh pr checks --json`) may run in parallel with the fetch as one single-turn multi-Bash batch per
> `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (…).

| Question | Answer | Where the paragraph says it |
|---|---|---|
| (1) Does it say `git rev-list` runs after `git fetch` has finished? | Yes | "run `git fetch` first and wait until it completes, then run `git rev-list --count --left-right`" |
| (2) Does it keep fetch and rev-list out of the same batch / list / table / parallel group? | Yes | "never in the same batch as the fetch"; the parallel batch it allows contains only the fetch plus `git status` and `gh pr checks --json` — the reads that do not consume the fetch result. No list or table in the section places fetch and rev-list together (`ac001-{local,template}-listgroup.md` empty) |

Scope note (REQ-GDP-001's last sentence): parallelizing `git status` and `gh pr checks --json` with
the fetch is explicitly permitted by the requirement and is what the paragraph allows.

Result: both answers "Yes" for the one paragraph in each copy. The automatic detector agrees:
autofail `test -e` 0 / `test -s` 1 for both copies (`ac001-judge-summary.txt`).

## M4 addendum — generated `.toml` copy (run-phase part 2)

Read by: manager-develop (run-phase part 2, card t622) after `make agents-emit` regenerated
`internal/template/templates/.codex/agents/moai/manager-git.toml` (commit `5708e04d2`). Source:
`ac001-toml-paras.md` — the only paragraph of the `.toml` `## Synchronization` section containing
both `fetch` and `rev-list` (`ac001-toml-paras.txt` = 1). `cmp ac001-toml-paras.md
ac001-template-paras.md` exit 0: it is byte-identical to the template paragraph read above, and the
whole `.toml` section equals the template section (`ac014-sync.diff` exit 0).

| Question | Answer | Where the paragraph says it |
|---|---|---|
| (1) Does it say `git rev-list` runs after `git fetch` has finished? | Yes | "run `git fetch` first and wait until it completes, then run `git rev-list --count --left-right`" |
| (2) Does it keep fetch and rev-list out of the same batch / list / table / parallel group? | Yes | "never in the same batch as the fetch"; the parallel batch holds only fetch + `git status` + `gh pr checks --json`; `ac001-toml-listgroup.md` empty (`test -s` exit 1) |

Result for the `.toml` copy: both answers "Yes". Automatic detector: autofail `test -e` 0 /
`test -s` 1 (`ac001-toml-autofail.md`).
