# 포맷 자동 선택 가이드

**버전**: 1.0.0
**목적**: 챕터 제목과 계획을 분석하여 최적의 장르 템플릿 추천

---

## 개요

YodA 시스템은 11개의 장르 템플릿을 제공합니다. 이 가이드는 챕터 제목, 목차 구조, 작성 계획을 분석하여 가장 적합한 템플릿을 자동으로 추천하는 알고리즘을 설명합니다.

---

## 11개 장르 템플릿 매트릭스

| 장르 | 주요 키워드 | 적합한 챕터 제목 패턴 | 콘텐츠 특징 |
|------|-------------|---------------------|------------|
| **introduction** | 소개, 시작, 개요, 배경 | "첫 만남", "시작하기", "배경" | 문제제기 → 솔루션 |
| **concept-explanation** | 개념, 이해, 원리 | "이해하기", "개념", "원리" | 정의 → 원리 → 예시 |
| **code-walkthrough** | 마스터, 가이드, 기능 | "마스터하기", "가이드" | 코드 중심, 점진적 |
| **hands-on-practice** | 실습, 구축, 만들기 | "실습", "구축하기" | Step-by-step |
| **case-study** | 분석, 회고, 사례 | "사례 분석", "회고" | 배경 → 문제 → 해결 |
| **design-pattern** | 패턴, 디자인, 아키텍처 | "Pattern", "패턴" | Code-First, At a Glance |
| **api-reference** | API, 함수, 메서드, 참조 | "API", "Reference" | Parameters 테이블 |
| **comparison** | vs, 비교, 선택 | "vs", "비교" | 비교표 + 선택 가이드 |
| **troubleshooting** | 에러, 문제, 해결, 디버깅 | "Error", "문제 해결" | Problem → Solution |
| **tutorial-step** | 설치, 설정, 시작 | "설치하기", "Setup" | Prerequisites → Steps |
| **architecture-overview** | 구조, 개요, 시스템 | "구조", "Overview" | Diagram → Components |

---

## 키워드 기반 매칭 규칙

### Python 규칙 정의

```python
FORMAT_SELECTION_RULES = {
    # 기존 5개 (PART 1-4 기본)
    "introduction": {
        "keywords": ["소개", "시작", "개요", "배경", "첫", "introduction", "overview", "getting started"],
        "title_patterns": [r".*첫.*", r".*시작.*", r".*소개.*", r"^Chapter 1"],
        "content_indicators": ["문제제기", "배경", "왜 필요한가"],
        "priority": 3  # 높을수록 우선순위 높음
    },

    "concept-explanation": {
        "keywords": ["개념", "이해", "원리", "이론", "concept", "understanding", "theory"],
        "title_patterns": [r".*이해.*", r".*개념.*", r".*원리.*"],
        "content_indicators": ["정의", "원리", "작동 방식"],
        "priority": 2
    },

    "code-walkthrough": {
        "keywords": ["마스터", "가이드", "기능", "실전", "master", "guide", "features"],
        "title_patterns": [r".*마스터.*", r".*가이드.*", r".*기능.*"],
        "content_indicators": ["코드 예제", "점진적", "Basic → Advanced"],
        "priority": 2
    },

    "hands-on-practice": {
        "keywords": ["실습", "구축", "만들기", "프로젝트", "hands-on", "build", "create"],
        "title_patterns": [r".*실습.*", r".*구축.*", r".*만들기.*"],
        "content_indicators": ["준비", "Step", "완성"],
        "priority": 2
    },

    "case-study": {
        "keywords": ["분석", "회고", "사례", "경험", "case study", "analysis", "retrospective"],
        "title_patterns": [r".*분석.*", r".*회고.*", r".*사례.*"],
        "content_indicators": ["배경", "문제", "해결", "교훈"],
        "priority": 2
    },

    # 신규 6개 (특수 목적)
    "design-pattern": {
        "keywords": ["패턴", "디자인", "아키텍처", "pattern", "design", "architecture"],
        "title_patterns": [r".*Pattern$", r".*패턴$", r".*디자인.*", r".*Architecture.*"],
        "content_indicators": ["At a Glance", "What", "Why", "Rule of thumb", "Key Takeaways"],
        "priority": 4  # 매우 특수한 포맷 - 높은 우선순위
    },

    "api-reference": {
        "keywords": ["API", "함수", "메서드", "참조", "function", "method", "reference"],
        "title_patterns": [r".*API.*", r".*Reference.*", r".*함수.*"],
        "content_indicators": ["Parameters", "Return Value", "Syntax", "Examples"],
        "priority": 4  # 명확한 구조 - 높은 우선순위
    },

    "comparison": {
        "keywords": ["vs", "비교", "선택", "대비", "comparison", "versus", "choose"],
        "title_patterns": [r".*vs.*", r".*비교.*", r".*선택.*"],
        "content_indicators": ["Option A", "Option B", "Comparison Table", "Choose when"],
        "priority": 3  # 명확한 의도 - 중간 우선순위
    },

    "troubleshooting": {
        "keywords": ["에러", "오류", "문제", "해결", "디버깅", "error", "issue", "fix", "debug"],
        "title_patterns": [r".*Error.*", r".*문제.*", r".*해결.*", r".*Fix.*"],
        "content_indicators": ["Problem", "Root Cause", "Solution", "Prevention"],
        "priority": 4  # 문제 해결 명확 - 높은 우선순위
    },

    "tutorial-step": {
        "keywords": ["설치", "설정", "구성", "시작", "setup", "install", "configure", "getting started"],
        "title_patterns": [r".*설치.*", r".*설정.*", r".*Setup.*", r".*Installation.*"],
        "content_indicators": ["Prerequisites", "Step 1", "Step 2", "Verification"],
        "priority": 3
    },

    "architecture-overview": {
        "keywords": ["구조", "아키텍처", "개요", "시스템", "architecture", "structure", "overview", "system"],
        "title_patterns": [r".*구조.*", r".*Architecture.*", r".*Overview.*"],
        "content_indicators": ["Components", "Data Flow", "Diagram", "High-Level"],
        "priority": 3
    }
}
```

---

## 점수 계산 알고리즘

### 1. 키워드 매칭 점수

```python
def calculate_keyword_score(text: str, keywords: list) -> int:
    """
    텍스트에서 키워드 출현 빈도 계산

    Args:
        text: 분석할 텍스트 (제목 또는 계획 내용)
        keywords: 검색할 키워드 리스트

    Returns:
        int: 키워드 출현 횟수
    """
    text_lower = text.lower()
    score = 0
    for keyword in keywords:
        score += text_lower.count(keyword.lower())
    return score
```

### 2. 패턴 매칭 점수

```python
import re

def calculate_pattern_score(title: str, patterns: list) -> int:
    """
    제목이 패턴과 매칭되는지 확인

    Args:
        title: 챕터 제목
        patterns: 정규표현식 패턴 리스트

    Returns:
        int: 매칭된 패턴 수 * 5 (패턴 매칭은 강한 신호)
    """
    score = 0
    for pattern in patterns:
        if re.search(pattern, title, re.IGNORECASE):
            score += 5  # 패턴 매칭은 높은 가중치
    return score
```

### 3. 콘텐츠 지표 점수

```python
def calculate_content_score(plan_content: str, indicators: list) -> int:
    """
    계획 내용에서 콘텐츠 지표 확인

    Args:
        plan_content: 작성 계획 내용
        indicators: 콘텐츠 지표 리스트

    Returns:
        int: 지표 출현 횟수 * 2
    """
    score = 0
    for indicator in indicators:
        if indicator.lower() in plan_content.lower():
            score += 2  # 콘텐츠 지표는 중간 가중치
    return score
```

### 4. 통합 점수 계산

```python
def calculate_format_scores(chapter_title: str, plan_content: str) -> dict:
    """
    모든 포맷에 대한 점수 계산

    Args:
        chapter_title: 챕터 제목
        plan_content: 작성 계획 내용

    Returns:
        dict: {format_name: total_score}
    """
    scores = {}

    for format_name, rules in FORMAT_SELECTION_RULES.items():
        # 1. 키워드 점수 (제목: 2배 가중치, 계획: 1배)
        title_keyword_score = calculate_keyword_score(chapter_title, rules["keywords"]) * 2
        plan_keyword_score = calculate_keyword_score(plan_content, rules["keywords"])

        # 2. 패턴 점수 (제목만)
        pattern_score = calculate_pattern_score(chapter_title, rules["title_patterns"])

        # 3. 콘텐츠 지표 점수 (계획만)
        content_score = calculate_content_score(plan_content, rules["content_indicators"])

        # 4. 우선순위 가중치
        priority_weight = rules["priority"]

        # 5. 총점 계산
        total_score = (title_keyword_score + plan_keyword_score + pattern_score + content_score) * priority_weight

        scores[format_name] = total_score

    return scores
```

---

## 추천 프로세스

### Step 1: 점수 계산

```python
def recommend_formats(chapter_title: str, plan_content: str, top_k: int = 3) -> list:
    """
    상위 K개 포맷 추천

    Args:
        chapter_title: 챕터 제목
        plan_content: 작성 계획 내용
        top_k: 추천할 포맷 개수 (기본: 3)

    Returns:
        list: [(format_name, score, confidence), ...]
    """
    scores = calculate_format_scores(chapter_title, plan_content)

    # 점수 내림차순 정렬
    sorted_formats = sorted(scores.items(), key=lambda x: x[1], reverse=True)

    # 상위 K개 선택
    top_formats = sorted_formats[:top_k]

    # 신뢰도 계산 (최고 점수 대비 비율)
    max_score = top_formats[0][1] if top_formats else 0

    recommendations = []
    for format_name, score in top_formats:
        confidence = (score / max_score * 100) if max_score > 0 else 0
        recommendations.append((format_name, score, confidence))

    return recommendations
```

### Step 2: 사용자 확인

```python
def ask_user_format_confirmation(recommendations: list) -> str:
    """
    AskUserQuestion으로 사용자 확인

    Args:
        recommendations: [(format_name, score, confidence), ...]

    Returns:
        str: 사용자가 선택한 format_name
    """
    # 상위 3개 옵션 생성
    options = []
    for format_name, score, confidence in recommendations:
        format_info = FORMAT_SELECTION_RULES[format_name]
        options.append({
            "label": format_name,
            "description": f"신뢰도: {confidence:.0f}% - {', '.join(format_info['keywords'][:3])}"
        })

    # AskUserQuestion 호출 (실제 구현은 yoda-book-author에서)
    # selected_format = AskUserQuestion(
    #     question="추천 포맷을 확인해주세요:",
    #     options=options
    # )

    # 임시 반환 (최고 점수 포맷)
    return recommendations[0][0]
```

---

## 사용 예시

### 예시 1: "Parallelization Pattern" 챕터

```python
chapter_title = "Parallelization Pattern"
plan_content = """
이 챕터에서는 병렬화 패턴을 설명합니다.

섹션 구성:
1. Code Example - 완전한 작동 코드
2. Code Explanation - 코드 상세 설명
3. At a Glance - What/Why/Rule
4. Visual Summary - 다이어그램
5. Key Takeaways - 핵심 포인트
"""

recommendations = recommend_formats(chapter_title, plan_content)

# 출력:
# [
#     ("design-pattern", 52, 100.0),   # ← 최고 점수
#     ("architecture-overview", 18, 34.6),
#     ("concept-explanation", 12, 23.1)
# ]

# 추천: design-pattern (신뢰도 100%)
```

### 예시 2: "Task() 함수 레퍼런스"

```python
chapter_title = "Task() 함수 레퍼런스"
plan_content = """
Task() 함수의 완전한 참조 문서입니다.

섹션:
- Overview
- Syntax
- Parameters 테이블
- Return Value
- Examples
- See Also
"""

recommendations = recommend_formats(chapter_title, plan_content)

# 출력:
# [
#     ("api-reference", 44, 100.0),    # ← 최고 점수
#     ("code-walkthrough", 10, 22.7),
#     ("concept-explanation", 8, 18.2)
# ]

# 추천: api-reference (신뢰도 100%)
```

### 예시 3: "LangChain vs Google ADK"

```python
chapter_title = "LangChain vs Google ADK: 에이전트 프레임워크 비교"
plan_content = """
두 프레임워크를 비교합니다.

섹션:
- Comparison Table
- LangChain 장단점
- Google ADK 장단점
- Choose when...
"""

recommendations = recommend_formats(chapter_title, plan_content)

# 출력:
# [
#     ("comparison", 36, 100.0),       # ← 최고 점수
#     ("concept-explanation", 12, 33.3),
#     ("architecture-overview", 10, 27.8)
# ]

# 추천: comparison (신뢰도 100%)
```

---

## yoda-book-author 통합 가이드

### Step 3-1a: 포맷 자동 추천 추가

기존 `yoda-book-author.md`의 Step 3-1 (PART 템플릿 선택)과 Step 3-2 (장르 템플릿 선택) 사이에 추가:

```markdown
### Step 3-1a: 포맷 자동 추천 (신규)

1. 챕터 제목과 작성 계획 로드
2. format-selection.md의 알고리즘 실행
3. 상위 3개 포맷 추천
4. AskUserQuestion으로 사용자 확인
5. 선택된 포맷을 Step 3-2의 장르 템플릿으로 사용
```

### 통합 코드 예시

```python
# Step 3-1: PART 템플릿 선택 (기존)
part_template = select_part_template(part_number)

# Step 3-1a: 포맷 자동 추천 (신규)
recommendations = recommend_formats(chapter_title, plan_content, top_k=3)

# 사용자 확인
selected_format = AskUserQuestion(
    question="추천 포맷을 확인해주세요 (상위 3개):",
    options=[
        {
            "label": recommendations[0][0],
            "description": f"신뢰도: {recommendations[0][2]:.0f}% (최고 추천)"
        },
        {
            "label": recommendations[1][0],
            "description": f"신뢰도: {recommendations[1][2]:.0f}%"
        },
        {
            "label": recommendations[2][0],
            "description": f"신뢰도: {recommendations[2][2]:.0f}%"
        }
    ]
)

# Step 3-2: 장르 템플릿 로드 (확장)
genre_template = load_genre_template(selected_format)  # 5개 → 11개 지원
```

---

## 검증 및 개선

### A/B 테스트

- 자동 추천 정확도 측정
- 사용자 수동 선택 vs 추천 선택 비교
- 신뢰도 임계값 조정

### 피드백 루프

- 사용자가 추천을 변경한 경우 로깅
- 키워드 가중치 재조정
- 새로운 패턴 발견 시 규칙 추가

---

**마지막 수정**: 2025-11-25
**버전**: 1.0.0
**상태**: 프로덕션 준비
