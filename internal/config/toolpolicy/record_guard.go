package toolpolicy

import (
	"errors"
	"os"
	"strings"
	"time"
)

var errRecordGuardBusy = errors.New("record guard held by another process")

// recordGuard is persistent: unlinking it on release would give two processes
// independent kernel locks on old and new guard inodes. Its descriptor remains
// open for the primary record's lifetime; the kernel lock covers mutations only.
type recordGuard struct {
	file *os.File
	path string
	id   string
}

func (g *recordGuard) verify() error {
	id, err := inspectRecordGuard(g.file, g.path)
	if err != nil {
		return err
	}
	if id != g.id {
		return errors.New("persistent record guard identity changed")
	}
	return nil
}

func (g *recordGuard) lockUntil(deadline time.Time) error {
	for {
		err := tryRecordGuardLock(g.file)
		if err == nil {
			if err := g.verify(); err != nil {
				return errors.Join(err, unlockRecordGuard(g.file))
			}
			return nil
		}
		if !errors.Is(err, errRecordGuardBusy) || !time.Now().Before(deadline) {
			return err
		}
		time.Sleep(recordLockRetryInterval)
	}
}

// Guard bindings are a protocol-generation identity, not an audit receipt.
// Legacy pid-only stamps retain their existing stale/liveness interpretation.
func checkRecordGuardBinding(body []byte, id string) error {
	seen := false
	for _, field := range strings.Fields(string(body)) {
		if len(field) < len("guard=") || field[:len("guard=")] != "guard=" {
			continue
		}
		if seen || field[len("guard="):] != id {
			return errors.New("record lock belongs to a different persistent guard")
		}
		seen = true
	}
	return nil
}
