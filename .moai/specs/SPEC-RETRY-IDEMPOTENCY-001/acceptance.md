# Acceptance Criteria — SPEC-RETRY-IDEMPOTENCY-001

## §D. AC Matrix

| AC ID | Covers REQ | Requirement | Verification | Severity |
|-------|------------|-------------|--------------|----------|
| AC-RI-001 | REQ-RI-001 | Idempotency-asymmetry augmentation present in deployed rule | grep the deployed file for asymmetry keywords | MUST |
| AC-RI-002 | REQ-RI-007 | Augmentation block byte-identical in template mirror (region-scoped) | region-scoped diff of the § Error Recovery Pattern augmentation block only (NOT whole-file — see D2 note) | MUST |
| AC-RI-003 | REQ-RI-003 | Observe-before-retry gate wording present | grep for "observe" + "side-effect" proximity | MUST |
| AC-RI-004 | REQ-RI-004 | Duplicate-effect hazard named | grep for "duplicate" (commit/PR/deploy) | MUST |
| AC-RI-005 | REQ-RI-005 | Existing 4-step Error Recovery Pattern unchanged | grep the 4 step lines verbatim | MUST |
| AC-RI-006 | REQ-RI-005 | Constitution "Maximum 3 retries" line unchanged | grep the constitution line verbatim | MUST |
| AC-RI-007 | REQ-RI-006 | Augmentation references step 3 ("do not retry the identical call") | grep for the step-3 reference | MUST |
| AC-RI-008 | REQ-RI-008 | Deployed rule neutrality — no internal trace | CI leak guard `go test ./internal/template/...` | MUST |
| AC-RI-009 | REQ-RI-008 | Template-neutrality guard passes | CI neutrality workflow / local test | MUST |
| AC-RI-010 | REQ-RI-007 | `make build` run; augmentation keywords present in BOTH template `.md` AND regenerated `embedded.go` | grep "side-effect", "idempotent", "observe" in the template `.md` AND `internal/template/embedded.go` (mirror-presence guard for a non-`workflowOptMirroredPaths` file) | MUST |
| AC-RI-011 | REQ-RI-002 | Idempotent-retry allowance wording present | grep for "idempotent" + "read-only" | MUST |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Augmentation present and mirrored (AC-RI-001, AC-RI-002)

- **Given** the run-phase edit is complete,
- **When** an auditor greps `.claude/rules/moai/core/agent-common-protocol.md` and
  `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` for the
  retry-safety-asymmetry augmentation,
- **Then** both files contain the augmentation block, and a `diff` of the two files shows no
  difference in the § Error Recovery Pattern region (byte-identical mirror).

### Scenario 2 — Observe-before-retry gate for side-effecting calls (AC-RI-003, AC-RI-004)

- **Given** the augmentation is present,
- **When** an auditor reads the side-effecting-calls bullet,
- **Then** it states that on an ambiguous failure the actor must first observe current state
  and retry only when the effect is confirmed absent, AND it names the duplicate-effect
  hazard (duplicate commit / duplicate pull request / double deploy).

### Scenario 3 — Existing policy preserved (AC-RI-005, AC-RI-006, AC-RI-007)

- **Given** the augmentation is appended,
- **When** an auditor compares the § Error Recovery Pattern steps 1-4 and the constitution
  "Maximum 3 retries per operation" line against their pre-edit form,
- **Then** the 4 steps and the 3-retry line are unchanged (verbatim), AND the augmentation
  explicitly references step 3 ("do not retry the identical call") as the rule it extends.

### Scenario 4 — Deployed-rule neutrality (AC-RI-008, AC-RI-009)

- **Given** the augmentation is present in the template source,
- **When** `go test ./internal/template/...` and the template-neutrality guard run,
- **Then** both pass — the augmentation prose contains no SPEC-ID token, no REQ/AC token, no
  audit citation, no internal date, no commit SHA, and no archive/memory path.

### Scenario 5 — Build parity + mirror-presence guard (AC-RI-010)

- **Given** the template mirror is edited,
- **When** `make build` runs and the embedded template is regenerated,
- **Then** a grep for the augmentation keywords ("side-effect", "idempotent", "observe")
  matches in BOTH the template `.md` AND the regenerated `internal/template/embedded.go`.
  This is the mirror-presence guard: because `agent-common-protocol.md` is outside
  `TestRuleTemplateMirrorDrift`, the leak guard alone would pass even if the template-mirror
  edit were forgotten — the embedded.go grep is what catches a forgotten mirror.

## §D.2 Edge Cases

- **Idempotent-only failure** — a Read/Grep/Glob transient failure retried up to the ceiling
  must remain permitted by the augmentation (AC-RI-011, MUST — covers the normative
  REQ-RI-002 shall-clause); the gate applies only to side-effecting calls.
- **Ambiguous vs unambiguous side-effect failure** — an *unambiguous* side-effect failure
  (the effect definitely did not land) is not gated by "observe first"; the gate targets the
  *ambiguous* case. The wording must scope the gate to ambiguous failures.

## §D.3 Definition of Done

- All MUST-severity ACs (AC-RI-001..011 — note AC-RI-011 is now MUST) pass.
- The augmentation block (§ Error Recovery Pattern insertion) is region-scoped byte-identical
  between deployed + template; whole-file byte-parity is NOT expected (the file carries
  internal SPEC-IDs the template strips — see spec.md §B REQ-RI-007 note). `make build` run;
  augmentation keywords present in `embedded.go`.
- CI internal-content leak guard + template-neutrality guard both green.
- Existing 4-step pattern + constitution 3-retry line verbatim-unchanged.

## §D.4 Quality Gate Criteria

- No Go code changed → no coverage delta expected; `go test ./...` remains green.
- `golangci-lint run` baseline clean (doc-only change touches no Go source).
- Doc-only Tier S change; sync-auditor may be a consolidated lightweight pass.
