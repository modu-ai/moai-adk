//go:build afred

// agentic_factory_update_preserve_afred_test.go: KNOWN DEFECT REPRODUCTION,
// BEFORE ANY REPAIR (RED). Excluded from the default test run by the afred
// build tag; run it with:
//
//	go test -tags afred -run '^TestAgenticFactoryUpdatePreservesUserFiles$' ./internal/template/
//
// Defect. The Deployer in force-update mode overwrites a file recorded
// user_modified or user_created in the manifest when that file sits at a
// template path. Only published-skill paths keep the provenance check
// (deployer.go: `!d.forceUpdate || protectedScope`). Measured on this tree, the
// test fails with these two paths overwritten:
//
//	.github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml  (user_modified)
//	.github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml   (user_created)
//
// A user-only file beside the feature files and the user's default pull
// request template stay byte-identical. The repair (protect these paths in
// force mode) touches deployer.go, which another open PR also changes, so it is
// deferred until that PR lands.
//
// Sentinel on failure: AF_PRESERVE.
package template_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestAgenticFactoryUpdatePreservesUserFiles: after a force update, a file the
// user edited (manifest user_modified), a file the user created at a feature
// path (user_created), and a user-only file next to the feature files stay
// byte-identical. Every path that changed is listed in the failure.
func TestAgenticFactoryUpdatePreservesUserFiles(t *testing.T) {
	t.Parallel()
	sub := afFeatureSubset(t)
	files := afExpectedFiles()
	root := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("AF_PRESERVE: manifest load: %v", err)
	}

	userEdited := files[0]
	userCreated := files[1]
	userOnly := ".github/ISSUE_TEMPLATE/my-team-bug-report.yml"
	userPR := ".github/PULL_REQUEST_TEMPLATE.md"
	content := map[string]string{
		userEdited:  "name: edited by the user\n",
		userCreated: "name: written by the user before install\n",
		userOnly:    "name: my team bug report\n",
		userPR:      "my team's default pull request template\n",
	}
	for rel, c := range content {
		afWriteFile(t, root, rel, c)
	}
	if err := mgr.Track(userEdited, manifest.UserModified, manifest.HashBytes([]byte("name: shipped\n"))); err != nil {
		t.Fatalf("AF_PRESERVE: track: %v", err)
	}
	if err := mgr.Track(userCreated, manifest.UserCreated, manifest.HashBytes([]byte(content[userCreated]))); err != nil {
		t.Fatalf("AF_PRESERVE: track: %v", err)
	}

	if err := template.NewDeployerWithForceUpdate(sub, true).Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("AF_PRESERVE: Deploy(force): %v", err)
	}

	var changed []string
	for rel, want := range content {
		if afReadFile(t, root, rel) != want {
			changed = append(changed, rel)
		}
	}
	sort.Strings(changed)
	if len(changed) > 0 {
		t.Errorf("AF_PRESERVE: update overwrote user files: %s", strings.Join(changed, ", "))
	}
}
