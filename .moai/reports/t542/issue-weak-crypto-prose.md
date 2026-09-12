## Summary

The Security Guardian's `weak-crypto` rule fires on the *word* `md5`, in any text, with no regard for whether a crypto API is being called. Prose that records an md5 file-identity check — the kind of line a verdict or an evidence log is made of — is reported as `weak-crypto (medium)`.

Originally observed by a lane while writing the card **t663** verdict: the line flagged was a record that a mutant had been restored and the restored file's md5 matched the committed copy's. No security use, no crypto API, no code. Independently reproduced below against the current source.

## Root cause, in two lines of source

The pattern is a bare word match (`internal/hook/security/patterns.go:157-165`):

```go
{
    Name:        "weak-crypto",
    Severity:    SevMedium,
    Description: "Weak hash or cipher mode for security-sensitive data",
    Patterns: []*regexp.Regexp{
        mp(`(?i)\b(MD5|SHA1)\b`), // md5/sha1 for password/token
        mp(`(?i)\bECB\b`),        // ECB cipher mode
    },
},
```

The scanner applies it to every buffer, with **no file-type filter** (`internal/hook/security/scan.go:26-62`). The only exclusions are a byte-size bound and a NUL-byte binary sentinel:

```go
func ScanBuffer(content string) []GuardianFinding {
	if content == "" { return nil }
	if len(content) > config.DefaultSecurityMaxScanBytes { content = content[:config.DefaultSecurityMaxScanBytes] }
	if looksBinary(content) { return nil }
	// ... every class, every pattern, every line
```

So Markdown, plain text, changelogs, and commit-message bodies are all scanned with rules written for source code. The comment beside the pattern (`// md5/sha1 for password/token`) and the description ("for security-sensitive data") both state the intended scope; the regex does not encode it.

Note the asymmetry with the ast-grep ruleset, which gets this right: `sec-weak-hash-md5` matches `pattern: md5.New()` — an actual API call — and would not fire on prose.

## Reproduction

```go
// tmpprobe/main.go, inside the module
package main

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

func main() {
	buf := "# verdict\n\n- Restore verified: the md5 of the restored file equals\n  the md5 of the committed copy; no security use here.\n"
	found := security.ScanBuffer(buf)
	fmt.Printf("findings: %d\n", len(found))
	for _, f := range found {
		fmt.Printf("%s (%s) line %d: %s | match=%q\n", f.Class, f.Severity, f.Line, f.Message, f.Match)
	}
}
```

```
$ go run ./tmpprobe
findings: 3
weak-crypto (medium) line 3: Weak hash or cipher mode for security-sensitive data | match="md5"
weak-crypto (medium) line 4: Weak hash or cipher mode for security-sensitive data | match="md5"
weak-crypto (medium) line 5: Weak hash or cipher mode for security-sensitive data | match="ECB"
```

(The buffer's fifth line is `- The ECB published the rate on Tuesday.` — a sentence about a central bank, reported as a cipher-mode finding.)

### It also fired on this issue

While this text was being written to disk, the in-session guardian reported it:

```
[MoAI Security Guardian] 15 finding(s): weak-crypto (medium) line 3: Weak hash or
cipher mode for security-sensitive data; weak-crypto (medium) line 5: …
… and 5 more
```

Fifteen medium findings, for a Markdown bug report *about* this rule. That is the clearest statement of the problem available: the document explaining that the rule cannot tell prose from crypto was itself classified as crypto.

Observed vs. intended:

| | Intended (per the description and the inline comment) | Observed |
|---|---|---|
| Trigger | md5/sha1 used for a password, token, or other security-sensitive datum | any occurrence of the token `md5`, `sha1`, or `ecb`, case-insensitive, on any line |
| Input scope | source code | any non-binary buffer under the size bound, Markdown included |
| This input | not a finding — prose describing a file-identity check | 2 findings, `weak-crypto (medium)` |

`ecb` is worth a glance too: `(?i)\bECB\b` matches the standalone word in prose, and "ECB" is also the European Central Bank.

## Why it matters

The guardian is advisory, so nothing is blocked — the cost is signal quality. Evidence-bearing verdicts routinely record hash comparisons to prove a mutant restore was byte-exact, which is a *good* practice this rule punishes. A medium finding that appears every time someone documents an integrity check trains readers to skim past the class, and the next real `md5` password hash rides in behind the noise.

## Suggested directions

Either would fix the reported case; the first is narrower:

1. **Scope the pattern to an API use rather than a word** — require a call or import shape (`md5.New()`, `hashlib.md5(`, `MessageDigest.getInstance("MD5")`, `crypto.createHash('md5')`), the way the ast-grep rule `sec-weak-hash-md5` already does.
2. **Scope the scanner by file type** — skip prose formats (`.md`, `.txt`, `.rst`) for the code-shaped classes, or run only the secret-shaped classes there.

A whitelist of "md5 used for file identity" phrasings is a third option, but it depends on wording and would need maintaining; the two above do not.

## Environment

- Reproduced at tree base `1d150a27d`; `moai-adk v3.2.0-rc.8`; macOS (darwin 25.6.0), `go run` from inside the module.
- Original sighting: card t663 verdict authoring, reported as `weak-crypto (medium) line 75: Weak hash or cipher mode`.

## Related

- Card t542 (this filing); original observation from card t663.
- Sibling filings from the same sweep: #1704, #1705.

🗿 MoAI
