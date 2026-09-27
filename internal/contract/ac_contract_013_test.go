package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestAC_CONTRACT_013 covers AC-CONTRACT-013 (the Go half; the go list
// dependency checks are the AC's shell block): verify reads a signed fixture
// project without changing any file, for the valid fixtures and every tamper
// case of AC-CONTRACT-010/011, and its verdict maps to exit 0 (valid) or 1
// (invalid) as the CLI reports it.
func TestAC_CONTRACT_013(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	digest := bodyDigest(body)
	human := sealed(humanSignature(digest, sigDefault))
	receipt := sealed(receiptSignature(digest))
	other64 := strings.Repeat("e", 64)

	// withSig renders body + the given signature verbatim.
	withSig := func(b string, s Signature) string { return b + renderSignature(s) }
	// editSig copies s, applies edit, and optionally recomputes the seal.
	editSig := func(s Signature, reseal bool, edit func(*Signature)) Signature {
		if s.Receipt != nil {
			r := *s.Receipt
			s.Receipt = &r
		}
		edit(&s)
		if reseal {
			s = sealed(s)
		}
		return s
	}

	tamperedAcc := strings.Replace(fixtureAcceptance, "first", "First", 1)
	moreAcc := fixtureAcceptance + "\n### AC-FIXTURE-003 — third\n\nGiven a change, then it passes.\n"
	cardBody := renderFixture(fixtureOpts{card: strPtr("t1235")})
	writeBody := renderFixture(fixtureOpts{write: []string{"internal/fixture/**", ".moai/specs/" + fixtureSpecID + "/**", "internal/**"}})
	accShaBody := strings.Replace(body, fixtureAcceptanceSHA256(), sha256Hex([]byte(tamperedAcc)), 1)

	type tcase struct {
		name       string
		contract   string
		acceptance *string // nil: acceptance.md absent
		wantExit   int
	}
	acc := func(s string) *string { return &s }
	cases := []tcase{
		{"valid human", withSig(body, human), acc(fixtureAcceptance), 0},
		{"valid receipt", withSig(body, receipt), acc(fixtureAcceptance), 0},
		// AC-CONTRACT-010
		{"acceptance one character changed", withSig(body, human), acc(tamperedAcc), 1},
		{"acceptance AC added", withSig(body, human), acc(moreAcc), 1},
		{"acceptance deleted", withSig(body, human), nil, 1},
		// AC-CONTRACT-011 — body
		{"ownership.write gains internal/**", withSig(writeBody, human), acc(fixtureAcceptance), 1},
		{"acceptance.sha256 edited to a tampered acceptance.md", withSig(accShaBody, human), acc(tamperedAcc), 1},
		{"card edited", withSig(cardBody, human), acc(fixtureAcceptance), 1},
		{"card and contract_sha256 edited without the seal",
			withSig(cardBody, editSig(human, false, func(s *Signature) { s.ContractSHA256 = bodyDigest(cardBody) })), acc(fixtureAcceptance), 1},
	}
	// AC-CONTRACT-011 — one signature field edited without recomputing the seal.
	for name, e := range map[string]struct {
		base Signature
		edit func(*Signature)
	}{
		"signed_at":          {human, func(s *Signature) { s.SignedAt = "2020-01-01T00:00:00Z" }},
		"head_sha":           {human, func(s *Signature) { s.HeadSHA = strings.Repeat("0", 40) }},
		"operator.email":     {human, func(s *Signature) { s.Operator.Email = "other@example.com" }},
		"batch_id":           {human, func(s *Signature) { s.BatchID = "forged-batch" }},
		"supersedes":         {human, func(s *Signature) { s.Supersedes = other64 }},
		"acceptance_sha256":  {human, func(s *Signature) { s.AcceptanceSHA256 = other64 }},
		"method":             {human, func(s *Signature) { s.Method = MethodReceipt }},
		"signer_kind":        {human, func(s *Signature) { s.SignerKind = "llm" }},
		"receipt.provenance": {receipt, func(s *Signature) { s.Receipt.Provenance = "store" }},
		"receipt.path":       {receipt, func(s *Signature) { s.Receipt.Path = "other-receipt.json" }},
	} {
		cases = append(cases, tcase{"unsealed " + name, withSig(body, editSig(e.base, false, e.edit)), acc(fixtureAcceptance), 1})
	}
	// AC-CONTRACT-011 — edits with the seal recomputed: (a)-(d) invalid, (e) valid.
	cases = append(cases,
		tcase{"resealed (a) receipt method interactive-tty",
			withSig(body, editSig(receipt, true, func(s *Signature) { s.Method = MethodInteractiveTTY })), acc(fixtureAcceptance), 1},
		tcase{"resealed (b) receipt provenance store",
			withSig(body, editSig(receipt, true, func(s *Signature) { s.Receipt.Provenance = "store" })), acc(fixtureAcceptance), 1},
		tcase{"resealed (c) human signer_kind llm",
			withSig(body, editSig(human, true, func(s *Signature) { s.SignerKind = "llm" })), acc(fixtureAcceptance), 1},
		tcase{"resealed (d) human acceptance_sha256",
			withSig(body, editSig(human, true, func(s *Signature) { s.AcceptanceSHA256 = other64 })), acc(fixtureAcceptance), 1},
		tcase{"resealed (e) complete keyless re-seal",
			withSig(body, editSig(receipt, true, func(s *Signature) {
				s.Method, s.SignerKind, s.Receipt = MethodInteractiveTTY, "human", nil
			})), acc(fixtureAcceptance), 0},
	)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			specDir := filepath.Join(root, ".moai", "specs", fixtureSpecID)
			writeProjectFile(t, filepath.Join(specDir, ContractFile), tc.contract)
			if tc.acceptance != nil {
				writeProjectFile(t, filepath.Join(specDir, AcceptanceFile), *tc.acceptance)
			}
			writeProjectFile(t, filepath.Join(specDir, ReceiptFile), fixtureReceipt)
			writeProjectFile(t, filepath.Join(specDir, "spec.md"), "---\nid: "+fixtureSpecID+"\nstatus: in-progress\n---\n")
			writeProjectFile(t, filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"), "workflow: {}\n")

			before := hashTree(t, root)
			in, err := LoadDir(specDir)
			if err != nil {
				t.Fatalf("LoadDir: %v", err)
			}
			ref := fixtureInputs(nil)
			in.Policy, in.RegistryRuleIDs, in.SpecStatus = ref.Policy, ref.RegistryRuleIDs, ref.SpecStatus
			rep := Verify(in)
			exit := 1
			if rep.Valid {
				exit = 0
			}
			if exit != tc.wantExit {
				t.Errorf("verify: want exit %d, got %d (state %s, reasons %v)", tc.wantExit, exit, rep.State, rep.Reasons)
			}
			after := hashTree(t, root)
			if !mapsEqual(before, after) {
				t.Errorf("verify changed the fixture project:\nbefore %v\nafter  %v", before, after)
			}
		})
	}
}

func writeProjectFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// hashTree records the SHA-256 of every file under root (and the entry kind
// of everything else), keyed by slash-relative path.
func hashTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !d.Type().IsRegular() {
			out[rel] = "kind:" + d.Type().String()
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("hash tree: %v", err)
	}
	return out
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		if bv, ok := b[k]; !ok || bv != a[k] {
			return false
		}
	}
	return true
}
