package codexapp

import "testing"

func TestAppServerWorkspaceSandboxPreservesExternalToolBoundary(t *testing.T) {
	for _, listen := range []string{"stdio://", "unix:///tmp/appserver.sock"} {
		values := map[string]bool{}
		args := appServerArgs(listen)
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "-c" {
				values[args[i+1]] = true
			}
		}
		for _, want := range []string{`sandbox_mode="workspace-write"`, `approval_policy="never"`, `web_search="disabled"`, "features.shell_tool=false", "features.view_image=false", "features.plugins=false", "features.apps=false", "features.codex_hooks=false", "features.hooks=false", "features.plugin_hooks=false", "agents.enabled=false", "features.multi_agent=false", "features.multi_agent_v2=false", "tools.update_plan.enabled=false", "tools.experimental_request_user_input.enabled=false"} {
			if !values[want] {
				t.Errorf("%s missing boundary %s", listen, want)
			}
		}
		for _, forbidden := range []string{`sandbox_mode="read-only"`, `sandbox_mode="danger-full-access"`} {
			if values[forbidden] {
				t.Errorf("%s unexpected sandbox %s", listen, forbidden)
			}
		}
	}
}
