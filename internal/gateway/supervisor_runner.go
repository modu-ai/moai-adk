package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

type identityProbe func(int) (string, homestate.ProcessIdentityState)

func RunChild(ctx context.Context, cfg ChildConfig, handler http.Handler, handoff io.Writer) error {
	return runChild(ctx, cfg, handler, handoff, homestate.ProbeProcessIdentity)
}

// RunChildWithControl accepts one shutdown byte on the private stdin pipe. EOF
// merely retires the control reader; exec'd parents are watched by identity.
func RunChildWithControl(ctx context.Context, cfg ChildConfig, handler http.Handler, handoff io.Writer, control io.ReadCloser) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	// @MX:WARN: [AUTO] Control read is unblocked by closing the owned pipe at shutdown.
	// @MX:REASON: The deferred join prevents an abandoned blocking reader.
	go func() {
		defer close(done)
		var b [1]byte
		if n, _ := control.Read(b[:]); n == 1 && b[0] == 1 {
			cancel()
		}
	}()
	err := RunChild(ctx, cfg, handler, handoff)
	_ = control.Close()
	<-done
	return err
}

func runChild(ctx context.Context, cfg ChildConfig, handler http.Handler, handoff io.Writer, probe identityProbe) error {
	return runChildWithListener(ctx, cfg, handler, handoff, probe, net.Listen)
}

func runChildWithListener(ctx context.Context, cfg ChildConfig, handler http.Handler, handoff io.Writer, probe identityProbe, listen func(string, string) (net.Listener, error)) error {
	if err := validateChildConfig(cfg); err != nil {
		return err
	}
	if handler == nil || handoff == nil {
		return errChildConfig
	}
	fingerprint, state := probe(cfg.ParentPID)
	if state != homestate.ProcessIdentityLive || fingerprint != cfg.ParentFingerprint {
		return errors.New("gateway parent identity unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	listener, err := listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return errors.New("gateway loopback bind failed")
	}
	defer listener.Close()
	var overlay *ownedOverlay
	if len(cfg.Overlay) > 0 {
		overlay, err = newOwnedOverlay(cfg.Overlay)
		if err != nil {
			return err
		}
		defer overlay.cleanup()
	}
	h := ChildHandoff{Address: listener.Addr().String()}
	if overlay != nil {
		h.OverlayPath = overlay.path
	}
	if err := json.NewEncoder(handoff).Encode(h); err != nil {
		return errors.New("gateway port handoff failed")
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	served := make(chan error, 1)
	// @MX:WARN: [AUTO] HTTP serve is always closed and joined on every monitor exit.
	// @MX:REASON: Parent death and lifetime expiry must free the listener and active connections.
	go func() { served <- server.Serve(listener) }()
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()
	lifetime := time.NewTimer(cfg.Lifetime)
	defer lifetime.Stop()
	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			stop = true
		case <-lifetime.C:
			stop = true
		case <-ticker.C:
			fp, s := probe(cfg.ParentPID)
			stop = s == homestate.ProcessIdentityDead || s == homestate.ProcessIdentityLive && fp != cfg.ParentFingerprint
		case err := <-served:
			_ = server.Close()
			if overlay != nil {
				_ = overlay.cleanup()
			}
			if err != http.ErrServerClosed {
				return errors.New("gateway server stopped unexpectedly")
			}
			return nil
		}
	}
	_ = server.Close()
	<-served
	if overlay != nil {
		return overlay.cleanup()
	}
	return nil
}

// ownedOverlay contains a child-created, nonsecret file in a private 0700
// directory. Callers cannot nominate deletion paths. Root anchoring prevents a
// substituted directory symlink from redirecting file cleanup. SameFile refuses
// a replaced file. Concurrent malicious mutation by the same UID is outside this
// exclusive-directory ownership contract; no atomic compare-and-unlink is claimed.
type ownedOverlay struct {
	path string
	root *os.Root
	info os.FileInfo
	once sync.Once
	err  error
}

func newOwnedOverlay(content []byte) (*ownedOverlay, error) {
	if !json.Valid(content) {
		return nil, errors.New("gateway overlay is invalid JSON")
	}
	dir, err := os.MkdirTemp("", "moai-gateway-overlay-")
	if err != nil {
		return nil, errors.New("gateway overlay directory failed")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		_ = os.Remove(dir)
		return nil, errors.New("gateway overlay root failed")
	}
	file, err := root.OpenFile("settings.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		_ = root.Close()
		_ = os.Remove(dir)
		return nil, errors.New("gateway overlay file failed")
	}
	_, writeErr := file.Write(content)
	info, statErr := file.Stat()
	closeErr := file.Close()
	if writeErr != nil || statErr != nil || closeErr != nil {
		_ = root.Remove("settings.json")
		_ = root.Close()
		_ = os.Remove(dir)
		return nil, errors.New("gateway overlay write failed")
	}
	return &ownedOverlay{path: filepath.Join(dir, "settings.json"), root: root, info: info}, nil
}
func (o *ownedOverlay) cleanup() error {
	o.once.Do(func() {
		defer o.root.Close()
		directory, dirErr := o.root.Stat(".")
		pathInfo, pathErr := os.Lstat(filepath.Dir(o.path))
		if dirErr != nil || pathErr != nil || !pathInfo.IsDir() || !os.SameFile(directory, pathInfo) {
			o.err = errors.New("gateway overlay directory ownership changed")
			return
		}
		info, err := o.root.Lstat("settings.json")
		if err != nil || !info.Mode().IsRegular() || !os.SameFile(info, o.info) {
			o.err = errors.New("gateway overlay ownership changed")
			return
		}
		if err := o.root.Remove("settings.json"); err != nil {
			o.err = errors.New("gateway overlay removal failed")
			return
		}
		_ = o.root.Close()
		if err := os.Remove(filepath.Dir(o.path)); err != nil {
			o.err = errors.New("gateway overlay directory removal failed")
		}
	})
	return o.err
}
