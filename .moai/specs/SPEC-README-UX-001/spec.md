---
id: README-UX-001
version: 0.1.0
status: completed
created: 2025-10-19
updated: 2025-10-19
priority: medium
category: docs
labels:
  - documentation
  - uv
  - installation
  - user-experience
related_issue: "https://github.com/modu-ai/moai-adk/issues/35"
scope:
  files:
    - README.md
---


## HISTORY

### v0.1.0 (2025-10-19)
- **DOCS COMPLETED**: README.md 수정 완료
- **CHANGED**: 3개 위치에서 `uv tool install` 권장, `uv pip install` 대안 병기
  - Quick Start (라인 360): uv tool install 추가
  - 업그레이드 (라인 1249): uv tool upgrade 추가
  - 재설치 (라인 1403): uv tool uninstall/install 추가
- **COMMITS**:
  - c04bb3d: 📝 DOCS: README.md uv 설치 방법을 tool 모드로 개선
  - bd8dcc8: 📝 SPEC: README.md uv 설치 방법 개선 명세 작성
- **REVIEW**: GitHub Issue #35 답변 완료

### v0.0.1 (2025-10-19)
- **INITIAL**: README.md 설치 방법을 uv 권장 방식으로 개선하는 명세 작성
- **SCOPE**: `uv pip install` → `uv tool install` 변경 (3곳)
- **CONTEXT**: GitHub Issue #35 - uv 공식 권장 방식인 tool install 적용
- **RELATED**: https://github.com/modu-ai/moai-adk/issues/35

---

## Environment (환경)

### 시스템 환경
- **대상 파일**: `README.md`
- **변경 위치**: 3곳 (라인 360, 1249, 1403)
- **uv 버전**: 최신 안정 버전 (권장사항 기준)

### 전제 조건
- uv 패키지 매니저가 설치되어 있음
- uv는 `pip`와 `tool` 두 가지 설치 방식을 지원:
  - **pip 모드**: 현재 가상 환경에 설치
  - **tool 모드**: 격리된 샌드박스에 설치 (권장)

---

## Assumptions (가정)

### uv tool install의 장점
1. **샌드박스 격리**: 각 도구가 독립된 환경에서 실행되어 의존성 충돌 방지
2. **전역 접근**: 어떤 디렉토리에서든 `moai-adk` CLI 사용 가능
3. **간편한 업데이트**: `uv tool upgrade moai-adk`로 최신 버전 유지
4. **uv 공식 권장**: uv 문서에서 CLI 도구 설치 시 권장하는 방식

### 사용자 경험 개선
- **명확한 안내**: 권장 방법(`uv tool install`)을 먼저 제시
- **대안 제공**: 레거시 환경을 위한 `uv pip install` 대안 명시
- **일관성**: 문서 전체에서 동일한 권장사항 적용

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

- 시스템은 uv 권장 설치 방식인 `uv tool install moai-adk`를 제공해야 한다
- 시스템은 README.md의 모든 설치 섹션에서 일관된 방법을 안내해야 한다
- 시스템은 사용자에게 샌드박스 격리의 장점을 설명해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN 사용자가 "Quick Start" 섹션을 읽으면, 시스템은 `uv tool install moai-adk`를 가장 먼저 제시해야 한다
- WHEN 사용자가 "업그레이드" 섹션을 읽으면, 시스템은 `uv tool upgrade moai-adk` 방법을 안내해야 한다
- WHEN 사용자가 "재설치" 섹션을 읽으면, 시스템은 tool 모드 기준 재설치 방법을 제시해야 한다

### State-driven Requirements (상태 기반)

- WHILE 일반 사용자 대상 문서일 때, 시스템은 `uv tool install`을 기본 권장해야 한다
- WHILE 개발자 대상 섹션일 때, 시스템은 `uv pip install -e ".[dev]"` (편집 가능 설치)를 유지해야 한다

### Optional Features (선택적 기능)

- WHERE 레거시 환경 호환성이 필요하면, 시스템은 `uv pip install` 대안을 제시할 수 있다
- WHERE pip 사용자를 위한 안내가 필요하면, 시스템은 `pip install moai-adk` 방법도 추가할 수 있다

### Constraints (제약사항)

- 설치 방법은 uv 공식 문서 권장사항과 일치해야 한다
- README.md의 3곳 모두 동일한 형식으로 수정해야 한다 (일관성)
- 개발자용 편집 가능 설치(`-e` 플래그)는 유지해야 한다

---

## Technical Design (기술 설계)

### 현재 구현 (문제점)

**라인 360 (Quick Start 섹션)**:
```bash
uv pip install moai-adk
```

**라인 1249 (업그레이드 섹션)**:
```bash
uv pip install --upgrade moai-adk
```

**라인 1403 (재설치 섹션)**:
```bash
uv pip install moai-adk --force-reinstall
```

**문제점**:
1. `uv pip install`은 현재 가상 환경에 설치 → 환경별 관리 필요
2. uv 공식 권장 방식인 `tool` 모드를 사용하지 않음
3. 샌드박스 격리의 장점을 활용하지 못함

---

### 개선된 구현 (제안)

**라인 360 (Quick Start 섹션)**:
```bash
# 권장: uv tool 모드 (샌드박스 격리)
uv tool install moai-adk

# 대안: 현재 가상 환경에 설치
uv pip install moai-adk
```

**라인 1249 (업그레이드 섹션)**:
```bash
# tool 모드 업그레이드
uv tool upgrade moai-adk

# pip 모드 업그레이드 (레거시)
uv pip install --upgrade moai-adk
```

**라인 1403 (재설치 섹션)**:
```bash
# tool 모드 재설치
uv tool uninstall moai-adk
uv tool install moai-adk

# pip 모드 재설치 (레거시)
uv pip install moai-adk --force-reinstall
```

**개선점**:
1. **권장 방법 우선**: `uv tool install` 명령어를 가장 먼저 제시
2. **대안 제공**: 기존 `uv pip install` 방법도 유지 (하위 호환성)
3. **명확한 설명**: 각 방법의 차이점 (샌드박스 vs 가상 환경) 명시

---


- **DOC**: `README.md` (3곳 수정)
- **RELATED ISSUE**: https://github.com/modu-ai/moai-adk/issues/35
- **REFERENCE**: https://github.com/astral-sh/uv (uv 공식 문서)

---

## Success Criteria (성공 기준)

### 기능 검증
1. README.md의 3곳 모두 `uv tool install moai-adk`를 권장 방법으로 제시
2. 각 섹션에 레거시 대안(`uv pip install`) 병기
3. 설명 추가: 샌드박스 격리 장점 명시

### 품질 기준
- 문서 일관성: 3곳 모두 동일한 형식으로 작성
- 링크 정상 작동: uv 공식 문서 참조 링크 추가
- 가독성: 코드 블록과 주석으로 명확히 구분

### 비기능 요구사항
- 사용자 친화성: 초보자도 이해할 수 있는 설명
- 접근성: 권장 방법과 대안을 모두 제시
- 유지보수성: 향후 uv 업데이트 시 쉽게 수정 가능

---

## Implementation Notes (구현 참고사항)

### 변경 위치 상세

#### 1. Quick Start 섹션 (라인 360 근처)
**검색 키워드**: `uv pip install moai-adk`
**변경 전**:
```bash
uv pip install moai-adk
```

**변경 후**:
```bash
# 권장: uv tool 모드 (샌드박스 격리, 전역 접근)
uv tool install moai-adk

# 대안: 현재 가상 환경에 설치
uv pip install moai-adk
```

---

#### 2. 업그레이드 섹션 (라인 1249 근처)
**검색 키워드**: `uv pip install --upgrade moai-adk`
**변경 전**:
```bash
uv pip install --upgrade moai-adk
```

**변경 후**:
```bash
# tool 모드 (권장)
uv tool upgrade moai-adk

# pip 모드 (레거시)
uv pip install --upgrade moai-adk
```

---

#### 3. 재설치 섹션 (라인 1403 근처)
**검색 키워드**: `uv pip install moai-adk --force-reinstall`
**변경 전**:
```bash
uv pip install moai-adk --force-reinstall
```

**변경 후**:
```bash
# tool 모드 (권장)
uv tool uninstall moai-adk
uv tool install moai-adk

# pip 모드 (레거시)
uv pip install moai-adk --force-reinstall
```

---

### 추가 개선 사항 (선택)

#### Quick Start 섹션에 uv tool 장점 설명 추가
```markdown
### uv tool install의 장점

- **격리된 환경**: 각 도구가 독립된 샌드박스에서 실행되어 의존성 충돌 방지
- **전역 접근**: 어떤 프로젝트에서든 `moai-adk` 명령어 사용 가능
- **간편한 관리**: `uv tool list`, `uv tool upgrade` 등으로 쉽게 관리

자세한 내용은 [uv 공식 문서](https://github.com/astral-sh/uv)를 참조하세요.
```

---

## References

- **GitHub Issue**: https://github.com/modu-ai/moai-adk/issues/35
- **uv 공식 문서**: https://github.com/astral-sh/uv
- **uv tool 가이드**: https://github.com/astral-sh/uv#tool-management
- **MoAI-ADK tech.md**: `.moai/project/tech.md` (uv 선택 이유 참조)

---

_이 SPEC은 문서 수정이므로 TDD 사이클이 아닌 직접 수정 방식으로 구현됩니다._
