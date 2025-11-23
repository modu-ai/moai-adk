# DevSecOps Automation

Integrate security into CI/CD pipeline for continuous security.

## Security Scanner Integration

### Static Application Security Testing (SAST)

```yaml
# GitHub Actions workflow
name: SAST Scan

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  sast:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Bandit
      run: |
        pip install bandit
        bandit -r src/ -f json -o bandit-report.json
    
    - name: Run SonarQube
      uses: SonarSource/sonarqube-scan-action@master
      env:
        SONAR_HOST_URL: ${{ secrets.SONAR_HOST_URL }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
    
    - name: Check coverage threshold
      run: |
        coverage run -m pytest tests/
        coverage report --fail-under=80
    
    - name: Upload results
      uses: actions/upload-artifact@v3
      with:
        name: security-reports
        path: |
          bandit-report.json
          .scannerwork/
```

### Dependency Vulnerability Scanning

```python
# Dependency check implementation
class DependencyScanner:
    def scan_vulnerabilities(self) -> dict:
        vulnerabilities = {}

        # Python dependencies
        result = subprocess.run(
            ['safety', 'check', '--json'],
            capture_output=True,
            text=True
        )

        vulnerabilities['python'] = json.loads(result.stdout)

        # Node.js dependencies
        result = subprocess.run(
            ['npm', 'audit', '--json'],
            capture_output=True,
            text=True
        )

        vulnerabilities['npm'] = json.loads(result.stdout)

        return vulnerabilities

    def fail_on_critical(self, vulnerabilities: dict) -> bool:
        for lang, vulns in vulnerabilities.items():
            critical_count = len([
                v for v in vulns
                if v.get('severity') == 'critical'
            ])

            if critical_count > 0:
                return True

        return False
```

## Container Security

### Docker Image Scanning

```yaml
name: Container Security

on:
  push:
    branches: [main]

jobs:
  container-scan:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Build Docker image
      run: docker build -t myapp:${{ github.sha }} .
    
    - name: Scan with Trivy
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: myapp:${{ github.sha }}
        format: 'sarif'
        output: 'trivy-results.sarif'
    
    - name: Upload Trivy results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'
```

## Secrets Management

### Secrets Detection in Code

```python
# Pre-commit hook for secrets detection
import subprocess
import re

class SecretsDetector:
    def detect_secrets_in_commit(self, files: list) -> list:
        patterns = {
            'aws_key': r'AKIA[0-9A-Z]{16}',
            'private_key': r'-----BEGIN RSA PRIVATE KEY-----',
            'api_key': r'[A-Za-z0-9]{32,}',
            'password': r'password\s*=\s*['"][^'"]+['"]',
            'token': r'(token|secret)\s*[:=]\s*['"][^'"]+['"]'
        }

        found_secrets = []

        for file_path in files:
            with open(file_path, 'r') as f:
                content = f.read()

                for pattern_name, pattern in patterns.items():
                    matches = re.finditer(pattern, content)

                    for match in matches:
                        found_secrets.append({
                            'file': file_path,
                            'pattern': pattern_name,
                            'match': match.group()[:20] + '...'
                        })

        return found_secrets
```

## Security Reporting

### Dashboard Integration

```python
class SecurityDashboard:
    def aggregate_metrics(self) -> dict:
        return {
            'vulnerabilities': {
                'critical': 0,
                'high': 5,
                'medium': 12,
                'low': 23
            },
            'code_coverage': 85,
            'test_pass_rate': 98,
            'last_security_audit': '2025-11-20',
            'compliance_status': {
                'soc2': 'Compliant',
                'iso27001': 'Audit Scheduled',
                'gdpr': 'Compliant'
            }
        }

    def generate_report(self) -> str:
        metrics = self.aggregate_metrics()

        report = f"""
        Security Dashboard Report
        
        Vulnerabilities: {metrics['vulnerabilities']['critical']} critical, {metrics['vulnerabilities']['high']} high
        Code Coverage: {metrics['code_coverage']}%
        Test Pass Rate: {metrics['test_pass_rate']}%
        Compliance: {metrics['compliance_status']['soc2']}
        """

        return report
```

---

**Tools**: GitHub Actions, GitLab CI, Jenkins, SonarQube, Trivy, Bandit
