---

name: moai-spec-intelligent-workflow
description: Intelligent SPEC generation decision engine with 3-level templates and analytics for MoAI-ADK

---

# SPEC Intelligent Workflow Skill

## 🎯 30초 요약

**SPEC Intelligent Workflow**는 Alfred의 자동 SPEC 판단 시스템입니다.

사용자 요청을 자연어로 분석하여:

- ✅ **0-1개 조건** → SPEC 불필요 (즉시 구현)
- ✅ **2-3개 조건** → SPEC 권장 (사용자 선택)
- ✅ **4-5개 조건** → SPEC 강력 권장 (강조)

자동으로 **3단계 템플릿** 중 하나를 선택하고, **통계 시스템**으로 효과를 추적합니다.


## 📚 주요 기능

### 1. Alfred의 5가지 자동 판단 기준

```
① 파일 수정 범위 (1파일 vs 여러 파일)
② 아키텍처 영향 (있음/없음)
③ 컴포넌트 통합 (단일 vs 복합)
④ 구현 시간 (30분 기준)
⑤ 향후 유지보수 (필요/불필요)
```

### 2. 3단계 SPEC 템플릿

```
Level 1 (Minimal)     → 간단한 작업, 5-10분 작성
Level 2 (Standard)    → 일반 기능, 10-15분 작성
Level 3 (Comprehensive) → 복잡한 작업, 20-30분 작성
```

### 3. 통계 및 분석

```
세션 시작 시: 최근 30일 SPEC 통계 자동 표시
세션 종료 시: SPEC 관련 데이터 자동 수집
월간 리포트: 효과 분석 및 개선 권장사항
```


## 🚀 빠른 시작

### 사용자 요청 → 자동 SPEC 판단

```
사용자: "사용자 프로필 이미지 업로드 기능을 추가해주세요"
  ↓
Alfred 분석: 4개 조건 충족 → SPEC 강력 권장
  ↓
사용자 선택: "예, SPEC 생성"
  ↓
자동 /moai:1-plan 실행
  ↓
Level 2 템플릿 자동 선택
  ↓
SPEC-XXX 생성 완료
  ↓
/moai:2-run SPEC-XXX 구현
```


## 📁 Skill 구조

| 파일                         | 목적                      | 크기 |
| ---------------------------- | ------------------------- | ---- |
| **README.md**                | 개요 및 빠른 시작         | 5KB  |
| **alfred-decision-logic.md** | Alfred 판단 알고리즘 상세 | 12KB |
| **templates.md**             | 3단계 SPEC 템플릿 및 예제 | 15KB |
| **analytics.md**             | 통계 및 분석 시스템 설계  | 10KB |
| **examples.md**              | 10+ 실전 사용 예제        | 12KB |
| **FAQ.md**                   | 자주 묻는 질문            | 5KB  |


## 💡 핵심 특징

- ✅ **자연어 판단**: 복잡한 점수 계산 없음
- ✅ **자동 선택**: 3단계 템플릿 자동 선택
- ✅ **사용자 존중**: 모든 제안 거부 가능
- ✅ **데이터 기반**: 통계로 효과 측정
- ✅ **TAG 교훈**: 과도한 자동화 방지


## 🔗 CLAUDE.md와의 관계

```
CLAUDE.md (개요)
  ↓
이 Skill (상세 구현)
  ├── Alfred 판단 알고리즘
  ├── 3단계 템플릿 완전 정의
  ├── 통계 시스템 설계
  └── 10+ 실전 예제
```

CLAUDE.md는 이 Skill의 개요만 포함하며, 상세 내용은 여기서 확인합니다.


## 📖 문서 참고

### 빠르게 이해하기

→ **README.md** (5분)

### Alfred의 판단 기준 학습

→ **alfred-decision-logic.md** (10분)

### SPEC 템플릿 선택

→ **templates.md** (15분)

### 통계 시스템 이해

→ **analytics.md** (10분)

### 실제 사용 사례

→ **examples.md** (20분)

### 궁금한 점 해결

→ **FAQ.md** (5분)


## 🎯 사용 시점

이 Skill은 다음 상황에서 참고됩니다:

1. **사용자가 새로운 작업 요청할 때**

   - Alfred가 SPEC 필요성을 판단
   - 5가지 조건 이해하기 위해 참고

2. **SPEC을 생성하기로 결정했을 때**

   - 3단계 템플릿 선택 기준 확인
   - 자신의 작업에 맞는 템플릿 찾기

3. **SPEC-First 워크플로우 효과를 알고 싶을 때**

   - 통계 시스템 이해
   - 월간 리포트 분석

4. **의문점이 생겼을 때**
   - FAQ 확인
   - 실전 예제로 이해


## ✨ 기대 효과

| 지표            | 예상 개선  |
| --------------- | ---------- |
| SPEC 사용률     | 50% → 85%  |
| 구현 시간       | 30% 단축   |
| 코드 품질       | 15% 향상   |
| 테스트 커버리지 | 80% → 90%+ |
| 버그 감소       | 25% 감소   |


**버전**: 1.0.0
**마지막 업데이트**: 2025-11-21
**상태**: Active - Production Ready
