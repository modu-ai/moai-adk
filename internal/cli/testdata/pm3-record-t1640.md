# PM-3 measurement record — tracked copy for AC-WHR-006's branch selector.
#
# Source of record: .moai/reports/t1640/pre-kickoff-measurements.md
# (gitignored, machine-local); the tracked run-phase evidence in
# .moai/specs/SPEC-WORKTREE-HOME-ROOT-001/progress.md section E.2 cites it.
# Operator ruling bd72 (board, 2026-10-19) authorized the lane-executed
# measurement. The RESULT line below is the verbatim one-line record; the
# AC-WHR-006 test parses exactly this line and fails on absence or GAP.
#
# Consequence (spec section F): FAIL selects branch Y of D-5 — the launcher
# passes a short name with no existing tree through unchanged and creates
# nothing (REQ-WHR-006 branch Y); the case is the single documented exception
# to "no creation path names a legacy root" (REQ-WHR-020).

RESULT PM-3: FAIL ;; COMMAND: cd "$HOME/.moai/scratch/t1640-pm3/repo" && claude --worktree "$HOME/.moai/worktrees/projects/pm3-scratch/pm3tree" -p "Run the shell commands pwd and git rev-parse --show-toplevel and reply with their two output lines only" --output-format text ;; OUTPUT: (stdout empty; stderr verbatim:) `Warning: claude.ai connectors are disabled because ANTHROPIC_API_KEY or another auth source is set and takes precedence over your claude.ai login · Unset it to load your organization's connectors` + `Error creating worktree: Invalid worktree name "/Users/goos/.moai/worktrees/projects/pm3-scratch/pm3tree": each "/"-separated segment must be non-empty and contain only letters, digits, dots, underscores, and [the name-charset set]` ;; EXIT: 1 ;; TOOLS: claude 2.1.303 (Claude Code) at /Users/goos/.local/bin/claude ;; TREE: aa3f9a79b484ac4fb64f13abc02a932ec2876b89
