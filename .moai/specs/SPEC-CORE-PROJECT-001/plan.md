# SPEC-CORE-PROJECT-001 구현 계획

## 우선순위별 마일스톤

### 1차 목표: LanguageDetector 구현
- [ ] detector.py 생성
- [ ] 20개 언어 패턴 정의
- [ ] detect, detect_multiple 메서드

### 2차 목표: ProjectInitializer 구현
- [ ] initializer.py 생성
- [ ] _create_directories 메서드
- [ ] _generate_language_templates 메서드

### 3차 목표: SystemChecker 구현
- [ ] checker.py 생성
- [ ] check_all 메서드
- [ ] 필수/선택 도구 분류

### 최종 목표: 통합 테스트
- [ ] 각 언어 감지 테스트
- [ ] 초기화 프로세스 테스트
- [ ] 시스템 체크 테스트

---

## 기술적 접근 방법

### 언어 감지 전략
1. **파일 확장자**: *.py, *.ts, *.java 등
2. **설정 파일**: package.json, pom.xml, go.mod 등
3. **우선순위**: 설정 파일 > 파일 확장자

### 디렉토리 구조 생성 전략
- pathlib.Path 사용 (OS 독립적)
- parents=True (중간 디렉토리 자동 생성)
- exist_ok=True (이미 존재 시 무시)

---

## 아키텍처 설계 방향

### 모듈 구조
```
core/project/
├── initializer.py    # 프로젝트 초기화
├── detector.py       # 언어 감지
└── checker.py        # 시스템 체크
```

### 의존성
- initializer → detector (언어 감지)
- initializer → template (템플릿 렌더링)

---

## 리스크 및 대응 방안

### 리스크 1: 언어 감지 정확도
- **문제**: 멀티 언어 프로젝트에서 주 언어 판별 어려움
- **대응**: detect_multiple로 모든 언어 반환, 사용자 선택

### 리스크 2: 시스템 도구 미설치
- **문제**: git, gh 등이 없을 수 있음
- **대응**: 필수/선택 도구 구분, 설치 가이드 제공

### 리스크 3: 디렉토리 권한 문제
- **문제**: .moai/ 생성 권한 없음
- **대응**: PermissionError 처리, 명확한 에러 메시지
