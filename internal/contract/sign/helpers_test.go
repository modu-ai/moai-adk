package sign_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// noMarkers is an environment in which no agent marker is set.
var noMarkers = map[string]string{}

// humanSeams returns human-path seams: TTY true, markers cleared, the given
// confirmation lines.
func humanSeams(lines ...string) (sign.Seams, *signtest.Recorder) {
	return signtest.Seams(true, noMarkers, lines...)
}

// mustSign runs Sign and fails the test on an error or a refusal.
func mustSign(t *testing.T, opts sign.Options, seams sign.Seams, rec *signtest.Recorder) sign.Result {
	t.Helper()
	res, err := sign.Sign(opts, seams)
	if err != nil {
		t.Fatalf("Sign: unexpected error: %v\noutput:\n%s", err, rec.Out.String())
	}
	if res.Refusal != "" {
		t.Fatalf("Sign: unexpected refusal %s (%s)\noutput:\n%s", res.Refusal, res.Cause, rec.Out.String())
	}
	return res
}

// assertRefuses runs Sign and asserts it refuses with code, prints the code,
// and leaves every fixture file byte-identical.
func assertRefuses(t *testing.T, p *signtest.Project, opts sign.Options, seams sign.Seams, rec *signtest.Recorder, code string) sign.Result {
	t.Helper()
	snap := p.Snapshot()
	res, err := sign.Sign(opts, seams)
	if err != nil {
		t.Fatalf("Sign: unexpected error %v, want refusal %s\noutput:\n%s", err, code, rec.Out.String())
	}
	if res.Refusal != code {
		t.Errorf("refusal = %q (%s), want %q\noutput:\n%s", res.Refusal, res.Cause, code, rec.Out.String())
	}
	if !contract.IsSignRefusalCode(res.Refusal) {
		t.Errorf("refusal %q is not a sign refusal code", res.Refusal)
	}
	if !strings.Contains(rec.Out.String(), code) {
		t.Errorf("output does not name %s:\n%s", code, rec.Out.String())
	}
	if len(res.Written) != 0 {
		t.Errorf("refusal wrote files: %v", res.Written)
	}
	p.AssertUnchanged(t, snap)
	return res
}

// decodeSigned reads and strictly decodes a SPEC's contract.yaml.
func decodeSigned(t *testing.T, p *signtest.Project, id string) *contract.Contract {
	t.Helper()
	c, err := contract.Decode(p.ReadFile(signtest.SpecRel(id, contract.ContractFile)))
	if err != nil {
		t.Fatalf("decode signed contract: %v", err)
	}
	return c
}

// assertSignedValid asserts verify reports signed-valid with no reasons.
func assertSignedValid(t *testing.T, p *signtest.Project, id string, opts sign.Options) {
	t.Helper()
	rep := p.Verify(id, opts)
	if rep.State != contract.StateSignedValid || len(rep.Reasons) != 0 {
		t.Errorf("verify %s: state %s reasons %v, want signed-valid []", id, rep.State, rep.Reasons)
	}
}
