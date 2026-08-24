# SPEC-SESSION-TELEMETRY-001 — Acceptance Criteria

Eleven criteria against nine requirements. Every criterion names a command and an expected
result; none is satisfied by reading a file and forming a judgement.

Where a criterion is satisfied by an **absence**, the pre-change baseline is stated, measured in
this tree at `dfbf828a6`, so a zero-hit result is new information rather than a vacuous pass. A
criterion that already passes on the untouched tree is a defect, not a criterion.

## §A The per-session record

**AC-ST-001** (REQ-ST-001) — Given a statusline render for session `S` under project root `P`,
When the builder completes, Then `P/.moai/state/context-usage/S.json` exists **and**
`P/.moai/state/context-usage.json` does not. Baseline: on the pre-change tree the first path
never exists (`ls -d .moai/state/context-usage` → `No such file or directory`) and the second
always does after any render, so both halves change state. The second half asserts the **hard
cut** and is correct only because D-3 chose a hard cut over a dual-write window; under a
dual-write window it would fail by construction. Written this way by decision, not by
presupposition.

**AC-ST-002** (REQ-ST-002) — Given a render input whose payload `session_id` is `S` and a
`.moai/state/current-session-id.txt` containing a **different** id `T`, When the builder
completes, Then the written file is `…/context-usage/S.json` and no file named for `T` exists;
and When `grep -rn 'CurrentSideChannelFile\|current-session-id' internal/statusline/` is run,
Then it returns zero hits. Baseline: the same grep returns zero on the pre-change tree, so the
grep half asserts preservation — the load-bearing half is the two-id fixture, which cannot pass
before the change because the pre-change writer produces one path regardless of either id.

**AC-ST-003** (REQ-ST-004) — Given a render input whose `model.display_name` is `Opus 5 (1M
context)` and an environment in which `ANTHROPIC_DEFAULT_OPUS_MODEL` names a z.ai model, When
the record is written, Then its model field equals that z.ai model name, carries no `[1m]`
suffix, and is not the Claude display name. Baseline: no model field exists on the pre-change
record (`grep -n 'model' internal/statusline/context_usage.go` returns no struct field), so the
assertion has no pre-change form.

**AC-ST-004** (REQ-ST-003) — Given a render input carrying `effort.level` and
`model.display_name`, When the record is written and read back, Then both values round-trip
unchanged; and Given a render input carrying neither, When the record is written, Then neither
key appears in the marshalled JSON (`omitempty`) and the write still succeeds.

**AC-ST-010** (REQ-ST-003) — Given a record file **produced by the pre-change writer** — that
is, generated in the test by marshalling the pre-change struct, not hand-authored — When it is
read by the post-change reader, Then the read succeeds, the context values are present, and the
model and effort values are reported as not recorded. The input is constrained to
writer-produced bytes so key order and indentation match the marshaller's own output; a
hand-indented fixture would fail a correct implementation for reasons the requirement does not
care about.

**AC-ST-008** (REQ-ST-007) — Given each of the four hostile key values `""`, `"../escape"`,
`"a/b"`, and `"/tmp/absolute"`, When a render is performed for each, Then (a) the count of files
created anywhere outside `P/.moai/state/context-usage/` is zero, measured by snapshotting the
tree before and after and diffing the file lists; (b) no file is created inside that directory
for these keys either; and (c) each render still returns its rendered statusline line with a
zero exit status. Baseline: on the pre-change tree the key is not a path component at all — the
filename is a constant — so no such refusal exists to preserve; all three halves are new.

## §B One reader

**AC-ST-005** (REQ-ST-005) — Given the merged tree, When
`grep -rEc '^func Read[A-Za-z]*ContextUsage' internal/statusline/*.go` is summed, Then the count
is **exactly 1**. Baseline: the same command returns **0** on the pre-change tree —
`readContextUsage` (`context_usage.go:186`) is unexported and no exported sibling exists — so
the criterion is new information rather than a restatement of the status quo.

**AC-ST-006** (REQ-ST-005) — Given the merged tree, When `grep -rn '"raw_pct"' internal/` is run
— scope pinned to `internal/`, **including** `_test.go` files — Then every hit lies inside
`internal/statusline`. Baseline, measured in this tree:

```
$ grep -rn '"raw_pct"' internal/
internal/statusline/context_usage.go:63        (struct field tag — stays)
internal/statusline/context_usage_test.go:150  (schema assertion — stays)
internal/cli/tokens.go:86                      (duplicate declaration — removed by REQ-ST-006)
internal/cli/tokens_test.go:283                (its fixture — moves with it)

$ grep -rln '"raw_pct"' internal/ | grep -v '^internal/statusline' | wc -l
2
```

So the outside-`internal/statusline` file count goes **2 → 0**, and the post-change assertion is
4 hits in 2 files.

**AC-ST-007** (REQ-ST-006) — Given the merged tree, two halves. **Removal:** When
`grep -rn 'context-usage' internal/cli/tokens.go` is run, Then it returns exactly one hit — the
command help string at `:426`, which the change does not touch — and specifically returns
neither the filename constant nor a struct declaration. Baseline, measured:
`grep -rn 'context-usage' internal/cli/` returns four hits — `tokens.go:30` (the constant,
removed), `tokens.go:79` (the struct's **doc comment**; the declaration itself is at `:81`),
`tokens.go:426` (the help string, out of scope), and `tokens_test.go:284` (the fixture path,
migrated). **Preservation:** Given a session with a readable per-session record, When
`moai tokens` is run, Then its output still carries the context block with the record's
`raw_pct`; and given no readable record, Then it exits zero with the block absent (fail-open,
matching `readTokensContextSnapshot`'s current contract at `tokens.go:393-397`).

## §C Consumers of the moved path

**AC-ST-009** (REQ-ST-008) — Given the merged tree, three halves. (a) When
`grep -rln "state/context-usage.json" .claude internal/template/templates` is run, Then it
returns zero files. (b) When the same paths are grepped for `state/context-usage/`, Then they
return exactly the **same four files** the baseline named. (c) When each mirror pair is compared
with `diff -q`, Then both comparisons print nothing and exit zero. Baseline, measured:

```
$ grep -rln "state/context-usage.json" .claude internal/template/templates
.claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md
```

and both `diff -q` comparisons already print nothing today, so half (c) asserts preservation —
it fails if the change updates one side of a pair and not the other, which is the actual hazard.

**AC-ST-011** (REQ-ST-009) — Given the merged tree, three halves. (a) When
`grep -rln "context-usage.json" docs-site/content | wc -l` is run, Then it returns **0**.
(b) When `grep -rln "context-usage/" docs-site/content | wc -l` is run, Then it returns **12**.
(c) When those twelve paths are grouped by locale, Then each of `en`, `ko`, `ja`, `zh` carries
exactly three. Baseline, measured: half (a) returns **12** today and half (b) returns **0**, and
the twelve are three per locale —
`content/{en,ko,ja,zh}/advanced/statusline.md`, `advanced/token-budget.md`,
`cli-reference/tokens.md`. This criterion exists because the parent SPEC enumerated these twelve
files and gave them no requirement, no criterion, and no Definition-of-Done entry, so a run
could have passed everything and left twelve published pages naming a path that no longer
exists.

## §D Traceability

| Requirement | Criteria |
|---|---|
| REQ-ST-001 | AC-ST-001 |
| REQ-ST-002 | AC-ST-002 |
| REQ-ST-003 | AC-ST-004, AC-ST-010 |
| REQ-ST-004 | AC-ST-003 |
| REQ-ST-005 | AC-ST-005, AC-ST-006 |
| REQ-ST-006 | AC-ST-007 |
| REQ-ST-007 | AC-ST-008 |
| REQ-ST-008 | AC-ST-009 |
| REQ-ST-009 | AC-ST-011 |

Nine requirements, eleven criteria. Every requirement carries at least one criterion; every
criterion names exactly one requirement as its parent.

## §E Edge cases

- **Two sessions rendering concurrently.** Distinct keys, distinct files, no shared slot — the
  case AC-ST-001 makes structurally impossible to lose. No locking is introduced; the atomic
  temp-file-plus-rename write the package already performs is per-file and remains sufficient.
- **A session whose payload omits `model` but carries `effort`.** Half-populated is a legitimate
  record: the present value is recorded, the absent one is reported as not recorded (AC-ST-004).
- **The per-session directory does not exist yet.** Created on first write, exactly as the
  single `state/` directory is created today; a creation failure is silent and the render
  completes (the existing best-effort contract).
- **Records accumulating for dead sessions.** No reaper is introduced. Each file is a few hundred
  bytes and the directory is gitignored; a reaper is a separate change with its own liveness
  question, and adding one here would make the disposal policy an unreviewed side effect of a
  path move.

## §F Definition of Done

- [ ] All eleven criteria pass, each with its cited command's verbatim output.
- [ ] Every absence-criterion's baseline re-measured at merge, not carried from this document.
- [ ] `go test ./internal/statusline/... ./internal/cli/...` passes; the full-suite verdict is
      read from CI on the pull-request head.
- [ ] `go vet ./...` and `golangci-lint run` clean on the touched packages.
- [ ] Both doctrine mirror pairs updated in the same commit, followed by `make build`; the
      template-neutrality guard passes.
- [ ] All twelve docs-site pages updated, three per locale, with a warning-free site build.
- [ ] `internal/spec/drift_cache.go:24`'s comment updated in the same sweep.
- [ ] No file under `internal/web` modified (this SPEC's exclusion boundary).
