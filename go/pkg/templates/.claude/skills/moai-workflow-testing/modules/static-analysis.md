# Static Analysis Tools Integration

> Module: pylint, flake8, bandit, mypy integration for comprehensive static analysis
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Intermediate
> Time: 15+ minutes
> Dependencies: Python 3.8+, pylint, flake8, bandit, mypy, subprocess, json

## Quick Reference

### Supported Tools

pylint: Code quality and style checking
- Comprehensive analysis of Python code
- Detects bugs, code smells, and style violations
- Provides code ratings and detailed reports
- Configurable with project-specific rules

flake8: Style guide enforcement
- Wrapper around pyflakes, pycodestyle, and McCabe
- Fast and lightweight style checking
- Enforces PEP 8 style guide
- Highly customizable with plugins

bandit: Security vulnerability scanning
- Finds common security issues in Python code
- Uses plugin-based architecture
- Configurable severity levels
- Integrates with security best practices

mypy: Type checking and validation
- Static type checker for Python
- Catches type errors before runtime
- Supports gradual typing
- Improves code reliability and documentation

### Core Implementation

```python
import subprocess
import json
from typing import Dict, List, Any

class StaticAnalysisTools:
    """Wrapper for various static analysis tools."""

    def __init__(self):
        self.tools = {
            'pylint': self._run_pylint,
            'flake8': self._run_flake8,
            'bandit': self._run_bandit,
            'mypy': self._run_mypy
        }

    async def run_all_analyses(self, file_path: str) -> Dict[str, Any]:
        """Run all available static analysis tools."""
        results = {}

        for tool_name, tool_func in self.tools.items():
            try:
                result = await tool_func(file_path)
                results[tool_name] = result
            except Exception as e:
                results[tool_name] = {'error': str(e)}

        return results
```

---

## Implementation Guide

### pylint Integration

```python
async def _run_pylint(self, file_path: str) -> Dict[str, Any]:
    """Run pylint analysis."""
    try:
        result = subprocess.run(
            ['pylint', file_path, '--output-format=json'],
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            return {'issues': []}

        try:
            issues = json.loads(result.stdout)
            return {
                'issues': issues,
                'summary': self._parse_pylint_summary(result.stderr)
            }
        except json.JSONDecodeError:
            return {
                'raw_output': result.stdout,
                'raw_errors': result.stderr
            }

    except FileNotFoundError:
        return {'error': 'pylint not installed'}
```

pylint Configuration:
Create .pylintrc file for project-specific rules:

```ini
[MASTER]
disable=C0111  # Disable missing-docstring warning if needed

[MESSAGES CONTROL]
disable=
    C0111,  # missing-docstring
    R0903,  # too-few-public-methods
    C0103,  # invalid-name

[DESIGN]
max-args=7
max-locals=15
max-returns=6
max-branches=12
max-statements=50
```

### flake8 Integration

```python
async def _run_flake8(self, file_path: str) -> Dict[str, Any]:
    """Run flake8 analysis."""
    try:
        result = subprocess.run(
            ['flake8', file_path, '--format=json'],
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            return {'issues': []}

        # Parse flake8 output
        issues = []
        for line in result.stdout.split('\n'):
            if line.strip():
                parts = line.split(':')
                if len(parts) >= 4:
                    issues.append({
                        'path': parts[0],
                        'line': int(parts[1]),
                        'column': int(parts[2]),
                        'code': parts[3].strip(),
                        'message': ':'.join(parts[4:]).strip()
                    })

        return {'issues': issues}

    except FileNotFoundError:
        return {'error': 'flake8 not installed'}
```

flake8 Configuration:
Create setup.cfg or .flake8 file:

```ini
[flake8]
max-line-length = 100
exclude =
    .git,
    __pycache__,
    venv,
    env
ignore =
    E203,  # whitespace before ':'
    W503,  # line break before binary operator
max-complexity = 10
```

### bandit Integration

```python
async def _run_bandit(self, file_path: str) -> Dict[str, Any]:
    """Run bandit security analysis."""
    try:
        result = subprocess.run(
            ['bandit', '-f', 'json', file_path],
            capture_output=True,
            text=True
        )

        try:
            bandit_results = json.loads(result.stdout)
            return bandit_results
        except json.JSONDecodeError:
            return {'raw_output': result.stdout}

    except FileNotFoundError:
        return {'error': 'bandit not installed'}
```

bandit Configuration:
Create .bandit file:

```yaml
exclude_dirs:
    - '/tests'
    - '/venv'
    - '/__pycache__'

tests:
  - B201  # flask_debug_true
  - B301  # pickle
  - B302  # marshal
  - B303  # md5
  - B304  # ciphers
  - B305  # cipher_modes
  - B306  # mktemp_q
  - B307  # eval
  - B308  # mark_safe
  - B309  # httpsconnection
  - B310  # urllib_urlopen
  - B311  # random
  - B312  # telnetlib
  - B313  # xml_bad_cElementTree
  - B314  # xml_bad_ElementTree
  - B315  # xml_bad_expatreader
  - B316  # xml_bad_expatbuilder
  - B317  # xml_bad_sax
  - B318  # xml_bad_minidom
  - B319  # xml_bad_pulldom
  - B320  # xml_bad_etree
  - B321  # ftplib
  - B322  # input
  - B323  # unverified_context
  - B324  # hashlib_new_insecure_functions
  - B325  # tempnam
  - B401  # import_telnetlib
  - B402  # import_ftplib
  - B403  # import_pickle
  - B404  # import_subprocess
  - B405  # import_xml_etree
  - B406  # import_xml_sax
  - B407  # import_xml_expat
  - B408  # import_xml_minidom
  - B409  # import_xml_pulldom
  - B410  # import_lxml
  - B411  # import_xmlrpclib
  - B412  # import_httpoxy
  - B413  # import_pycrypto
  - B501  # request_with_no_cert_validation
  - B502  # ssl_with_bad_version
  - B503  # ssl_with_bad_defaults
  - B504  # ssl_with_no_version
  - B505  # weak_cryptographic_key
  - B506  # yaml_load
  - B507  # ssh_no_host_key_verification
  - B601  # paramiko_calls
  - B602  # shell_injection_subprocess
  - B603  # subprocess_without_shell_equals_true
  - B604  # any_other_function_with_shell_equals_true
  - B605  # start_process_with_a_shell
  - B606  # start_process_with_no_shell
  - B607  # start_process_with_partial_path
  - B608  # hardcoded_sql_expressions
  - B609  # linux_commands_wildcard_injection
  - B610  # django_extra_used
  - B611  # django_rawsql_used
  - B701  # jinja2_autoescape_false
  - B702  # use_of_mako_templates
  - B703  # django_mark_safe
```

### mypy Integration

```python
async def _run_mypy(self, file_path: str) -> Dict[str, Any]:
    """Run mypy type analysis."""
    try:
        result = subprocess.run(
            ['mypy', file_path, '--show-error-codes'],
            capture_output=True,
            text=True
        )

        # Parse mypy output
        issues = []
        for line in result.stdout.split('\n'):
            if ':' in line and 'error:' in line:
                parts = line.split(':', 3)
                if len(parts) >= 4:
                    issues.append({
                        'path': parts[0],
                        'line': int(parts[1]),
                        'message': parts[3].strip()
                    })

        return {'issues': issues}

    except FileNotFoundError:
        return {'error': 'mypy not installed'}
```

mypy Configuration:
Create mypy.ini file:

```ini
[mypy]
python_version = 3.8
warn_return_any = True
warn_unused_configs = True
disallow_untyped_defs = False  # Set to True for strict typing
check_untyped_defs = True

[mypy-tests.*]
disallow_untyped_defs = False

[mypy-third_party_library.*]
ignore_missing_imports = True
```

---

## Tool-to-TRUST Category Mapping

```python
def _map_tool_to_trust_category(self, tool_name: str, issue_data: Dict) -> TrustCategory:
    """Map static analysis tool to TRUST category."""

    if tool_name == 'bandit':
        return TrustCategory.SAFETY
    elif tool_name == 'mypy':
        return TrustCategory.TRUTHFULNESS
    elif tool_name == 'pylint':
        message = issue_data.get('message', '').lower()
        if any(keyword in message for keyword in ['security', 'injection', 'unsafe']):
            return TrustCategory.SAFETY
        elif any(keyword in message for keyword in ['performance', 'inefficient']):
            return TrustCategory.TIMELINESS
        else:
            return TrustCategory.USABILITY
    else:
        return TrustCategory.USABILITY
```

---

## Result Normalization

```python
def _convert_static_issues(
    self, static_results: Dict[str, Any], file_path: str
) -> List[CodeIssue]:
    """Convert static analysis results to CodeIssue objects."""

    issues = []

    for tool_name, results in static_results.items():
        if 'error' in results:
            continue

        tool_issues = results.get('issues', [])
        for issue_data in tool_issues:
            category = self._map_tool_to_trust_category(tool_name, issue_data)

            issue = CodeIssue(
                id=f"{tool_name}_{len(issues)}",
                category=category,
                severity=self._map_severity(issue_data.get('severity', 'medium')),
                issue_type=self._map_issue_type(tool_name, issue_data),
                title=f"{tool_name.title()}: {issue_data.get('message', 'Unknown issue')}",
                description=issue_data.get('message', 'Static analysis issue'),
                file_path=file_path,
                line_number=issue_data.get('line', 0),
                column_number=issue_data.get('column', 0),
                code_snippet=issue_data.get('code_snippet', ''),
                suggested_fix=self._get_suggested_fix(tool_name, issue_data),
                confidence=0.8,
                rule_violated=issue_data.get('code', ''),
                external_reference=f"{tool_name} documentation"
            )
            issues.append(issue)

    return issues
```

---

## Error Handling

```python
class StaticAnalysisTools:
    """Enhanced static analysis tools with error handling."""

    def __init__(self, fallback_to_defaults: bool = True):
        self.tools = {
            'pylint': self._run_pylint_safe,
            'flake8': self._run_flake8_safe,
            'bandit': self._run_bandit_safe,
            'mypy': self._run_mypy_safe
        }
        self.fallback_to_defaults = fallback_to_defaults

    async def _run_pylint_safe(self, file_path: str) -> Dict[str, Any]:
        """Run pylint with fallback to basic checks."""
        try:
            return await self._run_pylint(file_path)
        except FileNotFoundError:
            if self.fallback_to_defaults:
                return self._get_basic_style_checks(file_path)
            return {'error': 'pylint not installed and no fallback available'}
        except Exception as e:
            return {'error': f'pylint execution failed: {str(e)}'}
```

---

## Best Practices

1. Tool Availability: Ensure all tools are installed in development environment
2. Configuration: Use configuration files for project-specific rules
3. CI Integration: Run static analysis in CI/CD pipeline
4. Incremental Adoption: Start with subset of tools, gradually add more
5. Error Handling: Implement fallback mechanisms for missing tools
6. Performance: Cache results to avoid redundant analysis
7. Team Consistency: Use same tool versions across team
8. Regular Updates: Keep tools updated for latest checks and fixes

---

## Related Modules

- [TRUST 5 Validation](./trust5-validation.md): Category mapping and scoring
- [Security Analysis](./security-analysis.md): Enhanced security detection
- [Quality Metrics](./quality-metrics.md): Code quality analysis

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/static-analysis.md`
