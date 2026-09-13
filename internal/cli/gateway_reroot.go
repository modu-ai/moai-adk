package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Launcher-side wedge recovery (SPEC-GATEWAY-WEDGE-REROOT-001, REQ-WRR-003).
// A transient upstream failure can leave the client transcript holding an
// assistant boundary the gateway never published; every later replay is then
// rejected with the chain cause and the conversation is wedged. This file
// re-roots the transcript by removing exactly that trailing boundary and its
// dependent tool results under four bounds:
//
//   - Trailing-boundary-bounded: only the last assistant boundary and the
//     tool-result rows dependent on it are removed; every other row is
//     preserved byte-identically, including plain user turns and the
//     terminal API-error display row.
//   - Single-shot: a durable attempt marker (attempted flag + preimage
//     digest) is written BEFORE any mutation, so the bound holds across
//     process restarts — a fresh invocation reading the marker removes
//     nothing and surfaces the fixed guidance (REQ-WRR-006).
//   - Non-destructive: the removed rows land in an aside file before the
//     transcript is rewritten and stay retrievable (aside-before-replace).
//   - No gateway write: this path holds no receipt-store handle and opens
//     none (locked structurally by TestGatewayRerootPathHoldsNoReceiptStore);
//     acceptance after recovery comes solely from the validator's existing
//     tail-truncation tolerance, which the characterization suite pins.
//
// Recovery is explicit-user-invoked only (REQ-WRR-005): the launcher runs it
// solely on the --reroot request paired with --resume; no rejection path
// calls into this file.
const (
	gatewayRerootMarkerSuffix = ".reroot.json"
	gatewayRerootAsideSuffix  = ".reroot-aside.jsonl"
)

const gatewayRerootExhaustedGuidance = "gateway wedge recovery was already attempted for this conversation (single-shot bound); retry the conversation to surface the gateway's classified guidance"

const gatewayRerootNothingGuidance = "no unpublished assistant boundary found to re-root; the conversation transcript was left untouched"

// gatewayRerootMarker is the durable attempt record. Its presence alone
// exhausts the wedge incident; the preimage digest names exactly what the
// one attempt was allowed to remove.
type gatewayRerootMarker struct {
	Attempted bool   `json:"attempted"`
	Preimage  string `json:"preimage"`
	Removed   int    `json:"removed"`
	Aside     string `json:"aside"`
}

// gatewayRerootOutcome reports one bounded attempt. Removed is 0 and
// Guidance non-empty whenever the attempt was refused (exhausted incident or
// nothing to re-root); the transcript is untouched in both cases.
type gatewayRerootOutcome struct {
	Removed  int
	Aside    string
	Guidance string
}

// rerootGatewayTranscript performs the single re-rooting attempt on the
// transcript file at path. The marker sidecar (path + .reroot.json) is the
// only state: the function holds none of its own, so any invocation — this
// process or a fresh one — observes the same durable bound.
func rerootGatewayTranscript(transcript string) (gatewayRerootOutcome, error) {
	markerPath := transcript + gatewayRerootMarkerSuffix
	if raw, err := os.ReadFile(markerPath); err == nil {
		var marker gatewayRerootMarker
		_ = json.Unmarshal(raw, &marker)
		guidance := gatewayRerootExhaustedGuidance
		if marker.Attempted && marker.Aside != "" {
			guidance += "; the removed content remains preserved at " + marker.Aside
		}
		return gatewayRerootOutcome{Aside: marker.Aside, Guidance: guidance}, nil
	}
	raw, err := os.ReadFile(transcript)
	if err != nil {
		return gatewayRerootOutcome{}, err
	}
	lines := strings.Split(string(raw), "\n")
	rows := make([]map[string]any, len(lines))
	last := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return gatewayRerootOutcome{}, fmt.Errorf("reroot: transcript row %d: %w", i+1, err)
		}
		rows[i] = row
		// An API-error display row is the client's terminal failure note, not
		// a model boundary: the gateway never saw it and it is never the
		// unpublished boundary (it stays preserved like any other row).
		if row["type"] == "assistant" && row["isApiErrorMessage"] != true {
			last = i
		}
	}
	if last < 0 {
		// Nothing to re-root: refuse cleanly and touch no file.
		return gatewayRerootOutcome{Guidance: gatewayRerootNothingGuidance}, nil
	}
	// The removal set: the trailing assistant boundary plus every row after
	// it whose content is entirely tool results bound to that boundary's
	// tool_use ids. Everything else is preserved, in order.
	toolIDs := map[string]bool{}
	if msg, ok := rows[last]["message"].(map[string]any); ok {
		if blocks, ok := msg["content"].([]any); ok {
			for _, v := range blocks {
				if b, ok := v.(map[string]any); ok && b["type"] == "tool_use" {
					if id, ok := b["id"].(string); ok {
						toolIDs[id] = true
					}
				}
			}
		}
	}
	dependent := func(row map[string]any) bool {
		if len(toolIDs) == 0 {
			return false
		}
		msg, ok := row["message"].(map[string]any)
		if !ok {
			return false
		}
		blocks, ok := msg["content"].([]any)
		if !ok || len(blocks) == 0 {
			return false
		}
		for _, v := range blocks {
			b, ok := v.(map[string]any)
			if !ok || b["type"] != "tool_result" {
				return false
			}
			id, ok := b["tool_use_id"].(string)
			if !ok || !toolIDs[id] {
				return false
			}
		}
		return true
	}
	remove := map[int]bool{last: true}
	for i := last + 1; i < len(rows); i++ {
		if rows[i] != nil && dependent(rows[i]) {
			remove[i] = true
		}
	}
	h := sha256.New()
	removed := 0
	var aside strings.Builder
	for i, line := range lines {
		if !remove[i] {
			continue
		}
		h.Write([]byte(line))
		h.Write([]byte("\n"))
		aside.WriteString(line)
		aside.WriteString("\n")
		removed++
	}
	asidePath := transcript + gatewayRerootAsideSuffix
	markerRaw, err := json.Marshal(gatewayRerootMarker{Attempted: true, Preimage: hex.EncodeToString(h.Sum(nil)), Removed: removed, Aside: asidePath})
	if err != nil {
		return gatewayRerootOutcome{}, err
	}
	// Durable marker first, exclusive-create so concurrent invocations
	// observe the same single-shot bound. Every failure below leaves the
	// transcript untouched: the incident is consumed, no content was lost.
	f, err := os.OpenFile(markerPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: attempt marker: %w", err)
	}
	if _, err = f.Write(markerRaw); err != nil {
		_ = f.Close()
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: attempt marker: %w", err)
	}
	if err = f.Close(); err != nil {
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: attempt marker: %w", err)
	}
	// Aside-before-replace: removed rows land aside before the transcript is
	// rewritten, so the removal stays recoverable (never hard-deleted).
	if err = os.WriteFile(asidePath, []byte(aside.String()), 0600); err != nil {
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: aside copy: %w", err)
	}
	kept := make([]string, 0, len(lines))
	for i, line := range lines {
		if !remove[i] {
			kept = append(kept, line)
		}
	}
	tmp := transcript + ".reroot-tmp"
	if err = os.WriteFile(tmp, []byte(strings.Join(kept, "\n")), 0600); err != nil {
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: transcript rewrite: %w", err)
	}
	if err = os.Rename(tmp, transcript); err != nil {
		_ = os.Remove(tmp)
		return gatewayRerootOutcome{}, fmt.Errorf("reroot: transcript rewrite: %w", err)
	}
	return gatewayRerootOutcome{Removed: removed, Aside: asidePath}, nil
}

// rerootRefusalError keeps the wiring error text uniform.
func rerootRefusalError(outcome gatewayRerootOutcome) error {
	if outcome.Guidance == "" {
		return errors.New("gateway reroot refused without guidance")
	}
	return errors.New("gateway reroot: " + outcome.Guidance)
}
