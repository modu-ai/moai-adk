# {{TOPIC}} 핸즈온 워크숍

**강사**: {{INSTRUCTOR}}
**대상**: {{AUDIENCE}}
**난이도**: {{DIFFICULTY_LEVEL}}
**예상 시간**: {{TOTAL_DURATION}}

---

## 🎯 워크숍 목표

### 이 워크숍을 마치면

**할 수 있게 되는 것**:
- [ ] {{SKILL_GOAL_1}}
- [ ] {{SKILL_GOAL_2}}
- [ ] {{SKILL_GOAL_3}}
- [ ] {{SKILL_GOAL_4}}

**이해하게 되는 것**:
- [ ] {{UNDERSTANDING_GOAL_1}}
- [ ] {{UNDERSTANDING_GOAL_2}}
- [ ] {{UNDERSTANDING_GOAL_3}}

**배우게 되는 것**:
- [ ] {{LEARNING_GOAL_1}}
- [ ] {{LEARNING_GOAL_2}}

### 워크숍 구조

```
시간 배분:
├─ 소개 및 준비 (15분)
├─ 실습 1 (45분)
├─ 휴식 (15분)
├─ 실습 2 (45분)
├─ 디버깅 세션 (30분)
├─ 팀 프로젝트 소개 (15분)
└─ 마무리 및 피드백 (15분)

총 소요 시간: {{TOTAL_DURATION}}
```

---

## 📋 필수 준비물

### 운영 환경 요구사항

**OS**: {{SUPPORTED_OS}}
- Windows 10/11 (또는 WSL2)
- macOS 11+
- Linux (Ubuntu 20.04+)

**하드웨어 사양**:
- CPU: {{CPU_REQUIREMENT}} (멀티코어 권장)
- RAM: {{RAM_REQUIREMENT}}
- 디스크: {{DISK_REQUIREMENT}} (권장 {{RECOMMENDED_DISK}})
- 인터넷: 안정적인 연결 (최소 {{MIN_BANDWIDTH}})

### 필수 소프트웨어

| 소프트웨어 | 버전 | 설치 시간 | 필수 여부 |
|-----------|------|---------|---------|
| {{SOFTWARE_1}} | {{VERSION_1}} | ~{{TIME_1}} | ✅ 필수 |
| {{SOFTWARE_2}} | {{VERSION_2}} | ~{{TIME_2}} | ✅ 필수 |
| {{SOFTWARE_3}} | {{VERSION_3}} | ~{{TIME_3}} | ✅ 필수 |
| {{OPTIONAL_TOOL_1}} | {{OPT_VERSION_1}} | ~{{OPT_TIME_1}} | ⭕ 권장 |

### 권장 도구 및 리소스

**IDE/에디터**:
- {{IDE_1}} (권장)
- {{IDE_2}} (대안)

**추가 도구**:
- {{TOOL_1}}: {{TOOL_1_PURPOSE}}
- {{TOOL_2}}: {{TOOL_2_PURPOSE}}

**사전 지식**:
- {{PREREQUISITE_1}} (기초 이해)
- {{PREREQUISITE_2}} (기초 이해)
- {{PREREQUISITE_3}} (선택사항)

---

## ⚙️ 환경 설정 (30분)

### Step 1: {{SETUP_1_TITLE}}

**목표**: {{SETUP_1_OBJECTIVE}}

**설치 명령어**:
```bash
# {{SETUP_1_COMMAND_DESC}}
{{SETUP_1_COMMAND}}
```

**설정 파일**:
```yaml
# config.yaml
{{SETUP_1_CONFIG}}
```

**검증 방법**:
```bash
# 설치 확인 명령어
{{SETUP_1_VERIFY_COMMAND}}
```

**예상 출력**:
```
{{SETUP_1_EXPECTED_OUTPUT}}
```

**❌ 문제 해결**:
- 에러 메시지 1: {{SETUP_1_ERROR_1}}
  해결책: {{SETUP_1_SOLUTION_1}}

- 에러 메시지 2: {{SETUP_1_ERROR_2}}
  해결책: {{SETUP_1_SOLUTION_2}}

**소요 시간**: {{SETUP_1_TIME}}

✅ **Step 1 완료 확인**:
- [ ] {{SETUP_1_CHECK_1}}
- [ ] {{SETUP_1_CHECK_2}}

---

### Step 2: {{SETUP_2_TITLE}}

**목표**: {{SETUP_2_OBJECTIVE}}

**명령어**:
```bash
{{SETUP_2_COMMAND}}
```

**구성 옵션**:
- 옵션 A: {{OPTION_A_2}} (권장)
- 옵션 B: {{OPTION_B_2}} (대안)

**검증**:
```bash
{{SETUP_2_VERIFY}}
```

**예상 결과**:
```
{{SETUP_2_EXPECTED_RESULT}}
```

**소요 시간**: {{SETUP_2_TIME}}

✅ **Step 2 완료 확인**:
- [ ] {{SETUP_2_CHECK_1}}
- [ ] {{SETUP_2_CHECK_2}}

---

### Step 3: {{SETUP_3_TITLE}}

**목표**: {{SETUP_3_OBJECTIVE}}

**환경 변수 설정**:
```bash
# .env 파일
{{SETUP_3_ENV_VARS}}
```

**로드 확인**:
```bash
{{SETUP_3_VERIFY}}
```

**소요 시간**: {{SETUP_3_TIME}}

✅ **전체 환경 설정 완료**:
- [ ] 필수 소프트웨어 설치됨
- [ ] 모든 검증 통과
- [ ] 문서 읽음

---

## 🔧 실습 1: {{LAB_1_TITLE}}

**소요 시간**: {{LAB_1_TIME}} (약 {{LAB_1_MINUTES}}분)

**난이도**: {{LAB_1_DIFFICULTY}}

### 실습 목표

이 실습을 통해:
- [ ] {{LAB_1_GOAL_1}}
- [ ] {{LAB_1_GOAL_2}}
- [ ] {{LAB_1_GOAL_3}}

### 🎬 프로젝트 시작하기

**프로젝트 생성**:
```bash
# 프로젝트 디렉토리 생성
{{LAB_1_PROJECT_CREATE}}

# 프로젝트 초기화
cd {{LAB_1_PROJECT_DIR}}
{{LAB_1_INIT_COMMAND}}
```

**프로젝트 구조 확인**:
```
{{LAB_1_PROJECT_DIR}}/
├─ src/
│  └─ main.{{LAB_1_EXTENSION}}
├─ tests/
│  └─ test_main.{{LAB_1_EXTENSION}}
├─ README.md
├─ {{LAB_1_CONFIG_FILE}}
└─ requirements.txt
```

---

### 단계별 진행

#### Step 1: {{LAB_1_STEP_1_TITLE}}

**목표**: {{LAB_1_STEP_1_GOAL}}

**시간**: {{LAB_1_STEP_1_TIME}}

**설명**:

{{TOPIC}}의 첫 번째 단계는 {{LAB_1_STEP_1_DESCRIPTION}}입니다.

**핵심 개념**:
```
{{LAB_1_STEP_1_CONCEPT}}

Input → Process → Output
  ↓         ↓         ↓
Data   Logic    Result
```

**코드 작성**:
```{{LAB_1_LANGUAGE}}
// src/main.{{LAB_1_EXTENSION}}
// {{LAB_1_STEP_1_DESC}}

{{LAB_1_STEP_1_CODE}}
```

**파일 저장**: `src/main.{{LAB_1_EXTENSION}}`

**실행 방법**:
```bash
# 코드 실행
{{LAB_1_STEP_1_RUN_COMMAND}}
```

**예상 출력**:
```
{{LAB_1_STEP_1_EXPECTED_OUTPUT}}
```

**상세 설명**:
- 라인 1-5: {{LAB_1_STEP_1_EXPLAIN_1}}
- 라인 6-10: {{LAB_1_STEP_1_EXPLAIN_2}}
- 라인 11+: {{LAB_1_STEP_1_EXPLAIN_3}}

**검증**:
```bash
# 결과 검증
{{LAB_1_STEP_1_VERIFY}}
```

**검증 체크리스트**:
- ✅ {{LAB_1_STEP_1_CHECK_1}}
- ✅ {{LAB_1_STEP_1_CHECK_2}}
- ✅ {{LAB_1_STEP_1_CHECK_3}}

**일반적인 실수**:
- 실수 1: {{LAB_1_STEP_1_MISTAKE_1}}
  → 해결책: {{LAB_1_STEP_1_FIX_1}}

- 실수 2: {{LAB_1_STEP_1_MISTAKE_2}}
  → 해결책: {{LAB_1_STEP_1_FIX_2}}

---

#### Step 2: {{LAB_1_STEP_2_TITLE}}

**목표**: {{LAB_1_STEP_2_GOAL}}

**시간**: {{LAB_1_STEP_2_TIME}}

**설명**:

이제 {{LAB_1_STEP_2_DESCRIPTION}}를 추가합니다.

**코드 추가**:
```{{LAB_1_LANGUAGE}}
// Step 1의 코드에 다음을 추가
{{LAB_1_STEP_2_CODE}}
```

**실행**:
```bash
{{LAB_1_STEP_2_RUN_COMMAND}}
```

**예상 출력**:
```
{{LAB_1_STEP_2_EXPECTED_OUTPUT}}
```

**검증**:
```bash
{{LAB_1_STEP_2_VERIFY}}
```

---

#### Step 3: {{LAB_1_STEP_3_TITLE}}

**목표**: {{LAB_1_STEP_3_GOAL}}

**시간**: {{LAB_1_STEP_3_TIME}}

**설명**: {{LAB_1_STEP_3_DESCRIPTION}}

**구현**:
```{{LAB_1_LANGUAGE}}
{{LAB_1_STEP_3_CODE}}
```

**테스트**:
```bash
{{LAB_1_STEP_3_TEST_COMMAND}}
```

**예상 결과**:
```
{{LAB_1_STEP_3_EXPECTED_RESULT}}
```

---

### 🏆 실습 1 완료 확인

**완료 체크리스트**:
- [ ] 모든 단계 완료
- [ ] 코드 실행 성공
- [ ] 예상 출력과 일치
- [ ] 테스트 통과
- [ ] 강사 확인 (옵션)

**다음으로**: 아래의 실습 2로 진행하세요.

---

## 🔧 실습 2: {{LAB_2_TITLE}}

**소요 시간**: {{LAB_2_TIME}}

**난이도**: {{LAB_2_DIFFICULTY}} (더 도전적)

### 실습 목표

- [ ] {{LAB_2_GOAL_1}}
- [ ] {{LAB_2_GOAL_2}}
- [ ] {{LAB_2_GOAL_3}}

### 🎬 시작하기

**새 프로젝트**:
```bash
{{LAB_2_PROJECT_CREATE}}
```

**또는 실습 1의 프로젝트 확장**:
```bash
cd {{LAB_1_PROJECT_DIR}}
# 실습 2 분기 생성
git checkout -b feature/{{LAB_2_BRANCH}}
```

---

### 단계별 진행

#### Step 1: {{LAB_2_STEP_1_TITLE}}

**목표**: {{LAB_2_STEP_1_GOAL}}

**시간**: {{LAB_2_STEP_1_TIME}}

**설명**: {{LAB_2_STEP_1_DESCRIPTION}}

**구현**:
```{{LAB_2_LANGUAGE}}
{{LAB_2_STEP_1_CODE}}
```

**실행**:
```bash
{{LAB_2_STEP_1_RUN_COMMAND}}
```

**예상 출력**:
```
{{LAB_2_STEP_1_EXPECTED_OUTPUT}}
```

---

#### Step 2: {{LAB_2_STEP_2_TITLE}}

**목표**: {{LAB_2_STEP_2_GOAL}}

**설명**: {{LAB_2_STEP_2_DESCRIPTION}}

```{{LAB_2_LANGUAGE}}
{{LAB_2_STEP_2_CODE}}
```

**테스트**:
```bash
{{LAB_2_STEP_2_TEST}}
```

---

#### Step 3: {{LAB_2_STEP_3_TITLE}}

**목표**: {{LAB_2_STEP_3_GOAL}}

**심화 내용**: {{LAB_2_STEP_3_ADVANCED}}

```{{LAB_2_LANGUAGE}}
{{LAB_2_STEP_3_CODE}}
```

---

#### Step 4: {{LAB_2_STEP_4_TITLE}} (선택사항)

**목표**: {{LAB_2_STEP_4_GOAL}}

**도전 과제**: {{LAB_2_STEP_4_CHALLENGE}}

```{{LAB_2_LANGUAGE}}
{{LAB_2_STEP_4_CODE}}
```

---

### 🏆 실습 2 완료 확인

- [ ] 모든 단계 완료 (Step 4는 선택)
- [ ] 코드 테스트 통과
- [ ] 성능 최적화 적용
- [ ] 코드 리뷰 완료

---

## 🏆 팀 프로젝트

**프로젝트명**: {{PROJECT_NAME}}

**소요 시간**: {{PROJECT_TIME}}

**난이도**: {{PROJECT_DIFFICULTY}}

### 프로젝트 개요

**비즈니스 요구사항**:

{{PROJECT_REQUIREMENT_DESC}}

**기대 효과**:
- {{PROJECT_BENEFIT_1}}
- {{PROJECT_BENEFIT_2}}
- {{PROJECT_BENEFIT_3}}

### 팀 구성

**팀 크기**: {{TEAM_SIZE}}명 (권장: {{RECOMMENDED_TEAM_SIZE}}명)

**역할 분담**:
- 리더: {{ROLE_LEADER}} (프로젝트 관리)
- 개발자 1: {{ROLE_DEV1}} (기능 개발)
- 개발자 2: {{ROLE_DEV2}} (기능 개발)
- 테스터: {{ROLE_QA}} (테스트)

### 요구사항 (MVP)

**필수 기능** (100% 구현 필요):

1. {{PROJECT_REQUIREMENT_1}}
   ```
   입력: {{PROJECT_REQ1_INPUT}}
   출력: {{PROJECT_REQ1_OUTPUT}}
   ```

2. {{PROJECT_REQUIREMENT_2}}
   ```
   입력: {{PROJECT_REQ2_INPUT}}
   출력: {{PROJECT_REQ2_OUTPUT}}
   ```

3. {{PROJECT_REQUIREMENT_3}}
   ```
   입력: {{PROJECT_REQ3_INPUT}}
   출력: {{PROJECT_REQ3_OUTPUT}}
   ```

**추가 기능** (선택, 점수 추가):

- {{OPTIONAL_FEATURE_1}} (+{{OPTIONAL_POINTS_1}}점)
- {{OPTIONAL_FEATURE_2}} (+{{OPTIONAL_POINTS_2}}점)

### 제출 형식

**코드 저장소**:
```bash
# GitHub 저장소 생성
https://github.com/{{GITHUB_ORG}}/{{PROJECT_REPO}}

# 필수 파일 구조
project-repo/
├─ src/
│  └─ main.{{PROJECT_EXTENSION}}
├─ tests/
│  ├─ test_feature1.{{PROJECT_EXTENSION}}
│  └─ test_feature2.{{PROJECT_EXTENSION}}
├─ README.md
├─ SETUP.md
└─ DOCUMENTATION.md
```

**README.md 필수 항목**:
- 프로젝트 설명 (1-2문장)
- 기능 설명
- 설치 및 실행 방법
- 팀원 소개
- 특이사항

**SETUP.md 필수 항목**:
```markdown
# 환경 설정

## 요구사항
- {{PROJECT_REQUIREMENT_1}}
- {{PROJECT_REQUIREMENT_2}}

## 설치 방법
{{PROJECT_SETUP_STEPS}}

## 실행 방법
{{PROJECT_RUN_STEPS}}

## 문제 해결
{{PROJECT_TROUBLESHOOTING}}
```

**데모 비디오** (2-3분):
- 기능 데모
- 코드 리뷰
- 테스트 결과

### 평가 기준

| 항목 | 배점 | 평가 기준 |
|------|------|---------|
| **기능 완성도** | 40점 | 모든 필수 기능 구현 |
| **코드 품질** | 30점 | 가독성, 구조, 에러 처리 |
| **테스트** | 15점 | 테스트 커버리지, 테스트 통과 |
| **문서화** | 10점 | README, 코드 주석, 설정 문서 |
| **추가 기능** | +5~10점 | 선택 기능 구현 |

**합격 기준**: 70점 이상

### 프로젝트 일정

```
Day 1-2: 설계 및 계획
  ├─ 요구사항 분석
  ├─ 아키텍처 설계
  └─ 작업 분담

Day 3-4: 개발 (Phase 1)
  ├─ 필수 기능 구현
  └─ 초기 테스트

Day 5: 개발 (Phase 2) + 테스트
  ├─ 기능 완성
  ├─ 전체 테스트
  └─ 버그 수정

Day 6: 최종 점검 + 제출
  ├─ 코드 리뷰
  ├─ 문서 작성
  ├─ 데모 영상 촬영
  └─ 최종 제출
```

### 팀 협업 도구

**버전 관리**:
```bash
# Git 초기화
git init
git remote add origin {{PROJECT_GIT_URL}}

# 브랜치 전략 (Git Flow)
main (안정화된 코드)
  └─ develop (개발 브랜치)
     ├─ feature/{{FEATURE_1}} (개발자 1)
     ├─ feature/{{FEATURE_2}} (개발자 2)
     └─ feature/{{FEATURE_3}} (테스터)
```

**협업 가이드**:
- 매일 standup 미팅 (15분)
- 코드 리뷰 (PR 필수)
- 일일 진도 보고

### 성공 팁

✅ **해야 할 것**:
- [ ] 일찍 시작 (첫날부터 코딩)
- [ ] 작은 기능부터 구현 (점진적 개발)
- [ ] 자주 테스트 (단위 테스트 작성)
- [ ] 커뮤니케이션 (팀원과 정기적 소통)
- [ ] 문서화 (진행 중 작성)

❌ **피해야 할 것**:
- [ ] 마지막 날에 몰아서 하기
- [ ] 테스트 없이 코드 작성
- [ ] 요구사항 무시하고 마음대로 개발
- [ ] 커뮤니케이션 부족
- [ ] 코드 리뷰 생략

---

## 🐛 트러블슈팅

### 설치 및 환경 설정 문제

#### 문제 1: {{TROUBLESHOOT_1_TITLE}}

**증상**:
```
{{TROUBLESHOOT_1_ERROR_MESSAGE}}
```

**원인**:
- 가능성 1: {{TROUBLESHOOT_1_CAUSE_1}}
- 가능성 2: {{TROUBLESHOOT_1_CAUSE_2}}

**해결책**:
```bash
# 방법 1: {{TROUBLESHOOT_1_SOLUTION_1_NAME}}
{{TROUBLESHOOT_1_SOLUTION_1_COMMAND}}

# 방법 2: {{TROUBLESHOOT_1_SOLUTION_2_NAME}}
{{TROUBLESHOOT_1_SOLUTION_2_COMMAND}}
```

**검증**:
```bash
{{TROUBLESHOOT_1_VERIFY}}
```

**성공 시 출력**:
```
{{TROUBLESHOOT_1_SUCCESS_OUTPUT}}
```

**더 알아보기**: {{TROUBLESHOOT_1_REFERENCE}}

---

#### 문제 2: {{TROUBLESHOOT_2_TITLE}}

**증상**:
```
{{TROUBLESHOOT_2_ERROR}}
```

**해결책**:
```bash
{{TROUBLESHOOT_2_SOLUTION}}
```

**검증**:
```bash
{{TROUBLESHOOT_2_VERIFY}}
```

---

### 실습 중 일반적인 문제

#### 문제 3: {{TROUBLESHOOT_3_TITLE}}

**상황**: {{TROUBLESHOOT_3_CONTEXT}}

**에러 메시지**:
```
{{TROUBLESHOOT_3_ERROR}}
```

**해결책**:

1단계: {{TROUBLESHOOT_3_STEP1}}
```bash
{{TROUBLESHOOT_3_STEP1_COMMAND}}
```

2단계: {{TROUBLESHOOT_3_STEP2}}
```bash
{{TROUBLESHOOT_3_STEP2_COMMAND}}
```

**성공 확인**:
```bash
{{TROUBLESHOOT_3_VERIFY}}
```

---

#### 문제 4: {{TROUBLESHOOT_4_TITLE}}

**상황**: {{TROUBLESHOOT_4_CONTEXT}}

**증상**: {{TROUBLESHOOT_4_SYMPTOM}}

**빠른 해결책**:
```bash
{{TROUBLESHOOT_4_QUICK_FIX}}
```

**상세 해결책**: {{TROUBLESHOOT_4_DETAILED_FIX}}

---

### 성능 및 최적화 문제

#### 문제 5: {{TROUBLESHOOT_5_TITLE}}

**문제 상황**: {{TROUBLESHOOT_5_SITUATION}}

**성능 진단**:
```bash
{{TROUBLESHOOT_5_DIAGNOSIS_COMMAND}}
```

**최적화 방법**:
- 방법 1: {{TROUBLESHOOT_5_OPTIMIZE_1}}
- 방법 2: {{TROUBLESHOOT_5_OPTIMIZE_2}}

---

### 온라인 리소스

**더 많은 도움**:
- Context7 MCP 문서: `{{LIBRARY_NAME}}`
- 공식 포럼: {{FORUM_URL}}
- GitHub Issues: {{GITHUB_ISSUES_URL}}
- Stack Overflow: {{STACKOVERFLOW_TAG}}
- 강사 이메일: {{INSTRUCTOR_EMAIL}}

---

## 📚 추가 학습 자료

### 심화 학습

**고급 주제**:
- {{ADVANCED_TOPIC_1}}: {{ADVANCED_TOPIC_1_URL}}
- {{ADVANCED_TOPIC_2}}: {{ADVANCED_TOPIC_2_URL}}
- {{ADVANCED_TOPIC_3}}: {{ADVANCED_TOPIC_3_URL}}

**추천 튜토리얼**:
- {{TUTORIAL_1_TITLE}}: {{TUTORIAL_1_URL}}
- {{TUTORIAL_2_TITLE}}: {{TUTORIAL_2_URL}}

**공식 문서**:
- Context7 MCP: `{{LIBRARY_NAME}}`
  - API 레퍼런스
  - 마이그레이션 가이드
  - Best Practices

### 관련 프로젝트

**샘플 프로젝트**:
- {{SAMPLE_PROJECT_1}}: {{SAMPLE_PROJECT_1_URL}}
- {{SAMPLE_PROJECT_2}}: {{SAMPLE_PROJECT_2_URL}}

**오픈소스**:
- {{OPENSOURCE_PROJECT_1}}: {{OPENSOURCE_PROJECT_1_URL}}

---

## ✅ 워크숍 완료 체크리스트

### 실습 완료 현황

**환경 설정**:
- [ ] 모든 필수 소프트웨어 설치
- [ ] 환경 변수 설정 완료
- [ ] 설정 검증 통과

**실습 1**:
- [ ] 모든 단계 완료
- [ ] 코드 실행 성공
- [ ] 테스트 통과

**실습 2**:
- [ ] 모든 필수 단계 완료
- [ ] 선택 단계 도전 (선택)
- [ ] 코드 최적화 적용

**팀 프로젝트**:
- [ ] 요구사항 분석 완료
- [ ] 설계 및 계획 완료
- [ ] MVP 기능 구현 완료
- [ ] 테스트 완료
- [ ] 문서화 완료
- [ ] 최종 제출 완료

### 학습 목표 달성도

**검증 (자체 평가)**:
- [ ] {{VERIFY_GOAL_1}}
- [ ] {{VERIFY_GOAL_2}}
- [ ] {{VERIFY_GOAL_3}}
- [ ] {{VERIFY_GOAL_4}}

### 피드백 및 인증

**피드백 양식**: {{FEEDBACK_FORM_URL}}

**인증서** (완료 후):
- 발급: {{CERTIFICATE_PROVIDER}}
- 링크: {{CERTIFICATE_URL}}
- 공유하기: LinkedIn, GitHub 등

---

## 📝 워크숍 평가 (선택)

### 자체 평가

**워크숍 만족도**:
- 내용 이해도: ⭐ ⭐ ⭐ ⭐ ⭐ (1-5)
- 난이도 적절성: ⭐ ⭐ ⭐ ⭐ ⭐ (1-5)
- 실습 유용성: ⭐ ⭐ ⭐ ⭐ ⭐ (1-5)

**피드백**:
- 가장 유익했던 부분: {{FEEDBACK_USEFUL}}
- 개선할 점: {{FEEDBACK_IMPROVEMENT}}
- 추가 학습 희망 주제: {{FEEDBACK_ADDITIONAL}}

### 강사 평가 (필수)

**완료 인증**:
- 강사명: {{INSTRUCTOR_OFFICIAL}}
- 서명: __________________
- 날짜: {{CERTIFICATION_DATE}}

---

## 🚀 다음 단계

### 즉시 실행 (이번 주)

- [ ] {{IMMEDIATE_ACTION_1}}
- [ ] {{IMMEDIATE_ACTION_2}}
- [ ] {{IMMEDIATE_ACTION_3}}

### 단기 계획 (1개월)

- {{SHORTTERM_ACTION_1}}
- {{SHORTTERM_ACTION_2}}

### 장기 계획 (3-6개월)

- {{LONGTERM_GOAL_1}}
- {{LONGTERM_GOAL_2}}

---

## 📞 지원 및 연락처

**강사 연락처**:
- 이메일: {{INSTRUCTOR_EMAIL}}
- 사무실: {{INSTRUCTOR_OFFICE}}
- 온라인: {{INSTRUCTOR_ONLINE}}

**팀 연락처**:
- 운영팀: {{OPERATIONS_EMAIL}}
- Slack 채널: #{{SLACK_CHANNEL}}
- 포럼: {{FORUM_URL}}

**추가 자료**:
- 워크숍 서버: {{WORKSHOP_SERVER}}
- 클라우드 환경 (선택): {{CLOUD_ENV_URL}}
- 협업 도구: {{COLLABORATION_TOOL}}

---

**마스터 요다의 워크숍 원칙: 배우는 것이 아니라 경험하라**

**템플릿 정보**:
- 버전: 1.0.0
- 예상 시간: {{TOTAL_DURATION}}
- 마지막 업데이트: 2025-11-14
- 대상: 개발 학습자
- 권장 방식: 라이브 인스트럭터 주도 + 자율 학습

---

*이 워크숍이 {{TOPIC}}의 실무 능력 향상에 도움이 되길 바랍니다.*
