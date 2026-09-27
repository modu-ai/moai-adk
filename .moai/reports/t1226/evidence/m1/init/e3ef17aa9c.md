# M1 — .claude/rules/moai/workflow/kanban-dispatch.md | The env-isolated verification form ¶6

surface = init
mech = M1
gross = 349
chars = 349
dest = workflow/kanban-dispatch-mechanics.md
pointer_file = .moai/reports/t1226/evidence/pointers/init/02f8357cbb.txt
pointer_chars = 87
c2 = Y (a paragraph row has no heading, so no `§ <title>` citation can target it)
c3 = Y (destination paths: see evidence/dest-paths.txt)
c4 = Y (dest size live 5224 / template 5224; cumulative check: AC-ALH-003 (4))
note = measurement record (incident class)

## moved text

````
Measured on Claude Code 2.1.276, inside a worktree session: `env -u FOO echo ok` and `env -u FOO git rev-parse --short HEAD` both ran; `( unset FOO; echo ok )` ran; `( unset FOO; git rev-parse --short HEAD )` and `git -C . rev-parse --short HEAD` were both refused. The last one is the control — it shows the guard was live while `env` was passing.
````
