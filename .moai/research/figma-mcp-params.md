# Figma MCP 도구 파라미터 검증 문서

**문서 목적**: Figma MCP 도구별 파라미터 명세, 제약사항, 에러 조건 정리
**작성일**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (조사 결과)

---

## 목차

1. [get_figma_data - 디자인 데이터 조회](#1-get_figma_data---디자인-데이터-조회)
2. [download_figma_images - 이미지 다운로드](#2-download_figma_images---이미지-다운로드)
3. [export_node_as_image - 노드 이미지 내보내기](#3-export_node_as_image---노드-이미지-내보내기)
4. [변수 조회 API (REST)](#4-변수-조회-api-rest)
5. [파라미터 제약사항 및 한계](#5-파라미터-제약사항-및-한계)
6. [에러 코드 매핑](#6-에러-코드-매핑)

---

## 1. get_figma_data - 디자인 데이터 조회

**도구 출처**: Figma Context MCP (`/glips/figma-context-mcp`)
**기능**: Figma 파일 또는 특정 노드의 구조화된 디자인 데이터 추출

### 1.1 전체 파라미터 명세

| 파라미터 | 타입 | 필수 | 기본값 | 설명 | 형식/제약 |
|---------|------|------|--------|------|---------|
| `fileKey` | string | ✅ | - | Figma 파일 키 | 22자 알파벳+숫자 |
| `nodeId` | string | ❌ | - | 특정 노드 ID (선택 시 해당 노드만 추출) | "1234:5678" 또는 "1234-5678" 형식 |
| `depth` | number | ❌ | 전체 | 트리 탐색 깊이 (1-10, 성능 조절용) | 1=메타만, 10=전체 |

### 1.2 nodeId 형식 및 추출 방법

**형식**:
```
"1234:5678"  // Figma 공식 형식 (권장)
"1234-5678"  // 대체 형식 (호환성)
```

**추출 방법** (Figma UI에서):
1. Figma 파일 열기
2. 노드 선택 (Frame, Component, Group 등)
3. 마우스 오른쪽 클릭 → "Copy node ID"
4. 또는 Figma URL에서 추출: `figma.com/file/{fileKey}/page-{nodeId}`

**파일 키 추출** (URL에서):
```
https://www.figma.com/file/abc123XYZ456/Design-System
                           ^^^^^^^^^^^^^^^^
                              fileKey
```

### 1.3 Depth 파라미터 가이드

| depth | 설명 | 사용 사례 | 성능 | 데이터 크기 |
|-------|------|---------|------|-----------|
| 1 | 파일 메타데이터만 | 파일 목록 확인, 페이지 개수 | 빠름 | <100KB |
| 2 | 페이지 + 주요 프레임 | 페이지 구조 파악 | 빠름 | 100KB-500KB |
| 3 | 프레임 + 컴포넌트 | 컴포넌트 라이브러리 탐색 | 중간 | 500KB-2MB |
| 5 | 전체 레이아웃 구조 | 디자인 시스템 분석 | 느림 | 2MB-10MB |
| 10+ | 모든 속성 포함 | 완전 코드 생성 | 매우 느림 | 10MB+ |

**권장**:
- 초기 탐색: `depth=1` 또는 `depth=2`
- 상세 분석: `depth=3` 또는 `depth=5`
- 완전 추출: `depth=10` (Rate Limit 주의)

### 1.4 반환값 구조

```typescript
{
  metadata: {
    name: string;                    // 파일명
    components: Record<string, any>; // 컴포넌트 정의
    componentSets: Record<string, any>;
  };
  nodes: Array<{
    id: string;                      // 노드 ID
    name: string;                    // 노드명
    type: string;                    // FRAME, COMPONENT, TEXT, etc.
    layout: string;                  // 레이아웃 ID
    children: Array<any>;            // 자식 노드
  }>;
  globalVars: {
    styles: Record<string, any>;    // CSS 변수, 레이아웃 정의
  };
}
```

### 1.5 사용 예제

**Example 1: 파일 전체 구조 조회**
```json
{
  "name": "get_figma_data",
  "arguments": {
    "fileKey": "abc123XYZ"
  }
}
```

**Example 2: 특정 노드 상세 조회**
```json
{
  "name": "get_figma_data",
  "arguments": {
    "fileKey": "abc123XYZ",
    "nodeId": "1234:5678",
    "depth": 3
  }
}
```

**Example 3: 얕은 탐색 (성능 최적화)**
```json
{
  "name": "get_figma_data",
  "arguments": {
    "fileKey": "abc123XYZ",
    "nodeId": "1234:5678",
    "depth": 1
  }
}
```

### 1.6 에러 조건

| 에러 코드 | 메시지 | 원인 | 복구 방법 |
|----------|--------|------|---------|
| **401** | Unauthorized | 잘못된 또는 만료된 API 키 | Figma 설정에서 새 토큰 생성 |
| **404** | Not Found | 존재하지 않는 파일 또는 노드 | 파일 키/노드 ID 확인 |
| **429** | Rate Limit | API 호출 분당 60회 초과 | 1초 이상 대기 후 재시도 |
| **500** | Server Error | Figma API 서버 오류 | 5초 후 재시도 (지수 백오프) |

---

## 2. download_figma_images - 이미지 다운로드

**도구 출처**: Figma Context MCP
**기능**: Figma 노드에서 이미지, 아이콘, 벡터 자산 다운로드

### 2.1 전체 파라미터 명세

| 파라미터 | 타입 | 필수 | 설명 | 제약 |
|---------|------|------|------|-----|
| `fileKey` | string | ✅ | Figma 파일 키 | 22자 알파벳+숫자 |
| `localPath` | string | ✅ | 로컬 저장 경로 | 절대 경로 필수, 디렉토리 미존재 시 자동 생성 |
| `pngScale` | number | ❌ | PNG 스케일 | 1, 2, 3, 4만 허용 (기본: 1) |
| `format` | string | ❌ | 이미지 형식 | "png" (기본), "svg", "jpg", "pdf" |
| `nodes` | array | ✅ | 다운로드할 노드 배열 | 최소 1개, 권장 최대 100개 |

### 2.2 nodes 배열 항목 상세

```typescript
{
  nodeId: string;              // ✅ 필수: 노드 ID ("1234:5678")
  fileName: string;            // ✅ 필수: 저장할 파일명 (확장자 포함)
  imageRef?: string;           // ❌ 선택: 이미지 참조 ID
  needsCropping?: boolean;     // ❌ 선택: 자동 크롭 여부
  cropTransform?: array;       // ❌ 선택: 크롭 변환 매트릭스 [[1,0,0], [0,1,0]]
  requiresImageDimensions?: boolean; // ❌ 선택: CSS 변수용 크기 추출
}
```

### 2.3 localPath 검증 규칙

| 조건 | 허용 | 예시 | 에러 메시지 |
|------|------|------|-----------|
| 절대 경로 | ✅ | `/Users/dev/assets/images` | - |
| 상대 경로 | ❌ | `./assets/images` | "Path for asset writes is invalid" |
| 디렉토리 미존재 | ✅ 자동생성 | `/Users/dev/assets/new-dir` | - |
| 권한 부족 | ❌ | `/root/protected` (EACCES) | "Path for asset writes is invalid" |
| 절대 경로 검증 코드 | | | |

```typescript
// ❌ 잘못된 예
const localPath = "./assets";

// ✅ 올바른 예
import path from 'path';
const localPath = path.resolve(process.cwd(), './assets');
// 또는
const localPath = '/Users/dev/project/assets';
```

### 2.4 pngScale 가이드

| Scale | 해상도 | 파일 크기 | 사용 사례 |
|-------|--------|---------|---------|
| **1** | 1x | 최소 | 모바일, 웹 기본 |
| **2** | 2x (@2x) | 4배 | 고해상도 디스플레이 |
| **3** | 3x | 9배 | 매우 고해상도 (거의 안 씀) |
| **4** | 4x | 16배 | 최고 품질 (성능 주의) |

**주의**: Scale 증가 → 이미지 크기 지수 증가 → Rate Limit 빠르게 도달

### 2.5 반환값 구조

```typescript
{
  content: [{
    type: "text",
    text: "Downloaded 2 images:\n- hero-bg.png: 750x1624 | --hero-bg-width: 375px; --hero-bg-height: 812px (cropped)\n- logo.svg: 120x40"
  }]
}
```

### 2.6 사용 예제

**Example 1: 단일 이미지 다운로드**
```json
{
  "name": "download_figma_images",
  "arguments": {
    "fileKey": "abc123XYZ",
    "localPath": "/Users/dev/project/assets/images",
    "nodes": [
      {
        "nodeId": "1234:5678",
        "fileName": "hero-bg.png"
      }
    ]
  }
}
```

**Example 2: 다중 이미지 + 크롭 + CSS 변수 생성**
```json
{
  "name": "download_figma_images",
  "arguments": {
    "fileKey": "abc123XYZ",
    "localPath": "/Users/dev/project/assets",
    "pngScale": 2,
    "nodes": [
      {
        "nodeId": "1234:5680",
        "fileName": "hero-bg.png",
        "needsCropping": true,
        "requiresImageDimensions": true
      },
      {
        "nodeId": "1234:5681",
        "fileName": "logo.svg"
      }
    ]
  }
}
```

**Example 3: SVG 형식으로 다운로드 (base64 에러 회피)**
```json
{
  "name": "download_figma_images",
  "arguments": {
    "fileKey": "abc123XYZ",
    "localPath": "/Users/dev/project/assets",
    "format": "svg",
    "nodes": [
      {
        "nodeId": "1234:5678",
        "fileName": "icon.svg"
      }
    ]
  }
}
```

### 2.7 에러 조건

| 에러 메시지 | HTTP Code | 원인 | 해결책 |
|-----------|----------|------|--------|
| **Path for asset writes is invalid** | 400 | 상대 경로, 권한 부족, 경로 형식 오류 | 절대 경로 사용, 디렉토리 생성 권한 확인 |
| **Image base64 format error** | 400 | 이미지 인코딩 실패, pngScale 너무 큼 | pngScale=1 또는 2로 감소, SVG 형식 시도 |
| **Node not found** | 404 | 존재하지 않는 노드 ID | `get_figma_data`로 유효한 노드 ID 재확인 |
| **Unauthorized** | 401 | 파일 접근 권한 없음 | 파일 소유자에게 읽기 권한 요청 |
| **Rate limit exceeded** | 429 | API 호출 분당 30회 초과 | 2초 이상 대기 후 재시도 |

---

## 3. export_node_as_image - 노드 이미지 내보내기

**도구 출처**: Talk To Figma MCP (`/sethdford/mcp-figma`)
**기능**: 노드를 이미지 형식으로 내보내기

### 3.1 전체 파라미터 명세

| 파라미터 | 타입 | 필수 | 설명 | 제약 |
|---------|------|------|------|-----|
| `node_id` | string | ✅ | 내보낼 노드 ID | "1234:5678" 형식 |
| `format` | string | ✅ | 이미지 형식 | "PNG", "JPG", "SVG", "PDF" |

### 3.2 format별 상세 정보

| 형식 | 용도 | 품질 | 파일 크기 | 주의사항 |
|------|------|------|---------|--------|
| **PNG** | 웹, UI 요소 | 손실 없음 | 중간 | 투명도 지원 |
| **JPG** | 사진, 복잡한 이미지 | 압축 손실 | 작음 | 투명도 미지원 |
| **SVG** | 벡터, 아이콘 | 손실 없음 | 작음 | 벡터만 가능 |
| **PDF** | 인쇄, 아카이브 | 손실 없음 | 큼 | 페이지 레이아웃 유지 |

### 3.3 반환값

**현재 제한**: base64 텍스트로만 반환 (직접 파일 저장 불가)

```json
{
  "result": {
    "base64": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
  }
}
```

### 3.4 사용 예제

```json
{
  "name": "export_node_as_image",
  "arguments": {
    "node_id": "1234:5678",
    "format": "PNG"
  }
}
```

### 3.5 에러 조건

| 에러 | 원인 | 해결책 |
|------|------|--------|
| **base64 encode 실패** | 노드 타입이 내보내기 미지원 | 노드 타입 확인 (FRAME, COMPONENT, RECTANGLE 등만 지원) |
| **Format not supported** | 형식 오류 | "PNG", "JPG", "SVG", "PDF" 중 선택 |

---

## 4. 변수 조회 API (REST)

**API 출처**: Figma 공식 REST API
**엔드포인트**: `GET /v1/files/{fileKey}/variables`
**인증**: `X-Figma-Token` 헤더

### 4.1 전체 파라미터 명세

| 파라미터 | 위치 | 타입 | 필수 | 설명 |
|---------|------|------|------|------|
| `file_key` | Path | string | ✅ | Figma 파일 키 |
| `published` | Query | boolean | ❌ | 게시된 변수만 조회 (기본: false) |

### 4.2 요청 헤더

```http
GET /v1/files/abc123XYZ/variables
X-Figma-Token: figd_your_token_here
Content-Type: application/json
```

### 4.3 응답 구조 (200 OK)

```typescript
{
  variables: [{
    id: string;                              // 변수 고유 ID
    name: string;                            // 변수 이름
    key: string;                             // 임포트 키
    variableCollectionId: string;            // 소속 컬렉션 ID
    resolvedType: "COLOR" | "FLOAT" | "STRING" | "BOOLEAN" | "VARIABLE_ALIAS";
    valuesByMode: {
      [modeId: string]: any;                 // 모드별 값
    };
    remote: boolean;                         // 원격(게시됨) 여부
    description: string;                     // 변수 설명
    hiddenFromPublishing: boolean;           // 라이브러리 게시 시 숨김
    scopes: string[];                        // UI 피커 범위
    codeSyntax: {
      WEB?: string;                          // CSS 변수명
      ANDROID?: string;                      // Android 상수명
      iOS?: string;                          // iOS 상수명
    };
  }];

  collections: [{
    id: string;
    name: string;
    modes: [{
      modeId: string;
      name: string;
    }];
    defaultModeId: string;
    variableIds: string[];
  }];
}
```

### 4.4 resolvedType별 값 형식

| Type | JavaScript 타입 | 예제 |
|------|---------------|------|
| **COLOR** | object | `{ "r": 1, "g": 0, "b": 0, "a": 1 }` |
| **FLOAT** | number | `16`, `2.5` |
| **STRING** | string | `"Roboto"`, `"bold"` |
| **BOOLEAN** | boolean | `true`, `false` |
| **VARIABLE_ALIAS** | object | `{ "type": "VARIABLE_ALIAS", "id": "123:789" }` |

### 4.5 사용 예제

**Example 1: 모든 변수 조회 (로컬 포함)**
```http
GET /v1/files/abc123XYZ/variables
X-Figma-Token: figd_your_token_here
```

**Example 2: 게시된 변수만 조회**
```http
GET /v1/files/abc123XYZ/variables?published=true
X-Figma-Token: figd_your_token_here
```

**Example 3: TypeScript 구현**
```typescript
async function getVariables(fileKey: string, apiToken: string) {
  const response = await fetch(
    `https://api.figma.com/v1/files/${fileKey}/variables`,
    {
      headers: { 'X-Figma-Token': apiToken }
    }
  );

  if (!response.ok) {
    throw new Error(`API Error: ${response.status}`);
  }

  return await response.json();
}
```

### 4.6 변수가 없을 때 응답

```json
{
  "variables": [],
  "collections": []
}
```

### 4.7 에러 조건

| HTTP Code | 에러 | 원인 | 해결책 |
|----------|------|------|--------|
| **400** | Bad Request | 파일 키 형식 오류 | 22자 알파벳+숫자 확인 |
| **401** | Unauthorized | 잘못되거나 만료된 토큰 | Figma 설정에서 새 토큰 생성 |
| **403** | Forbidden | 파일 접근 권한 없음 | 파일 소유자에게 권한 요청 |
| **404** | Not Found | 존재하지 않는 파일 | 파일 키 및 URL 재확인 |
| **429** | Rate Limit | 분당 100회 초과 | Retry-After 헤더 확인 후 재시도 |

---

## 5. 파라미터 제약사항 및 한계

### 5.1 API 제한사항

| 제한 | 값 | 영향 범위 | 우회 방법 |
|-----|------|---------|----------|
| **분당 API 호출 (일반)** | 60회 | get_figma_data, download_figma_images | 청킹: 10개씩 묶고 1초 대기 |
| **분당 API 호출 (이미지)** | 30회 | export_node_as_image | 300ms 이상 간격 유지 |
| **분당 API 호출 (변수)** | 100회 | 변수 API | 가장 여유로움 |
| **최대 노드 배열** | 100개 (권장) | download_figma_images | 배열 분할 후 순차 호출 |
| **최대 depth** | 10 | get_figma_data | depth 증가 시 지수 성능 저하 |
| **파일 크기 (추출)** | 50MB | 대규모 설계 시스템 | 페이지별로 분할 추출 |

### 5.2 nodeId 형식 제약

**허용되는 형식**:
```
"1234:5678"       ✅ 공식 형식
"1234-5678"       ✅ 대체 형식 (일부 도구)
```

**허용되지 않는 형식**:
```
"1234_5678"       ❌ 언더스코어
"1234.5678"       ❌ 마침표
"node-1234-5678"  ❌ 프리픽스
```

### 5.3 로컬 경로 제약

**Windows vs Unix 경로**:

```typescript
// ❌ 플랫폼별로 다른 경로 (이식성 없음)
const windowsPath = "C:\\Users\\dev\\assets";
const unixPath = "/Users/dev/assets";

// ✅ 플랫폼 독립적 (권장)
import path from 'path';
const portablePath = path.resolve(__dirname, './assets');
```

### 5.4 clientLanguages/clientFrameworks 미지원

**요청된 기능**:
- `clientLanguages`: TypeScript, JavaScript, etc. 지원 예상
- `clientFrameworks`: React, Vue, Angular 지원 예상

**현황**: 현재 공개 문서에서 지원 확인 안 됨

**대안**: 추출된 디자인 데이터를 코드 생성기에 직접 입력

```typescript
// 예: 디자인 → React 코드 수동 생성
const design = await figmaService.getRawNode(fileKey, nodeId);
const reactCode = generateReact(design);  // 별도 함수
await fs.writeFile('./Component.tsx', reactCode);
```

### 5.5 dirForAssetWrites 파라미터

**상태**: `download_figma_images`의 `localPath` 파라미터로 구현됨

**요구사항**:
- 필수: ✅
- 기본값: 없음 (명시 필수)
- 검증: 절대 경로만 허용

```typescript
// ✅ 올바른 사용
await figmaService.downloadImages(
  fileKey,
  "/Users/dev/assets",  // 명시 필수
  nodes
);

// ❌ 오류 유발
await figmaService.downloadImages(
  fileKey,
  "./assets",  // 상대 경로 불가
  nodes
);
```

### 5.6 forceCode 파라미터

**상태**: 현재 공개 MCP 도구에서 미지원

**추정 용도**: 이미지 다운로드 스킵하고 코드 생성만 진행

**우회 방법**:
```typescript
// 조건부 다운로드 (코드 생성 전 이미지 필요 여부 판단)
const skipImages = forceCode === true;

if (!skipImages) {
  await figmaService.downloadImages(fileKey, localPath, nodes);
}

const code = generateCode(design);
```

---

## 6. 에러 코드 매핑

### 6.1 HTTP 상태 코드별 에러

| HTTP Code | 상태 | 에러 타입 | 대상 도구 | 원인 | 권장 행동 |
|----------|------|---------|---------|------|---------|
| **200** | ✅ | - | 모두 | 성공 | 계속 진행 |
| **400** | ❌ | `invalid_request_error` | 모두 | 파라미터 오류 (파일키 형식, 경로) | 파라미터 검증, 재요청 |
| **401** | ❌ | `authentication_error` | 모두 | API 토큰 만료/오류 | 새 토큰 발급, 재인증 |
| **403** | ❌ | `forbidden_error` | 모두 | 파일 접근 권한 없음 | 파일 소유자에게 권한 요청 |
| **404** | ❌ | `not_found_error` | 모두 | 파일/노드 미존재 | 파일키/노드ID 재확인 |
| **429** | ⚠️ | `rate_limit_error` | 모두 | API 호출 제한 초과 | **Exponential Backoff** 재시도 |
| **500** | ❌ | `server_error` | 모두 | Figma 서버 오류 | 지수 백오프로 재시도 (최대 3회) |
| **503** | ❌ | `service_unavailable` | 모두 | Figma 서비스 정지 | 몇 분 후 재시도 |

### 6.2 에러별 Retry 전략

**429 Rate Limit - Exponential Backoff 구현**:

```typescript
async function retryWithBackoff(
  fn: () => Promise<any>,
  maxRetries: number = 3,
  initialDelayMs: number = 1000
): Promise<any> {
  let lastError: Error;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error: any) {
      lastError = error;

      // 429 Rate Limit 감지
      if (error.response?.status === 429) {
        // Retry-After 헤더 확인 (서버에서 권장 대기 시간)
        const retryAfter = error.response.headers['retry-after'];
        const delayMs = retryAfter
          ? parseInt(retryAfter) * 1000
          : initialDelayMs * Math.pow(2, attempt);  // 1s → 2s → 4s

        console.log(`Rate limited. Waiting ${delayMs}ms before retry...`);
        await new Promise(resolve => setTimeout(resolve, delayMs));
        continue;
      }

      // 5xx 서버 에러 (최대 3회 재시도)
      if (error.response?.status >= 500) {
        const delayMs = initialDelayMs * Math.pow(2, attempt);
        console.log(`Server error. Waiting ${delayMs}ms before retry...`);
        await new Promise(resolve => setTimeout(resolve, delayMs));
        continue;
      }

      // 4xx 클라이언트 에러 (재시도 불가)
      throw error;
    }
  }

  throw lastError;
}

// 사용
const data = await retryWithBackoff(
  () => figmaService.getRawFile("abc123XYZ"),
  3,     // 최대 3회 재시도
  1000   // 초기 1초 대기
);
```

### 6.3 에러 처리 플로우차트

```
요청 → API 호출
  ├─ 200 OK → 성공 반환
  ├─ 400 Bad Request
  │  └─ 원인: 파라미터 오류 (절대경로 아님, 파일키 형식)
  │     └─ 처리: 파라미터 검증 후 즉시 재요청 (재시도 없음)
  ├─ 401 Unauthorized
  │  └─ 원인: 토큰 만료/오류
  │     └─ 처리: 새 토큰 발급 후 재요청
  ├─ 403 Forbidden
  │  └─ 원인: 권한 없음
  │     └─ 처리: 사용자 알림, 권한 요청
  ├─ 404 Not Found
  │  └─ 원인: 파일/노드 미존재
  │     └─ 처리: 데이터 재확인 후 오류 보고
  ├─ 429 Too Many Requests
  │  └─ 원인: Rate Limit 초과
  │     └─ 처리: Exponential Backoff (1s → 2s → 4s) 최대 3회
  └─ 5xx Server Error
     └─ 원인: Figma 서버 오류
        └─ 처리: Exponential Backoff 재시도 (최대 3회)
```

### 6.4 MCP 도구별 에러 핸들링

**get_figma_data 에러**:

```typescript
try {
  const data = await figmaService.getRawFile(fileKey, depth);
} catch (error: any) {
  if (error.response?.status === 401) {
    console.error('API Key 만료됨. 새 키 발급 필요');
  } else if (error.response?.status === 404) {
    console.error(`파일 '${fileKey}' 없음. 파일 키 재확인`);
  } else if (error.response?.status === 429) {
    console.warn('Rate Limit 도달. 1초 이상 대기 후 재시도');
  } else {
    console.error('알 수 없는 에러:', error.message);
  }
}
```

**download_figma_images 에러**:

```typescript
try {
  await figmaService.downloadImages(fileKey, localPath, nodes);
} catch (error: any) {
  if (error.message.includes('Path for asset writes is invalid')) {
    console.error('경로 오류: 절대 경로 사용 필요 (상대경로 불가)');
  } else if (error.message.includes('Image base64 format error')) {
    console.error('이미지 인코딩 실패: pngScale 감소 또는 SVG 형식 시도');
  } else if (error.response?.status === 404) {
    console.error('노드를 찾을 수 없음. get_figma_data로 노드 ID 재확인');
  }
}
```

---

## 요약: 파라미터 체크리스트

### get_figma_data 호출 전
- [ ] `fileKey` 형식 확인 (22자 알파벳+숫자)
- [ ] `nodeId` 필요 여부 (전체 또는 특정 노드)
- [ ] `depth` 설정 (성능 vs 상세도)
- [ ] API 토큰 유효성 확인

### download_figma_images 호출 전
- [ ] `fileKey` 형식 확인
- [ ] `localPath` 절대 경로 확인 (path.resolve 사용)
- [ ] 디렉토리 생성 권한 확인
- [ ] `pngScale` 값 적절성 (1-4)
- [ ] `nodes` 배열 최대 100개 준수
- [ ] `fileName`에 확장자 포함

### 변수 조회 API 호출 전
- [ ] `X-Figma-Token` 헤더 설정
- [ ] `fileKey` 형식 확인
- [ ] `published` 파라미터 (로컬 vs 게시 변수)

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (Section 1-6)
**추가 검증 필요**: `clientLanguages`, `clientFrameworks`, `forceCode` 파라미터 (아직 공식 문서에서 미발견)
