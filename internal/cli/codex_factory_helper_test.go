package cli

import (
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// NextFactoryWorkerNumberForTest delegates to the kanban SSOT.
func NextFactoryWorkerNumberForTest(reg map[string]kanban.FactoryWorkerEntry, alive func(int) bool) int {
	return kanban.NextFactoryWorkerNumber(reg, alive)
}
