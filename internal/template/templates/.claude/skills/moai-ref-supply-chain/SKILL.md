---
name: moai-ref-supply-chain
description: >
  Software supply-chain defensive security reference: SBOM generation and
  verification (SPDX / CycloneDX), dependency-confusion defense, malicious-package
  triage playbook, SLSA provenance levels, Sigstore / cosign signing and
  verification, package-registry hardening, typosquatting defense, and
  transitive-dependency auditing. Agent-extending skill that amplifies backend,
  security, and release-engineering work with production-grade defensive patterns
  for the software supply chain.
  NOT for: offensive techniques (dependency-confusion attack execution, malicious
  package authoring, registry exploitation), LLM/AI-specific security (see
  moai-ref-llm-security), web-app OWASP Top 10 (see moai-ref-owasp-checklist), or
  general API design (see moai-ref-api-patterns).

when_to_use: >
  Use when generating or verifying an SBOM, defending against dependency confusion
  or typosquatting, triaging a suspicious package before adoption, raising a build
  to a SLSA provenance level, signing or verifying artifacts with Sigstore, hardening
  a package registry, or auditing transitive dependencies for vulnerabilities and
  license risk. Loads as background knowledge for release-engineering, dependency-
  hygiene, and supply-chain-hardening tasks across any language ecosystem.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-06-24"
  tags: "supply-chain, sbom, slsa, sigstore, cosign, dependency-confusion, typosquatting, provenance, transitive-dependencies, reference"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000
---

# Software Supply-Chain Defensive Security Reference

Defensive practitioner reference for hardening a software supply chain — the chain
from source, through build, to the artifact a consumer installs. Every section is
framed as defense, hardening, detection, or verification: it describes the weakness,
how to detect it, and how to prevent it, never how to exploit it. AI/LLM-specific
supply-chain concerns (model and training-data provenance) live in
`moai-ref-llm-security`; web-application vulnerabilities live in
`moai-ref-owasp-checklist`.

## Target Use

Apply when building, releasing, or consuming software components. The threat model
is an untrusted supply chain: any dependency you pull, any build step you run, and
any artifact you ship may have been substituted, tampered with, or impersonated.
The defenses below establish provenance (where did this come from?), integrity
(has it been altered?), and hygiene (is this the component I meant to use?).

## The Supply-Chain Trust Boundaries

The core defensive insight: each hand-off in the chain is a boundary where a
component can be substituted or tampered. Establish provenance and verify integrity
at every hand-off.

| Boundary | Substitution / tamper risk | Primary defense |
|----------|----------------------------|-----------------|
| Source resolution (name → package) | Dependency confusion, typosquatting | Namespace scoping, install-time verification, name allowlist |
| Dependency download | Compromised registry, MITM | Lockfile pinning by hash, signature verification |
| Transitive closure | Vulnerable or malicious deep dependency | Transitive audit, depth limits, SBOM diff |
| Build | Tampered build, injected step | SLSA provenance, isolated/ephemeral builders |
| Artifact publish | Substituted artifact | Sigstore signing, provenance attestation |
| Consumer install | Unverified artifact accepted | Signature + provenance verification at install/admission |

## SBOM — Generation and Verification

A Software Bill of Materials (SBOM) is the inventory of components in an artifact.
It is the foundation for every downstream defense: you cannot audit, scan, or verify
what you have not inventoried.

### Format selection

The two open SBOM formats are interoperable; pick by ecosystem fit and consumer needs.

| Format | Steward | Strength |
|--------|---------|----------|
| SPDX | Linux Foundation (ISO/IEC 5962 standard) | License-centric; broad regulatory and legal acceptance |
| CycloneDX | OWASP | Security-centric; native vulnerability and dependency-relationship modeling |

### Minimum-element baseline

A useful SBOM carries at least the NTIA minimum elements: the supplier, the
component name, the version, unique identifiers, the dependency relationships, the
author of the SBOM data, and a timestamp. An SBOM missing dependency relationships
is an inventory, not a graph — it cannot answer "what pulls in this vulnerable
component?".

### Generation and verification practice

The standard SBOM-generation tool (syft is the de-facto cross-ecosystem standard
that emits both SPDX and CycloneDX) inspects an artifact or source tree and produces
the component inventory. Generate the SBOM **as part of the build**, not after — an
SBOM generated later cannot see build-time-only dependencies. Verify an SBOM by
re-generating it from the artifact and diffing: a drift between the shipped SBOM and
the re-generated one signals tampering or an incomplete original.

- **Generate at build time** so the SBOM reflects the exact resolved closure.
- **Attach the SBOM as a signed attestation** (see Sigstore below) so a consumer can
  trust the inventory, not just read it.
- **Diff SBOMs across releases** to surface a newly-introduced or version-bumped
  dependency before it ships.

## Dependency-Confusion Defense

Dependency confusion (public disclosure, 2021) is when a resolver pulls a public
package that shadows an intended private/internal package of the same name, because
the resolver preferred the public registry. The defense is to make the resolver
prefer — or exclusively use — the trusted source for internal names. This section is
purely defensive: it describes how to prevent the substitution, not how to perform it.

| Control | Defensive rationale |
|---------|---------------------|
| Namespace / scope reservation | Reserve the organization's package namespace on the public registry so an internal name cannot be claimed externally |
| Source pinning per name | Configure the resolver to fetch internal names ONLY from the internal registry; never let an internal name fall through to a public registry |
| Install-time hash verification | Pin every dependency to a content hash in the lockfile so a same-name substitution with different content is rejected |
| Registry routing controls | Use a registry proxy that routes internal-namespace requests to the internal source and refuses public fallback for those names |

Lockfile hash-pinning is the cross-cutting control here: package managers across
ecosystems each record a content hash for every resolved dependency — `pip` writes
hashes in a requirements lockfile, `npm` records integrity hashes in its lockfile,
and `cargo` records checksums in its lockfile — and verify that hash on install, so
a same-name substitution with different bytes fails the check regardless of which
registry served it.

## Malicious-Package Triage Playbook

When a dependency is suspect — flagged by a feed, an advisory, or a reviewer — triage
it before it enters (or to decide whether to evict it). The playbook is detection and
response, never an analysis of how to author a malicious package.

### Triage signals (detection)

| Signal | What it indicates |
|--------|-------------------|
| Name near-match to a popular package | Possible typosquat (see Typosquatting Defense) |
| Recent maintainer change or new publisher | Possible account takeover or hand-off to a bad actor |
| Install/post-install script presence | Code runs at install time — a common abuse vector; inspect what it does |
| Unexpected network or filesystem access in scripts | Exfiltration or persistence behavior |
| Version published out of cadence / sudden major jump | Possible hijack of an abandoned package |
| Mismatch between repository and published artifact | The published artifact was not built from the claimed source |

### Response procedure

1. **Quarantine**: pin the last-known-good version and block the suspect version in
   the resolver / proxy so no build pulls it.
2. **Verify provenance**: check whether the artifact has a signature and provenance
   attestation (Sigstore / SLSA below); an unverifiable artifact stays quarantined.
3. **Inventory exposure**: use the SBOM to find every build that already pulled the
   suspect version (this is what the SBOM dependency graph is for).
4. **Report upstream**: notify the package registry's security/advisory channel and,
   if internal, the internal security owner. Most registries provide a malware/abuse
   reporting channel.
5. **Record**: capture the triage verdict and the evidence so a future occurrence is
   faster to adjudicate.

## SLSA Provenance Levels

SLSA (Supply-chain Levels for Software Artifacts) defines build-integrity levels.
Provenance is signed metadata describing how an artifact was built; the level states
how trustworthy that provenance is. The current Build track defines four levels
(L0–L3); the level rises as the build becomes harder to tamper with. (Earlier SLSA
versions numbered the levels L1–L4 — same idea, four graduated levels.)

| Level | What it guarantees | How to verify |
|-------|--------------------|---------------|
| Build L0 | No guarantee | n/a — treat as unverified |
| Build L1 | Provenance exists: the build process is documented and emits provenance | Check that provenance accompanies the artifact and describes the build |
| Build L2 | Signed provenance from a hosted build platform — tamper-evident via signing | Verify the provenance signature chains to the expected build platform identity |
| Build L3 | Hardened build: provenance is non-forgeable, the build runs in an isolated / ephemeral environment, secret material is isolated from the build | Verify the signature AND that the builder identity is a hardened, isolated platform |

Defensive practice: **require a minimum level for what you consume.** A release
pipeline should refuse to deploy an artifact whose provenance is below the level the
environment demands — production typically requires signed provenance (L2+) at
minimum, and high-assurance environments require a hardened builder (L3).

## Sigstore / cosign Signing and Verification

Sigstore is the de-facto cross-ecosystem standard for signing and verifying software
artifacts. The standard Sigstore signing tool, cosign, signs artifacts and
verifies signatures; it pairs with a certificate authority that issues short-lived
identity-bound certificates and a transparency log that records every signature. This
section frames Sigstore as the verification mechanism a consumer applies, never as a
way to forge a signature.

### Keyless signing model

Keyless signing removes the long-lived private key — the most-stolen secret in a
signing pipeline. Instead:

1. The signer authenticates with an identity provider (OIDC).
2. The CA issues a **short-lived certificate** bound to that identity.
3. The artifact is signed with the ephemeral key.
4. The signature and certificate are recorded in a **public transparency log**.

The result: a verifier can confirm *which identity signed this artifact* without any
party holding a long-lived key that could be stolen.

### Verification at the consumer boundary

The defensive value is on the **verify** side. A consumer (or an admission gate)
verifies an artifact's signature and checks the signer identity against an expected
identity policy before accepting the artifact. The standard verification invocation
shape is `cosign verify --certificate-identity <expected> --certificate-oidc-issuer
<expected> <artifact-ref>`. Verification fails closed: an unsigned or wrong-identity
artifact is rejected.

- **Verify at install / admission time**, not only at publish time — the boundary
  that matters is where the artifact is accepted.
- **Pin the expected signer identity** so a validly-signed-but-wrong-identity
  artifact is rejected.
- **Verify the SBOM and provenance attestations** the same way — they are signed
  artifacts too.

## Package-Registry Hardening

The registry is a high-value target: compromising it lets an attacker substitute
artifacts at the source. Harden both the registry you publish to and any internal
mirror you operate.

| Control | Defensive rationale |
|---------|---------------------|
| Two-factor on all publish/admin accounts | Account takeover is the most common registry-compromise vector; 2FA blocks credential-stuffing |
| Registry-side signing / required attestations | Require artifacts to carry a verifiable signature + provenance before they are publishable or consumable |
| Mirror / proxy pinning | Pin an internal mirror to specific upstream versions by hash so an upstream change cannot silently flow through |
| Scoped publish tokens | Issue per-package, least-privilege publish tokens; never a broad token that can publish any package |
| Immutable published versions | Disallow re-publishing over an existing version so a published artifact cannot be swapped after the fact |
| Audit logging on publish | Log every publish with the identity and artifact hash for tamper detection and incident response |

## Typosquatting Defense

Typosquatting registers a package whose name is a near-miss of a popular one
(transposed letters, added/removed character, lookalike spelling), hoping a developer
mistypes the name. The defense is name verification and allowlisting; this section
does not describe how to register a squat.

| Control | Defensive rationale |
|---------|---------------------|
| Dependency name allowlist | Resolve only from a vetted allowlist of exact package names; a near-miss name is not on the list |
| Automated typo / similarity detection | Run name-similarity checks in CI to flag a new dependency whose name is one edit-distance from a popular package |
| Namespace registration | Register the organization's own likely-mistyped names defensively so they cannot be squatted |
| Pin exact names + hashes | Combine exact-name pinning with hash-pinning so even a name collision serves the wrong bytes and fails the hash check |

A name-allowlist plus hash-pinning is the strongest combination: the allowlist
rejects an unexpected *name*, and the hash check rejects unexpected *bytes* under an
expected name.

## Transitive-Dependency Auditing

Most of a project's dependency surface is transitive — pulled in by direct
dependencies. A vulnerability or malicious package three levels deep is still in your
artifact. Audit the full closure, not just the direct dependencies.

### Audit dimensions

| Dimension | Defensive control |
|-----------|-------------------|
| Vulnerability | Scan the full transitive closure against vulnerability advisories; fail the build on a severity threshold |
| License | Scan transitive licenses against an allowed-license policy; flag a non-compliant deep dependency |
| Depth / count | Bound dependency depth and total count; an unexpectedly deep or wide closure is a review signal |
| Lockfile integrity | Pin the entire closure by hash in the lockfile so the audited closure is the installed closure |

### Ecosystem-neutral auditing

Every language ecosystem ships an advisory-database audit tool that walks the
transitive closure and reports known-vulnerable dependencies — for example `pip-audit`
(Python), `npm audit` (Node.js), `cargo audit` (Rust), `govulncheck` (Go), and
`bundler-audit` (Ruby) each consult their ecosystem's advisory database. Run the
ecosystem's audit tool in CI, fail the build on findings above a chosen severity, and
re-run on every lockfile change so a newly-introduced transitive vulnerability is
caught before merge. For cross-ecosystem projects, a Software Composition Analysis
(SCA) tool that consumes the SBOM gives a single audit across all ecosystems at once.

## Cross-References

- `moai-ref-llm-security` — AI/LLM defensive security, including model and
  training-data provenance (the supply-chain surface specific to ML artifacts that
  this skill's SBOM/SLSA/Sigstore controls underpin).
- `moai-ref-owasp-checklist` — web-application OWASP Top 10, including A06
  (Vulnerable and Outdated Components), which the transitive-dependency audit here
  operationalizes at the supply-chain layer.
- `moai-ref-api-patterns` — REST/GraphQL API design and error handling (the
  application surface, distinct from the supply-chain surface).

## Defensive Severity Levels

| Level | Label | Action | Example |
|-------|-------|--------|---------|
| P0 | CRITICAL | Block release | Unsigned artifact accepted at production deploy; no lockfile hash-pinning on internal-namespace packages |
| P1 | HIGH | Fix before merge | No transitive vulnerability audit in CI; SBOM not generated at build time |
| P2 | MEDIUM | Fix within iteration | No automated typosquat detection on new dependencies; registry publish without 2FA |
| P3 | LOW | Track in backlog | SBOM present but not attached as a signed attestation; license audit not yet enforced |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "We only use well-known packages, so we do not need an SBOM" | You cannot triage a newly-disclosed vulnerability without an inventory. The SBOM is what tells you whether the bad component is in your artifact at all. |
| "Lockfile pinning by version is enough" | Version pinning without hash pinning still accepts substituted bytes under the same version string. Pin by content hash so a same-version substitution fails. |
| "Dependency confusion only affects companies with private packages" | Any project that resolves an internal name from a resolver that can fall through to a public registry is exposed. Scope internal names to the trusted source. |
| "Signature verification slows down our pipeline" | Verification is the boundary between accepting a tampered artifact and rejecting it. An unverified artifact in production is the expensive outcome, not the verification step. |
| "Transitive dependencies are the upstream maintainer's problem" | A vulnerable transitive dependency ships inside your artifact under your name. Audit the full closure; the depth of the problem does not change whose artifact it is in. |
| "SLSA is overkill for an internal build" | The SLSA level you require is the assurance you get against a tampered build. Internal builds are tampered too; require at least signed provenance for anything you deploy. |

**Provenance-by-default**: every artifact you consume is unverified until you check
its signature and provenance, and every name you resolve is unverified until it
matches an expected name and hash. Trust is established by verification, not by
familiarity.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- An artifact is deployed to production without verifying its signature and provenance
- Internal-namespace package names can fall through to a public registry when the internal source misses
- The lockfile pins versions but not content hashes, so a same-version substitution would pass
- No SBOM is generated at build time, so the deployed artifact's component inventory is unknown
- The transitive dependency closure is never audited against vulnerability or license advisories
- A new dependency is added with no name-similarity check against popular package names
- Registry publish/admin accounts lack two-factor authentication
- A published artifact version can be re-published (overwritten) after the fact

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] An SBOM is generated at build time (SPDX or CycloneDX) and carries the minimum elements including dependency relationships
- [ ] Internal-namespace package names resolve only from the trusted internal source; public fallback for those names is disabled
- [ ] Every dependency is pinned by content hash in the lockfile, and the hash is verified on install
- [ ] The transitive dependency closure is audited for vulnerabilities and license compliance in CI, failing above a chosen severity
- [ ] New dependencies pass a name-similarity / typosquat check before adoption
- [ ] Consumed artifacts are verified against a signature and an expected signer identity (Sigstore) at install or admission time
- [ ] A minimum SLSA provenance level is required for deployed artifacts, and the pipeline refuses artifacts below it
- [ ] Registry publish/admin accounts use two-factor authentication and scoped, least-privilege publish tokens

<!-- moai:evolvable-end -->
