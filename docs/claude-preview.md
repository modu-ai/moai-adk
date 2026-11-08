# Claude preview — RC1 work in progress

Confirmed: Rust, one common skill, online-first without a MoAI server, and SETUP/UPDATE in both Claude Code and Codex are RC goals. P02 discovery precedence is decided; its exact schema and profile details remain open. P05 package/temporary bridge is undecided. This preview does not decide either.

Minimal Claude preview: presentation-only versions of MoAI, MoAI-Easy and MoAI-Learn; a stdin-only status renderer; project-only settings; and fixture tests. No hooks, provider secret loading, global settings, Go bridge or claims of complete RC. Renderer currently displays only provided model and context percentage; other data is unavailable, not zero.

Source: modu-ai/moai-adk commit 4fa2aa9db2e0add0f1299e35407141b2b41c5ec0, Apache-2.0, `.claude/output-styles/moai/`; source hashes: MoAI 8244a7bb3a53436aca882d7c46c7a59866f59fc3a12d232a5738a25a2f7b5f04, Easy c33a60904e7908fe6387d69d57f489b3565f66a2449b5103b931c8a526b8d050, Learn ef01dad72d0f76763209a94af38fee7392d12d1a98df44c6a38ff74d9f0c5638. These are adapted summaries, not full semantic equivalence; review and host testing remain required.

Run `python3 tests/claude_preview_test.py`. This provisional fixture is not the RC Rust installer. It installs only into a supplied project, refusing conflicts. Do not use it to replace user settings.

## Verification and limits

The source files and Apache-2.0 license are preserved under `references/legacy-output-styles/`; they are not loaded as styles. The adapted styles retain selected banner and learning patterns, but do not yet establish full presentation equivalence. Execution/authority instructions from the old implementation are intentionally not copied into the styles; a reviewed common workflow has not yet been implemented.

The installer preserves unknown JSON keys and connects owned-key three-way UPDATE to actual settings writes. Existing matching values are not adopted. Persistent pending receipts block subsequent updates after interruption; `--recovery-plan` diagnoses snapshots without modifying files. Automatic recovery is not implemented. Settings are checked again before receipt completion, including later edits and deletion. Check-then-replace is not OS-level compare-and-swap. See `claude-owned-update.md` for limits. License and adaptation notices are installed alongside styles under `.moai-cli/licenses/claude-output-styles/`.

The Claude statusline environment variable and visual rendering remain unverified in an actual host. Claude 2.1.319 version detection is not a host-loading test. No hooks are installed.
