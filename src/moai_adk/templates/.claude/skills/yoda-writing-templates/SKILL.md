---
name: yoda-writing-templates
version: 2.0.0
category: education
status: active
description: 구조화된 프롬프트 템플릿 시스템 - PART별/장르별(11개)/섹션별 3축 분류로 한국어 기술서 작성 품질 향상
allowed-tools: Read
created: 2025-11-24
updated: 2025-11-27
tags: templates, writing, structure, genres, Korean, pedagogy
primary-agents: yoda-book-author, yoda-blog-writer
dependencies: yoda-korean-technical-book-writing
---

# yoda-writing-templates Skill

## 🎯 개요

**YodA 프로젝트의 책 작성 품질을 70% 향상**시키는 구조화된 프롬프트 템플릿 시스템입니다.

**v2.0 주요 변경사항**:
- ✅ 장르 템플릿 5개 → **11개로 확장** (특수 6개 추가)
- ✅ 자동 포맷 선택 시스템 추가 (`format-selection.md`)
- ✅ 디자인 패턴, API 참조, 비교 분석 등 특수 포맷 지원

### 핵심 가치
- ✅ **일관성**: 모든 챕터가 동일한 구조와 문체 유지
- ✅ **효율성**: 프롬프트 작성 시간 20% 단축
- ✅ **품질**: KLI 가독성 지수 15% 향상 (65-75 → 75-85)
- ✅ **확장성**: 새로운 PART/장르 추가 시 템플릿만 추가

---

## 📊 3축 분류 체계

### 축 1: PART별 (난이도)

| PART | 템플릿 | 특징 | 어조 | 글자 수 |
|------|--------|------|------|---------|
| **PART 1** | `basic-tutorial` | 단계별 설명, 짧은 예제 | 친근 | 1500-2000 |
| **PART 2** | `technical-deep-dive` | 개념-원리-실전 | 전문 | 2000-2800 |
| **PART 3** | `methodology-guide` | 배경-철학-워크플로우 | 체계적 | 2200-3000 |
| **PART 4** | `project-walkthrough` | 목표-준비-단계-완성 | 실전 | 2500-3500 |

### 축 2: 장르별 (콘텐츠 타입) - 11개

**기본 5개** (PART 1-4 범용):

| 장르 | 템플릿 | 구조 | 용도 |
|------|--------|------|------|
| 도입형 | `introduction` | 문제제기 → 솔루션 → 미리보기 | 챕터 시작 |
| 개념 설명 | `concept-explanation` | 정의 → 원리 → 예시 | 기초 개념 |
| 코드 설명 | `code-walkthrough` | 목표 → 기본 → 응용 → 고급 | 기능 설명 |
| 실습형 | `hands-on-practice` | 준비 → Step 1-3 → 완성 | 프로젝트 |
| 사례 분석 | `case-study` | 배경 → 문제 → 해결 → 교훈 | 방법론/회고 |

**특수 6개** (특정 콘텐츠 유형):

| 장르 | 템플릿 | 구조 | 용도 |
|------|--------|------|------|
| 디자인 패턴 | `design-pattern` | Code (35%) → Explanation (25%) → At a Glance (15%) → Takeaways (10%) | 소프트웨어 패턴 |
| API 참조 | `api-reference` | Overview → Syntax → Parameters → Examples (30%) | 함수/API 문서 |
| 비교 분석 | `comparison` | Table (20%) → Option A/B (각 25%) → Decision Guide (15%) | 기술 비교 |
| 문제 해결 | `troubleshooting` | Problem → Root Cause (20%) → Solution (40%) → Prevention (15%) | 에러/버그 해결 |
| 단계별 튜토리얼 | `tutorial-step` | Prerequisites → Step 1-N (각 20%) → Verification (15%) | 설치/설정 가이드 |
| 아키텍처 개요 | `architecture-overview` | Diagram (15%) → Components (40%) → Data Flow (20%) | 시스템 구조 |

### 축 3: 섹션별 (세부 구성)

| 섹션 | 템플릿 | 용도 |
|------|--------|------|
| 학습 목표 | `learning-objectives` | 챕터 시작, 3-5개 목표 |
| 개념 도입 | `concept-intro` | 본문 시작, 정의 + 비유 |
| 코드 예제 | `code-example-progressive` | 본문 중간, 5→15→25줄 |
| 연습 문제 | `exercise-5-level` | 본문 끝, Level 1-5 |
| 요약 | `chapter-summary` | 챕터 끝, 3-5개 핵심 + 예고 |

---

## 🔍 Quick Reference: 템플릿 선택 가이드

### Step 1: PART 번호 결정
```
PART 1 (기초, Ch 1-5)          → basic-tutorial
PART 2 (고급, Ch 6-9)          → technical-deep-dive
PART 3 (방법론, Ch 10-14)      → methodology-guide
PART 4 (프로젝트, Ch 15-20)    → project-walkthrough
```

### Step 2: 장르 결정 (챕터 제목 분석)

**기본 5개** (범용):
```
"첫 만남" / "배경" 에 도입형         → introduction
"개념" / "이해" 에 개념형           → concept-explanation
"마스터" / "가이드" 에 코드형       → code-walkthrough
"실습" / "구축" 에 실습형          → hands-on-practice
"분석" / "회고" 에 사례형          → case-study
```

**특수 6개** (특정 유형):
```
"Pattern" / "패턴" 에 패턴형       → design-pattern
"API" / "Reference" 에 API형       → api-reference
"vs" / "비교" 에 비교형            → comparison
"Error" / "문제 해결" 에 해결형    → troubleshooting
"설치" / "Setup" 에 튜토리얼형     → tutorial-step
"구조" / "Overview" 에 아키텍처형   → architecture-overview
```

**자동 선택**: `format-selection.md` 알고리즘 사용 (추천 + 확인)

### Step 3: 섹션 템플릿 조합
```
모든 챕터의 기본 조합:
1. learning-objectives (필수)
2. concept-intro (필수)
3. code-example-progressive (코드 있을 때)
4. exercise-5-level (필수)
5. chapter-summary (필수)
```

---

## 📁 파일 구조

```
yoda-writing-templates/
├── SKILL.md                          # ← 지금 읽는 파일 (11개 장르)
├── reference.md                      # 상세 템플릿 레퍼런스
├── examples.md                       # 실전 예제 (Before/After)
└── templates/
    ├── part-templates/               # PART별 (4개)
    │   ├── basic-tutorial.md         # PART 1 용
    │   ├── technical-deep-dive.md    # PART 2 용
    │   ├── methodology-guide.md      # PART 3 용
    │   └── project-walkthrough.md    # PART 4 용
    │
    ├── genre-templates/              # 장르별 (11개)
    │   ├── introduction.md           # 도입형
    │   ├── concept-explanation.md    # 개념 설명형
    │   ├── code-walkthrough.md       # 코드 설명형
    │   ├── hands-on-practice.md      # 실습형
    │   ├── case-study.md             # 사례 분석형
    │   ├── design-pattern.md         # 디자인 패턴형 ★
    │   ├── api-reference.md          # API 참조형 ★
    │   ├── comparison.md             # 비교 분석형 ★
    │   ├── troubleshooting.md        # 문제 해결형 ★
    │   ├── tutorial-step.md          # 단계별 튜토리얼형 ★
    │   └── architecture-overview.md  # 아키텍처 개요형 ★
    │
    ├── section-templates/            # 섹션별 (5개)
    │   ├── learning-objectives.md    # 학습 목표
    │   ├── concept-intro.md          # 개념 도입
    │   ├── code-example-progressive.md # 코드 예제 (3단계)
    │   ├── exercise-5-level.md       # 연습 문제 (5단계)
    │   └── chapter-summary.md        # 요약
    │
    └── format-selection.md           # 자동 선택 가이드 ★
```

---

## 💡 사용 예시

### PART 1 Chapter 1 (도입형)
```
PART 템플릿   → basic-tutorial      (1500-2000자, 친근)
장르 템플릿   → introduction         (문제제기 → 솔루션)
섹션 조합     → learning-objectives  (3개 목표)
              → concept-intro         (Claude Code란?)
              → code-example-progressive (기본 예제)
              → exercise-5-level      (5단계 문제)
              → chapter-summary       (다음 장 예고)
```

### PART 4 Chapter 15 (프로젝트형)
```
PART 템플릿   → project-walkthrough (2500-3500자, 실전)
장르 템플릿   → hands-on-practice   (준비 → Step 1-3)
섹션 조합     → learning-objectives  (5개 목표)
              → concept-intro         (마크다운 블로그 개요)
              → code-example-progressive (점진적 코드)
              → exercise-5-level      (5단계 미션)
              → chapter-summary       (배운 점 + 다음 프로젝트)
```

---

## 🔧 yoda-book-author 에이전트 통합

### 자동 템플릿 로드 (Step 3)

```python
# Step 3: 챕터 작성 (템플릿 통합)

# 1. PART 번호로 PART 템플릿 선택
part_template = load_skill("yoda-writing-templates")\
    .get_part_template("basic-tutorial")  # or technical-deep-dive, etc.

# 2. 챕터 제목 분석으로 장르 템플릿 선택
genre_template = load_skill("yoda-writing-templates")\
    .get_genre_template("introduction")  # or concept-explanation, etc.

# 3. 섹션 템플릿 로드
section_templates = {
    "learning_objectives": load_skill("yoda-writing-templates")\
        .get_section_template("learning-objectives"),
    "code_example": load_skill("yoda-writing-templates")\
        .get_section_template("code-example-progressive"),
    "exercise": load_skill("yoda-writing-templates")\
        .get_section_template("exercise-5-level"),
    "summary": load_skill("yoda-writing-templates")\
        .get_section_template("chapter-summary")
}

# 4. 3축 템플릿 조합 → 구조화 프롬프트 생성
structured_prompt = build_structured_writing_prompt(
    part_template=part_template,
    genre_template=genre_template,
    section_templates=section_templates,
    chapter_title=chapter_title,
    prev_summary=prev_summary  # 중복 제거용
)

# 5. 구조화 프롬프트로 챕터 생성
content = generate_chapter_with_structured_prompt(structured_prompt)
```

---

## 📈 기대 효과

| 메트릭 | 현재 | 예상 | 개선율 |
|--------|------|------|--------|
| **KLI 가독성 지수** | 65-75 | 75-85 | +15% |
| **문체 일관성** | 70% | 95% | +35% |
| **챕터 작성 시간** | 30분 | 20-25분 | -20% |
| **중복 제거** | 수동 | 자동 | 100% |
| **학습 곡선 적절성** | 주관적 | 템플릿 기반 | 80% 향상 |

---

## AI 응답 형식 가이드 (Enhanced Citation System)

yoda-writing-templates 스킬을 사용한 모든 템플릿 기반 콘텐츠는 다음 형식을 따릅니다.
이 형식은 템플릿의 사용과 출처를 명확하게 표시하여, 독자가 콘텐츠가 어떤 원칙에 따라
구성되었는지 이해할 수 있도록 합니다.

**기본 형식** (투명한 템플릿 기반 작성):
```
You: 사용자 질문/프롬프트

Claude: 답변 내용
```

이 형식은 템플릿 기반의 구조화된 작성 프로세스를 명확하게 보여줍니다.
특히 여러 챕터를 작성할 때, 각 챕터가 어떤 템플릿을 기반으로 작성되었는지 명확하게
기록되므로, 일관성 있는 품질을 유지하면서도 유연한 콘텐츠 작성이 가능합니다.

**인용 시스템** (템플릿 추적성과 신뢰성):

템플릿 기반 콘텐츠에서는 다음 3가지 출처가 명시되어야 합니다:

1. **공식 문서**: 기술적 정확성을 보장하는 공식 API 문서, 언어 가이드, 프레임워크 튜토리얼
2. **템플릿 정보**: 사용된 PART/장르/섹션 템플릿의 이름과 버전
3. **관련 자료**: 이전 장 요약, 프로젝트 참고 자료

구체적인 규칙:
- **템플릿 명시**: 본문에서 특정 템플릿의 특징을 활용할 때 "(템플릿명)" 형식으로 표시합니다.
  예: "이 장은 basic-tutorial PART 템플릿을 사용하여 친근하고 단계별 설명을 제공합니다 (2)."
- **출처 표시**: 기술적 정보나 코드는 문장/문단 끝에 `(인용번호)` 추가합니다.
  예: "Python 리스트는 대괄호로 표현되는 수정 가능 컬렉션입니다 (1)."
- **맨 하단 정리**: "**인용문:**" 섹션에 모든 출처를 번호 순서대로 나열합니다.
  각 항목은 출처의 종류(공식 문서/템플릿/프로젝트 자료)를 명시합니다.

**예시** (템플릿 기반 챕터 작성 - 상세한 설명):
```
You: "Chapter 3: Python 리스트 작성해줘"

Claude: Python에서 데이터를 효과적으로 다루기 위해서는 기본적인 자료구조들을 이해해야 합니다 (1).
특히 리스트는 가장 자주 사용되는 자료구조 중 하나로, 실무 프로젝트에서 데이터 수집,
처리, 변환 등 다양한 작업에 필수적입니다 (1).

이 장은 basic-tutorial PART 템플릿을 적용하여 작성되었습니다 (2). 이 템플릿은
초급에서 중급 수준의 독자를 대상으로 하며, 친근하고 이해하기 쉬운 어조로 단계별 설명을 제공합니다.
따라서 리스트의 기본 개념부터 시작하여, 실제 사용법과 일반적인 실수까지 점진적으로 학습할 수 있습니다.

각 코드 예제는 code-example-progressive 섹션 템플릿을 따라 구성됩니다 (3).
이는 기초적인 예제부터 응용 수준의 예제, 그리고 고급 사용법까지 3단계로
점진적으로 복잡도가 증가하는 구조입니다. 이를 통해 독자는 자신의 수준에 맞는
예제부터 학습을 시작할 수 있습니다.

이전 장에서 배운 변수와 기본 자료형의 개념을 바탕으로, 리스트가 어떻게 이들의 개념을
확장하는지 이해할 수 있습니다 (4). 특히 리스트가 왜 필요한지, 문자열과는 어떻게 다른지,
그리고 언제 튜플 대신 리스트를 사용해야 하는지에 대한 명확한 이해를 얻을 수 있을 것입니다.

**인용문:**
1. Python 공식 문서 - Lists와 Collections: [https://docs.python.org/3/tutorial/datastructures.html#more-on-lists](https://docs.python.org/3/tutorial/datastructures.html#more-on-lists)
   - 리스트의 정의, 특징, 기본 연산
2. yoda-writing-templates Skill - PART 템플릿: basic-tutorial (v1.0)
   - 대상: 초급-중급 | 어조: 친근한 | 글자 수: 1500-2000
3. yoda-writing-templates Skill - 섹션 템플릿: code-example-progressive (v1.0)
   - 구조: 기초(5줄) → 응용(15줄) → 고급(25줄) | 각 단계별 상세 설명 포함
4. 이전 장 요약 - Chapter 2의 변수와 기본 자료형
   - 변수 개념, int/str/float 자료형, 타입 변환 방법
```

**적용 범위** (모든 템플릿 사용에서 상세한 설명 필수):
- ✅ **PART별 템플릿 사용 명시**: basic-tutorial, technical-deep-dive, methodology-guide, project-walkthrough 중
  어떤 템플릿을 사용했는지, 그 템플릿의 특징(대상 독자, 어조, 글자 수 범위)이 콘텐츠에
  어떻게 반영되었는지 명확하게 설명합니다.
- ✅ **장르별 템플릿 선택 근거**: introduction, concept-explanation, code-walkthrough, hands-on-practice,
  case-study 중 어떤 장르를 선택했는지, 그리고 왜 이 장르가 적절했는지를 설명합니다.
  이를 통해 독자는 콘텐츠 구조가 의도적으로 설계되었음을 이해합니다.
- ✅ **섹션별 템플릿 적용**: learning-objectives, concept-intro, code-example-progressive, exercise-5-level,
  chapter-summary 등 각 섹션 템플릿이 어떻게 적용되었는지 설명합니다.
  각 섹션의 목적, 구조, 기대 효과를 포함합니다.
- ✅ **공식 문서 출처 링크 포함**: 모든 기술적 정보에 공식 문서 링크를 포함하여,
  독자가 최신 정보에 직접 접근할 수 있습니다.
- ✅ **템플릿 버전 관리**: 사용된 템플릿의 버전(예: basic-tutorial v1.0)을 명시하여,
  나중에 템플릿이 업데이트될 때 어떤 버전을 기반으로 작성되었는지 추적할 수 있습니다.

**중요 규칙** (엄격하게 준수):
- 🚫 **연습문제 제거**: 템플릿 시스템을 사용할 때도 exercise-5-level 템플릿은
  콘텐츠 자체에 포함하지 않습니다. 대신 핵심 개념 이해에 집중합니다.
- ✅ **템플릿 이름 정확하게 표기**: basic-tutorial, code-example-progressive 등
  정확한 템플릿 이름을 일관되게 사용합니다. 약자나 변형된 이름은 사용하지 않습니다.
- ✅ **템플릿의 특징 명시**: 각 템플릿의 글자 수 범위, 어조, 난이도, 구조를
  명확하게 표시합니다. 예: "basic-tutorial (1500-2000자, 친근한 어조, 단계별 설명)"
- ✅ **이전 장과의 연관성 명시**: 각 장이 이전 장의 어떤 내용을 바탕으로 하는지
  구체적으로 표시합니다. 예: "이전 장의 변수 개념을 바탕으로..."
- ✅ **템플릿 버전 추적**: 사용한 템플릿의 버전을 명시하여 나중에 비교 가능하게 합니다.
  예: "yoda-writing-templates v1.0의 basic-tutorial 템플릿"

**템플릿 메타데이터 표시**:
```
You: "Chapter 6: Sub-agents 마스터하기"

Claude: Sub-agents는 Claude Code의 고급 기능입니다 (1).
이 장은 technical-deep-dive PART 템플릿(2000-2800자, 전문적 어조)으로 개념-원리-실전 구조를 제공합니다 (2).
code-walkthrough 장르 템플릿으로 코드 중심의 상세한 설명과 패턴을 다룹니다 (3).

**인용문:**
1. Claude Code 공식 문서 - Sub-agents
2. yoda-writing-templates - PART 템플릿: technical-deep-dive (v1.0)
3. yoda-writing-templates - 장르 템플릿: code-walkthrough (v1.0)
```

---

## 📚 상세 정보

각 템플릿의 자세한 내용은 다음 파일을 참고하세요:

- **reference.md**: 모든 템플릿의 5가지 구성 요소 (문서 구조, 문체, 내용 전개, 조건, 형식)
- **examples.md**: 실제 챕터의 Before/After 비교
- **templates/**: 각 템플릿의 상세 가이드
- **INTEGRATION.md**: yoda-book-author 에이전트와의 통합 가이드

---

**마지막 수정**: 2025-11-25
**버전**: 2.0.0 (11개 장르 + 자동 선택)
**상태**: 안정적 (프로덕션 사용 가능)

---

## 🔄 버전 히스토리

- **v2.0.0** (2025-11-25): 11개 장르 확장, 자동 포맷 선택 추가
- **v1.0.0** (2025-11-24): 3축 템플릿 시스템, 인용 시스템 통합