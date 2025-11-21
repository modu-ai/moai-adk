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


**Version**: 4.1.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready  
**Standards**: November 2025 Enterprise Standards  
**Compliance**: SOC 2, ISO 27001, NIST CSF Ready


## Advanced Patterns

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




## Context7 Integration

### Related Libraries & Tools
- [Vite](/vitejs/vite): Next generation frontend tooling
- [Webpack](/webpack/webpack): Module bundler

### Official Documentation
- [Documentation](https://vitejs.dev/guide/)
- [API Reference](https://vitejs.dev/config/)

### Version-Specific Guides
Latest stable version: 5.x
- [Release Notes](https://github.com/vitejs/vite/releases)
- [Migration Guide](https://vitejs.dev/guide/migration.html)
