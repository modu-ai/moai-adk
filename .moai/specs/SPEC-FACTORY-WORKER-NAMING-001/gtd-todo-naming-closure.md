# GTD todo naming-axis closure record (card t1085)

- **Date of decision**: 2026-09-22 (operator decision, relayed via card t1085 dispatch).
- **Decision**: the `moai todo` command KEEPS its name. It is NOT renamed to a GTD-branded name. The GTD naming axis is closed as an investigation outcome, not a rename.

## Scope of this closure

- The GTD card family — **t855 (naming), t867 (canon body), t899 (landed store), t939, t940, t941, autonomy ×2** (8 trees total) — is closed as **investigation-only scope**. No further GTD-axis rename or migration work is scheduled from this family.
- **t855 has zero work commits** (the naming card was never executed), so a closure record suffices for it — no code or branch disposition is owed from t855 itself.
- **Disposal boundary**: the physical DISPOSAL of the 8 old GTD work trees is **card t1084's** concern. This record performs no tree disposal and no SPEC status transitions.

## Related completed SPECs (pointers only — untouched by this closure)

- `SPEC-GTD-AUTONOMY-001` — status completed (v0.2.0, 2026-09-15): GTD knowledge graph + pre-delegation autonomy operation.
- `SPEC-GTD-CANON-BODY-001` — status completed (v0.3.2, 2026-09-18, card t867): moved the workflows body into `workflows/gtd.md` and retired the todo body, with compat aliasing.

## What survives unchanged

- The `gtd`↔`todo` compatibility alias surface, guarded by `internal/cli/gtd_compat_test.go` (`TestGTDAllTodoVerbsParity`, `TestGTDFiveStageCLIUsesSameSQLite`, `TestGTDCLIInputRefusalsDoNotFallThrough`), is untouched by this closure and remains in force (SPEC-FACTORY-WORKER-NAMING-001 REQ-009 / AC-008).
- The factory worker-vocabulary rename (`-f agent` → `-f worker`, `lane-N` → `worker-N`) is a separate axis (SPEC-FACTORY-WORKER-NAMING-001 M3/M4) and is unrelated to this GTD naming closure.
