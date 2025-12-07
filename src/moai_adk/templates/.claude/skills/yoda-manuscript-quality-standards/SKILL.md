---
name: yoda-manuscript-quality-standards
version: 2.0.0
category: quality
status: active
description: YodA 원고 품질 자동 검증 시스템 - writing-guidelines.md v3.0.0 준수 여부 실시간 검증 및 품질 점수 산정
allowed-tools: Read, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, WebSearch
created: 2025-11-28
updated: 2025-11-28
tags: quality-assurance, validation, writing-guidelines, manuscript, compliance, template-system
primary-agents: yoda-book-author
dependencies: yoda-writing-templates, yoda-korean-technical-book-writing, yoda-educational-image-prompts, mcp-context7-integrator
---

# yoda-manuscript-quality-standards Skill

**YodA 원고 품질 자동 검증 시스템 (writing-guidelines.md v3.0.0)**

---

## 🎯 개요 (30초 Quick Reference)

**목적**: `writing-guidelines.md v3.0.0`의 모든 규칙이 원고에 제대로 적용되었는지 자동으로 검증하고 품질 점수를 산정합니다.

**핵심 가치**:
- ✅ **3축 템플릿 시스템 준수 검증**: PART + 장르 + 섹션 자동 적용 여부 확인
- ✅ **분량 기준 템플릿별 검증**: PART/장르별 템플릿 기준 엄격 준수 확인
- ✅ **자동 점수 산정 시스템**: 0-100점 척도로 종합 품질 평가 및 보고서 생성
- ✅ **자동 개선 메커니즘**: 준수 실패 시 자동 수정 및 재시도
- ✅ **실시간 검증**: 작성 과정에서 실시간 품질 준수 여부 확인

**적용 대상**:
- yoda-book-author 에이전트 (자동 강제 적용)
- 모든 /yoda:book:* 커맨드 (자동 검증 통합)
- writing-guidelines.md v3.0.0 준수 (강제)

---

## 📊 Quick Reference: 검증 시스템 (v3.0.0)

### 🔍 자동 검증 시스템

**writing-guidelines.md v3.0.0** 기반 검증 항목:

#### ✅ 템플릿 시스템 준수 (40점)
1. **PART 템플릿 적용** (10점)
   - 챕터 번호에 따른 올바른 PART 템플릿 선택
   - 분량 범위 준수 (basic-tutorial: 1500-2000자 등)

2. **장르 템플릿 적용** (10점)
   - 제목 키워드 분석 기반 장르 자동 선택
   - 장르별 콘텐츠 분포 비율 준수

3. **섹션 템플릿 완비** (10점)
   - 5개 필수 섹션 모두 포함
   - 각 섹션별 구조 및 분량 기준 준수

4. **이전 장 연관성** (10점)
   - 이전 장 요약 참고 및 중복 제거
   - 용어 일관성 및 연결고리 명시

#### ✅ 품질 기준 준수 (30점)
1. **콘텐츠 깊이** (10점)
   - 모든 문단 최소 3문장 이상
   - 구체적 예시 및 비유 포함
   - AI 표현 완전 제거

2. **인용 시스템** (10점)
   - 통계/사실/기술 사양 모두 인용
   - 공식 출처만 사용
   - 인용문 섹션 완비

3. **분량 균형** (10점)
   - 장르별 이론/실습/분석 비율 준수
   - 코드 예제 수 준수

#### ✅ 한국어 작성 스타일 (30점)
1. **문체 일관성** (10점)
   - 전체 책 통해 반말(한다체) 유지
   - 문장 길이 평균 20-30단어 준수

2. **전문 용어 처리** (10점)
   - 한글 우선 + 영문 병기 (첫 등장 시)
   - 기술 용어 일관성 유지

3. **시각 자료** (10점)
   - 핸드 드로잉 표준 템플릿 사용
   - 최소 5개 Mermaid 다이어그램
   - 최소 4개 이미지 프롬프트

---

## 📊 품질 점수 산정 알고리즘

### 자동 검증 및 점수 계산

```python
def calculate_quality_score(content, guidelines, templates):
    """종합 품질 점수 산정 (0-100점) - writing-guidelines.md v3.0.0 기반"""

    score_breakdown = {
        'template_compliance': 0,  # 40점 만점
        'quality_standards': 0,     # 30점 만점
        'korean_writing': 0,         # 30점 만점
        'total_score': 0
    }

    # 템플릿 준수 검증 (40점)
    template_score = validate_template_compliance(content, templates)
    score_breakdown['template_compliance'] = min(template_score, 40)

    # 품질 기준 검증 (30점)
    quality_score = validate_quality_standards(content, guidelines)
    score_breakdown['quality_standards'] = min(quality_score, 30)

    # 한국어 작성 스타일 검증 (30점)
    writing_score = validate_korean_writing_style(content)
    score_breakdown['korean_writing'] = min(writing_score, 30)

    # 종합 점수 계산
    score_breakdown['total_score'] = sum([
        score_breakdown['template_compliance'],
        score_breakdown['quality_standards'],
        score_breakdown['korean_writing']
    ])

    return score_breakdown
```

### 품질 등급 판정

```python
def determine_quality_grade(score):
    """품질 등급 판정"""
    if score >= 90:
        return "A+", "최우수"
    elif score >= 80:
        return "A", "우수"
    elif score >= 70:
        return "B", "양호"
    elif score >= 60:
        return "C", "보통"
    else:
        return "D", "개선 필요"
```

---

## 🚨 자동 개선 메커니즘

### 준수 실패 시 자동 수정

```python
def auto_improve_compliance(content, validation_errors):
    """준수 실패 항목 자동 개선 - writing-guidelines.md v3.0.0 기반"""

    improved_content = content

    for error in validation_errors:
        if error.type == "word_count_violation":
            improved_content = adjust_content_length(
                improved_content, error.target_range
            )
        elif error.type == "missing_section":
            improved_content = add_missing_section(
                improved_content, error.section_template
            )
        elif error.type == "ai_expressions":
            improved_content = remove_ai_expressions(improved_content)
        elif error.type == "citation_missing":
            improved_content = add_required_citations(improved_content)
        elif error.type == "template_mismatch":
            improved_content = adjust_template_distribution(
                improved_content, error.template_requirements
            )

    return improved_content
```

---

## 📈 검증 보고서 생성

### 자동 검증 보고서 형식

```markdown
# YodA 원고 품질 검증 보고서 (v3.0.0)

## 📊 종합 평가

- **종합 점수**: 85/100점
- **등급**: A (우수)
- **검증 시간**: 2025-11-28 14:30:00
- **가이드라인 버전**: writing-guidelines.md v3.0.0

## 🔍 분야별 점수

| 평가 영역 | 점수 | 만점 | 준수율 | 평가 |
|----------|------|------|--------|------|
| 템플릿 준수 | 35점 | 40점 | 87.5% | 우수 |
| 품질 기준 | 25점 | 30점 | 83.3% | 우수 |
| 한국어 스타일 | 25점 | 30점 | 83.3% | 우수 |

## ✅ 준수 항목

- PART 템플릿 자동 적용 완료
- 5개 필수 섹션 모두 포함
- 사전 자료 조사 완료 (Context7 + WebSearch)
- 인용 시스템 부분 적용

## ⚠️ 개선 필요 항목

1. **인용 시스템 보강** (우선순위: 높음)
   - 누락된 통계 2건 발견
   - 인용문 섹션 보강 필요

## 🔧 자동 개선 실행

실행 결과: 5개 항목 자동 수정 완료
최종 점수: 85점 → 92점 (+7점 향상)
```

---

## 🛠️ 사용 방법

### 에이전트 자동 통합

```python
# yoda-book-author 에이전트에서 자동 검증 실행
from yoda_manuscript_quality_standards import validate_manuscript_v3

# 원고 작성 후 자동 검증 (writing-guidelines.md v3.0.0)
validation_result = validate_manuscript_v3(
    content=generated_chapter,
    guidelines_path='.moai/yoda/books/claude-code-agentic-coding-master/writing-guidelines.md',
    templates=applied_templates,
    chapter_number=chapter_num,
    chapter_title=chapter_title
)

# 결과 보고
if validation_result['total_score'] >= 80:
    print(f"✅ 품질 검증 통과: {validation_result['total_score']}/100점 ({validation_result['grade']})")
    print(f"✅ 템플릿 준수율: {validation_result['template_compliance']}/40점")
else:
    print(f"⚠️ 품질 개선 필요: {validation_result['total_score']}/100점 ({validation_result['grade']})")

    # 자동 개선 실행
    improved_content = auto_improve_compliance(content, validation_result['errors'])
```

### 강제 적용 시스템

**모든 /yoda:book:* 커맨드는 자동으로 검증을 실행**:

```python
# 커맨드 실행 시 자동으로 검증 통합
Task(subagent_type='yoda-book-author',
     prompt=f"""
     WRITING_GUIDELINES: {load_guidelines_v3()}
     ENFORCE_COMPLIANCE: true
     AUTO_VALIDATION: true
     QUALITY_THRESHOLD: 80

     Chapter parameters: {chapter_params}
     """)
```

---


---

## Detailed Guidelines

For detailed implementation guidelines covering all 6 quality aspects, see [Detailed Guidelines](modules/detailed-guidelines.md).

Key Topics:
- 3축 템플릿 시스템 준수 검증
- 인용 시스템
- 내용 분량 기준
- 품질 검사 기준 (5개 체크리스트)
- 문체 및 스타일
- 이미지 및 다이어그램


## 🔄 작성 워크플로우 (4단계)

### Phase 1: 자료 조사 (30-40%)

```
1. Context7 MCP로 공식 문서 조사
   - mcp__context7__resolve-library-id()
   - mcp__context7__get-library-docs()

2. WebSearch로 최신 자료 조사 (공식 출처만)
   - allowed_domains 필터 적용

3. 조사 결과 문서화
   - 출처, 제목, URL, 발행일 기록

4. 인용 목록 초안 작성
   - 인용 번호 순서대로 정리
```

### Phase 2: 초안 작성 (40-50%)

```
1. 개요 작성 (주요 섹션 구조)
2. 각 소절 작성 (최소 500 단어)
3. 코드 예제 추가 (기초→응용→고급)
4. 인용 번호 삽입 (본문에 (1), (2) 추가)
```

### Phase 3: 품질 검사 (10-20%)

```
1. 자동 체크리스트 검증 (5개)
2. AI 표현 제거
3. 할루시네이션 검증
4. 분량 확인
5. 인용문 섹션 완성
```

### Phase 4: 최종 검토 (5-10%)

```
1. 전체 흐름 확인
2. 문체 일관성 점검
3. 이미지 프롬프트 추가
4. 최종 저장
```

📖 **상세**: `@reference/workflow.md`

---

## 🤖 yoda-book-author 에이전트 통합

### 자동 로드 시점

**Step 2: 자료 조사** (워크플로우 Phase 1)

```python
# 품질 기준 로드
quality_standards = load_skill("yoda-manuscript-quality-standards")

# 공식 출처 확인
official_sources = quality_standards.get_allowed_domains()
# → ['anthropic.com', 'docs.anthropic.com', 'github.com/anthropics']

# WebSearch with 공식 도메인 필터
research_results = WebSearch(
    query="Claude Code 최신 기능",
    allowed_domains=official_sources
)

# 인용 시스템 자동 적용
citations = quality_standards.format_citations(research_results)
```

### 품질 검증 시점

**Step 6: 품질 검사** (워크플로우 Phase 3)

```python
# 5개 체크리스트 자동 검증
quality_report = quality_standards.validate_section(
    content=section_content,
    checklist_items=[
        "citation_system",
        "content_quantity",
        "depth_specificity",
        "ai_expression_removal",
        "hallucination_check"
    ]
)

if not quality_report.passed:
    print(f"품질 검증 실패: {quality_report.issues}")
    # 자동 수정 또는 사용자 경고
```

---

## 📈 기대 효과

| 메트릭 | 현재 | 목표 | 개선율 |
|--------|------|------|--------|
| **인용 시스템 준수** | 90% | 100% | +11% |
| **공식 출처 비율** | 70% | 100% | +43% |
| **품질 검증 통과율** | 75% | 95% | +27% |
| **작성 시간** | 25분/절 | 20분/절 | -20% |
| **할루시네이션 제거** | 85% | 99% | +16% |

---

## 🔗 Works Well With

이 스킬은 다음 스킬들과 함께 사용됩니다:

- **yoda-writing-templates** - 3축 템플릿 시스템 (구조 제공)
- **yoda-korean-technical-book-writing** - 한국어 작성 모범 사례 (문체 제공)
- **yoda-educational-image-prompts** - 이미지 생성 프롬프트
- **moai-quality-gate** - TRUST 5 품질 게이트
- **mcp-context7-integrator** - 최신 문서 조사

**역할 분리**:
| 스킬 | 역할 |
|------|------|
| yoda-writing-templates | 구조 (PART+장르+섹션) |
| yoda-korean-technical-book-writing | 문체 (반말, 용어, KLI) |
| **yoda-manuscript-quality-standards** | **품질 (분량, 인용, 검증, 워크플로우)** |

---

## 📚 상세 참고 자료

### reference/ 디렉토리 (7개 파일)

각 영역의 상세한 가이드를 제공합니다:

1. `research-process.md` - Context7 MCP, WebSearch 사용법, 공식 출처 검증
2. `citation-system.md` - 인용 번호 규칙, 인용문 형식, 인용 예시
3. `content-quantity.md` - 분량 기준, 문단 구조, 깊이 있는 설명 예시
4. `quality-checklists.md` - 5개 체크리스트 상세, 자동 검증 방법
5. `style-guide.md` - 반말 규칙, 금지 표현, 친근한 톤
6. `visual-elements.md` - 핸드 드로잉 템플릿, Mermaid 다이어그램
7. `workflow.md` - 4단계 워크플로우 상세 설명

### templates/ 디렉토리 (3개 템플릿)

작성 시 바로 사용 가능한 템플릿:

1. `research-template.md` - 자료 조사 문서화 템플릿
2. `citation-template.md` - 인용문 섹션 템플릿
3. `quality-checklist.md` - 작성 완료 후 체크리스트

### examples/ 디렉토리 (2개 예시)

실제 적용 예시:

1. `good-section-example.md` - 올바른 절 구성 예시
2. `bad-section-example.md` - 잘못된 절 구성 예시 (Before/After)

---

## ✅ 최종 체크리스트

작성 완료 후 다음 항목을 모두 확인:

- [ ] 사전 자료 조사 완료 (Context7 + WebSearch)
- [ ] 인용 번호 모두 추가 (통계, 사실, 기술 사양)
- [ ] 인용문 섹션 완성 (출처, 제목, URL, 날짜)
- [ ] 분량 기준 충족 (1,500+ 단어)
- [ ] AI 표현 제거 (과장, 추상, 할루시네이션)
- [ ] 반말 일관성 (한다체)
- [ ] 이미지 프롬프트 추가 (핸드 드로잉 템플릿)
- [ ] 코드 예제 3개 이상
- [ ] 깊이 있는 설명 (최소 3문장/문단)

---

**버전**: 1.0.0
**최종 업데이트**: 2025-11-27
**적용 대상**: 모든 YodA 책 프로젝트
**준수 필수**: ✅ 강제 적용
**다음 업데이트**: Claude Code v4.0 기능 반영 (예정)
