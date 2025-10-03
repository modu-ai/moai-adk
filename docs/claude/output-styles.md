# Output Styles

Alfred는 다양한 개발 상황에 맞는 4가지 대화 스타일을 제공합니다.

## 스타일 전환

Claude Code에서 언제든지 스타일을 전환할 수 있습니다:

```text
/output-style alfred-pro           # 전문 개발 모드 (기본)
/output-style beginner-learning  # 초보자 학습 모드
/output-style pair-collab        # 협업 모드
/output-style study-deep         # 심화 학습 모드
```

## 1️⃣ MoAI Professional (기본)

**대상**: 전문 개발자, 프로덕션 개발

### 특징
- 간결하고 기술적인 설명
- SPEC-First TDD 자동 적용
- TRUST 5원칙 및 @TAG 추적성 자동 검증
- 다중 언어 TDD 예제 (TypeScript, Python, Go, Rust)

### 사용 시점
- 실무 프로젝트 개발
- 빠른 기능 구현
- 프로덕션 코드 작성

### 예시

Alfred가 간결하고 기술적으로 응답합니다:

```text
사용자: "JWT 인증을 구현하고 싶어요"

Alfred:
✅ SPEC-AUTH-001 생성
- EARS 구문: Ubiquitous 요구사항 3개
- 브랜치: feature/SPEC-AUTH-001-jwt-auth
- 예상 구현: 2-3시간

다음: /alfred:2-build SPEC-AUTH-001
```

---

## 2️⃣ Beginner Learning

**대상**: 프로그래밍 초보자, 개발 입문자

### 특징
- 친절하고 상세한 단계별 가이드
- 기술 용어를 쉬운 말로 설명
- 실수 방지 및 격려 메시지
- Python, Flutter 기초 예제

### 사용 시점
- 프로그래밍 학습
- MoAI-ADK 입문
- 개발 기초 이해

### 예시

Alfred가 친절하게 설명합니다:

```text
사용자: "JWT가 뭔가요?"

Alfred:
🎯 JWT란? (쉽게 이해하기)

쉬운 설명: JWT는 인터넷 신분증 같은 거예요!
- 사용자가 로그인하면 서버가 "신분증"을 발급
- 다음부터는 이 신분증만 보여주면 OK
- 비밀번호 매번 입력 안 해도 됨

비유: 놀이공원 입장권처럼, 한 번 받으면 계속 사용 가능!

다음 단계: JWT를 직접 만들어볼까요? (y/n)
```

---

## 3️⃣ Pair Collaboration

**대상**: 아키텍처 설계, 코드 리뷰, 브레인스토밍

### 특징
- 실시간 협업 및 토론 중심
- 트레이드오프 분석 및 기술 선택 지원
- 다중 언어 코드 리뷰 (TypeScript, Python, Go)
- 아이디어 발굴 및 설계 세션

### 사용 시점
- 시스템 설계
- 기술 선택
- 코드 리뷰

### 예시

Alfred가 함께 고민합니다:

```text
사용자: "데이터베이스를 PostgreSQL로 할지 MongoDB로 할지 고민이에요"

Alfred:
💭 함께 분석해봅시다

PostgreSQL (관계형):
장점:
+ ACID 보장, 복잡한 쿼리 강력
+ 데이터 무결성 우수

단점:
- 스키마 변경 어려움
- 수평 확장 제한적

MongoDB (문서형):
장점:
+ 유연한 스키마
+ 수평 확장 용이

단점:
- JOIN 제한적
- ACID 보장 약함

💡 제 생각:
초기 MVP → MongoDB (빠른 반복)
프로덕션 → PostgreSQL (안정성)

현재 단계는? 우선순위는?
```

---

## 4️⃣ Study Deep

**대상**: 신기술 학습자, 프레임워크 전환자

### 특징
- 체계적인 학습 경로 제공
- 프레임워크별 심화 가이드
  - Express.js (TypeScript)
  - FastAPI (Python)
  - Gin (Go)
  - Axum (Rust)
- 개념 → 실습 → 전문가 팁 3단계 학습

### 사용 시점
- 새로운 언어/프레임워크 학습
- 기술 심화
- 실무 패턴 학습

### 예시

Alfred가 체계적으로 가르칩니다:

```text
사용자: "FastAPI를 배우고 싶어요"

Alfred:
📚 Learning Journey: FastAPI (Python)

🎯 Why This Matters:
FastAPI는 현대 Python 웹 프레임워크의 표준으로,
자동 문서화와 타입 검증을 제공합니다.

🏗️ Conceptual Foundation:
Pydantic 모델 → 자동 검증
async/await → 고성능 비동기

🔍 Basic Concept:
[간단한 예제 코드]

⚡ TDD with pytest:
[테스트 예제]

💡 Pro Tips:
✨ Dependency Injection: Depends()로 깔끔한 의존성 관리
⚠️ Pitfall: async와 sync 함수 혼용 주의

다음: 실습 프로젝트 시작할까요?
```

---

## Alfred SuperAgent와 Output Styles

모든 Output Styles에서 Alfred는 동일하게:
- ✅ 9개 전문 에이전트 조율
- ✅ SPEC-First TDD 워크플로우 적용
- ✅ TRUST 5원칙 자동 검증
- ✅ @TAG 추적성 보장

**차이점은 설명 방식과 학습 깊이**입니다.

---

## 스타일 비교표

| 스타일 | 대상 | 설명 방식 | 코드 예제 | 학습 깊이 |
|--------|------|-----------|-----------|-----------|
| **alfred-pro** | 전문 개발자 | 간결·기술적 | 다중 언어 TDD | 실무 중심 |
| **beginner-learning** | 초보자 | 친절·상세 | Python, Flutter | 기초 중심 |
| **pair-collab** | 설계·리뷰 | 협업·토론 | 다중 언어 리뷰 | 설계 중심 |
| **study-deep** | 학습자 | 체계적 | 프레임워크별 | 심화 학습 |

---

## 관련 문서

- [Alfred 에이전트](/claude/agents)
- [3단계 워크플로우](/guide/workflow)
- [SPEC-First TDD](/guide/spec-first-tdd)
