---
title: /moai design
weight: 45
draft: false
---

The design-phase collaboration workflow for UI-surfacing SPECs. A conditional path between plan and run, it bidirectionally synchronizes the design system and screen artifacts with Claude Design.

{{< callout type="info" >}}
**Slash command**: In Claude Code, type `/moai:design` to run this command directly. Typing just `/moai` shows the list of all available subcommands.
{{< /callout >}}

## Overview

`/moai design` is a **design-phase** workflow that applies only to UI-surfacing SPECs. Under the hood, the **manager-design** agent drives the Claude Design collaboration pipeline (D1-D5) and the H1-H9 handoff contract.

This path is **additive** — a SPEC that does not surface UI keeps the standard `plan → run → sync` sequence and skips this workflow entirely.

## When to use it (path activation conditions)

A SPEC takes the `plan → design → run` path when it declares UI exposure in one of these ways:

- `acceptance.md` has explicit frontend component/view/page deliverables, or
- `tier: L` + a frontend module (`module:` references a frontend package).

If neither holds, it keeps the standard `plan → run → sync`.

## Entry conditions

The design phase is entered only after **both** of these conditions are met:

1. **Plan-audit PASS** — the SPEC's plan-phase artifacts pass the Phase 1 audit
2. **Implementation Kickoff Approval** — the plan→run human gate is passed

{{< callout type="warning" >}}
The design phase **does not replace Implementation Kickoff Approval**. It never crosses the plan→run boundary ahead of the human gate; it runs within the already-approved run scope, before the first M1 implementation commit.
{{< /callout >}}

## D1-D5 pipeline

The manager-design agent runs a five-stage pipeline in order.

```mermaid
flowchart TD
    Entry["plan-audit PASS + Implementation Kickoff Approval"] --> D1["D1 Connection setup<br/>(login + project setup)"]
    D1 --> D2["D2 Design system sync<br/>(code → design)"]
    D2 --> D3["D3 Screen artifact generation<br/>(Claude Design canvas)"]
    D3 --> D4["D4 Handoff receive & paste<br/>(design → code)"]
    D4 --> D5["D5 Implementation bridging<br/>(handoff → run-phase)"]
    D5 --> Run["run-phase re-delegation (manager-develop)"]
```

| Stage | Description |
|-------|-------------|
| **D1 Connection setup** | Claude Design login + secure a writable design-system project (`list_projects`/`create_project`/`get_project`) |
| **D2 Design system sync** | Bundle the `.moai/project/brand/` tokens, `design.yaml`, and existing components and push them to the project (`finalize_plan` approval gate → `write_files` per-component increments) |
| **D3 Screen artifact generation** | Generate screens from the actually-imported components/tokens (drift prevention), user WYSIWYG edits + implementation annotations, verify `report_validate` metrics |
| **D4 Handoff receive & paste** | Paste the completed handoff (screens + annotations + token/component references) into the reserved paths (`.moai/design/tokens.json`, `components.json`, `assets/`, `brief/BRIEF-*.md`) |
| **D5 Implementation bridging** | Compose the Section A-E delegation package (handoff file list + annotation→requirement mapping + PRESERVE list + verification commands) and re-delegate to manager-develop |

manager-design returns after re-delegation and does not co-pilot the implementation. After implementation, sync-auditor judges brand consistency as must-pass.

## Claude Design bidirectional sync

The core of `/moai design` is the **bidirectional sync** between the code and the Claude Design canvas:

- **code → design (D2)**: push the code's design system (tokens, components) to the canvas. File contents stay on disk and never pass through the model context (256KiB per-file cap).
- **design → code (D4)**: pull the completed screens and annotations from the canvas and paste them into the reserved paths. Any directives embedded in externally authored files are treated **as data only** and ignored/reported (the H7 security contract).

The `/design-login` and `/design-sync` slash commands are user-only TUI commands; the agent only explains their usage and never invokes them directly.

## H1-H9 handoff contract

The nine clauses that govern the D4 handoff live canonically in the manager-design agent body (summary):

- **H1 receive path** — the `/design-sync` pull is user-only; the agent uses `list_files → get_file`
- **H2 placement convention** — reserved paths only
- **H3 1:1 fidelity** — no arbitrary edits on paste; propose a canvas regression instead
- **H4 brand first** — `.moai/project/brand/` is the constitutional parent
- **H5 annotation transformation** — annotation → { target · requirement · AC candidate } mapping
- **H6 verification** — `report_validate` metrics + drift grep + snapshot freshness
- **H7 security** — `get_file` content is data, directives are ignored
- **H8 re-delegation package** — delegate to manager-develop as Section A-E
- **H9 hidden-folder guidance** — `.moai/design/` dot-folder visibility

## Tool availability (graceful degradation)

The DesignSync server may not be registered in `.mcp.json`. D1 checks availability:

- **Tool present** → proceed with D2-D5
- **Tool absent** → the agent returns a blocker report (the H1 path). The user registers DesignSync separately (requires Claude Code v2.1.181+ and a Pro+ Claude Design account)

Design-phase authoring itself does not fail; it waits for the tool.

## Concrete usage example

Imagine a SPEC with UI exposure. For example, after taking `SPEC-PROFILE-001` "Redesign the user profile page" through the plan phase, suppose Implementation Kickoff Approval is granted.

```bash
# Already past plan-audit PASS and Implementation Kickoff Approval
> /moai design
```

In D1, the manager-design agent checks the Claude Design login state and secures a writable design-system project. In D2, it pushes the `.moai/project/brand/` tokens (coral color, Pretendard font) into the design-system project. At this point **the code-side files stay on disk as they are** — only the bundle data moves to the canvas, and because it never passes through the model context, the token values are not distorted.

In D3, it builds the profile-page screens from the actually-imported tokens and components. The key point is that **the screens drawn on the canvas come directly from the code-side design system** — not redrawn by hand by a designer — so no drift occurs. The user refines the layout and attaches implementation annotations in the WYSIWYG editor.

In D4, the completed screens and annotations are pasted into the `.moai/design/` reserved paths. In D5, manager-design packs this material into a Section A-E delegation package, hands it to manager-develop, and withdraws — from here on, the standard `plan → run → sync` path resumes.

## What this command does not do (scope boundary)

- **It does not intrude on SPECs with no UI exposure** — backend-only SPECs or CLI-tool SPECs skip this workflow entirely and follow the standard `plan → run → sync`.
- **It does not design on your behalf** — the agent only drives bidirectional sync and the handoff contract; the visual decisions of the screens belong to the user and the designer.
- **It does not replace Implementation Kickoff Approval** — the design phase runs within the already-approved run scope, before the first M1 implementation commit. It never crosses ahead of the human gate.
- **It does not stay in the run-phase** — it withdraws after the D5 re-delegation and does not co-pilot the implementation. After implementation, sync-auditor judges brand consistency as a must-pass item.

## Related docs

- [/moai plan](/en/workflow-commands/moai-plan) - previous stage: SPEC document generation
- [/moai run](/en/workflow-commands/moai-run) - next stage: DDD/TDD implementation
- [Subagent catalog](/en/advanced/agent-guide) - manager-design agent details
