---
name: "moai-artifacts-builder"
version: "4.1.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: Alfred
description: "아티팩트 생성, 관리, 거버넌스 패턴을 정의합니다. November 2025 enterprise standards 기반. 아티팩트 분류, Context7 MCP 통합, 생명주기 관리, 보안 검증. 사용: 프로덕션 아티팩트 생성, 아티팩트 거버넌스 설계, 저장소 구조 정의, SBOM 생성, 공급망 보안 검증 필요 시."
keywords: ['artifact', 'governance', 'lifecycle', 'context7-mcp', 'sbom', 'supply-chain-security', 'enterprise-standards', 'november-2025', 'devops', 'compliance']
allowed-tools: "Read, Glob, Grep, Bash, WebFetch, WebSearch"
---

## Skill 개요

**moai-artifacts-builder**는 November 2025 enterprise 아티팩트 관리 표준을 기반으로 합니다.

| 항목 | 값 |
|------|-----|
| 버전 | 4.1.0 (November 2025 stable) |
| 티어 | Alfred (상위 계층) |
| 자동 로드 | 아티팩트 생성, 저장소 설계, 거버넌스 정책 필요 시 |
| 관련 Skills | moai-alfred-rules, moai-security-governance |
| Context7 MCP | ✅ 통합 (artifact index 조회) |

---

## 무엇을 하는가?

### 핵심 책임

1. **아티팩트 분류 (Classification)**: 7가지 표준 형식 (80% enterprise coverage)
2. **생명주기 관리 (Lifecycle)**: 생성 → 검증 → 저장 → 배포 → 폐기
3. **거버넌스 정책 (Governance)**: RBAC, 감시, 감사 추적
4. **보안 & 준수 (Security)**: SBOM, provenance, immutability, SOC 2/ISO 27001
5. **Context7 MCP 통합**: 메타데이터 조회, 아티팩트 인덱스 검색
6. **Enterprise 패턴**: 멀티 저장소, 프록시 레지스트리, 캐싱 전략

---

## 언제 사용하는가?

### 필수 시나리오 (MUST use)

| 상황 | 사용 여부 |
|------|---------|
| ✅ 프로덕션 아티팩트 저장소 설계 | **필수** |
| ✅ Docker 이미지, Python wheel 배포 | **필수** |
| ✅ 공급망 보안 (supply chain security) 검증 | **필수** |
| ✅ SBOM(Software Bill of Materials) 생성 | **필수** |
| ✅ 규정 준수 감시(SOC 2, ISO 27001) | **필수** |
| ✅ 아티팩트 거버넌스 프레임워크 설계 | **필수** |
| ✅ 멀티 형식 저장소 구성 | **필수** |
| ✅ Artifact provenance & immutability 검증 | **필수** |

### 선택 시나리오 (MAY use)

- 로컬 캐시 관리 (선택)
- 성능 최적화 튜닝 (선택)
- CI/CD 파이프라인 통합 패턴 (moai-ci-cd-expert 권장)

---

## 빠른 참고: 7가지 Enterprise 아티팩트 형식

```
1. Container Images (Docker/OCI)
   - 80-90% enterprise usage
   - Registry: Docker Hub, ECR, JFrog
   - Governance: Tag policy, scan on push, retention

2. Language Packages (Maven, npm, PyPI)
   - Package managers, semantic versioning
   - Artifact: wheel, jar, tgz
   - Policy: Version immutability, checksum validation

3. Binary Artifacts (Executable, .so, .dll)
   - Native code, compiled binaries
   - Storage: Artifactory, Nexus
   - Compliance: Signature verification, SBOM

4. Documentation Artifacts
   - API docs, SDK docs, guides
   - Format: HTML, PDF, Markdown
   - Governance: Version lockdown, archive policy

5. Configuration & Templates
   - IaC (Terraform, CloudFormation, Helm)
   - Format: JSON, YAML, HCL
   - Validation: Schema check, secret scan

6. Test & Compliance Reports
   - Coverage reports, scan results, audit logs
   - Format: JSON, XML, CSV
   - Retention: 90+ days per compliance

7. Source Code Archives
   - Release tarballs, source builds
   - Format: .tar.gz, .zip
   - Verification: GPG sign, checksum
```

---

## 핵심 규칙: 아티팩트 거버넌스 프레임워크

### Rule #1: Classification First (분류 먼저)

모든 아티팩트는 위 7가지 중 **하나**로 분류:

```yaml
artifact:
  id: "app-service@1.0.0"
  type: "Container Image"  # ← 필수
  format: "Docker OCI"     # ← 필수
  registry: "docker.io"
  location: "docker.io/myorg/app-service:1.0.0"
  
metadata:
  created: 2025-11-12T14:30:00Z
  created_by: "GoosLab"
  signature: "sha256:abc123..."
  sbom_url: "https://sbom.registry/app-service@1.0.0.json"
```

### Rule #2: Provenance Tracking (출처 추적)

모든 아티팩트는 **완전한 provenance**를 포함:

```yaml
provenance:
  source_repo: "https://github.com/myorg/app-service"
  source_commit: "abc123def456"  # ← SOURCE_CONTROL 링크 필수
  builder: "GitHub Actions CI/CD"
  build_job: "workflow@run-1234"
  timestamp: 2025-11-12T14:30:00Z
  
audit_trail:
  - event: "built"
    actor: "ci-bot"
    time: "2025-11-12T14:30:00Z"
  - event: "scanned"
    actor: "security-scanner"
    tool: "Trivy v0.54.0"
    time: "2025-11-12T14:31:00Z"
  - event: "approved"
    actor: "release-manager"
    time: "2025-11-12T14:35:00Z"
  - event: "published"
    actor: "cd-bot"
    time: "2025-11-12T14:36:00Z"
```

### Rule #3: Immutability & Integrity (불변성 & 무결성)

모든 아티팩트는 **게시 후 수정 불가**:

```yaml
# ❌ 금지: 게시된 아티팩트 수정
docker tag app:1.0.0 app:1.0.0-fixed  # 버전 변경

# ✅ 정책: 새 버전으로 재배포
docker tag app:1.0.1 docker.io/myorg/app:1.0.1  # ← 새 버전
```

### Rule #4: SBOM Generation (부품 목록 생성)

모든 아티팩트는 **SBOM(Software Bill of Materials) 포함**:

```yaml
sbom:
  version: "CycloneDX 1.6"
  spec_version: "1.6"
  version: "1.0.0"
  
  metadata:
    timestamp: 2025-11-12T14:30:00Z
    tools:
      - name: "syft"
        version: "0.95.0"
  
  components:
    - type: "library"
      name: "requests"
      version: "2.31.0"
      purl: "pkg:pypi/requests@2.31.0"
      licenses: [{ name: "Apache-2.0" }]
    
    - type: "library"
      name: "pydantic"
      version: "2.5.0"
      purl: "pkg:pypi/pydantic@2.5.0"
      licenses: [{ name: "MIT" }]
```

### Rule #5: Compliance & Scanning (준수 & 스캔)

모든 아티팩트는 **배포 전** 보안 검증:

```yaml
security_checks:
  vulnerability_scan:
    tool: "Trivy"
    version: "0.54.0"
    severity_threshold: "HIGH"  # CRITICAL, HIGH 차단
    status: "passed"
    scan_time: 2025-11-12T14:31:00Z
    results:
      - cve: "CVE-2024-1234"
        severity: "CRITICAL"
        status: "patched"
  
  sbom_validation:
    compliant: true
    licenses_approved: true
    restricted_licenses: []
  
  signature_verification:
    signed: true
    signer: "release-key@myorg.com"
    algorithm: "RSA-2048"
    timestamp: 2025-11-12T14:35:00Z
```

---

## 아티팩트 생명주기 (5단계 파이프라인)

### 1️⃣ CREATION (생성)

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

### 2️⃣ VALIDATION (검증)

```yaml
phase: "validation"
gates:
  - ✅ Vulnerability scan (Trivy, Grype)
  - ✅ License compliance (FOSSA, Black Duck)
  - ✅ SBOM validation (cyclonedx-python)
  - ✅ Signature verification (GPG/RSA)
  - ✅ Artifact integrity (checksum match)
  
failure_behavior:
  critical: "block_deployment"  # CRITICAL/HIGH 취약점 → 배포 차단
  high: "require_approval"      # 수동 승인 필요
  medium_low: "log_warning"     # 로깅만
```

### 3️⃣ STORAGE (저장)

```yaml
phase: "storage"
registries:
  - registry: "docker.io"
    type: "Container"
    location: "docker.io/myorg/app-service:1.0.0"
    replication: "global"
    
  - registry: "pypi.org"
    type: "Python"
    location: "pypi.org/packages/app_service-1.0.0-py3-none-any.whl"
    immutable: true
    
proxy_cache:
  enabled: true
  upstream: "docker.io"
  local_cache: "private-registry.myorg.com"
  eviction_policy: "lru"
  ttl_days: 30
```

### 4️⃣ DEPLOYMENT (배포)

```yaml
phase: "deployment"
release_strategy:
  - Publish manifest (Docker Hub, PyPI, etc.)
  - Announce release (changelog, blog, email)
  - Monitor rollout (health checks, metrics)
  - Rollback plan (if critical issues)
  
monitoring:
  - Deployment success rate
  - Security incident detection
  - Performance baseline
  - User adoption rate
```

### 5️⃣ RETIREMENT (폐기)

```yaml
phase: "retirement"
retention_policy:
  - Archive old versions (S3, cold storage)
  - Purge deprecated tags
  - Remove from active registries
  - Maintain audit logs (7 years compliance)
  
retention_timeline:
  - Keep active: 2 years
  - Archive: 5 years
  - Purge: 7+ years (compliance)
```

---

## 3-Level Progressive Disclosure

### Level 1: 시작 가이드 (Beginner - 10분)

**당신이 필요한 것**: 간단한 아티팩트 저장소 설정

```yaml
# 최소 필수 구성
artifact_repository:
  type: "Docker"
  registry: "docker.io"
  namespace: "myorg"
  
  authentication:
    method: "token"
    token_source: "env:DOCKER_TOKEN"
  
  retention_policy:
    max_versions: 10
    ttl_days: 90
  
  security:
    scan_on_push: true
    vulnerability_threshold: "HIGH"
```

**→ 다음**: Level 2로 진행하기

### Level 2: Enterprise 패턴 (Intermediate - 30분)

**당신이 필요한 것**: 멀티 형식 저장소, 거버넌스, SBOM

```yaml
# Enterprise 구성
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

compliance:
  sbom_required: true
  sbom_format: "CycloneDX"
  signature_required: true
  audit_retention_years: 7
```

**→ 다음**: Level 3로 진행하기

### Level 3: 고급 거버넌스 (Advanced - 1시간)

**당신이 필요한 것**: 공급망 보안, SLA, 자동화, Context7 MCP

```yaml
# Advanced 거버넌스
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
    
    example_query: |
      # Artifact provenance lookup
      Context7.query({
        operation: "artifact_metadata",
        artifact_id: "app-service@1.0.0",
        fields: ["sbom", "provenance", "signatures"]
      })
  
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

---

## 실제 아티팩트 생성 예제 (10+ 패턴)

### 1. Docker 컨테이너 이미지

```yaml
artifact:
  id: "api-gateway@2.5.1"
  type: "Container Image"
  format: "Docker OCI"
  registry: "docker.io"
  repository: "myorg/api-gateway"
  tag: "2.5.1"
  
  # Full image reference
  image: "docker.io/myorg/api-gateway:2.5.1"
  
  metadata:
    created: 2025-11-12T14:30:00Z
    os: "linux"
    arch: "amd64"
    size_mb: 125
    layers: 8
    
    creator: "github-actions"
    source_repo: "https://github.com/myorg/api-gateway"
    source_commit: "abc123def456789"
    source_branch: "main"
    
    signature:
      algorithm: "RSA-4096"
      signature: "MEUCIQDx..."
      timestamp: 2025-11-12T14:35:00Z
  
  sbom:
    version: "CycloneDX 1.6"
    components:
      - type: "library"
        name: "fastapi"
        version: "0.110.0"
        purl: "pkg:pypi/fastapi@0.110.0"
      - type: "library"
        name: "pydantic"
        version: "2.5.0"
        purl: "pkg:pypi/pydantic@2.5.0"
      - type: "framework"
        name: "python"
        version: "3.12.0"
        purl: "pkg:pypi/python@3.12.0"
  
  security:
    vulnerability_scan:
      tool: "Trivy v0.54.0"
      scanned_at: 2025-11-12T14:31:00Z
      status: "passed"
      critical_vulnerabilities: 0
      high_vulnerabilities: 0
      result_url: "https://scan.registry.myorg.com/api-gateway@2.5.1"
    
    signature_verified: true
    sbom_compliant: true
```

### 2. Python 패키지 (PyPI Wheel)

```yaml
artifact:
  id: "data-pipeline@1.3.2"
  type: "Language Package"
  format: "Python Wheel"
  manager: "pip/PyPI"
  
  filename: "data_pipeline-1.3.2-py3-none-any.whl"
  pypi_url: "https://pypi.org/project/data-pipeline/1.3.2/"
  
  metadata:
    created: 2025-11-12T13:45:00Z
    python_version: "3.8+"
    size_bytes: 245678
    checksum:
      md5: "d8e8fca2dc0f896fd7cb4cb0031ba249"
      sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    
    creator: "ci-automation"
    source_repo: "https://github.com/myorg/data-pipeline"
    source_commit: "xyz789abc123"
    
    dependencies:
      - name: "pandas"
        version: "2.1.0"
        requirement: "pandas>=2.1.0,<3.0"
      - name: "numpy"
        version: "1.24.0"
        requirement: "numpy>=1.24.0"
  
  sbom:
    version: "CycloneDX 1.6"
    components_count: 18
    licenses:
      - name: "MIT"
        count: 12
      - name: "Apache-2.0"
        count: 6
  
  security:
    license_scan: "passed"
    restricted_licenses: []
    checksum_verified: true
```

### 3. Binary 아티팩트 (Native Code)

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
    created: 2025-11-12T12:00:00Z
    compiler: "gcc-13"
    optimization_flags: "-O3 -march=native"
    source_commit: "release/3.1.0"
  
  distribution:
    storage: "artifactory.myorg.com"
    bucket: "native-binaries"
    access: "restricted"
    approval_required: true
```

### 4. 설정 & IaC 아티팩트 (Terraform)

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
    - path: "outputs.tf"
      size_bytes: 512
      checksum_sha256: "opq345..."
  
  metadata:
    created: 2025-11-12T10:30:00Z
    terraform_version: ">= 1.5"
    cloud_provider: "AWS"
    modules_included: 5
    
    creator: "infrastructure-team"
    source_repo: "https://github.com/myorg/terraform-modules"
    source_commit: "4.2.0"
  
  validation:
    terraform_fmt: "passed"
    terraform_validate: "passed"
    security_scan: "passed"
    
    secrets_detected: 0
    hardcoded_values: 0
```

### 5. 문서 아티팩트 (API 문서)

```yaml
artifact:
  id: "api-docs@1.0.0"
  type: "Documentation"
  format: "HTML/OpenAPI"
  
  files:
    - path: "index.html"
      size_bytes: 45678
    - path: "openapi-spec.json"
      size_bytes: 12345
    - path: "styles.css"
      size_bytes: 5678
  
  metadata:
    created: 2025-11-12T14:00:00Z
    format: "OpenAPI 3.1.0"
    endpoints_documented: 42
    examples_included: 15
    
    creator: "docs-automation"
    source_repo: "https://github.com/myorg/api-service"
    source_branch: "main"
  
  hosting:
    url: "https://api-docs.myorg.com"
    cdn: "cloudflare"
    ttl: "1 hour"
```

### 6. 테스트 리포트 아티팩트

```yaml
artifact:
  id: "test-report@build-12345"
  type: "Test Report"
  format: "JSON/JUnit"
  
  files:
    - name: "junit-report.xml"
      size_bytes: 34567
    - name: "coverage-report.json"
      size_bytes: 23456
    - name: "test-summary.html"
      size_bytes: 12345
  
  metadata:
    created: 2025-11-12T14:32:00Z
    build_number: 12345
    test_framework: "pytest"
    
    results:
      total_tests: 487
      passed: 485
      failed: 2
      skipped: 0
      coverage_percent: 87.5
  
  retention:
    keep_days: 90
    archived_to_s3: "test-archives"
```

### 7. SBOM 아티팩트 (독립형)

```yaml
artifact:
  id: "sbom@app-service-1.0.0"
  type: "SBOM"
  format: "CycloneDX JSON"
  
  file: "app-service-1.0.0.sbom.json"
  
  metadata:
    created: 2025-11-12T14:30:30Z
    generator: "syft v0.95.0"
    spec_version: "CycloneDX 1.6"
    
    artifact_reference: "docker.io/myorg/app-service:1.0.0"
  
  statistics:
    total_components: 127
    libraries: 120
    frameworks: 5
    operating_systems: 2
    
    licenses:
      - name: "MIT"
        count: 85
      - name: "Apache-2.0"
        count: 32
      - name: "GPL-3.0"
        count: 3
    
    vulnerabilities:
      critical: 0
      high: 1
      medium: 3
      low: 12
```

### 8. 서명 증명서 (Attestation)

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
          invocation:
            configSource:
              uri: "github.com/myorg/app-service"
              digest:
                sha256: "xyz789..."
          completion:
            finishTime: "2025-11-12T14:35:00Z"
```

### 9. 릴리스 번들 (GitHub Release)

```yaml
artifact:
  id: "release@v2.0.0"
  type: "Release Bundle"
  format: "GitHub Release"
  
  release:
    tag: "v2.0.0"
    name: "Version 2.0.0 - Production Release"
    draft: false
    prerelease: false
    created: 2025-11-12T14:40:00Z
    
    assets:
      - name: "app-service-2.0.0.tar.gz"
        size_bytes: 5242880
        download_url: "https://github.com/myorg/app-service/releases/download/v2.0.0/app-service-2.0.0.tar.gz"
      
      - name: "app-service-2.0.0.tar.gz.sha256"
        content: "abc123def456..."
      
      - name: "SBOM.json"
        size_bytes: 123456
      
      - name: "SBOM.json.sig"
        size_bytes: 256
    
    changelog: |
      ## What's New
      - Feature A
      - Feature B
      - Bug fixes
      
      ## Breaking Changes
      - Configuration format updated
```

### 10. Helm 차트 아티팩트

```yaml
artifact:
  id: "helm-chart@5.1.0"
  type: "Kubernetes Configuration"
  format: "Helm Chart"
  
  chart:
    name: "myapp"
    version: "5.1.0"
    appVersion: "2.0.0"
    
    metadata:
      description: "Production Helm Chart for MyApp"
      type: "application"
      home: "https://github.com/myorg/myapp"
      sources:
        - "https://github.com/myorg/myapp"
      keywords:
        - "application"
        - "production"
  
  files:
    - path: "Chart.yaml"
    - path: "values.yaml"
    - path: "values.schema.json"
    - path: "templates/"
      templates_count: 12
  
  metadata:
    created: 2025-11-12T11:30:00Z
    maintainers:
      - name: "DevOps Team"
        email: "devops@myorg.com"
    
    validation:
      helm_lint: "passed"
      kubernetes_validate: "passed"
      security_scan: "passed"
  
  distribution:
    registry: "helm.myorg.com"
    repository: "stable"
    pullable: true
    signature_required: true
```

---

## Context7 MCP 통합 가이드

### Artifact Index 조회

```python
# Context7 쿼리: 아티팩트 메타데이터 조회
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

# 응답 구조
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

### SBOM Index 검색

```python
# Context7 쿼리: 특정 라이브러리 버전을 포함한 아티팩트 찾기
context7_query = {
    "operation": "sbom_index_search",
    "search_type": "component",
    "component_name": "requests",
    "version_range": ">=2.30.0",
    "fields": ["artifact_id", "vulnerability_status"]
}

# 응답
response = {
    "results": [
        {
            "artifact_id": "api-gateway@2.5.1",
            "component_version": "2.31.0",
            "vulnerability_status": "clean",
            "sbom_updated": "2025-11-12T14:31:00Z"
        },
        {
            "artifact_id": "data-pipeline@1.3.2",
            "component_version": "2.30.1",
            "vulnerability_status": "1 medium CVE",
            "sbom_updated": "2025-11-12T13:45:00Z"
        }
    ]
}
```

### Vulnerability 상관관계 분석

```python
# Context7 쿼리: CVE가 영향을 미치는 모든 아티팩트 찾기
context7_query = {
    "operation": "vulnerability_correlation",
    "cve_id": "CVE-2024-5678",
    "affected_components": ["requests", "urllib3"],
    "action": "find_artifacts"
}

# 응답
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
    ],
    "remediation_steps": [...]
}
```

---

## 실무 체크리스트

### 아티팩트 생성 체크리스트

- [ ] **분류**: 7가지 중 정확한 타입 선택
- [ ] **메타데이터**: 생성자, 시간, 소스 커밋, 빌더 정보 포함
- [ ] **Provenance**: 소스 repo, 커밋 SHA, 빌드 로그 링크
- [ ] **SBOM**: CycloneDX 또는 SPDX 형식
- [ ] **서명**: RSA-4096 또는 ECDSA-P256
- [ ] **스캔**: Trivy/Grype로 취약점 검사
- [ ] **라이선스**: 제한된 라이선스 확인
- [ ] **불변성**: 게시 후 수정 불가 정책 적용
- [ ] **감시**: 버전 태깅, 생명주기 추적
- [ ] **문서화**: 변경 로그, 릴리스 노트

### 저장소 설계 체크리스트

- [ ] **멀티 형식 지원**: Container, Python, Binary, IaC, Docs
- [ ] **레지스트리 구성**: 공식(upstream), 프록시 캐시(로컬)
- [ ] **RBAC**: 관리자, 게시자, 읽기 권한 분리
- [ ] **승인 워크플로우**: 보안팀, 릴리스 매니저 승인
- [ ] **스캔 자동화**: push 시 자동 취약점 검사
- [ ] **SBOM 필수**: 모든 아티팩트
- [ ] **서명 검증**: 배포 전 필수
- [ ] **감시 정책**: 버전 유지, TTL 설정, 자동 정리
- [ ] **감시 추적**: 모든 이벤트 로깅, 7년 보관
- [ ] **규정 준수**: SOC 2, ISO 27001 준수 확인

---

## 공식 자료 & 참고자료

### 2025 Enterprise Standards

- Cloudsmith Artifact Management Report 2025
  https://cloudsmith.com/blog/artifact-management-a-complete-guide

- JFrog Artifact Management Platform
  https://jfrog.com/artifact-management/

- Harness Artifact Lifecycle Management
  https://www.harness.io/harness-devops-academy/artifact-lifecycle-management-strategies

### SBOM & Supply Chain Security

- CycloneDX Official Specification v1.6
  https://cyclonedx.org/specification/

- SPDX License List
  https://spdx.org/licenses/

- SLSA Framework (Supply chain Levels for Software Artifacts)
  https://slsa.dev/

- in-toto: Attestation Framework
  https://in-toto.io/

### Security & Compliance

- OWASP Dependency Checking Best Practices
  https://owasp.org/

- Trivy Container Image Scanning
  https://aquasecurity.github.io/trivy/

- Syft SBOM Generation
  https://github.com/anchore/syft

- SOC 2 Compliance Requirements
  https://www.soc2.org/

- ISO 27001 Information Security Management
  https://www.iso.org/standard/27001

### Registry & Storage

- Docker Registry v2 Specification
  https://docs.docker.com/registry/spec/api/

- OCI Distribution Spec
  https://github.com/opencontainers/distribution-spec

- PyPI Package Repository
  https://pypi.org/

- Maven Central Repository
  https://mvnrepository.com/

---

## 다음 단계

1. **저장소 설계**: 아티팩트 분류 및 레지스트리 선택
2. **거버넌스 정책**: RBAC, 승인 워크플로우, 감시 정책
3. **자동화 구축**: CI/CD 통합, 스캔 자동화, SBOM 생성
4. **Context7 MCP**: 아티팩트 인덱스 조회 및 검색 통합
5. **규정 준수**: SOC 2, ISO 27001 감시 계획
6. **모니터링**: SLA 설정, 경고 구성, 성능 측정

---

**마지막 업데이트**: 2025-11-12 (November 2025 stable)
**버전**: 4.1.0
**유지보수자**: GoosLab (enterprise artifact governance)
