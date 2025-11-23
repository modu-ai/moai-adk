# Security Reference & Compliance

API reference, tools comparison, and compliance checklists.

## Security Headers Reference

| Header | Purpose | Example |
|--------|---------|---------|
| X-Content-Type-Options | Prevent MIME type sniffing | nosniff |
| X-Frame-Options | Prevent clickjacking | DENY |
| X-XSS-Protection | Enable XSS filter | 1; mode=block |
| Strict-Transport-Security | Force HTTPS | max-age=31536000 |
| Content-Security-Policy | XSS and data injection | default-src 'self' |
| Referrer-Policy | Control referrer info | no-referrer |

## Compliance Frameworks

### SOC 2 Type II Checklist

- [ ] Access controls (who can access what)
- [ ] Change management (controlled deployments)
- [ ] Monitoring & alerting (24/7 monitoring)
- [ ] Incident response (documented procedures)
- [ ] Audit logging (comprehensive logs)
- [ ] Encryption (data at rest & transit)
- [ ] Business continuity (disaster recovery)
- [ ] Risk assessment (regular reviews)

### ISO 27001 Checklist

- [ ] Information security policy
- [ ] Asset management
- [ ] Access control
- [ ] Cryptography
- [ ] Physical & environmental controls
- [ ] Operations security
- [ ] Communications security
- [ ] System acquisition & development
- [ ] Supplier relationships
- [ ] Information security incident management
- [ ] Business continuity management
- [ ] Compliance

### GDPR Compliance Checklist

- [ ] Data processing agreement (DPA)
- [ ] Privacy by design
- [ ] Data minimization
- [ ] Consent management
- [ ] Data subject rights (access, deletion)
- [ ] Data breach notification (72 hours)
- [ ] Privacy impact assessment (PIA)
- [ ] Data retention policy
- [ ] Third-party vendor reviews
- [ ] Staff training

## Vulnerability Severity Ratings

| CVSS Score | Severity | Impact |
|-----------|----------|--------|
| 0.0 | None | No impact |
| 0.1-3.9 | Low | Limited impact |
| 4.0-6.9 | Medium | Moderate impact |
| 7.0-8.9 | High | Serious impact |
| 9.0-10.0 | Critical | Severe impact |

## Tools Comparison

### SAST Tools

| Tool | Language | Cost | Accuracy |
|------|----------|------|----------|
| Bandit | Python | Free | Medium |
| SonarQube | Multi | Free/Paid | High |
| Checkmarx | Multi | Paid | Very High |
| Fortify | Multi | Paid | Very High |

### Dependency Scanners

| Tool | Method | Cost | Speed |
|------|--------|------|-------|
| Safety | Database | Free | Fast |
| npm audit | Database | Free | Fast |
| Snyk | ML | Free/Paid | Fast |
| Black Duck | AI | Paid | Slow |

## Incident Response Playbook

### Step 1: Detect
- [ ] Receive alert
- [ ] Confirm incident
- [ ] Estimate scope

### Step 2: Analyze
- [ ] Identify affected systems
- [ ] Determine entry point
- [ ] Assess damage

### Step 3: Contain
- [ ] Isolate affected systems
- [ ] Stop lateral movement
- [ ] Preserve evidence

### Step 4: Eradicate
- [ ] Remove malware
- [ ] Patch vulnerabilities
- [ ] Restore from backup

### Step 5: Recover
- [ ] Bring systems online
- [ ] Monitor for recurrence
- [ ] Restore user access

### Step 6: Document
- [ ] Write incident report
- [ ] Update runbooks
- [ ] Conduct post-mortem

## Security Testing Roadmap

### Phase 1: Baseline (Week 1-2)
- SAST scan
- Dependency audit
- Architecture review

### Phase 2: Deep Dive (Week 3-4)
- Threat modeling
- Penetration testing
- Code review

### Phase 3: Continuous (Ongoing)
- Automated scanning
- Vulnerability patching
- Security training

## Best Practices Checklist

### Development
- [ ] Input validation on all data
- [ ] Output encoding for XSS prevention
- [ ] Parameterized queries for SQL injection prevention
- [ ] CSRF tokens for state-changing operations
- [ ] Secure session management
- [ ] HTTPS/TLS for all data in transit
- [ ] Secrets in environment variables
- [ ] Security logging without sensitive data

### Deployment
- [ ] Security headers configured
- [ ] TLS 1.2+ only
- [ ] Automated dependency scanning
- [ ] Secrets management in place
- [ ] Monitoring & alerting enabled
- [ ] Incident response plan
- [ ] Regular backups
- [ ] Disaster recovery tested

### Operations
- [ ] Security patches applied promptly
- [ ] Access logs reviewed
- [ ] Unusual activity monitored
- [ ] Credentials rotated regularly
- [ ] Third-party risks assessed
- [ ] Compliance audits scheduled
- [ ] Security training conducted
- [ ] Incident response drills executed

---

**Last Updated**: 2025-11-23
