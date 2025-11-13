---
name: "moai-artifacts-builder"
version: "4.1.0"
created: 2025-11-13
updated: 2025-11-13
status: stable
tier: Alfred
description: "Enterprise artifact management with governance, lifecycle, SBOM, and supply chain security for November 2025 standards"
keywords: ['artifact', 'governance', 'lifecycle', 'context7-mcp', 'sbom', 'supply-chain-security', 'enterprise-standards']
allowed-tools: "Read, Glob, Bash, WebFetch, WebSearch"
---

# Enterprise Artifact Management & Governance - v4.1.0

## Skill Overview

**November 2025 Enterprise Standards**: Production artifact management with governance, SBOM, and supply chain security

| Feature | Coverage |
|---------|----------|
| **Artifact Types** | 7 standard formats (80% enterprise coverage) |
| **Security** | SBOM, provenance, immutability, SOC 2/ISO 27001 |
| **Context7 MCP** | ✅ Metadata lookup and artifact index search |
| **Compliance** | Automated scanning and signature verification |

## Core Responsibilities

1. **Artifact Classification**: 7 enterprise-standard formats
2. **Lifecycle Management**: Creation → Validation → Storage → Deployment → Retirement
3. **Governance**: RBAC, monitoring, audit trails
4. **Security**: SBOM, provenance, supply chain security
5. **Context7 Integration**: Metadata and vulnerability correlation

---

## Level 1: Quick Reference (50-150 lines)

### Essential Artifact Patterns

**Docker Container Artifact**:
```yaml
artifact:
  id: "api-gateway@2.5.1"
  type: "Container Image"
  format: "Docker OCI"
  registry: "docker.io"
  repository: "myorg/api-gateway"
  tag: "2.5.1"
  
  metadata:
    created: 2025-11-13T14:30:00Z
    source_repo: "https://github.com/myorg/api-gateway"
    source_commit: "abc123def456789"
    creator: "github-actions"
    
  security:
    vulnerability_scan:
      tool: "Trivy v0.54.0"
      status: "passed"
      critical_vulnerabilities: 0
      high_vulnerabilities: 0
```

**Python Package Artifact**:
```yaml
artifact:
  id: "data-pipeline@1.3.2"
  type: "Language Package"
  format: "Python Wheel"
  
  filename: "data_pipeline-1.3.2-py3-none-any.whl"
  pypi_url: "https://pypi.org/project/data-pipeline/1.3.2/"
  
  metadata:
    created: 2025-11-13T13:45:00Z
    python_version: "3.8+"
    checksum:
      sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    
    dependencies:
      - name: "pandas"
        version: "2.1.0"
        requirement: "pandas>=2.1.0,<3.0"
```

**SBOM Requirement** (CycloneDX 1.6):
```yaml
sbom:
  version: "CycloneDX 1.6"
  spec_version: "1.6"
  
  metadata:
    timestamp: 2025-11-13T14:30:00Z
    tools:
      - name: "syft"
        version: "0.95.0"
  
  components:
    - type: "library"
      name: "requests"
      version: "2.31.0"
      purl: "pkg:pypi/requests@2.31.0"
      licenses: [{ name: "Apache-2.0" }]
```

**7 Enterprise Artifact Types**:
1. **Container Images** (Docker/OCI) - 80-90% enterprise usage
2. **Language Packages** (Maven, npm, PyPI) - Package managers, semantic versioning
3. **Binary Artifacts** (Executable, .so, .dll) - Native code, compiled binaries
4. **Documentation** (API docs, guides) - HTML, PDF, Markdown formats
5. **Configuration** (IaC, templates) - Terraform, CloudFormation, Helm
6. **Test Reports** (Coverage, scans) - JSON, XML, CSV compliance data
7. **Source Archives** (Release tarballs) - .tar.gz, .zip with GPG signatures

**Core Governance Rules**:
- ✅ **Classification First**: Every artifact maps to one of 7 types
- ✅ **Provenance Tracking**: Complete source-to-production traceability
- ✅ **Immutability**: Published artifacts cannot be modified
- ✅ **SBOM Required**: All artifacts include Software Bill of Materials
- ✅ **Security Scanning**: Pre-deployment vulnerability verification

---

## Level 2: Practical Implementation (200-300 lines)

### Enterprise Artifact Lifecycle

**Phase 1: Creation**:
```yaml
phase: "creation"
activities:
  - Build artifact from source (CI/CD)
  - Generate metadata (timestamp, commit SHA, builder)
  - Create initial SBOM (syft, cyclonedx)
  - Assign semantic version (SemVer)
  
artifacts:
  - Raw binary/container
  - SBOM (CycloneDX or SPDX)
  - Build log
  - Metadata JSON
```

**Phase 2: Validation**:
```yaml
phase: "validation"
gates:
  - ✅ Vulnerability scan (Trivy, Grype)
  - ✅ License compliance (FOSSA, Black Duck)
  - ✅ SBOM validation (cyclonedx-python)
  - ✅ Signature verification (GPG/RSA)
  - ✅ Artifact integrity (checksum match)
  
failure_behavior:
  critical: "block_deployment"  # CRITICAL/HIGH vulnerabilities
  high: "require_approval"      # Manual approval required
  medium_low: "log_warning"     # Warning only
```

**Multi-Format Repository Setup**:
```yaml
repositories:
  - name: "container-registry"
    type: "Container (Docker/OCI)"
    upstream: "docker.io"
    proxy_cache:
      enabled: true
      retention_days: 30
    security:
      scan_on_push: true
      signature_required: true
      sbom_required: true
  
  - name: "python-packages"
    type: "Python (PyPI)"
    upstream: "pypi.org"
    retention_policy:
      strategy: "semantic_versioning"
      keep_release_versions: true
      keep_prerelease: 3
  
  - name: "binary-artifacts"
    type: "Generic Binary"
    retention_policy:
      max_size_gb: 500
      cleanup_strategy: "lru"

governance:
  rbac:
    admin_group: "release-engineering"
    publish_group: "ci-automation"
    read_group: "developers"
  
  approval_workflow:
    require_approval: true
    approvers: ["security-team", "release-manager"]
    timeout_hours: 24
```

**Context7 MCP Integration**:
```python
# Context7 artifact metadata lookup
context7_query = {
    "operation": "artifact_metadata_lookup",
    "artifact_id": "app-service@1.0.0",
    "fields": [
        "sbom",
        "provenance",
        "signatures",
        "vulnerability_scan",
        "compliance_status"
    ]
}

# Response structure
response = {
    "artifact": {
        "id": "app-service@1.0.0",
        "registry": "docker.io",
        "location": "docker.io/myorg/app-service:1.0.0",
        "sbom_url": "context7://sbom-index/app-service@1.0.0",
        "provenance_verified": True,
        "vulnerabilities_critical": 0
    }
}
```

**SBOM Index Search**:
```python
# Context7 vulnerability correlation
context7_query = {
    "operation": "vulnerability_correlation",
    "cve_id": "CVE-2024-5678",
    "affected_components": ["requests", "urllib3"],
    "action": "find_artifacts"
}

# Response
response = {
    "cve": "CVE-2024-5678",
    "severity": "HIGH",
    "affected_artifacts": [
        {
            "artifact_id": "api-gateway@2.5.1",
            "component": "requests@2.31.0",
            "status": "vulnerable",
            "patch_available": True,
            "patched_version": "requests@2.32.0",
            "recommended_action": "upgrade_component_rebuild"
        }
    ]
}
```

**Binary Artifact Example**:
```yaml
artifact:
  id: "performance-optimizer@3.1.0"
  type: "Binary Artifact"
  format: "Native Executable (.so)"
  
  files:
    - name: "libperformance_optimizer.so"
      arch: "x86_64"
      os: "linux"
      size_bytes: 512000
      checksum_sha256: "abc123..."
      signature: "RSA-4096"
    
    - name: "libperformance_optimizer.dylib"
      arch: "arm64"
      os: "macos"
      size_bytes: 480000
      checksum_sha256: "def456..."
      signature: "RSA-4096"
  
  metadata:
    created: 2025-11-13T12:00:00Z
    compiler: "gcc-13"
    optimization_flags: "-O3 -march=native"
    source_commit: "release/3.1.0"
  
  distribution:
    storage: "artifactory.myorg.com"
    bucket: "native-binaries"
    access: "restricted"
    approval_required: true
```

**Terraform/IaC Artifact**:
```yaml
artifact:
  id: "aws-infrastructure@4.2.0"
  type: "Configuration/IaC"
  format: "Terraform Module"
  
  files:
    - path: "main.tf"
      size_bytes: 2048
      checksum_sha256: "ijk789..."
    - path: "variables.tf"
      size_bytes: 1024
      checksum_sha256: "lmn012..."
  
  metadata:
    created: 2025-11-13T10:30:00Z
    terraform_version: ">= 1.5"
    cloud_provider: "AWS"
    modules_included: 5
    
  validation:
    terraform_fmt: "passed"
    terraform_validate: "passed"
    security_scan: "passed"
```

---

## Level 3: Advanced Integration (50-150 lines)

### Advanced Governance & Security

**Supply Chain Security Framework**:
```yaml
artifact_governance:
  supply_chain_security:
    provenance_tracking: true
    source_commit_required: true
    builder_attestation: true
    
    sbom_requirements:
      format: ["CycloneDX-1.6", "SPDX-2.3"]
      components_scanned: true
      license_compliance: true
      vulnerability_threshold:
        critical: "block"
        high: "require_approval"
        medium: "log_only"
    
    signature_verification:
      algorithms: ["RSA-4096", "ECDSA-P256"]
      trusted_signers: ["release-key@myorg.com"]
      timestamp_authority: "https://timestamp.comodoca.com"
  
  context7_mcp:
    enabled: true
    operations:
      - artifact_metadata_lookup
      - sbom_index_search
      - vulnerability_correlation
      - compliance_status_check
```

**Attestation Framework** (SLSA):
```yaml
artifact:
  id: "attestation@app-service-1.0.0"
  type: "Attestation"
  format: "in-toto/SLSA"
  
  attestation:
    version: "0.3"
    
    statement:
      _type: "https://in-toto.io/Statement/v0.1"
      
      subject:
        - name: "docker.io/myorg/app-service"
          digest:
            sha256: "abc123def456..."
      
      predicateType: "https://slsa.dev/provenance/v1"
      
      predicate:
        buildDefinition:
          buildType: "https://github.com/slsa-framework/slsa-github-generator@v1"
          externalParameters:
            source: "https://github.com/myorg/app-service"
            ref: "refs/tags/1.0.0"
        
        runDetails:
          builder:
            id: "https://github.com/slsa-framework/slsa-github-generator"
          completion:
            finishTime: "2025-11-13T14:35:00Z"
```

**Advanced Automation**:
```yaml
automation:
  auto_scan_on_push: true
  auto_sbom_generation: true
  auto_compliance_report: true
  auto_deprecation_warnings: true
  
  monitoring:
    sla_targets:
      - artifact_availability: "99.99%"
      - scan_completion: "< 5 minutes"
      - deployment_success: "> 95%"
    
    alerts:
      - high_vulnerability_detected: "notify_security_team"
      - signature_verification_failed: "block_deployment"
      - unauthorized_access: "incident_response"
```

**Enterprise Compliance**:
```yaml
compliance:
  sbom_required: true
  sbom_format: "CycloneDX"
  signature_required: true
  audit_retention_years: 7
  
  frameworks:
    - SOC 2 Type II
    - ISO 27001
    - NIST CSF
    - GDPR (for EU deployments)
    
  reporting:
    automated_reports: true
    vulnerability_reports: "daily"
    compliance_reports: "weekly"
    audit_trail: "7 years"
```

**Release Bundle Example**:
```yaml
artifact:
  id: "release@v2.0.0"
  type: "Release Bundle"
  format: "GitHub Release"
  
  release:
    tag: "v2.0.0"
    name: "Version 2.0.0 - Production Release"
    created: 2025-11-13T14:40:00Z
    
    assets:
      - name: "app-service-2.0.0.tar.gz"
        size_bytes: 5242880
        checksum_sha256: "abc123def456..."
      
      - name: "SBOM.json"
        size_bytes: 123456
      
      - name: "SBOM.json.sig"
        size_bytes: 256
    
    changelog: |
      ## What's New
      - Feature A
      - Feature B
      - Bug fixes
```

## Best Practices Checklist

### Artifact Creation
- [ ] **Classification**: Choose from 7 standard types
- [ ] **Metadata**: Include creator, timestamp, source commit
- [ ] **Provenance**: Source repo, commit SHA, build log links
- [ ] **SBOM**: CycloneDX or SPDX format
- [ ] **Signature**: RSA-4096 or ECDSA-P256
- [ ] **Scanning**: Trivy/Grype vulnerability detection
- [ ] **Immutability**: No post-publication modifications

### Repository Design
- [ ] **Multi-format Support**: Container, Python, Binary, IaC, Docs
- [ ] **Registry Configuration**: Official (upstream), proxy cache (local)
- [ ] **RBAC**: Admin, publisher, read permissions
- [ ] **Approval Workflow**: Security team and release manager approval
- [ ] **Auto-scanning**: Push-time vulnerability scanning
- [ ] **SBOM Required**: All artifacts must include SBOM
- [ ] **Audit Trail**: 7-year retention for compliance

---

**Version**: 4.1.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready  
**Standards**: November 2025 Enterprise Standards  
**Compliance**: SOC 2, ISO 27001, NIST CSF Ready
