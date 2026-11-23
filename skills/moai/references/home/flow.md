# MoAI evidence and handoff

Apply this file only in a project configured for MoAI; the project's own CLAUDE.md wins where it conflicts with this file.

- Make the smallest change that meets the agreed acceptance criteria. Use the project's verification commands; record the tested revision, the exact checks, and any check that was skipped or failed. Missing evidence is never a PASS.
- Keep implementation and independent review distinct when the configured workflow requires it; a reviewer checks the candidate revision, not a stale one.
- When blocked, record the scope, the revision, the remaining checks and the condition for resuming. Never overwrite changes you do not own, and never bypass ownership, permissions or a release gate.
- Keep secrets and private data out of shared artifacts. Land and publish only through the project's configured workflow; a local merge does not authorize a push or a release.
- Delete only inside your own worktree or a temporary folder, by a relative path; never remove an absolute or home path.

The lane, card and landing commands and the full flow are in the moai skill's `references/flows/next-flow.md`; read it when a card or a landing is involved.
