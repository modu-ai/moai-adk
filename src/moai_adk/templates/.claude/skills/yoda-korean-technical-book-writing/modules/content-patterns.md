# Content Patterns Module

## 개념 설명 패턴

**Analogy-First 접근법**:
```markdown
## [개념명]

### 일상 비유로 이해하기
[독자가 익숙한 상황과 연결]
예: "변수는 이름표가 붙은 상자와 같습니다..."

### 기술적 정의
[정확한 용어로 재정의]

### 실제 사용 사례
[현업 예시]

### 주의사항
[흔한 오해, 안티패턴]
```

## 코드 예제 패턴

**Progressive Disclosure (점진적 공개)**:

**Step 1: 최소 실행 가능 예제**
```python
# 가장 단순한 형태
def greet(name):
    return f"안녕하세요, {name}님!"

print(greet("철수"))  # 출력: 안녕하세요, 철수님!
```

**Step 2: 실용적 확장**
```python
# 에러 처리 추가
def greet(name):
    if not name or not isinstance(name, str):
        raise ValueError("이름은 비어있지 않은 문자열이어야 합니다")
    return f"안녕하세요, {name}님!"

try:
    print(greet(""))
except ValueError as e:
    print(f"오류: {e}")
```

**Step 3: 프로덕션 수준**
```python
# 타입 힌트, 독스트링, 로깅 추가
import logging
from typing import Optional

def greet(name: str, title: Optional[str] = None) -> str:
    """
    사용자에게 인사 메시지를 생성합니다.
    
    Args:
        name: 사용자 이름 (필수)
        title: 호칭 (선택, 예: "개발자", "학생")
    
    Returns:
        인사 메시지 문자열
    
    Raises:
        ValueError: name이 빈 문자열인 경우
    """
    if not name:
        logging.error("빈 이름으로 greet 호출됨")
        raise ValueError("이름은 비어있을 수 없습니다")
    
    greeting = f"안녕하세요, {name}님!"
    if title:
        greeting += f" ({title})"
    
    return greeting
```

## 연습 문제 설계

**난이도 계층 구조**:
```
Level 1: 기억 (Recall)
"다음 중 함수의 정의로 옳은 것은?"

Level 2: 이해 (Comprehension)
"다음 코드의 출력 결과를 예측하세요"

Level 3: 적용 (Application)
"리스트에서 짝수만 필터링하는 함수를 작성하세요"

Level 4: 분석 (Analysis)
"다음 두 코드의 시간 복잡도를 비교하고 설명하세요"

Level 5: 창조 (Creation)
"간단한 할 일 관리 CLI 앱을 설계하고 구현하세요"
```

**효과적인 연습 문제 구성** (장당):
- Level 1-2: 3문제 (기초 확인)
- Level 3: 2문제 (실습)
- Level 4-5: 1-2문제 (심화)

## 다이어그램 통합 전략

**다이어그램 타입별 용도**:

| 타입 | 사용 시점 | 도구 예시 |
|------|----------|----------|
| 플로우차트 | 알고리즘 흐름 설명 | Mermaid, draw.io |
| 시퀀스 다이어그램 | 시스템 간 상호작용 | PlantUML, Mermaid |
| 클래스 다이어그램 | 객체 관계 모델링 | UML 도구 |
| 아키텍처 다이어그램 | 시스템 구조 | Lucidchart, Excalidraw |
| 마인드맵 | 개념 관계 시각화 | XMind, Coggle |
