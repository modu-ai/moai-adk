package bugreport

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
)

// Payload is the machine-generated report (REQ-ANON-011): a closed set of
// fixed fields, no timestamp, no session id, no user name, no host, no path,
// no repository name, and no environment value anywhere. The in-memory Detail
// field holds the sealed closed-set type; the wire form serialises it through
// its Token and re-validates membership on read-back.
type Payload struct {
	Schema      string   `json:"-"`
	Kind        Kind     `json:"-"`
	Fingerprint string   `json:"-"`
	Version     string   `json:"-"`
	Commit      string   `json:"-"`
	OS          string   `json:"-"`
	Arch        string   `json:"-"`
	Frames      []string `json:"-"`
	Detail      Detail   `json:"-"`
}

// MarshalJSON renders the wire form. The detail, when present, serialises to
// its canonical token string — the only string a closed carrier can produce.
func (p Payload) MarshalJSON() ([]byte, error) {
	w := payloadWire{
		Schema:      p.Schema,
		Kind:        p.Kind,
		Fingerprint: p.Fingerprint,
		Version:     p.Version,
		Commit:      p.Commit,
		OS:          p.OS,
		Arch:        p.Arch,
		Frames:      p.Frames,
	}
	if p.Detail != nil {
		w.Detail = p.Detail.Token()
	}
	return json.Marshal(w)
}

// ErrBuildRejected wraps every payload construction refusal. Capture treats
// it as drop-the-signal: a payload that cannot satisfy the schema is never
// queued, never sent, never shown to the model.
var ErrBuildRejected = errors.New("bugreport: payload construction refused")

// Build assembles a payload from typed inputs only — kind, frames, detail,
// build identity (REQ-ANON-011). There is no error parameter, no any, and no
// free string: TestBuilderTakesNoErrorOrAny pins the signature from source.
// Every field is validated before the fingerprint is computed, and an invalid
// input is an error, never a degraded payload.
//
// @MX:ANCHOR: [AUTO] payload builder — the single constructor every capture site renders a payload through
// @MX:REASON: a second construction path could introduce a field outside the closed schema, which is the leak shape this SPEC exists to foreclose (REQ-ANON-011)
func Build(kind Kind, frames []string, detail Detail, build BuildIdentity) (Payload, error) {
	if !kind.Valid() {
		return Payload{}, fmt.Errorf("%w: unknown kind %q", ErrBuildRejected, kind)
	}
	if IsLocalOnly(kind, frames) {
		return Payload{}, ErrZeroFramesStayLocal
	}
	if len(frames) == 0 {
		return Payload{}, fmt.Errorf("%w: no frames for kind %s", ErrBuildRejected, kind)
	}
	if len(frames) > FrameLimit() {
		return Payload{}, fmt.Errorf("%w: %d frames over the %d limit", ErrBuildRejected, len(frames), FrameLimit())
	}
	for _, f := range frames {
		if err := ValidateFrameName(f); err != nil {
			return Payload{}, fmt.Errorf("%w: %v", ErrBuildRejected, err)
		}
	}
	if detail != nil {
		if err := validateDetailForKind(kind, detail); err != nil {
			return Payload{}, fmt.Errorf("%w: %v", ErrBuildRejected, err)
		}
	}

	p := Payload{
		Schema: SchemaV1,
		Kind:   kind,
		// The build identity is validated with the rest of the fields: a dev
		// build's "none" commit or a codename version fails the allowlist and
		// the signal stays local — a dev build must not publish.
		Version: build.Version,
		Commit:  build.Commit,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Frames:  frames,
		Detail:  detail,
	}
	p.Fingerprint = Fingerprint(CanonicalInput{
		Build:  build,
		OS:     p.OS,
		Arch:   p.Arch,
		Kind:   kind,
		Frames: frames,
	})
	if err := ValidatePayload(p); err != nil {
		return Payload{}, fmt.Errorf("%w: %v", ErrBuildRejected, err)
	}
	return p, nil
}
