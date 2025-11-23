# Vulnerable and Outdated Components (A06)

## Overview

Using components with known vulnerabilities exposes applications to attacks. This includes outdated dependencies, unpatched libraries, and insecure configurations.

## Critical Risks

### Known CVEs
- Components with published vulnerabilities
- Unpatched security issues
- End-of-life software

### Supply Chain Attacks
- Compromised packages
- Typosquatting
- Dependency confusion

## Remediation Patterns

### Dependency Scanning

**npm audit**:
```bash
# Check for vulnerabilities
npm audit

# Fix vulnerabilities automatically
npm audit fix

# Fix with breaking changes
npm audit fix --force
```

**Python (pip-audit)**:
```bash
# Install pip-audit
pip install pip-audit

# Scan dependencies
pip-audit

# Generate SBOM
pip-audit --format cyclonedx-json --output sbom.json
```

**Snyk Integration**:
```bash
# Install Snyk CLI
npm install -g snyk

# Authenticate
snyk auth

# Test project
snyk test

# Monitor continuously
snyk monitor
```

### Automated CI/CD Scanning

**GitHub Actions**:
```yaml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run Snyk
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}

      - name: npm audit
        run: npm audit --audit-level=high

      - name: OWASP Dependency Check
        uses: dependency-check/Dependency-Check_Action@main
        with:
          project: 'my-project'
          path: '.'
          format: 'HTML'
```

### Dependency Version Pinning

**package.json (Exact Versions)**:
```json
{
  "dependencies": {
    "express": "4.18.2",
    "helmet": "7.0.0",
    "bcrypt": "5.1.1"
  }
}
```

**package-lock.json**:
```bash
# Commit package-lock.json for reproducible builds
git add package-lock.json
git commit -m "Lock dependency versions"
```

**Python (requirements.txt)**:
```
Django==4.2.7
cryptography==41.0.7
requests==2.31.0
```

### Update Strategy

**Dependabot Configuration**:
```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5

  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
```

**Renovate Configuration**:
```json
{
  "extends": ["config:base"],
  "schedule": ["every weekend"],
  "vulnerabilityAlerts": {
    "enabled": true,
    "labels": ["security"]
  },
  "packageRules": [
    {
      "matchUpdateTypes": ["major"],
      "automerge": false
    },
    {
      "matchUpdateTypes": ["minor", "patch"],
      "automerge": true
    }
  ]
}
```

### Software Bill of Materials (SBOM)

**Generate SBOM**:
```bash
# CycloneDX (npm)
npx @cyclonedx/cyclonedx-npm --output-file sbom.json

# SPDX (Python)
pip-audit --format spdx > sbom.spdx.json
```

**SBOM Validation**:
```python
import json

def validate_sbom(sbom_path):
    """Validate SBOM against known vulnerabilities."""
    with open(sbom_path) as f:
        sbom = json.load(f)

    vulnerabilities = []

    for component in sbom.get('components', []):
        name = component['name']
        version = component['version']

        # Check against vulnerability database
        vulns = check_vulnerabilities(name, version)
        if vulns:
            vulnerabilities.append({
                'component': name,
                'version': version,
                'vulnerabilities': vulns
            })

    return vulnerabilities
```

### Vendor Security Policies

**Verify Package Integrity**:
```bash
# npm package verification
npm install --ignore-scripts

# Verify checksums
shasum -a 256 package.tar.gz
```

**PyPI Package Verification**:
```bash
# Verify with pip
pip install --require-hashes -r requirements.txt

# Generate hashes
pip-compile --generate-hashes requirements.in
```

### Monitoring & Alerting

**Vulnerability Monitoring Service**:
```javascript
class VulnerabilityMonitor {
  async checkDependencies() {
    // Fetch latest vulnerability data
    const vulnData = await this.fetchVulnerabilities();

    // Check installed packages
    const packages = await this.getInstalledPackages();

    const findings = [];

    for (const pkg of packages) {
      const vulns = vulnData[pkg.name]?.[pkg.version];
      if (vulns?.length > 0) {
        findings.push({
          package: pkg.name,
          version: pkg.version,
          vulnerabilities: vulns
        });
      }
    }

    // Alert on critical vulnerabilities
    const critical = findings.filter(f =>
      f.vulnerabilities.some(v => v.severity === 'critical')
    );

    if (critical.length > 0) {
      await this.sendAlert(critical);
    }

    return findings;
  }
}
```

## Best Practices

### Dependency Management

**DO**:
- ✅ Pin dependency versions
- ✅ Scan for vulnerabilities regularly
- ✅ Update dependencies within 30 days
- ✅ Generate and maintain SBOM
- ✅ Monitor security advisories
- ✅ Remove unused dependencies

**DON'T**:
- ❌ Use wildcard versions in production
- ❌ Ignore vulnerability warnings
- ❌ Skip security updates
- ❌ Trust unverified packages
- ❌ Use deprecated packages

### Update Policy

```markdown
## Security Update SLA

**Critical Vulnerabilities** (CVSS ≥9.0)
- Patch within: 24 hours
- Testing: Automated + Manual
- Deployment: Immediate

**High Vulnerabilities** (CVSS 7.0-8.9)
- Patch within: 7 days
- Testing: Automated
- Deployment: Next release

**Medium Vulnerabilities** (CVSS 4.0-6.9)
- Patch within: 30 days
- Testing: Automated
- Deployment: Scheduled

**Low Vulnerabilities** (CVSS <4.0)
- Patch within: 90 days
- Testing: Automated
- Deployment: Scheduled
```

## Validation Checklist

- [ ] Automated vulnerability scanning
- [ ] Dependencies up to date
- [ ] No known CVEs in dependencies
- [ ] Unused dependencies removed
- [ ] SBOM maintained
- [ ] Dependency checksums verified
- [ ] Security advisories monitored
- [ ] Update policy documented

## Testing

```javascript
describe('Dependency Security', () => {
  test('should have no critical vulnerabilities', async () => {
    const audit = await runNpmAudit();
    expect(audit.critical).toBe(0);
  });

  test('should have SBOM file', () => {
    expect(fs.existsSync('sbom.json')).toBe(true);
  });

  test('should have no outdated dependencies', async () => {
    const outdated = await getOutdatedPackages();
    const critical = outdated.filter(p => p.severity === 'critical');
    expect(critical).toHaveLength(0);
  });
});
```

---

**Last Updated**: 2025-11-24
**OWASP Category**: A06:2021
**CWE**: CWE-1035 (Vulnerable Third Party Component)
