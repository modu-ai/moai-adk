package cli

// migrate_ownership_env.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M7 (card t1865),
// REQ-OWR-001 and REQ-OWR-011: the pieces the migration and the `--dry-run`
// preview share so that both obtain a file's removal verdict from the one verdict
// function with the same inputs.
//
// Everything here reads. The environment builder itself (newOwnershipEnv) lives in
// migrate_project_assets.go next to the single launch-ledger read the update has
// always made; this file only adapts the two manifest views and answers the
// preview's one question. A preview that cannot assemble the environment (no
// home, an unreadable user manifest) announces no candidate: it never guesses a
// removal.

import (
	"path"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// manifestView is a project manifest as the verdict reads it: an entry lookup and
// the keys indexed by directory (for the letter-case alias).
type manifestView struct {
	entry     func(rel string) (*manifest.FileEntry, bool)
	keysByDir map[string][]string
}

// viewOfManager is the migration's view: the run's manifest manager.
func viewOfManager(mgr manifest.Manager) manifestView {
	return manifestView{entry: mgr.GetEntry, keysByDir: manifestKeysByDir(mgr)}
}

// viewOfManifest is the preview's view: a manifest loaded read-only. A nil
// manifest (absent or corrupt) is an empty one.
func viewOfManifest(mf *manifest.Manifest) manifestView {
	files := map[string]manifest.FileEntry{}
	if mf != nil && mf.Files != nil {
		files = mf.Files
	}
	byDir := make(map[string][]string, len(files))
	for k := range files {
		d := path.Dir(k)
		byDir[d] = append(byDir[d], k)
	}
	return manifestView{
		entry: func(rel string) (*manifest.FileEntry, bool) {
			f, ok := files[rel]
			if !ok {
				return nil, false
			}
			cp := f
			return &cp, true
		},
		keysByDir: byDir,
	}
}

// previewVerdicts answers the `--dry-run` preview's question — would the
// migration remove this file? — by asking the verdict function. The environment
// is assembled lazily on the first question, so a preview that never asks (the
// removal switch is off) reads nothing of the user's home.
type previewVerdicts struct {
	built bool
	ok    bool
	env   ownershipEnv
}

// removalEligible reports whether the verdict function marks rel (a
// project-relative slash path) removal-eligible for the project at projectRoot,
// judged from mf, the manifest as the preview loaded it read-only.
func (p *previewVerdicts) removalEligible(projectRoot string, mf *manifest.Manifest, rel string) bool {
	if !p.built {
		p.built = true
		p.env, p.ok = buildPreviewEnv(projectRoot, mf)
	}
	if !p.ok {
		return false
	}
	return classifyOwnershipFn(p.env, rel, filepath.Join(projectRoot, filepath.FromSlash(rel))).removalEligible
}

// buildPreviewEnv assembles the verdict environment the migration would use, from
// read-only sources. It reports false when any input cannot be read.
func buildPreviewEnv(projectRoot string, mf *manifest.Manifest) (ownershipEnv, bool) {
	homeDir, err := userHomeDirFn()
	if err != nil {
		return ownershipEnv{}, false
	}
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return ownershipEnv{}, false
	}
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		return ownershipEnv{}, false
	}
	userManifest, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		return ownershipEnv{}, false
	}
	env, err := newOwnershipEnv(ownershipEnvInputs{
		projectRoot: projectRoot, homeDir: homeDir, embedded: embedded, cat: cat,
		userManifest: userManifest, view: viewOfManifest(mf), evidence: newProvenanceSource(projectRoot),
	})
	if err != nil {
		return ownershipEnv{}, false
	}
	return env, true
}
