---
name: moai:9-update
description: MoAI-ADK 패키지 및 템플릿 업데이트 (백업 자동 생성, 설정 파일 보존)
argument-hint: [--check|--force]
tools: Read, Write, Bash, Grep, Glob
---

# 🔄 MoAI-ADK 프로젝트 업데이트

## 🎯 커맨드 목적

MoAI-ADK npm 패키지를 최신 버전으로 업데이트하고, 템플릿 파일을 안전하게 갱신합니다.

## 📋 실행 흐름

1. **버전 확인**: 현재 버전과 최신 버전 비교
2. **백업 생성**: 기존 파일을 타임스탬프 기반으로 백업
3. **패키지 업데이트**: npm install moai-adk@latest
4. **템플릿 복사**: Claude Code 도구로 안전하게 파일 복사
5. **검증**: 파일 존재 및 내용 무결성 확인

## 🔗 연관 에이전트

- **Primary**: None (직접 실행)
- **Quality Check**: trust-checker (✅ 품질 보증 리드) - 업데이트 후 검증 (선택적)
- **Secondary**: None (Claude Code 도구 직접 사용)

## 💡 사용 예시

```bash
/moai:9-update              # 업데이트 확인 및 실행
/moai:9-update --check      # 업데이트 가능 여부만 확인
/moai:9-update --force      # 강제 업데이트 (백업 없이)
```

## 명령어 개요

MoAI-ADK npm 패키지를 최신 버전으로 업데이트하고, `.claude`, `.moai`, `CLAUDE.md` 파일을 최신 템플릿으로 갱신합니다.

자동 백업, 안전한 파일 복사, 설정 보존을 보장하는 프로젝트 업데이트 시스템입니다.

## 사용법

```bash
/moai:9-update              # 업데이트 확인 및 실행
/moai:9-update --check      # 업데이트 가능 여부만 확인
/moai:9-update --force      # 강제 업데이트 (백업 없이)
```

**인수 처리**: `$ARGUMENTS`를 통해 `--check` 또는 `--force` 옵션 전달

## 실행 절차

### Phase 1: 버전 확인 및 검증

현재 버전과 최신 버전을 비교합니다.

```bash
# 1. 현재 설치된 버전 확인
npm list moai-adk --depth=0

# 2. 최신 버전 조회
npm view moai-adk version

# 3. 업데이트 필요 여부 판단
# 버전 비교 후 사용자에게 보고
```

**조건부 실행**: `--check` 옵션이면 여기서 중단하고 결과만 보고

### Phase 2: 백업 생성 (기본값)

기존 파일을 안전하게 백업합니다.

```bash
# 백업 디렉토리 생성 (타임스탬프 기반)
BACKUP_DIR=".moai-backup/$(date +%Y-%m-%d-%H-%M-%S)"
mkdir -p "$BACKUP_DIR"

# 백업 대상 파일 복사
cp -r .claude "$BACKUP_DIR/" 2>/dev/null || true
cp -r .moai "$BACKUP_DIR/" 2>/dev/null || true
cp CLAUDE.md "$BACKUP_DIR/" 2>/dev/null || true
```

**백업 구조**:
```
.moai-backup/
  └── YYYY-MM-DD-HH-mm-ss/
      ├── .claude/
      ├── .moai/
      └── CLAUDE.md
```

**예외**: `--force` 옵션이 제공되면 이 단계 건너뛰기

### Phase 3: npm 패키지 업데이트

MoAI-ADK 패키지 최신 버전을 설치합니다.

```bash
# package.json 존재 확인
if [ -f "package.json" ]; then
    npm install moai-adk@latest
else
    npm install -g moai-adk@latest
fi
```

**검증**: 설치 성공 여부 확인 및 새 버전 확인

### Phase 4: 템플릿 파일 복사

최신 템플릿 파일을 프로젝트에 복사합니다.

**복사 대상 파일**:
```
node_modules/moai-adk/templates/
  ├── .claude/
  │   ├── commands/moai/*.md     → .claude/commands/moai/
  │   ├── agents/moai/*.md       → .claude/agents/moai/
  │   └── hooks/moai/*.cjs       → .claude/hooks/moai/
  ├── .moai/
  │   ├── memory/
  │   │   └── development-guide.md → .moai/memory/
  │   └── project/
  │       ├── product.md         → .moai/project/
  │       ├── structure.md       → .moai/project/
  │       └── tech.md            → .moai/project/
  └── CLAUDE.md                  → CLAUDE.md
```

**보존 대상 (덮어쓰지 않음)**:
```
.moai/
  ├── specs/                     # 모든 SPEC 파일
  ├── indexes/                   # TAG 인덱스
  ├── reports/                   # 동기화 리포트
  └── config.json               # 프로젝트 설정
```

**실행 절차**:

#### Step 1: 패키지 설치 경로 찾기

```bash
# 로컬 node_modules 경로 확인
npm root

# 로컬에 없으면 글로벌 경로 확인
npm root -g
```

템플릿 경로 구성: `{npm_root}/moai-adk/templates`

#### Step 2: 템플릿 디렉토리 존재 확인

```bash
# 템플릿 디렉토리 검증
ls "{npm_root}/moai-adk/templates"
```

**오류 처리**: 템플릿 디렉토리가 없으면 Phase 3로 돌아가서 패키지 재설치

#### Step 3: 템플릿 파일 복사 (Claude Code 도구 기반)

Claude Code 도구를 활용하여 템플릿 파일을 안전하게 복사합니다:

**복사 대상 디렉토리 및 파일**:

1. `.claude/commands/moai/` - 명령어 파일 (*.md)
2. `.claude/agents/moai/` - 에이전트 파일 (*.md)
3. `.claude/hooks/moai/` - 훅 파일 (*.cjs)
4. `.moai/memory/development-guide.md` - 개발 가이드
5. `.moai/project/` - 프로젝트 문서 (product.md, structure.md, tech.md)
6. `CLAUDE.md` - 프로젝트 루트 설정 파일

**복사 절차 (각 카테고리별)**:

##### A. 명령어 파일 복사 (.claude/commands/moai/)

```text
1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Glob] "{npm_root}/moai-adk/templates/.claude/commands/moai/*.md" 패턴으로 파일 검색
3. 각 파일마다 반복:
   a. [Read] "{npm_root}/moai-adk/templates/.claude/commands/moai/{filename}"
   b. [Write] "./.claude/commands/moai/{filename}" (내용 그대로 복사)
4. 성공 메시지: "✅ .claude/commands/moai/ ({개수}개 파일 복사 완료)"
```

**오류 처리**:

- Glob 결과가 비어있으면 → 템플릿 디렉토리 경로 재확인 (Step 2로 복귀)
- Read 실패 시 → 해당 파일 건너뛰고 오류 로그 기록
- Write 실패 시 → 디렉토리 생성 후 재시도: `[Bash] mkdir -p .claude/commands/moai`

##### B. 에이전트 파일 복사 (.claude/agents/moai/)

```text
1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Glob] "{npm_root}/moai-adk/templates/.claude/agents/moai/*.md" 패턴으로 파일 검색
3. 각 파일마다 반복:
   a. [Read] "{npm_root}/moai-adk/templates/.claude/agents/moai/{filename}"
   b. [Write] "./.claude/agents/moai/{filename}"
4. 성공 메시지: "✅ .claude/agents/moai/ ({개수}개 파일 복사 완료)"
```

**오류 처리**: 명령어 파일과 동일

##### C. 훅 파일 복사 (.claude/hooks/moai/)

```text
1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Glob] "{npm_root}/moai-adk/templates/.claude/hooks/moai/*.cjs" 패턴으로 파일 검색
3. 각 파일마다 반복:
   a. [Read] "{npm_root}/moai-adk/templates/.claude/hooks/moai/{filename}"
   b. [Write] "./.claude/hooks/moai/{filename}"
4. [Bash] chmod +x .claude/hooks/moai/*.cjs (실행 권한 부여)
5. 성공 메시지: "✅ .claude/hooks/moai/ ({개수}개 파일 복사 완료)"
```

**오류 처리**:

- chmod 실패 시 → 경고 메시지만 표시하고 계속 진행

##### D. 개발 가이드 복사 (.moai/memory/development-guide.md)

```text
1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Read] "{npm_root}/moai-adk/templates/.moai/memory/development-guide.md"
3. [Write] "./.moai/memory/development-guide.md" (무조건 덮어쓰기)
4. 성공 메시지: "✅ .moai/memory/development-guide.md 업데이트 완료"
```

**오류 처리**:

- Write 실패 시 → `[Bash] mkdir -p .moai/memory` 후 재시도

##### E. 프로젝트 문서 복사 (.moai/project/)

```text
각 파일(product.md, structure.md, tech.md)마다:

1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Read] "./.moai/project/{filename}" (기존 파일 확인, 없으면 에러 무시)
3. [Grep] 기존 파일에서 "{{PROJECT_NAME}}" 패턴 검색
   - 검색 결과 있음 → 템플릿 상태로 판단 → 덮어쓰기
   - 검색 결과 없음 → 사용자 수정 상태 → 백업 후 덮어쓰기
   - 파일 없음 → 새로 생성
4. 백업이 필요한 경우:
   a. [Read] "./.moai/project/{filename}"
   b. [Write] "./.moai-backup/{timestamp}/.moai/project/{filename}" (백업 생성)
5. [Read] "{npm_root}/moai-adk/templates/.moai/project/{filename}"
6. [Write] "./.moai/project/{filename}"
7. 성공 메시지: "✅ .moai/project/{filename} (백업: {yes/no})"
```

**오류 처리**:

- Grep 도구가 없으면 → 무조건 백업 후 덮어쓰기
- Write 실패 시 → `[Bash] mkdir -p .moai/project` 후 재시도

##### F. CLAUDE.md 복사 (프로젝트 루트)

```text
1. [Bash] npm root 실행하여 {npm_root} 경로 확인
2. [Read] "./CLAUDE.md" (기존 파일 확인, 없으면 에러 무시)
3. [Grep] 기존 파일에서 "{{PROJECT_NAME}}" 패턴 검색
   - 검색 결과 있음 → 템플릿 상태 → 덮어쓰기
   - 검색 결과 없음 → 사용자 수정 상태 → 백업 후 덮어쓰기
   - 파일 없음 → 새로 생성
4. 백업이 필요한 경우:
   a. [Read] "./CLAUDE.md"
   b. [Write] "./.moai-backup/{timestamp}/CLAUDE.md"
5. [Read] "{npm_root}/moai-adk/templates/CLAUDE.md"
6. [Write] "./CLAUDE.md"
7. 성공 메시지: "✅ CLAUDE.md 업데이트 완료 (백업: {yes/no})"
```

**오류 처리**: 프로젝트 문서와 동일

**전체 복사 절차 요약**:

```text
Phase 4: 템플릿 파일 복사 시작...

[Step 1] npm root 경로 확인
  → [Bash] npm root
  → 로컬 node_modules: /Users/user/project/node_modules
  → 템플릿 경로: /Users/user/project/node_modules/moai-adk/templates

[Step 2] 명령어 파일 복사 (A)
  → [Glob] 10개 파일 발견
  → [Read/Write] 각 파일 복사...
  → ✅ 10개 파일 복사 완료

[Step 3] 에이전트 파일 복사 (B)
  → [Glob] 8개 파일 발견
  → [Read/Write] 각 파일 복사...
  → ✅ 8개 파일 복사 완료

[Step 4] 훅 파일 복사 (C)
  → [Glob] 3개 파일 발견
  → [Read/Write] 각 파일 복사...
  → [Bash] chmod +x 실행 권한 부여
  → ✅ 3개 파일 복사 완료

[Step 5] 개발 가이드 복사 (D)
  → [Read/Write] development-guide.md
  → ✅ 개발 가이드 업데이트 완료

[Step 6] 프로젝트 문서 복사 (E)
  → product.md: [Grep] 템플릿 상태 확인 → 덮어쓰기 ✅
  → structure.md: [Grep] 사용자 수정 감지 → 백업 후 덮어쓰기 ✅
  → tech.md: 파일 없음 → 새로 생성 ✅

[Step 7] CLAUDE.md 복사 (F)
  → [Grep] 사용자 수정 상태 확인 → 백업 후 덮어쓰기 ✅

Phase 4 완료!
```

### Phase 5: 업데이트 검증

업데이트 완료를 검증합니다.

#### 검증 단계

1. **파일 존재 확인 (Glob 도구)**:

   ```text
   .claude/commands/moai/*.md 파일 개수 확인
   .claude/agents/moai/*.md 파일 개수 확인
   .claude/hooks/moai/*.cjs 파일 개수 확인
   .moai/memory/development-guide.md 존재 확인
   .moai/project/*.md 파일 개수 확인 (product, structure, tech)
   CLAUDE.md 존재 확인
   ```

2. **파일 내용 검증 (Read 도구)**:

   ```text
   development-guide.md에서 버전 정보 확인
   CLAUDE.md에서 @TAG 시스템 섹션 존재 확인
   commands/moai/ 파일들의 YAML frontmatter 유효성 확인
   ```

3. **패키지 버전 확인 (Bash 도구)**:

   ```bash
   # 새 버전 설치 확인
   npm list moai-adk --depth=0

   # 예상 출력: moai-adk@{새버전}
   ```

4. **파일 카운트 비교**:

   ```text
   예상 파일 개수:
   - .claude/commands/moai/: ~10개 파일
   - .claude/agents/moai/: ~8개 파일
   - .claude/hooks/moai/: ~3개 파일
   - .moai/memory/: 1개 파일 (development-guide.md)
   - .moai/project/: 3개 파일 (product, structure, tech)
   - 루트: 1개 파일 (CLAUDE.md)
   ```

**검증 체크리스트**:

- [ ] .claude/commands/moai/*.md 파일 존재 (10개 예상)
- [ ] .claude/agents/moai/*.md 파일 존재 (8개 예상)
- [ ] .claude/hooks/moai/*.cjs 파일 존재 (3개 예상)
- [ ] .moai/memory/development-guide.md 최신 버전
- [ ] .moai/project/*.md 파일 존재 (3개)
- [ ] CLAUDE.md 업데이트 완료 (@TAG 섹션 확인)
- [ ] npm list moai-adk로 새 버전 확인

**검증 실패 시 조치**:

- 파일 누락: Phase 4 Step 3으로 돌아가 해당 파일 재복사
- 버전 불일치: Phase 3으로 돌아가 패키지 재설치
- 내용 손상: 백업에서 복원 후 Phase 4부터 재시작

### Phase 5.5: 업데이트 후 품질 검증 (선택적)

업데이트 완료 후 선택적으로 시스템 전체 무결성을 검증할 수 있습니다.

**실행 조건**: 사용자가 `--check-quality` 옵션을 제공한 경우

**검증 목적**:
- 업데이트 후 시스템 전체 무결성 확인
- 업데이트로 인한 품질 저하 감지
- 설정 파일 및 문서 일관성 검증

**실행 방식**:
```bash
# --check-quality 옵션 추가
/moai:9-update --check-quality
```

**검증 항목**:
- **파일 무결성**: 업데이트된 모든 파일의 구조 검증
- **설정 일관성**: config.json과 문서 간 일관성 확인
- **TAG 체계**: 문서 내 @TAG 형식 준수 확인
- **EARS 구문**: SPEC 템플릿의 EARS 형식 검증

**검증 실행**: Level 1 빠른 스캔 (3-5초)

**검증 결과 처리**:

✅ **Pass**: 업데이트 성공적으로 완료
- 모든 파일 정상
- 시스템 무결성 유지

⚠️ **Warning**: 경고 표시 후 완료
- 일부 문서 포맷 이슈
- 권장사항 미적용
- 사용자 확인 권장

❌ **Critical**: 업데이트 롤백 권장
- 파일 손상 감지
- 설정 불일치 발견
- 사용자 선택: "롤백" 또는 "무시하고 진행"

**검증 건너뛰기**:
- 기본적으로 검증은 실행되지 않음
- `--check-quality` 옵션 제공 시에만 실행

## 출력 예시

```text
🔍 MoAI-ADK 업데이트 확인 중...

📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
✅ 업데이트 가능

💾 백업 생성 중...
   → .moai-backup/2025-09-30-23-45-00/

📦 패키지 업데이트 중...
   npm install moai-adk@0.0.2
   ✅ 패키지 업데이트 완료

📁 패키지 경로 확인 중...
   npm root → /Users/user/project/node_modules
   ✅ 템플릿 경로: /Users/user/project/node_modules/moai-adk/templates

📄 템플릿 파일 복사 중...
   [Glob] .claude/commands/moai/*.md → 10개 파일 발견
   [Read/Write] 1-spec.md → .claude/commands/moai/1-spec.md
   [Read/Write] 2-build.md → .claude/commands/moai/2-build.md
   [Read/Write] 3-sync.md → .claude/commands/moai/3-sync.md
   ... (7개 더)
   ✅ .claude/commands/moai/ (10개 파일 복사 완료)

   [Glob] .claude/agents/moai/*.md → 8개 파일 발견
   [Read/Write] spec-builder.md → .claude/agents/moai/spec-builder.md
   [Read/Write] code-builder.md → .claude/agents/moai/code-builder.md
   ... (6개 더)
   ✅ .claude/agents/moai/ (8개 파일 복사 완료)

   [Glob] .claude/hooks/moai/*.cjs → 3개 파일 발견
   ✅ .claude/hooks/moai/ (3개 파일 복사 완료)

   [Read/Write] development-guide.md → .moai/memory/development-guide.md
   ✅ .moai/memory/development-guide.md

   [Read/Write] product.md → .moai/project/product.md
   [Read/Write] structure.md → .moai/project/structure.md
   [Read/Write] tech.md → .moai/project/tech.md
   ✅ .moai/project/*.md (3개 파일)

   [Read/Write] CLAUDE.md → ./CLAUDE.md
   ✅ CLAUDE.md

🔍 검증 중...
   [Glob] .claude/commands/moai/*.md → 10개 ✅
   [Glob] .claude/agents/moai/*.md → 8개 ✅
   [Glob] .claude/hooks/moai/*.cjs → 3개 ✅
   [Read] development-guide.md 내용 확인 ✅
   [Read] CLAUDE.md @TAG 섹션 확인 ✅
   [Bash] npm list moai-adk@0.0.2 ✅

✨ 업데이트 완료!

롤백이 필요하면: moai restore --from=2025-09-30-23-45-00
```

## 안전 장치

### 자동 백업

- `--force` 옵션이 아닌 한 항상 백업 생성
- 백업 위치: `.moai-backup/YYYY-MM-DD-HH-mm-ss/`
- 백업은 수동 삭제 전까지 유지

### 충돌 방지

- 사용자 생성 파일 (`.moai/specs/*`) 절대 건드리지 않음
- 프로젝트 설정 파일 (`.moai/config.json`) 보존
- TAG 인덱스 및 리포트 보존

### 롤백 지원

```bash
# 백업 목록 확인
moai restore --list

# 특정 백업으로 복원
moai restore --from=2025-09-30-23-45-00

# 최근 백업으로 복원
moai restore --latest
```

## 오류 처리

### 일반적인 오류 및 해결 방법

**오류 1**: `npm install` 실패

```text
증상: Phase 3에서 패키지 설치 실패
원인:
  - 인터넷 연결 문제
  - npm 캐시 손상
  - 권한 문제 (글로벌 설치 시)

해결:
  1. [Bash] npm cache clean --force
  2. 인터넷 연결 확인
  3. 글로벌 설치 시: sudo 권한 확인
  4. 재시도
```

**오류 2**: 템플릿 디렉토리를 찾을 수 없음

```text
증상: Phase 4 Step 2에서 템플릿 경로 검증 실패
원인:
  - 패키지 설치가 불완전함
  - npm root 경로가 잘못됨

해결:
  1. [Bash] npm root 출력 확인
  2. [Bash] ls {npm_root}/moai-adk로 패키지 존재 확인
  3. 없으면 Phase 3로 돌아가 패키지 재설치
  4. 글로벌 설치도 확인: npm root -g
```

**오류 3**: 파일 복사 실패 (Write 도구 오류)

```text
증상: Phase 4 Step 3에서 Write 도구 실패
원인:
  - 대상 디렉토리가 없음
  - 파일 권한 문제
  - 디스크 용량 부족

해결:
  1. [Bash] mkdir -p .claude/commands/moai (디렉토리 생성)
  2. [Bash] chmod -R 755 .claude (권한 부여)
  3. [Bash] df -h (디스크 용량 확인)
  4. 백업에서 복원 후 재시도
```

**오류 4**: 검증 실패 (파일 개수 불일치)

```text
증상: Phase 5에서 예상 파일 개수와 실제 개수가 다름
원인:
  - 일부 파일 복사 실패
  - 템플릿 버전 불일치

해결:
  1. [Glob] 각 디렉토리의 파일 개수 재확인
  2. 누락된 파일 목록 확인
  3. Phase 4 Step 3으로 돌아가 누락 파일만 재복사
  4. 또는 전체 Phase 4 재실행
```

**오류 5**: 버전 확인 불가

```text
증상: Phase 1이나 Phase 5에서 버전 확인 실패
원인:
  - npm 미설치
  - npm registry 접근 불가
  - package.json 손상

해결:
  1. [Bash] npm --version (npm 설치 확인)
  2. [Bash] npm config get registry (registry 확인)
  3. [Read] package.json 검증
  4. npm 재설치 또는 registry 변경
```

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

업데이트 완료 후:

1. **설정 확인**: `.moai/config.json`에서 프로젝트 설정 확인
2. **문서 검토**: 업데이트된 `CLAUDE.md`와 `development-guide.md` 확인
3. **테스트 실행**: 기존 SPEC이 있다면 `/moai:2-build` 재실행 권장
4. **커밋**: 변경사항을 Git에 커밋

## 관련 명령어

- `/moai:8-project` - 프로젝트 초기화/재설정
- `moai restore` - 백업에서 복원
- `moai doctor` - 시스템 진단
- `moai status` - 현재 상태 확인

## 버전 호환성

- **v0.0.1 → v0.0.2**: 호환 ✅
- **v0.0.x → v0.1.x**: 마이너 버전 업그레이드 - 설정 확인 필요
- **v0.x.x → v1.x.x**: 메이저 버전 업그레이드 - 마이그레이션 가이드 참조
