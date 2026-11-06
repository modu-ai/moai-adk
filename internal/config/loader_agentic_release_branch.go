package config

import (
	"os"
	"path/filepath"
	"strings"
)

// agenticTrunkReleaseBatchName is the fixed tail of the af-flow release
// branch: release_branch_prefix + "main-batch" (SPEC-FACTORY-COMPLETION-RECOVERY-001
// REQ-FCR-026, card t2013). The branch is resolved ONLY for the release
// completion route; the af-flow IntegrationTarget stays the fixed main.
const agenticTrunkReleaseBatchName = "main-batch"

// LoadAgenticTrunkReleaseBranch returns the fixed af-flow release branch
// of the project at projectRoot — the active profile's release_branch_prefix
// followed by "main-batch", e.g. "release/main-batch" — or "" when the active
// profile's workflow is not af-flow, the prefix is empty, or the file is
// unreadable. It is a separate read on purpose: GitFlowIntegrationConfig does
// not carry the prefix, and widening that struct would move the other
// consumers of the integration target.
//
// Card t2056: the workflow gate reads the same effective workflow as
// LoadGitFlowIntegrationConfig (effectiveWorkflow), so in a factory leader or
// lane session a profile that sets no workflow counts as af-flow here too.
func LoadAgenticTrunkReleaseBranch(projectRoot string) string {
	_, factorySession := DetectRoleMarker(os.Getenv)
	return loadAgenticTrunkReleaseBranch(projectRoot, factorySession)
}

// loadAgenticTrunkReleaseBranch is the pure body (factorySession is a
// parameter, never read from the environment).
func loadAgenticTrunkReleaseBranch(projectRoot string, factorySession bool) string {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &gitStrategyFileWrapper{}
	loaded, err := loadYAMLFile(dir, "git-strategy.yaml", wrapper)
	if err != nil || !loaded {
		return ""
	}
	profile, ok := wrapper.GitStrategy.ActiveModeProfile()
	if !ok || effectiveWorkflow(profile, factorySession) != WorkflowAFFlow {
		return ""
	}
	prefix := strings.TrimSpace(profile.ReleaseBranchPrefix)
	if prefix == "" {
		return ""
	}
	return prefix + agenticTrunkReleaseBatchName
}
