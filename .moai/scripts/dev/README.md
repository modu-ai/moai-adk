# MoAI-ADK 개발자 유틸리티 스크립트

> **주의**: 이 디렉토리의 스크립트는 **MoAI-ADK 패키지 개발자 전용** 도구입니다.
>
> 패키지 배포본에는 포함되지 않습니다. 로컬 개발 및 유지보수 목적으로만 사용하세요.

---

## 📋 스크립트 목록

### 1. `fix-missing-spec-tags.py`

**목적**: SPEC 파일에서 누락된 `@SPEC:` 태그를 자동으로 추가

**사용 대상**:
- SPEC 파일 구조가 변경되었을 때
- TAG 시스템 마이그레이션 후 누락된 태그 정정
- 기존 SPEC 파일에 TAG 메타데이터 추가

**기능**:
- `.moai/specs/` 디렉토리 모든 SPEC 파일 검사
- `plan.md`, `acceptance.md`에 누락된 `@SPEC:` 태그 추가
- 제목 라인에서 `@SPEC:{ID}` 형식으로 변환
- 변경 사항 상세 기록

**사용 방법**:
```bash
# 검토 모드 (변경 없음)
python3 .moai/scripts/dev/fix-missing-spec-tags.py --dry-run

# 실행 모드 (파일 수정)
python3 .moai/scripts/dev/fix-missing-spec-tags.py

# 특정 SPEC ID만 처리
python3 .moai/scripts/dev/fix-missing-spec-tags.py --spec-id SPEC-CLI-ANALYSIS-001
```

---

### 2. `lint_korean_docs.py`

**목적**: 한국어 문서의 문법, 스타일 검사

**사용 대상**:
- 한국어로 작성된 문서 품질 검증
- 띄어쓰기, 문법 오류 감지
- 문서 스타일 일관성 확인
- SPEC, README 등 한국어 파일 점검

**기능**:
- 한국어 띄어쓰기 규칙 검증
- 마크다운 형식 일관성 확인
- 문장 부호 사용 규칙 검증
- 모순되는 표현 감지

**사용 방법**:
```bash
# 전체 문서 검사
python3 .moai/scripts/dev/lint_korean_docs.py

# 특정 파일 검사
python3 .moai/scripts/dev/lint_korean_docs.py --file path/to/file.md

# 상세 보고서 생성
python3 .moai/scripts/dev/lint_korean_docs.py --report
```

---

### 3. `validate_mermaid_diagrams.py`

**목적**: 마크다운 파일의 Mermaid 다이어그램 유효성 검증

**사용 대상**:
- 문서에 포함된 다이어그램 문법 검사
- 다이어그램 렌더링 오류 감지
- 아키텍처 문서 다이어그램 검증
- 플로우차트, 클래스 다이어그램 등 검사

**기능**:
- 모든 `.md` 파일의 Mermaid 블록 추출
- 마크다운 문법 검증
- 렌더링 호환성 확인
- 오류가 있는 다이어그램 상세 보고

**사용 방법**:
```bash
# 전체 다이어그램 검사
python3 .moai/scripts/dev/validate_mermaid_diagrams.py

# 특정 디렉토리만 검사
python3 .moai/scripts/dev/validate_mermaid_diagrams.py --path docs/

# 상세 분석 보고서
python3 .moai/scripts/dev/validate_mermaid_diagrams.py --detailed
```

---

### 4. `init-dev-config.sh`

**목적**: 개발 환경의 `.moai/config.json` 초기화

**사용 대상**:
- 패키지를 `pip install -e .` 또는 `uv pip install -e .`로 설치한 후
- 개발 환경 설정 자동화
- 버전 정보 동기화

**기능**:
- `pyproject.toml`에서 실제 버전 정보 추출
- `config.json`의 템플릿 변수 치환
- 개발 환경의 필요한 설정값 자동 주입

**사용 방법**:
```bash
# 개발 환경 설정 초기화
bash .moai/scripts/dev/init-dev-config.sh

# 설정 내용 검증
cat .moai/config.json | jq '.version'
```

**사용 시기**:
- 패키지를 editable mode로 설치한 직후
- 버전 업데이트 후 개발 환경 재설정 필요 시

---

### 5. `skill-pattern-validator.sh`

**목적**: MoAI-ADK Skills의 구조와 형식 검증

**사용 대상**:
- 새 Skill 생성 후 구조 검증
- Skill 메타데이터 완성도 확인
- 패키지 배포 전 Skill 표준 준수 검사
- Skill 문서화 완성도 확인

**기능**:
- SKILL.md 메타데이터 검증 (name, version, status)
- 필수 섹션 존재 확인 (설명, 사용 방법, 예제)
- @TAG 시스템 연계 검증
- 파일 구조 일관성 확인

**사용 방법**:
```bash
# 모든 Skills 검증
bash .moai/scripts/dev/skill-pattern-validator.sh

# 특정 Skill만 검증
bash .moai/scripts/dev/skill-pattern-validator.sh moai-lang-python

# 상세 리포트 생성
bash .moai/scripts/dev/skill-pattern-validator.sh --detailed
```

---

## 🚀 개발자 워크플로우

### 일반적인 사용 순서

```bash
# 1. 환경 초기화 (처음 설치 시)
bash .moai/scripts/dev/init-dev-config.sh

# 2. SPEC 파일 검증 (새 SPEC 추가 후)
python3 .moai/scripts/dev/fix-missing-spec-tags.py --dry-run

# 3. 문서 품질 검사 (한국어 문서 작성 후)
python3 .moai/scripts/dev/lint_korean_docs.py

# 4. 다이어그램 검증 (아키텍처 문서 업데이트 후)
python3 .moai/scripts/dev/validate_mermaid_diagrams.py

# 5. Skill 검증 (새 Skill 추가 후)
bash .moai/scripts/dev/skill-pattern-validator.sh
```

---

## 📊 패키지 배포 설정

**중요**: 이 스크립트들은 패키지 배포본에 포함되지 않습니다.

`.gitignore` 설정:
```gitignore
# 개발자 스크립트 (배포 제외)
.moai/scripts/dev/
```

`pyproject.toml` 제외 설정:
```toml
[tool.poetry]
exclude = [
  ".moai/scripts/dev/*",
]
```

---

## 🔧 유지보수

### 스크립트 추가 기준

이 디렉토리에는 다음 조건을 만족하는 스크립트만 추가됩니다:

✅ **포함 기준**:
- MoAI-ADK 패키지 개발 및 유지보수 목적
- 패키지 배포본에 불필요
- 개발자 전용 도구
- 로컬 개발 환경에서만 사용

❌ **제외 기준**:
- 최종 사용자가 실행해야 하는 스크립트
- 패키지 기능의 일부
- 배포본에 포함되어야 하는 도구

### 새 스크립트 추가

새로운 개발자 스크립트를 추가할 때:

1. **명확한 목적 정의**: 무엇을 하는가?
2. **개발자 대상 확인**: 최종 사용자인가?
3. **이 README 업데이트**: 스크립트 문서화
4. **Git 커밋**: `chore(dev-scripts): Add {script-name}`

---

## 📚 관련 문서

- **Skill 시스템**: `.moai/skills/moai-foundation-tags/`
- **TAG 시스템**: `.moai/specs/TAG-REFERENCE.md`
- **개발 가이드**: `CONTRIBUTING.md`
- **패키지 구조**: `pyproject.toml`

---

**마지막 업데이트**: 2025-11-13
**상태**: Production Ready (개발자 도구)
