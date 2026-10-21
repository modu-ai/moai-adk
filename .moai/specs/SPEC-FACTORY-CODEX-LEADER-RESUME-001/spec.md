---
id: SPEC-FACTORY-CODEX-LEADER-RESUME-001
title: "Codex factory leader resume re-adoption (blocked: the Codex leader entry is refused)"
version: "0.1.0"
status: rejected
created: 2026-10-15
updated: 2026-10-15
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "factory, codex, leader, resume, run-adoption"
tier: M
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-LAUNCHER-ENTRY-FLAGS-001]
---

# SPEC-FACTORY-CODEX-LEADER-RESUME-001 — Codex factory leader resume re-adoption

## HISTORY

- 2026-10-15 · v0.1.0 · manager-spec · Plan-phase authoring from card t1745 (Tier M). The card asks that the Codex factory leader entry re-adopt the prior run on a resume-shaped start, as the cc and glm leader entries do. The premise was re-verified on tree `1ced8cab9` (branch `WT-codex-leader-resume`) and does not hold as stated: the Codex leader is refused by a completed SPEC, and the branch that would carry the adoption has no reachable production caller (§0). Status `draft`, blocked on decision Q1 in `decision-index.md`. This version authorizes no code change.

## §0 Blocking premise [HARD]

No Codex factory leader entry is reachable from the product on this tree, for two independent reasons.

1. **The refusal is a completed decision.** `moai codex` refuses every `-f` factory shape with one line that names `moai cc -f` or `moai glm -f` as the leader entries (`internal/cli/codex_launcher.go:801-802`). SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-004 requires that refusal, and its Out of Scope section states that a Codex leader "stays refused". SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-002 is the retirement that REQ-SD-004 narrowed.
2. **The code that would carry the adoption is not reached.** The Codex entry function that contains the leader branch (`internal/cli/codex_factory.go:90`) has one non-test caller, `internal/cli/codex_launcher.go:1214`, which runs only when the parsed factory entry is enabled. The parse enables it only for the lane token (`internal/cli/codex_factory.go:68`), and a lane token returns earlier, into the relaunch path (`internal/cli/codex_launcher.go:753`). SPEC-LAUNCHER-ENTRY-FLAGS-001 measured the same unreachability by overlay probe (its `progress.md`, the Codex lane reachability paragraph; cited here, not re-run) and left the removal of that branch as an operator decision (its spec.md O-5).

Consequence: wiring the adoption into that branch would change no user-visible behavior. Making the branch reachable reverses REQ-SD-004, a product-level change to a shipped command's refusal. That decision belongs to the operator (decision Q1). Until Q1 is answered `reopen`, the requirements in §3 that are gated on it are not binding and no run-phase work starts. If Q1 is answered `keep`, this SPEC is rejected (plan.md §C).

## §1 Background

- The adoption logic exists once. The cc and glm leader branches each call one shared combined gate and one shared start recorder (`internal/cli/factory_run_adopt.go:163-166` and `:173-184`; `internal/cli/cc.go:214,218`; `internal/cli/glm.go:274,281`). The candidate finder and the adoption write live in `internal/homestate/factory_run_adopt.go` (`:72-96` and `:112-145`). Commit `0ae31523f` (card t1707) wired the glm branch and touched those CLI files.
- The adoption rule is conservative. A single candidate whose owner classifies dead on its own recorded stamp, and which still holds live records, is adopted. Zero candidates mint a run id as before. Several candidates, or a single candidate whose owner is indeterminate, adopt nothing and produce one warning per candidate (`internal/homestate/factory_run_adopt.go:41-68`; `internal/cli/factory_run_adopt.go:64-112`).
- The resume-shaped signal is a Claude option spelling. The gate admits `--resume`, `-r`, their equals forms, and the attached short form, and it judges only explicit spellings (`internal/cli/factory_run_adopt.go:126-155`). `moai codex` has no resume verb: its verb table holds the empty verb, `cli`, `status`, and `app` (`internal/cli/codex_launcher.go:102-107`), and it forwards the operator's tail after `--` verbatim.
- The leader entry must capture the prior run id before it publishes the new one. The cc and glm branches do this (`internal/cli/cc.go:199`, `internal/cli/glm.go:262`), because the publication overwrites the environment (`internal/cli/factory.go:786-807`; `internal/cli/factory_launch_helpers.go:72-80`).
- No test drives the shared combined gate or the shared start recorder directly. The automatic path's live-legacy-peer refusal has no test; only the explicit path's refusal is tested (`internal/cli/factory_run_boundary_m1_test.go:57`).

## §2 Decision gate

Three decisions are recorded in `decision-index.md`. Q1 gates everything else: reopen the Codex leader entry, or keep it refused. Q2 is the resume-shaped carrier for a Codex start and applies only if Q1 is `reopen`. Q3 is removal of the unreachable interactive Codex lane branch and is independent of Q1 and Q2.

## §3 Requirements

REQ-CLR-001 and REQ-CLR-008 bind in both dispositions. REQ-CLR-002 binds only in the `keep` disposition. REQ-CLR-003 through REQ-CLR-007 bind only if Q1 is `reopen`.

- **REQ-CLR-001** (Ubiquitous) — The factory launcher shall apply one shared resume re-adoption decision to every factory leader entry it exposes, and no leader entry shall carry a second copy of the candidate selection or of the adoption write.
- **REQ-CLR-002** (Where the Codex leader entry remains refused) — Where decision Q1 is `keep`, the Codex launcher shall refuse every `-f` leader shape with the single REQ-SD-004 line and shall start no Codex-led factory run.
- **REQ-CLR-003** (Where the Codex leader entry is reopened, event-driven) — Where decision Q1 is `reopen`, when a Codex leader start is resume-shaped under the carrier that decision Q2 defines and its published run id was freshly minted, the Codex leader entry shall re-adopt the single prior active run that the shared decision selects, keeping that run's id.
- **REQ-CLR-004** (Where the Codex leader entry is reopened, event-driven) — Where decision Q1 is `reopen`, when no candidate run exists, the Codex leader entry shall mint a fresh run id, adopt nothing, and write no warning.
- **REQ-CLR-005** (Where the Codex leader entry is reopened, unwanted behavior) — Where decision Q1 is `reopen`, when more than one candidate run exists, or when the only candidate's owner is not proven dead on its own stamp, the Codex leader entry shall not adopt any run and shall write one warning per candidate that names the run and its owner classification.
- **REQ-CLR-006** (Where the Codex leader entry is reopened, unwanted behavior) — Where decision Q1 is `reopen`, when the operator supplied a run id explicitly or the launch environment already carried one, the Codex leader entry shall keep that run id and shall not adopt another run.
- **REQ-CLR-007** (Where the Codex leader entry is reopened, Ubiquitous) — Where decision Q1 is `reopen`, the Codex leader entry shall take its adoption decision and its start record from the shared decision and the shared recorder, the shared decision shall admit a Codex-led run to a Codex leader while refusing it to a cc or glm leader, and the shared decision shall refuse a run that holds a live legacy peer record to every leader.
- **REQ-CLR-008** (Ubiquitous) — Every leader entry shall adopt a run only when its owner classifies dead on the run's own recorded stamp; an indeterminate, peer-classified, or unstamped owner shall never be adopted.

## §4 Out of Scope

### Out of Scope — reopening the Codex leader inside this SPEC

- Lifting the REQ-SD-004 refusal. Decision Q1 makes that choice; this SPEC does not make it.
- Writing the Codex resume vocabulary. Decision Q2 makes that choice.

### Out of Scope — removing the unreachable Codex lane branch

- Deleting the leader branch of the Codex entry function, its sibling seams, or the fields that no parsed entry sets. Decision Q3 makes that choice.

### Out of Scope — the Codex lane relaunch

- The lane relaunch join and lane claims. Lanes join an existing run and never mint one, so run adoption does not apply to them.

### Out of Scope — cc and glm observable behavior

- Any change in what the cc or glm leader entries adopt, warn about, or record. The shared decision may gain a backend input only if every existing cc and glm outcome stays identical, as pinned by the regression guards in acceptance.md AC-007.

### Out of Scope — weakening adoption safety

- Adopting an indeterminate, peer-classified, or unstamped owner; adopting several candidates; replacing an explicit or environment-carried run id.

### Out of Scope — Codex hook coverage

- Codex does not wire PreCompact, PostCompact, PermissionRequest, or Interrupt, and Notification, PostToolUseFailure, TeammateIdle, and TaskCompleted never fire under Codex (AGENTS.md §9). Any leader behavior that depends on those events needs its own analysis if Q1 is `reopen`.

### Out of Scope — managed Codex sessions

- The managed-session divert (SPEC-FACTORY-MANAGED-SESSION-001) is not a leader entry and is not changed.

## §5 Acceptance

acceptance.md carries the acceptance criteria, the REQ-to-AC traceability, the edge cases, and the known gaps at plan time. plan.md carries the pre-flight gate, the known issues, the milestones, and the constraints.
