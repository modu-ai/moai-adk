package hook

import (
	"bytes"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// laggingInfo reports a smaller size than the real file, the way a path-based
// read can on Windows while other handles are still writing the file.
type laggingInfo struct {
	fs.FileInfo
	size int64
}

func (l laggingInfo) Size() int64 { return l.size }

// R228: the cap decision must come from the size the open append handle
// reports, not from a path-based read that can lag. With the path-based size
// forced to 0 and the real file at the cap, the append must still rotate.
func TestInboxCap_RotatesWhenPathSizeLagsBehindHandle(t *testing.T) {
	root := t.TempDir()
	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)) // exactly the cap

	orig := inboxPathStat
	inboxPathStat = func(name string) (fs.FileInfo, error) {
		info, err := os.Stat(name)
		if err != nil {
			return nil, err
		}
		return laggingInfo{FileInfo: info, size: 0}, nil
	}
	t.Cleanup(func() { inboxPathStat = orig })

	appendLessonsInboxStub(root, "tool_failure:Bash:ExitError", "lagging path size", "tool:Bash")

	if got := archiveGens(t, root); got != 1 {
		t.Fatalf("archive generations = %d, want 1: the handle-reported size is at the cap, so the append must rotate", got)
	}
}

func captureDebugLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(orig) })
	return &buf
}

// R228 observability: a cap check that returns early says why, so an absent
// rotation is never silent.
func TestInboxCap_EarlyReturnBelowCapIsLogged(t *testing.T) {
	root := t.TempDir()
	seedInbox(t, root, 2)
	buf := captureDebugLog(t)

	appendLessonsInboxStub(root, "k", "below cap", "tool:Bash")

	if !strings.Contains(buf.String(), "below cap") {
		t.Errorf("below-cap early return left no log line; log = %q", buf.String())
	}
}

func TestInboxCap_EarlyReturnStandDownIsLogged(t *testing.T) {
	root := t.TempDir()
	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+1)
	if err := os.MkdirAll(lselStateDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	buf := captureDebugLog(t)

	appendLessonsInboxStub(root, "k", "stand down", "tool:Bash")

	if !strings.Contains(buf.String(), "stand-down") {
		t.Errorf("curator stand-down early return left no log line; log = %q", buf.String())
	}
	if got := archiveGens(t, root); got != 0 {
		t.Errorf("stand-down must not rotate, got %d generations", got)
	}
}
