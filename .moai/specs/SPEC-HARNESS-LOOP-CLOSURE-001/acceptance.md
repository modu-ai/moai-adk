# Acceptance Criteria — SPEC-HARNESS-LOOP-CLOSURE-001

**AC count**: 8 (AC-HLC-001 .. AC-HLC-008)

All Go test idioms reference the harness package under `internal/harness/`. Fixtures use
`t.TempDir()` so no live `.moai/harness/` runtime artifact is mutated by the suite. Each AC
is independently testable. The "first clean live apply" against a real `my-harness-*` skill
is an OPERATIONAL follow-through (see spec.md §Exclusions) and is NOT an AC here — these ACs
prove the Go mechanism is complete + correct + reversible + auditable.

---

## AC-HLC-001 — Lineage entry appended on approved apply (REQ-HLC-001/003)

**Given** an approved proposal targeting a `my-harness-*/SKILL.md` fixture under `t.TempDir()`
and an empty manifest path,
**When** `Apply()` resolves to `DecisionApproved`, creates the snapshot, and modifies the
SKILL.md description,
**Then** exactly one lineage entry is appended to the manifest with `decision: "approved"`,
`applied_surface` equal to the modified field key (`"description"`), and non-empty
`proposal_id` / `target_path` / `timestamp`.

```bash
go test ./internal/harness/ -run 'TestLineage.*Approved|TestApply.*Lineage.*Approved' -count=1 -v 2>&1 | tee /tmp/ac-hlc-001.log
# Anti-vacuous guard (L_ac_run_pattern_vacuous_guard): the -run alternation MUST match a real test.
grep -q -- '--- PASS' /tmp/ac-hlc-001.log && ! grep -q 'no tests to run' /tmp/ac-hlc-001.log || { echo "AC-HLC-001 VACUOUS-OR-FAIL"; exit 1; }
# Expected: ≥1 '--- PASS' line AND no 'no tests to run'; one approved entry, applied_surface=="description", LoadManifest len==1.
```

---

## AC-HLC-002 — Lineage entry appended on rejected apply, active harness unchanged (REQ-HLC-004/008)

**Given** a proposal whose `target_path` is in the FROZEN zone (e.g. a relative
`.claude/agents/moai/x.md` fixture) and an empty manifest path,
**When** `Apply()` evaluates the safety pipeline and L1 Frozen Guard returns
`DecisionRejected` (`RejectedBy: 1`),
**Then** exactly one lineage entry is appended with `decision: "rejected"` and a non-empty
`reason`, AND no SKILL.md is written AND no snapshot directory is created for the rejected
proposal (active harness unchanged).

```bash
go test ./internal/harness/ -run 'TestLineage.*Rejected|TestApply.*Frozen.*Lineage' -count=1 -v
# Expected: PASS — one rejected entry; the rejection error is still returned; no file/snapshot mutation.
```

---

## AC-HLC-003 — Pending (human-gate) path writes NO lineage and returns ApplyPendingError (REQ-HLC-005/006)

**Given** a proposal that passes L1–L4 with `auto_apply: false` (the L5 human gate),
**When** `Apply()` evaluates and the pipeline returns `DecisionPendingApproval`,
**Then** `Apply()` returns an `*ApplyPendingError` carrying the `OversightProposal` payload
AND NO lineage entry is written (the manifest remains empty — pending is not a transition).

```bash
go test ./internal/harness/ -run 'TestApply.*Pending|TestLineage.*Pending.*NoEntry' -count=1 -v
# Expected: PASS — errors.As(*ApplyPendingError) true; OversightPayload non-nil; LoadManifest len==0.
```

---

## AC-HLC-004 — Apply writes ONLY to the my-harness-* description field, preserving the rest (REQ-HLC-007)

**Given** a `my-harness-x/SKILL.md` fixture with frontmatter (name, description, other
fields) and a body,
**When** an approved `Apply()` enriches the description,
**Then** ONLY the description field is changed; all other frontmatter fields and the entire
body are preserved byte-for-byte.

```bash
go test ./internal/harness/ -run 'TestApply.*PreservesFrontmatterBody|TestEnrichDescription' -count=1 -v 2>&1 | tee /tmp/ac-hlc-004.log
# Anti-vacuous guard (L_ac_run_pattern_vacuous_guard): the -run alternation MUST match a real test.
grep -q -- '--- PASS' /tmp/ac-hlc-004.log && ! grep -q 'no tests to run' /tmp/ac-hlc-004.log || { echo "AC-HLC-004 VACUOUS-OR-FAIL"; exit 1; }
# Expected: ≥1 '--- PASS' line AND no 'no tests to run'; non-description frontmatter lines and body bytes identical pre/post.
```

---

## AC-HLC-005 — Manifest is append-only and LoadManifest round-trips (REQ-HLC-001/002)

**Given** an empty manifest path,
**When** `WriteLineageEntry()` is called twice with two distinct entries,
**Then** `LoadManifest()` returns exactly two entries in write order, each round-tripping
its fields; AND `LoadManifest()` on a non-existent path returns an empty slice with no error
(backward compatibility).

```bash
go test ./internal/harness/ -run 'TestLineage.*AppendOnly|TestLoadManifest' -count=1 -v
# Expected: PASS — two writes → len==2 in order; missing-file → (empty, nil).
```

---

## AC-HLC-006 — Snapshots directory created on first apply (REQ-HLC-009)

**Given** a snapshot base directory that does not yet exist (mirroring the live tree where
`.moai/harness/snapshots/` is absent) and an approved proposal,
**When** the first `Apply()` runs,
**Then** the snapshot base directory is created and contains a per-apply snapshot dir with a
byte-for-byte backup of the pre-apply SKILL.md plus a `manifest.json`, written BEFORE the
SKILL.md modification.

```bash
go test ./internal/harness/ -run 'TestApply.*SnapshotCreatedFirstApply|TestSnapshot.*Created' -count=1 -v
# Expected: PASS — snapshot base dir + per-apply dir + manifest.json exist; backup == pre-apply bytes.
```

---

## AC-HLC-007 — RestoreSnapshot restores byte-identical prior SKILL.md (REQ-HLC-010)

**Given** a SKILL.md fixture that was modified by an approved `Apply()` (with a snapshot
created),
**When** `RestoreSnapshot()` is invoked against that snapshot directory,
**Then** the SKILL.md is restored byte-identically to its pre-apply content.

```bash
go test ./internal/harness/ -run 'TestRestoreSnapshot.*ByteIdentical|TestApply.*Rollback' -count=1 -v
# Expected: PASS — post-restore bytes == pre-apply bytes (rollback path proven).
```

---

## AC-HLC-008 — Backward compat + C-HRA-008 boundary + template neutrality (REQ-HLC-011/012)

**Given** the post-change tree,
**When** running the full harness suite and the boundary/neutrality checks,
**Then** the full suite is green, the C-HRA-008 subagent-boundary guard is green (no
`AskUserQuestion` in `internal/harness/` outside `_test.go`), the protected runtime
artifacts are unchanged in the working tree, and no lineage code leaked into the templates.

```bash
# Full suite
go test ./internal/harness/... -count=1
# Expected: ok (all packages).

# C-HRA-008 boundary (test + raw grep)
go test ./internal/harness/ -run TestSubagentBoundary -count=1
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[[:space:]]*//" || echo "BOUNDARY_CLEAN"
# Expected: TestSubagentBoundary PASS; grep prints BOUNDARY_CLEAN (no matches).

# Backward compat — protected artifacts untouched
git status --porcelain .moai/harness/usage-log.jsonl .moai/harness/learning-history/tier-promotions.jsonl
# Expected: empty output (no diff to either protected artifact).

# Template neutrality
grep -rn 'LineageEntry\|WriteLineageEntry\|manifest.jsonl' internal/template/templates/ || echo "TEMPLATE_CLEAN"
# Expected: TEMPLATE_CLEAN (no leak into templates).
```

---

## Definition of Done

- [ ] AC-HLC-001 .. AC-HLC-008 all pass on the post-change tree.
- [ ] `LineageEntry` struct added to `types.go` (additive, `omitempty` optional fields).
- [ ] `internal/harness/lineage.go` exists with `WriteLineageEntry()` + `LoadManifest()`
      (mirrors the `learner.go:145 WritePromotion` append-JSONL idiom; manifest path is a parameter).
- [ ] `Apply()` writes one `decision:"approved"` lineage entry after the SKILL.md modify and
      one `decision:"rejected"` lineage entry at the `DecisionRejected` early-return; the
      pending path writes NO lineage.
- [ ] `createSnapshot` / `RestoreSnapshot` / safety pipeline / L1 guard are UNCHANGED.
- [ ] `auto_apply: false` (harness.yaml) unchanged; no autonomous apply path added.
- [ ] FROZEN constants unchanged (tier thresholds, score dimensions, rubric anchors,
      Security must-pass floor).
- [ ] No quality verdict / scorer gate added (proof-of-mechanism only).
- [ ] Full `go test ./internal/harness/...` green; coverage on the lineage path ≥ 85%.
- [ ] Cross-platform build green (`go build` + `GOOS=windows GOARCH=amd64 go build`).
- [ ] C-HRA-008 boundary green; protected runtime artifacts unchanged; templates clean.
- [ ] `golangci-lint run ./internal/harness/...` clean.

## Edge Cases Covered

| Edge case | Handling | AC |
|-----------|----------|-----|
| Approved apply | snapshot → modify → append `approved` lineage | AC-HLC-001, AC-HLC-006 |
| Frozen-zone target (non-`my-harness-*`) | L1 reject → append `rejected` lineage, no file/snapshot mutation | AC-HLC-002 |
| Human-gate pending (`auto_apply: false`) | `ApplyPendingError`, NO lineage write | AC-HLC-003 |
| Other L2–L4 rejection | `rejected` lineage with layer reason, active harness unchanged | AC-HLC-002 (same reject path) |
| Manifest file absent | `LoadManifest` returns `(empty, nil)` — backward compat | AC-HLC-005 |
| Two consecutive transitions | two append-only entries in order | AC-HLC-005 |
| Snapshots dir absent (first ever apply) | `createSnapshot` MkdirAll creates base dir before write | AC-HLC-006 |
| Rollback after apply | `RestoreSnapshot` restores byte-identical SKILL.md | AC-HLC-007 |
| Protected runtime artifacts | untouched (manifest.jsonl is a new additive file) | AC-HLC-008 |
