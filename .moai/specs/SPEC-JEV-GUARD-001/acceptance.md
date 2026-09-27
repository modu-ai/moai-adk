---
id: SPEC-JEV-GUARD-001
title: "acceptance — Jev Consumer B withdrawal"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
---

# acceptance.md — SPEC-JEV-GUARD-001

## D. AC Matrix

| AC | Scenario | Severity | REQ |
|----|----------|----------|-----|
| AC-JEVG-001 | Guard test green | Blocker | REQ-JEVG-001 |
| AC-JEVG-002 | SkillSuggest marker absent from non-test Go | Blocker | REQ-JEVG-001 |
| AC-JEVG-003 | Guard test byte-identical | Blocker | REQ-JEVG-003 |
| AC-JEVG-004 | SKILL.md prose withdrawal, no divergence widening | Blocker | REQ-JEVG-005 |
| AC-JEVG-005 | Build + affected suites green; command unregistered | Blocker | REQ-JEVG-002, 004 |
| AC-JEVG-006 | Shared symbols intact | Blocker | REQ-JEVG-004 |
| AC-JEVG-007 | Lint clean on changed packages | Blocker | REQ-JEVG-001 |
| AC-JEVG-008 | Restoration contract recorded and mechanically backed | High | REQ-JEVG-006 |

## D.1 AC-JEVG-001 — Guard test GREEN (RED→GREEN pair)

**RED-now cell** (observed at plan phase, pinned):

- **Command**: `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips`
- **Verbatim stdout**: `gate_demo_test.go:137: a consumer call path is present before its measurement: [SkillSuggest in /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1083/internal/cli/jev_skill_suggest.go]` plus `FAIL` lines (full text in research.md §1)
- **Exit code**: 1
- **Tree SHA**: `cd99336bf`

**Given** the withdrawal of M1 is applied, **When** the same command runs on the post-M1 tree, **Then** it exits 0 and prints no consumer-marker failure. The flip is red-for-the-right-reason: the only failing marker was SkillSuggest, the only code removed.

## D.2 AC-JEVG-002 — Marker absence

**Given** the post-M1 tree, **When** `grep -rn 'SkillSuggest' --include='*.go' internal/ cmd/ | grep -v _test` runs, **Then** it returns 0 hits (exit 1 from grep is the pass signal for zero matches — the count, not the exit code alone, is the evidence). The other two markers (`NearDuplicateMark`, `LaneQuestionRoute`) already hold absence and remain absent.

## D.3 AC-JEVG-003 — Guard test integrity

**Given** the post-M1 tree, **When** `git diff <pre-M1-SHA> -- internal/jevmeasure/gate_demo_test.go` runs, **Then** the diff is empty (byte-identical). No marker-list edit, no walk-scope change, no build tag, no skip logic anywhere in the diff of any `*_test.go` under `internal/jevmeasure/`.

## D.4 AC-JEVG-004 — SKILL.md prose withdrawal without divergence widening (executable)

**Given** the post-M1 tree, **When** the following three commands run, **Then**:

(a) `grep -c 'jev-suggest' .claude/skills/moai/SKILL.md` → prints 0;
(b) `grep -c 'jev-suggest' internal/template/templates/.claude/skills/moai/SKILL.md` → prints 0;
(c) no-widening, shape-pinned and executable:

```bash
diff <(sed 's|\${CLAUDE_SKILL_DIR}|.claude/skills/moai|g' .claude/skills/moai/SKILL.md) \
     internal/template/templates/.claude/skills/moai/SKILL.md \
  | grep '^[<>]' | grep -vc -e 'moai c[gc] -w' -e 'Last Updated:'
```

→ prints **0**. The in-process `sed` normalizes comparison categories A+B (`${CLAUDE_SKILL_DIR}` → literal path), so the only divergence that may remain is the pre-existing baseline pair (C: the `moai cg -w`/`moai cc -w` hunk; D: the local-only `Last Updated:` deletion — spec.md §D.3). Any content line outside those shapes — including any *new* divergence introduced by this SPEC's edit — makes the count non-zero and fails the AC. The check cannot false-fail a correct edit (removing identical lines from both files adds no content lines) and cannot be satisfied by "normalizing" the files themselves (the sed normalizes the comparison only; editing either file to erase categories A-D is the axis violation plan.md §G prohibits and is not measured as a pass here).

## D.5 AC-JEVG-005 — Build, suites, and command unregistration

**Given** the post-M1 tree, **When** (a) `go build ./...` runs, **Then** exit 0; (b) `go test ./internal/jevmeasure/... ./internal/cli/...` runs, **Then** exit 0 (Consumer A/C suites pass — shared-symbol integrity observed behaviorally); (c) `go build -o /tmp/t1083-bin ./cmd/moai && /tmp/t1083-bin jev-suggest --help` runs, **Then** the subcommand resolves as unknown (non-zero exit, no `jev-suggest` usage text). Scratch binary under `/tmp`, cleaned after the check.

## D.6 AC-JEVG-006 — Shared symbols intact

**Given** the post-M1 tree, **When** `grep -n 'func jevNotice' internal/cli/todo_jev_finding.go` and `grep -n 'func jevEnabled' internal/cli/doctor_jev.go` run, **Then** both definitions are present, and the `./internal/cli/...` suite of AC-JEVG-005 passes with Consumer A tests intact (`installJevProbe` still resolves inside `todo_jev_finding_test.go`).

## D.7 AC-JEVG-007 — Lint

**Given** the post-M1 tree, **When** `golangci-lint run ./internal/cli/... ./internal/jevmeasure/...` runs, **Then** it exits clean (no new findings attributable to this SPEC's diff).

## D.8 AC-JEVG-008 — Restoration contract recorded and mechanically backed

**Given** the post-M1 tree and the restoration contract REQ-JEVG-006, **When** the SPEC set is inspected, **Then**:

(a) spec.md §B carries REQ-JEVG-006 naming all three restoration preconditions (t1066 F2 repair, gate run, baseline beaten) and both mandatory successor-SPEC acts (history restoration + guard-mechanism evolution under MEASURE ownership), and spec.md §F carries the chain linking REQ-JEVG-006 ↔ REQ-JEVN-016 ↔ REQ-JEVO-009/AC-JEVO-012 ↔ the guard test;
(b) the contract is enforced by the guard test, not by prose alone: until a successor SPEC performs both acts, any re-introduction of the SkillSuggest marker — by any actor, under any gate state — trips AC-JEVG-001/AC-JEVG-002. Verification of (b) at this SPEC's close is AC-JEVG-001 and AC-JEVG-002 passing (the guard is live against exactly the re-landing the contract governs). This AC is forward-looking by nature; its closing observation is (a)'s presence check plus the live guard, and its future obligation binds the successor SPEC named in REQ-JEVG-006.

## Edge cases

- **Gate still disabled** (`workflow.jev.enabled: false` is the default): withdrawal removes the command entirely, so the gate flag becomes inert for Consumer B — no dead-flag defect; the flag continues to serve `jev_ask` (MCP) and the doctor surface.
- **Stale installed binary**: `/tmp` scratch binary is built fresh in AC-JEVG-005; an installed `~/go/bin/moai` predating the fix is not evidence either way (tool-provenance rule).
- **Partial removal** (impl deleted, root.go registration left): caught by `go build` (undefined `newJevSuggestCmd`) — the build is the second net under the grep.

## Quality gates

- TRUST 5 Tested: RED→GREEN pair observed (AC-JEVG-001); affected suites green (AC-JEVG-005).
- TRUST 5 Trackable: conventional commit subjects naming SPEC-JEV-GUARD-001 + card t1083.

## Definition of Done

1. AC-JEVG-001 through AC-JEVG-008 all PASS with verbatim command outputs recorded in progress.md §E.2.
2. The withdrawal set (spec.md §D.1) is the complete diff footprint — no file outside it changed.
3. Restoration contract (REQ-JEVG-006) is recorded and citable by the future re-landing SPEC.
4. Lane hands off to sync with the codemap-regeneration obligation named.
