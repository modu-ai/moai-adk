---
id: SPEC-20261119-diff-go
title: "differential tests against the Go build for behaviour the Rust build keeps (C-104)"
status: draft
revision: 2
created: 2026-11-19
---

# differential tests against the Go build for behaviour the Rust build keeps (C-104)

## Intent

Card C-104 (verify J: "the old system is the oracle"). The Go build `moai.old` still exists on the machine of the person who moved from it, and the Rust build kept a few of its behaviours on purpose. A test can run both on the same input and compare, so a change in the Rust build that breaks a kept behaviour is seen, and a difference that the Rust build chose is written down with its reason instead of being found again later. The card's research-first rule holds: nothing here ports a Go behaviour; only what the Rust build already keeps is compared.
Measured on this machine against `moai.old` v3.2.1-rc.6 before any test was written: the status line text differs in layout on purpose, while the reading of the payload (model name from a duplicate key, a plain string, a key in another letter case, a wrong type, broken JSON) agrees; Go keeps the last model name in a cache file below its HOME (`HOME/.moai/state/last-model.txt`, measured in the tier 1 audit), so every case needs a fresh folder and HOME; a Go-generation hook call is answered by Go with an allow decision and ignored by the Rust build. The audit also measured two shared behaviours that the first head left out: both builds answer a deny decision with the same `hookEventName` and `permissionDecision` for `rm -rf /` (the reason text differs), and both read `.moai/config/sections/statusline.yaml` and obey the model switch in it.

## Non-goals

- Reading of `project.yaml`: the Go build keeps its project settings in `.moai/config/sections/project.yaml` (its warning names that folder), the Rust build in `.moai/project.yaml`, which Go never reads; no shared behaviour for that file, so no case exists. The settings file the Rust build does read from Go's folder, `statusline.yaml`, is compared (REQ-6); `llm.yaml` and `system.yaml` (read by `src/statusline/state.rs`) have no case yet and are a follow-up card.
- The text layout of the status line and the hook JSON of the Go build: they differ on purpose and are on the approved list, not compared for equality.
- Running the Go build in CI, the network, and the real HOME.
- Changing any Rust source.

## Requirements

- REQ-1: When the Go build is present, the status line tests shall run the Go build and the Rust build on the same payload, each in its own temporary HOME and folder, and compare which model names each shows. · failure: none — a difference fails the test with both outputs.
- REQ-2: When the Go build is absent, each test shall print the reason it did not compare and pass. · failure: none — the skip is the answer.
- REQ-3: When a behaviour differs on purpose, the approved-differences file shall hold the case name, what each build does and the reason, with a test that checks the difference is still present. · failure: none — a vanished difference fails the test.
- REQ-4: When a differential test starts a binary, the test shall give it a cleared environment, a temporary HOME and a temporary folder, and no network address. · failure: none — a leaked HOME is caught by the isolation scan.
- REQ-6: When the folder holds `.moai/config/sections/statusline.yaml` with `statusline.segments.model` set to `false` or `true`, or no such file, the status line tests shall compare both builds and expect the model name hidden for `false` and shown otherwise. · failure: none — a difference fails the test with both outputs.
- REQ-7: When a dangerous command reaches the Go hook call and the Rust tool-preflight hook, the hook test shall compare the `hookEventName` and `permissionDecision` of both answers and list the different reason text and the different silent allow on the approved list. · failure: none — a vanished difference fails the test.
- REQ-5: When the Go build answers a case, the test shall check that answer against the expectation written in the test, so two builds that both show nothing do not pass. · failure: none — a changed oracle fails the sanity assertion.
- REQ-8: When the approved-differences list is read, every entry shall name a case that a test of `tests/diff_go.rs` consults through the approved check with a non-empty reason, so a made-up or stale entry fails the suite (leader note 2687 N7). · failure: none — the fixture test fails naming the unconsulted entry.

## Acceptance

- AC-1 (REQ-1, REQ-5): seven payloads (object name, later duplicate key, plain string, key case, wrong type, broken JSON, empty object) show the same model names in both builds, and the Go answer equals the written expectation.
  verify: `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go the_status_line_reads_the_model_of_a_payload_like_the_go_build`
- AC-2 (REQ-2): with no Go build the tests print the reason and pass, and the lookup is tested for a missing path, no value, the variable and the home folder.
  verify: `MOAI_GO_BIN=/nonexistent/moai.old cargo test --locked --offline --test diff_go`; `cargo test --locked --offline --test diff_go an_absent_go_build_skips_with_a_reason`
- AC-3 (REQ-3): the layout of the status line and a Go-generation hook call are on the approved list with a reason each, and their tests fail when the difference disappears.
  verify: `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go the_layout_of_the_status_line_differs_on_purpose`; `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go a_go_hook_call_is_ignored_by_the_rust_build_on_purpose`
- AC-4 (REQ-4): the helper runs every binary with a cleared environment in a fresh temporary HOME and folder.
  verify: `cargo test --locked --offline --test test_isolation_rules`
- AC-6 (REQ-6): with the switch `false` neither build shows the model, with `true` or without the file both show it, and the Go answer equals the written expectation.
  verify: `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go the_status_line_obeys_the_model_switch_of_statusline_yaml_like_the_go_build`
- AC-7 (REQ-7): `rm -rf /` is denied by both with the same event name and decision and different reason text, and `ls` is answered by Go with an allow decision while the Rust build prints nothing; both differences are on the approved list.
  verify: `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go a_dangerous_command_is_denied_with_the_same_decision_as_the_go_build`; `MOAI_GO_BIN=$HOME/go/bin/moai.old cargo test --locked --offline --test diff_go a_harmless_command_is_allowed_silently_by_the_rust_build_on_purpose`
- AC-8 (REQ-8): every approved-differences entry names a consulted case with a reason; a fake entry added to the fixture fails the fixture test naming it.
  verify: `cargo test --locked --offline --test diff_go every_approved_difference_entry_is_consulted_by_a_test`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): the style and lint checks pass.
  verify: `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-diff-go/spec.md`, `tests/diff_go.rs`, `tests/fixtures/diff_go/approved-differences.json`
- read-only: `src/statusline/godecode.rs`, `tests/common/mod.rs`
- forbidden: every file under `src/`, the real HOME, the network, a live tool session

## Plan

1. Measure both builds by hand in a temporary HOME (done; see Intent).
2. Tests and the approved-differences file, run with the Go build and with a path that is absent.
3. Test-side mutants, neighbour checks, Result.

## Risks

- The Go build is not on CI, so the comparison runs only on a machine that has it → each test says why it skipped; the approved list and the lookup are tested without it.
- The Go build writes a `.moai` folder into its current folder and a model cache below its HOME → every case runs in its own temporary folder and HOME.
- A Go version other than v3.2.1-rc.6 may answer differently → the sanity assertion names the Go answer, so a changed oracle fails loudly instead of passing quietly.
- The mutants were applied to the test file only, because the status line source belongs to other cards → a change in `godecode.rs` is not mutation tested by this slice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Compare which model names are shown, not the whole text | the layouts differ on purpose (approved list) | compare more segments |
| D-2 | One fresh HOME and folder per case | Go caches the last model in `HOME/.moai/state/last-model.txt` (measured by the audit: a second run in the same HOME showed the earlier name) | share a home |
| D-3 | The Go build comes from `MOAI_GO_BIN`, else `$HOME/go/bin/moai.old`, and is only executed | the card says present-or-skip | require the variable |
| D-4 | `project.yaml` reading is a Non-goal; `statusline.yaml` reading is compared | research-first: `project.yaml` lives in different files, `statusline.yaml` is read by both (audit B2) | add a case for `llm.yaml` or `system.yaml` |

## Result

- AC-1: met — with `MOAI_GO_BIN=/Users/goos/go/bin/moai.old` (v3.2.1-rc.6) `the_status_line_reads_the_model_of_a_payload_like_the_go_build` passes: seven payloads, Rust equals Go, and Go equals the written expectation
- AC-2: met — with `MOAI_GO_BIN=/nonexistent/moai.old` all four tests pass and print `diff_go: skipped, the Go build /nonexistent/moai.old is absent (set MOAI_GO_BIN)`; `an_absent_go_build_skips_with_a_reason` tests the lookup for a missing path, no value, the variable and the home folder
- AC-3: met — `the_layout_of_the_status_line_differs_on_purpose` and `a_go_hook_call_is_ignored_by_the_rust_build_on_purpose` pass; both cases are in `tests/fixtures/diff_go/approved-differences.json` with a reason
- AC-4: met — `test_isolation_rules` 3 passed with the new file (every binary starts with `clean_env`, a temporary HOME and a temporary folder); also `git_env_scrub` 3, `ac_command_lint` 5, `cargo clippy --test diff_go -- -D warnings` 0, `python3 -m unittest tests.repo_scans_test tests.ci_scope_test` OK
- AC-6: met — `the_status_line_obeys_the_model_switch_of_statusline_yaml_like_the_go_build` passes (three cases: `model: false` hides the name, `model: true` and no file show it; Rust equals Go and Go equals the written expectation)
- AC-7: met — `a_dangerous_command_is_denied_with_the_same_decision_as_the_go_build` and `a_harmless_command_is_allowed_silently_by_the_rust_build_on_purpose` pass (same `hookEventName` and `permissionDecision` deny; different reason text; Go allow answer against Rust silence; both on the approved list)
- AC-5: met — `cargo fmt --check` 0
- AC-8 (r2 후속, 리더 노트 2687 N7): met — RED: fixture 에 가짜 항목(`fake_entry_probe`)을 넣자 `every_approved_difference_entry_is_consulted_by_a_test` 가
  "holds \"fake_entry_probe\", which no test consults" 로 실패(관찰 2026-11-19T05:4xZ) → probe 제거 후 GREEN: `--test diff_go` 8 통과(2.90초, Go 빌드 없이
  스킵 사유 출력 포함), fmt 0, `cargo clippy --test diff_go -- -D warnings` 0. CONSULTED_CASES 4종이 fixture 의 4 항목과 정확히 대응해요.

RED: before `tests/diff_go.rs` existed there was no comparison at all; the first run of the hook case failed on a wrong expectation of mine (`{}` against the measured allow decision, which depends on the folder having no MoAI settings), and the test was corrected to the measured answer. Repair after the tier 1 audit: the two new hook tests were run before their approved-list entries existed and FAILED with `not on the approved list` (2 failed, 5 passed); the `statusline.yaml` test passed at once because it describes a behaviour both builds already share, so it is checked by the mutants below.
Mutation of the first head: test side 6 mutants, 5 killed and 1 survived (the audit counted the same: D2 wrong expectation, D3 duplicate-key payload changed, D4 layout case compared Go with Go, D5 hook decision, D6 lookup ignores the variable killed; D1, Rust compared with itself, survives because a mutant that empties the Rust check cannot be seen by the test that holds it). Source side, in a scratch worktree removed afterwards: a string model emptied in `src/statusline/godecode.rs` made `the_status_line_reads_the_model_of_a_payload_like_the_go_build` FAIL; clearing only the display name, and display name plus name, survived because the status line falls back to the id, which Go also does. The audit added two source mutants that were killed (fold-name fallback removed, duplicate-key order reversed).
Mutation of the repair, 5 mutants, 4 killed and 1 equivalent: S1 Rust ignores the model switch (`src/statusline/render.rs`) KILLED by the switch test; S2 Rust Claude deny answers `ask` (`src/hook_engine.rs`) KILLED by the deny test; T1 the `model: false` case expects the name shown KILLED; T2 the `statusline.yaml` file is not written into the folder KILLED; T3 the deny test no longer compares `permissionDecision` SURVIVES on the unchanged source, which is the expected result because only a changed answer could differ (S2 shows the answer is caught). Both scratch worktrees were removed and the tree was clean after each mutant.
Not run: the full local `cargo test`, Linux, Windows, any Go build other than v3.2.1-rc.6, `llm.yaml` and `system.yaml` reading (follow-up card), a `sandbox-exec` network deny of my own (the audit ran one).
