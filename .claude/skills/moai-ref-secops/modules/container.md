# Container — Image, Kubernetes RBAC, and Runtime Defensive Hardening

Defensive reference for containers and their orchestrator. Every section frames a
misconfiguration, how to detect it, and how to prevent it. Where a topic inherently
names an attack vector — container escape exists, RBAC misconfiguration enables
privilege movement — the section describes the misconfiguration and the hardening that
closes it, and omits any step-by-step exploitation. There are no attack procedures here.

Back to the overview: [SKILL.md](../SKILL.md).

## Container Trust Boundaries

A container shares the host kernel. The isolation between a container and its host is
configuration, not a hard boundary — a misconfigured container weakens it. The
orchestrator adds an identity-and-authorization layer (RBAC) on top. Defense hardens
both: the container's own posture and the cluster's authorization.

| Boundary | Risk | Primary defense |
|----------|------|-----------------|
| Image content | Vulnerable dependency, embedded secret, untrusted base | Image scanning, minimal base, no embedded secrets |
| Container → host kernel | Escape to the node | seccomp, AppArmor/SELinux, non-root, read-only root, dropped capabilities, no privileged mode |
| Pod identity → cluster | Over-privileged ServiceAccount used to reach other resources | Least-privilege RBAC, ServiceAccount scoping |
| Workload admission | A non-compliant pod is allowed to run | PodSecurity admission, policy engine gate |
| Running workload | Post-deploy compromise with no signal | Runtime threat detection + alerting |

## Image Scanning

The image is the unit of deployment. A vulnerable dependency or an embedded secret in
any layer ships with it. Scan the image and gate admission on the result.

| Control | Defensive rationale |
|---------|---------------------|
| Scan every image in the pipeline | The standard container image scanner inspects each layer against vulnerability advisories before the image is pushed; fail the build above a severity threshold |
| Minimal / distroless base image | A smaller base has fewer packages and therefore fewer vulnerabilities and a smaller attack surface |
| No secrets in any layer | A secret in an early layer is recoverable even if a later layer deletes it; keep secrets out of the build entirely (inject at runtime) |
| Pin the base image by digest | Pinning by digest freezes the scanned content; a moving tag can change under you |
| Sign and verify the image | Sign the image and verify the signature at admission (see `moai-ref-supply-chain` for the signing mechanism) so only built-and-scanned images run |
| Re-scan on advisory update | A clean image becomes vulnerable when a new advisory drops; re-scan running images, not just build-time ones |

### Admission control on scan result

Admission control is the cluster-side gate: the standard admission-control policy engine
evaluates each pod against policy before it is scheduled. Use it to require that an image
was scanned and signed, that it is from an allowed registry, and that the pod's security
context meets the hardening baseline below. An image that fails policy is rejected at
admission, not caught after it runs.

## Kubernetes RBAC Hardening

RBAC governs which identity can perform which action on which resource. A misconfigured
RBAC — an over-broad RoleBinding, a ServiceAccount bound to cluster-admin — lets a single
compromised workload reach the whole cluster. The defense is least privilege. This
section describes how to scope RBAC tightly; it does not describe how to abuse a loose one.

| Control | Defensive rationale |
|---------|---------------------|
| Least-privilege Roles | Grant only the verbs and resources a workload needs; avoid wildcard verbs (`*`) and wildcard resources |
| No ServiceAccount bound to cluster-admin | A workload ServiceAccount must never hold cluster-admin; that binding turns any pod compromise into a cluster compromise |
| Namespace-scoped Roles over ClusterRoles | Prefer a namespaced Role + RoleBinding; reserve ClusterRoles for genuinely cluster-wide needs and review every ClusterRoleBinding |
| One ServiceAccount per workload | Give each workload its own ServiceAccount with its own minimal permissions, so a compromise is contained to that workload's scope |
| Disable automounting where unused | A pod that does not call the Kubernetes API should not automount its ServiceAccount token, removing a credential an attacker could otherwise reach |
| Review `escalate` / `bind` / `impersonate` verbs | These verbs let a holder grant themselves more access; restrict them to a small, audited set of administrative roles |
| Audit RBAC regularly | Periodically list who-can-do-what (the standard `kubectl auth can-i` review, or an RBAC-audit tool) to catch privilege creep |

### ServiceAccount token hygiene

A ServiceAccount token is a cluster credential. Treat it as one: scope it, do not
automount it where unused, prefer short-lived projected tokens over long-lived ones, and
never copy it out of the cluster. A token that leaks is a foothold; minimizing where
tokens exist minimizes the blast radius.

## Container-Escape Defense

Container escape is movement from inside a container to the host node. It is enabled by a
weak container security context — a privileged container, a root user, a writable root
filesystem, dangerous mounted host paths, or excess Linux capabilities. The defense is to
harden the security context so the escape paths are closed. This section names the
misconfigurations that enable escape and the hardening that prevents each; it provides no
escape procedure.

| Hardening control | Misconfiguration it closes | Defensive rationale |
|-------------------|----------------------------|---------------------|
| Run as non-root (`runAsNonRoot: true`) | Container runs as root (UID 0) | A non-root process has far less leverage against the host if the runtime is breached |
| Read-only root filesystem | Writable root filesystem | A read-only root prevents an attacker from writing a payload or modifying binaries inside the container |
| Drop all capabilities, add back only what is needed | Container holds the default (or `ALL`) Linux capabilities | Most workloads need no capabilities; dropping them removes kernel-level powers an escape would rely on |
| seccomp profile (`RuntimeDefault` or stricter) | seccomp disabled (`Unconfined`) | A seccomp profile blocks dangerous syscalls, narrowing the kernel surface a breakout could use |
| AppArmor / SELinux profile | No mandatory access control on the container | An MAC profile confines the container's file and capability access at the kernel level |
| No privileged mode (`privileged: false`) | `privileged: true` | A privileged container has near-host access; it is the single most dangerous misconfiguration and must be off |
| No dangerous host mounts | Host `/` , the container runtime socket, or `/proc` mounted in | Mounting the host filesystem or the runtime socket hands the container a direct path to the host |
| Disable privilege escalation (`allowPrivilegeEscalation: false`) | A process can gain more privileges than its parent | Blocking escalation prevents a setuid-style step-up inside the container |

### The hardened-pod baseline

A hardened pod security context, applied as the default and enforced by PodSecurity
admission (the `restricted` profile) or an admission policy: non-root, read-only root,
all capabilities dropped, `RuntimeDefault` seccomp, no privileged mode, no privilege
escalation, no host namespaces, no host-path mounts. Enforce this baseline at admission so
a pod that violates it is rejected before it runs, rather than detected after.

## Runtime Threat Detection

Static scanning and admission control are pre-deploy. Runtime detection covers what
happens after a workload is running — a post-deploy compromise, a zero-day, or anomalous
behavior. The standard runtime-detection engine watches kernel-level activity (syscalls,
process spawns, file and network events) and alerts on behavior that matches a detection
rule. This is detection-and-alerting, framed defensively; the rules describe what
malicious behavior looks like so it can be detected, not how to perform it.

### Detection-rule shape (concept, not vendor)

A runtime-detection rule names a behavior and the condition that triggers an alert. The
shape transfers to any runtime-detection tool:

- **Container drift**: a process runs that was not in the original image (an unexpected
  binary executing inside a container is a strong compromise signal).
- **Shell in a production container**: an interactive shell spawned in a container that
  should never have one.
- **Sensitive file access**: a read of a credential file, the ServiceAccount token path,
  or `/etc/shadow` by an unexpected process.
- **Credential-access behavior**: process behavior matching credential-dumping patterns —
  the rule detects the behavior so a responder is alerted; the rule is the defense, not an
  instruction to perform the dumping.
- **Outbound to an unexpected destination**: a container connecting to a host or port
  outside its expected egress, a possible command-and-control or exfiltration signal that
  the rule flags for investigation.
- **Privilege-change / escape attempt**: a syscall pattern consistent with an escape
  attempt — the rule alerts so the workload can be isolated and investigated.

### Runtime-detection practice

- **Alert, then route** — a detection with no responder is noise; route alerts to an
  on-call channel and an incident process.
- **Tune to the workload** — baseline each workload's normal behavior so the rules fire on
  genuine anomalies, not normal operation; an untuned ruleset trains responders to ignore it.
- **Pair with response** — on a high-confidence detection, the response is to isolate the
  pod (cordon the node or kill the workload), preserve evidence, and investigate.
- **Correlate to MITRE ATT&CK** — map detection rules to ATT&CK technique IDs for
  defensive correlation (see below), so coverage gaps are visible.

## MITRE ATT&CK — Defensive Correlation (containers)

The skill cites MITRE ATT&CK Containers-matrix technique IDs to correlate a defense with
the technique it counters — never to provide an attack procedure. Representative
correlations:

| ATT&CK technique | Technique name | Defensive control that counters it |
|------------------|----------------|------------------------------------|
| T1610 | Deploy Container | Admission control, signed-image requirement, registry allowlist |
| T1611 | Escape to Host | Non-root, read-only root, dropped capabilities, seccomp, no privileged mode, no host mounts |
| T1613 | Container and Resource Discovery | Least-privilege RBAC, ServiceAccount token hygiene, network policy |
| T1609 | Container Administration Command | RBAC restriction on exec, runtime detection of shell-in-container |
| T1525 | Implant Internal Image | Image signing + verification at admission, registry hardening (see `moai-ref-supply-chain`) |

Each technique is cited to state how to detect and defend, never how to execute. A
technique ID accompanied by an exploitation procedure for that technique is the anti-pattern
this skill forbids.

## Container Verification Checklist

- [ ] Every image is scanned for vulnerabilities in the pipeline and re-scanned on new advisories; builds fail above a severity threshold
- [ ] Images use a minimal base, carry no secrets in any layer, and are pinned by digest
- [ ] Admission control requires scanned, signed images from an allowed registry and enforces the hardened-pod baseline
- [ ] Pods run non-root with a read-only root filesystem, all capabilities dropped, a seccomp profile, no privileged mode, and no privilege escalation
- [ ] No host paths, the runtime socket, or host namespaces are mounted into workload containers
- [ ] RBAC is least-privilege; no workload ServiceAccount is bound to cluster-admin; ServiceAccount tokens are scoped and not automounted where unused
- [ ] PodSecurity admission (restricted) or an equivalent policy engine enforces the baseline at admission
- [ ] A runtime-detection engine alerts on drift, escape attempts, and credential access, with alerts routed to a responder
