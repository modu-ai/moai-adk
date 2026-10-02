# Decision Index — SPEC-MOAI-STATUS-MOD-001

Decisions surfaced while assembling this SPEC that the operator has not settled. Each row states what is unresolved and why; none carries a preferred answer (decision gate on, recommendation mode `pull`: `.moai/config/sections/interview.yaml`). Labels: `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. No row qualifies as `DECIDED` or `POLICY-COVERED`: none has an authority anchor verifiable in the committed tree that settles it, so none is relabeled to look settled. Measurements cited are `M-n` / `G-n` of `spec.md`, taken at tree `58dad3055`.

### Q1: What concrete signal does the design phrase "retention 상태" (retention state) name?

Label: FOUNDER
Authority anchor: —
Why unresolved: the design report detailing the retention signal was not readable during authoring (claude.ai login unavailable), so the phrase was derived from the repository. The derivation this SPEC carries (REQ-MSM-008, M-8) is the memory-store over-cap signal from `moai memory doctor --json`: per-store `topic_files` vs `cap` (primary-checkout store measured 1422 files against a cap of 50) plus the `MEMORY_TOPIC_COUNT_OVER_CAP` / `MEMORY_INDEX_OVERFLOW` finding codes. Two other readings exist in the tree and are NOT taken: `.moai/state` trace/runs retention (`DefaultTraceRetentionDays = 30` and kin, `internal/config/defaults.go:353-410`) and hook-runtime log retention (`DefaultHookRuntimeLogRetentionDays = 30`) — both are age-based clean policies, not live states a session can read without the clean command. Whether the intended signal was the memory store, the state files, or something the unreadable report defined otherwise is unconfirmed. Residue: which of the up-to-four resolved memory stores (G-7) the warning should attribute to the current session.
Operator verdict:

### Q2: Which usage levels does the AbovePrompt strip display, given the mirrored gate formulas?

Label: FOUNDER
Authority anchor: —
Why unresolved: REQ-MSM-003 pins the classification to the in-repo gates' default formulas (D-3, M-7; the gates themselves are env/config-overridable at runtime — the mod shows its own frozen values, divergence recorded as G-12): context warn at soft (50 for a window of 500,000+ tokens, else 90), critical at the hard ceiling (min(95, 85+10) clamped up to soft); rate windows warn at the t1347 gate holds (90 / 95). Whether the band should also show an INFO state below those levels (for example the raw percentages always, dimmed), or draw only at warn and above (the MVP's choice), and whether critical gets a visually distinct form, is a display decision no committed artifact states.
Operator verdict:

### Q3: What health poll interval, at or above the 15,000 ms floor of REQ-MSM-007, does the mod use?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: one cycle spawns three `moai` processes; the doctor single checks measured 1,578 bytes of output in well under the 20-second timeout (M-9) and `memory doctor --json` measured 186,726 bytes (M-8), but the wall time of a full cycle on a loaded machine is unmeasured (G-6), and the hook budget (10 s of own time, `$` calls in flight excluded) has not been exercised against three sequential runs. Each tick also runs the MCP check, which **deletes the CLI's own dead PID-stamp files as a side effect** (plan §B.3, spec §2) — a higher interval runs that delete less often. The floor is the sibling's poll floor; the value above it depends on how fresh the health line must be and on the measured cycle cost.
Operator verdict:

### Q4: Does the health line also show an OK state, or only warnings?

Label: FOUNDER
Authority anchor: —
Why unresolved: REQ-MSM-008 pins warnings-only for the MVP (`$.ui.status(undefined)` when nothing fails), on the reading that a permanently pinned "all healthy" line is noise beside the engine's own pinned notices. The opposite reading — a visible OK confirms the checks are running — is untested with a person (G-1). No committed artifact states which the operator wants.
Operator verdict:

### Q5: How does the spinner suffix compose when another writer already carries a suffix?

Label: FOUNDER
Authority anchor: —
Why unresolved: the Spinner event's `suffix` arrives with the engine's default (one ellipsis) or what an earlier hook in the chain rewrote it to (M-5: "A rewrite is drawn as given"). The MVP appends its marker to the incoming suffix verbatim and passes it on; it never removes or normalizes what it finds. Whether the operator wants the mod to skip appending when a non-engine suffix is already present (avoiding a two-marker line), or to compose a canonical order, is unspecified.
Operator verdict:

### Q6: Should the lane toast fire for every inbound delivery, or narrow to some origin kinds?

Label: FOUNDER
Authority anchor: —
Why unresolved: the design named "lane notification toast", and `session.receive` carries deliveries of several origins (`bridge`, `peer`, `coordinator`, `task-notification`, `scheduled-trigger`, `projects-relay`, `slack-ping`, `unclassified`; M-5). The MVP toasts all of them (REQ-MSM-006); a lane-heavy day could make that noisy, and a narrow filter (peer and coordinator only) would miss the relays. Which set the operator wants is a product choice; the delivery volume per origin was not measured (G-3).
Operator verdict:

### Q7: Where does the mod finally load from, and must it load under the operator's own profile?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: shared with the sibling card's Q7 (SPEC-MOAI-BOARD-MOD-001). Card t1434 (plugin load-scope measurement) is not done, so which scopes load a mod, and with what trust prompts, is unmeasured; the operator profile's rollout switch refused the sibling's engine runner and gates hooks-module loads (their M-13, G-11 — not re-read in this session). This SPEC fixes only the prototype location (`mods/moai-status/`, loaded with `claude --plugin-dir`) and says it moves with the t1434 verdict.
Operator verdict:
