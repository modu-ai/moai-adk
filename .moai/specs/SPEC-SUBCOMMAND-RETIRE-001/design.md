# Design — SPEC-SUBCOMMAND-RETIRE-001

> Plan-phase design rationale. This is a removal/refactor SPEC; the "design" is the set
> of scope-boundary, sequencing, and replacement-path decisions that keep the removal
> safe (no broken CI, no capability gap, no FROZEN-zone violation, no orphaned cross-ref).

## §A. Design Decision 1 — Design-rule-subsystem scope boundary (FROZEN-respecting)

**Problem.** Removing `/moai design` could be read as "remove the entire design
subsystem," but that subsystem includes a FROZEN/zone-protected rule
(`.claude/rules/moai/design/constitution.md`), brand config (`.moai/project/brand/`,
`design.yaml`), and a real Go package (`internal/design/dtcg/` DTCG tokens that may
serve the `moai web` console).

**Decision.** Remove only the **command entry point** (`/moai design` command + workflow
file) and the **five design-pack skills**. Do NOT touch the FROZEN design methodology
rule, its `zone-registry.md` mirror entries (051-149), brand config, or `internal/design/`.

**Rationale.**
1. The design constitution rule header self-declares "FROZEN/EVOLVABLE zone"; the
   zone-registry exposes 100 FROZEN mirror entries for it. Editing it violates the zone
   contract and would fail `zone-registry` doctor checks.
2. The design methodology (GAN loop, brand integration) is reference documentation that
   can outlive the slash command; a 1-person OSS author may still consult it or drive
   brand work through `moai web`.
3. The `internal/design/` Go package is code, not a `/moai` workflow surface — it is
   outside this SPEC's "subcommand menu slimming" mandate and is exercised by other
   tests (`internal/design/dtcg/frozen_guard_test.go`).

**Consequence for the 0-dangling-ref AC (REQ-SCR-008).** The AC is scoped to **non-FROZEN,
in-scope surfaces** only (moai.md, CLAUDE.md, spec-workflow.md, glm-web-tooling.md,
moai-domain-humanize). References to removed skills/subcommands that remain inside FROZEN
design/constitution.md or its zone-registry mirror are **tolerated and explicitly
exempted** — they describe the design methodology, not a live command. This is the honest
boundary; pretending to achieve "0 refs everywhere" would require a FROZEN-zone violation.

## §B. Design Decision 2 — Security-audit capability replacement (no gap)

**Problem.** `/moai security` is the only dedicated OWASP audit entry point; removing it
risks a capability gap.

**Decision.** Route the capability through two retained paths and document them in the
SKILL.md router + a cross-reference note:
1. `/moai review --security` — `review` already spawns `Agent(general-purpose)` with
   security scope and exposes a `--security` flag (SKILL.md lines 173-175).
2. Natural-language "security audit" → `Agent(general-purpose)` loading the four retained
   defensive-cybersecurity reference skills (`moai-ref-owasp-checklist`, `moai-ref-secops`,
   `moai-ref-supply-chain`, `moai-ref-llm-security` — all optional-pack:devops, NOT removed).

**Rationale.** The `moai-ref-*` security skills were added by
`SPEC-V3R6-SEC-SKILL-INTEGRATION-001` precisely as the reusable security-knowledge layer;
the dedicated `/moai security` workflow is now a thin wrapper over a capability that the
`review` path + ref skills already provide. The Priority 3 router line "Security language
… routes to security" is rerouted to the review/Agent path.

**Other four retirements** carry no gap either (research.md §E): coverage→`go test -cover`/
`/moai gate`; brain→`/moai plan`; design→`moai web`/brand config; e2e→none (out of domain,
accepted).

## §C. Design Decision 3 — docs-site i18n handling (4-locale parity)

**Problem.** docs-site has dedicated `moai-coverage.md` + `moai-e2e.md` pages in all four
locales (en/ko/ja/zh) plus menu entries and cross-links, governed by the §17 4-locale
parity doctrine (a partial removal that drops only one locale is a parity defect).

**Decision.** Treat docs-site cleanup as a **dedicated milestone (M5)** removing the
coverage+e2e pages symmetrically across all four locales, the `main.yaml` menu entries,
and the cross-links in `quality-commands/{_index.md,_meta.yaml,moai-review.md,
moai-codemaps.md}` + ko `getting-started`/`core-concepts` mentions. `security` needs no
page removal (none exists). `brain`/`design` already removed by commit `1ece11578`.

**Rationale.** Leaving stale docs for removed commands is a user-facing defect, and the
prior commit set the precedent that docs-site is kept in lockstep. M5 is structured as a
distinct milestone so it can be split into a follow-up SPEC if run-phase scope pressure
demands — but the default is to complete it here.

## §D. Design Decision 4 — CI-guard count reconciliation method (RED-avoidance)

**Problem.** The catalog count constants (`expectedSkillCount=35`, `expectedTotal=42`,
`wantTotal=42`) and the `gen-catalog-hashes.go --all` hash refresh form a coupled set:
removing skill dirs without updating constants → RED tests; leaving catalog entries while
removing dirs → `make build` hashing error on a missing `SKILL.md`.

**Decision.** Sequence the catalog-affecting changes as one atomic milestone (M2) so the
tree never commits in a RED state:
1. Delete the 7 catalog.yaml entries (2 core-skill entries lines 91-98; 5 design-pack
   entries in the 169-196 region, preserving the `moai-domain-humanize` entry).
2. Delete the 7 skill directories in both trees.
3. Update the three count constants (re-derived by recount, not assumed): skills 35→28,
   total 42→35, with an appended history-comment entry citing this SPEC.
4. Delete `TestBrainCommandThinPattern`; fix `agentless_audit_test.go` `MODE_UNKNOWN`
   path list (drop `design.md`, keep `run.md`).
5. `make build` (regenerates `embedded.go`, refreshes remaining-entry hashes).
6. `go test ./internal/template/...` GREEN gate.

**Rationale.** The generator is hash-only (verified), so entry removal is manual and must
precede the build; the count constants must change in the same milestone as the dir
removal or the suite is RED between commits. Re-deriving counts by recount (rather than
hardcoding 28/35 from this design doc) satisfies the verification-claim-integrity
invariant — the run-phase agent observes the actual post-removal count.

**Counts model (verified, to be re-confirmed):** 35 skills + 7 agents = 42 entries today;
all 7 removals are skills (2 core + 5 optional-pack:design) → 28 skills + 7 agents = 35.

## §E. Alternatives Considered

| Alternative | Rejected because |
|-------------|------------------|
| Local-only hide (delete from `.claude/` only) | `moai update` re-deploys from template → reverts the hide. User explicitly chose the permanent template-source route. |
| Add a new `subcommand_retire_audit_test.go` guard (like seq-thinking precedent) | Unnecessary — the existing count constants + dynamic catalog tests already enforce absence after constant updates. A new guard adds maintenance weight for no extra coverage. |
| Remove the entire design subsystem in one SPEC | Scope explosion + FROZEN-zone violation + touches `internal/design/` Go code. Bounded to the command/skill entry points; subsystem retirement deferred. |
| Defer all docs-site cleanup | Leaves user-facing stale docs for removed commands, contradicting the precedent set by commit `1ece11578`. Completed as M5 (splittable). |

## §F. Anti-Patterns to avoid (run-phase guidance)

- **AP-1**: Editing the FROZEN `design/constitution.md` or its zone-registry mirror to chase
  "0 refs" — violates the zone contract. Exempt those surfaces (Decision §A).
- **AP-2**: Removing skill dirs before catalog entries → `make build` hash error on missing
  `SKILL.md`. Remove catalog entries in the same step (Decision §D).
- **AP-3**: Committing the tree between dir-removal and constant-update → RED suite. Keep M2
  atomic.
- **AP-4**: Hardcoding 28/35 from this doc without recount → unobserved-claim. Re-derive by
  recount during run-phase.
- **AP-5**: Dropping only one locale's docs page → §17 4-locale parity defect. Remove
  coverage/e2e symmetrically across en/ko/ja/zh.
- **AP-6**: Forgetting the `moai-domain-humanize` → `moai-domain-copywriting` cross-ref
  rewrite → orphaned reference in a retained skill (REQ-SCR-010).
