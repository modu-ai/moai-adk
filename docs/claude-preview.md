# Claude preview — RC1 work in progress

Removed (Issue #2709). This page described the RC1 Python preview: presentation-only MoAI, MoAI-Easy and MoAI-Learn output styles, a stdin status renderer and a provisional project installer under `adapters/claude/`, with `tests/claude_preview_test.py`. The installer, renderer and test were deleted because no product route uses them; their last version is in the git history before the commit that closed #2709. The three adapted styles and their `LICENSE` and `NOTICE.md` were kept byte for byte under `templates/claude/output-styles/` (leader decision on #2709) for a later user-layer install (#2638).

What replaced each part: project settings — `moai claude-settings-plan` / `claude-settings-apply` (#2239); the status line — `moai-cli statusline` (#2376, `docs/rust-product-path.md`). No Rust route installs output style files yet. The older style sources stay where they were, under `references/legacy-output-styles/`, untouched by this change.
