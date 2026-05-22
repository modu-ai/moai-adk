package sandbox

import (
	"strings"
	"sync"
)

// @MX:ANCHOR: [AUTO] ScrubEnv is the single entry point for sandbox environment-variable scrubbing
// @MX:REASON: Fan_in >= 3: BubblewrapBackend.Exec, SeatbeltBackend.Exec,
//             DockerBackend.Exec, TestEnvScrub_* tests — security-critical function
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-006/031

// awsOnce pre-compiles the AWS_ prefix string for O(1) lookup.
// Using sync.Once prevents per-Exec string allocation.
var (
	awsOnce   sync.Once
	awsPrefix string
)

func initAWSPrefix() {
	awsOnce.Do(func() {
		awsPrefix = "AWS_"
	})
}

// defaultDenyList is the enumerated list of environment variable names that are
// stripped from the child process environment by default.
// Per REQ-V3R2-RT-003-006: enumerated (not regex) — projects extend via
// security.yaml sandbox.env_scrub_extra (additive).
var defaultDenyList = []string{
	"GITHUB_TOKEN",
	"ANTHROPIC_API_KEY",
	"OPENAI_API_KEY",
	"NPM_TOKEN",
	"GH_TOKEN",
}

// ScrubEnv removes sensitive environment variables from parent and returns a
// filtered copy. Variables in the passthrough list are preserved even if they
// match the denylist.
//
// Scrubbing rules (in priority order):
//  1. If the variable name is in passthrough → keep unconditionally.
//  2. If the variable name starts with "AWS_" → remove.
//  3. If the variable name is in defaultDenyList → remove.
//  4. Otherwise → keep.
//
// The function never mutates the input slices.
func ScrubEnv(parent []string, passthrough []string) []string {
	if len(parent) == 0 {
		return []string{}
	}

	initAWSPrefix()

	// Convert passthroughSet into a map for fast lookup
	ptSet := make(map[string]bool, len(passthrough))
	for _, v := range passthrough {
		ptSet[v] = true
	}

	// denySet
	denySet := make(map[string]bool, len(defaultDenyList))
	for _, k := range defaultDenyList {
		denySet[k] = true
	}

	result := make([]string, 0, len(parent))
	for _, kv := range parent {
		key := envKey(kv)

		// 1. passthrough takes priority
		if ptSet[key] {
			result = append(result, kv)
			continue
		}

		// 2. AWS_ prefix
		if strings.HasPrefix(key, awsPrefix) {
			continue
		}

		// 3. default denylist
		if denySet[key] {
			continue
		}

		// 4. Otherwise keep
		result = append(result, kv)
	}

	return result
}

// envKey extracts the key portion from a "KEY=VALUE" env string.
func envKey(kv string) string {
	i := strings.IndexByte(kv, '=')
	if i < 0 {
		return kv
	}
	return kv[:i]
}
