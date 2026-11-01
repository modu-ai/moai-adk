// factory_wait.go — card t2013: record a lane's wait in the factory record and
// print, once, who to wake when the landing edge satisfies it.
package cli

import (
	"context"
	"io"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryPrintWakes is a stub in the RED commit.
func factoryPrintWakes(ctx context.Context, out io.Writer, db *homestate.FactoryDB, card homestate.Card) {
}
