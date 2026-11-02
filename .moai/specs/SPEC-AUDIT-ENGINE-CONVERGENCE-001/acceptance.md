# Acceptance — SPEC-AUDIT-ENGINE-CONVERGENCE-001

## §A Adoption contract

Every AC below is adopted as a pair (`verification-completeness.md` §2):
- a RED-now cell measured on tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d` (a document-level pin, binding every entry that names no tree of its own);
- a green-path cell naming the milestone that flips it.

Each AC also names the mutant it blocks.

There are two classes:
- **release-blocking**: the RED was executed in this plan phase as a read-only single invocation, and it can be re-executed on the pinned tree.
- **paired guard**: a negative sub-criterion whose subject does not exist yet, so it passes vacuously today. It is adopted only together with the release-blocking sibling named in its row, and it is never recorded as a pass on its own.

Each RED below is a source observation of the defective path, not the AC's whole verdict. Every green cell pairs that observation's inverse with a named behavior test, so deleting the text without fixing the behavior does not pass.

## §B AC Matrix

| AC | Given-When-Then | REQ | Class | Evidence | Milestone |
|---|---|---|---|---|---|
| AC-AEC-001a | Given a delta in effect on `audit_multi`, When the codex leg sends its request, Then the request carries the delta SHA as its review target (or carries the `delta...HEAD` diff), and the scope note contains neither "prose-only" nor "prose only". | REQ-AEC-003 | release-blocking | EV-01 | M3 |
| AC-AEC-001b | Given a fixture card, When `audit_multi` is called with its `card_id` and no `delta_base`, Then: with no verdict file at all (first review), or a newest verdict that is PASS, or an `audited_sha` that is a non-ancestor or conflicting, the full scope is reviewed and the result says why; with a FAIL verdict carrying an ancestor `audited_sha`, that SHA is proposed and named with its file; an explicit SHA replaces the file's; `delta_base: "full"` reviews the resolved base. | REQ-AEC-001, REQ-AEC-002 | release-blocking | EV-02 | M3 |
| AC-AEC-001c | Given a chain of server-written second-review records (full-scope R0, delta R1 whose base is R0's head), each with a matching receipt, and a complete closure confirmation, When the proposed base equals R1's head, Then the delta scope is reviewed; When any link's base differs from the previous record's head (a gap), or the proposed base differs from the last record's head, or no full-scope record starts the chain, Then the full scope is reviewed and the result names the broken link. | REQ-AEC-004 (i) | release-blocking | EV-03 | M1 |
| AC-AEC-001d | Given an unbroken chain, When the call omits the BLOCKING-closure confirmation, or names only some `blocking_defect:` ids of the newest FAIL verdict, or names a closing commit outside `delta_base..HEAD`, or the id count differs from `blocking_count`, Then the full scope is reviewed and no delta-scope PASS is produced; When the confirmation is complete, Then every leg's request carries it. | REQ-AEC-004 (ii) | paired guard (with 001c) | EV-04 | M1 |
| AC-AEC-001e | Given a chain whose link record cites a receipt absent from the receipt store, or a receipt whose tool, tree root, or head commit disagrees with the record, or chain data supplied as a call argument or present only in a verdict file, When `audit_multi` evaluates the chain, Then that chain is refused and the full scope is reviewed. | REQ-AEC-004 (i) | paired guard (with 001c) | EV-03 | M1 |
| AC-AEC-001f | Given a delta call admitted by the chain gate, When it completes, Then exactly one second-review record is appended: scope base equals the delta base, it carries the delta marker and the receipt id of the record it extends, and `SelectSecondReview` keeps it. Given a delta call the gate sent to full scope, Then its record is an ordinary full-scope record. A full-scope call's record is byte-shaped as today. | REQ-AEC-005 | release-blocking | EV-05 | M1 |
| AC-AEC-002a | Given leg stubs that fail with a quota message, a deadline, and an authentication refusal, When `audit_multi` converges, Then each inconclusive leg carries `error_class` `quota`, `timeout`, `auth`, its `verdict` stays `inconclusive`, the receipt carries the class, and a `pass` or `fail` leg carries no class. | REQ-AEC-006 | release-blocking | EV-06 | M4 |
| AC-AEC-002b | Given a verdict file labelled `INCONCLUSIVE` with `required_backend: codex inconclusive quota`, When the admission rule reads it, Then it refuses with a reason naming `codex` and `quota`; Given a report directory holding iteration files labelled FAIL, INCONCLUSIVE, FAIL, When the round counter runs, Then it counts 2; and the convention's label line lists `INCONCLUSIVE`. | REQ-AEC-007 | release-blocking | EV-07 | M4 |
| AC-AEC-002c | Given a persisted convergence result for session S at tree state T, When `audit_multi` re-runs only the codex leg for S at T, Then the claude and glm leg call counters are 0 and their verdicts are carried; When the tree state differs from T, Then all legs run; When the caller supplies a verdict for a carried leg, Then it is ignored. | REQ-AEC-008 | paired guard (with 002a) | EV-06 | M4 |
| AC-AEC-003a | Given a receipt judged at commit C, When commits follow C that touch only `.moai/reports/t1/verdict.md`, `.moai/reports/t1/second-review.jsonl`, the `§E.1`/`§E.4` lines of `.moai/specs/SPEC-X-001/progress.md`, and — in a separate fixture — a full rewrite of that file's `§E.3` section with `§E.2` byte-unchanged, Then the reuse check accepts the receipt in each fixture (the `§E.3` rewrite inherits the receipt). | REQ-AEC-009 | release-blocking | EV-08 | M5 |
| AC-AEC-003b | Given the same receipt, When the following commit touches any one of: a `.go` file; `spec.md`; `.moai/reports/t1/evidence-ledger.md`; `.moai/reports-x/a`; exactly one changed line under `§E.2` of `progress.md` with every other section unchanged; a rename of a report into `internal/`; or HEAD is not a descendant of C; or an uncommitted code edit exists; or the receipt predates the new field, Then the reuse check refuses with `CauseReceiptTreeStateChanged`. A receipt whose head equals HEAD is judged exactly as before. | REQ-AEC-010 | paired guard (with 003a) | EV-08 | M5 |
| AC-AEC-004a | Given a tree with rejection marks for SPEC-A and SPEC-B, a terminal on standard input (seam), no agent markers, and an evidence file under `.moai/reports/` naming SPEC-A, When the verb clears the SPEC-A mark with a correct typed confirmation, Then one approval record is appended with the operator name, git identity, mark fields, evidence path and SHA-256, HEAD and UTC time; the SPEC-A mark is gone; the SPEC-B mark is byte-identical; and the spawn check no longer lists SPEC-A. | REQ-AEC-013 | release-blocking | EV-09 | M2 |
| AC-AEC-004b | Given any one of: `CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`, or the factory role marker set; standard input not a terminal (including `</dev/null`); a mark name that does not exist; a mark recorded for another tree in the same store; an evidence file outside `.moai/reports/` or not naming the mark's SPEC ID; or a typed confirmation naming another agent type or SPEC, When the verb runs, Then it refuses, the approvals log and every mark are unchanged. | REQ-AEC-012 | paired guard (with 004a) | EV-09 | M2 |
| AC-AEC-004c | Given an approval line appended to the approvals log by hand, Then the spawn check still denies; and given the built server's `tools/list` plus a source sweep of callers of the new store function, Then no MCP tool clears a rejection and the function has exactly one non-test caller, the verb. | REQ-AEC-011 | paired guard (with 004a) | EV-09 | M2 |
| AC-AEC-005 | Given the four agent files, Then none says an unmet gate returns or makes the verdict `fail`; each says it stays `inconclusive` with `gate_unmet`, yields no PASS, and is exported as `INCONCLUSIVE`; each keeps the unreadable-configuration `fail` case; and a test pins `applyGateUnmet` returning `inconclusive` for a required unmet gate and `fail` for an unreadable gate. | REQ-AEC-014 | release-blocking | EV-10, EV-11 | M6 |
| AC-AEC-006 | Given the convention doc and its mirror, Then each carries the repair-scope paragraph with the three permitted repair kinds, the amendment route, and the clause exempting work specified by a SPEC amendment with its own AC; the mirror carries no SPEC ID or card id; and the two files are byte-identical. | REQ-AEC-015, REQ-AEC-016 | release-blocking | EV-12 | M6 |

REQ-AEC-016 is also checked at every milestone close: `make build` exits 0 and mirror pairs change in the same commit (the sync audit reads `git show --stat` per commit).

## §C Mutants each AC blocks

| AC | Mutant that would satisfy a shallow check while violating the requirement | Why it cannot pass here |
|---|---|---|
| 001a | Rewrite the scope note, keep the prose-only codex prompt | the test asserts the request payload's target at the codex writer seam |
| 001b | Default from any newest verdict, PASS included, or propose a delta on a first review | the no-verdict, PASS, and non-ancestor fixtures expect full scope |
| 001c | Check only the last link, or only that the proposed base is an ancestor of HEAD | the gap fixture has a valid last link and an ancestor base, yet a broken earlier link; it expects full scope |
| 001d | Accept any non-empty confirmation, or a subset of ids, or a closing commit outside the range | fixtures with one id missing, a count mismatch, and an out-of-range commit each expect full scope and no delta PASS |
| 001e | Build the chain from verdict-file `audited_sha` values or call arguments, or skip the receipt cross-check | the fixture's verdict files form a perfect chain while the records do not, and the forged-receipt fixture differs only in the receipt; both expect full scope |
| 001f | File a delta record with the full-scope base, or without the delta marker, or for a delta the gate rejected | the record test asserts the base equals the delta SHA, the marker, the linked receipt id, and that a rejected delta files a full-scope record |
| 002a | Set every class to `other`, or attach a class to pass/fail legs | three fixtures expect three distinct classes; pass/fail legs expect none |
| 002b | Admit `INCONCLUSIVE`, or drop the class from the reason, or skip FAIL files in the counter too | the admission test expects a refusal naming the class; the counter fixture expects 2 of 3 |
| 002c | Re-run all legs, or trust a caller-supplied verdict | per-leg call counters must read 0; a caller verdict is ignored |
| 003a/b | Drop the head comparison, or allow all of `.moai/reports/`, or treat the whole `progress.md` as exempt | the `.go`, ledger, prefix-trick, rename, non-descendant, and uncommitted-edit fixtures expect stale; the single-line `§E.2` fixture (every other section unchanged) expects stale, so a whole-file exemption fails 003b |
| 003a | Treat any change under `§E.3` as stale, or allow only the `§E.1`/`§E.4` lines | the `§E.3` full-rewrite fixture (with `§E.2` byte-unchanged) expects the receipt inherited, so a mutant that marks `§E.3` stale fails 003a |
| 004a-c | A verb that clears any mark, a wildcard, an MCP wrapper, or one that honours a hand-written approval line | the SPEC-B mark must be byte-identical; every refusal fixture leaves state unchanged; the caller sweep allows one caller |
| 005 | Delete the wrong sentence without stating the real behavior | the AC requires the inconclusive and `INCONCLUSIVE` statements and the unreadable-config case, plus the code test |
| 006 | Add the paragraph without the exemption clause | the AC requires the exemption clause |

## §D RED-now Evidence Ledger

All entries: tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d`, run by manager-spec in this plan phase, read-only, single invocation; the exit code is its own field.

```
EV-01 (AC-AEC-001a · release-blocking · RED, measured)
command : grep -n "prose-only scope" internal/cli/audit_delta_scope.go
stdout  : 106:	return deltaBaseRef + " " + sha + " (prose-only scope: codex adversarial mode cannot be bounded mechanically)"
exit    : 0
red why : the audit_multi codex leg's delta scope exists only as prose;
          performCodexAudit (mcp_convergence.go:631) always sends turn/start.
green   : M3 — exit 1 with empty stdout, and the request-payload test passes.

EV-02 (AC-AEC-001b · release-blocking · RED, measured)
command : grep -l "audited_sha" internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go internal/cli/audit_delta_scope.go internal/cli/mcp_codex_audit.go
stdout  : (empty)
exit    : 1
red why : no audit tool file reads a verdict's audited_sha, so no default can exist.
          (Factory rebind code elsewhere in internal/cli reads it for another purpose.)
green   : M3 — the default-proposal fixture tests pass.

EV-03 (AC-AEC-001c, 001e · release-blocking · RED, measured)
command : grep -n "return r.Target == TargetBaseBranch" internal/closure/readiness.go
stdout  : 152:		return r.Target == TargetBaseBranch &&
exit    : 0
red why : the only reader of second-review records filters on target, head and
          changed-file count (readiness.go:151-155) and has no chain notion; no
          chain evaluation exists anywhere, so a delta cannot be gated.
green   : M1 — the chain fixtures (gap, wrong last link, no full-scope start,
          forged receipt, verdict-file-only chain) behave as AC-AEC-001c/e state.

EV-04 (AC-AEC-001d · release-blocking for the pair · RED, measured)
command : grep -n "blocking_defect\|blocking_closed" internal/auditverdict/verdict.go internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go .moai/docs/audit-artifact-convention.md
stdout  : (empty)
exit    : 1
red why : no verdict file can name BLOCKING defect ids and no call can confirm
          their closure.
green   : M1/M4 — the confirmation fixtures pass; the convention defines the line.

EV-05 (AC-AEC-001f · release-blocking · RED, measured)
command : grep -n 'cfg.CardID != "" && cfg.DeltaBase == ""' internal/cli/mcp_convergence.go
stdout  : 1059:	if cfg.CardID != "" && cfg.DeltaBase == "" {
exit    : 0
red why : a delta call never files a second-review record.
green   : M1 — the gated delta record test passes; this condition no longer
          excludes an admitted delta.

EV-06 (AC-AEC-002a, 002c · release-blocking · RED, measured)
command : grep -n "error_class" internal/cli/mcp_convergence.go
stdout  : (empty)
exit    : 1
red why : PerBackendVerdict (mcp_convergence.go:92-117) carries no error class.
green   : M4 — the three-class fixture test and the carried-leg counter test pass.

EV-07 (AC-AEC-002b · release-blocking · RED, measured; two observations)
command : grep -n "verdict: <PASS|PASS-WITH-DEBT|FAIL>" .moai/docs/audit-artifact-convention.md
stdout  : 87:  verdict: <PASS|PASS-WITH-DEBT|FAIL>
exit    : 0
command : grep -n "numbers count even when the verdict label" internal/runtime/audit_counter.go
stdout  : 10:// numbers count even when the verdict label, score, and audited hash are
exit    : 0
red why : no non-FAIL label exists for an unrun audit, and the round counter
          counts every iteration file regardless of label.
green   : M4 — the label line lists INCONCLUSIVE; the admission test refuses with
          the class; the counter fixture counts 2 of FAIL/INCONCLUSIVE/FAIL.

EV-08 (AC-AEC-003a, 003b · release-blocking · RED, measured)
command : grep -n "r.HeadCommit != head" internal/auditreceipt/store.go
stdout  : 1608:				if r.HeadCommit != head || r.TreeDigest != digest {
exit    : 0
red why : any commit moves both head and digest (TreeState folds the head in,
          store.go:242), so a reports-only commit makes every receipt stale.
green   : M5 — the allowlisted-range test accepts (including the §E.3
          full-rewrite fixture); every negative fixture refuses (including the
          single changed §E.2 line).

EV-09 (AC-AEC-004a..c · release-blocking · RED, measured)
command : grep -rln "ClearRejection" internal/cli
stdout  : (empty)
exit    : 1
red why : no CLI code clears a rejection; the only clears are in the hook's
          auditor-stop path (audit_receipt_guard.go:185-190, 225-235).
green   : M2 — the positive test passes, every refusal fixture refuses, and the
          sweep finds exactly one non-test caller.

EV-10 (AC-AEC-005 · release-blocking · RED, measured)
command : grep -c "verdict: fail" .claude/agents/moai/sync-auditor.md .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md
stdout  : .claude/agents/moai/sync-auditor.md:1
          .claude/agents/moai/plan-auditor.md:1
          internal/template/templates/.claude/agents/moai/plan-auditor.md:1
          internal/template/templates/.claude/agents/moai/sync-auditor.md:1
exit    : 0
red why : all four files say an unmet required gate returns `verdict: fail`, while
          applyGateUnmet keeps it inconclusive (mcp_codex_audit.go:2448).
green   : M6 — every count is 0 (exit 1), the inconclusive wording is present, and
          the applyGateUnmet test passes.

EV-11 (AC-AEC-005 · regression-guard grade, abridged stdout)
command : grep -n "overall verdict is \`fail\`\|makes the overall verdict \`fail\`" .claude/agents/moai/sync-auditor.md .claude/agents/moai/plan-auditor.md
stdout  : .claude/agents/moai/sync-auditor.md:201:… A non-empty `gate_unmet` makes the overall verdict `fail`.
          .claude/agents/moai/plan-auditor.md:251:3. Read `convergence_check`. … the overall verdict is `fail`.
exit    : 0
red why : a second wrong sentence per file; enforceRequiredGateUnmet keeps the
          overall inconclusive (mcp_convergence.go:1144-1146).
          The stdout is abridged with "…" (each match is a full paragraph), so this
          entry is regression-guard grade; AC-AEC-005 stays release-blocking
          through EV-10, which is verbatim.
green   : M6 — exit 1 on both files and both mirrors.

EV-12 (AC-AEC-006 · release-blocking · RED, measured)
command : grep -n "may only remove" .moai/docs/audit-artifact-convention.md internal/template/templates/.moai/docs/audit-artifact-convention.md
stdout  : (empty)
exit    : 1
red why : the chosen home and its mirror carry no repair-scope rule (the same grep
          over spec-workflow.md and its mirror also exited 1).
          Companion: `cmp .moai/docs/audit-artifact-convention.md
          internal/template/templates/.moai/docs/audit-artifact-convention.md`
          exited 0.
green   : M6 — the paragraph is present in both and cmp still exits 0.
```

## §E Edge cases

- A card directory with verdict files from two phases: the newest by the existing rebind selection decides. A tie or conflicting `audited_sha` means full scope.
- A chain whose full-scope anchor record predates a later full-scope record: the newest full-scope record starts the chain, and older links are ignored.
- A worktree without its own config, whose store lives in the primary checkout: the B4 verb clears only marks whose `tree_root` equals the current tree, and chain receipts must name the same tree root.
- A legacy rejection record with no `tree_root` is treated as the store root's own tree (the existing rule).
- `UnknownSpec` marks are clearable only by naming `unknown` exactly; the evidence check then requires the agent type in the evidence file.

## §F Quality gates

- Scoped `go test` green for every touched package, with the named tests present (swept count above zero for each `-run` selector).
- `golangci-lint run` clean on touched packages.
- `make build` exits 0 after each mirror commit.
- Coverage of new functions at or above 85%.

## §G Definition of Done

- Every AC's green cell is observed and recorded verbatim in `progress.md` §E.2.
- Template mirror pairs land in the same commits.
- The sync audit passes, and its verdict file is under `.moai/reports/t2017/`.
