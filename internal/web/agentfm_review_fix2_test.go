// agentfm_review_fix2_test.go — F11: the user-install root classifies as
// core (retained agents render), harness specialists under the same root do
// not (card t1509 review-fix round 2).
package web

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

func TestRF2F11_AgentfmClassifiesUserInstall(t *testing.T) {
	orig := homeAgentsDir
	defer func() { homeAgentsDir = orig }()
	homeAgentsDir = filepath.Join(string(filepath.Separator), "tmp", "rf2-home-agents")

	core := agentfm.AgentInfo{Path: filepath.Join(homeAgentsDir, "manager-spec.md")}
	harness := agentfm.AgentInfo{Path: filepath.Join(homeAgentsDir, "harness", "spec.md")}
	if !agentIsMoaiCore(core) {
		t.Error("F11: user-install retained agent not classified core — console rows vanish")
	}
	if agentIsMoaiCore(harness) {
		t.Error("F11: harness specialist misclassified core")
	}
}
