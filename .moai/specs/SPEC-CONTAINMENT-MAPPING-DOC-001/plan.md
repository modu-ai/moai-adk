# Plan: SPEC-CONTAINMENT-MAPPING-DOC-001

Tier M. One document, no code. Sections are ordered so that the decisions most likely to change come first and the mechanical steps come last.

## A. Context

- Deliverable: `.moai/docs/multi-agent-containment-mapping.md` (see `spec.md` § 1.2).
- Run-phase owner: `manager-develop`; sync-phase owner: `manager-docs`.
- Tree pin for every measurement in this plan and in `acceptance.md`: `629bb51cda73fc70422769988b8c28182c9bd2d5`.
- Premise check (planning run): the three anthropic.com engineering URLs are absent from `.moai/docs`, `.claude/rules`, `.claude/skills`, `AGENTS.md`, and `docs-site/content`; `ls .moai/docs | grep -c -i -E "contain|mapping|multi-agent"` printed `0`. One local rule mentions a merge queue in passing (the CI watch protocol, one line offering it as an option); it is not a comparison and was not opened beyond that line.

## B. Decisions (most likely to change first)

1. **Seed Fit calls** (§ D). Each `aligned` / `differs` / `gap` is a judgment. The author re-opens every cited file and may flip a call, provided the row still cites what was read.
2. **Document language.** Siblings in `.moai/docs` are written in the project conversation language (Korean), and the project's native-idiom rule applies to such artifacts. Assumption recorded here for correction: the body is Korean prose in a clean written register, with fixed tokens (`reference-only`, theme and Fit tokens, `VERIFIED` / `UNVERIFIED`), identifiers, paths, URLs, and verbatim source fragments kept in their original form. The humanize pass is applied inline where the humanize skill is not installed.
3. **Table shape.** Two tables with fixed columns (source register; mapping table). The acceptance commands key on the column order and the fixed tokens, so changing either means changing the commands in the same edit.
4. **Candidate rows not seeded** (the author may add one only if it can name a source row for it; otherwise drop it):
   - Independent review evidence (a correlated-failure defense): `.claude/skills/moai-ref-cross-model-audit/SKILL.md` § Independence rule (load-bearing) and `.claude/rules/moai/workflow/factory-dispatch-detail.md` § The verdict's home. No source row backs it.
   - Bounded unattended continuation: `.claude/rules/moai/workflow/goal-directive.md` § Goal-Presentation Timing names a ceiling; its value was not compared with the source's "two or three" figure.

## C. Verified source register (planning-run evidence)

Method: WebFetch of each URL, plus raw download and `grep -c -F` of each quoted fragment on the downloaded bytes. HTTP status of each raw download was 200.

| ID | WebFetch | Raw download | Fragments matched on raw bytes |
|---|---|---|---|
| BP-5 | yes | 200 | title 1; "source control became our biggest development bottleneck" 1; "around 2,500 code changes merged per day" 1; "we merged over 30,000 per day" 1; "A proper merge queue, for example" 1; "much sense at agent scale" 1; "global repository lock on the merge queue" 1; "post-merge and pre-deploy automated testing" 1; "concurrent changes can conflict" 1; "Claude can typically fix these conflicts faster than humans" 1 |
| BP-6 | yes | 200 | title 2; "users approved roughly 93% of permission prompts" 1; "Design for containment at the environment layer first, then steer behavior at the model layer" 1; "The goal is to set a hard boundary on what an agent can reach" 1 |
| BP-1 | yes | 200 | title 1; "Decoupling the brain from the hands meant the harness no longer lived inside the container" 1; "each could fail or be replaced independently" 1; "the tokens are never reachable from the sandbox where Claude" 1; "nothing in the harness needs to survive a crash" 1 |
| BP-2 | yes | 200 | title 2; "Treat a text-only end of turn as a report rather than as proof the task is done" 2; "stop after two or three automatic continuations on the same task" 2 |

A count above 1 means the page repeats the string; every fragment matched at least once. WebFetch output passes through a model extraction step, so only the raw-byte match counts as verbatim evidence.

## D. Seed mapping rows

Every file below was opened in the planning run, and every section name was confirmed as a real heading with `grep -n`. Fit is the planning-run judgment.

| Theme | Source | Source claim (fragment) | MoAI defense (opened) | Fit and basis |
|---|---|---|---|---|
| correlated-failure | BP-1 | "nothing in the harness needs to survive a crash" | `.claude/rules/moai/workflow/factory-dispatch.md` § The delegation channel is the queue; § Completion is read, never trusted; `.claude/rules/moai/workflow/auto-semantics.md` § 5.1 The standing awaken carrier | aligned. Delegation and completion evidence live on disk, not in a session; a stopped lane is re-read from disk. Scope of the claim: state kept outside the acting session, not a stateless harness. |
| correlated-failure | BP-1 | "each could fail or be replaced independently" | `.claude/rules/moai/workflow/factory-dispatch-gates.md` § Verification load incident record; `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Serializing a heavy run across lanes | differs. Lanes share one machine, which is a shared failure domain (the incident record: several lanes running the same full suite at once); the defense serializes heavy runs through a lease instead of separating the domains. |
| correlated-failure | BP-2 | "Treat a text-only end of turn as a report rather than as proof the task is done" | `.claude/rules/moai/workflow/cross-session-messaging-core.md` § An idle notice is a scheduling hint; `.claude/rules/moai/workflow/factory-dispatch.md` § Completion is read, never trusted | aligned. An idle notice says when to look and nothing about completion; the leader advances a card on evidence it read. |
| isolation-first | BP-6 | "Design for containment at the environment layer first, then steer behavior at the model layer" | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement; § Rules | differs. Same direction, different default posture: the doctrine layer is always loaded, while the mechanical PreToolUse layer is opt-in, inert by default in the distributed configuration, and fail-open on uncertainty. This repository's own setting was not read, so none is claimed. |
| isolation-first | BP-6 | "The goal is to set a hard boundary on what an agent can reach" | `.claude/rules/moai/workflow/worktree-integration.md` § Worktree Selection Rules; § Parallel-Session Branch Conflict Auto-Isolation; `.claude/rules/moai/workflow/factory-dispatch.md` § Isolation is provisioned by MoAI, then entered through a launcher | aligned for working-tree and branch-state reach (write-heavy agents isolated; a card runs in its own launcher-entered worktree). Scope note: this boundary is about the working tree, not about network or credential reach. |
| isolation-first | BP-6 | "users approved roughly 93% of permission prompts" | `.claude/rules/moai/workflow/auto-semantics.md` § 9. The gate inventory and dispositions | aligned in direction. Gates that were prompts for reversible work are autonomous with an evidence standard; a human keeps the three keep-set categories and the voluntary contract-signing path. The source's statistic is its own telemetry and is not restated as a fact about MoAI. |
| isolation-first | BP-1 | "the tokens are never reachable from the sandbox where Claude" | none located. Policy-level text only: `.claude/rules/moai/core/moai-constitution.md` § Security Boundaries | gap. The opened text states a policy (no secrets in version control, credentials via environment variables) and describes no structural separation of credentials from the agent's execution environment. Not evidence that none exists. |
| merge-queue-alternative | BP-5 | "post-merge and pre-deploy automated testing" | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Integration into the release branch is self-served; `.claude/rules/local/gitflow-lane-protocol.md` § 3. Serialization — one lane at a time in the merge window | differs. The factory keeps a first-in-first-out ticket queue behind one integration window, re-measures the candidate tree before the lane joins the queue, and keeps the in-window step to an identity check plus the merge, with no test suite inside the window. The source reports declining a global-lock merge queue at its volume and testing after merge. |
| merge-queue-alternative | BP-5 | "Claude can typically fix these conflicts faster than humans" | `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Integration into the release branch is self-served (the conflicts bullet) | aligned. The lane that owns the change resolves what its own merge raises; a semantic clash with another lane's merged change is a blocker report. |
| merge-queue-alternative | BP-5 | "source control became our biggest development bottleneck" | none located. No measurement of the factory's merge volume or of its bottleneck was opened | gap. The comparison therefore carries no scale claim about the factory. |

Also opened, not yet cited in a row: `.moai/docs/gitflow-integration-chain.md` (its operating procedure carries the re-measure-outside-the-window step; its operating-procedure label is a bold line, not a heading, so cite it as `§ operating procedure` only after re-reading), `.claude/rules/moai/workflow/resource-slot-lease.md`, `.claude/rules/moai/core/agent-common-protocol.md`.

## E. Template-deployment finding

Question: is `.moai/docs` shipped to user projects, so that a new document there would need a Template-First mirror?

Finding: partly. `.moai/docs` is tracked, but only a subset is mirrored into `internal/template/templates/.moai/docs/`, and the nearest siblings are not in that subset. A document that cites repo-local rule files and local-only siblings belongs with the local-only group. No mirror is needed.

Command evidence (planning run, tree pin in § A):

- `find internal/template/templates/.moai/docs -maxdepth 1 -type f | wc -l` printed `6`; `find .moai/docs -maxdepth 1 -type f | wc -l` printed `37`. The six mirrored files are `agent-lint.md`, `audit-artifact-convention.md`, `generic-patterns-guide.md`, `hook-stdin-fail-closed.md`, `session-handoff-appendix.md`, `todo-queue-storage.md`. None is a sibling named by the card.
- The local directory holds the git workflow doctrine, the integration-chain procedure, and the managed-session guide; none of the three is among the six mirrored files.
- `ls internal/template/templates/.moai/docs | grep -c -E "workflow-doctrine|integration-chain|managed-session"` printed `0`.
- `grep -n -F` on the Korean phrase for "local-only document" matched line 3 of `.moai/docs/gitflow-integration-chain.md`, which states in Korean that the document is local-only and is not mirrored to the template, because it contains internal card and SPEC identifiers and internal dates.
- `internal/template/spec_ceiling_mirror_test.go` pairs one `.moai/docs` file with its template mirror (`audit-artifact-convention.md`), which shows mirroring is chosen per file, not applied to the whole directory. `internal/template/github_flow_sweep_guard_test.go` sweeps `.moai/docs/*.md` as a surface with a file-count floor of 33; adding one file raises the count and the guard looks for text naming the integration branch as this project's base, which REQ-CMM-010 already excludes.

Gap: whether `moai update` preserves or removes untracked-by-template files in `.moai/docs` was not measured here. It does not change the finding, because the file is tracked in the repository.

## F. Milestones (execution order, priority High unless noted)

1. **M1 — Re-verify the four sources.** Fetch each URL, download raw, grep each fragment to be quoted. Record the flag per source from this run. (Mechanical; Priority High, because every later row depends on the flags.)
2. **M2 — Author the document skeleton.** Opening reference-only statement, source register (fixed columns), mapping table header, scale caveat for the merge-queue theme. Korean prose per § B.2.
3. **M3 — Fill the mapping rows.** Start from § D, re-open each cited file, keep or flip each Fit, paraphrase rather than copy. Keep both gap rows unless a defense is actually found. Add candidate rows from § B.4 only with a source.
4. **M4 — Run the acceptance battery.** Run E01 to E09 from `acceptance.md` on the document, then the scope guard E10 against the plan-commit SHA. Apply the inline humanize pass first, then re-run, because wording edits can introduce a forbidden token.

## G. Risks and residual risk

- **Paraphrase drift.** The defense cells paraphrase rule text; a paraphrase can overstate. Mitigation: each Fit basis in § D says what was and was not read. Residual: the marker-token check (AC-CMM-009) cannot see a paraphrase that restates a rule; a reviewer reads the defense cells against the cited sections (AC-CMM-014).
- **Moving upstream pages.** A source page can change after the planning run. The author re-verifies in M1; a failed re-fetch yields `UNVERIFIED`, and REQ-CMM-005 then applies by hand (AC-CMM-012 is reviewer-checked and dormant while all four flags are VERIFIED).
- **Repository drift.** Rule section names can change between planning and run. AC-CMM-007 checks paths only; a renamed section is caught by M3 re-reading, not by a command.
- **Observed-failure scope.** The acceptance commands were each observed red on the absent document and on a constructed failing input; the criterion-level evidence is in `acceptance.md`.

## H. Cross-references

- SPEC frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Verification claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md`
- Verification completeness (observed-failure completion, two-cell adoption): `.claude/rules/moai/development/verification-completeness.md`
- Template isolation doctrine: `.moai/docs/template-internal-isolation-doctrine.md`
