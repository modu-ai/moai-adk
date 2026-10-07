package publish

// revalidate.go — the send-time trust boundary (review-gate finding, P1):
// the queue file is a local file, so a queued item's stored body is
// UNTRUSTED input — exactly what could be modified on disk between enqueue
// and send. The sender never publishes stored body text as-is: every item
// is re-validated at send time and the public title and body are
// REGENERATED from the validated closed-set fields only. An intact item
// regenerates byte-identical output (the same render function, the same
// fields); a tampered one fails a cross-check and never reaches gh.

import (
	"runtime"
	"strings"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

// errUnvalidatedBody marks a queued item whose stored body failed send-time
// re-validation. The item stays queued (the attempt limit eventually drops
// it) and never reaches gh — no search, no comment, no create.
var errUnvalidatedBody = errorString("queued body failed send-time re-validation")

// revalidatedItem recovers the VALIDATED payload from a queued item's
// stored body marker and regenerates the exact title and body the send
// paths may publish. Ok is false when anything fails: the title key, the
// marker block, any field's anchored allowlist (ValidatePayload), the
// read-back detail validator (ParseDetail), or the cross-checks binding the
// stored fields to each other and to this machine's platform — the payload
// was built at capture time on the flushing machine, so a stored body
// claiming another platform is corrupt.
func revalidatedItem(item feedback.QueueItem) (payload bugreport.Payload, title, body string, ok bool) {
	kind, fp, ok := ParseTitleKey(item.Title)
	if !ok {
		return bugreport.Payload{}, "", "", false
	}
	if item.Kind != "" && item.Kind != kind {
		return bugreport.Payload{}, "", "", false
	}
	fields, valid := ParseMarker(item.Body)
	if !valid {
		return bugreport.Payload{}, "", "", false
	}
	if fields["schema"] != bugreport.SchemaV1 ||
		fields["fingerprint"] != fp || fp != item.Fingerprint ||
		fields["kind"] != kind {
		return bugreport.Payload{}, "", "", false
	}
	if fields["os_arch"] != runtime.GOOS+"/"+runtime.GOARCH {
		return bugreport.Payload{}, "", "", false
	}
	var frames []string
	if raw := fields["frames"]; raw != "" {
		frames = strings.Split(raw, ",")
	}
	var detail bugreport.Detail
	if raw := fields["detail"]; raw != "" {
		d, err := bugreport.ParseDetail(bugreport.Kind(kind), raw)
		if err != nil {
			return bugreport.Payload{}, "", "", false
		}
		detail = d
	}
	p := bugreport.Payload{
		Schema:      bugreport.SchemaV1,
		Kind:        bugreport.Kind(kind),
		Fingerprint: fp,
		Version:     fields["version"],
		Commit:      fields["commit"],
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Frames:      frames,
		Detail:      detail,
	}
	if err := bugreport.ValidatePayload(p); err != nil {
		return bugreport.Payload{}, "", "", false
	}
	// Required hold (review gate, P2): field format is not the whole trust
	// boundary — a payload the tripwire is required to hold (the
	// path-traversal sentinel) must not publish from a tampered queue any
	// more than from the spool. The rule lives in outbox (the tripwire's
	// owner) so the drain and the send path cannot fork.
	if outbox.RequiresHold(p) {
		return bugreport.Payload{}, "", "", false
	}
	regenTitle, regenBody := outbox.RenderReport(p)
	if regenTitle != item.Title {
		return bugreport.Payload{}, "", "", false
	}
	return p, regenTitle, regenBody, true
}
