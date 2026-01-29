# AST-Grep Security Rules

Security vulnerability detection patterns for common attack vectors.

## SQL Injection

### Python

```yaml
id: sql-injection-python-format
language: python
severity: error
rule:
  any:
    - pattern: 'cursor.execute($QUERY % $ARGS)'
    - pattern: 'cursor.execute($QUERY.format($$$ARGS))'
    - pattern: 'cursor.execute(f"$$$SQL")'
    - pattern: 'execute($QUERY + $ARGS)'
message: 'Potential SQL injection. Use parameterized queries.'
fix: 'cursor.execute($QUERY, ($ARGS,))'
```

### JavaScript/TypeScript

```yaml
id: sql-injection-js
language: javascript
severity: error
rule:
  any:
    - pattern: '$DB.query(`$$$SQL ${$VAR} $$$REST`)'
    - pattern: '$DB.query($SQL + $VAR)'
    - pattern: '$DB.raw($SQL + $VAR)'
message: 'Potential SQL injection. Use parameterized queries.'
```

## XSS (Cross-Site Scripting)

### React

```yaml
id: xss-dangerouslySetInnerHTML
language: typescriptreact
severity: warning
rule:
  pattern: 'dangerouslySetInnerHTML={{ __html: $CONTENT }}'
  not:
    has:
      pattern: 'DOMPurify.sanitize($CONTENT)'
message: 'XSS risk: sanitize content before using dangerouslySetInnerHTML'
```

### JavaScript DOM

```yaml
id: xss-innerHTML
language: javascript
severity: warning
rule:
  any:
    - pattern: '$EL.innerHTML = $CONTENT'
    - pattern: 'document.write($CONTENT)'
message: 'Potential XSS vulnerability. Sanitize user input.'
```

## Secrets Detection

### Hardcoded Credentials

```yaml
id: hardcoded-password
language: python
severity: error
rule:
  any:
    - pattern: 'password = "$$$VALUE"'
    - pattern: 'PASSWORD = "$$$VALUE"'
    - pattern: 'secret = "$$$VALUE"'
    - pattern: 'api_key = "$$$VALUE"'
constraints:
  VALUE:
    regex: '.{8,}'  # At least 8 characters
message: 'Hardcoded credential detected. Use environment variables.'
```

### API Keys

```yaml
id: exposed-api-key
language: javascript
severity: error
rule:
  any:
    - pattern: 'apiKey: "$$$KEY"'
    - pattern: 'API_KEY = "$$$KEY"'
    - pattern: 'Authorization: "Bearer $$$TOKEN"'
message: 'API key should not be hardcoded. Use environment variables.'
```

## Command Injection

### Python

```yaml
id: command-injection-python
language: python
severity: error
rule:
  any:
    - pattern: 'os.system($CMD)'
    - pattern: 'subprocess.call($CMD, shell=True)'
    - pattern: 'subprocess.run($CMD, shell=True)'
    - pattern: 'os.popen($CMD)'
message: 'Potential command injection. Avoid shell=True and use subprocess with list arguments.'
fix: 'subprocess.run(shlex.split($CMD), shell=False)'
```

### JavaScript

```yaml
id: command-injection-js
language: javascript
severity: error
rule:
  any:
    - pattern: 'exec($CMD)'
    - pattern: 'execSync($CMD)'
    - pattern: 'spawn($CMD, { shell: true })'
message: 'Potential command injection. Use spawn with shell: false.'
```

## Path Traversal

```yaml
id: path-traversal
language: python
severity: error
rule:
  any:
    - pattern: 'open($PATH + $USER_INPUT)'
    - pattern: 'open(f"$$$PATH{$USER_INPUT}$$$REST")'
    - pattern: 'os.path.join($BASE, $USER_INPUT)'
  not:
    precedes:
      any:
        - pattern: 'os.path.realpath($$$ARGS)'
        - pattern: 'os.path.abspath($$$ARGS)'
message: 'Path traversal risk. Validate and sanitize file paths.'
```

## Insecure Cryptography

### Weak Hashing

```yaml
id: weak-hash-algorithm
language: python
severity: warning
rule:
  any:
    - pattern: 'hashlib.md5($$$ARGS)'
    - pattern: 'hashlib.sha1($$$ARGS)'
message: 'Weak hash algorithm. Use SHA-256 or stronger.'
fix: 'hashlib.sha256($$$ARGS)'
```

### Insecure Random

```yaml
id: insecure-random
language: python
severity: warning
rule:
  pattern: 'random.random()'
  inside:
    any:
      - pattern: 'def $FUNC($$$ARGS): $$$BODY'
        constraints:
          FUNC:
            regex: '.*(token|secret|key|password).*'
message: 'Use secrets module for security-sensitive random values.'
fix: 'secrets.token_hex(16)'
```

## CSRF Protection

```yaml
id: missing-csrf-token
language: html
severity: warning
rule:
  pattern: '<form $$$ATTRS>'
  not:
    has:
      pattern: 'csrf_token'
message: 'Form may be missing CSRF token.'
```

## Authentication Issues

### Hardcoded JWT Secret

```yaml
id: hardcoded-jwt-secret
language: javascript
severity: error
rule:
  pattern: 'jwt.sign($PAYLOAD, "$SECRET")'
message: 'JWT secret should not be hardcoded.'
```

### Missing Token Verification

```yaml
id: jwt-no-verification
language: javascript
severity: error
rule:
  pattern: 'jwt.decode($TOKEN)'
  not:
    inside:
      has:
        pattern: 'jwt.verify($TOKEN, $SECRET)'
message: 'Use jwt.verify() instead of jwt.decode() for security.'
```

## Usage

Run security scan:

```bash
sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml --severity error
```

JSON output for CI:

```bash
sg scan --config sgconfig.yml --json | jq '.[] | select(.severity == "error")'
```
