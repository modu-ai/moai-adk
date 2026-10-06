package cli

// migrate_render.go — the production TemplateRender for the migration
// classifier (SPEC-INIT-SHRINK-001 REQ-010, plan M1): the embedded template
// tree for this project's render context. A path is carried when the tree
// holds a plain source at it, or a .tmpl source that renders (with the
// project's TemplateContext when one is wired, raw bytes otherwise).

import (
	"io/fs"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/template"
)

// templateRenderCarriage adapts an embedded-template FS (and optionally a
// renderer + context) to update.TemplateRender. Carriage is decided against
// the DEPLOY-relative path: a plain source at relPath, or relPath+".tmpl"
// whose rendered form is the carried content. A nil renderer reads .tmpl
// sources raw — fixtures that avoid .tmpl paths need no renderer, and the
// update flow's ValidateAll aborts before a migration could run on a
// template whose render fails.
func templateRenderCarriage(fsys fs.FS, renderer template.Renderer, ctx *template.TemplateContext) update.TemplateRender {
	return renderedCarriage{fsys: fsys, renderer: renderer, ctx: ctx}
}

type renderedCarriage struct {
	fsys     fs.FS
	renderer template.Renderer
	ctx      *template.TemplateContext
}

// Carries implements update.TemplateRender.
func (c renderedCarriage) Carries(relPath string) ([]byte, bool) {
	if data, err := fs.ReadFile(c.fsys, relPath); err == nil {
		return data, true
	}
	tmplPath := relPath + ".tmpl"
	data, err := fs.ReadFile(c.fsys, tmplPath)
	if err != nil {
		return nil, false
	}
	if c.renderer != nil && c.ctx != nil {
		rendered, renderErr := c.renderer.Render(tmplPath, c.ctx)
		if renderErr != nil {
			// A render failure is a template defect the update flow's
			// ValidateAll catches before any migration runs; degrade to raw
			// bytes here so classification stays decidable.
			return data, true
		}
		return rendered, true
	}
	return data, true
}
