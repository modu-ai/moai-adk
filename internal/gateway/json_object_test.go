package gateway

import "testing"

func TestValidateJSONObjectRejectsAmbiguousInput(t *testing.T) {
	for _, body := range [][]byte{[]byte(`{"nested":{"k":1,"k":2}}`), []byte(`[]`), []byte(`null`), []byte(`{} {}`), []byte{'{', '"', 'k', '"', ':', '"', 255, '"', '}'}} {
		if err := ValidateJSONObject(body); err == nil {
			t.Fatalf("accepted %q", body)
		}
	}
	if err := ValidateJSONObject([]byte(`{"한글":{"키":"값"}}`)); err != nil {
		t.Fatal(err)
	}
}
