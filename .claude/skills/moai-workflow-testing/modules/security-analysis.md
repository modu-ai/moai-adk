# Security Analysis with Documentation

> Module: Documentation-enhanced security pattern detection and vulnerability scanning
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: WebSearch/WebFetch, regex engine, source parser, security scanner

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

Documentation Integration:
- OWASP Top 10 patterns
- Semgrep security rules
- Real-time vulnerability database
- Industry best practices
- Compliance frameworks

### Core Implementation

The analyzer loads detection patterns from Documentation (or a built-in default set) and runs them over source text as regexes. The patterns themselves are language-sensitive and should be loaded per-language; the orchestration logic below is language-neutral.

```text
class SecurityAnalyzer:
    docs
    security_patterns = {}

    load_security_patterns():
        if docs is none:
            return default_security_patterns()
        try:
            owasp = docs.get_library_docs(
                "<security/owasp>",
                topic="OWASP Top 10 vulnerability patterns",
                tokens=5000)
            semgrep = docs.get_library_docs(
                "<security/semgrep>",
                topic="security vulnerability detection patterns",
                tokens=4000)
            return { owasp, semgrep }
        except e:
            log("Failed to load Documentation security patterns: " + e)
            return default_security_patterns()
```

---

## Implementation Guide

### SQL Injection Detection

```text
analyze_sql_injection(file_path, content):
    issues = []
    # Patterns are language-sensitive; the ones below are illustrative.
    # Map them to the host language's query-execution idioms.
    patterns = [
        "string concatenation inside a query-execute call",   # query + user input
        "string interpolation/formatting inside execute()",   # f-strings, %, format()
        "execute() with % formatting",
        "raw SQL built from concatenated user input"
    ]
    for (line_num, line) in enumerate(lines(content), from=1):
        for pattern in patterns:
            if matches(line, pattern):
                issues.append(CodeIssue(
                    id="sql_injection_" + line_num,
                    category=TrustCategory.SAFETY, severity="critical",
                    issue_type="security_vulnerability", title="SQL Injection Risk",
                    description="Potential SQL injection vulnerability detected",
                    file_path=file_path, line_number=line_num, column_number=1,
                    code_snippet=trim(line),
                    suggested_fix="Use parameterized queries or an ORM to prevent SQL injection",
                    confidence=0.8, rule_violated="SQL_INJECTION",
                    external_reference="OWASP SQL Injection Prevention Cheat Sheet"))
    return issues
```

SQL Injection Best Practices:
- Use parameterized queries
- Implement ORM frameworks
- Validate and sanitize user input
- Apply principle of least privilege
- Use stored procedures when appropriate

### Command Injection Detection

```text
analyze_command_injection(file_path, content):
    issues = []
    # Dangerous sinks: shell execution from user-controlled input.
    # Names differ per language (Python os.system/subprocess shell=True,
    # Node child_process exec, Ruby backticks/Kernel.system,
    # Go os/exec with sh -c, PHP shell_exec). Map per language.
    dangerous_sinks = [shell_exec_call, system_call, eval_call, exec_call]
    for (line_num, line) in enumerate(lines(content), from=1):
        for sink in dangerous_sinks:
            if matches(line, sink):
                if uses_shell_string(line) or concatenates_user_input(line):
                    issues.append(CodeIssue(
                        id="command_injection_" + line_num,
                        category=TrustCategory.SAFETY, severity="critical",
                        issue_type="security_vulnerability", title="Command Injection Risk",
                        description="Potential command injection vulnerability",
                        file_path=file_path, line_number=line_num, column_number=1,
                        code_snippet=trim(line),
                        suggested_fix="Pass arguments as a list (not a shell string) and validate input",
                        confidence=0.9, rule_violated="COMMAND_INJECTION",
                        external_reference="OWASP Command Injection Prevention"))
    return issues
```

### Path Traversal Detection

```text
analyze_path_traversal(file_path, content):
    issues = []
    patterns = [
        "string concatenation inside a file-open call",
        "../ parent-directory reference",
        "..\\ windows parent-directory reference",
        "string formatting building a file path from user input"
    ]
    for (line_num, line) in enumerate(lines(content), from=1):
        for pattern in patterns:
            if matches(line, pattern):
                issues.append(CodeIssue(
                    id="path_traversal_" + line_num,
                    category=TrustCategory.SAFETY, severity="high",
                    issue_type="security_vulnerability", title="Path Traversal Risk",
                    description="Potential path traversal vulnerability",
                    file_path=file_path, line_number=line_num, column_number=1,
                    code_snippet=trim(line),
                    suggested_fix="Validate and sanitize file paths; constrain to an allowed root",
                    confidence=0.7, rule_violated="PATH_TRAVERSAL",
                    external_reference="OWASP Path Traversal Prevention"))
    return issues
```

### Hardcoded Credentials Detection

```text
analyze_hardcoded_credentials(file_path, content):
    issues = []
    # Match common credential variable names bound to a string literal of
    # suspicious length. The variable names are language-neutral heuristics.
    patterns = [
        "password = <quoted string, 8+ chars>",
        "api_key  = <quoted string, 20+ chars>",
        "secret   = <quoted string, 16+ chars>",
        "token    = <quoted string, 20+ chars>",
        "aws_access_key reference",
        "private_key assignment"
    ]
    for (line_num, line) in enumerate(lines(content), from=1):
        for pattern in patterns:
            if matches(line, pattern, case_insensitive):
                issues.append(CodeIssue(
                    id="hardcoded_credential_" + line_num,
                    category=TrustCategory.SAFETY, severity="critical",
                    issue_type="security_vulnerability", title="Hardcoded Credential",
                    description="Hardcoded credential detected in source code",
                    file_path=file_path, line_number=line_num, column_number=1,
                    code_snippet=trim(line)[:50] + "...",   # truncate for security
                    suggested_fix="Move credentials to environment variables or a secrets manager",
                    confidence=0.9, rule_violated="HARDCODED_CREDENTIALS",
                    external_reference="OWASP Key Management Cheat Sheet"))
    return issues
```

### Weak Cryptography Detection

```text
analyze_weak_cryptography(file_path, content):
    issues = []
    # Map each weak algorithm name to the host language's crypto API.
    # (Python hashlib.md5, Node crypto.createHash('md5'), Java MessageDigest "MD5",
    #  Go crypto/md5, etc.)
    weak = {
        md5: "md5 digest call",
        sha1:"sha1 digest call",
        des: "DES cipher selection",
        rc4: "RC4 cipher selection"
    }
    for (line_num, line) in enumerate(lines(content), from=1):
        for (algo, pattern) in weak:
            if matches(line, pattern):
                issues.append(CodeIssue(
                    id="weak_crypto_" + algo + "_" + line_num,
                    category=TrustCategory.SAFETY, severity="high",
                    issue_type="security_vulnerability", title="Weak Cryptography: " + uppercase(algo),
                    description="Use of weak cryptographic algorithm " + algo,
                    file_path=file_path, line_number=line_num, column_number=1,
                    code_snippet=trim(line),
                    suggested_fix="Replace " + algo + " with a stronger alternative (e.g. SHA-256, AES)",
                    confidence=0.9, rule_violated="WEAK_CRYPTOGRAPHY",
                    external_reference="OWASP Cryptographic Storage Cheat Sheet"))
    return issues
```

---

## Documentation-Enhanced Analysis

### Real-Time Vulnerability Database

```text
analyze_with_docs_patterns(file_path, content):
    issues = []
    security_patterns = load_security_patterns()
    if 'owasp'   in security_patterns: issues.extend(analyze_owasp_patterns  (file_path, content, security_patterns.owasp))
    if 'semgrep' in security_patterns: issues.extend(analyze_semgrep_rules   (file_path, content, security_patterns.semgrep))
    return issues
```

### Business Logic Vulnerabilities

```text
analyze_business_logic_security(file_path, content):
    issues = []
    tree = parse(content)     # host language's parser
    for node in walk(tree, matching=IfStatement):
        if is_weak_authentication(node):
            issues.append(CodeIssue(
                id="weak_auth_" + node.line,
                category=TrustCategory.SAFETY, severity="high",
                issue_type="security_vulnerability", title="Weak Authentication",
                description="Potential authentication bypass vulnerability",
                file_path=file_path, line_number=node.line, column_number=node.column,
                code_snippet=source_of(node, content),
                suggested_fix="Implement proper authentication with strong session management",
                confidence=0.7, rule_violated="WEAK_AUTHENTICATION"))
    return issues
```

---

## Security Fix Suggestions

```text
security_fix_suggestion(vulnerability_type):
    suggestions = {
        sql_injection:          "Use parameterized queries or an ORM to prevent SQL injection",
        command_injection:      "Pass arguments as a list (not a shell string) and validate input",
        path_traversal:         "Validate and sanitize file paths; constrain to an allowed root",
        hardcoded_credentials:  "Move credentials to environment variables or a secrets manager",
        weak_cryptography:      "Replace with a stronger algorithm (e.g. SHA-256, AES)",
        xss:                    "Sanitize user input and use context-aware output encoding",
        csrf:                   "Implement CSRF tokens with unique, unpredictable values",
        authentication_bypass:  "Implement proper authentication with multi-factor support"
    }
    return suggestions[vulnerability_type] default "Review and fix the security vulnerability"
```

---

## Best Practices

1. Documentation Integration: Leverage real-time vulnerability databases for latest threats
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
- [static-analysis.md](./static-analysis.md): security scanner integration for vulnerability scanning
- [automated-code-review/docs-integration.md](./automated-code-review/docs-integration.md): WebSearch/WebFetch patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/security-analysis.md`
