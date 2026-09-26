package contract

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrContractMissing: <specDir>/contract.yaml does not exist.
	ErrContractMissing = errors.New("contract: contract.yaml not found")
	// ErrPathEscapesSpecDir: a SPEC file resolves outside its SPEC directory.
	ErrPathEscapesSpecDir = errors.New("contract: path escapes the SPEC directory")
	// ErrInvalidSpecID: a SPEC ID does not match SpecIDPattern.
	ErrInvalidSpecID = errors.New("contract: invalid SPEC ID")
)

// ResolveSpecDir returns <projectRoot>/.moai/specs/<specID>. The ID is
// validated against SpecIDPattern first, so a caller-supplied ID can never
// name a path outside `.moai/specs/`.
//
// @MX:ANCHOR: [AUTO] Path boundary for caller-supplied SPEC IDs.
// @MX:REASON: The contract CLI, the signer, and the signtest harness resolve SPEC directories here; the ID check is what keeps a path inside .moai/specs/.
func ResolveSpecDir(projectRoot, specID string) (string, error) {
	if !ValidSpecID(specID) {
		return "", fmt.Errorf("%w: %q", ErrInvalidSpecID, specID)
	}
	return filepath.Join(projectRoot, ".moai", "specs", specID), nil
}

// LoadDir reads the contract inputs of one SPEC directory: contract.yaml
// (required), acceptance.md and kickoff-receipt.json (optional; absence is
// recorded in the *Present flags). SpecID is the directory's base name.
//
// Every file is resolved through symlinks and must stay inside the resolved
// SPEC directory; one that resolves outside it is ErrPathEscapesSpecDir.
// All returned errors are I/O errors (the CLI maps them to exit 2), never
// verify reasons. LoadDir leaves Policy, the registry values, and SpecStatus
// for the caller to fill.
//
// @MX:ANCHOR: [AUTO] The one reader of a SPEC's contract inputs.
// @MX:REASON: The contract CLI, the signer, and the signtest harness load through it; its symlink containment is the ErrPathEscapesSpecDir guarantee.
func LoadDir(specDir string) (Inputs, error) {
	abs, err := filepath.Abs(specDir)
	if err != nil {
		return Inputs{}, fmt.Errorf("contract: resolve %s: %w", specDir, err)
	}
	realDir, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return Inputs{}, fmt.Errorf("contract: SPEC directory %s: %w", abs, err)
	}

	in := Inputs{SpecID: filepath.Base(abs)}
	data, ok, err := readContained(abs, realDir, ContractFile)
	if err != nil {
		return Inputs{}, err
	}
	if !ok {
		return Inputs{}, fmt.Errorf("%w: %s", ErrContractMissing, filepath.Join(abs, ContractFile))
	}
	in.Contract = data

	if in.Acceptance, in.AcceptancePresent, err = readContained(abs, realDir, AcceptanceFile); err != nil {
		return Inputs{}, err
	}
	if in.Receipt, in.ReceiptPresent, err = readContained(abs, realDir, ReceiptFile); err != nil {
		return Inputs{}, err
	}
	return in, nil
}

// readContained reads dir/name. It reports ok=false (no error) when the name
// does not exist at all, and an error when the name exists but resolves
// outside realDir, dangles, or is not a regular file.
func readContained(dir, realDir, name string) ([]byte, bool, error) {
	p := filepath.Join(dir, name)
	if _, err := os.Lstat(p); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("contract: stat %s: %w", p, err)
	}
	target, err := filepath.EvalSymlinks(p)
	if err != nil {
		return nil, false, fmt.Errorf("contract: resolve %s: %w", p, err)
	}
	rel, err := filepath.Rel(realDir, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, false, fmt.Errorf("%w: %s -> %s", ErrPathEscapesSpecDir, p, target)
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, false, fmt.Errorf("contract: stat %s: %w", target, err)
	}
	if !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("contract: %s is not a regular file", p)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, false, fmt.Errorf("contract: read %s: %w", p, err)
	}
	return data, true, nil
}
