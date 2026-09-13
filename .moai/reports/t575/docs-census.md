# docs-census.md — SPEC-DOCS-TODO-TEMP-GUARD-001 (card t575)

Census of queue-path claims reproduced from acceptance.md §D.5 with per-claim counts (D2 disposition: 21 claims across 5 rows). Truth classification baseline: spec.md §2 truth table (code-cited against `internal/kanban/todo_root.go:81-86,135-144,157-165`, `state_dir.go:28-40`, `temp_origin.go:60-107`). Measured on this tree, HEAD e83b4ec20 (pre-edit content identical to HEAD; edits verified post-hoc by the §E.2 greps in progress.md).

| # | Location | Claim | Pre-fix truth | Post-fix truth status |
|---|----------|-------|---------------|----------------------|
| 1 | `docs-site/content/ko/utility-commands/moai-todo.md:59` | queue stored at `~/.moai/db/<project-key>/todo/backlog.db` (unconditional) | FALSE for temp-origin bases | FIXED — conditioned on temporary origin, `.moai/state/todo/` carve-out + `MOAI_HOME` override note |
| 2 | `docs-site/content/en/utility-commands/moai-todo.md:59` | same | FALSE | FIXED (same conditioning) |
| 3 | `docs-site/content/ja/utility-commands/moai-todo.md:59` | same | FALSE | FIXED (same conditioning) |
| 4 | `docs-site/content/zh/utility-commands/moai-todo.md:59` | same | FALSE | FIXED (same conditioning) |
| 5 | `docs-site/content/ko/utility-commands/moai-todo.md:248` | projects without git metadata use the same home layout (unconditional) | FALSE for temp-origin bases | FIXED — conditioned; project-local path named |
| 6 | `docs-site/content/en/utility-commands/moai-todo.md:248` | same | FALSE | FIXED |
| 7 | `docs-site/content/ja/utility-commands/moai-todo.md:248` | same | FALSE | FIXED |
| 8 | `docs-site/content/zh/utility-commands/moai-todo.md:248` | same | FALSE | FIXED |
| 9 | `docs-site/content/ko/advanced/moai-web-console.md:108` | kanban search order: project-local `.moai/state/todo` first, then home `~/.moai/db/<project-key>/todo` | TRUE (describes lookup candidates in order) | TRUE — unchanged, in-scope-adjacent only |
| 10 | `docs-site/content/en/advanced/moai-web-console.md:108` | same | TRUE | TRUE — unchanged |
| 11 | `docs-site/content/ja/advanced/moai-web-console.md:108` | same | TRUE | TRUE — unchanged |
| 12 | `docs-site/content/zh/advanced/moai-web-console.md:108` | same | TRUE | TRUE — unchanged |
| 13 | `docs-site/content/ko/advanced/factory-mode.md:93` | factory lane ownership at `~/.moai/db/<project-key>/factory/factory.db` | NOT VERIFIED this pass (factory store, separate from todo queue) | unchanged — follow-up candidate (t706 owns factory path) |
| 14 | `docs-site/content/en/advanced/factory-mode.md:93` | same | NOT VERIFIED | unchanged — follow-up |
| 15 | `docs-site/content/ja/advanced/factory-mode.md:93` | same | NOT VERIFIED | unchanged — follow-up |
| 16 | `docs-site/content/zh/advanced/factory-mode.md:93` | same | NOT VERIFIED | unchanged — follow-up |
| 17-20 | `README.ko.md` / `README.en.md` / `README.ja.md` / `README.zh-CN.md` line 80 | same factory `.db` path claim (4 files) | NOT VERIFIED | unchanged — follow-up |
| 21 | `internal/template/templates/.moai/docs/todo-queue-storage.md` lines 4, 106 | home layout explained unconditionally | same defect class; template-shipped doc, out of card scope | unchanged — follow-up finding (t704 owns template doc) |

Count summary: 21 claims — 8 false in-scope (all FIXED), 4 TRUE (unchanged), 8 factory/README not-verified (out of scope), 1 template follow-up (out of scope). Census method per acceptance.md §D.5: `grep -rn 'moai/todo\|moai/db'` over `docs-site/content/` and the four README files, per-hit truth classification against spec.md §2.
