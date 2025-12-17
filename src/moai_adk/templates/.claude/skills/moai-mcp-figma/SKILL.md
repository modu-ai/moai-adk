---
name: moai-mcp-figma
description: Figma MCP integration specialist for design system extraction, component generation, and design-to-code workflows. Use when integrating Figma designs into development.
version: 1.0.0
category: integration
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
tags:
 - mcp
 - figma
 - design-system
 - design-tokens
 - components
updated: 2025-12-07
status: active
author: MoAI-ADK Team
---

# Figma MCP Integration Specialist

## Quick Reference (30 seconds)

Figma MCP 통합 전문가 - Figma API를 통한 디자인 시스템 추출, 컴포넌트 생성, 디자인 토큰 동기화, 그리고 디자인에서 코드로의 완전한 워크플로우를 제공하는 MCP 서버 기반 통합 시스템입니다.

핵심 기능:
- Figma API Integration: MCP 서버를 통한 Figma 파일 및 노드 접근
- Design System Extraction: 디자인 시스템 컴포넌트 자동 추출
- Design Tokens Sync: 색상, 타이포그래피, 간격 등 디자인 토큰 동기화
- Component Generation: React, Vue, Svelte 등 프레임워크별 컴포넌트 생성
- Auto-layout to CSS: Figma Auto-layout을 CSS Flexbox/Grid로 변환
- Asset Export: 아이콘, 이미지 등 디자인 자산 자동 추출

사용 시기:
- Figma 디자인을 개발 코드로 변환할 때
- 디자인 시스템을 프로젝트에 통합할 때
- 디자인 토큰을 코드와 동기화할 때
- 컴포넌트 라이브러리를 자동 생성할 때
- Figma 파일 변경 사항을 추적할 때

---

## Implementation Guide (5 minutes)

### Figma MCP 서버 설정

기본 서버 초기화:
```python
from moai_mcp_figma import FigmaMCPServer
import os

# Figma MCP 서버 생성
figma_server = FigmaMCPServer("figma-integration-server")

# Figma API 토큰 설정
figma_server.setup({
    'api_key': os.getenv('FIGMA_TOKEN'),
    'team_id': os.getenv('FIGMA_TEAM_ID')  # Optional
})

# 서버 시작
figma_server.start(port=3001)
```

Claude Desktop 설정 (claude_desktop_config.json):
```json
{
  "mcpServers": {
    "figma": {
      "command": "python",
      "args": ["-m", "moai_mcp_figma"],
      "env": {
        "FIGMA_TOKEN": "your-figma-personal-access-token"
      }
    }
  }
}
```

### Figma API 인증

Personal Access Token 생성:
1. Figma 계정 설정 → Personal Access Tokens
2. "Generate new token" 클릭
3. 토큰 이름 설정 (예: "MCP Integration")
4. 토큰 복사 및 환경 변수에 저장

토큰 검증:
```python
# 토큰 유효성 확인
validation = await figma_server.invoke_tool("validate_token", {})

# 사용자 정보 확인
user_info = await figma_server.invoke_tool("get_user_info", {})
print(f"Authenticated as: {user_info['name']}")
```

### 디자인 시스템 추출

Figma 파일에서 컴포넌트 추출:
```python
# 파일 ID는 Figma URL에서 확인 가능
# https://www.figma.com/file/{FILE_ID}/Design-System
file_id = "abc123xyz789"

# 모든 컴포넌트 추출
components = await figma_server.invoke_tool("extract_figma_components", {
    "file_id": file_id,
    "include_variants": True,
    "include_instances": True
})

# 결과 구조:
# {
#   "components": [
#     {
#       "id": "123:456",
#       "name": "Button/Primary",
#       "type": "COMPONENT",
#       "properties": {...},
#       "children": [...]
#     }
#   ],
#   "component_sets": [...]
# }
```

특정 페이지 또는 프레임에서 추출:
```python
# 특정 페이지의 컴포넌트만 추출
components = await figma_server.invoke_tool("extract_components_from_page", {
    "file_id": file_id,
    "page_name": "Components",
    "filter_by_type": ["COMPONENT", "COMPONENT_SET"]
})

# 특정 프레임의 컴포넌트 추출
frame_components = await figma_server.invoke_tool("extract_components_from_frame", {
    "file_id": file_id,
    "frame_name": "Design System/Buttons"
})
```

### 디자인 토큰 동기화

색상 토큰 추출:
```python
# 색상 스타일 추출
color_tokens = await figma_server.invoke_tool("extract_color_tokens", {
    "file_id": file_id,
    "output_format": "css"
})

# CSS 변수로 출력:
# :root {
#   --color-primary-500: #3b82f6;
#   --color-primary-600: #2563eb;
#   --color-gray-50: #f9fafb;
# }

# TypeScript 형식으로 추출
ts_colors = await figma_server.invoke_tool("extract_color_tokens", {
    "file_id": file_id,
    "output_format": "typescript"
})

# export const colors = {
#   primary: {
#     500: '#3b82f6',
#     600: '#2563eb',
#   },
#   gray: {
#     50: '#f9fafb',
#   }
# } as const;
```

타이포그래피 토큰 추출:
```python
# 텍스트 스타일 추출
typography_tokens = await figma_server.invoke_tool("extract_typography_tokens", {
    "file_id": file_id,
    "output_format": "css"
})

# CSS 클래스로 출력:
# .text-heading-1 {
#   font-family: 'Inter', sans-serif;
#   font-size: 36px;
#   font-weight: 700;
#   line-height: 1.2;
# }
```

간격 및 레이아웃 토큰:
```python
# 간격 시스템 추출
spacing_tokens = await figma_server.invoke_tool("extract_spacing_tokens", {
    "file_id": file_id,
    "base_unit": 4  # 4px 기준 (4, 8, 12, 16, 24, 32, ...)
})

# --spacing-1: 4px;
# --spacing-2: 8px;
# --spacing-3: 12px;
# --spacing-4: 16px;
```

통합 디자인 토큰 추출:
```python
# 모든 토큰을 한 번에 추출
all_tokens = await figma_server.invoke_tool("sync_figma_tokens", {
    "file_id": file_id,
    "token_types": ["colors", "typography", "spacing", "effects"],
    "output_format": "json",
    "include_variants": True
})

# JSON 형식:
# {
#   "colors": {...},
#   "typography": {...},
#   "spacing": {...},
#   "effects": {...}
# }
```

### 컴포넌트 코드 생성

React 컴포넌트 생성:
```python
# 단일 컴포넌트 생성
button_code = await figma_server.invoke_tool("generate_react_component", {
    "file_id": file_id,
    "component_id": "123:456",  # Figma 컴포넌트 ID
    "target_library": "shadcn",  # shadcn/ui 기반
    "include_typescript": True,
    "include_stories": True  # Storybook stories 포함
})

# 결과:
# {
#   "component": "export const Button = ({ variant = 'primary', ... }) => { ... }",
#   "types": "export interface ButtonProps { variant?: 'primary' | 'secondary'; ... }",
#   "story": "export default { component: Button, ... }"
# }
```

여러 컴포넌트 일괄 생성:
```python
# 페이지의 모든 컴포넌트 생성
all_components = await figma_server.invoke_tool("generate_component_library", {
    "file_id": file_id,
    "page_name": "Components",
    "target_framework": "react",
    "target_library": "shadcn",
    "output_directory": "./src/components/ui",
    "include_tests": True
})

# 생성된 파일 구조:
# src/components/ui/
# ├── button.tsx
# ├── button.test.tsx
# ├── button.stories.tsx
# ├── input.tsx
# ├── input.test.tsx
# └── input.stories.tsx
```

Vue 컴포넌트 생성:
```python
# Vue 3 Composition API 컴포넌트
vue_component = await figma_server.invoke_tool("generate_vue_component", {
    "file_id": file_id,
    "component_id": "123:456",
    "composition_api": True,
    "typescript": True
})

# <script setup lang="ts">
# defineProps<ButtonProps>()
# </script>
```

### Auto-layout을 CSS로 변환

Flexbox 변환:
```python
# Figma Auto-layout을 CSS Flexbox로 변환
layout_css = await figma_server.invoke_tool("convert_autolayout_to_css", {
    "file_id": file_id,
    "node_id": "789:012",  # Auto-layout이 적용된 노드
    "css_framework": "tailwind"  # tailwind, css, scss
})

# Tailwind 클래스:
# "flex flex-col gap-4 items-center justify-between"

# 일반 CSS:
# display: flex;
# flex-direction: column;
# gap: 16px;
# align-items: center;
# justify-content: space-between;
```

Grid 레이아웃 변환:
```python
# Grid 기반 레이아웃 변환
grid_css = await figma_server.invoke_tool("convert_to_grid", {
    "file_id": file_id,
    "node_id": "789:012",
    "responsive": True,  # 반응형 Grid 생성
    "breakpoints": {
        "sm": 640,
        "md": 768,
        "lg": 1024
    }
})
```

### 자산 추출 및 최적화

아이콘 추출:
```python
# SVG 아이콘 추출
icons = await figma_server.invoke_tool("export_icons", {
    "file_id": file_id,
    "page_name": "Icons",
    "format": "svg",
    "optimize": True,  # SVGO를 통한 최적화
    "output_directory": "./src/assets/icons"
})

# React 컴포넌트로 변환
icon_components = await figma_server.invoke_tool("convert_icons_to_components", {
    "icons": icons,
    "target_framework": "react",
    "include_typescript": True
})
```

이미지 및 자산 추출:
```python
# 다양한 형식으로 이미지 추출
assets = await figma_server.invoke_tool("export_assets", {
    "file_id": file_id,
    "node_ids": ["123:456", "789:012"],
    "format": "png",  # png, jpg, svg, pdf
    "scale": 2,  # 2x 해상도
    "output_directory": "./src/assets/images"
})

# 반응형 이미지 생성 (1x, 2x, 3x)
responsive_assets = await figma_server.invoke_tool("export_responsive_assets", {
    "file_id": file_id,
    "node_ids": ["123:456"],
    "formats": ["png", "webp"],
    "scales": [1, 2, 3]
})
```

---

## Advanced Patterns (10+ minutes)

고급 패턴 및 전체 워크플로우는 별도 문서를 참조하세요:
- `advanced-patterns.md` - 완전한 Design-to-Code 파이프라인, 컴포넌트 변형 처리, Figma Variables 변환
- `sync-workflows.md` - 실시간 동기화, 웹훅, 디자인 시스템 검증
- `component-strategies.md` - 복합 컴포넌트, 반응형 컴포넌트 생성 전략

핵심 고급 기능:
- Complete Design-to-Code Pipeline: 토큰 추출부터 문서 생성까지 전체 자동화
- Component Variants: Component Set의 모든 변형을 Props로 자동 변환
- Figma Variables: Light/Dark 테마별 디자인 토큰 생성
- Real-time Sync: 파일 변경 감지 및 자동 재생성
- Design System Validation: 디자인 일관성 검증 및 보고

---

## Works Well With

보완 스킬:
- `moai-domain-frontend` - 프론트엔드 컴포넌트 통합 및 최적화
- `moai-domain-uiux` - UI/UX 디자인 패턴 및 접근성
- `moai-library-shadcn` - shadcn/ui 컴포넌트 라이브러리 통합
- `moai-library-tailwind` - Tailwind CSS 유틸리티 클래스 생성
- `moai-docs-generation` - 디자인 시스템 문서 자동 생성

외부 서비스:
- Figma (디자인 시스템, 컴포넌트 라이브러리)
- Storybook (컴포넌트 문서화 및 테스트)
- Chromatic (시각적 회귀 테스트)

개발 도구:
- React, Vue, Svelte (컴포넌트 프레임워크)
- TypeScript (타입 안전성)
- Tailwind CSS (스타일링)
- shadcn/ui (컴포넌트 라이브러리)

---

*자세한 구현 패턴, API 레퍼런스, 고급 워크플로우는 별도 문서를 참조하세요.*
