# SPEC-GLM-JEV-KEY-001 — progress

- SPEC: SPEC-GLM-JEV-KEY-001 (card t1613)
- status: completed
- phase: plan (manager-spec, 2026-10-09)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (2026-10-09 — plan-audit final verdict PASS, overall 1.00 ≥ Tier M 0.80, blocking 0, audited_sha c1512b919; verdict: .moai/reports/t1613/plan-audit-verdict.md. Iteration history: FAIL 0.69 → repairs bb0bfbfcd/71040aec7/2bf9ac29a → FAIL 0.75 → repair 0d2036275 → FAIL 0.94 (D9 alone) → repair c1512b919 → PASS via hunk-scoped reread. Codex cross-opinion ran bounded twice; receipt-less raw-exec path recorded fail-closed in the verdict.)

## §F Phase 4 Mode Selection

- Input parameters: tier M · scope ~3 production files + 1-2 test files · domains 1 (internal/cli, reusing internal/glmcred + internal/jevcred) · file language mix Go · concurrency benefit LOW (coding-heavy) · agent-teams prereqs not requested.
- Mode evaluation: direct — no (semantic multi-file change); serial — SELECTED; fanout — no (coding-heavy, single domain, sequential milestones M1→M3); sweep — no (semantic new-code work, ~5 files, not ≥30-file mechanical).
- Decision: serial
- Justification: coding-heavy CLI implementation per Anthropic's coding-task parallelism caveat — sequential single-spawn (manager-develop) over M1→M2→M3 with RED-first tests; the storage layer already exists (glmcred/jevcred reuse) so the work is one-domain surface wiring; fan-out would split context without parallelizable independence.
- Kickoff gate: met in the default autonomous form — plan-audit verdict PASS (1.00 ≥ 0.80, blocking 0) on audited_sha c1512b919 = current HEAD, artifact-hash unchanged since the verdict, no open blockers. Decision record: .moai/reports/t1613/progress.md (2026-10-09).

## §E.2 Run-phase Evidence

Run phase: manager-develop, TDD (RED-GREEN-REFACTOR), branch WT-10-09-class, base f7606c7bc (develop). Commits (per-milestone, no push — lane lands via factory_complete):

| Milestone | SHA | Subject |
|---|---|---|
| M1 (RED + status transition) | f7454e31d | feat(SPEC-GLM-JEV-KEY-001): M1 RED tests for glm --key flag and moai jev --key |
| M2 (GREEN) | d7bdb3cce | feat(SPEC-GLM-JEV-KEY-001): M2 GREEN — glm --key scan, moai jev command |
| M3 (edge test + lint fix) | 99e5d271a | feat(SPEC-GLM-JEV-KEY-001): M3 verification — --key= edge test, lint fix |
| M3 (harness hardening) | 5306e4cb2 | fix(SPEC-GLM-JEV-KEY-001): M3 — execRoot resets command-tree outputs |
| M2 gate repair | 5de28887e | fix(SPEC-GLM-JEV-KEY-001): M2 gate repair — redact refusal, sweep whole region, refuse flag-shaped values |
| Card-review repair | 0cd03b4a0 | fix(SPEC-GLM-JEV-KEY-001): card-review repair — mask positional tokens, refuse flag-shaped jev values |
| Card-review r2 repair | 7c43c4034 | fix(SPEC-GLM-JEV-KEY-001): card-review r2 — allowlist-based redaction |
| Evidence refresh | (this commit) | feat(SPEC-GLM-JEV-KEY-001): run-phase evidence refresh — gate repair rows |

(spec.md `status:` draft → in-progress on M1; spec.md frontmatter is the only SPEC-body surface touched; updated: unchanged — same calendar day.)

### E8 — RED evidence (captured BEFORE each GREEN, this tree, base 2f73e3514 lineage)

- `go test ./internal/cli/ -run '<new family>' -count=1 -v` → **10 FAIL** (save/refuse/empty/newline/jev-command cases, each failing for its stated reason — e.g. `unknown command "jev" for "moai"`, `GLM API key not found`, masked-confirmation missing) + 2 PASS (vacuous-green passthrough by design, M2 characterization setup-routing); the 3 help tests separately: `--- FAIL: TestRootHelpListsJevCommand / TestGlmHelpDocumentsKeyFlag / TestJevHelpDocumentsKeyFlag`.
- AC-GJK-016 mutant-RED (plan §F): a temporary whole-args scan stub in runGLM (removed before GREEN) made `TestGlmKeyAfterDashDashPassthrough` FAIL with `jev_key_test.go:363: no save confirmation may appear for a post--- token, got: "GLM API key stored (test****7890)\n"` — exactly the mutant the RED cell requires.
- `--key=<value>` edge (acceptance §B, added in M3): with the scan's `--key=` branch temporarily removed, `TestGlmKeyEqualsFormSaves` FAIL observed; branch restored → green. (Test-first derived, not test-after.)
- M2 gate repair RED (turn-end gate defects, repaired in 5de28887e — verbatim RED observed before the fix): `TestGlmKeyConflictErrorMasksValue` — `refusal must not disclose the second key value, got: --key cannot be combined with other arguments (found "--key=sk-secret-9999")` (the P1 leak); `TestGlmKeyLeadingArgsRefused` — `a launch-flag mixed invocation must be refused` (P2-leading: `-p work --key K` stored); `TestGlmKeyExecFlagAsValueRefused` — `a flag-shaped token must not be stored as the key` (P2-execflag: `--key -f` stored).
- Card-review repair RED (round-1 findings, repaired in 0cd03b4a0 — verbatim RED observed before the fix): `TestGlmKeyPositionalDuplicateMasked` — `refusal must not disclose the key value, got: ... (found "sk-secret-9999")` (positional duplicate passed redactArg verbatim); `TestJevKeyOptionTokenValueRefused` — `an option-shaped value must be refused, not stored` (pflag consumed `--help` as the credential string).
- Card-review r2 repair RED (round-2 finding, repaired in 7c43c4034 — verbatim RED observed before the fix): `TestGlmKeyEqualsInsideValueMasked` — `... (found "sk-secret=****")` (the '=' heuristic preserved most of the key value); `TestGlmKeyDashLeadingDuplicateMasked` — `... (found "-foo")` (the dash heuristic passed a key value as a flag name). redactArg is now an allowlist — only confirmed glm flag names pass; every other token is `****`.

### E1 — AC matrix (all observed this run phase; HEAD 5306e4cb2 unless noted)

| AC | Status | Verification command | Observed output (verbatim, abbreviated to the deciding lines) |
|---|---|---|---|
| AC-GJK-001 | PASS | `go test ./internal/cli/ -run 'TestRootHelpListsJevCommand$'` | `--- PASS` (family run: `ok ... internal/cli 1.063s`) |
| AC-GJK-002 | PASS | `go test ./internal/cli/ -run 'TestGlmHelpDocumentsKeyFlag$'` | `--- PASS` |
| AC-GJK-003 | PASS | `go test ./internal/cli/ -run 'TestJevHelpDocumentsKeyFlag$'` | `--- PASS` |
| AC-GJK-004 | PASS | `TestGlmKeyFlagSavesKeySameStorage` + binary smoke `MOAI_HOME=/tmp/t1613-smoke2 /tmp/moai-t1613 glm --key sk-smoke-1234567890` | `GLM API key stored (sk-s****7890)`; file `GLM_API_KEY="sk-smoke-1234567890"`; `ls -l` → `-rw-------@` (0600); full key absent from stdout/stderr |
| AC-GJK-005 | PASS | `TestJevKeyFlagSavesCredential` + smoke `moai jev --key tsk-smoke-9876543210` | `Jev credential stored (…3210)`; `TYPESAFE_API_KEY="tsk-smoke-9876543210"`; `-rw-------@`; disclosure = final four chars only |
| AC-GJK-006 | PARTIAL — see Gaps | mixed glm-family subset `go test ./internal/cli/ -run 'Test(GLM|Glm)' -timeout 150s` | `ok github.com/modu-ai/moai-adk/internal/cli 18.390s` (pre-existing glm family + new tests together, after the output-reset fix). **Package-scope** run NOT observed green locally (structural, below); CI owns the repository-wide verdict — PENDING at report time |
| AC-GJK-007 | PASS | `TestKeyFormsShareStorageLastWriterWins` + smoke (setup 1111 → `--key` 2222 → file has 2222) | `GLM API key stored (sk-f****2222)` / `GLM_API_KEY="sk-flag-wins-2222"` |
| AC-GJK-008 | PASS | `TestGlmSetupRoutingUnchanged` + smoke `moai glm setup sk-legacy-1111` | `GLM API key stored (sk-l****1111)` / `GLM_API_KEY="sk-legacy-1111"` (setup path untouched) |
| AC-GJK-009 | PASS | `TestGlmKeyFlagRefusesExtraArgs` + `TestGlmKeyLeadingArgsRefused` + `TestGlmKeyPositionalDuplicateMasked` (card-review repair) + smoke `moai glm --key K status` | exit 1, `--Key cannot be combined with other arguments (found "status")...`; leading `-p work --key K` → `found "-p"`; positional duplicate → `found "****"` (masked, card-review binary smoke); no file written |
| AC-GJK-010 | PASS | `TestJevBareInvocationPrintsHelpExitZero` + smoke bare `moai jev` | jev help printed, `exit=0`, `.env.typesafe` absent |
| AC-GJK-011 | PASS | `TestGlmKeyFlagMissingValueErrors` + `TestGlmKeyEmptyValueErrors` + `TestJevKeyEmptyValueErrors` + `TestGlmKeyExecFlagAsValueRefused` + `TestJevKeyOptionTokenValueRefused` (card-review repair) | `--- PASS` ×5 (`--key requires a value` for a flag-shaped token; `empty API key` / `empty Jev credential`; jev option-shaped value refused, stored file byte-for-byte unchanged) |
| AC-GJK-012 | PASS (package level) | `go test ./internal/glmcred/ ./internal/jevcred/` (Save's explicit Chmod tests) + smoke `ls -l` | `ok ... glmcred 0.317s` / `ok ... jevcred 0.376s`; smoke files `-rw-------@` |
| AC-GJK-013 | PASS | premises: `git rev-parse --verify f7606c7bc` → resolves; `git diff --stat f7606c7bc..HEAD -- internal/cli/` → non-empty (607 insertions at M3; re-verified after the gate-repair commit); then `git diff f7606c7bc..HEAD -- internal/cli/ ':(exclude)**/*_test.go' \| grep '^+' \| grep "TYPESAFE_API_KEY\|GLM_API_KEY"` | `GREP_EXIT=1` — **0 rows** (re-observed at HEAD 5de28887e) |
| AC-GJK-014 | PASS | `TestJevKeyNewlineValueRefusesAndPreserves` | `--- PASS` (refused, file byte-for-byte identical) |
| AC-GJK-015 | PASS | `TestGlmKeyFlagNewlineValueRefusesAndPreserves` | `--- PASS` (refused, file byte-for-byte identical) |
| AC-GJK-016 | PASS (incl. mutant-RED) | `TestGlmKeyAfterDashDashPassthrough` + smoke `moai glm -- --key x` | `--- PASS`; smoke: no save confirmation, `.env.glm` unchanged (85 bytes, prior setup value), launch path's own error terminates (exit 1) |

### §C-gate results (this tree, HEAD 5306e4cb2)

- Full new family (incl. the three gate-repair tests, 19 names): `go test ./internal/cli/ -run 'Test(…)$' -count=1` → `ok ... internal/cli 0.769s`; earlier 16-name run `-count=2` → `ok` (stability, after the pflag flag-reset and output-reset fixes). After the card-review repairs (0cd03b4a0): focused selector `go test ./internal/cli/ -run 'Test(GLM|Glm|Jev|Key|Root)'` → `ok ... 25.374s`; binary smoke of all four repair cases observed (positional duplicate masked to `found "****"`, jev option token refused with an empty credential directory, both earlier gate repairs unchanged).
- `go test ./internal/cli/ -run 'Test(GLM|Glm)' -count=1 -timeout 150s` → `ok ... 18.390s`.
- `go test ./internal/glmcred/ ./internal/jevcred/ -count=1` → both `ok` (80.8% / 85.0% coverage — packages untouched this card).
- `go vet ./internal/cli/` → clean (exit 0). `gofmt -l` on touched files → empty.
- `golangci-lint run --timeout=2m` → `0 issues.` (one ST1005 introduced mid-M3 was fixed in 99e5d271a).
- `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → `FINAL_BUILDS_OK`.
- New-code coverage (selector-scoped profile): `handleGLMKeyFlag` 93.8%, `scanGLMKeyValue` 87.5%, `jev.init` 100.0%, `runJev` 81.2%. The internal/cli *package* percentage (5.7%) is the selector-executed fraction only — it is NOT a package verdict and is not claimed as one.
- PRESERVE: `git diff f7606c7bc..HEAD -- internal/glmcred internal/jevcred` → **0 lines** (storage SSOT bodies untouched); `runGLMSetup` body untouched; manual routing order untouched (scan inserted after the switch); envkeys.go untouched; no `.env.jev`; no new config sections; t1612 renderer surfaces untouched.

### Gaps (explicitly NOT observed)

- **Package-scope `go test ./internal/cli/`** (AC-GJK-006's literal front line and the §C preservation gate): one background run **timed out at 601.119s with ZERO individual `--- FAIL` rows** — the known card-worktree structural red (t1542 contention class: `thisRepoRoot` codex family / child env / suite-level contention hangs the suite rather than failing tests). Not retried per lane-local discipline; NOT recorded as a pass. Compensating local evidence: the mixed glm-family subset (`Test(GLM|Glm)`, 18.390s ok) sweeps the pre-existing glm family together with the new tests, and CI owns the repository-wide verdict — PENDING at report time.
- `moai doctor` Jev probe (acceptance §D.4) — optional manual probe, not run; the AC evidence is the on-disk file.

### Residual-risk

- The 601s suite timeout's true cause is un-diagnosed by design (no retry); if it were caused by this card rather than the t1542 class, CI would surface it — this is the residual the CI verdict covers.
- `glm --key status` (a literal value equal to a subcommand name) routes to the status subcommand via cobra's Find before the scan can see it — outside every AC, stores nothing, and fails safe; recorded here for t1612's guidance author.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-09
run_commit_sha: pending-backfill-run
run_status: complete-with-gaps (package-scope suite structural red local; CI owns repository-wide verdict)
ac_pass_count: 15
ac_fail_count: 0
ac_not_observed_locally: 1 (AC-GJK-006 package scope — see §E.2 Gaps)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (factory lane — no push from this lane; landing via factory_complete)
l44_post_push_fetch: n/a (same)
new_warnings_or_lints_introduced: 0 (one ST1005 introduced and fixed within M3)
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 3
m1_to_mN_commit_strategy: per-milestone commits (M1 RED f7454e31d / M2 GREEN d7bdb3cce / M3 99e5d271a + 5306e4cb2); RED evidence per §E.2 E8; no push, no --no-verify


## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-10-09
sync_commit_sha: acc8d071a
sync_status: complete
changelog_entry_position: CHANGELOG.md [Unreleased] § Added (top entry)
b12_self_test_a: pass — pre-emission `grep -c 'SPEC-GLM-JEV-KEY-001' CHANGELOG.md` = 0 (duplicate guard)
b12_self_test_b: pass — AC counter on acceptance.md: live=26 identifier strings = 16 canonical criteria (AC-GJK-001..016, §D.1) + 10 abbreviated §D.3 cross-reference spellings; excluded=0, ambiguous=0; CHANGELOG cites 16
b12_self_test_c: pass — entry-claimed paths ls-verified (internal/cli/jev.go, internal/cli/glm.go, internal/glmcred/, internal/jevcred/)
frontmatter_status_transitions:
  spec_md: in-progress → implemented → completed (single sync commit, merged close)
  plan_acceptance: `updated:` refreshed (frontmatter only; bodies untouched)
canary_compliance_check: n/a (this SPEC defines no forward-looking policy tested by its own sync tests)
docs_site: updated — docs-site/content/{ko,en,ja,zh}/cli-reference/launchers.md (ko canonical → derived; glm `--key` form + `moai jev --key` mention; no new headings, section-count parity unchanged)
mx_tag_validation: pass — existing @MX:NOTE [AUTO] tags on jev.go (jevCmd) and glm.go (glmCmd) well-formed; no new tags required (new functions tested, package-private)

