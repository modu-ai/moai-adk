# Security Analysis with Context7

> Module: Context7-enhanced security pattern detection and vulnerability scanning
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, Context7 MCP, re, ast, bandit

## Quick Reference

### Security Vulnerability Categories

Injection Attacks:
- SQL Injection: Parameterized query validation
- Command Injection: Shell command safety checks
- LDAP Injection: Directory service query safety
- XPath Injection: XML query validation
- NoSQL Injection: NoSQL query safety

Authentication & Authorization:
- Hardcoded credentials detection
- Weak password validation
- Session management issues
- Authorization bypass detection
- Multi-factor authentication gaps

Data Protection:
- Sensitive data exposure
- Cryptographic storage issues
- Insufficient encryption
- Key management problems
- Data leakage detection

API Security:
- Improper input validation
- Authentication token handling
- Rate limiting issues
- CORS misconfiguration
- API version management

Context7 Integration:
- OWASP Top 10 patterns
- Semgrep security rules
- Real-time vulnerability database
- Industry best practices
- Compliance frameworks

### Core Implementation

```python
import re
from typing import Dict, List, Any

class SecurityAnalyzer:
    """Security vulnerability analyzer with Context7 integration."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.security_patterns = {}

    async def load_security_patterns(self) -> Dict[str, Any]:
        """Load security patterns from Context7."""
        if not self.context7:
            return self._get_default_security_patterns()

        try:
            # Load OWASP Top 10 patterns
            owasp_patterns = await self.context7.get_library_docs(
                context7_library_id="/security/owasp",
                topic="OWASP Top 10 vulnerability patterns 2025",
                tokens=5000
            )

            # Load Semgrep security rules
            semgrep_patterns = await self.context7.get_library_docs(
                context7_library_id="/security/semgrep",
                topic="security vulnerability detection patterns",
                tokens=4000
            )

            return {
                'owasp': owasp_patterns,
                'semgrep': semgrep_patterns
            }

        except Exception as e:
            print(f"Failed to load Context7 security patterns: {e}")
            return self._get_default_security_patterns()
```

---

## Implementation Guide

### SQL Injection Detection

```python
async def analyze_sql_injection(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect SQL injection vulnerabilities."""

    issues = []
    lines = content.split('\n')

    # SQL injection patterns
    sql_injection_patterns = [
        r"execute\([^)]*\+[^)]*\)",  # String concatenation
        r"format\s*\(",               # String formatting
        r"%\s*[^,]*s",               # Old-style formatting
        r"\.execute\(.*\%.*\)",      # Execute with formatting
        r"\.exec\(.*\+.*\)",         # Exec with concatenation
    ]

    for line_num, line in enumerate(lines, 1):
        for pattern in sql_injection_patterns:
            if re.search(pattern, line, re.IGNORECASE):
                issue = CodeIssue(
                    id=f"sql_injection_{line_num}",
                    category=TrustCategory.SAFETY,
                    severity="critical",
                    issue_type="security_vulnerability",
                    title="SQL Injection Risk",
                    description="Potential SQL injection vulnerability detected",
                    file_path=file_path,
                    line_number=line_num,
                    column_number=1,
                    code_snippet=line.strip(),
                    suggested_fix="Use parameterized queries or ORM to prevent SQL injection",
                    confidence=0.8,
                    rule_violated="SQL_INJECTION",
                    external_reference="OWASP SQL Injection Prevention Cheat Sheet"
                )
                issues.append(issue)

    return issues
```

SQL Injection Best Practices:
- Use parameterized queries
- Implement ORM frameworks
- Validate and sanitize user input
- Apply principle of least privilege
- Use stored procedures when appropriate

### Command Injection Detection

```python
async def analyze_command_injection(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect command injection vulnerabilities."""

    issues = []
    lines = content.split('\n')

    # Command injection patterns
    command_injection_patterns = [
        r"os\.system\(",
        r"subprocess\.call\(",
        r"subprocess\.Popen\(",
        r"eval\(",
        r"exec\(",
        r"__import__\(.*os\.system",
    ]

    for line_num, line in enumerate(lines, 1):
        for pattern in command_injection_patterns:
            if re.search(pattern, line):
                # Check if using shell=True or user input
                if 'shell=True' in line or '+' in line or '%' in line:
                    issue = CodeIssue(
                        id=f"command_injection_{line_num}",
                        category=TrustCategory.SAFETY,
                        severity="critical",
                        issue_type="security_vulnerability",
                        title="Command Injection Risk",
                        description="Potential command injection vulnerability",
                        file_path=file_path,
                        line_number=line_num,
                        column_number=1,
                        code_snippet=line.strip(),
                        suggested_fix="Use subprocess.run with proper argument lists or validate input",
                        confidence=0.9,
                        rule_violated="COMMAND_INJECTION",
                        external_reference="OWASP Command Injection Prevention"
                    )
                    issues.append(issue)

    return issues
```

### Path Traversal Detection

```python
async def analyze_path_traversal(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect path traversal vulnerabilities."""

    issues = []
    lines = content.split('\n')

    # Path traversal patterns
    path_traversal_patterns = [
        r"open\([^)]*\+[^)]*\)",  # String concatenation in open
        r"\.\.\/",                 # Parent directory reference
        r"\.\.\\",                 # Windows parent directory
        r"format\(.*\%.*\)",       # String formatting in file path
    ]

    for line_num, line in enumerate(lines, 1):
        for pattern in path_traversal_patterns:
            if re.search(pattern, line):
                issue = CodeIssue(
                    id=f"path_traversal_{line_num}",
                    category=TrustCategory.SAFETY,
                    severity="high",
                    issue_type="security_vulnerability",
                    title="Path Traversal Risk",
                    description="Potential path traversal vulnerability",
                    file_path=file_path,
                    line_number=line_num,
                    column_number=1,
                    code_snippet=line.strip(),
                    suggested_fix="Validate and sanitize file paths, use absolute paths",
                    confidence=0.7,
                    rule_violated="PATH_TRAVERSAL",
                    external_reference="OWASP Path Traversal Prevention"
                )
                issues.append(issue)

    return issues
```

### Hardcoded Credentials Detection

```python
async def analyze_hardcoded_credentials(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect hardcoded credentials."""

    issues = []
    lines = content.split('\n')

    # Credential patterns
    credential_patterns = [
        r"password\s*=\s*['\"][^'\"]{8,}['\"]",  # Hardcoded password
        r"api_key\s*=\s*['\"][^'\"]{20,}['\"]",  # Hardcoded API key
        r"secret\s*=\s*['\"][^'\"]{16,}['\"]",   # Hardcoded secret
        r"token\s*=\s*['\"][^'\"]{20,}['\"]",    # Hardcoded token
        r"aws_access_key",                        # AWS credentials
        r"private_key\s*=",                       # Private key
    ]

    for line_num, line in enumerate(lines, 1):
        for pattern in credential_patterns:
            if re.search(pattern, line, re.IGNORECASE):
                issue = CodeIssue(
                    id=f"hardcoded_credential_{line_num}",
                    category=TrustCategory.SAFETY,
                    severity="critical",
                    issue_type="security_vulnerability",
                    title="Hardcoded Credential",
                    description="Hardcoded credential detected in source code",
                    file_path=file_path,
                    line_number=line_num,
                    column_number=1,
                    code_snippet=line.strip()[:50] + "...",  # Truncate for security
                    suggested_fix="Move credentials to environment variables or secure configuration",
                    confidence=0.9,
                    rule_violated="HARDCODED_CREDENTIALS",
                    external_reference="OWASP Key Management Cheat Sheet"
                )
                issues.append(issue)

    return issues
```

### Weak Cryptography Detection

```python
async def analyze_weak_cryptography(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect weak cryptographic practices."""

    issues = []
    lines = content.split('\n')

    # Weak cryptography patterns
    weak_crypto_patterns = {
        'md5': r"hashlib\.md5\(",
        'sha1': r"hashlib\.sha1\(",
        'des': r"Cipher\.algo\s*=\s*['\"]DES['\"]",
        'rc4': r"Cipher\.algo\s*=\s*['\"]RC4['\"]",
    }

    for line_num, line in enumerate(lines, 1):
        for crypto_type, pattern in weak_crypto_patterns.items():
            if re.search(pattern, line):
                issue = CodeIssue(
                    id=f"weak_crypto_{crypto_type}_{line_num}",
                    category=TrustCategory.SAFETY,
                    severity="high",
                    issue_type="security_vulnerability",
                    title=f"Weak Cryptography: {crypto_type.upper()}",
                    description=f"Use of weak cryptographic algorithm {crypto_type}",
                    file_path=file_path,
                    line_number=line_num,
                    column_number=1,
                    code_snippet=line.strip(),
                    suggested_fix=f"Replace {crypto_type} with stronger alternative (e.g., SHA-256, AES)",
                    confidence=0.9,
                    rule_violated="WEAK_CRYPTOGRAPHY",
                    external_reference="OWASP Cryptographic Storage Cheat Sheet"
                )
                issues.append(issue)

    return issues
```

---

## Context7-Enhanced Analysis

### Real-Time Vulnerability Database

```python
async def analyze_with_context7_patterns(
    self, file_path: str, content: str
) -> List[CodeIssue]:
    """Analyze code using Context7 security patterns."""

    issues = []

    # Load latest security patterns
    security_patterns = await self.load_security_patterns()

    # Analyze using OWASP patterns
    if 'owasp' in security_patterns:
        owasp_issues = await self._analyze_owasp_patterns(
            file_path, content, security_patterns['owasp']
        )
        issues.extend(owasp_issues)

    # Analyze using Semgrep rules
    if 'semgrep' in security_patterns:
        semgrep_issues = await self._analyze_semgrep_rules(
            file_path, content, security_patterns['semgrep']
        )
        issues.extend(semgrep_issues)

    return issues
```

### Business Logic Vulnerabilities

```python
async def analyze_business_logic_security(self, file_path: str, content: str) -> List[CodeIssue]:
    """Detect business logic security issues."""

    issues = []
    tree = ast.parse(content)

    # Check for authentication bypass patterns
    for node in ast.walk(tree):
        if isinstance(node, ast.If):
            # Check for weak authentication conditions
            if self._is_weak_authentication(node):
                issue = CodeIssue(
                    id=f"weak_auth_{node.lineno}",
                    category=TrustCategory.SAFETY,
                    severity="high",
                    issue_type="security_vulnerability",
                    title="Weak Authentication",
                    description="Potential authentication bypass vulnerability",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=node.col_offset,
                    code_snippet=self._get_node_source(node, content),
                    suggested_fix="Implement proper authentication with strong session management",
                    confidence=0.7,
                    rule_violated="WEAK_AUTHENTICATION"
                )
                issues.append(issue)

    return issues
```

---

## Security Fix Suggestions

```python
def get_security_fix_suggestion(self, vulnerability_type: str) -> str:
    """Get security fix suggestion."""

    suggestions = {
        'sql_injection': "Use parameterized queries or ORM to prevent SQL injection",
        'command_injection': "Use subprocess.run with proper argument lists or validate input",
        'path_traversal': "Validate and sanitize file paths, use absolute paths",
        'hardcoded_credentials': "Move credentials to environment variables or secure configuration",
        'weak_cryptography': "Replace with stronger cryptographic algorithms (e.g., SHA-256, AES)",
        'xss': "Sanitize user input and use context-aware output encoding",
        'csrf': "Implement CSRF tokens with unique, unpredictable values",
        'authentication_bypass': "Implement proper authentication with multi-factor support",
    }

    return suggestions.get(vulnerability_type, "Review and fix security vulnerability")
```

---

## Best Practices

1. Context7 Integration: Leverage real-time vulnerability databases for latest threats
2. Comprehensive Coverage: Check all OWASP Top 10 vulnerability categories
3. Severity Accuracy: Use confidence scores to prioritize fixes
4. Actionable Guidance: Provide specific fix suggestions with code examples
5. Reference Documentation: Link to OWASP and industry best practices
6. Regular Updates: Keep security patterns current with evolving threats
7. False Positive Reduction: Use multiple detection methods for accuracy
8. Team Training: Educate team on common security pitfalls

---

## Related Modules

- [TRUST 5 Validation](./trust5-validation.md): Safety category analysis
- [static-analysis.md](./static-analysis.md): bandit integration for security scanning
- [automated-code-review/context7-integration.md](./automated-code-review/context7-integration.md): Context7 MCP patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/security-analysis.md`
