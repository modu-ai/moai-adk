# 코드 리뷰 실전 예제

## Example 1: PR 자동 리뷰

**리뷰 체크리스트**:
```python
async def review_pull_request(pr_data):
    """PR 자동 리뷰."""

    # 1. 품질 검사
    quality = await analyze_code_quality(pr_data['files'])

    # 2. 보안 검사
    security = await check_security_issues(pr_data['files'])

    # 3. 테스트 커버리지
    coverage = await verify_test_coverage(pr_data['files'])

    # 4. 성능 분석
    performance = await analyze_performance(pr_data['files'])

    # 5. 리뷰 결과
    return {
        'approval': quality.score >= 80 and security.issues == 0,
        'issues': security.issues + performance.bottlenecks,
        'suggestions': generate_improvements(quality, performance)
    }
```

## Example 2: TRUST 5 검증

**TRUST 5 자동 검증**:
```python
async def validate_trust_5(code):
    """TRUST 5 원칙 검증."""

    trust_checks = {
        'T (Test)': {
            'coverage': await measure_coverage(code),
            'min_target': 0.85
        },
        'R (Readable)': {
            'complexity': calculate_cyclomatic_complexity(code),
            'max_target': 10
        },
        'U (Unified)': {
            'consistency': check_style_consistency(code),
            'violation_count': count_violations(code)
        },
        'S (Secured)': {
            'vulnerabilities': await scan_security(code),
            'max_target': 0
        },
        'T (Trackable)': {
            'documentation': check_documentation(code),
            'test_links': verify_test_links(code)
        }
    }

    return verify_all_checks(trust_checks)
```

## Example 3: 보안 취약점 검사

**OWASP 규칙 기반 검사**:
```python
async def check_security_vulnerabilities(code, language):
    """보안 취약점 검사."""

    vulnerabilities = []

    # SQL Injection 검사
    if 'SQL' in language:
        sql_issues = detect_sql_injection(code)
        vulnerabilities.extend(sql_issues)

    # XSS 검사
    if 'JavaScript' in language:
        xss_issues = detect_xss_vulnerability(code)
        vulnerabilities.extend(xss_issues)

    # 인증/인가 검사
    auth_issues = check_authentication(code)
    vulnerabilities.extend(auth_issues)

    return {
        'total_vulnerabilities': len(vulnerabilities),
        'critical': count_by_severity(vulnerabilities, 'critical'),
        'issues': vulnerabilities,
        'remediation': generate_fixes(vulnerabilities)
    }
```

## Example 4: 성능 분석

**성능 병목 식별**:
```python
async def analyze_performance(code):
    """성능 병목 분석."""

    bottlenecks = []

    # N+1 쿼리 검사
    n_plus_one = detect_n_plus_one_queries(code)
    if n_plus_one:
        bottlenecks.append({
            'type': 'N+1 Query',
            'severity': 'high',
            'fix': 'Use batch queries or eager loading'
        })

    # 무한 루프 검사
    infinite_loops = detect_infinite_loops(code)
    bottlenecks.extend(infinite_loops)

    # 메모리 누수 검사
    leaks = detect_memory_leaks(code)
    bottlenecks.extend(leaks)

    return {
        'bottlenecks': bottlenecks,
        'severity': max_severity(bottlenecks),
        'recommendations': generate_optimizations(bottlenecks)
    }
```

## Example 5: 자동 리뷰 코멘트

**GitHub PR 자동 코멘트**:
```python
async def post_review_comments(pr_number, review_results):
    """리뷰 결과를 PR에 포스팅."""

    # 주요 이슈 코멘트
    for issue in review_results['critical_issues']:
        comment = f"""
        🔴 Critical Issue: {issue['type']}

        **Location**: {issue['file']}:{issue['line']}
        **Severity**: {issue['severity']}

        **Problem**: {issue['description']}

        **Fix**: {issue['suggested_fix']}
        """
        post_comment(pr_number, comment)

    # 개선 제안 코멘트
    for suggestion in review_results['suggestions']:
        comment = f"""
        💡 Suggestion: {suggestion['category']}

        **Current**: {suggestion['current']}
        **Recommended**: {suggestion['recommended']}
        **Reason**: {suggestion['reason']}
        """
        post_comment(pr_number, comment)

    # 최종 승인 코멘트
    if review_results['approval']:
        post_comment(pr_number, "✅ Approved - Ready to merge")
    else:
        post_comment(pr_number, "⏸ Changes requested")
```

---

**Last Updated**: 2025-11-22
**Total Examples**: 5 practical code review scenarios
