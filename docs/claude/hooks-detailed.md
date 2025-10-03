---
title: MoAI-ADK 훅 시스템 상세 가이드
description: Claude Code 이벤트 훅의 심층 분석, 아키텍처, 보안, 커스터마이징
---

# MoAI-ADK 훅 시스템 상세 가이드

MoAI-ADK의 훅 시스템은 Claude Code의 이벤트 기반 아키텍처를 활용하여 개발 프로세스를 자동으로 보호하고 가이드합니다. TypeScript/JavaScript로 빌드된 7개의 훅이 고성능으로 실행되며, 개발 가이드 준수, TAG 무결성, 보안 정책을 자동으로 검증합니다.

## 개요

### 훅 시스템의 목적과 철학

MoAI-ADK 훅 시스템은 "개발 프로세스 자동화"라는 핵심 철학을 구현합니다. 개발자가 수동으로 확인해야 할 사항들을 훅이 자동으로 검증하여 인적 오류를 줄이고 일관된 품질을 보장합니다.

**핵심 원칙:**
- **비침투적 자동화**: 개발 흐름을 방해하지 않으면서 백그라운드에서 작동
- **실시간 검증**: 파일 쓰기, 명령 실행 등의 이벤트 발생 시점에 즉시 검증
- **점진적 가이드**: 치명적 오류는 차단하고, 개선 사항은 경고로 제안

훅 시스템은 Claude Code의 생명주기 이벤트에 반응하여 실행됩니다. 세션 시작 시 프로젝트 상태를 분석하고, 파일 변경 시 TAG 무결성을 검사하며, 위험한 명령어 실행을 사전에 차단합니다.

### 이벤트 기반 아키텍처

MoAI-ADK는 Claude Code의 이벤트 기반 아키텍처를 활용합니다. 각 훅은 특정 이벤트(세션 시작, 프롬프트 제출, 도구 사용, 파일 쓰기)에 등록되어 해당 이벤트 발생 시 자동으로 실행됩니다.

이벤트 기반 설계의 장점은 명확합니다:
1. **낮은 결합도**: 각 훅은 독립적으로 작동하며 서로 의존하지 않음
2. **확장성**: 새로운 훅을 추가해도 기존 훅에 영향을 주지 않음
3. **성능**: 필요한 시점에만 실행되어 불필요한 오버헤드 없음

훅은 stdin으로 JSON 형식의 입력을 받고, stdout으로 결과를 반환합니다. 이러한 표준 인터페이스 덕분에 어떤 언어로든 훅을 작성할 수 있지만, MoAI-ADK는 성능과 타입 안전성을 위해 TypeScript를 선택했습니다.

### 보안 및 성능 고려사항

보안은 훅 시스템 설계의 핵심입니다. 각 훅은 명시적으로 허용된 작업만 수행하며, 민감한 파일이나 위험한 명령어는 사전에 차단합니다.

**보안 계층:**
- **입력 검증**: 모든 훅은 입력 데이터를 검증하고 악의적인 페이로드를 거부
- **경로 제한**: `.moai/memory/`, `.env` 등 보호된 경로는 편집 불가
- **명령어 필터링**: `rm -rf`, `sudo` 등 위험한 명령어 패턴 차단

성능 최적화도 중요합니다. 훅은 개발 흐름을 방해하지 않도록 빠르게 실행되어야 합니다. MoAI-ADK 훅은 다음 최적화 기법을 사용합니다:
- **조기 종료**: 처리 대상이 아닌 경우 즉시 반환
- **파일 시스템 캐싱**: 반복적인 파일 읽기 최소화
- **타임아웃 설정**: 10초 내 응답하지 않으면 자동 중단

## 훅 생명주기

### 훅 등록 및 실행 순서

Claude Code가 시작되면 `.claude/settings.json`에 정의된 훅들이 순서대로 등록됩니다. 각 훅은 특정 이벤트 타입에 바인딩되며, 이벤트 발생 시 등록된 순서대로 실행됩니다.

**등록 프로세스:**
1. Claude Code가 프로젝트 루트의 `.claude/settings.json` 읽기
2. `hooks` 배열에 정의된 각 훅의 경로 확인
3. 훅 파일이 실행 가능한지 검증 (실행 권한, 파일 존재)
4. 이벤트 타입별로 훅을 내부 맵에 등록

훅 실행 순서는 중요합니다. 예를 들어, `steering-guard`는 위험한 프롬프트를 차단하기 위해 가장 먼저 실행되고, `pre-write-guard`는 파일 쓰기 직전에 실행되어 TAG 무결성을 검사합니다.

**실행 순서 예시 (파일 쓰기 시):**
```
1. user-prompt-submit-hook (프롬프트 검증)
2. steering-guard (정책 위반 체크)
3. tool-use-block-hook (도구 사용 모니터링)
4. pre-write-guard (파일 안전성 검사)
5. tag-enforcer (@IMMUTABLE TAG 검증)
```

### 컨텍스트 전달

훅 간 데이터 공유는 직접적으로 이루어지지 않습니다. 각 훅은 독립적으로 실행되며, 필요한 데이터는 입력 JSON 또는 프로젝트 파일 시스템에서 읽어옵니다.

**입력 컨텍스트 구조:**
```json
{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.ts",
    "content": "file content..."
  },
  "prompt": "user prompt text",
  "session_id": "unique-session-id"
}
```

각 훅은 이 컨텍스트에서 필요한 정보를 추출하여 처리합니다. 예를 들어, `pre-write-guard`는 `tool_name`이 "Write"인지 확인하고, `file_path`에서 파일 경로를 추출하여 안전성을 검사합니다.

훅은 실행 결과를 JSON 형식으로 반환하며, 이는 다음 훅의 입력으로 전달되지 않습니다. 대신 Claude Code가 결과를 수집하여 최종 판단을 내립니다.

### 에러 처리

훅 실행 중 에러가 발생하면 어떻게 될까요? MoAI-ADK는 안전한 기본 동작(fail-safe)을 따릅니다.

**에러 처리 전략:**
1. **비치명적 에러**: 경고 로그를 출력하고 계속 진행 (예: 파일 읽기 실패)
2. **치명적 에러**: 작업을 차단하고 에러 메시지 출력 (예: @IMMUTABLE TAG 수정)
3. **타임아웃**: 10초 내 응답하지 않으면 훅을 중단하고 계속 진행

```javascript
try {
  const result = await hook.execute(input);
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message}`);
    process.exit(2); // 작업 차단
  }
} catch (error) {
  console.error(`Hook error: ${error.message}`);
  // 에러를 로그하지만 계속 진행
}
```

에러 발생 시 `.moai/logs/hooks.log`에 상세한 정보가 기록되어 디버깅에 활용할 수 있습니다.

### 훅 생명주기 다이어그램

```mermaid
sequenceDiagram
    participant User as 사용자
    participant CC as Claude Code
    participant HM as Hook Manager
    participant SG as steering-guard
    participant PW as pre-write-guard
    participant TE as tag-enforcer
    participant FS as File System

    User->>CC: 파일 쓰기 명령
    CC->>HM: Write 이벤트 발생

    HM->>SG: execute(input)
    activate SG
    SG->>SG: 프롬프트 검증
    SG-->>HM: {success: true}
    deactivate SG

    HM->>PW: execute(input)
    activate PW
    PW->>FS: 기존 파일 읽기
    FS-->>PW: 파일 내용
    PW->>PW: 민감 파일 체크
    PW-->>HM: {success: true}
    deactivate PW

    HM->>TE: execute(input)
    activate TE
    TE->>FS: 기존 파일 읽기
    FS-->>TE: 파일 내용
    TE->>TE: "@IMMUTABLE 검증"
    alt TAG 수정 감지
        TE-->>HM: {success: false, blocked: true}
        HM-->>CC: 작업 차단
        CC-->>User: 에러 메시지
    else TAG 정상
        TE-->>HM: {success: true}
        deactivate TE
        HM-->>CC: 모든 훅 통과
        CC->>FS: 파일 쓰기
        FS-->>CC: 완료
        CC-->>User: 성공
    end
```

## 개별 훅 상세 분석

### session-notice: 세션 초기화

**목적:** Claude Code 세션 시작 시 프로젝트 상태를 분석하고 개발자에게 유용한 정보를 제공합니다.

**트리거 조건:**
- Claude Code 세션이 시작될 때 (`session-start` 이벤트)
- MoAI 프로젝트가 감지되었을 때 (`.moai` 디렉토리 존재)
- 세션당 한 번만 실행 (중복 실행 방지)

**파라미터:**
```typescript
interface SessionNoticeInput {
  session_id: string;
  project_root: string;
  environment: {
    shell: string;
    platform: string;
  };
}
```

**실행 로직:**
```javascript
async execute(input) {
  // 1. MoAI 프로젝트 여부 확인
  if (!this.isMoAIProject()) {
    return {
      success: true,
      message: "💡 Run `/alfred:8-project` to initialize MoAI-ADK"
    };
  }

  // 2. 프로젝트 상태 수집
  const status = await this.getProjectStatus();

  // 3. Git 정보 조회
  const gitInfo = await this.getGitInfo();

  // 4. SPEC 진행률 계산
  const specProgress = this.getSpecProgress();

  // 5. 개발 가이드 준수 여부 확인
  const constitutionStatus = this.checkConstitutionStatus();

  // 6. 사용자 친화적 출력 생성
  const output = await this.generateSessionOutput({
    projectName: status.projectName,
    gitInfo,
    specProgress,
    constitutionStatus
  });

  return {
    success: true,
    message: output,
    data: status
  };
}
```

**반환값 형식:**
```javascript
{
  success: true,
  message: `
🗿 MoAI-ADK 프로젝트: MoAI-ADK
🌿 현재 브랜치: develop (d0262d8 📚 DOCS: 전체 문서 동기화 완료...)
📝 변경사항: 15개 파일
📝 SPEC 진행률: 0/0 (미완료 0개)
✅ 통합 체크포인트 시스템 사용 가능
  `,
  data: {
    projectName: "MoAI-ADK",
    moaiVersion: "0.0.1",
    initialized: true,
    // ...
  }
}
```

**커스터마이징 방법:**

출력 메시지를 프로젝트 요구사항에 맞게 수정할 수 있습니다:
```javascript
// .claude/hooks/alfred/session-notice.js 수정

generateSessionOutput(status) {
  const lines = [];

  // 커스텀 로고 추가
  lines.push('🚀 Your Custom Logo');

  // 프로젝트별 메시지
  if (status.projectName === 'critical-project') {
    lines.push('⚠️  프로덕션 환경입니다. 신중하게 작업하세요!');
  }

  // 기존 정보 표시
  lines.push(`📦 프로젝트: ${status.projectName}`);

  return lines.join('\n');
}
```

**주의사항:**
- Git 명령어 실행에 2초 타임아웃이 설정되어 있어, Git이 느리면 정보가 표시되지 않을 수 있음
- `.moai/config.json`이 손상되면 기본값으로 동작
- 세션 알림 파일(`/tmp/moai_session_notified`)이 존재하면 메시지가 표시되지 않음

### language-detector: 언어 감지

**목적:** 프로젝트에서 사용하는 프로그래밍 언어를 자동으로 감지하고 적절한 개발 도구를 추천합니다.

**트리거 조건:**
- 세션 시작 시 자동 실행
- 프로젝트 루트에 언어 관련 파일이 존재할 때 (package.json, pyproject.toml 등)
- `moai doctor` 명령어 실행 시

**파라미터:**
입력 파라미터는 필요하지 않으며, 프로젝트 루트에서 파일 시스템을 스캔합니다.

**감지 알고리즘:**
```javascript
detectProjectLanguages() {
  const languages = [];

  // 1. 설정 파일 기반 감지 (높은 정확도)
  if (this.hasFile('package.json')) {
    languages.push('javascript');
  }
  if (this.hasFile('tsconfig.json')) {
    languages.push('typescript');
  }
  if (this.hasFile('pyproject.toml')) {
    languages.push('python');
  }

  // 2. 파일 확장자 기반 감지 (폴백)
  if (this.hasFiles('**/*.py')) {
    languages.push('python');
  }
  if (this.hasFiles('**/*.{ts,tsx}')) {
    languages.push('typescript');
  }

  // 3. 중복 제거 및 반환
  return Array.from(new Set(languages));
}
```

**도구 매핑:**
```javascript
const DEFAULT_MAPPINGS = {
  test_runners: {
    python: "pytest",
    typescript: "npm test",
    go: "go test ./...",
    rust: "cargo test"
  },
  linters: {
    python: "ruff",
    typescript: "eslint",
    go: "golangci-lint",
    rust: "cargo clippy"
  },
  formatters: {
    python: "black",
    typescript: "prettier",
    go: "gofmt",
    rust: "rustfmt"
  }
};
```

**실제 코드 예시:**

`.moai/config/language_mappings.json`을 생성하여 커스텀 도구를 지정할 수 있습니다:
```json
{
  "test_runners": {
    "typescript": "vitest",
    "python": "pytest --cov"
  },
  "linters": {
    "typescript": "biome check",
    "python": "ruff check"
  },
  "formatters": {
    "typescript": "biome format",
    "python": "ruff format"
  }
}
```

**출력 예시:**
```
🌐 감지된 언어: typescript, python
🔧 권장 도구:
- typescript: test=npm test, lint=eslint, format=prettier
- python: test=pytest, lint=ruff, format=black
💡 필요 시 /alfred:2-build 단계에서 해당 도구를 사용해 TDD를 실행하세요.
```

**커스터마이징 방법:**

새로운 언어를 추가하려면 `language-detector.js`를 직접 수정:
```javascript
// Kotlin 지원 추가
detectProjectLanguages() {
  // ...
  if (this.hasFile('build.gradle.kts') || this.hasFiles('**/*.kt')) {
    languages.push('kotlin');
  }
  // ...
}

// Kotlin 도구 매핑
DEFAULT_MAPPINGS.test_runners.kotlin = "gradle test";
DEFAULT_MAPPINGS.linters.kotlin = "ktlint";
DEFAULT_MAPPINGS.formatters.kotlin = "ktlint format";
```

**주의사항:**
- `node_modules`, `.git` 등 제외 디렉토리는 스캔하지 않음
- 파일 스캔은 재귀적으로 이루어지므로 대형 프로젝트에서는 약간의 시간이 걸릴 수 있음
- 언어 감지는 휴리스틱 기반이므로 100% 정확하지 않을 수 있음 (confidence: 0.85)

### steering-guard: 방향성 가이드

**목적:** 사용자 프롬프트에서 위험한 패턴을 감지하고 개발 가이드 위반을 사전에 차단합니다.

**트리거 조건:**
- 사용자가 프롬프트를 제출할 때마다 (`user-prompt-submit` 이벤트)
- MoAI 프로젝트에서만 활성화
- 세션 시작 시 한 번 안내 메시지 출력

**파라미터:**
```typescript
interface SteeringGuardInput {
  prompt: string;
  session_id: string;
  tool_name?: string;
}
```

**금지 패턴:**
```javascript
const BANNED_PATTERNS = [
  {
    pattern: /ignore (the )?(claude|constitution|steering|instructions)/i,
    message: "헌법/지침 무시는 허용되지 않습니다.",
    severity: "critical"
  },
  {
    pattern: /disable (all )?(hooks?|guards?|polic(y|ies))/i,
    message: "Hook/Guard 해제 요청은 차단되었습니다.",
    severity: "critical"
  },
  {
    pattern: /rm -rf/i,
    message: "위험한 셸 명령을 프롬프트로 제출할 수 없습니다.",
    severity: "high"
  },
  {
    pattern: /drop (all )?safeguards/i,
    message: "안전장치 제거 요청은 거부됩니다.",
    severity: "critical"
  }
];
```

**실행 로직:**
```javascript
async execute(input) {
  // 1. 세션 안내 표시 (첫 실행 시)
  this.showSessionNotice();

  // 2. 프롬프트 추출 및 검증
  const prompt = input.prompt;
  if (!prompt || typeof prompt !== 'string') {
    return { success: true };
  }

  // 3. 금지 패턴 검사
  for (const { pattern, message, severity } of BANNED_PATTERNS) {
    if (pattern.test(prompt)) {
      return {
        success: false,
        blocked: true,
        message: message,
        exitCode: 2
      };
    }
  }

  // 4. 정상 통과
  return {
    success: true,
    message: "Steering Guard: 개발 가이드와 TAG 규칙을 준수하며 작업을 진행합니다."
  };
}
```

**세션 안내 메시지:**
```
🚀 MoAI-ADK 하이브리드 프로젝트가 감지되었습니다!
📖 개발 가이드: CLAUDE.md | TRUST 원칙: .moai/memory/development-guide.md
⚡ 하이브리드 워크플로우: /alfred:1-spec → /alfred:2-build → /alfred:3-sync
🔗 시스템 상태: TypeScript (브릿지 없음) ⚠️
🔧 디버깅: /alfred:4-debug | 설정 관리: @agent-cc-manager
```

**커스터마이징 방법:**

프로젝트별 금지 패턴을 추가할 수 있습니다:
```javascript
// .claude/hooks/alfred/steering-guard.js 수정

const PROJECT_SPECIFIC_PATTERNS = [
  {
    pattern: /production|prod|main branch/i,
    message: "프로덕션 환경 작업은 승인이 필요합니다.",
    severity: "high"
  },
  {
    pattern: /delete (all |entire )?database/i,
    message: "데이터베이스 삭제 명령은 차단됩니다.",
    severity: "critical"
  }
];

// BANNED_PATTERNS와 병합
const ALL_PATTERNS = [...BANNED_PATTERNS, ...PROJECT_SPECIFIC_PATTERNS];
```

**주의사항:**
- 세션 안내는 `/tmp/moai_session_notified` 파일이 없을 때만 표시됨
- 프롬프트 전체를 검사하므로 정상적인 문장에도 false positive 발생 가능
- 차단된 프롬프트는 로그에 기록되지 않음 (보안상 이유)

### policy-block: 정책 위반 차단

**목적:** 위험한 Bash 명령어 실행을 사전에 차단하여 시스템 보안과 데이터 무결성을 보호합니다.

**트리거 조건:**
- Claude Code가 Bash 도구를 사용하려고 할 때 (`tool-use-block` 이벤트)
- `tool_name`이 "Bash"일 때만 실행
- 모든 Bash 명령어에 대해 검사

**파라미터:**
```typescript
interface PolicyBlockInput {
  tool_name: "Bash";
  tool_input: {
    command: string | string[];
  };
}
```

**위험 명령어 목록:**
```javascript
const DANGEROUS_COMMANDS = [
  "rm -rf /",                    // 루트 파일 시스템 삭제
  "rm -rf --no-preserve-root",   // 강제 루트 삭제
  "sudo rm",                      // 관리자 권한 삭제
  "dd if=/dev/zero",              // 디스크 덮어쓰기
  ":(){:|:&};:",                  // 포크 폭탄
  "mkfs."                         // 파일 시스템 포맷
];

const ALLOWED_PREFIXES = [
  "git ", "python", "pytest",
  "npm ", "node ", "go ", "cargo ",
  "poetry ", "pnpm ", "rg ",
  "ls ", "cat ", "echo ", "which ",
  "make ", "moai "
];
```

**실행 로직:**
```javascript
async execute(input) {
  // 1. Bash 명령어인지 확인
  if (input.tool_name !== 'Bash') {
    return { success: true };
  }

  // 2. 명령어 추출
  const command = this.extractCommand(input.tool_input || {});
  if (!command) {
    return { success: true };
  }

  // 3. 위험 명령어 검사 (소문자 변환 후 검사)
  const commandLower = command.toLowerCase();
  for (const dangerousCommand of DANGEROUS_COMMANDS) {
    if (commandLower.includes(dangerousCommand)) {
      return {
        success: false,
        blocked: true,
        message: `위험 명령이 감지되었습니다 (${dangerousCommand}).`,
        exitCode: 2
      };
    }
  }

  // 4. 허용 목록 확인 (경고만 출력)
  if (!this.isAllowedPrefix(command)) {
    console.error(
      'NOTICE: 등록되지 않은 명령입니다. 필요 시 settings.json 의 allow 목록을 갱신하세요.'
    );
  }

  // 5. 정상 통과
  return { success: true };
}
```

**명령어 추출:**
```javascript
extractCommand(toolInput) {
  const raw = toolInput['command'] || toolInput['cmd'];

  // 배열 형식 처리 (예: ["npm", "test"])
  if (Array.isArray(raw)) {
    return raw.map(String).join(' ');
  }

  // 문자열 형식 처리
  if (typeof raw === 'string') {
    return raw.trim();
  }

  return null;
}
```

**실제 코드 예시:**

`.claude/settings.json`에서 허용 명령어를 확장할 수 있습니다:
```json
{
  "hooks": {
    "security": {
      "allowedCommands": [
        "git", "npm", "node", "python",
        "docker", "kubectl", "terraform"
      ],
      "blockedPatterns": [
        "rm -rf", "sudo", "chmod 777",
        "DROP TABLE", "TRUNCATE"
      ]
    }
  }
}
```

**커스터마이징 방법:**

프로젝트별 위험 명령어를 추가:
```javascript
// .claude/hooks/alfred/policy-block.js 수정

const PROJECT_SPECIFIC_DANGEROUS = [
  "curl | bash",            // 원격 스크립트 실행
  "wget -O- | sh",          // 원격 스크립트 실행
  "eval $(aws ",            // AWS 자격증명 평가
  "> /etc/",                // 시스템 파일 덮어쓰기
];

const ALL_DANGEROUS = [...DANGEROUS_COMMANDS, ...PROJECT_SPECIFIC_DANGEROUS];
```

**주의사항:**
- 명령어는 소문자로 변환 후 검사하므로 대소문자 우회 불가
- 부분 문자열 매칭을 사용하므로 정상 명령어에도 오탐 가능 (예: "informed" → "rm" 포함)
- 차단된 명령어는 실행되지 않으며, 로그에만 기록됨

### pre-write-guard: 파일 쓰기 보안

**목적:** 파일 쓰기 전에 민감한 파일이나 보호된 경로를 검증하여 실수로 인한 손상을 방지합니다.

**트리거 조건:**
- Claude Code가 파일을 쓰려고 할 때 (`pre-write` 이벤트)
- `tool_name`이 "Write", "Edit", "MultiEdit"일 때
- 모든 파일 쓰기 작업에 대해 실행

**파라미터:**
```typescript
interface PreWriteGuardInput {
  tool_name: "Write" | "Edit" | "MultiEdit";
  tool_input: {
    file_path?: string;
    filePath?: string;
    path?: string;
    content?: string;
    new_string?: string;
  };
}
```

**민감 키워드 및 보호 경로:**
```javascript
const SENSITIVE_KEYWORDS = [
  ".env",           // 환경 변수 파일
  "/secrets",       // 비밀 디렉토리
  "/.git/",         // Git 내부 파일
  "/.ssh"           // SSH 키
];

const PROTECTED_PATHS = [
  ".moai/memory/"   // MoAI 메모리 (개발 가이드)
];
```

**실행 로직:**
```javascript
async execute(input) {
  // 1. 쓰기 작업인지 확인
  const toolName = input.tool_name;
  if (!toolName || !['Write', 'Edit', 'MultiEdit'].includes(toolName)) {
    return { success: true };
  }

  // 2. 파일 경로 추출
  const toolInput = input.tool_input || {};
  const filePath = this.extractFilePath(toolInput);

  // 3. 파일 안전성 검사
  if (!this.checkFileSafety(filePath || '')) {
    return {
      success: false,
      blocked: true,
      message: "민감한 파일은 편집할 수 없습니다.",
      exitCode: 2
    };
  }

  // 4. 정상 통과
  return { success: true };
}
```

**파일 안전성 검사:**
```javascript
checkFileSafety(filePath) {
  if (!filePath) {
    return true;
  }

  const pathLower = filePath.toLowerCase();

  // 1. 민감 키워드 검사
  for (const keyword of SENSITIVE_KEYWORDS) {
    if (pathLower.includes(keyword)) {
      return false;
    }
  }

  // 2. 보호 경로 검사
  for (const protectedPath of PROTECTED_PATHS) {
    if (filePath.includes(protectedPath)) {
      return false;
    }
  }

  return true;
}
```

**실제 코드 예시:**

프로젝트별 보호 경로를 설정할 수 있습니다:
```javascript
// .claude/hooks/alfred/pre-write-guard.js 수정

const PROJECT_PROTECTED_PATHS = [
  ".moai/memory/",
  "config/production/",
  "secrets/",
  "credentials/",
  ".aws/",
  ".gcp/"
];

const ALL_PROTECTED = [...PROTECTED_PATHS, ...PROJECT_PROTECTED_PATHS];
```

**커스터마이징 방법:**

파일 확장자별 추가 검증:
```javascript
checkFileSafety(filePath) {
  if (!filePath) return true;

  // 기존 검사
  if (!this.baseCheckFileSafety(filePath)) {
    return false;
  }

  // 프로덕션 설정 파일 보호
  if (filePath.includes('production') && filePath.endsWith('.env')) {
    console.warn('⚠️  프로덕션 환경 파일입니다. 신중하게 수정하세요.');
  }

  // SQL 파일 경고
  if (filePath.endsWith('.sql')) {
    console.warn('⚠️  SQL 파일입니다. 마이그레이션 도구를 사용하세요.');
  }

  return true;
}
```

**주의사항:**
- 파일 경로는 대소문자를 구분하지 않고 검사 (소문자 변환 후)
- 민감한 파일은 완전히 차단되며 편집 불가
- `.moai/memory/development-guide.md`는 보호되어 있어 실수로 수정 불가

### tag-enforcer: TAG 규칙 강제

**목적:** Code-First TAG 시스템의 불변성(immutability)을 보장하고 TAG 형식을 검증합니다.

**트리거 조건:**
- 소스 파일을 쓰거나 수정할 때
- 파일 확장자가 `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.md`, `.go`, `.rs`, `.java`, `.cpp`, `.hpp`일 때
- 테스트 파일 및 `node_modules`는 제외

**파라미터:**
```typescript
interface TagEnforcerInput {
  tool_name: "Write" | "Edit";
  tool_input: {
    file_path: string;
    content: string;      // 새 내용
  };
}
```

**검증 항목:**
1. **@IMMUTABLE TAG 불변성**: 기존에 `@IMMUTABLE` 마커가 있는 TAG 블록을 수정하려고 하면 차단
2. **TAG 형식 검증**: `@DOC:CATEGORY:DOMAIN-ID` 형식 준수
3. **체인 무결성**: `CHAIN` 라인의 TAG 참조가 올바른지 확인
4. **카테고리 유효성**: SPEC, REQ, DESIGN, TASK, TEST, FEATURE, API, FIX 등만 허용

**실행 로직:**
```javascript
async execute(input) {
  // 1. 파일 쓰기 작업인지 확인
  if (!this.isWriteOperation(input.tool_name)) {
    return { success: true };
  }

  // 2. 검증 대상 파일인지 확인
  const filePath = this.extractFilePath(input.tool_input || {});
  if (!filePath || !this.shouldEnforceTags(filePath)) {
    return { success: true };
  }

  // 3. 기존 파일 내용과 새 내용 비교
  const oldContent = await this.getOriginalFileContent(filePath);
  const newContent = this.extractFileContent(input.tool_input || {});

  // 4. @IMMUTABLE TAG 블록 수정 검사 (핵심!)
  const immutabilityCheck = this.checkImmutability(oldContent, newContent, filePath);
  if (immutabilityCheck.violated) {
    return {
      success: false,
      blocked: true,
      message: `🚫 @IMMUTABLE TAG 수정 금지: ${immutabilityCheck.violationDetails}`,
      suggestions: this.generateImmutabilityHelp(immutabilityCheck),
      exitCode: 2
    };
  }

  // 5. 새 TAG 블록 유효성 검증
  const validation = this.validateCodeFirstTag(newContent);
  if (!validation.isValid) {
    return {
      success: false,
      blocked: true,
      message: `🏷️ Code-First TAG 검증 실패: ${validation.violations.join(', ')}`,
      suggestions: this.generateTagSuggestions(filePath, newContent),
      exitCode: 2
    };
  }

  // 6. 경고 출력 (차단하지 않음)
  if (validation.warnings.length > 0) {
    console.error(`⚠️ TAG 개선 권장: ${validation.warnings.join(', ')}`);
  }

  return {
    success: true,
    message: validation.hasTag
      ? `✅ Code-First TAG 검증 완료`
      : `📝 TAG 블록이 없는 파일 (권장사항)`
  };
}
```

**@IMMUTABLE 검증:**
```javascript
checkImmutability(oldContent, newContent, filePath) {
  // 1. 기존 파일에서 TAG 블록 추출
  const oldTagBlock = this.extractTagBlock(oldContent);
  const newTagBlock = this.extractTagBlock(newContent);

  if (!oldTagBlock) {
    return { violated: false }; // 새 파일이므로 통과
  }

  // 2. @IMMUTABLE 마커 확인
  const wasImmutable = /@IMMUTABLE/m.test(oldTagBlock.content);
  if (!wasImmutable) {
    return { violated: false }; // @IMMUTABLE이 아니므로 수정 가능
  }

  // 3. TAG 블록이 삭제되었는지 확인
  if (!newTagBlock) {
    return {
      violated: true,
      modifiedTag: this.extractMainTag(oldTagBlock.content),
      violationDetails: '@IMMUTABLE TAG 블록이 삭제되었습니다'
    };
  }

  // 4. TAG 블록 내용 비교 (공백 정규화 후)
  const oldNormalized = this.normalizeTagBlock(oldTagBlock.content);
  const newNormalized = this.normalizeTagBlock(newTagBlock.content);

  if (oldNormalized !== newNormalized) {
    return {
      violated: true,
      modifiedTag: this.extractMainTag(oldTagBlock.content),
      violationDetails: '@IMMUTABLE TAG 블록의 내용이 변경되었습니다'
    };
  }

  return { violated: false };
}
```

**TAG 형식 검증:**
```javascript
validateCodeFirstTag(content) {
  const violations = [];
  const warnings = [];

  // 1. TAG 블록 추출
  const tagBlock = this.extractTagBlock(content);
  if (!tagBlock) {
    return {
      isValid: true,
      violations: [],
      warnings: ['파일 최상단에 TAG 블록이 없습니다 (권장사항)'],
      hasTag: false
    };
  }

  // 2. 메인 TAG 검증 (@DOC:CATEGORY:DOMAIN-ID)
  const tagMatch = /@DOC:([A-Z]+):([A-Z0-9-]+)/.exec(tagBlock.content);
  if (!tagMatch) {
    violations.push('@TAG 라인이 발견되지 않았습니다');
  } else {
    const [, category, domainId] = tagMatch;

    // 카테고리 유효성 검사
    const validCategories = ['SPEC', 'REQ', 'DESIGN', 'TASK', 'TEST', 'FEATURE', 'API', 'FIX'];
    if (!validCategories.includes(category)) {
      violations.push(`유효하지 않은 TAG 카테고리: ${category}`);
    }

    // 도메인 ID 형식 검사 (예: AUTH-001)
    if (!/^[A-Z0-9-]+-\d{3,}$/.test(domainId)) {
      warnings.push(`도메인 ID 형식 권장: ${domainId} -> DOMAIN-001`);
    }
  }

  // 3. 체인 검증 (CHAIN: REQ -> DESIGN -> TASK -> TEST)
  const chainMatch = /CHAIN:\s*(.+)/.exec(tagBlock.content);
  if (chainMatch) {
    const chainTags = chainMatch[1].split(/\s*->\s*/);
    for (const chainTag of chainTags) {
      if (!/@[A-Z]+:[A-Z0-9-]+/.test(chainTag.trim())) {
        warnings.push(`체인의 TAG 형식을 확인하세요: ${chainTag.trim()}`);
      }
    }
  }

  // 4. @IMMUTABLE 마커 권장
  if (!/@IMMUTABLE/.test(tagBlock.content)) {
    warnings.push('@IMMUTABLE 마커를 추가하여 TAG 불변성을 보장하는 것을 권장합니다');
  }

  return {
    isValid: violations.length === 0,
    violations,
    warnings,
    hasTag: true
  };
}
```

**실제 코드 예시:**

올바른 TAG 블록 작성:
```javascript
/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001 -> TEST:AUTH-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */
export class AuthenticationService {
  // 구현...
}
```

**커스터마이징 방법:**

TAG 검증 규칙을 완화하거나 강화할 수 있습니다:
```javascript
// .claude/hooks/alfred/tag-enforcer.js 수정

// 경고를 에러로 승격
validateCodeFirstTag(content) {
  const result = this.baseValidateCodeFirstTag(content);

  // @IMMUTABLE 마커 필수화
  if (!result.hasTag) {
    result.violations.push('모든 소스 파일은 TAG 블록이 필수입니다');
    result.isValid = false;
  }

  return result;
}
```

**주의사항:**
- @IMMUTABLE TAG는 절대 수정할 수 없으며, 새로운 TAG를 생성해야 함
- TAG 블록은 파일 최상단(30줄 이내)에 위치해야 감지됨
- 테스트 파일은 TAG 규칙이 적용되지 않음

### file-monitor: 파일 모니터링

**목적:** 프로젝트 파일 변경을 실시간으로 추적하고 주기적으로 체크포인트를 생성합니다.

**트리거 조건:**
- MoAI 프로젝트 세션 시작 시 자동으로 파일 감시 시작
- 파일 변경 이벤트 발생 시 즉시 반응
- 5분마다 체크포인트 생성

**파라미터:**
파일 시스템 이벤트로부터 자동으로 파일 경로와 변경 타입을 수신합니다.

**실행 로직:**
```javascript
async execute(input) {
  try {
    if (this.isMoAIProject()) {
      if (this.watchFiles()) {
        return {
          success: true,
          message: "📁 File monitoring started"
        };
      } else {
        return {
          success: true,
          message: "⚠️  Could not start file monitoring"
        };
      }
    }
    return { success: true };
  } catch (error) {
    return { success: true };
  }
}

// 파일 감시 시작
watchFiles() {
  try {
    if (this.isRunning) {
      return true;
    }

    // Node.js fs.watch를 사용한 재귀적 감시
    this.watcher = fs.watch(
      this.projectRoot,
      { recursive: true },
      (eventType, filename) => {
        if (filename) {
          const fullPath = path.join(this.projectRoot, filename);
          this.onFileChanged(fullPath);
        }
      }
    );

    this.isRunning = true;
    return true;
  } catch (error) {
    return false;
  }
}
```

**파일 변경 처리:**
```javascript
onFileChanged(filePath) {
  // 1. 모니터링 대상 파일인지 확인
  if (!this.shouldMonitorFile(filePath)) {
    return;
  }

  // 2. 변경된 파일 목록에 추가
  this.changedFiles.add(filePath);

  // 3. 파일 변경 이벤트 발생
  const event = {
    path: filePath,
    type: 'modified',
    timestamp: new Date()
  };
  this.emit('fileChanged', event);

  // 4. 체크포인트 생성 시점 확인
  if (this.shouldCreateCheckpoint()) {
    this.createCheckpoint();
  }
}
```

**체크포인트 생성:**
```javascript
createCheckpoint() {
  try {
    const currentTime = Date.now();

    // 변경된 파일 목록 저장 (실제 구현에서는 파일로 저장)
    const changedFilesList = Array.from(this.changedFiles);

    // 체크포인트 이벤트 발생
    this.emit('checkpoint', {
      timestamp: new Date(currentTime),
      changedFiles: changedFilesList
    });

    // 상태 초기화
    this.changedFiles.clear();
    this.lastCheckpointTime = currentTime;

    return true;
  } catch (error) {
    return false;
  }
}
```

**모니터링 대상 파일:**
```javascript
const watchPatterns = new Set([
  '.py',    // Python
  '.js',    // JavaScript
  '.ts',    // TypeScript
  '.md',    // Markdown
  '.json',  // JSON
  '.yml',   // YAML
  '.yaml'
]);

const ignorePatterns = new Set([
  '.git',
  '__pycache__',
  'node_modules',
  '.pytest_cache',
  'dist',
  'build'
]);
```

**실제 코드 예시:**

파일 변경 이벤트 리스너 등록:
```javascript
const monitor = new FileMonitor(process.cwd());

// 파일 변경 이벤트 구독
monitor.on('fileChanged', (event) => {
  console.log(`File changed: ${event.path} at ${event.timestamp}`);

  // 파일 변경에 따른 추가 작업 (예: 테스트 실행)
  if (event.path.endsWith('.test.ts')) {
    console.log('Test file changed, running tests...');
  }
});

// 체크포인트 이벤트 구독
monitor.on('checkpoint', (event) => {
  console.log(`Checkpoint created at ${event.timestamp}`);
  console.log(`Changed files: ${event.changedFiles.length}`);
});

// 모니터링 시작
monitor.execute({});
```

**커스터마이징 방법:**

모니터링 대상 파일 확장자 추가:
```javascript
// .claude/hooks/alfred/file-monitor.js 수정

const watchPatterns = new Set([
  '.py', '.js', '.ts', '.md', '.json', '.yml', '.yaml',
  // 추가 확장자
  '.go', '.rs', '.java', '.cpp', '.hpp',
  '.vue', '.svelte', '.tsx', '.jsx'
]);
```

체크포인트 주기 변경:
```javascript
// 5분 → 10분으로 변경
this.checkpointInterval = 600000; // 10 minutes in milliseconds
```

**주의사항:**
- 파일 감시는 재귀적으로 이루어지므로 대형 프로젝트에서 성능 영향 있을 수 있음
- `node_modules`와 같은 대형 디렉토리는 자동으로 제외됨
- 체크포인트는 메모리에만 저장되며 영구 저장되지 않음 (실제 프로젝트에서는 파일로 저장 구현 필요)

## 커스텀 훅 작성

### 훅 작성 패턴

MoAI-ADK 커스텀 훅을 작성하려면 표준 인터페이스를 준수해야 합니다. 훅은 stdin으로 JSON 입력을 받고, stdout으로 JSON 결과를 반환합니다.

**기본 템플릿:**
```javascript
/**
 * @DOC:CUSTOM:HOOK-NAME-001
 * CHAIN: REQ:CUSTOM-001 -> DESIGN:CUSTOM-001 -> TASK:CUSTOM-001 -> TEST:CUSTOM-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

class CustomHook {
  name = 'custom-hook';

  /**
   * 훅 실행 메서드
   * @param {Object} input - Claude Code로부터 받은 입력
   * @returns {Object} - 실행 결과
   */
  async execute(input) {
    try {
      // 1. 입력 검증
      if (!this.shouldExecute(input)) {
        return { success: true };
      }

      // 2. 비즈니스 로직 실행
      const result = await this.processInput(input);

      // 3. 결과 반환
      return {
        success: true,
        message: result.message,
        data: result.data
      };

    } catch (error) {
      // 4. 에러 처리
      return {
        success: false,
        message: `Custom hook failed: ${error.message}`,
        exitCode: 1
      };
    }
  }

  shouldExecute(input) {
    // 실행 조건 확인
    return true;
  }

  async processInput(input) {
    // 실제 로직 구현
    return {
      message: 'Processing completed',
      data: {}
    };
  }
}

// stdin 입력 처리
async function main() {
  try {
    let input = '';

    process.stdin.setEncoding('utf8');
    for await (const chunk of process.stdin) {
      input += chunk;
    }

    const parsedInput = input.trim() ? JSON.parse(input) : {};
    const hook = new CustomHook();
    const result = await hook.execute(parsedInput);

    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      process.exit(2);
    } else if (!result.success) {
      console.error(`ERROR: ${result.message}`);
      process.exit(result.exitCode || 1);
    } else if (result.message) {
      console.log(result.message);
    }

    process.exit(0);
  } catch (error) {
    console.error(`Hook error: ${error.message}`);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { CustomHook, main };
```

**실전 예시 1: API 키 감지 훅**
```javascript
class ApiKeyDetectorHook {
  name = 'api-key-detector';

  // API 키 패턴
  patterns = [
    /sk_live_[a-zA-Z0-9]{24,}/, // Stripe
    /AIza[a-zA-Z0-9_-]{35}/,    // Google
    /ghp_[a-zA-Z0-9]{36}/,      // GitHub
    /xox[baprs]-[a-zA-Z0-9-]+/  // Slack
  ];

  async execute(input) {
    if (input.tool_name !== 'Write' && input.tool_name !== 'Edit') {
      return { success: true };
    }

    const content = input.tool_input?.content || '';

    for (const pattern of this.patterns) {
      if (pattern.test(content)) {
        return {
          success: false,
          blocked: true,
          message: '🔐 API 키가 감지되었습니다. 환경 변수를 사용하세요.',
          exitCode: 2
        };
      }
    }

    return { success: true };
  }
}
```

**실전 예시 2: 코드 복잡도 체크 훅**
```javascript
class ComplexityCheckHook {
  name = 'complexity-check';
  maxComplexity = 10;

  async execute(input) {
    if (input.tool_name !== 'Write') {
      return { success: true };
    }

    const filePath = input.tool_input?.file_path;
    const content = input.tool_input?.content || '';

    // TypeScript/JavaScript 파일만 검사
    if (!filePath?.match(/\.(ts|js)$/)) {
      return { success: true };
    }

    const functions = this.extractFunctions(content);
    const violations = [];

    for (const func of functions) {
      const complexity = this.calculateComplexity(func.body);
      if (complexity > this.maxComplexity) {
        violations.push(`${func.name}: 복잡도 ${complexity} (최대 ${this.maxComplexity})`);
      }
    }

    if (violations.length > 0) {
      console.warn('⚠️  복잡도 초과 함수:\n' + violations.join('\n'));
    }

    return { success: true };
  }

  extractFunctions(content) {
    // 간단한 함수 추출 (실제로는 AST 파서 사용 권장)
    const functionPattern = /function\s+(\w+)|const\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>/g;
    const functions = [];
    let match;

    while ((match = functionPattern.exec(content)) !== null) {
      functions.push({
        name: match[1] || match[2],
        body: this.extractFunctionBody(content, match.index)
      });
    }

    return functions;
  }

  calculateComplexity(body) {
    // 간단한 복잡도 계산 (실제로는 cyclomatic complexity 사용)
    const complexityKeywords = ['if', 'else', 'for', 'while', 'case', '&&', '||', '?'];
    let complexity = 1;

    for (const keyword of complexityKeywords) {
      const regex = new RegExp(`\\b${keyword}\\b`, 'g');
      const matches = body.match(regex);
      if (matches) {
        complexity += matches.length;
      }
    }

    return complexity;
  }

  extractFunctionBody(content, startIndex) {
    // 함수 본문 추출 (간단한 구현)
    let depth = 0;
    let body = '';
    let started = false;

    for (let i = startIndex; i < content.length; i++) {
      const char = content[i];

      if (char === '{') {
        depth++;
        started = true;
      } else if (char === '}') {
        depth--;
        if (depth === 0 && started) {
          return body;
        }
      }

      if (started) {
        body += char;
      }
    }

    return body;
  }
}
```

### 보안 정책 설정

커스텀 훅에서 보안 정책을 구현할 때는 다음 원칙을 따르세요:

**1. 최소 권한 원칙 (Principle of Least Privilege)**
```javascript
class SecureHook {
  // 명시적으로 허용된 작업만 수행
  allowedOperations = ['read', 'validate'];

  async execute(input) {
    const operation = input.operation;

    if (!this.allowedOperations.includes(operation)) {
      return {
        success: false,
        blocked: true,
        message: `Operation '${operation}' is not allowed`,
        exitCode: 2
      };
    }

    // 허용된 작업 수행
    return await this.performOperation(operation, input);
  }
}
```

**2. 입력 검증**
```javascript
validateInput(input) {
  // 필수 필드 확인
  if (!input.tool_name) {
    throw new Error('tool_name is required');
  }

  // 타입 검증
  if (typeof input.tool_name !== 'string') {
    throw new Error('tool_name must be a string');
  }

  // 길이 제한
  if (input.prompt && input.prompt.length > 10000) {
    throw new Error('prompt is too long');
  }

  // XSS 방지
  if (this.containsXSS(input.prompt)) {
    throw new Error('XSS detected in prompt');
  }

  return true;
}

containsXSS(text) {
  const xssPatterns = [
    /<script[^>]*>.*?<\/script>/gi,
    /on\w+\s*=\s*["'][^"']*["']/gi,
    /javascript:/gi
  ];

  return xssPatterns.some(pattern => pattern.test(text));
}
```

**3. 경로 검증**
```javascript
validateFilePath(filePath) {
  // 절대 경로 확인
  if (!path.isAbsolute(filePath)) {
    throw new Error('Only absolute paths are allowed');
  }

  // 경로 탐색 공격 방지
  const normalizedPath = path.normalize(filePath);
  if (normalizedPath.includes('..')) {
    throw new Error('Path traversal detected');
  }

  // 허용된 디렉토리 확인
  const allowedDirs = [
    path.join(process.cwd(), 'src'),
    path.join(process.cwd(), 'tests')
  ];

  const isAllowed = allowedDirs.some(dir =>
    normalizedPath.startsWith(dir)
  );

  if (!isAllowed) {
    throw new Error('Access denied to this directory');
  }

  return normalizedPath;
}
```

**4. 민감 정보 마스킹**
```javascript
maskSensitiveData(content) {
  // API 키 마스킹
  content = content.replace(/sk_live_[a-zA-Z0-9]{24,}/g, 'sk_live_***MASKED***');

  // 이메일 마스킹
  content = content.replace(/([a-zA-Z0-9._%+-]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})/g, '***@$2');

  // 신용카드 번호 마스킹
  content = content.replace(/\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b/g, '****-****-****-****');

  return content;
}
```

### 테스트 방법

커스텀 훅을 테스트하려면 다음 방법을 사용하세요:

**1. 단위 테스트 (Vitest)**
```javascript
// custom-hook.test.js
import { describe, it, expect } from 'vitest';
import { CustomHook } from './custom-hook.js';

describe('CustomHook', () => {
  it('should execute successfully with valid input', async () => {
    const hook = new CustomHook();
    const input = {
      tool_name: 'Write',
      tool_input: {
        file_path: '/path/to/file.ts',
        content: 'console.log("test");'
      }
    };

    const result = await hook.execute(input);

    expect(result.success).toBe(true);
  });

  it('should block when API key is detected', async () => {
    const hook = new CustomHook();
    const input = {
      tool_name: 'Write',
      tool_input: {
        file_path: '/path/to/file.ts',
        content: 'const apiKey = "sk_live_abcdefghijklmnopqrstuvwxyz";'
      }
    };

    const result = await hook.execute(input);

    expect(result.success).toBe(false);
    expect(result.blocked).toBe(true);
  });
});
```

**2. 통합 테스트 (실제 stdin/stdout)**
```bash
# test-hook.sh
#!/bin/bash

# 테스트 입력 준비
INPUT='{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/tmp/test.ts",
    "content": "console.log(\"test\");"
  }
}'

# 훅 실행
RESULT=$(echo "$INPUT" | node .claude/hooks/alfred/custom-hook.js)

# 결과 확인
if [ $? -eq 0 ]; then
  echo "✅ Hook executed successfully"
  echo "Output: $RESULT"
else
  echo "❌ Hook failed"
  exit 1
fi
```

**3. 엣지 케이스 테스트**
```javascript
describe('CustomHook edge cases', () => {
  it('should handle empty input', async () => {
    const hook = new CustomHook();
    const result = await hook.execute({});
    expect(result.success).toBe(true);
  });

  it('should handle malformed input', async () => {
    const hook = new CustomHook();
    const result = await hook.execute({ invalid: 'data' });
    expect(result.success).toBe(true); // 안전한 기본 동작
  });

  it('should handle timeout', async () => {
    const hook = new CustomHook();
    const input = { /* 매우 큰 데이터 */ };

    const resultPromise = hook.execute(input);
    const timeoutPromise = new Promise((_, reject) =>
      setTimeout(() => reject(new Error('Timeout')), 10000)
    );

    await expect(Promise.race([resultPromise, timeoutPromise]))
      .rejects.toThrow('Timeout');
  });
});
```

### 배포 전략

커스텀 훅을 프로젝트에 배포하는 단계별 가이드:

**1. 훅 파일 작성 및 배치**
```bash
# 훅 파일 생성
touch .claude/hooks/alfred/my-custom-hook.js

# 실행 권한 부여
chmod +x .claude/hooks/alfred/my-custom-hook.js
```

**2. settings.json 업데이트**
```json
{
  "hooks": [
    {
      "name": "my-custom-hook",
      "path": ".claude/hooks/alfred/my-custom-hook.js",
      "event": "pre-write",
      "enabled": true,
      "timeout": 10000
    }
  ]
}
```

**3. 로컬 테스트**
```bash
# 단위 테스트 실행
npm test -- custom-hook.test.js

# 통합 테스트 실행
bash test-hook.sh

# Claude Code에서 실제 테스트
# (테스트 파일 생성하여 훅 트리거)
```

**4. 점진적 배포**
```javascript
// 훅에 feature flag 추가
class CustomHook {
  async execute(input) {
    // feature flag 확인
    const isEnabled = process.env.CUSTOM_HOOK_ENABLED === 'true';
    if (!isEnabled) {
      return { success: true, message: 'Custom hook disabled' };
    }

    // 실제 로직
    return this.processInput(input);
  }
}
```

**5. 모니터링 및 롤백**
```javascript
class CustomHook {
  async execute(input) {
    const startTime = Date.now();

    try {
      const result = await this.processInput(input);

      // 성능 로깅
      const duration = Date.now() - startTime;
      if (duration > 5000) {
        console.warn(`⚠️  Custom hook took ${duration}ms`);
      }

      return result;
    } catch (error) {
      // 에러 로깅
      console.error(`❌ Custom hook error: ${error.message}`);

      // 안전한 폴백
      return { success: true, message: 'Custom hook failed, continuing...' };
    }
  }
}
```

**6. 버전 관리**
```javascript
/**
 * @DOC:CUSTOM:HOOK-001
 * VERSION: 1.0.0
 * CHANGELOG:
 *   - 1.0.0: Initial release
 *   - 1.1.0: Added API key detection
 *   - 1.2.0: Improved performance
 */
class CustomHook {
  version = '1.2.0';
  // ...
}
```

## 훅 실행 흐름도

### 전체 시스템 아키텍처

```mermaid
flowchart TB
    Start([Claude Code 시작]) --> Init[훅 시스템 초기화]
    Init --> LoadConfig[.claude/settings.json 로드]
    LoadConfig --> RegisterHooks[훅 등록]

    RegisterHooks --> SessionStart[세션 시작 이벤트]
    SessionStart --> SN[session-notice 실행]
    SessionStart --> LD[language-detector 실행]
    SessionStart --> FM[file-monitor 시작]

    SN --> WaitEvent[이벤트 대기]
    LD --> WaitEvent
    FM --> WaitEvent

    WaitEvent --> UserPrompt{사용자 프롬프트?}
    WaitEvent --> FileWrite{파일 쓰기?}
    WaitEvent --> BashCmd{Bash 명령?}

    UserPrompt -->|Yes| SG[steering-guard]
    SG --> PromptBlocked{차단?}
    PromptBlocked -->|Yes| BlockAction[작업 차단]
    PromptBlocked -->|No| WaitEvent

    FileWrite -->|Yes| PW[pre-write-guard]
    PW --> PathBlocked{민감 파일?}
    PathBlocked -->|Yes| BlockAction
    PathBlocked -->|No| TE[tag-enforcer]

    TE --> TagBlocked{TAG 위반?}
    TagBlocked -->|Yes| BlockAction
    TagBlocked -->|No| AllowWrite[파일 쓰기 허용]
    AllowWrite --> FM

    BashCmd -->|Yes| PB[policy-block]
    PB --> CmdBlocked{위험 명령?}
    CmdBlocked -->|Yes| BlockAction
    CmdBlocked -->|No| AllowCmd[명령 실행 허용]
    AllowCmd --> WaitEvent

    BlockAction --> ErrorMsg[에러 메시지 출력]
    ErrorMsg --> WaitEvent
```

### 파일 쓰기 체인 상세도

```mermaid
sequenceDiagram
    participant User
    participant Claude as Claude Code
    participant HM as Hook Manager
    participant SG as steering-guard
    participant PW as pre-write-guard
    participant TE as tag-enforcer
    participant FS as File System
    participant FM as file-monitor

    User->>Claude: "auth.ts 파일 생성"
    Claude->>HM: Write 이벤트 발생

    rect rgb(200, 220, 240)
        Note over HM,SG: Phase 1: 프롬프트 검증
        HM->>SG: execute(input)
        SG->>SG: 금지 패턴 확인
        SG-->>HM: {success: true}
    end

    rect rgb(220, 240, 200)
        Note over HM,PW: Phase 2: 경로 보안 검사
        HM->>PW: execute(input)
        PW->>PW: 민감 파일 확인
        PW->>PW: 보호 경로 확인
        PW-->>HM: {success: true}
    end

    rect rgb(240, 220, 200)
        Note over HM,TE: Phase 3: TAG 무결성 검증
        HM->>TE: execute(input)
        TE->>FS: 기존 파일 읽기
        FS-->>TE: 파일 없음 (신규)
        TE->>TE: TAG 블록 검증

        alt TAG 블록 없음
            TE-->>HM: {success: true, warnings: [...]}
        else "TAG 블록 있음 & @IMMUTABLE"
            TE->>TE: 불변성 검사
            alt 수정 감지
                TE-->>HM: {blocked: true}
                HM-->>Claude: 작업 차단
                Claude-->>User: 에러 메시지
            else 정상
                TE-->>HM: {success: true}
            end
        end
    end

    rect rgb(200, 240, 220)
        Note over HM,FS: Phase 4: 파일 쓰기 & 모니터링
        HM-->>Claude: 모든 훅 통과
        Claude->>FS: 파일 쓰기
        FS-->>FM: 파일 변경 이벤트
        FM->>FM: 변경 목록 업데이트
        FM->>FM: 체크포인트 확인
        FS-->>Claude: 쓰기 완료
        Claude-->>User: 성공
    end
```

### 에러 전파 다이어그램

```mermaid
flowchart LR
    subgraph "훅 실행 계층"
        H1[Hook 1 - steering-guard] --> H2[Hook 2 - pre-write-guard]
        H2 --> H3[Hook 3 - tag-enforcer]
    end

    subgraph "에러 처리 계층"
        E1{에러 발생?} --> |Yes| E2[에러 타입 판별]
        E2 --> |치명적| E3[작업 차단 - exitCode=2]
        E2 --> |경고| E4[경고 로그 - 계속 진행]
        E2 --> |타임아웃| E5[훅 중단 - 계속 진행]

        E3 --> Stop([작업 중단])
        E4 --> Continue([다음 훅])
        E5 --> Continue
        E1 --> |No| Continue
    end

    H1 --> E1
    H2 --> E1
    H3 --> E1

    Continue --> Result{모든 훅 완료?}
    Result --> |Yes| Success([작업 허용])
    Result --> |No| NextHook[다음 훅 실행]
    NextHook --> H1

```

## 실전 예시

### 파일 쓰기 제한 훅

**시나리오:** 특정 브랜치에서 특정 파일 수정을 제한하고 싶습니다.

```javascript
/**
 * @DOC:SECURITY:BRANCH-PROTECTION-001
 * CHAIN: REQ:SEC-001 -> DESIGN:SEC-001 -> TASK:SEC-001 -> TEST:SEC-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

const { execSync } = require('child_process');

class BranchProtectionHook {
  name = 'branch-protection';

  // 보호 규칙 정의
  protectionRules = [
    {
      branch: 'main',
      protectedPaths: [
        'package.json',
        'package-lock.json',
        '.moai/config.json',
        'src/core/**/*'
      ],
      message: '⚠️  main 브랜치에서 핵심 파일 수정 불가. feature 브랜치를 사용하세요.'
    },
    {
      branch: 'production',
      protectedPaths: ['**/*'],
      message: '🚫 production 브랜치는 직접 수정할 수 없습니다.'
    }
  ];

  async execute(input) {
    // 1. 파일 쓰기 작업인지 확인
    if (!['Write', 'Edit'].includes(input.tool_name)) {
      return { success: true };
    }

    // 2. 현재 브랜치 확인
    const currentBranch = this.getCurrentBranch();

    // 3. 파일 경로 추출
    const filePath = this.extractFilePath(input.tool_input);
    if (!filePath) {
      return { success: true };
    }

    // 4. 보호 규칙 적용
    for (const rule of this.protectionRules) {
      if (currentBranch === rule.branch) {
        if (this.matchesProtectedPath(filePath, rule.protectedPaths)) {
          return {
            success: false,
            blocked: true,
            message: rule.message,
            exitCode: 2
          };
        }
      }
    }

    return { success: true };
  }

  getCurrentBranch() {
    try {
      const branch = execSync('git rev-parse --abbrev-ref HEAD', {
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'ignore']
      }).trim();
      return branch;
    } catch (error) {
      return 'unknown';
    }
  }

  extractFilePath(toolInput) {
    return toolInput?.file_path || toolInput?.filePath || null;
  }

  matchesProtectedPath(filePath, protectedPaths) {
    const path = require('path');
    const minimatch = require('minimatch');

    for (const pattern of protectedPaths) {
      if (minimatch(filePath, pattern, { matchBase: true })) {
        return true;
      }
    }

    return false;
  }
}

// stdin/stdout 처리
async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');

  for await (const chunk of process.stdin) {
    input += chunk;
  }

  const parsedInput = input.trim() ? JSON.parse(input) : {};
  const hook = new BranchProtectionHook();
  const result = await hook.execute(parsedInput);

  if (result.blocked) {
    console.error(`BLOCKED: ${result.message}`);
    process.exit(2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message}`);
    process.exit(result.exitCode || 1);
  } else if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

if (require.main === module) {
  main();
}

module.exports = { BranchProtectionHook, main };
```

### 비밀 정보 마스킹 훅

**시나리오:** 코드에 하드코딩된 비밀 정보를 감지하고 자동으로 환경 변수 사용을 제안합니다.

```javascript
/**
 * @DOC:SECURITY:SECRET-DETECTION-001
 * CHAIN: REQ:SEC-002 -> DESIGN:SEC-002 -> TASK:SEC-002 -> TEST:SEC-002
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

class SecretDetectionHook {
  name = 'secret-detection';

  // 비밀 정보 패턴
  secretPatterns = [
    {
      name: 'AWS Access Key',
      pattern: /AKIA[0-9A-Z]{16}/g,
      suggestion: 'process.env.AWS_ACCESS_KEY_ID'
    },
    {
      name: 'Stripe API Key',
      pattern: /sk_live_[a-zA-Z0-9]{24,}/g,
      suggestion: 'process.env.STRIPE_SECRET_KEY'
    },
    {
      name: 'GitHub Token',
      pattern: /ghp_[a-zA-Z0-9]{36}/g,
      suggestion: 'process.env.GITHUB_TOKEN'
    },
    {
      name: 'JWT Token',
      pattern: /eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}/g,
      suggestion: 'process.env.JWT_SECRET'
    },
    {
      name: 'Private Key',
      pattern: /-----BEGIN (?:RSA |EC )?PRIVATE KEY-----/g,
      suggestion: 'process.env.PRIVATE_KEY'
    },
    {
      name: 'Database URL',
      pattern: /(?:postgres|mysql|mongodb):\/\/[^:]+:[^@]+@[^\/]+/g,
      suggestion: 'process.env.DATABASE_URL'
    }
  ];

  async execute(input) {
    if (!['Write', 'Edit'].includes(input.tool_name)) {
      return { success: true };
    }

    const content = input.tool_input?.content || '';
    const filePath = input.tool_input?.file_path || 'unknown';

    const detections = [];

    // 각 패턴으로 검사
    for (const { name, pattern, suggestion } of this.secretPatterns) {
      const matches = content.match(pattern);
      if (matches) {
        detections.push({
          name,
          count: matches.length,
          suggestion
        });
      }
    }

    if (detections.length > 0) {
      const message = this.generateWarningMessage(filePath, detections);

      return {
        success: false,
        blocked: true,
        message: message,
        exitCode: 2
      };
    }

    return { success: true };
  }

  generateWarningMessage(filePath, detections) {
    const lines = [
      `🔐 비밀 정보가 감지되었습니다: ${filePath}`,
      '',
      '감지된 비밀 정보:'
    ];

    for (const { name, count, suggestion } of detections) {
      lines.push(`  - ${name}: ${count}개`);
      lines.push(`    대신 사용하세요: ${suggestion}`);
    }

    lines.push('');
    lines.push('💡 해결 방법:');
    lines.push('  1. .env 파일에 비밀 정보 저장');
    lines.push('  2. 코드에서 process.env 사용');
    lines.push('  3. .env를 .gitignore에 추가');

    return lines.join('\n');
  }
}

async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');

  for await (const chunk of process.stdin) {
    input += chunk;
  }

  const parsedInput = input.trim() ? JSON.parse(input) : {};
  const hook = new SecretDetectionHook();
  const result = await hook.execute(parsedInput);

  if (result.blocked) {
    console.error(`BLOCKED: ${result.message}`);
    process.exit(2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message}`);
    process.exit(result.exitCode || 1);
  }

  process.exit(0);
}

if (require.main === module) {
  main();
}

module.exports = { SecretDetectionHook, main };
```

### 성능 모니터링 훅

**시나리오:** 파일 쓰기 작업의 성능을 모니터링하고 느린 훅을 감지합니다.

```javascript
/**
 * @DOC:PERF:HOOK-MONITORING-001
 * CHAIN: REQ:PERF-001 -> DESIGN:PERF-001 -> TASK:PERF-001 -> TEST:PERF-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

const fs = require('fs');
const path = require('path');

class PerformanceMonitoringHook {
  name = 'performance-monitoring';
  logPath = path.join(process.cwd(), '.moai', 'logs', 'hook-performance.log');
  warningThreshold = 1000; // 1초

  async execute(input) {
    const startTime = Date.now();

    try {
      // 훅 실행 시뮬레이션
      await this.simulateHookExecution(input);

      const duration = Date.now() - startTime;

      // 성능 로그 기록
      this.logPerformance({
        hook: this.name,
        event: input.tool_name,
        duration,
        timestamp: new Date().toISOString()
      });

      // 느린 훅 경고
      if (duration > this.warningThreshold) {
        console.warn(
          `⚠️  Hook '${this.name}' took ${duration}ms (threshold: ${this.warningThreshold}ms)`
        );
      }

      return {
        success: true,
        message: `Performance monitored: ${duration}ms`,
        data: { duration }
      };

    } catch (error) {
      return {
        success: false,
        message: `Performance monitoring failed: ${error.message}`,
        exitCode: 1
      };
    }
  }

  async simulateHookExecution(input) {
    // 실제 훅 로직을 여기에 구현
    // 예: 파일 읽기, 검증 등
    return new Promise(resolve => setTimeout(resolve, 100));
  }

  logPerformance(data) {
    try {
      // 로그 디렉토리 생성
      const logDir = path.dirname(this.logPath);
      if (!fs.existsSync(logDir)) {
        fs.mkdirSync(logDir, { recursive: true });
      }

      // JSON Lines 형식으로 로그 추가
      const logLine = JSON.stringify(data) + '\n';
      fs.appendFileSync(this.logPath, logLine);

      // 로그 크기 제한 (1MB)
      const stats = fs.statSync(this.logPath);
      if (stats.size > 1024 * 1024) {
        this.rotateLog();
      }

    } catch (error) {
      console.error(`Failed to log performance: ${error.message}`);
    }
  }

  rotateLog() {
    try {
      const backupPath = `${this.logPath}.${Date.now()}.bak`;
      fs.renameSync(this.logPath, backupPath);

      // 오래된 백업 삭제 (최근 5개만 유지)
      const logDir = path.dirname(this.logPath);
      const backups = fs.readdirSync(logDir)
        .filter(f => f.startsWith('hook-performance.log.') && f.endsWith('.bak'))
        .sort()
        .reverse();

      for (let i = 5; i < backups.length; i++) {
        fs.unlinkSync(path.join(logDir, backups[i]));
      }

    } catch (error) {
      console.error(`Failed to rotate log: ${error.message}`);
    }
  }

  // 성능 리포트 생성 (별도 메서드)
  generatePerformanceReport() {
    try {
      const logs = fs.readFileSync(this.logPath, 'utf8')
        .split('\n')
        .filter(line => line.trim())
        .map(line => JSON.parse(line));

      const report = {
        totalExecutions: logs.length,
        averageDuration: logs.reduce((sum, log) => sum + log.duration, 0) / logs.length,
        slowestHooks: logs
          .sort((a, b) => b.duration - a.duration)
          .slice(0, 10),
        durationDistribution: {
          fast: logs.filter(log => log.duration < 100).length,
          medium: logs.filter(log => log.duration >= 100 && log.duration < 1000).length,
          slow: logs.filter(log => log.duration >= 1000).length
        }
      };

      return report;

    } catch (error) {
      return { error: error.message };
    }
  }
}

async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');

  for await (const chunk of process.stdin) {
    input += chunk;
  }

  const parsedInput = input.trim() ? JSON.parse(input) : {};
  const hook = new PerformanceMonitoringHook();
  const result = await hook.execute(parsedInput);

  if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

if (require.main === module) {
  main();
}

module.exports = { PerformanceMonitoringHook, main };
```

### 자동 백업 훅

**시나리오:** 중요한 파일을 수정하기 전에 자동으로 백업을 생성합니다.

```javascript
/**
 * @DOC:FEATURE:AUTO-BACKUP-001
 * CHAIN: REQ:BACKUP-001 -> DESIGN:BACKUP-001 -> TASK:BACKUP-001 -> TEST:BACKUP-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

const fs = require('fs');
const path = require('path');

class AutoBackupHook {
  name = 'auto-backup';
  backupDir = path.join(process.cwd(), '.moai', 'backups');
  maxBackupsPerFile = 5;

  // 백업 대상 파일 패턴
  backupPatterns = [
    '.moai/config.json',
    'package.json',
    'tsconfig.json',
    '.moai/project/**/*.md',
    'src/**/*.ts',
    'src/**/*.js'
  ];

  async execute(input) {
    if (!['Write', 'Edit'].includes(input.tool_name)) {
      return { success: true };
    }

    const filePath = this.extractFilePath(input.tool_input);
    if (!filePath) {
      return { success: true };
    }

    // 백업 대상인지 확인
    if (!this.shouldBackup(filePath)) {
      return { success: true };
    }

    try {
      // 파일이 이미 존재하는 경우에만 백업
      if (fs.existsSync(filePath)) {
        const backupPath = await this.createBackup(filePath);
        console.log(`💾 Backup created: ${path.basename(backupPath)}`);
      }

      return { success: true };

    } catch (error) {
      // 백업 실패해도 작업은 계속 진행 (경고만 출력)
      console.warn(`⚠️  Backup failed: ${error.message}`);
      return { success: true };
    }
  }

  extractFilePath(toolInput) {
    return toolInput?.file_path || toolInput?.filePath || null;
  }

  shouldBackup(filePath) {
    const minimatch = require('minimatch');

    for (const pattern of this.backupPatterns) {
      if (minimatch(filePath, pattern, { matchBase: true })) {
        return true;
      }
    }

    return false;
  }

  async createBackup(filePath) {
    // 백업 디렉토리 생성
    if (!fs.existsSync(this.backupDir)) {
      fs.mkdirSync(this.backupDir, { recursive: true });
    }

    // 백업 파일명 생성 (타임스탬프 포함)
    const fileName = path.basename(filePath);
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const backupFileName = `${fileName}.${timestamp}.bak`;
    const backupPath = path.join(this.backupDir, backupFileName);

    // 파일 복사
    fs.copyFileSync(filePath, backupPath);

    // 오래된 백업 정리
    this.cleanOldBackups(fileName);

    return backupPath;
  }

  cleanOldBackups(fileName) {
    try {
      // 해당 파일의 모든 백업 찾기
      const backups = fs.readdirSync(this.backupDir)
        .filter(f => f.startsWith(fileName) && f.endsWith('.bak'))
        .map(f => ({
          name: f,
          path: path.join(this.backupDir, f),
          time: fs.statSync(path.join(this.backupDir, f)).mtime
        }))
        .sort((a, b) => b.time - a.time); // 최신순 정렬

      // 최대 개수 초과 시 오래된 것 삭제
      if (backups.length > this.maxBackupsPerFile) {
        for (let i = this.maxBackupsPerFile; i < backups.length; i++) {
          fs.unlinkSync(backups[i].path);
        }
      }

    } catch (error) {
      console.warn(`⚠️  Failed to clean old backups: ${error.message}`);
    }
  }

  // 백업 복원 메서드 (별도 스크립트에서 호출)
  restoreBackup(backupPath, targetPath) {
    try {
      if (!fs.existsSync(backupPath)) {
        throw new Error(`Backup not found: ${backupPath}`);
      }

      // 현재 파일을 임시 백업
      if (fs.existsSync(targetPath)) {
        const tempBackup = `${targetPath}.temp.bak`;
        fs.copyFileSync(targetPath, tempBackup);
      }

      // 백업 복원
      fs.copyFileSync(backupPath, targetPath);

      console.log(`✅ Restored: ${targetPath} from ${path.basename(backupPath)}`);
      return true;

    } catch (error) {
      console.error(`❌ Restore failed: ${error.message}`);
      return false;
    }
  }
}

async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');

  for await (const chunk of process.stdin) {
    input += chunk;
  }

  const parsedInput = input.trim() ? JSON.parse(input) : {};
  const hook = new AutoBackupHook();
  const result = await hook.execute(parsedInput);

  if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

if (require.main === module) {
  main();
}

module.exports = { AutoBackupHook, main };
```

### 커스텀 검증 훅

**시나리오:** 프로젝트별 코딩 컨벤션을 자동으로 검증합니다.

```javascript
/**
 * @DOC:QUALITY:CONVENTION-CHECK-001
 * CHAIN: REQ:QA-001 -> DESIGN:QA-001 -> TASK:QA-001 -> TEST:QA-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-09-30
 * @IMMUTABLE
 */

class ConventionCheckHook {
  name = 'convention-check';

  // 프로젝트별 컨벤션 규칙
  conventions = {
    typescript: [
      {
        name: '인터페이스는 I prefix 없이 작성',
        pattern: /interface\s+I[A-Z]/g,
        suggestion: 'interface IUser -> interface User'
      },
      {
        name: '클래스명은 PascalCase',
        pattern: /class\s+[a-z]/g,
        suggestion: 'class myService -> class MyService'
      },
      {
        name: '상수는 UPPER_SNAKE_CASE',
        pattern: /const\s+[a-z][a-zA-Z]*\s*=\s*["'`]/g,
        suggestion: 'const apiKey -> const API_KEY'
      }
    ],
    python: [
      {
        name: '클래스명은 PascalCase',
        pattern: /class\s+[a-z]/g,
        suggestion: 'class my_service -> class MyService'
      },
      {
        name: '함수명은 snake_case',
        pattern: /def\s+[A-Z]/g,
        suggestion: 'def GetUser -> def get_user'
      }
    ]
  };

  async execute(input) {
    if (input.tool_name !== 'Write') {
      return { success: true };
    }

    const filePath = input.tool_input?.file_path || '';
    const content = input.tool_input?.content || '';

    // 파일 타입 감지
    const fileType = this.detectFileType(filePath);
    if (!fileType) {
      return { success: true };
    }

    // 컨벤션 검사
    const violations = this.checkConventions(content, fileType);

    if (violations.length > 0) {
      const message = this.generateViolationMessage(filePath, violations);
      console.warn(message); // 경고만 출력, 차단하지 않음
    }

    return { success: true };
  }

  detectFileType(filePath) {
    if (filePath.match(/\.(ts|tsx)$/)) return 'typescript';
    if (filePath.match(/\.(js|jsx)$/)) return 'javascript';
    if (filePath.match(/\.py$/)) return 'python';
    return null;
  }

  checkConventions(content, fileType) {
    const rules = this.conventions[fileType] || [];
    const violations = [];

    for (const rule of rules) {
      const matches = content.match(rule.pattern);
      if (matches) {
        violations.push({
          name: rule.name,
          count: matches.length,
          suggestion: rule.suggestion,
          examples: matches.slice(0, 3) // 처음 3개만
        });
      }
    }

    return violations;
  }

  generateViolationMessage(filePath, violations) {
    const lines = [
      `📋 코딩 컨벤션 위반: ${filePath}`,
      ''
    ];

    for (const v of violations) {
      lines.push(`  ⚠️  ${v.name} (${v.count}개)`);
      lines.push(`     ${v.suggestion}`);
      if (v.examples.length > 0) {
        lines.push(`     예: ${v.examples.join(', ')}`);
      }
      lines.push('');
    }

    return lines.join('\n');
  }
}

async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');

  for await (const chunk of process.stdin) {
    input += chunk;
  }

  const parsedInput = input.trim() ? JSON.parse(input) : {};
  const hook = new ConventionCheckHook();
  const result = await hook.execute(parsedInput);

  if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

if (require.main === module) {
  main();
}

module.exports = { ConventionCheckHook, main };
```

## 마무리

MoAI-ADK 훅 시스템은 Claude Code의 강력한 이벤트 기반 아키텍처를 활용하여 개발 프로세스를 자동화하고 보호합니다. 7개의 내장 훅이 개발 가이드 준수, TAG 무결성, 보안 정책을 자동으로 검증하며, 커스텀 훅을 작성하여 프로젝트별 요구사항을 구현할 수 있습니다.

훅 시스템의 핵심은 **비침투적 자동화**입니다. 개발자가 의식하지 못하는 사이 백그라운드에서 작동하여 인적 오류를 줄이고, 치명적 문제는 차단하며, 개선 사항은 제안합니다.

이 가이드에서 다룬 내용:
- 훅 생명주기와 실행 순서
- 7개 내장 훅의 상세 동작 방식
- 커스텀 훅 작성 패턴과 보안 정책
- 실전 예시: 파일 제한, 비밀 감지, 성능 모니터링, 자동 백업

더 자세한 내용은 [MoAI-ADK 공식 문서](https://moai-adk.vercel.app)를 참고하세요.