package cli

// managed_operator_input.go — SPEC-FACTORY-MANAGED-CARD-CHILD-001: the lane
// loop's managed card-child launch and the operator-input pump the successive
// managed sessions share. M1 INERT DECLARATIONS: the seam, the source, the
// creation counter and the pump type compile and do nothing yet; M2/M3 give
// them behavior.

import (
	"io"
	"os"
	"sync/atomic"
)

// managedLaneOperatorSource is the operator-input source the lane loop's pump
// reads (default: the process stdin). Tests point it at their own reader.
var managedLaneOperatorSource io.Reader = os.Stdin

// managedOperatorPumpsCreated counts pumps built, so a test can show the
// switch-off path builds none.
var managedOperatorPumpsCreated atomic.Int64

// defaultManagedCodexCardLaunch is the default body of the lane loop's managed
// launch seam. Inert until M3.
func defaultManagedCodexCardLaunch(bin string, args, env []string, dir string) error {
	return nil
}

// managedOperatorPump is the lane loop's one reader of the operator input.
type managedOperatorPump struct{}

// newManagedOperatorPump builds a pump over src. Inert until M3.
func newManagedOperatorPump(src io.Reader) *managedOperatorPump {
	managedOperatorPumpsCreated.Add(1)
	return &managedOperatorPump{}
}

// attach returns the reader one managed session reads its operator input
// from. Inert until M3: it reports end of input at once.
func (p *managedOperatorPump) attach() io.ReadCloser {
	return io.NopCloser(eofReader{})
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) { return 0, io.EOF }
