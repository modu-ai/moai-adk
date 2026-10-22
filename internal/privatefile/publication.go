package privatefile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Publication retains the staging object until a caller commits a multi-file
// update or revokes it. Revoke and Close act on that object, never its old name.
// A Publication has one owner and is not safe for concurrent method calls.
type Publication struct {
	anchor              *os.File
	source              string
	before              os.FileInfo
	closed, revoked     bool
	closeErr, revokeErr error
}

// Name is the original staging pathname; cleanup never relies on it.
func (p *Publication) Name() string { return p.source }

// CapturePublication takes ownership of staging, including on error. The caller
// must own a temporary created by CreateTemp and finish writing and any required
// sync first. An independent cleanup handle captures
// the same private, singly linked object before the original writer is closed;
// a writer-close failure prevents publication. Close releases the cleanup handle.
// @MX:WARN: [AUTO] Keep every staging ownership and cleanup failure on the captured inode.
// @MX:REASON: Identity, native policy, anchor acquisition and writer-close failures must stop publication without pathname cleanup.
func CapturePublication(staging *os.File) (_ *Publication, resultErr error) {
	if staging == nil {
		return nil, errors.New("private publication: nil staging file")
	}
	p := &Publication{anchor: staging, source: staging.Name()}
	writerClosed := false
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, p.Revoke(), p.Close())
		}
		if !writerClosed && p.anchor != staging {
			resultErr = errors.Join(resultErr, staging.Close())
		}
	}()
	before, err := staging.Stat()
	if err != nil {
		return nil, err
	}
	if err := checkPublicationHandleNative(staging); err != nil {
		return nil, err
	}
	held, err := openPublicationAnchorNative(p.source)
	if err != nil {
		return nil, err
	}
	opened, err := held.Stat()
	if err != nil || !os.SameFile(before, opened) {
		return nil, errors.Join(err, errors.New("private file: staging identity changed"), held.Close())
	}
	p.anchor, p.before = held, before
	writerClosed = true
	if err := staging.Close(); err != nil {
		return nil, err
	}
	return p, nil
}

// Install moves this staging inode within its directory and verifies the native
// owner-only policy and identity before and after publication. It retains the
// cleanup handle: the caller must Revoke on a failed transaction, then Close.
// A cross-directory refusal preserves the staging object and destination.
// @MX:WARN: [AUTO] Validate retained staging identity and native policy on both sides of replacement.
// @MX:REASON: Destination shape, native policy, link count and source identity checks are mandatory refusal boundaries.
func (p *Publication) Install(target string) error { return p.install(target, Check) }

func (p *Publication) install(target string, check func(string) error) error {
	if p.closed || p.revoked {
		return errors.New("private publication: ownership already released")
	}
	if filepath.Clean(filepath.Dir(p.source)) != filepath.Clean(filepath.Dir(target)) {
		return errors.New("private publication: staging_directory_mismatch")
	}
	if info, err := os.Lstat(target); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("private file: unsafe destination")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := checkPublicationHandleNative(p.anchor); err != nil {
		return err
	}
	if err := Check(p.source); err != nil {
		return err
	}
	if err := check(p.source); err != nil {
		return err
	}
	before, err := publicationInfoNative(p.source)
	if err != nil || !os.SameFile(p.before, before) {
		return errors.Join(err, errors.New("private file: staging identity changed"))
	}
	if err := atomicfile.Replace(p.source, target); err != nil {
		return err
	}
	if err := Check(target); err != nil {
		return err
	}
	if err := check(target); err != nil {
		return err
	}
	after, err := publicationInfoNative(target)
	if err != nil || !os.SameFile(p.before, after) {
		return errors.Join(err, fmt.Errorf("private file: publication identity changed"))
	}
	return checkPublicationHandleNative(p.anchor)
}

// Revoke invalidates secret bytes on the retained inode, preserving any other
// writer's replacement at its former pathname. Repeated calls return one result.
func (p *Publication) Revoke() error {
	if !p.revoked {
		p.revoked = true
		p.revokeErr = revokePublicationNative(p.anchor)
	}
	return p.revokeErr
}

// Close releases the retained object handle once. It never unlinks a pathname.
func (p *Publication) Close() error {
	if !p.closed {
		p.closed = true
		p.closeErr = p.anchor.Close()
	}
	return p.closeErr
}
