# moai-security-devsecops: References & External Resources

## Official Documentation & Standards

### SAST Tools

| Tool | Version | Documentation | License |
|------|---------|---|---|
| **SonarQube** | 10.8+ | https://docs.sonarqube.org | SSPL/Community |
| **Snyk** | 1.1300+ | https://docs.snyk.io | Commercial/Free |
| **CodeQL** | Latest | https://codeql.github.com/docs | Proprietary |
| **Semgrep** | 1.45+ | https://semgrep.dev/docs | Open Source (LGPL) |

### DAST Tools

| Tool | Version | Documentation | Best For |
|------|---------|---|---|
| **OWASP ZAP** | 2.14+ | https://www.zaproxy.org/docs | Open Source Web Apps |
| **Burp Suite** | 2024.x | https://portswigger.net/burp/documentation | Enterprise API Testing |

### SCA & Container Scanning

| Tool | Latest Version | Documentation |
|------|---|---|
| **Trivy** | 0.58+ | https://aquasecurity.github.io/trivy |
| **Dependency-Check** | 9.0+ | https://jeremylong.github.io/DependencyCheck |
| **Syft** | 1.0+ | https://github.com/anchore/syft |

### Security Standards & Frameworks

| Framework | Focus | Website |
|-----------|-------|---------|
| **OWASP Top 10** | Application Security Risks | https://owasp.org/www-project-top-ten |
| **CWE Top 25** | Software Weaknesses | https://cwe.mitre.org/top25 |
| **CVSS v3.1** | Vulnerability Scoring | https://www.first.org/cvss/v3.1 |
| **NIST Cybersecurity Framework** | Risk Management | https://www.nist.gov/cyberframework |
| **ISO 27001** | Information Security | https://www.iso.org/isoiec-27001-information-security-management.html |
| **PCI DSS 4.0** | Payment Security | https://www.pcisecuritystandards.org |

### CI/CD Security

| Platform | Security Docs |
|----------|---|
| **GitHub Actions** | https://docs.github.com/en/actions/security-guides |
| **GitLab CI/CD** | https://docs.gitlab.com/ee/ci/cloud_services_security |
| **Jenkins** | https://www.jenkins.io/security |

### Vulnerability Databases

| Database | Content | Access |
|----------|---------|--------|
| **National Vulnerability Database (NVD)** | Official CVE registry | https://nvd.nist.gov (Free API) |
| **Common Vulnerabilities and Exposures (CVE)** | CVE Details | https://cve.mitre.org |
| **GitHub Security Advisory** | GitHub-specific vulns | https://github.com/advisories |

---

## Configuration Files & Tools

### Pre-commit Hooks Setup

```bash
# Install pre-commit framework
pip install pre-commit

# .pre-commit-config.yaml example
repos:
  - repo: https://github.com/returntocorp/semgrep
    rev: v1.45.0
    hooks:
      - id: semgrep
        args: ['--config=p/security-audit']
  
  - repo: https://github.com/gitpython-developers/gitpython
    rev: 3.1.40
    hooks:
      - id: detect-secrets
        args: ['scan', '--baseline', '.secrets.baseline']
```

### Environment Setup

```bash
# Java 17 (required for SonarScanner)
java -version

# Python 3.11+
python --version

# Node 18+ (for npm packages)
node --version

# Docker (for container scanning)
docker --version
```

---

## Quick Integration Guide

### GitHub Actions One-Liner

```bash
# Copy example workflow
curl -o .github/workflows/security.yml \
  https://raw.githubusercontent.com/moai-framework/security-devsecops/main/examples/github-actions-security.yml

# Enable GitHub Advanced Security
# Settings → Code security and analysis → Enable
```

### Local Development Setup

```bash
# Install security tools locally
pip install bandit pylint safety

# Run SAST locally before commit
bandit -r src/

# Run SCA locally
safety check --requirements requirements.txt
```

---

## Useful Scripts & Automation

### Vulnerability Severity Calculator

```python
import math

def calculate_cvss3_score(av, ac, pr, ui, s, c, i, a):
    """Calculate CVSS 3.1 score from vectors"""
    # Vector values mapping
    av_map = {'N': 0.85, 'A': 0.62, 'L': 0.55, 'P': 0.2}
    ac_map = {'L': 0.77, 'H': 0.44}
    pr_map = {'N': 0.85, 'L': 0.62, 'H': 0.27}
    ui_map = {'N': 0.85, 'R': 0.62}
    impact_map = {'N': 0, 'L': 0.22, 'H': 0.56}
    
    # Calculate base score
    scope_changed = (s == 'C')
    impact = 1 - ((1 - impact_map[c]) * (1 - impact_map[i]) * (1 - impact_map[a]))
    exploitability = av_map[av] * ac_map[ac] * pr_map[pr] * ui_map[ui]
    
    if scope_changed:
        base_score = min(10, (7.52 * exploitability * impact) + 0.029 - 
                              0.02 * (1 - impact))
    else:
        base_score = min(10, exploitability * impact * 6.5)
    
    return round(base_score, 1)
```

### Remediation Status Tracker

```bash
#!/bin/bash
# Query all open security issues

CRITICAL=$(gh issue list --label "security,critical" --json title | wc -l)
HIGH=$(gh issue list --label "security,high" --json title | wc -l)
MEDIUM=$(gh issue list --label "security,medium" --json title | wc -l)

echo "Security Issues Status"
echo "Critical: $CRITICAL"
echo "High: $HIGH"
echo "Medium: $MEDIUM"

# Send Slack notification
curl -X POST $SLACK_WEBHOOK \
  -H 'Content-type: application/json' \
  -d "{\"text\":\"Security issues: $CRITICAL critical, $HIGH high, $MEDIUM medium\"}"
```

---

## Learning Resources

### Hands-on Labs

| Lab | Platform | Focus |
|-----|----------|-------|
| **WebGoat** | OWASP | Web application security |
| **Vulnerable Applications** | DVWA | Web vulnerability practice |
| **Juice Shop** | OWASP | Full-stack app security |
| **HackTheBox** | HTB | Real-world penetration testing |

### Certifications Aligned with DevSecOps

| Cert | Issuer | Focus |
|------|--------|-------|
| **CEH (Certified Ethical Hacker)** | EC-Council | Penetration Testing |
| **OSCP** | Offensive Security | Hands-on Hacking |
| **CKAD** | Linux Foundation | Kubernetes Security |
| **AWS Security** | Amazon | Cloud Security |

---

## Troubleshooting

### Common Issues

**Issue**: "SonarQube connection refused"
**Solution**: Check SONAR_HOST_URL and firewall rules

**Issue**: "Snyk authentication failed"
**Solution**: Verify SNYK_TOKEN in GitHub Secrets

**Issue**: "ZAP timeout during scan"
**Solution**: Increase timeout or reduce scope

**Issue**: "CVSS score mismatch"
**Solution**: Use https://www.first.org/cvss/calculator to verify

---

**Last Updated**: 2025-11-19
**Maintained By**: MoAI Security Team
**License**: Creative Commons Attribution 4.0
