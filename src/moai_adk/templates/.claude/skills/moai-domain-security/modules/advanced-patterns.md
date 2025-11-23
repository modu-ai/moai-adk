# Advanced Security Patterns

This module contains advanced enterprise security patterns, OWASP Top 10 2021 protection strategies, zero-trust architecture, threat modeling, and DevSecOps automation for production-ready applications.

---

## Overview


**Purpose**: Advanced security patterns for OWASP compliance, zero-trust architecture, and enterprise-grade threat protection.

**Key Topics**:
- OWASP Top 10 2021 advanced mitigation patterns
- Zero-trust authentication and authorization
- Automated threat modeling (STRIDE, PASTA)
- DevSecOps pipeline integration
- Cryptography and secrets management
- Security testing automation (SAST, DAST, IAST)
- Incident response and forensics
- Compliance automation (SOC 2, ISO 27001, GDPR)

**Target Audience**: Security engineers, DevSecOps specialists, security architects

---

## 1. Advanced OWASP Top 10 2021 Protection

**Concept**: Comprehensive protection against critical web application vulnerabilities with layered defense strategies.

### 1.1 Multi-Layer Injection Prevention (A03)

```python
# Advanced injection attack prevention with multiple validation layers
class InjectionDefenseMiddleware:
    """Multi-layer protection against injection attacks."""

    def __init__(self, app):
        """
        Initialize injection defense middleware.

        Args:
            app: WSGI/ASGI application instance
        """
        self.app = app
        self.sql_injection_patterns = self._compile_sql_patterns()
        self.xss_patterns = self._compile_xss_patterns()
        self.command_injection_patterns = self._compile_command_patterns()

    def _compile_sql_patterns(self) -> List[re.Pattern]:
        """Compile SQL injection detection patterns."""
        return [
            re.compile(r"(\bunion\b.*\bselect\b)", re.IGNORECASE),
            re.compile(r"(\bor\b\s+\d+\s*=\s*\d+)", re.IGNORECASE),
            re.compile(r"(;\s*drop\s+table)", re.IGNORECASE),
            re.compile(r"(--|\#|\/\*)", re.IGNORECASE),
            re.compile(r"(\bexec\b|\bexecute\b)", re.IGNORECASE)
        ]

    def _compile_xss_patterns(self) -> List[re.Pattern]:
        """Compile XSS attack detection patterns."""
        return [
            re.compile(r"<script[^>]*>.*?</script>", re.IGNORECASE | re.DOTALL),
            re.compile(r"javascript:", re.IGNORECASE),
            re.compile(r"on\w+\s*=", re.IGNORECASE),
            re.compile(r"<iframe", re.IGNORECASE),
            re.compile(r"<object", re.IGNORECASE)
        ]

    def _compile_command_patterns(self) -> List[re.Pattern]:
        """Compile command injection detection patterns."""
        return [
            re.compile(r"[;&|`$()]"),
            re.compile(r"(wget|curl|nc|bash|sh|python|perl|ruby)", re.IGNORECASE)
        ]

    def validate_input(self, value: str, input_type: str = "general") -> Tuple[bool, Optional[str]]:
        """
        Validate input against injection patterns.

        Args:
            value: Input value to validate
            input_type: Type of input (general, sql, xss, command)

        Returns:
            Tuple of (is_valid, threat_type)

        Example:
            >>> defense = InjectionDefenseMiddleware(app)
            >>> is_valid, threat = defense.validate_input("user@example.com", "email")
            >>> # Valid email: (True, None)
            >>> is_valid, threat = defense.validate_input("' OR 1=1 --", "sql")
            >>> # SQL injection detected: (False, "SQL_INJECTION")
        """
        # SQL Injection detection
        for pattern in self.sql_injection_patterns:
            if pattern.search(value):
                return False, "SQL_INJECTION"

        # XSS detection
        for pattern in self.xss_patterns:
            if pattern.search(value):
                return False, "XSS_ATTACK"

        # Command injection detection
        if input_type == "command":
            for pattern in self.command_injection_patterns:
                if pattern.search(value):
                    return False, "COMMAND_INJECTION"

        return True, None

    def sanitize_output(self, value: str, context: str = "html") -> str:
        """
        Sanitize output based on context.

        Args:
            value: Value to sanitize
            context: Output context (html, js, url, css)

        Returns:
            Sanitized output

        Example:
            >>> sanitized = defense.sanitize_output("<script>alert('XSS')</script>", "html")
            >>> # Output: "&lt;script&gt;alert('XSS')&lt;/script&gt;"
        """
        if context == "html":
            return html.escape(value)
        elif context == "js":
            return json.dumps(value)[1:-1]  # Remove quotes
        elif context == "url":
            return urllib.parse.quote(value)
        elif context == "css":
            # Remove potentially dangerous CSS
            return re.sub(r'[^a-zA-Z0-9\s\-_]', '', value)

        return value

    def parameterized_query_wrapper(self, query: str, params: Dict[str, Any]) -> Tuple[str, tuple]:
        """
        Enforce parameterized queries for database operations.

        Args:
            query: SQL query with placeholders
            params: Query parameters

        Returns:
            Tuple of (validated_query, sanitized_params)

        Example:
            >>> query, params = defense.parameterized_query_wrapper(
            ...     "SELECT * FROM users WHERE email = :email",
            ...     {"email": "user@example.com"}
            ... )
            >>> # Safe parameterized query execution
        """
        # Validate query structure
        if any(pattern.search(query) for pattern in self.sql_injection_patterns):
            raise SecurityError("Query contains potential SQL injection patterns")

        # Sanitize parameters
        sanitized_params = tuple(
            self._sanitize_sql_param(v) for v in params.values()
        )

        return query, sanitized_params

    def _sanitize_sql_param(self, param: Any) -> Any:
        """Sanitize individual SQL parameter."""
        if isinstance(param, str):
            # Remove SQL comment markers
            param = re.sub(r'(--|#|\/\*|\*\/)', '', param)
            # Escape single quotes
            param = param.replace("'", "''")

        return param
```

### 1.2 Advanced Authentication & Broken Access Control Prevention (A01)

```python
class ZeroTrustAccessControl:
    """Advanced access control with zero-trust principles."""

    def __init__(self, policy_engine):
        """
        Initialize zero-trust access control.

        Args:
            policy_engine: Policy decision point (PDP) engine
        """
        self.policy_engine = policy_engine
        self.risk_scorer = RiskScorer()

    def evaluate_access(
        self,
        user: User,
        resource: Resource,
        action: str,
        context: Dict[str, Any]
    ) -> AccessDecision:
        """
        Evaluate access request with zero-trust principles.

        Args:
            user: User requesting access
            resource: Resource being accessed
            action: Action to perform (read, write, delete)
            context: Request context (IP, device, location, time)

        Returns:
            Access decision with confidence score

        Example:
            >>> zta = ZeroTrustAccessControl(policy_engine)
            >>> decision = zta.evaluate_access(
            ...     user=current_user,
            ...     resource=sensitive_document,
            ...     action="read",
            ...     context={
            ...         "ip_address": "192.168.1.100",
            ...         "device_fingerprint": "abc123",
            ...         "time_of_day": "02:00",
            ...         "location": "unusual_country"
            ...     }
            ... )
            >>> # Decision: DENY (risk score: 85/100)
        """
        # Calculate risk score
        risk_score = self.risk_scorer.calculate_risk(user, resource, context)

        # Evaluate policy
        policy_decision = self.policy_engine.evaluate({
            "user_id": user.id,
            "user_roles": user.roles,
            "resource_id": resource.id,
            "resource_classification": resource.classification,
            "action": action,
            "risk_score": risk_score
        })

        # Apply step-up authentication for high-risk requests
        if risk_score > 70 and policy_decision == "ALLOW":
            return AccessDecision(
                decision="STEP_UP_REQUIRED",
                reason="High risk score requires additional authentication",
                risk_score=risk_score,
                mfa_required=True
            )

        # Apply time-based access restrictions
        if self._is_outside_business_hours(context.get("time_of_day")):
            if resource.classification in ["CONFIDENTIAL", "SECRET"]:
                return AccessDecision(
                    decision="DENY",
                    reason="Access to confidential resources restricted outside business hours",
                    risk_score=risk_score
                )

        return AccessDecision(
            decision=policy_decision,
            reason=f"Policy evaluation: {policy_decision}",
            risk_score=risk_score
        )

    def _is_outside_business_hours(self, time_str: str) -> bool:
        """Check if request is outside business hours (9 AM - 6 PM)."""
        if not time_str:
            return False

        hour = int(time_str.split(":")[0])
        return hour < 9 or hour >= 18


class RiskScorer:
    """Calculate risk scores for access requests."""

    def calculate_risk(
        self,
        user: User,
        resource: Resource,
        context: Dict[str, Any]
    ) -> int:
        """
        Calculate risk score (0-100) for access request.

        Risk Factors:
        - User behavior anomalies (30 points)
        - Device trust level (20 points)
        - Location anomalies (20 points)
        - Time of access (10 points)
        - Resource sensitivity (20 points)

        Returns:
            Risk score from 0 (low risk) to 100 (high risk)
        """
        risk_score = 0

        # User behavior anomalies
        if self._detect_behavior_anomaly(user, context):
            risk_score += 30

        # Device trust level
        device_trust = context.get("device_trust_score", 100)
        risk_score += int((100 - device_trust) * 0.2)

        # Location anomalies
        if self._detect_location_anomaly(user, context.get("location")):
            risk_score += 20

        # Time of access
        if self._is_unusual_time(user, context.get("time_of_day")):
            risk_score += 10

        # Resource sensitivity
        if resource.classification == "SECRET":
            risk_score += 20
        elif resource.classification == "CONFIDENTIAL":
            risk_score += 10

        return min(risk_score, 100)

    def _detect_behavior_anomaly(self, user: User, context: Dict[str, Any]) -> bool:
        """Detect anomalous user behavior patterns."""
        # Example: Detect unusual access patterns
        # In production: Use ML-based anomaly detection
        request_rate = context.get("requests_last_hour", 0)
        avg_rate = user.avg_hourly_requests or 10

        return request_rate > (avg_rate * 3)  # 3x normal rate

    def _detect_location_anomaly(self, user: User, location: Optional[str]) -> bool:
        """Detect location-based anomalies."""
        if not location or not user.usual_locations:
            return False

        return location not in user.usual_locations

    def _is_unusual_time(self, user: User, time_str: Optional[str]) -> bool:
        """Detect unusual access times."""
        if not time_str:
            return False

        hour = int(time_str.split(":")[0])

        # Outside typical working hours (9 AM - 6 PM)
        return hour < 9 or hour >= 18
```

---

## 2. Automated Threat Modeling

**Concept**: Systematic identification and mitigation of security threats using STRIDE and PASTA methodologies.

### 2.1 STRIDE Threat Modeling Automation

```python
class STRIDEThreatModeler:
    """Automate STRIDE threat modeling for system components."""

    THREAT_CATEGORIES = {
        "Spoofing": ["authentication", "identity"],
        "Tampering": ["data_integrity", "input_validation"],
        "Repudiation": ["logging", "audit_trail"],
        "Information_Disclosure": ["encryption", "access_control"],
        "Denial_of_Service": ["rate_limiting", "resource_management"],
        "Elevation_of_Privilege": ["authorization", "least_privilege"]
    }

    def __init__(self):
        """Initialize STRIDE threat modeler."""
        self.threat_database = self._load_threat_database()

    def analyze_component(self, component: SystemComponent) -> ThreatAnalysis:
        """
        Analyze component for STRIDE threats.

        Args:
            component: System component to analyze

        Returns:
            Threat analysis with identified threats and mitigations

        Example:
            >>> modeler = STRIDEThreatModeler()
            >>> component = SystemComponent(
            ...     name="Authentication API",
            ...     type="API",
            ...     data_flow=["user_credentials", "jwt_tokens"],
            ...     trust_boundary=True
            ... )
            >>> analysis = modeler.analyze_component(component)
            >>> # Identified threats:
            >>> # - Spoofing: Weak password policy (HIGH)
            >>> # - Tampering: Unvalidated JWT claims (MEDIUM)
            >>> # - Information Disclosure: Credentials in logs (HIGH)
        """
        identified_threats = []

        # Analyze each STRIDE category
