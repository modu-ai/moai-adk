package cli

// web_agentic_factory.go — connects the Web Console's Agentic Factory panels to
// the claim store (Issue #1948).
//
// The console only reads. It builds the two sources that need the claim store
// alone — the background intake requests and a per-Issue context ledger — and
// only when the project has set Agentic Factory up:
//
//   - no agentic-factory.yaml, or enabled absent/false: not set up. The sources
//     stay nil and the panels say "not provided"; nothing is shown to a project
//     that never asked for the feature.
//   - the file is set up but cannot be used (invalid, or its claim store
//     directory is not there or not a directory): the sources exist and return
//     the reason as their error, so the panels show why instead of nothing.
//   - set up and usable: the sources read the store named by
//     workflow.agentic_factory.claim_store.dir. A relative directory is relative
//     to the project root. The directory is only opened when it already exists:
//     claim.Open would create it, and a console that creates the store it was
//     asked to read hides a wrong path.
//
// MonitorRows is NOT wired here. The monitor collector needs a remote reader
// (GitHub reads) and a ledger reader, and the console has neither; building a
// GitHub reader here would make every page refresh a network call. It stays nil
// and the panels that depend on it say "not provided".

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/web"
)

// agenticFactoryWebSources is what the console receives from the project's
// Agentic Factory settings. A nil field means "not provided".
type agenticFactoryWebSources struct {
	IntakeRequests func(context.Context) ([]byte, error)
	ContextModel   func(context.Context, int) ([]byte, error)
}

// agenticFactoryWebSourcesFor reads projectRoot's settings and returns the
// console's sources. It never writes and never creates the claim store.
func agenticFactoryWebSourcesFor(projectRoot string) agenticFactoryWebSources {
	res := config.AgenticFactoryConfigured(projectRoot)
	if !res.Configured {
		if res.Err == nil {
			return agenticFactoryWebSources{} // not set up: the panels stay "not provided"
		}
		return agenticFactoryWebSourcesFailing(res.Reason)
	}
	s, err := config.LoadAgenticFactorySettings(projectRoot)
	if err != nil {
		return agenticFactoryWebSourcesFailing(err.Error())
	}
	dir := s.ClaimStoreDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(projectRoot, dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return agenticFactoryWebSourcesFailing(fmt.Sprintf(
			"claim store directory %q (workflow.agentic_factory.claim_store.dir) is not an existing directory; the console does not create it", dir))
	}
	st, err := claim.Open(dir)
	if err != nil {
		return agenticFactoryWebSourcesFailing(fmt.Sprintf("claim store directory %q cannot be opened: %v", dir, err))
	}
	return agenticFactoryWebSources{
		IntakeRequests: web.IntakeRequestsFrom(st),
		ContextModel:   web.ContextModelFrom(st),
	}
}

// agenticFactoryWebSourcesFailing returns sources that report why Agentic
// Factory cannot be shown, so the panels carry the reason.
func agenticFactoryWebSourcesFailing(reason string) agenticFactoryWebSources {
	err := errors.New(reason)
	return agenticFactoryWebSources{
		IntakeRequests: func(context.Context) ([]byte, error) { return nil, err },
		ContextModel:   func(context.Context, int) ([]byte, error) { return nil, err },
	}
}
