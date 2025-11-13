# moai-artifacts-builder Examples (v4.1.0)

## 실제 구현 예제 (7 Production Patterns)

### Example 1: Docker 이미지 빌드 & SBOM 생성

**상황**: Python FastAPI 애플리케이션을 Docker로 패키징

```bash
# 1. 이미지 빌드
docker build -t myorg/api-service:1.0.0 .

# 2. SBOM 생성 (Trivy)
trivy image --format cyclonedx \
  myorg/api-service:1.0.0 > sbom.json

# 3. 이미지 서명
docker push myorg/api-service:1.0.0
gpg --armor --detach-sign api-service:1.0.0.tar

# 4. 레지스트리에 메타데이터 저장
curl -X POST https://registry.myorg.com/artifacts/metadata \
  -H "Content-Type: application/json" \
  -d '{
    "artifact_id": "api-service@1.0.0",
    "type": "container_image",
    "registry": "docker.io"
  }'
```

---

### Example 2: Python 패키지 배포 (PyPI)

```bash
# 1. Wheel 생성
python -m build --wheel

# 2. SBOM 생성
syft dist/data_pipeline-1.3.2-py3-none-any.whl \
  --output cyclonedx-json > sbom.json

# 3. Checksum 계산
sha256sum dist/data_pipeline-1.3.2-py3-none-any.whl > checksum.sha256

# 4. 서명
gpg --armor --detach-sign dist/data_pipeline-1.3.2-py3-none-any.whl

# 5. 취약점 스캔
grype dist/ --output json > security-scan.json

# 6. PyPI에 배포
twine upload dist/*.whl --config-file ~/.pypirc
```

---

### Example 3: Terraform 모듈 거버넌스

```bash
# 1. 형식 검증
terraform fmt -recursive -check .
terraform validate

# 2. 보안 스캔
tflint --init
tflint --format compact

# 3. SBOM 생성
cat > sbom.json << 'SBOM'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.6",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "aws-provider",
      "version": "5.31.0"
    }
  ]
}
SBOM

# 4. 체크섬
sha256sum main.tf variables.tf > checksums.txt

# 5. 릴리스 생성
git tag -a v4.2.0 -m "Release v4.2.0"
gh release create v4.2.0 --notes "New features"
```

---

### Example 4: 아티팩트 거버넌스 모니터링 (Python)

```python
import json
import requests
from datetime import datetime, timedelta

class ArtifactMonitor:
    def __init__(self, api_url, api_token):
        self.api_url = api_url
        self.headers = {"Authorization": f"Bearer {api_token}"}
    
    def get_artifacts(self, days=30):
        endpoint = f"{self.api_url}/artifacts"
        params = {
            "created_after": (datetime.now() - timedelta(days=days)).isoformat()
        }
        response = requests.get(endpoint, headers=self.headers, params=params)
        return response.json().get("artifacts", [])
    
    def validate_sbom(self, artifact):
        checks = {
            "has_sbom": "sbom_url" in artifact,
            "format_valid": False,
            "components_scanned": False
        }
        
        if checks["has_sbom"]:
            sbom_response = requests.get(
                artifact["sbom_url"],
                headers=self.headers
            )
            sbom = sbom_response.json()
            
            checks["format_valid"] = sbom.get("bomFormat") == "CycloneDX"
            checks["components_scanned"] = len(sbom.get("components", [])) > 0
        
        return checks
    
    def generate_report(self):
        artifacts = self.get_artifacts()
        
        report = {
            "generated_at": datetime.now().isoformat(),
            "total_artifacts": len(artifacts),
            "compliance": {
                "sbom_compliant": 0,
                "signature_verified": 0,
                "vulnerability_free": 0
            }
        }
        
        for artifact in artifacts:
            sbom_checks = self.validate_sbom(artifact)
            if all(sbom_checks.values()):
                report["compliance"]["sbom_compliant"] += 1
        
        return report

# Usage
if __name__ == "__main__":
    monitor = ArtifactMonitor(
        api_url="https://artifacts.myorg.com",
        api_token="your-token"
    )
    report = monitor.generate_report()
    print(json.dumps(report, indent=2))
```

---

### Example 5: CI/CD 자동화 (GitHub Actions)

```yaml
# .github/workflows/artifact-release.yml
name: Build & Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build
        run: python -m build
      
      - name: Generate SBOM
        run: |
          syft dist/ \
            --output cyclonedx-json \
            > sbom.json
      
      - name: Vulnerability Scan
        run: |
          grype dist/ \
            --output json \
            > security-scan.json
      
      - name: Checksums
        run: sha256sum dist/* > checksums.sha256
      
      - name: Sign
        env:
          GPG_PRIVATE_KEY: ${{ secrets.GPG_PRIVATE_KEY }}
          GPG_PASSPHRASE: ${{ secrets.GPG_PASSPHRASE }}
        run: |
          echo "$GPG_PRIVATE_KEY" | gpg --import
          for file in dist/*; do
            echo "$GPG_PASSPHRASE" | gpg --armor --detach-sign "$file"
          done
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            dist/*
            sbom.json
            security-scan.json
            checksums.sha256
```

---

### Example 6: 저장소 설정 (YAML)

```yaml
# artifact-registry-config.yaml
repositories:
  
  container-registry:
    type: "container"
    upstream_url: "https://docker.io"
    proxy_cache:
      enabled: true
      max_age_days: 30
    security:
      scan_on_push: true
      scanner: "trivy"
      severity_threshold: "HIGH"
      sbom_required: true
    retention_policy:
      keep_latest: 10
      cleanup_after_days: 90
  
  python-registry:
    type: "python"
    upstream_url: "https://pypi.org"
    security:
      scan_on_push: true
      license_check: true
    retention_policy:
      immutable_releases: true
      keep_prerelease: 3

governance:
  rbac:
    admin: ["release-engineering"]
    publisher: ["ci-automation"]
    consumer: ["developers"]
  
  approval_workflow:
    require_approval: true
    timeout_hours: 24
  
  compliance:
    standards: ["SOC2", "ISO27001"]
    sbom_required: true
    audit_retention_years: 7
```

---

### Example 7: Vulnerability 추적 (Context7 MCP)

```python
import context7_client

client = context7_client.Context7Client(
    endpoint="https://context7.myorg.com",
    api_key="your-api-key"
)

# CVE-2024-5678 영향받는 아티팩트 찾기
response = client.query(
    operation="vulnerability_correlation",
    cve_id="CVE-2024-5678",
    affected_components=["requests", "urllib3"],
    action="find_artifacts"
)

print(f"Affected Artifacts: {len(response['affected_artifacts'])}")

for artifact in response["affected_artifacts"]:
    print(f"- {artifact['artifact_id']}")
    print(f"  Status: {artifact['status']}")
    if artifact.get("patch_available"):
        print(f"  Patch: {artifact['patched_version']}")

# Remediation Plan 생성
if response["affected_artifacts"]:
    client.notify(
        incident_type="critical_vulnerability",
        severity="CRITICAL",
        message=f"CVE affecting {len(response['affected_artifacts'])} artifacts",
        recipients=["security-team@myorg.com"]
    )
```

---

## 체크리스트

### 아티팩트 생성 체크리스트

- [ ] 분류: 7가지 중 정확한 타입 선택
- [ ] 메타데이터: 생성자, 시간, 소스 커밋
- [ ] SBOM: CycloneDX 또는 SPDX 형식
- [ ] 서명: RSA-4096 또는 ECDSA-P256
- [ ] 취약점 스캔: Trivy/Grype
- [ ] 라이선스: 제한된 라이선스 확인
- [ ] 불변성: 게시 후 수정 불가
- [ ] 문서화: 변경 로그, 릴리스 노트

### 저장소 설계 체크리스트

- [ ] 멀티 형식 지원
- [ ] RBAC 설정
- [ ] 승인 워크플로우
- [ ] 스캔 자동화
- [ ] SBOM 필수
- [ ] 서명 검증
- [ ] 감시 정책
- [ ] 규정 준수 확인

---

**마지막 업데이트**: 2025-11-12
**버전**: 4.1.0
