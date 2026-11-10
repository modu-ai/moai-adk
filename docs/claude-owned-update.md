# Claude owned settings UPDATE

The project installer now connects a three-way settings planner to actual file application. Use `python3 adapters/claude/install.py PROJECT --style MoAI` to select one of MoAI, MoAI-Easy, and MoAI-Learn in this provisional implementation.

Only keys newly created by this installer become owned. Existing identical keys are preserved without adoption. `.moai-cli/claude-settings-owner.json` stores schema version, transaction state and owned baseline. Unknown settings remain user-owned. Baseline/current/desired conflicts stop before asset application. When the desired template is unchanged, user edits are preserved.

Before settings mutation, a persistent pending receipt records the old baseline and desired owned values. After successful settings application it becomes committed. An interruption leaves pending and blocks further automatic updates. There is no automatic rollback or journal recovery yet: inspect and reconcile the affected settings and receipt without overwriting later user changes. The receipt is local mutable evidence, not cryptographic authorization.

Fixture tests cover actual settings updates, unchanged-template user edits, conflicts, matching unowned settings, repeat application, failure after settings write, pending detection and subsequent user-change preservation. Existing renderer/install and planner tests also pass.

The installer still uses Python, not the final Rust product path. Existing installations without a receipt are not silently adopted. Asset version updates, AGENTS merge, multi-host UPDATE and actual host loading are not implemented here. Check-then-replace is not OS-level compare-and-swap against uncooperative writers. Multiple files are not one atomic transaction.

## Read-only recovery and completion boundary

`python3 adapters/claude/install.py PROJECT --recovery-plan` classifies unowned, committed, legacy pending, before/applied snapshot matches, or conflicting current content. Before/after SHA-256 values are evidence only. The command does not restore settings, delete receipts, or expose their values. Invalid JSON constants and malformed receipts are rejected. No full backup is stored, so pending transactions require manual review; this is diagnostic support, not automatic known-good recovery.

Immediately before committing a receipt, the installer checks the applied settings bytes and rejects observed later edits, deletion or symlink changes while retaining pending state. This narrows the observation window; it does not eliminate concurrent writer races. Interrupted receipt and CLI fixtures use temporary directories and do not invoke hosts.

F01 installation/UPDATE is partially connected through the provisional Python adapter. The Rust `experiments/rc-cli` command remains a status/version experiment and does not invoke the adapter. The root Rust product now handles minimal portable profile creation and package integrity selection (see `rust-product-path.md`). It does not invoke this adapter. Both-host preflight/application, AGENTS merge and full interview reconfiguration remain incomplete. F07/F19 local TODO promotion was removed by the user decision of 2026-11-10 (#2384): the work ledger is forge Issues only.
