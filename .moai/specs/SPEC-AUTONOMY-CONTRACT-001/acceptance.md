# acceptance.md — SPEC-AUTONOMY-CONTRACT-001

> Verification layer. Each AC is a binary Given-When-Then scenario with a runnable command and an
> expected output. Requirements live in spec.md §D; this file does not restate them.

## §A. Conventions

- **Test naming (binding on run-phase):** every AC below has exactly one top-level Go test named
  `TestAC_CONTRACT_<NNN>` (subtests allowed), placed in the **owning package** named in the §B matrix.
- **Pass check (applies to every AC; defined once here).** From the worktree root, with `<NNN>` and
  `<pkg>` taken from the AC's matrix row:

  ```
  go test -v -count=1 -run '^TestAC_CONTRACT_<NNN>' <pkg> > ac.out 2>&1; echo "exit=$?"
  grep -E '^--- PASS: TestAC_CONTRACT_<NNN> ' ac.out
  grep -E '^--- FAIL|\[no tests to run\]' ac.out; echo "bad=$?"
  ```

  The AC passes only when all three hold: `exit=0`; the second command prints exactly one
  `--- PASS: TestAC_CONTRACT_<NNN> (…)` line; the third prints no line and `bad=1`. A run that prints
  `[no tests to run]`, or no `--- PASS:` line for the named test, is a **FAIL** — the test was not
  written or not found. (Plan-phase measurement: `go test -v -count=1 -run '^TestAC_CONTRACT_999'
  ./internal/atomicfile/` exits 0 and prints `ok … [no tests to run]` with zero `--- PASS:` lines, so a
  check on `ok` alone would pass a missing test; the existing test
  `TestReplace_OntoAbsentDestination` prints `--- PASS: TestReplace_OntoAbsentDestination (0.00s)`.)
- **Shell checks:** where an AC also lists shell commands, they run from the worktree root outside any
  Go test, and their expected output is stated in the AC. Both the Go pass check and the shell checks
  must hold.
- **CLI fixture:** CLI-level checks run against a throwaway project created by the test with
  `t.TempDir()`, containing `.moai/specs/SPEC-FIXTURE-001/{spec.md,acceptance.md,contract.yaml}`, a
  plan-audit report file, and a git repository with `user.name`/`user.email` set. No AC reads or writes
  the real `.moai/specs/`, except AC-CONTRACT-006, which reads (never writes) the real
  `.moai/specs/*/acceptance.md` corpus.
- **Environment:** tests that exercise signing clear the agent-environment markers
  (`design.md` § Agent-Environment Markers) through the signer's environment seam, except where an AC
  sets one deliberately.
- **JSON checks:** "`reasons` contains X" means the `reasons` array of `verify --json` / `show --json`
  output contains the string X (`design.md` § Verify Reason Codes). "Refuses with X" means `sign` exits 1
  and prints the code X from `design.md` § Sign Refusal Codes, and every fixture file is byte-identical
  to before.

## §B. AC Matrix

| AC | Requirement | Kind | Owning package `<pkg>` | Extra shell check |
|---|---|---|---|---|
| AC-CONTRACT-001 | REQ-CONTRACT-001 | positive + negative | `./internal/contract/` | — |
| AC-CONTRACT-002 | REQ-CONTRACT-002 | negative | `./internal/contract/` | — |
| AC-CONTRACT-003 | REQ-CONTRACT-003 | negative | `./internal/contract/` | — |
| AC-CONTRACT-004 | REQ-CONTRACT-004 | positive + negative | `./internal/contract/` | — |
| AC-CONTRACT-005 | REQ-CONTRACT-005 | positive | `./internal/contract/` | — |
| AC-CONTRACT-006 | REQ-CONTRACT-005 | parity | `./internal/spec/` | — |
| AC-CONTRACT-007 | REQ-CONTRACT-006 | negative | `./internal/cli/` | — |
| AC-CONTRACT-008 | REQ-CONTRACT-007 | negative | `./internal/contract/` | — |
| AC-CONTRACT-009 | REQ-CONTRACT-008 | positive | `./internal/cli/` | — |
| AC-CONTRACT-010 | REQ-CONTRACT-008 | negative (tamper) | `./internal/cli/` | — |
| AC-CONTRACT-011 | REQ-CONTRACT-008 | negative (tamper) | `./internal/cli/` | — |
| AC-CONTRACT-012 | REQ-CONTRACT-008, REQ-CONTRACT-009 | negative | `./internal/cli/` | — |
| AC-CONTRACT-013 | REQ-CONTRACT-009 | invariant | `./internal/contract/` | `go list` checks |
| AC-CONTRACT-014 | REQ-CONTRACT-010, REQ-CONTRACT-024 | negative | `./internal/cli/` | — |
| AC-CONTRACT-015 | REQ-CONTRACT-010, REQ-CONTRACT-011, REQ-CONTRACT-024 | positive | `./internal/cli/` | — |
| AC-CONTRACT-016 | REQ-CONTRACT-012, REQ-CONTRACT-023 | negative | `./internal/contract/sign/` | — |
| AC-CONTRACT-017 | REQ-CONTRACT-013 | positive + negative | `./internal/cli/` | — |
| AC-CONTRACT-018 | REQ-CONTRACT-014, REQ-CONTRACT-025 | positive | `./internal/cli/` | — |
| AC-CONTRACT-019 | REQ-CONTRACT-015 | positive + fail-safe | `./internal/config/` | — |
| AC-CONTRACT-020 | REQ-CONTRACT-016 | template | `./internal/template/` | template scan, local grep |
| AC-CONTRACT-021 | REQ-CONTRACT-017, REQ-CONTRACT-018 | negative | `./internal/contract/` | — |
| AC-CONTRACT-022 | REQ-CONTRACT-019 | neutrality | `./internal/cli/` | `git diff` |
| AC-CONTRACT-023 | REQ-CONTRACT-020 | negative + positive | `./internal/contract/` | — |
| AC-CONTRACT-024 | REQ-CONTRACT-022 | positive + negative | `./internal/cli/` | — |
| AC-CONTRACT-025 | REQ-CONTRACT-021 | negative | `./internal/cli/` | — |

Severity: all 25 are MUST-PASS. AC-CONTRACT-007, 010, 011, 013, 014, 016, 020, 024, and 025 are
load-bearing for the operator decisions (no main/release action, tamper detection and recovery, verify
safe for hooks, human-presence checks, receipt gating, template neutrality).

## §C. Scenarios

### AC-CONTRACT-001 — location, version, spec_id

**Given** a fixture SPEC directory `SPEC-FIXTURE-001` whose `contract.yaml` has `schema_version: 1` and `spec_id: SPEC-FIXTURE-001`,
**When** the contract is loaded by SPEC ID,
**Then** it decodes without error; **and given** the same file with `spec_id: SPEC-OTHER-001`, verify reports `spec_id_mismatch`; **and given** `schema_version: 2`, verify reports `schema_invalid`.

### AC-CONTRACT-002 — required sections

**Given** ten fixture signed contracts, each missing exactly one of `acceptance`, `invariants`, `ownership`, `approach`, `actions`, `reobserve`, `review`, `budget`, `escalate_on`, `plan_audit`, and one fixture omitting only the optional `ownership.scratch`,
**When** each is verified,
**Then** each of the ten is invalid and its `reasons` contains `schema_invalid` or the section-specific code from `design.md` § Field rules, and the scratch-less fixture reports no reason caused by the omission.

### AC-CONTRACT-003 — strict decoding

**Given** an otherwise-valid signed contract with an extra top-level field `notes: "x"`, and a second one with an extra nested field `ownership.maybe: []`,
**When** each is verified,
**Then** each exits 1 with `reasons` equal to `["schema_invalid"]`.

### AC-CONTRACT-004 — canonical digest

**Given** a contract C,
**When** its digest is computed for C, for C with `actions`, `ownership.write`, and `ownership.scratch` reordered, for C with a YAML comment added, and for C re-indented,
**Then** all four digests are identical 64-character lowercase hex strings; **and when** the digest is computed for C with `budget.turns` changed from 60 to 61, **then** it differs; **and when** the `signature` block is added or changed, **then** the digest is unchanged.

### AC-CONTRACT-005 — acceptance hash normalization

**Given** three `acceptance.md` byte variants with the same logical text — LF, CRLF, and UTF-8-BOM + LF,
**When** the acceptance hash and AC count are measured for each,
**Then** the three hashes are equal and equal the SHA-256 of the LF variant's bytes, and the three counts are equal; **and given** one changed character, **then** the hash differs.

### AC-CONTRACT-006 — AC counter parity (corpus and per-branch fixtures)

**Given** the counter program extracted from the manager-docs agent definition's `MOAI-AC-COUNTER` block, and two inputs: (a) every `.moai/specs/*/acceptance.md` (excluding `_archive`), and (b) synthetic fixtures — one with a `moai-ac-prefix` declaration, one with a `[RETIRED]`-marked ID, one with a `[REF]`-marked ID, one where a single ID appears both marked and unmarked (ambiguity), one with CRLF line endings, one with a UTF-8 BOM,
**When** both the extracted counter and the Go counter run on each input's normalized bytes,
**Then** they report the same live count and the same ambiguity verdict for every input (the awk exit 3 cases map to the Go ambiguous result); the test fails naming any input where they differ; it asserts the corpus is non-empty; as positive controls it asserts the ambiguity fixture is reported ambiguous by **both** implementations and the prefix fixture yields a count that differs from counting the same file under the default `AC` prefix; **and** it runs the extracted counter on the CRLF fixture's **raw** bytes and logs the raw count beside the normalized count (a `t.Logf` line containing `raw-crlf-count=` and `normalized-count=`), which is the measurement behind spec.md §C.5's "equal only for LF, BOM-free files" statement.

### AC-CONTRACT-007 — empty, forbidden, and unknown actions

**Given** a signed fixture contract whose `actions` is `[commit, push-main]`,
**When** `moai contract verify SPEC-FIXTURE-001 --json` runs in the fixture project,
**Then** the exit code is 1 and `reasons` contains `forbidden_action`; **and given** `actions: [commit, deploy-prod]`, **then** exit 1 and `reasons` contains `unknown_action`; **and given** `actions: []`, **then** exit 1 and `reasons` contains `actions_empty`; **and** each of `merge-main`, `force-push`, `release-branch`, `release-pr` yields `forbidden_action`.

### AC-CONTRACT-008 — escalation completeness

**Given** a contract whose `escalate_on` lists five of the six triggers, and another listing all six plus `boredom`,
**When** each is verified,
**Then** each is invalid with `reasons` containing `escalate_on_incomplete`.

### AC-CONTRACT-009 — valid signed contract

**Given** a fixture contract signed on the human path (TTY seam true, markers cleared, confirmation token supplied) with config `second_review: required`, `push_develop: true`,
**When** `moai contract verify SPEC-FIXTURE-001 --json` runs,
**Then** the exit code is 0, `valid` is `true`, `state` is `signed-valid`, and `reasons` is `[]`.

### AC-CONTRACT-010 — tampered acceptance.md

**Given** the signed fixture from AC-CONTRACT-009,
**When** one character of `acceptance.md` is changed (AC count unchanged) and verify runs,
**Then** exit 1, `state` is `signed-invalid`, and `reasons` contains `acceptance_hash_mismatch`; **and when** instead an AC is added to `acceptance.md`, **then** `reasons` contains both `acceptance_hash_mismatch` and `ac_count_mismatch`; **and when** `acceptance.md` is deleted, **then** `reasons` contains `acceptance_missing`.

### AC-CONTRACT-011 — tampered contract body and signature block

**Given** the signed fixture from AC-CONTRACT-009,
**When** `ownership.write` gains `internal/**` without re-signing and verify runs,
**Then** exit 1 and `reasons` contains `contract_digest_mismatch`; **and when** `acceptance.sha256` in the contract is edited to match a tampered `acceptance.md`, **then** `reasons` contains `contract_digest_mismatch`.
**And given** the untampered human-signed fixture and the untampered `llm` receipt-signed fixture from AC-CONTRACT-015, **when** verify runs, **then** each is `signed-valid` with `reasons` `[]`. **And when** exactly one signature field is edited **without** recomputing the seal — each of `signed_at`, `head_sha`, `operator.email`, `batch_id`, `supersedes`, `acceptance_sha256`, `method`, `signer_kind`, and, on the receipt fixture, `receipt.provenance` (`file` → `moai-store`) and `receipt.path` — **then** each verify exits 1 with `reasons` containing `signature_seal_mismatch`. **And when** the edit is made **with** the seal recomputed: (a) on the receipt fixture, only `method` changed to `interactive-tty` — **then** `reasons` contains `signature_inconsistent`; (b) on the receipt fixture, `receipt.provenance: moai-store` — **then** `reasons` contains `signature_inconsistent`; (c) on the human fixture, `signer_kind: llm` — **then** `reasons` contains `signature_inconsistent`; (d) on the human fixture, `signature.acceptance_sha256` set to a different 64-hex value — **then** `reasons` contains `signature_acceptance_mismatch`; (e) residual-risk control — on the receipt fixture, `method: interactive-tty`, `signer_kind: human`, the `receipt` block deleted, and the seal recomputed — **then** verify reports `signed-valid`: a complete keyless re-seal is not detectable by verify, as spec.md §C.6 and §H state.

### AC-CONTRACT-012 — missing signature, missing file, path escape

**Given** a fixture contract with no `signature` block,
**When** `moai contract verify SPEC-FIXTURE-001` runs,
**Then** exit 1 and output names `unsigned`; **and given** a SPEC directory with no `contract.yaml`, **then** exit 2 with an error naming the missing path; **and given** `contract.yaml` is a symlink resolving outside the SPEC directory, **then** exit 2 with an error naming the escape.

### AC-CONTRACT-013 — verify is read-only and has no process or network dependency

**Given** the signed fixture project and a recorded SHA-256 of every file under it,
**When** verify runs (valid case and each tamper case of AC-CONTRACT-010/011),
**Then** every file hash is unchanged afterwards and exit codes are 0 / 1 as expected.
**And when** these shell commands run from the worktree root (`deps.txt` is a scratch file):

```
go list ./internal/contract; echo "list_exit=$?"
go list -deps ./internal/contract > deps.txt; echo "deps_exit=$?"
test -s deps.txt; echo "nonempty_exit=$?"
! grep -xE 'os/exec|net' deps.txt; echo "noexecnet_exit=$?"
! grep -E 'moai-adk/internal/(spec|config|constitution|hook)$' deps.txt; echo "noimport_exit=$?"
```

**Then** the first prints `github.com/modu-ai/moai-adk/internal/contract` and `list_exit=0` (positive control: the package exists), and every other line prints `…_exit=0`. Any non-zero value fails the AC. Because the dependency checks read a file written by a `go list` whose own exit status is checked (`deps_exit`), a `go list` failure cannot pass as "no forbidden dependency". (Plan-phase measurement of these semantics: `go list ./internal/contract` prints `stat …/internal/contract: directory not found` and `list_exit=1` — the AC fails, as it must, before the package exists; on `./internal/atomicfile` the same steps give `list_exit=0`, `deps_exit=0`, `noexecnet_exit=0`; on `./internal/config` the `net` line is printed and `noexecnet_exit=1`.)

### AC-CONTRACT-014 — signing-path gates (non-TTY human path; receipt path mode gate)

**Given** an unsigned valid draft and, on the human path, the TTY seam reporting false (and, at CLI level with markers cleared, `moai contract sign SPEC-FIXTURE-001 </dev/null`),
**When** sign runs,
**Then** it refuses with `not_tty` and the output states that signing requires an interactive terminal;
**and given** config `mode: guided`, `kickoff.decider: llm`, and a receipt that would otherwise validate, **when** `sign --signer llm --receipt .moai/specs/SPEC-FIXTURE-001/kickoff-receipt.json` runs, **then** it refuses with `mode_not_contract`;
**and given** the same with `mode: contract` and the TTY seam false, **then** signing succeeds without reading a confirmation (the receipt path does not require a terminal).

### AC-CONTRACT-015 — signature record on both paths

**Given** an unsigned valid draft without `acceptance.sha256`/`ac_count`/`budget`, git `user.name`/`user.email` set, markers cleared, and the TTY seam true,
**When** sign runs on the human path and the confirmation reader supplies a wrong token,
**Then** it refuses with `confirmation_mismatch`; **and when** it supplies the displayed token,
**Then** exit 0, the file now carries `acceptance.sha256` equal to the measured hash, `ac_count` equal to the measured count, `budget` equal to `budget_default`, and a `signature` block with `signer_kind: human`, the git name and email under `operator`, a UTC RFC 3339 `signed_at`, a hex `head_sha` equal to the fixture HEAD, `method: interactive-tty`, `contract_sha256` equal to a fresh digest, and `acceptance_sha256` equal to the measured hash; a subsequent verify exits 0.
**And given** a second unsigned draft, config `mode: contract`, `kickoff.decider: llm`, and a receipt at the fixed path whose inputs equal `show --json`'s `signable_contract_sha256`, the acceptance hash, and the plan-audit report hash, with one `llm` decision `approve` (in A1 only an `llm` receipt can be accepted; any Jev decision routes to a human, spec.md §C.8), **when** `sign --signer llm --receipt <that path>` runs, **then** exit 0, the confirmation reader is never called, the signature has `signer_kind: llm`, `method: receipt`, `receipt.sha256` equal to the receipt file's SHA-256, and `receipt.provenance: file`, and verify exits 0; **and when** one byte of the receipt file is then changed, **then** verify exits 1 with `reasons` containing `receipt_mismatch`.

### AC-CONTRACT-016 — sign-time and receipt refusals

**Given** each of these unsigned drafts on the human path (TTY seam true, markers cleared, correct token): (a) a recorded `acceptance.sha256` differing from the measured hash, (b) an `acceptance.md` where one AC ID is both `[RETIRED]`-marked and unmarked, (c) `plan_audit.verdict: FAIL`, (d) `actions: [push-main]`, (e) git `user.email` empty, (f) an `acceptance.md` with zero AC IDs; and (g) an already signed-valid contract signed again without `--resign`,
**When** sign runs,
**Then** they refuse with, respectively, `draft_acceptance_stale`, `ac_count_ambiguous`, `plan_audit_not_passing`, `verify_failed`, `git_identity_missing`, `ac_count_zero`, and `already_signed`.
**And given** config `mode: contract`, `workflow.jev.enabled: true`, `jev_min_confidence: 0.50`, and receipt-path invocations whose receipts have: (h) a missing `reason_refs`, (i) `signer: llm` while `kickoff.decider` is `llm+jev`, (j) an `inputs.acceptance_sha256` that differs from the file, (k) `llm+jev` with `llm` approve and jev approve at confidence 0.30, (l) `llm+jev` with `llm` approve and jev reject at confidence 0.71 under `on_disagree: human`, (m) `llm+jev` with `llm` approve and jev approve at confidence 0.71 under `on_disagree: reject`, (n) an `llm` decision `escalate`, (o) a receipt path outside `.moai/specs/SPEC-FIXTURE-001/kickoff-receipt.json`, (p) `kickoff.decider: llm` and one `llm` decision `reject` under `on_disagree: reject`; and (q) `kickoff.decider: jev` with `workflow.jev.enabled: false`,
**When** sign runs for each,
**Then** they refuse with, respectively, `receipt_invalid`, `receipt_signer_mismatch`, `receipt_input_mismatch`, `receipt_requires_human`, `receipt_requires_human`, `receipt_requires_human`, `receipt_requires_human`, `receipt_invalid`, `receipt_rejected`, and `kickoff_decider_jev_disabled`. Cases (k), (l), and (m) all route to a human because any Jev decision is unmeasured in A1 (`design.md` § Agreement Rule, rule 0), independent of its confidence, its answer, and `on_disagree`; (p) exercises the `on_disagree: reject` branch without a Jev decision.

### AC-CONTRACT-017 — batch signing

**Given** config `batch_sign: true` and three valid unsigned drafts,
**When** human-path `sign` is invoked with all three IDs (one of them repeated) and the batch token,
**Then** the confirmation reader is called exactly once, exactly three signatures are written, all three verify exit 0, and all three carry the same non-empty `batch_id`; **and given** one of the three is invalid, **then** exit 1 and all three files are byte-identical to before; **and given** config `batch_sign: false`, **then** an invocation with two distinct IDs refuses with `batch_disabled` without prompting; **and given** `batch_sign: true` and `--signer llm` with two distinct IDs, **then** it refuses with `batch_non_human`.

### AC-CONTRACT-018 — show output and derived sets

**Given** the signed fixture, whose contract has `ownership.never: [internal/x/**]`, `ownership.scratch: [.moai/state/verify/**]`, and the invariant `frozen-files`, with the registry Frozen file list `[.claude/rules/moai/core/moai-constitution.md]` supplied,
**When** `moai contract show SPEC-FIXTURE-001 --json` runs,
**Then** exit 0 and the output parses as one JSON object containing the keys `spec_id`, `schema_version`, `state`, `valid`, `reasons`, `contract_sha256`, `recorded_contract_sha256`, `signable_contract_sha256`, `acceptance`, `actions`, `push_requires_lease`, `terminal`, `effective_never`, `scratch`, `frozen_files`, `second_review`, `mode`, `budget`, `signature`; `effective_never` contains `internal/x/**`, `.moai/specs/SPEC-FIXTURE-001/contract.yaml`, and `.moai/specs/SPEC-FIXTURE-001/acceptance.md`; `scratch` equals `[".moai/state/verify/**"]`; `frozen_files` equals, as a set, `["**/CLAUDE.md", "**/CLAUDE.local.md", ".claude/rules/moai/core/moai-constitution.md", "internal/x/**"]` (every element a glob; the bare strings `CLAUDE.md` and `CLAUDE.local.md` are absent); `terminal` is `false`; **and given** the unsigned draft, `state` is `unsigned` and `effective_never` omits the two SPEC files; **and given** the fixture's `spec.md` frontmatter `status: completed`, and separately `status: "archived"`, **then** `terminal` is `true`, and with `status: draft` it is `false`; **and** plain `show` (no `--json`) prints each schema section heading.

### AC-CONTRACT-019 — configuration defaults and fail-safe

**Given** a `workflow.yaml` with no `autonomy` section,
**When** the configuration is loaded,
**Then** mode is `guided`, `batch_sign` false, `second_review` `required`, `push_develop` false, budget `60/40/2`, `kickoff.decider` `human`, `kickoff.jev_min_confidence` `0.50`, and `kickoff.on_disagree` `human`; **and given** each of `mode: yolo`, `second_review: sometimes`, `decider: oracle`, `on_disagree: shrug`, `jev_min_confidence: 1.5`, **then** the value falls back to `guided`, `required`, `human`, `human`, `0.50` respectively and a warning naming that key is emitted; **and given** `mode: contract`, `second_review: advisory`, `decider: llm+jev`, `on_disagree: reject`, `jev_min_confidence: 0.6` with `workflow.jev.enabled: true`, **then** those values are returned; **and given** `decider: jev`, and separately `decider: llm+jev`, each with `workflow.jev.enabled: false`, **then** the reader reports a configuration error whose text names both `workflow.autonomy.kickoff.decider` and `workflow.jev.enabled`, and the returned decider is not `human` (no fallback).

### AC-CONTRACT-020 — template default, neutrality, and local values

**Given** the embedded template file `.moai/config/sections/workflow.yaml`,
**When** it is decoded,
**Then** `workflow.autonomy.mode` is `guided`, `workflow.autonomy.contract.second_review` is `required`, and `workflow.autonomy.kickoff.decider` is `human`.
**And when** this shell scan runs from the worktree root over the whole template file:

```
grep -nE 'SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}|(^|[^A-Za-z0-9])t[0-9]{3,}([^0-9]|$)|20[0-9]{2}-[0-9]{2}-[0-9]{2}|A-Q[0-9]|[Tt]his repository' \
  internal/template/templates/.moai/config/sections/workflow.yaml; echo "exit=$?"
```

**Then** it prints no matching line and `exit=1`. (Plan-phase measurement at HEAD `ca1d5dc43` lineage: the current template prints `exit=1`; a copy with `    # (default required; A-Q3)` and `    # this repository local config sets true` appended prints both planted lines (263, 264) and `exit=0`; an earlier copy with `# see SPEC-AUTONOMY-CONTRACT-001` appended also matched. The multi-segment `SPEC(-…)+` form is required — the single-segment `SPEC-[A-Z][A-Z0-9]+-[0-9]{3}` misses IDs such as `SPEC-AUTONOMY-CONTRACT-001`.)
**And when** `grep -nE '^ *(batch_sign|push_develop): true$' .moai/config/sections/workflow.yaml` runs, **then** it prints exactly two lines (this repository's local values).

### AC-CONTRACT-021 — second-review and push-develop coupling

**Given** config `second_review: required` and a signed contract with `review.second_model: none`,
**When** verify runs,
**Then** `reasons` contains `second_review_missing`; **and given** `second_review: advisory`, **then** that code is absent; **and given** config `push_develop: false` and `actions` containing `push-develop`, **then** `reasons` contains `push_develop_disabled`; **and given** `push_develop: true`, **then** the derived `push_requires_lease` is `true` (a `moai slot` lease on resource `push-develop`, `design.md` § Push Lease).

### AC-CONTRACT-022 — Kickoff-neutral notice in both modes and both paths; no gate document changed

**Given** config `mode: guided`, and separately `mode: contract`,
**When** a human-path sign completes in each, and a receipt-path sign completes under `mode: contract`,
**Then** every one of the three outputs contains the notice that the signature does not yet replace Implementation Kickoff Approval.
**And when** this shell command runs from the worktree root on the run-phase branch:

```
git diff --stat develop...HEAD -- .claude/skills .claude/rules .claude/agents internal/template/templates/.claude
```

**Then** it prints nothing (merge-base diff: no skill, rule, or agent file changed by this SPEC's branch).

### AC-CONTRACT-023 — ownership and invariant well-formedness

**Given** signed fixture contracts with, respectively, `ownership.write: []`, a write glob `/etc/**`, a write glob `../x/**`, the glob `internal/a/**` in both `write` and `never`, a `write` list `[internal/**, .moai/specs/*.md]` (does not match `.moai/specs/SPEC-FIXTURE-001/contract.yaml`), a scratch glob `../tmp/**`, the glob `tmp/**` in both `scratch` and `never`, and an invariant `constitution:CONST-NOPE-*`,
**When** each is verified with the supplied registry IDs `[CONST-V3R2-001]`,
**Then** the first seven report `ownership_invalid` and the last reports `invariant_unresolved`; **and** contracts whose `write` holds `.moai/**`, `.moai/specs/**`, or `.moai/specs/SPEC-FIXTURE-001/**` report no `ownership_invalid`; **and** `constitution:CONST-V3R2-*` reports no `invariant_unresolved`.

### AC-CONTRACT-024 — re-sign after an acceptance change

**Given** the signed fixture from AC-CONTRACT-009 with digest D0 and `plan_audit.verdict: PASS`, whose `acceptance.md` then gains one AC (verify now reports `acceptance_hash_mismatch` and `ac_count_mismatch`),
**When** `sign SPEC-FIXTURE-001` runs without `--resign`,
**Then** it refuses with `already_signed`;
**When** `sign --resign SPEC-FIXTURE-001` runs on the human path with the TTY seam true, markers cleared, and the displayed token,
**Then** the summary shows the recorded and newly measured acceptance hash prefix and AC count as `old → new` and shows the plan-audit verdict marked as carried, exit is 0, the contract's `acceptance.sha256` and `ac_count` equal the new measurement, `signature.supersedes` equals D0, and verify exits 0; **and when** the same `--resign` runs with the TTY seam false, **then** it refuses with `not_tty`; **and given** an unsigned draft, **when** `sign --resign` runs, **then** it refuses with `not_signed`.

### AC-CONTRACT-025 — agent-environment refusal (human path only)

**Given** an unsigned valid draft, the TTY seam true, and the correct token,
**When** human-path sign runs with `CLAUDECODE=1` in the signer's environment, and separately with `CLAUDE_CODE_SESSION_ID=abc`,
**Then** each refuses with `agent_marker` and the output names the detected variable; **and when** both are unset or empty, **then** signing proceeds (exit 0); **and given** `CLAUDECODE=1` with config `mode: contract`, `kickoff.decider: llm`, and a validating receipt, **when** `sign --signer llm --receipt <fixed path>` runs, **then** it signs (exit 0) — markers gate only the human path.

## §D. Quality Gates and Definition of Done

- Every AC's §A pass check holds in the worktree; every listed shell check prints its stated output.
- `go vet` and `golangci-lint run` clean for touched packages; coverage of `internal/contract` >= 85%.
- `make build` succeeds after the template change; `moai spec lint` reports no error for this SPEC.
- Sync phase documents the three commands and the configuration keys.
