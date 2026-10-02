# Decision index — SPEC-ASIDE-BROWSER-001

Decisions surfaced while authoring this SPEC. Each row states what was or is unresolved and why; none carries a preferred answer. Q1 to Q3 were answered by the operator on 2026-10-02 through the lane's question channel and are recorded below with that date. Q4 to Q8 are open; the SPEC text records the working assumption each was drafted against, and a verdict can change it.

### Q1: Does "never call from a subagent alone" mean orchestrator-only execution, or delegated execution with explicit authorization?

Label: FOUNDER
Authority anchor: (none — no committed artifact on the authority register decides it.)
Why unresolved: the card line reads two ways. Orchestrator-only: the orchestrator itself runs every Aside step and no subagent ever invokes Aside; this removes any subagent path to the operator's browser but moves Aside execution out of the e2e-tester. Delegated with authorization: a subagent may run Aside only on a spawn prompt that carries the operator's explicit request and the read-only boundary; this keeps the e2e-tester as executor but the boundary then rests on prompt content a subagent cannot authenticate.
Operator verdict: DECIDE (operator-answered 2026-10-02) — orchestrator only. No subagent, including the e2e-tester, ever invokes `aside` (exec or repl) or an Aside MCP tool; the orchestrator runs the Aside steps itself. Applied in REQ-ASB-006, REQ-ASB-007, REQ-ASB-010; the e2e-tester recipe block and the delegated-prompt subtests were removed, and one negative sentence (skill and e2e-tester definition) carries the mechanical anchor.

### Q2: Which catalog tier holds the skill?

Label: FOUNDER
Authority anchor: (none — `internal/template/catalog.yaml` defines the tiers but states no rule for placing an opt-in tool skill.)
Why unresolved: candidates were `optional-pack:testing` (an existing pack with an empty skill list, described as covering E2E testing), another optional pack, or core. Measured at plan-audit and re-read here: `TestSlimFS_HidesNonCoreEntries` pins that every non-core entry is hidden from slim installs, and `internal/cli/init.go:85` documents `--all` as the bypass, so an optional-pack skill would be absent from a default install while core `moai` (which carries `workflows/e2e.md`) and core `e2e-tester` name it. Core adds the skill's description to every install's skill listing.
Operator verdict: DECIDE (operator-answered 2026-10-02) — core. `moai-ref-aside-browser` is catalogued under `catalog.core.skills`. The optional-pack placement was rejected for the measured slim-install reason above. Consequences for guards are in `acceptance.md` AC-ASB-003 and `plan.md` § B (baseline run E11: no hard-coded core or total count).

### Q3: Must a silent fallback still appear in the run report?

Label: FOUNDER
Authority anchor: (none — `verification-claim-integrity.md` §3.1 covers a refused command inside a verification, not a toolchain substitution, so it does not answer this as written.)
Why unresolved: the card says Aside absent means "silently fall back". Read strictly, nothing is surfaced; read as "no prompt, no install, no failure", the run report could still carry one line naming the fallback. The two readings differ in whether a report can name the toolchain that ran without saying why Aside did not.
Operator verdict: DECIDE (operator-answered 2026-10-02) — completely silent. An absent Aside, or `CI=true`, continues on the default toolchain with no Aside-specific message, note, prompt, install attempt, or failure. The run report names the toolchain actually used in the normal way; that is not an Aside note. Applied in REQ-ASB-012; the fallback-note requirement was removed.

### Q4: With `--tool aside` and `CI=true`, does the workflow fall back or refuse?

Label: FOUNDER
Authority anchor: (none — `e2e.md` line 129 marks MCP-tier tools unavailable under `CI=true`, but a workflow file is not on the authority register and does not address an explicit request.)
Why unresolved: the card says Aside is "excluded when CI=true". An explicit request made in CI can either fall back to the default toolchain (silently, now that Q3 is answered) or stop with an error so the run does not substitute what the operator asked for. The draft (REQ-ASB-011, REQ-ASB-012) falls back silently.
Operator verdict:

### Q5: How is the screenshot evidence captured and persisted?

Label: EVIDENCE-NEEDED
Authority anchor: (none.)
Why unresolved: REQ-ASB-013 needs `aside repl` to produce a screenshot file saved under `e2e/`. The mechanism (a path option, returned bytes the orchestrator must write, or a sandbox that blocks writes) was not measured, because measuring it launches a browser in the operator's session. The requirement is worded without saying which side writes the bytes, and `plan.md` M3.0 measures it first.
Operator verdict:

### Q6: Should the skill be injected into testing missions through `domain_skills`?

Label: FOUNDER
Authority anchor: (none — `.moai/config/sections/delegation.yaml` `domain_skills` is an operator setting, but it names no rule for opt-in tools.)
Why unresolved: listing the skill under `domain_skills.testing` would make the orchestrator inject it into every matching mission, which makes Aside more reachable than "only when explicitly requested". Leaving it out means the skill is found only by description match or by the e2e workflow naming it. The draft (assumption A-4) leaves it out.
Operator verdict:

### Q7: How much documentation lands in docs-site?

Label: FOUNDER
Authority anchor: (none.)
Why unresolved: the draft limits docs-site to `guides/mcp-server.md` in four locales, including the numeral change from four to five documented-but-disabled entries. The card says "documented"; other candidates are `utility-commands/moai-e2e.md` in four locales, the docs-site skill-guide table, the README ref-skill lists and counts, or no docs-site change (skill and e2e workflow only). More pages raise the file count past the Tier M ceiling and add translation work in ja and zh; the README count of ref skills becomes stale by one in the draft. The draft (assumption A-5) takes the single-guide scope.
Operator verdict:

### Q8: With orchestrator-only execution, which e2e phases stay with the e2e-tester?

Label: FOUNDER
Authority anchor: (none — `e2e.md` assigns detection, journey mapping, script creation, execution, and recording to the e2e-tester, and no committed artifact says how an orchestrator-run toolchain divides that work.)
Why unresolved: Q1 moves Aside execution to the orchestrator. The e2e-tester could keep Phases 0 and 1 (detection and journey mapping, read-only) and hand Phases 2 and 3 for the Aside journeys to the orchestrator; or the orchestrator could run the whole Aside flow, journey mapping included, leaving the e2e-tester out of an Aside run. The first keeps the existing agent's read-only work and puts the step-by-step browser driving in the orchestrator's context (bounded by redirecting output to `e2e/.runs/`); the second removes the agent entirely at the cost of more orchestrator work. The draft (assumption A-8, `plan.md` M3.2) takes the first.
Operator verdict:
