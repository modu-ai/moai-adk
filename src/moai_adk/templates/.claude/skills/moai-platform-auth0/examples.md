# Auth0 Platform Security - Production Examples

Comprehensive, production-ready code examples for implementing Auth0 security features including attack protection, MFA, token management, sender constraining, and compliance.

## Table of Contents

1. [Environment Setup](#1-environment-setup)
2. [Client Initialization and Authentication](#2-client-initialization-and-authentication)
3. [MFA Implementation](#3-mfa-implementation)
4. [Attack Protection Configuration](#4-attack-protection-configuration)
5. [Token Management](#5-token-management)
6. [DPoP Implementation](#6-dpop-implementation)
7. [mTLS Implementation](#7-mtls-implementation)
8. [GDPR Compliance](#8-gdpr-compliance)
9. [Error Handling and Retry Logic](#9-error-handling-and-retry-logic)
10. [Security Monitoring](#10-security-monitoring)

---

## 1. Environment Setup

### 1.1 Environment Variables (.env)

```bash
# Auth0 Configuration
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
AUTH0_AUDIENCE=https://api.yourapp.com

# Token Configuration
TOKEN_ISSUER=https://your-tenant.auth0.com/
TOKEN_ALGORITHM=RS256
TOKEN_LEEWAY=30

# MFA Configuration
MFA_ENABLED=true
ADAPTIVE_MFA_ENABLED=true
MFA_FACTORS=otp,push,webauthn

# Attack Protection
BOT_DETECTION_SENSITIVITY=medium
BRUTE_FORCE_THRESHOLD=10
SUSPICIOUS_IP_THROTTLE_ENABLED=true

# DPoP Configuration (if enabled)
DPOP_ENABLED=false
DPOP_PRIVATE_KEY_PATH=/path/to/private_key.pem

# Compliance
GDPR_ENABLED=true
DATA_RETENTION_DAYS=365
CONSENT_MANAGEMENT_ENABLED=true

# Application URLs
APP_BASE_URL=https://yourapp.com
APP_CALLBACK_URL=https://yourapp.com/callback
APP_LOGOUT_URL=https://yourapp.com/logout

# API Configuration
API_BASE_URL=https://api.yourapp.com
API_TIMEOUT=30000
API_MAX_RETRIES=3

# Logging
LOG_LEVEL=info
SECURITY_LOG_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90
```

### 1.2 Configuration Loader (Python)

```python
# config.py
import os
from dataclasses import dataclass
from typing import Optional
from dotenv import load_dotenv

load_dotenv()

@dataclass
class Auth0Config:
    """Auth0 configuration from environment variables"""

    domain: str
    client_id: str
    client_secret: str
    audience: str
    token_issuer: str
    token_algorithm: str = "RS256"
    token_leeway: int = 30

    @classmethod
    def from_env(cls) -> "Auth0Config":
        """Load configuration from environment variables"""
        domain = os.getenv("AUTH0_DOMAIN")
        if not domain:
            raise ValueError("AUTH0_DOMAIN is required")

        client_id = os.getenv("AUTH0_CLIENT_ID")
        if not client_id:
            raise ValueError("AUTH0_CLIENT_ID is required")

        client_secret = os.getenv("AUTH0_CLIENT_SECRET")
        if not client_secret:
            raise ValueError("AUTH0_CLIENT_SECRET is required")

        audience = os.getenv("AUTH0_AUDIENCE")
        if not audience:
            raise ValueError("AUTH0_AUDIENCE is required")

        return cls(
            domain=domain,
            client_id=client_id,
            client_secret=client_secret,
            audience=audience,
            token_issuer=os.getenv("TOKEN_ISSUER", f"https://{domain}/"),
            token_algorithm=os.getenv("TOKEN_ALGORITHM", "RS256"),
            token_leeway=int(os.getenv("TOKEN_LEEWAY", "30"))
        )

@dataclass
class SecurityConfig:
    """Security configuration from environment variables"""

    mfa_enabled: bool = False
    adaptive_mfa_enabled: bool = False
    mfa_factors: list = None

    bot_detection_sensitivity: str = "medium"
    brute_force_threshold: int = 10
    suspicious_ip_throttle_enabled: bool = False

    dpop_enabled: bool = False
    dpop_private_key_path: Optional[str] = None

    gdpr_enabled: bool = False
    data_retention_days: int = 365
    consent_management_enabled: bool = False

    @classmethod
    def from_env(cls) -> "SecurityConfig":
        """Load security configuration from environment variables"""
        return cls(
            mfa_enabled=os.getenv("MFA_ENABLED", "false").lower() == "true",
            adaptive_mfa_enabled=os.getenv("ADAPTIVE_MFA_ENABLED", "false").lower() == "true",
            mfa_factors=os.getenv("MFA_FACTORS", "").split(",") if os.getenv("MFA_FACTORS") else [],
            bot_detection_sensitivity=os.getenv("BOT_DETECTION_SENSITIVITY", "medium"),
            brute_force_threshold=int(os.getenv("BRUTE_FORCE_THRESHOLD", "10")),
            suspicious_ip_throttle_enabled=os.getenv("SUSPICIOUS_IP_THROTTLE_ENABLED", "false").lower() == "true",
            dpop_enabled=os.getenv("DPOP_ENABLED", "false").lower() == "true",
            dpop_private_key_path=os.getenv("DPOP_PRIVATE_KEY_PATH"),
            gdpr_enabled=os.getenv("GDPR_ENABLED", "false").lower() == "true",
            data_retention_days=int(os.getenv("DATA_RETENTION_DAYS", "365")),
            consent_management_enabled=os.getenv("CONSENT_MANAGEMENT_ENABLED", "false").lower() == "true"
        )
```

---

## 2. Client Initialization and Authentication

### 2.1 Auth0 Client Factory (Python)

```python
# auth0_client.py
import httpx
from typing import Optional, Dict, Any
from config import Auth0Config, SecurityConfig

class Auth0Client:
    """
    Auth0 authentication client with comprehensive security features
    """

    def __init__(self, auth0_config: Auth0Config, security_config: SecurityConfig):
        self.config = auth0_config
        self.security = security_config
        self.domain = auth0_config.domain
        self.client_id = auth0_config.client_id
        self.client_secret = auth0_config.client_secret
        self.audience = auth0_config.audience

        # HTTP client with security settings
        self.http_client = httpx.AsyncClient(
            base_url=f"https://{self.domain}",
            timeout=30.0,
            headers={
                "Content-Type": "application/json",
                "User-Agent": f"Auth0PythonClient/1.0"
            }
        )

    async def close(self):
        """Close HTTP client"""
        await self.http_client.aclose()

    async def get_token(self, code: str, redirect_uri: str, code_verifier: Optional[str] = None) -> Dict[str, Any]:
        """
        Exchange authorization code for tokens

        Args:
            code: Authorization code from login
            redirect_uri: Redirect URI used in authorization
            code_verifier: PKCE code verifier (for public clients)

        Returns:
            Dictionary with access_token, refresh_token, id_token, etc.
        """
        token_endpoint = f"https://{self.domain}/oauth/token"

        payload = {
            "grant_type": "authorization_code",
            "client_id": self.client_id,
            "code": code,
            "redirect_uri": redirect_uri
        }

        # Add client secret for confidential clients
        if self.client_secret:
            payload["client_secret"] = self.client_secret

        # Add PKCE verifier for public clients
        if code_verifier:
            payload["code_verifier"] = code_verifier

        response = await self.http_client.post(token_endpoint, json=payload)
        response.raise_for_status()

        tokens = response.json()

        # Validate token response
        if "access_token" not in tokens:
            raise ValueError("Token response missing access_token")

        return tokens

    async def refresh_token(self, refresh_token: str) -> Dict[str, Any]:
        """
        Refresh access token using refresh token

        Args:
            refresh_token: Refresh token from initial authentication

        Returns:
            Dictionary with new access_token and refresh_token (if rotation enabled)
        """
        token_endpoint = f"https://{self.domain}/oauth/token"

        payload = {
            "grant_type": "refresh_token",
            "client_id": self.client_id,
            "refresh_token": refresh_token
        }

        if self.client_secret:
            payload["client_secret"] = self.client_secret

        response = await self.http_client.post(token_endpoint, json=payload)
        response.raise_for_status()

        return response.json()

    async def get_user_info(self, access_token: str) -> Dict[str, Any]:
        """
        Get user information from Auth0

        Args:
            access_token: Valid access token

        Returns:
            Dictionary with user profile information
        """
        userinfo_endpoint = f"https://{self.domain}/userinfo"

        response = await self.http_client.get(
            userinfo_endpoint,
            headers={"Authorization": f"Bearer {access_token}"}
        )
        response.raise_for_status()

        return response.json()

    async def revoke_token(self, token: str) -> None:
        """
        Revoke a token (requires Management API)

        Args:
            token: Token to revoke
        """
        # Note: This requires Management API credentials
        # Implementation depends on your setup
        pass
```

### 2.2 JWT Token Validator (Python)

```python
# token_validator.py
import jwt
import httpx
from typing import Dict, Any, Optional
from datetime import datetime, timedelta
from config import Auth0Config

class TokenValidator:
    """
    JWT token validation with comprehensive security checks
    """

    def __init__(self, config: Auth0Config):
        self.config = config
        self.jwks_url = f"https://{config.domain}/.well-known/jwks.json"
        self.jwks_cache: Dict[str, Any] = {}
        self.jwks_cache_expiry: Optional[datetime] = None

    async def get_jwks(self) -> Dict[str, Any]:
        """
        Fetch JSON Web Key Set from Auth0

        Returns:
            Dictionary with keys
        """
        # Check cache
        if self.jwks_cache and self.jwks_cache_expiry and datetime.now() < self.jwks_cache_expiry:
            return self.jwks_cache

        async with httpx.AsyncClient() as client:
            response = await client.get(self.jwks_url)
            response.raise_for_status()
            jwks = response.json()

        # Cache for 1 hour
        self.jwks_cache = jwks
        self.jwks_cache_expiry = datetime.now() + timedelta(hours=1)

        return jwks

    async def validate_access_token(self, token: str) -> Dict[str, Any]:
        """
        Validate access token with comprehensive checks

        Args:
            token: JWT access token

        Returns:
            Decoded token payload

        Raises:
            ValueError: If token is invalid
        """
        try:
            # Get JWKS for signature verification
            jwks = await self.get_jwks()

            # Decode and validate token
            payload = jwt.decode(
                token,
                key=self._get_public_key(jwks, token),
                algorithms=[self.config.token_algorithm],
                audience=self.config.audience,
                issuer=self.config.token_issuer,
                leeway=self.config.token_leeway
            )

            # Additional validation checks
            self._validate_token_claims(payload)

            return payload

        except jwt.ExpiredSignatureError:
            raise ValueError("Token has expired")
        except jwt.InvalidTokenError as e:
            raise ValueError(f"Invalid token: {str(e)}")

    def _get_public_key(self, jwks: Dict[str, Any], token: str) -> str:
        """
        Extract public key from JWKS using token header

        Args:
            jwks: JSON Web Key Set
            token: JWT token

        Returns:
            Public key string
        """
        unverified_header = jwt.get_unverified_header(token)
        kid = unverified_header.get("kid")

        if not kid:
            raise ValueError("Token missing 'kid' in header")

        # Find matching key
        for key in jwks.get("keys", []):
            if key.get("kid") == kid:
                return jwt.algorithms.RSAAlgorithm.from_jwk(key)

        raise ValueError(f"Unable to find key with kid: {kid}")

    def _validate_token_claims(self, payload: Dict[str, Any]) -> None:
        """
        Validate token claims for security

        Args:
            payload: Decoded token payload

        Raises:
            ValueError: If claims are invalid
        """
        # Required claims
        required_claims = ["iss", "sub", "aud", "exp", "iat"]
        for claim in required_claims:
            if claim not in payload:
                raise ValueError(f"Token missing required claim: {claim}")

        # Validate scope (if present)
        if "scope" in payload:
            scopes = payload["scope"].split()
            # Add custom scope validation here
            pass

        # Validate token type
        if payload.get("typ") and payload["typ"] != "Bearer":
            raise ValueError(f"Invalid token type: {payload['typ']}")

    def validate_id_token(self, token: str, nonce: Optional[str] = None) -> Dict[str, Any]:
        """
        Validate ID token with nonce verification

        Args:
            token: JWT ID token
            nonce: Original nonce from authorization request

        Returns:
            Decoded token payload

        Raises:
            ValueError: If token is invalid
        """
        payload = await self.validate_access_token(token)

        # Validate nonce if provided
        if nonce and payload.get("nonce") != nonce:
            raise ValueError("Token nonce does not match")

        return payload
```

---

## 3. MFA Implementation

### 3.1 MFA Challenge Handler (Python)

```python
# mfa_handler.py
from typing import Dict, Any, Optional
from enum import Enum
from auth0_client import Auth0Client

class MFAMethod(Enum):
    """Supported MFA methods"""
    OTP = "otp"
    PUSH = "push"
    WEBAUTHN = "webauthn"
    SMS = "sms"
    VOICE = "voice"

class MFAHandler:
    """
    Multi-Factor Authentication handler for Auth0
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client

    async def initiate_mfa_challenge(
        self,
        access_token: str,
        mfa_token: str,
        method: MFAMethod = MFAMethod.OTP
    ) -> Dict[str, Any]:
        """
        Initiate MFA challenge

        Args:
            access_token: Valid access token
            mfa_token: MFA token from authentication
            method: MFA method to use

        Returns:
            Challenge response with challenge_id or oob_code
        """
        mfa_endpoint = f"https://{self.client.domain}/mfa/challenge"

        payload = {
            "mfa_token": mfa_token,
            "client_id": self.client.client_id,
            "challenge_type": method.value
        }

        if self.client.client_secret:
            payload["client_secret"] = self.client.client_secret

        response = await self.client.http_client.post(
            mfa_endpoint,
            json=payload,
            headers={"Authorization": f"Bearer {access_token}"}
        )
        response.raise_for_status()

        return response.json()

    async def verify_mfa_challenge(
        self,
        access_token: str,
        challenge_id: str,
        code: str,
        oob_code: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Verify MFA challenge response

        Args:
            access_token: Valid access token
            challenge_id: Challenge ID from initiate_mfa_challenge
            code: OTP code from user
            oob_code: Out-of-band code (for push notifications)

        Returns:
            Verification result with tokens
        """
        verify_endpoint = f"https://{self.client.domain}/mfa/verify"

        payload = {
            "mfa_token": challenge_id,
            "client_id": self.client.client_id,
            "code": code
        }

        if oob_code:
            payload["oob_code"] = oob_code

        if self.client.client_secret:
            payload["client_secret"] = self.client.client_secret

        response = await self.client.http_client.post(
            verify_endpoint,
            json=payload,
            headers={"Authorization": f"Bearer {access_token}"}
        )
        response.raise_for_status()

        return response.json()

    async def associate_mfa_factor(
        self,
        access_token: str,
        factor_type: MFAMethod,
        factor_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Associate MFA factor with user account

        Args:
            access_token: Valid access token
            factor_type: Type of MFA factor
            factor_data: Factor-specific data (e.g., phone number for SMS)

        Returns:
            Association response
        """
        associate_endpoint = f"https://{self.client.domain}/mfa/associate"

        payload = {
            "client_id": self.client.client_id,
            "factor_type": factor_type.value,
            **factor_data
        }

        response = await self.client.http_client.post(
            associate_endpoint,
            json=payload,
            headers={"Authorization": f"Bearer {access_token}"}
        )
        response.raise_for_status()

        return response.json()
```

### 3.2 Adaptive MFA Integration (Python)

```python
# adaptive_mfa.py
from typing import Dict, Any, Optional
from datetime import datetime
from auth0_client import Auth0Client

class AdaptiveMFAEvaluator:
    """
    Adaptive MFA risk evaluation and challenge logic
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client

    async def evaluate_login_risk(
        self,
        user_id: str,
        ip_address: str,
        user_agent: str,
        location: Optional[Dict[str, float]] = None
    ) -> Dict[str, Any]:
        """
        Evaluate login risk using Auth0 Adaptive MFA signals

        Args:
            user_id: User ID
            ip_address: IP address of login attempt
            user_agent: User agent string
            location: Optional location data (lat, lon)

        Returns:
            Risk assessment with score and recommended action
        """
        # Collect risk signals
        signals = await self._collect_risk_signals(
            user_id, ip_address, user_agent, location
        )

        # Calculate risk score
        risk_score = self._calculate_risk_score(signals)

        # Determine action
        action = self._determine_action(risk_score, signals)

        return {
            "risk_score": risk_score,
            "action": action,
            "signals": signals,
            "timestamp": datetime.utcnow().isoformat()
        }

    async def _collect_risk_signals(
        self,
        user_id: str,
        ip_address: str,
        user_agent: str,
        location: Optional[Dict[str, float]]
    ) -> Dict[str, Any]:
        """Collect risk signals for evaluation"""

        signals = {
            "new_device": await self._is_new_device(user_id, user_agent),
            "impossible_travel": await self._check_impossible_travel(
                user_id, location
            ) if location else False,
            "untrusted_ip": await self._is_untrusted_ip(ip_address),
            "anomalous_location": await self._is_anomalous_location(
                user_id, location
            ) if location else False
        }

        return signals

    async def _is_new_device(self, user_id: str, user_agent: str) -> bool:
        """Check if device is new (not seen in 30 days)"""
        # Implementation: Check user agent against recent logins
        # This would typically query your user device tracking database
        return False  # Placeholder

    async def _check_impossible_travel(
        self,
        user_id: str,
        location: Dict[str, float]
    ) -> bool:
        """Check for impossible travel (geographic anomalies)"""
        # Implementation: Compare with last known location
        # Calculate travel velocity and assess feasibility
        return False  # Placeholder

    async def _is_untrusted_ip(self, ip_address: str) -> bool:
        """Check if IP is untrusted (has suspicious activity)"""
        # Implementation: Check against IP reputation databases
        # Could use Auth0's Suspicious IP Throttling data
        return False  # Placeholder

    async def _is_anomalous_location(
        self,
        user_id: str,
        location: Dict[str, float]
    ) -> bool:
        """Check if location is anomalous for user"""
        # Implementation: Compare with user's typical locations
        return False  # Placeholder

    def _calculate_risk_score(self, signals: Dict[str, Any]) -> float:
        """
        Calculate risk score from signals

        Returns:
            Risk score (0.0 = low risk, 1.0 = high risk)
        """
        risk_score = 0.0

        # Weighted risk factors
        if signals["new_device"]:
            risk_score += 0.3

        if signals["impossible_travel"]:
            risk_score += 0.4

        if signals["untrusted_ip"]:
            risk_score += 0.5

        if signals["anomalous_location"]:
            risk_score += 0.2

        return min(risk_score, 1.0)

    def _determine_action(self, risk_score: float, signals: Dict[str, Any]) -> str:
        """
        Determine action based on risk score

        Returns:
            Action: "allow", "mfa", "block"
        """
        if risk_score >= 0.7:
            return "block"

        if risk_score >= 0.4:
            return "mfa"

        return "allow"
```

---

## 4. Attack Protection Configuration

### 4.1 Attack Protection Monitor (Python)

```python
# attack_protection.py
from typing import Dict, Any, List
from datetime import datetime, timedelta
from auth0_client import Auth0Client
from enum import Enum

class AttackType(Enum):
    """Attack types to monitor"""
    CREDENTIAL_STUFFING = "credential_stuffing"
    BRUTE_FORCE = "brute_force"
    BOT_ATTACK = "bot_attack"
    SUSPICIOUS_IP = "suspicious_ip"

class AttackProtectionMonitor:
    """
    Monitor and respond to Auth0 attack protection events
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client
        self.base_url = f"https://{auth0_client.domain}/api/v2"

    async def get_attack_events(
        self,
        attack_type: Optional[AttackType] = None,
        from_date: Optional[datetime] = None,
        to_date: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        """
        Get attack protection events from Auth0 logs

        Args:
            attack_type: Filter by attack type
            from_date: Start date for events (default: 24 hours ago)
            to_date: End date for events (default: now)

        Returns:
            List of attack events
        """
        # Default time range: last 24 hours
        if not from_date:
            from_date = datetime.utcnow() - timedelta(hours=24)
        if not to_date:
            to_date = datetime.utcnow()

        # Query Auth0 logs for attack events
        # This requires Management API access
        logs_endpoint = f"{self.base_url}/logs"

        params = {
            "q": f"type:{attack_type.value} if attack_type else 'type:*attack*'",
            "from": from_date.isoformat(),
            "to": to_date.isoformat(),
            "per_page": 100
        }

        response = await self.client.http_client.get(
            logs_endpoint,
            params=params,
            headers={"Authorization": f"Bearer {await self._get_management_token()}"}
        )
        response.raise_for_status()

        return response.json().get("logs", [])

    async def get_suspicious_ip_throttling(self) -> Dict[str, Any]:
        """
        Get current suspicious IP throttling status

        Returns:
            Throttling configuration and statistics
        """
        # This would query Auth0 Management API
        # Implementation depends on API availability
        return {
            "enabled": True,
            "thresholds": {
                "login_per_day": 100,
                "signup_per_minute": 10
            },
            "blocked_ips": []
        }

    async def is_ip_blocked(self, ip_address: str) -> bool:
        """
        Check if IP address is currently blocked

        Args:
            ip_address: IP address to check

        Returns:
            True if IP is blocked
        """
        # Implementation: Check against Auth0's blocked IP list
        # Could also check local cache of blocked IPs
        return False

    async def handle_brute_force_protection(
        self,
        user_id: str,
        ip_address: str
    ) -> Dict[str, Any]:
        """
        Handle brute force protection event

        Args:
            user_id: User ID being targeted
            ip_address: IP address of attack

        Returns:
            Action taken
        """
        # Check if account is locked
        is_locked = await self._is_account_locked(user_id)

        if is_locked:
            return {
                "action": "account_locked",
                "message": "Account has been locked due to suspicious activity"
            }

        # Check if IP is blocked
        ip_blocked = await self.is_ip_blocked(ip_address)

        if ip_blocked:
            return {
                "action": "ip_blocked",
                "message": "IP address has been blocked"
            }

        return {
            "action": "monitoring",
            "message": "Monitoring suspicious activity"
        }

    async def _is_account_locked(self, user_id: str) -> bool:
        """Check if user account is locked"""
        # Implementation: Query Auth0 user status
        return False

    async def _get_management_token(self) -> str:
        """Get Management API token"""
        # Implementation: Use client credentials flow
        # to get Management API access token
        return ""
```

### 4.2 Breached Password Detection (Python)

```python
# breached_password.py
from typing import Dict, Any, Optional
from auth0_client import Auth0Client

class BreachedPasswordDetector:
    """
    Detect and block breached passwords using Auth0's Breached Password Detection
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client

    async def check_password_breached(
        self,
        password: str,
        user_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Check if password has been breached

        Args:
            password: Password to check
            user_id: Optional user ID for logging

        Returns:
            Breach status and action required
        """
        # Note: Actual breached password checking happens on Auth0 side
        # during authentication. This is a client-side check using
        # Have I Been Pwned API (optional enhancement).

        # For production, rely on Auth0's built-in breached password detection

        return {
            "breached": False,
            "action": "allow"
        }

    async def handle_breached_password(self, user_id: str) -> Dict[str, Any]:
        """
        Handle breached password detection

        Args:
            user_id: User ID

        Returns:
            Action taken (password reset required)
        """
        # Implementation:
        # 1. Block login attempt
        # 2. Trigger password reset email
        # 3. Log security event
        # 4. Notify user and admin

        return {
            "action": "password_reset_required",
            "message": "Your password has been detected in a data breach. Please reset your password."
        }

    def test_breached_password(self) -> str:
        """
        Generate test password for breached password detection

        Returns:
            Test password that will trigger breached password detection
        """
        # Auth0 provides test passwords that start with AUTH0-TEST-
        # These will always trigger breached password detection
        return "AUTH0-TEST-breached123"
```

---

## 5. Token Management

### 5.1 Refresh Token Rotation (Python)

```python
# token_rotation.py
from typing import Dict, Any, Optional
from datetime import datetime, timedelta
from auth0_client import Auth0Client
from token_validator import TokenValidator

class TokenRotationManager:
    """
    Manage refresh token rotation with security monitoring
    """

    def __init__(self, auth0_client: Auth0Client, validator: TokenValidator):
        self.client = auth0_client
        self.validator = validator

    async def rotate_token(
        self,
        refresh_token: str,
        detect_reuse: bool = True
    ) -> Dict[str, Any]:
        """
        Rotate refresh token and detect reuse attacks

        Args:
            refresh_token: Current refresh token
            detect_reuse: Enable reuse detection

        Returns:
            New token pair
        """
        # Store old token for reuse detection
        old_token = refresh_token

        # Get new tokens
        new_tokens = await self.client.refresh_token(refresh_token)

        # If reuse detection enabled, store old token hash
        if detect_reuse:
            await self._store_used_token(old_token)

        # Validate new tokens
        access_token_payload = await self.validator.validate_access_token(
            new_tokens["access_token"]
        )

        return {
            "access_token": new_tokens["access_token"],
            "refresh_token": new_tokens.get("refresh_token"),
            "id_token": new_tokens.get("id_token"),
            "token_type": new_tokens.get("token_type", "Bearer"),
            "expires_in": new_tokens.get("expires_in", 86400)
        }

    async def check_token_reuse(self, refresh_token: str) -> bool:
        """
        Check if refresh token has been reused (potential attack)

        Args:
            refresh_token: Refresh token to check

        Returns:
            True if token has been reused
        """
        # Implementation: Check against database of used tokens
        # If token is found in used tokens database, it's a reuse
        return False

    async def revoke_all_user_tokens(self, user_id: str) -> None:
        """
        Revoke all tokens for a user (after reuse attack detected)

        Args:
            user_id: User ID
        """
        # Implementation: Use Auth0 Management API to revoke all tokens
        # Also clear from local database
        pass

    async def _store_used_token(self, token: str) -> None:
        """Store used refresh token for reuse detection"""
        # Implementation: Store token hash in database
        # with expiration date matching token lifetime
        pass

    async def cleanup_expired_tokens(self) -> int:
        """
        Cleanup expired tokens from database

        Returns:
            Number of tokens cleaned up
        """
        # Implementation: Remove expired tokens from database
        return 0
```

### 5.2 Token Storage (Secure Patterns)

```python
# token_storage.py
from typing import Optional, Dict, Any
from datetime import datetime, timedelta
import jwt

class SecureTokenStorage:
    """
    Secure token storage with encryption

    Note: This is a simplified example. For production:
    - Use proper encryption libraries (cryptography)
    - Store encryption keys securely (AWS KMS, HashiCorp Vault)
    - Use secure session storage for SPAs
    - Use platform keystores for mobile (Keychain, Keystore)
    """

    def __init__(self, encryption_key: bytes):
        self.encryption_key = encryption_key
        self.storage: Dict[str, Any] = {}

    async def store_tokens(
        self,
        user_id: str,
        access_token: str,
        refresh_token: Optional[str] = None,
        id_token: Optional[str] = None
    ) -> None:
        """
        Securely store tokens

        Args:
            user_id: User ID
            access_token: Access token
            refresh_token: Optional refresh token
            id_token: Optional ID token
        """
        # Decode access token to get expiration
        payload = jwt.decode(access_token, options={"verify_signature": False})
        expires_at = datetime.fromtimestamp(payload["exp"])

        token_data = {
            "access_token": self._encrypt_token(access_token),
            "refresh_token": self._encrypt_token(refresh_token) if refresh_token else None,
            "id_token": self._encrypt_token(id_token) if id_token else None,
            "expires_at": expires_at.isoformat(),
            "issued_at": datetime.utcnow().isoformat()
        }

        self.storage[user_id] = token_data

    async def get_tokens(self, user_id: str) -> Optional[Dict[str, str]]:
        """
        Retrieve stored tokens

        Args:
            user_id: User ID

        Returns:
            Dictionary with tokens or None if not found
        """
        token_data = self.storage.get(user_id)

        if not token_data:
            return None

        # Check if expired
        expires_at = datetime.fromisoformat(token_data["expires_at"])
        if datetime.utcnow() >= expires_at:
            await self.delete_tokens(user_id)
            return None

        return {
            "access_token": self._decrypt_token(token_data["access_token"]),
            "refresh_token": self._decrypt_token(token_data["refresh_token"]) if token_data["refresh_token"] else None,
            "id_token": self._decrypt_token(token_data["id_token"]) if token_data["id_token"] else None
        }

    async def delete_tokens(self, user_id: str) -> None:
        """
        Delete stored tokens

        Args:
            user_id: User ID
        """
        if user_id in self.storage:
            del self.storage[user_id]

    def _encrypt_token(self, token: str) -> str:
        """Encrypt token (placeholder - use proper encryption)"""
        # In production, use AES-256-GCM or similar
        return f"encrypted:{token}"

    def _decrypt_token(self, encrypted_token: str) -> str:
        """Decrypt token (placeholder - use proper decryption)"""
        # In production, use AES-256-GCM or similar
        return encrypted_token.replace("encrypted:", "")
```

---

## 6. DPoP Implementation

### 6.1 DPoP Client (Python)

```python
# dpop_client.py
import jwt
import time
import hashlib
from typing import Dict, Any, Optional
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.backends import default_backend
from auth0_client import Auth0Client

class DPoPClient:
    """
    Demonstrating Proof-of-Possession (DPoP) client implementation
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client
        self.private_key: Optional[ec.EllipticCurvePrivateKey] = None
        self.public_key: Optional[ec.EllipticCurvePublicKey] = None
        self.jwk: Optional[Dict[str, Any]] = None
        self.nonce_cache: Dict[str, str] = {}

    def generate_key_pair(self) -> None:
        """
        Generate DPoP key pair (ES256)

        In production:
        - Generate once per installation
        - Store private key securely
        - Reuse across sessions
        """
        # Generate ES256 key pair
        self.private_key = ec.generate_private_key(ec.SECP256R1(), default_backend())
        self.public_key = self.private_key.public_key()

        # Create JWK from public key
        public_numbers = self.public_key.public_numbers()
        self.jwk = {
            "kty": "EC",
            "crv": "P-256",
            "x": self._base64url_encode(public_numbers.x.to_bytes(32, 'big')),
            "y": self._base64url_encode(public_numbers.y.to_bytes(32, 'big'))
        }

    def create_dpop_proof(
        self,
        http_method: str,
        http_uri: str,
        access_token: Optional[str] = None,
        nonce: Optional[str] = None
    ) -> str:
        """
        Create DPoP proof JWT

        Args:
            http_method: HTTP method (GET, POST, etc.)
            http_uri: Full HTTP URI
            access_token: Optional access token for ath claim
            nonce: Optional nonce from server

        Returns:
            DPoP proof JWT
        """
        if not self.private_key:
            raise ValueError("Key pair not generated. Call generate_key_pair() first.")

        # Calculate ath claim if access token provided
        ath = None
        if access_token:
            ath = self._calculate_access_token_hash(access_token)

        # Create payload
        now = int(time.time())
        payload = {
            "jti": self._generate_jti(),
            "htm": http_method.upper(),
            "htu": http_uri,
            "iat": now
        }

        if ath:
            payload["ath"] = ath

        if nonce:
            payload["nonce"] = nonce

        # Create header
        header = {
            "typ": "dpop+jwt",
            "alg": "ES256",
            "jwk": self.jwk
        }

        # Sign JWT
        proof = jwt.encode(
            payload,
            self.private_key,
            algorithm="ES256",
            headers=header
        )

        return proof

    async def get_token_with_dpop(self, code: str, redirect_uri: str) -> Dict[str, Any]:
        """
        Get tokens with DPoP binding

        Args:
            code: Authorization code
            redirect_uri: Redirect URI

        Returns:
            Token response with DPoP-bound tokens
        """
        token_endpoint = f"https://{self.client.domain}/oauth/token"

        # Create DPoP proof for token request
        dpop_proof = self.create_dpop_proof("POST", token_endpoint)

        # Request headers
        headers = {
            "DPoP": dpop_proof,
            "Content-Type": "application/json"
        }

        payload = {
            "grant_type": "authorization_code",
            "client_id": self.client.client_id,
            "code": code,
            "redirect_uri": redirect_uri
        }

        if self.client.client_secret:
            payload["client_secret"] = self.client.client_secret

        response = await self.client.http_client.post(token_endpoint, json=payload, headers=headers)

        # Handle nonce requirement
        if response.status_code == 400 and "use_dpop_nonce" in response.text:
            nonce = response.headers.get("DPoP-Nonce")

            if nonce:
                # Retry with nonce
                dpop_proof = self.create_dpop_proof("POST", token_endpoint, nonce=nonce)
                headers["DPoP"] = dpop_proof
                response = await self.client.http_client.post(token_endpoint, json=payload, headers=headers)

        response.raise_for_status()
        return response.json()

    async def call_api_with_dpop(
        self,
        access_token: str,
        api_endpoint: str,
        method: str = "GET"
    ) -> Any:
        """
        Call API with DPoP-bound token

        Args:
            access_token: DPoP-bound access token
            api_endpoint: API endpoint to call
            method: HTTP method

        Returns:
            API response
        """
        # Create DPoP proof with ath claim
        dpop_proof = self.create_dpop_proof(method, api_endpoint, access_token)

        headers = {
            "Authorization": f"DPoP {access_token}",
            "DPoP": dpop_proof
        }

        response = await self.client.http_client.request(method, api_endpoint, headers=headers)
        response.raise_for_status()

        return response.json()

    def _calculate_access_token_hash(self, access_token: str) -> str:
        """Calculate SHA-256 hash of access token for ath claim"""
        hash_bytes = hashlib.sha256(access_token.encode()).digest()
        return self._base64url_encode(hash_bytes)

    def _generate_jti(self) -> str:
        """Generate unique JTI for replay prevention"""
        import uuid
        return str(uuid.uuid4())

    def _base64url_encode(self, data: bytes) -> str:
        """Base64url encode without padding"""
        import base64
        return base64.urlsafe_b64encode(data).rstrip(b'=').decode('utf-8')

    def export_private_key_pem(self) -> str:
        """Export private key as PEM (for storage)"""
        if not self.private_key:
            raise ValueError("No private key")

        pem = self.private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        )

        return pem.decode('utf-8')

    def import_private_key_pem(self, pem: str) -> None:
        """Import private key from PEM"""
        from cryptography.hazmat.primitives.serialization import load_pem_private_key

        self.private_key = load_pem_private_key(
            pem.encode(),
            password=None,
            backend=default_backend()
        )
        self.public_key = self.private_key.public_key()

        # Recreate JWK
        public_numbers = self.public_key.public_numbers()
        self.jwk = {
            "kty": "EC",
            "crv": "P-256",
            "x": self._base64url_encode(public_numbers.x.to_bytes(32, 'big')),
            "y": self._base64url_encode(public_numbers.y.to_bytes(32, 'big'))
        }
```

---

## 7. mTLS Implementation

### 7.1 mTLS Client (Python)

```python
# mtls_client.py
import httpx
from typing import Dict, Any, Optional
from auth0_client import Auth0Client

class mTLSClient:
    """
    mTLS (mutual TLS) client for Auth0 token binding

    Note: mTLS requires:
    - Auth0 Enterprise Plan with HRI add-on
    - X.509 certificates
    - Confidential client
    """

    def __init__(
        self,
        auth0_client: Auth0Client,
        cert_path: str,
        key_path: str,
        ca_path: Optional[str] = None
    ):
        self.client = auth0_client
        self.cert_path = cert_path
        self.key_path = key_path
        self.ca_path = ca_path

        # Create HTTP client with mTLS
        self.http_client = httpx.AsyncClient(
            cert=(cert_path, key_path),
            verify=ca_path if ca_path else True
        )

    async def get_token_with_mtls(self, code: str, redirect_uri: str) -> Dict[str, Any]:
        """
        Get tokens with mTLS binding

        The certificate thumbprint will be embedded in the token's cnf claim

        Args:
            code: Authorization code
            redirect_uri: Redirect URI

        Returns:
            Token response with mTLS-bound tokens
        """
        token_endpoint = f"https://{self.client.domain}/oauth/token"

        payload = {
            "grant_type": "authorization_code",
            "client_id": self.client.client_id,
            "code": code,
            "redirect_uri": redirect_uri
        }

        # mTLS authentication (no client secret needed)
        # The certificate is used for client authentication

        response = await self.http_client.post(token_endpoint, json=payload)
        response.raise_for_status()

        tokens = response.json()

        # Token will contain cnf claim with x5t#S256 (certificate thumbprint)
        return tokens

    async def call_api_with_mtls(
        self,
        access_token: str,
        api_endpoint: str,
        method: str = "GET"
    ) -> Any:
        """
        Call API with mTLS-bound token

        The resource server will validate:
        1. Token signature
        2. Certificate thumbprint in cnf.x5t#S256 matches client certificate

        Args:
            access_token: mTLS-bound access token
            api_endpoint: API endpoint
            method: HTTP method

        Returns:
            API response
        """
        headers = {
            "Authorization": f"Bearer {access_token}"
        }

        response = await self.http_client.request(method, api_endpoint, headers=headers)
        response.raise_for_status()

        return response.json()

    def calculate_certificate_thumbprint(self) -> str:
        """
        Calculate SHA-256 thumbprint of certificate

        This should match the cnf.x5t#S256 claim in the token

        Returns:
            Base64url-encoded certificate thumbprint
        """
        from cryptography import x509
        from cryptography.hazmat.backends import default_backend
        import hashlib
        import base64

        # Load certificate
        with open(self.cert_path, 'rb') as f:
            cert_data = f.read()

        cert = x509.load_pem_x509_certificate(cert_data, default_backend())

        # Calculate thumbprint
        thumbprint = hashlib.sha256(cert.public_bytes()).digest()

        # Base64url encode
        return base64.urlsafe_b64encode(thumbprint).rstrip(b'=').decode('utf-8')

    async def close(self):
        """Close HTTP client"""
        await self.http_client.aclose()
```

---

## 8. GDPR Compliance

### 8.1 User Data Management (Python)

```python
# gdpr_handler.py
from typing import Dict, Any, Optional
from datetime import datetime
from auth0_client import Auth0Client

class GDPRHandler:
    """
    GDPR compliance handler for user data rights

    Implements:
    - Right to access (data export)
    - Right to portability (structured data export)
    - Right to erasure (account deletion)
    - Right to rectification (data correction)
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client
        self.base_url = f"https://{auth0_client.domain}/api/v2"

    async def export_user_data(self, user_id: str) -> Dict[str, Any]:
        """
        Export user data (Right to Access and Portability)

        Args:
            user_id: User ID

        Returns:
            User data in structured, machine-readable format
        """
        # Get user profile from Auth0
        user_data = await self._get_auth0_user_data(user_id)

        # Get user activity logs (last 90 days per GDPR)
        activity_logs = await self._get_user_activity_logs(user_id)

        # Get consent records
        consents = await self._get_user_consents(user_id)

        # Compile export
        export_data = {
            "user_id": user_id,
            "export_date": datetime.utcnow().isoformat(),
            "user_profile": user_data,
            "activity_logs": activity_logs,
            "consents": consents
        }

        return export_data

    async def delete_user_account(self, user_id: str) -> Dict[str, Any]:
        """
        Delete user account and all associated data (Right to Erasure)

        Args:
            user_id: User ID

        Returns:
            Deletion confirmation
        """
        # Delete from Auth0
        await self._delete_auth0_user(user_id)

        # Delete from application database
        await self._delete_application_data(user_id)

        # Log deletion event
        await self._log_deletion_event(user_id)

        return {
            "user_id": user_id,
            "deleted": True,
            "timestamp": datetime.utcnow().isoformat()
        }

    async def update_user_data(
        self,
        user_id: str,
        updates: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Update user data (Right to Rectification)

        Args:
            user_id: User ID
            updates: Data updates

        Returns:
            Updated user data
        """
        # Update in Auth0
        updated_user = await self._update_auth0_user(user_id, updates)

        # Update in application database
        await self._update_application_data(user_id, updates)

        return {
            "user_id": user_id,
            "updated": True,
            "timestamp": datetime.utcnow().isoformat(),
            "data": updated_user
        }

    async def record_consent(
        self,
        user_id: str,
        consent_type: str,
        granted: bool,
        purpose: str
    ) -> None:
        """
        Record user consent

        Args:
            user_id: User ID
            consent_type: Type of consent (e.g., "marketing", "analytics")
            granted: Whether consent was granted
            purpose: Purpose of data processing
        """
        consent_record = {
            "user_id": user_id,
            "consent_type": consent_type,
            "granted": granted,
            "purpose": purpose,
            "timestamp": datetime.utcnow().isoformat(),
            "valid_until": None  # Or set expiration if applicable
        }

        # Store in your consent management database
        await self._store_consent(consent_record)

    async def get_user_consents(self, user_id: str) -> list:
        """
        Get all user consents

        Args:
            user_id: User ID

        Returns:
            List of consent records
        """
        return await self._get_user_consents(user_id)

    async def _get_auth0_user_data(self, user_id: str) -> Dict[str, Any]:
        """Get user profile from Auth0"""
        user_endpoint = f"{self.base_url}/users/{user_id}"

        response = await self.client.http_client.get(
            user_endpoint,
            headers={"Authorization": f"Bearer {await self._get_management_token()}"}
        )
        response.raise_for_status()

        return response.json()

    async def _get_user_activity_logs(self, user_id: str) -> list:
        """Get user activity logs"""
        # Implementation: Query Auth0 logs for user activity
        # Limit to last 90 days per GDPR
        return []

    async def _delete_auth0_user(self, user_id: str) -> None:
        """Delete user from Auth0"""
        user_endpoint = f"{self.base_url}/users/{user_id}"

        await self.client.http_client.delete(
            user_endpoint,
            headers={"Authorization": f"Bearer {await self._get_management_token()}"}
        )

    async def _delete_application_data(self, user_id: str) -> None:
        """Delete user data from application database"""
        # Implementation: Delete from your application database
        pass

    async def _update_auth0_user(
        self,
        user_id: str,
        updates: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Update user in Auth0"""
        user_endpoint = f"{self.base_url}/users/{user_id}"

        response = await self.client.http_client.patch(
            user_endpoint,
            json=updates,
            headers={"Authorization": f"Bearer {await self._get_management_token()}"}
        )
        response.raise_for_status()

        return response.json()

    async def _update_application_data(
        self,
        user_id: str,
        updates: Dict[str, Any]
    ) -> None:
        """Update user data in application database"""
        # Implementation: Update your application database
        pass

    async def _store_consent(self, consent_record: Dict[str, Any]) -> None:
        """Store consent record in database"""
        # Implementation: Store in your consent management database
        pass

    async def _get_user_consents(self, user_id: str) -> list:
        """Get user consents from database"""
        # Implementation: Query your consent management database
        return []

    async def _log_deletion_event(self, user_id: str) -> None:
        """Log account deletion event"""
        # Implementation: Log to audit system
        pass

    async def _get_management_token(self) -> str:
        """Get Management API token"""
        # Implementation: Use client credentials flow
        return ""
```

---

## 9. Error Handling and Retry Logic

### 9.1 Retry Handler (Python)

```python
# retry_handler.py
import asyncio
import logging
from typing import Callable, Any, Optional
from datetime import datetime
import httpx

logger = logging.getLogger(__name__)

class Auth0RetryHandler:
    """
    Retry handler for Auth0 API calls with exponential backoff
    """

    def __init__(
        self,
        max_retries: int = 3,
        base_delay: float = 1.0,
        max_delay: float = 32.0,
        exponential_base: float = 2.0
    ):
        self.max_retries = max_retries
        self.base_delay = base_delay
        self.max_delay = max_delay
        self.exponential_base = exponential_base

    async def execute_with_retry(
        self,
        func: Callable,
        *args: Any,
        **kwargs: Any
    ) -> Any:
        """
        Execute function with retry logic

        Args:
            func: Function to execute
            *args: Function arguments
            **kwargs: Function keyword arguments

        Returns:
            Function result

        Raises:
            Exception: If all retries exhausted
        """
        last_exception = None

        for attempt in range(self.max_retries):
            try:
                return await func(*args, **kwargs)

            except httpx.HTTPStatusError as e:
                last_exception = e

                # Don't retry client errors (4xx)
                if 400 <= e.response.status_code < 500:
                    logger.error(f"Client error {e.response.status_code}: {e}")
                    raise

                # Retry on server errors (5xx) and rate limiting (429)
                if e.response.status_code >= 500 or e.response.status_code == 429:
                    delay = self._calculate_delay(attempt)

                    logger.warning(
                        f"Attempt {attempt + 1} failed with status {e.response.status_code}. "
                        f"Retrying in {delay:.2f} seconds..."
                    )

                    await asyncio.sleep(delay)
                    continue

                raise

            except (httpx.RequestError, httpx.TimeoutException) as e:
                last_exception = e

                delay = self._calculate_delay(attempt)

                logger.warning(
                    f"Attempt {attempt + 1} failed with network error. "
                    f"Retrying in {delay:.2f} seconds..."
                )

                await asyncio.sleep(delay)
                continue

            except Exception as e:
                # Don't retry on other exceptions
                logger.error(f"Unexpected error: {e}")
                raise

        # All retries exhausted
        logger.error(f"All {self.max_retries} retries exhausted")
        raise last_exception

    def _calculate_delay(self, attempt: int) -> float:
        """
        Calculate delay with exponential backoff

        Args:
            attempt: Attempt number (0-indexed)

        Returns:
            Delay in seconds
        """
        delay = self.base_delay * (self.exponential_base ** attempt)
        return min(delay, self.max_delay)

class Auth0ErrorHandler:
    """
    Centralized error handling for Auth0 operations
    """

    @staticmethod
    def handle_auth_error(error: Exception) -> Dict[str, Any]:
        """
        Handle authentication errors

        Args:
            error: Exception

        Returns:
            Error response dictionary
        """
        if isinstance(error, httpx.HTTPStatusError):
            status_code = error.response.status_code

            if status_code == 401:
                return {
                    "error": "invalid_token",
                    "error_description": "Token is invalid or expired",
                    "status_code": status_code
                }

            elif status_code == 403:
                return {
                    "error": "insufficient_scope",
                    "error_description": "Token lacks required scope",
                    "status_code": status_code
                }

            elif status_code == 429:
                return {
                    "error": "too_many_requests",
                    "error_description": "Rate limit exceeded",
                    "status_code": status_code
                }

            elif status_code >= 500:
                return {
                    "error": "server_error",
                    "error_description": "Auth0 server error",
                    "status_code": status_code
                }

        return {
            "error": "unknown_error",
            "error_description": str(error),
            "status_code": 500
        }

    @staticmethod
    def should_retry(error: Exception) -> bool:
        """
        Determine if error is retryable

        Args:
            error: Exception

        Returns:
            True if error is retryable
        """
        if isinstance(error, httpx.HTTPStatusError):
            status_code = error.response.status_code

            # Retry on server errors and rate limiting
            return status_code >= 500 or status_code == 429

        if isinstance(error, (httpx.RequestError, httpx.TimeoutException)):
            return True

        return False
```

### 9.2 Logging and Monitoring (Python)

```python
# security_logger.py
import logging
import json
from typing import Dict, Any, Optional
from datetime import datetime
from enum import Enum

class SecurityEventType(Enum):
    """Security event types"""
    LOGIN_SUCCESS = "login_success"
    LOGIN_FAILURE = "login_failure"
    MFA_CHALLENGE = "mfa_challenge"
    MFA_SUCCESS = "mfa_success"
    MFA_FAILURE = "mfa_failure"
    TOKEN_ISSUED = "token_issued"
    TOKEN_REFRESH = "token_refresh"
    TOKEN_REVOKED = "token_revoked"
    ATTACK_DETECTED = "attack_detected"
    BREACHED_PASSWORD = "breached_password"
    ACCOUNT_LOCKED = "account_locked"
    DATA_EXPORT = "data_export"
    DATA_DELETION = "data_deletion"

class SecurityLogger:
    """
    Centralized security event logging

    Logs security-relevant events for audit and compliance
    """

    def __init__(self, log_level: str = "INFO"):
        self.logger = logging.getLogger("auth0_security")
        self.logger.setLevel(getattr(logging, log_level.upper()))

        # Configure handler (in production, send to SIEM)
        handler = logging.StreamHandler()
        handler.setFormatter(
            logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
        )
        self.logger.addHandler(handler)

    def log_security_event(
        self,
        event_type: SecurityEventType,
        user_id: Optional[str] = None,
        ip_address: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None
    ) -> None:
        """
        Log security event

        Args:
            event_type: Type of security event
            user_id: User ID (if applicable)
            ip_address: IP address (if applicable)
            details: Additional event details
        """
        event = {
            "event_type": event_type.value,
            "timestamp": datetime.utcnow().isoformat(),
            "user_id": user_id,
            "ip_address": ip_address,
            "details": details or {}
        }

        # Log as structured JSON
        self.logger.info(json.dumps(event))

    def log_login_success(
        self,
        user_id: str,
        ip_address: str,
        mfa_used: bool = False
    ) -> None:
        """Log successful login"""
        self.log_security_event(
            SecurityEventType.LOGIN_SUCCESS,
            user_id=user_id,
            ip_address=ip_address,
            details={"mfa_used": mfa_used}
        )

    def log_login_failure(
        self,
        email: str,
        ip_address: str,
        reason: str
    ) -> None:
        """Log failed login attempt"""
        self.log_security_event(
            SecurityEventType.LOGIN_FAILURE,
            ip_address=ip_address,
            details={"email": email, "reason": reason}
        )

    def log_mfa_challenge(
        self,
        user_id: str,
        mfa_method: str
    ) -> None:
        """Log MFA challenge"""
        self.log_security_event(
            SecurityEventType.MFA_CHALLENGE,
            user_id=user_id,
            details={"mfa_method": mfa_method}
        )

    def log_attack_detected(
        self,
        attack_type: str,
        ip_address: str,
        target_user_id: Optional[str] = None
    ) -> None:
        """Log detected attack"""
        self.log_security_event(
            SecurityEventType.ATTACK_DETECTED,
            ip_address=ip_address,
            user_id=target_user_id,
            details={"attack_type": attack_type}
        )

    def log_breached_password(
        self,
        user_id: str,
        ip_address: str
    ) -> None:
        """Log breached password detection"""
        self.log_security_event(
            SecurityEventType.BREACHED_PASSWORD,
            user_id=user_id,
            ip_address=ip_address
        )

    def log_token_issued(
        self,
        user_id: str,
        token_type: str,
        scopes: list
    ) -> None:
        """Log token issuance"""
        self.log_security_event(
            SecurityEventType.TOKEN_ISSUED,
            user_id=user_id,
            details={"token_type": token_type, "scopes": scopes}
        )

    def log_data_export(
        self,
        user_id: str,
        requested_by: str
    ) -> None:
        """Log GDPR data export"""
        self.log_security_event(
            SecurityEventType.DATA_EXPORT,
            user_id=user_id,
            details={"requested_by": requested_by}
        )

    def log_data_deletion(
        self,
        user_id: str,
        deleted_by: str
    ) -> None:
        """Log GDPR data deletion"""
        self.log_security_event(
            SecurityEventType.DATA_DELETION,
            user_id=user_id,
            details={"deleted_by": deleted_by}
        )
```

---

## 10. Security Monitoring

### 10.1 Metrics Collector (Python)

```python
# metrics_collector.py
from typing import Dict, Any, List
from datetime import datetime, timedelta
from collections import defaultdict
from auth0_client import Auth0Client

class SecurityMetricsCollector:
    """
    Collect and analyze security metrics

    Tracks:
    - Authentication success/failure rates
    - MFA adoption and success rates
    - Attack detection rates
    - Token issuance patterns
    - Breached password detections
    """

    def __init__(self, auth0_client: Auth0Client):
        self.client = auth0_client
        self.metrics_cache: Dict[str, Any] = defaultdict(list)

    async def collect_metrics(
        self,
        time_range: timedelta = timedelta(hours=24)
    ) -> Dict[str, Any]:
        """
        Collect all security metrics

        Args:
            time_range: Time range for metrics

        Returns:
            Dictionary with all metrics
        """
        end_time = datetime.utcnow()
        start_time = end_time - time_range

        metrics = {
            "authentication": await self._collect_authentication_metrics(
                start_time, end_time
            ),
            "mfa": await self._collect_mfa_metrics(start_time, end_time),
            "attacks": await self._collect_attack_metrics(start_time, end_time),
            "tokens": await self._collect_token_metrics(start_time, end_time),
            "compliance": await self._collect_compliance_metrics(start_time, end_time)
        }

        return metrics

    async def _collect_authentication_metrics(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Collect authentication metrics"""
        # Implementation: Query Auth0 logs for authentication events

        return {
            "total_logins": 0,
            "successful_logins": 0,
            "failed_logins": 0,
            "success_rate": 0.0,
            "unique_users": 0,
            "unique_ips": 0
        }

    async def _collect_mfa_metrics(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Collect MFA metrics"""
        # Implementation: Query Auth0 logs for MFA events

        return {
            "mfa_challenges": 0,
            "mfa_successes": 0,
            "mfa_failures": 0,
            "success_rate": 0.0,
            "adoption_rate": 0.0,
            "by_factor": {
                "otp": 0,
                "push": 0,
                "webauthn": 0,
                "sms": 0
            }
        }

    async def _collect_attack_metrics(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Collect attack metrics"""
        # Implementation: Query Auth0 Security Center

        return {
            "bot_detection": {
                "blocked": 0,
                "challenged": 0
            },
            "brute_force": {
                "blocked_attempts": 0,
                "locked_accounts": 0
            },
            "breached_passwords": {
                "detected": 0,
                "blocked": 0
            },
            "suspicious_ip": {
                "throttled": 0,
                "blocked": 0
            }
        }

    async def _collect_token_metrics(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Collect token metrics"""
        # Implementation: Query token issuance and refresh patterns

        return {
            "tokens_issued": 0,
            "tokens_refreshed": 0,
            "tokens_revoked": 0,
            "refresh_rate": 0.0,
            "avg_lifetime": 0.0
        }

    async def _collect_compliance_metrics(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Collect compliance metrics"""
        # Implementation: Query GDPR and compliance events

        return {
            "gdpr": {
                "data_exports": 0,
                "data_deletions": 0,
                "consent_records": 0
            },
            "audit_logs": {
                "total_events": 0,
                "retention_compliance": 100.0
            }
        }

    async def generate_report(self) -> str:
        """
        Generate security metrics report

        Returns:
            Formatted report string
        """
        metrics = await self.collect_metrics()

        report = f"""
# Auth0 Security Metrics Report
Generated: {datetime.utcnow().isoformat()}

## Authentication
- Total Logins: {metrics['authentication']['total_logins']}
- Success Rate: {metrics['authentication']['success_rate']:.2%}
- Unique Users: {metrics['authentication']['unique_users']}

## MFA
- MFA Adoption: {metrics['mfa']['adoption_rate']:.2%}
- Success Rate: {metrics['mfa']['success_rate']:.2%}
- Challenges: {metrics['mfa']['mfa_challenges']}

## Attack Protection
- Bot Detection: {metrics['attacks']['bot_detection']['blocked']} blocked
- Brute Force: {metrics['attacks']['brute_force']['blocked_attempts']} blocked attempts
- Breached Passwords: {metrics['attacks']['breached_passwords']['detected']} detected
- Suspicious IP: {metrics['attacks']['suspicious_ip']['throttled']} throttled

## Tokens
- Issued: {metrics['tokens']['tokens_issued']}
- Refreshed: {metrics['tokens']['tokens_refreshed']}
- Revoked: {metrics['tokens']['tokens_revoked']}

## Compliance
- Data Exports: {metrics['compliance']['gdpr']['data_exports']}
- Data Deletions: {metrics['compliance']['gdpr']['data_deletions']}
"""

        return report
```

---

## Usage Examples

### Complete Authentication Flow

```python
# example_auth_flow.py
import asyncio
from config import Auth0Config, SecurityConfig
from auth0_client import Auth0Client
from token_validator import TokenValidator
from token_rotation import TokenRotationManager
from security_logger import SecurityLogger

async def authenticate_user(code: str, redirect_uri: str):
    """Complete authentication flow with security features"""

    # Initialize components
    auth0_config = Auth0Config.from_env()
    security_config = SecurityConfig.from_env()

    auth0_client = Auth0Client(auth0_config, security_config)
    token_validator = TokenValidator(auth0_config)
    token_manager = TokenRotationManager(auth0_client, token_validator)
    security_logger = SecurityLogger()

    try:
        # Exchange authorization code for tokens
        tokens = await auth0_client.get_token(code, redirect_uri)

        # Validate access token
        payload = await token_validator.validate_access_token(tokens["access_token"])

        # Get user info
        user_info = await auth0_client.get_user_info(tokens["access_token"])

        # Log successful login
        security_logger.log_login_success(
            user_id=payload["sub"],
            ip_address="user_ip_address",  # Get from request
            mfa_used="amr" in payload  # amr = Authentication Methods References
        )

        return {
            "user": user_info,
            "tokens": tokens
        }

    finally:
        await auth0_client.close()

# Run example
asyncio.run(authenticate_user("code_from_callback", "https://yourapp.com/callback"))
```

---

## Testing Examples

### Test Breached Password Detection

```python
# test_breached_password.py
import asyncio
from config import Auth0Config, SecurityConfig
from auth0_client import Auth0Client
from breached_password import BreachedPasswordDetector

async def test_breached_password():
    """Test breached password detection"""

    auth0_config = Auth0Config.from_env()
    security_config = SecurityConfig.from_env()

    auth0_client = Auth0Client(auth0_config, security_config)
    detector = BreachedPasswordDetector(auth0_client)

    # Use test password that always triggers detection
    test_password = detector.test_breached_password()

    print(f"Test password: {test_password}")
    print("This password will trigger Auth0's breached password detection")

    await auth0_client.close()

asyncio.run(test_breached_password())
```

---

## Best Practices Summary

1. **Always use HTTPS** for all token transmission
2. **Validate tokens** on every request (signature, expiration, issuer, audience)
3. **Enable refresh token rotation** to detect token theft
4. **Use RS256** algorithm for tokens (asymmetric, more secure)
5. **Implement MFA** for sensitive operations
6. **Monitor security events** with centralized logging
7. **Follow GDPR requirements** for user data rights
8. **Use DPoP or mTLS** for enhanced token security
9. **Configure attack protection** with appropriate thresholds
10. **Regularly audit** security configurations and access patterns

---

For more details, see:
- [Attack Protection Overview](modules/attack-protection-overview.md)
- [MFA Implementation](modules/mfa-overview.md)
- [Token Best Practices](modules/token-best-practices.md)
- [DPoP Implementation](modules/dpop-implementation.md)
- [GDPR Compliance](modules/gdpr-compliance.md)
