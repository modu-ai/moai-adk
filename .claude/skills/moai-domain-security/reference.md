# Application Security Reference

> OWASP Top 10, SAST, dependency scanning, and security best practices

---

## Official Documentation Links

| Tool/Standard | Documentation | Status |
|---------------|--------------|--------|
| **OWASP Top 10** | https://owasp.org/www-project-top-ten/ | ✅ Current (2025) |
| **SonarQube** | https://docs.sonarsource.com/sonarqube/latest/ | ✅ Current (2025) |
| **Snyk** | https://docs.snyk.io/ | ✅ Current (2025) |
| **Trivy** | https://aquasecurity.github.io/trivy/ | ✅ Current (2025) |

---

## OWASP Top 10 (2025)

1. **Broken Access Control**: Enforce least privilege
2. **Cryptographic Failures**: Use strong encryption
3. **Injection**: Use parameterized queries
4. **Insecure Design**: Threat modeling
5. **Security Misconfiguration**: Harden defaults
6. **Vulnerable Components**: Scan dependencies
7. **Authentication Failures**: MFA, strong passwords
8. **Data Integrity Failures**: Verify data integrity
9. **Logging Failures**: Comprehensive logging
10. **SSRF**: Validate URLs, use allowlists

---

## SAST (Static Analysis)

### SonarQube
```bash
# Run analysis
sonar-scanner   -Dsonar.projectKey=myapp   -Dsonar.sources=src   -Dsonar.host.url=http://localhost:9000   -Dsonar.token=mytoken
```

---

## Dependency Scanning

### Snyk
```bash
# Scan for vulnerabilities
snyk test

# Fix vulnerabilities
snyk fix
```

---

**Last Updated**: 2025-10-22
**Standards**: OWASP Top 10 2025, CWE Top 25
