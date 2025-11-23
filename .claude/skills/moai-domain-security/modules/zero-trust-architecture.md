# Zero-Trust Architecture

Implement "never trust, always verify" security model with adaptive authentication.

## Zero-Trust Principles

1. **Verify Every Access**: Every access request authenticated
2. **Least Privilege**: Grant minimum required permissions
3. **Assume Breach**: Design assuming compromise
4. **Verify Explicitly**: Use all data points for decisions
5. **Secure by Default**: Security-first approach

## Adaptive Authentication

### Risk-Based Authentication

```python
class AdaptiveAuthenticationEngine:
    def evaluate_risk(self, user_id: str, context: dict) -> dict:
        risk_score = 0

        # Geolocation risk
        if self._is_unusual_location(user_id, context['ip']):
            risk_score += 30

        # Device risk
        if self._is_new_device(user_id, context['device_fingerprint']):
            risk_score += 25

        # Time-based risk
        if self._is_unusual_time(user_id):
            risk_score += 15

        # Behavior risk
        if self._is_anomalous_behavior(user_id, context):
            risk_score += 20

        return {
            'risk_score': min(risk_score, 100),
            'risk_level': self._determine_level(risk_score),
            'required_auth_factors': self._determine_factors(risk_score)
        }

    def _determine_level(self, score: int) -> str:
        if score < 20:
            return 'low'
        elif score < 50:
            return 'medium'
        else:
            return 'high'

    def _determine_factors(self, score: int) -> list:
        factors = ['password']  # Base requirement

        if score >= 20:
            factors.append('totp')  # Two-factor authentication

        if score >= 50:
            factors.extend(['biometric', 'hardware_key'])  # Strong authentication

        return factors
```

### Multi-Factor Authentication

```python
import pyotp
import qrcode

class MFAManager:
    def setup_totp(self, user_id: str) -> dict:
        # Time-based one-time password
        secret = pyotp.random_base32()
        totp = pyotp.TOTP(secret)

        # Generate QR code for authenticator apps
        qr = qrcode.QRCode()
        qr.add_data(totp.provisioning_uri(
            name=f'user_{user_id}',
            issuer_name='SecureApp'
        ))
        qr.make()

        return {
            'secret': secret,
            'qr_code': qr.get_matrix(),
            'backup_codes': self._generate_backup_codes()
        }

    def verify_totp(self, user_id: str, token: str, secret: str) -> bool:
        totp = pyotp.TOTP(secret)

        # Allow time drift for network latency
        return totp.verify(token, valid_window=1)

    def setup_hardware_key(self, user_id: str) -> dict:
        # WebAuthn/FIDO2 hardware key registration
        challenge = secrets.token_bytes(32)

        return {
            'challenge': challenge.hex(),
            'user_id': user_id,
            'attestation_required': True
        }

    def _generate_backup_codes(self, count: int = 10) -> list:
        return [secrets.token_hex(4) for _ in range(count)]
```

## Session Management

### Secure Session Tokens

```python
import secrets
from datetime import datetime, timedelta

class SessionManager:
    def create_session(self, user_id: str, context: dict) -> dict:
        session_id = secrets.token_urlsafe(32)

        session = {
            'session_id': session_id,
            'user_id': user_id,
            'created_at': datetime.utcnow(),
            'expires_at': datetime.utcnow() + timedelta(hours=1),
            'last_activity': datetime.utcnow(),
            'device_fingerprint': context.get('device_fingerprint'),
            'ip_address': context.get('ip_address'),
            'user_agent': context.get('user_agent')
        }

        self.db.save(session)
        return session

    def validate_session(self, session_id: str, context: dict) -> bool:
        session = self.db.find_session(session_id)

        if not session:
            return False

        # Validate expiration
        if datetime.utcnow() > session['expires_at']:
            self.db.delete_session(session_id)
            return False

        # Validate device fingerprint
        if session['device_fingerprint'] != context.get('device_fingerprint'):
            raise SessionCompromisedError("Device fingerprint mismatch")

        # Update last activity
        session['last_activity'] = datetime.utcnow()
        self.db.save(session)

        return True

    def revoke_session(self, session_id: str):
        self.db.delete_session(session_id)

    def revoke_all_sessions(self, user_id: str):
        sessions = self.db.find_sessions_by_user(user_id)
        for session in sessions:
            self.db.delete_session(session['session_id'])
```

## Continuous Authentication

### Behavioral Biometrics

```python
class BehavioralBiometrics:
    def analyze_user_behavior(self, user_id: str, action: dict) -> dict:
        # Mouse movement patterns
        mouse_velocity = self._calculate_mouse_velocity(action['mouse_events'])

        # Keyboard patterns
        keystroke_dynamics = self._analyze_keystroke_dynamics(action['keyboard_events'])

        # Touch patterns (mobile)
        touch_pattern = self._analyze_touch_pattern(action.get('touch_events'))

        risk_score = 0
        if mouse_velocity > self._get_baseline('mouse_velocity', user_id):
            risk_score += 10

        if keystroke_dynamics['variance'] > 0.3:
            risk_score += 15

        return {
            'risk_score': risk_score,
            'anomaly_detected': risk_score > 30,
            'recommended_action': 'challenge' if risk_score > 50 else 'allow'
        }

    def _get_baseline(self, metric: str, user_id: str) -> float:
        historical_data = self.db.get_user_baseline(user_id, metric)
        if not historical_data:
            return float('inf')
        return sum(historical_data) / len(historical_data)
```

## Access Token Management

### JWT Tokens

```python
import jwt
from datetime import datetime, timedelta

class JWTManager:
    def create_access_token(self, user_id: str, claims: dict) -> str:
        payload = {
            'user_id': user_id,
            'iat': datetime.utcnow(),
            'exp': datetime.utcnow() + timedelta(minutes=15),
            **claims
        }

        token = jwt.encode(
            payload,
            self.secret_key,
            algorithm='HS256'
        )

        return token

    def create_refresh_token(self, user_id: str) -> str:
        payload = {
            'user_id': user_id,
            'type': 'refresh',
            'iat': datetime.utcnow(),
            'exp': datetime.utcnow() + timedelta(days=7)
        }

        token = jwt.encode(
            payload,
            self.refresh_secret,
            algorithm='HS256'
        )

        return token

    def verify_token(self, token: str) -> dict:
        try:
            payload = jwt.decode(
                token,
                self.secret_key,
                algorithms=['HS256']
            )

            # Check token blacklist (revocation)
            if self.is_token_blacklisted(token):
                raise jwt.InvalidTokenError("Token revoked")

            return payload

        except jwt.ExpiredSignatureError:
            raise jwt.InvalidTokenError("Token expired")
        except jwt.InvalidTokenError:
            raise jwt.InvalidTokenError("Invalid token")
```

## mTLS (Mutual TLS) Implementation

### Service-to-Service Authentication

```python
import ssl
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa

class mTLSManager:
    def generate_certificate(self, service_name: str) -> dict:
        # Generate private key
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048
        )

        # Generate certificate
        subject = issuer = x509.Name([
            x509.NameAttribute(NameOID.COUNTRY_NAME, "US"),
            x509.NameAttribute(NameOID.STATE_OR_PROVINCE_NAME, "CA"),
            x509.NameAttribute(NameOID.ORGANIZATION_NAME, "SecureOrg"),
            x509.NameAttribute(NameOID.COMMON_NAME, service_name),
        ])

        cert = x509.CertificateBuilder().subject_name(
            subject
        ).issuer_name(
            issuer
        ).public_key(
            private_key.public_key()
        ).serial_number(
            x509.random_serial_number()
        ).not_valid_before(
            datetime.utcnow()
        ).not_valid_after(
            datetime.utcnow() + timedelta(days=365)
        ).sign(private_key, hashes.SHA256())

        return {
            'certificate': cert,
            'private_key': private_key
        }

    def create_ssl_context(self, cert_path: str, key_path: str, ca_path: str) -> ssl.SSLContext:
        context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
        context.verify_mode = ssl.CERT_REQUIRED
        context.load_cert_chain(certfile=cert_path, keyfile=key_path)
        context.load_verify_locations(cafile=ca_path)

        return context
```

## Service Mesh Security

### Istio Authorization Policy

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: service-to-service-auth
spec:
  selector:
    matchLabels:
      app: backend-api
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/frontend"]
    to:
    - operation:
        methods: ["GET", "POST"]
        paths: ["/api/*"]
```

---

**Tools**: Okta, Auth0, Azure AD, Keycloak, HashiCorp Vault
