package backup

import "os"

// openRunMarker opens the run marker. Windows has no FIFO at a path, so a plain open is
// enough; readRunMarker's regular-file check still applies.
func openRunMarker(path string) (*os.File, error) {
	return os.Open(path)
}
