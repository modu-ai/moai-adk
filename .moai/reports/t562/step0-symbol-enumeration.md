# t562 M2 STEP 0 — test symbol collision enumeration (t562)

- Tree: `2d1dad058` (absorb merge of origin/develop `c72dc1baf`; t563's doctor stat seam and its
  test file `doctor_codex_stale_skill_test.go` arrived via the absorb, not authored by this card).
- Observed compile failure first (`go vet ./internal/cli/`): `statRecorder redeclared in this block`
  (`doctor_codex_stale_skill_test.go:331` vs `codex_skills_prune_readback_test.go:38`), with the
  `rec.paths undefined` cascade in the t563 file (its recorder carries a `paths` field; mine does
  not). Command and exit code recorded in the session report; verbatim lines below.

```
internal/cli/doctor_codex_stale_skill_test.go:331:6: statRecorder redeclared in this block
	internal/cli/codex_skills_prune_readback_test.go:38:6: other declaration of statRecorder
internal/cli/doctor_codex_stale_skill_test.go:342:26: rec.paths undefined (type *statRecorder has no field or method paths)
[... cascade: 402, 403, 406, 407, 705, 706, 709, 710 ...]
```

## Full enumeration — package-scope identifiers declared by this card's two M1 test files

Nine declarations (4 helpers, 5 Test funcs), each cross-checked against every OTHER `*_test.go`
in `internal/cli` via `/usr/bin/grep -l '\b<name>\b'`:

| Identifier | Kind | Declared at | Cross-check result |
|---|---|---|---|
| `overrideSeparator` | func helper | codex_skills_prune_readback_test.go:30 | CLEAN |
| `statRecorder` | type | codex_skills_prune_readback_test.go:38 | **COLLIDES** with `doctor_codex_stale_skill_test.go:331` (t563, absorbed) |
| `stubRecordingStat` | func helper | codex_skills_prune_readback_test.go:45 | CLEAN |
| `readbackEntry` | func helper | codex_skills_prune_readback_test.go:59 | CLEAN |
| `TestJudgeCodexSkillEntry_SeparatorConversion` | test | :73 | CLEAN |
| `TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion` | test | :99 | CLEAN |
| `TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative` | test | :122 | CLEAN |
| `TestJudgeCodexSkillEntry_EligibilityGatingPins` | test | :146 | CLEAN |
| `TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard` | test | codex_stale_skill_readback_test.go:32 | CLEAN |

Result: exactly ONE collision (`statRecorder`). No other identifier of this card collides with any
other test file in the package. Repair is a rename in THIS CARD'S file only
(`statRecorder` → `pruneReadbackStatRecorder`); the absorbed t563 file and all production files
are untouched.
