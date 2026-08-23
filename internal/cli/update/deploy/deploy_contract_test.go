package deploy_test

// Cross-contract test between the clean set and the deployment set
// (SPEC-CLI-CLEAN-SYMLINK-001, AC-CSL-009 / REQ-CSL-009). External test
// package: the contract reads the embedded template filesystem, and keeping
// that import in the _test package preserves deploy's leaf property for
// production code.
//
// "Held" is judged against the RENDERED destination set — the paths the
// deployer records after rendering (directory entries skipped, ".tmpl"
// suffix stripped; internal/template/deployer.go "Files ending in .tmpl are
// rendered and saved without the suffix") — NOT the raw embedded paths.
// Root 1 is the discriminator: the template holds settings.json.tmpl only,
// so a raw-path reading of the contract would false-fail on it (dossier
// §1.1(b)).

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// renderedDestinations walks the embedded template filesystem the way the
// deployer does and returns the set of slash-relative paths deployment
// records after rendering.
func renderedDestinations(t *testing.T) map[string]bool {
	t.Helper()
	tmplFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	dest := make(map[string]bool)
	err = fs.WalkDir(tmplFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dest[strings.TrimSuffix(filepath.ToSlash(p), ".tmpl")] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded templates: %v", err)
	}
	if len(dest) == 0 {
		t.Fatal("embedded template filesystem carries no files — contract cannot be evaluated")
	}
	return dest
}

// TestCleanSetCoveredByDeployment is AC-CSL-009: every non-glob clean root
// (the seven ManagedCleanTargets entries plus the eighth .moai/config root)
// must be a rendered-destination path — either an exact file or a directory
// holding at least one recorded file — and every glob clean pattern must
// match at least one recorded path. A violating root means clean deletes a
// user path deployment never rewrites (the silent-wipe defect class this
// contract exists to prevent).
func TestCleanSetCoveredByDeployment(t *testing.T) {
	dest := renderedDestinations(t)

	// The eighth removal root is handled inside CleanMoaiManagedPaths, not
	// listed in ManagedCleanTargets — the contract covers it all the same.
	roots := []string{filepath.ToSlash(filepath.Join(defs.MoAIDir, defs.ConfigSubdir))}
	var globPatterns []string
	for _, target := range deploy.ManagedCleanTargets(t.TempDir()) {
		rel := filepath.ToSlash(target.DisplayPath)
		if target.IsGlob {
			globPatterns = append(globPatterns, rel)
			continue
		}
		roots = append(roots, rel)
	}

	for _, root := range roots {
		if dest[root] {
			continue // exact recorded file (e.g. .claude/settings.json via .tmpl)
		}
		prefix := root + "/"
		for p := range dest {
			if strings.HasPrefix(p, prefix) {
				goto covered
			}
		}
		t.Errorf("clean root %s is not covered by deployment: no rendered destination at or under it (clean would delete a path nothing rewrites)", root)
	covered:
	}

	for _, pattern := range globPatterns {
		dir := path.Dir(pattern)
		namePat := path.Base(pattern)
		matched := false
		for p := range dest {
			if !strings.HasPrefix(p, dir+"/") {
				continue
			}
			seg := p[len(dir)+1:]
			if i := strings.Index(seg, "/"); i >= 0 {
				seg = seg[:i]
			}
			if ok, err := path.Match(namePat, seg); err == nil && ok {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("clean glob %s matches no rendered destination path (every match it removes on disk would be unrecoverable)", pattern)
		}
	}
}
