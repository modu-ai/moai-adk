package bugreport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

// The payload's fields are validated against anchored allowlists on the way
// in (Build) and on every read-back (ParsePayload): a queued payload is
// untrusted input because its store is a local file.
var (
	fingerprintPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)
	versionPattern     = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.\-]+)?$`)
	commitPattern      = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
)

// knownGOOS and knownGOARCH are the Go platform sets (go tool dist list,
// current set) the os and arch fields must be members of.
var knownGOOS = map[string]bool{
	"aix": true, "android": true, "darwin": true, "dragonfly": true,
	"freebsd": true, "illumos": true, "ios": true, "js": true,
	"linux": true, "netbsd": true, "openbsd": true, "plan9": true,
	"solaris": true, "wasip1": true, "windows": true,
}

var knownGOARCH = map[string]bool{
	"386": true, "amd64": true, "amd64p32": true, "arm": true,
	"arm64": true, "arm64be": true, "armbe": true, "loong64": true,
	"mips": true, "mips64": true, "mips64le": true, "mips64p32": true,
	"mips64p32le": true, "mipsle": true, "ppc": true, "ppc64": true,
	"ppc64le": true, "riscv": true, "riscv64": true, "s390": true,
	"s390x": true, "sparc": true, "sparc64": true, "wasm": true,
}

// ValidatePayload applies the anchored allowlists to every field of a
// payload, including the detail (closed-set membership via ParseDetail).
func ValidatePayload(p Payload) error {
	if p.Schema != SchemaV1 {
		return fmt.Errorf("bugreport: payload schema %q, want %q", p.Schema, SchemaV1)
	}
	if !p.Kind.Valid() {
		return fmt.Errorf("bugreport: payload kind %q is not one of the six kinds", string(p.Kind))
	}
	if !fingerprintPattern.MatchString(p.Fingerprint) {
		return fmt.Errorf("bugreport: payload fingerprint %q fails the 16-hex allowlist", p.Fingerprint)
	}
	if !versionPattern.MatchString(p.Version) {
		return fmt.Errorf("bugreport: payload version %q fails the semver allowlist", p.Version)
	}
	if !commitPattern.MatchString(p.Commit) {
		return fmt.Errorf("bugreport: payload commit %q fails the 7-40 hex allowlist", p.Commit)
	}
	if !knownGOOS[p.OS] {
		return fmt.Errorf("bugreport: payload os %q is not a Go platform", p.OS)
	}
	if !knownGOARCH[p.Arch] {
		return fmt.Errorf("bugreport: payload arch %q is not a Go architecture", p.Arch)
	}
	if len(p.Frames) == 0 {
		return fmt.Errorf("bugreport: payload carries no frames")
	}
	if len(p.Frames) > FrameLimit() {
		return fmt.Errorf("bugreport: payload carries %d frames, over the %d limit", len(p.Frames), FrameLimit())
	}
	for _, f := range p.Frames {
		if err := ValidateFrameName(f); err != nil {
			return err
		}
	}
	if p.Detail != nil {
		if err := validateDetailForKind(p.Kind, p.Detail); err != nil {
			return err
		}
	}
	return nil
}

// validateDetailForKind re-checks an in-memory Detail against its kind: the
// concrete carrier must be the one the kind's register row admits and its
// token form must pass ParseDetail, so a Detail built by hand outside the
// constructors cannot ride a payload.
func validateDetailForKind(kind Kind, d Detail) error {
	switch dd := d.(type) {
	case TemplateToken:
		if !dd.ValidForKind(kind) {
			return fmt.Errorf("%w: token %q is not valid for kind %s", ErrDetailRejected, dd, kind)
		}
		return nil
	case HookDetail:
		if kind != KindHookHandlerFailure && kind != KindHookTimeout {
			return fmt.Errorf("%w: hook detail is not valid for kind %s", ErrDetailRejected, kind)
		}
		hookIdentityRegistry.Lock()
		defer hookIdentityRegistry.Unlock()
		if !hookIdentityRegistry.events[dd.EventID] || !hookIdentityRegistry.handlers[dd.HandlerID] {
			return fmt.Errorf("%w: hook detail (%s, %s) is not registered", ErrDetailRejected, dd.EventID, dd.HandlerID)
		}
		return nil
	default:
		return fmt.Errorf("%w: detail carrier %T is not one of the closed types", ErrDetailRejected, d)
	}
}

// payloadWire is the on-disk form: detail is a string there, so read-back
// re-validates it by membership.
type payloadWire struct {
	Schema      string   `json:"schema"`
	Kind        Kind     `json:"kind"`
	Fingerprint string   `json:"fingerprint"`
	Version     string   `json:"version"`
	Commit      string   `json:"commit"`
	OS          string   `json:"os"`
	Arch        string   `json:"arch"`
	Frames      []string `json:"frames"`
	Detail      string   `json:"detail,omitempty"`
}

// ErrPayloadRejected wraps every read-back rejection.
var ErrPayloadRejected = errors.New("bugreport: payload rejected")

// ParsePayload decodes and re-validates a queued payload (REQ-ANON-011):
// unknown fields are rejected, every field must pass its anchored allowlist,
// the detail string must be a member of the closed set for the kind — and
// the input must be EXACTLY one value. json.Decoder binds only the first
// value, so trailing data (a second object, an array, plain junk) would
// otherwise ride in behind a valid payload; the second decode must reach
// EOF or the payload is rejected.
func ParsePayload(data []byte) (Payload, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var w payloadWire
	if err := dec.Decode(&w); err != nil {
		return Payload{}, fmt.Errorf("%w: decode: %v", ErrPayloadRejected, err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return Payload{}, fmt.Errorf("%w: trailing data after the payload", ErrPayloadRejected)
	}
	p := Payload{
		Schema:      w.Schema,
		Kind:        w.Kind,
		Fingerprint: w.Fingerprint,
		Version:     w.Version,
		Commit:      w.Commit,
		OS:          w.OS,
		Arch:        w.Arch,
		Frames:      w.Frames,
	}
	if w.Detail != "" {
		d, err := ParseDetail(p.Kind, w.Detail)
		if err != nil {
			return Payload{}, err
		}
		p.Detail = d
	}
	if err := ValidatePayload(p); err != nil {
		return Payload{}, err
	}
	return p, nil
}
