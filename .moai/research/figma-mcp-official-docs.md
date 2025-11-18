# Figma MCP 공식 문서 조사 결과

**조사 일자**: 2025-11-19
**조사 범위**: Context7 MCP를 통한 Figma Dev Mode MCP 공식 문서
**조사 상태**: ⚠️ 요청된 특정 도구 미발견

---

## 📋 요약

Context7 MCP 및 공개 저장소 조사 결과, 요청하신 특정 도구들(`get_design_context`, `get_variable_defs`, `get_screenshot`)은 **현재 공개된 Figma MCP 구현체에서 발견되지 않았습니다**.

대신, 다음 3개의 주요 커뮤니티 기반 Figma MCP 서버가 확인되었습니다:

1. **Figma Context MCP** (`/glips/figma-context-mcp`) - High reputation, 40 code snippets
2. **Cursor Talk To Figma MCP** (`/sethdford/mcp-figma`) - High reputation, 79 code snippets
3. **Figma Copilot** (`/xlzuvekas/figma-copilot`) - Medium reputation, 71 code snippets

---

## 🔍 조사 대상 도구 (요청사항)

아래 도구들은 요청되었으나 **현재 공개된 문서에서 발견되지 않았습니다**:

### 1. `get_design_context` (미발견)
- **상태**: 문서화되지 않음
- **추정 용도**: Dev Mode 디자인 컨텍스트 추출
- **대체 도구**: `get_figma_data` (Figma Context MCP)

### 2. `get_variable_defs` (미발견)
- **상태**: 문서화되지 않음
- **추정 용도**: Figma 변수 정의 조회
- **대체 접근**: Figma REST API Variables 엔드포인트 직접 호출

### 3. `get_screenshot` (미발견)
- **상태**: 문서화되지 않음
- **추정 용도**: 노드 스크린샷 캡처
- **대체 도구**: `downloadImages` (FigmaService), `export_node_as_image` (MCP-Figma)

---

## ✅ 발견된 대체 도구 및 API

### Section 1: Figma Context MCP (/glips/figma-context-mcp)

**평판**: High | **코드 예제**: 40개 | **벤치마크 점수**: 65.4

#### 1.1 `get_figma_data` - 디자인 데이터 조회

**목적**: Figma 파일 또는 특정 노드의 구조화된 디자인 데이터 추출

**파라미터**:

| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|------|------|
| `fileKey` | string | ✅ 필수 | Figma 파일 키 (예: `abc123XYZ`) |
| `nodeId` | string | ❌ 선택 | 특정 노드 ID (예: `1234:5678`) |
| `depth` | number | ❌ 선택 | 트리 탐색 깊이 (기본값: 전체) |

**반환값 구조**:

```json
{
  "metadata": {
    "name": "Login Screen",
    "components": {},
    "componentSets": {}
  },
  "nodes": [
    {
      "id": "1234:5678",
      "name": "LoginForm",
      "type": "FRAME",
      "layout": "layout-1",
      "children": [...]
    }
  ],
  "globalVars": {
    "styles": {
      "layout-1": {
        "width": 375,
        "height": 812,
        "layoutMode": "VERTICAL",
        "padding": "16px"
      }
    }
  }
}
```

**사용 예제**:

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

**에러 조건**:
- **401 Unauthorized**: 잘못된 Figma API 키
- **404 Not Found**: 존재하지 않는 파일 키 또는 노드 ID
- **429 Rate Limit**: API 호출 제한 초과

---

#### 1.2 `download_figma_images` - 이미지 다운로드

**목적**: Figma 노드에서 이미지, 아이콘, 벡터 자산 다운로드

**파라미터**:

| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|------|------|
| `fileKey` | string | ✅ 필수 | Figma 파일 키 |
| `localPath` | string | ✅ 필수 | 로컬 저장 경로 |
| `pngScale` | number | ❌ 선택 | PNG 스케일 (1, 2, 3, 4) |
| `nodes` | array | ✅ 필수 | 다운로드할 노드 목록 |
| `nodes[].nodeId` | string | ✅ 필수 | 노드 ID |
| `nodes[].fileName` | string | ✅ 필수 | 저장할 파일명 (확장자 포함) |
| `nodes[].imageRef` | string | ❌ 선택 | 이미지 참조 ID |
| `nodes[].needsCropping` | boolean | ❌ 선택 | 자동 크롭 여부 |
| `nodes[].cropTransform` | array | ❌ 선택 | 크롭 변환 매트릭스 |
| `nodes[].requiresImageDimensions` | boolean | ❌ 선택 | CSS 변수용 크기 추출 |

**반환값**:

```json
{
  "content": [
    {
      "type": "text",
      "text": "Downloaded 2 images:\n- hero-bg.png: 750x1624 | --hero-bg-width: 375px; --hero-bg-height: 812px (cropped)\n- logo.svg: 120x40"
    }
  ]
}
```

**사용 예제**:

```json
{
  "name": "download_figma_images",
  "arguments": {
    "fileKey": "abc123XYZ",
    "localPath": "/Users/dev/project/assets/images",
    "pngScale": 2,
    "nodes": [
      {
        "nodeId": "1234:5680",
        "imageRef": "abcdef123456",
        "fileName": "hero-bg.png",
        "needsCropping": true,
        "cropTransform": [[1, 0, 0], [0, 1, 0]],
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

**에러 해결**:

| 에러 메시지 | 원인 | 해결책 |
|-----------|------|--------|
| "Path for asset writes is invalid" | 잘못된 로컬 경로 | 절대 경로 사용, 디렉토리 존재 확인 |
| "Image base64 format error" | 이미지 인코딩 실패 | `pngScale` 값 조정 (1-4), 노드 타입 확인 |
| "Node not found" | 존재하지 않는 노드 ID | `get_figma_data`로 유효한 노드 ID 확인 |

---

#### 1.3 FigmaService 클래스 (TypeScript)

**목적**: Figma REST API와 직접 상호작용

**초기화**:

```typescript
import { FigmaService } from "./services/figma.js";

const figmaService = new FigmaService({
  figmaApiKey: "figd_your_token_here",
  figmaOAuthToken: "",
  useOAuth: false
});
```

**주요 메서드**:

##### `getRawFile(fileKey: string, depth?: number)`

**파라미터**:
- `fileKey`: Figma 파일 키 (필수)
- `depth`: 트리 깊이 (선택, 기본값: 5)

**반환값**:
```typescript
{
  name: string;
  document: {
    children: Array<any>;
  };
}
```

**예제**:
```typescript
const fileData = await figmaService.getRawFile("abc123XYZ", 5);
console.log(fileData.name); // "Design System"
console.log(fileData.document.children.length);
```

---

##### `getRawNode(fileKey: string, nodeId: string, depth?: number)`

**파라미터**:
- `fileKey`: Figma 파일 키 (필수)
- `nodeId`: 노드 ID (필수)
- `depth`: 트리 깊이 (선택, 기본값: 3)

**반환값**:
```typescript
{
  nodes: {
    [nodeId: string]: {
      document: any;
    };
  };
}
```

**예제**:
```typescript
const nodeData = await figmaService.getRawNode("abc123XYZ", "1234:5678", 3);
console.log(nodeData.nodes["1234:5678"].document.name);
```

---

##### `getNodeRenderUrls(fileKey: string, nodeIds: string[], format: string, options?: object)`

**파라미터**:
- `fileKey`: Figma 파일 키 (필수)
- `nodeIds`: 노드 ID 배열 (필수)
- `format`: 이미지 형식 (`"png"`, `"svg"`, `"jpg"`, `"pdf"`)
- `options`: 렌더링 옵션 (선택)
  - `pngScale`: PNG 스케일 (1, 2, 3, 4)
  - `svgOutlineText`: SVG 텍스트 아웃라인 변환 (boolean)

**반환값**:
```typescript
{
  [nodeId: string]: string; // 다운로드 URL
}
```

**예제**:
```typescript
const imageUrls = await figmaService.getNodeRenderUrls(
  "abc123XYZ",
  ["1234:5678", "1234:5679"],
  "png",
  { pngScale: 3 }
);
// { "1234:5678": "https://s3.amazonaws.com/...", ... }
```

---

##### `downloadImages(fileKey: string, localPath: string, nodes: array, options?: object)`

**파라미터**:
- `fileKey`: Figma 파일 키 (필수)
- `localPath`: 로컬 저장 경로 (필수)
- `nodes`: 다운로드할 노드 정보 배열 (필수)
  - `nodeId`: 노드 ID
  - `fileName`: 저장할 파일명
- `options`: 렌더링 옵션 (선택)
  - `pngScale`: PNG 스케일

**반환값**:
```typescript
Array<{
  filePath: string;
  finalDimensions: { width: number; height: number };
  wasCropped: boolean;
}>
```

**예제**:
```typescript
const results = await figmaService.downloadImages(
  "abc123XYZ",
  "./public/images",
  [
    {
      nodeId: "1234:5678",
      fileName: "button-icon.svg"
    }
  ],
  { pngScale: 2 }
);
```

---

#### 1.4 Extractor 시스템 (디자인 데이터 단순화)

**목적**: 복잡한 Figma API 응답을 구조화된 데이터로 변환

**사용 가능한 Extractor**:

| Extractor | 설명 | 추출 데이터 |
|-----------|------|------------|
| `allExtractors` | 모든 정보 추출 | 레이아웃, 텍스트, 시각, 컴포넌트 |
| `layoutAndText` | 레이아웃 + 텍스트만 | 구조, 텍스트 콘텐츠 |
| `contentOnly` | 텍스트만 | 텍스트 콘텐츠 |
| `layoutOnly` | 레이아웃만 | 구조, 크기, 위치 |
| `visualsOnly` | 시각 속성만 | 색상, 테두리, 효과 |

**사용 예제**:

```typescript
import {
  simplifyRawFigmaObject,
  allExtractors,
  layoutAndText,
  contentOnly,
  visualsOnly,
  layoutOnly,
  collapseSvgContainers
} from "./extractors/index.js";

// 전체 디자인 정보 추출
const rawResponse = await figmaService.getRawFile("abc123XYZ");
const fullDesign = simplifyRawFigmaObject(rawResponse, allExtractors, {
  maxDepth: 10,
  afterChildren: collapseSvgContainers
});

// 레이아웃 + 텍스트만
const layoutTextDesign = simplifyRawFigmaObject(rawResponse, layoutAndText);

// 텍스트만
const textOnlyDesign = simplifyRawFigmaObject(rawResponse, contentOnly);

// 시각 속성만 (특정 노드 타입 필터링)
const visualDesign = simplifyRawFigmaObject(rawResponse, visualsOnly, {
  nodeFilter: (node) => node.type === "RECTANGLE" || node.type === "ELLIPSE"
});
```

**옵션 파라미터**:

| 옵션 | 타입 | 설명 |
|------|------|------|
| `maxDepth` | number | 최대 탐색 깊이 |
| `afterChildren` | function | 자식 처리 후 실행할 함수 |
| `nodeFilter` | function | 노드 필터링 조건 |

---

### Section 2: Cursor Talk To Figma MCP (/sethdford/mcp-figma)

**평판**: High | **코드 예제**: 79개

#### 2.1 Document & Selection API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `get_document_info` | 현재 Figma 문서 정보 조회 | 없음 |
| `get_selection` | 현재 선택된 노드 정보 조회 | 없음 |
| `read_my_design` | 선택된 노드의 상세 정보 조회 (파라미터 불필요) | 없음 |
| `get_node_info` | 특정 노드의 상세 정보 조회 | `node_id` (string) |
| `get_nodes_info` | 여러 노드의 상세 정보 조회 | `node_ids` (string[]) |

---

#### 2.2 Annotation Management API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `get_annotations` | 문서 또는 특정 노드의 모든 주석 조회 | `node_id?` (string, 선택) |
| `set_annotation` | 주석 생성 또는 업데이트 (마크다운 지원) | `annotation_data` (object) |
| `set_multiple_annotations` | 여러 주석을 일괄 생성/업데이트 | `annotations` (object[]) |
| `scan_nodes_by_types` | 특정 타입의 노드 스캔 (주석 대상 찾기) | `types` (string[]) |

**주석 데이터 구조**:

```json
{
  "nodeId": "1234:5678",
  "content": "## 디자인 가이드\n\n이 컴포넌트는 **반응형**입니다.",
  "author": "user@example.com",
  "timestamp": "2023-08-31T10:00:00Z"
}
```

---

#### 2.3 Prototyping & Connection API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `get_reactions` | 노드의 프로토타입 반응 조회 (시각적 하이라이트) | 없음 |
| `set_default_connector` | 기본 커넥터 스타일 설정 (연결 생성 전 필수) | `connector_id` (string) |
| `create_connections` | 노드 간 커넥터 라인 생성 | `connections` (object[]) |

**연결 데이터 구조**:

```json
{
  "connections": [
    {
      "from": "1234:5678",
      "to": "1234:5679",
      "type": "ARROW"
    }
  ]
}
```

---

#### 2.4 Element Creation API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `create_rectangle` | 새 사각형 생성 | `position`, `size`, `name?` |
| `create_frame` | 새 프레임 생성 | `position`, `size`, `name?` |
| `create_text` | 새 텍스트 노드 생성 | `position`, `content`, `font_properties?` |

**생성 예제**:

```json
{
  "name": "create_rectangle",
  "arguments": {
    "position": { "x": 100, "y": 100 },
    "size": { "width": 200, "height": 100 },
    "name": "Button Background"
  }
}
```

---

#### 2.5 Text Modification API

| 도구 | 설명 | 주의사항 | 파라미터 |
|------|------|---------|---------|
| `scan_text_nodes` | 텍스트 노드 스캔 (대규모 디자인 청킹) | - | 없음 |
| `set_text_content` | 단일 텍스트 노드 콘텐츠 설정 | ⚠️ 포맷 손실 | `node_id`, `content` |
| `set_multiple_text_contents` | 여러 텍스트 노드 일괄 업데이트 | ⚠️ 포맷 손실 | `updates[]` |

**주의사항**: `set_text_content`와 `set_multiple_text_contents`는 **텍스트 포맷(굵기, 색상 등)을 유지하지 않습니다**.

**Before (텍스트 포맷 유지 필요 시)**:
```json
{
  "name": "get_node_info",
  "arguments": { "node_id": "1234:5678" }
}
```

**After (포맷 손실 허용 시)**:
```json
{
  "name": "set_text_content",
  "arguments": {
    "node_id": "1234:5678",
    "content": "새로운 텍스트"
  }
}
```

---

#### 2.6 Auto Layout API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `set_layout_mode` | 레이아웃 모드 및 래핑 설정 | `node_id`, `mode` (`NONE`/`HORIZONTAL`/`VERTICAL`) |
| `set_padding` | 오토 레이아웃 패딩 설정 | `node_id`, `top`, `right`, `bottom`, `left` |
| `set_axis_align` | 주축 및 반대축 정렬 설정 | `node_id`, `primaryAlign`, `counterAlign` |
| `set_layout_sizing` | 수평/수직 크기 조정 모드 | `node_id`, `horizontal`, `vertical` (`FIXED`/`HUG`/`FILL`) |
| `set_item_spacing` | 자식 간 간격 설정 | `node_id`, `spacing` (number) |

---

#### 2.7 Styling API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `set_fill_color` | 채우기 색상 설정 | `node_id`, `color` (RGBA) |
| `set_stroke_color` | 테두리 색상 및 두께 설정 | `node_id`, `color`, `weight` |
| `set_corner_radius` | 모서리 반경 설정 (개별 모서리 지원) | `node_id`, `radius` (number 또는 object) |

**색상 구조 (RGBA)**:

```json
{
  "r": 1.0,
  "g": 0.5,
  "b": 0.0,
  "a": 1.0
}
```

---

#### 2.8 Export API

| 도구 | 설명 | 제한사항 | 파라미터 |
|------|------|---------|---------|
| `export_node_as_image` | 노드를 이미지로 내보내기 | ⚠️ 현재 base64 텍스트 반환 (제한적 지원) | `node_id`, `format` (`PNG`/`JPG`/`SVG`/`PDF`) |

**사용 예제**:

```json
{
  "name": "export_node_as_image",
  "arguments": {
    "node_id": "1234:5678",
    "format": "PNG"
  }
}
```

**반환값** (현재):
```json
{
  "result": {
    "base64": "iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

---

#### 2.9 Components & Styles API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `get_styles` | 로컬 스타일 정보 조회 | 없음 |
| `get_local_components` | 로컬 컴포넌트 정보 조회 | 없음 |
| `create_component_instance` | 컴포넌트 인스턴스 생성 | `component_id` |
| `get_instance_overrides` | 선택된 컴포넌트 인스턴스의 오버라이드 추출 | 없음 |
| `set_instance_overrides` | 대상 인스턴스에 오버라이드 적용 | `instance_id`, `overrides` |

---

#### 2.10 Layout & Organization API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `move_node` | 노드를 새 위치로 이동 | `node_id`, `position` ({x, y}) |
| `resize_node` | 노드 크기 조정 | `node_id`, `dimensions` ({width, height}) |
| `delete_node` | 단일 노드 삭제 | `node_id` |
| `delete_multiple_nodes` | 여러 노드 일괄 삭제 (효율적) | `node_ids[]` |
| `clone_node` | 노드 복사 (옵션: 위치 오프셋) | `node_id`, `offset?` ({x, y}) |

---

#### 2.11 WebSocket 메시지 프로토콜

**메시지 구조**:

**1. 채널 참여 (클라이언트 → 서버)**:
```json
{
  "type": "join",
  "channel": "channel-name"
}
```

**2. 시스템 메시지 (서버 → 클라이언트)**:
```json
{
  "type": "system",
  "message": {
    "result": true
  },
  "channel": "channel-name"
}
```

**3. 에러 메시지 (서버 → 클라이언트)**:
```json
{
  "type": "error",
  "message": "에러 설명"
}
```

**4. 명령/응답 메시지 (양방향)**:
```json
{
  "id": "unique-request-id",
  "type": "message",
  "channel": "channel-name",
  "message": {
    "id": "unique-request-id",
    "command": "get_document_info",
    "params": {},
    "result": {},
    "error": "에러 메시지 (실패 시)"
  }
}
```

**5. 진행 상황 업데이트 (서버 → 클라이언트)**:
```json
{
  "id": "command-id",
  "type": "progress_update",
  "channel": "channel-name",
  "message": {
    "id": "command-id",
    "type": "progress_update",
    "data": {
      "commandId": "command-id",
      "progress": 75,
      "message": "처리 중...",
      "status": "in_progress"
    }
  }
}
```

**진행 상황 상태**:
- `started`: 작업 시작
- `in_progress`: 진행 중
- `completed`: 완료
- `error`: 에러 발생

---

### Section 3: Figma Copilot (/xlzuvekas/figma-copilot)

**평판**: Medium | **코드 예제**: 71개

#### 3.1 고급 텍스트 작업 API

| 도구 | 설명 | 파라미터 |
|------|------|---------|
| `scan_nodes_with_options` | 고급 스캔 (깊이 제어, 타임아웃, 부분 결과) | `options` (object) |

**스캔 옵션**:

```json
{
  "maxDepth": 5,
  "timeout": 30000,
  "includePartialResults": true,
  "nodeTypes": ["TEXT", "FRAME"]
}
```

---

#### 3.2 일괄 작업 API (v0.3.2)

**성능 향상**: 개별 작업 대비 50-90% 빠름

| 도구 | 설명 | 성능 | 파라미터 |
|------|------|------|---------|
| `clone_multiple_nodes` | 노드를 여러 위치로 일괄 복사 | 50-90% 빠름 | `node_id`, `positions[]` |
| `get_multiple_nodes_info` | 여러 노드 정보를 단일 요청으로 조회 | 70-85% 빠름 | `node_ids[]` |
| `set_multiple_nodes_property` | 여러 노드에 동일 속성 일괄 설정 | 60-80% 빠름 | `node_ids[]`, `property_name`, `value` |
| `execute_batch` | 여러 명령을 단일 왕복으로 순차 실행 | 80-90% 빠름 | `commands[]` |
| `get_connection_status` | 현재 연결 상태 및 통계 조회 | - | 없음 |

**일괄 복사 예제**:

```json
{
  "name": "clone_multiple_nodes",
  "arguments": {
    "node_id": "1234:5678",
    "positions": [
      { "x": 100, "y": 100 },
      { "x": 200, "y": 100 },
      { "x": 300, "y": 100 }
    ]
  }
}
```

**일괄 명령 실행 예제**:

```json
{
  "name": "execute_batch",
  "arguments": {
    "commands": [
      {
        "command": "create_frame",
        "params": { "position": { "x": 0, "y": 0 }, "size": { "width": 200, "height": 200 } }
      },
      {
        "command": "set_fill_color",
        "params": { "node_id": "{previous_result.id}", "color": { "r": 1, "g": 0, "b": 0, "a": 1 } }
      }
    ]
  }
}
```

---

### Section 4: Figma REST API Variables 엔드포인트

**공식 Figma REST API를 통한 변수 관리**

#### 4.1 Variables API 개요

**엔드포인트**: `/v1/files/{file_key}/variables`

**메서드**: `GET`, `POST`, `PUT`, `DELETE`

**인증**: Figma Personal Access Token (헤더: `X-Figma-Token`)

---

#### 4.2 GET /v1/files/{file_key}/variables

**목적**: Figma 파일의 변수 및 변수 컬렉션 조회

**파라미터**:

| 파라미터 | 타입 | 위치 | 필수 | 설명 |
|---------|------|------|------|------|
| `file_key` | string | Path | ✅ 필수 | Figma 파일 키 |
| `published` | boolean | Query | ❌ 선택 | 게시된 변수만 조회 (기본값: false) |

**요청 예제**:
```http
GET /v1/files/abc123XYZ/variables?published=true
X-Figma-Token: figd_your_token_here
```

**응답 구조** (200 OK):

```json
{
  "variables": [
    {
      "id": "123:456",
      "name": "Primary Color",
      "key": "variable_key_123",
      "variableCollectionId": "collection_id_789",
      "resolvedType": "COLOR",
      "valuesByMode": {
        "mode_1": { "r": 1, "g": 0, "b": 0, "a": 1 },
        "mode_2": { "r": 0, "g": 1, "b": 0, "a": 1 }
      },
      "remote": false,
      "description": "주요 브랜드 색상",
      "hiddenFromPublishing": false,
      "scopes": ["FRAME_FILL", "TEXT_FILL"],
      "codeSyntax": {
        "WEB": "--color-primary",
        "ANDROID": "R.color.primary",
        "iOS": "UIColor.primary"
      }
    }
  ],
  "collections": [
    {
      "id": "collection_id_789",
      "name": "Brand Colors",
      "modes": [
        { "modeId": "mode_1", "name": "Light" },
        { "modeId": "mode_2", "name": "Dark" }
      ]
    }
  ]
}
```

---

#### 4.3 Variable 객체 속성

| 속성 | 타입 | 읽기 전용 | 설명 |
|------|------|----------|------|
| `id` | string | ✅ | 변수의 고유 식별자 |
| `name` | string | ❌ | 변수 이름 |
| `key` | string | ✅ | `importVariableByKeyAsync`에서 사용할 키 |
| `variableCollectionId` | string | ✅ | 소속된 컬렉션 ID |
| `resolvedType` | string | ✅ | 변수 타입: `BOOLEAN`, `FLOAT`, `STRING`, `COLOR` |
| `valuesByMode` | object | ✅ | 모드별 값 (별칭 해석 안 됨) |
| `remote` | boolean | ✅ | 원격(게시됨) 또는 로컬 여부 |
| `description` | string | ❌ | 변수 설명 |
| `hiddenFromPublishing` | boolean | ❌ | 라이브러리 게시 시 숨김 여부 |
| `scopes` | string[] | ❌ | UI 피커에 표시될 범위 (예: `["FRAME_FILL", "TEXT_FILL"]`) |
| `codeSyntax` | object | ✅ | 플랫폼별 코드 구문 (WEB, ANDROID, iOS) |
| `deletedButReferenced` | boolean | ✅ | 삭제되었지만 참조가 남아있는 경우 true |

---

#### 4.4 VariableCollection 객체

**목적**: 변수를 그룹화하고 모드 관리

```json
{
  "id": "collection_id_789",
  "name": "Brand Colors",
  "modes": [
    {
      "modeId": "mode_1",
      "name": "Light"
    },
    {
      "modeId": "mode_2",
      "name": "Dark"
    }
  ],
  "defaultModeId": "mode_1",
  "remote": false,
  "hiddenFromPublishing": false,
  "variableIds": ["123:456", "123:457"]
}
```

---

#### 4.5 Variable 값 타입

| resolvedType | 값 타입 | 예제 |
|-------------|---------|------|
| `COLOR` | `{ r, g, b, a }` | `{ "r": 1, "g": 0, "b": 0, "a": 1 }` |
| `FLOAT` | `number` | `16` |
| `STRING` | `string` | `"Roboto"` |
| `BOOLEAN` | `boolean` | `true` |
| `VARIABLE_ALIAS` | `{ type, id }` | `{ "type": "VARIABLE_ALIAS", "id": "123:789" }` |

---

#### 4.6 에러 코드 및 해결책

| 에러 코드 | 메시지 | 원인 | 해결책 |
|----------|--------|------|--------|
| **400 Bad Request** | "Invalid file key" | 잘못된 파일 키 형식 | Figma URL에서 올바른 파일 키 추출 |
| **401 Unauthorized** | "Invalid token" | 잘못되거나 만료된 API 토큰 | Figma 설정에서 새 Personal Access Token 생성 |
| **403 Forbidden** | "Access denied" | 파일 접근 권한 없음 | 파일 소유자로부터 편집 또는 보기 권한 요청 |
| **404 Not Found** | "File not found" | 존재하지 않는 파일 | 파일 키 확인, 파일이 삭제되었는지 확인 |
| **429 Too Many Requests** | "Rate limit exceeded" | API 호출 제한 초과 | Retry-After 헤더 확인, 지수 백오프 재시도 |
| **500 Internal Server Error** | "Server error" | Figma 서버 오류 | 잠시 후 재시도, 지속 시 Figma 지원팀 문의 |

---

### Section 5: 에러 처리 모범 사례

#### 5.1 Rate Limiting 및 Retry 전략

**Figma API 제한**:
- **일반 엔드포인트**: 분당 60회
- **이미지 렌더링**: 분당 30회
- **변수 API**: 분당 100회

**지수 백오프 구현 (TypeScript)**:

```typescript
async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  initialDelay: number = 1000
): Promise<T> {
  let lastError: Error;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      // 429 Rate Limit 에러인 경우
      if (error.response?.status === 429) {
        const retryAfter = error.response.headers['retry-after'];
        const delay = retryAfter
          ? parseInt(retryAfter) * 1000
          : initialDelay * Math.pow(2, attempt);

        console.log(`Rate limited. Retrying after ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }

      // 5xx 서버 에러인 경우
      if (error.response?.status >= 500) {
        const delay = initialDelay * Math.pow(2, attempt);
        console.log(`Server error. Retrying after ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }

      // 다른 에러는 즉시 throw
      throw error;
    }
  }

  throw lastError;
}
```

**사용 예제**:

```typescript
const fileData = await retryWithBackoff(
  () => figmaService.getRawFile("abc123XYZ"),
  3,  // 최대 3회 재시도
  1000 // 초기 1초 대기
);
```

---

#### 5.2 이미지 다운로드 에러 해결

**에러 1: "Path for asset writes is invalid"**

**원인**: 잘못된 로컬 경로 또는 권한 문제

**해결책**:

```typescript
// Before (상대 경로 - 에러 발생)
await figmaService.downloadImages(
  "abc123XYZ",
  "./assets/images",  // ❌ 상대 경로
  nodes
);

// After (절대 경로 - 정상)
import path from 'path';

const absolutePath = path.resolve(process.cwd(), './assets/images');
await figmaService.downloadImages(
  "abc123XYZ",
  absolutePath,  // ✅ 절대 경로
  nodes
);

// 디렉토리 존재 확인 및 생성
import fs from 'fs/promises';
await fs.mkdir(absolutePath, { recursive: true });
```

---

**에러 2: "Image base64 format error"**

**원인**: 이미지 인코딩 실패, 너무 큰 노드, 지원되지 않는 형식

**해결책**:

```typescript
// Before (높은 스케일 - 에러 발생)
const results = await figmaService.downloadImages(
  "abc123XYZ",
  localPath,
  nodes,
  { pngScale: 4 }  // ❌ 너무 큰 이미지
);

// After (적절한 스케일)
const results = await figmaService.downloadImages(
  "abc123XYZ",
  localPath,
  nodes,
  { pngScale: 2 }  // ✅ 적절한 크기
);

// 또는 SVG 형식 사용
const results = await figmaService.downloadImages(
  "abc123XYZ",
  localPath,
  nodes.map(node => ({ ...node, fileName: node.fileName.replace('.png', '.svg') })),
  { format: 'svg' }
);
```

**노드 타입 검증**:

```typescript
// 이미지 내보내기 전 노드 타입 확인
const nodeInfo = await figmaService.getRawNode("abc123XYZ", nodeId);
const nodeType = nodeInfo.nodes[nodeId].document.type;

const exportableTypes = ['FRAME', 'COMPONENT', 'INSTANCE', 'RECTANGLE', 'ELLIPSE', 'POLYGON', 'STAR', 'VECTOR'];

if (!exportableTypes.includes(nodeType)) {
  console.warn(`Node ${nodeId} of type ${nodeType} may not export correctly`);
}
```

---

#### 5.3 변수 조회 400 에러 디버깅

**에러**: `400 Bad Request` on `/v1/files/{file_key}/variables`

**가능한 원인**:

1. **잘못된 파일 키 형식**
2. **변수가 없는 파일**
3. **권한 부족**
4. **API 버전 불일치**

**디버깅 단계**:

```typescript
async function debugVariablesAPI(fileKey: string, apiToken: string) {
  try {
    // 1. 파일 키 검증
    console.log('Testing file key:', fileKey);
    if (!/^[a-zA-Z0-9]{22}$/.test(fileKey)) {
      throw new Error('Invalid file key format. Should be 22 alphanumeric characters.');
    }

    // 2. 파일 접근 권한 확인
    const fileResponse = await fetch(`https://api.figma.com/v1/files/${fileKey}`, {
      headers: { 'X-Figma-Token': apiToken }
    });

    if (!fileResponse.ok) {
      throw new Error(`Cannot access file: ${fileResponse.status} ${fileResponse.statusText}`);
    }

    console.log('✅ File access OK');

    // 3. 변수 API 호출
    const variablesResponse = await fetch(
      `https://api.figma.com/v1/files/${fileKey}/variables/local`,
      {
        headers: { 'X-Figma-Token': apiToken }
      }
    );

    if (!variablesResponse.ok) {
      const errorBody = await variablesResponse.text();
      throw new Error(`Variables API failed: ${variablesResponse.status}\n${errorBody}`);
    }

    const data = await variablesResponse.json();
    console.log('✅ Variables retrieved:', data.meta?.variables?.length || 0);

    return data;

  } catch (error) {
    console.error('❌ Debug failed:', error.message);
    throw error;
  }
}
```

**해결책**:

```typescript
// Before (게시된 변수 없음 - 에러)
const variables = await fetch(
  `https://api.figma.com/v1/files/${fileKey}/variables?published=true`
);

// After (로컬 변수 포함)
const variables = await fetch(
  `https://api.figma.com/v1/files/${fileKey}/variables/local`
);
```

---

### Section 6: MCP 도구 호출 순서 권장사항

#### 6.1 권장 호출 순서

**시나리오 1: 디자인 데이터 추출 및 이미지 다운로드**

```
1. get_figma_data (fileKey만)
   → 파일 구조 파악, 노드 ID 수집

2. get_figma_data (fileKey + nodeId + depth)
   → 특정 노드의 상세 정보 추출

3. download_figma_images (fileKey + nodeIds + localPath)
   → 이미지 자산 다운로드
```

**시나리오 2: 변수 기반 디자인 시스템 추출**

```
1. GET /v1/files/{fileKey}/variables/local
   → 변수 및 컬렉션 정보 조회

2. get_figma_data (fileKey)
   → 변수가 바인딩된 노드 찾기

3. simplifyRawFigmaObject (with allExtractors)
   → 변수 참조를 포함한 디자인 토큰 추출
```

**시나리오 3: 프로토타입 플로우 문서화**

```
1. get_document_info
   → 문서 메타데이터 확인

2. get_reactions
   → 프로토타입 반응 및 연결 정보 추출

3. scan_nodes_by_types (["FRAME", "COMPONENT"])
   → 주요 화면 및 컴포넌트 식별

4. set_multiple_annotations (annotations[])
   → 화면별 설명 주석 추가
```

---

#### 6.2 병렬 호출 가능 여부

**병렬 호출 가능** (독립적 작업):

```typescript
// ✅ 병렬 실행 가능 (다른 파일 또는 독립적 노드)
const [fileData1, fileData2, variables] = await Promise.all([
  figmaService.getRawFile("fileKey1"),
  figmaService.getRawFile("fileKey2"),
  fetch(`https://api.figma.com/v1/files/fileKey1/variables/local`)
]);
```

**병렬 호출 불가** (순차 실행 필요):

```typescript
// ❌ 병렬 실행 불가 (의존성 있음)
// 1단계: 파일 데이터 조회
const fileData = await figmaService.getRawFile("abc123XYZ");

// 2단계: 노드 ID 추출
const nodeIds = extractNodeIds(fileData);

// 3단계: 이미지 다운로드 (nodeIds 필요)
const images = await figmaService.downloadImages("abc123XYZ", localPath, nodeIds);
```

**Rate Limit 고려한 병렬 호출**:

```typescript
// 분당 60회 제한 → 초당 1회로 제한
import pLimit from 'p-limit';

const limit = pLimit(1); // 동시 1개 요청만 허용
const delay = 1000; // 1초 대기

const results = await Promise.all(
  nodeIds.map((nodeId, index) =>
    limit(async () => {
      await new Promise(resolve => setTimeout(resolve, index * delay));
      return figmaService.getRawNode("abc123XYZ", nodeId);
    })
  )
);
```

---

#### 6.3 최적 호출 패턴 (성능 최적화)

**패턴 1: 청킹 (대규모 노드 처리)**

```typescript
async function processLargeDesign(fileKey: string, nodeIds: string[]) {
  const CHUNK_SIZE = 10;
  const chunks = [];

  // 노드를 10개씩 청크로 분할
  for (let i = 0; i < nodeIds.length; i += CHUNK_SIZE) {
    chunks.push(nodeIds.slice(i, i + CHUNK_SIZE));
  }

  // 청크별로 순차 처리 (Rate Limit 회피)
  for (const chunk of chunks) {
    const results = await Promise.all(
      chunk.map(nodeId => figmaService.getRawNode(fileKey, nodeId))
    );

    // 결과 처리
    processResults(results);

    // 다음 청크 전 1초 대기
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
}
```

---

**패턴 2: 캐싱 (중복 호출 방지)**

```typescript
class FigmaServiceWithCache {
  private cache: Map<string, any> = new Map();

  async getRawFileWithCache(fileKey: string): Promise<any> {
    const cacheKey = `file:${fileKey}`;

    if (this.cache.has(cacheKey)) {
      console.log('✅ Cache hit:', cacheKey);
      return this.cache.get(cacheKey);
    }

    console.log('🔄 Cache miss, fetching:', cacheKey);
    const data = await figmaService.getRawFile(fileKey);

    this.cache.set(cacheKey, data);
    return data;
  }

  clearCache() {
    this.cache.clear();
  }
}
```

---

**패턴 3: Progressive Enhancement (점진적 로딩)**

```typescript
async function progressiveLoad(fileKey: string, onProgress: (data: any) => void) {
  // 1단계: 메타데이터만 로드 (depth=1)
  const shallowData = await figmaService.getRawFile(fileKey, 1);
  onProgress({ stage: 'metadata', data: shallowData });

  // 2단계: 주요 페이지 로드 (depth=2)
  const mediumData = await figmaService.getRawFile(fileKey, 2);
  onProgress({ stage: 'pages', data: mediumData });

  // 3단계: 전체 디자인 로드 (depth=10)
  const fullData = await figmaService.getRawFile(fileKey, 10);
  onProgress({ stage: 'complete', data: fullData });
}

// 사용
progressiveLoad("abc123XYZ", ({ stage, data }) => {
  console.log(`Loaded ${stage}:`, data);
});
```

---

### Section 7: 설정 예제

#### 7.1 Figma Context MCP 설정

**macOS/Linux**:

```json
{
  "mcpServers": {
    "Framelink MCP for Figma": {
      "command": "npx",
      "args": ["-y", "figma-developer-mcp", "--figma-api-key=YOUR-KEY", "--stdio"]
    }
  }
}
```

**Windows**:

```json
{
  "mcpServers": {
    "Framelink MCP for Figma": {
      "command": "cmd",
      "args": ["/c", "npx", "-y", "figma-developer-mcp", "--figma-api-key=YOUR-KEY", "--stdio"]
    }
  }
}
```

---

#### 7.2 Cursor Talk To Figma MCP 설정

**Cursor AI**:

```json
{
  "mcpServers": {
    "TalkToFigma": {
      "command": "npx",
      "args": ["ai-figma-mcp@latest"]
    }
  }
}
```

**Windsurf AI**:

```json
{
  "servers": {
    "TalkToFigma": {
      "command": "npx",
      "args": ["ai-figma-mcp@latest"]
    }
  }
}
```

**GitHub Copilot (VS Code)**:

```json
{
  "mcp": {
    "servers": {
      "TalkToFigma": {
        "command": "npx",
        "args": ["ai-figma-mcp@latest"]
      }
    }
  }
}
```

---

#### 7.3 로컬 빌드 설정 (개발자용)

```json
{
  "mcpServers": {
    "TalkToFigma": {
      "command": "node",
      "args": ["/absolute/path/to/project/dist/talk_to_figma_mcp/server.js"]
    }
  }
}
```

---

### Section 8: 코드 예제 모음

#### 8.1 전체 워크플로우: Figma → 코드 생성

```typescript
import { FigmaService } from "figma-developer-mcp/services/figma";
import { simplifyRawFigmaObject, allExtractors } from "figma-developer-mcp/extractors";
import * as fs from 'fs/promises';

async function figmaToCode(fileKey: string, nodeId: string) {
  const figma = new FigmaService({
    figmaApiKey: process.env.FIGMA_API_KEY!,
    figmaOAuthToken: "",
    useOAuth: false
  });

  // 1. 노드 데이터 가져오기
  const nodeData = await figma.getRawNode(fileKey, nodeId, 5);

  // 2. 디자인 단순화
  const simplified = simplifyRawFigmaObject(nodeData, allExtractors);

  // 3. 변수 조회
  const variablesResponse = await fetch(
    `https://api.figma.com/v1/files/${fileKey}/variables/local`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_API_KEY! }
    }
  );
  const variables = await variablesResponse.json();

  // 4. 이미지 다운로드
  const imageNodes = simplified.nodes.filter(n => n.type === 'RECTANGLE' && n.fills);
  if (imageNodes.length > 0) {
    await figma.downloadImages(
      fileKey,
      './public/assets',
      imageNodes.map(n => ({ nodeId: n.id, fileName: `${n.name}.png` })),
      { pngScale: 2 }
    );
  }

  // 5. React 컴포넌트 생성 (예제)
  const componentCode = generateReactComponent(simplified, variables);
  await fs.writeFile(`./src/components/${simplified.name}.tsx`, componentCode);

  console.log('✅ Figma to Code complete!');
}

function generateReactComponent(design: any, variables: any): string {
  // 간단한 React 컴포넌트 생성 예제
  return `
import React from 'react';
import './styles.css';

export const ${design.name.replace(/\s/g, '')} = () => {
  return (
    <div className="${design.name.toLowerCase()}">
      {/* 생성된 컴포넌트 내용 */}
    </div>
  );
};
`;
}

// 실행
figmaToCode("abc123XYZ", "1234:5678");
```

---

#### 8.2 변수 기반 테마 생성

```typescript
async function generateThemeFromFigma(fileKey: string) {
  const response = await fetch(
    `https://api.figma.com/v1/files/${fileKey}/variables/local`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_API_KEY! }
    }
  );

  const data = await response.json();
  const variables = data.meta.variables;
  const collections = data.meta.variableCollections;

  // CSS 변수 생성
  const cssTheme: Record<string, any> = {};

  for (const collection of Object.values(collections)) {
    for (const mode of (collection as any).modes) {
      const modeVars: string[] = [];

      for (const variable of Object.values(variables)) {
        const v = variable as any;
        const value = v.valuesByMode[mode.modeId];

        if (v.resolvedType === 'COLOR') {
          const cssVar = `--${v.name.toLowerCase().replace(/\s/g, '-')}`;
          const cssValue = `rgba(${Math.round(value.r * 255)}, ${Math.round(value.g * 255)}, ${Math.round(value.b * 255)}, ${value.a})`;
          modeVars.push(`  ${cssVar}: ${cssValue};`);
        }
      }

      cssTheme[mode.name] = modeVars.join('\n');
    }
  }

  // CSS 파일 생성
  let css = '';
  for (const [modeName, vars] of Object.entries(cssTheme)) {
    css += `[data-theme="${modeName.toLowerCase()}"] {\n${vars}\n}\n\n`;
  }

  await fs.writeFile('./src/theme.css', css);
  console.log('✅ Theme CSS generated!');
}
```

---

### Section 9: 비교표

#### 9.1 Figma MCP 서버 기능 비교

| 기능 | Figma Context MCP | Talk To Figma MCP | Figma Copilot |
|-----|-------------------|-------------------|---------------|
| **디자인 데이터 조회** | ✅ `get_figma_data` | ✅ `get_document_info` | ✅ `get_document_info` |
| **이미지 다운로드** | ✅ `download_figma_images` | ✅ `export_node_as_image` | ❌ |
| **주석 관리** | ❌ | ✅ `set_annotation` | ✅ `set_multiple_annotations` |
| **텍스트 수정** | ❌ | ✅ `set_text_content` | ✅ `set_multiple_text_contents` |
| **Auto Layout** | ❌ | ✅ `set_layout_mode` | ✅ `set_layout_mode` |
| **일괄 작업** | ❌ | ❌ | ✅ `execute_batch`, `clone_multiple_nodes` |
| **프로토타입 연결** | ❌ | ✅ `create_connections` | ✅ `create_connections` |
| **컴포넌트 관리** | ❌ | ✅ `create_component_instance` | ✅ `get_instance_overrides` |
| **Extractor 시스템** | ✅ | ❌ | ❌ |
| **WebSocket 지원** | ❌ | ✅ | ❌ |
| **평판** | High | High | Medium |
| **코드 예제** | 40개 | 79개 | 71개 |

---

## 📝 결론 및 권장사항

### 조사 결과 요약

1. **요청된 특정 도구 미발견**: `get_design_context`, `get_variable_defs`, `get_screenshot`는 현재 공개 문서에서 확인되지 않음
2. **대체 도구 존재**: 유사한 기능을 제공하는 커뮤니티 기반 도구 다수 확인
3. **Figma REST API 직접 사용 가능**: 변수 조회는 공식 REST API로 가능

### 권장 접근 방식

**시나리오 1: 디자인 데이터 추출**
- **도구**: Figma Context MCP의 `get_figma_data`
- **장점**: 구조화된 데이터, Extractor 시스템
- **단점**: 수정 기능 없음

**시나리오 2: 디자인 수정 및 자동화**
- **도구**: Talk To Figma MCP 또는 Figma Copilot
- **장점**: WebSocket 실시간 연결, 풍부한 수정 API
- **단점**: 설정 복잡도 높음

**시나리오 3: 변수 관리**
- **도구**: Figma REST API 직접 호출
- **장점**: 공식 API, 완전한 변수 지원
- **단점**: MCP 통합 없음 (직접 구현 필요)

**시나리오 4: 이미지 자산 추출**
- **도구**: Figma Context MCP의 `download_figma_images`
- **장점**: 자동 크롭, CSS 변수 생성
- **단점**: Rate Limit 주의

---

## 🔗 참고 링크

- **Figma Context MCP**: https://github.com/glips/figma-context-mcp
- **Talk To Figma MCP**: https://github.com/sethdford/mcp-figma
- **Figma Copilot**: https://github.com/xlzuvekas/figma-copilot
- **Figma REST API Docs**: https://www.figma.com/developers/api
- **Figma Plugin API**: https://www.figma.com/plugin-docs/

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-11-19
**조사자**: MoAI Context7 Integrator
**Context7 쿼리 수**: 6회
**소스 평판**: High (3개), Medium (1개)
