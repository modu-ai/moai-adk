package claim

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// syncer is the fsync step of a durable write. The zero value syncs for real;
// tests set file or dir to inject a failure.
type syncer struct {
	file func(*os.File) error
	dir  func(string) error
}

func (y syncer) syncFile(f *os.File) error {
	if y.file != nil {
		return y.file(f)
	}
	return f.Sync()
}

func (y syncer) syncDir(path string) error {
	if y.dir != nil {
		return y.dir(path)
	}
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	err = d.Sync()
	if cerr := d.Close(); err == nil {
		err = cerr
	}
	return err
}

// writeAtomic writes b to dst so that a success means the bytes are on disk:
// temp file, fsync of the temp file, rename, fsync of the directory. Any failure
// is returned (the caller must not acknowledge the write); a failure before the
// rename leaves dst as it was and removes the temp file. A power or OS crash can
// otherwise leave a renamed but empty file behind an acknowledged event.
func (s *Store) writeAtomic(dir, pattern, dst string, b []byte) error {
	fail := func(err error) error { return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err) }
	tmp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return fail(err)
	}
	drop := func(err error) error {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return fail(err)
	}
	if _, err := tmp.Write(b); err != nil {
		return drop(err)
	}
	if err := s.sync.syncFile(tmp); err != nil {
		return drop(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return fail(err)
	}
	if err := os.Rename(tmp.Name(), dst); err != nil {
		_ = os.Remove(tmp.Name())
		return fail(err)
	}
	if err := s.sync.syncDir(dir); dirSyncFailureIsFatal(runtime.GOOS, err) {
		return fail(err)
	}
	return nil
}

// dirSyncFailureIsFatal reports whether a directory fsync error must fail the write.
// Windows cannot open a directory to sync it ("Access is denied"), so there the
// refusal is best effort, the same posture the ledger snapshot takes; any other
// platform or any other error keeps the fsync a hard durability step (card t2119).
func dirSyncFailureIsFatal(goos string, err error) bool {
	if err == nil {
		return false
	}
	return goos != "windows" || !errors.Is(err, fs.ErrPermission)
}
