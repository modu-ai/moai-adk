package contract

import "errors"

// Canonical returns a deep copy of c in canonical form.
func Canonical(c Contract) Contract {
	return Contract{}
}

// Digest returns the canonical body digest of c.
func Digest(c *Contract) (string, error) {
	return "", errors.New("not implemented")
}

// DigestBytes strictly decodes raw and returns its canonical body digest.
func DigestBytes(raw []byte) (string, error) {
	return "", errors.New("not implemented")
}
