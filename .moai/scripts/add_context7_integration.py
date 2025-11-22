#!/usr/bin/env python3
"""
Add Context7 Integration sections to security skills.
Part of SPEC-04-GROUP-E TDD implementation (GREEN phase).
"""

import re
from pathlib import Path

SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

CONTEXT7_TEMPLATES = {
    "moai-security-auth": """## Context7 Integration

### Related Libraries & Tools
- [bcryptjs](/dcodeIO/bcrypt.js): Bcrypt password hashing
- [argon2](/P-H-C/phc-winner-argon2): Argon2 password hashing
- [jwt](/auth0/node-jsonwebtoken): JWT authentication
- [oauth2-server](/oauthjs/node-oauth2-server): OAuth 2.0 authorization server
- [passport](/jaredhanson/passport): Authentication middleware

### Official Documentation
- [NIST SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html)
- [JWT.io](https://jwt.io/)
- [Passport.js](http://www.passportjs.org/)

### Version-Specific Guides
Latest stable versions: bcryptjs, argon2, JWT, OAuth 2.0
- [Password Storage Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [MFA Implementation Guide](https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html)
- [Session Management Guide](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)""",

    "moai-security-encryption": """## Context7 Integration

### Related Libraries & Tools
- [OpenSSL](/openssl/openssl): Cryptographic library
- [libsodium](/jedisct1/libsodium): Modern cryptography library
- [crypto](/nodejs/node): Node.js built-in crypto module
- [cryptography](/pyca/cryptography): Python cryptography library
- [TweetNaCl.js](/dchest/tweetnacl-js): Cryptographic library for JavaScript

### Official Documentation
- [NIST SP 800-38D](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf)
- [OpenSSL Documentation](https://www.openssl.org/docs/)
- [libsodium Documentation](https://doc.libsodium.org/)

### Version-Specific Guides
Latest stable versions: OpenSSL 3.x, libsodium, TLS 1.3
- [AES-GCM Implementation](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [Elliptic Curve Cryptography](https://safecurves.cr.yp.to/)
- [Key Derivation Functions](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Key_Management_Cheat_Sheet.html)""",

    "moai-security-ssrf": """## Context7 Integration

### Related Libraries & Tools
- [requestjs](/request/request): HTTP client for JavaScript
- [axios](/axios/axios): Promise-based HTTP client
- [requests](/psf/requests): HTTP library for Python
- [httpx](/encode/httpx): Modern HTTP client for Python
- [urllib3](/urllib3/urllib3): HTTP client for Python

### Official Documentation
- [OWASP SSRF](https://owasp.org/www-community/attacks/Server_Side_Request_Forgery)
- [CWE-918](https://cwe.mitre.org/data/definitions/918.html)
- [Blind SSRF](https://portswigger.net/web-security/ssrf)

### Version-Specific Guides
Latest stable versions: axios, requests, httpx
- [SSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [URL Validation](https://owasp.org/www-community/attacks/Server_Side_Request_Forgery)
- [Network Segmentation](https://cheatsheetseries.owasp.org/cheatsheets/Secure_Coding_Practices_Checklist.html)""",

    "moai-security-threat": """## Context7 Integration

### Related Libraries & Tools
- [threat-modeling](/OWASP/threat-modeling): OWASP threat modeling resources
- [attack-surface-analyzer](/microsoft/AttackSurfaceAnalyzer): Microsoft attack surface analyzer
- [draw.io](/jgraph/drawio): Diagram tool for threat models
- [ThreatDragon](/OWASP/threat-dragon): OWASP threat modeling tool
- [pytm](/nicodemos/pytm): Python threat modeling library

### Official Documentation
- [OWASP Threat Modeling](https://owasp.org/www-community/Threat_Modeling)
- [Threat Modeling Manifesto](https://www.threatmodelingmanifesto.org/)
- [STRIDE](https://en.wikipedia.org/wiki/STRIDE_(security))

### Version-Specific Guides
Latest stable versions: ThreatDragon, pytm, draw.io
- [Threat Modeling Process](https://owasp.org/www-community/Threat_Modeling)
- [STRIDE Methodology](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html)
- [Risk Assessment](https://cheatsheetseries.owasp.org/cheatsheets/Risk_Rating_Cheat_Sheet.html)""",
}

def add_context7_section(skill_name: str):
    """Add Context7 Integration section to skill."""
    skill_path = SKILLS_BASE_PATH / skill_name
    skill_file = skill_path / "SKILL.md"

    if not skill_file.exists():
        print(f"  {skill_name}: File not found")
        return False

    content = skill_file.read_text()

    # Skip if already has Context7 section
    if "## Context7 Integration" in content:
        print(f"  {skill_name}: Already has Context7 section, skipping")
        return False

    # Find insertion point (before ## Version History)
    insertion_pattern = r"(---\n\n)## Version History"
    if not re.search(insertion_pattern, content):
        print(f"  {skill_name}: Could not find insertion point")
        return False

    # Get the Context7 template
    if skill_name not in CONTEXT7_TEMPLATES:
        print(f"  {skill_name}: No template available")
        return False

    context7_section = CONTEXT7_TEMPLATES[skill_name]

    # Insert the Context7 section
    new_content = re.sub(
        insertion_pattern,
        f"\\1{context7_section}\n\n---\n\n## Version History",
        content
    )

    # Write back
    skill_file.write_text(new_content)
    print(f"  {skill_name}: Context7 Integration section added")
    return True

def main():
    """Main function."""
    skills_to_update = [
        "moai-security-auth",
        "moai-security-encryption",
        "moai-security-ssrf",
        "moai-security-threat",
    ]

    print("Adding Context7 Integration sections...\n")

    count = 0
    for skill in skills_to_update:
        if add_context7_section(skill):
            count += 1

    print(f"\nContext7 sections added to {count} skills")

if __name__ == "__main__":
    main()
