package cli

import (
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// NextFactoryAgentNumberForTest delegates to the kanban SSOT.
func NextFactoryAgentNumberForTest(reg map[string]kanban.FactoryWorkerEntry, alive func(int) bool) int {
	return kanban.NextFactoryAgentNumber(reg, alive)
}
