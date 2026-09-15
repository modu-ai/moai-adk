package translate

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// ReplayCause classifies why a history replay was rejected. A cause is fixed
// recovery guidance about the rejection class only — it never carries history
// data, counts, digests, or session identifiers. Card t672: every production
// wedge class observable at this boundary is named so the next occurrence
// self-identifies in the client-visible rejection body.
type ReplayCause uint8

const (
	// CauseChain: the replayed prefix chain does not match the recorded
	// receipt chain — history was edited, reordered, sliced, or contains a
	// boundary that was never published (for example after an upstream
	// failure the client still recorded).
	CauseChain ReplayCause = iota
	// CauseLineage: the receipt root holds no recorded history for this
	// session, so a replayed assistant history cannot be authorized against
	// any lineage. The sanctioned fork path seeds a child root through the
	// launcher; no request metadata can select or seed a receipt root here.
	CauseLineage
	// CauseReasoning: a replayed boundary omits the opaque envelope that a
	// required receipt proves was gateway-issued at that exact position.
	CauseReasoning
)

// historyReplayGuidance is the fixed recovery sentence this error has always
// carried; classified messages keep it verbatim as their prefix so past
// incident signatures keep matching.
const historyReplayGuidance = "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation"

// HistoryReplayError exposes only fixed recovery guidance, never history data.
// The zero value reports the chain cause, whose message is the historical one
// plus a fixed reason clause.
type HistoryReplayError struct {
	Cause ReplayCause
	code  string
}

func (e HistoryReplayError) CauseCode() string {
	if e.code != "" {
		return e.code
	}
	if e.Cause == CauseLineage {
		return "receipt_manifest_mismatch"
	}
	return "history_changed"
}

// rejectedHistory adds diagnostics only after the receipt verifier rejected
// the request. Shapes here cannot authorize history or relax replay checks.
func rejectedHistory(cause ReplayCause, raw []byte) HistoryReplayError {
	err := HistoryReplayError{Cause: cause}
	var messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(raw, &messages) != nil {
		return err
	}
	var previous json.RawMessage
	for _, message := range messages {
		if message.Role != "assistant" {
			previous = nil
			continue
		}
		var blocks []struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(message.Content, &blocks)
		for _, block := range blocks {
			if block.Type == "agent_summary" {
				err.code = "agent_summary_untrusted"
				return err
			}
		}
		if previous != nil && bytes.Equal(previous, message.Content) {
			err.code = "resume_duplicate_input"
			return err
		}
		previous = message.Content
	}
	return err
}

func (e HistoryReplayError) Error() string {
	switch e.Cause {
	case CauseLineage:
		return historyReplayGuidance + " (reason: no recorded history exists for this session; resume or fork through the moai launcher, or start a new conversation)"
	case CauseReasoning:
		return historyReplayGuidance + " (reason: replayed history omits gateway-issued reasoning recorded at this position; replay the history unmodified or start a new conversation)"
	default:
		return historyReplayGuidance + " (reason: replayed history does not match the recorded receipt chain; start a new conversation)"
	}
}

// classifiedError carries the replay cause of an observations-level rejection
// so Check can surface the same fixed recovery guidance the manifest-level
// classes use. Card t703: these paths previously leaked the bare
// receipt.ErrInvalid body, leaving the failing check unidentifiable. The
// cause is recovery guidance only — no request-derived detail is added.
type classifiedError struct {
	err   error
	cause ReplayCause
}

func (e classifiedError) Error() string { return e.err.Error() }
func (e classifiedError) Unwrap() error { return e.err }

func classify(cause ReplayCause, err error) error {
	return classifiedError{err: err, cause: cause}
}

type receiptHistory struct {
	store         *receipt.Store
	session       string
	family        string
	compatibleGPT bool
}

// NewReceiptHistory accepts only a launcher-authorized UUID and private store.
// Request metadata never selects a receipt root or authorizes a conversation.
func NewReceiptHistory(store *receipt.Store, session, family string) HistoryAuthority {
	return &receiptHistory{store: store, session: session, family: family}
}

// NewGPTSubscriptionReceiptHistory enables the measured GPT-6/GPT-5.6
// subscription compatibility pair. Existing receipts retain their source model
// family domain; only the new response is published under the target domain.
func NewGPTSubscriptionReceiptHistory(store *receipt.Store, session, family string) HistoryAuthority {
	return &receiptHistory{store: store, session: session, family: family, compatibleGPT: true}
}

func historyFamily(model string) (string, error) {
	switch model {
	case "gpt-6-astra":
		return model, nil
	case "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna":
		return "gpt-5.6", nil
	}
	return "", errors.New("unverified reasoning model family")
}
func (h *receiptHistory) observations(model, scope string, raw []byte) ([]receipt.Observation, error) {
	family, err := historyFamily(model)
	if err != nil {
		return nil, err
	}
	if h.store == nil || scope == "" {
		return nil, errors.New("private history authority unavailable")
	}
	var messages []map[string]any
	if json.Unmarshal(raw, &messages) != nil {
		return nil, classify(CauseChain, receipt.ErrInvalid)
	}
	ids := map[string]string{}
	observations := []receipt.Observation{}
	for _, m := range messages {
		if m["role"] != "assistant" {
			continue
		}
		blocks, err := contentBlocks(m["content"])
		if err != nil {
			return nil, err
		}
		envelope, err := messageEnvelope(blocks)
		if err != nil {
			return nil, err
		}
		o := receipt.Observation{Provider: "openai"}
		if envelope != nil {
			digest, _ := hex.DecodeString(envelope.Digest())
			copy(o.Opaque[:], digest)
			o.Items = uint32(len(envelope.Items()))
		}
		for _, b := range blocks {
			if b["type"] != "tool_use" {
				continue
			}
			id, err := stringField(b, "id")
			if err != nil {
				return nil, err
			}
			original := id
			if envelope != nil {
				original, err = opaque.RestoreToolID(id, envelope)
				if err != nil {
					return nil, classify(CauseChain, err)
				}
			} else if strings.HasPrefix(id, opaque.ToolPrefix) {
				// The reasoning envelope that binds this gateway-issued tool
				// marker was dropped from the replay, so the marker cannot be
				// restored and the request is untranslatable. Card t703: a
				// mid-conversation model switch produces exactly this shape
				// when the client re-encodes prior turns.
				return nil, classify(CauseReasoning, receipt.ErrInvalid)
			}
			if previous, ok := ids[id]; ok && previous != original {
				return nil, classify(CauseChain, receipt.ErrInvalid)
			}
			ids[id] = original
		}
		observations = append(observations, o)
	}
	boundaries, err := receipt.CanonicalPrefixes(raw, func(id string) (string, error) {
		if original, ok := ids[id]; ok {
			return original, nil
		}
		if strings.HasPrefix(id, opaque.ToolPrefix) {
			return "", receipt.ErrInvalid
		}
		return id, nil
	})
	if err != nil {
		return nil, classify(CauseChain, err)
	}
	if len(boundaries) != len(observations) {
		return nil, classify(CauseChain, receipt.ErrInvalid)
	}
	// Hash-only domain separation binds the existing receipt schema to the stable
	// credential owner and verified model family without retaining private data.
	domain := []byte("moai-history-v1\x00" + h.family + "\x00" + scope + "\x00" + family + "\x00")
	bind := func(d receipt.Digest) receipt.Digest {
		if d == (receipt.Digest{}) {
			return d
		}
		return receipt.Hash(append(append([]byte(nil), domain...), d[:]...))
	}
	for i, b := range boundaries {
		observations[i].Prefix = bind(b.Prefix)
		observations[i].Previous = bind(b.Previous)
	}
	return observations, nil
}
func (h *receiptHistory) Check(ctx context.Context, model, scope string, raw []byte) error {
	observations, err := h.observations(model, scope, raw)
	if err != nil {
		var classified classifiedError
		if errors.As(err, &classified) {
			return rejectedHistory(classified.cause, raw)
		}
		return err
	}
	manifest, err := h.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	if err = h.checkObserved(manifest, model, scope, raw, observations); err != nil {
		return rejectedHistory(replayCause(manifest, observations), raw)
	}
	return nil
}

// replayCause names the rejection class without exposing history data. An
// empty root against a nonempty replay is a lineage miss; a boundary whose
// required receipt exists while its opaque envelope is absent is a stripped
// replay; everything else is a chain mismatch. The compatible-domain check
// inside checkObserved is not re-derived here, so a stripped replay whose
// required receipt sits only in the alternate domain classifies as a chain
// mismatch — the class is recovery guidance, and both classes reject.
func replayCause(manifest *receipt.Manifest, observations []receipt.Observation) ReplayCause {
	if len(observations) > 0 && len(manifest.Candidates()) == 0 {
		return CauseLineage
	}
	for _, o := range observations {
		if o.Opaque != (receipt.Digest{}) || o.Items != 0 {
			continue
		}
		for _, c := range manifest.Candidates() {
			if c.Required && c.Prefix == o.Prefix && c.Previous == o.Previous {
				return CauseReasoning
			}
		}
	}
	return CauseChain
}
func (h *receiptHistory) Publish(ctx context.Context, model, scope string, raw []byte) error {
	observations, err := h.observations(model, scope, raw)
	if err != nil {
		return err
	}
	if len(observations) == 0 {
		return receipt.ErrInvalid
	}
	manifest, err := h.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	if err = h.checkObserved(manifest, model, scope, raw, observations[:len(observations)-1]); err != nil {
		return err
	}
	last := observations[len(observations)-1]
	return h.store.Publish(ctx, receipt.Candidate{Prefix: last.Prefix, Previous: last.Previous, Provider: last.Provider, Opaque: last.Opaque, Items: last.Items, Required: last.Items > 0, Complete: true})
}

// checkObserved checks each complete prefix in its original receipt domain.
// A mixed-model conversation therefore needs no receipt migration or ciphertext
// rewriting. Both alternatives retain the same session, owner and conversation
// family hashes and require the exact opaque digest and public-prefix chain.
func (h *receiptHistory) checkObserved(manifest *receipt.Manifest, model, scope string, raw []byte, observations []receipt.Observation) error {
	if !h.compatibleGPT || len(observations) == 0 {
		return manifest.Check(h.session, observations)
	}
	alternate := "gpt-6-astra"
	if model == "gpt-6-astra" {
		alternate = "gpt-5.6-sol"
	}
	other, err := h.observations(alternate, scope, raw)
	if err != nil {
		return err
	}
	for i, o := range observations {
		// A matching required receipt in either compatible domain forbids an
		// empty fallback, even if that public prefix also has an empty candidate.
		if o.Items == 0 && o.Opaque == (receipt.Digest{}) {
			for _, candidate := range manifest.Candidates() {
				if candidate.Required && ((candidate.Prefix == o.Prefix && candidate.Previous == o.Previous) || (candidate.Prefix == other[i].Prefix && candidate.Previous == other[i].Previous)) {
					return receipt.ErrInvalid
				}
			}
		}
		if manifest.Check(h.session, []receipt.Observation{o}) == nil {
			continue
		}
		if err = manifest.Check(h.session, other[i:i+1]); err != nil {
			return err
		}
	}
	return nil
}
