# /moai gtd — Canonical GTD Entry Point

This file is the canonical workflow identity for GTD task management. It is a
thin dispatch scaffold: policy and state transitions remain owned by the GTD
service and the existing queue implementation.

Pass the supplied arguments to `moai gtd` without rewriting them. The legacy
`/moai todo` entry point routes here as a compatibility alias and therefore
uses the same database, card identities, ordering, archive, and restore path.

Until a GTD-specific verb is selected, preserve the existing queue behavior
documented in `.claude/skills/moai/workflows/todo.md`.
