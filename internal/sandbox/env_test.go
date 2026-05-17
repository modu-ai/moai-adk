package sandbox

import (
	"testing"
)

// TestEnvScrub_DefaultDenylist verifies that the default 6-pattern denylist is applied.
// RED: fails until env.go::ScrubEnv is created.
func TestEnvScrub_DefaultDenylist(t *testing.T) {
	t.Parallel()

	parent := []string{
		"PATH=/usr/bin:/bin",
		"HOME=/root",
		"ANTHROPIC_API_KEY=sk-ant-test",
		"OPENAI_API_KEY=sk-openai-test",
		"GITHUB_TOKEN=ghp_test",
		"GH_TOKEN=ghp_test2",
		"NPM_TOKEN=npm_test",
		"MY_SAFE_VAR=keep",
	}

	result := ScrubEnv(parent, nil)

	// 제거되어야 할 변수들
	for _, key := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GITHUB_TOKEN", "GH_TOKEN", "NPM_TOKEN"} {
		if envContainsKey(result, key) {
			t.Errorf("ScrubEnv: %s should be removed from env", key)
		}
	}

	// 유지되어야 할 변수들
	for _, key := range []string{"PATH", "HOME", "MY_SAFE_VAR"} {
		if !envContainsKey(result, key) {
			t.Errorf("ScrubEnv: %s should be preserved in env", key)
		}
	}
}

// TestEnvScrub_AWSPrefixOnly verifies that AWS_* prefix-match removes all AWS_ vars.
// RED: fails until env.go handles prefix matching.
func TestEnvScrub_AWSPrefixOnly(t *testing.T) {
	t.Parallel()

	parent := []string{
		"AWS_ACCESS_KEY_ID=AKIA123",
		"AWS_SECRET_ACCESS_KEY=secret",
		"AWS_SESSION_TOKEN=token",
		"AWS_REGION=us-east-1",
		"AWESOME_VAR=keep", // not AWS_ prefix — must be kept
		"NORMAL=keep",
	}

	result := ScrubEnv(parent, nil)

	// 모든 AWS_ 접두사 변수 제거 확인
	for _, key := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_REGION"} {
		if envContainsKey(result, key) {
			t.Errorf("ScrubEnv: %s (AWS_ prefix) should be removed", key)
		}
	}

	// AWESOME_VAR은 "AWS_"가 아니므로 유지
	if !envContainsKey(result, "AWESOME_VAR") {
		t.Error("ScrubEnv: AWESOME_VAR should be preserved (not AWS_ prefix)")
	}
	if !envContainsKey(result, "NORMAL") {
		t.Error("ScrubEnv: NORMAL should be preserved")
	}
}

// TestEnvScrub_PassthroughPreserved verifies that explicitly listed vars survive scrubbing.
// RED: fails until env.go handles passthrough.
func TestEnvScrub_PassthroughPreserved(t *testing.T) {
	t.Parallel()

	parent := []string{
		"GH_TOKEN=ghp_should_be_passed",
		"ANTHROPIC_API_KEY=sk-ant-keep",
		"SAFE_VAR=keep",
	}
	passthrough := []string{"GH_TOKEN", "ANTHROPIC_API_KEY"}

	result := ScrubEnv(parent, passthrough)

	// passthrough 목록에 있는 변수는 유지
	if !envContainsKey(result, "GH_TOKEN") {
		t.Error("ScrubEnv: GH_TOKEN should be preserved via passthrough")
	}
	if !envContainsKey(result, "ANTHROPIC_API_KEY") {
		t.Error("ScrubEnv: ANTHROPIC_API_KEY should be preserved via passthrough")
	}
	if !envContainsKey(result, "SAFE_VAR") {
		t.Error("ScrubEnv: SAFE_VAR should be preserved")
	}
}

// TestEnvScrub_EmptyInput verifies that empty parent env returns empty result.
// RED: fails until env.go exists.
func TestEnvScrub_EmptyInput(t *testing.T) {
	t.Parallel()

	result := ScrubEnv(nil, nil)
	if result == nil {
		// nil is acceptable — just not panicking
		result = []string{}
	}
	if len(result) != 0 {
		t.Errorf("ScrubEnv(nil, nil): expected empty result, got %v", result)
	}

	result2 := ScrubEnv([]string{}, nil)
	if len(result2) != 0 {
		t.Errorf("ScrubEnv([], nil): expected empty result, got %v", result2)
	}
}

// envContainsKey returns true if the env slice contains KEY=... for the given key.
func envContainsKey(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
