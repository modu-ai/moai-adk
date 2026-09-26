# acceptance.md — SPEC-AUTONOMY-CONTRACT-001

> Verification layer. Each AC is a binary Given-When-Then scenario with a runnable command and an
> expected output. Requirements live in spec.md §D; this file does not restate them.

## §A. Conventions

- **Test naming (binding on run-phase):** every AC below has exactly one Go test whose name starts with
  `TestAC_CONTRACT_<NNN>` (e.g. `TestAC_CONTRACT_004`, subtests allowed). The AC's command runs it with
  `-count=1`; the expected output is a line starting with `ok` for each listed package and no `FAIL`.
- **CLI fixture:** CLI-level checks run against a throwaway project created by the test with
  `t.TempDir()`, containing `.moai/specs/SPEC-FIXTURE-001/{spec.md,acceptance.md,contract.yaml}` and a
  git repository with `user.name`/`user.email` set. No AC reads or writes the real `.moai/specs/`.
- **JSON checks:** "`reasons` contains X" means the `reasons` array of `verify --json` / `show --json`
  output contains the string X (`design.md` § Verify Reason Codes).

## §B. AC Matrix

| AC | Requirement | Kind | Command |
|---|---|---|---|
| AC-CONTRACT-001 | REQ-CONTRACT-001 | positive | `go test ./internal/contract/... -run TestAC_CONTRACT_001 -count=1` |
| AC-CONTRACT-002 | REQ-CONTRACT-002 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_002 -count=1` |
| AC-CONTRACT-003 | REQ-CONTRACT-003 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_003 -count=1` |
| AC-CONTRACT-004 | REQ-CONTRACT-004 | positive + negative | `go test ./internal/contract/... -run TestAC_CONTRACT_004 -count=1` |
| AC-CONTRACT-005 | REQ-CONTRACT-005 | positive | `go test ./internal/contract/... -run TestAC_CONTRACT_005 -count=1` |
| AC-CONTRACT-006 | REQ-CONTRACT-005 | parity | `go test ./internal/contract/... ./internal/spec/... -run TestAC_CONTRACT_006 -count=1` |
| AC-CONTRACT-007 | REQ-CONTRACT-006 | negative | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_007 -count=1` |
| AC-CONTRACT-008 | REQ-CONTRACT-007 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_008 -count=1` |
| AC-CONTRACT-009 | REQ-CONTRACT-008 | positive | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_009 -count=1` |
| AC-CONTRACT-010 | REQ-CONTRACT-008 | negative (tamper) | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_010 -count=1` |
| AC-CONTRACT-011 | REQ-CONTRACT-008 | negative (tamper) | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_011 -count=1` |
| AC-CONTRACT-012 | REQ-CONTRACT-008 | negative | `go test ./internal/cli/... -run TestAC_CONTRACT_012 -count=1` |
| AC-CONTRACT-013 | REQ-CONTRACT-009 | invariant | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_013 -count=1` |
| AC-CONTRACT-014 | REQ-CONTRACT-010 | negative | `go test ./internal/cli/... -run TestAC_CONTRACT_014 -count=1` |
| AC-CONTRACT-015 | REQ-CONTRACT-010, REQ-CONTRACT-011 | positive | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_015 -count=1` |
| AC-CONTRACT-016 | REQ-CONTRACT-012 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_016 -count=1` |
| AC-CONTRACT-017 | REQ-CONTRACT-013 | positive + negative | `go test ./internal/contract/... ./internal/cli/... -run TestAC_CONTRACT_017 -count=1` |
| AC-CONTRACT-018 | REQ-CONTRACT-014 | positive | `go test ./internal/cli/... -run TestAC_CONTRACT_018 -count=1` |
| AC-CONTRACT-019 | REQ-CONTRACT-015 | positive + fail-safe | `go test ./internal/config/... -run TestAC_CONTRACT_019 -count=1` |
| AC-CONTRACT-020 | REQ-CONTRACT-016 | template | `go test ./internal/template/... -run TestAC_CONTRACT_020 -count=1` |
| AC-CONTRACT-021 | REQ-CONTRACT-017, REQ-CONTRACT-018 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_021 -count=1` |
| AC-CONTRACT-022 | REQ-CONTRACT-019 | neutrality | `go test ./internal/cli/... -run TestAC_CONTRACT_022 -count=1` |
| AC-CONTRACT-023 | REQ-CONTRACT-020 | negative | `go test ./internal/contract/... -run TestAC_CONTRACT_023 -count=1` |
| AC-CONTRACT-024 | REQ-CONTRACT-021 | positive | `go test ./internal/contract/... -run TestAC_CONTRACT_024 -count=1` |

Severity: all 24 are MUST-PASS. AC-CONTRACT-007, 010, 011, 013, 014, and 020 are load-bearing for the
operator decisions (no main/release action, tamper detection, verify safe for hooks, human presence,
guided default).

## §C. Scenarios

### AC-CONTRACT-001 — location, version, spec_id

**Given** a fixture SPEC directory `SPEC-FIXTURE-001` whose `contract.yaml` has `schema_version: 1` and `spec_id: SPEC-FIXTURE-001`,
**When** the contract is loaded by SPEC ID,
**Then** it decodes without error; **and given** the same file with `spec_id: SPEC-OTHER-001`, verify reports `spec_id_mismatch`; **and given** `schema_version: 2`, verify reports `schema_invalid`.

### AC-CONTRACT-002 — required sections

**Given** ten fixture contracts, each missing exactly one of `acceptance`, `invariants`, `ownership`, `approach`, `actions`, `reobserve`, `review`, `budget` (on a signed contract — absent budget is fillable only at sign time), `escalate_on`, `plan_audit`,
**When** each is verified,
**Then** each result is invalid and its `reasons` contains `schema_invalid` or the section-specific code from `design.md` § Field rules.

### AC-CONTRACT-003 — strict decoding

**Given** a otherwise-valid signed contract with an extra top-level field `notes: "x"`, and a second one with an extra nested field `ownership.maybe: []`,
**When** each is verified,
**Then** each exits 1 with `reasons` equal to `["schema_invalid"]`.

### AC-CONTRACT-004 — canonical digest

**Given** a contract C,
**When** its digest is computed for C, for C with `actions` and `ownership.write` reordered, for C with a YAML comment added, and for C re-indented,
**Then** all four digests are identical 64-character lowercase hex strings; **and when** the digest is computed for C with `budget.turns` changed from 60 to 61, **then** it differs; **and when** the `signature` block is added or changed, **then** the digest is unchanged.

### AC-CONTRACT-005 — acceptance hash normalization

**Given** three `acceptance.md` byte variants with the same logical text — LF, CRLF, and UTF-8-BOM + LF,
**When** the acceptance hash is measured for each,
**Then** the three hashes are equal and equal the SHA-256 of the LF variant's bytes; **and given** one changed character, **then** the hash differs.

### AC-CONTRACT-006 — AC counter parity

**Given** every `.moai/specs/*/acceptance.md` in the repository (excluding `_archive`) and the counter program extracted from the manager-docs agent definition's `MOAI-AC-COUNTER` block,
**When** both the extracted counter and the Go counter run on each file,
**Then** they report the same live count for every file and the same ambiguity verdict (the awk exit 3 cases map to the Go ambiguous result); the test fails naming any file where they differ, and asserts the corpus is non-empty.

### AC-CONTRACT-007 — forbidden and unknown actions

**Given** a signed fixture contract whose `actions` is `[commit, push-main]`,
**When** `moai contract verify SPEC-FIXTURE-001 --json` runs in the fixture project,
**Then** the exit code is 1 and `reasons` contains `forbidden_action`; **and given** `actions: [commit, deploy-prod]`, **then** exit 1 and `reasons` contains `unknown_action`; **and** each of `merge-main`, `force-push`, `release-branch`, `release-pr` yields `forbidden_action`.

### AC-CONTRACT-008 — escalation completeness

**Given** a contract whose `escalate_on` lists five of the six triggers, and another listing all six plus `boredom`,
**When** each is verified,
**Then** each is invalid with `reasons` containing `escalate_on_incomplete`.

### AC-CONTRACT-009 — valid signed contract

**Given** a fixture contract signed through the signing core (TTY seam true, confirmation token supplied) with config `second_review: required`, `push_develop: true`,
**When** `moai contract verify SPEC-FIXTURE-001 --json` runs,
**Then** the exit code is 0, `valid` is `true`, `state` is `signed-valid`, and `reasons` is `[]`.

### AC-CONTRACT-010 — tampered acceptance.md

**Given** the signed fixture from AC-CONTRACT-009,
**When** one character of `acceptance.md` is changed (AC count unchanged) and verify runs,
**Then** exit 1, `state` is `signed-invalid`, and `reasons` contains `acceptance_hash_mismatch`; **and when** instead an AC is added to `acceptance.md`, **then** `reasons` contains both `acceptance_hash_mismatch` and `ac_count_mismatch`; **and when** `acceptance.md` is deleted, **then** `reasons` contains `acceptance_missing`.

### AC-CONTRACT-011 — tampered contract body

**Given** the signed fixture from AC-CONTRACT-009,
**When** `ownership.write` gains `internal/**` without re-signing and verify runs,
**Then** exit 1 and `reasons` contains `contract_digest_mismatch`; **and when** `acceptance.sha256` in the contract is edited to match a tampered `acceptance.md`, **then** `reasons` contains `contract_digest_mismatch`.

### AC-CONTRACT-012 — missing signature and missing file

**Given** a fixture contract with no `signature` block,
**When** `moai contract verify SPEC-FIXTURE-001` runs,
**Then** exit 1 and output names `unsigned`; **and given** a SPEC directory with no `contract.yaml`, **then** exit 2 with an error naming the missing path.

### AC-CONTRACT-013 — verify is read-only and subprocess-free

**Given** the signed fixture project and a recorded SHA-256 of every file under it,
**When** verify runs (valid case and each tamper case of AC-CONTRACT-010/011) with the verifier's process-spawn and network seams replaced by failing stubs,
**Then** no stub is called, every file hash is unchanged afterwards, and exit codes are 0 / 1 as expected.

### AC-CONTRACT-014 — non-TTY signing is refused

**Given** an unsigned valid draft and the TTY seam reporting false (and, at CLI level, `moai contract sign SPEC-FIXTURE-001 </dev/null`),
**When** sign runs,
**Then** the exit code is 1, the output states that signing requires an interactive terminal, and `contract.yaml` is byte-identical to before.

### AC-CONTRACT-015 — interactive signing writes the signature record

**Given** an unsigned valid draft without `acceptance.sha256`/`ac_count`/`budget`, git `user.name`/`user.email` set, and the TTY seam true,
**When** sign runs and the confirmation reader supplies a wrong token,
**Then** exit 1 and the file is unchanged; **and when** it supplies the displayed token,
**Then** exit 0, the file now carries `acceptance.sha256` equal to the measured hash, `ac_count` equal to the measured count, `budget` equal to `budget_default`, and a `signature` block with the git name and email, a UTC RFC 3339 `signed_at`, a hex `head_sha` equal to the fixture HEAD, `method: interactive-tty`, and `contract_sha256` equal to a fresh digest; a subsequent verify exits 0.

### AC-CONTRACT-016 — sign-time refusals

**Given** each of these drafts: (a) a recorded `acceptance.sha256` differing from the measured hash, (b) an `acceptance.md` where one AC ID is both `[RETIRED]`-marked and unmarked, (c) `plan_audit.verdict: FAIL`, (d) `actions: [push-main]`, (e) git `user.email` empty, (f) an already signed-valid contract without `--resign`,
**When** sign runs with the TTY seam true and the correct token,
**Then** each exits 1 with a message naming the refusal cause and each file is byte-identical to before; **and given** (f) with `--resign`, **then** exit 0 and `signature.supersedes` equals the previous `contract_sha256`.

### AC-CONTRACT-017 — batch signing

**Given** config `batch_sign: true` and three valid unsigned drafts,
**When** `sign` is invoked with all three IDs and the batch token,
**Then** the confirmation reader is called exactly once, all three verify exit 0, and all three signatures carry the same non-empty `batch_id`; **and given** one of the three is invalid, **then** exit 1 and all three files are byte-identical to before; **and given** config `batch_sign: false`, **then** an invocation with two IDs exits 1 without prompting.

### AC-CONTRACT-018 — show output

**Given** the signed fixture,
**When** `moai contract show SPEC-FIXTURE-001 --json` runs,
**Then** exit 0 and the output parses as one JSON object containing the keys `spec_id`, `schema_version`, `state`, `valid`, `reasons`, `contract_sha256`, `recorded_contract_sha256`, `acceptance`, `actions`, `push_requires_window`, `second_review`, `mode`, `budget`, `signature`; **and given** the unsigned draft, `state` is `unsigned`; **and** plain `show` (no `--json`) prints each schema section heading.

### AC-CONTRACT-019 — configuration defaults and fail-safe

**Given** a `workflow.yaml` with no `autonomy` section,
**When** the configuration is loaded,
**Then** mode is `guided`, `batch_sign` false, `second_review` `required`, `push_develop` false, and budget `60/40/2`; **and given** `mode: yolo`, **then** mode is `guided` and a warning naming `workflow.autonomy.mode` is emitted; **and given** `mode: contract`, `second_review: advisory`, **then** those values are returned.

### AC-CONTRACT-020 — template default

**Given** the embedded template file `.moai/config/sections/workflow.yaml`,
**When** it is decoded,
**Then** `workflow.autonomy.mode` is `guided` and `workflow.autonomy.contract.second_review` is `required`; and a scan of the template file finds no match for `SPEC-[A-Z]`, `\bt[0-9]{3,}\b`, or a `20[0-9]{2}-[0-9]{2}-[0-9]{2}` date. The existing template neutrality check (`.github/workflows/template-neutrality-check.yaml` path) also passes on the changed file.

### AC-CONTRACT-021 — second-review and push-develop coupling

**Given** config `second_review: required` and a signed contract with `review.second_model: none`,
**When** verify runs,
**Then** `reasons` contains `second_review_missing`; **and given** `second_review: advisory`, **then** that code is absent; **and given** config `push_develop: false` and `actions` containing `push-develop`, **then** `reasons` contains `push_develop_disabled`; **and given** `push_develop: true`, **then** `show --json` reports `push_requires_window: true`.

### AC-CONTRACT-022 — guided-mode neutrality

**Given** config `mode: guided`,
**When** sign completes successfully,
**Then** its output contains the notice that the signature does not replace Implementation Kickoff Approval in guided mode; **and** with `mode: contract` the notice is absent; **and** `git diff --stat develop -- .claude/ internal/template/templates/.claude/` on the run-phase branch lists no skill, rule, or agent file (no gate document changed in A1).

### AC-CONTRACT-023 — ownership and invariant well-formedness

**Given** signed fixture contracts with, respectively, `ownership.write: []`, a write glob `/etc/**`, a write glob `../x/**`, the glob `internal/a/**` in both `write` and `never`, a `write` list not covering `.moai/specs/SPEC-FIXTURE-001/**`, and an invariant `constitution:CONST-NOPE-*`,
**When** each is verified against a fixture constitution registry,
**Then** the first five report `ownership_invalid` and the last reports `invariant_unresolved`; **and** `constitution:CONST-V3R2-*` against a registry containing `CONST-V3R2-001` reports neither.

### AC-CONTRACT-024 — mission-contract projection

**Given** a signed-valid fixture contract,
**When** it is projected and sealed with the existing mission sealing function,
**Then** sealing succeeds, the sealed hash equals `signature.mission_contract_sha256`, the projection's allowed actions are the mapped mission actions of `design.md` § Action Vocabulary (`worktree` omitted), and its prohibited actions contain `main_merge`, `force_push`, `release_branch`, `release_pr`.

## §D. Edge Cases

- Empty `acceptance.md` (0 ACs): sign refuses (`ac_count` must be >= 1).
- `contract.yaml` symlinked outside the SPEC directory: load refuses (path containment), exit 2.
- Two IDs in a batch naming the same SPEC twice: deduplicated before validation; one signature.

## §E. Quality Gates and Definition of Done

- Every AC command in §B passes in the worktree.
- `go vet` and `golangci-lint run` clean for touched packages; coverage of `internal/contract` >= 85%.
- `make build` succeeds after the template change; `moai spec lint` reports no error for this SPEC.
- Sync phase documents the three commands and the configuration keys.
