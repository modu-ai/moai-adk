# @SPEC:UPDATE-REFACTOR-001 수락 기준

## 개요

본 문서는 `/alfred:9-update` 명령어 SPEC 문서화 작업의 수락 기준을 정의합니다. Given-When-Then 형식의 테스트 시나리오, 품질 게이트 기준, 검증 방법을 포함합니다.

## Definition of Done (완료 조건)

### 1. SPEC 문서 완성도

- ✅ **spec.md**: YAML Front Matter, HISTORY, EARS 구조 완전성
- ✅ **plan.md**: 구현 계획, 마일스톤, 기술적 접근 방법
- ✅ **acceptance.md**: 수락 기준, 테스트 시나리오 (본 문서)

### 2. EARS 명세 완전성

- ✅ **Environment (환경)**: 시스템 컨텍스트, 전제 조건 정의 완료
- ✅ **Assumptions (가정)**: 5개 주요 가정 명시 완료
- ✅ **Requirements (요구사항)**: Ubiquitous, Event-driven, State-driven, Optional, Constraints 정의 완료
- ✅ **Specifications (상세 명세)**: Phase 1-5 상세 절차 문서화 완료

### 3. 코드 품질 기준

- ✅ **restore.ts**: TypeScript 타입 안전성, 테스트 작성
- ✅ **9-update.md**: Alfred 직접 실행 방식, Claude Code 도구 기반
- ✅ **@TAG 체인**: SPEC, CODE, DOC 연결 무결성

### 4. 문서 품질 기준

- ✅ **추적성**: @TAG 참조 정확성
- ✅ **가독성**: 명확한 구조, 적절한 예시
- ✅ **완전성**: 오류 시나리오, 복구 절차 포함

## 테스트 시나리오

### 시나리오 1: 정상 업데이트 흐름 (Happy Path)

**Given** (사전 조건):
- MoAI-ADK v0.0.1이 설치되어 있다
- npm 레지스트리에 v0.0.2가 존재한다
- `.moai/`, `.claude/` 디렉토리가 존재한다
- `CLAUDE.md`에 프로젝트 정보가 작성되어 있다 (템플릿 상태 아님)

**When** (실행):
```bash
/alfred:9-update
```

**Then** (예상 결과):
1. Phase 1 성공:
   - "📦 현재 버전: v0.0.1" 출력
   - "⚡ 최신 버전: v0.0.2" 출력
   - "✅ 업데이트 가능" 출력

2. Phase 2 성공:
   - `.moai-backup/YYYY-MM-DD-HH-mm-ss/` 디렉토리 생성
   - `.claude/`, `.moai/`, `CLAUDE.md` 백업 완료

3. Phase 3 성공:
   - `npm install moai-adk@latest` 실행
   - `npm list moai-adk` → v0.0.2 확인

4. Phase 4 성공:
   - 명령어 파일 ~10개 복사
   - 에이전트 파일 ~9개 복사
   - 훅 파일 4개 복사 + 권한 설정 (Unix)
   - Output Styles 4개 복사
   - development-guide.md 업데이트
   - 프로젝트 문서 (product/structure/tech) 상태 확인:
     - 템플릿 상태이면 덮어쓰기
     - 사용자 수정 상태이면 보존
   - `CLAUDE.md` 지능형 병합:
     - 프로젝트 정보 (이름, 설명, 버전, 모드) 유지
     - 템플릿 구조 (Alfred 에이전트 목록 등) 최신화
   - `config.json` 딥 병합:
     - `project.*` 사용자 값 유지
     - `git_strategy.*` 사용자 값 유지
     - `tags.categories` 병합 (템플릿 + 사용자)
     - `pipeline.current_stage` 사용자 값 유지

5. Phase 5 성공:
   - 파일 개수 검증 통과
   - YAML frontmatter 검증 통과
   - 버전 정합성 확인 (v0.0.2)
   - 훅 파일 권한 확인 (755, Unix만)

6. 최종 출력:
   ```text
   ✨ 업데이트 완료!

   롤백이 필요하면: moai restore .moai-backup/2025-10-06-15-30-00
   ```

**검증 명령어**:
```bash
# 버전 확인
npm list moai-adk --depth=0
# 예상: moai-adk@0.0.2

# 파일 개수 확인
ls -1 .claude/commands/alfred/*.md | wc -l
# 예상: ~10

# 프로젝트 정보 유지 확인
grep "이름:" CLAUDE.md
# 예상: 기존 프로젝트 이름 유지

# 백업 존재 확인
ls -d .moai-backup/*
# 예상: 최소 1개 백업 디렉토리
```

---

### 시나리오 2: 버전 확인만 수행 (--check 옵션)

**Given**:
- MoAI-ADK v0.0.1이 설치되어 있다
- npm 레지스트리에 v0.0.2가 존재한다

**When**:
```bash
/alfred:9-update --check
```

**Then**:
1. Phase 1 실행:
   - "📦 현재 버전: v0.0.1" 출력
   - "⚡ 최신 버전: v0.0.2" 출력
   - "✅ 업데이트 가능" 출력

2. Phase 2-5 건너뛰기:
   - 백업 생성 안 함
   - npm 설치 안 함
   - 템플릿 복사 안 함

3. 최종 출력:
   ```text
   ℹ️  업데이트 확인 완료 (--check 옵션)
   ℹ️  업데이트를 실행하려면: /alfred:9-update
   ```

**검증**:
```bash
# 버전이 변경되지 않았는지 확인
npm list moai-adk --depth=0
# 예상: moai-adk@0.0.1 (그대로)

# 백업이 생성되지 않았는지 확인
ls .moai-backup/ 2>/dev/null
# 예상: 디렉토리 없음 또는 기존 백업만 존재
```

---

### 시나리오 3: 사용자 작업물 보존 (프로젝트 문서)

**Given**:
- `.moai/project/product.md`에 실제 프로젝트 정보가 작성되어 있다 (템플릿 상태 아님)
- `{{PROJECT_NAME}}` 패턴이 존재하지 않는다

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 - Step 7 실행:
   - `[Grep] "{{PROJECT_NAME}}" -n .moai/project/product.md` → 결과 없음
   - "ℹ️  product.md는 이미 프로젝트 정보가 작성되어 있어서 건너뜁니다" 출력
   - "💡 최신 템플릿 참조: {npm_root}/moai-adk/templates/.moai/project/product.md" 출력
   - `product.md` 파일 복사 건너뛰기

2. 사용자 파일 보존 확인:
   ```bash
   # 파일 수정 시간 확인 (변경되지 않음)
   stat -f "%Sm" .moai/project/product.md
   # 예상: 업데이트 이전 시간
   ```

**검증**:
```bash
# 사용자 작성 내용 유지 확인
grep "핵심 미션" .moai/project/product.md
# 예상: 사용자가 작성한 내용 그대로 존재
```

---

### 시나리오 4: 템플릿 상태 파일 덮어쓰기

**Given**:
- `.moai/project/product.md`가 템플릿 상태이다
- `{{PROJECT_NAME}}` 패턴이 존재한다

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 - Step 7 실행:
   - `[Grep] "{{PROJECT_NAME}}" -n .moai/project/product.md` → 결과 있음
   - "✅ .moai/project/product.md (템플릿 → 최신 버전)" 출력
   - 최신 템플릿으로 덮어쓰기

2. 템플릿 업데이트 확인:
   ```bash
   # 최신 템플릿 내용 확인
   grep "{{PROJECT_NAME}}" .moai/project/product.md
   # 예상: 패턴 존재
   ```

**검증**:
```bash
# 파일 수정 시간 확인 (최근 시간)
stat -f "%Sm" .moai/project/product.md
# 예상: 업데이트 실행 시간
```

---

### 시나리오 5: CLAUDE.md 지능형 병합

**Given**:
- `CLAUDE.md`에 프로젝트 정보가 작성되어 있다
- 프로젝트 이름: "MyProject"
- 프로젝트 설명: "A test project"
- 프로젝트 버전: "1.0.0"
- 프로젝트 모드: "team"

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 - Step 10 실행:
   - `[Grep] "{{PROJECT_NAME}}" -n ./CLAUDE.md` → 결과 없음 (사용자 수정 상태)
   - 프로젝트 정보 추출:
     ```text
     [Grep] "^- \*\*이름\*\*:" ./CLAUDE.md → "MyProject"
     [Grep] "^- \*\*설명\*\*:" ./CLAUDE.md → "A test project"
     [Grep] "^- \*\*버전\*\*:" ./CLAUDE.md → "1.0.0"
     [Grep] "^- \*\*모드\*\*:" ./CLAUDE.md → "team"
     ```
   - 최신 템플릿 읽기 + 정보 주입
   - "🔄 CLAUDE.md 병합 완료 (템플릿 최신화 + 프로젝트 정보 유지)" 출력

2. 병합 결과 확인:
   - 프로젝트 정보 유지:
     ```bash
     grep "이름.*MyProject" CLAUDE.md
     grep "설명.*A test project" CLAUDE.md
     grep "버전.*1.0.0" CLAUDE.md
     grep "모드.*team" CLAUDE.md
     ```
   - 템플릿 구조 최신화:
     ```bash
     grep "spec-builder" CLAUDE.md
     grep "code-builder" CLAUDE.md
     grep "doc-syncer" CLAUDE.md
     # 예상: 최신 에이전트 목록 존재
     ```

**검증**:
```bash
# 프로젝트 정보 유지 확인
grep -A1 "프로젝트 정보" CLAUDE.md
# 예상:
# - **이름**: MyProject
# - **설명**: A test project
# - **버전**: 1.0.0
# - **모드**: team

# Alfred 에이전트 목록 확인 (최신 템플릿)
grep "9개 전문 에이전트" CLAUDE.md
# 예상: 존재
```

---

### 시나리오 6: config.json 딥 병합

**Given**:
- `.moai/config.json`에 사용자 설정이 작성되어 있다
- `project.name`: "MyProject"
- `git_strategy.branch_prefix`: "custom/"
- `tags.categories`: ["AUTH", "API", "CUSTOM"]
- `pipeline.current_stage`: "build"

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 - Step 11 실행:
   - JSON 파싱 성공
   - 사용자 설정값 추출
   - 딥 병합 실행:
     ```javascript
     merged = {
       project: userConfig.project,  // "MyProject" 유지
       git_strategy: userConfig.git_strategy,  // "custom/" 유지
       tags: {
         ...templateConfig.tags,
         categories: ["AUTH", "API", "SPEC", "TEST", "CODE", "DOC", "CUSTOM"]  // 병합
       },
       pipeline: {
         ...templateConfig.pipeline,
         current_stage: "build"  // 사용자 값 유지
       }
     }
     ```
   - "🔄 config.json 병합 완료" 출력

2. 병합 결과 확인:
   ```bash
   # project 필드 유지
   jq '.project.name' .moai/config.json
   # 예상: "MyProject"

   # git_strategy 유지
   jq '.git_strategy.branch_prefix' .moai/config.json
   # 예상: "custom/"

   # tags.categories 병합 (템플릿 + 사용자)
   jq '.tags.categories' .moai/config.json
   # 예상: ["AUTH", "API", "SPEC", "TEST", "CODE", "DOC", "CUSTOM"]

   # pipeline.current_stage 유지
   jq '.pipeline.current_stage' .moai/config.json
   # 예상: "build"
   ```

**검증**:
```bash
# JSON 유효성 검사
jq . .moai/config.json > /dev/null
echo $?
# 예상: 0 (파싱 성공)

# 필드 존재 확인
jq 'has("project") and has("git_strategy") and has("tags") and has("pipeline")' .moai/config.json
# 예상: true
```

---

### 시나리오 7: 파일 복사 실패 복구

**Given**:
- 디스크 공간이 부족하여 일부 파일 복사 실패 예상

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 실행 중 오류 발생:
   - `[Write] .claude/commands/alfred/2-build.md` 실패
   - "❌ 2-build.md 복사 실패: 디스크 공간 부족" 로그
   - 실패 목록에 추가: `failed_files = [2-build.md]`

2. 나머지 파일 계속 처리:
   - 3-spec.md, 8-project.md 등 계속 복사
   - 성공한 파일 로그 출력

3. Phase 4 종료 후 실패 보고:
   ```text
   ⚠️ 1개 파일 복사 실패: [2-build.md]

   다음 조치를 선택하세요:
     1. Phase 4 재실행 (실패한 파일만)
     2. 백업 복원
     3. 무시하고 진행 (권장하지 않음)
   ```

4. 사용자 선택 대기

**검증**:
```bash
# 실패 파일 확인
ls .claude/commands/alfred/2-build.md
# 예상: 파일 없음

# 성공 파일 확인
ls .claude/commands/alfred/1-spec.md
# 예상: 파일 존재
```

---

### 시나리오 8: 버전 불일치 검증 실패

**Given**:
- npm 패키지는 v0.0.2로 업데이트되었으나
- development-guide.md는 v0.0.1 버전 정보 포함 (템플릿 복사 실패)

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 5.3 실행:
   - `[Grep] "version:" .moai/memory/development-guide.md` → v0.0.1
   - `[Bash] npm list moai-adk` → v0.0.2
   - "❌ 버전 불일치 감지" 출력

2. 원인 분석 안내:
   ```text
   ⚠️ development-guide.md 버전(v0.0.1)과 패키지 버전(v0.0.2)이 불일치합니다.
   이는 템플릿 복사가 제대로 되지 않았음을 의미합니다.

   가능한 원인:
   a. npm 캐시 손상
   b. 템플릿 디렉토리 경로 오류
   c. 파일 복사 중 네트워크 오류
   ```

3. 복구 옵션 제공:
   ```text
   다음 조치를 선택하세요:
     1. Phase 3 재실행 (npm 재설치)
     2. Phase 4 재실행 (템플릿 재복사)
     3. 무시 (권장하지 않음)
   ```

**검증**:
```bash
# 버전 확인
grep "version:" .moai/memory/development-guide.md
npm list moai-adk --depth=0
# 예상: 버전 불일치
```

---

### 시나리오 9: 롤백 수행 (moai restore)

**Given**:
- 업데이트 후 문제가 발생했다
- 백업이 `.moai-backup/2025-10-06-15-30-00/`에 존재한다

**When**:
```bash
moai restore .moai-backup/2025-10-06-15-30-00
```

**Then**:
1. 백업 경로 검증:
   - `validateBackupPath()` 실행
   - 필수 항목 확인: `.claude/`, `.moai/`, `CLAUDE.md`
   - "✅ 백업 검증 통과" 출력

2. 복원 실행:
   - `.claude/` 복원
   - `.moai/` 복원
   - `CLAUDE.md` 복원
   - "✅ 백업 복원 완료" 출력

3. 복원 항목 표시:
   ```text
   복원 완료:
     - .claude/
     - .moai/
     - CLAUDE.md
   ```

**검증**:
```bash
# 파일 수정 시간 확인 (복원 시간)
stat -f "%Sm" CLAUDE.md
# 예상: 복원 실행 시간

# 버전 확인 (이전 버전으로 복원)
npm list moai-adk --depth=0
# 예상: v0.0.1 (롤백됨)
```

---

### 시나리오 10: Dry-run 복원 (미리보기)

**Given**:
- 백업이 `.moai-backup/2025-10-06-15-30-00/`에 존재한다

**When**:
```bash
moai restore .moai-backup/2025-10-06-15-30-00 --dry-run
```

**Then**:
1. 미리보기 모드 활성화:
   - "🔍 Dry run - would restore to: {현재 디렉토리}" 출력

2. 복원 예상 항목 표시:
   ```text
   Would restore: .moai-backup/.../. claude → ./. claude
   Would restore: .moai-backup/.../. moai → ./. moai
   Would restore: .moai-backup/.../CLAUDE.md → ./CLAUDE.md
   ```

3. 실제 복원 없음:
   - "✅ Dry run completed successfully" 출력
   - "Would restore 3 items" 출력

**검증**:
```bash
# 파일이 실제로 변경되지 않았는지 확인
stat -f "%Sm" CLAUDE.md
# 예상: 이전 수정 시간 (변경 없음)
```

---

### 시나리오 11: 품질 검증 (--check-quality 옵션)

**Given**:
- 업데이트가 완료되었다
- trust-checker 에이전트가 사용 가능하다

**When**:
```bash
/alfred:9-update --check-quality
```

**Then**:
1. Phase 1-5 정상 실행 (시나리오 1과 동일)

2. Phase 5.5 추가 실행:
   - `@agent-trust-checker "Level 1 빠른 스캔 (3-5초)"` 호출
   - TRUST 5원칙 검증:
     - **T**est: 테스트 커버리지 ≥85%
     - **R**eadable: ESLint/Biome 통과
     - **U**nified: TypeScript 타입 안전성
     - **S**ecured: npm audit 보안 취약점 없음
     - **T**rackable: @TAG 체인 무결성

3. 검증 결과 표시:
   ```text
   🔍 품질 검증 중... (trust-checker)
      ✅ 테스트 커버리지 88% (목표: 85%)
      ✅ 린트 검사 통과
      ✅ 타입 안전성 확인
      ✅ 보안 취약점 없음
      ✅ TAG 체인 무결성 확인
      ✅ 품질 검증 통과
   ```

**검증**:
```bash
# trust-checker 실행 로그 확인
# (trust-checker가 실제로 호출되었는지)

# 테스트 커버리지 확인
npm run test:coverage
# 예상: 85% 이상
```

---

### 시나리오 12: JSON 파싱 실패 - 안전 모드

**Given**:
- `.moai/config.json`이 손상되어 JSON 파싱 불가능

**When**:
```bash
/alfred:9-update
```

**Then**:
1. Phase 4 - Step 11 실행 중 오류:
   - JSON 파싱 시도 실패
   - 안전 모드 진입

2. 안전 모드 실행:
   ```bash
   [Bash] cp .moai/config.json .moai/config.json.backup
   ```
   - "⚠️ config.json 파싱 실패, 백업 생성: .moai/config.json.backup" 출력

3. 템플릿으로 교체:
   ```bash
   [Read] "{TEMPLATE_ROOT}/.moai/config.json"
   [Write] ".moai/config.json"
   ```
   - "📝 수동 복구 필요: .moai/config.json.backup 참조" 출력

**검증**:
```bash
# 백업 파일 존재 확인
ls .moai/config.json.backup
# 예상: 파일 존재

# 새 config.json 유효성 확인
jq . .moai/config.json > /dev/null
echo $?
# 예상: 0 (파싱 성공)

# 템플릿 상태 확인
grep "{{PROJECT_NAME}}" .moai/config.json
# 예상: 패턴 존재 (템플릿 상태)
```

---

## 품질 게이트 (Quality Gates)

### 1. 문서 품질 게이트

| 검증 항목 | 기준 | 도구 | 통과 조건 |
|-----------|------|------|-----------|
| YAML Front Matter | 필수 필드 7개 | 수동 검토 | 모든 필드 존재 |
| HISTORY 섹션 | v0.1.0 INITIAL | 수동 검토 | 항목 작성 완료 |
| EARS 구조 | 4개 섹션 | 수동 검토 | 모든 섹션 존재 |
| Markdown 린트 | 표준 준수 | markdownlint | 0 경고 |
| @TAG 체인 | 무결성 | `rg '@(SPEC\|CODE\|DOC):'` | 모든 TAG 연결 |

**검증 명령어**:
```bash
# Markdown 린트
npx markdownlint .moai/specs/SPEC-UPDATE-REFACTOR-001/

# @TAG 체인 검증
rg '@SPEC:UPDATE-REFACTOR-001' -n
rg '@CODE:CLI-005' -n
rg '@DOC:UPDATE-001' -n
```

### 2. 코드 품질 게이트

| 검증 항목 | 기준 | 도구 | 통과 조건 |
|-----------|------|------|-----------|
| 타입 안전성 | TypeScript strict | `tsc --noEmit` | 0 에러 |
| 린트 | ESLint/Biome | `npm run lint` | 0 에러 |
| 테스트 커버리지 | ≥85% | Jest/Vitest | 85% 이상 |
| 보안 취약점 | 0 Critical | `npm audit` | 0 Critical |

**검증 명령어**:
```bash
# 타입 체크
npm run type-check

# 린트
npm run lint

# 테스트 커버리지
npm run test:coverage

# 보안 스캔
npm audit --production
```

### 3. 기능 품질 게이트

| 검증 항목 | 기준 | 방법 | 통과 조건 |
|-----------|------|------|-----------|
| Phase 1-5 실행 | 정상 완료 | 수동 테스트 | 모든 Phase 성공 |
| 백업 생성 | 디렉토리 생성 | 파일 시스템 확인 | `.moai-backup/` 존재 |
| 템플릿 복사 | 파일 개수 | Glob 검증 | 예상 개수 일치 |
| 지능형 병합 | 정보 유지 | Grep 검증 | 프로젝트 정보 존재 |
| 롤백 기능 | 복원 성공 | `moai restore` 테스트 | 파일 복원 완료 |

**검증 명령어**:
```bash
# Phase 1-5 실행
/alfred:9-update

# 백업 확인
ls -d .moai-backup/*

# 템플릿 파일 개수
ls -1 .claude/commands/alfred/*.md | wc -l

# 프로젝트 정보 유지
grep "이름:" CLAUDE.md

# 롤백 테스트
moai restore .moai-backup/$(ls -t .moai-backup/ | head -1)
```

### 4. 사용자 경험 게이트

| 검증 항목 | 기준 | 방법 | 통과 조건 |
|-----------|------|------|-----------|
| 오류 메시지 명확성 | 이해 가능 | 사용자 테스트 | 복구 방법 안내 |
| 진행 상태 표시 | Phase별 로그 | 수동 확인 | 모든 Phase 로그 존재 |
| 롤백 안내 | 명령어 표시 | 출력 확인 | 명령어 정확성 |
| 실행 시간 | - | - | **시간 표시 금지** |

**검증 방법**:
- 오류 시나리오별 메시지 확인
- 진행 로그 완전성 확인
- 롤백 명령어 정확성 확인

---

## 성능 기준

### 실행 시간 기준 (참고용, 문서 표시 금지)

| Phase | 예상 시간 | 측정 방법 |
|-------|-----------|-----------|
| Phase 1 | 2-3초 | `npm` 명령어 2회 |
| Phase 2 | 1-2초 | 파일 복사 |
| Phase 3 | 10-30초 | npm 설치 |
| Phase 4 | 5-10초 | 템플릿 복사 (~30개 파일) |
| Phase 5 | 2-3초 | 검증 |
| Phase 5.5 | 3-5초 | trust-checker (선택) |
| **총계** | **20-53초** | - |

**주의**: 이 시간은 내부 참고용이며, 사용자 문서에는 절대 표시하지 않습니다.

### 백업 크기 기준

- **최대 백업 크기**: ~5MB (SPEC 파일 포함 시)
- **예상 백업 크기**: ~1.5MB (SPEC 파일 제외 시)
- **백업 개수 제한**: 10개 권장 (자동 정리 정책)

---

## 에러 핸들링 기준

### 복구 가능 에러

| 에러 타입 | 자동 복구 | 사용자 선택 | 수동 복구 |
|-----------|-----------|-------------|-----------|
| 파일 복사 실패 | 2회 재시도 | Phase 4 재실행 | 백업 복원 |
| 디렉토리 없음 | `mkdir -p` | - | - |
| 권한 오류 (Unix) | `chmod +x` | - | - |
| 권한 오류 (Windows) | 경고만 | - | - |
| JSON 파싱 실패 | 백업 + 템플릿 교체 | - | 수동 병합 |
| 버전 불일치 | - | Phase 3/4 재실행 | 백업 복원 |
| 파일 누락 | - | Phase 4 재실행 | 백업 복원 |

### 복구 불가능 에러

| 에러 타입 | 처리 방법 | 사용자 안내 |
|-----------|-----------|-------------|
| npm 설치 실패 | 중단 | npm 캐시 정리 후 재시도 |
| 디스크 공간 부족 | 중단 | 공간 확보 후 재시도 |
| 네트워크 오류 | 중단 | 연결 확인 후 재시도 |
| 백업 생성 실패 | 중단 (--force 제외) | 디스크 확인 후 재시도 |

---

## 크로스 플랫폼 기준

### 플랫폼별 테스트 매트릭스

| 플랫폼 | Node.js | 테스트 항목 | 예상 결과 |
|--------|---------|-------------|-----------|
| **macOS** | 14, 16, 18, 20 | Phase 1-5, 롤백 | 모두 통과 |
| **Linux** | 14, 16, 18, 20 | Phase 1-5, 롤백 | 모두 통과 |
| **Windows** | 14, 16, 18, 20 | Phase 1-5, 롤백 | chmod 경고, 나머지 통과 |

### 플랫폼별 예외 처리

**Windows**:
- `chmod` 실패 → 경고 출력, 계속 진행
- 경로 구분자 → `path.join()` 사용

**Unix 계열**:
- `chmod +x` 실행 → 755 권한 부여
- 실행 권한 검증 포함

---

## 회귀 테스트 체크리스트

### 기존 기능 보존 확인

- [ ] Phase 1: 버전 확인 로직 변경 없음
- [ ] Phase 2: 백업 디렉토리 구조 동일
- [ ] Phase 3: npm 설치 방식 동일
- [ ] Phase 4: 템플릿 복사 정책 동일
- [ ] Phase 5: 검증 항목 동일
- [ ] `moai restore`: 복원 로직 동일
- [ ] 백업 검증: 필수 항목 목록 동일

### 하위 호환성 확인

- [ ] 기존 백업 디렉토리 복원 가능
- [ ] 기존 config.json 병합 가능
- [ ] 기존 CLAUDE.md 병합 가능
- [ ] v0.0.1 → v0.0.2 업그레이드 성공

---

## 최종 검증 체크리스트

### SPEC 문서

- ✅ `spec.md`: YAML, HISTORY, EARS 완전성
- ✅ `plan.md`: 마일스톤, 기술적 접근 방법
- ✅ `acceptance.md`: 테스트 시나리오, 품질 게이트

### 테스트 커버리지

- [ ] 시나리오 1-12 모두 검증 완료
- [ ] 품질 게이트 4가지 모두 통과
- [ ] 크로스 플랫폼 테스트 완료

### 문서 품질

- [ ] Markdown 린트 통과
- [ ] @TAG 체인 무결성 확인
- [ ] 오타 및 문법 검토 완료

### 코드 품질

- [ ] TypeScript 타입 체크 통과
- [ ] ESLint/Biome 통과
- [ ] 테스트 커버리지 ≥85%
- [ ] npm audit 통과 (0 Critical)

---

## 배포 전 최종 확인

### Git 작업

- [ ] SPEC 파일 3개 커밋 완료
- [ ] 커밋 메시지 명확성 확인
- [ ] @TAG 체인 검증 완료

### 문서 동기화

- [ ] 9-update.md 최신 상태 확인
- [ ] restore.ts 주석 업데이트 확인
- [ ] development-guide.md 참조 정확성 확인

### 최종 테스트

- [ ] 시나리오 1 (Happy Path) 실행 성공
- [ ] 시나리오 9 (롤백) 실행 성공
- [ ] 크로스 플랫폼 빌드 성공

---

**작성자**: @Goos
**작성일**: 2025-10-06
**버전**: 0.1.0
**상태**: Draft
**다음 단계**: `/alfred:2-build UPDATE-REFACTOR-001` (필요 시)
