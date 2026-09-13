package translate

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// HistoryReplayError exposes only fixed recovery guidance, never history data.
type HistoryReplayError struct{}

func (HistoryReplayError) Error() string {
	return "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation"
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
		return nil, receipt.ErrInvalid
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
					return nil, err
				}
			} else if strings.HasPrefix(id, opaque.ToolPrefix) {
				return nil, receipt.ErrInvalid
			}
			if previous, ok := ids[id]; ok && previous != original {
				return nil, receipt.ErrInvalid
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
	if err != nil || len(boundaries) != len(observations) {
		return nil, receipt.ErrInvalid
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
		return err
	}
	manifest, err := h.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	if err = h.checkObserved(manifest, model, scope, raw, observations); err != nil {
		return HistoryReplayError{}
	}
	return nil
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
