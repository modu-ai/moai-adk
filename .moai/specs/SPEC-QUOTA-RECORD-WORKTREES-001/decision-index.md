# decision-index.md — SPEC-QUOTA-RECORD-WORKTREES-001 (card t1442)

Decisions surfaced while planning that the plan does not settle. Each row states what is unresolved and why and carries no recommendation and no preferred answer; the verdict line was empty at authoring and now records the resolution (decision oracle, Jev, spec 0.2.0; Q2 and Q5 were below the confidence gate and carry the leader's verdict of spec 0.3.0, accepted provisionally). Labels follow the routing in the manager-spec definition: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. No row has a committed-tree authority that decides it, so none is DECIDED or POLICY-COVERED. The spec.md decision sections carry each decision's options and the default the run phase may start from; the card text itself ("linked worktrees (or the session's own project dir)") leaves Q1 open.

### Q1: Which record directories should the usage aggregator read besides the primary checkout's? (D1)

Label: FOUNDER
Authority anchor: none
Why unresolved: the card names two forms — every linked worktree, or the session's own project directory — and a third (a glob of the two known worktree layouts) is possible; they differ in how many of the scattered readings the gate can see, in whether a worktree placed outside the repository root is covered, and in per-call cost. Which trade-off the operator wants is a product call, not a measurement.
Operator verdict: Resolved by decision oracle — option A, the primary directory plus every linked worktree enumerated from `<git-common-dir>/worktrees/*/gitdir` (confidence 1.00).

### Q2: What is the per-call bound on directories scanned, and in what order are they taken? (D4)

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: no timing of the evaluation exists on this machine, so a bound of 64, 128, or none, and an order by name or by newest metadata-directory modification, are unmeasured; the cost model (about 1667 file stats per call against 1043 today) is a count, not a latency.
Operator verdict: LEADER VERDICT — ACCEPTED provisionally. The decision oracle leaned toward 64 directories ordered by newest metadata modification (probabilities 0.48 / 0.38 / 0.14) at confidence 0.22, LOW, under the 0.5 gate; the leader accepted the plan's default: 128 directories, name order, `gitdir` read up to 4 KiB, all unmeasured, and REQUIRED the directory bound to be a configuration key. Applied as `workflow.quota_gate.max_scan_dirs` (default 128, range 1-1024, out of range yields the default; REQ-QWR-011, AC-QWR-013), overriding the earlier "no new config keys" default for this key alone.

### Q3: Should a status or steering output name where a reading came from (the directory, or a count of directories read)? (D5)

Label: FOUNDER
Authority anchor: none
Why unresolved: the predecessor's output is pinned byte for byte when the gate is off or pressure is off; adding a cell or key would change its goldens and tests for the benefit of operator visibility. Whether that visibility is worth editing the predecessor's pinned output is a product call.
Operator verdict: Resolved by decision oracle — option A, no output change; the predecessor goldens and tests stay untouched (confidence 0.97).

### Q4: Which command does the real-lane measurement run around — the read-only `moai factory status`, or `moai factory next` itself? (D6)

Label: FOUNDER
Authority anchor: none
Why unresolved: `moai factory next` leases a card when it succeeds, so measuring it adds a side effect on the live queue; the read-only variant reaches the same evaluation function but never enters the lane-specific wait loop or latch. The card leaves the choice to the plan.
Operator verdict: Resolved by decision oracle — `moai factory status` always (read-only, before and after); `moai factory next` additionally only when a lease is genuinely due anyway (confidence 0.83).

### Q5: Should the aggregator's seam keep its present type with a new production value, or change type to carry directories? (D2)

Label: FOUNDER
Authority anchor: none
Why unresolved: keeping the type leaves every predecessor test unedited at the cost of deriving the repository root from the state-directory path shape; changing the type is a cleaner interface but edits two helper functions of the predecessor's completed test file.
Operator verdict: LEADER VERDICT — ACCEPTED provisionally. The decision oracle chose to keep the seam type and derive the root from the path shape (probability 0.71) at confidence 0.42, under the 0.5 gate; the leader accepted it. If deriving the root from the path shape proves unsafe in the run phase, switch to a type change of the seam and report to the leader (plan.md §B).
