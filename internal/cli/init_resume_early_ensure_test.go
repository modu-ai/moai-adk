package cli

// init_resume_early_ensure_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// finding B, re-aimed by card t1619. A fresh init interrupted after the executor must
// leave a resume checkpoint, so the rerun resumes instead of refusing with "already
// initialized". The user-asset ensure runs before the executor now (card t1587): a
// failing ensure leaves no project state at all (see
// TestEnsureFailureBeforeExecutorLeavesNoProjectState). The interruption this finding
// describes is therefore the post-executor window.
//
// The failing run leaves the project half deployed (llm.yaml without deployment_mode)
// and the harness selection record, both on disk when the window fails. No journal is
// written: the window is not an ensure, so the checkpoint rests on the init marker alone.
