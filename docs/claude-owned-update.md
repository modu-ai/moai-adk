# Claude owned settings UPDATE

Removed (Issue #2709). This page described the provisional Python installer `adapters/claude/install.py` (settings merge, asset copy, `--recovery-plan`). That installer and its tests were deleted; the product never called it (`docs/claude-adapter-contract.md`). Its last version is in the git history before the commit that closed #2709.

The settings merge is the Rust `moai-cli claude-settings-plan` / `claude-settings-apply` (#2239): owned keys, a pending receipt and its own recovery, described in `docs/claude-adapter-contract.md` and `docs/rust-product-path.md`. The status line is `moai-cli statusline` (#2376). No Rust route copies output style files into a project yet; the style files are kept under `templates/claude/output-styles/` (#2709).
