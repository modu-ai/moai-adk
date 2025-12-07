---
name: yoda-educational-image-prompts
version: 2.0.0
category: education
status: active
description: 교육용 콘텐츠를 위한 AI 이미지 생성 프롬프트 템플릿. 8가지 비주얼 스타일로 강의자료와 책 원고에 최적화된 한국어 프롬프트 생성. DALL-E 3, Gemini Imagen, Midjourney, Stable Diffusion 모두 호환.
allowed-tools: Read
created: 2025-11-22
updated: 2025-11-22
tags: image-generation, educational-design, visual-content, ai-prompts, korean, teaching-materials, book-writing
primary-agents: yoda-master, yoda-book-author
dependencies: yoda-content-generator, moai-mermaid-diagram-expert
---

# yoda-educational-image-prompts

**교육용 이미지 생성 프롬프트 Skill**  
강의자료와 책 원고에 포함될 시각 자료용 한국어 프롬프트를 자동 생성합니다.

---

## 🚀 Quick Start (30초)

### 기본 사용법

```python
# 섹션 내용 기반으로 한국어 프롬프트 생성
prompt = Skill("yoda-educational-image-prompts").generate(
    content="React 컴포넌트 생명주기와 Hook의 관계",
    style="auto",
    language="ko"
)

print(prompt)
# 출력:
# "React 컴포넌트 생명주기를 표현한 핸드드로잉 삽화 스타일입니다.
#  크림색 종이 질감 배경에 주황색과 파란색 색연필로 그린 듯한 따뜻한 색감..."
```

### 특정 스타일 지정

```python
# 특정 스타일 선택
prompt = Skill("yoda-educational-image-prompts").generate(
    content="마이크로서비스 아키텍처 다이어그램",
    style="isometric-3d",  # 아이소메트릭
    language="ko",
    context="기술 아키텍처 설명용"
)
```

### 에이전트 자동 사용

```python
# yoda-master와 yoda-book-author가 자동으로 호출
# 강의 생성 시 Introduction, Core Concepts 섹션에 자동 생성
# 책 챕터 작성 시 Hero, Basic, Advanced, Mistakes 섹션에 자동 생성
```

---

## 📚 핵심 기능 Overview

### 1. 8가지 비주얼 스타일

다양한 교육 콘텐츠를 지원하는 8가지 전문 스타일:

| 스타일 | 용도 | 상세 설명 |
|-------|------|----------|
| **핸드드로잉 스케치** | 개념 소개, 첫 강의 | [STYLES.md](STYLES.md#style-1) 참조 |
| **아이소메트릭 3D** | 시스템 아키텍처 | [STYLES.md](STYLES.md#style-2) 참조 |
| **미니멀 플랫** | 프로세스 흐름도 | [STYLES.md](STYLES.md#style-3) 참조 |
| **기술 도면** | 알고리즘 상세 설명 | [STYLES.md](STYLES.md#style-4) 참조 |
| **인포그래픽** | 통계 및 비교 | [STYLES.md](STYLES.md#style-5) 참조 |
| **포토리얼리스틱 3D** | 하드웨어/IoT | [STYLES.md](STYLES.md#style-6) 참조 |
| **만화/코믹** | 단계별 튜토리얼 | [STYLES.md](STYLES.md#style-7) 참조 |
| **그라디언트 현대** | 최신 기술(AI/블록체인) | [STYLES.md](STYLES.md#style-8) 참조 |

**전체 스타일 가이드**: [STYLES.md](STYLES.md)

### 2. 자동 스타일 선택

콘텐츠를 분석해서 최적의 스타일을 자동으로 추천합니다.

```python
# 자동 스타일 선택
prompt = Skill("yoda-educational-image-prompts").generate(
    content="마이크로서비스 아키텍처와 그 장점들",
    style="auto",  # 자동으로 "isometric-3d" 선택
    language="ko"
)
```

### 3. 한국어 프롬프트

모든 프롬프트가 자연스러운 한국어로 작성되며, 한글 렌더링 최적화가 포함됩니다.

**접근성 가이드**: [ACCESSIBILITY.md](ACCESSIBILITY.md)

### 4. 플랫폼 최적화

DALL-E 3, Gemini Imagen, Midjourney, Stable Diffusion 모두에 최적화된 프롬프트를 생성합니다.

**플랫폼별 가이드**: [PLATFORM_OPTIMIZATION.md](PLATFORM_OPTIMIZATION.md)

### 5. 접근성 (WCAG AA)

WCAG AA 기준을 만족하는 이미지 프롬프트로 모든 사용자가 접근 가능합니다.

**접근성 체크리스트**: [ACCESSIBILITY.md](ACCESSIBILITY.md)

---

## 🎯 스타일 선택 가이드 (요약)

### 콘텐츠 타입별 추천

| 콘텐츠 타입 | 추천 스타일 | 대상 |
|-----------|----------|------|
| **개념 소개** | 핸드드로잉 스케치 | 초급 학습자 |
| **시스템 아키텍처** | 아이소메트릭 3D | 중급 이상 |
| **프로세스 흐름** | 미니멀 플랫 | 모든 수준 |
| **상세 기술** | 기술 도면 | 고급 개발자 |
| **비교 및 데이터** | 인포그래픽 | 비즈니스/교육 |
| **물리적 장치** | 포토리얼리스틱 3D | IoT/하드웨어 |
| **단계별 학습** | 만화/코믹 | 초급 학습자 |
| **최신 기술** | 그라디언트 현대 | AI/혁신 기술 |

**전체 선택 가이드**: [STYLES.md](STYLES.md#selection-guide)

### 학습 대상별 추천

**초급 학습자**: 핸드드로잉 ⭐⭐⭐⭐⭐, 만화/코믹 ⭐⭐⭐⭐⭐, 미니멀 플랫 ⭐⭐⭐⭐  
**중급 학습자**: 미니멀 플랫 ⭐⭐⭐⭐⭐, 아이소메트릭 ⭐⭐⭐⭐, 인포그래픽 ⭐⭐⭐⭐  
**고급 학습자**: 기술 도면 ⭐⭐⭐⭐⭐, 포토리얼리스틱 ⭐⭐⭐⭐, 아이소메트릭 ⭐⭐⭐  
**최신 기술**: 그라디언트 현대 ⭐⭐⭐⭐⭐, 인포그래픽 ⭐⭐⭐

---

## 📞 사용 패턴

### 패턴 1: yoda-master에서 자동 사용

```python
# yoda-master 에이전트가 강의를 생성할 때 자동으로 호출
for section in sections:
    if section.type == "introduction":
        image_prompt = Skill("yoda-educational-image-prompts").generate(
            content=section.content,
            style="hand-drawn-sketch",
            language="ko",
            context="lecture hero image"
        )
        section.metadata["image_prompt"] = image_prompt
```

### 패턴 2: yoda-book-author에서 자동 사용

```python
# yoda-book-author가 책 챕터를 작성할 때 자동으로 호출
chapter_content = {
    "title": "리스트와 튜플",
    "hero_image": Skill("yoda-educational-image-prompts").generate(
        content="파이썬 리스트와 튜플의 기본 개념 소개",
        style="hand-drawn-sketch",
        language="ko",
        context="chapter 3 hero image"
    )
}
```

### 패턴 3: 수동 프롬프트 생성

```python
# 사용자가 직접 이미지 프롬프트 생성
prompt = Skill("yoda-educational-image-prompts").generate(
    content="Git 브랜칭 전략 비교: Git Flow vs GitHub Flow",
    style="infographic-vector",
    language="ko"
)

# DALL-E 3에 직접 전달
response = openai.Image.create(
    model="dall-e-3",
    prompt=prompt,
    size="1024x1024",
    quality="hd"
)
```

**더 많은 예제**: [EXAMPLES.md](EXAMPLES.md)

---

## 🔧 고급 기능

### 배치 생성

여러 섹션의 프롬프트를 한 번에 생성합니다.

```python
# 여러 섹션의 프롬프트 배치 생성
prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=[
        {"title": "소개", "content": "..."},
        {"title": "기초", "content": "..."},
        {"title": "고급", "content": "..."}
    ],
    language="ko",
    style_strategy="consistent"  # 일관된 스타일 또는 "varied"
)
```

**API 레퍼런스**: [REFERENCE.md](REFERENCE.md)

---

## 🌍 플랫폼 지원

### 완전 지원 ✅

- **DALL-E 3**: 한국어 프롬프트 그대로 사용 가능
- **Gemini Imagen 3**: 한국어 지원, DALL-E와 유사

### 제한적 지원 ⚠️

- **Midjourney**: 한국어 제한적, 영어 프롬프트 권장
- **Stable Diffusion**: 모델에 따라 다름, 영어 권장

**플랫폼별 최적화**: [PLATFORM_OPTIMIZATION.md](PLATFORM_OPTIMIZATION.md)

---

## ✅ 접근성 검증

모든 프롬프트 생성 시 다음 기준을 자동으로 검증합니다:

- [x] **색상 대비** (WCAG AA): 4.5:1 이상 명도 대비
- [x] **한글 렌더링**: 최소 16pt, 고딕 계열 서체
- [x] **텍스트 밀도**: 일행 최대 35자
- [x] **아이콘과 라벨**: 색상만으로 구분하지 않음

**전체 접근성 가이드**: [ACCESSIBILITY.md](ACCESSIBILITY.md)

---

## 📚 참고 문서

### 상세 가이드

- **[STYLES.md](STYLES.md)** - 8가지 비주얼 스타일 완전 설명 (800줄)
- **[EXAMPLES.md](EXAMPLES.md)** - 50+ 실전 예제 (900줄)
- **[PLATFORM_OPTIMIZATION.md](PLATFORM_OPTIMIZATION.md)** - 플랫폼별 최적화 (400줄)
- **[ACCESSIBILITY.md](ACCESSIBILITY.md)** - WCAG AA 기준 (300줄)
- **[REFERENCE.md](REFERENCE.md)** - API 레퍼런스 (200줄)

### 관련 Skill

- `yoda-content-generator` - 콘텐츠 생성 기반
- `moai-mermaid-diagram-expert` - 다이어그램 보충
- `yoda-korean-technical-book-writing` - 한국어 기술 책 작성

### 플랫폼 링크

- [DALL-E 3 API](https://platform.openai.com/)
- [Google Gemini API](https://ai.google.dev/)
- [Midjourney](https://www.midjourney.com/)
- [Stable Diffusion](https://stability.ai/)

---

## ✨ 최종 체크리스트

- [x] 8가지 스타일 완전 구현
- [x] 한국어 프롬프트 최적화
- [x] 플랫폼별 호환성 검증
- [x] 접근성 기준 (WCAG AA) 충족
- [x] 50+ 실전 예제 포함
- [x] yoda-master/yoda-book-author 통합 가능
- [x] Quick Start 30초 사용 가능
- [x] 자동 스타일 추천 로직
- [x] 배치 생성 기능
- [x] Claude Code 공식 표준 준수 (500줄 이하)

---

**Skill 생성일**: 2025-11-22  
**최종 업데이트**: 2025-11-22  
**버전**: 2.0.0 (Claude Code 표준 준수)  
**상태**: ✅ 프로덕션 준비 완료  
**품질**: TRUST 5 준수 - Test(covered), Readable(clear), Unified(consistent), Secured(safe), Trackable(versioned)

---

**이 Skill을 사용하여 모든 강의자료와 책 원고에 맞춤형 이미지 생성 프롬프트를 자동으로 생성하세요.**  
한국어 자연언어 프롬프트로 DALL-E 3, Gemini, Midjourney 등 모든 주요 이미지 생성 AI에서 활용 가능합니다.
