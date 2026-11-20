# DevSecOps — Pipeline and Infrastructure Defensive Hardening

Defensive reference for the CI/CD pipeline and the infrastructure-as-code it applies.
The pipeline is a high-value target: it holds credentials, runs code, and produces the
artifact that ships to production. Every section frames the weakness, how to detect it,
and how to prevent it — never how to subvert a pipeline.

Back to the overview: [SKILL.md](../SKILL.md).

## The Pipeline Threat Model

A CI/CD pipeline turns source into a deployed artifact. Each stage is a boundary where
a compromise can inject a build step, exfiltrate a secret, or alter the output. The
defense is to treat the pipeline as a production system: least privilege, isolation,
provenance, and secret hygiene at every stage.

| Stage | Risk | Primary defense |
|-------|------|-----------------|
| Source checkout | Poisoned dependency / build script | Pinned dependencies, reviewed pipeline definition |
| Build / test | Untrusted code on a shared runner | Ephemeral, isolated runners; no privileged mode |
| Secret access | Credential exposure in logs or environment | Scoped, short-lived secrets; log redaction |
| Artifact produce | Tampered or unsigned output | Signed provenance, reproducible build |
| Deploy | Misconfigured target resource | IaC scan before apply; least-privilege deploy identity |

## CI/CD Pipeline Hardening

The pipeline definition is code; harden it as code. The controls below reduce what a
compromised pipeline can reach and make tampering detectable.

| Control | Defensive rationale |
|---------|---------------------|
| Pin the pipeline definition and its actions/plugins by version or digest | An unpinned third-party pipeline action can change under you; pinning by digest freezes the reviewed code |
| Least-privilege pipeline tokens | A pipeline token should grant the minimum scope for its job — never a broad org-admin or deploy-anywhere token |
| Ephemeral, isolated runners | Run each job on a fresh, isolated runner so one job cannot read another job's secrets or persist a foothold |
| Disable privileged mode on runners | A privileged runner can reach the host; default to unprivileged and grant capabilities narrowly |
| Protected branches gate deploys | Require review + passing checks before a deploy stage runs; do not let an arbitrary branch trigger production deploy |
| Sign produced artifacts | Emit signed provenance so a consumer can verify the artifact came from this pipeline (see `moai-ref-supply-chain` for the signing mechanism) |
| Audit-log every pipeline run | Record who triggered each run, what it deployed, and the artifact digest, for tamper detection and incident response |

### Pipeline secret hygiene

Secrets are the most-targeted pipeline asset. The defense is to scope them, keep them
out of logs, and rotate them.

- **Scope secrets to the job that needs them** — a deploy secret is not available to a
  test job; a per-environment secret is not available across environments.
- **Use short-lived, federated credentials** where the platform supports identity
  federation (the pipeline exchanges its workload identity for a short-lived token),
  removing long-lived secrets from the pipeline entirely.
- **Redact secrets from logs** — the platform's secret-masking must be on, and build
  scripts must never `echo` a secret value or write it to an artifact.
- **Never persist a secret to the artifact or the image** — a secret baked into a
  layer is recoverable by anyone who pulls the image (see Image scanning in the
  Container module).

## Secret Scanning

A leaked secret in source or history is a standing breach. Secret scanning detects it
before it ships and detects history that already contains one.

| Control | Defensive rationale |
|---------|---------------------|
| Pre-commit secret scan | The standard secret-scanning tool runs as a pre-commit hook so a secret never reaches the remote |
| CI secret scan on every push | A second scan in CI catches what a bypassed local hook missed; fail the build on a finding |
| Full-history scan | Scan the entire git history, not just the diff — a secret committed and "removed" in a later commit is still in history and must be rotated, not just deleted |
| Rotate-on-detection | A detected secret is a compromised secret; the response is rotation, not deletion. Deletion hides the evidence; the secret is already exposed |
| Allowlist false positives narrowly | Suppress a confirmed false positive with a scoped, reviewed allowlist entry — never a global disable that blinds the scanner |

The standard secret scanner ships as both a pre-commit hook and a CI step; run it in
both places. A finding is treated as a live incident: rotate the credential, then purge
or accept the history exposure.

## IaC Misconfiguration Detection

Infrastructure-as-code (Terraform, CloudFormation, Kubernetes manifests, and the like)
defines cloud resources declaratively. A misconfiguration in IaC — an open storage
bucket, a permissive security group, an over-broad IAM policy — ships to production as
surely as a code bug. The defense is to scan the IaC against a misconfiguration policy
**before apply**.

### Misconfiguration classes (detect before apply)

| Class | Example misconfiguration | What the scan checks |
|-------|--------------------------|----------------------|
| Public exposure | Storage bucket or database open to the public internet | No resource is world-readable/writable unless explicitly intended and reviewed |
| Over-broad identity | IAM policy granting `*` actions on `*` resources | Policies are least-privilege; wildcards are flagged |
| Network openness | Security group allowing `0.0.0.0/0` on a sensitive port | Ingress is scoped to known sources; no broad inbound on admin ports |
| Missing encryption | Volume, bucket, or database without encryption-at-rest | Encryption is enabled on data stores |
| Disabled logging | Resource with audit/flow logging turned off | Logging is enabled where the policy requires it |
| Drift from baseline | Applied state diverges from the reviewed IaC | The deployed state matches the version-controlled definition |

### IaC scanning practice

The standard IaC scanner parses the IaC files and evaluates them against a policy set
(commonly CIS Benchmark-derived rules). Run it as a CI gate **before** the apply stage,
fail the pipeline on a finding above a chosen severity, and treat the policy set as code
that is reviewed and versioned. A scan that runs after apply is detection, not
prevention; the value is catching the misconfiguration before the resource exists.

- **Scan before apply** — the gate belongs between the plan and the apply stage.
- **Use a policy-as-code engine** — express the organization's guardrails as policy so
  the same rules apply across every IaC change, and the policy itself is reviewable.
- **Fail closed on high-severity findings** — a public-bucket or wildcard-IAM finding
  blocks the apply; a low-severity finding may warn.
- **Detect drift** — periodically compare the live infrastructure against the IaC so an
  out-of-band change (made directly in the console) is surfaced and reconciled.

## SAST / DAST Integration

Static and dynamic application security testing find code-level and runtime-level flaws
in the pipeline, before they reach production.

| Tier | What it analyzes | When it runs | What it catches |
|------|------------------|--------------|-----------------|
| SAST (static) | Source code and its data flows, without running it | On every push / PR | Injection sinks, unsafe deserialization, hardcoded secrets, dangerous API use |
| DAST (dynamic) | A running instance, by sending crafted requests | Against a staging deploy | Runtime-only flaws: auth gaps, server misconfiguration, reflected injection |
| IAST / runtime (optional) | Instruments a running app to combine static + dynamic signal | During integration tests | Flaws that need both code context and runtime to surface |

### Integration practice

- **Run SAST early** — the standard SAST tool runs on every pull request so a flaw is
  caught at review time, when it is cheapest to fix. Fail the PR on a high-severity
  finding; surface lower-severity findings as review comments.
- **Run DAST against staging** — point the dynamic scanner at a deployed staging
  instance, not production, and not the source. DAST exercises the running surface.
- **Tune for signal** — a noisy SAST configuration trains reviewers to ignore it.
  Suppress confirmed false positives narrowly and keep the high-severity gate strict.
- **Separate finding from filtering** — collect every finding with a severity and
  confidence, then rank; do not silently drop low-severity findings, because a
  clustered set of low findings can indicate a systemic gap.

## MITRE ATT&CK — Defensive Correlation (pipeline)

The skill cites MITRE ATT&CK technique IDs to correlate a defense with the technique it
counters — never to provide an attack procedure. Representative correlations for the
CI/CD surface:

| ATT&CK technique | Technique name | Defensive control that counters it |
|------------------|----------------|------------------------------------|
| T1195 | Supply Chain Compromise | Pinned dependencies, signed provenance, SBOM (see `moai-ref-supply-chain`) |
| T1552 | Unsecured Credentials | Secret scanning, scoped short-lived tokens, log redaction |
| T1078 | Valid Accounts | Least-privilege pipeline tokens, federated short-lived credentials, audit logging |

Each technique is cited to state how to detect and defend, never how to execute. Citing
a technique ID alongside an exploitation procedure is the anti-pattern this skill forbids.

## DevSecOps Verification Checklist

- [ ] The pipeline definition and its third-party actions/plugins are pinned by version or digest
- [ ] Pipeline tokens are least-privilege and scoped per job; deploy credentials are short-lived or federated
- [ ] Runners are ephemeral and isolated; privileged mode is off by default
- [ ] Secret scanning runs as both a pre-commit hook and a CI step, over the full history; detection triggers rotation
- [ ] IaC is scanned for misconfiguration before apply, failing closed on high-severity findings, with drift detection
- [ ] SAST runs on every PR and DAST runs against staging; high-severity findings gate the pipeline
- [ ] Every pipeline run is audit-logged with trigger identity, deployed target, and artifact digest
