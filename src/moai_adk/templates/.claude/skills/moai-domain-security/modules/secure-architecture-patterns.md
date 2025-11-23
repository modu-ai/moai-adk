# Secure Architecture & Cryptography

DevSecOps integration and production-grade encryption patterns.

                    findings.append({
                        "type": "EXPOSED_SECRET",
                        "file": secret.get("file"),
                        "detector": secret.get("DetectorName"),
                        "severity": "CRITICAL"
                    })

        return {
            "tool": "Secret Scanning (TruffleHog)",
            "findings": findings,
            "total_issues": len(findings),
            "status": "FAILED" if findings else "PASSED"
        }

    def _run_dast(self, commit_sha: str) -> Dict[str, Any]:
        """
        Run Dynamic Application Security Testing.

        Tools: OWASP ZAP, Burp Suite
        """
        # DAST requires deployed application
        if not self.config.dast_target_url:
            return {"status": "SKIPPED", "reason": "No target URL configured"}

        findings = []

        # Run OWASP ZAP baseline scan
        zap_cmd = f"zap-baseline.py -t {self.config.dast_target_url} -J /tmp/zap-{commit_sha}.json"
        subprocess.run(zap_cmd, shell=True, check=False)

        zap_results = self._parse_zap_results(f"/tmp/zap-{commit_sha}.json")
        findings.extend(zap_results)

        categorized = self._categorize_findings(findings)

        return {
            "tool": "DAST (OWASP ZAP)",
            "findings": categorized,
            "total_issues": len(findings),
            "status": "FAILED" if categorized["HIGH"] or categorized["CRITICAL"] else "PASSED"
        }

    def _categorize_findings(self, findings: List[Dict]) -> Dict[str, List]:
        """Categorize findings by severity."""
        categorized = {
            "CRITICAL": [],
            "HIGH": [],
            "MEDIUM": [],
            "LOW": []
        }

        for finding in findings:
            severity = finding.get("severity", "MEDIUM")
            categorized[severity].append(finding)

        return categorized

    def _determine_pipeline_status(self, results: Dict) -> str:
        """Determine overall pipeline status based on findings."""
        for stage_name, stage_result in results["stages"].items():
            if stage_result.get("status") == "FAILED":
                # Check if failure is blocking
                findings = stage_result.get("findings", {})
                if findings.get("CRITICAL") or findings.get("HIGH"):
                    return "FAILED"

        return "PASSED"

    def _extract_blocking_issues(self, results: Dict) -> List[Dict]:
        """Extract blocking security issues."""
        blocking = []

        for stage_name, stage_result in results["stages"].items():
            findings = stage_result.get("findings", {})

            for critical in findings.get("CRITICAL", []):
                blocking.append({
                    "stage": stage_name,
                    "severity": "CRITICAL",
                    "issue": critical
                })

            for high in findings.get("HIGH", []):
                blocking.append({
                    "stage": stage_name,
                    "severity": "HIGH",
                    "issue": high
                })

        return blocking
```

---

## 4. Cryptography & Secrets Management

**Concept**: Secure encryption patterns and centralized secrets management with HashiCorp Vault.

### 4.1 Production-Grade Encryption Patterns

```python
class CryptographyManager:
    """Enterprise-grade encryption for data at rest and in transit."""

    def __init__(self, key_management_service):
        """
        Initialize cryptography manager.

        Args:
            key_management_service: KMS for key storage (Vault, AWS KMS, GCP KMS)
        """
        self.kms = key_management_service

    def encrypt_at_rest(
        self,
        plaintext: bytes,
        context: Dict[str, str]
    ) -> EncryptedData:
        """
        Encrypt data at rest using AES-256-GCM.

        Args:
            plaintext: Data to encrypt
            context: Encryption context for key derivation

        Returns:
            Encrypted data with metadata

        Example:
            >>> crypto = CryptographyManager(vault_kms)
            >>> encrypted = crypto.encrypt_at_rest(
            ...     b"Sensitive user data",
            ...     context={"user_id": "12345", "data_type": "PII"}
            ... )
            >>> # Encrypted with AES-256-GCM + unique nonce
        """
        # Get data encryption key (DEK) from KMS
        dek = self.kms.get_data_key(context)

        # Generate random nonce (12 bytes for GCM)
        nonce = os.urandom(12)

        # Encrypt with AES-256-GCM
        cipher = Cipher(
            algorithms.AES(dek),
            modes.GCM(nonce),
            backend=default_backend()
        )

        encryptor = cipher.encryptor()
        ciphertext = encryptor.update(plaintext) + encryptor.finalize()

        # Get authentication tag
        tag = encryptor.tag

        return EncryptedData(
            ciphertext=ciphertext,
            nonce=nonce,
            tag=tag,
            algorithm="AES-256-GCM",
            key_id=dek.key_id,
            context=context
        )

    def decrypt_at_rest(self, encrypted_data: EncryptedData) -> bytes:
        """
        Decrypt data encrypted at rest.

        Args:
            encrypted_data: Encrypted data with metadata

        Returns:
            Decrypted plaintext

        Raises:
            AuthenticationError: If authentication tag verification fails
        """
        # Get DEK from KMS
        dek = self.kms.get_data_key(encrypted_data.context, encrypted_data.key_id)

        # Decrypt with AES-256-GCM
        cipher = Cipher(
            algorithms.AES(dek),
            modes.GCM(encrypted_data.nonce, encrypted_data.tag),
            backend=default_backend()
        )

        decryptor = cipher.decryptor()

        try:
            plaintext = decryptor.update(encrypted_data.ciphertext) + decryptor.finalize()
            return plaintext
        except Exception as e:
            raise AuthenticationError("Decryption failed - data may be tampered") from e

    def encrypt_in_transit(self, data: bytes, recipient_public_key: bytes) -> bytes:
        """
        Encrypt data for transmission using hybrid encryption.

        Approach:
            1. Generate ephemeral AES-256 key
            2. Encrypt data with AES-256-GCM
            3. Encrypt AES key with recipient's RSA-4096 public key
            4. Return encrypted key + encrypted data

        Args:
            data: Data to encrypt
            recipient_public_key: Recipient's RSA public key (PEM format)

        Returns:
            Encrypted package (encrypted_key + ciphertext)

        Example:
            >>> crypto = CryptographyManager(vault_kms)
            >>> encrypted_package = crypto.encrypt_in_transit(
            ...     b"Confidential message",
            ...     recipient_public_key=rsa_public_key
            ... )
            >>> # Hybrid encryption: RSA-4096 + AES-256-GCM
        """
        # Generate ephemeral AES key
        aes_key = os.urandom(32)  # 256 bits

        # Encrypt data with AES-256-GCM
        nonce = os.urandom(12)
        cipher = Cipher(
            algorithms.AES(aes_key),
            modes.GCM(nonce),
            backend=default_backend()
        )

        encryptor = cipher.encryptor()
        ciphertext = encryptor.update(data) + encryptor.finalize()
        tag = encryptor.tag

        # Encrypt AES key with recipient's RSA public key
        recipient_key = serialization.load_pem_public_key(recipient_public_key)
        encrypted_aes_key = recipient_key.encrypt(
            aes_key,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )

        # Package encrypted data
        package = {
            "encrypted_key": base64.b64encode(encrypted_aes_key).decode(),
            "nonce": base64.b64encode(nonce).decode(),
            "tag": base64.b64encode(tag).decode(),
            "ciphertext": base64.b64encode(ciphertext).decode()
        }

        return json.dumps(package).encode()
```

### 4.2 HashiCorp Vault Integration

```python
class VaultSecretsManager:
    """Manage secrets with HashiCorp Vault."""

    def __init__(self, vault_url: str, auth_token: str):
        """
        Initialize Vault secrets manager.

        Args:
            vault_url: Vault server URL
            auth_token: Authentication token
        """
        self.client = hvac.Client(url=vault_url, token=auth_token)

    def store_secret(
        self,
        path: str,
        secret_data: Dict[str, Any],
        metadata: Optional[Dict[str, str]] = None
    ) -> None:
        """
        Store secret in Vault with metadata.

        Args:
            path: Secret path in Vault (e.g., "database/prod/credentials")
            secret_data: Secret data to store
            metadata: Optional metadata (owner, expiry, rotation_policy)

        Example:
            >>> vault = VaultSecretsManager("https://vault.example.com", token)
            >>> vault.store_secret(
            ...     "database/prod/credentials",
            ...     {
            ...         "username": "app_user",
            ...         "password": "secure_password_123"
            ...     },
            ...     metadata={"rotation_days": "30", "owner": "backend-team"}
            ... )
        """
        # Store secret in KV v2 engine
        self.client.secrets.kv.v2.create_or_update_secret(
            path=path,
            secret=secret_data,
            metadata=metadata or {}
        )

    def retrieve_secret(self, path: str, version: Optional[int] = None) -> Dict[str, Any]:
        """
        Retrieve secret from Vault.

        Args:
            path: Secret path
            version: Optional version (defaults to latest)

        Returns:
            Secret data
        """
        response = self.client.secrets.kv.v2.read_secret_version(
            path=path,
            version=version
        )

        return response["data"]["data"]

    def rotate_secret(
        self,
        path: str,
        new_secret_data: Dict[str, Any]
    ) -> int:
        """
        Rotate secret to new version.

        Args:
            path: Secret path
            new_secret_data: New secret data

        Returns:
            New version number

        Example:
            >>> new_version = vault.rotate_secret(
            ...     "database/prod/credentials",
            ...     {"username": "app_user", "password": "new_secure_password_456"}
            ... )
            >>> # Rotated to version 2, old version 1 still accessible
        """
        response = self.client.secrets.kv.v2.create_or_update_secret(
            path=path,
            secret=new_secret_data
        )

        return response["data"]["version"]

    def enable_dynamic_secrets(
        self,
        database_config: Dict[str, Any]
    ) -> None:
        """
        Enable dynamic database credentials.

        Args:
            database_config: Database connection configuration

        Example:
            >>> vault.enable_dynamic_secrets({
            ...     "plugin_name": "postgresql-database-plugin",
            ...     "connection_url": "postgresql://{{username}}:{{password}}@localhost:5432/mydb",
            ...     "username": "vault_admin",
            ...     "password": "admin_password",
            ...     "allowed_roles": ["readonly", "readwrite"]
            ... })
            >>> # Dynamic credentials generated on-demand with TTL
        """
        self.client.sys.enable_secrets_engine(
            backend_type="database",
            path="database"
        )

        self.client.secrets.database.configure(
            name="my-postgresql-database",
            plugin_name=database_config["plugin_name"],
