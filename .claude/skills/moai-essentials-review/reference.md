# 코드 리뷰 빠른 참고

## 리뷰 도구

```bash
# Pylint (Python 품질)
pylint src/ --fail-under=8.0

# ESLint (JavaScript 스타일)
eslint . --max-warnings 0

# Black (Python 포맷)
black --check src/

# mypy (Python 타입 검사)
mypy src/ --strict

# Bandit (Python 보안)
bandit -r src/ -ll
```

## TRUST 5 검증 체크리스트

| 원칙 | 메트릭 | 목표 | 도구 |
|------|--------|------|------|
| **T**est | Coverage | ≥85% | pytest, coverage.py |
| **R**eadable | Complexity | <10 | pylint, lizard |
| **U**nified | Consistency | 0 violations | black, eslint |
| **S**ecured | Vulnerabilities | 0 critical | bandit, snyk |
| **T**rackable | Documentation | 100% | docstring, jsdoc |

## Context7 링크

- 보안 패턴: https://owasp.org/Top10
- 코드 품질: https://github.com/PyCQA
- 성능: https://refactoring.guru/

---

**Last Updated**: 2025-11-22
