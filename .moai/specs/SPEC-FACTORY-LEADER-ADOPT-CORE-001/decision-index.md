# SPEC-FACTORY-LEADER-ADOPT-CORE-001 — Decision Index

Interview decision gate: `on` (`.moai/config/sections/interview.yaml`). One row per decision surfaced during assembly that the operator did not settle in an interview (none was held; this SPEC was authored under the lane's card t1863 dispatch). This artifact is stateless on the status axis — the SPEC's lifecycle lives in `spec.md` alone.

Routing applied per the published register: the board records behind Q1 and Q3 exist and were read (`moai decision read`, installed build `ee981a932`), but the verb exposes no digest prefix, so no `board:<id>#<digest>` line can be pinned — an unpinnable citation routes to FOUNDER, never to DECIDED. No prior completed SPEC's HISTORY or Amendments decides any of these questions under identical conditions; no operator setting or constitution clause covers them (SPEC-FACTORY-RUN-RETIRE-001, completed in this tree, names the adopt command as a retire-guard exit but decides nothing about the leader start).

### Q1: When a leader start without a resume option meets an orphan run, does the launcher refuse (no run created), or print the guidance block and still mint?

Label: FOUNDER
Class: product-level
Authority anchor: (none pinnable — board `d-20261017T140304Z-62a6` read via `moai decision read --scope card:t1849`, decided_by=leader under operator delegation; no digest prefix exposed by the verb, so the `board:<id>#<digest>` pin the register requires cannot be formed)
Why unresolved: The refuse-with-guidance behavior was ruled by the leader under delegation on 2026-10-17 and carried into card t1863's shrink instruction; the ruling is not an operator verdict and cannot be digest-pinned from the board, and no committed artifact decides it.
Operator verdict:

Options as authored vs. ranked: the SPEC as authored refuses (REQ-FLAC-002/005). Default: print the guidance block and mint (rule: preserves current behavior — the launch outcome today is a minted run and a started session). Alternate: refuse with guidance, no run created.

### Q2: Does the unified adoption write on the explicit `--factory-run` leader path extend to runs whose owner is live or indeterminate, or only to dead-owner orphans?

Label: FOUNDER
Class: product-level
Authority anchor: (none — no prior SPEC or operator setting decides the write path of the explicit flag; the predecessor SPEC-FACTORY-LEADER-ADOPT-001 explicitly kept this path unchanged, but it is held at plan and decides nothing)
Why unresolved: The card mandates the adoption write's guarantees for the adopt use (a dead-owner orphan named by the guidance); whether every explicit `--factory-run` leader start takes over through that write — including live-owner and indeterminate takeovers, gaining the address republish and the `run.resumed` event but no longer overwriting `lane_capacity` — changes the observable outcomes of an existing flag beyond the mandated case.
Operator verdict:

Options: (a) unified — every `--factory-run` leader start takes over through the adoption write (as authored, REQ-FLAC-003; one write path; fixes the stale lead address everywhere on the explicit path); (b) split — only runs whose owner classifies dead on the stamp adopt through the adoption write, live/indeterminate takeovers keep today's unconditional restamp. Default: (b) (rule: preserves current behavior of the live-owner takeover). Alternate: (a).

### Q3: Until the deferred chooser card lands, is the refusal terminal-independent (applies to interactive starts too), or should minimal stream-based detection ship now and refuse only off-terminal?

Label: FOUNDER
Class: product-level
Authority anchor: (none pinnable — the shrink ruling that defers interactivity detection is a leader board record of 2026-10-17 read through the same digest-less verb; the card instruction defers the detection but is not a committed artifact)
Why unresolved: The card's binding scope defers interactivity detection to a follow-up card, which as authored makes an interactive `moai cc -f` over an orphan refused the same way — a change to a shipped command's default behavior on a terminal that the operator has not directly confirmed.
Operator verdict:

Options: (a) terminal-independent refusal until the chooser card lands (as authored, spec.md Terms and DP-3; no stream is read, no prompt exists); (b) ship minimal detection now (stdin and stderr both terminals → the refusal is withheld) and refuse only off-terminal. Default: (b) (rule: smaller user-visible surface — it refuses in strictly fewer situations). Alternate: (a).
