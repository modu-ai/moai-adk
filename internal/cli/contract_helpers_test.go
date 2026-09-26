package cli

// Test support for the `moai contract` AC tests (contract_ac_test.go). Every
// run is in-process against a signtest fixture project in t.TempDir(): the
// project root, the terminal check, the environment, and the confirmation
// reader are all replaced through package seams, so the runtime's own
// environment (a Claude Code session carries CLAUDECODE) never reaches the
// command, and nothing reads or writes the real .moai/specs.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// autoCfg renders a workflow.yaml for the fixture project.
type autoCfg struct {
	mode        string // guided | contract
	decider     string // "" omits the key
	batch       bool
	pushDevelop bool
	jevDisabled bool // writes workflow.jev.enabled: false
}

func (c autoCfg) yaml() string {
	var b strings.Builder
	b.WriteString("workflow:\n")
	if c.jevDisabled {
		b.WriteString("  jev:\n    enabled: false\n")
	}
	fmt.Fprintf(&b, "  autonomy:\n    mode: %s\n", c.mode)
	fmt.Fprintf(&b, "    contract:\n      batch_sign: %t\n      second_review: required\n      push_develop: %t\n",
		c.batch, c.pushDevelop)
	if c.decider != "" {
		fmt.Fprintf(&b, "    kickoff:\n      decider: %s\n", c.decider)
	}
	return b.String()
}

// Configurations used across the AC tests.
var (
	cfgGuided      = autoCfg{mode: "guided", pushDevelop: true}
	cfgContractLLM = autoCfg{mode: "contract", decider: "llm", pushDevelop: true}
)

const workflowYAMLRel = ".moai/config/sections/workflow.yaml"

// newContractProject returns a fixture project carrying cfg as its workflow.yaml.
func newContractProject(t *testing.T, cfg autoCfg) *signtest.Project {
	t.Helper()
	p := signtest.New(t)
	p.WriteFile(workflowYAMLRel, cfg.yaml())
	return p
}

// contractRun is one in-process `moai contract` invocation.
type contractRun struct {
	tty   bool              // TTY seam value
	env   map[string]string // the whole environment the command sees (nil: empty)
	stdin string            // the command's input stream
	// realTTY leaves stdinIsTerminalFn untouched (the real stdin check).
	realTTY bool
}

// contractResult is what the invocation produced.
type contractResult struct {
	stdout, stderr string
	code           int
	reads          int // confirmation-reader calls
}

func (r contractResult) String() string {
	return fmt.Sprintf("exit=%d reads=%d\n--- stdout ---\n%s--- stderr ---\n%s", r.code, r.reads, r.stdout, r.stderr)
}

// runContract executes `moai contract <args...>` in p.
func runContract(t *testing.T, p *signtest.Project, o contractRun, args ...string) contractResult {
	t.Helper()
	origRoot, origTTY, origEnv, origReader := findProjectRootFn, stdinIsTerminalFn, contractGetenvFn, newContractLineReader
	t.Cleanup(func() {
		findProjectRootFn, stdinIsTerminalFn, contractGetenvFn, newContractLineReader = origRoot, origTTY, origEnv, origReader
	})
	root := p.Root
	findProjectRootFn = func() (string, error) { return root, nil }
	if !o.realTTY {
		tty := o.tty
		stdinIsTerminalFn = func() bool { return tty }
	}
	env := o.env
	contractGetenvFn = func(k string) string { return env[k] }
	res := contractResult{}
	newContractLineReader = func(r io.Reader) func() (string, error) {
		inner := origReader(r)
		return func() (string, error) {
			res.reads++
			return inner()
		}
	}

	cmd := newContractCmd()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetIn(strings.NewReader(o.stdin))
	cmd.SetArgs(args)
	err := cmd.Execute()
	res.stdout, res.stderr = out.String(), errb.String()
	if err != nil {
		if code, ok := ResolveExitCode(err); ok {
			res.code = code
		} else {
			res.code = 1
			res.stderr += err.Error()
		}
	}
	return res
}

// humanRun is a human-path invocation: a terminal, no agent markers, and the
// given confirmation line.
func humanRun(token string) contractRun {
	return contractRun{tty: true, env: map[string]string{}, stdin: token + "\n"}
}

// signHuman signs ids on the human path and fails the test unless exit 0.
func signHuman(t *testing.T, p *signtest.Project, ids ...string) contractResult {
	t.Helper()
	if len(ids) == 0 {
		ids = []string{signtest.SpecID}
	}
	token := ids[0]
	if len(ids) > 1 {
		token = fmt.Sprintf("sign %d contracts", len(ids))
	}
	res := runContract(t, p, humanRun(token), append([]string{"sign"}, ids...)...)
	if res.code != 0 {
		t.Fatalf("human sign %v: want exit 0\n%s", ids, res)
	}
	return res
}

// showJSON runs `show <id> --json` and returns the parsed report.
func showJSON(t *testing.T, p *signtest.Project, id string) contract.Report {
	t.Helper()
	res := runContract(t, p, contractRun{env: map[string]string{}}, "show", id, "--json")
	if res.code != 0 {
		t.Fatalf("show --json: want exit 0\n%s", res)
	}
	var rep contract.Report
	if err := json.Unmarshal([]byte(res.stdout), &rep); err != nil {
		t.Fatalf("show --json is not a report object: %v\n%s", err, res)
	}
	return rep
}

// verifyJSON runs `verify <id> --json` and returns the result, the parsed
// report, and the raw object.
func verifyJSON(t *testing.T, p *signtest.Project, id string) (contractResult, contract.Report, map[string]any) {
	t.Helper()
	res := runContract(t, p, contractRun{env: map[string]string{}}, "verify", id, "--json")
	var rep contract.Report
	var raw map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &rep); err != nil {
		t.Fatalf("verify --json is not a report object: %v\n%s", err, res)
	}
	if err := json.Unmarshal([]byte(res.stdout), &raw); err != nil {
		t.Fatalf("verify --json is not a JSON object: %v\n%s", err, res)
	}
	return res, rep, raw
}

// wantVerify asserts verify --json's exit code and that reasons contain want.
func wantVerify(t *testing.T, p *signtest.Project, code int, want ...string) contract.Report {
	t.Helper()
	res, rep, _ := verifyJSON(t, p, signtest.SpecID)
	if res.code != code {
		t.Errorf("verify: want exit %d\n%s", code, res)
	}
	for _, w := range want {
		if !slices.Contains(rep.Reasons, w) {
			t.Errorf("verify: reasons %v do not contain %q", rep.Reasons, w)
		}
	}
	return rep
}

// contractRel is the fixture contract path.
func contractRel(id string) string { return signtest.SpecRel(id, contract.ContractFile) }

// decodeContract strictly decodes the current contract.yaml of id.
func decodeContract(t *testing.T, p *signtest.Project, id string) *contract.Contract {
	t.Helper()
	c, err := contract.Decode(p.ReadFile(contractRel(id)))
	if err != nil {
		t.Fatalf("decode %s: %v", id, err)
	}
	return c
}

// writeContract re-encodes c as the contract.yaml of id.
func writeContract(t *testing.T, p *signtest.Project, id string, c *contract.Contract) {
	t.Helper()
	data, err := yaml.Marshal(c)
	if err != nil {
		t.Fatalf("encode %s: %v", id, err)
	}
	p.WriteFile(contractRel(id), string(data))
}

// editContract decodes the fixture contract, applies edit, and writes it back.
func editContract(t *testing.T, p *signtest.Project, edit func(c *contract.Contract)) {
	t.Helper()
	c := decodeContract(t, p, signtest.SpecID)
	edit(c)
	writeContract(t, p, signtest.SpecID, c)
}

// reseal recomputes the signature seal of c (the seal only; contract_sha256
// is left as the edit set it).
func reseal(t *testing.T, c *contract.Contract) {
	t.Helper()
	seal, err := contract.ComputeSeal(*c.Signature)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	c.Signature.Seal = seal
}

// redigest records the fresh body digest in the signature and reseals it: a
// body edit that a signer re-signed, so only the body's own rules report.
func redigest(t *testing.T, c *contract.Contract) {
	t.Helper()
	d, err := contract.Digest(c)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	c.Signature.ContractSHA256 = d
	reseal(t, c)
}

// humanSigned returns a fixture project whose contract is signed on the human
// path under cfgGuided (second_review required, push_develop true).
func humanSigned(t *testing.T) *signtest.Project {
	t.Helper()
	p := newContractProject(t, cfgGuided)
	signHuman(t, p)
	return p
}

// writeValidReceipt writes a receipt for SpecID bound to show --json's
// signable_contract_sha256; mutate edits it further.
func writeValidReceipt(t *testing.T, p *signtest.Project, mutate func(*contract.KickoffReceipt)) {
	t.Helper()
	shown := showJSON(t, p, signtest.SpecID)
	if shown.SignableContractSHA256 == "" {
		t.Fatalf("show --json: empty signable_contract_sha256")
	}
	opts := p.ReceiptOptions("llm", "llm")
	data := p.Receipt(opts, func(r *contract.KickoffReceipt) {
		if r.Inputs.ContractSHA256 != shown.SignableContractSHA256 {
			t.Errorf("receipt input digest %s differs from show --json signable_contract_sha256 %s",
				r.Inputs.ContractSHA256, shown.SignableContractSHA256)
		}
		r.Inputs.ContractSHA256 = shown.SignableContractSHA256
		if mutate != nil {
			mutate(r)
		}
	})
	p.WriteReceipt(data)
}

// receiptArgs is `sign <SpecID> --signer <signer> --receipt <fixed path>`.
func receiptArgs(signer string) []string {
	return []string{"sign", signtest.SpecID, "--signer", signer, "--receipt", signtest.ReceiptRel()}
}

// receiptSigned returns a fixture project whose contract is signed on the
// receipt path (r1: mode contract, decider llm, --signer llm).
func receiptSigned(t *testing.T) *signtest.Project {
	t.Helper()
	p := newContractProject(t, cfgContractLLM)
	writeValidReceipt(t, p, nil)
	res := runContract(t, p, contractRun{env: map[string]string{}}, receiptArgs("llm")...)
	if res.code != 0 {
		t.Fatalf("receipt sign: want exit 0\n%s", res)
	}
	return p
}
