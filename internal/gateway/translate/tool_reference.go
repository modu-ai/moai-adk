package translate

import "errors"

// toolResultContent preserves text and represents discovered tool references using
// their request-local Responses names. Only supplied schemas can be activated.
func toolResultContent(v any, forward map[string]string, definitions map[string]any) ([]string, []string, error) {
	blocks, err := contentBlocks(v)
	if err != nil {
		return nil, nil, err
	}
	texts := []string{}
	references := []string{}
	for _, block := range blocks {
		if block["type"] != "tool_reference" {
			part, err := textContent([]any{block})
			if err != nil {
				return nil, nil, err
			}
			texts = append(texts, part...)
			continue
		}
		if err := keys(block, "type", "tool_name"); err != nil {
			return nil, nil, err
		}
		name, err := stringField(block, "tool_name")
		if err != nil || definitions[name] == nil || forward[name] == "" {
			return nil, nil, errors.New("invalid or unavailable tool reference")
		}
		texts = append(texts, "Tool available: "+forward[name])
		references = append(references, name)
	}
	return texts, references, nil
}
