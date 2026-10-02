# SPEC-RUN-EXTERNAL-DELEGATION-001 — Acceptance Criteria (revision 4)

> Verification layer. Every criterion is Given-When-Then, binary, and decided by the
> actual output of a named command. The GEARS requirements live in `spec.md` §C; each
> AC opens with a **Covers** line naming the `REQ-RXD-XXX` it verifies. 16 criteria
> (Tier M ceiling 16), 16 requirements.
>
> **Command conventions.**
> - `<SCRUB>` is the environment scrub, run as one compound call:
>   `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` (AGENTS.md §4). Ledger rows that need it are informational — the prefix carries `&&`, which is outside the single-invocation form.
> - `<BASE>` is the output of `git merge-base develop HEAD`, read at the moment of use and never pinned (a literal base SHA lets other cards' commits into the range after an absorb). At plan time it equals `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`. A range read after the card merged into `develop` is empty and measures nothing; it is evaluated before the merge only.
> - Every `go test -run` pattern is anchored (`^…$`); an unanchored pattern also selects longer names. A `-run` result is read together with its swept count: a pattern that selects zero tests prints `ok` and proves nothing, so the verbose `--- PASS` lines are the evidence, never the final `ok` alone.
> - The `grep -n -o -F -e …` rows of the ledger are the **cheap independent probes** of the doctrine anchors: empty output means none of the listed phrases exists. They look at the whole file. The decisive green check — each anchor present **inside its own H3 subsection**, which is what closes the scoping mutant — is the guard test `TestRunExternalDelegationDoctrine` (plan.md §D), and a criterion's green path always names its subtest.
>
> **Classification.** **release-blocking** = a single-invocation RED-now ledger entry exists, was observed red on the pinned tree, and is reproducible by re-running it. **regression-guard** = green on arrival; it holds a state that must not be lost, and each carries a positive control showing the probe fires and a mutant that turns it red.
>
> **RED-now / green-path pairing.** Each AC states why it is red now and which milestone (plan.md §E) flips it. The wording-anchor criteria (AC-RXD-003 to AC-RXD-009) decide presence of the contract phrases listed in plan.md §D; their subsection-scoped, mutant-checked form is the guard test (M1).

## Evidence ledger (RED-now observations, tree `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`)

**One document-level pin: every entry below was measured in this plan revision on the tree `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b` (HEAD, clean tracked tree; the untracked SPEC directory is the only addition), in the worktree `.moai/worktrees/t1424`.** It binds every entry that carries no pin of its own. The ledger is fenced rather than tabled because a table cell mangles shell metacharacters. Each entry is `id`, the plain single invocation, the verbatim stdout, the exit code as its own field, and why it is red. Exit codes of `grep` entries were read by appending `; echo "rc=$?"` to the invocation; the command line recorded is the plain one. For `git` entries the exit code is recorded as `0 (no error reported)` because the worktree guard refuses a `git` command with a trailing `echo`. Output order of a multi-file `grep -c` is the tool's own and not significant. Where an entry's `out:` reads `(no output)`, stdout was empty. Entries marked `(scratch, informational)` — E3e, E7g, E7h, E22 — demonstrate a check on fixtures or copies outside the tree (the scratch directory is not a citation target and the fixtures are described in the entry); they carry no release-blocking criterion's RED-now observation, only positive controls and observed failures of regression-guard checks.

### Capability grant, single home, pointers (AC-RXD-001, AC-RXD-002, AC-RXD-011)

```
E1    cmd:  grep -c -F "mcp__moai__codex_task" .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  .claude/agents/moai/manager-develop.md:0
            internal/template/templates/.claude/agents/moai/manager-develop.md:0
      exit: 1
      why:  neither copy carries the tool today (so the MCP-tools section cannot name it either)

E1b   cmd:  grep -c -E '^tools: Read, Write, Edit, Bash, Grep, Glob, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__moai__verify_snapshot, mcp__moai__verify_trend, mcp__moai__goal_status, mcp__moai__codex_task, mcp__moai__codex_job_status, mcp__moai__codex_job_result, mcp__moai__codex_job_cancel, mcp__moai__glm_task, mcp__moai__glm_job_status, mcp__moai__glm_job_result, mcp__moai__glm_job_cancel$' .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  internal/template/templates/.claude/agents/moai/manager-develop.md:0
            .claude/agents/moai/manager-develop.md:0
      exit: 1
      why:  the full-line pin (existing prefix + eight-tool tail) matches in neither copy

E1c   cmd:  grep -c -E '^tools: Read, Write, Edit, Bash, Grep, Glob, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__moai__verify_snapshot, mcp__moai__verify_trend, mcp__moai__goal_status$' .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  internal/template/templates/.claude/agents/moai/manager-develop.md:1
            .claude/agents/moai/manager-develop.md:1
      exit: 0
      why:  positive control: the literal prefix of the pin is exactly today's line, so E1b is red for the missing tail only and the pin is satisfiable

E1d   cmd:  grep -c '^tools:' .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  .claude/agents/moai/manager-develop.md:1
            internal/template/templates/.claude/agents/moai/manager-develop.md:1
      exit: 0
      why:  control: one `tools:` line per copy today; green keeps it at 1 (no second line carrying the grant)

E1e   cmd:  grep -c -F "mcp__moai__codex_task" .claude/agents/moai/super-advisor.md
      out:  2
      exit: 0
      why:  positive control: the same probe fires on the agent that carries the tool (tools line and body)

E2    cmd:  grep -c "^## External Model Delegation" .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  .claude/skills/moai/workflows/run.md:0
            internal/template/templates/.claude/skills/moai/workflows/run.md:0
      exit: 1
      why:  the section does not exist in either run.md

E2c   cmd:  grep -c "^## Recursive Self-Diagnosis Loop" .claude/skills/moai/workflows/run.md
      out:  1
      exit: 0
      why:  positive control: the heading form fires on an existing run.md section

E3    cmd:  grep -c "External Model Delegation" .claude/skills/moai/workflows/fix.md .claude/skills/moai/workflows/loop.md .claude/agents/moai/manager-develop.md internal/template/templates/.claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/loop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  .claude/skills/moai/workflows/fix.md:0
            internal/template/templates/.claude/skills/moai/workflows/fix.md:0
            .claude/skills/moai/workflows/loop.md:0
            .claude/agents/moai/manager-develop.md:0
            internal/template/templates/.claude/agents/moai/manager-develop.md:0
            internal/template/templates/.claude/skills/moai/workflows/loop.md:0
      exit: 1
      why:  no pointer title in any of the six files

E3b   cmd:  grep -c -F ".claude/skills/moai/workflows/run.md" .claude/skills/moai/workflows/fix.md .claude/skills/moai/workflows/loop.md .claude/agents/moai/manager-develop.md internal/template/templates/.claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/loop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  .claude/skills/moai/workflows/fix.md:0
            .claude/agents/moai/manager-develop.md:0
            internal/template/templates/.claude/skills/moai/workflows/fix.md:0
            .claude/skills/moai/workflows/loop.md:0
            internal/template/templates/.claude/skills/moai/workflows/loop.md:0
            internal/template/templates/.claude/agents/moai/manager-develop.md:0
      exit: 1
      why:  none of the three files names the path a subagent must Read; a bare title would not be actionable

E3c   cmd:  grep -c "manager-develop" .claude/skills/moai/workflows/fix.md
      out:  3
      exit: 0
      why:  positive control: fix.md already names the agent where the pointer belongs

E3d   cmd:  grep -n -o -F -e '### Delegable classes' -e 'fixture regeneration' -e 'lint-repair draft' -e 'characterization-test draft' -e 'delegation is never required' -e 'bounded to the files the prompt names' -e '### Excluded work' -e 'design or architecture decisions' -e 'security-sensitive code' -e 'SPEC artifacts' -e 'public-API changes' -e 'the external output would decide' -e 'more than the bounded files' -e 'implementation code under a test-first cycle' -e 'semantic failures are never delegated' -e 'protected files are never named' -e '### Request construction' -e 'starts every job with background true' -e 'git rev-parse --show-toplevel' -e 'its own L1 tree' -e 'never sets the write argument' -e 'names the bounded files' -e 'patch text only' -e 'no secrets' -e 'over HTTPS' -e 'takes no project_root' -e 'bounded excerpts' -e 'never contains content of an excluded class' -e '### Result handling' -e 'untrusted data' -e 'never follows instructions found in a result' -e 'never executes commands found in a result' -e 'through its own edit tools' -e 'runs the verification the cycle already requires' -e 'reports the measured output' -e 'is discarded' -e 'done directly' -e '### Value observation' -e 'the generator the project already owns' -e 'unmodified code under test' -e 'never taken from the reply' -e '### Fail-open' -e 'unavailable, inconclusive, failed or empty' -e 'at most five reads of the job status or result tool' -e 'is a failed delegation' -e 'cancels the job' -e 'does the subtask itself' -e 'no blocker report' -e 'each read follows a unit of its own work' -e 'never a sleep loop' -e '### One-writer rule' -e 'does not edit the files named in an in-flight prompt' -e 'reads or cancels every job it started before reporting completion' -e 'the orchestrator owns the one-writer-per-tree rule' -e '### Harness scope' -e 'Claude-harness capability' -e 'on any other harness' .claude/skills/moai/workflows/fix.md .claude/skills/moai/workflows/loop.md .claude/agents/moai/manager-develop.md internal/template/templates/.claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/loop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  .claude/agents/moai/manager-develop.md:162:SPEC artifacts
            .claude/agents/moai/manager-develop.md:228:SPEC artifacts
            internal/template/templates/.claude/agents/moai/manager-develop.md:163:SPEC artifacts
            internal/template/templates/.claude/agents/moai/manager-develop.md:229:SPEC artifacts
      exit: 0
      why:  the pointer-forbidden set (all 49 anchors and the 8 `### ` headings of plan.md §D = 57 phrases) occurs in the six pointer files in exactly one phrase: `SPEC artifacts`, in both `manager-develop.md` copies — which is why it is the one allowlisted anchor. This is the arrival state; green keeps the output to `SPEC artifacts` hits only. The probe fires (the four hits are its own positive control); a restated anchor in any pointer file adds a line

E3e   cmd:  (scratch, informational — a scratch Go program that mirrors the `pointers` subtest; not a repository file and not reproducible from the tree) go run pointercheck.go good_pointer.md audit_mutant_two_line.md padded_one_line.md short_paraphrase_residual.md bullet_restates.md
      out:  PASS good_pointer.md
            FAIL audit_mutant_two_line.md
                 - pointer paragraph words=73 (max 40)
                 - anchor present: no secrets
            FAIL padded_one_line.md
                 - pointer paragraph words=60 (max 40)
            PASS short_paraphrase_residual.md
            FAIL bullet_restates.md
                 - anchor present: over HTTPS
                 - anchor present: takes no project_root
      exit: 1 (the program exits 1 when any file fails; the expected mix is one pass, three fails, one residual pass)
      why:  fixtures: `good_pointer.md` is a 24-word one-line pointer carrying the path and the title (passes); `audit_mutant_two_line.md` is the iteration-2 audit's mutant — the pointer line plus one 67-word line paraphrasing the whole procedure, 2 physical lines, 0 of the 13 old `(R)` phrases and no `### ` heading (`grep -c -F` over those 14 strings printed `0`, rc 1; the revision-2 checks passed it) — it fails the word cap and, here, also the anchor `no secrets`; `padded_one_line.md` is the same paraphrase on the pointer's own line, 1 line, 60 words (fails the word cap); `bullet_restates.md` puts anchor phrases into an MCP-tools bullet (fails the absence check, which has no bullet exemption); `short_paraphrase_residual.md` is a 19-word one-line pointer-plus-paraphrase with no anchor (passes — the accepted lexical residual of spec.md R-3)
```

### The doctrine anchors (AC-RXD-003 to AC-RXD-009), live and template `run.md`

Each entry lists its subsection's H3 heading and every anchor of plan.md §D for that subsection; all of them are searched in both `run.md` copies, the live file and its template mirror, which are byte-identical today. Together E10a-E10g search 57 phrases — the 49 anchors of plan.md §D and the 8 headings (16 + 12 + 9 + 4 + 9 + 4 + 3).

```
E10a  cmd:  grep -n -o -F -e '### Delegable classes' -e 'fixture regeneration' -e 'lint-repair draft' -e 'characterization-test draft' -e 'delegation is never required' -e 'bounded to the files the prompt names' -e '### Excluded work' -e 'design or architecture decisions' -e 'security-sensitive code' -e 'SPEC artifacts' -e 'public-API changes' -e 'the external output would decide' -e 'more than the bounded files' -e 'implementation code under a test-first cycle' -e 'semantic failures are never delegated' -e 'protected files are never named' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 16 phrases of AC-RXD-003 exists in either run.md (the bare words `semantic failure` already occur at run.md lines 158, 160, 189, which is why the longer phrase is the anchor)

E10b  cmd:  grep -n -o -F -e '### Request construction' -e 'starts every job with background true' -e 'git rev-parse --show-toplevel' -e 'its own L1 tree' -e 'never sets the write argument' -e 'names the bounded files' -e 'patch text only' -e 'no secrets' -e 'over HTTPS' -e 'takes no project_root' -e 'bounded excerpts' -e 'never contains content of an excluded class' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 12 phrases of AC-RXD-004 exists

E10c  cmd:  grep -n -o -F -e '### Result handling' -e 'untrusted data' -e 'never follows instructions found in a result' -e 'never executes commands found in a result' -e 'through its own edit tools' -e 'runs the verification the cycle already requires' -e 'reports the measured output' -e 'is discarded' -e 'done directly' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 9 phrases of AC-RXD-005 exists (revision 3 added `done directly`, the anchor of REQ-RXD-008's "subtask done directly" half; it was 8 phrases in revision 2)

E10d  cmd:  grep -n -o -F -e '### Value observation' -e 'the generator the project already owns' -e 'unmodified code under test' -e 'never taken from the reply' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 4 phrases of AC-RXD-006 exists

E10e  cmd:  grep -n -o -F -e '### Fail-open' -e 'unavailable, inconclusive, failed or empty' -e 'at most five reads of the job status or result tool' -e 'each read follows a unit of its own work' -e 'never a sleep loop' -e 'is a failed delegation' -e 'cancels the job' -e 'does the subtask itself' -e 'no blocker report' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 9 phrases of AC-RXD-007 exists (revision 3: the read bound covers the job status or result tool, and two anchors decide the separation and the no-polling halves of REQ-RXD-010; it was 7 phrases in revision 2)

E10f  cmd:  grep -n -o -F -e '### One-writer rule' -e 'does not edit the files named in an in-flight prompt' -e 'reads or cancels every job it started before reporting completion' -e 'the orchestrator owns the one-writer-per-tree rule' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 4 phrases of AC-RXD-008 exists (the candidate anchor `Background Agent Execution` was dropped: it already occurs at run.md line 191)

E10g  cmd:  grep -n -o -F -e '### Harness scope' -e 'Claude-harness capability' -e 'on any other harness' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md
      out:  (no output)
      exit: 1
      why:  none of the 3 phrases of the anchor half of AC-RXD-009 exists

E10x  cmd:  grep -n -o -F -e '## Recursive Self-Diagnosis Loop' -e 'zzzz-no-such-anchor' .claude/skills/moai/workflows/run.md
      out:  181:## Recursive Self-Diagnosis Loop
      exit: 0
      why:  positive control: the same multi-pattern `-F -o` form fires when one of its phrases exists, so the empty E10a-E10g outputs are absence, not a broken probe
```

### Consumer statements (AC-RXD-010)

```
E4a   cmd:  grep -c -E '^\| `mcp__moai__(codex_task|codex_job_status|codex_job_result|codex_job_cancel|glm_task|glm_job_status|glm_job_result|glm_job_cancel)` .*manager-develop' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  none of the eight per-tool rows names `manager-develop` (green: 8)

E4a-c cmd:  grep -c -E '^\| `mcp__moai__(codex_task|codex_job_status|codex_job_result|codex_job_cancel|glm_task|glm_job_status|glm_job_result|glm_job_cancel)` .*super-advisor' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  8
      exit: 0
      why:  positive control: the same row regex matches all eight rows on their existing consumer

E4b   cmd:  grep -c -E '^\| (Codex|GLM) delegation \|.*manager-develop' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  neither family-table row names `manager-develop` (green: 2)

E4b-c cmd:  grep -c -E '^\| (Codex|GLM) delegation \|.*super-advisor' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  2
      exit: 0
      why:  positive control: the family-table row regex fires on both rows

E4c   cmd:  grep -c -F "External Model Delegation" .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  neither family description points at the section yet (green: 2). Wrap-sensitive on its own — a hard-wrapped catalogue could split the phrase across two physical lines and read 1 — so E4e and E4f below pin each family sentence to one physical line

E4e   cmd:  grep -c -E 'manager-develop.*codex_task.*External Model Delegation' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  no physical line of the codex family description names `manager-develop`, then `codex_task`, then the section title (green: 1). A description that points at the section without naming the consumer, a sentence wrapped over two lines, or the sentence placed in the GLM description all leave it at 0

E4f   cmd:  grep -c -E 'manager-develop.*glm_task.*External Model Delegation' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  the same for the GLM family description (green: 1); E4e and E4f together decide that each family's sentence names `manager-develop`, its start tool and the section on one physical line. Which description holds which sentence is not decided by a grep (a single line naming both start tools would satisfy E4e and E4f — E4c's count of 2 and the reader do the rest)

E4e-c cmd:  grep -c -E 'codex_task.*codex.*super-advisor' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  2
      exit: 0
      why:  positive control: the same three-term `.*` chain form fires on existing physical lines of the codex family

E4f-c cmd:  grep -c -E 'glm_task.*GLM.*super-advisor' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  1
      exit: 0
      why:  positive control: the same chain form fires on the existing GLM per-tool row

E4d   cmd:  grep -c -E '^\| `mcp__moai__codex_setup` .*manager-develop' .claude/rules/moai/core/moai-mcp-tools-catalogue.md
      out:  0
      exit: 1
      why:  the `codex_setup` row names no `manager-develop` consumer (green on arrival; must stay 0)

E5    cmd:  grep -c -F "read-only only while" .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md
      out:  internal/template/templates/.claude/rules/moai/development/agent-authoring.md:0
            .claude/rules/moai/development/agent-authoring.md:0
      exit: 1
      why:  the read-only-list note does not yet state the opt-in condition

E5b   cmd:  grep -c -F "External Model Delegation" .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md
      out:  .claude/rules/moai/development/agent-authoring.md:0
            internal/template/templates/.claude/rules/moai/development/agent-authoring.md:0
      exit: 1
      why:  the note does not yet point at the section that bounds the delegation

E5c   cmd:  grep -c "mcp__moai__codex_task" .claude/rules/moai/development/agent-authoring.md
      out:  1
      exit: 0
      why:  positive control: the warning sentence exists to be amended
```

### Change set, defaults, hygiene (AC-RXD-012 to AC-RXD-016)

```
E6    cmd:  git diff --name-only c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- . ':(exclude).moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001' ':(exclude).moai/reports/t1424'
      out:  (no output)
      exit: 0 (no error reported)
      why:  no planned file has changed on the whole tree; green prints the 15 paths of plan.md §B

E6a   cmd:  git diff --name-only c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- .claude/agents internal/template/templates/.claude/agents
      out:  (no output)
      exit: 0 (no error reported)
      why:  no agent definition changed; green prints exactly the two manager-develop.md paths

E6c   cmd:  git log --oneline -1 -- .claude/agents/moai/manager-develop.md
      out:  d8d164d90 docs(rules,agents): local doctrine inherits model and effort — card t1246
      exit: 0 (no error reported)
      why:  positive control: the pathspec resolves to a tracked file, so the empty entries above are not an empty sweep

E6d   cmd:  git diff --name-only d8d164d90~1..d8d164d90 -- .claude/agents/moai/manager-develop.md ':(exclude).moai/reports/t1424'
      out:  .claude/agents/moai/manager-develop.md
      exit: 0 (no error reported)
      why:  positive control for the `:(exclude)` form of E6: over a historical commit that touched the file, the same form prints it

E7    cmd:  grep -rn "allow_write: true" internal/template/templates .claude .moai/config
      out:  (no output)
      exit: 1
      why:  no opt-in is shipped or set in this worktree (green on arrival — a regression guard)

E7c   cmd:  grep -c "allow_write" .claude/rules/moai/development/agent-authoring.md
      out:  1
      exit: 0
      why:  positive control: the key is spelled this way where it is documented

E7d   cmd:  grep -n -i -E '\bwrite\b[^.]{0,40}\b(true|enabled|on)\b|allow_write[^.]{0,20}\btrue\b' .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md .claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/fix.md .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md
      out:  (no output)
      exit: 1
      why:  the widened write-instruction expression (case-insensitive proximity: `true`, `enabled` or `on` within 40 characters after the whole word `write`, or `true` within 20 characters after `allow_write`) finds nothing in the eight files the doctrine touches (green on arrival, must stay empty; the whole of `run.md` is scanned here, the section slice only by the test). Revision 2's three-spelling expression is superseded: it passed a section that instructs a write (E7h)

E7f   cmd:  grep -c -i -E '\bwrite\b[^.]{0,40}\b(true|enabled|on)\b|allow_write[^.]{0,20}\btrue\b' internal/cli/mcp_server.go
      out:  1
      exit: 0
      why:  positive control (tree file): the widened E7d expression matches where the opt-in is documented in the tool registration, so the empty E7d is absence, not a broken expression

E7g   cmd:  (scratch, informational) grep -n -i -E '\bwrite\b[^.]{0,40}\b(true|enabled|on)\b|allow_write[^.]{0,20}\btrue\b' bad_to_true.md ; the same over good_prescribed.md with -c ; the sentence `The agent never sets the write argument` counted with grep -c -F over good_prescribed.md
      out:  bad_to_true.md:
            3:The agent never sets the write argument to true.
            5:For lint-repair drafts it sets the `write` argument to `true`; pass `write` = `true`; or write: `true`; or "write" : true; or write:True; or turn write on; or a write-enabled turn.
            good_prescribed.md: 0
            anchor sentence count in good_prescribed.md: 1
      exit: 0 for the bad fixture (matches found), 1 for the good fixture count line (no match; grep -c exits 1 on a zero count), 0 for the anchor count
      why:  positive control (bad fixture) and negative control (prescribed wording). `bad_to_true.md` holds, as lines 1-5, a heading, the anchor sentence with " to true" appended (so the section may not append it — plan.md §D), and the audit's markdown-variant instruction in six spellings; the widened expression fires on lines 3 and 5. `good_prescribed.md` holds the two sentences plan.md §D prescribes (anchor sentence `The agent never sets the write argument, so a delegated turn stays read-only.` and the one-writer sentence naming `workflow.codex.task.allow_write` without a value) — it contains the anchor phrase `never sets the write argument` once and the expression finds nothing in it, so the prescribed wording passes the widened check

E7h   cmd:  (scratch, informational) grep -n -E 'write: ?true|write=true|"write": ?true' bad_to_true.md
      out:  (no output)
      exit: 1
      why:  the revision-2 narrow expression finds nothing in the same bad fixture that E7g flags on two lines — the evasion the iteration-2 audit demonstrated (N2), reproduced; it is kept only as the reason the expression was widened

E8    cmd:  grep -n "AllowWrite: false" internal/config/defaults.go
      out:  1245:				AllowWrite: false,
      exit: 0
      why:  the shipped default is false (green on arrival)

E8b   cmd:  grep -n "allow_write" /Users/goos/MoAI/moai-adk-go/.moai/config/sections/workflow.yaml
      out:  136:      allow_write: true
      exit: 0
      why:  the primary checkout's workflow.yaml — the file `readCodexTaskAllowWrite(projectDirResolver())` reads for a server rooted there — carries the opt-in open on this maintainer's machine (spec.md R-6); read outside the card tree, in this run

E9    cmd:  grep -c "super-advisor" .claude/rules/moai/core/moai-mcp-tools.md
      out:  0
      exit: 1
      why:  the always-loaded stub states no consumer for these tools, so it needs no edit

E9c   cmd:  git log --oneline -1 -- .claude/rules/moai/core/moai-mcp-tools.md
      out:  d43e50bb3 absorb develop 145c3d98c into WT-factory-self-dispatch (card t1240)
      exit: 0 (no error reported)
      why:  positive control: the stub is a tracked file

E11   cmd:  grep -rl "TestRunExternalDelegation" internal/template
      out:  (no output)
      exit: 1
      why:  the guard test does not exist

E11c  cmd:  grep -rl "TestAgentlessUtilityNoLLMControlFlow" internal/template
      out:  internal/template/agentless_audit_test.go
            internal/template/templates/.claude/skills/moai/workflows/fix.md
      exit: 0
      why:  positive control: the same form finds a guard test that exists

E12   cmd:  grep -c -F "mcp__moai__codex_setup" .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  internal/template/templates/.claude/agents/moai/manager-develop.md:0
            .claude/agents/moai/manager-develop.md:0
      exit: 1
      why:  `codex_setup` is absent today (green on arrival; must stay 0)

E12c  cmd:  grep -c -F "mcp__moai__codex_setup" .claude/agents/moai/super-advisor.md
      out:  2
      exit: 0
      why:  positive control: the probe fires where `codex_setup` is carried

E20   cmd:  grep -c -E 'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}|\bREQ-[A-Z0-9-]+|\bAC-[A-Z0-9-]+|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\bt[0-9]{3,5}\b' internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  1
      exit: 0
      why:  the token-class expression fires on a template file that already carries one earlier token — which is why REQ-RXD-015 is scoped to the lines this change adds, not to the whole file

E21   cmd:  git diff --name-only -G'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}|REQ-[A-Z]|AC-[A-Z]|20[0-9]{2}-[0-9]{2}-[0-9]{2}' c50da9c2f8aa1227073bd77caa07ca1c75b8d81b..HEAD -- .claude internal/template/templates
      out:  (no output)
      exit: 0 (no error reported)
      why:  no changed line matches a token class (green on arrival — a regression guard over the added lines)

E21c  cmd:  git diff --name-only -G'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}|REQ-[A-Z]|AC-[A-Z]|20[0-9]{2}-[0-9]{2}-[0-9]{2}' 4d7ec04e4~1..4d7ec04e4 -- internal/template/templates/.claude/agents/moai/manager-develop.md
      out:  internal/template/templates/.claude/agents/moai/manager-develop.md
      exit: 0 (no error reported)
      why:  positive control: over a historical commit whose changed lines carry a requirement token, the same `-G` form prints the file

E21d  cmd:  grep -rn -E "t1424|SPEC-RUN-EXTERNAL-DELEGATION" internal/template/templates .claude/agents .claude/skills .claude/rules/moai
      out:  (no output)
      exit: 1
      why:  this card's identifiers are in no shipped or live file (green on arrival)
```

### Informational baselines (the prefix or a pipe is outside the single-invocation form)

```
E13   cmd:  <SCRUB> go test -count=1 -run '^TestGoldenCommittedArtifactsMatchEmission$' ./internal/template/agentemit/
      out:  ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.397s
      exit: 0
      why:  the generated Codex files match the agent bodies today

E14   cmd:  <SCRUB> go test -count=1 -v -run '^TestCodexTaskAllowWrite_DistributedDefaultIsFalse$' ./internal/cli/
      out:  === RUN   TestCodexTaskAllowWrite_DistributedDefaultIsFalse
            --- PASS: TestCodexTaskAllowWrite_DistributedDefaultIsFalse (0.00s)
            PASS
            ok  	github.com/modu-ai/moai-adk/internal/cli	1.399s
      exit: 0
      why:  the default test names one test and it ran

E15   cmd:  <SCRUB> go test -count=1 -v -run '^(TestTemplateNoInternalContentLeak|TestContractModeLocalTemplateParity|TestContractModeBlocksWellFormed|TestAgentlessUtilityNoLLMControlFlow|TestAgentFrontmatterAudit|TestDeclaredRuleMirrorForks|TestSanitizedPairParity|TestRunSkillContainsModeUnknownSentinel|TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles|TestAllAgentsInCatalog|TestAllSkillsInCatalog)$' ./internal/template/
      out:  twelve top-level `--- PASS:` lines (TestTemplateNoInternalContentLeak, TestContractModeBlocksWellFormed, TestContractModeLocalTemplateParity, TestDeclaredRuleMirrorForks, TestAllSkillsInCatalog, TestAllAgentsInCatalog, TestRunSkillContainsModeUnknownSentinel, TestAgentlessUtilityNoLLMControlFlow, TestAgentFrontmatterAudit, TestCatalogHashCoversSkillSubfiles, TestManifestHashFormat, TestSanitizedPairParity) then:
            ok  	github.com/modu-ai/moai-adk/internal/template	1.343s
      exit: 0
      why:  the twelve template guards plan.md §F re-measures all pass today (the verbose stream was read, not only the final ok; the stream is longer than the 50-line ceiling and is not reproduced)

E16   cmd:  moai agent lint
      out:  last line: Summary: 25 total (0 errors, 25 warnings)
      exit: 0
      why:  baseline for AC-RXD-016; the installed build `moai version` prints `v3.2.0-rc.24   moai_cp/20260925_122548-1896-gc50da9c2f   built 2026-10-02T05:34:44Z`, built from this HEAD

E17   cmd:  GOOS=windows GOARCH=amd64 go vet ./internal/template/
      out:  (no output)
      exit: 0
      why:  baseline: the template package type-checks for Windows including its `_test.go` files (`go build ./...` does not compile them)

E18   cmd:  <SCRUB> go test -count=1 -v -run '^(TestMCPToolCatalogueDocsStayMirrorIdentical|TestMCPToolCatalogueFiguresMatchRegistry)$' ./internal/cli/
      out:  --- PASS: TestMCPToolCatalogueDocsStayMirrorIdentical (0.00s)
            --- PASS: TestMCPToolCatalogueFiguresMatchRegistry (0.00s)
            ok  	github.com/modu-ai/moai-adk/internal/cli	1.159s
      exit: 0
      why:  baseline of the two guards that own the catalogue's byte-identity and tool-count figures

E19   cmd:  diff .claude/agents/moai/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md | grep -c '^[<>]'
            diff .claude/skills/moai/workflows/fix.md internal/template/templates/.claude/skills/moai/workflows/fix.md | grep -c '^[<>]'
            diff .claude/skills/moai/workflows/loop.md internal/template/templates/.claude/skills/moai/workflows/loop.md | grep -c '^[<>]'
            diff .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md | grep -c '^[<>]'
      out:  5
            4
            2
            4
      exit: 0 each (measured by one loop over the four pairs in this order; the pipe puts this outside the single-invocation form)
      why:  the by-design differences of the four pairs; green keeps the four counts unchanged. The other two pairs are byte-identical: `cmp -s` of both run.md copies and of both catalogue copies exits 0 (rc=0 each, measured in this run)

E22   cmd:  (scratch, informational — the generator runs against a scratch copy via its own `--catalog` flag, so no tracked file is written) cp internal/template/catalog.yaml <scratch>/catalog.yaml ; go run ./internal/template/scripts/gen-catalog-hashes.go --all --catalog <scratch>/catalog.yaml ; cp <scratch>/catalog.yaml <scratch>/after1.yaml ; the same go run a second time ; cmp -s <scratch>/after1.yaml <scratch>/catalog.yaml ; cmp -s internal/template/catalog.yaml <scratch>/catalog.yaml
      out:  both generator runs ended `catalog.yaml updated successfully (13900 bytes)` (each also prints 49 `name: hash` lines, not reproduced); the two `cmp -s` calls printed nothing
      exit: 0 for each generator run; 0 for `cmp -s` after1 vs after2; 0 for `cmp -s` tracked catalog vs the scratch result; `git status --short` afterwards printed only `?? .moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001/`
      why:  the premise of AC-RXD-015's idempotence clause: the generator is a fixed point on this tree, and a second run changes nothing relative to the first. The run phase repeats the comparison on the real file after M3 — copy the file, run the generator, `cmp -s` — because a `git status` letter cannot see a second run's effect
```

## AC-RXD-001 — `manager-develop` carries exactly the eight delegation tools, in both copies, and names them in its body

**Covers**: maps REQ-RXD-001

- **Classification**: release-blocking (RED-now: E1b, controls E1c, E1d, E1e; E1 for the body half).
- **Given** the live agent file and its template mirror,
  **When** each `tools:` line and each MCP-tools section is read,
  **Then** the full-line command of E1b prints `1` for each of the two files (the literal prefix of today's line followed by exactly the eight-tool tail in the order of REQ-RXD-001); `grep -c '^tools:'` prints `1` for each (one `tools:` line, no second line carrying the grant); `grep -c -F "mcp__moai__codex_setup"` prints `0` for each (exit 1); and the guard test's `tools` subtest passes — it asserts the whole line, that **all eight names appear inside the agent body's `## MCP Tools` section**, each as the lead of a bullet line (`- ` + backtick + `mcp__moai__<name>`, the format of the section's three existing bullets), with the section sliced from that heading to the next `\n## ` and the slice asserted non-empty (REQ-RXD-001's second clause), and that `super-advisor.md` still carries `codex_setup`.
- **RED-now**: E1b — the full-line pin matches in neither copy, while E1c shows the prefix of that pin matches today.
- **Green** (M3): the anchored command prints `1` for both files, exit 0; `--- PASS: TestRunExternalDelegationDoctrine/tools ` (the name followed by a space).
- **Re-pin note**: if absorbing `develop` changes the existing prefix before M3, re-derive the literal prefix from `git show <BASE>:<path>` (first `tools:` line) in both the command and the test; the requirement is the line "as it is today" plus the tail.
- **Mutant probe**: deleting `Skill` or `Write` from the prefix, inserting `Agent` (which would break the flat-hierarchy guarantee) or `WebFetch` before `mcp__moai__goal_status`, reordering the tail, dropping one `glm_job_*` tool, granting only the codex family, or putting the grant on a second `tools:` line each fails the full-line anchor; adding `mcp__moai__codex_setup` makes the `codex_setup` count `1`; listing seven tools in `## MCP Tools` and naming the eighth only in the pointer paragraph, in a comment, in a later section or on the `tools:` line fails the section-scoped bullet assertion (the revision-2 count of "at least twice in each copy" passed that mutant).

## AC-RXD-002 — one home for the procedure; pointers name the path and the section and restate nothing

**Covers**: maps REQ-RXD-002

- **Classification**: release-blocking (RED-now: E2 with control E2c; E3 and E3b with control E3c). The anchor-absence half is green on arrival (E3d: only the allowlisted `SPEC artifacts` occurs); its observed failures are the scratch demonstration E3e now and the paraphrase mutant of the run phase (AC-RXD-016).
- **Given** the finished change,
  **When** the heading, the pointers and the pointer-forbidden set are checked,
  **Then** `grep -c "^## External Model Delegation"` prints `1` for the live and the template `run.md`; the guard test's `section` subtest passes (the heading exactly once per copy, a non-empty slice, the eight H3 subsections once each in order, the two copies byte-equal); its `pointers` subtest passes (in `fix.md`, `loop.md` and `manager-develop.md` of both trees exactly one physical line carries `.claude/skills/moai/workflows/run.md` and it also carries `External Model Delegation`; the pointer paragraph — that line to the line before the next blank line, list marker or `#` heading — is at most 2 physical lines and 40 words; and **no phrase of the pointer-forbidden set appears anywhere in the file**: every anchor and `### ` heading of plan.md §D except the allowlisted `SPEC artifacts`, which the agent body carries today — E3d); the E3d command prints only `SPEC artifacts` hits (four, one per copy line; the line numbers shift with the edit); `grep -c "^### Delegable classes"` prints `0` for those six files; and `git diff --numstat <BASE>..HEAD -- <the six pointer files>` shows at most 3 added lines for each `fix.md` and `loop.md` copy and at most 14 added lines for each `manager-develop.md` copy (the tools line, the MCP-tools bullets and the pointer paragraph — a coarse backstop; the paragraph cap and the anchor-absence check are the primary decision).
- **RED-now**: E2 and E3 — the section and every pointer title are absent; E3b — no pointer names the path.
- **Green** (M2 for the section, M3 for the pointers): the counts above and `--- PASS: TestRunExternalDelegationDoctrine/section ` and `--- PASS: TestRunExternalDelegationDoctrine/pointers `.
- **Mutant probe**: a pointer that says "see External Model Delegation" with no path fails `pointers`; a thirty-line restated procedure under any heading contains doctrine anchors and fails `pointers` and the added-line budget; **a paraphrase of the procedure** — the iteration-2 audit's mutant, the pointer line plus one 67-word line that restates the procedure without using an `(R)` phrase of revision 2 — fails the 40-word and the 2-line pointer cap (demonstrated on a scratch fixture, E3e: `pointer paragraph words=73 (max 40)`, and here also `anchor present: no secrets`), as does the same paraphrase padded onto the pointer's own line (E3e: `words=60`), and a bullet in the MCP-tools list that carries an anchor (`over HTTPS`) fails the absence check, which has no bullet or `tools:`-line exemption. **Not decided (accepted residual, spec.md R-3):** a paraphrase that fits both caps and uses none of the anchor phrases (E3e: the 19-word `short_paraphrase_residual.md` passes) — the plan-auditor and sync-auditor reading owns it. Renaming the `run.md` heading or leaving the template one sub-heading short fails `section`.

## AC-RXD-003 — the allowlist is closed, bounded and optional; the exclusion list names all nine kinds

**Covers**: maps REQ-RXD-003, REQ-RXD-004

- **Classification**: release-blocking (RED-now: E10a, control E10x).
- **Given** the `run.md` section,
  **When** the anchors of `### Delegable classes` (`fixture regeneration`, `lint-repair draft`, `characterization-test draft`, `delegation is never required`, `bounded to the files the prompt names`) and of `### Excluded work` (`design or architecture decisions`, `security-sensitive code`, `SPEC artifacts`, `public-API changes`, `the external output would decide`, `more than the bounded files`, `implementation code under a test-first cycle`, `semantic failures are never delegated`, `protected files are never named`) are searched,
  **Then** every anchor is present **inside its own subsection**, `### Delegable classes` holds exactly three class bullets, and the `allowlist` and `exclusions` subtests pass; the E10a command prints a match for each of the 16 phrases.
- **RED-now**: E10a — none of the phrases exists.
- **Green** (M1 RED, M2): `<SCRUB> go test -count=1 -v -run '^TestRunExternalDelegationDoctrine$' ./internal/template/` shows `--- PASS: TestRunExternalDelegationDoctrine/allowlist ` and `…/exclusions `, exit 0.
- **Mutant probe**: removing one class, adding a fourth class, dropping the bounded-files clause (`bounded to the files the prompt names`) or any single exclusion anchor fails its subtest, which names the missing phrase and the subsection; moving an exclusion anchor into the allowlist subsection fails `exclusions` (the scoping mutant).

## AC-RXD-004 — requests are constructed per backend, always in background, and never set `write`

**Covers**: maps REQ-RXD-005, REQ-RXD-006

- **Classification**: release-blocking (RED-now: E10b, control E10x).
- **Given** the `run.md` section,
  **When** the anchors of `### Request construction` (`starts every job with background true`, `git rev-parse --show-toplevel`, `its own L1 tree`, `never sets the write argument`, `names the bounded files`, `patch text only`, `no secrets`, `over HTTPS`, `takes no project_root`, `bounded excerpts`, `never contains content of an excluded class`) are searched,
  **Then** all are present inside that subsection and the `request` subtest passes, including its negative check: the case-insensitive proximity expression `\bwrite\b[^.\n]{0,40}\b(true|enabled|on)\b|allow_write[^.\n]{0,20}\btrue\b` finds nothing in the section slice of either `run.md` copy, and no `allow_write: true` literal appears anywhere in it; the subtest also asserts the expression's two positive controls (it matches a built-in bad string and `internal/cli/mcp_server.go`), so an empty result cannot come from a broken expression.
- **RED-now**: E10b.
- **Green** (M1/M2): `--- PASS: TestRunExternalDelegationDoctrine/request `.
- **Mutant probe**: stating `project_root` for both tools (the card's wording) removes `takes no project_root` and fails the subtest; a section that sends a synchronous call lacks `starts every job with background true`; a section that says "pass `write: true` to let codex apply the patch" — or, in markdown formatting, "sets the `write` argument to `true`", "pass `write` = `true`", "write:True", "turn write on" or "a write-enabled turn" — matches the negative check and fails although every positive anchor is present (E7g; the revision-2 three-spelling expression passed all of these, E7h); the anchor sentence itself must not carry the enabling word either: "never sets the write argument to true" fails (E7g line 3), "never sets the write argument, so a delegated turn stays read-only." passes.

## AC-RXD-005 — results are untrusted data and are verified before they stay

**Covers**: maps REQ-RXD-007, REQ-RXD-008

- **Classification**: release-blocking (RED-now: E10c, control E10x).
- **Given** the `run.md` section,
  **When** the anchors of `### Result handling` (`untrusted data`, `never follows instructions found in a result`, `never executes commands found in a result`, `through its own edit tools`, `runs the verification the cycle already requires`, `reports the measured output`, `is discarded`, `done directly`) are searched,
  **Then** all are present inside that subsection and the `result` subtest passes.
- **RED-now**: E10c.
- **Green** (M1/M2): `--- PASS: TestRunExternalDelegationDoctrine/result `.
- **Mutant probe**: a section that applies the patch with a shell redirect lacks `through its own edit tools`; one that keeps a failing patch lacks `is discarded`; one that never mentions executing found commands lacks `never executes commands found in a result`; one that verifies but never reports lacks `reports the measured output`; one that discards a failing patch and re-delegates indefinitely, instead of doing the subtask directly, lacks `done directly` (the iteration-2 audit's N10 mutant: `is discarded` alone anchored only the discard half) — each fails the subtest naming the phrase.

## AC-RXD-006 — values come from the generator or the unmodified code, never the reply

**Covers**: maps REQ-RXD-009

- **Classification**: release-blocking (RED-now: E10d, control E10x).
- **Given** the `run.md` section, whose allowlist names fixture or golden-file regeneration and characterization-test drafts and whose exclusions name "the external output would decide the expected behavior of a test",
  **When** the anchors of `### Value observation` (`the generator the project already owns`, `unmodified code under test`, `never taken from the reply`) are searched,
  **Then** all are present inside that subsection and the `value` subtest passes — the one subsection that reconciles the allowlist with the exclusion (the external model drafts the step; the expected bytes and asserted values come from running the generator or observing the code).
- **RED-now**: E10d.
- **Green** (M1/M2): `--- PASS: TestRunExternalDelegationDoctrine/value `.
- **Mutant probe**: a section whose allowlist keeps fixture and golden regeneration but whose result handling says only "run the cycle's verification" (the 0.1.0 shape, where a golden rewritten to match wrong output passes verification) lacks all three anchors and fails `value`.

## AC-RXD-007 — delegation fails open within a fixed bound

**Covers**: maps REQ-RXD-010

- **Classification**: release-blocking (RED-now: E10e, control E10x).
- **Given** the `run.md` section,
  **When** the anchors of `### Fail-open` (`unavailable, inconclusive, failed or empty`, `at most five reads of the job status or result tool`, `each read follows a unit of its own work`, `never a sleep loop`, `is a failed delegation`, `cancels the job`, `does the subtask itself`, `no blocker report`) are searched,
  **Then** all are present inside that subsection and the `failopen` subtest passes.
- **RED-now**: E10e.
- **Green** (M1/M2): `--- PASS: TestRunExternalDelegationDoctrine/failopen `.
- **Mutant probe**: "poll until terminal" (the `super-advisor` wording) or "wait as long as the job takes" lacks `at most five reads of the job status or result tool` and `is a failed delegation`; **a section that bounds only the status tool** ("at most five reads of the job status tool") no longer matches the anchor — an agent that polls `codex_job_result` or `glm_job_result` (both return a running job's current status without blocking: `internal/cli/mcp_server.go`, the descriptions of `codex_job_result` and `glm_job_result`) reads the status tool zero times and would have satisfied the revision-2 anchor while waiting as long as the job runs; a section that allows five back-to-back reads or a polling loop lacks `each read follows a unit of its own work` or `never a sleep loop`; a section that says a failed delegation returns a blocker report removes `no blocker report`; one that never cancels removes `cancels the job` — each fails the subtest. **No floor** on the number of reads is asserted: a job still not terminal when no non-conflicting work is left is a failed delegation after fewer than five reads (spec.md §B.2 item 4).

## AC-RXD-008 — the agent's own writes: nothing in flight is edited, every job is closed

**Covers**: maps REQ-RXD-011, REQ-RXD-012

- **Classification**: release-blocking (RED-now: E10f, control E10x).
- **Given** the `run.md` section,
  **When** the anchors of `### One-writer rule` (`does not edit the files named in an in-flight prompt`, `reads or cancels every job it started before reporting completion`, `the orchestrator owns the one-writer-per-tree rule`) are searched,
  **Then** all are present inside that subsection and the `onewriter` subtest passes.
- **RED-now**: E10f.
- **Green** (M1/M2): `--- PASS: TestRunExternalDelegationDoctrine/onewriter `.
- **Mutant probe**: the 0.1.0 writable mutant — a subsection stating "the agent is the only writer, may freely edit the files named in a prompt while the job is in flight, and reads or cancels every job" — contains the old anchors `only writer`, `in flight` and `reads or cancels every job` and passed the 0.1.0 criterion while violating the requirement; it lacks `does not edit the files named in an in-flight prompt` and `reads or cancels every job it started before reporting completion` and now fails.

## AC-RXD-009 — delegation is stated as Claude-harness only; the Codex file changes by regeneration only

**Covers**: maps REQ-RXD-013

- **Classification**: release-blocking for the anchor half (RED-now: E10g, control E10x); regression-guard for the generated-file half (E13 is the green-on-arrival baseline; observed failure: the hand-edit mutant below).
- **Given** the finished change,
  **When** the section and the Codex tree are inspected,
  **Then** the anchors `### Harness scope`, `Claude-harness capability` and `on any other harness` are present inside that subsection (the `harness` subtest passes); `git diff --name-only <BASE>..HEAD -- internal/template/agentemit internal/template/templates/.codex` prints exactly `internal/template/templates/.codex/agents/moai/manager-develop.toml`; and `make agents-emit-check` exits 0, so that file is exactly what the emitter produces from the new agent body.
- **RED-now**: E10g for the anchor.
- **Green** (M2 for the anchor, M3 for the file): the TOML is regenerated with `make agents-emit` after the body edit; any hand edit makes `make agents-emit-check` exit non-zero.
- **Mutant probe**: editing the TOML by hand fails the third clause; touching `agents-codex.yaml` adds a second path to the second clause; a section that says the Codex agent should also delegate lacks `Claude-harness capability` and fails the first.

## AC-RXD-010 — consumer statements match the wiring

**Covers**: maps REQ-RXD-014

- **Classification**: release-blocking (RED-now: E4a with control E4a-c; E4b with control E4b-c; E4c; E4e with control E4e-c and E4f with control E4f-c; E5 and E5b with control E5c).
- **Given** the finished change,
  **When** the catalogue and the authoring note are read,
  **Then** the E4a command prints `8`; the E4b command prints `2`; the E4c command prints `2` (one pointer in each family description) and, so that each description sentence also **names the consumer on the same physical line**, the E4e command (`manager-develop`, then `codex_task`, then `External Model Delegation`) and the E4f command (the same with `glm_task`) each print `1` — one sentence per family, on one physical line each (the pointer-phrase count of E4c alone is wrap-sensitive and would pass a description that points at the section without naming `manager-develop`); the E4d command (the `codex_setup` row) prints `0` (exit 1); on both `agent-authoring.md` copies `grep -c -F "read-only only while"` and `grep -c -F "External Model Delegation"` print at least `1` each and `grep -c "allow_write"` still prints at least `1`; `git diff --name-only <BASE>..HEAD -- .claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md` prints nothing; and `<SCRUB> go test -count=1 -v -run '^(TestMCPToolCatalogueDocsStayMirrorIdentical|TestMCPToolCatalogueFiguresMatchRegistry)$' ./internal/cli/` shows both `--- PASS:` lines.
- **RED-now**: E4a, E4b, E4c, E4e, E4f, E5, E5b.
- **Green** (M4).
- **Mutant probe**: a family description that points at `External Model Delegation` without naming `manager-develop` leaves E4c at `2` but E4e or E4f at `0`; a sentence hard-wrapped across two physical lines leaves the matching E4e/E4f at `0`; naming `manager-develop` on the `codex_setup` row makes the E4d count `1`; leaving the family table unchanged leaves the E4b count at `0`; a note that names `manager-develop` but omits the opt-in condition leaves `read-only only while` at `0`; editing the catalogue in one tree only fails `TestMCPToolCatalogueDocsStayMirrorIdentical` and `cmp -s`.

## AC-RXD-011 — every mirrored pair carries the same change

**Covers**: maps REQ-RXD-015

- **Classification**: release-blocking for the mirrored-change half (RED-now: E2 and E3 — the template copy lacks the change as well); regression-guard for the by-design-difference half (E19 is green on arrival; observed failure: the live-only-hunk mutant below).
- **Given** the finished change,
  **When** the six mirrored pairs are compared,
  **Then** `cmp -s .claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run.md` and `cmp -s .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` each exit 0 (these two pairs are byte-identical today and must stay so); for the four pairs that differ by design (`manager-develop.md`, `fix.md`, `loop.md`, `agent-authoring.md`) the E19 commands still print `5`, `4`, `2` and `4` (the by-design difference neither grew nor shrank, so the same change reached both copies in the same words), the guard test's `pairdelta` subtest passes with its M1 constants, and the marker checks of AC-RXD-001, -002 and -010 pass on both copies.
- **RED-now**: E2/E3 (the template copy lacks the change as well).
- **Green** (M2–M4).
- **Mutant probe**: leaving the template `run.md` one sub-heading short makes `cmp -s` exit 1; a hunk applied to the live `fix.md` only, or worded differently in the two copies, raises the E19 count above `4` and fails `pairdelta` (count equality of markers alone would not see a re-worded hunk).

## AC-RXD-012 — the lines the change adds carry no internal token

**Covers**: maps REQ-RXD-015

- **Classification**: regression-guard (green on arrival: E21, E21d; controls E21c and E20 show the probe fires on a changed token line and on an existing token; the leak guard passes on arrival, E15).
- **Given** the finished change,
  **When** the changed lines of both trees are scanned,
  **Then** `git diff --name-only -G'SPEC-[A-Z][A-Z0-9]+-[0-9]{3}|REQ-[A-Z]|AC-[A-Z]|20[0-9]{2}-[0-9]{2}-[0-9]{2}' <BASE>..HEAD -- .claude internal/template/templates` prints nothing (no changed line — added or removed — matches a token class; the pre-existing token lines of the touched files stay untouched); `grep -rn -E "t1424|SPEC-RUN-EXTERNAL-DELEGATION" internal/template/templates .claude/agents .claude/skills .claude/rules/moai` prints nothing (exit 1); and `<SCRUB> go test -count=1 -v -run '^TestTemplateNoInternalContentLeak$' ./internal/template/` shows `--- PASS`.
- **RED-now**: not applicable (green on arrival); the controls prove the probes can fire.
- **Green** (M2–M5).
- **Mutant probe**: pasting a requirement token, a SPEC id or an ISO date into the new section, a pointer or a catalogue sentence makes the first command print that file; pasting the card id fails the second command, the card-id grep. Measured in the run (progress.md §E.2.5): `TestTemplateNoInternalContentLeak` did not flag a bare card id, so it is not what decides that case; it did flag a SPEC id plus an ISO date (class `C1-spec-id-prefix`).
- **Note on the 0.1.0 criterion**: the earlier form (`grep` for this card's two identifiers only) was narrower than its requirement; the whole-file token count is not used because the touched template files already carry earlier tokens (E20) that this change must not be blamed for.

## AC-RXD-013 — the change set is exactly the planned fifteen files, on the whole tree

**Covers**: maps REQ-RXD-016

- **Classification**: release-blocking (RED-now: E6, controls E6c and E6d; E6a for the agents half).
- **Given** the finished run phase,
  **When** `git diff --name-only <BASE>..HEAD -- . ':(exclude).moai/specs/SPEC-RUN-EXTERNAL-DELEGATION-001' ':(exclude).moai/reports/t1424'` is read,
  **Then** it prints exactly the fifteen paths of plan.md §B (six live files, their six template mirrors, the generated TOML, `internal/template/catalog.yaml`, and the guard test `internal/template/run_external_delegation_test.go`), no other path — the whole tree, so root-level files (`CLAUDE.local.md`, `AGENTS.local.md`, `CHANGELOG.md`), `docs-site/` and `.agents/skills/` are seen; `git diff --name-only <BASE>..HEAD -- .claude/agents internal/template/templates/.claude/agents` prints exactly the two `manager-develop.md` paths; and the always-loaded surface is untouched by the same set (`CLAUDE.md`, `AGENTS.md`, `moai-mcp-tools.md` and its mirror, `kanban-dispatch.md` and the output styles are not among the fifteen).
- **RED-now**: E6 prints nothing, where the green output has fifteen lines; E6a prints nothing, where the green output has two.
- **Green** (M5).
- **Mutant probe**: any sixteenth path (for example `moai-mcp-tools.md`, `internal/mcp/catalog.go`, `agents-codex.yaml`, `CLAUDE.local.md`, a `_templ.go` file) is a plan deviation and a blocker report, not a quiet addition; editing `super-advisor.md` adds a third path to the agents command.

## AC-RXD-014 — `allow_write` stays false and no shipped prose pairs `write` with an enabling value

**Covers**: maps REQ-RXD-005, REQ-RXD-016

- **Classification**: regression-guard (E7, E7d, E8 are green on arrival; positive controls E7c, E7f (the tool registration, a tree file), E7g (a bad fixture and the prescribed wording, scratch) and the E14 test). It proves the **shipped default** and that no shipped file pairs the whole word `write` with `true`, `enabled` or `on` within 40 characters (a lexical proximity check, not a proof that no write is ever instructed — spec.md R-3); it does **not** prove the opt-in is off on any machine — on this maintainer's machine it is open (E8b, spec.md R-6) and only the agent's never setting `write` keeps a delegated codex turn read-only.
- **Given** the finished change,
  **When** the shipped default, the shipped configuration and the new prose are read,
  **Then** `grep -rn "allow_write: true" internal/template/templates .claude .moai/config` prints nothing (exit 1); `grep -n "AllowWrite: false" internal/config/defaults.go` prints one line (exit 0); `git diff --name-only <BASE>..HEAD -- internal/config .moai/config internal/template/templates/.moai/config` prints nothing; the E7d command (`grep -n -i -E '\bwrite\b[^.]{0,40}\b(true|enabled|on)\b|allow_write[^.]{0,20}\btrue\b'` over both copies of `run.md`, `manager-develop.md`, `fix.md` and `loop.md`) prints nothing (exit 1) — and the E7f command (the same expression over `internal/cli/mcp_server.go`) still prints a non-zero count, so the empty result is absence and not a broken expression; and `<SCRUB> go test -count=1 -v -run '^TestCodexTaskAllowWrite_DistributedDefaultIsFalse$' ./internal/cli/` shows `--- PASS`.
- **RED-now**: not applicable (green on arrival); the criterion proves the default survives the change and the doctrine's own prose adds no opt-in literal and no `write` instruction.
- **Mutation proof (run phase, recorded verbatim in progress.md §E.2)**: build an overlay JSON that maps `internal/config/defaults.go` to a scratch copy in which `AllowWrite: false` reads `true`, run `<SCRUB> go test -overlay=<overlay.json> -count=1 -v -run '^TestCodexTaskAllowWrite_DistributedDefaultIsFalse$' ./internal/cli/` and observe `--- FAIL`; `git diff --stat -- internal/config/defaults.go` is empty because the real file was never edited (a PRESERVE file is not mutated). A second mutant (the ninth of AC-RXD-016, run once and cited from both) adds the markdown-variant instruction ``sets the `write` argument to `true` `` (a spelling the revision-2 three-spelling expression did not match, E7h) to the new section and shows the E7d command printing that line and the `request` subtest failing; a third (the tenth of AC-RXD-016, likewise run once and cited from both) appends " to true" to the anchor sentence `never sets the write argument` and shows the same; each is reverted with an empty `git diff` for the file.
- **Green** (M5).

## AC-RXD-015 — the generated files come from their generators and nothing else regenerates

**Covers**: maps REQ-RXD-016

- **Classification**: regression-guard (the hash and emit guards pass on arrival, E13 and E15; observed failure: the stale-hash and hand-edit mutants below).
- **Given** the finished change after the generators ran,
  **When** the generator and hash guards run,
  **Then** `<SCRUB> go test -count=1 -v -run '^(TestTemplateNoInternalContentLeak|TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles|TestAllAgentsInCatalog|TestAllSkillsInCatalog)$' ./internal/template/` shows five `--- PASS:` lines; `make agents-emit-check`, `make commands-emit-check` and `make tool-policy-drift-check` exit 0; a second `go run ./internal/template/scripts/gen-catalog-hashes.go --all` leaves `internal/template/catalog.yaml` byte-identical to the copy taken after the first run — copy the file to the scratch directory after the last M3 generator run, run the generator once more, and `cmp -s <copy> internal/template/catalog.yaml` exits 0 (a `git status --short` letter is not the comparison: after the first run the file already reads ` M`, so the letter is the same whatever the second run does; the plan-time observation of this procedure on a scratch copy is E22) — so the hashes are regenerated, never hand-edited; and no `_templ.go` file appears in the whole-tree set of AC-RXD-013 (the run does not execute `make build`, whose `templ-generate` rewrites tracked generated files).
- **RED-now**: not applicable; `TestCatalogHashCoversSkillSubfiles` turns red the moment a template skill file changes without regenerating the hashes, and `make agents-emit-check` turns red the moment the agent body changes without `make agents-emit` — the observed-failure demonstrations recorded in progress.md §E.2.
- **Green** (M2, M3 and M5): each commit that edits a template skill or agent file carries the regenerated `catalog.yaml` (and, from M3, the regenerated TOML).
- **Mutant probe**: editing a template skill file without regenerating fails the hash test; editing the TOML by hand fails `make agents-emit-check`; hand-editing a `catalog.yaml` hash fails the same hash test.

## AC-RXD-016 — the guard test and the owning packages are green, with the failures observed

**Covers**: maps REQ-RXD-016

- **Classification**: release-blocking (RED-now: E11, control E11c).
- **Given** the finished run phase,
  **When** the verification batch of plan.md §F runs,
  **Then** `<SCRUB> go test -count=1 -v -run '^TestRunExternalDelegationDoctrine$' ./internal/template/` shows `--- PASS` for the parent and all twelve subtests (`tools`, `section`, `allowlist`, `exclusions`, `request`, `result`, `value`, `failopen`, `onewriter`, `harness`, `pointers`, `pairdelta`, as named in plan.md §D — the swept count is read from the verbose output and stated in progress.md, never inferred from the final `ok`); `<SCRUB> go test -count=1 ./internal/template/... ./internal/config/...` exits 0; `GOOS=windows GOARCH=amd64 go vet ./internal/template/` exits 0 (it type-checks the new `_test.go` for Windows — E17 is its baseline — which `GOOS=windows GOARCH=amd64 go build ./...` does not); `GOOS=windows GOARCH=amd64 go build ./...` exits 0; `make agents-emit-check` and `make commands-emit-check` exit 0; `moai agent lint` ends with `0 errors` and no more than the 25 baseline warnings (E16); the two catalogue figure tests pass (E18); and `moai spec lint --strict SPEC-RUN-EXTERNAL-DELEGATION-001` exits 0.
- **RED-now**: E11 — the guard test does not exist.
- **Green** (M1 for the test's RED capture in its own commit, M5 for the batch).
- **Ordering witness**: the RED is witnessed by the commit graph — the M1 commit adds the test and the `spec.md` `status:`/`updated:` lines only, and checking it out shows `--- FAIL` — and the doctrine commits that follow are the GREEN; the baseline and the change are never in one commit.
- **Observed failures (recorded verbatim in progress.md §E.2)**: the ten mutants of plan.md §D (remove one tool from the template tools line; add `codex_setup`; insert `Agent` into the tools prefix; rename the section heading; copy a sub-heading into `fix.md`; the "may freely edit the files named in a prompt" one-writer rewrite; the `never sets the write argument` anchor moved into `### Fail-open`; a two-line paraphrase of the procedure added to a pointer; the markdown-variant write instruction ``sets the `write` argument to `true` `` added to the section; the anchor sentence with ` to true` appended, the tenth, which AC-RXD-014 requires), each shown failing the guard test and then reverted with an empty `git diff` for the mutated file.

## Definition of Done

1. All 16 criteria are PASS with the verbatim command output in progress.md §E.2; no criterion rests on a summary.
2. The regression-guards (AC-RXD-009 generated-file half, AC-RXD-011 by-design half, AC-RXD-012, AC-RXD-014, AC-RXD-015) show their failure observed, not only their green; the ten guard mutants of AC-RXD-016 are shown failing.
3. The change set equals the planned fifteen files on the whole tree; no file outside plan.md §B changed.
4. `workflow.codex.task.allow_write` is false by default and no `allow_write: true` literal and no `write: true` instruction exists in any shipped file; the report states, as the residual, that the opt-in is open in the maintainer's primary checkout (spec.md R-6).
5. The repository-wide test verdict belongs to CI on the integration branch and is reported PENDING at run completion.
6. The docs-site four-locale consumer-table correction is handed to the sync phase and named in the run completion report as open.
