## API Reference

### Core SSRF Protection Operations
- `validate_url(url)` - Comprehensive URL validation and risk assessment
- `make_secure_request(url, options)` - Secure HTTP request with validation
- `is_ip_allowed(ip_address)` - IP address validation against allowlist/denylist
- `validate_network_access(host, port)` - Network access validation
- `calculate_risk_score(url)` - Risk scoring for suspicious patterns

### Official Documentation & Resources
- [OWASP SSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [CWE-918 SSRF Vulnerability](https://cwe.mitre.org/data/definitions/918.html)
- [Blind SSRF - PortSwigger](https://portswigger.net/web-security/ssrf)
- [SSRF Attack Examples](https://owasp.org/www-community/attacks/Server_Side_Request_Forgery)

### Context7 Integration
- `get_latest_ssrf_docs()` - SSRF protection via Context7
- `analyze_threat_patterns()` - Threat pattern analysis via Context7
- `optimize_url_validation()` - URL validation optimization via Context7

## Best Practices (November 2025)

### DO
- Use allowlist approach for domains and IP addresses
- Implement comprehensive URL parsing and validation
- Block access to internal network ranges and metadata endpoints
- Validate request responses for size and content type
- Implement rate limiting and anomaly detection
- Log all requests and security events
- Use network segmentation for additional protection
- Regularly update threat intelligence and protection rules

### DON'T
- Rely solely on blacklist approaches for protection
 Allow user-controlled URLs without validation
- Skip DNS resolution and IP address validation
- Forget to implement response size limits
- Ignore suspicious patterns in URLs and parameters
- Skip logging and monitoring of security events
- Use outdated threat intelligence or protection rules
- Forget to test SSRF protection regularly

## Works Well With

- `moai-security-api` (API security implementation)
- `moai-foundation-trust` (Security and compliance)
- `moai-security-compliance` (Compliance management)
- `moai-domain-backend` (Backend security)
- `moai-cc-configuration` (Security configuration)
- `moai-baas-foundation` (BaaS security patterns)
- `moai-security-owasp` (OWASP security standards)
- `moai-security-encryption` (Encryption and data protection)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced threat detection, and comprehensive protection patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, SSRF protection patterns, validation systems
- **v1.0.0** (2025-11-11): Initial SSRF security foundation

---

**End of Skill** | Updated 2025-11-13

## SSRF Security Framework

### Protection Layers
- URL validation with comprehensive parsing
- Domain and IP address filtering
- Network segmentation and isolation
- Request rate limiting and anomaly detection
- Response validation and size limits

### Enterprise Features
- Real-time threat intelligence integration
- Comprehensive logging and audit trails
- Automated vulnerability assessment
- Integration with security information and event management (SIEM)
- Compliance reporting and documentation

---

**End of Enterprise SSRF Security Protection Expert **
