# /alfred:9-update

MoAI-ADK 패키지와 템플릿을 안전하게 업데이트합니다.

## Overview

Package Update는 MoAI-ADK의 핵심 유지보수 워크플로우입니다. **"안전한 업데이트, 완벽한 복원"** 원칙을 따라 자동 백업, 스마트 병합, 무결성 검증을 제공합니다.

### 담당

- **Alfred** (직접 실행)
- **역할**: 버전 확인, 패키지 업데이트, 템플릿 동기화, 품질 검증
- **지원**: debug-helper (오류 발생 시), trust-checker (품질 검증 시)

---

## When to Use

다음과 같은 경우 `/alfred:9-update`를 사용합니다:

- ✅ 새 버전의 MoAI-ADK가 출시되었을 때
- ✅ 버그 수정 패치를 적용할 때
- ✅ 신규 기능을 사용하고 싶을 때
- ✅ 템플릿 파일이 손상되었을 때

### 중요 버전 정보

**✅ v0.2.17 이상 사용자**:

- `/alfred:9-update` 커맨드 사용 (권장)
- 사용자 SPEC/Reports 자동 보호
- 스마트 병합 지원

**📌 v0.2.16 이하 사용자**:

- 터미널에서 `moai init .` 사용 (안전)

```bash
npm install -g moai-adk@latest
cd your-project
moai init .
```

---

## Command Syntax

### Basic Usage

```bash
/alfred:9-update
```

### Advanced Options

```bash
# 업데이트 가능 여부만 확인
/alfred:9-update --check

# 백업 없이 강제 업데이트 (주의)
/alfred:9-update --force

# 품질 검증 포함 업데이트
/alfred:9-update --check-quality
```

---

## Workflow (2단계)

### Phase 1: 분석 및 계획 수립

Alfred가 다음 작업을 수행합니다:

1. **현재 버전 확인**

   ```bash
   # package.json에서 버전 조회
   npm list moai-adk

   # 출력 예시
   # moai-adk@0.2.17
   ```

2. **최신 버전 조회**

   ```bash
   # npm 레지스트리에서 최신 버전 확인
   npm view moai-adk version

   # 출력 예시
   # 0.2.17
   ```

3. **버전 비교 및 분석**
   - 현재: `v0.2.16`
   - 최신: `v0.2.17`
   - 유형: Patch 업데이트 (버그 수정/문서 개선)

4. **업데이트 계획 보고**

   ```markdown
   📦 MoAI-ADK 업데이트 계획

   현재 버전: v0.2.16
   최신 버전: v0.2.17
   업데이트 유형: Patch (완전 호환)

   실행 단계:
   1. 백업 생성 (.moai-backup/2025-10-11-15-30-00/)
   2. npm 패키지 업데이트
   3. 템플릿 파일 동기화
   4. 무결성 검증

   보호 대상:
   - .moai/specs/ (절대 건드리지 않음)
   - .moai/reports/ (절대 건드리지 않음)
   - .moai/project/*.md (사용자 수정 시 보존)
   - CLAUDE.md (지능형 병합)
   - .moai/config.json (스마트 딥 병합)

   진행하시겠습니까? (진행/수정/중단)
   ```

5. **사용자 확인 대기**
   - **"진행"**: Phase 2로 이동
   - **"수정 [내용]"**: 옵션 변경 (예: --force 추가/제거)
   - **"중단"**: 작업 취소

### Phase 2: 업데이트 실행

사용자가 "진행"하면 Alfred가 다음 5단계를 순차 실행합니다:

---

#### Step 1: 백업 생성

**목적**: 롤백 가능하도록 현재 상태 저장

```bash
# 백업 디렉토리 생성
mkdir -p .moai-backup/2025-10-11-15-30-00

# 파일 백업
cp -r .claude/ .moai-backup/2025-10-11-15-30-00/
cp -r .moai/ .moai-backup/2025-10-11-15-30-00/
cp CLAUDE.md .moai-backup/2025-10-11-15-30-00/

# 백업 확인
ls -la .moai-backup/2025-10-11-15-30-00/
```

**출력 예시**:

```text
💾 백업 생성 중...
   → .moai-backup/2025-10-11-15-30-00/
   ✅ .claude/ (35개 파일)
   ✅ .moai/ (8개 파일)
   ✅ CLAUDE.md

✅ 백업 완료
```

**예외 처리**:

- `--force` 옵션 시: 백업 건너뛰기 (⚠️ 롤백 불가)
- 디스크 공간 부족: 오류 메시지 + 작업 중단

---

#### Step 2: npm 패키지 업데이트

**목적**: moai-adk 패키지를 최신 버전으로 업그레이드

1. **패키지 매니저 자동 감지**

   ```bash
   # lock 파일 확인
   ls -la | grep -E "pnpm-lock|yarn.lock|bun.lockb|package-lock"

   # 감지 결과
   # pnpm-lock.yaml 존재 → pnpm 사용
   # yarn.lock 존재 → yarn 사용
   # bun.lockb 존재 → bun 사용
   # package-lock.json 존재 → npm 사용 (기본)
   ```

2. **설치 위치 판단**

   ```bash
   # package.json 확인
   if [ -f "package.json" ]; then
     # 로컬 설치
     pnpm install moai-adk@latest
   else
     # 글로벌 설치
     pnpm install -g moai-adk@latest
   fi
   ```

3. **패키지 업데이트 실행**

   ```bash
   # 예시: pnpm 로컬 설치
   pnpm install moai-adk@latest

   # 설치 확인
   pnpm list moai-adk
   # moai-adk@0.2.17
   ```

**출력 예시**:

```text
📦 패키지 업데이트 중...
   감지된 패키지 매니저: pnpm
   설치 위치: 로컬

   실행: pnpm install moai-adk@latest

   ✅ moai-adk@0.2.17 설치 완료
```

**오류 처리**:

- npm 캐시 손상 → `npm cache clean --force` 후 재시도
- 네트워크 오류 → 재시도 안내 (최대 3회)
- 버전 불일치 → Phase 2 재실행 제안

---

#### Step 3: 템플릿 파일 동기화

**목적**: 최신 템플릿으로 프로젝트 파일 업데이트 (사용자 데이터 보호)

##### 3.1 npm root 확인

```bash
# node_modules 경로 확인
npm root

# 출력 예시
# /Users/goos/MoAI/MoAI-ADK/node_modules

# 템플릿 경로 설정
TEMPLATE_ROOT=/Users/goos/MoAI/MoAI-ADK/node_modules/moai-adk/templates
```

##### 3.2 사용자 데이터 보호 (🔒 v0.2.17+)

**절대 건드리지 않는 디렉토리**:

```bash
# .moai/specs/ - 사용자 SPEC 파일
# ❌ 읽기 금지, ❌ 수정 금지, ❌ 삭제 금지, ❌ 덮어쓰기 금지

# .moai/reports/ - 동기화 리포트
# ❌ 읽기 금지, ❌ 수정 금지, ❌ 삭제 금지, ❌ 덮어쓰기 금지

# 템플릿 복사 시 자동 제외
excludePaths: ['specs', 'reports']
```

**보존/병합 대상**:

- `.moai/config.json` - 스마트 딥 병합 🔄
- `.moai/project/*.md` - 사용자 수정 시 보존 🔒
- `CLAUDE.md` - 지능형 병합 🔄

##### 3.3 템플릿 복사 (병렬 최적화 ⚡)

**카테고리 A-D: 시스템 파일 전체 교체 (병렬 실행)**

```bash
# Step 1: 명령어 파일 (~10개)
.claude/commands/alfred/*.md → 전체 교체 ✅

# Step 2: 에이전트 파일 (~9개)
.claude/agents/alfred/*.md → 전체 교체 ✅

# Step 3: 훅 파일 + 권한 (~4개)
.claude/hooks/alfred/*.cjs → 전체 교체 + chmod 755 ✅

# Step 4: Output Styles (4개)
.claude/output-styles/alfred/*.md → 전체 교체 ✅
```

**병렬 실행 효과**:

- 4단계 → 1단계로 단축
- 약 75% 시간 절감 (10-12초 → 3-4초)

**카테고리 E: 개발 가이드 (순차 실행)**

```bash
# Step 5: 개발 가이드
.moai/memory/development-guide.md → 무조건 덮어쓰기 ✅
```

**카테고리 F-H: 프로젝트 문서 (사용자 작업물 보존)**

```bash
# Step 6: product.md
if [[ $(grep "{{PROJECT_NAME}}" .moai/project/product.md) ]]; then
  # 템플릿 상태 → 덮어쓰기
  cp $TEMPLATE_ROOT/.moai/project/product.md .moai/project/product.md
else
  # 사용자 수정 → 보존 🔒
  echo "ℹ️ product.md는 이미 작성되어 있어서 건너뜁니다"
  echo "  → 최신 템플릿: $TEMPLATE_ROOT/.moai/project/product.md"
fi

# Step 7-8: structure.md, tech.md도 동일
```

**카테고리 I: CLAUDE.md 지능형 병합**

```bash
# Step 9: CLAUDE.md 병합
if [[ ! $(grep "{{PROJECT_NAME}}" CLAUDE.md) ]]; then
  # 사용자 프로젝트 정보 추출
  PROJECT_NAME=$(grep "^- \*\*이름\*\*:" CLAUDE.md | sed 's/.*: //')
  PROJECT_DESC=$(grep "^- \*\*설명\*\*:" CLAUDE.md | sed 's/.*: //')
  PROJECT_VERSION=$(grep "^- \*\*버전\*\*:" CLAUDE.md | sed 's/.*: //')
  PROJECT_MODE=$(grep "^- \*\*모드\*\*:" CLAUDE.md | sed 's/.*: //')

  # 최신 템플릿 읽기
  TEMPLATE_CLAUDE=$(cat $TEMPLATE_ROOT/CLAUDE.md)

  # 템플릿에 사용자 정보 주입
  echo "$TEMPLATE_CLAUDE" | \
    sed "s/{{PROJECT_NAME}}/$PROJECT_NAME/" | \
    sed "s/{{PROJECT_DESCRIPTION}}/$PROJECT_DESC/" | \
    sed "s/{{PROJECT_VERSION}}/$PROJECT_VERSION/" | \
    sed "s/{{PROJECT_MODE}}/$PROJECT_MODE/" > CLAUDE.md

  echo "🔄 CLAUDE.md 병합 완료 (템플릿 최신화 + 프로젝트 정보 유지)"
fi
```

**카테고리 J: config.json 스마트 딥 병합**

```bash
# Step 10: config.json 병합
if [[ ! $(grep "{{PROJECT_NAME}}" .moai/config.json) ]]; then
  # 사용자 설정 추출 (JSON 파싱)
  USER_CONFIG=$(cat .moai/config.json)
  TEMPLATE_CONFIG=$(cat $TEMPLATE_ROOT/.moai/config.json)

  # 딥 병합 전략 (필드별)
  # - project.*: 사용자 값 100% 유지
  # - constitution.*: 템플릿 필드 + 사용자 값 덮어쓰기
  # - git_strategy.*: 사용자 값 100% 유지
  # - tags.categories: 템플릿 + 사용자 병합 (중복 제거)
  # - pipeline.available_commands: 템플릿 최신
  # - pipeline.current_stage: 사용자 값 유지
  # - _meta.*: 템플릿 최신

  # 병합 결과 저장 (들여쓰기 2칸)
  echo "$MERGED_CONFIG" | jq '.' --indent 2 > .moai/config.json

  echo "🔄 config.json 스마트 병합 완료"
fi
```

**출력 예시**:

```text
📄 템플릿 동기화 중...

⚡ 병렬 처리 (Step 1-4):
   ✅ .claude/commands/alfred/ (10개 파일)
   ✅ .claude/agents/alfred/ (9개 파일)
   ✅ .claude/hooks/alfred/ (4개 파일 + 권한)
   ✅ .claude/output-styles/alfred/ (4개 파일)

📚 순차 처리 (Step 5-10):
   ✅ .moai/memory/development-guide.md (무조건 업데이트)
   ✅ .moai/project/product.md (템플릿 → 최신)
   ⏭️  .moai/project/structure.md (사용자 작업물 보존 🔒)
   ✨ .moai/project/tech.md (새로 생성)
   🔄 CLAUDE.md (지능형 병합)
   🔄 .moai/config.json (스마트 병합)

✅ 템플릿 동기화 완료
```

**오류 처리**:

- 파일 복사 실패 → 실패 목록 수집 + 재시도 제안
- 디렉토리 없음 → `mkdir -p` 자동 실행 후 재시도
- JSON 파싱 실패 → 백업 생성 + 템플릿 교체 (안전 모드)

---

#### Step 4: 무결성 검증

**목적**: 업데이트가 올바르게 완료되었는지 검증

##### 4.1 파일 개수 검증

```bash
# 명령어 파일 개수
COMMANDS_COUNT=$(ls .claude/commands/alfred/*.md | wc -l)
echo "명령어 파일: $COMMANDS_COUNT개 (예상: ~10개)"

# 에이전트 파일 개수
AGENTS_COUNT=$(ls .claude/agents/alfred/*.md | wc -l)
echo "에이전트 파일: $AGENTS_COUNT개 (예상: ~9개)"

# 훅 파일 개수
HOOKS_COUNT=$(ls .claude/hooks/alfred/*.cjs | wc -l)
echo "훅 파일: $HOOKS_COUNT개 (예상: 4개)"

# Output Styles 파일 개수
STYLES_COUNT=$(ls .claude/output-styles/alfred/*.md | wc -l)
echo "Output Styles: $STYLES_COUNT개 (예상: 4개)"
```

##### 4.2 YAML Frontmatter 검증

```bash
# 대표 파일 샘플링
SAMPLE_FILE=".claude/commands/alfred/1-spec.md"

# Frontmatter 추출 및 파싱
head -10 $SAMPLE_FILE | grep -E "^(name|description|allowed-tools):" > /dev/null

if [ $? -eq 0 ]; then
  echo "✅ YAML frontmatter 정상"
else
  echo "❌ YAML frontmatter 손상 감지"
fi
```

##### 4.3 버전 정보 확인

```bash
# development-guide.md 버전
GUIDE_VERSION=$(grep "version:" .moai/memory/development-guide.md | head -1 | awk '{print $2}')

# package.json 버전
PKG_VERSION=$(npm list moai-adk --depth=0 | grep moai-adk | awk '{print $2}' | sed 's/@//')

if [ "$GUIDE_VERSION" == "$PKG_VERSION" ]; then
  echo "✅ 버전 일치: $PKG_VERSION"
else
  echo "❌ 버전 불일치: guide=$GUIDE_VERSION, pkg=$PKG_VERSION"
fi
```

##### 4.4 훅 파일 권한 검증 (Unix만)

```bash
# Unix 계열에서만 실행
if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "win32" ]]; then
  for hook in .claude/hooks/alfred/*.cjs; do
    if [[ -x "$hook" ]]; then
      echo "✅ $hook (755)"
    else
      echo "⚠️ $hook (실행 권한 없음)"
    fi
  done
fi
```

**출력 예시**:

```text
🔍 무결성 검증 중...

파일 개수:
   ✅ 명령어 파일: 10개 (예상: ~10개)
   ✅ 에이전트 파일: 9개 (예상: ~9개)
   ✅ 훅 파일: 4개 (예상: 4개)
   ✅ Output Styles: 4개 (예상: 4개)

내용 검증:
   ✅ YAML frontmatter 정상
   ✅ 버전 일치: v0.2.17

권한 검증:
   ✅ .claude/hooks/alfred/pre-commit.cjs (755)
   ✅ .claude/hooks/alfred/post-commit.cjs (755)
   ✅ .claude/hooks/alfred/pre-push.cjs (755)
   ✅ .claude/hooks/alfred/post-merge.cjs (755)

✅ 검증 완료
```

**검증 실패 시 복구 전략**:

| 오류 유형 | 복구 조치 |
|----------|---------|
| 파일 누락 | Step 3 재실행 (템플릿 복사) |
| 버전 불일치 | Step 2 재실행 (npm 재설치) |
| 내용 손상 | 백업 복원 후 재시작 |
| 권한 오류 | `chmod +x .claude/hooks/alfred/*.cjs` |

---

#### Step 5: 품질 검증 (선택적)

**조건**: `--check-quality` 옵션 제공 시에만 실행

**목적**: TRUST 5원칙 준수 여부 검증

```bash
# trust-checker 에이전트 호출
@agent-trust-checker "Level 1 빠른 스캔 (3-5초)"
```

**검증 항목**:

1. **T**est Coverage

   ```bash
   # 테스트 커버리지 확인
   npm run test:coverage

   # 기준: 85% 이상
   Coverage: 87.5% ✅
   ```

2. **R**eadable Code

   ```bash
   # 린터 검사
   npm run lint

   # 0 errors, 0 warnings ✅
   ```

3. **U**nified Architecture

   ```bash
   # TypeScript 타입 검사
   npm run type-check

   # No type errors ✅
   ```

4. **S**ecured

   ```bash
   # 보안 취약점 검사
   npm audit

   # 0 vulnerabilities ✅
   ```

5. **T**rackable

   ```bash
   # @TAG 체인 무결성
   rg '@(SPEC|TEST|CODE|DOC):' -n

   # 고아 TAG: 0개 ✅
   ```

**결과별 처리**:

- **✅ Pass (통과)**:

  ```text
  ✅ 품질 검증 통과
     - 테스트 커버리지: 87.5% (목표: 85%)
     - 린터 오류: 0개
     - 타입 오류: 0개
     - 보안 취약점: 0개
     - TAG 체인: 정상

  ✨ 업데이트 완료!
  ```

- **⚠️ Warning (경고)**:

  ```text
  ⚠️ 품질 검증 경고
     - 테스트 커버리지: 82% (목표: 85%) ⚠️
     - 일부 문서 포맷 이슈 발견

  계속 진행 가능하지만 확인 권장
  ```

- **❌ Critical (치명적)**:

  ```text
  ❌ 품질 검증 실패
     - 파일 손상 감지
     - 설정 불일치

  권장 조치:
  1. 롤백 (권장): 백업에서 수동 복원
     cp -r .moai-backup/2025-10-11-15-30-00/* ./
  2. 무시하고 진행 (위험)

  선택: ___
  ```

**실행 시간**: 추가 3-5초

---

## Update Modes

### 일반 업데이트 (기본)

```bash
/alfred:9-update
```

**실행 흐름**:

1. ✅ 버전 확인
2. ✅ 백업 생성
3. ✅ npm 업데이트
4. ✅ 템플릿 동기화
5. ✅ 무결성 검증
6. ⏭️ 품질 검증 건너뜀

### --check (확인만)

```bash
/alfred:9-update --check
```

**실행 흐름**:

1. ✅ 버전 확인
2. ⏭️ 업데이트 중단
3. 📊 결과 보고

**출력 예시**:

```text
🔍 업데이트 확인 중...

현재 버전: v0.2.16
최신 버전: v0.2.17
업데이트 유형: Patch (완전 호환)

✅ 업데이트 가능

업데이트하려면: /alfred:9-update
```

### --force (강제 업데이트)

```bash
/alfred:9-update --force
```

**⚠️ 경고**: 백업 생성 없이 업데이트 (롤백 불가)

**실행 흐름**:

1. ✅ 버전 확인
2. ⏭️ 백업 건너뜀 (⚠️)
3. ✅ npm 업데이트
4. ✅ 템플릿 동기화
5. ✅ 무결성 검증

**사용 시나리오**:

- 디스크 공간 부족
- 이미 수동 백업 완료
- 테스트 환경

### --check-quality (품질 검증)

```bash
/alfred:9-update --check-quality
```

**실행 흐름**:

1. ✅ 버전 확인
2. ✅ 백업 생성
3. ✅ npm 업데이트
4. ✅ 템플릿 동기화
5. ✅ 무결성 검증
6. ✅ 품질 검증 (trust-checker)

**추가 시간**: +3-5초

---

## Version Compatibility

### Semantic Versioning (`v0.x.y`)

| 업데이트 유형 | 호환성 | 설명 | 예시 |
|------------|-------|------|------|
| **Patch** | 완전 호환 | 버그 수정, 문서 개선 | v0.2.16 → v0.2.17 |
| **Minor** | 호환 (설정 확인 권장) | 신규 기능 추가 | v0.2.x → v0.3.x |
| **Major** | Breaking Changes | 마이그레이션 가이드 필수 | v0.x.x → v1.x.x |

### Upgrade Paths

#### Patch (0.2.16 → 0.2.17)

```bash
# 안전한 업데이트
/alfred:9-update

# 예상 소요 시간: 30-60초
# 위험도: 낮음 ✅
# 롤백 가능: ✅
```

#### Minor (0.2.x → 0.3.x)

```bash
# 설정 확인 권장
/alfred:9-update --check-quality

# 예상 소요 시간: 45-90초
# 위험도: 중간 ⚠️
# 롤백 가능: ✅
# 추가 작업: config.json 신규 필드 확인
```

#### Major (0.x.x → 1.x.x)

```bash
# 마이그레이션 가이드 필수 확인
# docs/migration/v0-to-v1.md

# 1. 변경사항 읽기
# 2. 수동 백업
# 3. 업데이트 실행
/alfred:9-update --check-quality

# 예상 소요 시간: 2-5분
# 위험도: 높음 ❌
# 롤백 가능: ✅ (수동 백업 필수)
# 추가 작업: 코드 수정 필요할 수 있음
```

---

## Backup & Restore

### 자동 백업 구조

```text
.moai-backup/
├── 2025-10-11-15-30-00/     # 타임스탬프 기반
│   ├── .claude/
│   │   ├── commands/
│   │   ├── agents/
│   │   ├── hooks/
│   │   └── output-styles/
│   ├── .moai/
│   │   ├── config.json
│   │   ├── memory/
│   │   └── project/
│   └── CLAUDE.md
├── 2025-10-10-09-15-22/     # 이전 백업
└── 2025-10-09-14-45-33/     # 더 오래된 백업
```

### 백업 보존 정책

- **자동 삭제 없음**: 수동 삭제 전까지 영구 보존
- **권장 보존 기간**: 최근 3-5개
- **수동 정리**: `rm -rf .moai-backup/2025-10-09-*`

### 복원 방법

#### 1. 백업 목록 확인

```bash
# 터미널에서 실행
ls -la .moai-backup/

# 출력 예시
# drwxr-xr-x  2025-10-11-15-30-00
# drwxr-xr-x  2025-10-10-09-15-22
# drwxr-xr-x  2025-10-09-14-45-33
```

#### 2. 수동 복원

```bash
# 현재 파일 백업 (안전)
cp -r .claude/ .claude.current/
cp -r .moai/ .moai.current/
cp CLAUDE.md CLAUDE.current.md

# 백업에서 복원
cp -r .moai-backup/2025-10-11-15-30-00/.claude/ ./
cp -r .moai-backup/2025-10-11-15-30-00/.moai/ ./
cp .moai-backup/2025-10-11-15-30-00/CLAUDE.md ./

echo "✅ 수동 복원 완료"
```

#### 3. 선택적 복원

```bash
# 특정 파일만 복원
cp .moai-backup/2025-10-11-15-30-00/.moai/config.json .moai/
echo "✅ config.json만 복원 완료"

# 특정 디렉토리만 복원
cp -r .moai-backup/2025-10-11-15-30-00/.claude/commands/ .claude/
echo "✅ 명령어 파일만 복원 완료"
```

---

## Troubleshooting Scenarios

### 시나리오 1: 파일 복사 실패

**상황**:

```text
Phase 3 템플릿 동기화 중...
  → .claude/commands/alfred/1-spec.md ✅
  → .claude/commands/alfred/2-build.md ❌ (디스크 공간 부족)
```

**복구 절차**:

1. **오류 로그 기록**

   ```text
   ❌ 2-build.md 복사 실패: 디스크 공간 부족
   ```

2. **실패 목록 수집**

   ```text
   failed_files = [2-build.md, 3-sync.md]
   ```

3. **나머지 파일 계속 처리**

   ```text
   → .claude/commands/alfred/0-project.md ✅
   → .claude/commands/alfred/9-update.md ✅
   ```

4. **Phase 3 종료 후 보고**

   ```text
   ⚠️ 2개 파일 복사 실패
      - 2-build.md
      - 3-sync.md

   선택:
   1. 재시도 (실패한 파일만)
   2. 백업 복원 (cp -r .moai-backup/2025-10-11-15-30-00/* ./)
   3. 무시하고 진행 (권장하지 않음)
   ```

5. **자동 재시도**
   - 각 파일당 최대 2회
   - 재시도 간격: 3초

**해결 방법**:

```bash
# 디스크 공간 확보
df -h  # 공간 확인
rm -rf ~/Downloads/large-file.zip  # 불필요한 파일 삭제

# Step 3 재실행 (실패한 파일만)
# Alfred가 자동으로 재시도
```

---

### 시나리오 2: 검증 실패 (파일 누락)

**상황**:

```text
Phase 4 검증 중...
  → 명령어 파일: 8개 (예상: 10개) ❌
  → 누락 파일 감지
```

**복구 절차**:

1. **누락 파일 파악**

   ```bash
   # 템플릿 파일 목록
   ls $TEMPLATE_ROOT/.claude/commands/alfred/*.md

   # 실제 파일 목록
   ls .claude/commands/alfred/*.md

   # 차이 비교
   diff <(ls $TEMPLATE_ROOT/.claude/commands/alfred/*.md) \
        <(ls .claude/commands/alfred/*.md)

   # 누락된 파일
   # 3-sync.md
   # 0-project.md
   ```

2. **사용자 선택 제안**

   ```text
   ❌ 검증 실패: 2개 파일 누락
      - 3-sync.md
      - 0-project.md

   선택:
   1. Phase 3 재실행 (전체 복사 재시도)
   2. 백업 복원 (cp -r .moai-backup/2025-10-11-15-30-00/* ./)
   3. 무시하고 진행 (위험)
   ```

3. **"Phase 3 재실행" 선택 시**

   ```bash
   # Alfred가 자동 재실행
   → Phase 3 템플릿 동기화 재시작
   → 완료 후 Phase 4 재검증

   if [ 재검증 통과 ]; then
     echo "✅ 검증 통과 (재시도 성공)"
   else
     # 시나리오 2 반복 (최대 3회)
   fi
   ```

4. **"백업 복원" 선택 시**

   ```bash
   # 수동 복원
   cp -r .moai-backup/2025-10-11-15-30-00/.claude/ ./
   cp -r .moai-backup/2025-10-11-15-30-00/.moai/ ./
   cp .moai-backup/2025-10-11-15-30-00/CLAUDE.md ./

   echo "✅ 복원 완료"
   echo "재시도하시겠습니까? (y/n)"
   ```

---

### 시나리오 3: 버전 불일치

**상황**:

```text
Phase 4 검증 중...
  → development-guide.md 버전: v0.2.17
  → npm 패키지 버전: v0.2.18
  → ❌ 버전 불일치
```

**복구 절차**:

1. **사용자에게 보고**

   ```text
   ⚠️ 버전 불일치 감지
      development-guide.md: v0.2.17
      moai-adk 패키지: v0.2.18

   → 템플릿 복사가 제대로 되지 않았습니다
   ```

2. **원인 분석 안내**

   ```text
   가능한 원인:
   a. npm 캐시 손상
   b. 템플릿 디렉토리 경로 오류
   c. 파일 복사 중 네트워크 오류
   ```

3. **선택 제안**

   ```text
   선택:
   1. Phase 2 재실행 (npm 재설치)
   2. Phase 3 재실행 (템플릿 재복사)
   3. 무시 (권장하지 않음)
   ```

4. **자동 복구 (Phase 2 선택 시)**

   ```bash
   # npm 캐시 정리
   npm cache clean --force

   # 패키지 재설치
   npm install moai-adk@latest

   # Phase 3 재실행
   # Alfred가 템플릿 재복사

   # Phase 4 재검증
   ```

---

### 시나리오 4: Write 도구 실패 (디렉토리 없음)

**상황**:

```text
[Write] .claude/commands/alfred/1-spec.md
❌ 실패: 디렉토리 없음
```

**자동 복구**:

1. **오류 감지**

   ```text
   ❌ Write 실패: .claude/commands/alfred/ 디렉토리 없음
   ```

2. **디렉토리 자동 생성**

   ```bash
   mkdir -p .claude/commands/alfred
   echo "✅ 디렉토리 생성 완료"
   ```

3. **Write 재시도**

   ```bash
   # Alfred가 자동 재시도
   [Write] .claude/commands/alfred/1-spec.md
   ✅ 파일 복사 성공 (재시도)
   ```

**재시도 실패 시**:

```text
❌ Write 재시도 실패

확인 사항:
1. 디스크 공간: df -h
2. 권한 문제: ls -la .claude/
3. 파일 잠금: lsof .claude/commands/alfred/1-spec.md

→ 시나리오 1로 진행 (파일 복사 실패)
```

---

### 시나리오 5: config.json 병합 충돌

**상황**:

```text
Phase 3-10 config.json 병합 중...
❌ JSON 파싱 실패: Unexpected token at line 15
```

**복구 절차**:

1. **오류 감지 및 보고**

   ```text
   ❌ config.json 병합 실패
      → JSON 파싱 오류: line 15
      → 사용자 설정 파일이 손상되었을 수 있습니다
   ```

2. **안전 모드 진입**

   ```bash
   # 손상된 config.json 백업
   cp .moai/config.json .moai/config.json.broken

   # 템플릿으로 교체
   cp $TEMPLATE_ROOT/.moai/config.json .moai/config.json

   echo "⚠️ 안전 모드: 템플릿으로 교체 완료"
   echo "   손상된 파일: .moai/config.json.broken"
   ```

3. **사용자 안내**

   ```text
   ℹ️ 다음 수동 작업 권장:

   1. 손상된 파일 확인:
      cat .moai/config.json.broken

   2. 복구 가능한 설정 수동 이전:
      - project.name, version, mode
      - git_strategy.mode, base_branch

   3. 복구 완료 후 삭제:
      rm .moai/config.json.broken
   ```

---

### 시나리오 6: 품질 검증 실패 (Critical)

**상황**:

```text
Phase 5 품질 검증 중...
❌ Critical: 파일 손상 감지
   - CLAUDE.md: 필수 섹션 누락
   - config.json: 필수 필드 누락
```

**복구 절차**:

1. **즉시 중단**

   ```text
   ❌ 품질 검증 실패 (Critical)

   발견된 문제:
   - CLAUDE.md: "## @TAG Lifecycle" 섹션 누락
   - config.json: "project.name" 필드 누락

   → 업데이트가 올바르게 완료되지 않았습니다
   ```

2. **복구 옵션 제시**

   ```text
   권장 조치:

   1. 롤백 (강력 권장) ✅
      cp -r .moai-backup/2025-10-11-15-30-00/.claude/ ./
      cp -r .moai-backup/2025-10-11-15-30-00/.moai/ ./
      cp .moai-backup/2025-10-11-15-30-00/CLAUDE.md ./

   2. 수동 수정 (고급 사용자) ⚠️
      - CLAUDE.md 섹션 추가
      - config.json 필드 추가

   3. 무시하고 진행 (매우 위험) ❌
      → 시스템 불안정 가능성

   선택 (1/2/3): ___
   ```

3. **"롤백" 선택 시**

   ```bash
   # 수동 롤백 실행
   cp -r .moai-backup/2025-10-11-15-30-00/.claude/ ./
   cp -r .moai-backup/2025-10-11-15-30-00/.moai/ ./
   cp .moai-backup/2025-10-11-15-30-00/CLAUDE.md ./

   echo "✅ 롤백 완료"
   echo ""
   echo "재시도 방법:"
   echo "1. 네트워크 확인"
   echo "2. npm 캐시 정리: npm cache clean --force"
   echo "3. 재실행: /alfred:9-update"
   ```

---

## Data Protection Strategy

### 완전 보존 (Never Touch) 🔒

#### .moai/specs/ - 사용자 SPEC 파일

**보호 수준**: 최상 (절대 건드리지 않음)

```bash
# ❌ 금지 작업
- 읽기 금지
- 수정 금지
- 삭제 금지
- 덮어쓰기 금지
- 백업 생성 금지 (불필요)

# ✅ 보호 메커니즘
# moai init . 실행 시
excludePaths: ['specs']

# /alfred:9-update 실행 시
if [[ "$path" == *"/specs/"* ]]; then
  echo "⏭️ .moai/specs/ 건너뜀 (사용자 데이터 보호)"
  continue
fi
```

#### .moai/reports/ - 동기화 리포트

**보호 수준**: 최상 (절대 건드리지 않음)

```bash
# ❌ 금지 작업
- 읽기 금지
- 수정 금지
- 삭제 금지
- 덮어쓰기 금지

# ✅ 보호 메커니즘
# moai init . 실행 시
excludePaths: ['reports']

# /alfred:9-update 실행 시
if [[ "$path" == *"/reports/"* ]]; then
  echo "⏭️ .moai/reports/ 건너뜀 (작업 이력 보호)"
  continue
fi
```

### 조건부 보존 🔒

#### .moai/project/*.md - 프로젝트 문서

**보존 조건**: `{{PROJECT_NAME}}` 패턴 없음

```bash
# 템플릿 판단 로직
if grep -q "{{PROJECT_NAME}}" .moai/project/product.md; then
  # 템플릿 상태 → 덮어쓰기
  cp $TEMPLATE_ROOT/.moai/project/product.md .moai/project/product.md
  echo "✅ product.md 최신 템플릿으로 업데이트"
else
  # 사용자 수정 → 보존 🔒
  echo "ℹ️ product.md는 이미 작성되어 있어서 건너뜁니다"
  echo "  → 최신 템플릿: $TEMPLATE_ROOT/.moai/project/product.md"
fi
```

**보존되는 파일**:

- `product.md` - 제품 개요, 목표, 사용자
- `structure.md` - 디렉토리 구조, 아키텍처
- `tech.md` - 기술 스택, 도구, 라이브러리

### 지능형 병합 🔄

#### CLAUDE.md - 프로젝트 지침

**병합 전략**: 템플릿 최신 구조 + 프로젝트 정보 유지

```bash
# Phase 3-9: CLAUDE.md 병합

# 1. 템플릿 상태 확인
if grep -q "{{PROJECT_NAME}}" CLAUDE.md; then
  # 템플릿 상태 → 전체 교체
  cp $TEMPLATE_ROOT/CLAUDE.md ./
  echo "✅ CLAUDE.md 최신 버전으로 업데이트"
  exit 0
fi

# 2. 사용자 프로젝트 정보 추출
PROJECT_NAME=$(grep "^- \*\*이름\*\*:" CLAUDE.md | sed 's/.*: //')
PROJECT_DESC=$(grep "^- \*\*설명\*\*:" CLAUDE.md | sed 's/.*: //')
PROJECT_VERSION=$(grep "^- \*\*버전\*\*:" CLAUDE.md | sed 's/.*: //')
PROJECT_MODE=$(grep "^- \*\*모드\*\*:" CLAUDE.md | sed 's/.*: //')

echo "📊 추출된 프로젝트 정보:"
echo "   이름: $PROJECT_NAME"
echo "   설명: $PROJECT_DESC"
echo "   버전: $PROJECT_VERSION"
echo "   모드: $PROJECT_MODE"

# 3. 최신 템플릿 읽기
TEMPLATE_CLAUDE=$(cat $TEMPLATE_ROOT/CLAUDE.md)

# 4. 템플릿에 사용자 정보 주입
echo "$TEMPLATE_CLAUDE" | \
  sed "s/{{PROJECT_NAME}}/$PROJECT_NAME/" | \
  sed "s/{{PROJECT_DESCRIPTION}}/$PROJECT_DESC/" | \
  sed "s/{{PROJECT_VERSION}}/$PROJECT_VERSION/" | \
  sed "s/{{PROJECT_MODE}}/$PROJECT_MODE/" > CLAUDE.md

echo "🔄 CLAUDE.md 지능형 병합 완료"
echo "   → 템플릿 최신 구조 반영"
echo "   → 프로젝트 정보 유지 (이름, 설명, 버전, 모드)"
```

**유지되는 정보**:

- 프로젝트 이름
- 프로젝트 설명
- 버전 정보
- 모드 (team/personal)

**최신화되는 내용**:

- Alfred 에이전트 목록
- 워크플로우 설명
- TRUST 5원칙
- @TAG 시스템
- 개발 원칙

### 스마트 딥 병합 🔄

#### .moai/config.json - 프로젝트 설정

**병합 전략**: 필드별 지능적 병합

```bash
# Phase 3-10: config.json 병합

# 1. 템플릿 상태 확인
if grep -q "{{PROJECT_NAME}}" .moai/config.json; then
  # 템플릿 상태 → 전체 교체
  cp $TEMPLATE_ROOT/.moai/config.json .moai/config.json
  echo "✅ config.json 최신 버전으로 업데이트"
  exit 0
fi

# 2. 사용자 설정 추출 (JSON 파싱)
USER_CONFIG=$(cat .moai/config.json | jq '.')
TEMPLATE_CONFIG=$(cat $TEMPLATE_ROOT/.moai/config.json | jq '.')

# 3. 필드별 병합 전략 적용
# project.*: 사용자 값 100% 유지
PROJECT_DATA=$(echo "$USER_CONFIG" | jq '.project')

# constitution.*: 템플릿 필드 + 사용자 값 덮어쓰기
CONST_TEMPLATE=$(echo "$TEMPLATE_CONFIG" | jq '.constitution')
CONST_USER=$(echo "$USER_CONFIG" | jq '.constitution')
CONSTITUTION=$(echo "$CONST_TEMPLATE" | jq ". * $CONST_USER")

# git_strategy.*: 사용자 값 100% 유지
GIT_STRATEGY=$(echo "$USER_CONFIG" | jq '.git_strategy')

# tags.categories: 템플릿 + 사용자 병합 (중복 제거)
TAGS_TEMPLATE=$(echo "$TEMPLATE_CONFIG" | jq '.tags.categories')
TAGS_USER=$(echo "$USER_CONFIG" | jq '.tags.categories')
TAGS_MERGED=$(echo "[$TAGS_TEMPLATE, $TAGS_USER]" | jq 'flatten | unique')

# pipeline.available_commands: 템플릿 최신
PIPELINE_COMMANDS=$(echo "$TEMPLATE_CONFIG" | jq '.pipeline.available_commands')

# pipeline.current_stage: 사용자 값 유지
PIPELINE_STAGE=$(echo "$USER_CONFIG" | jq '.pipeline.current_stage')

# _meta.*: 템플릿 최신
META=$(echo "$TEMPLATE_CONFIG" | jq '._meta')

# 4. 병합 결과 조합
MERGED_CONFIG=$(jq -n \
  --argjson project "$PROJECT_DATA" \
  --argjson constitution "$CONSTITUTION" \
  --argjson git_strategy "$GIT_STRATEGY" \
  --argjson tags_categories "$TAGS_MERGED" \
  --argjson pipeline_commands "$PIPELINE_COMMANDS" \
  --argjson pipeline_stage "$PIPELINE_STAGE" \
  --argjson meta "$META" \
  '{
    project: $project,
    constitution: $constitution,
    git_strategy: $git_strategy,
    tags: { categories: $tags_categories },
    pipeline: {
      available_commands: $pipeline_commands,
      current_stage: $pipeline_stage
    },
    _meta: $meta
  }')

# 5. 병합 결과 저장 (들여쓰기 2칸)
echo "$MERGED_CONFIG" | jq '.' --indent 2 > .moai/config.json

echo "🔄 config.json 스마트 딥 병합 완료"
```

**필드별 병합 정책**:

| 필드 | 병합 전략 | 이유 |
|------|----------|------|
| `project.*` | 사용자 값 100% | 프로젝트 식별 정보 |
| `constitution.test_coverage_target` | 사용자 값 | 팀 정책 |
| `constitution.simplicity_threshold` | 사용자 값 | 팀 정책 |
| `git_strategy.*` | 사용자 값 100% | 워크플로우 설정 |
| `tags.categories` | 병합 (템플릿 + 사용자) | 확장 가능 |
| `pipeline.available_commands` | 템플릿 최신 | 시스템 명령어 |
| `pipeline.current_stage` | 사용자 값 | 진행 상태 |
| `_meta.*` | 템플릿 최신 | TAG 참조 |

---

## Best Practices

### 1. 정기 업데이트

```bash
# 주 1회 확인 권장
/alfred:9-update --check

# 출력: "✅ 최신 버전 사용 중" 또는 "✅ 업데이트 가능"
```

### 2. 중요 작업 전 백업

```bash
# 수동 백업 생성
mkdir -p .moai-backup/manual-$(date +%Y-%m-%d-%H-%M-%S)
cp -r .claude/ .moai/ CLAUDE.md .moai-backup/manual-*/

# 업데이트 실행
/alfred:9-update
```

### 3. 품질 검증 활용

```bash
# 프로덕션 환경 업데이트 시
/alfred:9-update --check-quality

# 개발 환경은 기본 모드
/alfred:9-update
```

### 4. 백업 정리

```bash
# 오래된 백업 삭제 (30일 이상)
find .moai-backup/ -type d -mtime +30 -exec rm -rf {} \;

# 또는 수동 선택
ls -la .moai-backup/
rm -rf .moai-backup/2025-09-*
```

### 5. 변경사항 확인

```bash
# 업데이트 후 변경사항 확인
git diff

# 주요 파일 검토
git diff .moai/config.json
git diff CLAUDE.md
```

### 6. 로그 기록

```bash
# 업데이트 이력 기록
echo "$(date): Updated to v0.2.17" >> .moai/update.log

# 이력 조회
cat .moai/update.log
```

---

## Performance Optimization

### 병렬 실행 전략 ⚡

**Before (순차 실행)**:

```text
Step 1: 명령어 파일 복사 (3초)
Step 2: 에이전트 파일 복사 (3초)
Step 3: 훅 파일 복사 (2초)
Step 4: Output Styles 복사 (2초)
───────────────────────────────
총 소요 시간: 10초
```

**After (병렬 실행)**:

```text
Step 1-4 병렬 실행 (최대 3초)
───────────────────────────────
총 소요 시간: 3초
절감: 7초 (70% 개선)
```

### 캐시 활용

```bash
# npm 캐시 활용 (2회차 업데이트부터)
npm install moai-adk@latest --prefer-offline

# 속도 향상: 30-40%
```

### Skip 옵션 (개발 중)

```bash
# 특정 단계 건너뛰기 (v0.3.x 예정)
/alfred:9-update --skip-hooks --skip-styles
```

---

## Related Commands

### 프로젝트 초기화

```bash
# 처음부터 다시 시작
/alfred:0-project
```

### 시스템 진단

```bash
# 터미널에서 실행
moai doctor

# 출력 예시
# ✅ Node.js: v20.10.0
# ✅ npm: v10.2.3
# ✅ moai-adk: v0.2.17
# ✅ 템플릿: 정상
# ✅ 설정: 정상
```

### 현재 상태 확인

```bash
# 터미널에서 실행
moai status

# 출력 예시
# 프로젝트: MoAI-ADK
# 버전: v0.2.17
# 모드: personal
# 진행 단계: 2-build
# 마지막 업데이트: 2025-10-11
```

---

## Error Messages

### 심각도별 아이콘

모든 에러 메시지는 일관된 심각도 표시를 사용합니다:

- **❌ Critical**: 작업 중단, 즉시 조치 필요
- **⚠️ Warning**: 주의 필요, 계속 진행 가능
- **ℹ️ Info**: 정보성 메시지, 참고용

### 메시지 형식

```text
[아이콘] [컨텍스트]: [문제 설명]
  → [권장 조치]
```

### 예시

```text
❌ 패키지 업데이트 실패: npm 캐시 손상
  → npm cache clean --force 실행 후 재시도

⚠️ 파일 복사 경고: 2개 파일 건너뜀 (사용자 수정)
  → 최신 템플릿: $TEMPLATE_ROOT/.moai/project/

ℹ️ product.md는 이미 작성되어 있어서 건너뜁니다
  → 필요 시 수동 업데이트 가능
```

---

## Architecture

### Alfred 중앙 오케스트레이션

```text
┌─────────────────────────────────────────────────────────┐
│                   UpdateOrchestrator                     │
├─────────────────────────────────────────────────────────┤
│ Phase 1: VersionChecker   (Alfred - Bash)                │
│          현재/최신 버전 비교                              │
│                                                          │
│ Phase 2: BackupManager    (Alfred - Bash)                │
│          .moai-backup/{timestamp}/ 생성                  │
│                                                          │
│ Phase 3: NpmUpdater       (Alfred - Bash)                │
│          패키지 매니저 감지 + 업데이트                    │
│                                                          │
│ Phase 4: TemplateSync     (Alfred - Claude Code Tools)   │
│          ⚡ Step 1-4 병렬 실행 (시스템 파일)              │
│          📚 Step 5-10 순차 실행 (병합/보존)               │
│          Glob, Read, Grep, Write, Bash                   │
│                                                          │
│ Phase 5: IntegrityCheck   (Alfred - Claude Code Tools)   │
│          파일 개수, YAML, 버전, 권한 검증                 │
│          Glob, Read, Grep, Bash                          │
│                                                          │
│ Phase 5.5: QualityCheck   (trust-checker @agent)         │
│            --check-quality 옵션 시에만 실행               │
│            TRUST 5원칙 검증 (3-5초)                       │
└─────────────────────────────────────────────────────────┘
```

### 핵심 원칙

1. **Alfred 직접 제어**: 모든 Phase를 Alfred가 직접 실행
2. **Claude Code 도구**: TypeScript 대신 Glob, Read, Write 등 활용
3. **자연어 지침**: 모든 로직이 자연어로 표현
4. **선택적 에이전트 호출**: 품질 검증 시에만 trust-checker 사용

---

## Migration Notes

### v0.2.16 → v0.2.17

**주요 변경사항**:

- ✅ 사용자 데이터 보호 강화 (.moai/specs/, .moai/reports/)
- ✅ 스마트 딥 병합 개선 (config.json)
- ✅ 병렬 실행 최적화 (75% 속도 향상)

**마이그레이션 단계**:

```bash
# 1. 백업 생성 (자동)
/alfred:9-update

# 2. 변경사항 확인
git diff .moai/config.json

# 3. 정상 동작 확인
moai doctor
moai status

# 4. 완료
```

### v0.2.x → v0.3.x (예정)

**Breaking Changes**:

- config.json 구조 변경
- 새로운 필수 필드 추가

**마이그레이션 가이드**: `docs/migration/v0.2-to-v0.3.md` 참조

---

## FAQ

### Q1. 업데이트 중 실패하면 어떻게 하나요?

```bash
# 백업에서 수동 복원
cp -r .moai-backup/2025-10-11-15-30-00/.claude/ ./
cp -r .moai-backup/2025-10-11-15-30-00/.moai/ ./
cp .moai-backup/2025-10-11-15-30-00/CLAUDE.md ./
```

### Q2. 백업 없이 업데이트하려면?

```bash
# --force 옵션 (위험)
/alfred:9-update --force

# ⚠️ 경고: 롤백 불가능
```

### Q3. 특정 파일만 업데이트하려면?

```bash
# 현재는 전체 업데이트만 지원
# v0.3.x에서 선택적 업데이트 지원 예정

# 임시 방법: 수동 복사
cp $TEMPLATE_ROOT/.moai/config.json .moai/
```

### Q4. 업데이트 후 설정이 초기화되었어요

```bash
# 백업에서 설정 복원
cp .moai-backup/2025-10-11-15-30-00/.moai/config.json .moai/

# 또는 수동 재설정
# .moai/config.json 편집
```

### Q5. 품질 검증은 필수인가요?

```text
필수 아님 (선택 사항)

기본 업데이트: /alfred:9-update
품질 검증 포함: /alfred:9-update --check-quality

권장 사용:
- 프로덕션: --check-quality ✅
- 개발 환경: 기본 모드
```

---

## Next Steps

업데이트 완료 후 다음 단계로 진행합니다:

1. **[Stage 0: Project Init](0-project.md)** - 프로젝트 초기화
2. **[Stage 1: SPEC Writing](1-spec.md)** - 명세 작성
3. **Troubleshooting** - 문제 해결 가이드

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>안전한 업데이트, 완벽한 복원</strong> 🔄</p>
  <p>정기적으로 업데이트하세요!</p>
</div>
