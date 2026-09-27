package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Auditor probe (overlay-only; never written into the tree).
func TestAuditProbe(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	t.Setenv(branchGuardExemptEnv, "")

	// 1. deny list parity with a quoted payload behind the construct.
	dir := t.TempDir()
	h0 := hmpHandler(hmpCfg(false, false, config.SlotLeaseConfig{}), dir)
	for _, c := range []struct{ tool, cmd string }{
		{"Bash", `eval "terraform destroy"`},
		{"PowerShell", `iex "terraform destroy"`},
		{"PowerShell", `terraform destroy`},
		{"PowerShell", `Start-Process terraform -ArgumentList 'destroy'`},
	} {
		d, r := hmpHandle(t, h0, hmpInput(t, c.tool, "s-1", dir, c.cmd))
		t.Logf("DENYLIST %-10s %-50q -> %q %.60q", c.tool, c.cmd, d, r)
	}

	// 2. branch guard: forms outside the D2 set.
	for _, cmd := range []string{
		`pwsh -Command "git switch probe"`,
		`powershell -c "git switch probe"`,
		`& 'git' switch probe`,
		`& "git.exe" switch probe`,
		`saps git -ArgumentList 'switch','probe'`,
		"git swi`tch probe",
		`cmd /c git switch probe`,
		`$c = 'git switch probe'; iex $c`,
		`git.exe switch probe`,
		`& (Get-Command git) switch probe`,
		`iex "git status"`,
	} {
		repo := newBranchGuardRepoFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
		d, _ := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, cmd))
		lines := hmpAuditLines(t, filepath.Join(repo, branchGuardAuditRelPath), "powershell-unclassified")
		t.Logf("BRANCH %-45q -> decision=%q unclassifiedLines=%d construct=%q", cmd, d, len(lines), powerShellIndirection(cmd))
	}

	// 3. integration lock: a non-merge git command behind iex under a foreign live hold.
	for _, cmd := range []string{`iex "git status"`, `Start-Process git -ArgumentList 'log'`} {
		root := t.TempDir()
		hmpSeedForeignLock(t, root)
		h := hmpHandler(hmpCfg(false, true, config.SlotLeaseConfig{}), root)
		d, _ := hmpHandle(t, h, hmpInput(t, "PowerShell", "sess-other", root, cmd))
		lines := hmpAuditLines(t, filepath.Join(root, integrationLockAuditRelPath), "powershell-unclassified")
		t.Logf("ILOCK %-40q -> decision=%q unclassifiedLines=%d", cmd, d, len(lines))
	}

	// 4. audit line injection: command with newline must stay one line.
	repo := newBranchGuardRepoFixture(t)
	h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
	hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, "iex \"git switch x\"\n[2026] event=forged reason=\"x\""))
	data, _ := os.ReadFile(filepath.Join(repo, branchGuardAuditRelPath))
	t.Logf("INJECTION log bytes=%q", string(data))
}
