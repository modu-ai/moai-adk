# Command text: citations that were removed

SPEC-20261120-cli-text-citations. The text the commands print (help, previews, hints, error and note strings) no longer quotes moai-adk's own history. This document keeps what was removed, under the path of the file it came from, for the developer who wants the origin of a sentence. Nothing here changes what a command tells the user; the wording that stays is the whole message.

`tests/cli_text_citations.rs` checks that every file listed in its `CLEANED` list has a section here and carries no citation, and that the files of its `PENDING` list are the only ones that still do.

## `src/claim.rs`

- Removed `(decision D2)` after "answering `met` and the line it prints" (the handoff hint that names `moai claim-timeout-check`).
- Removed `(decision D1, Issue #1975 comment 6840449133)` from the note about a renewed claim (a session's own expired claim is renewed with the next epoch and no handoff line).

## `src/claim_verify.rs`

- Removed `, decision D5 HOLD` from "not guarded (fail-open)" (a lane branch without a leading Issue number is not guarded).

## `src/gate_schedule.rs`

- Removed `(#2906 Q1)` from the Windows reason ("Windows Task Scheduler entries are out of scope").

## `src/handoff.rs`

- Removed `(REQ-06B-07, #2099)` from the note "some effects cannot be judged from Issue comments; pass --remote-probe to judge them".

## `src/heavy_slot.rs`

- Removed `(#2841)` from the description of the per-lane build folder.
- Removed `card C-82` from "(internal/build-folder, card C-82)" in the heavy-slot hint; the hint still names `internal/build-folder`.

## `src/help.rs`

Removed from the Korean command descriptions (each was a parenthesis or a clause naming an Issue; the sentence around it is unchanged):

- `project-detect` `(#2496)`, `language-setup` `(#2293)`, `index-build` `(C6 #2182)`, `update` `(#2522)`.
- `go-state-plan` `(#2561)`, `go-state-apply` `(#2561, #2696)` and the words "#2561 2단계" (now "파일 이동 단계"), `go-state-revert-plan` `(#2561 2단계)`, `go-state-revert` `(#2561, #2696)`.
- `home-state-plan` `(#2723)`, `home-state-apply` `(#2723)`, `home-state-revert` `(#2723)`.
- `home-retire-plan`, `home-retire-apply`, `home-retire-revert` `(#2871)`; `home-install-plan` `(#2871 U3a)`, `home-install-apply` `(#2871 U3c)`, `home-install-revert` `(#2871 U3c)`.
- `ledger-status` `(#3068)`, `card` `(#3070)`, `ledger-register-plan` and `ledger-register-apply` `(#3068)`, `card-import-plan` and `card-import-apply` `(#3071)`, `card-summary-plan` and `card-summary-apply` `(#3071)`, `ledger-recover-plan` and `ledger-recover-apply` `(#3068)`.
- `worktree-move-plan` `(#2723, #2835)`, `worktree-move-apply` `(#2723)`.
- `install-inspect` `(04-A #1998)`, `worktree-report` `(WT-1, #2380)`, `worktree-install-plan` `(#3010 P6a)`, `optimize-plan` `(#2724)`.
- `key-set` `(#2393)`, `key-status` `(#2393)`, `key-remove` `(#2393)`, `key-import-plan` and `key-import-apply` `(#2393 S3)`.
- `--confirm-token` descriptions: `cleanup.confirm_token` `(#2872)`, `home.confirm_token` `(#2871)`, `switch.confirm_token` `(C-33 S3)`.
- Argument descriptions: the per-item split `(#2815)` and the worktree root `#2835`.
- The sample ids of the `bundle` examples changed from `C-119 C-120` to `C-1 C-2` (sample ids, not citations).

## `src/home_state.rs`

- Removed `(#2811)` from the note on work folders and from the error "is user work and is never moved or copied".
- "the apply is the next step of #2723" became "the apply is a later step"; "the selective apply is the next card after #2815" became "the selective apply is a later step".

## `src/launcher_init.rs`

- The permission reference recorded in the apply receipts was `launcher-auto-init #2481 user decision 6810818827`; it is now `launcher-auto-init: the user chose first-launch setup`. The reference only has to name the reason for the writes; the Issue comment that recorded the user's choice is no longer in the text.

## `src/ledger/open.rs`

- Removed `(Issue #3068 Q1)` from "the ledger waits for the HOME move".

## `src/project_hosts.rs`

- Removed `(#2936 S2, default A)` from the reason that MoAI writes no `.omp/config.yml`.

## `src/recovery.rs`

- Removed `(#2091)` from "a receipt written before version_before existed" and `, #2268` from "(unmanaged root at apply time)".

## `src/shim_check.rs`

- Removed `(#3021)` from the shim note ("Nothing was changed.").

## `src/worktree.rs`

- Removed `(#2930 R-3)` from "is inside the Cargo build folder".

## `src/worktree_gc.rs`

- Removed `(#2498)` from the `--recent-minutes` error and from the apply hint; removed `(#2389 comment 6774899518 item 2)` from "on this machine the leader runs it only after the user confirms".

## `src/worktree_include.rs`

- Removed `, #2670` from "(not changed here)".

## `src/worktree_report.rs`

- Removed `(WT-4)` and the label `Q4`: the reason now reads "removal candidates need the in-use checks of worktree.gc.plan and a decision by the user".
