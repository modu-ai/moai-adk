package cli

import (
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
// version-only decision. A tree that cannot be read or measured does not skip,
// so the run syncs.
func versionMatchSkipsSync(projectRoot string) bool {
	m, err := manifest.NewManager().Load(projectRoot)
	if err != nil {
		return false
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
