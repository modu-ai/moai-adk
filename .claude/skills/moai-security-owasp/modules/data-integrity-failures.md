# Data Integrity Failures (A08)

## Overview

Data integrity failures involve insufficient verification of code and data integrity, including insecure deserialization, missing integrity checks, and CI/CD pipeline vulnerabilities.

## Critical Risks

### Insecure Deserialization
- Arbitrary code execution
- Remote code execution
- Object injection attacks

### Missing Integrity Checks
- Unsigned software updates
- Missing checksums
- Unverified downloads

## Remediation Patterns

### Safe Deserialization

**Avoid Unsafe Formats**:
```python
# DANGEROUS: pickle can execute arbitrary code
import pickle

def load_data(data):
    return pickle.loads(data)  # NEVER DO THIS

# SAFE: Use JSON instead
import json

def load_data(data):
    obj = json.loads(data)

    # Validate structure
    if not isinstance(obj, dict):
        raise ValueError("Invalid data format")

    # Validate required fields
    required = ['name', 'email', 'age']
    if not all(field in obj for field in required):
        raise ValueError("Missing required fields")

    # Validate types
    if not isinstance(obj['age'], int):
        raise ValueError("Age must be integer")

    return obj
```

**Safe Serialization Libraries**:
```javascript
// SAFE: Use JSON for data
const data = JSON.parse(userInput);

// SAFE: MessagePack (fast binary JSON)
const msgpack = require('msgpack');
const data = msgpack.decode(buffer);

// SAFE: Protocol Buffers (schema-based)
const protobuf = require('protobufjs');
const message = MyMessage.decode(buffer);
```

### Code Signing

**Sign Release Artifacts**:
```bash
# Sign with GPG
gpg --detach-sign --armor release.tar.gz

# Verify signature
gpg --verify release.tar.gz.asc release.tar.gz
```

**npm Package Signing**:
```bash
# Sign package
npm publish --sign

# Verify signature
npm audit signatures
```

**Docker Image Signing**:
```bash
# Enable Docker Content Trust
export DOCKER_CONTENT_TRUST=1

# Sign and push image
docker push myimage:latest

# Verify signature
docker trust inspect myimage:latest
```

### Integrity Verification

**Subresource Integrity (SRI)**:
```html
<!-- Verify external scripts -->
<script
  src="https://cdn.example.com/library.js"
  integrity="sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8wC"
  crossorigin="anonymous"></script>

<!-- Verify CSS -->
<link
  rel="stylesheet"
  href="https://cdn.example.com/style.css"
  integrity="sha384-50aSY3yN5J3PfBYvqJVkM8VCKqKKvjJ5pWCz7rvCfDn72rkQJ9r9kLn8YqkPDQn"
  crossorigin="anonymous">
```

**Generate SRI Hash**:
```bash
# Generate hash for file
cat library.js | openssl dgst -sha384 -binary | openssl base64 -A
```

### CI/CD Pipeline Security

**Secure GitHub Actions**:
```yaml
name: Secure Build

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest

    # Pin action versions
    steps:
      - uses: actions/checkout@v4.1.0  # Specific version

      # Verify dependencies
      - name: Verify Checksums
        run: |
          sha256sum -c checksums.txt

      # Sign artifacts
      - name: Sign Release
        run: |
          gpg --import ${{ secrets.GPG_PRIVATE_KEY }}
          gpg --detach-sign --armor dist/release.tar.gz

      # Upload with integrity
      - name: Upload Artifact
        uses: actions/upload-artifact@v3
        with:
          name: signed-release
          path: |
            dist/release.tar.gz
            dist/release.tar.gz.asc
```

**GitLab CI/CD Security**:
```yaml
build:
  stage: build
  script:
    # Verify dependencies
    - npm ci --ignore-scripts
    - npm audit --audit-level=high

    # Build with checksums
    - npm run build
    - sha256sum dist/* > checksums.txt

  artifacts:
    paths:
      - dist/
      - checksums.txt
    reports:
      dependency_scanning: gl-dependency-scanning-report.json
```

### Software Supply Chain Security

**Dependency Integrity**:
```javascript
// package-lock.json includes integrity hashes
{
  "name": "express",
  "version": "4.18.2",
  "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz",
  "integrity": "sha512-5/PsL6iGPdfQ/lKM1UuielYgv3BUoJfz1aUwU9vHZ+J7gyvwdQXFEBIEIaxeGf0GIcreATNyBExtalisDbuMqQ=="
}
```

**Verify Before Install**:
```bash
# npm verifies integrity automatically
npm ci

# Fail on integrity mismatch
npm install --strict-ssl --audit
```

### Update Verification

**Automated Update Verification**:
```python
import hashlib
import requests

def verify_update(url, expected_hash):
    """Verify update integrity before applying."""
    response = requests.get(url)
    actual_hash = hashlib.sha256(response.content).hexdigest()

    if actual_hash != expected_hash:
        raise ValueError(f"Integrity check failed: {actual_hash} != {expected_hash}")

    return response.content

# Usage
try:
    update = verify_update(
        'https://example.com/update.tar.gz',
        'a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3'
    )
    apply_update(update)
except ValueError as e:
    logger.error(f"Update verification failed: {e}")
```

## Best Practices

### Secure Development

**DO**:
- ✅ Use JSON instead of pickle/YAML
- ✅ Validate deserialized data structure
- ✅ Sign release artifacts
- ✅ Verify checksums
- ✅ Use SRI for external resources
- ✅ Pin CI/CD action versions
- ✅ Implement code signing

**DON'T**:
- ❌ Deserialize untrusted data
- ❌ Use pickle with user input
- ❌ Skip integrity verification
- ❌ Trust unsigned updates
- ❌ Use latest/unpinned versions in CI/CD

### Serialization Security

```python
# Schema validation with Pydantic
from pydantic import BaseModel, validator

class UserData(BaseModel):
    name: str
    email: str
    age: int

    @validator('age')
    def age_must_be_positive(cls, v):
        if v < 0:
            raise ValueError('Age must be positive')
        return v

# Safe deserialization
def load_user(data: str) -> UserData:
    return UserData.parse_raw(data)
```

## Validation Checklist

- [ ] No unsafe deserialization (pickle, YAML)
- [ ] JSON schema validation
- [ ] Code signing implemented
- [ ] Checksum verification
- [ ] SRI for external resources
- [ ] CI/CD pipeline secured
- [ ] Dependency integrity verified
- [ ] Update verification automated

## Testing

```python
def test_safe_deserialization():
    """Verify safe deserialization."""
    # Valid data
    valid = '{"name": "Alice", "email": "alice@example.com", "age": 30}'
    user = load_user(valid)
    assert user.name == "Alice"

    # Invalid structure
    invalid = '{"name": "Bob"}'
    with pytest.raises(ValidationError):
        load_user(invalid)

    # Type mismatch
    invalid_type = '{"name": "Bob", "email": "bob@example.com", "age": "thirty"}'
    with pytest.raises(ValidationError):
        load_user(invalid_type)
```

---

**Last Updated**: 2025-11-24
**OWASP Category**: A08:2021
**CWE**: CWE-502 (Deserialization of Untrusted Data)
