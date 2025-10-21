# Acceptance Criteria: README-UX-001

## 인수 기준 개요

이 문서는 SPEC-README-UX-001(README.md uv 설치 방법 개선)의 인수 기준을 정의합니다.

---

## AC-1: Quick Start 섹션 개선

### Given (전제 조건)
- README.md 파일의 Quick Start 섹션 (라인 360 근처)
- 현재: `uv pip install moai-adk` 단일 명령어만 제시

### When (실행)
- README.md Quick Start 섹션 확인

### Then (기대 결과)
- **권장 방법**이 먼저 표시되어야 함: `uv tool install moai-adk`
- **대안**이 추가로 제시되어야 함: `uv pip install moai-adk`
- **설명**이 포함되어야 함: 각 방법의 차이점 (샌드박스 vs 가상 환경)

### 검증 방법
```bash
# 라인 360 근처 확인
grep -A 5 "uv tool install moai-adk" README.md

# 기대 결과:
# - "권장" 키워드 포함
# - "샌드박스 격리" 또는 "전역 접근" 설명 포함
# - 대안으로 "uv pip install" 병기
```

---

## AC-2: 업그레이드 섹션 개선

### Given (전제 조건)
- README.md 파일의 업그레이드 섹션 (라인 1249 근처)
- 현재: `uv pip install --upgrade moai-adk`만 제시

### When (실행)
- README.md 업그레이드 섹션 확인

### Then (기대 결과)
- **tool 모드 업그레이드** 명령어가 먼저 제시: `uv tool upgrade moai-adk`
- **pip 모드 업그레이드** 명령어가 대안으로 제시: `uv pip install --upgrade moai-adk`
- 각 방법에 설명 추가

### 검증 방법
```bash
# 라인 1249 근처 확인
grep -A 3 "uv tool upgrade moai-adk" README.md

# 기대 결과:
# - "tool 모드" 키워드 포함
# - "pip 모드" 또는 "레거시" 설명 포함
```

---

## AC-3: 재설치 섹션 개선

### Given (전제 조건)
- README.md 파일의 재설치/문제 해결 섹션 (라인 1403 근처)
- 현재: `uv pip install moai-adk --force-reinstall`만 제시

### When (실행)
- README.md 재설치 섹션 확인

### Then (기대 결과)
- **tool 모드 재설치** 명령어가 먼저 제시:
  ```bash
  uv tool uninstall moai-adk
  uv tool install moai-adk
  ```
- **pip 모드 재설치** 명령어가 대안으로 제시: `uv pip install moai-adk --force-reinstall`

### 검증 방법
```bash
# 라인 1403 근처 확인
grep -A 5 "uv tool uninstall moai-adk" README.md

# 기대 결과:
# - "uv tool uninstall" 다음 "uv tool install" 명령어
# - 대안으로 "--force-reinstall" 포함
```

---

## AC-4: 문서 일관성 검증

### Given (전제 조건)
- README.md의 3곳 모두 수정 완료
- 각 섹션이 동일한 형식을 따름

### When (실행)
- README.md 전체 검색

### Then (기대 결과)
- `uv tool install` 명령어가 3곳 모두에서 **권장 방법**으로 제시되어야 함
- 각 섹션이 동일한 형식 (권장 → 대안 → 설명)을 따라야 함
- 레거시 `uv pip install` 방법이 모두 **대안**으로 명시되어야 함

### 검증 방법
```bash
# uv tool install 출현 횟수 확인
grep -c "uv tool install" README.md

# 기대 결과: 3회 이상 (Quick Start, 업그레이드, 재설치)

# 일관성 검증: "권장" 키워드 확인
grep -c "권장" README.md

# 기대 결과: 3회 이상
```

---

## AC-5: uv tool 장점 설명 추가 (선택사항)

### Given (전제 조건)
- README.md Quick Start 섹션

### When (실행)
- uv tool install 명령어 주변 확인

### Then (기대 결과)
- **샌드박스 격리** 장점 설명 포함
- **전역 접근** 장점 설명 포함
- uv 공식 문서 링크 추가 (선택)

### 검증 방법
```bash
# 샌드박스 관련 설명 확인
grep -i "샌드박스\|격리\|sandbox" README.md

# uv 공식 문서 링크 확인
grep "https://github.com/astral-sh/uv" README.md
```

---

## AC-6: 개발자용 설치 방법 유지

### Given (전제 조건)
- README.md의 개발자용 설치 섹션 (editable install)
- 기존: `uv pip install -e ".[dev,security]"`

### When (실행)
- 개발자용 설치 섹션 확인

### Then (기대 결과)
- **편집 가능 설치** 방법이 유지되어야 함: `uv pip install -e ".[dev,security]"`
- tool 모드로 변경하지 않아야 함 (개발자는 pip 모드 필요)

### 검증 방법
```bash
# 편집 가능 설치 확인
grep "uv pip install -e" README.md

# 기대 결과: 존재 (변경 없음)
```

---

## 비기능 요구사항 검증

### 가독성
- **기준**: 초보자도 이해할 수 있는 명확한 설명
- **검증**: README.md 읽고 각 설치 방법의 차이 이해 가능 여부

### 일관성
- **기준**: 3곳 모두 동일한 형식과 용어 사용
- **검증**: `grep -A 5 "uv tool install" README.md` 결과 비교

### 링크 검증
- **기준**: 모든 외부 링크가 정상 작동
- **검증**: uv 공식 문서 링크 클릭 시 접근 가능

---

## 체크리스트

- [ ] AC-1: Quick Start 섹션에 `uv tool install` 권장 방법 추가
- [ ] AC-2: 업그레이드 섹션에 `uv tool upgrade` 추가
- [ ] AC-3: 재설치 섹션에 `uv tool uninstall/install` 추가
- [ ] AC-4: 문서 일관성 검증 (3곳 모두 동일 형식)
- [ ] AC-5: uv tool 장점 설명 추가 (선택)
- [ ] AC-6: 개발자용 `-e` 설치 방법 유지
- [ ] 가독성: 초보자 친화적 설명
- [ ] 일관성: 동일한 용어 및 형식 사용
- [ ] 링크 검증: uv 공식 문서 링크 정상 작동

---

## 추가 검증 (선택사항)

### 실제 사용자 테스트
1. **신규 사용자**: `uv tool install moai-adk` 실행 후 `moai-adk --version` 확인
2. **기존 사용자**: `uv tool upgrade moai-adk` 실행 후 최신 버전 확인
3. **문제 해결**: `uv tool uninstall` → `uv tool install` 재설치 확인

### 크로스 플랫폼 테스트
- **macOS**: `uv tool install moai-adk` 정상 작동
- **Linux**: `uv tool install moai-adk` 정상 작동
- **Windows**: `uv tool install moai-adk` 정상 작동 (PowerShell/Git Bash)

---

_이 인수 기준은 `/alfred:2-run SPEC-README-UX-001` 또는 직접 수정 후 검증됩니다._
