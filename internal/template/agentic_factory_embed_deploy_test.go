// agentic_factory_embed_deploy_test.go: the Agentic Factory templates reach the
// embedded filesystem byte-for-byte, and a fresh-install deploy writes them
// without touching files a user already owns.
//
// Scope. The deploy case drives the real Deployer over the real embedded
// files in a t.TempDir() project, in the non-force (init) mode. The `moai
// update` force-update/merge path is NOT exercised here; isolated install and
// update tests belong to the integration lane and are recorded as a gap.
//
// Sentinels on failure: AF_EMBED / AF_DEPLOY.
package template_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// TestAgenticFactoryEmbedded: every shipped file is present in the embedded
// FS with the same bytes as the source tree, and the deployer would write it.
func TestAgenticFactoryEmbedded(t *testing.T) {
	t.Parallel()
	emb, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("AF_EMBED: EmbeddedTemplates: %v", err)
	}
	listed := map[string]bool{}
	for _, p := range template.NewDeployer(emb).ListTemplates() {
		listed[p] = true
	}
	for _, rel := range afExpectedFiles() {
		got, err := fs.ReadFile(emb, rel)
		if err != nil {
			t.Errorf("AF_EMBED: %s is not in the embedded FS (regenerate the embed manifest): %v", rel, err)
			continue
		}
		want, err := os.ReadFile(filepath.Join(afTemplatesRoot, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("AF_EMBED: read source %s: %v", rel, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("AF_EMBED: %s differs between the embedded FS and the source tree", rel)
		}
		if !listed[rel] {
			t.Errorf("AF_EMBED: deployer listing omits %s", rel)
		}
	}
	// the issue-chooser config is deliberately not shipped
	if _, err := fs.Stat(emb, ".github/ISSUE_TEMPLATE/config.yml"); err == nil {
		t.Errorf("AF_EMBED: .github/ISSUE_TEMPLATE/config.yml must not be embedded")
	}
}

// afFeatureSubset copies the feature files out of the embedded FS.
func afFeatureSubset(t *testing.T) fstest.MapFS {
	t.Helper()
	emb, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("AF_DEPLOY: EmbeddedTemplates: %v", err)
	}
	sub := fstest.MapFS{}
	for _, rel := range afExpectedFiles() {
		data, err := fs.ReadFile(emb, rel)
		if err != nil {
			t.Fatalf("AF_DEPLOY: %s missing from the embedded FS: %v", rel, err)
		}
		sub[rel] = &fstest.MapFile{Data: data}
	}
	return sub
}

// TestAgenticFactoryDeployNonDestructive: a fresh-install deploy creates every
// feature file, leaves a pre-existing user copy of one of them byte-identical
// (recorded user_created), leaves the user's own default pull request
// template alone, and is idempotent.
func TestAgenticFactoryDeployNonDestructive(t *testing.T) {
	t.Parallel()
	sub := afFeatureSubset(t)
	// the shipped labels.yml rides along so the user's own copy is a real
	// collision, not a vacuous one
	emb, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("AF_DEPLOY: EmbeddedTemplates: %v", err)
	}
	shippedLabels, err := fs.ReadFile(emb, ".github/labels.yml")
	if err != nil {
		t.Fatalf("AF_DEPLOY: read shipped labels.yml: %v", err)
	}
	sub[".github/labels.yml"] = &fstest.MapFile{Data: shippedLabels}

	root := t.TempDir()
	mustWrite := func(rel, content string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	userForm := ".github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml"
	userPR := ".github/PULL_REQUEST_TEMPLATE.md"
	userLabels := ".github/labels.yml"
	mustWrite(userForm, "name: my own customised form\n")
	mustWrite(userPR, "my team's default pull request template\n")
	mustWrite(userLabels, "- name: my-label\n  color: \"ffffff\"\n")

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("AF_DEPLOY: manifest load: %v", err)
	}
	d := template.NewDeployer(sub)
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("AF_DEPLOY: Deploy: %v", err)
	}

	for _, rel := range afExpectedFiles() {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("AF_DEPLOY: %s was not created: %v", rel, err)
		}
	}
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("AF_DEPLOY: read %s: %v", rel, err)
		}
		return string(b)
	}
	if got := read(userForm); got != "name: my own customised form\n" {
		t.Errorf("AF_DEPLOY: the user's form was overwritten: %q", got)
	}
	if got := read(userPR); got != "my team's default pull request template\n" {
		t.Errorf("AF_DEPLOY: the user's default pull request template was touched: %q", got)
	}
	if got := read(userLabels); got != "- name: my-label\n  color: \"ffffff\"\n" {
		t.Errorf("AF_DEPLOY: the user's labels.yml was touched: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "ISSUE_TEMPLATE", "config.yml")); err == nil {
		t.Errorf("AF_DEPLOY: deploy created an issue-chooser config.yml")
	}

	if e, found := mgr.GetEntry(userForm); !found || e.Provenance != manifest.UserCreated {
		t.Errorf("AF_DEPLOY: %s should be recorded user_created, got found=%v entry=%+v", userForm, found, e)
	}
	for _, a := range afTemplates {
		rel := a.gitlabPath()
		if e, found := mgr.GetEntry(rel); !found || e.Provenance != manifest.TemplateManaged {
			t.Errorf("AF_DEPLOY: %s should be tracked template_managed, got found=%v entry=%+v", rel, found, e)
		}
	}

	// the deployed output (not just the source tree) carries none of the
	// credential/path shapes the leak scanner knows
	for _, rel := range afExpectedFiles() {
		deployed := read(rel)
		if rel == userForm {
			continue // the user's own file, not template output
		}
		for name, re := range afLeakPatterns() {
			if re.MatchString(deployed) {
				t.Errorf("AF_TEMPLATE_LEAK: deployed %s matches the %s shape", rel, name)
			}
		}
	}

	// second deploy: no change to any feature file or user file
	before := map[string]string{}
	for _, rel := range append(afExpectedFiles(), userPR, userLabels) {
		before[rel] = read(rel)
	}
	if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
		t.Fatalf("AF_DEPLOY: second Deploy: %v", err)
	}
	for rel, want := range before {
		if got := read(rel); got != want {
			t.Errorf("AF_DEPLOY: %s changed on the second deploy", rel)
		}
	}
}
