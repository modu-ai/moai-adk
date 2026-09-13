package conversation

// Launcher-side reasoning-envelope repair (SPEC-GATEWAY-ENVELOPE-REPAIR-001,
// REQ-EVR-002..009). Explicit-invocation only, single-shot with a durable
// record, byte-exact verbatim re-injection verified by marker self-attestation,
// aside-before-replace. This file is the entire repair path: it must never
// import or call any gateway receipt-store symbol (REQ-EVR-009) — the repair
// decision rests on the tool marker's embedded digest alone.

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
)

var (
	// ErrRepairNotRepairable refuses a shape the repair cannot provably
	// restore: source carrier gone, digest mismatch, no surviving
	// self-attesting marker, or nothing to repair. A refusal performs zero
	// modification and does not consume the single-shot attempt.
	ErrRepairNotRepairable = errors.New("envelope repair refused")
	// ErrRepairAlreadyAttempted terminates recovery after the single-shot
	// attempt was durably recorded; the bound survives process restarts.
	ErrRepairAlreadyAttempted = errors.New("envelope repair already attempted")
)

// asideSuffix marks the preserved preimage of a repaired transcript. The file
// is never deleted by the repair path.
const asideSuffix = ".moai-repair-aside"

// RepairProvenance records one injected boundary: the digest the surviving
// marker self-attested and the transcript line position that received the
// verbatim envelope.
type RepairProvenance struct {
	Digest   string `json:"digest"`
	Position int    `json:"position"`
}

// repairRecord is the durable single-shot marker on the launcher conversation
// record namespace.
type repairRecord struct {
	Attempted   bool               `json:"attempted"`
	Boundaries  []RepairProvenance `json:"boundaries"`
	Aside       string             `json:"aside"`
	AsideSHA256 string             `json:"aside_sha256"`
}

func (m *Manager) repairStatePath(r record) string {
	return filepath.Join(m.root, "families", r.FamilyID, "repair", r.UUID+".json")
}

// RepairEnvelope restores gateway-issued reasoning envelopes that the client
// dropped from persisted assistant boundaries while their bound
// toolu_moai_v1_* markers survived. The verbatim carrier is read from the
// family transcript itself; sha256(carrier raw bytes) must equal the digest
// embedded in the marker (marker self-attestation) before anything is
// injected. All-or-nothing: any unprovable boundary aborts with zero
// modification.
//
// @MX:NOTE: [AUTO] admissibility is marker self-attestation only — no receipt
// store read exists on this path by design (REQ-EVR-003-4); the unchanged
// gateway Check adjudicates the repaired replay at request time.
// @MX:SPEC: SPEC-GATEWAY-ENVELOPE-REPAIR-001 REQ-EVR-002/003/006/007/008
func (m *Manager) RepairEnvelope(ctx context.Context, id string) ([]RepairProvenance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := m.refreshNative(ctx, id); err != nil && err != ErrIncomplete {
		return nil, err
	}
	m.mu.Lock()
	r, ok := m.entries[id]
	m.mu.Unlock()
	if !ok {
		return nil, ErrMissing
	}
	if r.Transcript == "" || r.Completion == 0 {
		return nil, fmt.Errorf("%w: conversation transcript is incomplete", ErrRepairNotRepairable)
	}
	p := filepath.Join(r.ConfigDir, filepath.FromSlash(r.Transcript))
	if !within(r.ConfigDir, p) {
		return nil, fmt.Errorf("%w: transcript path escapes the family namespace", ErrRepairNotRepairable)
	}
	info, err := os.Lstat(p)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxIndexBytes {
		return nil, fmt.Errorf("%w: transcript is unreadable", ErrRepairNotRepairable)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("%w: transcript is unreadable", ErrRepairNotRepairable)
	}
	statePath := m.repairStatePath(r)
	if _, err = os.Stat(statePath); err == nil {
		return nil, ErrRepairAlreadyAttempted
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	injections, pool, err := repairPlan(lines)
	if err != nil {
		return nil, err
	}
	// Aside-before-replace: the preimage is preserved before any repaired
	// form exists, and never hard-deleted. The durable record is written
	// before the transcript: a crash between the two burns the single shot
	// conservatively instead of allowing a second injection.
	asidePath := p + asideSuffix
	sum := sha256.Sum256(raw)
	if err = writeExclusive(asidePath, raw); err != nil {
		return nil, err
	}
	record := repairRecord{
		Attempted:   true,
		Boundaries:  injections,
		Aside:       filepath.Base(asidePath),
		AsideSHA256: hex.EncodeToString(sum[:]),
	}
	recordRaw, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(statePath), 0700); err != nil {
		return nil, err
	}
	if err = writeExclusive(statePath, recordRaw); err != nil {
		return nil, err
	}
	for _, in := range injections {
		patched, err := repairInjectLine(lines[in.Position], pool[in.Digest])
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRepairNotRepairable, err)
		}
		lines[in.Position] = patched
	}
	if err = atomicWrite(p, []byte(strings.Join(lines, "\n")+"\n")); err != nil {
		return nil, err
	}
	return injections, nil
}

// repairPlan scans transcript rows for stripped boundaries: assistant
// messages whose bound tool marker's digest has no matching
// redacted_thinking carrier in the same message. It collects every stripped
// boundary before verifying that the transcript retains a verbatim carrier
// for each — all-or-nothing. Subagent rows (isSidechain) are out of scope.
func repairPlan(lines []string) ([]RepairProvenance, map[string]string, error) {
	pool := map[string]string{}
	type strippedBoundary struct {
		line    int
		digests []string
	}
	var stripped []strippedBoundary
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, nil, fmt.Errorf("%w: transcript row %d is not JSON", ErrRepairNotRepairable, i)
		}
		if row["isSidechain"] == true {
			continue
		}
		msg, ok := row["message"].(map[string]any)
		if !ok {
			continue
		}
		blocks, ok := msg["content"].([]any)
		if !ok {
			continue
		}
		var carriers, markers []string
		for _, value := range blocks {
			block, ok := value.(map[string]any)
			if !ok {
				continue
			}
			switch block["type"] {
			case "redacted_thinking":
				data, _ := block["data"].(string)
				digest, err := repairCarrierDigest(data)
				if err != nil {
					return nil, nil, fmt.Errorf("%w: transcript row %d carries an undecodable envelope", ErrRepairNotRepairable, i)
				}
				carriers = append(carriers, digest)
				pool[digest] = data
			case "tool_use":
				markerID, _ := block["id"].(string)
				digest, err := repairMarkerDigest(markerID)
				if err != nil {
					continue // not a gateway-bound marker; public content
				}
				markers = append(markers, digest)
			}
		}
		missing := map[string]bool{}
		for _, d := range markers {
			if !containsString(carriers, d) {
				missing[d] = true
			}
		}
		if len(missing) > 0 {
			digests := make([]string, 0, len(missing))
			for d := range missing {
				digests = append(digests, d)
			}
			stripped = append(stripped, strippedBoundary{line: i, digests: digests})
		}
	}
	out := make([]RepairProvenance, 0, len(stripped))
	for _, p := range stripped {
		for _, d := range p.digests {
			if _, ok := pool[d]; !ok {
				return nil, nil, fmt.Errorf("%w: no verbatim source carrier for boundary at row %d (digest %s)", ErrRepairNotRepairable, p.line, shortDigest(d))
			}
		}
	}
	for _, p := range stripped {
		for _, d := range p.digests {
			out = append(out, RepairProvenance{Digest: d, Position: p.line})
		}
	}
	if len(out) == 0 {
		return nil, nil, fmt.Errorf("%w: no stripped boundary found", ErrRepairNotRepairable)
	}
	return out, pool, nil
}

// repairCarrierDigest returns sha256 over the carrier's decoded canonical raw
// bytes; opaque.Decode additionally rejects any non-canonical alternate
// spelling, so exactly one byte sequence can ever satisfy a marker.
func repairCarrierDigest(data string) (string, error) {
	env, err := opaque.Decode(data)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(env.Bytes())
	return hex.EncodeToString(sum[:]), nil
}

// repairMarkerDigest extracts the digest a surviving tool marker self-attests.
// The marker format is parsed here, not via opaque codec changes, because the
// codec is frozen by this SPEC; the {call_id, opaque_sha256} shape is the
// BindToolID wire contract.
func repairMarkerDigest(marker string) (string, error) {
	if !strings.HasPrefix(marker, opaque.ToolPrefix) || len(marker) > 4096 {
		return "", errors.New("not a gateway tool marker")
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimPrefix(marker, opaque.ToolPrefix))
	if err != nil {
		return "", err
	}
	var b struct {
		CallID string `json:"call_id"`
		Digest string `json:"opaque_sha256"`
	}
	if json.Unmarshal(raw, &b) != nil || len(b.Digest) != 64 {
		return "", errors.New("malformed tool marker binding")
	}
	return b.Digest, nil
}

// repairEnvelopeBlock is the injected block; the struct fixes a
// deterministic JSON key order.
type repairEnvelopeBlock struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// repairContentInsertOffset streams the row JSON only far enough to locate
// the byte offset just past the opening bracket of message.content, so the
// splice can insert without re-serializing anything.
func repairContentInsertOffset(line string) (int, bool) {
	dec := json.NewDecoder(strings.NewReader(line))
	if t, err := dec.Token(); err != nil {
		return 0, false
	} else if d, ok := t.(json.Delim); !ok || d != '{' {
		return 0, false
	}
	for {
		keyTok, err := dec.Token()
		if err != nil {
			return 0, false
		}
		key, ok := keyTok.(string)
		if !ok {
			return 0, false // row object closed: no message
		}
		if key != "message" {
			var skip json.RawMessage
			if dec.Decode(&skip) != nil {
				return 0, false
			}
			continue
		}
		if t, err := dec.Token(); err != nil {
			return 0, false
		} else if d, ok := t.(json.Delim); !ok || d != '{' {
			return 0, false
		}
		for {
			mKeyTok, err := dec.Token()
			if err != nil {
				return 0, false
			}
			mKey, ok := mKeyTok.(string)
			if !ok {
				return 0, false // message object closed: no content array
			}
			if mKey != "content" {
				var skip json.RawMessage
				if dec.Decode(&skip) != nil {
					return 0, false
				}
				continue
			}
			if t, err := dec.Token(); err != nil {
				return 0, false
			} else if d, ok := t.(json.Delim); !ok || d != '[' {
				return 0, false
			}
			return int(dec.InputOffset()), true
		}
	}
}

// repairInjectLine splices the verbatim envelope block into the head of the
// row's message.content array — the position the gateway itself emits it —
// WITHOUT re-serializing the row: every byte outside the inserted span is
// the input's own byte (key order, escaping, whitespace all preserved).
func repairInjectLine(line, data string) (string, error) {
	block, err := json.Marshal(repairEnvelopeBlock{Type: "redacted_thinking", Data: data})
	if err != nil {
		return "", err
	}
	pos, ok := repairContentInsertOffset(line)
	if !ok {
		return "", errors.New("message content array not located in row")
	}
	rest := line[pos:]
	if strings.HasPrefix(strings.TrimLeft(rest, " \t\r\n"), "]") {
		// Empty array: no separator needed.
		return line[:pos] + string(block) + rest, nil
	}
	return line[:pos] + string(block) + "," + rest, nil
}

func writeExclusive(path string, raw []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: aside already exists from a prior partial attempt", ErrRepairNotRepairable)
		}
		return err
	}
	if _, err = f.Write(raw); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func atomicWrite(path string, raw []byte) error {
	tmp := path + ".repair-tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func shortDigest(d string) string {
	if len(d) > 12 {
		return d[:12]
	}
	return d
}
