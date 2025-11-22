# 리팩토링 빠른 참고

## Rope 명령어 (Python)

```python
# 메서드 추출
Extract(project, resource, start, end).get_changes('method_name')

# 이름 변경
Rename(project, resource, offset).get_changes('new_name')

# 메서드 인라인
Inline(project, resource, offset).get_changes()

# 변수 추출
ExtractVariable(project, resource, start, end).get_changes('var_name')

# 클래스 추출
ExtractMethod(project, resource, start, end).get_changes('method_name')
```

## SOLID 원칙 체크리스트

- **S**ingle Responsibility: 한 클래스 = 한 이유로만 변경
- **O**pen/Closed: 확장에 열림, 수정에 닫힘
- **L**iskov Substitution: 기반 클래스 대체 가능
- **I**nterface Segregation: 인터페이스 분리
- **D**ependency Inversion: 추상화에 의존

## 기술 부채 점수

| 메트릭 | Good | Warning | Critical |
|------|------|---------|----------|
| 순환 복잡도 | <10 | 10-20 | >20 |
| 함수 길이 | <50줄 | 50-100줄 | >100줄 |
| 중복률 | <3% | 3-10% | >10% |
| 테스트 커버리지 | ≥85% | 70-85% | <70% |

## Context7 링크

- [Refactoring.Guru](https://refactoring.guru/)
- [Python Rope](https://github.com/python-rope/rope)
- [Black Formatter](https://github.com/psf/black)

---

**Last Updated**: 2025-11-22
