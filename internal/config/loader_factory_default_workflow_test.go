package config

// loader_factory_default_workflow_test.go — card t2056: entering a factory
// session (the -f leader or a -l lane) defaults an UNSET Git workflow to
// af-flow. An explicit setting always wins, a plain session is unchanged, and
// the user's git-strategy.yaml is never rewritten.

import (
	"os"
	"path/filepath"
	"testing"
)

// t2056Project writes a git-strategy.yaml (body "" writes no file) and returns
// the project root and the file path.
func t2056Project(t *testing.T, body string) (root, path string) {
	t.Helper()
	root = t.TempDir()
	path = filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	if body == "" {
		return root, path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, path
}

// t2056Session sets the factory markers a launcher stamps: "leader" carries
// only the run's lane count (the leader entry stamps no role value), "lane"
// carries the worker label, the role and the count, "" clears all three (a
// plain session).
func t2056Session(t *testing.T, kind string) {
	t.Helper()
	t.Setenv(EnvMoaiFactoryWorker, "")
	t.Setenv(EnvMoaiFactoryWorkers, "")
	t.Setenv(EnvFactoryRole, "")
	switch kind {
	case "leader":
		t.Setenv(EnvMoaiFactoryWorkers, "3")
	case "lane":
		t.Setenv(EnvMoaiFactoryWorker, "lane-1")
		t.Setenv(EnvFactoryRole, FactoryRoleLane)
		t.Setenv(EnvMoaiFactoryWorkers, "0")
	}
}

const (
	t2056KeyMissing = "git_strategy:\n  mode: manual\n  manual:\n    develop_branch: develop\n"
	t2056Blank      = "git_strategy:\n  mode: manual\n  manual:\n    workflow: \"\"\n"
)

func t2056Explicit(workflow string) string {
	return "git_strategy:\n  mode: manual\n  manual:\n    workflow: " + workflow + "\n    develop_branch: develop\n"
}

// TestT2056_UnsetWorkflowDefaultsToAFFInAFactorySession — an active profile
// with no workflow (key missing, or a blank value) reads as af-flow when the
// session is the factory leader or a lane.
func TestT2056_UnsetWorkflowDefaultsToAFFInAFactorySession(t *testing.T) {
	for _, session := range []string{"leader", "lane"} {
		for name, body := range map[string]string{"key-missing": t2056KeyMissing, "blank": t2056Blank} {
			t2056Session(t, session)
			root, _ := t2056Project(t, body)
			c := LoadGitFlowIntegrationConfig(root)
			if c.Workflow != WorkflowAFFlow {
				t.Errorf("%s/%s: Workflow = %q, want %q", session, name, c.Workflow, WorkflowAFFlow)
			}
			if c.Disposition != DispositionValidNonGitFlow {
				t.Errorf("%s/%s: Disposition = %q, want %q", session, name, c.Disposition, DispositionValidNonGitFlow)
			}
			if !c.AgenticTrunkWorkflow || !c.UsesLocalIntegration() {
				t.Errorf("%s/%s: the af-flow predicates are not set (AgenticTrunkWorkflow=%v UsesLocalIntegration=%v)", session, name, c.AgenticTrunkWorkflow, c.UsesLocalIntegration())
			}
			if c.IntegrationTarget != "main" {
				t.Errorf("%s/%s: IntegrationTarget = %q, want main", session, name, c.IntegrationTarget)
			}
		}
	}
}

// TestT2056_ExplicitWorkflowAlwaysWins — a workflow the user wrote is never
// replaced by the default, in any session kind; the legacy agentic-trunk
// spelling still reads as af-flow.
func TestT2056_ExplicitWorkflowAlwaysWins(t *testing.T) {
	want := map[string]string{
		"github-flow":   WorkflowGitHubFlow,
		"git-flow":      WorkflowGitFlow,
		"gitlab-flow":   WorkflowGitLabFlow,
		"release-flow":  WorkflowReleaseFlow,
		"af-flow":       WorkflowAFFlow,
		"agentic-trunk": WorkflowAFFlow,
	}
	for _, session := range []string{"", "leader", "lane"} {
		for raw, normalized := range want {
			t2056Session(t, session)
			root, _ := t2056Project(t, t2056Explicit(raw))
			if got := LoadGitFlowIntegrationConfig(root).Workflow; got != normalized {
				t.Errorf("session=%q workflow=%s: Workflow = %q, want %q", session, raw, got, normalized)
			}
		}
	}
}

// TestT2056_PlainSessionAndAbsentFileAreUnchanged — outside -f/-l an unset
// workflow keeps today's reading (invalid, no default), and a project with no
// git-strategy.yaml stays unknown in every session kind.
func TestT2056_PlainSessionAndAbsentFileAreUnchanged(t *testing.T) {
	t2056Session(t, "")
	for name, body := range map[string]string{"key-missing": t2056KeyMissing, "blank": t2056Blank} {
		root, _ := t2056Project(t, body)
		c := LoadGitFlowIntegrationConfig(root)
		if c.Workflow != "" || c.Disposition != DispositionInvalid {
			t.Errorf("plain/%s: Workflow=%q Disposition=%q, want \"\" / %q", name, c.Workflow, c.Disposition, DispositionInvalid)
		}
	}
	for _, session := range []string{"", "leader", "lane"} {
		t2056Session(t, session)
		root, _ := t2056Project(t, "")
		c := LoadGitFlowIntegrationConfig(root)
		if c.Workflow != "" || c.Disposition != "" {
			t.Errorf("absent file, session=%q: Workflow=%q Disposition=%q, want unknown", session, c.Workflow, c.Disposition)
		}
	}
}

// TestT2056_ReleaseBranchReaderSharesTheDefault — card-review (d7e46307e4):
// LoadAgenticTrunkReleaseBranch gates on the same workflow. Card t2098
// narrows the reader further: the release route is opt-in through
// release_route: true (default OFF by operator decision — the af-flow
// landing surface is local main), so a release_branch_prefix alone yields ""
// in every session kind, and the opt-in bodies keep the workflow gate. In a
// factory session an unset workflow still reads as af-flow — otherwise the
// two readers disagree and `factory complete` takes the main route instead
// of the release route.
func TestT2056_ReleaseBranchReaderSharesTheDefault(t *testing.T) {
	withPrefix := func(workflow string, route bool) string {
		body := "git_strategy:\n  mode: manual\n  manual:\n    release_branch_prefix: release/\n"
		if route {
			body += "    release_route: true\n"
		}
		if workflow != "" {
			body += "    workflow: " + workflow + "\n"
		}
		return body
	}
	type tc struct {
		name, session, workflow, want string
		route                         bool
	}
	for _, c := range []tc{
		{name: "factory leader, key missing, route off", session: "leader", workflow: withPrefix("", false)},
		{name: "factory lane, key missing, route off", session: "lane", workflow: withPrefix("", false)},
		{name: "factory lane, explicit af-flow, route off", session: "lane", workflow: withPrefix("af-flow", false)},
		{name: "factory lane, explicit github-flow, route off", session: "lane", workflow: withPrefix("github-flow", false)},
		{name: "plain session, key missing, route off", session: "", workflow: withPrefix("", false)},
		{name: "plain session, explicit af-flow, route off", session: "", workflow: withPrefix("af-flow", false)},
		{name: "factory leader, key missing, route on", session: "leader", workflow: withPrefix("", true), want: "release/main-batch"},
		{name: "factory lane, key missing, route on", session: "lane", workflow: withPrefix("", true), want: "release/main-batch"},
		{name: "factory lane, explicit af-flow, route on", session: "lane", workflow: withPrefix("af-flow", true), want: "release/main-batch"},
		{name: "factory lane, explicit github-flow, route on", session: "lane", workflow: withPrefix("github-flow", true)},
		{name: "plain session, key missing, route on", session: "", workflow: withPrefix("", true)},
		{name: "plain session, explicit af-flow, route on", session: "", workflow: withPrefix("af-flow", true), want: "release/main-batch"},
	} {
		t2056Session(t, c.session)
		root, _ := t2056Project(t, c.workflow)
		got := LoadAgenticTrunkReleaseBranch(root)
		if got != c.want {
			t.Errorf("%s: LoadAgenticTrunkReleaseBranch = %q, want %q", c.name, got, c.want)
		}
		// The two readers never disagree about whether the project is af-flow
		// when the route is opted in (card t2098: with the route off the
		// release reader is "" by the gate, not by disagreement).
		isAFF := LoadGitFlowIntegrationConfig(root).Workflow == WorkflowAFFlow
		if c.route && (got != "") != isAFF {
			t.Errorf("%s: the release-branch reader and the workflow reader disagree (workflow af-flow=%v, release branch=%q)", c.name, isAFF, got)
		}
	}
}

// TestT2056_TheUserFileIsNeverRewritten — resolving the default at read time
// leaves git-strategy.yaml byte-identical.
func TestT2056_TheUserFileIsNeverRewritten(t *testing.T) {
	t2056Session(t, "lane")
	root, path := t2056Project(t, t2056KeyMissing)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = LoadGitFlowIntegrationConfig(root)
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("git-strategy.yaml changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
