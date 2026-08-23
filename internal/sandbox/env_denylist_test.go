package sandbox

import "testing"

// The accessor exists so that consumers outside this package can reuse the name
// vocabulary. That is only safe if they cannot mutate the built-in list.
func TestDefaultEnvDenyListReturnsCopy(t *testing.T) {
	t.Parallel()

	first := DefaultEnvDenyList()
	if len(first) == 0 {
		t.Fatalf("expected a non-empty deny list")
	}

	original := first[0]
	first[0] = "MUTATED_BY_CALLER"

	second := DefaultEnvDenyList()
	if second[0] != original {
		t.Fatalf("caller mutation reached the built-in list: got %q, want %q", second[0], original)
	}
}
