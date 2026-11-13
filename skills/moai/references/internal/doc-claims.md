# Internal: doc-claims

When to read: when you write or check a document for users (a README, release notes, an install or usage guide, a compatibility table), as `moai-docs`.

Every checkable statement in such a document needs evidence you can name. Judge the claims; fix the document only as a separate, stated change.

1. **Scope.** Only documents written for users of the software. Design notes and Issue records are out of scope.
2. **List the evidence first.** For each item: what it is, where it came from, its version or commit, and when it was produced. An undated item cannot support a claim about the current state. Never copy a secret into the report; name its location for redaction.
3. **Split the claims.** Pull out every statement about platforms, versions, install and usage steps, defaults, limits, counts and availability. Split a sentence until each claim asserts one thing ("installs with one command on macOS and Linux" is three claims). Leave out statements of taste ("fast", "easy") and say why.
4. **Label each claim once**, taking the first label that applies, in this order:

| Label | When |
| --- | --- |
| `needs-human` | Settling it needs a command run, a private system, a UI or a maintainer's call. Name the exact command or file. |
| `stale-suspected` | It was true once, but current evidence differs on a version, date, count or name. Name the field and both values. |
| `verified` | Named evidence supports it directly. |
| `unsupported` | None of the above. Give one reason: `missing-evidence` (nothing speaks to it), `contradicted` (evidence says the opposite) or `insufficient-coverage` (evidence covers less than the claim, for example one platform of three). |

5. **Report** three parts: what was read (documents and evidence with versions and dates, excluded statements, evidence that was missing); one row per claim with its label and anchor; and what the check does not prove.

Background knowledge is not evidence, and the absence of a contradiction is not support.
