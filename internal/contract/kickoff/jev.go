package kickoff

import (
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
)

// defaultJev is the production Jev seam: the single internal/jev client with
// the stored credential. A disabled gate constructs no request.
func defaultJev(enabled bool) JevAsker {
	c := jev.New(enabled)
	c.LoadCredential = jevcred.Load
	return c.Ask
}
