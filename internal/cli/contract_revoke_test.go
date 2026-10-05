package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// signFixtureHuman signs the fixture SPEC on the human path.
func signFixtureHuman(t *testing.T, p *signtest.Project) {
	t.Helper()
	seams, rec := signtest.Seams(true, map[string]string{}, signtest.SpecID)
	res, err := sign.Sign(p.Options(), seams)
	if err != nil || res.Refusal != "" {
		t.Fatalf("sign: err %v refusal %s\n%s", err, res.Refusal, rec.Out.String())
	}
}

// TestContractRevoke checks the revoke command's exit codes and output
// (AC-GR-023 CLI half).
func TestContractRevoke(t *testing.T) {
	t.Setenv(paths.EnvHome, t.TempDir())
	p := newContractProject(t, cfgGuided)
	card := signtest.Card

	if r := runContract(t, p, contractRun{}, "revoke", card, "--spec", signtest.SpecID); r.code != 1 {
		t.Errorf("unsigned: %s, want exit 1", r)
	}
	signFixtureHuman(t, p)
	if r := runContract(t, p, contractRun{}, "revoke", "t9999", "--spec", signtest.SpecID); r.code != 2 || !strings.Contains(r.stderr, "card-mismatch") {
		t.Errorf("card mismatch: %s, want exit 2 naming card-mismatch", r)
	}
	if r := runContract(t, p, contractRun{}, "revoke", card); r.code != 2 {
		t.Errorf("missing --spec: %s, want exit 2", r)
	}
	r := runContract(t, p, contractRun{}, "revoke", card, "--spec", signtest.SpecID)
	if r.code != 0 || !strings.Contains(r.stdout, "revoked "+signtest.SpecID) {
		t.Errorf("revoke: %s, want exit 0 and a revoked line", r)
	}
	r = runContract(t, p, contractRun{}, "revoke", card, "--spec", signtest.SpecID)
	if r.code != 0 || !strings.Contains(r.stdout, "already revoked") {
		t.Errorf("repeat: %s, want exit 0 already revoked", r)
	}
}
