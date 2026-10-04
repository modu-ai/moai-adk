# SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 — Plan

> Tier M, three milestones, ordered by decision reversibility: the data model (manifest schema, entry grammar, categories) is the decision most likely to change, so it leads; guard behaviour follows; the mechanical liveness and evidence work closes. No time estimates — priority and order only. Document-level pin: tree SHA `e497f693608ac7ea45a08b06304dc585e927ff49`.

## §A Context

- Card t1510, operator decision D5, Class C. Scope bound by the leader: declaration, blocking, human routing. No new subsystem, no auditor-definition edit (card t1500 lands later).
- Plan-phase deliverables already on disk: the base-tree RED evidence under `evidence/` (probe, judge, recorded outputs, latency baseline). These are committed by the orchestrator ahead of any run-phase commit so the commit graph witnesses red-before-change.
- Reused rather than invented: the `HARNESS_FROZEN_*` sentinel family and its orchestrator pattern-match convention; the existing hook-block round in `agent-common-protocol.md` § Hook Invocation Surface as the human route; the config package's dedicated-loader pattern (`LoadHarnessConfig`, `LoadCrossSessionConfig`); the existing shell-segment splitter in `internal/hook`; the audit-row convention of the subagent write guard.

## §B Known issues to carry into run-phase

- **Config completeness audit.** `internal/config/audit_loader_completeness_test.go` fails `YAML_SECTION_NO_LOADER` for any template section file not registered. `protected-zone` is registered in the dedicated-loader list in the same commit that adds the file (the wiring audit beside it may need the same entry — read it before M1 ends).
- **Template-First.** The shipped manifest is authored under `internal/template/templates/.moai/config/sections/` first, `make build` regenerates the embed, then the local copy is made. `moai update` wipes `.moai/config/` and re-deploys the template; the dogfood overlay therefore lives at `.moai/project/protected-zone.yaml`, outside the managed roots (verify against the root list in `internal/cli/update/deploy/deploy.go` in M1 — do not rely on this plan's reading).
- **Template neutrality.** The shipped file must pass `template_neutrality_audit_test.go` and `internal_content_leak_test.go`: no SPEC identifiers, requirement tokens, audit citations, internal dates, commit hashes.
- **Matcher is pinned.** `Write|Edit|Bash` in both settings files is asserted by `agent_model_matcher_test.go` and `hmp_powershell_matcher_test.go`; this SPEC changes neither file.
- **Relative paths resolve against the process cwd** in the existing file-access check (measured). Reuse that resolution; do not introduce a second one.
- **The sentinel catalog test** (`internal/harness/sentinel_catalog_test.go`) asserts exactly the eight existing sentinels; add the ninth there.
- **Go test selector hazard.** `-run` matching zero tests prints `ok` and exits 0 (ledger L-6). Every test this SPEC adds to a package is a named subtest of one top-level `TestProtectedZone`, selected with the anchored `-run '^TestProtectedZone$'`, and every AC states a swept floor.
- **Tool provenance.** Run-phase measurements cite the judging build's commit next to the tree HEAD (`verification-claim-integrity.md` §2.2); the installed `moai` binary is not the judge — build from the tree and invoke by path.
- **Bash tool refusals in a worktree session.** A command that names git in a compound form, or runs a computed command, is refused by the worktree guard; use plain separate commands and script files. Name any refusal in Gaps (`verification-claim-integrity.md` §3.1).

## §C Milestones

### M1 — Manifest data model, loader, matcher (Priority High)

Decisions most likely to change; nothing here touches runtime behaviour yet.

- Define manifest schema v1: `version`, `categories.<name>.runtime` (bool) and `categories.<name>.paths` (list). Strict decode; unknown keys, a wrong `version`, an entry using a form outside the four, an entry containing `..`, or an absolute entry make the file invalid. Duplicate entries and empty categories are valid.
- Author the shipped manifest (spec §D shipped column) and the dogfood overlay (overlay column); the local `.moai/config/sections/protected-zone.yaml` is the shipped file plus nothing — the overlay carries the dogfood paths.
- Add the loader (two files, union, overlay add-only, category order = shipped first then overlay) in `internal/config` beside the other dedicated loaders; register the section in the completeness audit.
- Add the pure matcher: normalization (REQ-SIPZ-006) and the four-form comparison, taking the root and the path as strings so the same table runs on every OS.
- Tests (subtests of `TestProtectedZone` in their package): loader valid/invalid table including the edge cases in `acceptance.md` §D, overlay-cannot-narrow, template-neutrality, shipped⊆dogfood, Windows-shaped normalization table.
- Flips: AC-SIPZ-009, AC-SIPZ-012; AC-SIPZ-004 loader half.

### M2 — Guard behaviour for file tools (Priority High)

- In `internal/hook`, add the guard module and one call site in the Write/Edit branch of the PreToolUse handler. The order inside the branch: identity gate (compiled set) → normalize → compiled baseline first (so P1–P4 and S1 keep their legacy sentinels) → manifest. Normalization is also applied to the existing baseline check by routing it through the new normalizer (the one change to existing code; AC-SIPZ-002 and AC-SIPZ-005 are the pair that watches it).
- Failure modes: present-but-invalid → deny every identity Write/Edit/Bash-mutation with `manifest=invalid`; absent → baseline floor plus audit row.
- Deny reason format and the new sentinel constant; audit row append to `.moai/logs/protected-zone-audit.jsonl` (single append, fail-open on write error with a stderr notice).
- Cost seam: the manifest is opened only after the identity and tool gates; a test injects a read counter.
- Update the sentinel catalog test (ninth sentinel) and the baseline-coverage drift test (count 21).
- Flips: AC-SIPZ-001, 002, 004, 005, 006, 007 (read-counter half), 008, 010.

### M3 — Shell rule, liveness, evidence (Priority Medium)

- The Bash rule: split into segments with the existing splitter, tokenize, pair a mutating verb with a zone-covered argument or redirection target. Under-match and pass on anything unclassifiable.
- Liveness test (REQ-SIPZ-015): matcher group of both settings files, both manifests parse, dead-entry sweep with the `runtime` exemption and a non-zero swept count, end-to-end handler denial of a known input.
- Execute every mutant in `acceptance.md` §C and record the observed red in `progress.md` §E.2 — the checks are unfinished until each red has been seen.
- Re-run the probe and the judge against a binary built from the final tree: expect `JUDGE swept=47 expected=47 fail=0`. Run the paired latency A/B (base build vs final build, same session).
- `make build`, `go vet`, `golangci-lint run`, `go test -race ./internal/hook/...`, and the template and config packages.
- Flips: AC-SIPZ-003, 007 (latency half), 011, 013.

## §D Technical approach, in decision order

1. **Entry grammar** — four forms only, ASCII case-folded, no negation (spec §C.5). Changing the grammar later changes every manifest, so it is fixed first.
2. **Layering** — shipped ∪ overlay ∪ compiled floor; overlay add-only; invalid anywhere → closed for the identity set (spec §C.4).
3. **Identity gate** — compiled constant, checked before any file I/O.
4. **Order inside the guard** — baseline first (legacy sentinels survive), manifest second (new sentinel), shell rule third.
5. **Human route** — the denial reason; no new channel (spec §C.7). No rule or skill file edit is planned: the route rides the existing hook-block doctrine. If the audit finds the doctrine silent about `HARNESS_FROZEN_*` denials, record it as a finding rather than widening scope.
6. **Liveness** — CI test, not a runtime heartbeat (spec §C.8).

## §E Risks

| Risk | Mitigation |
|---|---|
| The guard blocks a legitimate learner write | The legitimate surface (`.claude/agents/harness/`, `hns-*`, `.moai/harness/main.md`, `.moai/specs/`) is probed as controls N5, N7, N8, N4 and stays outside the zone |
| `moai update` weakens the zone | Shipped manifest is regenerated (heals); overlay is outside the wiped roots; the dead-entry sweep and AC-SIPZ-009's superset assertion catch a shrunken dogfood zone |
| Invalid manifest locks the learner out | Intentional and loud: the reason names the file; the audit row records it; `moai update` restores the shipped file; non-identity callers are untouched |
| Shell rule over-blocks (e.g. `cp` from a zone path) | Accepted: the rule is scoped to the identity and under-matches elsewhere; read-only verbs are probed as controls |
| Existing normalization change alters a legacy decision | AC-SIPZ-002 and 005 pair; the legacy sentinels are asserted byte for byte |
| Latency regression on a hot path | Manifest read only after the identity gate; paired A/B budget in AC-SIPZ-007 |
| A path nobody listed stays unprotected | Disclosed (spec §F G4); the manifest is reviewed against each new apparatus file, and card t1500's files are Q4 |

## §F Rollback

Revert the run-phase commits. Operationally, deleting the manifest files returns the guard to the compiled baseline floor (absent ≠ invalid); a corrupt file is the case that fails closed, so the rollback is to delete or restore it, never to edit it.

## §G Cross-references

- `spec.md` §B requirements, §C decisions, §D initial content; `acceptance.md` §E ledger; `decision-index.md` open questions Q1–Q6.
- `.claude/rules/moai/development/verification-completeness.md` — the two-cell and mutant-probe obligations this plan follows.
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement — identity reach and the no-bypass lesson.
