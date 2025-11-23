# Advanced Threat Modeling

STRIDE threat modeling automation and risk assessment.

        for category, keywords in self.THREAT_CATEGORIES.items():
            threats = self._identify_threats(component, category, keywords)
            identified_threats.extend(threats)

        # Prioritize threats by risk
        prioritized = self._prioritize_threats(identified_threats)

        # Generate mitigations
        mitigations = self._generate_mitigations(prioritized)

        return ThreatAnalysis(
            component=component.name,
            threats=prioritized,
            mitigations=mitigations,
            risk_score=self._calculate_risk_score(prioritized)
        )

    def _identify_threats(
        self,
        component: SystemComponent,
        category: str,
        keywords: List[str]
    ) -> List[Threat]:
        """Identify threats in specific STRIDE category."""
        threats = []

        # Check if component handles relevant data
        for keyword in keywords:
            if keyword in component.data_flow or keyword in component.type.lower():
                # Query threat database
                db_threats = self.threat_database.get(category, [])

                for threat_template in db_threats:
                    threat = Threat(
                        category=category,
                        description=threat_template["description"].format(
                            component=component.name
                        ),
                        severity=threat_template["severity"],
                        likelihood=self._assess_likelihood(component, threat_template),
                        impact=threat_template["impact"]
                    )
                    threats.append(threat)

        # Special checks for trust boundaries
        if component.trust_boundary:
            threats.append(Threat(
                category="Elevation_of_Privilege",
                description=f"{component.name} crosses trust boundary - validate all inputs",
                severity="HIGH",
                likelihood="MEDIUM",
                impact="Data breach or unauthorized access"
            ))

        return threats

    def _prioritize_threats(self, threats: List[Threat]) -> List[Threat]:
        """Prioritize threats by risk (severity × likelihood)."""
        severity_map = {"CRITICAL": 4, "HIGH": 3, "MEDIUM": 2, "LOW": 1}
        likelihood_map = {"VERY_HIGH": 4, "HIGH": 3, "MEDIUM": 2, "LOW": 1}

        def risk_score(threat):
            severity = severity_map.get(threat.severity, 1)
            likelihood = likelihood_map.get(threat.likelihood, 1)
            return severity * likelihood

        return sorted(threats, key=risk_score, reverse=True)

    def _generate_mitigations(self, threats: List[Threat]) -> List[Mitigation]:
        """Generate mitigations for identified threats."""
        mitigations = []

        for threat in threats:
            if threat.category == "Spoofing":
                mitigations.append(Mitigation(
                    threat_id=threat.id,
                    strategy="Implement multi-factor authentication (MFA)",
                    implementation="Use TOTP or WebAuthn for second factor",
                    validation="Verify MFA enrollment rate > 90%"
                ))

            elif threat.category == "Tampering":
                mitigations.append(Mitigation(
                    threat_id=threat.id,
                    strategy="Implement input validation and integrity checks",
                    implementation="Use JSON Schema validation + HMAC signatures",
                    validation="Automated tests for all input vectors"
                ))

            elif threat.category == "Information_Disclosure":
                mitigations.append(Mitigation(
                    threat_id=threat.id,
                    strategy="Encrypt sensitive data at rest and in transit",
                    implementation="AES-256-GCM at rest, TLS 1.3 in transit",
                    validation="Scan for unencrypted sensitive data"
                ))

            elif threat.category == "Denial_of_Service":
                mitigations.append(Mitigation(
                    threat_id=threat.id,
                    strategy="Implement rate limiting and resource quotas",
                    implementation="Token bucket algorithm with Redis",
                    validation="Load test with 10x expected traffic"
                ))

        return mitigations

    def _load_threat_database(self) -> Dict[str, List[Dict]]:
        """Load threat database with STRIDE patterns."""
        return {
            "Spoofing": [
                {
                    "description": "{component} may allow credential stuffing attacks",
                    "severity": "HIGH",
                    "impact": "Unauthorized access to user accounts"
                },
                {
                    "description": "{component} lacks MFA enforcement",
                    "severity": "MEDIUM",
                    "impact": "Single-factor authentication weakness"
                }
            ],
            "Tampering": [
                {
                    "description": "{component} accepts unvalidated input",
                    "severity": "HIGH",
                    "impact": "Data corruption or injection attacks"
                }
            ],
            "Information_Disclosure": [
                {
                    "description": "{component} may leak sensitive data in error messages",
                    "severity": "MEDIUM",
                    "impact": "Information leakage"
                },
                {
                    "description": "{component} stores credentials in plaintext",
                    "severity": "CRITICAL",
                    "impact": "Credential exposure"
                }
            ],
            "Denial_of_Service": [
                {
                    "description": "{component} lacks rate limiting",
                    "severity": "MEDIUM",
                    "impact": "Service unavailability"
                }
            ],
            "Elevation_of_Privilege": [
                {
                    "description": "{component} may allow privilege escalation",
                    "severity": "HIGH",
                    "impact": "Unauthorized administrative access"
                }
            ]
        }

    def _assess_likelihood(self, component: SystemComponent, threat: Dict) -> str:
        """Assess likelihood based on component exposure."""
        if component.internet_facing:
            return "HIGH"
        elif component.trust_boundary:
            return "MEDIUM"
        else:
            return "LOW"

    def _calculate_risk_score(self, threats: List[Threat]) -> int:
        """Calculate overall risk score (0-100)."""
        if not threats:
            return 0

        severity_map = {"CRITICAL": 25, "HIGH": 15, "MEDIUM": 10, "LOW": 5}
        total_risk = sum(severity_map.get(t.severity, 0) for t in threats)

        return min(total_risk, 100)
```

---

## 3. DevSecOps Pipeline Integration

**Concept**: Integrate security testing into CI/CD pipelines with automated SAST, DAST, and dependency scanning.

### 3.1 Automated Security Testing Pipeline

```python
class DevSecOpsPipeline:
    """Integrate security testing into CI/CD pipeline."""

    def __init__(self, config: PipelineConfig):
        """
        Initialize DevSecOps pipeline.

        Args:
            config: Pipeline configuration with tool settings
        """
        self.config = config
        self.sast_tools = self._initialize_sast_tools()
        self.dast_tools = self._initialize_dast_tools()
        self.sca_tools = self._initialize_sca_tools()

    def execute_security_pipeline(self, commit_sha: str) -> SecurityReport:
        """
        Execute complete security testing pipeline.

        Args:
            commit_sha: Git commit SHA to test

        Returns:
            Comprehensive security report

        Pipeline Stages:
            1. SAST (Static Application Security Testing)
            2. SCA (Software Composition Analysis)
            3. Secret Scanning
            4. DAST (Dynamic Application Security Testing)
            5. Container Security Scanning

        Example:
            >>> pipeline = DevSecOpsPipeline(config)
            >>> report = pipeline.execute_security_pipeline("abc123def456")
            >>> # Security Report:
            >>> # - SAST: 2 HIGH, 5 MEDIUM vulnerabilities
            >>> # - SCA: 1 CRITICAL dependency (Log4j 2.14.0)
            >>> # - Secrets: 0 exposed credentials
            >>> # - DAST: 3 MEDIUM vulnerabilities
            >>> # Overall: FAILED (blocking issues found)
        """
        results = {
            "commit_sha": commit_sha,
            "timestamp": datetime.now().isoformat(),
            "stages": {}
        }

        # Stage 1: SAST
        sast_result = self._run_sast(commit_sha)
        results["stages"]["sast"] = sast_result

        # Stage 2: SCA (Dependency Scanning)
        sca_result = self._run_sca(commit_sha)
        results["stages"]["sca"] = sca_result

        # Stage 3: Secret Scanning
        secret_result = self._run_secret_scan(commit_sha)
        results["stages"]["secret_scan"] = secret_result

        # Stage 4: DAST (requires deployed environment)
        if self.config.enable_dast:
            dast_result = self._run_dast(commit_sha)
            results["stages"]["dast"] = dast_result

        # Stage 5: Container Security
        if self.config.enable_container_scan:
            container_result = self._run_container_scan(commit_sha)
            results["stages"]["container_scan"] = container_result

        # Generate comprehensive report
        report = SecurityReport(
            commit_sha=commit_sha,
            results=results,
            overall_status=self._determine_pipeline_status(results),
            blocking_issues=self._extract_blocking_issues(results)
        )

        return report

    def _run_sast(self, commit_sha: str) -> Dict[str, Any]:
        """
        Run Static Application Security Testing.

        Tools: Bandit (Python), Semgrep, SonarQube
        """
        findings = []

        # Run Bandit for Python
        bandit_cmd = f"bandit -r src/ -f json -o /tmp/bandit-{commit_sha}.json"
        subprocess.run(bandit_cmd, shell=True, check=False)

        bandit_results = self._parse_bandit_results(f"/tmp/bandit-{commit_sha}.json")
        findings.extend(bandit_results)

        # Run Semgrep for multi-language analysis
        semgrep_cmd = f"semgrep --config auto --json -o /tmp/semgrep-{commit_sha}.json src/"
        subprocess.run(semgrep_cmd, shell=True, check=False)

        semgrep_results = self._parse_semgrep_results(f"/tmp/semgrep-{commit_sha}.json")
        findings.extend(semgrep_results)

        # Categorize findings by severity
        categorized = self._categorize_findings(findings)

        return {
            "tool": "SAST (Bandit + Semgrep)",
            "findings": categorized,
            "total_issues": len(findings),
            "status": "FAILED" if categorized["CRITICAL"] or categorized["HIGH"] else "PASSED"
        }

    def _run_sca(self, commit_sha: str) -> Dict[str, Any]:
        """
        Run Software Composition Analysis (dependency scanning).

        Tools: Safety (Python), Snyk, OWASP Dependency-Check
        """
        findings = []

        # Run Safety for Python dependencies
        safety_cmd = "safety check --json"
        result = subprocess.run(
            safety_cmd,
            shell=True,
            capture_output=True,
            text=True,
            check=False
        )

        if result.returncode != 0:
            safety_data = json.loads(result.stdout)
            for vuln in safety_data.get("vulnerabilities", []):
                findings.append({
                    "type": "VULNERABLE_DEPENDENCY",
                    "package": vuln["package"],
                    "version": vuln["installed_version"],
                    "vulnerability": vuln["vulnerability"],
                    "severity": vuln["severity"],
                    "cve": vuln.get("cve"),
                    "fix": vuln.get("fix_version")
                })

        categorized = self._categorize_findings(findings)

        return {
            "tool": "SCA (Safety)",
            "findings": categorized,
            "total_issues": len(findings),
            "status": "FAILED" if categorized["CRITICAL"] else "PASSED"
        }

    def _run_secret_scan(self, commit_sha: str) -> Dict[str, Any]:
        """
        Scan for exposed secrets and credentials.

        Tools: TruffleHog, git-secrets, detect-secrets
        """
        findings = []

        # Run TruffleHog
        trufflehog_cmd = f"trufflehog git file://. --since-commit {commit_sha} --json"
        result = subprocess.run(
            trufflehog_cmd,
            shell=True,
            capture_output=True,
            text=True,
            check=False
        )

        if result.stdout:
            for line in result.stdout.strip().split("\n"):
                if line:
                    secret = json.loads(line)
