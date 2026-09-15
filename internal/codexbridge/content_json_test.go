package codexbridge

import (
	"encoding/json"
	"testing"
)

func TestContentJSONPreservesNativeTaggedVariants(t *testing.T) {
	for _, tt := range []struct {
		content Content
		want    string
	}{
		{Content{Type: "inputText"}, `{"type":"inputText","text":""}`},
		{Content{Type: "inputText", Text: "result"}, `{"type":"inputText","text":"result"}`},
		{Content{Type: "inputImage", ImageURL: "data:image/png;base64,aW1hZ2U="}, `{"type":"inputImage","imageUrl":"data:image/png;base64,aW1hZ2U="}`},
	} {
		raw, err := json.Marshal(tt.content)
		if err != nil || string(raw) != tt.want {
			t.Fatalf("got=%s err=%v want=%s", raw, err, tt.want)
		}
	}
}
