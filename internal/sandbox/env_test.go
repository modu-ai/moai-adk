package sandbox

import (
	"os"
	"strings"
	"testing"
)

// TestEnvScrub_DefaultDenylist verifies that the default 6 patterns
// are scrubbed from the environment.
// REQ-V3R2-RT-003-006
func TestEnvScrub_DefaultDenylist(t *testing.T) {
	// Set up environment with all 6 default scrubbed vars
	// Note: These are test values, NOT real credentials
	parentEnv := []string{
		"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7TEST",
		"AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG",
		"GITHUB_TOKEN=ghp_test_token_not_real",
		"ANTHROPIC_API_KEY=sk-ant-test12345",
		"OPENAI_API_KEY=sk-testfakekey",
		"NPM_TOKEN=npm_test_value",
		"GH_TOKEN=gh_test_value",
		"PATH=/usr/bin:/bin",
		"HOME=/home/user",
	}

	opts := SandboxOptions{
		EnvPassthrough: []string{},
	}

	result := ScrubEnv(parentEnv, opts.EnvPassthrough)

	// Verify scrubbed vars are NOT present
	scrubbedVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"GITHUB_TOKEN",
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"NPM_TOKEN",
		"GH_TOKEN",
	}

	for _, v := range scrubbedVars {
		if envContains(result, v) {
			t.Errorf("Scrubbed var %s should not be present in output", v)
		}
	}

	// Verify non-scrubbed vars ARE present
	if !envContains(result, "PATH=/usr/bin:/bin") {
		t.Error("PATH should be preserved")
	}

	if !envContains(result, "HOME=/home/user") {
		t.Error("HOME should be preserved")
	}
}

// TestEnvScrub_AWSPrefixOnly verifies that AWS_* prefix matching is
// strict (AWSOME_VAR should NOT be scrubbed).
// REQ-V3R2-RT-003-006
func TestEnvScrub_AWSPrefixOnly(t *testing.T) {
	parentEnv := []string{
		"AWS_ACCESS_KEY_ID=testkey",
		"AWS_REGION=us-east-1",
		"AWSOME_VAR=hello", // Should NOT be scrubbed
		"PATH=/usr/bin",
	}

	opts := SandboxOptions{
		EnvPassthrough: []string{},
	}

	result := ScrubEnv(parentEnv, opts.EnvPassthrough)

	// AWS_* vars should be scrubbed
	if envContains(result, "AWS_ACCESS_KEY_ID") {
		t.Error("AWS_ACCESS_KEY_ID should be scrubbed")
	}

	if envContains(result, "AWS_REGION") {
		t.Error("AWS_REGION should be scrubbed")
	}

	// AWSOME_VAR should be preserved
	if !envContains(result, "AWSOME_VAR=hello") {
		t.Error("AWSOME_VAR should be preserved (not AWS_ prefix)")
	}
}

// TestEnvScrub_PassthroughPreserved verifies that vars in the
// passthrough list are preserved even if they're in the denylist.
// REQ-V3R2-RT-003-031
func TestEnvScrub_PassthroughPreserved(t *testing.T) {
	parentEnv := []string{
		"GITHUB_TOKEN=ghp_test_value",
		"ANTHROPIC_API_KEY=sk-ant-testvalue",
		"PATH=/usr/bin",
	}

	opts := SandboxOptions{
		EnvPassthrough: []string{"GITHUB_TOKEN"},
	}

	result := ScrubEnv(parentEnv, opts.EnvPassthrough)

	// GITHUB_TOKEN should be preserved (in passthrough)
	if !envContains(result, "GITHUB_TOKEN=ghp_test_value") {
		t.Error("GITHUB_TOKEN should be preserved (in passthrough list)")
	}

	// ANTHROPIC_API_KEY should be scrubbed (not in passthrough)
	if envContains(result, "ANTHROPIC_API_KEY") {
		t.Error("ANTHROPIC_API_KEY should be scrubbed")
	}

	// PATH should be preserved (not in denylist)
	if !envContains(result, "PATH=/usr/bin") {
		t.Error("PATH should be preserved")
	}
}

// TestEnvScrub_EmptyInput verifies that empty environment input
// is handled gracefully.
func TestEnvScrub_EmptyInput(t *testing.T) {
	parentEnv := []string{}
	opts := SandboxOptions{
		EnvPassthrough: []string{},
	}

	result := ScrubEnv(parentEnv, opts.EnvPassthrough)

	if len(result) != 0 {
		t.Errorf("Empty input should produce empty output, got %d vars", len(result))
	}
}

func envContains(env []string, keyOrValue string) bool {
	for _, e := range env {
		if strings.HasPrefix(e, keyOrValue+"=") || e == keyOrValue {
			return true
		}
	}
	return false
}
