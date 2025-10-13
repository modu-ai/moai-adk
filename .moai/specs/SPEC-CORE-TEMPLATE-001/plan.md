# SPEC-CORE-TEMPLATE-001 구현 계획

## 우선순위별 마일스톤

### 1차 목표: TemplateProcessor 구현
- [ ] template/processor.py 생성
- [ ] Jinja2 Environment 설정
- [ ] render, render_to_file 메서드

### 2차 목표: ConfigManager 구현
- [ ] template/config.py 생성
- [ ] load, save, update 메서드
- [ ] deep_merge 유틸리티

### 3차 목표: 템플릿 파일 전환
- [ ] .ejs → .j2 변환
- [ ] 20개 언어 템플릿 검증
- [ ] 템플릿 변수 매핑

### 최종 목표: 통합 테스트
- [ ] 템플릿 렌더링 테스트
- [ ] config.json 읽기/쓰기 테스트
- [ ] 언어별 템플릿 조회 테스트

---

## 기술적 접근 방법

### Jinja2 vs EJS 비교

| 기능 | EJS (TS) | Jinja2 (Python) |
|------|----------|-----------------|
| 변수 | `<%= var %>` | `{{ var }}` |
| 조건 | `<% if %>` | `{% if %}` |
| 반복 | `<% for %>` | `{% for %}` |
| 주석 | `<%# comment %>` | `{# comment #}` |

### 템플릿 전환 전략
1. 기존 .ejs 파일 복사
2. 구문 변환 스크립트 실행
3. 수동 검증 및 수정

---

## 아키텍처 설계 방향

### 모듈 구조
```
core/template/
├── processor.py      # Jinja2 렌더링
├── config.py         # config.json 관리
└── variables.py      # 템플릿 변수 정의
```

### 의존성 최소화
- Jinja2만 사용 (외부 의존성 1개)
- 표준 라이브러리 json 사용

---

## 리스크 및 대응 방안

### 리스크 1: 템플릿 구문 차이
- **문제**: EJS → Jinja2 전환 시 누락 가능
- **대응**: 자동 변환 스크립트 + 수동 검증

### 리스크 2: 인코딩 문제
- **문제**: UTF-8 인코딩 미지원 환경
- **대응**: encoding="utf-8" 명시적 지정

### 리스크 3: 템플릿 캐싱
- **문제**: Jinja2는 기본 캐싱이 있음
- **대응**: 개발 시 auto_reload=True 설정
