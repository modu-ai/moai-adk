package gateway

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// EstimateInputTokens matches OpenAIConfig.MeasureInput and count_tokens' local
// json-runes-div4 estimate. It is neither an exact tokenizer nor an upper bound.
// The caller retains byte limits and provider-authoritative context rejection.
func EstimateInputTokens(_ ModelEntry, body []byte) (int64, error) {
	if err := translate.ValidateJSONObject(body); err != nil {
		return 0, err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	return estimateInputProjection(payload)
}

func estimateInputProjection(payload map[string]json.RawMessage) (int64, error) {
	input := make(map[string]json.RawMessage)
	for _, key := range []string{"messages", "system", "tools"} {
		if value, ok := payload[key]; ok {
			input[key] = value
		}
	}
	if value, ok := payload["output_config"]; ok {
		var config map[string]json.RawMessage
		if err := json.Unmarshal(value, &config); err != nil {
			return 0, err
		}
		if format, ok := config["format"]; ok {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(format, &fields); err != nil {
				return 0, err
			}
			if schema, ok := fields["schema"]; ok {
				input["output_schema"] = schema
			}
		}
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	// Convert before rounding to avoid overflowing a platform-sized int.
	return (int64(utf8.RuneCount(raw)) + 3) / 4, nil
}
