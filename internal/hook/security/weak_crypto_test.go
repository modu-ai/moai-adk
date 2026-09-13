package security

import (
	"regexp"
	"testing"
)

// t683 / ISSUE #1708: the weak-crypto class matched the BARE words md5/sha1/ECB
// (patterns.go `(?i)\b(MD5|SHA1)\b` / `(?i)\bECB\b`) with no file-type filter in
// ScanBuffer, so a verdict or evidence log that merely MENTIONS the words — e.g.
// a report recording that file identity was confirmed by md5 — surfaced as a
// weak-crypto finding. The narrowing direction is CALL CONTEXT, never wholesale
// prose-file exclusion: `md5(password)` in code MUST still be detected.
//
// These tests lock both directions of that boundary:
//
//	detect   — Go/Python/JS/Java/PHP/Ruby/C#/Rust code that USES a weak hash or
//	           ECB cipher mode still produces a weak-crypto finding;
//	silence  — prose / evidence-log lines that only MENTION the words produce zero
//	           findings;
//	preserve — the true-positive count over the code corpus does not shrink
//	           against the retired bare-word baseline (the old regexes are
//	           re-derived here, so the before/after comparison runs on every
//           execution, not just once).

// weakCryptoCodeCorpus is one true-positive detection site per line, in the
// idiomatic weak-crypto form of each language the card names.
var weakCryptoCodeCorpus = []string{
	`import "crypto/md5"`,                   // Go import
	`h := md5.New()`,                        // Go constructor
	`sum := sha1.Sum(data)`,                 // Go hash call
	`h = hashlib.md5(password).hexdigest()`, // Python hashlib
	`const h = crypto.createHash('md5').update(token).digest('hex');`, // JS createHash
	`MessageDigest md = MessageDigest.getInstance("MD5");`,            // Java named algorithm
	`$tok = md5($password);`,                                          // PHP bare call
	`Digest::MD5.hexdigest(pw)`,                                       // Ruby digest module
	`aes.Mode = CipherMode.ECB;`,                                      // C# cipher mode
	`cipher = AES.new(key, AES.MODE_ECB)`,                             // PyCryptodome ECB mode
	`Cipher c = Cipher.getInstance("AES/ECB/PKCS5Padding");`,          // Java cipher spec
	`let mut h = Md5::new();`,                                         // Rust md-5 crate
}

// weakCryptoProseCorpus is the defect's false-positive surface: prose and
// evidence-log lines that mention the words without calling anything.
var weakCryptoProseCorpus = []string{
	`파일 동일성 확인은 md5 해시로 기록했다 (t663 판정서).`,
	`MD5 and SHA1 are considered weak for password storage.`,
	`The ECB mode of operation is discouraged in new designs.`,
	`[2026-09-12] verdict.md: 동일성 md5 기록 — sha1로 재측정, ECB 언급 없음`,
	`# Weak crypto: MD5 vs SHA1 (RFC 6151)`,
	`MD5 (Message-Digest Algorithm 5) predates SHA1.`,
}

// TestWeakCryptoProseSilent is the defect direction: prose and evidence-log
// mentions of md5/sha1/ECB must produce ZERO weak-crypto findings.
func TestWeakCryptoProseSilent(t *testing.T) {
	for _, line := range weakCryptoProseCorpus {
		for _, f := range ScanBuffer(line) {
			if f.Class == "weak-crypto" {
				t.Errorf("prose line raised a weak-crypto finding (false positive): %q -> %+v", line, f)
			}
		}
	}
}

// TestWeakCryptoCodeDetected is the narrowing guard: code that uses a weak hash
// or ECB mode in call context must still be detected in every named language.
func TestWeakCryptoCodeDetected(t *testing.T) {
	for _, line := range weakCryptoCodeCorpus {
		hit := false
		for _, f := range ScanBuffer(line) {
			if f.Class == "weak-crypto" {
				hit = true
				break
			}
		}
		if !hit {
			t.Errorf("weak-crypto code line went undetected: %q", line)
		}
	}
}

// TestWeakCryptoTruePositivesPreserved is the card's before/after obligation:
// the detection count over the code corpus must not shrink against the retired
// bare-word baseline. The old regexes are re-derived here verbatim, so the
// comparison executes on every run instead of resting on a recorded number.
func TestWeakCryptoTruePositivesPreserved(t *testing.T) {
	oldBare := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(MD5|SHA1)\b`),
		regexp.MustCompile(`(?i)\bECB\b`),
	}
	oldCount := 0
	for _, line := range weakCryptoCodeCorpus {
		for _, re := range oldBare {
			if re.MatchString(line) {
				oldCount++
				break // one baseline hit per fixture line
			}
		}
	}
	newCount := 0
	for _, line := range weakCryptoCodeCorpus {
		for _, f := range ScanBuffer(line) {
			if f.Class == "weak-crypto" {
				newCount++
				break // one detection site per fixture line
			}
		}
	}
	if newCount < oldCount {
		t.Errorf("true-positive count shrank: old bare-word baseline %d, new call-context count %d — the narrowing over-narrowed", oldCount, newCount)
	}
}
