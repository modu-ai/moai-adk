---
id: SPEC-20261124-audit-prep
title: "An audit prepares its slot from the head with one incremental build and finishes clean, addressed by card"
status: draft
revision: 1
created: 2026-11-24
---

# An audit prepares its slot from the head with one incremental build and finishes clean, addressed by card

## Intent

Card C-364 and the leader decision of 2026-11-23T22:22Z (user delegation 2026-11-23T13:57Z). The user pointed at the cost again: each audit built the tree from nothing in an empty build folder, although the lane had already built the same head. The sources are the redesign `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3D and the review `audit-landing-process-review-2026-11-22.md` A1, A2, A4 and "감사가 레인 빌드를 이어받는 방법".

Today the leader does this by hand: make an audit worktree at the head, clone the lane's `target/` with `cp -c -R`, run the changed targets incrementally from that one `target/`, and never build the merged tree. This SPEC turns that hand rule into two commands that a leader or an auditor types with the card name:

- `moai audit-prep <card> --head <40hex>` takes an audit slot for the card, switches it detached to the head, places the lane's (or the seed's) build in it, and runs one incremental build of the named test targets.
- `moai audit-done <card>` finishes the card's slot: it releases it only when it is clean, and it never discards or forces.

What already exists (read at tip `b21063a45d`): `moai audit-slot prep|done` landed as `178fde8beb` (SPEC-20261123-audit-slot, audit PASS 7045). It already does the slot switch, the lane clone with `cp -c -R -p`, the touch of tracked sources (so a mutant build in the lane cannot pass as the head), one holder per slot, the link and own-worktree refusals, and the clean-up with `--discard`. It does not run the build, it is addressed by slot and holder instead of by card, and its tests do not show that a built test executable reads the audit folder's paths instead of the lane's. This SPEC adds exactly those, on top of it; it does not copy its safety code.

Readers: the leader who dispatches an audit, the auditor who runs it, the lane whose READY worktree is the build source.

## Non-goals

- A per-card audit worktree `.claude/worktrees/audit-<card>` that is removed at the end. The user-approved redesign §3A fixes `audit-1`..`audit-3` and reuses them (D-1). If the leader picks that shape, this SPEC is revised before any code.
- Fixing the compile-time `env!` paths. Today 163 files under `tests/` read `env!("CARGO_MANIFEST_DIR")` or `env!("CARGO_BIN_EXE_*")` (measured with `grep -lE` in this worktree at `df2cffd593`; the card says 135). They stay as they are. This slice only shows that a rebuilt test executable reads the audit slot, not the lane (REQ-9), and that nothing is shared that would not be.
- Sharing a nextest archive between the lane and the audit. It needs those paths read at run time first.
- Copying evidence logs at the verdict, and removing a slot. Slots are fixed.
- Running the audit's tests or mutants. `audit-prep` only builds; the auditor runs.
- Running `check affected` again for the head. The auditor cites the lane's receipt.
- Any host change outside this repository: no `~/.claude`, no `~/.moai` data.

## Requirements

- REQ-1: When `audit-prep` is given a card and a 40-hex head, the audit-prep command shall take a free audit slot with the card as holder and leave it detached at that head.
- REQ-2 (REQ-1): If the card already holds an audit slot, then the audit-prep command shall prepare that same slot again.
- REQ-3 (REQ-1): If every audit slot has a live holder other than the card, then the audit-prep command shall refuse, name every holder and run no build.
- REQ-4 (REQ-1): If the slot cannot be prepared, then the audit-prep command shall refuse with the reason, change nothing more and run no build.
- REQ-5: When the slot is prepared, the audit-prep command shall run one incremental `cargo test --no-run` of the test targets named with `--test`, or of the library when none is named, inside one heavy slot.
- REQ-6 (REQ-5): If more than two test targets are named, then the audit-prep command shall refuse before it changes anything.
- REQ-7: When the build ends, the audit-prep command shall report the units Cargo compiled, the units it found fresh and the exit code of the build. · failure: none — a build that fails is REQ-8, and the report is written in both cases.
- REQ-8 (REQ-5): If the build fails, then the audit-prep command shall exit non-zero, keep the slot held by the card and name the last lines of Cargo's output.
- REQ-9: When the slot's build came from a lane, the audit-prep command shall build every named target from the slot's own sources, so no compile-time path in a rebuilt test executable names the lane's folder. · failure: none — a compile error is REQ-8; a lane path found in a rebuilt executable fails AC-4.
- REQ-10: When `audit-done` is given a card, the audit-done command shall finish the slot that card holds like `audit-slot done` without discard: refuse and name the files when the slot holds a change, release it when it is clean, and keep its `target/`.
- REQ-11 (REQ-10): If the card holds no audit slot, then the audit-done command shall refuse and say so.
- REQ-12: The audit-done command shall accept no option that discards a change or forces a release.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): in a temporary repository with a lane at H, `audit-prep C-1 --head H --lane L` leaves one slot detached at H, held by `C-1`, and answers the slot name; a second call for `C-1` answers the same slot; three other holders make a call for `C-9` refuse naming all three; an unusable lane or a dirty slot is refused and no build runs. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-2 (REQ-5, REQ-6, REQ-7): a temporary crate with a dependency outside the repository, built in the lane and prepared into a slot, builds with one `audit-prep` call that names one test target; the answer lists the workspace units as compiled once each and every dependency as fresh; a third `--test` is refused before any change. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-3 (REQ-8): with a test target that does not compile, `audit-prep` exits non-zero, the slot is still held by the card, and the answer carries the compiler's last lines. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-4 (REQ-9): the lane builds a test executable that prints its `CARGO_MANIFEST_DIR`; after `audit-prep` it is rebuilt in the slot and prints the slot's folder, and the rebuilt executable holds no byte sequence of the lane's folder. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-5 (REQ-10, REQ-11, REQ-12): after a mutant changed a tracked file in the slot, `audit-done C-1` refuses, names the file and keeps the hold; once reverted it releases the slot, keeps `target/` and leaves an empty `status --porcelain`; a card without a slot is refused; `--discard` and `--force` are not accepted. · verify: `cargo nextest run --locked --offline --test audit_prep_cli`
- AC-6 (REQ-5): on a prepared slot a one-line mutant in the library recompiles only the workspace units, none of the dependencies; the real repository is not built by any test (no whole `cargo test`). · verify: `cargo nextest run --locked --offline --test audit_prep_cli`

## Scope

- write:
  - `.moai/specs/SPEC-20261124-audit-prep/spec.md`
  - `src/audit_prep.rs` (new)
  - `src/ops/audit_prep.rs` (new), one operation `audit.prep` with the two actions
  - `src/lib.rs` (one `pub mod` line)
  - `tests/audit_prep_cli.rs` (new)
  - `docs/design-review/appendix/operations/audit_prep.md` (new)
  - `src/audit_slot.rs`: only the helpers the card-addressed commands share become `pub(crate)` (the holder lookup by name, the prep and done cores); no change of behavior
  - `src/help.rs`: the rows and examples, under a one-command line lease
  - `tests/fixtures/help/all.txt`, `all-60.txt`, `all-80.txt`, `all-dumb.txt`: regenerated with `MOAI_HELP_GOLDEN_WRITE=1`, one row each, under a recorded leader decision as for audit-slot
- read-only: everything else
- forbidden: `.github`, `.claude/settings.json`, `.moai/project.yaml`

## Plan

1. This SPEC, then the leader's approval (a `LANE-BLOCKED kind=decision` note; no self-approval).
2. `SCOPE-REQUEST` for the code paths above.
3. Tests first (RED): `tests/audit_prep_cli.rs` on temporary repositories and a temporary HOME, AC-1 to AC-6.
4. `src/audit_prep.rs` and `src/ops/audit_prep.rs`; the help rows and goldens; the operation document.
5. Guard mutants, one line per new guard test in the READY.
6. One bare `moai check affected` at READY, then READY with head and base.

## Risks

- The new pair and `audit-slot` are two command surfaces for one job. The new one only calls the old one's functions, and the operation document names `audit-slot` for slot-level work (the leader's `--discard`).
- A build that fails leaves the slot held. The auditor still has to run `audit-done`, and a mutant left behind refuses it by design (REQ-10, REQ-12). A leader repairing a stuck slot uses `audit-slot done --discard`.
- Cargo may recompile the workspace crate more than once if a sibling crate is also dirty. AC-2 and AC-6 count compiled units from Cargo's JSON messages, so a regression shows as a number, not as a time.
- The heavy slot can make `audit-prep` wait. The answer names the wait, as `check affected` does.
- `RUSTC_WRAPPER` with incremental builds made `sccache` refuse (exit 101) in 2026-11-22 audits. The build runs without `RUSTC_WRAPPER` and `CARGO_BUILD_RUSTC_WRAPPER` (D-3).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The commands are the card-addressed front of the fixed slots `audit-1`..`audit-3` and call the code of `audit-slot`; they do not make `audit-<card>` worktrees. | The user-approved redesign §3A fixes the audit slots and reuses them; SPEC-20261123-audit-slot is landed and audited; a second worktree family would copy its link, own-worktree and holder safety code. The dispatch names `.claude/worktrees/audit-<card>`; that part is the leader's to confirm. | Leader picks another shape in the decision note; this SPEC is revised before code. |
| D-2 | The card name is the holder (`C-364` passes the holder-name rule), and the slot is chosen by `auto`. | One card, one audit, one slot; the holder record of audit-slot already refuses a second live holder and takes over a dead one. | Add a `--holder` option later. |
| D-3 | The build is `cargo test --no-run --locked --offline --message-format json` with `--test NAME` at most twice, or `--lib` alone, run in the slot with `RUSTC_WRAPPER` unset, in one heavy slot. `target/` is the slot's own; `CARGO_TARGET_DIR` stays unset. | Card C-364 (mutants build only 1-2 targets), review A2, the sccache-incremental clash of 2026-11-22; Cargo's JSON `compiler-artifact` carries `fresh`. | Change the argument list in one place. |
| D-4 | The cap is two named test targets. | Review A2: 1 to 2 targets per mutant. | Raise the constant. |
| D-5 | `audit-done` has no discard and no force; a left mutant refuses it. | The dispatch: no `--force`; the redesign's removal conditions (clean tree, no process, no claim). | The leader's repair path stays `audit-slot done --discard`. |
| D-6 | The compile-time `env!` paths are not fixed here; a test shows the rebuilt executables read the slot (REQ-9). The touch of tracked sources that `audit-slot prep` already does is what forces the workspace units to recompile in the slot. | Dispatch scope; measured count 163 files, not 135. | A later card moves the paths to run time and then shares nextest archives. |

## Result

- Draft: no code yet. Waiting for the leader's approval of this SPEC.
