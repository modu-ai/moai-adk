package codexbridge

import "encoding/json"

// MarshalJSON preserves the mandatory text field for empty tool responses.
// Image replies retain their separate imageUrl variant without a text field.
func (c Content) MarshalJSON() ([]byte, error) {
	if c.Type == "inputText" {
		return json.Marshal(struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: c.Type, Text: c.Text})
	}
	type plain Content
	return json.Marshal(plain(c))
}
