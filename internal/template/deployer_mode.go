package template

import "strings"

// deployer_mode.go — the deploy-mode surface after SPEC-USER-ASSET-INSTALL-001
// M7. The plugin-mode split is RETIRED with its carrier: the plugin payload
// constant, the mirror policy, the exclusion walk branch, the .mcp.json
// strip filter, the plugin re-home path, and the mirror re-home entry point
// are all gone. What remains:
//
//   - DeployMode — the project's deployment_mode RECORD surface only. The
//     record exists in deployed projects and update never flips it (REQ-018);
//     the deployer itself no longer branches on it and carries a single
//     project payload shape.
//   - isCommonAssetRoot — the REQ-005 walk exclusion: the project payload
//     carries no common skill or agent file in any mode; the four user roots
//     (plus the manifest/backup homes) are the install surface.

// DeployMode selects the deploy file set. Post-M7 only the LOCAL value
// remains meaningful.
type DeployMode string

const (
	// DeployModeLocal — the single project payload shape (the zero value
	// and every path).
	DeployModeLocal DeployMode = "local"
)

// isCommonAssetRoot reports whether a deploy-relative path lives under one
// of the project common-asset roots the user-folder install replaces
// (SPEC-USER-ASSET-INSTALL-001 REQ-005: the project payload carries no
// common skill or agent file in any mode).
func isCommonAssetRoot(relPath string) bool {
	for _, root := range []string{
		".claude/skills/",
		".claude/commands/moai/",
		".claude/agents/moai/",
		".agents/skills/",
		".codex/agents/moai/",
	} {
		if strings.HasPrefix(relPath, root) {
			return true
		}
	}
	return false
}
