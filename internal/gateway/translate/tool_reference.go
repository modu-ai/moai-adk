package translate

import (
	"encoding/json"
	"errors"
)

// toolResultContent preserves text and represents discovered tool references using
// their request-local Responses names. Only supplied schemas can be activated.
func toolResultContent(v any, forward map[string]string, definitions map[string]any, images bool) ([]any, []string, error) {
	blocks, err := contentBlocks(v)
	if err != nil {
		return nil, nil, err
	}
	content := []any{}
	references := []string{}
	for _, block := range blocks {
		if block["type"] == "image" && images {
			raw, _ := json.Marshal(block)
			imageURL, err := ImageSourceURL(raw)
			if err != nil {
				return nil, nil, err
			}
			content = append(content, map[string]any{"type": "input_image", "image_url": imageURL})
			continue
		}
		if block["type"] != "tool_reference" {
			part, err := textContent([]any{block})
			if err != nil {
				return nil, nil, err
			}
			for _, text := range part {
				content = append(content, map[string]any{"type": "input_text", "text": text})
			}
			continue
		}
		if err := keys(block, "type", "tool_name"); err != nil {
			return nil, nil, err
		}
		name, err := stringField(block, "tool_name")
		if err != nil || definitions[name] == nil || forward[name] == "" {
			return nil, nil, errors.New("invalid or unavailable tool reference")
		}
		content = append(content, map[string]any{"type": "input_text", "text": "Tool available: " + forward[name]})
		references = append(references, name)
	}
	return content, references, nil
}
