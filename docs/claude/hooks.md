---
title: 이벤트 훅
description: 9개 자동화 훅으로 개발 프로세스 보호
---

# 이벤트 훅

MoAI-ADK는 개발 프로세스를 자동으로 보호하고 가이드하는 **9개 이벤트 훅**을 제공합니다. TypeScript/JavaScript로 빌드되어 고성능으로 실행됩니다.

## 훅 개요

### 전체 훅 목록

| 훅 | 이벤트 | 주요 기능 | 자동 실행 |
|---|--------|-----------|-----------|
| **file-monitor** | 파일 변경 | 파일 모니터링 및 로깅 | ✅ |
| **language-detector** | 세션 시작 | 언어 자동 감지 | ✅ |
| **policy-block** | 명령 실행 전 | 정책 위반 차단 | ✅ |
| **pre-write-guard** | 파일 쓰기 전 | 파일 검증 및 백업 | ✅ |
| **session-notice** | 세션 시작 | 개발 가이드 안내 | ✅ |
| **steering-guard** | 주기적 | 방향성 가이드 | ✅ |
| **tag-enforcer** | 파일 쓰기 전 | TAG 규칙 강제 적용 | ✅ |

### 훅 위치

```
.claude/hooks/moai/
├── file-monitor.js
├── language-detector.js
├── policy-block.js
├── pre-write-guard.js
├── session-notice.js
├── steering-guard.js
├── tag-enforcer.js
└── package.json
```

## 1. file-monitor

### 목적

**파일 변경 모니터링 및 로깅**

코드 파일 변경을 실시간으로 추적하고 로그에 기록합니다.

### 동작 방식

```javascript
// .claude/hooks/moai/file-monitor.js

/**
 * @HOOK:FILE-MONITOR-001
 * 파일 변경 이벤트 모니터링
 */
export function onFileChange(event) {
  const { path, action } = event;

  // Winston logger로 기록
  logger.info('File changed', {
    path,
    action, // 'create' | 'modify' | 'delete'
    timestamp: new Date().toISOString(),
    user: process.env.USER
  });

  // TAG 파일 변경 감지
  if (isSourceFile(path)) {
    checkTagPresence(path);
  }

  // 민감 파일 변경 경고
  if (isSensitiveFile(path)) {
    logger.warn('Sensitive file modified', { path });
  }
}
```

### 모니터링 대상

#### 소스 코드 파일

```
감지 대상:
- src/**/*.ts
- src/**/*.js
- src/**/*.py
- src/**/*.java
- src/**/*.go
- __tests__/**/*

로그 내용:
- 파일 경로
- 변경 종류 (생성/수정/삭제)
- 타임스탬프
- TAG 존재 여부
```

#### 설정 파일

```
감지 대상:
- .moai/config.json
- .claude/settings.json
- package.json
- tsconfig.json

로그 내용:
- 변경 전후 diff
- 설정 검증 결과
```

### 로그 예시

```json
{
  "level": "info",
  "message": "File changed",
  "path": "src/auth/service.ts",
  "action": "modify",
  "hasTag": true,
  "tagId": "AUTH-001",
  "timestamp": "2024-01-15T14:30:45.123Z"
}
```

## 2. language-detector

### 목적

**프로젝트 언어 자동 감지**

세션 시작 시 프로젝트의 주 언어를 감지하고 적절한 도구를 추천합니다.

### 동작 방식

```javascript
/**
 * @HOOK:LANGUAGE-DETECTOR-001
 * 세션 시작 시 언어 자동 감지
 */
export function onSessionStart() {
  const languages = detectLanguages();

  logger.info('Languages detected', {
    languages,
    primary: languages[0]
  });

  // 도구 추천
  const tools = recommendTools(languages[0]);

  console.log(`
Detected Languages:
  - ${languages.map(l => `${l.name}: ${l.percentage}%`).join('\n  - ')}

Recommended Tools:
  - Test: ${tools.test}
  - Lint: ${tools.lint}
  - Format: ${tools.format}
  `);
}
```

### 감지 패턴

#### JavaScript/TypeScript

```
감지 파일:
- package.json
- tsconfig.json
- *.ts, *.tsx, *.js, *.jsx

도구 추천:
- Test: Vitest
- Lint: Biome
- Format: Biome
- Build: tsup
```

#### Python

```
감지 파일:
- requirements.txt
- pyproject.toml
- setup.py
- *.py

도구 추천:
- Test: pytest
- Lint: ruff
- Format: black
- Type: mypy
```

#### Java

```
감지 파일:
- pom.xml
- build.gradle
- *.java

도구 추천:
- Test: JUnit
- Build: Maven/Gradle
- Format: Google Java Format
```

### 출력 예시

```
✓ Language Detection
  - TypeScript: 65%
  - Python: 25%
  - Go: 10%

Primary Language: TypeScript

Recommended Tools:
  - Test Runner: Vitest
  - Linter: Biome
  - Formatter: Biome
  - Build Tool: tsup

Run 'moai doctor' for detailed diagnostics.
```

## 3. policy-block

### 목적

**정책 위반 차단**

위험한 명령어나 패턴을 자동으로 차단합니다.

### 동작 방식

```javascript
/**
 * @HOOK:POLICY-BLOCK-001
 * 정책 위반 명령어 차단
 */
export function onCommandExecute(command) {
  const blockedPatterns = [
    /rm\s+-rf/,           // 강제 삭제
    /sudo/,               // 관리자 권한
    /chmod\s+777/,        // 전체 권한
    /--force/,            // 강제 실행
    /DROP\s+TABLE/i       // SQL 삭제
  ];

  for (const pattern of blockedPatterns) {
    if (pattern.test(command)) {
      logger.warn('Blocked command', { command, pattern });

      throw new PolicyViolationError(
        `Command blocked by security policy: ${command}\n` +
        `Pattern: ${pattern}\n` +
        `If you need to run this, use git-manager with approval.`
      );
    }
  }
}
```

### 차단 대상

#### 파일 시스템

```
차단 명령어:
- rm -rf /
- rm -rf *
- sudo rm

허용 방법:
- 사용자 확인 필요
- @agent-git-manager로 요청
```

#### Git 작업

```
차단 명령어:
- git push --force
- git reset --hard
- git branch -D main

허용 방법:
- @agent-git-manager로 요청
- 사용자 명시적 승인
```

#### 데이터베이스

```
차단 명령어:
- DROP TABLE
- TRUNCATE TABLE
- DELETE without WHERE

허용 방법:
- 마이그레이션 스크립트 사용
- 사용자 승인
```

### 차단 예시

```
❌ Command Blocked

Command: rm -rf node_modules/
Reason: Potentially destructive operation
Policy: File deletion requires confirmation

Recommendation:
  Use: @agent-git-manager "clean node_modules"

This ensures:
  ✓ Backup before deletion
  ✓ User confirmation
  ✓ Audit logging
```

## 4. pre-write-guard

### 목적

**파일 쓰기 전 검증 및 백업**

파일을 쓰기 전에 검증하고 자동 백업을 생성합니다.

### 동작 방식

```javascript
/**
 * @HOOK:PRE-WRITE-GUARD-001
 * 파일 쓰기 전 검증
 */
export async function onBeforeWrite(file) {
  // 1. 파일 존재 확인
  if (existsSync(file.path)) {
    // 백업 생성
    await createBackup(file.path);
    logger.info('Backup created', { path: file.path });
  }

  // 2. TAG BLOCK 확인 (소스 파일)
  if (isSourceFile(file.path) && !hasTagBlock(file.content)) {
    logger.warn('Missing TAG BLOCK', { path: file.path });

    const suggestion = generateTagBlock(file.path);
    console.warn(`
⚠️  Missing TAG BLOCK in ${file.path}

Suggested TAG BLOCK:
${suggestion}

Add this to the top of your file.
    `);
  }

  // 3. TRUST 원칙 검증
  const violations = checkTrustViolations(file.content);
  if (violations.length > 0) {
    logger.warn('TRUST violations detected', { path: file.path, violations });

    console.warn(`
⚠️  TRUST Violations in ${file.path}

${violations.map(v => `- ${v.rule}: ${v.message}`).join('\n')}

Consider fixing these before committing.
    `);
  }

  // 4. 민감정보 검사
  const secrets = detectSecrets(file.content);
  if (secrets.length > 0) {
    throw new SecurityError(
      `Sensitive information detected in ${file.path}\n` +
      `Found: ${secrets.join(', ')}\n` +
      `Use environment variables instead.`
    );
  }
}
```

### 검증 항목

#### TAG BLOCK 존재

```typescript
// ✅ Good
// @CODE:AUTH-001 | Chain: @REQ → @DESIGN → @TASK → @TEST
export class AuthService {
  // ...
}

// ❌ Bad (경고 발생)
export class AuthService {
  // TAG BLOCK 없음
}
```

#### TRUST 원칙

```typescript
// ✅ Good: 함수 ≤50 LOC
function authenticate(email, password) {
  // 45 LOC
}

// ❌ Bad: 함수 >50 LOC (경고)
function processOrder(order) {
  // 65 LOC - 리팩토링 필요
}
```

#### 민감정보

```typescript
// ❌ Bad: 하드코딩된 비밀 (차단)
const apiKey = 'sk_live_abc123';
const password = 'mypassword';

// ✅ Good: 환경 변수
const apiKey = process.env.API_KEY;
const password = process.env.DB_PASSWORD;
```

### 백업 시스템

```
백업 위치: .moai/backups/
백업 형식: {filename}.{timestamp}.bak

예시:
  service.ts
  → .moai/backups/service.ts.20240115-143045.bak

보관 기간: 7일
자동 정리: 매일 자정
```

## 5. session-notice

### 목적

**세션 시작 안내**

Claude Code 세션 시작 시 개발 가이드를 표시합니다.

### 동작 방식

```javascript
/**
 * @HOOK:SESSION-NOTICE-001
 * 세션 시작 알림
 */
export function onSessionStart() {
  const projectName = readProjectName();
  const mode = readConfig().mode;

  console.log(`
🗿 MoAI-ADK v0.0.1

Project: ${projectName}
Mode: ${mode}

3-Stage Workflow:
  /moai:1-spec  → SPEC 작성
  /moai:2-build → TDD 구현
  /moai:3-sync  → 문서 동기화

On-Demand Support:
  @agent-debug-helper "오류내용"
  @agent-tag-agent "TAG 검증"

TRUST 5 Principles:
  ✓ Test First
  ✓ Readable
  ✓ Unified
  ✓ Secured
  ✓ Trackable

Run 'moai doctor' for diagnostics.
  `);
}
```

### 표시 내용

```
프로젝트 정보:
- 프로젝트명
- 모드 (Personal/Team)
- 버전

핵심 명령어:
- 3단계 워크플로우
- 에이전트 호출 방법
- TRUST 5원칙

다음 단계:
- 첫 SPEC 작성
- 도움말 확인
```

## 6. steering-guard

### 목적

**개발 방향성 가이드**

주기적으로 개발 방향을 체크하고 가이드합니다.

### 동작 방식

```javascript
/**
 * @HOOK:STEERING-GUARD-001
 * 개발 방향성 가이드 (30분마다)
 */
export function onPeriodic() {
  const status = analyzeProjectStatus();

  // 장시간 Draft SPEC 경고
  if (status.draftSpecs.length > 0) {
    const oldDrafts = status.draftSpecs.filter(s =>
      isOlderThan(s.createdAt, '2 hours')
    );

    if (oldDrafts.length > 0) {
      console.warn(`
⚠️  Draft SPECs Pending

${oldDrafts.map(s => `- ${s.id}: ${s.title} (${s.age})`).join('\n')}

Next Step:
  /moai:2-build ${oldDrafts[0].id}
      `);
    }
  }

  // 구현 완료, 동기화 필요
  if (status.needsSync) {
    console.info(`
ℹ️  Documentation Sync Needed

${status.modifiedFiles.length} files modified
${status.newTags.length} new TAGs detected

Next Step:
  /moai:3-sync
    `);
  }

  // 테스트 실패 경고
  if (status.failingTests > 0) {
    console.warn(`
⚠️  ${status.failingTests} Test(s) Failing

Run tests:
  npm test

Debug:
  @agent-debug-helper "테스트 실패 분석"
    `);
  }
}
```

### 체크 항목

- Draft SPEC 방치 여부
- 문서 동기화 필요 여부
- 테스트 실패 상태
- TRUST 준수율 하락
- TAG 체인 불완전

## 7. tag-enforcer

### 목적

**TAG 규칙 강제 적용**

신규 소스 파일 생성 시 TAG BLOCK 존재를 강제하고, TAG 체인 무결성을 보장합니다.

### 동작 방식

```javascript
/**
 * @HOOK:TAG-ENFORCER-001
 * TAG 규칙 강제 적용
 */
export async function onBeforeWrite(file) {
  // 소스 파일만 검증
  if (!isSourceFile(file.path)) {
    return;
  }

  // TAG BLOCK 필수 확인
  if (!hasTagBlock(file.content)) {
    throw new ValidationError(
      `TAG BLOCK is required in source files.\n` +
      `File: ${file.path}\n\n` +
      `Add a TAG BLOCK at the top of the file:\n` +
      `// @CODE:<DOMAIN-ID> | Chain: @REQ → @DESIGN → @TASK → @TEST\n` +
      `// Related: @CODE:<ID>, @CODE:<ID>, @CODE:<ID>`
    );
  }

  // TAG 형식 검증
  const tags = extractTags(file.content);
  for (const tag of tags) {
    if (!isValidTagFormat(tag)) {
      throw new ValidationError(
        `Invalid TAG format: ${tag}\n` +
        `Expected format: @CATEGORY:DOMAIN-NNN\n` +
        `Example: @SPEC:AUTH-001`
      );
    }
  }

  // TAG 체인 완결성 확인
  const chainStatus = validatePrimaryChain(tags);
  if (!chainStatus.complete) {
    logger.warn('Incomplete TAG chain', {
      path: file.path,
      missing: chainStatus.missingTags
    });

    console.warn(`
⚠️  Incomplete TAG Chain in ${file.path}

Missing TAGs:
${chainStatus.missingTags.map(t => `  - ${t}`).join('\n')}

Complete your chain: @REQ → @DESIGN → @TASK → @TEST
    `);
  }
}
```

### 검증 규칙

#### TAG BLOCK 필수

```typescript
// ✅ Good: TAG BLOCK 존재
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
// Related: @CODE:AUTH-001:API, @CODE:AUTH-001:INFRA
export class AuthService {
  // ...
}

// ❌ Bad: TAG BLOCK 없음 (에러)
export class AuthService {
  // TAG가 없어서 에러 발생
}
```

#### TAG 형식 검증

```typescript
// ✅ Good: 올바른 형식
@SPEC:AUTH-001

@CODE:AUTH-001

// ❌ Bad: 잘못된 형식 (에러)
@REQ-AUTH-001      // 잘못된 구분자
@AUTH-001          // 카테고리 누락
@SPEC:AUTH001       // 하이픈 누락
```

#### TAG 체인 완결성

```typescript
// ✅ Good: 완전한 체인
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001

// ⚠️ Warning: 불완전한 체인
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001
// Missing: @TEST:AUTH-001
```

### 에러 예시

```
❌ TAG BLOCK Required

File: src/payment/service.ts
Error: TAG BLOCK is required in source files

Add a TAG BLOCK at the top of the file:
// @CODE:PAYMENT-001 | Chain: @SPEC:PAYMENT-001 →  → @CODE:PAYMENT-001 → @TEST:PAYMENT-001
// Related: @CODE:PAYMENT-001:API, @CODE:PAYMENT-001:DATA

This ensures full traceability from requirements to tests.
```

### 경고 예시

```
⚠️  Incomplete TAG Chain

File: src/auth/service.ts

Missing TAGs:
  - @TEST:AUTH-001

Complete your chain: @REQ → @DESIGN → @TASK → @TEST

Use '@agent-tag-agent' to verify chain integrity.
```

## 훅 커스터마이징

### 훅 활성화/비활성화

```json
// .claude/settings.json
{
  "hooks": {
    "enabled": true,
    "individual": {
      "file-monitor": true,
      "language-detector": true,
      "policy-block": true,
      "pre-write-guard": true,
      "session-notice": true,
      "steering-guard": false,    // 비활성화
      "tag-enforcer": true
    }
  }
}
```

### 커스텀 훅 추가

```javascript
// .claude/hooks/moai/custom-hook.js

/**
 * @HOOK:CUSTOM-001
 * 커스텀 훅 예시
 */
export function onCustomEvent(data) {
  // 커스텀 로직
  logger.info('Custom hook triggered', { data });
}
```

### 훅 순서 조정

```json
// .claude/settings.json
{
  "hooks": {
    "order": [
      "session-notice",       // 가장 먼저
      "language-detector",
      "policy-block",
      "tag-enforcer",         // TAG 검증
      "pre-write-guard",
      "file-monitor",
      "steering-guard"        // 가장 나중
    ]
  }
}
```

## 트러블슈팅

### 훅이 실행되지 않음

```bash
# 1. 훅 활성화 확인
cat .claude/settings.json | jq '.hooks.enabled'

# 2. 훅 파일 존재 확인
ls -la .claude/hooks/moai/

# 3. 권한 확인
chmod +x .claude/hooks/moai/*.js

# 4. Claude Code 재시작
```

### 훅 오류 디버깅

```bash
# 훅 로그 확인
cat .moai/logs/hooks.log

# 특정 훅 테스트
node .claude/hooks/moai/pre-write-guard.js
```

### TAG 검증 오류

```bash
# TAG 스캔 및 검증
@agent-tag-agent "코드 전체 스캔하여 TAG 검증"

# TAG 형식 확인
@agent-tag-agent "TAG 형식 검증"

# 불완전한 체인 확인
@agent-tag-agent "TAG 체인 무결성 검사"
```

## 다음 단계

### 에이전트 활용

- **[에이전트 가이드](/claude/agents)**: 8개 에이전트
- **[명령어](/claude/commands)**: 워크플로우 명령어

### 고급 설정

- **[커스텀 에이전트](/advanced/custom-agents)**: 에이전트 생성
- **[설정 파일](/reference/configuration)**: 전체 설정 옵션

## 참고 자료

- **훅 소스**: `.claude/hooks/moai/`
- **설정 파일**: `.claude/settings.json`
- **로그 파일**: `.moai/logs/hooks.log`