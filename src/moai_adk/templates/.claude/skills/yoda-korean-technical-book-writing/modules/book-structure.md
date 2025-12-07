# Book Structure Best Practices

## 목차(Table of Contents) 설계

**MECE 원칙 적용** (Mutually Exclusive, Collectively Exhaustive):
- **상호 배타적**: 각 장/절의 내용이 중복되지 않음
- **전체 포괄적**: 주제의 모든 핵심 영역을 다룸

**목차 레벨 구조**:
```
Part (선택적)        # 대주제 구분 (10+ 장인 경우)
└─ Chapter (장)      # 4-8개 권장
   └─ Section (절)   # 장당 5-9개 권장
      └─ Subsection (항) # 필요시만 사용, 최대 3단계
```

**균형 잡힌 목차 예시**:
```
Part 1. 기초 다지기
├─ Chapter 1. Python 시작하기 (6절)
├─ Chapter 2. 변수와 자료형 (7절)
└─ Chapter 3. 제어문 (8절)

Part 2. 심화 학습
├─ Chapter 4. 함수와 모듈 (7절)
├─ Chapter 5. 객체지향 프로그래밍 (9절)
└─ Chapter 6. 예외 처리 (6절)
```

## Chapter 구조 템플릿

**표준 장 구성** (Pedagogical Framework 기반):

```markdown
# Chapter N. [장 제목]

## 학습 목표 (Opener - Learning Objectives)
- 이 장을 마치면 ~를 할 수 있습니다
- ~의 개념을 이해하고 설명할 수 있습니다
- ~와 ~의 차이점을 구별할 수 있습니다

## 도입 (Opener - Introduction)
[장의 핵심 질문 제시]
[실생활 문제 연결]
[이전 장과의 연계성]

## 핵심 개념 1 (Main Content)
### 개념 설명
### 코드 예제
### 심화 토론

## 핵심 개념 2
[반복]

## 실전 프로젝트 (Integrated Device - Case Study)
[종합 응용 예제]

## 요약 (Closer - Summary)
[핵심 내용 3-5개 요점 정리]

## 연습 문제 (Closer - Review Questions)
[기초 3문제 + 응용 2문제]

## 더 읽을거리 (Closer - External Resources)
[참고 문헌, 공식 문서 링크]
```

## Section 구조 패턴

**3단계 학습 패턴** (개념 → 예제 → 연습):

```markdown
## N.M 섹션 제목

### Why: 왜 필요한가?
[문제 상황 제시]
[기존 방법의 한계]

### What: 무엇인가?
[개념 정의]
[핵심 원리 설명]

### How: 어떻게 사용하는가?
[단계별 절차]
[코드 예제]

### Try It: 직접 해보기
[미니 실습 과제]
```

## 부록 및 참고 자료 구성

**부록 포함 항목**:
- **부록 A**: 개발 환경 설정 (OS별 상세 가이드)
- **부록 B**: 문제 해결 FAQ (독자 피드백 기반)
- **부록 C**: 참고 자료 (공식 문서, 커뮤니티 링크)
- **부록 D**: 용어 사전 (Glossary)
- **부록 E**: 정규 표현식/단축키 등 레퍼런스

**색인(Index) 작성 가이드**:
```
✓ 주요 개념어 모두 포함
✓ 영문 용어는 한글 표기와 교차 참조
  예: "컨테이너 → Container 참조"
✓ 약어는 풀네임과 연결
  예: "API → Application Programming Interface"
✓ 페이지 번호는 주 설명 페이지를 볼드 처리
```
