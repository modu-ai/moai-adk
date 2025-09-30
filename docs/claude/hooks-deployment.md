# MoAI-ADK Hooks 배포 전략

## 🎯 설계 목표

**"추가 설치 없이 Node.js만으로 모든 환경에서 실행 가능"**

- ✅ Node.js 18+ 기반 (추가 패키지 불필요)
- ✅ "type": "module" 환경과 충돌 없음
- ✅ Windows/Mac/Linux 모두 호환
- ✅ Python 의존성 없음

---

## 📦 선택한 방식: CommonJS (.cjs)

### 왜 .cjs인가?

| 기준 | .cjs | .mjs | .js (ESM) | .ts (Node 22.6+) | tsx |
|------|------|------|-----------|------------------|-----|
| **추가 설치** | ❌ 없음 | ❌ 없음 | ❌ 없음 | ❌ 없음 | ✅ 필요 |
| **Node 버전** | 12+ | 12+ | 14+ | **22.6+** | 18+ |
| **빌드 필요** | 개발자만 | 개발자만 | ❌ | ❌ | ❌ |
| **"type": "module" 충돌** | ❌ 없음 | ❌ 없음 | ✅ 있음 | ❌ 없음 | ❌ 없음 |
| **타입 안전성** | ✅ (빌드 시) | ✅ (빌드 시) | ❌ | ⚠️ (타입 체크 안함) | ❌ |
| **프로덕션 안정성** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |

**결론**: `.cjs`는 **최대 호환성**과 **제로 의존성**을 동시에 달성

---

## 🏗️ 빌드 & 배포 프로세스

### 1. 개발자 환경 (TypeScript)

```
moai-adk-ts/
└── src/
    └── claude/
        └── hooks/
            └── session/
                └── session-notice.ts  ← TypeScript 소스
```

### 2. 빌드 명령어

```bash
bun run build:hooks
```

**실행 과정**:
1. `tsup --config tsup.hooks.config.ts` - TypeScript → .cjs 변환
2. `cp ../.claude/hooks/moai/*.cjs templates/.claude/hooks/moai/` - 템플릿에 복사

**tsup 설정** (`tsup.hooks.config.ts`):
```typescript
export default defineConfig({
  entry: {
    'session-notice': 'src/claude/hooks/session/session-notice.ts',
  },
  format: ['cjs'],                        // CommonJS 형식
  target: 'node18',                       // Node.js 18+
  outDir: '../.claude/hooks/moai',        // 출력 디렉토리
  outExtension: () => ({ js: '.cjs' }),   // .cjs 확장자
  bundle: true,                           // 의존성 번들링
  minify: false,                          // 디버깅 가능하게
});
```

### 3. 배포 패키지 구조

```
moai-adk/
├── dist/                          # CLI 및 라이브러리
└── templates/
    └── .claude/
        ├── hooks/
        │   └── moai/
        │       ├── session-notice.cjs        ✅ 빌드됨
        │       ├── pre-write-guard.cjs       ✅ 빌드됨
        │       ├── tag-enforcer.cjs          ✅ 빌드됨
        │       ├── steering-guard.cjs        ✅ 빌드됨
        │       └── policy-block.cjs          ✅ 빌드됨
        └── settings.json
```

### 4. 사용자 환경 (설치 후)

```json
// .claude/settings.json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session-notice.cjs",
        "type": "command"
      }]
    }]
  }
}
```

**실행**: `node session-notice.cjs` ✅ 추가 설치 없이 바로 작동

---

## ✅ 검증

### 1. Node.js 버전별 호환성

```bash
# Node.js 18
node session-notice.cjs  ✅

# Node.js 20
node session-notice.cjs  ✅

# Node.js 22
node session-notice.cjs  ✅
```

### 2. "type": "module" 환경

```json
// package.json
{
  "type": "module"
}
```

```bash
# .js 파일 - ❌ CommonJS 문법 사용 시 오류
node session-notice.js

# .cjs 파일 - ✅ CommonJS로 명시적 인식
node session-notice.cjs
```

### 3. 크로스 플랫폼

- ✅ Windows: `node session-notice.cjs`
- ✅ macOS: `node session-notice.cjs`
- ✅ Linux: `node session-notice.cjs`

---

## 🚀 왜 다른 방식을 선택하지 않았나?

### ❌ Node.js 22.6+ Native TypeScript

```bash
node --experimental-strip-types session-notice.ts
```

**문제점**:
- **Node.js 22.6+ 필수** - 대부분 사용자가 18-20 사용
- **실험적 기능** - 프로덕션 불안정
- **타입 체크 안 함** - 런타임만 처리

**결론**: 호환성 문제로 **부적합**

---

### ❌ tsx

```bash
npx tsx session-notice.ts
```

**문제점**:
- **추가 설치 필요** - `npm install -D tsx`
- **목표 위반** - "추가 설치 없이" 원칙 위배

**결론**: 의존성 증가로 **부적합**

---

### ❌ 순수 JavaScript (ESM)

```javascript
// session-notice.js
export async function main() { ... }
```

**문제점**:
- **TypeScript 타입 안전성 상실**
- **개발 경험 저하**
- **유지보수 어려움**

**결론**: 품질 저하로 **부적합**

---

## 📊 최종 아키텍처

```
┌─────────────────────┐
│ 개발자 (MoAI Team)  │
│                     │
│ TypeScript 소스     │
│ session-notice.ts   │
└──────────┬──────────┘
           │
           │ bun run build:hooks
           │ (tsup)
           ▼
┌─────────────────────┐
│ 빌드된 파일          │
│ session-notice.cjs  │
│ (CommonJS)          │
└──────────┬──────────┘
           │
           │ npm publish / bun publish
           │
           ▼
┌─────────────────────┐
│ 사용자 (고객)        │
│                     │
│ node xxx.cjs  ✅    │
│                     │
│ Node.js 18+ 만 있으면 됨
└─────────────────────┘
```

---

## 🔄 CI/CD 파이프라인

### prepublishOnly Hook

```json
// package.json
{
  "scripts": {
    "prepublishOnly": "bun run ci",
    "ci": "bun run clean && bun run build && bun run check && bun run test:ci"
  }
}
```

**실행 순서**:
1. `clean` - 이전 빌드 정리
2. `build` - TypeScript 빌드 (CLI + hooks)
3. `check` - 타입 체크 + 린트
4. `test:ci` - 전체 테스트 + 커버리지

**보장**:
- ✅ 타입 안전성 검증됨
- ✅ 모든 테스트 통과
- ✅ 빌드된 .cjs 파일 포함

---

## 📝 베스트 프랙티스

### 1. 개발 시

```bash
# TypeScript 소스 수정
vim src/claude/hooks/session/session-notice.ts

# 타입 체크
bun run type-check

# 빌드
bun run build:hooks

# 테스트
node ../.claude/hooks/moai/session-notice.cjs
```

### 2. 배포 전

```bash
# 전체 CI 실행
bun run ci

# 패키지 검증
npm pack --dry-run
```

### 3. 사용자 지원

**FAQ: "훅이 실행되지 않아요"**

**체크리스트**:
1. Node.js 18+ 설치 확인: `node -v`
2. 파일 경로 확인: `ls .claude/hooks/moai/session-notice.cjs`
3. 실행 권한 확인: `ls -l .claude/hooks/moai/session-notice.cjs`
4. 수동 실행 테스트: `node .claude/hooks/moai/session-notice.cjs`

---

## 🎉 결론

**현재 .cjs 방식은 목표를 완벽히 달성했습니다**:

✅ **제로 의존성** - Node.js 18+ 만 있으면 됨
✅ **최대 호환성** - 모든 OS, 모든 Node 버전 (18+)
✅ **타입 안전** - 빌드 시 TypeScript 타입 체크
✅ **프로덕션 검증** - 대부분의 npm 패키지가 사용하는 표준 방식
✅ **사용자 편의** - `node xxx.cjs` 바로 실행 가능

**이것이 Node.js 생태계에서 추가 설치 없이 TypeScript를 배포하는 최선의 방법입니다.** 🚀