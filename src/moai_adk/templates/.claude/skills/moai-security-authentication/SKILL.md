# moai-security-authentication: Advanced Authentication Systems

**Expert-Level Authentication Implementation for Enterprise Applications**  
Trust Score: 9.9/10 | Version: 4.0.0 | Last Updated: 2025-11-11

## 🔐 Multi-Factor Authentication (MFA) Architecture

Modern authentication requires multiple layers of verification to protect against credential theft and replay attacks.

### Time-Based One-Time Password (TOTP) Implementation

```python
import pyotp
import qrcode
import io
import base64
from typing import Tuple, Optional, Dict, Any
from dataclasses import dataclass
from datetime import datetime, timedelta
import secrets

@dataclass
class MFACredentials:
    secret: str
    backup_codes: list
    qr_code: str
    setup_key: str

class TOTPService:
    def __init__(self, issuer_name: str = "MyApp"):
        self.issuer = issuer_name
        self.totp_validity_window = 1  # ±30 seconds
    
    def generate_mfa_credentials(self, user_id: str, email: str) -> MFACredentials:
        """Generate complete MFA credentials for user"""
        secret = pyotp.random_base32()
        backup_codes = [secrets.token_hex(4).upper() for _ in range(10)]
        
        totp_uri = pyotp.totp.TOTP(secret).provisioning_uri(
            name=email, issuer_name=self.issuer
        )
        
        qr = qrcode.QRCode(version=1, box_size=10, border=5)
        qr.add_data(totp_uri)
        qr.make(fit=True)
        
        img = qr.make_image(fill_color="black", back_color="white")
        buffer = io.BytesIO()
        img.save(buffer, format='PNG')
        qr_code_base64 = base64.b64encode(buffer.getvalue()).decode()
        
        return MFACredentials(
            secret=secret,
            backup_codes=backup_codes,
            qr_code=qr_code_base64,
            setup_key=secret
        )
    
    def verify_totp_token(self, secret: str, token: str) -> Tuple[bool, str]:
        """Verify TOTP token with time window tolerance"""
        try:
            totp = pyotp.TOTP(secret)
            is_valid = totp.verify(token, valid_window=self.totp_validity_window)
            return (True, "Token verified successfully") if is_valid else (False, "Invalid token")
        except Exception as e:
            return False, f"Token verification failed: {str(e)}"

# OAuth 2.0 Implementation
class OAuth2Service:
    def __init__(self):
        self.providers = {}
        self.state_store = {}
    
    def register_provider(self, name: str, config: dict) -> None:
        """Register OAuth provider configuration"""
        self.providers[name] = config
    
    def get_authorization_url(self, provider_name: str, redirect_uri: str) -> str:
        """Generate OAuth authorization URL with PKCE"""
        import secrets
        import hashlib
        import base64
        from urllib.parse import urlencode
        
        provider = self.providers.get(provider_name)
        if not provider:
            raise ValueError(f"Provider {provider_name} not registered")
        
        state = secrets.token_urlsafe(32)
        code_verifier = secrets.token_urlsafe(64)
        code_challenge = base64.urlsafe_b64encode(
            hashlib.sha256(code_verifier.encode()).digest()
        ).decode().rstrip('=')
        
        self.state_store[state] = {
            "provider": provider_name,
            "code_verifier": code_verifier
        }
        
        params = {
            "client_id": provider["client_id"],
            "redirect_uri": redirect_uri,
            "response_type": "code",
            "scope": " ".join(provider.get("scopes", ["read"])),
            "state": state,
            "code_challenge": code_challenge,
            "code_challenge_method": "S256"
        }
        
        return f"{provider['auth_url']}?{urlencode(params)}"

# JWT Authentication Service
import jwt
from datetime import datetime, timedelta

class JWTAuthenticationService:
    def __init__(self, secret_key: str, algorithm: str = "HS256"):
        self.secret_key = secret_key
        self.algorithm = algorithm
        self.token_expiry = 3600
        self.refresh_token_expiry = 86400 * 30
    
    def generate_token_pair(self, user_id: str, user_data: dict) -> dict:
        """Generate access token and refresh token pair"""
        current_time = datetime.utcnow()
        
        access_payload = {
            "user_id": user_id,
            "user_data": user_data,
            "token_type": "access",
            "iat": current_time,
            "exp": current_time + timedelta(seconds=self.token_expiry),
            "jti": secrets.token_urlsafe(32)
        }
        
        refresh_payload = {
            "user_id": user_id,
            "token_type": "refresh",
            "iat": current_time,
            "exp": current_time + timedelta(seconds=self.refresh_token_expiry),
            "jti": secrets.token_urlsafe(32)
        }
        
        return {
            "access_token": jwt.encode(access_payload, self.secret_key, algorithm=self.algorithm),
            "refresh_token": jwt.encode(refresh_payload, self.secret_key, algorithm=self.algorithm),
            "token_type": "Bearer",
            "expires_in": self.token_expiry
        }
    
    def verify_token(self, token: str) -> Optional[dict]:
        """Verify JWT token and return payload"""
        try:
            return jwt.decode(token, self.secret_key, algorithms=[self.algorithm])
        except jwt.ExpiredSignatureError:
            return None
        except jwt.InvalidTokenError:
            return None
