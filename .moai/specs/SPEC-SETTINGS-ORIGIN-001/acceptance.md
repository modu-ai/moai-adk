# acceptance.md — SPEC-SETTINGS-ORIGIN-001

Verification layer for the t487 investigation. Format: `AC-XXX` Given-When-Then, binary-testable. The GEARS obligation lives in spec.md §2 (REQ-001..REQ-008); nothing here restates a requirement in Given-When-Then form as though it were a GEARS clause.

## §D AC Matrix

- **AC-001 — Q1 inventory completeness.** Given the repository at the t487 base commit, When the repo-wide writer search runs (write-target grep covering `internal/`, `pkg/`, `cmd/`, `.claude/hooks/`, template sources, Makefile, `scripts/`, `.claude/workflows/`, and `.github/workflows/`), Then every code path that writes `.claude/settings.json` appears in the inventory table with file:line, and a cross-check search for write-target patterns — bounded only by the repository root, never by the enumerated list — surfaces no code path absent from the table (any surfaced miss is a FAIL, recorded verbatim).
- **AC-002 — Q1 shape reconciliation.** Given the completed inventory and the §1 fingerprint (key order / ask list / matcher form / status-transition count), When each inventoried path is evaluated, Then each entry carries an explicit could-produce / could-not-produce judgment with per-path reasoning tied to the fingerprint markers (no bare "ruled out" without a stated marker).
- **AC-003 — SWEEP distribution.** Given `git worktree list` output of N trees (the primary checkout is entry 1 of the listing), When the read-only dirty check (`git --no-optional-locks -C <wt> status --porcelain -- .claude/settings.json` — mandatory form, plain `git status` takes index write locks in trees other lanes may be using) runs on every tree, Then the verdict records N and the count checked (equal), and every tree whose `.claude/settings.json` is dirty vs its HEAD carries path + md5 + the three shape markers (ask-list length, matcher form, key order); no tree is modified, disposed, or pushed.
- **AC-004 — Q2 conditional external inflow.** Given Q1 concludes no repository code path produces the dirty shape, When external candidates are enumerated, Then each candidate (other checkouts, old binaries, TS predecessor templates, Claude Code runtime writes, manual hand-edit) carries evidence or an explicit not-fetchable/not-measurable Gap label; if Q1 DID find a producer, this AC is satisfied by recording that branch decision with its evidence.
- **AC-005 — Q3 single recommendation.** Given the Q1/Q2/SWEEP findings, When the verdict is composed, Then it contains exactly one recurrence-prevention recommendation, grounded in a stated evidence premise, and compliant with t485 C4 (process-level if it proposes live observation — no config-file-digging axis).
- **AC-006 — verdict artifact.** Given run-phase completion, When the card closes, Then `.moai/reports/t487/verdict.md` exists with all five sections (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), each Evidence entry carrying command + verbatim observed output.
- **AC-007 — evidence integrity.** Given any quantitative or negative claim anywhere in the investigation (counts, "no dirty trees", "absent key"), When the claim is asserted, Then the command that produced it and its observed output are cited in the same entry; empty output is labeled, never reported as zero; absence of signal is never reported as evidence in either direction.
- **AC-008 — baseline reuse (REQ-008).** Given the completed investigation, When the verdict is reviewed, Then t480's six origin-elimination measurements appear only as cited baseline — no re-execution of them appears in the verdict's Evidence section — and the §1 lane-8 fingerprint observations appear only as hypotheses in the H1–H4 tree, never as established conclusions.

## Edge cases

- A swept worktree is locked, mid-merge, or unreadable → record the tree + the error verbatim as a Gap; do not skip silently (a skipped tree is a false "all clean").
- The preserved copy's md5 drifts between pre-flight and verdict → halt and report; the SPEC's baseline is broken.
- A dirty settings.json found in the sweep contains secrets/tokens → record path+md5+markers only; never paste file contents into the verdict.
- The npm TS-predecessor templates are unreachable → AC-004 records the fetch attempt (command + error) as a Gap, not as elimination of the candidate.
- Sweep finds the SAME dirty shape in ≥1 other tree → H4 systematic branch fires; Q2's "external" framing narrows and the recommendation must address a still-active writer (this outcome does not fail any AC; it re-routes M4).

## Quality gate / Definition of Done

- Tier S plan-auditor threshold 0.75; harness minimal; no code, so no test/coverage gate applies — the verdict artifact IS the deliverable.
- All 8 AC binary-passable by inspection of `.moai/reports/t487/verdict.md` + the SPEC artifacts; no AC requires trusting an uncited summary. Every REQ-001..REQ-008 is covered by ≥1 AC (REQ-008 → AC-008).
- Card close report names: card id, branch + HEAD, evidence path (`.moai/reports/t487/verdict.md`), and the measurement scope. Sync phase expected N/A (evidence-only, t480 precedent).
