# 전면 회귀 분석 보고서: v0.12.0 → v0.14.0

**분석 일시**: 2025-11-03 19:00
**분석 범위**: v0.12.0 (2025-10-31) ~ v0.14.0 (2025-11-03)
**기간**: 3일
**총 커밋**: 81개
**회귀 커밋**: 20+ (Fix 타입 커밋)
**심각도**: 🔴 Critical

---

## 🚨 핵심 결론

**v0.12.0 이후 3일간 **최소 7개의 주요 회귀 사건(Regression Cascade)**이 발생했습니다.**

### 회귀의 특징

1. **연쇄 회귀 (Regression Cascade)**: 한 문제 해결 → 새 문제 발생 → 이전 문제 재발생
2. **부분 해결**: 일부만 고쳐지고 나머지는 그대로 유지
3. **템플릿 동기화 실패**: 로컬과 패키지 템플릿 간 불일치
4. **테스트 불안정성**: 패치마다 테스트 실패 반복

---

## 📋 주요 회귀 사건

### 1️⃣ Hook 환경 변수 3단계 회귀

**심각도**: 🔴 Critical | **영향**: Windows 사용자 | **상태**: ❌ 미해결

#### Timeline

```
2025-11-02 17:25:32  a2898697  ✅ 해결
  "command": "uv run .claude/hooks/alfred/session_start__show_project_info.py"
  → 환경 변수 제거, 상대 경로 사용

2025-11-03 17:12:45  3a3b8808  ⚠️  다른 시도
  Hook 섹션 전체 제거
  → auto-discovery 방식 시도

2025-11-03 17:13:44  aaff7388  ❌ 회귀
  "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/..."
  → a2898697의 해결책 완전히 제거, 환경 변수 복구
```

#### 현재 상태

```
❌ 로컬:      $CLAUDE_PROJECT_DIR 포함 (5개 hook 모두)
❌ 템플릿:    $CLAUDE_PROJECT_DIR 포함 (5개 hook 모두)
❌ Windows:   환경 변수 미설정으로 오류
```

#### 영향받는 Hook (5개)

- SessionStart:      `session_start__show_project_info.py`
- PreToolUse:        `pre_tool__auto_checkpoint.py`
- UserPromptSubmit:  `user_prompt__jit_load_docs.py`
- SessionEnd:        `session_end__cleanup.py`
- PostToolUse:       `post_tool__log_changes.py`

---

### 2️⃣ Hook Import 경로 3중 회귀

**심각도**: 🔴 Critical | **영향**: 모든 사용자 | **상태**: ⚠️ 부분 해결

#### 문제 사이클

```
2025-11-02 22:44:25  0e458bca  🔧 첫 복구
  fix: Restore package template hook file integrity
  → Hook 파일 무결성 손상 복구

2025-11-02 22:50:22  680bac00  🔧 두 번째 시도
  fix: Resolve PowerShell test failures
  - HookResult import 에러 수정
  - detect_language 문제 해결

2025-11-03 17:27:16  a07458f7  🔧 세 번째 시도
  fix: Resolve hook test import path issues
  - pytest conftest 경로 조정

2025-11-03 17:28:53  e626ab00  🔧 네 번째 시도
  fix: Resolve Alfred hooks sys.path and import issues
  - sys.path BEFORE imports 변경
  - handlers/__init__.py 백호환성 모듈 추가
```

#### 근본 원인

1. **sys.path 설정 타이밍**: Import 전에 설정되지 않음
2. **모듈 구조 불명확**: core vs shared vs handlers 간 순환 참조
3. **pytest 경로**: conftest 경로 불일치
4. **HookResult 클래스**: Import 경로 변경

#### 영향받는 파일

```
.claude/hooks/alfred/
├── alfred_hooks.py (sys.path 문제)
├── handlers/
│   ├── __init__.py (backward compatibility)
│   ├── session_start__show_project_info.py
│   ├── pre_tool__auto_checkpoint.py
│   ├── user_prompt__jit_load_docs.py
│   ├── session_end__cleanup.py
│   ├── post_tool__log_changes.py
│   └── core/ (import failures)
```

---

### 3️⃣ 사용자 설정 손실 및 복구

**심각도**: 🟡 Medium | **영향**: 프로젝트 설정 | **상태**: ✅ 해결됨

#### Timeline

```
2025-11-02 22:11:39  a52b156a  ✅ 복구
  fix: Restore full user nickname in config.json
  → 사용자 nickname이 손실되었다가 복구됨
```

#### 영향

```
손실된 설정:
{
  "user": {
    "nickname": "GOOS🪿엉아"  ← 손실됨
  }
}
```

#### 원인

초기화 프로세스에서 config.json이 부분적으로 생성되었거나 덮어써짐

---

### 4️⃣ Package Template 동기화 실패

**심각도**: 🔴 Critical | **영향**: 버전 관리 | **상태**: ⚠️ 부분 해결

#### 문제

```
로컬 프로젝트 (.claude/)
        ↕️ [동기화 불일치]
패키지 템플릿 (src/moai_adk/templates/.claude/)
```

#### 회귀 사건

1. **Hook 파일 무결성** (0e458bca)
   - Package template의 hook 파일이 손상됨
   - 로컬과 템플릿 동기화 불일치

2. **Settings 동기화** (aaff7388 + 3a3b8808)
   - 로컬과 템플릿의 settings.json이 다름
   - 환경 변수 버전으로 복구

3. **Command 파일 동기화**
   - 다국어 제거 → 영어로 통일
   - Template vs 로컬 간 메시지 불일치

---

### 5️⃣ Test 불안정성 / 실패 반복

**심각도**: 🟡 Medium | **영향**: 빌드/배포 | **상태**: ⚠️ 부분 해결

#### Regression

```
2025-11-02 22:50:22  680bac00
  fix: Resolve PowerShell test failures
  - HookResult import
  - detect_language failures
  - Performance test targets

2025-11-03 18:02:35  0c97350e
  Fix: Resolve 20 test failures for v0.14.0 deployment
  - 957 tests → 977 tests passing
  - 20개 새로운 test failure 해결
```

#### 패턴

```
v0.12.1: 996 tests
    ↓
v0.13.0: 957 tests (39개 실패 증가)
    ↓
680bac00: PowerShell 테스트 수정 시도
    ↓
v0.14.0 dev: 957 tests (여전히 실패)
    ↓
0c97350e: 20개 추가 수정 → 977 tests (신규 실패 감소)
```

**문제**: 테스트가 안정적이지 않음, 반복적인 수정 필요

---

### 6️⃣ TAG 검증 에러 반복

**심각도**: 🟡 Medium | **영향**: 배포 | **상태**: ✅ 해결됨

#### Timeline

```
2025-11-03 18:24:24  07daac41
  fix: Resolve all TAG validation errors
  - complete v0.14.0 deployment preparation
```

#### 영향받는 파일

- 여러 SPEC 문서의 중복 @TAG
- Skill 파일의 예제 @TAG
- 문서 내 reference @TAG

---

### 7️⃣ uv tool 명령 불일치

**심각도**: 🟡 Medium | **영향**: 사용자 경험 | **상태**: ✅ 해결됨

#### Timeline

```
2025-11-02 21:57:39  52d28afd
  fix: Update upgrade command to use uv tool instead of uv pip
  → README vs SessionStart hook 명령 불일치
```

#### 해결책

```
Before:  "uv pip install --upgrade moai-adk>=VERSION"
After:   "uv tool upgrade moai-adk"
```

---

## 📊 회귀 통계

### 회귀 유형 분류

| 유형 | 개수 | 심각도 | 상태 |
|------|------|--------|------|
| Hook 환경 변수 | 1 | 🔴 Critical | ❌ 미해결 |
| Hook Import 경로 | 4 | 🔴 Critical | ⚠️ 부분 해결 |
| Template 동기화 | 3+ | 🔴 Critical | ⚠️ 부분 해결 |
| 설정 손실/복구 | 1 | 🟡 Medium | ✅ 해결 |
| Test 실패 반복 | 2 | 🟡 Medium | ⚠️ 부분 해결 |
| TAG 검증 | 1 | 🟡 Medium | ✅ 해결 |
| 명령 불일치 | 1 | 🟡 Medium | ✅ 해결 |
| **총계** | **13+** | | **6/13 해결** |

### 시간대별 회귀 분포

```
2025-11-02
  22:11 - a52b156a (설정 손실 복구)
  22:44 - 0e458bca (Template 무결성 복구)
  22:50 - 680bac00 (PowerShell 테스트)
  21:57 - 52d28afd (uv tool 명령)
  17:25 - a2898697 (Hook 환경 변수 해결 시도)

2025-11-03
  17:12 - 3a3b8808 (Hook 섹션 제거 시도)
  17:13 - aaff7388 (Hook 환경 변수 회귀) ← ❌ 심각
  17:27 - a07458f7 (Hook import 경로 수정)
  17:28 - e626ab00 (Hook sys.path 수정)
  18:02 - 0c97350e (20개 Test failure 수정)
  18:24 - 07daac41 (TAG 검증 에러 수정)
```

---

## 🔴 회귀의 근본 원인

### 1. **템플릿/로컬 동기화 부족**

```
패키지 템플릿 (src/moai_adk/templates/.claude/)
  ↕️ [수동 동기화만 가능]
로컬 프로젝트 (.claude/)
  ↕️ [git ignore로 추적 불가]
```

**문제**:
- 자동 동기화 메커니즘 부재
- 한 쪽을 수정하면 다른 쪽이 뒤처짐
- 회귀하기 쉬운 구조

### 2. **Hook 시스템의 복잡성**

```
.claude/hooks/alfred/
├── alfred_hooks.py (진입점)
├── shared/
│   ├── core/
│   │   ├── checkpoint.py
│   │   ├── context.py
│   │   ├── ttl_cache.py
│   │   └── version_cache.py
│   └── handlers/
│       └── ...
└── handlers/ (backward compat alias)
    ├── session_start__show_project_info.py
    └── ...
```

**문제**:
- 순환 import 위험
- sys.path 관리의 어려움
- Windows vs Unix 경로 처리
- Python 3.13 호환성

### 3. **테스트 커버리지 부족**

- 20개 테스트가 반복적으로 실패
- Cross-platform 테스트 불안정
- Mock/fixtures 부족

### 4. **Git History 추적 부족**

- 이전 커밋의 의도 미파악
- "Restore" 커밋 시 잘못된 버전 복구
- Merge conflict 해결 미흡

---

## 🎯 문제점 및 권장사항

### 즉시 조치 필요

| 문제 | 우선순위 | 권장 조치 |
|------|---------|---------|
| Hook 환경 변수 | 🔴 P0 | a2898697 재적용 + 상대 경로로 통일 |
| Hook import 경로 | 🔴 P0 | sys.path 관리 정책 수립, 테스트 강화 |
| Template 동기화 | 🔴 P0 | 자동 동기화 스크립트 or symlink |
| Test 안정성 | 🟡 P1 | Cross-platform 테스트 인프라 개선 |

### 구조적 개선

1. **자동 동기화**
   ```bash
   # Option A: rsync 스크립트
   rsync -av src/moai_adk/templates/.claude/ .claude/

   # Option B: Symlink (권장)
   ln -s src/moai_adk/templates/.claude .claude

   # Option C: Git submodule (advanced)
   ```

2. **Hook 시스템 리팩토링**
   - Single entry point (alfred_hooks.py)
   - sys.path 정책 명확화
   - pytest conftest 통일
   - Windows 경로 처리 표준화

3. **테스트 강화**
   - Windows/Mac/Linux CI 파이프라인
   - Mock 개선
   - Fixtures 명확화
   - Regression 테스트

4. **Git 정책**
   - Merge commit에 이유 기록
   - Revert/Restore 전에 history 검토
   - Template 변경 → 로컬 변경 → 테스트 순서 강제

---

## 📈 회귀 패턴 분석

### 패턴 1: 반복적 시도 후 실패

```
문제 발생 → 첫 시도 (부분 해결) → 다른 시도 → 회귀
```

예시:
- a2898697 (환경 변수 제거) → 3a3b8808 (전체 제거) → aaff7388 (원점 복구)

### 패턴 2: 일부만 고쳐지는 문제

```
Hook import 에러가 4번 연속 수정됨 (0e458bca, 680bac00, a07458f7, e626ab00)
→ 근본 원인 미해결, 증상만 치료
```

### 패턴 3: Template 동기화 실패

```
로컬 수정 → 패키지 템플릿 미반영 → 버전 배포 → 사용자 문제 발생
```

---

## ✅ 검증 체크리스트

- [ ] Hook 환경 변수: a2898697 변경사항 재적용
- [ ] Hook import: sys.path 관리 정책 검토
- [ ] Template 동기화: 자동화 메커니즘 구현
- [ ] Test: Cross-platform 테스트 추가
- [ ] Settings: config.json 초기화 프로세스 검증
- [ ] TAG: 중복 @TAG 제거 확인
- [ ] uv tool: 문서와 코드 일치성 검증

---

## 📌 결론

**v0.12.0 이후 3일간의 개발에서:**

1. ✅ **7개 회귀 사건 중 3개 완전 해결**
2. ⚠️ **4개 부분 해결** (근본 원인 미해결)
3. ❌ **1개 미해결** (Hook 환경 변수)

**근본 원인**:
- Template/로컬 동기화 자동화 부족
- Hook 시스템의 복잡성
- Test 커버리지 부족
- Git history 추적 미흡

**권장**:
즉시 Hook 환경 변수 문제를 수정하고,
구조적인 자동화 메커니즘을 도입할 필요가 있습니다.

---

**분석 완료**: 🎩 Alfred
**상태**: 🔴 심각한 회귀 다수 발견
**신뢰도**: 🟢 High (81개 커밋 전수 분석)

