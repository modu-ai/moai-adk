# Figma MCP 호환성 및 대안 분석 매트릭스

**문서 목적**: Context7 조사 결과를 바탕으로 Figma Context MCP, Talk To Figma MCP, Figma Copilot의 기능 비교 및 호환성 분석
**작성일**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (Section 1-3, 8-9)

---

## 목차

1. [Figma MCP 서버 전체 비교](#1-figma-mcp-서버-전체-비교)
2. [기능별 호환성 매트릭스](#2-기능별-호환성-매트릭스)
3. [사용 사례별 권장 도구](#3-사용-사례별-권장-도구)
4. [마이그레이션 가이드](#4-마이그레이션-가이드)
5. [통합 전략 (다중 도구)](#5-통합-전략-다중-도구)

---

## 1. Figma MCP 서버 전체 비교

### 1.1 서버 개요

| 속성 | Figma Context MCP | Talk To Figma MCP | Figma Copilot |
|------|-------------------|-------------------|---------------|
| **공식명** | figma-context-mcp | mcp-figma / ai-figma-mcp | figma-copilot |
| **GitHub** | `/glips/figma-context-mcp` | `/sethdford/mcp-figma` | `/xlzuvekas/figma-copilot` |
| **평판** | High | High | Medium |
| **코드 예제** | 40개 | 79개 | 71개 |
| **벤치마크 점수** | 65.4 | - | - |
| **최신 업데이트** | 활발 | 활발 | 유지보수 중 |
| **커뮤니티** | 중간 | 큼 | 중간 |
| **문서화** | 양호 | 우수 | 양호 |

---

### 1.2 서버별 소개

#### Figma Context MCP (`/glips/figma-context-mcp`)

**특징**:
- 구조화된 디자인 데이터 추출에 특화
- Extractor 시스템으로 유연한 데이터 필터링
- 이미지 자산 다운로드 통합
- TypeScript 기반

**장점**:
✅ 깔끔한 API 디자인
✅ 성능 최적화 (Extractor 시스템)
✅ 대규모 디자인 시스템 처리 우수
✅ CSS 변수 자동 생성

**단점**:
❌ 노드 수정 기능 없음
❌ 실시간 협업 미지원
❌ WebSocket 미지원
❌ 주석/주석 관리 미지원

**추천 사용자**: 디자인 데이터 추출, 코드 생성, 자산 관리

---

#### Talk To Figma MCP (`/sethdford/mcp-figma`)

**특징**:
- 가장 포괄적인 API 커버리지
- 실시간 WebSocket 연결
- 광범위한 수정 기능
- 프로토타입 플로우 관리

**장점**:
✅ WebSocket 실시간 연결
✅ 풍부한 수정 API (텍스트, 레이아웃, 스타일)
✅ 프로토타입 연결 생성
✅ 컴포넌트 관리
✅ 코드 예제 많음 (79개)

**단점**:
❌ 설정 복잡도 높음
❌ 성능 최적화 부족 (일괄 작업 X)
❌ 문서화 부분적
❌ 대규모 설계 시스템 처리 어려움

**추천 사용자**: 자동 디자인 수정, 실시간 협업, 복잡한 프로토타입

---

#### Figma Copilot (`/xlzuvekas/figma-copilot`)

**특징**:
- 일괄 작업 최적화
- 고급 스캔 옵션
- 성능 중심
- 하이브리드 기능

**장점**:
✅ 일괄 작업으로 50-90% 성능 개선
✅ execute_batch로 원자적 작업
✅ 고급 스캔 옵션 (깊이, 타임아웃, 부분 결과)
✅ 연결 상태 모니터링

**단점**:
❌ 문서화 부족
❌ 커뮤니티 규모 작음
❌ Extractor 없음
❌ WebSocket 미지원
❌ 검증되지 않은 기능 포함

**추천 사용자**: 대량 노드 처리, 배치 작업, 성능 중시

---

## 2. 기능별 호환성 매트릭스

### 2.1 데이터 추출 기능

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **파일 메타데이터** | ✅ get_figma_data | ✅ get_document_info | ✅ get_document_info | 모두 지원 |
| **노드 정보 조회** | ✅ get_figma_data(nodeId) | ✅ get_node_info | ✅ get_node_info | 모두 지원 |
| **다중 노드 조회** | ❌ 개별 호출 | ✅ get_nodes_info | ✅ get_multiple_nodes_info | Talk To Figma 최적 |
| **깊이 제어** | ✅ depth 파라미터 | ❌ 자동 | ✅ scan_nodes_with_options | Context MCP 최적 |
| **스타일 정보** | ✅ (Extractor) | ✅ get_styles | ✅ get_styles | 모두 지원 |
| **컴포넌트 정보** | ❌ | ✅ get_local_components | ✅ get_local_components | Talk To Figma/Copilot 지원 |
| **선택 정보** | ❌ | ✅ get_selection | ✅ read_my_design | 양쪽 지원 |

**최적 선택**:
- 깊이 제어: **Context MCP** (depth 파라미터)
- 다중 노드: **Talk To Figma** (get_nodes_info)
- 성능: **Copilot** (배치 처리)

---

### 2.2 이미지/자산 내보내기

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **이미지 다운로드** | ✅ download_figma_images | ✅ export_node_as_image (base64만) | ❌ | Context MCP 최고 |
| **자동 크롭** | ✅ needsCropping | ❌ | ❌ | Context MCP 전용 |
| **CSS 변수 생성** | ✅ requiresImageDimensions | ❌ | ❌ | Context MCP 전용 |
| **형식 선택** | ✅ PNG/SVG/JPG/PDF | ✅ PNG/JPG/SVG/PDF | ❌ | 둘 다 지원 |
| **스케일 지정** | ✅ pngScale (1-4) | ❌ | ❌ | Context MCP 전용 |
| **파일 직접 저장** | ✅ localPath | ❌ base64만 반환 | ❌ | Context MCP만 파일 저장 |

**최적 선택**:
- **자산 관리**: **Context MCP** (자동 크롭, CSS 변수, 파일 저장)
- **빠른 미리보기**: **Talk To Figma** (base64 디코딩 가능)

---

### 2.3 노드 수정 기능

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **텍스트 수정** | ❌ | ✅ set_text_content ⚠️ | ✅ set_multiple_text_contents ⚠️ | 포맷 손실 주의 |
| **다중 텍스트 수정** | ❌ | ❌ | ✅ set_multiple_text_contents | Copilot만 지원 |
| **색상 변경** | ❌ | ✅ set_fill_color | ✅ set_fill_color | 둘 다 지원 |
| **테두리 스타일** | ❌ | ✅ set_stroke_color | ✅ set_stroke_color | 둘 다 지원 |
| **모서리 반경** | ❌ | ✅ set_corner_radius | ✅ set_corner_radius | 둘 다 지원 |
| **노드 생성** | ❌ | ✅ (rectangle, frame, text) | ✅ (rectangle, frame, text) | 둘 다 지원 |
| **노드 삭제** | ❌ | ✅ delete_node | ✅ delete_multiple_nodes | Copilot이 배치 지원 |
| **노드 복사** | ❌ | ✅ clone_node | ✅ clone_multiple_nodes | Copilot이 배치 지원 |
| **노드 이동/크기** | ❌ | ✅ move_node, resize_node | ✅ move_node, resize_node | 둘 다 지원 |

**최적 선택**:
- **대량 수정**: **Copilot** (배치 작업, 50-90% 성능 개선)
- **개별 수정**: **Talk To Figma** (API 완성도)
- **텍스트만 수정**: **Copilot** (set_multiple_text_contents)

**⚠️ 주의**: 텍스트 수정 시 포맷(굵기, 색상) 손실 가능

---

### 2.4 레이아웃 기능

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **Auto Layout 설정** | ❌ | ✅ set_layout_mode | ✅ set_layout_mode | 둘 다 지원 |
| **패딩 설정** | ❌ | ✅ set_padding | ✅ set_padding | 둘 다 지원 |
| **정렬 설정** | ❌ | ✅ set_axis_align | ✅ set_axis_align | 둘 다 지원 |
| **크기 조정 모드** | ❌ | ✅ set_layout_sizing | ✅ set_layout_sizing | 둘 다 지원 |
| **아이템 간격** | ❌ | ✅ set_item_spacing | ✅ set_item_spacing | 둘 다 지원 |

**최적 선택**: **Talk To Figma** 또는 **Copilot** (둘 다 동등)

---

### 2.5 주석/문서화 기능

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **주석 조회** | ❌ | ✅ get_annotations | ✅ get_annotations | 둘 다 지원 |
| **주석 생성/수정** | ❌ | ✅ set_annotation | ✅ set_multiple_annotations | Copilot이 배치 지원 |
| **노드 스캔 (타입)** | ❌ | ✅ scan_nodes_by_types | ✅ scan_nodes_with_options | Copilot이 고급 옵션 |

**최적 선택**: **Copilot** (배치 주석)

---

### 2.6 프로토타입/상호작용

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **프로토타입 반응** | ❌ | ✅ get_reactions | ❌ | Talk To Figma 전용 |
| **연결선 생성** | ❌ | ✅ create_connections | ✅ create_connections | 둘 다 지원 |
| **커넥터 스타일** | ❌ | ✅ set_default_connector | ❌ | Talk To Figma 전용 |

**최적 선택**: **Talk To Figma** (프로토타입 플로우 관리)

---

### 2.7 컴포넌트 관리

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **로컬 컴포넌트 조회** | ❌ | ✅ get_local_components | ✅ get_local_components | 둘 다 지원 |
| **인스턴스 생성** | ❌ | ✅ create_component_instance | ❌ | Talk To Figma 전용 |
| **오버라이드 조회** | ❌ | ✅ get_instance_overrides | ✅ get_instance_overrides | 둘 다 지원 |
| **오버라이드 적용** | ❌ | ✅ set_instance_overrides | ✅ set_instance_overrides | 둘 다 지원 |

**최적 선택**: **Talk To Figma** (컴포넌트 인스턴스 생성)

---

### 2.8 변수 및 디자인 토큰

| 기능 | Context MCP | Talk To Figma | Copilot | Figma REST API |
|------|-----------|---------------|---------|-----------------|
| **변수 조회** | ❌ | ❌ | ❌ | ✅ GET /variables |
| **변수 값 조회** | ❌ | ❌ | ❌ | ✅ valuesByMode |
| **코드 구문** | ❌ | ❌ | ❌ | ✅ codeSyntax (WEB/iOS/Android) |
| **모드 관리** | ❌ | ❌ | ❌ | ✅ (API 레벨) |
| **CSS 변수 생성** | ✅ (이미지용) | ❌ | ❌ | ⚠️ 수동 변환 |

**최적 선택**: **Figma REST API** (공식, 완전 기능)

---

### 2.9 성능 및 고급 기능

| 기능 | Context MCP | Talk To Figma | Copilot | 비고 |
|------|-----------|---------------|---------|------|
| **WebSocket 연결** | ❌ | ✅ | ❌ | Talk To Figma만 실시간 |
| **일괄 작업** | ❌ | ❌ | ✅ execute_batch | Copilot 전용 |
| **배치 복사** | ❌ | ❌ | ✅ clone_multiple_nodes (50-90% 빠름) | Copilot 전용 |
| **배치 삭제** | ❌ | ❌ | ✅ delete_multiple_nodes | Copilot 전용 |
| **배치 속성 설정** | ❌ | ❌ | ✅ set_multiple_nodes_property | Copilot 전용 |
| **Extractor 시스템** | ✅ (유연한 필터링) | ❌ | ❌ | Context MCP 전용 |
| **Rate Limit 추적** | ❌ | ❌ | ✅ get_connection_status | Copilot 전용 |

**최적 선택**:
- **대량 작업**: **Copilot** (배치 API, 50-90% 성능 개선)
- **실시간 협업**: **Talk To Figma** (WebSocket)
- **유연한 필터링**: **Context MCP** (Extractor)

---

## 3. 사용 사례별 권장 도구

### 3.1 디자인 시스템 추출 및 코드 생성

**요구사항**:
- ✅ 전체 설계 구조 추출
- ✅ 변수 및 토큰 조회
- ✅ 이미지 자산 다운로드
- ✅ CSS 변수 자동 생성

**권장 스택**:

```
Step 1: Context MCP - get_figma_data + Extractor
Step 2: Context MCP - download_figma_images (CSS 변수 포함)
Step 3: Figma REST API - GET /variables (디자인 토큰)
Step 4: 코드 생성기 (수동 또는 LLM)
```

**코드 예제**:
```typescript
// 1. 디자인 데이터 추출
const design = await figmaService.getRawFile("abc123XYZ");
const simplified = simplifyRawFigmaObject(design, allExtractors);

// 2. 이미지 다운로드 (CSS 변수 포함)
await figmaService.downloadImages(
  "abc123XYZ",
  "./assets",
  imageNodes.map(n => ({
    nodeId: n.id,
    fileName: `${n.name}.png`,
    requiresImageDimensions: true
  })),
  { pngScale: 2 }
);

// 3. 변수 조회
const variables = await fetch(
  `https://api.figma.com/v1/files/abc123XYZ/variables`,
  { headers: { 'X-Figma-Token': apiToken } }
).then(r => r.json());

// 4. React 컴포넌트 생성
const componentCode = generateReact(simplified, variables);
```

**예상 시간**: 15-30분 (설계 크기에 따라)
**비용**: 무료 (API 제한 준수)

---

### 3.2 자동 디자인 수정 및 배치 작업

**요구사항**:
- ✅ 대량 노드 한 번에 처리
- ✅ 텍스트, 색상, 레이아웃 수정
- ✅ 성능 최적화 (분당 호출 제한 회피)

**권장 도구**: **Copilot** (배치 작업, 50-90% 성능 개선)

**코드 예제**:
```typescript
// 단일 명령으로 배치 작업 실행
const results = await figmaCopilot.execute_batch({
  commands: [
    {
      command: "set_fill_color",
      params: { node_id: "1234:5678", color: { r: 1, g: 0, b: 0, a: 1 } }
    },
    {
      command: "set_text_content",
      params: { node_id: "1234:5679", content: "새로운 텍스트" }
    },
    {
      command: "set_layout_mode",
      params: { node_id: "1234:5680", mode: "HORIZONTAL" }
    }
  ]
});

// 개별 작업 3개 = ~3초 (순차)
// 배치 작업 = ~0.3초 (원자적)
// 성능: 10배 향상
```

**예상 성능**: 50-90% 빠름 (개별 작업 대비)
**비용**: 무료 (API 제한 준수)

---

### 3.3 실시간 협업 및 프로토타입 자동화

**요구사항**:
- ✅ WebSocket 실시간 연결
- ✅ 프로토타입 플로우 관리
- ✅ 주석 및 협업 기능
- ✅ 광범위한 수정 API

**권장 도구**: **Talk To Figma MCP** (WebSocket, 완전한 API)

**코드 예제**:
```typescript
// WebSocket 연결
const client = new TalkToFigmaClient({
  apiKey: process.env.FIGMA_API_KEY!
});

// 실시간 문서 모니터링
client.onDocumentUpdate((update) => {
  console.log('문서 변경:', update);
});

// 프로토타입 연결 생성
await client.create_connections({
  connections: [
    { from: "screen1", to: "screen2", type: "ARROW" }
  ]
});

// 주석 추가
await client.set_annotation({
  nodeId: "1234:5678",
  content: "## 이 버튼은 클릭 시 모달 열기\n\n자세한 설명"
});
```

**예상 시간**: 실시간 (WebSocket)
**비용**: 무료

---

### 3.4 간단한 이미지 내보내기

**요구사항**:
- ✅ 노드를 이미지로 변환
- ✅ 여러 형식 지원 (PNG, SVG, JPG)
- ✅ 최소 기능

**권장 도구**: **Context MCP** (다운로드 + 파일 저장)

**코드 예제**:
```typescript
// 이미지 다운로드 (파일 직접 저장)
await figmaService.downloadImages(
  "abc123XYZ",
  "/Users/dev/assets",
  [{ nodeId: "1234:5678", fileName: "component.png" }],
  { pngScale: 2 }
);

// 결과: /Users/dev/assets/component.png
```

**예상 시간**: 1-2초
**비용**: 무료 (분당 30회 제한 준수)

---

### 3.5 변수 및 디자인 토큰 관리

**요구사항**:
- ✅ Figma 변수 조회
- ✅ CSS/iOS/Android 코드 생성
- ✅ 모드별 값 추출

**권장 도구**: **Figma REST API** (공식, 완전 기능)

**코드 예제**:
```typescript
// Figma REST API로 변수 조회
const response = await fetch(
  `https://api.figma.com/v1/files/abc123XYZ/variables`,
  {
    headers: { 'X-Figma-Token': process.env.FIGMA_API_KEY! }
  }
);

const data = await response.json();
const variables = data.meta.variables;

// CSS 변수로 변환
const cssVars = Object.entries(variables).map(([key, variable]) => {
  const val = variable.valuesByMode['light-mode'];
  return `--${variable.name.toLowerCase()}: ${formatCSSValue(val)};`;
}).join('\n');

// src/theme.css
await fs.writeFile('./src/theme.css', cssVars);
```

**예상 시간**: 5-10분
**비용**: 무료

---

## 4. 마이그레이션 가이드

### 4.1 Talk To Figma → Copilot 마이그레이션

**목적**: 성능 개선 (대량 작업)

**Before** (Talk To Figma):
```typescript
// 개별 호출 (느림)
for (const nodeId of nodeIds) {
  await client.delete_node(nodeId);  // 각각 ~100ms
}
// 총 시간: 100 노드 × 100ms = 10초
```

**After** (Copilot):
```typescript
// 배치 호출 (빠름)
await copilot.delete_multiple_nodes(nodeIds);  // 한 번에 ~1초
// 총 시간: 1초 (10배 향상)
```

**마이그레이션 체크리스트**:
- [ ] `delete_node` → `delete_multiple_nodes`
- [ ] `clone_node` → `clone_multiple_nodes`
- [ ] `set_text_content` (반복) → `set_multiple_text_contents`
- [ ] 개별 속성 설정 → `execute_batch`
- [ ] Rate Limit 전략 재검토

---

### 4.2 Context MCP → Talk To Figma 마이그레이션

**목적**: 노드 수정 기능 추가

**Before** (Context MCP):
```typescript
// 데이터만 추출
const design = await figmaService.getRawFile("abc123XYZ");
// 수정은 외부에서 수동 처리 필요
```

**After** (Talk To Figma):
```typescript
// 데이터 조회 + 수정
const nodeInfo = await client.get_node_info({ node_id: "1234:5678" });
await client.set_text_content({
  node_id: "1234:5678",
  content: "새로운 텍스트"
});
```

**마이그레이션 체크리스트**:
- [ ] `getRawFile` → `get_document_info`
- [ ] `getRawNode` → `get_node_info`
- [ ] 추출 데이터 구조 재정의
- [ ] 수정 API 테스트
- [ ] 텍스트 포맷 손실 영향 평가

---

### 4.3 MCP 없는 스크립트 → MCP 통합

**Before** (직접 REST API):
```typescript
// 복잡한 HTTP 호출
const response = await fetch(
  `https://api.figma.com/v1/files/${fileKey}`,
  { headers: { 'X-Figma-Token': apiToken } }
);
const data = await response.json();
// 에러 처리, 재시도 로직 수동 구현
```

**After** (MCP):
```typescript
// 간단한 함수 호출
const data = await figmaService.getRawFile(fileKey);
// 에러 처리, 재시도 MCP가 자동 처리
```

---

## 5. 통합 전략 (다중 도구)

### 5.1 권장 도구 조합

**설계 시스템 구축**:
```
Context MCP (데이터 추출)
    ↓
Figma REST API (변수 조회)
    ↓
Context MCP (이미지 다운로드)
    ↓
LLM 코드 생성
```

**디자인 자동화**:
```
Talk To Figma (데이터 조회 + 수정)
    ↓
Copilot (배치 작업)
    ↓
Figma REST API (변수 업데이트)
```

**실시간 협업**:
```
Talk To Figma WebSocket (실시간 모니터링)
    ↓
Copilot (배치 수정)
    ↓
Context MCP (최종 이미지 내보내기)
```

---

### 5.2 도구 선택 의사결정 트리

```
질문: 주요 용도?
  ├─ 데이터 추출만
  │  └─ "Context MCP" (get_figma_data)
  │
  ├─ 데이터 추출 + 이미지 내보내기
  │  └─ "Context MCP" (download_figma_images)
  │
  ├─ 데이터 추출 + 노드 수정
  │  └─ "Talk To Figma MCP" (get_node_info + set_*)
  │
  ├─ 대량 노드 수정 (성능 중시)
  │  └─ "Figma Copilot" (execute_batch, 배치 작업)
  │
  ├─ 변수/토큰 조회
  │  └─ "Figma REST API" (GET /variables)
  │
  ├─ 실시간 협업/프로토타입
  │  └─ "Talk To Figma MCP" (WebSocket)
  │
  └─ 위 중 여러 개
     └─ 다중 도구 조합 (5.1 참고)
```

---

### 5.3 API 호출 순서 베스트 프랙티스

**1순위**: Context MCP (무료, 최적화)
```typescript
const design = await figmaService.getRawFile(fileKey);
```

**2순위**: Figma REST API (공식, 안정적)
```typescript
const variables = await fetch(
  `https://api.figma.com/v1/files/${fileKey}/variables`
).then(r => r.json());
```

**3순위**: Talk To Figma (완전한 기능)
```typescript
const updated = await client.set_text_content({
  node_id: "1234:5678",
  content: "수정됨"
});
```

**4순위**: Copilot (배치 성능)
```typescript
await copilot.execute_batch({ commands: [...] });
```

---

## 요약: 도구 선택 빠른 참조표

| 작업 | 권장 도구 | 대안 | 이유 |
|------|---------|------|------|
| 설계 데이터 추출 | Context MCP | Talk To Figma | 구조화된 Extractor |
| 이미지 다운로드 | Context MCP | - | 파일 직접 저장 |
| 변수 조회 | REST API | - | 공식 API |
| 텍스트 수정 | Talk To Figma | Copilot | 개별/배치 선택 |
| 배치 수정 | Copilot | Talk To Figma | 50-90% 성능 개선 |
| 레이아웃 변경 | Talk To Figma | Copilot | 둘 다 동등 |
| 실시간 협업 | Talk To Figma | - | WebSocket 지원 |
| 프로토타입 관리 | Talk To Figma | - | 연결 생성 API |
| 컴포넌트 관리 | Talk To Figma | - | 인스턴스 생성 |
| 성능 최적화 | Copilot | - | 배치 API |

---

**문서 버전**: 1.0
**마지막 업데이트**: 2025-11-19
**기반 자료**: figma-mcp-official-docs.md (Section 9)
**참고**: 이 매트릭스는 현재 공개된 문서를 기반으로 작성되었으며, 주기적 업데이트가 필요합니다.
