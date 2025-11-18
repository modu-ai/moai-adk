# Figma MCP 에러 코드 매핑 및 해결책

**문서 목적**: Figma MCP 도구 사용 시 발생하는 에러를 HTTP 상태 코드와 원인, 해결책으로 정리
**작성일**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (Section 5)

---

## 목차

1. [HTTP 상태 코드 매핑](#1-http-상태-코드-매핑)
2. [도구별 에러 타입](#2-도구별-에러-타입)
3. [Rate Limit 에러 처리](#3-rate-limit-에러-처리)
4. [실전 에러 디버깅 가이드](#4-실전-에러-디버깅-가이드)
5. [복구 전략 및 Best Practice](#5-복구-전략-및-best-practice)

---

## 1. HTTP 상태 코드 매핑

### 1.1 모든 MCP 도구 공통 에러

| HTTP Code | 상태 | 에러 타입 | 사용자 메시지 | 즉시 재시도 | 백오프 재시도 |
|----------|------|---------|-------------|----------|----------|
| **200** | ✅ OK | - | - | - | - |
| **201** | ✅ Created | - | - | - | - |
| **204** | ✅ No Content | - | - | - | - |
| **400** | ❌ Bad Request | `invalid_request_error` | "요청 파라미터가 잘못되었습니다" | ❌ 아니오 | ❌ 아니오 |
| **401** | ❌ Unauthorized | `authentication_error` | "API 키가 만료되었습니다" | ❌ 아니오 | ✅ 예 (새 토큰 발급 후) |
| **403** | ❌ Forbidden | `forbidden_error` | "이 파일에 접근할 권한이 없습니다" | ❌ 아니오 | ❌ 아니오 |
| **404** | ❌ Not Found | `not_found_error` | "파일 또는 노드를 찾을 수 없습니다" | ❌ 아니오 | ❌ 아니오 |
| **429** | ⚠️ Too Many Requests | `rate_limit_error` | "API 호출 제한에 도달했습니다" | ❌ 아니오 | ✅ 예 (Exponential Backoff) |
| **500** | ❌ Server Error | `server_error` | "Figma 서버 오류입니다" | ❌ 아니오 | ✅ 예 (지수 백오프) |
| **502** | ❌ Bad Gateway | `server_error` | "게이트웨이 오류입니다" | ❌ 아니오 | ✅ 예 |
| **503** | ❌ Service Unavailable | `service_unavailable` | "Figma 서비스가 일시적으로 중단되었습니다" | ❌ 아니오 | ✅ 예 |
| **504** | ❌ Gateway Timeout | `server_error` | "요청이 시간 초과되었습니다" | ❌ 아니오 | ✅ 예 |

---

### 1.2 400 Bad Request - 파라미터 오류

| 시나리오 | 파라미터 | 잘못된 예 | 올바른 예 | 에러 메시지 |
|---------|---------|---------|---------|-----------|
| **파일 키 형식** | fileKey | "abc123" (짧음) | "abc123XYZ456def789012" | Invalid file key format |
| **상대 경로 사용** | localPath | "./assets/images" | "/Users/dev/assets/images" | Path for asset writes is invalid |
| **디렉토리 권한** | localPath | "/root/protected" (EACCES) | "/Users/dev/assets" | Path for asset writes is invalid |
| **노드 ID 형식** | nodeId | "1234" (콜론 없음) | "1234:5678" | Invalid node ID format |
| **pngScale 범위** | pngScale | 5 (범위 초과) | 2 | Invalid scale value |
| **format 값** | format | "webp" (미지원) | "png", "svg" | Unsupported format |

---

### 1.3 401 Unauthorized - 인증 오류

| 원인 | 시나리오 | 해결책 | 예상 시간 |
|------|--------|--------|---------|
| **토큰 만료** | 30일 이상 사용 안 함 | Figma 설정 → Access tokens → 새 토큰 생성 | 5분 |
| **토큰 무효** | 잘못된 토큰 입력 | API 키 복사 재확인 | 5분 |
| **토큰 삭제됨** | 다른 기기에서 삭제 | 새 토큰 생성 | 5분 |
| **권한 부족** | 읽기 전용 토큰으로 쓰기 시도 | 쓰기 권한이 있는 토큰 발급 | 5분 |
| **헤더 누락** | X-Figma-Token 헤더 없음 | 헤더 추가: `X-Figma-Token: figd_...` | 1분 |

**해결 단계** (Figma 웹사이트):
1. figma.com 로그인
2. Settings (톱니바퀴 아이콘)
3. Account 탭
4. Personal access tokens
5. "Create a new token" 클릭
6. 복사하여 환경변수 설정

```bash
# .env 파일
FIGMA_API_KEY=figd_your_new_token_here

# 환경변수 확인
echo $FIGMA_API_KEY
```

---

### 1.4 403 Forbidden - 권한 없음

| 시나리오 | 원인 | 해결책 | 소유자 작업 |
|---------|------|--------|-----------|
| **파일 비공개** | 개인 파일 | 파일 소유자에게 공유 요청 | 파일 → 공유 → 이메일 추가 |
| **팀 권한 부족** | 팀 파일 접근 불가 | 팀 관리자에게 권한 요청 | 팀 → 멤버 관리 |
| **파일 권한 제한** | 읽기 권한만 있는데 쓰기 시도 | 편집 권한으로 업그레이드 요청 | 파일 → 공유 → 권한 변경 |

**공유 방법** (소유자가 해야 할 작업):
```
Figma 파일 → 상단 오른쪽 "Share"
→ "+ Add people"
→ 이메일 입력
→ 권한 선택 (View / Edit / Manage)
→ "Invite"
```

---

### 1.5 404 Not Found - 리소스 미존재

| 파라미터 | 확인 방법 | 원인 | 해결책 |
|---------|---------|------|--------|
| **fileKey** | Figma URL 확인 | 파일 삭제됨 또는 잘못된 키 | `https://figma.com/file/{fileKey}/...` URL에서 추출 |
| **nodeId** | get_figma_data로 목록 조회 | 노드 삭제됨 또는 잘못된 ID | 먼저 `get_figma_data(fileKey)` 실행 후 유효한 노드 ID 확인 |

**파일 키 추출 방법** (URL에서):
```
https://www.figma.com/file/abc123XYZ456def789012/Project-Name
                           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                   fileKey (복사)
```

**노드 ID 추출 방법** (Figma 클라이언트):
```
1. Figma 파일 열기
2. 노드(Frame, Component 등) 선택
3. 마우스 오른쪽 클릭 → "Copy node ID"
4. 또는 keyboard: Cmd+Shift+C (Mac) / Ctrl+Shift+C (Windows)
```

---

## 2. 도구별 에러 타입

### 2.1 get_figma_data 에러

| 에러 | HTTP Code | 원인 | 예방 방법 |
|------|----------|------|----------|
| **Invalid file key** | 400 | 파일 키 형식 오류 (22자 아님) | 파일 URL에서 복사 후 사용 |
| **File not found** | 404 | 파일 삭제됨 또는 잘못된 키 | Figma 웹에서 파일 존재 확인 |
| **Access denied** | 403 | 파일 접근 권한 없음 | 파일 소유자에게 공유 요청 |
| **Unauthorized** | 401 | API 토큰 만료 | 새 토큰 발급 |
| **Rate limit exceeded** | 429 | 분당 60회 초과 | 1초 이상 대기 후 재시도 |

**이미지 다운로드는 별도 API 호출이므로 429 에러 분리 필요**

---

### 2.2 download_figma_images 에러

| 에러 메시지 | HTTP Code | 원인 | 해결책 | 우회 방법 |
|-----------|----------|------|--------|----------|
| **Path for asset writes is invalid** | 400 | 상대 경로 / 디렉토리 권한 문제 | `path.resolve()` 사용 절대 경로 | 다른 디렉토리 시도 |
| **Image base64 format error** | 400 | 이미지 인코딩 실패 / pngScale 너무 큼 | pngScale을 1 또는 2로 감소 | SVG 형식으로 변경 |
| **Node not found** | 404 | 존재하지 않는 노드 ID | `get_figma_data` 실행 후 유효한 ID 확인 | 부모 노드 사용 |
| **No images found** | 400 | 노드에 이미지 없음 | 이미지가 있는 노드 선택 | 다른 노드 시도 |
| **Rate limit exceeded** | 429 | 분당 30회 초과 (이미지 렌더링) | 2초 이상 대기 후 재시도 | 청킹: 5개씩 처리 |
| **Unauthorized** | 401 | API 토큰 문제 | 토큰 갱신 | 새 세션 시작 |

**경로 오류 해결** (TypeScript):

```typescript
// ❌ 오류 발생
await figmaService.downloadImages(
  "abc123XYZ",
  "./assets/images",  // 상대 경로
  nodes
);

// ✅ 올바른 방법
import path from 'path';
const absolutePath = path.resolve(process.cwd(), './assets/images');
await figmaService.downloadImages(
  "abc123XYZ",
  absolutePath,  // 절대 경로
  nodes
);

// 또는
import * as fs from 'fs/promises';
const absolutePath = '/Users/dev/project/assets/images';
await fs.mkdir(absolutePath, { recursive: true });  // 디렉토리 생성
await figmaService.downloadImages("abc123XYZ", absolutePath, nodes);
```

---

### 2.3 export_node_as_image 에러

| 에러 | HTTP Code | 원인 | 해결책 |
|------|----------|------|--------|
| **Node not found** | 404 | 노드 ID 오류 | 노드 ID 재확인 |
| **Format not supported** | 400 | 형식 이름 오류 | "PNG", "JPG", "SVG", "PDF" 중 선택 |
| **base64 encode failed** | 500 | 노드 타입 미지원 (GROUP, TEXT 등) | FRAME, COMPONENT, RECTANGLE 등 사용 |
| **Unauthorized** | 401 | API 토큰 문제 | 토큰 갱신 |

**주의**: 반환값이 base64 문자열이므로 매우 큼 (10MB+ 가능)

---

### 2.4 변수 조회 API (REST) 에러

| 에러 | HTTP Code | 원인 | 해결책 |
|------|----------|------|--------|
| **Invalid file key** | 400 | 파일 키 형식 오류 | 22자 알파벳+숫자 확인 |
| **File not found** | 404 | 파일 삭제됨 | 파일 존재 여부 확인 |
| **Invalid token** | 401 | API 토큰 만료 | 새 토큰 발급 |
| **Access denied** | 403 | 파일 접근 권한 없음 | 파일 소유자에게 공유 요청 |
| **Rate limit exceeded** | 429 | 분당 100회 초과 | 지수 백오프 재시도 |

---

## 3. Rate Limit 에러 처리

### 3.1 Figma API Rate Limit 정책

| 엔드포인트 | 제한 | 예시 |
|-----------|------|------|
| **일반 API** (get_figma_data, 변수 조회) | 분당 60회 | 1초마다 1회, 분당 60회 |
| **이미지 렌더링** (download_figma_images, export_node_as_image) | 분당 30회 | 2초마다 1회, 분당 30회 |
| **변수 API** | 분당 100회 | 가장 여유로움 |

### 3.2 429 에러 응답 구조

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 61
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1234567890

{
  "err": "ERR_RATE_LIMIT",
  "status": 429,
  "message": "Rate limit exceeded"
}
```

**Retry-After 헤더**: 서버가 권장하는 재시도 대기 시간 (초)

### 3.3 Exponential Backoff 구현 (권장)

```typescript
import pLimit from 'p-limit';

async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  initialDelayMs: number = 1000
): Promise<T> {
  let lastError: Error;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error: any) {
      lastError = error;

      // 429 Rate Limit 감지
      if (error.response?.status === 429) {
        // 서버의 Retry-After 헤더 우선 사용
        const retryAfter = error.response.headers['retry-after'];
        const delayMs = retryAfter
          ? parseInt(retryAfter) * 1000
          : initialDelayMs * Math.pow(2, attempt);

        console.log(
          `🔄 Rate limited. Waiting ${delayMs}ms before retry ${attempt + 1}/${maxRetries}...`
        );
        await new Promise(resolve => setTimeout(resolve, delayMs));
        continue;
      }

      // 5xx 서버 에러 (3회 재시도)
      if (error.response?.status >= 500) {
        const delayMs = initialDelayMs * Math.pow(2, attempt);
        console.log(
          `⚠️ Server error (${error.response.status}). Waiting ${delayMs}ms before retry...`
        );
        await new Promise(resolve => setTimeout(resolve, delayMs));
        continue;
      }

      // 4xx 클라이언트 에러 (재시도 불가, 즉시 throw)
      throw error;
    }
  }

  throw lastError;
}

// 사용 예제
const fileData = await retryWithBackoff(
  () => figmaService.getRawFile("abc123XYZ"),
  3,     // 최대 3회 재시도
  1000   // 초기 대기: 1초
);
// Attempt 1: 실패 → 1초 대기
// Attempt 2: 실패 → 2초 대기
// Attempt 3: 실패 → 4초 대기
// 총 시간: 1 + 2 + 4 = 7초
```

### 3.4 청킹 전략 (Rate Limit 회피)

```typescript
async function processWithChunking<T>(
  items: T[],
  processFunc: (item: T) => Promise<any>,
  chunkSize: number = 10,
  delayBetweenChunksMs: number = 1000
): Promise<any[]> {
  const results = [];

  for (let i = 0; i < items.length; i += chunkSize) {
    const chunk = items.slice(i, i + chunkSize);
    console.log(`📦 Processing chunk ${Math.floor(i / chunkSize) + 1}...`);

    // 청크 내에서는 병렬 처리
    const chunkResults = await Promise.all(
      chunk.map(item => processFunc(item))
    );
    results.push(...chunkResults);

    // 다음 청크 전 대기 (Rate Limit 회피)
    if (i + chunkSize < items.length) {
      console.log(`⏳ Waiting ${delayBetweenChunksMs}ms before next chunk...`);
      await new Promise(resolve => setTimeout(resolve, delayBetweenChunksMs));
    }
  }

  return results;
}

// 사용 예제: 100개 노드 이미지 다운로드 (30회/분 제한 때문에 문제)
const nodeIds = [...]; // 100개

await processWithChunking(
  nodeIds,
  async (nodeId) => {
    return await figmaService.downloadImages(
      "abc123XYZ",
      "./assets",
      [{ nodeId, fileName: `${nodeId}.png` }]
    );
  },
  5,      // 5개씩 묶기 (병렬)
  2000    // 각 청크 사이 2초 대기 (분당 30회 제한 충족)
);
```

---

## 4. 실전 에러 디버깅 가이드

### 4.1 "Path for asset writes is invalid" 해결

**에러 시나리오**:
```
Error: Path for asset writes is invalid
  at FigmaService.downloadImages()
```

**진단 단계**:

```typescript
import path from 'path';
import fs from 'fs/promises';

async function debugPathError(localPath: string) {
  console.log('📍 원본 경로:', localPath);

  // 1단계: 상대 경로 vs 절대 경로
  if (!path.isAbsolute(localPath)) {
    console.warn('⚠️ 상대 경로 감지. 절대 경로로 변환...');
    localPath = path.resolve(process.cwd(), localPath);
    console.log('✅ 변환된 경로:', localPath);
  }

  // 2단계: 디렉토리 존재 확인
  try {
    await fs.access(localPath);
    console.log('✅ 디렉토리 접근 가능');
  } catch (error: any) {
    if (error.code === 'ENOENT') {
      console.warn('⚠️ 디렉토리 미존재. 생성 중...');
      await fs.mkdir(localPath, { recursive: true });
      console.log('✅ 디렉토리 생성 완료');
    } else if (error.code === 'EACCES') {
      console.error('❌ 권한 부족. 다른 디렉토리 시도 필요');
      return false;
    }
  }

  // 3단계: 쓰기 권한 테스트
  try {
    const testFile = path.join(localPath, '.test-write');
    await fs.writeFile(testFile, 'test');
    await fs.unlink(testFile);
    console.log('✅ 쓰기 권한 확인됨');
    return true;
  } catch (error) {
    console.error('❌ 쓰기 권한 없음:', error);
    return false;
  }
}

// 사용
const isValid = await debugPathError('./assets/images');
if (isValid) {
  await figmaService.downloadImages(
    "abc123XYZ",
    path.resolve(process.cwd(), './assets/images'),
    nodes
  );
}
```

---

### 4.2 "Image base64 format error" 해결

**에러 시나리오**:
```
Error: Image base64 format error
  at FigmaService.downloadImages()
```

**원인별 해결책**:

| 원인 | 징후 | 해결책 |
|------|------|--------|
| **pngScale 너무 큼** | 큰 이미지 노드 | pngScale=1 또는 2로 감소 |
| **노드 타입 미지원** | TEXT, GROUP 노드 | FRAME, COMPONENT, RECTANGLE 시도 |
| **이미지 손상** | 특정 노드만 실패 | 다른 노드 시도 또는 재생성 |
| **네트워크 오류** | 간헐적 실패 | Exponential Backoff 재시도 |

```typescript
async function debugImageError(
  fileKey: string,
  nodeIds: string[],
  localPath: string
) {
  for (const nodeId of nodeIds) {
    try {
      console.log(`🖼️ Downloading ${nodeId}...`);

      // 시도 1: 기본 설정
      try {
        await figmaService.downloadImages(fileKey, localPath, [
          { nodeId, fileName: `${nodeId}.png` }
        ]);
        console.log(`✅ Success: ${nodeId}`);
        continue;
      } catch (error: any) {
        if (!error.message.includes('base64')) throw error;
        console.warn(`⚠️ base64 error. Trying pngScale=1...`);
      }

      // 시도 2: 스케일 감소
      try {
        await figmaService.downloadImages(
          fileKey,
          localPath,
          [{ nodeId, fileName: `${nodeId}.png` }],
          { pngScale: 1 }
        );
        console.log(`✅ Success with pngScale=1: ${nodeId}`);
        continue;
      } catch (error: any) {
        if (!error.message.includes('base64')) throw error;
        console.warn(`⚠️ base64 error. Trying SVG format...`);
      }

      // 시도 3: SVG로 변경
      try {
        await figmaService.downloadImages(
          fileKey,
          localPath,
          [{ nodeId, fileName: `${nodeId}.svg` }],
          { format: 'svg' }
        );
        console.log(`✅ Success with SVG: ${nodeId}`);
      } catch (error) {
        console.error(`❌ All attempts failed for ${nodeId}:`, error.message);
      }

    } catch (error) {
      console.error(`❌ Unexpected error for ${nodeId}:`, error);
    }
  }
}

// 사용
await debugImageError(
  "abc123XYZ",
  ["1234:5678", "1234:5679", "1234:5680"],
  "/Users/dev/assets"
);
```

---

### 4.3 "File not found (404)" 해결

**에러 시나리오**:
```
Error: File not found (404)
  at FigmaService.getRawFile()
```

**진단 단계**:

```typescript
async function debugFileNotFound(fileKey: string, apiToken: string) {
  console.log('🔍 Diagnosing 404 error for fileKey:', fileKey);

  // 1단계: 파일 키 형식 검증
  if (!/^[a-zA-Z0-9]{22}$/.test(fileKey)) {
    console.error('❌ Invalid file key format');
    console.log('   Expected: 22 alphanumeric characters');
    console.log('   Example: abc123XYZ456def789012');
    return false;
  }
  console.log('✅ File key format valid');

  // 2단계: API 접근 확인
  try {
    const response = await fetch(
      `https://api.figma.com/v1/files/${fileKey}`,
      {
        headers: { 'X-Figma-Token': apiToken }
      }
    );

    if (response.status === 404) {
      console.error('❌ File not found. Possible causes:');
      console.log('   1. File has been deleted');
      console.log('   2. File key is incorrect');
      console.log('   3. File access permission revoked');
      return false;
    }

    if (!response.ok) {
      console.error(`❌ API error: ${response.status}`);
      return false;
    }

    console.log('✅ File accessible');
    return true;

  } catch (error) {
    console.error('❌ Network error:', error);
    return false;
  }
}

// 사용
const isValid = await debugFileNotFound("abc123XYZ", process.env.FIGMA_API_KEY!);
```

---

## 5. 복구 전략 및 Best Practice

### 5.1 에러 처리 우선순위

```typescript
enum ErrorSeverity {
  FATAL = 0,      // 즉시 중단
  BLOCKING = 1,   // 사용자 개입 필요
  RETRYABLE = 2,  // 자동 재시도
  WARNING = 3,    // 경고만 표시
  INFO = 4        // 정보 로깅
}

async function handleError(
  error: any
): Promise<ErrorSeverity> {
  const status = error.response?.status;

  // FATAL: 즉시 중단
  if (status === 400 && error.message.includes('Path for asset writes')) {
    console.error('🔴 FATAL: 경로 오류. 파라미터 확인 필요');
    return ErrorSeverity.FATAL;
  }

  // BLOCKING: 사용자 개입 필요
  if (status === 403) {
    console.error('🔴 BLOCKING: 파일 접근 권한 없음. 소유자에게 공유 요청');
    return ErrorSeverity.BLOCKING;
  }

  if (status === 401) {
    console.error('🔴 BLOCKING: API 토큰 만료. 새 토큰 발급 필요');
    return ErrorSeverity.BLOCKING;
  }

  // RETRYABLE: 자동 재시도
  if (status === 429 || status >= 500) {
    console.warn('🟡 RETRYABLE: Exponential Backoff 재시도');
    return ErrorSeverity.RETRYABLE;
  }

  // WARNING: 경고
  if (status === 404) {
    console.warn('🟡 WARNING: 리소스 미존재. 데이터 재확인 필요');
    return ErrorSeverity.WARNING;
  }

  // INFO: 정보
  console.log('ℹ️ INFO: 기타 에러');
  return ErrorSeverity.INFO;
}
```

### 5.2 재시도 정책 결정표

```typescript
interface RetryPolicy {
  maxRetries: number;
  initialDelayMs: number;
  backoffMultiplier: number;
  maxDelayMs: number;
}

const retryPolicies = {
  // 429 Rate Limit: 지수 백오프
  RATE_LIMIT: {
    maxRetries: 3,
    initialDelayMs: 1000,
    backoffMultiplier: 2,
    maxDelayMs: 60000  // 최대 1분
  },

  // 5xx Server Error: 보수적 백오프
  SERVER_ERROR: {
    maxRetries: 3,
    initialDelayMs: 2000,
    backoffMultiplier: 2,
    maxDelayMs: 30000  // 최대 30초
  },

  // 네트워크 타임아웃
  TIMEOUT: {
    maxRetries: 2,
    initialDelayMs: 3000,
    backoffMultiplier: 1.5,
    maxDelayMs: 10000  // 최대 10초
  }
};
```

### 5.3 에러 로깅 전략

```typescript
interface ErrorLog {
  timestamp: Date;
  errorType: string;
  httpStatus?: number;
  message: string;
  stack?: string;
  context: {
    fileKey?: string;
    nodeId?: string;
    operation: string;
  };
  recoveryAction: string;
  success?: boolean;
}

async function logAndRecover(
  error: any,
  context: any
): Promise<ErrorLog> {
  const log: ErrorLog = {
    timestamp: new Date(),
    errorType: error.name,
    httpStatus: error.response?.status,
    message: error.message,
    stack: error.stack,
    context,
    recoveryAction: 'pending',
    success: false
  };

  // 심각도별 처리
  const severity = await handleError(error);

  switch (severity) {
    case ErrorSeverity.FATAL:
      log.recoveryAction = 'MANUAL_INTERVENTION_REQUIRED';
      break;
    case ErrorSeverity.BLOCKING:
      log.recoveryAction = 'USER_ACTION_REQUIRED';
      break;
    case ErrorSeverity.RETRYABLE:
      log.recoveryAction = 'AUTO_RETRY_SCHEDULED';
      break;
  }

  // 파일에 로깅
  await fs.appendFile(
    './.moai/logs/figma-errors.log',
    JSON.stringify(log) + '\n'
  );

  return log;
}
```

---

## 요약: 에러 복구 체크리스트

### 에러 발생 시 순서도

```
에러 발생
  ↓
상태 코드 확인
  ├─ 400 (Bad Request)
  │  └─ 파라미터 오류 → 파라미터 검증 후 수정
  │
  ├─ 401 (Unauthorized)
  │  └─ 토큰 만료 → 새 토큰 발급
  │
  ├─ 403 (Forbidden)
  │  └─ 권한 없음 → 파일 소유자에게 공유 요청
  │
  ├─ 404 (Not Found)
  │  └─ 리소스 미존재 → 파일/노드 ID 재확인
  │
  ├─ 429 (Rate Limit)
  │  └─ 호출 제한 → Exponential Backoff 재시도
  │
  └─ 5xx (Server Error)
     └─ 서버 오류 → 지수 백오프로 재시도
```

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (Section 5)
