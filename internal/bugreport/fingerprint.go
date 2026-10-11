package bugreport

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// SchemaV1 is the payload and fingerprint schema marker.
const SchemaV1 = "v1"

// BuildIdentity is the build identity the pipeline carries into payloads and
// fingerprints: the version and commit strings the ldflags inject. They are
// passed as a typed struct — never as loose parameters — and validated by the
// payload allowlists before anything leaves the machine.
type BuildIdentity struct {
	Version string
	Commit  string
}

// CanonicalInput is the exact input the fingerprint is derived from
// (REQ-ANON-010): the moai version, the build commit, the operating system
// and architecture, the error kind, and the ordered moai-internal frame
// names. Nothing else enters it — no file path, line number, argument, or
// non-moai frame.
type CanonicalInput struct {
	Build  BuildIdentity
	OS     string
	Arch   string
	Kind   Kind
	Frames []string
}

// canonicalString renders the newline-separated canonical input: the schema
// marker first, then the identity, platform, kind, and the ordered frames.
func (in CanonicalInput) canonicalString() string {
	var b strings.Builder
	b.WriteString(SchemaV1)
	b.WriteByte('\n')
	b.WriteString(in.Build.Version)
	b.WriteByte('\n')
	b.WriteString(in.Build.Commit)
	b.WriteByte('\n')
	b.WriteString(in.OS)
	b.WriteByte('/')
	b.WriteString(in.Arch)
	b.WriteByte('\n')
	b.WriteString(string(in.Kind))
	for _, f := range in.Frames {
		b.WriteByte('\n')
		b.WriteString(f)
	}
	return b.String()
}

// Fingerprint returns the first 16 lowercase hex digits of the SHA-256 of the
// canonical input (design.md section 4). It is the dedupe key, the cap key,
// the title token, and the family key every consumer merges on, so its
// inputs are pinned by TestFingerprintInputsAllMatter.
//
// @MX:ANCHOR: [AUTO] fingerprint — every dedupe, cap, title, and comment key derives from it
// @MX:REASON: a second derivation (different fields, different hash slice) would split one defect into families that never meet (REQ-ANON-010)
func Fingerprint(in CanonicalInput) string {
	sum := sha256.Sum256([]byte(in.canonicalString()))
	return hex.EncodeToString(sum[:])[:16]
}
