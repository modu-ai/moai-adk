# Internal: security-checklist

When to read: when a change handles authentication, authorisation, user input, secrets, outgoing requests or an interface other programs call, as `moai-audit`, or as `moai-impl` writing that change.

Show which items you checked and how; an item you did not check is not_run, never a pass.

## API risks (OWASP API Security Top 10, 2023)

Names from https://owasp.org/API-Security/editions/2023/en/0x11-t10/ (read 2026-11-13T15:00Z).

| ID | Risk | Question to ask |
| --- | --- | --- |
| API1:2023 | Broken Object Level Authorization | Can one user read or change another user's object by changing an identifier? |
| API2:2023 | Broken Authentication | Can authentication be guessed, replayed or bypassed? |
| API3:2023 | Broken Object Property Level Authorization | Can a caller read or set a field it should not see or change? |
| API4:2023 | Unrestricted Resource Consumption | Can repeated or large requests exhaust time, memory, storage or paid quota? |
| API5:2023 | Broken Function Level Authorization | Can an ordinary caller reach an administrative operation? |
| API6:2023 | Unrestricted Access to Sensitive Business Flows | Can a sensitive flow be automated or repeated in bulk? |
| API7:2023 | Server Side Request Forgery | Can input make the server fetch an internal or unintended address? |
| API8:2023 | Security Misconfiguration | Are debug output, verbose errors or default settings exposed? |
| API9:2023 | Improper Inventory Management | Are old, undocumented or test endpoints still reachable? |
| API10:2023 | Unsafe Consumption of APIs | Is data from another service trusted without validation or a time limit? |

## Trust boundaries

- A cached or client-supplied session value is not proof of identity; check against the server-side source before each authorisation decision.
- A check at a proxy, gateway or middleware is a convenience, not the boundary; every operation that changes data checks authentication and ownership itself.
- A scheduled job that calls a URL calls a public URL; require a secret, compared in constant time.
- A webhook payload is untrusted until its signature is verified against the shared secret.
- Production builds ship no source maps, stack traces or build manifests.

## Input, secrets and output

- Validate type, structure and length at the boundary; build queries and commands with parameters, never by joining strings.
- Normalise paths and refuse anything outside the intended root, including through symbolic links.
- Escape output for the context it is written into.
- Keep secrets out of code, history, logs and error messages; write the variable name, never the value.

Settings with numbers (hash costs, token lifetimes, lockout counts, header values) are taken from the current official source for the library or standard in use, with its address and the date read, not from this list.
