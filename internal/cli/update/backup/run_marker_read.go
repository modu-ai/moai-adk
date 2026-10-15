package backup

import "io"

// runMarkerReadLimit bounds the marker read. A run marker is one run-id line, so the bound
// keeps a pathological file from being read whole.
const runMarkerReadLimit = 4096

// readRunMarker returns the bytes of the run marker at path. ok is false for an absent path,
// a non-regular object, or a read error; callers treat !ok as "no marker". The regular-file
// check is made on the handle the bytes are read from. Card t1619 (codex review P2): a FIFO
// named .moai-run must not hang the update's run collection, which reads every backup
// directory before it checks the directory name.
func readRunMarker(path string) (data []byte, ok bool) {
	f, err := openRunMarker(path)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, false
	}
	data, err = io.ReadAll(io.LimitReader(f, runMarkerReadLimit))
	if err != nil {
		return nil, false
	}
	return data, true
}
