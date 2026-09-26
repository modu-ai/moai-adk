package contract

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Acceptance binding (REQ-CONTRACT-005; design.md § Acceptance hash and count).
//
// The AC counter below is a Go port of the published awk counter, the
// MOAI-AC-COUNTER block in the manager-docs agent definition. A parity test in
// internal/spec (TestAC_CONTRACT_006) runs the extracted awk program and this
// port over the whole acceptance corpus plus per-branch fixtures, so the port
// cannot drift from the published program silently.
//
// Awk-to-Go correspondence, construct by construct:
//
//   - Records: awk splits on "\n" (default RS); a trailing newline yields no
//     extra record. strings.Split yields one extra empty record, which holds
//     no identifier and therefore changes nothing.
//   - Regex engine: awk `match()` is POSIX leftmost-longest; the port compiles
//     every pattern with regexp.Compile and calls Longest(). Leftmost-longest
//     also governs `sub()`, which replaces only the first match.
//   - The prefix is inserted unescaped as a regular-expression fragment, as
//     awk does — `AC|FLH` is an alternation, and a comma-separated list such as
//     `CR, FLH` becomes the literal `CR,FLH` after space removal (awk does the
//     same; the agent-definition prose calls it a list, the program does not).
//   - Syntax gaps between awk ERE and Go RE2 are reachable only through the
//     declared prefix. Go accepts some forms awk ERE does not treat the same
//     way (`\d`, `\w`, `\b`, `(?i)`, `(?:…)`, lazy quantifiers), and awk
//     implementations differ among themselves on interval expressions. A
//     prefix using such forms may count differently; every prefix in the
//     corpus is a plain letter-digit token, which both engines read
//     identically. A prefix Go cannot compile yields ErrACPrefixInvalid — the
//     awk program aborts on an unbalanced prefix too, so neither side produces
//     a count.
//   - Offsets: awk RSTART/RLENGTH count characters, Go indices count bytes.
//     The identifier grammar and the marker test are ASCII, so the matched
//     identifier and the remainder are the same text either way.

// ErrACPrefixInvalid reports that a file's `moai-ac-prefix` declaration does
// not compile as a regular-expression fragment, so no count can be produced.
var ErrACPrefixInvalid = errors.New("contract: moai-ac-prefix declaration does not compile")

// ACCountResult is the outcome of CountAC.
type ACCountResult struct {
	// Prefix is the effective prefix fragment (default "AC").
	Prefix string
	// Live counts distinct IDs never followed by a [RETIRED]/[REF] marker.
	Live int
	// Excluded counts distinct IDs only ever followed by a marker.
	Excluded int
	// Ambiguous lists IDs seen both marked and unmarked, in first-seen order.
	// A non-empty list is the awk counter's exit-3 case: the file has no
	// usable count until every such occurrence is resolved.
	Ambiguous []string
}

// IsAmbiguous reports whether any ID was both marked and unmarked.
func (r ACCountResult) IsAmbiguous() bool { return len(r.Ambiguous) > 0 }

const defaultACPrefix = "AC"

var (
	utf8BOM = []byte{0xEF, 0xBB, 0xBF}
	crlf    = []byte("\r\n")
	lf      = []byte("\n")

	prefixDeclLineRe  = longest(`^<!-- *moai-ac-prefix:`)
	prefixDeclStripRe = longest(`^<!-- *moai-ac-prefix: *`)
	prefixDeclTailRe  = longest(` *-->.*$`)
	acMarkerRe        = longest(`^[ \t]*(\[RETIRED\]|\[REF\])`)
)

func longest(expr string) *regexp.Regexp {
	re := regexp.MustCompile(expr)
	re.Longest()
	return re
}

// NormalizeAcceptance strips one leading UTF-8 BOM and replaces every CRLF
// with LF. The input is never modified; the result never aliases it.
//
// @MX:ANCHOR: [AUTO] Normalization under the acceptance hash and AC count.
// @MX:REASON: AcceptanceHash, the Verify AC rule (rules.go), and the signer all normalize here; any change moves every signed acceptance_sha256.
func NormalizeAcceptance(raw []byte) []byte {
	b := bytes.TrimPrefix(raw, utf8BOM)
	return bytes.ReplaceAll(b, crlf, lf)
}

// AcceptanceHash returns the lowercase-hex SHA-256 of the normalized bytes.
//
// @MX:ANCHOR: [AUTO] The acceptance hash a signed contract binds to.
// @MX:REASON: Consumed by Verify (acceptance_hash_mismatch), by `sign` when it
// records acceptance.sha256, and by the receipt inputs; a change to the
// normalization silently invalidates every signed contract.
func AcceptanceHash(raw []byte) string {
	sum := sha256.Sum256(NormalizeAcceptance(raw))
	return hex.EncodeToString(sum[:])
}

// CountAC ports the published awk AC counter. It expects normalized content
// (NormalizeAcceptance); on raw CRLF or BOM-prefixed bytes it behaves exactly
// as the awk program does on those bytes.
//
// @MX:ANCHOR: [AUTO] Go port of the MOAI-AC-COUNTER awk program.
// @MX:REASON: Consumed by Verify (ac_count_mismatch / ac_count_ambiguous), by
// `sign` when it records acceptance.ac_count, and pinned to the awk program by
// the internal/spec parity test; semantics must stay byte-for-byte faithful.
func CountAC(normalized []byte) (ACCountResult, error) {
	res := ACCountResult{Prefix: defaultACPrefix}
	idRe, err := compileACPattern(res.Prefix)
	if err != nil {
		return res, err
	}
	declared := false
	marked := map[string]bool{}
	unmarked := map[string]bool{}
	seen := map[string]bool{}
	var order []string

	for _, line := range strings.Split(string(normalized), "\n") {
		if !declared && prefixDeclLineRe.MatchString(line) {
			decl := subFirst(prefixDeclStripRe, line)
			decl = subFirst(prefixDeclTailRe, decl)
			decl = strings.ReplaceAll(decl, " ", "")
			if decl != "" {
				declared = true
				res.Prefix = decl
				if idRe, err = compileACPattern(decl); err != nil {
					return res, err
				}
			}
		}
		rest := line
		for {
			loc := idRe.FindStringIndex(rest)
			if loc == nil {
				break
			}
			id := rest[loc[0]:loc[1]]
			rest = rest[loc[1]:]
			if acMarkerRe.MatchString(rest) {
				marked[id] = true
			} else {
				unmarked[id] = true
			}
			if !seen[id] {
				seen[id] = true
				order = append(order, id)
			}
		}
	}

	for _, id := range order {
		switch {
		case marked[id] && unmarked[id]:
			res.Ambiguous = append(res.Ambiguous, id)
		case marked[id]:
			res.Excluded++
		default:
			res.Live++
		}
	}
	return res, nil
}

func compileACPattern(prefix string) (*regexp.Regexp, error) {
	re, err := regexp.Compile("(" + prefix + ")-([A-Z0-9]+-)*[0-9]+[a-z]?")
	if err != nil {
		return nil, fmt.Errorf("%w: %q: %v", ErrACPrefixInvalid, prefix, err)
	}
	re.Longest()
	return re, nil
}

// subFirst mirrors awk sub(re, "", s): remove the first (leftmost-longest)
// match, if any.
func subFirst(re *regexp.Regexp, s string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + s[loc[1]:]
}
