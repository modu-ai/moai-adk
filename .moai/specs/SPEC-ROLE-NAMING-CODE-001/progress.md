# progress — SPEC-ROLE-NAMING-CODE-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1256 · branch `WT-role-naming-code` · plan base `e62c3e183`
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L) + this file
- Evidence: `.moai/reports/t1256/census.md` (reproducible via `census.py`), `.moai/reports/t1256/conflicts.md`
- v0.1.0 (`6fe67c674`): REQ 20 · AC 20.
- v0.2.0 revision (2026-09-26): operator answers relayed by the leader applied — `lane` canonical, legacy spellings rejected (no aliases), persisted values write-new with the run-boundary rule, role-marker guard `lane`-only, lane self-dispatch help text, homonym qualifiers, `manager-lead` kept, t1193 demoted to a recorded dependency, new ordering t1242 → t1245 → t1256 run → t1240 → t1257 (t1193 excluded pending decision), design §3 zh note scoped to the code layer.
  - Rewritten REQ: 001, 003, 004, 005, 007, 009, 010, 011, 012, 013, 014, 016, 017, 018, 019, 020. New REQ: 021, 022, 023. Unchanged: 002, 006, 008, 015.
  - Rewritten AC: 002, 003, 004, 005, 007, 008, 009, 011, 013, 014, 015, 016, 018, 019, 020. New AC: 021, 022, 023. Unchanged: 001, 006, 010, 012, 017.
  - REQ count 23 · AC count 23 (AC-RNC-001..023, contiguous); every REQ maps to ≥1 AC (acceptance.md §C).
- v0.2.0 items to confirm at Kickoff: O0 derived `lead` rejection, O1 env names kept, O2 run-boundary persisted-data rule; O7 open. **Superseded by v0.3.0 below.**
- Plan-audit iteration 1: FAIL 0.77 (`.moai/reports/t1256/plan-audit-iter1.md`, audited at `d0770b9cc`) — MP-7 (O7 open) + blocking D2–D7, D9.
- v0.3.0 revision (2026-09-26):
  - Operator answers given directly in the lane window recorded in plan.md §B as RESOLVED: O7 rename role-sense Go identifiers; O0 `lead`/`lead-<suffix>` rejection confirmed; O1 env var names kept; O2 live legacy records refused with retire-and-relaunch guidance. No clarification marker remains in plan.md or research.md.
  - Rewritten REQ: 001 (D6), 003/005 (case variants, D16), 004 (`<n>` ≥ 1), 009 (detection carve-out), 011 (D18), 012 (D4 — equality assertion is the one coupling), 018 (D2/D5/D15), 019 (D1/D3), 022 (D7 — factory-run scope and membership basis). New REQ: 024 (D9 — legacy-only run retire), 025 (D7/D8 — kanban registry, legacy board role declaration, SessionStart session-record writer). REQ-number placement note added (D12).
  - Rewritten AC: 002, 003, 004, 005, 006, 008, 011, 012, 013 (D13 fixed string), 014, 016 (D5), 018, 019, 022, 023 (D14 fixed population command). New AC: 024, 025. Word-boundary rule stated once at the top of acceptance.md.
  - plan.md: §B all resolved; §C.3 sibling-SPEC re-check with plan-time branch @ SHA (D10); §E remeasure scope adds `./internal/spec` (plus `./internal/homestate`, `./internal/web`); M1 names the SessionStart session-record writer and the retire fallback (D8, D9); M6 records the partial supersession of SPEC-FACTORY-WORKER-NAMING-001 at sync (D11); R9 added.
  - design.md D2/D4/D7/D8 and §2/§4 updated; research.md §3 splits board role declarations from session records (D8).
  - Evidence hygiene: stray nested copy `.moai/reports/t1256/.moai/` removed (D17). No auditor helper scripts (`recount.py`, `sample.py`) found untracked in the worktree.
  - REQ count 25 · AC count 25 (AC-RNC-001..025, contiguous) — both at the Tier L ceiling of 25, neither over; every REQ maps to ≥1 AC (acceptance.md §C).
- Plan-audit iteration 2: FAIL 0.845 (`.moai/reports/t1256/plan-audit-iter2.md`, audited at `5102a69e9`) — blocking N1 (AC-RNC-008 required a role-declaration write no production path performs) + optional N2–N8.
- v0.3.1 revision (2026-09-26), no REQ or AC added (both stay at 25):
  - N1 (blocking): measured at `5102a69e9` over production `.go` files — `DeclareRole` 0 callers; `ResolveDeclaredRole` called only at `internal/kanban/board_store.go:191`; `WriteBoardState` only at `board_store.go:363` (inside `TransitionIntoRunOpts`), which is called only at `board_store.go:344` (inside `TransitionIntoRun`), which has 0 callers; `RecoverBoard` 0. AC-RNC-008 drops the declaration from the launch-written records and verifies the read side with a test-written `DeclareRole` declaration; REQ-RNC-025 / AC-RNC-025 board clause names the legacy role `lead` instead of a relaunch remedy and states it governs guard behavior only; plan R1 rewritten with the measured call chain; design.md §2/§4 wording aligned. No launcher declaration write added.
  - N2: AC-RNC-008 lists each record with its expected value in a table.
  - N3: `lead` matched case-insensitively (`(?i)\blead\b`) in REQ-RNC-018 and the acceptance word-boundary rule (current-tree measurement: case-insensitive and case-sensitive string-literal counts both 29, so no existing `Lead`/`LEAD` hit). The `MOAI_*` exclusion cannot produce a match under word-boundary semantics (`_` is a word character — `perl -ne 'print if /\blead\b/i'` prints nothing for `MOAI_KANBAN_LEAD_ADDR`), so AC-RNC-018's control is now paired: token-only string passes, the same token plus a free-standing `lead` fails (a real hit); a `Lead` sentence-start mutation added.
  - N4: AC-RNC-023 population command strips Go comments (`sed -E 's#[[:space:]]+//.*$##'` then re-grep); reproduced at `5102a69e9` → 2 rows (`internal/cli/factory.go:544`, `internal/cli/kanban.go:711`), the `internal/statusline/types.go:250` comment row gone.
  - N5: REQ-RNC-022 states an empty-`run_id` legacy claim belongs to no run, is not refused, is not recognized as a lane, and is stale once dead.
  - N6: AC-RNC-019 adds the M5 census re-run recording remaining identifier-internal role-sense rows, each tagged with its exclusion, untagged count 0.
  - N7: allowlist entries bound to file + exact literal; line numbers are recorded only (REQ-RNC-018, AC-RNC-018).
  - N8: AC-RNC-025 SessionStart case split into (a) label trigger with no record and (b) record trigger with a `leader` label.
- Plan-audit iteration 3 (final, Tier L ceiling): FAIL 0.89 (`.moai/reports/t1256/plan-audit-iter3.md`, audited at `7c2b4d528`) — mandatory gates 7/7, blocking P1 only (plan.md M5 and R5 still described the allowlist as file:line-bound), optional P2–P4.
- Post-audit fixes (2026-09-26), no REQ or AC added (both stay at 25):
  - P1: plan.md M5 and R5 now state allowlist entries are bound to file + exact literal, the line number recorded only — consistent with REQ-RNC-018 and AC-RNC-018.
  - P2: census script committed at `.moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py` with an `--out` argument (default output under the ignored `.moai/reports/t1256/raw/`); output byte-identical to the report copy on the current tree (16608 rows). AC-RNC-019, plan.md §C pre-flight step 4, and research.md §1 reference the tracked path.
- Accepted debt (operator override, not fixed at plan close):
  - P3: REQ-RNC-022's empty-`run_id` legacy-claim clause has no dedicated AC-RNC-022 case. Run phase should add the Given/When/Then (live legacy claim with `run_id=''`, owner `worker-5` → factory leader launch proceeds, claim not counted as a lane) when implementing AC-RNC-022.
  - P4: AC-RNC-025(a) requires "no session record file is created", which REQ-RNC-025 implies ("instead of re-deriving") but does not state; wording alignment deferred.
- Plan-audit final state: iteration 3 FAIL 0.89 → operator override to **PASS-WITH-DEBT** after the P1 fix, decided by the operator in the lane window on 2026-09-26; no re-audit.
- Implementation Kickoff: approved by the operator on 2026-09-26, progression mode semi-autonomous.
- Status: plan closed; frontmatter `status: draft` left for manager-develop's `draft → in-progress` at run entry.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
