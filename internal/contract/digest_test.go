package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDigest_MatchesDigestBytesAndIsHandComputable(t *testing.T) {
	raw := signFixture(renderFixture(fixtureOpts{reverseSets: true}))
	c, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	d, err := Digest(c)
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	db, err := DigestBytes(raw)
	if err != nil {
		t.Fatalf("DigestBytes: %v", err)
	}
	if d != db {
		t.Errorf("Digest %s != DigestBytes %s", d, db)
	}

	// Independent recomputation of the design.md § Digest recipe.
	canon := *c
	canon.Signature = nil
	canon.Actions = slices.Sorted(slices.Values(c.Actions))
	own := *c.Ownership
	own.Write = slices.Sorted(slices.Values(own.Write))
	own.Never = slices.Sorted(slices.Values(own.Never))
	own.Scratch = slices.Sorted(slices.Values(own.Scratch))
	canon.Ownership = &own
	canon.Invariants = slices.Sorted(slices.Values(c.Invariants))
	canon.EscalateOn = slices.Sorted(slices.Values(c.EscalateOn))
	data, err := json.Marshal(canon)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if want := hex.EncodeToString(sum[:]); d != want {
		t.Errorf("Digest %s, independent recomputation %s", d, want)
	}
	if strings.Contains(string(data), "signature") {
		t.Errorf("canonical JSON still carries the signature: %s", data)
	}
}

func TestCanonical_DoesNotMutateInput(t *testing.T) {
	c, err := Decode([]byte(renderFixture(fixtureOpts{reverseSets: true})))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	actionsBefore := slices.Clone(c.Actions)
	writeBefore := slices.Clone(c.Ownership.Write)
	canon := Canonical(*c)
	if !slices.Equal(c.Actions, actionsBefore) || !slices.Equal(c.Ownership.Write, writeBefore) {
		t.Errorf("Canonical mutated its input: actions=%v write=%v", c.Actions, c.Ownership.Write)
	}
	if !slices.IsSorted(canon.Actions) || !slices.IsSorted(canon.Ownership.Write) || !slices.IsSorted(canon.Ownership.Scratch) {
		t.Errorf("Canonical left sets unsorted: %+v %+v", canon.Actions, canon.Ownership)
	}
	if !slices.Equal(canon.Reobserve, c.Reobserve) {
		t.Errorf("Canonical reordered reobserve: %v -> %v", c.Reobserve, canon.Reobserve)
	}
	if canon.Ownership == c.Ownership || canon.Acceptance == c.Acceptance || canon.Budget == c.Budget {
		t.Errorf("Canonical shares nested pointers with its input")
	}
}

func TestDigest_Sensitivity(t *testing.T) {
	base := renderFixture(fixtureOpts{})
	d0, err := DigestBytes([]byte(base))
	if err != nil {
		t.Fatalf("DigestBytes: %v", err)
	}
	changed := map[string]string{
		"reobserve reordered": strings.Replace(base,
			"  - \"contract.yaml\"\n  - \"acceptance.md\"\n", "  - \"acceptance.md\"\n  - \"contract.yaml\"\n", 1),
		"approach edited":       strings.Replace(base, "one milestone", "two milestones", 1),
		"acceptance hash drops": strings.Replace(base, "  sha256: \""+fixtureAcceptanceSHA256()+"\"\n", "", 1),
		"verdict edited":        strings.Replace(base, "verdict: PASS", "verdict: PASS-WITH-DEBT", 1),
		"schema_version edited": strings.Replace(base, "schema_version: 1", "schema_version: 2", 1),
		"budget removed":        renderFixture(fixtureOpts{omit: map[string]bool{SectionBudget: true}}),
	}
	for name, raw := range changed {
		t.Run(name, func(t *testing.T) {
			if raw == base {
				t.Fatalf("variant did not change the text")
			}
			d, err := DigestBytes([]byte(raw))
			if err != nil {
				t.Fatalf("DigestBytes: %v", err)
			}
			if d == d0 {
				t.Errorf("digest unchanged by %s", name)
			}
		})
	}

	t.Run("scratch absent equals scratch empty", func(t *testing.T) {
		absent := renderFixture(fixtureOpts{omit: map[string]bool{SectionOwnershipScratch: true}})
		empty := strings.Replace(absent, "  never:\n", "  scratch: []\n  never:\n", 1)
		if empty == absent {
			t.Fatalf("variant did not change the text")
		}
		da, errA := DigestBytes([]byte(absent))
		de, errE := DigestBytes([]byte(empty))
		if errA != nil || errE != nil {
			t.Fatalf("DigestBytes: %v / %v", errA, errE)
		}
		if da != de {
			t.Errorf("absent scratch %s != empty scratch %s", da, de)
		}
	})
}

func TestDigest_Errors(t *testing.T) {
	if _, err := Digest(nil); err == nil {
		t.Errorf("Digest(nil) returned no error")
	}
	if _, err := DigestBytes([]byte("notes: x\n")); !errors.Is(err, ErrSchemaInvalid) {
		t.Errorf("DigestBytes on an invalid document: %v, want ErrSchemaInvalid", err)
	}
}
