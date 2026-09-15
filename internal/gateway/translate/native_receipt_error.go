package translate

// Native identity denials are deterministic input failures, not broken network
// connections. Keep the reason constant and never expose the authority's error.
type nativeReceiptAuthorizationError struct{}

func (nativeReceiptAuthorizationError) Error() string { return "native receipt authorization required" }
func (nativeReceiptAuthorizationError) CauseCode() string {
	return "native_receipt_authorization_required"
}
