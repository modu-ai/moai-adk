package translate

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
)

// HistoryAuthority is provided by the private, authorized conversation binding.
// Opaque envelopes carry consistency data; only this boundary authorizes replay.
type HistoryAuthority interface {
	Check(context.Context, string, string, []byte) error
	Publish(context.Context, string, string, []byte) error
}

func messageEnvelope(blocks []map[string]any) (*opaque.Envelope, error) {
	var envelope *opaque.Envelope
	for _, b := range blocks {
		if b["type"] != "redacted_thinking" {
			continue
		}
		if envelope != nil || keys(b, "type", "data") != nil {
			return nil, errors.New("duplicate or invalid opaque block")
		}
		data, err := stringField(b, "data")
		if err != nil {
			return nil, err
		}
		envelope, err = opaque.Decode(data)
		if err != nil {
			return nil, err
		}
	}
	return envelope, nil
}
func outputEnvelope(output []any) (*opaque.Envelope, error) {
	return outputEnvelopeWithRaw(output, nil)
}
func outputEnvelopeWithRaw(output []any, original map[int][]byte) (*opaque.Envelope, error) {
	items := []opaque.Item{}
	public := []opaque.PublicItem{}
	for i, v := range output {
		item, ok := v.(map[string]any)
		if !ok {
			return nil, errors.New("invalid output item")
		}
		if item["type"] != "reasoning" {
			typ, _ := item["type"].(string)
			phase, _ := item["phase"].(string)
			_, hasPhase := item["phase"]
			count := 1
			if typ == "message" {
				parts, ok := item["content"].([]any)
				if !ok {
					return nil, errors.New("invalid public content")
				}
				count = len(parts)
			}
			public = append(public, opaque.PublicItem{OutputIndex: i, Type: typ, Blocks: count, Phase: phase, PhaseNull: hasPhase && item["phase"] == nil})
			continue
		}
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		if saved := original[i]; saved != nil {
			raw = saved
		}
		items = append(items, opaque.Item{OutputIndex: i, Raw: raw})
	}
	if len(items) == 0 {
		return nil, nil
	}
	return opaque.EncodeWithPublic(items, public)
}
