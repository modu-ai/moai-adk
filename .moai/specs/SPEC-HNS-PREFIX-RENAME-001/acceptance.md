---
id: SPEC-HNS-PREFIX-RENAME-001
version: "0.1.1"
updated: 2026-07-13
document: acceptance
---

# Acceptance Criteria — SPEC-HNS-PREFIX-RENAME-001

## §D AC Matrix

All commands run from the repo root. Every AC that verifies matching logic requires an EXECUTED command with verbatim output (reachability, not token presence — §D.2). All greps are case-sensitive (`grep` WITHOUT `-i`, per plan.md §B.2).

| AC | Verifies | Procedure | PASS condition |
|----|----------|-----------|----------------|
| AC-HPR-001 | REQ-HPR-001 | `grep -rn 'harness-<name>' internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/workflows/harness-builder.md internal/template/templates/.claude/skills/moai/workflows/harness-build-entry.md internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` AND same grep with `hns-<name>` | `harness-<name>` = 0 matches; `hns-<name>` ≥ 1 match in EACH of the 4 files; all six artifact-type path contracts present in `harness-builder.md` (Runner `.claude/workflows/hns-<name>-run.js`, agents `.claude/agents/harness/hns-<name>-*-specialist.md`, skills `.claude/skills/hns-<name>-*/`, verify `hns-<name>-verify`, manifest `runner_workflow`, thin-command reference) |
| AC-HPR-002 | REQ-HPR-002 | `diff -q` each of the 4 template↔live pairs | 4/4 byte-identical |
| AC-HPR-003 | REQ-HPR-005, REQ-HPR-008 | `go test ./internal/cli/ -run 'TestUserOwned.*HNS|TestUpdateNamespace.*' -v` (new unit tests: classifier returns user-owned for `.claude/skills/hns-x/SKILL.md` and `.claude/workflows/hns-x-run.js`; returns NOT-user-owned for `.claude/skills/moai-harness-learner/SKILL.md` and `.claude/skills/hnsx-foo/SKILL.md`) | all PASS |
| AC-HPR-004 | REQ-HPR-006, REQ-HPR-007 | E2E sandbox go test: plant `hns-*`, `harness-*`, `my-harness-*` skill dirs + `hns-*`/`harness-*` workflow JS + agents under `.claude/agents/harness/` in `t.TempDir()`, execute the update preservation flow | every planted artifact survives byte-identical; zero deletions |
| AC-HPR-005 | REQ-HPR-008 | Existing template-managed overwrite test (moai-harness-* skills still template-synced) re-run | PASS unmodified — `moai-harness-learner` never classified user-owned |
| AC-HPR-006 | REQ-HPR-010, REQ-HPR-011 | `go test ./internal/cli/harness/ -run 'TestV4.*' -v` with new dual-prefix cases: list enumerates `hns-x-*` artifacts; remove on a mixed-generation harness (`hns-<n>-run.js` + `harness-<n>-a-specialist.md`) removes both | all PASS |
| AC-HPR-007 | REQ-HPR-012, REQ-HPR-009 | `go test ./internal/cli/harness/ -run 'TestDoctor' -v` + `go test ./internal/harness/ -run 'PrefixConflict|Frozen' -v` with `hns-` cases | all PASS; `runnerSpecialistRE` matches `hns-x-specialist` AND `harness-x-specialist`; a doctor test case with manifest `runner_workflow: hns-x-run.js` resolves the Runner with NO doctor Runner-resolution code change (manifest-driven, prefix-agnostic — REQ-HPR-012) |
| AC-HPR-008 | REQ-HPR-013, REQ-HPR-018..021 | Post-M3, live repo: `go run ./cmd/moai harness doctor; echo exit=$?` and `go run ./cmd/moai doctor 2>&1 \| grep 'hns-'` | `harness doctor` exit 0 with all specialist/Runner references resolving; `moai doctor` classifies `hns-*` skills as user customization (INFO, non-failing) |
| AC-HPR-009 | REQ-HPR-014 | `go test ./internal/template/ -run TestSplitHarnessNamespaceNoLeak -v` after extending name set | PASS; test source asserts absence for BOTH `harness-{release-update,github,release}*` AND `hns-{release-update,github,release}*` |
| AC-HPR-010 | REQ-HPR-015, REQ-HPR-016 | `go test ./internal/template/ -run 'NamespaceProtection|TemplateAgentsStructure' -v`; red-team: temporarily plant `internal/template/templates/.claude/skills/hns-probe/SKILL.md`, re-run, confirm FAIL, remove | guard PASS on clean tree AND FAIL on planted `hns-` leak (red-team evidence recorded, plant reverted) |
| AC-HPR-011 | REQ-HPR-021, REQ-HPR-022 | Scoped stale-ref sweep: `grep -rn 'harness-moaiadk-\|harness-github-specialist\|harness-release-specialist\|harness-release-update-specialist\|harness-release-update-run' .claude/ .moai/docs/ internal/ Makefile --exclude-dir=worktrees` (repo root) | 0 matches outside (a) legacy-recognition Go code/comments/tests that intentionally reference legacy names, (b) doctrine migration-history notes; each surviving match individually justified in §E.2 evidence |
| AC-HPR-012 | REQ-HPR-003, REQ-HPR-004 | Non-target baseline-delta, before M1 vs after M4, over the PINNED non-target file set (10 files): `grep -rc 'moai-harness-learner\|harness-spec.yaml\|\.moai/harness/\|handle-harness-observe\|harness-generated' internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl internal/template/templates/.claude/hooks/moai/handle-harness-observe-stop.sh.tmpl internal/template/templates/.claude/hooks/moai/handle-harness-observe-subagent-stop.sh.tmpl internal/template/templates/.claude/hooks/moai/handle-harness-observe-user-prompt-submit.sh.tmpl internal/template/templates/.claude/settings.json.tmpl internal/template/templates/.moai/config/sections/harness.yaml internal/template/templates/.moai/config/sections/interview.yaml internal/template/templates/.moai/config/sections/system.yaml.tmpl internal/template/templates/.moai/config/sections/workflow.yaml internal/template/catalog_loader.go` | per-file counts unchanged before vs after; classification table for all 29 template files recorded in progress.md §E.2 |
| AC-HPR-013 | REQ-HPR-023 | Read `.moai/docs/harness-namespace-doctrine.md` prefix table + `.moai/docs/dev-only-commands-isolation.md` artifact tables | doctrine states `hns-*` canonical + `harness-*`/`my-harness-*` legacy-preserved matrix; dev-only tables list `hns-` artifact names; Builder-emission statement says `hns-` only |
| AC-HPR-014 | REQ-HPR-024, REQ-HPR-026 | `make build && go test ./... ; echo exit=$?` + `go test ./internal/template/ -run 'InternalContentLeak|Neutrality' -v` | build exit 0; full suite PASS; neutrality tests PASS; no SPEC-HNS-PREFIX-RENAME-001 token inside `internal/template/templates/**` (`grep -r 'SPEC-HNS-PREFIX-RENAME' internal/template/templates/` = 0) |
| AC-HPR-015 | REQ-HPR-025 | `git log --stat` for SPEC commits + final report | CLAUDE.local.md absent from every SPEC commit; flagged §21/§24 pointer-update list delivered to user |
| AC-HPR-016 | REQ-HPR-017, REQ-HPR-020 | `go test ./internal/cli/... ./internal/harness/... ./internal/template/... -count=1` and `git log --diff-filter=D --stat -- '*_test.go'` scoped to SPEC commits | all legacy-prefix tests PASS with assertions unmodified; no test file deleted; Runner JS internal strings show `hns-release-update-specialist` |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Legacy user project survives update (backward compatibility)

- **Given** a sandbox user project containing `.claude/skills/harness-acme-verify/SKILL.md`, `.claude/skills/my-harness-old/SKILL.md`, `.claude/workflows/harness-acme-run.js`, and `.claude/agents/harness/harness-acme-core-specialist.md`
- **When** the `moai update` preservation flow runs with the post-rename binary
- **Then** every artifact survives byte-identical, none is backed-up-then-dropped, and the update summary reports them as preserved user-owned paths.

### Scenario 2 — New Builder output is recognized end-to-end

- **Given** a sandbox project containing a Builder-shaped artifact set named per the new contract (`.claude/workflows/hns-acme-run.js`, `.claude/agents/harness/hns-acme-core-specialist.md`, `.claude/skills/hns-acme-verify/SKILL.md`, `.claude/commands/harness/acme.md`, manifest with `runner_workflow: hns-acme-run.js`)
- **When** `moai harness list`, `moai harness doctor`, and the update preservation flow run
- **Then** `list` enumerates the `acme` harness, `doctor` resolves all references with exit 0, and update preserves all artifacts.

### Scenario 3 — Mixed-generation harness lifecycle

- **Given** a harness `acme` whose artifacts straddle generations (`harness-acme-run.js` + `hns-acme-core-specialist.md`)
- **When** `moai harness remove acme` executes
- **Then** artifacts of BOTH prefixes belonging to `acme` are removed atomically, and artifacts of other harnesses (either prefix) are untouched.

## §D.2 Indirect-Verification Rule (reachability, not token presence)

A grep hit on prose is NEVER sufficient for Group B/C ACs. Each recognition claim must be evidenced by an executed test or CLI invocation that traverses the matching code path (`go test -run` output, `moai harness doctor` exit code, E2E sandbox survival). Token-presence greps are admissible only for contract-doc wording (AC-HPR-001), doctrine wording (AC-HPR-013), and absence sweeps (AC-HPR-011, AC-HPR-012, AC-HPR-014's neutrality grep). All greps are case-sensitive (plan.md §B.2).

## §E Edge Cases

1. `hnsx-foo` / `hnsfoo` skill dirs — NOT matched (exact `hns-` HasPrefix); covered in AC-HPR-003.
2. `moai-harness-learner` — template-managed, still overwritten by update (AC-HPR-005).
3. `my-harness-*` third generation — still recognized (AC-HPR-004); no assertion weakened (AC-HPR-016).
4. Name shadowing in v4 matching: `hns-release-update-*` must not be claimed by harness `release` (the existing longest-name-first discipline in `v4lifecycle.go` extends to the `hns-` branch; unit case required in AC-HPR-006).
5. Empty `.claude/agents/harness/` after rename in a foreign project — directory-level preservation unchanged.
6. Case sensitivity: `strings.HasPrefix` is byte-exact — an upper- or mixed-case variant of the prefix is NOT recognized as user-owned. Acceptable by contract: the Builder emits lowercase `hns-` only, and all sweeps are case-sensitive (plan.md §B.2). No case-folding is added.

## §F Quality Gates + Definition of Done

- All 16 ACs PASS with verbatim evidence recorded in progress.md §E.2/§E.3.
- `make build`, `go test ./...` (`-count=1`), `golangci-lint run` on touched files: clean.
- Coverage of touched packages ≥ pre-SPEC baseline (measure both sides; no carry-over numbers).
- Template↔live byte-parity restored for every edited mirror pair.
- Zero deletions/modifications of any user-owned-prefix artifact by any test or build step.
- CLAUDE.local.md flag list delivered; file untouched.
