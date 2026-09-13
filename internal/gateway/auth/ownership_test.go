package auth

import "testing"

func TestStoreOwnsOnlyItsLiveReferenceType(t *testing.T) {
	s := &Store{}
	other := &Store{}
	for name, tc := range map[string]struct {
		ref  CredentialRef
		want bool
	}{
		"same store":  {&storeRef{s: s}, true},
		"other store": {&storeRef{s: other}, false},
		"snapshot":    {&snapshotRef{}, false},
		"nil":         {nil, false},
		"typed nil":   {(*storeRef)(nil), false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := s.Owns(tc.ref); got != tc.want {
				t.Fatalf("Owns = %v, want %v", got, tc.want)
			}
		})
	}
	var absent *Store
	if absent.Owns(&storeRef{}) {
		t.Fatal("nil store owns reference")
	}
}
