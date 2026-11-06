package cli

import "os"

// openJevCostLedger opens the decision ledger. Windows has no FIFO at a path,
// so a plain open is enough; readJevCostLedger's regular-file check still
// applies (the directory case of TestJevCost_DirectoryAtLedgerPath pins it
// here).
func openJevCostLedger(path string) (*os.File, error) {
	return os.Open(path)
}
