# t661 — C wrapper pre-scan (reading only, no execution)

Question: can wrapping `userHomeDirFn` in TestMain (redirect only the captured real home to a sandbox)
change the result of any other `internal/cli` test?

## Direct seam reads in tests
`grep -rn --include='*_test.go' -E '\buserHomeDirFn\(\)|\buserHomeDir\(\)|glmcred\.HomeDirFn\(\)|kanban\.HomeDirFn\(\)' internal/cli`
→ only the lines this card added (`update_clean_install_test.go:739`, `update_skip_sync_test.go:156`,
`main_test.go:440/463`). No pre-existing test reads the seam directly.

## Tests computing an expectation from `os.UserHomeDir()`
`grep -rln --include='*_test.go' 'os\.UserHomeDir()' internal/cli` → 9 files. Context read for each:
- `update_home_seam_test.go:138` — comment only.
- `launcher_test.go:52`, `launcher_worktree_l2_test.go:37-42`, `update_test.go:31/1487/1781` — set HOME
  (and USERPROFILE) to a temp dir first; the wrapper passes a non-real home through unchanged.
- `migrate_profiles_test.go:202`, `memory_test.go:198` — capture the real home only to assert the
  derivation does NOT land in it, then set HOME to a temp dir; unaffected.
- `coverage_improvement_test.go:2790/6275` — pass `os.UserHomeDir()` straight to
  `detectGoBinPathForUpdate(homeDir)`; the seam is not consulted; unaffected.
- `main_test.go` — this card's own guard. `preference/home_isolation_test.go` is a different package.

Reading verdict: no pre-existing test in `internal/cli` compares a seam-resolved home against the real
home without overriding HOME, so the wrapper is predicted not to flip any assertion.

## Remaining exposure (not an assertion flip)
A test that, without overriding HOME, reads real-home state through `userHomeDirFn` /
`glmcred.HomeDirFn` / `kanban.HomeDirFn` and skips when the state is absent would now see the empty
sandbox and could turn a pass into a skip.

Found instance: `TestAuditPinLive_GLMDifferential` (`audit_pin_live_test.go:182-191`) calls
`loadGLMKey()` → `glmcred.Load()` → `glmcred.Path()` → `glmcred.HomeDirFn` (aliased to
`userHomeDirFn`, `glm.go:36`) without overriding HOME. It is opt-in (`auditPinRequireLive`, skips unless
the live env var is `1`). With the live env set, the wrapper would make it read the empty sandbox
`.moai/.env.glm` and take its SKIP branch (writing the "credential absent" evidence marker) instead of
the live z.ai call. Default runs (live env unset) are unaffected. Other live test files
(`codex_live_*`, `codex_review_*_live_test.go`) were listed but not read.

Checked and unaffected: `TestRunGLM_SavesAPIKey` (`coverage_improvement_test.go:1358`) sets HOME and
USERPROFILE to a temp dir first; `TestLoadGLMKey_*` set HOME.

## Gaps
- Prediction only; confirmed or refuted by a full `internal/cli` package run in a separate slot.
