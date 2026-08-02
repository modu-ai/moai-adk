package config

// session_worktree_test.go — SPEC-SESSION-WORKTREE-001 M1 table-driven coverage
// for the SessionWorktreeEnabled activation decision over the
// (workflow.session_worktree.enabled, MOAI_SESSION_WORKTREE) matrix.
//
// The matrix pins REQ-SW-001 (default-off), REQ-SW-002 (config flag on), and
// REQ-SW-003 (env wins over config both directions, AC-SW-004a / AC-SW-004b).

import "testing"

// TestSessionWorktreeEnabledMatrix covers every combination of the config flag
// and the three env states (unset / "1" / "0"). Env wins over config;
// unset env falls through to config; default-off when both unset.
// Subtests mutate the process env via t.Setenv, so they MUST run serially
// (Go prohibits t.Setenv under t.Parallel). The matrix is small (8 cases).
func TestSessionWorktreeEnabledMatrix(t *testing.T) {
	cases := []struct {
		name       string
		configFlag bool
		envValue   string // "" = unset, "1" = force on, "0" = force off
		envSet     bool   // whether to set the env var at all (distinguishes "" from unset)
		want       bool
	}{
		// REQ-SW-001 — default-off when both unset.
		{
			name:       "config_false_env_unset_default_off",
			configFlag: false,
			envValue:   "",
			envSet:     false,
			want:       false,
		},
		// REQ-SW-002 — config flag ON, env unset → ON.
		{
			name:       "config_true_env_unset_falls_through_on",
			configFlag: true,
			envValue:   "",
			envSet:     false,
			want:       true,
		},
		// REQ-SW-003 / AC-SW-004a — env=1 forces ON regardless of config false.
		{
			name:       "config_false_env_1_forces_on",
			configFlag: false,
			envValue:   "1",
			envSet:     true,
			want:       true,
		},
		// REQ-SW-003 / AC-SW-004a — env=1 forces ON even when config true.
		{
			name:       "config_true_env_1_forces_on",
			configFlag: true,
			envValue:   "1",
			envSet:     true,
			want:       true,
		},
		// REQ-SW-003 / AC-SW-004b — env=0 forces OFF regardless of config true.
		{
			name:       "config_true_env_0_forces_off",
			configFlag: true,
			envValue:   "0",
			envSet:     true,
			want:       false,
		},
		// REQ-SW-003 / AC-SW-004b — env=0 forces OFF even when config false.
		{
			name:       "config_false_env_0_forces_off",
			configFlag: false,
			envValue:   "0",
			envSet:     true,
			want:       false,
		},
		// Unrecognized env value does NOT override — falls through to config.
		{
			name:       "config_false_env_garbage_falls_through_off",
			configFlag: false,
			envValue:   "yes",
			envSet:     true,
			want:       false,
		},
		{
			name:       "config_true_env_garbage_falls_through_on",
			configFlag: true,
			envValue:   "true",
			envSet:     true,
			want:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envSet {
				t.Setenv(EnvSessionWorktree, tc.envValue)
			}

			cfg := &Config{
				Workflow: WorkflowConfig{
					SessionWorktree: SessionWorktreeConfig{Enabled: tc.configFlag},
				},
			}

			got := SessionWorktreeEnabled(cfg)
			if got != tc.want {
				t.Fatalf("SessionWorktreeEnabled(cfg) = %v, want %v (configFlag=%v, env=%q)",
					got, tc.want, tc.configFlag, tc.envValue)
			}
		})
	}
}

// TestSessionWorktreeEnabledNilConfigSafety documents the nil-config contract:
// a nil config means "no activation" (false), so a caller that failed to load
// config never accidentally enables worktree isolation.
func TestSessionWorktreeEnabledNilConfigSafety(t *testing.T) {
	// Unset env + nil cfg → false (no activation). t.Setenv → serial.
	t.Setenv(EnvSessionWorktree, "")
	if got := SessionWorktreeEnabled(nil); got != false {
		t.Fatalf("SessionWorktreeEnabled(nil) with env unset = %v, want false", got)
	}
}

// TestSessionWorktreeEnabledDefaultOffWhenZeroValue confirms the distributed
// default: a freshly constructed default WorkflowConfig has the feature OFF
// (REQ-SW-001 / AC-SW-001 / AC-SW-002 — byte-identical baseline).
func TestSessionWorktreeEnabledDefaultOffWhenZeroValue(t *testing.T) {
	// t.Setenv → serial.
	t.Setenv(EnvSessionWorktree, "") // ensure env does not interfere

	cfg := &Config{Workflow: NewDefaultWorkflowConfig()}
	if cfg.Workflow.SessionWorktree.Enabled != false {
		t.Fatalf("NewDefaultWorkflowConfig() SessionWorktree.Enabled = %v, want false",
			cfg.Workflow.SessionWorktree.Enabled)
	}
	if got := SessionWorktreeEnabled(cfg); got != false {
		t.Fatalf("SessionWorktreeEnabled(default cfg) = %v, want false", got)
	}
}
