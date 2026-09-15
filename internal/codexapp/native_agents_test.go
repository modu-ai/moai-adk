package codexapp

import "testing"

func TestAppServerArgsDisableNativeAgents(t *testing.T) {
	for _, listen := range []string{"stdio://", "unix:///tmp/appserver.sock"} {
		t.Run(listen, func(t *testing.T) {
			args := appServerArgs(listen)
			values := make(map[string]bool)
			for i := 0; i+1 < len(args); i++ {
				if args[i] == "-c" {
					values[args[i+1]] = true
				}
			}
			for _, want := range []string{"agents.enabled=false", "features.multi_agent=false", "features.multi_agent_v2=false"} {
				if !values[want] {
					t.Errorf("missing native agent override %q", want)
				}
			}
		})
	}
}
