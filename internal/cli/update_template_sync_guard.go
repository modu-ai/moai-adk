package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// embeddedTemplateTreeHash hashes the embedded template tree this binary would
// deploy (SPEC-UPDATE-REPAIR-RC30-001 REQ-RC30-022). A deploy records the value,
// and a version-matched run compares its recorded value against it.
func embeddedTemplateTreeHash() (string, error) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return "", err
	}
	return template.ComputeDirTreeHash(embedded, ".")
}

// recordTemplateTree stores the embedded template tree this deploy wrote on the
// in-memory manifest, so the next version-matched run skips only while that tree
// is still the one shipped (REQ-RC30-022). The caller persists the manifest.
func recordTemplateTree(mgr manifest.Manager, errOut io.Writer) {
	treeHash, err := embeddedTemplateTreeHash()
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "  template tree not recorded: %v\n", err)
		return
	}
	mgr.Manifest().TemplateTreeHash = treeHash
}

// versionMatchSkipsSync reports whether a run whose template_version matches the
// binary may skip the template sync (REQ-RC30-022). The recorded template tree
// must also equal the embedded tree. A project with no recorded tree keeps the
// version-only decision. A tree that cannot be measured does not skip, so the
// run syncs.
//
// A manifest that fails to load is judged by whether the sync can rebuild it. A
// corrupt manifest does not skip: the sync moves it aside and rebuilds it. The
// judgement only reads the file (probeManifest), so it can run any number of times
// and every run reads the same project (card t1619, codex review round 3). Any
// other load failure (a directory or an unreadable file where the manifest
// belongs) skips, because the sync's own load would fail the same way. The
// integrity probe then names the file as a warning, so the update exits 0
// (REQ-UMF-002).
func versionMatchSkipsSync(projectRoot string) bool {
	m, err := probeManifest(projectRoot)
	if err != nil {
		return !errors.Is(err, manifest.ErrManifestCorrupt)
	}
	if m.TemplateTreeHash == "" {
		return true
	}
	current, err := embeddedTemplateTreeHash()
	if err != nil {
		return false
	}
	return m.TemplateTreeHash == current
}

// probeManifest reads the project manifest without changing it. Load moves a corrupt
// manifest aside as a side effect, so a second judgement over the moved file read the
// project as fresh and skipped the rebuild (card t1619, codex review round 3). The
// judgement reads the bytes and leaves the file to the sync stage, which owns the
// quarantine. A path that is not a readable regular file is a load failure, as Load's
// own read of it is.
func probeManifest(projectRoot string) (*manifest.Manifest, error) {
	path := filepath.Join(projectRoot, defs.MoAIDir, defs.ManifestJSON)
	data, ok := readManifestRecord(path)
	if !ok {
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			return manifest.NewManifest(), nil
		}
		return nil, fmt.Errorf("manifest probe: %s is not a readable regular file", path)
	}
	var mf manifest.Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("%w: %v", manifest.ErrManifestCorrupt, err)
	}
	return &mf, nil
}

// warnStampSkew prints a visible warning when the template_version stamp in the
// deployed system.yaml disagrees with the version this run deployed
// (REQ-RC30-023). With no system.yaml there is nothing deployed to compare.
func warnStampSkew(out io.Writer, projectRoot, stamp string) {
	configPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
	if _, err := os.Stat(configPath); err != nil {
		return
	}
	deployed, err := plan.GetProjectConfigVersion(projectRoot)
	if err != nil || deployed == stamp {
		return
	}
	th := resolveTheme()
	_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "template stamp disagrees with deployed bytes",
		fmt.Sprintf("system.yaml template_version %s, this run deployed %s", deployed, stamp), "", &th))
}
