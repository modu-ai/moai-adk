package contract

import "errors"

var (
	// ErrContractMissing: <specDir>/contract.yaml does not exist.
	ErrContractMissing = errors.New("contract: contract.yaml not found")
	// ErrPathEscapesSpecDir: a SPEC file resolves outside its SPEC directory.
	ErrPathEscapesSpecDir = errors.New("contract: path escapes the SPEC directory")
	// ErrInvalidSpecID: a SPEC ID does not match SpecIDPattern.
	ErrInvalidSpecID = errors.New("contract: invalid SPEC ID")
)

// ResolveSpecDir returns <projectRoot>/.moai/specs/<specID>.
func ResolveSpecDir(projectRoot, specID string) (string, error) {
	return "", errors.New("not implemented")
}

// LoadDir reads the SPEC directory's contract inputs.
func LoadDir(specDir string) (Inputs, error) {
	return Inputs{}, errors.New("not implemented")
}
