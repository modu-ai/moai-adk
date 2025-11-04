# 🔒 Git 태그 보호 정책 강화 가이드

## 개요

MoAI-ADK 팀 모드에서 GitFlow 순서를 엄격하게 강제하고 태그 우회를 완전히 차단하기 위한 종합적인 보호 정책입니다.

## 🚨 해결된 보안 취약점

### 기존 문제점
1. **pre-push hook 미흡**: 태그 생성에 대한 아무런 검증 없음
2. **GitHub Actions 부재**: 태그 푸시 자동 검증 부재
3. **수동 우회 가능**: `git tag` & `git push`로 제약 없이 생성 가능

### 강화된 보호 장치
1. **pre-push hook 강화**: 태그 형식, 소스 브랜치, 중복 검증
2. **GitHub Actions 검증**: CI/CD 파이프라인에서 태그 유효성 확인
3. **팀 모드 엄격 적용**: feature → develop → main 순서 강제

## 🛡️ 구현된 강화 정책

### 1. 로컬 pre-push Hook 강화

#### 파일 위치: `.git/hooks/pre-push`

**팀 모드 규칙:**
- ✅ **시맨틱 버전 형식만 허용**: `v1.2.3`
- ✅ **main 브랜치에서만 생성 가능**
- ✅ **develop과 main 동기화 확인**
- ✅ **중복 태그 생성 금지**
- ✅ **위반 시 즉시 차단 (exit 1)**

**개인 모드 규칙:**
- ⚠️  **권고 형식**: 시맨틱 버전 권장
- ⚠️  **확인 질문**: 비표준 태그 시 사용자 확인

#### 검증 프로세스
```bash
# 잘못된 태그 생성 시도 (팀 모드)
git checkout feature/SPEC-001
git tag v2.0.0 -m "Emergency release"
git push origin v2.0.0

# 결과: ❌ BLOCKED
# - 현재 브랜치: feature/SPEC-001 (main 아님)
# - 요구 브랜치: main
```

### 2. GitHub Actions 워크플로우 강화

#### 파일 위치: `.github/workflows/moai-release-pipeline.yml`

**추가된 검증:**
- ✅ **태그 푸시 이벤트 감지**
- ✅ **시맨틱 버전 형식 검증**
- ✅ **팀 모드 자동 감지**
- ✅ **소스 브랜치 검증 (main만 허용)**
- ✅ **CI/CD 파이프라인에서 위반 시 빌드 실패**

#### 워크플로우 트리거
```yaml
on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'  # 시맨틱 버전만
```

### 3. GitHub 브랜치 보호 규칙 (권장)

#### GitHub 설정 → Settings → Branches

**권장 보호 규칙:**
1. **main 브랜치 보호**
   - PR 요구 (최소 1명 리뷰어)
   - CI/CD 통과 필요
   - Force push 금지

2. **develop 브랜치 보호**
   - PR 요구 (팀 결정)
   - CI/CD 통과 필요
   - Force push 금지

## 🔄 강화된 GitFlow 워크플로우

### 팀 모드 엄격 순서
```
1. feature/SPEC-XXX 개발
   ↓ (PR 생성)
2. develop 브랜치로 병합
   ↓ (통합 테스트)
3. develop → main 병합
   ↓ (릴리즈 준비)
4. main 브랜치에서 태그 생성
   ↓ (자동 배포)
5. GitHub 릴리즈 자동 트리거
```

### 각 단계별 검증
```bash
# 1단계: feature 개발 (정상)
git checkout -b feature/SPEC-001
# ... 개발 작업 ...

# 2단계: develop 병합 (정상)
git push origin feature/SPEC-001
# PR 생성 → develop으로 병합

# 3단계: main 병합 (정상)
git checkout main
git merge develop
git push origin main

# 4단계: 태그 생성 (정상)
git tag v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3
# → ✅ 성공: main 브랜치에서 생성, 시맨틱 버전
```

## 🚫 차단되는 위반 시나리오

### 시나리오 1: feature 브랜치에서 직접 태그
```bash
git checkout feature/SPEC-001
git tag v1.0.0
git push origin v1.0.0

# 결과: ❌ BLOCKED by pre-push hook
# 오류: "Tags must be created from 'main' branch in team mode"
```

### 시나리오 2: 비표준 태그 형식
```bash
git checkout main
git tag release-v1.0
git push origin release-v1.0

# 결과: ❌ BLOCKED by pre-push hook
# 오류: "Invalid tag format: release-v1.0"
```

### 시나리오 3: 중복 태그
```bash
git checkout main
git tag v1.0.0  # 이미 원격에 존재
git push origin v1.0.0

# 결과: ❌ BLOCKED by pre-push hook & GitHub Actions
# 오류: "Tag 'v1.0.0' already exists remotely"
```

### 시나리오 4: develop 미동기화 상태에서 태그
```bash
# main이 develop보다 뒤처진 상태
git checkout main
git tag v1.0.0
git push origin v1.0.0

# 결과: ⚠️  WARNING (pre-push)
# 메시지: "Main branch is not synchronized with develop"
# 확인 필요: continue anyway? (y/N)
```

## 📋 적용 상태 확인

### 1. pre-push Hook 상태
```bash
# Hook이 적용되었는지 확인
ls -la .git/hooks/pre-push

# Hook 내용 확인 (팀 모드 검증)
grep -A 10 "TEAM_MODE.*true" .git/hooks/pre-push
```

### 2. GitHub Actions 상태
```bash
# 워크플로우에 태그 검증이 있는지 확인
grep -A 5 "validate-tag-push" .github/workflows/moai-release-pipeline.yml
```

### 3. 팀 모드 설정 확인
```bash
# 프로젝트 모드 확인
cat .moai/config.json | jq '.git_strategy.mode'
```

## 🔧 문제 해결

### Hook이 작동하지 않을 때
```bash
# Hook 재설치
chmod +x .git/hooks/pre-push

# 모드 설정 확인
cat .moai/config.json | jq '.git_strategy.mode'
```

### GitHub Actions 실패 시
1. **Settings → Actions** 확인
2. **워크플로우 권한** 확인
3. **jq 패키지 설치** 확인 (Ubuntu 기본 제공)

### 팀 전환 시
```bash
# 팀 모드로 전환
/alfred:0-project update
# 또는 직접 설정 수정
vim .moai/config.json
```

## 🎯 정책 준수 점검리스트

### 정기적 점검 항목
- [ ] pre-push hook이 모든 팀원에게 적용되었는가?
- [ ] GitHub Actions 태그 검증이 정상 작동하는가?
- [ ] 팀 모드 설정이 올바른가?
- [ ] 모든 태그가 시맨틱 버전 형식을 따르는가?
- [ ] 태그가 항상 main 브랜치에서 생성되는가?
- [ ] develop → main 동기화 후 태그가 생성되는가?

## 📞 지원 및 문의

정책 적용 중 문제 발생 시:
1. `.moai/reports/` 에서 로그 확인
2. `moai-adk doctor` 실행
3. GitHub Actions 로그 확인
4. 팀 관리자에게 문의

---

**🔒 이 정책은 MoAI-ADK 팀 모드에서 GitFlow 순서 엄격 준수와 태그 우회 방지를 보장합니다.**