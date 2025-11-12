# Implementation Plan: README-UX-001

## 구현 계획 개요

이 문서는 SPEC-README-UX-001(README.md uv 설치 방법 개선)의 구현 계획을 정의합니다.

**참고**: 이 SPEC은 문서 수정이므로 TDD 사이클 대신 직접 수정 방식을 사용합니다.

---

## 구현 전략

### 문서 수정 워크플로우

1. **백업**: 변경 전 현재 README.md 상태 확인
2. **수정**: 3곳 수정 (라인 360, 1249, 1403)
3. **검증**: 일관성 및 링크 확인
4. **커밋**: 변경사항 커밋

---

## Phase 1: 백업 및 현재 상태 확인

### 작업 목록

#### 1.1. 현재 설치 명령어 위치 확인

```bash
# uv pip install 명령어 검색
grep -n "uv pip install moai-adk" README.md

# 예상 결과:
# 360:uv pip install moai-adk
# 1249:uv pip install --upgrade moai-adk
# 1403:uv pip install moai-adk --force-reinstall
```

#### 1.2. 백업 생성 (선택사항)

```bash
# Checkpoint 자동 생성 (Alfred Hooks)
# 또는 수동 백업
cp README.md README.md.backup
```

---

## Phase 2: README.md 수정

### 작업 목록

#### 2.1. Quick Start 섹션 수정 (라인 360 근처)

**검색 패턴**: `uv pip install moai-adk` (처음 출현)

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

**추가 설명 (선택사항)**:
```markdown
#### uv tool install의 장점

- **격리된 환경**: 각 도구가 독립된 샌드박스에서 실행되어 의존성 충돌 방지
- **전역 접근**: 어떤 프로젝트에서든 `moai-adk` 명령어 사용 가능
- **간편한 관리**: `uv tool upgrade` 등으로 쉽게 업데이트

자세한 내용은 [uv 공식 문서](https://github.com/astral-sh/uv)를 참조하세요.
```

---

#### 2.2. 업그레이드 섹션 수정 (라인 1249 근처)

**검색 패턴**: `uv pip install --upgrade moai-adk`

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

#### 2.3. 재설치 섹션 수정 (라인 1403 근처)

**검색 패턴**: `uv pip install moai-adk --force-reinstall`

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

## Phase 3: 검증

### 작업 목록

#### 3.1. 일관성 검증

```bash
# uv tool install 출현 횟수 확인 (최소 3회 이상)
grep -c "uv tool install" README.md

# "권장" 키워드 확인 (최소 3회)
grep -c "권장" README.md

# "대안" 키워드 확인 (최소 3회)
grep -c "대안" README.md
```

**기대 결과**:
- `uv tool install`: 3회 이상
- `권장`: 3회 이상
- `대안`: 3회 이상

---

#### 3.2. 링크 검증

```bash
# uv 공식 문서 링크 확인
grep "https://github.com/astral-sh/uv" README.md

# 기대 결과: 링크 존재
```

**수동 검증**:
- 브라우저에서 https://github.com/astral-sh/uv 접속 확인

---

#### 3.3. 개발자용 설치 방법 유지 확인

```bash
# 편집 가능 설치 확인 (변경 없어야 함)
grep "uv pip install -e" README.md

# 기대 결과: 존재 (개발자용 섹션은 그대로 유지)
```

---

#### 3.4. Markdown 렌더링 확인 (선택사항)

```bash
# VSCode 또는 GitHub Preview로 README.md 확인
# - 코드 블록이 올바르게 렌더링되는지
# - 주석이 명확하게 표시되는지
# - 링크가 클릭 가능한지
```

---

## Phase 4: 커밋

### 커밋 메시지 (Locale: ko)

```bash
git add README.md

git commit -m "$(cat <<'EOF'
📝 DOCS: README.md uv 설치 방법을 tool 모드로 개선

- Quick Start: uv tool install 권장, uv pip install 대안 병기
- 업그레이드: uv tool upgrade 추가
- 재설치: uv tool uninstall/install 추가
- 전체 3곳 수정으로 일관성 유지

Fixes #35

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

---

## 예상 산출물

### 문서 변경
- **수정**: `README.md` (3곳)
  - 라인 360 근처: Quick Start 섹션
  - 라인 1249 근처: 업그레이드 섹션
  - 라인 1403 근처: 재설치 섹션
- **추가** (선택): uv tool install 장점 설명 섹션

### 커밋 이력
- **단일 커밋**: "DOCS: README.md uv 설치 방법 개선"
- **이슈 종료**: Fixes #35

---

## 검증 체크리스트

### 수정 완료 확인
- [ ] Quick Start 섹션: `uv tool install` 권장 추가
- [ ] 업그레이드 섹션: `uv tool upgrade` 추가
- [ ] 재설치 섹션: `uv tool uninstall/install` 추가

### 품질 확인
- [ ] 3곳 모두 동일한 형식 (권장 → 대안)
- [ ] 주석으로 각 방법 설명 추가
- [ ] uv 공식 문서 링크 추가 (선택)

### 기능 확인
- [ ] 개발자용 `-e` 설치 방법 유지
- [ ] Markdown 렌더링 정상
- [ ] 링크 클릭 가능

### 일관성 확인
- [ ] `grep -c "uv tool install" README.md`: 3회 이상
- [ ] `grep -c "권장" README.md`: 3회 이상
- [ ] 모든 섹션 동일한 용어 사용

---

## 다음 단계

1. README.md 직접 수정 (Edit tool 사용)
2. 검증 스크립트 실행
3. 커밋 및 푸시 (git-manager 사용)
4. GitHub Issue #35 종료
5. `/alfred:3-sync` 실행하여 문서 동기화

---

## 실제 사용자 테스트 (선택사항)

### 신규 사용자 관점
```bash
# README.md 보고 설치 시도
uv tool install moai-adk
moai-adk --version

# 기대 결과: 최신 버전 설치 및 실행
```

### 업그레이드 시나리오
```bash
# 기존 설치 확인
uv tool list

# 업그레이드
uv tool upgrade moai-adk

# 기대 결과: 최신 버전으로 업데이트
```

### 문제 해결 시나리오
```bash
# 재설치
uv tool uninstall moai-adk
uv tool install moai-adk

# 기대 결과: 깨끗하게 재설치
```

---

_이 계획은 `/alfred:2-run SPEC-README-UX-001` 또는 직접 수정으로 구현됩니다._
