package fixtures

// Violation Go fixture — demonstrates patterns matched by the shipped
// ruleset (.moai/astgrep-rules via sgconfig.yml). This file should produce
// >= 1 finding when scanned. Verified matchers on sg 0.40.5:
//   - go-interface-empty-not-any  (go/idioms.yml)
//   - sec-weak-hash-md5           (security/crypto.yml)
//   - sec-log-injection-unsanitized (security/web.yml)

import (
	"crypto/md5"
	"log"
)

// Dirty trips several shipped rules on purpose.
func Dirty(userInput string) {
	var payload interface{}
	payload = userInput
	_ = payload

	h := md5.New()
	_ = h

	log.Printf("user=%s", userInput)
}
