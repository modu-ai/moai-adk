# SPEC-HOOKS-001 인수 기준 (Acceptance Criteria)

> Pure JavaScript 훅 시스템 재설계

**SPEC ID**: HOOKS-001
**버전**: 0.0.1
**작성일**: 2025-10-12
**작성자**: @Goos

---

## ✅ 인수 기준 개요

이 SPEC은 다음 조건을 **모두 충족**해야 완료로 간주됩니다:

1. **크로스 플랫폼 실행**: 맥/윈도우/리눅스 모두 성공
2. **빌드 없는 배포**: `npm publish` 시 빌드 스텝 불필요
3. **성능 유지**: 훅 실행 < 100ms
4. **기존 테스트 통과**: Vitest 테스트 100% 통과
5. **파일 크기**: 각 훅 < 20KB
6. **외부 의존성**: 0개 (Node.js 내장 모듈만)

---

## 📋 Given-When-Then 시나리오

### Scenario 1: 크로스 플랫폼 실행

#### macOS 환경

**Given**:
- macOS 13+ 환경
- Node.js 18+ 설치됨
- `templates/.claude/hooks/alfred/policy-block.js` 존재

**When**:
```bash
node templates/.claude/hooks/alfred/policy-block.js
```

**Then**:
- ✅ 에러 없이 실행 완료
- ✅ exit code 0 반환
- ✅ 실행 시간 < 100ms

#### Windows 환경

**Given**:
- Windows 10+ 환경
- Node.js 18+ 설치됨
- `templates/.claude/hooks/alfred/policy-block.js` 존재

**When**:
```powershell
node templates\.claude\hooks\alfred\policy-block.js
```

**Then**:
- ✅ 에러 없이 실행 완료
- ✅ exit code 0 반환
- ✅ 실행 시간 < 100ms

#### Linux 환경

**Given**:
- Ubuntu 20.04+ 환경
- Node.js 18+ 설치됨
- `templates/.claude/hooks/alfred/policy-block.js` 존재

**When**:
```bash
node templates/.claude/hooks/alfred/policy-block.js
```

**Then**:
- ✅ 에러 없이 실행 완료
- ✅ exit code 0 반환
- ✅ 실행 시간 < 100ms

---

### Scenario 2: 위험 명령 차단

**Given**:
- `policy-block.js` 실행 가능
- stdin으로 위험 명령 전달

**When**:
```bash
echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}' | \
  node templates/.claude/hooks/alfred/policy-block.js
```

**Then**:
- ✅ "위험 명령이 감지되었습니다" 메시지 출력
- ✅ exit code 2 반환
- ✅ 실행 시간 < 100ms

---

### Scenario 3: 정상 명령 허용

**Given**:
- `policy-block.js` 실행 가능
- stdin으로 정상 명령 전달

**When**:
```bash
echo '{"tool_name":"Bash","tool_input":{"command":"git status"}}' | \
  node templates/.claude/hooks/alfred/policy-block.js
```

**Then**:
- ✅ 에러 없이 통과
- ✅ exit code 0 반환
- ✅ 실행 시간 < 100ms

---

### Scenario 4: JSON 파싱 에러 처리

**Given**:
- `policy-block.js` 실행 가능
- stdin으로 잘못된 JSON 전달

**When**:
```bash
echo 'invalid json' | \
  node templates/.claude/hooks/alfred/policy-block.js
```

**Then**:
- ✅ "Failed to parse input" 에러 메시지 출력
- ✅ exit code 1 반환
- ✅ 명확한 에러 위치 표시

---

### Scenario 5: moai init 실행

**Given**:
- 새 프로젝트 디렉토리
- `moai-adk` 패키지 설치됨

**When**:
```bash
mkdir test-project
cd test-project
moai init .
```

**Then**:
- ✅ `.claude/hooks/alfred/*.js` 파일 복사됨
- ✅ 각 파일 < 20KB
- ✅ 파일 실행 권한 644 (읽기/쓰기)
- ✅ `.claude/settings.json`에 훅 경로 설정됨

---

### Scenario 6: 빌드 없는 배포

**Given**:
- `moai-adk-ts` 프로젝트
- `templates/.claude/hooks/alfred/*.js` 존재

**When**:
```bash
npm publish
```

**Then**:
- ✅ 빌드 스텝 실행 안 됨 (tsup 호출 없음)
- ✅ Pure JS 파일 그대로 배포
- ✅ 배포 시간 < 10초

---

### Scenario 7: 성능 벤치마크

**Given**:
- 벤치마크 스크립트 작성됨
- 4개 훅 파일 모두 존재

**When**:
```bash
node benchmark.js
```

**Then**:
- ✅ policy-block: < 100ms
- ✅ pre-write-guard: < 100ms
- ✅ tag-enforcer: < 100ms
- ✅ session-notice: < 100ms

**벤치마크 스크립트 예시**:
```javascript
// benchmark.js
const { performance } = require('perf_hooks');

async function benchmark(hookPath, input) {
  const start = performance.now();
  // ... 훅 실행
  const duration = performance.now() - start;
  console.log(`${hookPath}: ${duration.toFixed(2)}ms`);
  return duration < 100;
}

// 각 훅 테스트
```

---

### Scenario 8: 기존 Vitest 테스트 통과

**Given**:
- TypeScript 소스 테스트 존재
- `moai-adk-ts/src/__tests__/claude/hooks/*.test.ts`

**When**:
```bash
cd moai-adk-ts
bun test src/__tests__/claude/hooks/
```

**Then**:
- ✅ 모든 테스트 통과 (100%)
- ✅ 커버리지 ≥ 85%
- ✅ 에러 0건

---

### Scenario 9: 파일 크기 제약

**Given**:
- Pure JS 파일 4개 작성 완료

**When**:
```bash
ls -lh templates/.claude/hooks/alfred/*.js
```

**Then**:
- ✅ policy-block.js: < 20KB
- ✅ pre-write-guard.js: < 20KB
- ✅ tag-enforcer.js: < 20KB
- ✅ session-notice.js: < 20KB

---

### Scenario 10: 외부 의존성 검증

**Given**:
- Pure JS 파일 작성 완료

**When**:
```bash
grep -r "require('.*')" templates/.claude/hooks/alfred/*.js | \
  grep -v "require('fs')" | \
  grep -v "require('path')" | \
  grep -v "require('process')"
```

**Then**:
- ✅ 결과 없음 (외부 패키지 사용 안 함)
- ✅ Node.js 내장 모듈만 사용 (fs, path, process)

---

## 🧪 수동 테스트 체크리스트

### 맥에서 실행

- [ ] policy-block.js 실행 성공
- [ ] pre-write-guard.js 실행 성공
- [ ] tag-enforcer.js 실행 성공
- [ ] session-notice.js 실행 성공

### 윈도우에서 실행 (GitHub Actions 또는 실제 환경)

- [ ] policy-block.js 실행 성공
- [ ] pre-write-guard.js 실행 성공
- [ ] tag-enforcer.js 실행 성공
- [ ] session-notice.js 실행 성공

### 리눅스에서 실행 (GitHub Actions 또는 Docker)

- [ ] policy-block.js 실행 성공
- [ ] pre-write-guard.js 실행 성공
- [ ] tag-enforcer.js 실행 성공
- [ ] session-notice.js 실행 성공

### 통합 테스트

- [ ] moai init 실행 → 훅 파일 복사 확인
- [ ] 훅 실행 → settings.json 경로 정상 작동
- [ ] 위험 명령 차단 → exit code 2 확인
- [ ] 정상 명령 허용 → exit code 0 확인

### 성능 테스트

- [ ] 각 훅 실행 시간 < 100ms
- [ ] 메모리 사용량 < 15MB
- [ ] CPU 사용률 < 10%

### 문서화

- [ ] README.md 업데이트 완료
- [ ] CHANGELOG.md 기록 완료
- [ ] 개발자 가이드 작성 완료

---

## 📊 최종 인수 기준 요약

| 기준 | 목표 | 측정 방법 | 통과 조건 |
|------|------|-----------|-----------|
| **크로스 플랫폼** | 100% 호환 | 맥/윈도우/리눅스 실행 | 모두 성공 |
| **성능** | < 100ms | 벤치마크 스크립트 | 모든 훅 통과 |
| **파일 크기** | < 20KB | `ls -lh` | 모든 파일 통과 |
| **외부 의존성** | 0개 | `grep require` | 내장 모듈만 |
| **기존 테스트** | 100% 통과 | `bun test` | 에러 0건 |
| **빌드 시간** | 0초 | `npm publish` | 빌드 스텝 없음 |

---

## 🚦 승인 조건

다음 조건을 **모두** 만족해야 SPEC-HOOKS-001이 완료됩니다:

1. ✅ 위 10개 시나리오 모두 통과
2. ✅ 수동 테스트 체크리스트 100% 완료
3. ✅ 최종 인수 기준 표 모두 달성
4. ✅ 코드 리뷰 2회 통과
5. ✅ 문서화 완료 (README + CHANGELOG)

---

**작성자**: @Goos
**검토자**: (TBD)
**승인일**: (TBD)
**테스트 실행일**: (TBD)
