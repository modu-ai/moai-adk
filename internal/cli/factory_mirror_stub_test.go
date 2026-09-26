package cli

// RED-phase contract stub for the dispatch mirror (t1239 M6): it fixes the
// write seam the mirror tests replace. The GREEN commit deletes this file and
// declares the seam with the mirror.

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

var factoryAssignmentWriter = func(context.Context, string, *kanban.BacklogStore, string, string, string) error { return nil }
