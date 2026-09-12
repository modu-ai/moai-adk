package receipt

import "testing"

func TestEmptyEnvelopeAmbiguityAcrossProviders(t *testing.T) {
	m, _ := New(testUUID)
	a := candidate("same", "prior", "A")
	plain := a
	plain.Provider = "anthropic"
	plain.Required = false
	plain.Opaque = Digest{}
	plain.Items = 0
	if e := m.Publish(plain); e != nil {
		t.Fatal(e)
	}
	empty := Observation{Prefix: a.Prefix, Previous: a.Previous, Provider: "anthropic"}
	if e := m.Check(testUUID, []Observation{empty}); e != nil {
		t.Fatal("plain foreign receipt rejected", e)
	}
	if e := m.Publish(a); e != nil {
		t.Fatal(e)
	}
	if e := m.Check(testUUID, []Observation{empty}); e == nil {
		t.Fatal("provider selected away required receipt")
	}
	actual := Observation{Prefix: a.Prefix, Previous: a.Previous, Provider: "openai", Opaque: a.Opaque, Items: a.Items}
	if e := m.Check(testUUID, []Observation{actual}); e != nil {
		t.Fatal("exact opaque rejected", e)
	}
	actual.Provider = "anthropic"
	if e := m.Check(testUUID, []Observation{actual}); e == nil {
		t.Fatal("opaque provider mismatch accepted")
	}
}
func TestZeroParentCannotCreateFork(t *testing.T) {
	var zero Manifest
	if _, e := zero.Fork(testUUID); e == nil {
		t.Fatal("zero parent fork accepted")
	}
}
