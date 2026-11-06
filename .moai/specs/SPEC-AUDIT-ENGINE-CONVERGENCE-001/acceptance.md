# Acceptance — SPEC-AUDIT-ENGINE-CONVERGENCE-001

## §A Adoption contract

Every AC below is adopted as a pair (`verification-completeness.md` §2):
- a RED-now cell measured on tree `cb30e05b3e5f64774b8dabd87e80c24332f4551d`. This is a document-level pin that binds every entry naming no tree of its own; entries measured in round 4 carry their own pin, `36148aeec99fd8d80e76645432f5a58802732c3e`, whose code equals `cb30e05b3e`;
- a green-path cell naming the milestone that flips it.

Each AC also names the mutant it blocks.

There are two classes:
- **release-blocking**: the RED was executed in this plan phase as a read-only single invocation, and it can be re-executed on the pinned tree.
- **paired guard**: a negative sub-criterion whose subject does not exist yet, so it passes vacuously today. It is adopted only together with the release-blocking sibling named in its row, and it is never recorded as a pass on its own.

Each RED below is a source observation of the defective path, not the AC's whole verdict. Every green cell pairs that observation's inverse with a named behavior test, so deleting the text without fixing the behavior does not pass.

**Multi-file `grep -c` output order.** The per-file line order of a multi-file `grep -c` may vary by environment. Each recorded order (EV-10, EV-11, EV-17) was observed on one machine, the lane's, and on that machine a plain multi-file `grep -c` over the local agent files printed its files in different orders on different runs (rounds 5 and 6). No cause for the ordering was measured, and none is claimed. Read a multi-file cell by its `file:count` pairs (in EV-20, `tree:path:count` triples), never by line position.

## §B AC Matrix

| AC | Given-When-Then | REQ | Class | Evidence | Milestone |
|---|---|---|---|---|---|
| AC-AEC-001a | Given a delta in effect on `audit_multi`, When the codex leg sends its request, Then the request is the adversarial `turn/start` turn, it contains the server-collected `delta_base...HEAD` diff bytes, it carries no native review target, and the scope note contains neither "prose-only" nor "prose only". | REQ-AEC-003 | release-blocking | EV-01 | M3 |
| AC-AEC-001b | Given a fixture card, When `audit_multi` is called with its `card_id` and no `delta_base`, Then: with no verdict file at all (first review), or a newest verdict that is PASS, or an `audited_sha` that is a non-ancestor or conflicting, the full scope is reviewed and the result says why; with a FAIL verdict carrying an ancestor `audited_sha`, that SHA is proposed and named with its file. An explicit SHA replaces the file's and, with a `card_id`, is reviewed at full scope when its chain is broken (the same fixture as 001c). `delta_base: "full"` reviews the resolved base. Without a `card_id`, an explicit SHA keeps the t1708 behavior. | REQ-AEC-001, REQ-AEC-002 | release-blocking | EV-02 | M3 |
| AC-AEC-001c | Given server-written second-review records (full-scope R0, delta R1 whose base is R0's head), each with a matching receipt and pass/fail from every required backend, and a complete closure confirmation, When the proposed base equals R1's head, Then the delta scope is reviewed. When any one of the following holds, Then the full scope is reviewed and the result names the failing link: a link's base differs from the previous record's head (a gap); the proposed base differs from the last record's head; no full-scope record starts the chain; R0 records a required backend `inconclusive`; R1 records a required backend `inconclusive`; with the required set {codex, glm}, R0 carries a codex pass and no glm entry at all; R1 likewise carries no glm entry; R0's or R1's receipt has a non-empty `gate_unmet` or an inconclusive codex verdict. | REQ-AEC-004 (i) | release-blocking | EV-03 | M1 |
| AC-AEC-001d | Given an unbroken chain, When any one of the following holds, Then the full scope is reviewed and no delta-scope PASS is produced: the confirmation is omitted; it names only some ids of the union; the closing commit of an id new since the previous link lies outside `delta_base..HEAD`; a FAIL file's `blocking_defect:` count differs from its `blocking_count`; a FAIL file carries no `blocking_count`; the card has no FAIL verdict file tied to the chain. Given two FAIL files on the chain (an older plan-audit FAIL with ids X, Y and a newer sync-audit FAIL with id Z only), Then a confirmation naming only Z is incomplete and one naming X, Y, Z is complete. Given an inherited closure (R0's FAIL names X, Y; commit C1 closes them; R1, a delta from R0, is admitted with a confirmation of X, Y closed by C1 and records X, Y as confirmed; R1's FAIL names only Z; Z's closing commit lies after R1's head), When R2 is proposed as a delta from R1's head with a confirmation covering X, Y, Z, Then the delta scope is reviewed although C1 lies outside R2's `delta_base..HEAD`; if R1 instead recorded only X as confirmed, the same R2 proposal is reviewed at full scope. Given a re-report (R0's FAIL names X; C1 claims to close X; R1, a delta from R0, is admitted with a confirmation of X and records X as confirmed, and R1's FAIL names X again), When R2 is proposed as a delta from R1's head with a confirmation covering X, Then with no commit after R1's head that closes X the full scope is reviewed, and with such a closing commit inside R2's `delta_base..HEAD` the delta scope is reviewed. When the confirmation is complete, Then every leg's request carries it. | REQ-AEC-004 (ii) | paired guard (with 001c) | EV-04 | M1 |
| AC-AEC-001e | Given a chain whose link record cites a receipt absent from the receipt store, or a receipt whose tool, tree root, or head commit disagrees with the record, or chain data supplied as a call argument or present only in verdict files, When `audit_multi` evaluates the chain, Then that chain is refused and the full scope is reviewed. | REQ-AEC-004 (i) | paired guard (with 001c) | EV-03 | M1 |
| AC-AEC-001f | Given a delta call admitted by the chain gate, When it completes, Then: exactly one second-review record is appended, with scope base equal to the delta base, the delta marker, the receipt id of the record it extends, and the SHA-256 of each relied-on FAIL verdict file; `SelectSecondReview` keeps it; and the session file holds its result even when its overall is pass. Given a delta call the gate sent to full scope, Then its record and persistence are those of a full-scope call. Given a delta PASS without a `card_id`, Then the session file is unchanged (t1873). A full-scope call's record is shaped as today. | REQ-AEC-005 | release-blocking | EV-05, EV-18 | M1 |
| AC-AEC-002a | Given leg stubs that fail with a quota message, a deadline, and an authentication refusal, When `audit_multi` converges, Then each inconclusive leg carries `error_class` `quota`, `timeout`, `auth`, its `verdict` stays `inconclusive`, the receipt carries the class, and a `pass` or `fail` leg carries no class. | REQ-AEC-006 | release-blocking | EV-06 | M4 |
| AC-AEC-002b | Given a verdict file labelled `INCONCLUSIVE` with `convergence_overall: inconclusive`, `required_backend: codex inconclusive quota`, and no `required_backend_fail:` line, When the admission rule reads it, Then it refuses with a reason naming `codex` and `quota`, while `required_backend: codex pass quota` and `required_backend: codex inconclusive lunch` are malformed and refuse. When an auditor's last line is `AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none`, Then the Stop step parses it (no `CauseVerdictLineMissing`), clears no rejection, and stage entry stays blocked with the class in the reason. When the factory card reader reads `verdict: INCONCLUSIVE`, Then the card stays in audit. Given iteration files FAIL (1), INCONCLUSIVE (2), FAIL (3), Then `CountAuditRounds` and `CountPlanAuditRounds` each count 2. Given FAIL (1), INCONCLUSIVE (2), Then the latest verdict selected is iteration 2. The convention's label line lists `INCONCLUSIVE`. | REQ-AEC-007 | release-blocking | EV-07, EV-13, EV-14, EV-15, EV-16 | M4 |
| AC-AEC-002c | Given a persisted result for session S carrying head H, digest D, and scope P, When `audit_multi` re-runs only the codex leg for S at H, D, P, Then the claude and glm leg call counters are 0 and their verdicts are carried. When the head, the digest, or the scope differs, or the persisted result lacks those fields, Then all legs run. When the caller supplies a verdict for a carried leg, Then it is ignored. | REQ-AEC-008 | paired guard (with 002a) | EV-06, EV-18 | M4 |
| AC-AEC-003a [RETIRED] | Retired with B3 by operator decision (2026-11-06): the freshness relaxation across evidence-only commits is excluded from this SPEC and returns as its own SPEC and card. | REQ-AEC-009 [RETIRED] | retired | EV-08 (retired) | M5 (retired) |
| AC-AEC-003b [RETIRED] | Retired with AC-AEC-003a [RETIRED], for the same reason. Receipt freshness stays as SPEC-MCP-NEUTRALITY-001 REQ-MN-020 sets it: a receipt whose head differs from the current HEAD is stale. | REQ-AEC-010 [RETIRED] | retired | EV-08 (retired) | M5 (retired) |
| AC-AEC-004a | Given a tree with rejection marks for SPEC-A and SPEC-B, a terminal on standard input (seam), neither `CLAUDECODE` nor `CLAUDE_CODE_SESSION_ID` set, and an evidence file under `.moai/reports/` naming SPEC-A, When the verb clears the SPEC-A mark with a correct typed confirmation, Then one approval record is appended with the operator name, git identity, mark fields, evidence path and SHA-256, HEAD and UTC time; the SPEC-A mark is gone; the SPEC-B mark is byte-identical; and the spawn check no longer lists SPEC-A. | REQ-AEC-013 | release-blocking | EV-09 | M2 |
| AC-AEC-004b | Given any one of the following, When the verb runs, Then it refuses, and the approvals log and every mark are unchanged: `CLAUDECODE` set; `CLAUDE_CODE_SESSION_ID` set; standard input not a terminal (including `</dev/null`); a mark name that does not exist; a mark recorded for another tree in the same store; an evidence file outside `.moai/reports/` or not naming the mark's SPEC ID; a typed confirmation naming another agent type or SPEC. The marker set the verb checks equals the contract signing path's set (one test compares the two). | REQ-AEC-011, REQ-AEC-012 | paired guard (with 004a) | EV-09, EV-19 | M2 |
| AC-AEC-004c | Given an approval line appended to the approvals log by hand, Then the spawn check still denies. Given the built server's `tools/list` and a source sweep of callers of the new store function, Then no MCP tool clears a rejection, and the function has exactly one non-test caller, the verb. | REQ-AEC-011 | paired guard (with 004a) | EV-09 | M2 |
| AC-AEC-005 | Given the two agent sources under `internal/template/agentemit/source/` and the four local definitions `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md`, `.codex/agents/moai/plan-auditor.toml`, and `.codex/agents/moai/sync-auditor.toml` (hand-updated, leader rulings R308, R313), Then each of the six files: <br>- contains none of the three wrong sentences (EV-10, EV-11 for the sources; EV-20 for the local definitions); <br>- states that an unmet gate with no reported defect stays `inconclusive` with `gate_unmet`, yields no PASS, and is exported as `INCONCLUSIVE`; <br>- states that a reported fail stays fail; <br>- keeps the unreadable-configuration `fail` case; <br>- lists `INCONCLUSIVE` in both machine-line templates; <br>- obliges `blocking_count` and `blocking_defect:` lines on every FAIL (EV-17, EV-20). <br>After `make agents-emit`, `make agents-emit-check` exits 0, so the template Claude files and the template Codex TOML files carry the source text; the emitter does not reach the four local definitions, so they are checked by the statements above. Tests pin: `applyGateUnmet` returns `inconclusive` for a required unmet gate and `fail` for an unreadable gate; `enforceRequiredGateUnmet` keeps `fail` for a convergence where one required backend reported fail and another required leg was inconclusive. | REQ-AEC-014 | release-blocking | EV-10, EV-11, EV-17, EV-20 | M6 |
| AC-AEC-006 | Given the convention doc and its mirror, Then each carries the repair-scope paragraph with the three permitted repair kinds, the amendment route, and the clause exempting work specified by a SPEC amendment with its own AC; the mirror carries no SPEC ID or card id; and the two files are byte-identical. | REQ-AEC-015, REQ-AEC-016 | release-blocking | EV-12 | M6 |

REQ-AEC-016 is also checked at every milestone close: `make build`, which runs `agents-emit-check`, exits 0, and mirror pairs and rendered copies change in the same commit (the sync audit reads `git show --stat` per commit).

## §C Mutants each AC blocks

| AC | Mutant that would satisfy a shallow check while violating the requirement | Why it cannot pass here |
|---|---|---|
| 001a | Rewrite the scope note but keep the prose-only codex prompt, or switch the leg to the native target | the test asserts, at the codex writer seam, the adversarial turn, the diff bytes in it, and no target field |
| 001b | Default from any newest verdict, PASS included; propose a delta on a first review; or let an explicit SHA with a `card_id` skip the gate | the no-verdict, PASS, and non-ancestor fixtures expect full scope, and the explicit-SHA fixture on a broken chain expects full scope |
| 001c | Check only the last link; or check only ancestry; or accept any record that exists | the gap fixture has a valid last link and an ancestor base; the two inconclusive-record fixtures and the `gate_unmet` receipt fixture are well-formed links that never ran a required backend; each expects full scope |
| 001c (absent entry) | Check only the backend entries a record contains, so a record with no glm entry passes because every entry present is a pass | the R0 and R1 fixtures with the required set {codex, glm} and no glm entry expect full scope; the check must iterate over the required set |
| 001d | Compare ids against the newest FAIL only; read a missing `blocking_count` as 0; accept a subset of ids; accept an out-of-range commit; or accept a delta with no FAIL verdict | the two-FAIL fixture expects Z-only to be incomplete; the count-absent fixture and the no-FAIL fixture expect full scope (an empty union with a zero count would otherwise pass); the subset and out-of-range fixtures expect full scope |
| 001d (inheritance) | Require every id's closing commit inside the current delta, so C1 must lie inside R2's `delta_base..HEAD`; or inherit an id the extending record never confirmed | the inherited-closure fixture expects R2 admitted with C1 outside its delta, so a current-delta-only rule fails it; the X-only variant expects full scope, so inheriting an id the extending record never confirmed fails it (REQ-AEC-005); and the out-of-range fixture, whose id is new since the previous link, still expects full scope |
| 001d (re-report) | Inherit by id regardless of a later re-naming, so R1's recorded confirmation of X carries X into R2 although R1's FAIL named X again | the re-report fixture without a closing commit after R1's head expects full scope; an id-level inheritance admits the delta and fails it, while the variant with that closing commit still expects delta scope, so a rule that never admits a re-named id fails too |
| 001e | Build the chain from verdict-file `audited_sha` values or call arguments, or skip the receipt cross-check | the fixture's verdict files form a perfect chain while the records do not, and the forged-receipt fixture differs only in the receipt; both expect full scope |
| 001f | File a delta record with the full-scope base, without the marker, or for a delta the gate rejected; persist every delta PASS; or persist none | the record test asserts the base, marker, linked receipt id, and verdict-file digests; the persistence fixtures expect the gated delta PASS persisted and the `card_id`-less delta PASS not persisted |
| 002a | Set every class to `other`, or attach a class to pass/fail legs | three fixtures expect three distinct classes; pass/fail legs expect none |
| 002b | Admit `INCONCLUSIVE`; drop the class from the reason; keep `convergence_overall: fail` (the reason then stops at "overall fail"); accept any third token; leave the Stop grammar or the card reader at three labels; change only one counter; skip FAIL files too; or skip INCONCLUSIVE in `Latest` | the admission fixture expects the class-naming reason; the two malformed-token fixtures expect refusal; the last-line fixture expects a parsed line and no `CauseVerdictLineMissing`; the card-reader fixture expects the card held; each counter is asserted separately at 2 of 3; the `Latest` fixture expects iteration 2 |
| 002c | Re-run all legs; trust a caller-supplied verdict; or carry from a result whose head, digest, or scope differs | per-leg call counters must read 0; a caller verdict is ignored; each of the three mismatch fixtures expects all legs to run |
| 003a, 003b (retired) | none — both criteria are retired with B3 (operator decision, 2026-11-06) | no mutant row is kept for a retired criterion |
| 004a-c | A verb that clears any mark, a wildcard, an MCP wrapper, one that honours a hand-written approval line, or one with a marker set different from the signing path | the SPEC-B mark must be byte-identical; every refusal fixture leaves state unchanged; the set-equality test fails a divergent set; the caller sweep allows one caller |
| 005 | Delete the wrong sentences without stating the real behavior; edit the template copies instead of the sources; fix the sources and rely on `make agents-emit` to reach the local definitions; or say "fail only when the configuration cannot be read" on the convergence path | the AC requires each positive statement in the sources and in all four local definitions; `agents-emit-check` fails a copy-only edit; the local definitions are checked directly, because the emitter's project targets exclude them (`sets.go:62-63`) and EV-20 measures them separately from the sources; the convergence test keeps a reported fail at `fail` |
| 005 (local Codex copies) | Hand-fix only the two `.claude/agents/moai/` copies and leave `.codex/agents/moai/{plan,sync}-auditor.toml` unchanged | AC-AEC-005 names all four local files; EV-20's green requires empty output for both wrong-sentence commands over all four, so either unchanged TOML file still prints `…toml:1` and the cell stays red |
| 006 | Add the paragraph without the exemption clause | the AC requires the exemption clause |

## §D RED-now Evidence Ledger

All entries were run by manager-spec in this plan phase as read-only, single invocations. The exit code is recorded as its own field. Unless an entry names its own tree, it was measured on `cb30e05b3e5f64774b8dabd87e80c24332f4551d`.

```
EV-01 (AC-AEC-001a · release-blocking · RED, measured)
command : grep -n "prose-only scope" internal/cli/audit_delta_scope.go
stdout  : 106:	return deltaBaseRef + " " + sha + " (prose-only scope: codex adversarial mode cannot be bounded mechanically)"
exit    : 0
red why : the audit_multi codex leg's delta scope exists only as prose.
green   : M3 — exit 1 with empty stdout, and the request-payload test passes.

EV-02 (AC-AEC-001b · release-blocking · RED, measured)
command : grep -l "audited_sha" internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go internal/cli/audit_delta_scope.go internal/cli/mcp_codex_audit.go
stdout  : (empty)
exit    : 1
red why : no audit tool file reads a verdict's audited_sha, so no default can exist.
green   : M3 — the default-proposal fixture tests pass.

EV-03 (AC-AEC-001c, 001e · release-blocking · RED, measured)
command : grep -n "return r.Target == TargetBaseBranch" internal/closure/readiness.go
stdout  : 152:		return r.Target == TargetBaseBranch &&
exit    : 0
red why : the only reader of second-review records filters per record and has
          no chain notion; no chain evaluation exists, so a delta cannot be gated.
green   : M1 — every chain fixture of AC-AEC-001c/e behaves as stated.

EV-04 (AC-AEC-001d · release-blocking for the pair · RED, measured)
command : grep -n "blocking_defect\|blocking_closed" internal/auditverdict/verdict.go internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go .moai/docs/audit-artifact-convention.md
stdout  : (empty)
exit    : 1
red why : no verdict file can name BLOCKING ids and no call can confirm closure.
green   : M1/M4 — the confirmation fixtures pass; the convention defines the line.

EV-05 (AC-AEC-001f · release-blocking · RED, measured)
command : grep -n 'cfg.CardID != "" && cfg.DeltaBase == ""' internal/cli/mcp_convergence.go
stdout  : 1059:	if cfg.CardID != "" && cfg.DeltaBase == "" {
exit    : 0
red why : a delta call never files a second-review record.
green   : M1 — the gated delta record test passes.

EV-06 (AC-AEC-002a, 002c · release-blocking · RED, measured)
command : grep -n "error_class" internal/cli/mcp_convergence.go
stdout  : (empty)
exit    : 1
red why : PerBackendVerdict carries no error class.
green   : M4 — the three-class fixture test and the carried-leg tests pass.

EV-07 (AC-AEC-002b · release-blocking · RED, measured; two observations)
command : grep -n "verdict: <PASS|PASS-WITH-DEBT|FAIL>" .moai/docs/audit-artifact-convention.md
stdout  : 87:  verdict: <PASS|PASS-WITH-DEBT|FAIL>
exit    : 0
command : grep -n "numbers count even when the verdict label" internal/runtime/audit_counter.go
stdout  : 10:// numbers count even when the verdict label, score, and audited hash are
exit    : 0
red why : no non-FAIL label exists for an unrun audit in the exported file, and
          the first round counter counts every iteration file whatever its label.
green   : M4 — the label line lists INCONCLUSIVE; the counter fixture counts 2 of 3.

EV-08 (AC-AEC-003a [RETIRED], 003b · retired)
status  : retired with B3 by operator decision (2026-11-06); no RED or green
          cell is kept. A receipt whose head differs from the current HEAD
          stays stale (SPEC-MCP-NEUTRALITY-001 REQ-MN-020, unchanged).

EV-09 (AC-AEC-004a..c · release-blocking · RED, measured)
command : grep -rln "ClearRejection" internal/cli
stdout  : (empty)
exit    : 1
red why : no CLI code clears a rejection; the only clears are in the hook's
          auditor-stop path.
green   : M2 — the positive test passes, every refusal fixture refuses, and the
          sweep finds exactly one non-test caller.

EV-10 (AC-AEC-005 · release-blocking · RED, re-measured in round 4)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -c 'returns `verdict: fail` with a non-empty' internal/template/agentemit/source/plan-auditor.md internal/template/agentemit/source/sync-auditor.md
stdout  : internal/template/agentemit/source/sync-auditor.md:1
          internal/template/agentemit/source/plan-auditor.md:1
exit    : 0
order   : observed on the lane's machine; re-run in round 5 on cc100c2909 with
          the same output. The per-file order may vary by environment (§A);
          read the pairs, not the line positions.
red why : both sources say an unmet required gate returns `verdict: fail` on
          the single-backend path, while applyGateUnmet keeps it inconclusive
          (mcp_codex_audit.go:2448). The pattern names that one sentence, so the
          kept "fail when the configuration cannot be read" sentence cannot
          turn a correct fix red.
green   : M6 — each count is 0 (exit 1), and the applyGateUnmet test passes.
          (Superseded entry: the round-1 cell grepped `verdict: fail` over the
          four rendered copies; its recorded line order could not be checked,
          and the target moved to the sources.)

EV-11 (AC-AEC-005 · release-blocking · RED, re-measured in round 4)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -c 'gate_unmet` makes the overall verdict `fail`\|the overall verdict is `fail`' internal/template/agentemit/source/plan-auditor.md internal/template/agentemit/source/sync-auditor.md
stdout  : internal/template/agentemit/source/sync-auditor.md:1
          internal/template/agentemit/source/plan-auditor.md:1
exit    : 0
order   : observed on the lane's machine; may vary by environment (§A).
red why : a second wrong sentence per source; enforceRequiredGateUnmet keeps the
          overall inconclusive unless a backend reported fail
          (mcp_convergence.go:1144-1146).
green   : M6 — each count is 0 (exit 1) in both sources; the rendered copies
          follow through agents-emit-check (EV-17's green).

EV-12 (AC-AEC-006 · release-blocking · RED, measured)
command : grep -n "may only remove" .moai/docs/audit-artifact-convention.md internal/template/templates/.moai/docs/audit-artifact-convention.md
stdout  : (empty)
exit    : 1
red why : the chosen home and its mirror carry no repair-scope rule.
green   : M6 — the paragraph is present in both and cmp still exits 0.

EV-13 (AC-AEC-002b · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n "AUDIT-VERDICT" internal/auditreceipt/store.go
stdout  : 173:	`^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL) spec=((?:` + specIDBody + `|` + cardIDBody + `)) receipts=(none|rcpt-[a-z0-9]{20,40}(?:,rcpt-[a-z0-9]{20,40})*)[ \t\r]*$`)
exit    : 0
red why : an INCONCLUSIVE last line fails this grammar and is recorded as
          CauseVerdictLineMissing.
green   : M4 — the last-line fixture parses and records no missing-line cause.

EV-14 (AC-AEC-002b · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n "verdictLinePattern" internal/homestate/card_evidence_readers.go
stdout  : 20:	verdictLinePattern  = regexp.MustCompile(`^verdict: (PASS|PASS-WITH-DEBT|FAIL)$`)
          128:		if m := verdictLinePattern.FindStringSubmatch(line); m != nil {
exit    : 0
red why : the factory card reader does not recognise `verdict: INCONCLUSIVE`.
green   : M4 — the card-reader fixture holds the card with the class in its reason.

EV-15 (AC-AEC-002b · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n "return len(files), nil" internal/runtime/audit_ceiling.go
stdout  : 1191:	return len(files), nil
exit    : 0
red why : the second round counter counts every iteration file.
green   : M4 — CountPlanAuditRounds counts 2 of FAIL/INCONCLUSIVE/FAIL.

EV-16 (AC-AEC-002b · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n 'v == "pass" || v == "fail"' internal/auditverdict/verdict.go
stdout  : 155:			if v := strings.ToLower(value); v == "pass" || v == "fail" {
          211:		return v, v == "pass" || v == "fail"
exit    : 0
red why : `convergence_overall: inconclusive` is a malformed receipt today.
green   : M4 — the admission fixture refuses with the class-naming reason.

EV-17 (AC-AEC-005 · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -c "blocking_defect" internal/template/agentemit/source/plan-auditor.md internal/template/agentemit/source/sync-auditor.md
stdout  : internal/template/agentemit/source/sync-auditor.md:0
          internal/template/agentemit/source/plan-auditor.md:0
exit    : 1
order   : observed on the lane's machine; may vary by environment (§A).
red why : neither auditor source obliges the BLOCKING-id lines the chain gate
          reads.
green   : M6 — each count is at least 1, and `make agents-emit-check` exits 0.

EV-18 (AC-AEC-001f, 002c · release-blocking · RED, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n 'cfg.DeltaBase == "" || result.OverallVerdict == overallVerdictFail' internal/cli/mcp_convergence.go
stdout  : 1046:	if cfg.SessionID != "" && (cfg.DeltaBase == "" || result.OverallVerdict == overallVerdictFail) {
exit    : 0
red why : a chain-admitted delta PASS is never persisted, and the persisted
          result carries no head, digest, or scope to re-run against.
green   : M1/M4 — the persistence fixtures and the carried-leg tests pass.

EV-19 (AC-AEC-004b · regression-guard grade, measured)
tree    : 36148aeec99fd8d80e76645432f5a58802732c3e
command : grep -n 'return \[\]string{config.EnvClaudeCode, config.EnvClaudeCodeSessionID}' internal/cli/contract.go
stdout  : 77:	return []string{config.EnvClaudeCode, config.EnvClaudeCodeSessionID}
exit    : 0
red why : this observes the set the verb must equal; it is green today and pins
          the reused set rather than a defect.
green   : M2 — the set-equality test passes.

EV-20 (AC-AEC-005 · release-blocking · RED, re-measured in round 6 over the four local definitions; three observations)
tree    : cc100c29093b02b7b117ec63a2eef7cc51de581f (the committed tree, read
          directly by `git grep <tree>`; every output line carries that SHA)
command : git grep -c 'returns `verdict: fail` with a non-empty' cc100c29093b02b7b117ec63a2eef7cc51de581f -- .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .codex/agents/moai/plan-auditor.toml .codex/agents/moai/sync-auditor.toml
stdout  : cc100c29093b02b7b117ec63a2eef7cc51de581f:.claude/agents/moai/plan-auditor.md:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.claude/agents/moai/sync-auditor.md:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.codex/agents/moai/plan-auditor.toml:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.codex/agents/moai/sync-auditor.toml:1
exit    : 0
command : git grep -c 'gate_unmet` makes the overall verdict `fail`\|the overall verdict is `fail`' cc100c29093b02b7b117ec63a2eef7cc51de581f -- .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .codex/agents/moai/plan-auditor.toml .codex/agents/moai/sync-auditor.toml
stdout  : cc100c29093b02b7b117ec63a2eef7cc51de581f:.claude/agents/moai/plan-auditor.md:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.claude/agents/moai/sync-auditor.md:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.codex/agents/moai/plan-auditor.toml:1
          cc100c29093b02b7b117ec63a2eef7cc51de581f:.codex/agents/moai/sync-auditor.toml:1
exit    : 0
command : git grep -c "blocking_defect" cc100c29093b02b7b117ec63a2eef7cc51de581f -- .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md .codex/agents/moai/plan-auditor.toml .codex/agents/moai/sync-auditor.toml
stdout  : (empty)
exit    : 1
reading : read each line as a `<tree>:<path>:<count>` triple, not by position;
          no line order is claimed. `git grep -c` prints no line for a file
          with zero matches, so the empty stdout with exit 1 means all four
          counts are 0.
red why : the four local definitions are not emitter targets
          (internal/template/agentemit/sets.go:62-63); each carries the two
          wrong lines and no BLOCKING-id obligation, so a source-only fix with
          `make agents-emit` leaves all four wrong, and a `.claude`-only hand
          fix leaves the two TOML copies wrong.
green   : M6, measured on the M6 commit <M6>:
          command : the first command with <M6> in place of the tree SHA
          stdout  : (empty)
          exit    : 1
          command : the second command with <M6> in place of the tree SHA
          stdout  : (empty)
          exit    : 1
          command : the third command with <M6> in place of the tree SHA
          stdout  : four lines, one per local definition, each
                    `<M6>:<path>:<N>` with N at least 1
          exit    : 0
          tree    : <M6>, the commit that lands the M6 hand edits
          and each of the four files states the positive statements of
          AC-AEC-005.
```

## §E Edge cases

- A card directory with verdict files from two phases: the newest by the existing rebind selection decides the proposal; the closure set spans every FAIL file tied to the chain.
- A chain whose full-scope starting record predates a later full-scope record: the newest full-scope record starts the chain; older links are ignored.
- A worktree without its own config, whose store lives in the primary checkout: the B4 verb clears only marks whose `tree_root` equals the current tree, and chain receipts must name the same tree root.
- A legacy rejection record with no `tree_root` is treated as the store root's own tree (the existing rule).
- `UnknownSpec` marks are clearable only by naming `unknown` exactly; the evidence check then requires the agent type in the evidence file.

## §F Quality gates

- Scoped `go test` green for every touched package, with the named tests present (swept count above zero for each `-run` selector).
- `golangci-lint run` clean on touched packages.
- `make build` (including `agents-emit-check`) exits 0 after each mirror or source commit.
- Coverage of new functions at or above 85%.

## §G Definition of Done

- Every AC's green cell is observed and recorded verbatim in `progress.md` §E.2.
- Template mirror pairs and rendered agent copies land in the same commits as their sources.
- The sync audit passes, and its verdict file is under `.moai/reports/t2017/`.
