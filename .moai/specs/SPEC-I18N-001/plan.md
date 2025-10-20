# SPEC-I18N-001 구현 계획서

> **SPEC**: 다국어 템플릿 시스템 (한/영)
> **버전**: v0.0.1
> **상태**: draft

---

## 1. 구현 우선순위

### 1차 목표: 템플릿 분리 및 기반 구축
- 기존 `.claude/` → `.claude-ko/` 이동
- `.claude-en/` 생성 및 영어 번역
- 템플릿 구조 동일성 검증

### 2차 목표: CLI 및 Processor 수정
- init.py에 locale 선택 프롬프트 추가
- processor.py에 locale 기반 템플릿 복사 로직 추가
- config.json 스키마 확장 (locale 필드)

### 3차 목표: 문서화
- README.md 한국어로 재작성
- README.en.md 영어 번역 작성
- 링크 및 참조 업데이트

---

## 2. 기술적 접근 방법

### 2.1 템플릿 디렉토리 구조 설계

**원칙**:
- 디렉토리 이름: `.claude-{locale}/` (locale은 "ko" 또는 "en")
- 파일 구조: 양쪽 동일 (명명 규칙, 디렉토리 계층 동일)
- 내용만 번역 (파일명은 변경하지 않음)

**검증 방법**:
```bash
# 디렉토리 구조 비교
diff <(cd .claude-ko && find . -type f | sort) \
     <(cd .claude-en && find . -type f | sort)

# 파일 개수 확인
ls -1 .claude-ko/ | wc -l  # 49개
ls -1 .claude-en/ | wc -l  # 49개 (동일)
```

### 2.2 init.py 수정 전략

**변경 사항**:
1. locale 선택 프롬프트 추가
   - 선택지: "Korean (한국어)" / "English"
   - 기본값: "Korean (한국어)"
2. 선택된 locale 값을 `config.json`에 저장
3. processor에 locale 값 전달

**API 변경**:
```python
# Before
def initialize_project(name: str, mode: str) -> None:
    ...

# After
def initialize_project(name: str, mode: str, locale: str = "ko") -> None:
    ...
    config["project"]["locale"] = locale
    processor.copy_claude_template(locale)
```

### 2.3 processor.py 수정 전략

**변경 사항**:
1. `copy_claude_template(locale)` 메서드에 locale 매개변수 추가
2. locale 기반 템플릿 디렉토리 선택 로직
3. Fallback 처리 (지원되지 않는 locale → "en")

**에러 처리**:
- 템플릿 디렉토리 누락 시: `FileNotFoundError` + 명확한 메시지
- 복사 실패 시: `shutil.Error` + 복구 가이드

---

## 3. 마일스톤 (시간 예측 제외)

### Milestone 1: 템플릿 분리
**작업 항목**:
- [ ] `.claude/` → `.claude-ko/` 이동 (Git mv)
- [ ] `.claude-en/` 디렉토리 생성
- [ ] 주요 파일 번역 (commands/, agents/, README.md)
- [ ] 구조 동일성 검증 (스크립트)

**완료 조건**:
- 디렉토리 구조 100% 일치
- 필수 파일 존재 확인
- Git 이력 보존

### Milestone 2: CLI 및 Processor 수정
**작업 항목**:
- [ ] init.py에 locale 선택 프롬프트 추가
- [ ] processor.py에 locale 기반 복사 로직 추가
- [ ] config.json 스키마 확장 (locale 필드)
- [ ] 단위 테스트 작성 (test_i18n_template.py)

**완료 조건**:
- 한국어 선택 시 `.claude-ko/` 복사 확인
- 영어 선택 시 `.claude-en/` 복사 확인
- 테스트 커버리지 ≥85%

### Milestone 3: 문서화 및 통합 테스트
**작업 항목**:
- [ ] README.md 한국어 재작성
- [ ] README.en.md 영어 번역
- [ ] 통합 테스트 작성 (test_init_i18n.py)
- [ ] 링크 유효성 검증

**완료 조건**:
- README 2개 버전 완성
- 통합 테스트 통과
- 링크 100% 유효

---

## 4. 아키텍처 설계 방향

### 4.1 템플릿 선택 로직

```
moai-adk init
    ↓
사용자 입력: locale ("ko" | "en")
    ↓
TemplateProcessor.copy_claude_template(locale)
    ↓
    ├─ locale == "ko" → .claude-ko/ 복사
    ├─ locale == "en" → .claude-en/ 복사
    └─ 그 외 → "en" (Fallback) + 경고
    ↓
.moai/config.json에 locale 저장
```

### 4.2 에러 처리 플로우

```
TemplateProcessor.copy_claude_template(locale)
    ↓
    ├─ 템플릿 디렉토리 없음?
    │   ↓
    │   FileNotFoundError: "Template directory not found: .claude-{locale}/"
    │
    ├─ 복사 권한 없음?
    │   ↓
    │   PermissionError: "Permission denied: {dest}"
    │
    └─ 복사 성공?
        ↓
        logger.info(f"✅ Template copied: .claude-{locale}/ → .claude/")
```

---

## 5. 리스크 및 대응 방안

### 리스크 1: 템플릿 구조 불일치
**시나리오**: `.claude-ko/`와 `.claude-en/`의 파일 구조가 다름
**영향**: 일부 기능 누락 또는 에러 발생
**대응**:
- 자동화 스크립트로 구조 검증
- CI/CD에서 차이 감지 시 빌드 실패

### 리스크 2: 번역 품질 부족
**시나리오**: 영어 번역이 부정확하거나 누락
**영향**: 영어 사용자 UX 저하
**대응**:
- 네이티브 리뷰 요청
- 커뮤니티 피드백 반영

### 리스크 3: 기존 사용자 호환성
**시나리오**: 기존 프로젝트에서 `.claude/` 디렉토리 참조 시 깨짐
**영향**: 기존 사용자 워크플로우 중단
**대응**:
- 마이그레이션 가이드 제공
- `moai-adk update` 명령어로 자동 마이그레이션 지원

---

## 6. 성능 최적화 전략

### 최적화 포인트 1: 템플릿 복사 속도
**목표**: 템플릿 복사 ≤1초
**방법**:
- `shutil.copytree(..., dirs_exist_ok=True)` 사용
- 불필요한 파일 제외 (`.DS_Store`, `__pycache__`)

### 최적화 포인트 2: 메모리 사용량
**목표**: 메모리 증가 ≤10MB
**방법**:
- 파일 단위 복사 (전체 로드하지 않음)
- 대용량 파일 스트리밍 복사

---

## 7. 테스트 계획

### 단위 테스트
- `test_copy_claude_template_korean()`
- `test_copy_claude_template_english()`
- `test_copy_claude_template_fallback_to_english()`
- `test_template_structure_consistency()`

### 통합 테스트
- `test_init_with_korean_locale()`
- `test_init_with_english_locale()`
- `test_init_with_unsupported_locale()`

### E2E 테스트
- 한국어 템플릿으로 전체 워크플로우 (1-spec → 2-build → 3-sync)
- 영어 템플릿으로 전체 워크플로우

---

## 8. 문서화 계획

### 사용자 문서
- **README.md**: 한국어 메인, 영어 링크 제공
- **README.en.md**: 영어 번역
- **CHANGELOG.md**: 버전 변경 이력

### 개발자 문서
- **docs/i18n-guide.md**: 템플릿 다국어화 가이드
- **docs/contributing.md**: 번역 기여 방법
- **API Reference**: TemplateProcessor API 문서 업데이트

---

## 9. 배포 전략

### 배포 단계
1. **Alpha**: 내부 테스트 (v0.0.1-alpha)
2. **Beta**: 커뮤니티 테스트 (v0.0.1-beta)
3. **Stable**: 정식 릴리즈 (v0.0.1)

### 롤백 계획
- 기존 `.claude/` 템플릿 유지 (백업)
- 문제 발생 시 이전 버전으로 즉시 복구

---

## 10. 다음 단계 (Post-Implementation)

### 즉시 다음 단계
- `/alfred:2-build I18N-001` 실행 (TDD 구현)

### 향후 확장 계획
- v0.2.0: 일본어(ja), 중국어(zh) 추가
- v1.0.0: CLI 메시지 다국어화

---

_이 계획서는 TDD 구현 전 전략을 제시합니다. 구현 중 필요 시 수정될 수 있습니다._
