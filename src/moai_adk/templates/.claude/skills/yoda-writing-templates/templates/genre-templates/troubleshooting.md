# 장르 템플릿: troubleshooting (문제 해결형)

## 📋 용도

에러, 버그, 성능 문제 해결 가이드

**구조**: Problem Description (15%) → Root Cause (20%) → Solution Steps (40%) → Prevention (15%) → Related Issues (10%)

**특징**:
- 문제 중심 접근
- 단계별 해결 방법
- 재발 방지책 제공
- 유사 문제 연결

---

## 🏗️ 5가지 구성 요소

### 1. 문서 구조

```
├─ Problem Description (15%, 증상 설명)
│  ├─ 에러 메시지
│  └─ 발생 상황
│
├─ Root Cause (20%, 원인 분석)
│  ├─ 근본 원인
│  └─ 발생 조건
│
├─ Solution Steps (40%, 해결 방법)
│  ├─ Step 1: 진단
│  ├─ Step 2: 수정
│  ├─ Step 3: 검증
│  └─ 각 단계 코드 예제
│
├─ Prevention (15%, 재발 방지)
│  ├─ Best practices
│  └─ 체크리스트
│
└─ Related Issues (10%, 관련 문제)
   └─ 유사 문제 링크

총 섹션: 5개
비율: 15+20+40+15+10 = 100%
```

### 2. 문체

**어조**: 문제 해결 중심

**종결어미**: "-다", "-요" (지시형)

**능동태**: 95% 이상

**문장 길이**: 평균 15-18단어

### 3. 내용 전개

1. **Problem**: 에러 메시지, 스크린샷
2. **Root Cause**: 왜 발생하는가
3. **Solution**: Step 1-3 상세 해결
4. **Prevention**: 재발 방지
5. **Related**: 유사 문제

### 4. 조건

**글자 수**: 1,500-2,000자

**코드 예제**: 3개 (Before/After/Verification)

**필수 요소**:
- 에러 메시지 원문
- 단계별 해결 방법
- 검증 방법

### 5. 형식

```markdown
# ModuleNotFoundError 해결하기

## Problem Description

다음과 같은 에러가 발생합니다:

\`\`\`
ModuleNotFoundError: No module named 'anthropic'
\`\`\`

## Root Cause

Python 가상 환경에 `anthropic` 패키지가 설치되지 않았기 때문입니다.

## Solution Steps

### Step 1: 가상 환경 확인

\`\`\`bash
which python
# 출력: /path/to/venv/bin/python
\`\`\`

### Step 2: 패키지 설치

\`\`\`bash
pip install anthropic
\`\`\`

### Step 3: 설치 검증

\`\`\`python
import anthropic
print(anthropic.__version__)
\`\`\`

## Prevention

- `requirements.txt`에 모든 의존성 명시
- CI/CD에서 패키지 설치 자동화

## Related Issues

- `ImportError: cannot import name 'Client'`
- 버전 호환성 문제
```

---

**마지막 수정**: 2025-11-25
**버전**: 1.0.0
