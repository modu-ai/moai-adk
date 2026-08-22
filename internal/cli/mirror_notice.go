package cli

// Shared consumption of the deployer's skill-mirror result for the two
// `moai update` paths. The init path consumes the same result in
// internal/core/project, where the notice goes into InitResult.Warnings
// instead of a writer.

import (
	"context"
	"io"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/mirrornotice"
	"github.com/modu-ai/moai-adk/internal/template"
)

// deployWithMirrorNotice deploys and writes the skill-mirror notice, if any, to
// errOut.
//
// The result extension is optional by design, so a Deployer that does not
// implement it deploys exactly as before and produces no notice — the type
// assertion, not a widened interface, is what keeps that true.
//
// errOut is passed explicitly rather than reusing whatever writer the caller
// already holds: both update paths write progress to stdout, and the notice is
// a warning, which internal/cli/CLAUDE.md:14 assigns to stderr ("Never mix").
func deployWithMirrorNotice(
	ctx context.Context,
	dep template.Deployer,
	projectRoot string,
	mgr manifest.Manager,
	tmplCtx *template.TemplateContext,
	errOut io.Writer,
) error {
	rd, ok := dep.(template.ResultDeployer)
	if !ok {
		return dep.Deploy(ctx, projectRoot, mgr, tmplCtx)
	}

	// The result is populated even when deployment errors, so the notice is
	// written before the error is returned.
	res, err := rd.DeployWithResult(ctx, projectRoot, mgr, tmplCtx)
	mirrornotice.WriteTo(errOut, res)
	return err
}
