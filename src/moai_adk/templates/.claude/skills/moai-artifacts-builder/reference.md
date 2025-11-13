# moai-artifacts-builder Reference (v4.1.0)

## 공식 자료 & 링크 (November 2025 Enterprise Standards)

### 1. Core References (핵심 표준)

| 자료 | 링크 | 설명 |
|-----|------|------|
| **Cloudsmith Artifact Management Report 2025** | https://cloudsmith.com/blog/artifact-management-a-complete-guide | Enterprise artifact governance 표준, 2025 트렌드 |
| **JFrog Artifact Management Platform** | https://jfrog.com/artifact-management/ | Enterprise registry solutions, 모범 사례 |
| **Harness Artifact Lifecycle** | https://www.harness.io/harness-devops-academy/artifact-lifecycle-management-strategies | ALM 전략, 자동화 패턴 |
| **GitLab Artifact Management** | https://about.gitlab.com/blog/streamline-enterprise-artifact-management-with-gitlab | CI/CD 통합, 스트림라이닝 |

### 2. SBOM & Supply Chain Security

| 자료 | 링크 | 설명 |
|-----|------|------|
| **CycloneDX Official v1.6** | https://cyclonedx.org/specification/ | SBOM 표준, schema |
| **SPDX License List** | https://spdx.org/licenses/ | 라이선스 정의 및 호환성 |
| **SLSA Framework (Levels 0-4)** | https://slsa.dev/ | 공급망 보안 레벨, provenance |
| **in-toto Attestation** | https://in-toto.io/ | 서명 및 증명 프레임워크 |

### 3. Container Registry Standards

| 자료 | 링크 | 설명 |
|-----|------|------|
| **Docker Registry v2 API** | https://docs.docker.com/registry/spec/api/ | Registry 프로토콜 |
| **OCI Distribution Spec** | https://github.com/opencontainers/distribution-spec | Open Container Initiative |
| **OCI Image Spec** | https://github.com/opencontainers/image-spec | 이미지 포맷 표준 |

### 4. Security Scanning Tools

| 도구 | 링크 | 설명 |
|-----|------|------|
| **Trivy** | https://aquasecurity.github.io/trivy/ | 컨테이너/이미지 취약점 스캔 |
| **Grype** | https://github.com/anchore/grype | 멀티 아티팩트 취약점 감지 |
| **Syft** | https://github.com/anchore/syft | SBOM 생성 (CycloneDX, SPDX) |

### 5. Language Package Managers

| 패키지 관리자 | 링크 | 아티팩트 형식 |
|-------------|------|--------------|
| **PyPI (Python)** | https://pypi.org/ | .whl, .tar.gz |
| **npm (JavaScript)** | https://www.npmjs.com/ | .tgz |
| **Maven Central** | https://mvnrepository.com/ | .jar, .pom |
| **RubyGems** | https://rubygems.org/ | .gem |

### 6. Compliance & Governance

| 표준 | 링크 | 설명 |
|-----|------|------|
| **SOC 2 Type II** | https://www.soc2.org/ | 보안, 가용성, 기밀성 감사 |
| **ISO 27001** | https://www.iso.org/standard/27001 | 정보 보안 관리 표준 |
| **NIST Supply Chain Risk** | https://csrc.nist.gov/Projects/supply-chain-risk-management/ | 연방 기준 |

### 7. Context7 MCP Integration

| 기능 | 설명 |
|-----|------|
| `artifact_metadata_lookup` | Artifact ID로 메타데이터, SBOM, provenance, signatures 조회 |
| `sbom_index_search` | Component 이름/버전으로 영향받는 아티팩트 검색 |
| `vulnerability_correlation` | CVE ID로 영향받는 모든 아티팩트 찾기 |
| `compliance_status_check` | SOC 2, ISO 27001 준수 상태 검증 |

---

## 아티팩트 분류 매트릭스

```
Type 1: Container Images (Docker/OCI)
├─ Registry: Docker Hub, ECR, JFrog
├─ Format: .tar (OCI), manifest.json
├─ Enterprise %: 80-90%
└─ Governance: Tag policy, scan on push, retention

Type 2: Language Packages (Maven, npm, PyPI)
├─ Registry: PyPI, npm, Maven Central
├─ Format: .whl, .jar, .tgz
├─ Enterprise %: 60-70%
└─ Governance: Version immutability, checksum validation

Type 3: Binary Artifacts
├─ Registry: Artifactory, Nexus
├─ Format: .so, .dll, executable
├─ Enterprise %: 40-50%
└─ Governance: Signature verification, SBOM

Type 4: Documentation
├─ Registry: Static hosting, CDN
├─ Format: HTML, PDF, Markdown
├─ Enterprise %: 30-40%
└─ Governance: Version lockdown, archive policy

Type 5: Configuration & IaC
├─ Registry: GitHub, artifact storage
├─ Format: .tf, .yaml, .json
├─ Enterprise %: 70-80%
└─ Governance: Schema validation, secret scan

Type 6: Test & Compliance Reports
├─ Registry: Artifact storage
├─ Format: JSON, XML, CSV
├─ Enterprise %: 50-60%
└─ Governance: Retention (90+ days compliance)

Type 7: Source Code Archives
├─ Registry: GitHub Releases
├─ Format: .tar.gz, .zip
├─ Enterprise %: 20-30%
└─ Governance: GPG sign, checksum verification
```

---

## 명령어 & 스크립트 참고

### Docker CLI - SBOM 생성

```bash
# Trivy로 SBOM 생성 (CycloneDX 형식)
trivy image --format cyclonedx docker.io/myorg/app:1.0.0 > sbom.json

# Syft로 SBOM 생성
syft docker.io/myorg/app:1.0.0 --output cyclonedx-json > sbom.json
```

### PyPI - 패키지 배포

```bash
# Wheel 생성
python -m build --wheel

# PyPI에 업로드 (twine)
twine upload dist/app-1.0.0-py3-none-any.whl

# Checksum 검증
sha256sum dist/app-1.0.0-py3-none-any.whl
```

### Artifact Signing

```bash
# GPG로 아티팩트 서명
gpg --armor --detach-sign artifact.tar.gz

# 서명 검증
gpg --verify artifact.tar.gz.asc artifact.tar.gz
```

### Context7 MCP Query (Python)

```python
import context7_client

client = context7_client.Context7Client()

# Artifact metadata 조회
artifact = client.query(
    operation="artifact_metadata_lookup",
    artifact_id="app-service@1.0.0",
    fields=["sbom", "provenance", "signatures"]
)

# SBOM index 검색
results = client.query(
    operation="sbom_index_search",
    search_type="component",
    component_name="requests",
    version_range=">=2.30.0"
)

# Vulnerability correlation
vulnerabilities = client.query(
    operation="vulnerability_correlation",
    cve_id="CVE-2024-5678",
    action="find_artifacts"
)
```

---

## 성능 벤치마크 (2025 Enterprise Baseline)

| 작업 | 목표 | 도구 | 비고 |
|-----|------|------|------|
| SBOM 생성 | < 30초 | Syft v0.95+ | Container image 기준 |
| 취약점 스캔 | < 5분 | Trivy v0.54+ | 1000+ components |
| 서명 검증 | < 1초 | GPG/RSA | 1회당 |
| SBOM 색인화 | < 10초 | Context7 | 1000 artifacts |
| Artifact 배포 | < 2분 | CI/CD 통합 | 검증 포함 |

---

## 운영 SLA 목표

| 지표 | 목표 | 모니터링 |
|-----|------|---------|
| Artifact 가용성 | 99.99% | Uptime monitoring |
| Scan 완료 | < 5분 | Pipeline logs |
| 배포 성공률 | > 95% | Deployment tracking |
| 평균 복구 시간 (MTTR) | < 15분 | Incident response |
| 규정 감시 | 7년 보관 | Audit logs |

---

**마지막 업데이트**: 2025-11-12
**버전**: 4.1.0
**유지보수자**: GoosLab
