# MoAI-ADK v0.2.10 릴리스 노트

🗿 **MoAI-ADK v0.2.10** - 설정 스키마 개선 및 자동 버전 관리

📅 **배포일**: 2025년 10월 7일
📦 **패키지**: [moai-adk@0.2.10](https://www.npmjs.com/package/moai-adk)
🏷️ **Git 태그**: `v0.2.10`

---

## 🎯 핵심 요약

이번 릴리스는 **설정 스키마 개선**과 **자동 버전 관리** 시스템을 도입하여, 버전 정보 혼란과 하드코딩 문제를 근본적으로 해결합니다.

### 주요 개선사항

- ✨ **moai.version 필드 신설**: config.json에서 패키지 버전을 명확하게 추적
- 🔄 **자동 버전 주입**: package.json 기반 동적 버전 관리 (하드코딩 제로)
- 🔙 **하위 호환성 보장**: 3단계 Fallback으로 기존 프로젝트 영향 없음
- 📋 **업데이트 흐름 강화**: `/alfred:9-update` Phase 4.5에서 자동 버전 동기화

---

## ✨ 새로운 기능

### 1. 설정 스키마 개선

**해결된 문제**: 기존 스키마에서는 `project.version`이 사용자 프로젝트 버전과 moai-adk 패키지 버전을 모두 의미해 혼란을 야기했습니다.

**새 스키마**:
```json
{
  "moai": {
    "version": "0.2.10"  // 신규: moai-adk 패키지 버전
  },
  "project": {
    "name": "MyProject",
    "version": "0.1.0",  // 사용자 프로젝트 버전 (moai.version과 별개)
    "mode": "team"
  }
}
```

**장점**:
- 🎯 명확한 분리: `moai.version` (패키지) vs `project.version` (사용자 프로젝트)
- 🔍 session-start-hook에서 정확한 버전 표시
- 📦 올바른 업데이트 감지 및 알림

### 2. 자동 버전 관리

**하드코딩 제로 원칙**: 모든 버전 정보를 동적으로 로드합니다.

#### `moai init` 실행 시:
```typescript
// config-builder.ts
import packageJson from '../../../package.json';

public buildConfig(answers: InitAnswers): MoAIConfig {
  return {
    moai: {
      version: packageJson.version  // 자동 주입: "0.2.10"
    },
    // ...
  };
}
```

#### `/alfred:9-update` 실행 시:
새로 추가된 **Phase 4.5**에서 `moai.version`을 자동 업데이트:
```bash
# Step 4.5.1: 설치된 버전 감지
npm list moai-adk --depth=0 | grep moai-adk
→ moai-adk@0.3.0

# Step 4.5.3: config.json 업데이트
config.moai.version = "0.3.0"
```

### 3. 하위 호환성 보장

**3단계 우선순위 Fallback**으로 기존 프로젝트에 영향 없음:

```typescript
// session-notice/utils.ts - getMoAIVersion()
// 1순위: moai.version (신규 스키마)
if (config.moai?.version) return config.moai.version;

// 2순위: project.version (구 스키마 - 하위 호환성)
if (config.project?.version) return config.project.version;

// 3순위: node_modules/moai-adk/package.json (최후 수단)
const packageJson = require('moai-adk/package.json');
return packageJson.version;
```

**결과**: 기존 사용자에게 Breaking Change 없음.

---

## 🔄 변경된 파일

### 핵심 구현
- `templates/.moai/config.json`: `moai.version` 필드 추가
- `src/cli/config/config-builder.ts`: package.json에서 버전 자동 주입
- `src/claude/hooks/session-notice/utils.ts`: 우선순위 기반 버전 감지
- `src/claude/hooks/session-notice/types.ts`: 명확화 주석 추가

### 문서
- `.claude/commands/alfred/9-update.md`: Phase 4.5 추가 (moai.version 자동 업데이트)
- `CHANGELOG.md`: v0.2.10 항목
- `package.json`: 버전 0.2.10으로 변경

### 배포
- `scripts/publish.sh`: 신규 - NPM 자동 배포 스크립트

---

## 🚀 마이그레이션 가이드

### 신규 프로젝트
조치 불필요. `moai init`이 자동으로 `moai.version`을 포함한 config를 생성합니다.

### 기존 프로젝트

**방법 1: 자동 업데이트** (권장)
```bash
# moai-adk 패키지 업데이트 및 config 자동 업데이트
/alfred:9-update
```

**방법 2: 수동 업데이트**
`.moai/config.json` 파일 수정:
```json
{
  "moai": {
    "version": "0.2.10"  // 이 필드 추가
  },
  "project": {
    // 기존 필드 유지
  }
}
```

**하위 호환성 참고**: 업데이트하지 않아도 시스템이 `project.version`으로 폴백하므로 정상 작동합니다.

---

## 🔧 기술 세부사항

### Session-Start Hook 동작

**v0.2.10 이전**:
```
📦 버전: v0.0.3 (incorrect - project version 표시)
```

**v0.2.10 이후**:
```
📦 버전: v0.2.10 (최신) (correct - package version 표시)
```

### 버전 감지 우선순위

```
getMoAIVersion()
  ↓
1순위: config.moai.version 확인
  ↓ (없으면)
2순위: config.project.version 확인 (구 스키마)
  ↓ (없으면)
3순위: node_modules/moai-adk/package.json 확인
  ↓ (없으면)
return 'unknown'
```

### 빌드 검증

모든 품질 게이트 통과:
```bash
✅ TypeScript 타입 체크: PASSED
✅ Biome lint: PASSED
✅ 테스트: PASSED
✅ 빌드: SUCCESS (dist/index.js, dist/index.cjs)
✅ Hook 빌드: SUCCESS (session-notice.cjs)
```

---

## 📦 설치

### NPM
```bash
npm install moai-adk@0.2.10
# 또는
npm install moai-adk@latest
```

### Bun
```bash
bun add moai-adk@0.2.10
```

### 기존 설치 업데이트
```bash
npm update moai-adk
# 그 다음 업데이트 명령 실행
/alfred:9-update
```

---

## 🐛 버그 수정

- **수정**: 패키지 버전과 프로젝트 버전 간 의미 혼란 해결
- **수정**: config-builder.ts의 하드코딩된 버전 "0.0.1" 제거
- **수정**: session-notice hook의 잘못된 버전 표시 (0.0.3) 수정
- **수정**: 자동 버전 업데이트 메커니즘 부재 문제 해결

---

## 📚 문서 업데이트

- **9-update.md**: `moai.version` 자동 업데이트를 위한 Phase 4.5 추가
- **CHANGELOG.md**: v0.2.10 전체 변경 내역
- **RELEASE_NOTES.md**: 본 문서

---

## 🔗 링크

- 📦 **NPM 패키지**: https://www.npmjs.com/package/moai-adk
- 🐙 **GitHub 저장소**: https://github.com/modu-ai/moai-adk
- 🐛 **이슈 트래커**: https://github.com/modu-ai/moai-adk/issues
- 📖 **문서**: https://moai-adk.vercel.app

---

## 🙏 크레딧

**핵심 기여자**:
- @Goos - 설정 스키마 재설계, 자동 버전 관리 구현
- Alfred SuperAgent - 오케스트레이션 및 품질 보증

**특별 감사**:
- cc-manager agent - 9-update.md Phase 4.5 구현
- trust-checker agent - TRUST 5원칙 검증

---

**전체 변경 내역**: https://github.com/modu-ai/moai-adk/compare/v0.2.6...v0.2.10

---

🗿 MoAI-ADK v0.2.10으로 생성
