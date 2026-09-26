package sign_test

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

var contractRel = signtest.SpecRel(signtest.SpecID, contract.ContractFile)

func TestSign_KeepsCRLFLineEndings(t *testing.T) {
	p := signtest.New(t)
	crlf := strings.ReplaceAll(signtest.DraftContract(signtest.Draft{}), "\n", "\r\n")
	p.WriteFile(contractRel, strings.TrimSuffix(crlf, "\r\n")) // no final line ending
	seams, rec := humanSeams(signtest.SpecID)
	mustSign(t, p.Options(), seams, rec)
	out := string(p.ReadFile(contractRel))
	if strings.Count(out, "\n") != strings.Count(out, "\r\n") {
		t.Errorf("signed file mixes line endings:\n%q", out)
	}
	assertSignedValid(t, p, signtest.SpecID, p.Options())
}

func TestSign_FillsNullBudget(t *testing.T) {
	for _, null := range []string{"budget:", "budget: ~"} {
		t.Run(null, func(t *testing.T) {
			p := signtest.New(t)
			p.WriteFile(contractRel, signtest.DraftContract(signtest.Draft{})+null+"\n")
			seams, rec := humanSeams(signtest.SpecID)
			mustSign(t, p.Options(), seams, rec)
			c := decodeSigned(t, p, signtest.SpecID)
			if c.Budget == nil || *c.Budget != p.Options().BudgetDefault {
				t.Errorf("budget = %v, want the default", c.Budget)
			}
			assertSignedValid(t, p, signtest.SpecID, p.Options())
		})
	}
}

func TestSign_FlowStyleAcceptanceWithRecordedHash(t *testing.T) {
	p := signtest.New(t)
	sha := contract.AcceptanceHash([]byte(signtest.Acceptance))
	text := strings.Replace(signtest.DraftContract(signtest.Draft{}),
		"acceptance:\n  file: acceptance.md   # the SPEC's acceptance criteria\n",
		"acceptance: {file: acceptance.md, sha256: \""+sha+"\", ac_count: 7}\n", 1)
	p.WriteFile(contractRel, text)
	seams, rec := humanSeams(signtest.SpecID)
	mustSign(t, p.Options(), seams, rec)
	c := decodeSigned(t, p, signtest.SpecID)
	if *c.Acceptance.SHA256 != sha || *c.Acceptance.ACCount != 2 {
		t.Errorf("binding = %s / %d, want %s / 2", *c.Acceptance.SHA256, *c.Acceptance.ACCount, sha)
	}
	assertSignedValid(t, p, signtest.SpecID, p.Options())
}

func TestSign_MissingAcceptanceRefusesVerifyFailed(t *testing.T) {
	p := signtest.New(t)
	if err := os.Remove(p.Path(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile))); err != nil {
		t.Fatal(err)
	}
	seams, rec := humanSeams(signtest.SpecID)
	res := assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseVerifyFailed)
	if !strings.Contains(strings.Join(res.Reasons, ","), contract.ReasonAcceptanceMissing) {
		t.Errorf("reasons = %v, want acceptance_missing", res.Reasons)
	}
}

func TestSign_UndecodableContractRefusesVerifyFailed(t *testing.T) {
	p := signtest.New(t)
	p.WriteFile(contractRel, signtest.DraftContract(signtest.Draft{})+"notes: x\n")
	seams, rec := humanSeams(signtest.SpecID)
	res := assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseVerifyFailed)
	if strings.Join(res.Reasons, ",") != contract.ReasonSchemaInvalid {
		t.Errorf("reasons = %v, want [schema_invalid]", res.Reasons)
	}
}

func TestSign_GitSeamErrors(t *testing.T) {
	p := signtest.New(t)
	t.Run("identity_error", func(t *testing.T) {
		seams, _ := humanSeams(signtest.SpecID)
		seams.GitIdentity = func(string) (string, string, error) { return "", "", errors.New("git exploded") }
		snap := p.Snapshot()
		if _, err := sign.Sign(p.Options(), seams); err == nil {
			t.Error("want an error from a failing git identity")
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("head_error", func(t *testing.T) {
		seams, _ := humanSeams(signtest.SpecID)
		seams.GitIdentity = func(string) (string, string, error) { return "n", "e@x", nil }
		seams.GitHead = func(string) (string, error) { return "", errors.New("no HEAD") }
		snap := p.Snapshot()
		if _, err := sign.Sign(p.Options(), seams); err == nil {
			t.Error("want an error from a failing HEAD read")
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("default_head_outside_repository", func(t *testing.T) {
		q := signtest.New(t)
		if err := os.RemoveAll(q.Path(".git")); err != nil {
			t.Fatal(err)
		}
		seams, _ := humanSeams(signtest.SpecID)
		seams.GitIdentity = func(string) (string, string, error) { return "n", "e@x", nil }
		if _, err := sign.Sign(q.Options(), seams); err == nil {
			t.Error("want an error when HEAD cannot be read")
		}
	})
}

func TestSign_DefaultWriterFailureReportsSplit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits do not block file creation on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permission bits")
	}
	p := signtest.New(t)
	dir := p.Path(".moai/specs/" + signtest.SpecID)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	seams, rec := humanSeams(signtest.SpecID)
	snap := p.Snapshot()
	res, err := sign.Sign(p.Options(), seams)
	if err == nil {
		t.Fatalf("want a write error, got %+v", res)
	}
	if len(res.Written) != 0 || len(res.Unwritten) != 1 {
		t.Errorf("split = %v / %v, want none / one", res.Written, res.Unwritten)
	}
	if !strings.Contains(rec.Out.String(), "written: none") {
		t.Errorf("output lacks the split:\n%s", rec.Out.String())
	}
	p.AssertUnchanged(t, snap)
}

// TestSign_DefaultStdinSeams swaps os.Stdin for a pipe: the default terminal
// check reports false for it, and the default line reader reads from it.
func TestSign_DefaultStdinSeams(t *testing.T) {
	withStdin := func(t *testing.T, input string) {
		t.Helper()
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.WriteString(input); err != nil {
			t.Fatal(err)
		}
		_ = w.Close()
		old := os.Stdin
		os.Stdin = r
		t.Cleanup(func() { os.Stdin = old; _ = r.Close() })
	}

	t.Run("pipe_is_not_a_terminal", func(t *testing.T) {
		p := signtest.New(t)
		withStdin(t, signtest.SpecID+"\n")
		seams, rec := humanSeams()
		seams.IsTTY, seams.ReadLine = nil, nil
		assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseNotTTY)
	})
	t.Run("null_device_is_not_a_terminal", func(t *testing.T) {
		p := signtest.New(t)
		devNull, err := os.Open(os.DevNull)
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdin
		os.Stdin = devNull
		t.Cleanup(func() { os.Stdin = old; _ = devNull.Close() })
		seams, rec := humanSeams()
		seams.IsTTY, seams.ReadLine = nil, nil
		assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseNotTTY)
	})
	for name, input := range map[string]string{"with_newline": signtest.SpecID + "\r\n", "without_newline": signtest.SpecID} {
		t.Run("line_reader_"+name, func(t *testing.T) {
			p := signtest.New(t)
			withStdin(t, input)
			seams, rec := humanSeams()
			seams.ReadLine = nil
			mustSign(t, p.Options(), seams, rec)
		})
	}
	t.Run("line_reader_empty_stream", func(t *testing.T) {
		p := signtest.New(t)
		withStdin(t, "")
		seams, rec := humanSeams()
		seams.ReadLine = nil
		assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseConfirmationMismatch)
	})
}

func TestSign_ReadLineErrorIsAnError(t *testing.T) {
	p := signtest.New(t)
	seams, _ := humanSeams()
	seams.ReadLine = func() (string, error) { return "", errors.New("terminal gone") }
	snap := p.Snapshot()
	if _, err := sign.Sign(p.Options(), seams); err == nil {
		t.Error("want an error when the confirmation cannot be read")
	}
	p.AssertUnchanged(t, snap)
}

func TestSign_ResignOnReceiptPath(t *testing.T) {
	p := signtest.New(t)
	opts := p.ReceiptOptions("llm", "llm")
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec := signtest.Seams(false, noMarkers)
	mustSign(t, opts, seams, rec)
	d0 := decodeSigned(t, p, signtest.SpecID).Signature.ContractSHA256

	p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile), signtest.AcceptanceThree)
	opts.Resign = true
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec = signtest.Seams(false, noMarkers)
	mustSign(t, opts, seams, rec)
	s := decodeSigned(t, p, signtest.SpecID).Signature
	if s.Supersedes != d0 || s.Method != contract.MethodReceipt {
		t.Errorf("re-signed receipt signature = %+v, want supersedes %s", s, d0)
	}
	assertSignedValid(t, p, signtest.SpecID, opts)
}

// TestSign_ResignRefusesTamperedBodyOnReceiptPath: the receipt path tolerates
// receipt_mismatch against the old signature, but not a body edited after
// signing — a fresh receipt over the edited body must not re-approve it.
func TestSign_ResignRefusesTamperedBodyOnReceiptPath(t *testing.T) {
	p := signtest.New(t)
	opts := p.ReceiptOptions("llm", "llm")
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec := signtest.Seams(false, noMarkers)
	mustSign(t, opts, seams, rec)

	rel := signtest.SpecRel(signtest.SpecID, contract.ContractFile)
	body := string(p.ReadFile(rel))
	const line = "    - \"internal/fixture/**\"\n"
	tampered := strings.Replace(body, line, line+"    - \"internal/widened/**\"\n", 1)
	if tampered == body {
		t.Fatalf("fixture has no ownership.write line %q to edit", line)
	}
	p.WriteFile(rel, tampered)
	opts.Resign = true
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec = signtest.Seams(false, noMarkers)
	assertRefuses(t, p, opts, seams, rec, contract.RefuseVerifyFailed)
	if !strings.Contains(rec.Out.String(), contract.ReasonContractDigestMismatch) {
		t.Errorf("the refusal must name %s:\n%s", contract.ReasonContractDigestMismatch, rec.Out.String())
	}
}
