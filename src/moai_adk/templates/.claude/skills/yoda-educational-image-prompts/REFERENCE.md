# API 레퍼런스 & 고급 기능

**yoda-educational-image-prompts Skill - API 참조 문서**

이 문서는 yoda-educational-image-prompts Skill의 API 레퍼런스와 고급 기능을 제공합니다.

---

## 목차

1. [API 레퍼런스](#api-reference)
2. [고급 기능](#advanced-features)
3. [자동 스타일 추천 알고리즘](#auto-style-recommendation)
4. [배치 생성](#batch-generation)
5. [커스터마이징](#customization)

---

<a name="api-reference"></a>
## 📚 API 레퍼런스

### generate()

단일 프롬프트 생성

#### 시그니처

```python
def generate(
    content: str,
    style: str = "auto",
    language: str = "ko",
    context: Optional[str] = None,
    detail_level: str = "standard"
) -> str
```

#### 파라미터

| 파라미터 | 타입 | 필수 | 기본값 | 설명 |
|---------|------|------|--------|------|
| `content` | str | ✅ | - | 이미지로 표현할 내용 |
| `style` | str | ❌ | "auto" | 비주얼 스타일 선택 |
| `language` | str | ❌ | "ko" | 프롬프트 언어 |
| `context` | str | ❌ | None | 사용 맥락 (강의/책/슬라이드) |
| `detail_level` | str | ❌ | "standard" | 디테일 수준 |

#### 스타일 옵션

| 값 | 설명 |
|----|------|
| `"auto"` | 자동 선택 (콘텐츠 분석) |
| `"hand-drawn-sketch"` | 핸드드로잉 스케치 |
| `"isometric-3d"` | 아이소메트릭 3D 기술도 |
| `"minimalist-flat"` | 미니멀 플랫 디자인 |
| `"technical-blueprint"` | 기술 도면 블루프린트 |
| `"infographic-vector"` | 인포그래픽 벡터 |
| `"photorealistic-3d"` | 포토리얼리스틱 3D 렌더 |
| `"comic-educational"` | 만화/교육용 코믹 |
| `"gradient-modern"` | 그라디언트 현대 기술 |

#### 디테일 레벨 옵션

| 값 | 설명 | 프롬프트 길이 |
|----|------|--------------|
| `"minimal"` | 기본만 | ~200자 |
| `"standard"` | 표준 (권장) | ~400자 |
| `"detailed"` | 상세 | ~600자 |
| `"comprehensive"` | 완전 | ~800자 |

#### 반환값

생성된 한국어 프롬프트 (str)

#### 예제

```python
# 기본 사용
prompt = Skill("yoda-educational-image-prompts").generate(
    content="React 컴포넌트 생명주기와 Hook의 관계"
)

# 스타일 지정
prompt = Skill("yoda-educational-image-prompts").generate(
    content="마이크로서비스 아키텍처 다이어그램",
    style="isometric-3d"
)

# 상세한 프롬프트 생성
prompt = Skill("yoda-educational-image-prompts").generate(
    content="Git 브랜칭 전략 비교",
    style="infographic-vector",
    detail_level="comprehensive"
)

# 맥락 정보 포함
prompt = Skill("yoda-educational-image-prompts").generate(
    content="파이썬 리스트 기초",
    style="hand-drawn-sketch",
    context="chapter 3 hero image for beginners"
)
```

---

### generate_batch()

여러 섹션의 프롬프트를 한 번에 생성

#### 시그니처

```python
def generate_batch(
    sections: List[Dict[str, str]],
    language: str = "ko",
    style_strategy: str = "consistent"
) -> Dict[str, str]
```

#### 파라미터

| 파라미터 | 타입 | 필수 | 기본값 | 설명 |
|---------|------|------|--------|------|
| `sections` | List[Dict] | ✅ | - | 섹션 리스트 |
| `language` | str | ❌ | "ko" | 프롬프트 언어 |
| `style_strategy` | str | ❌ | "consistent" | 스타일 선택 전략 |

#### 섹션 딕셔너리 구조

```python
{
    "title": str,      # 섹션 제목
    "content": str,    # 섹션 내용
    "style": str       # (Optional) 특정 스타일 지정
}
```

#### 스타일 전략 옵션

| 값 | 설명 |
|----|------|
| `"consistent"` | 모든 섹션에 동일한 스타일 사용 |
| `"varied"` | 섹션마다 자동으로 다른 스타일 선택 |
| `"progressive"` | 섹션 순서에 따라 점진적으로 변화 |

#### 반환값

섹션 제목을 키로, 프롬프트를 값으로 하는 딕셔너리 (Dict[str, str])

#### 예제

```python
# 여러 섹션 배치 생성
sections = [
    {"title": "소개", "content": "React Hooks란?"},
    {"title": "기초", "content": "useState와 useEffect 사용법"},
    {"title": "고급", "content": "Custom Hooks 만들기"}
]

prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=sections,
    style_strategy="consistent"
)

# 결과
# {
#     "소개": "React Hooks를 표현한 핸드드로잉 삽화...",
#     "기초": "useState와 useEffect를 표현한 핸드드로잉 삽화...",
#     "고급": "Custom Hooks를 표현한 핸드드로잉 삽화..."
# }
```

---

<a name="advanced-features"></a>
## 🔧 고급 기능

### 자동 스타일 추천

콘텐츠를 분석해서 최적의 스타일을 자동으로 선택합니다.

#### 작동 원리

1. **키워드 추출**: 콘텐츠에서 핵심 키워드 추출
2. **카테고리 분류**: 키워드를 기술 카테고리로 분류
3. **스타일 매칭**: 카테고리에 맞는 최적 스타일 선택

#### 카테고리별 스타일 매칭

| 카테고리 | 자동 선택 스타일 | 이유 |
|---------|----------------|------|
| **아키텍처/시스템** | isometric-3d | 입체감과 구조 표현 |
| **알고리즘/자료구조** | technical-blueprint | 정밀한 동작 원리 |
| **웹 개발/UI** | minimalist-flat | 명확한 흐름 표현 |
| **데이터/비교** | infographic-vector | 시각적 비교 효과 |
| **AI/블록체인** | gradient-modern | 미래지향적 느낌 |
| **하드웨어/IoT** | photorealistic-3d | 사실감 있는 표현 |
| **튜토리얼/단계** | comic-educational | 스토리를 통한 학습 |
| **기초 개념** | hand-drawn-sketch | 친근하고 접근 쉬움 |

#### 예제

```python
# 자동 스타일 선택
prompt1 = Skill("yoda-educational-image-prompts").generate(
    content="마이크로서비스 아키텍처와 그 장점들",
    style="auto"  # 자동으로 "isometric-3d" 선택
)

prompt2 = Skill("yoda-educational-image-prompts").generate(
    content="퀵소트 알고리즘의 단계별 실행",
    style="auto"  # 자동으로 "technical-blueprint" 선택
)

prompt3 = Skill("yoda-educational-image-prompts").generate(
    content="React useState 기초 사용법",
    style="auto"  # 자동으로 "hand-drawn-sketch" 선택
)
```

---

<a name="auto-style-recommendation"></a>
## 🤖 자동 스타일 추천 알고리즘

### 알고리즘 상세

#### 1단계: 키워드 추출

```python
def extract_keywords(content: str) -> List[str]:
    """
    콘텐츠에서 기술 키워드 추출
    """
    tech_keywords = {
        "architecture": ["마이크로서비스", "아키텍처", "시스템", "구조"],
        "algorithm": ["알고리즘", "자료구조", "정렬", "탐색", "시간복잡도"],
        "web": ["웹", "프론트엔드", "백엔드", "API", "REST"],
        "database": ["데이터베이스", "SQL", "NoSQL", "쿼리"],
        "ai": ["AI", "ML", "머신러닝", "신경망", "딥러닝"],
        "blockchain": ["블록체인", "암호화폐", "스마트계약"],
        "hardware": ["하드웨어", "IoT", "센서", "라즈베리파이"],
        # ... more categories
    }
    
    extracted = []
    for category, keywords in tech_keywords.items():
        for keyword in keywords:
            if keyword in content:
                extracted.append((category, keyword))
    
    return extracted
```

#### 2단계: 카테고리 분류

```python
def classify_category(keywords: List[Tuple[str, str]]) -> str:
    """
    추출된 키워드를 기반으로 카테고리 분류
    """
    category_scores = {}
    
    for category, keyword in keywords:
        category_scores[category] = category_scores.get(category, 0) + 1
    
    # 가장 많이 등장한 카테고리 선택
    if category_scores:
        return max(category_scores, key=category_scores.get)
    else:
        return "basic"  # 기본 카테고리
```

#### 3단계: 스타일 매칭

```python
def match_style(category: str, context: Optional[str] = None) -> str:
    """
    카테고리에 맞는 최적 스타일 선택
    """
    style_mapping = {
        "architecture": "isometric-3d",
        "algorithm": "technical-blueprint",
        "web": "minimalist-flat",
        "database": "technical-blueprint",
        "ai": "gradient-modern",
        "blockchain": "gradient-modern",
        "hardware": "photorealistic-3d",
        "basic": "hand-drawn-sketch"
    }
    
    # 맥락 정보가 있으면 추가 조정
    if context and "hero" in context.lower():
        return "hand-drawn-sketch"  # Hero 이미지는 친근한 스타일
    
    return style_mapping.get(category, "hand-drawn-sketch")
```

---

<a name="batch-generation"></a>
## 📦 배치 생성

### 사용 사례

#### 1. 강의 전체 섹션 프롬프트 생성

```python
# 강의 구조
lecture_sections = [
    {"title": "Introduction", "content": "React Hooks란 무엇인가?"},
    {"title": "Core Concepts", "content": "useState와 useEffect 사용법"},
    {"title": "Advanced", "content": "Custom Hooks 만들기"},
    {"title": "Examples", "content": "실전 예제 5가지"}
]

# 배치 생성
prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=lecture_sections,
    style_strategy="progressive"  # 점진적으로 변화
)

# 결과 활용
for section in lecture_sections:
    section["image_prompt"] = prompts[section["title"]]
    
save_lecture(lecture_sections)
```

#### 2. 책 챕터 전체 이미지 프롬프트 생성

```python
# 책 챕터 구조
chapter = {
    "title": "Chapter 3: 리스트와 튜플",
    "sections": [
        {"title": "Hero", "content": "파이썬 데이터 구조 소개"},
        {"title": "Basic", "content": "리스트 기본 연산"},
        {"title": "Advanced", "content": "성능 비교와 선택 기준"},
        {"title": "Mistakes", "content": "흔한 실수들"}
    ]
}

# 스타일 전략: 섹션마다 다르게
style_map = {
    "Hero": "hand-drawn-sketch",
    "Basic": "minimalist-flat",
    "Advanced": "infographic-vector",
    "Mistakes": "comic-educational"
}

# 배치 생성
prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=chapter["sections"],
    style_strategy="varied"
)

# 또는 개별 스타일 지정
for section in chapter["sections"]:
    section["style"] = style_map.get(section["title"], "auto")
    
prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=chapter["sections"]
)
```

---

<a name="customization"></a>
## 🎨 커스터마이징

### 커스텀 스타일 템플릿 추가

```python
# 커스텀 스타일 정의
custom_style = {
    "name": "watercolor-artistic",
    "template": """
    "{주제}의 수채화 예술 스타일 삽화입니다.
    부드러운 수채화 질감과 투명한 색상 레이어,
    {색상1}에서 {색상2}로의 자연스러운 번짐 효과,
    종이 질감과 물 번짐이 느껴지는 예술적 표현.
    한글 라벨은 깔끔한 서체로 명확하게 표시되며,
    예술적이면서도 교육용으로 적합한 디자인.
    """
}

# Skill에 커스텀 스타일 등록 (API 확장 시)
Skill("yoda-educational-image-prompts").register_custom_style(custom_style)

# 사용
prompt = Skill("yoda-educational-image-prompts").generate(
    content="인상주의 화풍의 역사",
    style="watercolor-artistic"
)
```

### 플랫폼별 최적화 오버라이드

```python
# 플랫폼별 프롬프트 변환
dalle_prompt = Skill("yoda-educational-image-prompts").generate(
    content="React Hooks 소개",
    style="hand-drawn-sketch",
    platform="dalle-3"  # 한국어 프롬프트 그대로
)

midjourney_prompt = Skill("yoda-educational-image-prompts").generate(
    content="React Hooks 소개",
    style="hand-drawn-sketch",
    platform="midjourney"  # 영어 프롬프트로 변환
)
```

---

## 📚 관련 문서

- [STYLES.md](STYLES.md) - 8가지 비주얼 스타일 완전 가이드
- [EXAMPLES.md](EXAMPLES.md) - 50+ 실전 예제
- [PLATFORM_OPTIMIZATION.md](PLATFORM_OPTIMIZATION.md) - 플랫폼별 최적화
- [ACCESSIBILITY.md](ACCESSIBILITY.md) - WCAG AA 접근성 가이드

---

**문서 최종 업데이트**: 2025-11-22  
**버전**: 2.0.0  
**상태**: ✅ 프로덕션 준비 완료

이 문서는 yoda-educational-image-prompts Skill의 완전한 API 레퍼런스와 고급 기능을 제공합니다.
